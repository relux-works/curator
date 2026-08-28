// Package swiftpminterop implements the swiftpm-source-v1 C-family target,
// module-map, header, system-library, and interop-boundary validation stage.
// It consumes the accepted selection-neutral source closure produced by
// internal/swiftpmsource, proves that every header, module, framework,
// library, and SDK read resolves either to admitted source or to exactly one
// C0/C4-selected binding node, and republishes the closure as an extended
// capture graph plus an exact selection binding.
//
// The stage never compiles, links, or otherwise starts a process on its own:
// observed compiler read sets arrive through the manager-owned ReadSetProvider
// seam, and portable assurance is reported honestly as not-observed.
package swiftpminterop

import (
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/swiftpmsource"
)

// Code is a stable diagnostic code. The interop stage deliberately shares one
// vocabulary with the SwiftPM source adapter so a shared cause is never
// renamed at a phase boundary.
type Code = swiftpmsource.Code

// Stable diagnostics owned or re-raised by the interop boundary.
const (
	// CodeMixedLanguageTarget rejects Swift and C-family source in one target.
	CodeMixedLanguageTarget Code = "swiftpm_mixed_language_target_unsupported"
	// CodeModuleMapEscape rejects an absolute, escaping, malformed, or
	// otherwise undeclared module-map reference.
	CodeModuleMapEscape Code = "swiftpm_modulemap_escape"
	// CodeHeaderInputUndeclared rejects a header or module read outside the
	// admitted closure and the selected toolchain/SDK roots.
	CodeHeaderInputUndeclared Code = "swiftpm_header_input_undeclared"
	// CodeTargetPlatformUnsupported rejects a destination/toolchain/language
	// combination that has no accepted profile.
	CodeTargetPlatformUnsupported Code = swiftpmsource.CodeTargetPlatformUnsupported
	// CodeUnsafeSettingForbidden rejects unsafe build settings.
	CodeUnsafeSettingForbidden Code = swiftpmsource.CodeUnsafeSettingForbidden
	// CodePluginUnsupported rejects reaching a build-tool or command plugin.
	CodePluginUnsupported Code = swiftpmsource.CodePluginUnsupported
	// CodeMacroUnsupported rejects reaching a macro implementation.
	CodeMacroUnsupported Code = swiftpmsource.CodeMacroUnsupported
	// CodeBinaryUnavailable rejects every binary target.
	CodeBinaryUnavailable Code = swiftpmsource.CodeBinaryUnavailable
	// CodeToolchainUntrusted rejects a system library, SDK, or host path
	// without external-toolchain evidence.
	CodeToolchainUntrusted Code = "artifact_toolchain_untrusted"
	// CodeWriteUndeclared rejects a write the declared closure never authorized.
	CodeWriteUndeclared Code = "closure_write_undeclared"
	// CodeToolchainChanged rejects toolchain drift immediately before use.
	CodeToolchainChanged Code = swiftpmsource.CodeToolchainChanged
	// CodeGeneratedInputUndeclared rejects a generated source, header, or
	// resource without accepted generator lineage.
	CodeGeneratedInputUndeclared Code = "artifact_generated_input_undeclared"
	// CodeInteropUndeclared rejects an undeclared or incompatible interop edge.
	CodeInteropUndeclared Code = "closure_interop_undeclared"
	// CodeTargetIdentityChanged rejects target-platform drift after binding.
	CodeTargetIdentityChanged Code = "closure_target_identity_changed"
	// CodeDerivationUnauthorized rejects incomplete interop authority.
	CodeDerivationUnauthorized Code = swiftpmsource.CodeDerivationUnauthorized
	// CodeGraphIncomplete rejects an incomplete interop closure.
	CodeGraphIncomplete Code = swiftpmsource.CodeGraphIncomplete
	// CodeGraphReferenceInvalid rejects duplicate, dangling, wrong-kind, or
	// capture-replacing binding records.
	CodeGraphReferenceInvalid Code = swiftpmsource.CodeGraphReferenceInvalid
	// CodeSourceInventoryDrift rejects an inventory that differs from the
	// admitted tree.
	CodeSourceInventoryDrift Code = swiftpmsource.CodeSourceInventoryDrift
)

// ErrorCode extracts the stable diagnostic carried by err. Shared graph,
// checkpoint, and execution boundaries render their stable code as the leading
// token of a plain error rather than as a typed adapter failure, so both
// encodings resolve to one vocabulary here.
func ErrorCode(err error) Code {
	if code := swiftpmsource.ErrorCode(err); code != "" {
		return code
	}
	if err == nil {
		return ""
	}
	token := err.Error()
	if index := strings.IndexAny(token, " :"); index > 0 {
		token = token[:index]
	}
	if sharedCodes[Code(token)] {
		return Code(token)
	}
	return ""
}

// sharedCodes are the stable codes the shared closure services raise as plain
// errors that this boundary may surface unchanged.
var sharedCodes = map[Code]bool{
	CodeInteropUndeclared: true, CodeGraphIncomplete: true, CodeGraphReferenceInvalid: true,
	CodeGeneratedInputUndeclared: true, CodeTargetIdentityChanged: true, CodeDerivationUnauthorized: true,
	CodeWriteUndeclared:                true,
	"closure_graph_schema_unsupported": true, "closure_checkpoint_invalid": true,
	"closure_build_cycle": true, "closure_derivation_drift": true,
}

func fail(code Code, format string, args ...any) error {
	return &swiftpmsource.Failure{Code: code, Detail: fmt.Sprintf(format, args...), Fields: map[string]string{}}
}

func failFields(code Code, fields map[string]string, format string, args ...any) error {
	return &swiftpmsource.Failure{Code: code, Detail: fmt.Sprintf(format, args...), Fields: fields}
}
