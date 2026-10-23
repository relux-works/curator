package scopes

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Machine hybrid reads and edits reject project-only schemas before mutation;
// schema 1 remains readable and supports both add and remove.
func TestHybridProjectSchemasRefused(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		for _, operation := range []string{"load", "add", "remove"} {
			t.Run(fmt.Sprintf("%d/%s", version, operation), func(t *testing.T) {
				home := t.TempDir()
				path := HybridManifestPath(home)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				payload := []byte(fmt.Sprintf(`{"schema_version":%d,"skills":[{"name":"review","tag":"v1.0.0","targets":["qa"]}]}`, version))
				if err := os.WriteFile(path, payload, 0600); err != nil {
					t.Fatal(err)
				}
				var err error
				switch operation {
				case "load":
					_, err = LoadHybridDecls(home)
				case "add":
					err = AddHybridDecl(home, "other", "tag", "v1.0.0", "", []string{"qa"})
				case "remove":
					err = RemoveHybridDecl(home, "review")
				}
				if version == 1 {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if err == nil || !strings.Contains(err.Error(), "requires a newer tool") {
					t.Fatalf("project schema admitted: %v", err)
				}
				after, readErr := os.ReadFile(path)
				if readErr != nil || !bytes.Equal(payload, after) {
					t.Fatalf("refusal changed manifest: %v", readErr)
				}
			})
		}
	}
}
