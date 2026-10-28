package envprofile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
	"github.com/relux-works/curator/internal/transaction"
)

// recordedOriginFixture reaches initial Resolve with only a raw-account seam.
// The actual publication owner must capture the private origin; this fixture
// never supplies a fabricated persisted authority record or a real OS account.
func recordedOriginFixture(t *testing.T, label string) (*managedFixture, ResolveRequest, []byte) {
	return recordedOriginSkillsFixture(t, label, false)
}

func recordedOriginSkillsFixture(t *testing.T, label string, five bool) (*managedFixture, ResolveRequest, []byte) {
	t.Helper()
	enableV1WritersForTest(t)
	fx := writeManagedFixture(t, "origin-fixture")
	if five {
		lock, _, err := readLock(fx.home, fx.profile)
		if err != nil {
			t.Fatal(err)
		}
		for i := 1; i <= 4; i++ {
			name := fmt.Sprintf("fixture-skill-%d", i)
			source := "github.com/example/" + name
			commit := writeGitStoreFixture(t, fx.home, "skill", name, source, map[string]string{"SKILL.md": "# Fixture skill\n"})
			lock.Members = append(lock.Members, contextlock.Member{Kind: "skill", Name: name, Source: source, Commit: commit, RequiredBy: []string{fx.profile}})
		}
		lock.Sort()
		fx.lockHash, err = contextlock.Write(lockPath(fx.home, fx.profile), lock)
		if err != nil {
			t.Fatal(err)
		}
	}
	anchor := t.TempDir()
	native := filepath.Join(t.TempDir(), label)
	if err := os.Mkdir(native, 0700); err != nil {
		t.Fatal(err)
	}
	fx.native["codex_cli"] = native
	t.Setenv("CODEX_HOME", native)
	if err := os.WriteFile(filepath.Join(native, "auth.json"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	req.originAccountLookup = func() (envorigin.Account, error) {
		return envorigin.Account{UID: uint64(os.Geteuid()), RealUID: uint64(os.Getuid()), EffectiveUID: uint64(os.Geteuid()), Home: anchor}, nil
	}
	if result, err := Resolve(req); err != nil || result == nil || len(result.Document) == 0 {
		t.Fatalf("recorded fixture provisioning: %v %v", result, err)
	}
	target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
	path := filepath.Join(target, envorigin.Name)
	raw, present, err := envorigin.ReadPrivateAuthority(fx.home, path, envorigin.MaxRecordBytes)
	if err != nil || !present {
		t.Fatalf("initial owner did not publish authority: present=%v error=%v", present, err)
	}
	binding := envorigin.Binding{Profile: fx.profile, LockSHA256: "sha256:" + strings.TrimPrefix(fx.lockHash, "sha256:"), ManagedTarget: target}
	origin, err := envorigin.Parse(raw, binding, "codex_cli", fx.home)
	if err != nil || origin.NativeRoot != native || origin.OperatorHome != anchor {
		t.Fatalf("initial capture differs from native selection: %v %v", origin, err)
	}
	req.Repair = false
	return fx, req, raw
}

// Proves actual Resolve credential/backend policy remains at recorded A under
// today's B, in both label directions and plain/repair. This is not a launcher
// no-start or real account transport witness. Fixture credential files are empty.
func TestProvisioningOriginRecordedAuthority(t *testing.T) {
	for _, direction := range []string{"a-to-b", "b-to-a"} {
		t.Run(direction, func(t *testing.T) {
			original, other := "a", "b"
			if direction == "b-to-a" {
				original, other = "b", "a"
			}
			fx, req, origin := recordedOriginFixture(t, original)
			a := fx.native["codex_cli"]
			b := filepath.Join(filepath.Dir(a), other)
			if err := os.Mkdir(b, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(b, "config.toml"), []byte("cli_auth_credentials_store = 'keyring'\n"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CODEX_HOME", b)
			req.NativeHomeOf = func(id string) (string, error) {
				if id == "codex_cli" {
					return b, nil
				}
				return fx.native[id], nil
			}
			accountCalls, bReads := 0, 0
			req.originAccountLookup = func() (envorigin.Account, error) {
				accountCalls++
				return envorigin.Account{}, fmt.Errorf("recorded maintenance must not discover account")
			}
			req.readRegularFile = func(path string) (stateread.File, error) {
				if path == filepath.Join(b, "config.toml") {
					bReads++
				}
				return stateread.ReadRegularFile(path)
			}
			for _, repair := range []bool{false, true} {
				req.Repair = repair
				result, err := Resolve(req)
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("recorded A must remain current under B (repair=%v): %v", repair, err)
				}
				link, err := os.Readlink(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), "auth.json"))
				if err != nil || link != filepath.Join(a, "auth.json") {
					t.Fatalf("recorded A link changed: %q %v", link, err)
				}
				got, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), envorigin.Name))
				if err != nil || !bytes.Equal(got, origin) {
					t.Fatal("maintenance recaptured origin")
				}
			}
			if accountCalls != 0 || bReads != 0 {
				t.Fatalf("recorded policy consulted mutable authority: account=%d B-config=%d", accountCalls, bReads)
			}
		})
	}
}

// Proves actual Resolve refuses legacy before backend observation and preserves
// marker/link state for plain, repair and dry-run. No arbitrary error qualifies.
func TestProvisioningOriginLegacyRefusal(t *testing.T) {
	fx, req, _ := recordedOriginFixture(t, "a")
	target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
	if err := os.Remove(filepath.Join(target, envorigin.Name)); err != nil {
		t.Fatal(err)
	}
	marker, err := os.ReadFile(filepath.Join(target, envmarker.Name))
	if err != nil {
		t.Fatal(err)
	}
	link, err := os.Readlink(filepath.Join(target, "auth.json"))
	if err != nil {
		t.Fatal(err)
	}
	observations := 0
	req.Detect = func(_adapter envregistry.Adapter) string { observations++; return "unknown" }
	for _, mode := range []string{"plain", "repair", "dry-run"} {
		req.Repair = mode != "plain"
		req.DryRun = mode == "dry-run"
		result, err := Resolve(req)
		if err == nil || !strings.Contains(err.Error(), envorigin.DiagRequired) {
			t.Fatalf("%s must refuse specifically origin-required, got %v", mode, err)
		}
		if result != nil && len(result.Document) != 0 {
			t.Fatal("legacy refusal emitted fragment")
		}
		got, err := os.ReadFile(filepath.Join(target, envmarker.Name))
		if err != nil || !bytes.Equal(got, marker) {
			t.Fatal("legacy refusal changed marker")
		}
		gotLink, err := os.Readlink(filepath.Join(target, "auth.json"))
		if err != nil || gotLink != link {
			t.Fatal("legacy refusal changed credential link")
		}
	}
	if observations != 0 {
		t.Fatalf("legacy refusal reached backend observation %d times", observations)
	}
}

// Proves the actual repair continuation updates all five stale native skill
// links on a healthy recorded-A home, but refuses a managed credential link to
// caller B before any such update. Snapshots cover this synthetic fixture only;
// this does not qualify a real provider/account or arbitrary session contents.
func TestProvisioningOriginFiveLinkRepair(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(map[bool]string{false: "healthy", true: "wrong-target"}[conflict], func(t *testing.T) {
			fx, req, _ := recordedOriginSkillsFixture(t, "a", true)
			a := fx.native["codex_cli"]
			op, err := beginOperation(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			_, err = materializeScopeWithNativeHome(op, fx.home, fx.profile, "codex_cli", Policy{}, req.NativeHomeOf)
			if closeErr := op.close(); err == nil {
				err = closeErr
			}
			if err != nil {
				t.Fatal(err)
			}
			lock, _, err := readLock(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			skills, err := skillsOf(fx.home, newGitManager(fx.home), lock)
			if err != nil || len(skills) != 5 {
				t.Fatalf("five skill setup: %d %v", len(skills), err)
			}
			beforeLinks := map[string]string{}
			for _, skill := range skills {
				path := filepath.Join(a, "skills", skill.name)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(skill.source, path); err != nil {
					t.Fatal(err)
				}
				beforeLinks[path] = skill.source
			}
			b := t.TempDir()
			if err := os.WriteFile(filepath.Join(b, "config.toml"), []byte("cli_auth_credentials_store = 'keyring'\n"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CODEX_HOME", b)
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			if conflict {
				if err := os.Remove(filepath.Join(target, "auth.json")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(b, "auth.json"), filepath.Join(target, "auth.json")); err != nil {
					t.Fatal(err)
				}
			}
			beforeNative, beforeManaged := nativeSnapshot(t, a), nativeSnapshot(t, target)
			beforeLock, err := os.ReadFile(lockPath(fx.home, fx.profile))
			if err != nil {
				t.Fatal(err)
			}
			bReads := 0
			req.readRegularFile = func(path string) (stateread.File, error) {
				if path == filepath.Join(b, "config.toml") {
					bReads++
				}
				return stateread.ReadRegularFile(path)
			}
			req.originAccountLookup = func() (envorigin.Account, error) {
				t.Error("maintenance rediscovered account")
				return envorigin.Account{}, fmt.Errorf("forbidden discovery")
			}
			req.Repair = true
			result, err := Resolve(req)
			if conflict {
				if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
					t.Fatalf("wrong target must refuse specifically: %v", err)
				}
				if result != nil && len(result.Document) > 0 {
					t.Fatal("refusal emitted fragment")
				}
				if nativeSnapshot(t, a) != beforeNative || nativeSnapshot(t, target) != beforeManaged {
					t.Fatal("credential refusal changed fixture preimages")
				}
				for path, want := range beforeLinks {
					got, err := os.Readlink(path)
					if err != nil || got != want {
						t.Fatalf("refusal changed skill link %s", filepath.Base(path))
					}
				}
			} else {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("healthy repair continuation: %v", err)
				}
				for _, skill := range skills {
					got, err := os.Readlink(filepath.Join(a, "skills", skill.name))
					if err != nil || got != skill.root {
						t.Fatalf("healthy continuation did not repair %s: %v", skill.name, err)
					}
				}
			}
			afterLock, err := os.ReadFile(lockPath(fx.home, fx.profile))
			if err != nil || !bytes.Equal(beforeLock, afterLock) {
				t.Fatal("repair changed lock")
			}
			if bReads != 0 {
				t.Fatalf("read caller-B backend %d times", bReads)
			}
		})
	}
}

// Proves real home publication retains a phase-bound unit through an abrupt
// rename-window stop, then a fresh engine recovers recorded A under caller B.
// Ordinary returned failure instead restores all enumerated home preimages.
// This is an in-process stop model, not an OS power-loss/process qualification.
func TestProvisioningOriginWholeUnitRecovery(t *testing.T) {
	for _, mode := range []string{"returned-error", "restart"} {
		t.Run(mode, func(t *testing.T) {
			fx, req, origin := recordedOriginFixture(t, "a")
			target := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			path := filepath.Join(target, "AGENTS.md")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, target)
			reached := false
			failure := fmt.Errorf("owned unit publication fault")
			req.Repair = true
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
				if event.Point == transaction.PointAfterInstall && event.Class == "marker" {
					reached = true
					if mode == "restart" {
						panic("owned unit simulated stop")
					}
					return failure
				}
				return nil
			}})}
			if mode == "returned-error" {
				result, err := Resolve(req)
				if !errors.Is(err, failure) || result != nil && len(result.Document) > 0 {
					t.Fatalf("intended error/fragment: %v", err)
				}
				if !reached || nativeSnapshot(t, target) != before {
					t.Fatal("returned error did not restore exact managed preimages")
				}
			} else {
				func() {
					defer func() {
						if r := recover(); r != "owned unit simulated stop" {
							t.Fatalf("wrong stop: %v", r)
						}
					}()
					_, _ = Resolve(req)
				}()
				if !reached {
					t.Fatal("marker publication not reached")
				}
				b := t.TempDir()
				t.Setenv("CODEX_HOME", b)
				req.transactionOptions = nil
				req.originAccountLookup = func() (envorigin.Account, error) {
					t.Error("recovery rediscovered account")
					return envorigin.Account{}, failure
				}
				result, err := Resolve(req)
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("fresh engine recorded-A recovery: %v", err)
				}
				got, err := os.ReadFile(filepath.Join(target, envorigin.Name))
				if err != nil || !bytes.Equal(got, origin) {
					t.Fatal("recovery recaptured origin")
				}
				entries, err := os.ReadDir(filepath.Join(fx.home, "state", "transactions", "v1"))
				if err != nil || len(entries) != 0 {
					t.Fatalf("cleanup incomplete: %v %v", entries, err)
				}
			}
		})
	}
}
