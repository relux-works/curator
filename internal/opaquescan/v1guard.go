package opaquescan

import (
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/hashing"
)

// FindingNUL is the stable opaque-NUL finding id shared by every v1
// pre-hash and pre-trust refusal (Spec §8 interim rule for v1 readers).
// Audit findings render this id; entry points without a finding channel
// (currentness readers, frozen hashes, store pin checks) carry it in
// their refusal error so operators and tests see one vocabulary.
const FindingNUL = "audit.opaque.nul-byte"

// RefuseNULV1 enforces the Spec §8 interim rule before a v1 identity is
// computed or trusted over root: when version selects the v1 framing it
// scans every regular file below root and refuses with an error carrying
// FindingNUL if any file contains a NUL byte, so no v1 digest is ever
// computed over those bytes. A v2 root hashes NUL as ordinary data and
// skips the scan entirely. An unknown version refuses without scanning
// rather than guessing a rule.
func RefuseNULV1(root string, version hashing.Version) error {
	switch version {
	case hashing.VersionV1:
		paths, err := NULPaths(root)
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			return nil
		}
		return fmt.Errorf("%s: regular file contains a NUL byte and is treated as opaque (files: %s)",
			FindingNUL, strings.Join(paths, ", "))
	case hashing.VersionV2:
		return nil
	default:
		return fmt.Errorf("unsupported content hash version %d", version)
	}
}
