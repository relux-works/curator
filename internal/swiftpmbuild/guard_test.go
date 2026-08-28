package swiftpmbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The build stage owns no process boundary of its own: every child crosses the
// shared commit-before-start executor seam. This extends the cross-adapter
// guard to the new SwiftPM build surface without widening its allowlist.
func TestBuildAdapterStartsNoProcessOutsideTheSharedSeams(t *testing.T) {
	sharedProcessSeams := map[string]bool{"acquisition.go": true, "portable_runner.go": true}
	local, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	interop, err := filepath.Glob(filepath.Join("..", "swiftpminterop", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Glob(filepath.Join("..", "swiftpmsource", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	shared, err := filepath.Glob(filepath.Join("..", "closureexec", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	inspected := 0
	for _, name := range append(append(append(local, interop...), source...), shared...) {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		payload, readErr := os.ReadFile(name) // #nosec G304 -- fixed package-relative production source glob.
		if readErr != nil {
			t.Fatal(readErr)
		}
		inspected++
		if !strings.Contains(string(payload), "exec.Command") {
			continue
		}
		if sharedProcessSeams[filepath.Base(name)] {
			continue
		}
		t.Fatalf("production SwiftPM surface %s bypasses the instrumented process boundary", name)
	}
	if inspected < 15 {
		t.Fatalf("guard inspected only %d production files", inspected)
	}
	for _, name := range local {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		payload, readErr := os.ReadFile(name) // #nosec G304 -- fixed package-relative production source glob.
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(payload), `"os/exec"`) {
			t.Fatalf("build production source %s imports os/exec", name)
		}
	}
}
