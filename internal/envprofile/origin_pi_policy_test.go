package envprofile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
)

// Initial capture uses only a synthetic raw account and selector. The Pi
// selector denotes the agent root itself, independently of native skill custody.
func piOriginRequest(t *testing.T, live, isolated bool, label string) (*managedFixture, ResolveRequest, string) {
	t.Helper()
	fx := writeManagedFixture(t, "pi-origin-fixture")
	anchor := t.TempDir()
	agent := filepath.Join(t.TempDir(), label, "pi", "agent")
	if err := os.MkdirAll(agent, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	if live {
		if err := os.WriteFile(filepath.Join(agent, "auth.json"), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	req := fx.request("pi")
	req.Repair = true
	if isolated {
		req.Machine.Isolation = map[string]map[string]string{fx.profile: {"pi": envregistry.IsolationIsolated}}
	}
	req.originAccountLookup = func() (envorigin.Account, error) {
		return envorigin.Account{UID: uint64(os.Geteuid()), RealUID: uint64(os.Getuid()), EffectiveUID: uint64(os.Geteuid()), Home: anchor}, nil
	}
	return fx, req, agent
}

// Adopted Pi shared dangling policy refuses before first publication, while a
// live shared agent root and isolated login-pending home reach real Resolve.
// This is not an OS-account/provider qualification or a credential-content read.
func TestProvisioningOriginPiInitialPolicy(t *testing.T) {
	enableV1WritersForTest(t)
	for _, tc := range []struct {
		name           string
		live, isolated bool
	}{{"shared-live", true, false}, {"shared-absent", false, false}, {"isolated-pending", false, true}} {
		t.Run(tc.name, func(t *testing.T) {
			fx, req, agent := piOriginRequest(t, tc.live, tc.isolated, "a")
			before := nativeSnapshot(t, fx.home)
			result, err := Resolve(req)
			target := ManagedHomeDir(fx.home, fx.profile, "pi")
			if !tc.live && !tc.isolated {
				if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) || !strings.Contains(err.Error(), "reason=dangling") {
					t.Fatalf("Pi shared absent must refuse before publication: %v", err)
				}
				if result != nil && len(result.Document) > 0 {
					t.Fatal("Pi refusal emitted fragment")
				}
				if nativeSnapshot(t, fx.home) != before {
					t.Fatal("Pi refusal mutated manager preimages")
				}
				return
			}
			if err != nil || result == nil || len(result.Document) == 0 {
				t.Fatalf("valid Pi initial continuation: %v", err)
			}
			if tc.isolated {
				if _, err := os.Lstat(filepath.Join(target, "auth.json")); !os.IsNotExist(err) {
					t.Fatalf("isolated Pi gained shared auth: %v", err)
				}
				return
			}
			link, err := os.Readlink(filepath.Join(target, "auth.json"))
			if err != nil || link != filepath.Join(agent, "auth.json") {
				t.Fatalf("Pi agent-root target: %q %v", link, err)
			}
		})
	}
}

// Recorded A is preserved under caller B in both label directions. Bare read,
// status and repair distinguish missing native leaf, old/double-agent/wrong B
// links and credential forks. Valid repair restores only an absent/empty entry.
// Snapshots cover owned synthetic data; target/account seams observe raw inputs.
func TestProvisioningOriginPiRecordedPolicy(t *testing.T) {
	enableV1WritersForTest(t)
	for _, direction := range []string{"a-to-b", "b-to-a"} {
		t.Run(direction, func(t *testing.T) {
			for _, state := range []string{"healthy", "native-absent", "managed-absent", "empty-directory", "old-target", "double-agent", "caller-target", "regular-fork", "unreadable-target"} {
				t.Run(state, func(t *testing.T) {
					fx, req, agent := piOriginRequest(t, true, false, strings.Split(direction, "-to-")[0])
					result, err := Resolve(req)
					if err != nil || result == nil || len(result.Document) == 0 {
						t.Fatalf("live Pi setup: %v", err)
					}
					target := ManagedHomeDir(fx.home, fx.profile, "pi")
					link := filepath.Join(target, "auth.json")
					origin, err := os.ReadFile(filepath.Join(target, envorigin.Name))
					if err != nil {
						t.Fatal(err)
					}
					b := filepath.Join(t.TempDir(), strings.Split(direction, "-to-")[1])
					if err := os.Mkdir(b, 0700); err != nil {
						t.Fatal(err)
					}
					t.Setenv("PI_CODING_AGENT_DIR", b)
					req.originAccountLookup = func() (envorigin.Account, error) {
						t.Error("recorded Pi rediscovered account")
						return envorigin.Account{}, fmt.Errorf("forbidden account lookup")
					}
					if state == "native-absent" {
						if err = os.Remove(filepath.Join(agent, "auth.json")); err != nil {
							t.Fatal(err)
						}
					}
					if state != "healthy" && state != "native-absent" && state != "unreadable-target" {
						if err = os.Remove(link); err != nil {
							t.Fatal(err)
						}
						switch state {
						case "empty-directory":
							err = os.Mkdir(link, 0700)
						case "old-target":
							err = os.Symlink(filepath.Join(filepath.Dir(agent), "auth.json"), link)
						case "double-agent":
							err = os.Symlink(filepath.Join(agent, "agent", "auth.json"), link)
						case "caller-target":
							err = os.Symlink(filepath.Join(b, "auth.json"), link)
						case "regular-fork":
							err = os.WriteFile(link, nil, 0600)
						}
						if err != nil {
							t.Fatal(err)
						}
					}
					targetReads := 0
					req.passthroughTargetStat = func(path string) (os.FileInfo, error) {
						targetReads++
						if path != filepath.Join(agent, "auth.json") {
							t.Fatalf("credential observation escaped recorded A: %s", path)
						}
						if state == "unreadable-target" {
							return nil, &os.PathError{Op: "stat", Path: path, Err: errors.New("fixture metadata denied")}
						}
						return os.Stat(path)
					}
					status, statusErr := StatusOf(StatusRequest{Home: fx.home, Machine: req.Machine, Detect: req.Detect, NativeHomeOf: req.NativeHomeOf, OperatorXDG: fx.xdg, LaunchDir: fx.launch, passthroughTargetStat: req.passthroughTargetStat})
					if statusErr != nil {
						t.Fatal(statusErr)
					}
					row := findHome(status, fx.profile, "pi")
					if row == nil || row.Current != (state == "healthy") {
						t.Fatalf("Pi status currency %s: %+v", state, row)
					}
					if state == "native-absent" {
						for _, want := range []string{envregistry.DiagPassthroughDetached, envregistry.DiagCredentialConflict, "reason=dangling"} {
							if !strings.Contains(strings.Join(row.Findings, ";"), want) {
								t.Fatalf("Pi status lost %s: %+v", want, row)
							}
						}
					}
					if state == "unreadable-target" && !strings.Contains(strings.Join(row.Findings, ";"), envregistry.DiagPassthroughUnreadable) {
						t.Fatalf("Pi unreadable status: %+v", row)
					}
					before := nativeSnapshot(t, target)
					nativeBefore := nativeSnapshot(t, agent)
					for _, mode := range []string{"plain", "dry-run", "repair"} {
						req.Repair = mode != "plain"
						req.DryRun = mode == "dry-run"
						got, e := Resolve(req)
						repairable := state == "managed-absent" || state == "empty-directory"
						success := state == "healthy" || repairable && mode == "repair"
						if success {
							if e != nil || got == nil || len(got.Document) == 0 {
								t.Fatalf("valid recorded Pi %s continuation: %v", mode, e)
							}
							if want := filepath.Join(agent, "auth.json"); !req.DryRun {
								if actual, e := os.Readlink(link); e != nil || actual != want {
									t.Fatalf("repair target %q %v", actual, e)
								}
							}
						} else {
							want := envregistry.DiagCredentialConflict
							if mode == "plain" || repairable {
								want = DiagHomeStale
							}
							if state == "unreadable-target" {
								want = envregistry.DiagPassthroughUnreadable
							}
							if e == nil || !strings.Contains(e.Error(), want) {
								t.Fatalf("Pi %s/%s intended %s: %v", state, mode, want, e)
							}
							if state == "native-absent" && !strings.Contains(e.Error(), "reason=dangling") {
								t.Fatalf("Pi dangling reason lost: %v", e)
							}
							if got != nil && len(got.Document) > 0 {
								t.Fatal("Pi rejection emitted fragment")
							}
							if nativeSnapshot(t, target) != before {
								t.Fatal("Pi rejection changed managed preimages")
							}
						}
						if nativeSnapshot(t, agent) != nativeBefore {
							t.Fatal("Pi maintenance changed native fixture")
						}
						raw, e := os.ReadFile(filepath.Join(target, envorigin.Name))
						if e != nil || string(raw) != string(origin) {
							t.Fatal("Pi maintenance recaptured origin")
						}
					}
					if (state == "healthy" || state == "native-absent" || state == "unreadable-target") && targetReads == 0 {
						t.Fatal("Pi credential metadata owner was not reached")
					}
				})
			}
		})
	}
}
