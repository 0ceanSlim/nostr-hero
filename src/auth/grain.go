package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/0ceanslim/grain/client/core/tools"
	"github.com/0ceanslim/grain/client/session"
	"nostr-hero/src/utils"
)

// AuthHandler handles all authentication-related operations using grain client
type AuthHandler struct {
	config *utils.Config
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(cfg *utils.Config) *AuthHandler {
	return &AuthHandler{config: cfg}
}

// LoginRequest represents a login request for Nostr Hero
type LoginRequest struct {
	PublicKey     string                         `json:"public_key,omitempty"`
	PrivateKey    string                         `json:"private_key,omitempty"`  // nsec format
	SigningMethod session.SigningMethod         `json:"signing_method"`
	Mode          session.SessionInteractionMode `json:"mode,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success     bool                `json:"success"`
	Message     string              `json:"message"`
	Session     *session.UserSession `json:"session,omitempty"`
	PublicKey   string              `json:"public_key,omitempty"`
	NPub        string              `json:"npub,omitempty"`
	Error       string              `json:"error,omitempty"`
}

// SessionResponse represents a session status response
type SessionResponse struct {
	Success  bool                `json:"success"`
	IsActive bool                `json:"is_active"`
	Session  *session.UserSession `json:"session,omitempty"`
	NPub     string              `json:"npub,omitempty"`
	Error    string              `json:"error,omitempty"`
}

// KeyPairResponse represents a key generation response
type KeyPairResponse struct {
	Success bool           `json:"success"`
	KeyPair *tools.KeyPair `json:"key_pair,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// HandleLogin handles user login/authentication using grain session system
func (auth *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := auth.validateLoginRequest(&req); err != nil {
		auth.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Default to write mode for game functionality
	if req.Mode == "" {
		req.Mode = session.WriteMode
	}

	// Create session init request for grain
	sessionReq := session.SessionInitRequest{
		RequestedMode: req.Mode,
		SigningMethod: req.SigningMethod,
	}

	// Handle different signing methods
	switch req.SigningMethod {
	case session.BrowserExtension:
		if req.PublicKey == "" {
			auth.sendErrorResponse(w, "Public key required for browser extension signing", http.StatusBadRequest)
			return
		}
		sessionReq.PublicKey = req.PublicKey

	case session.AmberSigning:
		if req.PublicKey == "" {
			auth.sendErrorResponse(w, "Public key required for Amber signing", http.StatusBadRequest)
			return
		}
		sessionReq.PublicKey = req.PublicKey

	case session.EncryptedKey:
		if req.PrivateKey == "" {
			auth.sendErrorResponse(w, "Private key required for encrypted key signing", http.StatusBadRequest)
			return
		}

		var privateKeyHex string
		var err error

		// Handle both nsec and hex format
		if strings.HasPrefix(req.PrivateKey, "nsec") {
			// Decode nsec to get hex private key
			privateKeyHex, err = tools.DecodeNsec(req.PrivateKey)
			if err != nil {
				auth.sendErrorResponse(w, fmt.Sprintf("Invalid nsec format: %v", err), http.StatusBadRequest)
				return
			}
		} else if len(req.PrivateKey) == 64 {
			// Assume it's already hex format
			if matched, _ := regexp.MatchString("^[0-9a-fA-F]{64}$", req.PrivateKey); !matched {
				auth.sendErrorResponse(w, "Invalid hex private key format", http.StatusBadRequest)
				return
			}
			privateKeyHex = req.PrivateKey
		} else {
			auth.sendErrorResponse(w, "Private key must be nsec format or 64-character hex", http.StatusBadRequest)
			return
		}

		pubkey, err := tools.DerivePublicKey(privateKeyHex)
		if err != nil {
			auth.sendErrorResponse(w, fmt.Sprintf("Failed to derive public key: %v", err), http.StatusBadRequest)
			return
		}

		sessionReq.PublicKey = pubkey
		sessionReq.PrivateKey = req.PrivateKey

	default:
		// For read-only mode or other cases
		if req.PublicKey != "" {
			sessionReq.PublicKey = req.PublicKey
		} else {
			auth.sendErrorResponse(w, "Either public key or private key must be provided", http.StatusBadRequest)
			return
		}
	}

	// Create user session using grain
	userSession, err := session.CreateUserSession(w, sessionReq)
	if err != nil {
		auth.sendErrorResponse(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusBadRequest)
		return
	}

	// Generate npub for response
	npub, _ := tools.EncodePubkey(userSession.PublicKey)

	log.Printf("🎮 Nostr Hero user logged in: %s (%s mode)", userSession.PublicKey[:16]+"...", userSession.Mode)

	response := LoginResponse{
		Success:   true,
		Message:   "Login successful",
		Session:   userSession,
		PublicKey: userSession.PublicKey,
		NPub:      npub,
	}

	auth.sendJSONResponse(w, response, http.StatusOK)
}

// HandleLogout handles user logout
func (auth *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear session using grain session manager
	if session.SessionMgr != nil {
		session.SessionMgr.ClearSession(w, r)
	}

	log.Println("🎮 Nostr Hero user logged out")

	response := map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	}

	auth.sendJSONResponse(w, response, http.StatusOK)
}

// HandleSession handles session status checks
func (auth *AuthHandler) HandleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from grain session manager
	if !session.IsSessionManagerInitialized() {
		response := SessionResponse{
			Success:  true,
			IsActive: false,
			Error:    "session manager not initialized",
		}
		auth.sendJSONResponse(w, response, http.StatusOK)
		return
	}

	userSession := session.SessionMgr.GetCurrentUser(r)
	if userSession == nil {
		response := SessionResponse{
			Success:  true,
			IsActive: false,
		}
		auth.sendJSONResponse(w, response, http.StatusOK)
		return
	}

	// Generate npub for response
	npub, _ := tools.EncodePubkey(userSession.PublicKey)

	response := SessionResponse{
		Success:  true,
		IsActive: true,
		Session:  userSession,
		NPub:     npub,
	}

	auth.sendJSONResponse(w, response, http.StatusOK)
}

// HandleGenerateKeys handles key pair generation
func (auth *AuthHandler) HandleGenerateKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate new key pair using grain tools
	keyPair, err := tools.GenerateKeyPair()
	if err != nil {
		auth.sendErrorResponse(w, fmt.Sprintf("Failed to generate keys: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("🎮 Generated new key pair for Nostr Hero: %s", keyPair.Npub)

	response := KeyPairResponse{
		Success: true,
		KeyPair: keyPair,
	}

	auth.sendJSONResponse(w, response, http.StatusOK)
}

// HandleAmberCallback processes callbacks from Amber app
func (auth *AuthHandler) HandleAmberCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎮 Amber callback received: method=%s, url=%s", r.Method, r.URL.String())

	// Parse query parameters
	eventParam := r.URL.Query().Get("event")
	if eventParam == "" {
		log.Printf("❌ Amber callback missing event parameter")
		auth.renderAmberError(w, "Missing event data from Amber")
		return
	}

	// Extract public key from event parameter
	publicKey, err := auth.extractPublicKeyFromAmber(eventParam)
	if err != nil {
		log.Printf("❌ Failed to extract public key from amber response: %v", err)
		auth.renderAmberError(w, "Invalid response from Amber: "+err.Error())
		return
	}

	log.Printf("✅ Amber callback processed successfully: %s...", publicKey[:16])

	// Create session
	sessionRequest := session.SessionInitRequest{
		PublicKey:     publicKey,
		RequestedMode: session.WriteMode, // Game requires write mode
		SigningMethod: session.AmberSigning,
	}

	_, err = session.CreateUserSession(w, sessionRequest)
	if err != nil {
		log.Printf("❌ Failed to create amber session: %v", err)
		auth.renderAmberError(w, "Failed to create session")
		return
	}

	log.Printf("✅ Amber session created successfully: %s...", publicKey[:16])

	// Render success page with auto-redirect
	auth.renderAmberSuccess(w, publicKey)
}

// GetCurrentUser is a utility function to get the current authenticated user
func GetCurrentUser(r *http.Request) *session.UserSession {
	if !session.IsSessionManagerInitialized() {
		return nil
	}
	return session.SessionMgr.GetCurrentUser(r)
}

// Helper methods

func (auth *AuthHandler) validateLoginRequest(req *LoginRequest) error {
	if req.SigningMethod == "" {
		return fmt.Errorf("signing method is required")
	}

	if req.Mode == "" {
		req.Mode = session.WriteMode // Default to write mode for game
	}

	return nil
}

func (auth *AuthHandler) extractPublicKeyFromAmber(eventParam string) (string, error) {
	// Handle compressed response (starts with "Signer1")
	if strings.HasPrefix(eventParam, "Signer1") {
		return "", fmt.Errorf("compressed Amber responses not supported")
	}

	// For get_public_key, event parameter should contain the public key directly
	publicKey := strings.TrimSpace(eventParam)

	// Validate public key format (64 hex characters)
	pubKeyRegex := regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
	if !pubKeyRegex.MatchString(publicKey) {
		return "", fmt.Errorf("invalid public key format from Amber")
	}

	return publicKey, nil
}

func (auth *AuthHandler) renderAmberSuccess(w http.ResponseWriter, publicKey string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Amber Login Success - Nostr Hero</title>
    <style>
        body {
            font-family: 'Pixelify Sans', monospace;
            margin: 0;
            padding: 20px;
            background: #001100;
            color: #00ff41;
            text-align: center;
        }
        .success { color: #00ff41; margin: 20px 0; }
        .loading { color: #888; }
    </style>
</head>
<body>
    <div class="success">
        <h2>✅ Amber Login Successful!</h2>
        <p>Connected successfully. Returning to Nostr Hero...</p>
    </div>
    <div class="loading">
        <p>Please wait...</p>
    </div>

    <script>
        const amberResult = {
            success: true,
            publicKey: '` + publicKey + `',
            timestamp: Date.now()
        };

        try {
            localStorage.setItem('amber_callback_result', JSON.stringify(amberResult));
            console.log('Stored Amber success result in localStorage');
        } catch (error) {
            console.error('Failed to store Amber result:', error);
        }

        if (window.opener && !window.opener.closed) {
            try {
                window.opener.postMessage({
                    type: 'amber_success',
                    publicKey: '` + publicKey + `'
                }, window.location.origin);
                console.log('Sent success message to opener window');
            } catch (error) {
                console.error('Failed to send message to opener:', error);
            }
        }

        setTimeout(() => {
            try {
                if (window.opener && !window.opener.closed) {
                    window.close();
                } else {
                    window.location.href = '/game?amber_login=success';
                }
            } catch (error) {
                console.error('Failed to navigate:', error);
                window.location.href = '/game';
            }
        }, 1500);
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func (auth *AuthHandler) renderAmberError(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Amber Login Error - Nostr Hero</title>
    <style>
        body {
            font-family: 'Pixelify Sans', monospace;
            margin: 0;
            padding: 20px;
            background: #001100;
            color: #ff4444;
            text-align: center;
        }
        .error { color: #ff4444; margin: 20px 0; }
        .retry { margin-top: 20px; }
        .retry a { color: #00ff41; text-decoration: none; }
    </style>
</head>
<body>
    <div class="error">
        <h2>❌ Amber Login Failed</h2>
        <p>` + errorMsg + `</p>
    </div>
    <div class="retry">
        <a href="/game">← Return to game</a>
    </div>

    <script>
        if (window.opener) {
            window.opener.postMessage({
                type: 'amber_error',
                error: '` + errorMsg + `'
            }, window.location.origin);
            setTimeout(() => window.close(), 3000);
        }
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func (auth *AuthHandler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (auth *AuthHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	auth.sendJSONResponse(w, response, statusCode)
}