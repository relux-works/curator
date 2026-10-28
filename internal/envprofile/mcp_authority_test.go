package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectmcpstate"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/transaction"
)

// A real installed, recorded neighbor supplies per-home authority. A node
// carrier is planted only after valid provisioning, never an orphan exemption.
func mcpAuthorityCarrier(t *testing.T, home string, node bool) {
	t.Helper()
	source := t.TempDir()
	writePackage(t, source, "authority-neighbor", "1.0.0", "fixture\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: source}); err != nil {
		t.Fatal(err)
	}
	native := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", native)
	fx := &managedFixture{home: home, profile: "authority-neighbor", native: map[string]string{"claude_code": native}, launch: t.TempDir(), xdg: t.TempDir(), accountAnchor: t.TempDir()}
	req := fx.request(envregistry.ClaudeCode)
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	if node {
		writeNodeAuthorityMarker(t, fx)
	}
	err := envauthority.Validate(home)
	if node {
		requireMCPError(t, err, envregistry.DiagCredentialInjectionUnavailable)
	} else if err != nil {
		t.Fatal(err)
	}
}

func observeMCPAuthorityCallbacks(t *testing.T, calls *int) {
	t.Helper()
	prior := newOperationEngine
	newOperationEngine = func(home string, options ...transaction.Option) (*transaction.Engine, error) {
		*calls++
		return prior(home, append(options, recoveryCounter(calls)...)...)
	}
	t.Cleanup(func() { newOperationEngine = prior })
}

// Every public profile mode refuses node authority before recovery, vendor
// qualification or publication callbacks and leaves the complete manager tree
// unchanged. Each per-home neighbor executes the same mode successfully.
// Snapshots cover bytes, inventory, types and links, not timestamps or ACLs.
func TestProfileMCPCredentialAuthority(t *testing.T) {
	for _, mode := range []string{"install", "status", "repair", "preview"} {
		for _, node := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/node=%v", mode, node), func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				mcpAuthorityCarrier(t, fx.home, node)
				req.Mode = mode
				if mode == "preview" {
					req.Mode, req.DryRun = "install", true
				}
				calls := 0
				observeMCPAuthorityCallbacks(t, &calls)
				req.qualified = func(string) bool { calls++; return true }
				req.beforePublish = func() { calls++ }
				before := nativeSnapshot(t, fx.home)
				result, err := ProfileMCP(req)
				if node {
					if err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialInjectionUnavailable+":") {
						t.Fatalf("profile MCP authority admission missing: %v", err)
					}
					if result != nil && result.Published || calls != 0 || nativeSnapshot(t, fx.home) != before {
						t.Fatal("profile MCP authority refusal wrote state or invoked a callback")
					}
				} else if err != nil || result == nil || (mode == "install" || mode == "repair") && !result.Published {
					t.Fatalf("per-home profile MCP control: %+v %v", result, err)
				}
			})
		}
	}
}

// Acquisition, publication, replay and restore each refuse the schema-valid
// node carrier without manager/project writes, a transaction callback or a Git
// invocation. Per-home controls consume a real locked Git fixture successfully.
// Git tracing is confined to the fixture and is enabled after fixture setup.
func TestProjectMCPCredentialAuthority(t *testing.T) {
	for _, entry := range []string{"resolve", "publish", "replay", "restore"} {
		for _, node := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/node=%v", entry, node), func(t *testing.T) {
				home, root, lock, decls, policy := lockedProjectMCP(t, Policy{}, false)
				mcpAuthorityCarrier(t, home, node)
				calls := 0
				observeMCPAuthorityCallbacks(t, &calls)
				trace := filepath.Join(t.TempDir(), "git-events.json")
				t.Setenv("GIT_TRACE2_EVENT", trace)
				before, projectBefore := nativeSnapshot(t, home), nativeSnapshot(t, root)
				var err error
				switch entry {
				case "resolve":
					_, _, err = ResolveProjectMCP(home, decls, policy)
				case "publish":
					var payload []byte
					payload, err = projectmcpstate.Canonical(lock.SkillsLock)
					if err == nil {
						err = PublishProjectMCPLock(home, root, lock, map[string][]byte{sourcelock.PathIn(root): payload})
					}
				case "replay":
					_, _, err = ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, policy)
				case "restore":
					_, _, err = RestoreProjectMCP(home, root, "project", lock.ManifestSHA256, decls, policy)
				}
				if node {
					if err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialInjectionUnavailable+":") {
						t.Fatalf("project MCP authority admission missing: %v", err)
					}
					if calls != 0 || nativeSnapshot(t, home) != before || nativeSnapshot(t, root) != projectBefore {
						t.Fatal("project MCP authority refusal wrote state or invoked a callback")
					}
					if body, e := os.ReadFile(trace); !os.IsNotExist(e) && (e != nil || len(body) != 0) {
						t.Fatalf("project MCP authority refusal invoked Git: %v", e)
					}
				} else if err != nil {
					t.Fatalf("per-home project MCP control: %v", err)
				} else if entry == "publish" {
					if calls == 0 {
						t.Fatal("per-home publication did not exercise transaction observation")
					}
				} else if body, e := os.ReadFile(trace); e != nil || len(body) == 0 {
					t.Fatalf("per-home Git observation control: %v", e)
				}
			})
		}
	}
}

// A node carrier introduced by the existing prepublication seam refuses before
// the writer creates private parents or stages payloads; per-home still writes.
// The exact active lock keeps its inventory/metadata coverage while Windows
// holds its byte-range lock; its bytes are compared before and after release.
func TestProfileMCPCredentialPublicationBoundary(t *testing.T) {
	for _, node := range []bool{true, false} {
		t.Run(fmt.Sprintf("node=%v", node), func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			mcpAuthorityCarrier(t, fx.home, false)
			lock := authorityTestLock(t, fx.home)
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			activeLock := claudeFreshnessLockPath(t, fx.home)
			lockBytes := nativeRead(t, activeLock)
			var expected map[string]claudeRefusalEntry
			req.beforePublish = func() {
				// The admitted neighbor already exists. Plant only its authority
				// transition under the held lock; never recursively Install here.
				if node {
					writeNodeAuthorityMarker(t, &managedFixture{home: fx.home, profile: "authority-neighbor"})
				}
				var err error
				expected, err = claudeFreshnessSnapshot(fx.home, activeLock)
				if err != nil {
					t.Fatal(err)
				}
			}
			result, err := ProfileMCP(req)
			if expected == nil {
				t.Fatal("MCP publication seam was not reached")
			}
			if !bytes.Equal(lockBytes, nativeRead(t, activeLock)) {
				t.Fatal("MCP publication changed manager lock bytes")
			}
			if node {
				if err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialInjectionUnavailable+":") {
					t.Fatalf("MCP publication authority recheck missing: %v", err)
				}
				after, snapshotErr := claudeFreshnessSnapshot(fx.home, activeLock)
				if snapshotErr != nil {
					t.Fatal(snapshotErr)
				}
				assertClaudeRefusalSnapshot(t, expected, after)
				if result.Published || t.Failed() {
					t.Fatal("MCP publication authority refusal wrote after the marker changed")
				}
			} else if err != nil || !result.Published {
				t.Fatalf("per-home publication control: %v", err)
			}
		})
	}
}

// Authority admission precedes even empty acquisition, schema processing or
// protected input reads. Per-home reaches an empty result, a lock-header
// refusal, or the missing protected-lock refusal, respectively.
func TestProjectMCPAuthorityBeforeState(t *testing.T) {
	for _, entry := range []string{"resolve", "publish", "replay", "restore"} {
		for _, node := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/node=%v", entry, node), func(t *testing.T) {
				home := gitFixtureHome(t)
				if err := privatedir.Protect(home); err != nil {
					t.Fatal(err)
				}
				mcpAuthorityCarrier(t, home, node)
				before := nativeSnapshot(t, home)
				var err error
				switch entry {
				case "resolve":
					_, _, err = ResolveProjectMCP(home, nil, Policy{})
				case "publish":
					err = PublishProjectMCPLock(home, home, &projectmcpstate.Lock{}, nil)
				case "replay":
					_, _, err = ReplayProjectMCP(home, home, "project", "", nil, Policy{})
				case "restore":
					_, _, err = RestoreProjectMCP(home, home, "project", "", nil, Policy{})
				}
				if node && (err == nil || !strings.HasPrefix(err.Error(), envregistry.DiagCredentialInjectionUnavailable+":")) {
					t.Fatalf("MCP early authority admission missing: %v", err)
				}
				if !node {
					switch entry {
					case "resolve":
						if err != nil {
							t.Fatalf("per-home empty acquisition: %v", err)
						}
					case "publish":
						if err == nil || !strings.Contains(err.Error(), "mcp_declaration_invalid: MCP lock header") {
							t.Fatalf("per-home lock-header control: %v", err)
						}
					default:
						if err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable: protected MCP lock is unavailable") {
							t.Fatalf("per-home missing-lock control: %v", err)
						}
					}
				}
				if nativeSnapshot(t, home) != before {
					t.Fatal("invalid-input authority check wrote state")
				}
			})
		}
	}
}
