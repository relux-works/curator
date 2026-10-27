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
	"sort"
	"strings"
)

// All ordinary lanes share discovery, partitioning, execution and admission.
// Windows keeps its measured four-shard/120m configuration unchanged.
type testLane struct {
	goos, lane, timeout string
	shards, margin      int
}

func ordinaryConfig(lane, goos string) (testLane, error) {
	if lane == "windows-test" && goos == "windows" {
		return testLane{goos, lane, "120m", 4, 3600}, nil
	}
	if lane == "unix-test" {
		switch goos {
		case "linux":
			return testLane{goos, lane, "60m", 3, 1800}, nil
		case "darwin":
			return testLane{goos, lane, "60m", 2, 1800}, nil
		}
	}
	return testLane{}, fmt.Errorf("%s lane does not support native %s", lane, goos)
}

var ordinaryPackages = []string{
	"github.com/relux-works/curator/internal/envprofile",
	"github.com/relux-works/curator/internal/install",
	"github.com/relux-works/curator/cmd/curator",
}

type testManifest struct {
	Commit, GOOS, Lane string
	Shard              int
	Inventories        map[string][]string
	Selector           string
	ExitCode           int
}

// The three measured slow packages share native jobs. Hashing only the name
// makes a union selector safe even when packages declare identical test names.
func (c testLane) partition(tests map[string][]string) (map[string][][]string, error) {
	if len(tests) != len(ordinaryPackages) {
		return nil, fmt.Errorf("ordinary package inventory mismatch")
	}
	result := map[string][][]string{}
	for _, p := range ordinaryPackages {
		shards := make([][]string, c.shards)
		seen := map[string]bool{}
		for _, name := range tests[p] {
			if name == "" || strings.Contains(name, "/") || seen[name] {
				return nil, fmt.Errorf("invalid or duplicate test %q in %s", name, p)
			}
			seen[name] = true
			h := fnv.New64a()
			_, _ = h.Write([]byte(name))
			// #nosec G115 -- Native configurations fix the positive shard count to 2, 3 or 4.
			i := int(h.Sum64() % uint64(c.shards))
			shards[i] = append(shards[i], name)
		}
		for i := range shards {
			sort.Strings(shards[i])
			if len(shards[i]) == 0 {
				return nil, fmt.Errorf("empty ordinary shard %d in %s", i, p)
			}
		}
		result[p] = shards
	}
	return result, nil
}

func (c testLane) selector(parts map[string][][]string, shard int) string {
	names := map[string]bool{}
	for _, p := range ordinaryPackages {
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

func (c testLane) inventory() (map[string][]string, error) {
	tests := map[string][]string{}
	for _, p := range ordinaryPackages {
		list, err := nativeInventory(p, false)
		if err != nil {
			return nil, err
		}
		tests[p] = list
	}
	return tests, nil
}

func (c testLane) validateStream(data []byte, parts map[string][][]string, shard int) error {
	streams := map[string]*bytes.Buffer{}
	for _, p := range ordinaryPackages {
		streams[p] = &bytes.Buffer{}
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("invalid ordinary test JSON: %w", err)
		}
		var e event
		if err := json.Unmarshal(raw, &e); err != nil {
			return err
		}
		stream, ok := streams[e.Package]
		if !ok {
			return fmt.Errorf("unexpected ordinary package %q", e.Package)
		}
		stream.Write(raw)
		stream.WriteByte('\n')
	}
	for _, p := range ordinaryPackages {
		elapsed, err := validatePackageStream(streams[p].Bytes(), parts[p][shard], p, float64(c.margin))
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		fmt.Printf("%s: shard %d %s: %d tests, %.3fs (<%ds)\n", c.lane, shard, p, len(parts[p][shard]), elapsed, c.margin)
	}
	return nil
}

func (c testLane) run(dir string, shard int, tests map[string][]string, commit string) error {
	if shard < 0 || shard >= c.shards {
		return fmt.Errorf("ordinary shard must be 0..%d", c.shards-1)
	}
	parts, err := c.partition(tests)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	m := testManifest{Commit: commit, GOOS: c.goos, Lane: c.lane, Shard: shard, Inventories: tests, Selector: c.selector(parts, shard)}
	// #nosec G304 -- The CLI evidence directory is the explicit output boundary.
	stream, err := os.Create(filepath.Join(dir, "go-test.json"))
	if err != nil {
		return err
	}
	args := []string{"test", "-json", "-count=1", "-timeout=" + c.timeout, "-run", m.Selector}
	// #nosec G204 -- Fixed packages and an escaped native-name selector; no shell.
	cmd := exec.Command("go", append(args, ordinaryPackages...)...)
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
	return c.validateStream(data, parts, shard)
}

func (c testLane) merge(dir, out string, tests map[string][]string, commit string) error {
	parts, err := c.partition(tests)
	if err != nil {
		return err
	}
	paths, err := filepath.Glob(filepath.Join(dir, "*", "manifest.json"))
	if err != nil {
		return err
	}
	if len(paths) != c.shards {
		return fmt.Errorf("want %d ordinary manifests, got %d", c.shards, len(paths))
	}
	seen := map[int]bool{}
	var merged bytes.Buffer
	for _, path := range paths {
		// #nosec G304 -- Manifests come only from the explicit artifact glob above.
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var m testManifest
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if m.Shard < 0 || m.Shard >= c.shards || seen[m.Shard] {
			return fmt.Errorf("invalid or duplicate ordinary shard %d", m.Shard)
		}
		seen[m.Shard] = true
		if m.Commit != commit || m.GOOS != c.goos || m.Lane != c.lane || m.ExitCode != 0 || !reflect.DeepEqual(m.Inventories, tests) || m.Selector != c.selector(parts, m.Shard) {
			return fmt.Errorf("ordinary shard %d: source/platform/lane/inventory/selector/exit mismatch", m.Shard)
		}
		data, err = os.ReadFile(filepath.Join(filepath.Dir(path), "go-test.json"))
		if err != nil {
			return err
		}
		if err := c.validateStream(data, parts, m.Shard); err != nil {
			return err
		}
		merged.Write(data)
		merged.WriteByte('\n')
	}
	// The shell gate uses only this admitted package list to filter execution.
	if err := os.WriteFile(out+".packages", []byte(strings.Join(ordinaryPackages, "\n")+"\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(out, merged.Bytes(), 0644); err != nil {
		return err
	}
	for _, p := range ordinaryPackages {
		fmt.Printf("%s: union equals full native list: %s: %d tests\n", c.lane, p, len(tests[p]))
	}
	return nil
}

func (c testLane) execute(mode, dir, out string, shard int) error {
	if mode != "run" && mode != "merge" {
		return fmt.Errorf("unknown ordinary mode %q", mode)
	}
	tests, err := c.inventory()
	if err != nil {
		return err
	}
	commit, err := sourceCommit()
	if err != nil {
		return err
	}
	if mode == "merge" {
		return c.merge(dir, out, tests, commit)
	}
	return c.run(dir, shard, tests, commit)
}
