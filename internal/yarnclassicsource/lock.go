package yarnclassicsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

// SupportedYarnVersion pins the exact Classic implementation whose lock and
// offline-mirror behavior is covered by this profile.
const SupportedYarnVersion = "1.22.22"

// Target is the exact Node platform selection used for optional dependency pruning.
type Target struct {
	OS, Architecture, Libc string
	IncludeDev             bool
}

// ParseRequest supplies the sole authoritative root lock, all workspace
// manifests, and root manager configuration. Subtree lockfiles are not input.
type ParseRequest struct {
	LockName               string
	LockBytes              []byte
	Manifests              map[string][]byte
	Configuration          map[string][]byte
	YarnVersion            string
	AllowedRegistryOrigins []string
	Target                 Target
}

// DependencyEdge retains descriptor, scope, and selected/pruned evidence.
type DependencyEdge struct {
	From, Name, Spec, Scope, To string
	Selected                    bool
	Reason                      string
}

// Package is either a root/workspace tree or one immutable yarn.lock entry.
type Package struct {
	Key, Name, Version, Resolved, Integrity, WorkspacePath string
	Selectors                                              []string
	Dependencies, DevDependencies, OptionalDependencies    map[string]string
	PeerDependencies                                       map[string]string
	PeerOptional                                           map[string]bool
	OS, CPU, Libc                                          []string
	Selected                                               bool
	PruneReason                                            string
	manifest                                               packageManifest
}

// Layout binds Yarn workspace hoisting/layout inputs without making the
// installed node_modules tree graph authority.
type Layout struct {
	ModulesFolder string
	Nohoist       []string
}

// Graph is the frozen Yarn Classic graph before raw artifact admission.
type Graph struct {
	LockName, LockDigest, RawLockSHA256, RootName, RootVersion string
	YarnVersion, ConfigurationDigest                           string
	Target                                                     Target
	Layout                                                     Layout
	Packages                                                   []Package
	Edges                                                      []DependencyEdge
	Workspaces                                                 []string
	packageIndex                                               map[string]int
	selectorIndex                                              map[string]string
	workspaceByName                                            map[string]string
	manifestBytes                                              map[string][]byte
	configurationBytes                                         map[string][]byte
}

type packageManifest struct {
	Name, Version                                                         string
	Dependencies, DevDependencies, OptionalDependencies, PeerDependencies map[string]string
	PeerDependenciesMeta                                                  map[string]struct {
		Optional bool `json:"optional"`
	}
	Scripts                                             map[string]string
	OS, CPU, Libc                                       []string
	BundleDependencies, BundledDependencies, Workspaces json.RawMessage
	Resolutions                                         map[string]string
	Gypfile                                             *bool
	Private                                             bool
}

type workspaceDeclaration struct {
	Packages []string `json:"packages"`
	Nohoist  []string `json:"nohoist"`
}

type lockEntry struct {
	Selectors                          []string
	Version, Resolved, Integrity       string
	Dependencies, OptionalDependencies map[string]string
}

// Parse accepts the closed Yarn lockfile v1 grammar and reconciles every root
// and workspace declaration against one deterministic descriptor graph.
func Parse(request ParseRequest) (Graph, error) {
	if request.LockName != "yarn.lock" {
		if request.LockName == "" || len(request.LockBytes) == 0 {
			return Graph{}, fail(CodeLockMissing, "an authoritative root yarn.lock is required", map[string]string{"expected": "yarn.lock"})
		}
		return Graph{}, fail(CodeLockFormatUnsupported, "unsupported Yarn lock name", map[string]string{"format": request.LockName})
	}
	if len(request.LockBytes) == 0 {
		return Graph{}, fail(CodeLockMissing, "the root yarn.lock is empty", nil)
	}
	if request.YarnVersion != SupportedYarnVersion {
		return Graph{}, fail(CodeLockFormatUnsupported, "unsupported Yarn Classic implementation", map[string]string{"expected": SupportedYarnVersion, "observed": request.YarnVersion})
	}
	if request.Target.OS == "" || request.Target.Architecture == "" {
		return Graph{}, fail(CodeGraphIncomplete, "Yarn target OS and architecture are required", nil)
	}
	manifestBytes, manifests, err := parseManifests(request.Manifests)
	if err != nil {
		return Graph{}, err
	}
	root, ok := manifests["package.json"]
	if !ok || root.Name == "" || root.Version == "" {
		return Graph{}, fail(CodeLockStale, "root package.json identity is missing", map[string]string{"path": "package.json"})
	}
	configBytes, layout, configDigest, err := parseConfiguration(request.Configuration)
	if err != nil {
		return Graph{}, err
	}
	workspacePaths, nohoist, err := reconcileWorkspaces(root.Workspaces, manifests)
	if err != nil {
		return Graph{}, err
	}
	layout.Nohoist = nohoist
	if len(workspacePaths) > 0 && !root.Private {
		return Graph{}, fail(CodeLockStale, "Yarn workspace root must be private", map[string]string{"path": "package.json"})
	}
	entries, err := parseLock(request.LockBytes, request.AllowedRegistryOrigins)
	if err != nil {
		return Graph{}, err
	}
	packages := make([]Package, 0, len(entries)+len(workspacePaths)+1)
	index := map[string]int{}
	workspaceByName := map[string]string{}
	addWorkspace := func(key, workspace, manifestPath string, manifest packageManifest) error {
		if _, exists := workspaceByName[manifest.Name]; exists {
			return fail(CodeGraphIncomplete, "workspace package name is not unique", map[string]string{"name": manifest.Name})
		}
		pkg := Package{Key: key, Name: manifest.Name, Version: manifest.Version, WorkspacePath: workspace,
			Dependencies: cloneMap(manifest.Dependencies), DevDependencies: cloneMap(manifest.DevDependencies), OptionalDependencies: cloneMap(manifest.OptionalDependencies),
			PeerDependencies: cloneMap(manifest.PeerDependencies), PeerOptional: peerOptional(manifest.PeerDependenciesMeta),
			OS: sortedStrings(manifest.OS), CPU: sortedStrings(manifest.CPU), Libc: sortedStrings(manifest.Libc), manifest: manifest}
		if rawBundlePresent(manifest.BundleDependencies) || rawBundlePresent(manifest.BundledDependencies) {
			return fail(CodeBundledDependencyUnsupported, "workspace declares bundled dependencies", map[string]string{"path": manifestPath})
		}
		index[key] = len(packages)
		workspaceByName[manifest.Name] = key
		packages = append(packages, pkg)
		return nil
	}
	if err = addWorkspace("workspace:.", ".", "package.json", root); err != nil {
		return Graph{}, err
	}
	for _, workspace := range workspacePaths {
		manifestPath := path.Join(workspace, "package.json")
		if err = addWorkspace("workspace:"+workspace, workspace, manifestPath, manifests[manifestPath]); err != nil {
			return Graph{}, err
		}
	}
	selectorIndex := map[string]string{}
	for _, entry := range entries {
		name, err := selectorName(entry.Selectors[0])
		if err != nil {
			return Graph{}, err
		}
		key := remoteKey(name, entry.Version, entry.Integrity)
		if _, exists := index[key]; exists {
			return Graph{}, fail(CodeGraphIncomplete, "two Yarn entries collapse to one immutable package key", map[string]string{"package": key})
		}
		pkg := Package{Key: key, Name: name, Version: entry.Version, Resolved: entry.Resolved, Integrity: entry.Integrity,
			Selectors: append([]string(nil), entry.Selectors...), Dependencies: cloneMap(entry.Dependencies), OptionalDependencies: cloneMap(entry.OptionalDependencies),
			DevDependencies: map[string]string{}, PeerDependencies: map[string]string{}, PeerOptional: map[string]bool{}}
		for _, selector := range entry.Selectors {
			if prior := selectorIndex[selector]; prior != "" {
				return Graph{}, fail(CodeLockFormatUnsupported, "Yarn descriptor appears in multiple lock entries", map[string]string{"selector": selector})
			}
			selectorIndex[selector] = key
		}
		index[key] = len(packages)
		packages = append(packages, pkg)
	}
	for pattern, spec := range root.Resolutions {
		if !validPackageName(pattern) || selectorIndex[pattern+"@"+spec] == "" {
			return Graph{}, fail(CodeLockStale, "root resolution is outside the exact-name grammar or absent from yarn.lock", map[string]string{"resolution": pattern})
		}
	}
	edges, err := buildEdges(packages, index, selectorIndex, workspaceByName, request.Target)
	if err != nil {
		return Graph{}, err
	}
	if err = markSelection(packages, edges, index, request.Target); err != nil {
		return Graph{}, err
	}
	rawSum := sha256.Sum256(request.LockBytes)
	semanticEntries := make([]any, len(entries))
	for i, entry := range entries {
		semanticEntries[i] = map[string]any{"selectors": stringsToAny(entry.Selectors), "version": entry.Version, "resolved": entry.Resolved, "integrity": entry.Integrity, "dependencies": stringMapAny(entry.Dependencies), "optional_dependencies": stringMapAny(entry.OptionalDependencies)}
	}
	canonical, err := protocoljson.MarshalCanonical(map[string]any{"configuration_digest": configDigest, "entries": semanticEntries, "layout": map[string]any{"modules_folder": layout.ModulesFolder, "nohoist": stringsToAny(layout.Nohoist)}, "manifests": bytesMapDigest(manifestBytes), "schema": "yarn-lock-v1", "yarn_version": request.YarnVersion})
	if err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "Yarn lock cannot be canonicalized: "+err.Error(), nil)
	}
	lockID, err := closuregraph.IDFromCanonical("yarn-classic-lock-v1", canonical)
	if err != nil {
		return Graph{}, err
	}
	return Graph{LockName: "yarn.lock", LockDigest: string(lockID), RawLockSHA256: "sha256:" + hex.EncodeToString(rawSum[:]), RootName: root.Name, RootVersion: root.Version,
		YarnVersion: request.YarnVersion, ConfigurationDigest: configDigest, Target: request.Target, Layout: layout, Packages: packages, Edges: edges,
		Workspaces: workspacePaths, packageIndex: index, selectorIndex: selectorIndex, workspaceByName: workspaceByName, manifestBytes: manifestBytes, configurationBytes: configBytes}, nil
}

func parseLock(payload []byte, allowed []string) ([]lockEntry, error) {
	if bytes.Contains(payload, []byte("\r")) || !bytes.HasSuffix(payload, []byte("\n")) {
		return nil, fail(CodeLockFormatUnsupported, "yarn.lock must use LF and a terminal newline", nil)
	}
	lines := strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
	entries := []lockEntry{}
	var current *lockEntry
	section := ""
	seenHeader := false
	for lineNo, line := range lines {
		if strings.TrimSpace(line) == "" || (strings.HasPrefix(line, "#") && !seenHeader) {
			continue
		}
		if strings.ContainsRune(line, '\t') || strings.HasSuffix(line, " ") {
			return nil, lockFormat(lineNo, "tabs and trailing whitespace are forbidden")
		}
		if !strings.HasPrefix(line, " ") {
			if !strings.HasSuffix(line, ":") {
				return nil, lockFormat(lineNo, "entry header must end with colon")
			}
			selectors, err := parseSelectorList(strings.TrimSuffix(line, ":"))
			if err != nil {
				return nil, lockFormat(lineNo, err.Error())
			}
			entries = append(entries, lockEntry{Selectors: selectors, Dependencies: map[string]string{}, OptionalDependencies: map[string]string{}})
			current = &entries[len(entries)-1]
			section, seenHeader = "", true
			continue
		}
		if current == nil {
			return nil, lockFormat(lineNo, "field precedes first entry")
		}
		if strings.HasPrefix(line, "    ") {
			if section == "" || strings.HasPrefix(line[4:], " ") {
				return nil, lockFormat(lineNo, "invalid dependency indentation")
			}
			name, value, err := parsePair(line[4:])
			if err != nil || !validPackageName(name) {
				return nil, lockFormat(lineNo, "invalid dependency declaration")
			}
			target := current.Dependencies
			if section == "optionalDependencies" {
				target = current.OptionalDependencies
			}
			if _, duplicate := target[name]; duplicate {
				return nil, lockFormat(lineNo, "duplicate dependency declaration")
			}
			target[name] = value
			continue
		}
		if !strings.HasPrefix(line, "  ") || strings.HasPrefix(line[2:], " ") {
			return nil, lockFormat(lineNo, "invalid field indentation")
		}
		field := line[2:]
		if strings.HasSuffix(field, ":") {
			section = strings.TrimSuffix(field, ":")
			if section != "dependencies" && section != "optionalDependencies" {
				return nil, lockFormat(lineNo, "unsupported nested field "+section)
			}
			continue
		}
		section = ""
		name, value, err := parsePair(field)
		if err != nil {
			return nil, lockFormat(lineNo, err.Error())
		}
		switch name {
		case "version":
			if current.Version != "" {
				return nil, lockFormat(lineNo, "duplicate version")
			}
			current.Version = value
		case "resolved":
			if current.Resolved != "" {
				return nil, lockFormat(lineNo, "duplicate resolved")
			}
			current.Resolved = value
		case "integrity":
			if current.Integrity != "" {
				return nil, lockFormat(lineNo, "duplicate integrity")
			}
			current.Integrity = value
		default:
			return nil, lockFormat(lineNo, "unsupported field "+name)
		}
	}
	if len(entries) == 0 {
		return nil, fail(CodeGraphIncomplete, "yarn.lock has no package entries", nil)
	}
	seenSelectors := map[string]bool{}
	for i := range entries {
		entry := &entries[i]
		if !exactVersion(entry.Version) {
			return nil, fail(CodeLockStale, "Yarn entry version is not exact", map[string]string{"selector": entry.Selectors[0]})
		}
		if err := validateRemoteIdentity(entry.Resolved, entry.Integrity, allowed); err != nil {
			return nil, err
		}
		firstName := mustSelectorName(entry.Selectors[0])
		for _, selector := range entry.Selectors {
			if seenSelectors[selector] {
				return nil, fail(CodeLockFormatUnsupported, "duplicate Yarn descriptor", map[string]string{"selector": selector})
			}
			seenSelectors[selector] = true
			name, err := selectorName(selector)
			if err != nil {
				return nil, err
			}
			if name != firstName {
				return nil, fail(CodeLockFormatUnsupported, "one Yarn entry mixes package names", nil)
			}
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.Join(entries[i].Selectors, "\x00") < strings.Join(entries[j].Selectors, "\x00")
	})
	return entries, nil
}

func parseSelectorList(value string) ([]string, error) {
	parts := splitComma(value)
	selectors := make([]string, 0, len(parts))
	for _, part := range parts {
		item, err := parseScalar(strings.TrimSpace(part))
		if err != nil || item == "" {
			return nil, fmt.Errorf("invalid selector")
		}
		if _, err = selectorName(item); err != nil {
			return nil, err
		}
		selectors = append(selectors, item)
	}
	if len(selectors) == 0 {
		return nil, fmt.Errorf("empty selector list")
	}
	sort.Strings(selectors)
	for i := 1; i < len(selectors); i++ {
		if selectors[i] == selectors[i-1] {
			return nil, fmt.Errorf("duplicate selector")
		}
	}
	return selectors, nil
}

func splitComma(value string) []string {
	parts, start, quoted, escaped := []string{}, 0, false, false
	for i, r := range value {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' && quoted {
			escaped = true
			continue
		}
		if r == '"' {
			quoted = !quoted
			continue
		}
		if r == ',' && !quoted {
			parts = append(parts, value[start:i])
			start = i + 1
		}
	}
	return append(parts, value[start:])
}

func parsePair(value string) (string, string, error) {
	index := strings.IndexByte(value, ' ')
	if index <= 0 {
		return "", "", fmt.Errorf("field lacks a value")
	}
	name := value[:index]
	scalar, err := parseScalar(strings.TrimSpace(value[index+1:]))
	if err != nil || scalar == "" {
		return "", "", fmt.Errorf("field has an invalid value")
	}
	return name, scalar, nil
}

func parseScalar(value string) (string, error) {
	if strings.HasPrefix(value, "\"") {
		if !strings.HasSuffix(value, "\"") {
			return "", fmt.Errorf("unterminated quoted value")
		}
		result, err := strconv.Unquote(value)
		if err != nil || strings.ContainsAny(result, "\x00\r\n") {
			return "", fmt.Errorf("invalid quoted value")
		}
		return result, nil
	}
	if value == "" || strings.ContainsAny(value, " \"\x00\r\n") {
		return "", fmt.Errorf("invalid bare value")
	}
	return value, nil
}

func selectorName(selector string) (string, error) {
	index := strings.LastIndex(selector, "@")
	if strings.HasPrefix(selector, "@") {
		index = strings.Index(selector[1:], "@") + 1
	}
	if index <= 0 || index == len(selector)-1 || !validPackageName(selector[:index]) {
		return "", fail(CodeLockFormatUnsupported, "invalid Yarn descriptor", map[string]string{"selector": selector})
	}
	return selector[:index], nil
}
func mustSelectorName(selector string) string { value, _ := selectorName(selector); return value }

func parseManifests(values map[string][]byte) (map[string][]byte, map[string]packageManifest, error) {
	result, decoded := map[string][]byte{}, map[string]packageManifest{}
	for name, payload := range values {
		if !validProjectFile(name, "package.json") {
			return nil, nil, fail(CodeLocalPathEscape, "manifest path escapes the project", map[string]string{"path": name})
		}
		if err := validateJSON(payload); err != nil {
			return nil, nil, fail(CodeLockStale, "invalid package manifest: "+err.Error(), map[string]string{"path": name})
		}
		var manifest packageManifest
		if err := json.Unmarshal(payload, &manifest); err != nil {
			return nil, nil, fail(CodeLockStale, "cannot decode package manifest", map[string]string{"path": name})
		}
		result[name], decoded[name] = append([]byte(nil), payload...), manifest
	}
	return result, decoded, nil
}

func reconcileWorkspaces(raw json.RawMessage, manifests map[string]packageManifest) ([]string, []string, error) {
	patterns, nohoist, err := workspacePatterns(raw)
	if err != nil {
		return nil, nil, err
	}
	dirs := []string{}
	for name := range manifests {
		if name != "package.json" {
			dirs = append(dirs, path.Dir(name))
		}
	}
	sort.Strings(dirs)
	selected := []string{}
	for _, dir := range dirs {
		if matchesAnyWorkspace(patterns, dir) {
			selected = append(selected, dir)
		}
	}
	for _, pattern := range patterns {
		found := false
		for _, dir := range selected {
			found = found || matchWorkspace(pattern, dir)
		}
		if !found {
			return nil, nil, fail(CodeLockStale, "workspace pattern has no captured manifest", map[string]string{"workspace": pattern})
		}
	}
	if len(selected) != len(dirs) {
		return nil, nil, fail(CodeLockStale, "captured manifest is outside root workspace declaration", nil)
	}
	return selected, nohoist, nil
}

func workspacePatterns(raw json.RawMessage) ([]string, []string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return []string{}, []string{}, nil
	}
	var direct []string
	if json.Unmarshal(raw, &direct) == nil {
		values, err := validateWorkspacePatterns(direct)
		return values, []string{}, err
	}
	var value workspaceDeclaration
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, nil, fail(CodeLockFormatUnsupported, "unsupported Yarn workspaces declaration", nil)
	}
	patterns, err := validateWorkspacePatterns(value.Packages)
	if err != nil {
		return nil, nil, err
	}
	nohoist := sortedStrings(value.Nohoist)
	for i, item := range nohoist {
		if item == "" || strings.ContainsAny(item, "\\\x00\r\n") || path.IsAbs(item) || (i > 0 && nohoist[i-1] == item) {
			return nil, nil, fail(CodeLockFormatUnsupported, "invalid Yarn nohoist layout rule", map[string]string{"nohoist": item})
		}
	}
	return patterns, nohoist, nil
}

func validateWorkspacePatterns(values []string) ([]string, error) {
	result := sortedStrings(values)
	for i, value := range result {
		if value == "" || path.IsAbs(value) || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || strings.Contains(value, "\\") || strings.ContainsAny(value, "?[") || (strings.Contains(value, "*") && !strings.HasSuffix(value, "/*")) || (i > 0 && result[i-1] == value) {
			return nil, fail(CodeLockFormatUnsupported, "workspace pattern is outside the closed exact/trailing-star grammar", map[string]string{"workspace": value})
		}
	}
	return result, nil
}

func parseConfiguration(values map[string][]byte) (map[string][]byte, Layout, string, error) {
	result := map[string][]byte{}
	layout := Layout{ModulesFolder: "node_modules"}
	canonical := map[string]any{}
	for name, payload := range values {
		if name != ".yarnrc" {
			return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "unsupported Yarn configuration file", map[string]string{"path": name})
		}
		if bytes.Contains(payload, []byte("\r")) {
			return nil, Layout{}, "", fail(CodeLockFormatUnsupported, "Yarn configuration must use LF", nil)
		}
		settings := map[string]string{}
		for lineNo, line := range strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, err := parsePair(line)
			if err != nil {
				return nil, Layout{}, "", lockFormat(lineNo, "invalid .yarnrc setting")
			}
			switch key {
			case "--install.modules-folder", "modules-folder":
				if value == "" || path.IsAbs(value) || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") {
					return nil, Layout{}, "", fail(CodeLocalPathEscape, "Yarn modules folder escapes the project", map[string]string{"path": value})
				}
				layout.ModulesFolder = value
			case "workspaces-experimental":
				if value != "true" {
					return nil, Layout{}, "", fail(CodeLockFormatUnsupported, "workspaces-experimental must remain enabled", nil)
				}
			default:
				return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "Yarn configuration extends resolution or execution", map[string]string{"setting": key})
			}
			if _, duplicate := settings[key]; duplicate {
				return nil, Layout{}, "", fail(CodeLockFormatUnsupported, "duplicate Yarn configuration setting", map[string]string{"setting": key})
			}
			settings[key] = value
		}
		result[name] = append([]byte(nil), payload...)
		canonical[name] = stringMapAny(settings)
	}
	id, err := closuregraph.DomainID("yarn-classic-configuration-v1", canonical)
	if err != nil {
		return nil, Layout{}, "", err
	}
	return result, layout, string(id), nil
}

func buildEdges(packages []Package, index map[string]int, selectors, workspaces map[string]string, _ Target) ([]DependencyEdge, error) {
	edges := []DependencyEdge{}
	for _, pkg := range packages {
		sets := []struct {
			scope  string
			values map[string]string
		}{{"runtime", pkg.Dependencies}, {"development", pkg.DevDependencies}, {"optional", pkg.OptionalDependencies}, {"peer", pkg.PeerDependencies}}
		for _, set := range sets {
			for _, name := range sortedMapKeys(set.values) {
				if set.scope == "runtime" {
					if _, optional := pkg.OptionalDependencies[name]; optional {
						continue
					}
				}
				spec := set.values[name]
				to := resolveDependency(name, spec, selectors, workspaces, packages, index)
				if to == "" && set.scope == "peer" {
					to = resolveUniquePeer(name, spec, packages)
				}
				optional := set.scope == "optional" || (set.scope == "peer" && pkg.PeerOptional[name])
				if to == "" && !optional {
					code := CodeGraphIncomplete
					if pkg.WorkspacePath != "" {
						code = CodeLockStale
					}
					return nil, fail(code, "dependency has no Yarn lock/workspace instance", map[string]string{"from": pkg.Key, "dependency": name, "specifier": spec})
				}
				reason := ""
				if to == "" {
					reason = "optional_missing"
				}
				edges = append(edges, DependencyEdge{From: pkg.Key, Name: name, Spec: spec, Scope: set.scope, To: to, Selected: to != "", Reason: reason})
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool { return edgeKey(edges[i]) < edgeKey(edges[j]) })
	return edges, nil
}

func resolveUniquePeer(name, spec string, packages []Package) string {
	matched := ""
	for _, pkg := range packages {
		if pkg.Name != name || !versionSatisfies(pkg.Version, spec) {
			continue
		}
		if matched != "" && matched != pkg.Key {
			return ""
		}
		matched = pkg.Key
	}
	return matched
}

// versionSatisfies is intentionally a small closed grammar for peer
// reconciliation. Complex ranges fail closed instead of delegating authority
// to a live resolver.
func versionSatisfies(version, spec string) bool {
	if spec == "*" || spec == version {
		return true
	}
	prefix := ""
	for _, candidate := range []string{"^", "~", ">=", "<=", ">", "<"} {
		if strings.HasPrefix(spec, candidate) {
			prefix = candidate
			break
		}
	}
	if prefix == "" || strings.ContainsAny(strings.TrimPrefix(spec, prefix), " <>=~^*|") {
		return false
	}
	left, right := semverParts(version), semverParts(strings.TrimPrefix(spec, prefix))
	if left == nil || right == nil {
		return false
	}
	compare := compareVersion(left, right)
	switch prefix {
	case "^":
		if compare < 0 {
			return false
		}
		switch {
		case right[0] != 0:
			return left[0] == right[0]
		case right[1] != 0:
			return left[0] == 0 && left[1] == right[1]
		default:
			return left[0] == 0 && left[1] == 0 && left[2] == right[2]
		}
	case "~":
		return compare >= 0 && left[0] == right[0] && left[1] == right[1]
	case ">=":
		return compare >= 0
	case "<=":
		return compare <= 0
	case ">":
		return compare > 0
	case "<":
		return compare < 0
	}
	return false
}

func semverParts(value string) []int {
	// Ranges deliberately accept only complete release versions. Exact string
	// equality above remains valid for prereleases, but range/prerelease
	// precedence is outside this profile's closed peer grammar.
	if strings.ContainsAny(value, "-+") {
		return nil
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil
	}
	result := make([]int, 3)
	for i, part := range parts {
		if len(part) > 1 && part[0] == '0' {
			return nil
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return nil
		}
		result[i] = number
	}
	return result
}
func compareVersion(left, right []int) int {
	for i := range left {
		if left[i] < right[i] {
			return -1
		}
		if left[i] > right[i] {
			return 1
		}
	}
	return 0
}

func resolveDependency(name, spec string, selectors, workspaces map[string]string, packages []Package, index map[string]int) string {
	if key := workspaces[name]; key != "" && workspaceSpecMatches(spec, packages[index[key]].Version, packages[index[key]].WorkspacePath) {
		return key
	}
	return selectors[name+"@"+spec]
}
func workspaceSpecMatches(spec, version, workspace string) bool {
	switch spec {
	case "*", version, "workspace:*", "workspace:^", "workspace:~", "link:" + workspace:
		return true
	}
	return false
}

func markSelection(packages []Package, edges []DependencyEdge, index map[string]int, target Target) error {
	for i := range packages {
		packages[i].Selected, packages[i].PruneReason = false, "unreachable"
	}
	queue := []string{"workspace:."}
	packages[index["workspace:."]].Selected, packages[index["workspace:."]].PruneReason = true, ""
	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for i := range edges {
			edge := &edges[i]
			if edge.From != from || edge.To == "" {
				continue
			}
			if edge.Scope == "development" && !target.IncludeDev {
				edge.Selected, edge.Reason = false, "development_omitted"
				continue
			}
			targetPkg := &packages[index[edge.To]]
			if ok, reason := matchesTarget(*targetPkg, target); !ok {
				if edge.Scope != "optional" && (edge.Scope != "peer" || !packages[index[from]].PeerOptional[edge.Name]) {
					return fail(CodeGraphIncomplete, "required Yarn dependency is incompatible with the selected target", map[string]string{"from": from, "package": edge.To, "reason": reason})
				}
				edge.Selected, edge.Reason = false, reason
				targetPkg.PruneReason = reason
				continue
			}
			edge.Selected, edge.Reason = true, ""
			if !targetPkg.Selected {
				targetPkg.Selected, targetPkg.PruneReason = true, ""
				queue = append(queue, edge.To)
			}
		}
	}
	return nil
}

func matchesTarget(pkg Package, target Target) (bool, string) {
	for _, item := range []struct {
		name, value string
		rules       []string
	}{{"os", target.OS, pkg.OS}, {"cpu", target.Architecture, pkg.CPU}, {"libc", target.Libc, pkg.Libc}} {
		if len(item.rules) > 0 && !selectorMatches(item.rules, item.value) {
			return false, item.name + "_pruned"
		}
	}
	return true, ""
}
func selectorMatches(rules []string, value string) bool {
	positive := false
	for _, rule := range rules {
		if strings.HasPrefix(rule, "!") {
			if strings.TrimPrefix(rule, "!") == value {
				return false
			}
		} else {
			positive = true
			if rule == value {
				return true
			}
		}
	}
	return !positive
}

func validateRemoteIdentity(resolved, integrity string, allowed []string) error {
	if !strings.HasPrefix(integrity, "sha512-") || strings.ContainsAny(integrity, " \t") {
		return fail(CodeIntegrityMissing, "Yarn entry lacks one canonical sha512 integrity", map[string]string{"resolved": resolved})
	}
	parsed, err := url.Parse(resolved)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || !strings.HasSuffix(strings.ToLower(parsed.Path), ".tgz") {
		return fail(CodeOriginUnpinned, "Yarn locator is not a supported immutable registry tarball", map[string]string{"scheme": locatorScheme(resolved)})
	}
	if !registryOriginAllowed(parsed, allowed) {
		return fail(CodeOriginUnpinned, "Yarn registry origin is outside the allowlist", map[string]string{"origin": parsed.Scheme + "://" + parsed.Host})
	}
	return nil
}
func registryOriginAllowed(resolved *url.URL, allowed []string) bool {
	for _, value := range allowed {
		candidate, err := url.Parse(value)
		if err == nil && candidate.Scheme == resolved.Scheme && strings.EqualFold(candidate.Host, resolved.Host) && (candidate.Path == "" || candidate.Path == "/" || strings.HasPrefix(resolved.Path, strings.TrimSuffix(candidate.Path, "/")+"/")) {
			return true
		}
	}
	return false
}

func validProjectFile(value, base string) bool {
	return value != "" && !path.IsAbs(value) && path.Clean(value) == value && value != ".." && !strings.HasPrefix(value, "../") && path.Base(value) == base
}
func validPackageName(value string) bool {
	if value == "" || strings.ContainsAny(value, "\\\x00\r\n ") {
		return false
	}
	if strings.HasPrefix(value, "@") {
		return strings.Count(value, "/") == 1 && !strings.Contains(strings.TrimPrefix(value, "@"), "@")
	}
	return !strings.ContainsAny(value, "@/")
}
func exactVersion(value string) bool {
	return value != "" && !strings.ContainsAny(value, " <>~=^*|/\\\x00\r\n")
}
func remoteKey(name, version, integrity string) string {
	sum := sha256.Sum256([]byte(name + "\x00" + version + "\x00" + integrity))
	return "remote:" + name + "@" + version + ":" + hex.EncodeToString(sum[:8])
}
func locatorScheme(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return "invalid"
	}
	return parsed.Scheme
}
func rawBundlePresent(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) != 0 && !bytes.Equal(trimmed, []byte("null")) && !bytes.Equal(trimmed, []byte("[]")) && !bytes.Equal(trimmed, []byte("false"))
}
func matchWorkspace(pattern, dir string) bool {
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(dir, prefix) && !strings.Contains(strings.TrimPrefix(dir, prefix), "/")
	}
	return pattern == dir
}
func matchesAnyWorkspace(patterns []string, dir string) bool {
	for _, pattern := range patterns {
		if matchWorkspace(pattern, dir) {
			return true
		}
	}
	return false
}
func edgeKey(value DependencyEdge) string {
	return strings.Join([]string{value.From, value.Scope, value.Name, value.Spec, value.To}, "\x00")
}
func lockFormat(line int, detail string) error {
	return fail(CodeLockFormatUnsupported, detail, map[string]string{"line": fmt.Sprint(line + 1)})
}
func cloneMap(value map[string]string) map[string]string {
	result := map[string]string{}
	for key, item := range value {
		result[key] = item
	}
	return result
}
func sortedStrings(value []string) []string {
	result := append([]string(nil), value...)
	sort.Strings(result)
	return result
}
func sortedMapKeys(value map[string]string) []string {
	result := make([]string, 0, len(value))
	for key := range value {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func peerOptional(value map[string]struct {
	Optional bool `json:"optional"`
}) map[string]bool {
	result := map[string]bool{}
	for key, item := range value {
		if item.Optional {
			result[key] = true
		}
	}
	return result
}
func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}
func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}
func stringMapAny(values map[string]string) map[string]any {
	result := map[string]any{}
	for key, value := range values {
		result[key] = value
	}
	return result
}
func bytesMapDigest(values map[string][]byte) map[string]any {
	result := map[string]any{}
	for key, value := range values {
		sum := sha256.Sum256(value)
		result[key] = "sha256:" + hex.EncodeToString(sum[:])
	}
	return result
}

func validateJSON(payload []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if token, err := decoder.Token(); err == nil || token != nil {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}
func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok || seen[key] {
				return fmt.Errorf("duplicate or invalid object key")
			}
			seen[key] = true
			if err = scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return fmt.Errorf("invalid object end")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return fmt.Errorf("invalid array end")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
	return nil
}

func targetCondition(pkg Package) *closuregraph.Condition {
	clauses := []string{}
	for _, item := range []struct {
		name  string
		rules []string
	}{{"os", pkg.OS}, {"architecture", pkg.CPU}, {"libc", pkg.Libc}} {
		if len(item.rules) > 0 {
			clauses = append(clauses, item.name+"="+strings.Join(item.rules, ","))
		}
	}
	if len(clauses) == 0 {
		return nil
	}
	return &closuregraph.Condition{EvaluatorID: "yarn-classic-platform-v1", Expression: strings.Join(clauses, ";")}
}
