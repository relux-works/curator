package envjournal

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/relux-works/curator/internal/entrydigest"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envunit"
	"github.com/relux-works/curator/internal/projectleaserecord"
)

// LeaseReferenceGap recognizes only the fixed shared ledger's backup/install
// window. It supplies bounded closed metadata for equality checks; it grants
// no home exception, launch, new reference or forward effect. The executor
// still validates the complete journal before any recovery write.
func LeaseReferenceGap(j *Journal, manager string) (before, after []byte, gap bool, err error) {
	if j.ProjectIdentity != "project-leases" || j.Unit != nil || j.Global != nil || len(j.Opaque) != 0 || (j.Schema != "curator-install-transaction-v1" && j.Schema != "curator-install-transaction-v2") || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`).MatchString(j.TransactionID) {
		return nil, nil, false, unitError("lease replacement header differs")
	}
	if j.Phase != PhaseCommitting && j.Phase != PhaseRollingBack {
		return nil, nil, false, nil
	}
	path := envunit.LeaseReferencePath(manager)
	found := false
	for index, target := range j.Targets {
		if target.LivePath != path {
			continue
		}
		if found {
			return nil, nil, false, unitError("duplicate lease replacement")
		}
		found = true
		if target.Class != "project-lease" || target.Kind != KindBytes || target.Identifier != path || target.PreimageDigest == "absent" || target.DesiredDigest == "absent" {
			return nil, nil, false, unitError("lease replacement target differs")
		}
		stem := fmt.Sprintf(".curator-txn-%s-%03d.", entrydigest.ShortIdentity(j.TransactionID), index)
		parent := filepath.Dir(path)
		if target.BackupPath != filepath.Join(parent, stem+"backup") || target.RollbackPath != filepath.Join(parent, stem+"rollback") || target.StagedPath != filepath.Join(parent, stem+"desired") {
			return nil, nil, false, unitError("lease replacement sidecars differ")
		}
		_, present, e := envorigin.ReadPrivateAuthority(manager, path, projectleaserecord.MaxRecord)
		if e != nil {
			return nil, nil, false, e
		}
		if present {
			return nil, nil, false, nil
		}
		if target.State != StatePending && target.State != StateBackedUp && (j.Phase != PhaseRollingBack || target.State != StateCommitted) {
			return nil, nil, false, unitError("lease replacement phase differs")
		}
		read := func(candidate, digest string) ([]byte, error) {
			raw, present, e := envorigin.ReadPrivateAuthority(manager, candidate, projectleaserecord.MaxRecord)
			if e != nil {
				return nil, e
			}
			if !present {
				return nil, unitError("lease replacement evidence absent")
			}
			actual, e := target.Digest(candidate)
			if e != nil || actual != digest {
				return nil, unitError("lease replacement evidence digest differs")
			}
			if _, e := projectleaserecord.DecodeLedger(raw); e != nil {
				return nil, e
			}
			return raw, nil
		}
		before, e = read(target.BackupPath, target.PreimageDigest)
		if e != nil {
			return nil, nil, false, e
		}
		if j.Phase == PhaseCommitting {
			after, e = read(target.StagedPath, target.DesiredDigest)
			if e != nil {
				return nil, nil, false, e
			}
		}
		gap = true
	}
	return before, after, gap, nil
}
