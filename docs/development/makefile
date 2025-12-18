# Nostr Hero - Development Setup
# First-time setup for developers
# Usage: make setup

.PHONY: setup setup-configs install-deps clean-configs help

# Complete first-time setup
setup: setup-configs install-deps
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit config.yml to set your port (default: 8080)"
	@echo "  2. Run 'air' to start the Go backend with live reload"
	@echo "     - Air will automatically run 'npm run build' on every Go file change"
	@echo "  3. Or run 'npm run dev' in another terminal for faster frontend development"
	@echo "     - Provides instant hot module replacement for JS/CSS changes"
	@echo ""

# Copy example configs to root
setup-configs:
	@echo "Setting up configuration files..."
	@if [ ! -f config.yml ]; then \
		cp docs/development/examples/example.config.yml config.yml && echo "  ✅ Created config.yml"; \
	else \
		echo "  ⚠️  config.yml already exists (skipping)"; \
	fi
	@if [ ! -f .air.toml ]; then \
		cp docs/development/examples/example.air.toml .air.toml && echo "  ✅ Created .air.toml"; \
	else \
		echo "  ⚠️  .air.toml already exists (skipping)"; \
	fi
	@if [ ! -f package.json ]; then \
		cp docs/development/examples/example.package.json package.json && echo "  ✅ Created package.json"; \
	else \
		echo "  ⚠️  package.json already exists (skipping)"; \
	fi
	@if [ ! -f vite.config.js ]; then \
		cp docs/development/examples/example.vite.config.js vite.config.js && echo "  ✅ Created vite.config.js"; \
	else \
		echo "  ⚠️  vite.config.js already exists (skipping)"; \
	fi
	@if [ ! -f postcss.config.js ]; then \
		cp docs/development/examples/example.postcss.config.js postcss.config.js && echo "  ✅ Created postcss.config.js"; \
	else \
		echo "  ⚠️  postcss.config.js already exists (skipping)"; \
	fi

# Install npm dependencies
install-deps:
	@echo ""
	@echo "Installing npm dependencies..."
	@npm install
	@echo "  ✅ npm packages installed"

# Remove local config files (not examples)
clean-configs:
	@echo "Removing local configuration files..."
	@rm -f .air.toml package.json package-lock.json vite.config.js postcss.config.js
	@rm -rf node_modules
	@echo "  ✅ Local configs removed (examples preserved)"

# Show help
help:
	@echo "Nostr Hero - Development Setup"
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo "  make setup         - Complete first-time setup (configs + dependencies)"
	@echo "  make setup-configs - Copy example configs to root (without installing deps)"
	@echo "  make install-deps  - Install npm dependencies only"
	@echo "  make clean-configs - Remove local configs (keeps examples)"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Example configs are in docs/development/examples/"
	@echo "  - example.air.toml         → .air.toml (Go live reload config)"
	@echo "  - example.config.yml       → config.yml (Server config)"
	@echo "  - example.package.json     → package.json (npm dependencies)"
	@echo "  - example.vite.config.js   → vite.config.js (Vite bundler config)"
	@echo "  - example.postcss.config.js → postcss.config.js (PostCSS/Tailwind)"
	@echo ""
	@echo "After setup, run:"
	@echo "  air              - Start Go backend with live reload"
	@echo "                     (automatically runs npm build on Go file changes)"
	@echo "  npm run dev      - Start Vite dev server for frontend development"
	@echo "                     (hot module replacement, faster than air)"
	@echo "  npm run build    - Build production bundles manually"
	@echo ""

.DEFAULT_GOAL := help
