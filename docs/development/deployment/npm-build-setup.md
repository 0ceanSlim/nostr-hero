# NPM Build System - Production Deployment Setup

## Overview

As of December 2024, Nostr Hero uses an npm-based build system for frontend assets (CSS/JavaScript). The production deployment process now requires Node.js and npm to build the frontend before starting the Go server.

## What Changed

**Before**: Only Go was needed on the production server
```bash
git pull origin main
go run main.go  # or restart service
```

**After**: Node.js + npm are required for frontend builds
```bash
git pull origin main
npm install      # Install dependencies
npm run build    # Build CSS/JS assets
go run main.go   # or restart service
```

## Production Server Requirements

### 1. Install Node.js (v18+)

The production server must have Node.js v18 or higher installed:

```bash
# Check if Node.js is installed
node --version   # Should be v18.0.0 or higher
npm --version

# If not installed, install Node.js:
# Ubuntu/Debian:
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Verify installation
node --version
npm --version
```

### 2. Update Your Deployment Script

The webhook deployment script (`webhook-listener-advanced.sh`) has been updated to include:

1. **npm install** - Installs/updates dependencies from `package.json`
2. **npm run build** - Builds the frontend (generates `www/dist/main.css` and JS bundles)

See the updated template in `webhook-listener-advanced.sh.template` (lines 53-71).

### 3. Update Your Service File (if using Air in production)

**⚠️ IMPORTANT**: If your production service currently uses `Air` for live-reload, you should switch to a compiled binary or `go run`:

**Option A: Use compiled binary** (recommended for production)
```bash
# Build once
go build -o nostr-hero main.go

# Service file uses the binary
ExecStart=/path/to/nostr-hero/nostr-hero
```

**Option B: Use `go run`** (simpler, but slower startup)
```bash
# Service file
ExecStart=/usr/bin/go run /path/to/nostr-hero/main.go
```

**Option C: Keep using Air** (if you prefer dev-mode in production)
```bash
# No changes needed - Air will work fine
ExecStart=/usr/bin/air
```

### 4. Gitignore Configuration

The following files are now tracked in git (no longer gitignored):
- `package.json` ✅ Tracked (dependencies)
- `package-lock.json` ⚠️ **Should be tracked** (lock file for reproducible builds)
- `vite.config.js` ✅ Tracked (build configuration)
- `postcss.config.js` ✅ Tracked (CSS processing config)
- `tailwind.config.js` ❌ Not needed (Tailwind v4 uses CSS-based config)

The build output directory `www/dist/` remains gitignored (generated on deployment).

### 5. Initial Setup on Production Server

When deploying for the first time with the new build system:

```bash
# Navigate to repo
cd /path/to/nostr-hero

# Pull latest code
git pull origin main

# Install npm dependencies
npm install

# Build frontend assets
npm run build

# Verify build output exists
ls -lh www/dist/main.css    # Should exist and be ~31KB
ls -lh www/dist/*.js        # Should have game.js, index.js, etc.

# Restart your service
sudo systemctl restart nostr-hero  # or whatever your service name is
```

## Deployment Workflow

### Automated Deployment (Webhook)

The webhook script now handles everything automatically:

```
1. Git push → GitHub webhook triggered
2. Server pulls latest code
3. npm install (installs/updates dependencies)
4. npm run build (builds CSS/JS)
5. Service restart (Go server with new assets)
```

### Manual Deployment

If deploying manually:

```bash
cd /path/to/nostr-hero
git pull origin main
npm install
npm run build
sudo systemctl restart nostr-hero
```

## Build Outputs

The `npm run build` command generates:

```
www/dist/
├── main.css           # Single CSS file with all styles (~31KB)
├── game.js            # Game page JavaScript bundle
├── gameIntro.js       # Game intro page bundle
├── newGame.js         # New game page bundle
├── index.js           # Home page bundle
├── settings.js        # Settings page bundle
├── discover.js        # Discover page bundle
├── saves.js           # Saves page bundle
├── *.js.map           # Source maps for debugging
└── chunks/            # Shared code chunks
```

These files are served by the Go server from `/dist/*` URLs.

## Troubleshooting

### Build fails with "Cannot find module"
```bash
# Delete node_modules and reinstall
rm -rf node_modules
npm install
npm run build
```

### CSS missing or styles look wrong
```bash
# Rebuild the CSS
npm run build

# Check if main.css exists and has content
ls -lh www/dist/main.css   # Should be ~31KB
head -c 100 www/dist/main.css  # Should show CSS content
```

### Old styles still showing in browser
- Hard refresh the browser: `Ctrl + F5` (Windows/Linux) or `Cmd + Shift + R` (Mac)
- The Go server serves static files with cache headers

### Service won't start after update
```bash
# Check service logs
sudo journalctl -u nostr-hero -n 50

# Verify Go server still works
cd /path/to/nostr-hero
go run main.go  # Should start without errors
```

## Development vs Production

### Development (Local)
```bash
# Watch mode - rebuilds on file changes
npm run build:watch

# Or run Vite dev server (with HMR)
npm run dev
```

### Production (Server)
```bash
# Single build
npm run build

# Then start Go server
go run main.go  # or use compiled binary
```

## Reverting to Old System (Emergency)

If you need to quickly rollback to the old system without npm:

1. **Check out the commit before the npm migration**:
   ```bash
   git log --oneline  # Find the last commit before npm was added
   git checkout <commit-hash>
   ```

2. **Restart service** (no npm build needed for old commits)

3. **Fix forward**: Once stable, update your production server with Node.js and redeploy properly

---

**Last Updated**: 2024-12-18
**Status**: Active - npm build system deployed
