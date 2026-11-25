package harnessboundary

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
	"github.com/relux-works/curator/internal/harnessvalue"
)

const baselineHead = "267bbf477354f4f3d35c9530ec8da0503cd86b8a"
const frozenDigest = "a186ea55e67809751365ae5d5d08e9af47af3d16fc629b682e1ee544cc03ae1c"

func readJSON(t *testing.T, name string, output any) []byte {
	t.Helper()
	bytes, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(bytes, output); err != nil {
		t.Fatal(err)
	}
	return bytes
}
func vocabulary(t *testing.T) map[string][]string {
	t.Helper()
	var v map[string][]string
	readJSON(t, "vocabulary.json", &v)
	for _, id := range harnessregistry.IDs() {
		value := string(id)
		if _, ok := v[value]; !ok {
			v[value] = []string{value}
		}
	}
	return v
}
func plant(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
func scan(t *testing.T, root string) []Finding {
	t.Helper()
	findings, err := ScanBranches(root, vocabulary(t))
	if err != nil {
		t.Fatal(err)
	}
	return findings
}
func requireFinding(t *testing.T, findings []Finding, path string, line int) {
	t.Helper()
	for _, f := range findings {
		if f.Path == path && f.Line == line {
			return
		}
	}
	t.Fatalf("missing %s:%d finding: %v", path, line, findings)
}

// The production scan accounts for remaining legacy policy, rejects every new
// occurrence, and pins the original baseline independently of the active list.
// Constant folding is bounded to source constants; dynamic identifiers remain
// an exact-candidate review obligation.
func TestNoHarnessBranches(t *testing.T) {
	t.Run("predecessor-ratchet", testPredecessorContract)
	t.Run("physical-source-locations", testPhysicalSourceLocations)
	t.Run("canonical-source-lines", testCanonicalSourceLines)
	var frozen, active ExceptionSet
	bytes := readJSON(t, "frozen-exceptions.json", &frozen)
	readJSON(t, "exceptions.json", &active)
	if digest(bytes) != frozenDigest || frozen.SourceHead != baselineHead {
		t.Fatal("frozen baseline changed without a contract revision")
	}
	if err := CheckExceptionSubset(frozen, active); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..")
	production := scan(t, root)
	if len(production) != len(active.Exceptions) {
		t.Errorf("active exception occurrences=%d source occurrences=%d; remove stale allowances", len(active.Exceptions), len(production))
	}
	if unmatched := ApplyExceptions(production, active); len(unmatched) > 0 {
		for _, f := range unmatched {
			t.Error(f.String())
		}
	}
	t.Run("future-registry-ID", func(t *testing.T) {
		root := t.TempDir()
		policy := vocabulary(t)
		policy["future_cli"] = []string{"future_cli"}
		plant(t, root, "internal/harness/future_cli/adapter.go", "package future\nconst ID=\"future_cli\"\n")
		findings, err := ScanBranches(root, policy)
		if err != nil || len(findings) != 0 {
			t.Fatalf("new module positive: %v %v", findings, err)
		}
		rel := "internal/shared/plant.go"
		plant(t, root, rel, "package shared\nfunc selected(id string)bool{return id==\"future_cli\"}\n")
		findings, err = ScanBranches(root, policy)
		if err != nil {
			t.Fatal(err)
		}
		requireFinding(t, findings, rel, 2)
	})
	t.Run("generic-positive", func(t *testing.T) {
		root := t.TempDir()
		plant(t, root, "internal/shared/generic.go", "package shared\nfunc same(a,b string) bool { return a==b }\n")
		if findings := scan(t, root); len(findings) != 0 {
			t.Fatal(findings)
		}
	})
	for name, body := range map[string]string{
		"new-branch":      "package shared\nfunc selected(id string) bool { return id == \"claude_code\" }\n",
		"cursor-branch":   "package shared\nfunc selected(id string) bool { return id == \"cursor\" }\n",
		"windsurf-branch": "package shared\nfunc selected(id string) bool { return id == \"windsurf\" }\n",
		"computed-id":     "package shared\ntype ID string\nconst selected = \"claude\" + \"_code\"\nfunc selectedID(id ID) bool { return id == selected }\n",
		"native-path":     "package shared\nconst nativePath = \".codex/skills\"\n",
		"cursor-path":     "package shared\nconst nativePath = \".cursor/rules\"\n",
		"concrete-import": "package shared\nimport _ \"github.com/relux-works/curator/internal/harness/cursor\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			rel := "internal/shared/plant.go"
			plant(t, root, rel, body)
			line := 2
			if name == "computed-id" {
				line = 4
			}
			requireFinding(t, scan(t, root), rel, line)
		})
	}
	t.Run("one-to-one-duplicate", func(t *testing.T) {
		root := t.TempDir()
		rel := "internal/shared/plant.go"
		line := "const nativePath = \".codex/skills\""
		plant(t, root, rel, "package shared\n"+line+"\n")
		initial := scan(t, root)
		if len(initial) != 1 {
			t.Fatalf("positive occurrences=%d", len(initial))
		}
		e := ExceptionSet{SourceHead: baselineHead, Exceptions: []Exception{{rel, initial[0].HarnessID, initial[0].LineSHA256}}}
		if len(ApplyExceptions(initial, e)) != 0 {
			t.Fatal("valid allowance refused")
		}
		// The duplicate is deliberately identical, including its digest.
		plant(t, root, rel, "package shared\n"+line+"\n"+line+"\n")
		unmatched := ApplyExceptions(scan(t, root), e)
		if len(unmatched) != 1 {
			t.Fatalf("duplicate did not consume separate allowance: %v", unmatched)
		}
		requireFinding(t, unmatched, rel, 3)
	})
	t.Run("shrink-only", func(t *testing.T) {
		base := ExceptionSet{SourceHead: baselineHead, Exceptions: []Exception{{"internal/shared/old.go", "codex_cli", strings.Repeat("a", 64)}}}
		if err := CheckExceptionSubset(base, base); err != nil {
			t.Fatal(err)
		}
		if err := CheckExceptionSubset(base, ExceptionSet{SourceHead: baselineHead}); err != nil {
			t.Fatal(err)
		}
		grown := ExceptionSet{SourceHead: baselineHead, Exceptions: append(append([]Exception(nil), base.Exceptions...), base.Exceptions[0])}
		if err := CheckExceptionSubset(base, grown); err == nil || !strings.Contains(err.Error(), "harness_exception_growth") {
			t.Fatalf("growth verdict: %v", err)
		}
		changed := base
		changed.SourceHead = "other"
		if err := CheckExceptionSubset(base, changed); err == nil {
			t.Fatal("baseline substitution accepted")
		}
	})
	t.Run("generated-source-positive-and-negative", func(t *testing.T) {
		root := t.TempDir()
		rel := "internal/shared/generated.go"
		plant(t, root, rel, "package shared\n// A normal source comment.\n")
		if len(scan(t, root)) != 0 {
			t.Fatal("normal source refused")
		}
		plant(t, root, rel, "// Code generated by fixture. DO NOT EDIT.\npackage shared\n")
		requireFinding(t, scan(t, root), rel, 1)
		plant(t, root, rel, "package shared\n//go:generate fixture\n")
		requireFinding(t, scan(t, root), rel, 2)
	})
}

func TestRemovedCodexSkillAllowanceCannotReturn(t *testing.T) {
	var active ExceptionSet
	readJSON(t, "exceptions.json", &active)
	root := t.TempDir()
	rel := "internal/adapters/adapters.go"
	plant(t, root, rel, "package adapters\nvar legacyAgentPaths = map[string]string{\n\t\"codex_cli\":   \".codex/skills\",\n}\n")
	findings, err := ScanBranches(root, vocabulary(t))
	if err != nil {
		t.Fatal(err)
	}
	requireFinding(t, findings, rel, 3)
	if unmatched := ApplyExceptions(findings, active); len(unmatched) != 1 {
		t.Fatalf("restored Codex skill-path branch was not rejected: %v", unmatched)
	}
}

// Each effect plant is parsed, never executed; imports and concrete call lines
// are rejected. Positive fixtures cover the complete direct import allowlist
// and lexical layouts, not native vendor behavior.
func TestAdapterEffectBoundary(t *testing.T) {
	t.Run("physical-effect-locations", testPhysicalEffectLocations)
	t.Run("bounded-layout-values", testLegitimateLayouts)
	t.Run("symlinked-allowlisted-helper", testSymlinkedAllowlistedHelper)
	findings, err := ScanEffects(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range findings {
		t.Error(f.String())
	}
	positive := `package fixture
import (
 "bytes"
 "encoding/json"
 "fmt"
 "path"
 "path/filepath"
 "sort"
 "strconv"
 "strings"
 "github.com/relux-works/curator/internal/harnessapi"
 "github.com/relux-works/curator/internal/harnessvalue"
 "github.com/relux-works/curator/internal/harnessformats"
)
func lexical() string { return fmt.Sprintf("%s",filepath.Join("parent","native")) }
`
	root := t.TempDir()
	rel := "internal/harness/fixture/adapter.go"
	plant(t, root, rel, positive)
	findings, err = ScanEffects(root)
	if err != nil || len(findings) != 0 {
		t.Fatalf("positive import/lexical control: %v %v", findings, err)
	}
	for name, row := range map[string][2]string{
		"credential-read": {"os", `os.ReadFile("fixture/native/auth.json")`},
		"external-write":  {"os", `os.WriteFile("fixture/outside/output",nil,0600)`},
		"process-launch":  {"os/exec", `exec.Command("fixture")`},
		"network":         {"net/http", `http.Get("https://example.invalid")`},
		"hidden-helper":   {"github.com/relux-works/curator/internal/envprofile", `envprofile.Current("fixture")`},
		"registry-cycle":  {"github.com/relux-works/curator/internal/harnessregistry", `harnessregistry.IDs()`},
		"filesystem-path": {"path/filepath", `filepath.EvalSymlinks("fixture")`},
		"format-output":   {"fmt", `fmt.Println("fixture")`},
		"unsafe":          {"unsafe", `unsafe.Sizeof(1)`},
		"reflection":      {"reflect", `reflect.TypeOf(1)`},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			plant(t, root, rel, "package fixture\nimport \""+row[0]+"\"\nfunc planted() {\n "+row[1]+"\n}\n")
			findings, err := ScanEffects(root)
			if err != nil {
				t.Fatal(err)
			}
			requireFinding(t, findings, rel, 4)
		})
	}
	for _, body := range []string{"package fixture\n//go:linkname hidden elsewhere\n", "package fixture\nimport \"C\"\n", "package fixture\nimport . \"fmt\"\n"} {
		root := t.TempDir()
		plant(t, root, rel, body)
		findings, err := ScanEffects(root)
		if err != nil || len(findings) == 0 {
			t.Fatalf("directive/import plant escaped: %v %v", findings, err)
		}
	}
	for _, helper := range []string{"harnessapi", "harnessvalue", "harnessformats"} {
		root := t.TempDir()
		rel := "internal/" + helper + "/plant.go"
		plant(t, root, rel, "package "+helper+"\nimport \"os\"\nfunc bad() { os.Remove(\"fixture\") }\n")
		findings, err := ScanEffects(root)
		if err != nil {
			t.Fatal(err)
		}
		requireFinding(t, findings, rel, 3)
	}
}

func testSymlinkedAllowlistedHelper(t *testing.T) {
	const adapter = `package claude_code
import "github.com/relux-works/curator/internal/harnessvalue"
func invoke() { harnessvalue.Run() }
`
	t.Run("regular-helper-positive", func(t *testing.T) {
		root := t.TempDir()
		plant(t, root, "internal/harnessvalue/helper.go", "package harnessvalue\nimport \"strings\"\nfunc Run() string { return strings.TrimSpace(\" fixture \") }\n")
		plant(t, root, "internal/harness/claude_code/adapter.go", adapter)
		findings, err := ScanEffects(root)
		if err != nil || len(findings) != 0 {
			t.Fatalf("regular allowlisted helper was rejected: findings=%v err=%v", findings, err)
		}
	})
	t.Run("reviewer-process-helper-symlink-is-refused", func(t *testing.T) {
		root := t.TempDir()
		plant(t, root, "internal/evilhelper/helper.go", "package harnessvalue\nimport \"os/exec\"\nfunc Run() { _ = exec.Command(\"fixture\") }\n")
		plant(t, root, "internal/harness/claude_code/adapter.go", adapter)
		link := filepath.Join(root, "internal", "harnessvalue")
		target := filepath.Join(root, "internal", "evilhelper")
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("host cannot create the required symlink control: %v", err)
		}
		findings, err := ScanEffects(root)
		if err == nil || !strings.Contains(err.Error(), "source_directory_symlink_unsupported") {
			t.Fatalf("reviewer helper-symlink plant escaped the effect gate: findings=%v err=%v", findings, err)
		}
	})
}

// Legitimate layouts are exercised as values, independently of the staged
// adapters' unported operations. These constructors have no ambient effects.
func testLegitimateLayouts(t *testing.T) {
	paths := func(names ...string) harnessapi.NamedPaths {
		t.Helper()
		var values []harnessvalue.RelativePath
		for _, name := range names {
			value, err := harnessvalue.NewRelativePath(name)
			if err != nil {
				t.Fatal(err)
			}
			values = append(values, value)
		}
		result, err := harnessapi.NewNamedPaths(values)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	for _, row := range []struct {
		name   string
		roots  []string
		parent bool
	}{
		{"direct-home", []string{"home"}, false},
		{"parent-valued-xdg", []string{"xdg/native"}, true},
		{"split-roots", []string{"config/native", "data/native", "state/native", "cache/native"}, false},
	} {
		layout, err := harnessapi.NewHomeLayout("FIXTURE_HOME", paths(row.roots...), row.parent)
		if err != nil || layout.ParentValued() != row.parent || len(layout.Roots().Paths()) != len(row.roots) {
			t.Fatalf("%s: %v", row.name, err)
		}
	}
	for _, scope := range []harnessapi.Scope{harnessapi.UserScope, harnessapi.ProjectScope} {
		discovery, err := harnessapi.NewSkillDiscovery(paths(".agents/skills"), paths("generated/skills"), scope, true, false)
		if err != nil || !discovery.NativeDiscovery() || discovery.Mirror() || discovery.Scope() != scope {
			t.Fatalf("native discovery value: %v", err)
		}
		inspection, err := harnessapi.NewMCPChannel(paths("mcp.json"), harnessapi.JSON, true, false, true)
		if err != nil || !inspection.Inspect() || inspection.Write() || !inspection.RequiresTrust() {
			t.Fatalf("MCP inspection/write separation: %v", err)
		}
	}
	if _, err := harnessapi.NewMCPChannel(paths("mcp.json"), harnessapi.JSON, false, true, false); err == nil {
		t.Fatal("write without inspection accepted")
	}
	for _, form := range []harnessapi.Form{harnessapi.Monolithic, harnessapi.Referenced} {
		instructions, err := harnessapi.NewInstructionSpec(paths("instructions/root.md"), form, []byte("Fixture ownership notice\n"))
		if err != nil || instructions.Form() != form {
			t.Fatalf("instruction form: %v", err)
		}
	}
	route, err := harnessapi.NewCredentialRoute(paths("auth.json"), "file-link", true)
	if err != nil || !route.Qualified() || route.Strategy() != "file-link" || len(route.Paths().Paths()) != 1 {
		t.Fatalf("credential metadata route: %v", err)
	}
	property, err := harnessapi.NewProperty("credential_route", "file-link")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := harnessapi.NewSnapshot([]harnessapi.Property{property}); err != nil {
		t.Fatal(err)
	}
}
