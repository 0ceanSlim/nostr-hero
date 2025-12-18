# Troubleshooting Webhook Deployment

## Issue: GitHub shows "npm install failed" but unsure if it actually failed

### Step 1: Check the Webhook Log File

The webhook script logs everything to a file (defined as `LOG_FILE` in the script):

```bash
# Check the deployment log (replace path with your actual log path)
tail -n 100 /var/log/nostr-hero-deploy.log

# Or follow it in real-time during a deployment
tail -f /var/log/nostr-hero-deploy.log
```

**What to look for:**
- `🔍 Environment Info:` - Shows user, PATH, Node/NPM versions
- `❌ npm install failed!` - Actual failure message
- `✅ Dependencies installed` - Success message

### Step 2: Check the Webhook Service Logs

If using the `webhook` tool as a service:

```bash
# Check webhook service status
sudo systemctl status webhook-listener  # or your service name

# View recent logs
sudo journalctl -u webhook-listener -n 100 --no-pager

# Follow logs in real-time
sudo journalctl -u webhook-listener -f
```

### Step 3: Test npm Manually as Webhook User

The webhook likely runs as a different user than you. Test if npm works for that user:

```bash
# Find out which user runs the webhook
ps aux | grep webhook

# Switch to that user and test npm
sudo -u <webhook-user> npm --version
sudo -u <webhook-user> node --version

# Test full build as that user
sudo -u <webhook-user> bash -c "cd /path/to/nostr-hero && npm install && npm run build"
```

### Step 4: Fix PATH Issues

If `npm` is not found in the webhook's PATH, you need to either:

**Option A: Add npm to PATH in the webhook script**

Edit your actual webhook script (not the template) and add at the top:

```bash
# Add Node.js to PATH (adjust version number based on your installation)
export PATH="/usr/bin:$PATH"
# OR if using nvm:
# export NVM_DIR="$HOME/.nvm"
# [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
```

**Option B: Use absolute path to npm**

The updated template now auto-detects npm location:

```bash
NPM_CMD=$(command -v npm || echo "/usr/bin/npm")
$NPM_CMD install
```

If this still fails, manually set the path:

```bash
NPM_CMD="/usr/local/bin/npm"  # or wherever npm is installed
$NPM_CMD install
```

### Step 5: Check GitHub Status Reporting

The script reports status to GitHub using the `update_github_status` function. To see what's being reported:

```bash
# In your deployment log, look for lines with curl commands
grep "curl.*github.com.*statuses" /var/log/nostr-hero-deploy.log
```

**To see GitHub status details:**

1. Go to your GitHub repo → **Commits** tab
2. Click on the latest commit
3. Look for "deployment/webhook" check
4. Click "Details" (if available)

### Step 6: Test Webhook Manually

Trigger the webhook script manually to see detailed output:

```bash
# Run the script directly (adjust path)
cd /path/to/nostr-hero
sudo -u <webhook-user> bash ./webhook-listener-advanced.sh
```

## Common Issues & Solutions

### Issue: "npm: command not found"

**Cause**: npm is not in the PATH for the webhook user

**Solution**:
```bash
# Find where npm is installed
which npm
# Example output: /usr/bin/npm

# Update webhook script to use full path
NPM_CMD="/usr/bin/npm"
```

### Issue: "EACCES: permission denied"

**Cause**: The webhook user doesn't have write permissions

**Solution**:
```bash
# Check ownership of the repo
ls -la /path/to/nostr-hero

# Fix ownership (replace webhook-user with actual user)
sudo chown -R webhook-user:webhook-user /path/to/nostr-hero

# Or give write permissions to node_modules
chmod -R 755 /path/to/nostr-hero/node_modules
```

### Issue: Build succeeds but GitHub still shows failure

**Cause**: The GitHub status was reported incorrectly earlier

**Solution**:
- The next successful deployment will update the status
- Or manually trigger another push to test

### Issue: No logs appearing

**Cause**: Log file doesn't exist or wrong permissions

**Solution**:
```bash
# Create log file with correct permissions
sudo touch /var/log/nostr-hero-deploy.log
sudo chown webhook-user:webhook-user /var/log/nostr-hero-deploy.log
sudo chmod 644 /var/log/nostr-hero-deploy.log

# Or use a different log location in user's home
LOG_FILE="$HOME/nostr-hero-deploy.log"
```

## Debugging Checklist

Run through these checks on your production server:

```bash
# 1. Is Node.js installed?
node --version  # Should be v18+

# 2. Is npm accessible?
npm --version

# 3. Can the webhook user run npm?
sudo -u <webhook-user> npm --version

# 4. Does the repo have correct permissions?
ls -la /path/to/nostr-hero

# 5. Can you manually build successfully?
cd /path/to/nostr-hero
npm install
npm run build
ls -lh www/dist/main.css  # Should exist

# 6. Check webhook logs
tail -n 50 /var/log/nostr-hero-deploy.log

# 7. Test GitHub status reporting
# (Make a test push and check commit status)
```

## Testing Without Pushing to GitHub

Test the entire deployment flow locally on the server:

```bash
# 1. SSH into your production server
ssh user@your-server

# 2. Navigate to repo
cd /path/to/nostr-hero

# 3. Pull latest (simulates webhook)
git pull origin main

# 4. Run the build steps manually
npm install
npm run build

# 5. Check if build succeeded
ls -lh www/dist/main.css
cat www/dist/main.css | head -c 200  # Should show CSS

# 6. Restart service
sudo systemctl restart your-service-name

# 7. Test in browser
curl http://localhost:your-port/dist/main.css | head -c 200
```

## Still Having Issues?

If the deployment is actually working (service restarts, new code is running) but GitHub just shows the wrong status:

**Temporary workaround**: Comment out the GitHub status reporting in your webhook script

```bash
# In webhook-listener-advanced.sh, comment out these lines:
# update_github_status "pending" "Deployment in progress..."
# update_github_status "failure" "npm install failed"
# update_github_status "success" "Deployment completed successfully"
```

The deployment will still work; you just won't see status on GitHub commits.

---

**Last Updated**: 2024-12-18
