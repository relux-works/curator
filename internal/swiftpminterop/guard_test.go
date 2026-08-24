package swiftpminterop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The interop stage owns no process boundary at all: it starts nothing and
// derives observed reads only through the manager-owned seam. The two shared
// seams under internal/closureexec remain the only allowlisted crossings, and
// this adapter surface is now covered by the same guard.
func TestInteropAdapterStartsNoProcessOutsideTheSharedSeams(t *testing.T) {
	sharedProcessSeams := map[string]bool{"acquisition.go": true, "portable_runner.go": true}
	local, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := filepath.Glob(filepath.Join("..", "swiftpmsource", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	shared, err := filepath.Glob(filepath.Join("..", "closureexec", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	inspected := 0
	for _, name := range append(append(local, upstream...), shared...) {
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
		t.Fatalf("production interop/source surface %s bypasses the instrumented process boundary", name)
	}
	if inspected < 10 {
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
			t.Fatalf("interop production source %s imports os/exec", name)
		}
	}
}
