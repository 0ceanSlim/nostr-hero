# Self-Hosted GitHub Actions Runner Setup

This guide covers setting up a self-hosted GitHub Actions runner on the Ubuntu server for automated deployments.

## Prerequisites

- Ubuntu server with Go 1.24+, Node.js 18+, and `swag` installed
- SSH access to the server
- Admin access to the GitHub repository

## 1. Create the Runner

1. Go to the repo on GitHub: **Settings > Actions > Runners > New self-hosted runner**
2. Select **Linux** and **x64**
3. Follow the download and configure steps shown on GitHub:

```bash
# Create a directory for the runner
mkdir ~/actions-runner && cd ~/actions-runner

# Download (URL from GitHub UI)
curl -o actions-runner-linux-x64-2.XXX.X.tar.gz -L https://github.com/actions/runner/releases/download/vX.X.X/actions-runner-linux-x64-X.X.X.tar.gz

# Extract
tar xzf ./actions-runner-linux-x64-2.XXX.X.tar.gz

# Configure (token from GitHub UI)
./config.sh --url https://github.com/OWNER/REPO --token YOUR_TOKEN
```

## 2. Install as a Service

```bash
cd ~/actions-runner
sudo ./svc.sh install
sudo ./svc.sh start
```

Verify it's running:

```bash
sudo ./svc.sh status
```

The runner should appear as "Idle" in GitHub Settings > Actions > Runners.

## 3. Sudoers for Service Management

The deploy workflow needs to stop/start systemd services without a password prompt. Add these entries:

```bash
sudo visudo -f /etc/sudoers.d/pubkey-quest
```

Add:

```
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl stop pubkey-quest-test
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl stop pubkey-quest
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl stop codex
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl start pubkey-quest-test
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl start pubkey-quest
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl start codex
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart pubkey-quest-test
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart pubkey-quest
USER ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart codex
```

Replace `USER` with the actual username running the runner.

## 4. Install the Systemd Services

Copy the service templates from `docs/development/deployment/` and replace `USER` with your username:

```bash
# Test server
sudo cp pubkey-quest-test.service.template /etc/systemd/system/pubkey-quest-test.service
sudo sed -i 's/USER/your-username/g' /etc/systemd/system/pubkey-quest-test.service

# Codex
sudo cp codex.service.template /etc/systemd/system/codex.service
sudo sed -i 's/USER/your-username/g' /etc/systemd/system/codex.service

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable pubkey-quest-test codex
sudo systemctl start pubkey-quest-test codex
```

## 5. Server Directory Layout

The self-hosted runner checks out code to `/home/pubkey-test/` (configured as the runner's work directory or via checkout action). The layout after a successful deploy:

```
/home/pubkey-test/
├── config.yml              # Test server config (port, debug_mode: true)
├── codex-config.yml        # Codex config
├── pubkey-quest            # Built server binary
├── codex                   # Built codex binary
├── www/
│   ├── game.db             # Migrated database
│   └── dist/               # Built frontend
├── game-data/              # JSON source data
├── data/saves/             # Test save files
└── ... (full repo checkout)
```

## 6. Runner Labels

The runner is registered with these labels (used in workflow `runs-on`):

- `self-hosted`
- `linux`
- `x64`

## 7. Verify

After setup, push a commit to `main` and check:

1. **GitHub Actions tab** — `Deploy Test Server` workflow should trigger
2. **Runner status** — Should show as "Active" during the run
3. **Services** — `sudo systemctl status pubkey-quest-test` and `sudo systemctl status codex` should show the services running
4. **Version** — `./pubkey-quest -version` should print the version with commit hash

## Troubleshooting

| Problem | Fix |
|---------|-----|
| Runner offline | `cd ~/actions-runner && sudo ./svc.sh start` |
| Permission denied on restart | Check sudoers file — `sudo visudo -f /etc/sudoers.d/pubkey-quest` |
| Build fails but old service still runs | This is expected — the workflow stops before the deploy step on build failure |
| swag not found | `go install github.com/swaggo/swag/cmd/swag@latest` and ensure `~/go/bin` is in PATH |
