package contextlock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Validate key spellings before typed decoding: encoding/json otherwise accepts
// case aliases even with DisallowUnknownFields. RawMessage subtrees keep their
// own strict reader (the plan) or their existing empty-object/array checks.
func validateExactJSONKeys(payload []byte, target any) error {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	return exactJSONKeyValue(value, reflect.TypeOf(target), "$")
}

func exactJSONKeyValue(value any, shape reflect.Type, path string) error {
	for shape.Kind() == reflect.Pointer {
		shape = shape.Elem()
	}
	if shape == reflect.TypeFor[json.RawMessage]() {
		return nil
	}
	switch shape.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return nil
		} // The typed decoder and carrier validation check values.
		fields := make(map[string]reflect.Type, shape.NumField())
		for i := 0; i < shape.NumField(); i++ {
			field := shape.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name != "" && name != "-" {
				fields[name] = field.Type
			}
		}
		for key, child := range object {
			field, ok := fields[key]
			if !ok {
				return fmt.Errorf("exact JSON key required: unknown field %q at %s", key, path)
			}
			if err := exactJSONKeyValue(child, field, path+"."+key); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if items, ok := value.([]any); ok {
			for i, child := range items {
				if err := exactJSONKeyValue(child, shape.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
