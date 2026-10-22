package projectcontext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Claim: the actual reader agrees with all 57 published composition grammar
// fixtures at the pinned spec commit, including their complete-record identity.
// These are syntax vectors, not native adapter or runtime qualification.
func TestCompositionPublishedVectors(t *testing.T) {
	root := filepath.Join("testdata", "composition")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Cases []struct {
			Name         string
			File         string
			BytesSHA256  string `json:"bytes_sha256"`
			RecordSHA256 string `json:"record_sha256"`
			Expected     *struct{ Code string }
		}
	}
	if err = json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Cases) != 57 {
		t.Fatal("pinned composition vector count differs")
	}
	for _, tc := range manifest.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(root, tc.File))
			if err != nil {
				t.Fatal(err)
			}
			if digest(b) != tc.BytesSHA256 {
				t.Fatal("published fixture bytes differ")
			}
			record, err := ParseCompositionRecord(b)
			if tc.Expected != nil {
				required(t, err, tc.Expected.Code)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if identity(record) != tc.RecordSHA256 {
				t.Fatal("published record identity differs")
			}
		})
	}
}
