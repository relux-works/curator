package harnessparity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Capture only the unchanged original launcher APIs and original CLI binary.
func TestFreezeOriginalLaunchers(t *testing.T) {
	out := os.Getenv("HARNESS_CAPTURE_OUT")
	if out == "" {
		t.Fatal("explicit private output required")
	}
	for _, c := range launcherCases() {
		t.Run(c.Name, func(t *testing.T) {
			actual := runLauncherFixture(t, c)
			payload, err := json.MarshalIndent(actual, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, c.Name+".json"), append(payload, '\n'), 0600); err != nil {
				t.Fatal(err)
			}
		})
	}
}
