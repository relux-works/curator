package claude_code

import (
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
)

func claudeContext(t *testing.T, platform string, scope harnessapi.Scope) harnessapi.Context {
	t.Helper()
	release, err := harnessapi.NewVerifiedRelease("2.1.261")
	if err != nil {
		t.Fatal(err)
	}
	context, err := harnessapi.NewContext(platform, scope, release)
	if err != nil {
		t.Fatal(err)
	}
	return context
}

func paths(t *testing.T, named harnessapi.NamedPaths) []string {
	t.Helper()
	result := []string{}
	for _, path := range named.Paths() {
		result = append(result, path.String())
	}
	return result
}

func TestAdapterMetadataAndProbe(t *testing.T) {
	adapter := Adapter{}
	if adapter.ID() != "claude_code" || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter identity/version = %q/%d", adapter.ID(), adapter.APIVersion())
	}
	capabilities := adapter.Capabilities().Values()
	want := []harnessapi.Capability{
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
		harnessapi.NativeMCPWrite,
		harnessapi.CredentialMetadata,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	}
	if !reflect.DeepEqual(capabilities, want) {
		t.Fatalf("capabilities = %v, want %v", capabilities, want)
	}
	probe, err := adapter.ReleaseProbe(harnessapi.Context{})
	if err != nil || probe.ExecutableAlias() != "claude" || !reflect.DeepEqual(probe.Arguments(), []string{"--version"}) || probe.Timeout() != harnessapi.ProbeTimeoutV1 || len(probe.Environment()) != 0 {
		t.Fatalf("release probe = %#v, error %v", probe, err)
	}
}

func TestAdapterHomeInstructionsSkillsAndSettings(t *testing.T) {
	adapter := Adapter{}
	home, err := adapter.Home(harnessapi.Context{})
	if err != nil || home.EnvironmentVariable() != "CLAUDE_CONFIG_DIR" || !reflect.DeepEqual(paths(t, home.Roots()), []string{".claude"}) || home.ParentValued() {
		t.Fatalf("home = %#v, error %v", home, err)
	}
	for _, form := range []harnessapi.Form{harnessapi.Monolithic, harnessapi.Referenced} {
		instructions, err := adapter.Instructions(harnessapi.Context{}, form)
		if err != nil || instructions.Form() != form || !reflect.DeepEqual(paths(t, instructions.Paths()), []string{"CLAUDE.md"}) {
			t.Fatalf("instructions for %q = %#v, error %v", form, instructions, err)
		}
	}
	if _, err := adapter.Instructions(harnessapi.Context{}, "unsupported"); err == nil || err.Error() != "environment_form_unsupported" {
		t.Fatalf("unsupported instruction form accepted: %v", err)
	}

	userSkills, err := adapter.Skills(harnessapi.Context{}, harnessapi.UserScope)
	if err != nil || !reflect.DeepEqual(paths(t, userSkills.Paths()), []string{"skills"}) || len(paths(t, userSkills.IgnorePaths())) != 0 || userSkills.NativeDiscovery() || !userSkills.Mirror() {
		t.Fatalf("user skills = %#v, error %v", userSkills, err)
	}
	projectSkills, err := adapter.Skills(harnessapi.Context{}, harnessapi.ProjectScope)
	if err != nil || !reflect.DeepEqual(paths(t, projectSkills.Paths()), []string{".claude/skills"}) || !reflect.DeepEqual(paths(t, projectSkills.IgnorePaths()), []string{".claude/skills"}) {
		t.Fatalf("project skills = %#v, error %v", projectSkills, err)
	}
	if _, err := adapter.Skills(harnessapi.Context{}, "invalid"); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("unsupported skill scope accepted: %v", err)
	}

	userSettings, err := adapter.Settings(claudeContext(t, "linux", harnessapi.UserScope))
	if err != nil || userSettings.Format() != harnessapi.JSON || !reflect.DeepEqual(paths(t, userSettings.Paths()), []string{"settings.json"}) {
		t.Fatalf("user settings = %#v, error %v", userSettings, err)
	}
	projectSettings, err := adapter.Settings(claudeContext(t, "linux", harnessapi.ProjectScope))
	if err != nil || projectSettings.Format() != harnessapi.JSON || !reflect.DeepEqual(paths(t, projectSettings.Paths()), []string{".claude/settings.json", ".claude/settings.local.json"}) {
		t.Fatalf("project settings = %#v, error %v", projectSettings, err)
	}
	if _, err := adapter.Settings(harnessapi.Context{}); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("unsupported settings scope accepted: %v", err)
	}
	if _, err := namedPaths("../escape"); err == nil {
		t.Fatal("escaping path passed the adapter's path constructor")
	}
}

func TestAdapterMCPCredentialsTrustAndPreflight(t *testing.T) {
	adapter := Adapter{}
	userMCP, err := adapter.MCP(claudeContext(t, "linux", harnessapi.UserScope), harnessapi.UserScope)
	if err != nil || !reflect.DeepEqual(paths(t, userMCP.Paths()), []string{".claude.json"}) || userMCP.Format() != harnessapi.JSON || !userMCP.Inspect() || userMCP.Write() || userMCP.RequiresTrust() {
		t.Fatalf("user MCP = %#v, error %v", userMCP, err)
	}
	projectMCP, err := adapter.MCP(claudeContext(t, "linux", harnessapi.ProjectScope), harnessapi.ProjectScope)
	if err != nil || !reflect.DeepEqual(paths(t, projectMCP.Paths()), []string{".mcp.json"}) || projectMCP.Format() != harnessapi.JSON || !projectMCP.Inspect() || !projectMCP.Write() || !projectMCP.RequiresTrust() {
		t.Fatalf("project MCP = %#v, error %v", projectMCP, err)
	}
	if _, err := adapter.MCP(harnessapi.Context{}, "invalid"); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("unsupported MCP scope accepted: %v", err)
	}

	linuxRoute, err := adapter.Credentials(claudeContext(t, "linux", harnessapi.UserScope))
	if err != nil || linuxRoute.Strategy() != "file-link" || !linuxRoute.Qualified() || !reflect.DeepEqual(paths(t, linuxRoute.Paths()), []string{".credentials.json"}) {
		t.Fatalf("Linux credential route = %#v, error %v", linuxRoute, err)
	}
	if _, err := adapter.Credentials(claudeContext(t, "darwin", harnessapi.UserScope)); err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("Darwin Keychain route unexpectedly fit v1: %v", err)
	}
	windowsRoute, err := adapter.Credentials(claudeContext(t, "windows", harnessapi.UserScope))
	if err != nil || windowsRoute.Strategy() != "unsupported" || windowsRoute.Qualified() || len(paths(t, windowsRoute.Paths())) != 0 {
		t.Fatalf("Windows credential route = %#v, error %v", windowsRoute, err)
	}

	userTrust, err := adapter.Trust(claudeContext(t, "linux", harnessapi.UserScope))
	if err != nil || userTrust.Required() || !userTrust.ObservationOnly() {
		t.Fatalf("user trust = %#v, error %v", userTrust, err)
	}
	projectTrust, err := adapter.Trust(claudeContext(t, "linux", harnessapi.ProjectScope))
	if err != nil || !projectTrust.Required() || !projectTrust.ObservationOnly() {
		t.Fatalf("project trust = %#v, error %v", projectTrust, err)
	}
	preflight, err := adapter.Preflight(claudeContext(t, "linux", harnessapi.ProjectScope), harnessapi.Snapshot{}, harnessapi.Request{})
	if err != nil || preflight == nil || len(preflight) != 0 {
		t.Fatalf("preflight = %v, error %v", preflight, err)
	}
}

func TestUnportedOperationsRemainExplicit(t *testing.T) {
	adapter := Adapter{}
	context, snapshot, request := harnessapi.Context{}, harnessapi.Snapshot{}, harnessapi.Request{}
	if _, err := adapter.PlanSeed(context, snapshot, request); err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("dynamic seed merge was not left with its shared owner: %v", err)
	}
	if _, err := adapter.ProjectContext(context, snapshot, request); err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("project context was not left with its shared owner: %v", err)
	}
	if _, err := adapter.Status(context, snapshot); err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("status was not left with its shared owner: %v", err)
	}
	if _, err := adapter.Drift(context, snapshot); err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("drift was not left with its shared owner: %v", err)
	}
}
