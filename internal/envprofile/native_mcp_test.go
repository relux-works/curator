package envprofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/relux-works/curator/internal/envorigin"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/mcpnative"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

func nativeWriterFixture(t *testing.T) (*managedFixture, ProfileMCPRequest) {
	t.Helper()
	fx := writeManagedFixture(t, "sample")
	// Every actual writer consumer has inert PATH tripwires for server, auth
	// and network entrypoints. No real native home or credential store is used.
	bin := t.TempDir()
	tripwire := filepath.Join(bin, "executed")
	if runtime.GOOS != "windows" {
		for _, name := range []string{"fixture-mcp", "claude", "codex", "security", "curl"} {
			script := []byte("#!/bin/sh\nprintf called > '" + tripwire + "'\n")
			if e := os.WriteFile(filepath.Join(bin, name), script, 0o700); e != nil {
				t.Fatal(e)
			}
		}
		t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
		t.Cleanup(func() {
			if _, e := os.Lstat(tripwire); !os.IsNotExist(e) {
				t.Errorf("FORBIDDEN_RUNTIME_OR_AUTH_CALL: %v", e)
			}
		})
	}

	_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	source := "github.com/example/native-mcp"
	pin := writeGitStoreFixture(t, fx.home, "mcp", "figma-devmode", source, map[string]string{"agent-mcp.json": `{"schema_version":1,"name":"figma-devmode","version":"1.2.0","server":{"transport":"stdio","command":"fixture-mcp","args":["--stdio"],"env_names":[],"environments":["claude_code","codex_cli","muse","opencode","pi"]}}`})
	for i := range lock.Members {
		m := &lock.Members[i]
		if m.Kind == contextlock.KindMCP {
			m.Source = source
			m.Commit = pin
		}
	}
	fx.lockHash, err = contextlock.Write(lockPath(fx.home, fx.profile), lock)
	if err != nil {
		t.Fatal(err)
	}
	// The first MCP installation is distinct from first home provisioning.
	// Establish the real recorded homes through Resolve, with only fixture account
	// evidence. ProfileMCP must preserve this authority rather than invent it.
	for _, id := range []string{"claude_code", "codex_cli"} {
		resolve := fx.request(id)
		resolve.Repair = true
		if _, err := Resolve(resolve); err != nil {
			t.Fatal(err)
		}
		// The operator may remove an owned generated MCP alias before enrolling the
		// separate native writer. Fixture setup never follows or copies that alias.
		if err := os.Remove(nativeWriterPath(fx, id)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		adapter, _ := envregistry.ByID(id)
		leaf := filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), adapter.SkillsDir, "myskill")
		if info, err := os.Lstat(leaf); err == nil && info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(leaf); err != nil {
				t.Fatal(err)
			}
		}
	}
	return fx, ProfileMCPRequest{Home: fx.home, Profile: fx.profile, Scope: "profile", Mode: "install", Environments: []string{"claude_code", "codex_cli"}, qualified: func(string) bool { return true }}
}
func nativeWriterPath(fx *managedFixture, id string) string {
	return filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), filepath.FromSlash(mcpnative.Path(id)))
}
func nativeWrite(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := privatedir.MakeAll(filepath.Dir(path)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := privatedir.ProtectFile(path); err != nil {
		t.Fatal(err)
	}
}
func nativeRead(t *testing.T, path string) []byte {
	t.Helper()
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func nativeAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected absent payload: %s (%v)", filepath.Base(path), err)
	}
}
func requireMCPError(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), code) {
		t.Fatalf("error = %v, want %s", err, code)
	}
}

// Through the production writer, both adapters install alongside disjoint
// foreign bytes and jointly publish native entries, complete protected records,
// applicable skill context and the exact lock. Status is read-only and stable.
func TestProfileMCPFirstInstall(t *testing.T) {
	fx, req := nativeWriterFixture(t)
	beforeLock := nativeRead(t, lockPath(fx.home, fx.profile))
	foreign := map[string][]byte{"claude_code": []byte("{ \"setting\": true, \"mcpServers\": {\"foreign\" : {\"command\":\"external\",\"args\":[]}} }\n"), "codex_cli": []byte("# keep exact\nsetting = 'value'\n[mcp_servers.foreign]\ncommand = 'external'\nargs = []\n")}
	for id, body := range foreign {
		nativeWrite(t, nativeWriterPath(fx, id), body)
	}
	result, err := ProfileMCP(req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Published || result.Notice != "MCP configuration installed. Authentication not checked." || len(result.Rows) != 2 {
		t.Fatalf("install result: %+v", result)
	}
	snapshots := map[string][]byte{}
	for _, id := range req.Environments {
		path := nativeWriterPath(fx, id)
		body := nativeRead(t, path)
		entries, e := mcpnative.Entries(id, body)
		if e != nil {
			t.Fatal(e)
		}
		old, _ := mcpnative.Entries(id, foreign[id])
		if !bytes.Equal(old["foreign"], entries["foreign"]) {
			t.Fatal("FOREIGN_BYTES_CHANGED")
		}
		if id == "codex_cli" && !bytes.HasPrefix(body, foreign[id]) {
			t.Fatal("foreign TOML prefix changed")
		}
		markerPath := nativeMCPLedger(fx.home, fx.profile, id)
		markerBytes, e := os.ReadFile(markerPath)
		if e != nil {
			t.Fatalf("NATIVE_WITHOUT_OWNERSHIP: %v", e)
		}
		marker, e := mcpnative.DecodeOwnership(markerBytes, fx.profile, id)
		if e != nil || len(marker.Records) != 1 {
			t.Fatalf("invalid ownership: %+v %v", marker, e)
		}
		record := marker.Records[0]
		if record.Owner != "curator" || record.Name != "figma-devmode" || record.PreimageSHA256 != nil || record.InstalledSHA256 != mcpnative.Digest(entries[record.Name]) {
			t.Fatalf("wrong ownership: %+v", record)
		}
		if e = privatedir.ValidateFile(markerPath); e != nil {
			t.Fatal(e)
		}
		adapter, _ := envregistry.ByID(id)
		if string(nativeRead(t, filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), adapter.SkillsDir, "myskill", "SKILL.md"))) != "# myskill\n" {
			t.Fatal("skill payload missing")
		}
		snapshots[path] = body
		snapshots[markerPath] = markerBytes
	}
	if !bytes.Equal(nativeRead(t, lockPath(fx.home, fx.profile)), beforeLock) {
		t.Fatal("exact profile lock changed")
	}
	req.Mode = "status"
	req.qualified = func(string) bool { t.Fatal("status attempted native qualification"); return false }
	recoveries := 0
	req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
	result, err = ProfileMCP(req)
	if err != nil || result.Published || result.Notice != "" || recoveries != 0 {
		t.Fatalf("status: %+v %v recoveries=%d", result, err, recoveries)
	}
	for _, row := range result.Rows {
		if row.State != "unchanged" {
			t.Fatalf("status row: %+v", row)
		}
	}
	for path, old := range snapshots {
		if !bytes.Equal(old, nativeRead(t, path)) {
			t.Fatal("status mutated payload")
		}
	}
}

// Edited and missing owned entries remain untouched. Status distinguishes the
// observations; repair/install cannot restore, replace, remove or enroll them.
func TestProfileMCPDriftAndAuthorityRefusals(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, change := range []string{"edited", "missing", "removed-declaration", "replacement"} {
			t.Run(id+"/"+change, func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				req.Environments = []string{id}
				if _, err := ProfileMCP(req); err != nil {
					t.Fatal(err)
				}
				path := nativeWriterPath(fx, id)
				markerPath := nativeMCPLedger(fx.home, fx.profile, id)
				marker := nativeRead(t, markerPath)
				want := "drift"
				switch change {
				case "edited":
					body := nativeRead(t, path)
					body = bytes.ReplaceAll(body, []byte("fixture-mcp"), []byte("external-mcp"))
					nativeWrite(t, path, body)
				case "missing":
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					want = "missing"
				case "removed-declaration":
					_, lock, _, e := loadResolveInputs(fx.home, fx.profile)
					if e != nil {
						t.Fatal(e)
					}
					members := lock.Members[:0]
					for _, m := range lock.Members {
						if m.Kind != contextlock.KindMCP {
							members = append(members, m)
						}
					}
					lock.Members = members
					if _, e = contextlock.Write(lockPath(fx.home, fx.profile), lock); e != nil {
						t.Fatal(e)
					}
					want = "removed-declaration"
				case "replacement":
					var markerObject map[string]any
					_ = json.Unmarshal(marker, &markerObject)
					markerObject["records"].([]any)[0].(map[string]any)["server_sha256"] = mcpnative.Digest([]byte("old server"))
					changed, e := canonicalMCPTest(markerObject)
					if e != nil {
						t.Fatal(e)
					}
					nativeWrite(t, markerPath, changed)
					marker = changed
					want = "replacement-required"
				}
				before, _ := os.ReadFile(path)
				req.Mode = "status"
				result, e := ProfileMCP(req)
				if e != nil || len(result.Rows) != 1 || result.Rows[0].State != want {
					t.Fatalf("status: %+v %v", result, e)
				}
				for _, mode := range []string{"install", "repair"} {
					req.Mode = mode
					result, e = ProfileMCP(req)
					requireMCPError(t, e, mcpnative.AuthorityUnavailable)
					if result.Published {
						t.Fatal("authority-free repair published")
					}
					after, _ := os.ReadFile(path)
					if !bytes.Equal(before, after) || !bytes.Equal(marker, nativeRead(t, markerPath)) {
						t.Fatal("refused action changed payload")
					}
				}
			})
		}
	}
}
func canonicalMCPTest(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	var o mcpnative.Ownership
	if e = json.Unmarshal(b, &o); e != nil {
		return nil, e
	}
	return o.Bytes()
}

// Foreign same-NAME entries never acquire ownership, including exact desired
// bytes. Codex checks the live base for A, B and edits after B independently of
// seed metadata. Every mixed plan aborts all tentative sibling mutations.
func TestProfileMCPForeignAndBaseConflicts(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, where := range []string{"same", "different", "base-a", "base-b", "post-b"} {
			if id == "claude_code" && strings.HasPrefix(where, "base") || id == "claude_code" && where == "post-b" {
				continue
			}
			t.Run(id+"/"+where, func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				path := nativeWriterPath(fx, id)
				command := "fixture-mcp"
				if where == "different" {
					command = "external"
				}
				body := []byte(`{"mcpServers":{"figma-devmode":{"type":"stdio","command":"` + command + `","args":["--stdio"]}}}`)
				if id == "codex_cli" {
					body = []byte("[mcp_servers.figma-devmode]\ncommand = \"" + command + "\"\nargs = [\"--stdio\"]\n")
				}
				if where != "same" && where != "different" {
					path = filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), "config.toml")
				}
				nativeWrite(t, path, body)
				result, err := ProfileMCP(req)
				if !bytes.Equal(body, nativeRead(t, path)) {
					t.Fatal("FOREIGN_BYTES_CHANGED")
				}
				requireMCPError(t, err, mcpnative.Conflict)
				if result.Published || len(result.Rows) != 2 {
					t.Fatalf("mixed conflict result: %+v", result)
				}
				if !bytes.Equal(body, nativeRead(t, path)) {
					t.Fatal("FOREIGN_BYTES_CHANGED")
				}
				for _, row := range result.Rows {
					if row.State == "added" {
						t.Fatal("successful row survived failed plan")
					}
					if row.State == "conflict" && row.Code != "" {
						t.Fatal("conflict row must have only summary code")
					}
				}
				for _, env := range req.Environments {
					nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, env))
					if env != id {
						nativeAbsent(t, nativeWriterPath(fx, env))
					}
				}
			})
		}
	}
}

// Scope/action flags never confer authority, while dry-run and status remain
// useful without mutations or recovery. Unqualified required adapters refuse;
// known unsupported adapters retain null path/code.
func TestProfileMCPScopeActionsAndPreview(t *testing.T) {
	fx, req := nativeWriterFixture(t)
	req.Scope = "project"
	_, err := ProfileMCP(req)
	requireMCPError(t, err, "project_scope_shared_home")
	req.Scope = "profile"
	for _, action := range []string{"restore", "replace", "remove", "release", "takeover"} {
		req.Operation = action
		_, err = ProfileMCP(req)
		requireMCPError(t, err, mcpnative.AuthorityUnavailable)
	}
	req.Operation = ""
	req.DryRun = true
	result, err := ProfileMCP(req)
	if err != nil || len(result.Rows) != 2 || result.Published {
		t.Fatalf("preview: %+v %v", result, err)
	}
	for _, row := range result.Rows {
		if row.State != "would:added" {
			t.Fatalf("preview row: %+v", row)
		}
	}
	for _, id := range req.Environments {
		nativeAbsent(t, nativeWriterPath(fx, id))
		nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
	}
	req.DryRun = false
	req.Mode = "status"
	result, err = ProfileMCP(req)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range result.Rows {
		if row.State != "missing" {
			t.Fatal(row)
		}
	}
	req.Mode = "install"
	req.Environments = []string{"claude_code", "codex_cli", "muse", "pi", "opencode"}
	result, err = ProfileMCP(req)
	requireMCPError(t, err, "native_writer_unqualified")
	if len(result.Rows) != 5 {
		t.Fatalf("incomplete pair report: %+v", result)
	}
	for _, row := range result.Rows {
		if row.Environment == "pi" || row.Environment == "opencode" {
			if row.Path != nil || row.Code != "" || row.State != "unsupported" {
				t.Fatalf("unsupported: %+v", row)
			}
		}
	}
	req.Environments = []string{"claude_code"}
	req.qualified = func(string) bool { return false }
	_, err = ProfileMCP(req)
	requireMCPError(t, err, "native_writer_unqualified")
}

// Native/ownership unreadability and no-follow violations refuse the entire
// batch. A successful first install above controls against reject-everything.
func TestProfileMCPUnreadableAndNoFollow(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, kind := range []string{"bad-native", "empty-json", "bad-marker", "missing-owner", "wrong-profile", "unprotected-marker", "link", "parent-link"} {
			if id == "codex_cli" && kind == "empty-json" {
				continue
			}
			t.Run(id+"/"+kind, func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				req.Environments = []string{id}
				path := nativeWriterPath(fx, id)
				switch kind {
				case "bad-native":
					nativeWrite(t, path, []byte("invalid syntax ["))
				case "empty-json":
					nativeWrite(t, path, []byte{})
				case "bad-marker":
					nativeWrite(t, nativeMCPLedger(fx.home, fx.profile, id), []byte("{}"))
				case "wrong-profile", "unprotected-marker":
					profile := fx.profile
					if kind == "wrong-profile" {
						profile = "another"
					}
					body, e := mcpnative.NewOwnership(profile, id).Bytes()
					if e != nil {
						t.Fatal(e)
					}
					p := nativeMCPLedger(fx.home, fx.profile, id)
					nativeWrite(t, p, body)
					if kind == "unprotected-marker" {
						// Read-only foreign access passes the mutation route check
						// but must still fail the ownership ledger's stricter
						// privacy boundary: chmod on Unix, a real ACE on Windows.
						restore, err := makeWorldReadableFileForTest(p)
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() {
							if err := restore(); err != nil {
								t.Error(err)
							}
						})
					}
				case "missing-owner":
					if _, err := ProfileMCP(req); err != nil {
						t.Fatal(err)
					}
					p := nativeMCPLedger(fx.home, fx.profile, id)
					body := nativeRead(t, p)
					body = bytes.Replace(body, []byte(`,"owner":"curator"`), nil, 1)
					nativeWrite(t, p, body)
				case "link":
					target := filepath.Join(t.TempDir(), "external")
					nativeWrite(t, target, []byte("{}"))
					if err := privatedir.MakeAll(filepath.Dir(path)); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(target, path); err != nil {
						t.Fatal(err)
					}
				case "parent-link":
					if err := privatedir.MakeAll(filepath.Dir(ManagedHomeDir(fx.home, fx.profile, id))); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(gitFixtureHome(t), ManagedHomeDir(fx.home, fx.profile, id)); err != nil {
						t.Fatal(err)
					}
				}
				before, _ := os.ReadFile(path)
				want := mcpnative.Unreadable
				if kind == "parent-link" {
					// The manager-wide authority gate rejects a linked managed-home
					// parent before native planning; both modes preserve its bytes.
					want = envregistry.DiagStoreUntrusted
				}
				readonly := req
				readonly.Mode = "status"
				_, readErr := ProfileMCP(readonly)
				requireMCPError(t, readErr, want)
				result, err := ProfileMCP(req)
				requireMCPError(t, err, want)
				if result.Published {
					t.Fatal("unreadable plan published")
				}
				after, _ := os.ReadFile(path)
				if !bytes.Equal(before, after) {
					t.Fatal("unreadable refusal changed native bytes")
				}
			})
		}
	}
}

// Full native preimages and source/lock inputs are frozen before publication;
// an external edit immediately before staging is preserved, never overwritten.
func TestProfileMCPChangedInputsRefuse(t *testing.T) {
	for _, kind := range []string{"native", "lock", "source", "parent", "parent-same-bytes"} {
		t.Run(kind, func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			req.Environments = []string{"claude_code"}
			path := nativeWriterPath(fx, "claude_code")
			nativeWrite(t, path, []byte("{}\n"))
			req.beforePublish = func() {
				switch kind {
				case "native":
					nativeWrite(t, path, []byte("{\"external\":true}\n"))
				case "lock":
					f := lockPath(fx.home, fx.profile)
					nativeWrite(t, f, append(nativeRead(t, f), ' '))
				case "source":
					_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
					if err != nil {
						t.Fatal(err)
					}
					m := lock.Members[0]
					entry := newGitManager(fx.home).entryPath(fx.home, resolvedOf(m))
					file := filepath.Join(entry, "context", "a.md")
					if err := os.Chmod(filepath.Dir(file), 0o700); err != nil {
						t.Fatal(err)
					}
					nativeWrite(t, file, []byte("changed\n"))
				case "parent-same-bytes":
					parent := filepath.Dir(path)
					if err := os.Rename(parent, parent+"-old"); err != nil {
						t.Fatal(err)
					}
					nativeWrite(t, path, []byte("{}\n"))
				case "parent":
					parent := filepath.Dir(path)
					if err := os.Rename(parent, parent+"-old"); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(parent+"-old", parent); err != nil {
						t.Fatal(err)
					}
				}
			}
			result, err := ProfileMCP(req)
			if kind == "native" && !bytes.Contains(nativeRead(t, path), []byte("external")) {
				t.Fatal("EXTERNAL_PREIMAGE_OVERWRITTEN")
			}
			switch kind {
			case "parent", "parent-same-bytes":
				if err == nil {
					t.Fatal("changed parent accepted")
				}
			case "lock":
				requireMCPError(t, err, envorigin.DiagInvalid)
			default:
				requireMCPError(t, err, "transaction_input_changed")
			}
			if result.Published {
				t.Fatal("changed input published")
			}
			nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, "claude_code"))
			if kind == "native" && !bytes.Contains(nativeRead(t, path), []byte("external")) {
				t.Fatal("external edit overwritten")
			}
		})
	}
}

// A failure at any first-install target boundary rolls native entries, ownership,
// skills and lock back together for each adapter; no unmarked native install survives.
func TestProfileMCPJointRollback(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, class := range []string{"native", "ownership", "skill", "lock"} {
			t.Run(id+"/"+class, func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				req.Environments = []string{id}
				beforeLock := nativeRead(t, lockPath(fx.home, fx.profile))
				path := nativeWriterPath(fx, id)
				original := []byte("{}\n")
				if id == "codex_cli" {
					original = []byte("# original\n")
				}
				nativeWrite(t, path, original)
				fired := false
				injected := errors.New("injected payload boundary")
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
					if event.Point != transaction.PointAfterInstall {
						return nil
					}
					matches := class == "native" && event.LivePath == path || class == "ownership" && strings.Contains(event.LivePath, "mcp-native-v1") || class == "skill" && strings.HasSuffix(event.LivePath, "SKILL.md") || class == "lock" && event.LivePath == lockPath(fx.home, fx.profile)
					if matches {
						fired = true
						return injected
					}
					return nil
				}})}
				result, err := ProfileMCP(req)
				if !fired || !errors.Is(err, injected) || result.Published {
					t.Fatalf("boundary was not reached/refused: fired=%v result=%+v err=%v", fired, result, err)
				}
				if !bytes.Equal(original, nativeRead(t, path)) || !bytes.Equal(beforeLock, nativeRead(t, lockPath(fx.home, fx.profile))) {
					t.Fatal("joint rollback changed old generation")
				}
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
				adapter, _ := envregistry.ByID(id)
				nativeAbsent(t, filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), adapter.SkillsDir, "myskill", "SKILL.md"))
			})
		}
	}
}

// Legacy materialization refuses an enrolled surface, even when the current
// MCP declarations are empty. It cannot silently replace per-record ownership.
func TestProfileMCPLegacyWriterRefusesEnrollment(t *testing.T) {
	fx, req := nativeWriterFixture(t)
	req.Environments = []string{"claude_code"}
	if _, err := ProfileMCP(req); err != nil {
		t.Fatal(err)
	}
	before := nativeRead(t, nativeWriterPath(fx, "claude_code"))
	adapter, _ := envregistry.ByID("claude_code")
	plan := &homePlan{adapter: adapter, docs: map[string][]byte{}, links: map[string]string{}, fileHashes: map[string][]byte{}}
	legacy := fx.request("claude_code")
	err := plan.mcp(&legacy, []contextmaterialize.MCPServer{{Name: "figma-devmode", Transport: "stdio", Command: "fixture-mcp"}}, map[string]envmarker.Surface{})
	requireMCPError(t, err, mcpnative.AuthorityUnavailable)
	if !bytes.Equal(before, nativeRead(t, nativeWriterPath(fx, "claude_code"))) {
		t.Fatal("legacy writer changed native bytes")
	}
}

// Qualification reads executable bytes without running them on any platform.
// The mismatched digest refuses, while an explicitly invoked copy of this
// test binary proves the bounded execution tripwire actually detects a call.
func TestProfileMCPExecutableQualificationInert(t *testing.T) {
	root := t.TempDir()
	marker := filepath.Join(root, "executed")
	path := filepath.Join(root, "mcp-qualification-tripwire")
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	payload := nativeRead(t, executable)
	if err := os.WriteFile(path, payload, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_TEST_MCP_QUALIFICATION_TRIPWIRE", marker)
	want := strings.TrimPrefix(mcpnative.Digest(payload), "sha256:")
	if !nativeMCPExecutableMatches(path, want) {
		t.Fatal("matching bytes refused")
	}
	if nativeMCPExecutableMatches(path, strings.Repeat("0", 64)) {
		t.Fatal("different bytes qualified")
	}
	nativeAbsent(t, marker)
	if err := exec.Command(path).Run(); err != nil {
		t.Fatal(err)
	}
	if string(nativeRead(t, marker)) != "called" {
		t.Fatal("execution tripwire plant did not fire")
	}
}

// The fixture home is private on each platform and installs successfully.
// Broadened home or managed-home mutation rights still cause the production
// manager-lock or route check to refuse, without publishing native/ownership/skill payloads
// or repairing the caller's permissions. Windows exercises actual DACLs.
func TestProfileMCPPrivateFixtureBoundary(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, mode := range []string{"private", "home", "managed-home"} {
			t.Run(id+"/"+mode, func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				req.Environments = []string{id}
				if err := privatedir.Validate(fx.home); err != nil {
					t.Fatalf("fixture home is not private: %v", err)
				}
				body := []byte{}
				if id == "claude_code" {
					body = []byte("{}\n")
				}
				nativeWrite(t, nativeWriterPath(fx, id), body)
				// Establish the normal empty operation-lock footprint before
				// planting permissions, so the snapshot covers payload effects.
				op, err := beginOperation(fx.home)
				if err != nil {
					t.Fatal(err)
				}
				if err := op.close(); err != nil {
					t.Fatal(err)
				}
				widened := ""
				if mode != "private" {
					widened = fx.home
					if mode == "managed-home" {
						widened = ManagedHomeDir(fx.home, fx.profile, id)
					}
					restore, err := makeWorldWritableSingleDirectoryForTest(widened)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() {
						if err := restore(); err != nil {
							t.Error(err)
						}
					})
				}
				before := snapshotMCPRecovery(t, fx.home)
				result, err := ProfileMCP(req)
				if mode == "private" {
					if err != nil || !result.Published || len(result.Rows) != 1 || result.Rows[0].State != "added" {
						t.Fatalf("private install: %+v %v", result, err)
					}
					return
				}
				// Main's Windows manager-home lock refuses the foreign grant
				// before the native writer's later route-permission check.
				refusal := "permissions:"
				if runtime.GOOS == "windows" && mode == "home" {
					refusal = "environment_lock_unavailable:"
				}
				if err == nil || !strings.Contains(err.Error(), refusal) || result.Published {
					t.Fatalf("broadened mutation rights accepted: %+v %v", result, err)
				}
				if runtime.GOOS == "windows" && mode == "home" && !strings.Contains(err.Error(), "DACL grants access to another principal") {
					t.Fatalf("manager-home refusal did not identify the foreign grant: %v", err)
				}
				if privatedir.Validate(widened) == nil {
					t.Fatal("refusal repaired broadened permissions")
				}
				if !reflect.DeepEqual(before, snapshotMCPRecovery(t, fx.home)) {
					t.Fatal("refusal changed payload bytes")
				}
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
				adapter, _ := envregistry.ByID(id)
				nativeAbsent(t, filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), adapter.SkillsDir, "myskill"))
			})
		}
	}
}

// An interruption after durable target publication leaves real resumable
// intent. Status/preview preserve it; a fresh engine completes one coherent
// native/ownership/skills/lock generation. Changed frozen policy refuses and
// rolls back, even when its file identity has not changed.
func TestProfileMCPDurableRecovery(t *testing.T) {
	for _, id := range []string{"claude_code", "codex_cli"} {
		for _, changed := range []bool{false, true} {
			t.Run(id+map[bool]string{false: "/valid", true: "/changed-policy"}[changed], func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				req.Environments = []string{id}
				cfgPath := filepath.Join(fx.home, "config.json")
				configBody, _ := json.Marshal(map[string]any{"schema_version": 1, "projects": map[string]any{}, "skills_root": filepath.Join(fx.home, "unused-skills")})
				nativeWrite(t, cfgPath, configBody)
				cfg, e := config.Load(cfgPath, nil)
				if e != nil {
					t.Fatal(e)
				}
				req.Config = cfg
				beforeLock := nativeRead(t, lockPath(fx.home, fx.profile))
				path := nativeWriterPath(fx, id)
				interrupted := false
				req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Observe: func(event transaction.Event) {
					if event.Point == transaction.PointAfterInstall && event.LivePath == path {
						panic("fixture durable interruption")
					}
				}})}
				func() {
					defer func() {
						if p := recover(); p != "fixture durable interruption" {
							t.Fatalf("unexpected interruption: %v", p)
						} else {
							interrupted = true
						}
					}()
					_, _ = ProfileMCP(req)
				}()
				if !interrupted {
					t.Fatal("durable native boundary not reached")
				}
				inventory := filepath.Join(fx.home, "state", "transactions", "v1")
				before := snapshotMCPRecovery(t, fx.home)
				for _, mode := range []string{"status", "install"} {
					readOnly := req
					readOnly.transactionOptions = nil
					readOnly.Mode = mode
					readOnly.DryRun = mode == "install"
					report, err := ProfileMCP(readOnly)
					if err == nil || !strings.Contains(err.Error(), DiagMaterializeRecoveryRequired) || !report.RecoveryRequired || report.Published {
						t.Fatalf("pending read-only report: %+v %v", report, err)
					}
					after := snapshotMCPRecovery(t, fx.home)
					if !reflect.DeepEqual(before, after) {
						t.Fatal("read-only request changed pending generation")
					}
				}
				if changed {
					nativeWrite(t, cfgPath, []byte(`{"schema_version":1,"projects":{},"allowed_sources":["github.com/elsewhere"]}`))
				}
				op, err := beginOperation(fx.home)
				if changed {
					requireMCPError(t, err, "transaction_input_changed")
					nativeAbsent(t, path)
					nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if !op.recovered {
						t.Fatal("fresh recovery was not recorded")
					}
					if err = op.close(); err != nil {
						t.Fatal(err)
					}
					body := nativeRead(t, path)
					entries, e := mcpnative.Entries(id, body)
					if e != nil {
						t.Fatal(e)
					}
					marker, e := mcpnative.DecodeOwnership(nativeRead(t, nativeMCPLedger(fx.home, fx.profile, id)), fx.profile, id)
					if e != nil || len(marker.Records) != 1 || marker.Records[0].InstalledSHA256 != mcpnative.Digest(entries["figma-devmode"]) {
						t.Fatalf("recovery split native/ownership: %+v %v", marker, e)
					}
					adapter, _ := envregistry.ByID(id)
					_ = nativeRead(t, filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), adapter.SkillsDir, "myskill", "SKILL.md"))
				}
				if !bytes.Equal(beforeLock, nativeRead(t, lockPath(fx.home, fx.profile))) {
					t.Fatal("recovery changed exact profile lock")
				}
				rows, e := os.ReadDir(inventory)
				if e != nil || len(rows) != 0 {
					t.Fatalf("recovery inventory not cleaned: %v %v", rows, e)
				}
			})
		}
	}
}
func snapshotMCPRecovery(t *testing.T, home string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(home, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		body, e := os.ReadFile(path)
		if e == nil {
			out[path] = string(body)
		}
		return e
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Existing profile admission gates are replayed against the locked snapshots,
// with no payload effects after policy denial. Fresh config bytes override a
// caller's stale permissive Policy; admitted installs are controlled above.
func TestProfileMCPCurrentAdmission(t *testing.T) {
	for _, gate := range []string{"source-allowlist", "mcp-allowlist", "signers", "revocation", "current-config", "policy-race", "http"} {
		t.Run(gate, func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			req.Environments = []string{"claude_code", "codex_cli"}
			code := ""
			switch gate {
			case "source-allowlist":
				req.Policy.AllowedSources = []string{"github.com/elsewhere"}
				code = "profile_source_invalid"
			case "mcp-allowlist":
				req.Policy.MCPAllowlist = []string{"github.com/elsewhere"}
				code = "mcp_package_not_allowed"
			case "signers":
				req.Policy.RequireSourceSigners = true
				code = "source_signers_missing"
			case "revocation":
				req.Policy.Revocations = []string{"source:github.com/example/native-mcp"}
				code = "revoked"
			case "current-config", "policy-race":
				path := filepath.Join(fx.home, "config.json")
				body, _ := json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(fx.home, "unused-skills"), "projects": map[string]any{}})
				nativeWrite(t, path, body)
				cfg, e := config.Load(path, nil)
				if e != nil {
					t.Fatal(e)
				}
				req.Config = cfg
				if gate == "current-config" {
					body, _ = json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(fx.home, "unused-skills"), "projects": map[string]any{}, "allowed_sources": []string{"github.com/elsewhere"}})
					nativeWrite(t, path, body)
					code = "profile_source_invalid"
				} else {
					req.beforePublish = func() { nativeWrite(t, path, append(body, ' ')) }
					code = "transaction_input_changed"
				}
			case "http":
				_, lock, _, e := loadResolveInputs(fx.home, fx.profile)
				if e != nil {
					t.Fatal(e)
				}
				pin := writeGitStoreFixture(t, fx.home, "mcp", "figma-devmode", "github.com/example/native-http", map[string]string{"agent-mcp.json": `{"schema_version":1,"name":"figma-devmode","version":"1.2.0","server":{"transport":"http","url":"https://service.example/mcp","env_names":[],"environments":["claude_code","codex_cli"]}}`})
				for i := range lock.Members {
					if lock.Members[i].Kind == contextlock.KindMCP {
						lock.Members[i].Commit = pin
						lock.Members[i].Source = "github.com/example/native-http"
					}
				}
				if _, e = contextlock.Write(lockPath(fx.home, fx.profile), lock); e != nil {
					t.Fatal(e)
				}
				code = "native_writer_unqualified"
			}
			result, err := ProfileMCP(req)
			requireMCPError(t, err, code)
			if result.Published || len(result.Rows) != 2 {
				t.Fatalf("denial omitted independently known pairs: %+v", result)
			}
			for _, row := range result.Rows {
				if row.State == "added" {
					t.Fatal("denied batch retained mutation row")
				}
			}
			for _, id := range req.Environments {
				nativeAbsent(t, nativeWriterPath(fx, id))
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
			}
		})
	}
}

// A signed version tag can authorize an unsigned exact commit in the legacy
// profile cache without project-only origin metadata. Moving that tag to a
// different commit cannot authorize the frozen snapshot; both cases replay
// the current source signer list without fetching. Valid signed provenance
// cannot bypass a home whose mutation permissions broaden before publication.
func TestProfileMCPSignedTagReplay(t *testing.T) {
	for _, variant := range []string{"valid", "different-commit", "broadened-home"} {
		t.Run(variant, func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			req.Environments = []string{"codex_cli"}
			source := "github.com/example/native-mcp"
			repo := newGitManager(fx.home).repoDir(source)
			if remotes, err := gitOutput(repo, "remote"); err != nil || strings.TrimSpace(remotes) != "" {
				t.Fatalf("legacy profile cache must have no remote metadata: %v", err)
			}
			key, pub := statusSignerKey(t)
			gitStatusCommand(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "tag", "-s", "-m", "fixture", "v1.2.0")
			req.Policy.SourceSigners = map[string][]contextresolve.Signer{source: {{Type: "ssh", Key: pub}}}
			if variant == "different-commit" {
				gitStatusCommand(t, repo, "commit", "--allow-empty", "-m", "different")
				gitStatusCommand(t, repo, "tag", "-d", "v1.2.0")
				gitStatusCommand(t, repo, "-c", "gpg.format=ssh", "-c", "user.signingkey="+key, "tag", "-s", "-m", "fixture", "v1.2.0")
			}
			if variant == "broadened-home" {
				req.beforePublish = func() {
					restore, e := makeWorldWritableSingleDirectoryForTest(fx.home)
					if e != nil {
						t.Fatal(e)
					}
					t.Cleanup(func() {
						if e := restore(); e != nil {
							t.Error(e)
						}
					})
				}
			}
			result, err := ProfileMCP(req)
			switch variant {
			case "broadened-home":
				requireMCPError(t, err, "permissions:")
				if result.Published {
					t.Fatal("signed replay published through broadened ACL")
				}
				nativeAbsent(t, nativeWriterPath(fx, "codex_cli"))
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, "codex_cli"))
			case "different-commit":
				requireMCPError(t, err, contextresolve.DiagSourceUnsigned)
				if result.Published {
					t.Fatalf("wrong tag commit admitted: %+v %v", result, err)
				}
				nativeAbsent(t, nativeWriterPath(fx, "codex_cli"))
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, "codex_cli"))
			default:
				if err != nil || !result.Published {
					t.Fatalf("valid tag refused: %+v %v", result, err)
				}
			}
		})
	}
}

// Disjoint base servers remain byte-identical and explicitly ungoverned. A
// base edit introducing a collision after a successful install still conflicts.
func TestProfileMCPCodexForeignBasePreserved(t *testing.T) {
	fx, req := nativeWriterFixture(t)
	req.Environments = []string{"codex_cli"}
	basePath := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "codex_cli"), "config.toml")
	base := []byte("# original base\n[mcp_servers.foreign]\ncommand='external'\n")
	nativeWrite(t, basePath, base)
	result, err := ProfileMCP(req)
	if err != nil || !result.Published {
		t.Fatalf("disjoint base: %+v %v", result, err)
	}
	if !bytes.Equal(base, nativeRead(t, basePath)) {
		t.Fatal("foreign base changed")
	}
	if len(result.Warnings) != 2 || result.Warnings[0] != "mcp_native_servers_ungoverned" || result.Warnings[1] != "mcp_seed_unstripped" {
		t.Fatalf("base warnings lost: %+v", result.Warnings)
	}
	nativeWrite(t, basePath, append(base, []byte("[mcp_servers.figma-devmode]\ncommand='external'\n")...))
	req.Mode = "status"
	result, err = ProfileMCP(req)
	requireMCPError(t, err, mcpnative.Conflict)
	if result.Published {
		t.Fatal("base conflict published")
	}
}

// Two NAME additions share one complete ledger per adapter. Adding a second
// name later retains the existing protected sibling; an edited second entry
// then refuses repair without changing either ledger or either native file.
func TestProfileMCPMultipleRecords(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		t.Run(map[bool]string{false: "together", true: "incremental"}[incremental], func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			if incremental {
				if _, err := ProfileMCP(req); err != nil {
					t.Fatal(err)
				}
			}
			_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			source := "github.com/example/native-second"
			pin := writeGitStoreFixture(t, fx.home, "mcp", "second-tool", source, map[string]string{"agent-mcp.json": `{"schema_version":1,"name":"second-tool","version":"1.2.0","server":{"transport":"stdio","command":"fixture-mcp","args":[],"environments":["claude_code","codex_cli"]}}`})
			lock.Members = append(lock.Members, contextlock.Member{Kind: contextlock.KindMCP, Name: "second-tool", Source: source, Version: "1.2.0", Commit: pin, RequiredBy: []string{fx.profile}})
			lock.Sort()
			canonical, e := lock.Canonical()
			if e != nil {
				t.Fatal(e)
			}
			// Preserve captured origin while changing only this fixture's installed
			// declaration. Deliberate wrong-lock refusal is tested separately.
			old, oldHash, e := readLock(fx.home, fx.profile)
			_ = old
			if e != nil {
				t.Fatal(e)
			}
			for _, id := range []string{"claude_code", "codex_cli"} {
				target := ManagedHomeDir(fx.home, fx.profile, id)
				marker, e := envmarker.Read(target)
				if e != nil {
					t.Fatal(e)
				}
				raw := nativeRead(t, filepath.Join(target, envorigin.Name))
				before := envorigin.Binding{Profile: fx.profile, LockSHA256: "sha256:" + strings.TrimPrefix(oldHash, "sha256:"), ManagedTarget: target}
				after := before
				after.LockSHA256 = "sha256:" + strings.TrimPrefix(contextlock.HashBytes(canonical), "sha256:")
				rebound, e := envorigin.Rebind(raw, before, after, id, fx.home)
				if e != nil {
					t.Fatal(e)
				}
				marker.Profile.LockSHA256 = strings.TrimPrefix(after.LockSHA256, "sha256:")
				body, e := marker.Marshal()
				if e != nil {
					t.Fatal(e)
				}
				nativeWrite(t, filepath.Join(target, envmarker.Name), body)
				nativeWrite(t, filepath.Join(target, envorigin.Name), rebound)
			}
			nativeWrite(t, lockPath(fx.home, fx.profile), canonical)
			result, err := ProfileMCP(req)
			if err != nil || !result.Published || len(result.Rows) != 4 {
				t.Fatalf("multiple records: %+v %v", result, err)
			}
			snapshots := map[string][]byte{}
			for _, id := range req.Environments {
				path := nativeWriterPath(fx, id)
				body := nativeRead(t, path)
				entries, e := mcpnative.Entries(id, body)
				if e != nil || len(entries) != 2 {
					t.Fatalf("native siblings missing: %+v %v", entries, e)
				}
				markerPath := nativeMCPLedger(fx.home, fx.profile, id)
				markerBytes := nativeRead(t, markerPath)
				marker, e := mcpnative.DecodeOwnership(markerBytes, fx.profile, id)
				if e != nil || len(marker.Records) != 2 {
					t.Fatalf("OWNERSHIP_SIBLING_LOST: %+v %v", marker, e)
				}
				for _, record := range marker.Records {
					if record.InstalledSHA256 != mcpnative.Digest(entries[record.Name]) {
						t.Fatalf("sibling hash mismatch: %+v", record)
					}
				}
				snapshots[path], snapshots[markerPath] = body, markerBytes
			}
			req.Mode = "status"
			result, err = ProfileMCP(req)
			if err != nil || len(result.Rows) != 4 {
				t.Fatalf("multiple status: %+v %v", result, err)
			}
			for _, row := range result.Rows {
				if row.State != "unchanged" {
					t.Fatalf("multiple status row: %+v", row)
				}
			}
			path := nativeWriterPath(fx, "codex_cli")
			body := bytes.Replace(snapshots[path], []byte("[mcp_servers.second-tool]\ncommand = \"fixture-mcp\""), []byte("[mcp_servers.second-tool]\ncommand = \"external-mcp\""), 1)
			if bytes.Equal(body, snapshots[path]) {
				t.Fatal("second-entry drift plant did not apply")
			}
			nativeWrite(t, path, body)
			snapshots[path] = body
			req.Mode = "repair"
			_, err = ProfileMCP(req)
			requireMCPError(t, err, mcpnative.AuthorityUnavailable)
			for path, want := range snapshots {
				if !bytes.Equal(want, nativeRead(t, path)) {
					t.Fatal("refusal changed a sibling payload")
				}
			}
		})
	}
}

// An explicitly permitted requested environment name allows first install;
// the same declaration outside the ceiling refuses without native/marker
// publication. Native documents do not carry the requested environment name.
func TestProfileMCPEnvironmentNamesCeiling(t *testing.T) {
	for _, allowed := range []bool{true, false} {
		t.Run(map[bool]string{true: "allowed", false: "forbidden"}[allowed], func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			_, lock, _, err := loadResolveInputs(fx.home, fx.profile)
			if err != nil {
				t.Fatal(err)
			}
			source := "github.com/example/native-env"
			pin := writeGitStoreFixture(t, fx.home, "mcp", "figma-devmode", source, map[string]string{"agent-mcp.json": `{"schema_version":1,"name":"figma-devmode","version":"1.2.0","server":{"transport":"stdio","command":"fixture-mcp","args":[],"env_names":["TOOL_TOKEN"],"environments":["claude_code","codex_cli"]}}`})
			for i := range lock.Members {
				if lock.Members[i].Kind == contextlock.KindMCP {
					lock.Members[i].Source, lock.Members[i].Commit = source, pin
				}
			}
			if _, err = contextlock.Write(lockPath(fx.home, fx.profile), lock); err != nil {
				t.Fatal(err)
			}
			req.Policy.PassableEnvNamesSet = true
			req.Policy.PassableEnvNames = []string{"TOOL_TOKEN"}
			if !allowed {
				req.Policy.PassableEnvNames = []string{"OTHER"}
			}
			result, err := ProfileMCP(req)
			if !allowed {
				requireMCPError(t, err, "mcp_contribution_policy_refused: environment-name ceiling")
				if result.Published {
					t.Fatal("ceiling refusal published")
				}
				for _, id := range req.Environments {
					nativeAbsent(t, nativeWriterPath(fx, id))
					nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, id))
				}
				return
			}
			if err != nil || !result.Published {
				t.Fatalf("allowed environment name: %+v %v", result, err)
			}
			for _, id := range req.Environments {
				if bytes.Contains(nativeRead(t, nativeWriterPath(fx, id)), []byte("TOOL_TOKEN")) {
					t.Fatal("native configuration carried an environment name")
				}
			}
		})
	}
}

// Both TOML string kinds and legal closing lengths install with nonempty
// ownership digests and immediately report unchanged without rewriting bytes.
func TestProfileMCPTOMLMultilineClosingRuns(t *testing.T) {
	for _, quote := range []string{"\"", "'"} {
		for count := 3; count <= 5; count++ {
			t.Run(fmt.Sprintf("%s/%d", quote, count), func(t *testing.T) {
				fx, req := nativeWriterFixture(t)
				req.Environments = []string{"codex_cli"}
				input := []byte("[mcp_servers.foreign]\ncommand='external'\nnote=" + strings.Repeat(quote, 3) + "hello" + strings.Repeat(quote, count) + "\n")
				path := nativeWriterPath(fx, "codex_cli")
				nativeWrite(t, path, input)
				result, err := ProfileMCP(req)
				if err != nil || !result.Published {
					t.Fatalf("install: %+v %v", result, err)
				}
				body := nativeRead(t, path)
				if !bytes.HasPrefix(body, input) {
					t.Fatal("foreign bytes changed")
				}
				entries, err := mcpnative.Entries("codex_cli", body)
				if err != nil || len(entries["figma-devmode"]) == 0 {
					t.Fatalf("INSERTED_SPAN_ABSENT: %v", err)
				}
				ledger, err := mcpnative.DecodeOwnership(nativeRead(t, nativeMCPLedger(fx.home, fx.profile, "codex_cli")), fx.profile, "codex_cli")
				if err != nil || len(ledger.Records) != 1 || ledger.Records[0].InstalledSHA256 != mcpnative.Digest(entries["figma-devmode"]) {
					t.Fatalf("ownership: %+v %v", ledger, err)
				}
				before := snapshotMCPRecovery(t, fx.home)
				req.Mode = "status"
				result, err = ProfileMCP(req)
				if err != nil || len(result.Rows) != 1 || result.Rows[0].State != "unchanged" || !reflect.DeepEqual(before, snapshotMCPRecovery(t, fx.home)) {
					t.Fatalf("status: %+v %v", result, err)
				}
			})
		}
	}
}

// An absent ledger is not permission to publish beneath a pre-existing shared
// ownership parent. Refusals preserve its mode and all payload bytes; 0700
// installs and immediately reports unchanged. Unix modes do not qualify ACLs.
func TestProfileMCPLedgerParentBeforeFirstInstall(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission-bit witnesses; Windows DACL controls run separately")
	}
	for _, mode := range []os.FileMode{0o700, 0o755, 0o777} {
		t.Run(fmt.Sprintf("%o", mode), func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			req.Environments = []string{"claude_code"}
			ledger := nativeMCPLedger(fx.home, fx.profile, "claude_code")
			parent := filepath.Dir(ledger)
			if err := privatedir.MakeAll(parent); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(parent, mode); err != nil {
				t.Fatal(err)
			}
			before := snapshotMCPRecovery(t, fx.home)
			result, err := ProfileMCP(req)
			info, statErr := os.Lstat(parent)
			if statErr != nil || info.Mode().Perm() != mode {
				t.Fatalf("parent mode changed: %v", statErr)
			}
			if mode == 0o700 {
				if err != nil || !result.Published {
					t.Fatalf("private install: %+v %v", result, err)
				}
				req.Mode = "status"
				result, err = ProfileMCP(req)
				if err != nil || len(result.Rows) != 1 || result.Rows[0].State != "unchanged" {
					t.Fatalf("private status: %+v %v", result, err)
				}
			} else {
				requireMCPError(t, err, mcpnative.Unreadable)
				if result.Published {
					t.Fatal("UNTRUSTED_PARENT_PUBLISHED")
				}
				for path, body := range before {
					if string(nativeRead(t, path)) != body {
						t.Fatal("refusal changed existing payload")
					}
				}
				adapter, _ := envregistry.ByID("claude_code")
				nativeAbsent(t, filepath.Join(ManagedHomeDir(fx.home, fx.profile, "claude_code"), adapter.SkillsDir, "myskill"))
				nativeAbsent(t, ledger)
				nativeAbsent(t, nativeWriterPath(fx, "claude_code"))
			}
		})
	}
}

// Preview and mutation agree on deterministic skill conflicts (changed bytes,
// extra file and wrong mode). Preview of absent/identical context stays useful
// and neither creates a path nor recovers or acquires the mutation lock.
func TestProfileMCPPreviewSkillPreimages(t *testing.T) {
	for _, variant := range []string{"absent", "identical", "changed", "extra", "wrong-mode"} {
		t.Run(variant, func(t *testing.T) {
			if variant == "wrong-mode" && runtime.GOOS == "windows" {
				t.Skip("POSIX permission-bit witnesses; Windows DACL controls run separately")
			}
			fx, req := nativeWriterFixture(t)
			req.Environments = []string{"claude_code"}
			adapter, _ := envregistry.ByID("claude_code")
			path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, "claude_code"), adapter.SkillsDir, "myskill", "SKILL.md")
			if variant != "absent" {
				nativeWrite(t, path, []byte("# myskill\n"))
			}
			switch variant {
			case "changed":
				nativeWrite(t, path, []byte("changed\n"))
			case "extra":
				nativeWrite(t, filepath.Join(filepath.Dir(path), "extra.txt"), []byte("extra\n"))
			case "wrong-mode":
				if err := os.Chmod(path, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			before := snapshotMCPReviewTree(t, fx.home)
			recoveries := 0
			req.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{ObserveRecovery: func() { recoveries++ }})}
			req.DryRun = true
			result, err := ProfileMCP(req)
			conflict := variant != "absent" && variant != "identical"
			if conflict {
				requireMCPError(t, err, mcpnative.AuthorityUnavailable)
				if len(result.Rows) != 1 || result.Rows[0].State != "would:refused" || result.Rows[0].Code != mcpnative.AuthorityUnavailable {
					t.Fatalf("preview conflict row: %+v", result)
				}
			} else if err != nil || len(result.Rows) != 1 || result.Rows[0].State != "would:added" {
				t.Fatalf("preview: %+v %v", result, err)
			}
			if result.Published || recoveries != 0 || !reflect.DeepEqual(before, snapshotMCPReviewTree(t, fx.home)) {
				t.Fatal("PREVIEW_MUTATED_STATE")
			}
			req.DryRun = false
			result, err = ProfileMCP(req)
			if conflict {
				requireMCPError(t, err, mcpnative.AuthorityUnavailable)
				nativeAbsent(t, nativeWriterPath(fx, "claude_code"))
				nativeAbsent(t, nativeMCPLedger(fx.home, fx.profile, "claude_code"))
				if string(nativeRead(t, path)) != before[path].body {
					t.Fatal("conflicting skill changed")
				}
			} else if err != nil || !result.Published {
				t.Fatalf("install: %+v %v", result, err)
			}
		})
	}
}

type mcpReviewNode struct {
	mode os.FileMode
	body string
}

func snapshotMCPReviewTree(t *testing.T, root string) map[string]mcpReviewNode {
	t.Helper()
	nodes := map[string]mcpReviewNode{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		node := mcpReviewNode{mode: info.Mode()}
		if info.Mode().IsRegular() {
			body, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			node.body = string(body)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			body, e := os.Readlink(path)
			if e != nil {
				return e
			}
			node.body = body
		}
		nodes[path] = node
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return nodes
}

// Legacy planning accepts shared operator-home permissions only with absent
// enrollment; present/malformed markers and linked ownership routes still
// refuse without touching the plan or foreign target. Protected roots remain
// covered through production Resolve by the checked-root conformance control.
func TestProfileMCPLegacyEnrollmentRoute(t *testing.T) {
	for _, variant := range []string{"absent", "empty-parent", "marker", "malformed", "root-link", "parent-link", "leaf-link"} {
		t.Run(variant, func(t *testing.T) {
			fx, _ := nativeWriterFixture(t)
			ledger := nativeMCPLedger(fx.home, fx.profile, "codex_cli")
			switch variant {
			case "empty-parent":
				if err := privatedir.MakeAll(filepath.Dir(ledger)); err != nil {
					t.Fatal(err)
				}
			case "marker":
				body, _ := mcpnative.NewOwnership(fx.profile, "codex_cli").Bytes()
				nativeWrite(t, ledger, body)
			case "malformed":
				nativeWrite(t, ledger, []byte("invalid\n"))
			case "root-link", "parent-link", "leaf-link":
				target := t.TempDir()
				link := filepath.Join(fx.home, "mcp-native-v1")
				if variant == "parent-link" {
					link = filepath.Dir(ledger)
				}
				if variant == "leaf-link" {
					link = ledger
					target = filepath.Join(target, "marker")
					nativeWrite(t, target, []byte("external\n"))
				}
				if err := privatedir.MakeAll(filepath.Dir(link)); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, link); err != nil {
					t.Skipf("fixture symlink unavailable: %v", err)
				}
			}
			if runtime.GOOS != "windows" {
				restore, err := makeWorldWritableSingleDirectoryForTest(fx.home)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := restore(); err != nil {
						t.Error(err)
					}
				})
			}
			before := snapshotMCPReviewTree(t, fx.home)
			adapter, _ := envregistry.ByID("codex_cli")
			plan := &homePlan{adapter: adapter, docs: map[string][]byte{}, links: map[string]string{}, fileHashes: map[string][]byte{}}
			legacy := fx.request("codex_cli")
			err := plan.mcp(&legacy, nil, map[string]envmarker.Surface{})
			if variant == "absent" || variant == "empty-parent" {
				if err != nil {
					t.Fatalf("LEGACY_OPERATOR_HOME_REFUSED: %v", err)
				}
			} else {
				code := mcpnative.Unreadable
				if variant == "marker" || variant == "malformed" {
					code = mcpnative.AuthorityUnavailable
				}
				requireMCPError(t, err, code)
				if len(plan.docs) != 0 || len(plan.links) != 0 || len(plan.fileHashes) != 0 {
					t.Fatal("legacy refusal altered plan")
				}
			}
			if !reflect.DeepEqual(before, snapshotMCPReviewTree(t, fx.home)) {
				t.Fatal("legacy guard mutated state")
			}
		})
	}
}

// Tight planning-time permissions cannot authorize publication after the same
// directory's mode changes. The unchanged mode is a successful control above.
func TestProfileMCPLedgerParentPermissionRace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission-bit witnesses; Windows DACL controls run separately")
	}
	fx, req := nativeWriterFixture(t)
	req.Environments = []string{"claude_code"}
	ledger := nativeMCPLedger(fx.home, fx.profile, "claude_code")
	parent := filepath.Dir(ledger)
	if err := privatedir.MakeAll(parent); err != nil {
		t.Fatal(err)
	}
	req.beforePublish = func() {
		if err := os.Chmod(parent, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	result, err := ProfileMCP(req)
	requireMCPError(t, err, mcpnative.Unreadable)
	if result.Published {
		t.Fatal("UNTRUSTED_PARENT_PUBLISHED")
	}
	nativeAbsent(t, ledger)
	nativeAbsent(t, nativeWriterPath(fx, "claude_code"))
	info, err := os.Lstat(parent)
	if err != nil || info.Mode().Perm() != 0o755 {
		t.Fatalf("refusal repaired external mode: %v", err)
	}
}

// Invalid private authority refuses at the real owner before parent allocation
// or protection, including an inactive adapter. This pairs the actual first-
// install successes above, rather than treating an early generic error as proof.
func TestProfileMCPOriginBeforeParents(t *testing.T) {
	for _, state := range []string{"missing", "malformed", "inactive"} {
		t.Run(state, func(t *testing.T) {
			fx, req := nativeWriterFixture(t)
			id := "codex_cli"
			if state == "inactive" {
				req.Environments = []string{"claude_code"}
			}
			path := filepath.Join(ManagedHomeDir(fx.home, fx.profile, id), envorigin.Name)
			if state == "missing" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else {
				nativeWrite(t, path, []byte("{}"))
			}
			before := nativeSnapshot(t, fx.home)
			calls := 0
			req.qualified = func(string) bool { calls++; return true }
			req.beforePublish = func() { calls++ }
			result, err := ProfileMCP(req)
			want := envorigin.DiagInvalid
			if state == "missing" {
				want = envorigin.DiagRequired
			}
			requireMCPError(t, err, want)
			if calls != 0 || result.Published || nativeSnapshot(t, fx.home) != before {
				t.Fatal("invalid origin reached MCP effects/qualification")
			}
		})
	}
}
