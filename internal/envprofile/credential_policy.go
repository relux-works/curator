package envprofile

import (
	"fmt"

	"github.com/relux-works/curator/internal/envregistry"
)

// admitCredentialStore prevents a stored selector from silently becoming native
// on an older producer. This is an activation boundary, not runtime qualification.
// Effective-policy adoption and successor login/fragment consumers are separate.
func (req *ResolveRequest) admitCredentialStore(adapter envregistry.Adapter) error {
	selection, err := req.Machine.CredentialStoreFor(req.Profile, adapter.ID)
	if err != nil {
		return err
	}
	if selection == envregistry.CredentialStoreNative {
		return nil
	}
	if adapter.ID == envregistry.CodexCLI {
		native, err := req.nativeHome(adapter.ID)
		if err != nil {
			return err
		}
		// Even an explicit file selection cannot turn failed or unsupported
		// native inspection into an absent selector and a file fallback.
		if _, err := req.codexCredentialStore(native); err != nil {
			return err
		}
		if selection == envregistry.CredentialStoreFile {
			return fmt.Errorf("%s: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback", envregistry.DiagCodexSettingsUnsupported)
		}
	}
	guidance := "this adapter has no qualified forced-store mapping"
	switch adapter.ID {
	case envregistry.ClaudeCode:
		guidance = "Claude has no established force-file or force-Keychain interface; log in from the same kind of session as the launcher, or explicitly enroll an environment-token source for inference"
	case envregistry.Muse:
		guidance = "managed Muse requires file storage; successor file-policy login and fragment delivery are not enabled"
	}
	return fmt.Errorf("%s: %s/%s selection %s: %s", envregistry.DiagCredentialStoreUnsupported, req.Profile, adapter.ID, selection, guidance)
}
