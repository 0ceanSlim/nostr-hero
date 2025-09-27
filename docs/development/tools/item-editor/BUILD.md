# Item Editor - Build & Run Instructions

## 🏗️ Building from Repo Root

### Prerequisites
- Go 1.21 or later
- Git (for dependency management)
- Make (optional, for easier commands)

### Using Make (Recommended)

```bash
# From the repo root directory
make run-item-editor    # Build and run
make item-editor        # Just build
make clean-item-editor  # Clean build files
make help              # Show available commands
```

### Manual Build Commands

```bash
# From the repo root directory

# Build the item editor
cd docs/development/tools/item-editor
go mod tidy
go build -o ../../../../item-editor-gui.exe .
cd ../../../../

# The executable will be created as: item-editor-gui.exe
```

### Quick Build Script

You can also use this one-liner from the repo root:

```bash
cd docs/development/tools/item-editor && go build -o ../../../../item-editor-gui.exe . && cd ../../../../
```

## 🚀 Running the Tool

### From Repo Root (Required)

```bash
# Must be run from the repository root directory
cd /path/to/nostr-hero
./item-editor-gui.exe
```

**Important**: The tool MUST be run from the repo root because it uses relative paths to access:
- `docs/data/equipment/items/` - All item JSON files
- `docs/data/character/starting-gear.json` - Character starting equipment

### What Happens When You Run

1. **Web server starts** on `http://localhost:8080`
2. **Browser auto-opens** to the item editor interface
3. **Items are loaded** from the game data files
4. **Terminal-themed GUI** appears ready for editing

## 🎯 Quick Start

```bash
# Clone/navigate to repo
cd /path/to/nostr-hero

# Using Make (recommended)
make run-item-editor

# Or manually build and run
cd docs/development/tools/item-editor && go build -o ../../../../item-editor-gui.exe . && cd ../../../../
./item-editor-gui.exe
```

## 🔧 Development Workflow

### Making Changes to the Tool

1. **Edit source code** in `docs/development/tools/item-editor/main.go`
2. **Rebuild from repo root**:
   ```bash
   cd docs/development/tools/item-editor && go build -o ../../../../item-editor-gui.exe . && cd ../../../../
   ```
3. **Test by running** `./item-editor-gui.exe`

### File Structure

```
nostr-hero/
├── item-editor-gui.exe          # Built executable (run from here)
├── docs/
│   └── development/
│       └── tools/
│           └── item-editor/
│               ├── main.go      # Source code
│               ├── go.mod       # Go module
│               ├── go.sum       # Dependencies
│               ├── README.md    # Feature documentation
│               └── BUILD.md     # This file
└── docs/data/
    ├── equipment/items/         # Item data files (read/write)
    └── character/starting-gear.json  # Starting gear data (read/write)
```

## 🛠️ Troubleshooting

### "No such file or directory" errors
- **Cause**: Running from wrong directory
- **Fix**: Always run from repo root (`/c/code/nostr-hero`)

### "Failed to load items" error
- **Cause**: Missing data files or wrong working directory
- **Fix**: Ensure you're in repo root and data files exist

### Browser doesn't open automatically
- **Solution**: Manually open `http://localhost:8080`

### Port 8080 already in use
- **Solution**: Kill any existing processes using port 8080 or restart the tool

## 📦 Dependencies

The tool uses these Go modules (managed automatically):
- `github.com/gorilla/mux` - HTTP routing
- Standard library for JSON, file I/O, etc.

## 🔄 Updating

When you make changes to the codebase:

1. **Always rebuild** before testing
2. **Test thoroughly** with actual data
3. **Backup important data** before major refactoring operations
4. **Use git commits** to track changes to game data

## ⚡ Quick Reference

| Action | Command |
|--------|---------|
| Build | `cd docs/development/tools/item-editor && go build -o ../../../../item-editor-gui.exe . && cd ../../../../` |
| Run | `./item-editor-gui.exe` |
| Access | `http://localhost:8080` |
| Working Dir | Must be repo root directory |