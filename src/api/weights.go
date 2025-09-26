package api

import (
	"encoding/json"
	"net/http"
)

// WeightsHandler serves character generation weights from DuckDB
func WeightsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get weights data from DuckDB
	weightsData, err := getWeightsFromDB()
	if err != nil {
		http.Error(w, "Failed to load weights data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(weightsData)
}

// getWeightsFromDB retrieves all character generation weights from DuckDB
func getWeightsFromDB() (map[string]interface{}, error) {
	// For now, we'll construct the weights structure manually from the migrated data
	// In the future, this could be stored as a single JSON document in the database

	weights := map[string]interface{}{
		"Races": []string{
			"Human", "Elf", "Dwarf", "Halfling", "Gnome", "Orc",
			"Half-Elf", "Dragonborn", "Tiefling", "Half-Orc",
		},
		"RaceWeights": []int{36, 12, 14, 8, 6, 10, 7, 2, 3, 2},
		"classWeightsByRace": map[string]map[string]int{
			"Human": {
				"Fighter": 12, "Rogue": 8, "Bard": 8, "Cleric": 8,
				"Paladin": 8, "Ranger": 8, "Monk": 8, "Barbarian": 8,
				"Druid": 8, "Sorcerer": 8, "Warlock": 8, "Wizard": 8,
			},
			"Elf": {
				"Wizard": 16, "Ranger": 16, "Rogue": 10, "Bard": 8,
				"Sorcerer": 8, "Fighter": 8, "Druid": 8, "Cleric": 8,
				"Monk": 8, "Warlock": 8, "Paladin": 4, "Barbarian": 4,
			},
			"Dwarf": {
				"Fighter": 18, "Barbarian": 18, "Cleric": 14, "Paladin": 14,
				"Monk": 8, "Ranger": 8, "Druid": 8, "Rogue": 2,
				"Sorcerer": 2, "Bard": 2, "Warlock": 2, "Wizard": 2,
			},
			"Halfling": {
				"Rogue": 20, "Ranger": 14, "Fighter": 14, "Monk": 10,
				"Bard": 10, "Sorcerer": 10, "Warlock": 10, "Cleric": 10,
				"Druid": 10, "Barbarian": 4, "Paladin": 4, "Wizard": 4,
			},
			"Gnome": {
				"Wizard": 20, "Rogue": 20, "Sorcerer": 8, "Fighter": 8,
				"Warlock": 8, "Bard": 8, "Ranger": 6, "Monk": 6,
				"Paladin": 6, "Cleric": 4, "Druid": 4, "Barbarian": 2,
			},
			"Orc": {
				"Barbarian": 22, "Fighter": 16, "Rogue": 12, "Druid": 12,
				"Warlock": 6, "Ranger": 6, "Monk": 6, "Cleric": 6,
				"Paladin": 6, "Sorcerer": 4, "Bard": 4, "Wizard": 4,
			},
			"Half-Elf": {
				"Bard": 14, "Sorcerer": 14, "Monk": 12, "Paladin": 10,
				"Rogue": 10, "Fighter": 8, "Cleric": 8, "Druid": 8,
				"Ranger": 8, "Wizard": 8, "Warlock": 6, "Barbarian": 4,
			},
			"Dragonborn": {
				"Paladin": 22, "Sorcerer": 14, "Warlock": 14, "Bard": 14,
				"Fighter": 8, "Barbarian": 8, "Monk": 8, "Rogue": 2,
				"Cleric": 2, "Druid": 2, "Ranger": 2, "Wizard": 2,
			},
			"Tiefling": {
				"Sorcerer": 18, "Warlock": 18, "Wizard": 10, "Bard": 10,
				"Paladin": 8, "Fighter": 8, "Rogue": 8, "Monk": 8,
				"Ranger": 8, "Cleric": 2, "Druid": 2, "Barbarian": 2,
			},
			"Half-Orc": {
				"Barbarian": 22, "Fighter": 22, "Monk": 10, "Cleric": 10,
				"Paladin": 10, "Warlock": 10, "Bard": 6, "Ranger": 6,
				"Druid": 6, "Rogue": 2, "Sorcerer": 2, "Wizard": 2,
			},
		},
		"BackgroundWeightsByClass": map[string]map[string]int{
			"Paladin": {
				"Acolyte": 13, "Folk Hero": 13, "Gladiator": 12,
				"Guard": 12, "Knight": 12, "Noble": 12, "Soldier": 13, "Wayfarer": 13,
			},
			"Sorcerer": {
				"Charlatan": 16, "Entertainer": 12, "Hermit": 11,
				"Noble": 12, "Sage": 14, "Urchin": 11, "Wayfarer": 11, "Scribe": 13,
			},
			"Warlock": {
				"Charlatan": 17, "Criminal": 9, "Spy": 8, "Hermit": 17,
				"Outlander": 16, "Sage": 16, "Sailor": 7, "Urchin": 10,
			},
			"Bard": {
				"Charlatan": 13, "Entertainer": 13, "Artisan": 6, "Merchant": 7,
				"Noble": 13, "Sage": 12, "Sailor": 12, "Urchin": 12, "Wayfarer": 12,
			},
			"Fighter": {
				"Folk Hero": 15, "Gladiator": 15, "Knight": 16, "Outlander": 12,
				"Pirate": 7, "Sailor": 6, "Soldier": 15, "Farmer": 13,
			},
			"Barbarian": {
				"Folk Hero": 20, "Outlander": 23, "Pirate": 20,
				"Soldier": 21, "Urchin": 16,
			},
			"Monk": {
				"Acolyte": 16, "Hermit": 25, "Outlander": 13,
				"Sage": 18, "Wayfarer": 10, "Scribe": 18,
			},
			"Rogue": {
				"Charlatan": 17, "Criminal": 17, "Artisan": 8, "Merchant": 9,
				"Pirate": 16, "Sailor": 16, "Urchin": 17,
			},
			"Cleric": {
				"Acolyte": 21, "Folk Hero": 16, "Hermit": 21,
				"Sage": 13, "Wayfarer": 18, "Scribe": 13,
			},
			"Druid": {
				"Farmer": 20, "Hermit": 20, "Outlander": 20,
				"Sage": 20, "Wayfarer": 20,
			},
			"Ranger": {
				"Farmer": 17, "Folk Hero": 17, "Guide": 17,
				"Outlander": 17, "Sailor": 16, "Soldier": 16,
			},
			"Wizard": {
				"Acolyte": 16, "Artisan": 10, "Merchant": 10,
				"Hermit": 10, "Noble": 20, "Sage": 17, "Scribe": 17,
			},
		},
		"Alignments": []string{
			"Lawful Good", "Neutral Good", "Chaotic Good",
			"Lawful Neutral", "True Neutral", "Chaotic Neutral",
			"Lawful Evil", "Neutral Evil", "Chaotic Evil",
		},
		"AlignmentWeights": []int{10, 10, 10, 10, 20, 10, 10, 10, 10},
	}

	return weights, nil
}

// IntroductionsHandler serves character introduction data from DuckDB
func IntroductionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// For now, return a simple structure - this could be expanded to query from DB
	introductions := map[string]interface{}{
		"base_intro": map[string]string{
			"scene1":       "The rain falls relentlessly, drumming against the worn cobblestones of your village. Tonight marks eighteen years since your arrival in this world - a life spent without the guidance of parents, but not without care.",
			"scene2":       "For years, the old caretaker looked after you when no one else would. Teaching you, guiding you, preparing you for a world that can be both harsh and wondrous.",
			"scene3":       "Tonight, as the rain falls outside, they've passed peacefully in their sleep. Their final words still echo in your mind:",
			"final_words":  "Your destiny awaits beyond these village borders. You have a hero's heart - don't let it remain hidden here.",
			"letter_intro": "Before passing, they left you a letter. It speaks of dreams unfulfilled and adventures never taken - and a wish that you might find your own path.",
		},
		"background_intros": map[string]map[string]string{
			"Acolyte": {
				"scene":  "[Fade to: A small shrine in the corner of the room, candles flickering.] Their faith sustained them, even at the end. You were raised amidst prayers and rituals, guided by the old caretaker's unwavering devotion.",
				"letter": "My child, faith has been my anchor through life's storms. I found purpose in service to higher powers, in the rituals that mark our days, in the comfort of ancient texts. But I remained too long within temple walls when the wider world called.",
			},
			"Criminal": {
				"scene":  "[Fade to: A hidden compartment being opened in the floor, revealing concealed items.] Even in their final days, they kept secrets well. The old caretaker taught you to survive in the shadows, to observe without being seen.",
				"letter": "I never told you everything about my past. Some secrets are best carried to the grave. What I've taught you - the silent step, the watchful eye, the patient hand - these were born of necessity in darker times.",
			},
			"Folk Hero": {
				"scene":  "[Fade to: Maps and natural implements - seeds, dried herbs, survival tools.] They knew the wilds better than the village streets. The old caretaker taught you to read the stars, to track game, to live in harmony with nature's harsh beauty.",
				"letter": "The earth provides, the sky guides, and between them is where we make our way. The road called to me all my life, but I answered it too rarely. I found you, and that journey was enough to fill my heart, but there were countless paths I left unexplored.",
			},
			// Add more backgrounds as needed...
		},
		"equipment_intros": map[string]map[string]string{
			"warrior": {
				"scene": "Along the wall rest weapons and armor, maintained with meticulous care despite their age.",
				"quote": "Take what suits your strength. A warrior's tools should feel like extensions of the body, not burdens upon it.",
			},
			"faithful": {
				"scene": "A modest altar sits in the corner, testament to devotion maintained through decades.",
				"quote": "Faith and focus are weapons that never dull. When body and spirit align, even mountains can be moved.",
			},
			// Add more equipment types...
		},
		"final_note": map[string]string{
			"text":  "Among these practical items rests a simple pack, weathered but sturdy, large enough to carry what you'll need but not so large as to become a burden. Attached to it is a small note in your caretaker's hand:",
			"quote": "The weight you carry shapes the journey. Choose wisely.",
		},
		"departure": map[string]string{
			"text": "With the letter carefully folded and tucked away, you gather what you need for the journey ahead. The road stretches before you, damp from the night's rain but illuminated by the breaking dawn. Where fate will lead, only time will tell. But one thing is certain - you won't find your destiny by staying here.",
		},
	}

	json.NewEncoder(w).Encode(introductions)
}

// StartingGearHandler serves starting equipment data
func StartingGearHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// For now, return starting gear structure directly
	// This should eventually be loaded from DuckDB
	startingGear := map[string]interface{}{
		"equipment_packs": map[string]map[string]interface{}{
			"Fighter": map[string]interface{}{
				"given": []map[string]interface{}{
					{"item": "leather-armor", "quantity": 1, "equipped": true, "slot": "armor"},
					{"item": "shortsword", "quantity": 1, "equipped": true, "slot": "weapon"},
					{"item": "shield", "quantity": 1, "equipped": true, "slot": "shield"},
					{"item": "rations", "quantity": 5},
					{"item": "bedroll", "quantity": 1},
					{"item": "rope", "quantity": 1},
				},
				"choice": []map[string]interface{}{
					{
						"options": []string{"longsword", "battleaxe", "warhammer"},
						"description": "Choose your primary weapon",
					},
				},
			},
			"Wizard": map[string]interface{}{
				"given": []map[string]interface{}{
					{"item": "robes", "quantity": 1, "equipped": true, "slot": "armor"},
					{"item": "quarterstaff", "quantity": 1, "equipped": true, "slot": "weapon"},
					{"item": "spellbook", "quantity": 1},
					{"item": "spell-component-pouch", "quantity": 1},
					{"item": "rations", "quantity": 3},
					{"item": "bedroll", "quantity": 1},
				},
				"choice": []map[string]interface{}{
					{
						"options": []string{"healing-potion", "mage-armor-scroll", "magic-missile-scroll"},
						"description": "Choose your starting magical aid",
					},
				},
			},
			"Rogue": map[string]interface{}{
				"given": []map[string]interface{}{
					{"item": "leather-armor", "quantity": 1, "equipped": true, "slot": "armor"},
					{"item": "shortsword", "quantity": 1, "equipped": true, "slot": "weapon"},
					{"item": "dagger", "quantity": 2},
					{"item": "thieves-tools", "quantity": 1},
					{"item": "rations", "quantity": 4},
					{"item": "rope", "quantity": 1},
				},
				"choice": []map[string]interface{}{
					{
						"options": []string{"shortbow", "crossbow", "throwing-knives"},
						"description": "Choose your ranged weapon",
					},
				},
			},
			"Cleric": map[string]interface{}{
				"given": []map[string]interface{}{
					{"item": "chainmail", "quantity": 1, "equipped": true, "slot": "armor"},
					{"item": "mace", "quantity": 1, "equipped": true, "slot": "weapon"},
					{"item": "shield", "quantity": 1, "equipped": true, "slot": "shield"},
					{"item": "holy-symbol", "quantity": 1},
					{"item": "rations", "quantity": 4},
					{"item": "bedroll", "quantity": 1},
				},
				"choice": []map[string]interface{}{
					{
						"options": []string{"healing-potion", "blessing-scroll", "turn-undead-scroll"},
						"description": "Choose your divine blessing",
					},
				},
			},
		},
	}

	json.NewEncoder(w).Encode(startingGear)
}