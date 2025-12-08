#!/bin/bash
# Advanced GitHub Webhook Listener using 'webhook' tool
# Install: https://github.com/adnanh/webhook
# Usage: webhook -hooks hooks.json -verbose

# This script is called by the webhook tool when a push event is received
# Define this in hooks.json (see webhook-hooks.json)

REPO_PATH="/path/to/nostr-hero"
SERVICE_NAME="nostr-hero-test"
LOG_FILE="/var/log/nostr-hero-deploy.log"

{
  echo "========================================="
  echo "🚀 Deployment started: $(date)"
  echo "========================================="

  # Navigate to repo
  cd "$REPO_PATH" || {
    echo "❌ Failed to navigate to repo path: $REPO_PATH"
    exit 1
  }

  # Pull latest code
  echo "📥 Pulling latest code from main branch..."
  git pull origin main

  if [ $? -eq 0 ]; then
    echo "✅ Code updated successfully"

    # Optional: Run tests before restarting
    # echo "🧪 Running tests..."
    # go test ./... || { echo "❌ Tests failed, aborting deployment"; exit 1; }

    # Restart service
    echo "🔄 Restarting $SERVICE_NAME service..."
    sudo systemctl restart "$SERVICE_NAME"

    if [ $? -eq 0 ]; then
      echo "✅ Service restarted successfully"
      echo "🎉 Deployment completed: $(date)"
    else
      echo "❌ Service restart failed!"
      exit 1
    fi
  else
    echo "❌ Git pull failed!"
    exit 1
  fi

} 2>&1 | tee -a "$LOG_FILE"
