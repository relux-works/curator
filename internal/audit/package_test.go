package audit

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// packageTestCommit is one fixed full object id shared by the package
// cache fixtures below.
const packageTestCommit = "0123456789abcdef0123456789abcdef01234567"

// identifiedSubject seeds one audit subject stating the complete
// network-git package identity over the given snapshot.
func identifiedSubject(snapshot string) Subject {
	identity, ok := SkillPackage("github.com/example/role-skills", "", "backend", packageTestCommit, "skills/backend")
	if !ok {
		panic("test package identity is not statable")
	}
	return Subject{Name: "backend", Commit: packageTestCommit, Snapshot: snapshot,
		SchemaVersion: 3, Capabilities: capabilities.ImplicitNone(),
		Directory: "skills/backend", Package: identity}
}

// TestVerdictCacheBindsFullPackageIdentity pins the skillfile-sources §4
// audit-cache equality: a verdict cached for one package identity is a
// hit for that same identity and a miss for another directory,
// repository, or commit, even when the content hash matches — and a
// subject that states an identity never matches a record that states
// none.
func TestVerdictCacheBindsFullPackageIdentity(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("contact https://unlisted.example/")})
	digest, err := hashing.ContentSHA256(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	seeded := identifiedSubject(root)
	storeCachedFindings(cfg, digest, seeded, nil, nil)

	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, seeded); !hit {
		t.Fatalf("same-identity verdict missed")
	}
	same := seeded
	if report, err := auditSubject(cfg, same, false); err != nil || !report.CacheHit {
		t.Fatalf("same-identity audit = %+v, err = %v, want a cache hit", report, err)
	}

	probes := map[string]Subject{}
	movedDirectory := seeded
	movedDirectory.Directory = "skills/frontend"
	movedDirectory.Package.Directory = "skills/frontend"
	probes["directory"] = movedDirectory

	movedRepository := seeded
	movedRepository.Package.Repository = "github.com/example/other-skills"
	probes["repository"] = movedRepository

	movedCommit := seeded
	movedCommit.Commit = strings.Repeat("ab", 20)
	movedCommit.Package.Commit = &SourceCommit{ObjectFormat: "sha1", Hex: movedCommit.Commit}
	probes["commit"] = movedCommit

	otherKind := seeded
	otherKind.Package = SourcePackage{Kind: SourcePackageConfiguredGit, Source: "backend",
		Commit: &SourceCommit{ObjectFormat: "sha1", Hex: packageTestCommit}, Directory: "."}
	otherKind.Directory = "."
	probes["kind"] = otherKind

	unstated := seeded
	unstated.Package = SourcePackage{}
	probes["unstated"] = unstated

	for name, probe := range probes {
		if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, probe); hit {
			t.Fatalf("verdict cached for the seeded identity was reused for another %s", name)
		}
	}
	if report, err := auditSubject(cfg, probes["repository"], false); err != nil || report.CacheHit {
		t.Fatalf("other-repository audit = %+v, err = %v, want a fresh detection", report, err)
	}

	// A legacy record without a package member states no identity: an
	// identified subject misses it, while an unstated subject still hits.
	legacyDigest := "sha256:" + strings.Repeat("cc", 32)
	legacyPayload, err := json.Marshal(map[string]any{
		"schema_version": 1, "skill": "backend", "commit": packageTestCommit,
		"directory": "skills/backend", "findings": []Finding{},
	})
	if err != nil {
		t.Fatal(err)
	}
	legacyPath := verdictPath(cfg, legacyDigest)
	if err := os.MkdirAll(trustDir(cfg.Home(), legacyDigest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, legacyPayload, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, hit := loadCachedFindings(cfg, legacyDigest, hashing.VersionV1, seeded); hit {
		t.Fatalf("identified subject reused a legacy record without a package member")
	}
	if _, hit := loadCachedFindings(cfg, legacyDigest, hashing.VersionV1, unstated); !hit {
		t.Fatalf("unstated subject missed a legacy record without a package member")
	}

	// The local-snapshot arm binds its inventory digest the same way.
	snapshotSubject := Subject{Name: "local", Snapshot: root, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(),
		Package:      SourcePackage{Kind: SourcePackageLocalSnapshot, Snapshot: "sha256:" + strings.Repeat("aa", 32)}}
	storeCachedFindings(cfg, digest, snapshotSubject, nil, nil)
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, snapshotSubject); !hit {
		t.Fatalf("same local-snapshot verdict missed")
	}
	otherSnapshot := snapshotSubject
	otherSnapshot.Package.Snapshot = "sha256:" + strings.Repeat("bb", 32)
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, otherSnapshot); hit {
		t.Fatalf("verdict cached for one local snapshot was reused for another")
	}
}

// TestSkillPackageArms pins the audited package constructor: a canonical
// network identity records the network-git arm with the selected
// directory, a local source records the configured-git root arm with the
// schema-1 source default, and neither a malformed commit nor a
// subdirectory without a network identity states anything.
func TestSkillPackageArms(t *testing.T) {
	network, ok := SkillPackage("github.com/example/role-skills", "", "backend", packageTestCommit, "skills/backend")
	if !ok {
		t.Fatal("network-git identity was not statable")
	}
	if network.Kind != SourcePackageNetworkGit || network.Repository != "github.com/example/role-skills" ||
		network.Commit == nil || network.Commit.ObjectFormat != "sha1" || network.Commit.Hex != packageTestCommit ||
		network.Directory != "skills/backend" {
		t.Fatalf("network-git arm = %+v", network)
	}

	local, ok := SkillPackage("", "", "consumer", packageTestCommit, "")
	if !ok {
		t.Fatal("configured-git identity was not statable")
	}
	if local.Kind != SourcePackageConfiguredGit || local.Source != "consumer" || local.Directory != "." {
		t.Fatalf("configured-git arm = %+v", local)
	}

	sha256Commit := strings.Repeat("ab", 32)
	sha256Identity, ok := SkillPackage("github.com/example/role-skills", "", "backend", sha256Commit, "skills/backend")
	if !ok {
		t.Fatal("sha256 network-git identity was not statable")
	}
	if sha256Identity.Commit == nil || sha256Identity.Commit.ObjectFormat != "sha256" || sha256Identity.Commit.Hex != sha256Commit {
		t.Fatalf("sha256 arm = %+v", sha256Identity)
	}

	if _, ok := SkillPackage("github.com/example/role-skills", "", "backend", "not-a-commit", "skills/backend"); ok {
		t.Fatal("malformed commit stated a package identity")
	}
	if _, ok := SkillPackage("", "backend", "backend", packageTestCommit, "skills/backend"); ok {
		t.Fatal("subdirectory without a network identity stated a package identity")
	}
}
