package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envregistry"
)

// claudeResolvedPath resolves the deepest existing ancestor of path and keeps
// the rest lexically, as a read through a symlinked directory resolves.
func claudeResolvedPath(t *testing.T, path string) string {
	t.Helper()
	rest := ""
	for dir := filepath.Clean(path); ; dir = filepath.Dir(dir) {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			return filepath.Join(resolved, rest)
		}
		if filepath.Dir(dir) == dir {
			t.Fatalf("no existing ancestor of %s", path)
		}
		rest = filepath.Join(filepath.Base(dir), rest)
	}
}

// claudeRuleMatches models a literal absolute Read(//P/**) allow rule: it
// matches files below P. Other rule shapes match nothing here.
func claudeRuleMatches(rules []string, path string) bool {
	posix := filepath.ToSlash(path)
	for _, rule := range rules {
		if !strings.HasPrefix(rule, "Read(//") || !strings.HasSuffix(rule, "/**)") {
			continue
		}
		escaped := rule[len("Read(/") : len(rule)-len("/**)")]
		var root strings.Builder
		for i := 0; i < len(escaped); i++ {
			if escaped[i] == '\\' && i+1 < len(escaped) {
				i++
			}
			root.WriteByte(escaped[i])
		}
		if strings.HasPrefix(posix, root.String()+"/") {
			return true
		}
	}
	return false
}

// claudeAllowsRead models Claude's documented symlink rule for allow rules:
// a read is allowed only when the requested path and the path it resolves to
// both match. This is a model of the vendor contract, not Claude itself.
func claudeAllowsRead(t *testing.T, rules []string, requested string) bool {
	t.Helper()
	return claudeRuleMatches(rules, requested) && claudeRuleMatches(rules, claudeResolvedPath(t, requested))
}

func claudeSettingsAllow(t *testing.T, path string) []string {
	t.Helper()
	object := claudeNumberObject(t, readClaudeState(t, path))
	permissions, _ := object["permissions"].(map[string]any)
	raw, _ := permissions["allow"].([]any)
	rules := make([]string, 0, len(raw))
	for _, entry := range raw {
		if rule, ok := entry.(string); ok {
			rules = append(rules, rule)
		}
	}
	return rules
}

func claudeReadScopeFixture(t *testing.T) (*managedFixture, ResolveRequest, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the read-scope model below uses POSIX paths and symlinks")
	}
	fx, req, claudeState := claudeRepairFixture(t, envregistry.FormMonolithic)
	return fx, req, filepath.Join(filepath.Dir(claudeState), claudeSettingsName)
}

// A repaired Claude home lets its own declared skill and selected context be
// read without the outside-read prompt, and nothing else. The positive case is
// the 2026-10-29 regression shape (a Read of
// CLAUDE_CONFIG_DIR/skills/<declared skill>/references/firewall.md). The
// negative controls are a sibling home's link into the same immutable skill
// entry, an unselected context package, an a2a-swarm configuration path and a
// retargeted skill link. Limit: Claude's matcher is modelled from its
// documented allow and symlink rules; a live Claude 2.1.312 Read is qualified
// on the host separately.
func TestClaudeReadScopeAdmitsOnlyOwnSkillsAndContexts(t *testing.T) {
	fx, _, settings := claudeReadScopeFixture(t)
	homeDir := ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode)
	rules := claudeSettingsAllow(t, settings)
	if len(rules) == 0 {
		t.Fatalf("repair projected no read scope into %s", settings)
	}
	for _, rule := range rules {
		if !strings.HasPrefix(rule, "Read(//") || !strings.HasSuffix(rule, "/**)") {
			t.Fatalf("projected rule %q is not a literal absolute Read rule", rule)
		}
	}
	skillLink := filepath.Join(homeDir, "skills", "myskill")
	firewall := filepath.Join(skillLink, "references", "firewall.md")
	if !claudeAllowsRead(t, rules, firewall) {
		t.Fatalf("own declared skill reference %s is not admitted by %v", firewall, rules)
	}
	if !claudeAllowsRead(t, rules, filepath.Join(skillLink, "SKILL.md")) {
		t.Fatalf("own declared SKILL.md is not admitted by %v", rules)
	}
	contexts := filepath.Join(contextstore.Root(fx.home), "context", fx.profile)
	entries, err := os.ReadDir(contexts)
	if err != nil || len(entries) != 1 {
		t.Fatalf("selected context store entries %v: %v", entries, err)
	}
	contextModule := filepath.Join(contexts, entries[0].Name(), "context", "a.md")
	if !claudeAllowsRead(t, rules, contextModule) {
		t.Fatalf("own selected context module %s is not admitted by %v", contextModule, rules)
	}

	target, err := os.Readlink(skillLink)
	if err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(EnvRoot(fx.home), "sibling", envregistry.ClaudeCode, "skills")
	if err := os.MkdirAll(sibling, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(sibling, "myskill")); err != nil {
		t.Fatal(err)
	}
	unselected := contextstore.EntryDir(fx.home, "context", "other-package", strings.Repeat("e", 64))
	if err := os.MkdirAll(filepath.Join(unselected, "context"), 0o700); err != nil {
		t.Fatal(err)
	}
	elsewhere := t.TempDir()
	for name, path := range map[string]string{
		"sibling home link into the same skill entry": filepath.Join(sibling, "myskill", "references", "firewall.md"),
		"unselected context package":                  filepath.Join(unselected, "context", "a.md"),
		"a2a-swarm configuration":                     filepath.Join(filepath.Dir(fx.home), ".config", "a2a-swarm", "agents", "peer.json"),
		"whole contexts store sibling":                filepath.Join(contextstore.Root(fx.home), "skill", "secret.md"),
		"managed home outside skills":                 filepath.Join(homeDir, ".claude.json"),
	} {
		if claudeAllowsRead(t, rules, path) {
			t.Fatalf("%s %s is admitted by %v", name, path, rules)
		}
	}
	// A retargeted link resolves outside the projected entry and is refused.
	if err := os.Remove(skillLink); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, skillLink); err != nil {
		t.Fatal(err)
	}
	if claudeAllowsRead(t, rules, firewall) {
		t.Fatalf("retargeted skill link to %s is admitted by %v", elsewhere, rules)
	}
}

// An existing home without the projection (a pre-fix member) is stale for
// verification, and the launch-time repair adds the scope; a repeated verify
// is then current. A deleted settings file is restored the same way.
func TestClaudeReadScopeReachesExistingHomes(t *testing.T) {
	_, req, settings := claudeReadScopeFixture(t)
	want := readClaudeState(t, settings)
	for _, variant := range []string{"pre-fix-absent", "pre-fix-unrelated-settings"} {
		t.Run(variant, func(t *testing.T) {
			if err := os.Remove(settings); err != nil {
				t.Fatal(err)
			}
			if variant == "pre-fix-unrelated-settings" {
				if err := os.WriteFile(settings, []byte("{\"model\":\"opus\"}\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			verify := req
			verify.Repair = false
			_, err := Resolve(verify)
			if err == nil || !strings.Contains(err.Error(), "read scope") {
				t.Fatalf("verify of a home without the read scope: %v", err)
			}
			if _, err := Resolve(req); err != nil {
				t.Fatalf("repair: %v", err)
			}
			if got := claudeSettingsAllow(t, settings); !reflect.DeepEqual(got, claudeSettingsAllowOf(t, want)) {
				t.Fatalf("repaired scope %v, want %v", got, claudeSettingsAllowOf(t, want))
			}
			if _, err := Resolve(verify); err != nil {
				t.Fatalf("verify after repair: %v", err)
			}
		})
	}
}

func claudeSettingsAllowOf(t *testing.T, payload []byte) []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), claudeSettingsName)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return claudeSettingsAllow(t, path)
}

// Repair merges the scope into operator settings without changing anything
// it does not own: unrelated keys and number lexemes, deny and ask rules, an
// operator allow rule, an operator glob over the store and the outside-read
// block stay; a retired projected rule is removed; no additionalDirectories,
// Edit rule or default mode is introduced.
func TestClaudeReadScopePreservesOperatorSettings(t *testing.T) {
	fx, req, settings := claudeReadScopeFixture(t)
	homeDir := ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode)
	projected := claudeSettingsAllow(t, settings)
	retired := claudeReadRule(filepath.Join(homeDir, "skills", "retired"))
	operatorGlob := "Read(/" + filepath.ToSlash(contextstore.Root(fx.home)) + "/**)"
	before := `{
	  "model": "opus", "cleanupPeriodDays": 30.50, "big": 9007199254740993123,
	  "permissions": {
	    "blockReadsOutsideWorkingDirectories": false,
	    "deny": ["Read(//` + strings.TrimPrefix(filepath.ToSlash(homeDir), "/") + `/skills/myskill/secret/**)"],
	    "ask": ["Bash(git push *)"],
	    "allow": ["Bash(make test)", ` + jsonString(t, retired) + `, ` + jsonString(t, operatorGlob) + `, 7]
	  },
	  "env": {"A": "<&>"}
	}`
	if err := os.WriteFile(settings, []byte(before), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(req); err != nil {
		t.Fatalf("repair: %v", err)
	}
	got := claudeNumberObject(t, readClaudeState(t, settings))
	want := claudeNumberObject(t, []byte(before))
	allow := []any{"Bash(make test)", operatorGlob, json.Number("7")}
	for _, rule := range projected {
		allow = append(allow, rule)
	}
	want["permissions"].(map[string]any)["allow"] = allow
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge changed unowned settings:\ngot  %#v\nwant %#v", got, want)
	}
	if !bytes.Contains(readClaudeState(t, settings), []byte(`"<&>"`)) {
		t.Fatalf("merge HTML-escaped an operator value")
	}
	req.Repair = false
	if _, err := Resolve(req); err != nil {
		t.Fatalf("verify after merge: %v", err)
	}
}

func jsonString(t *testing.T, value string) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

// Settings Curator does not override stay byte-identical (and a link keeps its
// target): resolve still succeeds and reports the skipped projection as a
// warning. The valid-object case above is the positive control.
func TestClaudeReadScopeLeavesConflictingSettings(t *testing.T) {
	for name, payload := range map[string]string{
		"array document":          "[1]\n",
		"invalid JSON":            "{\"permissions\":\n",
		"permissions not object":  "{\"permissions\":[\"Read\"]}\n",
		"allow not array":         "{\"permissions\":{\"allow\":\"Read(//x/**)\"}}\n",
		"settings is a link":      "{\"model\":\"opus\"}\n",
		"settings is a directory": "",
	} {
		t.Run(name, func(t *testing.T) {
			_, req, settings := claudeReadScopeFixture(t)
			if err := os.Remove(settings); err != nil {
				t.Fatal(err)
			}
			check := settings
			switch name {
			case "settings is a link":
				check = filepath.Join(t.TempDir(), "operator-settings.json")
				if err := os.WriteFile(check, []byte(payload), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(check, settings); err != nil {
					t.Fatal(err)
				}
			case "settings is a directory":
				if err := os.Mkdir(settings, 0o700); err != nil {
					t.Fatal(err)
				}
			default:
				if err := os.WriteFile(settings, []byte(payload), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			result, err := Resolve(req)
			if err != nil {
				t.Fatalf("repair with conflicting settings: %v", err)
			}
			if !strings.Contains(strings.Join(result.Warnings, "\n"), "read scope") {
				t.Fatalf("conflict was not reported as a warning: %v", result.Warnings)
			}
			if name == "settings is a directory" {
				if info, err := os.Lstat(settings); err != nil || !info.IsDir() {
					t.Fatalf("settings directory changed: %v %v", info, err)
				}
				return
			}
			if got := string(readClaudeState(t, check)); got != payload {
				t.Fatalf("conflicting settings changed to %q", got)
			}
			if name == "settings is a link" {
				if target, err := os.Readlink(settings); err != nil || target != check {
					t.Fatalf("settings link changed: %q %v", target, err)
				}
			}
		})
	}
}

// Ownership covers only the projected shape below the managed skills
// directory and the contexts store: a literal path with the /** suffix.
// Globs, prefixes without a path, other tools and other roots are operator
// rules and are never retired.
func TestClaudeReadScopeOwnership(t *testing.T) {
	scope := claudeReadScope{prefixes: []string{"Read(//h/skills/", "Read(//s/contexts/"}}
	for rule, want := range map[string]bool{
		"Read(//h/skills/a2a/**)":                     true,
		"Read(//s/contexts/skill-context/v1/v2/x/**)": true,
		`Read(//s/contexts/context/p\[1\]/abc/**)`:    true,
		"Read(//s/contexts/**)":                       false,
		"Read(//s/contexts/context/*/**)":             false,
		"Read(//h/skills//**)":                        false,
		"Read(//h/skills/**)":                         false,
		"Read(//h/skills/a2a)":                        false,
		"Edit(//h/skills/a2a/**)":                     false,
		"Read(//h/other/a2a/**)":                      false,
		"Read(//h/skills/a2a/**)x":                    false,
	} {
		if got := scope.owns(rule); got != want {
			t.Errorf("owns(%q) = %v, want %v", rule, got, want)
		}
	}
}

// Rules render literal paths: gitignore metacharacters are escaped, and the
// merge is idempotent (a second merge reports no change).
func TestClaudeReadScopeRuleEscapingAndIdempotence(t *testing.T) {
	if got, want := claudeReadRule("/a b/[x]*?/c"), `Read(//a b/\[x\]\*\?/c/**)`; got != want {
		t.Fatalf("claudeReadRule = %q, want %q", got, want)
	}
	scope := claudeReadScope{rules: []string{"Read(//h/skills/a/**)"}, prefixes: []string{"Read(//h/skills/"}}
	first, changed, err := mergeClaudeReadScope(nil, false, scope)
	if err != nil || !changed {
		t.Fatalf("first merge: changed=%v err=%v", changed, err)
	}
	if _, changed, err := mergeClaudeReadScope(first, true, scope); err != nil || changed {
		t.Fatalf("second merge: changed=%v err=%v", changed, err)
	}
	if _, changed, err := mergeClaudeReadScope(nil, false, claudeReadScope{}); err != nil || changed {
		t.Fatalf("empty scope wrote an absent file: changed=%v err=%v", changed, err)
	}
}
