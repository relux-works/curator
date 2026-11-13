package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/manifest"
)

// declareGit writes a schema-1 Skillfile declaring one skill at tag v1 with
// an explicit git source.
func (fixture upgradeFixture) declareGit(t *testing.T, alias, name, gitURL string) {
	t.Helper()
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"codex_cli"},
		"skills":         []map[string]any{{"name": name, "tag": "v1", "git": gitURL}},
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(fixture.projects[alias], manifest.Name), string(payload))
}

// Claim: `upgrade --all` binds every selected project's declaration to the
// shared checkout, even after an earlier project refreshed it: a later
// project declaring another source for the "direct" checkout fails exactly
// as it does when upgraded alone (source_origin_mismatch) and is not
// installed, while the earlier project is. Positive control: two projects
// with the same declaration both install and the checkout is fetched once.
func TestUpgradeAllBindsEveryProjectToTheSharedCheckout(t *testing.T) {
	t.Run("mismatching later project", func(t *testing.T) {
		fixture := newUpgradeFixture(t, "app", "other")
		upstream := filepath.Join(filepath.Dir(fixture.skillsRoot), "upstream")
		fixture.declareGit(t, "app", "direct", filepath.Join(upstream, "direct"))
		fixture.declareGit(t, "other", "direct", filepath.Join(upstream, "unrelated"))

		code, _, stderr := capture(t, fixture.configPath, "upgrade", "other")
		if code == exitOK || !strings.Contains(stderr, "source_origin_mismatch: direct declares") {
			t.Fatalf("upgrade other alone = %d, want source_origin_mismatch\nstderr:\n%s", code, stderr)
		}
		code, stdout, stderr := capture(t, fixture.configPath, "upgrade", "--all")
		if code == exitOK {
			t.Fatalf("upgrade --all accepted the mismatching project\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
		if !strings.Contains(stdout+stderr, "source_origin_mismatch: direct declares") {
			t.Fatalf("upgrade --all did not name the mismatch\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
		if _, err := os.Stat(filepath.Join(fixture.projects["other"], ".agents", "skills", "direct")); !os.IsNotExist(err) {
			t.Fatalf("the mismatching project was installed: %v", err)
		}
		if _, err := os.Stat(filepath.Join(fixture.projects["app"], ".agents", "skills", "direct")); err != nil {
			t.Fatalf("the matching project was not installed: %v", err)
		}
	})
	t.Run("identical declarations", func(t *testing.T) {
		fixture := newUpgradeFixture(t, "app", "other")
		upstream := filepath.Join(filepath.Dir(fixture.skillsRoot), "upstream")
		for _, alias := range []string{"app", "other"} {
			fixture.declareGit(t, alias, "direct", filepath.Join(upstream, "direct"))
		}
		code, stdout, stderr := capture(t, fixture.configPath, "upgrade", "--all")
		if code != exitOK {
			t.Fatalf("upgrade --all = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		for _, alias := range []string{"app", "other"} {
			if _, err := os.Stat(filepath.Join(fixture.projects[alias], ".agents", "skills", "direct")); err != nil {
				t.Fatalf("project %s was not installed: %v", alias, err)
			}
		}
		if fetches := countFetches(stdout, "direct"); fetches != 1 {
			t.Fatalf("direct fetched %d times, want 1\n%s", fetches, stdout)
		}
	})
}
