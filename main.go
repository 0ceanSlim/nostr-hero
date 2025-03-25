package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
)

type RaceClassWeight struct {
	Race   string
	Class  string
	Weight int
}

type RaceBackgroundWeight struct {
	Race       string
	Background string
	Weight     int
}

var races = []string{"Human", "Elf", "Dwarf", "Halfling", "Orc", "Tiefling", "Gnome", "Dragonborn"}
var raceWeights = []int{30, 15, 15, 10, 10, 10, 5, 5}

// var classes = []string{"Fighter", "Wizard", "Rogue", "Cleric", "Paladin", "Ranger", "Bard", "Warlock"}

var classWeightsByRace = []RaceClassWeight{
	{"Human", "Fighter", 20}, {"Human", "Wizard", 15}, {"Human", "Rogue", 15}, {"Human", "Cleric", 12}, {"Human", "Paladin", 10}, {"Human", "Ranger", 10}, {"Human", "Bard", 10}, {"Human", "Warlock", 8},
	{"Elf", "Ranger", 25}, {"Elf", "Wizard", 20}, {"Elf", "Rogue", 15}, {"Elf", "Bard", 12}, {"Elf", "Fighter", 10}, {"Elf", "Cleric", 8}, {"Elf", "Warlock", 5}, {"Elf", "Paladin", 5},
	{"Dwarf", "Fighter", 25}, {"Dwarf", "Cleric", 20}, {"Dwarf", "Paladin", 15}, {"Dwarf", "Ranger", 12}, {"Dwarf", "Rogue", 10}, {"Dwarf", "Wizard", 8}, {"Dwarf", "Warlock", 5}, {"Dwarf", "Bard", 5},
	{"Halfling", "Rogue", 30}, {"Halfling", "Bard", 20}, {"Halfling", "Ranger", 15}, {"Halfling", "Fighter", 10}, {"Halfling", "Wizard", 8}, {"Halfling", "Cleric", 7}, {"Halfling", "Warlock", 5}, {"Halfling", "Paladin", 5},
	{"Orc", "Fighter", 30}, {"Orc", "Barbarian", 25}, {"Orc", "Ranger", 15}, {"Orc", "Cleric", 10}, {"Orc", "Warlock", 8}, {"Orc", "Rogue", 5}, {"Orc", "Wizard", 4}, {"Orc", "Bard", 3},
	{"Tiefling", "Warlock", 30}, {"Tiefling", "Sorcerer", 25}, {"Tiefling", "Rogue", 15}, {"Tiefling", "Bard", 10}, {"Tiefling", "Fighter", 8}, {"Tiefling", "Wizard", 5}, {"Tiefling", "Cleric", 4}, {"Tiefling", "Paladin", 3},
	{"Gnome", "Wizard", 25}, {"Gnome", "Bard", 20}, {"Gnome", "Rogue", 15}, {"Gnome", "Warlock", 12}, {"Gnome", "Fighter", 10}, {"Gnome", "Cleric", 8}, {"Gnome", "Ranger", 5}, {"Gnome", "Paladin", 5},
	{"Dragonborn", "Paladin", 30}, {"Dragonborn", "Fighter", 20}, {"Dragonborn", "Sorcerer", 15}, {"Dragonborn", "Cleric", 12}, {"Dragonborn", "Warlock", 8}, {"Dragonborn", "Bard", 6}, {"Dragonborn", "Ranger", 5}, {"Dragonborn", "Rogue", 4},
}

// var backgrounds = []string{"Soldier", "Sage", "Outlander", "Noble", "Acolyte", "Charlatan", "Entertainer", "Hermit"}

var backgroundWeightsByRace = []RaceBackgroundWeight{
	{"Human", "Soldier", 20}, {"Human", "Noble", 20}, {"Human", "Charlatan", 15}, {"Human", "Sage", 15}, {"Human", "Outlander", 10}, {"Human", "Acolyte", 10}, {"Human", "Entertainer", 5}, {"Human", "Hermit", 5},
	{"Elf", "Sage", 25}, {"Elf", "Noble", 20}, {"Elf", "Outlander", 15}, {"Elf", "Hermit", 10}, {"Elf", "Acolyte", 10}, {"Elf", "Entertainer", 8}, {"Elf", "Soldier", 7}, {"Elf", "Charlatan", 5},
	{"Dwarf", "Soldier", 25}, {"Dwarf", "Acolyte", 20}, {"Dwarf", "Outlander", 15}, {"Dwarf", "Noble", 12}, {"Dwarf", "Sage", 10}, {"Dwarf", "Hermit", 8}, {"Dwarf", "Entertainer", 5}, {"Dwarf", "Charlatan", 5},
	{"Halfling", "Charlatan", 30}, {"Halfling", "Entertainer", 25}, {"Halfling", "Hermit", 15}, {"Halfling", "Outlander", 10}, {"Halfling", "Acolyte", 8}, {"Halfling", "Sage", 5}, {"Halfling", "Soldier", 4}, {"Halfling", "Noble", 3},
	{"Orc", "Soldier", 30}, {"Orc", "Outlander", 25}, {"Orc", "Hermit", 15}, {"Orc", "Charlatan", 10}, {"Orc", "Entertainer", 8}, {"Orc", "Acolyte", 5}, {"Orc", "Noble", 4}, {"Orc", "Sage", 3},
	{"Tiefling", "Charlatan", 25}, {"Tiefling", "Noble", 20}, {"Tiefling", "Hermit", 15}, {"Tiefling", "Entertainer", 12}, {"Tiefling", "Sage", 10}, {"Tiefling", "Outlander", 8}, {"Tiefling", "Soldier", 5}, {"Tiefling", "Acolyte", 5},
	{"Gnome", "Sage", 25}, {"Gnome", "Entertainer", 20}, {"Gnome", "Charlatan", 15}, {"Gnome", "Hermit", 12}, {"Gnome", "Acolyte", 10}, {"Gnome", "Noble", 8}, {"Gnome", "Outlander", 5}, {"Gnome", "Soldier", 5},
	{"Dragonborn", "Soldier", 30}, {"Dragonborn", "Noble", 25}, {"Dragonborn", "Acolyte", 15}, {"Dragonborn", "Sage", 10}, {"Dragonborn", "Outlander", 8}, {"Dragonborn", "Entertainer", 5}, {"Dragonborn", "Charlatan", 4}, {"Dragonborn", "Hermit", 3},
}

var alignments = []string{"Lawful Good", "Neutral Good", "Chaotic Good", "Lawful Neutral", "True Neutral", "Chaotic Neutral", "Lawful Evil", "Neutral Evil", "Chaotic Evil"}
var alignmentWeights = []int{10, 10, 10, 10, 20, 10, 10, 10, 10}

func rollStat(rng *rand.Rand) int {
	dice := []int{rng.Intn(6) + 1, rng.Intn(6) + 1, rng.Intn(6) + 1, rng.Intn(6) + 1}
	minIndex := 0
	for i := 1; i < 4; i++ {
		if dice[i] < dice[minIndex] {
			minIndex = i
		}
	}
	return dice[0] + dice[1] + dice[2] + dice[3] - dice[minIndex]
}

func weightedChoice(options []string, weights []int, rng *rand.Rand) string {
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	randVal := rng.Intn(totalWeight)
	for i, w := range weights {
		randVal -= w
		if randVal < 0 {
			return options[i]
		}
	}
	return options[len(options)-1]
}

func chooseClassByRace(race string, rng *rand.Rand) string {
	filteredClasses := []string{}
	filteredWeights := []int{}
	for _, rc := range classWeightsByRace {
		if rc.Race == race {
			filteredClasses = append(filteredClasses, rc.Class)
			filteredWeights = append(filteredWeights, rc.Weight)
		}
	}
	return weightedChoice(filteredClasses, filteredWeights, rng)
}

func chooseBackgroundByRace(race string, rng *rand.Rand) string {
	filteredBackgrounds := []string{}
	filteredWeights := []int{}
	for _, rb := range backgroundWeightsByRace {
		if rb.Race == race {
			filteredBackgrounds = append(filteredBackgrounds, rb.Background)
			filteredWeights = append(filteredWeights, rb.Weight)
		}
	}
	return weightedChoice(filteredBackgrounds, filteredWeights, rng)
}

func generateCharacter(hexKey string) {
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil || len(keyBytes) != 32 {
		fmt.Println("Invalid key. Must be a 32-byte hex string.")
		return
	}

	hash := sha256.Sum256(keyBytes)
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	rng := rand.New(rand.NewSource(seed))

	race := weightedChoice(races, raceWeights, rng)
	class := chooseClassByRace(race, rng)
	background := chooseBackgroundByRace(race, rng)
	alignment := weightedChoice(alignments, alignmentWeights, rng)

	stats := map[string]int{
		"Strength":     rollStat(rng),
		"Dexterity":    rollStat(rng),
		"Constitution": rollStat(rng),
		"Intelligence": rollStat(rng),
		"Wisdom":       rollStat(rng),
		"Charisma":     rollStat(rng),
	}

	fmt.Println("Generated Character:")
	fmt.Println("Race:", race)
	fmt.Println("Class:", class)
	fmt.Println("Background:", background)
	fmt.Println("Alignment:", alignment)
	fmt.Println("Stats:")
	for stat, value := range stats {
		fmt.Printf("  %s: %d\n", stat, value)
	}
}

func main() {
	var inputKey string
	fmt.Print("Enter a 32-byte hex key: ")
	fmt.Scanln(&inputKey)
	generateCharacter(inputKey)
}
