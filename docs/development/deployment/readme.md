# Nostr Hero - Deployment & Versioning Strategy

This document outlines the multi-server deployment architecture, versioning strategy, and build automation plans for Nostr Hero.

## Server Environments

Nostr Hero runs three separate server environments with different purposes, access controls, and save systems.

### 1. Test Server (`test.nostrhero.quest`)

**Purpose**: Development and testing environment

**Configuration**:
- Runs in **dev mode** (Air live-reload)
- Debug logging **always enabled**
- Direct deployment from git pushes

**Save System**:
- Saves stored **on disk only** (no Nostr relay integration)
- No event ID tracking
- Full local save file management

**Access Control**:
- **Whitelist by pubkey** (curated list)
- Restricted to developers and select testers

**Deployment**:
- **Auto-deploy on every push** to main/test branch
- Service automatically restarts after deployment
- No version tracking required
- See `SETUP-AUTO-DEPLOY.md` for setup instructions
- **Privacy**: Server details never committed to repo (webhook or self-hosted runner)

**Timeline**: Active now

---

### 2. Official Server (Primary Domain - TBD)

**Purpose**: Production environment for official releases

**Configuration**:
- Runs **versioned releases** (compiled binaries)
- Production optimizations enabled
- Stable, tested code only

**Save System**:
- Saves tracked via **Nostr event ID**
- Saves published to **official Nostr relays**
- Server fetches saves from official relays only
- Versioned save compatibility tracking

**Access Control**:
- **Alpha**: HappyTavern members only
- **Beta+**: Free public access

**Deployment**:
- Deploy on **versioned releases only** (semver: v0.1.0-alpha, v0.2.0-beta, v1.0.0, etc.)
- Manual promotion from test server after QA
- Service restart on new version deployment

**Timeline**: Launches with 1st alpha release

---

### 3. Modded Server

**Purpose**: Community server with relaxed save restrictions

**Configuration**:
- Runs **same version as official server**
- May allow modified game rules/content in future

**Save System**:
- Fetches **any game save event** from Nostr
- Allows loading saves from any relay
- No official version restrictions
- Players can load community/modified saves

**Access Control**:
- **Beta+**: HappyTavern members only (exclusive feature)

**Deployment**:
- Deploys **when official server gets new version**
- Mirrors official release schedule
- Separate service instance

**Timeline**: Launches with beta release

---

## Deployment Infrastructure

All servers run on **Cloudflare Argo Tunnel** as Linux services.

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Linux Server (Local Machine)                                │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Test Service │  │ Official Svc │  │ Modded Svc   │      │
│  │ (Air/Dev)    │  │ (Binary)     │  │ (Binary)     │      │
│  │ Port: TBD    │  │ Port: TBD    │  │ Port: TBD    │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │               │
│  ┌──────┴─────────────────┴─────────────────┴───────┐      │
│  │ Cloudflare Argo Tunnel (cloudflared)             │      │
│  └──────┬───────────────────┬─────────────────┬─────┘      │
└─────────┼───────────────────┼─────────────────┼────────────┘
          │                   │                 │
          ▼                   ▼                 ▼
  test.nostrhero.quest   nostrhero.com   modded.nostrhero.com
      (or similar)        (example)         (or similar)
```

### Deployment Triggers

| Event | Test Server | Official Server | Modded Server |
|-------|-------------|-----------------|---------------|
| Git push to main/test branch | ✅ Auto-deploy + restart | ❌ No change | ❌ No change |
| Versioned release (git tag) | ✅ Deploy new version | ✅ Deploy new version | ✅ Deploy new version |
| Manual trigger | ✅ Allowed | ✅ Allowed | ✅ Allowed |

---

## Versioning Strategy

### Semantic Versioning

Nostr Hero follows **semantic versioning** (semver):

- **Alpha**: `v0.1.0-alpha`, `v0.2.0-alpha`, etc.
- **Beta**: `v0.1.0-beta`, `v0.2.0-beta`, etc.
- **Release**: `v1.0.0`, `v1.1.0`, `v2.0.0`, etc.

### Version Compatibility

**Official Server**:
- Tracks save file version compatibility
- May refuse to load saves from incompatible versions
- Event metadata includes game version

**Modded Server**:
- Best-effort compatibility with all versions
- Allows loading any save regardless of version

---

## Build Automation

### Auto-Deployment Options

**Test Server** has two options for auto-deployment (see `SETUP-AUTO-DEPLOY.md`):

**Option 1: GitHub Webhook + Server Listener** (Recommended)
- Server listens for GitHub webhook events
- On push → automatically pulls and restarts
- **Privacy**: No server credentials in GitHub repo
- **Implementation**: `webhook-listener-advanced.sh` + `webhook-hooks.json`

**Option 2: Self-Hosted GitHub Actions Runner**
- GitHub runner software on server
- Workflow triggers on push (`.github/workflows/deploy-test.yml`)
- **Privacy**: No server credentials in GitHub repo
- **Implementation**: Runner authenticates directly with GitHub

### CI/CD Pipeline

**On Git Push (Test Server)**:
1. ~~Run tests (if available)~~ (Optional, add later)
2. ~~Build test binary~~ (Air handles this in dev mode)
3. **Pull latest code** from GitHub
4. **Restart test service** (systemd)
5. ~~Post Nostr note as Dungeon Master npub with commit details~~ (TODO)

**On Versioned Release** (git tag) - TODO:
1. Run full test suite
2. Build production binaries (official + modded)
3. Generate release notes
4. Deploy to official server
5. Deploy to modded server (if beta+)
6. Restart services
7. Post Nostr note as Dungeon Master npub with release announcement

### Nostr Integration

**Dungeon Master npub**: `npub1...` (TBD)

**Note Types**:
- **Code pushes**: Short changelog note with commit hash
- **Versioned releases**: Detailed release notes with features/fixes
- **Server status**: Maintenance notifications, downtime alerts

**Relay Publishing**:
- Notes published to official Nostr relays
- Tagged appropriately for filtering (e.g., `#NostrHero`, `#DevUpdate`)

---

## Access Control Implementation

### Alpha Phase (Official Server Launch)

**HappyTavern Member Verification**:
- Option 1: Maintain whitelist of HappyTavern member npubs
- Option 2: Check for HappyTavern badge/credential on Nostr profile
- Option 3: Token-based access (distributed to members)

**Test Server**:
- Hardcoded pubkey whitelist in `config.yml` or database
- Check npub on login
- Reject non-whitelisted users

### Beta Phase

**Official Server**: Open to public

**Modded Server**: HappyTavern members only (same verification as alpha)

### Post-Beta (v1.0+)

**Official Server**: Fully public

**Modded Server**: HappyTavern members only (premium feature)

---

## Save System Implementation Details

### Official Server Save Flow

```
Player saves game
    ↓
Frontend creates save object
    ↓
POST to /api/saves/{npub}
    ↓
Backend saves to disk + publishes Nostr event
    ↓
Event includes:
    - Save data (encrypted or public)
    - Event ID (for tracking)
    - Game version tag
    - Timestamp
    ↓
Event published to official relays
    ↓
On load: Fetch events from relays by npub
    ↓
Filter by official version compatibility
    ↓
Load most recent compatible save
```

### Modded Server Save Flow

```
Player requests save list
    ↓
Fetch ALL save events from Nostr (any relay)
    ↓
Filter by npub (or allow cross-npub if desired)
    ↓
Present all saves (no version filtering)
    ↓
Player selects save
    ↓
Load save (best-effort compatibility)
```

### Test Server Save Flow

```
Player saves game
    ↓
POST to /api/saves/{npub}
    ↓
Save to disk only
    ↓
No Nostr relay interaction
    ↓
On load: Read from disk
```

---

## Timeline & Milestones

| Phase | Test Server | Official Server | Modded Server |
|-------|-------------|-----------------|---------------|
| **Now** | ✅ Active (disk saves, dev mode) | ❌ Not live | ❌ Not live |
| **Alpha Release** | ✅ Continues (whitelist) | ✅ Launch (HappyTavern only) | ❌ Not live |
| **Beta Release** | ✅ Continues (whitelist) | ✅ Public access | ✅ Launch (HappyTavern only) |
| **v1.0 Release** | ✅ Continues (whitelist) | ✅ Public access | ✅ HappyTavern only |

---

## Implementation Checklist

### Phase 1: Test Server (Current)
- [x] Basic server running
- [ ] Auto-deploy script on git push
- [ ] Service restart automation
- [ ] Pubkey whitelist enforcement
- [ ] Disk-based save system

### Phase 2: Alpha Prep (Official Server)
- [ ] Nostr relay integration for saves
- [ ] Event ID tracking system
- [ ] HappyTavern member verification
- [ ] Production build pipeline
- [ ] Cloudflare Argo Tunnel setup (official domain)
- [ ] Dungeon Master npub creation
- [ ] Nostr note publishing on releases

### Phase 3: Beta Prep (Modded Server)
- [ ] Modded server codebase (forked or config-based)
- [ ] "Fetch any save event" logic
- [ ] Cloudflare Argo Tunnel setup (modded domain)
- [ ] HappyTavern member verification (modded)
- [ ] Cross-version save loading (best-effort)

### Phase 4: Build Automation
- [ ] CI/CD pipeline (GitHub Actions or custom)
- [ ] Automated testing suite
- [ ] Release note generation
- [ ] Binary builds for Linux
- [ ] Deploy scripts (rsync, scp, or API-based)
- [ ] Service management scripts (systemd)
- [ ] Nostr note publishing automation

---

## Technical TODOs

### Nostr Save Event Schema

**Kind**: TBD (likely `kind: 30000+` for parameterized replaceable event)

**Tags**:
- `["d", "{npub}_{timestamp}"]` - Unique identifier
- `["game", "nostr-hero"]` - Game identifier
- `["version", "v0.1.0-alpha"]` - Game version
- `["server", "official"]` or `["server", "modded"]`

**Content**: JSON save data (possibly encrypted)

### Relay List

**Official Relays** (for official server):
- TBD (curated list of reliable relays)

**Community Relays** (for modded server):
- Fetch from all known public relays

### Service Management

**Systemd Service Files**:
- `nostr-hero-test.service`
- `nostr-hero-official.service`
- `nostr-hero-modded.service`

**Deployment Scripts**:
- `deploy-test.sh` - Git pull, build (Air), restart service
- `deploy-official.sh` - Download release binary, restart service
- `deploy-modded.sh` - Download release binary, restart service

---

## Questions to Resolve

1. **Dungeon Master npub**: Generate new keypair or use existing?
2. **HappyTavern verification**: Which method (whitelist, badge, token)?
3. **Save encryption**: Should Nostr save events be encrypted (NIP-04/NIP-44)?
4. **Modded server rules**: Will modded server allow game rule modifications, or just unrestricted save loading?
5. **Domain names**: Final domain names for official and modded servers?
6. **Relay selection**: Which relays for official save publishing?
7. **CI/CD platform**: GitHub Actions, GitLab CI, or custom scripts?

---

**Last Updated**: 2025-12-07
**Status**: Planning & Documentation Phase
