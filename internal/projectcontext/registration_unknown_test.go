package projectcontext

import (
	"errors"
	"os"
	"testing"
)

// Claim: only an indeterminate read gets the no-lapse disposition. Observed
// changed tuples and the explicitly open repository/root cases keep refusal
// and the conservative lapse disposition. Equal tuples remain valid controls.
func TestBindingObservationDisposition(t *testing.T) {
	for _, tc := range []struct {
		name    string
		err     error
		unknown bool
	}{
		{"start", &localGitError{class: "start refused", errno: "EACCES", code: -1}, true},
		{"timeout", &localGitError{class: "timed out", code: -1}, true},
		{"signal", &localGitError{class: "signal", code: -1}, true},
		{"output-bound", &localGitError{class: "output bound exceeded", code: -1}, true},
		{"indeterminate-exit", &localGitError{class: "exit", code: 129}, true},
		{"repository-exit128-open", &localGitError{class: "exit", code: 128}, false},
		{"unreadable-metadata", bindingUnknown("root cannot be verified"), true},
		{"unknown-error", errors.New("/do-not-report"), true},
		{"missing-root-open", bindingPathReadFailure(os.ErrNotExist, "root is missing"), false},
		{"permission-read", bindingPathReadFailure(os.ErrPermission, "root cannot be read"), true},
		{"multiple-roots-open", bindingFailure("repository root commit is unavailable or ambiguous"), false},
		{"ambiguous-remote-open", bindingFailure("primary remote is ambiguous"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := registered{Binding: binding{Filesystem: map[string]string{"inode": "1"}}}
			r.Record.RootIdentity = identity(r.Binding)
			before := identity(r)
			if err := verifyBindingResult(r, r.Binding, nil); err != nil {
				t.Fatal("equal control", err)
			}
			err := verifyBindingResult(r, r.Binding, tc.err)
			var unknown *bindingUnknownError
			if !isCode(err, "project_registration_mismatch") || errors.As(err, &unknown) != tc.unknown {
				t.Fatalf("got %v; unknown=%v want=%v", err, unknown != nil, tc.unknown)
			}
			if identity(r) != before {
				t.Fatal("classification mutated saved binding")
			}
			changed := binding{Filesystem: map[string]string{"inode": "2"}}
			err = verifyBindingResult(r, changed, nil)
			unknown = nil
			if !isCode(err, "project_registration_mismatch") || errors.As(err, &unknown) {
				t.Fatal("observed real change became unknown", err)
			}
		})
	}
}
