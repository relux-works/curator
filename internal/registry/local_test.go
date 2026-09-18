package registry

import (
	"strings"
	"testing"
)

func TestCheckLocalPackage(t *testing.T) {
	// A strict registry policy requires a network attestation that local
	// content cannot supply: installation must fail, not pass unattested.
	if err := CheckLocalPackage("review", "strict", PackageLocalSnapshot); err == nil ||
		!strings.Contains(err.Error(), "no network attestation") {
		t.Fatalf("strict local err = %v", err)
	}
	// Narrowing the policy to advisory admits the same package: the refusal
	// is the policy bound, not the package.
	if err := CheckLocalPackage("review", "advisory", PackageLocalSnapshot); err != nil {
		t.Fatalf("advisory local err = %v", err)
	}
	// Git kinds always pass here; their evidence resolves on the normal
	// path with exact name/repository/commit/context matching.
	for _, kind := range []string{PackageNetworkGit, PackageConfiguredGit} {
		if err := CheckLocalPackage("review", "strict", kind); err != nil {
			t.Fatalf("strict %s err = %v", kind, err)
		}
	}
}

func TestResolveWithoutIdentityIsUnknown(t *testing.T) {
	// Even if a record's content hash coincides with the query hash, a
	// query with no network identity and no commit resolves unknown: the
	// registry never forges an attestation for local content.
	fetch := func(_, _, _, _ string) ([]map[string]any, error) {
		return []map[string]any{{
			"schema_version":  1,
			"name":            "review",
			"source_identity": "example.org/kit",
			"commit":          "0123456789abcdef0123456789abcdef01234567",
			"content_sha256":  "sha256:1111111111111111111111111111111111111111111111111111111111111111",
			"status":          StatusAudited,
			"sig": map[string]any{
				"algorithm": "ed25519",
				"key_id":    "0123456789abcdef",
				"signature": strings.Repeat("A", 88),
			},
		}}, nil
	}
	registries := []Registry{{Name: "test-reg", URL: "https://example.org", PublicKeys: []string{}}}
	resolution := Resolve(registries, "", "",
		"sha256:1111111111111111111111111111111111111111111111111111111111111111", fetch)
	if resolution.Result != ResultUnknown || resolution.Attestation != nil {
		t.Fatalf("identity-less query resolved: %+v", resolution)
	}
}
