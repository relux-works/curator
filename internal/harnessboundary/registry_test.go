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
	vh, err := adapter.Home(c)
	errors = append(errors, effects.observe(vh, err))
	vi, err := adapter.Instructions(c, harnessapi.Monolithic)
	errors = append(errors, effects.observe(vi, err))
	if !capabilities.Has(harnessapi.SkillProjection) && !capabilities.Has(harnessapi.NativeSkillDiscovery) {
		vs, err := adapter.Skills(c, harnessapi.ProjectScope)
		errors = append(errors, effects.observe(vs, err))
	}
	vset, err := adapter.Settings(c)
	errors = append(errors, effects.observe(vset, err))
	// MCP is checked separately because one adapter declares an unsupported
	// channel while the remaining operations stay staged.
	vc, err := adapter.Credentials(c)
	errors = append(errors, effects.observe(vc, err))
	vt, err := adapter.Trust(c)
	errors = append(errors, effects.observe(vt, err))
	vr, err := adapter.Preflight(c, s, r)
	errors = append(errors, effects.observe(vr, err))
	vp, err := adapter.PlanSeed(c, s, r)
	errors = append(errors, effects.observe(vp, err))
	vpc, err := adapter.ProjectContext(c, s, r)
	errors = append(errors, effects.observe(vpc, err))
	vstatus, err := adapter.Status(c, s)
	errors = append(errors, effects.observe(vstatus, err))
	vd, err := adapter.Drift(c, s)
	errors = append(errors, effects.observe(vd, err))
	return errors
}

// Staged capabilities describe implementations, rather than future native
// support. A false operational declaration fails this same conformance check.
func stagedCapabilities(adapter harnessapi.Adapter) error {
	capabilities := adapter.Capabilities()
	if adapter.ID() == "pi" {
		want, _ := harnessapi.NewCapabilitySet(
			harnessapi.ManagedHome,
			harnessapi.InstructionOutput,
			harnessapi.SkillProjection,
			harnessapi.NativeSkillDiscovery,
			harnessapi.CredentialMetadata,
			harnessapi.TrustObservation,
			harnessapi.ReleaseProbing,
		)
		if capabilities != want {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		ctx, snapshot, request := harnessapi.Context{}, harnessapi.Snapshot{}, harnessapi.Request{}
		if _, err := adapter.Home(ctx); err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Instructions(ctx, harnessapi.Monolithic); err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Skills(ctx, harnessapi.UserScope); err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Skills(ctx, harnessapi.ProjectScope); code(err) != harnessapi.CapabilityUnsupported {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Credentials(ctx); err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Trust(ctx); err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.ReleaseProbe(ctx); err != nil {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.MCP(ctx, harnessapi.UserScope); code(err) != harnessapi.CapabilityUnsupported {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.ParseReleaseProbe(ctx, harnessapi.ProbeOutput{}); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Settings(ctx); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Preflight(ctx, snapshot, request); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.PlanSeed(ctx, snapshot, request); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.ProjectContext(ctx, snapshot, request); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Status(ctx, snapshot); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		if _, err := adapter.Drift(ctx, snapshot); code(err) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
		return nil
	}
	for _, capability := range capabilities.Values() {
		switch capability {
		case harnessapi.ReleaseProbing:
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
	if !capabilities.Has(harnessapi.MCPInspection) && !capabilities.Has(harnessapi.NativeMCPWrite) {
		mcp, mcpErr := adapter.MCP(harnessapi.Context{}, harnessapi.UserScope)
		if code(effects.observe(mcp, mcpErr)) != harnessapi.OperationNotPorted {
			return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
		}
	}
	if effects.count != 0 {
		return harnessapi.NewRefusal(harnessapi.RegistryMismatch)
	}
	expectedEffects := 0
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

type falseCapability struct{ harnessapi.Adapter }

func (falseCapability) Capabilities() harnessapi.CapabilitySet {
	set, _ := harnessapi.NewCapabilitySet(harnessapi.ManagedHome)
	return set
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
			if err := stagedCapabilities(falseCapability{adapter}); code(err) != harnessapi.RegistryMismatch {
				t.Fatalf("false claim escaped: %v", err)
			}
		})
	}
}
