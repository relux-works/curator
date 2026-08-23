package main

import (
	"context"
	"fmt"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
)

// resolveCLIProvider is the platform-neutral installation seam for a future
// separately installed host-execution-provider-v1. This release ships no host
// provider, so explicit verified selection fails closed before cache or child
// process activity.
var resolveCLIProvider = func(config.Execution) install.VerifiedBuildSessionProvider { return nil }

// preflightCLIExecution is the single production boundary between machine
// assurance selection and every CLI path that can inspect a build cache or
// start a compiler process.
func preflightCLIExecution(ctx context.Context, cfg *config.Config) (*install.BuildAuthority, error) {
	if cfg == nil {
		return nil, fmt.Errorf("execution configuration is absent")
	}
	selected := closureexec.AssuranceConfig{
		Mode:                  closureexec.AssuranceMode(cfg.Execution.Mode),
		ProviderID:            cfg.Execution.ProviderID,
		ProviderVersion:       cfg.Execution.ProviderVersion,
		ProviderBinarySHA256:  closuregraph.ID(cfg.Execution.ProviderBinarySHA256),
		ProviderTrustEvidence: cfg.Execution.ProviderTrustEvidence,
	}
	return install.NewBuildAuthority(ctx, selected, resolveCLIProvider(cfg.Execution))
}
