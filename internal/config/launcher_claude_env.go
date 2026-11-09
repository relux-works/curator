package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// ClaudeEnvSetting is the only admitted machine-owned Claude child setting.
const ClaudeEnvSetting = "BASH_MAX_OUTPUT_LENGTH"

// Claude environment refusal codes never contain input names or values.
const (
	ClaudeEnvInvalid           = "claude_env_invalid"
	ClaudeEnvNameNotAllowed    = "claude_env_name_not_allowed"
	ClaudeEnvValueNotInteger   = "claude_env_value_not_integer"
	ClaudeEnvValueNonpositive  = "claude_env_value_nonpositive"
	ClaudeEnvValueOverBound    = "claude_env_value_over_bound"
	ClaudeEnvRuntimeIneligible = "claude_env_runtime_ineligible"
)

// LauncherClaudeEnv is a closed declaration, not a generic environment map.
// The empty object selects no override. It does not change native settings.
type LauncherClaudeEnv struct {
	BashMaxOutputLength string `json:"BASH_MAX_OUTPUT_LENGTH,omitempty"`
}

type launcherClaudeEnvError struct{ code string }

func (e launcherClaudeEnvError) Error() string { return "extensions.claude_env: " + e.code }
func claudeEnvError(code string) error         { return launcherClaudeEnvError{code: code} }

// configJSONError renders only the source path, never parser details or input.
type configJSONError struct{ path string }

func (e configJSONError) Error() string { return fmt.Sprintf("malformed JSON in config %s", e.path) }

// LauncherDiagnostic preserves the loader's input-free diagnostic and identity.
// Discovery refusal is an independent property, not a rendered member code.
func LauncherDiagnostic(err error) error {
	return err
}

type launcherSelectionRefusal struct{ error }

func (e launcherSelectionRefusal) Unwrap() error { return e.error }

// IsLauncherSelectionRefusal identifies invalid explicit input that must not
// fall through Core's optional-config umbrella-provider discovery route.
func IsLauncherSelectionRefusal(err error) bool {
	var selection launcherClaudeEnvError
	var refusal launcherSelectionRefusal
	return errors.As(err, &selection) || errors.As(err, &refusal)
}

// A known explicit declaration still belongs to this refusal boundary when
// its containing version is wrong. It must not fall through optional-config
// provider discovery merely because version validation runs before members.
func hasMachineClaudeEnv(data map[string]any) bool {
	extensions, ok := data["extensions"].(map[string]any)
	if !ok {
		return false
	}
	_, present := extensions["claude_env"]
	return present
}

func markLauncherSelectionRefusal(err error) error {
	if err == nil || IsLauncherSelectionRefusal(err) {
		return err
	}
	return launcherSelectionRefusal{err}
}

// hasClaudeEnvMemberPrefix recognizes only a decoded key at the actual
// top-level extensions.claude_env path. It never repairs or accepts input.
// Stop on an undecodable prefix: unrelated malformed legacy files retain
// optional-config discovery. Values and nested lookalike keys are not signals.
func hasClaudeEnvMemberPrefix(payload []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return false
	}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return false
		}
		if key == "extensions" {
			token, err := decoder.Token()
			if err != nil {
				return false
			}
			if token != json.Delim('{') {
				// A duplicate later extensions object may still declare the
				// member. Skip this non-object without treating it as absence.
				if token == json.Delim('[') {
					for decoder.More() {
						var ignored json.RawMessage
						if decoder.Decode(&ignored) != nil {
							return false
						}
					}
					if _, err := decoder.Token(); err != nil {
						return false
					}
				}
				continue
			}
			for decoder.More() {
				member, err := decoder.Token()
				if err != nil {
					return false
				}
				if member == "claude_env" {
					return true
				}
				var ignored json.RawMessage
				if decoder.Decode(&ignored) != nil {
					return false
				}
			}
			if _, err := decoder.Token(); err != nil {
				return false
			}
		} else {
			var ignored json.RawMessage
			if decoder.Decode(&ignored) != nil {
				return false
			}
		}
	}
	return false
}

func malformedConfigError(payload []byte, path string) error {
	err := error(configJSONError{path: path})
	if hasClaudeEnvMemberPrefix(payload) {
		return markLauncherSelectionRefusal(err)
	}
	return err
}

var claudeEnvInteger = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)

func parseLauncherClaudeEnv(raw any) (*LauncherClaudeEnv, error) {
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, claudeEnvError(ClaudeEnvInvalid)
	}
	// Names precede values, regardless of object iteration order.
	for key := range object {
		if key != ClaudeEnvSetting {
			return nil, claudeEnvError(ClaudeEnvNameNotAllowed)
		}
	}
	selection := &LauncherClaudeEnv{}
	rawValue, present := object[ClaudeEnvSetting]
	if !present {
		return selection, nil
	}
	value, ok := rawValue.(string)
	if !ok || !claudeEnvInteger.MatchString(value) {
		return nil, claudeEnvError(ClaudeEnvValueNotInteger)
	}
	if value == "0" || strings.HasPrefix(value, "-") {
		return nil, claudeEnvError(ClaudeEnvValueNonpositive)
	}
	number, err := strconv.ParseUint(value, 10, 32)
	if err != nil || number > 65536 {
		return nil, claudeEnvError(ClaudeEnvValueOverBound)
	}
	selection.BashMaxOutputLength = value
	return selection, nil
}

// LauncherClaudeEnvStatus reports machine selection, never observed runtime
// effect. It contains neither an environment dump nor credential material.
type LauncherClaudeEnvStatus struct {
	Runtime    string `json:"runtime"`
	State      string `json:"state"`
	ConfigPath string `json:"config_path"`
}

// ClaudeEnvStatus distinguishes absence, configuration and runtime ineligibility.
// A valid shared selection does not refuse a non-Claude launch.
func (c *Config) ClaudeEnvStatus() []LauncherClaudeEnvStatus {
	rows := make([]LauncherClaudeEnvStatus, 0, 5)
	configured := c.Extensions != nil && c.Extensions.ClaudeEnv != nil && c.Extensions.ClaudeEnv.BashMaxOutputLength != ""
	for _, runtime := range []string{"claude_code", "codex_cli", "opencode", "pi", "muse"} {
		state := "absent"
		if configured {
			state = "runtime-ineligible"
			if runtime == "claude_code" {
				state = "configured"
			}
		}
		rows = append(rows, LauncherClaudeEnvStatus{Runtime: runtime, State: state, ConfigPath: c.Path})
	}
	return rows
}

// LauncherClaudeEnvContribution contains only this extension's contribution.
// It is never a complete child environment or authorization to start a process.
type LauncherClaudeEnvContribution struct {
	Runtime string            `json:"runtime"`
	Env     LauncherClaudeEnv `json:"env"`
}

func claudeEnvRuntime(runtime string) bool {
	switch runtime {
	case "claude_code", "codex_cli", "opencode", "pi", "muse":
		return true
	}
	return false
}

// ClaudeEnvContribution derives a fresh, closed contribution from a validated
// selected machine configuration. Other admitted runtimes receive an empty
// contribution; the shared declaration does not make their launch ineligible.
func (c *Config) ClaudeEnvContribution(runtime string) (LauncherClaudeEnvContribution, error) {
	if !claudeEnvRuntime(runtime) {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvInvalid)
	}
	result := LauncherClaudeEnvContribution{Runtime: runtime}
	if c.Extensions != nil && c.Extensions.ClaudeEnv != nil {
		// Recheck the typed value rather than letting programmatic mutation bypass
		// the same value grammar used by the selected machine-file parser.
		selected := c.Extensions.ClaudeEnv
		object := map[string]any{}
		if selected.BashMaxOutputLength != "" {
			object[ClaudeEnvSetting] = selected.BashMaxOutputLength
		}
		if _, err := parseLauncherClaudeEnv(object); err != nil {
			return LauncherClaudeEnvContribution{}, err
		}
		if runtime == "claude_code" {
			result.Env = *selected
		}
	}
	return result, nil
}

// ParseClaudeEnvContribution validates the narrow contribution transport before
// a consumer uses it. Shape/runtime precede declaration errors, then eligibility;
// no refusal returns a usable partial contribution or reflects supplied input.
func ParseClaudeEnvContribution(payload []byte) (LauncherClaudeEnvContribution, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvInvalid)
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvInvalid)
	}
	object, ok := raw.(map[string]any)
	if !ok || len(object) != 2 {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvInvalid)
	}
	runtime, ok := object["runtime"].(string)
	if !ok || !claudeEnvRuntime(runtime) {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvInvalid)
	}
	rawEnv, present := object["env"]
	if !present {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvInvalid)
	}
	env, err := parseLauncherClaudeEnv(rawEnv)
	if err != nil {
		return LauncherClaudeEnvContribution{}, err
	}
	if runtime != "claude_code" && env.BashMaxOutputLength != "" {
		return LauncherClaudeEnvContribution{}, claudeEnvError(ClaudeEnvRuntimeIneligible)
	}
	return LauncherClaudeEnvContribution{Runtime: runtime, Env: *env}, nil
}
