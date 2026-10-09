package audit

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
)

// TestVerdictCacheBindsSelectedDirectory pins the core §4.4 audit-cache
// key: a verdict cached for one selected package directory is a miss for
// another directory, even when the content hash matches, so an audit
// result for one directory is never reused for another merely because
// repository and commit match.
func TestVerdictCacheBindsSelectedDirectory(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("contact https://unlisted.example/")})
	digest, err := hashing.ContentSHA256(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	seeded := Subject{Name: "seeded", Snapshot: root, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), Directory: "skills/backend"}
	storeCachedFindings(cfg, digest, seeded, nil, nil)
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, seeded); !hit {
		t.Fatalf("same-directory verdict missed")
	}
	for _, other := range []string{"skills/frontend", "."} {
		probe := seeded
		probe.Directory = other
		if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, probe); hit {
			t.Fatalf("verdict cached for skills/backend was reused for %q", other)
		}
	}
	other := Subject{Name: "other", Snapshot: root, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), Directory: "skills/frontend"}
	report, err := auditSubject(cfg, other, false)
	if err != nil {
		t.Fatal(err)
	}
	if report.CacheHit || len(report.Findings) == 0 {
		t.Fatalf("report %+v, want a fresh detection, not the cached verdict of another directory", report)
	}
}

// TestVerdictCacheLegacyRecordWithoutDirectoryBindsRoot proves records
// that predate the directory member keep binding the repository root:
// a root subject still hits them, a subdirectory subject misses.
func TestVerdictCacheLegacyRecordWithoutDirectoryBindsRoot(t *testing.T) {
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{"scripts/tool": []byte("contact https://unlisted.example/")})
	digest, err := hashing.ContentSHA256(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "advisory", "off")
	dir := trustDir(cfg.Home(), digest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"schema_version": 1, "findings": []Finding{}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(verdictPath(cfg, digest), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	rootSubject := Subject{Name: "root", Snapshot: root, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), Directory: "."}
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, rootSubject); !hit {
		t.Fatalf("root subject missed a legacy record without a directory member")
	}
	subdir := rootSubject
	subdir.Directory = "skills/backend"
	if _, hit := loadCachedFindings(cfg, digest, hashing.VersionV1, subdir); hit {
		t.Fatalf("subdirectory subject reused a legacy root verdict")
	}
}
