# Frontend Styling Guide & Troubleshooting

This document provides information on the project's CSS build process using Tailwind CSS.

## Styling Issues?

If the application is running but the styles look broken or are missing, it's likely an issue with how the CSS is being loaded or compiled. Here are the things to check:

### 1. Are you using the pre-compiled `output.css`?

The default HTML layout (`www/views/templates/layout.html`) links to a static, pre-compiled CSS file:
```html
<link href="/res/style/output.css" rel="stylesheet" />
```
This file should be checked into the repository and contain all necessary styles. If it's missing or out of date, you may need to rebuild it.

### 2. Are you running the Tailwind CSS build process?

If you are actively developing the frontend and making changes to styles, you need to run the Tailwind CSS compiler. This process watches your files and automatically rebuilds `output.css` when it detects changes.

You can run this process using `npx`:
```bash
npx tailwindcss -i ./www/res/style/input.css -o ./www/res/style/output.css --watch
```
Ensure this command is running in a separate terminal while the main Go server is running.

### 3. Are you using the Tailwind Play CDN?

For a quick start without needing to build any CSS, you can use the Tailwind Play CDN. This is a script that compiles the necessary styles directly in your browser. To use it, you must edit `www/views/templates/layout.html` and replace the `output.css` link with the CDN script:

**Replace this:**
```html
<link href="/res/style/output.css?v=20250109-23" rel="stylesheet" />
```

**With this:**
```html
<script src="https://cdn.tailwindcss.com"></script>
```
**Note:** The Play CDN is great for development and quick testing, but it is not recommended for production. For official builds, a static `output.css` file should always be generated.