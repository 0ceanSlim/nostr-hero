# 🔧 Item Editor Tool

A powerful web-based item editor with terminal theming, dynamic field management, global ID refactoring capabilities, and **integrated AI image generation**.

## 🚀 Quick Start

### Using Make (Recommended)
```bash
# From repo root, navigate to tools directory
cd docs/development/tools

# Build and run the item editor
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
- **🖼️ AI Image Generation** - Generate pixel art images directly in the editor
- **💰 Cost Tracking** - Real-time PixelLab API balance and cost per generation
- **🎯 Smart Prompts** - Uses `ai_description` field or auto-generates from item data
- **📜 Image History** - Keeps all generated images in `_history` folders
- **✓ Replace/Discard** - Preview and choose whether to use generated images

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

## 🎨 AI Image Generation Workflow

### Setup
1. Add your PixelLab API key to `config.yml`:
```yaml
pixellab:
  api_key: "your-api-key-here"
```

### Generating Images
1. **Select an item** from the list
2. **View current image** (if exists) in the preview section
3. **Choose AI model**:
   - `Bitforge` (~$0.03/image) - Faster, cheaper
   - `Pixflux` (~$0.05/image) - Higher quality
4. **Click "Generate Image"** button
5. **Wait for generation** (10-30 seconds)
6. **Preview the result**:
   - View generated image
   - See the prompt used
   - Check generation cost
7. **Choose action**:
   - ✓ **Use This Image** - Copies to `www/res/img/items/{id}.png` and saves to history
   - ✗ **Discard** - Delete and try again

### Prompt Priority
The generator uses the best available description:
1. **`ai_description`** field (best - specifically for image gen)
2. **`description`** field (fallback - game description)
3. **Auto-generated** from item name + rarity (last resort)

### Image Storage
- **Active images**: `www/res/img/items/{item-id}.png`
- **Generation history**: `www/res/img/items/_history/{item-id}/{timestamp}_{model}.png`

### Cost Tracking
- View current balance in the editor
- See cost per generation
- Balance updates automatically after each generation

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

### Required
- **Go 1.21+** for building
- **Web browser** for interface
- **Run from repo root** (file paths are relative)

### Optional
- **Make** (for easier building)
- **PixelLab API key** (for image generation - add to `config.yml`)

## 🎨 Interface

- **Left Sidebar**: Item list with search and validation
- **Main Panel**: Item editing form with save/refactor buttons
- **Status Bar**: Real-time feedback and operation status
- **Modals**: Refactoring dialogs with preview

## 💡 Benefits

This enhanced tool provides:
- **Unified workflow** - Edit items and generate images in one place
- **ID consistency** - Automated cross-reference updates across all game data files
- **Cost efficiency** - See costs before generating, choose models based on budget
- **Quality control** - Preview and approve images before committing
- **History tracking** - Never lose a generated image, all saved to history
- **Smart prompting** - Leverages `ai_description` fields for better results

## 📝 Notes

- The editor works without PixelLab API - image generation is optional
- Generated images are saved to history before being used
- The refactoring system still works independently of image generation
- All `ai_description` and `image` fields are now properly supported
- **IMPORTANT**: Must be run from repo root (`C:\code\nostr-hero`) - all paths are relative

## ⚠️ Known Issues & Future Improvements

### Current Limitations

1. **Field Visibility Issues**
   - Some item fields don't show in the editor
   - Dynamic field rendering may miss certain field types

2. **Unnecessary Fields**
   - Many items have legacy D&D fields they don't need
   - Items inherited from D&D base need cleanup

3. **Item Structure Inconsistencies**
   - Food/consumable effects not standardized
   - Stat modification fields vary between items (e.g., rations have effects that need restructuring)
   - No unified effects system

4. **Validation Gaps**
   - No comprehensive item validation across all items
   - No type-specific schema validation
   - No enforcement of required fields per item type

### Planned Improvements

#### 1. Item Type Schemas
Define proper schemas for each item category:
- **Weapons**: damage, damage-type, range, ammunition, etc.
- **Armor**: ac, gear_slot, armor type
- **Food/Consumables**: effects array with stat modifications, duration, etc.
- **Tools**: proficiency requirements, uses
- **Containers**: contents array, capacity

#### 2. Standardized Effects System
Create unified format for item effects:
```json
{
  "effects": [
    {
      "type": "modify_stat",
      "stat": "hunger",
      "amount": -20,
      "duration": "instant"
    },
    {
      "type": "heal",
      "amount": "1d8+2",
      "duration": "instant"
    }
  ]
}
```

#### 3. Validation System
- Type-specific validation rules
- Required field checking per item type
- Cross-reference validation (ensure referenced items exist)
- Value range validation (e.g., weight > 0, price >= 0)
- Automated validation reports

#### 4. Migration Tools
- Convert legacy D&D items to game-specific format
- Remove unnecessary fields automatically
- Standardize effect formats
- Batch update tools

#### 5. Editor Enhancements
- Type-specific field templates
- Smart field suggestions based on item type
- Real-time validation feedback
- Bulk edit capabilities
- Field type enforcement (number, string, array, etc.)

### Migration Path

As the game mechanics evolve:
1. **Audit Phase**: Catalog all current item types and their actual usage
2. **Schema Definition**: Create formal schemas for each item category
3. **Validation**: Build validation tools to identify inconsistencies
4. **Migration**: Update items to match new schemas
5. **Enforcement**: Add editor validation to prevent future issues

## 🔍 Item Field Reference

### Standard Fields (All Items)
- `id` - Unique identifier (must match filename)
- `name` - Display name
- `description` - In-game description
- `ai_description` - Description for AI image generation (optional but recommended)
- `price` - Gold piece cost
- `type` - Item category (see types below)
- `weight` - Weight in pounds
- `stack` - Max stack size (1 for non-stackable)
- `rarity` - common, uncommon, rare, very rare, legendary
- `tags` - Array of tags for filtering/searching
- `notes` - Internal development notes (not shown to players)
- `image` - Path to item image (e.g., `/res/img/items/{id}.png`)

### Type-Specific Fields

**Weapons:**
- `damage` - Damage dice (e.g., "1d8" or "1d8,1d10" for versatile)
- `damage-type` - slashing, piercing, bludgeoning, etc.
- `range` - Weapon range in feet
- `range-long` - Long range for ranged weapons
- `ammunition` - Ammo type required (for ranged weapons)
- `gear_slot` - "hands" for equippable weapons

**Armor:**
- `ac` - Armor class bonus
- `gear_slot` - armor, helmet, boots, gloves, etc.

**Consumables (Food, Potions):**
- `heal` - Healing amount (if applicable)
- `effects` - Array of stat modifications (needs standardization)

**Containers:**
- `contents` - Array of `[item_id, quantity]` pairs

### Item Types in Use
Based on current codebase:
- Martial Melee Weapons
- Simple Melee Weapons
- Martial Ranged Weapons
- Simple Ranged Weapons
- Adventuring Gear
- Tools
- Armor (Light, Medium, Heavy)
- Ammunition
- Clothes
- Potion
- Food
- Container

## 🛠️ Development Workflow Recommendations

1. **Before editing items**: Understand which fields are actually used in game code
2. **When adding new items**: Use `ai_description` field for better image generation
3. **For consumables**: Document the intended effect before implementing
4. **For validation**: Run validation checks before committing item changes
5. **For refactoring**: Use the preview feature to see all references before applying

## 🚧 Active Development Areas

- Game mechanics still evolving
- Item effects system being standardized
- Food/hunger/fatigue system in flux
- Equipment stats and effects being refined

**Recommendation**: When in doubt about a field, add it to `notes` array rather than deleting it. This preserves information for future reference while keeping the item functional.