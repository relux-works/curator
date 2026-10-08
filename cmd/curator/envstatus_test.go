package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

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

func TestEnvStatusReportsNotInheritedNativeCodexServers(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	configPath := filepath.Join(os.Getenv("CODEX_HOME"), "config.toml")
	nativeConfig := "model = \"gpt-5-codex\"\n\n[mcp_servers.figma]\ncommand = \"server\"\nargs = []\n"
	if err := os.WriteFile(configPath, []byte(nativeConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("env resolve codex_cli --repair = %d\nstderr:\n%s", code, stderr)
	} else if !strings.Contains(stderr, "mcp_native_servers_not_inherited: native Codex MCP servers figma") || strings.Contains(strings.ToLower(stderr), "accept the loss") {
		t.Fatalf("env resolve omitted the revision-B not-inherited warning, or retained a migration hint:\n%s", stderr)
	}
	managedConfig := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.CodexCLI), "config.toml")
	seeded, err := os.ReadFile(managedConfig)
	if err != nil {
		t.Fatal(err)
	}
	assertCLISeedStripped(t, seeded)
	_, _, marker := cliEnvMarker(t, source, envregistry.CodexCLI)
	if record := marker.CodexSeedRecord; record == nil || record.Revision != envregistry.CodexSeedRevisionB || len(record.NativeMCPServers) != 1 || record.NativeMCPServers[0] != "figma" {
		t.Fatalf("CLI provisioning seed record = %+v, want B with figma", record)
	}
	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\nstderr:\n%s", code, stderr)
	}
	for _, want := range []string{
		"codex-seed: revision B (shipped)",
		"codex-seed-record: revision B; native MCP servers figma (not-inherited)",
		"warning: mcp_native_servers_not_inherited: native Codex MCP servers figma were stripped from config.toml and are not inherited",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("env status omitted %q:\n%s", want, stdout)
		}
	}
}

func TestEnvResolveStripsAndReportsInlineNativeCodexMCPTable(t *testing.T) {
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
	} else if !strings.Contains(stderr, "mcp_native_servers_not_inherited: native Codex MCP servers local") || strings.Contains(strings.ToLower(stderr), "accept the loss") {
		t.Fatalf("env resolve omitted the revision-B inline-table warning, or retained a migration hint:\n%s", stderr)
	}

	seededPath := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.CodexCLI), "config.toml")
	seeded, err := os.ReadFile(seededPath)
	if err != nil {
		t.Fatal(err)
	}
	assertCLISeedStripped(t, seeded)

	code, stdout, stderr := runProfile(t, source, "env", "status")
	if code != exitOK {
		t.Fatalf("env status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "native MCP servers local (not-inherited)") {
		t.Fatalf("env status omitted inline-table server name:\n%s", stdout)
	}
	if strings.Contains(stdout, `command = "server"`) {
		t.Fatalf("env status exposed the native MCP server command:\n%s", stdout)
	}
}

func assertCLISeedStripped(t *testing.T, payload []byte) {
	t.Helper()
	var seeded map[string]any
	if err := toml.Unmarshal(payload, &seeded); err != nil {
		t.Fatalf("CLI seed is invalid TOML: %v", err)
	}
	if _, exists := seeded["mcp_servers"]; exists {
		t.Fatalf("revision-B CLI seed inherited native MCP servers: %s", payload)
	}
	if len(seeded) != 1 || seeded["model"] != "gpt-5-codex" {
		t.Fatalf("revision-B CLI seed failed to preserve non-MCP members: %#v", seeded)
	}
}
