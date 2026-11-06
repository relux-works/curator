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

func inspectBindingControl(t *testing.T, f fixture) bindingRecords {
	t.Helper()
	v, err := f.store.Inspect(f.registration.CheckoutID)
	if err != nil || v.Registration.CheckoutID != f.registration.CheckoutID {
		t.Fatal("healthy tuple control", err)
	}
	return readBindingRecords(t, f)
}
func assertBindingUnknown(t *testing.T, f fixture, before bindingRecords, want string) {
	t.Helper()
	v, err := f.store.Inspect(f.registration.CheckoutID)
	if !isCode(err, "project_registration_mismatch") || v.Registration.CheckoutID != "" || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "no state changed") || !strings.Contains(err.Error(), "retrying is safe") || strings.Contains(err.Error(), "explicitly rebind") {
		t.Fatal("unknown authorized or wrong recovery advice", err)
	}
	sameBindingRecords(t, f, before)
}
func assertBindingLapsed(t *testing.T, f fixture, before bindingRecords, want string) {
	t.Helper()
	v, err := f.store.Inspect(f.registration.CheckoutID)
	if !isCode(err, "project_registration_mismatch") || v.Registration.CheckoutID != "" || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "explicitly rebind") {
		t.Fatal("observed change authorized or wrong refusal", err)
	}
	retainedBindingLapse(t, f, before)
}

// Claim: each already-read differing component survives a later real spawn
// refusal. Restoring before a healthy use cannot erase its durable lapse.
// The no-change/all-starts-refused control denies use without writes and recovers.
func TestRegistrationObservedBeforeFailure(t *testing.T) {
	for _, kind := range []string{"all-starts-refused", "root-before-rev-parse", "root-before-rev-list", "root-before-config", "commit-before-config"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			before := inspectBindingControl(t, f)
			restore := func() {}
			site := 1
			if strings.HasPrefix(kind, "root-before-") {
				site = map[string]int{"root-before-rev-parse": 1, "root-before-rev-list": 2, "root-before-config": 3}[kind]
				saved := f.root + "-saved"
				if err := os.Rename(f.root, saved); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(f.root, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(filepath.Join(saved, ".git"), filepath.Join(f.root, ".git")); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(filepath.Join(f.root, ".git"), filepath.Join(saved, ".git")); err != nil {
						t.Fatal(err)
					}
					if err := os.Remove(f.root); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(saved, f.root); err != nil {
						t.Fatal(err)
					}
				}
			} else if kind == "commit-before-config" {
				site = 3
				head, other := differentRootAncestor(t, f)
				gitTest(t, f.root, "update-ref", "HEAD", other)
				if got := gitTest(t, f.root, "rev-list", "--max-parents=0", "HEAD"); got != other {
					t.Fatal("ancestor did not change")
				}
				restore = func() { gitTest(t, f.root, "update-ref", "HEAD", head) }
			}
			dir, _ := bindingGitPlant(t, site, "spawn")
			orig := os.Getenv("PATH")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
			if kind == "all-starts-refused" {
				assertBindingUnknown(t, f, before, "at rev-parse: start refused (EACCES)")
			} else {
				field := "filesystem.inode"
				if kind == "commit-before-config" {
					field = "repository.root_commit"
				}
				v, err := f.store.Inspect(f.registration.CheckoutID)
				if !isCode(err, "project_registration_mismatch") || v.Registration.CheckoutID != "" || !strings.Contains(err.Error(), field) || !strings.Contains(err.Error(), "start refused (EACCES)") || !strings.Contains(err.Error(), "explicitly rebind") {
					t.Fatal("observed difference lost to later failure", err)
				}
				retainedBindingLapse(t, f, before)
			}
			t.Setenv("PATH", orig)
			restore()
			if kind == "all-starts-refused" {
				inspectBindingControl(t, f)
				sameBindingRecords(t, f, before)
			} else {
				assertBindingLapsed(t, f, before, "authority lapsed")
			}
		})
	}
}

func differentRootAncestor(t *testing.T, f fixture) (string, string) {
	t.Helper()
	head := gitTest(t, f.root, "rev-parse", "HEAD")
	other := filepath.Join(t.TempDir(), "other")
	initRepo(t, other)
	write(t, filepath.Join(other, "different"), "different ancestor\n")
	gitTest(t, other, "add", "different")
	gitTest(t, other, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--amend", "-qm", "different ancestor")
	oid := gitTest(t, other, "rev-parse", "HEAD")
	gitTest(t, f.root, "fetch", "-q", other, "HEAD")
	return head, oid
}

// Claim: real Git exit128 means no tuple answer for each tested cause, so use
// refuses without state/lapse writes and the same tuple works after recovery.
// No stderr grammar or retry distinguishes these cases.
func TestRegistrationGitExit128Unknown(t *testing.T) {
	for _, kind := range []string{"not-a-repository", "bad-config", "unreadable-config", "missing-object", "ownership-refused"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			before := inspectBindingControl(t, f)
			restore := func() {}
			orig := os.Getenv("PATH")
			switch kind {
			case "not-a-repository":
				original := filepath.Join(f.root, ".git")
				saved := f.root + "-git"
				if err := os.Rename(original, saved); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(saved, original); err != nil {
						t.Fatal(err)
					}
				}
			case "bad-config", "unreadable-config":
				cfg := filepath.Join(f.root, ".git", "config")
				was, err := os.ReadFile(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if kind == "bad-config" {
					write(t, cfg, string(was)+"\n[bad\n")
				} else if err = os.Chmod(cfg, 0); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Chmod(cfg, 0600); err != nil {
						t.Fatal(err)
					}
					write(t, cfg, string(was))
				}
			case "missing-object":
				head := gitTest(t, f.root, "rev-parse", "HEAD")
				obj := filepath.Join(f.root, ".git", "objects", head[:2], head[2:])
				saved := obj + "-saved"
				if err := os.Rename(obj, saved); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(saved, obj); err != nil {
						t.Fatal(err)
					}
				}
			case "ownership-refused":
				// Real Git's ownership refusal is forced with its test selector; no chown.
				// This witnesses the real exit, not a different-owner native filesystem.
				realGit, err := exec.LookPath("git")
				if err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				script := fmt.Sprintf("#!/bin/sh\nGIT_TEST_ASSUME_DIFFERENT_OWNER=1 exec %q \"$@\"\n", realGit)
				if err = os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
				restore = func() { t.Setenv("PATH", orig) }
			}
			assertBindingUnknown(t, f, before, "exit 128")
			restore()
			inspectBindingControl(t, f)
			sameBindingRecords(t, f, before)
		})
	}
}

// Claim: actual structural ancestor loss (ENOTDIR/loop) retains the conservative
// durable lapse. A genuine ancestor read denial stays unknown and recovers.
// Manager state is outside the affected ancestor; no root-mode policy is moved.
func TestRegistrationAncestorDisposition(t *testing.T) {
	for _, kind := range []string{"parent-file", "parent-loop", "ancestor-unreadable"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			parent := filepath.Join(filepath.Dir(f.root), "ancestor")
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(parent, "checkout")
			if err := os.Rename(f.root, root); err != nil {
				t.Fatal(err)
			}
			reg, err := f.store.Rebind(f.registration.CheckoutID, root)
			if err != nil {
				t.Fatal(err)
			}
			f.root, f.registration = reg.ProjectRoot, reg
			parent = filepath.Dir(f.root)
			before := inspectBindingControl(t, f)
			saved := parent + "-saved"
			var restore func()
			if kind == "ancestor-unreadable" {
				if err = os.Chmod(parent, 0); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Chmod(parent, 0700); err != nil {
						t.Fatal(err)
					}
				}
				t.Cleanup(restore)
				assertBindingUnknown(t, f, before, "filesystem metadata read failed: permission denied")
			} else {
				if err = os.Rename(parent, saved); err != nil {
					t.Fatal(err)
				}
				if kind == "parent-file" {
					write(t, parent, "not a directory\n")
				} else if err = os.Symlink(parent, parent); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Remove(parent); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(saved, parent); err != nil {
						t.Fatal(err)
					}
				}
				assertBindingLapsed(t, f, before, "checkout binding recheck failed")
			}
			restore()
			if kind == "ancestor-unreadable" {
				inspectBindingControl(t, f)
				sameBindingRecords(t, f, before)
			} else {
				assertBindingLapsed(t, f, before, "authority lapsed")
			}
		})
	}
}

// Claim: each owner-held open row still refuses and retains a durable lapse,
// even after the original condition is restored. These controls pin the holding
// decision; they do not decide a future policy for missing/mode/type/tuple rows.
func TestRegistrationHeldOpenRows(t *testing.T) {
	for _, kind := range []string{"root-mode-lost", "root-type-change", "exact-root", "ambiguous-remote", "noncanonical-remote", "multiple-roots"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			before := inspectBindingControl(t, f)
			restore := func() {}
			switch kind {
			case "root-mode-lost":
				if err := os.Chmod(f.root, 0100); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Chmod(f.root, 0700); err != nil {
						t.Fatal(err)
					}
				}
				t.Cleanup(restore)
			case "root-type-change":
				saved := f.root + "-saved"
				if err := os.Rename(f.root, saved); err != nil {
					t.Fatal(err)
				}
				write(t, f.root, "file instead\n")
				restore = func() {
					if err := os.Remove(f.root); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(saved, f.root); err != nil {
						t.Fatal(err)
					}
				}
			case "exact-root":
				original := filepath.Join(f.root, ".git")
				parentGit := filepath.Join(filepath.Dir(f.root), ".git")
				if err := os.Rename(original, parentGit); err != nil {
					t.Fatal(err)
				}
				restore = func() {
					if err := os.Rename(parentGit, original); err != nil {
						t.Fatal(err)
					}
				}
			case "ambiguous-remote", "noncanonical-remote":
				cfg := filepath.Join(f.root, ".git", "config")
				was, err := os.ReadFile(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if kind == "ambiguous-remote" {
					gitTest(t, f.root, "config", "--add", "remote.origin.url", "https://example.invalid/first.git")
					gitTest(t, f.root, "config", "--add", "remote.origin.url", "https://example.invalid/second.git")
				} else {
					gitTest(t, f.root, "config", "remote.origin.url", "/unadmitted/local")
				}
				restore = func() { write(t, cfg, string(was)) }
			case "multiple-roots":
				head, other := differentRootAncestor(t, f)
				tree := gitTest(t, f.root, "rev-parse", "HEAD^{tree}")
				merge := gitTest(t, f.root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit-tree", tree, "-p", head, "-p", other, "-m", "two roots")
				gitTest(t, f.root, "update-ref", "HEAD", merge)
				restore = func() { gitTest(t, f.root, "update-ref", "HEAD", head) }
			}
			assertBindingLapsed(t, f, before, "explicitly rebind")
			restore()
			assertBindingLapsed(t, f, before, "authority lapsed")
		})
	}
}

// Claim: even a final filesystem read failure cannot discard a remote answer
// already observed to differ. The same-remote control stays unknown, writes no
// records and recovers; restoring an observed remote change cannot renew it.
func TestRegistrationRemoteBeforeFilesystemFailure(t *testing.T) {
	for _, kind := range []string{"same-remote", "changed-remote"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			parent := filepath.Join(filepath.Dir(f.root), "ancestor")
			if err := os.Mkdir(parent, 0700); err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(parent, "checkout")
			if err := os.Rename(f.root, root); err != nil {
				t.Fatal(err)
			}
			reg, err := f.store.Rebind(f.registration.CheckoutID, root)
			if err != nil {
				t.Fatal(err)
			}
			f.root, f.registration = reg.ProjectRoot, reg
			parent = filepath.Dir(f.root)
			before := inspectBindingControl(t, f)
			cfg := filepath.Join(f.root, ".git", "config")
			was, err := os.ReadFile(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "changed-remote" {
				gitTest(t, f.root, "config", "remote.origin.url", "https://example.invalid/changed.git")
			}
			restoreMode := func() {
				if e := os.Chmod(parent, 0700); e != nil {
					t.Fatal(e)
				}
			}
			t.Cleanup(restoreMode)
			dir := bindingReadWindowPlant(t, "config --get-all remote.origin.url", fmt.Sprintf("/bin/chmod 000 %q", parent))
			orig := os.Getenv("PATH")
			t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
			if kind == "same-remote" {
				assertBindingUnknown(t, f, before, "filesystem metadata read failed: permission denied")
			} else {
				v, e := f.store.Inspect(f.registration.CheckoutID)
				if !isCode(e, "project_registration_mismatch") || v.Registration.CheckoutID != "" || !strings.Contains(e.Error(), "repository.primary_remote") || !strings.Contains(e.Error(), "filesystem metadata read failed: permission denied") || !strings.Contains(e.Error(), "explicitly rebind") {
					t.Fatal("observed remote lost to final read failure", e)
				}
				retainedBindingLapse(t, f, before)
			}
			t.Setenv("PATH", orig)
			restoreMode()
			write(t, cfg, string(was))
			if kind == "same-remote" {
				inspectBindingControl(t, f)
				sameBindingRecords(t, f, before)
			} else {
				assertBindingLapsed(t, f, before, "authority lapsed")
			}
		})
	}
}
