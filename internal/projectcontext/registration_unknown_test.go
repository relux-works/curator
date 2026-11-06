package projectcontext

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
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
		{"repository-exit128-unknown", &localGitError{class: "exit", code: 128}, true},
		{"unreadable-metadata", bindingUnknown("root cannot be verified"), true},
		{"unknown-error-open", errors.New("/do-not-report"), false},
		{"missing-root-open", bindingPathReadFailure(os.ErrNotExist, "root is missing"), false},
		{"parent-file-open", bindingPathReadFailure(syscall.ENOTDIR, "root unavailable"), false},
		{"symlink-loop-open", bindingPathReadFailure(errors.New("loop"), "root unavailable"), false},
		{"io-read", bindingPathReadFailure(syscall.EIO, "root unavailable"), true},
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

// Claim: a successfully observed differing component takes precedence over a
// later unknown failure. An unobserved component cannot manufacture a change;
// equal observations are valid controls and classification mutates no inputs.
func TestBindingPartialObservation(t *testing.T) {
	remote := "https://example.invalid/recorded.git"
	saved := binding{Repository: repository{Root: "/checkout", PrimaryRemote: &remote, RootCommit: strings.Repeat("1", 40)}, Filesystem: map[string]string{"device": "3", "inode": "4"}}
	r := registered{Binding: saved}
	r.Record.RootIdentity = identity(saved)
	for _, field := range []string{"root", "remote", "commit", "filesystem"} {
		t.Run(field, func(t *testing.T) {
			if err := verifyBindingResult(r, saved, nil); err != nil {
				t.Fatal("equal control", err)
			}
			live := binding{Repository: saved.Repository, Filesystem: map[string]string{"device": "3", "inode": "4"}}
			observed := &bindingObservationError{error: &localGitError{site: "config", class: "start refused", errno: "EACCES", code: -1}}
			want := ""
			switch field {
			case "root":
				live.Repository.Root = "/different"
				observed.root = true
				want = "repository.root"
			case "remote":
				live.Repository.PrimaryRemote = nil
				observed.remote = true
				want = "repository.primary_remote"
			case "commit":
				live.Repository.RootCommit = strings.Repeat("2", 40)
				observed.commit = true
				want = "repository.root_commit"
			case "filesystem":
				live.Filesystem["inode"] = "5"
				observed.filesystem = true
				want = "filesystem.inode"
			}
			before := identity(struct {
				Saved registered
				Live  binding
			}{r, live})
			err := verifyBindingResult(r, live, observed)
			var unknown *bindingUnknownError
			if !isCode(err, "project_registration_mismatch") || errors.As(err, &unknown) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "start refused (EACCES)") || !strings.Contains(err.Error(), "explicitly rebind") {
				t.Fatal("observed difference discarded", err)
			}
			if identity(struct {
				Saved registered
				Live  binding
			}{r, live}) != before {
				t.Fatal("classification mutated binding")
			}
			observed.root, observed.remote, observed.commit, observed.filesystem = false, false, false, false
			err = verifyBindingResult(r, live, observed)
			unknown = nil
			if !errors.As(err, &unknown) || strings.Contains(err.Error(), "explicitly rebind") {
				t.Fatal("unobserved field manufactured a lapse", err)
			}
		})
	}
}
