package projectcontext

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: every bounded object reader admits its exact byte limit and refuses
// the next byte, retaining at most the limit and reporting an object refusal.
// These portable copy controls exercise the actual writer, not a Git fixture.
func TestGitObjectOutputBounds(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int
	}{
		{"ls-tree", maxGitTreeBytes}, {"cat-file", MaxFileBytes}, {"tag-line", MaxRecord},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, extra := range []int{0, 1} {
				out := &boundedOutput{max: tc.limit}
				_, err := io.Copy(out, io.LimitReader(strings.NewReader(strings.Repeat("x", tc.limit+extra)), int64(tc.limit+extra)))
				if extra == 0 {
					if err != nil || out.exceeded || out.buffer.Len() != tc.limit {
						t.Fatal("exact limit refused", err)
					}
				} else if err == nil || !out.exceeded || out.buffer.Len() > tc.limit {
					t.Fatal("overflow admitted", err)
				}
			}
			site := tc.name
			if site == "tag-line" {
				site = "tag"
			}
			err := (&localGitError{class: "output bound exceeded", site: site}).Error()
			if !strings.HasPrefix(err, "project_context_invalid: local Git object read failed at "+site+":") || strings.Contains(err, "identity") {
				t.Fatal("object failure misclassified", err)
			}
		})
	}
}

func gitFixtureInput(t *testing.T, root string, input []byte, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "GIT_") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
	cmd.Stdin = bytes.NewReader(input)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fixture Git input: %v: %s", err, b)
	}
	return strings.TrimSpace(string(b))
}

// Claim: a real MaxFiles tree with 240-byte portable paths is admitted,
// including entries outside a directory selection; a next file is refused.
// Reads leave Git's index unchanged. Shared empty blobs keep total bytes in cap.
func TestGitTreeInventoryCaps(t *testing.T) {
	root := t.TempDir()
	initRepo(t, root)
	blob := gitFixtureInput(t, root, nil, "hash-object", "-w", "--stdin")
	gitTest(t, root, "read-tree", "--empty")
	var index bytes.Buffer
	for i := 0; i < MaxFiles; i++ {
		prefix := "other/"
		if i == 0 {
			prefix = "selected/"
		}
		name := fmt.Sprintf("%sf%05d-", prefix, i)
		name += strings.Repeat("x", maxPathBytes-len(name))
		fmt.Fprintf(&index, "100644 %s\t%s\n", blob, name)
	}
	gitFixtureInput(t, root, index.Bytes(), "update-index", "--add", "--index-info")
	commit := gitTest(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit-tree", gitTest(t, root, "write-tree"), "-m", "inventory caps")
	before, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := captureCommit(root, commit, "selected")
	if err != nil || len(c.Bytes) != 1 {
		t.Fatalf("admitted capped tree refused: files=%d: %v", len(c.Bytes), err)
	}
	listing, err := gitBytes(root, maxGitTreeBytes, "ls-tree", "-r", "-l", "-z", commit)
	if err != nil || len(listing) <= MaxRecord {
		t.Fatalf("control did not cross old listing limit: %d: %v", len(listing), err)
	}
	after, err := os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("capture mutated index", err)
	}
	gitFixtureInput(t, root, []byte("100644 "+blob+"\textra\n"), "update-index", "--add", "--index-info")
	commit = gitTest(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit-tree", gitTest(t, root, "write-tree"), "-m", "one extra file")
	before, err = os.ReadFile(filepath.Join(root, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	c, err = captureCommit(root, commit, "selected")
	if !isCode(err, "project_context_invalid") || !strings.Contains(err.Error(), "file bound exceeded") || len(c.Bytes) != 0 {
		t.Fatal("extra file admitted or wrong refusal", err)
	}
	after, e := os.ReadFile(filepath.Join(root, ".git", "index"))
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("refusal mutated index", e)
	}
}

// Claim: cat-file admits the capped blob and rejects one extra byte before
// exposing partial output. Both observations leave the immutable object intact.
func TestGitBlobOutputCap(t *testing.T) {
	root := t.TempDir()
	initRepo(t, root)
	for _, extra := range []int{0, 1} {
		data := bytes.Repeat([]byte{'x'}, MaxFileBytes+extra)
		oid := gitFixtureInput(t, root, data, "hash-object", "-w", "--stdin")
		got, err := gitBytes(root, MaxFileBytes, "cat-file", "blob", oid)
		if extra == 0 {
			if err != nil || !bytes.Equal(got, data) {
				t.Fatal("capped blob refused", err)
			}
		} else if !isCode(err, "project_context_invalid") || !strings.Contains(err.Error(), "at cat-file: output bound exceeded") || got != nil {
			t.Fatal("oversized blob admitted or wrong refusal", err)
		}
		if size := gitTest(t, root, "cat-file", "-s", oid); size != fmt.Sprint(len(data)) {
			t.Fatal("read mutated object", size)
		}
	}
}

// Claim: inert tag names totaling more than one record are consumed without an
// aggregate cap; the valid candidate survives. A single over-bound line refuses
// without retaining an unbounded name. Candidate resolution is tested elsewhere.
func TestGitTagEnumerationBounds(t *testing.T) {
	root := t.TempDir()
	initRepo(t, root)
	head := gitTest(t, root, "rev-parse", "HEAD")
	var packed strings.Builder
	packed.WriteString("# pack-refs with: peeled fully-peeled sorted\n")
	for i := 0; i < 90000; i++ {
		fmt.Fprintf(&packed, "%s refs/tags/inert-%08d\n", head, i)
	}
	fmt.Fprintf(&packed, "%s refs/tags/v1.2.3\n", head)
	write(t, filepath.Join(root, ".git", "packed-refs"), packed.String())
	before := packed.String()
	tags := &versionTagOutput{}
	if err := gitRead(root, tags, func() bool { return tags.line.exceeded }, "tag", "--list"); err != nil || len(tags.tags) != 1 || tags.tags[0].Tag != "v1.2.3" {
		t.Fatal("large tag listing refused", err)
	}
	after, err := os.ReadFile(filepath.Join(root, ".git", "packed-refs"))
	if err != nil || string(after) != before {
		t.Fatal("listing mutated refs", err)
	}
	positive := &versionTagOutput{}
	if _, err = positive.Write([]byte(strings.Repeat("x", MaxRecord) + "\n")); err != nil || positive.line.exceeded {
		t.Fatal("exact line refused", err)
	}
	negative := &versionTagOutput{}
	if _, err = negative.Write([]byte(strings.Repeat("x", MaxRecord+1) + "\n")); err == nil || !negative.line.exceeded || len(negative.tags) != 0 {
		t.Fatal("line overflow admitted", err)
	}
}
