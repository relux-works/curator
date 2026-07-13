package skillspec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPortablePathConformanceVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "portable-paths.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Input string `json:"input"`
		Valid bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		_, err := validateRelativePath(testCase.Input, "path", true)
		if (err == nil) != testCase.Valid {
			t.Errorf("path %q valid=%v, error=%v", testCase.Input, testCase.Valid, err)
		}
	}
}
