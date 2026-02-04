THIS WILL HAVE TO BE REDONE AFTER PROJECT RESTRUCTURING

# Swagger API Documentation Setup

This guide explains how to generate and serve Swagger/OpenAPI documentation for the Nostr Hero API.

## Prerequisites

1. Go 1.21+ installed
2. Server builds successfully (`go build ./...`)

## Installation

### 1. Install Swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

Ensure `$GOPATH/bin` is in your PATH.

### 2. Add Go Dependencies

```bash
cd server
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/swag
```

## Generating Documentation

### Generate from annotations

Run from the `server/` directory:

```bash
swag init -g api/routes.go -o docs
```

This creates/updates:

- `docs/docs.go` - Go code embedding the spec
- `docs/swagger.json` - OpenAPI 3.0 spec (JSON)
- `docs/swagger.yaml` - OpenAPI 3.0 spec (YAML)

### Regenerate after changes

Re-run `swag init` whenever you modify handler annotations:

```bash
swag init -g api/routes.go -o docs
```

## Enabling Swagger UI

### 1. Uncomment imports in `api/routes.go`

Find these lines and uncomment them:

```go
import (
    // ... other imports ...

    // Uncomment after running: swag init -g api/routes.go -o docs
    _ "nostr-hero/docs"
    httpSwagger "github.com/swaggo/http-swagger"
)
```

### 2. Uncomment the route handler

Find this line in `RegisterRoutes()` and uncomment it:

```go
// Swagger UI - uncomment after running: swag init -g api/routes.go -o docs
mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
```

### 3. Access Swagger UI

Start the server and visit:

```
http://localhost:8080/swagger/
```

## Annotation Reference

### General API Info (in `api/routes.go`)

```go
// @title           Nostr Hero API
// @version         1.0
// @description     REST API for the Nostr Hero game server.
// @host            localhost:8080
// @BasePath        /api
```

### Handler Annotations

Place these comments directly above handler functions:

```go
// HandlerName godoc
// @Summary      Short one-line description
// @Description  Longer detailed description (optional)
// @Tags         GroupName
// @Accept       json
// @Produce      json
// @Param        name   query/path/body   type      required   "description"
// @Success      200    {object}          TypeName
// @Failure      400    {object}          ErrorType  "error description"
// @Router       /endpoint [method]
func HandlerName(w http.ResponseWriter, r *http.Request) {
```

### Parameter Types

| Location | Usage            | Example                                                 |
| -------- | ---------------- | ------------------------------------------------------- |
| `query`  | URL query string | `@Param npub query string true "Nostr pubkey"`          |
| `path`   | URL path segment | `@Param id path string true "Item ID"`                  |
| `body`   | Request body     | `@Param request body CreateRequest true "Request body"` |
| `header` | HTTP header      | `@Param X-Auth header string false "Auth token"`        |

### Response Types

```go
// Single object
// @Success 200 {object} ResponseType

// Array of objects
// @Success 200 {array} ItemType

// Primitive type
// @Success 200 {string} string "Success message"

// Empty response
// @Success 204
```

### Common Patterns

**GET with query params:**

```go
// @Param   npub      query   string  true   "Nostr public key"
// @Param   save_id   query   string  true   "Save file ID"
// @Success 200       {object} SessionState
// @Router  /session/state [get]
```

**POST with JSON body:**

```go
// @Param   request   body    CreateCharacterRequest  true  "Character creation data"
// @Success 200       {object} CreateCharacterResponse
// @Router  /character/create-save [post]
```

**Path parameters:**

```go
// @Param   id   path   string   true   "Spell ID"
// @Success 200  {object} Spell
// @Router  /spells/{id} [get]
```

**Multiple response codes:**

```go
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 404 {object} ErrorResponse "Not found"
// @Failure 500 {object} ErrorResponse "Server error"
```

## Model Annotations

Add `swagger:model` to struct types for better documentation:

```go
// ProfileMetadata represents a Nostr user profile
// swagger:model ProfileMetadata
type ProfileMetadata struct {
    Name        string `json:"name" example:"satoshi"`
    DisplayName string `json:"display_name" example:"Satoshi Nakamoto"`
    About       string `json:"about" example:"Creator of Bitcoin"`
}
```

### Field Tags

```go
type Example struct {
    ID       string `json:"id" example:"abc123"`           // Example value
    Name     string `json:"name" minLength:"1" maxLength:"100"` // Validation
    Age      int    `json:"age" minimum:"0" maximum:"150"` // Range
    Email    string `json:"email" format:"email"`          // Format hint
    Optional string `json:"optional,omitempty"`            // Optional field
}
```

## Grouping with Tags

Handlers are grouped in Swagger UI by their `@Tags` annotation:

- `GameData` - Static game content (items, spells, monsters)
- `Character` - Character generation and creation
- `Auth` - Nostr authentication
- `Saves` - Save file management
- `Session` - In-memory session state
- `Game` - Game actions and state
- `Shop` - Shop transactions
- `Profile` - Player profiles
- `Debug` - Debug endpoints (dev only)

## Troubleshooting

### "swag: command not found"

Ensure `$GOPATH/bin` is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Changes not appearing

1. Re-run `swag init -g api/routes.go -o docs`
2. Restart the server
3. Hard refresh browser (Ctrl+Shift+R)

### Import errors after generating

Ensure you've run:

```bash
go get -u github.com/swaggo/http-swagger
go mod tidy
```

### Annotations not parsed

- Annotations must be directly above the function (no blank lines)
- Must start with `// @` (space after //)
- Function must be exported (capital letter)

## File Locations

```
server/
├── api/
│   ├── routes.go          # Main API info annotations + route registration
│   ├── doc.go             # Package-level swagger:meta (optional)
│   ├── profile.go         # Handler with annotations
│   ├── saves.go           # Handler with annotations
│   ├── character/         # Character handlers
│   ├── data/              # Game data handlers
│   └── game/              # Game action handlers
└── docs/
    ├── SWAGGER_SETUP.md   # This file
    ├── docs.go            # Generated - DO NOT EDIT
    ├── swagger.json       # Generated - OpenAPI spec
    └── swagger.yaml       # Generated - OpenAPI spec
```

## CI/CD Integration

Add to your build pipeline:

```yaml
- name: Generate Swagger docs
  run: |
    go install github.com/swaggo/swag/cmd/swag@latest
    cd server && swag init -g api/routes.go -o docs

- name: Verify docs are up to date
  run: |
    git diff --exit-code server/docs/
```
