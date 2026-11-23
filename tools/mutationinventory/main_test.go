package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/curatormutations"
	"gopkg.in/yaml.v3"
)

// Claim: inventory parsing keeps runnable Go test names and drops tool noise.
// Limit: this checks parser output only; it does not invoke Go or validate test behavior.
func TestParseInventoryKeepsRunnableGoTestNames(t *testing.T) {
	got := parseInventory("TestOne\nExampleOne\nFuzzOne\nBenchmarkOne\nnoise\n")
	want := []string{"ExampleOne", "FuzzOne", "TestOne"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("parsed inventory = %v, want %v", got, want)
	}
}

// Claim: the inventory command admits only the repository's supported CI presets.
// Limit: this does not prove a workflow invokes those presets on the intended runners.
func TestInventoryCommandProfileAllowsOnlySupportedCIPresets(t *testing.T) {
	// Claim: these rows cover accepted package, timeout, and race combinations.
	// Limit: they do not cover arbitrary runner environments or execute the commands.
	tests := []struct {
		name        string
		packagePath string
		timeout     string
		flags       []string
		wantPackage string
		wantTimeout string
		wantFlags   []string
	}{
		{
			name:        "repository path with default timeout",
			packagePath: curatorPackagePath,
			timeout:     inventoryTimeout,
			wantPackage: curatorPackagePath,
			wantTimeout: inventoryTimeout,
		},
		{
			name:        "module path with default timeout",
			packagePath: curatorPackageImportPath,
			timeout:     inventoryTimeout,
			wantPackage: curatorPackageImportPath,
			wantTimeout: inventoryTimeout,
		},
		{
			name:        "Windows timeout",
			packagePath: curatorPackagePath,
			timeout:     windowsInventoryTimeout,
			wantPackage: curatorPackagePath,
			wantTimeout: windowsInventoryTimeout,
		},
		{
			name:        "module path with Windows timeout",
			packagePath: curatorPackageImportPath,
			timeout:     windowsInventoryTimeout,
			wantPackage: curatorPackageImportPath,
			wantTimeout: windowsInventoryTimeout,
		},
		{
			name:        "race lane",
			packagePath: curatorPackagePath,
			timeout:     inventoryTimeout,
			flags:       []string{"-race"},
			wantPackage: curatorPackagePath,
			wantTimeout: inventoryTimeout,
			wantFlags:   []string{"-race"},
		},
		{
			name:        "module path race lane",
			packagePath: curatorPackageImportPath,
			timeout:     inventoryTimeout,
			flags:       []string{"-race"},
			wantPackage: curatorPackageImportPath,
			wantTimeout: inventoryTimeout,
			wantFlags:   []string{"-race"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile, err := newInventoryCommandProfile(test.packagePath, test.timeout, test.flags)
			if err != nil {
				t.Fatalf("supported CI invocation rejected: %v", err)
			}
			if profile.packagePath != test.wantPackage || profile.timeout != test.wantTimeout || !reflect.DeepEqual(profile.goTestFlags(), test.wantFlags) {
				t.Fatalf("profile = %#v, want package=%q timeout=%q flags=%v", profile, test.wantPackage, test.wantTimeout, test.wantFlags)
			}
			// Claim: each preset maps to the ordinary, mutation, or post-landing selector.
			// Limit: argument construction alone does not prove the selected tests pass.
			for _, selection := range []struct {
				value        inventorySelection
				wantSelector string
			}{
				{value: unfilteredInventory, wantSelector: "."},
				{value: mutationInventory, wantSelector: curatormutations.Selector()},
				{value: postLandingInventory, wantSelector: curatormutations.PostLandingSelector()},
			} {
				cmd, err := inventoryCommand(profile, selection.value)
				if err != nil {
					t.Fatalf("supported inventory command rejected: %v", err)
				}
				wantArgs := []string{"go", "test", "-list", selection.wantSelector, "-timeout", test.wantTimeout}
				wantArgs = append(wantArgs, test.wantFlags...)
				wantArgs = append(wantArgs, test.wantPackage)
				if !reflect.DeepEqual(cmd.Args, wantArgs) {
					t.Fatalf("inventory command args = %q, want %q", cmd.Args, wantArgs)
				}
			}
		})
	}
}

// Claim: unsupported packages, timeouts, flags, and selector values are refused.
// Limit: this covers the local allowlist boundary, not every CLI failure mode.
func TestInventoryCommandProfileRejectsCallerControlledValues(t *testing.T) {
	// Claim: each row is outside the supported CI preset allowlist.
	// Limit: the table does not exhaust malformed command-line syntax.
	tests := []struct {
		name        string
		packagePath string
		timeout     string
		flags       []string
	}{
		{name: "unknown package", packagePath: "../../outside", timeout: inventoryTimeout},
		{name: "unknown timeout", packagePath: curatorPackagePath, timeout: "10s"},
		{name: "arbitrary flag", packagePath: curatorPackagePath, timeout: inventoryTimeout, flags: []string{"-coverprofile=/tmp/out"}},
		{name: "combined race flags", packagePath: curatorPackagePath, timeout: inventoryTimeout, flags: []string{"-race", "-count=1"}},
		{name: "race with Windows timeout", packagePath: curatorPackagePath, timeout: windowsInventoryTimeout, flags: []string{"-race"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := newInventoryCommandProfile(test.packagePath, test.timeout, test.flags); err == nil {
				t.Fatalf("caller-controlled inventory values were admitted: package=%q timeout=%q flags=%v", test.packagePath, test.timeout, test.flags)
			}
		})
	}
	profile, err := newInventoryCommandProfile(curatorPackagePath, inventoryTimeout, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inventoryCommand(profile, inventorySelection(99)); err == nil {
		t.Fatal("unknown inventory selector was admitted")
	}
}

// Claim: the launcher version must match the version that built this inventory tool.
// Limit: matching version text does not authenticate the Go executable or its source.
func TestVerifyGoLauncherIdentityRequiresTheToolchainThatBuiltTheInventoryTool(t *testing.T) {
	if err := verifyGoLauncherIdentity("go version go1.25.5 darwin/arm64", "go1.25.5"); err != nil {
		t.Fatalf("matching launcher rejected: %v", err)
	}
	// Claim: a different or unrecognized launcher identity is rejected.
	// Limit: these examples do not test executable replacement after identity checking.
	for _, tc := range []struct {
		name    string
		version string
	}{
		{name: "different toolchain version", version: "go version go1.25.4 darwin/arm64"},
		{name: "unrecognized launcher output", version: "some other toolchain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := verifyGoLauncherIdentity(tc.version, "go1.25.5"); err == nil {
				t.Fatal("different Go launcher identity was admitted")
			}
		})
	}
}

// Claim: the Mac race inventory preserves the PR/mutation/exhaustive partition and exact names.
// Limit: the fixtures do not run deferred tests or establish hosted-runner behavior.
func TestRaceMacInventoryUsesNamedPartitionInvariants(t *testing.T) {
	profile, err := newInventoryCommandProfile(curatorPackagePath, inventoryTimeout, []string{"-race"})
	if err != nil {
		t.Fatal(err)
	}
	mutationNames := curatormutations.Names()
	exhaustiveNames := curatormutations.ExhaustiveNames()
	ordinaryNames := []string{"TestOrdinaryOne", "TestOrdinaryTwo"}
	inventory := inventoryWithPartition(ordinaryNames, mutationNames, exhaustiveNames)
	if err := validateMacRaceInventory("darwin", profile, inventory); err != nil {
		t.Fatalf("complete Darwin race inventory rejected: %v", err)
	}
	if err := validateInventoryPartition(inventory); err != nil {
		t.Fatalf("complete inventory partition rejected: %v", err)
	}
	if err := curatormutations.RequirePostLandingComplete(inventory); err != nil {
		t.Fatalf("complete post-landing inventory rejected: %v", err)
	}

	addedOrdinary := append(append([]string(nil), inventory.All...), "TestAddedOrdinary")
	withAddedOrdinary, err := curatormutations.Split(addedOrdinary)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateMacRaceInventory("darwin", profile, withAddedOrdinary); err != nil {
		t.Fatalf("an added ordinary test was rejected: %v", err)
	}

	lostOrdinary := inventory
	lostOrdinary.Pull = append([]string(nil), inventory.Pull[:1]...)
	if err := validateMacRaceInventory("darwin", profile, lostOrdinary); err == nil {
		t.Fatal("an ordinary test missing from PR ∪ M ∪ exhaustive was admitted")
	}

	duplicateAcross := inventory
	duplicateAcross.Mutations = append(append([]string(nil), inventory.Mutations...), inventory.Pull[0])
	if err := validateMacRaceInventory("darwin", profile, duplicateAcross); err == nil {
		t.Fatal("a test duplicated across PR and M was admitted")
	}

	eighthMutation := inventory
	eighthMutation.All = append(append([]string(nil), inventory.All...), "TestUnexpectedAdditionalMutation")
	eighthMutation.Mutations = append(append([]string(nil), inventory.Mutations...), "TestUnexpectedAdditionalMutation")
	if err := validateInventoryPartition(eighthMutation); err != nil {
		t.Fatalf("the eighth-name fixture is not a conserved partition: %v", err)
	}
	if err := validateMacRaceInventory("darwin", profile, eighthMutation); err == nil {
		t.Fatal("an eighth mutation name was admitted")
	}

	missingMutation := inventoryWithPartition(ordinaryNames, mutationNames[:len(mutationNames)-1], exhaustiveNames)
	if err := validateInventoryPartition(missingMutation); err != nil {
		t.Fatalf("the missing-name fixture is not a conserved partition: %v", err)
	}
	if err := validateMacRaceInventory("darwin", profile, missingMutation); err == nil {
		t.Fatal("a missing canonical mutation name was admitted")
	}

	renamedMutationNames := append([]string(nil), mutationNames...)
	renamedMutationNames[0] = "TestRenamedMutation"
	renamedMutation := inventoryWithPartition(ordinaryNames, renamedMutationNames, exhaustiveNames)
	if err := validateInventoryPartition(renamedMutation); err != nil {
		t.Fatalf("the renamed-name fixture is not a conserved partition: %v", err)
	}
	if err := validateMacRaceInventory("darwin", profile, renamedMutation); err == nil {
		t.Fatal("a renamed mutation name was admitted")
	}
	missingExhaustive := inventoryWithPartition(ordinaryNames, mutationNames, nil)
	if err := validateInventoryPartition(missingExhaustive); err != nil {
		t.Fatalf("missing-exhaustive fixture is not a conserved partition: %v", err)
	}
	if err := validateMacRaceInventory("darwin", profile, missingExhaustive); err == nil {
		t.Fatal("post-landing inventory without the exhaustive product-behavior test was admitted")
	}

	if err := validateMacRaceInventory("linux", profile, missingMutation); err != nil {
		t.Fatalf("non-macOS invocation received the Mac mutation inventory policy: %v", err)
	}
	if _, err := newInventoryCommandProfile(curatorPackagePath, inventoryTimeout, []string{"-race", "-tags=curator_test_platform"}); err == nil {
		t.Fatal("caller-controlled race flags were admitted")
	}
}

func inventoryWithPartition(ordinary, mutations, exhaustive []string) curatormutations.Inventory {
	all := append(append(append([]string(nil), ordinary...), mutations...), exhaustive...)
	return curatormutations.Inventory{
		All:        all,
		Pull:       append([]string(nil), ordinary...),
		Mutations:  append([]string(nil), mutations...),
		Exhaustive: append([]string(nil), exhaustive...),
	}
}

// Claim: ordinary Go test invocations must carry the canonical post-landing skip.
// Limit: this does not prove the main-push workflow actually runs the skipped tests.
func TestOrdinaryGoTestInvocationRejectsMissingPostLandingSkip(t *testing.T) {
	base := []string{"-mode", "verify-invocation", "--", "env", "CURATOR_CONFORMANCE_ROOT=/tmp/root", "go", "test", "-json", "-run", "^TestOrdinary$"}
	withSkip := append(append([]string(nil), base...), curatormutations.SkipArgs()...)
	withSkip = append(withSkip, "./cmd/curator")
	if err := run(withSkip); err != nil {
		t.Fatalf("ordinary invocation with canonical skip rejected: %v", err)
	}
	// This is the ordinary-lane negative control: remove just the skip pair
	// from the same invocation while leaving its package and run selector intact.
	withoutSkip := removePair(withSkip, "-skip")
	if err := run(withoutSkip); err == nil {
		t.Fatal("ordinary invocation without the post-landing selector was admitted")
	}
}

func removePair(args []string, flag string) []string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return append(append([]string(nil), args[:i]...), args[i+2:]...)
		}
	}
	return append([]string(nil), args...)
}

// Claim: a mutation selector must contain exactly the canonical mutation inventory.
// Limit: this does not prove workflow wiring or the behavior of selected tests.
func TestMutationRunInventoryMustMatchCanonicalPartition(t *testing.T) {
	all := append(append(curatormutations.Names(), curatormutations.ExhaustiveNames()...), "TestOrdinary")
	inventory, err := curatormutations.Split(all)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyMutationSelection(inventory.Mutations, curatormutations.Names()); err != nil {
		t.Fatalf("canonical -run set rejected: %v", err)
	}
	postLanding := append(append([]string(nil), inventory.Mutations...), inventory.Exhaustive...)
	if err := verifyPostLandingSelection(postLanding, curatormutations.PostLandingNames()); err != nil {
		t.Fatalf("complete post-landing -run set rejected: %v", err)
	}
	missing := curatormutations.Names()[:len(curatormutations.Names())-1]
	if err := verifyMutationSelection(inventory.Mutations, missing); err == nil {
		t.Fatal("mutation selector missing a canonical test was admitted")
	}
	if err := verifyMutationSelection(inventory.Mutations, append(curatormutations.Names(), "TestOrdinary")); err == nil {
		t.Fatal("mutation selector including an ordinary test was admitted")
	}
	missingExhaustive := append([]string(nil), inventory.Mutations...)
	if err := verifyPostLandingSelection(postLanding, missingExhaustive); err == nil {
		t.Fatal("post-landing selector missing the exhaustive test was admitted")
	}
	if err := verifyPostLandingSelection(postLanding, append(postLanding, "TestOrdinary")); err == nil {
		t.Fatal("post-landing selector including an ordinary test was admitted")
	}
}

// Claim: every ordinary test-gate invocation explicitly skips deferred tests.
// Limit: this does not establish that the main-push lane runs them.
func TestEveryCITestGateInvocationSkipsDeferredSet(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil || len(parsed) == 0 {
		t.Fatalf("ci.yml is not valid YAML: %v", err)
	}
	var blocks []string
	var current strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "      - ") {
			if current.Len() != 0 {
				blocks = append(blocks, current.String())
				current.Reset()
			}
		}
		if current.Len() != 0 || strings.HasPrefix(line, "      - ") {
			current.WriteString(line)
			current.WriteByte('\n')
		}
	}
	if current.Len() != 0 {
		blocks = append(blocks, current.String())
	}

	count := 0
	for _, block := range blocks {
		if !strings.Contains(block, ".github/ci/test-gate.sh") {
			continue
		}
		count++
		if !strings.Contains(block, `CI_SKIP_POST_LANDING_TESTS: "1"`) {
			t.Errorf("test-gate workflow step does not skip the post-landing set:\n%s", block)
		}
		if strings.Contains(block, "GO_TEST_FLAGS: -race -skip") || strings.Contains(block, "GO_TEST_FLAGS: '-race -skip") {
			t.Errorf("post-landing skip was folded into GO_TEST_FLAGS:\n%s", block)
		}
	}
	if count != 5 {
		t.Fatalf("found %d test-gate steps, want five", count)
	}
}

// Claim: the test-gate compatibility flag expression remains one safe Bash argument.
// Limit: this checks the expression and local Bash behavior, not a complete CI run.
func TestTestGateStoreCompatFlagsExpansionIsNounsetSafe(t *testing.T) {
	const safeExpansion = `"${STORE_COMPAT_FLAGS[*]+"${STORE_COMPAT_FLAGS[*]}"}"`

	data, err := os.ReadFile("../../.github/ci/test-gate.sh")
	if err != nil {
		t.Fatal(err)
	}
	var flagLines []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || !strings.Contains(trimmed, "-extra-go-test-flags") {
			continue
		}
		flagLines = append(flagLines, trimmed)
	}
	wantFlagLine := "-extra-go-test-flags " + safeExpansion + " -evidence \"$EVIDENCE/post-landing-inventory/$stage\")"
	if len(flagLines) != 1 || flagLines[0] != wantFlagLine {
		t.Fatalf("active test-gate store flags argument lines = %q, want exactly %q", flagLines, wantFlagLine)
	}

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("bash is unavailable for the expansion behavior check: %v", err)
	}

	// Claim: empty and populated compatibility values retain the expected argument shape.
	// Limit: the examples do not cover every Bash release or environment configuration.
	for _, tc := range []struct {
		name       string
		assignment string
		want       string
	}{
		{
			name:       "empty compatibility flags retain the required empty argument",
			assignment: "STORE_COMPAT_FLAGS=()",
			want:       "count=4\nflag=<>\n",
		},
		{
			name:       "populated compatibility flags remain one argument",
			assignment: "STORE_COMPAT_FLAGS=(-race -count=1)",
			want:       "count=4\nflag=<-race -count=1>\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := "set -u\n" + tc.assignment + "\nargs=(-extra-go-test-flags " + safeExpansion + " -evidence evidence)\nprintf 'count=%d\\nflag=<%s>\\n' \"${#args[@]}\" \"${args[1]}\"\n"
			got, err := exec.Command(bash, "-c", script).CombinedOutput()
			if err != nil {
				t.Fatalf("nounset-safe expansion failed: %v\n%s", err, got)
			}
			if string(got) != tc.want {
				t.Fatalf("argument shape = %q, want %q", got, tc.want)
			}
		})
	}
	// The source check and argument-shape cases above cover the contract. Bash
	// releases differ on whether the legacy empty-array form fails under nounset.
}

func normalizeLineEndings(data []byte) string {
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// Claim: LF and CRLF workflow text normalize to the same LF form.
// Limit: this does not parse YAML or check workflow semantics.
func TestNormalizeLineEndingsAcceptsLFAndCRLF(t *testing.T) {
	// Claim: both line-ending forms are accepted by the normalizer.
	// Limit: the table covers only these two byte sequences.
	for _, tc := range []struct {
		name string
		data string
	}{
		{name: "LF", data: "on:\n  push:\n"},
		{name: "CRLF", data: "on:\r\n  push:\r\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeLineEndings([]byte(tc.data)); got != "on:\n  push:\n" {
				t.Fatalf("normalized workflow = %q, want LF form", got)
			}
		})
	}
}

func mainPushOnlyTrigger(workflow map[string]any) bool {
	triggers, ok := workflow["on"].(map[string]any)
	if !ok || len(triggers) != 1 {
		return false
	}
	push, ok := triggers["push"].(map[string]any)
	if !ok || len(push) != 1 {
		return false
	}
	branches, ok := push["branches"].([]any)
	return ok && len(branches) == 1 && branches[0] == "main"
}

// Claim: only a push to main satisfies the post-landing trigger predicate.
// Limit: this does not validate jobs or triggers outside the modeled mapping.
func TestMainPushOnlyTriggerRejectsOtherEventsAndBranches(t *testing.T) {
	// Claim: the table covers main push, another branch, an extra event, and malformed input.
	// Limit: it does not enumerate every valid GitHub Actions trigger form.
	for _, tc := range []struct {
		name     string
		workflow map[string]any
		want     bool
	}{
		{
			name: "main push only",
			workflow: map[string]any{"on": map[string]any{
				"push": map[string]any{"branches": []any{"main"}},
			}},
			want: true,
		},
		{
			name: "other branch",
			workflow: map[string]any{"on": map[string]any{
				"push": map[string]any{"branches": []any{"release"}},
			}},
		},
		{
			name: "additional event",
			workflow: map[string]any{"on": map[string]any{
				"push":         map[string]any{"branches": []any{"main"}},
				"pull_request": map[string]any{},
			}},
		},
		{
			name:     "trigger text is not a mapping",
			workflow: map[string]any{"on": "push to main"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mainPushOnlyTrigger(tc.workflow); got != tc.want {
				t.Fatalf("main-push-only trigger = %t, want %t", got, tc.want)
			}
		})
	}
}

// Claim: the post-landing workflow inventories and runs every deferred test on main.
// Limit: this validates workflow text and YAML, not a completed hosted run.
func TestPostLandingWorkflowIsMainPushOnlyAndRunsCompleteDeferredSet(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/post-landing-mutation-tests.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := validatePostLandingInvocationGates(data); err != nil {
		t.Fatalf("valid post-landing invocation rejected: %v", err)
	}
	workflow := normalizeLineEndings(data)
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil || len(parsed) == 0 {
		t.Fatalf("post-landing workflow is not valid YAML: %v", err)
	}
	if !mainPushOnlyTrigger(parsed) {
		t.Fatal("post-landing workflow must trigger only on pushes to main")
	}
	jobs, ok := parsed["jobs"].(map[string]any)
	if !ok {
		t.Fatal("post-landing workflow has no jobs mapping")
	}
	job, ok := jobs["curator-deferred-tests"].(map[string]any)
	if !ok || job["runs-on"] != "macos-latest" {
		t.Fatalf("post-landing job is missing or uses the wrong runner: %#v", job)
	}
	if _, ok := job["continue-on-error"]; ok {
		t.Fatal("post-landing test verdict is allowed to fail")
	}
	for _, required := range []string{
		"-mode inventory -package ./cmd/curator -timeout 60m",
		"-go-test-flags=-race -require-all",
		"selector=\"$(go run ./tools/mutationinventory -mode post-landing-selector)\"",
		"go test -race -count=1 -timeout=60m -v",
		"-run \"$selector\" ./cmd/curator",
		"go-test.txt",
		"name: post-landing-mutation-evidence",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("post-landing workflow is missing %q", required)
		}
	}
	if strings.Contains(workflow, "continue-on-error:") {
		t.Fatal("post-landing verdict is not required")
	}
}

// Claim: the post-landing test invocation refuses skip flags and continue-on-error at job or step scope.
// Limit: this validates workflow configuration and planted variants, not GitHub Actions execution semantics.
func TestPostLandingWorkflowRejectsSkipAndContinueOnErrorPlants(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/post-landing-mutation-tests.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := validatePostLandingInvocationGates(data); err != nil {
		t.Fatalf("positive control: valid workflow rejected: %v", err)
	}

	plants := []struct {
		name, before, after string
	}{
		{"separate_skip", "go test -race", "go test -skip TestOrdinary -race"},
		{"equals_skip", "go test -race", "go test -skip=TestOrdinary -race"},
		{
			"job_continue_on_error",
			"  curator-deferred-tests:\n    name:",
			"  curator-deferred-tests:\n    continue-on-error: true\n    name:",
		},
		{
			"step_continue_on_error",
			"      - name: Run post-landing mutation and exhaustive tests\n        shell: bash",
			"      - name: Run post-landing mutation and exhaustive tests\n        continue-on-error: true\n        shell: bash",
		},
	}

	lf := normalizeLineEndings(data)
	variants := []struct {
		name string
		data []byte
	}{
		{name: "lf", data: []byte(lf)},
		{name: "crlf", data: []byte(strings.ReplaceAll(lf, "\n", "\r\n"))},
	}
	for _, variant := range variants {
		variant := variant
		t.Run(variant.name, func(t *testing.T) {
			if err := validatePostLandingInvocationGates(variant.data); err != nil {
				t.Fatalf("positive control: valid %s workflow rejected: %v", variant.name, err)
			}
			for _, plant := range plants {
				plant := plant
				t.Run(plant.name, func(t *testing.T) {
					normalized := normalizeLineEndings(variant.data)
					planted := strings.Replace(normalized, plant.before, plant.after, 1)
					if planted == normalized {
						t.Fatal("workflow plant was not applied")
					}
					if strings.Contains(string(variant.data), "\r\n") {
						planted = strings.ReplaceAll(planted, "\n", "\r\n")
					}
					if err := validatePostLandingInvocationGates([]byte(planted)); err == nil {
						t.Fatal("forbidden post-landing invocation plant was admitted")
					}
				})
			}
		})
	}
}

func validatePostLandingInvocationGates(data []byte) error {
	var workflow struct {
		Jobs map[string]struct {
			ContinueOnError *bool `yaml:"continue-on-error"`
			Steps           []struct {
				ContinueOnError *bool  `yaml:"continue-on-error"`
				Run             string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return err
	}
	job, ok := workflow.Jobs["curator-deferred-tests"]
	if !ok {
		return fmt.Errorf("post-landing workflow has no deferred-test job")
	}
	if job.ContinueOnError != nil {
		return fmt.Errorf("post-landing job allows continue-on-error")
	}
	testInvocationFound := false
	for _, step := range job.Steps {
		if step.ContinueOnError != nil {
			return fmt.Errorf("post-landing step allows continue-on-error")
		}
		args := strings.Fields(step.Run)
		for i := 0; i+1 < len(args); i++ {
			if args[i] != "go" || args[i+1] != "test" {
				continue
			}
			testInvocationFound = true
			for _, arg := range args[i+2:] {
				if arg == "-skip" || arg == "--skip" || strings.HasPrefix(arg, "-skip=") || strings.HasPrefix(arg, "--skip=") {
					return fmt.Errorf("post-landing go test invocation skips deferred tests")
				}
			}
		}
	}
	if !testInvocationFound {
		return fmt.Errorf("post-landing workflow has no go test invocation")
	}
	return nil
}

func checkMacOSRaceCapacityStep(data []byte) error {
	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return err
	}
	jobs, ok := parsed["jobs"].(map[string]any)
	if !ok {
		return fmt.Errorf("ci.yml has no jobs mapping")
	}
	race, ok := jobs["race"].(map[string]any)
	if !ok {
		return fmt.Errorf("ci.yml has no Race job")
	}
	steps, ok := race["steps"].([]any)
	if !ok {
		return fmt.Errorf("Race job has no steps")
	}
	found := 0
	for _, value := range steps {
		step, ok := value.(map[string]any)
		if !ok || step["name"] != "Record macOS race capacity" {
			continue
		}
		found++
		if step["if"] != "runner.os == 'macOS'" {
			return fmt.Errorf("capacity step is not limited to the macOS runner")
		}
		run, ok := step["run"].(string)
		if !ok {
			return fmt.Errorf("capacity step has no shell script")
		}
		for _, required := range []string{
			"fmt.Println(runtime.NumCPU())",
			"fmt.Println(runtime.GOMAXPROCS(0))",
			"GOTOOLCHAIN=local GOENV=off go run \"$RUNNER_TEMP/go-runner-capacity.go\"",
			"tee .temp/ci-evidence/race/runner-capacity.txt",
		} {
			if !strings.Contains(run, required) {
				return fmt.Errorf("capacity step is missing %q", required)
			}
		}
		if strings.Count(run, "fmt.Println(") != 2 || strings.Contains(run, "fmt.Printf(") {
			return fmt.Errorf("capacity step prints data beyond the two requested numeric values")
		}
	}
	if found != 1 {
		return fmt.Errorf("found %d macOS race capacity steps, want exactly one", found)
	}
	return nil
}

// Claim: the PR macOS Race job logs CPU count and effective GOMAXPROCS as two numeric lines.
// Limit: this checks workflow source, not the hosted runner's emitted values.
func TestMacOSRaceJobLogsRunnerCapacity(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkMacOSRaceCapacityStep(data); err != nil {
		t.Fatalf("valid macOS capacity step rejected: %v", err)
	}
	for _, test := range []struct {
		name  string
		plant func([]byte) ([]byte, error)
	}{
		{
			name: "missing_cpu_count",
			plant: func(workflow []byte) ([]byte, error) {
				mutant := strings.Replace(string(workflow), "fmt.Println(runtime.NumCPU())", "fmt.Println(runtime.NumGoroutine())", 1)
				if mutant == string(workflow) {
					return nil, fmt.Errorf("CPU count statement not found")
				}
				return []byte(mutant), nil
			},
		},
		{
			name: "missing_effective_gomaxprocs",
			plant: func(workflow []byte) ([]byte, error) {
				mutant := strings.Replace(string(workflow), "fmt.Println(runtime.GOMAXPROCS(0))", "fmt.Println(runtime.NumCPU())", 1)
				if mutant == string(workflow) {
					return nil, fmt.Errorf("GOMAXPROCS statement not found")
				}
				return []byte(mutant), nil
			},
		},
		{
			name: "wrong_runner_condition",
			plant: func(workflow []byte) ([]byte, error) {
				return setWorkflowStepCondition(workflow, "race", "Record macOS race capacity", "runner.os == 'Linux'")
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutant, err := test.plant(data)
			if err != nil {
				t.Fatalf("capacity-step plant not applied: %v", err)
			}
			if err := checkMacOSRaceCapacityStep(mutant); err == nil {
				t.Fatal("invalid macOS capacity step admitted")
			}
		})
	}
	t.Run("wrong_runner_condition_on_crlf", func(t *testing.T) {
		crlf := []byte(strings.ReplaceAll(normalizeLineEndings(data), "\n", "\r\n"))
		if err := checkMacOSRaceCapacityStep(crlf); err != nil {
			t.Fatalf("valid CRLF workflow rejected: %v", err)
		}
		mutant, err := setWorkflowStepCondition(crlf, "race", "Record macOS race capacity", "runner.os == 'Linux'")
		if err != nil {
			t.Fatalf("CRLF condition plant failed: %v", err)
		}
		condition, err := workflowStepConditionValue(mutant, "race", "Record macOS race capacity")
		if err != nil {
			t.Fatal(err)
		}
		if condition != "runner.os == 'Linux'" {
			t.Fatalf("CRLF plant condition = %q, want Linux", condition)
		}
		if err := checkMacOSRaceCapacityStep(mutant); err == nil {
			t.Fatal("wrong runner condition admitted after CRLF plant")
		}
	})
}

type pullRequestStep struct {
	Name string    `yaml:"name"`
	If   string    `yaml:"if"`
	Run  string    `yaml:"run"`
	Uses string    `yaml:"uses"`
	With yaml.Node `yaml:"with"`
	Env  yaml.Node `yaml:"env"`
}

type pullRequestJob struct {
	Name     string            `yaml:"name"`
	If       string            `yaml:"if"`
	Uses     string            `yaml:"uses"`
	Needs    yaml.Node         `yaml:"needs"`
	Env      yaml.Node         `yaml:"env"`
	Strategy yaml.Node         `yaml:"strategy"`
	Steps    []pullRequestStep `yaml:"steps"`
}

type pullRequestWorkflow struct {
	On   yaml.Node                 `yaml:"on"`
	Env  yaml.Node                 `yaml:"env"`
	Jobs map[string]pullRequestJob `yaml:"jobs"`
}

const legacyMutationJobID = "curator-mutation-pr"

var mutationJobResultReference = regexp.MustCompile(`(?i)\bneeds\s*(?:\.\s*([A-Za-z_][A-Za-z0-9_-]*)|\[\s*['"]([^'"]+)['"]\s*\])\s*\.\s*result\b`)
var dynamicNeedsResultReference = regexp.MustCompile(`(?i)\bneeds\s*\[\s*[^'"][^\]]+\]\s*\.\s*result\b`)
var githubExpression = regexp.MustCompile(`\$\{\{(?s:(.*?))\}\}`)
var simpleMatrixExpression = regexp.MustCompile(`^matrix\s*\.\s*([A-Za-z_][A-Za-z0-9_-]*)$`)
var workspaceEnvironmentPath = regexp.MustCompile(`^\$\{\{\s*github\.workspace\s*\}\}(?:/[A-Za-z0-9._/-]*)?$`)
var pinnedPackageVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// The current PR workflow uses no job-level reusable workflows. Keep this
// exact-target allowlist empty until a reviewed target is deliberately admitted.
var allowedPRReusableWorkflowTargets = map[string]struct{}{}

// These are the complete static matrix values the PR workflow currently passes
// into run/env: Go versions, runner labels, and shard indexes.
var allowedPRMatrixValues = map[string]map[string]struct{}{
	"go": {
		"1.25.5": {}, "1.26.0": {}, "1.27.1": {},
	},
	"os": {
		"ubuntu-latest": {}, "macos-latest": {}, "windows-latest": {},
	},
	"shard": {
		"0": {}, "1": {}, "2": {}, "3": {}, "4": {}, "5": {}, "6": {}, "7": {}, "8": {},
	},
}

func checkPRDefersPostLandingTests(data []byte) error {
	_, root, err := parseWorkflowDocument(data)
	if err != nil {
		return err
	}
	if err := validateWorkflowMappingKeys(root); err != nil {
		return err
	}
	var workflow pullRequestWorkflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return err
	}
	if err := requirePullRequestTrigger(&workflow.On); err != nil {
		return err
	}
	if workflow.Jobs == nil {
		return fmt.Errorf("workflow has no jobs mapping")
	}
	testJob, ok := workflow.Jobs["test"]
	if !ok {
		return fmt.Errorf("test job is missing")
	}
	wantNeeds := []string{"windows-test-shards", "unix-test-shards", "store-v1-compatibility"}
	needs, err := pullRequestNeedIDs(&testJob.Needs)
	if err != nil || testJob.Needs.Kind != yaml.SequenceNode || !reflect.DeepEqual(needs, wantNeeds) {
		return fmt.Errorf("test job has unexpected dependencies: %v", needs)
	}

	mutationJobIDs := map[string]struct{}{legacyMutationJobID: {}}
	for id, job := range workflow.Jobs {
		if isMutationSuiteJob(id, job) {
			mutationJobIDs[strings.ToLower(id)] = struct{}{}
		}
	}
	for id, job := range workflow.Jobs {
		if !canRunOnPullRequest(job.If) {
			continue
		}
		if uses := strings.TrimSpace(job.Uses); uses != "" {
			if _, ok := allowedPRReusableWorkflowTargets[uses]; !ok {
				return fmt.Errorf("pull-request job %q calls an unapproved reusable workflow %q", id, uses)
			}
		}
		if err := checkPRDeferredTestCommands(id, job, &workflow.Env); err != nil {
			return err
		}
		if err := checkPRLocalTargets(id, job, &workflow.Env); err != nil {
			return err
		}
		if err := checkPRDynamicInputs(id, job, &workflow.Env); err != nil {
			return err
		}
		if isMutationSuiteJob(id, job) {
			return fmt.Errorf("mutation suite job %q can run on pull-request CI", id)
		}
		needs, err := pullRequestNeedIDs(&job.Needs)
		if err != nil {
			return fmt.Errorf("job %q has unreadable needs: %w", id, err)
		}
		for _, dependency := range needs {
			if isMutationJobID(dependency, mutationJobIDs) {
				return fmt.Errorf("pull-request job %q depends on deferred mutation job %q", id, dependency)
			}
		}
		if hasMutationResultReference(job.If, mutationJobIDs) || hasMutationResultReferenceInNode(&job.Needs, mutationJobIDs) || hasMutationResultReferenceInNode(&job.Env, mutationJobIDs) {
			return fmt.Errorf("pull-request job %q consumes a deferred mutation result", id)
		}
		for _, step := range job.Steps {
			if hasMutationResultReference(step.If, mutationJobIDs) || hasMutationResultReference(step.Run, mutationJobIDs) || hasMutationResultReferenceInNode(&step.Env, mutationJobIDs) {
				return fmt.Errorf("pull-request job %q consumes a deferred mutation result", id)
			}
		}
	}
	return nil
}

func requirePullRequestTrigger(on *yaml.Node) error {
	on = resolveWorkflowAlias(on)
	if on == nil {
		return fmt.Errorf("workflow has no trigger")
	}
	switch on.Kind {
	case yaml.ScalarNode:
		if on.Tag == "!!str" && on.Value == "pull_request" {
			return nil
		}
		return fmt.Errorf("workflow scalar trigger has no pull_request event")
	case yaml.SequenceNode:
		seen := make(map[string]struct{}, len(on.Content))
		found := false
		for _, event := range on.Content {
			if event.Kind != yaml.ScalarNode || event.Tag != "!!str" {
				return fmt.Errorf("workflow trigger has an unreadable event")
			}
			if _, ok := seen[event.Value]; ok {
				return fmt.Errorf("workflow trigger repeats event %q", event.Value)
			}
			seen[event.Value] = struct{}{}
			if event.Value == "pull_request" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("workflow has no pull_request trigger")
		}
		return nil
	case yaml.MappingNode:
		// Continue below with the event-filter mapping form.
	default:
		return fmt.Errorf("workflow trigger must be a scalar, sequence, or mapping")
	}
	seen := make(map[string]struct{}, len(on.Content)/2)
	found := false
	for i := 0; i+1 < len(on.Content); i += 2 {
		key := on.Content[i]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
			return fmt.Errorf("workflow trigger has an unreadable event key")
		}
		if _, ok := seen[key.Value]; ok {
			return fmt.Errorf("workflow trigger repeats event %q", key.Value)
		}
		seen[key.Value] = struct{}{}
		if key.Value != "pull_request" {
			continue
		}
		event := resolveWorkflowAlias(on.Content[i+1])
		if event == nil || (event.Kind != yaml.MappingNode && (event.Kind != yaml.ScalarNode || event.Tag != "!!null")) {
			return fmt.Errorf("pull_request trigger has an unsupported value")
		}
		found = true
	}
	if !found {
		return fmt.Errorf("workflow has no pull_request trigger")
	}
	return nil
}

func resolveWorkflowAlias(node *yaml.Node) *yaml.Node {
	for node != nil && node.Kind == yaml.AliasNode {
		if node.Alias == nil {
			return nil
		}
		node = node.Alias
	}
	return node
}

func validateWorkflowMappingKeys(root *yaml.Node) error {
	seenNodes := make(map[*yaml.Node]struct{})
	activeNodes := make(map[*yaml.Node]struct{})
	var visit func(*yaml.Node) error
	visit = func(node *yaml.Node) error {
		if node == nil {
			return nil
		}
		if node.Kind == yaml.AliasNode {
			if node.Alias == nil {
				return fmt.Errorf("workflow has an unresolved YAML alias")
			}
			node = node.Alias
		}
		if _, ok := activeNodes[node]; ok {
			return fmt.Errorf("workflow has a cyclic YAML alias")
		}
		if _, ok := seenNodes[node]; ok {
			return nil
		}
		activeNodes[node] = struct{}{}
		defer delete(activeNodes, node)
		if node.Kind == yaml.MappingNode {
			if len(node.Content)%2 != 0 {
				return fmt.Errorf("workflow has an incomplete mapping")
			}
			keys := make(map[string]struct{}, len(node.Content)/2)
			for i := 0; i+1 < len(node.Content); i += 2 {
				key := node.Content[i]
				if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
					return fmt.Errorf("workflow has a non-string mapping key")
				}
				if _, ok := keys[key.Value]; ok {
					return fmt.Errorf("workflow mapping repeats key %q", key.Value)
				}
				keys[key.Value] = struct{}{}
				if err := visit(node.Content[i+1]); err != nil {
					return err
				}
			}
		} else {
			for _, child := range node.Content {
				if err := visit(child); err != nil {
					return err
				}
			}
		}
		seenNodes[node] = struct{}{}
		return nil
	}
	return visit(root)
}

func checkPRDynamicInputs(jobID string, job pullRequestJob, workflowEnv *yaml.Node) error {
	matrixValues, err := pullRequestMatrixValues(&job.Strategy)
	if err != nil {
		return fmt.Errorf("pull-request job %q has an unclassifiable matrix: %w", jobID, err)
	}
	checkText := func(text, location string) error {
		if strings.Count(text, "${{") != strings.Count(text, "}}") {
			return fmt.Errorf("pull-request job %q has an unbalanced dynamic input in %s", jobID, location)
		}
		for _, match := range githubExpression.FindAllStringSubmatch(text, -1) {
			expression := strings.TrimSpace(match[1])
			if strings.Contains(strings.ToLower(expression), "matrix") {
				fieldMatch := simpleMatrixExpression.FindStringSubmatch(expression)
				if len(fieldMatch) != 2 {
					return fmt.Errorf("pull-request job %q has an unclassifiable matrix input %q in %s", jobID, expression, location)
				}
				if err := validatePRMatrixField(fieldMatch[1], matrixValues); err != nil {
					return fmt.Errorf("pull-request job %q: %w", jobID, err)
				}
				continue
			}
			switch expression {
			case "github.workspace":
				if location != "env" || !workspaceEnvironmentPath.MatchString(text) {
					return fmt.Errorf("pull-request job %q has an unclassifiable workspace input in %s", jobID, location)
				}
			case "env.PNPM_PIN":
				pin, ok := workflowStringValue(workflowEnv, "PNPM_PIN")
				if location != "run" || !ok || !pinnedPackageVersion.MatchString(pin) {
					return fmt.Errorf("pull-request job %q has an unclassifiable package-pin input in %s", jobID, location)
				}
			case "runner.os == 'Windows' && '120m' || '60m'":
				if location != "env" || strings.TrimSpace(text) != match[0] {
					return fmt.Errorf("pull-request job %q has an unclassifiable runner input in %s", jobID, location)
				}
			default:
				return fmt.Errorf("pull-request job %q has an unclassifiable dynamic input %q in %s", jobID, expression, location)
			}
		}
		return nil
	}
	for _, value := range workflowScalarValues(&job.Env) {
		if err := checkText(value, "env"); err != nil {
			return err
		}
	}
	for _, step := range job.Steps {
		if err := checkText(step.Run, "run"); err != nil {
			return err
		}
		for _, value := range workflowScalarValues(&step.Env) {
			if err := checkText(value, "env"); err != nil {
				return err
			}
		}
	}
	return nil
}

type workflowShellWord struct {
	value    string
	quote    byte
	operator bool
}

var workflowShellVariable = regexp.MustCompile(`^\$(?:\{([A-Za-z_][A-Za-z0-9_]*)\}|([A-Za-z_][A-Za-z0-9_]*))$`)
var localWorkflowScriptInvocation = regexp.MustCompile(`(?i)\b(?:bash|sh|python3?|node|pwsh|powershell)\s+["']?((?:\./|\.\./|\.github/|[A-Za-z0-9_.-]+/|/)[^"' \t\r\n;&|]+)`)
var localWorkflowScriptToken = regexp.MustCompile(`(?:^|[\s;&|])((?:\./|\.\./|\.github/|[A-Za-z0-9_.-]+/|/)[^"' \t\r\n;&|]+\.(?:sh|bash|zsh|py|js|ps1|cmd|bat|rb|pl))(?:$|[\s;&|])`)
var localWorkflowExecutable = regexp.MustCompile(`(?m)(?:^|[;&|]\s*)((?:\./|\.\./|[A-Za-z0-9_.-]+/)[^\s;&|]+)(?:$|[\s;&|])`)
var dynamicWorkflowScriptInvocation = regexp.MustCompile(`(?i)\b(?:bash|sh|python3?|node|pwsh|powershell)\s+["']?\$(?:\{|[A-Za-z_(])`)
var dynamicWorkflowSourceInvocation = regexp.MustCompile(`(?im)(?:^|[;&|]\s*)(?:source|\.)\s+["']?\$(?:\{|[A-Za-z_(])`)
var inlineWorkflowScriptBody = regexp.MustCompile(`(?i)\b(?:bash|sh|zsh)\s+-[A-Za-z]*c(?:\s|$)|\bpython3?\s+-c(?:\s|$)|\bnode\s+-e(?:\s|$)|\beval(?:\s|$)`)
var inlineWorkflowHeredoc = regexp.MustCompile(`(?im)^\s*(?:python3?|bash|sh|zsh|node|pwsh|powershell)\s+[^\n]*<<-?\s*["']?[A-Za-z_][A-Za-z0-9_]*["']?\s*$`)
var testGateSkipOverride = regexp.MustCompile(`(?i)(?:^|[\s;&|])CI_SKIP_POST_LANDING_TESTS=([^\s;&|]+)`)
var testGateSkipUnset = regexp.MustCompile(`(?i)(?:^|[\s;&|])(?:-u\s+|unset\s+)CI_SKIP_POST_LANDING_TESTS(?:[\s;&|]|$)`)
var isolatedWorkflowEnvironment = regexp.MustCompile(`(?i)\benv\s+-i(?:\s|$)`)

// Local scripts are admitted by exact path. The test gate is special: its
// workflow step must also set CI_SKIP_POST_LANDING_TESTS=1.
var allowedPRLocalScripts = map[string]struct{}{
	".github/ci/candidate-suite.sh":                 {},
	".github/ci/default-build-transport-control.py": {},
	".github/ci/gate-selftest.sh":                   {},
	".github/ci/install-rust-toolchain.sh":          {},
	".github/ci/isolate-go-state.sh":                {},
	".github/ci/isolated-test.sh":                   {},
	".github/ci/ledger-consistency.sh":              {},
	".github/ci/naming-gate.sh":                     {},
	".github/ci/no-broad-suppression.sh":            {},
	".github/ci/pnpm-pin-guard.sh":                  {},
	".github/ci/release-source-gate.sh":             {},
	".github/ci/rust-pin-guard.sh":                  {},
	".github/ci/test-gate.sh":                       {},
	".github/ci/toolchain-identity.sh":              {},
	"scripts/restore-runner-tmp.sh":                 {},
}

// No local action is currently admitted in PR workflows. A future local action
// must add an exact path and exact input map here after its target is reviewed.
var allowedPRLocalActions = map[string]map[string]string{}

var allowedPRLocalGoTools = map[string]struct{}{
	"./tools/check-store-compatibility": {},
	"./tools/envprofilerace":            {},
	"./tools/sanitize-ci-evidence":      {},
}

func checkPRDeferredTestCommands(jobID string, job pullRequestJob, workflowEnv *yaml.Node) error {
	for _, step := range job.Steps {
		var goFlags []workflowShellWord
		goFlagsLoaded := false
		words := workflowShellWords(step.Run)
		for i := 0; i+1 < len(words); i++ {
			if words[i].operator || !isWorkflowGoCommand(words[i].value) || words[i+1].operator || words[i+1].value != "test" {
				continue
			}
			if !goFlagsLoaded {
				value, present, readable := workflowEnvironmentLookup("GOFLAGS", workflowEnv, &job.Env, &step.Env)
				if !readable {
					return fmt.Errorf("pull-request job %q has unclassifiable GOFLAGS", jobID)
				}
				if present {
					goFlags = workflowShellWords(value)
				}
				goFlagsLoaded = true
			}
			end := i + 2
			for end < len(words) && !words[end].operator {
				end++
			}
			if err := checkPRGoTestArguments(jobID, words[i+2:end], goFlags, workflowEnv, &job.Env, &step.Env); err != nil {
				return err
			}
			i = end - 1
		}
	}
	return nil
}

func workflowShellWords(text string) []workflowShellWord {
	text = strings.ReplaceAll(text, "\\\r\n", " ")
	text = strings.ReplaceAll(text, "\\\n", " ")
	var words []workflowShellWord
	var current strings.Builder
	var quote byte
	var wordQuote byte
	started := false
	flush := func() {
		if started {
			words = append(words, workflowShellWord{value: current.String(), quote: wordQuote})
			current.Reset()
			wordQuote = 0
			started = false
		}
	}
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if quote == '\'' {
			if ch == '\'' {
				quote = 0
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		if ch == '\\' && i+1 < len(text) {
			if text[i+1] == '\n' {
				i++
				continue
			}
			if isWindowsDrivePathPrefix(current.String()) {
				current.WriteByte('/')
				started = true
				continue
			}
			current.WriteByte(text[i+1])
			started = true
			i++
			continue
		}
		if quote == '"' {
			if ch == '"' {
				quote = 0
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			if started {
				wordQuote = 0
			} else {
				wordQuote = ch
			}
			started = true
			quote = ch
			continue
		}
		if ch == '\n' {
			flush()
			words = append(words, workflowShellWord{value: ";", operator: true})
			continue
		}
		if ch == ' ' || ch == '\t' || ch == '\r' {
			flush()
			continue
		}
		if strings.ContainsRune(";|&()<> ", rune(ch)) {
			flush()
			operator := string(ch)
			if i+1 < len(text) && ((ch == '&' && text[i+1] == '&') || (ch == '|' && text[i+1] == '|')) {
				operator += string(text[i+1])
				i++
			}
			words = append(words, workflowShellWord{value: operator, operator: true})
			continue
		}
		current.WriteByte(ch)
		started = true
	}
	flush()
	return words
}

func isWindowsDrivePathPrefix(value string) bool {
	return len(value) >= 2 && ((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) && value[1] == ':'
}

func isWorkflowGoCommand(value string) bool {
	if value == "go" || value == "$GO" || value == "${GO}" || strings.EqualFold(value, "go.exe") {
		return true
	}
	path := strings.ToLower(strings.ReplaceAll(value, "\\", "/"))
	return strings.HasSuffix(path, "/go") || strings.HasSuffix(path, "/go.exe")
}

func checkPRGoTestArguments(jobID string, args, goFlags []workflowShellWord, workflowEnv, jobEnv, stepEnv *yaml.Node) error {
	var runSelector, skipSelector string
	hasRun, hasSkip := false, false
	deferredPackage := false
	goTestArguments := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		value := arg.value
		if value == "-args" {
			goTestArguments = true
			continue
		}
		switch {
		case value == "-run" || value == "--run" || value == "-test.run" || value == "--test.run":
			if hasRun || i+1 >= len(args) {
				return fmt.Errorf("pull-request job %q has an unreadable go test -run selector", jobID)
			}
			i++
			resolved, err := resolvePRTestSelector(args[i], workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -run selector: %w", jobID, err)
			}
			runSelector, hasRun = resolved, true
		case strings.HasPrefix(value, "-run=") || strings.HasPrefix(value, "--run=") || strings.HasPrefix(value, "-test.run=") || strings.HasPrefix(value, "--test.run="):
			if hasRun {
				return fmt.Errorf("pull-request job %q repeats go test -run", jobID)
			}
			resolved, err := resolvePRTestSelector(workflowShellWord{value: value[strings.IndexByte(value, '=')+1:], quote: arg.quote}, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -run selector: %w", jobID, err)
			}
			runSelector, hasRun = resolved, true
		case value == "-skip" || value == "--skip" || value == "-test.skip" || value == "--test.skip":
			if hasSkip || i+1 >= len(args) {
				return fmt.Errorf("pull-request job %q has an unreadable go test -skip selector", jobID)
			}
			i++
			resolved, err := resolvePRTestSelector(args[i], workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -skip selector: %w", jobID, err)
			}
			skipSelector, hasSkip = resolved, true
		case strings.HasPrefix(value, "-skip=") || strings.HasPrefix(value, "--skip=") || strings.HasPrefix(value, "-test.skip=") || strings.HasPrefix(value, "--test.skip="):
			if hasSkip {
				return fmt.Errorf("pull-request job %q repeats go test -skip", jobID)
			}
			resolved, err := resolvePRTestSelector(workflowShellWord{value: value[strings.IndexByte(value, '=')+1:], quote: arg.quote}, workflowEnv, jobEnv, stepEnv)
			if err != nil {
				return fmt.Errorf("pull-request job %q has an unclassifiable go test -skip selector: %w", jobID, err)
			}
			skipSelector, hasSkip = resolved, true
		case value == "-exec" || value == "--exec" || strings.HasPrefix(value, "-exec=") || strings.HasPrefix(value, "--exec="):
			return fmt.Errorf("pull-request job %q uses a custom go test executor", jobID)
		default:
			if goTestArguments || strings.HasPrefix(value, "-") {
				continue
			}
			packagePath := value
			if hasDynamicShellReference(value) {
				resolved, err := resolvePRTestSelector(arg, workflowEnv, jobEnv, stepEnv)
				if err != nil {
					return fmt.Errorf("pull-request job %q has an unclassifiable go test package target: %w", jobID, err)
				}
				packagePath = resolved
			}
			if packageReachesPRDeferredRoot(packagePath) {
				deferredPackage = true
			}
		}
	}
	goRun, goSkip, hasGoRun, hasGoSkip, err := parsePRGoFlags(jobID, goFlags)
	if err != nil {
		return err
	}
	if !hasRun && hasGoRun {
		runSelector, hasRun = goRun, true
	}
	if !hasSkip && hasGoSkip {
		skipSelector, hasSkip = goSkip, true
	}
	if hasRun && prSelectorMatchesDeferredTests(runSelector) && !(hasSkip && prSelectorSkipsAllDeferredTests(skipSelector)) {
		return fmt.Errorf("pull-request job %q selects deferred tests with go test -run", jobID)
	}
	if deferredPackage && (!hasRun || prSelectorMatchesDeferredTests(runSelector)) && !(hasSkip && prSelectorSkipsAllDeferredTests(skipSelector)) {
		return fmt.Errorf("pull-request job %q runs an unfiltered deferred test root", jobID)
	}
	return nil
}

func parsePRGoFlags(jobID string, words []workflowShellWord) (string, string, bool, bool, error) {
	var runSelector, skipSelector string
	hasRun, hasSkip := false, false
	for _, word := range words {
		if word.operator {
			return "", "", false, false, fmt.Errorf("pull-request job %q has unclassifiable GOFLAGS syntax", jobID)
		}
		value := word.value
		switch {
		case value == "-run" || value == "--run" || value == "-test.run" || value == "--test.run":
			return "", "", false, false, fmt.Errorf("pull-request job %q has a GOFLAGS selector without a static value", jobID)
		case strings.HasPrefix(value, "-run=") || strings.HasPrefix(value, "--run=") || strings.HasPrefix(value, "-test.run=") || strings.HasPrefix(value, "--test.run="):
			if hasRun {
				return "", "", false, false, fmt.Errorf("pull-request job %q repeats GOFLAGS -run", jobID)
			}
			runSelector = value[strings.IndexByte(value, '=')+1:]
			if runSelector == "" || strings.Contains(runSelector, "${{") {
				return "", "", false, false, fmt.Errorf("pull-request job %q has an unclassifiable GOFLAGS -run selector", jobID)
			}
			hasRun = true
		case value == "-skip" || value == "--skip" || value == "-test.skip" || value == "--test.skip":
			return "", "", false, false, fmt.Errorf("pull-request job %q has a GOFLAGS skip selector without a static value", jobID)
		case strings.HasPrefix(value, "-skip=") || strings.HasPrefix(value, "--skip=") || strings.HasPrefix(value, "-test.skip=") || strings.HasPrefix(value, "--test.skip="):
			if hasSkip {
				return "", "", false, false, fmt.Errorf("pull-request job %q repeats GOFLAGS -skip", jobID)
			}
			skipSelector = value[strings.IndexByte(value, '=')+1:]
			if skipSelector == "" || strings.Contains(skipSelector, "${{") {
				return "", "", false, false, fmt.Errorf("pull-request job %q has an unclassifiable GOFLAGS -skip selector", jobID)
			}
			hasSkip = true
		case value == "-exec" || value == "--exec" || strings.HasPrefix(value, "-exec=") || strings.HasPrefix(value, "--exec="):
			return "", "", false, false, fmt.Errorf("pull-request job %q uses a custom GOFLAGS test executor", jobID)
		default:
			if !strings.HasPrefix(value, "-") {
				return "", "", false, false, fmt.Errorf("pull-request job %q has an unclassifiable GOFLAGS entry", jobID)
			}
		}
	}
	return runSelector, skipSelector, hasRun, hasSkip, nil
}

func resolvePRTestSelector(word workflowShellWord, workflowEnv, jobEnv, stepEnv *yaml.Node) (string, error) {
	selector := strings.TrimSpace(word.value)
	if word.quote == '\'' {
		return selector, nil
	}
	if selector == "" || strings.Contains(selector, "${{") || strings.ContainsRune(selector, '`') {
		return "", fmt.Errorf("selector is empty or dynamic")
	}
	if match := workflowShellVariable.FindStringSubmatch(selector); len(match) == 3 {
		name := match[1]
		if name == "" {
			name = match[2]
		}
		value, ok := workflowEnvironmentValue(name, workflowEnv, jobEnv, stepEnv)
		if !ok {
			return "", fmt.Errorf("shell variable %q is not statically defined", name)
		}
		return value, nil
	}
	if hasDynamicShellReference(selector) {
		return "", fmt.Errorf("selector contains an unresolved shell expansion")
	}
	return selector, nil
}

func hasDynamicShellReference(value string) bool {
	if value == "$" {
		return true
	}
	for i := 0; i < len(value); i++ {
		if value[i] != '$' || i+1 >= len(value) {
			continue
		}
		next := value[i+1]
		if next == '{' || next == '(' || next == '?' || next == '!' || next == '$' || next == '-' || next >= '0' && next <= '9' || next == '_' || next >= 'A' && next <= 'Z' || next >= 'a' && next <= 'z' {
			return true
		}
		if next != '|' && next != ')' && next != '/' && next != ']' {
			return true
		}
	}
	return false
}

func workflowEnvironmentValue(name string, levels ...*yaml.Node) (string, bool) {
	value, present, readable := workflowEnvironmentLookup(name, levels...)
	return value, present && readable
}

func workflowEnvironmentLookup(name string, levels ...*yaml.Node) (string, bool, bool) {
	var value *yaml.Node
	for _, level := range levels {
		level = resolveWorkflowAlias(level)
		if current := workflowMappingValue(level, name); current != nil {
			value = current
		}
	}
	if value == nil {
		return "", false, true
	}
	value = resolveWorkflowAlias(value)
	if value == nil || value.Kind != yaml.ScalarNode || value.Tag != "!!str" || strings.Contains(value.Value, "${{") {
		return "", true, false
	}
	return value.Value, true, true
}

func packageReachesPRDeferredRoot(value string) bool {
	return value == "./..." || value == "..." || value == "./cmd/curator" ||
		value == "github.com/relux-works/curator/cmd/curator" || strings.HasSuffix(value, "/cmd/curator") ||
		strings.HasPrefix(value, "./cmd/curator/") || strings.HasSuffix(value, "/cmd/curator/...")
}

func prSelectorMatchesDeferredTests(selector string) bool {
	parts := strings.Split(selector, "/")
	pattern, err := regexp.Compile(parts[0])
	if err != nil {
		return true
	}
	for _, name := range curatormutations.PostLandingNames() {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

func prSelectorSkipsAllDeferredTests(selector string) bool {
	parts := strings.Split(selector, "/")
	pattern, err := regexp.Compile(parts[0])
	if err != nil {
		return false
	}
	for _, name := range curatormutations.PostLandingNames() {
		if !pattern.MatchString(name) {
			return false
		}
	}
	return true
}

func checkPRLocalTargets(jobID string, job pullRequestJob, workflowEnv *yaml.Node) error {
	for _, step := range job.Steps {
		words := workflowShellWords(step.Run)
		for i := 0; i+1 < len(words); i++ {
			if words[i].operator || !isWorkflowGoCommand(words[i].value) || words[i+1].operator || words[i+1].value != "run" {
				continue
			}
			for j := i + 2; j < len(words) && !words[j].operator; j++ {
				candidate := words[j].value
				if strings.HasPrefix(candidate, "-") {
					continue
				}
				if strings.HasPrefix(candidate, "$") {
					if !isAllowedPRDynamicGoRun(step) {
						return fmt.Errorf("pull-request job %q uses a dynamic local go run target", jobID)
					}
					break
				}
				if strings.HasPrefix(candidate, "./") || strings.HasSuffix(candidate, ".go") {
					if _, ok := allowedPRLocalGoTools[candidate]; !ok {
						return fmt.Errorf("pull-request job %q calls an unapproved local Go tool %q", jobID, candidate)
					}
					break
				}
			}
		}
		uses := strings.TrimSpace(step.Uses)
		if isPRLocalPath(uses) {
			allowedInputs, ok := allowedPRLocalActions[uses]
			actualInputs, readable := workflowStringMap(&step.With)
			if !ok || !readable || !reflect.DeepEqual(actualInputs, allowedInputs) {
				return fmt.Errorf("pull-request job %q calls an unapproved local action %q or changes its inputs", jobID, uses)
			}
		}
		if dynamicWorkflowScriptInvocation.MatchString(step.Run) {
			return fmt.Errorf("pull-request job %q calls a local script through a dynamic path", jobID)
		}
		if dynamicWorkflowSourceInvocation.MatchString(step.Run) {
			return fmt.Errorf("pull-request job %q sources a local script through a dynamic path", jobID)
		}
		if inlineWorkflowScriptBody.MatchString(step.Run) {
			return fmt.Errorf("pull-request job %q embeds an unclassifiable shell script body", jobID)
		}
		if inlineWorkflowHeredoc.MatchString(step.Run) && !isAllowedPRInlineHeredoc(step) {
			return fmt.Errorf("pull-request job %q embeds an unapproved interpreter heredoc", jobID)
		}
		for _, script := range prLocalScripts(step.Run) {
			if _, ok := allowedPRLocalScripts[script]; !ok {
				return fmt.Errorf("pull-request job %q calls an unapproved local script %q", jobID, script)
			}
			if script == ".github/ci/test-gate.sh" {
				value, ok := workflowEnvironmentValue("CI_SKIP_POST_LANDING_TESTS", workflowEnv, &job.Env, &step.Env)
				if !ok || value != "1" {
					return fmt.Errorf("pull-request job %q calls test-gate.sh without the deferred-suite skip", jobID)
				}
				if testGateSkipUnset.MatchString(step.Run) || isolatedWorkflowEnvironment.MatchString(step.Run) {
					return fmt.Errorf("pull-request job %q removes the test-gate deferred-suite skip", jobID)
				}
				for _, match := range testGateSkipOverride.FindAllStringSubmatch(step.Run, -1) {
					if strings.Trim(match[1], "\"'") != "1" {
						return fmt.Errorf("pull-request job %q overrides the test-gate deferred-suite skip", jobID)
					}
				}
			}
		}
	}
	return nil
}

func isAllowedPRDynamicGoRun(step pullRequestStep) bool {
	return step.Name == "Record macOS race capacity" && step.If == "runner.os == 'macOS'" &&
		fmt.Sprintf("%x", sha256.Sum256([]byte(step.Run))) == "6132d13dc9aea819d1cd3aac9107a2742c6e83ea9e87807709fe545d91ab8780"
}

func isAllowedPRInlineHeredoc(step pullRequestStep) bool {
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(step.Run)))
	switch step.Name {
	case "Driver qualification and reproducibility":
		return digest == "34dd145d07394a40b8159c102e5665a1363050b74e6623a802a8a57eb92ff0d8" ||
			digest == "dfeb880111ffb7b4a3b48aa2adb757ae2d344042a4a0cc13564b5c66522d1502"
	case "Compare thirty isolated counts at each revision":
		return digest == "34e2488b9fba5fbb5a0d64c4ca1fac95a431be59499ba12b02271c4ae08d9fd1"
	default:
		return false
	}
}

func isPRLocalPath(value string) bool {
	return strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") || strings.HasPrefix(value, ".github/")
}

func prLocalScripts(run string) []string {
	found := make(map[string]struct{})
	for _, pattern := range []*regexp.Regexp{localWorkflowScriptInvocation, localWorkflowScriptToken, localWorkflowExecutable} {
		for _, match := range pattern.FindAllStringSubmatch(run, -1) {
			if len(match) < 2 {
				continue
			}
			value := strings.Trim(match[1], "\"'")
			if strings.HasPrefix(value, "./") {
				value = strings.TrimPrefix(value, "./")
			}
			found[value] = struct{}{}
		}
	}
	paths := make([]string, 0, len(found))
	for value := range found {
		paths = append(paths, value)
	}
	sort.Strings(paths)
	return paths
}

func workflowStringMap(node *yaml.Node) (map[string]string, bool) {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind == 0 || node.Tag == "!!null" {
		return map[string]string{}, true
	}
	if node.Kind != yaml.MappingNode {
		return nil, false
	}
	values := make(map[string]string, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		value := resolveWorkflowAlias(node.Content[i+1])
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || value == nil || value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			return nil, false
		}
		values[key.Value] = value.Value
	}
	return values, true
}

func pullRequestMatrixValues(strategy *yaml.Node) (map[string][]string, error) {
	values := make(map[string][]string)
	strategy = resolveWorkflowAlias(strategy)
	if strategy == nil || strategy.Kind == 0 || strategy.Tag == "!!null" {
		return values, nil
	}
	if strategy.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("strategy is not a mapping")
	}
	matrix := resolveWorkflowAlias(workflowMappingValue(strategy, "matrix"))
	if matrix == nil || matrix.Kind == 0 || matrix.Tag == "!!null" {
		return values, nil
	}
	if matrix.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("matrix is not a static mapping")
	}
	seen := make(map[string]struct{}, len(matrix.Content)/2)
	for i := 0; i+1 < len(matrix.Content); i += 2 {
		key := matrix.Content[i]
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
			return nil, fmt.Errorf("matrix has a non-string key")
		}
		if _, ok := seen[key.Value]; ok {
			return nil, fmt.Errorf("matrix repeats key %q", key.Value)
		}
		seen[key.Value] = struct{}{}
		switch key.Value {
		case "include":
			if err := collectPRMatrixRows(matrix.Content[i+1], values, true); err != nil {
				return nil, fmt.Errorf("matrix include: %w", err)
			}
		case "exclude":
			if err := collectPRMatrixRows(matrix.Content[i+1], values, false); err != nil {
				return nil, fmt.Errorf("matrix exclude: %w", err)
			}
		default:
			items, err := prMatrixDimensionValues(matrix.Content[i+1])
			if err != nil {
				return nil, fmt.Errorf("matrix field %q: %w", key.Value, err)
			}
			values[key.Value] = append(values[key.Value], items...)
		}
	}
	return values, nil
}

func collectPRMatrixRows(node *yaml.Node, values map[string][]string, include bool) error {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind != yaml.SequenceNode {
		return fmt.Errorf("rows are not a sequence")
	}
	for _, row := range node.Content {
		row = resolveWorkflowAlias(row)
		if row == nil || row.Kind != yaml.MappingNode {
			return fmt.Errorf("row is not a mapping")
		}
		seen := make(map[string]struct{}, len(row.Content)/2)
		for i := 0; i+1 < len(row.Content); i += 2 {
			key := row.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return fmt.Errorf("row has a non-string key")
			}
			if _, ok := seen[key.Value]; ok {
				return fmt.Errorf("row repeats key %q", key.Value)
			}
			seen[key.Value] = struct{}{}
			value, err := prMatrixScalar(row.Content[i+1])
			if err != nil {
				return fmt.Errorf("field %q: %w", key.Value, err)
			}
			if include {
				values[key.Value] = append(values[key.Value], value)
			}
		}
	}
	return nil
}

func prMatrixDimensionValues(node *yaml.Node) ([]string, error) {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("dimension is not a sequence")
	}
	values := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		value, err := prMatrixScalar(item)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("dimension is empty")
	}
	return values, nil
}

func prMatrixScalar(node *yaml.Node) (string, error) {
	node = resolveWorkflowAlias(node)
	if node == nil || node.Kind != yaml.ScalarNode || (node.Tag != "!!str" && node.Tag != "!!int" && node.Tag != "!!bool") {
		return "", fmt.Errorf("value is not a static string, integer, or boolean")
	}
	if strings.Contains(node.Value, "${{") || strings.Contains(node.Value, "}}") {
		return "", fmt.Errorf("value contains an unclassifiable dynamic expression")
	}
	return node.Value, nil
}

func validatePRMatrixField(field string, values map[string][]string) error {
	allowed, ok := allowedPRMatrixValues[field]
	if !ok {
		return fmt.Errorf("matrix field %q is not allowlisted", field)
	}
	fieldValues, ok := values[field]
	if !ok || len(fieldValues) == 0 {
		return fmt.Errorf("matrix field %q has no classifiable values", field)
	}
	for _, value := range fieldValues {
		if isDeferredSuiteMatrixValue(value) {
			return fmt.Errorf("matrix field %q carries deferred-suite selector or command %q", field, value)
		}
		if _, ok := allowed[value]; !ok {
			return fmt.Errorf("matrix field %q has unapproved value %q", field, value)
		}
	}
	return nil
}

func isDeferredSuiteMatrixValue(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, legacyMutationJobID) ||
		strings.Contains(value, "tools/mutationinventory") ||
		strings.Contains(value, "mutationinventory") ||
		hasMutationSuiteLabel(value) ||
		strings.Contains(value, "exhaustive") ||
		strings.Contains(value, "post-landing")
}

func workflowStringValue(mapping *yaml.Node, key string) (string, bool) {
	mapping = resolveWorkflowAlias(mapping)
	value := resolveWorkflowAlias(workflowMappingValue(mapping, key))
	if value == nil || value.Kind != yaml.ScalarNode || value.Tag != "!!str" || strings.Contains(value.Value, "${{") {
		return "", false
	}
	return value.Value, true
}

func canRunOnPullRequest(condition string) bool {
	condition = strings.TrimSpace(condition)
	if strings.HasPrefix(condition, "${{") && strings.HasSuffix(condition, "}}") {
		condition = strings.TrimSpace(condition[3 : len(condition)-2])
	}
	condition = strings.Join(strings.Fields(condition), " ")
	switch condition {
	case "vars.ROSE_AIR_RUNNER == 'true' && github.event_name == 'push' && github.ref == 'refs/heads/main'":
		return false
	case "github.event_name == 'workflow_dispatch' && (inputs.candidate_ref != '' || inputs.candidate_root != '')":
		return false
	case "github.event_name == 'workflow_dispatch' && inputs.fixture_isolation_diagnostics":
		return false
	default:
		// Unknown conditions are included so a new or changed condition cannot
		// silently remove a job from the pull-request scan.
		return true
	}
}

func isMutationSuiteJob(id string, job pullRequestJob) bool {
	identity := strings.ToLower(id + " " + job.Name)
	if strings.Contains(identity, legacyMutationJobID) || hasMutationSuiteLabel(identity) || strings.Contains(strings.ToLower(job.Uses), "mutation") {
		return true
	}
	for _, step := range job.Steps {
		run := strings.ToLower(step.Run)
		if hasMutationSuiteLabel(strings.ToLower(step.Name)) || hasMutationSuiteLabel(run) || strings.Contains(run, "tools/mutationinventory") {
			return true
		}
	}
	return false
}

func hasMutationSuiteLabel(identity string) bool {
	if !strings.Contains(identity, "mutation") {
		return false
	}
	return strings.Contains(identity, "suite") || strings.Contains(identity, "control") || strings.Contains(identity, "inventory")
}

func pullRequestNeedIDs(needs *yaml.Node) ([]string, error) {
	if needs == nil || needs.Kind == 0 || needs.Tag == "!!null" {
		return nil, nil
	}
	if needs.Kind == yaml.AliasNode && needs.Alias != nil {
		needs = needs.Alias
	}
	if needs.Kind == yaml.ScalarNode {
		return []string{needs.Value}, nil
	}
	if needs.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected scalar or sequence, got %s", needs.ShortTag())
	}
	ids := make([]string, 0, len(needs.Content))
	for _, dependency := range needs.Content {
		if dependency.Kind != yaml.ScalarNode || dependency.Tag != "!!str" {
			return nil, fmt.Errorf("dependency is not a string scalar")
		}
		ids = append(ids, dependency.Value)
	}
	return ids, nil
}

func isMutationJobID(id string, mutationJobIDs map[string]struct{}) bool {
	_, ok := mutationJobIDs[strings.ToLower(id)]
	return ok
}

func hasMutationResultReference(text string, mutationJobIDs map[string]struct{}) bool {
	if dynamicNeedsResultReference.MatchString(text) {
		// The current PR workflow has no dynamic needs.result selectors. Their
		// target cannot be proven non-mutation from the workflow text alone.
		return true
	}
	for _, match := range mutationJobResultReference.FindAllStringSubmatch(text, -1) {
		id := match[1]
		if id == "" {
			id = match[2]
		}
		if isMutationJobID(id, mutationJobIDs) {
			return true
		}
	}
	return false
}

func hasMutationResultReferenceInNode(node *yaml.Node, mutationJobIDs map[string]struct{}) bool {
	for _, value := range workflowScalarValues(node) {
		if hasMutationResultReference(value, mutationJobIDs) {
			return true
		}
	}
	return false
}

func workflowScalarValues(node *yaml.Node) []string {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.AliasNode && node.Alias != nil {
		return workflowScalarValues(node.Alias)
	}
	if node.Kind == yaml.ScalarNode {
		return []string{node.Value}
	}
	var values []string
	for _, child := range node.Content {
		values = append(values, workflowScalarValues(child)...)
	}
	return values
}

type workflowStepAddition struct {
	Name string
	Run  string
	Uses string
	With map[string]string
	Env  map[string]string
}

func addStepToWorkflowJob(data []byte, jobName string, step workflowStepAddition) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	jobs := workflowMappingValue(root, "jobs")
	job := workflowMappingValue(jobs, jobName)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobName)
	}
	steps := workflowMappingValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("workflow job %q has no steps sequence", jobName)
	}
	steps.Content = append(steps.Content, workflowStepNode(step))
	return yaml.Marshal(document)
}

func addWorkflowJob(data []byte, jobID, name, run, uses, condition string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	jobs := workflowMappingValue(root, "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("workflow has no jobs mapping")
	}
	if workflowMappingValue(jobs, jobID) != nil {
		return nil, fmt.Errorf("workflow job %q already exists", jobID)
	}
	job := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendWorkflowMappingPair(job, "name", workflowStringNode(name))
	if condition != "" {
		appendWorkflowMappingPair(job, "if", workflowStringNode(condition))
	}
	if uses != "" {
		appendWorkflowMappingPair(job, "uses", workflowStringNode(uses))
	} else {
		appendWorkflowMappingPair(job, "runs-on", workflowStringNode("ubuntu-latest"))
		steps := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		if run != "" {
			steps.Content = append(steps.Content, workflowStepNode(workflowStepAddition{Run: run}))
		}
		appendWorkflowMappingPair(job, "steps", steps)
	}
	appendWorkflowMappingPair(jobs, jobID, job)
	return yaml.Marshal(document)
}

func addWorkflowMatrixInclude(data []byte, jobID string, rows []map[string]string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	job := workflowMappingValue(workflowMappingValue(root, "jobs"), jobID)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobID)
	}
	if workflowMappingValue(job, "strategy") != nil {
		return nil, fmt.Errorf("workflow job %q already has a strategy", jobID)
	}
	include := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, values := range rows {
		row := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendWorkflowMappingPair(row, key, workflowStringNode(values[key]))
		}
		include.Content = append(include.Content, row)
	}
	matrix := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendWorkflowMappingPair(matrix, "include", include)
	strategy := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	appendWorkflowMappingPair(strategy, "matrix", matrix)
	appendWorkflowMappingPair(job, "strategy", strategy)
	return yaml.Marshal(document)
}

func setWorkflowTriggerEvents(data []byte, events []string) ([]byte, error) {
	on := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, event := range events {
		appendWorkflowMappingPair(on, event, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"})
	}
	return setWorkflowTriggerNode(data, on)
}

func setWorkflowTriggerNode(data []byte, on *yaml.Node) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	workflowSetMappingValue(root, "on", on)
	return yaml.Marshal(document)
}

func setWorkflowStepCondition(data []byte, jobID, stepName, condition string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	job := workflowMappingValue(workflowMappingValue(root, "jobs"), jobID)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobID)
	}
	steps := workflowMappingValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("workflow job %q has no steps sequence", jobID)
	}
	found := false
	for _, step := range steps.Content {
		name := workflowMappingValue(step, "name")
		if name == nil || name.Value != stepName {
			continue
		}
		if found {
			return nil, fmt.Errorf("workflow job %q has multiple steps named %q", jobID, stepName)
		}
		workflowSetMappingValue(step, "if", workflowStringNode(condition))
		found = true
	}
	if !found {
		return nil, fmt.Errorf("workflow job %q has no step named %q", jobID, stepName)
	}
	return yaml.Marshal(document)
}

func workflowStepConditionValue(data []byte, jobID, stepName string) (string, error) {
	_, root, err := parseWorkflowDocument(data)
	if err != nil {
		return "", err
	}
	job := workflowMappingValue(workflowMappingValue(root, "jobs"), jobID)
	if job == nil {
		return "", fmt.Errorf("workflow job %q is missing", jobID)
	}
	steps := workflowMappingValue(job, "steps")
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return "", fmt.Errorf("workflow job %q has no steps sequence", jobID)
	}
	for _, step := range steps.Content {
		name := workflowMappingValue(step, "name")
		if name == nil || name.Value != stepName {
			continue
		}
		condition := workflowMappingValue(step, "if")
		if condition == nil || condition.Kind != yaml.ScalarNode {
			return "", fmt.Errorf("workflow job %q step %q has no scalar if condition", jobID, stepName)
		}
		return condition.Value, nil
	}
	return "", fmt.Errorf("workflow job %q has no step named %q", jobID, stepName)
}

func appendWorkflowJobNeed(data []byte, jobID, dependency string) ([]byte, error) {
	document, root, err := parseWorkflowDocument(data)
	if err != nil {
		return nil, err
	}
	job := workflowMappingValue(workflowMappingValue(root, "jobs"), jobID)
	if job == nil {
		return nil, fmt.Errorf("workflow job %q is missing", jobID)
	}
	needs := workflowMappingValue(job, "needs")
	if needs == nil {
		needs = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		appendWorkflowMappingPair(job, "needs", needs)
	} else if needs.Kind == yaml.ScalarNode {
		needs = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode(needs.Value)}}
		workflowSetMappingValue(job, "needs", needs)
	} else if needs.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("workflow job %q needs is not a sequence", jobID)
	}
	needs.Content = append(needs.Content, workflowStringNode(dependency))
	return yaml.Marshal(document)
}

func workflowStepNode(step workflowStepAddition) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if step.Name != "" {
		appendWorkflowMappingPair(node, "name", workflowStringNode(step.Name))
	}
	if step.Run != "" {
		appendWorkflowMappingPair(node, "run", workflowStringNode(step.Run))
	}
	if step.Uses != "" {
		appendWorkflowMappingPair(node, "uses", workflowStringNode(step.Uses))
	}
	if len(step.With) != 0 {
		with := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(step.With))
		for key := range step.With {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendWorkflowMappingPair(with, key, workflowStringNode(step.With[key]))
		}
		appendWorkflowMappingPair(node, "with", with)
	}
	if len(step.Env) != 0 {
		env := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(step.Env))
		for key := range step.Env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appendWorkflowMappingPair(env, key, workflowStringNode(step.Env[key]))
		}
		appendWorkflowMappingPair(node, "env", env)
	}
	return node
}

func parseWorkflowDocument(data []byte) (*yaml.Node, *yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, nil, err
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("workflow document has no root mapping")
	}
	return &document, document.Content[0], nil
}

func workflowMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func appendWorkflowMappingPair(mapping *yaml.Node, key string, value *yaml.Node) {
	mapping.Content = append(mapping.Content, workflowStringNode(key), value)
}

func workflowSetMappingValue(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	appendWorkflowMappingPair(mapping, key, value)
}

func workflowStringNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

// Claim: PR test gates defer the expensive suite, require a PR trigger, and only
// expose reviewed static matrix inputs to commands and environment values.
// Limit: this checks static wiring; hosted execution and product behavior remain unverified.
func TestPullRequestWorkflowDefersMutationAndExhaustiveSuites(t *testing.T) {
	good := mutationInventoryWorkflowBytes(t)
	if strings.Contains(string(good), "tools/mutationinventory") {
		t.Fatal("PR workflow invokes the deferred-suite inventory tool directly")
	}
	if err := checkPRDefersPostLandingTests(good); err != nil {
		t.Fatalf("valid PR deferral rejected: %v", err)
	}
	triggerForms := []struct {
		name  string
		plant func([]byte) ([]byte, error)
	}{
		{
			name: "scalar_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				return setWorkflowTriggerNode(data, workflowStringNode("pull_request"))
			},
		},
		{
			name: "sequence_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				on := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode("push"), workflowStringNode("pull_request")}}
				return setWorkflowTriggerNode(data, on)
			},
		},
		{
			name: "mapping_null_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				on := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(on, "pull_request", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"})
				return setWorkflowTriggerNode(data, on)
			},
		},
		{
			name: "mapping_filtered_pull_request_trigger",
			plant: func(data []byte) ([]byte, error) {
				branches := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode("main")}}
				filter := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(filter, "branches", branches)
				on := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(on, "pull_request", filter)
				return setWorkflowTriggerNode(data, on)
			},
		},
	}
	for _, test := range triggerForms {
		t.Run("valid_"+test.name, func(t *testing.T) {
			workflow, err := test.plant(good)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPRDefersPostLandingTests(workflow); err != nil {
				t.Fatalf("supported trigger form rejected: %v", err)
			}
		})
	}
	tests := []struct {
		name  string
		plant func([]byte) ([]byte, error)
	}{
		{
			name: "restored_pr_job",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, legacyMutationJobID, "Curator mutation controls", "go run ./tools/mutationinventory -mode inventory", "", "")
			},
		},
		{
			name: "renamed_pr_mutation_job",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "deferred-suite", "Curator mutation controls", "go run ./tools/mutationinventory -mode inventory", "", "")
			},
		},
		{
			name: "mutation_job_with_unknown_condition",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "suite-lane", "Scheduled analysis", "go run ./tools/mutationinventory -mode inventory", "", "github.event_name == 'pull_request' || inputs.run_suite")
			},
		},
		{
			name: "neutral_local_reusable_workflow_calls_deferred_suite",
			plant: func(data []byte) ([]byte, error) {
				target := []byte("name: Quality\non: workflow_call\njobs:\n  routine:\n    runs-on: ubuntu-latest\n    steps:\n      - run: go run ./tools/mutationinventory -mode inventory\n")
				var targetWorkflow pullRequestWorkflow
				if err := yaml.Unmarshal(target, &targetWorkflow); err != nil {
					return nil, err
				}
				if !isMutationSuiteJob("routine", targetWorkflow.Jobs["routine"]) {
					return nil, fmt.Errorf("reusable-workflow target fixture does not call the deferred suite")
				}
				return addWorkflowJob(data, "quality-check", "Quality check", "", "./.github/workflows/quality.yml", "")
			},
		},
		{
			name: "renamed_reusable_suite_job",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "quality-check", "Quality check", "", "./.github/workflows/post-landing-mutation-tests.yml", "")
			},
		},
		{
			name: "matrix_include_supplies_deferred_suite_selector",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ matrix.suite_selector }}'`, "", "")
				if err != nil {
					return nil, err
				}
				return addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "TestMutationInventory|TestExhaustive"}})
			},
		},
		{
			name: "matrix_include_supplies_deferred_suite_selector_through_env",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run "$SUITE_SELECTOR"`, "", "")
				if err != nil {
					return nil, err
				}
				mutant, err = addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "TestMutationInventory|TestExhaustive"}})
				if err != nil {
					return nil, err
				}
				document, root, err := parseWorkflowDocument(mutant)
				if err != nil {
					return nil, err
				}
				job := workflowMappingValue(workflowMappingValue(root, "jobs"), "routine-check")
				env := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(env, "SUITE_SELECTOR", workflowStringNode("${{ matrix.suite_selector }}"))
				appendWorkflowMappingPair(job, "env", env)
				return yaml.Marshal(document)
			},
		},
		{
			name: "matrix_include_has_unclassifiable_dynamic_input",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ matrix.suite_selector }}'`, "", "")
				if err != nil {
					return nil, err
				}
				return addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "${{ inputs.suite_filter }}"}})
			},
		},
		{
			name: "matrix_lookup_is_unclassifiable",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ matrix[inputs.selector] }}'`, "", "")
				if err != nil {
					return nil, err
				}
				return addWorkflowMatrixInclude(mutant, "routine-check", []map[string]string{{"suite_selector": "TestRoutine"}})
			},
		},
		{
			name: "unclassified_dynamic_run_selector",
			plant: func(data []byte) ([]byte, error) {
				return addWorkflowJob(data, "routine-check", "Routine check", `go test ./... -run '${{ inputs.suite_filter }}'`, "", "")
			},
		},
		{
			name: "scalar_push_only_trigger",
			plant: func(data []byte) ([]byte, error) {
				return setWorkflowTriggerNode(data, workflowStringNode("push"))
			},
		},
		{
			name: "sequence_push_only_trigger",
			plant: func(data []byte) ([]byte, error) {
				on := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{workflowStringNode("push")}}
				return setWorkflowTriggerNode(data, on)
			},
		},
		{
			name: "push_only_trigger",
			plant: func(data []byte) ([]byte, error) {
				return setWorkflowTriggerEvents(data, []string{"push"})
			},
		},
		{
			name: "duplicate_trigger_mapping_is_rejected",
			plant: func(data []byte) ([]byte, error) {
				document, root, err := parseWorkflowDocument(data)
				if err != nil {
					return nil, err
				}
				push := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(push, "push", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"})
				appendWorkflowMappingPair(root, "on", push)
				return yaml.Marshal(document)
			},
		},
		{
			name: "restored_pr_dependency",
			plant: func(data []byte) ([]byte, error) {
				return appendWorkflowJobNeed(data, "test", legacyMutationJobID)
			},
		},
		{
			name: "legacy_bridge_in_linux_build_controls",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Require PR mutation gate success",
					Run:  `test "${{ needs.curator-mutation-pr.result }}" = success`,
				})
			},
		},
		{
			name: "renamed_bridge_run",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Run:  `echo "${{ needs.curator-mutation-pr.result }}"`,
				})
			},
		},
		{
			name: "renamed_bridge_env",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Env:  map[string]string{"MUTATION_RESULT": "${{ needs.curator-mutation-pr.result }}"},
				})
			},
		},
		{
			name: "dynamic_mutation_result_reference",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Run:  `echo "${{ needs[vars.MUTATION_JOB].result }}"`,
				})
			},
		},
		{
			name: "renamed_mutation_result_reference",
			plant: func(data []byte) ([]byte, error) {
				mutant, err := addWorkflowJob(data, "deferred-suite", "Curator mutation controls", "go run ./tools/mutationinventory -mode inventory", "", "vars.ROSE_AIR_RUNNER == 'true' && github.event_name == 'push' && github.ref == 'refs/heads/main'")
				if err != nil {
					return nil, err
				}
				return addStepToWorkflowJob(mutant, "linux-build-controls", workflowStepAddition{
					Name: "Record shard outcome",
					Run:  `echo "${{ needs.deferred-suite.result }}"`,
				})
			},
		},
		{
			name: "restored_pr_test_bridge",
			plant: func(data []byte) ([]byte, error) {
				return addStepToWorkflowJob(data, "test", workflowStepAddition{
					Name: "Require PR mutation gate success",
					Run:  `test "${{ needs.curator-mutation-pr.result }}" = success`,
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := checkPRDefersPostLandingTests(good); err != nil {
				t.Fatalf("unmodified workflow is not the passing control: %v", err)
			}
			mutant, err := test.plant(good)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPRDefersPostLandingTests(mutant); err == nil {
				t.Fatal("pull-request deferred-suite violation admitted")
			}
		})
	}
}

func TestAddStepToWorkflowJobRequiresExistingStepsSequence(t *testing.T) {
	good := mutationInventoryWorkflowBytes(t)
	if err := checkPRDefersPostLandingTests(good); err != nil {
		t.Fatalf("unmodified workflow is not the passing control: %v", err)
	}
	if _, err := addStepToWorkflowJob(good, "linux-build-controls", workflowStepAddition{Name: "control", Run: "true"}); err != nil {
		t.Fatalf("valid workflow helper use failed: %v", err)
	}
	for _, test := range []struct {
		name     string
		workflow string
		job      string
	}{
		{name: "missing_job", workflow: "jobs:\n  present:\n    steps: []\n", job: "missing"},
		{name: "missing_steps", workflow: "jobs:\n  present:\n    runs-on: ubuntu-latest\n", job: "present"},
		{name: "non_sequence_steps", workflow: "jobs:\n  present:\n    steps:\n      run: echo\n", job: "present"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := addStepToWorkflowJob([]byte(test.workflow), test.job, workflowStepAddition{Name: "control"}); err == nil {
				t.Fatal("helper accepted a missing job or invalid steps field")
			}
		})
	}
}

func mutationInventoryWorkflowBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Claim: other workflow cmd/curator invocations remain narrowly selected.
// Limit: this line-oriented scan does not execute workflows or parse expressions.
func TestOtherWorkflowCmdCuratorInvocationsAreNarrow(t *testing.T) {
	files, err := filepath.Glob("../../.github/workflows/*.yml")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, path := range files {
		if strings.HasSuffix(path, "post-landing-mutation-tests.yml") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var blocks []string
		var current strings.Builder
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "      - ") {
				if current.Len() != 0 {
					blocks = append(blocks, current.String())
					current.Reset()
				}
			}
			if current.Len() != 0 || strings.HasPrefix(line, "      - ") {
				current.WriteString(line)
				current.WriteByte('\n')
			}
		}
		if current.Len() != 0 {
			blocks = append(blocks, current.String())
		}
		for _, block := range blocks {
			if !strings.Contains(block, "go test") || !strings.Contains(block, "cmd/curator") || strings.Contains(block, "test-gate.sh") {
				continue
			}
			checked++
			if !strings.Contains(block, "-run") {
				t.Errorf("direct cmd/curator workflow test is not narrowly selected (%s):\n%s", filepath.Base(path), block)
			}
			if strings.Contains(block, `-run "$`) || strings.Contains(block, `-run "${`) {
				t.Errorf("direct cmd/curator workflow test uses a dynamic selector (%s):\n%s", filepath.Base(path), block)
			}
			for _, name := range curatormutations.PostLandingNames() {
				if strings.Contains(block, name) {
					t.Errorf("direct workflow test selects a post-landing test (%s): %s", filepath.Base(path), name)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no direct cmd/curator workflow invocations were checked")
	}
}

// Claim: the PR guard rejects deferred test selectors resolved from workflow,
// job, step, package, and GOFLAGS inputs while admitting safe neighbors.
// Limit: this is a static workflow check, not hosted-runner evidence.
func TestPullRequestGuardResolvesDeferredSelectorsFromEnvironment(t *testing.T) {
	good := mutationInventoryWorkflowBytes(t)
	if err := checkPRDefersPostLandingTests(good); err != nil {
		t.Fatalf("unmodified workflow is not the passing control: %v", err)
	}
	makeWorkflow := func(scope, selector string) ([]byte, error) {
		data, err := addWorkflowJob(good, "selector-control", "Selector control", `go test ./... -run "$SUITE_SELECTOR"`, "", "")
		if err != nil {
			return nil, err
		}
		document, root, err := parseWorkflowDocument(data)
		if err != nil {
			return nil, err
		}
		target := root
		if scope != "workflow" {
			target = workflowMappingValue(workflowMappingValue(root, "jobs"), "selector-control")
			if scope == "step" {
				steps := workflowMappingValue(target, "steps")
				if steps == nil || len(steps.Content) != 1 {
					return nil, fmt.Errorf("selector-control job has no single step")
				}
				target = steps.Content[0]
			}
		}
		env := workflowMappingValue(target, "env")
		if env == nil {
			env = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			appendWorkflowMappingPair(target, "env", env)
		}
		appendWorkflowMappingPair(env, "SUITE_SELECTOR", workflowStringNode(selector))
		return yaml.Marshal(document)
	}
	for _, scope := range []string{"workflow", "job", "step"} {
		t.Run(scope+"_environment_refuses_deferred_selector", func(t *testing.T) {
			workflow, err := makeWorkflow(scope, curatormutations.PostLandingSelector())
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPRDefersPostLandingTests(workflow); err == nil {
				t.Fatalf("deferred selector from %s environment was admitted", scope)
			}
			workflow, err = makeWorkflow(scope, "^TestRoutine$")
			if err != nil {
				t.Fatal(err)
			}
			if err := checkPRDefersPostLandingTests(workflow); err != nil {
				t.Fatalf("ordinary selector from %s environment rejected: %v", scope, err)
			}
		})
	}
	for _, test := range []struct {
		name      string
		packageID string
		wantError bool
	}{
		{name: "deferred_package_from_environment_refuses", packageID: "./cmd/curator", wantError: true},
		{name: "ordinary_package_from_environment_admits", packageID: "./tools/mutationinventory"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addWorkflowJob(good, "package-control", "Package control", `go test "$TEST_PACKAGE"`, "", "")
			if err != nil {
				t.Fatal(err)
			}
			document, root, err := parseWorkflowDocument(workflow)
			if err != nil {
				t.Fatal(err)
			}
			job := workflowMappingValue(workflowMappingValue(root, "jobs"), "package-control")
			steps := workflowMappingValue(job, "steps")
			if steps == nil || len(steps.Content) != 1 {
				t.Fatal("package-control job has no single step")
			}
			step := steps.Content[0]
			env := workflowMappingValue(step, "env")
			if env == nil {
				env = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				appendWorkflowMappingPair(step, "env", env)
			}
			appendWorkflowMappingPair(env, "TEST_PACKAGE", workflowStringNode(test.packageID))
			workflow, err = yaml.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			err = checkPRDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("deferred package target from the step environment was admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("ordinary package target from the step environment was rejected: %v", err)
			}
		})
	}
	workflow, err := addWorkflowJob(good, "unknown-package-control", "Unknown package control", `go test "$MISSING_PACKAGE"`, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPRDefersPostLandingTests(workflow); err == nil {
		t.Fatal("unresolved dynamic package target was admitted")
	}
	for _, test := range []struct {
		name      string
		goPath    string
		selector  string
		wantError bool
	}{
		{name: "absolute_go_path_refuses_deferred_selector", goPath: "/opt/go/bin/go", selector: curatormutations.PostLandingSelector(), wantError: true},
		{name: "absolute_go_exe_path_refuses_deferred_selector", goPath: "C:/Go/bin/go.exe", selector: curatormutations.PostLandingSelector(), wantError: true},
		{name: "absolute_windows_backslash_go_path_refuses_deferred_selector", goPath: `C:\Go\bin\go.exe`, selector: curatormutations.PostLandingSelector(), wantError: true},
		{name: "absolute_go_path_admits_ordinary_selector", goPath: "/opt/go/bin/go", selector: "^TestRoutine$"},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := test.goPath + " test ./... -run '" + test.selector + "'"
			workflow, err := addWorkflowJob(good, "absolute-go-control", "Absolute Go control", run, "", "")
			if err != nil {
				t.Fatal(err)
			}
			err = checkPRDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("absolute Go invocation admitted a deferred selector")
			}
			if !test.wantError && err != nil {
				t.Fatalf("absolute Go invocation rejected a safe selector: %v", err)
			}
		})
	}
	makeGOFlagsWorkflow := func(scope, goFlags, run string) ([]byte, error) {
		data, err := addWorkflowJob(good, "goflags-control", "GOFLAGS control", run, "", "")
		if err != nil {
			return nil, err
		}
		document, root, err := parseWorkflowDocument(data)
		if err != nil {
			return nil, err
		}
		target := root
		if scope != "workflow" {
			target = workflowMappingValue(workflowMappingValue(root, "jobs"), "goflags-control")
			if scope == "step" {
				steps := workflowMappingValue(target, "steps")
				if steps == nil || len(steps.Content) != 1 {
					return nil, fmt.Errorf("GOFLAGS-control job has no single step")
				}
				target = steps.Content[0]
			}
		}
		env := workflowMappingValue(target, "env")
		if env == nil {
			env = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			appendWorkflowMappingPair(target, "env", env)
		}
		appendWorkflowMappingPair(env, "GOFLAGS", workflowStringNode(goFlags))
		return yaml.Marshal(document)
	}
	goFlagsCases := []struct {
		name      string
		scope     string
		goFlags   string
		run       string
		wantError bool
	}{
		{name: "workflow_goflags_deferred_refuses", scope: "workflow", goFlags: `-run="` + curatormutations.PostLandingSelector() + `"`, run: "go test ./tools/ciworkflow", wantError: true},
		{name: "workflow_goflags_ordinary_admits", scope: "workflow", goFlags: `-run="^TestRoutine$"`, run: "go test ./tools/ciworkflow"},
		{name: "job_goflags_test_run_alias_refuses", scope: "job", goFlags: `-test.run="` + curatormutations.PostLandingSelector() + `"`, run: "go test ./tools/ciworkflow", wantError: true},
		{name: "step_goflags_complete_skip_admits", scope: "step", goFlags: `-skip="` + curatormutations.PostLandingSelector() + `"`, run: "go test ./..."},
		{name: "command_run_overrides_goflags_selector", scope: "step", goFlags: `-run="` + curatormutations.PostLandingSelector() + `"`, run: `go test ./tools/ciworkflow -run '^TestRoutine$'`},
		{name: "dynamic_goflags_refuses", scope: "step", goFlags: "${{ env.SUITE_SELECTOR }}", run: "go test ./tools/ciworkflow", wantError: true},
		{name: "step_goflags_custom_executor_refuses", scope: "step", goFlags: "-exec=./tools/deferred-wrapper", run: "go test ./tools/ciworkflow", wantError: true},
	}
	for _, test := range goFlagsCases {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := makeGOFlagsWorkflow(test.scope, test.goFlags, test.run)
			if err != nil {
				t.Fatal(err)
			}
			err = checkPRDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("GOFLAGS-selected deferred tests were admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("safe GOFLAGS selection rejected: %v", err)
			}
		})
	}
	if !isWorkflowGoCommand(`C:\Go\bin\go.exe`) {
		t.Fatal("Windows absolute Go executable path was not recognized")
	}
	if isWorkflowGoCommand("gopher") {
		t.Fatal("non-Go command was recognized as the Go executable")
	}
	workflow, err = addWorkflowJob(good, "selector-control", "Selector control", `go test ./... -run "$MISSING_SELECTOR"`, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPRDefersPostLandingTests(workflow); err == nil {
		t.Fatal("unresolved shell selector was admitted")
	}
	workflow, err = addWorkflowJob(good, "literal-selector", "Routine", "go test ./... -run '"+curatormutations.PostLandingSelector()+"'", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPRDefersPostLandingTests(workflow); err == nil {
		t.Fatal("literal deferred selector was admitted")
	}
	for _, test := range []struct {
		name      string
		run       string
		wantError bool
	}{
		{name: "unfiltered_recursive_packages", run: "go test ./...", wantError: true},
		{name: "unfiltered_deferred_root", run: "go test ./cmd/curator", wantError: true},
		{name: "ordinary_recursive_selector", run: `go test ./... -run '^TestRoutine$'`},
		{name: "ordinary_deferred_root_selector", run: `go test ./cmd/curator -run '^TestRoutine$'`},
		{name: "recursive_with_complete_deferred_skip", run: "go test ./... -skip '" + curatormutations.PostLandingSelector() + "'"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addWorkflowJob(good, "selector-shape", "Selector shape", test.run, "", "")
			if err != nil {
				t.Fatal(err)
			}
			err = checkPRDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("deferred test execution was admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("safe test selection rejected: %v", err)
			}
		})
	}
}

// Claim: local actions and scripts are refused unless their exact target is
// admitted, while supported remote actions and guarded CI scripts remain valid.
// Limit: an allowlisted target remains trusted by path; this does not run it.
func TestPullRequestGuardChecksLocalActionsAndScripts(t *testing.T) {
	good := mutationInventoryWorkflowBytes(t)
	document, root, err := parseWorkflowDocument(good)
	if err != nil {
		t.Fatal(err)
	}
	jobs := workflowMappingValue(root, "jobs")
	changedHeredoc := false
	for i := 1; i < len(jobs.Content) && !changedHeredoc; i += 2 {
		steps := workflowMappingValue(jobs.Content[i], "steps")
		if steps == nil {
			continue
		}
		for _, step := range steps.Content {
			name := workflowMappingValue(step, "name")
			if name == nil || name.Value != "Driver qualification and reproducibility" {
				continue
			}
			run := workflowMappingValue(step, "run")
			run.Value += "\npython3 - <<'MUTATION'\nimport subprocess\nsubprocess.run(['go', 'test', './...'])\nMUTATION\n"
			changedHeredoc = true
			break
		}
	}
	if !changedHeredoc {
		t.Fatal("did not find the allowlisted Python heredoc to mutate")
	}
	mutatedWorkflow, err := yaml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPRDefersPostLandingTests(mutatedWorkflow); err == nil {
		t.Fatal("modified allowlisted interpreter heredoc was admitted")
	}
	document, root, err = parseWorkflowDocument(good)
	if err != nil {
		t.Fatal(err)
	}
	jobs = workflowMappingValue(root, "jobs")
	changedGeneratedGoRun := false
	for i := 1; i < len(jobs.Content) && !changedGeneratedGoRun; i += 2 {
		steps := workflowMappingValue(jobs.Content[i], "steps")
		if steps == nil {
			continue
		}
		for _, step := range steps.Content {
			name := workflowMappingValue(step, "name")
			if name == nil || name.Value != "Record macOS race capacity" {
				continue
			}
			run := workflowMappingValue(step, "run")
			run.Value += "\necho changed\n"
			changedGeneratedGoRun = true
			break
		}
	}
	if !changedGeneratedGoRun {
		t.Fatal("did not find the allowlisted generated Go command to mutate")
	}
	mutatedWorkflow, err = yaml.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPRDefersPostLandingTests(mutatedWorkflow); err == nil {
		t.Fatal("modified allowlisted generated Go command was admitted")
	}
	addStep := func(uses, run string, with, env map[string]string) ([]byte, error) {
		data, err := addWorkflowJob(good, "local-target-control", "Local target control", "", "", "")
		if err != nil {
			return nil, err
		}
		return addStepToWorkflowJob(data, "local-target-control", workflowStepAddition{Uses: uses, Run: run, With: with, Env: env})
	}
	for _, test := range []struct {
		name      string
		uses      string
		run       string
		with      map[string]string
		env       map[string]string
		wantError bool
	}{
		{name: "unapproved_local_action_with_selector", uses: "./.github/actions/deferred-suite", with: map[string]string{"selector": curatormutations.PostLandingSelector()}, wantError: true},
		{name: "remote_action_neighbor", uses: "actions/checkout@v4", with: map[string]string{"fetch-depth": "1"}},
		{name: "unapproved_local_script", run: "bash .github/ci/deferred-suite.sh", wantError: true},
		{name: "unapproved_local_go_tool", run: "go run ./tools/deferred-suite", wantError: true},
		{name: "allowlisted_local_go_tool_neighbor", run: "go run ./tools/sanitize-ci-evidence -source input -destination output"},
		{name: "go_test_custom_executor", run: `go test ./tools/ciworkflow -run '^TestRoutine$' -exec ./tools/deferred-wrapper`, wantError: true},
		{name: "go_test_without_custom_executor", run: `go test ./tools/ciworkflow -run '^TestRoutine$'`},
		{name: "dynamic_local_go_target", run: `go run "$LOCAL_GO_FILE"`, wantError: true},
		{name: "dynamic_interpreter_script", run: `bash "$LOCAL_SCRIPT"`, wantError: true},
		{name: "inline_shell_body", run: `bash -c 'go test ./... -run "${SUITE_SELECTOR}"'`, wantError: true},
		{name: "unapproved_interpreter_heredoc", run: "python3 - <<'PY'\nimport subprocess\nsubprocess.run(['go', 'test', './...'])\nPY", wantError: true},
		{name: "dynamic_sourced_script", run: `source "$LOCAL_SCRIPT"`, wantError: true},
		{name: "allowlisted_helper_script_neighbor", run: "bash .github/ci/isolate-go-state.sh"},
		{name: "allowlisted_sourced_helper_script_neighbor", run: "source .github/ci/isolate-go-state.sh"},
		{name: "test_gate_without_skip", run: "bash .github/ci/test-gate.sh", wantError: true},
		{name: "test_gate_with_skip", run: "bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}},
		{name: "test_gate_inline_skip_override", run: "CI_SKIP_POST_LANDING_TESTS=0 bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
		{name: "test_gate_case_insensitive_skip_override", run: "ci_skip_post_landing_tests=0 bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
		{name: "test_gate_skip_unset", run: "env -u CI_SKIP_POST_LANDING_TESTS bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
		{name: "test_gate_shell_unset", run: "unset CI_SKIP_POST_LANDING_TESTS; bash .github/ci/test-gate.sh", env: map[string]string{"CI_SKIP_POST_LANDING_TESTS": "1"}, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			workflow, err := addStep(test.uses, test.run, test.with, test.env)
			if err != nil {
				t.Fatal(err)
			}
			err = checkPRDefersPostLandingTests(workflow)
			if test.wantError && err == nil {
				t.Fatal("unapproved local action or deferred script was admitted")
			}
			if !test.wantError && err != nil {
				t.Fatalf("supported local or remote action was rejected: %v", err)
			}
		})
	}
}
