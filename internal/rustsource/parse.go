package rustsource

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type manifestDependency struct {
	Version         string   `toml:"version"`
	Package         string   `toml:"package"`
	Registry        string   `toml:"registry"`
	Git             string   `toml:"git"`
	Rev             string   `toml:"rev"`
	Branch          string   `toml:"branch"`
	Tag             string   `toml:"tag"`
	Path            string   `toml:"path"`
	Optional        bool     `toml:"optional"`
	DefaultFeatures *bool    `toml:"default-features"`
	Features        []string `toml:"features"`
}

func (dependency *manifestDependency) UnmarshalTOML(value any) error {
	if text, ok := value.(string); ok {
		dependency.Version, dependency.DefaultFeatures = text, boolPointer(true)
		return nil
	}
	table, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("dependency must be a version or table")
	}
	allowed := map[string]bool{"version": true, "package": true, "registry": true, "git": true, "rev": true, "branch": true, "tag": true, "path": true, "optional": true, "default-features": true, "features": true}
	for key := range table {
		if !allowed[key] {
			return fmt.Errorf("unsupported dependency field %q", key)
		}
	}
	payload, err := toml.Marshal(table)
	if err != nil {
		return err
	}
	type plain manifestDependency
	if _, err = toml.Decode(string(payload), (*plain)(dependency)); err != nil {
		return err
	}
	if dependency.DefaultFeatures == nil {
		dependency.DefaultFeatures = boolPointer(true)
	}
	return nil
}

func boolPointer(value bool) *bool { return &value }

type manifestDocument struct {
	Package *struct {
		Name, Version, Links string
		Build                any `toml:"build"`
	} `toml:"package"`
	Workspace *struct {
		Members []string
	} `toml:"workspace"`
	Dependencies      map[string]manifestDependency `toml:"dependencies"`
	DevDependencies   map[string]manifestDependency `toml:"dev-dependencies"`
	BuildDependencies map[string]manifestDependency `toml:"build-dependencies"`
	Target            map[string]struct {
		Dependencies      map[string]manifestDependency `toml:"dependencies"`
		DevDependencies   map[string]manifestDependency `toml:"dev-dependencies"`
		BuildDependencies map[string]manifestDependency `toml:"build-dependencies"`
	} `toml:"target"`
	Features map[string][]string `toml:"features"`
	Lib      *struct {
		ProcMacro bool     `toml:"proc-macro"`
		CrateType []string `toml:"crate-type"`
	} `toml:"lib"`
	Bin     []struct{ Name string }                  `toml:"bin"`
	Patch   map[string]map[string]manifestDependency `toml:"patch"`
	Replace map[string]manifestDependency            `toml:"replace"`
}

// ParseManifest parses Cargo declarations without executing Cargo. Unknown
// top-level keys are allowed only when they are non-authoritative metadata;
// every dependency table is closed and all target predicates stay unevaluated.
func ParseManifest(filename string, payload []byte) (Manifest, error) {
	if err := validateManifestShape(payload); err != nil {
		return Manifest{}, err
	}
	var document manifestDocument
	_, err := toml.Decode(string(payload), &document)
	if err != nil {
		return Manifest{}, fail(CodeGraphIncomplete, "Cargo.toml is malformed", map[string]string{"path": filename})
	}
	result := Manifest{Path: filename, Features: map[string][]string{}}
	if document.Package != nil {
		result.PackageName, result.PackageVersion, result.Links = document.Package.Name, document.Package.Version, document.Package.Links
		result.HasBuildScript = document.Package.Build != nil && document.Package.Build != false
	}
	if document.Workspace != nil {
		result.WorkspaceMembers = append([]string(nil), document.Workspace.Members...)
	}
	if document.Lib != nil {
		result.ProcMacro = document.Lib.ProcMacro
		for _, crateType := range document.Lib.CrateType {
			if crateType == "proc-macro" {
				result.ProcMacro = true
			}
		}
	}
	for _, binary := range document.Bin {
		result.Bins = append(result.Bins, binary.Name)
	}
	appendDependencies := func(kind, target string, values map[string]manifestDependency) {
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, name := range keys {
			value := values[name]
			defaults := true
			if value.DefaultFeatures != nil {
				defaults = *value.DefaultFeatures
			}
			features, _ := sortedUnique(value.Features)
			result.Dependencies = append(result.Dependencies, DependencyDeclaration{Name: name, Package: value.Package, Version: value.Version, Registry: value.Registry, Git: value.Git, Rev: value.Rev, Branch: value.Branch, Tag: value.Tag, Path: value.Path, Optional: value.Optional, DefaultFeatures: defaults, Features: features, Kind: kind, Target: target})
		}
	}
	appendDependencies("normal", "", document.Dependencies)
	appendDependencies("development", "", document.DevDependencies)
	appendDependencies("build", "", document.BuildDependencies)
	targets := make([]string, 0, len(document.Target))
	for target := range document.Target {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	for _, target := range targets {
		table := document.Target[target]
		appendDependencies("normal", target, table.Dependencies)
		appendDependencies("development", target, table.DevDependencies)
		appendDependencies("build", target, table.BuildDependencies)
	}
	patchRegistries := make([]string, 0, len(document.Patch))
	for registry := range document.Patch {
		patchRegistries = append(patchRegistries, registry)
	}
	sort.Strings(patchRegistries)
	for _, registry := range patchRegistries {
		appendDependencies("patch", "patch:"+registry, document.Patch[registry])
	}
	if len(document.Replace) != 0 {
		return Manifest{}, fail(CodeConfigUntrusted, "Cargo [replace] is unsupported", map[string]string{"path": filename})
	}
	for feature, values := range document.Features {
		sorted, unique := sortedUnique(values)
		if !unique {
			return Manifest{}, fail(CodeGraphIncomplete, "duplicate feature member", map[string]string{"feature": feature})
		}
		result.Features[feature] = sorted
	}
	result.WorkspaceMembers, _ = sortedUnique(result.WorkspaceMembers)
	sort.Strings(result.Bins)
	result.Digest = digest(payload)
	return result, nil
}

func validateManifestShape(payload []byte) error {
	var root map[string]any
	if _, err := toml.Decode(string(payload), &root); err != nil {
		return fail(CodeGraphIncomplete, "Cargo.toml is malformed", nil)
	}
	allowedTop := map[string]bool{"package": true, "workspace": true, "dependencies": true, "dev-dependencies": true, "build-dependencies": true, "target": true, "features": true, "patch": true, "replace": true, "lib": true, "bin": true, "example": true, "test": true, "bench": true, "lints": true, "badges": true, "profile": true}
	for key := range root {
		if !allowedTop[key] {
			return fail(CodeGraphIncomplete, "unsupported Cargo manifest table", map[string]string{"field": key})
		}
	}
	if table, ok := root["package"].(map[string]any); ok {
		allowed := map[string]bool{"name": true, "version": true, "authors": true, "edition": true, "rust-version": true, "description": true, "documentation": true, "readme": true, "homepage": true, "repository": true, "license": true, "license-file": true, "keywords": true, "categories": true, "workspace": true, "build": true, "links": true, "exclude": true, "include": true, "publish": true, "metadata": true, "default-run": true, "autolib": true, "autobins": true, "autoexamples": true, "autotests": true, "autobenches": true}
		for key := range table {
			if !allowed[key] {
				return fail(CodeGraphIncomplete, "unsupported Cargo package field", map[string]string{"field": "package." + key})
			}
		}
	}
	if targets, ok := root["target"].(map[string]any); ok {
		allowed := map[string]bool{"dependencies": true, "dev-dependencies": true, "build-dependencies": true}
		for predicate, raw := range targets {
			table, ok := raw.(map[string]any)
			if !ok {
				return fail(CodeGraphIncomplete, "Cargo target table is malformed", map[string]string{"target": predicate})
			}
			for key := range table {
				if !allowed[key] {
					return fail(CodeGraphIncomplete, "unsupported target declaration", map[string]string{"target": predicate, "field": key})
				}
			}
		}
	}
	if library, ok := root["lib"].(map[string]any); ok {
		allowed := map[string]bool{"name": true, "path": true, "crate-type": true, "bench": true, "doctest": true, "test": true, "doc": true, "plugin": true, "proc-macro": true, "harness": true, "edition": true, "required-features": true}
		for key := range library {
			if !allowed[key] {
				return fail(CodeGraphIncomplete, "unsupported library declaration", map[string]string{"field": "lib." + key})
			}
		}
	}
	for _, section := range []string{"bin", "example", "test", "bench"} {
		if raw, present := root[section]; present {
			items, ok := raw.([]map[string]any)
			if !ok {
				return fail(CodeGraphIncomplete, "Cargo target array is malformed", map[string]string{"field": section})
			}
			allowed := map[string]bool{"name": true, "path": true, "test": true, "bench": true, "doc": true, "doctest": true, "harness": true, "edition": true, "required-features": true, "crate-type": true}
			for _, item := range items {
				for key := range item {
					if !allowed[key] {
						return fail(CodeGraphIncomplete, "unsupported Cargo target field", map[string]string{"field": section + "." + key})
					}
				}
			}
		}
	}
	if _, exists := root["profile"]; exists {
		return fail(CodeConfigUntrusted, "package-defined Cargo profiles are unsupported", nil)
	}
	return nil
}

type lockDocument struct {
	Version int
	Package []struct {
		Name, Version, Source, Checksum string
		Dependencies                    []string
	}
}

// ParseLock parses supported Cargo.lock v3/v4 into an all-target/all-feature superset.
func ParseLock(payload []byte) (LockFile, error) {
	return ParseLockWithApprovedRegistries(payload, []string{"registry+https://github.com/rust-lang/crates.io-index"})
}

// ParseLockWithApprovedRegistries parses a lock under an explicit immutable
// registry policy. Registry locators absent from the policy fail closed.
func ParseLockWithApprovedRegistries(payload []byte, approvedRegistries []string) (LockFile, error) {
	if len(payload) == 0 {
		return LockFile{}, fail(CodeLockRequired, "Cargo.lock is required", nil)
	}
	var document lockDocument
	metadata, err := toml.Decode(string(payload), &document)
	if err != nil || (document.Version != 3 && document.Version != 4) {
		return LockFile{}, fail(CodeLockRequired, "unsupported Cargo.lock", nil)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) != 0 {
		return LockFile{}, fail(CodeLockMismatch, "unknown Cargo.lock field", map[string]string{"field": undecoded[0].String()})
	}
	approved := map[string]bool{}
	for _, source := range approvedRegistries {
		approved[source] = true
	}
	result := LockFile{Version: document.Version, Digest: digest(payload)}
	seen := map[string]bool{}
	for _, item := range document.Package {
		key := PackageKey{Name: item.Name, Version: item.Version, Source: item.Source}
		if err := validatePackageKey(key); err != nil {
			return LockFile{}, fail(CodeLockMismatch, err.Error(), nil)
		}
		if seen[key.String()] {
			return LockFile{}, fail(CodeLockMismatch, "duplicate lock package", map[string]string{"package": key.String()})
		}
		seen[key.String()] = true
		var kind SourceKind
		switch {
		case strings.HasPrefix(item.Source, "registry+"):
			kind = SourceRegistry
		case strings.HasPrefix(item.Source, "git+"):
			kind = SourceGit
		case item.Source == "":
			kind = SourcePath
		default:
			return LockFile{}, fail(CodeLockMismatch, "unsupported lock source", map[string]string{"source": item.Source})
		}
		if kind == SourceRegistry && !approved[item.Source] {
			return LockFile{}, fail(CodeRegistryIdentityInvalid, "registry source is not manager-approved", map[string]string{"source": item.Source})
		}
		if kind == SourceGit && !validGitSource(item.Source) {
			return LockFile{}, fail(CodeGitIdentityInvalid, "Git lock source lacks a full lowercase commit", map[string]string{"source": item.Source})
		}
		if kind == SourceRegistry && !validHexDigest(item.Checksum) {
			return LockFile{}, fail(CodeRegistryIdentityInvalid, "registry checksum is missing or invalid", map[string]string{"package": key.String()})
		}
		if kind != SourceRegistry && item.Checksum != "" {
			return LockFile{}, fail(CodeLockMismatch, "checksum on non-registry package", map[string]string{"package": key.String()})
		}
		dependencies := make([]LockDependency, len(item.Dependencies))
		for i, raw := range item.Dependencies {
			dependencies[i] = parseLockDependency(raw)
		}
		sort.Slice(dependencies, func(i, j int) bool {
			a, b := dependencies[i], dependencies[j]
			return a.Name+"\x00"+a.Version+"\x00"+a.Source < b.Name+"\x00"+b.Version+"\x00"+b.Source
		})
		result.Packages = append(result.Packages, LockPackage{Key: key, Kind: kind, Checksum: item.Checksum, Dependencies: dependencies})
	}
	sort.Slice(result.Packages, func(i, j int) bool { return result.Packages[i].Key.String() < result.Packages[j].Key.String() })
	return result, nil
}

func validGitSource(source string) bool {
	index := strings.LastIndexByte(source, '#')
	if index < 4 || index+41 != len(source) {
		return false
	}
	commit := source[index+1:]
	return validLowerHex(commit, 40)
}

func parseLockDependency(value string) LockDependency {
	parts := strings.Split(value, " ")
	dependency := LockDependency{Name: parts[0]}
	if len(parts) > 1 {
		dependency.Version = parts[1]
	}
	if len(parts) > 2 {
		dependency.Source = strings.Trim(parts[2], "()")
	}
	return dependency
}

func validHexDigest(value string) bool {
	return validLowerHex(value, 64)
}
func validLowerHex(value string, size int) bool {
	if len(value) != size {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && strings.ToLower(value) == value
}
func digest(payload []byte) string { sum := sha256.Sum256(payload); return hex.EncodeToString(sum[:]) }
func safeRelative(value string) bool {
	return value != "" && value != "." && !strings.HasPrefix(value, "/") && path.Clean(value) == value && value != ".." && !strings.HasPrefix(value, "../") && !strings.Contains(value, "\\")
}
