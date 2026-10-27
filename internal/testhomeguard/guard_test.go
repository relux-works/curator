package testhomeguard

import (
	"github.com/relux-works/curator/internal/testgitenv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
