package manifest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Previewing a valid replacement preserves unrelated declarations and metadata,
// leaves its input and live file unchanged, and yields the ordinary add bytes.
func TestWithDeclPrivatePreview(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, Name)
	original := []byte(`{"schema_version":1,"locale":"en","skills":[{"name":"kept","tag":"v1"},{"name":"changed","tag":"old"}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	input := bytes.Clone(original)
	preview, err := WithDecl(input, path, "changed", "tag", "new", "https://example.test/changed.git", "")
	if err != nil {
		t.Fatal(err)
	}
	live, _ := os.ReadFile(path)
	if !bytes.Equal(input, original) || !bytes.Equal(live, original) {
		t.Fatal("preview modified declaration input")
	}
	parsed, err := ParseBytes(preview, path)
	if err != nil || len(parsed.Skills) != 2 || parsed.Skills[0].Name != "kept" || parsed.Skills[1].Ref.Value != "new" || parsed.Locale != "en" {
		t.Fatalf("preview lost declarations or metadata: %+v %v", parsed, err)
	}
	if err := AddDecl(root, "changed", "tag", "new", "https://example.test/changed.git", ""); err != nil {
		t.Fatal(err)
	}
	live, _ = os.ReadFile(path)
	if !bytes.Equal(live, preview) {
		t.Fatal("preview and publication bytes differ")
	}
}

// Malformed carrier bytes and invalid prospective names refuse rather than
// being repaired by the edit; rejection leaves the caller's bytes untouched.
func TestWithDeclRejectsInvalidInput(t *testing.T) {
	for _, input := range [][]byte{[]byte(`null`), []byte(`{"schema_version":1,"schema_version":1,"skills":[]}`), []byte(`{`)} {
		before := bytes.Clone(input)
		if _, err := WithDecl(input, Name, "skill", "tag", "v1", "", ""); err == nil {
			t.Fatal("invalid preview admitted")
		}
		if !bytes.Equal(input, before) {
			t.Fatal("rejected preview modified input")
		}
	}
	if _, err := WithDecl([]byte(`{"schema_version":1,"skills":[]}`), Name, "../escape", "tag", "v1", "", ""); err == nil {
		t.Fatal("invalid prospective name admitted")
	}
}
