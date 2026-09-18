// Package contextmaterialize system-module admission (environments §3,
// §5.5, §12.1): a package is direct when it is the root itself, an active
// overlay, or named by the root's or an active overlay's requires.contexts
// entry; every other context member is transitive. Only the system modules
// of direct packages, and of transitive packages admitted by a
// system_module_waivers entry, are admitted: the system-prompt output and
// the launch fragment carry only admitted modules. Under the drop policy
// (default) a non-admitted applicable system module is skipped at
// materialization with the context_system_module_dropped warning; under
// error, resolution of the same module fails with
// context_system_module_transitive.
package contextmaterialize

import (
	"fmt"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextpkg"
)

// Admission diagnostics (environments §3.1, §5.7).
const (
	// DiagSystemModuleDropped warns that a transitive system module was
	// skipped at materialization under the drop policy, naming the
	// package and the module path.
	DiagSystemModuleDropped = "context_system_module_dropped"
	// DiagSystemModuleTransitive refuses a transitive system module under
	// the error policy, naming the package and the module path.
	DiagSystemModuleTransitive = "context_system_module_transitive"
)

// Transitive system-module policies (environments §12.1).
const (
	TransitiveDrop  = "drop"
	TransitiveError = "error"
)

// Admission is the effective §12.1 system-module admission policy:
// the transitive_system_modules value with the waived package names.
type Admission struct {
	// Transitive is drop or error; "" selects the drop default.
	Transitive string
	// Waivers admits transitive packages by name.
	Waivers map[string]bool
}

// Policy resolves the effective policy value, refusing anything outside
// the closed enum. Machine configuration validates the spelling at the
// reader; this guards programmatic callers.
func (a Admission) Policy() (string, error) {
	switch a.Transitive {
	case "", TransitiveDrop:
		return TransitiveDrop, nil
	case TransitiveError:
		return TransitiveError, nil
	default:
		return "", fmt.Errorf("transitive_system_modules %q is not drop or error", a.Transitive)
	}
}

// Waived reports whether a waiver admits the package's system modules.
func (a Admission) Waived(pkg string) bool {
	return a.Waivers[pkg]
}

// DirectSet computes the direct package set from the lock (environments
// §3): the root, every active overlay, and every package named by the
// root's or an active overlay's requires.contexts entry. The lock's
// required_by lists carry exactly those edges — one name, one kind, so an
// edge into a context member is a contexts edge — and the overlay flags
// mark the active overlays, so no manifest read is needed.
func DirectSet(lock *contextlock.Lock) map[string]bool {
	direct := map[string]bool{}
	if lock == nil {
		return direct
	}
	overlays := map[string]bool{}
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindContext {
			continue
		}
		if member.Overlay {
			overlays[member.Name] = true
		}
	}
	direct[lock.Root] = true
	for name := range overlays {
		direct[name] = true
	}
	for _, member := range lock.Members {
		if member.Kind != contextlock.KindContext || direct[member.Name] {
			continue
		}
		for _, requirer := range member.RequiredBy {
			if requirer == lock.Root || overlays[requirer] {
				direct[member.Name] = true
				break
			}
		}
	}
	return direct
}

// DroppedModule names one non-admitted system module by package and
// manifest path.
type DroppedModule struct {
	Package string
	Path    string
}

// ClassifiedModule is one admitted system module with its bytes.
type ClassifiedModule struct {
	Package string
	Module  Module
}

// TransitiveSystemModuleError is the §3 error-policy refusal, naming the
// package and the module path.
type TransitiveSystemModuleError struct {
	Package string
	Module  string
}

func (e *TransitiveSystemModuleError) Error() string {
	return fmt.Sprintf("%s: package %q system module %q is transitive and the transitive_system_modules policy is error",
		DiagSystemModuleTransitive, e.Package, e.Module)
}

// ClassifySystemModules splits every applicable system module of the
// emitted order into admitted and dropped (environments §3): a module is
// admitted when its package is direct or waived. Both lists follow emitted
// order with manifest order within a package, so dropped[0] is the module
// the error policy names. Members without package content contribute
// nothing; callers that need the content error check presence themselves.
func ClassifySystemModules(lock *contextlock.Lock, order []contextlock.Member, packages map[string]Package, environment string, admission Admission) (admitted []ClassifiedModule, dropped []DroppedModule) {
	direct := DirectSet(lock)
	for _, member := range order {
		pkg, ok := packages[member.Name]
		if !ok {
			continue
		}
		for _, module := range Applicable(pkg, "system", environment) {
			if direct[member.Name] || admission.Waived(member.Name) {
				admitted = append(admitted, ClassifiedModule{Package: member.Name, Module: module})
			} else {
				dropped = append(dropped, DroppedModule{Package: member.Name, Path: module.Path})
			}
		}
	}
	return admitted, dropped
}

// FirstTransitiveSystemModule reports the first non-admitted system module
// in emitted order (manifest order within a package) that applies to at
// least one of envIDs, or nil when the closure carries none. Under the
// drop policy it always reports nil: drop never fails resolution,
// installation, or update for admission. A module whose selector names no
// registered environment selects nothing (§3) and can never reach a
// system-prompt surface, so it never refuses resolution.
func FirstTransitiveSystemModule(lock *contextlock.Lock, precedence Precedence, modules map[string][]contextpkg.Module, admission Admission, envIDs []string) (*DroppedModule, error) {
	policy, err := admission.Policy()
	if err != nil {
		return nil, err
	}
	if policy != TransitiveError {
		return nil, nil
	}
	order, err := EmittedOrder(lock, precedence)
	if err != nil {
		return nil, err
	}
	direct := DirectSet(lock)
	for _, member := range order {
		if direct[member.Name] || admission.Waived(member.Name) {
			continue
		}
		for _, module := range modules[member.Name] {
			class := module.Class
			if class == "" {
				class = "root"
			}
			if class != "system" {
				continue
			}
			applies := false
			for _, env := range envIDs {
				if module.Applies(env) {
					applies = true
					break
				}
			}
			if !applies {
				continue
			}
			return &DroppedModule{Package: member.Name, Path: module.Path}, nil
		}
	}
	return nil, nil
}
