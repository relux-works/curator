package main

import (
	"context"
	"fmt"
	"go/build"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// TestHostGoToolchainIsSelectableOnAnInventoryPlatform pins the one precondition
// every compiled-command case in this package depends on and none of them state:
// on a platform rc5-native-control-inventory-v1 covers, the go-v1 boundary must
// be able to select the host's own Go installation from the trusted selection
// variables alone.
//
// Without it, each of those cases fails on its own setup install with the same
// `outcome=toolchain-unavailable` line and no indication of which host fact the
// boundary rejected, because the driver deliberately reports a verdict rather
// than a path. This case reports the selection inputs instead, so a host whose
// GOROOT the boundary cannot accept is diagnosed once, here, rather than seven
// times in cases that are about status reporting.
func TestHostGoToolchainIsSelectableOnAnInventoryPlatform(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	snapshot, err := godriver.Probe(context.Background(), godriver.ConfigFromEnvironment(t.TempDir()))
	if err != nil {
		t.Fatalf("the go-v1 boundary could not select the host Go installation: %v\n%s", err, hostToolchainFacts())
	}
	if snapshot.GOROOT == "" || snapshot.Executable == "" {
		t.Fatalf("selected toolchain snapshot is incomplete: %+v", snapshot)
	}
	if snapshot.Target.GOOS != runtime.GOOS || snapshot.Target.GOARCH != runtime.GOARCH {
		t.Fatalf("selected toolchain targets %s/%s, want the host %s/%s",
			snapshot.Target.GOOS, snapshot.Target.GOARCH, runtime.GOOS, runtime.GOARCH)
	}
}

// hostToolchainFacts renders the selection inputs the go-v1 boundary reads,
// each resolved the same way it resolves them. It is only ever printed on a
// failure, and it exists so that failure names the rejected host fact.
func hostToolchainFacts() string {
	var report strings.Builder
	fmt.Fprintf(&report, "host toolchain facts (%s/%s):\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&report, "  %s=%q\n", godriver.SelectionCuratorGo, os.Getenv(godriver.SelectionCuratorGo))
	fmt.Fprintf(&report, "  %s=%q\n", godriver.SelectionGOROOT, os.Getenv(godriver.SelectionGOROOT))
	fmt.Fprintf(&report, "  build.Default.GOROOT=%q\n", build.Default.GOROOT)

	root := os.Getenv(godriver.SelectionGOROOT)
	if root == "" {
		root = build.Default.GOROOT
	}
	describePathFact(&report, "selected GOROOT", root)
	resolved, err := filepath.EvalSymlinks(filepath.Clean(root))
	fmt.Fprintf(&report, "  EvalSymlinks(GOROOT)=%q err=%v\n", resolved, err)
	if err == nil {
		describePathFact(&report, "resolved GOROOT", resolved)
		launcher := "go"
		if runtime.GOOS == "windows" {
			launcher = "go.exe"
		}
		describePathFact(&report, "derived launcher", filepath.Join(resolved, "bin", launcher))
	}
	return report.String()
}

func describePathFact(report *strings.Builder, label, path string) {
	info, err := os.Lstat(path)
	if err != nil {
		fmt.Fprintf(report, "  %s %q: lstat err=%v%s\n", label, path, err, platformPathFact(path))
		return
	}
	followed, followErr := os.Stat(path)
	fmt.Fprintf(report, "  %s %q: lstatDir=%t regular=%t symlink=%t mode=%v size=%d statDir=%t statErr=%v%s\n",
		label, path, info.IsDir(), info.Mode().IsRegular(),
		info.Mode()&fs.ModeSymlink != 0, info.Mode(), info.Size(),
		followErr == nil && followed.IsDir(), followErr, platformPathFact(path))
}
