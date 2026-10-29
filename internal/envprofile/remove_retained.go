package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/projectlease"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// removeRetained commits installed-record absence and durable nonlaunchable
// custody in one unit. Retained homes are observed only through their exact
// authority files and directory metadata: sessions and credentials are opaque.
func removeRetained(op *operation, manager, profile string) (resultErr error) {
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return err
	}
	lock, _, err := readLock(manager, profile)
	if err != nil {
		return err
	}
	canonical, err := lock.Canonical()
	if err != nil {
		return err
	}
	source, err := stateread.ReadRegularFile(sourcePath(manager, profile))
	if err != nil {
		return err
	}
	if source.Kind != stateread.KindPresent {
		return fmt.Errorf("%s: installed source absent", envorigin.DiagInvalid)
	}
	image := envunit.RecordImage{Kind: "file", Bytes: canonical, Digest: envorigin.Digest(canonical)}
	var closure []string
	for _, member := range lock.Members {
		raw, err := json.Marshal(member)
		if err != nil {
			return err
		}
		closure = append(closure, envorigin.Digest(raw))
	}
	unit := &envunit.Unit{Record: envunit.Version, ID: newTransactionID(), Manager: identity, Action: envunit.Remove, Homes: []envunit.Home{}, Profiles: []envunit.ProfileTransition{{Name: profile, Source: envorigin.Digest(source.Bytes), Closure: closure, OldLock: image, NewLock: envunit.RecordImage{Kind: "absent"}}}}
	for _, adapter := range envregistry.Registry {
		target := envorigin.ManagedTarget(manager, profile, adapter.ID)
		marker, err := envmarker.Read(target)
		if err != nil {
			return err
		}
		origin, present, err := envorigin.ReadPrivateAuthority(manager, filepath.Join(target, envorigin.Name), envorigin.MaxRecordBytes)
		if err != nil {
			return err
		}
		if marker == nil {
			if present {
				return fmt.Errorf("%s: sidecar without managed marker", envorigin.DiagInvalid)
			}
			continue
		}
		if err := requirePerHomeCredentialSource(marker); err != nil {
			return err
		}
		if !present {
			return fmt.Errorf("%s: adopt the legacy origin before retaining a materialized home", envorigin.DiagRequired)
		}
		if marker.Mode != envmarker.ModeManagedHome || marker.Profile.Name != profile || marker.Profile.LockSHA256 != image.Digest[len("sha256:"):] {
			return fmt.Errorf("%s: removal marker binding differs", envorigin.DiagInvalid)
		}
		if _, err := envorigin.ParseCaptured(origin, envorigin.Binding{Profile: profile, LockSHA256: image.Digest, ManagedTarget: target}, adapter.ID, manager); err != nil {
			return err
		}
		mark, err := stateread.ReadRegularFile(filepath.Join(target, envmarker.Name))
		if err != nil {
			return err
		}
		root := envorigin.ManagedRoot(manager, profile, adapter.ID)
		directory, err := envunit.ObserveDirectory(root)
		if err != nil {
			return err
		}
		before := envunit.HomeImage{Kind: "recorded", Marker: mark.Bytes, Origin: origin}
		unit.Homes = append(unit.Homes, envunit.Home{Profile: profile, Adapter: adapter.ID, Mode: "managed-home", Target: target, Root: root, DirectoryBefore: directory, DirectoryAfter: directory, Before: before, After: before})
	}
	witness, ok := op.lock.(interface{ Held() managerlock.Held })
	if !ok {
		return managerlock.ErrNotHeld
	}
	references, err := projectlease.PrepareProfileRemoval(manager, profile, witness.Held())
	if err != nil {
		return err
	}
	installed := ProfileDir(manager, profile)
	entries, err := os.ReadDir(installed)
	if err != nil {
		return err
	}
	var records []envlifecycle.OwnedRecord
	for _, entry := range entries {
		name := entry.Name()
		// rendered is the fixed manager-authored document projection namespace,
		// not a managed tool home. The enclosing record target binds its complete
		// link-free preimage; retained home links intentionally become nonlaunchable.
		if name == "rendered" && entry.IsDir() {
			continue
		}
		if name != "source.json" && name != "lock.json" && name != "lock.prev.json" {
			return fmt.Errorf("%s: unclassified installed record %q prevents removal", envorigin.DiagInvalid, name)
		}
		state, err := stateread.ReadRegularFile(filepath.Join(installed, name))
		if err != nil {
			return err
		}
		if state.Kind != stateread.KindPresent {
			return fmt.Errorf("%s: installed record disappeared", envorigin.DiagInvalid)
		}
		raw := state.Bytes
		if name == "lock.json" {
			if !bytes.Equal(bytes.TrimSpace(raw), canonical) {
				return fmt.Errorf("%s: installed lock changed", envorigin.DiagInvalid)
			}
			raw = canonical
		}
		records = append(records, envlifecycle.OwnedRecord{Path: filepath.Join(installed, name), Image: envunit.RecordImage{Kind: "file", Bytes: raw, Digest: envorigin.Digest(raw)}})
	}
	_, prior, err := envlifecycle.Read(identity)
	if err != nil {
		return err
	}
	if err := unit.ValidateReadScope("installed-profile", installed); err != nil {
		return err
	}
	digest, err := transaction.DigestTarget(transaction.KindEntry, installed)
	if err != nil {
		return err
	}
	plan := transaction.Plan{TransactionID: unit.ID, ProjectIdentity: "managed-home-unit", Unit: unit, Lifecycle: &envlifecycle.Change{Prior: prior, RemovedRecords: records, Revocations: []envlifecycle.Revocation{}}, Targets: []transaction.Target{{Class: "installed-profile", Identifier: installed, Kind: transaction.KindEntry, LivePath: installed, PreimageDigest: digest}}}
	if references != nil {
		pub, e := newHomePublication(op)
		if e != nil {
			return e
		}
		defer finishPublicationCleanup(&resultErr, pub.close)
		if e = pub.file(references.Path, "reference-record", references.After, 0600, true); e != nil {
			return e
		}
		for _, target := range pub.targets {
			plan.Targets = append(plan.Targets, target)
		}
		plan.Lifecycle.Revocations = append(plan.Lifecycle.Revocations, envlifecycle.Revocation{Path: references.Path, Before: envunit.RecordImage{Kind: "file", Bytes: references.Before, Digest: envorigin.Digest(references.Before)}, After: envunit.RecordImage{Kind: "file", Bytes: references.After, Digest: envorigin.Digest(references.After)}})
	}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return err
	}
	return op.engine.Commit(op.lock, plan.TransactionID)
}

// The common scan and actual retained receipt, rather than missing source alone,
// make a repeated remove completion idempotent after its final journal cleanup.
func removedProfileRecorded(manager, profile string) (bool, error) {
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return false, err
	}
	ledger, _, err := envlifecycle.Read(identity)
	if err != nil {
		return false, err
	}
	for _, entry := range ledger.Entries {
		if entry.Removed != nil || entry.Purged != nil {
			for _, name := range entry.Profiles() {
				if name == profile {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
