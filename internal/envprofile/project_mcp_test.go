package envprofile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/projectmcpstate"
	"github.com/relux-works/curator/internal/sourcelock"
)

func mcpDeclarations(t *testing.T) map[string]projectmcp.Declaration {
	t.Helper()
	d, err := projectmcp.ParseDeclaration(map[string]any{"package": map[string]any{"git": "https://packages.example/tool", "tag": "v1.0.0"}})
	if err != nil {
		t.Fatal(err)
	}
	return map[string]projectmcp.Declaration{"tool": d}
}

// Both source policies are segment-aware and checked before any acquisition.
// A nearby admitted source and permissive-empty control prevent reject-all
// implementations from satisfying either allowlist or posture rejection.
func TestProjectMCPPreflight(t *testing.T) {
	decls := mcpDeclarations(t)
	for _, tc := range []struct {
		name   string
		policy Policy
		code   string
	}{
		{"allowed", Policy{AllowedSources: []string{"packages.example"}, MCPAllowlist: []string{"packages.example/tool"}}, ""},
		{"permissive empty", Policy{}, ""},
		{"general source", Policy{AllowedSources: []string{"packages.example/toolkit"}}, "profile_source_invalid"},
		{"MCP segment", Policy{MCPAllowlist: []string{"packages.example/tools"}}, "mcp_package_not_allowed"},
		{"hardened empty", Policy{SecurityPosture: config.SecurityPostureHardened, AllowedSources: []string{"packages.example"}}, "mcp_package_allowlist_empty"},
		{"hardened admitted", Policy{SecurityPosture: config.SecurityPostureHardened, AllowedSources: []string{"packages.example"}, MCPAllowlist: []string{"packages.example"}}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := PreflightProjectMCP(decls, tc.policy)
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.code) {
				t.Fatalf("want %s: %v", tc.code, err)
			}
		})
	}
}

// Each candidate passes through actual Git acquisition, the shared selector and
// signer resolver, stage-A package validation and strict audit. Rejections leave
// the prior consumer records and native configuration byte-identical.
func TestProjectMCPCandidateGates(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(map[string]any, *Policy, map[string]string)
		code   string
	}{
		{"valid", nil, ""},
		{"unpinned", func(m map[string]any, _ *Policy, _ map[string]string) {
			m["server"].(map[string]any)["args"] = []string{"tool"}
		}, "mcp_declaration_invalid"},
		{"secret", func(m map[string]any, _ *Policy, _ map[string]string) {
			m["server"].(map[string]any)["args"] = []string{"tool@1.0.0", "AKIA1234567890ABCDEF"}
		}, "context-secret-material"},
		{"unknown environment", func(m map[string]any, _ *Policy, _ map[string]string) {
			m["server"].(map[string]any)["environments"] = []string{"codex"}
		}, "mcp_declaration_invalid"},
		{"wrong name", func(m map[string]any, _ *Policy, _ map[string]string) { m["name"] = "other" }, "mcp_declaration_invalid"},
		{"version mismatch", func(m map[string]any, _ *Policy, _ map[string]string) { m["version"] = "2.0.0" }, "context_version_mismatch"},
		{"missing signer", func(_ map[string]any, p *Policy, _ map[string]string) { p.RequireSourceSigners = true }, "context_source_signers_missing"},
		{"unsigned", func(_ map[string]any, p *Policy, _ map[string]string) {
			p.SourceSigners = map[string][]contextresolve.Signer{"packages.example/tool": {}}
		}, "context_source_unsigned"},
		{"revoked", func(_ map[string]any, p *Policy, _ map[string]string) {
			p.Revocations = []string{"source:packages.example/tool"}
		}, "revoked"},
		{"ceiling", func(m map[string]any, _ *Policy, _ map[string]string) {
			m["server"].(map[string]any)["env_names"] = []string{"SERVICE_TOKEN"}
		}, "mcp_contribution_policy_refused"},
		{"ceiling admitted", func(m map[string]any, p *Policy, _ map[string]string) {
			m["server"].(map[string]any)["env_names"] = []string{"SERVICE_TOKEN"}
			p.PassableEnvNames = []string{"SERVICE_TOKEN"}
			p.PassableEnvNamesSet = true
		}, ""},
		{"strict audit", func(_ map[string]any, _ *Policy, f map[string]string) {
			f["README.md"] = "POST https://api.example.net/ingest\n"
		}, "audit blocked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ids := newGitIdentities(t)
			home := gitFixtureHome(t)
			native := filepath.Join(home, "native.json")
			if err := os.WriteFile(native, []byte("foreign"), 0600); err != nil {
				t.Fatal(err)
			}
			manifest := map[string]any{"schema_version": 1, "name": "tool", "version": "1.0.0", "server": map[string]any{"transport": "stdio", "command": "npx", "args": []string{"tool@1.0.0"}, "environments": []string{"codex_cli"}}}
			policy := Policy{AllowedSources: []string{"packages.example"}, MCPAllowlist: []string{"packages.example/tool"}}
			files := map[string]string{}
			if tc.change != nil {
				tc.change(manifest, &policy, files)
			}
			payload, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			files["agent-mcp.json"] = string(payload)
			ids.serve(gitRepo(t, files, "v1.0.0"), "https://packages.example/tool")
			entries, _, err := ResolveProjectMCP(home, mcpDeclarations(t), policy)
			if tc.code == "" {
				if err != nil || len(entries) != 1 {
					t.Fatalf("entries %d: %v", len(entries), err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.code) {
				t.Fatalf("want %s, got %v", tc.code, err)
			}
			if err != nil && strings.Contains(err.Error(), "AKIA1234567890ABCDEF") {
				t.Fatal("secret in diagnostic")
			}
			if got, err := os.ReadFile(native); err != nil || string(got) != "foreign" {
				t.Fatalf("native changed: %q %v", got, err)
			}
			if _, err := os.Stat(filepath.Join(home, "mcp-project-locks")); !os.IsNotExist(err) {
				t.Fatal("candidate published authority")
			}
		})
	}
}

func lockedProjectMCP(t *testing.T, policy Policy, signedTag bool) (string, string, *projectmcpstate.Lock, map[string]projectmcp.Declaration, Policy) {
	t.Helper()
	ids := newGitIdentities(t)
	home, root := gitFixtureHome(t), gitFixtureHome(t)
	for _, dir := range []string{home, root} {
		if err := privatedir.Protect(dir); err != nil {
			t.Fatal(err)
		}
	}
	repo := gitRepo(t, map[string]string{"agent-mcp.json": `{"schema_version":1,"name":"tool","version":"1.0.0","server":{"transport":"stdio","command":"npx","args":["tool@1.0.0"],"environments":["codex_cli","muse"]}}`}, "v1.0.0")
	if signedTag {
		key, pub := statusSignerKey(t)
		gitStatusCommand(t, repo, "tag", "-d", "v1.0.0")
		gitStatusCommand(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "tag", "-s", "-m", "fixture", "v1.0.0")
		policy.SourceSigners = map[string][]contextresolve.Signer{"packages.example/tool": {{Type: "ssh", Key: pub}}}
	}
	ids.serve(repo, "https://packages.example/tool")
	decls := mcpDeclarations(t)
	entries, _, err := ResolveProjectMCP(home, decls, policy)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"schema_version":3,"skills":[]}`)
	if err := os.WriteFile(filepath.Join(root, "Skillfile.json"), payload, 0600); err != nil {
		t.Fatal(err)
	}
	hash, err := sourcelock.ManifestDigest(payload)
	if err != nil {
		t.Fatal(err)
	}
	skills, err := sourcelock.New(hash, []sourcelock.Member{})
	if err != nil {
		t.Fatal(err)
	}
	lock := &projectmcpstate.Lock{SchemaVersion: 1, Capability: projectmcp.Capability, ProjectID: "project", ManifestSHA256: hash, SkillsLock: skills.Object(), Entries: entries}
	if err := lock.Seal(); err != nil {
		t.Fatal(err)
	}
	skillBytes, err := projectmcpstate.Canonical(skills.Object())
	if err != nil {
		t.Fatal(err)
	}
	if err := PublishProjectMCPLock(home, root, lock, map[string][]byte{sourcelock.PathIn(root): skillBytes}); err != nil {
		t.Fatal(err)
	}
	return home, root, lock, decls, policy
}

// Replay rechecks protected local authority, exact project/document/source lock,
// full committed snapshot bytes/modes and current source/revocation policy. Each case
// has a usable locked positive fixture and leaves its post-plant state untouched.
func TestProjectMCPAuthorizedReplayGates(t *testing.T) {
	for _, mode := range []string{"valid", "project", "manifest", "portable", "skills lock", "snapshot", "snapshot mode", "missing snapshot", "authority", "policy", "revoked"} {
		t.Run(mode, func(t *testing.T) {
			if mode == "snapshot mode" && runtime.GOOS == "windows" {
				t.Skip("Windows does not expose portable executable permission bits")
			}
			policy := Policy{}
			home, root, lock, decls, _ := lockedProjectMCP(t, policy, false)
			project, hash := "project", lock.ManifestSHA256
			switch mode {
			case "project":
				project = "other"
			case "manifest":
				hash = "sha256:" + strings.Repeat("2", 64)
			case "portable":
				if err := os.WriteFile(ProjectMCPLockPath(root), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "skills lock":
				if err := os.WriteFile(sourcelock.PathIn(root), []byte(`{}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "snapshot mode":
				entry := contextstore.EntryDir(home, "mcp", "tool", lock.Entries[0].Resolved.Commit)
				if err := os.Chmod(filepath.Join(entry, "agent-mcp.json"), 0555); err != nil {
					t.Fatal(err)
				}
			case "snapshot":
				entry := contextstore.EntryDir(home, "mcp", "tool", lock.Entries[0].Resolved.Commit)
				if err := os.Chmod(entry, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(entry, "extra.txt"), []byte("uncommitted"), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing snapshot":
				if err := os.RemoveAll(contextstore.EntryDir(home, "mcp", "tool", lock.Entries[0].Resolved.Commit)); err != nil {
					t.Fatal(err)
				}
			case "authority":
				if err := os.Remove(privateProjectMCPLock(home, root)); err != nil {
					t.Fatal(err)
				}
			case "policy":
				policy.MCPAllowlist = []string{"packages.example/other"}
			case "revoked":
				policy.Revocations = []string{"source:packages.example/tool"}
			}
			before := preserveSnapshot(t, home)
			rootBefore := preserveSnapshot(t, root)
			_, _, err := ReplayProjectMCP(home, root, project, hash, decls, policy)
			code := map[string]string{"project": "source_lock_stale", "manifest": "source_lock_stale", "portable": "source_lock_stale", "skills lock": "source_lock_stale", "snapshot": "source_snapshot_changed", "snapshot mode": "source_snapshot_changed", "missing snapshot": "source_snapshot_unavailable", "authority": "source_snapshot_unavailable", "policy": "mcp_package_not_allowed", "revoked": "revoked"}[mode]
			if mode == "valid" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), code) {
				t.Fatalf("want %s: %v", code, err)
			}
			if !reflect.DeepEqual(before, preserveSnapshot(t, home)) || !reflect.DeepEqual(rootBefore, preserveSnapshot(t, root)) {
				t.Fatal("read-only replay changed state")
			}
		})
	}
}

// A tag signature on an unsigned commit remains admissible during exact local
// replay; a changed signer policy rejects it without fetching or reselection.
func TestProjectMCPTagSignatureReplay(t *testing.T) {
	home, root, lock, decls, policy := lockedProjectMCP(t, Policy{}, true)
	source := lock.Entries[0].Resolved.Source
	if _, _, err := ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, policy); err != nil {
		t.Fatal(err)
	}
	_, other := statusSignerKey(t)
	policy.SourceSigners[source] = []contextresolve.Signer{{Type: "ssh", Key: other}}
	if _, _, err := ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, policy); err == nil || !strings.Contains(err.Error(), "context_source_signer_rejected") {
		t.Fatal(err)
	}
}

// Publication validates the current complete manifest and no-follow target
// routes before replacing any portable/protected consumer record. A normal
// republish is the positive control for the same transaction publisher.
func TestProjectMCPPublicationGates(t *testing.T) {
	for _, mode := range []string{"valid", "stale manifest", "portable symlink", "authority symlink"} {
		t.Run(mode, func(t *testing.T) {
			home, root, lock, _, _ := lockedProjectMCP(t, Policy{}, false)
			switch mode {
			case "stale manifest":
				if err := os.WriteFile(filepath.Join(root, "Skillfile.json"), []byte(`{"schema_version":3,"locale":"en","skills":[]}`), 0600); err != nil {
					t.Fatal(err)
				}
			case "portable symlink", "authority symlink":
				target := ProjectMCPLockPath(root)
				if mode == "authority symlink" {
					target = privateProjectMCPLock(home, root)
				}
				bytes, err := os.ReadFile(target)
				if err != nil {
					t.Fatal(err)
				}
				other := filepath.Join(gitFixtureHome(t), "foreign.json")
				if err := os.WriteFile(other, bytes, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(other, target); err != nil {
					t.Skipf("symlink fixture unavailable: %v", err)
				}
			}
			before := preserveSnapshot(t, home)
			rootBefore := preserveSnapshot(t, root)
			err := PublishProjectMCPLock(home, root, lock, nil)
			if mode == "valid" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				want := "symbolic-link"
				if mode == "stale manifest" {
					want = "source_lock_stale"
				}
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Fatalf("want %s: %v", want, err)
				}
				if !reflect.DeepEqual(before, preserveSnapshot(t, home)) || !reflect.DeepEqual(rootBefore, preserveSnapshot(t, root)) {
					t.Fatal("refused publication changed state")
				}
			}
		})
	}
}

// Preview refuses a missing exact snapshot without writes; an actual install
// reconstructs that same commit from its cached Git input, without changing the
// lock or consulting a moving selector. Subsequent preview uses restored bytes.
func TestProjectMCPMissingSnapshotRestore(t *testing.T) {
	home, root, lock, decls, _ := lockedProjectMCP(t, Policy{}, false)
	entry := contextstore.EntryDir(home, "mcp", "tool", lock.Entries[0].Resolved.Commit)
	if err := os.RemoveAll(entry); err != nil {
		t.Fatal(err)
	}
	before := preserveSnapshot(t, home)
	if _, _, err := ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, Policy{}); err == nil || !strings.Contains(err.Error(), "source_snapshot_unavailable") {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, preserveSnapshot(t, home)) {
		t.Fatal("preview wrote missing snapshot")
	}
	if _, _, err := RestoreProjectMCP(home, root, "project", lock.ManifestSHA256, decls, Policy{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReplayProjectMCP(home, root, "project", lock.ManifestSHA256, decls, Policy{}); err != nil {
		t.Fatal(err)
	}
}
