package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
)

// writeAttestMarker stages one installed skill directory holding a valid
// marker: the content hash always matches the staged bytes, so every row
// below classifies on identity, never on drift.
func writeAttestMarker(t *testing.T, root, name string, m *marker.Marker) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	m.Name = name
	if err := marker.Write(dir, m); err != nil {
		t.Fatalf("Write = %v", err)
	}
}

func v5AttestNetworkMarker() *marker.Marker {
	return &marker.Marker{
		Name:               "net",
		Package:            &marker.Package{Kind: "network-git", Repository: "git.example.com/skills/skill-a", Commit: &marker.Commit{ObjectFormat: "sha1", Hex: testCommit}, Directory: "skills/skill-a"},
		LockSHA256:         "sha256:" + strings.Repeat("c", 64),
		ContentSHA256:      testContentSHA256,
		Locale:             "en",
		Agents:             []string{"claude_code"},
		Commands:           []string{},
		Dependencies:       []string{},
		SkillSchemaVersion: 4,
		RuntimeRoots:       []string{},
		BuildRoots:         []string{},
		InstalledAt:        "2026-09-18T00:00:00Z",
		Files:              []string{"SKILL.md"},
		Builds:             map[string]marker.Build{},
		Activation:         &marker.Activation{Context: true, Commands: []string{}},
		Requirers:          []string{"<project>"},
	}
}

func v5AttestLocalMarker() *marker.Marker {
	m := v5AttestNetworkMarker()
	m.Package = &marker.Package{Kind: "local-snapshot", Snapshot: "sha256:" + strings.Repeat("a", 64)}
	return m
}

func v5AttestConfiguredMarker() *marker.Marker {
	m := v5AttestNetworkMarker()
	m.Package = &marker.Package{Kind: "configured-git", Source: "skill-a", Commit: &marker.Commit{ObjectFormat: "sha1", Hex: testCommit}, Directory: "."}
	return m
}

func attestResultsBySkill(results []AttestResult) map[string]AttestResult {
	bySkill := map[string]AttestResult{}
	for _, result := range results {
		bySkill[result.Skill] = result
	}
	return bySkill
}

// TestAttestRootV5NetworkGitResolvesThroughThePackage is the positive row:
// a network-git marker re-resolves through its canonical repository and
// locked commit exactly like the install lane's registry resolution.
func TestAttestRootV5NetworkGitResolvesThroughThePackage(t *testing.T) {
	s := newSigner(t)
	registries := []Registry{{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}}
	fetch := staticFetch(map[string][]map[string]any{
		"https://one": {s.sign(record(StatusAudited))},
	}, nil)
	root := t.TempDir()
	writeAttestMarker(t, root, "net", v5AttestNetworkMarker())
	results := attestResultsBySkill(AttestRoot("test", root, registries, fetch))
	got, ok := results["net"]
	if !ok {
		t.Fatalf("no result for net: %+v", results)
	}
	if got.Result != ResultAudited || got.Registry != "one" {
		t.Fatalf("result = %+v, want audited via one", got)
	}
	if HasRevocation(AttestRoot("test", root, registries, fetch)) {
		t.Fatalf("audited marker reported a revocation")
	}
}

// TestAttestRootV5RevocationSurfacesWithoutReinstall is the revocation
// row: a revocation issued after install surfaces from the marker alone.
func TestAttestRootV5RevocationSurfacesWithoutReinstall(t *testing.T) {
	s := newSigner(t)
	registries := []Registry{{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}}
	fetch := staticFetch(map[string][]map[string]any{
		"https://one": {s.sign(record(StatusRevoked))},
	}, nil)
	root := t.TempDir()
	writeAttestMarker(t, root, "net", v5AttestNetworkMarker())
	results := AttestRoot("test", root, registries, fetch)
	if len(results) != 1 || results[0].Result != ResultRevoked {
		t.Fatalf("results = %+v, want one revoked", results)
	}
	if !HasRevocation(results) {
		t.Fatalf("revoked marker reported no revocation")
	}
}

// TestAttestRootV5ArmsWithoutRegistryIdentityAreUnattestable is the
// negative row: local snapshots have no network identity and a
// configured source path alone is not a canonical registry identity, so
// neither arm can attest — and neither is reported as audited.
func TestAttestRootV5ArmsWithoutRegistryIdentityAreUnattestable(t *testing.T) {
	s := newSigner(t)
	registries := []Registry{{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}}
	fetch := staticFetch(map[string][]map[string]any{
		"https://one": {s.sign(record(StatusAudited))},
	}, nil)
	root := t.TempDir()
	writeAttestMarker(t, root, "local", v5AttestLocalMarker())
	writeAttestMarker(t, root, "cfg", v5AttestConfiguredMarker())
	results := attestResultsBySkill(AttestRoot("test", root, registries, fetch))
	local, ok := results["local"]
	if !ok || local.Result != "unattestable" || local.Detail != "local snapshot has no registry identity" {
		t.Fatalf("local = %+v, want unattestable with the local detail", local)
	}
	cfg, ok := results["cfg"]
	if !ok || cfg.Result != "unattestable" || cfg.Detail != "no canonical source identity" {
		t.Fatalf("cfg = %+v, want unattestable with the identity detail", cfg)
	}
}

// TestAttestRootLegacyMarkersUnchanged pins the frozen lane: a legacy
// marker attests through its parsed Git identity exactly as before.
func TestAttestRootLegacyMarkersUnchanged(t *testing.T) {
	s := newSigner(t)
	registries := []Registry{{Name: "one", URL: "https://one", PublicKeys: []string{s.pinned}}}
	fetch := staticFetch(map[string][]map[string]any{
		"https://one": {s.sign(record(StatusAudited))},
	}, nil)
	root := t.TempDir()
	writeAttestMarker(t, root, "legacy", &marker.Marker{
		Name: "legacy", Source: "skill-a", Git: "https://git.example.com/skills/skill-a",
		RefKind: "revision", Ref: testCommit, Commit: testCommit,
		ContentSHA256: testContentSHA256, Locale: "en",
		Agents: []string{"claude_code"}, Commands: []string{}, Dependencies: []string{},
		SkillSchemaVersion: 6, RuntimeRoots: []string{}, BuildRoots: []string{},
		InstalledAt: "2026-07-21T00:00:00Z", Files: []string{"SKILL.md"},
		Builds: map[string]marker.Build{},
	})
	results := attestResultsBySkill(AttestRoot("test", root, registries, fetch))
	got, ok := results["legacy"]
	if !ok {
		t.Fatalf("no result for legacy: %+v", results)
	}
	if got.Result != ResultAudited || got.Registry != "one" {
		t.Fatalf("result = %+v, want audited via one", got)
	}
}

func TestAttestRootKeepsInvalidMarkerDistinctFromAbsence(t *testing.T) {
	root := t.TempDir()
	installed := filepath.Join(root, "broken")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	if results := AttestRoot("test", root, nil, nil); len(results) != 0 {
		t.Fatalf("absent marker attestation = %+v, want no row", results)
	}
	path := filepath.Join(installed, marker.Name)
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	results := AttestRoot("test", root, nil, nil)
	if len(results) != 1 || results[0].Skill != "broken" || results[0].Result != ResultUnknown || !strings.Contains(results[0].Detail, marker.DiagInvalid) || strings.Contains(results[0].Detail, "manager_state_unreadable") {
		t.Fatalf("invalid marker attestation = %+v, want unknown with %s and no unreadable diagnostic", results, marker.DiagInvalid)
	}
}

// TestV5AttestIdentityRefusesUnprovableMarkers pins the helper seam:
// shapes the marker reader must refuse (empty repository, missing
// commit, missing hash) never attest, so a defect in the reader still
// fails closed here.
func TestV5AttestIdentityRefusesUnprovableMarkers(t *testing.T) {
	base := v5AttestNetworkMarker()
	cases := map[string]func(*marker.Marker){
		"empty-repository": func(m *marker.Marker) { m.Package.Repository = "" },
		"nil-commit":       func(m *marker.Marker) { m.Package.Commit = nil },
		"empty-commit":     func(m *marker.Marker) { m.Package.Commit.Hex = "" },
		"empty-hash":       func(m *marker.Marker) { m.ContentSHA256 = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := *base
			pkg := *base.Package
			m.Package = &pkg
			if base.Package.Commit != nil {
				commit := *base.Package.Commit
				m.Package.Commit = &commit
			}
			mutate(&m)
			if id, commit, detail := v5AttestIdentity(&m); detail == "" || id != "" || commit != "" {
				t.Fatalf("v5AttestIdentity = (%q, %q, %q), want a refusal", id, commit, detail)
			}
		})
	}
	id, commit, detail := v5AttestIdentity(base)
	if detail != "" || id != "git.example.com/skills/skill-a" || commit != testCommit {
		t.Fatalf("v5AttestIdentity = (%q, %q, %q), want the canonical passthrough", id, commit, detail)
	}
}
