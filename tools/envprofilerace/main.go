// Command envprofilerace partitions CI test lanes and verifies their evidence.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const hashShardCount = 8
const shardCount = hashShardCount + 1
const ownershipControls = "TestGlobalNativeOwnershipNarrowingControls"
const pkg = "github.com/relux-works/curator/internal/envprofile"
const timeout = "60m"
const marginSeconds = 1800

type manifest struct {
	Commit   string
	GOOS     string
	Shard    int
	Tests    []string
	Selector string
	ExitCode int
}

type event struct {
	ImportPath string
	Action     string
	Package    string
	Test       string
	Elapsed    float64
}

func shardFor(name string) int {
	// The most expensive measured mutation suite gets its own runner. Keep
	// all other assignments stable, including tests discovered in the future.
	if name == ownershipControls {
		return hashShardCount
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int(h.Sum64() % hashShardCount)
}

func partition(tests []string) ([][]string, error) {
	shards := make([][]string, shardCount)
	seen := map[string]bool{}
	for _, name := range tests {
		if name == "" || strings.Contains(name, "/") || seen[name] {
			return nil, fmt.Errorf("invalid or duplicate test %q", name)
		}
		seen[name] = true
		i := shardFor(name)
		shards[i] = append(shards[i], name)
	}
	for i := range shards {
		sort.Strings(shards[i])
		if len(shards[i]) == 0 {
			return nil, fmt.Errorf("empty shard %d", i)
		}
	}
	return shards, nil
}

func selector(tests []string) string {
	escaped := make([]string, len(tests))
	for i, name := range tests {
		escaped[i] = regexp.QuoteMeta(name)
	}
	return "^(" + strings.Join(escaped, "|") + ")$"
}

func inventory() ([]string, error) {
	return nativeInventory("./internal/envprofile", true)
}

func nativeInventory(path string, race bool) ([]string, error) {
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	// #nosec G204 -- Callers supply fixed package constants; no shell is invoked.
	cmd := exec.Command("go", append(args, "-list", ".", path)...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("list tests for %s: %w", path, err)
	}
	var tests []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Test") || strings.HasPrefix(line, "Example") || strings.HasPrefix(line, "Fuzz") {
			tests = append(tests, line)
		}
	}
	if len(tests) == 0 {
		return nil, fmt.Errorf("empty test inventory for %s", path)
	}
	sort.Strings(tests)
	return tests, nil
}

func sourceCommit() (string, error) {
	// Actions checks out GITHUB_SHA (including the PR merge revision). Local
	// callers with detached Git metadata pass the same explicit source pin
	// before clearing Git selectors from the test environment.
	if sha := os.Getenv("GITHUB_SHA"); sha != "" {
		if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(sha) {
			return "", fmt.Errorf("GITHUB_SHA must be a full commit revision")
		}
		return sha, nil
	}
	out, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("source revision: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func runShard(dir string, shard int) error {
	if shard < 0 || shard >= shardCount {
		return fmt.Errorf("shard must be 0..%d", shardCount-1)
	}
	tests, err := inventory()
	if err != nil {
		return err
	}
	shards, err := partition(tests)
	if err != nil {
		return err
	}
	commit, err := sourceCommit()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// #nosec G304 -- The explicit CLI evidence directory is the intended output boundary.
	stream, err := os.Create(filepath.Join(dir, "go-test.json"))
	if err != nil {
		return err
	}
	m := manifest{Commit: commit, GOOS: runtime.GOOS, Shard: shard, Tests: tests, Selector: selector(shards[shard])}
	// #nosec G204 -- Selector is an escaped list of discovered test names; no shell is invoked.
	cmd := exec.Command("go", "test", "-race", "-json", "-count=1", "-timeout="+timeout, "-run", m.Selector, "./internal/envprofile")
	cmd.Stdout = stream
	cmd.Stderr = os.Stderr
	fmt.Printf("envprofilerace: shard %d/%d, %d of %d tests, timeout=%s\n", shard, shardCount, len(shards[shard]), len(tests), timeout)
	runErr := cmd.Run()
	if err := stream.Close(); err != nil {
		return fmt.Errorf("close test JSON: %w", err)
	}
	if runErr != nil {
		m.ExitCode = 1
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
		return err
	}
	if runErr != nil {
		return runErr
	}
	// #nosec G304 -- Read back the JSON just written to the explicit evidence directory.
	data, err = os.ReadFile(filepath.Join(dir, "go-test.json"))
	if err != nil {
		return err
	}
	elapsed, err := validateStream(data, shards[shard])
	if err != nil {
		return err
	}
	fmt.Printf("envprofilerace: shard %d PASS %.3fs (<%ds)\n", shard, elapsed, marginSeconds)
	return nil
}

// Require completed top-level tests, not just the planned -run expression.
// Subtest obligations and skip classification remain with platform-case-gate.sh.
func validateStream(data []byte, tests []string) (float64, error) {
	return validatePackageStream(data, tests, pkg, marginSeconds)
}

func validatePackageStream(data []byte, tests []string, packageName string, margin float64) (float64, error) {
	want := map[string]bool{}
	for _, name := range tests {
		want[name] = true
	}
	seen := map[string]bool{}
	packages := 0
	var elapsed float64
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var e event
		err := dec.Decode(&e)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("invalid test JSON: %w", err)
		}
		// Go 1.24+ build JSON has ImportPath and no Package. Admit only the
		// selected package/test binary variants; build output is not completion.
		if strings.HasPrefix(e.Action, "build-") {
			if e.Test != "" || e.Package != "" || e.ImportPath == "" {
				return 0, fmt.Errorf("invalid build event")
			}
			valid := e.ImportPath == packageName || e.ImportPath == packageName+".test" || e.ImportPath == packageName+" ["+packageName+".test]" || e.ImportPath == packageName+"_test ["+packageName+".test]"
			if !valid {
				return 0, fmt.Errorf("unexpected build import %q", e.ImportPath)
			}
			if e.Action == "build-fail" {
				return 0, fmt.Errorf("build failure: %s", e.ImportPath)
			}
			if e.Action != "build-output" {
				return 0, fmt.Errorf("unsupported event action %q", e.Action)
			}
			continue
		}
		if e.ImportPath != "" {
			return 0, fmt.Errorf("unexpected build import on test event")
		}
		switch e.Action {
		case "start", "run", "pause", "cont", "output", "pass", "skip", "fail":
		default:
			return 0, fmt.Errorf("unsupported event action %q", e.Action)
		}
		if e.Package != packageName {
			return 0, fmt.Errorf("unexpected package %q", e.Package)
		}
		if e.Action == "fail" {
			return 0, fmt.Errorf("failed test/package %s", e.Test)
		}
		if e.Test == "" && e.Action == "pass" {
			packages++
			elapsed = e.Elapsed
		}
		if e.Test != "" && !strings.Contains(e.Test, "/") && (e.Action == "pass" || e.Action == "skip") {
			if !want[e.Test] || seen[e.Test] {
				return 0, fmt.Errorf("unexpected or duplicate test %s", e.Test)
			}
			seen[e.Test] = true
		}
	}
	for _, name := range tests {
		if !seen[name] {
			return 0, fmt.Errorf("missing completed test %s", name)
		}
	}
	if packages != 1 {
		return 0, fmt.Errorf("want one passing package, got %d", packages)
	}
	if elapsed <= 0 || elapsed >= margin {
		return 0, fmt.Errorf("package elapsed %.3fs outside (0,%.0f)s margin", elapsed, margin)
	}
	return elapsed, nil
}

func mergeEvidence(dir, out string, tests []string, commit, goos string) error {
	shards, err := partition(tests)
	if err != nil {
		return err
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*", "manifest.json"))
	if err != nil {
		return err
	}
	if len(paths) != shardCount {
		return fmt.Errorf("want %d shard manifests, got %d", shardCount, len(paths))
	}
	seen := map[int]bool{}
	var merged bytes.Buffer
	for _, path := range paths {
		// #nosec G304 -- Manifests are read only from the explicit CI artifact glob above.
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if m.Shard < 0 || m.Shard >= shardCount || seen[m.Shard] {
			return fmt.Errorf("invalid or duplicate shard %d", m.Shard)
		}
		seen[m.Shard] = true
		if m.Commit != commit || m.GOOS != goos || m.ExitCode != 0 || !reflect.DeepEqual(m.Tests, tests) || m.Selector != selector(shards[m.Shard]) {
			return fmt.Errorf("shard %d: source/platform/inventory/selector/exit mismatch", m.Shard)
		}
		data, err = os.ReadFile(filepath.Join(filepath.Dir(path), "go-test.json"))
		if err != nil {
			return err
		}
		elapsed, err := validateStream(data, shards[m.Shard])
		if err != nil {
			return fmt.Errorf("shard %d: %w", m.Shard, err)
		}
		fmt.Printf("envprofilerace: verified shard %d: %d tests, %.3fs\n", m.Shard, len(shards[m.Shard]), elapsed)
		merged.Write(data)
		merged.WriteByte('\n')
	}
	// Publish only after the complete union has been verified.
	if err := os.WriteFile(out, merged.Bytes(), 0644); err != nil {
		return err
	}
	fmt.Printf("envprofilerace: union equals full race list: %d tests across %d shards\n", len(tests), shardCount)
	return nil
}

func main() {
	lane := flag.String("lane", "race", "race, windows-test or unix-test")
	mode := flag.String("mode", "run", "run or merge")
	dir := flag.String("dir", "", "shard output, or downloaded artifact directory")
	out := flag.String("out", "", "merged JSON output")
	shard := flag.Int("shard", -1, "zero-based shard index")
	flag.Parse()
	var err error
	if *lane == "windows-test" || *lane == "unix-test" {
		var config testLane
		config, err = ordinaryConfig(*lane, runtime.GOOS)
		if err == nil {
			err = config.execute(*mode, *dir, *out, *shard)
		}
	} else if *lane != "race" {
		err = fmt.Errorf("unknown lane %q", *lane)
	} else {
		switch *mode {
		case "run":
			err = runShard(*dir, *shard)
		case "merge":
			var tests []string
			var commit string
			tests, err = inventory()
			if err == nil {
				commit, err = sourceCommit()
			}
			if err == nil {
				err = mergeEvidence(*dir, *out, tests, commit, runtime.GOOS)
			}
		default:
			err = fmt.Errorf("unknown mode %q", *mode)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "envprofilerace:", err)
		os.Exit(1)
	}
}
