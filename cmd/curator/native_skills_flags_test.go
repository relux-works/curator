package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// Every unrelated registered command rejects takeover through run(), both with
// and without help; the same help invocation without takeover succeeds. Help
// is the positive control for group topics and commands requiring live input.
func TestNativeSkillsTakeoverEveryRegisteredCommand(t *testing.T) {
	allowed := map[string]bool{
		"profile install": true, "profile use": true, "profile sync": true,
		"profile update": true, "env resolve": true,
	}
	for _, command := range commandRegistry {
		if allowed[command.name] {
			continue
		}
		t.Run(strings.ReplaceAll(command.name, " ", "/"), func(t *testing.T) {
			source, home := profileHome(t)
			pinOperatorHome(t)
			args := strings.Fields(command.name)
			before := snapshotHelpTree(t, home)
			if code, _, stderr := runProfile(t, source, append(append([]string{}, args...), "--help")...); code != exitOK {
				t.Fatalf("positive help control = %d: %s", code, stderr)
			}
			for _, flag := range []string{"--takeover", "--takeover=true", "--takeover=false", "-takeover"} {
				for _, help := range []bool{false, true} {
					invocation := append(append([]string{}, args...), flag)
					if help {
						invocation = append(invocation, "--help")
					}
					if code, _, stderr := runProfile(t, source, invocation...); code != exitUsage || !strings.Contains(stderr, "--takeover") {
						t.Fatalf("%v: unrelated takeover admitted: exit %d, want 2; %s", invocation, code, stderr)
					}
				}
			}
			if !reflect.DeepEqual(before, snapshotHelpTree(t, home)) {
				t.Fatal("refusal wrote manager state")
			}
		})
	}
}

// The ignored-argument regression and neighboring noninteractive wrappers have
// real successful no-flag controls, then reject consent without changing state.
func TestNativeSkillsTakeoverArgumentDiscardRegression(t *testing.T) {
	for _, name := range []string{"list", "update", "gc", "version", "--version"} {
		t.Run(name, func(t *testing.T) {
			source, home := profileHome(t)
			pinOperatorHome(t)
			source.cfg.SkillsRoot = t.TempDir()
			if code, _, stderr := runProfile(t, source, name); code != exitOK {
				t.Fatalf("positive no-flag control = %d: %s", code, stderr)
			}
			before := snapshotHelpTree(t, home)
			if code, _, stderr := runProfile(t, source, name, "--takeover"); code != exitUsage || !strings.Contains(stderr, "--takeover") {
				t.Fatalf("unrelated takeover admitted: exit %d, want 2; %s", code, stderr)
			}
			if !reflect.DeepEqual(before, snapshotHelpTree(t, home)) {
				t.Fatal("refusal wrote manager state")
			}
		})
	}
}

// Install/resolve require the activating operation; -- terminates consent
// parsing, so a following literal remains an operand rather than a flag.
func TestNativeSkillsTakeoverActivationBoundary(t *testing.T) {
	source, _ := profileHome(t)
	pinOperatorHome(t)
	for _, args := range [][]string{
		{"profile", "install", "sample"}, {"profile", "install", "sample", "--use=false"},
		{"env", "resolve", "claude_code"}, {"env", "resolve", "claude_code", "--repair=false"},
	} {
		if code, _, stderr := runProfile(t, source, append(args, "--takeover")...); code != exitUsage || !strings.Contains(stderr, "--takeover") {
			t.Fatalf("%v: exit %d: %s", args, code, stderr)
		}
	}
	if code, _, stderr := runProfile(t, source, "list", "--", "--takeover"); code != exitOK {
		t.Fatalf("literal operand refused: exit %d: %s", code, stderr)
	}
}

// Permit only list to bypass the still-present consent gate. The behavioral
// sweep must fail by name with exit 1, not by setup/compile failure or timeout.
func TestNativeSkillsTakeoverRegistryNarrowingMutant(t *testing.T) {
	original, err := filepath.Abs("command_help.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "switch command.name {\n\t\tcase \"profile use\", \"profile sync\", \"profile update\":"
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.Count(source, anchor) != 1 {
		t.Fatal("mutation anchor is not unique")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "replacement.go")
	mutated := strings.Replace(source, anchor, "switch command.name {\n\t\tcase \"list\":\n\t\t\treturn false\n\t\tcase \"profile use\", \"profile sync\", \"profile update\":", 1)
	if err := os.WriteFile(replacement, []byte(mutated), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, data, 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-count=1", "-timeout=3m", "-v", "-run", "^TestNativeSkillsTakeover(EveryRegisteredCommand|ArgumentDiscardRegression)$/^list$", ".")
	command.Env = append(os.Environ(), cliHelperProcess+"=1", "GOFLAGS="+os.Getenv("GOFLAGS")+" "+strconv.Quote("-overlay="+overlay))
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "unrelated takeover admitted: exit 0, want 2") ||
		!strings.Contains(string(output), "--- FAIL: TestNativeSkillsTakeoverEveryRegisteredCommand/list") ||
		!strings.Contains(string(output), "--- FAIL: TestNativeSkillsTakeoverArgumentDiscardRegression/list") || strings.Contains(string(output), "build failed") {
		t.Fatalf("list-only mutant lacks intended behavioral exit 1: %v\n%s", err, output)
	}
	t.Log("list-only-consent-bypass: child exit=1; killed by TestNativeSkillsTakeoverEveryRegisteredCommand/list and TestNativeSkillsTakeoverArgumentDiscardRegression/list")
}
