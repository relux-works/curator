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

	"github.com/relux-works/curator/internal/testtoolchain"
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
		{"profile", "install", "sample", "--use=0"}, {"env", "resolve", "claude_code", "--repair=0"},
	} {
		if code, _, stderr := runProfile(t, source, append(args, "--takeover")...); code != exitUsage || !strings.Contains(stderr, "--takeover") {
			t.Fatalf("%v: exit %d: %s", args, code, stderr)
		}
	}
	for _, args := range [][]string{
		{"profile", "install", "sample", "--use=1"}, {"env", "resolve", "claude_code", "--repair=TRUE"},
	} {
		if code, _, stderr := runProfile(t, source, append(args, "--takeover", "--help")...); code != exitOK {
			t.Fatalf("Go boolean spelling refused: %v: exit %d: %s", args, code, stderr)
		}
	}
	if code, _, stderr := runProfile(t, source, "list", "--", "--takeover"); code != exitOK {
		t.Fatalf("literal operand refused: exit %d: %s", code, stderr)
	}
}

// Resolve help refuses takeover without true, non-dry repair, leaves both homes
// unchanged, and retains successful no-consent and activating help controls.
func TestReviewNativeTakeoverHelpBoundary(t *testing.T) {
	source, home := profileHome(t)
	pinOperatorHome(t)
	nativeHome := os.Getenv("HOME")
	before, nativeBefore := snapshotHelpTree(t, home), snapshotHelpTree(t, nativeHome)
	for _, extra := range [][]string{nil, {"--repair=false"}, {"--repair=0"}, {"--repair", "--dry-run"}} {
		base := append([]string{"env", "resolve", "claude_code"}, extra...)
		if code, _, stderr := runProfile(t, source, append(append([]string{}, base...), "--help")...); code != exitOK {
			t.Fatalf("positive help control: %d %s", code, stderr)
		}
		args := append(append([]string{}, base...), "--takeover", "--help")
		if code, _, stderr := runProfile(t, source, args...); code != exitUsage || !strings.Contains(stderr, "--takeover") {
			t.Errorf("%v: takeover without an activating repair admitted: exit %d, want 2; %s", args, code, stderr)
		}
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "claude_code", "--repair", "--takeover", "--help"); code != exitOK {
		t.Errorf("activating help control: %d %s", code, stderr)
	}
	if !reflect.DeepEqual(before, snapshotHelpTree(t, home)) || !reflect.DeepEqual(nativeBefore, snapshotHelpTree(t, nativeHome)) {
		t.Fatal("help or refusal wrote fixture state")
	}
}

// All registry topics and activating boolean forms enforce the same admission
// before -h, --help, and help <command>; no-consent controls all succeed.
func TestNativeSkillsTakeoverHelpCrossProduct(t *testing.T) {
	source, home := profileHome(t)
	pinOperatorHome(t)
	nativeHome := os.Getenv("HOME")
	before, nativeBefore := snapshotHelpTree(t, home), snapshotHelpTree(t, nativeHome)
	for _, command := range commandRegistry {
		variants := []struct {
			flags []string
			admit bool
		}{{nil, command.name == "profile use" || command.name == "profile sync" || command.name == "profile update"}}
		activation := ""
		switch command.name {
		case "profile install":
			activation = "--use"
		case "env resolve":
			activation = "--repair"
		}
		if activation != "" {
			for _, value := range []string{"", "=true", "=1", "=TRUE", "=false", "=0"} {
				variants = append(variants, struct {
					flags []string
					admit bool
				}{[]string{activation + value}, value != "=false" && value != "=0"})
			}
		}
		if command.name == "env resolve" {
			for _, variant := range append(variants[:0:0], variants...) {
				for _, value := range []string{"", "=true", "=1", "=false", "=0"} {
					variants = append(variants, struct {
						flags []string
						admit bool
					}{append(append([]string{}, variant.flags...), "--dry-run"+value), variant.admit && (value == "=false" || value == "=0")})
				}
			}
		}
		for _, variant := range variants {
			for _, help := range []string{"-h", "--help", "prefix -h", "prefix --help", "help"} {
				t.Run(command.name+"/"+strings.Join(variant.flags, ",")+"/"+help, func(t *testing.T) {
					args := append(strings.Fields(command.name), variant.flags...)
					withHelp := func(args []string) []string {
						if help == "help" {
							return append([]string{"help"}, args...)
						}
						if strings.HasPrefix(help, "prefix ") {
							return append([]string{strings.TrimPrefix(help, "prefix ")}, args...)
						}
						return append(append([]string{}, args...), help)
					}
					if code, _, stderr := runProfile(t, source, withHelp(args)...); code != exitOK {
						t.Fatalf("no-consent help control = %d: %s", code, stderr)
					}
					for _, flag := range []string{"--takeover", "--takeover=true", "--takeover=false", "--takeover=0", "-takeover"} {
						invocation := withHelp(append(append([]string{}, args...), flag))
						want := exitUsage
						if variant.admit {
							want = exitOK
						}
						if code, _, stderr := runProfile(t, source, invocation...); code != want || (want == exitUsage && !strings.Contains(stderr, "--takeover")) {
							t.Errorf("%v: exit %d, want %d: %s", invocation, code, want, stderr)
						}
					}
				})
			}
		}
	}
	if !reflect.DeepEqual(before, snapshotHelpTree(t, home)) || !reflect.DeepEqual(nativeBefore, snapshotHelpTree(t, nativeHome)) {
		t.Fatal("help sweep wrote fixture state")
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
	anchor := "switch command.name {"
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.Count(source, anchor) != 1 {
		t.Fatal("mutation anchor is not unique")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "replacement.go")
	mutated := strings.Replace(source, anchor, anchor+"\n\t\tcase \"list\":\n\t\t\treturn false", 1)
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
	command := exec.Command(testtoolchain.Go(t), "test", "-count=1", "-timeout=3m", "-v", "-run", "^TestNativeSkillsTakeover(EveryRegisteredCommand|ArgumentDiscardRegression)$/^list$", ".")
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

// Each plant retains dispatch refusal but permits one invalid resolve-help
// class; the permanent reviewer witness must fail with successful admission.
func TestNativeSkillsTakeoverHelpNarrowingMutant(t *testing.T) {
	original, err := filepath.Abs("command_help.go")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "return !repair || dryRun"
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.Count(source, anchor) != 1 {
		t.Fatal("mutation anchor is not unique")
	}
	for _, mutant := range []struct{ name, after string }{
		{"missing-repair", "return (!repair && !hasHelpFlag(args)) || dryRun"},
		{"dry-run", "return !repair || (dryRun && !hasHelpFlag(args))"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			dir := t.TempDir()
			replacement := filepath.Join(dir, "replacement.go")
			mutated := strings.Replace(source, anchor, mutant.after, 1)
			if err := os.WriteFile(replacement, []byte(mutated), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlay, data, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(testtoolchain.Go(t), "test", "-count=1", "-timeout=3m", "-v", "-run", "^TestReviewNativeTakeoverHelpBoundary$", ".")
			command.Env = append(os.Environ(), cliHelperProcess+"=1", "GOFLAGS="+os.Getenv("GOFLAGS")+" "+strconv.Quote("-overlay="+overlay))
			output, err := command.CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "takeover without an activating repair admitted: exit 0, want 2") ||
				!strings.Contains(string(output), "--- FAIL: TestReviewNativeTakeoverHelpBoundary") || strings.Contains(string(output), "build failed") {
				t.Fatalf("resolve-help mutant lacks intended behavioral exit 1: %v\n%s", err, output)
			}
			t.Logf("resolve-help-%s: child exit=1; killed by TestReviewNativeTakeoverHelpBoundary", mutant.name)
		})
	}
}
