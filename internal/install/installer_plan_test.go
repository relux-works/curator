package install

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

// These subtests share the existing POSIX-installer platform gate. All archives,
// destinations, homes and staging writes are confined to the fixture prefix.
func testInstallScriptPlan(t *testing.T) {
	t.Run("locales", func(t *testing.T) {
		// Selection is tested even for locale names not installed on the host;
		// it must not depend on the machine's locale catalog.
		rows := []struct {
			name  string
			env   map[string]string
			title string
		}{
			{"ru", map[string]string{"LANG": "ru_RU.UTF-8"}, "План установки"},
			{"en", map[string]string{"LANG": "en_US.UTF-8"}, "Installation plan"},
			{"unknown", map[string]string{"LANG": "zz_ZZ.UTF-8"}, "Installation plan"},
			{"unset", nil, "Installation plan"},
			{"LC_ALL_wins", map[string]string{"LC_ALL": "ru", "LC_MESSAGES": "en", "LANG": "en"}, "План установки"},
			{"LC_MESSAGES_wins", map[string]string{"LC_MESSAGES": "ru", "LANG": "en"}, "План установки"},
			{"unknown_LC_ALL_still_wins", map[string]string{"LC_ALL": "de_DE", "LANG": "ru"}, "Installation plan"},
			{"empty_LC_ALL_falls_through", map[string]string{"LC_ALL": "", "LC_MESSAGES": "ru"}, "План установки"},
		}
		for _, row := range rows {
			t.Run(row.name, func(t *testing.T) {
				result := runInstallScriptFixture(t, installScriptFixture{args: []string{"--plan"}, env: row.env})
				assertInstallerUnchanged(t, result)
				if result.exitCode != 0 || !strings.Contains(result.output, row.title) {
					t.Fatalf("wrong plan locale: exit=%d\n%s", result.exitCode, result.output)
				}
				assertInstallerPlan(t, result, row.title)
			})
		}
	})

	t.Run("latest_metadata", func(t *testing.T) {
		// Default-unset lookup failures still show every plan field. Decline and
		// repeated failure after consent leave the scratch tree unchanged; a later
		// successful lookup installs only after consent. Timeout is a curl exit-28
		// fixture that also requires the actual connection/total timeout arguments.
		for _, row := range []struct {
			name, mode, input string
			args              []string
			tty, install      bool
			calls, exitCode   int
		}{
			{"failure_plan", "failure", "", []string{"--plan"}, false, false, 1, 0},
			{"failure_decline", "failure", "n\n", nil, true, false, 1, 1},
			{"failure_consent", "failure", "y\n", nil, true, false, 2, 1},
			{"failure_yes", "failure", "", []string{"--yes"}, false, false, 2, 1},
			{"timeout_plan", "timeout", "", []string{"--plan"}, false, false, 1, 0},
			{"timeout_consent", "timeout", "", []string{"--yes"}, false, false, 2, 1},
			{"empty_plan", "empty", "", []string{"--plan"}, false, false, 1, 0},
			{"empty_consent", "empty", "", []string{"--yes"}, false, false, 2, 1},
			{"partial_failure_consent", "partial_failure", "", []string{"--yes"}, false, false, 2, 1},
			{"recovers_after_consent", "recover", "y\n", nil, true, true, 2, 0},
		} {
			t.Run(row.name, func(t *testing.T) {
				result := runInstallScriptFixture(t, installScriptFixture{
					metadataMode: row.mode, args: row.args, tty: row.tty, input: row.input,
					env:                 map[string]string{"NO_COLOR": "1"},
					fallbackDestination: true, ghAvailable: true, ghAttestationCommand: true,
					ghSupportsSignerWorkflow: true, ghAttestationValid: true,
				})
				assertInstallerPlanVersion(t, result, "Installation plan", "latest (could not be resolved)")
				if result.exitCode != row.exitCode || strings.Count(result.curlRequests, "latest\n") != row.calls {
					t.Fatalf("exit=%d requests=%q; want exit=%d metadata calls=%d\n%s", result.exitCode, result.curlRequests, row.exitCode, row.calls, result.output)
				}
				if row.install {
					if !result.installed || !installerPlanPrecedesMutation(result.output) || !strings.Contains(result.output, "downloading curator "+fixtureVersion) {
						t.Fatalf("retry did not install resolved version after plan/consent:\n%s", result.output)
					}
				} else {
					assertInstallerUnchanged(t, result)
					if row.calls == 2 {
						assertInstallRefused(t, result, "could not determine the latest version after consent")
					} else if row.tty && !strings.Contains(result.output, "Installation cancelled") {
						t.Fatalf("failed lookup bypassed normal decline:\n%s", result.output)
					}
				}
			})
		}
		// The pinned positive control must bypass even a failing metadata API.
		t.Run("pinned_skips_lookup", func(t *testing.T) {
			result := runInstallScriptFixture(t, installScriptFixture{
				args: []string{"--plan"}, metadataMode: "failure",
				env: map[string]string{"CURATOR_VERSION": fixtureVersion},
			})
			assertInstallerUnchanged(t, result)
			assertInstallerPlan(t, result, "Installation plan")
			if result.exitCode != 0 || result.curlRequests != "" {
				t.Fatalf("pinned plan contacted metadata API: exit=%d requests=%q", result.exitCode, result.curlRequests)
			}
		})
		// The unresolved plan and metadata disclosure also stay localized.
		t.Run("failure_plan_ru", func(t *testing.T) {
			result := runInstallScriptFixture(t, installScriptFixture{
				args: []string{"--plan"}, metadataMode: "failure",
				env: map[string]string{"LANG": "ru"},
			})
			assertInstallerUnchanged(t, result)
			assertInstallerPlanVersion(t, result, "План установки", "latest (не удалось определить)")
			if result.exitCode != 0 || result.curlRequests != "latest\n" {
				t.Fatalf("failed Russian preview: exit=%d requests=%q\n%s", result.exitCode, result.curlRequests, result.output)
			}
		})
	})

	// Verification rows above remain the positive/negative security controls
	// after consent. These rows isolate consent, destination and presentation.
	for _, row := range []struct {
		name    string
		fixture installScriptFixture
		install bool
		message string
	}{
		{"plan_only", installScriptFixture{args: []string{"--plan"}}, false, "Installation plan"},
		{"plan_wins_over_yes", installScriptFixture{args: []string{"--yes", "--plan"}}, false, "Installation plan"},
		{"plan_fallback_without_mkdir", installScriptFixture{args: []string{"--plan"}, fallbackDestination: true}, false, "Installation plan"},
		{"yes_flag", installScriptFixture{args: []string{"--yes"}}, true, "Installation plan"},
		{"yes_env", installScriptFixture{env: map[string]string{"CURATOR_YES": "1"}}, true, "Installation plan"},
		{"yes_fallback_creates_after_plan", installScriptFixture{args: []string{"--yes"}, fallbackDestination: true}, true, "Installation plan"},
		{"confirmed", installScriptFixture{tty: true, input: "y\n"}, true, "Continue?"},
		{"confirmed_ru", installScriptFixture{tty: true, input: "да\n", env: map[string]string{"LANG": "ru"}}, true, "Продолжить?"},
		{"declined", installScriptFixture{tty: true, input: "n\n", fallbackDestination: true}, false, "Installation cancelled"},
		{"declined_ru", installScriptFixture{tty: true, input: "нет\n", env: map[string]string{"LANG": "ru"}}, false, "Установка отменена"},
		{"blank_declines", installScriptFixture{tty: true, input: "\n"}, false, "Installation cancelled"},
		{"insecure_plan_is_honest", installScriptFixture{args: []string{"--plan"}, skipVerification: true}, false, "Verification DISABLED"},
		{"unknown_option", installScriptFixture{args: []string{"--wat"}}, false, "unknown option"},
	} {
		t.Run(row.name, func(t *testing.T) {
			fixture := row.fixture
			fixture.ghAvailable, fixture.ghAttestationCommand = true, true
			fixture.ghSupportsSignerWorkflow, fixture.ghAttestationValid = true, true
			result := runInstallScriptFixture(t, fixture)
			if !strings.Contains(result.output, row.message) {
				t.Fatalf("missing %q:\n%s", row.message, result.output)
			}
			if row.install {
				if result.exitCode != 0 || !result.installed || result.ghCalls == "" {
					t.Fatalf("confirmed install failed: exit=%d\n%s", result.exitCode, result.output)
				}
				if !installerPlanPrecedesMutation(result.output) {
					t.Fatalf("filesystem change preceded the complete plan:\n%s", result.output)
				}
				if !fixture.tty && (strings.Contains(result.output, "Continue?") || strings.Contains(result.output, "Продолжить?")) {
					t.Fatalf("--yes/CURATOR_YES still prompted:\n%s", result.output)
				}
			} else {
				assertInstallerUnchanged(t, result)
				if fixture.tty && result.exitCode == 0 {
					t.Fatal("decline must return a non-success status")
				}
			}
		})
	}

	t.Run("terminal_colour", func(t *testing.T) {
		for _, row := range []struct {
			name   string
			tty    bool
			env    map[string]string
			colour bool
		}{
			{"pipe", false, nil, false},
			{"tty", true, nil, true},
			{"NO_COLOR", true, map[string]string{"NO_COLOR": "1"}, false},
			{"empty_NO_COLOR", true, map[string]string{"NO_COLOR": ""}, false},
		} {
			t.Run(row.name, func(t *testing.T) {
				result := runInstallScriptFixture(t, installScriptFixture{args: []string{"--plan"}, tty: row.tty, env: row.env})
				assertInstallerUnchanged(t, result)
				if result.exitCode != 0 || strings.Contains(result.output, "\x1b[") != row.colour {
					t.Fatalf("colour=%t, want %t, exit=%d: %q", strings.Contains(result.output, "\x1b["), row.colour, result.exitCode, result.output)
				}
			})
		}
	})

	t.Run("early_mutation_monitor_control", func(t *testing.T) {
		// A single planted mktemp before the script proves that both ordering
		// instrumentation and scratch-tree comparison detect a real violation.
		result := runInstallScriptFixture(t, installScriptFixture{args: []string{"--plan"}, plantEarlyMutation: true})
		if result.exitCode != 0 || installerPlanPrecedesMutation(result.output) || len(result.changes) == 0 || !strings.Contains(result.mutations, "mktemp") {
			t.Fatalf("known early write escaped the monitor: changes=%v\n%s", result.changes, result.output)
		}
	})
}

func assertInstallerPlan(t *testing.T, result installScriptResult, title string) {
	t.Helper()
	assertInstallerPlanVersion(t, result, title, fixtureVersion)
}

func assertInstallerPlanVersion(t *testing.T, result installScriptResult, title, version string) {
	t.Helper()
	for _, value := range []string{title, "curator", version, result.installPath,
		"https://github.com/relux-works/curator#readme",
		"https://github.com/relux-works/curator/blob/main/docs/book/README.md",
		"https://github.com/relux-works/curator-spec", "GitHub build attestation", "cosign", "SHA-256"} {
		if !strings.Contains(result.output, value) {
			t.Errorf("plan omits %q:\n%s", value, result.output)
		}
	}
	if strings.Contains(result.curlRequests, "latest\n") {
		disclosure := "Before consent: read-only latest metadata GET, no file writes (5s timeout)."
		if title == "План установки" {
			disclosure = "До подтверждения: read-only GET метаданных latest, без записи файлов (timeout 5 с)."
		}
		if !strings.Contains(result.output, disclosure) {
			t.Errorf("plan omits metadata disclosure %q:\n%s", disclosure, result.output)
		}
	}
	if title == "План установки" {
		if !strings.Contains(result.output, "Нужен sudo:     нет") {
			t.Fatal("Russian plan must explicitly say sudo is not needed")
		}
	} else if !strings.Contains(result.output, "Needs sudo:     no") {
		t.Fatal("English plan must explicitly say sudo is not needed")
	}
	if strings.Contains(result.output, "\x1b[") {
		t.Fatal("redirected output contains terminal colour")
	}
}

func assertInstallerUnchanged(t *testing.T, result installScriptResult) {
	t.Helper()
	// A read-only latest metadata GET is allowed; artifact requests and all
	// installer filesystem writes remain forbidden before consent or resolution.
	artifactRequests := strings.ReplaceAll(result.curlRequests, "latest\n", "")
	if result.installed || len(result.changes) != 0 || result.mutations != "" || artifactRequests != "" || result.ghCalls != "" || result.cosignCalls != "" {
		t.Fatalf("plan/decline had effects: installed=%t changes=%v mutations=%q requests=%q\n%s",
			result.installed, result.changes, result.mutations, result.curlRequests, result.output)
	}
}

func installerPlanPrecedesMutation(output string) bool {
	planEnd := strings.Index(output, "SHA-256")
	mutation := strings.Index(output, "FIXTURE_MUTATION")
	return planEnd >= 0 && mutation > planEnd
}

func installerFixtureTree(t *testing.T, root, ignoredLog string) map[string]string {
	t.Helper()
	tree := make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == ignoredLog {
			return nil // The curl stub's request log is observation, not installer data.
		}
		value := info.Mode().String()
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += fmt.Sprintf("/%d/%s", info.ModTime().UnixNano(), data)
		}
		tree[path] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func installerFixtureChanges(before, after map[string]string) []string {
	var changes []string
	for path, value := range before {
		if after[path] != value {
			changes = append(changes, path)
		}
	}
	for path := range after {
		if _, exists := before[path]; !exists {
			changes = append(changes, path)
		}
	}
	sort.Strings(changes)
	return changes
}

// Reply only after the prompt arrives, keeping stdin open until child exit.
// BSD script otherwise injects EOF before a preloaded pipe's answer.
type installerTTYOutput struct {
	mu       sync.Mutex
	output   bytes.Buffer
	input    io.Writer
	answer   string
	answered bool
}

func (out *installerTTYOutput) Write(data []byte) (int, error) {
	out.mu.Lock()
	defer out.mu.Unlock()
	n, err := out.output.Write(data)
	if !out.answered && out.answer != "" && strings.Contains(out.output.String(), "? [") {
		out.answered = true
		_, err = io.WriteString(out.input, out.answer)
	}
	return n, err
}

func runInstallerTTY(t *testing.T, cmd *exec.Cmd, answer string) ([]byte, error) {
	t.Helper()
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = input.Close() }()
	out := &installerTTYOutput{input: input, answer: answer}
	cmd.Stdout, cmd.Stderr = out, out
	err = cmd.Run()
	return bytes.Clone(out.output.Bytes()), err
}
