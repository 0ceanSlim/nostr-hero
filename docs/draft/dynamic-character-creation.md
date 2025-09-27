# Dynamic Character Creation Flow

## Overview
1. **Derive character** from pubkey (race, class, stats) automatically
2. **Present sequential choice windows** for class-specific starting gear options
3. **Smooth transitions** between choices with HTMX
4. **Final summary** of all gear + auto-given items
5. **Generate save file** and redirect to starting city

## Step 1: Character Derivation from Pubkey

```python
def derive_character_from_pubkey(pubkey):
    """Generate deterministic character from pubkey"""

    # Use pubkey as seed for consistent results
    import hashlib
    seed = int(hashlib.sha256(pubkey.encode()).hexdigest(), 16)
    random.seed(seed)

    # Load game data
    races = load_races()
    classes = load_classes()

    # Derive race and class
    race = random.choice(races)
    char_class = random.choice(classes)

    # Generate stats (4d6 drop lowest, 6 times)
    stats = {}
    for stat in ['strength', 'dexterity', 'constitution', 'intelligence', 'wisdom', 'charisma']:
        rolls = [random.randint(1, 6) for _ in range(4)]
        stats[stat] = sum(sorted(rolls, reverse=True)[:3])

    # Apply racial bonuses
    stats = apply_racial_bonuses(stats, race)

    return {
        'pubkey': pubkey,
        'race': race,
        'class': char_class,
        'stats': stats,
        'level': 1,
        'hp': calculate_hp(char_class, stats['constitution']),
        'mana': calculate_mana(char_class) if is_spellcaster(char_class) else 0
    }
```

## Step 2: Sequential Choice System

```python
def get_class_choice_sequence(class_name):
    """Get ordered list of choices for a class"""

    with open('www/data/starting-gear.json', 'r') as f:
        starting_gear = json.load(f)

    class_data = next(item for item in starting_gear if item['class'].lower() == class_name.lower())

    choices = []
    for i, gear_group in enumerate(class_data['starting_gear']):
        if 'option' in gear_group:
            choices.append({
                'id': i,
                'options': gear_group['option'],
                'title': derive_choice_title(gear_group['option']),
                'type': derive_choice_type(gear_group['option'])
            })

    return choices

def derive_choice_title(options):
    """Generate title based on option content"""
    first_option = options[0]

    if 'sword' in str(first_option).lower() or 'axe' in str(first_option).lower():
        return "Choose Your Starting Weapon"
    elif 'armor' in str(first_option).lower() or 'mail' in str(first_option).lower():
        return "Choose Your Armor"
    elif 'pack' in str(first_option).lower():
        return "Choose Your Equipment Pack"
    elif 'wand' in str(first_option).lower() or 'orb' in str(first_option).lower():
        return "Choose Your Spellcasting Focus"
    elif 'component' in str(first_option).lower() or 'water' in str(first_option).lower():
        return "Choose Your Spell Components"
    else:
        return "Choose Your Equipment"
```

## Step 3: HTMX Flow Routes

```python
@app.get("/start")
async def start_character_creation(pubkey: str):
    """Initial entry point - derive character and start choices"""

    # Derive character from pubkey
    character = derive_character_from_pubkey(pubkey)

    # Get choice sequence for their class
    choices = get_class_choice_sequence(character['class'])

    # Store in session
    session_id = generate_session_id()
    store_character_session(session_id, {
        'character': character,
        'choices': choices,
        'selections': {},
        'current_choice': 0
    })

    # Return first choice window
    return render_choice_window(session_id, 0)

@app.post("/make-choice/{session_id}/{choice_id}")
async def make_choice(session_id: str, choice_id: int, selected_option: str):
    """Process a choice and move to next window"""

    session_data = get_character_session(session_id)

    # Store the selection
    session_data['selections'][choice_id] = selected_option
    session_data['current_choice'] = choice_id + 1

    update_character_session(session_id, session_data)

    # Check if more choices remain
    if session_data['current_choice'] < len(session_data['choices']):
        # Return next choice window
        return render_choice_window(session_id, session_data['current_choice'])
    else:
        # Return final summary
        return render_final_summary(session_id)

@app.post("/finalize-character/{session_id}")
async def finalize_character(session_id: str):
    """Generate save file and redirect to game"""

    session_data = get_character_session(session_id)

    # Generate final inventory
    final_character = generate_final_character(session_data)

    # Save to database
    save_character(final_character)

    # Clear session
    clear_character_session(session_id)

    # Redirect to starting city
    starting_location = get_starting_location(final_character['race'])
    return redirect(f"/game?pubkey={final_character['pubkey']}&location={starting_location}")
```

## Step 4: Choice Window HTML Template

```html
<!-- Single choice window that gets replaced -->
<div id="choice-window" class="choice-container">
    <div class="character-info">
        <h2>{{ character.race }} {{ character.class }}</h2>
        <p>Choice {{ current_choice + 1 }} of {{ total_choices }}</p>
    </div>

    <div class="choice-title">
        <h3>{{ choice.title }}</h3>
    </div>

    <div class="options-grid">
        {% for option in choice.options %}
        <div
            class="option-card"
            hx-post="/make-choice/{{ session_id }}/{{ choice.id }}"
            hx-vals='{"selected_option": "{{ option[0] if option is list else option }}"}'
            hx-target="#choice-window"
            hx-swap="outerHTML"
        >
            <div class="option-name">{{ get_item_name(option) }}</div>
            <div class="option-description">{{ get_item_description(option) }}</div>
            {% if get_item_image(option) %}
            <img src="{{ get_item_image(option) }}" alt="{{ get_item_name(option) }}" class="option-image">
            {% endif %}
        </div>
        {% endfor %}
    </div>
</div>
```

## Step 5: Final Summary Window

```html
<div id="character-complete" class="final-summary">
    <div class="character-header">
        <h2>{{ character.race }} {{ character.class }} - Ready!</h2>
        <div class="character-stats">
            <div>HP: {{ character.hp }}</div>
            {% if character.mana > 0 %}
            <div>Mana: {{ character.mana }}</div>
            {% endif %}
        </div>
    </div>

    <div class="selected-gear">
        <h3>Your Selected Equipment</h3>
        <div class="gear-list">
            {% for selection in selections %}
            <div class="gear-item selected">
                <strong>{{ get_item_name(selection) }}</strong>
                <p>{{ get_item_description(selection) }}</p>
            </div>
            {% endfor %}
        </div>
    </div>

    <div class="guaranteed-gear">
        <h3>Additional Starting Equipment</h3>
        <div class="gear-list">
            {% for item in guaranteed_items %}
            <div class="gear-item guaranteed">
                <strong>{{ item.name }}</strong> ({{ item.quantity }}x)
                <p>{{ item.description }}</p>
            </div>
            {% endfor %}
        </div>
    </div>

    {% if character.spells %}
    <div class="starting-spells">
        <h3>Your Starting Spells</h3>
        <div class="spell-list">
            {% for spell in character.spells %}
            <div class="spell-item">
                <strong>{{ spell.name }}</strong> - {{ spell.description }}
            </div>
            {% endfor %}
        </div>
    </div>
    {% endif %}

    <button
        class="start-game-btn"
        hx-post="/finalize-character/{{ session_id }}"
        hx-target="body"
    >
        Begin Your Adventure in {{ starting_city }}
    </button>
</div>
```

## Step 6: Save File Generation

```python
def generate_final_character(session_data):
    """Generate complete character with all gear and selections"""

    character = session_data['character']
    selections = session_data['selections']

    # Build final inventory from selections + guaranteed items
    final_inventory = []

    # Add selected items
    for choice_id, selected_option in selections.items():
        item_data = get_item_data(selected_option)
        final_inventory.append({
            'item_id': selected_option,
            'quantity': 1,
            'source': 'selected'
        })

    # Add guaranteed items for this class
    guaranteed_items = get_guaranteed_items(character['class'])
    for item in guaranteed_items:
        final_inventory.append({
            'item_id': item['item'],
            'quantity': item['quantity'],
            'source': 'guaranteed'
        })

    # Add starting spells if spellcaster
    if is_spellcaster(character['class']):
        character['spells'] = get_starting_spells(character['class'])

    # Add final inventory
    character['inventory'] = final_inventory
    character['created_at'] = datetime.utcnow().isoformat()

    return character
```

## Example Flow for Wizard:

1. **Derive:** "You are a Human Wizard with high Intelligence"
2. **Choice 1:** "Choose Your Starting Weapon" → Quarterstaff vs Dagger
3. **Choice 2:** "Choose Your Spellcasting Focus" → Wand/Orb/Rod/Staff/Crystal
4. **Choice 3:** "Choose Your Spell Components" → Component Pouch vs Bat Guano & Sulfur
5. **Choice 4:** "Choose Your Equipment Pack" → Scholar's Pack vs Explorer's Pack
6. **Summary:** Shows selections + guaranteed items (Spellbook, 1x Bat Guano & Sulfur, starting spells)
7. **Start Game:** Creates save file, redirects to starting city

Each choice is a clean window with big clickable cards, smooth HTMX transitions, and a final "here's everything you got" summary before entering the game world.