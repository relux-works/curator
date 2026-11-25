package mcp

import (
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

type refusalMCPAdapterFixture struct {
	harnessapi.UnportedOperations
	capabilities harnessapi.CapabilitySet
	channel      harnessapi.MCPChannel
	mcpCalls     int
	homeCalls    int
	mcpErr       error
	homeErr      error
}

func (a *refusalMCPAdapterFixture) ID() harnessapi.ID { return "synthetic" }

func (a *refusalMCPAdapterFixture) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.OperationNotPorted)
}

func (a *refusalMCPAdapterFixture) APIVersion() harnessapi.APIVersion {
	return harnessapi.AdapterAPIVersionV1
}

func (a *refusalMCPAdapterFixture) Capabilities() harnessapi.CapabilitySet { return a.capabilities }

func (a *refusalMCPAdapterFixture) MCP(harnessapi.Context, harnessapi.Scope) (harnessapi.MCPChannel, error) {
	a.mcpCalls++
	return a.channel, a.mcpErr
}

func (a *refusalMCPAdapterFixture) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	a.homeCalls++
	return harnessapi.HomeLayout{}, a.homeErr
}

func newRefusalMCPAdapterFixture(t *testing.T, capabilities ...harnessapi.Capability) *refusalMCPAdapterFixture {
	t.Helper()
	set, err := harnessapi.NewCapabilitySet(capabilities...)
	if err != nil {
		t.Fatal(err)
	}
	path, err := harnessvalue.NewRelativePath("config.toml")
	if err != nil {
		t.Fatal(err)
	}
	paths, err := harnessapi.NewNamedPaths([]harnessvalue.RelativePath{path})
	if err != nil {
		t.Fatal(err)
	}
	channel, err := harnessapi.NewMCPChannel(paths, harnessapi.TOML, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
	return &refusalMCPAdapterFixture{capabilities: set, channel: channel}
}

func withMCPFixture(adapter *refusalMCPAdapterFixture) func(harnessapi.ID) (harnessapi.Adapter, error) {
	return func(harnessapi.ID) (harnessapi.Adapter, error) { return adapter, nil }
}

func TestMCPPropagatesDeclaredInspectionRefusal(t *testing.T) {
	adapter := newRefusalMCPAdapterFixture(t, harnessapi.MCPInspection)
	adapter.mcpErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	sources, err := mcpSourcesWithLookup("synthetic", harnessapi.UserScope, []string{"legacy.json"}, withMCPFixture(adapter))
	if err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("declared MCP refusal became legacy sources: sources=%v err=%v", sources, err)
	}
}

func TestMCPPropagatesDeclaredHomeRefusal(t *testing.T) {
	adapter := newRefusalMCPAdapterFixture(t, harnessapi.MCPInspection, harnessapi.ManagedHome)
	adapter.homeErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	sources, err := mcpSourcesWithLookup("synthetic", harnessapi.UserScope, []string{"legacy.json"}, withMCPFixture(adapter))
	if err == nil || err.Error() != harnessapi.OperationNotPorted {
		t.Fatalf("declared Home refusal returned partial MCP sources: sources=%v err=%v", sources, err)
	}
}

func TestMCPUsesLegacySourcesOnlyWhenInspectionIsUndeclared(t *testing.T) {
	adapter := newRefusalMCPAdapterFixture(t)
	adapter.mcpErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	want := []string{"legacy.json"}
	sources, err := mcpSourcesWithLookup("synthetic", harnessapi.UserScope, want, withMCPFixture(adapter))
	if err != nil || len(sources) != 1 || sources[0] != want[0] {
		t.Fatalf("undeclared inspection did not preserve legacy source: sources=%v err=%v", sources, err)
	}
	if adapter.mcpCalls != 0 {
		t.Fatalf("caller invoked undeclared MCP %d times", adapter.mcpCalls)
	}
}

func TestMCPKeepsRelativePathWhenManagedHomeIsUndeclared(t *testing.T) {
	adapter := newRefusalMCPAdapterFixture(t, harnessapi.MCPInspection)
	adapter.homeErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	sources, err := mcpSourcesWithLookup("synthetic", harnessapi.UserScope, []string{"legacy.json"}, withMCPFixture(adapter))
	if err != nil || len(sources) != 1 || sources[0] != "config.toml" {
		t.Fatalf("missing optional home changed declared MCP source: sources=%v err=%v", sources, err)
	}
	if adapter.homeCalls != 0 {
		t.Fatalf("caller invoked undeclared Home %d times", adapter.homeCalls)
	}
}
