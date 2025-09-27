# Item Editor GUI - Terminal Theme

A powerful web-based item editor with a dark terminal theme and full global ID refactoring capabilities.

## 🎯 Features

### ✨ **Modern Web GUI**
- **Dark Terminal Theme** - Beautiful dark UI with terminal aesthetics
- **Monospace Fonts** - Consolas/Monaco for that authentic terminal feel
- **Click-to-Edit** - Point and click interface, no terminal commands needed
- **Responsive Design** - Works on any screen size
- **Auto-open Browser** - Launches automatically when started

### 🔧 **Core Functionality**
- **Item List & Search** - Browse all items with real-time search
- **Visual Validation** - See validation status (✓/✗) at a glance
- **Full Item Editing** - Edit all item properties in a clean form
- **Auto-save** - Changes saved immediately to JSON files

### 🚀 **Global ID Refactoring**
- **Cross-Reference Scanning** - Finds ALL references across the codebase
- **Preview Mode** - Shows exactly what will change before applying
- **One-Click Refactor** - Updates everything automatically
- **Safe Operations** - All changes happen atomically

## 🖥️ Usage

```bash
cd docs/development/tools/item-editor
./item-editor-gui.exe
```

**The tool will:**
1. Start a web server on `http://localhost:8080`
2. Automatically open your browser
3. Load all items from the game data

## 🎨 Interface

### **Left Sidebar**
- **Search box** - Filter items by name or ID
- **Item list** - All items with validation status
  - ✓ = ID matches filename
  - ✗ = ID mismatch (needs fixing)
- **Validate All** - Check all items for consistency
- **Refresh** - Reload items from disk

### **Main Panel**
- **Item Details Form** - Edit all item properties
  - ID, Name, Description
  - Price, Type, Weight, Stack
  - Gear Slot, Rarity
- **Save Changes** - Update the item file
- **Refactor ID** - Global ID change with preview

### **Status Bar**
- Real-time feedback on operations
- Success/error messages
- Item counts and validation results

## 🔄 Global ID Refactoring Workflow

1. **Select item** from the list
2. **Click "Refactor ID"** button
3. **Enter new ID** in the dialog
4. **Click "Preview Changes"** to see what will update
5. **Review the preview** showing:
   - File rename operation
   - All references that will be updated
   - Locations in starting-gear.json and pack contents
6. **Click "Apply Refactor"** to execute all changes

## 🎯 What Gets Updated

When you change an item ID, the tool automatically updates:

- ✅ **Item filename** - `old-id.json` → `new-id.json`
- ✅ **Item.id field** - Internal JSON property
- ✅ **Starting gear** - All character class equipment references
- ✅ **Pack contents** - All pack files containing the item
- ✅ **Nested structures** - Complex weapon choices, bundles, etc.

## 🛡️ Safety Features

- **Preview before apply** - Always shows what will change
- **Atomic operations** - All changes succeed or all fail
- **Real-time validation** - Immediate feedback on issues
- **Auto-browser opening** - No manual URL entry needed

## 🎨 Terminal Theme Colors

- **Background**: `#121212` (Dark charcoal)
- **Sidebar**: `#1e1e1e` (Darker gray)
- **Text**: `#ffffff` (Pure white)
- **Accent**: `#50fa7b` (Terminal green)
- **Selection**: `#44475a` (Purple-gray)
- **Error**: `#ff5555` (Red)
- **Warning**: `#f1fa8c` (Yellow)

## 🔧 Technical Details

- **Built with Go** - Consistent with main project
- **Web server** - Gorilla Mux router
- **Single file** - Everything embedded in one executable
- **Auto-browser** - Cross-platform browser opening
- **REST API** - Clean separation of frontend/backend

## 🚀 Why This Approach?

- **Familiar interface** - Point and click like you're used to
- **Terminal aesthetics** - Maintains the developer feel you want
- **No dependencies** - Just run the executable
- **Cross-platform** - Works on Windows, Mac, Linux
- **Modern** - All the benefits of web UI technology

This solves the ID consistency problem once and for all with a tool that's both powerful and easy to use!