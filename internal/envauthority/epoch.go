package envauthority

import (
	"github.com/relux-works/curator/internal/envlifecycle"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
)

// Epoch binds operation-local policy observations to the actual private ledger,
// including absence, revision and every receipt's complete identity/phase. It
// never replaces classification, target admission or post-recovery rechecks.
func Epoch(manager string) (string, error) {
	identity, err := envunit.ObserveManager(manager)
	if err != nil {
		return "", err
	}
	_, raw, err := envlifecycle.Read(identity)
	if err != nil {
		return "", err
	}
	if len(raw) == 0 {
		return "absent", nil
	}
	return envorigin.Digest(raw), nil
}
