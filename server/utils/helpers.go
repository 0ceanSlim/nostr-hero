package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/bech32"

	"nostr-hero/types"
)

// Helper function to prepend a directory path to a list of filenames
func PrependDir(dir string, files []string) []string {
	var fullPaths []string
	for _, file := range files {
		fullPaths = append(fullPaths, dir+file)
	}
	return fullPaths
}

// GenerateRandomToken creates a cryptographically secure random token
// of the specified length in bytes (output will be twice this length as hex)
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// In a real application, handle this error better
		// For now, let's log and generate something less secure but still random
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}

func DecodeNpub(npub string) (string, error) {
	hrp, data, err := bech32.Decode(npub)
	if err != nil {
		return "", err
	}
	if hrp != "npub" {
		return "", errors.New("invalid hrp")
	}

	decodedData, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%x", decodedData)), nil
}

func LoadWeights(filename string) (*types.WeightData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var data types.WeightData
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// Sorts JSON input deterministically
func SortJSONInput(data map[string][]types.WeightedOption) map[string][]types.WeightedOption {
	sortedData := make(map[string][]types.WeightedOption)
	
	// Extract keys and sort them
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Sort values within each key
	for _, key := range keys {
		options := data[key]
		sort.Slice(options, func(i, j int) bool {
			return options[i].Name < options[j].Name
		})
		sortedData[key] = options
	}

	return sortedData
}