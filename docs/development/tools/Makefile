# Nostr Hero - Development Tools Build Commands
# Run from: docs/development/tools/
# Usage: cd docs/development/tools && make <command>

.PHONY: item-editor run-item-editor clean-item-editor item-sprite-gen png-to-svg pixellab-gen clean-tools generate-items test-items test-pixellab pixellab-balance pixellab-dry-run pixellab-vectorize pixellab-scale-test pixellab-weapons pixellab-gear pixellab-starting-gear pixellab-ai-ready help

# Build the item editor
item-editor:
	@echo "Building item editor..."
	cd item-editor && go mod tidy && go build -o item-editor-gui.exe .
	@echo "Item editor built: item-editor/item-editor-gui.exe"

# Build and run the item editor
run-item-editor: item-editor
	@echo "Starting item editor..."
	@echo "Running from repo root..."
	cd ../../../ && docs\development\tools\item-editor\item-editor-gui.exe

# Clean item editor build
clean-item-editor:
	@echo "Cleaning item editor..."
	-del /Q item-editor\item-editor-gui.exe 2>nul || echo "No file to clean"
	@echo "Item editor cleaned"

# Build item sprite generator
item-sprite-gen:
	@echo "Building item sprite generator..."
	cd item-sprite-generator && go build -o item-sprite-gen.exe .
	@echo "Item sprite generator built: item-sprite-generator/item-sprite-gen.exe"

# Build PNG to SVG converter
png-to-svg:
	@echo "Building PNG to SVG converter..."
	cd png-to-svg && go build -o png-to-svg.exe .
	@echo "PNG to SVG converter built: png-to-svg/png-to-svg.exe"

# Build PixelLab generator
pixellab-gen:
	@echo "Building PixelLab generator..."
	cd pixellab-generator && go mod tidy && go build -o pixellab-gen.exe .
	@echo "PixelLab generator built: pixellab-generator/pixellab-gen.exe"

# Clean all tools
clean-tools: clean-item-editor
	@echo "Cleaning all tools..."
	rm -f item-sprite-generator/item-sprite-gen.exe png-to-svg/png-to-svg.exe pixellab-generator/pixellab-gen.exe
	@echo "All tools cleaned"

# Test with first 10 items
test-items: item-sprite-gen
	@echo "Building test generator..."
	cd item-sprite-generator && go build -o item-sprite-test.exe test.go lib.go
	@echo ""
	@echo "=== Test run: generating 10 items ==="
	@echo "Make sure ComfyUI is running on http://localhost:8188"
	@echo ""
	cd item-sprite-generator && ./item-sprite-test.exe -count 10
	@echo ""
	@echo "Check output at: C:/code/comfyui/output/nostr_items/"

# Check PixelLab balance
pixellab-balance: pixellab-gen
	@echo "Checking PixelLab balance..."
	cd pixellab-generator && ./pixellab-gen.exe balance

# PixelLab dry run for cost estimation
pixellab-dry-run: pixellab-gen
	@echo "Running PixelLab cost estimation..."
	cd pixellab-generator && ./pixellab-gen.exe dry-run --count 2 --model bitforge

# Test PixelLab with 2 items
test-pixellab: pixellab-gen
	@echo "=== PixelLab Test Run: 2 items ==="
	@echo "Checking balance first..."
	cd pixellab-generator && ./pixellab-gen.exe balance
	@echo ""
	@echo "Running dry-run estimation..."
	cd pixellab-generator && ./pixellab-gen.exe dry-run --count 2 --model bitforge
	@echo ""
	@echo "Generating 2 test items..."
	cd pixellab-generator && ./pixellab-gen.exe generate --count 2 --model bitforge
	@echo ""
	@echo "Check output at: ../../../www/res/img/items/"

# Convert PNG images to pixel-perfect SVG
pixellab-vectorize: pixellab-gen
	@echo "Converting PNG images to pixel-perfect SVG..."
	cd pixellab-generator && ./pixellab-gen.exe vectorize

# Create scaled test images from SVG
pixellab-scale-test: pixellab-gen
	@echo "Creating scaled test images..."
	cd pixellab-generator && ./pixellab-gen.exe scale-test

# Generate images by category
pixellab-weapons: pixellab-gen
	@echo "=== Generating weapon images ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-category weapons --model bitforge --count 5

pixellab-gear: pixellab-gen
	@echo "=== Generating adventuring gear images ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-category gear --model bitforge --count 5

pixellab-starting-gear: pixellab-gen
	@echo "=== Generating starting gear images ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-category starting-gear --model bitforge

pixellab-ai-ready: pixellab-gen
	@echo "=== Generating AI-ready items ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-category ai-ready --model bitforge

# Generate by item type
pixellab-simple-weapons: pixellab-gen
	@echo "=== Generating Simple Melee Weapons ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-type "Simple Melee Weapons" --model bitforge

pixellab-martial-weapons: pixellab-gen
	@echo "=== Generating Martial Melee Weapons ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-type "Martial Melee Weapons" --model bitforge

pixellab-ranged-weapons: pixellab-gen
	@echo "=== Generating Simple Ranged Weapons ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-type "Simple Ranged Weapons" --model bitforge

pixellab-adventuring-gear: pixellab-gen
	@echo "=== Generating Adventuring Gear ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-type "Adventuring Gear" --model bitforge

pixellab-tools: pixellab-gen
	@echo "=== Generating Tools ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-type "Tools" --model bitforge

pixellab-armor: pixellab-gen
	@echo "=== Generating Armor ==="
	cd pixellab-generator && ./pixellab-gen.exe generate-type "Armor" --model bitforge

# Generate specific item by ID
pixellab-item: pixellab-gen
	@echo "=== Generating specific item ==="
	@echo "Usage: make pixellab-item ITEM=club"
	@if [ -z "$(ITEM)" ]; then echo "Error: Please specify ITEM=<item-id>"; exit 1; fi
	cd pixellab-generator && ./pixellab-gen.exe generate-id $(ITEM) --model bitforge

# Complete workflow: build tools, generate sprites, convert to SVG
generate-items: item-sprite-gen png-to-svg
	@echo "=== Starting item generation workflow ==="
	@echo ""
	@echo "Step 1: Generating sprites via ComfyUI..."
	@echo "Make sure ComfyUI is running on http://localhost:8188"
	@echo ""
	cd item-sprite-generator && ./item-sprite-gen.exe
	@echo ""
	@echo "Step 2: Converting PNGs to SVG..."
	@echo ""
	cd png-to-svg && ./png-to-svg.exe -input "C:/code/comfyui/output/nostr_items" -output "../../../assets/items/svg"
	@echo ""
	@echo "=== Item generation complete! ==="

# Show available commands
help:
	@echo "Nostr Hero - Development Tools"
	@echo "Run from: docs/development/tools/"
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo "Basic commands:"
	@echo "  make test-items         - Generate first 10 items (for testing/tweaking)"
	@echo "  make generate-items     - Build tools & run complete sprite generation workflow"
	@echo "  make item-editor        - Build the item editor"
	@echo "  make run-item-editor    - Build and run the item editor"
	@echo ""
	@echo "PixelLab commands:"
	@echo "  make pixellab-gen       - Build PixelLab generator"
	@echo "  make pixellab-balance   - Check PixelLab API balance"
	@echo "  make pixellab-dry-run   - Estimate costs for 2 items"
	@echo "  make test-pixellab      - Test PixelLab with 2 items"
	@echo "  make pixellab-vectorize - Convert PNG to pixel-perfect SVG"
	@echo "  make pixellab-scale-test- Create scaled test images"
	@echo ""
	@echo "Category generation:"
	@echo "  make pixellab-weapons   - Generate 5 weapon images"
	@echo "  make pixellab-gear      - Generate 5 adventuring gear images"
	@echo "  make pixellab-starting-gear - Generate all starting gear"
	@echo "  make pixellab-ai-ready  - Generate all AI-ready items"
	@echo ""
	@echo "Type-specific generation:"
	@echo "  make pixellab-simple-weapons    - Generate Simple Melee Weapons"
	@echo "  make pixellab-martial-weapons   - Generate Martial Melee Weapons"
	@echo "  make pixellab-ranged-weapons    - Generate Simple Ranged Weapons"
	@echo "  make pixellab-adventuring-gear  - Generate Adventuring Gear"
	@echo "  make pixellab-tools             - Generate Tools"
	@echo "  make pixellab-armor             - Generate Armor"
	@echo "  make pixellab-item ITEM=<id>    - Generate specific item by ID"
	@echo ""
	@echo "Other commands:"
	@echo "  make clean-item-editor  - Remove item editor executable"
	@echo "  make clean-tools        - Remove all tool executables"
	@echo "  make help               - Show this help message"

# Default target
.DEFAULT_GOAL := help
