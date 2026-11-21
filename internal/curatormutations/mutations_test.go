package curatormutations

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// Claim: canonical mutation and exhaustive names map to exact selectors, and returned names cannot mutate the inventory.
// Limit: this checks declared names and selectors, not package discovery or test execution.
func TestCanonicalInventoryAndSelector(t *testing.T) {
	want := []string{
		"TestCLIPathIsolationNarrowingMutant",
		"TestCommandHelpNarrowingMutants",
		"TestMuseCredentialPolicyNarrowingMutants",
		"TestMuseRootContextNarrowingMutants",
		"TestNativeSkillsTakeoverHelpNarrowingMutant",
		"TestNativeSkillsTakeoverRegistryNarrowingMutant",
		"TestWindowsGitBashSelectionNarrowingMutant",
	}
	if got := Names(); !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical inventory = %v, want %v", got, want)
	}
	if got := Selector(); got != "^(TestCLIPathIsolationNarrowingMutant|TestCommandHelpNarrowingMutants|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextNarrowingMutants|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestWindowsGitBashSelectionNarrowingMutant)$" {
		t.Fatalf("canonical selector drifted: %s", got)
	}
	if got, want := ExhaustiveNames(), []string{"TestGlobalCommandsMemoryPolicyDoesNotFollowPreviousCWD"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("exhaustive inventory = %v, want %v", got, want)
	}
	if got := PostLandingSelector(); got != "^(TestCLIPathIsolationNarrowingMutant|TestCommandHelpNarrowingMutants|TestGlobalCommandsMemoryPolicyDoesNotFollowPreviousCWD|TestMuseCredentialPolicyNarrowingMutants|TestMuseRootContextNarrowingMutants|TestNativeSkillsTakeoverHelpNarrowingMutant|TestNativeSkillsTakeoverRegistryNarrowingMutant|TestWindowsGitBashSelectionNarrowingMutant)$" {
		t.Fatalf("post-landing selector drifted: %s", got)
	}
	// Callers cannot mutate the canonical list through a returned slice.
	names := Names()
	names[0] = "TestCallerMutation"
	if Names()[0] != want[0] {
		t.Fatal("Names exposed mutable canonical state")
	}
}

// Claim: partitioning conserves inventory and rejects missing, duplicate, or unclassified narrowing tests.
// Limit: the inputs are synthetic and do not verify discovery from cmd/curator.
func TestSplitConservesInventoryAndRejectsUnclassifiedMutants(t *testing.T) {
	all := append(append(Names(), ExhaustiveNames()...), "TestOrdinary", "ExampleOrdinary")
	inventory, err := Split(all)
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.All) != len(inventory.Pull)+len(inventory.Mutations)+len(inventory.Exhaustive) {
		t.Fatalf("inventory is not conserved: all=%d PR=%d mutation=%d exhaustive=%d", len(inventory.All), len(inventory.Pull), len(inventory.Mutations), len(inventory.Exhaustive))
	}
	if len(inventory.Pull) != 2 || len(inventory.Mutations) != 7 || len(inventory.Exhaustive) != 1 {
		t.Fatalf("wrong partition: %+v", inventory)
	}
	selector := regexp.MustCompile(Selector())
	postLandingSelector := regexp.MustCompile(PostLandingSelector())
	for _, name := range inventory.All {
		matched := selector.MatchString(name)
		if matched != contains(inventory.Mutations, name) {
			t.Fatalf("selector partition drift for %s: matched=%v", name, matched)
		}
		matched = postLandingSelector.MatchString(name)
		if matched != (contains(inventory.Mutations, name) || contains(inventory.Exhaustive, name)) {
			t.Fatalf("post-landing selector partition drift for %s: matched=%v", name, matched)
		}
	}
	if err := RequireComplete(inventory); err != nil {
		t.Fatal(err)
	}
	if err := RequirePostLandingComplete(inventory); err != nil {
		t.Fatal(err)
	}

	if _, err := Split(append(all, "TestFutureNarrowingMutant")); err == nil || !strings.Contains(err.Error(), "unclassified narrowing mutant") {
		t.Fatalf("unclassified mutant admitted: %v", err)
	}
	if _, err := Split(append(all, "TestOrdinary")); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate inventory name admitted: %v", err)
	}
	partial, err := Split([]string{"TestOrdinary", Names()[0]})
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireComplete(partial); err == nil {
		t.Fatal("incomplete canonical mutation inventory admitted")
	}
	partialPostLanding, err := Split(append(Names(), "TestOrdinary"))
	if err != nil {
		t.Fatal(err)
	}
	if err := RequirePostLandingComplete(partialPostLanding); err == nil {
		t.Fatal("post-landing inventory without the exhaustive test admitted")
	}
}

// Claim: descriptive mutant-named tests in another package remain in its ordinary partition.
// Limit: this synthetic inventory does not prove actual shard execution or workflow behavior.
func TestSplitOtherPackageKeepsDescriptiveMutantTestsOrdinary(t *testing.T) {
	modeTest := "TestClaudeNativeMemoryModeNarrowingMutants"
	preflightTest := "TestClaudeNativeMemoryPreflightNarrowingMutant"
	all := append(append(Names(), ExhaustiveNames()...), modeTest, preflightTest)
	inventory, err := SplitOtherPackage(all)
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.All) != 10 || len(inventory.Pull) != 2 || len(inventory.Mutations) != 7 || len(inventory.Exhaustive) != 1 {
		t.Fatalf("other-package partition = all:%d PR:%d mutations:%d exhaustive:%d, want 10/2/7/1", len(inventory.All), len(inventory.Pull), len(inventory.Mutations), len(inventory.Exhaustive))
	}
	if !contains(inventory.Pull, modeTest) || !contains(inventory.Pull, preflightTest) {
		t.Fatalf("descriptive mutant tests were not kept in the ordinary set: %v", inventory.Pull)
	}
	if err := RequireComplete(inventory); err != nil {
		t.Fatalf("canonical names were not preserved for placement validation: %v", err)
	}
}

// Claim: ordinary invocations accept the canonical post-landing skip and reject missing, incomplete, or overbroad selectors.
// Limit: this checks argument validation, not whether workflow callers use the validated arguments.
func TestSkipInvocationAndMissingSelectorNegativeControl(t *testing.T) {
	args := append([]string{"test", "-json", "-run", "^TestOrdinary$"}, SkipArgs()...)
	if err := ValidateInvocation(args); err != nil {
		t.Fatalf("canonical ordinary invocation rejected: %v", err)
	}
	// Removing only the selector pair from the ordinary invocation must make
	// the contract check fail, even though its -run selector is unchanged.
	withoutSkip := []string{"test", "-json", "-run", "^TestOrdinary$"}
	if err := ValidateInvocation(withoutSkip); err == nil {
		t.Fatal("ordinary invocation without the post-landing skip was admitted")
	}
	selectorWithoutOneMutation := strings.Replace(PostLandingSelector(), "|TestWindowsGitBashSelectionNarrowingMutant", "", 1)
	if err := ValidateSkipArgs([]string{"-skip", selectorWithoutOneMutation}); err == nil {
		t.Fatal("post-landing skip selector missing one mutation was admitted")
	}
	selectorWithoutExhaustive := strings.Replace(PostLandingSelector(), "|TestGlobalCommandsMemoryPolicyDoesNotFollowPreviousCWD", "", 1)
	if err := ValidateSkipArgs([]string{"-skip", selectorWithoutExhaustive}); err == nil {
		t.Fatal("post-landing skip selector missing the exhaustive test was admitted")
	}
	selectorWithOrdinaryTest := strings.TrimSuffix(PostLandingSelector(), ")$") + "|TestOrdinary)$"
	if err := ValidateSkipArgs([]string{"-skip", selectorWithOrdinaryTest}); err == nil {
		t.Fatal("skip selector including an ordinary test was admitted")
	}
	if err := ValidateSkipArgs([]string{"-skip=" + PostLandingSelector()}); err == nil {
		t.Fatal("combined -skip argument was admitted")
	}
	if err := ValidateSkipArgs([]string{"-skip"}); err == nil {
		t.Fatal("missing -skip value was admitted")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
