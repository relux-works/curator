package yarnmodernsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
	"gopkg.in/yaml.v3"
)

// SupportedYarnVersion is the exact implementation covered by this profile.
const SupportedYarnVersion = "4.9.2"

// ConditionGrammarID binds the pinned Yarn/tinylogic condition semantics used
// both while admitting the lock and while selecting the target graph.
const ConditionGrammarID = "yarn-modern-condition-v1:yarn-4.9.2-tinylogic"

// PeerVirtualizationAlgorithmID binds the closed projection of Yarn 4.9.2's
// applyVirtualResolutionMutations pass. Runtime virtual hashes remain Yarn
// encodings; this identity is derived from the authoritative base locator and
// exact provider instance set instead of trusting those generated hashes.
const PeerVirtualizationAlgorithmID = "yarn-modern-peer-virtualization-v1:yarn-4.9.2"

// BuiltinPlugins is the closed manager-owned plugin set in the pinned release.
var BuiltinPlugins = []string{
	"@yarnpkg/plugin-compat", "@yarnpkg/plugin-constraints", "@yarnpkg/plugin-dlx", "@yarnpkg/plugin-essentials", "@yarnpkg/plugin-exec",
	"@yarnpkg/plugin-file", "@yarnpkg/plugin-git", "@yarnpkg/plugin-github",
	"@yarnpkg/plugin-http", "@yarnpkg/plugin-init", "@yarnpkg/plugin-interactive-tools", "@yarnpkg/plugin-jsr",
	"@yarnpkg/plugin-link", "@yarnpkg/plugin-nm", "@yarnpkg/plugin-npm",
	"@yarnpkg/plugin-npm-cli", "@yarnpkg/plugin-pack", "@yarnpkg/plugin-patch",
	"@yarnpkg/plugin-pnp", "@yarnpkg/plugin-pnpm", "@yarnpkg/plugin-stage",
	"@yarnpkg/plugin-typescript", "@yarnpkg/plugin-version", "@yarnpkg/plugin-workspace-tools",
}

// Target is the exact platform/production selection used for closure.
type Target struct {
	OS, Architecture, Libc string
	IncludeDev             bool
}

// ParseRequest supplies all immutable lock, manifest, rc, patch, and tool inputs.
type ParseRequest struct {
	LockName                          string
	LockBytes                         []byte
	Manifests, Configuration, Patches map[string][]byte
	YarnVersion                       string
	BuiltinPluginSet                  []string
	Target                            Target
}

// DependencyEdge records a resolved descriptor and its selection evidence.
type DependencyEdge struct {
	From, Name, Spec, Scope, To string
	Selected                    bool
	Reason                      string
}

// PeerBinding records the exact provider context of a derived virtual package.
// An empty Provider is permitted only for an explicitly optional peer.
type PeerBinding struct {
	Name, Spec, Provider string
	Optional             bool
}

// Package is one workspace or exact modern Yarn resolution instance.
type Package struct {
	Key, BaseKey, Name, Version, Resolution, Checksum, WorkspacePath string
	// Resolved and Integrity are compatibility projections consumed by the shared capture engine.
	Resolved, Integrity                                                   string
	Selectors                                                             []string
	Dependencies, DevDependencies, OptionalDependencies, PeerDependencies map[string]string
	PeerOptional                                                          map[string]bool
	PeerContext                                                           []PeerBinding
	OS, CPU, Libc, Conditions                                             []string
	Selected                                                              bool
	PruneReason, PatchPath                                                string
	manifest                                                              packageManifest
}

// Layout binds cache normalization, linker, plugin, and condition identity.
type Layout struct {
	NodeLinker                 string
	CompressionLevel           int
	CacheKey                   string
	Conditions, BuiltinPlugins []string
	ConditionGrammar           string
	PeerVirtualization         string
	PnpMode, ModulesFolder     string
	EnableTelemetry            bool
	PnpEnableEsmLoader         bool
	DefaultProtocol            string
	NpmRegistryServer          string
}

// Patch binds one declared lock patch to exact project bytes.
type Patch struct{ Path, SHA256, Locator string }

// Graph is the frozen modern Yarn lock and configuration projection.
type Graph struct {
	LockName, LockDigest, RawLockSHA256, RootName, RootVersion, YarnVersion, ConfigurationDigest string
	Target                                                                                       Target
	Layout                                                                                       Layout
	Packages                                                                                     []Package
	Edges                                                                                        []DependencyEdge
	Workspaces                                                                                   []string
	Patches                                                                                      []Patch
	packageIndex                                                                                 map[string]int
	selectorIndex, workspaceByName                                                               map[string]string
	manifestBytes, configurationBytes, patchBytes                                                map[string][]byte
}
type packageManifest struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	PeerDependenciesMeta map[string]struct {
		Optional bool `json:"optional"`
	} `json:"peerDependenciesMeta"`
	Scripts             map[string]string `json:"scripts"`
	OS                  []string          `json:"os"`
	CPU                 []string          `json:"cpu"`
	Libc                []string          `json:"libc"`
	BundleDependencies  json.RawMessage   `json:"bundleDependencies"`
	BundledDependencies json.RawMessage   `json:"bundledDependencies"`
	Workspaces          json.RawMessage   `json:"workspaces"`
	Resolutions         map[string]string `json:"resolutions"`
	Gypfile             *bool             `json:"gypfile"`
	Private             bool              `json:"private"`
	PackageManager      string            `json:"packageManager"`
}
type lockEntry struct {
	Selectors                                     []string
	Version, Resolution, Checksum, Language, Link string
	Dependencies, PeerDependencies                map[string]string
	Optional, PeerOptional                        map[string]bool
	Conditions                                    []string
}

// Parse validates and reconciles the supported lock v8 profile without executing Yarn.
func Parse(request ParseRequest) (Graph, error) {
	if request.LockName != "yarn.lock" || len(request.LockBytes) == 0 {
		if request.LockName == "" || len(request.LockBytes) == 0 {
			return Graph{}, fail(CodeLockMissing, "an authoritative root yarn.lock is required", nil)
		}
		return Graph{}, fail(CodeLockFormatUnsupported, "unsupported Yarn lock name", map[string]string{"format": request.LockName})
	}
	if request.YarnVersion != SupportedYarnVersion {
		return Graph{}, fail(CodeLockFormatUnsupported, "unsupported modern Yarn implementation", map[string]string{"expected": SupportedYarnVersion, "observed": request.YarnVersion})
	}
	if request.Target.OS == "" || request.Target.Architecture == "" {
		return Graph{}, fail(CodeGraphIncomplete, "target OS and architecture are required", nil)
	}
	manifestBytes, manifests, err := parseManifests(request.Manifests)
	if err != nil {
		return Graph{}, err
	}
	root, ok := manifests["package.json"]
	if !ok || root.Name == "" || root.Version == "" {
		return Graph{}, fail(CodeLockStale, "root manifest identity is missing", nil)
	}
	if root.PackageManager != "yarn@"+SupportedYarnVersion {
		return Graph{}, fail(CodeLockStale, "packageManager does not bind supported Yarn", map[string]string{"expected": "yarn@" + SupportedYarnVersion, "observed": root.PackageManager})
	}
	workspaces, err := reconcileWorkspaces(root.Workspaces, manifests)
	if err != nil {
		return Graph{}, err
	}
	configBytes, layout, configID, err := parseConfiguration(request.Configuration, request.BuiltinPluginSet, request.Target)
	if err != nil {
		return Graph{}, err
	}
	entries, cacheKey, err := parseLock(request.LockBytes)
	if err != nil {
		return Graph{}, err
	}
	if cacheKey != layout.CacheKey {
		return Graph{}, fail(CodeLockStale, "lock and rc cache keys differ", map[string]string{"lock": cacheKey, "rc": layout.CacheKey})
	}
	for _, entry := range entries {
		if entry.Checksum != "" && !validYarnChecksum(entry.Checksum, cacheKey) {
			return Graph{}, fail(CodeIntegrityMissing, "modern Yarn checksum is not canonical for the bound cache key", map[string]string{"resolution": entry.Resolution})
		}
	}
	patches, patchBytes, err := reconcilePatches(entries, request.Patches)
	if err != nil {
		return Graph{}, err
	}
	packages := []Package{}
	index := map[string]int{}
	workByName := map[string]string{}
	addWorkspace := func(key, workspace string, m packageManifest) error {
		if m.Name == "" || m.Version == "" || workByName[m.Name] != "" {
			return fail(CodeGraphIncomplete, "workspace identity is missing or duplicated", map[string]string{"workspace": workspace})
		}
		if rawPresent(m.BundleDependencies) || rawPresent(m.BundledDependencies) {
			return fail(CodeBundledDependencyUnsupported, "bundled dependencies are unsupported", map[string]string{"workspace": workspace})
		}
		p := Package{Key: key, Name: m.Name, Version: m.Version, WorkspacePath: workspace, Dependencies: cloneMap(m.Dependencies), DevDependencies: cloneMap(m.DevDependencies), OptionalDependencies: cloneMap(m.OptionalDependencies), PeerDependencies: cloneMap(m.PeerDependencies), PeerOptional: peerOptional(m.PeerDependenciesMeta), OS: sortedStrings(m.OS), CPU: sortedStrings(m.CPU), Libc: sortedStrings(m.Libc), manifest: m}
		index[key] = len(packages)
		workByName[m.Name] = key
		packages = append(packages, p)
		return nil
	}
	if err = addWorkspace("workspace:.", ".", root); err != nil {
		return Graph{}, err
	}
	for _, w := range workspaces {
		if err = addWorkspace("workspace:"+w, w, manifests[path.Join(w, "package.json")]); err != nil {
			return Graph{}, err
		}
	}
	selectors := map[string]string{}
	workspaceEntries := map[string]bool{}
	for _, e := range entries {
		name, protocol, identityErr := resolutionIdentity(e.Resolution)
		if identityErr != nil {
			return Graph{}, identityErr
		}
		if protocol == "workspace" {
			workspace, workspaceErr := workspacePathFromResolution(e.Resolution)
			if workspaceErr != nil {
				return Graph{}, workspaceErr
			}
			key := "workspace:" + workspace
			packagePosition, exists := index[key]
			if !exists || workspaceEntries[workspace] {
				return Graph{}, fail(CodeLockStale, "workspace lock entry is unmatched or duplicated", map[string]string{"workspace": workspace})
			}
			if err = reconcileWorkspaceLockEntry(e, packages[packagePosition], manifests); err != nil {
				return Graph{}, err
			}
			workspaceEntries[workspace] = true
			for _, s := range e.Selectors {
				if selectors[s] != "" {
					return Graph{}, fail(CodeLockFormatUnsupported, "descriptor appears twice", map[string]string{"descriptor": s})
				}
				selectors[s] = key
			}
			continue
		}
		key := e.Resolution
		if _, exists := index[key]; exists {
			return Graph{}, fail(CodeGraphIncomplete, "duplicate resolution", map[string]string{"resolution": key})
		}
		patchPath := ""
		if protocol == "patch" {
			patchPath, err = patchPathFromResolution(e.Resolution)
			if err != nil {
				return Graph{}, err
			}
		}
		p := Package{Key: key, Name: name, Version: e.Version, Resolution: e.Resolution, Checksum: e.Checksum, Resolved: e.Resolution, Integrity: e.Checksum, Selectors: append([]string(nil), e.Selectors...), Dependencies: cloneMap(e.Dependencies), OptionalDependencies: map[string]string{}, PeerDependencies: cloneMap(e.PeerDependencies), PeerOptional: cloneBoolMap(e.PeerOptional), Conditions: append([]string(nil), e.Conditions...), PatchPath: patchPath}
		for dep, optional := range e.Optional {
			if optional {
				p.OptionalDependencies[dep] = p.Dependencies[dep]
				delete(p.Dependencies, dep)
			}
		}
		for _, s := range e.Selectors {
			if selectors[s] != "" {
				return Graph{}, fail(CodeLockFormatUnsupported, "descriptor appears twice", map[string]string{"descriptor": s})
			}
			selectors[s] = key
		}
		index[key] = len(packages)
		packages = append(packages, p)
	}
	if len(workspaceEntries) != len(workByName) {
		return Graph{}, fail(CodeLockStale, "a declared workspace lacks an authoritative lock entry", map[string]string{"expected": fmt.Sprint(len(workByName)), "observed": fmt.Sprint(len(workspaceEntries))})
	}
	packages, edges, index, err := buildPackageGraph(packages, selectors, workByName)
	if err != nil {
		return Graph{}, err
	}
	if err = markSelection(packages, edges, index, request.Target); err != nil {
		return Graph{}, err
	}
	raw := sha256.Sum256(request.LockBytes)
	records := make([]any, 0, len(entries))
	for _, e := range entries {
		records = append(records, map[string]any{"selectors": stringsAny(e.Selectors), "version": e.Version, "resolution": e.Resolution, "checksum": e.Checksum, "language": e.Language, "link": e.Link, "dependencies": mapAny(e.Dependencies), "dependency_optional": boolMapAny(e.Optional), "peers": mapAny(e.PeerDependencies), "peer_optional": boolMapAny(e.PeerOptional), "conditions": stringsAny(e.Conditions)})
	}
	derivedPackages := make([]any, 0, len(packages))
	for _, pkg := range packages {
		if pkg.BaseKey == "" {
			continue
		}
		derivedPackages = append(derivedPackages, map[string]any{"key": pkg.Key, "base_key": pkg.BaseKey, "peer_context": peerBindingsAny(pkg.PeerContext)})
	}
	derivedEdges := make([]any, 0, len(edges))
	for _, edge := range edges {
		derivedEdges = append(derivedEdges, map[string]any{"from": edge.From, "name": edge.Name, "spec": edge.Spec, "scope": edge.Scope, "to": edge.To, "reason": edge.Reason})
	}
	canonical, err := protocoljson.MarshalCanonical(map[string]any{"schema": "yarn-modern-lock-v1", "manager": request.YarnVersion, "cache_key": cacheKey, "configuration": configID, "manifests": bytesDigest(manifestBytes), "patches": patchesAny(patches), "entries": records, "peer_virtualization": PeerVirtualizationAlgorithmID, "derived_packages": derivedPackages, "derived_edges": derivedEdges})
	if err != nil {
		return Graph{}, err
	}
	lockID, err := closuregraph.IDFromCanonical("yarn-modern-lock-v1", canonical)
	if err != nil {
		return Graph{}, err
	}
	return Graph{LockName: "yarn.lock", LockDigest: string(lockID), RawLockSHA256: "sha256:" + hex.EncodeToString(raw[:]), RootName: root.Name, RootVersion: root.Version, YarnVersion: request.YarnVersion, ConfigurationDigest: configID, Target: request.Target, Layout: layout, Packages: packages, Edges: edges, Workspaces: workspaces, Patches: patches, packageIndex: index, selectorIndex: selectors, workspaceByName: workByName, manifestBytes: manifestBytes, configurationBytes: configBytes, patchBytes: patchBytes}, nil
}

func parseLock(payload []byte) ([]lockEntry, string, error) {
	if bytes.Contains(payload, []byte("\r")) || !bytes.HasSuffix(payload, []byte("\n")) {
		return nil, "", fail(CodeLockFormatUnsupported, "yarn.lock must use LF and terminal newline", nil)
	}
	var root map[string]any
	dec := yaml.NewDecoder(bytes.NewReader(payload))
	if err := dec.Decode(&root); err != nil {
		return nil, "", fail(CodeLockFormatUnsupported, "cannot parse yarn.lock: "+err.Error(), nil)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return nil, "", fail(CodeLockFormatUnsupported, "yarn.lock must contain exactly one document", nil)
	}
	meta, ok := root["__metadata"].(map[string]any)
	if !ok || !onlyKeys(meta, "version", "cacheKey") {
		return nil, "", fail(CodeLockFormatUnsupported, "unsupported modern Yarn lock schema", nil)
	}
	version, versionOK := meta["version"].(int)
	cacheKey, cacheKeyOK := meta["cacheKey"].(string)
	if !versionOK || version != 8 || !cacheKeyOK || cacheKey == "" {
		return nil, "", fail(CodeLockFormatUnsupported, "lock cacheKey is missing", nil)
	}
	entries := []lockEntry{}
	for descriptor, raw := range root {
		if descriptor == "__metadata" {
			continue
		}
		fields, ok := raw.(map[string]any)
		if !ok || !onlyKeys(fields, "version", "resolution", "checksum", "languageName", "linkType", "dependencies", "peerDependencies", "dependenciesMeta", "peerDependenciesMeta", "conditions") {
			return nil, "", fail(CodeLockFormatUnsupported, "unsupported lock entry", map[string]string{"descriptor": descriptor})
		}
		selectors, err := descriptorList(descriptor)
		if err != nil {
			return nil, "", err
		}
		version, versionErr := lockString(fields, "version", true)
		resolution, resolutionErr := lockString(fields, "resolution", true)
		checksum, checksumErr := lockString(fields, "checksum", false)
		language, languageErr := lockString(fields, "languageName", true)
		link, linkErr := lockString(fields, "linkType", true)
		dependencies, dependenciesErr := lockStringMap(fields, "dependencies")
		peers, peersErr := lockStringMap(fields, "peerDependencies")
		optional, optionalErr := lockMetaOptional(fields, "dependenciesMeta")
		peerOptional, peerOptionalErr := lockMetaOptional(fields, "peerDependenciesMeta")
		conditions, conditionsErr := lockStringSlice(fields, "conditions")
		if firstError(versionErr, resolutionErr, checksumErr, languageErr, linkErr, dependenciesErr, peersErr, optionalErr, peerOptionalErr, conditionsErr) != nil {
			return nil, "", fail(CodeLockFormatUnsupported, "lock entry contains a missing or type-confused field", map[string]string{"descriptor": descriptor})
		}
		for _, condition := range conditions {
			if _, _, conditionErr := evaluateYarnCondition(condition, map[string]string{"os": "validation-os", "cpu": "validation-cpu", "libc": "validation-libc"}); conditionErr != nil {
				return nil, "", fail(CodeLockFormatUnsupported, "lock entry contains a condition outside the pinned Yarn grammar", map[string]string{"descriptor": descriptor, "condition": condition})
			}
		}
		e := lockEntry{Selectors: selectors, Version: version, Resolution: resolution, Checksum: checksum, Language: language, Link: link, Dependencies: dependencies, PeerDependencies: peers, Optional: optional, PeerOptional: peerOptional, Conditions: conditions}
		if e.Version == "" || e.Resolution == "" {
			return nil, "", fail(CodeLockFormatUnsupported, "lock entry lacks identity", map[string]string{"descriptor": descriptor})
		}
		_, protocol, err := resolutionIdentity(e.Resolution)
		if err != nil {
			return nil, "", err
		}
		if protocol != "workspace" && e.Checksum == "" {
			return nil, "", fail(CodeIntegrityMissing, "package checksum is missing", map[string]string{"resolution": e.Resolution})
		}
		if protocol == "workspace" {
			if e.Language != "unknown" || e.Link != "soft" {
				return nil, "", fail(CodeLockFormatUnsupported, "workspace lock entry lacks Yarn-owned language/link identity", map[string]string{"resolution": e.Resolution})
			}
		} else if e.Language != "node" || e.Link != "hard" {
			return nil, "", fail(CodeLockFormatUnsupported, "package lock entry lacks Yarn-owned language/link identity", map[string]string{"resolution": e.Resolution})
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Resolution < entries[j].Resolution })
	return entries, cacheKey, nil
}

func parseConfiguration(values map[string][]byte, supplied []string, target Target) (map[string][]byte, Layout, string, error) {
	payload, ok := values[".yarnrc.yml"]
	if !ok || len(values) != 1 {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "exactly one root .yarnrc.yml is required", nil)
	}
	var fields map[string]any
	decoder := yaml.NewDecoder(bytes.NewReader(payload))
	if err := decoder.Decode(&fields); err != nil || fields == nil {
		return nil, Layout{}, "", fail(CodeLockFormatUnsupported, "cannot parse .yarnrc.yml", nil)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, Layout{}, "", fail(CodeLockFormatUnsupported, ".yarnrc.yml must contain exactly one document", nil)
	}
	allowed := []string{"nodeLinker", "compressionLevel", "cacheFolder", "enableGlobalCache", "enableNetwork", "enableImmutableInstalls", "enableScripts", "enableTelemetry", "checksumBehavior", "pnpMode", "pnpEnableEsmLoader", "supportedArchitectures", "defaultProtocol", "npmRegistryServer"}
	if !onlyKeys(fields, allowed...) {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, ".yarnrc.yml contains unsupported extension or setting", nil)
	}
	linker, err := requiredStringField(fields, "nodeLinker")
	if err != nil {
		return nil, Layout{}, "", err
	}
	if linker != "pnp" && linker != "node-modules" {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "unsupported Yarn linker", map[string]string{"nodeLinker": linker})
	}
	compression, err := strictIntegerField(fields, "compressionLevel")
	if err != nil || compression < 0 || compression > 9 {
		return nil, Layout{}, "", fail(CodeLockFormatUnsupported, "compressionLevel must be 0..9", nil)
	}
	cacheFolder, cacheErr := requiredStringField(fields, "cacheFolder")
	enableGlobalCache, globalErr := requiredBoolField(fields, "enableGlobalCache")
	enableNetwork, networkErr := requiredBoolField(fields, "enableNetwork")
	enableImmutableInstalls, immutableErr := requiredBoolField(fields, "enableImmutableInstalls")
	enableScripts, scriptsErr := requiredBoolField(fields, "enableScripts")
	checksumBehavior, checksumErr := requiredStringField(fields, "checksumBehavior")
	if firstError(cacheErr, globalErr, networkErr, immutableErr, scriptsErr, checksumErr) != nil {
		return nil, Layout{}, "", fail(CodeLockFormatUnsupported, ".yarnrc.yml contains a missing or type-confused policy setting", nil)
	}
	if cacheFolder != ".yarn/cache" || enableGlobalCache || enableNetwork || !enableImmutableInstalls || enableScripts || checksumBehavior != "throw" {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "immutable offline cache and skip-script policy is required", nil)
	}
	pnpMode, err := optionalStringField(fields, "pnpMode", "strict")
	if err != nil {
		return nil, Layout{}, "", err
	}
	if linker == "pnp" && pnpMode != "strict" {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "PnP must use strict mode", nil)
	}
	enableTelemetry, err := optionalBoolField(fields, "enableTelemetry", false)
	if err != nil {
		return nil, Layout{}, "", err
	}
	if enableTelemetry {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "Yarn telemetry must be disabled", nil)
	}
	pnpEnableEsmLoader, err := optionalBoolField(fields, "pnpEnableEsmLoader", false)
	if err != nil {
		return nil, Layout{}, "", err
	}
	if pnpEnableEsmLoader {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "the experimental PnP ESM loader is unsupported", nil)
	}
	defaultProtocol, err := optionalStringField(fields, "defaultProtocol", "npm:")
	if err != nil {
		return nil, Layout{}, "", err
	}
	if defaultProtocol != "npm:" {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "defaultProtocol must use the pinned npm resolver", map[string]string{"defaultProtocol": defaultProtocol})
	}
	npmRegistryServer, err := optionalStringField(fields, "npmRegistryServer", "https://registry.yarnpkg.com")
	if err != nil {
		return nil, Layout{}, "", err
	}
	if npmRegistryServer != "https://registry.yarnpkg.com" {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "custom npm registry configuration is unsupported", map[string]string{"npmRegistryServer": npmRegistryServer})
	}
	builtins := sortedStrings(supplied)
	if len(builtins) == 0 {
		builtins = sortedStrings(BuiltinPlugins)
	}
	if !equalStrings(builtins, sortedStrings(BuiltinPlugins)) {
		return nil, Layout{}, "", fail(CodeManagerPluginUndeclared, "built-in plugin set differs from pinned release", nil)
	}
	conditions := []string{}
	if raw := fields["supportedArchitectures"]; raw != nil {
		architectures, ok := raw.(map[string]any)
		if !ok || !onlyKeys(architectures, "os", "cpu", "libc") {
			return nil, Layout{}, "", fail(CodeLockFormatUnsupported, "supportedArchitectures is malformed", nil)
		}
		for key, actual := range map[string]string{"os": target.OS, "cpu": target.Architecture, "libc": target.Libc} {
			values, valuesErr := strictArchitectureValues(architectures, key)
			if valuesErr != nil {
				return nil, Layout{}, "", valuesErr
			}
			if len(values) > 0 && !contains(values, actual) {
				return nil, Layout{}, "", fail(CodeGraphIncomplete, "target outside supportedArchitectures", map[string]string{"field": key, "target": actual})
			}
			for _, v := range values {
				conditions = append(conditions, key+"="+v)
			}
		}
	}
	sort.Strings(conditions)
	cacheKey := fmt.Sprintf("10c%d", compression)
	layout := Layout{NodeLinker: linker, CompressionLevel: compression, CacheKey: cacheKey, Conditions: conditions, BuiltinPlugins: builtins, ConditionGrammar: ConditionGrammarID, PeerVirtualization: PeerVirtualizationAlgorithmID, PnpMode: pnpMode, ModulesFolder: "node_modules", EnableTelemetry: enableTelemetry, PnpEnableEsmLoader: pnpEnableEsmLoader, DefaultProtocol: defaultProtocol, NpmRegistryServer: npmRegistryServer}
	id, err := closuregraph.DomainID("yarn-modern-configuration-v1", map[string]any{"linker": linker, "compression": compression, "cache_key": cacheKey, "conditions": stringsAny(conditions), "condition_grammar": ConditionGrammarID, "peer_virtualization": PeerVirtualizationAlgorithmID, "plugins": stringsAny(builtins), "pnp_mode": pnpMode, "enable_telemetry": enableTelemetry, "pnp_enable_esm_loader": pnpEnableEsmLoader, "default_protocol": defaultProtocol, "npm_registry_server": npmRegistryServer})
	if err != nil {
		return nil, Layout{}, "", err
	}
	return map[string][]byte{".yarnrc.yml": append([]byte(nil), payload...)}, layout, string(id), nil
}

func reconcilePatches(entries []lockEntry, supplied map[string][]byte) ([]Patch, map[string][]byte, error) {
	needed := map[string]string{}
	for _, e := range entries {
		_, protocol, _ := resolutionIdentity(e.Resolution)
		if protocol == "patch" {
			p, err := patchPathFromResolution(e.Resolution)
			if err != nil {
				return nil, nil, err
			}
			needed[p] = e.Resolution
		}
	}
	if len(needed) != len(supplied) {
		return nil, nil, fail(CodeManagerPluginUndeclared, "patch set differs from lock", map[string]string{"expected": fmt.Sprint(len(needed)), "observed": fmt.Sprint(len(supplied))})
	}
	out := []Patch{}
	copies := map[string][]byte{}
	for p, locator := range needed {
		b, ok := supplied[p]
		if !ok || len(b) == 0 || !validRelative(p, ".yarn/patches") || bytes.Contains(b, []byte("\r")) {
			return nil, nil, fail(CodeManagerPluginUndeclared, "declared patch is missing or unsupported", map[string]string{"path": p})
		}
		sum := sha256.Sum256(b)
		out = append(out, Patch{Path: p, SHA256: "sha256:" + hex.EncodeToString(sum[:]), Locator: locator})
		copies[p] = append([]byte(nil), b...)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, copies, nil
}
func buildPackageGraph(basePackages []Package, selectors, workspaces map[string]string) ([]Package, []DependencyEdge, map[string]int, error) {
	packages := make([]Package, 0, len(basePackages))
	index := map[string]int{}
	base := map[string]Package{}
	for _, pkg := range basePackages {
		if pkg.BaseKey != "" {
			continue
		}
		pkg.Selected = false
		pkg.PruneReason = ""
		pkg.PeerContext = nil
		base[pkg.Key] = pkg
		index[pkg.Key] = len(packages)
		packages = append(packages, pkg)
	}
	edges := []DependencyEdge{}
	expanded := map[string]bool{}
	expanding := map[string]bool{}
	activeContextBySource := map[string]string{}

	var expand func(string) error
	expand = func(instanceKey string) error {
		if expanded[instanceKey] {
			return nil
		}
		if expanding[instanceKey] {
			// Runtime and peer dependency edges are non-ordering graph evidence.
			// Revisiting the exact in-progress instance therefore closes a valid
			// SCC; only a return to the same source through a different derived
			// peer context is non-well-founded.
			return nil
		}
		instance := packages[index[instanceKey]]
		sourceKey := packageSourceKey(instance)
		if activeContext, ok := activeContextBySource[sourceKey]; ok && activeContext != instanceKey {
			return fail(CodeGraphIncomplete, "peer virtualization contains a non-well-founded context cycle", map[string]string{"package": sourceKey, "active_context": activeContext, "recursive_context": instanceKey})
		}
		expanding[instanceKey] = true
		activeContextBySource[sourceKey] = instanceKey
		defer func() {
			delete(expanding, instanceKey)
			delete(activeContextBySource, sourceKey)
		}()
		source := base[sourceKey]
		declarations := packageDependencyDeclarations(source)
		children := map[string]string{}
		for _, declaration := range declarations {
			to := resolveDependency(declaration.name, declaration.spec, selectors, workspaces)
			if to == "" {
				if declaration.scope == "optional" {
					edges = append(edges, DependencyEdge{From: instanceKey, Name: declaration.name, Spec: declaration.spec, Scope: declaration.scope, Reason: "optional_missing"})
					continue
				}
				return fail(CodeGraphIncomplete, "dependency descriptor is absent", map[string]string{"from": instanceKey, "descriptor": declaration.name + "@" + declaration.spec, "scope": declaration.scope})
			}
			children[declaration.name] = to
		}
		for _, binding := range instance.PeerContext {
			if binding.Provider != "" {
				children[binding.Name] = binding.Provider
			}
		}

		limit := len(children)*len(children) + len(children) + 1
		for pass := 0; pass < limit; pass++ {
			changed := false
			for _, declaration := range declarations {
				baseKey := resolveDependency(declaration.name, declaration.spec, selectors, workspaces)
				if baseKey == "" {
					continue
				}
				childKey, child, err := virtualPackageFor(base[baseKey], instance, children, selectors, workspaces, base)
				if err != nil {
					return err
				}
				if childKey != baseKey {
					if _, ok := index[childKey]; !ok {
						index[childKey] = len(packages)
						packages = append(packages, child)
					}
				}
				if children[declaration.name] != childKey {
					children[declaration.name] = childKey
					changed = true
				}
			}
			if !changed {
				break
			}
			if pass == limit-1 {
				return fail(CodeGraphIncomplete, "peer virtualization did not converge", map[string]string{"package": instanceKey})
			}
		}

		for _, declaration := range declarations {
			to := children[declaration.name]
			if to == "" {
				continue
			}
			edges = append(edges, DependencyEdge{From: instanceKey, Name: declaration.name, Spec: declaration.spec, Scope: declaration.scope, To: to})
		}
		for _, binding := range instance.PeerContext {
			reason := ""
			if binding.Provider == "" {
				reason = "optional_peer_unresolved"
			}
			edges = append(edges, DependencyEdge{From: instanceKey, Name: binding.Name, Spec: binding.Spec, Scope: "peer", To: binding.Provider, Reason: reason})
		}
		for _, child := range children {
			if err := expand(child); err != nil {
				return err
			}
		}
		expanded[instanceKey] = true
		return nil
	}

	for _, pkg := range append([]Package(nil), packages...) {
		if err := expand(pkg.Key); err != nil {
			return nil, nil, nil, err
		}
	}
	referenced := map[string]bool{}
	for _, edge := range edges {
		if edge.To != "" {
			referenced[edge.To] = true
		}
	}
	retained := packages[:0]
	for _, pkg := range packages {
		if pkg.BaseKey == "" || referenced[pkg.Key] {
			retained = append(retained, pkg)
		}
	}
	packages = retained
	sort.Slice(packages, func(i, j int) bool { return packages[i].Key < packages[j].Key })
	index = map[string]int{}
	for i := range packages {
		index[packages[i].Key] = i
	}
	sort.Slice(edges, func(i, j int) bool { return edgeKey(edges[i]) < edgeKey(edges[j]) })
	return packages, edges, index, nil
}

type packageDependencyDeclaration struct{ name, spec, scope string }

func packageDependencyDeclarations(pkg Package) []packageDependencyDeclaration {
	result := []packageDependencyDeclaration{}
	for _, set := range []struct {
		scope string
		items map[string]string
	}{{"runtime", pkg.Dependencies}, {"development", pkg.DevDependencies}, {"optional", pkg.OptionalDependencies}} {
		for _, name := range sortedKeys(set.items) {
			result = append(result, packageDependencyDeclaration{name: name, spec: set.items[name], scope: set.scope})
		}
	}
	return result
}

func virtualPackageFor(base, parent Package, providers, selectors, workspaces map[string]string, basePackages map[string]Package) (string, Package, error) {
	peers, optional := effectivePeerDeclarations(base)
	if len(peers) == 0 {
		return base.Key, base, nil
	}
	bindings := make([]PeerBinding, 0, len(peers))
	for _, name := range sortedKeys(peers) {
		spec := peers[name]
		provider := providers[name]
		if provider == "" && parent.Name == name {
			provider = parent.Key
		}
		if provider == "" {
			provider = resolveOwnPeerDefault(base, name, selectors, workspaces)
		}
		if provider == "" {
			if !optional[name] {
				return "", Package{}, fail(CodeGraphIncomplete, "required peer provider is absent", map[string]string{"package": base.Key, "peer": name, "range": spec, "parent": parent.Key})
			}
			bindings = append(bindings, PeerBinding{Name: name, Spec: spec, Optional: true})
			continue
		}
		providerPackage, ok := basePackages[providerSourceKey(provider)]
		if provider == parent.Key {
			providerPackage, ok = parent, true
		}
		if !ok || !modernPeerVersionSatisfies(providerPackage.Version, spec) {
			return "", Package{}, fail(CodeGraphIncomplete, "peer provider is incompatible or ambiguous", map[string]string{"package": base.Key, "peer": name, "range": spec, "provider": provider})
		}
		bindings = append(bindings, PeerBinding{Name: name, Spec: spec, Provider: provider, Optional: optional[name]})
	}
	id, err := closuregraph.DomainID(PeerVirtualizationAlgorithmID, map[string]any{"base_locator": base.Key, "providers": peerBindingsAny(bindings)})
	if err != nil {
		return "", Package{}, err
	}
	virtual := base
	virtual.BaseKey = base.Key
	virtual.Key = base.Key + "#peer:" + strings.TrimPrefix(string(id), "sha256:")
	virtual.PeerContext = bindings
	virtual.Selected = false
	virtual.PruneReason = ""
	return virtual.Key, virtual, nil
}

func providerSourceKey(key string) string {
	if marker := strings.Index(key, "#peer:"); marker >= 0 {
		return key[:marker]
	}
	return key
}

func resolveOwnPeerDefault(pkg Package, name string, selectors, workspaces map[string]string) string {
	for _, values := range []map[string]string{pkg.Dependencies, pkg.OptionalDependencies, pkg.DevDependencies} {
		if spec := values[name]; spec != "" {
			return resolveDependency(name, spec, selectors, workspaces)
		}
	}
	return ""
}

func effectivePeerDeclarations(pkg Package) (map[string]string, map[string]bool) {
	peers := cloneMap(pkg.PeerDependencies)
	optional := cloneBoolMap(pkg.PeerOptional)
	for name := range pkg.PeerDependencies {
		if strings.HasPrefix(name, "@types/") {
			continue
		}
		typesName := "@types/" + name
		if strings.HasPrefix(name, "@") {
			parts := strings.SplitN(strings.TrimPrefix(name, "@"), "/", 2)
			if len(parts) == 2 {
				typesName = "@types/" + parts[0] + "__" + parts[1]
			}
		}
		if _, peerExists := peers[typesName]; peerExists || pkg.Dependencies[typesName] != "" || pkg.OptionalDependencies[typesName] != "" || pkg.DevDependencies[typesName] != "" {
			continue
		}
		peers[typesName] = "*"
		optional[typesName] = true
	}
	return peers, optional
}

func packageSourceKey(pkg Package) string {
	if pkg.BaseKey != "" {
		return pkg.BaseKey
	}
	return pkg.Key
}

func peerBindingsAny(values []PeerBinding) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = map[string]any{"name": value.Name, "spec": value.Spec, "provider": value.Provider, "optional": value.Optional}
	}
	return result
}
func markSelection(packages []Package, edges []DependencyEdge, index map[string]int, target Target) error {
	root, ok := index["workspace:."]
	if !ok {
		return fail(CodeGraphIncomplete, "root workspace is absent", nil)
	}
	for i := range packages {
		packages[i].PruneReason = "unreachable"
	}
	queue := []string{}
	for i := range packages {
		if packages[i].BaseKey == "" && packages[i].WorkspacePath != "" {
			packages[i].Selected = true
			packages[i].PruneReason = ""
			queue = append(queue, packages[i].Key)
		}
	}
	if !packages[root].Selected {
		packages[root].Selected = true
		packages[root].PruneReason = ""
		queue = append(queue, "workspace:.")
	}
	sort.Strings(queue)
	for len(queue) > 0 {
		from := queue[0]
		queue = queue[1:]
		for i := range edges {
			e := &edges[i]
			if e.From != from || e.To == "" {
				continue
			}
			if e.Scope == "development" && !target.IncludeDev {
				e.Reason = "development_omitted"
				continue
			}
			position, exists := index[e.To]
			if !exists {
				return fail(CodeGraphIncomplete, "dependency context target is absent", map[string]string{"package": e.To})
			}
			p := &packages[position]
			ok, reason, conditionErr := conditionsMatch(p.Conditions, target)
			if conditionErr != nil {
				return conditionErr
			}
			if !ok {
				if e.Scope != "optional" && (e.Scope != "peer" || !packages[index[from]].PeerOptional[e.Name]) {
					return fail(CodeGraphIncomplete, "required condition rejects target", map[string]string{"package": e.To})
				}
				e.Reason = reason
				p.PruneReason = reason
				continue
			}
			e.Selected = true
			e.Reason = ""
			if !p.Selected {
				p.Selected = true
				p.PruneReason = ""
				queue = append(queue, e.To)
				if p.BaseKey != "" {
					basePosition, baseExists := index[p.BaseKey]
					if !baseExists {
						return fail(CodeGraphIncomplete, "virtual package base locator is absent", map[string]string{"package": p.Key})
					}
					base := &packages[basePosition]
					if !base.Selected {
						base.Selected = true
						base.PruneReason = ""
						queue = append(queue, base.Key)
					}
				}
			}
		}
	}
	return nil
}

func parseManifests(values map[string][]byte) (map[string][]byte, map[string]packageManifest, error) {
	copies := map[string][]byte{}
	out := map[string]packageManifest{}
	for p, b := range values {
		if p != "package.json" && (!validRelative(p, ".") || !strings.HasSuffix(p, "/package.json")) {
			return nil, nil, fail(CodeLocalPathEscape, "manifest path escapes project", map[string]string{"path": p})
		}
		var m packageManifest
		if err := validateJSON(b); err != nil {
			return nil, nil, fail(CodeLockStale, "manifest outside closed grammar: "+err.Error(), map[string]string{"path": p})
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, nil, fail(CodeLockStale, "manifest cannot be decoded", map[string]string{"path": p})
		}
		copies[p] = append([]byte(nil), b...)
		out[p] = m
	}
	return copies, out, nil
}
func reconcileWorkspaces(raw json.RawMessage, manifests map[string]packageManifest) ([]string, error) {
	if !rawPresent(raw) {
		if len(manifests) != 1 {
			return nil, fail(CodeLockStale, "undeclared workspace manifest", nil)
		}
		return nil, nil
	}
	var patterns []string
	if json.Unmarshal(raw, &patterns) != nil {
		var wrapped struct {
			Packages []string `json:"packages"`
		}
		if json.Unmarshal(raw, &wrapped) != nil {
			return nil, fail(CodeLockStale, "invalid workspaces", nil)
		}
		patterns = wrapped.Packages
	}
	out := []string{}
	for p := range manifests {
		if p == "package.json" {
			continue
		}
		dir := path.Dir(p)
		matched := false
		for _, pattern := range patterns {
			if pattern == dir || (strings.HasSuffix(pattern, "/*") && path.Dir(dir) == strings.TrimSuffix(pattern, "/*")) {
				matched = true
			}
		}
		if !matched {
			return nil, fail(CodeLockStale, "workspace manifest is undeclared", map[string]string{"path": p})
		}
		out = append(out, dir)
	}
	sort.Strings(out)
	return out, nil
}

func workspacePathFromResolution(resolution string) (string, error) {
	marker := "@workspace:"
	position := strings.LastIndex(resolution, marker)
	if position <= 0 {
		return "", fail(CodeLockStale, "workspace resolution is malformed", map[string]string{"resolution": resolution})
	}
	workspace := resolution[position+len(marker):]
	if workspace != "." && !validRelative(workspace, ".") {
		return "", fail(CodeLocalPathEscape, "workspace resolution escapes the project", map[string]string{"workspace": workspace})
	}
	return workspace, nil
}

func reconcileWorkspaceLockEntry(entry lockEntry, pkg Package, manifests map[string]packageManifest) error {
	name, _, err := resolutionIdentity(entry.Resolution)
	if err != nil {
		return err
	}
	if name != pkg.Name || entry.Resolution != pkg.Name+"@workspace:"+pkg.WorkspacePath || entry.Version != "0.0.0-use.local" || entry.Checksum != "" || len(entry.Conditions) != 0 {
		return fail(CodeLockStale, "workspace lock identity differs from its manifest", map[string]string{"workspace": pkg.WorkspacePath})
	}
	expectedDependencies, err := workspaceDependencyProjection(pkg.manifest)
	if err != nil {
		return err
	}
	expectedOptional := map[string]bool{}
	for dependency := range pkg.manifest.OptionalDependencies {
		expectedOptional[dependency] = true
	}
	if !equalStringMap(entry.Dependencies, expectedDependencies) || !equalBoolMap(entry.Optional, expectedOptional) || !equalStringMap(entry.PeerDependencies, pkg.manifest.PeerDependencies) || !equalBoolMap(entry.PeerOptional, peerOptional(pkg.manifest.PeerDependenciesMeta)) {
		return fail(CodeLockStale, "workspace lock dependency metadata differs from its manifest", map[string]string{"workspace": pkg.WorkspacePath})
	}
	expectedSelectors := map[string]bool{pkg.Name + "@workspace:" + pkg.WorkspacePath: true}
	for _, manifest := range manifests {
		for dependency, spec := range workspaceDependencyDeclarations(manifest) {
			if dependency == pkg.Name && strings.HasPrefix(spec, "workspace:") {
				expectedSelectors[pkg.Name+"@"+spec] = true
			}
		}
	}
	if len(entry.Selectors) != len(expectedSelectors) {
		return fail(CodeLockStale, "workspace lock selectors differ from declared workspace references", map[string]string{"workspace": pkg.WorkspacePath})
	}
	for _, selector := range entry.Selectors {
		if !expectedSelectors[selector] {
			return fail(CodeLockStale, "workspace lock selector is undeclared", map[string]string{"workspace": pkg.WorkspacePath, "descriptor": selector})
		}
	}
	return nil
}

func workspaceDependencyProjection(manifest packageManifest) (map[string]string, error) {
	out := map[string]string{}
	for scope, dependencies := range map[string]map[string]string{
		"dependency": manifest.Dependencies, "development": manifest.DevDependencies, "optional": manifest.OptionalDependencies,
	} {
		for name, spec := range dependencies {
			if prior, exists := out[name]; exists && prior != spec {
				return nil, fail(CodeLockStale, "workspace dependency scopes conflict", map[string]string{"dependency": name, "scope": scope})
			}
			out[name] = spec
		}
	}
	return out, nil
}

func workspaceDependencyDeclarations(manifest packageManifest) map[string]string {
	out, _ := workspaceDependencyProjection(manifest)
	for name, spec := range manifest.PeerDependencies {
		if _, exists := out[name]; !exists {
			out[name] = spec
		}
	}
	return out
}

func resolutionIdentity(value string) (string, string, error) {
	if marker := strings.Index(value, "@patch:"); marker > 0 {
		return value[:marker], "patch", nil
	}
	at := strings.LastIndex(value, "@")
	if at <= 0 {
		return "", "", fail(CodeOriginUnpinned, "malformed resolution", map[string]string{"resolution": value})
	}
	tail := value[at+1:]
	colon := strings.IndexByte(tail, ':')
	if colon <= 0 {
		return "", "", fail(CodeOriginUnpinned, "resolution protocol absent", nil)
	}
	protocol := tail[:colon]
	if protocol != "npm" && protocol != "workspace" && protocol != "patch" {
		return "", "", fail(CodeOriginUnpinned, "unsupported resolution protocol", map[string]string{"protocol": protocol})
	}
	if protocol == "npm" && !exactVersion(tail[colon+1:]) {
		return "", "", fail(CodeOriginUnpinned, "npm resolution is mutable", map[string]string{"resolution": value})
	}
	return value[:at], protocol, nil
}
func validYarnChecksum(value, cacheKey string) bool {
	prefix := cacheKey + "/"
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	digest := strings.TrimPrefix(value, prefix)
	if len(digest) != 64 && len(digest) != 128 {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil && strings.ToLower(digest) == digest
}
func patchPathFromResolution(value string) (string, error) {
	hash := strings.IndexByte(value, '#')
	if hash < 0 {
		return "", fail(CodeManagerPluginUndeclared, "patch locator lacks path", nil)
	}
	tail := value[hash+1:]
	if end := strings.Index(tail, "::"); end >= 0 {
		tail = tail[:end]
	}
	tail = strings.ReplaceAll(tail, "%2F", "/")
	tail = strings.TrimPrefix(tail, "./")
	if !validRelative(tail, ".yarn/patches") {
		return "", fail(CodeLocalPathEscape, "patch escapes .yarn/patches", map[string]string{"path": tail})
	}
	return tail, nil
}
func resolveDependency(name, spec string, selectors, workspaces map[string]string) string {
	if strings.HasPrefix(spec, "workspace:") {
		return workspaces[name]
	}
	return selectors[name+"@"+spec]
}

func modernPeerVersionSatisfies(version, spec string) bool {
	spec = strings.TrimPrefix(spec, "npm:")
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
	left, right := modernSemverParts(version), modernSemverParts(strings.TrimPrefix(spec, prefix))
	if left == nil || right == nil {
		return false
	}
	compare := compareModernVersion(left, right)
	switch prefix {
	case "^":
		if compare < 0 {
			return false
		}
		if right[0] != 0 {
			return left[0] == right[0]
		}
		if right[1] != 0 {
			return left[0] == 0 && left[1] == right[1]
		}
		return left[0] == 0 && left[1] == 0 && left[2] == right[2]
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
	default:
		return false
	}
}

func modernSemverParts(value string) []int {
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

func compareModernVersion(left, right []int) int {
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
func conditionsMatch(values []string, target Target) (bool, string, error) {
	for _, condition := range values {
		matched, field, err := evaluateYarnCondition(condition, map[string]string{"os": target.OS, "cpu": target.Architecture, "libc": target.Libc})
		if err != nil {
			return false, "", fail(CodeLockFormatUnsupported, "package condition is outside the pinned Yarn grammar", map[string]string{"condition": condition})
		}
		if !matched {
			return false, field + "_pruned", nil
		}
	}
	return true, "", nil
}

// evaluateYarnCondition implements the pinned Yarn 4.9.2 tinylogic grammar.
// Its terms are exact os/cpu/libc selectors; unary !, parentheses, &, |, and ^
// have the same left-to-right semantics as tinylogic 2.0.0. Manifest-generated
// conditions are a narrower subset: fields are conjoined and same-field
// alternatives are grouped. Repeated ! therefore behaves exactly like Yarn:
// even counts are positive and odd counts are negated.
func evaluateYarnCondition(expression string, target map[string]string) (bool, string, error) {
	parser := yarnConditionParser{expression: expression, target: target}
	matched, field, err := parser.parseExpression()
	if err != nil {
		return false, "condition", err
	}
	if parser.offset != len(expression) {
		return false, "condition", fmt.Errorf("unexpected Yarn condition input at byte %d", parser.offset)
	}
	return matched, field, nil
}

type yarnConditionParser struct {
	expression string
	offset     int
	target     map[string]string
}

func (p *yarnConditionParser) parseExpression() (bool, string, error) {
	left, leftField, err := p.parseTerm()
	if err != nil {
		return false, "condition", err
	}
	for {
		checkpoint := p.offset
		p.skipWhitespace()
		if p.offset >= len(p.expression) || !strings.ContainsRune("|&^", rune(p.expression[p.offset])) {
			p.offset = checkpoint
			return left, leftField, nil
		}
		operator := p.expression[p.offset]
		p.offset++
		p.skipWhitespace()
		right, rightField, parseErr := p.parseTerm()
		if parseErr != nil {
			return false, "condition", parseErr
		}
		left, leftField = combineYarnCondition(operator, left, leftField, right, rightField)
	}
}

func (p *yarnConditionParser) parseTerm() (bool, string, error) {
	if p.offset >= len(p.expression) {
		return false, "condition", fmt.Errorf("yarn condition term is missing")
	}
	if p.expression[p.offset] == '!' {
		p.offset++
		matched, field, err := p.parseTerm()
		return !matched, field, err
	}
	if p.expression[p.offset] == '(' {
		p.offset++
		p.skipWhitespace()
		matched, field, err := p.parseExpression()
		if err != nil {
			return false, "condition", err
		}
		p.skipWhitespace()
		if p.offset >= len(p.expression) || p.expression[p.offset] != ')' {
			return false, "condition", fmt.Errorf("yarn condition group is unbalanced")
		}
		p.offset++
		return matched, field, nil
	}
	return p.parseSelector()
}

func (p *yarnConditionParser) parseSelector() (bool, string, error) {
	p.skipWhitespace()
	start := p.offset
	for p.offset < len(p.expression) && !strings.ContainsRune(" \t\n\r()!|&^", rune(p.expression[p.offset])) {
		p.offset++
	}
	selector := p.expression[start:p.offset]
	parts := strings.Split(selector, "=")
	if len(parts) != 2 || !validYarnConditionField(parts[0]) || !validYarnConditionValue(parts[1]) {
		return false, "condition", fmt.Errorf("malformed Yarn selector %q", selector)
	}
	actual := p.target[parts[0]]
	if actual == "" {
		return false, parts[0], fmt.Errorf("target lacks Yarn selector field %q", parts[0])
	}
	return actual == parts[1], parts[0], nil
}

func (p *yarnConditionParser) skipWhitespace() {
	for p.offset < len(p.expression) && strings.ContainsRune(" \t\n\r", rune(p.expression[p.offset])) {
		p.offset++
	}
}

func combineYarnCondition(operator byte, left bool, leftField string, right bool, rightField string) (bool, string) {
	field := "condition"
	if leftField == rightField {
		field = leftField
	}
	switch operator {
	case '&':
		if !left {
			return false, leftField
		}
		if !right {
			return false, rightField
		}
		return true, field
	case '|':
		return left || right, field
	default:
		return left != right, field
	}
}

func validYarnConditionField(value string) bool {
	return value == "os" || value == "cpu" || value == "libc"
}

func validYarnConditionValue(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func onlyKeys(values map[string]any, allowed ...string) bool {
	set := map[string]bool{}
	for _, v := range allowed {
		set[v] = true
	}
	for k := range values {
		if !set[k] {
			return false
		}
	}
	return true
}
func descriptorList(value string) ([]string, error) {
	parts := strings.Split(value, ", ")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
		if !strings.Contains(parts[i], "@") {
			return nil, fail(CodeLockFormatUnsupported, "malformed descriptor", map[string]string{"descriptor": value})
		}
	}
	sort.Strings(parts)
	return parts, nil
}
func requiredStringField(values map[string]any, key string) (string, error) {
	value, present := values[key]
	text, ok := value.(string)
	if !present || !ok || text == "" {
		return "", fail(CodeLockFormatUnsupported, key+" must be a non-empty string", nil)
	}
	return text, nil
}

func optionalStringField(values map[string]any, key, fallback string) (string, error) {
	if _, present := values[key]; !present {
		return fallback, nil
	}
	return requiredStringField(values, key)
}

func requiredBoolField(values map[string]any, key string) (bool, error) {
	value, present := values[key]
	boolean, ok := value.(bool)
	if !present || !ok {
		return false, fail(CodeLockFormatUnsupported, key+" must be a boolean", nil)
	}
	return boolean, nil
}

func optionalBoolField(values map[string]any, key string, fallback bool) (bool, error) {
	if _, present := values[key]; !present {
		return fallback, nil
	}
	return requiredBoolField(values, key)
}

func strictIntegerField(values map[string]any, key string) (int, error) {
	value, present := values[key]
	integer, ok := value.(int)
	if !present || !ok {
		return 0, fail(CodeLockFormatUnsupported, key+" must be an integer", nil)
	}
	return integer, nil
}

func strictArchitectureValues(values map[string]any, key string) ([]string, error) {
	value, present := values[key]
	if !present {
		return nil, nil
	}
	var result []string
	switch typed := value.(type) {
	case string:
		result = []string{typed}
	case []any:
		result = make([]string, len(typed))
		for index, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fail(CodeLockFormatUnsupported, "supportedArchitectures values must be strings", map[string]string{"field": key})
			}
			result[index] = text
		}
	default:
		return nil, fail(CodeLockFormatUnsupported, "supportedArchitectures field must be a string or string sequence", map[string]string{"field": key})
	}
	seen := map[string]bool{}
	for _, selector := range result {
		if !validArchitectureSelector(selector) || seen[selector] {
			return nil, fail(CodeLockFormatUnsupported, "supportedArchitectures contains a malformed or duplicate selector", map[string]string{"field": key})
		}
		seen[selector] = true
	}
	sort.Strings(result)
	return result, nil
}

func validArchitectureSelector(value string) bool {
	if value == "" || value == "current" {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '.' || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func firstError(values ...error) error {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
func lockString(fields map[string]any, key string, required bool) (string, error) {
	value, exists := fields[key]
	if !exists {
		if required {
			return "", fmt.Errorf("%s is required", key)
		}
		return "", nil
	}
	text, ok := value.(string)
	if !ok || text == "" {
		return "", fmt.Errorf("%s must be a non-empty string", key)
	}
	return text, nil
}

func lockStringMap(fields map[string]any, key string) (map[string]string, error) {
	value, exists := fields[key]
	if !exists {
		return map[string]string{}, nil
	}
	in, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a mapping", key)
	}
	out := make(map[string]string, len(in))
	for name, raw := range in {
		text, stringOK := raw.(string)
		if name == "" || !stringOK || text == "" {
			return nil, fmt.Errorf("%s entries must bind non-empty strings", key)
		}
		out[name] = text
	}
	return out, nil
}

func lockStringSlice(fields map[string]any, key string) ([]string, error) {
	value, exists := fields[key]
	if !exists {
		return nil, nil
	}
	var out []string
	switch typed := value.(type) {
	case string:
		out = []string{typed}
	case []any:
		out = make([]string, len(typed))
		for index, raw := range typed {
			text, ok := raw.(string)
			if !ok {
				return nil, fmt.Errorf("%s entries must be strings", key)
			}
			out[index] = text
		}
	default:
		return nil, fmt.Errorf("%s must be a string or string sequence", key)
	}
	seen := make(map[string]bool, len(out))
	for _, text := range out {
		if text == "" || seen[text] {
			return nil, fmt.Errorf("%s contains an empty or duplicate value", key)
		}
		seen[text] = true
	}
	sort.Strings(out)
	return out, nil
}

func lockMetaOptional(fields map[string]any, key string) (map[string]bool, error) {
	value, exists := fields[key]
	if !exists {
		return map[string]bool{}, nil
	}
	in, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a mapping", key)
	}
	out := make(map[string]bool, len(in))
	for name, raw := range in {
		metadata, mapOK := raw.(map[string]any)
		if name == "" || !mapOK || !onlyKeys(metadata, "optional") || len(metadata) != 1 {
			return nil, fmt.Errorf("%s entries must contain exactly optional", key)
		}
		optional, boolOK := metadata["optional"].(bool)
		if !boolOK {
			return nil, fmt.Errorf("%s optional must be a boolean", key)
		}
		out[name] = optional
	}
	return out, nil
}
func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
func validRelative(value, base string) bool {
	return value != "" && !path.IsAbs(value) && path.Clean(value) == value && !strings.HasPrefix(value, "../") && (base == "." || value == base || strings.HasPrefix(value, base+"/"))
}
func exactVersion(value string) bool {
	base := strings.SplitN(strings.SplitN(value, "+", 2)[0], "-", 2)[0]
	parts := strings.Split(base, ".")
	if len(parts) != 3 {
		return false
	}
	for _, v := range parts {
		if _, err := strconv.Atoi(v); err != nil {
			return false
		}
	}
	return true
}
func rawPresent(value json.RawMessage) bool {
	s := strings.TrimSpace(string(value))
	return s != "" && s != "null" && s != "[]" && s != "{}"
}
func cloneMap(value map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range value {
		out[k] = v
	}
	return out
}
func cloneBoolMap(value map[string]bool) map[string]bool {
	out := map[string]bool{}
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
func sortedKeys[T any](value map[string]T) []string {
	out := make([]string, 0, len(value))
	for k := range value {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func sortedStrings(value []string) []string {
	out := append([]string(nil), value...)
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
func equalBoolMap(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if observed, ok := b[k]; !ok || observed != v {
			return false
		}
	}
	return true
}
func edgeKey(e DependencyEdge) string {
	return e.From + "\x00" + e.Scope + "\x00" + e.Name + "\x00" + e.To
}
func stringsAny(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
func mapAny(values map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range values {
		out[k] = v
	}
	return out
}
func boolMapAny(values map[string]bool) map[string]any {
	out := map[string]any{}
	for k, v := range values {
		out[k] = v
	}
	return out
}
func bytesDigest(values map[string][]byte) map[string]any {
	out := map[string]any{}
	for k, v := range values {
		s := sha256.Sum256(v)
		out[k] = "sha256:" + hex.EncodeToString(s[:])
	}
	return out
}
func patchesAny(values []Patch) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = map[string]any{"path": v.Path, "sha256": v.SHA256, "locator": v.Locator}
	}
	return out
}
func targetCondition(pkg Package) *closuregraph.Condition {
	clauses := append([]string(nil), pkg.Conditions...)
	if len(clauses) == 0 {
		for _, x := range []struct {
			k string
			v []string
		}{{"os", pkg.OS}, {"cpu", pkg.CPU}, {"libc", pkg.Libc}} {
			if len(x.v) > 0 {
				selectors := make([]string, len(x.v))
				for index, value := range x.v {
					negation := ""
					if strings.HasPrefix(value, "!") {
						negation = "!"
						value = strings.TrimPrefix(value, "!")
					}
					selectors[index] = negation + x.k + "=" + value
				}
				if len(selectors) == 1 {
					clauses = append(clauses, selectors[0])
				} else {
					clauses = append(clauses, "("+strings.Join(selectors, " | ")+")")
				}
			}
		}
	}
	if len(clauses) == 0 {
		return nil
	}
	return &closuregraph.Condition{EvaluatorID: "yarn-modern-platform-v1", Expression: strings.Join(clauses, " & ")}
}

func workspacePatterns(raw json.RawMessage) ([]string, []string, error) {
	if !rawPresent(raw) {
		return nil, nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return values, nil, nil
	}
	var wrapped struct {
		Packages []string `json:"packages"`
		Nohoist  []string `json:"nohoist"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, nil, fail(CodeLockStale, "workspaces declaration is malformed", nil)
	}
	return wrapped.Packages, wrapped.Nohoist, nil
}

func validateJSON(payload []byte) error {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func rawBundlePresent(value json.RawMessage) bool { return rawPresent(value) }
