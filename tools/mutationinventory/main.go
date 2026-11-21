// Command mutationinventory exposes the canonical deferred selectors and
// records conservation evidence for unfiltered cmd/curator test inventories.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/curatormutations"
)

type inventoryReport struct {
	Commit              string   `json:"commit"`
	GOOS                string   `json:"goos"`
	GOARCH              string   `json:"goarch"`
	GoVersion           string   `json:"go_version"`
	GOFLAGSHash         string   `json:"goflags_sha256"`
	Package             string   `json:"package"`
	Timeout             string   `json:"timeout"`
	GoTestFlags         []string `json:"go_test_flags"`
	MutationSelector    string   `json:"mutation_selector"`
	PostLandingSelector string   `json:"post_landing_selector"`
	Unfiltered          []string `json:"unfiltered"`
	PullRequest         []string `json:"pull_request"`
	Mutations           []string `json:"mutations"`
	Exhaustive          []string `json:"exhaustive"`
	PostLanding         []string `json:"post_landing"`
}

const curatorPackagePath = "./cmd/curator"
const curatorPackageImportPath = "github.com/relux-works/curator/cmd/curator"
const inventoryTimeout = "60m"
const windowsInventoryTimeout = "120m"

type inventoryCommandProfile struct {
	packagePath string
	timeout     string
	race        bool
}

type inventorySelection uint8

const (
	unfilteredInventory inventorySelection = iota
	mutationInventory
	postLandingInventory
)

func newInventoryCommandProfile(packagePath, timeout string, flags []string) (inventoryCommandProfile, error) {
	profile := inventoryCommandProfile{}
	switch packagePath {
	case curatorPackagePath:
		profile.packagePath = curatorPackagePath
	case curatorPackageImportPath:
		profile.packagePath = curatorPackageImportPath
	default:
		return inventoryCommandProfile{}, fmt.Errorf("unsupported inventory package %q", packagePath)
	}

	switch timeout {
	case inventoryTimeout:
		profile.timeout = inventoryTimeout
	case windowsInventoryTimeout:
		profile.timeout = windowsInventoryTimeout
	default:
		return inventoryCommandProfile{}, fmt.Errorf("unsupported inventory timeout %q", timeout)
	}

	switch {
	case len(flags) == 0:
	case len(flags) == 1 && flags[0] == "-race" && profile.timeout == inventoryTimeout:
		profile.race = true
	default:
		return inventoryCommandProfile{}, fmt.Errorf("unsupported inventory Go test flags %q", flags)
	}
	return profile, nil
}

func (profile inventoryCommandProfile) goTestFlags() []string {
	if profile.race {
		return []string{"-race"}
	}
	return nil
}

func (profile inventoryCommandProfile) selector(selection inventorySelection) (string, error) {
	switch selection {
	case unfilteredInventory:
		return ".", nil
	case mutationInventory:
		return curatormutations.Selector(), nil
	case postLandingInventory:
		return curatormutations.PostLandingSelector(), nil
	default:
		return "", fmt.Errorf("unsupported inventory selection %d", selection)
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "mutationinventory:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("mutationinventory", flag.ContinueOnError)
	mode := fs.String("mode", "", "selector, post-landing-selector, inventory, or verify-invocation")
	packagePath := fs.String("package", "./cmd/curator", "package to inventory")
	timeout := fs.String("timeout", "60m", "go test package timeout")
	goTestFlags := fs.String("go-test-flags", "", "additional go test flags, separated by spaces")
	extraGoTestFlags := fs.String("extra-go-test-flags", "", "additional flags supplied by the gate, separated by spaces")
	evidence := fs.String("evidence", "", "directory for the inventory report")
	requireAll := fs.Bool("require-all", false, "require the complete post-landing test set")
	if err := fs.Parse(args); err != nil {
		return err
	}
	remaining := fs.Args()

	switch *mode {
	case "selector":
		if len(remaining) != 0 {
			return errors.New("selector mode takes no arguments")
		}
		fmt.Println(curatormutations.Selector())
		return nil
	case "post-landing-selector":
		if len(remaining) != 0 {
			return errors.New("post-landing-selector mode takes no arguments")
		}
		fmt.Println(curatormutations.PostLandingSelector())
		return nil
	case "verify-invocation":
		if err := curatormutations.ValidateInvocation(remaining); err != nil {
			return err
		}
		return nil
	case "inventory":
		if len(remaining) != 0 {
			return errors.New("inventory mode takes no arguments")
		}
		if *packagePath == "" || *timeout == "" || *evidence == "" {
			return errors.New("inventory mode requires package, timeout, and evidence")
		}
		flags := append(strings.Fields(*extraGoTestFlags), strings.Fields(*goTestFlags)...)
		profile, err := newInventoryCommandProfile(*packagePath, *timeout, flags)
		if err != nil {
			return err
		}
		return writeInventory(profile, *evidence, *requireAll)
	default:
		return fmt.Errorf("unknown mode %q", *mode)
	}
}

func writeInventory(profile inventoryCommandProfile, evidence string, requireAll bool) error {
	goVersion, err := commandOutput(exec.Command("go", "version"))
	if err != nil {
		return err
	}
	if err := verifyGoLauncherIdentity(goVersion, runtime.Version()); err != nil {
		return err
	}

	all, err := listInventory(profile, unfilteredInventory)
	if err != nil {
		return fmt.Errorf("unfiltered inventory: %w", err)
	}
	if len(all) == 0 {
		return fmt.Errorf("empty unfiltered test inventory for %s", profile.packagePath)
	}
	inventory, err := curatormutations.Split(all)
	if err != nil {
		return err
	}
	selectedMutations, err := listInventory(profile, mutationInventory)
	if err != nil {
		return fmt.Errorf("mutation -run inventory: %w", err)
	}
	if err := verifyMutationSelection(inventory.Mutations, selectedMutations); err != nil {
		return err
	}
	selectedPostLanding, err := listInventory(profile, postLandingInventory)
	if err != nil {
		return fmt.Errorf("post-landing -run inventory: %w", err)
	}
	expectedPostLanding := append(append([]string(nil), inventory.Mutations...), inventory.Exhaustive...)
	if err := verifyPostLandingSelection(expectedPostLanding, selectedPostLanding); err != nil {
		return err
	}
	if err := validateInventoryPartition(inventory); err != nil {
		return err
	}
	if requireAll {
		if err := curatormutations.RequirePostLandingComplete(inventory); err != nil {
			return err
		}
	}

	commit, err := commandOutput(exec.Command("git", "rev-parse", "HEAD"))
	if err != nil {
		return err
	}
	goEnv, err := commandOutput(exec.Command("go", "env", "GOOS", "GOARCH"))
	if err != nil {
		return err
	}
	platform := strings.Fields(goEnv)
	if len(platform) != 2 {
		return fmt.Errorf("unexpected go env platform output %q", goEnv)
	}
	if platform[0] == "" {
		platform[0] = runtime.GOOS
	}
	if platform[1] == "" {
		platform[1] = runtime.GOARCH
	}
	if err := validateMacRaceInventory(platform[0], profile, inventory); err != nil {
		return err
	}
	postLanding := append(append([]string(nil), inventory.Mutations...), inventory.Exhaustive...)
	sort.Strings(postLanding)
	effectiveGoFlags, err := commandOutput(exec.Command("go", "env", "GOFLAGS"))
	if err != nil {
		return err
	}
	report := inventoryReport{
		Commit:              strings.TrimSpace(commit),
		GOOS:                platform[0],
		GOARCH:              platform[1],
		GoVersion:           strings.TrimSpace(goVersion),
		GOFLAGSHash:         fmt.Sprintf("%x", sha256.Sum256([]byte(effectiveGoFlags))),
		Package:             profile.packagePath,
		Timeout:             profile.timeout,
		GoTestFlags:         profile.goTestFlags(),
		MutationSelector:    curatormutations.Selector(),
		PostLandingSelector: curatormutations.PostLandingSelector(),
		Unfiltered:          inventory.All,
		PullRequest:         inventory.Pull,
		Mutations:           inventory.Mutations,
		Exhaustive:          inventory.Exhaustive,
		PostLanding:         postLanding,
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(evidence, 0755); err != nil {
		return err
	}
	path := filepath.Join(evidence, "mutation-inventory.json")
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Printf("inventory %s: unfiltered=%d PR=%d mutation=%d exhaustive=%d post-landing=%d evidence=%s\n", profile.packagePath, len(inventory.All), len(inventory.Pull), len(inventory.Mutations), len(inventory.Exhaustive), len(postLanding), path)
	return nil
}

func listInventory(profile inventoryCommandProfile, selection inventorySelection) ([]string, error) {
	cmd, err := inventoryCommand(profile, selection)
	if err != nil {
		return nil, err
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command %q failed: %w", cmd.Args, err)
	}
	return parseInventory(stdout.String()), nil
}

func inventoryCommand(profile inventoryCommandProfile, selection inventorySelection) (*exec.Cmd, error) {
	selector, err := profile.selector(selection)
	if err != nil {
		return nil, err
	}
	args := []string{"test", "-list", selector, "-timeout", profile.timeout}
	args = append(args, profile.goTestFlags()...)
	args = append(args, profile.packagePath)
	//nolint:gosec // The profile is built only from the supported CI compatibility table; selectors, flags, timeout, and package are tool-owned constants.
	return exec.Command("go", args...), nil
}

func verifyGoLauncherIdentity(versionOutput, toolVersion string) error {
	fields := strings.Fields(versionOutput)
	if len(fields) < 3 || fields[0] != "go" || fields[1] != "version" || fields[2] != toolVersion {
		return fmt.Errorf("go launcher version does not match inventory tool: launcher=%q tool=%q", versionOutput, toolVersion)
	}
	return nil
}

func verifyMutationSelection(expected, selected []string) error {
	return verifyNamedSelection("mutation", expected, selected)
}

func verifyPostLandingSelection(expected, selected []string) error {
	return verifyNamedSelection("post-landing", expected, selected)
}

func verifyNamedSelection(lane string, expected, selected []string) error {
	expected = append([]string(nil), expected...)
	selected = append([]string(nil), selected...)
	sort.Strings(expected)
	sort.Strings(selected)
	if !reflect.DeepEqual(expected, selected) {
		return fmt.Errorf("%s selector inventory mismatch: expected partition=%v selected by -run=%v", lane, expected, selected)
	}
	return nil
}

func validateInventoryPartition(inventory curatormutations.Inventory) error {
	all := make(map[string]struct{}, len(inventory.All))
	for _, name := range inventory.All {
		if name == "" {
			return fmt.Errorf("unfiltered inventory contains an empty test name")
		}
		if _, exists := all[name]; exists {
			return fmt.Errorf("unfiltered inventory contains duplicate test %q", name)
		}
		all[name] = struct{}{}
	}

	partition := make(map[string]string, len(inventory.Pull)+len(inventory.Mutations)+len(inventory.Exhaustive))
	for _, group := range []struct {
		name  string
		tests []string
	}{{name: "PR", tests: inventory.Pull}, {name: "mutation", tests: inventory.Mutations}, {name: "exhaustive", tests: inventory.Exhaustive}} {
		for _, name := range group.tests {
			if name == "" {
				return fmt.Errorf("%s inventory contains an empty test name", group.name)
			}
			if previous, exists := partition[name]; exists {
				if previous != group.name {
					return fmt.Errorf("test %q appears in both %s and %s inventories", name, previous, group.name)
				}
				return fmt.Errorf("%s inventory contains duplicate test %q", group.name, name)
			}
			if _, exists := all[name]; !exists {
				return fmt.Errorf("%s inventory contains test %q absent from unfiltered inventory", group.name, name)
			}
			partition[name] = group.name
		}
	}
	for name := range all {
		if _, exists := partition[name]; !exists {
			return fmt.Errorf("unfiltered test %q is missing from PR, mutation, and exhaustive inventories", name)
		}
	}
	return nil
}

func validateMacRaceInventory(goos string, profile inventoryCommandProfile, inventory curatormutations.Inventory) error {
	if goos != "darwin" || !profile.race {
		return nil
	}
	if err := validateInventoryPartition(inventory); err != nil {
		return fmt.Errorf("race macOS inventory partition for %s: %w", profile.packagePath, err)
	}
	if err := curatormutations.RequirePostLandingComplete(inventory); err != nil {
		return fmt.Errorf("race macOS post-landing inventory for %s: %w", profile.packagePath, err)
	}
	return nil
}

func parseInventory(output string) []string {
	var names []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Test") || strings.HasPrefix(line, "Example") || strings.HasPrefix(line, "Fuzz") {
			names = append(names, line)
		}
	}
	sort.Strings(names)
	return names
}

func commandOutput(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", strings.Join(cmd.Args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}
