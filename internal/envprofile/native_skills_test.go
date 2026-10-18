package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

func nativeSkillsFixture(t *testing.T) (*managedFixture, map[string]string) {
	t.Helper()
	pinOperatorHome(t, t.TempDir())
	fx := writeManagedFixture(t, "native-skills")
	homes := pinHomes(t)
	for _, dir := range homes {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	homes["opencode-native"] = homes["opencode"]
	fx.native = homes
	if err := EnsureDefaultWithPolicy(fx.home, Policy{}); err != nil {
		t.Fatal(err)
	}
	seedLiveNativeCredentials(t, fx)
	return fx, homes
}

func requireNativeSkill(t *testing.T, home, native string) {
	t.Helper()
	target := filepath.Join(native, "skills", "myskill")
	link, err := os.Readlink(target)
	if err != nil || !sameStoreTree(link, filepath.Join(home, "contexts")) {
		t.Fatalf("native skill link missing or outside store: %q %v", link, err)
	}
	marker := homeMarker(t, native)
	paths := marker.Surfaces[envmarker.SurfaceSkills].Paths
	if len(paths) != 1 || paths[0] != "skills/myskill" {
		t.Fatalf("native skill record = %v", paths)
	}
}

// Use and sync create and record native links for all three in-place adapters; opencode retains staging.
func TestNativeSkillsUseAndSync(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	for _, op := range []string{"use", "sync"} {
		var err error
		if op == "use" {
			_, err = UseWithPolicy(fx.home, fx.profile, "", "", false, Policy{})
		} else {
			_, err = SyncWithPolicy(fx.home, Policy{})
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{envregistry.ClaudeCode, envregistry.CodexCLI, envregistry.Pi} {
			requireNativeSkill(t, fx.home, homes[id])
		}
	}
	if _, err := os.Readlink(filepath.Join(homes["opencode"], ".agent-context", "skills", "myskill")); err != nil {
		t.Fatal(err)
	}
}

// Foreign links require takeover and preserve exact relative link text in backup; all unrecorded plain/store entries require consent too.
func TestNativeSkillsTakeover(t *testing.T) {
	for _, kind := range []string{"foreign", "directory", "file", "store", "relative-store"} {
		t.Run(kind, func(t *testing.T) {
			fx, homes := nativeSkillsFixture(t)
			native := homes["claude_code"]
			target := filepath.Join(native, "skills", "myskill")
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				t.Fatal(err)
			}
			link := "../../retired/missing"
			if kind == "store" {
				link = filepath.Join(fx.home, "contexts", "missing")
			}
			if kind == "relative-store" {
				var err error
				link, err = filepath.Rel(filepath.Dir(target), filepath.Join(fx.home, "contexts", "missing"))
				if err != nil {
					t.Fatal(err)
				}
			}
			switch kind {
			case "directory":
				if err := os.Mkdir(target, 0755); err != nil {
					t.Fatal(err)
				}
				writeGitFile(t, target, "keep", "original")
			case "file":
				if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
					t.Fatal(err)
				}
			default:
				if err := os.Symlink(link, target); err != nil {
					t.Fatal(err)
				}
			}
			before := nativeSnapshot(t, native)
			results, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{})
			diagnostic := DiagUnmanagedConflict
			if kind == "foreign" {
				diagnostic = DiagForeignManager
			}
			assertSwitchDiagnostic(t, err, results, &diagnostic)
			if after := nativeSnapshot(t, native); after != before {
				t.Fatal("refusal wrote native home or backup")
			}
			if _, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{Takeover: true}); err != nil {
				t.Fatal(err)
			}
			backup := filepath.Join(native, ".agent-environment-backup", "1", "skills", "myskill")
			if kind == "directory" {
				if got := homeBytes(t, backup, "keep"); got != "original" {
					t.Fatal(got)
				}
			} else if kind == "file" {
				if got := homeBytes(t, filepath.Dir(backup), "myskill"); got != "original" {
					t.Fatal(got)
				}
			} else if got, err := os.Readlink(backup); err != nil || got != link {
				t.Fatalf("backup link text = %q %v; want %q", got, err, link)
			}
			requireNativeSkill(t, fx.home, native)
		})
	}
}

func nativeSnapshot(t *testing.T, root string) string {
	t.Helper()
	var out strings.Builder
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == root {
				return nil
			}
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out.WriteString(rel + ":" + entry.Type().String() + ":")
		if entry.Type()&os.ModeSymlink != 0 {
			value, err := os.Readlink(path)
			if err != nil {
				return err
			}
			out.WriteString(value)
		} else if !entry.IsDir() {
			value, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			out.Write(value)
		}
		out.WriteByte('\n')
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// Neither takeover setting may traverse a linked skills directory or modify its external destination.
func TestNativeSkillsLinkedParent(t *testing.T) {
	for _, takeover := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary", true: "takeover"}[takeover], func(t *testing.T) {
			fx, homes := nativeSkillsFixture(t)
			native, outside := homes["claude_code"], t.TempDir()
			if err := os.Symlink(outside, filepath.Join(native, "skills")); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, native)
			results, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{Takeover: takeover})
			diagnostic := DiagWriteWouldFollowLink
			assertSwitchDiagnostic(t, err, results, &diagnostic)
			if nativeSnapshot(t, native) != before || nativeSnapshot(t, outside) != ".:d---------:\n" {
				t.Fatal("linked parent received writes")
			}
		})
	}
}

// Ledger ownership overrides even an old environment record; the entry and ledger survive takeover and status reports non-current.
func TestNativeSkillsLedger(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	if _, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	native := homes["claude_code"]
	ledger := filepath.Join(native, "skills", ".csk-managed.json")
	if err := os.WriteFile(ledger, []byte(`{"schema_version":1,"entries":["myskill"]}`), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.Readlink(filepath.Join(native, "skills", "myskill"))
	if err != nil {
		t.Fatal(err)
	}
	results, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{Takeover: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(results[0].Detail, "legacy global scope") {
		t.Fatalf("ledger entry not reported: %+v", results)
	}
	if got, err := os.Readlink(filepath.Join(native, "skills", "myskill")); err != nil || got != before {
		t.Fatal("ledger entry changed")
	}
	if got := homeBytes(t, native, "skills/.csk-managed.json"); got != `{"schema_version":1,"entries":["myskill"]}` {
		t.Fatal("ledger changed")
	}
	if paths := homeMarker(t, native).Surfaces[envmarker.SurfaceSkills].Paths; len(paths) != 0 {
		t.Fatalf("environment claimed ledger entry: %v", paths)
	}
	status, err := StatusOf(StatusRequest{Home: fx.home, NativeHomeOf: func(id string) (string, error) { return homes[id], nil }})
	if err != nil {
		t.Fatal(err)
	}
	if !status.NonCurrent || !strings.Contains(statusNativeDetails(status), "legacy global scope") {
		t.Fatalf("ledger status not non-current: %+v", status.Scopes)
	}
}

func statusNativeDetails(status *Status) string {
	var out strings.Builder
	for _, scope := range status.Scopes {
		if scope.NativeSkills != nil {
			out.WriteString(scope.NativeSkills.Detail)
		}
	}
	return out.String()
}

// An unreadable native marker fails closed before replacement or backup even with takeover.
func TestNativeSkillsUnreadableMarker(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	native := homes["claude_code"]
	if err := os.Mkdir(filepath.Join(native, envmarker.Name), 0755); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, native)
	results, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{Takeover: true})
	diagnostic := envmarker.DiagMarkerUnreadable
	assertSwitchDiagnostic(t, err, results, &diagnostic)
	if nativeSnapshot(t, native) != before {
		t.Fatal("unreadable marker allowed writes")
	}
}

// Current-profile repair restores a drifted native entry while explicit non-current repair leaves native state alone.
func TestNativeSkillsRepair(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	requireNativeSkill(t, fx.home, homes["codex_cli"])
	target := filepath.Join(homes["codex_cli"], "skills", "myskill")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../drift", target); err != nil {
		t.Fatal(err)
	}
	status, err := StatusOf(StatusRequest{Home: fx.home, NativeHomeOf: req.NativeHomeOf})
	if err != nil || !strings.Contains(statusNativeDetails(status), "drift") {
		t.Fatalf("drift status = %+v %v", status, err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	requireNativeSkill(t, fx.home, homes["codex_cli"])
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	beforeDryRun := nativeSnapshot(t, homes["codex_cli"])
	dry := req
	dry.DryRun = true
	if _, err := Resolve(dry); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("dry-run missed native drift: %v", err)
	}
	if nativeSnapshot(t, homes["codex_cli"]) != beforeDryRun {
		t.Fatal("dry-run wrote native home")
	}
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	if err := SetCurrent(fx.home, DefaultProfile); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, homes["codex_cli"])
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	if nativeSnapshot(t, homes["codex_cli"]) != before {
		t.Fatal("non-current repair rewrote native home")
	}
}

// Recorded staging paths converge and profile switching removes only recorded skills, preserving unnamed foreign bytes and links.
func TestNativeSkillsConvergenceAndSwitch(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	if _, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	native := homes["claude_code"]
	target, old := filepath.Join(native, "skills", "myskill"), filepath.Join(native, ".agent-context", "skills", "myskill")
	if err := os.MkdirAll(filepath.Dir(old), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(target, old); err != nil {
		t.Fatal(err)
	}
	marker := homeMarker(t, native)
	surface := marker.Surfaces[envmarker.SurfaceSkills]
	surface.Paths = []string{".agent-context/skills/myskill"}
	marker.Surfaces[envmarker.SurfaceSkills] = surface
	payload, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, envmarker.Name), payload, 0644); err != nil {
		t.Fatal(err)
	}
	writeGitFile(t, native, "skills/unnamed/keep", "untouched")
	if err := os.Symlink("../missing", filepath.Join(native, "skills", "unnamed-link")); err != nil {
		t.Fatal(err)
	}
	untouched := nativeSnapshot(t, filepath.Join(native, "skills", "unnamed"))
	if _, err := SyncWithPolicy(fx.home, Policy{Takeover: true}); err != nil {
		t.Fatal(err)
	}
	requireNativeSkill(t, fx.home, native)
	if _, err := os.Lstat(old); !os.IsNotExist(err) {
		t.Fatal("old staging entry retained")
	}
	if _, err := UseWithPolicy(fx.home, DefaultProfile, "claude_code", "", false, Policy{Takeover: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(target); !os.IsNotExist(err) {
		t.Fatal("retired recorded skill retained")
	}
	if nativeSnapshot(t, filepath.Join(native, "skills", "unnamed")) != untouched {
		t.Fatal("unnamed skill changed")
	}
	if got, err := os.Readlink(filepath.Join(native, "skills", "unnamed-link")); err != nil || got != "../missing" {
		t.Fatal("unnamed link changed")
	}
}

// Install --use and update replace native skill links for every participating adapter and preserve unnamed entries.
func TestNativeSkillsInstallAndUpdate(t *testing.T) {
	pinOperatorHome(t, t.TempDir())
	homes := pinHomes(t)
	home := t.TempDir()
	ids := newGitIdentities(t)
	repo := gitRepo(t, map[string]string{"SKILL.md": "# myskill\n"}, "v1.0.0")
	operand := ids.serve(repo, "https://example.com/native-skill")
	source := t.TempDir()
	writeManifestPackage(t, source, `{"schema_version":1,"name":"native-profile","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"myskill":{"git":"`+operand+`","range":"*"}}}}`, map[string]string{"a.md": "original\n"})
	if _, _, _, err := Install(home, InstallOptions{Operand: source, Use: true}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"claude_code", "codex_cli", "pi"} {
		requireNativeSkill(t, home, homes[id])
		writeGitFile(t, homes[id], "skills/unnamed/keep", "untouched")
	}
	if err := os.WriteFile(filepath.Join(source, "context", "a.md"), []byte("updated\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := UpdateWithPolicy(home, "native-profile", Policy{Takeover: true}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"claude_code", "codex_cli", "pi"} {
		requireNativeSkill(t, home, homes[id])
		if got := homeBytes(t, homes[id], "skills/unnamed/keep"); got != "untouched" {
			t.Fatal("update changed unnamed entry")
		}
	}
	second := t.TempDir()
	writeManifestPackage(t, second, `{"schema_version":1,"name":"second-profile","version":"1.0.0","requires":{"skills":{"another":{"git":"`+operand+`","range":"*"}}}}`, nil)
	if _, _, _, err := Install(home, InstallOptions{Operand: second}); err != nil {
		t.Fatal(err)
	}
	if _, err := UseWithPolicy(home, "second-profile", "codex_cli", "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	if _, err := UseWithPolicy(home, "native-profile", "", "", false, Policy{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Readlink(filepath.Join(homes["codex_cli"], "skills", "another")); err != nil {
		t.Fatalf("new scoped skill missing: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(homes["codex_cli"], "skills", "myskill")); !os.IsNotExist(err) {
		t.Fatal("scope retained old skill")
	}
	for _, id := range []string{"claude_code", "pi"} {
		requireNativeSkill(t, home, homes[id])
	}

}

// A stale directory listing cannot authorize reuse of an existing next backup generation; Use leaves every native byte unchanged.
func TestNativeSkillsBackupCollision(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	native := homes["claude_code"]
	writeGitFile(t, native, "skills/myskill/keep", "operator")
	writeGitFile(t, native, ".agent-environment-backup/1/keep", "existing generation")
	before := nativeSnapshot(t, native)
	policy := Policy{Takeover: true, backupReadDirectory: func(path string) (stateread.Directory, error) {
		listing, err := stateread.ReadDir(path)
		listing.Entries = nil
		return listing, err
	}}
	results, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, policy)
	diagnostic := DiagBackupExists
	assertSwitchDiagnostic(t, err, results, &diagnostic)
	if nativeSnapshot(t, native) != before {
		t.Fatal("backup collision partially wrote native home")
	}
	// The same real entry point accepts the unchanged fixture when it observes generation 1 and selects generation 2.
	if _, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{Takeover: true}); err != nil {
		t.Fatal(err)
	}
	requireNativeSkill(t, fx.home, native)
	if got := homeBytes(t, native, ".agent-environment-backup/2/skills/myskill/keep"); got != "operator" {
		t.Fatal(got)
	}
}

// A ledgered unrecorded foreign link is neither adopted nor backed up; unreadable ownership state refuses before native writes.
func TestNativeSkillsUnrecordedLedger(t *testing.T) {
	for _, unreadable := range []bool{false, true} {
		t.Run(map[bool]string{false: "ledgered", true: "unreadable-ledger"}[unreadable], func(t *testing.T) {
			fx, homes := nativeSkillsFixture(t)
			native := homes["claude_code"]
			writeGitFile(t, native, "skills/.csk-managed.json", `{"schema_version":1,"entries":["myskill"]}`)
			ledger := filepath.Join(native, "skills", ".csk-managed.json")
			if unreadable {
				if err := os.Remove(ledger); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(ledger, 0755); err != nil {
					t.Fatal(err)
				}
			}
			target := filepath.Join(native, "skills", "myskill")
			if err := os.Symlink("../../legacy/missing", target); err != nil {
				t.Fatal(err)
			}
			before := nativeSnapshot(t, native)
			results, err := UseWithPolicy(fx.home, fx.profile, "claude_code", "", false, Policy{Takeover: true})
			if unreadable {
				if err == nil || !strings.Contains(results[0].Detail, stateread.DiagUnreadable) {
					t.Fatalf("ownership read did not refuse: %v %+v", err, results)
				}
				if nativeSnapshot(t, native) != before {
					t.Fatal("unreadable ledger allowed writes")
				}
				return
			}
			if err != nil || !strings.Contains(results[0].Detail, "legacy global scope") {
				t.Fatalf("ledger entry not reported: %v %+v", err, results)
			}
			if got, err := os.Readlink(target); err != nil || got != "../../legacy/missing" {
				t.Fatal("ledger entry changed")
			}
			if _, err := os.Lstat(filepath.Join(native, ".agent-environment-backup")); !os.IsNotExist(err) {
				t.Fatal("ledger entry backed up")
			}
		})
	}
}

// A skills-free profile never traverses or modifies an unnamed native skills directory, even if that directory is a link.
func TestNativeSkillsEmptyProfileLeavesUnnamedParent(t *testing.T) {
	fx, homes := nativeSkillsFixture(t)
	native, outside := homes["claude_code"], t.TempDir()
	writeGitFile(t, outside, "keep", "untouched")
	if err := os.Symlink(outside, filepath.Join(native, "skills")); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, outside)
	if _, err := UseWithPolicy(fx.home, DefaultProfile, "claude_code", "", false, Policy{Takeover: true}); err != nil {
		t.Fatal(err)
	}
	if nativeSnapshot(t, outside) != before {
		t.Fatal("unnamed native directory changed")
	}
}
