package envunit

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/staging"
)

func fixtureUnit(t *testing.T) *Unit {
	t.Helper()
	root := t.TempDir()
	manager := filepath.Join(root, "manager")
	if err := privatedir.Make(manager); err != nil {
		t.Fatal(err)
	}
	m, err := ObserveManager(manager)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := (&contextlock.Lock{SchemaVersion: contextlock.SchemaVersion2, HashVersion: 2, Root: "fixture", Members: []contextlock.Member{{Kind: contextlock.KindContext, Name: "fixture", Version: "1.0.0", Commit: strings.Repeat("b", 40), Source: "https://example.com/fixture", RequiredBy: []string{}}}}).Canonical()
	if err != nil {
		t.Fatal(err)
	}
	hash := envorigin.Digest(lock)
	target := envorigin.ManagedTarget(manager, "p", "codex_cli")
	r := &envorigin.Record{Record: envorigin.Version, Adapter: "codex_cli", OperatorHome: root, Selector: envorigin.Selector{Name: "CODEX_HOME", State: "absent"}, NativeRoot: filepath.Join(root, "a"), Capture: envorigin.Capture{Method: "provisioning", Authority: "native-selection", Time: "2026-10-28T00:00:00Z"}, Binding: envorigin.Binding{Profile: "p", LockSHA256: hash, ManagedTarget: target}}
	origin, err := r.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	passthrough := []envmarker.Passthrough{}
	seeds := []string{}
	marker := &envmarker.Marker{Version: 2, Profile: envmarker.Profile{Name: "p", Root: "fixture", Kind: "git", LockSHA256: strings.TrimPrefix(hash, "sha256:"), Source: "https://example.com/fixture", Requirement: &envmarker.Requirement{Range: "^1.0.0"}}, Members: []envmarker.Member{{Name: "fixture", Version: "1.0.0", Commit: strings.Repeat("b", 40), Weight: 0}}, Precedence: envmarker.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Mode: envmarker.ModeManagedHome, Surfaces: map[string]envmarker.Surface{}, Passthrough: &passthrough, Seeds: &seeds}
	mark, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	absent := Image{Kind: "absent"}
	// Descriptor models use actual target-native owner/protection metadata.
	// The private template is outside the modeled live namespace and its
	// payload is noncredential fixture data; no account lookup is involved.
	template := filepath.Join(root, "private-template")
	entry, err := privatedir.Create(template, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = entry.WriteString("descriptor fixture"); err != nil {
		t.Fatal(err)
	}
	if err = entry.Close(); err != nil {
		t.Fatal(err)
	}
	fileImage, err := ObserveEntry(template, false)
	if err != nil {
		t.Fatal(err)
	}
	directoryImage, err := ObserveDirectory(manager)
	if err != nil {
		t.Fatal(err)
	}
	file := func(data []byte) Image {
		image := fileImage
		image.Digest = envorigin.Digest(data)
		return image
	}
	u := &Unit{Record: Version, ID: "fixture-unit", Manager: m, Action: Provision, Profiles: []ProfileTransition{{Name: "p", Source: envorigin.Digest([]byte("source")), Closure: []string{envorigin.Digest([]byte("closure"))}, OldLock: RecordImage{Kind: "file", Bytes: lock, Digest: hash}, NewLock: RecordImage{Kind: "file", Bytes: lock, Digest: hash}}}, Homes: []Home{{Profile: "p", Adapter: "codex_cli", Mode: "managed-home", Target: target, Before: HomeImage{Kind: "absent"}, After: HomeImage{Kind: "recorded", Marker: mark, Origin: origin}}}, Targets: []Target{{Class: "origin", Path: filepath.Join(target, envorigin.Name), Before: absent, After: file(origin)}, {Class: "marker", Path: filepath.Join(target, envmarker.Name), Before: absent, After: file(mark)}}}
	u.Homes[0].DirectoryBefore = absent
	u.Homes[0].Root = target
	u.Homes[0].DirectoryAfter = directoryImage
	u.Targets = append(u.Targets, Target{Class: "home-directory", Path: target, Before: absent, After: u.Homes[0].DirectoryAfter})
	var plan staging.Plan
	for _, target := range u.Targets {
		plan.ReplaceEntry(target.Class, target.Path, target.Path, "unused-stage")
	}
	snapshot, err := plan.Snapshot(nil)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := snapshot.Durable()
	if err != nil {
		t.Fatal(err)
	}
	u.BoundaryProof = proof
	return u
}

// Proves declarations bind exact capture/profile/target sets and a strict closed
// transport. Engine/live-prefix admission is tested separately, not by this fixture.
func TestProvisioningOriginUnitDescriptor(t *testing.T) {
	u := fixtureUnit(t)
	if err := u.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, err := u.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(raw, u.Manager); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Unit){
		func(v *Unit) { v.Action = "unknown" },
		func(v *Unit) { v.Homes[0].Target = filepath.Join(v.Manager.Path, "environments", "other", "codex_cli") },
		func(v *Unit) { v.Targets = v.Targets[:1] },
		func(v *Unit) { v.Targets = append(v.Targets, v.Targets[0]) },
		func(v *Unit) {
			v.Homes[0].After.Origin = bytes.Replace(v.Homes[0].After.Origin, []byte(`"native-selection"`), []byte(`"explicit-operator-input"`), 1)
		},
		func(v *Unit) { v.Profiles[0].NewLock.Bytes = []byte(`{"different":true}`) },
	} {
		var v Unit
		_ = json.Unmarshal(raw, &v)
		mutate(&v)
		if err := v.Validate(); err == nil {
			t.Fatalf("accepted invalid unit %+v", v)
		}
	}
	if _, err := Parse(bytes.Replace(raw, []byte(`"action":`), []byte(`"Action":`), 1), u.Manager); err == nil {
		t.Fatal("case-folded authority accepted")
	}
}

// Proves reachable target prefixes differ from arbitrary mixed states. It does
// not allow a journal's existence to supply an admission verdict.
func TestProvisioningOriginUnitReachablePrefix(t *testing.T) {
	u := fixtureUnit(t)
	states := make([]Progress, len(u.Targets))
	for i, target := range u.Targets {
		states[i] = Progress{State: "pending", Live: target.Before}
	}
	if err := u.ValidateProgress("prepared", states); err != nil {
		t.Fatal(err)
	}
	states[0] = Progress{State: "committed", Live: u.Targets[0].After}
	if err := u.ValidateProgress("committing", states); err != nil {
		t.Fatal(err)
	}
	states[0] = Progress{State: "pending", Live: u.Targets[0].Before}
	states[1] = Progress{State: "committed", Live: u.Targets[1].After}
	if err := u.ValidateProgress("committing", states); err == nil {
		t.Fatal("non-prefix mixture accepted")
	}
	states[0] = Progress{State: "committed", Live: u.Targets[0].After}
	for i := 1; i < len(states); i++ {
		states[i] = Progress{State: "committed", Live: u.Targets[i].After}
	}
	if err := u.ValidateProgress("cleanup", states); err != nil {
		t.Fatal(err)
	}
	states[0].Live.Digest = envorigin.Digest([]byte("substitution"))
	if err := u.ValidateProgress("cleanup", states); err == nil {
		t.Fatal("substituted committed target accepted")
	}
}

// Proves the descriptor can express genuine retained-L1/old-home-L0 takeover
// without treating a generic previous lock as authority or folding in L2.
// Receipt/live-journal validation and actual takeover remain separate controls.
func TestProvisioningOriginUnitRetainedLocks(t *testing.T) {
	u := fixtureUnit(t)
	u.Action = Takeover
	h := &u.Homes[0]
	h.Before = h.After
	h.DirectoryBefore = h.DirectoryAfter
	oldLock := u.Profiles[0].OldLock
	lock, err := contextlock.Parse(oldLock.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	lock.Members[0].Weight++
	newBytes, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	newLock := RecordImage{Kind: "file", Bytes: newBytes, Digest: envorigin.Digest(newBytes)}
	u.Profiles[0].OldLock = newLock
	u.Profiles[0].NewLock = newLock
	marker, err := envmarker.Parse(h.Before.Marker)
	if err != nil {
		t.Fatal(err)
	}
	marker.Profile.LockSHA256 = strings.TrimPrefix(newLock.Digest, "sha256:")
	marker.Members[0].Weight = lock.Members[0].Weight
	h.After.Marker, err = marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	oldBinding := envorigin.Binding{Profile: "p", LockSHA256: oldLock.Digest, ManagedTarget: h.Target}
	nextBinding := oldBinding
	nextBinding.LockSHA256 = newLock.Digest
	h.After.Origin, err = envorigin.Rebind(h.Before.Origin, oldBinding, nextBinding, h.Adapter, u.Manager.Path)
	if err != nil {
		t.Fatal(err)
	}
	for i := range u.Targets {
		target := &u.Targets[i]
		target.Before = target.After
		switch target.Class {
		case "marker":
			target.After.Digest = envorigin.Digest(h.After.Marker)
		case "origin":
			target.After.Digest = envorigin.Digest(h.After.Origin)
		}
	}
	u.LedgerPublication = &LedgerPublication{Path: filepath.Join(u.Manager.Path, ".state", "environment-lifecycle.json"), PreimageDigest: "sha256:" + strings.Repeat("a", 64), Revision: 1, NextRevision: 2, EntryIDs: []string{}}
	u.RetainedEntry = &RetainedEntry{ID: "retained-group", Digest: "sha256:" + strings.Repeat("b", 64), HomeLocks: map[string]RecordImage{"p": oldLock}}
	if err := u.Validate(); err != nil {
		t.Fatalf("genuine L1/old-home-L0 descriptor refused: %v", err)
	}
	u.RetainedEntry.HomeLocks["p"] = newLock
	if err := u.Validate(); err == nil || !strings.Contains(err.Error(), "marker binding") {
		t.Fatalf("invented old-home authority accepted: %v", err)
	}
	u.RetainedEntry.HomeLocks["p"] = oldLock
	lock.Members[0].Weight++
	l2, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	u.Profiles[0].NewLock = RecordImage{Kind: "file", Bytes: l2, Digest: envorigin.Digest(l2)}
	if err := u.Validate(); err == nil || !strings.Contains(err.Error(), "cannot change installed lock") {
		t.Fatalf("takeover also performed L2: %v", err)
	}
}
