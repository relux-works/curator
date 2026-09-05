package gitops

import (
	"bytes"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

// byteExactSHA256 is the expected core §8 content hash of the spec vector
// conformance/v1/vectors/snapshot-acquisition.json (curator-spec ec695ba).
const byteExactSHA256 = "sha256:500ea934403d10a2a0b6b7e8874790e489ee002328d3dc0edbda2fe5be2bced0"

var byteExactFiles = []string{".gitattributes", "crlf.txt", "lf.txt", "mixed.txt", "subst.txt"}

// fixtureName maps a tree path to its testdata file. The fixture's
// .gitattributes ("* text=auto") is stored under a neutral name so it cannot
// normalize its sibling fixtures inside curator's own index.
func fixtureName(path string) string {
	if path == ".gitattributes" {
		return "gitattributes.fixture"
	}
	return path
}

func readFixture(t *testing.T, fixture, path string) []byte {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(fixture, fixtureName(path)))
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

// commitExactTree commits every file of dir into a fresh repository with its
// exact bytes, bypassing attribute conversion (hash-object --no-filters +
// update-index --cacheinfo), so the in-tree "* text=auto" rule cannot
// normalize the blobs. modes maps a path to a tree mode; default 100644.
func commitExactTree(t *testing.T, dir string, modes map[string]string) (string, string) {
	t.Helper()
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range byteExactFiles {
		oid := gitRun(t, repo, "hash-object", "-w", "--no-filters", "--", filepath.Join(absDir, fixtureName(name)))
		mode := modes[name]
		if mode == "" {
			mode = "100644"
		}
		gitRun(t, repo, "update-index", "--add", "--cacheinfo", mode+","+oid+","+name)
	}
	tree := gitRun(t, repo, "write-tree")
	commit := gitRun(t, repo, "commit-tree", tree, "-m", "exact")
	gitRun(t, repo, "update-ref", "refs/heads/main", commit)
	return repo, commit
}

func TestExtractReproducesByteExactVector(t *testing.T) {
	fixture := filepath.Join("testdata", "byte-exact")
	repo, commit := commitExactTree(t, fixture, nil)
	// Prove the committed blobs equal the fixture bytes before acquiring.
	for _, name := range byteExactFiles {
		want := readFixture(t, fixture, name)
		cmd := exec.Command("git", "cat-file", "-p", commit+":"+name)
		cmd.Dir = repo
		cmd.Env = gitEnv()
		got, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("committed blob %s differs from fixture", name)
		}
	}
	for _, autocrlf := range []string{"true", "false"} {
		t.Run("autocrlf="+autocrlf, func(t *testing.T) {
			gitRun(t, repo, "config", "core.autocrlf", autocrlf)
			dest := filepath.Join(t.TempDir(), "snap")
			if err := Extract(repo, commit, dest); err != nil {
				t.Fatal(err)
			}
			assertByteExactSnapshot(t, fixture, dest)
		})
	}
}

// assertByteExactSnapshot checks the acquisition contract of the vector.
func assertByteExactSnapshot(t *testing.T, fixture, dest string) {
	t.Helper()
	var listed []string
	err := filepath.WalkDir(dest, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(dest, path)
			listed = append(listed, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(listed)
	if strings.Join(listed, ",") != strings.Join(byteExactFiles, ",") {
		t.Fatalf("snapshot files = %v, want %v", listed, byteExactFiles)
	}
	for _, name := range byteExactFiles {
		want := readFixture(t, fixture, name)
		got, err := os.ReadFile(filepath.Join(dest, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("%s: snapshot bytes differ from committed blob:\n got %q\nwant %q", name, got, want)
		}
	}
	subst, _ := os.ReadFile(filepath.Join(dest, "subst.txt"))
	if !bytes.Contains(subst, []byte("$Format:%H$")) || !bytes.Contains(subst, []byte("$Format:%h$")) {
		t.Fatalf("export-subst placeholder was expanded: %q", subst)
	}
	crlf, _ := os.ReadFile(filepath.Join(dest, "crlf.txt"))
	if !bytes.Contains(crlf, []byte("\r\n")) {
		t.Fatalf("crlf.txt lost CRLF: %q", crlf)
	}
	mixed, _ := os.ReadFile(filepath.Join(dest, "mixed.txt"))
	if !bytes.Contains(mixed, []byte("\r\n")) || !bytes.Contains(bytes.ReplaceAll(mixed, []byte("\r\n"), nil), []byte("\n")) {
		t.Fatalf("mixed.txt lost mixed endings: %q", mixed)
	}
	hash, err := hashing.ContentSHA256(dest, nil)
	if err != nil {
		t.Fatal(err)
	}
	if hash != byteExactSHA256 {
		t.Fatalf("content hash = %s, want %s", hash, byteExactSHA256)
	}
}

// TestExtractIgnoresWorkingTreeConversion is the negative proof that the
// git-archive path is gone: with "* text=auto" committed and
// core.autocrlf=true, "git archive" rewrites lf.txt to CRLF and expands the
// export-subst placeholder. Extraction must reproduce the committed bytes
// regardless; this test fails if extraction is switched back to git archive.
func TestExtractIgnoresWorkingTreeConversion(t *testing.T) {
	fixture := filepath.Join("testdata", "byte-exact")
	repo, commit := commitExactTree(t, fixture, nil)
	gitRun(t, repo, "config", "core.autocrlf", "true")

	// Establish that the conversion is live in this repository: git archive
	// under this configuration changes lf.txt and subst.txt.
	archive := exec.Command("git", "-C", repo, "archive", "--format=tar", commit)
	archive.Env = gitEnv()
	tarBytes, err := archive.Output()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(tarBytes, []byte("alpha\r\nbeta")) {
		t.Fatal("git archive did not apply core.autocrlf=true here; the negative control is not live")
	}
	if bytes.Contains(tarBytes, []byte("$Format:%H$")) {
		t.Fatal("git archive left export-subst unexpanded; the negative control is not live")
	}

	dest := filepath.Join(t.TempDir(), "snap")
	if err := Extract(repo, commit, dest); err != nil {
		t.Fatal(err)
	}
	lf, err := os.ReadFile(filepath.Join(dest, "lf.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(lf, []byte("\r")) {
		t.Fatalf("lf.txt acquired CRLF from core.autocrlf: %q", lf)
	}
	assertByteExactSnapshot(t, fixture, dest)
}

func TestExtractPreservesExecutableBit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose portable executable permission bits")
	}
	fixture := filepath.Join("testdata", "byte-exact")
	repo, commit := commitExactTree(t, fixture, map[string]string{"lf.txt": "100755"})
	dest := filepath.Join(t.TempDir(), "snap")
	if err := Extract(repo, commit, dest); err != nil {
		t.Fatal(err)
	}
	execInfo, err := os.Stat(filepath.Join(dest, "lf.txt"))
	if err != nil {
		t.Fatal(err)
	}
	plainInfo, err := os.Stat(filepath.Join(dest, "crlf.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if execInfo.Mode().Perm()&0o100 == 0 {
		t.Fatalf("100755 entry lost the executable bit: %v", execInfo.Mode())
	}
	if plainInfo.Mode().Perm()&0o111 != 0 {
		t.Fatalf("100644 entry gained an executable bit: %v", plainInfo.Mode())
	}
}

// commitRawTree writes a tree from raw ls-tree lines through mktree so tests
// can commit entries a working tree cannot express (submodules, escapes).
func commitRawTree(t *testing.T, repo string, lines []string) string {
	t.Helper()
	cmd := exec.Command("git", "mktree", "-z")
	cmd.Dir = repo
	cmd.Env = gitEnv()
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\x00") + "\x00")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mktree: %v\n%s", err, out)
	}
	tree := strings.TrimSpace(string(out))
	return gitRun(t, repo, "commit-tree", tree, "-m", "raw")
}

func blobOID(t *testing.T, repo, content string) string {
	t.Helper()
	cmd := exec.Command("git", "hash-object", "-w", "--stdin")
	cmd.Dir = repo
	cmd.Env = gitEnv()
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func TestExtractRefusesSubmodules(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "x")
	commit := commitRawTree(t, repo, []string{
		"100644 blob " + oid + "\tSKILL.md",
		"160000 commit " + strings.Repeat("a", 40) + "\tvendor",
	})
	err := Extract(repo, commit, filepath.Join(t.TempDir(), "snap"))
	if err == nil || !strings.Contains(err.Error(), "unsupported entry type") {
		t.Fatalf("err = %v, want gitlink refusal", err)
	}
}

// literalEntry is one raw tree entry for commitLiteralTree.
type literalEntry struct {
	name string
	oid  string
}

// commitLiteralTree writes a tree object with arbitrary entry names (git
// mktree refuses "..", ".", ".git" and slashes) via hash-object --literally.
func commitLiteralTree(t *testing.T, repo string, entries []literalEntry) string {
	t.Helper()
	var body bytes.Buffer
	for _, entry := range entries {
		raw, err := hex.DecodeString(entry.oid)
		if err != nil {
			t.Fatal(err)
		}
		body.WriteString("100644 " + entry.name + "\x00")
		body.Write(raw)
	}
	cmd := exec.Command("git", "hash-object", "-t", "tree", "-w", "--literally", "--stdin")
	cmd.Dir = repo
	cmd.Env = gitEnv()
	cmd.Stdin = &body
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hash-object tree: %v\n%s", err, out)
	}
	return gitRun(t, repo, "commit-tree", strings.TrimSpace(string(out)), "-m", "literal")
}

func TestExtractRefusesEscapingPaths(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "x")
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	for _, name := range []string{"..", ".", "a/../../escape2", "sub/./x", "", "/abs"} {
		commit := commitLiteralTree(t, repo, []literalEntry{{name: name, oid: oid}})
		err := Extract(repo, commit, dest)
		if err == nil {
			t.Fatalf("%q: extraction must be refused", name)
		}
		// git itself rejects an empty entry name while listing; every other
		// shape must reach and trip the package's own path check.
		if name != "" && !strings.Contains(err.Error(), "unsafe path") {
			t.Fatalf("%q: err = %v, want escape refusal", name, err)
		}
	}
	entries, _ := os.ReadDir(parent)
	for _, entry := range entries {
		if entry.Name() != "snap" {
			t.Fatalf("escaped write reached %s", entry.Name())
		}
	}
}

func TestExtractRefusesOversizeBlob(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "x")
	commit := commitRawTree(t, repo, []string{"100644 blob " + oid + "\tsmall"})
	saved := maxSnapshotFileBytes
	t.Cleanup(func() { maxSnapshotFileBytes = saved })
	maxSnapshotFileBytes = 0
	err := Extract(repo, commit, filepath.Join(t.TempDir(), "snap"))
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("err = %v, want size refusal", err)
	}
	maxSnapshotFileBytes = 1
	if err := Extract(repo, commit, filepath.Join(t.TempDir(), "snap2")); err != nil {
		t.Fatalf("a blob at the bound must be accepted: %v", err)
	}
}

func TestExtractRefusesDuplicatePlatformPaths(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "x")
	dest := filepath.Join(t.TempDir(), "snap")
	probe := filepath.Join(t.TempDir(), "probe")
	if err := os.WriteFile(probe, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	upper := filepath.Join(filepath.Dir(probe), "PROBE")
	_, caseInsensitive := os.Stat(upper)
	commit := commitRawTree(t, repo, []string{
		"100644 blob " + oid + "\tReadme.md",
		"100644 blob " + oid + "\treadme.md",
	})
	err := Extract(repo, commit, dest)
	if caseInsensitive == nil {
		if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
			t.Fatalf("err = %v, want platform-path collision refusal on a case-insensitive filesystem", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("case-distinct paths must coexist on a case-sensitive filesystem: %v", err)
	}
}

func TestExtractRefusesExistingDestinationEntries(t *testing.T) {
	// A pre-existing file at a tree path is a collision, never overwritten.
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "new")
	commit := commitRawTree(t, repo, []string{"100644 blob " + oid + "\tSKILL.md"})
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Extract(repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
		t.Fatalf("err = %v, want refusal", err)
	}
	got, _ := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if string(got) != "old" {
		t.Fatalf("existing file was overwritten: %q", got)
	}
}

func TestExtractRefusesSuspiciousCommit(t *testing.T) {
	repo := makeRepo(t)
	for _, commit := range []string{"", "--output=/tmp/x", " "} {
		if err := Extract(repo, commit, filepath.Join(t.TempDir(), "snap")); err == nil {
			t.Fatalf("commit %q must be refused", commit)
		}
	}
}
