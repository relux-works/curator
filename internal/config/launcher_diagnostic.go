package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Diagnostic carries the fixed code/path vocabulary of the notification
// corpus while retaining the ordinary loader's compatibility text and cause.
// Neither field contains rejected configuration names or values.
type Diagnostic struct {
	Code  string
	Path  string
	cause error
}

func (e *Diagnostic) Error() string { return e.cause.Error() }
func (e *Diagnostic) Unwrap() error { return e.cause }
func configDiagnostic(code, path string, err error) error {
	var existing *Diagnostic
	if errors.As(err, &existing) {
		return err
	}
	return &Diagnostic{Code: code, Path: path, cause: err}
}

// Keep predecessor numeric decoding unchanged. The schema-8 and schema-9
// raw-file versions retain their tokens, so decimal/exponent spellings cannot
// become the required integer tokens.
// Programmatic Parse callers have no raw-file token to validate.
func decodeConfigObject(payload []byte, path string) (map[string]any, error) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, configDiagnostic("config_json_invalid", "$", fmt.Errorf("malformed JSON in config %s", path))
	}
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, configDiagnostic("config_shape_invalid", "$", fmt.Errorf("config %s must be a JSON object", path))
	}
	if version, ok := object["schema_version"].(float64); ok && (version == SchemaVersion8 || version == SchemaVersion9) {
		var tokens map[string]json.RawMessage
		if err := json.Unmarshal(payload, &tokens); err != nil {
			return nil, err
		}
		token := strings.TrimSpace(string(tokens["schema_version"]))
		if token != strconv.Itoa(int(version)) {
			object["schema_version"] = json.Number(token)
		}
	}
	return object, nil
}
