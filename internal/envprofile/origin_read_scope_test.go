package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/transaction"
)

// The executor refuses an arbitrary home/session recursive preimage before
// digest/staging, even when the caller labels it owned-tree and supplies a unit.
// CopiedSkillUnit is the nearby positive control for an actual permitted tree.
func TestProvisioningOriginRejectsArbitraryTree(t *testing.T) {
	for _, tc := range []struct{ leaf, class, want string }{{".", "owned-tree", "recursive preimage is outside"}, {"sessions", "owned-tree", "recursive preimage is outside"}, {"auth.json", "owned-tree", "credential leaf requires"}, {"auth.json", "owned-bytes", "credential leaf requires"}, {"sessions", "owned-bytes", "tool state is outside"}, {"sessions/transcript", "owned-bytes", "tool state is outside"}, {"sessions", "installed-profile", "exact removal binding"}, {"sessions", "marker", "fixed home leaf"}, {"sessions", "origin", "fixed home leaf"}} {
		t.Run(tc.class+"/"+tc.leaf, func(t *testing.T) {
			leaf := tc.leaf
			fx, req, _ := recordedOriginFixture(t, "a")
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			marker, e := envmarker.Read(target)
			if e != nil {
				t.Fatal(e)
			}
			if e := req.observeOrigin("codex_cli", marker); e != nil {
				t.Fatal(e)
			}
			source, lock, hash, e := loadResolveInputs(fx.home, fx.profile)
			if e != nil {
				t.Fatal(e)
			}
			canonical, e := lock.Canonical()
			if e != nil {
				t.Fatal(e)
			}
			sourceRaw, e := os.ReadFile(sourcePath(fx.home, fx.profile))
			if e != nil {
				t.Fatal(e)
			}
			_ = source
			manager, e := envunit.ObserveManager(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			mark, e := os.ReadFile(filepath.Join(target, envmarker.Name))
			if e != nil {
				t.Fatal(e)
			}
			dir, e := envunit.ObserveDirectory(target)
			if e != nil {
				t.Fatal(e)
			}
			record := envunit.RecordImage{Kind: "file", Bytes: canonical, Digest: "sha256:" + strings.TrimPrefix(hash, "sha256:")}
			image := envunit.HomeImage{Kind: "recorded", Marker: mark, Origin: req.origin.raw}
			id := newTransactionID()
			unit := &envunit.Unit{Record: envunit.Version, ID: id, Manager: manager, Action: envunit.Maintain, Profiles: []envunit.ProfileTransition{{Name: fx.profile, Source: envorigin.Digest(sourceRaw), Closure: []string{envorigin.Digest([]byte("fixture closure"))}, OldLock: record, NewLock: record}}, Homes: []envunit.Home{{Profile: fx.profile, Adapter: "codex_cli", Mode: "managed-home", Target: target, Root: target, DirectoryBefore: dir, DirectoryAfter: dir, Before: image, After: image}}}
			path := filepath.Join(target, leaf)
			op, e := beginOperation(fx.home)
			if e != nil {
				t.Fatal(e)
			}
			defer op.close()
			before := nativeSnapshot(t, fx.home)
			_, e = op.engine.Prepare(op.lock, transaction.Plan{TransactionID: id, ProjectIdentity: "managed-home-unit", Unit: unit, Targets: []transaction.Target{{Class: tc.class, Identifier: path, Kind: transaction.KindEntry, LivePath: path, PreimageDigest: "absent"}}})
			if e == nil || !strings.Contains(e.Error(), tc.want) {
				t.Fatalf("intended read-scope refusal before digest: %v", e)
			}
			if nativeSnapshot(t, fx.home) != before {
				t.Fatal("read-scope refusal changed manager/home preimages")
			}
		})
	}
}
