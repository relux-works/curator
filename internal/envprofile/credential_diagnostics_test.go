//go:build !windows

package envprofile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

// The metadata diagnostic distinguishes synthetic link shapes on non-Windows hosts,
// preserves each preimage, and never includes fixture payloads in its findings.
// This proves file-link readiness, not OAuth validity or refresh reliability.
func TestClaudeCredentialLinkDiagnostics(t *testing.T) {
	requireLinkCapability(t)
	for _, tc := range []struct {
		name    string
		want    CredentialLinkReason
		warning bool
	}{
		{"healthy", CredentialLinkOK, false},
		{"relative", CredentialLinkWrongTarget, false},
		{"symlink-dotdot", CredentialLinkWrongTarget, false},
		{"dangling", CredentialLinkDangling, true},
		{"replaced", CredentialLinkReplaced, false},
		{"wrong-target", CredentialLinkWrongTarget, false},
		{"missing", CredentialLinkMissing, false},
		{"directory", CredentialLinkWrongShape, false},
		{"directory-target", CredentialLinkWrongShape, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "native", ".credentials.json")
			full := filepath.Join(root, "managed", ".credentials.json")
			alt := filepath.Join(root, "other", ".credentials.json")
			for _, dir := range []string{filepath.Dir(target), filepath.Dir(full), filepath.Join(root, "other", "subdir")} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			const payload = "synthetic-credential-payload-must-not-appear"
			if err := os.WriteFile(target, []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(alt, []byte("synthetic-other-payload"), 0o600); err != nil {
				t.Fatal(err)
			}
			linkTarget := target
			switch tc.name {
			case "relative":
				linkTarget = filepath.Join("..", "native", ".credentials.json")
			case "symlink-dotdot":
				if err := os.Symlink(filepath.Join(root, "other", "subdir"), filepath.Join(root, "native", "alias")); err != nil {
					t.Fatal(err)
				}
				// Keep the raw traversal: filepath.Join would erase the witness.
				linkTarget = filepath.Join(root, "native", "alias") + string(filepath.Separator) + ".." + string(filepath.Separator) + ".credentials.json"
			case "wrong-target":
				linkTarget = alt
			case "dangling":
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
			case "directory-target":
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(target, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			switch tc.name {
			case "replaced":
				if err := os.WriteFile(full, []byte(payload), 0o600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(full, 0o700); err != nil {
					t.Fatal(err)
				}
			case "missing":
			default:
				if err := os.Symlink(linkTarget, full); err != nil {
					t.Fatal(err)
				}
			}
			// Absolute and relative controls reach the expected inode. The
			// Unix symlink/.. witness reaches another inode.
			if tc.name == "healthy" || tc.name == "relative" || tc.name == "symlink-dotdot" {
				actual, err := os.Stat(full)
				if err != nil {
					t.Fatal(err)
				}
				expected, err := os.Stat(target)
				if err != nil {
					t.Fatal(err)
				}
				if os.SameFile(actual, expected) != (tc.name != "symlink-dotdot") {
					t.Fatal("fixture does not demonstrate the intended target identity")
				}
			}
			unchanged := credentialPreimage(t, full, target, alt, filepath.Join(root, "native", "alias"))
			req := ResolveRequest{}
			got := req.inspectCredentialLink(full, target)
			unchanged()
			if got != tc.want {
				t.Fatalf("reason=%s, want %s", got, tc.want)
			}
			adapter, _ := envregistry.ByID(envregistry.ClaudeCode)
			plan := &homePlan{adapter: adapter, homeDir: filepath.Dir(full)}
			verdict := &verification{}
			// Supply a synthetic Linux file-link policy to exercise the real
			// verifier hook without changing backend admission.
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
			verdict.checkPassthrough(&req, plan, &envmarker.Marker{Version: envmarker.VersionV1, Passthrough: &entries})
			unchanged()
			findings := strings.Join(append(verdict.reasons, verdict.warnings...), "; ")
			if tc.want == CredentialLinkOK {
				if findings != "" {
					t.Fatalf("healthy link reported a finding: %s", findings)
				}
			} else {
				if !strings.Contains(findings, DiagCredentialLink) || !strings.Contains(findings, "reason="+string(tc.want)) {
					t.Fatalf("missing typed reason: %s", findings)
				}
				if tc.want == CredentialLinkWrongTarget && !strings.Contains(findings, envregistry.DiagCredentialConflict) {
					t.Fatalf("wrong target must retain the credential conflict refusal: %s", findings)
				}
				if tc.warning != (len(verdict.warnings) > 0) || tc.warning == (len(verdict.reasons) > 0) {
					t.Fatalf("wrong warning/stale classification: %+v", verdict)
				}
			}
			if strings.Contains(findings, payload) || strings.Contains(findings, "synthetic-other-payload") {
				t.Fatal("diagnostic disclosed synthetic credential content")
			}
			unchanged()
		})
	}
}

// Status and resolve use the same diagnostic on Linux's actual file backend.
// macOS does not declare this shared link and must not gain one here.
func TestClaudeCredentialDiagnosticsStatusResolve(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Claude shared file-link backend is exercised on Linux; metadata/hook cases run on non-Windows hosts")
	}
	requireLinkCapability(t)
	for _, shape := range []string{"healthy", "dangling", "replaced", "wrong-target", "missing"} {
		t.Run(shape, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			seedLiveNativeCredentials(t, fx)
			provision(t, fx, envregistry.ClaudeCode, envregistry.DefaultMachineConfig())
			full := filepath.Join(ManagedHomeDir(fx.home, fx.profile, envregistry.ClaudeCode), ".credentials.json")
			target := filepath.Join(fx.native[envregistry.ClaudeCode], ".credentials.json")
			alt := filepath.Join(fx.native[envregistry.ClaudeCode], "other.json")
			if err := os.WriteFile(alt, []byte("synthetic-other"), 0o600); err != nil {
				t.Fatal(err)
			}
			want := CredentialLinkOK
			switch shape {
			case "dangling":
				want = CredentialLinkDangling
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
			case "replaced", "wrong-target", "missing":
				if err := os.Remove(full); err != nil {
					t.Fatal(err)
				}
				switch shape {
				case "replaced":
					want = CredentialLinkReplaced
					if err := os.WriteFile(full, []byte("synthetic-refreshed"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "wrong-target":
					want = CredentialLinkWrongTarget
					if err := os.Symlink(alt, full); err != nil {
						t.Fatal(err)
					}
				case "missing":
					want = CredentialLinkMissing
				}
			}
			unchanged := credentialPreimage(t, full, target, alt, filepath.Join(filepath.Dir(full), envmarker.Name))
			result, err := Resolve(fx.request(envregistry.ClaudeCode))
			current := shape == "healthy" || shape == "dangling"
			if current {
				if err != nil || result == nil || len(result.Document) == 0 {
					t.Fatalf("current fixture failed resolve: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), DiagHomeStale) || result == nil || len(result.Document) != 0 {
				t.Fatalf("detached fixture must refuse launch fragment: %v", err)
			}
			if want != CredentialLinkOK {
				findings := strings.Join(append(result.StaleReasons, result.Warnings...), "; ")
				if !strings.Contains(findings, "reason="+string(want)) {
					t.Fatalf("resolve missing reason: %s", findings)
				}
			}
			unchanged()
			status, err := StatusOf(statusRequest(fx))
			if err != nil {
				t.Fatal(err)
			}
			row := findHome(status, fx.profile, envregistry.ClaudeCode)
			if row == nil || row.Current != current {
				t.Fatalf("wrong status currency: %+v", row)
			}
			if want != CredentialLinkOK {
				findings := strings.Join(append(row.Findings, row.Warnings...), "; ")
				if !strings.Contains(findings, "reason="+string(want)) {
					t.Fatalf("status missing reason: %s", findings)
				}
			}
			unchanged()
		})
	}
}

// Inspection failures are unreadable, never healthy or absent; error payloads
// are not echoed, and no error path changes the synthetic link or target.
func TestClaudeCredentialDiagnosticsUnreadable(t *testing.T) {
	requireLinkCapability(t)
	root := t.TempDir()
	target := filepath.Join(root, "target")
	full := filepath.Join(root, ".credentials.json")
	if err := os.WriteFile(target, []byte("synthetic"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, full); err != nil {
		t.Fatal(err)
	}
	unchanged := credentialPreimage(t, full, target)
	for _, readlink := range []bool{false, true} {
		req := ResolveRequest{}
		if readlink {
			req.passthroughReadlink = func(string) (string, error) { return "", os.ErrPermission }
		} else {
			req.passthroughLstat = func(string) (os.FileInfo, error) { return nil, os.ErrPermission }
		}
		adapter, _ := envregistry.ByID(envregistry.ClaudeCode)
		verdict := &verification{}
		if !verdict.checkClaudeCredentialLink(&req, &homePlan{adapter: adapter, homeDir: root}, ".credentials.json", target) {
			t.Fatal("link not handled")
		}
		if !verdict.passthroughReadFailed || len(verdict.warnings) != 0 || len(verdict.reasons) != 1 || !strings.Contains(verdict.reasons[0], "reason=unreadable") || !strings.Contains(verdict.reasons[0], envregistry.DiagPassthroughUnreadable) {
			t.Fatalf("wrong unreadable verdict: %+v", verdict)
		}
		unchanged()
	}
}
