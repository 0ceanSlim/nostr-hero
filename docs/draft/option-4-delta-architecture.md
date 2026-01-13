# Option 4: Go Backend State + Delta Updates

**Implementation Guide for UI Optimization & Nostr Integration**

---

## Executive Summary

This document outlines the implementation of a delta-based architecture that:
1. **Fixes UI flickering** (99% reduction in DOM mutations)
2. **Optimizes Nostr save files** (5-10KB instead of 20-50KB)
3. **Aligns with Go-first philosophy** (backend authoritative, minimal JS)
4. **Enables future features** (multiplayer, real-time updates, state rollback)

**Estimated Effort:** 16-24 hours
**Target Completion:** 4 weeks

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Architecture Overview](#architecture-overview)
3. [Nostr Integration](#nostr-integration)
4. [Backend Implementation](#backend-implementation)
5. [Frontend Implementation](#frontend-implementation)
6. [Migration Guide](#migration-guide)
7. [Testing Strategy](#testing-strategy)
8. [Performance Benchmarks](#performance-benchmarks)

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

## Nostr Integration

### Save File Optimization for Nostr

**Current Save Format (Too Large):**

```json
{
  "inventory": {
    "general_slots": [
      {
        "item": {
          "id": "longsword",
          "name": "Longsword",
          "description": "A versatile martial weapon...",
          "damage": "1d8",
          "damage_type": "slashing",
          "weight": 3,
          "value": 15,
          "rarity": "common",
          "properties": ["versatile"],
          "image": "longsword.png"
        },
        "quantity": 1
      }
    ]
  }
}
```
**Size:** ~500 bytes per item × 20 slots = **10KB** just for inventory!

---

**Optimized Save Format (Minimal):**

```json
{
  "hp": 15,
  "f": 3,           // fatigue (shortened field names)
  "h": 2,           // hunger
  "g": 250,         // gold
  "xp": 100,
  "loc": "kingdom",
  "dist": "center",
  "bld": "",
  "inv": [
    {"id": "longsword", "q": 1},
    {"id": "health-potion", "q": 3}
  ],
  "eq": {
    "mainHand": "longsword",
    "armor": "leather-armor"
  },
  "sp": ["fire-bolt", "magic-missile"],
  "eff": [
    {"id": "fatigue-accumulation", "d": -1}
  ],
  "t": 720,
  "d": 1
}
```
**Size:** ~500 bytes total for entire save!

---

### Save/Load Flow with Nostr

#### 1. Loading Game from Nostr

```
User clicks "Load Game"
  ↓
Backend fetches save from Nostr relay (minimal JSON, ~5-10KB)
  ↓
Backend HYDRATES save:
  1. Parse minimal save JSON
  2. Load item data from game-data/items/*.json
     - longsword.json → full stats, description, image
  3. Load spell data from game-data/magic/spells/*.json
  4. Load location data from game-data/locations/*.json
  5. Load NPC data from game-data/npcs/*.json
  6. Calculate NPC positions (schedules + current time)
  7. Calculate building states (open/close times + current time)
  8. Compute inventory weight
  ↓
Backend creates rich SessionState (in-memory, ~100KB+)
  ↓
Backend returns initial DELTA to frontend (full state on first load)
  ↓
Frontend applies delta and renders UI
```

#### 2. During Gameplay (Every 5 seconds)

```
Time clock ticks
  ↓
Frontend sends time update to backend
  ↓
Backend updates SessionState:
  - Process effects (fatigue++, hunger--)
  - Update NPC positions if hour changed
  - Update building states if time changed
  ↓
Backend calculates DELTA (only changes)
  Example delta: {"character": {"fatigue": 4}}  (~50 bytes)
  ↓
Backend returns delta to frontend
  ↓
Frontend applies surgical DOM update
  ↓
⚠️ NO SAVE TO NOSTR (would spam relays)
```

#### 3. Auto-Save (Every 5 minutes)

```
5-minute timer triggers
  ↓
Backend serializes SessionState to minimal save format
  - Extract item IDs (not full data)
  - Extract spell IDs
  - Extract location IDs
  - Compress JSON
  ↓
Backend publishes save to Nostr relay as event
  - Event kind: TBD (custom kind for game saves)
  - Content: minimal JSON (~5-10KB)
  - Optional: Encrypt with NIP-17
  ↓
Relay stores event
  ↓
Other devices can fetch latest save
```

---

### Nostr Save File Types

**Go Types:**

```go
// Optimized for Nostr storage
type NostrSaveFile struct {
    HP         int                    `json:"hp"`
    Fatigue    int                    `json:"f"`    // Shortened
    Hunger     int                    `json:"h"`
    Gold       int                    `json:"g"`
    XP         int                    `json:"xp"`

    Location   string                 `json:"loc"`
    District   string                 `json:"dist"`
    Building   string                 `json:"bld"`

    Items      []InventorySlotMinimal `json:"inv"`
    Equipment  map[string]string      `json:"eq"`
    Spells     []string               `json:"sp"`
    Effects    []ActiveEffectMinimal  `json:"eff"`

    Time       int                    `json:"t"`
    Day        int                    `json:"d"`
}

type InventorySlotMinimal struct {
    ID  string `json:"id"`   // "longsword"
    Q   int    `json:"q"`    // quantity
}

type ActiveEffectMinimal struct {
    ID  string `json:"id"`   // "fatigue-accumulation"
    D   int    `json:"d"`    // duration (-1 = permanent)
}
```

**Estimated Sizes:**
- Minimal save: 5-10KB
- Compressed (gzip): 2-5KB
- Encrypted (NIP-17): 3-6KB
- **Well within Nostr 64KB limit** ✅

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
import { logger } from '../lib/logger.js';
import { createLocationButton } from '../ui/locationDisplay.js';
import { getNPCById } from '../state/staticData.js';

/**
 * DeltaApplier handles surgical DOM updates
 * This is the ONLY place that modifies DOM in response to state changes
 */
export class DeltaApplier {
    applyDelta(delta) {
        logger.debug('Applying delta:', delta);

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
            const hpEl = document.getElementById('current-hp');
            if (hpEl) hpEl.textContent = charDelta.hp;

            const maxHpEl = document.getElementById('max-hp');
            const hpBar = document.getElementById('hp-bar');
            if (maxHpEl && hpBar) {
                const maxHp = parseInt(maxHpEl.textContent);
                hpBar.style.width = `${(charDelta.hp / maxHp) * 100}%`;
            }
        }

        // Fatigue
        if (charDelta.fatigue !== undefined) {
            const fatigueEl = document.getElementById('fatigue-level');
            if (fatigueEl) fatigueEl.textContent = charDelta.fatigue;
            this.updateFatigueEmoji(charDelta.fatigue);
        }

        // Hunger
        if (charDelta.hunger !== undefined) {
            const hungerEl = document.getElementById('hunger-level');
            if (hungerEl) hungerEl.textContent = charDelta.hunger;
            this.updateHungerEmoji(charDelta.hunger);
        }

        // Gold
        if (charDelta.gold !== undefined) {
            const goldEl = document.getElementById('gold-amount');
            if (goldEl) goldEl.textContent = charDelta.gold;
        }

        // Time
        if (charDelta.time_of_day !== undefined) {
            document.dispatchEvent(new CustomEvent('time:changed', {
                detail: { timeOfDay: charDelta.time_of_day }
            }));
        }
    }

    applyNPCDelta(npcDelta) {
        const container = document.querySelector('#npc-buttons div');
        if (!container) return;

        // Remove NPCs
        if (npcDelta.removed && npcDelta.removed.length > 0) {
            npcDelta.removed.forEach(npcId => {
                const button = container.querySelector(`[data-npc-id="${npcId}"]`);
                if (button) {
                    button.remove();
                    logger.debug(`Removed NPC: ${npcId}`);
                }
            });
        }

        // Add NPCs
        if (npcDelta.added && npcDelta.added.length > 0) {
            npcDelta.added.forEach(npcId => {
                const npcData = getNPCById(npcId);
                const displayName = npcData?.name || npcId.replace(/_/g, ' ');
                const button = createLocationButton(
                    displayName,
                    () => window.location.talkToNPC(npcId),
                    'npc'
                );
                button.dataset.npcId = npcId;
                container.appendChild(button);
                logger.debug(`Added NPC: ${npcId}`);
            });
        }

        // Update empty state
        if (container.children.length === 0) {
            container.innerHTML = '<div class="text-gray-400 text-xs p-2 text-center italic">No one here.</div>';
        }
    }

    applyBuildingDelta(buildingDelta) {
        if (!buildingDelta.state_changed) return;

        for (const [buildingId, isOpen] of Object.entries(buildingDelta.state_changed)) {
            const button = document.querySelector(`[data-building-id="${buildingId}"]`);
            if (button) {
                if (isOpen) {
                    button.style.background = '#6b8e6b';
                    button.style.color = '#ffffff';
                    button.disabled = false;
                } else {
                    button.style.background = '#808080';
                    button.style.color = '#000000';
                    button.disabled = true;
                }
                logger.debug(`Building ${buildingId}: ${isOpen ? 'OPEN' : 'CLOSED'}`);
            }
        }
    }

    applyInventoryDelta(inventoryDelta) {
        if (inventoryDelta.general_slots) {
            for (const [slotIndex, slotDelta] of Object.entries(inventoryDelta.general_slots)) {
                this.updateInventorySlot('general', parseInt(slotIndex), slotDelta);
            }
        }

        if (inventoryDelta.backpack_slots) {
            for (const [slotIndex, slotDelta] of Object.entries(inventoryDelta.backpack_slots)) {
                this.updateInventorySlot('backpack', parseInt(slotIndex), slotDelta);
            }
        }
    }

    updateInventorySlot(type, slotIndex, slotDelta) {
        const selector = type === 'general'
            ? `[data-item-slot="${slotIndex}"]`
            : `[data-backpack-slot="${slotIndex}"]`;

        const slotDiv = document.querySelector(selector);
        if (!slotDiv) return;

        const existingImg = slotDiv.querySelector('img');
        const existingQty = slotDiv.querySelector('.quantity-label');

        // Empty slot
        if (!slotDelta.item_id || slotDelta.item_id === '') {
            existingImg?.parentElement.remove();
            existingQty?.remove();
            slotDiv.dataset.itemId = '';
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
            const imgDiv = document.createElement('div');
            imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
            const img = document.createElement('img');
            img.src = newSrc;
            img.className = 'w-full h-full object-contain';
            img.style.imageRendering = 'pixelated';
            imgDiv.appendChild(img);
            slotDiv.appendChild(imgDiv);
        }

        // Update quantity
        if (slotDelta.quantity && slotDelta.quantity > 1) {
            if (existingQty) {
                existingQty.textContent = slotDelta.quantity;
            } else {
                const qtyLabel = document.createElement('div');
                qtyLabel.className = 'quantity-label absolute bottom-0 right-0 text-white';
                qtyLabel.style.fontSize = '10px';
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

            const existingImg = slotDiv.querySelector('img');

            if (!itemId) {
                existingImg?.parentElement.remove();
                slotDiv.dataset.itemId = '';
            } else {
                const newSrc = `/res/img/items/${itemId}.png`;

                if (existingImg) {
                    existingImg.src = newSrc;
                } else {
                    const imgDiv = document.createElement('div');
                    imgDiv.className = 'w-full h-full flex items-center justify-center p-1';
                    const img = document.createElement('img');
                    img.src = newSrc;
                    img.className = 'w-full h-full object-contain';
                    img.style.imageRendering = 'pixelated';
                    imgDiv.appendChild(img);
                    slotDiv.appendChild(imgDiv);
                }

                slotDiv.dataset.itemId = itemId;
            }
        }
    }

    updateFatigueEmoji(fatigue) {
        const emojiEl = document.getElementById('fatigue-emoji');
        if (!emojiEl) return;
        const emojis = ['😊', '😐', '😑', '😪', '😴', '🥱', '😵', '💀', '⚰️', '👻', '☠️'];
        emojiEl.textContent = emojis[Math.min(fatigue, emojis.length - 1)];
    }

    updateHungerEmoji(hunger) {
        const emojiEl = document.getElementById('hunger-emoji');
        if (!emojiEl) return;
        const emojis = ['☠️', '🥺', '😋', '😊'];
        emojiEl.textContent = emojis[Math.min(hunger, emojis.length - 1)];
    }
}

export const deltaApplier = new DeltaApplier();
```

---

### Update Time Clock

**File:** `src/systems/timeClock.js`

```javascript
import { deltaApplier } from './deltaApplier.js';

async function sendTimeUpdateToBackend(character) {
    if (!gameAPI.initialized) return;

    try {
        const response = await gameAPI.sendAction('update_time', {
            time_of_day: character.time_of_day,
            current_day: character.current_day
        });

        // Apply delta (not full state)
        if (response && response.delta) {
            logger.debug('Received delta:', response.delta);
            deltaApplier.applyDelta(response.delta);
        }

    } catch (error) {
        logger.error('Failed to sync time:', error);
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

### Phase 2: Frontend (4-6 hours)

1. **Create deltaApplier.js**
   ```bash
   # Create file
   src/systems/deltaApplier.js
   ```

2. **Update timeClock.js**
   - Import deltaApplier
   - Replace full state updates with delta application

3. **Remove client-side state caching**
   - Simplify `gameState.js` (optional - can keep for backwards compat)

4. **Test frontend**
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
describe('DeltaApplier', () => {
    it('applies character HP delta', () => {
        document.body.innerHTML = '<span id="current-hp">10</span>';

        const delta = { character: { hp: 5 } };
        deltaApplier.applyDelta(delta);

        expect(document.getElementById('current-hp').textContent).toBe('5');
    });

    it('adds NPC to DOM', () => {
        document.body.innerHTML = '<div id="npc-buttons"><div></div></div>';

        const delta = { npcs: { added: ['barkeep'] } };
        deltaApplier.applyDelta(delta);

        const npcButton = document.querySelector('[data-npc-id="barkeep"]');
        expect(npcButton).toBeTruthy();
    });
});
```

---

## Performance Benchmarks

### Before (Current System)

| Metric | Value |
|--------|-------|
| DOM mutations per 5s | ~50+ |
| Network payload | 5-10KB |
| NPC API calls | Every 5s |
| Save file size | 20-50KB |
| Flickering | Visible |

### After (Option 4)

| Metric | Value | Improvement |
|--------|-------|-------------|
| DOM mutations per 5s | ~2-5 | 90% reduction |
| Network payload | 50-200 bytes | 95% reduction |
| NPC API calls | Only on hour change | 85% reduction |
| Save file size | 5-10KB | 75% reduction |
| Flickering | None | 100% eliminated |

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

| Week | Focus | Hours | Deliverables |
|------|-------|-------|-------------|
| 1 | Backend setup | 4-5h | session_state.go, delta.go |
| 2 | Backend endpoints | 4-5h | Modified game_actions.go |
| 3 | Frontend delta applier | 4-6h | deltaApplier.js, updated timeClock.js |
| 4 | Testing & polish | 4-6h | Tests, performance validation |

**Total:** 16-22 hours over 4 weeks

---

## Questions?

- Discord: [TBD]
- GitHub Issues: [TBD]
- Documentation: `docs/draft/option-4-delta-architecture.md`

---

**Last Updated:** 2026-01-13
**Status:** Draft
**Author:** Claude (AI Assistant)
