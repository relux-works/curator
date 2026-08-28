package swiftpminterop

import (
	"context"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// validateConfig proves that every external identity the interop stage will
// trust is exact before a single admitted byte is read.
func validateConfig(config Config, destination swiftpmsource.Destination) error {
	if err := validateTool(config.Clang); err != nil {
		return err
	}
	if config.ClangCXX.Role != "" {
		if err := validateTool(config.ClangCXX); err != nil {
			return err
		}
	}
	if config.Recheck == nil {
		return fail(CodeDerivationUnauthorized, "interop toolchain recheck is absent")
	}
	if config.Assurance != closureexec.AssurancePortable && config.Assurance != closureexec.AssuranceVerified {
		return fail(CodeDerivationUnauthorized, "interop assurance mode is unsupported")
	}
	if err := validateComponent(config.SDK); err != nil {
		return err
	}
	if config.Sysroot != nil {
		if err := validateComponent(*config.Sysroot); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, library := range config.SystemLibraries {
		if library.Package == "" || library.Target == "" || library.ModuleMapPath == "" {
			return fail(CodeToolchainUntrusted, "system-library binding is incomplete")
		}
		key := library.Package + ":" + library.Target
		if seen[key] {
			return failFields(CodeGraphReferenceInvalid, map[string]string{"target": key}, "duplicate system-library binding")
		}
		seen[key] = true
		if err := validateComponent(library.Component); err != nil {
			return err
		}
	}
	if config.Profile.ID == "" || len(config.Profile.TargetTriples) == 0 {
		return fail(CodeTargetPlatformUnsupported, "no accepted destination profile is selected")
	}
	if !containsString(config.Profile.TargetTriples, destination.Platform.TargetTriple) {
		return failFields(CodeTargetPlatformUnsupported, map[string]string{"profile": config.Profile.ID, "triple": destination.Platform.TargetTriple}, "destination triple has no accepted interop profile")
	}
	if config.SDK.SDKFactsDigest.Valid() && destination.Platform.SDKID == "" {
		return fail(CodeTargetPlatformUnsupported, "destination omits the exact SDK identity bound by the selected SDK component")
	}
	return nil
}

func validateTool(tool swiftpmsource.ToolIdentity) error {
	if tool.Role == "" || tool.ExecutableRelativePath == "" || tool.VersionOutput == "" || !tool.Fingerprint.Valid() || !tool.ExecutableSHA256.Valid() {
		return failFields(CodeTargetPlatformUnsupported, map[string]string{"role": tool.Role}, "C-family driver requires an exact C0 identity")
	}
	return nil
}

func validateComponent(component ExternalComponent) error {
	if component.Role == "" || component.ExecutableRelativePath == "" || component.PolicySelector == "" || component.VersionOutput == "" || component.PlatformABI == "" {
		return failFields(CodeToolchainUntrusted, map[string]string{"role": component.Role}, "selected external component identity is incomplete")
	}
	if !component.Fingerprint.Valid() {
		return failFields(CodeToolchainUntrusted, map[string]string{"role": component.Role}, "selected external component has no content fingerprint")
	}
	if len(component.Roots) == 0 {
		return failFields(CodeToolchainUntrusted, map[string]string{"role": component.Role}, "selected external component declares no contained root")
	}
	return nil
}

// componentFromTool lifts an exact C0 tool identity into the shared external
// component shape used for every selected binding node.
func componentFromTool(tool swiftpmsource.ToolIdentity) ExternalComponent {
	return ExternalComponent{Role: tool.Role, ExecutableRelativePath: tool.ExecutableRelativePath, PlatformABI: tool.PlatformABI, PolicySelector: tool.PolicySelector, VersionOutput: tool.VersionOutput, Fingerprint: tool.Fingerprint, ExecutableSHA256: tool.ExecutableSHA256}
}

// componentNode builds the exact binding node for one selected component.
func componentNode(component ExternalComponent) (closuregraph.Node, closuregraph.ID, error) {
	fingerprints := []closuregraph.ID{}
	if component.ExecutableSHA256.Valid() {
		fingerprints = append(fingerprints, component.ExecutableSHA256)
	}
	sort.Slice(fingerprints, func(i, j int) bool { return fingerprints[i] < fingerprints[j] })
	payload := closuregraph.ToolchainComponentPayload{
		ComponentRole: component.Role, ContentFingerprint: component.Fingerprint,
		ExecutableRelativePath: component.ExecutableRelativePath, PlatformABI: component.PlatformABI,
		PolicySelector: component.PolicySelector, VersionOutput: component.VersionOutput,
		SDKFactsDigest: component.SDKFactsDigest, TimeOfUseRecheckRule: "immediate-exact-v1",
		ExecutionDomain: closuregraph.ExecutionTarget, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget},
	}
	if len(fingerprints) != 0 {
		payload.LinkFingerprintIDs = fingerprints
	}
	node := closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "swiftpm.interop.component." + component.Role, Payload: payload}
	if err := node.Validate(); err != nil {
		return closuregraph.Node{}, "", failFields(CodeToolchainUntrusted, map[string]string{"role": component.Role}, "selected external component is not a valid binding node: %v", err)
	}
	id, err := node.ID()
	return node, id, err
}

// requireLanguageProfile gates the restricted language families. A destination
// without an accepted fixture has no support, not a best-effort attempt.
func requireLanguageProfile(profile PlatformProfile, destination swiftpmsource.Destination, pkg, target string, languages []Language) error {
	key := pkg + ":" + target
	if containsLanguage(languages, LanguageObjC) || containsLanguage(languages, LanguageObjCXX) {
		if profile.ObjectiveCRuntime == "" {
			return failFields(CodeTargetPlatformUnsupported, map[string]string{"target": key, "triple": destination.Platform.TargetTriple, "profile": profile.ID}, "Objective-C family has no accepted runtime profile for this destination")
		}
	}
	if containsLanguage(languages, LanguageCXX) || containsLanguage(languages, LanguageObjCXX) {
		if len(profile.CxxStandardModes) == 0 {
			return failFields(CodeTargetPlatformUnsupported, map[string]string{"target": key, "profile": profile.ID}, "C++ family has no accepted standard/toolchain profile for this destination")
		}
		if profile.CxxStandardModes[0] == "" {
			return failFields(CodeTargetPlatformUnsupported, map[string]string{"target": key, "profile": profile.ID}, "C++ standard profile is empty")
		}
	}
	return nil
}

// recheck proves the exact selected identity immediately before use.
func recheck(ctx context.Context, config Config, tool swiftpmsource.ToolIdentity) error {
	observed, err := config.Recheck(ctx, tool)
	if err != nil || observed.Fingerprint != tool.Fingerprint || observed.ExecutableSHA256 != tool.ExecutableSHA256 {
		return failFields(CodeToolchainChanged, map[string]string{"role": tool.Role}, "C-family toolchain changed immediately before use")
	}
	return nil
}

// cxxInteropDeclared reports whether a Swift target declares direct C++
// interoperation at all, with every condition ignored. The boundary mode is a
// capture fact, so the declaration-level gate must read a condition-neutral
// input: evaluating the setting's condition here published two different
// capture records for one admitted closure and hard-rejected a destination on
// which the entire C++ declaration is pruned.
func cxxInteropDeclared(target swiftpmsource.Target) bool {
	for _, setting := range target.Settings {
		if cxxInteropSetting(setting) {
			return true
		}
	}
	return false
}

// cxxInteropSelected reports whether the exact destination keeps a Swift
// target's C++ interoperation opt-in. SwiftPM never propagates the mode
// implicitly, and a conditional opt-in is a destination fact.
func cxxInteropSelected(target swiftpmsource.Target, markers map[string]string) bool {
	for _, setting := range target.Settings {
		if cxxInteropSetting(setting) && conditionSelected(setting.Condition, markers) {
			return true
		}
	}
	return false
}

// cxxInteropSetting reads the decoded kind rather than the folded record kind,
// because `swiftpmsource.decodeBuildSetting` labels every setting decoded from
// a real manifest `swiftpm-setting` and keeps the declared kind only inside the
// retained JSON. Reading the folded kind meant this gate could see an
// `.interoperabilityMode(.Cxx)` declared by a directly constructed record but
// never one SwiftPM actually emitted.
func cxxInteropSetting(setting swiftpmsource.BuildSetting) bool {
	decoded, err := decodeSettingKind(setting)
	if err != nil || !strings.EqualFold(decoded.kind, "interoperabilityMode") {
		return false
	}
	for _, value := range decoded.values {
		if strings.EqualFold(value, "Cxx") || strings.EqualFold(value, "C++") {
			return true
		}
	}
	return false
}

func conditionSelected(condition *closuregraph.Condition, markers map[string]string) bool {
	if condition == nil {
		return true
	}
	parts := strings.Split(condition.Expression, "=")
	if len(parts) != 2 {
		return false
	}
	switch parts[0] {
	case "platform", "configuration", "architecture":
		return strings.EqualFold(markers[parts[0]], parts[1])
	case "trait":
		return markers["trait:"+parts[1]] == "true"
	default:
		return false
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
