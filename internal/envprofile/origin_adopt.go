package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/transaction"
)

// OriginAdoptRequest supplies explicit operator input. None of these origin
// facts is inferred from process HOME, a default or an existing credential link.
type OriginAdoptRequest struct {
	Resolve       ResolveRequest
	OperatorHome  string
	NativeRoot    string
	SelectorState string
	SelectorValue *string
}

// AdoptOrigin validates a legacy home and publishes only its private authority
// record. The home remains unchanged; its exact marker and lock are frozen
// transaction inputs, including for restart recovery. False means idempotent.
func AdoptOrigin(input OriginAdoptRequest) (bool, error) {
	req := input.Resolve
	if req.Profile == "" || req.Repair || req.DryRun || req.PreserveLock {
		return false, fmt.Errorf("usage: env adopt requires an explicit profile and trusted origin input")
	}
	adapter, err := envregistry.ByID(req.EnvID)
	if err != nil {
		return false, err
	}
	_, _, hash, err := loadResolveInputsWithOwner(req.Home, req.Profile, req.readRegularFile, req.boundaryOwnerLookup)
	if err != nil {
		return false, err
	}
	candidate := &envorigin.Record{Record: envorigin.Version, Adapter: adapter.ID, OperatorHome: input.OperatorHome, NativeRoot: input.NativeRoot, Selector: envorigin.Selector{Name: adapter.EnvVar, State: input.SelectorState, Value: input.SelectorValue}, Capture: envorigin.Capture{Method: "adoption", Authority: "explicit-operator-input", Time: time.Now().UTC().Format(time.RFC3339Nano)}, Binding: req.originBinding(adapter, hash)}
	payload, err := candidate.Marshal()
	if err != nil {
		return false, err
	}
	if err = req.validateRecordedNative(candidate); err != nil {
		return false, err
	}
	home := ManagedHomeDir(req.Home, req.Profile, adapter.ID)
	var unchanged bool
	validate := func() error {
		if _, err := validateResolveRoots(req.Home, req.boundaryOwnerLookup); err != nil {
			return err
		}
		if err := validateMarkerBoundary(req.Home, req.Profile, adapter.ID, req.boundaryOwnerLookup); err != nil {
			return err
		}
		source, lock, currentHash, err := loadResolveInputsWithOwner(req.Home, req.Profile, req.readRegularFile, req.boundaryOwnerLookup)
		if err != nil {
			return err
		}
		if currentHash != hash {
			return fmt.Errorf("%s: adoption lock changed", envorigin.DiagInvalid)
		}
		marker, err := req.readOriginMarker(adapter)
		if err != nil {
			return err
		}
		if marker == nil {
			return fmt.Errorf("%s: adoption requires a provisioned legacy home", envorigin.DiagInvalid)
		}
		if err = requirePerHomeCredentialSource(marker); err != nil {
			return err
		}
		raw, present, err := req.readProvisioningOrigin(adapter)
		if err != nil {
			return fmt.Errorf("%s: origin cannot be read", envorigin.DiagInvalid)
		}
		unchanged = false
		if present {
			existing, err := envorigin.Parse(raw)
			if err != nil {
				return err
			}
			if existing.Adapter != candidate.Adapter || existing.OperatorHome != candidate.OperatorHome || existing.NativeRoot != candidate.NativeRoot || existing.Binding != candidate.Binding || !reflect.DeepEqual(existing.Selector, candidate.Selector) {
				return fmt.Errorf("%s: adoption cannot replace an existing origin", envorigin.DiagInvalid)
			}
			unchanged = true
		}
		req.provisioningOrigin = candidate
		req.adoptingOrigin = true
		req.credentialPolicy = nil
		verdict := verifyHome(&req, adapter, source, lock, currentHash)
		if !verdict.current {
			if verdict.credentialSourceError != nil {
				return verdict.credentialSourceError
			}
			return fmt.Errorf("%s: home does not match explicit adoption input", envregistry.DiagCredentialConflict)
		}
		return nil
	}
	// Failed admission creates no lock or transaction state.
	if err = validate(); err != nil {
		return false, err
	}
	if unchanged {
		return false, nil
	}
	op, err := beginOperationValidated(req.Home, validate, req.transactionOptions...)
	if err != nil {
		return false, err
	}
	defer func() { _ = op.close() }()
	if err = validate(); err != nil {
		return false, op.recoveryRefusal(err)
	}
	if unchanged {
		return false, nil
	}
	stage, err := privatedir.CreateTemp("", "curator-origin-adopt-*")
	if err != nil {
		return false, err
	}
	defer func() { _ = os.Remove(stage.Name()) }()
	if _, err = stage.Write(payload); err != nil {
		_ = stage.Close()
		return false, err
	}
	if err = stage.Close(); err != nil {
		return false, err
	}
	inputs := map[string]string{}
	for _, path := range []string{filepath.Join(home, envmarker.Name), lockPath(req.Home, req.Profile)} {
		digest, err := transaction.DigestPath(path)
		if err != nil {
			return false, err
		}
		inputs[path] = digest
	}
	live := filepath.Join(home, envorigin.Name)
	preimage, err := transaction.DigestPath(live)
	if err != nil {
		return false, err
	}
	if preimage != transaction.DigestAbsent {
		return false, fmt.Errorf("%s: authority changed during adoption", envorigin.DiagInvalid)
	}
	plan := transaction.Plan{TransactionID: newTransactionID(), ProjectIdentity: "provisioning-origin-adoption", ReadInputs: inputs, Targets: []transaction.Target{{PrivateFile: true, Class: "provisioning-origin", Identifier: live, LivePath: live, StagedSource: stage.Name(), PreimageDigest: preimage}}}
	if _, err = op.engine.Prepare(op.lock, plan); err != nil {
		return false, err
	}
	if err = op.engine.Commit(op.lock, plan.TransactionID); err != nil {
		return false, err
	}
	return true, nil
}
