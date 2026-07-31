//go:build windows

package transaction

import (
	"fmt"
	"os"
)

// completeNamespaceIdentity finishes the identity read Windows leaves for later.
//
// os.Stat and os.Lstat return a record that still carries the path there: the
// volume serial and file index os.SameFile actually compares are read from the
// live filesystem the first time SameFile is asked for them. Two things follow
// for the pairwise sweep, and neither is what the rest of it assumes.
//
// The pass snapshot is not a snapshot. A path resolved early in a pass would
// have its object read at whatever later moment the first surviving pair
// reached it, so the sweep could compare one path as it was against another as
// it had since become.
//
// A failed read is fail-open. os.SameFile reports only a bool, so a read it
// could not perform — the object replaced or removed under the pass, a handle
// the process cannot open — arrives at the sweep as "these are different
// objects", which is the answer that lets an aliased pair through. The unix
// hosts have no such branch: a stat that cannot answer fails there, and the
// pair fails closed on it.
//
// Completing the read here restores both properties: the identity is bound to
// the object the pass named, every later comparison is a pure in-memory one,
// and a read that cannot be completed becomes the inspection failure the sweep
// already knows how to refuse.
func completeNamespaceIdentity(candidate targetNamespacePath, info os.FileInfo) (os.FileInfo, error) {
	// Comparing the record with itself is what forces the deferred read; the
	// answer is meaningful only as "the identity could be read".
	if os.SameFile(info, info) {
		return info, nil
	}
	// SameFile does not say why. Ask the filesystem again so the caller sees
	// the real reason, and in particular so a path that has gone away since the
	// stat is reported as the not-exist an eager host would have reported.
	if _, err := namespaceStat(candidate); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("read the file identity of %q", candidate.path)
}
