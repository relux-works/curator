package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func windowsFixtureInventory() map[string][]string {
	all := map[string][]string{}
	for i, p := range windowsPackages {
		for j := 0; j < 80; j++ {
			all[p] = append(all[p], fmt.Sprintf("TestP%dCase%03d", i, j))
		}
		all[p] = append(all[p], "TestShared", "ExampleFuture", "FuzzFuture")
	}
	return all
}

// One common selector must select every native test once in its own package,
// including future tests and names shared by packages, with no cross-selection.
func TestWindowsPartitionCoverage(t *testing.T) {
	tests := windowsFixtureInventory()
	parts, err := windowsPartition(tests)
	if err != nil {
		t.Fatal(err)
	}
	for p, names := range tests {
		for _, name := range names {
			count := 0
			for i := 0; i < windowsShardCount; i++ {
				matched := regexp.MustCompile(windowsSelector(parts, i)).MatchString(name)
				expected := false
				for _, selected := range parts[p][i] {
					expected = expected || selected == name
				}
				if matched != expected {
					t.Fatalf("cross-package selection: %s %s shard %d", p, name, i)
				}
				if matched {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("%s %s selected %d times", p, name, count)
			}
		}
	}
	delete(tests, windowsPackages[0])
	if _, err := windowsPartition(tests); err == nil || !strings.Contains(err.Error(), "inventory mismatch") {
		t.Fatalf("missing package admitted: %v", err)
	}
	tests = windowsFixtureInventory()
	tests[windowsPackages[0]] = append(tests[windowsPackages[0]], "TestShared")
	if _, err := windowsPartition(tests); err == nil || !strings.Contains(err.Error(), "duplicate test") {
		t.Fatalf("duplicate test admitted: %v", err)
	}
}

func windowsFixtureStream(parts map[string][][]string, shard int) []byte {
	var out strings.Builder
	for _, p := range windowsPackages {
		for _, name := range parts[p][shard] {
			fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Package\":%q,\"Test\":%q}\n", p, name)
		}
		fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Package\":%q,\"Elapsed\":42}\n", p)
	}
	return []byte(out.String())
}

// Complete positive evidence is admitted; each bounded plant must fail for
// its intended inventory, execution, source or margin violation, without outputs.
func TestWindowsMergeControls(t *testing.T) {
	for _, plant := range []string{"valid", "missing-shard", "duplicate-shard", "source", "platform", "lane", "inventory", "selector", "exit", "lost-test", "duplicate-test", "missing-package", "unexpected-package", "failed-package", "slow", "future-test"} {
		t.Run(plant, func(t *testing.T) {
			tests := windowsFixtureInventory()
			parts, err := windowsPartition(tests)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			for i := 0; i < windowsShardCount; i++ {
				path := filepath.Join(dir, fmt.Sprint(i))
				if err := os.Mkdir(path, 0755); err != nil {
					t.Fatal(err)
				}
				m := windowsManifest{Commit: "head", GOOS: "windows", Lane: "windows-test", Shard: i, Inventories: tests, Selector: windowsSelector(parts, i)}
				data, err := json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				for name, body := range map[string][]byte{"manifest.json": data, "go-test.json": windowsFixtureStream(parts, i)} {
					if err := os.WriteFile(filepath.Join(path, name), body, 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			path := filepath.Join(dir, "0", "manifest.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var m windowsManifest
			if err := json.Unmarshal(data, &m); err != nil {
				t.Fatal(err)
			}
			stream := string(windowsFixtureStream(parts, 0))
			want := "source/platform/lane/inventory/selector/exit mismatch"
			switch plant {
			case "missing-shard":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				want = "want 4 Windows manifests, got 3"
			case "duplicate-shard":
				m.Shard = 1
				m.Selector = windowsSelector(parts, 1)
				stream = string(windowsFixtureStream(parts, 1))
				want = "invalid or duplicate Windows shard"
			case "source":
				m.Commit = "other"
			case "platform":
				m.GOOS = "darwin"
			case "lane":
				m.Lane = "race"
			case "inventory":
				delete(m.Inventories, windowsPackages[0])
			case "selector":
				m.Selector = "^TestShared$"
			case "exit":
				m.ExitCode = 1
			case "future-test":
				tests[windowsPackages[0]] = append(tests[windowsPackages[0]], "TestAddedAfterArtifacts")
			case "lost-test":
				stream = strings.Join(strings.Split(stream, "\n")[1:], "\n")
				want = "missing completed test"
			case "duplicate-test":
				stream = strings.Split(stream, "\n")[0] + "\n" + stream
				want = "unexpected or duplicate test"
			case "missing-package":
				var kept []string
				for _, line := range strings.Split(stream, "\n") {
					if !strings.Contains(line, windowsPackages[0]) {
						kept = append(kept, line)
					}
				}
				stream = strings.Join(kept, "\n")
				want = "missing completed test"
			case "unexpected-package":
				stream = strings.ReplaceAll(stream, windowsPackages[0], "other")
				want = "unexpected Windows package"
			case "failed-package":
				stream = strings.Replace(stream,
					fmt.Sprintf(`{"Action":"pass","Package":%q,"Elapsed":42}`, windowsPackages[0]),
					fmt.Sprintf(`{"Action":"fail","Package":%q,"Elapsed":42}`, windowsPackages[0]), 1)
				want = "failed test/package"
			case "slow":
				stream = strings.Replace(stream, `"Elapsed":42`, `"Elapsed":3600`, 1)
				want = "outside (0,3600)s margin"
			}
			if plant != "missing-shard" {
				data, err = json.Marshal(m)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "0", "go-test.json"), []byte(stream), 0644); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(t.TempDir(), "merged.json")
			err = mergeWindowsEvidence(dir, out, tests, "head")
			if plant == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				packages, err := os.ReadFile(out + ".packages")
				if err != nil || string(packages) != strings.Join(windowsPackages, "\n")+"\n" {
					t.Fatalf("admitted package list: %s, %v", packages, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v, want %s", err, want)
			}
			for _, name := range []string{out, out + ".packages"} {
				if _, err := os.Stat(name); !os.IsNotExist(err) {
					t.Fatalf("refused evidence published %s", name)
				}
			}
		})
	}
}
