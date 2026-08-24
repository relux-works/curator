package crossconformance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/npmsource"
	"github.com/relux-works/curator/internal/rustsource"
	"github.com/relux-works/curator/internal/swiftpmbuild"
	"github.com/relux-works/curator/internal/yarnclassicsource"
)

// The integration package owns no process boundary of its own. This extends
// the cross-adapter guard to the new conformance surface without widening its
// allowlist: the only production files that may launch a child process are
// the two shared commit-before-start executor seams.
func TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams(t *testing.T) {
	sharedProcessSeams := map[string]bool{"acquisition.go": true, "portable_runner.go": true}
	local, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	shared, err := filepath.Glob(filepath.Join("..", "closureexec", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	// The needles are assembled at run time so this guard can scan its own
	// source as strictly as it scans everything else.
	launchNeedle := "exec" + ".Command"
	importNeedle := `"os/` + `exec"`
	inspected := 0
	for _, name := range append(local, shared...) {
		payload, readErr := os.ReadFile(name) // #nosec G304 -- fixed package-relative source glob.
		if readErr != nil {
			t.Fatal(readErr)
		}
		inspected++
		if !strings.Contains(string(payload), launchNeedle) {
			continue
		}
		if sharedProcessSeams[filepath.Base(name)] {
			continue
		}
		t.Fatalf("%s bypasses the instrumented process boundary", name)
	}
	if inspected < 15 {
		t.Fatalf("guard inspected only %d files", inspected)
	}
	for _, name := range local {
		payload, readErr := os.ReadFile(name) // #nosec G304 -- fixed package-relative source glob.
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(payload), importNeedle) {
			t.Fatalf("integration source %s imports os/exec", name)
		}
	}
}

// The integration package's production files must depend on nothing but the
// standard library. An oracle that imported the encoder it checks, or an
// adapter it integrates, would be checking itself.
func TestIntegrationProductionSourceImportsNoRepositoryPackage(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	production := 0
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		payload, readErr := os.ReadFile(name) // #nosec G304 -- fixed package-relative source glob.
		if readErr != nil {
			t.Fatal(readErr)
		}
		production++
		if strings.Contains(string(payload), "relux-works/curator/internal/") {
			t.Fatalf("%s imports a repository package; the oracle must stay independent", name)
		}
	}
	if production < 4 {
		t.Fatalf("guard inspected only %d production files", production)
	}
}

// A vector delegated to an owning package must still name a diagnostic that
// package actually declares. These are compile-time references, so a renamed
// or deleted constant breaks the build rather than quietly emptying the
// published matrix.
func TestDelegatedDiagnosticsAreDeclaredByTheirOwners(t *testing.T) {
	// Each entry is one owning package's own constant, referenced at compile
	// time. Several owners legitimately declare the same stable code: that is
	// the point of a shared vocabulary.
	owned := []struct{ owner, code string }{
		{"internal/swiftpmbuild", string(swiftpmbuild.CodeNetworkAttempted)},
		{"internal/swiftpmbuild", string(swiftpmbuild.CodeWriteUndeclared)},
		{"internal/swiftpmbuild", string(swiftpmbuild.CodeInputUndeclared)},
		{"internal/swiftpmbuild", string(swiftpmbuild.CodeProcessUndeclared)},
		{"internal/npmsource", npmsource.CodeNetworkAttempted},
		{"internal/yarnclassicsource", yarnclassicsource.CodeNetworkAttempted},
		{"internal/artifactpolicy", string(artifactpolicy.CodeLocalOutputDrift)},
		{"internal/artifactpolicy", string(artifactpolicy.CodeLocalOutputUnreceipted)},
		{"internal/rustsource", string(rustsource.CodeUndeclaredInput)},
		{"internal/closuregraph", string(closuregraph.CodeBuildCycle)},
		{"internal/closuregraph", string(closuregraph.CodeGraphReferenceInvalid)},
	}
	declared := map[string]bool{}
	for _, entry := range owned {
		if entry.code == "" {
			t.Fatalf("%s declares an empty diagnostic constant", entry.owner)
		}
		declared[entry.code] = true
	}
	for _, vector := range crossconformance.RejectionVectors() {
		if vector.CrossDrivable() {
			continue
		}
		covered := false
		for _, code := range vector.Codes {
			covered = covered || declared[code]
		}
		if !covered {
			t.Errorf("delegated vector %s names no diagnostic its owners declare: %v", vector.ID, vector.Codes)
		}
	}
	// The shared executor is the single boundary that reports a network
	// attempt for every adapter, so the delegation points at one place.
	if _, err := closureexec.NewCaptureStore(filepath.Join(t.TempDir(), "guard")); err != nil {
		t.Fatal(err)
	}
}
