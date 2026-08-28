// Package swiftpmbuild implements the swiftpm-source-v1 C5 planning, C6
// offline build, and C7 publication boundary. It consumes the accepted
// selection-neutral source closure produced by internal/swiftpmsource and the
// accepted C-family interop closure produced by internal/swiftpminterop,
// binds the exact platform, SwiftPM, swiftc, PackageDescription, Clang,
// linker, and SDK identities, plans the native SwiftPM build, executes it from
// fresh isolated roots with network denied and resolution frozen, reconciles
// the observed command and read/write sets, and publishes sorted observations
// and receipts through the shared protected store.
//
// The stage starts no process of its own: every child crosses the shared
// commit-before-start executor seam in internal/closureexec, so this package
// never imports os/exec.
package swiftpmbuild

import (
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// Code is a stable diagnostic code. The build stage shares one vocabulary with
// the SwiftPM source and interop stages: a shared cause is never renamed at a
// phase boundary.
type Code = swiftpmsource.Code

// Stable diagnostics owned or re-raised by the SwiftPM build boundary.
const (
	// CodeBuildGraphDrift rejects planned commands, target order, host/target
	// edges, system edges, or an FFI graph that differ from the checkpoint.
	CodeBuildGraphDrift Code = swiftpmsource.CodeBuildGraphDrift
	// CodeOfflineRebuildFailed rejects a frozen local-mirror replay or build
	// that tried an unavailable input or otherwise failed.
	CodeOfflineRebuildFailed Code = swiftpmsource.CodeOfflineReplayFailed
	// CodeMirrorMissing rejects a lock pin without a captured local mirror.
	CodeMirrorMissing Code = swiftpmsource.CodeDependencyMirrorMissing
	// CodeResolutionUnfrozen rejects a build requested before the root lock and
	// captured recursive graph were committed.
	CodeResolutionUnfrozen Code = swiftpmsource.CodeResolutionUnfrozen
	// CodeTargetPlatformUnsupported rejects a destination/toolchain combination
	// that has no accepted profile.
	CodeTargetPlatformUnsupported Code = swiftpmsource.CodeTargetPlatformUnsupported
	// CodeUnsafeSettingForbidden rejects unsafe build settings.
	CodeUnsafeSettingForbidden Code = swiftpmsource.CodeUnsafeSettingForbidden
	// CodeHeaderInputUndeclared rejects a header or module read outside the
	// admitted closure and the selected toolchain/SDK roots.
	CodeHeaderInputUndeclared Code = swiftpminterop.CodeHeaderInputUndeclared
	// CodeToolchainUntrusted rejects a linker, SDK, or host path without
	// external-toolchain evidence.
	CodeToolchainUntrusted Code = swiftpminterop.CodeToolchainUntrusted
	// CodeToolchainChanged rejects toolchain drift immediately before use.
	CodeToolchainChanged Code = swiftpmsource.CodeToolchainChanged
	// CodeDerivationUnauthorized rejects incomplete or foreign build authority.
	CodeDerivationUnauthorized Code = swiftpmsource.CodeDerivationUnauthorized
	// CodeGraphIncomplete rejects an incomplete build closure.
	CodeGraphIncomplete Code = swiftpmsource.CodeGraphIncomplete
	// CodeGraphReferenceInvalid rejects missing, duplicate, dangling,
	// wrong-kind, or capture-replacing binding records.
	CodeGraphReferenceInvalid Code = swiftpmsource.CodeGraphReferenceInvalid
	// CodeOutputUnreceipted rejects a pre-existing or unreceipted local output.
	CodeOutputUnreceipted Code = "artifact_local_output_unreceipted"
	// CodeOutputDrift rejects an output whose observed bytes differ from the
	// exact enforcement receipt.
	CodeOutputDrift Code = "artifact_local_output_drift"
	// CodeInputUndeclared rejects a read the committed permit never declared.
	CodeInputUndeclared Code = "closure_input_undeclared"
	// CodeWriteUndeclared rejects a write the committed permit never declared.
	CodeWriteUndeclared Code = "closure_write_undeclared"
	// CodeProcessUndeclared rejects a process the committed permit never allowed.
	CodeProcessUndeclared Code = "closure_process_undeclared"
	// CodeNetworkAttempted rejects any network use during the offline build.
	CodeNetworkAttempted Code = "closure_network_attempted"
	// CodeCheckpointInvalid rejects a broken C4/C5/C6/C7 chain.
	CodeCheckpointInvalid Code = "closure_checkpoint_invalid"
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
	CodeGraphIncomplete: true, CodeGraphReferenceInvalid: true, CodeDerivationUnauthorized: true,
	CodeCheckpointInvalid: true, CodeInputUndeclared: true, CodeWriteUndeclared: true,
	CodeProcessUndeclared: true, CodeNetworkAttempted: true, CodeOutputUnreceipted: true,
	CodeOutputDrift: true, CodeToolchainChanged: true, CodeToolchainUntrusted: true,
	"closure_derivation_drift": true, "closure_build_cycle": true,
	"closure_target_identity_changed": true, "closure_graph_schema_unsupported": true,
	"artifact_generated_input_undeclared": true,
}

func fail(code Code, format string, args ...any) error {
	return &swiftpmsource.Failure{Code: code, Detail: fmt.Sprintf(format, args...), Fields: map[string]string{}}
}

func failFields(code Code, fields map[string]string, format string, args ...any) error {
	return &swiftpmsource.Failure{Code: code, Detail: fmt.Sprintf(format, args...), Fields: fields}
}
