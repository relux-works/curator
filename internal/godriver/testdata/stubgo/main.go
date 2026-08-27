// Command stubgo is a test double for the fingerprinted Go launcher. It is a
// real native executable, so the worker session, process graph, native
// controls, and artifact verification run against real processes rather than a
// mock. Its behavior is driven exclusively by a manager-owned script file
// inside the fake GOROOT, never by an argument or environment value that the
// driver does not already fix.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type script struct {
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

func main() { os.Exit(run(os.Args[1:])) }

func run(argv []string) int {
	if len(argv) >= 3 && argv[0] == "sleep" {
		seconds, _ := strconv.Atoi(argv[1])
		appendLine(argv[2], strconv.Itoa(os.Getpid()))
		time.Sleep(time.Duration(seconds) * time.Second)
		return 0
	}
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "stubgo: cannot resolve executable:", err)
		return 1
	}
	goroot := filepath.Dir(filepath.Dir(executable))
	loaded, err := load(filepath.Join(goroot, "curator-stub.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "stubgo: cannot read script:", err)
		return 1
	}
	record(argv)
	if len(argv) == 0 {
		return 2
	}
	switch argv[0] {
	case "telemetry":
		return 0
	case "version":
		fmt.Print(loaded.Version)
		return 0
	case "env":
		return environment(goroot)
	case "list":
		return list(loaded)
	case "build":
		return build(loaded, argv)
	default:
		fmt.Fprintln(os.Stderr, "stubgo: unsupported command", argv[0])
		return 2
	}
}

func load(path string) (script, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- manager-owned script inside the fake GOROOT
	if err != nil {
		return script{}, err
	}
	var loaded script
	if err := json.Unmarshal(payload, &loaded); err != nil {
		return script{}, err
	}
	if loaded.Version == "" {
		loaded.Version = "go version go1.25.5 " + runtime.GOOS + "/" + runtime.GOARCH + "\n"
	}
	return loaded, nil
}

// record appends one call record so tests can assert the exact argument vector,
// working directory, and environment the worker used.
func record(argv []string) {
	cache := os.Getenv("GOCACHE")
	if cache == "" {
		return
	}
	directory, _ := os.Getwd()
	entry := map[string]any{"argv": argv, "dir": directory, "env": os.Environ()}
	encoded, err := json.Marshal(entry)
	if err != nil {
		return
	}
	appendLine(filepath.Join(cache, "curator-stub-calls.jsonl"), string(encoded))
}

func appendLine(path, line string) {
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304 -- manager-owned private path
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(line + "\n")
}

func environment(goroot string) int {
	config := os.Getenv("XDG_CONFIG_HOME")
	telemetry := filepath.Join(config, "go", "telemetry")
	if err := os.MkdirAll(telemetry, 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "stubgo:", err)
		return 1
	}
	values := map[string]string{
		"GOROOT": goroot, "GOHOSTOS": runtime.GOOS, "GOHOSTARCH": runtime.GOARCH,
		"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH,
		"GO386": "sse2", "GOAMD64": "v1", "GOARM": "7", "GOARM64": "v8.0",
		"GOMIPS": "hardfloat", "GOMIPS64": "hardfloat", "GOPPC64": "power8",
		"GORISCV64": "rva20u64", "GOWASM": "satconv,signext",
		"GOTELEMETRY": "off", "GOTELEMETRYDIR": telemetry,
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stubgo:", err)
		return 1
	}
	_, _ = os.Stdout.Write(encoded)
	return 0
}

func list(loaded script) int {
	if loaded.MutateSourcePath != "" {
		_ = os.WriteFile(loaded.MutateSourcePath, []byte(loaded.MutateSourceContent), 0o600) // #nosec G306 -- deliberate mutation probe
	}
	if loaded.ListSleepMS > 0 {
		time.Sleep(time.Duration(loaded.ListSleepMS) * time.Millisecond)
	}
	_, _ = os.Stdout.WriteString(loaded.ListStdout)
	if loaded.ListPadBytes > 0 {
		_, _ = os.Stdout.Write(make([]byte, loaded.ListPadBytes))
	}
	if loaded.ListStderr != "" {
		_, _ = os.Stderr.WriteString(loaded.ListStderr)
	}
	return loaded.ListExit
}

func build(loaded script, argv []string) int {
	output := ""
	for index := 0; index+1 < len(argv); index++ {
		if argv[index] == "-o" {
			output = argv[index+1]
		}
	}
	if output == "" {
		fmt.Fprintln(os.Stderr, "stubgo: build without a manager-derived output")
		return 2
	}
	if loaded.BuildSleepMS > 0 {
		time.Sleep(time.Duration(loaded.BuildSleepMS) * time.Millisecond)
	}
	if loaded.PrivateFileBytes > 0 && loaded.PrivateFileName != "" {
		target := filepath.Join(os.Getenv("GOCACHE"), loaded.PrivateFileName)
		if err := os.WriteFile(target, make([]byte, loaded.PrivateFileBytes), 0o600); err != nil { // #nosec G306 -- deliberate private-state probe
			fmt.Fprintln(os.Stderr, "stubgo: private write failed:", err)
			return 1
		}
	}
	if loaded.SpawnChildren > 0 {
		executable, err := os.Executable()
		if err != nil {
			return 1
		}
		seconds := loaded.SpawnSeconds
		if seconds == 0 {
			seconds = 120
		}
		for index := 0; index < loaded.SpawnChildren; index++ {
			child := exec.Command(executable, "sleep", strconv.Itoa(seconds), loaded.SpawnPidFile) // #nosec G204 -- test double models a Go tool child
			child.Env = os.Environ()
			if err := child.Start(); err != nil {
				fmt.Fprintln(os.Stderr, "stubgo: cannot start tool child:", err)
				return 1
			}
		}
		// Let the children publish their identities before the build result
		// returns, so teardown is observed against live processes.
		time.Sleep(500 * time.Millisecond)
	}
	if loaded.BuildStderr != "" {
		_, _ = os.Stderr.WriteString(loaded.BuildStderr)
	}
	if loaded.BuildExit != 0 {
		return loaded.BuildExit
	}
	switch loaded.ArtifactMode {
	case "none":
	case "symlink":
		_ = os.Symlink(filepath.Join(filepath.Dir(output), "missing"), output)
	case "hardlink":
		if err := os.Link(loaded.HardlinkSource, output); err != nil {
			fmt.Fprintln(os.Stderr, "stubgo: cannot hard link:", err)
			return 1
		}
	default:
		payload := []byte(loaded.Artifact)
		if loaded.ArtifactPad > 0 {
			payload = append(payload, make([]byte, loaded.ArtifactPad)...)
		}
		if err := os.WriteFile(output, payload, 0o600); err != nil { // #nosec G306 -- manager applies final permissions
			fmt.Fprintln(os.Stderr, "stubgo: cannot write output:", err)
			return 1
		}
	}
	if loaded.ExtraOutput != "" {
		_ = os.WriteFile(filepath.Join(filepath.Dir(output), loaded.ExtraOutput), []byte("extra"), 0o600) // #nosec G306 -- deliberate extra-output probe
	}
	return 0
}
