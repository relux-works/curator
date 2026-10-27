package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

const windowsShardCount = 4
const windowsTimeout = "120m"
const windowsMargin = 3600

var windowsPackages = []string{
	"github.com/relux-works/curator/internal/envprofile",
	"github.com/relux-works/curator/internal/install",
	"github.com/relux-works/curator/cmd/curator",
}

type windowsManifest struct {
	Commit, GOOS, Lane string
	Shard              int
	Inventories        map[string][]string
	Selector           string
	ExitCode           int
}

// The three measured slow packages share four jobs. Hashing only the name
// makes a union selector safe even when packages declare identical test names.
func windowsPartition(tests map[string][]string) (map[string][][]string, error) {
	if len(tests) != len(windowsPackages) {
		return nil, fmt.Errorf("windows package inventory mismatch")
	}
	result := map[string][][]string{}
	for _, p := range windowsPackages {
		shards := make([][]string, windowsShardCount)
		seen := map[string]bool{}
		for _, name := range tests[p] {
			if name == "" || strings.Contains(name, "/") || seen[name] {
				return nil, fmt.Errorf("invalid or duplicate test %q in %s", name, p)
			}
			seen[name] = true
			h := fnv.New64a()
			_, _ = h.Write([]byte(name))
			i := int(h.Sum64() % windowsShardCount)
			shards[i] = append(shards[i], name)
		}
		for i := range shards {
			sort.Strings(shards[i])
			if len(shards[i]) == 0 {
				return nil, fmt.Errorf("empty Windows shard %d in %s", i, p)
			}
		}
		result[p] = shards
	}
	return result, nil
}

func windowsSelector(parts map[string][][]string, shard int) string {
	names := map[string]bool{}
	for _, p := range windowsPackages {
		for _, name := range parts[p][shard] {
			names[name] = true
		}
	}
	var list []string
	for name := range names {
		list = append(list, name)
	}
	sort.Strings(list)
	return selector(list)
}

func windowsInventory() (map[string][]string, error) {
	tests := map[string][]string{}
	for _, p := range windowsPackages {
		list, err := nativeInventory(p, false)
		if err != nil {
			return nil, err
		}
		tests[p] = list
	}
	return tests, nil
}

func validateWindowsStream(data []byte, parts map[string][][]string, shard int) error {
	streams := map[string]*bytes.Buffer{}
	for _, p := range windowsPackages {
		streams[p] = &bytes.Buffer{}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("invalid Windows test JSON: %w", err)
		}
		var e event
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}
		stream, ok := streams[e.Package]
		if !ok {
			return fmt.Errorf("unexpected Windows package %q", e.Package)
		}
		stream.Write(raw)
		stream.WriteByte('\n')
	}
	for _, p := range windowsPackages {
		elapsed, err := validatePackageStream(streams[p].Bytes(), parts[p][shard], p, windowsMargin)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		fmt.Printf("windows-test: shard %d %s: %d tests, %.3fs (<%ds)\n", shard, p, len(parts[p][shard]), elapsed, windowsMargin)
	}
	return nil
}

func runWindowsShard(dir string, shard int, tests map[string][]string, commit string) error {
	if shard < 0 || shard >= windowsShardCount {
		return fmt.Errorf("windows shard must be 0..%d", windowsShardCount-1)
	}
	parts, err := windowsPartition(tests)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	m := windowsManifest{Commit: commit, GOOS: "windows", Lane: "windows-test", Shard: shard, Inventories: tests, Selector: windowsSelector(parts, shard)}
	// #nosec G304 -- The CLI evidence directory is the explicit output boundary.
	stream, err := os.Create(filepath.Join(dir, "go-test.json"))
	if err != nil {
		return err
	}
	args := []string{"test", "-json", "-count=1", "-timeout=" + windowsTimeout, "-run", m.Selector}
	// #nosec G204 -- Fixed packages and an escaped native-name selector; no shell.
	cmd := exec.Command("go", append(args, windowsPackages...)...)
	cmd.Stdout, cmd.Stderr = stream, os.Stderr
	runErr := cmd.Run()
	if err := stream.Close(); err != nil {
		return err
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
	// #nosec G304 -- Read back the stream written in the explicit evidence directory.
	data, err = os.ReadFile(filepath.Join(dir, "go-test.json"))
	if err != nil {
		return err
	}
	return validateWindowsStream(data, parts, shard)
}

func mergeWindowsEvidence(dir, out string, tests map[string][]string, commit string) error {
	parts, err := windowsPartition(tests)
	if err != nil {
		return err
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*", "manifest.json"))
	if err != nil {
		return err
	}
	if len(paths) != windowsShardCount {
		return fmt.Errorf("want %d Windows manifests, got %d", windowsShardCount, len(paths))
	}
	seen := map[int]bool{}
	var merged bytes.Buffer
	for _, path := range paths {
		// #nosec G304 -- Manifests come only from the explicit artifact glob above.
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var m windowsManifest
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if m.Shard < 0 || m.Shard >= windowsShardCount || seen[m.Shard] {
			return fmt.Errorf("invalid or duplicate Windows shard %d", m.Shard)
		}
		seen[m.Shard] = true
		if m.Commit != commit || m.GOOS != "windows" || m.Lane != "windows-test" || m.ExitCode != 0 || !reflect.DeepEqual(m.Inventories, tests) || m.Selector != windowsSelector(parts, m.Shard) {
			return fmt.Errorf("windows shard %d: source/platform/lane/inventory/selector/exit mismatch", m.Shard)
		}
		data, err = os.ReadFile(filepath.Join(filepath.Dir(path), "go-test.json"))
		if err != nil {
			return err
		}
		if err := validateWindowsStream(data, parts, m.Shard); err != nil {
			return err
		}
		merged.Write(data)
		merged.WriteByte('\n')
	}
	// The shell gate uses only this admitted package list to filter execution.
	if err := os.WriteFile(out+".packages", []byte(strings.Join(windowsPackages, "\n")+"\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(out, merged.Bytes(), 0644); err != nil {
		return err
	}
	for _, p := range windowsPackages {
		fmt.Printf("windows-test: union equals full native list: %s: %d tests\n", p, len(tests[p]))
	}
	return nil
}

func windowsLane(mode, dir, out string, shard int) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("windows-test lane requires native Windows")
	}
	if mode != "run" && mode != "merge" {
		return fmt.Errorf("unknown Windows mode %q", mode)
	}
	tests, err := windowsInventory()
	if err != nil {
		return err
	}
	commit, err := sourceCommit()
	if err != nil {
		return err
	}
	if mode == "merge" {
		return mergeWindowsEvidence(dir, out, tests, commit)
	}
	return runWindowsShard(dir, shard, tests, commit)
}
