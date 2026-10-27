package state

import (
	"os"
	"path/filepath"
	"strings"
)

// PublicationTarget is the journal adapter's closed, path-backed evidence.
// It conveys no permission to mutate: the engine independently validates its
// own journal, ownership, preimages and sidecar digests.
type PublicationTarget struct {
	Live, Class, Identifier, Kind, Desired, Preimage string
	Private                                          bool
	Candidates, Originals, AllPaths                  []string
}

// ValidatePublication prevents a forged owner string from bypassing home
// guards. Forward recovery validates the complete desired ledger and immutable
// evidence; rollback validates the original projection it would restore.
// readCandidate accepts only bytes matching the engine's recorded digest.
func ValidatePublication(home, phase string, targets []PublicationTarget, readCandidate func(string, string) ([]byte, bool, error)) error {
	rollback := phase == "rolling_back" || phase == "rollback_cleanup" || phase == "preparing"
	files := map[string][]byte{}
	hasLedger, absentLedger := false, false
	for _, target := range targets {
		if target.Class != "home-lease" || target.Identifier != target.Live || target.Kind != "" && target.Kind != "bytes" || !target.Private || target.Desired == "absent" {
			return invalid()
		}
		rel, err := filepath.Rel(Namespace(home), target.Live)
		if err != nil || filepath.Clean(target.Live) != target.Live {
			return invalid()
		}
		if rel != "ledger.json" {
			parts := strings.Split(filepath.ToSlash(rel), "/")
			if len(parts) != 2 || parts[0] != "bindings" && parts[0] != "events" || !strings.HasSuffix(parts[1], ".json") || !ValidID(strings.TrimSuffix(parts[1], ".json")) || target.Preimage != "absent" {
				return invalid()
			}
		} else {
			hasLedger = true
		}
		for _, path := range target.AllPaths {
			if path != "" {
				if err := CheckRoute(home, path); err != nil {
					return Refuse("environment_lease_unreadable", "state_unreadable")
				}
				parent, err := filepath.Rel(Namespace(home), filepath.Dir(path))
				if err != nil || parent != "." && parent != "bindings" && parent != "events" {
					return invalid()
				}
			}
		}
		paths, digest := target.Candidates, target.Desired
		if rollback {
			paths, digest = target.Originals, target.Preimage
		}
		if digest == "absent" {
			if rel == "ledger.json" {
				absentLedger = true
			}
			files[target.Live] = nil
			continue
		}
		found := false
		for _, path := range paths {
			if path == "" {
				continue
			}
			raw, matches, err := readCandidate(path, digest)
			if err != nil {
				return Refuse("environment_lease_unreadable", "state_unreadable")
			}
			if matches {
				files[target.Live], found = raw, true
				break
			}
		}
		if !found {
			return invalid()
		}
	}
	if !hasLedger {
		return invalid()
	}
	if absentLedger {
		return nil
	}
	_, err := DecodeLedger(home, files[LedgerPath(home)], func(path string) ([]byte, error) {
		if raw, ok := files[path]; ok {
			if raw == nil {
				return nil, os.ErrNotExist
			}
			return raw, nil
		}
		return Read(home, path)
	})
	return err
}
