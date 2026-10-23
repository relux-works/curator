package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
)

// cmdProfileMCP is deliberately separate from project install and legacy
// whole-file materialization. No CLI input can mint ownership-action authority.
func (c cli) cmdProfileMCP(cfg *config.Config, args []string) int {
	usage := "curator: profile mcp <install|status|repair|restore|replace|remove|release> <profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]"
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, usage)
		return exitUsage
	}
	mode, operation := args[0], ""
	switch mode {
	case "install", "status", "repair":
	case "restore", "replace", "remove", "release":
		operation = mode
		mode = "repair"
	default:
		_, _ = fmt.Fprintln(c.stderr, usage)
		return exitUsage
	}
	flags := c.newFlagSet("profile mcp")
	environments := flags.String("env", "", "comma-separated canonical environment ids (default: claude_code,codex_cli)")
	scope := flags.String("scope", "profile", "writer scope; project scope refuses shared-home writes")
	dry := flags.Bool("dry-run", false, "read-only preview; never recover or publish")
	takeover := flags.Bool("takeover", false, "request takeover (ownership-action authority is unavailable)")
	positional, err := parseInterspersed(flags, args[1:])
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, usage)
		return exitUsage
	}
	if *takeover {
		operation = "takeover"
	}
	var envs []string
	if *environments != "" {
		envs = strings.Split(*environments, ",")
	}
	result, err := envprofile.ProfileMCP(envprofile.ProfileMCPRequest{Home: cfg.Home(), Profile: positional[0], Scope: *scope, Mode: mode, Operation: operation, Environments: envs, DryRun: *dry, Config: cfg})
	if result != nil {
		if encodeErr := json.NewEncoder(c.stdout).Encode(result); encodeErr != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", encodeErr)
			return exitFail
		}
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
