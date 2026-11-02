package envprofile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/transaction"
)

func captureFixture(t *testing.T) (*managedFixture, MaterializeCaptureRequest) {
	t.Helper()
	fx := writeManagedFixture(t, "capture")
	seedLiveNativeCredentials(t, fx)
	provision := fx.request(envregistry.CodexCLI)
	provision.Repair = true
	if _, err := Resolve(provision); err != nil {
		t.Fatal(err)
	}
	req := MaterializeCaptureRequest{Resolve: fx.request(envregistry.CodexCLI)}
	req.Resolve.Detect = nil // Exercise the production no-probe boundary.
	req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
		return envregistry.DefaultMachineConfig(), Policy{}, nil
	}
	req.Capture = func(c *MaterializeCapture) ([]byte, error) {
		path := filepath.Join(ManagedHomeDir(c.Request.Home, c.Request.Profile, c.Request.EnvID), "AGENTS.md")
		path, err := filepath.EvalSymlinks(path) // Fixture follows the marker-authorized rendered link.
		if err != nil {
			return nil, err
		}
		file, err := openReadNoFollow(path)
		if err != nil {
			return nil, err
		}
		return c.ReadObject(path, file)
	}
	return fx, req
}

func requireCaptureRefusal(t *testing.T, req MaterializeCaptureRequest, code string) {
	t.Helper()
	got, err := CaptureMaterialize(req)
	if err == nil || !strings.Contains(err.Error(), code) || got != nil {
		t.Fatalf("capture = %v, %v; want %s and no output", got, err, code)
	}
}

func captureTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		value := info.Mode().String()
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			value += target
		} else if info.Mode().IsRegular() {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			// Windows byte-range locks forbid a second reader. Normal lock
			// bookkeeping is observed by name, mode and size, not source bytes.
			if filepath.Dir(relative) == filepath.Join("state", "locks", "v1") && strings.HasSuffix(relative, ".lock") {
				result[path] = fmt.Sprintf("%s size=%d", value, info.Size())
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += string(data)
		}
		result[path] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// Snapshot owned fixture roots without dereferencing credential links.
func captureFixtureState(t *testing.T, fx *managedFixture) ([]byte, error) {
	t.Helper()
	state := map[string]map[string]string{"manager": captureTree(t, fx.home), "xdg": captureTree(t, fx.xdg), "launch": captureTree(t, fx.launch)}
	for id, root := range fx.native {
		state["native-"+id] = captureTree(t, root)
	}
	source, err := readSource(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	if source.Path != "" {
		state["path-source"] = captureTree(t, source.Path)
	}
	return json.Marshal(state)
}

// CaptureMaterialize returns verified bytes and releases its real lock before
// the caller can publish. Successful capture changes neither source nor target.
func TestMaterializeCaptureCurrentAndRelease(t *testing.T) {
	fx, req := captureFixture(t)
	target := t.TempDir()
	before, _ := captureFixtureState(t, fx)
	targetBefore, _ := json.Marshal(captureTree(t, target))
	got, err := CaptureMaterialize(req)
	if err != nil || got == nil || len(got.Fragment) == 0 || !bytes.Contains(got.Content, []byte("hello")) {
		t.Fatalf("capture = %v, %v", got, err)
	}
	after, _ := captureFixtureState(t, fx)
	targetAfter, _ := json.Marshal(captureTree(t, target))
	if !bytes.Equal(before, after) || !bytes.Equal(targetBefore, targetAfter) {
		t.Fatal("observation mutated source or target")
	}
	manager, err := managerlock.New(fx.home)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	lock, err := manager.AcquireExistingHomeOnly(ctx)
	if err != nil {
		t.Fatal("lock retained after capture:", err)
	}
	_ = lock.Close()
}

// Native storage captures; unavailable explicit Codex stores refuse exactly as
// Resolve does, before the closure callback, with no output or fixture changes.
func TestMaterializeCaptureCredentialStoreAdmission(t *testing.T) {
	for _, tc := range []struct {
		selection envregistry.CredentialStoreSelection
		code      string
	}{
		{envregistry.CredentialStoreFile, envregistry.DiagCodexSettingsUnsupported},
		{envregistry.CredentialStoreKeychain, envregistry.DiagCredentialStoreUnsupported},
	} {
		t.Run(string(tc.selection), func(t *testing.T) {
			fx, req := captureFixture(t)
			control, err := CaptureMaterialize(req)
			if err != nil || control == nil || !bytes.Contains(control.Content, []byte("hello")) {
				t.Fatalf("native control = %v, %v", control, err)
			}
			machine := envregistry.DefaultMachineConfig()
			machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{
				fx.profile: {envregistry.CodexCLI: tc.selection},
			}
			req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
				return machine, Policy{}, nil
			}
			captured := false
			capture := req.Capture
			req.Capture = func(c *MaterializeCapture) ([]byte, error) {
				captured = true
				return capture(c)
			}
			before, _ := captureFixtureState(t, fx)
			got, captureErr := CaptureMaterialize(req)
			if captureErr == nil || !strings.Contains(captureErr.Error(), tc.code) || got != nil || captured {
				t.Fatalf("capture = %v, %v, callback=%v; want %s before capture", got, captureErr, captured, tc.code)
			}
			resolve := req.Resolve
			resolve.Machine = machine
			resolved, resolveErr := Resolve(resolve)
			if resolveErr == nil || resolveErr.Error() != captureErr.Error() || resolved != nil {
				t.Fatalf("Resolve = %v, %v; want capture refusal %v and no fragment", resolved, resolveErr, captureErr)
			}
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("credential-store refusal changed source or credential state")
			}
		})
	}
}

// Every durable journal phase refuses at the production capture entry point;
// real recovery would produce observed transaction events and/or mutate records.
func TestMaterializeCapturePendingDoesNotRecover(t *testing.T) {
	for _, phase := range []transaction.Phase{transaction.PhasePrepared, transaction.PhaseCommitting, transaction.PhaseCleanup, transaction.PhaseRollingBack, transaction.PhaseRollbackCleanup} {
		t.Run(string(phase), func(t *testing.T) {
			fx, req := captureFixture(t)
			if _, err := CaptureMaterialize(req); err != nil {
				t.Fatal("positive control:", err)
			}
			manager, _ := managerlock.New(fx.home)
			lock, err := manager.AcquireExistingHomeOnly(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			engine, err := transaction.New(fx.home)
			if err != nil {
				t.Fatal(err)
			}
			live := filepath.Join(t.TempDir(), "live")
			stage := filepath.Join(t.TempDir(), "stage")
			if err := os.WriteFile(live, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(stage, []byte("new"), 0600); err != nil {
				t.Fatal(err)
			}
			digest, err := transaction.DigestTarget(transaction.KindBytes, live)
			if err != nil {
				t.Fatal(err)
			}
			journal, err := engine.Prepare(lock, transaction.Plan{TransactionID: "capture-pending", ProjectIdentity: "capture", Targets: []transaction.Target{{Class: "profile", Identifier: "live", LivePath: live, StagedSource: stage, PreimageDigest: digest}}})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(fx.home, "state", "transactions", "v1", "capture-pending.json")
			journal.Phase = phase
			// encoding/json preserves the journal struct's canonical field order.
			payload, err := json.Marshal(journal)
			if err != nil {
				t.Fatal(err)
			}
			payload = append(payload, '\n')
			if err := os.WriteFile(path, payload, 0600); err != nil {
				t.Fatal(err)
			}
			_ = lock.Close()
			events := 0
			recoveries := 0
			req.Resolve.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Observe: func(transaction.Event) { events++ }, ObserveRecovery: func() { recoveries++ }})}
			before, _ := captureFixtureState(t, fx)
			requireCaptureRefusal(t, req, DiagMaterializeRecoveryRequired)
			after, _ := captureFixtureState(t, fx)
			liveBytes, err := os.ReadFile(live)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) || string(liveBytes) != "old" || events != 0 || recoveries != 0 {
				t.Fatal("capture recovered or changed pending state")
			}
		})
	}
}

// Corrupt and unreadable inventory remain errors, never clean-source fallback;
// an absent inventory on a current source is the adjacent positive control.
func TestMaterializeCaptureBadRecoveryEvidence(t *testing.T) {
	for _, kind := range []string{"corrupt", "not-directory"} {
		t.Run(kind, func(t *testing.T) {
			fx, req := captureFixture(t)
			if _, err := CaptureMaterialize(req); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(fx.home, "state", "transactions", "v1")
			if kind == "corrupt" {
				if err := os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path, "broken.json"), []byte("{"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("blocked"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := captureFixtureState(t, fx)
			code := "invalid transaction journal"
			if kind == "not-directory" {
				code = "manager_state_unreadable"
			}
			requireCaptureRefusal(t, req, code)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("failed observation changed evidence")
			}
		})
	}
}

// A busy real lock times out without invoking selection or capture. Release
// permits a fresh scoped selection and machine policy rather than stale inputs.
func TestMaterializeCaptureWaitReselectsAndTimeout(t *testing.T) {
	fx, req := captureFixture(t)
	manager, _ := managerlock.New(fx.home)
	lock, err := manager.AcquireExistingHomeOnly(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	old := lockTimeout
	lockTimeout = 30 * time.Millisecond
	t.Cleanup(func() { lockTimeout = old })
	calls := 0
	req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
		calls++
		return envregistry.DefaultMachineConfig(), Policy{}, nil
	}
	requireCaptureRefusal(t, req, DiagLockUnavailable)
	if calls != 0 {
		t.Fatal("selected policy before acquiring lock")
	}
	_ = lock.Close()
	// The released positive control uses the production lock budget; only
	// the refusal above needs a deliberately short deadline.
	lockTimeout = old
	if _, err := CaptureMaterialize(req); err != nil {
		t.Fatal("released control:", err)
	}
	// While the waiter is blocked, a cooperating writer updates current selection.
	req.Resolve.Profile = ""
	if err := SetCurrent(fx.home, "missing-before-wait"); err != nil {
		t.Fatal(err)
	}
	req.Resolve.Machine.Permissions = map[string]string{"capture": "yolo"}
	lock, err = manager.AcquireExistingHomeOnly(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	req.Capture = func(c *MaterializeCapture) ([]byte, error) {
		if c.Request.Profile != "capture" || c.Request.Machine.Permissions["capture"] != "native" {
			return nil, errors.New("stale selection")
		}
		return []byte("new generation"), nil
	}
	policyCalls := make(chan struct{}, 1)
	req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
		select {
		case policyCalls <- struct{}{}:
		default:
		}
		m := envregistry.DefaultMachineConfig()
		m.Permissions = map[string]string{"capture": "native"}
		return m, Policy{}, nil
	}
	go func() { _, err := CaptureMaterialize(req); result <- err }()
	select {
	case <-policyCalls:
		t.Fatal("machine selected while lock held")
	case <-time.After(30 * time.Millisecond):
	}
	// Simulate the already locked writer without recursively acquiring its lock.
	if err := os.WriteFile(CurrentFile(fx.home), []byte("capture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_ = lock.Close()
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}

// Same-fragment ABA/replacement, mode/byte changes, failed reads and policy
// changes must abort private output; the unmodified callback is the control.
func TestMaterializeCaptureSameFragmentDrift(t *testing.T) {
	for _, mutation := range []string{"aba", "mode", "bytes", "missing", "closed-handle", "policy", "composition-policy", "store-bytes"} {
		t.Run(mutation, func(t *testing.T) {
			fx, req := captureFixture(t)
			if _, err := CaptureMaterialize(req); err != nil {
				t.Fatal("stable control:", err)
			}
			var expected []byte
			original := req.Capture
			machine := envregistry.DefaultMachineConfig()
			policy := Policy{}
			req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) { return machine, policy, nil }
			req.Capture = func(c *MaterializeCapture) ([]byte, error) {
				content, err := original(c)
				if err != nil {
					return nil, err
				}
				path := c.objects[len(c.objects)-1].path
				switch mutation {
				case "aba":
					info, err := os.Stat(path)
					if err != nil {
						return nil, err
					}
					replacement := path + ".replacement"
					if err := os.WriteFile(replacement, content, info.Mode()); err != nil {
						return nil, err
					}
					if err := os.Chtimes(replacement, info.ModTime(), info.ModTime()); err != nil {
						return nil, err
					}
					err = replaceCapturedFixture(t, replacement, path)
					if err != nil {
						return nil, err
					}
					verified, err := resolveVerified(c.Request, verifyHomeContent)
					if err != nil || !bytes.Equal(verified.Document, c.Fragment) {
						return nil, fmt.Errorf("ABA witness changed the fragment: %v", err)
					}

				case "mode":
					if runtime.GOOS == "windows" {
						t.Skip("Windows does not expose portable executable permission bits")
					}
					info, err := os.Stat(path)
					if err != nil {
						return nil, err
					}
					if err := os.Chmod(path, info.Mode()^0100); err != nil {
						return nil, err
					}
				case "bytes":
					if err := os.WriteFile(path, []byte("changed"), 0600); err != nil {
						return nil, err
					}
				case "missing":
					if err := os.Remove(path); err != nil {
						return nil, err
					}
				case "closed-handle":
					_ = c.objects[len(c.objects)-1].file.Close()
				case "policy":
					machine.Permissions = map[string]string{"capture": "yolo"}
				case "composition-policy":
					policy.OverlaysAllowed = true
				case "store-bytes":
					_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
					if err != nil {
						return nil, err
					}
					for _, member := range lock.Members {
						if member.Kind == "skill" {
							root := newGitManager(fx.home).entryPath(fx.home, resolvedOf(member))
							p := filepath.Join(root, "SKILL.md")
							if err := os.Chmod(p, 0600); err != nil {
								return nil, err
							}
							if err := os.WriteFile(p, []byte("changed"), 0600); err != nil {
								return nil, err
							}
						}
					}
				}
				expected, _ = captureFixtureState(t, fx)
				return content, nil
			}
			requireCaptureRefusal(t, req, DiagMaterializeCaptureDrift)
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(expected, after) {
				t.Fatal("capture changed the planted source or credential state")
			}
		})
	}
}

// Real executable discovery is planted on PATH; capture and metadata status
// succeed without executing it when Detect is nil. Muse native/file and both
// isolation modes retain the v4 file projection and leave source state unchanged.
func TestMaterializeCaptureAndStatusNeverProbeVendor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell sentinel is Unix-only")
	}
	for _, tc := range []struct {
		name, environment, isolation string
		selection                    envregistry.CredentialStoreSelection
	}{
		{name: "codex", environment: envregistry.CodexCLI},
		{"muse-native-shared", envregistry.Muse, envregistry.IsolationShared, envregistry.CredentialStoreNative},
		{"muse-file-shared", envregistry.Muse, envregistry.IsolationShared, envregistry.CredentialStoreFile},
		{"muse-native-isolated", envregistry.Muse, envregistry.IsolationIsolated, envregistry.CredentialStoreNative},
		{"muse-file-isolated", envregistry.Muse, envregistry.IsolationIsolated, envregistry.CredentialStoreFile},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx, req := captureFixture(t)
			if tc.environment == envregistry.Muse {
				fx.native[envregistry.Muse] = t.TempDir()
				if err := os.WriteFile(filepath.Join(fx.native[envregistry.Muse], "auth.json"), []byte("synthetic-muse\n"), 0600); err != nil {
					t.Fatal(err)
				}
				provision := fx.request(envregistry.Muse)
				provision.Machine.Isolation[fx.profile] = map[string]string{envregistry.Muse: tc.isolation}
				provision.Machine.CredentialStore = map[string]map[string]envregistry.CredentialStoreSelection{fx.profile: {envregistry.Muse: tc.selection}}
				provision.Repair = true
				if _, err := Resolve(provision); err != nil {
					t.Fatal(err)
				}
				req.Resolve = provision
				req.Resolve.Repair = false
				req.Resolve.Detect = nil
				req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
					return provision.Machine, Policy{}, nil
				}
				req.Capture = func(c *MaterializeCapture) ([]byte, error) {
					path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.Muse), "config", "muse", "AGENTS.md")
					path, err := filepath.EvalSymlinks(path)
					if err != nil {
						return nil, err
					}
					file, err := openReadNoFollow(path)
					if err != nil {
						return nil, err
					}
					return c.ReadObject(path, file)
				}
			}
			before, _ := captureFixtureState(t, fx)
			bin := t.TempDir()
			sentinel := filepath.Join(t.TempDir(), "ran")
			for _, tool := range []string{"codex", "claude", "muse", "pi", "opencode"} {
				if err := os.WriteFile(filepath.Join(bin, tool), []byte("#!/bin/sh\nprintf ran > '"+sentinel+"'\nprintf '1.2.3\\n'\n"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			for _, form := range []string{"export", "manifest-only"} {
				t.Run(form, func(t *testing.T) {
					captured, err := CaptureMaterialize(req)
					if err != nil {
						t.Fatal(err)
					}
					if tc.environment == envregistry.Muse {
						var fragment struct {
							Fragment        string
							Env             map[string]string
							CredentialStore struct{ Selection string } `json:"credential_store"`
						}
						if err := json.Unmarshal(captured.Fragment, &fragment); err != nil {
							t.Fatal(err)
						}
						if fragment.Fragment != "launch-env-fragment-v4" || fragment.Env["TBH_CREDENTIAL_BACKEND"] != "file" || fragment.CredentialStore.Selection != "file" {
							t.Fatalf("capture lost Muse file projection: %s", captured.Fragment)
						}
					}
				})
			}
			status, err := StatusOf(StatusRequest{Home: fx.home, Machine: req.Resolve.Machine, NativeHomeOf: req.Resolve.NativeHomeOf, OperatorXDG: fx.xdg, LaunchDir: fx.launch})
			if err != nil {
				t.Fatal(err)
			}
			for _, adapter := range status.Adapters {
				if adapter.Detected != "unknown" {
					t.Fatalf("unsafe release observation: %v", adapter)
				}
			}
			row := findHome(status, fx.profile, tc.environment)
			if row == nil || !row.Current {
				t.Fatalf("metadata status lost current home: %+v", row)
			}
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(before, after) {
				t.Fatal("capture or status changed source or credential state")
			}
			if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
				t.Fatal("vendor executable ran")
			}
		})
	}
}

// Missing source homes and missing manager state cannot be provisioned by an
// observation; an already provisioned source succeeds in the neighboring test.
func TestMaterializeCaptureDoesNotProvision(t *testing.T) {
	for _, exists := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "empty"}[exists], func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "manager")
			if exists {
				if err := os.Mkdir(home, 0700); err != nil {
					t.Fatal(err)
				}
			}
			req := MaterializeCaptureRequest{Resolve: ResolveRequest{Home: home, EnvID: envregistry.CodexCLI}, ReloadMachine: func() (envregistry.MachineConfig, Policy, error) {
				t.Fatal("selected unprovisioned source")
				return envregistry.MachineConfig{}, Policy{}, nil
			}, Capture: func(*MaterializeCapture) ([]byte, error) { t.Fatal("captured absent source"); return nil, nil }}
			before, _ := json.Marshal(captureTree(t, root))
			requireCaptureRefusal(t, req, DiagHomeStale)
			after, _ := json.Marshal(captureTree(t, root))
			if !bytes.Equal(before, after) {
				t.Fatal("created source state")
			}
		})
	}
}

// An object exceeding the capture byte quota refuses before reading its bytes.
// This local limit is not an archive-size qualification.
func TestMaterializeCaptureObjectBounds(t *testing.T) {
	_, req := captureFixture(t)
	req.Capture = func(c *MaterializeCapture) ([]byte, error) {
		path := filepath.Join(t.TempDir(), "large")
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		if err := f.Truncate(captureMaxBytes + 1); err != nil {
			return nil, err
		}
		return c.ReadObject(path, f)
	}
	requireCaptureRefusal(t, req, DiagMaterializeCaptureDrift)
}

// Real profile use and scope sync wait behind capture, then complete against
// isolated native homes. Multi-file output remains one verified generation.
func TestMaterializeCaptureSerializesProfileWriters(t *testing.T) {
	for _, writer := range []string{"use", "sync"} {
		t.Run(writer, func(t *testing.T) {
			fx, req := captureFixture(t)
			t.Setenv("CLAUDE_CONFIG_DIR", fx.native[envregistry.ClaudeCode])
			t.Setenv("CODEX_HOME", fx.native[envregistry.CodexCLI])
			t.Setenv("XDG_CONFIG_HOME", fx.xdg)
			t.Setenv("PI_CODING_AGENT_DIR", fx.native[envregistry.Pi])
			completed := make(chan error, 1)
			original := req.Capture
			req.Capture = func(c *MaterializeCapture) ([]byte, error) {
				first, err := original(c)
				if err != nil {
					return nil, err
				}
				path, err := filepath.EvalSymlinks(filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.Resolve.EnvID), ".agent-context", "system-prompt.md"))
				if err != nil {
					return nil, err
				}
				file, err := openReadNoFollow(path)
				if err != nil {
					return nil, err
				}
				second, err := c.ReadObject(path, file)
				if err != nil {
					return nil, err
				}
				go func() {
					var err error
					if writer == "use" {
						_, err = UseWithPolicy(fx.home, fx.profile, envregistry.CodexCLI, "", false, Policy{})
					} else {
						_, err = SyncWithPolicy(fx.home, Policy{})
					}
					completed <- err
				}()
				select {
				case err := <-completed:
					return nil, errors.New("writer escaped capture lock: " + fmt.Sprint(err))
				case <-time.After(30 * time.Millisecond):
				}
				return append(first, second...), nil
			}
			captured, err := CaptureMaterialize(req)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(captured.Content, []byte("hello")) || !bytes.Contains(captured.Content, []byte("Be terse.")) {
				t.Fatal("torn content")
			}
			if err := <-completed; err != nil {
				t.Fatal("writer after capture:", err)
			}
		})
	}
}

// A repair that sees an external multi-file change cannot race its writes into
// capture: capture refuses drift, releases the lock, then real repair restores
// a current source which the next invocation can capture.
func TestMaterializeCaptureConcurrentRepair(t *testing.T) {
	fx, req := captureFixture(t)
	original := req.Capture
	repair := fx.request(envregistry.CodexCLI)
	repair.Repair = true
	waiting := make(chan struct{})
	repair.beforeRepairLock = func() { close(waiting) }
	completed := make(chan error, 1)
	req.Capture = func(c *MaterializeCapture) ([]byte, error) {
		first, err := original(c)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(c.objects[len(c.objects)-1].path, []byte("external generation"), 0600); err != nil {
			return nil, err
		}
		go func() { _, err := Resolve(repair); completed <- err }()
		select {
		case <-waiting:
		case <-t.Context().Done():
			return nil, errors.New("repair did not reach mutation lock")
		}
		select {
		case err := <-completed:
			return nil, errors.New("repair escaped capture lock: " + fmt.Sprint(err))
		case <-time.After(30 * time.Millisecond):
		}
		return first, nil
	}
	requireCaptureRefusal(t, req, DiagMaterializeCaptureDrift)
	if err := <-completed; err != nil {
		t.Fatal("repair after refusal:", err)
	}
	req.Capture = original
	if _, err := CaptureMaterialize(req); err != nil {
		t.Fatal("repaired positive control:", err)
	}
}

// Captured store objects keep their byte, executable-mode and file identities
// even when the profile fragment is unchanged. Stable objects are admitted.
func TestMaterializeCaptureStoreObjectIdentity(t *testing.T) {
	for _, change := range []string{"stable", "identity", "mode", "bytes"} {
		t.Run(change, func(t *testing.T) {
			fx, req := captureFixture(t)
			_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			var path string
			for _, member := range lock.Members {
				if member.Kind == "skill" {
					path = filepath.Join(newGitManager(fx.home).entryPath(fx.home, resolvedOf(member)), "SKILL.md")
				}
			}
			var expected []byte
			req.Capture = func(c *MaterializeCapture) ([]byte, error) {
				f, err := openReadNoFollow(path)
				if err != nil {
					return nil, err
				}
				data, err := c.ReadObject(path, f)
				if err != nil {
					return nil, err
				}
				info, err := os.Stat(path)
				if err != nil {
					return nil, err
				}
				switch change {
				case "identity":
					replacement := path + ".replacement"
					if err := os.WriteFile(replacement, data, info.Mode()); err != nil {
						return nil, err
					}
					if err := os.Chtimes(replacement, info.ModTime(), info.ModTime()); err != nil {
						return nil, err
					}
					if err := replaceCapturedFixture(t, replacement, path); err != nil {
						return nil, err
					}
				case "mode":
					if runtime.GOOS == "windows" {
						t.Skip("Windows does not expose portable executable permission bits")
					}
					if err := os.Chmod(path, info.Mode()^0100); err != nil {
						return nil, err
					}
				case "bytes":
					if err := os.Chmod(path, 0600); err != nil {
						return nil, err
					}
					if err := os.WriteFile(path, []byte("different"), 0600); err != nil {
						return nil, err
					}
				}
				expected, _ = captureFixtureState(t, fx)
				return data, nil
			}
			if change == "stable" {
				if _, err := CaptureMaterialize(req); err != nil {
					t.Fatal(err)
				}
			} else {
				requireCaptureRefusal(t, req, DiagMaterializeCaptureDrift)
				after, _ := captureFixtureState(t, fx)
				if !bytes.Equal(expected, after) {
					t.Fatal("capture changed the planted source or credential state")
				}
			}
		})
	}
}

// Failed final reads and same-byte replacements of authorizing records cannot
// be hidden by a valid, equal fragment. Callback mutations are the only writes.
func TestMaterializeCaptureAuthorizingDrift(t *testing.T) {
	for _, change := range []string{"policy-read", "selection", "marker-identity", "lock-identity", "origin-identity", "origin-bytes", "origin-mode"} {
		t.Run(change, func(t *testing.T) {
			fx, req := captureFixture(t)
			if _, err := CaptureMaterialize(req); err != nil {
				t.Fatal(err)
			}
			original := req.Capture
			if change == "selection" {
				req.Resolve.Profile = ""
			}
			calls := 0
			req.ReloadMachine = func() (envregistry.MachineConfig, Policy, error) {
				calls++
				if change == "policy-read" && calls == 2 {
					return envregistry.MachineConfig{}, Policy{}, os.ErrPermission
				}
				return envregistry.DefaultMachineConfig(), Policy{}, nil
			}
			var expected []byte
			var originPath string
			var planted envunit.Image
			req.Capture = func(c *MaterializeCapture) ([]byte, error) {
				data, err := original(c)
				if err != nil {
					return nil, err
				}
				switch change {
				case "selection":
					// Deliberate noncooperating fixture edit under the capture lock.
					if err := os.WriteFile(CurrentFile(fx.home), []byte("replacement-current\n"), 0600); err != nil {
						return nil, err
					}
				case "marker-identity", "lock-identity", "origin-identity", "origin-bytes", "origin-mode":
					var path string
					if strings.HasPrefix(change, "origin-") {
						path = filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.Resolve.EnvID), envorigin.Name)
					} else if change == "marker-identity" {
						path = c.objects[2].path
					} else {
						paths, err := filepath.Glob(filepath.Join(fx.home, "state", "locks", "v1", "*.lock"))
						if err != nil || len(paths) != 1 {
							return nil, fmt.Errorf("lock fixture: %v %v", paths, err)
						}
						path = paths[0]
					}
					info, err := os.Stat(path)
					if err != nil {
						return nil, err
					}
					var payload []byte
					if change == "marker-identity" || strings.HasPrefix(change, "origin-") {
						payload, err = os.ReadFile(path)
						if err != nil {
							return nil, err
						}
					} else if info.Size() != 0 {
						return nil, errors.New("lock fixture is not empty")
					}
					if change == "origin-mode" {
						if err := os.Chmod(path, 0400); err != nil {
							return nil, err
						}
						expected, _ = captureFixtureState(t, fx)
						if strings.HasPrefix(change, "origin-") {
							originPath = filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.Resolve.EnvID), envorigin.Name)
							planted, err = envunit.ObserveEntry(originPath, false)
							if err != nil {
								return nil, err
							}
						}
						return data, nil
					}
					if change == "origin-bytes" {
						var record envorigin.Record
						if err := json.Unmarshal(payload, &record); err != nil {
							return nil, err
						}
						record.Capture.Time = "2026-10-28T00:00:00Z"
						payload, err = record.Marshal()
						if err != nil {
							return nil, err
						}
					}
					replacement := path + ".new"
					if err := os.WriteFile(replacement, payload, info.Mode()); err != nil {
						return nil, err
					}
					if err := os.Chtimes(replacement, info.ModTime(), info.ModTime()); err != nil {
						return nil, err
					}
					if err := replaceCapturedFixture(t, replacement, path); err != nil {
						return nil, err
					}
				}
				expected, _ = captureFixtureState(t, fx)
				if strings.HasPrefix(change, "origin-") {
					originPath = filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.Resolve.EnvID), envorigin.Name)
					planted, err = envunit.ObserveEntry(originPath, false)
					if err != nil {
						return nil, err
					}
				}
				return data, nil
			}
			requireCaptureRefusal(t, req, DiagMaterializeCaptureDrift)
			if originPath != "" {
				got, err := envunit.ObserveEntry(originPath, false)
				if err != nil || got != planted {
					t.Fatalf("private origin plant overwritten: %+v %+v %v", planted, got, err)
				}
			}
			after, _ := captureFixtureState(t, fx)
			if !bytes.Equal(expected, after) {
				t.Fatal("refusal mutated source")
			}
		})
	}
}

// A denied journal read is unreadable state, never absence or a recovery
// request. A current source without a journal is the adjacent positive control.
func TestMaterializeCaptureUnreadableJournal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode-bit unreadability")
	}
	fx, req := captureFixture(t)
	if _, err := CaptureMaterialize(req); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(fx.home, "state", "transactions", "v1")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "denied.json")
	if err := os.WriteFile(path, []byte("{}"), 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0600) })
	if file, err := os.Open(path); err == nil {
		_ = file.Close()
		t.Skip("this environment can read a mode-000 file")
	}
	requireCaptureRefusal(t, req, "manager_state_unreadable")
	info, err := os.Lstat(path)
	if err != nil || info.Mode().Perm() != 0 {
		t.Fatal("refusal changed journal permissions or removed it")
	}
}

// A deterministic external writer changes both content objects between reads.
// The private old/new mixture must never escape as successful capture, and
// refusal preserves the externally planted generation without repairing it.
func TestMaterializeCaptureExternalTornGeneration(t *testing.T) {
	fx, req := captureFixture(t)
	firstReader := req.Capture
	secondPath, err := filepath.EvalSymlinks(filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.Resolve.EnvID), ".agent-context", "system-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	mutate := false
	var expected []byte
	req.Capture = func(c *MaterializeCapture) ([]byte, error) {
		first, err := firstReader(c)
		if err != nil {
			return nil, err
		}
		if mutate {
			changed := make(chan error, 1)
			firstPath := c.objects[len(c.objects)-1].path
			go func() {
				if err := os.WriteFile(firstPath, []byte("new root generation"), 0600); err != nil {
					changed <- err
					return
				}
				changed <- os.WriteFile(secondPath, []byte("new system generation"), 0600)
			}()
			if err := <-changed; err != nil {
				return nil, err
			}
		}
		file, err := openReadNoFollow(secondPath)
		if err != nil {
			return nil, err
		}
		second, err := c.ReadObject(secondPath, file)
		if err != nil {
			return nil, err
		}
		expected, _ = captureFixtureState(t, fx)
		return append(first, second...), nil
	}
	captured, err := CaptureMaterialize(req)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(captured.Content, []byte("hello")) || !bytes.Contains(captured.Content, []byte("Be terse.")) {
		t.Fatal("stable control missed a content object")
	}
	mutate = true
	requireCaptureRefusal(t, req, DiagMaterializeCaptureDrift)
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(expected, after) {
		t.Fatal("refusal repaired or mutated the external generation")
	}
}

// Capture must not reuse a known-release policy for metadata-only verification:
// a now-invalid native selector refuses without output, callback or state writes.
func TestMaterializeCaptureCredentialPolicyReleaseObservation(t *testing.T) {
	fx, req := captureFixture(t)
	if _, err := CaptureMaterialize(req); err != nil {
		t.Fatal(err)
	}
	adapter, err := envregistry.ByID(req.Resolve.EnvID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := req.Resolve.resolveCredentialPolicyObserved(adapter, adapter.VerifiedRelease); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fx.native[adapter.ID], "config.toml")
	if err := os.WriteFile(path, []byte("cli_auth_credentials_store = 'invalid'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	before, _ := captureFixtureState(t, fx)
	req.Capture = func(*MaterializeCapture) ([]byte, error) {
		t.Fatal("invalid native selector reached capture")
		return nil, nil
	}
	got, err := CaptureMaterialize(req)
	if err == nil || got != nil || !strings.Contains(err.Error(), DiagHomeStale) || !strings.Contains(err.Error(), envregistry.DiagCredentialUnsupported) {
		t.Fatalf("capture = %v, %v; want native-selector refusal with no output", got, err)
	}
	after, _ := captureFixtureState(t, fx)
	if !bytes.Equal(before, after) {
		t.Fatal("release-cache refusal changed source or credential state")
	}
}

// The shared policy cache keys both successes and errors by supplied release.
// macOS Claude changes isolation at the pinned release; other platforms retain
// their shared default but must still obtain a fresh projection per observation.
func TestMaterializeCaptureCredentialPolicyReleaseCache(t *testing.T) {
	fx := writeManagedFixture(t, "release-cache")
	req := fx.request(envregistry.ClaudeCode)
	req.Detect = func(envregistry.Adapter) string {
		t.Fatal("supplied release triggered vendor observation")
		return "unknown"
	}
	adapter, err := envregistry.ByID(envregistry.ClaudeCode)
	if err != nil {
		t.Fatal(err)
	}
	var previous *effectiveCredentialPolicy
	for _, release := range []string{"0.0.1", "unknown", adapter.VerifiedRelease, "0.0.1"} {
		p, err := req.resolveCredentialPolicyObserved(adapter, release)
		want := envregistry.IsolationShared
		if runtime.GOOS == "darwin" && release != "0.0.1" {
			want = envregistry.IsolationIsolated
		}
		if err != nil || p == nil || p.isolation != want || p == previous {
			t.Fatalf("release %s: policy=%+v err=%v; want fresh %s", release, p, err, want)
		}
		again, err := req.resolveCredentialPolicyObserved(adapter, release)
		if err != nil || again != p {
			t.Fatalf("same release did not reuse projection: %v", err)
		}
		previous = p
	}
	req.Machine.Isolation[fx.profile] = map[string]string{adapter.ID: envregistry.IsolationIsolated}
	req.credentialPolicy = nil
	_, err = req.resolveCredentialPolicyObserved(adapter, "0.0.1")
	if runtime.GOOS == "darwin" && (err == nil || !strings.Contains(err.Error(), envregistry.DiagIsolatedUnsupported)) {
		t.Fatalf("below-pinned isolation must refuse: %v", err)
	}
	p, err := req.resolveCredentialPolicyObserved(adapter, "unknown")
	if err != nil || p == nil || p.isolation != envregistry.IsolationIsolated {
		t.Fatalf("old refusal reused for unknown release: %+v, %v", p, err)
	}
}

// Capture retains private recorded A under a caller selecting B; its portable
// result contains neither private origin nor lifecycle authority. No discovery
// callback is permitted, and the fixture callback reaches actual content reads.
func TestMaterializeCaptureRecordedOrigin(t *testing.T) {
	fx, req := captureFixture(t)
	originPath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, req.Resolve.EnvID), envorigin.Name)
	before := nativeRead(t, originPath)
	t.Setenv("CODEX_HOME", t.TempDir())
	req.Resolve.originAccountLookup = func() (envorigin.Account, error) {
		t.Fatal("capture discovered caller B")
		return envorigin.Account{}, errors.New("forbidden lookup")
	}
	reached := false
	capture := req.Capture
	req.Capture = func(c *MaterializeCapture) ([]byte, error) { reached = true; return capture(c) }
	got, err := CaptureMaterialize(req)
	if err != nil || got == nil || !reached || len(got.Content) == 0 {
		t.Fatalf("recorded A capture: reached=%t %v", reached, err)
	}
	if !bytes.Equal(nativeRead(t, originPath), before) {
		t.Fatal("capture changed origin authority")
	}
	for _, output := range [][]byte{got.Fragment, got.Content} {
		if bytes.Contains(output, []byte(envorigin.Version)) || bytes.Contains(output, []byte(envlifecycle.Version)) {
			t.Fatal("portable capture leaked private authority")
		}
	}
}
