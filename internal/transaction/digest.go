package transaction

import (
	"github.com/relux-works/curator/internal/entrydigest"
	"os"
)

func DigestPath(path string) (string, error) { return entrydigest.DigestPath(path) }
func DigestTarget(kind TargetKind, path string) (string, error) {
	return entrydigest.DigestTarget(string(kind), path)
}
func DigestRegularPayload(payload []byte, mode os.FileMode) (string, error) {
	return entrydigest.DigestRegularPayload(payload, mode)
}
func digestRegularPayload(payload []byte, mode os.FileMode) (string, error) {
	return entrydigest.DigestRegularPayload(payload, mode)
}

func shortIdentity(value string) string { return entrydigest.ShortIdentity(value) }
