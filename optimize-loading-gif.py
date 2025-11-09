#!/usr/bin/env python3
"""
Optimize loading.gif - make background transparent and reduce file size
"""
from PIL import Image
import os

# Paths
input_path = 'www/res/img/loading.gif'
output_path = 'www/res/img/loading-optimized.gif'

print(f"Loading {input_path}...")

# Open the GIF
img = Image.open(input_path)

# Get all frames
frames = []
durations = []

try:
    while True:
        # Convert to RGBA to support transparency
        frame = img.convert('RGBA')

        # Get the background color (assuming it's white or light colored)
        # We'll make white/light colors transparent
        datas = frame.getdata()

        newData = []
        for item in datas:
            # If pixel is white or very light (close to white), make it transparent
            if item[0] > 240 and item[1] > 240 and item[2] > 240:
                newData.append((255, 255, 255, 0))  # Transparent
            else:
                newData.append(item)

        frame.putdata(newData)
        frames.append(frame)

        # Get frame duration
        duration = img.info.get('duration', 100)
        durations.append(duration)

        img.seek(img.tell() + 1)
except EOFError:
    pass  # End of sequence

print(f"Processed {len(frames)} frames")

# Save optimized GIF with transparency
frames[0].save(
    output_path,
    save_all=True,
    append_images=frames[1:],
    duration=durations,
    loop=0,
    transparency=0,
    disposal=2,
    optimize=True
)

# Get file sizes
original_size = os.path.getsize(input_path)
optimized_size = os.path.getsize(output_path)

print(f"Original size: {original_size} bytes")
print(f"Optimized size: {optimized_size} bytes")
print(f"Saved {original_size - optimized_size} bytes ({100 - (optimized_size/original_size*100):.1f}% reduction)")
print(f"Saved to {output_path}")
print("\nTo use the optimized version, rename it:")
print(f"   move {output_path} {input_path}")
