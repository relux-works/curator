package npmsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Target is the exact npm platform selection used to classify optional edges.
type Target struct {
	OS, Architecture, Libc string
	IncludeDev             bool
}

// ParseRequest supplies the authoritative root lock and every root/workspace manifest.
// Manifest keys are project-relative package.json paths using forward slashes.
type ParseRequest struct {
	LockName               string
	LockBytes              []byte
	Manifests              map[string][]byte
	AllowedRegistryOrigins []string
	Target                 Target
}

// DependencyEdge retains selected and pruned dependency declarations.
type DependencyEdge struct {
	From, Name, Spec, Scope, To string
	Selected                    bool
	Reason                      string
}

// Package is one package-lock install-tree instance. InstallPath is its
// package-lock packages-table key; peer/layout contexts therefore remain distinct.
type Package struct {
	InstallPath, Key, Name, Version, Resolved, Integrity, WorkspacePath   string
	Dependencies, DevDependencies, OptionalDependencies, PeerDependencies map[string]string
	PeerOptional                                                          map[string]bool
	OS, CPU, Libc                                                         []string
	Dev, Optional, Link, InBundle, HasInstallScript                       bool
	Selected                                                              bool
	PruneReason                                                           string
	manifest                                                              packageManifest
}

// Graph is the frozen npm lock/workspace graph before raw artifact admission.
type Graph struct {
	LockName, LockDigest, RawLockSHA256, RootName, RootVersion string
	LockfileVersion                                            int
	Target                                                     Target
	Packages                                                   []Package
	Edges                                                      []DependencyEdge
	Workspaces                                                 []string
	packageIndex                                               map[string]int
	linkTargets                                                map[string]string
	manifestBytes                                              map[string][]byte
}

type lockWire struct {
	Name            string                     `json:"name"`
	Version         string                     `json:"version"`
	LockfileVersion int                        `json:"lockfileVersion"`
	Packages        map[string]json.RawMessage `json:"packages"`
	Dependencies    json.RawMessage            `json:"dependencies"`
}

type packageWire struct {
	Name, Version, Resolved, Integrity                                    string
	Dependencies, DevDependencies, OptionalDependencies, PeerDependencies map[string]string
	PeerDependenciesMeta                                                  map[string]struct {
		Optional bool `json:"optional"`
	}
	OS, CPU, Libc                                                []string
	Dev, Optional, DevOptional, Link, InBundle, HasInstallScript bool
	BundleDependencies, BundledDependencies, Workspaces          json.RawMessage
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
	Gypfile                                             *bool
}

type legacyDependency struct {
	Version, Resolved, Integrity string
	Dev, Optional                bool
	Requires                     map[string]string
	Dependencies                 map[string]legacyDependency
}

// Parse validates supported npm lock schemas (v2 and v3), reconciles root and
// workspace manifests, and records every selected or pruned dependency edge.
func Parse(request ParseRequest) (Graph, error) {
	if request.LockName != "package-lock.json" && request.LockName != "npm-shrinkwrap.json" {
		if request.LockName == "" || len(request.LockBytes) == 0 {
			return Graph{}, fail(CodeLockMissing, "an authoritative npm lock is required", map[string]string{"expected": "package-lock.json|npm-shrinkwrap.json"})
		}
		return Graph{}, fail(CodeLockFormatUnsupported, "unsupported npm lock name", map[string]string{"format": request.LockName})
	}
	if len(request.LockBytes) == 0 {
		return Graph{}, fail(CodeLockMissing, "the npm lock is empty", map[string]string{"lock": request.LockName})
	}
	if err := validateJSON(request.LockBytes); err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "npm lock is not closed JSON: "+err.Error(), map[string]string{"lock": request.LockName})
	}
	var lock lockWire
	if err := json.Unmarshal(request.LockBytes, &lock); err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "cannot decode npm lock: "+err.Error(), nil)
	}
	if lock.LockfileVersion != 2 && lock.LockfileVersion != 3 {
		return Graph{}, fail(CodeLockFormatUnsupported, "only package-lock v2 and v3 are supported", map[string]string{"lockfile_version": fmt.Sprint(lock.LockfileVersion)})
	}
	if len(lock.Packages) == 0 {
		return Graph{}, fail(CodeGraphIncomplete, "npm lock packages table is absent", nil)
	}
	manifestBytes, manifests, err := parseManifests(request.Manifests)
	if err != nil {
		return Graph{}, err
	}
	rootManifest, ok := manifests["package.json"]
	if !ok {
		return Graph{}, fail(CodeLockStale, "root package.json is missing", map[string]string{"path": "package.json"})
	}
	if request.Target.OS == "" || request.Target.Architecture == "" {
		return Graph{}, fail(CodeGraphIncomplete, "npm target OS and architecture are required", nil)
	}
	keys := make([]string, 0, len(lock.Packages))
	for key := range lock.Packages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	packages := make([]Package, 0, len(keys))
	index := make(map[string]int, len(keys))
	links := map[string]string{}
	for _, installPath := range keys {
		if err := validateInstallPath(installPath); err != nil {
			return Graph{}, err
		}
		var wire packageWire
		if err := json.Unmarshal(lock.Packages[installPath], &wire); err != nil {
			return Graph{}, fail(CodeLockFormatUnsupported, fmt.Sprintf("invalid package entry: %v", err), map[string]string{"field": "packages." + installPath})
		}
		name := wire.Name
		if name == "" {
			name = packageNameFromInstallPath(installPath)
		}
		if installPath == "" {
			name = firstNonempty(wire.Name, lock.Name, rootManifest.Name)
		}
		pkg := Package{InstallPath: installPath, Key: packageKey(installPath), Name: name, Version: wire.Version,
			Resolved: wire.Resolved, Integrity: wire.Integrity, Dependencies: cloneMap(wire.Dependencies),
			DevDependencies: cloneMap(wire.DevDependencies), OptionalDependencies: cloneMap(wire.OptionalDependencies),
			PeerDependencies: cloneMap(wire.PeerDependencies), PeerOptional: peerOptional(wire.PeerDependenciesMeta),
			OS: sortedStrings(wire.OS), CPU: sortedStrings(wire.CPU), Libc: sortedStrings(wire.Libc),
			Dev: wire.Dev, Optional: wire.Optional || wire.DevOptional, Link: wire.Link, InBundle: wire.InBundle,
			HasInstallScript: wire.HasInstallScript}
		if installPath == "" || (!strings.Contains(installPath, "node_modules/") && !strings.HasPrefix(installPath, "node_modules/")) {
			pkg.WorkspacePath = installPath
		}
		if wire.Link {
			if installPath == "" || wire.Resolved == "" || path.IsAbs(wire.Resolved) || path.Clean(wire.Resolved) != wire.Resolved || strings.HasPrefix(wire.Resolved, "../") {
				return Graph{}, fail(CodeLocalPathEscape, "npm workspace link escapes the project", map[string]string{"path": installPath, "resolved": wire.Resolved})
			}
			links[installPath] = wire.Resolved
		} else if isExternalInstallPath(installPath) {
			if wire.InBundle || rawBundlePresent(wire.BundleDependencies) || rawBundlePresent(wire.BundledDependencies) {
				return Graph{}, fail(CodeBundledDependencyUnsupported, "bundled npm dependency is unsupported", map[string]string{"package": installPath})
			}
			if err := validateRemoteIdentity(pkg, request.AllowedRegistryOrigins); err != nil {
				return Graph{}, err
			}
		}
		index[installPath] = len(packages)
		packages = append(packages, pkg)
	}
	rootIndex, ok := index[""]
	if !ok {
		return Graph{}, fail(CodeGraphIncomplete, "npm lock has no root packages entry", nil)
	}
	packages[rootIndex].Version = firstNonempty(packages[rootIndex].Version, lock.Version, rootManifest.Version)
	if err := reconcileManifest(&packages[rootIndex], rootManifest, "package.json"); err != nil {
		return Graph{}, err
	}
	workspacePaths, err := reconcileWorkspaces(rootManifest.Workspaces, manifests, packages, index, links)
	if err != nil {
		return Graph{}, err
	}
	for _, workspace := range workspacePaths {
		manifestPath := path.Join(workspace, "package.json")
		manifest := manifests[manifestPath]
		idx := index[workspace]
		if err := reconcileManifest(&packages[idx], manifest, manifestPath); err != nil {
			return Graph{}, err
		}
	}
	if lock.LockfileVersion == 2 {
		if err := reconcileV2DependencyTree(lock.Dependencies, packages, index, links); err != nil {
			return Graph{}, err
		}
	}
	edges, err := buildEdges(packages, index, links)
	if err != nil {
		return Graph{}, err
	}
	if err := markSelection(packages, edges, index, links, request.Target); err != nil {
		return Graph{}, err
	}
	for i := range edges {
		if edges[i].To != "" {
			target := packages[index[edges[i].To]]
			edges[i].Selected = target.Selected
			edges[i].Reason = target.PruneReason
		}
	}
	rawDigest := sha256.Sum256(request.LockBytes)
	var semantic any
	semanticDecoder := json.NewDecoder(bytes.NewReader(request.LockBytes))
	semanticDecoder.UseNumber()
	if err := semanticDecoder.Decode(&semantic); err != nil {
		return Graph{}, err
	}
	canonical, err := protocoljson.MarshalCanonical(semantic)
	if err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "npm lock cannot be canonicalized: "+err.Error(), nil)
	}
	lockID, err := closuregraph.IDFromCanonical("npm-lock-v1", canonical)
	if err != nil {
		return Graph{}, err
	}
	return Graph{LockName: request.LockName, LockDigest: string(lockID), RawLockSHA256: "sha256:" + hex.EncodeToString(rawDigest[:]), RootName: packages[rootIndex].Name,
		RootVersion: packages[rootIndex].Version, LockfileVersion: lock.LockfileVersion, Target: request.Target,
		Packages: packages, Edges: edges, Workspaces: workspacePaths, packageIndex: index, linkTargets: links, manifestBytes: manifestBytes}, nil
}

func reconcileV2DependencyTree(raw json.RawMessage, packages []Package, index map[string]int, links map[string]string) error {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fail(CodeLockStale, "package-lock v2 requires its legacy dependency tree", map[string]string{"field": "dependencies"})
	}
	var roots map[string]legacyDependency
	if err := json.Unmarshal(raw, &roots); err != nil {
		return fail(CodeLockFormatUnsupported, "package-lock v2 dependency tree is invalid", map[string]string{"field": "dependencies"})
	}
	seen := map[string]bool{}
	var walk func(string, map[string]legacyDependency) error
	walk = func(parent string, values map[string]legacyDependency) error {
		names := make([]string, 0, len(values))
		for name := range values {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			legacy := values[name]
			installPath := resolveDependency(parent, name, index, links)
			if installPath == "" {
				if legacy.Optional {
					continue
				}
				return fail(CodeGraphIncomplete, "package-lock v2 dependency tree references a missing package", map[string]string{"parent": parent, "dependency": name})
			}
			pkg := packages[index[installPath]]
			if pkg.Link {
				installPath = links[installPath]
				pkg = packages[index[installPath]]
			}
			if legacy.Version != pkg.Version || legacy.Resolved != pkg.Resolved || legacy.Integrity != pkg.Integrity || legacy.Dev != pkg.Dev || legacy.Optional != pkg.Optional {
				return fail(CodeLockStale, "package-lock v2 dependency tree differs from packages table", map[string]string{"package": installPath})
			}
			declared := cloneMap(pkg.Dependencies)
			for key, value := range pkg.OptionalDependencies {
				declared[key] = value
			}
			for key, value := range legacy.Requires {
				if declared[key] != value {
					return fail(CodeLockStale, "package-lock v2 requires map differs from packages table", map[string]string{"package": installPath, "dependency": key})
				}
			}
			seen[installPath] = true
			if err := walk(installPath, legacy.Dependencies); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk("", roots); err != nil {
		return err
	}
	for _, pkg := range packages {
		if isExternalInstallPath(pkg.InstallPath) && !pkg.Link && !seen[pkg.InstallPath] {
			return fail(CodeLockStale, "package-lock v2 packages table member is absent from dependency tree", map[string]string{"package": pkg.InstallPath})
		}
	}
	return nil
}

func parseManifests(values map[string][]byte) (map[string][]byte, map[string]packageManifest, error) {
	bytesByPath := map[string][]byte{}
	result := map[string]packageManifest{}
	for name, payload := range values {
		if name == "" || path.IsAbs(name) || path.Clean(name) != name || strings.HasPrefix(name, "../") || !strings.HasSuffix(name, "package.json") {
			return nil, nil, fail(CodeLocalPathEscape, "manifest path escapes the project", map[string]string{"path": name})
		}
		if err := validateJSON(payload); err != nil {
			return nil, nil, fail(CodeLockStale, fmt.Sprintf("invalid package manifest: %v", err), map[string]string{"field": name})
		}
		var manifest packageManifest
		if err := json.Unmarshal(payload, &manifest); err != nil {
			return nil, nil, fail(CodeLockStale, fmt.Sprintf("invalid package manifest: %v", err), map[string]string{"field": name})
		}
		bytesByPath[name] = append([]byte(nil), payload...)
		result[name] = manifest
	}
	return bytesByPath, result, nil
}

func reconcileManifest(pkg *Package, manifest packageManifest, manifestPath string) error {
	if pkg.Name == "" || manifest.Name == "" || pkg.Name != manifest.Name || pkg.Version == "" || manifest.Version == "" || pkg.Version != manifest.Version {
		return fail(CodeLockStale, "manifest package identity differs from npm lock", map[string]string{"path": manifestPath, "lock_name": pkg.Name, "manifest_name": manifest.Name, "lock_version": pkg.Version, "manifest_version": manifest.Version})
	}
	for _, item := range []struct {
		name           string
		lock, manifest map[string]string
	}{
		{"dependencies", pkg.Dependencies, manifest.Dependencies}, {"devDependencies", pkg.DevDependencies, manifest.DevDependencies},
		{"optionalDependencies", pkg.OptionalDependencies, manifest.OptionalDependencies}, {"peerDependencies", pkg.PeerDependencies, manifest.PeerDependencies},
	} {
		if !equalStringMap(item.lock, item.manifest) {
			return fail(CodeLockStale, "manifest dependency declaration differs from npm lock", map[string]string{"path": manifestPath, "field": item.name})
		}
	}
	pkg.manifest = manifest
	return nil
}

func reconcileWorkspaces(raw json.RawMessage, manifests map[string]packageManifest, packages []Package, index map[string]int, links map[string]string) ([]string, error) {
	patterns, err := workspacePatterns(raw)
	if err != nil {
		return nil, err
	}
	manifestDirs := []string{}
	for name := range manifests {
		if name != "package.json" {
			manifestDirs = append(manifestDirs, path.Dir(name))
		}
	}
	sort.Strings(manifestDirs)
	selected := []string{}
	for _, dir := range manifestDirs {
		matched := false
		for _, pattern := range patterns {
			matched = matched || matchWorkspace(pattern, dir)
		}
		if matched {
			selected = append(selected, dir)
		}
	}
	for _, pattern := range patterns {
		found := false
		for _, dir := range selected {
			found = found || matchWorkspace(pattern, dir)
		}
		if !found {
			return nil, fail(CodeLockStale, "workspace pattern has no captured manifest", map[string]string{"workspace": pattern})
		}
	}
	selectedSet := map[string]bool{}
	for _, dir := range selected {
		selectedSet[dir] = true
		idx, ok := index[dir]
		if !ok {
			return nil, fail(CodeGraphIncomplete, "workspace is absent from npm lock packages", map[string]string{"workspace": dir})
		}
		name := manifests[path.Join(dir, "package.json")].Name
		alias := "node_modules/" + name
		if target, ok := links[alias]; !ok || target != dir {
			return nil, fail(CodeGraphIncomplete, "workspace link is absent or disagrees with npm lock", map[string]string{"workspace": dir, "alias": alias})
		}
		_ = packages[idx]
	}
	for _, pkg := range packages {
		if pkg.WorkspacePath != "" && pkg.InstallPath != "" && !selectedSet[pkg.InstallPath] {
			return nil, fail(CodeLockStale, "npm lock contains undeclared workspace", map[string]string{"workspace": pkg.InstallPath})
		}
	}
	return selected, nil
}

func workspacePatterns(raw json.RawMessage) ([]string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return []string{}, nil
	}
	var direct []string
	if json.Unmarshal(raw, &direct) == nil {
		return validateWorkspacePatterns(direct)
	}
	var object struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, fail(CodeLockFormatUnsupported, "unsupported workspaces declaration", nil)
	}
	return validateWorkspacePatterns(object.Packages)
}

func validateWorkspacePatterns(values []string) ([]string, error) {
	result := append([]string(nil), values...)
	sort.Strings(result)
	for i, value := range result {
		if value == "" || path.IsAbs(value) || path.Clean(value) != value || strings.HasPrefix(value, "../") || strings.Contains(value, "\\") || (strings.ContainsAny(value, "?[") || (strings.Contains(value, "*") && !strings.HasSuffix(value, "/*"))) || (i > 0 && result[i-1] == value) {
			return nil, fail(CodeLockFormatUnsupported, "workspace pattern is outside the closed exact/trailing-star grammar", map[string]string{"workspace": value})
		}
	}
	return result, nil
}

func matchWorkspace(pattern, dir string) bool {
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(dir, prefix) && !strings.Contains(strings.TrimPrefix(dir, prefix), "/")
	}
	return pattern == dir
}

func buildEdges(packages []Package, index map[string]int, links map[string]string) ([]DependencyEdge, error) {
	edges := []DependencyEdge{}
	for _, pkg := range packages {
		if pkg.Link {
			continue
		}
		sets := []struct {
			scope  string
			values map[string]string
		}{{"runtime", pkg.Dependencies}, {"development", pkg.DevDependencies}, {"optional", pkg.OptionalDependencies}, {"peer", pkg.PeerDependencies}}
		for _, set := range sets {
			names := sortedMapKeys(set.values)
			for _, name := range names {
				if set.scope == "runtime" {
					if _, overridden := pkg.OptionalDependencies[name]; overridden {
						continue
					}
				}
				to := resolveDependency(pkg.InstallPath, name, index, links)
				optional := set.scope == "optional" || (set.scope == "peer" && pkg.PeerOptional[name])
				if to == "" && !optional {
					return nil, fail(CodeGraphIncomplete, "dependency has no npm lock package", map[string]string{"from": pkg.InstallPath, "dependency": name, "scope": set.scope})
				}
				reason := ""
				if to == "" {
					reason = "optional_missing"
				}
				edges = append(edges, DependencyEdge{From: pkg.InstallPath, Name: name, Spec: set.values[name], Scope: set.scope, To: to, Reason: reason})
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool { return edgeKey(edges[i]) < edgeKey(edges[j]) })
	return edges, nil
}

func resolveDependency(from, name string, index map[string]int, links map[string]string) string {
	base := from
	if from == "" || !isExternalInstallPath(from) {
		base = ""
		for alias, target := range links {
			if target == from {
				base = alias
				break
			}
		}
	}
	for {
		candidate := path.Join(base, "node_modules", name)
		if _, ok := index[candidate]; ok {
			if target, linked := links[candidate]; linked {
				return target
			}
			return candidate
		}
		marker := strings.LastIndex(base, "/node_modules/")
		if marker < 0 {
			if base != "" {
				base = ""
				continue
			}
			return ""
		}
		base = base[:marker]
	}
}

func markSelection(packages []Package, edges []DependencyEdge, index map[string]int, links map[string]string, target Target) error {
	queue := []string{""}
	for _, pkg := range packages {
		if pkg.WorkspacePath != "" && pkg.InstallPath != "" {
			queue = append(queue, pkg.InstallPath)
		}
	}
	seen := map[string]bool{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		idx := index[current]
		pkg := &packages[idx]
		if ok, reason := matchesTarget(*pkg, target); !ok {
			if pkg.Optional {
				pkg.PruneReason = reason
				continue
			}
			return fail(CodeGraphIncomplete, "required npm package is incompatible with selected target", map[string]string{"package": pkg.InstallPath, "reason": reason})
		}
		if pkg.Dev && !target.IncludeDev {
			pkg.PruneReason = "development_omitted"
			continue
		}
		pkg.Selected = true
		for _, edge := range edges {
			if edge.From == current && edge.To != "" {
				if edge.Scope == "development" && !target.IncludeDev {
					continue
				}
				queue = append(queue, edge.To)
			}
		}
	}
	for i := range packages {
		if packages[i].Link {
			packages[i].Selected = packages[index[links[packages[i].InstallPath]]].Selected
			continue
		}
		if !seen[packages[i].InstallPath] {
			if packages[i].Optional {
				packages[i].PruneReason = "unreachable_optional"
			} else if packages[i].Dev && !target.IncludeDev {
				packages[i].PruneReason = "development_omitted"
			} else {
				return fail(CodeGraphIncomplete, "npm lock contains an unreachable package instance", map[string]string{"package": packages[i].InstallPath})
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
	matched := false
	for _, rule := range rules {
		if strings.HasPrefix(rule, "!") {
			if strings.TrimPrefix(rule, "!") == value {
				return false
			}
			continue
		}
		positive = true
		matched = matched || rule == value
	}
	return !positive || matched
}

func validateRemoteIdentity(pkg Package, allowedOrigins []string) error {
	if pkg.Integrity == "" {
		return fail(CodeIntegrityMissing, "npm package has no integrity", map[string]string{"package": pkg.InstallPath})
	}
	if !strings.HasPrefix(pkg.Integrity, "sha512-") || strings.Contains(pkg.Integrity, " ") {
		return fail(CodeIntegrityMissing, "npm package integrity must be one sha512 SRI", map[string]string{"package": pkg.InstallPath})
	}
	resolved, err := url.Parse(pkg.Resolved)
	if err != nil || resolved.Scheme != "https" || resolved.Host == "" || resolved.User != nil || resolved.RawQuery != "" || resolved.Fragment != "" || !strings.HasSuffix(resolved.Path, ".tgz") || !registryOriginAllowed(resolved, allowedOrigins) || strings.ContainsAny(pkg.Resolved, "\x00\r\n") {
		return fail(CodeOriginUnpinned, "npm package locator is not a supported registry tarball", map[string]string{"package": pkg.InstallPath, "scheme": locatorScheme(pkg.Resolved)})
	}
	if pkg.Name == "" || pkg.Version == "" {
		return fail(CodeGraphIncomplete, "npm package identity is incomplete", map[string]string{"package": pkg.InstallPath})
	}
	return nil
}

func registryOriginAllowed(resolved *url.URL, allowed []string) bool {
	for _, raw := range allowed {
		origin, err := url.Parse(raw)
		if err != nil || origin.Scheme != "https" || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" {
			continue
		}
		prefix := strings.TrimSuffix(origin.EscapedPath(), "/") + "/"
		if resolved.Scheme == origin.Scheme && strings.EqualFold(resolved.Host, origin.Host) && strings.HasPrefix(resolved.EscapedPath(), prefix) {
			return true
		}
	}
	return false
}

func validateInstallPath(value string) error {
	if value != "" && (path.IsAbs(value) || path.Clean(value) != value || strings.HasPrefix(value, "../") || strings.Contains(value, "\\")) {
		return fail(CodeLocalPathEscape, "npm lock package path escapes the project", map[string]string{"path": value})
	}
	return nil
}
func isExternalInstallPath(value string) bool {
	return strings.HasPrefix(value, "node_modules/") || strings.Contains(value, "/node_modules/")
}
func packageNameFromInstallPath(value string) string {
	parts := strings.Split(value, "/node_modules/")
	leaf := parts[len(parts)-1]
	leaf = strings.TrimPrefix(leaf, "node_modules/")
	segments := strings.Split(leaf, "/")
	if len(segments) >= 2 && strings.HasPrefix(segments[0], "@") {
		return segments[0] + "/" + segments[1]
	}
	return segments[0]
}
func packageKey(value string) string {
	if value == "" {
		return "root"
	}
	return value
}
func rawBundlePresent(raw json.RawMessage) bool {
	trim := bytes.TrimSpace(raw)
	return len(trim) > 0 && !bytes.Equal(trim, []byte("null")) && !bytes.Equal(trim, []byte("false")) && !bytes.Equal(trim, []byte("[]"))
}
func locatorScheme(value string) string {
	if at := strings.Index(value, ":"); at > 0 {
		return value[:at]
	}
	return "unknown"
}
func firstNonempty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
func cloneMap(value map[string]string) map[string]string {
	if value == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(value))
	for k, v := range value {
		out[k] = v
	}
	return out
}
func peerOptional(value map[string]struct {
	Optional bool `json:"optional"`
}) map[string]bool {
	out := map[string]bool{}
	for k, v := range value {
		out[k] = v.Optional
	}
	return out
}
func sortedStrings(value []string) []string {
	out := append([]string(nil), value...)
	sort.Strings(out)
	return out
}
func sortedMapKeys(value map[string]string) []string {
	out := make([]string, 0, len(value))
	for k := range value {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func equalStringMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
func edgeKey(e DependencyEdge) string {
	return e.From + "\x00" + e.Scope + "\x00" + e.Name + "\x00" + e.To
}

func validateJSON(payload []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}
func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = true
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return fmt.Errorf("unterminated object")
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return fmt.Errorf("unterminated array")
		}
	default:
		return fmt.Errorf("unexpected delimiter")
	}
	return nil
}

func targetCondition(pkg Package) *closuregraph.Condition {
	parts := []string{}
	if len(pkg.OS) > 0 {
		parts = append(parts, "os="+strings.Join(pkg.OS, ","))
	}
	if len(pkg.CPU) > 0 {
		parts = append(parts, "cpu="+strings.Join(pkg.CPU, ","))
	}
	if len(pkg.Libc) > 0 {
		parts = append(parts, "libc="+strings.Join(pkg.Libc, ","))
	}
	if len(parts) == 0 {
		return nil
	}
	return &closuregraph.Condition{EvaluatorID: "npm-platform-v1", Expression: strings.Join(parts, ";")}
}
