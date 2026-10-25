package envprofile

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// These fixtures exercise production native selection/Resolve, not live homes.
// No credential payload is read: auth is only a synthetic regular target.
func managedCallerFixture(t *testing.T) (*managedFixture, ResolveRequest) {
	return managedCallerFixtureAt(t, false)
}

func managedCallerFixtureAt(t *testing.T, custom bool) (*managedFixture, ResolveRequest) {
	t.Helper()
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "review-origin")
	// Synthetic subprocess-only user home/default; no live HOME or native store is touched.
	userHome := filepath.Join(t.TempDir(), "operator")
	native := filepath.Join(userHome, ".codex")
	if custom {
		native = filepath.Join(t.TempDir(), "custom-native")
	}
	if err := os.MkdirAll(native, 0700); err != nil {
		t.Fatal(err)
	}
	fx.native["codex_cli"] = native
	t.Setenv("HOME", userHome)
	t.Setenv("USERPROFILE", userHome)
	t.Setenv("CURATOR_CONFIG", filepath.Join(fx.home, "config.json"))
	if err := os.WriteFile(filepath.Join(native, "auth.json"), []byte("fixture-only\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "config.toml"), []byte("cli_auth_credentials_store = \"file\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", native)
	req := fx.request("codex_cli")
	req.NativeHomeOf = nil // Use the same current-process selection path as the real CLI.
	req.PreserveLock = true
	req.readRegularFile = func(path string) (stateread.File, error) {
		if filepath.Base(path) == "auth.json" {
			t.Fatalf("credential payload read attempted: %s", filepath.Base(path))
		}
		return stateread.ReadRegularFile(path)
	}
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatalf("native provisioning positive control: %v", err)
	}
	req.Repair = false
	req.PreserveLock = false // Bare resolve forbids this repair-only flag.
	return fx, req
}

// Snapshot fixture-managed surfaces/profile and optional native trees; never
// follow links or read credential payloads. Auth uses metadata only.
func managedCallerSnapshot(t *testing.T, fx *managedFixture, nativeRoots ...string) map[string]string {
	t.Helper()
	result := map[string]string{}
	roots := append([]string{ManagedHomeDir(fx.home, fx.profile, "codex_cli"), ProfileDir(fx.home, fx.profile)}, nativeRoots...)
	for index, root := range roots {
		if _, err := os.Lstat(root); os.IsNotExist(err) {
			result[root] = "absent"
			continue
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			info, e := d.Info()
			if e != nil {
				return e
			}
			rel, e := filepath.Rel(fx.home, path)
			if e != nil {
				return e
			}
			value := info.Mode().String()
			if info.Mode()&os.ModeSymlink != 0 {
				target, e := os.Readlink(path)
				if e != nil {
					return e
				}
				value += " -> " + target
			} else if info.Mode().IsRegular() {
				if filepath.Base(path) == "auth.json" {
					if index < 2 {
						return fmt.Errorf("unexpected regular credential in managed snapshot")
					}
					result[rel] = value + fmt.Sprintf(" %d %d", info.Size(), info.ModTime().UnixNano())
					return nil
				}
				b, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				value += fmt.Sprintf(" %x", sha256.Sum256(b))
			}
			result[rel] = value
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return result
}

// Proves the specific invariant: an existing managed root, including a symlink
// alias, must not itself be returned as native origin. No fallback choice is assumed.
func TestManagedCallerHomeIsNeverNative(t *testing.T) {
	for _, caller := range []string{"native", "managed", "managed-alias"} {
		t.Run(caller, func(t *testing.T) {
			fx, _ := managedCallerFixture(t)
			managed := ManagedHomeDir(fx.home, fx.profile, "codex_cli")
			selected := fx.native["codex_cli"]
			if caller != "native" {
				selected = managed
			}
			if caller == "managed-alias" {
				selected = filepath.Join(t.TempDir(), "alias")
				if err := os.Symlink(managed, selected); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("CODEX_HOME", selected)
			adapter, ok := adapterByID("codex_cli")
			if !ok {
				t.Fatal("missing adapter")
			}
			native, err := NativeHome(adapter)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := filepath.EvalSymlinks(native)
			if err != nil {
				t.Fatal(err)
			}
			managedActual, err := filepath.EvalSymlinks(managed)
			if err != nil {
				t.Fatal(err)
			}
			if actual == managedActual {
				t.Errorf("managed caller returned as native origin (%s)", caller)
			}
		})
	}
}

// Required default-native legacy continuity is grounded in the historical source
// contract. The selected default is a current origin, not a historical capture;
// while unchanged original caller controls and state preservation still execute.
func TestManagedCallerDefaultOriginPreserved(t *testing.T) {
	for _, mode := range []string{"verify", "repair-dry-run"} {
		for _, caller := range []string{"native-unset", "native-empty", "native", "managed", "managed-alias"} {
			t.Run(mode+"/"+caller, func(t *testing.T) {
				fx, req := managedCallerFixture(t)
				if mode == "repair-dry-run" {
					req.Repair = true
					req.PreserveLock = true
					req.DryRun = true
				}
				before := managedCallerSnapshot(t, fx)
				if caller == "native-unset" {
					if e := os.Unsetenv("CODEX_HOME"); e != nil {
						t.Fatal(e)
					}
				}
				if caller == "native-empty" {
					t.Setenv("CODEX_HOME", "")
				}
				if caller == "managed" {
					t.Setenv("CODEX_HOME", ManagedHomeDir(fx.home, fx.profile, "codex_cli"))
				}
				if caller == "managed-alias" {
					alias := filepath.Join(t.TempDir(), "alias")
					if e := os.Symlink(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), alias); e != nil {
						t.Fatal(e)
					}
					t.Setenv("CODEX_HOME", alias)
				}
				result, err := Resolve(req)
				if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx)) {
					t.Error("resolve changed managed surfaces/profile/lock/link despite unchanged fixture inputs")
				}
				if err != nil {
					t.Errorf("unchanged direct-link home gets caller-dependent verdict (%s): %v", caller, err)
				}
				if err == nil && (caller == "managed" || caller == "managed-alias") {
					if result == nil || !strings.Contains(strings.Join(result.Warnings, "\n"), nativeHomeFallbackNotice) {
						t.Error("managed fallback must report its current-selection notice")
					}
				}
			})
		}
	}
}

// Negative control: the contract does not admit arbitrary retargeting. A changed
// direct link must still receive the exact credential-conflict refusal, and all
// post-plant managed/profile/lock/link bytes remain untouched on the refused dry-run.
func TestManagedCallerChangedCredentialLinkRefuses(t *testing.T) {
	fx, req := managedCallerFixture(t)
	link := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), "auth.json")
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "auth.json"), link); err != nil {
		t.Fatal(err)
	}
	before := managedCallerSnapshot(t, fx)
	req.Repair = true
	req.DryRun = true
	_, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Errorf("want exact %s refusal, got %v", envregistry.DiagCredentialConflict, err)
	}
	if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx)) {
		t.Error("refused dry-run changed fixture state")
	}
}

// All existing mechanism adapters preserve explicit outside roots, exclude
// managed roots and aliases, and support missing default directories without
// creating them. OpenCode checks the effective child, not only its XDG parent.
func TestNativeHomeSelectionBoundaries(t *testing.T) {
	for _, adapter := range Adapters {
		t.Run(adapter.ID, func(t *testing.T) {
			operator := t.TempDir()
			manager := filepath.Join(t.TempDir(), "manager")
			t.Setenv("HOME", operator)
			t.Setenv("USERPROFILE", operator)
			root := EnvRoot(manager)
			managed := filepath.Join(root, "profile", adapter.ID)
			if err := os.MkdirAll(managed, 0700); err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(operator, adapter.DefaultDir)
			if adapter.ID == "opencode" {
				want = filepath.Join(operator, ".config", "opencode")
			}
			for _, name := range []string{"outside", "outside-missing", "sibling-prefix", "managed", "alias"} {
				t.Run(name, func(t *testing.T) {
					value := managed
					fallback := true
					switch name {
					case "outside":
						value = t.TempDir()
						fallback = false
					case "outside-missing":
						value = filepath.Join(t.TempDir(), "absent", "native")
						fallback = false
					case "sibling-prefix":
						value = root + "-other"
						fallback = false
					case "alias":
						value = filepath.Join(t.TempDir(), "alias")
						if err := os.Symlink(managed, value); err != nil {
							t.Skipf("symlink unavailable: %v", err)
						}
					}
					t.Setenv(adapter.EnvVar, value)
					got, notice, err := nativeHomeAt(adapter, manager)
					if err != nil {
						t.Fatal(err)
					}
					expected := value
					if adapter.ID == "opencode" {
						expected = filepath.Join(value, "opencode")
					}
					if fallback {
						expected = want
					}
					if got != expected || notice != fallback {
						t.Errorf("selection got %q notice=%v, want %q notice=%v", got, notice, expected, fallback)
					}
					if _, err := os.Lstat(want); !os.IsNotExist(err) {
						t.Errorf("selection created/readied missing native default: %v", err)
					}
				})
			}
			if adapter.ID == "opencode" {
				t.Run("effective-child-alias", func(t *testing.T) {
					parent := t.TempDir()
					if err := os.Symlink(managed, filepath.Join(parent, "opencode")); err != nil {
						t.Skipf("symlink unavailable: %v", err)
					}
					t.Setenv(adapter.EnvVar, parent)
					got, notice, err := nativeHomeAt(adapter, manager)
					if err != nil || got != want || !notice {
						t.Errorf("effective child alias got %q %v %v", got, notice, err)
					}
				})
			}
		})
	}
}

// Broken/looping selector links are inspection failures, never a default or
// absent origin; a non-directory ancestor and a managed HOME default also refuse.
func TestNativeHomeUnsafeSelectionRefuses(t *testing.T) {
	adapter, _ := adapterByID("codex_cli")
	for _, name := range []string{"dangling", "loop", "file-parent", "managed-default"} {
		t.Run(name, func(t *testing.T) {
			operator := t.TempDir()
			manager := filepath.Join(t.TempDir(), "manager")
			t.Setenv("HOME", operator)
			t.Setenv("USERPROFILE", operator)
			value := filepath.Join(t.TempDir(), "selector")
			if name == "managed-default" {
				operator = filepath.Join(EnvRoot(manager), "profile", "codex_cli")
				if err := os.MkdirAll(operator, 0700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("HOME", operator)
				t.Setenv("USERPROFILE", operator)
				value = operator
			} else if name == "file-parent" {
				if err := os.WriteFile(value, []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
				value = filepath.Join(value, "child")
			} else {
				target := value
				if name == "dangling" {
					target = value + "-missing"
				}
				if err := os.Symlink(target, value); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			t.Setenv("CODEX_HOME", value)
			if got, _, err := nativeHomeAt(adapter, manager); err == nil || got != "" {
				t.Errorf("unsafe native input must refuse, got %q %v", got, err)
			}
		})
	}
}

// A genuine explicit custom origin is valid. With current default-native skills,
// a managed caller must get the credential conflict in verify/dry-run/repair.
// An absent default instead keeps the earlier exact native-skills refusal.
// Every refusal preserves managed/profile/lock and both native fixture trees;
// no credential payload is read, and no launch fragment may escape.
func TestManagedCallerCustomOriginConflictRefuses(t *testing.T) {
	for _, mode := range []string{"verify", "repair-dry-run", "repair", "missing-native-default"} {
		t.Run(mode, func(t *testing.T) {
			fx, req := managedCallerFixtureAt(t, true)
			userHome, err := os.UserHomeDir()
			if err != nil {
				t.Fatal(err)
			}
			defaultNative := filepath.Join(userHome, ".codex")
			if mode != "missing-native-default" {
				// Real profile-use setup, not a nativeSkillsState stub: prepare the
				// selected default so the credential-conflict boundary is reached.
				t.Setenv("CODEX_HOME", defaultNative)
				results, err := UseWithPolicy(fx.home, fx.profile, "codex_cli", "", false, req.Policy)
				if err != nil {
					t.Fatalf("default native profile-use positive control: %v", err)
				}
				for _, result := range results {
					if !result.OK {
						t.Fatalf("default native profile-use: %s", result.Detail)
					}
				}
				row, err := nativeSkillsState(fx.home, fx.profile, "codex_cli", nil)
				if err != nil || (row != nil && row.State != "current") {
					t.Fatalf("native skills positive control: row=%+v err=%v", row, err)
				}
			}
			t.Setenv("CODEX_HOME", fx.native["codex_cli"])
			if _, err := Resolve(req); err != nil {
				t.Fatalf("explicit custom origin positive control: %v", err)
			}
			if mode != "verify" {
				req.Repair = true
				req.PreserveLock = true
				req.DryRun = mode != "repair"
			}
			t.Setenv("CODEX_HOME", ManagedHomeDir(fx.home, fx.profile, "codex_cli"))
			before := managedCallerSnapshot(t, fx, defaultNative, fx.native["codex_cli"])
			result, err := Resolve(req)
			if mode == "missing-native-default" {
				want := DiagHomeStale + ": native skills home missing"
				if err == nil || err.Error() != want {
					t.Errorf("absent default must retain the earlier exact refusal %q, got %v", want, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
				t.Errorf("custom-origin mismatch must remain %s, got %v", envregistry.DiagCredentialConflict, err)
			}
			if result == nil || !strings.Contains(strings.Join(result.Warnings, "\n"), nativeHomeFallbackNotice) {
				t.Error("fallback must be reported, including an early refusal")
			}
			if result != nil && len(result.Document) != 0 {
				t.Error("refused origin conflict emitted a fragment")
			}
			if !reflect.DeepEqual(before, managedCallerSnapshot(t, fx, defaultNative, fx.native["codex_cli"])) {
				t.Error("custom-origin refusal changed managed/profile/lock or native fixture state")
			}
		})
	}
}

// Physical containment must recognize same-directory case aliases where the
// fixture filesystem supports them, without folding genuinely distinct names.
// Missing children below that same managed root stay managed and are not made.
func TestNativeHomePhysicalContainment(t *testing.T) {
	adapter, _ := adapterByID("codex_cli")
	manager := filepath.Join(t.TempDir(), "manager")
	root := EnvRoot(manager)
	managed := filepath.Join(root, "profile", "codex_cli")
	if err := os.MkdirAll(managed, 0700); err != nil {
		t.Fatal(err)
	}
	operator := t.TempDir()
	t.Setenv("HOME", operator)
	t.Setenv("USERPROFILE", operator)
	want := filepath.Join(operator, ".codex")
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.MkdirAll(outside, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", outside)
	if got, fallback, err := nativeHomeAt(adapter, manager); err != nil || got != outside || fallback {
		t.Fatalf("outside-native positive control: got %q fallback=%v err=%v", got, fallback, err)
	}

	aliasRoot := filepath.Join(manager, "ENVIRONMENTS")
	original, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	alias, aliasErr := os.Stat(aliasRoot)
	caseAlias := aliasErr == nil && os.SameFile(original, alias)
	for _, mode := range []string{"selector", "selector-missing-child", "manager", "manager-missing-child"} {
		t.Run(mode, func(t *testing.T) {
			if !caseAlias {
				t.Skip("fixture does not expose this same-directory case alias")
			}
			selected, managerPath := filepath.Join(aliasRoot, "profile", "codex_cli"), manager
			if strings.HasPrefix(mode, "manager") {
				managerPath = filepath.Join(filepath.Dir(manager), "MANAGER")
				actualManager, err := os.Stat(manager)
				if err != nil {
					t.Fatal(err)
				}
				aliasManager, err := os.Stat(managerPath)
				if err != nil || !os.SameFile(actualManager, aliasManager) {
					t.Skip("fixture does not expose a same-directory manager case alias")
				}
				selected = managed
			}
			missing := strings.HasSuffix(mode, "missing-child")
			if missing {
				selected = filepath.Join(selected, "absent", "child")
			}
			t.Setenv("CODEX_HOME", selected)
			got, fallback, err := nativeHomeAt(adapter, managerPath)
			if err != nil || got != want || !fallback {
				t.Errorf("same-directory managed alias must fall back: got %q fallback=%v err=%v", got, fallback, err)
			}
			if missing {
				if _, err := os.Lstat(selected); !os.IsNotExist(err) {
					t.Errorf("selection materialized absent child: %v", err)
				}
			}
		})
	}
	// Both an absent selector and a managed selector must refuse when the
	// HOME default has a missing tail under the same physical managed root.
	for _, mode := range []string{"default-without-selector", "default-after-managed-selector"} {
		t.Run(mode, func(t *testing.T) {
			if !caseAlias {
				t.Skip("fixture does not expose this same-directory case alias")
			}
			aliasHome := filepath.Join(aliasRoot, "profile", "codex_cli")
			missingDefault := filepath.Join(aliasHome, ".codex")
			if _, err := os.Lstat(missingDefault); !os.IsNotExist(err) {
				t.Fatalf("default fixture must start absent: %v", err)
			}
			t.Setenv("HOME", aliasHome)
			t.Setenv("USERPROFILE", aliasHome)
			selector := ""
			wantFallback := mode == "default-after-managed-selector"
			if wantFallback {
				selector = aliasHome
			}
			t.Setenv("CODEX_HOME", selector)
			got, fallback, err := nativeHomeAt(adapter, manager)
			if err == nil || !strings.Contains(err.Error(), "native HOME default for codex_cli is inside the managed environments root") || got != "" || fallback != wantFallback {
				t.Errorf("managed default must refuse: got %q fallback=%v err=%v", got, fallback, err)
			}
			if _, err := os.Lstat(missingDefault); !os.IsNotExist(err) {
				t.Errorf("refusal materialized absent default: %v", err)
			}
		})
	}
	t.Run("case-distinct-outside", func(t *testing.T) {
		if caseAlias {
			t.Skip("fixture aliases these names instead of keeping distinct directories")
		}
		if err := os.MkdirAll(aliasRoot, 0700); err != nil {
			t.Fatal(err)
		}
		distinct, err := os.Stat(aliasRoot)
		if err != nil {
			t.Fatal(err)
		}
		if os.SameFile(original, distinct) {
			t.Fatal("case-distinct fixture unexpectedly names the managed directory")
		}
		t.Setenv("CODEX_HOME", aliasRoot)
		if got, fallback, err := nativeHomeAt(adapter, manager); err != nil || got != aliasRoot || fallback {
			t.Errorf("case-distinct outside home must remain native: got %q fallback=%v err=%v", got, fallback, err)
		}
	})
}
