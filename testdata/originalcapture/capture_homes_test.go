package harnessparity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// This original-checkout capture never imports any new harness implementation.
func TestFreezeOriginalBaseline(t *testing.T) {
	out := os.Getenv("HARNESS_CAPTURE_OUT")
	if out == "" {
		t.Fatal("explicit private output required")
	}
	for _, c := range fixtureCases() {
		t.Run(c.Name, func(t *testing.T) {
			actual := runLegacyFixture(t, c)
			payload, err := json.MarshalIndent(actual, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, c.Name+".json"), append(payload, '\n'), 0600); err != nil {
				t.Fatal(err)
			}
			if c.State == "empty" || c.State == "user-settings" || c.State == "current" {
				if actual.Refusal != "" {
					t.Fatalf("required successful fixture refused: %s", actual.Refusal)
				}
			}
		})
	}
	t.Run("repeat-determinism", func(t *testing.T) {
		c := fixtureCase{Name: "claude_code-empty", ID: "claude_code", Operation: "managed", State: "empty"}
		first := runLegacyFixture(t, c)
		second := runLegacyFixture(t, c)
		if err := compareResult(first, second); err != nil {
			t.Fatal(err)
		}
	})
}
