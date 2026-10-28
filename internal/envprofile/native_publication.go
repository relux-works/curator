package envprofile

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

func newHomePublicationGroup() *homePublicationGroup {
	return &homePublicationGroup{targets: map[string]transaction.Target{}, inputs: map[string]string{}, emptyMuse: map[string]string{}}
}

// planNativePublication runs after the read-only ownership inventory. Only
// selected context/skills entries and their fixed backup namespace are enlisted;
// native credentials, settings and sessions are not read, copied or digested.
// The group can join this exact custody set to a managed home before any write.
func planNativePublication(op *operation, home, native, profile string, source Source, lock *contextlock.Lock, hash string, precedence contextmaterialize.Precedence, order []contextlock.Member, skills []skillOf, adapter Adapter, prior *envmarker.Marker, recorded, want map[string]bool, skillLinks map[string]string, skillHashes map[string][]byte, skillPaths []string, document []byte, written bool) (resultErr error) {
	if err := envorigin.ValidateExternal(home, native); err != nil {
		return err
	}
	pub, err := newHomePublication(op)
	if err != nil {
		return err
	}
	defer func() {
		if !pub.deferred {
			_ = pub.close()
		}
	}()
	pub.readScope.Native = []envunit.Native{{Adapter: adapter.ID, Root: native}}
	backup := map[string]bool{}
	for rel := range want {
		state, e := stateread.Lstat(filepath.Join(native, filepath.FromSlash(rel)))
		if e != nil {
			return e
		}
		if state.Kind == stateread.KindPresent {
			backup[rel] = true
		}
	}
	if err := pub.backup(native, backup); err != nil {
		return err
	}
	for rel := range recorded {
		if !want[rel] {
			if err := pub.remove(filepath.Join(native, filepath.FromSlash(rel)), "owned-bytes"); err != nil {
				return err
			}
		}
	}
	surface := envmarker.Surface{Paths: []string{}, Form: "monolithic"}
	copies := []envmarker.Copy{}
	if written {
		if adapter.ID == "claude_code" {
			if err := pub.file(filepath.Join(native, adapter.Target), "owned-bytes", document, 0644, false); err != nil {
				return err
			}
			copies = append(copies, envmarker.Copy{Path: adapter.Target, Reason: envmarker.ReasonClaudeCodeRootContext})
		} else {
			// This manager-owned document is a target of the same unit as its alias.
			store := filepath.Join(ProfileDir(home, profile), "rendered", adapter.ID, adapter.Target)
			if err := pub.file(store, "owned-bytes", document, managedAuthoredFileMode, false); err != nil {
				return err
			}
			copied, err := pub.projectionLink(filepath.Join(native, adapter.Target), store, filepath.Join(native, envmarker.Name), envmarker.SurfaceRootContext, adapter.Target)
			if err != nil {
				return err
			}
			if copied {
				copies = append(copies, envmarker.Copy{Path: adapter.Target, Reason: envmarker.ReasonSymlinkFallback})
			}
		}
		surface.Paths = []string{adapter.Target}
		surface.ContentSHA256, _ = contextmaterialize.SurfaceHashWithVersion(map[string][]byte{adapter.Target: document}, lock.ContentHashVersion())
	}
	surface.Copies = &copies
	marker := &envmarker.Marker{Version: 1, Profile: envmarker.Profile{Name: profile, Root: lock.Root, Kind: markerKind(source), LockSHA256: strings.TrimPrefix(hash, "sha256:"), Source: markerSource(source), Requirement: markerRequirement(source), Directory: source.Directory, SourcePath: markerSourcePath(source), ImportedFromNative: source.ImportedFromNative}, Precedence: envmarker.Precedence{Winner: precedence.Winner, Placement: precedence.Placement}, Mode: envmarker.ModeLinked, Surfaces: map[string]envmarker.Surface{}}
	if lock.ContentHashVersion() == 2 {
		marker.Version = envmarker.VersionV3
		marker.HashVersion = 2
	}
	for _, member := range order {
		entry := envmarker.Member{Name: member.Name, Version: member.Version, Weight: member.Weight, Overlay: member.Overlay}
		if member.Commit != "" {
			entry.Commit = member.Commit
		} else {
			entry.StateSHA256 = member.StateHash
		}
		marker.Members = append(marker.Members, entry)
	}
	if written {
		marker.Surfaces[envmarker.SurfaceRootContext] = surface
	}
	for _, skill := range skills {
		if err := publishSkillProjection(op, home, skill); err != nil {
			return err
		}
	}
	sort.Strings(skillPaths)
	skillCopies := []envmarker.Copy{}
	for _, rel := range skillPaths {
		if recordedSkillFallback(prior, rel) {
			if err := pub.tree(filepath.Join(native, filepath.FromSlash(rel)), skillLinks[rel]); err != nil {
				return err
			}
			skillCopies = append(skillCopies, envmarker.Copy{Path: rel, Reason: envmarker.ReasonSymlinkFallback})
		} else {
			copied, err := pub.projectionLink(filepath.Join(native, filepath.FromSlash(rel)), skillLinks[rel], filepath.Join(native, envmarker.Name), envmarker.SurfaceSkills, rel)
			if err != nil {
				return err
			}
			if copied {
				skillCopies = append(skillCopies, envmarker.Copy{Path: rel, Reason: envmarker.ReasonSymlinkFallback})
			}
		}
	}
	if len(skillPaths) > 0 {
		digest, _ := contextmaterialize.SurfaceHashWithVersion(skillHashes, lock.ContentHashVersion())
		marker.Surfaces[envmarker.SurfaceSkills] = envmarker.Surface{Paths: skillPaths, ContentSHA256: digest, Copies: &skillCopies}
	}
	raw, err := marker.Marshal()
	if err != nil {
		return err
	}
	if err := pub.file(filepath.Join(native, envmarker.Name), "native-marker", raw, 0600, true); err != nil {
		return err
	}
	identity, err := envunit.ObserveManager(home)
	if err != nil {
		return err
	}
	canonical, err := lock.Canonical()
	if err != nil {
		return err
	}
	sourcePayload := []byte(nil)
	oldRecord := envunit.RecordImage{Kind: "absent"}
	if op.initialProfile != nil {
		sourcePayload = op.initialProfile[sourcePath(home, profile)]
		if len(sourcePayload) == 0 {
			return fmt.Errorf("%s: first-use source capture missing", envorigin.DiagInvalid)
		}
	} else {
		sourceState, err := stateread.ReadRegularFile(sourcePath(home, profile))
		if err != nil {
			return err
		}
		if sourceState.Kind != stateread.KindPresent {
			return fmt.Errorf("%s: native installed source absent", envorigin.DiagInvalid)
		}
		sourcePayload = sourceState.Bytes
		liveLock, _, err := readLock(home, profile)
		if err != nil {
			return err
		}
		liveCanonical, err := liveLock.Canonical()
		if err != nil {
			return err
		}
		oldRecord = envunit.RecordImage{Kind: "file", Bytes: liveCanonical, Digest: envorigin.Digest(liveCanonical)}
	}
	record := envunit.RecordImage{Kind: "file", Bytes: canonical, Digest: envorigin.Digest(canonical)}
	action := envunit.Maintain
	if oldRecord.Digest != record.Digest {
		action = envunit.Rebind
	}
	var closure []string
	for _, member := range lock.Members {
		raw, e := json.Marshal(member)
		if e != nil {
			return e
		}
		closure = append(closure, envorigin.Digest(raw))
	}
	directory, err := envunit.ObserveDirectory(native)
	if err != nil {
		return err
	}
	before, err := stateread.ReadRegularFile(filepath.Join(native, envmarker.Name))
	if err != nil {
		return err
	}
	var paths []string
	for path := range pub.targets {
		if strings.HasPrefix(path, native+string(filepath.Separator)) {
			rel, e := filepath.Rel(native, path)
			if e != nil {
				return e
			}
			paths = append(paths, filepath.ToSlash(rel))
		}
	}
	sort.Strings(paths)
	unit := &envunit.Unit{Record: envunit.Version, ID: newTransactionID(), Manager: identity, Action: action, Homes: []envunit.Home{}, Profiles: []envunit.ProfileTransition{{Name: profile, Source: envorigin.Digest(sourcePayload), Closure: closure, OldLock: oldRecord, NewLock: record}}, Native: []envunit.Native{{Profile: profile, Adapter: adapter.ID, Root: native, DirectoryBefore: directory, DirectoryAfter: directory, MarkerBefore: before.Bytes, MarkerAfter: raw, Paths: paths}}}
	plan := transaction.Plan{TransactionID: unit.ID, ProjectIdentity: "managed-home-unit", Unit: unit, ReadInputs: map[string]string{}}
	for _, path := range []string{sourcePath(home, profile), lockPath(home, profile)} {
		digest, e := transaction.DigestPath(path)
		if e != nil {
			return e
		}
		plan.ReadInputs[path] = digest
	}
	plan.Opaque = append([]transaction.OpaqueTarget(nil), pub.opaque...)
	for _, target := range pub.targets {
		plan.Targets = append(plan.Targets, target)
	}
	group := op.publicationGroup
	own := group == nil
	if own {
		group = newHomePublicationGroup()
		defer group.close()
	}
	if err := group.add(plan, pub); err != nil {
		return err
	}
	if own {
		return op.commitHomeGroup(home, group, nil, nil, nil, nil)
	}
	return nil
}
