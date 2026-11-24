package harnessboundary

import (
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/harnessapi"
	"github.com/relux-works/curator/internal/harnessregistry"
)

func code(err error) string {
	var refusal harnessapi.Refusal
	if errors.As(err, &refusal) {
		return refusal.Code()
	}
	return ""
}

// Unsupported registration versions refuse before construction; a reported
// version mismatch refuses after metadata only. Correct metadata passes, while
// ID/capability mismatches never return an adapter for an operational call.

// effectSpy counts a successful operation result or a nonzero result paired
// with refusal. Staged refusals must produce neither; the source effect gate
// separately rejects direct filesystem, process, and network effects.
type effectSpy struct{ count int }

func (s *effectSpy) observe(value any, err error) error {
	if err == nil || !reflect.ValueOf(value).IsZero() {
		s.count++
	}
	return err
}

// Every unported method is observed through the same effect spy.
func notPortedResults(adapter harnessapi.Adapter, effects *effectSpy, capabilities harnessapi.CapabilitySet) []error {
	c, s, r := harnessapi.Context{}, harnessapi.Snapshot{}, harnessapi.Request{}
	var errors []error
	v, err := adapter.ParseReleaseProbe(c, harnessapi.ProbeOutput{})
	errors = append(errors, effects.observe(v, err))
	if !capabilities.Has(harnessapi.ManagedHome) {
		vh, err := adapter.Home(c)
		errors = append(errors, effects.observe(vh, err))
	}
	if !capabilities.Has(harnessapi.InstructionOutput) {
		vi, err := adapter.Instructions(c, harnessapi.Monolithic)
		errors = append(errors, effects.observe(vi, err))
	}
	if !capabilities.Has(harnessapi.SkillProjection) && !capabilities.Has(harnessapi.NativeSkillDiscovery) {
		vs, err := adapter.Skills(c, harnessapi.ProjectScope)
		errors = append(errors, effects.observe(vs, err))
	}
	vset, err := adapter.Settings(c)
	errors = append(errors, effects.observe(vset, err))
	if !capabilities.Has(harnessapi.MCPInspection) && !capabilities.Has(harnessapi.NativeMCPWrite) {
		vmcp, err := adapter.MCP(c, harnessapi.UserScope)
		errors = append(errors, effects.observe(vmcp, err))
	}
	if !capabilities.Has(harnessapi.CredentialMetadata) {
		vc, err := adapter.Credentials(c)
		errors = append(errors, effects.observe(vc, err))
	}
	if !capabilities.Has(harnessapi.TrustObservation) {
		vt, err := adapter.Trust(c)
		errors = append(errors, effects.observe(vt, err))
	}
	if !capabilities.Has(harnessapi.ManagedHome) && !capabilities.Has(harnessapi.NativeMCPWrite) {
		vr, err := adapter.Preflight(c, s, r)
		errors = append(errors, effects.observe(vr, err))
	}
	if !capabilities.Has(harnessapi.ManagedHome) {
		vp, err := adapter.PlanSeed(c, s, r)
		errors = append(errors, effects.observe(vp, err))
		vstatus, err := adapter.Status(c, s)
		errors = append(errors, effects.observe(vstatus, err))
		vd, err := adapter.Drift(c, s)
		errors = append(errors, effects.observe(vd, err))
	}
	if !capabilities.Has(harnessapi.ProjectContextOutput) {
		vpc, err := adapter.ProjectContext(c, s, r)
		errors = append(errors, effects.observe(vpc, err))
	}
	return errors
}

// Staged capabilities describe implementations, rather than future native
// support. A false operational declaration fails this same conformance check.
func stagedCapabilities(adapter harnessapi.Adapter) error {
	capabilities := adapter.Capabilities()
	for _, capability := range capabilities.Values() {
		switch capability {
		case harnessapi.ManagedHome:
			layout, err := adapter.Home(harnessapi.Context{})
			if err != nil || layout.EnvironmentVariable() == "" || len(layout.Roots().Paths()) == 0 {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
		case harnessapi.InstructionOutput:
			instructions, err := adapter.Instructions(harnessapi.Context{}, harnessapi.Monolithic)
			if err != nil || instructions.Form() != harnessapi.Monolithic || len(instructions.Paths().Paths()) == 0 {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
		case harnessapi.ReleaseProbing:
		case harnessapi.TrustObservation:
			userContext, err := harnessapi.NewContext("linux", harnessapi.UserScope, harnessapi.VerifiedRelease{})
			if err != nil {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			userPolicy, err := adapter.Trust(userContext)
			if err != nil || userPolicy.Required() || !userPolicy.ObservationOnly() {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			projectContext, err := harnessapi.NewContext("linux", harnessapi.ProjectScope, harnessapi.VerifiedRelease{})
			if err != nil {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			projectPolicy, err := adapter.Trust(projectContext)
			if err != nil || !projectPolicy.Required() || !projectPolicy.ObservationOnly() {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
		case harnessapi.SkillProjection, harnessapi.NativeSkillDiscovery:
			discovery, err := adapter.Skills(harnessapi.Context{}, harnessapi.ProjectScope)
			if err != nil || discovery.Scope() != harnessapi.ProjectScope {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			if capability == harnessapi.SkillProjection && (!discovery.Mirror() || discovery.NativeDiscovery()) {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			if capability == harnessapi.NativeSkillDiscovery && !discovery.NativeDiscovery() {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
		case harnessapi.MCPInspection, harnessapi.NativeMCPWrite:
			channel, err := adapter.MCP(harnessapi.Context{}, harnessapi.UserScope)
			if err != nil || !channel.Inspect() {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			if capability == harnessapi.MCPInspection && channel.Write() {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
			if capability == harnessapi.NativeMCPWrite && !channel.Write() {
				return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
			}
		default:
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
	}
	effects := effectSpy{}
	for _, err := range notPortedResults(adapter, &effects, capabilities) {
		if code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
	}
	if effects.count != 0 {
		return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	expectedEffects := 0
	if capabilities.Has(harnessapi.ManagedHome) {
		value, err := adapter.Home(harnessapi.Context{})
		if !declaredResult(value, err, true, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		expectedEffects++
	}
	if capabilities.Has(harnessapi.InstructionOutput) {
		value, err := adapter.Instructions(harnessapi.Context{}, harnessapi.Monolithic)
		if !declaredResult(value, err, true, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		expectedEffects++
	}
	if capabilities.Has(harnessapi.TrustObservation) {
		context, err := harnessapi.NewContext("linux", harnessapi.UserScope, harnessapi.VerifiedRelease{})
		if err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		value, err := adapter.Trust(context)
		if !declaredResult(value, err, true, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		expectedEffects++
	}
	if capabilities.Has(harnessapi.NativeSkillDiscovery) || capabilities.Has(harnessapi.SkillProjection) {
		value, err := adapter.Skills(harnessapi.Context{}, harnessapi.ProjectScope)
		if !declaredResult(value, err, true, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		expectedEffects++
	} else {
		value, err := adapter.Skills(harnessapi.Context{}, harnessapi.ProjectScope)
		if !declaredResult(value, err, false, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
	}
	if capabilities.Has(harnessapi.MCPInspection) || capabilities.Has(harnessapi.NativeMCPWrite) {
		value, err := adapter.MCP(harnessapi.Context{}, harnessapi.UserScope)
		if !declaredResult(value, err, true, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		expectedEffects++
	} else {
		value, err := adapter.MCP(harnessapi.Context{}, harnessapi.UserScope)
		if !declaredResult(value, err, false, &effects) {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
	}
	value, err := adapter.ReleaseProbe(harnessapi.Context{})
	if capabilities.Has(harnessapi.ReleaseProbing) {
		if err != nil || reflect.ValueOf(value).IsZero() {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if observeErr := effects.observe(value, nil); observeErr != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		expectedEffects++
	} else if code(err) != harnessapi.CapabilityUnsupported || !reflect.ValueOf(value).IsZero() {
		return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	if effects.count != expectedEffects {
		return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	return nil
}

func declaredResult(value any, err error, declared bool, effects *effectSpy) bool {
	if declared {
		if err != nil || reflect.ValueOf(value).IsZero() {
			return false
		}
		return effects.observe(value, nil) == nil
	}
	return code(err) == harnessapi.OperationNotPorted && reflect.ValueOf(value).IsZero()
}

type falseCapability struct {
	harnessapi.Adapter
	capabilities harnessapi.CapabilitySet
}

func (a falseCapability) Capabilities() harnessapi.CapabilitySet {
	return a.capabilities
}

func wrongCapabilities(adapter harnessapi.Adapter) harnessapi.CapabilitySet {
	for capability := harnessapi.ManagedHome; capability <= harnessapi.ReleaseProbing; capability++ {
		if !adapter.Capabilities().Has(capability) {
			set, _ := harnessapi.NewCapabilitySet(capability)
			return set
		}
	}
	return harnessapi.CapabilitySet{}
}

func TestStagedOperationsAndCapabilities(t *testing.T) {
	var positive effectSpy
	if err := positive.observe(harnessapi.HomeLayout{}, nil); err != nil {
		t.Fatalf("positive effect observation failed: %v", err)
	}
	if positive.count != 1 {
		t.Fatal("effect spy positive control is disconnected")
	}
	for _, id := range harnessregistry.IDs() {
		t.Run(string(id), func(t *testing.T) {
			adapter, err := harnessregistry.Lookup(id)
			if err != nil {
				t.Fatal(err)
			}
			if err := stagedCapabilities(adapter); err != nil {
				t.Fatal(err)
			}
			if err := stagedCapabilities(falseCapability{Adapter: adapter, capabilities: wrongCapabilities(adapter)}); code(err) != harnessapi.RegistryMismatch {
				t.Fatalf("false claim escaped: %v", err)
			}
		})
	}
}
