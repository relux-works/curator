package swiftpminterop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

type componentBinding struct {
	component ExternalComponent
	node      closuregraph.Node
	nodeID    closuregraph.ID
	system    string
}

type closeState struct {
	config       Config
	capture      *swiftpmsource.Capture
	destination  swiftpmsource.Destination
	markers      map[string]string
	roots        roots
	packages     map[string]swiftpmsource.PackageEvidence
	order        []string
	packageRoots map[string]string
	nodesByKey   map[string]closuregraph.Node
	idsByKey     map[string]closuregraph.ID
	selected     map[closuregraph.ID]bool
	components   []componentBinding
	byRole       map[string]componentBinding
	systemByKey  map[string]componentBinding
	targets      []TargetInterop
	byTargetKey  map[string]int
	// settingDefines is the second input to each selected target's macro
	// oracle: the macro names its `.define` build settings bind, which no
	// source `#define` spells. See disposeBuildSettings.
	settingDefines map[string]map[string]bool
}

// declaredTarget is one target declaration reachable from the selected product
// when every condition is ignored. Capture is built from this condition-neutral
// superset; which member the destination actually selects is a binding fact.
type declaredTarget struct {
	pkg, name string
	key       string
	target    swiftpmsource.Target
	nodeID    closuregraph.ID
}

// targetEdge is one declared target-to-target edge with the selection-neutral
// condition that governs it.
type targetEdge struct {
	index     int
	condition *closuregraph.Condition
}

// Close validates the C-family, module, header, system-library, and interop
// boundaries of an accepted SwiftPM source closure and republishes it as an
// extended capture graph with an exact selection binding. It fails before any
// downstream compiler, plugin, macro, or extension can run.
func Close(ctx context.Context, config Config, capture *swiftpmsource.Capture) (*Result, error) {
	if capture == nil {
		return nil, fail(CodeDerivationUnauthorized, "interop closure requires an accepted SwiftPM capture")
	}
	state := &closeState{config: config, capture: capture, destination: capture.Destination(), packages: map[string]swiftpmsource.PackageEvidence{}, packageRoots: map[string]string{}, nodesByKey: map[string]closuregraph.Node{}, idsByKey: map[string]closuregraph.ID{}, selected: map[closuregraph.ID]bool{}, byRole: map[string]componentBinding{}, systemByKey: map[string]componentBinding{}, byTargetKey: map[string]int{}, settingDefines: map[string]map[string]bool{}}
	state.markers = state.destination.Markers
	if err := validateConfig(config, state.destination); err != nil {
		return nil, err
	}
	if err := recheck(ctx, config, config.Clang); err != nil {
		return nil, err
	}
	if config.ClangCXX.Role != "" {
		if err := recheck(ctx, config, config.ClangCXX); err != nil {
			return nil, err
		}
	}
	if err := state.indexCapture(); err != nil {
		return nil, err
	}
	if err := state.bindComponents(); err != nil {
		return nil, err
	}
	if err := state.classifyTargets(); err != nil {
		return nil, err
	}
	if err := state.inspectTargets(); err != nil {
		return nil, err
	}
	boundaries, err := state.deriveBoundaries()
	if err != nil {
		return nil, err
	}
	reads, err := state.verifyReads(ctx)
	if err != nil {
		return nil, err
	}
	return state.publish(boundaries, reads)
}

func (state *closeState) indexCapture() error {
	for _, node := range state.capture.Records.CaptureNodes {
		id, err := node.ID()
		if err != nil {
			return err
		}
		if _, duplicate := state.idsByKey[node.LogicalKey]; duplicate {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"logical_key": node.LogicalKey}, "accepted capture contains a duplicate logical key")
		}
		state.nodesByKey[node.LogicalKey], state.idsByKey[node.LogicalKey] = node, id
	}
	for _, id := range state.capture.TargetNodeIDs {
		state.selected[id] = true
	}
	if len(state.selected) == 0 {
		return fail(CodeGraphIncomplete, "accepted capture selected no target")
	}
	for _, pkg := range state.capture.Packages {
		if _, duplicate := state.packages[pkg.Identity]; duplicate {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"package": pkg.Identity}, "accepted capture contains a duplicate package identity")
		}
		root, err := pkg.ProtectedRoot()
		if err != nil {
			return err
		}
		state.packages[pkg.Identity] = pkg
		state.packageRoots[pkg.Identity] = root
		state.order = append(state.order, pkg.Identity)
		if err = state.roots.addAdmitted(pkg.Identity, root); err != nil {
			return err
		}
	}
	sort.Strings(state.order)
	return nil
}

func (state *closeState) bindComponents() error {
	add := func(component ExternalComponent, systemKey string) error {
		node, id, err := componentNode(component)
		if err != nil {
			return err
		}
		if existing, duplicate := state.byRole[component.Role]; duplicate {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"role": component.Role, "existing": string(existing.nodeID)}, "duplicate selected component role")
		}
		binding := componentBinding{component: component, node: node, nodeID: id, system: systemKey}
		state.components = append(state.components, binding)
		state.byRole[component.Role] = binding
		if systemKey != "" {
			state.systemByKey[systemKey] = binding
		}
		for _, root := range component.Roots {
			if err = state.roots.addTrusted(component.Role, root, id); err != nil {
				return err
			}
		}
		return nil
	}
	if err := add(componentFromTool(state.config.Clang), ""); err != nil {
		return err
	}
	if state.config.ClangCXX.Role != "" {
		if err := add(componentFromTool(state.config.ClangCXX), ""); err != nil {
			return err
		}
	}
	if err := add(state.config.SDK, ""); err != nil {
		return err
	}
	if state.config.Sysroot != nil {
		if err := add(*state.config.Sysroot, ""); err != nil {
			return err
		}
	}
	libraries := append([]SystemLibrary(nil), state.config.SystemLibraries...)
	sort.Slice(libraries, func(i, j int) bool {
		return libraries[i].Package+":"+libraries[i].Target < libraries[j].Package+":"+libraries[j].Target
	})
	for _, library := range libraries {
		if err := add(library.Component, library.Package+":"+library.Target); err != nil {
			return err
		}
	}
	return nil
}

// classifyTargets classifies every C-family and Swift target the selected
// product declares, ignoring every condition. Capture must stay
// selection-neutral: a conditional edge that one destination prunes is a
// selected/pruned verdict recorded by the active projection, not a reason for
// two destinations to publish different capture records. Destination-specific
// gates — restricted-language profiles and unsafe settings — still apply, but
// only to the subset the destination actually selects.
func (state *closeState) classifyTargets() error {
	declared, ordered := state.declaredTargets()
	reach := state.declaredReach(declared, ordered)
	for _, entry := range ordered {
		if !reach[entry.key] {
			continue
		}
		selected := state.selected[entry.nodeID]
		switch entry.target.Type {
		case "plugin", "macro", "binary":
			if !selected {
				// A dormant extension declaration stays captured source; it is
				// never compiled, invoked, or classified as a compile target.
				continue
			}
			_, _, err := classifyTarget(entry.pkg, entry.name, entry.target)
			if err == nil {
				err = failFields(CodePluginUnsupported, map[string]string{"target": entry.key}, "selected graph reaches an executable extension target")
			}
			return err
		}
		kind, languages, err := classifyTarget(entry.pkg, entry.name, entry.target)
		if err != nil {
			return err
		}
		if selected {
			if err = requireLanguageProfile(state.config.Profile, state.destination, entry.pkg, entry.name, languages); err != nil {
				return err
			}
			defines, settingErr := state.disposeBuildSettings(entry)
			if settingErr != nil {
				return settingErr
			}
			state.settingDefines[entry.key] = defines
		}
		sources := []string{}
		for _, source := range entry.target.Sources {
			if _, classified := sourceLanguage(source); classified {
				sources = append(sources, source)
			}
		}
		sort.Strings(sources)
		interop := TargetInterop{Package: entry.pkg, Target: entry.name, Kind: kind, Languages: languages, Sources: sources, SourceRoot: entry.target.SourceRoot(), NodeID: entry.nodeID, CxxInteropMode: cxxInteropSelected(entry.target, state.markers), CxxInteropDeclared: cxxInteropDeclared(entry.target), Selected: selected}
		state.byTargetKey[entry.key] = len(state.targets)
		state.targets = append(state.targets, interop)
	}
	if len(state.targets) == 0 {
		return fail(CodeGraphIncomplete, "selected product reaches no classifiable target")
	}
	for key := range state.systemByKey {
		index, ok := state.byTargetKey[key]
		if !ok || state.targets[index].Kind != KindSystem {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"target": key}, "system-library binding names no selected system target")
		}
	}
	return nil
}

// declaredTargets indexes every declared target that the accepted capture
// published as a target unit, in deterministic package and manifest order.
func (state *closeState) declaredTargets() (map[string]declaredTarget, []declaredTarget) {
	declared := map[string]declaredTarget{}
	ordered := []declaredTarget{}
	for _, identity := range state.order {
		pkg := state.packages[identity]
		for _, target := range pkg.Manifest.Targets {
			key := identity + ":" + target.Name
			id, ok := state.idsByKey["swiftpm.target."+identity+"."+target.Name]
			if !ok {
				continue
			}
			entry := declaredTarget{pkg: identity, name: target.Name, key: key, target: target, nodeID: id}
			declared[key] = entry
			ordered = append(ordered, entry)
		}
	}
	return declared, ordered
}

// declaredReach walks every declared target edge from the destination-selected
// seed set with every condition ignored. The result is the condition-neutral
// superset the capture records: it never omits a target merely because the
// current destination prunes the edge that reaches it.
func (state *closeState) declaredReach(declared map[string]declaredTarget, ordered []declaredTarget) map[string]bool {
	queue := []string{}
	for _, entry := range ordered {
		if state.selected[entry.nodeID] {
			queue = append(queue, entry.key)
		}
	}
	reach := map[string]bool{}
	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]
		if reach[key] {
			continue
		}
		reach[key] = true
		entry, ok := declared[key]
		if !ok {
			continue
		}
		switch entry.target.Type {
		case "plugin", "macro", "binary":
			continue
		}
		for _, name := range state.declaredEdgeTargets(entry) {
			if !reach[name] {
				queue = append(queue, name)
			}
		}
	}
	return reach
}

// declaredEdgeTargets resolves one target's declared target and product edges
// to package-qualified target keys, ignoring every condition.
func (state *closeState) declaredEdgeTargets(entry declaredTarget) []string {
	result := []string{}
	for _, dependency := range entry.target.Dependencies {
		identity := entry.pkg
		if dependency.Package != "" {
			identity = strings.ToLower(dependency.Package)
		}
		names := []string{dependency.Name}
		if dependency.Product != "" {
			other, exists := state.packages[identity]
			if !exists {
				continue
			}
			product, exists := findProduct(other.Manifest, dependency.Product)
			if !exists {
				continue
			}
			names = product.Targets
		}
		for _, name := range names {
			if name == "" {
				continue
			}
			result = append(result, identity+":"+name)
		}
	}
	return result
}

func (state *closeState) inspectTargets() error {
	for index := range state.targets {
		interop := &state.targets[index]
		pkg := state.packages[interop.Package]
		target, ok := findTarget(pkg.Manifest, interop.Target)
		if !ok {
			return failFields(CodeGraphIncomplete, map[string]string{"target": interop.Package + ":" + interop.Target}, "selected target is absent from its accepted manifest")
		}
		if interop.Kind == KindSystem {
			if err := state.inspectSystemTarget(interop); err != nil {
				return err
			}
			continue
		}
		root := state.packageRoots[interop.Package]
		if interop.Kind == KindClang {
			headerRoot, err := publicHeaderRoot(interop.Package, interop.Target, target.Path, target.PublicHeadersPath)
			if err != nil {
				return err
			}
			interop.PublicHeaderRoot = headerRoot
			if err = confineModuleMapLayout(interop.Package, interop.Target, root, targetTreeRoot(target, interop), headerRoot); err != nil {
				return err
			}
			headers, err := inventoryHeaders(root, interop.PublicHeaderRoot)
			if err != nil {
				return err
			}
			interop.Headers = headers
			evidence, err := state.moduleMapEvidence(interop, headers)
			if err != nil {
				return err
			}
			interop.ModuleMap = evidence
		}
	}
	for index := range state.targets {
		if err := state.scanAndConfineIncludes(&state.targets[index]); err != nil {
			return err
		}
	}
	return nil
}

func (state *closeState) inspectSystemTarget(interop *TargetInterop) error {
	key := interop.Package + ":" + interop.Target
	binding, ok := state.systemByKey[key]
	if !ok {
		return failFields(CodeToolchainUntrusted, map[string]string{"target": key}, "system-library target has no Curator-selected external binding")
	}
	library := SystemLibrary{}
	for _, candidate := range state.config.SystemLibraries {
		if candidate.Package+":"+candidate.Target == key {
			library = candidate
		}
	}
	resolution := state.roots.resolve(library.ModuleMapPath)
	if resolution.Class != ResolvedBinding || resolution.BindingNode != binding.nodeID {
		return failFields(CodeToolchainUntrusted, map[string]string{"target": key, "module_map": library.ModuleMapPath}, "system-library module map is outside its selected SDK, sysroot, or toolchain root")
	}
	payload, err := readTrustedFile(resolution.AbsolutePath)
	if err != nil {
		return failFields(CodeToolchainUntrusted, map[string]string{"target": key, "module_map": library.ModuleMapPath}, "selected system module map is unreadable")
	}
	parsed, err := ParseModuleMap(library.ModuleMapPath, payload)
	if err != nil {
		return err
	}
	baseDirectory := path.Dir(resolution.AbsolutePath)
	resolved := []Resolution{}
	for _, reference := range parsed.References {
		reference := reference
		result, referenceErr := state.roots.resolveReference(baseDirectory, reference.Path)
		if referenceErr != nil {
			return referenceErr
		}
		if result.Class != ResolvedBinding {
			return failFields(CodeToolchainUntrusted, map[string]string{"target": key, "reference": reference.Path}, "system module map names a path outside every selected external root")
		}
		resolved = append(resolved, result)
	}
	if err = state.confineLinks(key, parsed, &binding); err != nil {
		return err
	}
	interop.ModuleMap = &ModuleMapEvidence{Package: interop.Package, Target: interop.Target, Relative: library.ModuleMapPath, GrammarID: ModuleMapGrammarID, SHA256: digestBytes(payload), Parsed: parsed, ResolvedRefs: resolved}
	interop.ToolNodeID = binding.nodeID
	return nil
}

func (state *closeState) moduleMapEvidence(interop *TargetInterop, headers []HeaderFile) (*ModuleMapEvidence, error) {
	root := state.packageRoots[interop.Package]
	custom := ""
	for _, header := range headers {
		if isModuleMap(header.Relative) && path.Dir(header.Relative) == path.Clean(interop.PublicHeaderRoot) {
			custom = header.Relative
			break
		}
	}
	if custom != "" {
		payload, err := readAdmittedFile(root, custom)
		if err != nil {
			return nil, failFields(CodeModuleMapEscape, map[string]string{"module_map": custom}, "declared module map is unreadable")
		}
		parsed, err := ParseModuleMap(custom, payload)
		if err != nil {
			return nil, err
		}
		resolved, err := state.confineModuleMapReferences(interop, root, custom, parsed)
		if err != nil {
			return nil, err
		}
		if err = state.confineLinks(interop.Package+":"+interop.Target, parsed, nil); err != nil {
			return nil, err
		}
		return &ModuleMapEvidence{Package: interop.Package, Target: interop.Target, Relative: custom, GrammarID: ModuleMapGrammarID, SHA256: digestBytes(payload), Parsed: parsed, ResolvedRefs: resolved, PublicHeaderRoot: interop.PublicHeaderRoot, PublicHeaderFiles: headers}, nil
	}
	generated, err := GenerateModuleMap(interop.Target, interop.PublicHeaderRoot, headers)
	if err != nil {
		return nil, err
	}
	logical := path.Join(interop.PublicHeaderRoot, "module.modulemap")
	parsed, err := ParseModuleMap(logical, []byte(generated))
	if err != nil {
		return nil, err
	}
	resolved, err := state.confineModuleMapReferences(interop, root, logical, parsed)
	if err != nil {
		return nil, err
	}
	return &ModuleMapEvidence{Package: interop.Package, Target: interop.Target, Relative: logical, Generated: true, GrammarID: GeneratedModuleMapGrammarID, SHA256: digestBytes([]byte(generated)), Parsed: parsed, ResolvedRefs: resolved, PublicHeaderRoot: interop.PublicHeaderRoot, PublicHeaderFiles: headers}, nil
}

func (state *closeState) confineModuleMapReferences(interop *TargetInterop, root, logical string, parsed ModuleMap) ([]Resolution, error) {
	return state.confineModuleMapClosure(interop, root, logical, parsed, map[string]bool{logical: true})
}

// confineModuleMapClosure confines one module map and every admitted module map
// it names through `extern module`. An extern map was admitted by path and then
// never parsed, so every header it declared — and Clang does read them while
// building that module — reached neither confinement nor the include scan.
func (state *closeState) confineModuleMapClosure(interop *TargetInterop, root, logical string, parsed ModuleMap, visited map[string]bool) ([]Resolution, error) {
	key := interop.Package + ":" + interop.Target
	baseDirectory := path.Join(root, path.Dir(logical))
	resolved := []Resolution{}
	for _, reference := range parsed.References {
		result, err := state.roots.resolveReference(baseDirectory, reference.Path)
		if err != nil {
			return nil, err
		}
		if result.Class == ResolvedAdmitted && result.Package != interop.Package {
			return nil, failFields(CodeModuleMapEscape, map[string]string{"target": key, "reference": reference.Path, "resolved_package": result.Package}, "module map reaches another package's tree without a declared target edge")
		}
		if reference.Kind == ReferenceExternModule && result.Class != ResolvedAdmitted {
			return nil, failFields(CodeModuleMapEscape, map[string]string{"target": key, "reference": reference.Path}, "extern module names a map outside the admitted package")
		}
		resolved = append(resolved, result)
		if reference.Kind != ReferenceExternModule || visited[result.Relative] {
			continue
		}
		visited[result.Relative] = true
		payload, readErr := readAdmittedFile(root, result.Relative)
		if readErr != nil {
			return nil, failFields(CodeModuleMapEscape, map[string]string{"target": key, "reference": reference.Path}, "extern module map is unreadable in the admitted tree")
		}
		nested, parseErr := ParseModuleMap(result.Relative, payload)
		if parseErr != nil {
			return nil, parseErr
		}
		if err = state.confineLinks(key, nested, nil); err != nil {
			return nil, err
		}
		nestedRefs, nestedErr := state.confineModuleMapClosure(interop, root, result.Relative, nested, visited)
		if nestedErr != nil {
			return nil, nestedErr
		}
		resolved = append(resolved, nestedRefs...)
	}
	return resolved, nil
}

// confineLinks proves that every module-map link edge names a library or
// framework already declared by a selected external component. Provider hints
// and arbitrary host libraries confer no trust.
func (state *closeState) confineLinks(targetKey string, parsed ModuleMap, owner *componentBinding) error {
	for _, link := range parsed.Links {
		if state.linkDeclared(link, owner) {
			continue
		}
		return failFields(CodeToolchainUntrusted, map[string]string{"target": targetKey, "link": link.Name, "framework": boolText(link.Framework)}, "module map links a library or framework that no selected external component declares")
	}
	return nil
}

func (state *closeState) linkDeclared(link Link, owner *componentBinding) bool {
	candidates := state.components
	if owner != nil {
		candidates = append([]componentBinding{*owner}, state.components...)
	}
	for _, candidate := range candidates {
		values := candidate.component.Libraries
		if link.Framework {
			values = candidate.component.Frameworks
		}
		if containsString(values, link.Name) {
			return true
		}
	}
	return false
}

// sourceUnit is one admitted file this stage opens and scans, named by the
// package that owns it so a reference that resolves into a dependency's tree
// is scanned against that package's root.
type sourceUnit struct{ pkg, relative string }

// scanAndConfineIncludes scans the transitive closure of one target's declared
// C-family sources, public headers, and every admitted include they reach, and
// proves each reference resolves to admitted source or exactly one selected
// binding component. Scanning only the declared sources and the public-header
// root left the ordinary private-header layout admitted but never opened, so
// the directives of `Sources/CLib/private.h` were invisible; in portable mode
// this declared closure is the entire header proof.
func (state *closeState) scanAndConfineIncludes(interop *TargetInterop) error {
	if interop.Kind == KindSystem {
		return nil
	}
	searchRoots := state.includeSearchRoots(interop)
	queue := []sourceUnit{}
	for _, source := range interop.Sources {
		// Swift has no textual inclusion: the Swift compiler never performs
		// preprocessing on it, so applying the C grammar there would reject
		// ordinary Swift `#` syntax without proving anything.
		if language, classified := sourceLanguage(source); classified && language == LanguageSwift {
			continue
		}
		queue = append(queue, sourceUnit{pkg: interop.Package, relative: source})
	}
	for _, header := range headerPaths(interop.Headers) {
		queue = append(queue, sourceUnit{pkg: interop.Package, relative: header})
	}
	moduleSeeds, err := state.moduleMapSeeds(interop)
	if err != nil {
		return err
	}
	queue = append(queue, moduleSeeds...)
	visited := map[sourceUnit]bool{}
	scanned := []IncludeReference{}
	// The macro oracle is the union of two sources: every source `#define` the
	// scanned closure binds, and every `.define` build setting the destination
	// selects for this target. A build-setting define reaches the compiler as
	// `-D` without appearing in any admitted file, so an oracle built from
	// source alone leaves both closing rules blind to it.
	macros := map[string]bool{}
	for name := range state.settingDefines[interop.Package+":"+interop.Target] {
		macros[name] = true
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		result, scanErr := scanIncludes(interop.Package, interop.Target, current.pkg, state.packageRoots[current.pkg], current.relative)
		if scanErr != nil {
			return scanErr
		}
		for _, reference := range result.references {
			next, confineErr := state.confineInclude(interop, current, searchRoots, reference)
			if confineErr != nil {
				return confineErr
			}
			if next != nil && !visited[*next] {
				queue = append(queue, *next)
			}
		}
		scanned = append(scanned, result.references...)
		for name := range result.macros {
			macros[name] = true
		}
	}
	if err := rejectMacroDefinedModuleNames(interop.Package+":"+interop.Target, scanned, macros); err != nil {
		return err
	}
	sort.Slice(scanned, func(i, j int) bool {
		left := scanned[i].SourcePackage + "\x00" + scanned[i].Source + "\x00" + scanned[i].Spelling
		right := scanned[j].SourcePackage + "\x00" + scanned[j].Source + "\x00" + scanned[j].Spelling
		return left < right
	})
	interop.Includes = scanned
	return nil
}

// rejectMacroDefinedModuleNames closes the one identifier position this scanner
// records that the compiler resolves after macro expansion.
//
// `confineInclude` gates a module import by its recorded spelling and the
// closure records that spelling as the module the target reads. For `@import`
// and for the Microsoft `__pragma(clang module import …)` operator neither
// holds when the name is macro-defined: verified on the accepted Darwin
// profile, `#define NoSuchKitXYZ SecretKit` with `@import NoSuchKitXYZ;` builds
// SecretKit and reads its header, so the `moduleDeclared` gate is satisfied by
// a name the compiler never resolves and the retained evidence names the wrong
// module. That is an evidence-integrity defect, not a channel-keyword one, and
// no rule about which keywords may follow `@` reaches it.
//
// Portable mode does not expand the name — that would be a macro expander it
// has no business carrying — it refuses the construct: an `@import` whose name
// this stage cannot prove is the module the compiler resolves is rejected. The
// literal, non-macro spelling, which is every legitimate use, still admits.
//
// The macro set is the union over the target's whole scanned closure rather
// than per translation unit, because a definition in a header reaching a `.c`
// file is the realistic vector and the scanner does not model conditional
// inclusion. That over-approximates in the safe direction: at worst a literal
// import is rejected because an unrelated file of the same target binds a macro
// of that name.
//
// The set also unions the target's selected `.define` build settings, which
// bind a macro no admitted file spells: verified on the accepted Darwin
// profile, `-DNoSuchKitXYZ=SecretKit` with `@import NoSuchKitXYZ;` builds
// SecretKit and reads its header exactly as the source `#define` does. See
// disposeBuildSettings for the enumerated build-setting kind axis behind it.
func rejectMacroDefinedModuleNames(key string, references []IncludeReference, macros map[string]bool) error {
	for _, reference := range references {
		if !reference.ModuleImport || !reference.ExpandedName {
			continue
		}
		for _, component := range strings.Split(reference.Spelling, ".") {
			if !macros[component] {
				continue
			}
			return failFields(CodeHeaderInputUndeclared, map[string]string{"target": key, "source": reference.Source, "module": reference.Spelling, "macro": component}, "Clang module import names a macro-defined identifier the compiler expands before resolving, so the module this target reads is not the one the source names")
		}
	}
	return nil
}

// moduleMapSeeds returns every admitted file the target's module map declares.
// `confineModuleMapReferences` admits an in-package reference that resolves
// outside the public-header root, and the public-header inventory never lists
// such a file, so the module's own header set joined no scan and every
// directive it declared was invisible. Clang really does read those headers: it
// opens each module member while building the module from that map.
func (state *closeState) moduleMapSeeds(interop *TargetInterop) ([]sourceUnit, error) {
	if interop.ModuleMap == nil {
		return nil, nil
	}
	key := interop.Package + ":" + interop.Target
	units := []sourceUnit{}
	for _, resolution := range interop.ModuleMap.ResolvedRefs {
		if resolution.Class != ResolvedAdmitted {
			continue
		}
		info, err := os.Lstat(resolution.AbsolutePath)
		if err != nil {
			return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": key, "reference": resolution.Relative}, "admitted module-map reference is absent from the admitted tree")
		}
		if !info.IsDir() {
			// An extern module map is confined and parsed by the module-map
			// closure, not scanned with the C preprocessor grammar.
			if !isModuleMap(resolution.Relative) {
				units = append(units, sourceUnit{pkg: resolution.Package, relative: resolution.Relative})
			}
			continue
		}
		members, err := inventoryHeaders(state.packageRoots[resolution.Package], resolution.Relative)
		if err != nil {
			return nil, err
		}
		for _, member := range members {
			if isModuleMap(member.Relative) {
				continue
			}
			units = append(units, sourceUnit{pkg: resolution.Package, relative: member.Relative})
		}
	}
	sort.Slice(units, func(i, j int) bool {
		return units[i].pkg+"\x00"+units[i].relative < units[j].pkg+"\x00"+units[j].relative
	})
	return units, nil
}

// confineInclude classifies one scanned reference. An admitted resolution is
// returned so the caller scans it in turn: a reference this stage admits but
// never opens would hide every directive that file declares.
func (state *closeState) confineInclude(interop *TargetInterop, current sourceUnit, searchRoots []string, reference IncludeReference) (*sourceUnit, error) {
	key := interop.Package + ":" + interop.Target
	if reference.ModuleImport {
		if state.moduleDeclared(interop, reference.Spelling) {
			return nil, nil
		}
		return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": key, "source": reference.Source, "module": reference.Spelling}, "Clang module import resolves to no admitted module map or selected external module")
	}
	if strings.ContainsAny(reference.Spelling, "\x00\r\n") {
		return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": key, "spelling": reference.Spelling}, "include spelling contains a control character")
	}
	if isAbsoluteSpelling(reference.Spelling) {
		return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": key, "source": reference.Source, "spelling": reference.Spelling}, "source includes an absolute header path")
	}
	bases := searchRoots
	if !reference.Angled {
		bases = append([]string{path.Join(state.packageRoots[current.pkg], path.Dir(current.relative))}, searchRoots...)
	}
	if resolution, resolved := state.resolveIncludeSpelling(bases, reference.Spelling); resolved {
		return &sourceUnit{pkg: resolution.Package, relative: resolution.Relative}, nil
	}
	if state.systemHeaderDeclared(reference.Spelling) {
		return nil, nil
	}
	return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": key, "source": reference.Source, "spelling": reference.Spelling}, "include resolves outside the admitted closure and every selected binding root")
}

func (state *closeState) includeSearchRoots(interop *TargetInterop) []string {
	values := []string{}
	if interop.PublicHeaderRoot != "" {
		values = append(values, path.Join(state.packageRoots[interop.Package], interop.PublicHeaderRoot))
	}
	for _, edge := range state.directTargetDependencies(interop) {
		other := state.targets[edge.index]
		if other.PublicHeaderRoot == "" {
			continue
		}
		values = append(values, path.Join(state.packageRoots[other.Package], other.PublicHeaderRoot))
	}
	sort.Strings(values)
	return values
}

func (state *closeState) resolveIncludeSpelling(bases []string, spelling string) (Resolution, bool) {
	for _, base := range bases {
		resolution, err := state.roots.resolveReference(base, spelling)
		if err == nil && resolution.Class == ResolvedAdmitted {
			return resolution, true
		}
	}
	return Resolution{}, false
}

func (state *closeState) systemHeaderDeclared(spelling string) bool {
	for _, candidate := range state.components {
		for _, root := range candidate.component.Roots {
			resolution, err := state.roots.resolveReference(root, spelling)
			if err == nil && resolution.Class == ResolvedBinding {
				return true
			}
		}
		if containsString(candidate.component.Modules, spelling) {
			return true
		}
	}
	return false
}

func (state *closeState) moduleDeclared(interop *TargetInterop, name string) bool {
	root := strings.Split(name, ".")[0]
	for _, edge := range state.directTargetDependencies(interop) {
		other := state.targets[edge.index]
		if other.ModuleMap == nil {
			continue
		}
		if containsString(other.ModuleMap.Parsed.Modules, root) || containsString(other.ModuleMap.Parsed.Modules, name) {
			return true
		}
	}
	if interop.ModuleMap != nil && (containsString(interop.ModuleMap.Parsed.Modules, root) || containsString(interop.ModuleMap.Parsed.Modules, name)) {
		return true
	}
	for _, candidate := range state.components {
		if containsString(candidate.component.Modules, root) || containsString(candidate.component.Frameworks, root) {
			return true
		}
	}
	return false
}

// directTargetDependencies resolves one target's declared target and product
// edges to indices into the classified target table, together with the
// selection-neutral condition each edge carries. Conditions are never filtered
// here: the declaration is a capture fact and the destination verdict belongs
// to the active projection.
func (state *closeState) directTargetDependencies(interop *TargetInterop) []targetEdge {
	pkg := state.packages[interop.Package]
	target, ok := findTarget(pkg.Manifest, interop.Target)
	if !ok {
		return nil
	}
	result := []targetEdge{}
	seen := map[int]bool{}
	for _, dependency := range target.Dependencies {
		identity := interop.Package
		if dependency.Package != "" {
			identity = strings.ToLower(dependency.Package)
		}
		names := []string{dependency.Name}
		if dependency.Product != "" {
			other, exists := state.packages[identity]
			if !exists {
				continue
			}
			product, exists := findProduct(other.Manifest, dependency.Product)
			if !exists {
				continue
			}
			names = product.Targets
		}
		for _, name := range names {
			index, exists := state.byTargetKey[identity+":"+name]
			if !exists || seen[index] {
				continue
			}
			seen[index] = true
			result = append(result, targetEdge{index: index, condition: dependency.Condition})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].index < result[j].index })
	return result
}

// targetTreeRoot is the admitted subtree one target owns. SwiftPM always
// declares a target path; the source projection is the fallback when it does
// not.
func targetTreeRoot(target swiftpmsource.Target, interop *TargetInterop) string {
	if cleaned := path.Clean(target.Path); cleaned != "" && cleaned != "." {
		return cleaned
	}
	return targetSourceRoot(interop)
}

func headerPaths(headers []HeaderFile) []string {
	values := []string{}
	for _, header := range headers {
		if isModuleMap(header.Relative) {
			continue
		}
		values = append(values, header.Relative)
	}
	return values
}

func isAbsoluteSpelling(value string) bool {
	return strings.HasPrefix(value, "/") || windowsAbsolute(value)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func findTarget(manifest swiftpmsource.Manifest, name string) (swiftpmsource.Target, bool) {
	for _, value := range manifest.Targets {
		if value.Name == name {
			return value, true
		}
	}
	return swiftpmsource.Target{}, false
}

func findProduct(manifest swiftpmsource.Manifest, name string) (swiftpmsource.Product, bool) {
	for _, value := range manifest.Products {
		if value.Name == name {
			return value, true
		}
	}
	return swiftpmsource.Product{}, false
}

var _ = closureexec.AssurancePortable

func readTrustedFile(absolute string) ([]byte, error) {
	return os.ReadFile(absolute) // #nosec G304 -- path already confined to a selected external root.
}

func readAdmittedFile(root, relative string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(relative))) // #nosec G304 -- admitted immutable file below a verified protected package root.
}

func digestBytes(payload []byte) closuregraph.ID {
	sum := sha256.Sum256(payload)
	return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
}
