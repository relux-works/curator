package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func fixtureTests() []string {
	var names []string
	for i := 0; i < 80; i++ {
		names = append(names, fmt.Sprintf("TestCase%03d", i))
	}
	names = append(names, ownershipControls)
	return names
}

// Exact selectors partition every discovered name once, including future tests
// and names that would otherwise have regexp meaning. Duplicate input refuses.
func TestPartitionCoverage(t *testing.T) {
	tests := append(fixtureTests(), "TestFuture", "ExampleOutput", "FuzzInput", "TestLiteral.+")
	shards, err := partition(tests)
	if err != nil {
		t.Fatal(err)
	}
	if len(shards[hashShardCount]) != 1 || shards[hashShardCount][0] != ownershipControls {
		t.Fatal("ownership controls must run alone")
	}
	// Removing or renaming the dedicated suite must fail instead of quietly
	// retiring its runner; the full-list and platform gates still check cases.
	ordinary := fixtureTests()
	if _, err := partition(ordinary[:len(ordinary)-1]); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("empty shard %d", hashShardCount)) {
		t.Fatalf("missing ownership suite: %v", err)
	}
	for _, name := range tests {
		matches := 0
		for _, s := range shards {
			if regexp.MustCompile(selector(s)).MatchString(name) {
				matches++
			}
		}
		if matches != 1 {
			t.Fatalf("%s matches %d shards", name, matches)
		}
	}
	if regexp.MustCompile(selector([]string{"TestLiteral.+"})).MatchString("TestLiteralX") {
		t.Fatal("selector is not literal")
	}
	for _, bad := range [][]string{nil, {"TestOne", "TestOne"}, {"TestParent/subtest"}} {
		if _, err := partition(bad); err == nil {
			t.Fatal("invalid inventory admitted")
		}
	}
}

func fixtureStream(tests []string) []byte {
	var out strings.Builder
	for _, name := range tests {
		fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Package\":%q,\"Test\":%q}\n", pkg, name)
	}
	fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Package\":%q,\"Elapsed\":42}\n", pkg)
	return []byte(out.String())
}

// Completed events and a passing package are necessary: a selector, startup,
// package-only pass, unrelated failure or exhausted margin cannot prove coverage.
func TestStreamControls(t *testing.T) {
	tests := []string{"TestOne", "TestTwo"}
	good := string(fixtureStream(tests))
	if _, err := validateStream([]byte(good), tests); err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct{ data, want string }{
		"missing":         {string(fixtureStream(tests[:1])), "missing completed test TestTwo"},
		"duplicate":       {strings.Split(good, "\n")[0] + "\n" + good, "duplicate test TestOne"},
		"unselected":      {strings.ReplaceAll(good, "TestTwo", "TestOther"), "unexpected or duplicate test TestOther"},
		"no-package":      {strings.Join(strings.Split(good, "\n")[:2], "\n"), "want one passing package"},
		"failed":          {strings.Replace(good, "pass", "fail", 1), "failed test/package TestOne"},
		"invalid-json":    {good + "broken", "invalid test JSON"},
		"slow":            {strings.Replace(good, ":42", ":1800", 1), "outside (0,1800)s margin"},
		"wrong-package":   {strings.ReplaceAll(good, pkg, "other"), "unexpected package"},
		"twice-package":   {good + strings.Split(good, "\n")[2], "want one passing package"},
		"package-skipped": {strings.Replace(good, `"Action":"pass","Package":"`+pkg+`","Elapsed"`, `"Action":"skip","Package":"`+pkg+`","Elapsed"`, 1), "want one passing package"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateStream([]byte(tc.data), tests); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}

// Build-output uses ImportPath rather than Package. Only the selected test
// package's compiler/linker output is admitted; it is never test completion.
func TestBuildEventControls(t *testing.T) {
	good := string(fixtureStream([]string{"TestOne"}))
	build := fmt.Sprintf(`{"Action":"build-output","ImportPath":%q,"Output":"linker warning\n"}`+"\n", pkg)
	if _, err := validateStream([]byte(build+good), []string{"TestOne"}); err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct{ data, want string }{
		"missing-import":   {strings.Replace(build, `"ImportPath":`+fmt.Sprintf("%q", pkg), `"ImportPath":""`, 1) + good, "invalid build event"},
		"mixed-package":    {strings.TrimSuffix(build, "}\n") + `,"Package":` + fmt.Sprintf("%q", pkg) + "}\n" + good, "invalid build event"},
		"test-with-import": {strings.TrimSuffix(strings.Split(good, "\n")[0], "}") + `,"ImportPath":` + fmt.Sprintf("%q", pkg) + "}\n" + good, "unexpected build import on test event"},
		"foreign-import":   {strings.Replace(build, pkg, "foreign/package", 1) + good, "unexpected build import"},
		"build-fail":       {strings.Replace(build, "build-output", "build-fail", 1) + good, "build failure"},
		"build-test":       {strings.TrimSuffix(build, "}\n") + `,"Test":"TestOne"}` + "\n" + good, "invalid build event"},
		"missing-test":     {build + string(fixtureStream(nil)), "missing completed test"},
		"unknown-action":   {strings.Replace(build, "build-output", "build-skip", 1) + good, "unsupported event action"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := validateStream([]byte(tc.data), []string{"TestOne"}); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}

func writeEvidence(t *testing.T, dir string, tests []string) {
	t.Helper()
	shards, err := partition(tests)
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range shards {
		path := filepath.Join(dir, fmt.Sprint(i))
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
		m := manifest{Commit: "head", GOOS: "linux", Shard: i, Tests: tests, Selector: selector(s)}
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "manifest.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "go-test.json"), fixtureStream(s), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// The union fails closed on missing/duplicate/stale/wrong-platform artifacts,
// inventory or selector drift, nonzero test exit, or lost actual test events.
// A refused union must not publish a partial merged stream.
func TestMergeControls(t *testing.T) {
	tests := fixtureTests()
	for _, plant := range []string{"valid", "missing-shard", "duplicate-shard", "stale", "platform", "inventory", "selector", "exit", "missing-test"} {
		t.Run(plant, func(t *testing.T) {
			dir := t.TempDir()
			writeEvidence(t, dir, tests)
			path := filepath.Join(dir, "0", "manifest.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var m manifest
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatal(err)
			}
			switch plant {
			case "missing-shard":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			case "duplicate-shard":
				m.Shard = 1
				shards, _ := partition(tests)
				m.Selector = selector(shards[1])
				if err := os.WriteFile(filepath.Join(dir, "0", "go-test.json"), fixtureStream(shards[1]), 0644); err != nil {
					t.Fatal(err)
				}
			case "stale":
				m.Commit = "old"
			case "platform":
				m.GOOS = "darwin"
			case "inventory":
				m.Tests = append(m.Tests, "TestNew")
			case "selector":
				m.Selector = "^TestNothing$"
			case "exit":
				m.ExitCode = 1
			case "missing-test":
				shards, _ := partition(tests)
				if err := os.WriteFile(filepath.Join(dir, "0", "go-test.json"), fixtureStream(shards[0][1:]), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if plant != "missing-shard" {
				data, _ = json.Marshal(m)
				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			out := filepath.Join(t.TempDir(), "merged.json")
			err = mergeEvidence(dir, out, tests, "head", "linux")
			if plant == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			want := "source/platform/inventory/selector/exit mismatch"
			switch plant {
			case "missing-shard":
				want = fmt.Sprintf("want %d shard manifests, got %d", shardCount, shardCount-1)
			case "duplicate-shard":
				want = "invalid or duplicate shard"
			case "missing-test":
				want = "missing completed test"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v, want %s", err, want)
			}
			if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
				t.Fatalf("refused union published output: %v", statErr)
			}
		})
	}
}

// The workflow cannot silently run fewer shards when the tool's count changes.
func TestWorkflowShardCount(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	var wf struct {
		Jobs map[string]struct {
			Strategy struct {
				Matrix struct {
					Shard []int `yaml:"shard"`
				} `yaml:"matrix"`
			} `yaml:"strategy"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatal(err)
	}
	indices := wf.Jobs["envprofile-race"].Strategy.Matrix.Shard
	if len(indices) != shardCount {
		t.Fatalf("workflow has %d shards, tool has %d", len(indices), shardCount)
	}
	for i, index := range indices {
		if index != i {
			t.Fatalf("workflow shard %d = %d", i, index)
		}
	}
}

// Actions/local explicit pins must be full revisions, never branch names.
func TestSourceCommitPin(t *testing.T) {
	good := strings.Repeat("a", 40)
	t.Setenv("GITHUB_SHA", good)
	if got, err := sourceCommit(); err != nil || got != good {
		t.Fatalf("pin: %s, %v", got, err)
	}
	t.Setenv("GITHUB_SHA", "main")
	if _, err := sourceCommit(); err == nil || !strings.Contains(err.Error(), "full commit revision") {
		t.Fatalf("invalid source pin admitted: %v", err)
	}
}
