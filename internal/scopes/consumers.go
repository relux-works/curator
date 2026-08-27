// Package scopes implements the global and hybrid install scopes, the
// consumer registry, and runtime garbage collection (Spec §8.7, §9).
package scopes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// ConsumersName is the machine-level registry of project checkouts.
const ConsumersName = "consumers.json"

// LoadConsumers returns the registered checkout paths.
func LoadConsumers(home string) []string {
	payload, err := os.ReadFile(filepath.Join(home, ConsumersName)) // #nosec G304 -- machine home
	if err != nil {
		return nil
	}
	var data struct {
		Consumers []string `json:"consumers"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil
	}
	return data.Consumers
}

// RecordConsumer adds a checkout to the registry, deduplicated and sorted.
func RecordConsumer(home, projectRoot string) error {
	resolved, err := filepath.Abs(projectRoot)
	if err != nil {
		resolved = projectRoot
	}
	set := map[string]bool{resolved: true}
	for _, existing := range LoadConsumers(home) {
		set[existing] = true
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
	payload, err := json.MarshalIndent(map[string]any{"schema_version": 1, "consumers": list}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
