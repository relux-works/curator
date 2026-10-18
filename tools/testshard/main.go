// testshard partitions top-level cases; children and assertions stay with their parent.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const shardCount = 4

var heavy = map[string]bool{
	"github.com/relux-works/curator/cmd/curator":         true,
	"github.com/relux-works/curator/internal/envprofile": true,
	"github.com/relux-works/curator/internal/install":    true,
}

type event struct {
	Action  string
	Package string
	Test    string
	Output  string
}
type selection struct {
	Package  string
	Tests    []string
	Selected []string
}
type manifest struct {
	Index    int
	Count    int
	Packages []selection
}

func partition(tests []string, weights map[string]float64) [][]string {
	names := append([]string(nil), tests...)
	weight := func(name string) float64 {
		if n := weights[name]; n > 0 {
			return n
		}
		return 5
	}
	sort.Slice(names, func(i, j int) bool {
		a, b := weight(names[i]), weight(names[j])
		if a == b {
			return names[i] < names[j]
		}
		return a > b
	})
	bins := make([][]string, shardCount)
	loads := make([]float64, shardCount)
	for _, name := range names {
		at := 0
		for i := 1; i < shardCount; i++ {
			if loads[i] < loads[at] {
				at = i
			}
		}
		bins[at] = append(bins[at], name)
		loads[at] += weight(name)
	}
	for _, bin := range bins {
		sort.Strings(bin)
	}
	return bins
}

func inventory(goTool string, packages []string) ([]selection, error) {
	args := append([]string{"test", "-json", "-list", "^(Test|Example|Fuzz)"}, packages...)
	cmd := exec.Command(goTool, args...) // #nosec G204,G702 -- CI selects the Go launcher and package/test arguments.
	payload, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("test inventory: %w: %s", err, payload)
	}
	return parseInventory(payload, packages)
}

func parseInventory(payload []byte, packages []string) ([]selection, error) {
	found := map[string]*selection{}
	passed := map[string]bool{}
	for _, pkg := range packages {
		found[pkg] = &selection{Package: pkg, Tests: []string{}}
	}
	dec := json.NewDecoder(strings.NewReader(string(payload)))
	for {
		var e event
		if err := dec.Decode(&e); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		row, ok := found[e.Package]
		if !ok {
			return nil, fmt.Errorf("unexpected inventory package %s", e.Package)
		}
		if e.Action == "pass" || e.Action == "skip" {
			passed[e.Package] = true
		}
		if e.Action == "fail" {
			return nil, fmt.Errorf("failed inventory %s", e.Package)
		}
		name := strings.TrimSpace(e.Output)
		if e.Action == "output" && regexp.MustCompile(`^(Test|Example|Fuzz)[^ \t\r\n]*$`).MatchString(name) {
			row.Tests = append(row.Tests, name)
		}
	}
	var result []selection
	for _, pkg := range packages {
		if !passed[pkg] {
			return nil, fmt.Errorf("incomplete inventory %s", pkg)
		}
		sort.Strings(found[pkg].Tests)
		result = append(result, *found[pkg])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Package < result[j].Package })
	return result, nil
}

func run(dir string, index int, packages []string) error {
	if index < 0 || index >= shardCount {
		return errors.New("shard index outside 0..3")
	}
	goTool := os.Getenv("GO")
	if goTool == "" {
		goTool = "go"
	}
	rows, err := inventory(goTool, packages)
	if err != nil {
		return err
	}
	data, err := os.ReadFile("tools/testshard/windows-weights.json")
	if err != nil {
		return err
	}
	var weights map[string]map[string]float64
	if err := json.Unmarshal(data, &weights); err != nil {
		return err
	}
	for i := range rows {
		if heavy[rows[i].Package] {
			rows[i].Selected = partition(rows[i].Tests, weights[rows[i].Package])[index]
		} else if index == 0 {
			rows[i].Selected = append([]string(nil), rows[i].Tests...)
		}
	}
	m := manifest{Index: index, Count: shardCount, Packages: rows}
	data, err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "shard-manifest.json"), data, 0600); err != nil {
		return err
	}
	var wg sync.WaitGroup
	results := make(chan error, (len(rows)+1)*3)
	var light []string
	launch := func(args []string, label string) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// One file per process: never interleave JSON writes from concurrent tests.
			out, err := os.Create(filepath.Join(dir, "shard-"+label+".json")) // #nosec G304 -- CI-owned evidence directory and fixed package labels.
			if err != nil {
				results <- err
				return
			}
			defer func() {
				if err := out.Close(); err != nil {
					results <- err
				}
			}()
			stderr, err := os.Create(filepath.Join(dir, "shard-"+label+".stderr")) // #nosec G304 -- CI-owned evidence directory and fixed package labels.
			if err != nil {
				results <- err
				return
			}
			defer func() {
				if err := stderr.Close(); err != nil {
					results <- err
				}
			}()
			timeout := os.Getenv("GO_TEST_TIMEOUT")
			if timeout == "" {
				timeout = "120m"
			}
			base := []string{"test", "-json", "-count=1", "-timeout", timeout}
			base = append(base, strings.Fields(os.Getenv("GO_TEST_FLAGS"))...)
			cmd := exec.Command(goTool, append(base, args...)...) // #nosec G204,G702 -- CI selects the Go launcher; selected names are regexp-escaped.
			cmd.Stdout = out
			cmd.Stderr = stderr
			if err := cmd.Run(); err != nil {
				results <- fmt.Errorf("%s: %w", label, err)
			}
		}()
	}
	for _, row := range rows {
		if heavy[row.Package] {
			if len(row.Selected) == 0 {
				continue
			}
			escaped := make([]string, len(row.Selected))
			for i, n := range row.Selected {
				escaped[i] = regexp.QuoteMeta(n)
			}
			launch([]string{"-run", "^(" + strings.Join(escaped, "|") + ")$", row.Package}, filepath.Base(row.Package))
		} else if index == 0 {
			light = append(light, row.Package)
		}
	}
	if len(light) > 0 {
		launch(light, "light")
	}
	wg.Wait()
	close(results)
	var failures []error
	for err := range results {
		failures = append(failures, err)
	}
	// Keep the ordinary gate's merged stream format for artifact consumers.
	files, err := filepath.Glob(filepath.Join(dir, "shard-*.json"))
	if err != nil {
		return err
	}
	for _, path := range files {
		if strings.HasSuffix(path, "shard-manifest.json") {
			continue
		}
		data, err := os.ReadFile(path) // #nosec G304 -- Glob enumerates only this run's evidence files.
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
	}
	return errors.Join(failures...)
}

func verify(manifests []manifest, streams [][]byte) error {
	if len(manifests) != shardCount || len(streams) != shardCount {
		return errors.New("missing shard")
	}
	seenShard := map[int]bool{}
	expected := map[string]bool{}
	var baseline []selection
	for i, m := range manifests {
		if m.Count != shardCount || m.Index < 0 || m.Index >= shardCount || seenShard[m.Index] {
			return errors.New("invalid or duplicate shard")
		}
		seenShard[m.Index] = true
		inventory := make([]selection, len(m.Packages))
		for j, row := range m.Packages {
			inventory[j] = selection{Package: row.Package, Tests: row.Tests}
		}
		if i == 0 {
			baseline = inventory
		} else if !reflect.DeepEqual(inventory, baseline) {
			return errors.New("shard inventories disagree")
		}
		selected := map[string]bool{}
		packagePass := map[string]bool{}
		for _, row := range m.Packages {
			known := map[string]bool{}
			for _, name := range row.Tests {
				known[name] = true
				if i == 0 {
					expected[row.Package+"/"+name] = true
				}
			}
			for _, name := range row.Selected {
				key := row.Package + "/" + name
				if _, duplicate := selected[key]; !known[name] || duplicate {
					return fmt.Errorf("invalid selection %s", key)
				}
				selected[key] = false
			}
			if len(row.Selected) > 0 || (!heavy[row.Package] && m.Index == 0) {
				packagePass[row.Package] = false
			}
		}
		dec := json.NewDecoder(strings.NewReader(string(streams[i])))
		for {
			var e event
			if err := dec.Decode(&e); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if e.Action == "fail" {
				return fmt.Errorf("failed event %s/%s", e.Package, e.Test)
			}
			if e.Test == "" {
				if e.Action == "pass" || e.Action == "skip" {
					if _, ok := packagePass[e.Package]; !ok {
						return fmt.Errorf("unexpected package %s", e.Package)
					}
					packagePass[e.Package] = true
				}
				continue
			}
			if strings.Contains(e.Test, "/") {
				continue
			}
			if e.Action == "pass" || e.Action == "skip" {
				key := e.Package + "/" + e.Test
				old, ok := selected[key]
				if !ok || old {
					return fmt.Errorf("unexpected or duplicate test %s", key)
				}
				selected[key] = true
			}
		}
		for pkg, pass := range packagePass {
			if !pass {
				return fmt.Errorf("missing package verdict %s", pkg)
			}
		}
		for key, pass := range selected {
			if !pass {
				return fmt.Errorf("missing test verdict %s", key)
			}
			if !expected[key] {
				return fmt.Errorf("duplicate cross-shard test %s", key)
			}
			expected[key] = false
		}
	}
	for key, missing := range expected {
		if missing {
			return fmt.Errorf("unselected test %s", key)
		}
	}
	return nil
}

func merge(dir string) (err error) {
	var manifests []manifest
	var streams [][]byte
	for i := 0; i < shardCount; i++ {
		root := filepath.Join(dir, "test-evidence-windows-latest-"+strconv.Itoa(i), "test")
		data, err := os.ReadFile(filepath.Join(root, "shard-manifest.json")) // #nosec G304 -- CI-owned artifact root and fixed filename.
		if err != nil {
			return err
		}
		var m manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		if m.Index != i {
			return fmt.Errorf("artifact index mismatch: directory %d manifest %d", i, m.Index)
		}
		manifests = append(manifests, m)
		stream, err := os.ReadFile(filepath.Join(root, "go-test.json")) // #nosec G304 -- CI-owned artifact root and fixed filename.
		if err != nil {
			return err
		}
		streams = append(streams, stream)
	}
	if err := verify(manifests, streams); err != nil {
		return err
	}
	out, err := os.Create(filepath.Join(dir, "go-test.json")) // #nosec G304 -- CI-owned aggregate evidence directory.
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, out.Close()) }()
	for _, stream := range streams {
		if _, err := out.Write(stream); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var err error
	if len(os.Args) >= 5 && os.Args[1] == "run" {
		var index int
		index, err = strconv.Atoi(os.Args[3])
		if err == nil {
			err = run(os.Args[2], index, os.Args[4:])
		}
	} else if len(os.Args) == 3 && os.Args[1] == "merge" {
		err = merge(os.Args[2])
	} else {
		err = errors.New("usage: testshard run DIR INDEX PACKAGES... | merge DIR")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
