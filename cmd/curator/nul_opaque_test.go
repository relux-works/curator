package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
)

func pinAuditHashWriters(t *testing.T, v2 bool) {
	t.Helper()
	priorWriter := hashing.EnableV2Writers
	hashing.EnableV2Writers = v2
	t.Cleanup(func() { hashing.EnableV2Writers = priorWriter })
}

// The opaque-NUL interim rule (Spec §8) follows the audited identity: the
// external-repository lane blocks a NUL-bearing snapshot under v1 and
// admits it under v2, where NUL bytes are ordinary data.
func TestProductionExternalAuditOpaqueNULFollowsHashVersion(t *testing.T) {
	root := t.TempDir()
	const rel = "assets/deep/opaque.unsupported"
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("prefix\x00suffix"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Path: filepath.Join(t.TempDir(), "config.json")}
	if cfg.Audit.Enabled {
		t.Fatal("test requires the default disabled audit configuration")
	}
	subject := buildrepo.AuditSubject{
		Declared:     buildrepo.DeclaredState{Repository: "tools", Identity: "example.test/tools"},
		Effective:    buildrepo.EffectiveState{Commit: strings.Repeat("1", 40)},
		SnapshotRoot: root,
	}

	t.Run("v1-blocks", func(t *testing.T) {
		pinAuditHashWriters(t, false)
		for _, dryRun := range []bool{false, true} {
			t.Run(map[bool]string{false: "install", true: "dry-run"}[dryRun], func(t *testing.T) {
				deps := productionExternalDeps(cfg, dryRun)
				_, err := deps.AuditWarnings(context.Background(), subject)
				if err == nil || !strings.Contains(err.Error(), "audit.opaque.nul-byte") || !strings.Contains(err.Error(), rel) {
					t.Fatalf("production external audit error = %v, want blocking opaque finding naming %s", err, rel)
				}
			})
		}
	})

	t.Run("v2-admits", func(t *testing.T) {
		pinAuditHashWriters(t, true)
		for _, dryRun := range []bool{false, true} {
			t.Run(map[bool]string{false: "install", true: "dry-run"}[dryRun], func(t *testing.T) {
				deps := productionExternalDeps(cfg, dryRun)
				warnings, err := deps.AuditWarnings(context.Background(), subject)
				if err != nil || len(warnings) != 0 {
					t.Fatalf("v2 production external audit warnings=%v err=%v, want silent admission", warnings, err)
				}
			})
		}
	})
}

// `curator audit` reports what an install would trust: the blocking opaque
// finding under v1, a passing audit under v2.
func TestCLIAuditOpaqueNULFollowsHashVersion(t *testing.T) {
	build := func(t *testing.T) string {
		t.Helper()
		configPath := auditLabelProject(t, 7, false, nil)
		cfg, err := config.Load(configPath, nil)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Audit.Enabled {
			t.Fatal("test requires the default disabled audit configuration")
		}
		repo := filepath.Join(cfg.SkillsRoot, "label-skill")
		const rel = "docs/deep/opaque.unknown"
		path := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("prefix\x00suffix"), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit(t, repo, "add", ".")
		runGit(t, repo, "commit", "-qm", "add opaque audit file")
		runGit(t, repo, "tag", "-f", "v1")
		return configPath
	}

	t.Run("v1-blocks", func(t *testing.T) {
		pinAuditHashWriters(t, false)
		configPath := build(t)
		code, stdout, stderr := capture(t, configPath, "audit", "app")
		combined := stdout + "\n" + stderr
		if code == exitOK || !strings.Contains(combined, "audit.opaque.nul-byte") || !strings.Contains(combined, "docs/deep/opaque.unknown") {
			t.Fatalf("curator audit = %d\nstdout:\n%s\nstderr:\n%s; want blocking opaque finding naming %s", code, stdout, stderr, "docs/deep/opaque.unknown")
		}
	})

	t.Run("v2-admits", func(t *testing.T) {
		pinAuditHashWriters(t, true)
		configPath := build(t)
		code, stdout, stderr := capture(t, configPath, "audit", "app")
		combined := stdout + "\n" + stderr
		if code != exitOK || strings.Contains(combined, "audit.opaque.nul-byte") {
			t.Fatalf("v2 curator audit = %d\nstdout:\n%s\nstderr:\n%s; want a passing audit", code, stdout, stderr)
		}
	})
}
