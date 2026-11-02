package protocoljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// DecodeClosed is for new internal authority carriers. Unlike encoding/json,
// it requires exact field spelling and presence, recursively rejects unknown
// members, and distinguishes omitted optional values from explicit null. Raw
// JSON members are transport-validated here and require their own semantic
// decoder. Existing public readers keep their independently adopted grammar.
func DecodeClosed(raw []byte, dst any) error {
	if err := Validate(raw); err != nil {
		return err
	}
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("closed destination must be a nonnil pointer")
	}
	if err := closedShape(raw, v.Type().Elem(), "$", 0); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

var rawMessageType = reflect.TypeFor[json.RawMessage]()

func closedShape(raw []byte, typ reflect.Type, path string, depth int) error {
	if depth > 64 {
		return fmt.Errorf("%s: authority nesting limit", path)
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("%s: null is not an authority value", path)
	}
	if typ == rawMessageType {
		return nil
	}
	if typ.Kind() == reflect.Pointer {
		return closedShape(raw, typ.Elem(), path, depth+1)
	}
	switch typ.Kind() {
	case reflect.Struct:
		var members map[string]json.RawMessage
		if err := json.Unmarshal(raw, &members); err != nil || members == nil {
			return fmt.Errorf("%s: expected authority object", path)
		}
		allowed := make(map[string]bool, typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.PkgPath != "" {
				continue
			}
			tags := strings.Split(f.Tag.Get("json"), ",")
			name := tags[0]
			if name == "-" {
				continue
			}
			if name == "" || f.Anonymous {
				return fmt.Errorf("%s: authority field requires an explicit name", path)
			}
			allowed[name] = true
			value, present := members[name]
			optional := false
			for _, tag := range tags[1:] {
				optional = optional || tag == "omitempty"
			}
			if !present {
				if optional {
					continue
				}
				return fmt.Errorf("%s.%s: missing authority member", path, name)
			}
			if err := closedShape(value, f.Type, path+"."+name, depth+1); err != nil {
				return err
			}
		}
		for key := range members {
			if !allowed[key] {
				return fmt.Errorf("%s.%s: unknown authority member", path, key)
			}
		}
	case reflect.Slice, reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return nil
		} // json's base64 byte transport; semantic limits remain with the carrier.
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil || values == nil {
			return fmt.Errorf("%s: expected authority array", path)
		}
		for i, value := range values {
			if err := closedShape(value, typ.Elem(), fmt.Sprintf("%s[%d]", path, i), depth+1); err != nil {
				return err
			}
		}
	case reflect.Map:
		if typ.Key().Kind() != reflect.String {
			return fmt.Errorf("%s: unsupported authority map", path)
		}
		var values map[string]json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil || values == nil {
			return fmt.Errorf("%s: expected authority map", path)
		}
		for key, value := range values {
			if err := closedShape(value, typ.Elem(), path+"["+key+"]", depth+1); err != nil {
				return err
			}
		}
	case reflect.Interface:
		return fmt.Errorf("%s: untyped authority member", path)
	}
	return nil
}
