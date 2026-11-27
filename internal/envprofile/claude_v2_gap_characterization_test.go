package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// These tests pin the Claude behaviors that harnessapi v1 cannot express.
// They remain in the shared owner until the contract gains the missing fields.
func TestClaudeHarnessV2GapCharacterization(t *testing.T) {
	adapter, err := envregistry.ByID(envregistry.ClaudeCode)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("launch argument channels", func(t *testing.T) {
		wantPrompt := []envregistry.Channel{
			{Kind: envregistry.KindFlag, Semantics: envregistry.SemanticsAppend, Flag: "--append-system-prompt-file", Argument: envregistry.ArgumentPath},
			{Kind: envregistry.KindFlag, Semantics: envregistry.SemanticsReplace, Flag: "--system-prompt-file", Argument: envregistry.ArgumentPath},
		}
		if !reflect.DeepEqual(adapter.SystemPrompt, wantPrompt) {
			t.Fatalf("system-prompt channels = %#v, want %#v", adapter.SystemPrompt, wantPrompt)
		}
		wantMCP := &envregistry.Channel{Kind: envregistry.KindFlag, Flag: "--mcp-config", Argument: envregistry.ArgumentPath, With: []string{"--strict-mcp-config"}}
		if !reflect.DeepEqual(adapter.MCP, wantMCP) {
			t.Fatalf("MCP channel = %#v, want %#v", adapter.MCP, wantMCP)
		}
	})

	t.Run("per-home Keychain route", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("Claude's per-home Keychain strategy is supported on macOS")
		}
		machine := envregistry.DefaultMachineConfig()
		machine.Isolation["fixture"] = map[string]string{envregistry.ClaudeCode: envregistry.IsolationIsolated}
		request := ResolveRequest{Profile: "fixture", Machine: machine}
		policy, err := request.computeCredentialPolicy(adapter, adapter.VerifiedRelease)
		if err != nil {
			t.Fatal(err)
		}
		if policy.isolation != envregistry.IsolationIsolated || len(policy.records) != 1 {
			t.Fatalf("policy isolation/records = %q/%#v", policy.isolation, policy.records)
		}
		record := policy.records[0]
		if record.Strategy != envregistry.StrategyPerHomeKeychain || record.SourceRole != "managed" || record.Backend != "keychain" {
			t.Fatalf("per-home route = %#v", record)
		}

		machine.Isolation["fixture"][envregistry.ClaudeCode] = envregistry.IsolationShared
		request.Machine = machine
		if _, err := request.computeCredentialPolicy(adapter, adapter.VerifiedRelease); err == nil {
			t.Fatal("shared macOS route was accepted above the verified Claude release")
		}
	})

	t.Run("dynamic seed merge", func(t *testing.T) {
		root := t.TempDir()
		homeRel := "managed"
		launchDir := filepath.Join(string(filepath.Separator), "fixture-project")
		path := filepath.Join(root, homeRel, ".claude.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		input, err := json.Marshal(map[string]any{
			"hasCompletedOnboarding": true,
			"foreign":                json.Number("17"),
			"projects": map[string]any{
				"/other": map[string]any{"unknown": "keep", "cost": json.Number("1e-7")},
				launchDir: map[string]any{
					"hasTrustDialogAccepted":              false,
					"hasClaudeMdExternalIncludesApproved": false,
					"unknown":                             []any{json.Number("1e-7")},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, input, 0o600); err != nil {
			t.Fatal(err)
		}

		prepared, err := prepareClaudeSeed(root, homeRel, launchDir, envregistry.FormReferenced)
		if err != nil {
			t.Fatal(err)
		}
		if !prepared.present || len(prepared.payload) == 0 {
			t.Fatalf("prepared seed = %#v, want a changed existing seed", prepared)
		}
		if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, input) {
			t.Fatalf("preparation changed source bytes: error=%v bytes=%q", err, got)
		}
		var merged map[string]any
		decoder := json.NewDecoder(bytes.NewReader(prepared.payload))
		decoder.UseNumber()
		if err := decoder.Decode(&merged); err != nil {
			t.Fatal(err)
		}
		if merged["foreign"] != json.Number("17") {
			t.Fatalf("foreign top-level property = %#v", merged["foreign"])
		}
		projects, ok := merged["projects"].(map[string]any)
		if !ok || projects["/other"] == nil {
			t.Fatalf("existing projects were not preserved: %#v", merged["projects"])
		}
		other := projects["/other"].(map[string]any)
		if other["unknown"] != "keep" || other["cost"] != json.Number("1e-7") {
			t.Fatalf("unowned project data changed: %#v", other)
		}
		current := projects[launchDir].(map[string]any)
		if current["hasTrustDialogAccepted"] != true || current["hasClaudeMdExternalIncludesApproved"] != true || current["unknown"] == nil {
			t.Fatalf("launch project merge = %#v", current)
		}

		if err := os.WriteFile(path, []byte(`{"projects":[]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		normalized, err := prepareClaudeSeed(root, homeRel, launchDir, envregistry.FormReferenced)
		if err != nil {
			t.Fatalf("non-map projects field was not normalized: %v", err)
		}
		var normalizedObject map[string]any
		if err := json.Unmarshal(normalized.payload, &normalizedObject); err != nil {
			t.Fatal(err)
		}
		if _, ok := normalizedObject["projects"].(map[string]any); !ok {
			t.Fatalf("normalized projects field = %#v", normalizedObject["projects"])
		}

		if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := prepareClaudeSeed(root, homeRel, launchDir, envregistry.FormReferenced); err == nil {
			t.Fatal("non-object seed document was accepted")
		}
	})
}
