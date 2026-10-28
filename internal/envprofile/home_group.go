package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

type homePublicationGroup struct {
	unit         *envunit.Unit
	targets      map[string]transaction.Target
	opaque       []transaction.OpaqueTarget
	inputs       map[string]string
	publications []*homePublication
	emptyMuse    map[string]string
}

func isActivationPath(home, path string) bool { return envunit.ActivationPath(home, path) }

func (g *homePublicationGroup) close() error {
	var result error
	for _, p := range g.publications {
		result = errors.Join(result, p.close())
	}
	return result
}
func (g *homePublicationGroup) add(plan transaction.Plan, pub *homePublication) error {
	if g.unit == nil {
		g.unit = plan.Unit
		g.unit.ID = newTransactionID()
	} else {
		for _, p := range plan.Unit.Profiles {
			exists := false
			for _, prior := range g.unit.Profiles {
				if p.Name == prior.Name {
					a, _ := json.Marshal(p)
					b, _ := json.Marshal(prior)
					if !bytes.Equal(a, b) {
						return fmt.Errorf("%s: group profile authority changed", envorigin.DiagInvalid)
					}
					exists = true
				}
			}
			if !exists {
				g.unit.Profiles = append(g.unit.Profiles, p)
			}
		}
		g.unit.Homes = append(g.unit.Homes, plan.Unit.Homes...)
		g.unit.Native = append(g.unit.Native, plan.Unit.Native...)
	}
	for _, target := range plan.Targets {
		if prior, duplicate := g.targets[target.LivePath]; duplicate {
			if prior.Class != target.Class || prior.Kind != target.Kind || prior.PrivateFile != target.PrivateFile || prior.PreimageDigest != target.PreimageDigest {
				return fmt.Errorf("incompatible group target: %s", target.LivePath)
			}
			a, e := transaction.DigestTarget(prior.Kind, prior.StagedSource)
			if e != nil {
				return e
			}
			b, e := transaction.DigestTarget(target.Kind, target.StagedSource)
			if e != nil {
				return e
			}
			if a != b {
				return fmt.Errorf("divergent group target: %s", target.LivePath)
			}
			continue
		}
		g.targets[target.LivePath] = target
	}
	for _, action := range plan.Opaque {
		for _, prior := range g.opaque {
			if prior.LivePath == action.LivePath {
				return fmt.Errorf("duplicate group opaque target")
			}
		}
		g.opaque = append(g.opaque, action)
	}
	for path, digest := range plan.ReadInputs {
		if prior, ok := g.inputs[path]; ok && prior != digest {
			return fmt.Errorf("group read input changed")
		}
		g.inputs[path] = digest
	}
	for _, dir := range plan.Directories {
		if dir.EmptyMuseHome != "" {
			g.emptyMuse[dir.LivePath] = dir.EmptyMuseHome
		}
	}
	g.publications = append(g.publications, pub)
	pub.deferred = true
	return nil
}

// publishManagedRebind is the central lock-publication owner for materialized
// managed homes. It plans every adapter (including inactive homes) from L0,
// preserves captured authority, then commits L1 and all homes in one journal.
// No current/scoped selection can shrink that materialized target inventory.
func (op *operation) publishManagedRebind(home string, files map[string][]byte, removals []string, boundaries *staging.Boundaries, validate func() error) (_ bool, resultErr error) {
	profiles := []string{}
	for path := range files {
		if filepath.Base(path) == "lock.json" && filepath.Dir(filepath.Dir(path)) == ProfilesDir(home) {
			profiles = append(profiles, filepath.Base(filepath.Dir(path)))
		}
	}
	sort.Strings(profiles)
	if len(profiles) == 0 {
		return false, nil
	}
	// Exact no-op lock publication must not manufacture a partial unit or
	// rewrite home generations. Other record changes still use the full group.
	if len(removals) == 0 && op.retained == nil {
		unchanged := true
		for path, payload := range files {
			state, e := stateread.ReadRegularFile(path)
			if e != nil {
				return true, e
			}
			if state.Kind != stateread.KindPresent || !bytes.Equal(bytes.TrimSpace(state.Bytes), bytes.TrimSpace(payload)) {
				unchanged = false
			}
		}
		if unchanged {
			if validate != nil {
				return true, validate()
			}
			return true, nil
		}
	}
	g := op.publicationGroup
	collecting := g != nil
	if g == nil {
		g = newHomePublicationGroup()
	}
	if !collecting {
		defer finishPublicationCleanup(&resultErr, g.close)
	}
	for _, profile := range profiles {
		old, oldHash, err := readLock(home, profile)
		if isStateAbsent(err) {
			// A new installed namespace has no home binding to rebind. The
			// ordinary publisher still applies complete actual effect admission;
			// an existing orphan marker cannot pass the common authority scan.
			continue
		}
		if err != nil {
			return true, err
		}
		if old == nil {
			continue
		}
		next, err := contextlock.Parse(files[lockPath(home, profile)])
		if err != nil {
			return true, err
		}
		canonical, err := next.Canonical()
		if err != nil {
			return true, err
		}
		nextHash := contextlock.HashBytes(canonical)
		if nextHash == oldHash && op.retained == nil {
			unchanged := true
			for path, payload := range files {
				if filepath.Dir(path) == ProfileDir(home, profile) {
					state, e := stateread.ReadRegularFile(path)
					if e != nil {
						return true, e
					}
					if state.Kind != stateread.KindPresent || !bytes.Equal(bytes.TrimSpace(state.Bytes), bytes.TrimSpace(payload)) {
						unchanged = false
					}
				}
			}
			if unchanged {
				continue
			}
		}
		source, err := readSource(home, profile)
		if err != nil {
			return true, err
		}
		if payload, changing := files[sourcePath(home, profile)]; changing {
			if err := json.Unmarshal(payload, &source); err != nil {
				return true, err
			}
		}
		order, err := contextmaterialize.EmittedOrder(next, op.policy.Precedence())
		if err != nil {
			return true, err
		}
		machine := op.policy.EnvironmentMachine
		for _, adapter := range envregistry.Registry {
			target := ManagedHomeDir(home, profile, adapter.ID)
			prior, err := envmarker.Read(target)
			if err != nil {
				return true, err
			}
			if prior == nil {
				continue
			}
			req := &ResolveRequest{Home: home, Profile: profile, EnvID: adapter.ID, Machine: machine, Policy: op.policy, Repair: true, rebindLock: next, publicationGroup: g}
			if op.retained != nil {
				if err := observeRetainedOrigin(req, prior, op.retained); err != nil {
					return true, err
				}
			} else if err := req.observeOrigin(adapter.ID, prior); err != nil {
				return true, err
			}
			if err := req.preflightOriginCredentials(adapter); err != nil {
				return true, err
			}
			req.originBefore = append([]byte{}, req.origin.raw...)
			binding := req.origin.record.Binding
			binding.LockSHA256 = "sha256:" + strings.TrimPrefix(nextHash, "sha256:")
			rebound, err := envorigin.Rebind(req.origin.raw, req.origin.record.Binding, binding, adapter.ID, home)
			if err != nil {
				return true, err
			}
			record, err := envorigin.Parse(rebound, binding, adapter.ID, home)
			if err != nil {
				return true, err
			}
			req.origin = &originObservation{record: record, raw: rebound}
			plan, err := assembleHomeObserved(req, source, next, nextHash, op.policy.Precedence(), order, adapter, prior, "unknown")
			if err != nil {
				return true, err
			}
			// Existing one-time seeds are not recaptured during a lock transition.
			// XDG seed links retain their own recorded text and do not follow today's
			// caller selectors into another operator seed/data/state namespace.
			seeds := &seedBundle{files: map[string]seedFile{}, xdg: map[string]string{}, claudeInit: adapter.ID == envregistry.ClaudeCode}
			for _, name := range prior.SeedLinks {
				link, err := stateread.Readlink(filepath.Join(plan.parent, name))
				if err != nil {
					return true, err
				}
				if link.Kind == stateread.KindPresent {
					seeds.xdg[name] = link.Target
				}
			}
			recorded := map[string]bool{}
			for _, surface := range prior.Surfaces {
				for _, path := range surface.Paths {
					recorded[path] = true
				}
			}
			if err := publishManagedHome(op, req, plan, seeds, recorded, prior, false); err != nil {
				return true, err
			}
		}
	}
	if g.unit == nil {
		return false, nil
	}
	if collecting {
		return true, nil
	}
	return true, op.commitHomeGroup(home, g, files, removals, boundaries, validate)
}

func (op *operation) commitHomeGroup(home string, g *homePublicationGroup, files map[string][]byte, removals []string, boundaries *staging.Boundaries, validate func() error) (resultErr error) {
	if g == nil || g.unit == nil {
		return nil
	}
	recordPub, err := newHomePublication(op)
	if err != nil {
		return err
	}
	defer finishPublicationCleanup(&resultErr, recordPub.close)
	for path, payload := range files {
		class := "profile-record"
		if isActivationPath(home, path) {
			class = "activation"
		}
		if err := recordPub.file(path, class, payload, 0600, true); err != nil {
			return err
		}
	}
	for _, path := range removals {
		class := "profile-record"
		if isActivationPath(home, path) {
			class = "activation"
		}
		if err := recordPub.remove(path, class); err != nil {
			return err
		}
	}
	for path, target := range recordPub.targets {
		if _, duplicate := g.targets[path]; duplicate {
			return fmt.Errorf("duplicate group record target")
		}
		g.targets[path] = target
		delete(g.inputs, path)
	}
	for i := range g.unit.Profiles {
		p := &g.unit.Profiles[i]
		if payload, changing := files[sourcePath(home, p.Name)]; changing {
			next := envorigin.Digest(payload)
			if next != p.Source {
				p.OldSource = p.Source
				p.Source = next
			}
		}
	}
	plan := transaction.Plan{TransactionID: g.unit.ID, ProjectIdentity: "managed-home-unit", Unit: g.unit, ReadInputs: g.inputs}
	paths := []string{}
	targets := []staging.Target{}
	for path, target := range g.targets {
		plan.Targets = append(plan.Targets, target)
		paths = append(paths, path)
		targets = append(targets, staging.Target{LivePath: path, Kind: staging.KindEntry})
	}
	plan.Opaque = append([]transaction.OpaqueTarget(nil), g.opaque...)
	for _, action := range plan.Opaque {
		targets = append(targets, staging.Target{LivePath: action.LivePath, Kind: staging.KindEntry})
	}
	plan.Directories, err = op.engine.PublicationDirectories(paths...)
	if err != nil {
		return err
	}
	seenDirs := map[string]bool{}
	for _, dir := range plan.Directories {
		seenDirs[dir.LivePath] = true
	}
	for _, native := range g.unit.Native {
		var nativePaths []string
		for path := range g.targets {
			if path == native.Root || strings.HasPrefix(path, native.Root+string(filepath.Separator)) {
				nativePaths = append(nativePaths, path)
			}
		}
		dirs, e := op.engine.NativePublicationDirectories(native.Root, nativePaths...)
		if e != nil {
			return e
		}
		for _, dir := range dirs {
			if !seenDirs[dir.LivePath] {
				plan.Directories = append(plan.Directories, dir)
				seenDirs[dir.LivePath] = true
			}
		}
	}
	for i := range plan.Directories {
		plan.Directories[i].EmptyMuseHome = g.emptyMuse[plan.Directories[i].LivePath]
	}
	snapshot, err := (staging.Plan{Targets: targets}).Snapshot(nil)
	if err != nil {
		return err
	}
	proof, err := snapshot.Durable()
	if err != nil {
		return err
	}
	plan.BoundaryProof = &proof
	if validate != nil {
		if err := validate(); err != nil {
			return err
		}
	}
	if boundaries != nil {
		plan.BoundaryCheck = func(_ int, path string) error {
			if _, ok := boundaries.Durable.Targets[path]; ok {
				return staging.RecheckDurableOne(staging.Target{LivePath: path, Kind: staging.KindEntry}, boundaries.Durable)
			}
			return nil
		}
	}
	if op.publishPlan != nil {
		return op.publishPlan(plan)
	}
	if _, err := op.engine.Prepare(op.lock, plan); err != nil {
		return err
	}
	if err := op.engine.Commit(op.lock, plan.TransactionID); err != nil {
		return err
	}
	return nil
}
