package install

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/relux-works/curator/internal/managerlock"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/transaction"
)

// The actual global owner publishes external namespaces/trees with profile,
// native and consumer effects. Acquisition/compiler/process-list transports
// are synthetic; no native provider, real account or network is exercised.
// Both configured and default stores remain absent through Prepare. A fault
// after namespace or artifact publication restores absence and first-use
// profile state; a healthy neighbor proves protected final lookups succeed.
func TestProvisioningOriginExternalPublication(t *testing.T) {
	for _, custom := range []bool{false, true} {
		for _, fault := range []string{"healthy", "namespace", "artifact", "artifact-stop"} {
			t.Run(map[bool]string{false: "default", true: "configured"}[custom]+"/"+fault, func(t *testing.T) {
				e := newEnv(t)
				writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1"})
				e.globalDeclare("golden-skill")
				user := t.TempDir()
				t.Setenv("HOME", user)
				t.Setenv("CODEX_HOME", filepath.Join(user, ".codex"))
				t.Setenv("CLAUDE_CONFIG_DIR", filepath.Join(user, ".claude"))
				t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(user, ".pi", "agent"))
				t.Setenv("XDG_CONFIG_HOME", filepath.Join(user, ".config"))
				t.Setenv("XDG_DATA_HOME", filepath.Join(user, ".local", "share"))
				t.Setenv("XDG_STATE_HOME", filepath.Join(user, ".local", "state"))
				deps, _ := lifecycleBuildDeps(t)
				external := lifecycleExternalDeps()
				root := filepath.Join(e.home, "external-build-cache")
				if custom {
					root = filepath.Join(t.TempDir(), "external-cache")
					external.StoreRoot = root
				}
				assertAbsent := func(path string) {
					t.Helper()
					if _, err := os.Lstat(path); !os.IsNotExist(err) {
						t.Fatalf("forbidden early/failed publication at %s: %v", path, err)
					}
				}
				staged, prepared, hit, collected := false, false, false, false
				planted := errors.New("external-unit-fault")
				opts := Options{Platform: installPlatform(), Build: deps, External: external, OnStaged: func(Staged) error { staged = true; assertAbsent(root); return nil }, Commit: CommitDeps{Collect: func(req scopes.MaintenanceRequest) (scopes.MaintenanceResult, error) {
					collected = true
					cache, err := buildcache.New(req.Home)
					if err != nil {
						return scopes.MaintenanceResult{}, err
					}
					cache.ExecutablePaths = func() ([]string, error) { return nil, nil }
					req.Cache = cache
					return scopes.Collect(req)
				}, Hooks: transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point == transaction.PointPrepared {
						raw, err := os.ReadFile(filepath.Join(e.home, "state", "transactions", "v1", event.TransactionID+".json"))
						if err != nil {
							t.Fatal(err)
						}
						var journal transaction.Journal
						if err = json.Unmarshal(raw, &journal); err != nil {
							t.Fatal(err)
						}
						if journal.Unit != nil && journal.Unit.Legacy != nil {
							prepared = true
							assertAbsent(root)
							n := 0
							for _, target := range journal.Targets {
								if target.Class == "05-external-cache" || target.Class == "05-external-snapshot" {
									n++
									if !target.PrivateTree {
										t.Fatal("external tree lacks durable protection")
									}
								}
							}
							if n != 2 {
								t.Fatalf("external journal omitted artifact/snapshot: %d", n)
							}
						}
					}
					if fault == "artifact-stop" && event.Point == transaction.PointAfterInstall && event.Class == "05-external-cache" {
						hit = true
						panic(planted)
					}
					if fault == "namespace" && event.Point == transaction.PointAfterDirectoryInstall && event.LivePath == root || fault == "artifact" && event.Point == transaction.PointAfterInstall && event.Class == "05-external-cache" {
						hit = true
						return planted
					}
					return nil
				}}}}
				var result Result
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							if fault != "artifact-stop" || recovered != planted {
								panic(recovered)
							}
						}
					}()
					result = Global(e.cfg, user, opts)
				}()
				if fault == "artifact-stop" {
					if !hit || !prepared || collected {
						t.Fatalf("restart stop not reached: hit=%t prepared=%t collected=%t", hit, prepared, collected)
					}
					manager, err := managerlock.New(e.home)
					if err != nil {
						t.Fatal(err)
					}
					held, err := manager.AcquireHomeOnly(context.Background(), false)
					if err != nil {
						t.Fatal(err)
					}
					engine, err := transaction.New(e.home)
					if err == nil {
						err = engine.Recover(held)
					}
					closeErr := held.Close()
					if err != nil || closeErr != nil {
						t.Fatalf("fresh global compensation: %v %v", err, closeErr)
					}
					assertAbsent(root)
					assertAbsent(filepath.Join(e.home, "profiles", "default", "source.json"))
					assertAbsent(filepath.Join(e.home, "profiles", "default", "lock.json"))
					assertAbsent(filepath.Join(e.home, "global", "env.sh"))
					return
				}
				if !staged || !prepared {
					t.Fatalf("actual joined external owner not reached: %+v", result)
				}
				if fault != "healthy" {
					if !hit || result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), planted.Error()) || collected {
						t.Fatalf("wrong external refusal/continuation: %+v hit=%t collected=%t", result, hit, collected)
					}
					assertAbsent(root)
					assertAbsent(filepath.Join(e.home, "profiles", "default", "source.json"))
					assertAbsent(filepath.Join(e.home, "profiles", "default", "lock.json"))
					assertAbsent(filepath.Join(e.home, "global", "env.sh"))
					return
				}
				if result.Status != "ok" || !collected {
					t.Fatalf("healthy external publication: %+v collected=%t", result, collected)
				}
				entries, err := os.ReadDir(filepath.Join(root, buildrepo.ArtifactsDir(2)))
				if err != nil || len(entries) != 1 {
					t.Fatalf("artifact entry: %v %v", entries, err)
				}
				if err = (&buildrepo.DiskProtectedStore{Root: root}).VerifyArtifact(2, "sha256:"+entries[0].Name()); err != nil {
					t.Fatal(err)
				}
				entries, err = os.ReadDir(filepath.Join(root, "snapshots"))
				if err != nil || len(entries) != 1 {
					t.Fatalf("snapshot entry: %v %v", entries, err)
				}
				if err = (&buildrepo.DiskProtectedStore{Root: root}).VerifySnapshot("sha256:" + entries[0].Name()); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
