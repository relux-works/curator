package testhomeguard

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Runtime sibling churn preserves directory identity. Each protected mutation
// must still change the snapshot; all writes stay within a synthetic root.
func TestSnapshotIdentity(t *testing.T) {
	for _, shape := range []string{"runtime-sibling", "context-bytes", "context-link", "marker-bytes", "skill-link", "directory-replacement"} {
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
			must(os.Symlink("first", filepath.Join(root, "AGENTS.md")))
			must(os.Symlink("first", filepath.Join(root, "skills", "myskill")))
			before, err := snapshot([]string{root})
			must(err)
			switch shape {
			case "runtime-sibling":
				must(os.WriteFile(filepath.Join(root, "runtime-sidecar"), []byte("unrelated"), 0600))
				// Force a metadata change even on coarse timestamp filesystems.
				must(os.Chtimes(root, time.Unix(1, 0), time.Unix(1, 0)))
			case "context-bytes", "marker-bytes":
				name := "CLAUDE.md"
				if shape == "marker-bytes" {
					name = ".agent-environment.json"
				}
				path := filepath.Join(root, name)
				info, err := os.Stat(path)
				must(err)
				must(os.WriteFile(path, []byte("after!"), 0600))
				must(os.Chtimes(path, info.ModTime(), info.ModTime()))
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
