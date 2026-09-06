package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
)

// Production entry points under test: Install, Use, Sync, List,
// ScopedCurrents, Current.
//
// F16 covers the machine-scope switch over a scope record (environments
// §9.3): a machine-scope use — and, through the same path, an install
// activation — must skip adapters that carry a scope record, so the recorded
// state and the materialized bytes never disagree. A mutant that restores
// the unconditional machine pass (adapters = Adapters for the unnarrowed
// case) overwrites the scoped home with the machine profile and must fail
// these tests.

// scopeFixture installs alpha (machine current) and beta (scoped to
// codex_cli), leaving the machine current on alpha.
func scopeFixture(t *testing.T, home string) {
	t.Helper()
	alpha, beta := t.TempDir(), t.TempDir()
	writePackage(t, alpha, "alpha", "1.0.0", "alpha\n")
	writePackage(t, beta, "beta", "1.0.0", "beta\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: alpha}); err != nil {
		t.Fatalf("install alpha: %v", err)
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: beta}); err != nil {
		t.Fatalf("install beta: %v", err)
	}
	if _, err := Use(home, "beta", "codex_cli", "", false); err != nil {
		t.Fatalf("scoped use beta: %v", err)
	}
}

func homeBytes(t *testing.T, dir, file string) string {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(dir, file)) // #nosec G304 -- test home
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func homeMarker(t *testing.T, dir string) *envmarker.Marker {
	t.Helper()
	marker, err := envmarker.Read(dir)
	if err != nil || marker == nil {
		t.Fatalf("marker %+v %v", marker, err)
	}
	return marker
}

// TestMachineUseSkipsScopedAdapter drives the production Use for a
// machine-scope switch while env:codex_cli is scoped to beta: the codex home
// stays on beta (bytes and marker), the other three homes move to gamma, the
// machine current moves to gamma, and the listing still reports
// env:codex_cli=beta. A mutant that restores the unconditional machine pass
// overwrites the codex home with gamma and must fail this test.
func TestMachineUseSkipsScopedAdapter(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	scopeFixture(t, home)
	gamma := t.TempDir()
	writePackage(t, gamma, "gamma", "1.0.0", "gamma\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: gamma}); err != nil {
		t.Fatalf("install gamma: %v", err)
	}
	results, err := Use(home, "gamma", "", "", false)
	if err != nil {
		t.Fatalf("machine use gamma: %v (%+v)", err, results)
	}
	for _, result := range results {
		if result.Adapter == "codex_cli" {
			t.Fatalf("machine pass must skip the scoped adapter, results %+v", results)
		}
		if !result.OK {
			t.Fatalf("result %+v", result)
		}
	}
	if len(results) != len(Adapters)-1 {
		t.Fatalf("results %+v, want every adapter but the scoped one", results)
	}
	if current, _ := Current(home); current != "gamma" {
		t.Fatalf("current=%q, want gamma", current)
	}
	if scoped, err := ScopedCurrents(home); err != nil || scoped["env:codex_cli"] != "beta" {
		t.Fatalf("scoped %+v %v", scoped, err)
	}
	if got := homeBytes(t, homes["codex_cli"], "AGENTS.md"); !strings.Contains(got, "## Context: beta 1.0.0") {
		t.Fatalf("codex home does not carry beta:\n%s", got)
	}
	if marker := homeMarker(t, homes["codex_cli"]); marker.Profile.Name != "beta" {
		t.Fatalf("codex marker names %q, want beta", marker.Profile.Name)
	}
	for id, file := range map[string]string{"claude_code": "CLAUDE.md", "opencode": "AGENTS.md", "pi": "AGENTS.md"} {
		if got := homeBytes(t, homes[id], file); !strings.Contains(got, "## Context: gamma 1.0.0") {
			t.Fatalf("%s home does not carry gamma:\n%s", id, got)
		}
		if marker := homeMarker(t, homes[id]); marker.Profile.Name != "gamma" {
			t.Fatalf("%s marker names %q, want gamma", id, marker.Profile.Name)
		}
	}
	profiles, err := List(home)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Info{}
	for _, info := range profiles {
		byName[info.Name] = info
	}
	if !byName["gamma"].Current {
		t.Fatalf("gamma is not the machine current: %+v", byName["gamma"])
	}
	found := false
	for _, scope := range byName["beta"].ScopedFor {
		if scope == "env:codex_cli" {
			found = true
		}
	}
	if !found {
		t.Fatalf("beta ScopedFor %+v, want env:codex_cli", byName["beta"].ScopedFor)
	}
}

// TestInstallUseSkipsScopedAdapter drives the same shape through the
// production Install activation: install --use performs the §9.2 switch, so
// it skips the scoped home exactly like profile use does. The mutant from
// TestMachineUseSkipsScopedAdapter must fail this test too.
func TestInstallUseSkipsScopedAdapter(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	scopeFixture(t, home)
	gamma := t.TempDir()
	writePackage(t, gamma, "gamma", "1.0.0", "gamma\n")
	info, activated, _, err := Install(home, InstallOptions{Operand: gamma, Use: true})
	if err != nil {
		t.Fatalf("install gamma --use: %v", err)
	}
	if !activated {
		t.Fatal("install --use must report activated")
	}
	for _, result := range info.Activation {
		if result.Adapter == "codex_cli" {
			t.Fatalf("activation must skip the scoped adapter, results %+v", info.Activation)
		}
		if !result.OK {
			t.Fatalf("activation %+v", result)
		}
	}
	if current, _ := Current(home); current != "gamma" {
		t.Fatalf("current=%q, want gamma", current)
	}
	if got := homeBytes(t, homes["codex_cli"], "AGENTS.md"); !strings.Contains(got, "## Context: beta 1.0.0") {
		t.Fatalf("codex home does not carry beta:\n%s", got)
	}
	if marker := homeMarker(t, homes["codex_cli"]); marker.Profile.Name != "beta" {
		t.Fatalf("codex marker names %q, want beta", marker.Profile.Name)
	}
	if got := homeBytes(t, homes["claude_code"], "CLAUDE.md"); !strings.Contains(got, "## Context: gamma 1.0.0") {
		t.Fatalf("claude home does not carry gamma:\n%s", got)
	}
}

// TestScopedUseStillSwitchesOnlyThatHome pins the other half of the shape:
// a narrowed switch still writes exactly its own home and moves no machine
// current, even after a machine-scope switch ran over a scope record.
func TestScopedUseStillSwitchesOnlyThatHome(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	scopeFixture(t, home)
	gamma, delta := t.TempDir(), t.TempDir()
	writePackage(t, gamma, "gamma", "1.0.0", "gamma\n")
	writePackage(t, delta, "delta", "1.0.0", "delta\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: gamma}); err != nil {
		t.Fatalf("install gamma: %v", err)
	}
	if _, err := Use(home, "gamma", "", "", false); err != nil {
		t.Fatalf("machine use gamma: %v", err)
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: delta}); err != nil {
		t.Fatalf("install delta: %v", err)
	}
	results, err := Use(home, "delta", "codex_cli", "", false)
	if err != nil {
		t.Fatalf("scoped use delta: %v (%+v)", err, results)
	}
	if len(results) != 1 || results[0].Adapter != "codex_cli" || !results[0].OK {
		t.Fatalf("scoped results %+v, want exactly the codex entry", results)
	}
	if current, _ := Current(home); current != "gamma" {
		t.Fatalf("scoped use moved the machine current to %q", current)
	}
	if scoped, err := ScopedCurrents(home); err != nil || scoped["env:codex_cli"] != "delta" {
		t.Fatalf("scoped %+v %v", scoped, err)
	}
	if got := homeBytes(t, homes["codex_cli"], "AGENTS.md"); !strings.Contains(got, "## Context: delta 1.0.0") {
		t.Fatalf("codex home does not carry delta:\n%s", got)
	}
	if marker := homeMarker(t, homes["codex_cli"]); marker.Profile.Name != "delta" {
		t.Fatalf("codex marker names %q, want delta", marker.Profile.Name)
	}
	if got := homeBytes(t, homes["claude_code"], "CLAUDE.md"); !strings.Contains(got, "## Context: gamma 1.0.0") {
		t.Fatalf("claude home no longer carries gamma:\n%s", got)
	}
}

// TestSyncWritesScopedHomeOnce pins the sync half of the fix: the machine
// pass skips the scoped home and the scoped pass writes it once, so every
// adapter appears exactly once in the results and every home gains exactly
// one backup generation. The unconditional-pass mutant writes the scoped
// home twice (two results entries, two generations) and must fail this
// test.
func TestSyncWritesScopedHomeOnce(t *testing.T) {
	home := t.TempDir()
	homes := pinHomes(t)
	scopeFixture(t, home)
	gamma := t.TempDir()
	writePackage(t, gamma, "gamma", "1.0.0", "gamma\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: gamma}); err != nil {
		t.Fatalf("install gamma: %v", err)
	}
	if _, err := Use(home, "gamma", "", "", false); err != nil {
		t.Fatalf("machine use gamma: %v", err)
	}
	for _, dir := range homes {
		if err := os.RemoveAll(filepath.Join(dir, ".agent-environment-backup")); err != nil {
			t.Fatal(err)
		}
	}
	results, err := Sync(home)
	if err != nil {
		t.Fatalf("sync: %v (%+v)", err, results)
	}
	seen := map[string]int{}
	for _, result := range results {
		seen[result.Adapter]++
		if !result.OK {
			t.Fatalf("result %+v", result)
		}
	}
	for _, adapter := range Adapters {
		if seen[adapter.ID] != 1 {
			t.Fatalf("adapter %s appears %d times in %+v, want exactly once", adapter.ID, seen[adapter.ID], results)
		}
	}
	for id, dir := range homes {
		entries, err := os.ReadDir(filepath.Join(dir, ".agent-environment-backup"))
		if err != nil {
			t.Fatalf("%s backups: %v", id, err)
		}
		if len(entries) != 1 {
			t.Fatalf("%s has %d backup generations, want exactly one", id, len(entries))
		}
	}
	if got := homeBytes(t, homes["codex_cli"], "AGENTS.md"); !strings.Contains(got, "## Context: beta 1.0.0") {
		t.Fatalf("codex home does not carry beta after sync:\n%s", got)
	}
	if marker := homeMarker(t, homes["codex_cli"]); marker.Profile.Name != "beta" {
		t.Fatalf("codex marker names %q after sync, want beta", marker.Profile.Name)
	}
}

// TestScopedRecordFilenameIsPortable narrows the scope-record gate
// (environments §9.3): the scope record file must name no character Windows
// reserves, while ScopedCurrents keeps reporting the scope key with its
// colon. A mutant that records the raw scope key as the filename writes
// "env:codex_cli" and must fail this test on every platform; on Windows it
// additionally fails the switch itself with profile_use_partial.
func TestScopedRecordFilenameIsPortable(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	first, second := t.TempDir(), t.TempDir()
	writePackage(t, first, "one", "1.0.0", "one\n")
	writePackage(t, second, "two", "1.0.0", "two\n")
	if _, _, _, err := Install(home, InstallOptions{Operand: first}); err != nil {
		t.Fatalf("install one: %v", err)
	}
	if _, _, _, err := Install(home, InstallOptions{Operand: second}); err != nil {
		t.Fatalf("install two: %v", err)
	}
	if _, err := Use(home, "two", "codex_cli", "", false); err != nil {
		t.Fatalf("scoped use two: %v", err)
	}
	entries, err := os.ReadDir(ScopedDir(home))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("scope records %+v, want exactly one", entries)
	}
	if name := entries[0].Name(); strings.ContainsAny(name, `<>:"/\|?*`) {
		t.Fatalf("scope record %q names a Windows-reserved character", name)
	}
	if name := entries[0].Name(); name != scopeFileName("env:codex_cli") {
		t.Fatalf("scope record %q, want the encoded scope key", name)
	}
	scoped, err := ScopedCurrents(home)
	if err != nil {
		t.Fatal(err)
	}
	if scoped["env:codex_cli"] != "two" {
		t.Fatalf("scoped %+v, want env:codex_cli=two", scoped)
	}
	// The mapping round-trips, and a pre-encoding literal record decodes to
	// itself, so manager homes created where a colon is legal keep reading.
	if scopeKeyName(scopeFileName("env:codex_cli")) != "env:codex_cli" {
		t.Fatal("scope filename mapping does not round-trip")
	}
	if scopeKeyName("env:codex_cli") != "env:codex_cli" {
		t.Fatal("pre-encoding literal record does not decode to itself")
	}
	// A stale pre-encoding spelling never shadows its encoded successor:
	// both filenames below are free of reserved characters on every
	// platform, yet decode to one key the encoded record must win.
	stale, fresh := "stale%key", scopeFileName("stale%key")
	if stale == fresh {
		t.Fatalf("want distinct stale and encoded spellings, got %q", stale)
	}
	if err := os.MkdirAll(ScopedDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ScopedDir(home), stale), []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ScopedDir(home), fresh), []byte("fresh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scoped, err = ScopedCurrents(home)
	if err != nil {
		t.Fatal(err)
	}
	if scoped["stale%key"] != "fresh" {
		t.Fatalf("scoped %+v, want the encoded record to win", scoped)
	}
}
