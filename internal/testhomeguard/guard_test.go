package testhomeguard

import (
	"encoding/json"
	"github.com/relux-works/curator/internal/testgitenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Copied skill trees may contain credential carriers and their backup names.
// They retain only metadata; an adjacent instruction file is actually hashed.
func TestCredentialCarriersMetadataOnly(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "skills", "copied")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	names := []string{"auth.json", "auth.json.previous", "oauth_creds.json", "auth-backup.json", "secret.txt", "id_ed25519", ".env.local", ".npmrc"}
	for _, name := range append(names, "SKILL.md") {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("synthetic fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	image, err := snapshot([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		v := image[filepath.Join(dir, name)]
		if v.info == nil || v.hash != ([32]byte{}) {
			t.Fatalf("credential carrier was not metadata-only: %s", name)
		}
	}
	if image[filepath.Join(dir, "SKILL.md")].hash == ([32]byte{}) {
		t.Fatal("instruction positive control was not hashed")
	}
}

// F2: the neutral helper supplies the isolated HOME. Its retained provider
// override must fail our pre-test check; clearing that synthetic override is
// the valid runner control. No provider or native writer is called here.
func TestNeutralIsolationBoundary(t *testing.T) {
	for _, name := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "PI_CODING_AGENT_DIR", "XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME"} {
		t.Setenv(name, "")
	}
	previous, _ := os.UserHomeDir()
	restore := testgitenv.Isolate()
	defer restore()
	if err := CheckIsolation(previous); err != nil {
		t.Fatalf("neutral helper control: %v", err)
	}
	t.Setenv("CODEX_HOME", t.TempDir())
	if err := CheckIsolation(previous); err == nil {
		t.Fatal("retained ambient provider override admitted")
	}
}

// Oversized instruction files refuse before reading; credential carriers are
// metadata-only even when their size exceeds the content bound. A nearby small
// context file remains observable. All files are synthetic sparse fixtures.
func TestSnapshotContentBounds(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "auth.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(maxFileBytes + 1); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	image, err := snapshot([]string{root})
	if err != nil || image[path].hash != ([32]byte{}) {
		t.Fatalf("credential metadata control: %v", err)
	}
	if err := os.Rename(path, filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot([]string{root}); err == nil {
		t.Fatal("oversized context read admitted")
	}
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("bounded"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot([]string{root}); err != nil {
		t.Fatal(err)
	}
}

// Runtime sibling churn preserves directory identity. Each protected mutation
// must still change the snapshot; all writes stay within a synthetic root.
func TestSnapshotIdentity(t *testing.T) {
	for _, shape := range []string{"runtime-sibling", "context-bytes", "context-link", "marker-bytes", "settings-bytes", "nested-skill-bytes", "credential-metadata", "skill-link", "directory-replacement"} {
		t.Run(shape, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "native")
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			must(os.MkdirAll(filepath.Join(root, "skills"), 0700))
			must(os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("before"), 0600))
			must(os.WriteFile(filepath.Join(root, ".agent-environment.json"), []byte("before"), 0600))
			must(os.WriteFile(filepath.Join(root, "settings.json"), []byte("before"), 0600))
			must(os.MkdirAll(filepath.Join(root, "skills", "copied", "nested"), 0700))
			must(os.WriteFile(filepath.Join(root, "skills", "copied", "nested", "SKILL.md"), []byte("before"), 0600))
			must(os.WriteFile(filepath.Join(root, "auth.json"), []byte("synthetic credential carrier"), 0600))
			if shape == "context-link" || shape == "skill-link" {
				path := filepath.Join(root, "AGENTS.md")
				if shape == "skill-link" {
					path = filepath.Join(root, "skills", "myskill")
				}
				if err := os.Symlink("first", path); err != nil {
					t.Skipf("symlink unavailable: %v", err)
				}
			}
			before, err := snapshot([]string{root})
			must(err)
			if before[filepath.Join(root, "auth.json")].hash != ([32]byte{}) {
				t.Fatal("credential contents were hashed")
			}
			switch shape {
			case "runtime-sibling":
				must(os.WriteFile(filepath.Join(root, "runtime-sidecar"), []byte("unrelated"), 0600))
				// Force a metadata change even on coarse timestamp filesystems.
				must(os.Chtimes(root, time.Unix(1, 0), time.Unix(1, 0)))
			case "context-bytes", "marker-bytes", "settings-bytes", "nested-skill-bytes":
				name := "CLAUDE.md"
				if shape == "marker-bytes" {
					name = ".agent-environment.json"
				}
				if shape == "settings-bytes" {
					name = "settings.json"
				}
				if shape == "nested-skill-bytes" {
					name = filepath.Join("skills", "copied", "nested", "SKILL.md")
				}
				path := filepath.Join(root, name)
				info, err := os.Stat(path)
				must(err)
				must(os.WriteFile(path, []byte("after!"), 0600))
				must(os.Chtimes(path, info.ModTime(), info.ModTime()))
			case "credential-metadata":
				must(os.Chtimes(filepath.Join(root, "auth.json"), time.Unix(1, 0), time.Unix(1, 0)))
			case "context-link", "skill-link":
				path := filepath.Join(root, "AGENTS.md")
				if shape == "skill-link" {
					path = filepath.Join(root, "skills", "myskill")
				}
				must(os.Remove(path))
				must(os.Symlink("second", path))
			case "directory-replacement":
				must(os.Rename(root, root+"-old"))
				must(os.Mkdir(root, 0700))
			}
			after, err := snapshot([]string{root})
			must(err)
			wantSame := shape == "runtime-sibling"
			if same(before, after) != wantSame {
				t.Fatalf("snapshot equality = %v, want %v", !wantSame, wantSame)
			}
		})
	}
}

const fixtureBucketID = "01234567-0123-0123-0123-0123456789ab_abcdef01-abcd-abcd-abcd-abcdef012345"

// Root runtime churn and regular leaves of a registered native runtime bucket
// are ignored. Profile payloads at adjacent and deeper paths still fail with an
// equal-size/restored-mtime content mutation, including manifests and backups.
func TestRuntimeChurnAndRenderedPayloads(t *testing.T) {
	root := t.TempDir()
	bucket := "skills/synced/" + fixtureBucketID
	protected := []string{"settings.json", ".claude.json", ".agent-environment.json", ".agent-environment-backup/settings.json", ".agent-context/modules/fixture/instruction.md", "skills/copied/manifest.json", bucket + "/SKILL.md", bucket + "/nested/manifest.json"}
	write := func(name, body string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range protected {
		write(name, "before")
	}
	write("skills/synced/.bucket-"+fixtureBucketID, "registered")
	write(bucket+"/manifest.json", "before")
	write(bucket+"/.last-complete-round", "before")
	before, err := snapshot([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	// These are native-root runtime names, never exclusions inside skill payloads.
	for _, name := range []string{"sessions/event", "history/event", "history.jsonl", "sqlite/state", "state_5.sqlite", "state_5.sqlite-wal", "state_5.sqlite-shm", "models_cache/event", "models_cache.json", "global-state/event", "global-state.json", ".global-state.json", "shell-snapshots/event", "todos/event", "statsig/event", "backups/event", "logs/event", "projects/event", bucket + "/manifest.json", bucket + "/.last-complete-round"} {
		write(name, "runtime")
	}
	after, err := snapshot([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if !same(before, after) {
		t.Fatal("runtime-only churn changed protected image")
	}
	for _, name := range protected {
		t.Run(name, func(t *testing.T) {
			before, err := snapshot([]string{root})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, filepath.FromSlash(name))
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			write(name, "after!")
			if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
				t.Fatal(err)
			}
			after, err := snapshot([]string{root})
			if err != nil {
				t.Fatal(err)
			}
			if same(before, after) {
				t.Fatal("rendered payload mutation was excluded")
			}
		})
	}
}

// The optional private report names the exact diff and lstat metadata without
// recording protected contents or their hashes. Unchanged images stay empty.
func TestPrivateChangedPathReport(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.json")
	if err := os.WriteFile(path, []byte("before-payload"), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := snapshot([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	report := t.TempDir()
	t.Setenv("CURATOR_TEST_HOME_GUARD_REPORT", report)
	g := &Guard{roots: []string{root}, before: before}
	if g.Finish(0) != 0 {
		t.Fatal("unchanged control failed")
	}
	if entries, err := os.ReadDir(report); err != nil || len(entries) != 0 {
		t.Fatal("unchanged control wrote a diff")
	}
	if err := os.WriteFile(path, []byte("after--payload"), 0600); err != nil {
		t.Fatal(err)
	}
	if g.Finish(0) != 1 {
		t.Fatal("rendered write admitted")
	}
	entries, err := os.ReadDir(report)
	if err != nil || len(entries) != 1 {
		t.Fatal("private report missing")
	}
	raw, err := os.ReadFile(filepath.Join(report, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var row map[string]any
	if err := json.Unmarshal(raw, &row); err != nil {
		t.Fatal(err)
	}
	if row["path"] != path || row["before"] == nil || row["after"] == nil {
		t.Fatal("exact metadata diff missing")
	}
	for _, forbidden := range []string{"before-payload", "after--payload", "hash"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatal("report disclosed content identity")
		}
	}
}

// Account-level Claude runtime state is outside the managed-home footprint;
// a native context write inside that footprint still changes its image.
func TestAccountRuntimeOutsideManagedHomes(t *testing.T) {
	home := t.TempDir()
	roots := nativeHomePaths(home)
	before, err := snapshot(roots)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude.json"), []byte("runtime"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err := snapshot(roots)
	if err != nil {
		t.Fatal(err)
	}
	if !same(before, after) {
		t.Fatal("account runtime state protected")
	}
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".claude", "CLAUDE.md"), []byte("profile"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err = snapshot(roots)
	if err != nil {
		t.Fatal(err)
	}
	if same(before, after) {
		t.Fatal("native context write omitted")
	}
}
