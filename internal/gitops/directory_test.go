package gitops

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func directoryRepo(t *testing.T, links map[string]string) (string, string) {
	t.Helper()
	repo := makeRepo(t)
	for name, text := range map[string]string{"agents/skills/demo/SKILL.md": "selected\x00bytes\r\n", "sibling/ignored": "outside", "agents/skills/demo/run": "#!/bin/sh\n"} {
		p := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0755); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "add", ".")
	for name, target := range links {
		f := filepath.Join(t.TempDir(), "target")
		if err := os.WriteFile(f, []byte(target), 0600); err != nil {
			t.Fatal(err)
		}
		oid := gitRun(t, repo, "hash-object", "-w", "--no-filters", "--", f)
		gitRun(t, repo, "update-index", "--add", "--cacheinfo", "120000,"+oid+","+name)
	}
	tree := gitRun(t, repo, "write-tree")
	commit := gitRun(t, repo, "commit-tree", tree, "-m", "selected fixture")
	return repo, commit
}

// Sibling discovery links and files do not enter the selected closure; original
// blob bytes survive; executable mode is checked where portable permission bits
// exist. Full-root extraction still fails.
func TestExtractDirectoryOutsideLinks(t *testing.T) {
	repo, commit := directoryRepo(t, map[string]string{".claude/skills/demo": "../../agents/skills/demo", ".codex/skills/demo": "../../agents/skills/demo"})
	dst := filepath.Join(t.TempDir(), "snapshot")
	if err := ExtractDirectory(repo, commit, "agents/skills/demo", dst); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dst, "agents/skills/demo/SKILL.md"))
	if err != nil || string(b) != "selected\x00bytes\r\n" {
		t.Fatalf("selected bytes=%q err=%v", b, err)
	}
	for _, name := range []string{".claude", ".codex", "sibling", "SKILL.md"} {
		if _, err := os.Lstat(filepath.Join(dst, name)); !os.IsNotExist(err) {
			t.Fatalf("outside path %s materialized: %v", name, err)
		}
	}
	info, err := os.Stat(filepath.Join(dst, "agents/skills/demo/run"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
		t.Fatalf("executable mode lost: %v", info.Mode())
	}
	for _, dir := range []string{"", "."} {
		out := filepath.Join(t.TempDir(), "root")
		err := ExtractDirectory(repo, commit, dir, out)
		if err == nil {
			t.Fatalf("root selection %q admitted links", dir)
		}
		if _, e := os.Lstat(out); !os.IsNotExist(e) {
			t.Fatalf("root refusal wrote: %v", e)
		}
	}
}

// An in-selection link, symlink ancestor, missing/non-tree selection, or unsafe
// selector refuses before destination creation, beside the valid selection above.
func TestExtractDirectoryRefusals(t *testing.T) {
	for _, tc := range []struct {
		name, dir string
		links     map[string]string
		link      string
	}{
		{"inside", "agents/skills/demo", map[string]string{"agents/skills/demo/escape": "../../../../outside"}, "agents/skills/demo/escape"},
		{"ancestor", "alias/demo", map[string]string{"alias": "agents/skills"}, "alias"},
		{"leaf", "alias", map[string]string{"alias": "agents/skills/demo"}, "alias"},
		{"missing", "absent", nil, ""}, {"blob", "agents/skills/demo/SKILL.md", nil, ""},
		{"blob-ancestor", "agents/skills/demo/SKILL.md/child", nil, ""},
		{"parent", "agents/../skills", nil, ""}, {"absolute", "/agents/skills/demo", nil, ""},
		{"dot", "agents/./skills/demo", nil, ""}, {"backslash", `agents\skills\demo`, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, commit := directoryRepo(t, tc.links)
			dst := filepath.Join(t.TempDir(), "output")
			err := ExtractDirectory(repo, commit, tc.dir, dst)
			if err == nil {
				t.Fatal("invalid selection admitted")
			}
			if strings.HasPrefix(tc.name, "blob") && !strings.Contains(err.Error(), "is not a Git tree") {
				t.Fatalf("non-tree selection reached the wrong refusal: %v", err)
			}
			if tc.link != "" {
				var e *UnsupportedLinkError
				if !errors.As(err, &e) || e.Path != tc.link {
					t.Fatalf("wrong link refusal: %v", err)
				}
			}
			if _, err := os.Lstat(dst); !os.IsNotExist(err) {
				t.Fatalf("refusal wrote destination: %v", err)
			}
		})
	}
}

// Destination symlinks and existing files never redirect or overwrite a scoped
// extraction, and a failure leaves the pre-existing sentinel unchanged.
func TestExtractDirectoryDestinationGuards(t *testing.T) {
	repo, commit := directoryRepo(t, nil)
	for _, link := range []bool{false, true} {
		t.Run(map[bool]string{false: "existing", true: "symlink"}[link], func(t *testing.T) {
			dst := t.TempDir()
			outside := t.TempDir()
			target := filepath.Join(dst, "agents")
			if link {
				if err := os.Symlink(outside, target); err != nil {
					t.Fatal(err)
				}
			} else {
				target = filepath.Join(dst, "agents/skills/demo/SKILL.md")
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := ExtractDirectory(repo, commit, "agents/skills/demo", dst)
			if err == nil {
				t.Fatal("destination conflict admitted")
			}
			if link {
				files, e := os.ReadDir(outside)
				if e != nil || len(files) != 0 {
					t.Fatalf("escape wrote: %v", e)
				}
			} else {
				b, e := os.ReadFile(target)
				if e != nil || string(b) != "keep" {
					t.Fatalf("overwritten %q %v", b, e)
				}
			}
			if strings.Contains(err.Error(), "cat-file") {
				t.Fatalf("conflict was not refused before streaming: %v", err)
			}
		})
	}
}

// Git's own mktree is the independent reference for selected snapshot identity
// in both supported object formats. A full repository tree cannot stand in for
// the projection's pin even when all selected file bytes are identical.
func TestPinnedDirectorySnapshotTree(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			repo := t.TempDir()
			gitRun(t, repo, "init", "-q", "-b", "main", "--object-format="+format)
			for name, text := range map[string]string{"nested/selection/SKILL.md": "bytes\r\n", "outside": "unselected"} {
				path := filepath.Join(repo, name)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(text), 0644); err != nil {
					t.Fatal(err)
				}
			}
			gitRun(t, repo, "add", ".")
			gitRun(t, repo, "commit", "-qm", "tree identity fixture")
			commit := gitRun(t, repo, "rev-parse", "HEAD")
			selected := gitRun(t, repo, "rev-parse", commit+":nested/selection")
			for _, name := range []string{"selection", "nested"} {
				cmd := gitCommand("-C", repo, "mktree", "-z")
				cmd.Env = gitEnv()
				cmd.Stdin = strings.NewReader("040000 tree " + selected + "\t" + name + "\x00")
				b, err := cmd.Output()
				if err != nil {
					t.Fatal(err)
				}
				selected = strings.TrimSpace(string(b))
			}
			pin, gotFormat, err := PinnedDirectorySnapshotTree(repo, commit, "nested/selection")
			if err != nil || pin != selected || gotFormat != format {
				t.Fatalf("projection pin=%s format=%s err=%v want Git %s/%s", pin, gotFormat, err, selected, format)
			}
			dst := filepath.Join(t.TempDir(), "snapshot")
			if err := ExtractDirectory(repo, commit, "nested/selection", dst); err != nil {
				t.Fatal(err)
			}
			actual, err := TreeObjectIDFromDir(dst, format)
			if err != nil || actual != pin {
				t.Fatalf("snapshot identity=%s err=%v want %s", actual, err, pin)
			}
			whole, _, err := PinnedCommitTree(repo, commit)
			if err != nil || whole == pin {
				t.Fatalf("whole tree reused as selected identity: %s %v", whole, err)
			}
		})
	}
}

// A Git submodule inside or along the selection refuses without a destination;
// the same outside submodule leaves the valid selected package extractable.
// Git-index fixtures exercise this contract without OS symlink privileges.
func TestExtractDirectorySubmodules(t *testing.T) {
	for _, tc := range []struct{ name, submodule, selection, refusal string }{
		{"leaf", "dependency", "dependency", "is not a Git tree"},
		{"ancestor", "dependency", "dependency/child", "is not a Git tree"},
		{"inside", "agents/skills/demo/dependency", "agents/skills/demo", "submodule"},
		{"outside", "dependency", "agents/skills/demo", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, child := directoryRepo(t, nil)
			gitRun(t, repo, "update-index", "--add", "--cacheinfo", "160000,"+child+","+tc.submodule)
			tree := gitRun(t, repo, "write-tree")
			commit := gitRun(t, repo, "commit-tree", tree, "-m", "submodule selection fixture")
			dst := filepath.Join(t.TempDir(), "snapshot")
			err := ExtractDirectory(repo, commit, tc.selection, dst)
			if tc.refusal != "" {
				if err == nil || !strings.Contains(err.Error(), tc.refusal) {
					t.Fatalf("want %s refusal, got %v", tc.refusal, err)
				}
				if _, e := os.Lstat(dst); !os.IsNotExist(e) {
					t.Fatalf("refusal wrote destination: %v", e)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(dst, tc.selection, "SKILL.md"))
			if err != nil || string(b) != "selected\x00bytes\r\n" {
				t.Fatalf("selected bytes=%q err=%v", b, err)
			}
			if _, e := os.Lstat(filepath.Join(dst, tc.submodule)); !os.IsNotExist(e) {
				t.Fatalf("outside submodule materialized: %v", e)
			}
		})
	}
}
