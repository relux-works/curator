package swiftpminterop

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
)

// ResolutionClass records where one observed or declared read resolved.
type ResolutionClass string

// The closed resolution classes. Every header, module, framework, library, and
// SDK read must land in exactly one of the first two.
const (
	// ResolvedAdmitted means the read resolved inside an admitted package tree.
	ResolvedAdmitted ResolutionClass = "admitted_source"
	// ResolvedBinding means the read resolved inside exactly one C0/C4-selected
	// external toolchain, SDK, or sysroot binding node.
	ResolvedBinding ResolutionClass = "selected_binding"
	// ResolvedUndeclared means the read escaped the closure entirely.
	ResolvedUndeclared ResolutionClass = "undeclared"
)

// Resolution is the exact evidence for one path read.
type Resolution struct {
	Class        ResolutionClass
	Package      string
	Relative     string
	BindingRole  string
	BindingNode  closuregraph.ID
	AbsolutePath string
}

// admittedRoot is one admitted package tree available to containment.
type admittedRoot struct {
	identity string
	root     string
}

// trustedRoot is one C0/C4-selected external root outside the closure.
type trustedRoot struct {
	role   string
	root   string
	nodeID closuregraph.ID
}

// roots is the complete containment authority for one interop closure.
type roots struct {
	admitted []admittedRoot
	trusted  []trustedRoot
}

func (r *roots) addAdmitted(identity, root string) error {
	cleaned, err := absoluteRoot(root)
	if err != nil {
		return failFields(CodeGraphIncomplete, map[string]string{"package": identity}, "admitted package root is not an absolute real directory")
	}
	r.admitted = append(r.admitted, admittedRoot{identity: identity, root: cleaned})
	sort.Slice(r.admitted, func(i, j int) bool { return len(r.admitted[i].root) > len(r.admitted[j].root) })
	return nil
}

func (r *roots) addTrusted(role, root string, nodeID closuregraph.ID) error {
	cleaned, err := absoluteRoot(root)
	if err != nil {
		return failFields(CodeToolchainUntrusted, map[string]string{"role": role, "root": root}, "selected external root is not an absolute real directory")
	}
	for _, admitted := range r.admitted {
		if within(admitted.root, cleaned) || within(cleaned, admitted.root) {
			return failFields(CodeToolchainUntrusted, map[string]string{"role": role, "root": root}, "selected external root overlaps an admitted dependency tree")
		}
	}
	r.trusted = append(r.trusted, trustedRoot{role: role, root: cleaned, nodeID: nodeID})
	sort.Slice(r.trusted, func(i, j int) bool { return len(r.trusted[i].root) > len(r.trusted[j].root) })
	return nil
}

// resolve classifies one absolute candidate path. A path that is not a real
// regular file or directory, that traverses a symlink, or that lands outside
// every admitted and selected root is undeclared.
func (r *roots) resolve(candidate string) Resolution {
	cleaned := filepath.Clean(candidate)
	if !filepath.IsAbs(cleaned) {
		return Resolution{Class: ResolvedUndeclared, AbsolutePath: candidate}
	}
	for _, admitted := range r.admitted {
		if !within(admitted.root, cleaned) {
			continue
		}
		if !realPathWithin(admitted.root, cleaned) {
			return Resolution{Class: ResolvedUndeclared, AbsolutePath: cleaned}
		}
		relative, err := filepath.Rel(admitted.root, cleaned)
		if err != nil {
			return Resolution{Class: ResolvedUndeclared, AbsolutePath: cleaned}
		}
		return Resolution{Class: ResolvedAdmitted, Package: admitted.identity, Relative: filepath.ToSlash(relative), AbsolutePath: cleaned}
	}
	for _, trusted := range r.trusted {
		if !within(trusted.root, cleaned) {
			continue
		}
		if !presentNode(cleaned) {
			return Resolution{Class: ResolvedUndeclared, AbsolutePath: cleaned}
		}
		relative, err := filepath.Rel(trusted.root, cleaned)
		if err != nil {
			return Resolution{Class: ResolvedUndeclared, AbsolutePath: cleaned}
		}
		return Resolution{Class: ResolvedBinding, BindingRole: trusted.role, BindingNode: trusted.nodeID, Relative: filepath.ToSlash(relative), AbsolutePath: cleaned}
	}
	return Resolution{Class: ResolvedUndeclared, AbsolutePath: cleaned}
}

// resolveReference confines one module-map or include reference. Absolute and
// parent-escaping spellings are rejected before the filesystem is consulted so
// a missing file can never be mistaken for a contained one.
func (r *roots) resolveReference(baseDirectory, reference string) (Resolution, error) {
	if reference == "" || strings.ContainsAny(reference, "\x00\r\n") {
		return Resolution{}, failFields(CodeModuleMapEscape, map[string]string{"reference": reference}, "reference is empty or contains control characters")
	}
	if filepath.IsAbs(reference) || strings.HasPrefix(reference, "/") || windowsAbsolute(reference) {
		return Resolution{}, failFields(CodeModuleMapEscape, map[string]string{"reference": reference}, "reference names an absolute path outside the admitted package")
	}
	candidate := filepath.Clean(filepath.Join(baseDirectory, filepath.FromSlash(reference)))
	resolution := r.resolve(candidate)
	if resolution.Class == ResolvedUndeclared {
		return resolution, failFields(CodeModuleMapEscape, map[string]string{"reference": reference, "resolved": candidate}, "reference escapes the admitted closure and every selected binding root")
	}
	return resolution, nil
}

func windowsAbsolute(value string) bool {
	return len(value) >= 3 && value[1] == ':' && (value[2] == '\\' || value[2] == '/')
}

func absoluteRoot(value string) (string, error) {
	if value == "" {
		return "", os.ErrInvalid
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", os.ErrInvalid
	}
	return filepath.Clean(absolute), nil
}

// presentNode reports whether candidate exists as a real file or directory.
// A selected external root grants trust to the bytes it actually contains, not
// to every path spelling that happens to start with it.
func presentNode(candidate string) bool {
	info, err := os.Stat(candidate)
	return err == nil && (info.Mode().IsRegular() || info.IsDir())
}

func within(root, candidate string) bool {
	if root == candidate {
		return true
	}
	return strings.HasPrefix(candidate, root+string(filepath.Separator))
}

// realPathWithin walks every component below root and rejects symlinks and
// non-regular nodes so an admitted tree cannot forward a read outside itself.
func realPathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	if relative == "." {
		return true
	}
	current := root
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, statErr := os.Lstat(current)
		if statErr != nil {
			return false
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return false
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return false
		}
	}
	return true
}
