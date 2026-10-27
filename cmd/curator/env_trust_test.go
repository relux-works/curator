package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/operatorpolicy"
	"github.com/relux-works/curator/internal/projectcontext"
)

// Claim: the actual CLI enrolls a public operator key and records/revokes a
// signed keeper request without prompting or changing the profile/native home.
// An unsigned package-shaped request cannot grant trust/hooks. Production read
// refuses unqualified adapters with empty stdout, unknown projects refuse first.
func TestOperatorPolicyCLI(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("protected registration supports macOS/Linux")
	}
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "manager")
	root := filepath.Join(base, "checkout")
	if err = os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init")
	runGit(t, root, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Fixture")
	runGit(t, root, "remote", "add", "origin", "https://code.example/project.git")
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.Repeat("1", 64), RequiredBy: []string{}}}}
	b, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	profileDir := filepath.Join(home, "profiles", "base")
	if err = os.MkdirAll(profileDir, 0700); err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(profileDir, "lock.json")
	if err = os.WriteFile(lockPath, b, 0600); err != nil {
		t.Fatal(err)
	}
	registration, err := (projectcontext.Store{Home: home}).Register("fixture", root, "base", "codex_cli")
	if err != nil {
		t.Fatal(err)
	}
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	source := stubConfigSource{path: filepath.Join(home, "config.json"), cfg: &config.Config{Path: filepath.Join(home, "config.json")}}
	invoke := func(want int, diagnostic string, args ...string) string {
		t.Helper()
		var out, stderr strings.Builder
		code := run(args, source, &out, &stderr)
		if code != want || diagnostic != "" && !strings.Contains(stderr.String(), diagnostic) {
			t.Fatalf("%v code=%d output=%s stderr=%s", args, code, out.String(), stderr.String())
		}
		if want != exitOK && out.Len() != 0 {
			t.Fatal("refusal emitted usable stdout", out.String())
		}
		return out.String()
	}
	invoke(exitOK, "", "env", "trust", "key", "--principal", "principal:operator", "--public-key", base64.StdEncoding.EncodeToString(pub))
	now := time.Now().UTC().Add(-time.Second).Format(time.RFC3339)
	issuer := operatorpolicy.Issuer{Role: "operator", Principal: "principal:operator", Approval: strings.Repeat("a", 64), Actor: operatorpolicy.Actor{Kind: "operator", Principal: "principal:operator"}}
	// Root commit is read through the real command without relying on Git process
	// environment inherited from the running Werkstatt checkout.
	gitState, err := os.ReadFile(filepath.Join(home, "project-context", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Registrations map[string]struct {
			Binding struct {
				Repository struct {
					RootCommit string `json:"root_commit"`
				} `json:"repository"`
			} `json:"binding"`
		} `json:"registrations"`
	}
	if err = json.Unmarshal(gitState, &saved); err != nil {
		t.Fatal(err)
	}
	p := operatorpolicy.Policy{ID: "project", RootCommit: saved.Registrations[registration.CheckoutID].Binding.Repository.RootCommit, Destination: operatorpolicy.Destination{Transport: "https", Host: "code.example", Port: 443, Path: "/project.git"}, Spellings: []string{"https://code.example/project.git"}, Environments: []string{"codex_cli"}, Lanes: []string{"lane:test"}, ProjectTrust: true, Hooks: []operatorpolicy.Hook{}, Issuer: issuer, Time: now, Reason: "CLI approval", ReviewBy: time.Now().UTC().Add(time.Hour).Format(time.RFC3339)}
	sel := operatorpolicy.Selection{Checkout: registration.CheckoutID, Path: registration.ProjectRoot, Environment: "codex_cli", Lane: "lane:test", Profile: registration.Profile}
	command := operatorpolicy.Command{Kind: "operator-policy-command/1", Operation: "grant", Issuer: issuer, Time: now, Reason: p.Reason, EntryID: p.ID, Policy: &p, Selection: &sel, ActorPublicKey: base64.StdEncoding.EncodeToString(pub)}
	commandPath := filepath.Join(base, "command.json")
	saveCommand := func(c operatorpolicy.Command, signed bool) {
		t.Helper()
		canonical, e := projectcontext.MarshalRecord(c)
		if e != nil {
			t.Fatal(e)
		}
		sig := ""
		if signed {
			sig = base64.StdEncoding.EncodeToString(ed25519.Sign(key, canonical))
		}
		payload, e := json.Marshal(operatorpolicy.SignedCommand{Command: c, ApprovalSignature: sig, ActorSignature: sig})
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(commandPath, payload, 0600); e != nil {
			t.Fatal(e)
		}
	}
	saveCommand(command, false)
	invoke(exitFail, "workspace_trust_policy_authority", "env", "trust", "grant", "--command", commandPath)
	saveCommand(command, true)
	shown := invoke(exitOK, "", "env", "trust", "grant", "--command", commandPath)
	if !strings.Contains(shown, `"event":"policy-grant"`) {
		t.Fatal(shown)
	}
	before, err := os.ReadFile(filepath.Join(home, "operator-policy", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	readArgs := []string{"env", "trust", "read", "codex_cli", registration.ProjectRoot, "--checkout", registration.CheckoutID, "--lane", "lane:test", "--profile", "base", "--lock-sha256", registration.Profile.LockSHA256}
	invoke(exitFail, "workspace_trust_adapter_unqualified", readArgs...)
	readArgs[6] = "unknown"
	invoke(exitFail, "workspace_untrusted", readArgs...)
	readArgs[6] = registration.CheckoutID
	invoke(exitOK, "", "env", "trust", "list")
	after, err := os.ReadFile(filepath.Join(home, "operator-policy", "state.json"))
	if err != nil || string(before) != string(after) {
		t.Fatal("read/list changed state", err)
	}
	command.Operation = "revoke"
	command.Policy = nil
	command.Selection = nil
	command.Issuer.Approval = strings.Repeat("b", 64)
	command.Time = time.Now().UTC().Format(time.RFC3339)
	command.Reason = "Revoke CLI policy"
	saveCommand(command, true)
	invoke(exitOK, "", "env", "trust", "revoke", "--command", commandPath)
	invoke(exitFail, "workspace_trust_grant_revoked", readArgs...)
	if _, err = (operatorpolicy.Store{Home: home}).List(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, err = os.ReadFile(lockPath)
	if err != nil || string(after) != string(b) {
		t.Fatal("policy command changed profile", err)
	}
	for _, name := range []string{".claude", ".codex", "environments", "project-context/homes/" + registration.CheckoutID + "/config.toml"} {
		if _, err = os.Lstat(filepath.Join(home, name)); !os.IsNotExist(err) {
			t.Fatal("unexpected native write", name, err)
		}
	}
}

// Claim: package TestMain pins all host homes and removes native override
// variables before test code runs; a native path under the ambient home cannot
// resolve to the operator's actual profile. This is a startup-boundary check,
// not a claim about tests that later deliberately mutate their environment.
func TestOperatorPolicyHomeIsolation(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" || os.Getenv("USERPROFILE") != home || os.Getenv("XDG_CONFIG_HOME") != home || os.Getenv("CODEX_HOME") != "" || os.Getenv("CLAUDE_CONFIG_DIR") != "" || !strings.Contains(filepath.Base(home), "curator-cli-home-") {
		t.Fatal("test package homes not isolated")
	}
}
