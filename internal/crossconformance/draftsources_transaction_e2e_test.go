package crossconformance

// Transaction integration for draft sources at the compiled CLI.
//
// This test proves a transaction PREPARE refusal through the real
// binary: the source changes, so the upgrade has replacement work,
// but the adapter-mirror parent is read-only, so staging the desired
// generation fails (prepare the install transaction: mkdir of the
// .curator-txn desired dir reports permission denied) before any
// replacement is staged. The previously installed generation must
// survive byte-identical. Post-prepare rollback is proven by
// write-boundary-retarget (a Go-API hook swaps the output after
// PointPrepared) and by the accepted install-package per-class sweep
// (TestDraftFailureAtEveryTargetClassRestoresPriorState), not here.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/marker"
)

func TestDraftSourcesCLIInstallRestoresPriorState(t *testing.T) {
	root := t.TempDir()
	configPath, project, home := setupCLIProject(t, root)
	ignored := ".agents/\nSkillfile.dev.json\n" + adapters.AgentPaths["codex_cli"] + "/\n"
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(ignored), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review")
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "project", "resolve", "app"); code != 0 {
		t.Fatalf("resolve = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if code, stdout, stderr := runCurator(t, home, configPath, nil, "install", "app"); code != 0 {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "review", "SKILL.md")
	before, err := os.ReadFile(installed)
	if err != nil {
		t.Fatalf("installed context missing: %v", err)
	}
	if recorded := marker.Read(filepath.Dir(installed)); recorded == nil || recorded.Name != "review" {
		t.Fatalf("marker = %+v, want the installed review record", recorded)
	}
	mirror := filepath.Join(project, adapters.AgentPaths["codex_cli"], "review", "SKILL.md")
	mirrorBefore, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatalf("adapter mirror missing: %v", err)
	}

	// The upgrade has genuine replacement work (the source changes)
	// but the transaction cannot prepare: the adapter parent is
	// read-only, so staging the desired generation fails. The refused
	// install must leave the prior generation intact.
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "info.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Atomic replacement renames through the parent directory, so
	// the parent — not the mirror itself — is what must refuse.
	mirrorParent := filepath.Dir(filepath.Dir(mirror))
	if err := os.Chmod(mirrorParent, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(mirrorParent, 0o755) }()
	probe := filepath.Join(mirrorParent, ".writability-probe")
	if err := os.WriteFile(probe, []byte("x"), 0o644); err == nil {
		_ = os.Remove(probe)
		t.Skip("this process can write through a read-only directory: permission failure not simulable")
	}
	code, _, stderr := runCurator(t, home, configPath, nil, "install", "app")
	if code == 0 {
		t.Fatal("install succeeded with a read-only mirror, want the prepare refusal")
	}
	t.Logf("refusal diagnostic: %s", firstLine(stderr))
	after, err := os.ReadFile(installed)
	if err != nil {
		t.Fatalf("installed context lost after failed install: %v", err)
	}
	if string(after) != string(before) {
		t.Fatal("installed context changed under a failed install")
	}
	mirrorAfter, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatalf("adapter mirror lost after failed install: %v", err)
	}
	if string(mirrorAfter) != string(mirrorBefore) {
		t.Fatal("adapter mirror changed under a failed install")
	}
	if recorded := marker.Read(filepath.Dir(installed)); recorded == nil || recorded.Name != "review" {
		t.Fatalf("marker after failure = %+v, want the prior record intact", recorded)
	}
}

func firstLine(text string) string {
	for i, r := range text {
		if r == '\n' {
			return text[:i]
		}
	}
	return text
}
