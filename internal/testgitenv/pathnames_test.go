package testgitenv

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// Claim: the probe accepts distinct ordinary names and reports a collapse it
// can observe: a duplicated name (the planted collision every platform can
// produce) is reported as not distinct, and two names differing only in case
// are distinct exactly when this filesystem does not fold case (checked
// independently by a stat). On Linux and macOS a final-space name and its
// trimmed sibling are distinct for git. Limit: the Windows behaviour of the
// final-space row is observed by hosted CI, not asserted here.
func TestGitRepositoryNamesDistinct(t *testing.T) {
	probe := func(names ...string) (bool, string) {
		t.Helper()
		ok, detail, err := GitRepositoryNamesDistinct(t.TempDir(), names...)
		if err != nil {
			t.Fatal(err)
		}
		return ok, detail
	}
	if ok, detail := probe("origin.git", "other.git"); !ok {
		t.Fatalf("distinct ordinary names reported as not distinct: %s", detail)
	}
	if ok, detail := probe("same.git", "same.git"); ok || detail == "" {
		t.Fatalf("a duplicated name was accepted as distinct (detail %q)", detail)
	}
	folds := hostFoldsCase(t)
	if ok, detail := probe("case.git", "CASE.git"); ok == folds {
		t.Fatalf("case-only names: distinct = %v, host folds case = %v (%s)", ok, folds, detail)
	}
	spaced, detail := probe("origin.git ", "origin.git")
	if runtime.GOOS != "windows" && !spaced {
		t.Fatalf("final-space name not distinct on %s: %s", runtime.GOOS, detail)
	}
	t.Logf("final-space name distinct for git on %s: %v %s", runtime.GOOS, spaced, detail)
}

func hostFoldsCase(t *testing.T) bool {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fold-probe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := os.Stat(filepath.Join(dir, "FOLD-PROBE"))
	return err == nil
}

// recordingTB records Skipf and Fatalf instead of ending the test, so the
// helper's verdict can be asserted.
type recordingTB struct {
	testing.TB
	skipped, failed string
}

func (r *recordingTB) Helper()                           {}
func (r *recordingTB) Skipf(format string, args ...any)  { r.skipped = fmt.Sprintf(format, args...) }
func (r *recordingTB) Fatalf(format string, args ...any) { r.failed = fmt.Sprintf(format, args...) }

// Claim: RequireGitRepositoryNames never skips a row of portable names (an
// ordinary positive control fails instead), skips a row with a significant
// byte that the probe rejects, and that skip reason is one line accepted by
// an allow-policy host-capability class of .github/ci/skip-classes.tsv, so
// CI's platform-case gate records it rather than failing on it. The planted
// collision is a duplicated name, which every platform rejects. Control:
// distinct names neither skip nor fail.
func TestRequireGitRepositoryNamesSkipsOnlySignificantRows(t *testing.T) {
	control := &recordingTB{TB: t}
	RequireGitRepositoryNames(control, t.TempDir(), "origin.git", "other.git")
	if control.skipped != "" || control.failed != "" {
		t.Fatalf("distinct names: skipped %q, failed %q", control.skipped, control.failed)
	}
	ordinary := &recordingTB{TB: t}
	RequireGitRepositoryNames(ordinary, t.TempDir(), "same.git", "same.git")
	if ordinary.skipped != "" || !strings.HasPrefix(ordinary.failed, "ordinary positive control: ") {
		t.Fatalf("portable collision: skipped %q, failed %q; want a failure, never a skip", ordinary.skipped, ordinary.failed)
	}
	significant := &recordingTB{TB: t}
	RequireGitRepositoryNames(significant, t.TempDir(), "same.git ", "same.git ")
	if significant.failed != "" || significant.skipped == "" || strings.ContainsAny(significant.skipped, "\r\n") {
		t.Fatalf("significant-byte collision: skipped %q, failed %q; want a one-line skip", significant.skipped, significant.failed)
	}
	if !hostCapabilityAllows(t, significant.skipped) {
		t.Fatalf("skip reason %q matches no allow-policy host-capability class", significant.skipped)
	}
}

// hostCapabilityAllows reports whether reason matches an allow-policy
// host-capability pattern of the CI skip-class table.
func hostCapabilityAllows(t *testing.T, reason string) bool {
	t.Helper()
	table, err := os.ReadFile(filepath.Join("..", "..", ".github", "ci", "skip-classes.tsv"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(table), "\n") {
		fields := strings.Split(line, "\t")
		if strings.HasPrefix(line, "#") || len(fields) < 3 || fields[0] != "host-capability" || fields[2] != "allow" {
			continue
		}
		if pattern, err := regexp.Compile(fields[1]); err == nil && pattern.MatchString(reason) {
			return true
		}
	}
	return false
}
