package claudememory

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func testExclusions(t *testing.T, payload []byte) []string {
	t.Helper()
	_, entries, err := exclusionSettings(payload)
	if err != nil {
		t.Fatal(err)
	}
	return entries
}

// Refresh removes only Curator-owned occurrences. A pre-existing user pattern,
// duplicate user occurrence, unrelated number lexeme and user glob all survive.
func TestReconcileOwnershipAcrossProjects(t *testing.T) {
	base := t.TempDir()
	native := filepath.Join(base, "operator", ".claude", "CLAUDE.md")
	old := filepath.Join(base, "old[1]*", "CLAUDE.md")
	next := filepath.Join(base, "next", "CLAUDE.local.md")
	user := filepath.Join(base, "chosen", "CLAUDE.md")
	userPattern := literalPattern(filepath.ToSlash(user))
	input, _ := json.Marshal(map[string]any{"claudeMdExcludes": []string{"**/vendor/CLAUDE.md", userPattern}})
	input = bytes.Replace(input, []byte(`"claudeMdExcludes"`), []byte(`"n":1e999,"claudeMdExcludes"`), 1)
	first, record, err := Reconcile(input, nil, native, []string{old, user})
	if err != nil {
		t.Fatal(err)
	}
	owned, err := readOwnership(record)
	if err != nil || !reflect.DeepEqual(owned.OwnedPatterns, []string{literalPattern(filepath.ToSlash(old))}) {
		t.Fatalf("pre-existing user pattern claimed: %s / %v", record, err)
	}
	// A user explicitly duplicates the generated old pattern between launches.
	object, entries, _ := exclusionSettings(first)
	entries = append(entries, literalPattern(filepath.ToSlash(old)))
	object["claudeMdExcludes"], _ = json.Marshal(entries)
	edited, _ := json.Marshal(object)
	second, secondRecord, err := Reconcile(edited, record, native, []string{next})
	if err != nil {
		t.Fatal(err)
	}
	entries = testExclusions(t, second)
	count := map[string]int{}
	for _, entry := range entries {
		count[entry]++
	}
	if count[literalPattern(filepath.ToSlash(old))] != 1 || count[userPattern] != 1 || count["**/vendor/CLAUDE.md"] != 1 || count[literalPattern(filepath.ToSlash(next))] != 1 || !bytes.Contains(second, []byte("1e999")) {
		t.Fatalf("refresh lost user state or retained wrong owned count: %s", second)
	}
	again, againRecord, err := Reconcile(second, secondRecord, native, []string{next})
	if err != nil || !bytes.Equal(again, second) || !bytes.Equal(againRecord, secondRecord) {
		t.Fatal("unchanged policy is not byte-idempotent")
	}
}

// Missing ownership state never reclassifies a matching user exclusion as
// generated. Later project changes therefore preserve that existing choice.
func TestReconcileConservativeMigration(t *testing.T) {
	base := t.TempDir()
	ancestor := filepath.Join(base, "CLAUDE.md")
	pattern := literalPattern(filepath.ToSlash(ancestor))
	input, _ := json.Marshal(map[string]any{"claudeMdExcludes": []string{pattern}})
	first, record, err := Reconcile(input, nil, filepath.Join(base, ".claude", "CLAUDE.md"), []string{ancestor})
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := Reconcile(first, record, filepath.Join(base, ".claude", "CLAUDE.md"), nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range testExclusions(t, second) {
		found = found || entry == pattern
	}
	if !found {
		t.Fatal("migration removed pre-existing user choice")
	}
}

// The operator-native exception survives a HOME project root, while HOME's
// project CLAUDE.md and AGENTS.md receive no newly generated exclusion. This
// proves the emitted policy boundary, not native loader behavior.
func TestReconcileNativeExceptionAtHomeRoot(t *testing.T) {
	home := t.TempDir()
	managed := filepath.Join(home, "managed", "CLAUDE.md")
	ancestors, err := AncestorFiles(home, home, managed)
	if err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(home, ".claude", "CLAUDE.md")
	settings, _, err := Reconcile(nil, nil, native, ancestors)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, entry := range testExclusions(t, settings) {
		seen[entry] = true
	}
	if !seen[literalPattern(filepath.ToSlash(native))] || !seen[literalPattern(filepath.ToSlash(filepath.Join(home, ".claude", "rules")))+"/**"] {
		t.Fatal("operator-native exception omitted")
	}
	for _, name := range []string{"CLAUDE.md", "AGENTS.md", "CLAUDE.local.md", filepath.Join(".claude", "AGENTS.md")} {
		if seen[literalPattern(filepath.ToSlash(filepath.Join(home, name)))] {
			t.Fatalf("project instruction excluded: %s", name)
		}
	}
}

// Unknown or malformed ownership cannot remove user exclusions or yield either
// publication payload; valid migration and refresh controls are adjacent above.
func TestReconcileRefusesInvalidOwnership(t *testing.T) {
	native := filepath.Join(t.TempDir(), ".claude", "CLAUDE.md")
	for _, record := range []string{
		``, `null`, `{}`, `{"version":2,"owned_patterns":[]}`, `{"version":1,"owned_patterns":null}`,
		`{"version":1,"owned_patterns":[],"extra":true}`, `{"version":1,"version":1,"owned_patterns":[]}`,
		`{"version":1,"owned_patterns":["**/CLAUDE.md"]}`, `{"version":1,"owned_patterns":["relative/CLAUDE.md"]}`,
		`{"version":1,"owned_patterns":[1]}`,
	} {
		settings, state, err := Reconcile([]byte(`{"claudeMdExcludes":["**/chosen/**"]}`), []byte(record), native, nil)
		if err == nil || settings != nil || state != nil {
			t.Fatalf("invalid state produced a publication: %s", record)
		}
	}
	for _, input := range []string{`null`, `{"claudeMdExcludes":null}`, `{"claudeMdExcludes":[1]}`, `{"x":1,"x":2}`} {
		if settings, state, err := Reconcile([]byte(input), nil, native, nil); err == nil || settings != nil || state != nil {
			t.Fatal("invalid settings produced a publication")
		}
	}
	if settings, state, err := Reconcile(nil, nil, native, []string{filepath.Join(filepath.Dir(native), "AGENTS.md")}); err == nil || settings != nil || state != nil {
		t.Fatal("AGENTS filename accepted as generated exclusion")
	}
}
