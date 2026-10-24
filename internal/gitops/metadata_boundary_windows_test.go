//go:build windows

package gitops

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

// Regular/isolated clones and fresh exact-commit fetches admit the adjacent supported roots and refuse
// the next UTF-16 unit before invoking Git or creating even a cache parent.
// Unicode and relative spellings must obey the same expanded absolute bound.
func TestWindowsCloneMetadataBoundary(t *testing.T) {
	src := makeRepo(t)
	commit := gitRun(t, src, "rev-parse", "v1^{commit}")
	for _, lane := range []struct {
		name  string
		clone func(string, string) error
	}{
		{"regular", Clone}, {"isolated", CloneIsolated},
		{"exact-commit", func(source, destination string) error {
			return FetchCommitFromURLIsolated(destination, fileURL(t, source), commit)
		}},
	} {
		t.Run(lane.name, func(t *testing.T) {
			for _, tc := range []struct {
				name              string
				length            int
				unicode, relative bool
			}{
				{"just-under", 132, false, false},
				{"at-bound", 133, false, false},
				{"over-bound", 134, false, false},
				{"unicode-valid", 133, true, false},
				{"unicode-refused", 134, true, false},
				{"relative-refused", 134, false, true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					root := cloneTestRoot(t)
					parent := filepath.Join(root, "new")
					name := ""
					if tc.unicode {
						name = "\U0001f600"
					}
					padding := tc.length - utf16PathLength(parent) - 1 - utf16PathLength(name)
					if padding < 1 {
						t.Fatal("fixture root leaves no room for boundary control")
					}
					dest := filepath.Join(parent, name+strings.Repeat("c", padding))
					if tc.relative {
						t.Chdir(root)
						dest, _ = filepath.Rel(root, dest)
					}
					err := lane.clone(src, dest)
					if tc.length > 133 {
						var refusal *MetadataPathError
						if !errors.As(err, &refusal) || refusal.Actual != tc.length+126 || refusal.Limit != 259 || refusal.RootLength != tc.length || refusal.RootLimit != 133 || !strings.Contains(err.Error(), "git_metadata_path_too_long") {
							t.Fatalf("want typed metadata refusal at the exact projected length: %v", err)
						}
						if _, err := os.Lstat(parent); !os.IsNotExist(err) {
							t.Fatalf("refused clone created cache parent: %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if lane.name == "exact-commit" {
						// Reuse the freshly initialized cache through every existing-
						// repository fetch entry; none may reselect the locked object.
						gitRun(t, dest, "remote", "add", "origin", fileURL(t, src))
						if err := FetchCommitIsolated(dest, commit); err != nil {
							t.Fatal(err)
						}
						if err := FetchURLIsolated(dest, fileURL(t, src)); err != nil {
							t.Fatal(err)
						}
						if err := FetchIsolated(dest); err != nil {
							t.Fatal(err)
						}
					}
					if err := Fetch(dest); err != nil {
						t.Fatal(err)
					}
					ref, err := Resolve(dest, "tag", "v1")
					if err != nil {
						t.Fatal(err)
					}
					snapshot := filepath.Join(t.TempDir(), "snapshot")
					if err := Extract(dest, ref.Commit, snapshot); err != nil {
						t.Fatal(err)
					}
					if body, err := os.ReadFile(filepath.Join(snapshot, "SKILL.md")); err != nil || string(body) != "v1" {
						t.Fatalf("admitted payload=%q: %v", body, err)
					}
				})
			}
		})
	}
}

// Existing short-name aliases and extended prefixes cannot lower the charged
// expanded-root length. The existing ancestor and sentinel remain untouched.
// When 8.3 names are disabled, the same test covers the returned long spelling.
func TestWindowsCloneMetadataAliases(t *testing.T) {
	root := cloneTestRoot(t)
	ancestor := filepath.Join(root, strings.Repeat("l", 134-utf16PathLength(root)-1))
	if err := os.MkdirAll(ancestor, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(ancestor, "sentinel")
	if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := windows.UTF16PtrFromString(ancestor)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]uint16, 32768)
	n, err := windows.GetShortPathName(p, &buf[0], uint32(len(buf)))
	if err != nil || n == 0 || int(n) >= len(buf) {
		t.Fatalf("short path control: %d %v", n, err)
	}
	alias := windows.UTF16ToString(buf[:n])
	t.Logf("8.3 alias available: %t", alias != ancestor)
	for _, parent := range []string{ancestor, alias, `\\?\` + ancestor} {
		var refusal *MetadataPathError
		err := CheckCloneDestination(filepath.Join(parent, "new"))
		if !errors.As(err, &refusal) || refusal.RootLength != utf16PathLength(filepath.Join(ancestor, "new")) {
			t.Fatalf("alias bypassed expanded metadata budget: %v", err)
		}
	}
	if body, err := os.ReadFile(marker); err != nil || string(body) != "keep" {
		t.Fatalf("ancestor changed: %q %v", body, err)
	}
	entries, err := os.ReadDir(ancestor)
	if err != nil || len(entries) != 1 {
		t.Fatalf("alias check wrote paths: %v %v", entries, err)
	}
}

// Junctions charge the longer lexical or resolved root in either direction;
// refusing a child must leave the linked target's sentinel and entries intact.
func TestWindowsCloneMetadataJunctions(t *testing.T) {
	for _, longLink := range []bool{false, true} {
		t.Run(map[bool]string{false: "long-target", true: "long-link"}[longLink], func(t *testing.T) {
			root := cloneTestRoot(t)
			// A short junction remains admissible beside the long-target
			// refusal; rejecting every junction cannot satisfy this control.
			validTarget, validLink := filepath.Join(root, "valid-target"), filepath.Join(root, "valid-link")
			if err := os.Mkdir(validTarget, 0o700); err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("cmd", "/c", "mklink", "/J", validLink, validTarget).CombinedOutput(); err != nil {
				t.Fatalf("valid junction fixture: %v: %s", err, out)
			}
			t.Cleanup(func() {
				if err := os.Remove(validLink); err != nil {
					t.Errorf("remove valid junction: %v", err)
				}
			})
			if err := CheckCloneDestination(filepath.Join(validLink, "new")); err != nil {
				t.Fatalf("short junction refused: %v", err)
			}
			long := filepath.Join(root, strings.Repeat("j", 134-utf16PathLength(root)-1))
			short := filepath.Join(root, "short")
			target, link := long, short
			if longLink {
				target, link = short, long
			}
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(target, "sentinel")
			if err := os.WriteFile(marker, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
				t.Fatalf("junction fixture: %v: %s", err, out)
			}
			t.Cleanup(func() {
				if err := os.Remove(link); err != nil {
					t.Errorf("remove junction: %v", err)
				}
			})
			var refusal *MetadataPathError
			if err := CheckCloneDestination(filepath.Join(link, "new")); !errors.As(err, &refusal) || refusal.RootLength != 138 {
				t.Fatalf("junction bypassed longest root: %v", err)
			}
			if body, err := os.ReadFile(marker); err != nil || string(body) != "keep" {
				t.Fatalf("target changed: %q %v", body, err)
			}
			if entries, err := os.ReadDir(target); err != nil || len(entries) != 1 {
				t.Fatalf("junction check wrote paths: %v %v", entries, err)
			}
		})
	}
}

// The admission applies to new caches only. Every existing-repository fetch
// entry still works at root 134 when its actual metadata fits Git's limit;
// Curator neither relocates nor rejects that pre-existing checkout.
func TestWindowsExistingCacheMetadataExemption(t *testing.T) {
	src := makeRepo(t)
	commit := gitRun(t, src, "rev-parse", "v1^{commit}")
	base := cloneTestRoot(t)
	repo := filepath.Join(base, strings.Repeat("e", 134-utf16PathLength(base)-1))
	gitRun(t, base, "-c", "core.longpaths=true", "clone", "--", src, repo)
	for _, lane := range []struct {
		name  string
		fetch func() error
	}{
		{"regular", func() error { return Fetch(repo) }},
		{"isolated", func() error { return FetchIsolated(repo) }},
		{"url", func() error { return FetchURLIsolated(repo, fileURL(t, src)) }},
		{"commit", func() error { return FetchCommitIsolated(repo, commit) }},
		{"commit-url", func() error { return FetchCommitFromURLIsolated(repo, fileURL(t, src), commit) }},
	} {
		t.Run(lane.name, func(t *testing.T) {
			if err := lane.fetch(); err != nil {
				t.Fatalf("existing cache unexpectedly subject to new-root admission: %v", err)
			}
			ref, err := Resolve(repo, "tag", "v1")
			if err != nil || ref.Commit != commit {
				t.Fatalf("existing locked object changed: %+v %v", ref, err)
			}
		})
	}
}
