package startupverification

import "encoding/json"

// ParseLevelObject admits only the closed, integer-valued policy object.
func ParseLevelObject(raw any) (int, error) {
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 1 {
		return 0, Invalid("expected only level")
	}
	var level int
	switch value := object["level"].(type) {
	case int:
		level = value
	case float64:
		// Existing config mutation projections use float64 for decoded JSON.
		// File readers validate numeric spelling before this lossy projection.
		if value < 1 || value > 4 || value != float64(int(value)) {
			return 0, Invalid("level must be an integer from 1 to 4")
		}
		level = int(value)
	case int64:
		if value < 1 || value > 4 {
			return 0, Invalid("level must be 1 through 4")
		}
		level = int(value)
	case json.Number:
		if len(value.String()) != 1 || value.String()[0] < '1' || value.String()[0] > '4' {
			return 0, Invalid("level must be an integer from 1 to 4")
		}
		level = int(value.String()[0] - '0')
	default:
		return 0, Invalid("level must be an integer from 1 to 4")
	}
	if _, err := EffectiveLevel(level); err != nil {
		return 0, err
	}
	return level, nil
}
