# Nostr Hero - Item Editor Build Commands

.PHONY: item-editor run-item-editor clean-item-editor help

# Build the item editor
item-editor:
	@echo "Building item editor..."
	cd docs/development/tools/item-editor && go build -o ../../../../item-editor-gui.exe .
	@echo "Item editor built successfully: item-editor-gui.exe"

# Build and run the item editor
run-item-editor: item-editor
	@echo "Starting item editor..."
	./item-editor-gui.exe

# Clean item editor build
clean-item-editor:
	@echo "Cleaning item editor..."
	rm -f item-editor-gui.exe
	@echo "Item editor cleaned"

# Show available commands
help:
	@echo "Available commands:"
	@echo "  make item-editor      - Build the item editor"
	@echo "  make run-item-editor  - Build and run the item editor"
	@echo "  make clean-item-editor - Remove item editor executable"
	@echo "  make help            - Show this help message"

# Default target
.DEFAULT_GOAL := help