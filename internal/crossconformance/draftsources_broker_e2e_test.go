package crossconformance

// Broker-transport integration at the compiled-binary boundary.
//
// The semantic suites drive the resolved lane in-process and through
// `project resolve` with a git-argument shim, asserting endpoint
// selection, refusal classes, and zero clone attempts. The per-package
// suites prove the SSH wrapper dispatch on the production binary, the
// provenance sink shape and its real-install records, and the
// no-material/exhaustion matrix. This test covers the one broker gap
// they leave: the private askpass basename dispatch in
// cmd/curator/main.go answers git's two pinned credential prompts and
// refuses everything else. A positive authenticated network fetch
// against a private test CA stays out of reach by design — the strict
// lane pins http.sslVerify with a scrubbed git environment.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/testcli"
)

// copyBrokerBinary copies the compiled CLI under a broker basename so
// the manager dispatch answers as the private broker copy.
func copyBrokerBinary(t *testing.T, basename string) string {
	t.Helper()
	payload, err := os.ReadFile(curatorBinary(t))
	if err != nil {
		t.Fatal(err)
	}
	name := basename
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dest := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(dest, payload, 0o700); err != nil {
		t.Fatal(err)
	}
	return dest
}

func runBrokerBinary(t *testing.T, bin string, env []string, stdin []string, args ...string) (int, string, string) {
	t.Helper()
	return testcli.Run(t, "", env, strings.Join(stdin, ""), bin, args...)
}

// legacyBrokerSecretEnv is the removed environment hand-off; the
// broker must ignore it so a secret placed there is never answered.
const legacyBrokerSecretEnv = "CURATOR_BUILD_HTTPS_ASKPASS_SECRET"

// runBrokerWithSecret delivers secret to the broker copy through an
// inherited pipe handle, the production hand-off, and asserts the
// child environment carries neither the secret nor the legacy name.
func runBrokerWithSecret(t *testing.T, bin string, env []string, secret string, args ...string) (int, string) {
	t.Helper()
	code, stdout, childEnv := testcli.RunWithHTTPSBrokerSecret(t, env, buildrepo.EnvHTTPSBrokerTransport, secret, bin, args...)
	for _, entry := range childEnv {
		if strings.Contains(entry, secret) || strings.HasPrefix(entry, legacyBrokerSecretEnv+"=") {
			t.Fatalf("child environment carries the secret hand-off: %q", strings.SplitN(entry, "=", 2)[0])
		}
	}
	return code, stdout
}

func writeAskpassState(t *testing.T, host, username string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"Host": host, "Username": username})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "askpass-state.json")
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDraftSourcesBrokerAskpassDispatch(t *testing.T) {
	askpass := copyBrokerBinary(t, buildrepo.HTTPSBrokerName)
	const host, username, secret = "git.example.test", "oauth2", "broker-secret"
	state := writeAskpassState(t, host, username)
	env := []string{buildrepo.EnvHTTPSBrokerState + "=" + state}
	usernamePrompt := "Username for 'https://" + host + "': "
	passwordPrompt := "Password for 'https://" + username + "@" + host + "': "

	t.Run("username prompt", func(t *testing.T) {
		code, stdout := runBrokerWithSecret(t, askpass, env, secret, usernamePrompt)
		if code != 0 || stdout != username+"\n" {
			t.Fatalf("code = %d, stdout = %q, want %q", code, stdout, username+"\n")
		}
	})
	t.Run("password prompt", func(t *testing.T) {
		code, stdout := runBrokerWithSecret(t, askpass, env, secret, passwordPrompt)
		if code != 0 || stdout != secret+"\n" {
			t.Fatalf("code = %d, stdout redacted, want the secret once", code)
		}
	})
	t.Run("missing secret handle", func(t *testing.T) {
		code, stdout, _ := runBrokerBinary(t, askpass, env, nil, passwordPrompt)
		if code != 1 || stdout != "" {
			t.Fatalf("code = %d, stdout = %q, want silent refusal", code, stdout)
		}
	})
	t.Run("legacy env secret is not answered", func(t *testing.T) {
		// Mutant control: a broker that still reads the removed
		// environment variable would answer here.
		legacy := append([]string{legacyBrokerSecretEnv + "=" + secret}, env...)
		code, stdout, _ := runBrokerBinary(t, askpass, legacy, nil, passwordPrompt)
		if code != 1 || strings.Contains(stdout, secret) {
			t.Fatalf("code = %d, stdout redacted, want silent refusal", code)
		}
	})
	t.Run("plain binary answers no prompt", func(t *testing.T) {
		// Mutant control: the dispatch is basename-gated, so the
		// ordinary CLI entry must never answer a credential prompt.
		code, stdout := runBrokerWithSecret(t, curatorBinary(t), env, secret, passwordPrompt)
		if code == 0 || strings.Contains(stdout, secret) {
			t.Fatalf("code = %d, stdout redacted, want no answer", code)
		}
	})
	for _, testCase := range []struct {
		name string
		args []string
		env  []string
	}{
		{"foreign host", []string{"Username for 'https://other.test': "}, env},
		{"foreign user", []string{"Password for 'https://root@" + host + "': "}, env},
		{"bare prompt", []string{"Password: "}, env},
		{"no arguments", nil, env},
		{"extra argument", []string{usernamePrompt, "extra"}, env},
		{"missing state", []string{passwordPrompt}, nil},
		{"relative state", []string{passwordPrompt}, []string{buildrepo.EnvHTTPSBrokerState + "=relative.json"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			code, stdout := runBrokerWithSecret(t, askpass, testCase.env, secret, testCase.args...)
			if code == 1 && stdout == "" {
				return
			}
			t.Fatalf("code = %d, stdout = %q, want silent refusal", code, stdout)
		})
	}
	t.Run("malformed state", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(bad, []byte(`{"Host":"x","Username":"y","secret":"leak"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		code, stdout := runBrokerWithSecret(t, askpass,
			[]string{buildrepo.EnvHTTPSBrokerState + "=" + bad}, secret, passwordPrompt)
		if code == 1 && stdout == "" {
			return
		}
		t.Fatalf("code = %d, stdout = %q, want silent refusal", code, stdout)
	})
}
