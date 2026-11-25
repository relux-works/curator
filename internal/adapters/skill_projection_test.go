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

func TestProjectSkillResolverLeavesNativeDiscoveryWithoutMirror(t *testing.T) {
	pi, err := harnessregistry.Lookup("pi")
	if err != nil {
		t.Fatal(err)
	}
	paths := projectAgentPaths()
	if _, ok := paths["pi"]; ok {
		t.Fatalf("native discovery project path = %q, want no project mirror", paths["pi"])
	}
	if _, ok, err := registeredSkillProjection("pi", harnessapi.Context{}, harnessapi.ProjectScope, pi); ok || err != nil {
		t.Fatalf("native discovery project scope = supported %t, error %v; want no mirror", ok, err)
	}
}

func TestProjectSkillResolverStillFailsForOtherHarnessRefusals(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{name: "ordinary error", err: errors.New("unexpected skill failure")},
		{name: "other typed refusal", err: fmt.Errorf("wrapped: %w", harnessapi.NewRefusal(harnessapi.OperationNotPorted))},
		{name: "non-Pi project capability refusal", err: fmt.Errorf("wrapped: %w", harnessapi.NewRefusal(harnessapi.CapabilityUnsupported))},
	}
	capabilities, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := skillProjectionTestAdapter{capabilities: capabilities, err: tc.err}
			if _, supported, err := registeredSkillProjection("fixture", harnessapi.Context{}, harnessapi.ProjectScope, adapter); supported || err == nil {
				t.Fatalf("project refusal = supported %t, error %v; want propagated error", supported, err)
			}
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

	got, ok, err := registeredSkillProjection("fixture", harnessapi.Context{}, harnessapi.ProjectScope, adapter)
	if err != nil || !ok || !got.Mirror() || got.Scope() != harnessapi.ProjectScope {
		t.Fatalf("project skill declaration = %#v, supported=%v, error=%v", got, ok, err)
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
	discovery, ok, err := registeredSkillProjection("pi", harnessapi.Context{}, harnessapi.UserScope, pi)
	if err != nil || !ok || discovery.Scope() != harnessapi.UserScope || !discovery.NativeDiscovery() || discovery.Mirror() {
		t.Fatalf("Pi user-scope declaration = %#v, supported=%v, error=%v; want native discovery without a mirror", discovery, ok, err)
	}
	paths := discovery.Paths().Paths()
	if len(paths) != 1 || paths[0].String() != "skills" {
		t.Fatalf("Pi user-scope paths = %v, want [skills]", paths)
	}
}

func TestSkillResolverRejectsUndeclaredFlagCombinations(t *testing.T) {
	cases := []struct {
		name   string
		scope  harnessapi.Scope
		native bool
		mirror bool
	}{
		{name: "user neither", scope: harnessapi.UserScope, native: false, mirror: false},
		{name: "user both", scope: harnessapi.UserScope, native: true, mirror: true},
		{name: "project native no mirror", scope: harnessapi.ProjectScope, native: true, mirror: false},
		{name: "project no flags", scope: harnessapi.ProjectScope, native: false, mirror: false},
	}
	capabilities, err := harnessapi.NewCapabilitySet(harnessapi.SkillProjection)
	if err != nil {
		t.Fatal(err)
	}
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
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			discovery, err := harnessapi.NewSkillDiscovery(paths, ignores, tc.scope, tc.native, tc.mirror)
			if err != nil {
				t.Fatal(err)
			}
			adapter := skillProjectionTestAdapter{capabilities: capabilities, discovery: discovery}
			if _, _, err := registeredSkillProjection("fixture", harnessapi.Context{}, tc.scope, adapter); err == nil {
				t.Fatal("resolver accepted a flag combination outside the scope allowlist")
			}
		})
	}
}
