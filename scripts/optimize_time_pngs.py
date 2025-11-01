#!/usr/bin/env python3
"""
PNG Cleanup and Optimization Script for Time-of-Day Images

Features:
- Backs up originals to 'originals/' subfolder
- Creates circular mask to remove artifacts outside main circle
- Optional color reduction for retro/stylized look
- Optional pixelation effect
- PNG optimization

Usage:
    python optimize_time_pngs.py [options]

Options:
    --pixelate FACTOR    Pixelate by reducing size by factor (e.g., 4 = 256->64->256)
    --colors NUM         Reduce to NUM colors (posterize effect, e.g., 32)
    --radius PERCENT     Circle radius as percent of image size (default: 90)
    --no-backup         Skip backing up originals
"""

import os
import sys
import shutil
from pathlib import Path
from PIL import Image, ImageDraw, ImageOps

def backup_originals(image_dir, backup_dir):
    """Backup original images to a subfolder."""
    backup_path = image_dir / backup_dir
    backup_path.mkdir(exist_ok=True)

    png_files = list(image_dir.glob("*.png"))
    if not png_files:
        print(f"No PNG files found in {image_dir}")
        return False

    print(f"Backing up {len(png_files)} images to {backup_dir}/")
    for png_file in png_files:
        if backup_dir not in str(png_file):  # Don't backup the backup folder
            backup_file = backup_path / png_file.name
            if not backup_file.exists():
                shutil.copy2(png_file, backup_file)
                print(f"  [OK] Backed up: {png_file.name}")
            else:
                print(f"  [SKIP] Already backed up: {png_file.name}")

    return True

def create_circular_mask(size, radius_percent=90, feather=10):
    """Create a circular alpha mask with soft edges (radial gradient)."""
    width, height = size
    mask = Image.new('L', (width, height), 0)

    # Calculate circle parameters (centered)
    center_x, center_y = width // 2, height // 2
    max_radius = min(width, height) * radius_percent / 200

    # Create radial gradient
    for y in range(height):
        for x in range(width):
            # Calculate distance from center
            dx = x - center_x
            dy = y - center_y
            distance = (dx * dx + dy * dy) ** 0.5

            # Calculate alpha based on distance
            if distance <= max_radius - feather:
                # Fully opaque inside inner circle
                alpha = 255
            elif distance <= max_radius:
                # Gradient fade in feather zone
                fade = (max_radius - distance) / feather
                alpha = int(255 * fade)
            else:
                # Fully transparent outside
                alpha = 0

            mask.putpixel((x, y), alpha)

    return mask

def reduce_colors(image, num_colors):
    """Reduce image to a limited color palette (posterize effect)."""
    # Convert to P mode (palette) with specified colors
    return image.convert('P', palette=Image.ADAPTIVE, colors=num_colors).convert('RGBA')

def pixelate(image, factor):
    """Pixelate image by downscaling then upscaling."""
    width, height = image.size
    small_size = (width // factor, height // factor)

    # Downscale with nearest neighbor (no antialiasing)
    small = image.resize(small_size, Image.NEAREST)

    # Upscale back to original size
    return small.resize((width, height), Image.NEAREST)

def process_image(input_path, radius_percent=90, pixelate_factor=None, num_colors=None):
    """Process a single PNG image."""
    print(f"\nProcessing: {input_path.name}")

    # Open image
    img = Image.open(input_path)

    # Ensure RGBA mode
    if img.mode != 'RGBA':
        img = img.convert('RGBA')

    original_size = img.size
    print(f"  Original size: {original_size[0]}x{original_size[1]}")

    # Apply pixelation first (if requested)
    if pixelate_factor and pixelate_factor > 1:
        print(f"  Applying pixelation (factor: {pixelate_factor})...")
        img = pixelate(img, pixelate_factor)

    # Reduce colors (if requested)
    if num_colors and num_colors > 0:
        print(f"  Reducing to {num_colors} colors...")
        img = reduce_colors(img, num_colors)

    # Create circular mask with soft edges
    print(f"  Applying circular mask ({radius_percent}% radius)...")
    circular_mask = create_circular_mask(img.size, radius_percent)

    # Get existing alpha channel or create one
    if img.mode == 'RGBA':
        existing_alpha = img.split()[3]
    else:
        existing_alpha = Image.new('L', img.size, 255)

    # Combine masks: multiply existing alpha with circular mask
    # This preserves the blur/gradient from the original
    from PIL import ImageChops
    combined_mask = ImageChops.multiply(existing_alpha, circular_mask)

    # Apply combined mask to alpha channel
    img.putalpha(combined_mask)

    # Save optimized PNG
    print(f"  Saving optimized version...")
    img.save(input_path, 'PNG', optimize=True)

    print(f"  [DONE]")

def main():
    # Parse simple command-line arguments
    pixelate_factor = None
    num_colors = None
    radius_percent = 90
    do_backup = True

    args = sys.argv[1:]
    i = 0
    while i < len(args):
        if args[i] == '--pixelate' and i + 1 < len(args):
            pixelate_factor = int(args[i + 1])
            i += 2
        elif args[i] == '--colors' and i + 1 < len(args):
            num_colors = int(args[i + 1])
            i += 2
        elif args[i] == '--radius' and i + 1 < len(args):
            radius_percent = int(args[i + 1])
            i += 2
        elif args[i] == '--no-backup':
            do_backup = False
            i += 1
        elif args[i] in ['-h', '--help']:
            print(__doc__)
            return
        else:
            print(f"Unknown argument: {args[i]}")
            print("Use --help for usage information")
            return

    # Find image directory
    script_dir = Path(__file__).parent
    image_dir = script_dir.parent / 'www' / 'res' / 'img' / 'time'

    if not image_dir.exists():
        print(f"Error: Image directory not found: {image_dir}")
        return

    print("=" * 60)
    print("PNG Cleanup and Optimization")
    print("=" * 60)
    print(f"Image directory: {image_dir}")
    print(f"Settings:")
    print(f"  - Circle radius: {radius_percent}%")
    if pixelate_factor:
        print(f"  - Pixelation factor: {pixelate_factor}")
    if num_colors:
        print(f"  - Color reduction: {num_colors} colors")
    print()

    # Backup originals
    if do_backup:
        if not backup_originals(image_dir, "originals"):
            return
    else:
        print("Skipping backup (--no-backup specified)")

    # Process each PNG
    png_files = sorted([f for f in image_dir.glob("*.png") if 'originals' not in str(f)])

    if not png_files:
        print("\nNo PNG files to process!")
        return

    print(f"\nProcessing {len(png_files)} images...")

    for png_file in png_files:
        try:
            process_image(png_file, radius_percent, pixelate_factor, num_colors)
        except Exception as e:
            print(f"  [ERROR] Processing {png_file.name}: {e}")

    print("\n" + "=" * 60)
    print("All done!")
    print("=" * 60)
    print("\nOriginal files preserved in: www/res/img/time/originals/")
    print("\nTo experiment with different settings:")
    print("  python scripts/optimize_time_pngs.py --pixelate 4 --colors 32")
    print("  python scripts/optimize_time_pngs.py --radius 95")
    print("  python scripts/optimize_time_pngs.py --colors 64 --pixelate 2")

if __name__ == '__main__':
    main()
