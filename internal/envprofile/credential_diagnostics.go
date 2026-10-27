package envprofile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/stateread"
)

// CredentialLinkReason describes file-link readiness, never token validity.
type CredentialLinkReason string

// Credential link reasons describe only the observed filesystem shape.
const (
	CredentialLinkOK          CredentialLinkReason = "ok"
	CredentialLinkMissing     CredentialLinkReason = "missing"
	CredentialLinkDangling    CredentialLinkReason = "dangling"
	CredentialLinkReplaced    CredentialLinkReason = "replaced_regular_file"
	CredentialLinkWrongTarget CredentialLinkReason = "wrong_target"
	CredentialLinkWrongShape  CredentialLinkReason = "wrong_shape"
	CredentialLinkUnreadable  CredentialLinkReason = "unreadable"
	DiagCredentialLink                             = "environment_shared_credential_link"
)

// inspectCredentialLink performs metadata-only inspection. It never opens the
// credential file, even when the link was replaced by a provider refresh.
func (req *ResolveRequest) inspectCredentialLink(full, target string) CredentialLinkReason {
	var metadata stateread.Metadata
	var err error
	if req.passthroughLstat == nil {
		metadata, err = stateread.Lstat(full)
	} else {
		metadata, err = stateread.LstatWith(full, req.passthroughLstat)
	}
	if err != nil || metadata.Kind == stateread.KindUnreadable {
		return CredentialLinkUnreadable
	}
	if metadata.Kind == stateread.KindAbsent {
		return CredentialLinkMissing
	}
	if metadata.Info == nil {
		return CredentialLinkUnreadable
	}
	if metadata.Info.Mode().IsRegular() {
		return CredentialLinkReplaced
	}
	if metadata.Info.Mode()&os.ModeSymlink == 0 {
		return CredentialLinkWrongShape
	}
	var link stateread.Link
	if req.passthroughReadlink == nil {
		link, err = stateread.Readlink(full)
	} else {
		link, err = stateread.ReadlinkWith(full, req.passthroughReadlink)
	}
	if err != nil || link.Kind == stateread.KindUnreadable {
		return CredentialLinkUnreadable
	}
	if link.Kind == stateread.KindAbsent {
		return CredentialLinkMissing
	}
	// Preserve the verifier's exact-target contract. Lexical normalization
	// can erase symlink/.. traversal and cannot establish file identity.
	if link.Target != target {
		return CredentialLinkWrongTarget
	}
	metadata, err = stateread.Stat(full)
	if err != nil || metadata.Kind == stateread.KindUnreadable {
		return CredentialLinkUnreadable
	}
	if metadata.Kind == stateread.KindAbsent {
		return CredentialLinkDangling
	}
	if metadata.Info == nil {
		return CredentialLinkUnreadable
	}
	if !metadata.Info.Mode().IsRegular() {
		return CredentialLinkWrongShape
	}
	return CredentialLinkOK
}

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
