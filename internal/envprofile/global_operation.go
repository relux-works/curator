package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// GlobalSkillTransaction joins profile publication and the legacy continuation
// under the caller's manager-home lock. Ordinary returned errors restore the
// profile's mutable records and surfaces. The §9.4 unmanaged-conflict exception
// retains the extended lock and its install record. Immutable admitted store
// entries may remain.
// The continuation owns its own install journal and rolls its targets back
// before returning an error. This compensation does not replace crash recovery
// of individual profile and install transactions.
func GlobalSkillTransaction(home string, policy Policy, machine envregistry.MachineConfig, addition *DirectSkill, nativeHomeOf func(string) (string, error), lock transaction.HomeLock, next func() error) (info Info, resultErr error) {
	if err := lock.AssertHeld(); err != nil {
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
			resultErr = errors.Join(resultErr, fmt.Errorf("restore global profile publication: %w", err))
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
	managerHome   string
	entries       map[string]*globalImage
	absentParents map[string]bool
	parentModes   map[string]os.FileMode
}
type globalImage struct {
	mode     os.FileMode
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
	image := &globalImage{mode: info.Mode()}
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
		if op.undo.absentParents == nil {
			op.undo.absentParents = map[string]bool{}
		}
		if err := op.noteParents(path); err != nil {
			return err
		}
	}
	return nil
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
		if _, err := managedDirectory(root, rel); err != nil {
			return err
		}
		for name, child := range image.children {
			if err := restoreGlobalImage(filepath.Join(path, name), child); err != nil {
				return err
			}
		}

		return os.Chmod(path, image.mode.Perm())
	}
	if err := removeManagedTree(root, rel); err != nil {
		return err
	}
	if image.mode&os.ModeSymlink != 0 {
		return atomicManagedLink(root, rel, image.link)
	}
	return atomicManagedFile(root, rel, image.data, image.mode.Perm())
}
func (undo *globalUndo) restore() error {
	paths := make([]string, 0, len(undo.entries))
	for path := range undo.entries {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool { return len(paths[i]) > len(paths[j]) })
	var errs []error
	for _, path := range paths {
		// A parent image supersedes any leaves captured before the parent.
		nested := false
		for parent := range undo.entries {
			if path != parent && strings.HasPrefix(path, parent+string(filepath.Separator)) {
				nested = true
				break
			}
		}
		if !nested {
			if err := restoreGlobalImage(path, undo.entries[path]); err != nil {
				errs = append(errs, err)
			}
		}
	}
	parents := make([]string, 0, len(undo.absentParents))
	for parent := range undo.absentParents {
		parents = append(parents, parent)
	}
	sort.Slice(parents, func(i, j int) bool { return len(parents[i]) > len(parents[j]) })
	for _, parent := range parents {
		listing, err := stateread.ReadDir(parent)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if listing.Kind == stateread.KindAbsent {
			continue
		}
		if listing.Kind != stateread.KindPresent {
			errs = append(errs, stateread.UnusableError(parent, fmt.Errorf("compensation parent unreadable")))
			continue
		}
		if len(listing.Entries) == 0 {
			if err := os.Remove(parent); err != nil {
				errs = append(errs, err)
			}
		}
	}

	for parent, mode := range undo.parentModes {
		state, err := stateread.Lstat(parent)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if state.Kind != stateread.KindPresent || state.Info == nil {
			errs = append(errs, stateread.UnusableError(parent, fmt.Errorf("compensation parent disappeared")))
			continue
		}
		info := state.Info
		if !info.IsDir() {
			errs = append(errs, fmt.Errorf("compensation parent changed kind: %s", parent))
			continue
		}
		if info.Mode().Perm() != mode.Perm() {
			if err := os.Chmod(parent, mode.Perm()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
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
			op.noteParentMode(parent, state.Info)
			break
		}
		if state.Kind != stateread.KindAbsent {
			return stateread.UnusableError(parent, fmt.Errorf("compensation parent metadata unreadable"))
		}
		op.undo.absentParents[parent] = true
	}
	return nil
}

func (op *operation) noteParentMode(path string, info os.FileInfo) {
	if op.undo.absentParents[path] {
		return
	}
	if !info.IsDir() {
		return
	}
	if op.undo.parentModes == nil {
		op.undo.parentModes = map[string]os.FileMode{}
	}
	if _, ok := op.undo.parentModes[path]; !ok {
		op.undo.parentModes[path] = info.Mode()
	}
}

type globalInPlaceConflict struct{ error }

func globalInPlaceError(err error) error {
	if strings.Contains(err.Error(), DiagUnmanagedConflict) {
		return &globalInPlaceConflict{err}
	}
	return err
}
