package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/skillruntime"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// GlobalSkillTransaction joins profile publication and the legacy continuation
// under the caller's manager-home lock. Ordinary returned errors restore the
// profile's mutable records and surfaces after verifying their postimages.
// Failed native compensation retains published profile recovery records and
// returns GlobalCompensationError. The §9.4 unmanaged-conflict exception
// retains the extended lock and its install record. Immutable admitted store
// entries may remain.
// The continuation owns its own install journal and rolls its targets back
// before returning an error. This compensation does not replace crash recovery
// of individual profile and install transactions.
func GlobalSkillTransaction(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error), lock transaction.HomeLock, next func() error) (info Info, resultErr error) {
	if err := refusePreparedSkillPlanAdoption(home, ""); err != nil {
		return Info{}, err
	}
	if err := skillruntime.RefuseUnadmitted(home); err != nil {
		return Info{}, err
	}
	if err := lock.AssertHeld(); err != nil {
		return Info{}, err
	}
	if err := envauthority.Validate(home); err != nil {
		return Info{}, err
	}
	if addition != nil && (addition.Name == "" || addition.Git == "" || (addition.Tag == "") == (addition.Revision == "")) {
		return Info{}, fmt.Errorf("%s: global add requires a skill name, Git source, and exactly one of tag or revision", DiagRefConflict)
	}
	engine, err := transaction.New(home)
	if err != nil {
		return Info{}, err
	}
	op := &operation{lock: borrowedGlobalLock{lock}, engine: engine, undo: &globalUndo{entries: map[string]*globalImage{}, managerHome: home}}
	var retainedRecords map[string][]byte
	defer func() {
		if resultErr == nil {
			return
		}
		// Only a profile-phase in-place unmanaged conflict qualifies for retaining
		// the extended lock. A later legacy refusal never inherits this exception.
		if err := op.undo.restore(); err != nil {
			resultErr = errors.Join(resultErr, &GlobalCompensationError{Cause: err})
			return
		}
		if retainedRecords != nil {
			op.undo = nil
			if err := op.publish(retainedRecords); err != nil {
				resultErr = errors.Join(resultErr, err)
			}
		}
	}()
	info, err = globalSkillOperationWithOperation(home, policy, machine, addition, nativeHomeOf, op)
	if err != nil {
		var conflict *globalInPlaceConflict
		if errors.As(err, &conflict) && info.Lock != nil {
			var marshalErr error
			var retainedLock []byte
			retainedLock, marshalErr = info.Lock.Canonical()
			if marshalErr != nil {
				return info, errors.Join(err, marshalErr)
			}
			// First-use default provisioning must retain the install record that
			// makes the pending lock usable by profile sync --takeover.
			source, present, readErr := readManagedRegular(ProfileDir(home, info.Name), "source.json")
			if readErr != nil {
				return info, errors.Join(err, readErr)
			}
			if !present {
				return info, errors.Join(err, fmt.Errorf("pending profile install record disappeared"))
			}
			retainedRecords = map[string][]byte{lockPath(home, info.Name): retainedLock, sourcePath(home, info.Name): source}
		}
		return info, err
	}
	if addition != nil {
		root := filepath.Join(home, "global")
		if err := op.capture(manifest.PathIn(root)); err != nil {
			return info, err
		}
		kind, value := "tag", addition.Tag
		if addition.Revision != "" {
			kind, value = "revision", addition.Revision
		}
		if err := manifest.AddDecl(root, addition.Name, kind, value, addition.Git, ""); err != nil {
			return info, err
		}
		if err := op.recordPostimages(manifest.PathIn(root)); err != nil {
			return info, err
		}
	}
	return info, next()
}

type borrowedGlobalLock struct{ transaction.HomeLock }

func (borrowedGlobalLock) Close() error { return nil }
func (lock borrowedGlobalLock) AssertCurrent() error {
	if current, ok := lock.HomeLock.(interface{ AssertCurrent() error }); ok {
		return current.AssertCurrent()
	}
	return fmt.Errorf("global operation requires current manager lock")
}

// Images cover only the profile records and paths an admitted materialization
// can write. Native homes and credential stores are never walked wholesale;
// symbolic links are recorded as entries, never followed. Capturing backup
// generations also compensates retention pruning on an unsuccessful attempt.
type globalUndo struct {
	managerHome      string
	entries          map[string]*globalImage
	absentParents    map[string]bool
	parentModes      map[string]os.FileMode
	postimages       map[string]*globalImage
	parentPostimages map[string]*globalImage
}

// GlobalCompensationError forbids retrying a failed paired publication. The
// published profile lock and install record remain available for explicit sync
// after the conflicting or unwritable surface has been resolved.
type GlobalCompensationError struct{ Cause error }

func (err *GlobalCompensationError) Error() string {
	return fmt.Sprintf("global profile compensation failed; pending profile lock retained: %v", err.Cause)
}
func (err *GlobalCompensationError) Unwrap() error { return err.Cause }

type globalImage struct {
	mode     os.FileMode
	owner    pathboundary.OwnerIdentity
	data     []byte
	link     string
	children map[string]*globalImage
}

func readGlobalImage(path string) (*globalImage, error) {
	state, err := stateread.Lstat(path)
	if err != nil {
		return nil, err
	}
	if state.Kind == stateread.KindAbsent {
		return nil, nil
	}
	if state.Kind != stateread.KindPresent || state.Info == nil {
		return nil, stateread.UnusableError(path, fmt.Errorf("profile surface metadata unreadable"))
	}
	info := state.Info
	owner, err := pathboundary.DefaultOwnerLookup()(path, info)
	if err != nil {
		return nil, err
	}
	image := &globalImage{mode: info.Mode(), owner: owner}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		var link stateread.Link
		link, err = stateread.Readlink(path)
		if err == nil && link.Kind != stateread.KindPresent {
			err = stateread.UnusableError(path, fmt.Errorf("profile surface link changed during capture"))
		}
		image.link = link.Target
	case info.IsDir():
		image.children = map[string]*globalImage{}
		var listing stateread.Directory
		listing, err = stateread.ReadDir(path)
		if err == nil && listing.Kind != stateread.KindPresent {
			err = stateread.UnusableError(path, fmt.Errorf("profile surface directory changed during capture"))
		}
		if err == nil {
			for _, entry := range listing.Entries {
				var child *globalImage
				child, err = readGlobalImage(filepath.Join(path, entry.Name()))
				if err != nil {
					break
				}
				image.children[entry.Name()] = child
			}
		}
	case info.Mode().IsRegular():
		var present bool
		image.data, present, err = readManagedRegular(filepath.Dir(path), filepath.Base(path))
		if err == nil && !present {
			err = fmt.Errorf("profile surface disappeared during capture: %s", path)
		}
	default:
		err = fmt.Errorf("cannot journal non-file profile surface %s", path)
	}
	return image, err
}
func (op *operation) capture(paths ...string) error {
	if op == nil || op.undo == nil {
		return nil
	}
	for _, path := range paths {
		path = filepath.Clean(path)
		covered := false
		for prior := range op.undo.entries {
			if path == prior || strings.HasPrefix(path, prior+string(filepath.Separator)) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		if op.undo.managerHome != "" {
			rel, err := filepath.Rel(op.undo.managerHome, path)
			if err != nil {
				return err
			}
			if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				if _, err := managedPath(op.undo.managerHome, filepath.ToSlash(rel), false); err != nil {
					return err
				}
			}
		}
		image, err := readGlobalImage(path)
		if err != nil {
			return err
		}
		op.undo.entries[path] = image
		if op.undo.postimages == nil {
			op.undo.postimages = map[string]*globalImage{}
		}
		op.undo.postimages[path] = image
		if op.undo.absentParents == nil {
			op.undo.absentParents = map[string]bool{}
		}
		if err := op.noteParents(path); err != nil {
			return err
		}
	}
	return nil
}

// Record immediately at the publisher boundary, before another phase can run
// external processes or the continuation. Never learn a postimage at rollback.
func (op *operation) recordPostimages(paths ...string) error {
	if op == nil || op.undo == nil {
		return nil
	}
	for captured := range op.undo.entries {
		for _, path := range paths {
			if captured == path || strings.HasPrefix(path, captured+string(filepath.Separator)) {
				image, err := readGlobalImage(captured)
				if err != nil {
					return err
				}
				op.undo.postimages[captured] = image
				break
			}
		}
	}
	for parent := range op.undo.parentPostimages {
		for _, path := range paths {
			if path == parent || strings.HasPrefix(path, parent+string(filepath.Separator)) {
				image, err := readGlobalParent(parent)
				if err != nil {
					return err
				}
				op.undo.parentPostimages[parent] = image
				break
			}
		}
	}
	return nil
}

func (op *operation) recordHomePostimages(root string) error {
	if op == nil || op.undo == nil {
		return nil
	}
	var paths []string
	for path := range op.undo.entries {
		if strings.HasPrefix(path, root+string(filepath.Separator)) {
			paths = append(paths, path)
		}
	}
	// Managed-home provisioning can change empty directory modes without a
	// captured leaf below them (notably Muse data, state and cache homes).
	for parent := range op.undo.parentPostimages {
		if parent == root || strings.HasPrefix(parent, root+string(filepath.Separator)) {
			paths = append(paths, parent)
		}
	}
	return op.recordPostimages(paths...)
}

func (undo *globalUndo) checkPostimage(path string) error {
	current, err := readGlobalImage(path)
	if err != nil {
		return err
	}
	// An unchanged or already restored preimage needs no mutation.
	if reflect.DeepEqual(current, undo.entries[path]) {
		return nil
	}
	if !reflect.DeepEqual(current, undo.postimages[path]) {
		return fmt.Errorf("%w: global profile compensation postimage changed: %s", transaction.ErrImplementationCorruption, path)
	}
	return nil
}

// Chmod accepts permission and special bits, not the entry-type bits.
func globalPermissionMode(mode os.FileMode) os.FileMode {
	return mode & (os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky)
}

func restoreGlobalImage(path string, image *globalImage) error {
	root, rel := filepath.Dir(path), filepath.Base(path)
	current, err := readGlobalImage(path)
	if err != nil {
		return err
	}
	if reflect.DeepEqual(current, image) {
		return nil
	}
	if image == nil {
		return removeManagedTree(root, rel)
	}
	if image.mode.IsDir() {
		if err := removeManagedTree(root, rel); err != nil {
			return err
		}
		if err := os.Mkdir(path, 0700); err != nil {
			return err
		}
		for name, child := range image.children {
			if err := restoreGlobalImage(filepath.Join(path, name), child); err != nil {
				return err
			}
		}

		return os.Chmod(path, globalPermissionMode(image.mode))
	}
	if err := removeManagedTree(root, rel); err != nil {
		return err
	}
	if image.mode&os.ModeSymlink != 0 {
		return atomicManagedLink(root, rel, image.link)
	}
	return atomicManagedFile(root, rel, image.data, globalPermissionMode(image.mode))
}
func (undo *globalUndo) restore() error {
	paths := make([]string, 0, len(undo.entries))
	for path := range undo.entries {
		paths = append(paths, path)
	}
	// Native surfaces must finish compensating before manager records (notably
	// the extended lock) can revert. A failed surface rollback retains recovery
	// evidence instead of leaving an active skill outside an empty lock.
	recoveryRecord := func(path string) bool {
		return undo.managerHome != "" && (path == CurrentFile(undo.managerHome) || strings.HasPrefix(path, ProfilesDir(undo.managerHome)+string(filepath.Separator)))
	}
	sort.Slice(paths, func(i, j int) bool {
		if recoveryRecord(paths[i]) != recoveryRecord(paths[j]) {
			return !recoveryRecord(paths[i])
		}
		if len(paths[i]) != len(paths[j]) {
			return len(paths[i]) > len(paths[j])
		}
		return paths[i] < paths[j]
	})
	// Refuse unknown state before changing any target, including its marker or
	// lock. Inspect recorded parents without following a replacement symlink.
	for parent := range undo.parentPostimages {
		if err := undo.checkParentPostimage(parent); err != nil {
			return err
		}
	}
	for _, path := range paths {
		if err := undo.checkPostimage(path); err != nil {
			return err
		}
	}
	parentsRestored := false
	for _, path := range paths {
		if recoveryRecord(path) && !parentsRestored {
			if err := undo.restoreParents(); err != nil {
				return err
			}
			parentsRestored = true
		}
		// A parent image supersedes any leaves captured before the parent.
		nested := false
		for parent := range undo.entries {
			if path != parent && strings.HasPrefix(path, parent+string(filepath.Separator)) {
				nested = true
				break
			}
		}
		if !nested {
			if err := undo.checkPostimage(path); err != nil {
				return err
			}
			if err := restoreGlobalImage(path, undo.entries[path]); err != nil {
				return err
			}
		}
	}
	if !parentsRestored {
		if err := undo.restoreParents(); err != nil {
			return err
		}
	}
	// Recovery-record removals may have emptied additional created directories.
	return undo.removeAbsentParents()
}

// Parents are shallow metadata images: never enumerate unrelated runtime data.
func readGlobalParent(path string) (*globalImage, error) {
	state, err := stateread.Lstat(path)
	if err != nil {
		return nil, err
	}
	if state.Kind == stateread.KindAbsent {
		return nil, nil
	}
	if state.Kind != stateread.KindPresent || state.Info == nil || !state.Info.IsDir() {
		return nil, fmt.Errorf("%w: compensation parent changed kind: %s", transaction.ErrImplementationCorruption, path)
	}
	owner, err := pathboundary.DefaultOwnerLookup()(path, state.Info)
	if err != nil {
		return nil, err
	}
	return &globalImage{mode: state.Info.Mode(), owner: owner}, nil
}

func (undo *globalUndo) checkParentPostimage(parent string) error {
	current, err := readGlobalParent(parent)
	if err != nil {
		return err
	}
	if current == nil && undo.absentParents[parent] {
		return nil
	}
	published, recorded := undo.parentPostimages[parent]
	if !recorded || current == nil || published == nil || current.mode != published.mode || current.owner != published.owner {
		return fmt.Errorf("%w: compensation parent postimage changed: %s", transaction.ErrImplementationCorruption, parent)
	}
	return nil
}

func (undo *globalUndo) coveredParent(parent string) bool {
	for path := range undo.entries {
		if parent == path || strings.HasPrefix(parent, path+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func (undo *globalUndo) removeAbsentParents() error {
	parents := make([]string, 0, len(undo.absentParents))
	for parent := range undo.absentParents {
		parents = append(parents, parent)
	}
	sort.Slice(parents, func(i, j int) bool { return len(parents[i]) > len(parents[j]) })
	for _, parent := range parents {
		if undo.coveredParent(parent) {
			continue
		}
		if err := undo.checkParentPostimage(parent); err != nil {
			return err
		}
		listing, err := stateread.ReadDir(parent)
		if err != nil {
			return err
		}
		if listing.Kind == stateread.KindAbsent {
			continue
		}
		if listing.Kind != stateread.KindPresent {
			return stateread.UnusableError(parent, fmt.Errorf("compensation parent unreadable"))
		}
		if len(listing.Entries) == 0 {
			if err := undo.checkParentPostimage(parent); err != nil {
				return err
			}
			if err := os.Remove(parent); err != nil {
				return err
			}
		}
	}
	return nil
}

func (undo *globalUndo) restoreParents() error {
	if err := undo.removeAbsentParents(); err != nil {
		return err
	}
	for parent, mode := range undo.parentModes {
		// A complete captured tree already restores its own metadata.
		if undo.coveredParent(parent) {
			continue
		}
		if err := undo.checkParentPostimage(parent); err != nil {
			return err
		}
		state, err := stateread.Lstat(parent)
		if err != nil {
			return err
		}
		if state.Kind != stateread.KindPresent || state.Info == nil || !state.Info.IsDir() {
			return fmt.Errorf("%w: compensation parent changed kind: %s", transaction.ErrImplementationCorruption, parent)
		}
		if state.Info.Mode() != mode {
			if err := undo.checkParentPostimage(parent); err != nil {
				return err
			}
			if err := os.Chmod(parent, globalPermissionMode(mode)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (op *operation) captureNative(native string, adapter Adapter, skills []skillOf) error {
	if op.undo == nil {
		return nil
	}
	paths := []string{filepath.Join(native, adapter.Target), filepath.Join(native, envmarker.Name), filepath.Join(native, ".agent-environment-backup")}
	prior, err := envmarker.Read(native)
	if err != nil {
		return err
	}
	if prior != nil {
		for _, surface := range prior.Surfaces {
			for _, path := range surface.Paths {
				full, err := managedPath(native, path, false)
				if err != nil {
					return err
				}
				paths = append(paths, full)
			}
		}
	}
	for _, skill := range skills {
		paths = append(paths, filepath.Join(native, inPlaceSkillsDir(adapter.ID), skill.name))
	}
	state, err := stateread.Stat(native)
	if err != nil {
		return err
	}
	if state.Kind == stateread.KindPresent {
		for _, path := range paths {
			rel, err := filepath.Rel(native, path)
			if err != nil {
				return err
			}
			if _, err := managedPath(native, filepath.ToSlash(rel), false); err != nil {
				return err
			}
		}
	}
	return op.capture(paths...)
}
func (op *operation) captureManaged(plan *homePlan, seeds *seedBundle, recorded map[string]bool) error {
	if op.undo == nil {
		return nil
	}
	paths := []string{filepath.Join(plan.homeDir, envmarker.Name), filepath.Join(plan.homeDir, ".agent-environment-backup")}
	if plan.adapter.ID == envregistry.Muse {
		// Directory provisioning changes only existence and permissions; do not
		// snapshot unrelated runtime data already held under these directories.
		for _, directory := range []string{"config/muse", "data/muse", "state", "cache"} {
			if err := op.noteParents(filepath.Join(plan.homeDir, directory, ".directory-mode")); err != nil {
				return err
			}
		}
	}
	names := map[string]bool{}
	for path := range recorded {
		names[path] = true
	}
	for path := range plan.copies {
		names[path] = true
	}
	for path := range plan.links {
		names[path] = true
	}
	for path := range seeds.files {
		names[path] = true
	}
	if seeds.claudeInit || plan.adapter.ID == envregistry.ClaudeCode {
		names[".claude.json"] = true
	}
	for path := range names {
		paths = append(paths, filepath.Join(plan.homeDir, filepath.FromSlash(path)))
	}
	return op.capture(paths...)
}

func (op *operation) noteParents(path string) error {
	if op.undo == nil {
		return nil
	}
	if op.undo.absentParents == nil {
		op.undo.absentParents = map[string]bool{}
	}
	for parent := filepath.Dir(path); parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
		state, err := stateread.Lstat(parent)
		if err != nil {
			return err
		}
		if state.Kind == stateread.KindPresent && state.Info != nil {
			return op.noteParentMode(parent, state.Info)
		}
		if state.Kind != stateread.KindAbsent {
			return stateread.UnusableError(parent, fmt.Errorf("compensation parent metadata unreadable"))
		}
		op.undo.absentParents[parent] = true
		if op.undo.parentPostimages == nil {
			op.undo.parentPostimages = map[string]*globalImage{}
		}
		if _, ok := op.undo.parentPostimages[parent]; !ok {
			op.undo.parentPostimages[parent] = nil
		}
	}
	return nil
}

func (op *operation) noteParentMode(path string, info os.FileInfo) error {
	if op.undo.absentParents[path] {
		return nil
	}
	if !info.IsDir() {
		return nil
	}
	if op.undo.parentModes == nil {
		op.undo.parentModes = map[string]os.FileMode{}
	}
	if _, ok := op.undo.parentModes[path]; !ok {
		owner, err := pathboundary.DefaultOwnerLookup()(path, info)
		if err != nil {
			return err
		}
		op.undo.parentModes[path] = info.Mode()
		if op.undo.parentPostimages == nil {
			op.undo.parentPostimages = map[string]*globalImage{}
		}
		op.undo.parentPostimages[path] = &globalImage{mode: info.Mode(), owner: owner}
	}
	return nil
}

type globalInPlaceConflict struct{ error }

func globalInPlaceError(err error) error {
	if strings.Contains(err.Error(), DiagUnmanagedConflict) {
		return &globalInPlaceConflict{err}
	}
	return err
}
