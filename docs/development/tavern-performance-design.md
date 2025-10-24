# Tavern & Performance System Design

## Musical Instruments Available
- **Lute** (3500g) - Sophisticated strings
- **Viol** (3000g) - Classical strings
- **Bagpipe** (3000g) - Loud wind
- **Lyre** (3000g) - Refined strings
- **Flute** (1200g) - Soft wind
- **Horn** (300g) - Bold brass
- **Drum** (600g) - Percussion
- **Pan Flute** (1200g) - Folk wind
- **Shawm** (200g) - Reed wind

## Tavern Profiles by Location

### 1. Kingdom - The Golden Griffin (Upscale Urban)
**Innkeeper**: Gareth Thornfield (Human, friendly)
**Doorman**: Marcus Silverstring (Half-Elf, refined)
**Rest Cost**: 50g
**Atmosphere**: Sophisticated, refined clientele
**Accepted Instruments**:
- Lute: 15g base ✓
- Viol: 16g base ✓ (favorite)
- Lyre: 14g base ✓
- Flute: 12g base ✓
- Horn: 10g base ✓
- Drum: 8g base ✓
- Bagpipe: ✗ (too loud)
- Pan Flute: ✗ (too rustic)
- Shawm: ✗ (too harsh)

---

### 2. City-East (Goldenhaven) - The Salty Anchor (Port Tavern)
**Innkeeper**: Captain Reylia Stormwave (Dragonborn, salty)
**Doorman**: "Crusher" Morg (Half-Orc, tough but fair)
**Rest Cost**: 60g (expensive port city)
**Atmosphere**: Rowdy sailors, maritime culture
**Accepted Instruments**:
- Horn: 14g base ✓ (favorite - sea shanties)
- Drum: 12g base ✓ (rhythmic work songs)
- Lute: 13g base ✓
- Flute: 10g base ✓
- Bagpipe: 11g base ✓ (sailors love it)
- Viol: ✗ (too delicate for rough crowd)
- Lyre: ✗ (too refined)
- Pan Flute: ✗ (can't hear over noise)
- Shawm: 9g base ✓ (loud enough)

---

### 3. City-South (Verdant) - The Moonlit Glade Inn (Nature-themed)
**Innkeeper**: Faelyn Nightbreeze (Elf, serene)
**Doorman**: Theren Oakwhisper (Elf, traditional)
**Rest Cost**: 45g
**Atmosphere**: Peaceful, natural, elvish culture
**Accepted Instruments**:
- Lyre: 16g base ✓ (favorite - traditional)
- Flute: 15g base ✓ (forest melodies)
- Pan Flute: 14g base ✓ (nature sounds)
- Lute: 13g base ✓
- Viol: 12g base ✓
- Horn: ✗ (too loud, disturbs peace)
- Drum: ✗ (too harsh)
- Bagpipe: ✗ (absolutely not)
- Shawm: ✗ (too discordant)

---

### 4. Town-Northeast (Ironpeak) - The Pickaxe & Pint (Miner's Tavern)
**Innkeeper**: Helga Stonebrew (Dwarf, hearty)
**Doorman**: Durin Ironlung (Dwarf, gruff)
**Rest Cost**: 40g
**Atmosphere**: Hard-working miners, drinking songs
**Accepted Instruments**:
- Drum: 15g base ✓ (favorite - work rhythms)
- Horn: 14g base ✓ (drinking songs)
- Bagpipe: 13g base ✓ (dwarven tradition)
- Lute: 10g base ✓
- Viol: 8g base ✓
- Flute: ✗ (too soft, can't hear)
- Lyre: ✗ (too fancy)
- Pan Flute: ✗ (too delicate)
- Shawm: 11g base ✓ (loud enough)

---

### 5. Village-West (Millhaven) - The Grain & Grape (Country Inn)
**Innkeeper**: Marigold Sweetwater (Halfling, motherly)
**Doorman**: Tobias "Toby" Underhill (Halfling, friendly)
**Rest Cost**: 25g (cheapest)
**Atmosphere**: Cozy, rural, community gathering
**Accepted Instruments**:
- Flute: 10g base ✓ (favorite - folk songs)
- Pan Flute: 10g base ✓ (pastoral)
- Lute: 9g base ✓
- Drum: 8g base ✓
- Lyre: 8g base ✓
- Horn: 7g base ✓
- Viol: 7g base ✓
- Bagpipe: 6g base ✓ (village festivals)
- Shawm: 6g base ✓ (all welcome)

**Note**: Millhaven accepts all instruments but pays the least

---

### 6. Village-Southwest (Marshlight) - The Glowing Lantern (Swamp Tavern)
**Innkeeper**: Grishka Marshborn (Orc, practical)
**Doorman**: Rokgar the Grim (Orc, intimidating)
**Rest Cost**: 30g
**Atmosphere**: Rough, survivalist, tribal culture
**Accepted Instruments**:
- Drum: 16g base ✓ (favorite - tribal rhythms)
- Horn: 13g base ✓ (war calls)
- Shawm: 12g base ✓ (harsh = good)
- Bagpipe: 11g base ✓ (loud)
- Lute: 8g base ✓
- Flute: ✗ (too soft, shows weakness)
- Lyre: ✗ (too fancy)
- Pan Flute: ✗ (too gentle)
- Viol: ✗ (too refined)

---

## Performance Payment Formula

```
Total Pay = Base Pay + Charisma Bonus
Charisma Bonus = (CHA - 10) × 2 gold
```

**Examples**:
- CHA 10: No bonus
- CHA 14: +8 gold
- CHA 18: +16 gold
- CHA 20: +20 gold

**Example Calculation**:
- Location: Kingdom (Golden Griffin)
- Instrument: Lute (15g base)
- Player CHA: 16
- Charisma Bonus: (16-10) × 2 = 12g
- **Total Pay: 27 gold**

## Regional Preferences Summary

**Sophisticated** (Kingdom, Verdant):
- Prefer: Strings (lute, lyre, viol), soft wind (flute)
- Reject: Loud instruments (bagpipe, drum)

**Rowdy** (Goldenhaven, Ironpeak, Marshlight):
- Prefer: Loud instruments (drum, horn, bagpipe)
- Reject: Delicate instruments (lyre, flute)

**Rural** (Millhaven):
- Accept: Everything (community-focused)
- Pay Less: But everyone is welcome

## Rest System

**Cost by Location**:
- Goldenhaven (Port): 60g (most expensive)
- Kingdom (Capital): 50g
- Verdant (City): 45g
- Ironpeak (Town): 40g
- Marshlight (Swamp): 30g
- Millhaven (Village): 25g (cheapest)

**Benefits of Resting**:
- Fatigue restored to 0
- HP restored to maximum
- Temporary "Well Rested" buff (optional mechanic)

**Restrictions**:
- Can only rest once per day
- Must have enough gold
- Cannot rest if not fatigued

## NPC Roles

**Innkeeper** (Handles Lodging):
- Rents rooms
- Sells food/drink
- Provides rumors/information
- Friendly, hospitable

**Doorman/Bouncer** (Handles Entertainment):
- Books performances
- Explains rates and restrictions
- Enforces instrument policies
- Collects performance once per day

## Implementation Notes

### Database Tables

#### player_performances
```sql
CREATE TABLE player_performances (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  npub TEXT,
  tavern_id TEXT,
  instrument_id TEXT,
  base_pay INTEGER,
  charisma_bonus INTEGER,
  total_pay INTEGER,
  performed_at DATE,
  timestamp TIMESTAMP
);
```

#### player_rest_log
```sql
CREATE TABLE player_rest_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  npub TEXT,
  inn_id TEXT,
  cost INTEGER,
  fatigue_restored INTEGER,
  hp_restored INTEGER,
  rested_at DATE,
  timestamp TIMESTAMP
);
```

### API Endpoints

**POST /api/inn/:npc_id/rest**
```json
{
  "npub": "npub1..."
}
```

Response:
```json
{
  "success": true,
  "gold_spent": 50,
  "fatigue_restored": 85,
  "hp_restored": 15,
  "new_fatigue": 0,
  "new_hp": 45
}
```

**POST /api/tavern/:npc_id/perform**
```json
{
  "npub": "npub1...",
  "instrument_id": "lute"
}
```

Response:
```json
{
  "success": true,
  "instrument": "Lute",
  "base_pay": 15,
  "charisma_bonus": 12,
  "total_earned": 27,
  "can_perform_again": "2025-10-21"
}
```

### Save File Integration

```json
{
  "last_rest_date": "2025-10-20",
  "last_performances": {
    "tavern-doorman": "2025-10-20",
    "port-doorman": "2025-10-19"
  },
  "total_performances": 15,
  "total_performance_earnings": 385
}
```

## Balance Notes

- **High CHA characters** (bards) earn significantly more from performances
- **Expensive instruments** (lute, viol) generally pay better
- **Regional fit matters**: Bagpipes pay well in Ironpeak, nothing in Kingdom
- **Urban venues** pay more than rural
- **Performance income** supplements adventuring, not replaces it
- **Once per day limit** prevents farming gold through performances
- **Instrument ownership** required (can't perform without owning instrument)
