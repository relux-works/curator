// Rework-3 regressions for review finding v1-commit-context-bypass: a
// commit-pinned v1 context is refused before any v1 identity is computed
// or trusted, at the production Resolve and switch-materialization
// entries. The v1 NUL refusal shapes are the reviewer's rev4 probe,
// maintained here; the clean and v2 controls are
// production-provisioned so they stay consistent on every platform.
package envprofile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextpkg"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/opaquescan"
	"github.com/relux-works/curator/internal/privatedir"
)

// commitContextProfile installs a git-backed context profile through the
// production store helper: a real commit extracted to the profile store
// entry, a canonical lock at the requested framing, and a git profile
// source.
func commitContextProfile(t *testing.T, files map[string]string, schemaVersion, hashVersion int) (home, profile string, lock *contextlock.Lock, lockHash string) {
	t.Helper()
	home = t.TempDir()
	profile = "review"
	source := "example.test/review"
	commit := writeGitStoreFixture(t, home, contextlock.KindContext, profile, source, files)
	lock = &contextlock.Lock{
		SchemaVersion: schemaVersion,
		HashVersion:   hashVersion,
		Root:          profile,
		Members: []contextlock.Member{{
			Kind: contextlock.KindContext, Name: profile,
			Source: source, Version: "1.0.0", Commit: commit, Weight: 100,
		}},
	}
	lock.Sort()
	var err error
	lockHash, err = contextlock.Write(lockPath(home, profile), lock)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := marshalSource(Source{Kind: KindGit, Git: source, Req: Requirement{Tag: "v1.0.0"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath(home, profile), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	return home, profile, lock, lockHash
}

// commitResolveHomes pins the operator-side homes for one subtest: the
// native home, launch dir, and XDG dir are created once and reused by
// every Resolve call (provision + bare re-check). On Linux the
// claude_code .credentials.json passthrough is a file link into the
// native home, so a fresh native per call detaches the provisioned link
// (environment_home_stale) instead of exercising the NUL guard.
type commitResolveHomes struct {
	native string
	launch string
	xdg    string
}

func newCommitResolveHomes(t *testing.T) *commitResolveHomes {
	t.Helper()
	native := t.TempDir()
	// Live native credential target, as in seedLiveNativeCredentials:
	// harmless where the adapter links nothing (macOS keychain), and it
	// keeps the Linux file-link passthrough live instead of
	// detached-pending.
	if err := os.WriteFile(filepath.Join(native, ".credentials.json"), []byte("{\"t\":\"operator-claude\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return &commitResolveHomes{native: native, launch: t.TempDir(), xdg: t.TempDir()}
}

// request builds a bare claude_code Resolve request over the pinned homes.
func (h *commitResolveHomes) request(home, profile string) ResolveRequest {
	return ResolveRequest{
		Home: home, Profile: profile, EnvID: envregistry.ClaudeCode,
		Machine:   envregistry.DefaultMachineConfig(),
		LaunchDir: h.launch, OperatorXDG: h.xdg,
		Detect:       func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(string) (string, error) { return h.native, nil },
	}
}

// writeLegacyManagedHome records a managed claude_code home whose marker
// matches the lock's current v1 plan byte-for-byte: the surfaces carry
// the v1 identities of the store bytes as assembled, NUL included. A v1
// reader that skipped the opaque guard would report this home current;
// the guard must refuse it first. (A NUL-bearing v1 home cannot be
// provisioned through production Resolve: the pin check refuses it
// before the marker is ever written.)
func writeLegacyManagedHome(t *testing.T, home, profile, lockHash string, lock *contextlock.Lock) {
	t.Helper()
	entry := contextstore.EntryDir(home, contextlock.KindContext, profile, lock.Members[0].PinKey())
	manifest, err := contextpkg.LoadManifest(entry)
	if err != nil {
		t.Fatal(err)
	}
	pkg := contextmaterialize.Package{HasContext: manifest.HasContext}
	for _, module := range manifest.Modules {
		body, err := os.ReadFile(filepath.Join(entry, "context", filepath.FromSlash(module.Path)))
		if err != nil {
			t.Fatal(err)
		}
		pkg.Modules = append(pkg.Modules, contextmaterialize.Module{Module: module, Bytes: body})
	}
	doc, written, err := contextmaterialize.Monolithic(lock, lockHash, contextmaterialize.DefaultPrecedence, envregistry.ClaudeCode, map[string]contextmaterialize.Package{profile: pkg})
	if err != nil || !written {
		t.Fatalf("assemble fixture document: %v written=%v", err, written)
	}
	adapter, err := envregistry.ByID(envregistry.ClaudeCode)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, err := contextmaterialize.SurfaceHashWithVersion(map[string][]byte{adapter.RootTarget: doc}, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	skillsHash, err := contextmaterialize.SurfaceHashWithVersion(map[string][]byte{}, hashing.VersionV1)
	if err != nil {
		t.Fatal(err)
	}
	pass := []envmarker.Passthrough{}
	seeds := []string{}
	mark := &envmarker.Marker{
		Version: 1,
		Profile: envmarker.Profile{
			Name: profile, Root: profile, Kind: "git", Source: lock.Members[0].Source,
			Requirement: &envmarker.Requirement{Tag: "v1.0.0"}, LockSHA256: hashing.Normalize(lockHash),
		},
		Members:     []envmarker.Member{{Name: profile, Version: "1.0.0", Commit: lock.Members[0].Commit, Weight: 100}},
		Precedence:  envmarker.Precedence{Winner: "higher-weight", Placement: "winner-last"},
		Mode:        envmarker.ModeManagedHome,
		Surfaces:    map[string]envmarker.Surface{},
		Passthrough: &pass,
		Seeds:       &seeds,
	}
	copies := []envmarker.Copy{{Path: adapter.RootTarget, Reason: envmarker.ReasonClaudeCodeRootContext}}
	emptyCopies := []envmarker.Copy{}
	mark.Surfaces[envmarker.SurfaceRootContext] = envmarker.Surface{Paths: []string{adapter.RootTarget}, Form: "monolithic", ContentSHA256: rootHash, Copies: &copies}
	mark.Surfaces[envmarker.SurfaceSkills] = envmarker.Surface{Paths: []string{}, ContentSHA256: skillsHash, Copies: &emptyCopies}
	markerBytes, err := mark.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	managed := ManagedHomeDir(home, profile, envregistry.ClaudeCode)
	if err := privatedir.MakeAll(managed); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(managed, envmarker.Name), markerBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(managed, adapter.RootTarget), doc, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestResolveRefusesV1CommitPinnedContextNUL is the maintained form of the
// reviewer's rev4 probe: a legacy v1 lock over a commit-pinned NUL-bearing
// context is refused by the store pin check with the opaque finding at the
// production Resolve entry, before any v1 identity is computed — both when
// the NUL byte sits in a surfaced module and when it hides in a file the
// module projection omits.
//
// Mutants: M1 moves the full-snapshot guard below the member.Commit
// return of storeEntryPinHashes (the rev4 code), so the pin check passes
// and the downstream loadMaterial guard converts the outcome to a stale
// home whose reason carries the opaque text — the pin-routing assertion
// below kills it. M1+M2 additionally drops the loadMaterial guard (the
// full rev4 bypass): both shapes then verify a v1 identity over NUL and,
// where the hand-built marker verifies, Resolve reports the home current
// with nil error — the opaque-text and zero-counter assertions kill it on
// every platform.
func TestResolveRefusesV1CommitPinnedContextNUL(t *testing.T) {
	manifest := `{"schema_version":1,"name":"review","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`
	shapes := []struct {
		name  string
		files map[string]string
	}{
		{"module", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "x\x00y\n",
		}},
		{"excluded", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "clean\n",
			"assets/deep/a.bin":  "x\x00y",
		}},
	}
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			home, profile, lock, lockHash := commitContextProfile(t, shape.files, contextlock.SchemaVersion, 0)
			writeLegacyManagedHome(t, home, profile, lockHash, lock)
			homes := newCommitResolveHomes(t)
			var result *ResolveResult
			var resolveErr error
			calls := hashing.CountV1Hashes(func() {
				result, resolveErr = Resolve(homes.request(home, profile))
			})
			if calls != 0 {
				t.Errorf("v1 commit-pinned NUL context computed %d v1 identities before refusal, want 0", calls)
			}
			if resolveErr == nil || !strings.Contains(resolveErr.Error(), opaquescan.FindingNUL) {
				t.Fatalf("Resolve = (%+v, %v), want the opaque refusal %q", result, resolveErr, opaquescan.FindingNUL)
			}
			var failure *storeEntryFailure
			if !errors.As(resolveErr, &failure) || failure.Check != diagPinHash {
				t.Fatalf("Resolve = (%+v, %v), want the refusal routed through the pin_hash store check", result, resolveErr)
			}
		})
	}
}

// TestResolveCleanV1CommitContextStaysCurrent provisions a clean
// commit-pinned v1 home through production Resolve and re-resolves it
// bare: the home is current, and the v1 counter observes the re-check's
// v1 identities — proving the refusal tests' zero is a guarded path, not
// an unwired seam.
//
// The home is provisioned under the v1 writer so the lock keeps its v1
// framing (repair under the v2 writer would migrate it); the bare
// re-check runs under the restored writer. Mutant: refuse clean v1
// reads (an inverted version dispatch) — the bare resolve fails.
func TestResolveCleanV1CommitContextStaysCurrent(t *testing.T) {
	home, profile, _, _ := commitContextProfile(t, map[string]string{
		"agent-context.json": `{"schema_version":1,"name":"review","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`,
		"context/a.md":       "clean\n",
	}, contextlock.SchemaVersion, 0)
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	defer func() { hashing.EnableV2Writers = prior }()
	homes := newCommitResolveHomes(t)
	provision := homes.request(home, profile)
	provision.Repair = true
	if _, err := Resolve(provision); err != nil {
		t.Fatalf("provision clean v1 home: %v", err)
	}
	hashing.EnableV2Writers = prior
	var result *ResolveResult
	var resolveErr error
	calls := hashing.CountV1Hashes(func() {
		result, resolveErr = Resolve(homes.request(home, profile))
	})
	if resolveErr != nil {
		t.Fatalf("bare resolve of a clean v1 home: %v", resolveErr)
	}
	if result == nil || len(result.Document) == 0 {
		t.Fatal("bare resolve of a clean v1 home emitted no fragment")
	}
	if calls == 0 {
		t.Fatal("clean v1 re-check observed no v1 hash; the counter seam is not wired to this path")
	}
}

// TestResolveAdmitsV2CommitPinnedContextNUL provisions and re-resolves
// NUL-bearing commit-pinned contexts under a v2 lock: v2 hashes 0x00 as
// ordinary data, so both the surfaced-module and the
// projection-excluded shapes resolve current with zero v1 identities.
//
// Mutant: apply the NUL rule to v2 (drop the version dispatch) —
// provisioning refuses with the opaque finding.
func TestResolveAdmitsV2CommitPinnedContextNUL(t *testing.T) {
	manifest := `{"schema_version":1,"name":"review","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`
	shapes := []struct {
		name  string
		files map[string]string
	}{
		{"v2-module", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "x\x00y\n",
		}},
		{"v2-excluded", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "clean\n",
			"assets/deep/a.bin":  "x\x00y",
		}},
	}
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			home, profile, _, _ := commitContextProfile(t, shape.files, contextlock.SchemaVersion2, 2)
			homes := newCommitResolveHomes(t)
			provision := homes.request(home, profile)
			provision.Repair = true
			if _, err := Resolve(provision); err != nil {
				t.Fatalf("provision v2 NUL home: %v", err)
			}
			var result *ResolveResult
			var resolveErr error
			calls := hashing.CountV1Hashes(func() {
				result, resolveErr = Resolve(homes.request(home, profile))
			})
			if resolveErr != nil {
				t.Fatalf("bare resolve of a v2 NUL home: %v", resolveErr)
			}
			if result == nil || len(result.Document) == 0 {
				t.Fatal("bare resolve of a v2 NUL home emitted no fragment")
			}
			if calls != 0 {
				t.Errorf("v2 re-check computed %d v1 identities, want 0", calls)
			}
		})
	}
}

// TestSwitchMaterializeRefusesV1CommitPinnedContextNUL drives the
// production switch materialization (the profile-use path, which reads
// store entries without a preceding pin check) under the v1 writer: a
// commit-pinned v1 NUL context refuses with the opaque finding before any
// v1 identity is computed, for surfaced and excluded NUL alike, while
// the clean control materializes.
//
// Mutant: drop the loadMaterial guard — the NUL shapes materialize and
// record v1 surface identities over NUL bytes instead of refusing.
func TestSwitchMaterializeRefusesV1CommitPinnedContextNUL(t *testing.T) {
	enableV1WritersForTest(t)
	manifest := `{"schema_version":1,"name":"review","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`
	shapes := []struct {
		name        string
		files       map[string]string
		wantRefusal bool
	}{
		{"module", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "x\x00y\n",
		}, true},
		{"excluded", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "clean\n",
			"assets/deep/a.bin":  "x\x00y",
		}, true},
		{"clean", map[string]string{
			"agent-context.json": manifest,
			"context/a.md":       "clean\n",
		}, false},
	}
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			home, profile, _, _ := commitContextProfile(t, shape.files, contextlock.SchemaVersion, 0)
			native := t.TempDir()
			nativeHomeOf := func(string) (string, error) { return native, nil }
			var results []EntryResult
			var materializeErr error
			calls := hashing.CountV1Hashes(func() {
				results, materializeErr = materializeScopeWithNativeHome(home, profile, envregistry.ClaudeCode, Policy{}, nativeHomeOf)
			})
			if !shape.wantRefusal {
				if materializeErr != nil {
					t.Fatalf("materialize clean v1 scope: %v", materializeErr)
				}
				if len(results) != 1 || !results[0].OK {
					t.Fatalf("clean v1 scope results = %+v, want one ok entry", results)
				}
				if calls == 0 {
					t.Fatal("clean v1 switch observed no v1 hash; the counter seam is not wired to this path")
				}
				return
			}
			if calls != 0 {
				t.Errorf("v1 commit-pinned NUL switch computed %d v1 identities before refusal, want 0", calls)
			}
			if materializeErr == nil || !strings.Contains(materializeErr.Error(), opaquescan.FindingNUL) {
				t.Fatalf("materialize = (%+v, %v), want the opaque refusal %q", results, materializeErr, opaquescan.FindingNUL)
			}
			if len(results) != 0 {
				t.Fatalf("refused switch returned %d entry results, want none", len(results))
			}
		})
	}
}
