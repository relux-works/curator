package codex_cli

import (
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
)

func TestAdapterDeclaresV1CodexSurfaces(t *testing.T) {
	adapter := New()
	if adapter.ID() != ID {
		t.Fatalf("ID() = %q, want %q", adapter.ID(), ID)
	}
	if adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("APIVersion() = %d, want %d", adapter.APIVersion(), harnessapi.AdapterAPIVersionV1)
	}
	want, err := harnessapi.NewCapabilitySet(
		harnessapi.ManagedHome,
		harnessapi.InstructionOutput,
		harnessapi.SkillProjection,
		harnessapi.MCPInspection,
		harnessapi.CredentialMetadata,
		harnessapi.ReleaseProbing,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := adapter.Capabilities().Values(); !reflect.DeepEqual(got, want.Values()) {
		t.Fatalf("Capabilities() = %v, want %v", got, want.Values())
	}
	probe, err := adapter.ReleaseProbe(harnessapi.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if probe.ExecutableAlias() != "codex" || !reflect.DeepEqual(probe.Arguments(), []string{"--version"}) || probe.Timeout() != harnessapi.ProbeTimeoutV1 || len(probe.Environment()) != 0 {
		t.Fatalf("ReleaseProbe() = %#v", probe)
	}
}

func TestHomeAndInstructions(t *testing.T) {
	adapter := New()
	home, err := adapter.Home(harnessapi.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if home.EnvironmentVariable() != "CODEX_HOME" || home.ParentValued() {
		t.Fatalf("Home() selector/parent = %q/%t", home.EnvironmentVariable(), home.ParentValued())
	}
	if got := pathStrings(home.Roots()); !reflect.DeepEqual(got, []string{".codex"}) {
		t.Fatalf("Home().Roots() = %v", got)
	}

	spec, err := adapter.Instructions(harnessapi.Context{}, harnessapi.Monolithic)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Form() != harnessapi.Monolithic || !reflect.DeepEqual(pathStrings(spec.Paths()), []string{"AGENTS.md"}) || len(spec.Notice()) != 0 {
		t.Fatalf("Instructions() = form:%q paths:%v notice:%q", spec.Form(), pathStrings(spec.Paths()), spec.Notice())
	}
	if _, err := adapter.Instructions(harnessapi.Context{}, harnessapi.Referenced); refusalCode(err) != "environment_form_unsupported" {
		t.Fatalf("referenced instructions error = %v, want unsupported-form refusal", err)
	}
}

func TestSkillsAndMCP(t *testing.T) {
	adapter := New()
	for _, scope := range []harnessapi.Scope{harnessapi.UserScope, harnessapi.ProjectScope} {
		skills, err := adapter.Skills(harnessapi.Context{}, scope)
		if err != nil {
			t.Fatalf("Skills(%q): %v", scope, err)
		}
		wantSkills := []string{"skills"}
		if scope == harnessapi.ProjectScope {
			wantSkills = []string{".codex/skills"}
		}
		if got := pathStrings(skills.Paths()); !reflect.DeepEqual(got, wantSkills) {
			t.Fatalf("Skills(%q).Paths() = %v", scope, got)
		}
		if skills.Scope() != scope || skills.NativeDiscovery() || !skills.Mirror() {
			t.Fatalf("Skills(%q) scope/native/mirror = %q/%t/%t", scope, skills.Scope(), skills.NativeDiscovery(), skills.Mirror())
		}
		gotIgnores := pathStrings(skills.IgnorePaths())
		if scope == harnessapi.ProjectScope {
			wantIgnores := []string{".codex/skills"}
			if !reflect.DeepEqual(gotIgnores, wantIgnores) {
				t.Fatalf("Skills(%q).IgnorePaths() = %v, want %v", scope, gotIgnores, wantIgnores)
			}
		} else if len(gotIgnores) != 0 {
			t.Fatalf("Skills(%q).IgnorePaths() = %v, want no paths", scope, gotIgnores)
		}

		mcp, err := adapter.MCP(harnessapi.Context{}, scope)
		if err != nil {
			t.Fatalf("MCP(%q): %v", scope, err)
		}
		wantMCP := []string{"config.toml"}
		if scope == harnessapi.ProjectScope {
			wantMCP = []string{".codex/config.toml"}
		}
		if got := pathStrings(mcp.Paths()); !reflect.DeepEqual(got, wantMCP) {
			t.Fatalf("MCP(%q).Paths() = %v", scope, got)
		}
		if mcp.Format() != harnessapi.TOML || !mcp.Inspect() || mcp.Write() || mcp.RequiresTrust() != (scope == harnessapi.ProjectScope) {
			t.Fatalf("MCP(%q) format/inspect/write/trust = %q/%t/%t/%t", scope, mcp.Format(), mcp.Inspect(), mcp.Write(), mcp.RequiresTrust())
		}
	}
	if _, err := adapter.Skills(harnessapi.Context{}, "invalid"); refusalCode(err) != "harness_scope_unsupported" {
		t.Fatalf("Skills(invalid) error = %v, want scope refusal", err)
	}
	if _, err := adapter.MCP(harnessapi.Context{}, "invalid"); refusalCode(err) != "harness_scope_unsupported" {
		t.Fatalf("MCP(invalid) error = %v, want scope refusal", err)
	}
}

func TestCredentialMetadataAndUnportedV2Operations(t *testing.T) {
	adapter := New()
	route, err := adapter.Credentials(harnessapi.Context{})
	if err != nil {
		t.Fatal(err)
	}
	if got := pathStrings(route.Paths()); !reflect.DeepEqual(got, []string{"auth.json"}) || route.Strategy() != "keyring-preferred" || !route.Qualified() {
		t.Fatalf("Credentials() = paths:%v strategy:%q qualified:%t", got, route.Strategy(), route.Qualified())
	}
	if _, err := adapter.Settings(harnessapi.Context{}); refusalCode(err) != harnessapi.OperationNotPorted {
		t.Fatalf("Settings() error = %v, want not-ported refusal", err)
	}
	if _, err := adapter.PlanSeed(harnessapi.Context{}, harnessapi.Snapshot{}, harnessapi.Request{}); refusalCode(err) != harnessapi.OperationNotPorted {
		t.Fatalf("PlanSeed() error = %v, want not-ported refusal", err)
	}
}

func TestNamedPathsRejectsEscapingRelativePath(t *testing.T) {
	if _, err := namedPaths("../outside"); err == nil {
		t.Fatal("namedPaths accepted a path outside the adapter root")
	}
}

func pathStrings(paths harnessapi.NamedPaths) []string {
	values := paths.Paths()
	got := make([]string, 0, len(values))
	for _, path := range values {
		got = append(got, path.String())
	}
	return got
}

func refusalCode(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}
