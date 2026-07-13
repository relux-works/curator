package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
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
var snapshotStateNameRE = regexp.MustCompile(`^snapshot-[0-9a-f]{16}\.json$`)

const snapshotStateCatalogName = "known-registries.json"

type snapshotState struct {
	HighestVersion int    `json:"highest_version"`
	Head           string `json:"head,omitempty"`
	MerkleRoot     string `json:"merkle_root,omitempty"`
	LogSize        int    `json:"log_size,omitempty"`
}

type snapshotStateCatalog struct {
	SchemaVersion int      `json:"schema_version"`
	States        []string `json:"states"`
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
	tampered := map[string]bool{}
	var warnings []string
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		for _, reg := range registries {
			if len(reg.PublicKeys) > 0 {
				tampered[reg.URL] = true
				warnings = append(warnings, fmt.Sprintf("registry %s rollback state directory is unavailable: %v", reg.Name, err))
			}
		}
		return tampered, warnings
	}
	_ = os.Chmod(cacheDir, 0o700) // #nosec G302 -- owner-only directory access requires traversal bits
	knownStates, err := loadSnapshotStateCatalog(cacheDir)
	if err != nil {
		for _, reg := range registries {
			if len(reg.PublicKeys) > 0 {
				tampered[reg.URL] = true
				warnings = append(warnings, fmt.Sprintf("registry %s rollback state catalog is unavailable: %v", reg.Name, err))
			}
		}
		return tampered, warnings
	}
	for _, reg := range registries {
		if len(reg.PublicKeys) == 0 {
			continue
		}
		stateName := "snapshot-" + urlDigest(reg.URL) + ".json"
		stateFile := filepath.Join(cacheDir, stateName)
		state, stateExists, err := readSnapshotState(stateFile)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("registry %s rollback state is unreadable: %v", reg.Name, err))
			tampered[reg.URL] = true
			continue
		}
		if !stateExists && knownStates[stateName] {
			warnings = append(warnings, fmt.Sprintf("registry %s rollback state is missing after prior use", reg.Name))
			tampered[reg.URL] = true
			continue
		}
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
		if err := writeSnapshotState(stateFile, snapshotState{HighestVersion: parsed.Version, Head: parsed.Head, MerkleRoot: parsed.MerkleRoot, LogSize: parsed.LogSize}); err != nil {
			warnings = append(warnings, fmt.Sprintf("registry %s rollback state could not be persisted: %v", reg.Name, err))
			tampered[reg.URL] = true
			continue
		}
		if !knownStates[stateName] {
			knownStates[stateName] = true
			if err := writeSnapshotStateCatalog(cacheDir, knownStates); err != nil {
				warnings = append(warnings, fmt.Sprintf("registry %s rollback state catalog could not be persisted: %v", reg.Name, err))
				tampered[reg.URL] = true
			}
		}
	}
	return tampered, warnings
}

// MigrateSnapshotStates moves legacy rollback state out of the disposable
// record cache. It refuses malformed or conflicting state so an upgrade cannot
// silently reset the highest accepted registry snapshot.
func MigrateSnapshotStates(legacyDir, stateDir string) error {
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		return err
	}
	_ = os.Chmod(stateDir, 0o700) // #nosec G302 -- owner-only directory access requires traversal bits
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return rebuildSnapshotStateCatalog(stateDir)
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !snapshotStateNameRE.MatchString(entry.Name()) {
			continue
		}
		legacyPath := filepath.Join(legacyDir, entry.Name())
		legacyState, legacyExists, err := readSnapshotState(legacyPath)
		if err != nil {
			return fmt.Errorf("legacy rollback state %s is unreadable: %w", entry.Name(), err)
		}
		if !legacyExists {
			return fmt.Errorf("legacy rollback state %s disappeared during migration", entry.Name())
		}
		targetPath := filepath.Join(stateDir, entry.Name())
		targetState, targetExists, targetErr := readSnapshotState(targetPath)
		if targetErr != nil {
			return fmt.Errorf("rollback state %s is unreadable: %w", entry.Name(), targetErr)
		}
		if targetExists {
			if targetState.HighestVersion < legacyState.HighestVersion ||
				(targetState.HighestVersion == legacyState.HighestVersion && statesConflict(targetState, legacyState)) {
				return fmt.Errorf("rollback state %s conflicts with legacy state", entry.Name())
			}
			if err := os.Remove(legacyPath); err != nil {
				return err
			}
			if err := syncDirectory(legacyDir); err != nil {
				return err
			}
			continue
		}
		if err := os.Chmod(legacyPath, 0o600); err != nil {
			return err
		}
		if err := os.Rename(legacyPath, targetPath); err != nil {
			return err
		}
		if err := syncDirectory(stateDir); err != nil {
			return err
		}
		if err := syncDirectory(legacyDir); err != nil {
			return err
		}
	}
	return rebuildSnapshotStateCatalog(stateDir)
}

func statesConflict(left, right snapshotState) bool {
	return left.Head != right.Head || left.MerkleRoot != right.MerkleRoot || left.LogSize != right.LogSize
}

func urlDigest(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])[:16]
}

func readSnapshotState(path string) (snapshotState, bool, error) {
	payload, exists, err := readProtectedStateFile(path)
	if err != nil {
		return snapshotState{}, exists, err
	}
	if !exists {
		return snapshotState{}, false, nil
	}
	var data snapshotState
	if err := decodeJSON(payload, &data); err != nil {
		return snapshotState{}, true, err
	}
	if data.HighestVersion < 0 || data.LogSize < 0 || data.LogSize > data.HighestVersion {
		return snapshotState{}, true, fmt.Errorf("state has invalid version or log size")
	}
	if !hex256RE.MatchString(data.Head) || !hex256RE.MatchString(data.MerkleRoot) {
		return snapshotState{}, true, fmt.Errorf("state has invalid head or Merkle root")
	}
	return data, true, nil
}

func writeSnapshotState(path string, state snapshotState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return writeProtectedStateFile(path, payload)
}

func loadSnapshotStateCatalog(stateDir string) (map[string]bool, error) {
	path := filepath.Join(stateDir, snapshotStateCatalogName)
	payload, exists, err := readProtectedStateFile(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		entries, readErr := os.ReadDir(stateDir)
		if readErr != nil {
			return nil, readErr
		}
		for _, entry := range entries {
			if !entry.IsDir() && snapshotStateNameRE.MatchString(entry.Name()) {
				return nil, fmt.Errorf("catalog is missing while rollback state exists")
			}
		}
		empty := map[string]bool{}
		if err := writeSnapshotStateCatalog(stateDir, empty); err != nil {
			return nil, err
		}
		return empty, nil
	}
	var catalog snapshotStateCatalog
	if err := decodeJSON(payload, &catalog); err != nil {
		return nil, err
	}
	if catalog.SchemaVersion != 1 || catalog.States == nil {
		return nil, fmt.Errorf("unsupported rollback state catalog schema")
	}
	known := make(map[string]bool, len(catalog.States))
	for _, name := range catalog.States {
		if !snapshotStateNameRE.MatchString(name) || known[name] {
			return nil, fmt.Errorf("rollback state catalog contains an invalid entry")
		}
		known[name] = true
	}
	return known, nil
}

func rebuildSnapshotStateCatalog(stateDir string) error {
	known := map[string]bool{}
	path := filepath.Join(stateDir, snapshotStateCatalogName)
	if _, exists, err := readProtectedStateFile(path); err != nil {
		return err
	} else if exists {
		loaded, err := loadSnapshotStateCatalog(stateDir)
		if err != nil {
			return err
		}
		known = loaded
	}
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !snapshotStateNameRE.MatchString(entry.Name()) {
			continue
		}
		if _, exists, err := readSnapshotState(filepath.Join(stateDir, entry.Name())); err != nil {
			return fmt.Errorf("rollback state %s is unreadable: %w", entry.Name(), err)
		} else if !exists {
			return fmt.Errorf("rollback state %s disappeared during migration", entry.Name())
		}
		known[entry.Name()] = true
	}
	return writeSnapshotStateCatalog(stateDir, known)
}

func writeSnapshotStateCatalog(stateDir string, known map[string]bool) error {
	names := make([]string, 0, len(known))
	for name := range known {
		names = append(names, name)
	}
	slices.Sort(names)
	payload, err := json.Marshal(snapshotStateCatalog{SchemaVersion: 1, States: names})
	if err != nil {
		return err
	}
	return writeProtectedStateFile(filepath.Join(stateDir, snapshotStateCatalogName), payload)
}

func readProtectedStateFile(path string) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, true, fmt.Errorf("security state is not a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, true, fmt.Errorf("security state permissions are too broad")
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- protected state under the machine home
	return payload, true, err
}

func writeProtectedStateFile(path string, payload []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".registry-state-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}

func syncDirectory(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(path) // #nosec G304 -- path is a tool-managed protected state directory
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
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
