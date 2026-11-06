//go:build linux || darwin

package projectcontext

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Run a real fixture operation after the first successful Git observation and
// before the following observations. The sentinel fires once, not as a retry.
func bindingWindowPlant(t *testing.T, action string) string {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	sentinel := filepath.Join(dir, "fired")
	script := fmt.Sprintf(`#!/bin/sh
%[1]q "$@"; rc=$?
case " $* " in
 *" rev-parse --show-toplevel "*)
  if [ ! -e %[2]q ]; then
   : > %[2]q
   %[3]s
  fi
 ;;
esac
exit "$rc"
`, realGit, sentinel, action)
	if err = os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Claim: an actual branch switch inside the binding read window preserves the
// unchanged tuple and both records. An admitted mode change is not identity. A real root replacement in the same window
// refuses and durably lapses; restoring that original inode cannot revive it.
// The file-content same() predicate retains its metadata checks elsewhere.
func TestRegistrationRootReadWindow(t *testing.T) {
	for _, kind := range []string{"ordinary-checkout", "valid-mode-change", "replaced-root"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			id := f.registration.CheckoutID
			gitTest(t, f.root, "checkout", "-q", "-b", "feature")
			write(t, filepath.Join(f.root, "branch-entry"), "ordinary branch content\n")
			gitTest(t, f.root, "add", "branch-entry")
			gitTest(t, f.root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "ordinary branch")
			base := gitTest(t, f.root, "rev-parse", "HEAD~1")
			gitTest(t, f.root, "checkout", "-q", "--detach", base)
			if _, err := f.store.Inspect(id); err != nil {
				t.Fatal("unchanged tuple control", err)
			}
			before := readBindingRecords(t, f)
			realGit, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			saved := f.root + "-saved"
			action := fmt.Sprintf("%q -C %q checkout -q feature", realGit, f.root)
			if kind == "valid-mode-change" {
				action = fmt.Sprintf("/bin/chmod 500 %q", f.root)
			}
			if kind == "replaced-root" {
				action = fmt.Sprintf("/bin/mv %q %q && /bin/mkdir %q && /bin/mv %q %q", f.root, saved, f.root, filepath.Join(saved, ".git"), filepath.Join(f.root, ".git"))
			}
			dir := bindingWindowPlant(t, action)
			orig := os.Getenv("PATH")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
			got, err := f.store.Inspect(id)
			t.Setenv("PATH", orig)
			if kind != "replaced-root" {
				if err != nil || got.Registration.CheckoutID != id {
					t.Fatal("ordinary checkout lapsed an unchanged tuple", err)
				}
				if _, err = os.Stat(filepath.Join(f.root, "branch-entry")); kind == "ordinary-checkout" && err != nil {
					t.Fatal("real checkout plant did not run", err)
				}
				live, e := bind(f.root)
				if e != nil || identity(live) != f.registration.RootIdentity {
					t.Fatal("checkout changed tuple", e)
				}
				sameBindingRecords(t, f, before)
				if e := os.Chmod(f.root, 0700); e != nil {
					t.Fatal(e)
				}
				gitTest(t, f.root, "checkout", "-q", "--detach", base)
				if _, err = f.store.Inspect(id); err != nil {
					t.Fatal("restored branch control", err)
				}
				sameBindingRecords(t, f, before)
			} else {
				if !isCode(err, "project_registration_mismatch") || got.Registration.CheckoutID != "" || !strings.Contains(err.Error(), "root changed while binding") {
					t.Fatal("replaced root authorized or wrong refusal", err)
				}
				retainedBindingLapse(t, f, before)
				if err = os.Rename(filepath.Join(f.root, ".git"), filepath.Join(saved, ".git")); err != nil {
					t.Fatal(err)
				}
				if err = os.Remove(f.root); err != nil {
					t.Fatal(err)
				}
				if err = os.Rename(saved, f.root); err != nil {
					t.Fatal(err)
				}
				if _, err = f.store.Inspect(id); !isCode(err, "project_registration_mismatch") || !strings.Contains(err.Error(), "authority lapsed") {
					t.Fatal("restored root revived lapse", err)
				}
				retainedBindingLapse(t, f, before)
			}
		})
	}
}
