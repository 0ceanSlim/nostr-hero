# GitHub Commit Status Reporting Setup

## Overview

The webhook can report deployment status back to GitHub, showing success/failure checks on each commit. This requires a GitHub Personal Access Token.

## How It Works

When configured properly, you'll see deployment status on each commit:

```
✅ Deployment Status / report - Deployment completed successfully
❌ Deployment Status / report - npm install failed
⏳ Deployment Status / report - Deployment in progress...
```

## Setup Steps

### 1. Create a GitHub Personal Access Token

1. Go to GitHub → **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Click **Generate new token** → **Generate new token (classic)**
3. Set:
   - **Note**: `Nostr Hero Deployment Status Reporter`
   - **Expiration**: 90 days (or No expiration for less maintenance)
   - **Scopes**: Check `repo:status` (only this permission is needed)
4. Click **Generate token**
5. **Copy the token immediately** (you won't see it again!)

### 2. Update webhook-hooks.json

Edit your actual `webhook-hooks.json` file (not the template) on your production server:

```json
{
  "envname": "GITHUB_TOKEN",
  "source": "string",
  "name": "ghp_YOUR_ACTUAL_TOKEN_HERE"
}
```

Replace `REPLACE_WITH_YOUR_GITHUB_TOKEN_HERE` with your actual token (starts with `ghp_`).

**Important**: Keep this file secure! The token has write access to your repo's commit statuses.

### 3. Restart the Webhook Service

After updating the config:

```bash
# Restart the webhook listener service
sudo systemctl restart webhook-listener  # or your service name

# Verify it's running
sudo systemctl status webhook-listener

# Check logs
sudo journalctl -u webhook-listener -n 20
```

### 4. Test It

Make a test commit and push:

```bash
# On your local machine
git commit --allow-empty -m "Test deployment status reporting"
git push origin main

# Then check GitHub:
# 1. Go to your repo → Commits tab
# 2. Click on the latest commit
# 3. You should see "deployment/webhook" check with status
```

## Verification

### Check if Environment Variables Are Passed

Look at your deployment log:

```bash
tail -n 100 /var/log/nostr-hero-deploy.log
```

You should see:
```
🚀 Deployment started: <date>
Commit: <commit-sha>
Repo: <username>/<repo-name>
Pushed by: <your-github-username>
```

If these show actual values (not empty), the webhook is passing the variables correctly.

### Check if Status Is Being Sent

Look for curl commands in the log:

```bash
grep "Authorization" /var/log/nostr-hero-deploy.log
# (This won't show anything if output is redirected to /dev/null)
```

Or add debug output to the script temporarily:

```bash
# In webhook-listener-advanced.sh, modify update_github_status:
update_github_status() {
  local state=$1
  local description=$2

  echo "📡 Sending status to GitHub: $state - $description"  # ADD THIS LINE

  if [ -n "$GITHUB_TOKEN" ] && [ -n "$COMMIT_SHA" ] && [ -n "$REPO_NAME" ]; then
    # ... rest of function
```

## Troubleshooting

### Status Not Showing on GitHub

**Check 1**: Is GITHUB_TOKEN set?
```bash
# In your deployment log, look for:
echo "🔍 GITHUB_TOKEN: ${GITHUB_TOKEN:0:10}..."  # Shows first 10 chars
```

**Check 2**: Is the token valid?
```bash
# Test the API manually (replace with your values)
curl -H "Authorization: token ghp_YOUR_TOKEN" \
  https://api.github.com/repos/YOUR_USERNAME/nostr-hero/statuses/COMMIT_SHA \
  -d '{"state":"success","description":"Test","context":"deployment/webhook"}'
```

**Check 3**: Does the token have the right permissions?
- Go to GitHub → Settings → Developer settings → Personal access tokens
- Find your token
- Verify `repo:status` is checked

### "401 Unauthorized" Error

**Cause**: Invalid or expired token

**Solution**: Generate a new token and update webhook-hooks.json

### Status Shows But Stays "Pending"

**Cause**: The script is failing before it can report success/failure

**Solution**: Check the deployment log for actual errors:
```bash
tail -n 200 /var/log/nostr-hero-deploy.log | grep "❌"
```

### Don't Want Status Reporting

If you don't want to set up GitHub status reporting, you can:

**Option 1**: Leave GITHUB_TOKEN empty (status reporting will be skipped)

**Option 2**: Remove the `pass-environment-to-command` section from webhook-hooks.json

**Option 3**: Comment out the `update_github_status` calls in the deployment script

## Security Notes

- **Never commit webhook-hooks.json to git** (it contains your token!)
- The `.gitignore` already excludes it: `webhook-hooks.json`
- If you accidentally commit the token, **immediately revoke it** on GitHub
- Consider using a GitHub App instead of personal tokens for better security (more complex setup)

## Alternative: Skip Status Reporting

If the status reporting is more trouble than it's worth, you can simplify the webhook:

1. **Remove environment variables** from webhook-hooks.json
2. **Comment out status updates** in the deployment script:

```bash
# update_github_status "pending" "Deployment in progress..."
# update_github_status "failure" "npm install failed"
# update_github_status "success" "Deployment completed successfully"
```

The deployment will still work; you just won't see status checks on GitHub commits.

---

**Last Updated**: 2024-12-18
