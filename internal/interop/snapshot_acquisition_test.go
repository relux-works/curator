package interop

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
)

// snapshotAcquisitionVector mirrors vectors/snapshot-acquisition.json
// (environments §1.2). Fields the test does not consume are left undeclared.
type snapshotAcquisitionVector struct {
	Cases []struct {
		Name           string `json:"name"`
		Fixture        string `json:"fixture"`
		Expected       string `json:"expected"`
		ExpectedSHA256 string `json:"expected_sha256"`
		Files          []struct {
			Path   string `json:"path"`
			Bytes  int64  `json:"bytes"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	} `json:"cases"`
}

func acquisitionGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestConformanceSnapshotAcquisition drives gitops.Extract (the production
// acquisition path of internal/snapshot and internal/closure) against the
// suite's byte-exact vector. A root without the vector (the pinned rc.9 suite)
// is skipped with that reason; a root that has it must pass.
func TestConformanceSnapshotAcquisition(t *testing.T) {
	root := suiteRoot(t)
	vectorPath := filepath.Join(root, "vectors", "snapshot-acquisition.json")
	payload, err := os.ReadFile(vectorPath)
	if errors.Is(err, fs.ErrNotExist) {
		t.Skipf("conformance root %s has no vectors/snapshot-acquisition.json (pre-environments suite)", root)
	}
	if err != nil {
		t.Fatalf("reading %s: %v", vectorPath, err)
	}
	var vector snapshotAcquisitionVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatalf("decoding %s: %v", vectorPath, err)
	}
	if len(vector.Cases) == 0 {
		t.Fatalf("%s declares no cases", vectorPath)
	}
	for _, tc := range vector.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			fixture := filepath.Join(root, filepath.FromSlash(tc.Fixture))
			wantHash := readGolden(t, tc.Expected)
			if wantHash != tc.ExpectedSHA256 {
				t.Fatalf("expected file %s (%s) disagrees with vector expected_sha256 (%s)", tc.Expected, wantHash, tc.ExpectedSHA256)
			}
			// Commit every listed file with its exact bytes, bypassing the
			// fixture's own "* text=auto" attribute.
			repo := t.TempDir()
			acquisitionGit(t, repo, "init", "-q", "-b", "main")
			var wantPaths []string
			for _, file := range tc.Files {
				source := filepath.Join(fixture, filepath.FromSlash(file.Path))
				raw, err := os.ReadFile(source)
				if err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(raw)
				if got := "sha256:" + hex.EncodeToString(sum[:]); got != file.SHA256 || int64(len(raw)) != file.Bytes {
					t.Fatalf("fixture %s on disk (%s, %d bytes) does not match the vector (%s, %d bytes); the checkout normalized it", file.Path, got, len(raw), file.SHA256, file.Bytes)
				}
				oid := acquisitionGit(t, repo, "hash-object", "-w", "--no-filters", "--", source)
				acquisitionGit(t, repo, "update-index", "--add", "--cacheinfo", "100644,"+oid+","+file.Path)
				wantPaths = append(wantPaths, file.Path)
			}
			sort.Strings(wantPaths)
			tree := acquisitionGit(t, repo, "write-tree")
			commit := acquisitionGit(t, repo, "commit-tree", tree, "-m", "vector")
			acquisitionGit(t, repo, "update-ref", "refs/heads/main", commit)

			for _, autocrlf := range []string{"true", "false"} {
				t.Run("autocrlf="+autocrlf, func(t *testing.T) {
					acquisitionGit(t, repo, "config", "core.autocrlf", autocrlf)
					dest := filepath.Join(t.TempDir(), "snap")
					if err := gitops.Extract(repo, commit, dest); err != nil {
						t.Fatal(err)
					}
					var gotPaths []string
					err := filepath.WalkDir(dest, func(path string, d fs.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !d.IsDir() {
							rel, _ := filepath.Rel(dest, path)
							gotPaths = append(gotPaths, filepath.ToSlash(rel))
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					sort.Strings(gotPaths)
					if strings.Join(gotPaths, "\n") != strings.Join(wantPaths, "\n") {
						t.Fatalf("snapshot entries %v, want exactly %v", gotPaths, wantPaths)
					}
					for _, file := range tc.Files {
						want, _ := os.ReadFile(filepath.Join(fixture, filepath.FromSlash(file.Path)))
						got, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(file.Path)))
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(got, want) {
							t.Fatalf("%s: snapshot bytes differ from the committed blob", file.Path)
						}
					}
					subst, _ := os.ReadFile(filepath.Join(dest, "subst.txt"))
					if !bytes.Contains(subst, []byte("$Format:%H$")) || !bytes.Contains(subst, []byte("$Format:%h$")) {
						t.Fatalf("export-subst placeholders expanded: %q", subst)
					}
					crlf, _ := os.ReadFile(filepath.Join(dest, "crlf.txt"))
					if !bytes.Contains(crlf, []byte("\r\n")) {
						t.Fatalf("crlf.txt lost CRLF: %q", crlf)
					}
					mixed, _ := os.ReadFile(filepath.Join(dest, "mixed.txt"))
					if !bytes.Contains(mixed, []byte("\r\n")) || !bytes.Contains(bytes.ReplaceAll(mixed, []byte("\r\n"), nil), []byte("\n")) {
						t.Fatalf("mixed.txt lost mixed endings: %q", mixed)
					}
					got, err := hashing.ContentSHA256(dest, nil)
					if err != nil {
						t.Fatal(err)
					}
					if got != wantHash {
						t.Fatalf("content hash %s, want %s", got, wantHash)
					}
				})
			}
		})
	}
}
