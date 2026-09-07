package environments

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

// The CI tables this package's fail-closed behaviour is built out of. They are
// read as committed files, never mirrored into a constant here: a copy would go
// stale exactly when the tables changed, which is the moment these cases exist
// to catch.
const (
	rootArtefacts = "../../../.github/ci/root-artifacts.tsv"
	platformCases = "../../../.github/ci/platform-cases.tsv"
	thisPackage   = "internal/interop/environments"
)

// crossReferencedTrees are declared in root-artifacts.tsv but are not read by a
// `requireFamily` call: the vectors name individual files inside them, and the
// cases read those paths straight out of the vector. `vectors/environments.json`
// names every `expected/environments/...` surface it compares byte for byte, and
// `vectors/snapshot-acquisition.json` names the `fixtures/byte-exact` tree it
// commits and re-extracts. A root serving the vector but not the tree would fail
// mid-case with an open() error instead of failing the lane by name, so both
// trees are declared alongside the vectors.
var crossReferencedTrees = []string{"expected/environments", "fixtures/byte-exact"}

// packageFiles parses every Go file beside this case.
//
// The two scans below work on the SYNTAX TREE and never on the file text. A
// text scan is fooled by the obvious evasion -- bind the *testing.T to another
// name and call Skipf on that -- while still matching its own source, comments
// and string literals. The parser sees a call expression either way.
func packageFiles(t *testing.T) (*token.FileSet, map[string]*ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]*ast.File{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry.Name(), nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		files[entry.Name()] = file
	}
	if len(files) == 0 {
		t.Fatal("no Go sources found beside this case; the scans below would pass vacuously")
	}
	return fset, files
}

// skipCall reports the receiver and method of a `<ident>.Skip*(...)` call.
func skipCall(node ast.Node) (recv, method string, ok bool) {
	call, isCall := node.(*ast.CallExpr)
	if !isCall {
		return "", "", false
	}
	sel, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return "", "", false
	}
	switch sel.Sel.Name {
	case "Skip", "Skipf", "SkipNow":
	default:
		return "", "", false
	}
	ident, isIdent := sel.X.(*ast.Ident)
	if !isIdent {
		return "", "", false
	}
	return ident.Name, sel.Sel.Name, true
}

// requiredFamily reports the family literal of a `requireFamily(_, _, "x")` call.
func requiredFamily(node ast.Node) (string, bool) {
	call, isCall := node.(*ast.CallExpr)
	if !isCall {
		return "", false
	}
	ident, isIdent := call.Fun.(*ast.Ident)
	if !isIdent || ident.Name != "requireFamily" || len(call.Args) != 3 {
		return "", false
	}
	lit, isLit := call.Args[2].(*ast.BasicLit)
	if !isLit || lit.Kind != token.STRING {
		return "", false
	}
	return strings.Trim(lit.Value, `"`), true
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

// TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts is the recurrence
// guard for the hole this package was split out to close.
//
// `.github/ci/suite-plan.sh` can only defer this package -- and a
// CI_REQUIRE_FULL_ROOT lane can only fail by name -- for an artefact
// root-artifacts.tsv actually declares. A case that starts reading a sixth
// family without declaring it puts that family back in the silent-skip state:
// a candidate root could drop it, the plan would still serve the package, and
// the case would fail on an open() error at best or skip at worst.
//
// So the declared set must be EXACTLY the set the cases read, plus the trees
// the vectors cross-reference. Not a superset (a declared artefact nothing
// reads defers the package for no reason) and not a subset.
func TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts(t *testing.T) {
	_, files := packageFiles(t)
	read := map[string]bool{}
	for name, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			if family, ok := requiredFamily(node); ok {
				t.Logf("%s requires %s", name, family)
				read[family] = true
			}
			return true
		})
	}
	if len(read) == 0 {
		t.Fatal("no requireFamily call found; either the cases stopped reading the root or the scan broke")
	}

	declared, ok := declaredArtefacts(t, thisPackage)
	if !ok {
		t.Fatalf("%s declares no artefacts for %s.\n"+
			"\tWithout a row, suite-plan.sh serves this package against ANY root and a\n"+
			"\tcandidate that dropped an environments family would pass the lane green.",
			rootArtefacts, thisPackage)
	}

	want := map[string]bool{}
	for family := range read {
		want[family] = true
	}
	for _, tree := range crossReferencedTrees {
		want[tree] = true
	}
	got := map[string]bool{}
	for _, artefact := range declared {
		got[artefact] = true
	}

	for artefact := range want {
		if !got[artefact] {
			t.Errorf("%s is read by this package but NOT declared in %s for %s.\n"+
				"\tA root that drops it would still SERVE this package, so the candidate lane\n"+
				"\twould not fail by name -- exactly the hole the split closed.",
				artefact, rootArtefacts, thisPackage)
		}
	}
	for artefact := range got {
		if !want[artefact] {
			t.Errorf("%s is declared in %s for %s but nothing here reads it.\n"+
				"\tAn artefact nothing reads defers the package on roots that could have served it.",
				artefact, rootArtefacts, thisPackage)
		}
	}
	t.Logf("declared=%s", strings.Join(sortedKeys(got), " "))
}

// TestNoCaseHereSkipsForAnythingButTheDeferredRoot keeps the package's single
// legitimate skip single.
//
// A per-family `t.Skipf("... publishes no vectors/x.json ...")` classifies as
// `root-content` in `.github/ci/skip-classes.tsv`, which is policy `allow` in
// EVERY lane including the candidate one. One such guard anywhere in this
// package reopens the hole for that family. The only skip allowed here is the
// `root-unset` one `suiteRoot` raises, which skip-classes.tsv admits solely for
// a package suite-plan.sh actually deferred.
func TestNoCaseHereSkipsForAnythingButTheDeferredRoot(t *testing.T) {
	fset, files := packageFiles(t)
	skips := 0
	for name, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			recv, method, ok := skipCall(node)
			if !ok {
				return true
			}
			position := fset.Position(node.Pos())
			// The one legitimate skip: suiteRoot reporting that the suite plan
			// deferred this package. Recognised by WHERE it is, not by the text
			// of the line, so renaming the receiver cannot smuggle another one in.
			if strings.HasSuffix(name, "suite_test.go") && enclosingFunc(file, node) == "suiteRoot" {
				skips++
				t.Logf("the deferred-root skip: %s:%d %s.%s", position.Filename, position.Line, recv, method)
				return true
			}
			t.Errorf("%s:%d calls %s.%s outside suiteRoot.\n"+
				"\tEvery family here is declared in %s, so a missing one must DEFER the package\n"+
				"\tand fail a CI_REQUIRE_FULL_ROOT lane by name. A skip makes the lane pass green\n"+
				"\tinstead -- a per-family root-content skip is policy `allow` in EVERY lane.",
				position.Filename, position.Line, recv, method, rootArtefacts)
			return true
		})
	}
	if skips != 1 {
		t.Errorf("expected exactly one skip in suiteRoot, found %d", skips)
	}
}

// TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass checks the column
// that decides whether a skip of these cases is survivable.
//
// `.github/ci/platform-case-gate.sh` tolerates a skip of a ledger case only when
// the reason carries the class the row declares. `root-unset` is policy
// `deferred-only`: legitimate solely for a package suite-plan.sh deferred.
// `root-content` is policy `allow` everywhere -- a row carrying it lets these
// cases skip in the candidate lane and pass. That is what the rows said before
// the split, and it is the single character of this ledger a regression would
// most plausibly touch.
func TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass(t *testing.T) {
	payload, err := os.ReadFile(platformCases)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]string{}
	for _, line := range strings.Split(string(payload), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 5 || fields[0] != thisPackage {
			continue
		}
		name, mustRun, skipOK, class := fields[1], fields[2], fields[3], fields[4]
		declared[name] = class
		if mustRun != "linux,darwin,windows" {
			t.Errorf("%s :: %s must run on every platform, not %q", thisPackage, name, mustRun)
		}
		if !strings.HasPrefix(name, "TestConformance") {
			// The contract cases below read committed files only. They never
			// touch the conformance root, so they tolerate no skip at all.
			if skipOK != "-" || class != "-" {
				t.Errorf("%s :: %s reads no conformance root, so it must tolerate no skip; row says skip=%q class=%q",
					thisPackage, name, skipOK, class)
			}
			continue
		}
		if class != "root-unset" {
			t.Errorf("%s :: %s declares skip class %q; only root-unset is survivable here.\n"+
				"\troot-content is policy `allow` in every lane, so the candidate lane would stay\n"+
				"\tgreen while this case skipped, which is exactly the state before the split.",
				thisPackage, name, class)
		}
		if skipOK != "linux,darwin,windows" {
			t.Errorf("%s :: %s tolerates the deferred-root skip on every platform, not %q", thisPackage, name, skipOK)
		}
	}

	// Every conformance case here needs a row: an undeclared case can be
	// renamed or deleted without the platform-case gate noticing.
	_, files := packageFiles(t)
	cases := 0
	for name, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "TestConformance") {
				continue
			}
			cases++
			if _, ok := declared[fn.Name.Name]; !ok {
				t.Errorf("%s declares %s, but %s has no row for it under %s",
					name, fn.Name.Name, platformCases, thisPackage)
			}
		}
	}
	if cases == 0 {
		t.Fatal("no TestConformance case found in this package; the ledger check would pass vacuously")
	}
	for _, name := range sortedKeys(declared) {
		t.Logf("ledger row: %s :: %s [%s]", thisPackage, name, declared[name])
	}
}

// enclosingFunc names the top-level function a node sits inside, or "".
func enclosingFunc(file *ast.File, node ast.Node) string {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if node.Pos() >= fn.Pos() && node.End() <= fn.End() {
			return fn.Name.Name
		}
	}
	return ""
}

func sortedKeys[T any](set map[string]T) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
