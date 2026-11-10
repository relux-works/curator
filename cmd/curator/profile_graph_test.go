package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/profilegraph"
	"github.com/relux-works/curator/internal/profilegraph/graphtest"
)

// Both builders must establish the real lock boundary, including when a
// test-owned lock already has broad mutation permissions. The native Windows
// helper plants a foreign-write DACL; Unix plants group/other write bits.
// After construction, the command must still refuse that same violation
// atomically, preserve its bytes/metadata, and leave its permissions unsafe.
func TestProfileGraphFixtureLockBoundary(t *testing.T) {
	for _, name := range []string{"small", "dag-overlays"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("CURATOR_SYSTEM_CONFIG", filepath.Join(home, "absent-policy.json"))
			lockPath := filepath.Join(home, "profiles/sample/lock.json")
			if err := graphtest.Write(home, "profiles/sample/lock.json", []byte("{}")); err != nil {
				t.Fatal(err)
			}
			boundary := func() error {
				return pathboundary.ValidateLeafWithOwner(home, lockPath, pathboundary.DefaultOwnerLookup())
			}
			plant := func() {
				breakCacheProtection(t, lockPath)
				var failure *pathboundary.Failure
				if err := boundary(); !errors.As(err, &failure) || failure.Check != pathboundary.CheckPermissions {
					t.Fatalf("foreign-write control did not fail permissions: %v", err)
				}
			}
			plant()
			var err error
			if name == "small" {
				_, err = graphtest.Small(home, "../../internal/profilegraph/testdata/small")
			} else {
				_, err = graphtest.DAG(home)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := boundary(); err != nil {
				t.Fatalf("fixture lock boundary: %v", err)
			}
			for _, valid := range []bool{true, false} {
				if !valid {
					plant()
				}
				before := graphCommandSnapshot(t, home)
				var stdout, stderr bytes.Buffer
				code := run([]string{"profile", "graph", "sample", "--json"}, fileConfigSource(filepath.Join(home, "config.json")), &stdout, &stderr)
				if valid && (code != exitOK || stdout.Len() == 0) {
					t.Fatalf("private lock: code %d: %s", code, stderr.String())
				}
				if !valid && (code != exitFail || stdout.Len() != 0 || !strings.Contains(stderr.String(), "graph_invalid_input")) {
					t.Fatalf("writable lock: code %d: %s", code, stderr.String())
				}
				if graphCommandSnapshot(t, home) != before {
					t.Fatal("command changed fixture bytes or metadata")
				}
				if !valid && boundary() == nil {
					t.Fatal("command repaired the refused lock permissions")
				}
			}
		})
	}
}

// Windows absolute/UNC shapes and backslash traversal refuse before payload
// admission on every host; an ordinary relative file remains readable. This
// exercises lexical admission, not native drive resolution or reparse races.
func TestProfileGraphReaderPathRefusals(t *testing.T) {
	root := t.TempDir()
	if err := graphtest.Write(root, "context/base.md", []byte("allowed\n")); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{`C:\outside\base.md`, `\\server\share\base.md`, "//server/share/base.md", `..\outside.md`, `context\..\..\outside.md`} {
		t.Run(rel, func(t *testing.T) {
			r := profilegraph.NewReader(profilegraph.V1Limits)
			opens := 0
			r.OpenHook = func(_, _ string) { opens++ }
			if raw, err := r.Read(root, "source", "context/base.md"); err != nil || string(raw) != "allowed\n" || opens != 1 {
				t.Fatalf("relative positive control: %q %v opens=%d", raw, err, opens)
			}
			before := graphCommandSnapshot(t, root)
			raw, err := r.Read(root, "source", rel)
			var refusal *profilegraph.Error
			if raw != nil || !errors.As(err, &refusal) || refusal.Class != "graph_invalid_input" || opens != 1 {
				t.Fatalf("path refusal: %q %v opens=%d", raw, err, opens)
			}
			if graphCommandSnapshot(t, root) != before {
				t.Fatal("path refusal changed the fixture")
			}
		})
	}
}

// Both goldens are produced by real CLI dispatch over immutable snapshots and
// local Git objects. This also proves that shared dependencies remain one node
// while overlay occurrences with different directory/weight coordinates survive.
func TestProfileGraphCommandGoldens(t *testing.T) {
	for _, name := range []string{"small", "dag-overlays"} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			var err error
			if name == "small" {
				_, err = graphtest.Small(home, "../../internal/profilegraph/testdata/small")
			} else {
				_, err = graphtest.DAG(home)
			}
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			before := graphCommandSnapshot(t, home)
			code := run([]string{"profile", "graph", "sample", "--json"}, fileConfigSource(filepath.Join(home, "config.json")), &stdout, &stderr)
			if code != exitOK {
				t.Fatalf("code %d: %s", code, stderr.String())
			}
			if graphCommandSnapshot(t, home) != before {
				t.Fatal("command wrote managed files, object cache, refs or config")
			}
			var graph profilegraph.Graph
			if err := json.Unmarshal(stdout.Bytes(), &graph); err != nil {
				t.Fatal(err)
			}
			for _, d := range graph.Documents {
				offset := 0
				for _, entry := range d.Coverage {
					if entry.Range.Start != offset || entry.Range.End <= offset {
						t.Fatal("coverage gap or overlap")
					}
					offset = entry.Range.End
				}
				if offset != d.Bytes {
					t.Fatal("coverage does not reconcile")
				}
			}
			if name == "dag-overlays" {
				overlays := map[string]bool{}
				leaf := ""
				incoming := 0
				for _, n := range graph.Nodes {
					if n.Kind == "overlay" {
						overlays[n.Source.Directory] = true
					}
					if n.Label == "leaf" {
						if leaf != "" {
							t.Fatal("duplicated leaf package")
						}
						leaf = n.ID
					}
				}
				for _, e := range graph.Edges {
					if e.To == leaf && strings.HasPrefix(e.Kind, "requires-") {
						incoming++
					}
				}
				if len(overlays) != 2 || !overlays["overlay-a"] || !overlays["overlay-b"] || incoming != 2 {
					t.Fatalf("overlays=%v leaf incoming=%d", overlays, incoming)
				}
			}
			golden := filepath.Join("../../internal/profilegraph/testdata", name, "graph.json")
			if os.Getenv("UPDATE_PROFILE_GRAPH_GOLDEN") == "1" {
				if err := os.MkdirAll(filepath.Dir(golden), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, stdout.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
			}
			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(stdout.Bytes(), expected) {
				t.Fatalf("%s command golden differs", name)
			}
			if bytes.Contains(stdout.Bytes(), []byte(home)) {
				t.Fatal("absolute home leaked")
			}
		})
	}
}

func graphCommandSnapshot(t *testing.T, root string) string {
	t.Helper()
	var out strings.Builder
	err := filepath.WalkDir(root, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		fmt.Fprintf(&out, "%s %s %d %d\n", rel, info.Mode(), info.Size(), info.ModTime().UnixNano())
		if info.Mode().IsRegular() {
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(&out, "%x\n", sha256.Sum256(raw))
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}
			out.WriteString(target)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// Preview works without bootstrap; unknown environment and malformed arguments
// refuse atomically with the specified 1/2 exit codes. A denied declaration is
// never echoed in a diagnostic or emitted as a partial graph.
func TestProfileGraphCommandProtocol(t *testing.T) {
	source := t.TempDir()
	if err := graphtest.Write(source, "agent-context.json", []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"base.md"}]}}`)); err != nil {
		t.Fatal(err)
	}
	if err := graphtest.Write(source, "context/base.md", []byte("# Baseline\n")); err != nil {
		t.Fatal(err)
	}
	config := fileConfigSource(filepath.Join(t.TempDir(), "absent.json"))
	for _, tc := range []struct {
		args       []string
		code       int
		diagnostic string
	}{
		{[]string{"sample", "--source", source, "--env", "codex", "--json"}, 0, "graph_load_unknown"},
		{[]string{"sample", "--source", source, "--env", "unregistered", "--json"}, 1, "graph_invalid_input"},
		{[]string{"sample", "--bogus"}, 2, "profile graph"},
		{nil, 2, "profile graph"},
		{[]string{"sample", "--source", filepath.Join(source, "absent"), "--json"}, 1, "graph_missing_input"},
	} {
		var stdout, stderr bytes.Buffer
		code := run(append([]string{"profile", "graph"}, tc.args...), config, &stdout, &stderr)
		if code != tc.code || !strings.Contains(stderr.String(), tc.diagnostic) {
			t.Fatalf("%v: %d %s", tc.args, code, stderr.String())
		}
		if code != 0 && stdout.Len() != 0 {
			t.Fatal("partial output on failure")
		}
		if code == 0 {
			var graph profilegraph.Graph
			if json.Unmarshal(stdout.Bytes(), &graph) != nil || graph.Scope.Mode != "source" || len(graph.Scope.Environments) != 1 || graph.Scope.Environments[0] != "codex_cli" {
				t.Fatal("wrong preview/alias projection")
			}
		}
	}
	if err := graphtest.Write(source, "agent-context.json", []byte(`{"schema_version":1,"name":"sample","version":"1.0.0","context":{"modules":[{"path":"auth.json"}]}}`)); err != nil {
		t.Fatal(err)
	}
	if err := graphtest.Write(source, "context/auth.json", []byte("secret sentinel\n")); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"profile", "graph", "sample", "--source", source, "--json"}, config, &stdout, &stderr)
	if code != 1 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "graph_path_denied") || strings.Contains(stderr.String(), "auth.json") || strings.Contains(stderr.String(), "sentinel") {
		t.Fatalf("code=%d out=%s err=%s", code, stdout.String(), stderr.String())
	}
}
