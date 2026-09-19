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
		if recorded.Package != nil {
			// Draft schema-5 marker: the legacy source identity fields
			// are absent by construction, so attestation re-resolves
			// through the locked package (skillfile-sources §4). Only
			// the network-git arm carries a canonical registry
			// identity; a configured source path alone is not one, and
			// a local snapshot has no network identity at all.
			id, commit, detail := v5AttestIdentity(recorded)
			if detail != "" {
				results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: "unattestable", Detail: detail})
				continue
			}
			resolution := Resolve(registries, id, commit, recorded.ContentSHA256, fetch)
			registryName := ""
			if resolution.Attestation != nil {
				registryName = resolution.Attestation.Registry
			}
			results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: resolution.Result, Registry: registryName})
			continue
		}
		if recorded.Commit == "" || recorded.ContentSHA256 == "" {
			results = append(results, AttestResult{Scope: scope, Skill: recorded.Name, Result: "unattestable", Detail: "marker lacks commit or hash"})
			continue
		}
		id, identityErr := identity.Parse(recorded.Git)
		if identityErr != nil || id == "" {
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

// v5AttestIdentity resolves the registry identity and commit a draft
// marker attests. It returns a non-empty detail — and no identity — when
// the marker cannot attest: the local-snapshot and configured-git arms
// carry no canonical network identity, and a network-git marker without a
// parseable repository, a commit, or a content hash proves nothing.
func v5AttestIdentity(recorded *marker.Marker) (id, commit, detail string) {
	pkg := recorded.Package
	// The arms are the frozen source-types-v1 vocabulary (literals, as in
	// the marker reader: importing the lock package here would cycle
	// through its test dependencies).
	switch pkg.Kind {
	case "network-git":
		// The repository is already the canonical identity install
		// resolved and the marker reader validated (never a URL to
		// parse); an empty one proves nothing.
		if pkg.Repository == "" || pkg.Commit == nil || pkg.Commit.Hex == "" || recorded.ContentSHA256 == "" {
			return "", "", "marker lacks commit or hash"
		}
		return pkg.Repository, pkg.Commit.Hex, ""
	case "configured-git":
		return "", "", "no canonical source identity"
	default:
		return "", "", "local snapshot has no registry identity"
	}
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
