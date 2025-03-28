package functions

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"nostr-hero/src/types"
	"sort"
)

// CreateDeterministicRNG creates a seeded random number generator from a hex key
func CreateDeterministicRNG(hexKey string) (*rand.Rand, error) {
	// Decode the hex key
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil || len(keyBytes) != 32 {
		return nil, fmt.Errorf("invalid key, must be a 32-byte hex string")
	}

	// Create deterministic RNG
	hash := sha256.Sum256(keyBytes)
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	return rand.New(rand.NewSource(seed)), nil
}

// CreateDeterministicSeed generates a consistent seed from a given hexKey and a context string.
func CreateDeterministicSeed(hexKey, context string) int64 {
	hash := sha256.Sum256([]byte(hexKey + context))
	return int64(binary.BigEndian.Uint64(hash[:8]))
}

// Chooses an option based on weights deterministically
func DeterministicWeightedChoice(options []string, weights []int, seed int64) string {
	if len(options) == 0 {
		return ""
	}

	// Sort options for consistency
	sortedIndices := make([]int, len(options))
	for i := range options {
		sortedIndices[i] = i
	}
	sort.Slice(sortedIndices, func(i, j int) bool {
		return options[sortedIndices[i]] < options[sortedIndices[j]]
	})

	sortedOptions := make([]string, len(options))
	sortedWeights := make([]int, len(weights))
	for i, idx := range sortedIndices {
		sortedOptions[i] = options[idx]
		sortedWeights[i] = weights[idx]
	}

	// Compute cumulative weights
	totalWeight := 0
	for _, w := range sortedWeights {
		totalWeight += w
	}

	rng := rand.New(rand.NewSource(seed))
	randomValue := rng.Intn(totalWeight)

	accumulatedWeight := 0
	for i, w := range sortedWeights {
		accumulatedWeight += w
		if randomValue < accumulatedWeight {
			return sortedOptions[i]
		}
	}

	return sortedOptions[len(sortedOptions)-1] // Fallback (should never be reached)
}

// Normalizes weights to sum up to 100
func NormalizeWeights(options []types.WeightedOption) []types.WeightedOption {
	totalWeight := 0
	for _, opt := range options {
		totalWeight += opt.Weight
	}

	if totalWeight == 100 {
		return options // Already normalized
	}

	// Scale weights proportionally to sum to 100
	normalized := make([]types.WeightedOption, len(options))
	for i, opt := range options {
		normalized[i] = types.WeightedOption{
			Name:   opt.Name,
			Weight: (opt.Weight * 100) / totalWeight, // Normalize to 100
		}
	}

	return normalized
}

// Generate a race based on deterministic choice
func GenerateRace(hexKey string, weightData *types.WeightData) string {
	seed := CreateDeterministicSeed(hexKey, "race")
	return DeterministicWeightedChoice(weightData.Races, weightData.RaceWeights, seed)
}

// Generate a class based on race
func GenerateClass(hexKey string, weightData *types.WeightData, race string) string {
	seed := CreateDeterministicSeed(hexKey, "class_"+race)
	classOptions := []string{}
	classWeights := []int{}

	for class, weight := range weightData.ClassWeightsByRace[race] {
		classOptions = append(classOptions, class)
		classWeights = append(classWeights, weight)
	}

	return DeterministicWeightedChoice(classOptions, classWeights, seed)
}

// Generate a background based on class
func GenerateBackground(hexKey string, weightData *types.WeightData, class string) string {
	seed := CreateDeterministicSeed(hexKey, "background_"+class)
	backgroundOptions := []string{}
	backgroundWeights := []int{}

	for background, weight := range weightData.BackgroundWeightsByClass[class] {
		backgroundOptions = append(backgroundOptions, background)
		backgroundWeights = append(backgroundWeights, weight)
	}

	return DeterministicWeightedChoice(backgroundOptions, backgroundWeights, seed)
}

// Generate an alignment
func GenerateAlignment(hexKey string, weightData *types.WeightData) string {
	seed := CreateDeterministicSeed(hexKey, "alignment")
	return DeterministicWeightedChoice(weightData.Alignments, weightData.AlignmentWeights, seed)
}

// Minimum stat values for each class (unchanged from original)
var MinimumStatsByClass = map[string]map[string]int{
	"Paladin":    {"Strength": 12, "Charisma": 12},
	"Sorcerer":   {"Charisma": 12, "Constitution": 12},
	"Warlock":    {"Charisma": 12, "Wisdom": 12},
	"Bard":       {"Charisma": 12, "Dexterity": 12},
	"Fighter":    {"Strength": 12, "Dexterity": 12},
	"Barbarian":  {"Strength": 12, "Constitution": 12},
	"Monk":       {"Dexterity": 12, "Wisdom": 12},
	"Rogue":      {"Dexterity": 12, "Intelligence": 12},
	"Cleric":     {"Wisdom": 12, "Charisma": 12},
	"Druid":      {"Wisdom": 12, "Intelligence": 12},
	"Ranger":     {"Dexterity": 12, "Wisdom": 12},
	"Wizard":     {"Intelligence": 12, "Wisdom": 12},
}

// Generate stats with class-based minimums
func GenerateStats(hexKey string, class string) map[string]int {
	seed := CreateDeterministicSeed(hexKey, "stats")
	rng := rand.New(rand.NewSource(seed))

	// Roll initial stats
	stats := map[string]int{
		"Strength":     rollStat(rng),
		"Dexterity":    rollStat(rng),
		"Constitution": rollStat(rng),
		"Intelligence": rollStat(rng),
		"Wisdom":       rollStat(rng),
		"Charisma":     rollStat(rng),
	}

	// Enforce class minimums
	if minStats, exists := MinimumStatsByClass[class]; exists {
		for stat, minValue := range minStats {
			for stats[stat] < minValue {
				stats[stat] = rollStat(rng)
			}
		}
	}

	// Ensure stats remain within 8-16 range
	for stat, value := range stats {
		if value < 8 {
			stats[stat] = 8
		} else if value > 16 {
			stats[stat] = 16
		}
	}

	return stats
}

// Generates a deterministic character
func GenerateCharacter(hexKey string, weightData *types.WeightData) types.Character {
	race := GenerateRace(hexKey, weightData)
	class := GenerateClass(hexKey, weightData, race)
	background := GenerateBackground(hexKey, weightData, class)
	alignment := GenerateAlignment(hexKey, weightData)
	stats := GenerateStats(hexKey, class)

	return types.Character{
		Race:       race,
		Class:      class,
		Background: background,
		Alignment:  alignment,
		Stats:      stats,
	}
}

// Existing helper functions (rollStat and weightedChoice) remain the same
func rollStat(rng *rand.Rand) int {
	dice := []int{
		rng.Intn(6) + 1,
		rng.Intn(6) + 1,
		rng.Intn(6) + 1,
		rng.Intn(6) + 1,
	}
	minIndex := 0
	for i, v := range dice {
		if v < dice[minIndex] {
			minIndex = i
		}
	}
	sum := 0
	for i, v := range dice {
		if i != minIndex {
			sum += v
		}
	}
	return sum
}
