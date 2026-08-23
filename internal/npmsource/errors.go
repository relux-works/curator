// Package npmsource implements the closed npm source-closure profile.
package npmsource

// ProfileID identifies the closed npm lock/materialization profile.
const ProfileID = "npm-source-v1"

// Stable shared closure diagnostics emitted by npm-source-v1.
const (
	// CodeLockMissing reports an absent authoritative npm lock.
	CodeLockMissing                  = "closure_lock_missing"
	CodeLockFormatUnsupported        = "closure_lock_format_unsupported"
	CodeLockStale                    = "closure_lock_stale"
	CodeIntegrityMissing             = "closure_integrity_missing"
	CodeIntegrityMismatch            = "closure_integrity_mismatch"
	CodeOriginUnpinned               = "closure_origin_unpinned"
	CodeGraphIncomplete              = "closure_graph_incomplete"
	CodeLocalPathEscape              = "closure_local_path_escape"
	CodeBundledDependencyUnsupported = "closure_bundled_dependency_unsupported"
	CodeHookUndeclared               = "closure_hook_undeclared"
	CodeNativeBuildUnsupported       = "closure_native_build_unsupported"
	CodeOfflineInputMissing          = "closure_offline_input_missing"
	CodeNetworkAttempted             = "closure_network_attempted"
	CodeMetadataMismatch             = "closure_metadata_mismatch"
	CodeProcessUndeclared            = "closure_process_undeclared"
	CodeInputUndeclared              = "closure_input_undeclared"
	CodeWriteUndeclared              = "closure_write_undeclared"
	CodeDerivationUnauthorized       = "closure_derivation_unauthorized"
	CodeRuntimeIdentityChanged       = "closure_runtime_identity_changed"
)

// Error carries one stable shared closure diagnostic and structured fields.
type Error struct {
	Code   string
	Detail string
	Fields map[string]string
}

func (e *Error) Error() string {
	if e.Detail == "" {
		return e.Code
	}
	return e.Code + ": " + e.Detail
}

func fail(code, detail string, fields map[string]string) error {
	return &Error{Code: code, Detail: detail, Fields: fields}
}

// ErrorCode extracts a stable closure diagnostic from an npm adapter error.
func ErrorCode(err error) string {
	if typed, ok := err.(*Error); ok {
		return typed.Code
	}
	return ""
}
