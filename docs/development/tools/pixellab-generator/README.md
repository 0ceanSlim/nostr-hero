# PixelLab Generator

A Go tool for generating 32x32 pixel art images for fantasy game items using the PixelLab API.

## Features

- Generate pixel art images for D&D/fantasy game items
- Support for both `pixflux` and `bitforge` models
- **Uses JSON descriptions** from item files for better accuracy
- Fantasy-themed prompts with pixel-perfect settings
- **Run-specific folders** for organized output
- **SVG vectorization** for infinite scaling without blur
- Balance checking and cost estimation
- Configurable generation counts
- Transparent background support

## Setup

1. Add your PixelLab API key to `config.yml`:
```yaml
pixellab:
  api_key: "your-api-key-here"
```

2. Build the tool:
```bash
make pixellab-gen
```

## Usage

### Check API Balance
```bash
make pixellab-balance
# or
./pixellab-gen.exe balance
```

### Estimate Costs (Dry Run)
```bash
make pixellab-dry-run
# or
./pixellab-gen.exe dry-run --count 2 --model bitforge
```

### Generate Images
```bash
# Test with 2 items
make test-pixellab

# Generate specific count
./pixellab-gen.exe generate --count 5 --model bitforge

# Generate maximum possible with current balance
./pixellab-gen.exe generate --max-balance --model pixflux
```

### Convert to Scalable SVG
```bash
# Convert latest run to SVG
make pixellab-vectorize
# or
./pixellab-gen.exe vectorize

# Convert specific run
./pixellab-gen.exe vectorize run_bitforge_20231003_094500
```

### Scale Testing
```bash
# Create test images at multiple scales
make pixellab-scale-test
# or
./pixellab-gen.exe scale-test
```

## Models

- **bitforge**: Better for smaller images, potentially cheaper
- **pixflux**: Standard pixel art generation

## Output Structure

Each generation run creates a timestamped folder:
```
www/res/img/items/
├── run_bitforge_20231003_094500/
│   ├── png/
│   │   ├── acid.png
│   │   └── abacus.png
│   └── svg/
│       ├── acid.svg (references ../png/acid.png)
│       └── abacus.svg (references ../png/abacus.png)
```

## Item Data & Prompts

The tool reads item data from `docs/data/equipment/items/*.json` files:

1. **First priority**: Uses `description` field from JSON if available
2. **Fallback**: Uses hardcoded descriptions for common items
3. **Final fallback**: Generates based on item type and name

Enhanced with:
- Item rarity (common → simple, legendary → glowing effects)
- Item type categorization (weapons, armor, tools, etc.)
- Pixel-perfect rendering settings
- Fantasy D&D styling

## SVG Benefits

- **Infinite scaling**: No blur when resized
- **Small file size**: Vector-based
- **Browser compatible**: Works in web applications
- **Pixel-perfect CSS**: Maintains crisp edges at any size