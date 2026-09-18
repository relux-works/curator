package registry

import "fmt"

// Local package kinds admitted by draft source identities. Only packages
// with a network identity can carry a registry attestation; local content
// has no audit-record-v1 identity and the registry never forges one.
const (
	PackageLocalSnapshot = "local-snapshot"
	PackageNetworkGit    = "network-git"
	PackageConfiguredGit = "configured-git"
)

// NetworkAttestationRequired reports whether the registry policy demands a
// network attestation for every installed package. Strict mode admits no
// unattested package; advisory mode records whatever evidence resolves.
func NetworkAttestationRequired(registryPolicy string) bool {
	return registryPolicy == "strict"
}

// CheckLocalPackage enforces skillfile-sources §4 for a package with no
// network identity: where the registry policy requires a network
// attestation that local content cannot supply, installation fails. The
// refusal happens before any cache or compiler work at the call site. Git
// package kinds always pass here; their evidence is resolved and matched on
// the normal path, which never synthesizes records for local content.
func CheckLocalPackage(skill, registryPolicy, packageKind string) error {
	if packageKind == PackageLocalSnapshot && NetworkAttestationRequired(registryPolicy) {
		return fmt.Errorf(
			"%s is a local snapshot and registry_policy is strict: local content has no network attestation identity",
			skill)
	}
	return nil
}
