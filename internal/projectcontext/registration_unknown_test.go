package projectcontext

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
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

// Claim: fresh public registration succeeds on qualified platforms and keeps
// the existing unsupported-platform refusal before writes elsewhere. The equal
// binding control is portable; it does not qualify a native filesystem.
func TestRegistrationFilesystemQualification(t *testing.T) {
	r := registered{Binding: binding{Filesystem: map[string]string{"inode": "1"}}}
	r.Record.RootIdentity = identity(r.Binding)
	if err := verifyBindingResult(r, r.Binding, nil); err != nil {
		t.Fatal("equal binding control", err)
	}
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		f := newFixture(t)
		got, err := f.store.Inspect(f.registration.CheckoutID)
		if err != nil || got.Registration.CheckoutID != f.registration.CheckoutID {
			t.Fatal("qualified registration control", err)
		}
		return
	}
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "manager")
	store := Store{Home: home}
	got, err := store.Register("fixture", root, "base", "codex_cli")
	if !isCode(err, "project_context_unsupported") || !reflect.DeepEqual(got, Registration{}) {
		t.Fatalf("unqualified registration: got %v, want project_context_unsupported and no registration", err)
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatal("unsupported registration wrote manager state", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatal("unsupported registration changed checkout", err)
	}
}
