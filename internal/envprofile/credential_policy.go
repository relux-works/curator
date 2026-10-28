package envprofile

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/relux-works/curator/internal/envauthority"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// admitCredentialStore prevents a stored selector from silently becoming native
// on an older producer. This is an activation boundary, not runtime qualification.
// Muse file is admitted by D10; other forced mappings remain gated.
func admitCredentialStore(adapter envregistry.Adapter, profile string, selection envregistry.CredentialStoreSelection) error {
	if selection == envregistry.CredentialStoreNative || adapter.ID == envregistry.Muse && selection == envregistry.CredentialStoreFile {
		return nil
	}
	if adapter.ID == envregistry.CodexCLI && selection == envregistry.CredentialStoreFile {
		return fmt.Errorf("%s: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback", envregistry.DiagCodexSettingsUnsupported)
	}
	guidance := "this adapter has no qualified forced-store mapping"
	switch adapter.ID {
	case envregistry.ClaudeCode:
		guidance = "Claude has no established force-file or force-Keychain interface; log in from the same kind of session as the launcher, or explicitly enroll an environment-token source for inference"
	case envregistry.Muse:
		guidance = "managed Muse requires file storage; select native or file, then use TBH_CREDENTIAL_BACKEND=file for native file login"
	}
	return fmt.Errorf("%s: %s/%s selection %s: %s", envregistry.DiagCredentialStoreUnsupported, profile, adapter.ID, selection, guidance)
}

// effectiveCredentialPolicy is the operation-local projection for the existing
// managed-home credential contract. It is not a runtime qualification or a
// readiness verdict. Muse maps native to file here; other forced mappings
// retain activation refusals until qualification and strict consumers land.
type effectiveCredentialPolicy struct {
	selection      envregistry.CredentialStoreSelection
	effectiveStore envregistry.CredentialStoreSelection
	launchEnv      map[string]string
	isolation      string
	nativeStore    string
	links          map[string]string
	strategies     map[string]string
	records        []envmarker.Passthrough
}

// credentialPolicyResolution retains both successful and failed observations
// for one operation; a repair retry must not turn an observation failure into
// an absence or a different backend.
type credentialPolicyResolution struct {
	profile        string
	environment    string
	release        string
	originDigest   string
	lifecycleEpoch string
	policy         *effectiveCredentialPolicy
	err            error
}

// credentialPolicyRelease respects the content verifier's observation boundary.
func (req *ResolveRequest) credentialPolicyRelease(adapter envregistry.Adapter) string {
	if req.credentialRelease != nil {
		return *req.credentialRelease
	}
	return req.detected(adapter)
}

// resolveCredentialPolicy admits isolation and derives links and records from
// one native-store observation per release. New operations observe policy afresh.
func (req *ResolveRequest) resolveCredentialPolicy(adapter envregistry.Adapter) (*effectiveCredentialPolicy, error) {
	return req.resolveCredentialPolicyObserved(adapter, req.credentialPolicyRelease(adapter))
}

// resolveCredentialPolicyObserved uses supplied evidence, including unknown,
// without discovering a tool. Cache entries never cross release observations.
func (req *ResolveRequest) resolveCredentialPolicyObserved(adapter envregistry.Adapter, release string) (*effectiveCredentialPolicy, error) {
	epoch, epochErr := envauthority.Epoch(req.Home)
	if epochErr != nil {
		return nil, epochErr
	}
	if cached := req.credentialPolicy; cached != nil && cached.profile == req.Profile && cached.environment == adapter.ID && cached.release == release && cached.originDigest == req.originKey() && cached.lifecycleEpoch == epoch {
		return cached.policy, cached.err
	}
	policy, err := req.computeCredentialPolicy(adapter, release)
	req.credentialPolicy = &credentialPolicyResolution{profile: req.Profile, environment: adapter.ID, release: release, originDigest: req.originKey(), lifecycleEpoch: epoch, policy: policy, err: err}
	return policy, err
}

func (req *ResolveRequest) computeCredentialPolicy(adapter envregistry.Adapter, release string) (*effectiveCredentialPolicy, error) {
	selection, err := req.Machine.CredentialStoreFor(req.Profile, adapter.ID)
	if err != nil {
		return nil, err
	}
	p := &effectiveCredentialPolicy{
		selection:      selection,
		effectiveStore: selection,
		launchEnv:      map[string]string{},
		links:          map[string]string{}, strategies: map[string]string{}, records: []envmarker.Passthrough{},
	}
	var native string
	if adapter.ID == envregistry.CodexCLI {
		native, err = req.credentialNativeRoot(adapter.ID)
		if err != nil {
			return nil, err
		}
		p.nativeStore, err = req.codexCredentialStore(native)
		if err != nil {
			return nil, err
		}
	}
	if err := admitCredentialStore(adapter, req.Profile, selection); err != nil {
		return nil, err
	}
	if adapter.ID == envregistry.Muse {
		p.effectiveStore = envregistry.CredentialStoreFile
		p.launchEnv["TBH_CREDENTIAL_BACKEND"] = "file"
	}
	atAbove := true
	if ok, known := adapter.AtOrAbovePinned(release); known {
		atAbove = ok
	}
	p.isolation, err = req.Machine.EffectiveIsolation(req.Profile, adapter, atAbove)
	if err != nil {
		return nil, err
	}
	if adapter.ID == envregistry.CodexCLI && p.isolation == envregistry.IsolationIsolated && p.nativeStore != "file" {
		return nil, fmt.Errorf("%s: isolated is unsupported for codex_cli under native %s storage: the keyring identity is operator-global, independent of CODEX_HOME", envregistry.DiagIsolatedUnsupported, p.nativeStore)
	}
	if p.isolation == envregistry.IsolationIsolated {
		if adapter.ID == envregistry.ClaudeCode && runtime.GOOS == "darwin" {
			p.records = append(p.records, envmarker.Passthrough{
				Isolation: p.isolation, Strategy: envregistry.StrategyPerHomeKeychain,
				SourceRole: "managed", Backend: "keychain", BackendVersion: adapter.VerifiedRelease,
			})
		}
	} else {
		entries := adapter.PassthroughFor(runtime.GOOS)
		if len(entries) > 0 && adapter.VerifiedRelease == "" {
			return nil, fmt.Errorf("%s: %s has no verified credential backend release", envregistry.DiagCredentialUnsupported, adapter.ID)
		}
		if native == "" && len(entries) > 0 {
			native, err = req.credentialNativeRoot(adapter.ID)
			if err != nil {
				return nil, err
			}
		}
		for _, entry := range entries {
			record := envmarker.Passthrough{
				Isolation: p.isolation, Strategy: entry.Strategy, SourceRole: "native",
				Backend: "file", BackendVersion: adapter.VerifiedRelease,
			}
			switch entry.Strategy {
			case envregistry.StrategyFileLink:
				record.Path = entry.Path
			case envregistry.StrategyKeyringPreferred:
				if p.nativeStore == "keyring" {
					record.Backend = "ambient"
				} else {
					record.Path = entry.Path
				}
			case envregistry.StrategyAmbient:
				record.Backend = "ambient"
			case envregistry.StrategyPerHomeKeychain:
				record.SourceRole = "managed"
				record.Backend = "keychain"
			default:
				return nil, fmt.Errorf("%s: %s declares unsupported credential strategy %q", envregistry.DiagCredentialUnsupported, adapter.ID, entry.Strategy)
			}
			if record.Path != "" {
				p.links[record.Path] = filepath.Join(native, filepath.FromSlash(entry.FileLinkTarget))
				p.strategies[record.Path] = envregistry.StrategyFileLink
			}
			p.records = append(p.records, record)
		}
	}
	sort.Slice(p.records, func(i, j int) bool {
		if p.records[i].Path != p.records[j].Path {
			return p.records[i].Path < p.records[j].Path
		}
		return p.records[i].Strategy < p.records[j].Strategy
	})
	return p, nil
}

func (req *ResolveRequest) effectivePassthrough(adapter envregistry.Adapter) (map[string]string, map[string]string, error) {
	p, err := req.resolveCredentialPolicy(adapter)
	if err != nil {
		return nil, nil, err
	}
	return p.links, p.strategies, nil
}

func (req *ResolveRequest) credentialRecords(adapter envregistry.Adapter, provenance string) ([]envmarker.Passthrough, error) {
	p, err := req.resolveCredentialPolicy(adapter)
	if err != nil {
		return nil, err
	}
	records := append([]envmarker.Passthrough{}, p.records...)
	for i := range records {
		records[i].Provenance = provenance
	}
	return records, nil
}
