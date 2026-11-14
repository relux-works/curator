package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A home with the file exclusion but no native rules exclusion is stale in
// both forms, including metacharacter homes. Read-only refusal writes nothing;
// real repair retains managed content, unrelated settings and the old exclusion.
func TestClaudeNativeRulesMissingIsStale(t *testing.T) {
	for _, tc := range claudeMemoryHomeCases(t) {
		t.Run(tc.form+"/"+tc.name, func(t *testing.T) {
			t.Setenv("HOME", tc.first)
			t.Setenv("USERPROFILE", tc.first)
			_, req, seed := claudeRepairFixture(t, tc.form)
			settings := filepath.Join(filepath.Dir(seed), "settings.json")
			context := filepath.Join(filepath.Dir(seed), "CLAUDE.md")
			beforeContext := readClaudeState(t, context)
			req.Repair = false
			if _, err := Resolve(req); err != nil {
				t.Fatalf("complete settings positive control: %v", err)
			}
			excludes, err := json.Marshal([]string{tc.firstPattern})
			if err != nil {
				t.Fatal(err)
			}
			input := append([]byte(`{"n":7.50,"claudeMdExcludes":`), excludes...)
			input = append(input, '}')
			if err := os.WriteFile(settings, input, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), "native memory exclusion") {
				t.Fatalf("missing native rules did not make home stale: %v", err)
			}
			if !bytes.Equal(input, readClaudeState(t, settings)) {
				t.Fatal("read-only rules refusal changed settings")
			}
			req.Repair = true
			if _, err := Resolve(req); err != nil {
				t.Fatal(err)
			}
			var got struct {
				Excludes []string        `json:"claudeMdExcludes"`
				N        json.RawMessage `json:"n"`
			}
			if err := json.Unmarshal(readClaudeState(t, settings), &got); err != nil {
				t.Fatal(err)
			}
			rules := strings.TrimSuffix(tc.firstPattern, "CLAUDE.md") + "rules/**"
			if !slices.Contains(got.Excludes, tc.firstPattern) || !slices.Contains(got.Excludes, rules) || string(got.N) != "7.50" {
				t.Fatalf("rules repair lost settings: %+v", got)
			}
			if !bytes.Equal(beforeContext, readClaudeState(t, context)) {
				t.Fatal("rules repair changed managed context")
			}
			req.Repair = false
			if _, err := Resolve(req); err != nil {
				t.Fatalf("repaired rules positive control: %v", err)
			}
		})
	}
}
