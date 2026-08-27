package godriver

import (
	"bytes"
	"context"
	"encoding/json"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/buildsource"
)

// TestMain gives the test binary the same fixed hidden worker mode the
// installed manager has, so every worker test launches a real identity-verified
// process instead of an in-process mock.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == WorkerMode {
		os.Exit(RunWorker(os.Stdin, os.Stdout))
	}
	os.Exit(m.Run())
}

// stubScript mirrors the manager-owned script consumed by testdata/stubgo.
type stubScript struct {
	Version string `json:"version"`

	ListStdout   string `json:"list_stdout"`
	ListStderr   string `json:"list_stderr"`
	ListExit     int    `json:"list_exit"`
	ListSleepMS  int    `json:"list_sleep_ms"`
	ListPadBytes int    `json:"list_pad_bytes"`

	BuildStderr    string `json:"build_stderr"`
	BuildExit      int    `json:"build_exit"`
	BuildSleepMS   int    `json:"build_sleep_ms"`
	Artifact       string `json:"artifact"`
	ArtifactPad    int    `json:"artifact_pad"`
	ArtifactMode   string `json:"artifact_mode"`
	ExtraOutput    string `json:"extra_output"`
	HardlinkSource string `json:"hardlink_source"`

	MutateSourcePath    string `json:"mutate_source_path"`
	MutateSourceContent string `json:"mutate_source_content"`

	PrivateFileName  string `json:"private_file_name"`
	PrivateFileBytes int    `json:"private_file_bytes"`

	SpawnChildren int    `json:"spawn_children"`
	SpawnPidFile  string `json:"spawn_pid_file"`
	SpawnSeconds  int    `json:"spawn_seconds"`
}

// stubCall is one recorded invocation of the stub Go launcher.
type stubCall struct {
	Argv        []string `json:"argv"`
	Dir         string   `json:"dir"`
	Environment []string `json:"env"`
}

var stubOnce struct {
	sync.Once
	path string
	err  error
}

// stubGoBinary compiles testdata/stubgo once per package test run with the real
// toolchain, so the fake GOROOT contains a real native executable.
func stubGoBinary(t *testing.T) string {
	t.Helper()
	stubOnce.Do(func() {
		directory, err := os.MkdirTemp("", "curator-stubgo-")
		if err != nil {
			stubOnce.err = err
			return
		}
		output := filepath.Join(directory, "stubgo")
		if runtime.GOOS == "windows" {
			output += ".exe"
		}
		command := exec.Command(filepath.Join(build.Default.GOROOT, "bin", platformGoName), "build", "-o", output, "./testdata/stubgo") // #nosec G204 -- fixed test-only build of the local test double
		command.Env = os.Environ()
		if combined, err := command.CombinedOutput(); err != nil {
			stubOnce.err = err
			t.Logf("stubgo build output: %s", combined)
			return
		}
		stubOnce.path = output
	})
	if stubOnce.err != nil {
		t.Fatalf("cannot build the stub Go launcher: %v", stubOnce.err)
	}
	return stubOnce.path
}

// stubGOROOT assembles a fingerprintable fake GOROOT around the stub launcher.
func stubGOROOT(t *testing.T, script stubScript) string {
	t.Helper()
	root := t.TempDir()
	payload, err := os.ReadFile(stubGoBinary(t)) // #nosec G304 -- test-owned stub path
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "bin", platformGoName), payload, 0o755)
	writeTestFile(t, filepath.Join(root, "VERSION"), []byte("go1.25.5\n"), 0o644)
	writeTestFile(t, filepath.Join(root, "pkg", "tool", runtime.GOOS+"_"+runtime.GOARCH, "compile"), payload, 0o755)
	encoded, err := json.Marshal(script)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "curator-stub.json"), encoded, 0o644)
	return mustPhysical(t, root)
}

// workerFixture is one frozen snapshot plus a live session over the stub
// toolchain. Build runs the real worker, so every test observes the real
// process graph.
type workerFixture struct {
	t         *testing.T
	root      string
	buildRoot string
	sourceDir string
	goroot    string
	token     *buildsource.Token
	session   *Session
}

// newSnapshotFixture creates the frozen snapshot only, so a test can compute
// script values from its canonical paths before the session starts.
func newSnapshotFixture(t *testing.T) *workerFixture {
	t.Helper()
	snapshot := filepath.Join(t.TempDir(), "snapshot")
	buildRoot := filepath.Join(snapshot, "build")
	sourceDir := filepath.Join(buildRoot, "cmd", "tool")
	writeTestFile(t, filepath.Join(buildRoot, "go.mod"), []byte("module example.test/build\n\ngo 1.23\n"), 0o644)
	writeTestFile(t, filepath.Join(sourceDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)
	snapshot = mustPhysical(t, snapshot)
	return &workerFixture{
		t: t, root: snapshot,
		buildRoot: filepath.Join(snapshot, "build"),
		sourceDir: filepath.Join(snapshot, "build", "cmd", "tool"),
	}
}

// start freezes the snapshot and establishes the session over a stub GOROOT
// carrying script.
func (fixture *workerFixture) start(script stubScript) *workerFixture {
	fixture.t.Helper()
	token, err := buildsource.Validate(fixture.root)
	if err != nil {
		fixture.t.Fatal(err)
	}
	fixture.token = token
	fixture.goroot = stubGOROOT(fixture.t, script)
	session, err := Establish(context.Background(), Config{
		PrivateBase: fixture.t.TempDir(),
		CuratorGo:   filepath.Join(fixture.goroot, "bin", platformGoName),
	})
	if err != nil {
		fixture.t.Fatal(err)
	}
	fixture.session = session
	fixture.t.Cleanup(func() {
		_ = fixture.token.Close()
		_ = session.Close()
	})
	return fixture
}

func (fixture *workerFixture) request(limits ResourceLimits) BuildRequest {
	return BuildRequest{
		Session:       fixture.session,
		Source:        fixture.token,
		CommandObject: BuildCommand{"type": "build", "driver": "go-v1", "source_dir": "build/cmd/tool"},
		BuildRoot:     "build",
		SourceDir:     "build/cmd/tool",
		Command:       "golden-tool",
		Limits:        limits,
	}
}

func (fixture *workerFixture) rootPackage() packageJSON {
	return packageJSON{
		Dir: fixture.sourceDir, ImportPath: "example.test/build/cmd/tool", Name: "main", GoFiles: []string{"main.go"},
		Module: &moduleJSON{
			Path: "example.test/build", Main: true, Dir: fixture.buildRoot,
			GoMod: filepath.Join(fixture.buildRoot, "go.mod"), GoVersion: "1.23",
		},
	}
}

// calls returns every recorded stub launcher invocation, including the three
// package-independent bootstrap probes.
func (fixture *workerFixture) calls() []stubCall {
	fixture.t.Helper()
	cache := environmentMap(fixture.session.Environment())["GOCACHE"]
	payload, err := os.ReadFile(filepath.Join(cache, "curator-stub-calls.jsonl")) // #nosec G304 -- operation-private test path
	if err != nil {
		fixture.t.Fatalf("cannot read stub call log: %v", err)
	}
	var calls []stubCall
	decoder := json.NewDecoder(bytes.NewReader(payload))
	for decoder.More() {
		var call stubCall
		if err := decoder.Decode(&call); err != nil {
			fixture.t.Fatalf("cannot decode stub call log: %v", err)
		}
		calls = append(calls, call)
	}
	return calls
}

// sourceAwareCalls returns only the go list and go build invocations.
func (fixture *workerFixture) sourceAwareCalls() []stubCall {
	var result []stubCall
	for _, call := range fixture.calls() {
		if len(call.Argv) != 0 && (call.Argv[0] == "list" || call.Argv[0] == "build") {
			result = append(result, call)
		}
	}
	return result
}

func encodePackages(t *testing.T, packages ...packageJSON) []byte {
	t.Helper()
	var payload []byte
	for _, item := range packages {
		encoded, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		payload = append(payload, encoded...)
		payload = append(payload, '\n')
	}
	return payload
}
