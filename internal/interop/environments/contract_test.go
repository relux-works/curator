package environments

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strings"
	"testing"
)

// The ledger this package's cases are declared in. root-artifacts.tsv is read
// through declaredArtefacts in suite_test.go, beside the guard that enforces it.
const platformCases = "../../../.github/ci/platform-cases.tsv"

// The helpers in suite_test.go the conformance root may legitimately reach, and
// the reporting methods it may be formatted into. Everything else that touches
// `root` outside suite_test.go is a read the artefact declaration cannot see.
var (
	accountableRootCalls = []string{"rootPath", "requireFamily", "readRootFile", "rootHasPath", "rootPathsNamedIn"}
	reportingMethods     = []string{"Fatal", "Fatalf", "Error", "Errorf", "Log", "Logf"}
)

// rootIdent is the name suiteRoot's result is bound to everywhere in this
// package. TestNoRootReadHereEscapesTheDeclaredArtefactGuard tracks it by name,
// and TestTheRootIsAlwaysBoundToThatOneName keeps the name from moving.
const rootIdent = "root"

// packageFiles parses every Go file beside this case.
//
// The scans below work on the SYNTAX TREE and never on the file text. A text
// scan is fooled by the obvious evasion -- bind the *testing.T to another name
// and call Skipf on that -- while still matching its own source, comments and
// string literals. The parser sees the expression either way.
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

// skipSelector reports a `Skip`, `Skipf` or `SkipNow` selector, whatever it is
// selected ON and whether or not it is called.
//
// It matches the SELECTOR and never the receiver's shape. An earlier version
// required the receiver to be a bare *ast.Ident and two shapes walked straight
// past it while really skipping the case:
//
//	h := holder{t: t}; h.t.Skipf(...)   -- a selector chain: X is a SelectorExpr
//	skip := t.Skipf; skip(...)          -- a method value: never a CallExpr.Fun
//
// Both are `<expression>.Skip*`, so matching on the selector alone catches
// them, and catches the method expression `(*testing.T).Skip` as well.
//
// The bound this leaves is named rather than implied: a skip reached through
// reflection, or through a helper in another package whose own name does not
// start with `Skip`, is not a `Skip*` selector here and this scan cannot see
// it. That residue is covered behaviourally, not statically --
// `.github/ci/platform-case-gate.sh` classifies the reason of every skip that
// actually happens, and a reason it does not recognise, or a `root-unset` one
// in a package this lane serves, is fatal there.
func skipSelector(node ast.Node) (*ast.SelectorExpr, bool) {
	sel, isSel := node.(*ast.SelectorExpr)
	if !isSel || !strings.HasPrefix(sel.Sel.Name, "Skip") {
		return nil, false
	}
	return sel, true
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

// familiesReadHere is the set of artefacts `requireFamily` names in this
// package's sources.
func familiesReadHere(t *testing.T) []string {
	t.Helper()
	_, files := packageFiles(t)
	read := map[string]bool{}
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			if family, ok := requiredFamily(node); ok {
				read[family] = true
			}
			return true
		})
	}
	if len(read) == 0 {
		t.Fatal("no requireFamily call found; either the cases stopped reading the root or the scan broke")
	}
	return sortedKeys(read)
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
// This is the STATIC half, and it needs no conformance root: every family the
// sources require must be declared. The other half -- that the declaration is
// not a superset, and that it also covers the paths the vectors themselves name
// -- is derived from the vectors by
// TestConformanceEveryPathTheVectorsNameIsDeclared, which needs a served root.
func TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts(t *testing.T) {
	read := familiesReadHere(t)

	declared, ok := declaredArtefacts(t, thisPackage)
	if !ok {
		t.Fatalf("%s declares no artefacts for %s.\n"+
			"\tWithout a row, suite-plan.sh serves this package against ANY root and a\n"+
			"\tcandidate that dropped an environments family would pass the lane green.",
			rootArtefacts, thisPackage)
	}

	for _, family := range read {
		if _, covered := coveringArtefact(declared, family); !covered {
			t.Errorf("%s is read by this package but NOT declared in %s for %s.\n"+
				"\tA root that drops it would still SERVE this package, so the candidate lane\n"+
				"\twould not fail by name -- exactly the hole the split closed.",
				family, rootArtefacts, thisPackage)
		}
	}
	t.Logf("required=%s", strings.Join(read, " "))
	t.Logf("declared=%s", strings.Join(declared, " "))
}

// TestConformanceEveryPathTheVectorsNameIsDeclared derives the artefact
// declaration from the vectors instead of restating it in a list beside them.
//
// Four of the declared artefacts are vector families a `requireFamily` call
// names, so the static case above covers them. The rest are CROSS-REFERENCED:
// nothing names them in Go source, because the vectors name them in their own
// `expected` and `fixture` fields and the cases read whatever those fields say.
// A hand-kept list of those is a place to forget -- and one WAS forgotten:
// `expected/byte-exact-snapshot_sha256.txt`, read through readRootFile on every
// snapshot case, was declared nowhere.
//
// So this case reads the vectors, collects every string in them that the root
// actually resolves as a path, and requires both directions:
//
//   - every such path is covered by a declared artefact -- otherwise a root
//     could drop it while suite-plan.sh still served this package;
//   - every declared artefact is either a family the sources require or covers
//     at least one such path -- otherwise it defers this package on roots that
//     could have served it.
func TestConformanceEveryPathTheVectorsNameIsDeclared(t *testing.T) {
	root := suiteRoot(t)
	declared, ok := declaredArtefacts(t, thisPackage)
	if !ok {
		t.Fatalf("%s declares no artefacts for %s", rootArtefacts, thisPackage)
	}

	covers := map[string]int{}
	crossReferenced := 0
	for _, family := range familiesReadHere(t) {
		covers[family]++
		var document any
		if err := json.Unmarshal(requireFamily(t, root, family), &document); err != nil {
			t.Fatalf("decoding %s: %v", family, err)
		}
		for _, rel := range rootPathsNamedIn(document, root) {
			crossReferenced++
			artefact, covered := coveringArtefact(declared, rel)
			if !covered {
				t.Errorf("%s names %s, which this root publishes and these cases read, but %s declares no artefact covering it for %s.\n"+
					"\tsuite-plan.sh checks only the declared set, so a root dropping %s would still SERVE this\n"+
					"\tpackage and the candidate lane could not fail by name.",
					family, rel, rootArtefacts, thisPackage, rel)
				continue
			}
			covers[artefact]++
		}
	}
	if crossReferenced == 0 {
		t.Fatal("no vector named a path this root resolves; the derivation found nothing and would pass vacuously")
	}

	for _, artefact := range declared {
		if covers[artefact] == 0 {
			t.Errorf("%s declares %s for %s, but no case requires it and no vector names a path under it.\n"+
				"\tAn artefact nothing reads defers this package on roots that could have served it.",
				rootArtefacts, artefact, thisPackage)
		}
	}
	for _, artefact := range sortedKeys(covers) {
		t.Logf("declared %s covers %d read path(s)", artefact, covers[artefact])
	}
}

// TestNoRootReadHereEscapesTheDeclaredArtefactGuard keeps `rootPath` the only
// way this package turns the conformance root into a path on disk.
//
// `rootPath` refuses a path no declared artefact covers, which is what stops a
// new read from drifting undeclared -- but only while every read goes through
// it. One `filepath.Join(root, ...)` or `os.ReadFile(root + "/x")` in a case
// re-opens the gap silently: the read succeeds on the roots that publish it,
// and the plan never learns the path exists.
//
// So outside suite_test.go, where the guard itself lives, the identifier `root`
// may only be bound, returned, passed to one of the accountable helpers, or
// formatted into a failure message. Anywhere else it is a finding.
func TestNoRootReadHereEscapesTheDeclaredArtefactGuard(t *testing.T) {
	fset, files := packageFiles(t)
	occurrences := 0
	for _, name := range sortedKeys(files) {
		if name == "suite_test.go" {
			continue
		}
		file := files[name]

		sanctioned := map[token.Pos]bool{}
		mark := func(expr ast.Expr) {
			if ident, ok := expr.(*ast.Ident); ok && ident.Name == rootIdent {
				sanctioned[ident.Pos()] = true
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch statement := node.(type) {
			case *ast.CallExpr:
				if !accountableCall(statement) {
					return true
				}
				// Only a BARE `root` argument is sanctioned. `rootPath(t,
				// filepath.Join(root, ".."), rel)` must still be a finding.
				for _, arg := range statement.Args {
					mark(arg)
				}
			case *ast.AssignStmt:
				for _, lhs := range statement.Lhs {
					mark(lhs)
				}
			case *ast.ReturnStmt:
				for _, result := range statement.Results {
					mark(result)
				}
			}
			return true
		})

		ast.Inspect(file, func(node ast.Node) bool {
			ident, ok := node.(*ast.Ident)
			if !ok || ident.Name != rootIdent {
				return true
			}
			occurrences++
			if sanctioned[ident.Pos()] {
				return true
			}
			position := fset.Position(ident.Pos())
			t.Errorf("%s:%d uses the conformance root outside %v.\n"+
				"\tEvery root path this package reads must go through rootPath, which refuses a path\n"+
				"\t%s declares no artefact for. A read that skips the guard is invisible to\n"+
				"\tsuite-plan.sh: the root can drop it while this package is still served, and the\n"+
				"\tcandidate lane cannot fail by name. Put the helper in suite_test.go instead.",
				position.Filename, position.Line, accountableRootCalls, rootArtefacts)
			return true
		})
	}
	if occurrences == 0 {
		t.Fatal("no use of the conformance root found outside suite_test.go; the scan would pass vacuously")
	}
	t.Logf("%d uses of %q checked outside suite_test.go", occurrences, rootIdent)
}

// accountableCall reports whether a call may receive the conformance root: one
// of the suite_test.go helpers that runs the artefact check, or a testing
// reporting method that only formats it into text.
func accountableCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		for _, name := range accountableRootCalls {
			if fun.Name == name {
				return true
			}
		}
	case *ast.SelectorExpr:
		for _, name := range reportingMethods {
			if fun.Sel.Name == name {
				return true
			}
		}
	}
	return false
}

// TestTheRootIsAlwaysBoundToThatOneName keeps the scan above honest. It tracks
// the conformance root by the identifier `root`, so binding suiteRoot's result
// to a second name would hide every read of it.
func TestTheRootIsAlwaysBoundToThatOneName(t *testing.T) {
	fset, files := packageFiles(t)
	bindings := 0
	for _, name := range sortedKeys(files) {
		file := files[name]
		ast.Inspect(file, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for index, rhs := range assignment.Rhs {
				call, isCall := rhs.(*ast.CallExpr)
				if !isCall {
					continue
				}
				callee, isIdent := call.Fun.(*ast.Ident)
				if !isIdent || (callee.Name != "suiteRoot" && callee.Name != "loadEnvironmentsVector") {
					continue
				}
				// suiteRoot returns the root; loadEnvironmentsVector returns it
				// first. Either way the root is the assignment's first target.
				if index != 0 || len(assignment.Lhs) == 0 {
					continue
				}
				bindings++
				// `_` is not a binding: nothing can be read through the blank
				// identifier, so discarding the root is always safe.
				bound, isIdent := assignment.Lhs[0].(*ast.Ident)
				if !isIdent || (bound.Name != rootIdent && bound.Name != "_") {
					position := fset.Position(assignment.Pos())
					t.Errorf("%s:%d binds the conformance root to something other than %q.\n"+
						"\tTestNoRootReadHereEscapesTheDeclaredArtefactGuard tracks it by that name;\n"+
						"\tunder another one, every read of it becomes invisible to the scan.",
						position.Filename, position.Line, rootIdent)
				}
			}
			return true
		})
	}
	if bindings == 0 {
		t.Fatal("nothing binds the conformance root; the scan would pass vacuously")
	}
	t.Logf("%d root bindings checked", bindings)
}

// TestNoCaseHereSkipsForAnythingButTheDeferredRoot keeps the package's single
// legitimate skip single.
//
// A per-family `t.Skipf("... publishes no vectors/x.json ...")` classifies as
// `root-content` in `.github/ci/skip-classes.tsv`, which is policy `allow` in
// EVERY lane including the candidate one. One such guard anywhere in this
// package reopens the hole for that family. The only skip allowed here is the
// `root-unset` one `suiteRoot` raises, which skip-classes.tsv admits solely for
// a package `suite-plan.sh` actually deferred.
//
// The scan matches the `Skip*` SELECTOR, so the receiver's shape is irrelevant:
// a renamed binding, a selector chain and a method value are all caught. See
// skipSelector for the bound it does not cover, and where that residue is
// covered behaviourally instead.
func TestNoCaseHereSkipsForAnythingButTheDeferredRoot(t *testing.T) {
	fset, files := packageFiles(t)
	skips := 0
	for _, name := range sortedKeys(files) {
		file := files[name]
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := skipSelector(node)
			if !ok {
				return true
			}
			position := fset.Position(selector.Pos())
			// The receiver is reported as source text and position, never
			// matched by name: naming it is what let two shapes through.
			receiver := types.ExprString(selector.X)
			// The one legitimate skip: suiteRoot reporting that the suite plan
			// deferred this package. Recognised by WHERE it is, not by the text
			// of the line, so renaming the receiver cannot smuggle another one in.
			if name == "suite_test.go" && enclosingFunc(file, selector) == "suiteRoot" {
				skips++
				t.Logf("the deferred-root skip: %s:%d %s.%s", position.Filename, position.Line, receiver, selector.Sel.Name)
				return true
			}
			t.Errorf("%s:%d reaches %s.%s outside suiteRoot.\n"+
				"\tEvery artefact here is declared in %s, so a missing one must DEFER the package\n"+
				"\tand fail a CI_REQUIRE_FULL_ROOT lane by name. A skip makes the lane pass green\n"+
				"\tinstead -- a per-family root-content skip is policy `allow` in EVERY lane.",
				position.Filename, position.Line, receiver, selector.Sel.Name, rootArtefacts)
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
// the reason carries the class the row declares AND the class's own policy
// admits it in this lane. `root-unset` is policy `deferred-only`: legitimate
// solely for a package suite-plan.sh deferred. `root-content` is policy `allow`
// everywhere -- a row carrying it lets these cases skip in the candidate lane
// and pass. That is what the rows said before the split, and it is the single
// character of this ledger a regression would most plausibly touch.
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
