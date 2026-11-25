package envprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/pathboundary"
)

// DiagNativeHomeBindingRequired refuses an absent or inadmissible native target.
const DiagNativeHomeBindingRequired = "native_home_binding_required"

// NativeHomeBindings names exact native adapter directories admitted by the
// caller. Parent-valued layouts bind their actual child directory, not the
// selector parent. Bindings grant no takeover or credential authority;
// operation gates still apply.
type NativeHomeBindings map[string]string

// nativeHomeBindingAdmission carries the directory identity observed before
// validation. Purge must compare an opened root to this identity instead of
// blessing whatever a later pathname lookup happens to find.
type nativeHomeBindingAdmission struct {
	path string
	info os.FileInfo
}

// ValidateNativeHomeBindingValue checks an explicitly supplied binding's syntax
// without inspecting a home. Resolution performs the filesystem checks later.
func ValidateNativeHomeBindingValue(id, home string) error {
	if _, err := envregistry.ByID(id); err != nil {
		return fmt.Errorf("%s: %w", DiagNativeHomeBindingRequired, err)
	}
	if home == "" {
		return fmt.Errorf("%s: bind the native home for %s explicitly", DiagNativeHomeBindingRequired, id)
	}
	if strings.IndexByte(home, 0) >= 0 || !filepath.IsAbs(home) || filepath.Clean(home) != home || filepath.Dir(home) == home {
		return fmt.Errorf("%s: %s needs a clean absolute directory other than a filesystem root", DiagNativeHomeBindingRequired, id)
	}
	return nil
}

// ResolveNativeHomeBinding selects and validates the exact bound directory
// before any marker inspection. Neither the environment nor HOME supplies a
// binding. The manager argument identifies state that cannot be a native home.
func ResolveNativeHomeBinding(manager string, bindings NativeHomeBindings, id string) (string, error) {
	if err := ValidateNativeHomeBindingValue(id, bindings[id]); err != nil {
		return "", err
	}
	var admitted []nativeHomeBindingAdmission
	selected := ""
	for _, key := range bindingKeys(bindings) {
		binding, err := admitNativeHomeBinding(manager, bindings, key)
		if err != nil {
			return "", err
		}
		for _, prior := range admitted {
			if nativeHomeBindingsOverlap(prior, binding) {
				return "", fmt.Errorf("%s: native bindings overlap", DiagNativeHomeBindingRequired)
			}
		}
		admitted = append(admitted, binding)
		if key == id {
			selected = binding.path
		}
	}
	return selected, nil
}

func admitNativeHomeBinding(manager string, bindings NativeHomeBindings, id string) (nativeHomeBindingAdmission, error) {
	home := bindings[id]
	if err := ValidateNativeHomeBindingValue(id, home); err != nil {
		return nativeHomeBindingAdmission{}, err
	}
	before, err := os.Lstat(home)
	if err != nil {
		return nativeHomeBindingAdmission{}, fmt.Errorf("%s: inspect bound native directory: %w", DiagNativeHomeBindingRequired, err)
	}
	if !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nativeHomeBindingAdmission{}, fmt.Errorf("%s: bound native path is not a direct directory", DiagNativeHomeBindingRequired)
	}
	if _, err := resolveNativeHomeBindingWithOwner(manager, bindings, id, nil); err != nil {
		return nativeHomeBindingAdmission{}, err
	}
	after, err := os.Lstat(home)
	if err != nil {
		return nativeHomeBindingAdmission{}, fmt.Errorf("%s: re-inspect bound native directory: %w", DiagNativeHomeBindingRequired, err)
	}
	if !after.IsDir() || after.Mode()&os.ModeSymlink != 0 || !os.SameFile(before, after) {
		return nativeHomeBindingAdmission{}, fmt.Errorf("%s: bound native directory changed during admission", DiagNativeHomeBindingRequired)
	}
	return nativeHomeBindingAdmission{path: home, info: after}, nil
}

func nativeHomeBindingsOverlap(left, right nativeHomeBindingAdmission) bool {
	return bindingPathsOverlap(left.path, right.path) || os.SameFile(left.info, right.info)
}

func resolveNativeHomeBindingWithOwner(manager string, bindings NativeHomeBindings, id string, owner pathboundary.OwnerLookup) (string, error) {
	home := bindings[id]
	if err := ValidateNativeHomeBindingValue(id, home); err != nil {
		return "", err
	}
	if err := validateNativeBindingTarget(manager, home, owner); err != nil {
		return "", fmt.Errorf("%s: %w", DiagNativeHomeBindingRequired, err)
	}
	return home, nil
}

// Freeze and validate the complete set before returning any target. Sorted
// adapter order makes both plans and refusals deterministic.
func nativeBindingResolver(manager string, bindings NativeHomeBindings, ids []string) (func(string) (nativeHomeBindingAdmission, error), error) {
	ids = append([]string(nil), ids...)
	sort.Strings(ids)
	frozen := make(map[string]nativeHomeBindingAdmission, len(ids))
	// Missing/syntactically invalid bindings precede all filesystem inspection.
	for _, id := range ids {
		if err := ValidateNativeHomeBindingValue(id, bindings[id]); err != nil {
			return nil, err
		}
	}
	keys := make([]string, 0, len(bindings))
	for id := range bindings {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	for _, id := range keys {
		if err := ValidateNativeHomeBindingValue(id, bindings[id]); err != nil {
			return nil, err
		}
	}
	var admitted []nativeHomeBindingAdmission
	for _, id := range ids {
		binding, err := admitNativeHomeBinding(manager, bindings, id)
		if err != nil {
			return nil, err
		}
		for _, prior := range admitted {
			if nativeHomeBindingsOverlap(prior, binding) {
				return nil, fmt.Errorf("%s: native bindings overlap", DiagNativeHomeBindingRequired)
			}
		}
		admitted = append(admitted, binding)
		frozen[id] = binding
	}
	return func(id string) (nativeHomeBindingAdmission, error) {
		binding, ok := frozen[id]
		if !ok {
			return nativeHomeBindingAdmission{}, fmt.Errorf("%s: adapter is outside the admitted target set", DiagNativeHomeBindingRequired)
		}
		return binding, nil
	}, nil
}
