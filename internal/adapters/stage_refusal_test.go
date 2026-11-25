package adapters

import (
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessvalue"
)

type refusalAdapterFixture struct {
	harnessapi.UnportedOperations
	capabilities harnessapi.CapabilitySet
	discovery    harnessapi.SkillDiscovery
	skillsCalls  int
	homeCalls    int
	skillsErr    error
	homeErr      error
}

func (a *refusalAdapterFixture) ID() harnessapi.ID { return "synthetic" }

func (a *refusalAdapterFixture) ReleaseProbe(harnessapi.Context) (harnessapi.ProbeSpec, error) {
	return harnessapi.ProbeSpec{}, harnessapi.NewRefusal(harnessapi.OperationNotPorted)
}

func (a *refusalAdapterFixture) APIVersion() harnessapi.APIVersion {
	return harnessapi.AdapterAPIVersionV1
}

func (a *refusalAdapterFixture) Capabilities() harnessapi.CapabilitySet { return a.capabilities }

func (a *refusalAdapterFixture) Skills(harnessapi.Context, harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	a.skillsCalls++
	return a.discovery, a.skillsErr
}

func (a *refusalAdapterFixture) Home(harnessapi.Context) (harnessapi.HomeLayout, error) {
	a.homeCalls++
	return harnessapi.HomeLayout{}, a.homeErr
}

func newRefusalAdapterFixture(t *testing.T, capabilities ...harnessapi.Capability) *refusalAdapterFixture {
	t.Helper()
	set, err := harnessapi.NewCapabilitySet(capabilities...)
	if err != nil {
		t.Fatal(err)
	}
	path, err := harnessvalue.NewRelativePath("skills")
	if err != nil {
		t.Fatal(err)
	}
	paths, err := harnessapi.NewNamedPaths([]harnessvalue.RelativePath{path})
	if err != nil {
		t.Fatal(err)
	}
	discovery, err := harnessapi.NewSkillDiscovery(paths, harnessapi.NamedPaths{}, harnessapi.UserScope, false, true)
	if err != nil {
		t.Fatal(err)
	}
	return &refusalAdapterFixture{capabilities: set, discovery: discovery}
}

func runStageGlobalWithFixture(t *testing.T, adapter *refusalAdapterFixture) (Mirror, string, string) {
	t.Helper()
	home, userHome, stageRoot := t.TempDir(), t.TempDir(), t.TempDir()
	makeSkill(t, filepath.Join(home, "global", "skills"), "fixture-skill")
	lookup := func(harnessapi.ID) (harnessapi.Adapter, error) { return adapter, nil }
	mirror, err := stageGlobalWithLookup(
		stageRoot, home, userHome, []string{"synthetic"}, []string{"fixture-skill"}, "copy", nil, nil, nil, lookup,
	)
	if err != nil {
		return Mirror{}, err.Error(), ""
	}
	return mirror, "", filepath.Join(userHome, "skills", "fixture-skill")
}

func TestStageGlobalPropagatesDeclaredSkillsRefusal(t *testing.T) {
	adapter := newRefusalAdapterFixture(t, harnessapi.SkillProjection)
	adapter.skillsErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	_, errText, _ := runStageGlobalWithFixture(t, adapter)
	if errText != harnessapi.OperationNotPorted {
		t.Fatalf("declared Skills refusal was hidden by legacy fallback: err=%q", errText)
	}
}

func TestStageGlobalPropagatesDeclaredHomeRefusal(t *testing.T) {
	adapter := newRefusalAdapterFixture(t, harnessapi.SkillProjection, harnessapi.ManagedHome)
	adapter.homeErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	_, errText, _ := runStageGlobalWithFixture(t, adapter)
	if errText != harnessapi.OperationNotPorted {
		t.Fatalf("declared Home refusal was hidden by the default home: err=%q", errText)
	}
}

func TestStageGlobalKeepsDefaultRootWhenHomeCapabilityIsAbsent(t *testing.T) {
	adapter := newRefusalAdapterFixture(t, harnessapi.SkillProjection)
	adapter.homeErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	mirror, errText, expected := runStageGlobalWithFixture(t, adapter)
	if errText != "" {
		t.Fatalf("optional Home changed the staging result: %s", errText)
	}
	if _, found := find(mirror.Plan(), expected); !found {
		t.Fatalf("skill was not staged beneath the legacy default root: %+v", mirror.Plan().Targets)
	}
	if adapter.homeCalls != 0 {
		t.Fatalf("staging called undeclared Home %d times", adapter.homeCalls)
	}
}

func TestRegisteredSkillsUsesLegacyRouteWhenCapabilitiesAreAbsent(t *testing.T) {
	adapter := newRefusalAdapterFixture(t)
	adapter.skillsErr = harnessapi.NewRefusal(harnessapi.OperationNotPorted)
	lookup := func(harnessapi.ID) (harnessapi.Adapter, error) { return adapter, nil }
	_, ported, err := registeredSkillsWithLookup("synthetic", harnessapi.ProjectScope, lookup)
	if err != nil || ported {
		t.Fatalf("undeclared skills capability did not select the legacy route: ported=%t err=%v", ported, err)
	}
	if adapter.skillsCalls != 0 {
		t.Fatalf("caller invoked undeclared Skills %d times", adapter.skillsCalls)
	}
}
