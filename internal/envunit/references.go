package envunit

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/entrydigest"
	"github.com/relux-works/curator/internal/projectleaserecord"
)

// ReferencePreimage captures only the closed manager-owned lease metadata.
// It contains no session payload, credential contents or native root state.
// The full byte/mode input is frozen for the creating transaction; retained
// admission subsequently compares only the complete affected profile subset,
// allowing another profile to update the shared container without changing it.
type ReferencePreimage struct {
	Image RecordImage `json:"image"`
	Mode  uint32      `json:"mode"`
}

// LeaseReferencePath returns the fixed manager-owned lease reference record.
func LeaseReferencePath(manager string) string {
	return filepath.Join(manager, "state", "project-leases", "v1", "ledger.json")
}

// InputDigest returns the digest of the captured lease-reference preimage.
func (p ReferencePreimage) InputDigest() (string, error) {
	if p.Image.Kind == "absent" {
		return "absent", nil
	}
	return entrydigest.DigestRegularPayload(p.Image.Bytes, os.FileMode(p.Mode))
}
func (u *Unit) validateReferenceCapture() error {
	p := u.LeasePreimage
	if p == nil {
		if u.Action == GlobalExtend && u.LedgerPublication != nil {
			return invalid("global retained reference capture absent")
		}
		return nil
	}
	if u.Action != GlobalExtend || p.Image.validate() != nil || p.Mode > 07777 || p.Image.Kind == "absent" && p.Mode != 0 {
		return invalid("reference preimage binding")
	}
	if p.Image.Kind == "file" {
		if _, err := projectleaserecord.DecodeLedger(p.Image.Bytes); err != nil {
			return invalid("reference preimage grammar: %v", err)
		}
	}
	return nil
}
func leaseProjection(raw []byte, profiles []ProfileTransition) ([]byte, error) {
	l := &projectleaserecord.Ledger{Kind: "project-lease-state/1", Leases: map[string]projectleaserecord.Entry{}, Views: map[string]projectleaserecord.View{}}
	if len(raw) != 0 {
		decoded, err := projectleaserecord.DecodeLedger(raw)
		if err != nil {
			return nil, invalid("reference ledger: %v", err)
		}
		selected := map[string]bool{}
		for _, p := range profiles {
			selected[p.Name] = true
		}
		for k, v := range decoded.Leases {
			if selected[v.Composition.Profile] {
				l.Leases[k] = v
			}
		}
		for k, v := range decoded.Views {
			if selected[v.Composition.Profile] {
				l.Views[k] = v
			}
		}
	}
	return projectleaserecord.Canonical(l)
}

// CheckLeaseReferences validates equality in both directions: neither a new
// protected reference nor removal/mutation of an original one is disjoint.
func (u *Unit) CheckLeaseReferences(raw []byte) error {
	if u.LeasePreimage == nil {
		return invalid("global retained reference capture absent")
	}
	if err := u.validateReferenceCapture(); err != nil {
		return err
	}
	before, err := leaseProjection(u.LeasePreimage.Image.Bytes, u.Profiles)
	if err != nil {
		return err
	}
	after, err := leaseProjection(raw, u.Profiles)
	if err != nil {
		return err
	}
	if !bytes.Equal(before, after) {
		return invalid("retained profile lease reference preimages changed")
	}
	return nil
}
