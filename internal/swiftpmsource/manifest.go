package swiftpmsource

import (
	"context"
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

func normalizeManifest(manifest Manifest) (Manifest, error) {
	for _, value := range []string{manifest.PackageName, manifest.ToolsVersion, manifest.SelectedManifest} {
		if value == "" || strings.ContainsAny(value, "\x00\r\n") {
			return Manifest{}, fail(CodeManifestReplayDrift, "manifest identity is incomplete")
		}
	}
	manifest.PackageName = strings.ToLower(manifest.PackageName)
	manifest.Dependencies = append([]ManifestDependency(nil), manifest.Dependencies...)
	sort.Slice(manifest.Dependencies, func(i, j int) bool {
		return dependencySortKey(manifest.Dependencies[i]) < dependencySortKey(manifest.Dependencies[j])
	})
	for index, dependency := range manifest.Dependencies {
		dependency.Identity = strings.ToLower(dependency.Identity)
		if dependency.Identity == "" || (dependency.Kind != SourceRemote && dependency.Kind != SourceLocal && dependency.Kind != SourcePath) {
			return Manifest{}, fail(CodeDependencyOriginUnsupported, "manifest dependency is incomplete")
		}
		if dependency.Kind == SourcePath {
			if dependency.LocalPath == "" || path.IsAbs(dependency.LocalPath) || path.Clean(dependency.LocalPath) != dependency.LocalPath || dependency.LocalPath == ".." || strings.HasPrefix(dependency.LocalPath, "../") {
				return Manifest{}, failFields(CodeLocalDependencyOutside, map[string]string{"identity": dependency.Identity}, "local package escapes its admitted package root")
			}
		} else {
			canonical, err := canonicalLocation(dependency.Kind, dependency.Location)
			if err != nil {
				return Manifest{}, err
			}
			dependency.Location = canonical
		}
		manifest.Dependencies[index] = dependency
		if index > 0 && dependencySortKey(manifest.Dependencies[index-1]) == dependencySortKey(dependency) {
			return Manifest{}, failFields(CodeGraphReferenceInvalid, map[string]string{"identity": dependency.Identity}, "duplicate package dependency declaration")
		}
	}
	manifest.Products = append([]Product(nil), manifest.Products...)
	sort.Slice(manifest.Products, func(i, j int) bool { return manifest.Products[i].Name < manifest.Products[j].Name })
	for index := range manifest.Products {
		product := &manifest.Products[index]
		if product.Name == "" || product.Type == "" {
			return Manifest{}, fail(CodeManifestReplayDrift, "product declaration is incomplete")
		}
		product.Targets = sortedUnique(product.Targets)
		if len(product.Targets) == 0 {
			return Manifest{}, fail(CodeManifestReplayDrift, "product %s has no targets", product.Name)
		}
		if index > 0 && manifest.Products[index-1].Name == product.Name {
			return Manifest{}, fail(CodeGraphReferenceInvalid, "duplicate product %s", product.Name)
		}
	}
	manifest.Targets = append([]Target(nil), manifest.Targets...)
	sort.Slice(manifest.Targets, func(i, j int) bool { return manifest.Targets[i].Name < manifest.Targets[j].Name })
	for index := range manifest.Targets {
		target := &manifest.Targets[index]
		if target.Name == "" || target.Type == "" {
			return Manifest{}, fail(CodeManifestReplayDrift, "target declaration is incomplete")
		}
		if target.Type == "binary" {
			return Manifest{}, failFields(CodeBinaryUnavailable, map[string]string{"target": target.Name}, "binaryTarget is unavailable even when dormant")
		}
		target.Sources = sortedUnique(target.Sources)
		for _, source := range target.Sources {
			if path.IsAbs(source) || path.Clean(source) != source || source == ".." || strings.HasPrefix(source, "../") {
				return Manifest{}, fail(CodeSourceInventoryDrift, "target source escapes package: %s", source)
			}
		}
		sort.Slice(target.Dependencies, func(i, j int) bool {
			return targetDependencyKey(target.Dependencies[i]) < targetDependencyKey(target.Dependencies[j])
		})
		sort.Slice(target.Settings, func(i, j int) bool { return settingKey(target.Settings[i]) < settingKey(target.Settings[j]) })
		if index > 0 && manifest.Targets[index-1].Name == target.Name {
			return Manifest{}, fail(CodeGraphReferenceInvalid, "duplicate target %s", target.Name)
		}
	}
	manifest.Traits = sortedUnique(manifest.Traits)
	return manifest, nil
}

func manifestDigest(manifest Manifest) (closuregraph.ID, error) {
	dependencies := make([]any, len(manifest.Dependencies))
	for i, d := range manifest.Dependencies {
		dependencies[i] = map[string]any{"identity": d.Identity, "kind": string(d.Kind), "local_path": d.LocalPath, "location": d.Location, "requirement": d.Requirement}
	}
	products := make([]any, len(manifest.Products))
	for i, p := range manifest.Products {
		products[i] = map[string]any{"name": p.Name, "targets": stringsAny(p.Targets), "type": p.Type}
	}
	targets := make([]any, len(manifest.Targets))
	for i, t := range manifest.Targets {
		deps := make([]any, len(t.Dependencies))
		for j, d := range t.Dependencies {
			var condition any
			if d.Condition != nil {
				condition = map[string]any{"evaluator_id": d.Condition.EvaluatorID, "expression": d.Condition.Expression}
			}
			deps[j] = map[string]any{"condition": condition, "name": d.Name, "package": d.Package, "product": d.Product}
		}
		settings := make([]any, len(t.Settings))
		for j, s := range t.Settings {
			var condition any
			if s.Condition != nil {
				condition = map[string]any{"evaluator_id": s.Condition.EvaluatorID, "expression": s.Condition.Expression}
			}
			settings[j] = map[string]any{"condition": condition, "kind": s.Kind, "unsafe": s.Unsafe, "value": s.Value}
		}
		targets[i] = map[string]any{"dependencies": deps, "name": t.Name, "path": t.Path, "public_headers_path": t.PublicHeadersPath, "settings": settings, "sources": stringsAny(t.Sources), "type": t.Type}
	}
	return closuregraph.DomainID("swiftpm-normalized-manifest-v1", map[string]any{"dependencies": dependencies, "package_name": manifest.PackageName, "products": products, "selected_manifest": manifest.SelectedManifest, "targets": targets, "tools_version": manifest.ToolsVersion, "traits": stringsAny(manifest.Traits)})
}

func reconcileManifestSources(manifest Manifest, files []string) error {
	present := map[string]bool{}
	for _, file := range files {
		present[file] = true
	}
	if !present[manifest.SelectedManifest] {
		return fail(CodeManifestReplayDrift, "selected manifest is absent from admitted package tree")
	}
	for _, target := range manifest.Targets {
		for _, source := range target.Sources {
			if !present[source] {
				return failFields(CodeSourceInventoryDrift, map[string]string{"target": target.Name, "source": source}, "SwiftPM source declaration is absent from admitted tree")
			}
		}
	}
	return nil
}

func reconcileManifestDependencies(packages []PackageEvidence, lock Lock) error {
	pins := map[string]Pin{}
	for _, pin := range lock.Pins {
		pins[pin.Identity] = pin
	}
	discovered := map[string]bool{}
	for _, pkg := range packages {
		if err := reconcileOneManifest(pkg, pins); err != nil {
			return err
		}
		for _, dependency := range pkg.Manifest.Dependencies {
			if dependency.Kind != SourcePath {
				discovered[dependency.Identity] = true
			}
		}
	}
	for _, pin := range lock.Pins {
		if !discovered[pin.Identity] {
			return failFields(CodeResolvedFileOutOfDate, map[string]string{"identity": pin.Identity}, "root lock contains a dangling pin")
		}
	}
	return nil
}

func reconcileOneManifest(pkg PackageEvidence, pins map[string]Pin) error {
	for _, dependency := range pkg.Manifest.Dependencies {
		if dependency.Kind == SourcePath {
			continue
		}
		pin, ok := pins[dependency.Identity]
		if !ok {
			return failFields(CodeResolvedFileOutOfDate, map[string]string{"package": pkg.Identity, "dependency": dependency.Identity}, "manifest dependency is absent from root lock")
		}
		if pin.Kind != dependency.Kind || pin.CanonicalLocation != dependency.Location {
			return failFields(CodeDependencyPinMismatch, map[string]string{"identity": dependency.Identity}, "manifest and lock source-control origin differ")
		}
		if !requirementMatchesPin(dependency.Requirement, pin) {
			return failFields(CodeResolvedFileOutOfDate, map[string]string{"identity": dependency.Identity, "requirement": dependency.Requirement}, "frozen pin does not satisfy the manifest requirement")
		}
	}
	return nil
}

func requirementMatchesPin(requirement string, pin Pin) bool {
	prefix, value, ok := strings.Cut(requirement, ":")
	if !ok || value == "" {
		return false
	}
	switch prefix {
	case "exact":
		return pin.Version == value
	case "revision":
		return strings.EqualFold(pin.Revision, value)
	case "branch":
		return pin.Branch == value
	case "range":
		lower, upper, found := strings.Cut(value, "..<")
		if !found || pin.Version == "" {
			return false
		}
		actualVersion, actualOK := parseVersion(pin.Version)
		lowerVersion, lowerOK := parseVersion(lower)
		upperVersion, upperOK := parseVersion(upper)
		return actualOK && lowerOK && upperOK && compareVersion(actualVersion, lowerVersion) >= 0 && compareVersion(actualVersion, upperVersion) < 0
	default:
		return false
	}
}

func validateSelectedReachability(packages []PackageEvidence, rootProduct string, markers map[string]string) ([]string, error) {
	byPackage := map[string]Manifest{}
	for _, pkg := range packages {
		byPackage[pkg.Identity] = pkg.Manifest
	}
	root := packages[0].Manifest
	product, ok := findProduct(root, rootProduct)
	if !ok || product.Type != "executable" {
		return nil, fail(CodeGraphIncomplete, "selected product %s is not an executable", rootProduct)
	}
	type item struct{ Package, Target string }
	queue := []item{}
	for _, target := range product.Targets {
		queue = append(queue, item{packages[0].Identity, target})
	}
	seen := map[string]bool{}
	result := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		key := current.Package + ":" + current.Target
		if seen[key] {
			continue
		}
		seen[key] = true
		manifest, exists := byPackage[current.Package]
		if !exists {
			return nil, fail(CodeGraphIncomplete, "selected package %s is absent", current.Package)
		}
		target, exists := findTarget(manifest, current.Target)
		if !exists {
			return nil, fail(CodeGraphIncomplete, "selected target %s is absent", key)
		}
		switch target.Type {
		case "plugin":
			return nil, failFields(CodePluginUnsupported, map[string]string{"target": key}, "selected graph reaches a plugin")
		case "macro":
			return nil, failFields(CodeMacroUnsupported, map[string]string{"target": key}, "selected graph reaches a macro")
		}
		for _, setting := range target.Settings {
			if setting.Unsafe && conditionSelected(setting.Condition, markers) {
				return nil, failFields(CodeUnsafeSettingForbidden, map[string]string{"target": key, "setting": setting.Kind}, "selected target uses unsafe flags")
			}
		}
		result = append(result, key)
		for _, dependency := range target.Dependencies {
			if !conditionSelected(dependency.Condition, markers) {
				continue
			}
			packageID := current.Package
			if dependency.Package != "" {
				packageID = strings.ToLower(dependency.Package)
			}
			next := dependency.Name
			if dependency.Product != "" {
				depManifest, found := byPackage[packageID]
				if !found {
					return nil, fail(CodeGraphIncomplete, "product dependency package %s is absent", packageID)
				}
				depProduct, found := findProduct(depManifest, dependency.Product)
				if !found {
					return nil, fail(CodeGraphIncomplete, "product dependency %s:%s is absent", packageID, dependency.Product)
				}
				for _, name := range depProduct.Targets {
					queue = append(queue, item{packageID, name})
				}
				continue
			}
			queue = append(queue, item{packageID, next})
		}
	}
	sort.Strings(result)
	return result, nil
}

func validateRootSelection(manifest Manifest, productName string, markers map[string]string) error {
	product, ok := findProduct(manifest, productName)
	if !ok || product.Type != "executable" {
		return fail(CodeGraphIncomplete, "selected product %s is not an executable", productName)
	}
	queue := append([]string(nil), product.Targets...)
	seen := map[string]bool{}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if seen[name] {
			continue
		}
		seen[name] = true
		target, exists := findTarget(manifest, name)
		if !exists {
			return fail(CodeGraphIncomplete, "selected root target %s is absent", name)
		}
		switch target.Type {
		case "plugin":
			return failFields(CodePluginUnsupported, map[string]string{"target": name}, "selected graph reaches a plugin")
		case "macro":
			return failFields(CodeMacroUnsupported, map[string]string{"target": name}, "selected graph reaches a macro")
		}
		for _, setting := range target.Settings {
			if setting.Unsafe && conditionSelected(setting.Condition, markers) {
				return failFields(CodeUnsafeSettingForbidden, map[string]string{"target": name, "setting": setting.Kind}, "selected target uses unsafe flags")
			}
		}
		for _, dependency := range target.Dependencies {
			if dependency.Package == "" && conditionSelected(dependency.Condition, markers) {
				if dependency.Product != "" {
					if localProduct, found := findProduct(manifest, dependency.Product); found {
						queue = append(queue, localProduct.Targets...)
					}
				} else {
					queue = append(queue, dependency.Name)
				}
			}
		}
	}
	return nil
}

func validateToolchain(toolchain Toolchain) error {
	for _, tool := range []ToolIdentity{toolchain.Swift, toolchain.SwiftPM, toolchain.PackageDescription, toolchain.Git} {
		if tool.Role == "" || tool.ExecutableRelativePath == "" || tool.VersionOutput == "" || !tool.Fingerprint.Valid() || !tool.ExecutableSHA256.Valid() {
			return fail(CodeTargetPlatformUnsupported, "Swift, SwiftPM, PackageDescription, and Git require exact C0 identities")
		}
	}
	if toolchain.Recheck == nil {
		return fail(CodeDerivationUnauthorized, "toolchain recheck is absent")
	}
	return nil
}
func validateDestination(destination Destination) error {
	if destination.Platform.OS == "" || destination.Platform.Architecture == "" || destination.Platform.TargetTriple == "" || destination.Platform.SDKID == "" || destination.Markers == nil {
		return fail(CodeTargetPlatformUnsupported, "destination is incomplete")
	}
	return nil
}
func recheckTool(ctx context.Context, toolchain Toolchain, expected ToolIdentity) error {
	observed, err := toolchain.Recheck(ctx, expected)
	if err != nil || observed.Fingerprint != expected.Fingerprint || observed.ExecutableSHA256 != expected.ExecutableSHA256 {
		return failFields(CodeToolchainChanged, map[string]string{"role": expected.Role}, "toolchain changed immediately before use")
	}
	return nil
}
func findProduct(manifest Manifest, name string) (Product, bool) {
	for _, value := range manifest.Products {
		if value.Name == name {
			return value, true
		}
	}
	return Product{}, false
}
func findTarget(manifest Manifest, name string) (Target, bool) {
	for _, value := range manifest.Targets {
		if value.Name == name {
			return value, true
		}
	}
	return Target{}, false
}
func conditionSelected(condition *closuregraph.Condition, markers map[string]string) bool {
	if condition == nil {
		return true
	}
	selected, err := evaluateCondition(condition.Expression, markers)
	return err == nil && selected
}
func sortedUnique(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	result := out[:0]
	for _, v := range out {
		if v != "" && (len(result) == 0 || result[len(result)-1] != v) {
			result = append(result, v)
		}
	}
	return result
}
func dependencySortKey(value ManifestDependency) string {
	return value.Identity + "\x00" + string(value.Kind) + "\x00" + value.Location + "\x00" + value.LocalPath + "\x00" + value.Requirement
}
func targetDependencyKey(value TargetDependency) string {
	condition := ""
	if value.Condition != nil {
		condition = value.Condition.EvaluatorID + "\x00" + value.Condition.Expression
	}
	return value.Package + "\x00" + value.Product + "\x00" + value.Name + "\x00" + condition
}
func settingKey(value BuildSetting) string {
	condition := ""
	if value.Condition != nil {
		condition = value.Condition.EvaluatorID + "\x00" + value.Condition.Expression
	}
	return value.Kind + "\x00" + value.Value + "\x00" + condition
}
func packageInventoryIDs(packages []PackageEvidence) []closuregraph.ID {
	values := make([]closuregraph.ID, len(packages))
	for i, pkg := range packages {
		values[i] = pkg.SourceInventoryDigest
	}
	return sortedIDs(values)
}

type swiftPMConditionEvaluator struct{ markers map[string]string }

func (swiftPMConditionEvaluator) ID() string { return ConditionEvaluatorID }
func (e swiftPMConditionEvaluator) Evaluate(condition closuregraph.Condition, _ closuregraph.EvaluationInput) (bool, error) {
	if condition.EvaluatorID != ConditionEvaluatorID {
		return false, fail(CodeGraphReferenceInvalid, "wrong SwiftPM condition evaluator")
	}
	return evaluateCondition(condition.Expression, e.markers)
}
func evaluateCondition(expression string, markers map[string]string) (bool, error) {
	parts := strings.Split(expression, "=")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false, fail(CodeGraphReferenceInvalid, "unsupported SwiftPM condition %q", expression)
	}
	key, value := parts[0], parts[1]
	switch key {
	case "platform", "configuration", "architecture":
		return strings.EqualFold(markers[key], value), nil
	case "trait":
		return markers["trait:"+value] == "true", nil
	default:
		return false, fail(CodeGraphReferenceInvalid, "unsupported SwiftPM condition key %q", key)
	}
}

var _ = closureexec.ToolchainIdentity{}
