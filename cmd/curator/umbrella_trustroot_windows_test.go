//go:build windows

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

// shortPathName returns the 8.3 short spelling of dir through the
// platform API. Volumes without short-name generation answer with the
// long spelling itself.
func shortPathName(t *testing.T, dir string) string {
	t.Helper()
	long16, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		t.Fatal(err)
	}
	buffer := make([]uint16, 1024)
	n, err := windows.GetShortPathName(long16, &buffer[0], uint32(len(buffer)))
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 || n > uint32(len(buffer)) {
		t.Fatalf("GetShortPathName(%q) answered %d", dir, n)
	}
	return windows.UTF16ToString(buffer[:n])
}

// sameProviderFile reports whether two provider spellings name one file,
// mirroring the SameFile identity production directory comparisons use.
func sameProviderFile(t *testing.T, a, b string) bool {
	t.Helper()
	aInfo, aErr := os.Stat(a)
	bInfo, bErr := os.Stat(b)
	if aErr != nil || bErr != nil {
		t.Fatalf("stat providers %q (%v) and %q (%v)", a, aErr, b, bErr)
	}
	return os.SameFile(aInfo, bInfo)
}

// TestUmbrellaTrustRootShortSpellingIdentity pins spelling identity for
// the §11 verdicts on Windows: a trust root in 8.3 short spelling and a
// PATH entry in long spelling name one directory, in both directions,
// and a managed root in short spelling still refuses its providers.
func TestUmbrellaTrustRootShortSpellingIdentity(t *testing.T) {
	tmp := t.TempDir()
	long, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		t.Fatal(err)
	}
	short := shortPathName(t, tmp)
	if strings.EqualFold(short, long) {
		t.Skip("this host cannot create 8.3 short-name aliases for temp paths")
	}
	install := filepath.Join(tmp, "install")
	if err := os.MkdirAll(install, 0o755); err != nil {
		t.Fatal(err)
	}
	silent := writeProviderFile(t, tmp, "curator-run", true)
	newInputs := func(roots, entries []string) providerInputs {
		return providerInputs{
			installDir:    install,
			providerDirs:  roots,
			publishedDirs: nil,
			managedRoots:  nil,
			pathEntries:   entries,
			platform:      runtime.GOOS,
			pathExt:       append([]string{}, defaultWindowsPathExt...),
			revision:      providerRevisionA,
		}
	}
	// Trust root short, PATH entry long: one directory, silent, and the
	// resolved provider is the planted file — compared by filesystem
	// identity because the two spellings differ beyond case.
	if outcome := resolveProvider("run", newInputs([]string{short}, []string{long})); outcome.resolved == "" || outcome.warned || outcome.diagnostic != "" {
		t.Fatalf("short root, long PATH outcome = %+v, want a silent selection", outcome)
	} else if !sameProviderFile(t, outcome.resolved, silent) {
		t.Fatalf("resolved %q, want the planted provider %q", outcome.resolved, silent)
	}
	// Trust root long, PATH entry short: one directory, silent.
	if outcome := resolveProvider("run", newInputs([]string{long}, []string{short})); outcome.resolved == "" || outcome.warned || outcome.diagnostic != "" {
		t.Fatalf("long root, short PATH outcome = %+v, want a silent selection", outcome)
	} else if !sameProviderFile(t, outcome.resolved, silent) {
		t.Fatalf("resolved %q, want the planted provider %q", outcome.resolved, silent)
	}
	// A managed root in short spelling still refuses its providers.
	managed := newInputs(nil, []string{long})
	managed.managedRoots = []labeledDir{{path: short, label: "a managed directory"}}
	if outcome := resolveProvider("run", managed); outcome.resolved != "" || outcome.diagnostic != providerDiagnosticUntrusted {
		t.Fatalf("short managed root outcome = %+v, want an untrusted refusal", outcome)
	}
}
