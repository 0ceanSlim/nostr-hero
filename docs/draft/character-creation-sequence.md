# Character Creation Sequence for Nostr Hero

## New Player Flow (First Time with Pubkey)

### 1. **Pubkey Detection & Welcome**
```
GET /start?pubkey=npub1...
-> Check if pubkey exists in database
-> If new: redirect to character creation
-> If exists: redirect to game world
```

### 2. **Character Creation Steps**

#### Step 1: Race Selection
- Display all available races with descriptions
- Show racial bonuses and starting locations
- Submit race choice

#### Step 2: Class Selection
- Display all classes with descriptions
- Show class abilities, spell access, and playstyle
- Submit class choice

#### Step 3: Starting Stats
- Roll stats or point buy system
- Show race + class modifiers
- Confirm final stats

#### Step 4: Starting Gear Selection ⭐ **THIS IS THE FOCUS**
- Load class-specific starting gear from `starting-gear.json`
- Present each option set as interactive choices
- For spellcasters: include spell component selections
- For spellcasters: display starting spells from `starting-spells.json`
- Generate final inventory

#### Step 5: Starting Location
- Based on race selection from `racial-starting-cities.json`
- Place character in appropriate starting city
- Initialize save file

## Starting Gear Selection Flow

### Data Sources:
- `/www/data/starting-gear.json` - Equipment options by class
- `/www/data/starting-spells.json` - Spell lists by class
- `/www/data/items/*.json` - Individual item data

### Selection Process:
1. **Load Class Gear Options**
   - Parse starting-gear.json for selected class
   - Identify option sets vs guaranteed items

2. **Present Interactive Choices**
   - Each "option" array becomes a choice selection
   - Radio buttons or dropdown for single choice
   - Checkboxes for multiple selections (if applicable)

3. **Spellcaster Additions**
   - If class has spellcasting: show starting spells
   - Display spell descriptions and mechanics
   - Show material component requirements

4. **Generate Final Inventory**
   - Combine all selections + guaranteed items
   - Create character save with inventory
   - Redirect to starting location

## Example: Wizard Creation Flow

### Step 4A: Weapon Choice
```
Choose your starting weapon:
○ Quarterstaff - Simple melee weapon, versatile
○ Dagger - Light, finesse, throwable weapon
```

### Step 4B: Magical Focus
```
Choose your spellcasting focus:
○ Wand - Elegant and precise
○ Orb - Channels raw magical energy
○ Rod - Sturdy and reliable
○ Staff - Versatile, can be used as quarterstaff
○ Crystal - Pure magical resonance
```

### Step 4C: Spell Components
```
Choose your spell component approach:
○ Component Pouch - Universal pouch for most spells
○ Bat Guano and Sulfur (2x) - Specific materials for Fireball
```

### Step 4D: Equipment Pack
```
Choose your equipment pack:
○ Scholar's Pack - Books, ink, parchment for research
○ Explorer's Pack - Survival gear for adventuring
```

### Step 4E: Starting Spells Display
```
Your Starting Spells:

CANTRIPS (Unlimited use):
• Fire Bolt - 1d10 fire damage, ranged attack
• Ray of Frost - 1d8 cold damage, slows enemy
• Shocking Grasp - 1d8 lightning, prevents reactions

SPELLS IN SPELLBOOK (6 total):
• Magic Missile - 3d4+3 force damage, always hits
• Shield - +5 AC as reaction when attacked
• Burning Hands - 3d6 fire damage in cone area
• Cure Wounds - 1d8+3 healing, touch range
• Thunderwave - 2d8 thunder damage, pushes enemies
• Fireball - 8d6 fire damage, area effect (requires Bat Guano & Sulfur)

GUARANTEED ITEMS:
• Spellbook - Contains all your known spells
• Bat Guano and Sulfur (1x) - For casting Fireball
```

## Technical Implementation Notes

### HTMX Structure:
- Each selection triggers `hx-post` to update character state
- Server maintains temporary character creation session
- Progressive enhancement - each step unlocks the next
- Final step commits to database and redirects to game

### Data Flow:
1. `GET /create-character` - Start creation
2. `POST /create-character/race` - Select race
3. `POST /create-character/class` - Select class
4. `POST /create-character/stats` - Assign stats
5. `POST /create-character/gear` - Make gear selections
6. `POST /create-character/finalize` - Create save & start game

### Session Management:
- Store partial character in server session
- Each step validates previous selections
- Can navigate back to previous steps
- Clear session on completion or timeout