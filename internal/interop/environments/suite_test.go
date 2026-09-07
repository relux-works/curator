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
//   - a root that publishes every declared artefact SERVES this package, and
//     every case here runs;
//   - a root that publishes none of them DEFERS the whole package, `suiteRoot`
//     takes the `root is not set` path, and the deferral is named in the suite
//     plan -- fatal in a lane that sets `CI_REQUIRE_FULL_ROOT=1`;
//   - a root that publishes SOME of them is the case this package exists to
//     make loud: the plan names the missing artefact and the candidate lane
//     refuses to run at all.
//
// Nothing here may reintroduce a per-family skip. An artefact the suite plan
// declared present, that a case then cannot read, is a contradiction, and
// `requireFamily` reports it as a failure.
//
// Every root path this package reads goes through `rootPath`, which REFUSES a
// path no declared artefact covers. That is what keeps the three bullets above
// true as the cases change: a case that starts reading a new root path either
// declares it -- and inherits the deferral behaviour -- or fails immediately,
// naming the table it has to be added to. `contract_test.go` proves the guard
// cannot be walked around.
package environments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The CI tables this package's fail-closed behaviour is built out of. They are
// read as committed files, never mirrored into a constant here: a copy would go
// stale exactly when the tables changed, which is the moment these cases exist
// to catch.
const (
	rootArtefacts = "../../../.github/ci/root-artifacts.tsv"
	thisPackage   = "internal/interop/environments"
)

// suiteRoot resolves the conformance root, or skips because the suite plan
// deferred this package. This is the ONLY skip in the package: it is classified
// `root-unset`, which `.github/ci/skip-classes.tsv` admits only for a package
// `suite-plan.sh` actually deferred, so it cannot fire unnoticed in a lane that
// requires a fully serving root.
//
// No hosted lane reaches it any more. The committed SPEC_PIN publishes every
// declared artefact, so the plan serves this package everywhere and
// `.github/ci/platform-cases.tsv` tolerates no skip of these cases at all --
// the guard survives for a local `go test ./...` run with no root exported, and
// is fatal wherever the gate is what runs.
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

// declaredArtefacts returns the artefact set root-artifacts.tsv declares for a
// package, and whether the package has a row at all. An absent row is the state
// this whole package exists to prevent, so it is reported rather than defaulted.
func declaredArtefacts(t *testing.T, pkg string) ([]string, bool) {
	t.Helper()
	payload, err := os.ReadFile(rootArtefacts)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(payload), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 || fields[0] != pkg {
			continue
		}
		return strings.Split(fields[1], ","), true
	}
	return nil, false
}

// coveringArtefact reports which declared artefact covers a root-relative path:
// the artefact itself when it names a file, or the tree it sits under.
func coveringArtefact(declared []string, rel string) (string, bool) {
	for _, artefact := range declared {
		if rel == artefact || strings.HasPrefix(rel, artefact+"/") {
			return artefact, true
		}
	}
	return "", false
}

// rootPath is the ONLY place this package turns a root-relative path into a
// path on disk, and it refuses one that `.github/ci/root-artifacts.tsv` does
// not declare for this package.
//
// The declaration is what makes a missing artefact loud: `suite-plan.sh` checks
// exactly the declared set against the supplied root, defers the package when
// one is absent, and a `CI_REQUIRE_FULL_ROOT` lane fails there BY NAME. A path
// read outside that set is invisible to the plan -- the root can drop it, the
// package is still served, and the case fails mid-run on an open() error at
// best. Checking at the moment of the read, against the committed table rather
// than against a list kept beside it, is what stops a new read from drifting
// undeclared.
func rootPath(t *testing.T, root, rel string) string {
	t.Helper()
	declared, ok := declaredArtefacts(t, thisPackage)
	if !ok {
		t.Fatalf("%s declares no artefacts for %s.\n"+
			"\tWithout a row, suite-plan.sh serves this package against ANY root and a\n"+
			"\tcandidate that dropped an environments family would pass the lane green.",
			rootArtefacts, thisPackage)
	}
	if _, ok := coveringArtefact(declared, rel); !ok {
		t.Fatalf("this package reads %s out of the conformance root, but %s declares no artefact covering it for %s.\n"+
			"\tdeclared: %s\n"+
			"\tsuite-plan.sh checks only the declared set, so a root that dropped %s would still SERVE\n"+
			"\tthis package: the candidate lane could not fail by name. Declare it -- and expect\n"+
			"\tTestConformanceEveryPathTheVectorsNameIsDeclared to hold you to reading it.",
			rel, rootArtefacts, thisPackage, strings.Join(declared, " "), rel)
	}
	return filepath.Join(root, filepath.FromSlash(rel))
}

// rootHasPath reports whether the root publishes a path.
//
// It is the one root access that is deliberately NOT subject to the artefact
// declaration, and it exists so the declaration can be DERIVED rather than
// restated: TestConformanceEveryPathTheVectorsNameIsDeclared walks the vectors'
// own path fields and asks the root which of them resolve. It stats and never
// returns bytes, so it cannot be used to read an undeclared artefact.
func rootHasPath(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

// rootPathsNamedIn collects every string anywhere in a decoded vector that the
// supplied root resolves as a path. Resolution against the root -- rather than
// a list of field names -- is what makes this a derivation: a vector field the
// cases start consuming is collected the moment the root publishes it.
func rootPathsNamedIn(document any, root string) []string {
	found := map[string]bool{}
	var walk func(node any)
	walk = func(node any) {
		switch value := node.(type) {
		case map[string]any:
			for _, item := range value {
				walk(item)
			}
		case []any:
			for _, item := range value {
				walk(item)
			}
		case string:
			if strings.Contains(value, "/") && !strings.HasPrefix(value, "/") && rootHasPath(root, value) {
				found[value] = true
			}
		}
	}
	walk(document)
	return sortedKeys(found)
}

// requireFamily reads a declared root artefact, and FAILS rather than skips
// when it is absent. `.github/ci/root-artifacts.tsv` declares every family
// named here, so a root that serves this package and then does not publish one
// has already contradicted the suite plan. Treating that as a skip is exactly
// the silent shrink this package was split out to prevent.
func requireFamily(t *testing.T, root, family string) []byte {
	t.Helper()
	payload, err := os.ReadFile(rootPath(t, root, family))
	if err != nil {
		t.Fatalf("conformance root %s serves this package but does not publish %s: %v\n"+
			"\t%s is declared for internal/interop/environments in .github/ci/root-artifacts.tsv,\n"+
			"\tso a root that omits it must be DEFERRED by .github/ci/suite-plan.sh -- never silently skipped",
			root, family, err, family)
	}
	return payload
}

// readRootFile returns the trimmed contents of a declared root path. Like
// requireFamily it fails rather than skips: the path is covered by a declared
// artefact, so a root that serves this package and cannot produce it has
// contradicted the plan.
func readRootFile(t *testing.T, root, rel string) string {
	t.Helper()
	payload, err := os.ReadFile(rootPath(t, root, rel))
	if err != nil {
		t.Fatalf("conformance root %s serves this package but cannot produce %s: %v\n"+
			"\tit is covered by an artefact declared for internal/interop/environments in\n"+
			"\t.github/ci/root-artifacts.tsv, so a root that omits it must be DEFERRED, never skipped",
			root, rel, err)
	}
	return strings.TrimSpace(string(payload))
}
