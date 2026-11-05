package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
)

// Adapted upstream witness, retaining its name. Install -> resolver ->
// VerifyCandidate must acquire the exact signed object before authorization.
func TestIndependentSignedMissingPin(t *testing.T) {
	for _, mode := range []string{"stale", "cached-offline", "wrong-signer", "unsigned"} {
		t.Run(mode, func(t *testing.T) {
			f := newPinTransportFixture(t)
			key, pub := statusSignerKey(t)
			if mode != "unsigned" {
				gitRun(t, f.repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "commit", "--amend", "-S", "--no-edit")
			}
			f.pin = fixtureGitOutput(t, f.repo, "rev-parse", "HEAD")
			assertPinTransportMissing(t, f.cache, f.pin)
			if mode == "wrong-signer" {
				_, pub = statusSignerKey(t)
			}
			writeManifestPackage(t, f.root, fmt.Sprintf(`{"schema_version":1,"name":"review-transport-root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, f.operand, f.pin), nil)
			if mode == "cached-offline" {
				// Same signed selection, but cached: unreachable origin proves no fetch.
				gitRun(t, f.cache, "fetch", "--no-tags", "--no-write-fetch-head", "origin", f.pin)
				gitRun(t, f.cache, "remote", "set-url", "origin", gitFileURL(filepath.Join(t.TempDir(), "absent")))
			}
			refs := fixtureGitOutput(t, f.cache, "show-ref")
			trace := filepath.Join(t.TempDir(), "trace")
			t.Setenv("GIT_TRACE", trace)
			policy := Policy{SourceSigners: map[string][]contextresolve.Signer{f.source: {{Type: "ssh", Key: pub}}}}
			info, _, _, installErr := Install(f.home, InstallOptions{Operand: f.root, Policy: policy})
			success := mode == "stale" || mode == "cached-offline"
			if success {
				// Independent pin and raw payload assertions, not merely absence of error.
				if installErr != nil {
					t.Fatal(installErr)
				}
				member, ok := info.Lock.Find(contextlock.KindContext, "dep")
				if !ok || member.Commit != f.pin || member.Source != f.source {
					t.Fatalf("wrong selection: %+v", member)
				}
				payload, err := os.ReadFile(filepath.Join(contextstore.EntryDir(f.home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
				if err != nil || string(payload) != pinTransportBytes {
					t.Fatalf("wrong published bytes: %q %v", payload, err)
				}
				if _, err := os.Stat(ProfileDir(f.home, "review-transport-root")); err != nil {
					t.Fatal(err)
				}
			} else {
				// Wrong signer and absent signature must refuse after object acquisition,
				// before any dependency store entry or profile/lock/home publication.
				diagnostic := contextresolve.DiagSourceSignerRejected
				if mode == "unsigned" {
					diagnostic = contextresolve.DiagSourceUnsigned
				}
				if installErr == nil || !strings.Contains(installErr.Error(), diagnostic) {
					t.Fatalf("wrong signer refusal: %v", installErr)
				}
				if strings.Contains(installErr.Error(), "read commit signature") {
					t.Fatalf("object was not acquired: %v", installErr)
				}
				if _, err := os.Stat(filepath.Join(contextstore.Root(f.home), contextlock.KindContext, "dep")); !os.IsNotExist(err) {
					t.Fatalf("rejected dependency published: %v", err)
				}
			}
			f.assertPreserved(t, success)
			if fixtureGitOutput(t, f.cache, "show-ref") != refs {
				t.Fatal("selected refs changed")
			}
			if got := fixtureGitOutput(t, f.cache, "cat-file", "-t", f.pin); got != "commit" {
				t.Fatalf("exact object not acquired: %s", got)
			}
			raw, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			want := 1
			if mode == "cached-offline" {
				want = 0
			}
			if got := strings.Count(string(raw), "built-in: git fetch "); got != want {
				t.Fatalf("fetches = %d, want %d", got, want)
			}
			if want == 1 && !strings.Contains(string(raw), "fetch --no-tags --no-write-fetch-head --no-auto-maintenance --recurse-submodules=no --refmap= origin "+f.pin) {
				t.Fatal("exact acquisition flags missing")
			}
		})
	}
}

// Install -> resolver -> VerifyCandidate -> acquireMissingPin must preserve
// configured existing refs even after signer refusal. Objects may survive a
// refusal: this proves ref/metadata and publication boundaries, not rollback.
func TestInstallMissingPinPreservesConfiguredRefs(t *testing.T) {
	for _, destination := range []string{"refs/tags/v1.0.0", "refs/remotes/origin/preserved"} {
		for _, absent := range []bool{false, true} {
			for _, mode := range []string{"stale", "wrong-signer", "unsigned", "cached-offline"} {
				t.Run(fmt.Sprintf("%s/absent-%t/%s", destination, absent, mode), func(t *testing.T) {
					f := newPinTransportFixture(t)
					key, pub := statusSignerKey(t)
					if mode != "unsigned" {
						gitRun(t, f.repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "commit", "--amend", "-S", "--no-edit")
					}
					f.pin = fixtureGitOutput(t, f.repo, "rev-parse", "HEAD")
					assertPinTransportMissing(t, f.cache, f.pin)
					if mode == "wrong-signer" {
						_, pub = statusSignerKey(t)
					}
					writeManifestPackage(t, f.root, fmt.Sprintf(`{"schema_version":1,"name":"review-transport-root","version":"1.0.0","requires":{"contexts":{"dep":{"git":%q,"revision":%q}}}}`, f.operand, f.pin), nil)
					if mode == "cached-offline" {
						// Same signed selection, but cached: unreachable origin proves no fetch.
						gitRun(t, f.cache, "fetch", "--no-tags", "--no-write-fetch-head", "origin", f.pin)
						gitRun(t, f.cache, "remote", "set-url", "origin", gitFileURL(filepath.Join(t.TempDir(), "absent")))
					}
					// The expected destination is grounded in the original remote history,
					// independently of the new pin and the implementation under test.
					expectedRef := fixtureGitOutput(t, f.repo, "rev-parse", "v1.0.0^{commit}")
					if expectedRef == f.pin {
						t.Fatal("fixture requires distinct old and new commits")
					}
					gitRun(t, f.cache, "update-ref", destination, expectedRef)
					gitRun(t, f.cache, "config", "--replace-all", "remote.origin.fetch", "+"+f.pin+":"+destination)
					fetchHead := filepath.Join(f.cache, ".git", "FETCH_HEAD")
					if absent {
						if err := os.Remove(fetchHead); err != nil {
							t.Fatal(err)
						}
					}
					refs := fixtureGitOutput(t, f.cache, "show-ref")
					trace := filepath.Join(t.TempDir(), "trace")
					t.Setenv("GIT_TRACE", trace)
					policy := Policy{SourceSigners: map[string][]contextresolve.Signer{f.source: {{Type: "ssh", Key: pub}}}}
					info, _, _, installErr := Install(f.home, InstallOptions{Operand: f.root, Policy: policy})
					success := mode == "stale" || mode == "cached-offline"
					if success {
						// Independent pin and raw payload assertions, not merely absence of error.
						if installErr != nil {
							t.Fatal(installErr)
						}
						member, ok := info.Lock.Find(contextlock.KindContext, "dep")
						if !ok || member.Commit != f.pin || member.Source != f.source {
							t.Fatalf("wrong selection: %+v", member)
						}
						payload, err := os.ReadFile(filepath.Join(contextstore.EntryDir(f.home, member.Kind, member.Name, member.PinKey()), "context", "a.md"))
						if err != nil || string(payload) != pinTransportBytes {
							t.Fatalf("wrong published bytes: %q %v", payload, err)
						}
						if _, err := os.Stat(ProfileDir(f.home, "review-transport-root")); err != nil {
							t.Fatal(err)
						}
					} else {
						// Wrong signer and absent signature must refuse after object acquisition,
						// before any dependency store entry or profile/lock/home publication.
						diagnostic := contextresolve.DiagSourceSignerRejected
						if mode == "unsigned" {
							diagnostic = contextresolve.DiagSourceUnsigned
						}
						if installErr == nil || !strings.Contains(installErr.Error(), diagnostic) {
							t.Fatalf("wrong signer refusal: %v", installErr)
						}
						if strings.Contains(installErr.Error(), "read commit signature") {
							t.Fatalf("object was not acquired: %v", installErr)
						}
						if _, err := os.Stat(filepath.Join(contextstore.Root(f.home), contextlock.KindContext, "dep")); !os.IsNotExist(err) {
							t.Fatalf("rejected dependency published: %v", err)
						}
					}
					// Absence is checked before restoring only the fixture sentinel used by
					// the shared state assertion; failed reads never count as absence.
					if absent {
						if _, err := os.Lstat(fetchHead); !os.IsNotExist(err) {
							t.Fatalf("FETCH_HEAD created or unreadable: %v", err)
						}
						if err := os.WriteFile(fetchHead, []byte("keep fetch metadata\n"), 0o600); err != nil {
							t.Fatal(err)
						}
					}
					f.assertPreserved(t, success)
					if got := fixtureGitOutput(t, f.cache, "rev-parse", destination); got != expectedRef {
						t.Errorf("existing destination ref changed: %s got %s want %s", destination, got, expectedRef)
					}
					if fixtureGitOutput(t, f.cache, "show-ref") != refs {
						t.Fatal("selected refs changed")
					}
					if got := fixtureGitOutput(t, f.cache, "cat-file", "-t", f.pin); got != "commit" {
						t.Fatalf("exact object not acquired: %s", got)
					}
					raw, err := os.ReadFile(trace)
					if err != nil {
						t.Fatal(err)
					}
					want := 1
					if mode == "cached-offline" {
						want = 0
					}
					if got := strings.Count(string(raw), "built-in: git fetch "); got != want {
						t.Fatalf("fetches = %d, want %d", got, want)
					}
				})
			}
		}
	}
}
