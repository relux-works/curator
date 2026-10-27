package envprofile

import (
	"fmt"
	"path/filepath"

	"github.com/relux-works/curator/internal/envregistry"
)

// CredentialLinkReason describes file-link readiness, never token validity.
type CredentialLinkReason string

// Credential link reasons describe observed shape or unavailable inspection.
const (
	CredentialLinkOK          CredentialLinkReason = "ok"
	CredentialLinkUnsupported CredentialLinkReason = "unsupported"
	CredentialLinkMissing     CredentialLinkReason = "missing"
	CredentialLinkDangling    CredentialLinkReason = "dangling"
	CredentialLinkReplaced    CredentialLinkReason = "replaced_regular_file"
	CredentialLinkWrongTarget CredentialLinkReason = "wrong_target"
	CredentialLinkWrongShape  CredentialLinkReason = "wrong_shape"
	CredentialLinkUnreadable  CredentialLinkReason = "unreadable"
	DiagCredentialLink                             = "environment_shared_credential_link"
)

// checkClaudeCredentialLink handles only the effective shared file link. On
// platforms using a different backend no file-link readiness claim is made.
func (v *verification) checkClaudeCredentialLink(req *ResolveRequest, plan *homePlan, path, target string) bool {
	if plan.adapter.ID != envregistry.ClaudeCode || path != ".credentials.json" {
		return false
	}
	reason := req.inspectCredentialLink(filepath.Join(plan.homeDir, path), target)
	if reason == CredentialLinkOK {
		return true
	}
	finding := fmt.Sprintf("%s: %s reason=%s", DiagCredentialLink, path, reason)
	switch reason {
	case CredentialLinkUnsupported:
		v.reasons = append(v.reasons, envregistry.DiagCredentialUnsupported+": "+finding+": exact credential-link inspection is unsupported on this platform")
	case CredentialLinkDangling:
		// Absence before first login remains warning-only (environments §7.4).
		v.warnings = append(v.warnings, finding+": detached-pending; shared credential target is absent")
	case CredentialLinkUnreadable:
		v.passthroughReadFailed = true
		v.reasons = append(v.reasons, envregistry.DiagPassthroughUnreadable+": "+finding)
	default:
		v.reasons = append(v.reasons, envregistry.DiagCredentialConflict+": "+finding+": detached shared credential link; no credential repair performed")
	}
	return true
}
