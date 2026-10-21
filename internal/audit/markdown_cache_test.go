package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/hashing"
)

// A real pre-Markdown Gate allowance cannot bypass either public gate after
// upgrade, in either hash framing. New clean and declared-host allowances still
// cache; old records remain byte-identical. This covers the detector's former
// scripts/manifest surface, not every historical detector implementation.
func TestReviewPriorMarkdownAllowance(t *testing.T) {
	const seedEnv = "CURATOR_TEST_PRIOR_MARKDOWN_CACHE"
	versions := []hashing.Version{hashing.VersionV1, hashing.VersionV2}
	if root := os.Getenv(seedEnv); root != "" {
		for _, version := range versions {
			cfg, subject := priorMarkdownFixture(t, root, version)
			if warnings, errs := Gate(cfg, []Subject{subject}); len(warnings) != 0 || len(errs) != 0 {
				t.Fatalf("pre-Markdown Gate seed refused: warnings=%v errors=%v", warnings, errs)
			}
			report, err := auditSubject(cfg, subject, false)
			if err != nil || !report.CacheHit || report.Decision != DecisionAllow || len(report.Findings) != 0 {
				t.Fatalf("pre-Markdown Gate did not persist an allowance: %+v, %v", report, err)
			}
		}
		return
	}

	root := t.TempDir()
	sourcePath, err := filepath.Abs("audit.go")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(payload), "\r\n", "\n")
	versionPattern := regexp.MustCompile(`RulesetVersion = "[0-9]+"`)
	markdown := ` && !markdown`
	if len(versionPattern.FindAllString(source, -1)) != 1 || strings.Count(source, markdown) != 1 {
		t.Fatal("pre-Markdown overlay anchors are not unique")
	}
	oldSource := versionPattern.ReplaceAllString(source, `RulesetVersion = "4"`)
	oldSource = strings.Replace(oldSource, markdown, "", 1)
	replacement := filepath.Join(root, "old-audit.go")
	if err := os.WriteFile(replacement, []byte(oldSource), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{sourcePath: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(root, "old.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", "^TestReviewPriorMarkdownAllowance$", ".")
	command.Env = append(os.Environ(), seedEnv+"="+root)
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "--- PASS: TestReviewPriorMarkdownAllowance") {
		t.Fatalf("pre-Markdown Gate seed failed: %v\n%s", err, output)
	}
	t.Logf("pre-Markdown Gate seed exit 0:\n%s", output)

	for _, version := range versions {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			cfg, subject := priorMarkdownFixture(t, root, version)
			digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, version)
			if err != nil {
				t.Fatal(err)
			}
			oldPath := filepath.Join(trustDir(cfg.Home(), digest), "verdict-"+strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, "4"}, "-")+".json")
			old, err := os.ReadFile(oldPath)
			if err != nil {
				t.Fatal(err)
			}
			assertBlock := func(warnings, errs []string) {
				t.Helper()
				if len(warnings) != 0 || len(errs) != 1 || !strings.Contains(errs[0], "audit.capability.network-undeclared") || !strings.Contains(errs[0], "references/notes.md") || !strings.Contains(errs[0], "undeclared.example.net") {
					t.Fatalf("prior ruleset-4 allowance bypassed current Markdown network gate: warnings=%v errors=%v", warnings, errs)
				}
			}
			assertBlock(GateReadOnly(cfg, []Subject{subject}))
			if _, err := os.Stat(verdictPath(cfg, digest)); !os.IsNotExist(err) {
				t.Fatalf("read-only gate wrote a current verdict: %v", err)
			}
			for _, wantHit := range []bool{false, true} {
				report, err := auditSubject(cfg, subject, false)
				if err != nil || report.CacheHit != wantHit || report.Decision != DecisionBlock {
					t.Fatalf("upgraded audit: %+v error=%v, want block, cache hit=%v", report, err, wantHit)
				}
				assertBlock(Gate(cfg, []Subject{subject}))
				assertBlock(GateReadOnly(cfg, []Subject{subject}))
			}
			if retained, err := os.ReadFile(oldPath); err != nil || !bytes.Equal(retained, old) {
				t.Fatalf("old allowance changed: %v", err)
			}

			for _, control := range []string{"clean", "declared-host"} {
				t.Run(control, func(t *testing.T) {
					cfg, subject := priorMarkdownFixture(t, t.TempDir(), version)
					if control == "clean" {
						if err := os.WriteFile(filepath.Join(subject.Snapshot, "references/notes.md"), []byte("Local formatting guidance.\n"), 0600); err != nil {
							t.Fatal(err)
						}
					} else {
						subject.Capabilities = capabilities.Manifest{Network: []string{"undeclared.example.net"}}
					}
					for _, wantHit := range []bool{false, true} {
						report, err := auditSubject(cfg, subject, false)
						if err != nil || report.CacheHit != wantHit || report.Decision != DecisionAllow || len(report.Findings) != 0 {
							t.Fatalf("fresh control: %+v error=%v, want allow, cache hit=%v", report, err, wantHit)
						}
						if warnings, errs := Gate(cfg, []Subject{subject}); len(warnings) != 0 || len(errs) != 0 {
							t.Fatalf("fresh control gate refused: warnings=%v errors=%v", warnings, errs)
						}
					}
				})
			}
		})
	}
}

func priorMarkdownFixture(t *testing.T, root string, version hashing.Version) (*config.Config, Subject) {
	t.Helper()
	root = filepath.Join(root, fmt.Sprintf("v%d", version))
	snapshot := filepath.Join(root, "snapshot")
	if err := os.MkdirAll(filepath.Join(snapshot, "references"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snapshot, "references/notes.md"), []byte("curl https://undeclared.example.net/x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := newCfg(t, "strict", "high")
	cfg.Path = filepath.Join(root, "home", "config.json")
	subject := Subject{
		Name: "formatter", Source: "example.com/skills", Git: "https://example.com/skills.git",
		Commit: strings.Repeat("1", 40), Snapshot: snapshot, SchemaVersion: 3,
		Capabilities: capabilities.ImplicitNone(), HashVersion: version, Directory: "skills/formatter",
	}
	var ok bool
	subject.Package, ok = SkillPackage(subject.Source, "", subject.Name, subject.Commit, subject.Directory)
	if !ok {
		t.Fatal("fixture package identity invalid")
	}
	return cfg, subject
}
