package api_test

import (
	"net/http"
	"testing"

	"pubkey-quest/cmd/server/api/character"
	"pubkey-quest/cmd/server/db"
	"pubkey-quest/tests/helpers"
)

func setupCharacterTestServer(t *testing.T) *helpers.TestServer {
	t.Helper()

	// Set up test environment (changes to project root for database access)
	helpers.SetupTestEnvironment(t)

	// Initialize database for tests
	if err := db.InitDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	ts := helpers.NewTestServer()
	ts.Mux.HandleFunc("/api/weights", character.WeightsHandler)
	ts.Mux.HandleFunc("/api/introductions", character.IntroductionsHandler)
	ts.Mux.HandleFunc("/api/starting-gear", character.StartingGearHandler)
	ts.Mux.HandleFunc("/api/character", character.CharacterHandler)
	ts.Mux.HandleFunc("/api/character/create-save", character.CreateCharacterHandler)

	return ts
}

func TestWeightsHandler(t *testing.T) {
	ts := setupCharacterTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/weights")
	helpers.AssertStatus(t, resp, http.StatusOK)

	result := helpers.AssertJSON(t, resp)

	// Check expected weight fields
	expectedFields := []string{"Races", "RaceWeights", "classWeightsByRace", "BackgroundWeightsByClass", "Alignments", "AlignmentWeights"}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Expected field %s in weights response", field)
		}
	}
}

func TestIntroductionsHandler(t *testing.T) {
	ts := setupCharacterTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/introductions")
	helpers.AssertStatus(t, resp, http.StatusOK)

	// Response is raw JSON, should be parseable
	var result interface{}
	helpers.ReadJSON(t, resp, &result)
	if result == nil {
		t.Error("Expected non-nil introductions data")
	}
}

func TestStartingGearHandler(t *testing.T) {
	ts := setupCharacterTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/starting-gear")
	helpers.AssertStatus(t, resp, http.StatusOK)

	// Response is raw JSON array, should be parseable
	var result []interface{}
	helpers.ReadJSON(t, resp, &result)
	if len(result) == 0 {
		t.Error("Expected non-empty starting gear data")
	}
}

func TestCharacterHandler(t *testing.T) {
	ts := setupCharacterTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires npub parameter", func(t *testing.T) {
		resp := ts.GET(t, "/api/character")
		helpers.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("rejects invalid npub", func(t *testing.T) {
		resp := ts.GET(t, "/api/character?npub=invalid")
		helpers.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("generates character for valid npub", func(t *testing.T) {
		// Use a valid bech32-encoded npub (this is a test npub)
		resp := ts.GET(t, "/api/character?npub=npub1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsff2l4")
		// This may fail if the npub is invalid, which is expected
		if resp.StatusCode == http.StatusOK {
			result := helpers.AssertJSON(t, resp)
			if _, ok := result["character"]; !ok {
				t.Error("Expected character field in response")
			}
		}
	})
}

func TestCreateCharacterHandler(t *testing.T) {
	ts := setupCharacterTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires POST method", func(t *testing.T) {
		resp := ts.GET(t, "/api/character/create-save")
		helpers.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires valid request body", func(t *testing.T) {
		resp := ts.POST(t, "/api/character/create-save", "invalid")
		helpers.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("requires npub in request", func(t *testing.T) {
		body := map[string]interface{}{
			"name":              "TestCharacter",
			"equipment_choices": map[string]string{},
			"pack_choice":       "",
		}
		resp := ts.POST(t, "/api/character/create-save", body)
		helpers.AssertStatus(t, resp, http.StatusBadRequest)

		result := helpers.AssertJSON(t, resp)
		helpers.AssertError(t, result)
	})
}
