package opencode

import (
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

func testContext(t *testing.T, scope harnessapi.Scope) harnessapi.Context {
	t.Helper()
	context, err := harnessapi.NewContext("darwin", scope, harnessapi.VerifiedRelease{})
	if err != nil {
		t.Fatal(err)
	}
	return context
}

func pathStrings(paths harnessapi.NamedPaths) []string {
	values := paths.Paths()
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = value.String()
	}
	return out
}

func TestAdapterIdentityCapabilitiesAndReleaseProbe(t *testing.T) {
	adapter := Adapter{}
	if adapter.ID() != ID || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter metadata = %q/%d", adapter.ID(), adapter.APIVersion())
	}
	want := []harnessapi.Capability{
		harnessapi.ManagedHome,
		harnessapi.InstructionOutput,
		harnessapi.NativeSkillDiscovery,
		harnessapi.MCPInspection,
		harnessapi.TrustObservation,
		harnessapi.ReleaseProbing,
	}
	if got := adapter.Capabilities().Values(); !reflect.DeepEqual(got, want) {
		t.Fatalf("capabilities = %v, want %v", got, want)
	}
	probe, err := adapter.ReleaseProbe(testContext(t, harnessapi.UserScope))
	if err != nil {
		t.Fatal(err)
	}
	if !probe.ValidValue() || probe.ExecutableAlias() != "opencode" || !reflect.DeepEqual(probe.Arguments(), []string{"--version"}) || probe.Timeout() != harnessapi.ProbeTimeoutV1 || len(probe.Environment()) != 0 {
		t.Fatalf("release probe = %+v", probe)
	}
}

func TestHomeAndInstructionForms(t *testing.T) {
	adapter := Adapter{}
	home, err := adapter.Home(testContext(t, harnessapi.UserScope))
	if err != nil {
		t.Fatal(err)
	}
	if home.EnvironmentVariable() != "XDG_CONFIG_HOME" || !home.ParentValued() || !reflect.DeepEqual(pathStrings(home.Roots()), []string{"opencode"}) {
		t.Fatalf("home layout = %q, parent=%v, roots=%v", home.EnvironmentVariable(), home.ParentValued(), pathStrings(home.Roots()))
	}
	for _, test := range []struct {
		form harnessapi.Form
		want []string
	}{
		{form: "", want: []string{"AGENTS.md"}},
		{form: harnessapi.Monolithic, want: []string{"AGENTS.md"}},
		{form: harnessapi.Referenced, want: []string{"AGENTS.md", "opencode.json"}},
	} {
		spec, err := adapter.Instructions(testContext(t, harnessapi.UserScope), test.form)
		if err != nil {
			t.Fatalf("form %q: %v", test.form, err)
		}
		wantForm := test.form
		if wantForm == "" {
			wantForm = harnessapi.Monolithic
		}
		if spec.Form() != wantForm || !reflect.DeepEqual(pathStrings(spec.Paths()), test.want) || len(spec.Notice()) != 0 {
			t.Errorf("form %q spec = %q %v %q", test.form, spec.Form(), pathStrings(spec.Paths()), spec.Notice())
		}
	}
	if _, err := adapter.Instructions(testContext(t, harnessapi.UserScope), "unsupported"); err == nil || err.Error() != "environment_form_unsupported" {
		t.Fatalf("unsupported form refusal = %v", err)
	}
}

func TestNativeSkillDiscoveryScopes(t *testing.T) {
	adapter := Adapter{}
	for _, test := range []struct {
		scope       harnessapi.Scope
		mirror      bool
		ignorePaths []string
	}{
		{scope: harnessapi.UserScope, mirror: true, ignorePaths: []string{}},
		{scope: harnessapi.ProjectScope, mirror: false, ignorePaths: []string{".agents"}},
	} {
		discovery, err := adapter.Skills(testContext(t, test.scope), test.scope)
		if err != nil {
			t.Fatalf("scope %q: %v", test.scope, err)
		}
		if discovery.Scope() != test.scope || !discovery.NativeDiscovery() || discovery.Mirror() != test.mirror || !reflect.DeepEqual(pathStrings(discovery.Paths()), []string{".agents/skills"}) || !reflect.DeepEqual(pathStrings(discovery.IgnorePaths()), test.ignorePaths) {
			t.Errorf("scope %q discovery = %+v, paths=%v ignores=%v", test.scope, discovery, pathStrings(discovery.Paths()), pathStrings(discovery.IgnorePaths()))
		}
	}
	if _, err := adapter.Skills(testContext(t, harnessapi.UserScope), "invalid"); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("invalid skill scope refusal = %v", err)
	}
}

func TestMCPInspectionAndTrust(t *testing.T) {
	adapter := Adapter{}
	for _, test := range []struct {
		scope   harnessapi.Scope
		paths   []string
		trusted bool
	}{
		{scope: harnessapi.UserScope, paths: []string{".config/opencode/opencode.json", ".config/opencode/opencode.jsonc"}},
		{scope: harnessapi.ProjectScope, paths: []string{"opencode.json", "opencode.jsonc"}, trusted: true},
	} {
		channel, err := adapter.MCP(testContext(t, test.scope), test.scope)
		if err != nil {
			t.Fatalf("scope %q: %v", test.scope, err)
		}
		if channel.Format() != harnessapi.JSONC || !channel.Inspect() || channel.Write() || channel.RequiresTrust() != test.trusted || !reflect.DeepEqual(pathStrings(channel.Paths()), test.paths) {
			t.Errorf("scope %q MCP = format %q inspect=%v write=%v trust=%v paths=%v", test.scope, channel.Format(), channel.Inspect(), channel.Write(), channel.RequiresTrust(), pathStrings(channel.Paths()))
		}
		policy, err := adapter.Trust(testContext(t, test.scope))
		if err != nil || policy.Required() != test.trusted || !policy.ObservationOnly() {
			t.Errorf("scope %q trust = required %v observation-only %v (%v)", test.scope, policy.Required(), policy.ObservationOnly(), err)
		}
	}
	if _, err := adapter.MCP(testContext(t, harnessapi.UserScope), "invalid"); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("invalid MCP scope refusal = %v", err)
	}
	if _, err := adapter.Trust(harnessapi.Context{}); err == nil || err.Error() != "harness_scope_unsupported" {
		t.Fatalf("invalid trust context refusal = %v", err)
	}
}

func TestManagedHomeValueMethods(t *testing.T) {
	adapter := Adapter{}
	context := testContext(t, harnessapi.UserScope)
	property, err := harnessapi.NewProperty("present", "true")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := harnessapi.NewSnapshot([]harnessapi.Property{property})
	if err != nil {
		t.Fatal(err)
	}
	request, err := harnessapi.NewRequest(harnessapi.ManagedHome, harnessapi.Monolithic, harnessapi.UserScope)
	if err != nil {
		t.Fatal(err)
	}
	if refusals, err := adapter.Preflight(context, snapshot, request); err != nil || len(refusals) != 0 {
		t.Fatalf("preflight = %v (%v)", refusals, err)
	}
	seed, err := adapter.PlanSeed(context, snapshot, request)
	if err != nil || len(seed.Actions()) != 0 || len(seed.Refusals()) != 0 {
		t.Fatalf("seed plan = %v / %v (%v)", seed.Actions(), seed.Refusals(), err)
	}
	status, err := adapter.Status(context, snapshot)
	if err != nil || !reflect.DeepEqual(status.Properties(), []harnessapi.Property{property}) {
		t.Fatalf("status = %v (%v)", status.Properties(), err)
	}
	drift, err := adapter.Drift(context, snapshot)
	if err != nil || len(drift.Paths().Paths()) != 0 || len(drift.Refusals()) != 0 {
		t.Fatalf("drift = %v / %v (%v)", pathStrings(drift.Paths()), drift.Refusals(), err)
	}
}

func TestNamedPathsRejectInvalidPath(t *testing.T) {
	if _, err := namedPaths("../outside"); err == nil {
		t.Fatal("invalid relative path was accepted")
	}
	if _, err := harnessvalue.NewRelativePath("../outside"); err == nil {
		t.Fatal("contract negative control accepted an escaping path")
	}
}
