package swiftpminterop

import (
	"context"
	"path"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

// deriveBoundaries turns every selected consumer-to-C-family target edge into
// an explicit typed interop boundary. A boundary is never implied by a name,
// a suffix, or a successful compile: it must be derivable from the accepted
// target graph and the exact selected destination profile.
func (state *closeState) deriveBoundaries() ([]Boundary, error) {
	boundaries := []Boundary{}
	seen := map[string]bool{}
	for consumerIndex := range state.targets {
		consumer := &state.targets[consumerIndex]
		if consumer.Kind == KindSystem {
			continue
		}
		for _, edge := range state.directTargetDependencies(consumer) {
			provider := &state.targets[edge.index]
			if provider.Kind == KindSwift {
				continue
			}
			boundary, err := state.boundaryFor(consumer, provider, edge.condition)
			if err != nil {
				return nil, err
			}
			key := boundary.Provider + "->" + boundary.Consumer
			if seen[key] {
				return nil, failFields(CodeGraphReferenceInvalid, map[string]string{"boundary": key}, "duplicate interop boundary declaration")
			}
			seen[key] = true
			boundaries = append(boundaries, boundary)
		}
	}
	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i].Provider+"->"+boundaries[i].Consumer < boundaries[j].Provider+"->"+boundaries[j].Consumer
	})
	return boundaries, nil
}

func (state *closeState) boundaryFor(consumer, provider *TargetInterop, condition *closuregraph.Condition) (Boundary, error) {
	consumerKey := consumer.Package + ":" + consumer.Target
	providerKey := provider.Package + ":" + provider.Target
	boundary := Boundary{
		Provider: providerKey, Consumer: consumerKey,
		ProviderLanguages: provider.Languages, ConsumerLanguages: consumer.Languages,
		LinkLoadSemantics: "static-link", ToolchainRole: state.config.Clang.Role,
		Condition: condition,
		Selected:  consumer.Selected && provider.Selected && conditionSelected(condition, state.markers),
	}
	if provider.Kind != KindSystem && provider.ModuleMap == nil {
		return Boundary{}, failFields(CodeInteropUndeclared, map[string]string{"provider": providerKey}, "C-family provider has no admitted module map evidence")
	}
	if provider.Kind == KindSystem {
		boundary.Mode = closuregraph.InteropCABI
		boundary.ABI = CABIContract
		boundary.CallingConvention = "c"
		boundary.InterfaceContract = ModuleMapGrammarID
		boundary.ProviderLanguages = []Language{LanguageC}
		return state.finishBoundary(boundary)
	}
	implementationCxx := containsLanguage(provider.Languages, LanguageCXX) || containsLanguage(provider.Languages, LanguageObjCXX)
	interfaceCxx := interfaceExposesCxx(provider)
	if implementationCxx {
		if state.config.ClangCXX.Role == "" {
			return Boundary{}, failFields(CodeToolchainUntrusted, map[string]string{"provider": providerKey}, "C++ family provider has no selected C++ driver identity")
		}
		boundary.ToolchainRole = state.config.ClangCXX.Role
	}
	switch {
	case consumer.Kind == KindSwift && consumer.CxxInteropDeclared && (implementationCxx || interfaceCxx):
		if boundary.Selected && !state.config.Profile.CxxInterop {
			return Boundary{}, failFields(CodeTargetPlatformUnsupported, map[string]string{"provider": providerKey, "consumer": consumerKey, "profile": state.config.Profile.ID}, "direct Swift/C++ interoperation has no accepted destination profile")
		}
		boundary.Mode = closuregraph.InteropCXX
		boundary.ABI = CXXABIContract
		boundary.Runtime = CXXRuntimeContract
		boundary.CallingConvention = "cxx"
		boundary.InterfaceContract = "cxx-header-v1"
	case interfaceCxx && consumer.Kind == KindSwift:
		// The gate reads the condition-neutral declaration, so a consumer that
		// declares the opt-in only for some destinations is not rejected on the
		// destinations that prune the whole C++ dependency; a C++ target the
		// destination actually selects is still gated by requireLanguageProfile.
		return Boundary{}, failFields(CodeInteropUndeclared, map[string]string{"provider": providerKey, "consumer": consumerKey}, "provider exposes a C++ interface but the Swift consumer does not declare interoperabilityMode(.Cxx); SwiftPM never propagates it implicitly")
	case containsLanguage(provider.Languages, LanguageObjC) || containsLanguage(provider.Languages, LanguageObjCXX):
		if boundary.Selected && state.config.Profile.ObjectiveCRuntime == "" {
			return Boundary{}, failFields(CodeTargetPlatformUnsupported, map[string]string{"provider": providerKey, "consumer": consumerKey, "triple": state.destination.Platform.TargetTriple}, "Objective-C runtime boundary has no accepted destination profile")
		}
		boundary.Mode = closuregraph.InteropObjCRuntime
		boundary.ABI = CABIContract
		boundary.Runtime = ObjCRuntimeContract
		boundary.CallingConvention = "objc-msgsend"
		boundary.InterfaceContract = "objective-c-header-v1"
	default:
		boundary.Mode = closuregraph.InteropCABI
		boundary.ABI = CABIContract
		boundary.CallingConvention = "c"
		boundary.InterfaceContract = ModuleMapGrammarID
	}
	return state.finishBoundary(boundary)
}

func (state *closeState) finishBoundary(boundary Boundary) (Boundary, error) {
	if _, ok := state.byRole[boundary.ToolchainRole]; !ok {
		return Boundary{}, failFields(CodeToolchainUntrusted, map[string]string{"role": boundary.ToolchainRole}, "interop boundary names no selected toolchain component")
	}
	return boundary, nil
}

// interfaceExposesCxx reports whether a provider's admitted public interface
// is a C++ header rather than a C-compatible shim. A C++-only header extension
// is a declared C++ interface: consuming it from Swift needs the explicit
// interoperability opt-in, not a convention.
func interfaceExposesCxx(provider *TargetInterop) bool {
	if provider.ModuleMap == nil {
		return false
	}
	for _, reference := range provider.ModuleMap.Parsed.References {
		if reference.Kind == ReferenceExternModule || reference.Kind == ReferenceUmbrellaDirectory {
			continue
		}
		if cxxHeaderExtension(reference.Path) {
			return true
		}
	}
	for _, header := range provider.ModuleMap.PublicHeaderFiles {
		if cxxHeaderExtension(header.Relative) {
			return true
		}
	}
	return false
}

func cxxHeaderExtension(value string) bool {
	switch strings.ToLower(path.Ext(value)) {
	case ".hh", ".hpp", ".hxx", ".h++":
		return true
	default:
		return false
	}
}

// verifyReads resolves every provider-observed compiler read against the
// admitted closure and the selected binding roots. Portable assurance records
// not-observed honestly instead of claiming coverage it does not have; a
// verified run without an observed read set fails closed.
func (state *closeState) verifyReads(ctx context.Context) (ReadSetEvidence, error) {
	evidence := ReadSetEvidence{Mode: "not-observed", ReceiptIDs: []closuregraph.ID{}, Resolutions: []Resolution{}}
	observedAll := false
	for _, interop := range state.targets {
		if interop.Selected {
			observedAll = true
		}
	}
	if state.config.ReadSets == nil {
		observedAll = false
	}
	for index := range state.targets {
		interop := &state.targets[index]
		if !interop.Selected {
			// A pruned target is captured but never compiled, so it has no
			// compiler read set to observe and cannot weaken the verdict.
			continue
		}
		if state.config.ReadSets == nil {
			continue
		}
		request := ReadSetRequest{Package: interop.Package, Target: interop.Target, Languages: languageStrings(interop.Languages), Sources: append([]string(nil), interop.Sources...), PublicHeaderRoot: interop.PublicHeaderRoot}
		if interop.ModuleMap != nil {
			request.ModuleMap = interop.ModuleMap.Relative
		}
		result, err := state.config.ReadSets.ObserveReads(ctx, request)
		if err != nil {
			return ReadSetEvidence{}, failFields(CodeHeaderInputUndeclared, map[string]string{"target": interop.Package + ":" + interop.Target}, "compiler read-set observation failed: %v", err)
		}
		if !result.Observed {
			observedAll = false
			continue
		}
		if !result.ReceiptID.Valid() {
			return ReadSetEvidence{}, failFields(CodeDerivationUnauthorized, map[string]string{"target": interop.Package + ":" + interop.Target}, "observed read set omitted its issued derivation receipt")
		}
		evidence.ReceiptIDs = append(evidence.ReceiptIDs, result.ReceiptID)
		for _, read := range result.Reads {
			resolution := state.roots.resolve(read.Path)
			if resolution.Class == ResolvedUndeclared {
				return ReadSetEvidence{}, failFields(CodeHeaderInputUndeclared, map[string]string{"target": interop.Package + ":" + interop.Target, "path": read.Path, "class": read.Class}, "compiler read resolves outside the admitted closure and every selected binding node")
			}
			evidence.Resolutions = append(evidence.Resolutions, resolution)
		}
	}
	if observedAll {
		evidence.Mode = "observed"
	}
	if state.config.Assurance == closureexec.AssuranceVerified && evidence.Mode != "observed" {
		return ReadSetEvidence{}, fail(CodeHeaderInputUndeclared, "verified assurance requires an observed compiler read set for every selected target")
	}
	sort.Slice(evidence.Resolutions, func(i, j int) bool {
		return evidence.Resolutions[i].AbsolutePath < evidence.Resolutions[j].AbsolutePath
	})
	sort.Slice(evidence.ReceiptIDs, func(i, j int) bool { return evidence.ReceiptIDs[i] < evidence.ReceiptIDs[j] })
	return evidence, nil
}

// Selection-neutral interop contract identities. The exact ABI, C++ standard
// library, and Objective-C runtime that satisfy a contract are selection facts
// and therefore live only in the binding overlay's toolchain, SDK, and
// platform nodes; capture records the contract, never the destination.
const (
	// CABIContract is the platform C ABI contract.
	CABIContract = "c-abi-v1"
	// CXXABIContract is the C++ ABI contract.
	CXXABIContract = "itanium-cxx-abi-v1"
	// CXXRuntimeContract is the C++ standard library contract.
	CXXRuntimeContract = "cxx-standard-library-v1"
	// ObjCRuntimeContract is the Objective-C runtime contract.
	ObjCRuntimeContract = "objc-runtime-v1"
)
