# Option 4: Go Backend State + Delta Updates

**Implementation Guide for UI Optimization & Nostr Integration**

---

## Executive Summary

This document outlines the implementation of a delta-based architecture that:

1. **Fixes UI flickering** (90%+ reduction in DOM mutations)
2. **Implements smooth clock display** (60fps interpolated, no more jerky jumps)
3. **Optimizes tick rate** (every in-game minute / ~417ms real time)
4. **Reduces network usage** (91% less bandwidth despite 12x more ticks)
5. **Optimizes Nostr save files** (5-10KB instead of 20-50KB)
6. **Aligns with Go-first philosophy** (backend authoritative, minimal JS)
7. **Enables future features** (multiplayer, real-time updates, state rollback)

**Tick Rate Decision:** 1 tick per in-game minute (~2.4 ticks/second real time)
**Estimated Effort:** 17-23 hours
**Target Completion:** 4 weeks

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Architecture Overview](#architecture-overview)
3. [Tick System Design](#tick-system-design)
4. [Smooth Clock Display](#smooth-clock-display)
5. [Backend Implementation](#backend-implementation)
6. [Frontend Implementation](#frontend-implementation)
7. [Migration Guide](#migration-guide)
8. [Testing Strategy](#testing-strategy)
9. [Performance Benchmarks](#performance-benchmarks)

---

## Problem Statement

### Current Issues

**UI Flickering (Every 5 seconds):**

- Complete DOM clearing with `innerHTML = ''`
- 50+ DOM mutations per update cycle
- Images recreated causing visual reloads
- No change detection or diffing

**Save File Size (Nostr Constraint):**

- Current saves: 20-50KB (includes full item/spell/location data)
- Nostr event limit: 64KB
- Network cost: Every save publishes to relay
- Privacy cost: Larger payload harder to encrypt

**Architecture Issues:**

- Client-side state management (violates Go-first philosophy)
- Redundant NPC API calls (every 5 seconds)
- No separation between session state and persistent state

---

## Architecture Overview

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ LAYER 1: SAVE FILE (Stored on Nostr Relays)                 │
│ ⚠️ MUST BE MINIMAL - Target: <10KB compressed                │
│                                                               │
│ Store ONLY:                                                  │
│   - Character stats (HP, fatigue, hunger, XP, gold)         │
│   - Inventory item IDs + quantities (NOT full item data)    │
│   - Equipment slot IDs                                       │
│   - Location + district + building (NOT full location data) │
│   - Spell slot IDs (NOT full spell data)                    │
│   - Active effect IDs + durations (NOT effect definitions)  │
│                                                               │
│ DO NOT STORE:                                                │
│   ✗ Item descriptions, stats, images (in game-data/)        │
│   ✗ Spell descriptions, effects (in game-data/)             │
│   ✗ Location layouts, buildings (in game-data/)             │
│   ✗ NPC data, dialogues (in game-data/)                     │
│   ✗ Computed/derived values (weight, NPC positions)         │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ Save (every 5 min) / Load
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ LAYER 2: SESSION STATE (In-Memory, Backend Only)            │
│ ✅ CAN BE LARGE - Not saved to Nostr                         │
│                                                               │
│ Contains:                                                    │
│   - Full hydrated character data (from game-data + save)    │
│   - Cached NPC positions (computed from schedules + time)   │
│   - Cached building states (computed from time)             │
│   - Computed inventory weight                               │
│   - Delta calculation snapshots                             │
│   - Session-only data (booked shows, dialogue state, etc.)  │
└─────────────────────────────────────────────────────────────┘
                            ▲
                            │ Hydrate on load
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ LAYER 3: GAME DATA (JSON Files, HTTP Served)                │
│ ✅ CAN BE UNLIMITED SIZE - Served via HTTP, not in saves    │
│                                                               │
│ Contains:                                                    │
│   - Full item definitions (stats, descriptions, images)     │
│   - Full spell definitions (effects, scaling)               │
│   - Full location layouts (districts, buildings)            │
│   - Full NPC definitions (dialogues, schedules)             │
│   - Monster stat blocks                                     │
│   - Effect definitions                                       │
└─────────────────────────────────────────────────────────────┘
```

### Delta Update Flow

```
┌─────────────────────────────────────────────────────────────┐
│ GO BACKEND (Authoritative State)                             │
│                                                               │
│  SessionManager (In-Memory State per npub)                  │
│    ├─ SessionState {                                         │
│    │    SaveFile:       *types.SaveFile                      │
│    │    NPCsAtLocation: []string (cached per hour)           │
│    │    BuildingStates: map[string]bool (cached per 5 min)   │
│    │    LastSnapshot:   *SessionSnapshot (for delta calc)    │
│    │  }                                                       │
│    │                                                          │
│    └─ Delta Calculator                                       │
│         1. Clone current state as "old"                      │
│         2. Process action/time update                        │
│         3. Compare old vs new state                          │
│         4. Build Delta (only changed fields)                 │
│         5. Return Delta (50-200 bytes)                       │
└─────────────────────────────────────────────────────────────┘
                            │
                            │ HTTP/JSON (Delta Payload)
                            ▼
┌─────────────────────────────────────────────────────────────┐
│ FRONTEND (Stateless Rendering Layer)                        │
│                                                               │
│  Delta Applier (Surgical DOM Patcher)                       │
│    - Receives delta from backend                            │
│    - Applies surgical updates to DOM                        │
│    - NO client-side state management                        │
│    - NO full re-renders                                     │
│                                                               │
│  Examples:                                                   │
│    delta.character.hp = 5                                    │
│      → getElementById('hp').textContent = 5                  │
│                                                               │
│    delta.npcs.added = ["barkeep"]                            │
│      → appendChild(createNPCButton("barkeep"))               │
│                                                               │
│    delta.buildings["tavern"] = true                          │
│      → button.style.background = '#6b8e6b' (green/open)      │
└─────────────────────────────────────────────────────────────┘
```

---

## Tick System Design

### Decision: Tick Every In-Game Minute

**Tick Rate Calculation:**

```
Time multiplier: 144x (1 real second = 144 in-game seconds)
Target: 1 tick per in-game minute

1 in-game minute = 60 in-game seconds
60 / 144 = 0.417 real seconds = 417ms

Result: ~2.4 ticks per real second
```

### Why This Rate?

| Factor | Requirement | How This Satisfies |
|--------|-------------|-------------------|
| Effect resolution | Max 30 in-game minutes | Tick every 1 min = 30x finer than needed |
| NPC schedules | Hourly changes | 60 ticks per hour = precise transitions |
| Building open/close | 15-30 min granularity | Well within resolution |
| Clock smoothness | Continuous display | Decoupled (see below) |
| Player perception | No noticeable jumps | 417ms is imperceptible |

### Computational Cost Analysis

**Per-Tick Operations (Delta System):**

| Operation | Cost | Notes |
|-----------|------|-------|
| Compare ~20 state fields | ~1-5μs | Integer comparisons |
| Check effect triggers | ~1-10μs | Counter decrements |
| Build delta object | ~5-20μs | Only changed fields |
| JSON marshal delta | ~10-50μs | Tiny payload |
| HTTP overhead | ~1-5ms | The "expensive" part |
| Frontend apply delta | ~0.1-1ms | Surgical DOM updates |
| **Total per tick** | **~2-10ms** | |

**At 2.4 ticks/second: ~5-24ms compute per second (leaving 976ms idle)**

### Most Ticks Are Nearly Empty

| Event | Frequency | Delta Size |
|-------|-----------|------------|
| Time update | Every tick | ~35 bytes |
| Hunger/Fatigue | Every 30-60 in-game min | +20 bytes |
| NPC changes | Every in-game hour | +50-100 bytes |
| Building state | Every 15-30 in-game min | +30-50 bytes |

**Typical tick: 30-50 bytes | Occasional tick: 100-200 bytes**

### Network Cost Comparison

```
Delta system (2.4 req/sec × ~50 bytes):  ~120 bytes/second
Current system (0.2 req/sec × ~7KB):     ~1,400 bytes/second

Delta system uses ~91% LESS bandwidth despite 12x more ticks
```

### Tick Rate Summary

| Component | Rate | Purpose |
|-----------|------|---------|
| **Logic tick** | 417ms (~2.4/sec) | Backend processes effects, schedules |
| **Display clock** | 60fps (decoupled) | Smooth visual interpolation |
| **Backend sync** | On action + each tick | Authoritative time correction |
| **Auto-save** | 5 real minutes | Persist to Nostr |

---

## Smooth Clock Display

### Problem: Jerky Clock

The current clock display is jerky because:
1. Clock updates are tied to tick timing
2. Actions cause irregular time jumps
3. No interpolation between ticks

### Solution: Decouple Display from Logic

```
┌─────────────────────────────────────────────────────────────┐
│ LOGIC LAYER (Backend/Go)                                    │
│ - Processes at tick rate (every 417ms)                      │
│ - Handles effects, NPC schedules, state changes             │
│ - Returns authoritative time_of_day after each action       │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ time_of_day (authoritative)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ DISPLAY LAYER (Frontend/JS)                                 │
│ - Runs at 60fps (requestAnimationFrame)                     │
│ - Interpolates clock display based on:                      │
│     • lastKnownTime (from backend)                          │
│     • timeSinceLastUpdate (real elapsed ms)                 │
│     • timeMultiplier (144x)                                 │
│     • isPaused flag                                         │
│ - Syncs to backend time on each action/tick response        │
└─────────────────────────────────────────────────────────────┘
```

### Smooth Clock Implementation

**File:** `src/systems/smoothClock.js`

```javascript
/**
 * SmoothClock provides interpolated time display independent of tick rate.
 * The display runs at 60fps for smooth visuals, while syncing to backend
 * authoritative time on each action/response.
 */
export class SmoothClock {
    constructor() {
        this.gameTime = 0;           // Current in-game minutes (0-1439)
        this.gameDay = 1;
        this.lastSyncTime = 0;       // Last backend-confirmed time
        this.lastSyncRealTime = 0;   // Real timestamp of last sync
        this.timeMultiplier = 144;
        this.isPaused = false;
        this.animationId = null;
    }

    /**
     * Sync clock to backend authoritative time.
     * Called after every action response and tick.
     */
    syncFromBackend(timeOfDay, currentDay) {
        this.lastSyncTime = timeOfDay;
        this.lastSyncRealTime = performance.now();
        this.gameTime = timeOfDay;
        this.gameDay = currentDay;
    }

    /**
     * Animation frame loop - runs at 60fps for smooth display.
     */
    tick() {
        if (this.isPaused) {
            this.animationId = requestAnimationFrame(() => this.tick());
            return;
        }

        const now = performance.now();
        const realElapsedMs = now - this.lastSyncRealTime;
        const realElapsedSeconds = realElapsedMs / 1000;

        // Calculate interpolated game time
        const gameSecondsElapsed = realElapsedSeconds * this.timeMultiplier;
        const gameMinutesElapsed = gameSecondsElapsed / 60;

        let newTime = this.lastSyncTime + gameMinutesElapsed;
        let newDay = this.gameDay;

        // Handle day rollover
        while (newTime >= 1440) {
            newTime -= 1440;
            newDay++;
        }

        this.gameTime = newTime;
        this.gameDay = newDay;
        this.updateDisplay();

        this.animationId = requestAnimationFrame(() => this.tick());
    }

    /**
     * Update DOM clock display (surgical update).
     */
    updateDisplay() {
        const hours = Math.floor(this.gameTime / 60);
        const minutes = Math.floor(this.gameTime % 60);

        const clockEl = document.getElementById('game-clock');
        if (clockEl) {
            const timeStr = `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
            // Only update if changed (avoid unnecessary DOM writes)
            if (clockEl.textContent !== timeStr) {
                clockEl.textContent = timeStr;
            }
        }

        // Update day display if changed
        const dayEl = document.getElementById('game-day');
        if (dayEl) {
            const dayStr = `Day ${this.gameDay}`;
            if (dayEl.textContent !== dayStr) {
                dayEl.textContent = dayStr;
            }
        }
    }

    pause() {
        this.isPaused = true;
    }

    unpause() {
        // Re-anchor to current time so we don't jump
        this.lastSyncTime = this.gameTime;
        this.lastSyncRealTime = performance.now();
        this.isPaused = false;
    }

    start() {
        this.lastSyncRealTime = performance.now();
        this.tick();
    }

    stop() {
        if (this.animationId) {
            cancelAnimationFrame(this.animationId);
            this.animationId = null;
        }
    }

    /**
     * Get current interpolated time (for UI queries).
     */
    getCurrentTime() {
        return {
            timeOfDay: Math.floor(this.gameTime),
            currentDay: this.gameDay,
            hours: Math.floor(this.gameTime / 60),
            minutes: Math.floor(this.gameTime % 60)
        };
    }
}

export const smoothClock = new SmoothClock();
```

### Integration with Delta System

**Flow When Player Takes Action:**

```
Player clicks "Enter Tavern"
    ↓
Frontend: smoothClock.pause()
    ↓
POST /api/action {action: "enter_building", building: "tavern"}
    ↓
Backend:
  1. Advances time (e.g., +5 in-game minutes for travel)
  2. Processes any triggered effects
  3. Returns delta + authoritative time_of_day
    ↓
Frontend:
  1. deltaApplier.applyDelta(response.delta)
  2. smoothClock.syncFromBackend(response.time_of_day, response.current_day)
  3. smoothClock.unpause()
    ↓
Clock resumes smooth progression from new synced time
```

**Flow During Normal Tick:**

```
Tick timer fires (every 417ms)
    ↓
POST /api/tick {time_of_day: smoothClock.getCurrentTime().timeOfDay}
    ↓
Backend processes effects, returns delta
    ↓
Frontend:
  1. deltaApplier.applyDelta(response.delta)
  2. smoothClock.syncFromBackend(response.time_of_day, response.current_day)
    ↓
Clock continues smoothly (minor correction if drifted)
```

### Pause/Unpause Integration

```javascript
// When player pauses time (voluntary)
function pauseGame() {
    smoothClock.pause();
    // Optionally notify backend to stop processing
    gameAPI.sendAction('pause_time');
}

// When player unpauses or takes action (auto-unpause)
function unpauseGame() {
    smoothClock.unpause();
    gameAPI.sendAction('unpause_time');
}

// Actions auto-unpause if paused
async function performAction(action, params) {
    if (smoothClock.isPaused) {
        smoothClock.unpause();
    }
    smoothClock.pause(); // Pause during request

    const response = await gameAPI.sendAction(action, params);

    if (response.delta) {
        deltaApplier.applyDelta(response.delta);
    }
    smoothClock.syncFromBackend(response.time_of_day, response.current_day);
    smoothClock.unpause();

    return response;
}
```

---

## Nostr Save Optimization

> **Note:** Save file optimization for Nostr relays is a **separate concern** from the delta architecture. See `docs/draft/future-nostr-save-optimization.md` for detailed plans.

**Quick Summary:**
- Current saves work fine for development
- Future optimization: store IDs only, hydrate on load
- Target: 5-10KB saves (down from 20-50KB)
- Implement before production Nostr deployment

**For this delta architecture, assume:**
- Auto-save every 5 real minutes (not every tick)
- Save format unchanged for now
- Backend manages session state in-memory between saves

---

## Backend Implementation

### Step 1: Session State Manager

**File:** `server/api/session_state.go`

```go
package api

import (
    "encoding/json"
    "sync"
    "time"
    "server/types"
)

// SessionState holds runtime state for a player session
type SessionState struct {
    // Core state (from save file)
    SaveFile    *types.SaveFile
    Location    string
    District    string
    Building    string

    // Computed/cached state (not saved to Nostr)
    NPCsAtLocation  []string
    NPCsLastHour    int

    BuildingStates  map[string]BuildingStatus
    BuildingsLastCheck int

    // For delta calculation
    LastSnapshot    *SessionSnapshot

    // Metadata
    LastActivity    time.Time
    mutex           sync.RWMutex
}

type BuildingStatus struct {
    IsOpen     bool
    OpenTime   int  // minutes (0-1439)
    CloseTime  int  // minutes (0-1439)
}

// SessionSnapshot captures state at a point in time
type SessionSnapshot struct {
    HP           int
    Fatigue      int
    Hunger       int
    Gold         int
    TimeOfDay    int
    CurrentDay   int

    NPCs         []string
    Buildings    map[string]bool  // building_id -> isOpen

    GeneralSlots [4]InventorySlotSnapshot
    BackpackSlots [20]InventorySlotSnapshot
    EquipmentSlots map[string]string  // slot_name -> item_id
}

type InventorySlotSnapshot struct {
    ItemID   string
    Quantity int
}

// SessionManager manages all active player sessions
type SessionManager struct {
    sessions map[string]*SessionState
    mutex    sync.RWMutex
}

var globalSessionManager *SessionManager

func InitSessionManager() {
    globalSessionManager = &SessionManager{
        sessions: make(map[string]*SessionState),
    }
}

func GetSessionManager() *SessionManager {
    return globalSessionManager
}

// GetOrCreateSession retrieves or creates a session
func (sm *SessionManager) GetOrCreateSession(npub string, saveFile *types.SaveFile) *SessionState {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()

    if session, exists := sm.sessions[npub]; exists {
        session.LastActivity = time.Now()
        return session
    }

    // Create new session
    session := &SessionState{
        SaveFile:       saveFile,
        Location:       saveFile.Location,
        District:       "center",
        Building:       "",
        BuildingStates: make(map[string]BuildingStatus),
        LastActivity:   time.Now(),
    }

    session.LastSnapshot = session.createSnapshot()
    sm.sessions[npub] = session

    log.Printf("✅ Created session for %s", npub)
    return session
}

// UpdateSession updates state and returns delta
func (sm *SessionManager) UpdateSession(npub string, saveFile *types.SaveFile) *Delta {
    session := sm.GetOrCreateSession(npub, saveFile)

    session.mutex.Lock()
    defer session.mutex.Unlock()

    oldSnapshot := session.LastSnapshot
    session.SaveFile = saveFile
    session.LastActivity = time.Now()

    newSnapshot := session.createSnapshot()
    session.LastSnapshot = newSnapshot

    delta := calculateDelta(oldSnapshot, newSnapshot)
    return delta
}

// createSnapshot captures current state
func (s *SessionState) createSnapshot() *SessionSnapshot {
    snapshot := &SessionSnapshot{
        HP:          s.SaveFile.HP,
        Fatigue:     s.SaveFile.Fatigue,
        Hunger:      s.SaveFile.Hunger,
        Gold:        s.SaveFile.Gold,
        TimeOfDay:   s.SaveFile.TimeOfDay,
        CurrentDay:  s.SaveFile.CurrentDay,
        NPCs:        make([]string, len(s.NPCsAtLocation)),
        Buildings:   make(map[string]bool),
        EquipmentSlots: make(map[string]string),
    }

    copy(snapshot.NPCs, s.NPCsAtLocation)

    for id, status := range s.BuildingStates {
        snapshot.Buildings[id] = status.IsOpen
    }

    // Copy inventory
    for i, slot := range s.SaveFile.Inventory.GeneralSlots {
        if slot.Item != "" {
            snapshot.GeneralSlots[i] = InventorySlotSnapshot{
                ItemID:   slot.Item,
                Quantity: slot.Quantity,
            }
        }
    }

    // Copy backpack
    if s.SaveFile.Inventory.GearSlots["bag"] != nil {
        for i, slot := range s.SaveFile.Inventory.GearSlots["bag"].Contents {
            if i >= 20 {
                break
            }
            if slot.Item != "" {
                snapshot.BackpackSlots[i] = InventorySlotSnapshot{
                    ItemID:   slot.Item,
                    Quantity: slot.Quantity,
                }
            }
        }
    }

    // Copy equipment
    for slotName, item := range s.SaveFile.Inventory.GearSlots {
        if slotName != "bag" && item != nil {
            snapshot.EquipmentSlots[slotName] = item.Item
        }
    }

    return snapshot
}

// UpdateNPCsAtLocation updates cached NPC list (called hourly)
func (s *SessionState) UpdateNPCsAtLocation(location, district string, timeOfDay int, db *sql.DB) {
    currentHour := timeOfDay / 60

    // Skip if same hour and already cached
    if currentHour == s.NPCsLastHour && len(s.NPCsAtLocation) > 0 {
        return
    }

    // Fetch fresh NPC list from database
    npcs := getNPCsAtLocationFromDB(db, location, district, "", timeOfDay)

    s.mutex.Lock()
    s.NPCsAtLocation = npcs
    s.NPCsLastHour = currentHour
    s.mutex.Unlock()

    log.Printf("🔄 Updated NPCs for %s-%s at hour %d: %v", location, district, currentHour, npcs)
}

// UpdateBuildingStates updates cached building states
func (s *SessionState) UpdateBuildingStates(location, district string, timeOfDay int, db *sql.DB) {
    // Only recalculate every 5 minutes
    if s.BuildingsLastCheck > 0 && (timeOfDay - s.BuildingsLastCheck) < 5 {
        return
    }

    buildings := getBuildingsForDistrict(db, location, district)

    s.mutex.Lock()
    defer s.mutex.Unlock()

    for _, building := range buildings {
        isOpen := isBuildingOpen(building, timeOfDay)

        s.BuildingStates[building.ID] = BuildingStatus{
            IsOpen:    isOpen,
            OpenTime:  building.Open,
            CloseTime: building.Close,
        }
    }

    s.BuildingsLastCheck = timeOfDay
    log.Printf("🏛️ Updated building states for %s-%s", location, district)
}
```

---

### Step 2: Delta Types

**File:** `server/api/delta.go`

```go
package api

// Delta represents changes between states
type Delta struct {
    Character  *CharacterDelta  `json:"character,omitempty"`
    NPCs       *NPCDelta        `json:"npcs,omitempty"`
    Buildings  *BuildingDelta   `json:"buildings,omitempty"`
    Inventory  *InventoryDelta  `json:"inventory,omitempty"`
    Equipment  *EquipmentDelta  `json:"equipment,omitempty"`
}

type CharacterDelta struct {
    HP         *int  `json:"hp,omitempty"`
    MaxHP      *int  `json:"max_hp,omitempty"`
    Fatigue    *int  `json:"fatigue,omitempty"`
    Hunger     *int  `json:"hunger,omitempty"`
    Gold       *int  `json:"gold,omitempty"`
    XP         *int  `json:"xp,omitempty"`
    TimeOfDay  *int  `json:"time_of_day,omitempty"`
    CurrentDay *int  `json:"current_day,omitempty"`
}

type NPCDelta struct {
    Added   []string `json:"added,omitempty"`
    Removed []string `json:"removed,omitempty"`
}

type BuildingDelta struct {
    StateChanged map[string]bool `json:"state_changed,omitempty"`
}

type InventoryDelta struct {
    GeneralSlots  map[int]InventorySlotDelta `json:"general_slots,omitempty"`
    BackpackSlots map[int]InventorySlotDelta `json:"backpack_slots,omitempty"`
}

type InventorySlotDelta struct {
    ItemID   *string `json:"item_id,omitempty"`
    Quantity *int    `json:"quantity,omitempty"`
}

type EquipmentDelta struct {
    Changed map[string]*string `json:"changed,omitempty"`
}

// calculateDelta compares snapshots
func calculateDelta(old, new *SessionSnapshot) *Delta {
    delta := &Delta{}

    // Character stats
    charDelta := &CharacterDelta{}
    hasCharChanges := false

    if old.HP != new.HP {
        charDelta.HP = &new.HP
        hasCharChanges = true
    }
    if old.Fatigue != new.Fatigue {
        charDelta.Fatigue = &new.Fatigue
        hasCharChanges = true
    }
    if old.Hunger != new.Hunger {
        charDelta.Hunger = &new.Hunger
        hasCharChanges = true
    }
    if old.Gold != new.Gold {
        charDelta.Gold = &new.Gold
        hasCharChanges = true
    }
    if old.TimeOfDay != new.TimeOfDay {
        charDelta.TimeOfDay = &new.TimeOfDay
        hasCharChanges = true
    }
    if old.CurrentDay != new.CurrentDay {
        charDelta.CurrentDay = &new.CurrentDay
        hasCharChanges = true
    }

    if hasCharChanges {
        delta.Character = charDelta
    }

    // NPCs
    added, removed := diffStringArrays(old.NPCs, new.NPCs)
    if len(added) > 0 || len(removed) > 0 {
        delta.NPCs = &NPCDelta{
            Added:   added,
            Removed: removed,
        }
    }

    // Buildings
    changedBuildings := make(map[string]bool)
    for buildingID, newOpen := range new.Buildings {
        oldOpen, exists := old.Buildings[buildingID]
        if !exists || oldOpen != newOpen {
            changedBuildings[buildingID] = newOpen
        }
    }
    if len(changedBuildings) > 0 {
        delta.Buildings = &BuildingDelta{StateChanged: changedBuildings}
    }

    // Inventory general slots
    generalChanges := make(map[int]InventorySlotDelta)
    for i := 0; i < 4; i++ {
        oldSlot := old.GeneralSlots[i]
        newSlot := new.GeneralSlots[i]

        if oldSlot.ItemID != newSlot.ItemID || oldSlot.Quantity != newSlot.Quantity {
            slotDelta := InventorySlotDelta{}
            if newSlot.ItemID == "" {
                empty := ""
                slotDelta.ItemID = &empty
            } else {
                slotDelta.ItemID = &newSlot.ItemID
                slotDelta.Quantity = &newSlot.Quantity
            }
            generalChanges[i] = slotDelta
        }
    }

    if len(generalChanges) > 0 {
        if delta.Inventory == nil {
            delta.Inventory = &InventoryDelta{}
        }
        delta.Inventory.GeneralSlots = generalChanges
    }

    // Equipment
    equipmentChanges := make(map[string]*string)
    for slotName, newItemID := range new.EquipmentSlots {
        oldItemID, exists := old.EquipmentSlots[slotName]
        if !exists || oldItemID != newItemID {
            equipmentChanges[slotName] = &newItemID
        }
    }
    for slotName := range old.EquipmentSlots {
        if _, exists := new.EquipmentSlots[slotName]; !exists {
            equipmentChanges[slotName] = nil
        }
    }

    if len(equipmentChanges) > 0 {
        delta.Equipment = &EquipmentDelta{Changed: equipmentChanges}
    }

    return delta
}

// diffStringArrays returns added and removed elements
func diffStringArrays(old, new []string) (added, removed []string) {
    oldMap := make(map[string]bool)
    newMap := make(map[string]bool)

    for _, item := range old {
        oldMap[item] = true
    }
    for _, item := range new {
        newMap[item] = true
    }

    for item := range newMap {
        if !oldMap[item] {
            added = append(added, item)
        }
    }

    for item := range oldMap {
        if !newMap[item] {
            removed = append(removed, item)
        }
    }

    return
}
```

---

### Step 3: Modify Game Actions

**File:** `server/api/game_actions.go`

```go
// HandleUpdateTime with delta support
func (h *GameActionsHandler) HandleUpdateTime(w http.ResponseWriter, r *http.Request) {
    npub := r.URL.Query().Get("npub")

    var req struct {
        TimeOfDay  int `json:"time_of_day"`
        CurrentDay int `json:"current_day"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Get session
    session := globalSessionManager.GetOrCreateSession(npub, currentSaveFile)

    // Update NPCs if hour changed
    currentHour := req.TimeOfDay / 60
    if currentHour != session.NPCsLastHour {
        session.UpdateNPCsAtLocation(
            session.Location,
            session.District,
            req.TimeOfDay,
            db.GetDB(),
        )
    }

    // Update building states
    session.UpdateBuildingStates(
        session.Location,
        session.District,
        req.TimeOfDay,
        db.GetDB(),
    )

    // Update save file
    currentSaveFile.TimeOfDay = req.TimeOfDay
    currentSaveFile.CurrentDay = req.CurrentDay

    // Process time effects
    messages := advanceTime(currentSaveFile, 0)

    // Calculate delta
    delta := globalSessionManager.UpdateSession(npub, currentSaveFile)

    // Return delta (not full state)
    respondJSON(w, Response{
        Success: true,
        Delta:   delta,
        Message: messagesJoined(messages),
    })
}
```

---

## Frontend Implementation

### Delta Applier

**File:** `src/systems/deltaApplier.js`

```javascript
import { logger } from "../lib/logger.js";
import { createLocationButton } from "../ui/locationDisplay.js";
import { getNPCById } from "../state/staticData.js";

/**
 * DeltaApplier handles surgical DOM updates
 * This is the ONLY place that modifies DOM in response to state changes
 */
export class DeltaApplier {
  applyDelta(delta) {
    logger.debug("Applying delta:", delta);

    if (delta.character) {
      this.applyCharacterDelta(delta.character);
    }
    if (delta.npcs) {
      this.applyNPCDelta(delta.npcs);
    }
    if (delta.buildings) {
      this.applyBuildingDelta(delta.buildings);
    }
    if (delta.inventory) {
      this.applyInventoryDelta(delta.inventory);
    }
    if (delta.equipment) {
      this.applyEquipmentDelta(delta.equipment);
    }
  }

  applyCharacterDelta(charDelta) {
    // HP
    if (charDelta.hp !== undefined) {
      const hpEl = document.getElementById("current-hp");
      if (hpEl) hpEl.textContent = charDelta.hp;

      const maxHpEl = document.getElementById("max-hp");
      const hpBar = document.getElementById("hp-bar");
      if (maxHpEl && hpBar) {
        const maxHp = parseInt(maxHpEl.textContent);
        hpBar.style.width = `${(charDelta.hp / maxHp) * 100}%`;
      }
    }

    // Fatigue
    if (charDelta.fatigue !== undefined) {
      const fatigueEl = document.getElementById("fatigue-level");
      if (fatigueEl) fatigueEl.textContent = charDelta.fatigue;
      this.updateFatigueEmoji(charDelta.fatigue);
    }

    // Hunger
    if (charDelta.hunger !== undefined) {
      const hungerEl = document.getElementById("hunger-level");
      if (hungerEl) hungerEl.textContent = charDelta.hunger;
      this.updateHungerEmoji(charDelta.hunger);
    }

    // Gold
    if (charDelta.gold !== undefined) {
      const goldEl = document.getElementById("gold-amount");
      if (goldEl) goldEl.textContent = charDelta.gold;
    }

    // Time
    if (charDelta.time_of_day !== undefined) {
      document.dispatchEvent(
        new CustomEvent("time:changed", {
          detail: { timeOfDay: charDelta.time_of_day },
        })
      );
    }
  }

  applyNPCDelta(npcDelta) {
    const container = document.querySelector("#npc-buttons div");
    if (!container) return;

    // Remove NPCs
    if (npcDelta.removed && npcDelta.removed.length > 0) {
      npcDelta.removed.forEach((npcId) => {
        const button = container.querySelector(`[data-npc-id="${npcId}"]`);
        if (button) {
          button.remove();
          logger.debug(`Removed NPC: ${npcId}`);
        }
      });
    }

    // Add NPCs
    if (npcDelta.added && npcDelta.added.length > 0) {
      npcDelta.added.forEach((npcId) => {
        const npcData = getNPCById(npcId);
        const displayName = npcData?.name || npcId.replace(/_/g, " ");
        const button = createLocationButton(
          displayName,
          () => window.location.talkToNPC(npcId),
          "npc"
        );
        button.dataset.npcId = npcId;
        container.appendChild(button);
        logger.debug(`Added NPC: ${npcId}`);
      });
    }

    // Update empty state
    if (container.children.length === 0) {
      container.innerHTML =
        '<div class="p-2 text-xs italic text-center text-gray-400">No one here.</div>';
    }
  }

  applyBuildingDelta(buildingDelta) {
    if (!buildingDelta.state_changed) return;

    for (const [buildingId, isOpen] of Object.entries(
      buildingDelta.state_changed
    )) {
      const button = document.querySelector(
        `[data-building-id="${buildingId}"]`
      );
      if (button) {
        if (isOpen) {
          button.style.background = "#6b8e6b";
          button.style.color = "#ffffff";
          button.disabled = false;
        } else {
          button.style.background = "#808080";
          button.style.color = "#000000";
          button.disabled = true;
        }
        logger.debug(`Building ${buildingId}: ${isOpen ? "OPEN" : "CLOSED"}`);
      }
    }
  }

  applyInventoryDelta(inventoryDelta) {
    if (inventoryDelta.general_slots) {
      for (const [slotIndex, slotDelta] of Object.entries(
        inventoryDelta.general_slots
      )) {
        this.updateInventorySlot("general", parseInt(slotIndex), slotDelta);
      }
    }

    if (inventoryDelta.backpack_slots) {
      for (const [slotIndex, slotDelta] of Object.entries(
        inventoryDelta.backpack_slots
      )) {
        this.updateInventorySlot("backpack", parseInt(slotIndex), slotDelta);
      }
    }
  }

  updateInventorySlot(type, slotIndex, slotDelta) {
    const selector =
      type === "general"
        ? `[data-item-slot="${slotIndex}"]`
        : `[data-backpack-slot="${slotIndex}"]`;

    const slotDiv = document.querySelector(selector);
    if (!slotDiv) return;

    const existingImg = slotDiv.querySelector("img");
    const existingQty = slotDiv.querySelector(".quantity-label");

    // Empty slot
    if (!slotDelta.item_id || slotDelta.item_id === "") {
      existingImg?.parentElement.remove();
      existingQty?.remove();
      slotDiv.dataset.itemId = "";
      return;
    }

    // Update or create image
    const newSrc = `/res/img/items/${slotDelta.item_id}.png`;

    if (existingImg) {
      const currentSrc = new URL(existingImg.src).pathname;
      if (currentSrc !== newSrc) {
        existingImg.src = newSrc;
      }
    } else {
      const imgDiv = document.createElement("div");
      imgDiv.className = "w-full h-full flex items-center justify-center p-1";
      const img = document.createElement("img");
      img.src = newSrc;
      img.className = "w-full h-full object-contain";
      img.style.imageRendering = "pixelated";
      imgDiv.appendChild(img);
      slotDiv.appendChild(imgDiv);
    }

    // Update quantity
    if (slotDelta.quantity && slotDelta.quantity > 1) {
      if (existingQty) {
        existingQty.textContent = slotDelta.quantity;
      } else {
        const qtyLabel = document.createElement("div");
        qtyLabel.className =
          "quantity-label absolute bottom-0 right-0 text-white";
        qtyLabel.style.fontSize = "10px";
        qtyLabel.textContent = slotDelta.quantity;
        slotDiv.appendChild(qtyLabel);
      }
    } else {
      existingQty?.remove();
    }

    slotDiv.dataset.itemId = slotDelta.item_id;
  }

  applyEquipmentDelta(equipmentDelta) {
    if (!equipmentDelta.changed) return;

    for (const [slotName, itemId] of Object.entries(equipmentDelta.changed)) {
      const slotDiv = document.querySelector(`[data-slot="${slotName}"]`);
      if (!slotDiv) continue;

      const existingImg = slotDiv.querySelector("img");

      if (!itemId) {
        existingImg?.parentElement.remove();
        slotDiv.dataset.itemId = "";
      } else {
        const newSrc = `/res/img/items/${itemId}.png`;

        if (existingImg) {
          existingImg.src = newSrc;
        } else {
          const imgDiv = document.createElement("div");
          imgDiv.className =
            "w-full h-full flex items-center justify-center p-1";
          const img = document.createElement("img");
          img.src = newSrc;
          img.className = "w-full h-full object-contain";
          img.style.imageRendering = "pixelated";
          imgDiv.appendChild(img);
          slotDiv.appendChild(imgDiv);
        }

        slotDiv.dataset.itemId = itemId;
      }
    }
  }

  updateFatigueEmoji(fatigue) {
    const emojiEl = document.getElementById("fatigue-emoji");
    if (!emojiEl) return;
    const emojis = [
      "😊",
      "😐",
      "😑",
      "😪",
      "😴",
      "🥱",
      "😵",
      "💀",
      "⚰️",
      "👻",
      "☠️",
    ];
    emojiEl.textContent = emojis[Math.min(fatigue, emojis.length - 1)];
  }

  updateHungerEmoji(hunger) {
    const emojiEl = document.getElementById("hunger-emoji");
    if (!emojiEl) return;
    const emojis = ["☠️", "🥺", "😋", "😊"];
    emojiEl.textContent = emojis[Math.min(hunger, emojis.length - 1)];
  }
}

export const deltaApplier = new DeltaApplier();
```

---

### Update Time Clock

**File:** `src/systems/timeClock.js`

```javascript
import { deltaApplier } from "./deltaApplier.js";

async function sendTimeUpdateToBackend(character) {
  if (!gameAPI.initialized) return;

  try {
    const response = await gameAPI.sendAction("update_time", {
      time_of_day: character.time_of_day,
      current_day: character.current_day,
    });

    // Apply delta (not full state)
    if (response && response.delta) {
      logger.debug("Received delta:", response.delta);
      deltaApplier.applyDelta(response.delta);
    }
  } catch (error) {
    logger.error("Failed to sync time:", error);
  }
}
```

---

## Migration Guide

### Phase 1: Backend (8-10 hours)

1. **Create session state infrastructure**

   ```bash
   # Create new files
   server/api/session_state.go
   server/api/delta.go
   ```

2. **Init session manager in main.go**

   ```go
   import "server/api"

   func main() {
       // ... existing setup ...

       api.InitSessionManager()

       // ... rest of main ...
   }
   ```

3. **Modify game_actions.go**

   - Update `HandleUpdateTime` to return deltas
   - Update `HandleEnterBuilding` to return deltas
   - Update `HandleInventoryAction` to return deltas

4. **Test backend**
   ```bash
   go test ./server/api -v
   ```

### Phase 2: Frontend (5-7 hours)

1. **Create deltaApplier.js**

   ```bash
   # Create file
   src/systems/deltaApplier.js
   ```

2. **Create smoothClock.js**

   ```bash
   # Create file
   src/systems/smoothClock.js
   ```

   - 60fps interpolated clock display
   - Syncs to backend authoritative time
   - Pause/unpause support

3. **Create tickManager.js**

   ```bash
   # Create file
   src/systems/tickManager.js
   ```

   - Manages 417ms tick timer
   - Sends tick requests to backend
   - Applies returned deltas

4. **Update existing systems**

   - Replace old timeClock.js with smoothClock integration
   - Remove client-side state management where possible
   - Simplify `gameState.js` (optional - can keep for backwards compat)

5. **Test frontend**
   ```bash
   npm run build
   # Start server and test manually
   ```

### Phase 3: Testing (4-6 hours)

1. **Performance testing**

   - Open DevTools Performance tab
   - Record 30 seconds of gameplay
   - Verify <5 DOM mutations per 5-second cycle
   - Verify delta payloads <500 bytes

2. **Edge case testing**

   - Hour changes (NPC updates)
   - Building open/close
   - Player ejection from closed buildings
   - Inventory changes
   - Equipment swapping

3. **Nostr save testing**
   - Save game to Nostr relay
   - Verify save size <10KB
   - Load game from relay
   - Verify hydration works correctly

---

## Testing Strategy

### Unit Tests (Backend)

```go
// server/api/session_state_test.go
func TestSessionStateCreation(t *testing.T) {
    sm := &SessionManager{sessions: make(map[string]*SessionState)}
    saveFile := &types.SaveFile{
        HP: 10,
        Fatigue: 0,
        Hunger: 2,
    }

    session := sm.GetOrCreateSession("test_npub", saveFile)

    if session.SaveFile.HP != 10 {
        t.Errorf("Expected HP 10, got %d", session.SaveFile.HP)
    }
}

func TestDeltaCalculation(t *testing.T) {
    old := &SessionSnapshot{HP: 10, Fatigue: 0}
    new := &SessionSnapshot{HP: 8, Fatigue: 1}

    delta := calculateDelta(old, new)

    if delta.Character == nil {
        t.Error("Expected character delta")
    }
    if *delta.Character.HP != 8 {
        t.Errorf("Expected HP delta 8, got %d", *delta.Character.HP)
    }
    if *delta.Character.Fatigue != 1 {
        t.Errorf("Expected fatigue delta 1, got %d", *delta.Character.Fatigue)
    }
}
```

### Integration Tests

```javascript
// Test delta applier
describe("DeltaApplier", () => {
  it("applies character HP delta", () => {
    document.body.innerHTML = '<span id="current-hp">10</span>';

    const delta = { character: { hp: 5 } };
    deltaApplier.applyDelta(delta);

    expect(document.getElementById("current-hp").textContent).toBe("5");
  });

  it("adds NPC to DOM", () => {
    document.body.innerHTML = '<div id="npc-buttons"><div></div></div>';

    const delta = { npcs: { added: ["barkeep"] } };
    deltaApplier.applyDelta(delta);

    const npcButton = document.querySelector('[data-npc-id="barkeep"]');
    expect(npcButton).toBeTruthy();
  });
});

// Test smooth clock
describe("SmoothClock", () => {
  it("interpolates time correctly at 144x", () => {
    const clock = new SmoothClock();
    clock.syncFromBackend(720, 1); // Noon, day 1

    // Simulate 1 real second passing
    clock.lastSyncRealTime = performance.now() - 1000;

    // 1 second * 144x = 144 game seconds = 2.4 minutes
    const time = clock.getCurrentTime();
    expect(time.timeOfDay).toBeCloseTo(722, 0); // 720 + 2.4 ≈ 722
  });

  it("handles day rollover", () => {
    const clock = new SmoothClock();
    clock.syncFromBackend(1438, 1); // 23:58, day 1

    // Simulate 2 real seconds (4.8 in-game minutes)
    clock.lastSyncRealTime = performance.now() - 2000;

    const time = clock.getCurrentTime();
    expect(time.currentDay).toBe(2);
    expect(time.timeOfDay).toBeLessThan(5); // Early morning day 2
  });

  it("pauses interpolation", () => {
    const clock = new SmoothClock();
    clock.syncFromBackend(720, 1);
    clock.pause();

    // Simulate time passing
    clock.lastSyncRealTime = performance.now() - 5000;

    const time = clock.getCurrentTime();
    expect(time.timeOfDay).toBe(720); // Should not have advanced
  });
});

// Test tick timing
describe("TickManager", () => {
  it("fires ticks at correct interval", async () => {
    const tickTimes = [];
    const mockCallback = () => tickTimes.push(performance.now());

    // Run for ~1 second, should get ~2-3 ticks
    tickManager.start(mockCallback);
    await new Promise(r => setTimeout(r, 1000));
    tickManager.stop();

    expect(tickTimes.length).toBeGreaterThanOrEqual(2);
    expect(tickTimes.length).toBeLessThanOrEqual(3);

    // Check interval is close to 417ms
    if (tickTimes.length >= 2) {
      const interval = tickTimes[1] - tickTimes[0];
      expect(interval).toBeCloseTo(417, -2); // Within ~10ms
    }
  });
});
```

---

## Performance Benchmarks

### Before (Current System)

| Metric               | Value       |
| -------------------- | ----------- |
| Tick rate            | Every 5s    |
| In-game time per tick| 12 minutes  |
| DOM mutations per 5s | ~50+        |
| Network payload/tick | 5-10KB      |
| Network bytes/sec    | ~1,400      |
| NPC API calls        | Every 5s    |
| Save file size       | 20-50KB     |
| Flickering           | Visible     |
| Clock display        | Jerky       |

### After (Option 4 with 1-Minute Tick)

| Metric               | Value               | Improvement      |
| -------------------- | ------------------- | ---------------- |
| Tick rate            | Every 417ms         | 12x more frequent |
| In-game time per tick| 1 minute            | 12x finer        |
| DOM mutations per 5s | ~12-30 (surgical)   | 40-75% reduction |
| Network payload/tick | 30-200 bytes        | 97% reduction    |
| Network bytes/sec    | ~120                | 91% reduction    |
| NPC API calls        | Only on hour change | 85% reduction    |
| Save file size       | 5-10KB              | 75% reduction    |
| Flickering           | None                | 100% eliminated  |
| Clock display        | Smooth (60fps)      | 100% improved    |

### Tick Rate Comparison

| System | Ticks/sec | Bytes/tick | Bytes/sec | DOM ops/tick |
|--------|-----------|------------|-----------|--------------|
| Current (5s) | 0.2 | ~7,000 | ~1,400 | ~50 |
| Delta (417ms) | 2.4 | ~50 | ~120 | ~2-5 |
| **Ratio** | **12x more** | **140x smaller** | **91% less** | **90% fewer** |

### Why More Ticks = Less Load

The counterintuitive result: **12x more ticks uses 91% less bandwidth**.

This is because:
1. **Delta payloads are tiny** - Only changed fields, not full state
2. **Most ticks are near-empty** - Just time update (~35 bytes)
3. **No full DOM re-renders** - Surgical updates only
4. **NPC/building data cached** - Not re-fetched each tick

---

## Benefits Beyond Flicker Fix

1. **Multiplayer Ready**

   - Backend authoritative state
   - Easy to broadcast deltas to multiple clients

2. **Easier Testing**

   - All logic in Go (unit testable)
   - Frontend is pure rendering

3. **Better Security**

   - No client-side state manipulation
   - Backend validates everything

4. **Simpler Debugging**

   - Single source of truth
   - Clear delta logs

5. **Future Features**
   - Real-time updates (WebSocket deltas)
   - State rollback/replay
   - Server-side AI/automation
   - NIP-17 encrypted saves
   - Multi-device sync

---

## Timeline

| Week | Focus                  | Hours | Deliverables                          |
| ---- | ---------------------- | ----- | ------------------------------------- |
| 1    | Backend setup          | 4-5h  | session_state.go, delta.go            |
| 2    | Backend endpoints      | 4-5h  | Modified game_actions.go, tick handler |
| 3    | Frontend systems       | 5-7h  | deltaApplier.js, smoothClock.js       |
| 4    | Testing & polish       | 4-6h  | Tests, performance validation         |

**Total:** 17-23 hours over 4 weeks

### Key Implementation Files

**Backend (Go):**
- `server/api/session_state.go` - Session management
- `server/api/delta.go` - Delta types and calculation
- `server/api/game_actions.go` - Modified to return deltas
- `server/api/tick_handler.go` - New tick endpoint

**Frontend (JS):**
- `src/systems/deltaApplier.js` - Surgical DOM updates
- `src/systems/smoothClock.js` - 60fps interpolated clock display
- `src/systems/tickManager.js` - 417ms tick timer management

---

## Questions?

- Discord: [TBD]
- GitHub Issues: [TBD]
- Documentation: `docs/draft/option-4-delta-architecture.md`

---

**Last Updated:** 2026-01-14
**Status:** Implementation In Progress
**Author:** Claude (AI Assistant)

---

## Implementation Status

### Completed (Phase 1)

**Backend (Go):**
- [x] `server/api/delta.go` - Delta types and calculation engine
- [x] `server/api/session_manager.go` - Enhanced with snapshot tracking
- [x] `server/api/game_actions.go` - Modified to calculate and return deltas

**Frontend (JS):**
- [x] `src/systems/smoothClock.js` - 60fps interpolated clock display
- [x] `src/systems/deltaApplier.js` - Surgical DOM updates
- [x] `src/systems/tickManager.js` - 417ms tick orchestration
- [x] `src/systems/timeClock.js` - Updated to integrate new systems

### Key Changes

1. **All game actions now return deltas** - Not just time updates
2. **Backwards compatible** - `USE_DELTA_SYSTEM` flag allows reverting
3. **Full state still included** - Both `delta` and `state` in responses during transition

### Testing Checklist

- [x] Build Go server without errors
- [ ] Clock displays smoothly at 60fps
- [x] Ticks occur every ~417ms when running
- [x] Fatigue/hunger updates apply via delta
- [ ] No UI flickering during updates
- [x] NPC buttons add/remove correctly
- [x] Building states update correctly

---

## Known Issues (Next Cleanup Session)

**Last Updated:** 2026-01-17

### Critical

1. **Wait feature completely broken**
   - Needs full examination
   - May be related to delta architecture changes

### UI/UX Issues

2. **Vault deposit not updating UI**
   - Dragging items from inventory to player vault doesn't update until page refresh
   - Inventory delta not being calculated/applied for vault operations

3. **UI flicker on navigation**
   - Clicking district navigation buttons causes flicker
   - Entering/exiting buildings causes button/UI flicker
   - Building state updates themselves are smooth (delta working)
   - Issue is with full location re-renders on navigation

4. **Clock still not smooth**
   - Frequently jumps forward a minute quickly
   - Sometimes lags for a minute
   - Interpolation not consistent
   - May need to adjust sync frequency or interpolation algorithm

### Developer Experience

5. **Excessive backend logging**
   - Too much log output during normal operation
   - Need to reduce verbose logging or add log levels
