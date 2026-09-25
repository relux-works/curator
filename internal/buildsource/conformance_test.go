package buildsource

import (
	"encoding/base64"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

// TestBuildSourceConformanceVectors drives the record digest and identity
// assertions for every published build-source case through one counted
// consumer.
func TestBuildSourceConformanceVectors(t *testing.T) {
	cases := loadBuildSourceCases(t)
	if len(cases) == 0 {
		t.Skip("this conformance root publishes no build-source cases")
	}
	conformancecoverage.Run(t, "build-source/build-source-cases", cases,
		func(tc buildSourceCase) string { return tc.Name }, func(t *testing.T, testCase buildSourceCase) {
			records := testCase.Records
			if len(records) == 0 {
				records = testCase.InputOrder
			}
			if testCase.Result == "accepted" && len(records) > 0 {
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
			}
			assertBuildSourceIdentityVector(t, testCase)
		})
}
