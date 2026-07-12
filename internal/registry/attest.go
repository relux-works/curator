package registry

import (
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/identity"
	"github.com/relux-works/curator/internal/marker"
)

// AttestResult is one re-check outcome (Spec §13.3, status --attest).
type AttestResult struct {
	Scope    string
	Skill    string
	Result   string // audited | revoked | deprecated | unknown | no-registries | unattestable
	Registry string
	Detail   string
}

// AttestRoot re-resolves every install marker under a skills root against
// the trusted registries. It reads markers, not sources, so a revocation
// issued after install surfaces without reinstalling.
func AttestRoot(scope, skillsRoot string, registries []Registry, fetch FetchFn) []AttestResult {
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil
	}
	var results []AttestResult
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		recorded := marker.Read(filepath.Join(skillsRoot, entry.Name()))
		if recorded == nil {
			continue
		}
		if len(registries) == 0 {
			results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: "no-registries"})
			continue
		}
		if recorded.Commit == "" || recorded.ContentSHA256 == "" {
			results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: "unattestable", Detail: "marker lacks commit or hash"})
			continue
		}
		id := identity.Canonical(recorded.Git)
		if id == "" {
			results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: "unattestable", Detail: "no canonical source identity"})
			continue
		}
		resolution := Resolve(registries, id, recorded.Commit, recorded.ContentSHA256, fetch)
		registryName := ""
		if resolution.Attestation != nil {
			registryName = resolution.Attestation.Registry
		}
		results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: resolution.Result, Registry: registryName})
	}
	return results
}

// HasRevocation reports whether any result is a revocation.
func HasRevocation(results []AttestResult) bool {
	for _, result := range results {
		if result.Result == ResultRevoked {
			return true
		}
	}
	return false
}
