package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/skillspec"
)

// The documentation is a contract with skill authors and operators, so it is
// gated like one. These tests parse the shipped documents rather than a copy:
// an example that stopped parsing, a vocabulary that grew without a row, or a
// link that stopped resolving fails the build.
const (
	readmePath = "../../README.md"
	buildsDoc  = "../../docs/compiled-builds.md"
)

var (
	fenceRE = regexp.MustCompile("(?s)```([a-z]*)\n(.*?)\n```")
	// linkRE matches inline markdown links; bare autolinks and reference
	// definitions are deliberately out of scope.
	linkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)
	// headingRE captures ATX headings so anchor targets can be derived.
	headingRE = regexp.MustCompile(`(?m)^#{1,6} +(.+?) *$`)
)

// driverPackageDir is the go-v1 driver package the failure table is checked
// against.
const driverPackageDir = "../../internal/godriver"

// relayedCodeArguments are the boundary-code arguments that deliberately
// forward a code produced elsewhere instead of naming a new one. Every other
// unresolvable argument fails the scan, so a new indirection cannot silently
// shrink the discovered set.
var relayedCodeArguments = map[string]string{
	// The worker reports its own diagnostic over the session protocol; the
	// manager relays it verbatim. The worker emits it from this same package.
	"message.Failure.Code": "worker-reported diagnostic relayed by the manager",
}

func readDoc(t *testing.T, path string) string {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(payload)
}

// fencedBlocks returns every fenced block of one language, in document order.
func fencedBlocks(text, language string) []string {
	var blocks []string
	for _, match := range fenceRE.FindAllStringSubmatch(text, -1) {
		if match[1] == language {
			blocks = append(blocks, match[2])
		}
	}
	return blocks
}

func mustContain(t *testing.T, text, needle, document, what string) {
	t.Helper()
	if !strings.Contains(text, needle) {
		t.Errorf("%s does not document the %s %q", document, what, needle)
	}
}

// TestDocumentedJSONBlocksParse proves every JSON example in the shipped
// documentation is real JSON rather than illustrative pseudo-syntax.
func TestDocumentedJSONBlocksParse(t *testing.T) {
	for _, path := range []string{readmePath, buildsDoc} {
		blocks := fencedBlocks(readDoc(t, path), "json")
		if len(blocks) == 0 && path == buildsDoc {
			t.Fatalf("%s carries no JSON example", path)
		}
		for index, block := range blocks {
			var value any
			if err := json.Unmarshal([]byte(block), &value); err != nil {
				t.Errorf("%s json block %d does not parse: %v", path, index, err)
			}
		}
	}
}

// TestDocumentedMixedManifestLoads materializes the documented mixed package
// exactly as the guide presents it and loads it through the real parser, so the
// example cannot drift from the schema 6 rules it claims to demonstrate.
func TestDocumentedMixedManifestLoads(t *testing.T) {
	doc := readDoc(t, buildsDoc)

	manifests := fencedBlocks(doc, "json")
	if len(manifests) == 0 {
		t.Fatal("the guide carries no manifest example")
	}
	manifest := manifests[0]

	goMods := fencedBlocks(doc, "")
	var goMod string
	for _, block := range goMods {
		if strings.HasPrefix(block, "module ") {
			goMod = block
			break
		}
	}
	if goMod == "" {
		t.Fatal("the guide carries no go.mod example")
	}

	sources := fencedBlocks(doc, "go")
	if len(sources) == 0 {
		t.Fatal("the guide carries no Go source example")
	}

	dir := t.TempDir()
	files := map[string]string{
		"agent-skill.json":          manifest,
		"SKILL.md":                  "# Example skill\n\nRun `indexer`.\n",
		"scripts/report":            "#!/bin/sh\nexit 0\n",
		"build/go.mod":              goMod + "\n",
		"build/cmd/indexer/main.go": sources[0] + "\n",
	}
	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	spec, err := skillspec.Load(dir)
	if err != nil {
		t.Fatalf("the documented example does not load: %v", err)
	}
	if spec.SchemaVersion != 6 {
		t.Fatalf("schema_version = %d, want 6", spec.SchemaVersion)
	}
	if spec.SourceFile != skillspec.CanonicalManifestName {
		t.Fatalf("source file = %q, want %q", spec.SourceFile, skillspec.CanonicalManifestName)
	}
	if len(spec.BuildRoots) != 1 || spec.BuildRoots[0] != "build" {
		t.Fatalf("build_roots = %v, want [build]", spec.BuildRoots)
	}
	if len(spec.RuntimeRoots) != 1 || spec.RuntimeRoots[0] != "scripts" {
		t.Fatalf("runtime_roots = %v, want [scripts]", spec.RuntimeRoots)
	}

	compiled := spec.Commands["indexer"]
	if compiled.Type != "build" || compiled.Driver != "go-v1" || compiled.SourceDir != "build/cmd/indexer" {
		t.Fatalf("compiled command = %+v", compiled)
	}
	if got := spec.Commands["report"]; got.Type != "script" || got.UnixPath != "scripts/report" {
		t.Fatalf("script command = %+v", got)
	}
	if got := spec.Commands["git"]; got.Type != "system" || got.Command != "git" {
		t.Fatalf("system command = %+v", got)
	}
	if len(spec.Commands) != 3 {
		t.Fatalf("the example must stay a mixed package, got %d commands", len(spec.Commands))
	}
}

// TestDocumentedCurrentnessVocabulary proves the reachable status vocabulary
// and the documented one are the same set, in both directions.
func TestDocumentedCurrentnessVocabulary(t *testing.T) {
	readme := readDoc(t, readmePath)
	for _, code := range currentnessCodes() {
		mustContain(t, readme, "`"+code+"`", "README.md", "currentness code")
	}
	for _, cause := range inputCauses() {
		mustContain(t, readme, "`"+cause+"`", "README.md", "input-drift cause")
	}

	known := map[string]bool{}
	for _, code := range currentnessCodes() {
		known[code] = true
	}
	for _, cause := range inputCauses() {
		known[cause] = true
	}

	// The reverse direction, scoped to the two tables that claim to enumerate
	// the vocabulary: a first-column entry that no longer names a reachable
	// state or cause is a stale row. Prose elsewhere in the section is not
	// policed, so ordinary hyphenated terms are unaffected.
	var stale []string
	for _, cell := range statusTableKeyCells(t, readme) {
		for _, match := range regexp.MustCompile("`([a-z][a-z-]{2,})`").FindAllStringSubmatch(cell, -1) {
			if !known[match[1]] {
				stale = append(stale, match[1])
			}
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("README.md documents unknown build states or causes: %s", strings.Join(stale, ", "))
	}
}

// statusTableKeyCells returns the first column of every data row in the README
// tables of the compiled-status section: the code table and the cause table.
func statusTableKeyCells(t *testing.T, readme string) []string {
	t.Helper()
	_, after, found := strings.Cut(readme, "\n## Compiled-command status, diagnostics, and repair\n")
	if !found {
		t.Fatal("README.md has no compiled-status section")
	}
	if next := strings.Index(after, "\n## "); next >= 0 {
		after = after[:next]
	}
	var cells []string
	for _, line := range strings.Split(after, "\n") {
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		columns := strings.Split(line, "|")
		if len(columns) < 3 {
			t.Fatalf("compiled-status row is not a two-column row: %s", line)
		}
		cells = append(cells, columns[1])
	}
	if len(cells) == 0 {
		t.Fatal("the compiled-status section carries no table rows")
	}
	return cells
}

// TestDocumentedDryRunOutcomes proves every planner outcome a dry run can
// report is documented in both the README and the authoring guide.
func TestDocumentedDryRunOutcomes(t *testing.T) {
	outcomes := []install.BuildOutcome{
		install.BuildCacheHit,
		install.BuildWouldPreflightAndBuild,
		install.BuildWouldRebuildUntrustedCache,
		install.BuildCorrupt,
		install.BuildUnsupported,
		install.BuildToolchainUnavailable,
	}
	readme := readDoc(t, readmePath)
	guide := readDoc(t, buildsDoc)
	for _, outcome := range outcomes {
		mustContain(t, readme, "`"+string(outcome)+"`", "README.md", "dry-run outcome")
		mustContain(t, guide, "`"+string(outcome)+"`", "docs/compiled-builds.md", "dry-run outcome")
	}
}

// TestDocumentedToolchainSelection proves the documented selection mechanisms
// and tested release families are the ones the driver actually accepts.
func TestDocumentedToolchainSelection(t *testing.T) {
	readme := readDoc(t, readmePath)
	guide := readDoc(t, buildsDoc)
	for _, document := range []struct {
		name string
		text string
	}{{"README.md", readme}, {"docs/compiled-builds.md", guide}} {
		mustContain(t, document.text, godriver.SelectionCuratorGo, document.name, "selection mechanism")
		mustContain(t, document.text, godriver.SelectionGOROOT, document.name, "selection mechanism")
		if !strings.Contains(document.text, "never searches `PATH`") {
			t.Errorf("%s does not state that PATH is never searched", document.name)
		}
	}
	for _, family := range godriver.TestedFamilies() {
		mustContain(t, guide, "`"+family+"`", "docs/compiled-builds.md", "tested Go release family")
	}
}

// TestDocumentedCompiledBuildAuthoringContract locks the task's required
// author and operator guidance into the shipped documentation. These are
// security and compatibility boundaries, not optional explanatory detail.
func TestDocumentedCompiledBuildAuthoringContract(t *testing.T) {
	guide := readDoc(t, buildsDoc)
	requiredGuideText := []struct {
		needle string
		what   string
	}{
		{"Everything in schemas 1 through 5 keeps working exactly as before.", "schema 1 through 5 compatibility"},
		{"Build roots never reach agent context", "build-root context exclusion"},
		{"Vendoring is mandatory.", "vendor-only prerequisite"},
		{"Embedded inputs must be regular files inside the build root.", "embed-input prerequisite"},
		{"Lifecycle hooks.", "unsupported lifecycle hooks"},
		{"Package-supplied argv or environment.", "unsupported package argv and environment"},
		{"cgo.", "unsupported cgo"},
		{"Go workspaces.", "unsupported workspaces"},
		{"Network access of any kind during a build.", "unsupported downloads"},
		{"External linking.", "unsupported external linking"},
		{"Root modules other than the declared build root.", "unsupported root modules"},
		{"Any driver other than `go-v1`.", "unsupported generic drivers"},
		{"The compiled output is untrusted, and Curator never runs it.", "untrusted-output boundary"},
		{"does not execute the artifact during `install`, `upgrade`, `status`,", "no install-time execution"},
		{"Portable logical identity", "portable logical identity"},
		{"Curator-local paths", "implementation-local paths"},
		{"`install` and `upgrade` are", "install and upgrade repair"},
		{"`gc` removes an unreferenced protected cache entry", "locked garbage collection behavior"},
		{".agents/bin/<command>", "Unix shim invocation"},
		{`.agents\bin\<command>.cmd`, "Windows shim invocation"},
	}
	for _, required := range requiredGuideText {
		mustContain(t, guide, required.needle, "docs/compiled-builds.md", required.what)
	}

	readme := readDoc(t, readmePath)
	for _, required := range []struct {
		needle string
		what   string
	}{
		{"`status --json`", "JSON status"},
		{"`status --check`", "status verdict"},
		{"`install --dry-run`", "install dry run"},
		{"`upgrade --dry-run`", "upgrade dry run"},
		{"exclusive\nmanager-home mutation lock", "locked garbage collection"},
	} {
		mustContain(t, readme, required.needle, "README.md", required.what)
	}
}

// TestDocumentedBoundaryCodesAreComplete proves the guide's failure table names
// every go-v1 boundary code the driver can emit, and no code it cannot. The
// table claims to be complete, so it is checked against the driver sources
// rather than against a second hand-maintained list.
func TestDocumentedBoundaryCodesAreComplete(t *testing.T) {
	emitted := emittedBoundaryCodes(t)
	if len(emitted) == 0 {
		t.Fatal("no go-v1 boundary codes were discovered")
	}

	// Only the failure table's Codes column claims to enumerate boundary codes,
	// so the comparison is scoped to it. Underscore-cased identifiers elsewhere
	// in the guide, and in the table's remedy column, are manifest fields and
	// marker keys rather than codes.
	documented := map[string]bool{}
	for _, codes := range failureTableCodeCells(t, readDoc(t, buildsDoc)) {
		for _, match := range regexp.MustCompile("`([a-z][a-z0-9_]{4,})`").FindAllStringSubmatch(codes, -1) {
			documented[match[1]] = true
		}
	}
	if len(documented) == 0 {
		t.Fatal("the failure-classes table carries no boundary codes")
	}

	var missing []string
	for code := range emitted {
		if !documented[code] {
			missing = append(missing, code)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("docs/compiled-builds.md omits go-v1 boundary codes: %s", strings.Join(missing, ", "))
	}

	var stale []string
	for code := range documented {
		if !emitted[code] {
			stale = append(stale, code)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("docs/compiled-builds.md documents unknown go-v1 boundary codes: %s", strings.Join(stale, ", "))
	}
}

// emittedBoundaryCodes returns every go-v1 boundary code the driver package can
// emit. The scan is AST-based on purpose: a code reaches diagnostic or
// diagnosticErr as a string literal, as a package-level constant, or through a
// local variable, and a textual scan silently misses the last form. Whatever
// this resolver cannot reduce to a literal is reported as a failure rather than
// skipped, so a new indirection cannot make the completeness check pass by
// discovering fewer codes than the driver can raise.
func emittedBoundaryCodes(t *testing.T) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(driverPackageDir)
	if err != nil {
		t.Fatalf("read driver package: %v", err)
	}
	fileSet := token.NewFileSet()
	var files []*ast.File
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		// Build constraints are deliberately ignored: the documented set is the
		// whole go-v1 surface, not the subset this host compiles.
		parsed, err := parser.ParseFile(fileSet, filepath.Join(driverPackageDir, name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		files = append(files, parsed)
	}

	// A name can carry more than one literal: a package constant is declared
	// once per build-constrained file, and a local is often reassigned on a
	// branch. Every value it can hold is a code the driver can raise.
	constants := map[string][]string{}
	for _, file := range files {
		for _, declaration := range file.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, specification := range group.Specs {
				if values, ok := specification.(*ast.ValueSpec); ok {
					recordStringValues(constants, values)
				}
			}
		}
	}

	emitted := map[string]bool{}
	for _, file := range files {
		resolved := 0
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			resolved += resolveDiagnosticCalls(t, fileSet, function, constants, emitted)
		}
		// A diagnostic raised outside a function body is never visited above, so
		// the two counts must agree or the scan is incomplete.
		if total := countDiagnosticCalls(file); total != resolved {
			t.Errorf("%s: %d diagnostic calls sit outside a function body and were not scanned",
				fileSet.Position(file.Package).Filename, total-resolved)
		}
	}
	return emitted
}

// resolveDiagnosticCalls records the code of every diagnostic raised in one
// function and returns how many calls it visited.
func resolveDiagnosticCalls(t *testing.T, fileSet *token.FileSet, function *ast.FuncDecl, constants map[string][]string, emitted map[string]bool) int {
	t.Helper()
	locals := map[string][]string{}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		switch statement := node.(type) {
		case *ast.AssignStmt:
			for index, target := range statement.Lhs {
				if index >= len(statement.Rhs) {
					break
				}
				if name, ok := target.(*ast.Ident); ok && name.Name != "_" {
					if literal, ok := stringLiteral(statement.Rhs[index]); ok {
						locals[name.Name] = append(locals[name.Name], literal)
					}
				}
			}
		case *ast.ValueSpec:
			recordStringValues(locals, statement)
		}
		return true
	})

	visited := 0
	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isDiagnosticCall(call) {
			return true
		}
		visited++
		argument := call.Args[0]
		if literal, ok := stringLiteral(argument); ok {
			emitted[literal] = true
			return true
		}
		if name, ok := argument.(*ast.Ident); ok {
			if literals, ok := locals[name.Name]; ok {
				for _, literal := range literals {
					emitted[literal] = true
				}
				return true
			}
			if literals, ok := constants[name.Name]; ok {
				for _, literal := range literals {
					emitted[literal] = true
				}
				return true
			}
		}
		if _, relayed := relayedCodeArguments[types.ExprString(argument)]; relayed {
			return true
		}
		t.Errorf("%s: boundary code %s resolves to no literal; give it a constant or allow it as a relay",
			fileSet.Position(call.Pos()), types.ExprString(argument))
		return true
	})
	return visited
}

// recordStringValues copies every string-literal initializer of one value
// specification into values, keyed by declared name.
func recordStringValues(values map[string][]string, specification *ast.ValueSpec) {
	for index, name := range specification.Names {
		if index >= len(specification.Values) || name.Name == "_" {
			continue
		}
		if literal, ok := stringLiteral(specification.Values[index]); ok {
			values[name.Name] = append(values[name.Name], literal)
		}
	}
}

// isDiagnosticCall reports whether one call raises a go-v1 boundary diagnostic.
func isDiagnosticCall(call *ast.CallExpr) bool {
	if len(call.Args) == 0 {
		return false
	}
	name, ok := call.Fun.(*ast.Ident)
	return ok && (name.Name == "diagnostic" || name.Name == "diagnosticErr")
}

// countDiagnosticCalls counts every diagnostic raised anywhere in one file.
func countDiagnosticCalls(file *ast.File) int {
	count := 0
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && isDiagnosticCall(call) {
			count++
		}
		return true
	})
	return count
}

// stringLiteral unquotes one expression when it is an untyped string literal.
func stringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		return "", false
	}
	return value, true
}

// failureTableCodeCells returns the Codes cell of every data row in the guide's
// failure-classes table: the second column of each markdown row between the
// "Failure classes" heading and the next heading, excluding the header row and
// its separator.
func failureTableCodeCells(t *testing.T, guide string) []string {
	t.Helper()
	_, after, found := strings.Cut(guide, "\n## Failure classes\n")
	if !found {
		t.Fatal("docs/compiled-builds.md has no failure-classes section")
	}
	if next := strings.Index(after, "\n## "); next >= 0 {
		after = after[:next]
	}
	var cells []string
	for _, line := range strings.Split(after, "\n") {
		if !strings.HasPrefix(line, "|") || strings.HasPrefix(line, "|---") || strings.HasPrefix(line, "| Group ") {
			continue
		}
		columns := strings.Split(line, "|")
		if len(columns) < 4 {
			t.Fatalf("failure-classes row is not a three-column row: %s", line)
		}
		cells = append(cells, columns[2])
	}
	return cells
}

// TestDocumentedLinksResolve proves every relative link and in-document anchor
// in the shipped documentation points at something that exists.
func TestDocumentedLinksResolve(t *testing.T) {
	documents := map[string]string{
		readmePath: readDoc(t, readmePath),
		buildsDoc:  readDoc(t, buildsDoc),
	}
	for path, text := range documents {
		anchors := map[string]bool{}
		for _, match := range headingRE.FindAllStringSubmatch(text, -1) {
			anchors[anchorFor(match[1])] = true
		}
		for _, match := range linkRE.FindAllStringSubmatch(text, -1) {
			target := match[1]
			switch {
			case strings.HasPrefix(target, "http://"), strings.HasPrefix(target, "https://"):
				continue
			case strings.HasPrefix(target, "#"):
				if !anchors[strings.TrimPrefix(target, "#")] {
					t.Errorf("%s: anchor %q has no heading", path, target)
				}
			default:
				file, fragment, _ := strings.Cut(target, "#")
				resolved := filepath.Join(filepath.Dir(path), filepath.FromSlash(file))
				if _, err := os.Stat(resolved); err != nil {
					t.Errorf("%s: link %q does not resolve: %v", path, target, err)
					continue
				}
				if fragment == "" {
					continue
				}
				payload, err := os.ReadFile(resolved) // #nosec G304 -- path is derived from a checked-in document
				if err != nil {
					t.Errorf("%s: cannot read link target %q: %v", path, target, err)
					continue
				}
				found := false
				for _, heading := range headingRE.FindAllStringSubmatch(string(payload), -1) {
					if anchorFor(heading[1]) == fragment {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: link %q names a heading that does not exist", path, target)
				}
			}
		}
	}
}

// anchorFor derives the GitHub heading anchor of one heading title.
func anchorFor(title string) string {
	lowered := strings.ToLower(strings.TrimSpace(title))
	var builder strings.Builder
	for _, character := range lowered {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9', character == '-', character == '_':
			builder.WriteRune(character)
		case character == ' ':
			builder.WriteRune('-')
		}
	}
	return builder.String()
}
