// This test transport rewrites only fixture endpoints and delegates every
// operation to real Git. Its adjacent configuration survives the production
// isolated environment without admitting extra environment variables.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	payload, err := os.ReadFile(filepath.Join(filepath.Dir(executable), "route.json"))
	if err != nil {
		panic(err)
	}
	var route struct {
		Git, Local, Contacts string
	}
	if err := json.Unmarshal(payload, &route); err != nil {
		panic(err)
	}
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "https://example.org/kit.git" || arg == "git@example.org:kit.git" {
			f, err := os.OpenFile(route.Contacts, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
			if err != nil {
				panic(err)
			}
			if _, err := fmt.Fprintln(f, "contact"); err != nil {
				panic(err)
			}
			if err := f.Close(); err != nil {
				panic(err)
			}
			args[i] = route.Local
		}
	}
	cmd := exec.Command(route.Git, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		panic(err)
	}
}
