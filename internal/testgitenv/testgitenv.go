// Package testgitenv keeps a test process from inheriting the operator's
// git configuration. A host whose ~/.gitconfig enables commit or tag
// signing (or rewrites URLs, or installs hooks) would otherwise leak into
// every fixture repository a test builds, so a commit a test expects to be
// unsigned comes out signed.
//
// TestMain in every package whose tests run git calls Isolate before
// m.Run. A test that needs its own global config (an insteadOf rewrite, a
// credential helper) still sets GIT_CONFIG_GLOBAL with t.Setenv, which
// overrides the process-wide value for that test. Production code never
// imports this package.
package testgitenv

import (
	"fmt"
	"os"
	"path/filepath"
)

// Isolate points GIT_CONFIG_GLOBAL at an empty file it owns and sets
// GIT_CONFIG_NOSYSTEM=1 for the rest of the process. It replaces any
// inherited value: an ambient GIT_CONFIG_GLOBAL is exactly what it must
// not trust. The returned function removes the scratch file; call it after
// m.Run. A failure to create the file is fatal: a test run that silently
// falls back to the ambient config is the defect this guards against.
func Isolate() func() {
	dir, err := os.MkdirTemp("", "curator-test-gitconfig-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: create isolated git config dir:", err)
		os.Exit(2)
	}
	path := filepath.Join(dir, "gitconfig")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "testgitenv: create isolated git config:", err)
		os.Exit(2)
	}
	_ = os.Setenv("GIT_CONFIG_GLOBAL", path)
	_ = os.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	return func() { _ = os.RemoveAll(dir) }
}
