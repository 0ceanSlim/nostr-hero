# Styling System

This directory contains the Tailwind CSS configuration and styles for Nostr Hero.

## Files

- **`input.css`** - Tailwind source file with custom base styles and CSS variables
- **`output.css`** - Compiled Tailwind CSS (auto-generated, minified)
- **`tailwind.config.js`** - Tailwind configuration with theme system and custom utilities
- **`fonts.css`** - Dogica pixel font declarations

## Development

### Using Air (Recommended)

Air automatically builds Tailwind CSS when files change. Just run from project root:

```bash
air
```

Air will:

- Watch for changes in `www/views/**/*.html` and `www/scripts/**/*.js`
- Rebuild CSS automatically
- Hot reload the Go server

### Manual Build

Build CSS manually when needed:

```bash
# Development build
tailwindcss -c ./www/res/style/tailwind.config.js -i ./www/res/style/input.css -o ./www/res/style/output.css

# Production build (minified)
tailwindcss -c ./www/res/style/tailwind.config.js -i ./www/res/style/input.css -o ./www/res/style/output.css --minify
```

### Watch Mode

For manual development without Air:

```bash
tailwindcss -c ./www/res/style/tailwind.config.js -i ./www/res/style/input.css -o ./www/res/style/output.css --watch
```

## Theme System

The project uses CSS custom properties for flexible theming:

### Available Themes

- **`dark`** (default) - Dark theme with purple accents
- **`light`** - Light theme
- **`midnight`** - Dark blue with gold accents
- **`lava`** - Red/black volcanic theme
- **`solar`** - Cyan/teal solarized theme

### Using Themes

Switch themes by setting the `data-theme` attribute on the `<html>` element:

```javascript
document.documentElement.setAttribute("data-theme", "midnight");
```

### Theme-Aware Colors

Use these Tailwind classes for colors that adapt to the current theme:

```html
<div class="bg-theme-bg-primary text-theme-text-primary">
  Content that changes with theme
</div>
```

**Available theme color classes:**

- `bg-theme-bg-primary` / `text-theme-bg-primary`
- `bg-theme-bg-secondary` / `text-theme-bg-secondary`
- `bg-theme-bg-tertiary` / `text-theme-bg-tertiary`
- `bg-theme-bg-inverted` / `text-theme-bg-inverted`
- `text-theme-text-primary`
- `text-theme-text-secondary`
- `text-theme-text-muted`
- `text-theme-text-inverted`
- `text-theme-text-highlighted`

### Win95 Game UI (Static Colors)

The game interface uses fixed Win95-style retro colors that don't change with themes:

```html
<div class="win95-inset">Dark retro panel (always the same)</div>
```

**Available Win95 color classes:**

- `bg-win95-bg`
- `bg-win95-bg-dark`
- `bg-win95-bg-light`
- `bg-win95-bg-lighter`
- `bg-win95-black`

## Custom Utilities

### Win95 Components

Pre-built component classes for retro Win95 styling:

- **`.win95-inset`** - Sunken panel with inset borders
- **`.win95-outset`** - Raised panel with outset borders
- **`.win95-button`** - Interactive button with hover/active states
- **`.win95-button-pressed`** - Pressed button appearance

Example:

```html
<button class="px-4 py-2 text-white win95-button">Click Me</button>
```

### Pixel-Art Utilities

- **`.pixel-perfect`** - Crisp pixel rendering (no anti-aliasing)
- **`.pixel-clip`** - 3px corner clipping for retro borders
- **`.pixel-clip-sm`** - 2px corner clipping

Example:

```html
<img src="sprite.png" class="w-16 h-16 pixel-perfect pixel-clip" />
```

## Font System

The project uses **Dogica**, a pixel-perfect font designed for retro aesthetics.

### Font Families

- **`font-dogica`** - Standard Dogica font
- **`font-dogica-pixel`** - Dogica Pixel variant

### Font Rendering

Global pixel-perfect font rendering is applied automatically:

- No font smoothing
- Crisp edges
- Optimized for speed

All images and canvas elements also render with pixelated mode by default.

## Adding Custom Styles

### Modifying Themes

Edit theme colors in `input.css`:

```css
:root[data-theme="custom"] {
  --color-bgPrimary: rgb(r, g, b);
  --color-textPrimary: rgb(r, g, b);
  /* ... */
}
```

### Adding Utilities

Edit `tailwind.config.js` to add new utilities or components:

```javascript
plugins: [
  function ({ addUtilities }) {
    addUtilities({
      ".my-custom-class": {
        /* styles */
      },
    });
  },
];
```

### Custom CSS

Add custom CSS to `input.css` using Tailwind layers:

```css
@layer components {
  .my-component {
    /* styles */
  }
}
```

## CSS Variables Reference

See `input.css` for the complete list of CSS custom properties used in themes. All theme variables follow the pattern `--color-*`.
