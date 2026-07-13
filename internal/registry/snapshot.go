package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// DefaultSnapshotMaxAge is the staleness bound: an older reachable snapshot
// counts as a frozen view (Spec §13.4).
const DefaultSnapshotMaxAge = 7 * 24 * time.Hour

// DefaultSnapshotClockSkew is the maximum accepted future timestamp.
const DefaultSnapshotClockSkew = 5 * time.Minute

var hex256RE = regexp.MustCompile(`^[0-9a-f]{64}$`)

type snapshotState struct {
	HighestVersion int    `json:"highest_version"`
	Head           string `json:"head,omitempty"`
	MerkleRoot     string `json:"merkle_root,omitempty"`
	LogSize        int    `json:"log_size,omitempty"`
}

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
		state := readSnapshotState(stateFile)
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
		schema, schemaOK := intValue(snapshot["schema_version"])
		version, versionOK := intValue(snapshot["version"])
		logSize, logSizeOK := intValue(snapshot["log_size"])
		head, headOK := snapshot["head"].(string)
		merkleRoot, rootOK := snapshot["merkle_root"].(string)
		if !schemaOK || schema != 1 || !versionOK || !logSizeOK || version < 0 || logSize < 0 || version < logSize ||
			!headOK || !rootOK || !hex256RE.MatchString(head) || !hex256RE.MatchString(merkleRoot) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot is malformed", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if version < state.HighestVersion {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot version moved backward; possible rollback", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if version == state.HighestVersion && state.Head != "" &&
			(state.Head != head || state.MerkleRoot != merkleRoot || state.LogSize != logSize) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot changed without advancing version; possible equivocation", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		created, ok := parseISO8601(snapshot["created_at"])
		createdText, _ := snapshot["created_at"].(string)
		if !ok || created.UTC().Format(time.RFC3339) != createdText || now.Sub(created) > maxAge {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot is stale", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if created.After(now.Add(DefaultSnapshotClockSkew)) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot timestamp is too far in the future", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		writeSnapshotState(stateFile, snapshotState{HighestVersion: version, Head: head, MerkleRoot: merkleRoot, LogSize: logSize})
	}
	return tampered, warnings
}

func urlDigest(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])[:16]
}

func readSnapshotState(path string) snapshotState {
	payload, err := os.ReadFile(path) // #nosec G304 -- cache state under the machine home
	if err != nil {
		return snapshotState{}
	}
	var data snapshotState
	if err := json.Unmarshal(payload, &data); err != nil {
		return snapshotState{}
	}
	return data
}

func writeSnapshotState(path string, state snapshotState) {
	payload, _ := json.Marshal(state)
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, payload, 0o600); err == nil {
		_ = os.Rename(temporary, path)
	}
}

func intValue(raw any) (int, bool) {
	switch value := raw.(type) {
	case float64:
		if value == float64(int(value)) {
			return int(value), true
		}
	case int:
		return value, true
	case int64:
		if value >= 0 && value <= 9007199254740991 {
			return int(value), true
		}
	case json.Number:
		parsed, err := strconv.ParseInt(string(value), 10, 64)
		if err == nil && parsed >= 0 && parsed <= 9007199254740991 {
			return int(parsed), true
		}
	}
	return 0, false
}
