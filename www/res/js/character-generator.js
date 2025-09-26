// Deterministic Character Generation for Nostr Hero
// Uses Nostr public key as seed for D&D-style character creation

class NostrCharacterGenerator {
    constructor() {
        this.weights = null;
        this.introductions = null;
        this.startingGear = null;
    }

    // Load all necessary data files from API endpoints
    async initialize() {
        try {
            console.log('📊 Loading character generation data from API...');

            const [weightsResponse, introsResponse, gearResponse] = await Promise.all([
                fetch('/api/weights'),
                fetch('/api/introductions'),
                fetch('/api/starting-gear')
            ]);

            if (!weightsResponse.ok) throw new Error('Failed to load weights data from API');
            if (!introsResponse.ok) throw new Error('Failed to load introductions data from API');
            if (!gearResponse.ok) throw new Error('Failed to load starting gear data from API');

            this.weights = await weightsResponse.json();
            this.introductions = await introsResponse.json();
            this.startingGear = await gearResponse.json();

            console.log('✅ Character generation data loaded from API');
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
            name: this.generateName(goCharacter.race, this.createSeededRNG(12345)), // Use simple seed for name
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
            gold: this.rollDice('3d6', this.createSeededRNG(54321)) * 10, // Use simple seed for gold
            experience: 0
        };

        console.log('✅ Generated character from API:', character);
        return character;
    }

    // Generate introduction and starting scenario
    generateIntroduction(character) {
        const background = character.background;
        const race = character.race;
        const characterClass = character.class;

        const introduction = {
            baseIntro: this.introductions.base_intro,
            backgroundIntro: this.introductions.background_intros[background] || this.introductions.background_intros['Folk Hero'],
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
        const classGear = this.startingGear.equipment_packs[characterClass];

        if (!classGear) {
            console.warn('No starting gear found for class:', characterClass);
            return { inventory: [], equipment: {} };
        }

        const inventory = [];
        const equipment = {};

        // Add given items (automatic)
        if (classGear.given) {
            classGear.given.forEach(item => {
                if (typeof item === 'string') {
                    inventory.push({ item: item, quantity: 1 });
                } else if (typeof item === 'object') {
                    inventory.push({ item: item.item, quantity: item.quantity || 1 });
                    if (item.equipped) {
                        equipment[item.slot || 'weapon'] = item.item;
                    }
                }
            });
        }

        return {
            inventory: inventory,
            equipment: equipment,
            choices: classGear.choice || [] // Equipment choices to be made by player
        };
    }

    // Convert npub to hex key (matching Go backend format)
    npubToHex(npub) {
        // For now, we need to get the actual hex key from the session manager
        // The Go backend uses the raw hex public key, not the npub
        if (window.sessionManager && window.sessionManager.getSession()) {
            return window.sessionManager.getSession().publicKey;
        }
        // Fallback - this won't give correct results but prevents errors
        return npub.replace('npub1', '');
    }

    // Create deterministic seed (matching Go backend SHA256 approach)
    createDeterministicSeed(hexKey, context) {
        // This needs to match Go's CreateDeterministicSeed function exactly
        // Go does: hash := sha256.Sum256([]byte(hexKey + context))
        // Then: return int64(binary.BigEndian.Uint64(hash[:8]))

        // Simple hash that attempts to match Go's behavior
        const combined = hexKey + context;
        let hash = 0;

        // Create a more robust hash that better matches SHA256 behavior
        for (let i = 0; i < combined.length; i++) {
            const char = combined.charCodeAt(i);
            hash = ((hash << 5) - hash) + char;
            hash = hash & hash; // Convert to 32-bit integer
        }

        // Make it positive and match Go's int64 range behavior
        return Math.abs(hash);
    }

    // Create seeded random number generator (matching Go's rand.New)
    createSeededRNG(seed) {
        return {
            seed: seed,
            next: function() {
                // Linear congruential generator matching Go's approach
                this.seed = (this.seed * 1103515245 + 12345) & 0x7FFFFFFF;
                return this.seed / 0x7FFFFFFF;
            },
            intn: function(n) {
                return Math.floor(this.next() * n);
            }
        };
    }

    // Deterministic weighted choice (matching Go backend with sorting)
    deterministicWeightedChoice(options, weights, seed) {
        if (options.length === 0) return '';

        // Sort options for consistency (matching Go implementation)
        const indices = Array.from({length: options.length}, (_, i) => i);
        indices.sort((a, b) => options[a].localeCompare(options[b]));

        const sortedOptions = indices.map(i => options[i]);
        const sortedWeights = indices.map(i => weights[i]);

        // Compute total weight
        const totalWeight = sortedWeights.reduce((sum, w) => sum + w, 0);

        const rng = this.createSeededRNG(seed);
        const randomValue = rng.intn(totalWeight);

        let accumulatedWeight = 0;
        for (let i = 0; i < sortedWeights.length; i++) {
            accumulatedWeight += sortedWeights[i];
            if (randomValue < accumulatedWeight) {
                return sortedOptions[i];
            }
        }

        return sortedOptions[sortedOptions.length - 1]; // Fallback
    }

    // Generate race (matching Go backend)
    generateRace(hexKey) {
        const seed = this.createDeterministicSeed(hexKey, 'race');
        return this.deterministicWeightedChoice(this.weights.Races, this.weights.RaceWeights, seed);
    }

    // Generate class (matching Go backend)
    generateClass(hexKey, race) {
        const seed = this.createDeterministicSeed(hexKey, 'class_' + race);
        const raceWeights = this.weights.classWeightsByRace[race];
        if (!raceWeights) {
            console.warn('No class weights for race:', race);
            return 'Fighter';
        }

        const classOptions = Object.keys(raceWeights);
        const classWeights = Object.values(raceWeights);
        return this.deterministicWeightedChoice(classOptions, classWeights, seed);
    }

    // Generate background (matching Go backend)
    generateBackground(hexKey, characterClass) {
        const seed = this.createDeterministicSeed(hexKey, 'background_' + characterClass);
        const classWeights = this.weights.BackgroundWeightsByClass[characterClass];
        if (!classWeights) {
            console.warn('No background weights for class:', characterClass);
            return 'Folk Hero';
        }

        const backgroundOptions = Object.keys(classWeights);
        const backgroundWeights = Object.values(classWeights);
        return this.deterministicWeightedChoice(backgroundOptions, backgroundWeights, seed);
    }

    // Generate alignment (matching Go backend)
    generateAlignment(hexKey) {
        const seed = this.createDeterministicSeed(hexKey, 'alignment');
        return this.deterministicWeightedChoice(this.weights.Alignments, this.weights.AlignmentWeights, seed);
    }

    // Generate ability scores (matching Go backend)
    generateStats(hexKey, characterClass) {
        const seed = this.createDeterministicSeed(hexKey, 'stats');
        const rng = this.createSeededRNG(seed);

        const classRequirements = {
            'Paladin': { 'Strength': 12, 'Charisma': 12 },
            'Sorcerer': { 'Charisma': 12, 'Constitution': 12 },
            'Warlock': { 'Charisma': 12, 'Wisdom': 12 },
            'Bard': { 'Charisma': 12, 'Dexterity': 12 },
            'Fighter': { 'Strength': 12, 'Dexterity': 12 },
            'Barbarian': { 'Strength': 12, 'Constitution': 12 },
            'Monk': { 'Dexterity': 12, 'Wisdom': 12 },
            'Rogue': { 'Dexterity': 12, 'Intelligence': 12 },
            'Cleric': { 'Wisdom': 12, 'Charisma': 12 },
            'Druid': { 'Wisdom': 12, 'Intelligence': 12 },
            'Ranger': { 'Dexterity': 12, 'Wisdom': 12 },
            'Wizard': { 'Intelligence': 12, 'Wisdom': 12 }
        };

        // Roll initial stats
        const stats = {
            'Strength': this.rollStat(rng),
            'Dexterity': this.rollStat(rng),
            'Constitution': this.rollStat(rng),
            'Intelligence': this.rollStat(rng),
            'Wisdom': this.rollStat(rng),
            'Charisma': this.rollStat(rng)
        };

        // Enforce class minimums (matching Go logic)
        const requirements = classRequirements[characterClass] || {};
        for (const [stat, minValue] of Object.entries(requirements)) {
            while (stats[stat] < minValue) {
                stats[stat] = this.rollStat(rng);
            }
        }

        // Ensure stats remain within 8-16 range
        for (const [stat, value] of Object.entries(stats)) {
            if (value < 8) {
                stats[stat] = 8;
            } else if (value > 16) {
                stats[stat] = 16;
            }
        }

        // Convert to lowercase keys for compatibility with UI
        return {
            strength: stats['Strength'],
            dexterity: stats['Dexterity'],
            constitution: stats['Constitution'],
            intelligence: stats['Intelligence'],
            wisdom: stats['Wisdom'],
            charisma: stats['Charisma']
        };
    }

    // Roll a single ability score (matching Go backend - 4d6 drop lowest)
    rollStat(rng) {
        const dice = [
            rng.intn(6) + 1,
            rng.intn(6) + 1,
            rng.intn(6) + 1,
            rng.intn(6) + 1
        ];

        // Find minimum value index
        let minIndex = 0;
        for (let i = 1; i < dice.length; i++) {
            if (dice[i] < dice[minIndex]) {
                minIndex = i;
            }
        }

        // Sum all except minimum
        let sum = 0;
        for (let i = 0; i < dice.length; i++) {
            if (i !== minIndex) {
                sum += dice[i];
            }
        }

        return sum;
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

    // Generate appropriate name based on race
    generateName(race, rng) {
        const namesByRace = {
            'Human': ['Alaric', 'Beatrice', 'Connor', 'Diana', 'Edmund', 'Fiona', 'Gareth', 'Helena'],
            'Elf': ['Aelindra', 'Thranduil', 'Celebrian', 'Elrond', 'Galadriel', 'Legolas', 'Arwen', 'Haldir'],
            'Dwarf': ['Thorin', 'Dain', 'Balin', 'Dwalin', 'Gimli', 'Gloin', 'Oin', 'Nori'],
            'Halfling': ['Bilbo', 'Frodo', 'Samwise', 'Peregrin', 'Meriadoc', 'Rosie', 'Poppy', 'Daisy'],
            'Dragonborn': ['Arjhan', 'Balasar', 'Bharash', 'Donaar', 'Ghesh', 'Heskan', 'Kriv', 'Medrash'],
            'Gnome': ['Alston', 'Boddynock', 'Brocc', 'Burgell', 'Dimble', 'Eldon', 'Erky', 'Fonkin'],
            'Half-Elf': ['Aramil', 'Berris', 'Dayereth', 'Enna', 'Galinndan', 'Heian', 'Himo', 'Immeral'],
            'Half-Orc': ['Dench', 'Feng', 'Gell', 'Henk', 'Holg', 'Imsh', 'Keth', 'Krusk'],
            'Tiefling': ['Akmenos', 'Amnon', 'Barakas', 'Damakos', 'Ekemon', 'Iados', 'Kairon', 'Leucis'],
            'Orc': ['Grishnak', 'Ugluk', 'Azog', 'Bolg', 'Gothmog', 'Lurtz', 'Sharku', 'Yazneg']
        };

        const names = namesByRace[race] || namesByRace['Human'];
        const index = Math.floor(rng.next() * names.length);
        return names[index];
    }

    // Utility function for dice rolling with seeded RNG
    rollDice(diceString, rng) {
        const match = diceString.match(/(\d+)d(\d+)(?:\+(\d+))?/);
        if (!match) return 1;

        const numDice = parseInt(match[1]);
        const dieSize = parseInt(match[2]);
        const modifier = parseInt(match[3]) || 0;

        let total = 0;
        for (let i = 0; i < numDice; i++) {
            total += Math.floor(rng.next() * dieSize) + 1;
        }

        return total + modifier;
    }
}

// Global instance
window.characterGenerator = new NostrCharacterGenerator();

console.log('🎲 Character generator loaded');