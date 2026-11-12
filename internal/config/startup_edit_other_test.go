//go:build !unix

package config

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/startupverification"
)

// The unqualified editor refuses an ordinary valid source without changing it;
// the same configuration remains readable with its default local level.
func TestStartupEditUnqualifiedPlatformPreservesSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "machine.json")
	raw := []byte(`{"schema_version":7,"skills_root":"fixture","projects":{}}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if cfg, err := Parse(map[string]any{"schema_version": 7, "skills_root": "fixture", "projects": map[string]any{}}, path); err != nil || cfg.StartupVerification.Level != 3 {
		t.Fatalf("valid read control: %v", err)
	}
	err := SetStartupVerification(context.Background(), path, 2)
	var diagnostic *startupverification.Diagnostic
	if !errors.As(err, &diagnostic) || diagnostic.Code != "startup_verification_invalid" {
		t.Fatalf("unsupported edit: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("refused edit changed source")
	}
}
