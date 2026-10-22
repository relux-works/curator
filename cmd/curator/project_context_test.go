package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/projectcontext"
)

// Claim: the public noun/verb grammar drives register, explicit package,
// inspect, preview, exact approval and revoke without mutating profile bytes.
// The flow never creates a launch fragment or claims native qualification.
func TestProjectAdmissionCLIFlow(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("CIP root identities support macOS/Linux")
	}
	root := t.TempDir()
	home := filepath.Join(root, "manager")
	checkout := filepath.Join(root, "checkout")
	if err := os.MkdirAll(checkout, 0700); err != nil {
		t.Fatal(err)
	}
	runGit(t, checkout, "init")
	runGit(t, checkout, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "Fixture")
	lock := &contextlock.Lock{SchemaVersion: 4, HashVersion: 2, Precedence: contextlock.Precedence{Winner: "higher-weight", Placement: "winner-last"}, Root: "base", Members: []contextlock.Member{{Kind: "context", Name: "base", Version: "1.0.0", StateHash: strings.Repeat("1", 64), RequiredBy: []string{}}}}
	b, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(home, "profiles", "base")
	if err = os.MkdirAll(p, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(p, "lock.json"), b, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(checkout, "AGENTS.md"), []byte("Explicit project context.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	source := stubConfigSource{path: filepath.Join(home, "config.json"), cfg: &config.Config{Path: filepath.Join(home, "config.json")}}
	invoke := func(args ...string) string {
		t.Helper()
		var out, stderr strings.Builder
		code := run(args, source, &out, &stderr)
		if code != 0 {
			t.Fatalf("%v = %d\n%s", args, code, stderr.String())
		}
		return out.String()
	}
	var r projectcontext.Registration
	if err = json.Unmarshal([]byte(invoke("project", "register", checkout, "--name", "fixture", "--profile", "base", "--environment", "codex_cli")), &r); err != nil {
		t.Fatal(err)
	}
	inspect := invoke("project", "inspect", r.CheckoutID)
	if !strings.Contains(inspect, `"mode": "off"`) || !strings.Contains(inspect, `"memory": "off"`) {
		t.Fatal(inspect)
	}
	var preview projectcontext.Preview
	if err = json.Unmarshal([]byte(invoke("project", "package", r.CheckoutID, "guide", "--source", "AGENTS.md")), &preview); err != nil {
		t.Fatal(err)
	}
	shown := invoke("project", "preview", r.CheckoutID)
	if !strings.Contains(shown, "agent-context.json") || !strings.Contains(shown, "Explicit project context.") {
		t.Fatal("preview omitted bytes/manifest")
	}
	invoke("project", "approve", r.CheckoutID, "guide", "--digest", preview.Review.Items[0].ItemSHA256, "--package-digest", preview.PackageSHA256)
	invoke("project", "revoke", r.CheckoutID, "guide")
	before, err := os.ReadFile(filepath.Join(p, "lock.json"))
	if err != nil || string(before) != string(b) {
		t.Fatal("profile bytes changed")
	}
	// CLI syntax refuses authority strings and surplus operands before loading
	// configuration; a repository cannot choose an approver or policy ceiling.
	for _, args := range [][]string{{"project", "approve", r.CheckoutID, "guide", "--as", "operator"}, {"project", "register", checkout, "extra"}, {"project", "inspect", r.CheckoutID, "extra"}, {"project", "package", r.CheckoutID, "guide", "--source", "../AGENTS.md"}} {
		var out, stderr strings.Builder
		if code := run(args, source, &out, &stderr); code == 0 {
			t.Fatalf("forbidden syntax succeeded: %v", args)
		}
	}
}
