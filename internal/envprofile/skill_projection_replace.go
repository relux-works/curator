//go:build !darwin && !linux

package envprofile

import "github.com/relux-works/curator/internal/transaction"

// Windows retains its recoverable backup/install sequence: directory symlink
// replacement does not provide the Unix atomic name-exchange contract.
const skillProjectionReplacementKind = transaction.KindBytes
