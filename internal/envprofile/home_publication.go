package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/linkcap"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/staging"
	"github.com/relux-works/curator/internal/transaction"
)

// homePublication stages only enumerated manager-owned entries. It never
// snapshots the home tree: credentials are link text, and unlisted tool state
// (including sessions) is neither read nor copied. Parent metadata is a separate
// durable action, so process-local undo is not the restart recovery contract.
type homePublication struct {
	readScope   envunit.Unit
	op          *operation
	scratch     string
	targets     map[string]transaction.Target
	opaque      []transaction.OpaqueTarget
	directories []string
	unit        *envunit.Unit
	deferred    bool
}

func newHomePublication(op *operation) (*homePublication, error) {
	scratch, err := os.MkdirTemp("", "curator-home-unit-*")
	if err != nil {
		return nil, err
	}
	if err := privatedir.Protect(scratch); err != nil {
		_ = os.Remove(scratch)
		return nil, err
	}
	return &homePublication{op: op, scratch: scratch, targets: map[string]transaction.Target{}, readScope: envunit.Unit{Manager: envunit.Manager{Path: op.engine.Home()}}}, nil
}
func (p *homePublication) close() error { return os.RemoveAll(p.scratch) }
func (p *homePublication) entry(path, class, source string, private bool) error {
	if _, duplicate := p.targets[path]; duplicate {
		return fmt.Errorf("duplicate home publication target: %s", path)
	}
	// A prior copied skill is a recursive projection even when this operation
	// removes it or returns it to a link. Classify before computing its digest.
	if class == "owned-bytes" {
		for _, candidate := range []string{path, source} {
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
	if err := p.readScope.ValidateReadScope(class, path); err != nil {
		return err
	}
	for _, candidate := range []string{path, source} {
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
	before, err := transaction.DigestTarget(kind, path)
	if err != nil {
		return err
	}
	p.targets[path] = transaction.Target{Class: class, Identifier: path, Kind: kind, LivePath: path, StagedSource: source, PreimageDigest: before, PrivateFile: private}
	return nil
}
func (p *homePublication) file(path, class string, payload []byte, mode os.FileMode, private bool) error {
	source := filepath.Join(p.scratch, strconv.Itoa(len(p.targets)))
	if err := os.WriteFile(source, payload, mode); err != nil {
		return err
	}
	return p.entry(path, class, source, private)
}
func (p *homePublication) link(path, class, target string) error {
	source := filepath.Join(p.scratch, strconv.Itoa(len(p.targets)))
	if err := os.Symlink(target, source); err != nil {
		return err
	}
	return p.entry(path, class, source, false)
}

// projectionLink registers only a context/skill alternative. Credential links
// deliberately use link(), and can never become copied auth material.
func (p *homePublication) projectionLink(path, target, markerPath, surface, rel string) (bool, error) {
	source := target
	if prepared, ok := p.targets[target]; ok {
		source = prepared.StagedSource
	}
	info, err := os.Lstat(source)
	if err != nil {
		return false, err
	}
	class := "owned-bytes"
	if info.IsDir() {
		class = "owned-tree"
	}
	if err := p.link(path, "owned-bytes", target); err != nil {
		if !linkcap.Unavailable(err) {
			return false, err
		}
		return true, p.entry(path, class, source, false)
	}
	entry := p.targets[path]
	entry.LinkFallback = &transaction.LinkFallback{Source: source, Class: class, MarkerPath: markerPath, Surface: surface, RelativePath: rel}
	p.targets[path] = entry
	return false, nil
}

func (p *homePublication) remove(path, class string) error { return p.entry(path, class, "", false) }

func (p *homePublication) tree(path, source string) error {
	name := strconv.Itoa(len(p.targets))
	if err := copyTree(source, p.scratch, name); err != nil {
		return err
	}
	return p.entry(path, "owned-tree", filepath.Join(p.scratch, name), false)
}

// backup plans the next immutable generation from only the exact noncredential
// surface replacement set. Each entry and retention removal joins the same
// journal as the home, so refusal cannot leave a generation or lose an old one.
func (p *homePublication) backup(home string, paths map[string]bool) error {
	if len(paths) == 0 {
		return nil
	}
	// Check the entire no-follow parent route before listing any generation.
	// A broken/link backup parent is not an absent backup namespace.
	probe, err := managedPath(home, ".agent-environment-backup/.inventory", false)
	if err != nil {
		return err
	}
	base := filepath.Dir(probe)
	image, err := envunit.ObserveDirectory(base)
	if err != nil {
		return err
	}
	var entries []os.DirEntry
	if image.Kind != "absent" {
		entries, err = os.ReadDir(base)
		if err != nil {
			return err
		}
		after, err := envunit.ObserveDirectory(base)
		if err != nil || image != after {
			return fmt.Errorf("%s: backup namespace changed", envorigin.DiagInvalid)
		}
	}
	highest := 0
	var generations []int
	for _, entry := range entries {
		n, err := strconv.Atoi(entry.Name())
		if err != nil || n <= 0 || strconv.Itoa(n) != entry.Name() || !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: unknown backup generation entry", envorigin.DiagInvalid)
		}
		if err == nil && n > 0 {
			generations = append(generations, n)
			if n > highest {
				highest = n
			}
		}
	}
	next := highest + 1
	generation := filepath.Join(base, strconv.Itoa(next))
	if _, err := os.Lstat(generation); err == nil {
		return fmt.Errorf("%s: next generation exists", DiagBackupExists)
	} else if !os.IsNotExist(err) {
		return err
	}
	for rel := range paths {
		source := filepath.Join(home, filepath.FromSlash(rel))
		info, err := os.Lstat(source)
		if err != nil {
			return err
		}
		destination := filepath.Join(generation, filepath.FromSlash(rel))
		if info.Mode()&os.ModeSymlink != 0 {
			text, err := os.Readlink(source)
			if err != nil {
				return err
			}
			if err := p.link(destination, "owned-bytes", text); err != nil {
				return err
			}
		} else if info.Mode().IsRegular() {
			payload, present, err := readManagedRegular(home, rel)
			if err != nil {
				return err
			}
			if !present {
				return fmt.Errorf("backup input disappeared")
			}
			if err := p.file(destination, "owned-bytes", payload, info.Mode().Perm(), false); err != nil {
				return err
			}
		} else if info.IsDir() {
			if err := p.tree(destination, source); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s: unowned backup source type", DiagUnmanagedConflict)
		}
	}
	sort.Ints(generations)
	for len(generations)+1 > BackupRetention {
		old := generations[0]
		generations = generations[1:]
		path := filepath.Join(base, strconv.Itoa(old))
		before, err := envunit.ObserveDirectory(path)
		if err != nil {
			return err
		}
		p.opaque = append(p.opaque, transaction.OpaqueTarget{Class: envunit.BackupQuarantine, LivePath: path, Before: before})
	}
	return nil
}
func (p *homePublication) publish(req *ResolveRequest, plan *homePlan, prior *envmarker.Marker, marker []byte) error {
	transactionPlan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "managed-home-unit", ReadInputs: map[string]string{}}
	var paths []string
	var boundaryTargets []staging.Target
	for _, target := range p.targets {
		transactionPlan.Targets = append(transactionPlan.Targets, target)
		paths = append(paths, target.LivePath)
		boundaryTargets = append(boundaryTargets, staging.Target{LivePath: target.LivePath, Kind: staging.KindEntry})
	}
	transactionPlan.Opaque = append([]transaction.OpaqueTarget(nil), p.opaque...)
	for _, action := range p.opaque {
		boundaryTargets = append(boundaryTargets, staging.Target{LivePath: action.LivePath, Kind: staging.KindEntry})
	}
	// Empty Muse directories have durable existence/protection too.
	for _, dir := range p.directories {
		paths = append(paths, filepath.Join(dir, envmarker.Name))
	}
	var err error
	transactionPlan.Directories, err = p.op.engine.PublicationDirectories(paths...)
	if err != nil {
		return err
	}
	for index := range transactionPlan.Directories {
		for _, empty := range p.directories {
			if transactionPlan.Directories[index].LivePath == empty {
				transactionPlan.Directories[index].EmptyMuseHome = plan.homeDir
			}
		}
	}
	snapshot, err := (staging.Plan{Targets: boundaryTargets}).Snapshot(nil)
	if err != nil {
		return err
	}
	proof, err := snapshot.Durable()
	if err != nil {
		return err
	}
	transactionPlan.BoundaryProof = &proof
	for _, path := range []string{sourcePath(req.Home, req.Profile), lockPath(req.Home, req.Profile)} {
		digest, err := transaction.DigestPath(path)
		if err != nil {
			return err
		}
		transactionPlan.ReadInputs[path] = digest
	}
	if p.unit != nil {
		transactionPlan.Unit = p.unit
		transactionPlan.TransactionID = p.unit.ID
	} else {
		transactionPlan.Unit, err = req.homeUnit(transactionPlan.TransactionID, plan, prior, marker)
		if err != nil {
			return err
		}
	}
	if req.publicationGroup != nil {
		return req.publicationGroup.add(transactionPlan, p)
	}
	if _, err := p.op.engine.Prepare(p.op.lock, transactionPlan); err != nil {
		return err
	}
	if err := p.op.engine.Commit(p.op.lock, transactionPlan.TransactionID); err != nil {
		return err
	}
	return nil
}

func publishManagedHome(op *operation, req *ResolveRequest, plan *homePlan, seeds *seedBundle, recorded map[string]bool, prior *envmarker.Marker, provisioned bool) (resultErr error) {
	if err := req.preflightOriginCredentials(plan.adapter); err != nil {
		return err
	}
	pub, err := newHomePublication(op)
	if err != nil {
		return err
	}
	defer func() {
		if !pub.deferred {
			resultErr = errors.Join(resultErr, pub.close())
		}
	}()
	pub.readScope.Homes = []envunit.Home{{Adapter: plan.adapter.ID, Target: plan.homeDir}}
	root := EnvRoot(req.Home)
	homeRel, err := managedRelative(root, plan.homeDir)
	if err != nil {
		return err
	}
	if seeds.claudeInit || plan.adapter.ID == "claude_code" {
		seeds.claudePrepared, err = refreshClaudeSeed(op.claudePrepared, root, homeRel, req.LaunchDir, plan.form)
		if err != nil {
			return op.recoveryRefusal(err)
		}
	}
	want := map[string]bool{}
	for path := range plan.copies {
		want[path] = true
	}
	for path := range plan.links {
		want[path] = true
	}
	if seeds.claudeInit {
		want[".claude.json"] = true
	}
	for seed := range seeds.files {
		want[seed] = true
	}
	unmanaged := unmanagedPlanTargets(plan, recorded)
	if len(unmanaged) > 0 {
		if _, err := inventoryUnmanaged(plan.homeDir, unmanaged, plan.storeRoot); err != nil && !req.Policy.Takeover {
			return err
		}
	}
	if err := preflightManagedWriteTargets(root, homeRel, want); err != nil {
		return err
	}
	// Immutable skill projections remain independent content-addressed artifacts.
	// Their own publisher must journal its parent and tree publication before a
	// home unit can reference them. They confer no home/credential authority.
	for _, skill := range plan.skillProjections {
		if err := publishSkillProjection(op, req.Home, skill); err != nil {
			return err
		}
	}
	for path, payload := range plan.docs {
		if err := pub.file(path, "owned-bytes", payload, managedAuthoredFileMode, false); err != nil {
			return err
		}
	}
	replacing := map[string]bool{}
	for path := range want {
		full, err := plannedManagedPath(root, managedJoin(homeRel, path))
		if err != nil {
			return err
		}
		info, err := os.Lstat(full)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 && provisioned && req.Policy.Takeover && !recorded[path] {
			replacing[path] = true
		}
	}
	if err := pub.backup(plan.homeDir, replacing); err != nil {
		return err
	}
	for path := range recorded {
		if !want[path] && !isSeedPath(plan, path) {
			if err := pub.remove(filepath.Join(plan.homeDir, filepath.FromSlash(path)), "owned-bytes"); err != nil {
				return err
			}
		}
	}
	for path, payload := range plan.copies {
		if err := pub.file(filepath.Join(plan.homeDir, filepath.FromSlash(path)), "owned-bytes", payload, managedAuthoredFileMode, false); err != nil {
			return err
		}
	}
	for path, target := range plan.links {
		if recordedSkillFallback(prior, path) {
			if err := pub.tree(filepath.Join(plan.homeDir, filepath.FromSlash(path)), target); err != nil {
				return err
			}
			recordLinkFallback(plan, path)
			continue
		}
		surfaceName := ""
		for key, surface := range plan.marker.Surfaces {
			for _, declared := range surface.Paths {
				if declared == path {
					surfaceName = key
				}
			}
		}
		if surfaceName == "" {
			return fmt.Errorf("projection link has no declared surface")
		}
		copied, err := pub.projectionLink(filepath.Join(plan.homeDir, filepath.FromSlash(path)), target, filepath.Join(plan.homeDir, envmarker.Name), surfaceName, path)
		if err != nil {
			return err
		}
		if copied {
			recordLinkFallback(plan, path)
		}
	}
	if provisioned {
		for path, seed := range seeds.files {
			if err := pub.file(filepath.Join(plan.homeDir, filepath.FromSlash(path)), "owned-bytes", seed.payload, seed.mode, false); err != nil {
				return err
			}
		}
	}
	if plan.adapter.ID == "muse" {
		for _, dir := range []string{"config/muse", "data/muse", "state", "cache"} {
			pub.directories = append(pub.directories, filepath.Join(plan.homeDir, dir))
		}
	}
	provenance := "repaired"
	if provisioned {
		provenance = "provisioned"
	}
	marker, err := finalizeMarker(op, req, plan, seeds, prior, provenance, pub)
	if err != nil {
		return err
	}
	payload, err := marker.Marshal()
	if err != nil {
		return err
	}
	if prior != nil && prior.Version == envmarker.VersionV1 && legacyMarkerProjectionMatches(prior, marker) {
		payload, err = os.ReadFile(filepath.Join(plan.homeDir, envmarker.Name))
		if err != nil {
			return err
		}
	}
	if err := pub.file(filepath.Join(plan.homeDir, envmarker.Name), "marker", payload, 0600, true); err != nil {
		return err
	}
	if req.origin == nil {
		return fmt.Errorf("%s: authority absent at publication", envorigin.DiagInvalid)
	}
	if err := pub.file(filepath.Join(plan.homeDir, envorigin.Name), "origin", req.origin.raw, 0600, true); err != nil {
		return err
	}
	return pub.publish(req, plan, prior, payload)
}
