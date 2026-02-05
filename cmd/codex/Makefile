# CODEX Makefile
# Run from project root: make -f cmd/codex/Makefile <target>
# Or from cmd/codex/: make <target>

.PHONY: all build build-js build-go build-linux run clean clean-js clean-go deps deps-js deps-go migrate help

# Detect if we're running from project root or cmd/codex
ifeq ($(wildcard go.mod),go.mod)
	# Running from project root
	ROOT_DIR := .
	CODEX_DIR := cmd/codex
else
	# Running from cmd/codex
	ROOT_DIR := ../..
	CODEX_DIR := .
endif

# Detect OS
ifeq ($(OS),Windows_NT)
	BINARY_NAME := codex.exe
	DETECTED_OS := Windows
else
	BINARY_NAME := codex
	DETECTED_OS := Linux
endif

BINARY_PATH := $(ROOT_DIR)/$(BINARY_NAME)

# Default target - build everything
all: build

# Build everything (JS + Go)
build: build-js build-go

# Build JavaScript/CSS assets (CODEX frontend)
build-js:
	@echo "📦 Building CODEX frontend..."
	@cd $(CODEX_DIR) && npm run build
	@echo "✅ Frontend build complete!"

# Build Go executable (auto-detect OS)
build-go:
	@echo "🔨 Building CODEX server for $(DETECTED_OS)..."
	@cd $(ROOT_DIR) && go build -o $(BINARY_NAME) ./cmd/codex
	@echo "✅ Server build complete! ($(BINARY_PATH))"

# Build Go executable (Linux cross-compile)
build-linux:
	@echo "🔨 Building CODEX server for Linux (cross-compile)..."
ifeq ($(OS),Windows_NT)
	@cd $(ROOT_DIR) && powershell -Command "$$env:GOOS='linux'; $$env:GOARCH='amd64'; go build -o codex ./cmd/codex"
else
	@cd $(ROOT_DIR) && GOOS=linux GOARCH=amd64 go build -o codex ./cmd/codex
endif
	@echo "✅ Linux server build complete! ($(ROOT_DIR)/codex)"

# Build and run CODEX
run: build
	@echo "🚀 Starting CODEX..."
	@cd $(ROOT_DIR) && ./$(BINARY_NAME)

# Run with auto-restart on changes (requires Air)
dev:
	@echo "🔥 Starting CODEX in development mode..."
	@cd $(ROOT_DIR) && air -c cmd/codex/.air.toml

# Clean all build artifacts
clean: clean-js clean-go
	@echo "✅ All build artifacts cleaned!"

# Clean JavaScript build output
clean-js:
	@echo "🧹 Cleaning CODEX frontend build..."
	@rm -rf $(ROOT_DIR)/www/dist/codex
	@rm -rf $(CODEX_DIR)/dist

# Clean Go executable
clean-go:
	@echo "🧹 Cleaning CODEX server executables..."
	@rm -f $(ROOT_DIR)/codex.exe $(ROOT_DIR)/codex
	@echo "✅ Cleaned codex binaries"

# Install all dependencies
deps: deps-js deps-go

# Install JavaScript dependencies
deps-js:
	@echo "📦 Installing CODEX frontend dependencies..."
	@cd $(CODEX_DIR) && npm install
	@echo "✅ Frontend dependencies installed!"

# Install Go dependencies
deps-go:
	@echo "📦 Installing CODEX server dependencies..."
	@cd $(ROOT_DIR) && go mod download
	@echo "✅ Server dependencies installed!"

# Run database migration
migrate:
	@echo "🔄 Running database migration..."
	@cd $(ROOT_DIR) && ./$(BINARY_NAME) --migrate
	@echo "✅ Migration complete!"

# Help
help:
	@echo "🎯 CODEX - Content Organization & Data Entry eXperience"
	@echo ""
	@echo "📋 Available targets:"
	@echo "  make all          - Build everything (JS + Go)"
	@echo "  make build        - Build everything (JS + Go)"
	@echo "  make build-js     - Build frontend only"
	@echo "  make build-go     - Build server for current OS"
	@echo "  make build-linux  - Build server for Linux (cross-compile)"
	@echo "  make run          - Build and run CODEX"
	@echo "  make dev          - Run with auto-restart (requires Air)"
	@echo "  make clean        - Remove all build artifacts"
	@echo "  make clean-js     - Remove frontend build"
	@echo "  make clean-go     - Remove server executable"
	@echo "  make deps         - Install all dependencies"
	@echo "  make deps-js      - Install frontend dependencies"
	@echo "  make deps-go      - Install server dependencies"
	@echo "  make migrate      - Run database migration"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "🔧 Configuration:"
	@echo "  Config file: cmd/codex/config.yml (copy from config.example.yml)"
	@echo "  Port: Set in config.yml (server.port, default: 8080)"
	@echo "  PixelLab API: Set in config.yml (pixellab.api_key)"
