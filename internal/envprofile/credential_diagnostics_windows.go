//go:build windows

package envprofile

// Windows link creation/readback can normalize absolute target spelling before
// inspection. A metadata comparison cannot certify the Unix exact-target
// contract, and Claude has no admitted Windows shared file-link backend.
// Report that boundary without inspecting credentials or claiming readiness.
func (req *ResolveRequest) inspectCredentialLink(full, target string) CredentialLinkReason {
	return CredentialLinkUnsupported
}
