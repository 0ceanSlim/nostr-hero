package functions

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"

	"nostr-hero/src/types"
	"nostr-hero/src/utils"
)

// GenerateCharacter creates a character based on a given hex key
func GenerateCharacter(hexKey string) (types.Character, error) {
	// Decode the hex key
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil || len(keyBytes) != 32 {
		return types.Character{}, fmt.Errorf("invalid key, must be a 32-byte hex string")
	}

	// Load weights from JSON
	weightData, err := utils.LoadWeights("data/weights.json")
	if err != nil {
		return types.Character{}, fmt.Errorf("failed to load weights: %v", err)
	}

	// Create deterministic random number generator from key
	hash := sha256.Sum256(keyBytes)
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	rng := rand.New(rand.NewSource(seed))

	// Generate race first
	race := weightedChoice(weightData.Races, weightData.RaceWeights, rng)

	// Prepare class weights for the chosen race
	classOptions := []string{}
	classWeights := []int{}
	for class, weight := range weightData.ClassWeightsByRace[race] {
		classOptions = append(classOptions, class)
		classWeights = append(classWeights, weight)
	}

	// Choose class
	class := weightedChoice(classOptions, classWeights, rng)
	if class == "" {
		return types.Character{}, fmt.Errorf("no valid class found for race: %s", race)
	}

	// Prepare background weights for the chosen race
	backgroundOptions := []string{}
	backgroundWeights := []int{}
	for background, weight := range weightData.BackgroundWeightsByRace[race] {
		backgroundOptions = append(backgroundOptions, background)
		backgroundWeights = append(backgroundWeights, weight)
	}

	// Choose background
	background := weightedChoice(backgroundOptions, backgroundWeights, rng)
	if background == "" {
		return types.Character{}, fmt.Errorf("no valid background found for race: %s", race)
	}

	// Choose alignment using weights
	alignment := weightedChoice(weightData.Alignments, weightData.AlignmentWeights, rng)

	// Generate character
	character := types.Character{
		Race:       race,
		Class:      class,
		Background: background,
		Alignment:  alignment,
		Stats: map[string]int{
			"Strength":     rollStat(rng),
			"Dexterity":    rollStat(rng),
			"Constitution": rollStat(rng),
			"Intelligence": rollStat(rng),
			"Wisdom":       rollStat(rng),
			"Charisma":     rollStat(rng),
		},
	}
	return character, nil
}

// rollStat generates a stat using 4d6 drop lowest method
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

// weightedChoice selects an option based on provided weights
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
