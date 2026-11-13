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

// republishV1 moves tag v1 of a bare repository to a new commit whose
// SKILL.md carries marker, and adds the tag marker on that commit.
func republishV1(t *testing.T, bare, marker string) {
	t.Helper()
	work := filepath.Join(t.TempDir(), "work")
	runGit(t, filepath.Dir(work), "clone", "-q", bare, work)
	skill := filepath.Join(work, "SKILL.md")
	content, err := os.ReadFile(skill)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, skill, string(content)+"\n"+marker+"\n")
	runGit(t, work, "commit", "-qam", marker)
	runGit(t, work, "tag", "-f", "v1")
	runGit(t, work, "tag", marker)
	runGit(t, work, "push", "-q", "--force", "origin", "HEAD:refs/heads/main", "refs/tags/v1", "refs/tags/"+marker)
}

// Claim (review R2 witness, real CLI): `upgrade app` with the "direct"
// checkout's ordinary origin file://<root>/mirror, mirror -> a repository
// whose name ends in a space, and a distinct sibling without the space,
// fetches the symlink target: its fresh-intended tag arrives, the sibling's
// tag does not, and the installed skill is the target's v1 payload, not the
// sibling's; with no declared source and with the matching file://
// declaration. Control: an ordinary target name with a distinct sibling
// behaves the same.
func TestUpgradeFetchesTheExactPhysicalOrigin(t *testing.T) {
	for _, names := range []struct{ name, intended, sibling string }{
		{"final space", "direct.git ", "direct.git"},
		{"ordinary path control", "direct.git", "other.git"},
	} {
		for _, declared := range []bool{false, true} {
			label := names.name + "/undeclared"
			if declared {
				label = names.name + "/declared file URL"
			}
			t.Run(label, func(t *testing.T) {
				fixture := newUpgradeFixture(t, "app")
				root, err := filepath.EvalSymlinks(filepath.Dir(fixture.skillsRoot))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(filepath.Join(root, names.intended), 0o755); err != nil {
					t.Skipf("filesystem refuses %q: %v", names.intended, err)
				}
				if err := os.Remove(filepath.Join(root, names.intended)); err != nil {
					t.Fatal(err)
				}
				for name, marker := range map[string]string{names.intended: "fresh-intended", names.sibling: "only-in-trimmed-sibling"} {
					bare := filepath.Join(root, name)
					runGit(t, root, "clone", "-q", "--bare", filepath.Join(root, "upstream", "direct"), bare)
					republishV1(t, bare, marker)
				}
				mirror := filepath.Join(root, "mirror")
				if err := os.Symlink(filepath.Join(root, names.intended), mirror); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
				mirrorURL := "file://" + filepath.ToSlash(mirror)
				if filepath.VolumeName(mirror) != "" {
					mirrorURL = "file:///" + filepath.ToSlash(mirror)
				}
				checkout := filepath.Join(fixture.skillsRoot, "direct")
				runGit(t, checkout, "remote", "set-url", "origin", mirrorURL)
				if declared {
					fixture.declareGit(t, "app", "direct", mirrorURL)
				} else {
					fixture.declare(t, "app", "direct")
				}

				code, stdout, stderr := capture(t, fixture.configPath, "upgrade", "app")
				if code != exitOK {
					t.Fatalf("upgrade app = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
				}
				if fetches := countFetches(stdout, "direct"); fetches != 1 {
					t.Fatalf("direct fetched %d times, want 1\n%s", fetches, stdout)
				}
				tags := runGitOutput(t, checkout, "tag", "-l")
				if !strings.Contains(tags, "fresh-intended") || strings.Contains(tags, "only-in-trimmed-sibling") {
					t.Fatalf("tags %q: want the symlink target's tag and not the sibling's", tags)
				}
				installed, err := os.ReadFile(filepath.Join(fixture.projects["app"], ".agents", "skills", "direct", "SKILL.md"))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(installed), "fresh-intended") || strings.Contains(string(installed), "only-in-trimmed-sibling") {
					t.Fatalf("installed SKILL.md is not the symlink target's v1 payload:\n%s", installed)
				}
			})
		}
	}
}
