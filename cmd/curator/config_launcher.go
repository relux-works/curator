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
