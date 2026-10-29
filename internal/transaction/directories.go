package transaction

import (
	"fmt"
	"github.com/relux-works/curator/internal/envjournal"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/privatedir"
)

const metadataJournalSchema = "curator-install-transaction-v2"

// PublicationDirectories captures missing manager-owned parents and protection
// of existing managed-environment parents without mutating either. Other
// existing manager/native parents keep their current metadata. External native
// surfaces need their separately admitted boundary plan.
func (engine *Engine) PublicationDirectories(paths ...string) ([]DirectoryAction, error) {
	byPath := map[string]DirectoryAction{}
	for _, path := range paths {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return nil, journalf("unclean publication path")
		}
		parent := filepath.Dir(path)
		if !insideDirectory(parent, engine.home) {
			continue
		}
		for {
			before, err := envunit.ObserveDirectory(parent)
			if err != nil {
				return nil, err
			}
			if before.Kind != "absent" {
				if err := pathboundary.ValidateRouteWithOwner(engine.home, parent, pathboundary.DefaultOwnerLookup()); err != nil {
					return nil, err
				}
				if insideDirectory(parent, filepath.Join(engine.home, "environments")) {
					if before.Mode != 0700 {
						byPath[parent] = DirectoryAction{LivePath: parent, Before: before, Mode: 0700}
					}
					parent = filepath.Dir(parent)
					continue
				}
				break
			}
			if parent == engine.home {
				return nil, corruptionf("manager home is absent")
			}
			byPath[parent] = DirectoryAction{LivePath: parent, Before: before, Mode: 0700}
			parent = filepath.Dir(parent)
		}
	}
	result := make([]DirectoryAction, 0, len(byPath))
	for _, action := range byPath {
		result = append(result, action)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].LivePath < result[j].LivePath })
	return result, nil
}

func directorySidecar(base, id string, index int) string {
	return filepath.Join(base, fmt.Sprintf(".curator-txn-%s-dir-%03d.desired", shortIdentity(id), index))
}

func insideDirectory(path, dir string) bool {
	return path == dir || strings.HasPrefix(path, dir+string(filepath.Separator))
}

// buildDirectories is read-only. Staging directories are allocated only after
// all plan validation succeeds; an interruption before journal creation can
// leave a private empty staging artifact, never a live home or an authority.
func buildDirectories(plan Plan) ([]DirectoryRecord, error) {
	actions := append([]DirectoryAction(nil), plan.Directories...)
	sort.Slice(actions, func(i, j int) bool { return actions[i].LivePath < actions[j].LivePath })
	var records []DirectoryRecord
	for _, action := range actions {
		if !filepath.IsAbs(action.LivePath) || filepath.Clean(action.LivePath) != action.LivePath || action.Mode != 0700 {
			return nil, journalf("directory action requires a canonical private path")
		}
		if len(records) > 0 && records[len(records)-1].LivePath == action.LivePath {
			return nil, journalf("duplicate directory action")
		}
		before, err := envunit.ObserveDirectory(action.LivePath)
		if err != nil || before != action.Before {
			return nil, corruptionf("directory preimage changed: %s: %v", action.LivePath, err)
		}
		base := filepath.Dir(action.LivePath)
		for {
			state, err := envunit.ObserveDirectory(base)
			if err != nil {
				return nil, err
			}
			if state.Kind != "absent" {
				break
			}
			known := false
			for _, prior := range records {
				if prior.LivePath == base {
					known = true
					break
				}
			}
			if !known {
				return nil, journalf("missing parent has no directory action: %s", base)
			}
			base = filepath.Dir(base)
		}
		baseImage, err := envunit.ObserveDirectory(base)
		if err != nil {
			return nil, err
		}
		after := before
		after.Kind, after.Mode = "directory", action.Mode
		record := DirectoryRecord{EmptyMuseHome: action.EmptyMuseHome, LivePath: action.LivePath, Before: before, After: after, BasePath: base, Base: baseImage, State: StatePending}
		if before.Kind == "absent" {
			record.StagedPath = directorySidecar(base, plan.TransactionID, len(records))
		}
		records = append(records, record)
	}
	return records, nil
}

// desiredParent keeps staging on the destination filesystem while its live
// parent is absent. Backup and rollback entries still use the live parent.
func desiredParent(live string, directories []DirectoryRecord) string {
	for _, dir := range directories {
		if dir.Before.Kind == "absent" && insideDirectory(filepath.Dir(live), dir.LivePath) {
			return dir.BasePath
		}
	}
	return filepath.Dir(live)
}

func stagedTargetPath(live, id string, index int, dirs []DirectoryRecord) string {
	return filepath.Join(desiredParent(live, dirs), filepath.Base(sidecarPath(live, id, index, "desired")))
}

func (engine *Engine) verifyDirectoryBase(journal *Journal, dir DirectoryRecord) error {
	if !insideDirectory(dir.BasePath, engine.home) {
		return engine.verifyNativeDirectoryBase(journal, dir)
	}
	if err := pathboundary.ValidateRouteWithOwner(engine.home, dir.BasePath, pathboundary.DefaultOwnerLookup()); err != nil {
		return err
	}
	expected := dir.Base
	for _, other := range journal.Directories {
		if other.LivePath != dir.BasePath {
			continue
		}
		if other.State == StateCommitted {
			expected = other.After
		} else if other.State == StateRolledBack {
			expected = other.Before
		} else if journal.Phase != PhasePreparing && journal.Phase != PhasePrepared {
			// A metadata change may be durable just before its progress write.
			// Admit only this journal's exact pinned postimage in that window;
			// existence, current permissions or an unrelated journal are no proof.
			actual, err := envunit.ObserveDirectory(dir.BasePath)
			if err != nil {
				return err
			}
			if actual == other.After {
				expected = other.After
			}
		}
	}
	return verifyDirectoryImage(dir.BasePath, expected)
}

func (engine *Engine) allocateDirectoryStaging(journal *Journal) error {
	for index := range journal.Directories {
		dir := &journal.Directories[index]
		if dir.StagedPath == "" {
			after, err := prepareExistingDirectoryProtection(*dir, index, journal.TransactionID)
			if err != nil {
				return err
			}
			dir.After = after
			continue
		}
		if err := engine.verifyDirectoryBase(journal, *dir); err != nil {
			return err
		}
		if err := privatedir.Make(dir.StagedPath); err != nil {
			return err
		}
		after, err := envunit.ObserveDirectory(dir.StagedPath)
		if err != nil {
			return err
		}
		dir.After = after
		if err := syncDirectory(dir.StagedPath); err != nil {
			return err
		}
		if err := syncDirectory(dir.BasePath); err != nil {
			return err
		}
	}
	return nil
}

func verifyDirectoryImage(path string, expected envunit.Image) error {
	got, err := envunit.ObserveDirectory(path)
	if err != nil || got != expected {
		return corruptionf("directory identity/protection changed: %s: %v", path, err)
	}
	return nil
}

// buildJournal may hold an unallocated placeholder while it validates the
// read-only plan. No placeholder can cross the durable save/load boundary:
// interrupted recovery must identify the exact prepared inode, not an empty
// identity interpreted as permission to allocate a new one.
func validateDurableDirectories(journal *Journal) error {
	for _, dir := range journal.Directories {
		pin := envunit.Manager{Path: dir.LivePath, Identity: dir.After.Identity, Owner: dir.After.Owner}
		if err := pin.Validate(); err != nil {
			return journalf("unbound durable directory: %v", err)
		}
	}
	return nil
}

func validateDirectories(journal *Journal) error {
	if journal.Schema == journalSchema {
		if len(journal.Directories) != 0 {
			return journalf("legacy journal cannot carry metadata actions")
		}
		return nil
	}
	if journal.Schema != metadataJournalSchema || len(journal.Directories) == 0 {
		return journalf("metadata journal has no actions")
	}
	for i, dir := range journal.Directories {
		if !filepath.IsAbs(dir.LivePath) || filepath.Clean(dir.LivePath) != dir.LivePath || !validText(dir.LivePath) || !filepath.IsAbs(dir.BasePath) || filepath.Clean(dir.BasePath) != dir.BasePath || !insideDirectory(filepath.Dir(dir.LivePath), dir.BasePath) {
			return journalf("invalid directory path")
		}
		if i > 0 && journal.Directories[i-1].LivePath >= dir.LivePath {
			return journalf("directory actions not canonical")
		}
		validImage := func(im envunit.Image) bool {
			return im.Kind == "directory" && im.Identity != "" && im.Owner != "" && im.Mode <= 0777 && im.Mode&0022 == 0 && im.Digest == "" && im.LinkTarget == "" && envunit.ValidProtection(im.Protection)
		}
		if !validImage(dir.Base) || dir.After.Kind != "directory" || dir.After.Mode != 0700 {
			return journalf("invalid directory identity/protection")
		}
		if dir.Before.Kind == "absent" {
			if dir.Before != (envunit.Image{Kind: "absent"}) || dir.StagedPath != directorySidecar(dir.BasePath, journal.TransactionID, i) {
				return journalf("invalid absent directory action")
			}
			if dir.After.Identity == "" && journal.Phase == PhasePreparing { /* read-only build only */
			} else if !validImage(dir.After) {
				return journalf("missing prepared directory identity")
			}
		} else if !validImage(dir.Before) || dir.After.Identity != dir.Before.Identity || dir.After.Owner != dir.Before.Owner || dir.StagedPath != "" {
			return journalf("directory action changes unrelated metadata")
		}
		if dir.State != StatePending && dir.State != StateCommitted && dir.State != StateRolledBack {
			return journalf("invalid directory progress")
		}
		if (journal.Phase == PhasePreparing || journal.Phase == PhasePrepared) && dir.State != StatePending {
			return journalf("directory progress before commit")
		}
		used := false
		for _, target := range journal.Targets {
			if target.LivePath == dir.LivePath || insideDirectory(dir.LivePath, target.LivePath) {
				return journalf("directory metadata overlaps target replacement")
			}
			if insideDirectory(filepath.Dir(target.LivePath), dir.LivePath) {
				used = true
			}
		}
		if !used && dir.EmptyMuseHome == "" {
			return journalf("directory action has no owned publication")
		}
		if dir.EmptyMuseHome != "" {
			if !filepath.IsAbs(dir.EmptyMuseHome) || filepath.Clean(dir.EmptyMuseHome) != dir.EmptyMuseHome {
				return journalf("unclean empty Muse home")
			}
			rel, err := filepath.Rel(dir.EmptyMuseHome, dir.LivePath)
			if err != nil || (rel != "state" && rel != "cache" && rel != filepath.Join("config", "muse") && rel != filepath.Join("data", "muse")) {
				return journalf("unowned empty Muse directory")
			}
		}
		for path := range journal.ReadInputs {
			if overlapsReadPath(path, dir.LivePath) || (dir.StagedPath != "" && overlapsReadPath(path, dir.StagedPath)) {
				return journalf("read input overlaps changing directory")
			}
		}
	}
	return nil
}

func (engine *Engine) emitDirectory(journal *Journal, index int) error {
	event := Event{Point: PointAfterDirectoryInstall, TransactionID: journal.TransactionID, TargetIndex: -1, Class: "directory", Identifier: journal.Directories[index].LivePath, LivePath: journal.Directories[index].LivePath}
	if engine.hooks.Observe != nil {
		engine.hooks.Observe(event)
	}
	if engine.hooks.Fault != nil {
		return engine.hooks.Fault(event)
	}
	return nil
}

func (engine *Engine) commitDirectories(journal *Journal) error {
	return engine.commitDirectorySubset(journal, func(string) bool { return true })
}
func (engine *Engine) commitDirectorySubset(journal *Journal, selected func(string) bool) error {
	for i := range journal.Directories {
		if !selected(journal.Directories[i].LivePath) {
			continue
		}
		dir := &journal.Directories[i]
		if err := engine.verifyDirectoryBase(journal, *dir); err != nil {
			return err
		}
		if err := engine.checkCommit(journal); err != nil {
			return err
		}
		if err := checkReadInputs(journal.ReadInputs); err != nil {
			return err
		}
		got, err := envunit.ObserveDirectory(dir.LivePath)
		if err != nil {
			return err
		}
		if dir.State == StateCommitted {
			if got != dir.After {
				return corruptionf("committed directory changed")
			}
			continue
		}
		if dir.State != StatePending {
			return corruptionf("cannot forward rolled-back directory")
		}
		if got != dir.After {
			if got != dir.Before {
				return corruptionf("directory preimage changed")
			}
			if err := engine.checkDirectoryParent(journal, *dir); err != nil {
				return err
			}
			if dir.Before.Kind == "absent" {
				if err := verifyDirectoryImage(dir.StagedPath, dir.After); err != nil {
					return err
				}
				if err := engine.renameNoReplace(dir.StagedPath, dir.LivePath); err != nil {
					return err
				}
			} else if err := envunit.ApplyDirectoryProtection(dir.LivePath, dir.After); err != nil {
				return err
			}
			if err := syncDirectory(dir.LivePath); err != nil {
				return err
			}
			if err := syncDirectory(filepath.Dir(dir.LivePath)); err != nil {
				return err
			}
			if err := verifyDirectoryImage(dir.LivePath, dir.After); err != nil {
				return err
			}
			if err := engine.emitDirectory(journal, i); err != nil {
				return err
			}
		}
		dir.State = StateCommitted
		delete(engine.nsCache, journal.TransactionID)
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
	}
	return nil
}

func (engine *Engine) checkDirectoryParent(journal *Journal, dir DirectoryRecord) error {
	parent := filepath.Dir(dir.LivePath)
	for _, prior := range journal.Directories {
		if prior.LivePath == parent {
			if prior.State != StateCommitted {
				return corruptionf("directory parent not committed")
			}
			return verifyDirectoryImage(parent, prior.After)
		}
	}
	if parent != dir.BasePath {
		return corruptionf("directory parent not bound")
	}
	return verifyDirectoryImage(parent, dir.Base)
}

func verifyCommittedDirectories(journal *Journal) error {
	for _, dir := range journal.Directories {
		if journal.Global != nil && journal.Global.State != "forward" && !envjournal.GlobalPublishedDirectory(journal, dir.LivePath) {
			if dir.State != StatePending {
				return corruptionf("retained nonpublication directory changed")
			}
			if err := verifyDirectoryImage(dir.LivePath, dir.Before); err != nil {
				return err
			}
			continue
		}
		if dir.State != StateCommitted {
			return corruptionf("directory publication incomplete")
		}
		if err := verifyDirectoryImage(dir.LivePath, dir.After); err != nil {
			return err
		}
	}
	return nil
}

func (engine *Engine) rollbackDirectories(journal *Journal) error {
	for i := len(journal.Directories) - 1; i >= 0; i-- {
		dir := &journal.Directories[i]
		if err := engine.verifyDirectoryBase(journal, *dir); err != nil {
			return err
		}
		got, err := envunit.ObserveDirectory(dir.LivePath)
		if err != nil {
			return err
		}
		if got != dir.Before {
			if got != dir.After {
				return corruptionf("directory rollback refuses substituted entry")
			}
			if err := engine.checkDirectoryParent(journal, *dir); err != nil {
				return err
			}
			if dir.Before.Kind == "absent" {
				if err := os.Remove(dir.LivePath); err != nil {
					return err
				}
			} else {
				if err := envunit.ApplyDirectoryProtection(dir.LivePath, dir.Before); err != nil {
					return err
				}
				if err := syncDirectory(dir.LivePath); err != nil {
					return err
				}
			}
			if err := syncDirectory(filepath.Dir(dir.LivePath)); err != nil {
				return err
			}
		}
		dir.State = StateRolledBack
		if err := engine.saveJournal(journal); err != nil {
			return err
		}
	}
	return engine.cleanupDirectoryStaging(journal)
}

func (engine *Engine) cleanupDirectoryStaging(journal *Journal) error {
	for _, dir := range journal.Directories {
		if dir.StagedPath == "" {
			continue
		}
		if err := engine.verifyDirectoryBase(journal, dir); err != nil {
			return err
		}
		got, err := envunit.ObserveDirectory(dir.StagedPath)
		if err != nil {
			return err
		}
		if got.Kind == "absent" {
			continue
		}
		if got != dir.After {
			return corruptionf("unknown directory staging preserved")
		}
		// Remove only an empty inode whose complete metadata was recorded. Never
		// enumerate or recursively delete an arbitrary home or session tree.
		if err := os.Remove(dir.StagedPath); err != nil {
			return err
		}
		if err := syncDirectory(dir.BasePath); err != nil {
			return err
		}
	}
	return nil
}
