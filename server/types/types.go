package types

type RegistryEntry struct {
	Npub      string    `json:"npub"`
	PubKey    string    `json:"pubkey"`
	Character Character `json:"character"`
}

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

type Character struct {
	Race       string         `json:"race"`
	Class      string         `json:"class"`
	Background string         `json:"background"`
	Alignment  string         `json:"alignment"`
	Stats      map[string]int `json:"stats"`
}

type WeightData struct {
	Races                    []string                  `json:"Races"`
	RaceWeights              []int                     `json:"RaceWeights"`
	ClassWeightsByRace       map[string]map[string]int `json:"classWeightsByRace"`
	BackgroundWeightsByClass map[string]map[string]int `json:"backgroundWeightsByClass"`
	Alignments               []string                  `json:"Alignments"`
	AlignmentWeights         []int                     `json:"AlignmentWeights"`
}

// Weighted option structure
type WeightedOption struct {
	Name   string
	Weight int
}