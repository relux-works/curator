// Package harnessboundary checks source architecture without executing adapters.
package harnessboundary

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/harnessregistry"
)

// Exception pins one preexisting source-policy occurrence.
type Exception struct {
	Path       string `json:"path"`
	HarnessID  string `json:"harness_id"`
	LineSHA256 string `json:"line_sha256"`
}

// ExceptionSet binds occurrence allowances to the original source head.
type ExceptionSet struct {
	SourceHead string      `json:"source_head"`
	Exceptions []Exception `json:"exceptions"`
}

// Finding identifies a concrete source path and line violation.
type Finding struct {
	Path                        string
	Line                        int
	HarnessID, Rule, LineSHA256 string
}

// String returns the validated relative spelling.
func (f Finding) String() string {
	return fmt.Sprintf("%s:%d: %s (%s)", f.Path, f.Line, f.Rule, f.HarnessID)
}

func digest(bytes []byte) string { sum := sha256.Sum256(bytes); return hex.EncodeToString(sum[:]) }
func key(e Exception) string     { return e.Path + "\x00" + e.HarnessID + "\x00" + e.LineSHA256 }

// CheckExceptionSubset uses counts, not set membership: a duplicate consumes a
// distinct allowance. Baseline allowances cannot be replenished by new entries.
func CheckExceptionSubset(frozen, active ExceptionSet) error {
	if frozen.SourceHead == "" || active.SourceHead != frozen.SourceHead {
		return fmt.Errorf("harness_exception_baseline_mismatch")
	}
	counts := map[string]int{}
	for _, e := range frozen.Exceptions {
		counts[key(e)]++
	}
	for _, e := range active.Exceptions {
		if counts[key(e)] == 0 {
			return fmt.Errorf("%s: harness_exception_growth (%s)", e.Path, e.HarnessID)
		}
		counts[key(e)]--
	}
	return nil
}

func productionFiles(root string) ([]string, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// CI checks out the authoritative protocol as a sibling repository.
		// It is an input to Curator's tests, not Curator-owned production
		// source, so its Go generators do not belong in this source scan.
		if filepath.Clean(relative) == "protocol-spec" {
			return filepath.SkipDir
		}
		// This declared test-tool submodule is not Curator production source.
		if filepath.Clean(relative) == filepath.Join("agents", "skills", "skill-go-testing-tools") {
			return filepath.SkipDir
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if isDeclaredSkillSubmoduleLink(absoluteRoot, path) {
				return nil
			}
			productionSource := strings.HasSuffix(relative, ".go") && !strings.HasSuffix(relative, "_test.go")
			if productionSource {
				return fmt.Errorf("%s: source_symlink_unsupported", path)
			}
			target, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("%s: source_symlink_unresolved", path)
			}
			if target.IsDir() {
				return fmt.Errorf("%s: source_directory_symlink_unsupported", path)
			}
		}
		if entry.IsDir() {
			if excludedProductionDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(relative, ".go") && !strings.HasSuffix(relative, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func excludedProductionDirectory(name string) bool {
	switch name {
	case ".git", ".temp", ".task-board", "testdata", "vendor":
		return true
	default:
		return false
	}
}

var declaredSkillSubmoduleLinkPaths = map[string]struct{}{
	filepath.Join(strings.Join([]string{".cl", "aude"}, ""), "skills", "skill-go-testing-tools"): {},
	filepath.Join(strings.Join([]string{".co", "dex"}, ""), "skills", "skill-go-testing-tools"):  {},
}

func isDeclaredSkillSubmoduleLink(root, path string) bool {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	linkPath, err := filepath.Rel(root, absolutePath)
	if err != nil {
		return false
	}
	if _, allowed := declaredSkillSubmoduleLinkPaths[filepath.Clean(linkPath)]; !allowed {
		return false
	}
	linkTarget, err := os.Readlink(absolutePath)
	if err != nil {
		return false
	}
	if filepath.IsAbs(linkTarget) {
		return false
	}
	target := filepath.Clean(filepath.Join(filepath.Dir(absolutePath), linkTarget))
	declaredTarget := filepath.Join(root, "agents", "skills", "skill-go-testing-tools")
	return target == declaredTarget
}

func adapterPath(path string) bool { return strings.HasPrefix(path, "internal/harness/") }
func exemptBranchPath(path string) bool {
	return adapterPath(path) || strings.HasPrefix(path, "internal/harnessapi/") || strings.HasPrefix(path, "internal/harnessregistry/")
}

type source struct {
	path  string
	bytes []byte
	lines []string
	tree  *ast.File
	set   *token.FileSet
}

type constantDefinition struct {
	expression  ast.Expr
	sourceIndex int
}
type constantIndex struct {
	sources     []source
	definitions map[string][]constantDefinition
}

func (index constantIndex) value(expr ast.Expr, sourceIndex int, visiting map[string]bool) (string, bool) {
	s := index.sources[sourceIndex]
	resolve := func(name string, definitions []constantDefinition) (string, bool) {
		if len(definitions) != 1 || visiting[name] {
			return "", false
		}
		visiting[name] = true
		defer delete(visiting, name)
		return index.value(definitions[0].expression, definitions[0].sourceIndex, visiting)
	}
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			value, err := strconv.Unquote(e.Value)
			return value, err == nil
		}
	case *ast.ParenExpr:
		return index.value(e.X, sourceIndex, visiting)
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			a, ak := index.value(e.X, sourceIndex, visiting)
			b, bk := index.value(e.Y, sourceIndex, visiting)
			return a + b, ak && bk
		}
	case *ast.Ident:
		if e.Obj != nil && e.Obj.Kind == ast.Con {
			if spec, ok := e.Obj.Decl.(*ast.ValueSpec); ok {
				for i, name := range spec.Names {
					if name.Name == e.Name && i < len(spec.Values) {
						return resolve(s.path+":"+strconv.Itoa(int(spec.Pos())), []constantDefinition{{spec.Values[i], sourceIndex}})
					}
				}
			}
		}
		name := filepath.ToSlash(filepath.Dir(s.path)) + "#" + e.Name
		return resolve(name, index.definitions[name])
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			for _, imp := range s.tree.Imports {
				path, _ := strconv.Unquote(imp.Path.Value)
				alias := filepath.Base(path)
				if imp.Name != nil {
					alias = imp.Name.Name
				}
				if alias == ident.Name && strings.HasPrefix(path, "github.com/relux-works/curator/") {
					name := strings.TrimPrefix(path, "github.com/relux-works/curator/") + "#" + e.Sel.Name
					return resolve(name, index.definitions[name])
				}
			}
		}
	case *ast.CallExpr:
		if len(e.Args) == 1 {
			conversion := false
			switch name := e.Fun.(type) {
			case *ast.Ident:
				conversion = name.Name == "string" || name.Obj != nil && name.Obj.Kind == ast.Typ
			case *ast.SelectorExpr:
				conversion = name.Sel.Name == "ID"
			}
			if conversion {
				return index.value(e.Args[0], sourceIndex, visiting)
			}
		}
	}
	return "", false
}

func newConstantIndex(sources []source) constantIndex {
	index := constantIndex{sources: sources, definitions: map[string][]constantDefinition{}}
	for n, s := range sources {
		for _, declaration := range s.tree.Decls {
			decl, ok := declaration.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				continue
			}
			var previous []ast.Expr
			for _, spec := range decl.Specs {
				value := spec.(*ast.ValueSpec)
				expressions := value.Values
				if len(expressions) == 0 {
					expressions = previous
				} else {
					previous = expressions
				}
				for i, name := range value.Names {
					if i < len(expressions) {
						key := filepath.ToSlash(filepath.Dir(s.path)) + "#" + name.Name
						index.definitions[key] = append(index.definitions[key], constantDefinition{expressions[i], n})
					}
				}
			}
		}
	}
	return index
}

func loadSource(root, path string) (source, error) {
	bytes, err := os.ReadFile(path) // #nosec G304 -- source walk supplies paths and rejects linked Go files.
	if err != nil {
		return source{}, err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return source{}, err
	}
	set := token.NewFileSet()
	tree, err := parser.ParseFile(set, path, bytes, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return source{}, fmt.Errorf("%s: source_parse_failed: %w", rel, err)
	}
	lines := strings.Split(string(bytes), "\n")
	// Line content excludes the complete CRLF delimiter. Preserve all other
	// whitespace and a terminal lone CR so exception hashes cannot be repaired.
	for i := 0; i+1 < len(lines); i++ {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	return source{filepath.ToSlash(rel), bytes, lines, tree, set}, nil
}

// ScanBranches includes all checked-in production boundaries, including
// packages absent from earlier inventories. Comments are scanned conservatively
// too, so existing native policy facts stay visible in the frozen allowance set.
func ScanBranches(root string, vocabulary map[string][]string) ([]Finding, error) {
	files, err := productionFiles(root)
	if err != nil {
		return nil, err
	}
	branchVocabulary := make(map[string][]string, len(vocabulary))
	for marker, ids := range vocabulary {
		branchVocabulary[marker] = append([]string(nil), ids...)
	}
	managedHomeValues := map[string]string{}
	addManagedHomeValue := func(value, id string) {
		if value == "" {
			return
		}
		managedHomeValues[value] = id
		for _, existing := range branchVocabulary[value] {
			if existing == value {
				return
			}
		}
		branchVocabulary[value] = append(branchVocabulary[value], value)
	}
	for _, spec := range harnessregistry.ManagedHomeSpecs() {
		id := string(spec.ID)
		addManagedHomeValue(id, id)
		addManagedHomeValue(spec.Alias, id)
	}
	patterns := map[string]*regexp.Regexp{}
	for marker := range vocabulary {
		patterns[marker] = regexp.MustCompile(`(^|[^A-Za-z0-9_])` + regexp.QuoteMeta(marker) + `([^A-Za-z0-9_]|$)`)
	}
	var findings []Finding
	var sources []source
	for _, file := range files {
		s, err := loadSource(root, file)
		if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	constants := newConstantIndex(sources)
	for sourceIndex, s := range sources {
		if exemptBranchPath(s.path) {
			continue
		}
		seen := map[string]bool{}
		add := func(line int, id, rule string) {
			k := strconv.Itoa(line) + ":" + id
			if seen[k] {
				return
			}
			seen[k] = true
			findings = append(findings, Finding{s.path, line, id, rule, digest([]byte(s.lines[line-1]))})
		}
		for n, line := range s.lines {
			for marker, pattern := range patterns {
				// A literal regex cannot match without its literal marker. Avoid
				// invoking the regex engine for the overwhelmingly neutral lines.
				if strings.Contains(line, marker) && pattern.MatchString(line) {
					for _, id := range vocabulary[marker] {
						add(n+1, id, "harness_policy_outside_owner")
					}
				}
			}
		}
		ast.Inspect(s.tree, func(node ast.Node) bool {
			checkID := func(expr ast.Expr) {
				value, known := constants.value(expr, sourceIndex, map[string]bool{})
				if known {
					for _, id := range branchVocabulary[value] {
						add(s.set.PositionFor(expr.Pos(), false).Line, id, "harness_id_branch_outside_owner")
					}
				}
			}
			checkCallPredicate := func(expr ast.Expr) {
				ast.Inspect(expr, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok {
						return true
					}
					for _, arg := range call.Args {
						value, known := constants.value(arg, sourceIndex, map[string]bool{})
						if id, managedHomeValue := managedHomeValues[value]; known && managedHomeValue {
							add(s.set.PositionFor(call.Pos(), false).Line, id, "harness_id_call_predicate_outside_owner")
						}
					}
					return true
				})
			}
			switch n := node.(type) {
			case *ast.IfStmt:
				checkCallPredicate(n.Cond)
			case *ast.SwitchStmt:
				if n.Tag != nil {
					checkCallPredicate(n.Tag)
				}
			case *ast.BinaryExpr:
				if n.Op == token.EQL || n.Op == token.NEQ {
					checkID(n.X)
					checkID(n.Y)
				}
				if n.Op == token.ADD {
					value, known := constants.value(n, sourceIndex, map[string]bool{})
					if known && !strings.ContainsAny(value, " \t\r\n") {
						for marker, pattern := range patterns {
							if strings.Contains(value, marker) && pattern.MatchString(value) {
								for _, id := range vocabulary[marker] {
									add(s.set.PositionFor(n.Pos(), false).Line, id, "harness_computed_literal_outside_owner")
								}
							}
						}
					}
				}
			case *ast.CaseClause:
				for _, expr := range n.List {
					checkID(expr)
					checkCallPredicate(expr)
				}
			}
			return true
		})
		for _, imported := range s.tree.Imports {
			path, _ := strconv.Unquote(imported.Path.Value)
			if strings.Contains(path, "/internal/harness/") {
				add(s.set.PositionFor(imported.Pos(), false).Line, "adapter", "concrete_adapter_import")
			}
		}
		for _, group := range s.tree.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:generate") || strings.HasPrefix(comment.Text, "// Code generated ") {
					add(s.set.PositionFor(comment.Pos(), false).Line, "generated", "generated_source_requires_scanned_build_manifest")
				}
			}
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.HarnessID < b.HarnessID
	})
	return findings, nil
}

// ApplyExceptions consumes allowances one-to-one, preserving exact diagnostics
// for every additional occurrence, even when its source text was copied.
func ApplyExceptions(findings []Finding, exceptions ExceptionSet) []Finding {
	counts := map[string]int{}
	for _, e := range exceptions.Exceptions {
		counts[key(e)]++
	}
	var unmatched []Finding
	for _, f := range findings {
		k := key(Exception{f.Path, f.HarnessID, f.LineSHA256})
		if counts[k] > 0 {
			counts[k]--
			continue
		}
		unmatched = append(unmatched, f)
	}
	return unmatched
}

var pureImports = map[string]bool{
	"bytes": true, "encoding/json": true, "fmt": true, "path": true, "path/filepath": true, "sort": true, "strconv": true, "strings": true,
	"github.com/relux-works/curator/internal/harnessapi":       true,
	"github.com/relux-works/curator/internal/harnessvalue":     true,
	"github.com/relux-works/curator/internal/harnessformats":   true,
	"github.com/relux-works/curator-runtimes/pkg/harnessapi":   true,
	"github.com/relux-works/curator-runtimes/pkg/harnessvalue": true,
}

const upstreamSystemsPrefix = "github.com/relux-works/curator-runtimes/pkg/agentic/systems/"

func isPureImport(name, sourcePath string) bool {
	if pureImports[name] {
		return true
	}
	if adapterPath(sourcePath) && strings.HasPrefix(name, upstreamSystemsPrefix) {
		rest := strings.TrimPrefix(name, upstreamSystemsPrefix)
		return rest != "" && !strings.Contains(rest, "/")
	}
	return false
}

var pureFormat = map[string]bool{"Errorf": true, "Sprint": true, "Sprintf": true, "Sprintln": true, "Append": true, "Appendf": true, "Appendln": true, "FormatString": true}
var lexicalFilepath = map[string]bool{"Base": true, "Clean": true, "Dir": true, "Ext": true, "FromSlash": true, "IsAbs": true, "IsLocal": true, "Join": true, "Localize": true, "Match": true, "Rel": true, "Separator": true, "Split": true, "SplitList": true, "ToSlash": true, "VolumeName": true}

// ScanEffects parses selectors even when an allowed package has an alias.
// Import denial also rejects effects hidden behind an unapproved helper edge.
func ScanEffects(root string) ([]Finding, error) {
	files, err := productionFiles(root)
	if err != nil {
		return nil, err
	}
	var findings []Finding
	for _, file := range files {
		s, err := loadSource(root, file)
		if err != nil {
			return nil, err
		}
		apiPath := strings.HasPrefix(s.path, "internal/harnessapi/")
		purePath := adapterPath(s.path) || apiPath || strings.HasPrefix(s.path, "internal/harnessvalue/") || strings.HasPrefix(s.path, "internal/harnessformats/")
		if !purePath {
			continue
		}
		add := func(pos token.Pos, rule string) {
			line := s.set.PositionFor(pos, false).Line
			findings = append(findings, Finding{s.path, line, "effect", rule, digest([]byte(s.lines[line-1]))})
		}
		imports := map[string]string{}
		for _, imp := range s.tree.Imports {
			name, _ := strconv.Unquote(imp.Path.Value)
			local := filepath.Base(name)
			if imp.Name != nil {
				local = imp.Name.Name
			}
			imports[local] = name
			allowed := isPureImport(name, s.path) || (apiPath && name == "time")
			if !allowed || local == "." || local == "_" {
				add(imp.Pos(), "adapter_import_forbidden")
			}
		}
		ast.Inspect(s.tree, func(node ast.Node) bool {
			sel, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			name := imports[ident.Name]
			apiTime := apiPath && name == "time" && (sel.Sel.Name == "Duration" || sel.Sel.Name == "Second")
			if name != "" && !isPureImport(name, s.path) && !apiTime || name == "fmt" && !pureFormat[sel.Sel.Name] || name == "path/filepath" && !lexicalFilepath[sel.Sel.Name] {
				add(sel.Pos(), "adapter_effect_api_forbidden")
			}
			return true
		})
		for _, group := range s.tree.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:linkname") || strings.HasPrefix(comment.Text, "//go:generate") {
					add(comment.Pos(), "adapter_directive_forbidden")
				}
			}
		}
	}
	return findings, nil
}
