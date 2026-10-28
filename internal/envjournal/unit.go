package envjournal

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/entrydigest"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
)

func unitError(message string) error {
	return fmt.Errorf("%s: pending environment unit: %s", envorigin.DiagInvalid, message)
}

// ValidateUnit binds the durable authority descriptor to the real complete
// journal inventory. A legacy journal conveys no provisioning-origin exception.
func ValidateUnit(j *Journal, manager string) error {
	if j == nil || j.Unit == nil {
		return nil
	}
	u := j.Unit
	if u.ID != j.TransactionID || u.Manager.Path != manager || j.ProjectIdentity != "managed-home-unit" {
		return unitError("journal identity differs")
	}
	if err := u.Validate(); err != nil {
		return err
	}
	if u.LeasePreimage != nil {
		digest, err := u.LeasePreimage.InputDigest()
		if err != nil {
			return err
		}
		if j.ReadInputs[envunit.LeaseReferencePath(manager)] != digest {
			return unitError("reference preimage differs from frozen input")
		}
	}
	if err := ValidateOpaque(j); err != nil {
		return err
	}
	ledgerCount := 0
	if u.LedgerPublication != nil {
		ledgerCount = 1
	}
	if len(u.Targets)+ledgerCount != len(j.Targets)+len(j.Directories)+len(j.Opaque) {
		return unitError("target inventory differs")
	}
	byPath := map[string]envunit.Target{}
	for _, target := range u.Targets {
		byPath[target.Path] = target
	}
	for index, target := range j.Targets {
		stem := fmt.Sprintf(".curator-txn-%s-%03d.", entrydigest.ShortIdentity(j.TransactionID), index)
		if target.BackupPath != filepath.Join(filepath.Dir(target.LivePath), stem+"backup") || target.RollbackPath != filepath.Join(filepath.Dir(target.LivePath), stem+"rollback") {
			return unitError("sidecar identity differs")
		}
		parent := filepath.Dir(target.LivePath)
		for _, directory := range j.Directories {
			rel, err := filepath.Rel(directory.LivePath, parent)
			if directory.Before.Kind == "absent" && err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
				parent = directory.BasePath
				break
			}
		}
		if target.DesiredDigest != "absent" && target.StagedPath != filepath.Join(parent, stem+"desired") {
			return unitError("desired sidecar identity differs")
		}
		if target.Class == "lifecycle" {
			if u.LedgerPublication == nil || target.LivePath != u.LedgerPublication.Path || target.Identifier != target.LivePath || !target.PrivateFile || target.Kind != KindBytes || target.ExpectedGeneration != "" || target.GenerationPath != "" {
				return unitError("structured lifecycle target binding differs")
			}
			continue
		}
		image, ok := byPath[target.LivePath]
		if !ok || image.Class != target.Class || target.Identifier != target.LivePath || (target.Kind != KindEntry && (target.Kind != KindBytes || (!target.PrivateFile && !u.LegacyOwns(target.Class, target.LivePath)))) || target.ExpectedGeneration != "" || target.GenerationPath != "" {
			return unitError("file target binding differs")
		}
		if target.PreimageDigest == "" || (image.Before.Kind == "absent") != (target.PreimageDigest == "absent") || (image.After.Kind == "absent") != (target.DesiredDigest == "absent") {
			return unitError("pre/postimage absence differs")
		}
		if (target.Class == "05-external-cache" || target.Class == "05-external-snapshot") && image.After.Kind != "absent" && !target.PrivateTree {
			return unitError("external replacement lacks durable protection")
		}
		if (target.Class == "origin" || target.Class == "marker" || target.Class == "native-marker" || target.Class == "56-hook-approvals") && image.After.Kind != "absent" && !target.PrivateFile {
			return unitError("authority replacement lacks durable protection")
		}
	}
	for _, directory := range j.Directories {
		if directory.EmptyMuseHome != "" {
			bound := false
			for _, home := range u.Homes {
				if home.Adapter == "muse" && home.Target == directory.EmptyMuseHome {
					bound = true
				}
			}
			if !bound {
				return unitError("empty Muse directory is outside the admitted home")
			}
		}
		image, ok := byPath[directory.LivePath]
		if !ok || image.Before != directory.Before || image.After != directory.After || (image.Class != "parent" && image.Class != "home-directory") {
			return unitError("directory metadata binding differs")
		}
	}
	for _, action := range j.Opaque {
		image, ok := byPath[action.LivePath]
		if !ok || image.Class != OpaqueClass(action) || image.Before != action.Before || image.After.Kind != "absent" {
			return unitError("opaque target inventory differs")
		}
	}
	switch j.Phase {
	case PhasePreparing, PhasePrepared, PhaseCommitting, PhaseCleanup, PhaseRollingBack, PhaseRollbackCleanup:
	default:
		return unitError("unknown phase")
	}
	return nil
}
func sameImage(actual, want envunit.Image) bool {
	if want.Identity == "" && (want.Kind == "file" || want.Kind == "link" || want.Kind == "tree") {
		actual.Identity = ""
		actual.Protection = ""
	}
	return actual == want
}
func observed(path string, want envunit.Image, credential bool) (bool, error) {
	class := "owned-bytes"
	if credential {
		class = "credential-entry"
	}
	if want.Kind == "tree" {
		class = "owned-tree"
	}
	actual, err := envunit.ObserveTarget(path, class)
	if err != nil {
		return false, err
	}
	return sameImage(actual, want), nil
}

// VerifyUnitState observes reachable entry states, including the exact rename
// windows between a durable move and its state write. It authorizes neither an
// unrelated home nor any downstream continuation; callers rescan after recovery.
// The transaction executor additionally validates all sidecar and removal names.
func VerifyUnitState(j *Journal, manager string) error {
	if err := ValidateUnit(j, manager); err != nil {
		return err
	}
	if j.Unit == nil {
		return nil
	}
	u := j.Unit
	if err := ValidateGlobal(j); err != nil {
		return err
	}
	if err := u.VerifyNativeRoots(); err != nil {
		return err
	}
	if err := u.Manager.Verify(); err != nil {
		return err
	}
	if err := VerifyOpaqueState(j); err != nil {
		return err
	}
	beforeByPath := map[string]envunit.Target{}
	for _, target := range u.Targets {
		beforeByPath[target.Path] = target
	}
	seenRemaining := false
	for _, target := range j.Targets {
		if target.Class == "lifecycle" {
			if err := verifyLifecycleTarget(j, target); err != nil {
				return err
			}
			continue
		}
		image := beforeByPath[target.LivePath]
		credential := image.Class == "credential-entry"
		before, err := observed(target.LivePath, image.Before, credential)
		if err != nil {
			return err
		}
		after, err := observed(target.LivePath, image.After, credential)
		if err != nil {
			return err
		}
		absent, err := observed(target.LivePath, envunit.Image{Kind: "absent"}, credential)
		if err != nil {
			return err
		}
		backup, err := observed(target.BackupPath, image.Before, credential)
		if err != nil {
			return err
		}
		backupAbsent, err := observed(target.BackupPath, envunit.Image{Kind: "absent"}, credential)
		if err != nil {
			return err
		}
		// An absent preimage has no backup. A present preimage can occupy only its
		// bound backup entry while the live entry is moved/replaced.
		moved := backup && image.Before.Kind != "absent" || backupAbsent && image.Before.Kind == "absent"
		valid := false
		switch j.Phase {
		case PhasePreparing, PhasePrepared:
			valid = target.State == StatePending && before && backupAbsent
		case PhaseCommitting:
			switch target.State {
			case StatePending:
				seenRemaining = true
				valid = before && backupAbsent || moved && (absent || after)
			case StateBackedUp:
				seenRemaining = true
				valid = moved && (absent || after)
			case StateCommitted:
				valid = (!seenRemaining || j.Global != nil && GlobalPublishedTarget(j, target.LivePath)) && after && moved
			}
		case PhaseCleanup:
			valid = target.State == StateCommitted && after && (moved || backupAbsent)
		case PhaseRollingBack:
			switch target.State {
			case StateRolledBack:
				valid = before && backupAbsent
			case StatePending, StateBackedUp, StateCommitted:
				valid = before && backupAbsent || moved && (after || absent)
			}
		case PhaseRollbackCleanup:
			valid = target.State == StateRolledBack && before && backupAbsent
		}
		if GlobalRetained(j) && !GlobalPublishedTarget(j, target.LivePath) && j.Phase == PhaseCleanup {
			valid = target.State == StatePending && before && backupAbsent
		}
		if !valid {
			return unitError("unreachable live/backup state for " + target.Class)
		}
		if j.Phase == PhasePrepared || j.Phase == PhaseCommitting {
			if target.StagedPath != "" {
				staged, err := observed(target.StagedPath, image.After, credential)
				if err != nil {
					return err
				}
				if !staged && !after {
					return unitError("desired staging differs from unit postimage")
				}
			}
		}
		proof, ok := u.BoundaryProof.Targets[target.LivePath]
		if !ok {
			return unitError("missing boundary")
		}
		_ = proof
		if err := staging.RecheckDurableOne(staging.Target{LivePath: target.LivePath, Kind: staging.KindEntry}, u.BoundaryProof); err != nil {
			return err
		}
	}
	for _, directory := range j.Directories {
		before, err := observed(directory.LivePath, directory.Before, false)
		if err != nil {
			return err
		}
		after, err := observed(directory.LivePath, directory.After, false)
		if err != nil {
			return err
		}
		valid := false
		switch j.Phase {
		case PhasePreparing, PhasePrepared:
			valid = directory.State == StatePending && before
		case PhaseCommitting:
			valid = directory.State == StatePending && (before || after) || directory.State == StateCommitted && after
		case PhaseCleanup:
			valid = directory.State == StateCommitted && after
		case PhaseRollingBack:
			valid = before || after
		case PhaseRollbackCleanup:
			valid = directory.State == StateRolledBack && before
		}
		if GlobalRetained(j) && j.Phase == PhaseCleanup && !GlobalPublishedDirectory(j, directory.LivePath) {
			valid = directory.State == StatePending && before
		}
		if !valid {
			return unitError("unreachable directory state")
		}
	}
	for _, native := range u.Native {
		if _, changing := beforeByPath[native.Root]; !changing {
			directory, err := envunit.ObserveDirectory(native.Root)
			if err != nil {
				return err
			}
			if directory != native.DirectoryBefore {
				return unitError("native custody root was substituted")
			}
		}
	}
	// Source and canonical installed lock remain independent authority inputs.
	// No HOME, selector, credential link or previous-lock filename participates.
	for _, profile := range u.Profiles {
		root := filepath.Join(manager, "profiles", profile.Name)
		if u.Action == envunit.Purge {
			state, err := stateread.Lstat(root)
			if err != nil {
				return err
			}
			if state.Kind != stateread.KindAbsent {
				return unitError("purge installed namespace reappeared")
			}
			continue
		}
		if u.Action == envunit.Remove {
			var present bool
			for _, target := range j.Targets {
				if target.LivePath != root || target.Class != "installed-profile" {
					continue
				}
				live, err := stateread.Lstat(root)
				if err != nil {
					return err
				}
				if live.Kind == stateread.KindAbsent {
					if j.Phase == PhaseCleanup {
						present = true
						root = ""
						break
					}
					backup, err := stateread.Lstat(target.BackupPath)
					if err != nil {
						return err
					}
					if backup.Kind == stateread.KindPresent {
						root = target.BackupPath
						present = true
					} else if j.Phase == PhaseCleanup {
						present = true
						root = ""
					}
				} else {
					present = true
				}
			}
			if !present {
				return unitError("removal lacks exact installed-record target")
			}
			if root == "" {
				continue
			}
		}
		sourceState, err := readUnitRecord(j, filepath.Join(root, "source.json"))
		if err != nil {
			return err
		}
		initialAbsent := (u.Action == envunit.GlobalExtend || u.Action == envunit.Reuse) && profile.OldLock.Kind == "absent" && j.Phase != PhaseCleanup
		if sourceState.Kind != stateread.KindPresent && !initialAbsent {
			return unitError("source is absent outside admitted first-use/removal")
		}
		source := sourceState.Bytes
		if digest := envorigin.Digest(source); sourceState.Kind == stateread.KindPresent && digest != profile.Source && (profile.OldSource == "" || digest != profile.OldSource) {
			return unitError("source identity changed")
		}
		lockState, err := readUnitRecord(j, filepath.Join(root, "lock.json"))
		if err != nil {
			return err
		}
		if lockState.Kind != stateread.KindPresent && !initialAbsent {
			return unitError("installed lock is absent outside admitted first-use/removal")
		}
		canonical := bytes.TrimSpace(lockState.Bytes)
		if lockState.Kind == stateread.KindPresent && !bytes.Equal(canonical, profile.OldLock.Bytes) && !bytes.Equal(canonical, profile.NewLock.Bytes) {
			return unitError("installed lock differs from bound old/new authority")
		}
	}
	for _, home := range u.Homes {
		if u.Action == envunit.Purge {
			directory, err := envunit.ObserveDirectory(home.Root)
			if err != nil {
				return err
			}
			if directory.Kind == "absent" {
				continue
			}
		}
		if _, changing := beforeByPath[home.Root]; !changing {
			directory, err := envunit.ObserveDirectory(home.Root)
			if err != nil {
				return err
			}
			if directory != home.DirectoryBefore {
				return unitError("unchanged home directory was substituted")
			}
		}
		for _, authority := range []struct {
			name string
			raw  []byte
		}{{envorigin.Name, home.Before.Origin}, {".agent-environment.json", home.Before.Marker}} {
			path := filepath.Join(home.Target, authority.name)
			if _, changing := beforeByPath[path]; changing {
				continue
			}
			var raw []byte
			var present bool
			var err error
			if authority.name == envorigin.Name {
				raw, present, err = envorigin.ReadPrivateAuthority(manager, path, envorigin.MaxRecordBytes)
			} else {
				state, readErr := stateread.ReadRegularFile(path)
				raw, present, err = state.Bytes, state.Kind == stateread.KindPresent, readErr
			}
			if err != nil {
				return err
			}
			if present != (len(authority.raw) != 0) || !bytes.Equal(raw, authority.raw) {
				return unitError("unchanged home authority was substituted")
			}
		}
	}
	return nil
}

func VerifyForwardOrigins(j *Journal, manager string) error {
	if j.Unit == nil {
		return nil
	}
	u := j.Unit
	if u.Action == envunit.Remove || u.Action == envunit.Purge {
		return nil
	}
	if j.Phase != PhaseRollingBack && j.Phase != PhaseRollbackCleanup && j.Phase != PhasePreparing {
		for _, home := range u.Homes {
			if home.After.Kind == "recorded" {
				var lock string
				for _, profile := range u.Profiles {
					if profile.Name == home.Profile {
						lock = profile.NewLock.Digest
					}
				}
				if _, err := envorigin.Parse(home.After.Origin, envorigin.Binding{Profile: home.Profile, LockSHA256: lock, ManagedTarget: home.Target}, home.Adapter, manager); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Decode canonicality is checked by independent readers before granting a
// pending exception. The concrete wire admits no extra duplicate/case variants.
func CanonicalEqual(raw []byte, j *Journal, marshal func(any) ([]byte, error)) bool {
	canonical, err := marshal(j)
	return err == nil && bytes.Equal(raw, append(canonical, '\n'))
}

// readUnitRecord handles only rename windows already verified against the full
// unit inventory above. It never consults lock.prev or an ambient alternate.
func readUnitRecord(j *Journal, path string) (stateread.File, error) {
	state, err := stateread.ReadRegularFile(path)
	if err != nil || state.Kind != stateread.KindAbsent {
		return state, err
	}
	for _, target := range j.Targets {
		if target.LivePath != path || target.Class != "profile-record" {
			continue
		}
		for _, candidate := range []string{target.BackupPath, target.StagedPath} {
			if candidate == "" {
				continue
			}
			state, err = stateread.ReadRegularFile(candidate)
			if err != nil || state.Kind == stateread.KindPresent {
				return state, err
			}
		}
	}
	return state, nil
}
