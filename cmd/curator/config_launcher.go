package main

import (
	"encoding/json"
	"fmt"
)

// cmdConfigLauncher uses the same injected/configured source as every normal
// Core operation. It exports metadata only; credential resolution is separate.
func (c cli) cmdConfigLauncher() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	payload, err := json.Marshal(cfg.LauncherProjection())
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: cannot encode launcher configuration")
		return exitFail
	}
	_, err = fmt.Fprintln(c.stdout, string(payload))
	if err != nil {
		return exitFail
	}
	return exitOK
}

// cmdConfigLauncherStatus reports selection and eligibility only. No runtime
// is started or probed, and no native settings or environment are inspected.
func (c cli) cmdConfigLauncherStatus() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	return c.printJSON(cfg.ClaudeEnvStatus())
}

// cmdConfigLauncherContribution exports only the selected extension contribution;
// the separately installed launcher still owns actual child launch and resume.
func (c cli) cmdConfigLauncherContribution(runtime string) int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	contribution, err := cfg.ClaudeEnvContribution(runtime)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return c.printJSON(contribution)
}
