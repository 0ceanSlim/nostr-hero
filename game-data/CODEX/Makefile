# CODEX Makefile
# NOTE: This Makefile should be run from the CODEX directory (game-data/CODEX/)
# You can also run it from the project root using: make -f game-data/CODEX/Makefile <target>

.PHONY: all build build-js build-go run clean clean-js clean-go deps deps-js deps-go migrate help

# Default target - build everything
all: build

# Build everything (JS + Go)
build: build-js build-go

# Build JavaScript/CSS assets
build-js:
	@echo "📦 Building CODEX frontend..."
	@npm run build
	@echo "✅ Frontend build complete! (output: www/dist/codex/)"

# Build Go executable
build-go:
	@echo "🔨 Building CODEX server..."
	@go build -o ../../codex.exe .
	@echo "✅ Server build complete! (executable: ./codex.exe)"

# Build and run CODEX
run: build
	@echo "🚀 Starting CODEX..."
	@cd ../.. && codex.exe

# Run with auto-restart on changes (requires Air)
dev:
	@echo "🔥 Starting CODEX in development mode..."
	@air

# Clean all build artifacts
clean: clean-js clean-go
	@echo "✅ All build artifacts cleaned!"

# Clean JavaScript build output
clean-js:
	@echo "🧹 Cleaning CODEX frontend build..."
	@rm -rf ../../www/dist/codex
	@rm -rf dist

# Clean Go executable
clean-go:
	@echo "🧹 Cleaning CODEX server executable..."
	@rm -f ../../codex.exe ../../codex

# Install all dependencies
deps: deps-js deps-go

# Install JavaScript dependencies
deps-js:
	@echo "📦 Installing CODEX frontend dependencies..."
	@npm install
	@echo "✅ Frontend dependencies installed!"

# Install Go dependencies
deps-go:
	@echo "📦 Installing CODEX server dependencies..."
	@go mod download
	@echo "✅ Server dependencies installed!"

# Run database migration
migrate:
	@echo "🔄 Running database migration..."
	@cd ../.. && codex.exe --migrate
	@echo "✅ Migration complete!"

# Help
help:
	@echo "🎯 CODEX - Content Organization & Data Entry eXperience"
	@echo ""
	@echo "📋 Available targets:"
	@echo "  make all        - Build everything (JS + Go)"
	@echo "  make build      - Build everything (JS + Go)"
	@echo "  make build-js   - Build frontend only (→ www/dist/codex/)"
	@echo "  make build-go   - Build server only (→ ./codex.exe)"
	@echo "  make run        - Build and run CODEX"
	@echo "  make dev        - Run with auto-restart (requires Air)"
	@echo "  make clean      - Remove all build artifacts"
	@echo "  make clean-js   - Remove frontend build"
	@echo "  make clean-go   - Remove server executable"
	@echo "  make deps       - Install all dependencies"
	@echo "  make deps-js    - Install frontend dependencies"
	@echo "  make deps-go    - Install server dependencies"
	@echo "  make migrate    - Run database migration"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "🔧 Configuration:"
	@echo "  Config file: codex-config.yml (copy from codex-config.example.yml)"
	@echo "  Port: Set in codex-config.yml (server.port, default: 8080)"
	@echo "  PixelLab API: Set in codex-config.yml (pixellab.api_key)"
