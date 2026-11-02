package transaction

import (
	"github.com/relux-works/curator/internal/entrydigest"
	"os"
)

// DigestPath computes the no-follow entry preimage digest.
func DigestPath(path string) (string, error) { return entrydigest.DigestPath(path) }

// DigestTarget computes the digest using the declared target kind.
func DigestTarget(kind TargetKind, path string) (string, error) {
	return entrydigest.DigestTarget(string(kind), path)
}

// DigestRegularPayload computes the desired regular-file payload digest.
func DigestRegularPayload(payload []byte, mode os.FileMode) (string, error) {
	return entrydigest.DigestRegularPayload(payload, mode)
}
func digestRegularPayload(payload []byte, mode os.FileMode) (string, error) {
	return entrydigest.DigestRegularPayload(payload, mode)
}

func shortIdentity(value string) string { return entrydigest.ShortIdentity(value) }
