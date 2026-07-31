//go:build windows

package godriver

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// A Windows directory junction is the redirection the boundary meets on a real
// host: actions/setup-go leaves the runner's GOROOT as one, and Go reports it
// as fs.ModeIrregular without fs.ModeDir while every open and traversal through
// it succeeds. filepath.EvalSymlinks follows fs.ModeSymlink only, so it hands
// the junction back unresolved, and the directory check that follows is then
// asked about the redirection rather than the directory it names. These cases
// pin that the selection resolves the redirection instead, and that everything
// it freezes afterwards belongs to the physical directory.

func TestJunctionedGOROOTSelectsThePhysicalToolchain(t *testing.T) {
	root := testToolchain(t)
	physicalRoot := mustPhysical(t, root)
	junction := createDirectoryJunction(t, filepath.Join(t.TempDir(), "x64"), root)
	assertJunctionPlatformFact(t, junction)

	for name, config := range map[string]Config{
		"GOROOT":     {GOROOT: junction},
		"CURATOR_GO": {CuratorGo: filepath.Join(junction, "bin", platformGoName)},
	} {
		t.Run(name, func(t *testing.T) {
			executable, goroot, rootInfo, err := selectToolchain(config, junction)
			if err != nil {
				t.Fatalf("selectToolchain through a junction: %v", err)
			}
			if goroot != physicalRoot {
				t.Fatalf("selected GOROOT = %q, want the physical root %q", goroot, physicalRoot)
			}
			if want := filepath.Join(physicalRoot, "bin", platformGoName); executable != want {
				t.Fatalf("selected launcher = %q, want %q", executable, want)
			}
			// The frozen root identity has to be the directory's, so a later
			// re-aiming of the junction cannot pass the drift check by leaving
			// the junction's own identity untouched.
			info, err := os.Lstat(goroot)
			if err != nil || !info.IsDir() || info.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 {
				t.Fatalf("os.Lstat(%s) = %v, %v, want a plain directory", goroot, info, err)
			}
			if !os.SameFile(rootInfo, info) {
				t.Fatal("the frozen root identity is not the physical directory's")
			}
			if err := verifySelectedRoot(goroot, rootInfo, executable); err != nil {
				t.Fatalf("verifySelectedRoot: %v", err)
			}
		})
	}
}

// TestJunctionedRuntimeGOROOTProbesAsOnePhysicalToolchain runs the whole
// establishment the way the runner presents it: build.Default.GOROOT is the
// junction, so the fallback selection starts there, and the probe reports its
// GOROOT in the junction spelling. Both the private probe base and the selected
// root are reached through junctions, so a resolver that only handled one of
// them would still fail here.
func TestJunctionedRuntimeGOROOTProbesAsOnePhysicalToolchain(t *testing.T) {
	root := testToolchain(t)
	physicalRoot := mustPhysical(t, root)
	junction := createDirectoryJunction(t, filepath.Join(t.TempDir(), "x64"), root)
	assertJunctionPlatformFact(t, junction)
	base := createDirectoryJunction(t, filepath.Join(t.TempDir(), "base"), t.TempDir())
	assertJunctionPlatformFact(t, base)

	host := hostFacts{runtimeGOROOT: junction, goos: runtime.GOOS, goarch: runtime.GOARCH}
	executor := validProbeExecutor(t, junction, host)
	session, err := establish(context.Background(), Config{PrivateBase: base, Executor: executor}, host)
	if err != nil {
		t.Fatalf("establish through a junctioned GOROOT: %v", err)
	}
	if session.GOROOT() != physicalRoot {
		t.Fatalf("session GOROOT = %q, want the physical root %q", session.GOROOT(), physicalRoot)
	}
	if want := filepath.Join(physicalRoot, "bin", platformGoName); session.Executable() != want {
		t.Fatalf("session launcher = %q, want %q", session.Executable(), want)
	}
	// The worker re-proves the launcher path independently before every exec,
	// so the selection and that re-proof have to agree on the spelling.
	if err := verifyGoLauncher(session.Executable()); err != nil {
		t.Fatalf("verifyGoLauncher on the selected launcher: %v", err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestOrdinaryPathSpellingRejectsANameItCannotJoinOnto(t *testing.T) {
	cases := []struct {
		name  string
		final string
		want  string
	}{
		{name: "drive", final: `\\?\C:\Go`, want: `C:\Go`},
		{name: "UNC", final: `\\?\UNC\server\share\Go`, want: `\\server\share\Go`},
		{name: "already ordinary", final: `C:\Go`, want: `C:\Go`},
		{name: "volume GUID", final: `\\?\Volume{11111111-2222-3333-4444-555555555555}\Go`, want: ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			ordinary, err := ordinaryPathSpelling(`C:\selected`, test.final)
			if test.want == "" {
				if err == nil {
					t.Fatalf("ordinaryPathSpelling(%q) = %q, want a rejection", test.final, ordinary)
				}
				return
			}
			if err != nil || ordinary != test.want {
				t.Fatalf("ordinaryPathSpelling(%q) = %q, %v, want %q", test.final, ordinary, err, test.want)
			}
		})
	}
}
