package projectcontext

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/stateread"
)

// Lapses have their own bounded record so even an existing, completely full
// preview/approval ledger can durably lose authority. There are at most 256
// registrations, each with at most a 64-byte ID; this record cannot reach 32 KiB.
// Both records are read and written under the same manager home lock.
const maxLapseRecord = 32 << 10

type rootLapses struct {
	Version     int      `json:"version"`
	CheckoutIDs []string `json:"checkout_ids"`
}

func (op *operation) loadLapses(root *os.Root) error {
	st, err := stateread.Lstat(filepath.Join(op.root, "lapses.json"))
	if err != nil {
		return fail("project_context_drift", "cannot inspect retained root-authority lapses")
	}
	if st.Kind == stateread.KindAbsent {
		// Older writers retained lapses inline in state.json. No companion
		// is required until that envelope has been migrated to version 2.
		if op.state.Version == 1 {
			return nil
		}
		return fail("project_context_drift", "missing retained root-authority lapses")
	}
	b, _, err := read(root, "lapses.json", maxLapseRecord)
	if err != nil {
		return fail("project_context_drift", "cannot read retained root-authority lapses")
	}
	var lapses rootLapses
	if err := decode(b, &lapses); err != nil {
		return fail("project_context_drift", "invalid retained root-authority lapses")
	}
	can, err := canonical(lapses)
	if err != nil || string(can) != string(b) || lapses.Version != 1 || len(lapses.CheckoutIDs) > 256 {
		return fail("project_context_drift", "invalid retained root-authority lapses")
	}
	previous := ""
	for _, id := range lapses.CheckoutIDs {
		r, ok := op.state.Registrations[id]
		if !ok || !idRE.MatchString(id) || id <= previous {
			return fail("project_context_drift", "invalid retained root-authority lapse identity")
		}
		r.Lapsed = true
		op.state.Registrations[id] = r
		previous = id
	}
	return nil
}

func (op *operation) retainLapses() error {
	if err := op.validate(); err != nil {
		return err
	}
	lapses := rootLapses{Version: 1, CheckoutIDs: []string{}}
	for id, r := range op.state.Registrations {
		if r.Lapsed {
			lapses.CheckoutIDs = append(lapses.CheckoutIDs, id)
		}
	}
	sort.Strings(lapses.CheckoutIDs)
	b, err := canonical(lapses)
	if err != nil {
		return err
	}
	if len(b) > maxLapseRecord {
		return fail("project_context_drift", "root-authority lapse capacity exceeded")
	}
	return writeAtomic(op.root, "lapses.json", b)
}
