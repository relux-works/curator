package envsnapshot

import "testing"

// Embedded schemas must compile offline before any structural test can pass.
func TestEmbeddedSchemasCompile(t *testing.T) {
	if _, err := validators(); err != nil {
		t.Fatal(err)
	}
}
