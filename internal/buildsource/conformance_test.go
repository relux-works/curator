package buildsource

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildSourceConformanceVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Path          string `json:"path"`
		ContentBase64 string `json:"content_base64"`
	}
	type vectorCase struct {
		Name          string   `json:"name"`
		Result        string   `json:"result"`
		ContentSHA256 string   `json:"content_sha256"`
		Records       []record `json:"records"`
		InputOrder    []record `json:"input_order"`
	}
	var vectors struct {
		BuildSourceCases []vectorCase `json:"build_source_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range vectors.BuildSourceCases {
		records := testCase.Records
		if len(records) == 0 {
			records = testCase.InputOrder
		}
		if testCase.Result != "accepted" || len(records) == 0 {
			continue
		}
		t.Run(testCase.Name, func(t *testing.T) {
			tree := t.TempDir()
			for _, item := range records {
				content, err := base64.StdEncoding.DecodeString(item.ContentBase64)
				if err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, tree, item.Path, content)
			}
			token := validateTestTree(t, tree)
			if got := token.Identity().ContentSHA256; got != testCase.ContentSHA256 {
				t.Fatalf("digest = %s, want %s", got, testCase.ContentSHA256)
			}
		})
	}
}
