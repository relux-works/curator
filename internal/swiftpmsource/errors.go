// Package swiftpmsource implements the restricted swiftpm-source-v1 source
// acquisition and graph-closure boundary. Compilation and C-family read-set
// validation are intentionally owned by the downstream SwiftPM build and
// interop adapters.
package swiftpmsource

import "fmt"

// Code is a stable swiftpm-source-v1 diagnostic code.
type Code string

// Stable SwiftPM diagnostics implemented by this adapter boundary.
const (
	CodeResolutionUnfrozen          Code = "swiftpm_resolution_unfrozen"
	CodeResolvedFileOutOfDate       Code = "swiftpm_resolved_file_out_of_date"
	CodeDependencyPinMismatch       Code = "swiftpm_dependency_pin_mismatch"
	CodeDependencyOriginUnsupported Code = "swiftpm_dependency_origin_unsupported"
	CodeDependencyMirrorMissing     Code = "swiftpm_dependency_mirror_missing"
	CodeLocalDependencyOutside      Code = "swiftpm_local_dependency_outside_closure"
	CodeManifestReplayDrift         Code = "swiftpm_manifest_replay_drift"
	CodeSourceInventoryDrift        Code = "swiftpm_source_inventory_drift"
	CodeUnsafeSettingForbidden      Code = "swiftpm_unsafe_build_setting_forbidden"
	CodePluginUnsupported           Code = "swiftpm_plugin_execution_unsupported"
	CodeMacroUnsupported            Code = "swiftpm_macro_execution_unsupported"
	CodeTargetPlatformUnsupported   Code = "swiftpm_target_platform_unsupported"
	CodeOfflineReplayFailed         Code = "swiftpm_offline_rebuild_failed"
	CodeBuildGraphDrift             Code = "swiftpm_build_graph_drift"
	CodeBinaryUnavailable           Code = "artifact_binary_admission_unavailable"
	CodeToolchainChanged            Code = "artifact_toolchain_identity_changed"
	CodeDerivationUnauthorized      Code = "closure_derivation_unauthorized"
	CodeDerivationDrift             Code = "closure_derivation_drift"
	CodeGraphIncomplete             Code = "closure_graph_incomplete"
	CodeGraphReferenceInvalid       Code = "closure_graph_reference_invalid"
)

// Failure retains a stable code and deterministic structured fields.
type Failure struct {
	Code   Code
	Detail string
	Fields map[string]string
}

func (failure *Failure) Error() string {
	if failure.Detail == "" {
		return string(failure.Code)
	}
	return string(failure.Code) + ": " + failure.Detail
}

func fail(code Code, format string, args ...any) error {
	return &Failure{Code: code, Detail: fmt.Sprintf(format, args...), Fields: map[string]string{}}
}

func failFields(code Code, fields map[string]string, format string, args ...any) error {
	return &Failure{Code: code, Detail: fmt.Sprintf(format, args...), Fields: fields}
}

// ErrorCode extracts the stable adapter code carried by err.
func ErrorCode(err error) Code {
	if failure, ok := err.(*Failure); ok {
		return failure.Code
	}
	return ""
}
