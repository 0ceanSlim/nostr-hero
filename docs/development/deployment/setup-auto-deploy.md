# Auto-Deployment Setup Guide

This guide explains how to set up automatic deployment for the test server while keeping server details private.

## Option 1: GitHub Webhook + Server Listener (Recommended)

**Pros**: Server controls updates, no GitHub secrets, simple
**Cons**: Requires installing webhook tool

### Step 1: Install webhook tool on server

```bash
# On your Linux server
wget https://github.com/adnanh/webhook/releases/download/2.8.1/webhook-linux-amd64.tar.gz
tar -xvf webhook-linux-amd64.tar.gz
sudo mv webhook-linux-amd64/webhook /usr/local/bin/
```

### Step 2: Configure the scripts

1. Create `webhook-hooks.json` from template:
   ```bash
   cd /path/to/nostr-hero/scripts
   cp webhook-hooks.json.template webhook-hooks.json
   ```

2. Generate webhook secret:
   ```bash
   openssl rand -hex 32
   ```
   Copy this secret - you'll need it for GitHub and the config file.

3. Edit `scripts/webhook-hooks.json`:
   - Replace `/path/to/nostr-hero` with actual path (2 places)
   - Replace `REPLACE_WITH_WEBHOOK_SECRET` with the secret from step 2

4. Edit `scripts/webhook-listener-advanced.sh`:
   - Replace `/path/to/nostr-hero` with actual path
   - Replace `nostr-hero-test` with your actual service name

5. Make script executable:
   ```bash
   chmod +x scripts/webhook-listener-advanced.sh
   ```

**Note**: `webhook-hooks.json` is gitignored to keep secrets private.

### Step 3: Set up systemd service

1. Create `webhook-listener.service` from template:
   ```bash
   cp webhook-listener.service.template webhook-listener.service
   ```

2. Edit `scripts/webhook-listener.service`:
   - Replace `REPLACE_WITH_USERNAME` with your Linux username
   - Replace all `/path/to/nostr-hero` with actual path

3. Install the service:
   ```bash
   sudo cp scripts/webhook-listener.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable webhook-listener
   sudo systemctl start webhook-listener
   ```

3. Check status:
   ```bash
   sudo systemctl status webhook-listener
   ```

### Step 4: Configure Cloudflare Tunnel (if needed)

If your webhook listener needs to be accessible from GitHub:

1. Add to your `cloudflared` config:
   ```yaml
   ingress:
     - hostname: webhook.nostrhero.quest
       service: http://localhost:9000
   ```

2. Restart tunnel:
   ```bash
   sudo systemctl restart cloudflared
   ```

**Alternative**: Use Cloudflare Tunnel's built-in ingress without exposing publicly (if webhook port is local-only)

### Step 5: Configure GitHub Webhook

1. Go to your GitHub repo → Settings → Webhooks → Add webhook
2. **Payload URL**: `https://webhook.nostrhero.quest` (or your tunnel URL)
3. **Content type**: `application/json`
4. **Secret**: Paste the secret from Step 2
5. **Events**: Select "Just the push event"
6. Click "Add webhook"

### Step 6: Test the deployment

1. Make a commit and push to main:
   ```bash
   git commit --allow-empty -m "Test auto-deployment"
   git push
   ```

2. Check webhook listener logs:
   ```bash
   sudo journalctl -u webhook-listener -f
   ```

3. Check deployment logs:
   ```bash
   tail -f /var/log/nostr-hero-deploy.log
   ```

---

## Option 2: Self-Hosted GitHub Actions Runner

**Pros**: Native GitHub integration, official solution
**Cons**: Requires runner software on server

### Step 1: Add runner to your repository

1. Go to GitHub repo → Settings → Actions → Runners → New self-hosted runner
2. Select Linux x64
3. Follow the installation commands on your server:
   ```bash
   # Example (use the commands GitHub provides):
   mkdir actions-runner && cd actions-runner
   curl -o actions-runner-linux-x64-2.311.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz
   tar xzf ./actions-runner-linux-x64-2.311.0.tar.gz
   ./config.sh --url https://github.com/your-username/nostr-hero --token YOUR_TOKEN
   ```

### Step 2: Install runner as service

```bash
cd ~/actions-runner
sudo ./svc.sh install
sudo ./svc.sh start
```

### Step 3: Give runner sudo permissions (for service restart)

1. Edit sudoers file:
   ```bash
   sudo visudo
   ```

2. Add line (replace `your-username`):
   ```
   your-username ALL=(ALL) NOPASSWD: /bin/systemctl restart nostr-hero-test
   your-username ALL=(ALL) NOPASSWD: /bin/systemctl status nostr-hero-test
   ```

### Step 4: Configure the workflow

The workflow is already created at `.github/workflows/deploy-test.yml`.

1. Edit the file and replace:
   - `/path/to/nostr-hero` with actual path
   - `nostr-hero-test` with your service name

2. Commit and push:
   ```bash
   git add .github/workflows/deploy-test.yml
   git commit -m "Add auto-deployment workflow"
   git push
   ```

### Step 5: Test the deployment

1. Make a commit and push:
   ```bash
   git commit --allow-empty -m "Test auto-deployment"
   git push
   ```

2. Check GitHub Actions tab in your repo to see the workflow run

---

## Security Considerations

### Option 1 (Webhook)
- ✅ Server details never in GitHub
- ✅ Webhook secret validates GitHub is the sender
- ⚠️ Webhook endpoint must be accessible (use Cloudflare Tunnel)
- ⚠️ Ensure webhook secret is strong (32+ characters)

### Option 2 (Self-Hosted Runner)
- ✅ Server details never in GitHub
- ✅ Runner authenticates with GitHub token
- ✅ No inbound connections needed
- ⚠️ Runner has access to repo secrets (if any)
- ⚠️ Requires sudo permissions for service restart

---

## Troubleshooting

### Webhook not triggering

1. Check GitHub webhook delivery status (Settings → Webhooks → Recent Deliveries)
2. Check webhook listener logs:
   ```bash
   sudo journalctl -u webhook-listener -f
   ```
3. Verify Cloudflare Tunnel is routing correctly
4. Test webhook locally:
   ```bash
   curl -X POST http://localhost:9000
   ```

### GitHub Actions runner offline

1. Check runner status:
   ```bash
   cd ~/actions-runner
   sudo ./svc.sh status
   ```
2. Check runner logs:
   ```bash
   sudo journalctl -u actions.runner.* -f
   ```
3. Restart runner:
   ```bash
   sudo ./svc.sh restart
   ```

### Git pull fails

1. Ensure server has SSH key added to GitHub (or use HTTPS)
2. Check git remote:
   ```bash
   cd /path/to/nostr-hero
   git remote -v
   ```
3. Test manual pull:
   ```bash
   git pull origin main
   ```

### Service restart fails

1. Check service status:
   ```bash
   sudo systemctl status nostr-hero-test
   ```
2. Check sudo permissions (for self-hosted runner)
3. Manually test restart:
   ```bash
   sudo systemctl restart nostr-hero-test
   ```

---

## Recommended: Option 1 (Webhook)

For your use case, I recommend **Option 1 (Webhook)** because:
- ✅ Server fully controls when and how it updates
- ✅ No GitHub runner software to maintain
- ✅ Simple and lightweight
- ✅ Easy to debug (just check webhook logs)
- ✅ No sudo permission concerns

---

## Next Steps

After setting up auto-deployment:

1. Update `DEPLOYMENT.md` with actual webhook URL / runner details
2. Test deployment with a dummy commit
3. Set up log rotation for `/var/log/nostr-hero-deploy.log`:
   ```bash
   sudo nano /etc/logrotate.d/nostr-hero-deploy
   ```
   ```
   /var/log/nostr-hero-deploy.log {
       daily
       rotate 7
       compress
       missingok
       notifempty
   }
   ```

4. Consider adding Nostr note publishing to the deployment script (for Dungeon Master npub updates)

---

**Last Updated**: 2025-12-07
