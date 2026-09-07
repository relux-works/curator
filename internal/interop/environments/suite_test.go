// Package environments consumes the environments families of the authoritative
// external Curator Protocol suite. It contains no implementation-owned expected
// values.
//
// It is a package of its own, and not part of `internal/interop`, so that
// `.github/ci/root-artifacts.tsv` can declare the environments families these
// cases read WITHOUT deferring the pre-environments conformance cases that live
// in `internal/interop` and that the default lane runs against the committed
// SPEC_PIN root.
//
// The consequence of that declaration is the whole point of the split:
//
//   - a root that publishes every declared family SERVES this package, and
//     every case here runs;
//   - a root that publishes none of them DEFERS the whole package, `suiteRoot`
//     takes the `root is not set` path, and the deferral is named in the suite
//     plan -- fatal in a lane that sets `CI_REQUIRE_FULL_ROOT=1`;
//   - a root that publishes SOME of them is the case this package exists to
//     make loud: the plan names the missing family and the candidate lane
//     refuses to run at all.
//
// Nothing here may reintroduce a per-family skip. A family the suite plan
// declared present, that a case then cannot read, is a contradiction, and
// `requireFamily` reports it as a failure.
package environments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// suiteRoot resolves the conformance root, or skips because the suite plan
// deferred this package. This is the ONLY legitimate skip in the package: it is
// classified `root-unset`, which `.github/ci/skip-classes.tsv` admits only for
// a package `suite-plan.sh` actually deferred, so it cannot fire unnoticed in a
// lane that requires a fully serving root.
func suiteRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(absolute, "manifest.json")); err != nil {
		t.Fatalf("invalid CURATOR_CONFORMANCE_ROOT: %v", err)
	}
	return absolute
}

// requireFamily reads a declared root artefact, and FAILS rather than skips
// when it is absent. `.github/ci/root-artifacts.tsv` declares every family
// named here, so a root that serves this package and then does not publish one
// has already contradicted the suite plan. Treating that as a skip is exactly
// the silent shrink this package was split out to prevent.
func requireFamily(t *testing.T, root, family string) []byte {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(family)))
	if err != nil {
		t.Fatalf("conformance root %s serves this package but does not publish %s: %v\n"+
			"\t%s is declared for internal/interop/environments in .github/ci/root-artifacts.tsv,\n"+
			"\tso a root that omits it must be DEFERRED by .github/ci/suite-plan.sh -- never silently skipped",
			root, family, err, family)
	}
	return payload
}

// readRootFile returns the trimmed contents of a file the root publishes.
func readRootFile(t *testing.T, root, rel string) string {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(payload))
}
