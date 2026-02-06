# Pubkey Quest - Development Setup
# Usage: make setup

.PHONY: setup config deps migrate swagger build clean help

# Complete first-time setup
setup: config deps migrate swagger
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit config.yml to set your port and enable debug_mode"
	@echo "  2. Run 'air' to start the dev server with live reload"
	@echo ""

# Copy example config (if not already present)
config:
	@if [ ! -f ../../config.yml ]; then \
		cp examples/example.config.yml ../../config.yml && echo "✅ Created config.yml"; \
	else \
		echo "⚠️  config.yml already exists (skipping)"; \
	fi

# Install dependencies
deps:
	@echo "📦 Installing npm dependencies..."
	@cd ../.. && npm install
	@echo "📦 Installing Go dependencies..."
	@cd ../.. && go mod download
	@echo "✅ Dependencies installed"

# Build codex and run database migration
migrate:
	@echo "🔨 Building codex..."
	@cd ../.. && go build -o codex ./cmd/codex
	@echo "🔄 Running database migration..."
	@cd ../.. && ./codex -migrate
	@echo "✅ Migration complete"

# Generate Swagger API docs
swagger:
	@echo "📝 Generating Swagger API docs..."
	@cd ../.. && swag init -g cmd/server/api/routes.go -o docs/api/swagger --parseDependency
	@echo "✅ Swagger docs generated"

# Full production build (frontend + swagger + codex + migration + server)
build: deps swagger migrate
	@echo "🔨 Building frontend..."
	@cd ../.. && npm run build
	@echo "🔨 Building server binary..."
	@cd ../.. && go build -o pubkey-quest ./cmd/server
	@echo "✅ Build complete"

# Remove generated files
clean:
	@echo "🧹 Cleaning generated files..."
	@rm -f ../../www/game.db
	@rm -rf ../../www/dist
	@rm -rf ../../docs/api/swagger
	@rm -f ../../pubkey-quest ../../pubkey-quest.exe
	@rm -f ../../codex ../../codex.exe
	@echo "✅ Clean complete"

# Show help
help:
	@echo "Pubkey Quest - Development Setup"
	@echo ""
	@echo "  make setup    - First-time setup (config + deps + migrate + swagger)"
	@echo "  make config   - Copy example config.yml (if missing)"
	@echo "  make deps     - Install npm + Go dependencies"
	@echo "  make migrate  - Build codex + run database migration"
	@echo "  make swagger  - Generate Swagger API docs"
	@echo "  make build    - Full production build"
	@echo "  make clean    - Remove generated files"
	@echo "  make help     - Show this message"
	@echo ""

.DEFAULT_GOAL := help
