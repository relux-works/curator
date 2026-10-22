package projectcontext

import (
	"errors"
	"io/fs"
	"strings"

	"github.com/relux-works/curator/internal/stateread"
)

func isCode(err error, code string) bool {
	return err != nil && strings.HasPrefix(err.Error(), code+":")
}

// The public project boundary owns these diagnostics. Owning validators keep
// their own error types internally; filesystem details must not leak paths.
func projectError(err error) error {
	if err == nil {
		return nil
	}
	for _, code := range []string{"project_context_invalid", "project_registration_mismatch", "project_item_unapproved", "project_context_drift", "project_context_unreadable", "project_context_unsupported", "project_view_busy", "project_lease_mismatch", "project_path_refused"} {
		if isCode(err, code) {
			return err
		}
	}
	var pathErr *fs.PathError
	var readErr *stateread.Error
	if errors.As(err, &readErr) || errors.As(err, &pathErr) || errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrNotExist) {
		return fail("project_context_unreadable", "cannot read required project evidence")
	}
	return fail("project_context_invalid", "project declaration, manifest or closure validation failed")
}
