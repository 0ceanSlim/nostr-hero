# Nostr Hero - Item Editor Build Commands

.PHONY: item-editor run-item-editor clean-item-editor item-sprite-gen png-to-svg clean-tools generate-items test-items help

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

# Build item sprite generator
item-sprite-gen:
	@echo "Building item sprite generator..."
	cd docs/development/tools/item-sprite-generator && go build -o ../../../../item-sprite-gen.exe .
	@echo "Item sprite generator built: item-sprite-gen.exe"

# Build PNG to SVG converter
png-to-svg:
	@echo "Building PNG to SVG converter..."
	cd docs/development/tools/png-to-svg && go build -o ../../../../png-to-svg.exe .
	@echo "PNG to SVG converter built: png-to-svg.exe"

# Clean all tools
clean-tools: clean-item-editor
	@echo "Cleaning all tools..."
	rm -f item-sprite-gen.exe png-to-svg.exe
	@echo "All tools cleaned"

# Test with first 10 items
test-items:
	@echo "Building test generator..."
	cd docs/development/tools/item-sprite-generator && go build -o ../../../../item-sprite-test.exe test.go lib.go
	@echo ""
	@echo "=== Test run: generating 10 items ==="
	@echo "Make sure ComfyUI is running on http://localhost:8188"
	@echo ""
	./item-sprite-test.exe -count 10
	@echo ""
	@echo "Check output at: C:/code/comfyui/output/nostr_items/"

# Complete workflow: build tools, generate sprites, convert to SVG
generate-items: item-sprite-gen png-to-svg
	@echo "=== Starting item generation workflow ==="
	@echo ""
	@echo "Step 1: Generating sprites via ComfyUI..."
	@echo "Make sure ComfyUI is running on http://localhost:8188"
	@echo ""
	./item-sprite-gen.exe
	@echo ""
	@echo "Step 2: Converting PNGs to SVG..."
	@echo ""
	./png-to-svg.exe -input "C:/code/comfyui/output/nostr_items" -output "./assets/items/svg"
	@echo ""
	@echo "=== Item generation complete! ==="

# Show available commands
help:
	@echo "Available commands:"
	@echo "  make test-items        - Generate first 10 items (for testing/tweaking)"
	@echo "  make generate-items    - Build tools & run complete sprite generation workflow"
	@echo "  make item-editor       - Build the item editor"
	@echo "  make run-item-editor   - Build and run the item editor"
	@echo "  make item-sprite-gen   - Build item sprite generator"
	@echo "  make png-to-svg        - Build PNG to SVG converter"
	@echo "  make clean-item-editor - Remove item editor executable"
	@echo "  make clean-tools       - Remove all tool executables"
	@echo "  make help              - Show this help message"

# Default target
.DEFAULT_GOAL := help