#!/bin/bash
# GitHub Webhook Listener for Auto-Deployment
# Usage: Run this as a systemd service on the test server

PORT=9000  # Webhook listener port (not exposed publicly, only via Cloudflare tunnel if needed)
SECRET="your-webhook-secret-here"  # Set this in GitHub webhook settings
REPO_PATH="/path/to/nostr-hero"  # Absolute path to repo on server
SERVICE_NAME="nostr-hero-test"

echo "🎣 Starting webhook listener on port $PORT..."

# Simple webhook receiver using netcat (or use a proper tool like webhook/adnanh)
while true; do
  echo "Waiting for webhook..."

  # Listen for POST request (basic implementation)
  # For production, use https://github.com/adnanh/webhook instead
  REQUEST=$(echo -e "HTTP/1.1 200 OK\n\n" | nc -l -p $PORT)

  echo "📨 Webhook received!"

  # Navigate to repo
  cd "$REPO_PATH" || exit

  # Pull latest code
  echo "📥 Pulling latest code..."
  git pull origin main

  if [ $? -eq 0 ]; then
    echo "✅ Code updated successfully"

    # Restart service
    echo "🔄 Restarting service..."
    sudo systemctl restart "$SERVICE_NAME"

    if [ $? -eq 0 ]; then
      echo "✅ Service restarted successfully"
    else
      echo "❌ Service restart failed!"
    fi
  else
    echo "❌ Git pull failed!"
  fi

  echo "---"
done
