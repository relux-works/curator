package install

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/staging"
)

const projectBinLedgerName = ".curator-bin-ownership.json"
const projectBinConflict = "project_bin_unmanaged_conflict"
const projectBinPreserved = "project_bin_unmanaged_preserved"

type projectBinLedger struct {
	SchemaVersion int               `json:"schema_version"`
	Entries       map[string]string `json:"entries"`
}

type projectBinOwnership struct {
	owned      map[string]bool
	preimages  map[string]string
	messages   []string
	ledgerPath string
}

// Ownership is a previous journaled fingerprint plus an exact live match.
// Before the first ledger, legacy adoption requires an installed marker naming
// this command and a byte-exact old launcher to that marker's runtime/cache.
// A changed live entry never inherits ownership from its name or old ledger.
func inspectProjectBin(home, binDir, platform, externalRoot string, desired map[string]bool) (projectBinOwnership, error) {
	state := projectBinOwnership{owned: map[string]bool{}, preimages: map[string]string{}, ledgerPath: filepath.Join(filepath.Dir(binDir), projectBinLedgerName)}
	desiredKeys := map[string]bool{}
	for path := range desired {
		state.preimages[path] = ""
		desiredKeys[projectBinKey(path, platform)] = true
	}
	var ledger projectBinLedger
	legacy := false
	info, err := os.Lstat(state.ledgerPath)
	if os.IsNotExist(err) {
		legacy = true
	} else if err != nil {
		return state, err
	} else {
		if !info.Mode().IsRegular() || info.Size() > 1<<20 {
			return state, fmt.Errorf("project_bin_ownership_invalid: %s", state.ledgerPath)
		}
		file, err := os.Open(state.ledgerPath)
		if err != nil {
			return state, err
		}
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&ledger)
		tail := decoder.Decode(&struct{}{})
		_ = file.Close()
		if err != nil || tail != io.EOF || ledger.SchemaVersion != 1 || ledger.Entries == nil {
			return state, fmt.Errorf("project_bin_ownership_invalid: %s", state.ledgerPath)
		}
	}
	if info, err := os.Lstat(binDir); err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
		return state, fmt.Errorf("project_bin_ownership_invalid: %s", binDir)
	} else if err != nil && !os.IsNotExist(err) {
		return state, err
	}
	entries, err := os.ReadDir(binDir)
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	var previous map[string][]*marker.Marker
	if legacy {
		previous, err = previousProjectCommands(home, binDir)
		if err != nil {
			return state, err
		}
	}
	for _, entry := range entries {
		path := filepath.Join(binDir, entry.Name())
		fingerprint, err := shimFingerprint(path)
		if err != nil {
			return state, err
		}
		owned := fingerprint != "" && ledger.Entries[entry.Name()] == fingerprint
		if legacy {
			name := entry.Name()
			if platform == "windows" {
				name = strings.TrimSuffix(name, ".cmd")
			}
			for _, recorded := range previous[name] {
				if runtimestore.MatchesLegacyShim(path, platform, func(target string) bool { return legacyTargetOwned(target, home, externalRoot, name, recorded) }) {
					owned = true
					break
				}
			}
		}
		if owned {
			state.preimages[path] = fingerprint
			state.owned[projectBinKey(path, platform)] = true
			continue
		}
		if desiredKeys[projectBinKey(path, platform)] {
			return state, fmt.Errorf("%s: %s", projectBinConflict, path)
		}
		state.messages = append(state.messages, fmt.Sprintf("%s: %s", projectBinPreserved, path))
	}
	return state, nil
}

func previousProjectCommands(home, binDir string) (map[string][]*marker.Marker, error) {
	commands := map[string][]*marker.Marker{}
	for _, store := range []string{filepath.Join(filepath.Dir(binDir), "skills"), scopes.HybridSkillsRoot(home)} {
		entries, err := os.ReadDir(store)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() || !identifiers.Valid(entry.Name()) {
				continue
			}
			recorded, _, err := marker.ReadState(filepath.Join(store, entry.Name()))
			if err != nil || recorded == nil || recorded.Name != entry.Name() {
				continue
			}
			for _, name := range recorded.Commands {
				commands[name] = append(commands[name], recorded)
			}
		}
	}
	return commands, nil
}

func legacyTargetOwned(target, home, externalRoot, command string, recorded *marker.Marker) bool {
	if build, ok := recorded.Builds[command]; ok {
		key := strings.TrimPrefix(string(build.CacheKey), "sha256:")
		if build.Driver == "go-repository-v1" {
			return target == filepath.Join(externalRoot, buildrepo.ArtifactsDir(build.ReceiptSchemaVersion), key, "artifact") || target == filepath.Join(externalRoot, buildrepo.ArtifactsDir(build.ReceiptSchemaVersion), key, "artifact.exe")
		}
		namespace := buildcache.Namespace
		if recorded.Package != nil {
			namespace = buildcache.SourceAwareNamespace
		}
		return target == filepath.Join(home, "cache", "build", namespace, key, filepath.FromSlash(build.ArtifactPath))
	}
	key := recorded.Commit
	if recorded.Package != nil {
		digest, err := recorded.Package.Digest()
		if err != nil {
			return false
		}
		key, err = runtimestore.SourceV1Key(digest)
		if err != nil {
			return false
		}
	}
	root := runtimestore.Dir(home, recorded.Name, key)
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	if len(recorded.RuntimeRoots) == 0 {
		return rel == filepath.Join("bin", command) || rel == filepath.Join("bin", command+".cmd")
	}
	for _, runtimeRoot := range recorded.RuntimeRoots {
		prefix := filepath.FromSlash(runtimeRoot)
		if rel == prefix || strings.HasPrefix(rel, prefix+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func shimFingerprint(path string) (string, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		hash := sha256.Sum256([]byte(target))
		return "link:" + hex.EncodeToString(hash[:]), nil
	}
	if !info.Mode().IsRegular() {
		return "", nil
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return "", fmt.Errorf("%s: %s", projectBinConflict, path)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	current, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !current.Mode().IsRegular() || !os.SameFile(opened, current) {
		return "", fmt.Errorf("%s: %s", projectBinConflict, path)
	}
	return "file:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func (state projectBinOwnership) stageLedger(stageRoot string, plan *staging.Plan) error {
	entries := map[string]string{}
	for _, target := range plan.Targets {
		if target.Class != staging.ClassCanonicalShim || target.Removal() {
			continue
		}
		fingerprint, err := shimFingerprint(target.StagedPath)
		if err != nil {
			return err
		}
		if fingerprint == "" {
			return fmt.Errorf("project_bin_ownership_invalid: staged shim is not a file")
		}
		entries[filepath.Base(target.LivePath)] = fingerprint
	}
	payload, err := json.Marshal(projectBinLedger{SchemaVersion: 1, Entries: entries})
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if old, err := os.ReadFile(state.ledgerPath); err == nil && bytes.Equal(old, payload) {
		return nil
	}
	staged := filepath.Join(stageRoot, "project-bin-ledger.json")
	if err := os.WriteFile(staged, payload, 0o600); err != nil {
		return err
	}
	plan.Replace(staging.ClassMirrorLedger, "project-bin-ownership", state.ledgerPath, staged)
	return nil
}

func projectBinKey(path, platform string) string {
	if platform == "windows" {
		return strings.ToLower(path)
	}
	return path
}
func isShimLink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}
