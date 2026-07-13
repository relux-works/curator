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
	"strings"
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

type parsedSnapshot struct {
	Version    int
	LogSize    int
	Head       string
	MerkleRoot string
	CreatedAt  time.Time
}

// SnapshotFetchFn returns the signed snapshot object of a registry.
type SnapshotFetchFn func(url string) (map[string]any, error)

// CheckSnapshots verifies each registry snapshot and returns the URLs to
// treat as tampered plus warnings (Spec §13.4). A registry is excluded when
// its snapshot signature fails, its version moved backward relative to the
// persisted highest accepted version, or it is stale. An unreachable
// snapshot warns but does not exclude.
func CheckSnapshots(registries []Registry, cacheDir string, fetch SnapshotFetchFn, now time.Time, maxAge time.Duration) (map[string]bool, []string) {
	return CheckSnapshotsWithPolicy(registries, cacheDir, fetch, now, maxAge, DefaultSnapshotClockSkew)
}

// CheckSnapshotsWithPolicy applies explicit manager-config age and clock-skew
// bounds. A zero clock skew is literal; a zero max age retains the historical
// default for callers of CheckSnapshots.
func CheckSnapshotsWithPolicy(registries []Registry, cacheDir string, fetch SnapshotFetchFn, now time.Time, maxAge, clockSkew time.Duration) (map[string]bool, []string) {
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
		parsed, err := parseSnapshot(snapshot)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot is malformed: %v", reg.Name, err))
			tampered[reg.URL] = true
			continue
		}
		if !VerifySigned(snapshot, reg.PublicKeys) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot signature failed verification", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if parsed.Version < state.HighestVersion {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot version moved backward; possible rollback", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if parsed.Version == state.HighestVersion && state.Head != "" &&
			(state.Head != parsed.Head || state.MerkleRoot != parsed.MerkleRoot || state.LogSize != parsed.LogSize) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot changed without advancing version; possible equivocation", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if now.Sub(parsed.CreatedAt) > maxAge {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot is stale", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		if parsed.CreatedAt.After(now.Add(clockSkew)) {
			warnings = append(warnings, fmt.Sprintf("registry %s snapshot timestamp is too far in the future", reg.Name))
			tampered[reg.URL] = true
			continue
		}
		writeSnapshotState(stateFile, snapshotState{HighestVersion: parsed.Version, Head: parsed.Head, MerkleRoot: parsed.MerkleRoot, LogSize: parsed.LogSize})
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

func parseSnapshot(snapshot map[string]any) (parsedSnapshot, error) {
	if unknown := unknownKeys(snapshot, "schema_version", "merkle_root", "log_size", "head", "version", "created_at", "sig"); len(unknown) > 0 || len(snapshot) != 7 {
		return parsedSnapshot{}, fmt.Errorf("must contain exactly the schema 1 snapshot fields")
	}
	if !integerEquals(snapshot["schema_version"], 1) {
		return parsedSnapshot{}, fmt.Errorf("schema_version must be 1")
	}
	version, versionOK := nonNegativeSafeInteger(snapshot["version"])
	logSize, logSizeOK := nonNegativeSafeInteger(snapshot["log_size"])
	head, headOK := snapshot["head"].(string)
	merkleRoot, rootOK := snapshot["merkle_root"].(string)
	if !versionOK || !logSizeOK || version < logSize || !headOK || !rootOK ||
		!hex256RE.MatchString(head) || !hex256RE.MatchString(merkleRoot) {
		return parsedSnapshot{}, fmt.Errorf("has invalid version, log size, head, or Merkle root")
	}
	created, ok := parseISO8601(snapshot["created_at"])
	createdText, _ := snapshot["created_at"].(string)
	if !ok || created.UTC().Format("2006-01-02T15:04:05Z") != createdText {
		return parsedSnapshot{}, fmt.Errorf("created_at must be an exact UTC seconds timestamp")
	}
	if err := validateSignatureEnvelope(snapshot["sig"]); err != nil {
		return parsedSnapshot{}, fmt.Errorf("signature: %w", err)
	}
	if _, err := CanonicalBytesChecked(snapshot); err != nil {
		return parsedSnapshot{}, fmt.Errorf("is not valid CCJ-1: %w", err)
	}
	return parsedSnapshot{Version: version, LogSize: logSize, Head: head, MerkleRoot: merkleRoot, CreatedAt: created}, nil
}

func nonNegativeSafeInteger(raw any) (int, bool) {
	const maximum = int64(9007199254740991)
	var parsed int64
	switch value := raw.(type) {
	case int:
		parsed = int64(value)
	case int64:
		parsed = value
	case json.Number:
		var err error
		parsed, err = strconv.ParseInt(string(value), 10, 64)
		if err != nil || strings.ContainsAny(string(value), ".eE+") || strconv.FormatInt(parsed, 10) != string(value) {
			return 0, false
		}
	default:
		return 0, false
	}
	if parsed < 0 || parsed > maximum {
		return 0, false
	}
	return int(parsed), true
}
