package gameutil

// GetIntValue safely extracts an int from a map, handling both int and float64
func GetIntValue(m map[string]interface{}, key string, defaultValue int) int {
	val, ok := m[key]
	if !ok {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return defaultValue
	}
}

// GetFloatValue safely extracts a float64 from a map, handling both int and float64
func GetFloatValue(m map[string]interface{}, key string, defaultValue float64) float64 {
	val, ok := m[key]
	if !ok {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultValue
	}
}

// Pluralize returns "s" if count != 1, empty string otherwise
func Pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
