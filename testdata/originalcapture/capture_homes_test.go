package harnessparity

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var (
	nativeParentWorker       = flag.Bool("native-parent-worker", false, "")
	nativeParentWorkerCase   = flag.String("native-parent-worker-case", "", "")
	nativeParentWorkerRoot   = flag.String("native-parent-worker-root", "", "")
	nativeParentWorkerResult = flag.String("native-parent-worker-result", "", "")
)

func captureCallerSelectorNames() []string {
	var names []string
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		normalized := strings.ToUpper(name)
		if normalized == "CODEX_HOME" || normalized == "CLAUDE_CONFIG_DIR" || strings.HasPrefix(normalized, "A2A_") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func captureCallerEnvironmentError() error {
	names := captureCallerSelectorNames()
	if len(names) == 0 {
		return nil
	}
	return fmt.Errorf("capture-driver/caller-selector-set:%s", strings.Join(names, ","))
}

func onlyHomeAndPathEnvironment() bool {
	entries := os.Environ()
	if len(entries) != 2 {
		return false
	}
	names := map[string]bool{}
	for _, entry := range entries {
		name, _, _ := strings.Cut(entry, "=")
		names[name] = true
	}
	return len(names) == 2 && names["HOME"] && names["PATH"]
}

func TestCaptureDriverRejectsCallerSelectors(t *testing.T) {
	for _, selector := range []string{"CODEX_HOME", "CLAUDE_CONFIG_DIR", "A2A_CAPTURE_CONTROL", "A2A_UNLISTED_SELECTOR"} {
		t.Run(selector, func(t *testing.T) {
			home := t.TempDir()
			cmd := exec.Command(os.Args[0], "-test.run=^TestFreezeOriginalBaseline$")
			cmd.Env = []string{"HOME=" + home, "PATH=" + os.Getenv("PATH"), selector + "=fixture-control"}
			output, err := cmd.CombinedOutput()
			if err == nil || !bytes.Contains(output, []byte("capture-driver/caller-selector-set:"+selector)) {
				t.Fatalf("capture driver did not refuse selector %s before starting", selector)
			}
		})
	}
}

type repeatFixtureObservation struct {
	Observation            string        `json:"observation"`
	FixtureRootBinding     string        `json:"fixture_root_binding"`
	FixtureRootOrdinal     int           `json:"fixture_root_ordinal"`
	FixtureRootPathOmitted bool          `json:"fixture_root_path_omitted"`
	DistinctRootsInMemory  bool          `json:"distinct_roots_in_memory"`
	ResultSHA256           string        `json:"result_sha256"`
	Result                 fixtureResult `json:"result"`
}

func writeRepeatFixtureObservation(dir, name string, observation repeatFixtureObservation) error {
	payload, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".json"), append(payload, '\n'), 0600)
}

// This original-checkout capture never imports any new harness implementation.
func TestFreezeOriginalBaseline(t *testing.T) {
	if *nativeParentWorker {
		if !onlyHomeAndPathEnvironment() {
			t.Fatal("capture-driver/native-parent-environment-not-isolated")
		}
		if *nativeParentWorkerCase == "" || *nativeParentWorkerRoot == "" || *nativeParentWorkerResult == "" {
			t.Fatal("capture-driver/native-parent-worker-arguments-invalid")
		}
		var c fixtureCase
		for _, candidate := range fixtureCases() {
			if candidate.Name == *nativeParentWorkerCase {
				c = candidate
				break
			}
		}
		if c.Name == "" || c.State != "native-parent" {
			t.Fatal("capture-driver/native-parent-worker-case-invalid")
		}
		actual, _ := runLegacyFixtureAtRoot(t, c, *nativeParentWorkerRoot, true)
		payload, err := json.MarshalIndent(actual, "", "  ")
		if err != nil {
			t.Fatal("native-parent result encoding failed")
		}
		if err := os.WriteFile(*nativeParentWorkerResult, append(payload, '\n'), 0o600); err != nil {
			t.Fatal("native-parent result write failed")
		}
		return
	}
	out := os.Getenv("HARNESS_CAPTURE_OUT")
	if out == "" {
		t.Fatal("explicit private output required")
	}
	for _, c := range fixtureCases() {
		t.Run(c.Name, func(t *testing.T) {
			var actual fixtureResult
			if c.State == "native-parent" {
				actual = runNativeParentFixtureIsolated(t, c, out)
			} else {
				actual = runLegacyFixture(t, c)
			}
			payload, err := json.MarshalIndent(actual, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, c.Name+".json"), append(payload, '\n'), 0600); err != nil {
				t.Fatal(err)
			}
			if !actual.Executed || actual.Case != c {
				t.Fatal("capture-driver/fixture-result-incomplete")
			}
			if c.State == "native-parent" && (len(actual.Document) == 0 || len(actual.Tree) == 0 || len(actual.Rendered) == 0) {
				t.Fatal("capture-driver/native-parent-output-incomplete")
			}
			if c.State == "empty" || c.State == "user-settings" || c.State == "current" || c.State == "native-parent" {
				if actual.Refusal != "" {
					t.Fatalf("required successful fixture refused: %s", actual.Refusal)
				}
			}
		})
	}
	t.Run("repeat-determinism", func(t *testing.T) {
		c := fixtureCase{Name: "claude_code-empty", ID: "claude_code", Operation: "managed", State: "empty"}
		repeatOut := os.Getenv("HARNESS_REPEAT_CAPTURE_OUT")
		if repeatOut == "" {
			t.Fatal("explicit private repeat output required")
		}
		if err := os.MkdirAll(repeatOut, 0o700); err != nil {
			t.Fatal("repeat output initialization failed")
		}
		first, firstRoot := runLegacyFixtureWithRoot(t, c)
		second, secondRoot := runLegacyFixtureWithRoot(t, c)
		distinctRoots := firstRoot != secondRoot
		bindingSource := "t.TempDir -> filepath.EvalSymlinks"
		firstBytes, err := json.MarshalIndent(first, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		secondBytes, err := json.MarshalIndent(second, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := writeRepeatFixtureObservation(repeatOut, "first", repeatFixtureObservation{
			Observation: "first", FixtureRootBinding: bindingSource, FixtureRootOrdinal: 1,
			FixtureRootPathOmitted: true, DistinctRootsInMemory: distinctRoots,
			ResultSHA256: hashBytes(append(firstBytes, '\n')), Result: first,
		}); err != nil {
			t.Fatal("first repeat observation export failed")
		}
		if err := writeRepeatFixtureObservation(repeatOut, "second", repeatFixtureObservation{
			Observation: "second", FixtureRootBinding: bindingSource, FixtureRootOrdinal: 2,
			FixtureRootPathOmitted: true, DistinctRootsInMemory: distinctRoots,
			ResultSHA256: hashBytes(append(secondBytes, '\n')), Result: second,
		}); err != nil {
			t.Fatal("second repeat observation export failed")
		}
		if !distinctRoots {
			t.Fatal("repeat fixture roots were not independent")
		}
		if err := compareResult(first, second); err != nil {
			t.Fatal(err)
		}
	})
}

func runNativeParentFixtureIsolated(t *testing.T, c fixtureCase, out string) fixtureResult {
	t.Helper()
	if c.State != "native-parent" {
		t.Fatal("native-parent isolation requested for another fixture")
	}
	parentHome := os.Getenv("HOME")
	trustedPath := os.Getenv("PATH")
	if parentHome == "" || trustedPath == "" {
		t.Fatal("native-parent isolated caller inputs unavailable")
	}
	workerHome, err := os.MkdirTemp(parentHome, "native-parent-worker-")
	if err != nil {
		t.Fatal("native-parent isolated home setup failed")
	}
	defer os.RemoveAll(workerHome)
	workerRoot := filepath.Join(workerHome, "fixture-root")
	workerResult := filepath.Join(filepath.Dir(out), "native-parent-worker-result.json")
	if _, err := os.Lstat(workerResult); err == nil || !os.IsNotExist(err) {
		t.Fatal("native-parent worker result path is not fresh")
	}
	defer os.Remove(workerResult)
	envPath, err := exec.LookPath("env")
	if err != nil {
		t.Fatal("env utility unavailable for native-parent isolation")
	}
	cmd := exec.Command(
		envPath, "-i", "HOME="+workerHome, "PATH="+trustedPath,
		os.Args[0], "-test.run=^TestFreezeOriginalBaseline$",
		"-native-parent-worker=true",
		"-native-parent-worker-case="+c.Name,
		"-native-parent-worker-root="+workerRoot,
		"-native-parent-worker-result="+workerResult,
	)
	if _, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("native-parent isolated capture failed: %v", err)
	}
	payload, err := os.ReadFile(workerResult)
	if err != nil {
		t.Fatal("native-parent worker result unavailable")
	}
	var result fixtureResult
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatal("native-parent worker result invalid")
	}
	return result
}
