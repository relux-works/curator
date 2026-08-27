package godriver

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type recordingExecutor struct {
	calls []Process
	run   func(int, Process) (Output, error)
}

func (executor *recordingExecutor) Run(_ context.Context, process Process) (Output, error) {
	process.Arguments = append([]string(nil), process.Arguments...)
	process.Environment = append([]string(nil), process.Environment...)
	executor.calls = append(executor.calls, process)
	return executor.run(len(executor.calls)-1, process)
}

func TestEstablishUsesExactPrivateProbeAndFrozenEnvironment(t *testing.T) {
	root := testToolchain(t)
	physicalRoot := mustPhysical(t, root)
	base := t.TempDir()
	host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
	executor := validProbeExecutor(t, root, host)

	session, err := establish(context.Background(), Config{PrivateBase: base, CuratorGo: filepath.Join(root, "bin", platformGoName), Executor: executor}, host)
	if err != nil {
		t.Fatal(err)
	}
	operation := session.operation
	defer session.Close()

	wantArguments := [][]string{
		{"telemetry", "off"},
		{"version"},
		append([]string{"env", "-json"}, probeEnvNames...),
	}
	if len(executor.calls) != len(wantArguments) {
		t.Fatalf("calls = %d, want %d", len(executor.calls), len(wantArguments))
	}
	for index, call := range executor.calls {
		if call.Executable != filepath.Join(physicalRoot, "bin", platformGoName) || !reflect.DeepEqual(call.Arguments, wantArguments[index]) {
			t.Fatalf("call %d = %q %q", index, call.Executable, call.Arguments)
		}
		if call.Directory != filepath.Join(operation, "empty") {
			t.Fatalf("call %d cwd = %q", index, call.Directory)
		}
		if call.Timeout <= 0 || call.OutputLimit != defaultOutputLimit {
			t.Fatalf("call %d bounds = %v/%d", index, call.Timeout, call.OutputLimit)
		}
		probeEnv := environmentMap(call.Environment)
		for _, forbidden := range []string{"GOROOT", "GOOS", "GOARCH", "GOFLAGS", "GOWORK", "GOPROXY", "GOPRIVATE", "CC", "CXX", "SSH_AUTH_SOCK", "HTTP_PROXY"} {
			if _, present := probeEnv[forbidden]; present {
				t.Fatalf("bootstrap call %d inherited %s", index, forbidden)
			}
		}
		for _, fixed := range []string{"GOENV", "GOTOOLCHAIN", "GOPATH", "GOMODCACHE", "GOCACHE", "GOTMPDIR", "HOME", "XDG_CONFIG_HOME", "PATH", "TMPDIR", "LC_ALL", "LANG"} {
			value, present := probeEnv[fixed]
			if !present {
				t.Fatalf("bootstrap call %d missing %s", index, fixed)
			}
			if fixed != "LC_ALL" && fixed != "LANG" && fixed != "GOENV" && fixed != "GOTOOLCHAIN" && !isWithin(value, operation) {
				t.Fatalf("bootstrap %s = %q, outside operation", fixed, value)
			}
		}
	}

	target := session.Target()
	if target.GOOS != host.goos || target.GOARCH != host.goarch {
		t.Fatalf("target = %+v", target)
	}
	if key := tuningVariable(host.goarch); key != "" && target.Tuning[key] == "" {
		t.Fatalf("target tuning = %+v", target.Tuning)
	}
	finalEnv := environmentMap(session.Environment())
	for key, want := range map[string]string{
		"GOROOT": physicalRoot, "GOOS": host.goos, "GOARCH": host.goarch,
		"GO111MODULE": "on", "GOFLAGS": "", "GOPROXY": "off", "GOSUMDB": "off",
		"GOPRIVATE": "", "GONOPROXY": "none", "GONOSUMDB": "none", "GOVCS": "*:off",
		"GOWORK": "off", "CGO_ENABLED": "0", "GO_EXTLINK_ENABLED": "0", "GOEXPERIMENT": "",
	} {
		if finalEnv[key] != want {
			t.Errorf("environment %s = %q, want %q", key, finalEnv[key], want)
		}
	}
	for _, forbidden := range []string{"CC", "CXX", "AR", "PKG_CONFIG", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "SSH_AUTH_SOCK", "GIT_CONFIG_GLOBAL", "LANGUAGE"} {
		if _, present := finalEnv[forbidden]; present {
			t.Errorf("operation environment inherited %s", forbidden)
		}
	}
	if session.Toolchain().Algorithm != buildmetaAlgorithm || session.Toolchain().GoRelpath != "bin/go" || !strings.HasPrefix(session.Toolchain().ContentSHA256, "sha256:") {
		t.Fatalf("toolchain = %+v", session.Toolchain())
	}

	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(operation); !os.IsNotExist(err) {
		t.Fatalf("private operation still exists: %v", err)
	}
}

const buildmetaAlgorithm = "curator-go-toolchain-v1"

func TestSelectionPriorityAndNoPathLookup(t *testing.T) {
	selected := testToolchain(t)
	physicalSelected := mustPhysical(t, selected)
	ignored := testToolchain(t)
	base := t.TempDir()
	host := hostFacts{runtimeGOROOT: ignored, goos: runtime.GOOS, goarch: runtime.GOARCH}
	executor := validProbeExecutor(t, selected, host)
	config := Config{
		PrivateBase: base, CuratorGo: filepath.Join(selected, "bin", platformGoName), GOROOT: ignored,
		Executor: executor,
	}
	session, err := establish(context.Background(), config, host)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if session.GOROOT() != physicalSelected {
		t.Fatalf("GOROOT = %q, want CURATOR_GO root %q", session.GOROOT(), physicalSelected)
	}
	for _, call := range executor.calls {
		if call.Executable != filepath.Join(physicalSelected, "bin", platformGoName) {
			t.Fatalf("executed unselected path %q", call.Executable)
		}
	}
}

func TestConfigFromEnvironmentAdmitsOnlyTrustedSelectors(t *testing.T) {
	t.Setenv("CURATOR_GO", "/trusted/bin/"+platformGoName)
	t.Setenv("GOROOT", "/trusted")
	t.Setenv("PATH", "/repository/fake-bin")
	t.Setenv("GOFLAGS", "-toolexec=evil")
	t.Setenv("GOWORK", "/repository/go.work")
	config := ConfigFromEnvironment("/private", "/repository")
	if config.CuratorGo != "/trusted/bin/"+platformGoName || config.GOROOT != "/trusted" || !reflect.DeepEqual(config.ForbiddenRoots, []string{"/repository"}) {
		t.Fatalf("config = %+v", config)
	}
}

func TestSelectionFallsBackOnlyWhenHigherPriorityIsMissing(t *testing.T) {
	root := testToolchain(t)
	physicalRoot := mustPhysical(t, root)
	host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
	for name, config := range map[string]Config{
		"GOROOT":  {PrivateBase: t.TempDir(), GOROOT: root},
		"runtime": {PrivateBase: t.TempDir()},
	} {
		t.Run(name, func(t *testing.T) {
			executor := validProbeExecutor(t, root, host)
			config.Executor = executor
			session, err := establish(context.Background(), config, host)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			if session.GOROOT() != physicalRoot {
				t.Fatalf("GOROOT = %q", session.GOROOT())
			}
		})
	}
}

func TestUnsafeToolchainCandidatesFailBeforeExecution(t *testing.T) {
	validRoot := testToolchain(t)
	wrapperRoot := t.TempDir()
	writeTestFile(t, filepath.Join(wrapperRoot, "bin", platformGoName), []byte("#!/bin/sh\nexec go \"$@\"\n"), 0o755)
	repository := t.TempDir()
	repositoryRoot := filepath.Join(repository, "toolchain")
	copyTestToolchain(t, validRoot, repositoryRoot)

	cases := []struct {
		name   string
		config Config
		code   string
	}{
		{name: "relative CURATOR_GO", config: Config{CuratorGo: filepath.Join("bin", platformGoName)}, code: "untrusted_go_executable"},
		{name: "wrapper", config: Config{CuratorGo: filepath.Join(wrapperRoot, "bin", platformGoName)}, code: "untrusted_go_executable"},
		{name: "repository toolchain", config: Config{CuratorGo: filepath.Join(repositoryRoot, "bin", platformGoName), ForbiddenRoots: []string{repository}}, code: "untrusted_go_executable"},
		{name: "missing GOROOT", config: Config{GOROOT: filepath.Join(t.TempDir(), "missing")}, code: "go_toolchain_missing"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			executor := &recordingExecutor{run: func(int, Process) (Output, error) { return Output{}, nil }}
			test.config.PrivateBase = t.TempDir()
			test.config.Executor = executor
			_, err := establish(context.Background(), test.config, hostFacts{runtimeGOROOT: validRoot, goos: runtime.GOOS, goarch: runtime.GOARCH})
			if DiagnosticCode(err) != test.code {
				t.Fatalf("error = %v, code = %q, want %q", err, DiagnosticCode(err), test.code)
			}
			if len(executor.calls) != 0 {
				t.Fatalf("unsafe candidate executed %d times", len(executor.calls))
			}
		})
	}
}

func TestCandidateLinkOutsideDerivedGOROOTFails(t *testing.T) {
	inside := t.TempDir()
	outside := testToolchain(t)
	if err := os.MkdirAll(filepath.Join(inside, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "bin", platformGoName), filepath.Join(inside, "bin", platformGoName)); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, _, _, err := selectToolchain(Config{CuratorGo: filepath.Join(inside, "bin", platformGoName)}, outside)
	if DiagnosticCode(err) != "toolchain_executable_mismatch" {
		t.Fatalf("error = %v", err)
	}
}

func TestProbeFailuresCleanPrivateStateAndHaveStableDiagnostics(t *testing.T) {
	root := testToolchain(t)
	otherRoot := testToolchain(t)
	host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
	cases := []struct {
		name   string
		mutate func(int, Process, Output) (Output, error)
		code   string
	}{
		{name: "telemetry failure", code: "telemetry_initialization_failed", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 0 {
				return output, errors.New("exit 2")
			}
			return output, nil
		}},
		{name: "probe timeout", code: "process_timeout", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 0 {
				return output, errProcessTimeout
			}
			return output, nil
		}},
		{name: "invalid version", code: "malformed_go_version", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 1 {
				output.Stdout = []byte("not go\n")
			}
			return output, nil
		}},
		{name: "future family", code: "unsupported_go_family", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 1 {
				output.Stdout = []byte("go version go1.26.1 " + host.goos + "/" + host.goarch + "\n")
			}
			return output, nil
		}},
		{name: "target mismatch", code: "target_mismatch", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 2 {
				output.Stdout = mutateJSON(t, output.Stdout, "GOOS", "other")
			}
			return output, nil
		}},
		{name: "GOROOT mismatch", code: "toolchain_executable_mismatch", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 2 {
				output.Stdout = mutateJSON(t, output.Stdout, "GOROOT", otherRoot)
			}
			return output, nil
		}},
		{name: "telemetry remains local", code: "telemetry_initialization_failed", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 2 {
				output.Stdout = mutateJSON(t, output.Stdout, "GOTELEMETRY", "local")
			}
			return output, nil
		}},
		{name: "invalid env JSON", code: "invalid_go_env", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 2 {
				output.Stdout = []byte(`{"GOROOT":`)
			}
			return output, nil
		}},
		{name: "telemetry escape", code: "telemetry_directory_untrusted", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 2 {
				output.Stdout = mutateJSON(t, output.Stdout, "GOTELEMETRYDIR", t.TempDir())
			}
			return output, nil
		}},
		{name: "bounded output", code: "process_output_limit", mutate: func(index int, _ Process, output Output) (Output, error) {
			if index == 0 {
				output.Stdout = []byte(strings.Repeat("x", 65))
			}
			return output, nil
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			base := t.TempDir()
			valid := validProbeExecutor(t, root, host)
			executor := &recordingExecutor{run: func(index int, process Process) (Output, error) {
				output, err := valid.run(index, process)
				if err != nil {
					return output, err
				}
				return test.mutate(index, process, output)
			}}
			config := Config{PrivateBase: base, CuratorGo: filepath.Join(root, "bin", platformGoName), Executor: executor}
			if test.name == "bounded output" {
				config.OutputLimit = 64
			}
			_, err := establish(context.Background(), config, host)
			if DiagnosticCode(err) != test.code {
				t.Fatalf("error = %v, code = %q, want %q", err, DiagnosticCode(err), test.code)
			}
			assertDirectoryEmpty(t, base)
		})
	}
}

func TestMalformedVersionAndEnvironmentInputs(t *testing.T) {
	for _, payload := range [][]byte{
		[]byte("go version go1.25.5 " + runtime.GOOS + "/" + runtime.GOARCH),
		[]byte("go version go1.25.5 " + runtime.GOOS + "/" + runtime.GOARCH + "\n\n"),
		[]byte("go version go1.25.5 " + runtime.GOOS + "/" + runtime.GOARCH + "\rX\n"),
		append([]byte("go version go1.25.5 "), 0xff, '\n'),
	} {
		if _, _, _, _, err := parseGoVersion(payload); DiagnosticCode(err) != "malformed_go_version" {
			t.Errorf("parseGoVersion(%q) = %v", payload, err)
		}
	}
	if _, _, _, _, err := parseGoVersion([]byte("go version go1.22.12 " + runtime.GOOS + "/" + runtime.GOARCH + "\n")); DiagnosticCode(err) != "unsupported_go_family" {
		t.Fatalf("pre-1.23 error = %v", err)
	}

	duplicate := `{"GOROOT":"a","GOROOT":"b"}`
	if _, err := decodeProbeEnvironment([]byte(duplicate)); DiagnosticCode(err) != "invalid_go_env" {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestProbeRemovesStateAndNeverMemoizes(t *testing.T) {
	root := testToolchain(t)
	base := t.TempDir()
	host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
	executor := validProbeExecutor(t, root, host)
	config := Config{PrivateBase: base, CuratorGo: filepath.Join(root, "bin", platformGoName), Executor: executor}
	for range 2 {
		snapshot, err := probe(context.Background(), config, host)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Toolchain.Algorithm != buildmetaAlgorithm {
			t.Fatalf("snapshot toolchain = %+v", snapshot.Toolchain)
		}
		assertDirectoryEmpty(t, base)
	}
	if len(executor.calls) != 6 {
		t.Fatalf("probe calls = %d, want 6 (no memo)", len(executor.calls))
	}
}

func TestToolchainMutationFailsRecheck(t *testing.T) {
	root := testToolchain(t)
	host := hostFacts{runtimeGOROOT: root, goos: runtime.GOOS, goarch: runtime.GOARCH}
	executor := validProbeExecutor(t, root, host)
	session, err := establish(context.Background(), Config{PrivateBase: t.TempDir(), CuratorGo: filepath.Join(root, "bin", platformGoName), Executor: executor}, host)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "VERSION"), []byte("mutated"), 0o644)
	if err := session.VerifyToolchain(context.Background()); DiagnosticCode(err) != "toolchain_mutated" {
		t.Fatalf("VerifyToolchain error = %v", err)
	}
	if err := session.Close(); DiagnosticCode(err) != "toolchain_mutated" {
		t.Fatalf("Close error = %v", err)
	}
	if _, err := os.Lstat(session.operation); !os.IsNotExist(err) {
		t.Fatalf("mutated session probe was not removed: %v", err)
	}
}

func TestRealTrustedGoProbe(t *testing.T) {
	if os.Getenv("CURATOR_REAL_GO_TEST") != "1" {
		t.Skip("set CURATOR_REAL_GO_TEST=1 for the bounded native toolchain probe")
	}
	config := ConfigFromEnvironment(t.TempDir())
	config.CuratorGo = filepath.Join(runtime.GOROOT(), "bin", platformGoName)
	session, err := Establish(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if session.Target().GOOS != runtime.GOOS || session.Target().GOARCH != runtime.GOARCH {
		t.Fatalf("target = %+v", session.Target())
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
}

func validProbeExecutor(t *testing.T, root string, host hostFacts) *recordingExecutor {
	t.Helper()
	return &recordingExecutor{run: func(index int, process Process) (Output, error) {
		switch index % 3 {
		case 0:
			return Output{}, nil
		case 1:
			return Output{Stdout: []byte("go version go1.25.5 " + host.goos + "/" + host.goarch + "\n")}, nil
		case 2:
			environment := environmentMap(process.Environment)
			telemetry := filepath.Join(environment["XDG_CONFIG_HOME"], "go", "telemetry")
			if err := os.MkdirAll(telemetry, 0o700); err != nil {
				t.Fatal(err)
			}
			values := map[string]string{
				"GOROOT": root, "GOHOSTOS": host.goos, "GOHOSTARCH": host.goarch, "GOOS": host.goos, "GOARCH": host.goarch,
				"GO386": "sse2", "GOAMD64": "v1", "GOARM": "7", "GOARM64": "v8.0", "GOMIPS": "hardfloat",
				"GOMIPS64": "hardfloat", "GOPPC64": "power8", "GORISCV64": "rva20u64", "GOWASM": "satconv,signext",
				"GOTELEMETRY": "off", "GOTELEMETRYDIR": telemetry,
			}
			payload, err := json.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			return Output{Stdout: payload}, nil
		default:
			panic("unreachable")
		}
	}}
}

func testToolchain(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "bin", platformGoName), testExecutableBytes(), 0o755)
	writeTestFile(t, filepath.Join(root, "VERSION"), []byte("go1.25.5\n"), 0o644)
	writeTestFile(t, filepath.Join(root, "pkg", "tool", runtime.GOOS+"_"+runtime.GOARCH, "compile"), testExecutableBytes(), 0o755)
	return root
}

func copyTestToolchain(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o755)
	})
	if err != nil {
		t.Fatal(err)
	}
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

func mutateJSON(t *testing.T, payload []byte, key, value string) []byte {
	t.Helper()
	var object map[string]string
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	object[key] = value
	result, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertDirectoryEmpty(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		sort.Strings(names)
		t.Fatalf("directory %s is not empty: %v", path, names)
	}
}

func mustPhysical(t *testing.T, path string) string {
	t.Helper()
	physical, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return physical
}
