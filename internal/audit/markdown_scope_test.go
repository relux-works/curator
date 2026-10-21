package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/skillspec"
)

const markdownContact = "Read https://api.example.net/guide\n"

func markdownScopeSubject(t *testing.T, files map[string]string, manifest string) Subject {
	t.Helper()
	snapshot := t.TempDir()
	filesCopy := map[string][]byte{"SKILL.md": []byte("# Guide\n")}
	for path, text := range files {
		filesCopy[path] = []byte(text)
	}
	if manifest == "" && filesCopy["agent-skill.json"] == nil && filesCopy["csk-skill.json"] == nil {
		manifest = `{"schema_version":3,"capabilities":{}}`
	}
	if manifest != "" {
		filesCopy["agent-skill.json"] = []byte(manifest)
	}
	writeFiles(t, snapshot, filesCopy)
	spec, err := skillspec.Load(snapshot)
	if err != nil {
		t.Fatalf("scope fixture manifest is invalid: %v", err)
	}
	return Subject{Name: "guide", Snapshot: snapshot, SchemaVersion: spec.SchemaVersion, Capabilities: spec.Capabilities}
}

func assertMarkdownScope(t *testing.T, subject Subject, wantFile string) {
	t.Helper()
	cfg := newCfg(t, "strict", "high")
	warnings, errs := GateReadOnly(cfg, []Subject{subject})
	assertMarkdownVerdict(t, warnings, errs, wantFile)
	for _, wantHit := range []bool{false, true} {
		report, err := auditSubject(cfg, subject, false)
		if err != nil || report.CacheHit != wantHit {
			t.Fatalf("persisted audit: %+v, %v", report, err)
		}
		if wantFile == "" {
			if report.Decision != DecisionAllow || len(report.Findings) != 0 {
				t.Fatalf("excluded Markdown produced findings: %+v", report)
			}
		} else {
			if report.Decision != DecisionBlock || len(report.Findings) != 1 {
				t.Fatalf("visible Markdown was not blocked: %+v", report)
			}
			f := report.Findings[0]
			if f.ID != "audit.capability.network-undeclared" || f.Severity != SeverityHigh || f.File != wantFile || !strings.Contains(f.Evidence, "contacts api.example.net,") {
				t.Fatalf("wrong Markdown finding: %+v", f)
			}
		}
		warnings, errs = Gate(cfg, []Subject{subject})
		assertMarkdownVerdict(t, warnings, errs, wantFile)
		warnings, errs = GateReadOnly(cfg, []Subject{subject})
		assertMarkdownVerdict(t, warnings, errs, wantFile)
	}
}

func assertMarkdownVerdict(t *testing.T, warnings, errs []string, wantFile string) {
	t.Helper()
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if wantFile == "" {
		if len(errs) != 0 {
			t.Fatalf("excluded Markdown refused: %v", errs)
		}
	} else if len(errs) != 1 || !strings.Contains(errs[0], "audit.capability.network-undeclared") || !strings.Contains(errs[0], wantFile) {
		t.Fatalf("visible Markdown verdict: %v, want network refusal for %s", errs, wantFile)
	}
}

// Repo research, public board resources and build-only vendor docs cannot
// create network findings; nested build/runtime roots are removed even inside
// an eligible context root. Both public gates and the persisted verdict agree.
func TestMarkdownScopeRepositoryNoise(t *testing.T) {
	subject := markdownScopeSubject(t, map[string]string{
		".research/notes.md":                 markdownContact,
		".task-board/.resources/x.md":        markdownContact,
		"tools/board-cli/vendor/m/README.md": markdownContact,
		"tools/board-cli/go.mod":             "module example.test/board\ngo 1.25.5\n",
		"tools/board-cli/cmd/tool/main.go":   "package main\nfunc main() {}\n",
		"assets/build/guide.md":              markdownContact,
		"assets/build/go.mod":                "module example.test/asset\ngo 1.25.5\n",
		"assets/build/cmd/tool/main.go":      "package main\nfunc main() {}\n",
		"references/runtime/guide.md":        markdownContact,
	}, `{"schema_version":6,"capabilities":{},"build_roots":["tools/board-cli","assets/build"],"runtime_roots":["references/runtime"],"commands":{"board":{"type":"build","driver":"go-v1","source_dir":"tools/board-cli/cmd/tool"},"asset":{"type":"build","driver":"go-v1","source_dir":"assets/build/cmd/tool"}}}`)
	assertMarkdownScope(t, subject, "")
}

// Each core context root still rejects an undeclared contact and admits the
// same bytes with the host declared. Uppercase Markdown extensions stay scanned.
func TestMarkdownScopeVisibleContacts(t *testing.T) {
	for _, file := range []string{"SKILL.md", "agents/a.md", "references/guide.md", ".skill_triggers/a.md", "assets/a.md", "templates/a.md", "examples/a.md", "data/a.MD"} {
		t.Run(file, func(t *testing.T) {
			subject := markdownScopeSubject(t, map[string]string{file: markdownContact}, "")
			assertMarkdownScope(t, subject, file)
			subject.Capabilities = capabilities.Manifest{Network: []string{"api.example.net"}}
			assertMarkdownScope(t, subject, "")
		})
	}
}

// Fixed exclusions apply at every depth. Nearby guide.md controls above
// prevent a detector that ignores the entire references root from passing.
func TestMarkdownScopeExcludedNames(t *testing.T) {
	for _, file := range []string{"references/README.md", "references/deep/CHANGELOG-old.md", "assets/LICENSE.md", "references/tests/guide.md", "references/deep/node_modules/m/guide.md", "references/.git/guide.md", "references/deep/__tests__/guide.md", "notes.md", "references-other/guide.md"} {
		t.Run(file, func(t *testing.T) {
			assertMarkdownScope(t, markdownScopeSubject(t, map[string]string{file: markdownContact}, ""), "")
		})
	}
}

// scripts Markdown follows the exports rule and context exclusions, while
// non-Markdown scripts (including excluded-looking names) and the selected
// manifest retain their historical scanning surface.
func TestMarkdownScopeScriptsAndManifest(t *testing.T) {
	for _, exports := range []bool{false, true} {
		name := "commandless"
		manifest := `{"schema_version":3,"capabilities":{}}`
		if exports {
			name = "exports"
			manifest = `{"schema_version":3,"capabilities":{},"runtime_roots":["runtime"],"commands":{"tool":{"unix_path":"runtime/tool","type":"script"}}}`
		}
		t.Run(name, func(t *testing.T) {
			subject := markdownScopeSubject(t, map[string]string{"scripts/tool": "echo ok\n", "runtime/tool": "echo ok\n", "scripts/x.md": markdownContact}, manifest)
			want := "scripts/x.md"
			if exports {
				want = ""
			}
			assertMarkdownScope(t, subject, want)
			if err := os.Remove(filepath.Join(subject.Snapshot, "scripts/x.md")); err != nil {
				t.Fatal(err)
			}
			writeFiles(t, subject.Snapshot, map[string][]byte{"scripts/tests/contact.txt": []byte(markdownContact)})
			assertMarkdownScope(t, subject, "scripts/tests/contact.txt")
		})
	}
	subject := markdownScopeSubject(t, map[string]string{"scripts/README.md": markdownContact}, "")
	assertMarkdownScope(t, subject, "")
	for _, file := range []string{"agent-skill.json", "csk-skill.json"} {
		t.Run(file, func(t *testing.T) {
			subject := markdownScopeSubject(t, map[string]string{file: `{"schema_version":1,"homepage":"https://api.example.net/guide"}`}, "")
			findings := detect(subject.Snapshot, subject.Capabilities)
			if len(findings) != 1 || findings[0].ID != "audit.capability.network-undeclared" || findings[0].File != file {
				t.Fatalf("manifest finding: %+v", findings)
			}
		})
	}
}

// A ruleset-5 block record for repo-only Markdown must miss in both hash
// framings, with read-only gates leaving it untouched and new allowances
// caching normally. The seeded bytes match the former detector's finding;
// the historical detector itself is exercised by the every-Markdown mutant.
func TestMarkdownScopeInvalidatesRulesetFive(t *testing.T) {
	for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			subject := markdownScopeSubject(t, map[string]string{".research/notes.md": markdownContact}, "")
			subject.HashVersion = version
			cfg := newCfg(t, "strict", "high")
			digest, err := hashing.ContentSHA256WithVersion(subject.Snapshot, nil, version)
			if err != nil {
				t.Fatal(err)
			}
			oldFindings := []Finding{{ID: "audit.capability.network-undeclared", Severity: SeverityHigh, File: ".research/notes.md", Evidence: "contacts api.example.net, which no capabilities.network entry covers", Verifiable: true}}
			storeCachedFindings(cfg, digest, subject, oldFindings, nil)
			currentPath := verdictPath(cfg, digest)
			oldPath := filepath.Join(filepath.Dir(currentPath), "verdict-"+strings.Join([]string{cfg.Audit.Backend, cfg.Audit.Model, PromptVersion, "5"}, "-")+".json")
			old, err := os.ReadFile(currentPath)
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]any
			if err := json.Unmarshal(old, &record); err != nil {
				t.Fatal(err)
			}
			record["ruleset_version"] = "5"
			old, err = json.Marshal(record)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(currentPath); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(oldPath, old, 0600); err != nil {
				t.Fatal(err)
			}
			warnings, errs := GateReadOnly(cfg, []Subject{subject})
			if len(warnings) != 0 || len(errs) != 0 {
				t.Fatalf("stale Markdown refusal reused: warnings=%v errors=%v", warnings, errs)
			}
			if currentPath != oldPath {
				if _, err := os.Stat(currentPath); !os.IsNotExist(err) {
					t.Fatalf("read-only gate wrote verdict: %v", err)
				}
			}
			for _, wantHit := range []bool{false, true} {
				report, err := auditSubject(cfg, subject, false)
				if err != nil || report.CacheHit != wantHit || report.Decision != DecisionAllow || len(report.Findings) != 0 {
					t.Fatalf("stale Markdown refusal reused: report=%+v error=%v want hit=%v", report, err, wantHit)
				}
				warnings, errs := Gate(cfg, []Subject{subject})
				assertMarkdownVerdict(t, warnings, errs, "")
			}
			retained, err := os.ReadFile(oldPath)
			if err != nil || !bytes.Equal(old, retained) {
				t.Fatalf("old block record changed: %v", err)
			}
		})
	}
}
