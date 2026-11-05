package envprofile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// observePendingOrigin acquires only an existing lock and never invokes
// recovery. A normal legacy home cannot acquire authority from an unrelated
// journal, the invoking selector, or its current credential link.
func (req *ResolveRequest) observePendingOrigin(adapter envregistry.Adapter, hash string) (bool, error) {
	if req.borrowedOperation != nil {
		return req.preparePendingOrigin(req.borrowedOperation, adapter, hash)
	}
	if err := envauthority.Validate(req.Home); err != nil {
		return false, err
	}
	manager, err := managerlock.New(req.Home)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	lock, err := manager.AcquireExistingHomeOnly(ctx)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s: observe provisioning recovery: %w", DiagLockUnavailable, err)
	}
	defer func() { _ = lock.Close() }()
	if err := envauthority.Validate(req.Home); err != nil {
		return false, err
	}
	engine, err := newOperationEngine(req.Home, req.transactionOptions...)
	if err != nil {
		return false, err
	}
	return req.preparePendingOrigin(&operation{lock: lock, engine: engine}, adapter, hash)
}

// preparePendingOrigin admits only an interrupted first publication whose
// journal binds both absent-before marker and origin to this exact home and
// current lock. This is not legacy adoption or permission to replace a present
// invalid origin. Normal authority is read again after actual recovery.
func (req *ResolveRequest) preparePendingOrigin(op *operation, adapter envregistry.Adapter, hash string) (bool, error) {
	if err := op.lock.AssertCurrent(); err != nil {
		return false, err
	}
	_, present, err := req.readProvisioningOrigin(adapter)
	if err != nil || present {
		return false, err
	}
	journals, err := op.engine.PendingJournals(op.lock)
	if err != nil {
		return false, err
	}
	home := ManagedHomeDir(req.Home, req.Profile, adapter.ID)
	originPath := filepath.Join(home, envorigin.Name)
	markerPath := filepath.Join(home, envmarker.Name)
	var candidate *envorigin.Record
	var candidateID string
	touchingHome := map[string]bool{}
	for _, journal := range journals {
		var origin, marker *transaction.TargetRecord
		for i := range journal.Targets {
			target := &journal.Targets[i]
			_, below := managedRelative(home, target.LivePath)
			_, above := managedRelative(target.LivePath, home)
			if below == nil || above == nil || target.LivePath == home {
				touchingHome[journal.TransactionID] = true
			}
			switch target.LivePath {
			case originPath:
				origin = target
			case markerPath:
				marker = target
			}
		}
		if origin == nil {
			continue
		}
		if candidate != nil || journal.ProjectIdentity != "profiles" || marker == nil ||
			(journal.Phase != transaction.PhasePrepared && journal.Phase != transaction.PhaseCommitting && journal.Phase != transaction.PhaseCleanup) ||
			origin.PreimageDigest != transaction.DigestAbsent || marker.PreimageDigest != transaction.DigestAbsent {
			return false, fmt.Errorf("%s: pending origin is not an unambiguous first publication", envorigin.DiagInvalid)
		}
		originBytes, err := req.pendingOriginPayload(*origin)
		if err != nil {
			return false, err
		}
		markerBytes, err := req.pendingOriginPayload(*marker)
		if err != nil {
			return false, err
		}
		candidate, err = envorigin.Parse(originBytes)
		if err != nil {
			return false, err
		}
		candidateID = journal.TransactionID
		if err := candidate.Match(adapter.ID, req.originBinding(adapter, hash)); err != nil {
			return false, err
		}
		if candidate.Capture.Method != "provisioning" || candidate.Capture.Authority != "native-selection" {
			return false, fmt.Errorf("%s: pending record is not provisioning authority", envorigin.DiagInvalid)
		}
		if err := req.validateRecordedNative(candidate); err != nil {
			return false, err
		}
		parsedMarker, err := envmarker.Parse(markerBytes)
		if err != nil {
			return false, err
		}
		if err := requirePerHomeCredentialSource(parsedMarker); err != nil {
			return false, err
		}
		if parsedMarker.Mode != envmarker.ModeManagedHome || parsedMarker.Profile.Name != req.Profile ||
			parsedMarker.Profile.LockSHA256 != strings.TrimPrefix(hash, "sha256:") {
			return false, fmt.Errorf("%s: pending marker binding does not match", envorigin.DiagInvalid)
		}
	}
	if candidate == nil {
		return false, nil
	}
	for id := range touchingHome {
		if id != candidateID {
			return false, fmt.Errorf("%s: competing home publication requires explicit recovery disposition", envorigin.DiagInvalid)
		}
	}
	if err := op.lock.AssertCurrent(); err != nil {
		return false, err
	}
	req.provisioningOrigin = candidate
	req.credentialPolicy = nil
	return true, nil
}

// Only manager-owned private regular record bytes may supply recovery
// authority. Read the exact live/staged entry without following links, and
// require its digest both before and after the read. No credential is opened.
func (req *ResolveRequest) pendingOriginPayload(target transaction.TargetRecord) ([]byte, error) {
	invalid := func() error {
		return fmt.Errorf("%s: pending provisioning record is unavailable or changed", envorigin.DiagInvalid)
	}
	if target.Class != "profile" || target.Kind != transaction.KindBytes || target.DesiredDigest == transaction.DigestAbsent {
		return nil, invalid()
	}
	for _, path := range []string{target.LivePath, target.StagedPath} {
		if path == "" {
			continue
		}
		rel, err := managedRelative(req.Home, path)
		if err != nil {
			return nil, invalid()
		}
		if _, err := managedPath(req.Home, rel, false); err != nil {
			return nil, invalid()
		}
		digest, err := transaction.DigestTarget(target.Kind, path)
		if err != nil {
			return nil, invalid()
		}
		if digest == transaction.DigestAbsent {
			continue
		}
		if digest != target.DesiredDigest || privatedir.ValidateFile(path) != nil {
			return nil, invalid()
		}
		payload, present, err := readManagedRegular(req.Home, rel)
		if err != nil || !present {
			return nil, invalid()
		}
		digest, err = transaction.DigestTarget(target.Kind, path)
		if err != nil || digest != target.DesiredDigest {
			return nil, invalid()
		}
		return payload, nil
	}
	return nil, invalid()
}
