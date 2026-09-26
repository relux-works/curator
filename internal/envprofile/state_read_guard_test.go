package envprofile

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type managerReadFunction struct {
	file           string
	name           string
	key            string
	directory      string
	imports        map[string]string
	dotImports     map[string]bool
	body           *ast.BlockStmt
	objects        map[*ast.Ident]types.Object
	selections     map[*ast.SelectorExpr]*types.Selection
	calls          map[string]bool
	directSeamCall bool
	absenceReads   map[string]bool
	throughSeam    bool
}

type managerReadAudit struct {
	filesScanned int
	sitesScanned int
	seamGuarded  []string
	allowlisted  []string
	violations   []string
}

// Exceptions are deliberately keyed by file and function and explain why the
// absent result belongs to an operator-selected input, source tree, or cache.
// Add entries only after reviewing the read's ownership and fallback behavior.
var reviewedManagerReadAllowlist = reviewedManagerReadExceptions()

func reviewedManagerReadExceptions() map[string]string {
	exceptions := make(map[string]string)
	allow := func(reason string, entries ...string) {
		for _, entry := range entries {
			exceptions[entry] = reason
		}
	}
	allow("Reads an operator-selected hook file to record or compare its current bytes; absence is an input fact and other read failures remain errors.",
		"cmd/curator/hook.go:cli.cmdHookApprove",
		"internal/hookapproval/hookapproval.go:Classify",
	)
	allow("Reads operator-owned configuration or policy input; a documented missing configuration has its own outcome, and other I/O failures propagate.",
		"cmd/curator/main.go:cli.cmdBootstrap",
		"internal/config/config.go:readObject",
		"internal/config/sourcepolicy.go:LoadSourcePolicy",
		"internal/config/sourceproviders.go:LoadSourceProviders",
		"internal/rustsource/pipeline.go:rejectAmbientCargoConfig",
	)
	allow("Reads an operator-selected project declaration or source tree; absence describes missing input, while unreadable and malformed inputs keep distinct errors.",
		"cmd/curator/project_resolve.go:cli.cmdProjectResolve",
		"internal/contextpkg/contextpkg.go:LoadMCP",
		"internal/contextpkg/contextpkg.go:ValidateModules",
		"internal/contextstore/contextstore.go:EnsureState",
		"internal/devsub/devsub.go:LoadManifest",
		"internal/envprofile/envprofile.go:pathManifestDiag",
		"internal/install/generation.go:readDocument",
		"internal/install/install.go:resolveRegistries",
		"internal/manifest/manifest.go:Load",
		"internal/manifest/manifest.go:readObject",
		"internal/swiftpmsource/capture.go:verifySuppliedRootLock",
	)
	allow("Initializes a private object database for one locked Git commit; absence means the isolated replay checkout does not exist yet, while every other Lstat failure stops before fetching.",
		"internal/gitops/gitops.go:FetchCommitFromURLIsolated",
	)
	allow("Inspects a caller-selected source snapshot for declared content; absence means the snapshot lacks that input and other read failures do not count as absence.",
		"internal/audit/audit.go:detect",
		"internal/contextaudit/contextaudit.go:Detect",
		"internal/closure/resolve.go:openGitFrozen",
		"internal/closure/resolve.go:serveGitSubtree",
		"internal/closure/resolve.go:stagePriorFile",
		"internal/gitops/gitops.go:Extract",
		"internal/gitops/gitops.go:destinationFoldsCase",
		"internal/gitops/gitops.go:planWrites",
		"internal/gitops/gitops.go:writeBlobs",
		"internal/godriver/graph.go:validatePackageInputs",
		"internal/godriver/moduleroots.go:readVendorModules",
		"internal/manifest/expand.go:checkOutputBoundary",
		"internal/manifest/expand.go:regularMetadata",
		"internal/manifest/expand.go:rejectDirectorySymlinks",
		"internal/manifest/expand.go:selectionDirectory",
		"internal/moduleroots/moduleroots.go:validateDirectory",
		"internal/skillspec/parse.go:pathExists",
		"internal/skillspec/parse.go:validateLinkFreeDirectory",
		"internal/skillspec/parse.go:validateNearestGoMod",
		"internal/snapshot/boundaries.go:EnumerateInputs",
		"internal/snapshot/boundaries.go:ValidateLocalPackage",
		"internal/snapshot/boundaries.go:ValidateRootInputs",
		"internal/snapshot/boundaries.go:checkDeclaredInputsDisjoint",
		"internal/snapshot/boundaries.go:checkLinkFree",
		"internal/snapshot/boundaries.go:checkRootCoverage",
		"internal/snapshot/boundaries.go:sameFile",
		"internal/snapshot/capture.go:BuildInventory",
		"internal/snapshot/capture.go:PublishLocal",
		"internal/snapshot/rename_noreplace_other.go:renameNoReplace",
		"internal/snapshot/rename_noreplace_other_unix.go:renameNoReplace",
		"internal/snapshot/snapshot.go:Get",
		"internal/snapshot/snapshot.go:publishPreparedSnapshot",
		"internal/staging/boundaries.go:RecheckDurableOne",
		"internal/staging/boundaries.go:Within",
		"internal/staging/boundaries.go:recheckTarget",
	)
	allow("Inspects a caller-owned native environment surface or seed; an absent external file is skipped or reported as missing, and other read errors are preserved.",
		"internal/adapters/stage.go:*Mirror.stageStale",
		"internal/envprofile/managed.go:*ResolveRequest.gatherSeeds",
		"internal/envprofile/managed.go:*verification.checkCopy",
		"internal/envprofile/managed.go:*verification.checkSurfaces",
		"internal/envprofile/managed.go:checkProfileCollision",
		"internal/envprofile/managed.go:claudeProjects",
		"internal/envprofile/managed.go:claudeSeed",
		"internal/envprofile/managed.go:inventoryUnmanaged",
		"internal/envprofile/managed.go:referencedBlocked",
	)
	allow("Checks an external destination or target namespace before a write or removal; ENOENT describes path topology, while other metadata failures abort the operation.",
		"internal/adapters/adapters.go:unmanagedConflict",
		"internal/adapters/boundaries.go:ValidateDestinations",
		"internal/globalbins/globalbins.go:unmanagedConflict",
		"internal/install/commit.go:makeMissingDirectories",
		"internal/install/targets.go:stageStaleSkillRemovals",
		"internal/transaction/digest.go:DigestPath",
		"internal/transaction/digest.go:DigestTarget",
		"internal/transaction/engine.go:*Engine.Prepare",
		"internal/transaction/engine.go:*Engine.buildJournal",
		"internal/transaction/files.go:copyTarget",
		"internal/transaction/journal.go:removeDurablyUnrecorded",
		"internal/transaction/journal.go:removeDurablyWithEntries",
		"internal/transaction/journal.go:removeTreeDurably",
		"internal/transaction/journal.go:validateRemovalStart",
		"internal/transaction/namespace_case_darwin.go:existingNamespaceAncestor",
		"internal/transaction/rename_noreplace_other_unix.go:durableRenameNoReplace",
		"internal/transaction/staging.go:*Engine.createStagingEntry",
		"internal/transaction/staging.go:validatePreparingTarget",
	)
	allow("Checks the manager's disposable build or runtime cache and artifact paths; ENOENT is a cache miss or stale entry, not manager state used as a default.",
		"internal/buildcache/cache.go:*Store.Inspect",
		"internal/buildcache/collect.go:openSweepRoot",
		"internal/buildcache/collect.go:reserveRetiredName",
		"internal/buildcache/publish.go:*Store.quarantinePath",
		"internal/buildcache/publish.go:withdrawnTo",
		"internal/buildrepo/gc.go:Collect",
		"internal/buildrepo/gc.go:collectArtifacts",
		"internal/marker/marker.go:currentBuilds",
		"internal/runtimestore/runtimestore.go:RemoveStaleShims",
		"internal/runtimestore/scripts.go:PrepareScriptRuntime",
		"internal/runtimestore/targets.go:ManagedShimsIn",
		"internal/rustsource/build_execution.go:*managerState.executeBuildCargo",
		"internal/rustsource/build_toolchain.go:sdkFactsFingerprint",
		"internal/rustsource/manager_vendor.go:*managerVendorRunner.stageCargoHome",
		"internal/rustsource/toolchain_registry.go:cargoHostCapabilityReason",
	)
	allow("Checks a private execution workspace, temporary checkout, or staging directory; absence means no prior workspace and all other inspection failures remain errors.",
		"internal/closureexec/portable_runner.go:*ManagerProcessRunner.prepareReplay",
		"internal/closureexec/portable_runner.go:*ManagerProcessRunner.prepareWorkCopy",
		"internal/closureexec/portable_runner.go:ensureEmptyDirectory",
		"internal/closureexec/rename_noreplace_other.go:renameTreeNoReplace",
		"internal/closureexec/source_control_mirror.go:*SourceControlMirrorRunner.Run",
		"internal/closureexec/store.go:preparePrivateDirectory",
		"internal/closureexec/workspace.go:PrepareWorkspace",
		"internal/npmsource/materialize.go:cleanAbsentAbsolute",
		"internal/npmsource/materialize.go:mergeWritableTree",
		"internal/pnpmsource/materialize.go:cleanAbsentAbsolute",
		"internal/pnpmsource/materialize.go:reconcileWritableStoreOverlay",
		"internal/pnpmsource/materialize.go:validateDirectNodeModules",
		"internal/rustsource/build.go:rejectBuildConfiguration",
		"internal/rustsource/pipeline.go:captureAndVendor",
		"internal/swiftpmbuild/build.go:requireEmptyOutputRoot",
		"internal/yarnclassicsource/materialize.go:cleanAbsentAbsolute",
		"internal/yarnmodernsource/capture.go:reconcileCapturedAuthorities",
		"internal/yarnmodernsource/materialize.go:cleanAbsentAbsolute",
	)
	return exceptions
}

func TestManagerOwnedAbsenceReadsAreGuarded(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not locate the test source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(testFile), "../.."))
	audit, err := scanManagerReadSources(root)
	if err != nil {
		t.Fatal(err)
	}
	expectedFiles, err := independentlyCountManagerReadSources(root)
	if err != nil {
		t.Fatal(err)
	}
	if audit.filesScanned != expectedFiles {
		t.Fatalf("manager-read scanner parsed %d production files, independent walk found %d", audit.filesScanned, expectedFiles)
	}
	if audit.filesScanned == 0 || audit.sitesScanned == 0 {
		t.Fatalf("manager-read scan is empty: %d production files, %d relevant readers", audit.filesScanned, audit.sitesScanned)
	}
	covered := len(audit.seamGuarded) + len(audit.allowlisted)
	coveragePercent := 100 * float64(covered) / float64(audit.sitesScanned)
	t.Logf("manager-owned reads: %d/%d covered (%.1f%%): %d guarded via seam, %d allowlisted (%d production files scanned)",
		covered, audit.sitesScanned, coveragePercent, len(audit.seamGuarded), len(audit.allowlisted), audit.filesScanned)
	if covered != audit.sitesScanned {
		t.Errorf("manager-read audit accounts for %d of %d relevant readers", covered, audit.sitesScanned)
	}
	for _, violation := range audit.violations {
		t.Error(violation)
	}
}

func TestManagerReadScannerFindsAliasedNotExistCollapse(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"internal/envprofile", "internal/stateread", "cmd/curator"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test/curator\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mutant := `package envprofile

import (
	errorcheck "errors"
	iofs "io/fs"
	hostfs "os"
)

func readNewManagerState() string {
	data, err := hostfs.ReadFile("manager-state")
	if errorcheck.Is(err, iofs.ErrNotExist) {
		return "default"
	}
	if err != nil {
		return "default"
	}
	return string(data)
}
`
	if err := os.WriteFile(filepath.Join(root, "internal/envprofile/mutant.go"), []byte(mutant), 0o644); err != nil {
		t.Fatal(err)
	}
	audit, err := scanManagerReadSourcesWithAllowlist(root, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.violations) != 1 || !strings.Contains(audit.violations[0], "internal/envprofile/mutant.go:readNewManagerState") {
		t.Fatalf("AST scan of errors.Is(fs.ErrNotExist) mutant = %v, want its one unreviewed reader", audit.violations)
	}
}

func scanManagerReadSources(root string) (managerReadAudit, error) {
	return scanManagerReadSourcesWithAllowlist(root, reviewedManagerReadAllowlist)
}

func scanManagerReadSourcesWithAllowlist(root string, allowlist map[string]string) (managerReadAudit, error) {
	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return managerReadAudit{}, err
	}
	files, err := managerReadSourceFiles(root)
	if err != nil {
		return managerReadAudit{}, err
	}

	fset := token.NewFileSet()
	functions := make([]*managerReadFunction, 0, len(files))
	byKey := make(map[string]*managerReadFunction)
	for _, path := range files {
		parsed, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return managerReadAudit{}, fmt.Errorf("parse %s: %w", path, err)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return managerReadAudit{}, err
		}
		directory := filepath.ToSlash(filepath.Dir(rel))
		imports := importAliases(parsed)
		dotImports := dotImportPaths(parsed)
		objects, selections := resolveLocalObjects(fset, parsed)
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			name := functionDisplayName(fset, function)
			key := directory + "::" + name
			info := &managerReadFunction{
				file:         filepath.ToSlash(rel),
				name:         name,
				key:          key,
				directory:    directory,
				imports:      imports,
				dotImports:   dotImports,
				body:         function.Body,
				objects:      objects,
				selections:   selections,
				calls:        make(map[string]bool),
				absenceReads: make(map[string]bool),
			}
			collectFunctionReads(info, function, modulePath)
			functions = append(functions, info)
			byKey[key] = info
		}
	}

	// Resolve local and imported calls to the shared seam transitively. This
	// counts callers routed through existing typed wrappers such as envmarker.Read.
	for changed := true; changed; {
		changed = false
		for _, function := range functions {
			if function.throughSeam {
				continue
			}
			if function.directSeamCall {
				function.throughSeam = true
				changed = true
				continue
			}
			for target := range function.calls {
				if callee := byKey[target]; callee != nil && callee.throughSeam {
					function.throughSeam = true
					changed = true
					break
				}
			}
		}
	}

	var audit managerReadAudit
	audit.filesScanned = len(files)
	usedAllowlist := make(map[string]bool)
	for _, function := range functions {
		entry := function.file + ":" + function.name
		if len(function.absenceReads) > 0 {
			audit.sitesScanned++
			reason, allowed := allowlist[entry]
			if !allowed || strings.TrimSpace(reason) == "" || strings.ContainsAny(reason, "\r\n") {
				reads := sortedManagerReadKeys(function.absenceReads)
				audit.violations = append(audit.violations, fmt.Sprintf("%s tests not-exist after %s without a seam route or reviewed allowlist reason", entry, strings.Join(reads, ", ")))
				continue
			}
			usedAllowlist[entry] = true
			audit.allowlisted = append(audit.allowlisted, entry+" — "+reason)
			continue
		}
		if function.throughSeam {
			audit.sitesScanned++
			audit.seamGuarded = append(audit.seamGuarded, entry)
		}
	}
	for entry := range allowlist {
		if !usedAllowlist[entry] {
			audit.violations = append(audit.violations, "reviewed allowlist entry has no matching absence-sensitive reader: "+entry)
		}
	}
	sort.Strings(audit.seamGuarded)
	sort.Strings(audit.allowlisted)
	sort.Strings(audit.violations)
	return audit, nil
}

func collectFunctionReads(info *managerReadFunction, function *ast.FuncDecl, modulePath string) {
	var reads []*ast.CallExpr
	calledFunctions := make(map[ast.Expr]bool)
	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		calledFunctions[call.Fun] = true
		_, isRead := filesystemReadCall(call, info.imports, info.dotImports, info.selections)
		if isRead {
			reads = append(reads, call)
		}
		if target, directSeam := calledFunction(call, info, modulePath); directSeam {
			info.directSeamCall = true
		} else if target != "" {
			info.calls[target] = true
		}
		return true
	})
	ast.Inspect(function.Body, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok || calledFunctions[selector] {
			return true
		}
		if api, isRead := filesystemReadSelector(selector, info.imports, info.selections); isRead {
			info.absenceReads["function-value."+api] = true
		}
		return true
	})

	notExistVariables := make(map[types.Object]bool)
	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		for _, object := range notExistErrorObjects(call, info.imports, info.dotImports, info.objects) {
			notExistVariables[object] = true
		}
		return true
	})
	for _, call := range reads {
		api, _ := filesystemReadCall(call, info.imports, info.dotImports, info.selections)
		for object := range assignedErrorObjects(function.Body, call, info.objects) {
			if notExistVariables[object] {
				info.absenceReads[api] = true
			}
		}
	}
}

func filesystemReadCall(call *ast.CallExpr, imports map[string]string, dotImports map[string]bool, selections map[*ast.SelectorExpr]*types.Selection) (string, bool) {
	readMethods := map[string]bool{"ReadFile": true, "ReadDir": true, "Stat": true, "Lstat": true, "Open": true}
	fsFunctions := map[string]bool{"ReadFile": true, "ReadDir": true, "Stat": true, "WalkDir": true, "Glob": true}
	switch function := call.Fun.(type) {
	case *ast.SelectorExpr:
		return filesystemReadSelector(function, imports, selections)
	case *ast.Ident:
		if (dotImports["os"] && readMethods[function.Name]) || (dotImports["io/fs"] && fsFunctions[function.Name]) {
			return "dot-import." + function.Name, true
		}
	}
	return "", false
}

func filesystemReadSelector(function *ast.SelectorExpr, imports map[string]string, selections map[*ast.SelectorExpr]*types.Selection) (string, bool) {
	readMethods := map[string]bool{"ReadFile": true, "ReadDir": true, "Stat": true, "Lstat": true, "Open": true}
	fsFunctions := map[string]bool{"ReadFile": true, "ReadDir": true, "Stat": true, "WalkDir": true, "Glob": true}
	if !readMethods[function.Sel.Name] && !fsFunctions[function.Sel.Name] {
		return "", false
	}
	if qualifier, ok := function.X.(*ast.Ident); ok {
		if importPath, imported := imports[qualifier.Name]; imported {
			switch importPath {
			case "os":
				if readMethods[function.Sel.Name] {
					return "os." + function.Sel.Name, true
				}
			case "io/fs":
				if fsFunctions[function.Sel.Name] {
					return "io/fs." + function.Sel.Name, true
				}
			}
			return "", false
		}
	}
	// io/fs interfaces expose Open/ReadFile/ReadDir methods on an FS value.
	// Inspect the resolved selection so unrelated methods with the same name,
	// such as archive/zip.File.Open, are not treated as manager filesystem reads.
	selection := selections[function]
	if selection == nil || selection.Obj().Pkg() == nil {
		return "", false
	}
	methodPackage := selection.Obj().Pkg().Path()
	switch {
	case methodPackage == "os" && readMethods[function.Sel.Name]:
		return "os method." + function.Sel.Name, true
	case methodPackage == "io/fs" && fsFunctions[function.Sel.Name]:
		return "io/fs method." + function.Sel.Name, true
	default:
		return "", false
	}
}

func notExistErrorObjects(call *ast.CallExpr, imports map[string]string, dotImports map[string]bool, objects map[*ast.Ident]types.Object) []types.Object {
	var errorExpression ast.Expr
	switch function := call.Fun.(type) {
	case *ast.SelectorExpr:
		qualifier, ok := function.X.(*ast.Ident)
		if !ok {
			return nil
		}
		importPath := imports[qualifier.Name]
		switch {
		case importPath == "os" && function.Sel.Name == "IsNotExist" && len(call.Args) == 1:
			errorExpression = call.Args[0]
		case importPath == "errors" && function.Sel.Name == "Is" && len(call.Args) == 2 && isNotExistSentinel(call.Args[1], imports, dotImports):
			errorExpression = call.Args[0]
		}
	case *ast.Ident:
		switch {
		case function.Name == "IsNotExist" && dotImports["os"] && len(call.Args) == 1:
			errorExpression = call.Args[0]
		case function.Name == "Is" && dotImports["errors"] && len(call.Args) == 2 && isNotExistSentinel(call.Args[1], imports, dotImports):
			errorExpression = call.Args[0]
		}
	}
	if errorExpression == nil {
		return nil
	}
	var found []types.Object
	ast.Inspect(errorExpression, func(node ast.Node) bool {
		if ident, ok := node.(*ast.Ident); ok {
			if object := objects[ident]; object != nil {
				found = append(found, object)
			}
		}
		return true
	})
	return found
}

func isNotExistSentinel(expression ast.Expr, imports map[string]string, dotImports map[string]bool) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "ErrNotExist" {
		return false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	return imports[qualifier.Name] == "io/fs" || imports[qualifier.Name] == "os" ||
		(dotImports["io/fs"] && qualifier.Name == "fs") || (dotImports["os"] && qualifier.Name == "os")
}

func assignedErrorObjects(body *ast.BlockStmt, target *ast.CallExpr, objects map[*ast.Ident]types.Object) map[types.Object]bool {
	assigned := make(map[types.Object]bool)
	ast.Inspect(body, func(node ast.Node) bool {
		switch statement := node.(type) {
		case *ast.AssignStmt:
			if len(statement.Rhs) != 1 || statement.Rhs[0] != target || len(statement.Lhs) < 2 {
				return true
			}
			if ident, ok := statement.Lhs[len(statement.Lhs)-1].(*ast.Ident); ok {
				if object := objects[ident]; object != nil {
					assigned[object] = true
				}
			}
		case *ast.ValueSpec:
			if len(statement.Values) != 1 || statement.Values[0] != target || len(statement.Names) < 2 {
				return true
			}
			if object := objects[statement.Names[len(statement.Names)-1]]; object != nil {
				assigned[object] = true
			}
		}
		return true
	})
	return assigned
}

func resolveLocalObjects(fset *token.FileSet, file *ast.File) (map[*ast.Ident]types.Object, map[*ast.SelectorExpr]*types.Selection) {
	info := &types.Info{
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	config := types.Config{
		Importer: scannerImporter{fallback: importer.Default()},
		Error:    func(error) {},
	}
	_, _ = config.Check("managerread.scan", fset, []*ast.File{file}, info)
	objects := make(map[*ast.Ident]types.Object, len(info.Defs)+len(info.Uses))
	for ident, object := range info.Defs {
		objects[ident] = object
	}
	for ident, object := range info.Uses {
		objects[ident] = object
	}
	return objects, info.Selections
}

type scannerImporter struct {
	fallback types.Importer
}

func (i scannerImporter) Import(path string) (*types.Package, error) {
	if imported, err := i.fallback.Import(path); err == nil {
		return imported, nil
	}
	placeholder := types.NewPackage(path, filepath.Base(path))
	placeholder.MarkComplete()
	return placeholder, nil
}

func calledFunction(call *ast.CallExpr, function *managerReadFunction, modulePath string) (string, bool) {
	switch target := call.Fun.(type) {
	case *ast.Ident:
		if function.dotImports[modulePath+"/internal/stateread"] && isStateReadSeamFunction(target.Name) {
			return "", true
		}
		return function.directory + "::" + target.Name, false
	case *ast.SelectorExpr:
		qualifier, ok := target.X.(*ast.Ident)
		if !ok {
			return "", false
		}
		importPath, imported := function.imports[qualifier.Name]
		if !imported {
			return "", false
		}
		seamPath := modulePath + "/internal/stateread"
		if importPath == seamPath && isStateReadSeamFunction(target.Sel.Name) {
			return "", true
		}
		prefix := modulePath + "/"
		if strings.HasPrefix(importPath, prefix) {
			directory := strings.TrimPrefix(importPath, prefix)
			return filepath.ToSlash(directory) + "::" + target.Sel.Name, false
		}
	}
	return "", false
}

func isStateReadSeamFunction(name string) bool {
	switch name {
	case "ReadFile", "ReadDir", "Stat", "Lstat":
		return true
	default:
		return false
	}
}

func importAliases(file *ast.File) map[string]string {
	aliases := make(map[string]string)
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		if spec.Name != nil && spec.Name.Name != "." && spec.Name.Name != "_" {
			aliases[spec.Name.Name] = importPath
		} else if spec.Name == nil {
			aliases[filepath.Base(importPath)] = importPath
		}
	}
	return aliases
}

func dotImportPaths(file *ast.File) map[string]bool {
	paths := make(map[string]bool)
	for _, spec := range file.Imports {
		if spec.Name == nil || spec.Name.Name != "." {
			continue
		}
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err == nil {
			paths[importPath] = true
		}
	}
	return paths
}

func functionDisplayName(fset *token.FileSet, function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) == 0 {
		return function.Name.Name
	}
	var receiver strings.Builder
	if err := formatNode(&receiver, fset, function.Recv.List[0].Type); err != nil {
		return function.Name.Name
	}
	return receiver.String() + "." + function.Name.Name
}

func formatNode(builder *strings.Builder, fset *token.FileSet, node ast.Node) error {
	return format.Node(builder, fset, node)
}

func readModulePath(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open module file %s: %w", path, err)
	}
	scanner := bufio.NewScanner(file)
	modulePath := ""
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && fields[0] == "module" {
			modulePath = fields[1]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close module file %s: %w", path, err)
	}
	if modulePath == "" {
		return "", fmt.Errorf("module path missing from %s", path)
	}
	return modulePath, nil
}

func managerReadSourceFiles(root string) ([]string, error) {
	var files []string
	for _, relativeRoot := range []string{"internal", "cmd"} {
		base := filepath.Join(root, relativeRoot)
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if entry.IsDir() && filepath.ToSlash(relative) == "internal/stateread" {
				return filepath.SkipDir
			}
			if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", base, err)
		}
	}
	sort.Strings(files)
	return files, nil
}

// This second walk independently checks the scan denominator so accidentally
// narrowing the parser's input list cannot still produce a green ratio.
func independentlyCountManagerReadSources(root string) (int, error) {
	count := 0
	for _, relativeRoot := range []string{"internal", "cmd"} {
		base := filepath.Join(root, relativeRoot)
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relativeSlash := filepath.ToSlash(relative)
			if entry.IsDir() && (relativeSlash == "internal/stateread" || strings.HasPrefix(relativeSlash, "internal/stateread/")) {
				return filepath.SkipDir
			}
			if !entry.IsDir() && filepath.Ext(path) == ".go" && !strings.HasSuffix(path, "_test.go") {
				count++
			}
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("count %s: %w", base, err)
		}
	}
	return count, nil
}

func sortedManagerReadKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
