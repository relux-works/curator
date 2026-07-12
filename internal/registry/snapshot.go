package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultSnapshotMaxAge is the staleness bound: an older reachable snapshot
// counts as a frozen view (Spec §13.4).
const DefaultSnapshotMaxAge = 7 * 24 * time.Hour

// SnapshotFetchFn returns the signed snapshot object of a registry.
type SnapshotFetchFn func(url string) (map[string]any, error)

// CheckSnapshots verifies each registry snapshot and returns the URLs to
// treat as tampered plus warnings (Spec §13.4). A registry is excluded when
// its snapshot signature fails, its version moved backward relative to the
// persisted highest accepted version, or it is stale. An unreachable
// snapshot warns but does not exclude.
func CheckSnapshots(registries []Registry, cacheDir string, fetch SnapshotFetchFn, now time.Time, maxAge time.Duration) (map[string]bool, []string) {
	if maxAge == 0 {
		maxAge = DefaultSnapshotMaxAge
	}
	_ = os.MkdirAll(cacheDir, 0o755)
	tampered := map[string]bool{}
	var warnings []string
	for _, reg := range registries {
		if len(reg.PublicKeys) == 0 {
			continue
		}
		stateFile := filepath.Join(cacheDir, "snapshot-"+urlDigest(reg.URL)+".json")
		highest := readHighestVersion(stateFile)
		snapshot, err := fetch(reg.URL)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot unavailable: %v", reg.Name, err))
			continue
		}
		if !VerifySigned(snapshot, reg.PublicKeys) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot signature failed verification", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		version, ok := intValue(snapshot["version"])
		if !ok || version < highest {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot version moved backward; possible rollback", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		created, ok := parseISO8601(snapshot["created_at"])
		if !ok || now.Sub(created) > maxAge {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot is stale", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		writeHighestVersion(stateFile, version)
	}
	return tampered, warnings
}

func urlDigest(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])[:16]
}

func readHighestVersion(path string) int {
	payload, err := os.ReadFile(path) // #nosec G304 -- cache state under the machine home
	if err != nil {
		return 0
	}
	var data struct {
		HighestVersion int `json:"highest_version"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return 0
	}
	return data.HighestVersion
}

func writeHighestVersion(path string, version int) {
	payload, _ := json.Marshal(map[string]int{"highest_version": version})
	_ = os.WriteFile(path, payload, 0o644)
}

func intValue(raw any) (int, bool) {
	switch value := raw.(type) {
	case float64:
		if value == float64(int(value)) {
			return int(value), true
		}
	case int:
		return value, true
	}
	return 0, false
}
