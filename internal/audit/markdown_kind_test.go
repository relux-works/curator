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
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/hashing"
)

func contextMarkdownSubject(t *testing.T, file, body, selector string) Subject {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, map[string][]byte{
		"agent-context.json": []byte(`{"schema_version":1,"name":"review-context","version":"1.0.0","context":{"modules":[{"path":"` + file + `"` + selector + `}]}}`),
		"context/" + file:    []byte(body),
	})
	return Subject{Name: "review-context", Kind: contextlock.KindContext, Snapshot: root, SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
}

// Both gates audit every declared module, including system/selected modules,
// non-Markdown extensions and skill-excluded names. Clean and declared contacts
// remain admissible. These cases exercise the shared materialization reader.
func TestMarkdownScopeContextModules(t *testing.T) {
	for _, tc := range []struct{ file, selector string }{
		{"a.md", ""}, {"README.md", ""}, {"notes.txt", ""},
		{"test/system.md", `,"class":"system","environments":["codex_cli"]`},
	} {
		t.Run(tc.file, func(t *testing.T) {
			subject := contextMarkdownSubject(t, tc.file, markdownContact, tc.selector)
			assertMarkdownScope(t, subject, "context/"+tc.file)
			subject.Capabilities.Network = []string{"api.example.net"}
			assertMarkdownScope(t, subject, "")
			clean := contextMarkdownSubject(t, tc.file, "# Clean module\n", tc.selector)
			assertMarkdownScope(t, clean, "")
		})
	}
}

// MCP has no Markdown publication surface; its historical all-Markdown detector
// still rejects an undeclared contact outside skill roots. Unknown kinds and
// invalid context manifests use the same conservative fallback. Clean and
// declared controls show that these paths do not reject every package.
func TestMarkdownScopeOtherKinds(t *testing.T) {
	for _, kind := range []string{contextlock.KindMCP, "future-kind", contextlock.KindContext} {
		t.Run(kind, func(t *testing.T) {
			subject := Subject{Name: "package", Kind: kind, Snapshot: t.TempDir(), SchemaVersion: 3, Capabilities: capabilities.ImplicitNone()}
			writeFiles(t, subject.Snapshot, map[string][]byte{"unlisted/README.md": []byte("# Clean\n")})
			assertMarkdownScope(t, subject, "")
			writeFiles(t, subject.Snapshot, map[string][]byte{"unlisted/README.md": []byte(markdownContact)})
			assertMarkdownScope(t, subject, "unlisted/README.md")
			subject.Capabilities.Network = []string{"api.example.net"}
			assertMarkdownScope(t, subject, "")
		})
	}
}

// Missing declared bytes refuse specifically rather than yielding an empty
// allowance. Restoring the same module permits the clean package.
func TestMarkdownScopeContextMissingModule(t *testing.T) {
	subject := contextMarkdownSubject(t, "a.md", "# Clean\n", "")
	if err := os.Remove(filepath.Join(subject.Snapshot, "context", "a.md")); err != nil {
		t.Fatal(err)
	}
	for _, gate := range []func(){
		func() {
			_, errs := Gate(newCfg(t, "strict", "high"), []Subject{subject})
			if len(errs) != 1 || !strings.Contains(errs[0], "audit.context.scan-incomplete") {
				t.Fatalf("missing module verdict: %v", errs)
			}
		},
		func() {
			_, errs := GateReadOnly(newCfg(t, "strict", "high"), []Subject{subject})
			if len(errs) != 1 || !strings.Contains(errs[0], "audit.context.scan-incomplete") {
				t.Fatalf("missing module read-only verdict: %v", errs)
			}
		},
	} {
		gate()
	}
	writeFiles(t, subject.Snapshot, map[string][]byte{"context/a.md": []byte("# Clean\n")})
	assertMarkdownScope(t, subject, "")
}

// The same raw tree can be clean under the skill contract and hostile under the
// context/MCP/unknown contract. Cached allowances cannot cross that boundary.
func TestMarkdownScopeCacheBindsKind(t *testing.T) {
	for _, kind := range []string{contextlock.KindContext, contextlock.KindMCP, "future-kind"} {
		t.Run(kind, func(t *testing.T) {
			subject := contextMarkdownSubject(t, "a.md", markdownContact, "")
			cfg := newCfg(t, "strict", "high")
			subject.Kind = ""
			if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
				t.Fatal(errs)
			}
			subject.Kind = kind
			for attempt, persist := range []bool{false, true, true} {
				report, err := auditSubject(cfg, subject, persist)
				if err != nil || report.CacheHit != (attempt == 2) || report.Decision != DecisionBlock || len(report.Findings) != 1 || report.Findings[0].File != "context/a.md" {
					t.Fatalf("kind cache borrowed allowance: %+v %v", report, err)
				}
			}
		})
	}
}

// rulesetSevenSource matches anchors with LF internally and restores the input's
// line endings so Windows checkout fixtures remain CRLF when overlaid.
func rulesetSevenSource(source string) (string, error) {
	crlf := strings.Contains(source, "\r\n")
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = regexp.MustCompile(`RulesetVersion = "[0-9]+"`).ReplaceAllString(source, `RulesetVersion = "7"`)
	for before, after := range map[string]string{
		`if kind == contextlock.KindContext {`:                               `if false && kind == contextlock.KindContext {`,
		`if kind == contextlock.KindSkill {`:                                 `if true {`,
		`if markdown && (kind == contextlock.KindSkill || contextModules) {`: `if markdown || contextModules {`,
		`"subject_kind":    subject.Kind,`:                                   ``,
		"\tif data.Kind != subject.Kind {\n\t\treturn nil, false\n\t}":       "",
	} {
		if strings.Count(source, before) != 1 {
			return "", fmt.Errorf("ruleset-7 anchor not unique: %q", before)
		}
		source = strings.Replace(source, before, after, 1)
	}
	if crlf {
		source = strings.ReplaceAll(source, "\n", "\r\n")
	}
	return source, nil
}

// LF and CRLF fixtures retain their endings and receive the ruleset-7 changes.
// Missing and duplicate multiline anchors must return the specific error rather
// than silently producing an incomplete seed.
func TestRulesetSevenSourceLineEndings(t *testing.T) {
	payload, err := os.ReadFile("audit.go")
	if err != nil {
		t.Fatal(err)
	}
	original := strings.ReplaceAll(string(payload), "\r\n", "\n")
	const anchor = "\tif data.Kind != subject.Kind {\n\t\treturn nil, false\n\t}"
	for _, ending := range []struct{ name, value string }{{"LF", "\n"}, {"CRLF", "\r\n"}} {
		for _, variant := range []string{"unique", "missing", "duplicate"} {
			t.Run(ending.name+"/"+variant, func(t *testing.T) {
				source := original
				switch variant {
				case "missing":
					source = strings.Replace(source, anchor, "", 1)
				case "duplicate":
					source += anchor + "\n"
				}
				source = strings.ReplaceAll(source, "\n", ending.value)
				got, err := rulesetSevenSource(source)
				if variant != "unique" {
					want := fmt.Sprintf("ruleset-7 anchor not unique: %q", anchor)
					if err == nil || err.Error() != want || got != "" {
						t.Fatalf("ambiguous anchor: seed empty=%t, error=%v; want %s", got == "", err, want)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				normalized := strings.ReplaceAll(got, "\r\n", "\n")
				if got != strings.ReplaceAll(normalized, "\n", ending.value) {
					t.Fatal("ruleset-7 seed changed source line endings")
				}
				if !strings.Contains(got, `RulesetVersion = "7"`) || strings.Contains(normalized, anchor) || strings.Contains(got, `"subject_kind":    subject.Kind,`) {
					t.Fatal("ruleset-7 seed retained current ruleset or kind binding")
				}
			})
		}
	}
}

// A production Gate writes and reuses an empty ruleset-7 verdict with the r2
// skill-only selection and without a kind binding. The current gates reject the
// same context in both hash frames, leave the old record intact and cache the
// new refusal. No handwritten finding record is used as the seed.
func TestMarkdownScopeInvalidatesRulesetSevenAllowance(t *testing.T) {
	const seedEnv = "CURATOR_TEST_RULESET_SEVEN_CACHE"
	seed := os.Getenv(seedEnv)
	root := seed
	if root == "" {
		root = t.TempDir()
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
		source, err := rulesetSevenSource(string(payload))
		if err != nil {
			t.Fatal(err)
		}
		replacement := filepath.Join(root, "seven.go")
		if err := os.WriteFile(replacement, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{sourcePath: replacement}})
		overlayPath := filepath.Join(root, "seven.json")
		if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("go", "test", "-overlay", overlayPath, "-count=1", "-timeout=2m", "-v", "-run", "^TestMarkdownScopeInvalidatesRulesetSevenAllowance$", ".")
		cmd.Env = append(os.Environ(), seedEnv+"="+root)
		output, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(output), "--- PASS: TestMarkdownScopeInvalidatesRulesetSevenAllowance") {
			t.Fatalf("real ruleset-7 seed failed: %v\n%s", err, output)
		}
		t.Logf("real ruleset-7 Gate seed exit 0:\n%s", output)
	}
	for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			dir := filepath.Join(root, fmt.Sprintf("v%d", version))
			snapshot := filepath.Join(dir, "snapshot")
			writeFiles(t, snapshot, map[string][]byte{"agent-context.json": []byte(`{"schema_version":1,"name":"review-context","version":"1.0.0","context":{"modules":[{"path":"a.md"}]}}`), "context/a.md": []byte(markdownContact)})
			subject := Subject{Name: "review-context", Kind: contextlock.KindContext, Snapshot: snapshot, SchemaVersion: 3, HashVersion: version, Capabilities: capabilities.ImplicitNone()}
			cfg := newCfg(t, "strict", "high")
			cfg.Path = filepath.Join(dir, "home", "config.json")
			if seed != "" {
				if _, errs := Gate(cfg, []Subject{subject}); len(errs) != 0 {
					t.Fatal(errs)
				}
				report, err := auditSubject(cfg, subject, false)
				if err != nil || !report.CacheHit || report.Decision != DecisionAllow || len(report.Findings) != 0 {
					t.Fatalf("ruleset-7 allowance not seeded: %+v %v", report, err)
				}
				return
			}
			hash, err := hashing.ContentSHA256WithVersion(snapshot, nil, version)
			if err != nil {
				t.Fatal(err)
			}
			oldPath := strings.Replace(verdictPath(cfg, hash), "-"+RulesetVersion+".json", "-7.json", 1)
			old, err := os.ReadFile(oldPath)
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]json.RawMessage
			if err := json.Unmarshal(old, &record); err != nil {
				t.Fatal(err)
			}
			if _, ok := record["subject_kind"]; ok {
				t.Fatal("ruleset-7 seed unexpectedly binds kind")
			}
			if string(record["findings"]) != "[]" && string(record["findings"]) != "null" {
				t.Fatalf("seed findings: %s", record["findings"])
			}
			for attempt, persist := range []bool{false, true, true} {
				report, err := auditSubject(cfg, subject, persist)
				if err != nil || report.CacheHit != (attempt == 2) || report.Decision != DecisionBlock || len(report.Findings) != 1 || report.Findings[0].ID != "audit.capability.network-undeclared" || report.Findings[0].File != "context/a.md" {
					t.Fatalf("ruleset-7 allowance reused: %+v %v", report, err)
				}
			}
			if _, errs := GateReadOnly(cfg, []Subject{subject}); len(errs) != 1 {
				t.Fatalf("read-only accepted old allowance: %v", errs)
			}
			current, err := os.ReadFile(oldPath)
			if err != nil || !bytes.Equal(old, current) {
				t.Fatal("old ruleset-7 record changed")
			}
		})
	}
}
