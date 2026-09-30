package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
)

func TestProductionExternalAuditBlocksOpaqueNULWhenAuditIsDisabled(t *testing.T) {
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

	for _, dryRun := range []bool{false, true} {
		t.Run(map[bool]string{false: "install", true: "dry-run"}[dryRun], func(t *testing.T) {
			deps := productionExternalDeps(cfg, dryRun)
			_, err := deps.AuditWarnings(context.Background(), buildrepo.AuditSubject{
				Declared:     buildrepo.DeclaredState{Repository: "tools", Identity: "example.test/tools"},
				Effective:    buildrepo.EffectiveState{Commit: strings.Repeat("1", 40)},
				SnapshotRoot: root,
			})
			if err == nil || !strings.Contains(err.Error(), "audit.opaque.nul-byte") || !strings.Contains(err.Error(), rel) {
				t.Fatalf("production external audit error = %v, want blocking opaque finding naming %s", err, rel)
			}
		})
	}
}

func TestCLIAuditBlocksOpaqueNULWhenAuditIsDisabled(t *testing.T) {
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

	code, stdout, stderr := capture(t, configPath, "audit", "app")
	combined := stdout + "\n" + stderr
	if code == exitOK || !strings.Contains(combined, "audit.opaque.nul-byte") || !strings.Contains(combined, rel) {
		t.Fatalf("curator audit = %d\nstdout:\n%s\nstderr:\n%s; want blocking opaque finding naming %s", code, stdout, stderr, rel)
	}
}
