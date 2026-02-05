// Package character handles character generation and related logic
package character

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"

	"pubkey-quest/types"
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

// DeterministicWeightedChoice chooses an option based on weights deterministically
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

// NormalizeWeights normalizes weights to sum up to 100
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

// MinimumStatsByClass defines minimum stat values for each class's primary stats
var MinimumStatsByClass = map[string]map[string]int{
	"Paladin":   {"Strength": 15, "Charisma": 15},
	"Sorcerer":  {"Charisma": 15, "Constitution": 15},
	"Warlock":   {"Charisma": 15, "Wisdom": 15},
	"Bard":      {"Charisma": 15, "Dexterity": 15},
	"Fighter":   {"Strength": 15, "Dexterity": 15},
	"Barbarian": {"Strength": 15, "Constitution": 15},
	"Monk":      {"Dexterity": 15, "Wisdom": 15},
	"Rogue":     {"Dexterity": 15, "Intelligence": 15},
	"Cleric":    {"Wisdom": 15, "Charisma": 15},
	"Druid":     {"Wisdom": 15, "Intelligence": 15},
	"Ranger":    {"Dexterity": 15, "Wisdom": 15},
	"Wizard":    {"Intelligence": 15, "Wisdom": 15},
}

// Stat generation constants
const (
	StatMin       = 10 // Minimum value for any stat
	StatMax       = 16 // Maximum value for any stat
	ClassStatMin  = 15 // Minimum value for class primary stats
	TotalStatBase = 60 // 6 stats × 10 minimum
)

// StatTier represents a tier of total stat points with its weight
type StatTier struct {
	MinTotal int
	MaxTotal int
	Weight   int // Percentage weight (should sum to 100)
}

// StatTiers defines the weighted distribution for starting stat totals
// 79-81: 25%, 82-86: 65%, 87-89: 10%
var StatTiers = []StatTier{
	{MinTotal: 79, MaxTotal: 81, Weight: 25},
	{MinTotal: 82, MaxTotal: 86, Weight: 65},
	{MinTotal: 87, MaxTotal: 89, Weight: 10},
}

// StatNames is the canonical ordering of stats for deterministic iteration
var StatNames = []string{
	"Strength",
	"Dexterity",
	"Constitution",
	"Intelligence",
	"Wisdom",
	"Charisma",
}

// GenerateRace generates a race based on deterministic choice
func GenerateRace(hexKey string, weightData *types.WeightData) string {
	seed := CreateDeterministicSeed(hexKey, "race")
	return DeterministicWeightedChoice(weightData.Races, weightData.RaceWeights, seed)
}

// GenerateClass generates a class based on race
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

// GenerateBackground generates a background based on class
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

// GenerateAlignment generates an alignment
func GenerateAlignment(hexKey string, weightData *types.WeightData) string {
	seed := CreateDeterministicSeed(hexKey, "alignment")
	return DeterministicWeightedChoice(weightData.Alignments, weightData.AlignmentWeights, seed)
}

// GenerateStats generates stats using weighted tier system
// Total stat points range from 79-89 with weighted distribution:
// - 79-81: 25% chance
// - 82-86: 65% chance
// - 87-89: 10% chance
// Individual stats are clamped between 10-16, with class primary stats at minimum 15
func GenerateStats(hexKey string, class string) map[string]int {
	// Step 1: Select tier based on weights
	tierSeed := CreateDeterministicSeed(hexKey, "stat_tier")
	tier := selectStatTier(tierSeed)

	// Step 2: Select exact total within the tier
	totalSeed := CreateDeterministicSeed(hexKey, "stat_total")
	totalRng := rand.New(rand.NewSource(totalSeed))
	totalStats := tier.MinTotal + totalRng.Intn(tier.MaxTotal-tier.MinTotal+1)

	// Step 3: Initialize all stats to minimum (10)
	stats := make(map[string]int)
	for _, name := range StatNames {
		stats[name] = StatMin
	}
	currentTotal := TotalStatBase // 60

	// Step 4: Set class primary stats to 15
	if classStats, exists := MinimumStatsByClass[class]; exists {
		for statName := range classStats {
			stats[statName] = ClassStatMin
			currentTotal += (ClassStatMin - StatMin) // Add 5 per class stat
		}
	}

	// Step 5: Distribute remaining points
	remainingPoints := totalStats - currentTotal
	if remainingPoints > 0 {
		distributeStatPoints(hexKey, stats, remainingPoints, class)
	}

	return stats
}

// selectStatTier picks a tier based on weighted probability
func selectStatTier(seed int64) StatTier {
	rng := rand.New(rand.NewSource(seed))
	roll := rng.Intn(100) // 0-99

	cumulative := 0
	for _, tier := range StatTiers {
		cumulative += tier.Weight
		if roll < cumulative {
			return tier
		}
	}

	// Fallback to last tier (shouldn't happen)
	return StatTiers[len(StatTiers)-1]
}

// distributeStatPoints distributes extra points deterministically across stats
func distributeStatPoints(hexKey string, stats map[string]int, points int, class string) {
	// Create a deterministic shuffle order for stats
	shuffleSeed := CreateDeterministicSeed(hexKey, "stat_shuffle")
	rng := rand.New(rand.NewSource(shuffleSeed))

	// Create shuffled order of stat indices
	order := make([]int, len(StatNames))
	for i := range order {
		order[i] = i
	}
	rng.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	// Distribute points one at a time, cycling through the shuffled order
	distributed := 0
	maxIterations := points * len(StatNames) // Safety limit

	for distributed < points && maxIterations > 0 {
		maxIterations--

		for _, idx := range order {
			if distributed >= points {
				break
			}

			statName := StatNames[idx]
			if stats[statName] < StatMax {
				stats[statName]++
				distributed++
			}
		}
	}
}

// GenerateCharacter generates a deterministic character from a hex key
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

