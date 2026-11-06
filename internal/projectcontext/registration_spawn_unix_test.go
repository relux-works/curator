//go:build linux || darwin

package projectcontext

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// bindingGitPlant changes only fixture PATH. A non-executable interpreter makes
// the selected read fail in Start itself, rather than simulating an exit code.
// The exit control counts every invocation so a silent bind retry is detected.
func bindingGitPlant(t *testing.T, site int, mode string) (string, string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	count := filepath.Join(dir, "count")
	write(t, count, "0\n")
	interpreter := filepath.Join(dir, "not-executable")
	write(t, interpreter, "inert\n")
	denied := filepath.Join(dir, "denied")
	if err := os.WriteFile(denied, []byte("#!"+interpreter+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	script := fmt.Sprintf(`#!/bin/sh
read n < %[1]q
n=$((n+1)); echo "$n" > %[1]q
if [ "$n" -eq %[2]d ]; then
 case %[3]q in
 exit128) exit 128 ;;
 signal) kill -TERM $$ ;;
 timeout) exec /bin/sleep 30 ;;
 output) i=0; while [ "$i" -lt 16384 ]; do printf 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'; i=$((i+1)); done; exit 0 ;;
 esac
fi
%[4]q "$@"; rc=$?
if [ %[3]q = spawn ] && [ "$((n+1))" -eq %[2]d ]; then /bin/mv %[5]q %[6]q; fi
exit "$rc"
`, count, site, mode, realGit, denied, filepath.Join(dir, "git"))
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if mode == "spawn" && site == 1 {
		if err := os.Rename(denied, filepath.Join(dir, "git")); err != nil {
			t.Fatal(err)
		}
	}
	return dir, count
}

// Claim: each real bind read site preserves start/timeout/signal/output-bound
// classes, including a real EACCES spawn refusal, with no path in diagnostics.
// An exit failure followed by a healthy command is not retried. Direct binding
// leaves the checkout and saved tuple unchanged, and healthy controls still pass.
func TestBindingGitReadFailures(t *testing.T) {
	f := newFixture(t)
	b, err := bind(f.root)
	if err != nil {
		t.Fatal(err)
	}
	r := registered{Record: f.registration, Binding: b}
	orig := os.Getenv("PATH")
	for index, site := range []string{"rev-parse", "rev-list", "config"} {
		t.Run(site, func(t *testing.T) {
			for _, tc := range []struct{ mode, want string }{{"spawn", "start refused (EACCES)"}, {"exit128", "exit 128"}, {"signal", "signal"}, {"timeout", "timed out"}, {"output", "output bound exceeded"}} {
				t.Run(tc.mode, func(t *testing.T) {
					if err := verifyBinding(r); err != nil {
						t.Fatal("healthy control", err)
					}
					state := identity(r)
					config, err := os.ReadFile(filepath.Join(f.root, ".git", "config"))
					if err != nil {
						t.Fatal(err)
					}
					dir, count := bindingGitPlant(t, index+1, tc.mode)
					t.Setenv("PATH", dir+string(os.PathListSeparator)+orig)
					err = verifyBinding(r)
					want := "checkout binding recheck failed: local Git identity read failed at " + site + ": " + tc.want
					if err == nil || !strings.Contains(err.Error(), want) || strings.Contains(err.Error(), dir) {
						t.Fatalf("got %v; want %s without path", err, want)
					}
					calls, e := os.ReadFile(count)
					if e != nil {
						t.Fatal(e)
					}
					expected := index + 1
					if tc.mode == "spawn" {
						expected = index
					}
					if strings.TrimSpace(string(calls)) != strconv.Itoa(expected) {
						t.Fatalf("read retried: calls=%s want=%d", calls, expected)
					}
					after, e := os.ReadFile(filepath.Join(f.root, ".git", "config"))
					if e != nil || string(after) != string(config) || identity(r) != state {
						t.Fatal("refusal mutated checkout or tuple", e)
					}
					t.Setenv("PATH", orig)
					if err := verifyBinding(r); err != nil {
						t.Fatal("restored control", err)
					}
				})
			}
		})
	}
}
