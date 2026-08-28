package pnpmsource

// ProfileID is the pinned pure-source pnpm closure profile.
const ProfileID = "pnpm-source-v1"

// SupportedPNPMVersion is the only external manager release whose lock,
// patch-hash, store, and virtual-layout semantics are admitted by ProfileID.
const SupportedPNPMVersion = "10.33.0"

// Stable diagnostics use the shared closure namespace.
const (
	CodeLockMissing                  = "closure_lock_missing"
	CodeLockFormatUnsupported        = "closure_lock_format_unsupported"
	CodeLockStale                    = "closure_lock_stale"
	CodeIntegrityMissing             = "closure_integrity_missing"
	CodeIntegrityMismatch            = "closure_integrity_mismatch"
	CodeOriginUnpinned               = "closure_origin_unpinned"
	CodeGraphIncomplete              = "closure_graph_incomplete"
	CodeLocalPathEscape              = "closure_local_path_escape"
	CodeBundledDependencyUnsupported = "closure_bundled_dependency_unsupported"
	CodeManagerPluginUndeclared      = "closure_manager_plugin_undeclared"
	CodeHookUndeclared               = "closure_hook_undeclared"
	CodeNativeBuildUnsupported       = "closure_native_build_unsupported"
	CodeOfflineInputMissing          = "closure_offline_input_missing"
	CodeRuntimeIdentityChanged       = "closure_runtime_identity_changed"
	CodeMetadataMismatch             = "closure_metadata_mismatch"
	CodeInputUndeclared              = "closure_input_undeclared"
	CodeDerivationUnauthorized       = "closure_derivation_unauthorized"
)

// Error is a stable fail-closed adapter diagnostic.
type Error struct {
	Code, Detail string
	Fields       map[string]string
}

func (e *Error) Error() string { return e.Code + ": " + e.Detail }

func fail(code, detail string, fields map[string]string) error {
	return &Error{Code: code, Detail: detail, Fields: fields}
}

// ErrorCode extracts the stable diagnostic from err.
func ErrorCode(err error) string {
	if value, ok := err.(*Error); ok {
		return value.Code
	}
	return ""
}
