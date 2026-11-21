// Package curatormutations defines tests reserved for the post-landing lane.
package curatormutations

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var mutationTests = []string{
	"TestCommandHelpNarrowingMutants",
	"TestMuseCredentialPolicyNarrowingMutants",
	"TestMuseRootContextNarrowingMutants",
	"TestNativeSkillsTakeoverHelpNarrowingMutant",
	"TestNativeSkillsTakeoverRegistryNarrowingMutant",
	"TestCLIPathIsolationNarrowingMutant",
	"TestWindowsGitBashSelectionNarrowingMutant",
}

var exhaustiveTests = []string{
	"TestGlobalCommandsMemoryPolicyDoesNotFollowPreviousCWD",
}

// Names returns a copy of the exact mutation-test inventory.
func Names() []string {
	names := append([]string(nil), mutationTests...)
	sort.Strings(names)
	return names
}

// ExhaustiveNames returns a copy of the bounded exhaustive product-behavior
// tests reserved for the post-landing lane.
func ExhaustiveNames() []string {
	names := append([]string(nil), exhaustiveTests...)
	sort.Strings(names)
	return names
}

// PostLandingNames returns the exact union of mutation and exhaustive tests.
func PostLandingNames() []string {
	names := append(Names(), ExhaustiveNames()...)
	sort.Strings(names)
	return names
}

// Selector is the exact top-level Go test selector for the mutation suite.
func Selector() string {
	return selector(Names())
}

// PostLandingSelector selects every deferred mutation and exhaustive test.
func PostLandingSelector() string {
	return selector(PostLandingNames())
}

func selector(names []string) string {
	for i, name := range names {
		names[i] = regexp.QuoteMeta(name)
	}
	return "^(" + strings.Join(names, "|") + ")$"
}

// Inventory partitions one package's unfiltered test names into ordinary PR,
// mutation, and exhaustive post-landing sets.
type Inventory struct {
	All        []string
	Pull       []string
	Mutations  []string
	Exhaustive []string
}

// Split rejects duplicate names and any newly added mutant-style test that is
// not explicitly admitted to the deferred inventory. Use it for the package
// that owns the post-landing tests.
func Split(names []string) (Inventory, error) {
	return split(names, true)
}

// SplitOtherPackage partitions another package's tests without treating its
// descriptive "NarrowingMutant" names as members of the cmd/curator suite.
// Canonical deferred names remain classified so callers can reject their
// appearance outside the owning package.
func SplitOtherPackage(names []string) (Inventory, error) {
	return split(names, false)
}

func split(names []string, rejectUnclassifiedMutants bool) (Inventory, error) {
	known := make(map[string]bool, len(mutationTests))
	for _, name := range mutationTests {
		known[name] = true
	}
	exhaustive := make(map[string]bool, len(exhaustiveTests))
	for _, name := range exhaustiveTests {
		exhaustive[name] = true
	}

	seen := make(map[string]bool, len(names))
	result := Inventory{All: append([]string(nil), names...)}
	for _, name := range names {
		if name == "" || seen[name] {
			return Inventory{}, fmt.Errorf("empty or duplicate test name %q", name)
		}
		seen[name] = true
		if known[name] {
			result.Mutations = append(result.Mutations, name)
			continue
		}
		if exhaustive[name] {
			result.Exhaustive = append(result.Exhaustive, name)
			continue
		}
		if rejectUnclassifiedMutants && strings.Contains(name, "NarrowingMutant") {
			return Inventory{}, fmt.Errorf("unclassified narrowing mutant test %q", name)
		}
		result.Pull = append(result.Pull, name)
	}
	sort.Strings(result.All)
	sort.Strings(result.Pull)
	sort.Strings(result.Mutations)
	sort.Strings(result.Exhaustive)
	return result, nil
}

// RequireComplete checks that the current platform exposes all seven
// canonical mutation tests.
func RequireComplete(inventory Inventory) error {
	if len(inventory.Mutations) != len(mutationTests) {
		return fmt.Errorf("mutation inventory has %d tests, want %d", len(inventory.Mutations), len(mutationTests))
	}
	got := make(map[string]bool, len(inventory.Mutations))
	for _, name := range inventory.Mutations {
		got[name] = true
	}
	for _, name := range mutationTests {
		if !got[name] {
			return fmt.Errorf("mutation inventory is missing %s", name)
		}
	}
	return nil
}

// RequirePostLandingComplete checks the full post-landing partition: all seven
// mutation tests and every explicitly listed exhaustive test must be present.
func RequirePostLandingComplete(inventory Inventory) error {
	if err := RequireComplete(inventory); err != nil {
		return err
	}
	if len(inventory.Exhaustive) != len(exhaustiveTests) {
		return fmt.Errorf("post-landing exhaustive inventory has %d tests, want %d", len(inventory.Exhaustive), len(exhaustiveTests))
	}
	got := make(map[string]bool, len(inventory.Exhaustive))
	for _, name := range inventory.Exhaustive {
		got[name] = true
	}
	for _, name := range exhaustiveTests {
		if !got[name] {
			return fmt.Errorf("post-landing inventory is missing exhaustive test %s", name)
		}
	}
	return nil
}

// SkipArgs returns the Go testing flags as a separate flag and value.
func SkipArgs() []string {
	return []string{"-skip", PostLandingSelector()}
}

// ValidateSkipArgs requires one standalone -skip with the complete post-landing
// selector. It rejects -skip=<selector> so CI keeps the flag and value separate.
func ValidateSkipArgs(args []string) error {
	count := 0
	waitingForSelector := false
	for _, arg := range args {
		if waitingForSelector {
			if arg != PostLandingSelector() {
				return fmt.Errorf("-skip must be followed by the canonical post-landing selector")
			}
			waitingForSelector = false
			continue
		}
		if strings.HasPrefix(arg, "-skip=") {
			return fmt.Errorf("post-landing selector must use a separate -skip argument")
		}
		if arg != "-skip" {
			continue
		}
		count++
		waitingForSelector = true
	}
	if waitingForSelector {
		return fmt.Errorf("-skip must be followed by the canonical post-landing selector")
	}
	if count != 1 {
		return fmt.Errorf("want exactly one canonical post-landing -skip pair, got %d", count)
	}
	return nil
}

// ValidateInvocation checks the Go test portion of a command vector that may
// begin with env assignments. It is used immediately before test-gate runs a
// full package invocation.
func ValidateInvocation(args []string) error {
	for i, arg := range args {
		if arg == "test" {
			return ValidateSkipArgs(args[i+1:])
		}
	}
	return fmt.Errorf("command does not contain a go test invocation")
}
