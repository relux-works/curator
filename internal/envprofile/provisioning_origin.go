package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

func (req *ResolveRequest) originBinding(adapter envregistry.Adapter, hash string) envorigin.Binding {
	return envorigin.Binding{Profile: req.Profile, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:"), ManagedTarget: ManagedHomeDir(req.Home, req.Profile, adapter.ID)}
}

// prepareProvisioningOrigin observes the standalone authority before any home
// mutation. An existing marker never permits a missing record to be inferred.
func (req *ResolveRequest) prepareProvisioningOrigin(adapter envregistry.Adapter, hash string, prior *envmarker.Marker) error {
	payload, present, err := req.readProvisioningOrigin(adapter)
	if err != nil {
		return fmt.Errorf("%s: private record cannot be read", envorigin.DiagInvalid)
	}
	binding := req.originBinding(adapter, hash)
	if present {
		if prior == nil {
			return fmt.Errorf("%s: orphan origin record", envorigin.DiagInvalid)
		}
		record, e := envorigin.Parse(payload)
		if e != nil {
			return e
		}
		if e = record.Match(adapter.ID, binding); e != nil {
			return e
		}
		if e = req.validateRecordedNative(record); e != nil {
			return e
		}
		req.provisioningOrigin = record
		return nil
	}
	if prior != nil {
		if req.adoptingOrigin && req.provisioningOrigin != nil {
			if err := req.provisioningOrigin.Match(adapter.ID, binding); err != nil {
				return err
			}
			return req.validateRecordedNative(req.provisioningOrigin)
		}
		return fmt.Errorf("%s: legacy home requires explicit trusted-input env adopt", envorigin.DiagRequired)
	}
	// A pre-lock capture remains usable only for this exact operation binding;
	// callers clear it at fresh lock/recovery boundaries.
	if req.provisioningOrigin != nil && req.provisioningOrigin.Match(adapter.ID, binding) == nil {
		return nil
	}
	operator, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("%s: operator-home anchor unavailable", envorigin.DiagInvalid)
	}
	operator, err = nativeComparablePath(operator)
	if err != nil {
		return fmt.Errorf("%s: operator-home anchor cannot be resolved", envorigin.DiagInvalid)
	}
	native, err := req.nativeHome(adapter.ID)
	if err != nil {
		return err
	}
	native, err = filepath.Abs(native)
	if err != nil {
		return fmt.Errorf("%s: native root cannot be resolved", envorigin.DiagInvalid)
	}
	selector := envorigin.Selector{Name: adapter.EnvVar, State: "absent"}
	if value, ok := os.LookupEnv(adapter.EnvVar); ok {
		selector.Value = &value
		selector.State = "present-value"
		if value == "" {
			selector.State = "present-empty"
		}
	}
	record := &envorigin.Record{Record: envorigin.Version, Adapter: adapter.ID, OperatorHome: operator, Selector: selector, NativeRoot: native, Capture: envorigin.Capture{Method: "provisioning", Authority: "native-selection", Time: time.Now().UTC().Format(time.RFC3339Nano)}, Binding: binding}
	if _, err = record.Marshal(); err != nil {
		return err
	}
	if err = req.validateRecordedNative(record); err != nil {
		return err
	}
	req.provisioningOrigin = record
	return nil
}

func originHomeSuffix(adapter envregistry.Adapter) string {
	if adapter.ID == envregistry.OpenCode {
		return "opencode"
	}
	return ""
}

func (req *ResolveRequest) validateRecordedNative(record *envorigin.Record) error {
	for _, candidate := range []string{record.NativeRoot, record.OperatorHome} {
		managed, err := nativePathIsManaged(candidate, req.Home)
		if err != nil || managed {
			return fmt.Errorf("%s: origin intersects managed state or cannot be inspected", envorigin.DiagInvalid)
		}
	}
	return nil
}

// Credential routing is separate from native-skill routing. Only validated
// origin is supplied by assembly/resolve; low-level policy fixtures without a
// managed home retain their explicitly injected native-selection seam.
func (req *ResolveRequest) credentialNativeHome(id string) (string, error) {
	if req.provisioningOrigin != nil {
		if req.provisioningOrigin.Adapter != id {
			return "", fmt.Errorf("%s: adapter mismatch", envorigin.DiagInvalid)
		}
		return req.provisioningOrigin.NativeRoot, nil
	}
	return req.nativeHome(id)
}

func (req *ResolveRequest) originPolicyIdentity() string {
	if req.provisioningOrigin == nil {
		return ""
	}
	digest, err := req.provisioningOrigin.Digest()
	if err != nil {
		return "invalid"
	}
	return digest
}

// readOriginMarker uses the shared absence/error boundary and admits no writes.
func (req *ResolveRequest) readOriginMarker(adapter envregistry.Adapter) (*envmarker.Marker, error) {
	read := req.readStateFile
	if read == nil {
		read = stateread.ReadFile
	}
	return envmarker.ReadWith(ManagedHomeDir(req.Home, req.Profile, adapter.ID), read)
}

// A present non-regular, linked, public or unreadable record is invalid, never
// legacy absence. Parents and the final entry are inspected without following.
func (req *ResolveRequest) readProvisioningOrigin(adapter envregistry.Adapter) ([]byte, bool, error) {
	rel := filepath.Join("environments", req.Profile, adapter.ID, originHomeSuffix(adapter), envorigin.Name)
	full, err := managedPath(req.Home, rel, false)
	if err != nil {
		return nil, false, err
	}
	metadata, err := stateread.Lstat(full)
	if err != nil {
		return nil, false, err
	}
	if metadata.Kind == stateread.KindAbsent {
		return nil, false, nil
	}
	if metadata.Kind != stateread.KindPresent || metadata.Info == nil || !metadata.Info.Mode().IsRegular() {
		return nil, false, fmt.Errorf("origin must be a regular file")
	}
	if err := privatedir.ValidateFile(full); err != nil {
		return nil, false, err
	}
	read := req.readRegularFile
	if read == nil {
		read = stateread.ReadRegularFile
	}
	file, err := read(full)
	if err != nil || file.Kind != stateread.KindPresent {
		return nil, false, fmt.Errorf("origin record observation failed")
	}
	return file.Bytes, true, nil
}
