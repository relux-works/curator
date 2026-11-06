package install

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/scriptpolicy"
	"github.com/relux-works/curator/internal/scriptworker"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
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
// this command and either an exact old script or a marker-derived enforced sidecar.
// A changed live entry never inherits ownership from its name or old ledger.
func inspectProjectBin(home, binDir, platform, externalRoot string, desired map[string]bool) (state projectBinOwnership, resultErr error) {
	defer func() {
		var stateErr *stateread.Error
		if errors.As(resultErr, &stateErr) {
			resultErr = fmt.Errorf("%w; restore read access to this manager state before reinstalling", resultErr)
		}
	}()
	state = projectBinOwnership{owned: map[string]bool{}, preimages: map[string]string{}, ledgerPath: filepath.Join(filepath.Dir(binDir), projectBinLedgerName)}
	desiredKeys := map[string]bool{}
	for path := range desired {
		state.preimages[path] = ""
		desiredKeys[projectBinKey(path, platform)] = true
	}
	var ledger projectBinLedger
	legacy := false
	metadata, err := stateread.Lstat(state.ledgerPath)
	if err != nil {
		return state, err
	}
	if metadata.Kind == stateread.KindAbsent {
		legacy = true
	} else {
		info := metadata.Info
		if !info.Mode().IsRegular() || info.Size() > 1<<20 {
			return state, projectBinRefusal("project_bin_ownership_invalid", state.ledgerPath)
		}
		file, err := stateread.ReadFile(state.ledgerPath)
		if err != nil {
			return state, err
		}
		decoder := json.NewDecoder(bytes.NewReader(file.Bytes))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&ledger)
		tail := decoder.Decode(&struct{}{})
		if err != nil || tail != io.EOF || ledger.SchemaVersion != 1 || ledger.Entries == nil {
			return state, projectBinRefusal("project_bin_ownership_invalid", state.ledgerPath)
		}
	}
	metadata, err = stateread.Lstat(binDir)
	if err != nil {
		return state, err
	}
	if metadata.Kind == stateread.KindPresent && (!metadata.Info.IsDir() || metadata.Info.Mode()&os.ModeSymlink != 0) {
		return state, projectBinRefusal("project_bin_ownership_invalid", binDir)
	}
	listing, err := stateread.ReadDir(binDir)
	if err != nil {
		return state, err
	}
	if listing.Kind == stateread.KindAbsent {
		return state, nil
	}
	entries := listing.Entries
	var previous map[string][]*marker.Marker
	if legacy {
		previous, err = previousProjectCommands(home, binDir)
		if err != nil {
			return state, err
		}
	}
	// Legacy enforced ownership derives the complete sidecar from the old marker's
	// frozen manifest. The old manager executable bytes are deliberately not proof.
	live := map[string]bool{}
	for _, entry := range entries {
		live[projectBinKey(filepath.Join(binDir, entry.Name()), platform)] = true
	}
	enforced := map[string]bool{}
	legacyEnforced := map[string]bool{}
	if legacy {
		for name, records := range previous {
			destination, err := runtimestore.NewManagedEnforcedShim(runtimestore.ProjectShim, binDir, name, platform)
			if err != nil {
				return state, err
			}
			// A stale marker with no live launcher or sidecar needs no frozen
			// manifest. Only live ownership candidates may require that load.
			if !live[projectBinKey(destination.Path(), platform)] && !live[projectBinKey(destination.SidecarPath(), platform)] {
				continue
			}
			for _, recorded := range records {
				expected, candidate, err := legacyEnforcedSidecar(home, binDir, platform, name, recorded)
				if err != nil {
					// Decoding frozen ownership evidence is not repaired by
					// changing permissions; manual relocation is sufficient.
					return state, projectBinRefusal("project_bin_legacy_enforced", destination.Path())
				}
				if !candidate {
					continue
				}
				legacyEnforced[destination.Path()] = true
				legacyEnforced[destination.SidecarPath()] = true
				sidecar, err := stateread.ReadRegularFile(destination.SidecarPath())
				if err != nil {
					return state, err
				}
				if sidecar.Kind == stateread.KindPresent && bytes.Equal(sidecar.Bytes, expected) {
					enforced[destination.Path()] = true
					enforced[destination.SidecarPath()] = true
				}
			}
		}
	}
	for _, entry := range entries {
		path := filepath.Join(binDir, entry.Name())
		key := projectBinKey(path, platform)
		name := entry.Name()
		if platform == "windows" {
			name = strings.TrimSuffix(name, ".cmd")
		}
		relevant := desiredKeys[key] || ledger.Entries[entry.Name()] != "" || len(previous[name]) != 0 || legacyEnforced[path] || (legacy && (strings.HasSuffix(path, scriptworker.ShimSidecarSuffix) || hasEnforcedSidecar(path)))
		if !relevant {
			state.messages = append(state.messages, projectBinRefusal(projectBinPreserved, path).Error())
			continue // An unrelated entry's bytes, permissions and size are never read.
		}
		fingerprint, err := shimFingerprint(path)
		if err != nil {
			return state, err
		}
		owned := fingerprint != "" && ledger.Entries[entry.Name()] == fingerprint
		if legacy && fingerprint != "" {
			if enforced[path] && entry.Type().IsRegular() {
				owned = true
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
			state.owned[key] = true
			continue
		}
		if legacy && (legacyEnforced[path] || strings.HasSuffix(path, scriptworker.ShimSidecarSuffix) || hasEnforcedSidecar(path)) {
			return state, projectBinRefusal("project_bin_legacy_enforced", path)
		}
		if desiredKeys[key] {
			return state, projectBinRefusal(projectBinConflict, path)
		}
		state.messages = append(state.messages, projectBinRefusal(projectBinPreserved, path).Error())
	}

	return state, nil
}

func previousProjectCommands(home, binDir string) (map[string][]*marker.Marker, error) {
	commands := map[string][]*marker.Marker{}
	for _, store := range []string{filepath.Join(filepath.Dir(binDir), "skills"), scopes.HybridSkillsRoot(home)} {
		listing, err := stateread.ReadDir(store)
		if err != nil {
			return nil, err
		}
		if listing.Kind == stateread.KindAbsent {
			continue
		}
		for _, entry := range listing.Entries {
			if !entry.IsDir() || !identifiers.Valid(entry.Name()) {
				continue
			}
			recorded, _, err := marker.ReadState(filepath.Join(store, entry.Name()))
			if err != nil {
				var invalid *marker.InvalidError
				if !errors.As(err, &invalid) {
					return nil, err
				}
				continue
			}
			if recorded == nil || recorded.Name != entry.Name() {
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
	metadata, err := stateread.Lstat(path)
	if err != nil {
		return "", err
	}
	if metadata.Kind == stateread.KindAbsent {
		return "", nil
	}
	info := metadata.Info
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", stateread.UnusableError(path, err)
		}
		hash := sha256.Sum256([]byte(target))
		return "link:" + hex.EncodeToString(hash[:]), nil
	}
	if !info.Mode().IsRegular() {
		return "", nil
	}
	file, err := os.Open(path) // #nosec G304 -- read-only ownership probe of an enumerated bin entry; type and opened identity are checked.
	if err != nil {
		return "", stateread.UnusableError(path, err)
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil {
		return "", stateread.UnusableError(path, err)
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return "", projectBinRefusal(projectBinConflict, path)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", stateread.UnusableError(path, err)
	}
	current, err := os.Lstat(path)
	if err != nil {
		return "", stateread.UnusableError(path, err)
	}
	if !current.Mode().IsRegular() || !os.SameFile(opened, current) {
		return "", projectBinRefusal(projectBinConflict, path)
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
			return projectBinRefusal("project_bin_ownership_invalid", target.StagedPath)
		}
		entries[filepath.Base(target.LivePath)] = fingerprint
	}
	payload, err := json.Marshal(projectBinLedger{SchemaVersion: 1, Entries: entries})
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	old, err := stateread.ReadFile(state.ledgerPath)
	if err != nil {
		return err
	}
	if old.Kind == stateread.KindPresent && bytes.Equal(old.Bytes, payload) {
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

// Each diagnostic carries one action; no refusal authorizes automatic recovery.
func projectBinRefusal(code, path string) error {
	remedy := "move this foreign entry before reinstalling"
	switch code {
	case projectBinPreserved:
		remedy = "leave this foreign entry in place or move it manually"
	case "project_bin_ownership_invalid":
		remedy = "restore the ownership ledger from a trusted backup before reinstalling"
	case "project_bin_legacy_enforced":
		remedy = "move the legacy launcher and its sidecar aside before reinstalling"
	}
	return fmt.Errorf("%s: %s; %s", code, path, remedy)
}

func hasEnforcedSidecar(path string) bool {
	metadata, err := stateread.Lstat(path + scriptworker.ShimSidecarSuffix)
	return err == nil && metadata.Kind == stateread.KindPresent
}

// The marker selects a frozen manifest, whose enforced declaration and sidecar
// derivation supply legacy evidence. Missing evidence never authorizes adoption.
func legacyEnforcedSidecar(home, binDir, platform, name string, recorded *marker.Marker) ([]byte, bool, error) {
	key, source, commit, directory := recorded.Commit, recorded.Source, recorded.Commit, "."
	if recorded.Git != "" {
		source = identity.Canonical(recorded.Git)
	}
	var roots []string
	if recorded.Package != nil {
		pkg := recorded.Package
		digest, err := pkg.Digest()
		if err != nil {
			return nil, false, err
		}
		key, err = runtimestore.SourceV1Key(digest)
		if err != nil {
			return nil, false, err
		}
		if pkg.Kind == "local-snapshot" {
			roots = []string{filepath.Join(snapshot.LocalStoreDir(home), strings.TrimPrefix(pkg.Snapshot, "sha256:"), "snapshot")}
		} else if pkg.Commit != nil {
			commit, directory = pkg.Commit.Hex, pkg.Directory
			roots = []string{snapshot.Dir(home, pkg.Source, commit), snapshot.Dir(home, pkg.Repository, commit), snapshot.Dir(home, recorded.Name, commit)}
		}
	} else {
		roots = []string{snapshot.Dir(home, source, commit)}
	}
	for _, root := range roots {
		if directory != "" && directory != "." {
			root = filepath.Join(root, filepath.FromSlash(directory))
		}
		metadata, err := stateread.Lstat(root)
		if err != nil {
			return nil, false, err
		}
		if metadata.Kind == stateread.KindAbsent {
			continue
		}
		spec, err := skillspec.Load(root)
		if err != nil {
			return nil, false, stateread.UnusableError(root, err)
		}
		command, ok := spec.Commands[name]
		if !ok || !scriptpolicy.Enforced(command) {
			continue
		}
		capabilities, err := rawDeclaredCapabilities(root)
		if err != nil {
			return nil, true, err
		}
		runtimeDir := runtimestore.Dir(home, recorded.Name, key)
		entry := command.UnixPath
		if platform == "windows" {
			entry = command.WinPath
		}
		sidecar, err := scriptworker.NewShimSidecar(recorded.Name, name, command.Interpreter, filepath.Join(runtimeDir, filepath.FromSlash(entry)), runtimeDir, filepath.Dir(filepath.Dir(binDir)), recorded.SkillSchemaVersion, capabilities)
		if err != nil {
			return nil, true, err
		}
		payload, err := sidecar.Marshal()
		return payload, true, err
	}
	return nil, false, nil
}
