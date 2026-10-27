//go:build !windows

package envprofile

import (
	"os"

	"github.com/relux-works/curator/internal/stateread"
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
