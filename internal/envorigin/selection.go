package envorigin

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Account is raw OS evidence. A lookup seam supplies no selector, native root,
// containment verdict or final authority decision.
type Account struct {
	UID, RealUID, EffectiveUID uint64
	Home                       string
}
type AccountLookup func() (Account, error)

// Select is initial provisioning only. Existing homes must use Read and never
// call this function, regardless of current HOME/selectors or observed links.
func Select(manager string, b Binding, adapter string, lookup AccountLookup) (*Record, error) {
	selected, err := selectProvisioningNative(manager, adapter, lookup)
	if err != nil {
		return nil, err
	}
	r := &Record{Record: Version, Adapter: adapter, OperatorHome: selected.OperatorHome, Selector: selected.Selector, NativeRoot: selected.NativeRoot, Capture: Capture{Method: "provisioning", Authority: "native-selection", Time: time.Now().UTC().Format(time.RFC3339Nano)}, Binding: b}
	if _, err := r.Marshal(); err != nil {
		return nil, err
	}
	if b.ManagedTarget != ManagedTarget(manager, b.Profile, adapter) {
		return nil, invalid("capture target namespace mismatch")
	}
	return r, nil
}

// provisioningSelection is an initial credential-origin discovery observation, never an existing home's
// credential authority. It preserves raw selector presence independently of HOME.
type provisioningSelection struct {
	OperatorHome string
	Selector     Selector
	NativeRoot   string
}

// selectProvisioningNative applies account/physical policy only for initial credential-origin discovery.
// A raw-account seam supplies no path selection or final authority verdict.
func selectProvisioningNative(manager, adapter string, lookup AccountLookup) (*provisioningSelection, error) {
	real, effective, err := processAccountIdentity()
	if err != nil {
		return nil, err
	}
	if real != effective {
		return nil, invalid("ambiguous real/effective process identity")
	}
	if lookup == nil {
		lookup = LookupAccount
	}
	a, err := lookup()
	if err != nil {
		return nil, invalid("trusted account lookup unavailable: %v", err)
	}
	if a.UID != effective || a.RealUID != real || a.EffectiveUID != effective || !cleanAbsolute(a.Home) {
		return nil, invalid("ambiguous account identity or invalid account home")
	}
	if err := ValidateExternal(manager, a.Home); err != nil {
		return nil, err
	}
	name := selectorName(adapter)
	if name == "" {
		return nil, invalid("unsupported adapter")
	}
	value, present := os.LookupEnv(name)
	sel := Selector{Name: name, State: "absent"}
	if present {
		sel.Value = &value
		sel.State = "present-value"
		if value == "" {
			sel.State = "present-empty"
		}
	}
	root := ""
	if present && value != "" {
		if hasControl(value) {
			return nil, invalid("selector contains control characters")
		}
		raw := value
		if !filepath.IsAbs(raw) {
			cwd, err := os.Getwd()
			if err != nil || !cleanAbsolute(cwd) {
				return nil, invalid("cannot establish absolute launch cwd")
			}
			raw = filepath.Join(cwd, raw)
		}
		raw = filepath.Clean(raw)
		managed, err := IsManaged(manager, raw)
		if err != nil {
			return nil, err
		}
		if !managed {
			root = raw
			if adapter == "opencode" || adapter == "muse" {
				root = filepath.Join(raw, adapter)
			}
			managed, err = IsManaged(manager, root)
			if err != nil {
				return nil, err
			}
			if managed {
				root = ""
			}
		}
	}
	if root == "" {
		root, err = defaultRoot(a.Home, adapter)
		if err != nil {
			return nil, err
		}
	}
	if err := ValidateExternal(manager, root); err != nil {
		return nil, err
	}
	return &provisioningSelection{OperatorHome: a.Home, Selector: sel, NativeRoot: root}, nil
}

func defaultRoot(anchor, adapter string) (string, error) {
	switch adapter {
	case "codex_cli":
		return filepath.Join(anchor, ".codex"), nil
	case "claude_code":
		return filepath.Join(anchor, ".claude"), nil
	case "pi":
		return filepath.Join(anchor, ".pi", "agent"), nil
	case "opencode", "muse":
		return filepath.Join(anchor, ".config", adapter), nil
	}
	return "", fmt.Errorf("%s: unsupported adapter", DiagInvalid)
}
