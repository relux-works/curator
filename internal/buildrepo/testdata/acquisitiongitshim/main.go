// Command acquisitiongitshim records acquisition inputs before forwarding them
// to the trusted Git executable selected by the test fixture.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/relux-works/curator/internal/stateread"
)

type config struct {
	Git        string `json:"git"`
	Transport  string `json:"transport"`
	Source     string `json:"source"`
	Repository string `json:"repository"`
	LogPath    string `json:"log_path"`
}

func main() {
	executable, err := os.Executable()
	if err != nil {
		fail("cannot locate shim executable")
	}
	file, err := stateread.ReadRegularFile(executable + ".json")
	if err != nil {
		fail("cannot read shim configuration")
	}
	if file.Kind != stateread.KindPresent {
		fail("shim configuration is absent")
	}
	var settings config
	if err := json.Unmarshal(file.Bytes, &settings); err != nil {
		fail("cannot decode shim configuration")
	}
	args := append([]string(nil), os.Args[1:]...)
	log, err := os.OpenFile(settings.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		fail("cannot open shim log")
	}
	if err := json.NewEncoder(log).Encode(struct {
		Args        []string `json:"args"`
		Environment []string `json:"environment"`
	}{Args: args, Environment: os.Environ()}); err != nil {
		_ = log.Close()
		fail("cannot write shim log")
	}
	if err := log.Close(); err != nil {
		fail("cannot close shim log")
	}
	for index, arg := range args {
		if settings.Source != "" && arg == settings.Source {
			args[index] = settings.Repository
			continue
		}
		if arg == "protocol."+settings.Transport+".allow=always" {
			args[index] = "protocol.file.allow=always"
		}
	}
	command := exec.Command(settings.Git, args...) // #nosec G204 G702 -- fixture-owned configuration selects the pinned Git executable and acquisition argv.
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fail("cannot execute pinned Git")
	}
}

func fail(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
