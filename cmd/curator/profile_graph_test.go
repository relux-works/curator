package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/profilegraph"
	"github.com/relux-works/curator/internal/profilegraph/graphtest"
)

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
