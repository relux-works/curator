package harnessparity

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/relux-works/curator/internal/testgitenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if err := captureCallerEnvironmentError(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	flag.Parse()
	if *nativeParentWorker {
		os.Exit(m.Run())
	}
	cleanup := testgitenv.Isolate()
	root, err := os.MkdirTemp("", "original-capture-system-")
	if err != nil {
		os.Exit(90)
	}
	path := filepath.Join(root, "system.json")
	if os.WriteFile(path, []byte(`{"schema_version":2,"locked":[]}`), 0600) != nil {
		os.Exit(91)
	}
	if os.Setenv("CURATOR_SYSTEM_CONFIG", path) != nil {
		os.Exit(92)
	}
	code := m.Run()
	_ = os.RemoveAll(root)
	cleanup()
	os.Exit(code)
}

// Both exact positive trees must pass and both actual one-byte plants must fail.
func TestOriginalComparatorControls(t *testing.T) {
	for _, c := range []fixtureCase{{Name: "claude_code-mcp", ID: "claude_code", Operation: "managed", State: "user-settings"}, {Name: "claude-launcher-typed-projection", ID: "claude_code", Operation: "launcher", State: "typed-projection"}} {
		t.Run(c.Name, func(t *testing.T) {
			original := runCase(t, c)
			if original.Refusal != "" {
				t.Fatal("positive control refused")
			}
			if err := compareResult(original, original); err != nil {
				t.Fatal(err)
			}
			altered := original
			nodes := append([]treeNode(nil), original.Rendered...)
			target := ".agent-context/mcp/claude_code.json"
			if c.Operation == "launcher" {
				nodes = append([]treeNode(nil), original.Tree...)
				target = "projection.json"
			}
			found := false
			want, got := "", ""
			for i, node := range nodes {
				if node.Path == target && node.Kind == "file" && len(node.Bytes) > 0 {
					nodes[i].Bytes = append([]byte(nil), node.Bytes...)
					nodes[i].Bytes[0] ^= 1
					want = node.SHA256
					got = hashBytes(nodes[i].Bytes)
					nodes[i].SHA256 = got
					found = true
					break
				}
			}
			if !found {
				t.Fatal("mutation not applied")
			}
			if c.Operation == "launcher" {
				altered.Tree = nodes
			} else {
				altered.Rendered = nodes
			}
			err := compareResult(original, altered)
			if err == nil || !strings.Contains(err.Error(), target) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), got) || want == got {
				t.Fatalf("wrong mutation verdict: %v", err)
			}
		})
	}
}

// Explicit root bindings project only exact path-boundary forms and retain all
// suffixes, entries, non-root bytes, and unrelated paths.
func TestWindowsRootParameter(t *testing.T) {
	valid := func(root string) fixtureRootEvidence {
		return fixtureRootEvidence{Root: root, Origin: "driver-created", RecordedBeforeRun: true}
	}
	rows := []struct {
		id      string
		binding fixtureRootEvidence
		windows bool
		input   []byte
		want    []byte
		refuse  bool
	}{
		{"windows-slash-order", valid(`C:\fixture\one`), true,
			[]byte(`{"claudeMdExcludes":["C:/fixture/one/operator/.claude/CLAUDE.md","C:/fixture/one/operator/.claude/rules/**"],"keep":17}`),
			[]byte(`{"claudeMdExcludes":["${FIXTURE_ROOT}/operator/.claude/CLAUDE.md","${FIXTURE_ROOT}/operator/.claude/rules/**"],"keep":17}`), false},
		{"windows-native", valid(`C:\fixture\one`), true, []byte(`C:\fixture\one\child`), []byte(`${FIXTURE_ROOT}\child`), false},
		{"windows-json-escaped", valid(`C:\fixture\one`), true, []byte(`"C:\\fixture\\one\\child"`), []byte(`"${FIXTURE_ROOT}\\child"`), false},
		{"linux-darwin", valid("/fixture/one"), false, []byte(`{"p":"/fixture/one/child","keep":[2,1]}`), []byte(`{"p":"${FIXTURE_ROOT}/child","keep":[2,1]}`), false},
		{"near-prefix-sibling", valid(`C:\fixture\one`), true, []byte(`C:/fixture/one-sibling/file`), []byte(`C:/fixture/one-sibling/file`), false},
		{"near-prefix-posix", valid("/fixture/one"), false, []byte(`/fixture/one-sibling/file`), []byte(`/fixture/one-sibling/file`), false},
		{"unrelated-path", valid(`C:\fixture\one`), true, []byte(`D:/elsewhere/file`), []byte(`D:/elsewhere/file`), false},
		{"canonical-unchanged", valid("/fixture/one"), false, []byte(`${FIXTURE_ROOT}/child`), []byte(`${FIXTURE_ROOT}/child`), false},
		{"changed-suffix", valid("/fixture/one"), false, []byte(`/fixture/one/different`), []byte(`${FIXTURE_ROOT}/different`), false},
		{"changed-nonroot", valid("/fixture/one"), false, []byte(`{"p":"/fixture/one/child","keep":18}`), []byte(`{"p":"${FIXTURE_ROOT}/child","keep":18}`), false},
		{"windows-equivalent-root", valid(`C:\fixture\two`), true,
			[]byte(`{"claudeMdExcludes":["C:/fixture/two/operator/.claude/CLAUDE.md","C:/fixture/two/operator/.claude/rules/**"],"keep":17}`),
			rowsExpectedWindowsOrder(), false},
		{"changed-exclusion-entry", valid(`C:\fixture\one`), true,
			[]byte(`{"claudeMdExcludes":["C:/fixture/one/operator/.claude/CLAUDE.md","C:/fixture/one/operator/.claude/changed/**"],"keep":17}`),
			[]byte(`{"claudeMdExcludes":["${FIXTURE_ROOT}/operator/.claude/CLAUDE.md","${FIXTURE_ROOT}/operator/.claude/changed/**"],"keep":17}`), false},
		{"quoted-root-value", valid("/fixture/one"), false, []byte("{\"p\":\"/fixture/one\",\"keep\":17}"), []byte("{\"p\":\"${FIXTURE_ROOT}\",\"keep\":17}"), false},
		{"colon-path-boundary", valid("/fixture/one"), false, []byte("path:/fixture/one/child"), []byte("path:${FIXTURE_ROOT}/child"), false},
		{"missing-binding", fixtureRootEvidence{}, false, []byte(`/fixture/one/child`), nil, true},
		{"ambiguous-binding", fixtureRootEvidence{Root: "/fixture/one", Origin: "driver-created", RecordedBeforeRun: true, Ambiguous: true}, false, []byte(`/fixture/one/child`), nil, true},
		{"environment-binding", fixtureRootEvidence{Root: "/fixture/one", Origin: "environment", RecordedBeforeRun: true}, false, []byte(`/fixture/one/child`), nil, true},
		{"output-binding", fixtureRootEvidence{Root: "/fixture/one", Origin: "captured-output", RecordedBeforeRun: true}, false, []byte(`/fixture/one/child`), nil, true},
		{"late-binding", fixtureRootEvidence{Root: "/fixture/one", Origin: "driver-created"}, false, []byte(`/fixture/one/child`), nil, true},
		{"non-absolute-binding", fixtureRootEvidence{Root: "fixture/one", Origin: "driver-created", RecordedBeforeRun: true}, false, []byte("/fixture/one/child"), nil, true},
		{"trailing-separator-binding", fixtureRootEvidence{Root: "/fixture/one/", Origin: "driver-created", RecordedBeforeRun: true}, false, []byte("/fixture/one/child"), nil, true},
	}

	results := make(map[string][]byte, len(rows))
	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			if !row.refuse {
				binding, err := bindFixtureRoot(row.binding)
				if err != nil || binding.windows != row.windows {
					t.Fatal("fixture-root form was not derived from its explicit root")
				}
			}
			got, err := projectFixtureRoot(row.binding, row.input)
			if row.refuse {
				if err == nil || !isFixtureRootBindingRefusal(err) || len(got) != 0 {
					t.Fatal("invalid fixture-root binding was not refused without output")
				}
				return
			}
			if err != nil || !bytes.Equal(row.want, got) {
				t.Fatal("fixture-root projection differed from the literal expectation")
			}
			results[row.id] = got
		})
	}
	if !bytes.Equal(results["windows-slash-order"], results["windows-equivalent-root"]) {
		t.Fatal("equivalent explicit Windows roots did not parameterize equally")
	}
	if bytes.Equal(results["linux-darwin"], results["changed-suffix"]) || bytes.Equal(results["linux-darwin"], results["changed-nonroot"]) {
		t.Fatal("changed suffix or non-root field was not observable")
	}
	if bytes.Equal(results["windows-slash-order"], results["changed-exclusion-entry"]) {
		t.Fatal("changed exclusion entry was not observable")
	}
}

func rowsExpectedWindowsOrder() []byte {
	return []byte(`{"claudeMdExcludes":["${FIXTURE_ROOT}/operator/.claude/CLAUDE.md","${FIXTURE_ROOT}/operator/.claude/rules/**"],"keep":17}`)
}
