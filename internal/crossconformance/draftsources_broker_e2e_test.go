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
	env := []string{
		buildrepo.EnvHTTPSBrokerState + "=" + state,
		buildrepo.EnvHTTPSBrokerSecret + "=" + secret,
	}
	usernamePrompt := "Username for 'https://" + host + "': "
	passwordPrompt := "Password for 'https://" + username + "@" + host + "': "

	t.Run("username prompt", func(t *testing.T) {
		code, stdout, _ := runBrokerBinary(t, askpass, env, nil, usernamePrompt)
		if code != 0 || stdout != username+"\n" {
			t.Fatalf("code = %d, stdout = %q, want %q", code, stdout, username+"\n")
		}
	})
	t.Run("password prompt", func(t *testing.T) {
		code, stdout, _ := runBrokerBinary(t, askpass, env, nil, passwordPrompt)
		if code != 0 || stdout != secret+"\n" {
			t.Fatalf("code = %d, stdout redacted, want the secret once", code)
		}
	})
	t.Run("plain binary answers no prompt", func(t *testing.T) {
		// Mutant control: the dispatch is basename-gated, so the
		// ordinary CLI entry must never answer a credential prompt.
		code, stdout, _ := runBrokerBinary(t, curatorBinary(t), env, nil, passwordPrompt)
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
		{"missing secret", []string{passwordPrompt}, []string{buildrepo.EnvHTTPSBrokerState + "=" + state}},
		{"missing state", []string{passwordPrompt}, []string{buildrepo.EnvHTTPSBrokerSecret + "=" + secret}},
		{"relative state", []string{passwordPrompt}, []string{buildrepo.EnvHTTPSBrokerState + "=relative.json", buildrepo.EnvHTTPSBrokerSecret + "=" + secret}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			code, stdout, _ := runBrokerBinary(t, askpass, testCase.env, nil, testCase.args...)
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
		code, stdout, _ := runBrokerBinary(t, askpass,
			[]string{buildrepo.EnvHTTPSBrokerState + "=" + bad, buildrepo.EnvHTTPSBrokerSecret + "=" + secret},
			nil, passwordPrompt)
		if code == 1 && stdout == "" {
			return
		}
		t.Fatalf("code = %d, stdout = %q, want silent refusal", code, stdout)
	})
}
