package closuregraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"golang.org/x/text/unicode/norm"
)

func decodeCanonicalObject(payload []byte, context string) (map[string]any, error) {
	var raw map[string]any
	if err := protocoljson.UnmarshalCanonical(payload, &raw); err != nil {
		return nil, fmt.Errorf("decode %s: %w", context, err)
	}
	return raw, nil
}

func exactFields(raw map[string]any, context string, required, optional []string) error {
	allowed := make(map[string]bool, len(required)+len(optional))
	for _, field := range required {
		allowed[field] = true
		if _, ok := raw[field]; !ok {
			return fmt.Errorf("%s is missing required field %q", context, field)
		}
	}
	for _, field := range optional {
		allowed[field] = true
	}
	for _, field := range sortedMapKeys(raw) {
		if !allowed[field] {
			return fmt.Errorf("%s contains unknown field %q", context, field)
		}
	}
	return nil
}

func requiredString(raw map[string]any, field, context string) (string, error) {
	value, ok := raw[field].(string)
	if !ok {
		return "", fmt.Errorf("%s field %q must be a string", context, field)
	}
	return value, nil
}

func optionalString(raw map[string]any, field, context string) (string, error) {
	value, present := raw[field]
	if !present {
		return "", nil
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%s field %q must be a string", context, field)
	}
	return text, nil
}

func decodeStringFields(raw map[string]any, context string, required, optional []string) (map[string]string, error) {
	values := make(map[string]string, len(required)+len(optional))
	for _, field := range required {
		value, err := requiredString(raw, field, context)
		if err != nil {
			return nil, err
		}
		values[field] = value
	}
	for _, field := range optional {
		value, err := optionalString(raw, field, context)
		if err != nil {
			return nil, err
		}
		values[field] = value
	}
	return values, nil
}

func requireDecodedRecordRoundTrip(payload []byte, record canonicalRecord) error {
	canonical, err := canonicalBytes(record)
	if err != nil {
		return err
	}
	if !bytes.Equal(payload, canonical) {
		return fmt.Errorf("decoded canonical record does not round-trip exactly")
	}
	return nil
}

func requiredBool(raw map[string]any, field, context string) (bool, error) {
	value, ok := raw[field].(bool)
	if !ok {
		return false, fmt.Errorf("%s field %q must be a boolean", context, field)
	}
	return value, nil
}

func requiredInteger(raw map[string]any, field, context string) (int64, error) {
	number, ok := raw[field].(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s field %q must be an integer", context, field)
	}
	value, err := number.Int64()
	if err != nil || value < -protocoljson.MaxSafeInteger || value > protocoljson.MaxSafeInteger {
		return 0, fmt.Errorf("%s field %q must be a safe integer", context, field)
	}
	return value, nil
}

func requiredObject(raw map[string]any, field, context string) (map[string]any, error) {
	value, ok := raw[field].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s field %q must be an object", context, field)
	}
	return value, nil
}

func optionalObject(raw map[string]any, field, context string) (map[string]any, bool, error) {
	value, present := raw[field]
	if !present {
		return nil, false, nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s field %q must be an object", context, field)
	}
	return object, true, nil
}

func requiredStringSlice(raw map[string]any, field, context string) ([]string, error) {
	value, ok := raw[field].([]any)
	if !ok {
		return nil, fmt.Errorf("%s field %q must be an array", context, field)
	}
	result := make([]string, len(value))
	for index, item := range value {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s field %q item %d must be a string", context, field, index)
		}
		result[index] = text
	}
	return result, nil
}

func optionalStringSlice(raw map[string]any, field, context string) ([]string, error) {
	if _, present := raw[field]; !present {
		return nil, nil
	}
	return requiredStringSlice(raw, field, context)
}

func requiredIDSlice(raw map[string]any, field, context string) ([]ID, error) {
	values, err := requiredStringSlice(raw, field, context)
	if err != nil {
		return nil, err
	}
	ids := make([]ID, len(values))
	for index, value := range values {
		ids[index] = ID(value)
	}
	return ids, nil
}

func requiredStringMap(raw map[string]any, field, context string) (map[string]string, error) {
	object, err := requiredObject(raw, field, context)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(object))
	for _, key := range sortedMapKeys(object) {
		value := object[key]
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%s field %q member %q must be a string", context, field, key)
		}
		result[key] = text
	}
	return result, nil
}

func sortedMapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedPlatformRoles(values map[PlatformRole]ID) []PlatformRole {
	roles := make([]PlatformRole, 0, len(values))
	for role := range values {
		roles = append(roles, role)
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i] < roles[j] })
	return roles
}

func validatePortableTextFields(values map[string]string, allowEmpty, skipEmpty bool) error {
	for _, field := range sortedMapKeys(values) {
		value := values[field]
		if skipEmpty && value == "" {
			continue
		}
		if err := validatePortableText(value, field, allowEmpty); err != nil {
			return err
		}
	}
	return nil
}

func validatePortableStringMap(values map[string]string, field string, allowEmptyValue bool) error {
	for _, key := range sortedMapKeys(values) {
		if err := validatePortableText(key, field+" key", false); err != nil {
			return err
		}
		if err := validatePortableText(values[key], field+" value", allowEmptyValue); err != nil {
			return err
		}
	}
	return nil
}

func validateStringSliceFields(values map[string][]string, sortedValues bool) error {
	for _, field := range sortedMapKeys(values) {
		if err := validateStringSlice(values[field], field, sortedValues); err != nil {
			return err
		}
	}
	return nil
}

func validateIDFields(values map[string]ID) error {
	for _, field := range sortedMapKeys(values) {
		if err := validateID(values[field], field); err != nil {
			return err
		}
	}
	return nil
}

func validateIDSliceFields(values map[string][]ID, sortedValues bool) error {
	for _, field := range sortedMapKeys(values) {
		if err := validateIDSlice(values[field], field, sortedValues); err != nil {
			return err
		}
	}
	return nil
}

func idStrings(ids []ID) []string {
	values := make([]string, len(ids))
	for index, id := range ids {
		values[index] = string(id)
	}
	return values
}

func stringsToAny(values []string) []any {
	items := make([]any, len(values))
	for index, value := range values {
		items[index] = value
	}
	return items
}

func idsToAny(values []ID) []any {
	items := make([]any, len(values))
	for index, value := range values {
		items[index] = string(value)
	}
	return items
}

func stringMapToAny(values map[string]string) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func validatePortableText(value, field string, allowEmpty bool) error {
	if (!allowEmpty && value == "") || !utf8.ValidString(value) || !norm.NFC.IsNormalString(value) ||
		utf8.RuneCountInString(value) > 8192 || strings.ContainsRune(value, '\x00') {
		return fmt.Errorf("%s must be portable NFC UTF-8 text", field)
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return fmt.Errorf("%s must not contain control characters", field)
		}
	}
	return nil
}

func validatePortablePath(value, field string) error {
	if !identifiers.PortablePath(value) || !norm.NFC.IsNormalString(value) {
		return fmt.Errorf("%s must be a portable NFC relative path", field)
	}
	return nil
}

func validateID(id ID, field string) error {
	if !id.Valid() {
		return fmt.Errorf("%s must be a canonical sha256 identity", field)
	}
	return nil
}

func validateIDSlice(values []ID, field string, sorted bool) error {
	if values == nil {
		return fmt.Errorf("%s must be an explicit array", field)
	}
	for index, value := range values {
		if err := validateID(value, fmt.Sprintf("%s[%d]", field, index)); err != nil {
			return err
		}
	}
	return validateUniqueStrings(idStrings(values), field, sorted)
}

func validateStringSlice(values []string, field string, sorted bool) error {
	if values == nil {
		return fmt.Errorf("%s must be an explicit array", field)
	}
	for index, value := range values {
		if err := validatePortableText(value, fmt.Sprintf("%s[%d]", field, index), false); err != nil {
			return err
		}
	}
	return validateUniqueStrings(values, field, sorted)
}

func validateUniqueStrings(values []string, field string, sortedValues bool) error {
	seen := make(map[string]bool, len(values))
	for index, value := range values {
		if seen[value] {
			return fmt.Errorf("%s contains duplicate value %q", field, value)
		}
		seen[value] = true
		if sortedValues && index > 0 && values[index-1] > value {
			return fmt.Errorf("%s must be bytewise sorted", field)
		}
	}
	return nil
}

func sortedIDs(values []ID) []ID {
	if values == nil {
		return nil
	}
	result := append([]ID{}, values...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func sortedStrings(values []string) []string {
	if values == nil {
		return nil
	}
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}

func canonicalMapBytes(value map[string]any) ([]byte, error) {
	return protocoljson.MarshalCanonical(value)
}
