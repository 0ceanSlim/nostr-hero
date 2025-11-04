# Nostr Hero - Development Tools

Comprehensive guide to all development tools for Nostr Hero.

## Table of Contents

- [Overview](#overview)
- [Item Editor](#item-editor)
- [PixelLab Image Generator](#pixellab-image-generator)
- [Monster Manager](#monster-manager)
- [World Map Visualizer](#world-map-visualizer)
- [Tool Commands (Makefile)](#tool-commands-makefile)
- [Setup & Configuration](#setup--configuration)

---

## Overview

All tools are located in `docs/development/tools/` and can be run via the Makefile.

### Quick Access

```bash
# From anywhere in the repo
cd docs/development/tools

# See all available commands
make help
```

### Common Tools

| Tool | Purpose | Command |
|------|---------|---------|
| **Item Editor** | Edit items, generate images | `make run-item-editor` |
| **PixelLab Generator** | Batch generate item images | `make pixellab-gen` |
| **Monster Manager** | Manage creature stats | `python monster_manager.py` |
| **World Map Visualizer** | Visualize locations | `python world_map_visualizer.py` |

---

## Item Editor

**Full documentation**: [item-editor.md](./item-editor.md)

### Features

- Web-based GUI for editing items
- AI-powered image generation (PixelLab integration)
- Global ID refactoring (updates all references)
- Dynamic field management
- Smart tagging and filtering
- Image preview and history

### Quick Start

```bash
cd docs/development/tools
make run-item-editor
```

Opens browser to `http://localhost:8080`

**IMPORTANT**: Must run from repo root (Makefile handles this)

### Requirements

- Go 1.21+
- PixelLab API key (optional, for image generation)

### Configuration

Add to `config.yml` in repo root:

```yaml
pixellab:
  api_key: "your-api-key-here"
```

### Common Workflows

**Edit an item:**
1. Select item from list
2. Modify fields
3. Save

**Generate an image:**
1. Select item
2. Choose model (Bitforge or Pixflux)
3. Click "Generate Image"
4. Preview result
5. Click "Use This Image" or "Discard"

**Refactor item ID:**
1. Select item
2. Click "Refactor ID"
3. Enter new ID
4. Preview changes
5. Apply refactor

---

## PixelLab Image Generator

Batch generate pixel art images for game items using PixelLab API.

### Location

`docs/development/tools/pixellab-generator/`

### Build

```bash
make pixellab-gen
```

### Commands

**Check Balance:**
```bash
make pixellab-balance
```

**Dry Run (cost estimation):**
```bash
make pixellab-dry-run
```

**Generate by Category:**
```bash
make pixellab-weapons          # Generate 5 weapon images
make pixellab-gear             # Generate 5 gear images
make pixellab-starting-gear    # Generate all starting gear
make pixellab-ai-ready         # Generate all AI-ready items
```

**Generate by Type:**
```bash
make pixellab-simple-weapons
make pixellab-martial-weapons
make pixellab-ranged-weapons
make pixellab-adventuring-gear
make pixellab-tools
make pixellab-armor
```

**Generate Specific Item:**
```bash
make pixellab-item ITEM=longsword
```

### Output

Generated images saved to:
```
www/res/img/items/run_{type}_{model}_{timestamp}/png/{item-id}.png
```

### Models

- **Bitforge** (~$0.03/image) - Faster, cheaper
- **Pixflux** (~$0.05/image) - Higher quality

### Configuration

Requires PixelLab API key in `config.yml`:

```yaml
pixellab:
  api_key: "your-api-key-here"
```

### Prompt Generation

Uses item data in priority order:
1. `ai_description` field (best)
2. `description` field (fallback)
3. Auto-generated from item name + rarity (last resort)

---

## Monster Manager

Python tool for managing creature stat blocks.

### Location

`docs/development/tools/monster_manager.py`

### Usage

```bash
python monster_manager.py
```

### Features

- Add/edit/delete monsters
- Manage stat blocks
- Export to JSON
- Import from D&D databases

---

## World Map Visualizer

Python tool for visualizing game locations and connections.

### Location

`docs/development/tools/world_map_visualizer.py`

### Usage

```bash
python world_map_visualizer.py
```

### Features

- Visualize location graph
- Show connections between cities
- Export map images
- Validate location data

---

## Tool Commands (Makefile)

All commands from `docs/development/tools/Makefile`:

### Item Editor

```bash
make item-editor           # Build item editor
make run-item-editor       # Build and run item editor
make clean-item-editor     # Clean build files
```

### PixelLab Generator

```bash
# Setup
make pixellab-gen          # Build generator
make pixellab-balance      # Check API balance
make pixellab-dry-run      # Estimate costs

# Generation
make pixellab-weapons      # Generate weapons
make pixellab-gear         # Generate gear
make pixellab-armor        # Generate armor
make pixellab-item ITEM=id # Generate specific item

# Advanced
make pixellab-vectorize    # Convert PNG to SVG
make pixellab-scale-test   # Test scaled versions
```

### Utility

```bash
make clean-tools           # Clean all tool builds
make help                  # Show all commands
```

---

## Setup & Configuration

### Go Tools Setup

All Go tools require:

1. **Go 1.21+** installed
2. **Run from repo root** (paths are relative)

Example structure:
```bash
C:\code\nostr-hero\                    # Repo root
├── config.yml                          # Configuration
├── docs/development/tools/             # Tools directory
│   ├── Makefile                       # Build commands
│   ├── item-editor/                   # Item editor tool
│   └── pixellab-generator/            # Image generator
```

### Python Tools Setup

Python tools require:

```bash
pip install -r requirements.txt  # If requirements file exists
```

### PixelLab API Setup

1. Get API key from [PixelLab](https://pixellab.ai)
2. Add to `config.yml`:

```yaml
pixellab:
  api_key: "your-api-key-here"
```

3. Check balance:
```bash
cd docs/development/tools
make pixellab-balance
```

---

## Tool Development

### Adding a New Tool

1. Create directory: `docs/development/tools/your-tool/`
2. Add source code
3. Add Makefile target:

```makefile
your-tool:
	@echo "Building your tool..."
	cd your-tool && go build -o your-tool.exe .
```

4. Update this README
5. Update help command in Makefile

### Tool Conventions

- **Build output**: `{tool-name}.exe` in tool directory
- **Configuration**: Use root `config.yml`
- **Data paths**: Relative to repo root
- **Logging**: Use descriptive emoji-based logs

Example:
```go
log.Printf("✅ Success message")
log.Printf("❌ Error message")
log.Printf("⚠️ Warning message")
log.Printf("📂 Loading data...")
```

---

## Troubleshooting

### Item Editor won't start

**Problem**: `CreateFile docs/data/equipment/items: The system cannot find the path specified`

**Solution**: Run from repo root, not from tool directory:
```bash
cd C:\code\nostr-hero
.\docs\development\tools\item-editor\item-editor-gui.exe
```

Or use Makefile which handles this:
```bash
cd docs/development/tools
make run-item-editor
```

### PixelLab generation fails

**Problem**: API errors or unauthorized

**Solutions**:
1. Check API key in `config.yml`
2. Verify balance: `make pixellab-balance`
3. Check model name (bitforge or pixflux)

### Images don't show in Item Editor

**Problem**: Images not loading in preview

**Solutions**:
1. Check image exists: `www/res/img/items/{item-id}.png`
2. Verify image path in item JSON
3. Check browser console for 404 errors

### Makefile commands fail on Windows

**Problem**: Unix commands (like `rm`, `./`) don't work

**Solution**: We've updated Makefile to use Windows-compatible commands:
- Use `del` instead of `rm`
- Use backslashes in paths
- Don't use `./` prefix

---

## Best Practices

### When to Use Each Tool

| Task | Tool | Why |
|------|------|-----|
| Edit single item | Item Editor | Visual, easy, validates |
| Generate images for multiple items | PixelLab Generator | Batch processing, cost-efficient |
| Refactor item ID | Item Editor | Updates all references automatically |
| Add new monster | Monster Manager | Structured stat block entry |
| Visualize world map | World Map Visualizer | See location connections |

### Workflow Tips

1. **Always use Item Editor for ID changes** - It updates all references
2. **Use `ai_description` field** - Better image generation results
3. **Check balance before batch generation** - Avoid unexpected costs
4. **Test with dry-run first** - Estimate costs before committing
5. **Keep generated images in history** - Never lose good generations

### Development Workflow

```bash
# 1. Edit items
make run-item-editor

# 2. Generate missing images
make pixellab-balance
make pixellab-weapons

# 3. Test in game
cd ../../..
air

# 4. Commit changes
git add docs/data/equipment/items/
git add www/res/img/items/
git commit -m "Add new weapons with generated images"
```

---

## Resources

- [Item Editor Full Guide](./item-editor.md)
- [PixelLab API Docs](https://docs.pixellab.ai)
- [Development Guide](../readme.md)
- [Project Architecture](../../../CLAUDE.md)

---

## Development Notes & Changelog

**Note to Claude**: This section is for leaving organized notes as you help with development. Keep entries dated and categorized.

### 2025-11-03 - Item Editor Integration & Documentation Cleanup

**Item Editor Enhancements**:
- ✅ Integrated PixelLab image generation directly into item editor
- ✅ Added image preview and generation UI
- ✅ Implemented generate/replace/discard workflow
- ✅ Added cost tracking and balance display
- ✅ All generated images saved to `_history` folders
- ✅ Support for `ai_description` and `image` fields
- ✅ Fixed Windows path compatibility in Makefile
- ✅ Must run from repo root (paths are relative)

**Documentation Reorganization**:
- ✅ Created `docs/development/readme.md` as main development hub
- ✅ Created `docs/development/tools/readme.md` for all tools documentation
- ✅ Moved item-editor.md into tools/ directory
- ✅ Renamed all markdown files to lowercase
- ✅ Created `example.air.toml` for live-reload configuration
- ✅ Moved 11 outdated planning docs to `docs/draft/`
- ✅ All files now lowercase for consistency

**Draft Folder Consolidation**:
- ✅ Created `locations-plan.md` - Simplified current location needs (general store, vault, inn)
- ✅ Created `shops-and-services.md` - Consolidated all shop/tavern systems
- ✅ Created `unimplemented-ideas.md` - Preserved original concepts from planning.txt
- ✅ Created `xlsx-archive-notes.md` - Documentation for monsters/spells spreadsheets
- ✅ Updated `hunger-system-plan.md` - Marked as implemented, noted refactoring needed
- ✅ Renamed `house-of-keeping-system.md` to `vault-system.md` with simplified scope notes
- ✅ Removed 16 outdated files (gems.txt, character-creation-sequence.md, etc.)
- ✅ Final draft folder has 8 organized files vs. original 27+ scattered files

**Known Issues & Next Steps**:
- ⚠️ Item effects system needs standardization (currently inconsistent)
- ⚠️ Some item fields may not display properly in editor
- ⚠️ Need comprehensive item validation system
- 📋 Shop/inn/vault systems planned but not yet implemented
- 📋 Need to verify if all spells from spells.xlsx are in JSON format

**File Structure**:
```
docs/development/
├── readme.md                  # Main dev guide
├── example.air.toml          # Air config example
└── tools/
    ├── readme.md             # This file - all tools docs
    ├── item-editor.md        # Detailed item editor guide
    └── [tool subdirectories]
```

---

**Last Updated**: 2025-11-03
