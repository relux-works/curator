package scriptworker

import (
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/godriver"
)

// TestMain gives the test binary the same fixed hidden script-worker mode
// the installed manager has, so every worker test launches a real
// identity-verified process instead of an in-process mock.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == WorkerMode {
		os.Exit(RunWorker(os.Stdin, os.Stdout))
	}
	if len(os.Args) == 2 && os.Args[1] == ScriptNetNSProbeMode {
		os.Exit(RunNetNSProbe())
	}
	if len(os.Args) == 3 && os.Args[1] == ScriptLandlockProbeMode {
		os.Exit(RunLandlockProbe(os.Args[2]))
	}
	code := m.Run()
	if stubOnce.directory != "" {
		_ = os.RemoveAll(stubOnce.directory)
	}
	if curatorOnce.directory != "" {
		_ = os.RemoveAll(curatorOnce.directory)
	}
	if forgeOnce.directory != "" {
		_ = os.RemoveAll(forgeOnce.directory)
	}
	os.Exit(code)
}

var stubOnce struct {
	sync.Once
	directory string
	path      string
	err       error
}

// stubInterpreterBinary compiles testdata/stubinterp once per package test
// run with the real toolchain, so the fake interpreter is a real native
// executable with a stable identity.
func stubInterpreterBinary(t *testing.T) string {
	t.Helper()
	stubOnce.Do(func() {
		directory, err := os.MkdirTemp("", "curator-stubinterp-")
		if err != nil {
			stubOnce.err = err
			return
		}
		stubOnce.directory = directory
		output := filepath.Join(directory, "stubinterp")
		if runtime.GOOS == "windows" {
			output += ".exe"
		}
		command := exec.Command(filepath.Join(build.Default.GOROOT, "bin", goBinaryName()), "build", "-o", output, "./testdata/stubinterp") // #nosec G204 -- fixed test-only build of the local test double
		command.Env = os.Environ()
		if combined, err := command.CombinedOutput(); err != nil {
			stubOnce.err = err
			t.Logf("stubinterp build output: %s", combined)
			return
		}
		stubOnce.path = output
	})
	if stubOnce.err != nil {
		t.Fatalf("cannot build the stub interpreter: %v", stubOnce.err)
	}
	return stubOnce.path
}

func goBinaryName() string {
	if runtime.GOOS == "windows" {
		return "go.exe"
	}
	return "go"
}

var forgeOnce struct {
	sync.Once
	directory string
	path      string
	err       error
}

// forgeWorkerBinary compiles testdata/forgeworker once per package test
// run: a helper that speaks the session framing and lies in its ready
// proof or result, so tests prove the parent validates before permitting.
func forgeWorkerBinary(t *testing.T) string {
	t.Helper()
	forgeOnce.Do(func() {
		directory, err := os.MkdirTemp("", "curator-forgeworker-")
		if err != nil {
			forgeOnce.err = err
			return
		}
		forgeOnce.directory = directory
		output := filepath.Join(directory, "forgeworker")
		if runtime.GOOS == "windows" {
			output += ".exe"
		}
		command := exec.Command(filepath.Join(build.Default.GOROOT, "bin", goBinaryName()), "build", "-o", output, "./testdata/forgeworker") // #nosec G204 -- fixed test-only build of the local test double
		command.Env = os.Environ()
		if combined, err := command.CombinedOutput(); err != nil {
			forgeOnce.err = err
			t.Logf("forgeworker build output: %s", combined)
			return
		}
		forgeOnce.path = output
	})
	if forgeOnce.err != nil {
		t.Fatalf("cannot build the forge worker: %v", forgeOnce.err)
	}
	return forgeOnce.path
}

// executableFixtureName appends the platform executable suffix so a copied
// manager binary stays launchable on Windows, mirroring the go-v1 worker
// rows (internal/godriver/identity_test.go launcher fixtures). Without the
// suffix the copy is not executable there and the launch fails before the
// worker can refuse with its own diagnostic.
func executableFixtureName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

var curatorOnce struct {
	sync.Once
	directory string
	path      string
	err       error
}

// builtCuratorBinary compiles the real ./cmd/curator once per package test
// run, so the production dispatch wiring is proven against the shipped
// binary rather than assumed from shared code.
func builtCuratorBinary(t *testing.T) string {
	t.Helper()
	curatorOnce.Do(func() {
		directory, err := os.MkdirTemp("", "curator-scriptworker-prod-")
		if err != nil {
			curatorOnce.err = err
			return
		}
		curatorOnce.directory = directory
		output := filepath.Join(directory, "curator")
		if runtime.GOOS == "windows" {
			output += ".exe"
		}
		command := exec.Command(filepath.Join(build.Default.GOROOT, "bin", goBinaryName()), "build", "-o", output, "../../cmd/curator") // #nosec G204 -- fixed test-only build of the production binary under test
		command.Env = os.Environ()
		if combined, err := command.CombinedOutput(); err != nil {
			curatorOnce.err = err
			t.Logf("curator build output: %s", combined)
			return
		}
		curatorOnce.path = output
	})
	if curatorOnce.err != nil {
		t.Fatalf("cannot build the production curator binary: %v", curatorOnce.err)
	}
	return curatorOnce.path
}

func writeTestFile(t *testing.T, path string, payload []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, mode); err != nil {
		t.Fatal(err)
	}
}

func copyTestFile(t *testing.T, source, destination string) {
	t.Helper()
	payload, err := os.ReadFile(source) // #nosec G304 -- test-owned source path
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(source)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, destination, payload, info.Mode().Perm())
}

// mustPhysical resolves path the way the launch boundary does, so fixtures
// compare canonical paths on hosts where TempDir itself travels through a
// link (macOS /var -> /private/var, Windows junctions).
func mustPhysical(t *testing.T, path string) string {
	t.Helper()
	physical, err := godriver.CanonicalPhysicalPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return physical
}
