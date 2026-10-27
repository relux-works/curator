//go:build windows

package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// Real synthetic Windows link/file shapes never certify exact-target readiness.
// In particular, host normalization of the symlink/.. witness cannot turn it
// into an ok result. The verifier reports unsupported and preserves preimages;
// this does not claim any Windows credential backend qualification.
func TestClaudeCredentialDiagnosticsWindowsUnsupported(t *testing.T) {
	for _, shape := range []string{"missing", "replaced", "healthy", "symlink-dotdot"} {
		t.Run(shape, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "native", ".credentials.json")
			full := filepath.Join(root, "managed", ".credentials.json")
			alias := filepath.Join(root, "native", "alias")
			for _, dir := range []string{filepath.Dir(target), filepath.Dir(full), filepath.Join(root, "other", "subdir")} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			const payload = "synthetic-windows-credential-not-for-diagnostics"
			if err := os.WriteFile(target, []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			switch shape {
			case "replaced":
				if err := os.WriteFile(full, []byte(payload), 0o600); err != nil {
					t.Fatal(err)
				}
			case "healthy", "symlink-dotdot":
				requireLinkCapability(t)
				linkTarget := target
				if shape == "symlink-dotdot" {
					if err := os.Symlink(filepath.Join(root, "other", "subdir"), alias); err != nil {
						t.Fatal(err)
					}
					linkTarget = alias + string(filepath.Separator) + ".." + string(filepath.Separator) + ".credentials.json"
				}
				if err := os.Symlink(linkTarget, full); err != nil {
					t.Fatal(err)
				}
			}
			unchanged := credentialPreimage(t, full, target, alias)
			req := ResolveRequest{}
			if got := req.inspectCredentialLink(full, target); got != CredentialLinkUnsupported {
				t.Fatalf("reason=%s, want unsupported (never healthy)", got)
			}
			adapter, _ := envregistry.ByID(envregistry.ClaudeCode)
			release := "unknown"
			req.credentialRelease = &release
			req.credentialPolicy = &credentialPolicyResolution{
				environment: adapter.ID, release: release,
				policy: &effectiveCredentialPolicy{
					links:      map[string]string{".credentials.json": target},
					strategies: map[string]string{".credentials.json": envregistry.StrategyFileLink},
				},
			}
			entries := []envmarker.Passthrough{{Path: ".credentials.json", Strategy: envregistry.StrategyFileLink}}
			verdict := &verification{}
			verdict.checkPassthrough(&req, &homePlan{adapter: adapter, homeDir: filepath.Dir(full)},
				&envmarker.Marker{Version: envmarker.VersionV1, Passthrough: &entries})
			if len(verdict.reasons) != 1 || len(verdict.warnings) != 0 || verdict.passthroughReadFailed {
				t.Fatalf("unsupported must refuse without a warning or unreadable claim: %+v", verdict)
			}
			finding := verdict.reasons[0]
			if !strings.Contains(finding, envregistry.DiagCredentialUnsupported) ||
				!strings.Contains(finding, DiagCredentialLink) || !strings.Contains(finding, "reason=unsupported") ||
				strings.Contains(finding, payload) {
				t.Fatalf("wrong unsupported diagnostic: %s", finding)
			}
			unchanged()
		})
	}
}

// Unsupported inspection performs neither metadata read nor readlink, and the
// unsupported branch must not introduce a Windows shared-file backend.
// The common scope test remains the positive control for unaffected providers.
func TestClaudeCredentialDiagnosticsWindowsNoInspection(t *testing.T) {
	req := ResolveRequest{
		passthroughLstat: func(string) (os.FileInfo, error) {
			t.Fatal("unsupported diagnostic inspected filesystem metadata")
			return nil, os.ErrPermission
		},
		passthroughReadlink: func(string) (string, error) {
			t.Fatal("unsupported diagnostic read a link")
			return "", os.ErrPermission
		},
	}
	if got := req.inspectCredentialLink("unused", "unused"); got != CredentialLinkUnsupported {
		t.Fatalf("reason=%s, want unsupported", got)
	}
	adapter, _ := envregistry.ByID(envregistry.ClaudeCode)
	if len(adapter.PassthroughFor("windows")) != 0 {
		t.Fatal("diagnostic must not introduce a Windows credential backend")
	}
}
