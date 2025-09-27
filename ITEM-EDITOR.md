# 🔧 Item Editor Tool

A powerful web-based item editor with terminal theming, dynamic field management, and global ID refactoring capabilities.

## 🚀 Quick Start

### Using Make (Recommended)
```bash
# Build and run from repo root
make run-item-editor
```

### Manual Build
```bash
# Navigate to repo root
cd /path/to/nostr-hero

# Build the tool (first time or after changes)
cd docs/development/tools/item-editor && go build -o ../../../../item-editor-gui.exe . && cd ../../../../

# Run the tool
./item-editor-gui.exe
```

## ✨ Features

- **🎨 Terminal Theme** - Dark interface with green accents
- **🖱️ Click-to-Edit** - Modern web interface
- **🔍 Search & Filter** - Find items by name, type, or tag
- **📝 Dynamic Fields** - Show only fields that exist, add/remove any field
- **🏷️ Smart Tags** - Dropdown with existing tags, filter by tags
- **📔 Internal Notes** - Separate from game descriptions
- **✅ Validation** - Visual ID consistency checking
- **🔄 Global Refactoring** - Change IDs everywhere automatically
- **📋 Preview Mode** - See what will change before applying
- **📐 Structured Data** - Enforces consistent field ordering in JSON

## 🎯 Main Use Cases

### 1. **Dynamic Field Management**
- Only shows fields that exist in each item
- Add any new field with "Add Field" button
- Remove fields with ✗ button next to each field
- Enforces standard field ordering in saved JSON

### 2. **Fix ID Inconsistencies**
- Items show ✗ if `item.id` ≠ `filename`
- Click "Validate All" to see all issues
- Edit individual items to fix

### 3. **Global ID Refactoring**
- Select item → "Refactor ID" → Enter new ID
- Preview shows all references that will update
- One-click applies changes everywhere

### 4. **Smart Filtering & Tagging**
- Filter by item type or tag
- Search by name, ID, or description
- Add tags with autocomplete dropdown
- Internal notes separate from game descriptions

## 🔄 Global Refactoring Process

1. **Select item** from list
2. **Click "Refactor ID"** button
3. **Enter new ID** in dialog
4. **Click "Preview Changes"**
5. **Review what will be updated**:
   - File rename (`old.json` → `new.json`)
   - Starting gear references
   - Pack contents references
6. **Click "Apply Refactor"** to execute

## 📁 What Gets Updated

✅ **Item filename** - Renamed to match new ID
✅ **Item.id field** - Updated in JSON
✅ **Starting gear** - All character class equipment
✅ **Pack contents** - All pack files containing the item
✅ **Complex structures** - Nested weapon choices, bundles

## 🛡️ Safety Features

- **Preview before apply** - Always see what will change
- **Atomic operations** - All changes succeed or all fail
- **Real-time validation** - Immediate feedback
- **Path validation** - Ensures files exist before operations

## 📝 Development

For build instructions and development workflow, see:
- `docs/development/tools/item-editor/BUILD.md` - Build & run instructions
- `docs/development/tools/item-editor/README.md` - Technical details

## 🔧 Requirements

- **Go 1.21+** for building
- **Web browser** for interface
- **Make** (optional, for easier building)
- **Run from repo root** (file paths are relative)

## 🎨 Interface

- **Left Sidebar**: Item list with search and validation
- **Main Panel**: Item editing form with save/refactor buttons
- **Status Bar**: Real-time feedback and operation status
- **Modals**: Refactoring dialogs with preview

This tool solves the ID consistency problems by automating cross-reference updates across all game data files.