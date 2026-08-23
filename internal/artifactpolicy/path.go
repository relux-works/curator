package artifactpolicy

import (
	"encoding/base64"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// VirtualPath is a validated portable archive or tree path.
type VirtualPath struct {
	Canonical    string
	CollisionKey string
	Components   []string
}

type pathFailure struct {
	reason       string
	collisionKey string
	limitName    string
	limit        int64
	observed     int64
}

func (failure *pathFailure) Error() string {
	return failure.reason
}

// ValidateVirtualPath requires exact NFC UTF-8, slash separators, portable
// components, and the fixed policy-v1 path limits. It never cleans an unsafe
// name into a different name.
func ValidateVirtualPath(name string) (VirtualPath, error) {
	return validateVirtualPath(name, DefaultLimits())
}

func validateVirtualPath(name string, limits LimitVector) (VirtualPath, error) {
	fail := func(reason string) (VirtualPath, error) {
		return VirtualPath{}, &pathFailure{reason: reason}
	}
	if name == "" {
		return fail("empty_path")
	}
	if !utf8.ValidString(name) {
		return fail("invalid_utf8")
	}
	if norm.NFC.String(name) != name {
		return fail("not_nfc")
	}
	if int64(len(name)) > limits.MaxPathBytes {
		return VirtualPath{}, &pathFailure{
			reason: "path_too_long", limitName: "max_path_bytes",
			limit: limits.MaxPathBytes, observed: int64(len(name)),
		}
	}
	if strings.ContainsRune(name, '\\') {
		return fail("backslash_separator")
	}
	if strings.Contains(name, "!/") {
		return fail("reserved_container_separator")
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "//") {
		return fail("absolute_path")
	}
	for _, character := range name {
		if character == 0 || unicode.IsControl(character) {
			return fail("control_character")
		}
	}
	components := strings.Split(name, "/")
	for _, component := range components {
		if component == "" || component == "." || component == ".." {
			return fail("nonportable_component")
		}
		if int64(len(component)) > limits.MaxComponentBytes {
			return VirtualPath{}, &pathFailure{
				reason: "component_too_long", limitName: "max_component_bytes",
				limit: limits.MaxComponentBytes, observed: int64(len(component)),
			}
		}
		if strings.ContainsRune(component, ':') {
			return fail("drive_or_alternate_stream")
		}
		if strings.HasSuffix(component, ".") || strings.HasSuffix(component, " ") {
			return fail("windows_trailing_dot_or_space")
		}
		if windowsDeviceComponent(component) {
			return fail("windows_device_name")
		}
	}
	if path.Clean(name) != name {
		return fail("noncanonical_path")
	}
	return VirtualPath{
		Canonical: name, CollisionKey: portableCollisionKey(name),
		Components: append([]string(nil), components...),
	}, nil
}

func windowsDeviceComponent(component string) bool {
	base := component
	if dot := strings.IndexByte(base, '.'); dot >= 0 {
		base = base[:dot]
	}
	base = strings.ToUpper(base)
	switch base {
	case "CON", "PRN", "AUX", "NUL", "CLOCK$":
		return true
	}
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) {
		return base[3] >= '1' && base[3] <= '9'
	}
	return false
}

func portableCollisionKey(name string) string {
	components := strings.Split(norm.NFC.String(name), "/")
	for index, component := range components {
		component = strings.TrimRight(component, ". ")
		components[index] = cases.Fold().String(component)
	}
	return norm.NFC.String(strings.Join(components, "/"))
}

func joinTreePath(root, relative string) string {
	if relative == "" {
		return root
	}
	return root + "/" + relative
}

func joinContainerPath(container, member string) string {
	return container + "!/" + member
}

func originalNameBase64(name string) string {
	return base64.StdEncoding.EncodeToString([]byte(name))
}

func containerPathDiagnostic(parent, original string, chain []string, err error) Diagnostic {
	return unsafePathDiagnostic(
		joinContainerPath(parent, original), original,
		append(append([]string(nil), chain...), parent), err,
	)
}

func treePathDiagnostic(parent, original string, err error) Diagnostic {
	return unsafePathDiagnostic(joinTreePath(parent, original), original, nil, err)
}

func unsafePathDiagnostic(pathValue, original string, chain []string, err error) Diagnostic {
	diagnostic := Diagnostic{
		Code: CodeArchiveUnsafePath, Path: pathValue,
		OriginalNameBase64: originalNameBase64(original),
		ContainerChain:     append([]string(nil), chain...), Reason: err.Error(),
	}
	var failure *pathFailure
	if ok := errorAs(err, &failure); ok {
		diagnostic.CollisionKey = failure.collisionKey
		diagnostic.LimitName = failure.limitName
		diagnostic.Limit = failure.limit
		diagnostic.Observed = failure.observed
	}
	return diagnostic
}
