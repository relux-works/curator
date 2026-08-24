//go:build darwin || linux

package swiftpmbuild

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/swiftpminterop"
)

// One real tool vector keeps the planned command honest. The adapter tests own
// policy; this smoke proves that the exact planned argv still reaches SwiftPM,
// that the scratch directory reproduces SwiftPM's triple naming, and that the
// per-source object layout the build stage reconciles is the real one.
func TestRealSwiftPMBuildMatchesThePlannedLayout(t *testing.T) {
	swift, err := exec.LookPath("swift")
	if err != nil {
		t.Skip("swift toolchain is not installed")
	}
	root := t.TempDir()
	files := map[string]string{
		"Package.swift": "// swift-tools-version: 5.9\nimport PackageDescription\n" +
			"let package = Package(name: \"Fixture\", products: [.executable(name: \"fixture\", targets: [\"App\"])], " +
			"targets: [.target(name: \"CLib\"), .executableTarget(name: \"App\", dependencies: [\"CLib\"])])\n",
		"Package.resolved":             `{"version":3,"pins":[]}`,
		"Sources/CLib/include/CLib.h":  "#ifndef CLIB_H\n#define CLIB_H\nint first(void);\nint second(void);\nint left(void);\nint right(void);\n#endif\n",
		"Sources/CLib/first.c":         "#include \"CLib.h\"\nint first(void) { return 1; }\n",
		"Sources/CLib/nested/second.c": "#include \"CLib.h\"\nint second(void) { return 2; }\n",
		// Two sources of one Clang target that share a base name in different
		// subdirectories: SwiftPM mirrors the tree, so both objects must
		// resolve against the target-relative source path.
		"Sources/CLib/a/shared.c":  "#include \"CLib.h\"\nint left(void) { return 3; }\n",
		"Sources/CLib/b/shared.c":  "#include \"CLib.h\"\nint right(void) { return 4; }\n",
		"Sources/App/main.swift":   "import CLib\nprint(first() + second() + left() + right())\n",
		"Sources/App/helper.swift": "let helper = 3\n",
	}
	for relative, payload := range files {
		full := filepath.Join(root, filepath.FromSlash(relative))
		if err = os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(full, []byte(payload), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	triple := hostTriple(t, swift)
	argv := plannedBuildArgv(fixtureConfiguration, triple, "fixture")
	process := exec.CommandContext(t.Context(), swift, argv...) // #nosec G204 -- resolved test tool and production permit argv.
	process.Dir = root
	process.Env = []string{"HOME=" + filepath.Join(root, "empty-home"), "PATH=" + filepath.Dir(swift), "TZ=UTC"}
	if output, runErr := process.CombinedOutput(); runErr != nil {
		t.Fatalf("real SwiftPM build with the planned argv: %v: %s", runErr, output)
	}

	scratch := swiftpmScratchDirectory(triple, fixtureConfiguration)
	product := filepath.Join(root, filepath.FromSlash(path.Join(scratch, "fixture")))
	if info, statErr := os.Stat(product); statErr != nil || !info.Mode().IsRegular() {
		t.Fatalf("planned product path %q is not the real one: %v", product, statErr)
	}
	for target, sources := range map[string][]string{
		"CLib": {"Sources/CLib/a/shared.c", "Sources/CLib/b/shared.c", "Sources/CLib/first.c", "Sources/CLib/nested/second.c"},
		"App":  {"Sources/App/main.swift", "Sources/App/helper.swift"},
	} {
		directory := filepath.Join(root, filepath.FromSlash(scratch), target+".build")
		objects, listErr := collectObjectFiles(directory)
		if listErr != nil {
			t.Fatalf("target %s build directory: %v", target, listErr)
		}
		if len(objects) != len(sources) {
			t.Fatalf("target %s produced %v, want one object per source %v", target, objects, sources)
		}
		resolved := map[string]bool{}
		for _, source := range sources {
			slot := ObjectSlot{Package: "fixture", Target: target, Source: source, SourceRoot: "Sources/" + target, Kind: "clang"}
			match, resolveErr := resolveProducedObject(slot, objects)
			if resolveErr != nil {
				t.Fatalf("declared object for %s did not resolve to a real produced object: %v", source, resolveErr)
			}
			if resolved[match] {
				t.Fatalf("two declared sources of %s resolved to one produced object %s", target, match)
			}
			resolved[match] = true
		}
		if len(resolved) != len(objects) {
			t.Fatalf("target %s resolved %d of %d produced objects", target, len(resolved), len(objects))
		}
	}
	dependencies := []string{}
	_ = filepath.WalkDir(filepath.Join(root, filepath.FromSlash(scratch)), func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr == nil && !entry.IsDir() && strings.HasSuffix(entry.Name(), ".d") {
			dependencies = append(dependencies, current)
		}
		return nil
	})
	if len(dependencies) == 0 {
		t.Fatal("real SwiftPM build emitted no compiler dependency file for the observed read set")
	}
	sort.Strings(dependencies)
	payload, err := os.ReadFile(dependencies[0]) // #nosec G304 -- test-owned scratch tree.
	if err != nil {
		t.Fatal(err)
	}
	if len(parseDependencyFile(string(payload))) == 0 {
		t.Fatalf("dependency grammar produced no observed read from %s", dependencies[0])
	}
	var _ swiftpminterop.ObservedRead
}

func hostTriple(t *testing.T, swift string) string {
	t.Helper()
	process := exec.CommandContext(t.Context(), swift, "-print-target-info") // #nosec G204 -- resolved test tool.
	output, err := process.Output()
	if err != nil {
		t.Skipf("swift -print-target-info unavailable: %v", err)
	}
	marker := `"unversionedTriple": "`
	index := strings.Index(string(output), marker)
	if index < 0 {
		t.Skip("swift target info has no unversioned triple")
	}
	rest := string(output)[index+len(marker):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		t.Skip("swift target info triple is malformed")
	}
	return rest[:end]
}
