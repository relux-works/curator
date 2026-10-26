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
	lstat := req.passthroughLstat
	if lstat == nil {
		lstat = os.Lstat
	}
	metadata, err := stateread.LstatWith(full, lstat)
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
	readlink := req.passthroughReadlink
	if readlink == nil {
		readlink = os.Readlink
	}
	link, err := stateread.ReadlinkWith(full, readlink)
	if err != nil || link.Kind == stateread.KindUnreadable {
		return CredentialLinkUnreadable
	}
	if link.Kind == stateread.KindAbsent {
		return CredentialLinkMissing
	}
	got := link.Target
	if !filepath.IsAbs(got) {
		got = filepath.Join(filepath.Dir(full), got)
	}
	if filepath.Clean(got) != filepath.Clean(target) {
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
