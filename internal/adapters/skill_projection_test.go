package adapters

import (
	"errors"
	"fmt"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
	"github.com/relux-works/curator/internal/harnessvalue"
)

type skillProjectionTestAdapter struct {
	harnessapi.Adapter
	capabilities harnessapi.CapabilitySet
	discovery    harnessapi.SkillDiscovery
	err          error
}

func (a skillProjectionTestAdapter) Capabilities() harnessapi.CapabilitySet { return a.capabilities }

func (a skillProjectionTestAdapter) Skills(_ harnessapi.Context, _ harnessapi.Scope) (harnessapi.SkillDiscovery, error) {
	return a.discovery, a.err
}

func TestProjectSkillResolverTreatsPiScopeRefusalAsNoMirror(t *testing.T) {
	pi, err := harnessregistry.Lookup("pi")
	if err != nil {
		t.Fatal(err)
	}
	paths := projectAgentPaths()
	if _, ok := paths["pi"]; ok {
		t.Fatalf("Pi project path = %q, want no project mirror", paths["pi"])
	}
	if _, ok := registeredSkillProjection("pi", harnessapi.ProjectScope, pi); ok {
		t.Fatal("Pi project scope unexpectedly produced a mirror")
	}
}

func TestProjectSkillResolverStillFailsForOtherErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "ordinary error", err: errors.New("unexpected skill failure")},
		{name: "other typed refusal", err: fmt.Errorf("wrapped: %w", harnessapi.NewRefusal(harnessapi.OperationNotPorted))},
	}
	capabilities, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := skillProjectionTestAdapter{capabilities: capabilities, err: tc.err}
			defer func() {
				if recover() == nil {
					t.Fatal("unexpected Skills error did not fail the resolver")
				}
			}()
			_, _ = registeredSkillProjection("fixture", harnessapi.ProjectScope, adapter)
		})
	}
}

func TestProjectSkillResolverKeepsDeclaredMirrorPath(t *testing.T) {
	path, err := harnessvalue.NewRelativePath("fixture/skills")
	if err != nil {
		t.Fatal(err)
	}
	paths, err := harnessapi.NewNamedPaths([]harnessvalue.RelativePath{path})
	if err != nil {
		t.Fatal(err)
	}
	ignores, err := harnessapi.NewNamedPaths(nil)
	if err != nil {
		t.Fatal(err)
	}
	discovery, err := harnessapi.NewSkillDiscovery(paths, ignores, harnessapi.ProjectScope, false, true)
	if err != nil {
		t.Fatal(err)
	}
	capabilities, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection)
	if err != nil {
		t.Fatal(err)
	}
	adapter := skillProjectionTestAdapter{capabilities: capabilities, discovery: discovery}

	got, ok := registeredSkillProjection("fixture", harnessapi.ProjectScope, adapter)
	if !ok || !got.Mirror() || got.Scope() != harnessapi.ProjectScope {
		t.Fatalf("project skill declaration = %#v, supported=%v", got, ok)
	}
	if path := oneSkillPath("fixture", got.Paths()); path != "fixture/skills" {
		t.Fatalf("project skill path = %q, want fixture/skills", path)
	}
}

func TestUserSkillResolverAcceptsPiNativeDiscovery(t *testing.T) {
	pi, err := harnessregistry.Lookup("pi")
	if err != nil {
		t.Fatal(err)
	}
	discovery, ok := registeredSkillProjection("pi", harnessapi.UserScope, pi)
	if !ok || discovery.Scope() != harnessapi.UserScope || !discovery.NativeDiscovery() || discovery.Mirror() {
		t.Fatalf("Pi user-scope declaration = %#v, supported=%v; want native discovery without a mirror", discovery, ok)
	}
	paths := discovery.Paths().Paths()
	if len(paths) != 1 || paths[0].String() != "skills" {
		t.Fatalf("Pi user-scope paths = %v, want [skills]", paths)
	}
}
