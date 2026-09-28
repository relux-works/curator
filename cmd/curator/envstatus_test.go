package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

func TestPrintEnvStatusShowsCodexSeedPostureAndNames(t *testing.T) {
	var output bytes.Buffer
	printEnvStatus(&output, &envprofile.Status{
		CodexSeedRule: envprofile.CodexSeedRuleState{Revision: envregistry.CodexSeedRevisionA, Provenance: "shipped"},
		Homes: []envprofile.HomeState{{
			Profile:                     "acme",
			Environment:                 envregistry.CodexCLI,
			Provisioned:                 true,
			Current:                     true,
			CodexSeedRecord:             &envmarker.CodexSeedRecord{Revision: envregistry.CodexSeedRevisionA, NativeMCPServers: []string{"figma"}},
			NativeMCPServers:            []string{"figma"},
			NativeMCPServersDisposition: "ungoverned",
			Warnings:                    []string{envregistry.DiagMCPNativeServersUngoverned + ": figma"},
		}},
	})
	text := output.String()
	for _, want := range []string{
		"codex-seed: revision A (shipped)",
		"codex-seed-record: revision A; native MCP servers figma (ungoverned)",
		"warning: mcp_native_servers_ungoverned: figma",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("env status output omitted %q:\n%s", want, text)
		}
	}
}

func TestEnvStatusReportsUngovernedNativeCodexServers(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	configPath := filepath.Join(os.Getenv("CODEX_HOME"), "config.toml")
	nativeConfig := "model = \"gpt-5-codex\"\n\n[mcp_servers.figma]\ncommand = \"server\"\nargs = []\n"
	if err := os.WriteFile(configPath, []byte(nativeConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("env resolve codex_cli --repair = %d\nstderr:\n%s", code, stderr)
	} else if !strings.Contains(stderr, "mcp_native_servers_ungoverned: native Codex MCP servers figma") || !strings.Contains(strings.ToLower(stderr), "declare each server in the profile's mcp set, or accept the loss") {
		t.Fatalf("env resolve omitted the revision-A warning or migration hint:\n%s", stderr)
	}
	managedConfig := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.CodexCLI), "config.toml")
	seeded, err := os.ReadFile(managedConfig)
	if err != nil {
		t.Fatal(err)
	}
	if string(seeded) != nativeConfig {
		t.Fatalf("revision-A CLI seed changed config.toml:\n got %q\nwant %q", seeded, nativeConfig)
	}
	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\nstderr:\n%s", code, stderr)
	}
	for _, want := range []string{
		"codex-seed: revision A (shipped)",
		"codex-seed-record: revision A; native MCP servers figma (ungoverned)",
		"warning: mcp_native_servers_ungoverned: native Codex MCP servers figma remain in this managed home outside the profile lock and MCP allowlist",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("env status omitted %q:\n%s", want, stdout)
		}
	}
}

func TestEnvResolveCopiesAndReportsInlineNativeCodexMCPTable(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	configPath := filepath.Join(os.Getenv("CODEX_HOME"), "config.toml")
	nativeConfig := `model = "gpt-5-codex"
mcp_servers = { local = { command = "server", args = [] } }
`
	if err := os.WriteFile(configPath, []byte(nativeConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("env resolve codex_cli --repair = %d\nstderr:\n%s", code, stderr)
	} else if !strings.Contains(stderr, "mcp_native_servers_ungoverned: native Codex MCP servers local") || !strings.Contains(strings.ToLower(stderr), "accept the loss") {
		t.Fatalf("env resolve omitted the inline-table warning or migration hint:\n%s", stderr)
	}

	seededPath := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.CodexCLI), "config.toml")
	seeded, err := os.ReadFile(seededPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(seeded) != nativeConfig {
		t.Fatalf("revision-A seed changed inline native Codex config:\n got %q\nwant %q", seeded, nativeConfig)
	}

	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "native MCP servers local (ungoverned)") {
		t.Fatalf("env status omitted inline-table server name:\n%s", stdout)
	}
	if strings.Contains(stdout, `command = "server"`) {
		t.Fatalf("env status exposed the native MCP server command:\n%s", stdout)
	}
}
