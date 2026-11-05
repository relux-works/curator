package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envprofile"
)

// The CLI requires every origin input explicitly; absent, empty and nonempty
// selectors cannot silently substitute for each other or infer missing facts.
func TestEnvAdoptRequiresExplicitInputs(t *testing.T) {
	source, _ := profileHome(t)
	base := []string{"env", "adopt", "codex_cli", "--profile", "acme", "--operator-home", t.TempDir(), "--native-root", t.TempDir()}
	for _, selector := range [][]string{nil, {"--selector-state", "absent", "--selector-value", ""}, {"--selector-state", "present-empty"}, {"--selector-state", "present-empty", "--selector-value", "nonempty"}, {"--selector-state", "present-value", "--selector-value", ""}, {"--selector-state", "unknown"}} {
		args := append(append([]string{}, base...), selector...)
		code, stdout, stderr := runProfile(t, source, args...)
		if code != exitUsage || stdout != "" || !strings.Contains(stderr, "curator: env adopt") {
			t.Fatalf("invalid explicit input: code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
	}
	for _, flag := range []string{"--profile", "--operator-home", "--native-root"} {
		args := []string{}
		for i := 0; i < len(base); i++ {
			if base[i] == flag {
				i++
				continue
			}
			args = append(args, base[i])
		}
		args = append(args, "--selector-state", "absent")
		if code, stdout, _ := runProfile(t, source, args...); code != exitUsage || stdout != "" {
			t.Fatalf("missing %s: code=%d stdout=%q", flag, code, stdout)
		}
	}
}

// The production CLI adopts an existing legacy fixture with supplied facts,
// keeps identical repeat adoption read-only and refuses replacement authority.
func TestEnvAdoptExplicitLegacyHome(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatal(stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatal(stderr)
	}
	path := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", "codex_cli"), envorigin.Name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	original, err := envorigin.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	args := []string{"env", "adopt", "codex_cli", "--profile", "acme", "--operator-home", original.OperatorHome, "--native-root", original.NativeRoot, "--selector-state", original.Selector.State}
	if original.Selector.Value != nil {
		args = append(args, "--selector-value", *original.Selector.Value)
	}
	if code, stdout, stderr := runProfile(t, source, args...); code != exitOK || stdout != "provisioning origin adopted\n" {
		t.Fatalf("adopt code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := runProfile(t, source, args...); code != exitOK || stdout != "provisioning origin already matches\n" {
		t.Fatalf("repeat code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	afterInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(info, afterInfo) || !info.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("repeat rewrote authority")
	}
	for i := range args {
		if args[i] == "--native-root" {
			args[i+1] = t.TempDir()
			break
		}
	}
	if code, stdout, stderr := runProfile(t, source, args...); code != exitFail || stdout != "" || !strings.Contains(stderr, envorigin.DiagInvalid) {
		t.Fatalf("replace code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("replacement changed authority")
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--profile", "acme"); code != exitOK {
		t.Fatal(stderr)
	}
}
