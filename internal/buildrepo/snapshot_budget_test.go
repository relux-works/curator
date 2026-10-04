package buildrepo

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// These fixtures use real Git repositories and raw objects so aliases really
// share an OID, including directories that share a tree OID.
func makeBudgetFixture(t *testing.T, count, blobSize int, distinct, treeDAG bool) gitFixture {
	t.Helper()
	fixture := makeGitFixture(t, "sha1", false)
	gitDir := filepath.Join(fixture.work, ".git")
	var tree []byte
	for i := 0; i < count; i++ {
		content := bytes.Repeat([]byte{'x'}, blobSize)
		if distinct {
			copy(content, fmt.Sprintf("%04d", i))
		}
		blob := computeOID("sha1", "blob", content)
		// Git-created loose objects can be read-only (including the existing
		// empty blob); reusing an OID does not require rewriting its file.
		if _, err := os.Stat(filepath.Join(gitDir, "objects", blob[:2], blob[2:])); errors.Is(err, os.ErrNotExist) {
			blob = writeLooseObject(t, gitDir, "blob", content)
		} else if err != nil {
			t.Fatal(err)
		}
		target, mode := blob, "100644"
		if treeDAG {
			id, err := hex.DecodeString(blob)
			if err != nil {
				t.Fatal(err)
			}
			target = writeLooseObject(t, gitDir, "tree", append([]byte("100644 leaf\x00"), id...))
			mode = "40000"
		}
		id, err := hex.DecodeString(target)
		if err != nil {
			t.Fatal(err)
		}
		tree = append(tree, []byte(fmt.Sprintf("%s f%03d\x00", mode, i))...)
		tree = append(tree, id...)
	}
	treeOID := writeLooseObject(t, gitDir, "tree", tree)
	commit := []byte("tree " + treeOID + "\nauthor Fixture <fixture@example.test> 1 +0000\ncommitter Fixture <fixture@example.test> 1 +0000\n\nbudget\n")
	fixture.commit = writeLooseObject(t, gitDir, "commit", commit)
	// A detached HEAD avoids depending on the workstation's default branch.
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(fixture.commit+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestAdmitLocalExpandedSnapshotBudget(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		count                    int
		distinct, treeDAG, admit bool
	}{
		{"one-blob-control", 1, false, false, true},
		{"64-distinct-control", 64, true, false, false},
		{"64-aliases", 64, false, false, false},
		{"repeated-tree-DAG", 64, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := makeBudgetFixture(t, tc.count, 4096, tc.distinct, tc.treeDAG)
			snapshot, err := AdmitLocal(context.Background(), LocalRequest{
				Path: fixture.work, Tool: realGitTool(t), Limits: Limits{MaxExpandedBytes: 16384},
			})
			if tc.admit {
				if err != nil || snapshot == nil || len(snapshot.Files) != 1 {
					t.Fatalf("one-blob control: snapshot=%v err=%v", snapshot != nil, err)
				}
				want := append([]byte("curator-build-source-v1\x00F\x00\x00\x00\x00\x00\x00\x00\x04f000\x00\x00\x00\x00\x00\x00\x10\x00"), bytes.Repeat([]byte{'x'}, 4096)...)
				if !bytes.Equal(snapshot.CanonicalBytes, want) {
					t.Fatal("in-budget canonical bytes changed")
				}
				return
			}
			if ErrorCode(err) != CodeIncompleteSource || snapshot != nil {
				t.Fatalf("over-budget admission: snapshot=%v err=%v, want %s", snapshot != nil, err, CodeIncompleteSource)
			}
		})
	}
}

func TestAcquireNetworkExpandedSnapshotBudget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	for _, count := range []int{1, 64} {
		t.Run(fmt.Sprintf("aliases-%d", count), func(t *testing.T) {
			fixture := makeBudgetFixture(t, count, 4096, false, false)
			// Clone after installing the adversarial objects and detached HEAD.
			bare := filepath.Join(t.TempDir(), "remote.git")
			runTestGit(t, "", realGitPath(t), "clone", "--quiet", "--bare", "--", fixture.work, bare)
			tool, logPath := fakeHTTPGitTool(t, bare)
			source, err := ParseSource("https://fixture.test/repository.git")
			if err != nil {
				t.Fatal(err)
			}
			snapshot, err := AcquireNetwork(context.Background(), NetworkRequest{
				Source: source, Lock: LockedCommit{ObjectFormat: "sha1", Hex: fixture.commit},
				Tool: tool, Limits: Limits{MaxExpandedBytes: 16384},
			})
			if count == 1 {
				if err != nil || snapshot == nil {
					t.Fatalf("network control: snapshot=%v err=%v", snapshot != nil, err)
				}
			} else if ErrorCode(err) != CodeIncompleteSource || snapshot != nil {
				t.Fatalf("network aliases: snapshot=%v err=%v, want %s", snapshot != nil, err, CodeIncompleteSource)
			}
			log, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(log), "cat-file") {
				t.Fatal("network entry did not reach the shared raw-object proof")
			}
		})
	}
}

func TestAdmitLocalSnapshotBudgetBoundaries(t *testing.T) {
	fixture := makeBudgetFixture(t, 4, 4096, false, false)
	// Frozen framing: 24-byte header, then four (1 + 8 + 4 + 8 + 4096)
	// records. The content alone fits 16384, but the framing does not.
	const canonicalSize = 24 + 4*(1+8+4+8+4096)
	for _, tc := range []struct {
		name   string
		limits Limits
		admit  bool
	}{
		{"canonical-exact", Limits{MaxExpandedBytes: canonicalSize}, true},
		{"canonical-one-byte-over", Limits{MaxExpandedBytes: canonicalSize - 1}, false},
		{"content-fits-framing-does-not", Limits{MaxExpandedBytes: 16384}, false},
		{"files-exact", Limits{MaxFiles: 4}, true},
		{"files-one-over", Limits{MaxFiles: 3}, false},
		{"unique-objects-exact", Limits{MaxObjects: 3}, true},
		{"unique-objects-one-over", Limits{MaxObjects: 2}, false},
		{"object-bytes-one-over", Limits{MaxObjectBytes: 4095}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := AdmitLocal(context.Background(), LocalRequest{Path: fixture.work, Tool: realGitTool(t), Limits: tc.limits})
			if tc.admit {
				if err != nil || snapshot == nil || len(snapshot.CanonicalBytes) != canonicalSize {
					t.Fatalf("at-limit snapshot=%v err=%v", snapshot != nil, err)
				}
				if cap(snapshot.CanonicalBytes) != len(snapshot.CanonicalBytes) {
					t.Fatal("canonical allocation grew beyond its reserved size")
				}
			} else if ErrorCode(err) != CodeIncompleteSource || snapshot != nil {
				t.Fatalf("over-limit snapshot=%v err=%v, want %s", snapshot != nil, err, CodeIncompleteSource)
			}
		})
	}
}

func TestAdmitLocalExpandedTreeEntryBudget(t *testing.T) {
	fixture := makeBudgetFixture(t, 64, 0, false, true)
	// 64 directory aliases plus 64 leaf entries, though only two trees and
	// one empty blob are reachable. An entry budget is independent of bytes.
	for _, tc := range []struct {
		name    string
		entries int
		admit   bool
	}{
		{"exact", 128, true},
		{"one-over", 127, false},
		{"unique-only-would-admit", 4, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := AdmitLocal(context.Background(), LocalRequest{
				Path: fixture.work, Tool: realGitTool(t), Limits: Limits{MaxTreeEntries: tc.entries},
			})
			if tc.admit {
				if err != nil || snapshot == nil || len(snapshot.Files) != 64 {
					t.Fatalf("at-entry-limit snapshot=%v err=%v", snapshot != nil, err)
				}
			} else if ErrorCode(err) != CodeIncompleteSource || snapshot != nil {
				t.Fatalf("expanded tree entries: snapshot=%v err=%v, want %s", snapshot != nil, err, CodeIncompleteSource)
			}
		})
	}
}

func TestAdmitLocalEmptyTreeDAGEntryBudget(t *testing.T) {
	fixture := makeGitFixture(t, "sha1", false)
	gitDir := filepath.Join(fixture.work, ".git")
	empty := writeLooseObject(t, gitDir, "tree", nil)
	id, err := hex.DecodeString(empty)
	if err != nil {
		t.Fatal(err)
	}
	var tree []byte
	for i := 0; i < 64; i++ {
		tree = append(tree, []byte(fmt.Sprintf("40000 d%03d\x00", i))...)
		tree = append(tree, id...)
	}
	treeOID := writeLooseObject(t, gitDir, "tree", tree)
	commit := []byte("tree " + treeOID + "\nauthor Fixture <fixture@example.test> 1 +0000\ncommitter Fixture <fixture@example.test> 1 +0000\n\nempty DAG\n")
	commitOID := writeLooseObject(t, gitDir, "commit", commit)
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(commitOID+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, entries := range []int{64, 63} {
		t.Run(fmt.Sprintf("entries-%d", entries), func(t *testing.T) {
			snapshot, err := AdmitLocal(context.Background(), LocalRequest{
				Path: fixture.work, Tool: realGitTool(t),
				Limits: Limits{MaxTreeEntries: entries, MaxFiles: 1, MaxExpandedBytes: 4096},
			})
			if entries == 64 {
				if err != nil || snapshot == nil || len(snapshot.Files) != 0 || string(snapshot.CanonicalBytes) != "curator-build-source-v1\x00" {
					t.Fatalf("empty-directory control: snapshot=%v err=%v", snapshot != nil, err)
				}
			} else if ErrorCode(err) != CodeIncompleteSource || snapshot != nil {
				t.Fatalf("empty-directory fan-out: snapshot=%v err=%v, want %s", snapshot != nil, err, CodeIncompleteSource)
			}
		})
	}
}

func cachedBudgetTree(count int) (*objectReader, string) {
	content := bytes.Repeat([]byte{'x'}, 4096)
	blobOID := computeOID("sha1", "blob", content)
	var tree []byte
	id, _ := hex.DecodeString(blobOID)
	for i := 0; i < count; i++ {
		tree = append(tree, []byte(fmt.Sprintf("100644 f%03d\x00", i))...)
		tree = append(tree, id...)
	}
	treeOID := computeOID("sha1", "tree", tree)
	return &objectReader{format: "sha1", cache: map[string]rawObject{
		blobOID: {oid: blobOID, kind: "blob", data: content},
		treeOID: {oid: treeOID, kind: "tree", data: tree},
	}}, treeOID
}

// Entry-point regressions above prove the production wiring. Inspect the
// same walker here to prove refusal occurs before copying/appending a file,
// which is otherwise invisible once AdmitLocal discards the rejected result.
func TestSnapshotBudgetRefusesBeforeEmission(t *testing.T) {
	for _, tc := range []struct {
		name   string
		limits Limits
	}{
		{"expanded-bytes", Limits{MaxExpandedBytes: 16384}},
		{"canonical-bytes", Limits{MaxExpandedBytes: 24 + 4*(1+8+4+8+4096) - 1}},
		{"files", Limits{MaxFiles: 3}},
		{"entries", Limits{MaxTreeEntries: 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader, tree := cachedBudgetTree(64)
			budget := snapshotBudget{canonical: int64(len(snapshotHeader))}
			var files []File
			seen := map[string]string{}
			err := walkTree(context.Background(), reader, tree, "", 0, normalizedLimits(tc.limits), &budget, seen, &files)
			if ErrorCode(err) != CodeIncompleteSource {
				t.Fatalf("walk error=%v, want %s", err, CodeIncompleteSource)
			}
			if len(files) != 3 || budget.files != 3 || budget.expanded != 3*4096 || budget.canonical != 24+3*(1+8+4+8+4096) {
				t.Fatalf("refused file was emitted or charged: len=%d budget=%+v", len(files), budget)
			}
			if tc.name == "entries" && len(seen) != 3 {
				t.Fatalf("over-budget entry was retained: %d paths", len(seen))
			}
		})
	}
}

type cancelDuringWalkContext struct {
	context.Context
	cancel context.CancelFunc
	checks int
}

func (c *cancelDuringWalkContext) Err() error {
	c.checks++
	if c.checks == 4 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestSnapshotBudgetCachedWalkCancellation(t *testing.T) {
	reader, tree := cachedBudgetTree(64)
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &cancelDuringWalkContext{Context: parent, cancel: cancel}
	budget := snapshotBudget{canonical: int64(len(snapshotHeader))}
	var files []File
	err := walkTree(ctx, reader, tree, "", 0, DefaultLimits(), &budget, map[string]string{}, &files)
	if ErrorCode(err) != CodeIncompleteSource || !errors.Is(err, context.Canceled) || len(files) != 2 {
		t.Fatalf("cached walk ignored cancellation: files=%d checks=%d err=%v", len(files), ctx.checks, err)
	}
}
