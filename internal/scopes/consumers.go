// Package scopes implements the global and hybrid install scopes, the
// consumer registry, and runtime garbage collection (Spec §8.7, §9).
package scopes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// ConsumersName is the machine-level registry of project checkouts.
const ConsumersName = "consumers.json"

// consumersSchemaVersion is the only registry schema Curator writes or trusts.
const consumersSchemaVersion = 1

// LoadConsumers returns the registered checkout paths for read-only callers.
// An absent or untrustworthy registry reads as no consumers; every caller that
// rewrites the registry uses readConsumers instead, which distinguishes the two
// so an unreadable registry is never overwritten.
func LoadConsumers(home string) []string {
	consumers, err := readConsumers(home)
	if err != nil {
		return nil
	}
	return consumers
}

// RecordConsumer adds a checkout to the registry, deduplicated and sorted.
//
// It refuses to write over a registry it could not read: merging into an empty
// set would silently unregister every other checkout, and those checkouts are
// what keeps their installed artifacts from being collected.
func RecordConsumer(home, projectRoot string) error {
	resolved, err := filepath.Abs(projectRoot)
	if err != nil {
		resolved = projectRoot
	}
	set := map[string]bool{resolved: true}
	existing, err := readConsumers(home)
	if err != nil {
		return fmt.Errorf("read the consumer registry %s: %w", filepath.Join(home, ConsumersName), err)
	}
	for _, entry := range existing {
		set[entry] = true
	}
	return writeConsumers(home, set)
}

// ReplaceConsumers rewrites the registry (used by GC pruning).
func ReplaceConsumers(home string, consumers []string) error {
	set := map[string]bool{}
	for _, entry := range consumers {
		set[entry] = true
	}
	return writeConsumers(home, set)
}

func writeConsumers(home string, set map[string]bool) error {
	payload, err := ConsumersPayload(set)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, ConsumersName), payload, 0o644)
}

// ConsumersPayload renders the canonical registry bytes for one consumer set.
// Staging and direct writes share it so a staged ledger is byte-identical to
// the one a direct write would have produced.
func ConsumersPayload(set map[string]bool) ([]byte, error) {
	var list []string
	for entry, present := range set {
		if present {
			list = append(list, entry)
		}
	}
	sort.Strings(list)
	if list == nil {
		list = []string{}
	}
	payload, err := json.MarshalIndent(map[string]any{"schema_version": consumersSchemaVersion, "consumers": list}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// readConsumers reports an absent registry as no consumers and any registry
// that exists but cannot be trusted as an error, so a caller can tell "no
// consumers" from "unknown consumers".
func readConsumers(home string) ([]string, error) {
	payload, err := os.ReadFile(filepath.Join(home, ConsumersName)) // #nosec G304 -- machine home
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return parseConsumers(payload)
}

// parseConsumers accepts only the exact registry shape Curator writes.
//
// Leniency here is unsafe: a registry that is JSON but not this shape would
// otherwise read as "no consumers registered", which is indistinguishable from
// a machine where every project was legitimately removed — and that reading is
// what would let maintenance prune live checkouts and then collect the build
// artifacts they still reference.
//
// The document is therefore read as a token stream instead of being decoded
// into a struct. Struct decoding silently accepts a repeated object member and
// keeps the last one, so {"schema_version":1,"consumers":["/live"],"consumers":[]}
// would read as an empty but *trusted* registry: maintenance would rewrite it
// empty, and the next pass would no longer visit the checkout whose marker
// protects a build artifact. An ambiguous document is not a registry, so it is
// refused here, before any writer can normalize the ambiguity away.
//
// Repetition inside the consumers array is a different matter and stays
// accepted: deduplicating it cannot drop a checkout, so it carries none of the
// ambiguity that makes a repeated member unsafe.
func parseConsumers(payload []byte) ([]string, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := expectDelim(decoder, '{', "registry"); err != nil {
		return nil, err
	}
	var (
		version   *int
		consumers []string
		seen      = map[string]bool{}
	)
	for decoder.More() {
		name, err := memberName(decoder)
		if err != nil {
			return nil, err
		}
		if seen[name] {
			return nil, fmt.Errorf("registry repeats the %q member", name)
		}
		seen[name] = true
		switch name {
		case "schema_version":
			value, err := decodeRegistryVersion(decoder)
			if err != nil {
				return nil, err
			}
			version = &value
		case "consumers":
			list, err := decodeConsumerPaths(decoder)
			if err != nil {
				return nil, err
			}
			consumers = list
		default:
			return nil, fmt.Errorf("registry holds the unsupported member %q", name)
		}
	}
	if err := expectDelim(decoder, '}', "registry"); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("trailing content after the registry object")
	}
	if version == nil {
		return nil, fmt.Errorf("schema_version is missing")
	}
	if *version != consumersSchemaVersion {
		return nil, fmt.Errorf("unsupported registry schema_version %d", *version)
	}
	if consumers == nil {
		return nil, fmt.Errorf("consumers is missing or null")
	}
	return consumers, nil
}

// decodeRegistryVersion reads schema_version as an exact JSON integer. A
// fractional or exponent spelling is refused rather than rounded, because a
// registry Curator did not write is a registry it cannot reason about.
func decodeRegistryVersion(decoder *json.Decoder) (int, error) {
	token, err := decoder.Token()
	if err != nil {
		return 0, err
	}
	number, ok := token.(json.Number)
	if !ok {
		return 0, fmt.Errorf("schema_version is not a number")
	}
	value, err := strconv.Atoi(number.String())
	if err != nil {
		return 0, fmt.Errorf("schema_version %s is not an integer", number.String())
	}
	return value, nil
}

// decodeConsumerPaths reads the checkout array. An empty array is a registry
// with no consumers; anything that is not an array of absolute paths is an
// unknown registry.
func decodeConsumerPaths(decoder *json.Decoder) ([]string, error) {
	if err := expectDelim(decoder, '[', "consumers"); err != nil {
		return nil, err
	}
	list := []string{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		entry, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("consumers holds a checkout path that is not a string")
		}
		if entry == "" {
			return nil, fmt.Errorf("consumers holds an empty checkout path")
		}
		if !filepath.IsAbs(entry) {
			return nil, fmt.Errorf("consumer %q is not an absolute checkout path", entry)
		}
		list = append(list, entry)
	}
	if err := expectDelim(decoder, ']', "consumers"); err != nil {
		return nil, err
	}
	return list, nil
}

func memberName(decoder *json.Decoder) (string, error) {
	token, err := decoder.Token()
	if err != nil {
		return "", err
	}
	name, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("registry holds a member name that is not a string")
	}
	return name, nil
}

func expectDelim(decoder *json.Decoder, want json.Delim, subject string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); !ok || delim != want {
		return fmt.Errorf("%s is not the expected %s structure", subject, want.String())
	}
	return nil
}
