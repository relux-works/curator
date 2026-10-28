package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// bindMigration captures complete affected home authority before publication.
// Identity migration rebinds every materialized adapter to the staged lock;
// credential migration preserves the original record byte-for-byte. Credentials
// are entry metadata/link text only, never arbitrary files or home snapshots.
func (m *identityMigration) bindMigration(home, id string) (*envunit.Unit, map[string]string, error) {
	manager, err := envunit.ObserveManager(home)
	if err != nil {
		return nil, nil, err
	}
	unit := &envunit.Unit{Record: envunit.Version, ID: id, Manager: manager, Action: envunit.Maintain}
	profiles := map[string]bool{}
	for path := range m.targets {
		// A credential-only correction can preserve a schema-1 marker exactly.
		// Its actual home target still requires complete origin/unit custody.
		if rel, err := filepath.Rel(EnvRoot(home), path); err == nil {
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) >= 3 && parts[0] != ".." && parts[0] != "." {
				profiles[parts[0]] = true
			}
		}
		if filepath.Base(path) == "lock.json" && filepath.Dir(filepath.Dir(path)) == ProfilesDir(home) {
			profiles[filepath.Base(filepath.Dir(path))] = true
		}
		if filepath.Base(path) == envmarker.Name {
			rel, e := filepath.Rel(EnvRoot(home), path)
			parts := strings.Split(rel, string(filepath.Separator))
			if e == nil && len(parts) == 3 {
				profiles[parts[0]] = true
			}
		}
	}
	var names []string
	for name := range profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	inputs := map[string]string{}
	for _, profile := range names {
		old, _, e := readLock(home, profile)
		if e != nil {
			return nil, nil, e
		}
		before, e := old.Canonical()
		if e != nil {
			return nil, nil, e
		}
		after := before
		next := old
		if stage, ok := m.targets[lockPath(home, profile)]; ok {
			raw, e := os.ReadFile(stage)
			if e != nil {
				return nil, nil, e
			}
			next, e = contextlock.Parse(raw)
			if e != nil {
				return nil, nil, e
			}
			after, e = next.Canonical()
			if e != nil {
				return nil, nil, e
			}
		}
		changed := !bytes.Equal(before, after)
		if changed {
			unit.Action = envunit.Rebind
		}
		source, present, e := envorigin.ReadOwnedRecord(home, sourcePath(home, profile), 16<<20)
		if e != nil || !present {
			return nil, nil, fmt.Errorf("migration installed source unavailable: %w", e)
		}
		p := envunit.ProfileTransition{Name: profile, Source: envorigin.Digest(source), OldLock: envunit.RecordImage{Kind: "file", Bytes: before, Digest: envorigin.Digest(before)}, NewLock: envunit.RecordImage{Kind: "file", Bytes: after, Digest: envorigin.Digest(after)}}
		for _, member := range next.Members {
			raw, e := json.Marshal(member)
			if e != nil {
				return nil, nil, e
			}
			p.Closure = append(p.Closure, envorigin.Digest(raw))
		}
		unit.Profiles = append(unit.Profiles, p)
		for _, path := range []string{sourcePath(home, profile), lockPath(home, profile)} {
			if _, targeted := m.targets[path]; !targeted {
				digest, e := transaction.DigestPath(path)
				if e != nil {
					return nil, nil, e
				}
				inputs[path] = digest
			}
		}
		for _, adapter := range envregistry.Registry {
			target := ManagedHomeDir(home, profile, adapter.ID)
			markerPath := filepath.Join(target, envmarker.Name)
			staged, targeted := m.targets[markerPath]
			affected := targeted
			root := envorigin.ManagedRoot(home, profile, adapter.ID)
			for path := range m.targets {
				affected = affected || path == root || strings.HasPrefix(path, root+string(filepath.Separator))
			}
			if !changed && !affected {
				continue
			}
			marker, e := envmarker.Read(target)
			if e != nil {
				return nil, nil, e
			}
			if marker == nil {
				continue
			}
			rr := ResolveRequest{Home: home, Profile: profile, EnvID: adapter.ID}
			if e = rr.observeOrigin(adapter.ID, marker); e != nil {
				return nil, nil, e
			}
			prior, e := os.ReadFile(markerPath)
			if e != nil {
				return nil, nil, e
			}
			desired := prior
			if targeted {
				desired, e = os.ReadFile(staged)
				if e != nil {
					return nil, nil, e
				}
			} else if changed {
				return nil, nil, fmt.Errorf("%s: identity migration omitted materialized adapter marker", envorigin.DiagInvalid)
			}
			oldOrigin := append([]byte(nil), rr.origin.raw...)
			newOrigin := oldOrigin
			if changed {
				binding := rr.origin.record.Binding
				binding.LockSHA256 = "sha256:" + strings.TrimPrefix(contextlock.HashBytes(after), "sha256:")
				newOrigin, e = envorigin.Rebind(oldOrigin, rr.origin.record.Binding, binding, adapter.ID, home)
				if e != nil {
					return nil, nil, e
				}
				if e = m.bytes(filepath.Join(target, envorigin.Name), newOrigin); e != nil {
					return nil, nil, e
				}
			}
			directory, e := envunit.ObserveDirectory(root)
			if e != nil {
				return nil, nil, e
			}
			unit.Homes = append(unit.Homes, envunit.Home{Profile: profile, Adapter: adapter.ID, Mode: "managed-home", Target: target, Root: root, DirectoryBefore: directory, DirectoryAfter: directory, Before: envunit.HomeImage{Kind: "recorded", Marker: prior, Origin: oldOrigin}, After: envunit.HomeImage{Kind: "recorded", Marker: desired, Origin: newOrigin}})
		}
	}
	if len(unit.Homes) == 0 {
		return nil, inputs, nil
	}
	return unit, inputs, nil
}

func (m *identityMigration) publish(op *operation) error {
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "profile-hash-migration"}
	var err error
	plan.Unit, plan.ReadInputs, err = m.bindMigration(op.engine.Home(), plan.TransactionID)
	if err != nil {
		return err
	}
	if plan.Unit != nil {
		plan.ProjectIdentity = "managed-home-unit"
	}
	var paths []string
	for path := range m.targets {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var boundaries []staging.Target
	for _, path := range paths {
		class, private := "owned-bytes", false
		if filepath.Base(path) == envmarker.Name {
			class, private = "marker", true
		}
		if filepath.Base(path) == envorigin.Name {
			class, private = "origin", true
		}
		if filepath.Dir(filepath.Dir(path)) == ProfilesDir(op.engine.Home()) && (filepath.Base(path) == "lock.json" || filepath.Base(path) == "source.json") {
			class, private = "profile-record", true
		}
		if plan.Unit != nil {
			for _, h := range plan.Unit.Homes {
				auth := "auth.json"
				if h.Adapter == "claude_code" {
					auth = ".credentials.json"
				}
				if h.Adapter == "muse" {
					auth = filepath.Join("config", "muse", "auth.json")
				}
				if path == filepath.Join(h.Target, auth) && h.Adapter != "opencode" {
					class, private = "credential-entry", false
				}
			}
		}
		if class == "owned-bytes" {
			for _, candidate := range []string{path, m.targets[path]} {
				if candidate == "" {
					continue
				}
				info, err := os.Lstat(candidate)
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				if err == nil && info.IsDir() {
					class = "owned-tree"
				}
			}
		}
		if plan.Unit != nil {
			if err := plan.Unit.ValidateReadScope(class, path); err != nil {
				return err
			}
		}
		for _, candidate := range []string{path, m.targets[path]} {
			if candidate != "" {
				if err := envunit.ValidateEntryRead(candidate, class); err != nil {
					return err
				}
			}
		}
		kind := transaction.KindEntry
		if private {
			kind = transaction.KindBytes
		}
		digest, e := transaction.DigestTarget(kind, path)
		if e != nil {
			return e
		}
		plan.Targets = append(plan.Targets, transaction.Target{Class: class, Identifier: path, Kind: kind, LivePath: path, StagedSource: m.targets[path], PreimageDigest: digest, PrivateFile: private})
		boundaries = append(boundaries, staging.Target{LivePath: path, Kind: string(kind)})
	}
	plan.Directories, err = op.engine.PublicationDirectories(paths...)
	if err != nil {
		return err
	}
	snapshot, err := (staging.Plan{Targets: boundaries}).Snapshot(nil)
	if err != nil {
		return err
	}
	proof, err := snapshot.Durable()
	if err != nil {
		return err
	}
	plan.BoundaryProof = &proof
	if _, err = op.engine.Prepare(op.lock, plan); err != nil {
		return fmt.Errorf("prepare profile migration: %w", err)
	}
	if err = op.engine.Commit(op.lock, plan.TransactionID); err != nil {
		return fmt.Errorf("publish profile migration: %w", err)
	}
	return nil
}

// New credential migrations share the common whole-unit publisher. The old
// migration journal remains a guarded recovery input only; it grants no origin
// exception and is never emitted for new operations.
func publishCredentialMigration(op *operation, req MigrateRequest, report *MigrateReport) (resultErr error) {
	root, err := privatedir.TempStaging("", "curator-credential-migration-*")
	if err != nil {
		return err
	}
	m := &identityMigration{root: root, targets: map[string]string{}}
	defer m.cleanup()
	changed, _, err := changedMarkers(report)
	if err != nil {
		return err
	}
	for path, payload := range changed {
		if err = m.bytes(path, payload); err != nil {
			return err
		}
	}
	for _, action := range report.Ops() {
		live := filepath.Join(ManagedHomeDir(req.Home, action.Profile, action.EnvID), filepath.FromSlash(action.Path))
		if err = validateMigrationOp(live, action); err != nil {
			return err
		}
		if action.Kind == MigrateOpUnlink {
			m.targets[live] = ""
		} else if err = m.link(live, action.To); err != nil {
			return err
		}
	}
	return m.publish(op)
}
