package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/protocoljson"
)

type claudeEnvVector struct {
	ID          string `json:"id"`
	Claim       string `json:"claim"`
	Entry       string `json:"entry"`
	Input       string `json:"input_json"`
	Want        string `json:"want"`
	SchemaValid bool   `json:"schema_valid"`
}

type claudeEnvPlant struct {
	ID      string `json:"id"`
	Entry   string `json:"entry"`
	Valid   string `json:"valid_json"`
	Invalid string `json:"invalid_json"`
	Want    string `json:"want"`
}

func claudeEnvVerdict(err error) string {
	if err == nil {
		return "accept"
	}
	var selection launcherClaudeEnvError
	if errors.As(err, &selection) {
		return selection.code
	}
	return "unrelated-config-error"
}

const claudeEnvControl = `{"BASH_MAX_OUTPUT_LENGTH":"8000"}`

func claudeEnvCarrier(member string) string {
	return `{"schema_version":9,"skills_root":"x","projects":{},"extensions":{"claude_env":` + member + `}}`
}

// File-level JSON failures are distinct from member semantic refusals. First
// prove that this exact carrier accepts a valid member. Otherwise a broken
// setup must never count as detecting a declaration vector, even when its
// malformed member would mask the unrelated carrier defect in Load.
func claudeEnvDeclarationVerdict(t *testing.T, member string, carrier func(string) string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "selected.json")
	load := func(text string) (*Config, error) {
		t.Helper()
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load(path, nil)
		if err != nil && cfg != nil {
			t.Fatal("refusal returned a usable configuration")
		}
		after, readErr := os.ReadFile(path)
		if readErr != nil || string(after) != text {
			t.Fatal("loader changed source")
		}
		return cfg, err
	}
	if _, err := load(carrier(claudeEnvControl)); err != nil {
		return "invalid-containing-file"
	}
	_, err := load(carrier(member))
	if protocoljson.Validate([]byte(member)) != nil {
		var malformed configJSONError
		if !errors.As(err, &malformed) || malformed.path != path ||
			LauncherDiagnostic(err).Error() != "malformed JSON in config "+path {
			return "unexpected-file-diagnostic"
		}
		t.Log("classification: invalid declaration JSON; neutral containing-file diagnostic (source spec note pending)")
		return ClaudeEnvInvalid
	}
	t.Log("classification: member semantic verdict")
	return claudeEnvVerdict(err)
}

// Every pinned declaration runs through the actual selected machine-file
// loader. Projection vectors exercise the strict contribution reader, not a
// mock process. Neither passing path is evidence of native Claude execution.
func TestLauncherClaudeEnvPublishedCorpus(t *testing.T) {
	root := filepath.Join("testdata", "machine-claude-env-v1")
	var vectors []claudeEnvVector
	data, err := os.ReadFile(filepath.Join(root, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 41 {
		t.Fatal("truncated published corpus")
	}
	var plants []claudeEnvPlant
	data, err = os.ReadFile(filepath.Join(root, "plants.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &plants); err != nil {
		t.Fatal(err)
	}
	if len(plants) != 6 {
		t.Fatal("six plants required")
	}
	for _, p := range plants {
		vectors = append(vectors, claudeEnvVector{ID: p.ID + "-control", Entry: p.Entry, Input: p.Valid, Want: "accept"}, claudeEnvVector{ID: p.ID + "-plant", Entry: p.Entry, Input: p.Invalid, Want: p.Want})
	}
	for _, v := range vectors {
		t.Run(v.ID, func(t *testing.T) {
			var err error
			var verdict string
			switch v.Entry {
			case "declaration":
				verdict = claudeEnvDeclarationVerdict(t, v.Input, claudeEnvCarrier)
			case "projection":
				var got LauncherClaudeEnvContribution
				got, err = ParseClaudeEnvContribution([]byte(v.Input))
				if err != nil && !reflect.DeepEqual(got, LauncherClaudeEnvContribution{}) {
					t.Fatal("refusal returned a partial contribution")
				}
			default:
				t.Fatal("unknown corpus entry")
			}
			if verdict == "" {
				verdict = claudeEnvVerdict(err)
			}
			if verdict != v.Want {
				t.Fatalf("verdict %s, want %s", verdict, v.Want)
			}
		})
	}
}

// Each read reselects the machine bytes and source identity. The contribution
// never copies any other environment entry; valid shared selections keep all
// admitted non-Claude targets usable with an empty contribution.
func TestLauncherClaudeEnvProjectionAndReselection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "selected.json")
	for _, member := range []string{"", `,"extensions":{}`, `,"extensions":{"claude_env":{}}`, `,"extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"8000"}}`} {
		text := `{"schema_version":9,"skills_root":"x","projects":{}` + member + `}`
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load(path, nil)
		if err != nil {
			t.Fatal(err)
		}
		projection := cfg.LauncherProjection()
		if projection.Schema != "curator.launcher-config/3" || projection.ConfigSchema != 9 || projection.ConfigPath != path {
			t.Fatal("wrong carrier identity")
		}
		configured := strings.Contains(member, "8000")
		for i, runtime := range []string{"claude_code", "codex_cli", "opencode", "pi", "muse"} {
			got, err := cfg.ClaudeEnvContribution(runtime)
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			if runtime == "claude_code" && configured {
				want = "8000"
			}
			if got.Runtime != runtime || got.Env.BashMaxOutputLength != want {
				t.Fatal("incorrect runtime contribution")
			}
			payload, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParseClaudeEnvContribution(payload); err != nil {
				t.Fatal("producer/reader disagreement")
			}
			state := "absent"
			if configured {
				state = "runtime-ineligible"
				if runtime == "claude_code" {
					state = "configured"
				}
			}
			row := cfg.ClaudeEnvStatus()[i]
			if row.Runtime != runtime || row.State != state || row.ConfigPath != path {
				t.Fatal("wrong source or eligibility status")
			}
		}
		if projection.Extensions.ClaudeEnv != nil {
			projection.Extensions.ClaudeEnv.BashMaxOutputLength = "1"
			if cfg.Extensions.ClaudeEnv.BashMaxOutputLength == "1" {
				t.Fatal("projection aliases selected state")
			}
		}
	}
}

func TestLauncherClaudeEnvClosedCarrierAndSource(t *testing.T) {
	for _, schema := range []int{1, 2, 3, 5, 6, 7, 8} {
		path := filepath.Join(t.TempDir(), "old.json")
		text := fmt.Sprintf(`{"schema_version":%d,"skills_root":"x","projects":{},"extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"8000"}}}`, schema)
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		if cfg, err := Load(path, nil); cfg != nil || err == nil {
			t.Fatalf("old schema %d admitted new member", schema)
		}
	}
	// Prior selections retain both keychain/file forms, disabled/enabled Host
	// notification forms, proxy, credential-store and hash-version capabilities.
	for _, store := range []string{`{"type":"keychain","service":"example","account":"operator"}`, `{"type":"file","path":"/nonexistent/credential-reference"}`} {
		for _, notification := range []string{`{"backend":"off"}`, notificationSelectionFixture} {
			path := filepath.Join(t.TempDir(), "selected.json")
			text := `{"schema_version":9,"skills_root":"x","projects":{},"preferred_locale":"en","extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"8000"},"proxy":{"profile":"filtered"},"claude_auth":{"type":"oauth","store":` + store + `},"notification":` + notification + `},"environments":{"credential_store":{"sample":{"codex_cli":"file"}}}}`
			if err := os.WriteFile(path, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path, nil)
			if err != nil {
				t.Fatal(err)
			}
			project := filepath.Join(t.TempDir(), "project")
			if err := os.Mkdir(project, 0700); err != nil {
				t.Fatal(err)
			}
			if err := AddProject(path, "sample", project, nil); err != nil {
				t.Fatal(err)
			}
			updated, err := Load(path, nil)
			if err != nil {
				t.Fatal(err)
			}
			if updated.Schema != 9 || updated.PreferredLocale != "en" || !reflect.DeepEqual(cfg.Extensions, updated.Extensions) || !reflect.DeepEqual(cfg.Env.CredentialStore, updated.Env.CredentialStore) {
				t.Fatal("normal edit lost predecessor selection")
			}
		}
	}
	_, _, err := loadWithSystem(t, `{"schema_version":9,"skills_root":"x","projects":{}}`, `{"schema_version":2,"locked":[],"extensions":{"claude_env":{"BASH_MAX_OUTPUT_LENGTH":"8000"}}}`)
	if err == nil || LauncherDiagnostic(err).Error() != "extensions: launcher selections belong only to the selected machine configuration" {
		t.Fatal("system supplied forbidden authority or unsafe refusal")
	}
	// Dangerous variable names cannot become generic environment passthrough.
	for _, name := range []string{"BASH_ENV", "ENV", "LD_PRELOAD", "NODE_OPTIONS", "PATH", "HOME", "bash_max_output_length"} {
		_, err := parseLauncherClaudeEnv(map[string]any{ClaudeEnvSetting: "8000", name: "DO_NOT_ECHO"})
		if claudeEnvVerdict(err) != ClaudeEnvNameNotAllowed {
			t.Fatal("unexpected dangerous-name verdict")
		}
	}
}

func TestLauncherClaudeEnvDeterministicRefusalPrecedence(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{`{"runtime":"unknown","env":{"UNLISTED":"DO_NOT_ECHO"}}`, ClaudeEnvInvalid},
		{`{"runtime":"codex_cli","env":{"UNLISTED":"DO_NOT_ECHO","BASH_MAX_OUTPUT_LENGTH":"0"}}`, ClaudeEnvNameNotAllowed},
		{`{"runtime":"codex_cli","env":{"BASH_MAX_OUTPUT_LENGTH":"0"}}`, ClaudeEnvValueNonpositive},
		{`{"runtime":"codex_cli","env":{"BASH_MAX_OUTPUT_LENGTH":"8000"}}`, ClaudeEnvRuntimeIneligible},
	} {
		for i := 0; i < 50; i++ {
			_, err := ParseClaudeEnvContribution([]byte(tc.input))
			if claudeEnvVerdict(err) != tc.want {
				t.Fatal("unstable precedence")
			}
		}
	}
}
