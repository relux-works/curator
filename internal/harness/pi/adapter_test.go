package pi

import (
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
)

func TestAdapterV1Values(t *testing.T) {
	adapter := New()
	if adapter.ID() != ID || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter identity/version = %q/%d, want %q/%d", adapter.ID(), adapter.APIVersion(), ID, harnessapi.AdapterAPIVersionV1)
	}
	ctx, err := harnessapi.NewContext("darwin", harnessapi.UserScope, harnessapi.VerifiedRelease{})
	if err != nil {
		t.Fatal(err)
	}
	probe, err := adapter.ReleaseProbe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if probe.ExecutableAlias() != "pi" || len(probe.Arguments()) != 1 || probe.Arguments()[0] != "--version" || probe.Timeout() != harnessapi.ProbeTimeoutV1 || len(probe.Environment()) != 0 {
		t.Fatalf("release probe = %#v, want pi --version with the v1 timeout and no environment overrides", probe)
	}

	home, err := adapter.Home(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if home.EnvironmentVariable() != "PI_CODING_AGENT_DIR" || home.ParentValued() {
		t.Fatalf("home selector = %q parent=%v, want direct Pi home", home.EnvironmentVariable(), home.ParentValued())
	}
	if got := relativePaths(home.Roots()); len(got) != 1 || got[0] != ".pi" {
		t.Fatalf("home roots = %v, want [.pi]", got)
	}

	instructions, err := adapter.Instructions(ctx, harnessapi.Monolithic)
	if err != nil {
		t.Fatal(err)
	}
	if instructions.Form() != harnessapi.Monolithic || len(instructions.Notice()) != 0 {
		t.Fatalf("instruction form/notice = %q/%q, want monolithic with no static notice", instructions.Form(), instructions.Notice())
	}
	if got := relativePaths(instructions.Paths()); len(got) != 1 || got[0] != "AGENTS.md" {
		t.Fatalf("instruction paths = %v, want [AGENTS.md]", got)
	}
	if _, err := adapter.Instructions(ctx, harnessapi.Referenced); refusalCode(err) != "environment_form_unsupported" {
		t.Fatalf("referenced instruction refusal = %v, want environment_form_unsupported", err)
	}

	skills, err := adapter.Skills(ctx, harnessapi.UserScope)
	if err != nil {
		t.Fatal(err)
	}
	if skills.Scope() != harnessapi.UserScope || !skills.NativeDiscovery() || skills.Mirror() {
		t.Fatalf("skill scope/native/mirror = %q/%v/%v", skills.Scope(), skills.NativeDiscovery(), skills.Mirror())
	}
	if got := relativePaths(skills.Paths()); len(got) != 1 || got[0] != "skills" {
		t.Fatalf("skill paths = %v, want [skills]", got)
	}
	if len(skills.IgnorePaths().Paths()) != 0 {
		t.Fatalf("skill ignore paths = %v, want empty", relativePaths(skills.IgnorePaths()))
	}
	if _, err := adapter.Skills(ctx, harnessapi.ProjectScope); refusalCode(err) != harnessapi.CapabilityUnsupported {
		t.Fatalf("project skills refusal = %v, want %s", err, harnessapi.CapabilityUnsupported)
	}

	credentials, err := adapter.Credentials(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := relativePaths(credentials.Paths()); len(got) != 1 || got[0] != "agent/auth.json" || credentials.Strategy() != "file-link" || !credentials.Qualified() {
		t.Fatalf("credential route = %v/%q/%v", got, credentials.Strategy(), credentials.Qualified())
	}

	trust, err := adapter.Trust(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if trust.Required() || !trust.ObservationOnly() {
		t.Fatalf("trust policy = required:%v observation-only:%v, want no required trust and observation-only", trust.Required(), trust.ObservationOnly())
	}

	for _, scope := range []harnessapi.Scope{harnessapi.UserScope, harnessapi.ProjectScope} {
		_, err = adapter.MCP(ctx, scope)
		if refusalCode(err) != harnessapi.CapabilityUnsupported {
			t.Fatalf("MCP(%q) error = %v, want %s refusal", scope, err, harnessapi.CapabilityUnsupported)
		}
	}

	capabilities := adapter.Capabilities()
	for _, capability := range []harnessapi.Capability{harnessapi.ManagedHome, harnessapi.InstructionOutput, harnessapi.SkillProjection, harnessapi.NativeSkillDiscovery, harnessapi.CredentialMetadata, harnessapi.TrustObservation, harnessapi.ReleaseProbing} {
		if !capabilities.Has(capability) {
			t.Fatalf("supported capability %d was removed", capability)
		}
	}
	if capabilities.Has(harnessapi.MCPInspection) || capabilities.Has(harnessapi.NativeMCPWrite) {
		t.Fatal("unsupported MCP inspection or writing was declared as a capability")
	}
}

func relativePaths(paths harnessapi.NamedPaths) []string {
	values := paths.Paths()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

func refusalCode(err error) string {
	refusal, ok := err.(harnessapi.Refusal)
	if !ok {
		return ""
	}
	return refusal.Code()
}
