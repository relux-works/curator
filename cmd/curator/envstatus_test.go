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
		CodexSeedRule: envprofile.CodexSeedRuleState{Revision: envregistry.CodexSeedRevisionB, Provenance: "shipped"},
		Homes: []envprofile.HomeState{{
			Profile:                     "acme",
			Environment:                 envregistry.CodexCLI,
			Provisioned:                 true,
			Current:                     true,
			CodexSeedRecord:             &envmarker.CodexSeedRecord{Revision: envregistry.CodexSeedRevisionB, NativeMCPServers: []string{"figma"}},
			NativeMCPServers:            []string{"figma"},
			NativeMCPServersDisposition: "not-inherited",
			Warnings:                    []string{envregistry.DiagMCPNativeServersNotInherited + ": figma"},
		}},
	})
	text := output.String()
	for _, want := range []string{
		"codex-seed: revision B (shipped)",
		"codex-seed-record: revision B; native MCP servers figma (not-inherited)",
		"warning: mcp_native_servers_not_inherited: figma",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("env status output omitted %q:\n%s", want, text)
		}
	}
}

func TestEnvStatusReportsStrippedNativeCodexServers(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	configPath := filepath.Join(os.Getenv("CODEX_HOME"), "config.toml")
	nativeConfig := "model = \"gpt-5-codex\"\n\n[mcp_servers.figma]\ncommand = \"server\"\nargs = []\n"
	if err := os.WriteFile(configPath, []byte(nativeConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("env resolve codex_cli --repair = %d\nstderr:\n%s", code, stderr)
	} else if !strings.Contains(stderr, "mcp_native_servers_not_inherited: native Codex MCP servers figma") {
		t.Fatalf("env resolve omitted the stripped-server warning:\n%s", stderr)
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
	} else if !strings.Contains(stderr, "mcp_native_servers_not_inherited: native Codex MCP servers local") {
		t.Fatalf("env resolve omitted the inline-table warning:\n%s", stderr)
	}

	seededPath := filepath.Join(envprofile.ManagedHomeDir(source.cfg.Home(), "acme", envregistry.CodexCLI), "config.toml")
	seeded, err := os.ReadFile(seededPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(seeded), "mcp_servers") || strings.Contains(string(seeded), `command = "server"`) {
		t.Fatalf("inline native MCP table reached the managed config.toml: %s", seeded)
	}
	if !strings.Contains(string(seeded), `model = "gpt-5-codex"`) {
		t.Fatalf("stripping inline native MCP table lost other Codex config: %s", seeded)
	}

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
