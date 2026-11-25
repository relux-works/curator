package gemini

import (
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

func contextFor(t *testing.T, scope harnessapi.Scope) harnessapi.Context {
	t.Helper()
	context, err := harnessapi.NewContext("darwin", scope, harnessapi.VerifiedRelease{})
	if err != nil {
		t.Fatal(err)
	}
	return context
}

func pathsOf(paths harnessapi.NamedPaths) []string {
	values := paths.Paths()
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.String()
	}
	return result
}

func refusalCode(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}

func TestAdapterMetadataAndUnsupportedProbe(t *testing.T) {
	adapter := New()
	if adapter.ID() != ID || adapter.APIVersion() != harnessapi.AdapterAPIVersionV1 {
		t.Fatalf("adapter metadata = (%q, %d)", adapter.ID(), adapter.APIVersion())
	}
	want, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection, harnessapi.MCPInspection)
	if err != nil {
		t.Fatal(err)
	}
	if adapter.Capabilities() != want || !Capabilities().Has(harnessapi.SkillProjection) ||
		!Capabilities().Has(harnessapi.MCPInspection) || Capabilities().Has(harnessapi.NativeMCPWrite) {
		t.Fatalf("capabilities = %v", adapter.Capabilities().Values())
	}
	if _, err := adapter.ReleaseProbe(contextFor(t, harnessapi.UserScope)); refusalCode(err) != harnessapi.CapabilityUnsupported {
		t.Fatalf("release probe refusal = %v", err)
	}
}

func TestSkillDiscoveryByScope(t *testing.T) {
	for _, scope := range []harnessapi.Scope{harnessapi.ProjectScope, harnessapi.UserScope} {
		t.Run(string(scope), func(t *testing.T) {
			discovery, err := (Adapter{}).Skills(contextFor(t, scope), scope)
			if err != nil {
				t.Fatal(err)
			}
			if got := pathsOf(discovery.Paths()); !reflect.DeepEqual(got, []string{".gemini/skills"}) {
				t.Fatalf("skill paths = %v", got)
			}
			wantIgnores := []string{}
			if scope == harnessapi.ProjectScope {
				wantIgnores = []string{".gemini/skills"}
			}
			if got := pathsOf(discovery.IgnorePaths()); !reflect.DeepEqual(got, wantIgnores) {
				t.Fatalf("ignore paths = %v, want %v", got, wantIgnores)
			}
			if discovery.Scope() != scope || discovery.NativeDiscovery() || !discovery.Mirror() {
				t.Fatalf("skill policy = scope:%s native:%t mirror:%t", discovery.Scope(), discovery.NativeDiscovery(), discovery.Mirror())
			}
		})
	}
	if _, err := (Adapter{}).Skills(contextFor(t, harnessapi.ProjectScope), harnessapi.UserScope); refusalCode(err) != "harness_scope_unsupported" {
		t.Fatalf("mismatched skill scope refusal = %v", err)
	}
	if _, err := (Adapter{}).Skills(contextFor(t, harnessapi.ProjectScope), "invalid"); refusalCode(err) != "harness_scope_unsupported" {
		t.Fatalf("invalid skill scope refusal = %v", err)
	}
}

func TestMCPInspectionByScope(t *testing.T) {
	for _, scope := range []harnessapi.Scope{harnessapi.ProjectScope, harnessapi.UserScope} {
		t.Run(string(scope), func(t *testing.T) {
			channel, err := (Adapter{}).MCP(contextFor(t, scope), scope)
			if err != nil {
				t.Fatal(err)
			}
			if got := pathsOf(channel.Paths()); !reflect.DeepEqual(got, []string{".gemini/settings.json"}) {
				t.Fatalf("MCP paths = %v", got)
			}
			if channel.Format() != harnessapi.JSON || !channel.Inspect() || channel.Write() ||
				channel.RequiresTrust() != (scope == harnessapi.ProjectScope) {
				t.Fatalf("MCP policy = format:%s inspect:%t write:%t trust:%t", channel.Format(), channel.Inspect(), channel.Write(), channel.RequiresTrust())
			}
		})
	}
	if _, err := (Adapter{}).MCP(contextFor(t, harnessapi.ProjectScope), harnessapi.UserScope); refusalCode(err) != "harness_scope_unsupported" {
		t.Fatalf("mismatched MCP scope refusal = %v", err)
	}
	if _, err := (Adapter{}).MCP(contextFor(t, harnessapi.ProjectScope), "invalid"); refusalCode(err) != "harness_scope_unsupported" {
		t.Fatalf("invalid MCP scope refusal = %v", err)
	}
}

func TestInvalidNativePathIsRefused(t *testing.T) {
	if _, err := namedPath("../escape"); err == nil {
		t.Fatal("escaping native path accepted")
	}
	var invalid harnessvalue.RelativePath
	if _, err := harnessapi.NewNamedPaths([]harnessvalue.RelativePath{invalid}); refusalCode(err) != "harness_path_invalid" {
		t.Fatalf("invalid named path refusal = %v", err)
	}
}
