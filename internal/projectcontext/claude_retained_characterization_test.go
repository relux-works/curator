package projectcontext

import (
	"bytes"
	"context"
	"testing"

	"github.com/relux-works/curator/internal/projectadapter"
)

func TestClaudeRetainedProjectContextSourceProvenanceAndProtectedControls(t *testing.T) {
	_, req, preview, _ := claudeFixture(t, false, true)
	documents := map[string][]byte{
		"CLAUDE.md":                       []byte("Profile-only root.\n"),
		".agent-context/system-prompt.md": []byte("Profile-only system prompt.\n"),
	}
	source := profileFixture(t, preview, documents)
	prepared, err := PrepareClaudeMode(context.Background(), projectadapter.NativeProject, preview, *req.Claude, source)
	if err != nil {
		t.Fatal(err)
	}
	if !source.called || prepared.SourceContextRef != preview.Ref || prepared.SourceContextSHA256 != preview.SHA256 {
		t.Fatalf("profile source binding = called:%t ref:%q sha:%q; want ref:%q sha:%q", source.called, prepared.SourceContextRef, prepared.SourceContextSHA256, preview.Ref, preview.SHA256)
	}
	for path, want := range documents {
		if !bytes.Equal(prepared.Files[path], want) {
			t.Fatalf("profile document %s = %q, want %q", path, prepared.Files[path], want)
		}
	}
	for _, path := range []string{"settings.json", claudeOwnershipPath} {
		if !bytes.Equal(prepared.Files[path], preview.Files[path]) {
			t.Fatalf("protected manager control %s changed", path)
		}
	}
	for path, body := range prepared.Files {
		if bytes.Contains(body, []byte("Project document.")) || bytes.Contains(body, []byte("Unapproved")) {
			t.Fatalf("project-controlled prose survived in %s: %q", path, body)
		}
	}
	if err := validateClaudeFileSet(prepared.Files, prepared.Receipt.Files); err != nil {
		t.Fatalf("candidate receipt does not bind output: %v", err)
	}

	t.Run("bad source receipt has no candidate", func(t *testing.T) {
		bad := profileFixture(t, preview, map[string][]byte{"CLAUDE.md": []byte("Profile-only root.\n")})
		bad.receipt[0].SHA256 = digest([]byte("different source bytes"))
		candidate, err := PrepareClaudeMode(context.Background(), projectadapter.NativeProject, preview, *req.Claude, bad)
		required(t, err, "project_context_invalid")
		if candidate.Files != nil {
			t.Fatal("invalid profile-source provenance produced a candidate")
		}
	})

	t.Run("changed protected manager controls have no candidate", func(t *testing.T) {
		changed := *req.Claude
		changed.AllowTools = []string{"Read"}
		validSource := profileFixture(t, preview, map[string][]byte{"CLAUDE.md": []byte("Profile-only root.\n")})
		candidate, err := PrepareClaudeMode(context.Background(), projectadapter.NativeProject, preview, changed, validSource)
		required(t, err, "project_context_invalid")
		if candidate.Files != nil {
			t.Fatal("changed protected manager control produced a candidate")
		}
	})
}
