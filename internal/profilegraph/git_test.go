package profilegraph

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/profilegraph/graphtest"
)

// Real Git produces both REF_DELTA and OFS_DELTA packs in isolated fixtures.
// The producer reads their exact verified blobs without starting Git itself;
// an unavailable object id refuses instead of returning substitute bytes.
func TestOfflinePackedObjects(t *testing.T) {
	for _, offset := range []bool{false, true} {
		home := t.TempDir()
		source := "example.invalid/packed"
		first := bytes.Repeat([]byte("Synthetic packed instructions with a stable common prefix.\n"), 150)
		second := append(append([]byte{}, first...), []byte("A distinct final sentence.\n")...)
		commit, err := graphtest.Commit(home, source, map[string][]byte{"a.md": first, "b.md": second})
		if err != nil {
			t.Fatal(err)
		}
		root := filepath.Join(envprofile.GraphRepository(home, source), ".git")
		if err := graphtest.Write(root, "HEAD", []byte("ref: refs/heads/main\n")); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "refs"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "objects/pack"), 0700); err != nil {
			t.Fatal(err)
		}
		var ids, paths []string
		err = filepath.WalkDir(filepath.Join(root, "objects"), func(p string, e os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(filepath.Join(root, "objects"), p)
			id := strings.ReplaceAll(filepath.ToSlash(rel), "/", "")
			if oid(id) {
				ids = append(ids, id)
				paths = append(paths, p)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		git := func(args ...string) []byte {
			t.Helper()
			cmd := exec.Command("git", append([]string{"--git-dir", root}, args...)...)
			for _, kv := range os.Environ() {
				k, _, _ := strings.Cut(kv, "=")
				if k != "GIT_DIR" && k != "GIT_WORK_TREE" && k != "GIT_INDEX_FILE" && k != "GIT_COMMON_DIR" {
					cmd.Env = append(cmd.Env, kv)
				}
			}
			cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_COUNT=0")
			cmd.Stdin = strings.NewReader(strings.Join(ids, "\n") + "\n")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Git fixture pack: %v %s", err, out)
			}
			return out
		}
		args := []string{"pack-objects", "--window=10", "--depth=20", "--no-reuse-delta"}
		if offset {
			args = append(args, "--delta-base-offset")
		}
		args = append(args, filepath.Join(root, "objects/pack/pack"))
		pack := strings.TrimSpace(string(git(args...)))
		verification := git("verify-pack", "-v", filepath.Join(root, "objects/pack/pack-"+pack+".idx"))
		if !bytes.Contains(verification, []byte("chain length")) {
			t.Fatal("fixture did not exercise a delta")
		}
		for _, p := range paths {
			if err := os.Remove(p); err != nil {
				t.Fatal(err)
			}
		}
		b := &build{req: Request{Home: home}, r: NewReader(V1Limits)}
		raw, err := b.db(source).file(commit, "b.md")
		if err != nil || !bytes.Equal(raw, second) {
			t.Fatalf("offset=%v: %v", offset, err)
		}
		_, _, err = b.db(source).object(strings.Repeat("f", 40), 0)
		if err == nil || err.Error() != "graph_missing_input: graph input refused" {
			t.Fatalf("missing object: %v", err)
		}
	}
}
