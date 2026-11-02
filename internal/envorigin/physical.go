package envorigin

import (
	"fmt"
	"github.com/relux-works/curator/internal/pathboundary"
	"os"
	"path/filepath"
	"strings"
)

// route resolves existing ancestors and appends only proven absent components.
// Adapted bounded error/no-create/ancestor-SameFile kernels from public PR179,
// commits50f488db and2166919 (alexis); current-default authority is not reused.
// A route is an operation observation, not persisted hostile-same-UID identity.
type route struct {
	ancestor string
	info     os.FileInfo
	tail     []string
}

func inspectRoute(path string) (route, error) {
	if !cleanAbsolute(path) {
		return route{}, invalid("unclean or relative origin path")
	}
	var tail []string
	current := path
	for {
		info, err := os.Lstat(current)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return route{}, invalid("broken or unreadable origin route: %v", err)
			}
			actual, err := os.Stat(resolved)
			if err != nil || !actual.IsDir() {
				return route{}, invalid("origin ancestor is not a readable directory")
			}
			if info.Mode()&os.ModeSymlink == 0 && !info.IsDir() {
				return route{}, invalid("origin ancestor is not a directory")
			}
			handle, err := openDirectoryMetadata(resolved)
			if err != nil {
				return route{}, invalid("origin ancestor cannot be opened for directory metadata: %v", err)
			}
			opened, statErr := handle.Stat()
			closeErr := handle.Close()
			if statErr != nil || closeErr != nil || !opened.IsDir() || !os.SameFile(actual, opened) {
				return route{}, invalid("origin ancestor changed or failed during directory observation")
			}
			return route{ancestor: resolved, info: actual, tail: tail}, nil
		}
		if !os.IsNotExist(err) {
			return route{}, invalid("origin metadata: %v", err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return route{}, invalid("no existing origin ancestor")
		}
		tail = append([]string{filepath.Base(current)}, tail...)
		current = parent
	}
}
func below(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && (rel == "." || rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// IsManaged checks the original spelling first and then physical ancestor
// identities. It compares at the last existing managed ancestor even when the
// environments root is absent. Unknown missing-tail case aliases refuse.
func IsManaged(manager, path string) (bool, error) {
	if !cleanAbsolute(manager) || !cleanAbsolute(path) {
		return false, invalid("manager/native paths must be clean absolute")
	}
	base := filepath.Join(manager, "environments")
	if below(base, path) {
		return true, nil
	}
	br, err := inspectRoute(base)
	if err != nil {
		return false, err
	}
	cr, err := inspectRoute(path)
	if err != nil {
		return false, err
	}
	current := cr.ancestor
	suffix := append([]string{}, cr.tail...)
	for {
		info, err := os.Stat(current)
		if err != nil {
			return false, invalid("physical ancestor inspection: %v", err)
		}
		if os.SameFile(br.info, info) {
			if len(suffix) < len(br.tail) {
				return false, nil
			}
			exact, fold := true, true
			for i, v := range br.tail {
				exact = exact && v == suffix[i]
				fold = fold && strings.EqualFold(v, suffix[i])
			}
			if exact {
				return true, nil
			}
			if !fold {
				return false, nil
			}
			// Existing distinct entries prove case-sensitive separation. Both absent
			// spellings cannot prove separation, so never guess filesystem case rules.
			bp := filepath.Join(append([]string{br.ancestor}, br.tail...)...)
			cp := filepath.Join(append([]string{current}, suffix[:len(br.tail)]...)...)
			bi, be := os.Stat(bp)
			ci, ce := os.Stat(cp)
			if be == nil && ce == nil {
				return os.SameFile(bi, ci), nil
			}
			if os.IsNotExist(be) && ce == nil {
				return false, nil
			}
			if be != nil && !os.IsNotExist(be) || ce != nil && !os.IsNotExist(ce) {
				return false, invalid("managed alias inspection failed")
			}
			return false, invalid("ambiguous absent managed case alias")
		}
		parent := filepath.Dir(current)
		if parent == current {
			return false, nil
		}
		suffix = append([]string{filepath.Base(current)}, suffix...)
		current = parent
	}
}

// ValidateExternal proves outside-managed custody and a protected existing
// native ancestor without opening native settings/auth or creating a directory.
func ValidateExternal(manager, path string) error {
	managed, err := IsManaged(manager, path)
	if err != nil {
		return err
	}
	if managed {
		return invalid("native origin is inside managed environments")
	}
	r, err := inspectRoute(path)
	if err != nil {
		return err
	}
	if err := pathboundary.CheckPrivateFile(r.ancestor, r.info); err != nil {
		return invalid("native ancestor owner/protection: %v", err)
	}
	return nil
}

// NativeTarget returns adapter credential-relative metadata paths. It never
// opens the leaf; Pi's root is already the agent directory, not its parent.
func NativeTarget(root, adapter string) (string, error) {
	switch adapter {
	case "codex_cli", "pi", "muse":
		return filepath.Join(root, "auth.json"), nil
	case "claude_code":
		return filepath.Join(root, ".credentials.json"), nil
	case "opencode":
		return "", nil
	}
	return "", fmt.Errorf("%s: unsupported adapter", DiagInvalid)
}
