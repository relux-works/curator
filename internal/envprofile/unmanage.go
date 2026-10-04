package envprofile

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// UnmanageRequest selects native in-place environment surfaces to return to
// operator ownership (environments §9.2).
type UnmanageRequest struct {
	Home           string
	EnvID          string
	TargetID       string
	RestoreBackups bool
	NativeHomeOf   func(string) (string, error)
	// BackupLstat and BackupReadDir are narrow seams for deterministic
	// backup-inventory failures. Production leaves them nil and uses stateread.
	BackupLstat   func(string) (stateread.Metadata, error)
	BackupReadDir func(string) (stateread.Directory, error)
}

// UnmanageHome reports the in-place work for one native adapter home.
type UnmanageHome struct {
	Environment   string
	Home          string
	Removed       []string
	Restored      []string
	BackupPath    string
	MarkerRemoved bool
}

// UnmanageResult is the report for one unmanage operation.
type UnmanageResult struct {
	Homes []UnmanageHome
}

type unmanagePlan struct {
	environment string
	home        string
	marker      *envmarker.Marker
	markerPath  string
	backupPath  string
	removals    map[string]bool
	restores    map[string]backupRestoreEntry
}

// backupRestoreEntry retains the lstat entry type separately from permission
// bits. Regular files are supported here; link restoration is a separate
// operation and must never be implemented by reading through the link.
type backupRestoreEntry struct {
	kind    fs.FileMode
	mode    fs.FileMode
	payload []byte
}

// Unmanage removes only native surfaces recorded by their environment marker
// and can restore the newest §8.3 backup generation before deleting that
// marker. It holds the manager-home mutation lock throughout, preflights all
// selected markers and backup records before its first native-home write, and
// clears the selected current-profile records last.
func Unmanage(req UnmanageRequest) (*UnmanageResult, error) {
	if req.Home == "" {
		return nil, fmt.Errorf("manager home is empty")
	}
	environment := envregistry.NormalizeEnvID(req.EnvID)
	if environment != "" {
		if _, ok := adapterByID(environment); !ok {
			return nil, fmt.Errorf("%s: explicit operand names the unregistered environment %q", DiagUnknownEnvironment, environment)
		}
	}
	if req.TargetID != "" {
		if environment != "" {
			if _, err := envregistry.TargetFor(environment, req.TargetID); err != nil {
				return nil, err
			}
		} else if _, err := envregistry.TargetByID(req.TargetID); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("secondary fixed-home target %q writes are deferred: unmanage does not touch target homes until surface writes are implemented", req.TargetID)
	}

	op, err := beginOperation(req.Home)
	if err != nil {
		return nil, err
	}
	defer func() { _ = op.close() }()

	currentRemovals, err := unmanageCurrentRemovals(req.Home, environment)
	if err != nil {
		return nil, err
	}
	plans, err := planUnmanageHomes(req)
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		if err := preflightUnmanagePlan(plan); err != nil {
			return nil, err
		}
	}

	result := &UnmanageResult{Homes: make([]UnmanageHome, 0, len(plans))}
	for _, plan := range plans {
		outcome, err := applyUnmanagePlan(plan)
		if err != nil {
			return result, fmt.Errorf("unmanage %s: %w", plan.environment, err)
		}
		result.Homes = append(result.Homes, outcome)
	}
	if len(currentRemovals) > 0 {
		if err := op.publish(nil, currentRemovals...); err != nil {
			return result, fmt.Errorf("native environment surfaces were unmanaged but the current-profile record could not be cleared: %w", err)
		}
	}
	return result, nil
}

func planUnmanageHomes(req UnmanageRequest) ([]unmanagePlan, error) {
	var selected []Adapter
	if req.EnvID != "" {
		adapter, _ := adapterByID(envregistry.NormalizeEnvID(req.EnvID))
		selected = []Adapter{adapter}
	} else {
		selected = append(selected, Adapters...)
	}

	plans := make([]unmanagePlan, 0, len(selected))
	for _, adapter := range selected {
		native := ""
		var err error
		if req.NativeHomeOf != nil {
			native, err = req.NativeHomeOf(adapter.ID)
		} else {
			native, err = NativeHome(adapter)
		}
		if err != nil {
			return nil, err
		}
		if native == "" {
			return nil, fmt.Errorf("native home for %s is empty", adapter.ID)
		}
		marker, err := envmarker.Read(native)
		if err != nil {
			return nil, err
		}
		plan := unmanagePlan{
			environment: adapter.ID,
			home:        native,
			marker:      marker,
			markerPath:  filepath.Join(native, envmarker.Name),
			backupPath:  filepath.Join(native, ".agent-environment-backup"),
			removals:    map[string]bool{},
			restores:    map[string]backupRestoreEntry{},
		}
		if marker != nil {
			for _, key := range marker.SortedSurfaceKeys() {
				for _, surfacePath := range marker.Surfaces[key].Paths {
					if credentialPath(adapter.ID, surfacePath) {
						continue
					}
					plan.removals[surfacePath] = true
				}
			}
		}
		if req.RestoreBackups {
			files, err := newestBackupFiles(req, adapter.ID, plan.backupPath)
			if err != nil {
				return nil, err
			}
			plan.restores = files
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func newestBackupFiles(req UnmanageRequest, environment, backupRoot string) (map[string]backupRestoreEntry, error) {
	metadata, err := readBackupLstat(req, backupRoot)
	if err != nil {
		return nil, backupRecordUnreadable(backupRoot, err)
	}
	if metadata.Kind == stateread.KindAbsent {
		return map[string]backupRestoreEntry{}, nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil || !metadata.Info.IsDir() || metadata.Info.Mode()&os.ModeSymlink != 0 {
		return nil, backupRecordUnreadable(backupRoot, stateread.UnusableError(backupRoot, fmt.Errorf("backup inventory is not a directory")))
	}
	listing, err := readBackupDir(req, backupRoot)
	if err != nil {
		return nil, backupRecordUnreadable(backupRoot, err)
	}
	if listing.Kind != stateread.KindPresent {
		return nil, backupRecordUnreadable(backupRoot, stateread.UnusableError(backupRoot, fmt.Errorf("unknown directory read state %q", listing.Kind)))
	}

	newest := 0
	for _, entry := range listing.Entries {
		generation, err := strconv.Atoi(entry.Name())
		if err != nil || generation <= 0 || strconv.Itoa(generation) != entry.Name() {
			continue
		}
		if generation > newest {
			newest = generation
		}
	}
	if newest == 0 {
		return map[string]backupRestoreEntry{}, nil
	}

	generationRoot := filepath.Join(backupRoot, strconv.Itoa(newest))
	generationMetadata, err := stateread.Lstat(generationRoot)
	if err != nil {
		return nil, backupRecordUnreadable(generationRoot, err)
	}
	if generationMetadata.Kind != stateread.KindPresent || generationMetadata.Info == nil || !generationMetadata.Info.IsDir() || generationMetadata.Info.Mode()&os.ModeSymlink != 0 {
		return nil, backupRecordUnreadable(generationRoot, stateread.UnusableError(generationRoot, fmt.Errorf("newest backup generation is not a directory")))
	}
	files, err := readBackupTree(req, environment, generationRoot, "")
	if err != nil {
		return nil, err
	}
	return files, nil
}

func readBackupTree(req UnmanageRequest, environment, directory, prefix string) (map[string]backupRestoreEntry, error) {
	listing, err := readBackupDir(req, directory)
	if err != nil {
		return nil, backupRecordUnreadable(directory, err)
	}
	if listing.Kind != stateread.KindPresent {
		return nil, backupRecordUnreadable(directory, stateread.UnusableError(directory, fmt.Errorf("backup generation directory is absent or unusable")))
	}
	files := map[string]backupRestoreEntry{}
	for _, entry := range listing.Entries {
		name := entry.Name()
		rel := path.Join(prefix, name)
		if !safeBackupPath(rel) || reservedRestorePath(rel) {
			return nil, backupRecordUnreadable(filepath.Join(directory, name), stateread.UnusableError(filepath.Join(directory, name), fmt.Errorf("invalid backup path %q", rel)))
		}
		full := filepath.Join(directory, name)
		metadata, err := stateread.Lstat(full)
		if err != nil {
			return nil, backupRecordUnreadable(full, err)
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return nil, backupRecordUnreadable(full, stateread.UnusableError(full, fmt.Errorf("backup entry cannot be established")))
		}
		if metadata.Info.IsDir() {
			nested, err := readBackupTree(req, environment, full, rel)
			if err != nil {
				return nil, err
			}
			for nestedPath, payload := range nested {
				files[nestedPath] = payload
			}
			continue
		}
		if !metadata.Info.Mode().IsRegular() {
			return nil, backupRecordUnreadable(full, stateread.UnusableError(full, fmt.Errorf("backup entry is not a regular file")))
		}
		if credentialPath(environment, rel) {
			return nil, backupRecordUnreadable(full, stateread.UnusableError(full, fmt.Errorf("backup record names a credential path")))
		}
		state, err := stateread.ReadFile(full)
		if err != nil {
			return nil, backupRecordUnreadable(full, err)
		}
		if state.Kind != stateread.KindPresent {
			return nil, backupRecordUnreadable(full, stateread.UnusableError(full, fmt.Errorf("backup file is absent")))
		}
		files[rel] = backupRestoreEntry{
			kind:    metadata.Info.Mode().Type(),
			mode:    metadata.Info.Mode().Perm(),
			payload: append([]byte(nil), state.Bytes...),
		}
	}
	return files, nil
}

func readBackupDir(req UnmanageRequest, directory string) (stateread.Directory, error) {
	if req.BackupReadDir != nil {
		return req.BackupReadDir(directory)
	}
	return stateread.ReadDir(directory)
}

func readBackupLstat(req UnmanageRequest, path string) (stateread.Metadata, error) {
	if req.BackupLstat != nil {
		return req.BackupLstat(path)
	}
	return stateread.Lstat(path)
}

func backupRecordUnreadable(path string, err error) error {
	if err == nil {
		err = stateread.UnusableError(path, fmt.Errorf("backup inventory cannot be established"))
	}
	return fmt.Errorf("%s: %w", envregistry.DiagBackupRecordUnreadable, err)
}

func preflightUnmanagePlan(plan unmanagePlan) error {
	paths := make(map[string]bool, len(plan.removals)+len(plan.restores))
	for rel := range plan.removals {
		paths[rel] = true
	}
	for rel, entry := range plan.restores {
		if !safeBackupPath(rel) || reservedRestorePath(rel) {
			return backupRecordUnreadable(filepath.Join(plan.backupPath, rel), stateread.UnusableError(filepath.Join(plan.backupPath, rel), fmt.Errorf("invalid backup path %q", rel)))
		}
		if !entry.kind.IsRegular() {
			return backupRecordUnreadable(filepath.Join(plan.backupPath, rel), fmt.Errorf("unsupported backup entry type %v", entry.kind))
		}
		paths[rel] = true
	}
	for rel := range paths {
		if err := inspectUnmanagePath(plan.home, rel); err != nil {
			return fmt.Errorf("%s: %w", plan.environment, err)
		}
	}
	return nil
}

func inspectUnmanagePath(home, rel string) error {
	if !safeBackupPath(rel) {
		return fmt.Errorf("%s: invalid home-relative path %q", DiagSurfaceUnreadable, rel)
	}
	parts := strings.Split(rel, "/")
	current := home
	for index, part := range parts {
		current = filepath.Join(current, filepath.FromSlash(part))
		metadata, err := stateread.Lstat(current)
		if err != nil {
			return err
		}
		if metadata.Kind == stateread.KindAbsent {
			return nil
		}
		if metadata.Kind != stateread.KindPresent || metadata.Info == nil {
			return stateread.UnusableError(current, fmt.Errorf("unknown entry state %q", metadata.Kind))
		}
		last := index == len(parts)-1
		if !last && metadata.Info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: refusing to traverse link at %s", DiagWriteWouldFollowLink, current)
		}
		if !last && !metadata.Info.IsDir() {
			return stateread.UnusableError(current, fmt.Errorf("parent path component is not a directory"))
		}
		if last && !metadata.Info.Mode().IsRegular() && metadata.Info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("%s: recorded surface is not a regular file or link: %s", DiagSurfaceUnreadable, current)
		}
	}
	return nil
}

func applyUnmanagePlan(plan unmanagePlan) (UnmanageHome, error) {
	outcome := UnmanageHome{
		Environment: plan.environment,
		Home:        plan.home,
		BackupPath:  plan.backupPath,
	}
	paths := make([]string, 0, len(plan.removals)+len(plan.restores))
	seen := map[string]bool{}
	for rel := range plan.removals {
		seen[rel] = true
	}
	for rel := range plan.restores {
		seen[rel] = true
	}
	for rel := range seen {
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	for _, rel := range paths {
		full := filepath.Join(plan.home, filepath.FromSlash(rel))
		entry, restore := plan.restores[rel]
		if restore {
			// Stage privately, preserve the saved permission bits, and replace
			// the entry without opening the current managed link's target.
			// On Windows the writer creates an owner-only protected DACL;
			// FileMode does not encode or restore the original Windows ACL.
			if err := atomicManagedFile(plan.home, rel, entry.payload, entry.mode); err != nil {
				return outcome, err
			}
			outcome.Restored = append(outcome.Restored, rel)
			continue
		}
		if !plan.removals[rel] {
			continue
		}
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			return outcome, err
		}
		outcome.Removed = append(outcome.Removed, rel)
	}
	if plan.marker != nil {
		if err := os.Remove(plan.markerPath); err != nil && !os.IsNotExist(err) {
			return outcome, err
		}
		outcome.MarkerRemoved = true
	}
	return outcome, nil
}

func safeBackupPath(rel string) bool {
	if rel == "" || path.IsAbs(rel) || !fs.ValidPath(rel) {
		return false
	}
	clean := path.Clean(rel)
	return clean == rel && clean != "." && !strings.HasPrefix(clean, "../")
}

func reservedRestorePath(rel string) bool {
	return rel == envmarker.Name || rel == ".agent-environment-backup" || strings.HasPrefix(rel, ".agent-environment-backup/")
}

func credentialPath(environment, rel string) bool {
	adapter, err := envregistry.ByID(environment)
	if err != nil {
		return false
	}
	for _, passthrough := range adapter.PassthroughFor(runtime.GOOS) {
		if filepath.ToSlash(filepath.Clean(passthrough.Path)) == rel {
			return true
		}
	}
	return false
}

func unmanageCurrentRemovals(home, environment string) ([]string, error) {
	current, err := Current(home)
	if err != nil {
		return nil, err
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		return nil, err
	}
	var scopes []string
	removeMachine := environment == ""
	if environment == "" {
		for scope := range scoped {
			scopes = append(scopes, scope)
		}
	} else {
		scope := "env:" + environment
		if _, exists := scoped[scope]; exists {
			scopes = append(scopes, scope)
		}
	}
	sort.Strings(scopes)
	var removals []string
	if removeMachine && current != "" {
		removals = append(removals, CurrentFile(home))
	}
	for _, scope := range scopes {
		paths, err := scopeRecordPaths(home, scope)
		if err != nil {
			return nil, err
		}
		removals = append(removals, paths...)
	}
	sort.Strings(removals)
	return removals, nil
}
