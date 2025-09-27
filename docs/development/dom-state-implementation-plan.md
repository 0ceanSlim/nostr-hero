# DOM-Based State Management & DuckDB Implementation Plan

## Revised Architecture: Client-Side State + Minimal API

### Core Concept
- **Game State Lives in DOM**: All current character state stored in hidden DOM elements
- **Static Game Data from DB**: Items, spells, monsters served from DuckDB
- **Manual Save to Nostr**: Player clicks "Save Game" → current DOM state → Nostr relay
- **Minimal Backend**: Just data serving + Nostr relay interaction

### Why This Approach is Better
1. **Simpler Backend**: No complex game state APIs needed
2. **Instant Responsiveness**: No network calls for game actions
3. **Offline Capable**: Game works without internet (until save)
4. **Nostr Native**: Saves directly to decentralized relays
5. **HTMX Friendly**: Perfect for your current tech stack

## Revised Architecture

### Backend Responsibilities (Minimal)
```go
// Game data serving only
GET /api/items              → All items from DuckDB
GET /api/spells             → All spells from DuckDB
GET /api/monsters           → All monsters from DuckDB
GET /api/locations          → All locations from DuckDB
GET /api/game-data          → Combined game data bundle

// Character & save management
GET /api/character/{npub}   → Generate/load character
POST /api/save-game         → Save state to Nostr relay
GET /api/load-save/{npub}   → Load saved state from relay
```

### Frontend Responsibilities (Everything Else)
```html
<!-- Game state stored in hidden DOM elements -->
<div id="game-state" style="display: none;">
    <div id="character-data">
        {"hp": 45, "max_hp": 50, "mana": 12, "location": "waterdeep"}
    </div>
    <div id="inventory-data">
        [{"item": "longsword", "quantity": 1, "equipped": true}]
    </div>
    <div id="spell-data">
        [{"spell": "magic-missile", "prepared": true}]
    </div>
    <div id="location-data">
        {"current": "waterdeep", "discovered": ["waterdeep", "neverwinter"]}
    </div>
    <div id="combat-data">
        null
    </div>
</div>

<!-- Static game data loaded once -->
<div id="static-data" style="display: none;">
    <div id="all-items"><!-- All items from DB --></div>
    <div id="all-spells"><!-- All spells from DB --></div>
    <div id="all-monsters"><!-- All monsters from DB --></div>
    <div id="all-locations"><!-- All locations from DB --></div>
</div>
```

## Simplified Database Schema

### Focus on Static Game Data Only

```sql
-- Items (weapons, armor, consumables, etc.)
CREATE TABLE items (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    item_type VARCHAR NOT NULL,
    properties JSON,
    tags VARCHAR[],
    rarity VARCHAR DEFAULT 'common',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Spells
CREATE TABLE spells (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    level INTEGER NOT NULL,
    school VARCHAR NOT NULL,
    damage VARCHAR,
    mana_cost INTEGER,
    classes VARCHAR[],
    properties JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Monsters
CREATE TABLE monsters (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    challenge_rating DECIMAL(3,1),
    stats JSON,
    actions JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Locations
CREATE TABLE locations (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    location_type VARCHAR,
    description TEXT,
    properties JSON,
    connections JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Character Generation Data (for deterministic creation)
CREATE TABLE character_classes (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    hit_die INTEGER,
    spell_progression JSON,
    starting_equipment JSON
);

CREATE TABLE races (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    ability_modifiers JSON,
    traits JSON
);

-- No character instances, inventory, or game sessions!
-- All that lives in DOM and Nostr relays
```

## Game State Management in DOM

### Core Game Loop with Hyperscript

```html
<div id="game-container">
    <!-- Character Panel -->
    <div id="character-panel" _="on load call updateCharacterDisplay()">
        <div class="hp-bar">
            HP: <span id="current-hp">45</span>/<span id="max-hp">50</span>
        </div>
        <div class="mana-bar" _="on gameStateChange call updateManaDisplay()">
            Mana: <span id="current-mana">12</span>/<span id="max-mana">15</span>
        </div>
    </div>

    <!-- Location Panel -->
    <div id="location-panel" _="on load call displayCurrentLocation()">
        <h3 id="location-name">Waterdeep</h3>
        <p id="location-description"></p>
        <div id="location-actions"></div>
    </div>

    <!-- Action Buttons -->
    <div id="action-panel">
        <button _="on click call moveToLocation('neverwinter')">
            Travel North
        </button>
        <button _="on click call useItem('health-potion')">
            Use Health Potion
        </button>
        <button _="on click call castSpell('magic-missile')">
            Cast Magic Missile
        </button>
    </div>

    <!-- Save Game Button -->
    <button id="save-btn"
            _="on click call saveGameToRelay()"
            class="save-button">
        💾 Save Game
    </button>
</div>
```

### Game Logic in JavaScript

```html
<script>
// Game state management
function getGameState() {
    return {
        character: JSON.parse(document.getElementById('character-data').textContent),
        inventory: JSON.parse(document.getElementById('inventory-data').textContent),
        spells: JSON.parse(document.getElementById('spell-data').textContent),
        location: JSON.parse(document.getElementById('location-data').textContent),
        combat: JSON.parse(document.getElementById('combat-data').textContent)
    };
}

function updateGameState(newState) {
    if (newState.character) {
        document.getElementById('character-data').textContent = JSON.stringify(newState.character);
    }
    if (newState.inventory) {
        document.getElementById('inventory-data').textContent = JSON.stringify(newState.inventory);
    }
    if (newState.spells) {
        document.getElementById('spell-data').textContent = JSON.stringify(newState.spells);
    }
    if (newState.location) {
        document.getElementById('location-data').textContent = JSON.stringify(newState.location);
    }
    if (newState.combat) {
        document.getElementById('combat-data').textContent = JSON.stringify(newState.combat);
    }

    // Trigger UI updates
    document.dispatchEvent(new CustomEvent('gameStateChange'));
}

// Game actions
function moveToLocation(locationId) {
    const state = getGameState();
    const locationData = getLocationById(locationId);

    if (!locationData) return;

    // Update location
    state.location.current = locationId;
    if (!state.location.discovered.includes(locationId)) {
        state.location.discovered.push(locationId);
    }

    // Handle travel fatigue, time passage, etc.
    state.character.fatigue = calculateFatigue(state.character.fatigue, locationData.travel_time);

    updateGameState(state);
    displayCurrentLocation();
}

function useItem(itemId) {
    const state = getGameState();
    const item = findInInventory(state.inventory, itemId);
    const itemData = getItemById(itemId);

    if (!item || item.quantity <= 0) {
        showMessage("You don't have that item!");
        return;
    }

    // Apply item effects
    if (itemData.properties.heal) {
        state.character.hp = Math.min(
            state.character.max_hp,
            state.character.hp + itemData.properties.heal
        );
    }

    // Remove item from inventory
    item.quantity -= 1;
    if (item.quantity === 0) {
        state.inventory = state.inventory.filter(i => i.item !== itemId);
    }

    updateGameState(state);
    showMessage(`Used ${itemData.name}!`);
}

function castSpell(spellId) {
    const state = getGameState();
    const spell = findInSpells(state.spells, spellId);
    const spellData = getSpellById(spellId);

    if (!spell || !spell.prepared) {
        showMessage("Spell not prepared!");
        return;
    }

    if (state.character.mana < spellData.mana_cost) {
        showMessage("Not enough mana!");
        return;
    }

    // Cast spell
    state.character.mana -= spellData.mana_cost;

    // Apply spell effects (damage, healing, etc.)
    if (state.combat) {
        handleCombatSpell(spellId, spellData);
    } else {
        handleNonCombatSpell(spellId, spellData);
    }

    updateGameState(state);
}

// Save to Nostr relay
async function saveGameToRelay() {
    const state = getGameState();
    const npub = getCurrentNpub();

    const saveData = {
        npub: npub,
        timestamp: Date.now(),
        gameState: state,
        version: "1.0"
    };

    try {
        const response = await fetch('/api/save-game', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(saveData)
        });

        if (response.ok) {
            showMessage("✅ Game saved to Nostr relay!");
            document.getElementById('save-btn').textContent = "💾 Saved!";
            setTimeout(() => {
                document.getElementById('save-btn').textContent = "💾 Save Game";
            }, 2000);
        } else {
            showMessage("❌ Failed to save game");
        }
    } catch (error) {
        showMessage("❌ Error saving game: " + error.message);
    }
}

// Static data helpers
function getItemById(itemId) {
    const allItems = JSON.parse(document.getElementById('all-items').textContent);
    return allItems.find(item => item.id === itemId);
}

function getSpellById(spellId) {
    const allSpells = JSON.parse(document.getElementById('all-spells').textContent);
    return allSpells.find(spell => spell.id === spellId);
}

function getLocationById(locationId) {
    const allLocations = JSON.parse(document.getElementById('all-locations').textContent);
    return allLocations.find(location => location.id === locationId);
}
</script>
```

## Minimal Backend Implementation

### Simplified API Layer

```go
// src/api/gamedata.go
package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
)

// Serve all game data in one request for efficient loading
func GameDataHandler(w http.ResponseWriter, r *http.Request) {
    db, err := GetDB()
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    gameData := struct {
        Items     []Item     `json:"items"`
        Spells    []Spell    `json:"spells"`
        Monsters  []Monster  `json:"monsters"`
        Locations []Location `json:"locations"`
    }{}

    // Load all static data
    gameData.Items, _ = LoadAllItems(db)
    gameData.Spells, _ = LoadAllSpells(db)
    gameData.Monsters, _ = LoadAllMonsters(db)
    gameData.Locations, _ = LoadAllLocations(db)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(gameData)
}

// Character generation (existing, just simplified)
func CharacterHandler(w http.ResponseWriter, r *http.Request) {
    npub := r.URL.Query().Get("npub")

    // Generate character deterministically
    character := GenerateCharacterFromNpub(npub)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(character)
}

// Save game state to Nostr relay
func SaveGameHandler(w http.ResponseWriter, r *http.Request) {
    var saveData struct {
        Npub      string      `json:"npub"`
        Timestamp int64       `json:"timestamp"`
        GameState interface{} `json:"gameState"`
        Version   string      `json:"version"`
    }

    if err := json.NewDecoder(r.Body).Decode(&saveData); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Save to Nostr relay
    err := SaveToNostrRelay(saveData.Npub, saveData)
    if err != nil {
        http.Error(w, "Failed to save to relay", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}
```

### Game Initialization

```html
<!-- Initial page load -->
<div id="game-app" _="on load call initializeGame()">
    <!-- Game UI here -->
</div>

<script>
async function initializeGame() {
    const npub = getCurrentNpub();

    // Load static game data
    const gameDataResponse = await fetch('/api/game-data');
    const gameData = await gameDataResponse.json();

    // Store static data in DOM
    document.getElementById('all-items').textContent = JSON.stringify(gameData.items);
    document.getElementById('all-spells').textContent = JSON.stringify(gameData.spells);
    document.getElementById('all-monsters').textContent = JSON.stringify(gameData.monsters);
    document.getElementById('all-locations').textContent = JSON.stringify(gameData.locations);

    // Try to load existing save
    try {
        const saveResponse = await fetch(`/api/load-save/${npub}`);
        if (saveResponse.ok) {
            const saveData = await saveResponse.json();
            initializeFromSave(saveData.gameState);
        } else {
            // Create new character
            await createNewCharacter(npub);
        }
    } catch {
        // Create new character if load fails
        await createNewCharacter(npub);
    }

    // Start game
    displayCurrentLocation();
    updateAllDisplays();
}

async function createNewCharacter(npub) {
    const characterResponse = await fetch(`/api/character?npub=${npub}`);
    const character = await characterResponse.json();

    // Initialize fresh game state
    const initialState = {
        character: character,
        inventory: generateStartingInventory(character.class),
        spells: generateStartingSpells(character.class),
        location: {
            current: character.starting_location || 'waterdeep',
            discovered: [character.starting_location || 'waterdeep']
        },
        combat: null
    };

    updateGameState(initialState);
}
</script>
```

## Benefits of This Approach

### 1. **Extreme Simplicity**
- **Backend**: Just data serving + Nostr integration
- **No complex APIs**: No game state management on server
- **No database sessions**: All state in DOM

### 2. **Performance**
- **Instant Actions**: No network calls for game mechanics
- **Batch Loading**: All static data loaded once
- **Offline Play**: Works without internet until save

### 3. **Scalability**
- **Stateless Server**: Easy to deploy/scale
- **Client Processing**: Offload computation to player's browser
- **Nostr Decentralized**: No central save database needed

### 4. **Developer Experience**
- **Simple Debugging**: Game state visible in DOM
- **Easy Testing**: Manipulate DOM directly
- **HTMX Native**: Perfect for your existing approach

## Implementation Steps

### Phase 1: Database Setup (Week 1)
- [ ] Set up DuckDB with static data schema
- [ ] Migrate JSON files to database
- [ ] Create data serving APIs

### Phase 2: DOM State Management (Week 2)
- [ ] Build game state DOM structure
- [ ] Implement core JavaScript game functions
- [ ] Create character generation flow

### Phase 3: Game Mechanics (Week 3-4)
- [ ] Movement system
- [ ] Inventory management
- [ ] Spell casting
- [ ] Combat system

### Phase 4: Nostr Integration (Week 5)
- [ ] Save game to relay
- [ ] Load game from relay
- [ ] Handle save conflicts

### Phase 5: Polish (Week 6)
- [ ] Error handling
- [ ] UI improvements
- [ ] Performance optimization

This approach is much more aligned with your HTMX/Hyperscript philosophy and eliminates the need for complex backend game state management!