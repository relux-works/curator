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
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/whitelist"
)

func caseInsensitiveSnapshot(t *testing.T, root string) bool {
	t.Helper()
	path := filepath.Join(root, "case-probe")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	_, err := os.Stat(filepath.Join(root, "CASE-PROBE"))
	return err == nil
}

// On case-insensitive hosts, each root alias that CopyContext publishes is
// refused by both gates, including cached verdicts. A declared host admits the
// same bytes. On case-sensitive hosts, directory aliases remain out of context;
// a missing canonical SKILL.md is instead refused by CopyContext itself.
func TestMarkdownScopeCaseAliases(t *testing.T) {
	for _, tc := range []struct{ alias, canonical string }{
		{"SKILL.MD", "SKILL.md"}, {"References/guide.md", "references/guide.md"},
		{"AGENTS/guide.md", "agents/guide.md"}, {"ASSETS/guide.md", "assets/guide.md"},
		{".SKILL_TRIGGERS/guide.md", ".skill_triggers/guide.md"}, {"Templates/guide.md", "templates/guide.md"},
		{"Examples/guide.md", "examples/guide.md"}, {"Data/guide.md", "data/guide.md"}, {"Scripts/guide.md", "scripts/guide.md"},
	} {
		t.Run(tc.alias, func(t *testing.T) {
			subject := markdownScopeSubject(t, nil, "")
			insensitive := caseInsensitiveSnapshot(t, subject.Snapshot)
			if tc.alias == "SKILL.MD" {
				if err := os.Rename(filepath.Join(subject.Snapshot, "SKILL.md"), filepath.Join(subject.Snapshot, tc.alias)); err != nil {
					t.Fatal(err)
				}
			}
			writeFiles(t, subject.Snapshot, map[string][]byte{tc.alias: []byte(markdownContact)})
			destination := filepath.Join(t.TempDir(), "context")
			copied, err := whitelist.CopyContext(subject.Snapshot, destination, true, nil)
			if !insensitive && tc.alias == "SKILL.MD" {
				if err == nil {
					t.Fatal("case-sensitive copy admitted a missing canonical SKILL.md")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if insensitive {
				want = tc.canonical
				payload, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(tc.canonical)))
				if err != nil || string(payload) != markdownContact {
					t.Fatalf("alias copy witness missing: %q, %v; copied=%v", payload, err, copied)
				}
			} else if _, err := os.Stat(filepath.Join(destination, filepath.FromSlash(tc.canonical))); !os.IsNotExist(err) {
				t.Fatalf("case-sensitive root alias copied: %v", copied)
			}
			assertMarkdownScope(t, subject, want)
			subject.Capabilities = capabilities.Manifest{Network: []string{"api.example.net"}}
			assertMarkdownScope(t, subject, "")
		})
	}
}

// Every fixed exclusion is case-sensitive at depth: its canonical spelling
// drops the contact, while the neighboring case variant is copied and audited.
// This includes readme.md, ReadMe.MD and Tests, which are not spec exclusions.
func TestMarkdownScopeExclusionCaseParity(t *testing.T) {
	for _, tc := range []struct{ excluded, visible string }{
		{".git", ".GIT"}, {".github", ".GITHUB"}, {".gitlab-ci.yml", ".GITLAB-CI.YML"},
		{".venv", ".VENV"}, {"__pycache__", "__PYCACHE__"}, {"code.pyc", "code.PYC"},
		{"node_modules", "NODE_MODULES"}, {"tests", "Tests"}, {"test", "Test"}, {"__tests__", "__TESTS__"},
		{"README.md", "readme.md"}, {"README.md", "ReadMe.MD"}, {"CHANGELOG.md", "changelog.md"},
		{"LICENSE.md", "license.md"}, {"Makefile", "MAKEFILE"}, {"setup.py", "SETUP.PY"},
		{"pyproject.toml", "PYPROJECT.TOML"}, {"requirements-dev.txt", "REQUIREMENTS-dev.TXT"},
		{".DS_Store", ".ds_store"}, {".gitignore", ".GITIGNORE"},
	} {
		for _, variant := range []struct {
			name    string
			visible bool
		}{{tc.excluded, false}, {tc.visible, true}} {
			t.Run(variant.name, func(t *testing.T) {
				file := "references/deep/" + variant.name + "/guide.md"
				subject := markdownScopeSubject(t, map[string]string{file: markdownContact}, "")
				destination := filepath.Join(t.TempDir(), "context")
				if _, err := whitelist.CopyContext(subject.Snapshot, destination, true, nil); err != nil {
					t.Fatal(err)
				}
				payload, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(file)))
				want := ""
				if variant.visible {
					if err != nil || string(payload) != markdownContact {
						t.Fatalf("case variant not copied: %q, %v", payload, err)
					}
					want = file
				} else if !os.IsNotExist(err) {
					t.Fatalf("excluded contact copied: %q, %v", payload, err)
				}
				assertMarkdownScope(t, subject, want)
			})
		}
	}
}

// Excluded-name case variants used as Markdown leaves, not just ancestor
// directories, are copied and blocked; canonical README.md remains excluded.
func TestMarkdownScopeExclusionCaseLeaves(t *testing.T) {
	for _, name := range []string{"README.md", "readme.md", "ReadMe.MD"} {
		t.Run(name, func(t *testing.T) {
			file := "references/" + name
			subject := markdownScopeSubject(t, map[string]string{file: markdownContact}, "")
			destination := filepath.Join(t.TempDir(), "context")
			if _, err := whitelist.CopyContext(subject.Snapshot, destination, true, nil); err != nil {
				t.Fatal(err)
			}
			payload, err := os.ReadFile(filepath.Join(destination, filepath.FromSlash(file)))
			want := file
			if name == "README.md" {
				want = ""
				if !os.IsNotExist(err) {
					t.Fatalf("excluded README leaf copied: %q, %v", payload, err)
				}
			} else if err != nil || string(payload) != markdownContact {
				t.Fatalf("visible case variant leaf missing: %q, %v", payload, err)
			}
			assertMarkdownScope(t, subject, want)
		})
	}
}

// rulesetSixSource restores the actual ruleset-6 enumeration algorithm: walk
// physical directory entries, then apply the lexical context predicate. The
// detector, Gate, hashing and cache writer remain production code.
func rulesetSixSource(t *testing.T, source string) string {
	t.Helper()
	start := strings.Index(source, "\tif kind == contextlock.KindSkill {")
	if start < 0 {
		t.Fatal("ruleset-6 walk anchor missing")
	}
	end := strings.Index(source[start:], "\n\t// Preserve the historical non-Markdown") + start
	if end < start {
		t.Fatal("ruleset-6 walk anchors missing")
	}
	return source[:start] + `
 _ = filepath.WalkDir(snapshot, func(path string, entry os.DirEntry, err error) error {
  if err != nil || entry.IsDir() { return nil }
  rel, err := filepath.Rel(snapshot, path)
  if err != nil { return nil }
  posix := filepath.ToSlash(rel)
  if !strings.EqualFold(filepath.Ext(posix), ".md") || !whitelist.ContextPathIncluded(posix, includeScripts, excludeRoots) { return nil }
  return scan(path, posix)
 })` + source[end:]
}

// A real Gate-generated ruleset-6 allowance for References/guide.md cannot
// bypass the corrected gates in either hash framing. Read-only evaluation
// writes nothing, new refusals cache, and the old record remains unchanged.
func TestMarkdownScopeInvalidatesRulesetSixAllowance(t *testing.T) {
	const seedEnv = "CURATOR_TEST_RULESET_SIX_CACHE"
	seed := os.Getenv(seedEnv)
	root := seed
	if root == "" {
		root = t.TempDir()
	}
	if !caseInsensitiveSnapshot(t, root) {
		t.Skip("ruleset-6 alias allowance requires a case-insensitive filesystem")
	}
	if seed == "" {
		sourcePath, err := filepath.Abs("audit.go")
		if err != nil {
			t.Fatal(err)
		}
		payload, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		source := rulesetSixSource(t, string(payload))
		versionPattern := regexp.MustCompile(`RulesetVersion = "[0-9]+"`)
		if len(versionPattern.FindAllString(source, -1)) != 1 {
			t.Fatal("ruleset version anchor is not unique")
		}
		source = versionPattern.ReplaceAllString(source, `RulesetVersion = "6"`)
		replacement := filepath.Join(root, "ruleset-six.go")
		if err := os.WriteFile(replacement, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{sourcePath: replacement}})
		if err != nil {
			t.Fatal(err)
		}
		overlayPath := filepath.Join(root, "six.json")
		if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
			t.Fatal(err)
		}
		command := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", "^TestMarkdownScopeInvalidatesRulesetSixAllowance$", ".")
		command.Env = append(os.Environ(), seedEnv+"="+root)
		output, err := command.CombinedOutput()
		if err != nil || !strings.Contains(string(output), "--- PASS: TestMarkdownScopeInvalidatesRulesetSixAllowance") {
			t.Fatalf("ruleset-6 production seed failed: %v\n%s", err, output)
		}
		t.Logf("real ruleset-6 Gate seed exit 0:\n%s", output)
	}
	for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			path := filepath.Join(root, fmt.Sprintf("v%d", version))
			snapshot := filepath.Join(path, "snapshot")
			if err := os.MkdirAll(snapshot, 0700); err != nil {
				t.Fatal(err)
			}
			writeFiles(t, snapshot, map[string][]byte{"SKILL.md": []byte("# Guide\n"), "References/guide.md": []byte(markdownContact), "agent-skill.json": []byte(`{"schema_version":3,"capabilities":{}}`)})
			subject := Subject{Name: "guide", Snapshot: snapshot, SchemaVersion: 3, HashVersion: version, Capabilities: capabilities.ImplicitNone()}
			cfg := newCfg(t, "strict", "high")
			cfg.Path = filepath.Join(path, "home", "config.json")
			if seed != "" {
				warnings, errs := Gate(cfg, []Subject{subject})
				assertMarkdownVerdict(t, warnings, errs, "")
				report, err := auditSubject(cfg, subject, false)
				if err != nil || !report.CacheHit || report.Decision != DecisionAllow || len(report.Findings) != 0 {
					t.Fatalf("ruleset-6 allowance was not really cached: %+v, %v", report, err)
				}
				return
			}
			digest, err := hashing.ContentSHA256WithVersion(snapshot, nil, version)
			if err != nil {
				t.Fatal(err)
			}
			oldPath := filepath.Join(trustDir(cfg.Home(), digest), "verdict-"+strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, "6"}, "-")+".json")
			old, err := os.ReadFile(oldPath)
			if err != nil {
				t.Fatal(err)
			}
			warnings, errs := GateReadOnly(cfg, []Subject{subject})
			assertMarkdownVerdict(t, warnings, errs, "references/guide.md")
			if verdictPath(cfg, digest) != oldPath {
				if _, err := os.Stat(verdictPath(cfg, digest)); !os.IsNotExist(err) {
					t.Fatalf("read-only gate wrote verdict: %v", err)
				}
			}
			for _, hit := range []bool{false, true} {
				report, err := auditSubject(cfg, subject, false)
				if err != nil || report.CacheHit != hit || report.Decision != DecisionBlock {
					t.Fatalf("ruleset-6 allowance reused: %+v, %v", report, err)
				}
				warnings, errs = Gate(cfg, []Subject{subject})
				assertMarkdownVerdict(t, warnings, errs, "references/guide.md")
			}
			retained, err := os.ReadFile(oldPath)
			if err != nil || !bytes.Equal(old, retained) {
				t.Fatalf("ruleset-6 allowance mutated: %v", err)
			}
		})
	}
}
