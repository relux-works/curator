package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/stateread"
)

// originObservation belongs to one request, never a process/global cache.
// Re-observation at lock/recovery boundaries rechecks the actual private file;
// initial capture retains its original raw selector across those boundaries.
type originObservation struct {
	record     *envorigin.Record
	raw        []byte
	newCapture bool
}

func (req *ResolveRequest) observeOrigin(adapterID string, marker *envmarker.Marker) error {
	if req.EnvID != "" && req.EnvID != adapterID {
		return fmt.Errorf("%s: origin adapter differs from selected operation", envorigin.DiagInvalid)
	}
	lock, hash, err := readLock(req.Home, req.Profile)
	if err != nil {
		return err
	}
	if lock == nil {
		return fmt.Errorf("%s: origin requires exact installed lock", envorigin.DiagInvalid)
	}
	binding := envorigin.Binding{Profile: req.Profile, LockSHA256: "sha256:" + strings.TrimPrefix(hash, "sha256:"), ManagedTarget: envorigin.ManagedTarget(req.Home, req.Profile, adapterID)}
	if marker != nil {
		if err := requirePerHomeCredentialSource(marker); err != nil {
			return err
		}
		if marker.Mode != envmarker.ModeManagedHome || marker.Profile.Name != req.Profile || strings.TrimPrefix(marker.Profile.LockSHA256, "sha256:") != strings.TrimPrefix(hash, "sha256:") {
			return fmt.Errorf("%s: marker/installed lock binding mismatch", envorigin.DiagInvalid)
		}
		record, raw, err := envorigin.Read(req.Home, binding, adapterID)
		if err != nil {
			return err
		}
		if req.origin != nil && req.origin.newCapture && !bytes.Equal(raw, req.origin.raw) {
			return fmt.Errorf("%s: authority changed during new provisioning", envorigin.DiagInvalid)
		}
		req.origin = &originObservation{record: record, raw: raw}
		return nil
	}
	path := filepath.Join(binding.ManagedTarget, envorigin.Name)
	if _, present, err := envorigin.ReadPrivateAuthority(req.Home, path, envorigin.MaxRecordBytes); err != nil {
		return err
	} else if present {
		return fmt.Errorf("%s: origin sidecar without its managed marker", envorigin.DiagInvalid)
	}
	root := envorigin.ManagedRoot(req.Home, req.Profile, adapterID)
	if info, err := os.Lstat(root); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: unmarked home is not an owned directory", envorigin.DiagInvalid)
		}
		if err := pathboundary.ValidateRouteWithOwner(req.Home, root, pathboundary.DefaultOwnerLookup()); err != nil {
			return fmt.Errorf("%s: unmarked home boundary: %w", envorigin.DiagInvalid, err)
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			return fmt.Errorf("%s: unmarked home inventory: %w", envorigin.DiagInvalid, err)
		}
		if len(entries) != 0 {
			return fmt.Errorf("%s: unmarked existing payload is not new provisioning", envorigin.DiagInvalid)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("%s: home metadata: %w", envorigin.DiagInvalid, err)
	}
	if req.origin != nil && req.origin.newCapture {
		record, err := envorigin.Parse(req.origin.raw, binding, adapterID, req.Home)
		if err != nil {
			return err
		}
		req.origin.record = record
		return nil
	}
	record, err := envorigin.Select(req.Home, binding, adapterID, req.originAccountLookup)
	if err != nil {
		return err
	}
	raw, err := record.Marshal()
	if err != nil {
		return err
	}
	req.origin = &originObservation{record: record, raw: raw, newCapture: true}
	return nil
}

// refreshOrigin uses the fixed managed marker, never a current native-home/link
// inference. It precedes backend/probe observation, including ambient adapters.
func (req *ResolveRequest) refreshOrigin(adapterID string) error {
	if err := validateMarkerBoundary(req.Home, req.Profile, adapterID, req.boundaryOwnerLookup); err != nil {
		return err
	}
	marker, err := envmarker.Read(ManagedHomeDir(req.Home, req.Profile, adapterID))
	if err != nil {
		return err
	}
	return req.observeOrigin(adapterID, marker)
}
func (req *ResolveRequest) credentialNativeRoot(adapterID string) (string, error) {
	if req.origin == nil || req.origin.record.Adapter != adapterID || req.origin.record.Binding.Profile != req.Profile {
		if err := req.refreshOrigin(adapterID); err != nil {
			return "", err
		}
	}
	return req.origin.record.NativeRoot, nil
}
func (req *ResolveRequest) originKey() string {
	if req.origin == nil {
		return ""
	}
	return envorigin.Digest(req.origin.raw)
}

// preflightOriginCredentials checks entry metadata/link text only. A missing
// credential leaf is backend-specific pending state, never a universal login
// gate. No native credential bytes, directory children or sessions are read.
// This gate precedes recovery, store repair, backups and native skill writes.
func (req *ResolveRequest) preflightOriginCredentials(adapter envregistry.Adapter) error {
	policy, err := req.resolveCredentialPolicyObserved(adapter, "unknown")
	if err != nil {
		return err
	}
	home := ManagedHomeDir(req.Home, req.Profile, adapter.ID)
	for path, target := range policy.links {
		full := filepath.Join(home, filepath.FromSlash(path))
		if adapter.ID == envregistry.Muse {
			if auth := req.inspectMuseAuth(full, target); auth.err != nil {
				return auth.err
			}
			continue
		}
		if _, err := req.inspectNativeCredentialTarget(adapter, path, target); err != nil {
			return err
		}
		state, err := req.credentialEntryMetadata(full)
		if err != nil {
			return fmt.Errorf("%s: credential entry metadata unavailable: %w", envregistry.DiagPassthroughUnreadable, err)
		}
		if state.Kind == stateread.KindAbsent {
			continue
		}
		if state.Kind != stateread.KindPresent || state.Info == nil {
			return fmt.Errorf("%s: credential entry metadata unavailable", envregistry.DiagPassthroughUnreadable)
		}
		if state.Info.IsDir() {
			entries, err := os.ReadDir(full)
			if err != nil {
				return fmt.Errorf("%s: credential directory metadata unavailable: %w", envregistry.DiagPassthroughUnreadable, err)
			}
			if len(entries) == 0 {
				continue
			}
		}
		if state.Info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("passthrough entry %s is detached: %s: credential entry must be absent, empty directory or its recorded native link", path, envregistry.DiagCredentialConflict)
		}
		link, err := req.credentialEntryLink(full)
		if err != nil {
			return fmt.Errorf("%s: credential link unreadable: %w", envregistry.DiagPassthroughUnreadable, err)
		}
		got := link.Target
		if !filepath.IsAbs(got) {
			got = filepath.Join(filepath.Dir(full), got)
		}
		if filepath.Clean(got) != filepath.Clean(target) {
			return fmt.Errorf("passthrough entry %s is detached: %s: credential link targets %s, expected recorded %s; migration needed: inspect with `curator env migrate --plan --profile %s --env %s`; refusing to re-point it here", full, envregistry.DiagCredentialConflict, got, target, req.Profile, adapter.ID)
		}
	}
	// A backend transition cannot silently orphan an existing recorded link.
	marker, err := envmarker.Read(home)
	if err != nil {
		return err
	}
	return removeStaleCredentialLinks(adapter, home, req.origin.record.NativeRoot, marker, policy.links,
		fmt.Sprintf("curator env migrate --plan --profile %s --env %s", req.Profile, adapter.ID))
}

// Inspection seams return only raw metadata/link text; production selects the
// verdict. No credential payload is opened, copied or included in a journal.
func (req *ResolveRequest) credentialEntryMetadata(path string) (stateread.Metadata, error) {
	if req.passthroughLstat != nil {
		return stateread.LstatWith(path, req.passthroughLstat)
	}
	return stateread.Lstat(path)
}
func (req *ResolveRequest) credentialEntryLink(path string) (stateread.Link, error) {
	if req.passthroughReadlink != nil {
		return stateread.ReadlinkWith(path, req.passthroughReadlink)
	}
	return stateread.Readlink(path)
}

// Pi shared requires a live native leaf (environments §7.4). Codex and Linux
// Claude retain login-pending compatibility; Muse has its separate strict gate.
func (req *ResolveRequest) inspectNativeCredentialTarget(adapter envregistry.Adapter, path, target string) (stateread.Metadata, error) {
	var metadata stateread.Metadata
	var err error
	if req.passthroughTargetStat != nil {
		metadata, err = stateread.StatWith(target, req.passthroughTargetStat)
	} else {
		metadata, err = stateread.Stat(target)
	}
	if err != nil || metadata.Kind == stateread.KindUnreadable || metadata.Kind == stateread.KindPresent && metadata.Info == nil {
		return metadata, fmt.Errorf("%s: passthrough entry %s target cannot be inspected (%s): %v", envregistry.DiagPassthroughUnreadable, path, envregistry.DiagCredentialConflict, err)
	}
	if adapter.ID == envregistry.Pi && metadata.Kind == stateread.KindAbsent {
		return metadata, fmt.Errorf("passthrough entry %s is detached: %s reason=dangling: recorded native target %s is absent; log in to the recorded native Pi agent root", path, envregistry.DiagCredentialConflict, target)
	}
	return metadata, nil
}
