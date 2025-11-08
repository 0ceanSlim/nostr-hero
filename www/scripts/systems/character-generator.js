// Deterministic Character Generation for Nostr Hero
// Uses Nostr public key as seed for D&D-style character creation

class NostrCharacterGenerator {
    constructor() {
        this.weights = null;
        this.introductions = null;
        this.startingGear = null;
        this.startingGold = null;
        this.startingSpells = null;
        this.spellProgression = null;
        this.spellSlots = null;
        this.racialStartingCities = null;
    }

    // Load all necessary data files from API endpoints
    async initialize() {
        try {
            console.log('📊 Loading character generation data...');

            const [
                weightsResponse,
                introsResponse,
                gearResponse,
                goldResponse,
                spellsResponse,
                progressionResponse,
                slotsResponse,
                citiesResponse
            ] = await Promise.all([
                fetch('/data/systems/weights.json'),
                fetch('/data/character/introductions.json'),
                fetch('/data/character/starting-gear.json'),
                fetch('/data/character/starting-gold.json'),
                fetch('/data/character/starting-spells.json'),
                fetch('/data/character/spell-progression.json'),
                fetch('/data/character/spell-slots.json'),
                fetch('/data/character/racial-starting-cities.json')
            ]);

            if (!weightsResponse.ok) throw new Error('Failed to load weights data');
            if (!introsResponse.ok) throw new Error('Failed to load introductions data');
            if (!gearResponse.ok) throw new Error('Failed to load starting gear data');
            if (!goldResponse.ok) throw new Error('Failed to load starting gold data');
            if (!spellsResponse.ok) throw new Error('Failed to load starting spells data');
            if (!progressionResponse.ok) throw new Error('Failed to load spell progression data');
            if (!slotsResponse.ok) throw new Error('Failed to load spell slots data');
            if (!citiesResponse.ok) throw new Error('Failed to load racial starting cities data');

            this.weights = await weightsResponse.json();
            this.introductions = await introsResponse.json();
            this.startingGear = await gearResponse.json();
            this.startingGold = await goldResponse.json();
            this.startingSpells = await spellsResponse.json();
            this.spellProgression = await progressionResponse.json();
            this.spellSlots = await slotsResponse.json();
            const citiesData = await citiesResponse.json();
            this.racialStartingCities = citiesData.racial_starting_cities;

            console.log('✅ Character generation data loaded');
            return true;
        } catch (error) {
            console.error('❌ Failed to load character generation data:', error);
            throw error;
        }
    }

    // Generate a complete character from a Nostr public key using existing API
    async generateCharacter(npub) {
        if (!this.introductions || !this.startingGear) {
            throw new Error('Character generator not initialized');
        }

        console.log('🎲 Generating character for:', npub);

        // Use the existing /api/character endpoint that has the correct algorithm
        const response = await fetch(`/api/character?npub=${encodeURIComponent(npub)}`);
        if (!response.ok) {
            throw new Error(`Failed to generate character: ${response.status} ${response.statusText}`);
        }

        const apiResult = await response.json();
        const goCharacter = apiResult.character;

        // Convert Go character format to our expected format
        const character = {
            name: '', // Player enters their own name (d field)
            race: goCharacter.race,
            class: goCharacter.class,
            background: goCharacter.background,
            alignment: goCharacter.alignment,
            stats: {
                strength: goCharacter.stats.Strength,
                dexterity: goCharacter.stats.Dexterity,
                constitution: goCharacter.stats.Constitution,
                intelligence: goCharacter.stats.Intelligence,
                wisdom: goCharacter.stats.Wisdom,
                charisma: goCharacter.stats.Charisma
            },
            level: 1,
            hp: this.calculateHP(goCharacter.stats.Constitution, goCharacter.class),
            max_hp: this.calculateHP(goCharacter.stats.Constitution, goCharacter.class),
            mana: this.calculateMana({
                strength: goCharacter.stats.Strength,
                dexterity: goCharacter.stats.Dexterity,
                constitution: goCharacter.stats.Constitution,
                intelligence: goCharacter.stats.Intelligence,
                wisdom: goCharacter.stats.Wisdom,
                charisma: goCharacter.stats.Charisma
            }, goCharacter.class),
            max_mana: this.calculateMana({
                strength: goCharacter.stats.Strength,
                dexterity: goCharacter.stats.Dexterity,
                constitution: goCharacter.stats.Constitution,
                intelligence: goCharacter.stats.Intelligence,
                wisdom: goCharacter.stats.Wisdom,
                charisma: goCharacter.stats.Charisma
            }, goCharacter.class),
            fatigue: 0,
            gold: this.generateStartingGold(goCharacter.background),
            experience: 0
        };

        // Add equipment, spells, and city
        const equipmentResult = this.generateStartingEquipment(character);
        character.inventory = equipmentResult.inventory;
        character.equipment = equipmentResult.equipment;
        character.choices = equipmentResult.choices;

        character.spells = this.generateStartingSpells(character);
        character.spell_slots = this.generateSpellSlots(character);
        character.city = this.generateStartingCity(character);

        console.log('✅ Generated character from API:', character);
        return character;
    }

    // Generate introduction and starting scenario
    generateIntroduction(character) {
        const background = character.background;
        const race = character.race;
        const characterClass = character.class;

        const backgroundIntro = this.introductions.background_intros.find(entry =>
            entry.backgrounds.includes(background)
        ) || this.introductions.background_intros.find(entry =>
            entry.backgrounds.includes('Folk Hero')
        );

        const introduction = {
            baseIntro: this.introductions.base_intro,
            backgroundIntro: backgroundIntro,
            equipmentIntro: this.getEquipmentIntro(characterClass),
            finalNote: this.introductions.final_note,
            departure: this.introductions.departure
        };

        return introduction;
    }

    // Get appropriate equipment introduction based on class
    getEquipmentIntro(characterClass) {
        const classToType = {
            'Fighter': 'warrior',
            'Barbarian': 'warrior',
            'Paladin': 'faithful',
            'Cleric': 'faithful',
            'Monk': 'faithful',
            'Ranger': 'wilderness',
            'Druid': 'wilderness',
            'Wizard': 'arcane',
            'Sorcerer': 'arcane',
            'Warlock': 'arcane',
            'Rogue': 'clever',
            'Bard': 'clever'
        };

        const equipType = classToType[characterClass] || 'warrior';
        return this.introductions.equipment_intros[equipType];
    }

    // Generate starting equipment based on class
    generateStartingEquipment(character) {
        const characterClass = character.class;
        const classGearData = this.startingGear.find(g => g.class === characterClass);

        if (!classGearData) {
            console.warn('No starting gear found for class:', characterClass);
            return { inventory: [], equipment: {}, choices: [] };
        }

        const inventory = [];
        const equipment = {};
        const choices = [];
        const gearConfig = classGearData.starting_gear;

        // Process given items (auto-added to inventory)
        if (gearConfig.given_items) {
            gearConfig.given_items.forEach(givenItem => {
                inventory.push({ item: givenItem.item, quantity: givenItem.quantity });
            });
        }

        // Process equipment choices
        if (gearConfig.equipment_choices) {
            console.log(`🎒 Processing ${gearConfig.equipment_choices.length} equipment choices for ${characterClass}`);
            gearConfig.equipment_choices.forEach((equipChoice, index) => {
                console.log(`  Choice ${index}: ${equipChoice.options.length} options`);
                const choice = {
                    id: `choice-${index}`,
                    description: equipChoice.description || '',
                    options: equipChoice.options.map((opt, optIndex) => {
                        if (opt.type === 'single') {
                            // Single item choice
                            return {
                                item: opt.item,
                                quantity: opt.quantity,
                                type: 'single'
                            };
                        } else if (opt.type === 'bundle') {
                            // Bundle of items
                            const items = opt.items.map(i => `${i.item} (x${i.quantity})`).join(' + ');
                            return {
                                item: items,
                                quantity: 1,
                                isBundle: true,
                                bundle: opt.items.map(i => [i.item, i.quantity]),
                                type: 'bundle'
                            };
                        } else if (opt.type === 'multi_slot') {
                            // Complex multi-slot choice (weapon+shield OR 2 weapons)
                            const slotDescriptions = opt.slots.map(slot => {
                                if (slot.type === 'weapon_choice') {
                                    return 'Choose weapon';
                                } else if (slot.type === 'fixed') {
                                    return `${slot.item} (x${slot.quantity})`;
                                }
                                return 'Unknown slot';
                            });

                            return {
                                item: slotDescriptions.join(' + '),
                                quantity: 1,
                                isComplexChoice: true,
                                type: 'multi_slot',
                                weaponSlots: opt.slots.map((slot, slotIndex) => {
                                    if (slot.type === 'weapon_choice') {
                                        // Convert weapon options to old format for compatibility
                                        const weaponOptions = slot.options.map(w => [w, 1]);
                                        return {
                                            type: 'weapon_choice',
                                            options: weaponOptions,
                                            index: slotIndex
                                        };
                                    } else if (slot.type === 'fixed') {
                                        return {
                                            type: 'fixed_item',
                                            item: [slot.item, slot.quantity],
                                            index: slotIndex
                                        };
                                    }
                                    return { type: 'unknown', index: slotIndex };
                                })
                            };
                        }

                        console.warn('Unknown option type:', opt);
                        return { item: 'Unknown', quantity: 1 };
                    })
                };
                choices.push(choice);
            });
        }

        // Store pack choice separately (not in choices array)
        let packChoice = null;
        if (gearConfig.pack_choice) {
            packChoice = {
                id: 'pack-choice',
                description: gearConfig.pack_choice.description || 'Choose your adventuring pack',
                options: gearConfig.pack_choice.options.map(packId => ({
                    item: packId,
                    quantity: 1,
                    type: 'single'
                }))
            };
        }

        console.log(`✅ Generated ${choices.length} equipment choices for ${characterClass}:`, choices);
        return {
            inventory: inventory,
            equipment: equipment, // TODO: Handle default equipped items
            choices: choices,
            pack_choice: packChoice
        };
    }

    // Generate starting gold based on background
    generateStartingGold(background) {
        if (!this.startingGold || !this.startingGold['starting-gold']) {
            console.warn('Starting gold data not loaded');
            return 0;
        }

        const goldData = this.startingGold['starting-gold'];
        const entry = goldData.find(item => item[0] === background);

        if (entry) {
            return entry[1];
        } else {
            console.warn('No starting gold found for background:', background);
            return 1000; // Default gold
        }
    }

    // Generate starting city based on race
    generateStartingCity(character) {
        if (!this.racialStartingCities) {
            console.warn('Racial starting cities data not loaded');
            return 'kingdom'; // Default city
        }

        const startingCity = this.racialStartingCities[character.race];
        if (startingCity) {
            return startingCity;
        } else {
            console.warn('No starting city found for race:', character.race);
            return 'kingdom'; // Default to kingdom instead of 'Nexus'
        }
    }

    // Generate starting spells based on class
    generateStartingSpells(character) {
        if (!this.startingSpells) {
            console.warn('Spell data not loaded');
            return [];
        }

        const classSpells = this.startingSpells[character.class.toLowerCase()];
        if (!classSpells) {
            console.log(`${character.class} has no starting spells`);
            return [];
        }

        const knownSpells = [];

        // Add ALL cantrips from starting spells (they know all of them)
        if (classSpells.cantrips && classSpells.cantrips.length > 0) {
            knownSpells.push(...classSpells.cantrips);
        }

        // Add ALL level 1 spells from starting spells (they know all of them)
        if (classSpells.level1 && classSpells.level1.length > 0) {
            knownSpells.push(...classSpells.level1);
        }

        console.log(`Generated ${knownSpells.length} known spells for ${character.class}:`, knownSpells);
        return knownSpells;
    }

    // Generate spell slots for character
    generateSpellSlots(character) {
        if (!this.spellSlots) {
            console.warn('Spell slots data not loaded');
            return {};
        }

        const classSlots = this.spellSlots[character.class.toLowerCase()];
        if (!classSlots) {
            console.log(`${character.class} is not a spellcasting class or has no spell slots`);
            return {};
        }

        // Get slot counts for level 1
        const level1SlotCounts = classSlots["1"] || {};
        console.log(`Spell slot counts for ${character.class} level 1:`, level1SlotCounts);

        // Convert slot counts to actual slot objects
        const spellSlots = {};

        for (const [slotType, count] of Object.entries(level1SlotCounts)) {
            if (count > 0) {
                spellSlots[slotType] = [];
                for (let i = 0; i < count; i++) {
                    spellSlots[slotType].push({
                        slot: i,
                        spell: null,      // No spell prepared initially
                        quantity: 0       // Can prepare up to 5 of same spell
                    });
                }
            }
        }

        console.log(`Generated spell slots for ${character.class}:`, spellSlots);
        return spellSlots;
    }

    // Calculate HP based on constitution and class
    calculateHP(constitution, characterClass) {
        const hitDice = {
            'Barbarian': 12,
            'Fighter': 10,
            'Paladin': 10,
            'Monk': 8,
            'Ranger': 10,
            'Rogue': 8,
            'Bard': 8,
            'Cleric': 8,
            'Druid': 8,
            'Sorcerer': 6,
            'Warlock': 8,
            'Wizard': 6
        };

        const hitDie = hitDice[characterClass] || 8;
        const conModifier = Math.floor((constitution - 10) / 2);

        return hitDie + conModifier;
    }

    // Calculate mana based on stats and class
    calculateMana(stats, characterClass) {
        const spellcasters = {
            'Wizard': 'intelligence',
            'Sorcerer': 'charisma',
            'Warlock': 'charisma',
            'Bard': 'charisma',
            'Cleric': 'wisdom',
            'Druid': 'wisdom',
            'Paladin': 'charisma',
            'Ranger': 'wisdom'
        };

        const spellcastingStat = spellcasters[characterClass];
        if (!spellcastingStat) return 0; // Non-casters have no mana

        const statValue = stats[spellcastingStat];
        const statModifier = Math.floor((statValue - 10) / 2);

        return Math.max(0, statModifier + 1); // Base mana from casting stat
    }
}

// Global instance
window.characterGenerator = new NostrCharacterGenerator();

console.log('🎲 Character generator loaded');
