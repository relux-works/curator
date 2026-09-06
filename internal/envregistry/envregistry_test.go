package envregistry

import (
	"strings"
	"testing"
)

func mustAdapter(t *testing.T, id string) Adapter {
	t.Helper()
	adapter, err := ByID(id)
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func TestRegistryClosed(t *testing.T) {
	if len(Registry) != 4 {
		t.Fatalf("the revision-1 registry holds %d adapters, want 4", len(Registry))
	}
	seen := map[string]bool{}
	for _, adapter := range Registry {
		if seen[adapter.ID] {
			t.Fatalf("duplicate adapter %s", adapter.ID)
		}
		seen[adapter.ID] = true
		if adapter.EnvVar == "" || adapter.RootTarget == "" {
			t.Fatalf("adapter %s names no home variable or root target", adapter.ID)
		}
	}
	for _, id := range []string{ClaudeCode, CodexCLI, OpenCode, Pi} {
		if !seen[id] {
			t.Fatalf("adapter %s is missing from the registry", id)
		}
	}
}

// TestUnknownEnvironment narrows the registry gate: an explicit operand
// naming an unregistered environment must fail with environment_unknown,
// and weakening the lookup to warn-and-continue must fail this test.
func TestUnknownEnvironment(t *testing.T) {
	_, err := ByID("cursor")
	if err == nil || !strings.Contains(err.Error(), DiagUnknown) {
		t.Fatalf("unregistered environment must fail with %s, got %v", DiagUnknown, err)
	}
}

func TestUnknownTarget(t *testing.T) {
	_, err := TargetByID("nope")
	if err == nil || !strings.Contains(err.Error(), DiagTargetUnknown) {
		t.Fatalf("undeclared target must fail with %s, got %v", DiagTargetUnknown, err)
	}
}

func TestFormDefaults(t *testing.T) {
	for id, want := range map[string]string{ClaudeCode: FormMonolithic, CodexCLI: FormMonolithic, OpenCode: FormMonolithic, Pi: FormMonolithic} {
		form, err := mustAdapter(t, id).ResolveForm("")
		if err != nil || form != want {
			t.Fatalf("%s default form %q, want %q (err %v)", id, form, want, err)
		}
	}
}

// TestFormUnsupported narrows the §7.2 gate: referenced for codex_cli or
// pi must fail, never fall back silently — a fallback would emit bytes the
// operator did not ask for under a form the tool cannot read.
func TestFormUnsupported(t *testing.T) {
	for _, id := range []string{CodexCLI, Pi} {
		if _, err := mustAdapter(t, id).ResolveForm(FormReferenced); err == nil || !strings.Contains(err.Error(), DiagFormUnsupported) {
			t.Fatalf("%s referenced must fail with %s, got %v", id, DiagFormUnsupported, err)
		}
	}
	for _, id := range []string{ClaudeCode, OpenCode} {
		if form, err := mustAdapter(t, id).ResolveForm(FormReferenced); err != nil || form != FormReferenced {
			t.Fatalf("%s referenced must resolve, got %q (%v)", id, form, err)
		}
	}
	if _, err := mustAdapter(t, ClaudeCode).ResolveForm("condensed"); err == nil {
		t.Fatal("an unknown form name must fail, not fall back")
	}
}

// TestIsolationMatrix drives the §7.4 matrix on every platform: opencode
// refuses isolated everywhere; claude_code on macOS defaults to isolated
// at the pinned release and refuses shared; below the pinned release it
// refuses isolated; elsewhere shared is the default and isolated is
// available.
func TestIsolationMatrix(t *testing.T) {
	opencode := mustAdapter(t, OpenCode)
	if _, err := opencode.resolveIsolation("linux", IsolationIsolated, true); err == nil || !strings.Contains(err.Error(), DiagIsolatedUnsupported) {
		t.Fatalf("opencode isolated must fail with %s, got %v", DiagIsolatedUnsupported, err)
	}
	claude := mustAdapter(t, ClaudeCode)
	mode, err := claude.resolveIsolation("darwin", "", true)
	if err != nil || mode != IsolationIsolated {
		t.Fatalf("claude macOS default at pinned is isolated, got %q (%v)", mode, err)
	}
	if _, err := claude.resolveIsolation("darwin", IsolationShared, true); err == nil || !strings.Contains(err.Error(), DiagSharedUnsupported) {
		t.Fatalf("claude macOS shared at pinned must fail with %s, got %v", DiagSharedUnsupported, err)
	}
	if _, err := claude.resolveIsolation("darwin", IsolationIsolated, false); err == nil || !strings.Contains(err.Error(), DiagIsolatedUnsupported) {
		t.Fatalf("claude macOS isolated below pinned must fail with %s, got %v", DiagIsolatedUnsupported, err)
	}
	for _, goos := range []string{"linux", "windows"} {
		mode, err := claude.resolveIsolation(goos, "", true)
		if err != nil || mode != IsolationShared {
			t.Fatalf("claude %s default is shared, got %q (%v)", goos, mode, err)
		}
		if mode, err := claude.resolveIsolation(goos, IsolationIsolated, true); err != nil || mode != IsolationIsolated {
			t.Fatalf("claude %s isolated is available, got %q (%v)", goos, mode, err)
		}
	}
	for _, id := range []string{CodexCLI, Pi} {
		adapter := mustAdapter(t, id)
		if mode, err := adapter.resolveIsolation("linux", "", true); err != nil || mode != IsolationShared {
			t.Fatalf("%s default is shared, got %q (%v)", id, mode, err)
		}
		if mode, err := adapter.resolveIsolation("linux", IsolationIsolated, true); err != nil || mode != IsolationIsolated {
			t.Fatalf("%s isolated is available, got %q (%v)", id, mode, err)
		}
	}
	if _, err := mustAdapter(t, Pi).resolveIsolation("linux", "airgapped", true); err == nil {
		t.Fatal("an unknown isolation mode must fail")
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		have, want string
		comparison int
		ok         bool
	}{
		{"2.1.261", "2.1.261", 0, true},
		{"2.1.262", "2.1.261", 1, true},
		{"2.1.260", "2.1.261", -1, true},
		{"2.2", "2.1.261", 1, true},
		{"2.1.261-beta", "2.1.261", -1, true},
		{"0.153.2", "0.153.2", 0, true},
		{"unknown", "2.1.261", 0, false},
		{"", "2.1.261", 0, false},
	}
	for _, tc := range cases {
		got, ok := CompareVersions(tc.have, tc.want)
		if got != tc.comparison || ok != tc.ok {
			t.Fatalf("CompareVersions(%q, %q) = (%d, %v), want (%d, %v)", tc.have, tc.want, got, ok, tc.comparison, tc.ok)
		}
	}
}

func TestAtOrAbovePinned(t *testing.T) {
	claude := mustAdapter(t, ClaudeCode)
	if ok, known := claude.AtOrAbovePinned("2.1.261"); !ok || !known {
		t.Fatal("the pinned release itself is at the pin")
	}
	if ok, known := claude.AtOrAbovePinned("2.1.200"); ok || !known {
		t.Fatal("a release below the pin is known-below")
	}
	if _, known := claude.AtOrAbovePinned("unknown"); known {
		t.Fatal("an unknown release is never reported as matching")
	}
	opencode := mustAdapter(t, OpenCode)
	if _, known := opencode.AtOrAbovePinned("1.0.0"); known {
		t.Fatal("an adapter with no recorded release reports every release as unknown")
	}
}

func TestMachineConfigDefaults(t *testing.T) {
	config := DefaultMachineConfig()
	if config.Retention() != DefaultBackupRetention {
		t.Fatalf("default retention %d", config.Retention())
	}
	if len(config.XDGSeedAllowlist) != 3 {
		t.Fatalf("default XDG allowlist %v", config.XDGSeedAllowlist)
	}
	if config.PassableEnvNames != nil {
		t.Fatal("passable_env_names defaults to unbounded (nil)")
	}
	form, err := config.EffectiveForm(mustAdapter(t, CodexCLI))
	if err != nil || form != FormMonolithic {
		t.Fatalf("default codex form %q (%v)", form, err)
	}
}

func TestTargetConsent(t *testing.T) {
	target, err := TargetByID("xcode-coding-assistant")
	if err != nil {
		t.Fatal(err)
	}
	config := DefaultMachineConfig()
	if err := config.TargetConsent(target); err == nil || !strings.Contains(err.Error(), DiagTargetConsent) {
		t.Fatalf("the first auto write without consent must stop, got %v", err)
	}
	config.TargetParticipation[target.ID] = "enabled"
	if err := config.TargetConsent(target); err != nil {
		t.Fatalf("an explicit enable is consent by that act: %v", err)
	}
	config = DefaultMachineConfig()
	config.TargetConsented[target.ID] = true
	if err := config.TargetConsent(target); err != nil {
		t.Fatalf("recorded consent admits the write: %v", err)
	}
	if config.TargetParticipates(target, func(string) bool { return false }) {
		t.Fatal("auto participation without a probe path materializes nothing")
	}
	if !config.TargetParticipates(target, func(string) bool { return true }) {
		t.Fatal("auto participation with a probe path participates")
	}
}

func TestShadowAcknowledgment(t *testing.T) {
	config := DefaultMachineConfig()
	if config.ShadowAcknowledges(Pi, "AGENTS.override.md") {
		t.Fatal("shadowing defaults to fail-closed: no acknowledgment is recorded")
	}
	config.ShadowAcknowledged = []ShadowAck{{Env: Pi, Path: "AGENTS.override.md"}}
	if !config.ShadowAcknowledges(Pi, "AGENTS.override.md") {
		t.Fatal("a recorded acknowledgment downgrades exactly that row")
	}
	if config.ShadowAcknowledges(Pi, "OTHER.md") {
		t.Fatal("an acknowledgment covers exactly the recorded path")
	}
}
