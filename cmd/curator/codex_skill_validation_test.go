package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
)

// The real dispatcher emits exactly one bounded JSON object for content,
// setup and recognized-handler usage results, including flags after errors.
// Global/ordinary help remains normal help and is not a content verdict.
func TestCodexSkillValidationJSON(t *testing.T) {
	dir := t.TempDir()
	data := []byte("---\nname: unrelated-name\ndescription: Complete description\n---\nInstructions\n")
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), data, 0o640); err != nil {
		t.Fatal(err)
	}
	missing := t.TempDir()
	cases := []struct {
		name    string
		args    []string
		exit    int
		outcome string
		issue   string
	}{
		{"valid", []string{dir, "--json"}, 0, "valid", ""},
		{"flag-first", []string{"--json", dir}, 0, "valid", ""},
		{"missing-file", []string{missing, "--json"}, 3, "setup_error", "skill_io"},
		{"no-directory", []string{"--json"}, 2, "usage_error", "usage_invalid"},
		{"extra-directory", []string{dir, missing, "--json"}, 2, "usage_error", "usage_invalid"},
		{"unsupported-option-first", []string{"--unsupported", dir, "--json"}, 2, "usage_error", "usage_invalid"},
		{"unsupported-option-last", []string{"--json", dir, "--unsupported"}, 2, "usage_error", "usage_invalid"},
		{"takeover-option", []string{dir, "--takeover", "--json"}, 2, "usage_error", "usage_invalid"},
		{"takeover-single-dash", []string{dir, "-takeover", "--json"}, 2, "usage_error", "usage_invalid"},
		{"takeover-true-json-first", []string{"--json", dir, "--takeover=true"}, 2, "usage_error", "usage_invalid"},
		{"takeover-false", []string{dir, "--takeover=false", "--json"}, 2, "usage_error", "usage_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"skill", "validate-codex"}, tc.args...)
			exit, stdout, stderr := runProfile(t, stubConfigSource{path: filepath.Join(missing, "absent-config.json")}, args...)
			if exit != tc.exit || stderr != "" {
				t.Fatalf("exit=%d stderr=%q; want %d/empty", exit, stderr, tc.exit)
			}
			issues := []map[string]any{}
			if tc.issue != "" {
				issues = append(issues, map[string]any{"code": tc.issue})
			}
			want := map[string]any{"schema_version": float64(1), "contract": "codex-quick-validate-v1", "outcome": tc.outcome, "issues": issues}
			var decoded map[string]any
			if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
				t.Fatalf("result is not one JSON object: %v; %q", err, stdout)
			}
			wantBytes, _ := json.Marshal(want)
			var normalized map[string]any
			_ = json.Unmarshal(wantBytes, &normalized)
			if !reflect.DeepEqual(decoded, normalized) || !strings.HasSuffix(stdout, "\n") || strings.Count(stdout, "\n") != 1 {
				t.Fatalf("result=%q; want %#v", stdout, normalized)
			}
			if strings.Contains(stdout, dir) || strings.Contains(stdout, missing) || strings.Contains(stdout, "unrelated-name") {
				t.Fatal("result leaked an input value or absolute path")
			}
		})
	}
	for _, args := range [][]string{{"skill", "validate-codex", "--help", "--json"}, {"help", "skill", "validate-codex"}} {
		code, stdout, stderr := runProfile(t, stubConfigSource{}, args...)
		if code != 0 || !strings.Contains(stdout, "skill validate-codex") || strings.HasPrefix(stdout, "{") || stderr != "" {
			t.Fatalf("help became a content result: %d %q %q", code, stdout, stderr)
		}
	}
	after, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil || string(after) != string(data) {
		t.Fatalf("CLI changed input: %v", err)
	}
	entries, err := os.ReadDir(missing)
	if err != nil || len(entries) != 0 {
		t.Fatalf("validation wrote configuration: %v, %v", entries, err)
	}
}

// Plain takeover refusals retain the shared canonical diagnostic for every
// spelling. A literal --json after -- does not select JSON; no-flag validation
// is the positive control, and configuration access would panic.
func TestCodexSkillValidationTakeoverPlainAdmission(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	data := []byte("---\nname: sample\ndescription: Complete\n---\n")
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := run([]string{"skill", "validate-codex", dir}, codexValidationForbiddenConfig{}, &stdout, &stderr); code != 0 || stdout.String() != "Codex skill: valid\n" || stderr.Len() != 0 {
		t.Fatalf("positive validation: exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	for _, flag := range []string{"--takeover", "--takeover=true", "--takeover=false", "-takeover"} {
		for _, suffix := range [][]string{nil, {"--help"}, {"--", "--json"}} {
			stdout.Reset()
			stderr.Reset()
			args := append([]string{"skill", "validate-codex", dir, flag}, suffix...)
			code := run(args, codexValidationForbiddenConfig{}, &stdout, &stderr)
			if code != 2 || stdout.Len() != 0 || !strings.Contains(stderr.String(), "curator: --takeover applies only") {
				t.Fatalf("%v: exit=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
			}
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(data) {
		t.Fatalf("refusal changed input: %v", err)
	}
}

type codexValidationForbiddenConfig struct{}

func (codexValidationForbiddenConfig) Path() string {
	panic("Codex skill validation accessed machine configuration path")
}

func (codexValidationForbiddenConfig) Load(func(string)) (*config.Config, error) {
	panic("Codex skill validation loaded machine configuration")
}

// Configuration access would panic, detecting accidental provider fallback or
// manager initialization on both a valid input and neighboring refusal paths.
func TestCodexSkillValidationNeedsNoMachineConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte("---\nname: sample\ndescription: Complete\n---\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		args []string
		exit int
	}{
		{[]string{dir}, 0},
		{[]string{dir, "--json"}, 0},
		{[]string{"--json"}, 2},
		{[]string{dir, "--unsupported", "--json"}, 2},
		{[]string{filepath.Join(dir, "absent"), "--json"}, 3},
	} {
		var stdout, stderr strings.Builder
		code := run(append([]string{"skill", "validate-codex"}, tc.args...), codexValidationForbiddenConfig{}, &stdout, &stderr)
		if code != tc.exit {
			t.Fatalf("%v: exit=%d want=%d; %s", tc.args, code, tc.exit, stderr.String())
		}
	}
}

// Both output modes must agree on the same content refusal; this witness
// prevents a success/null result from satisfying the new command's contract.
func TestCodexSkillValidationRefusalAndPlainOutput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: Invalid\ndescription: Complete\n---\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	for _, jsonOut := range []bool{false, true} {
		args := []string{"skill", "validate-codex", dir}
		if jsonOut {
			args = append(args, "--json")
		}
		exit, stdout, stderr := runProfile(t, stubConfigSource{}, args...)
		if exit != 1 {
			t.Fatalf("invalid name accepted: exit=%d, stdout=%q stderr=%q", exit, stdout, stderr)
		}
		if jsonOut {
			if stdout != "{\"schema_version\":1,\"contract\":\"codex-quick-validate-v1\",\"outcome\":\"invalid\",\"issues\":[{\"code\":\"name_invalid\"}]}\n" || stderr != "" {
				t.Fatalf("wrong bounded refusal: %q %q", stdout, stderr)
			}
		} else if !strings.Contains(stderr, "name") || stdout != "" {
			t.Fatalf("missing useful plain diagnostic: %q %q", stdout, stderr)
		}
	}
}
