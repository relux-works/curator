package closureexec

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/closuregraph"
)

// Workspace is a task-private clean execution layout. Cache is derived state,
// never an intake authority. All writable roots start empty and disjoint.
type Workspace struct{ Root, Home, Config, Cache, Output, Temp string }

// PrepareWorkspace creates a new absent root and its empty owner-only children.
func PrepareWorkspace(root string) (Workspace, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Workspace{}, err
	}
	if _, err = os.Lstat(abs); err == nil {
		return Workspace{}, fmt.Errorf("workspace root already exists")
	}
	if !os.IsNotExist(err) {
		return Workspace{}, err
	}
	if err = os.Mkdir(abs, 0o700); err != nil {
		return Workspace{}, err
	}
	w := Workspace{Root: abs, Home: filepath.Join(abs, "home"), Config: filepath.Join(abs, "config"), Cache: filepath.Join(abs, "cache"), Output: filepath.Join(abs, "output"), Temp: filepath.Join(abs, "tmp")}
	for _, p := range []string{w.Home, w.Config, w.Cache, w.Output, w.Temp} {
		if err = os.Mkdir(p, 0o700); err != nil {
			return Workspace{}, err
		}
	}
	return w, nil
}

// DerivedCacheFile is one immutable observation of manager cache state built
// inside the task-private workspace from a receipted derivation.
type DerivedCacheFile struct {
	Path   string
	SHA256 closuregraph.ID
	Size   int64
}

// DerivedCacheReceipt proves that cache state is derived output, not capture
// provenance. The referenced derivation already binds admitted inputs,
// toolchain, environment, process/read/write policy, and network=none.
type DerivedCacheReceipt struct {
	DerivationReceiptID closuregraph.ID
	AssuranceMode       AssuranceMode
	PolicyID            string
	ExecutionPolicyID   string
	ProviderContract    *string
	Provider            *ProviderIdentity
	CapabilityReceiptID *closuregraph.ID
	ActualCapabilities  []CapabilityEvidence
	Files               []DerivedCacheFile
}

func (r DerivedCacheReceipt) canonicalValue() map[string]any {
	files := make([]any, len(r.Files))
	for i, f := range r.Files {
		files[i] = map[string]any{"path": f.Path, "sha256": string(f.SHA256), "size": f.Size}
	}
	return map[string]any{"actual_capabilities": capabilitiesValue(r.ActualCapabilities), "assurance_mode": string(r.AssuranceMode), "capability_receipt_sha256": optionalID(r.CapabilityReceiptID), "derivation_receipt_id": string(r.DerivationReceiptID), "execution_policy": r.ExecutionPolicyID, "files": files, "policy_id": r.PolicyID, "provider": optionalProvider(r.Provider), "provider_contract": optionalString(r.ProviderContract), "schema_id": "closure-derived-manager-cache-v2"}
}

// ID derives the exact admitted-input-bound manager-cache identity.
func (r DerivedCacheReceipt) ID() (closuregraph.ID, error) {
	if !r.DerivationReceiptID.Valid() {
		return "", fmt.Errorf("invalid derivation receipt identity")
	}
	if err := validateAssuranceBinding(r.AssuranceMode, r.PolicyID, r.ExecutionPolicyID, r.ProviderContract, r.Provider, r.CapabilityReceiptID, r.ActualCapabilities); err != nil {
		return "", err
	}
	for i, f := range r.Files {
		if err := portablePath(f.Path); err != nil {
			return "", err
		}
		if !f.SHA256.Valid() || f.Size < 0 || (i > 0 && r.Files[i-1].Path >= f.Path) {
			return "", fmt.Errorf("derived cache files must be valid, sorted, and unique")
		}
	}
	return closuregraph.DomainID("curator-derived-manager-cache-v2", r.canonicalValue())
}

// ValidateFor rejects reuse under another assurance mode or provider policy.
func (r DerivedCacheReceipt) ValidateFor(config AssuranceConfig, provider *ProviderIdentity) error {
	config = config.normalized()
	if err := config.validate(); err != nil {
		return err
	}
	if _, err := r.ID(); err != nil {
		return err
	}
	if r.AssuranceMode != config.Mode {
		return failure("assurance_evidence_mismatch", "cache mode differs from required policy")
	}
	if config.Mode == AssurancePortable {
		if provider != nil || r.Provider != nil {
			return failure("assurance_evidence_mismatch", "portable cache cannot carry provider authority")
		}
		return nil
	}
	if provider == nil || r.Provider == nil || *provider != *r.Provider || provider.validate(config) != nil {
		return failure("verified_execution_receipt_invalid", "cache provider differs from required provider")
	}
	return nil
}

// ObserveDerivedCache rechecks the actual private cache after its permitted
// construction and returns a canonical receipt. No ambient cache path enters.
func (w Workspace) ObserveDerivedCache(derivation DerivationReceipt) (DerivedCacheReceipt, error) {
	receiptID, err := derivation.ID()
	if err != nil {
		return DerivedCacheReceipt{}, err
	}
	for _, write := range derivation.Audit.Writes {
		if write != "cache" && write[:min(len(write), 6)] != "cache/" {
			return DerivedCacheReceipt{}, failure("closure_write_undeclared", "derivation wrote outside private cache")
		}
	}
	var files []DerivedCacheFile
	err = filepath.WalkDir(w.Cache, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == w.Cache {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return failure("closure_input_undeclared", "derived cache contains a link")
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return failure("closure_input_undeclared", "derived cache contains a special node")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- current is supplied by WalkDir rooted in the private cache.
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(w.Cache, current)
		if err != nil {
			return err
		}
		files = append(files, DerivedCacheFile{Path: filepath.ToSlash(rel), SHA256: digestBytes(payload), Size: int64(len(payload))})
		return nil
	})
	if err != nil {
		return DerivedCacheReceipt{}, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	result := DerivedCacheReceipt{DerivationReceiptID: receiptID, AssuranceMode: derivation.AssuranceMode, PolicyID: derivation.PolicyID, ExecutionPolicyID: derivation.ExecutionPolicyID, ProviderContract: derivation.ProviderContract, Provider: derivation.Provider, CapabilityReceiptID: derivation.CapabilityReceiptID, ActualCapabilities: append([]CapabilityEvidence(nil), derivation.ActualCapabilities...), Files: files}
	if _, err = result.ID(); err != nil {
		return DerivedCacheReceipt{}, err
	}
	return result, nil
}

// RecheckEmptyWritableRoots rejects ambient or poisoned state before use.
func (w Workspace) RecheckEmptyWritableRoots() error {
	for name, p := range map[string]string{"home": w.Home, "config": w.Config, "cache": w.Cache, "output": w.Output, "temp": w.Temp} {
		entries, err := os.ReadDir(p)
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return failure("closure_input_undeclared", "%s root is not empty", name)
		}
	}
	return nil
}
