// Command fixture is a test-owned, non-race Git/SSH subprocess dispatcher.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Binary          string
	Routes          map[string]string
	Refusals        map[string]string
	EnforceHeadless bool
	Direct          []string
}
type pinTransportConfig struct {
	Repo, Marker string
	Fail         bool
}

func main() { os.Exit(run()) }
func run() int {
	defer func() {
		for _, cancel := range cancellations {
			cancel()
		}
	}()
	self, err := os.Executable()
	if err != nil {
		return 98
	}
	name := strings.TrimSuffix(filepath.Base(self), ".exe")
	if name == "ssh" || name == "pin-transport" {
		return runPinTransport()
	}
	path := filepath.Join(filepath.Dir(self), "declared-fixture.json")
	if _, err := os.Stat(path); err == nil {
		return Run(path)
	}
	return runMissingPinGit()
}

// command bounds every subprocess, including remote probes and upload-pack.
var cancellations []context.CancelFunc

func boundedCommand(binary string, args ...string) *exec.Cmd {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.WaitDelay = time.Second
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=", "SSH_ASKPASS_REQUIRE=never")
	cancellations = append(cancellations, cancel)
	return cmd
}
func Run(path string) int {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 98
	}
	var fixture Config
	if json.Unmarshal(raw, &fixture) != nil {
		return 98
	}
	args := append([]string(nil), os.Args[1:]...)
	command := 0
	for command+1 < len(args) && (args[command] == "-C" || args[command] == "-c") {
		command += 2
	}
	if command >= len(args) {
		return 98
	}
	operation := args[command]
	if fixture.EnforceHeadless && (operation == "clone" || operation == "fetch") {
		if os.Getenv("GIT_TERMINAL_PROMPT") != "0" || !strings.Contains(os.Getenv("GIT_SSH_COMMAND"), "BatchMode=yes") || !strings.Contains(os.Getenv("GIT_SSH_COMMAND"), "StrictHostKeyChecking=yes") {
			fmt.Fprintln(os.Stderr, "headless invariant missing")
			return 97
		}
		for _, key := range []string{"GIT_CONFIG_GLOBAL", "GIT_CONFIG_SYSTEM"} {
			data, err := os.ReadFile(os.Getenv(key))
			if err != nil || len(data) != 0 {
				fmt.Fprintln(os.Stderr, "config isolation missing")
				return 97
			}
		}
	}

	original, destination := "", ""
	if operation == "clone" || operation == "fetch" {
		f, err := os.OpenFile(path+".traffic", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return 98
		}
		_, err = fmt.Fprintln(f, strings.Join(args, "\t"))
		_ = f.Close()
		if err != nil {
			return 98
		}
		for i := command + 1; i < len(args); i++ {
			if refusal, ok := fixture.Refusals[args[i]]; ok {
				fmt.Fprintln(os.Stderr, refusal)
				return 1
			}
			if networkEndpoint(args[i]) {
				_, routed := fixture.Routes[args[i]]
				direct := false
				for _, endpoint := range fixture.Direct {
					direct = direct || endpoint == args[i]
				}
				if (!routed && !direct) || (direct && !allowedWire(args[i])) {
					fmt.Fprintln(os.Stderr, "fixture endpoint unavailable")
					return 1
				}
			}
			if repo, ok := fixture.Routes[args[i]]; ok {
				if networkEndpoint(repo) && !allowedWire(repo) {
					return 98
				}
				original = args[i]
				args[i] = fileURL(repo)
				if strings.Contains(repo, "://") {
					args[i] = repo
				}
				if operation == "clone" && i+1 < len(args) {
					destination = args[i+1]
				}
			}
		}
	}
	// Named remotes retain their real Git refspec behavior. Route only
	// explicitly registered fixture URLs, using a test-owned invocation-local
	// rewrite below the observed product boundary. This keeps real Git remote
	// refspec semantics; it is not inherited operator configuration.
	// The product command and its environment were observed above, before this
	// test-only dispatch. Nothing is persisted to the cached Git config.
	if operation == "fetch" {
		for _, option := range args[command+1:] {
			if option != "--all" {
				continue
			}
			query := append(append([]string(nil), args[:command]...), "remote")
			remotes, err := boundedCommand(fixture.Binary, query...).Output()
			if err != nil {
				return 98
			}
			for _, remote := range strings.Fields(string(remotes)) {
				query := append(append([]string(nil), args[:command]...), "config", "--get", "remote."+remote+".url")
				raw, err := boundedCommand(fixture.Binary, query...).Output()
				if err != nil {
					return 98
				}
				endpoint := strings.TrimSpace(string(raw))
				if !networkEndpoint(endpoint) {
					continue
				}
				target, ok := fixture.Routes[endpoint]
				if !ok || (networkEndpoint(target) && !allowedWire(target)) {
					fmt.Fprintln(os.Stderr, "fixture endpoint unavailable")
					return 1
				}
				if !strings.Contains(target, "://") {
					target = fileURL(target)
				}
				args = append([]string{"-c", "url." + target + ".insteadOf=" + endpoint}, args...)
				command += 2
			}
		}
		for _, operand := range args[command+1:] {
			if strings.HasPrefix(operand, "-") || strings.ContainsAny(operand, "/:\\") {
				continue
			}
			query := append(append([]string(nil), args[:command]...), "config", "--get", "remote."+operand+".url")
			raw, err := boundedCommand(fixture.Binary, query...).Output()
			if err != nil {
				continue
			}
			endpoint := strings.TrimSpace(string(raw))
			if target, ok := fixture.Routes[endpoint]; ok {
				if networkEndpoint(target) && !allowedWire(target) {
					return 98
				}
				if !strings.Contains(target, "://") {
					target = fileURL(target)
				}
				args = append([]string{"-c", "url." + target + ".insteadOf=" + strings.TrimSpace(string(raw))}, args...)
				break
			} else if networkEndpoint(endpoint) {
				direct := false
				for _, allowed := range fixture.Direct {
					direct = direct || endpoint == allowed
				}
				if !direct || !allowedWire(endpoint) {
					fmt.Fprintln(os.Stderr, "fixture endpoint unavailable")
					return 1
				}
			}
		}
	}
	cmd := boundedCommand(fixture.Binary, args...)
	if ssh, err := exec.LookPath("ssh"); err == nil {
		if _, err := os.Stat(ssh + ".json"); err == nil {
			value := os.Getenv("GIT_SSH_COMMAND")
			if strings.HasPrefix(value, "ssh ") {
				cmd.Env = append(cmd.Env, "GIT_SSH_COMMAND='"+strings.ReplaceAll(filepath.ToSlash(ssh), "'", "'\\''")+"' "+strings.TrimPrefix(value, "ssh "))
			}
		}
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return e.ExitCode()
		}
		return 98
	}
	if destination != "" {
		cmd := boundedCommand(fixture.Binary, "-C", destination, "config", "remote.origin.url", original)
		if cmd.Run() != nil {
			return 98
		}
	}
	return 0
}

func fileURL(path string) string {
	if filepath.VolumeName(path) != "" {
		return "file:///" + filepath.ToSlash(path)
	}
	return "file://" + filepath.ToSlash(path)
}

type pinGitConfig struct{ Binary, Mode, Trace string }

func runMissingPinGit() int {
	self, err := os.Executable()
	if err != nil {
		return 98
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(self), "pin-git.json"))
	if err != nil {
		return 98
	}
	var config pinGitConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return 98
	}
	args := os.Args[1:]
	mode := config.Mode
	// Isolated Git deliberately discards ambient tracing. The executable
	// observer's sidecar survives that environment boundary.
	if os.Getenv("GIT_TRACE") == "" && mode == "observe" {
		f, err := os.OpenFile(config.Trace, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return 98
		}
		command := args
		for len(command) >= 2 && (command[0] == "-C" || command[0] == "-c") {
			command = command[2:]
		}
		_, err = fmt.Fprintln(f, "observer: built-in: git "+strings.Join(command, " "))
		if closeErr := f.Close(); err != nil || closeErr != nil {
			return 98
		}
	}
	command := args
	for len(command) >= 2 && (command[0] == "-C" || command[0] == "-c") {
		command = command[2:]
	}
	if len(command) > 0 && command[0] == "fetch" && mode == "still-absent" {
		prior, _ := os.ReadFile(config.Trace)
		f, err := os.OpenFile(config.Trace, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return 98
		}
		_, err = fmt.Fprintln(f, "fault plant: built-in: git fetch successful without object")
		if closeErr := f.Close(); err != nil || closeErr != nil {
			return 98
		}
		// Bound the known-violation control even if the retry guard mutates.
		if strings.Contains(string(prior), "fault plant: built-in: git fetch") {
			return 96
		}
		return 0
	}
	if len(args) >= 3 && args[2] == "cat-file" && mode == "probe-failure" {
		return 97
	}
	cmd := boundedCommand(config.Binary, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		return 99
	}
	return 0
}

func runPinTransport() int {
	for i, arg := range os.Args[1:] {
		if arg == "-F" && i+2 < len(os.Args) {
			raw, err := os.ReadFile(os.Args[i+2])
			if err != nil || len(raw) != 0 {
				fmt.Fprintln(os.Stderr, "SSH configuration was not isolated")
				return 97
			}
		}
	}
	for _, arg := range os.Args[1:] {
		if arg == "-G" {
			return 0
		}
	}

	self, err := os.Executable()
	if err != nil {
		return 98
	}
	raw, err := os.ReadFile(self + ".json")
	if err != nil {
		return 98
	}
	var config pinTransportConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		return 98
	}
	if err := os.WriteFile(config.Marker, []byte(strings.Join(os.Args, "\n")), 0o600); err != nil {
		return 98
	}
	if config.Fail {
		return 1
	}
	cmd := boundedCommand("git", "upload-pack", config.Repo)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return 1
	}
	return 0
}

var scpEndpoint = regexp.MustCompile(`^(?:[^/@:\s]+@)?[A-Za-z0-9][A-Za-z0-9.-]*:[^/\\\s]`)

func networkEndpoint(value string) bool {
	return (strings.Contains(value, "://") && !strings.HasPrefix(value, "file://")) || scpEndpoint.MatchString(value)
}
func localHTTP(value string) bool {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return false
	}
	ip := net.ParseIP(u.Hostname())
	return ip != nil && ip.IsLoopback()
}

func allowedWire(value string) bool {
	if localHTTP(value) {
		return true
	}
	if strings.HasPrefix(value, "ssh://") || scpEndpoint.MatchString(value) {
		ssh, err := exec.LookPath("ssh")
		if err != nil {
			return false
		}
		_, err = os.Stat(ssh + ".json")
		return err == nil
	}
	return false
}
