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
	Race      string
	Background string
	Weight    int
}

var races = []string{"Human", "Elf", "Dwarf", "Halfling", "Orc", "Tiefling", "Gnome", "Dragonborn"}
var raceWeights = []int{30, 15, 15, 10, 10, 10, 5, 5}

var classWeightsByRace = []RaceClassWeight{
	{"Human", "Fighter", 20}, {"Human", "Wizard", 15}, {"Human", "Rogue", 15},
	{"Elf", "Ranger", 25}, {"Elf", "Wizard", 20}, {"Elf", "Rogue", 15},
	{"Dwarf", "Fighter", 25}, {"Dwarf", "Cleric", 20}, {"Dwarf", "Paladin", 15},
	{"Halfling", "Rogue", 30}, {"Halfling", "Bard", 20},
	{"Tiefling", "Warlock", 30}, {"Tiefling", "Sorcerer", 25},
	{"Gnome", "Wizard", 25}, {"Gnome", "Bard", 20},
	{"Dragonborn", "Paladin", 30},
}

var backgrounds = []string{"Soldier", "Sage", "Outlander", "Noble", "Acolyte", "Charlatan", "Entertainer", "Hermit"}

var backgroundWeightsByRace = []RaceBackgroundWeight{
	{"Human", "Soldier", 20}, {"Human", "Noble", 20}, {"Human", "Charlatan", 15},
	{"Elf", "Sage", 25}, {"Elf", "Noble", 20}, {"Elf", "Outlander", 15},
	{"Dwarf", "Soldier", 25}, {"Dwarf", "Acolyte", 20}, {"Dwarf", "Outlander", 15},
	{"Halfling", "Charlatan", 30}, {"Halfling", "Entertainer", 25}, {"Halfling", "Hermit", 15},
	{"Orc", "Soldier", 30}, {"Orc", "Outlander", 25}, {"Orc", "Hermit", 15},
	{"Tiefling", "Charlatan", 25}, {"Tiefling", "Noble", 20}, {"Tiefling", "Hermit", 15},
	{"Gnome", "Sage", 25}, {"Gnome", "Entertainer", 20}, {"Gnome", "Charlatan", 15},
	{"Dragonborn", "Soldier", 30}, {"Dragonborn", "Noble", 25}, {"Dragonborn", "Acolyte", 15},
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
