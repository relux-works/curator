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
	for i, p := range ordinaryPackages {
		for j := 0; j < 80; j++ {
			all[p] = append(all[p], fmt.Sprintf("TestP%dCase%03d", i, j))
		}
		all[p] = append(all[p], "TestShared", "ExampleFuture", "FuzzFuture")
	}
	return all
}

// One common selector must select every native test once in its own package,
// including future tests and names shared by packages, with no cross-selection.
func testOrdinaryPartitionCoverage(t *testing.T, c testLane) {
	tests := windowsFixtureInventory()
	parts, err := c.partition(tests)
	if err != nil {
		t.Fatal(err)
	}
	for p, names := range tests {
		for _, name := range names {
			count := 0
			for i := 0; i < c.shards; i++ {
				matched := regexp.MustCompile(c.selector(parts, i)).MatchString(name)
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
	delete(tests, ordinaryPackages[0])
	if _, err := c.partition(tests); err == nil || !strings.Contains(err.Error(), "inventory mismatch") {
		t.Fatalf("missing package admitted: %v", err)
	}
	tests = windowsFixtureInventory()
	tests[ordinaryPackages[0]] = append(tests[ordinaryPackages[0]], "TestShared")
	if _, err := c.partition(tests); err == nil || !strings.Contains(err.Error(), "duplicate test") {
		t.Fatalf("duplicate test admitted: %v", err)
	}
}

func windowsFixtureStream(parts map[string][][]string, shard int) []byte {
	var out strings.Builder
	for _, p := range ordinaryPackages {
		for _, name := range parts[p][shard] {
			fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Package\":%q,\"Test\":%q}\n", p, name)
		}
		fmt.Fprintf(&out, "{\"Action\":\"pass\",\"Package\":%q,\"Elapsed\":42}\n", p)
	}
	return []byte(out.String())
}

// Complete positive evidence is admitted; each bounded plant must fail for
// its intended inventory, execution, source or margin violation, without outputs.
func testOrdinaryMergeControls(t *testing.T, c testLane) {
	for _, plant := range []string{"valid", "missing-shard", "duplicate-shard", "source", "platform", "lane", "inventory", "selector", "exit", "lost-test", "duplicate-test", "missing-package", "unexpected-package", "failed-package", "slow", "future-test"} {
		t.Run(plant, func(t *testing.T) {
			tests := windowsFixtureInventory()
			parts, err := c.partition(tests)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			for i := 0; i < c.shards; i++ {
				path := filepath.Join(dir, fmt.Sprint(i))
				if err := os.Mkdir(path, 0755); err != nil {
					t.Fatal(err)
				}
				m := testManifest{Commit: "head", GOOS: c.goos, Lane: c.lane, Shard: i, Inventories: tests, Selector: c.selector(parts, i)}
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
			var m testManifest
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
				want = fmt.Sprintf("want %d ordinary manifests, got %d", c.shards, c.shards-1)
			case "duplicate-shard":
				m.Shard = 1
				m.Selector = c.selector(parts, 1)
				stream = string(windowsFixtureStream(parts, 1))
				want = "invalid or duplicate ordinary shard"
			case "source":
				m.Commit = "other"
			case "platform":
				m.GOOS = "unsupported"
			case "lane":
				m.Lane = "race"
			case "inventory":
				delete(m.Inventories, ordinaryPackages[0])
			case "selector":
				m.Selector = "^TestShared$"
			case "exit":
				m.ExitCode = 1
			case "future-test":
				tests[ordinaryPackages[0]] = append(tests[ordinaryPackages[0]], "TestAddedAfterArtifacts")
			case "lost-test":
				stream = strings.Join(strings.Split(stream, "\n")[1:], "\n")
				want = "missing completed test"
			case "duplicate-test":
				stream = strings.Split(stream, "\n")[0] + "\n" + stream
				want = "unexpected or duplicate test"
			case "missing-package":
				var kept []string
				for _, line := range strings.Split(stream, "\n") {
					if !strings.Contains(line, ordinaryPackages[0]) {
						kept = append(kept, line)
					}
				}
				stream = strings.Join(kept, "\n")
				want = "missing completed test"
			case "unexpected-package":
				stream = strings.ReplaceAll(stream, ordinaryPackages[0], "other")
				want = "unexpected ordinary package"
			case "failed-package":
				stream = strings.Replace(stream,
					fmt.Sprintf(`{"Action":"pass","Package":%q,"Elapsed":42}`, ordinaryPackages[0]),
					fmt.Sprintf(`{"Action":"fail","Package":%q,"Elapsed":42}`, ordinaryPackages[0]), 1)
				want = "failed test/package"
			case "slow":
				stream = strings.Replace(stream, `"Elapsed":42`, fmt.Sprintf(`"Elapsed":%d`, c.margin), 1)
				want = fmt.Sprintf("outside (0,%d)s margin", c.margin)
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
			err = c.merge(dir, out, tests, "head")
			if plant == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				packages, err := os.ReadFile(out + ".packages")
				if err != nil || string(packages) != strings.Join(ordinaryPackages, "\n")+"\n" {
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

// Each native configuration uses the same union engine and bounded refusal
// controls, while retaining its measured shard count and package deadline.
func TestOrdinaryNativeLanes(t *testing.T) {
	for _, tc := range []struct {
		lane, goos, timeout string
		shards, margin      int
	}{
		{"windows-test", "windows", "120m", 4, 3600},
		{"unix-test", "linux", "60m", 3, 1800},
		{"unix-test", "darwin", "60m", 2, 1800},
	} {
		t.Run(tc.goos, func(t *testing.T) {
			c, err := ordinaryConfig(tc.lane, tc.goos)
			if err != nil || c.shards != tc.shards || c.timeout != tc.timeout || c.margin != tc.margin {
				t.Fatalf("native lane drift: %+v, %v", c, err)
			}
			t.Run("partition", func(t *testing.T) { testOrdinaryPartitionCoverage(t, c) })
			t.Run("admission", func(t *testing.T) { testOrdinaryMergeControls(t, c) })
		})
	}
	for _, tc := range [][2]string{{"unix-test", "windows"}, {"windows-test", "linux"}, {"windows-test", "darwin"}, {"unix-test", "freebsd"}, {"other", "linux"}} {
		if _, err := ordinaryConfig(tc[0], tc[1]); err == nil || !strings.Contains(err.Error(), "does not support native") {
			t.Fatalf("unsupported lane/platform admitted: %v, %v", tc, err)
		}
	}
}
