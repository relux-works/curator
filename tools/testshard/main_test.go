package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Greedy bins keep every original top-level name exactly once; a heavy case
// cannot be obscured by distributing only whole packages.
func TestPartitionPreservesInventoryAndBalancesKnownCosts(t *testing.T) {
	names := []string{"TestA", "TestB", "TestC", "TestD", "TestE", "ExampleHello", "FuzzInput"}
	weights := map[string]float64{"TestA": 100, "TestB": 90, "TestC": 80, "TestD": 70}
	got := partition(names, weights)
	seen := map[string]int{}
	for _, bin := range got {
		for _, name := range bin {
			seen[name]++
		}
	}
	for _, name := range names {
		if seen[name] != 1 {
			t.Fatalf("%s assigned %d times", name, seen[name])
		}
	}
	if len(seen) != len(names) {
		t.Fatalf("unexpected names %v", seen)
	}
	for i, name := range names[:4] {
		if !slices.Contains(got[i], name) {
			t.Fatalf("heavy case %s not in bin %d: %v", name, i, got)
		}
	}
	reversed := append([]string(nil), names...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	if !reflect.DeepEqual(got, partition(reversed, weights)) {
		t.Fatal("partition depends on enumeration order")
	}
}

func fixtureShards(t *testing.T) ([]manifest, [][]byte) {
	t.Helper()
	pkg := "github.com/relux-works/curator/cmd/curator"
	var manifests []manifest
	var streams [][]byte
	names := []string{"TestA", "TestB", "TestC", "TestD"}
	for i, name := range names {
		manifests = append(manifests, manifest{Index: i, Count: 4, Packages: []selection{{Package: pkg, Tests: names, Selected: []string{name}}}})
		events := []event{{Action: "start", Package: pkg}, {Action: "run", Package: pkg, Test: name}, {Action: "pass", Package: pkg, Test: name}, {Action: "pass", Package: pkg}}
		var stream []byte
		for _, e := range events {
			data, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			stream = append(stream, append(data, '\n')...)
		}
		streams = append(streams, stream)
	}
	return manifests, streams
}

// The production merge verifier accepts exhaustive passing shards and refuses
// each missing, failed, duplicate, truncated or unselected-case class by name.
func TestVerifyRequiresEveryShardAndEverySelectedTest(t *testing.T) {
	cases := []struct {
		name   string
		mutate func([]manifest, [][]byte) ([]manifest, [][]byte)
		want   string
	}{
		{"valid", func(m []manifest, s [][]byte) ([]manifest, [][]byte) { return m, s }, ""},
		{"missing-shard", func(m []manifest, s [][]byte) ([]manifest, [][]byte) { return m[:3], s[:3] }, "missing shard"},
		{"duplicate-shard", func(m []manifest, s [][]byte) ([]manifest, [][]byte) { m[3].Index = 2; return m, s }, "invalid or duplicate shard"},
		{"wrong-count", func(m []manifest, s [][]byte) ([]manifest, [][]byte) { m[0].Count = 3; return m, s }, "invalid or duplicate shard"},
		{"inventory-drift", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			m[1].Packages[0].Tests = []string{"TestB"}
			return m, s
		}, "shard inventories disagree"},
		{"missing-test-verdict", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			s[1] = []byte("{\"Action\":\"pass\",\"Package\":\"github.com/relux-works/curator/cmd/curator\"}\n")
			return m, s
		}, "missing test verdict"},
		{"failed-test", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			s[1] = []byte(strings.Replace(string(s[1]), `"Action":"pass"`, `"Action":"fail"`, 1))
			return m, s
		}, "failed event"},
		{"malformed-stream", func(m []manifest, s [][]byte) ([]manifest, [][]byte) { s[2] = append(s[2], '{'); return m, s }, "unexpected EOF"},
		{"missing-package-verdict", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			lines := strings.Split(string(s[2]), "\n")
			s[2] = []byte(strings.Join(lines[:3], "\n") + "\n")
			return m, s
		}, "missing package verdict"},
		{"unselected-test", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			m[3].Packages[0].Selected = nil
			s[3] = nil
			return m, s
		}, "unselected test"},
		{"duplicate-cross-shard-test", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			m[3].Packages[0].Selected = []string{"TestC"}
			s[3] = []byte(strings.ReplaceAll(string(s[3]), "TestD", "TestC"))
			return m, s
		}, "duplicate cross-shard test"},
		{"duplicate-selection", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			m[0].Packages[0].Selected = []string{"TestA", "TestA"}
			return m, s
		}, "invalid selection"},
		{"invented-test", func(m []manifest, s [][]byte) ([]manifest, [][]byte) {
			m[0].Packages[0].Selected = []string{"TestInvented"}
			return m, s
		}, "invalid selection"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, s := fixtureShards(t)
			m, s = tc.mutate(m, s)
			err := verify(m, s)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %s", err, tc.want)
			}
		})
	}
}

// go test -list failures/truncation must not become an empty admitted inventory.
func TestParseInventoryRequiresPackageVerdicts(t *testing.T) {
	pkg := "example.test/p"
	for _, tc := range []struct{ stream, want string }{
		{`{"Action":"output","Package":"example.test/p","Output":"TestOriginal\n"}` + "\n" + `{"Action":"pass","Package":"example.test/p"}`, ""},
		{`{"Action":"output","Package":"example.test/p","Output":"TestOriginal\n"}`, "incomplete inventory"},
		{`{"Action":"fail","Package":"example.test/p"}`, "failed inventory"},
		{`{"Action":"pass","Package":"example.test/other"}`, "unexpected inventory package"},
	} {
		rows, err := parseInventory([]byte(tc.stream), []string{pkg})
		if tc.want == "" {
			if err != nil || len(rows) != 1 || !reflect.DeepEqual(rows[0].Tests, []string{"TestOriginal"}) {
				t.Fatalf("%v %v", rows, err)
			}
		} else if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%v, want %s", err, tc.want)
		}
	}
}

// Drive merge through the actual CLI: a missing fourth artifact must give a
// nonzero verdict and must not publish a combined stream.
func TestMergeCLIRefusesMissingArtifact(t *testing.T) {
	if os.Getenv("TEST_SHARD_CHILD") == "1" {
		os.Args = []string{"testshard", "merge", os.Getenv("TEST_SHARD_ROOT")}
		main()
		return
	}
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestMergeCLIRefusesMissingArtifact$")
	cmd.Env = append(os.Environ(), "TEST_SHARD_CHILD=1", "TEST_SHARD_ROOT="+root)
	out, err := cmd.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "shard-manifest.json") {
		t.Fatalf("child=%v %s", err, out)
	}
	if _, err := os.Stat(filepath.Join(root, "go-test.json")); !os.IsNotExist(err) {
		t.Fatalf("published stream after refusal: %v", err)
	}
}

func TestShardCLIHelperProcess(t *testing.T) {
	if os.Getenv("TEST_SHARD_HELPER") != "1" {
		return
	}
	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("TEST_SHARD_ARGS")), &args); err != nil {
		t.Fatal(err)
	}
	os.Args = append([]string{"testshard"}, args...)
	main()
	os.Exit(0)
}

func shardCLI(t *testing.T, cwd string, args ...string) ([]byte, error) {
	t.Helper()
	data, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestShardCLIHelperProcess$")
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "TEST_SHARD_HELPER=1", "TEST_SHARD_ARGS="+string(data))
	return cmd.CombinedOutput()
}

// Real go test subprocesses execute all cases (including fuzz seeds/examples),
// then the production CLI merges their actual JSON. Removing a passing case's
// verdict afterwards must refuse without replacing the previous output.
func TestShardCLIExecutesEveryCaseBeforeCombinedAcceptance(t *testing.T) {
	lockShardGoTools(t)
	root := t.TempDir()
	write := func(name, body string) {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module github.com/relux-works/curator\n\ngo 1.25.5\n")
	write("tools/testshard/windows-weights.json", "{}")
	write("cmd/curator/cases_test.go", `package main
import("fmt";"testing")
func TestA(t *testing.T){t.Log("ran A")}
func TestB(t *testing.T){t.Log("ran B")}
func TestC(t *testing.T){t.Log("ran C")}
func TestD(t *testing.T){t.Log("ran D")}
func Example(){fmt.Println("hello")
// Output: hello
}
func FuzzInput(f *testing.F){f.Add("seed");f.Fuzz(func(t *testing.T,s string){if s!="seed"{t.Fatal(s)}})}
`)
	for i := 0; i < 4; i++ {
		dir := filepath.Join(root, "test-evidence-windows-latest-"+string(rune('0'+i)), "test")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		output, err := shardCLI(t, root, "run", dir, string(rune('0'+i)), "github.com/relux-works/curator/cmd/curator")
		if err != nil {
			t.Fatalf("shard %d: %v\n%s", i, err, output)
		}
		if err := os.WriteFile(filepath.Join(dir, "go-test.json"), output, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if output, err := shardCLI(t, root, "merge", root); err != nil {
		t.Fatalf("merge: %v %s", err, output)
	}
	before, err := os.ReadFile(filepath.Join(root, "go-test.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"TestA", "TestB", "TestC", "TestD", "Example", "FuzzInput"} {
		if !strings.Contains(string(before), `"Test":"`+name+`"`) {
			t.Fatalf("missing %s", name)
		}
	}
	dir := filepath.Join(root, "test-evidence-windows-latest-0", "test", "go-test.json")
	data, err := os.ReadFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	removed := false
	for _, line := range strings.Split(string(data), "\n") {
		var e event
		if line != "" {
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				t.Fatal(err)
			}
		}
		if !removed && e.Action == "pass" && e.Test == "Example" {
			removed = true
			continue
		}
		lines = append(lines, line)
	}
	if !removed {
		t.Fatal("control not applied")
	}
	if err := os.WriteFile(dir, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatal(err)
	}
	output, err := shardCLI(t, root, "merge", root)
	if err == nil || !strings.Contains(string(output), "missing test verdict") {
		t.Fatalf("control: %v %s", err, output)
	}
	after, err := os.ReadFile(filepath.Join(root, "go-test.json"))
	if err != nil || string(after) != string(before) {
		t.Fatal("refusal changed prior combined evidence")
	}
}

// The CLI holds the macOS host toolchain for its entire package. The unchanged
// baseline race CLI takes 1,488s, longer than LockHostGOROOT's per-test 20m
// budget. Use this package's existing test deadline for that cross-package wait;
// no CI timeout or production trust check is relaxed.
func lockShardGoTools(t *testing.T) {
	t.Helper()
	if os.Getenv("CURATOR_TEST_SHARD_PARENT_LOCK") == "1" {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if deadline, ok := t.Deadline(); ok {
		cancel()
		ctx, cancel = context.WithDeadline(context.Background(), deadline.Add(-time.Second))
	}
	t.Cleanup(cancel)
	lock, err := testtoolchain.AcquireHostGOROOT(ctx)
	if err != nil {
		t.Fatalf("acquire host Go toolchain for shard tests: %v", err)
	}
	if lock != nil {
		t.Cleanup(func() {
			if err := lock.Close(); err != nil {
				t.Errorf("release host Go toolchain: %v", err)
			}
		})
	}
}
