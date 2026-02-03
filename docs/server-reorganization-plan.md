# Server Reorganization & Build Process Plan

This document outlines the full reorganization of the `server/` codebase and establishment of a proper build/versioning process.

---

## Overview

| Phase | Focus                      | Status  |
| ----- | -------------------------- | ------- |
| 1     | Game Logic Cleanup         | Planned |
| 2     | Nostr Package (from Grain) | Planned |
| 3     | Routes Reorganization      | Planned |
| 4     | Main.go Simplification     | Planned |
| 5     | Go Workspace (go.work)     | Planned |
| 6     | Build Process & Versioning | Planned |

---

## Phase 1: Game Logic Cleanup

### Goal

Move game logic from `utils/` into proper `game/` subpackages.

### Function Analysis

**utils/time_utils.go → game/npc/schedule.go**

| Function                 | Used By                                  | Purpose                       |
| ------------------------ | ---------------------------------------- | ----------------------------- |
| `GetCurrentScheduleSlot` | `api/data/npcs.go`, `ResolveNPCSchedule` | Find active schedule slot     |
| `ResolveNPCSchedule`     | `game/npc/dialogue.go`                   | Get NPC availability/location |
| `DetermineLocationType`  | `game/npc/dialogue.go`                   | Check building vs district    |
| `getAllDialogueKeys`     | private helper                           | Extract dialogue keys         |
| `getAllActions`          | private helper                           | Extract NPC actions           |
| `GetDayOfWeek`           | `game/npc/entertainment.go`              | Day of week calculation       |

**utils/building_utils.go → game/building/building.go**

| Function                          | Used By                                    | Purpose                      |
| --------------------------------- | ------------------------------------------ | ---------------------------- |
| `IsBuildingOpen`                  | `game/movement/`, `api/session_manager.go` | Check building access        |
| `GetAllBuildingStatesForDistrict` | `game/npc/housing.go`, `game/gametime/`    | Building states for district |
| `BuildingHours` struct            | Above functions                            | Schedule data structure      |

### Implementation Steps

1. Create `game/building/building.go`
   - Move from `utils/building_utils.go`
   - Change package to `building`

2. Create `game/npc/schedule.go`
   - Move from `utils/time_utils.go`
   - Change package to `npc`

3. Update imports:
   - `game/npc/housing.go` → add `building` import
   - `game/npc/dialogue.go` → remove `utils` prefix (same package)
   - `game/npc/entertainment.go` → remove `utils` prefix
   - `game/gametime/gametime.go` → add `building` import
   - `game/movement/movement.go` → add `building` import
   - `api/session_manager.go` → add `building` import
   - `api/data/npcs.go` → add `npc` import for schedule functions

4. Delete old files:
   - `utils/building_utils.go`
   - `utils/time_utils.go`

5. Remove unused `utils` imports from affected files

### Files Changed

| File                        | Action              |
| --------------------------- | ------------------- |
| `game/building/building.go` | Create (~205 lines) |
| `game/npc/schedule.go`      | Create (~111 lines) |
| `utils/building_utils.go`   | Delete              |
| `utils/time_utils.go`       | Delete              |
| 7 files                     | Import updates      |

---

## Phase 2: Nostr Package

### Goal

Create a `nostr/` package using code from the [Grain project](https://github.com/0ceanslim/grain).

### Current Files to Reorganize

| File                          | Contents                       | Action                  |
| ----------------------------- | ------------------------------ | ----------------------- |
| `utils/encodeNpub.go`         | `EncodeNpub`                   | Replace with Grain code |
| `utils/fetchUserMetaData.go`  | `FetchUserMetadata`            | Replace with Grain code |
| `utils/fetchUserRelays.go`    | `FetchUserRelays`, `Mailboxes` | Replace with Grain code |
| `utils/helpers.go`            | `DecodeNpub`                   | Move to nostr/          |
| `types/nostrEvent.go`         | `NostrEvent`                   | Move to nostr/          |
| `types/subscriptionFilter.go` | `SubscriptionFilter`           | Move to nostr/          |
| `types/userMetadata.go`       | `UserMetadata`                 | Move to nostr/          |

### Target Structure

```
server/nostr/
├── encode.go       # EncodeNpub, DecodeNpub (from Grain)
├── metadata.go     # FetchUserMetadata (from Grain)
├── relays.go       # FetchUserRelays, Mailboxes (from Grain)
├── client.go       # Relay client (from Grain if applicable)
└── types.go        # NostrEvent, SubscriptionFilter, UserMetadata
```

### Implementation Steps

1. Review Grain project for reusable code:
   - `github.com/0ceanslim/grain` - identify utility functions
   - Check for relay client, encoding utilities

2. Create `server/nostr/` package

3. Either:
   - Import Grain as dependency (`go get github.com/0ceanslim/grain`)
   - Or copy relevant functions directly

4. Move types from `types/` to `nostr/types.go`

5. Update all imports in `auth/`, `api/`, `utils/`

6. Delete old utils files

---

## Phase 3: Routes Reorganization

### Goal

Consolidate page handlers in `routes/` for cleaner organization.

### Current Structure

```
routes/
├── InitializeRoutes.go  # Route registration
├── index.go             # / handler
├── game.go              # /game handler
├── new-game.go          # /new-game handler
├── loadSave.go          # /load-save handler
├── saves.go             # /saves handler
├── discover.go          # /discover handler
├── settings.go          # /settings handler
├── alpha-registry.go    # /alpha-registry handler
└── legacy-registry.go   # /legacy-registry handler
```

### Target Structure

```
routes/
├── routes.go        # InitializeRoutes, RegisterPages
├── index.go         # / handler
├── game_pages.go    # /game, /new-game, /load-save, /saves (consolidated)
├── discover.go      # /discover
├── settings.go      # /settings
└── registry.go      # /alpha-registry, /legacy-registry (consolidated)
```

### Implementation Steps

1. Merge game-related handlers into `game_pages.go`
2. Merge registry handlers into `registry.go`
3. Create `RegisterPages(mux)` function
4. Update `InitializeRoutes.go`

---

## Phase 4: Main.go Simplification

### Goal

Simplify `main.go` to ~30 lines with clear structure.

### Current Issues

- All route registration inline (~50 lines of route setup)
- Mixed concerns (API, pages, static files, debug)

### Target Structure

```go
package main

func main() {
    // 1. Load config
    cfg := utils.LoadConfig("config.yml")

    // 2. Initialize services
    initServices(cfg)
    defer shutdownServices()

    // 3. Create router
    mux := http.NewServeMux()

    // 4. Register routes
    registerStaticFiles(mux)
    routes.RegisterPages(mux)
    api.RegisterEndpoints(mux, cfg)
    registerDebugEndpoints(mux, cfg)

    // 5. Start server
    startServer(mux, cfg)
}
```

### Implementation Steps

1. Create `api/register.go` with `RegisterEndpoints(mux, cfg)`
2. Create `routes/register.go` with `RegisterPages(mux)`
3. Extract service init to helper functions
4. Refactor `main.go` to use new registration functions

---

## Phase 5: Go Workspace

### Goal

Create `go.work` at project root to unify Go modules.

### Current Modules

```
nostr-hero/
├── server/go.mod                           # module nostr-hero
├── game-data/CODEX/go.mod                  # module codex
├── docs/development/tools/*/go.mod         # tool modules
└── tmp/go-scraper/go.mod                   # temporary
```

### Implementation

1. Create `go.work` at project root:

   ```go
   go 1.24

   use (
       ./server
       ./game-data/CODEX
   )
   ```

2. Verify commands work from root:
   - `go build ./...`
   - `go test ./...`

3. Decision: Commit or gitignore `go.work`
   - Recommend: Commit for monorepo consistency

---

## Phase 6: Build Process & Versioning

### Goal

Establish proper versioning and automated build process.

### Versioning Scheme

Format: `MAJOR.MINOR.PATCH-HASH`

| Stage       | Version Format   | Trigger     |
| ----------- | ---------------- | ----------- |
| Development | `0.0.0-<commit>` | Every push  |
| Alpha       | `0.0.X-<commit>` | Manual flag |
| Beta        | `0.X.X-<commit>` | Manual flag |
| Release     | `X.X.X`          | Manual flag |

**Current:** All pushes create `0.0.0-<8 char commit hash>`

**Example progression:**

```
0.0.0-a1b2c3d4  # Development push
0.0.0-e5f6g7h8  # Development push
0.0.1-i9j0k1l2  # Flagged as alpha ready
0.0.1-m3n4o5p6  # Alpha push
0.0.2-q7r8s9t0  # Flagged alpha increment
0.1.0-u1v2w3x4  # Flagged as beta ready
1.0.0           # First full release
```

### GitHub Actions Workflow

**File:** `.github/workflows/build.yml`

```yaml
name: Build & Version

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      bump_type:
        description: "Version bump type"
        required: false
        type: choice
        options:
          - none
          - alpha
          - beta
          - release

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.24"

      - name: Generate version
        id: version
        run: |
          COMMIT_HASH=$(git rev-parse --short=8 HEAD)
          # Read current version from VERSION file or default
          if [ -f VERSION ]; then
            BASE_VERSION=$(cat VERSION)
          else
            BASE_VERSION="0.0.0"
          fi
          VERSION="${BASE_VERSION}-${COMMIT_HASH}"
          echo "version=$VERSION" >> $GITHUB_OUTPUT

      - name: Build
        run: |
          cd server
          go build -ldflags "-X main.Version=${{ steps.version.outputs.version }}" -o ../nostr-hero

      - name: Run tests
        run: |
          cd server
          go test ./test/...

      - name: Deploy to test server
        if: success()
        run: |
          # Webhook or deployment script
          echo "Deploying version ${{ steps.version.outputs.version }}"
```

### Version Bump Process

1. **Manual version bumps** via `VERSION` file:

   ```bash
   echo "0.0.1" > VERSION  # Flag alpha ready
   git add VERSION
   git commit -m "bump: alpha 0.0.1"
   git push
   ```

2. Or via **workflow dispatch**:
   - Go to Actions → Build & Version → Run workflow
   - Select bump type

### Build Artifacts

| Environment | Binary Name       | Version      |
| ----------- | ----------------- | ------------ |
| Test Server | `nostr-hero-test` | `0.0.X-hash` |
| Production  | `nostr-hero`      | `X.X.X`      |

### Implementation Steps

1. Create `VERSION` file with `0.0.0`
2. Update `.github/workflows/deployment-status.yml`:
   - Add build step
   - Add version generation
   - Add test execution
   - Update deployment trigger
3. Add `-ldflags` to embed version in binary
4. Add `/api/version` endpoint to expose version
5. Update test server deployment to use versioned builds

---

## Verification Checklist

### After Phase 1

- [ ] `go build ./...` passes
- [ ] `go test ./test/...` passes (26 tests)
- [ ] Building open/closed states work
- [ ] NPC schedules resolve correctly

### After Phase 2

- [ ] Nostr authentication works
- [ ] Profile fetching works
- [ ] No duplicate code with Grain

### After Phase 3

- [ ] All pages render correctly
- [ ] Route registration is cleaner

### After Phase 4

- [ ] Server starts correctly
- [ ] All endpoints work
- [ ] main.go is ~30 lines

### After Phase 5

- [ ] `go build ./...` works from project root
- [ ] `go test ./...` works from project root
- [ ] Go Report Card can analyze the project

### After Phase 6

- [ ] Versions appear in build output
- [ ] Test server shows correct version
- [ ] Manual version bumps work
- [ ] Tests run on every push

---

## Timeline

This is a multi-session project. Each phase can be completed independently:

- **Phase 1**: ~1 session (cleanest, most impactful)
- **Phase 2**: ~1 session (depends on Grain code review)
- **Phase 3**: ~0.5 session (straightforward consolidation)
- **Phase 4**: ~0.5 session (refactoring)
- **Phase 5**: ~0.25 session (simple)
- **Phase 6**: ~1 session (CI/CD work)

---

## Notes

- Phase 1 should be done first (most dependencies)
- Phase 2 can be done anytime (isolated)
- Phases 3-4 can be combined
- Phase 5 is quick win
- Phase 6 can be done in parallel with others
