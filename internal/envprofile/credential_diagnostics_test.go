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

// The metadata diagnostic distinguishes synthetic link shapes on every host,
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
		{"relative", CredentialLinkOK, false},
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
			for _, dir := range []string{filepath.Dir(target), filepath.Dir(full)} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			const payload = "synthetic-credential-payload-must-not-appear"
			if err := os.WriteFile(target, []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			alt := filepath.Join(root, "other")
			if err := os.WriteFile(alt, []byte("synthetic-other-payload"), 0o600); err != nil {
				t.Fatal(err)
			}
			linkTarget := target
			switch tc.name {
			case "relative":
				linkTarget = filepath.Join("..", "native", ".credentials.json")
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
			unchanged := credentialPreimage(t, full, target, alt)
			req := ResolveRequest{}
			if got := req.inspectCredentialLink(full, target); got != tc.want {
				t.Fatalf("reason=%s, want %s", got, tc.want)
			}
			adapter, _ := envregistry.ByID(envregistry.ClaudeCode)
			plan := &homePlan{adapter: adapter, homeDir: filepath.Dir(full)}
			verdict := &verification{}
			if !verdict.checkClaudeCredentialLink(&req, plan, ".credentials.json", target) {
				t.Fatal("Claude credential link was not handled")
			}
			findings := strings.Join(append(verdict.reasons, verdict.warnings...), "; ")
			if tc.want == CredentialLinkOK {
				if findings != "" {
					t.Fatalf("healthy link reported a finding: %s", findings)
				}
			} else {
				if !strings.Contains(findings, DiagCredentialLink) || !strings.Contains(findings, "reason="+string(tc.want)) {
					t.Fatalf("missing typed reason: %s", findings)
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
// macOS/Windows do not declare this shared link and must not gain one here.
func TestClaudeCredentialDiagnosticsStatusResolve(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Claude shared file-link backend is exercised on Linux; metadata/hook cases run on all hosts")
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

// credentialPreimage reads only synthetic fixture bytes to detect mutation;
// production inspection has no content-reader dependency.
func credentialPreimage(t *testing.T, paths ...string) func() {
	t.Helper()
	type image struct {
		info   os.FileInfo
		data   string
		absent bool
	}
	capture := func(path string) image {
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			return image{absent: true}
		}
		if err != nil {
			t.Fatal(err)
		}
		out := image{info: info}
		if info.Mode()&os.ModeSymlink != 0 {
			out.data, err = os.Readlink(path)
		} else if info.Mode().IsRegular() {
			var b []byte
			b, err = os.ReadFile(path)
			out.data = string(b)
		}
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	before := map[string]image{}
	for _, path := range paths {
		before[path] = capture(path)
	}
	return func() {
		t.Helper()
		for _, path := range paths {
			a, b := before[path], capture(path)
			if a.absent != b.absent {
				t.Fatal("diagnostic changed fixture path existence")
			}
			if a.absent {
				continue
			}
			if !os.SameFile(a.info, b.info) || a.info.Mode() != b.info.Mode() || a.info.Size() != b.info.Size() || !a.info.ModTime().Equal(b.info.ModTime()) || a.data != b.data {
				t.Fatal("diagnostic changed fixture preimage")
			}
		}
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

// Other providers and non-credential paths retain their existing verifier.
func TestClaudeCredentialDiagnosticsScope(t *testing.T) {
	for _, tc := range []struct{ id, path string }{{envregistry.CodexCLI, "auth.json"}, {envregistry.ClaudeCode, "other.json"}} {
		adapter, _ := envregistry.ByID(tc.id)
		req := ResolveRequest{passthroughLstat: func(string) (os.FileInfo, error) {
			t.Fatal("out-of-scope path inspected")
			return nil, os.ErrPermission
		}}
		verdict := &verification{}
		if verdict.checkClaudeCredentialLink(&req, &homePlan{adapter: adapter}, tc.path, "unused") {
			t.Fatal("out-of-scope path handled")
		}
		if len(verdict.reasons) != 0 || len(verdict.warnings) != 0 {
			t.Fatal("out-of-scope finding")
		}
	}
}
