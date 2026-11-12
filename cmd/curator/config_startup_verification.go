package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/relux-works/curator/internal/config"
)

func (c cli) cmdConfigStartupVerification(args []string) int {
	jsonOutput := false
	var level int
	if len(args) == 1 && args[0] == "--json" {
		jsonOutput = true
	} else if len(args) == 2 && args[0] == "--level" && len(args[1]) == 1 && args[1] >= "1" && args[1] <= "4" {
		level, _ = strconv.Atoi(args[1])
	} else if len(args) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: usage: config startup-verification [--json | --level 1..4]")
		return exitUsage
	}
	if level != 0 {
		if err := config.SetStartupVerification(context.Background(), c.config.Path(), level); err != nil {
			_, _ = fmt.Fprintln(c.stderr, err)
			return exitFail
		}
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	if jsonOutput {
		payload, err := json.Marshal(cfg.StartupProjection())
		if err != nil {
			return exitFail
		}
		if _, err := fmt.Fprintln(c.stdout, string(payload)); err != nil {
			return exitFail
		}
	} else {
		if _, err := fmt.Fprintf(c.stdout, "Startup verification level: %d (%s)\n", cfg.StartupVerification.Level, cfg.StartupVerificationOrigin); err != nil {
			return exitFail
		}
	}
	return exitOK
}
