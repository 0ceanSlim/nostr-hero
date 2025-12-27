# Martial Class Abilities System

## Overview

A progression system that gives non-spellcasting classes (Fighter, Barbarian, Monk, Rogue) unique combat abilities that unlock with level progression and scale in power as the character levels up.

**Key Features**:
- Each martial class gets 6 unique abilities
- Abilities unlock at specific level milestones (1, 3, 5, 7-8, 10, 12-15)
- Abilities become more powerful every 4-5 levels
- Each class has unique resource management (Stamina, Rage, Ki, Cunning)
- Resources regenerate during combat (unlike spell slots)
- Balances martials vs spellcasters

---

## Design Philosophy

**Why This System?**
- Spellcasters have versatility (many spells, scrolls, utility)
- Martials need reliability (consistent power, always-available abilities)
- Level progression should feel rewarding (new abilities + existing ones get stronger)
- Each class should play differently (unique resources and mechanics)

**Balance Goals**:
- Martials can't run out of resources like casters run out of mana
- Abilities regenerate in-combat (sustainable)
- Power scales linearly with level (no dead levels)
- No tracking between combats (resources reset)

---

## Resource Systems

### Fighter - Stamina
- **Max**: 10 points
- **Regeneration**: +2 per turn
- **Theme**: Endurance and tactical combat maneuvers
- **Playstyle**: Sustained damage dealer with self-healing

### Barbarian - Rage
- **Max**: 100 points
- **Regeneration**: +10 per hit taken
- **Theme**: Build-up mechanic, gets stronger when damaged
- **Playstyle**: Tank that converts damage into devastating attacks

### Monk - Ki Points
- **Max**: Wisdom modifier + Character Level
- **Regeneration**: +1 per turn
- **Theme**: Spiritual energy, precise strikes
- **Playstyle**: High mobility, many small attacks, crowd control

### Rogue - Cunning Points
- **Max**: 10 points
- **Regeneration**: +2 per turn, +1 per critical hit
- **Theme**: Opportunistic, rewards smart play
- **Playstyle**: Burst damage, stealth, critical hit focused

---

## Fighter Abilities

**Core Resource**: Stamina (Max 10, +2 per turn)

### 1. Second Wind (Unlocks Level 1)
**Cost**: 3 Stamina | **Uses**: Once per combat

| Level Range | Effect |
|-------------|--------|
| 1-4 | Heal 25% max HP |
| 5-9 | Heal 40% max HP |
| 10-14 | Heal 60% max HP + remove 1 debuff |
| 15+ | Heal 80% max HP + remove all debuffs + gain 20 temp HP, usable twice per combat |

**Description**: Fighter catches their breath and recovers from wounds mid-battle.

---

### 2. Power Strike (Unlocks Level 3)
**Cost**: 2 Stamina | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 3-6 | Next attack deals 150% damage |
| 7-11 | Next attack deals 200% damage |
| 12-16 | Next attack deals 250% damage + ignores armor |
| 17+ | Next attack deals 300% damage + ignores armor + cleaves to nearby enemy |

**Description**: Fighter puts extra force into their next strike.

---

### 3. Action Surge (Unlocks Level 5)
**Cost**: 5 Stamina (reduces to 4 at level 10) | **Uses**: Once per combat (twice at level 20)

| Level Range | Effect |
|-------------|--------|
| 5-9 | Take 2 actions this turn |
| 10-14 | Take 2 actions this turn (costs 4 Stamina) |
| 15-19 | Take 3 actions this turn |
| 20+ | Take 3 actions this turn (costs 4 Stamina), usable twice per combat |

**Description**: Fighter pushes beyond normal limits to act multiple times in rapid succession.

---

### 4. Disarming Blow (Unlocks Level 7)
**Cost**: 3 Stamina | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 7-11 | Disarm enemy for 2 turns, they deal -50% damage |
| 12-16 | Disarm enemy for 3 turns, they deal -60% damage |
| 17+ | Disarm enemy for 4 turns, they deal -75% damage, steal their weapon (can equip it) |

**Description**: Fighter strikes the enemy's weapon, forcing them to drop it.

---

### 5. Rally (Unlocks Level 10)
**Cost**: 4 Stamina | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 10-14 | Heal 10% max HP per turn for 3 turns |
| 15-19 | Heal 15% max HP per turn for 4 turns |
| 20+ | Heal 20% max HP per turn for 5 turns + gain +20% damage during effect |

**Description**: Fighter rallies themselves, recovering health over time while continuing to fight.

---

### 6. Indomitable (Unlocks Level 15)
**Cost**: 7 Stamina (reduces to 6 at level 20) | **Uses**: Once per day (twice at level 20)

| Level Range | Effect |
|-------------|--------|
| 15-17 | Cannot die for 2 turns (HP can't drop below 1) |
| 18-19 | Cannot die for 3 turns + heal 5% HP per turn |
| 20+ | Cannot die for 4 turns + heal 10% HP per turn + deal double damage, usable twice per day |

**Description**: Fighter refuses to fall, their willpower keeping them alive against impossible odds.

---

## Barbarian Abilities

**Core Resource**: Rage (Max 100, +10 per hit taken)

### 1. Enter Rage (Unlocks Level 1)
**Cost**: 50 Rage | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 1-4 | +50% damage, -25% incoming damage, lasts 4 turns |
| 5-9 | +60% damage, -35% incoming damage, lasts 5 turns |
| 10-14 | +75% damage, -45% incoming damage, lasts 6 turns, immune to fear/charm |
| 15+ | +100% damage, -50% incoming damage, lasts 8 turns, immune to fear/charm/stun |

**Description**: Barbarian enters a primal rage, becoming a devastating force on the battlefield.

---

### 2. Reckless Attack (Unlocks Level 3)
**Cost**: 20-30 Rage | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 3-6 | Next attack auto-crits, take +50% damage next turn (20 Rage) |
| 7-11 | Next attack auto-crits for 3x damage, take +30% damage next turn (20 Rage) |
| 12-16 | Next 2 attacks auto-crit for 3x damage, take +20% damage next turn (25 Rage) |
| 17+ | Next 3 attacks auto-crit for 4x damage, no damage penalty (30 Rage) |

**Description**: Barbarian attacks with reckless abandon, sacrificing defense for overwhelming offense.

---

### 3. Intimidating Roar (Unlocks Level 5)
**Cost**: 30 Rage (reduces to 25 at level 15) | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 5-9 | Enemies deal -30% damage for 2 turns |
| 10-14 | Enemies deal -40% damage for 3 turns + 20% flee chance |
| 15-19 | Enemies deal -50% damage for 4 turns + 35% flee chance (costs 25 Rage) |
| 20+ | Enemies deal -60% damage for 5 turns + 50% flee chance + can't heal |

**Description**: Barbarian unleashes a terrifying roar that shakes the resolve of all enemies.

---

### 4. Savage Leap (Unlocks Level 7)
**Cost**: 25 Rage (reduces to 20 at level 15) | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 7-11 | Jump to any enemy, deal 100% weapon damage on landing (25 Rage) |
| 12-16 | Jump to any enemy, deal 150% weapon damage in AoE around target (25 Rage) |
| 17-19 | Jump to any enemy, deal 200% weapon damage in AoE, knock enemies prone (20 Rage) |
| 20+ | Jump to any enemy, deal 250% weapon damage in AoE, knock prone, stun for 1 turn (20 Rage) |

**Description**: Barbarian leaps across the battlefield, crashing down on enemies.

---

### 5. Blood Frenzy (Unlocks Level 8)
**Cost**: Passive | **Uses**: Always Active

| Level Range | Effect |
|-------------|--------|
| 8-11 | Killing enemy restores 20% HP |
| 12-15 | Killing enemy restores 30% HP + extends Rage by 2 turns |
| 16-19 | Killing enemy restores 40% HP + extends Rage by 3 turns + gain 20 Rage |
| 20+ | Killing enemy restores 50% HP + extends Rage indefinitely + gain 30 Rage |

**Description**: The sight of blood drives the Barbarian into deeper frenzy, healing their wounds.

---

### 6. Berserker Mode (Unlocks Level 12)
**Cost**: 75 Rage (reduces to 70 at level 20) | **Uses**: Once per combat (twice at level 20)

| Level Range | Effect |
|-------------|--------|
| 12-16 | Deal +100% damage, can't drop below 1 HP for 4 turns |
| 17-19 | Deal +125% damage, can't drop below 1 HP, heal 10 HP per hit for 5 turns |
| 20+ | Deal +150% damage, can't drop below 1 HP, heal 15 HP per hit, immune to CC for 6 turns, twice per combat |

**Description**: Barbarian enters an unstoppable berserker state, becoming an immortal force of destruction.

---

## Monk Abilities

**Core Resource**: Ki Points (Max = Wisdom modifier + Level, +1 per turn)

### 1. Flurry of Blows (Unlocks Level 1)
**Cost**: 1 Ki | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 1-4 | Attack 2 times this turn |
| 5-9 | Attack 3 times this turn |
| 10-14 | Attack 4 times this turn |
| 15+ | Attack 5 times this turn |

**Description**: Monk unleashes a rapid series of strikes in a single moment.

---

### 2. Patient Defense (Unlocks Level 3)
**Cost**: 1 Ki | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 3-6 | Dodge all attacks for 1 turn |
| 7-11 | Dodge all attacks for 1 turn + reflect 25% damage back to attacker |
| 12-16 | Dodge all attacks for 2 turns + reflect 50% damage |
| 17+ | Dodge all attacks for 2 turns + reflect 100% damage + heal for damage reflected |

**Description**: Monk enters a defensive stance, evading attacks and redirecting force.

---

### 3. Stunning Strike (Unlocks Level 5)
**Cost**: 2 Ki | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 5-9 | Stun enemy for 1 turn |
| 10-14 | Stun enemy for 2 turns |
| 15-19 | Stun enemy for 2 turns + they take double damage while stunned |
| 20+ | Stun enemy for 3 turns + they take triple damage + can't act for 1 turn after stun ends |

**Description**: Monk strikes a pressure point, temporarily paralyzing the enemy.

---

### 4. Step of the Wind (Unlocks Level 7)
**Cost**: 1 Ki | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 7-11 | Teleport anywhere on battlefield |
| 12-16 | Teleport + next attack deals +50% damage |
| 17-19 | Teleport + next attack deals +100% damage + auto-crit |
| 20+ | Teleport + next 2 attacks deal +150% damage + auto-crit |

**Description**: Monk moves with supernatural speed, appearing anywhere on the battlefield.

---

### 5. Deflect Missiles (Unlocks Level 10)
**Cost**: 2 Ki | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 10-14 | Catch projectile and throw back for 150% damage |
| 15-19 | Catch projectile and throw back for 200% damage + hits 2 enemies |
| 20+ | Catch ANY attack (melee or ranged) and reflect 300% damage to all enemies |

**Description**: Monk's reflexes are so fast they can catch arrows and even deflect weapons mid-swing.

---

### 6. Quivering Palm (Unlocks Level 15)
**Cost**: 4 Ki | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 15-17 | Deal 300 damage after 3 turns (CON save to resist) |
| 18-19 | Deal 500 damage after 2 turns (CON save to resist) |
| 20+ | Instant kill if enemy HP < 30%, otherwise deal 800 damage (CON save to resist) |

**Description**: Monk touches an enemy and sets vibrations in their body that cause delayed catastrophic damage.

---

## Rogue Abilities

**Core Resource**: Cunning Points (Max 10, +2 per turn, +1 per crit)

### 1. Sneak Attack (Unlocks Level 1)
**Cost**: 3 Cunning | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 1-4 | +2d6 damage, auto-crit from stealth |
| 5-9 | +4d6 damage, auto-crit from stealth |
| 10-14 | +6d6 damage, auto-crit from stealth, 30% instant kill if HP < 20% |
| 15+ | +10d6 damage, auto-crit from stealth, 50% instant kill if HP < 30% |

**Description**: Rogue strikes at a vulnerable point, dealing massive damage.

---

### 2. Hide in Shadows (Unlocks Level 3)
**Cost**: 2 Cunning | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 3-6 | Invisible for 2 turns or until you attack |
| 7-11 | Invisible for 3 turns or until you attack, +25% crit chance while hidden |
| 12-16 | Invisible for 4 turns or until you attack, +50% crit chance, first attack doesn't break stealth |
| 17+ | Invisible for 5 turns, +75% crit chance, first 2 attacks don't break stealth |

**Description**: Rogue melts into the shadows, becoming invisible to enemies.

---

### 3. Poison Blade (Unlocks Level 5)
**Cost**: 3 Cunning | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 5-9 | Next 3 attacks deal +15 poison damage |
| 10-14 | Next 5 attacks deal +25 poison damage + slow enemy (-30% damage) |
| 15-19 | Next 7 attacks deal +40 poison damage + slow + reduce healing by 50% |
| 20+ | Next 10 attacks deal +60 poison damage + slow + no healing + DoT (5 per turn for 5 turns) |

**Description**: Rogue coats their weapon in deadly poison.

---

### 4. Evasion (Unlocks Level 8)
**Cost**: Passive | **Uses**: Always Active

| Level Range | Effect |
|-------------|--------|
| 8-12 | Take 0 damage from AoE attacks |
| 13-17 | Take 0 damage from AoE + gain 2 Cunning when evading |
| 18+ | Take 0 damage from AoE + gain 3 Cunning + reflect 50% of avoided damage back |

**Description**: Rogue's reflexes allow them to completely avoid area attacks.

---

### 5. Assassinate (Unlocks Level 10)
**Cost**: 5 Cunning | **Uses**: Once per combat (twice at level 20)

| Level Range | Effect |
|-------------|--------|
| 10-14 | Instant kill if enemy HP < 25%, otherwise 250% damage |
| 15-19 | Instant kill if enemy HP < 35%, otherwise 350% damage |
| 20+ | Instant kill if enemy HP < 50%, otherwise 500% damage, twice per combat |

**Description**: Rogue executes a perfect killing blow on a weakened enemy.

---

### 6. Shadow Step (Unlocks Level 15)
**Cost**: 4 Cunning | **Uses**: Unlimited

| Level Range | Effect |
|-------------|--------|
| 15-17 | Teleport behind enemy, next attack auto-crit for 300% damage |
| 18-19 | Teleport behind enemy, next 2 attacks auto-crit for 350% damage |
| 20+ | Teleport behind enemy, next 3 attacks auto-crit for 400% damage + apply bleed (20 damage/turn) |

**Description**: Rogue steps through shadows to appear behind their target for a devastating strike.

---

## Data Structures

### Ability Definition (`game-data/abilities/`)

Each ability is defined in a JSON file with scaling tiers:

```json
{
  "id": "second-wind",
  "name": "Second Wind",
  "class": "fighter",
  "unlock_level": 1,
  "resource_cost": 3,
  "resource_type": "stamina",
  "cooldown": "once_per_combat",
  "description": "Fighter catches their breath and recovers from wounds mid-battle.",
  "icon": "second-wind",

  "scaling_tiers": [
    {
      "min_level": 1,
      "max_level": 4,
      "effects": {
        "heal_percent": 0.25,
        "uses_per_combat": 1
      }
    },
    {
      "min_level": 5,
      "max_level": 9,
      "effects": {
        "heal_percent": 0.40,
        "uses_per_combat": 1
      }
    },
    {
      "min_level": 10,
      "max_level": 14,
      "effects": {
        "heal_percent": 0.60,
        "remove_debuffs": 1,
        "uses_per_combat": 1
      }
    },
    {
      "min_level": 15,
      "max_level": 99,
      "effects": {
        "heal_percent": 0.80,
        "remove_debuffs": "all",
        "temp_hp": 20,
        "uses_per_combat": 2
      }
    }
  ]
}
```

### Character Save Data

```json
{
  "class": "fighter",
  "level": 12,
  "stamina": 6,
  "max_stamina": 10,
  "stamina_regen": 2,

  "unlocked_abilities": [
    "second-wind",
    "power-strike",
    "action-surge",
    "disarming-blow",
    "rally"
  ],

  "ability_uses": {
    "second-wind": 1,
    "action-surge": 1,
    "indomitable": 0
  }
}
```

---

## Backend API Design

### Endpoint: Get Class Abilities

**Request**: `GET /api/abilities?class={class}&level={level}`

**Response**:
```json
{
  "success": true,
  "class": "fighter",
  "level": 12,
  "resource": {
    "type": "stamina",
    "current": 6,
    "max": 10,
    "regen_per_turn": 2
  },
  "abilities": [
    {
      "id": "second-wind",
      "name": "Second Wind",
      "cost": 3,
      "current_tier": {
        "level_range": "10-14",
        "effects": {
          "heal_percent": 0.60,
          "remove_debuffs": 1,
          "uses_per_combat": 1
        }
      },
      "next_tier": {
        "level_range": "15+",
        "unlock_at": 15,
        "effects": {
          "heal_percent": 0.80,
          "remove_debuffs": "all",
          "temp_hp": 20,
          "uses_per_combat": 2
        }
      },
      "uses_remaining": 1,
      "can_use": true
    }
  ]
}
```

### Endpoint: Use Ability

**Request**: `POST /api/abilities/use`

**Body**:
```json
{
  "npub": "npub1...",
  "save_id": "save_1234567890",
  "ability_id": "second-wind",
  "target_id": "self"
}
```

**Backend Validation**:
1. Load character from save
2. Load ability definition
3. Check class matches
4. Check level >= unlock_level
5. Check resource >= cost
6. Check uses remaining (if limited)
7. Calculate effects based on current level tier
8. Apply effects to character
9. Deduct resource cost
10. Decrement uses remaining
11. Save game state

**Success Response**:
```json
{
  "success": true,
  "ability_used": "second-wind",
  "effects_applied": {
    "healed": 45,
    "debuffs_removed": ["poisoned"]
  },
  "new_resource": 3,
  "uses_remaining": 0
}
```

**Error Response**:
```json
{
  "success": false,
  "error": "Not enough stamina (need 3, have 2)"
}
```

---

## Frontend UI Design

### Combat Abilities Panel

```
┌────────────────────────────────────────────────────────┐
│  ⚔️ Fighter Abilities                 Stamina: 6/10   │
├────────────────────────────────────────────────────────┤
│                                                         │
│  [Second Wind] (3 Stamina)           [1 use left]      │
│  Heal 60% HP + Remove 1 Debuff                         │
│  Next tier at Lv15: 80% HP + Remove All Debuffs       │
│  [USE]                                                  │
│                                                         │
│  [Power Strike] (2 Stamina)                            │
│  Next attack deals 250% damage + Ignores Armor         │
│  Next tier at Lv17: 300% + Cleave                     │
│  [USE]                                                  │
│                                                         │
│  [Action Surge] (4 Stamina)          [1 use left]      │
│  Take 2 actions this turn                              │
│  Next tier at Lv15: Take 3 actions                    │
│  [USE]                                                  │
│                                                         │
│  [Disarming Blow] (3 Stamina)                          │
│  Disarm for 3 turns, enemy deals -60% damage           │
│  Next tier at Lv17: Steal weapon                      │
│  [USE]                                                  │
│                                                         │
│  [Rally] (4 Stamina)                                   │
│  Heal 10% HP per turn for 3 turns                      │
│  Next tier at Lv15: 15% for 4 turns                   │
│  [USE]                                                  │
│                                                         │
│  🔒 Indomitable (Unlocks at Level 15)                  │
│      Cannot die for 2 turns                            │
└────────────────────────────────────────────────────────┘
```

### Level Up Notification

```
┌────────────────────────────────────────────────────────┐
│                    🎉 LEVEL UP! 🎉                     │
│                    You are now Level 10                │
├────────────────────────────────────────────────────────┤
│  🆕 New Ability Unlocked: Rally                        │
│                                                         │
│  Rally (4 Stamina)                                     │
│  Heal 10% max HP per turn for 3 turns                  │
│                                                         │
│  ⬆️ Abilities Improved:                                │
│  • Second Wind: Now heals 60% HP + removes debuffs     │
│  • Action Surge: Cost reduced to 4 Stamina             │
│  • Flurry of Blows: Now attack 4 times                │
│                                                         │
│  📊 Stats Increased:                                   │
│  • HP: 85 → 95 (+10)                                   │
│  • Max Ki: 13 → 14 (+1)                                │
│                                                         │
│  [Continue]                                            │
└────────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Data & Types
- [ ] Create ability JSON files for all 24 abilities (6 per class × 4 classes)
- [ ] Define ability scaling tiers in JSON
- [ ] Create Go types for abilities, resources, effects
- [ ] Add resource fields to character save structure

### Phase 2: Database & Migration
- [ ] Add `abilities` table to database
- [ ] Add `class_resources` table (stamina, rage, ki, cunning)
- [ ] Migrate ability data from JSON to SQLite on startup
- [ ] Add ability unlock tracking to save files

### Phase 3: Backend Logic
- [ ] Implement resource regeneration system
- [ ] Create ability validation logic (cost, uses, level requirements)
- [ ] Implement ability effect calculations based on level tier
- [ ] Create ability usage API endpoints
- [ ] Add ability list API (filtered by class/level)

### Phase 4: Combat Integration
- [ ] Add resource display to combat UI
- [ ] Create abilities panel in combat screen
- [ ] Implement ability usage in combat flow
- [ ] Add ability cooldown/uses tracking per combat
- [ ] Reset resources and uses at combat end

### Phase 5: Frontend UI
- [ ] Create ability button components
- [ ] Display current tier effects and next tier preview
- [ ] Show resource costs and availability
- [ ] Add level-up notification for new/improved abilities
- [ ] Create ability tooltips with full descriptions

### Phase 6: Balance & Polish
- [ ] Test all abilities at different level ranges
- [ ] Balance resource costs vs regeneration rates
- [ ] Tune damage numbers and effect durations
- [ ] Add visual effects for ability usage
- [ ] Create ability icons/animations

---

## Balance Considerations

### Resource Economy

**Fighter (Stamina)**:
- Regen: +2/turn, Max 10
- Can use most abilities every 1-2 turns
- Sustained combat effectiveness
- Example: Power Strike (2) + Power Strike (2) + Second Wind (3) = 7 stamina over 4 turns (regenerated)

**Barbarian (Rage)**:
- Regen: +10 per hit taken, Max 100
- Needs to take damage to build resource
- Rewards aggressive tanking
- Example: Take 5 hits = 50 rage = Enter Rage ability

**Monk (Ki)**:
- Regen: +1/turn, Max = WIS + Level (typically 15-20 at high level)
- Slow regeneration, must spend wisely
- Rewards tactical ability usage
- Example at Lv10 with WIS 14: Max 14 ki, regen 1/turn = ~3 Flurry of Blows per combat

**Rogue (Cunning)**:
- Regen: +2/turn, +1 per crit, Max 10
- Rewards crit-focused builds
- Can burst with multiple abilities quickly
- Example: Start with 0, Turn 1 = 2 cunning, Crit = 3 total, Turn 2 = 5, use Sneak Attack

### Ability Power Scaling

**Early Game (Levels 1-5)**:
- Few abilities (1-3 per class)
- Low scaling (abilities are 50-80% as powerful as late game)
- Resource management is tight

**Mid Game (Levels 6-12)**:
- Most abilities unlocked (4-5 per class)
- Moderate scaling (abilities are 100-150% base power)
- Resource pools are comfortable

**Late Game (Levels 13-20)**:
- All abilities unlocked (6 per class)
- High scaling (abilities are 200-400% base power)
- Resources abundant, can use abilities frequently
- Some abilities become unlimited use or "spam-able"

### Martial vs Caster Balance

**Spellcasters**:
- Limited mana pool
- Must return to town to restore mana (or use potions)
- Spell scrolls cost gold/components
- Versatility (many spells for different situations)

**Martials**:
- Resources regenerate in combat
- Never "run out" between fights
- No crafting costs
- Specialization (fewer abilities, but always available)

**Trade-off**:
- Casters have problem-solving tools (utility spells)
- Martials have consistent combat power (always ready)
- Casters require resource management between combats
- Martials only manage resources within combat

---

## Progression Examples

### Fighter Level 1 → 20 Journey

**Level 1**: Unlock Second Wind
- Can heal 25% HP once per combat
- Only ability available

**Level 3**: Unlock Power Strike
- Now has self-healing + burst damage
- Starting to feel like a skilled warrior

**Level 5**: Unlock Action Surge + Abilities Scale
- Second Wind heals 40% now
- Can take 2 actions in one turn (huge power spike)
- Feels like a veteran combatant

**Level 7**: Unlock Disarming Blow
- Can now control enemy damage output
- Power Strike deals 200%
- Tactical options increasing

**Level 10**: Unlock Rally + Abilities Scale
- Second Wind heals 60% + removes debuff
- Action Surge costs 4 stamina (cheaper)
- Rally provides sustained healing
- Feels nearly unkillable in prolonged fights

**Level 15**: Unlock Indomitable + Abilities Scale
- Second Wind heals 80% + removes all debuffs + temp HP, usable twice
- Action Surge gives 3 actions
- Indomitable makes you immortal for 2 turns
- Power Strike deals 250% + ignores armor
- Truly a legendary warrior

**Level 20**: All Abilities at Max Power
- Second Wind: 80% heal, all debuffs, temp HP, 2 uses
- Power Strike: 300% damage + cleave
- Action Surge: 3 actions, 2 uses
- Disarming Blow: Steal enemy weapons
- Rally: 20% heal/turn for 5 turns + bonus damage
- Indomitable: 4 turns immortal + double damage, 2 uses per day
- Unstoppable killing machine

---

### Monk Level 1 → 20 Journey

**Level 1**: Unlock Flurry of Blows
- Attack 2 times per turn (1 Ki)
- With WIS 14 + Lv1 = 15 max Ki, can use ~15 times per combat

**Level 3**: Unlock Patient Defense
- Can now dodge tank (1 Ki to dodge all attacks)
- Option between offense (Flurry) or defense (Dodge)

**Level 5**: Unlock Stunning Strike + Abilities Scale
- Flurry now attacks 3 times
- Can stun enemies (2 Ki)
- Strong crowd control

**Level 7**: Unlock Step of the Wind
- Can teleport anywhere (1 Ki)
- Incredible mobility, can focus priority targets

**Level 10**: Unlock Deflect Missiles + Abilities Scale
- Flurry attacks 4 times
- Stunning Strike lasts 2 turns
- Can catch arrows (2 Ki)
- Step of Wind adds +50% damage to next attack

**Level 15**: Unlock Quivering Palm + Abilities Scale
- Flurry attacks 5 times per turn
- Stunning Strike deals double damage
- Step of Wind auto-crits for +100% damage
- Deflect Missiles hits 2 enemies
- Quivering Palm deals 300 delayed damage
- Glass cannon with extreme mobility

**Level 20**: All Abilities at Max Power
- Flurry: 5 attacks (1 Ki) = can attack 20 times in one combat
- Patient Defense: 2 turns dodge + 100% reflect + heal
- Stunning Strike: 3 turns stun + triple damage
- Step of Wind: 2 auto-crit attacks at +150%
- Deflect Missiles: Reflect ALL attacks to ALL enemies
- Quivering Palm: Instant kill if HP < 30%
- Ultimate martial artist, untouchable and deadly

---

## Open Questions

### 1. Ability Unlocking Display
- Should abilities show as "locked" in UI before reaching level?
- Or hidden until unlocked (surprise mechanics)?

**Recommendation**: Show locked abilities with level requirement (builds anticipation)

### 2. Resource Reset Timing
- Reset at start of each combat?
- Or persist between combats (partial carry-over)?

**Recommendation**: Full reset at combat start (simpler, prevents resource hoarding)

### 3. Ability Respeccing
- Can players reset abilities if they want different build?
- Or locked once chosen (no choices in this system currently)?

**Note**: Current system has no choices, all abilities unlock automatically. May add "talent trees" later.

### 4. Visual Feedback
- What animations/effects should abilities have?
- Screen shake, particle effects, sound?

**Recommendation**: Start simple (text notifications), add effects later

---

## Future Enhancements

### Phase 2 Features (Post-MVP)

**Talent Trees**:
- At certain levels, choose between 2 variants of an ability
- Example Fighter Level 5: Choose "Offensive Surge" (3 attacks, no defense) OR "Defensive Surge" (2 actions + shield)
- Creates build diversity

**Ultimate Abilities** (Level 20):
- One super-powerful ability per class
- Usable once per day
- Feels epic and game-changing
- Examples:
  - Fighter: "Avatar of War" - 10 turns of godmode
  - Barbarian: "Primal Fury" - Transform into unstoppable beast
  - Monk: "Perfect Form" - 100% dodge, infinite Ki for 5 turns
  - Rogue: "Death Mark" - Instantly kill any enemy

**Ability Synergies**:
- Using certain abilities in sequence grants bonuses
- Example: Rogue uses Hide in Shadows → Shadow Step → Assassinate = guaranteed instant kill
- Rewards skilled play

**Passive Abilities**:
- Always-on bonuses that unlock with level
- Example: Fighter Level 8 "Weapon Expert" - +10% damage with all weapons
- Don't cost resources, just make class stronger

**Equipment Bonuses**:
- Certain items reduce ability costs or add effects
- Example: "Fighter's Gauntlets" - Power Strike costs 1 stamina instead of 2
- Creates itemization goals

---

## Summary

This martial abilities system provides:

- **Progression satisfaction** - New abilities every few levels, existing ones get stronger
- **Class identity** - Each martial class plays uniquely with distinct resources
- **Combat depth** - Tactical decisions about when to use abilities
- **Balance** - Competes with spellcasters without magic
- **No save bloat** - All calculated from level, no tracking needed beyond combat
- **Scalability** - Easy to add new abilities or tiers later

**Key Design Pillars**:
1. **Unlock & Scale** - Abilities unlock at milestones and improve automatically with level
2. **Resource Management** - Each class has unique resource that regenerates differently
3. **Power Fantasy** - High-level abilities feel incredibly powerful
4. **Always Available** - Resources regenerate in combat, never "run dry" like mana
5. **Simple to Track** - No complex cooldowns or cross-combat persistence

Next steps: Implement Phase 1 (ability data files) for all 24 abilities (6 per class × 4 classes).
