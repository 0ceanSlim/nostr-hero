# CODEX Makefile

.PHONY: build run clean help

# Default target
all: build

# Build the CODEX executable
build:
	@echo "Building CODEX..."
	@go build -o codex.exe main.go
	@echo "Build complete! Run with: ./codex.exe"

# Build and run
run: build
	@echo "Starting CODEX..."
	@./codex.exe

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f codex.exe codex
	@echo "Clean complete!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@echo "Dependencies installed!"

# Help
help:
	@echo "CODEX - Content Organization & Data Entry eXperience"
	@echo ""
	@echo "Available targets:"
	@echo "  make build    - Build CODEX executable"
	@echo "  make run      - Build and run CODEX"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make deps     - Install Go dependencies"
	@echo "  make help     - Show this help message"
