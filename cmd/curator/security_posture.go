package main

import (
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envfragment"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/shell"
)

func securityPostureRevisions() config.SecurityPostureRevisions {
	providerTrustRoots := "revision-A"
	if activeProviderRevision == providerRevisionB {
		providerTrustRoots = "revision-B"
	}
	return config.SecurityPostureRevisions{
		HookTrust:          string(shell.DefaultTrustProfile),
		EnvPassthrough:     string(envfragment.ActiveS4Profile),
		ProviderTrustRoots: providerTrustRoots,
		UpdateConfirmation: "A-warning",
		CodexSeed:          envregistry.CodexSeedRevisionB,
	}
}

func curatorSecurityPostureRows(cfg *config.Config) []config.SecurityPostureRow {
	return cfg.SecurityPostureStatusRows(securityPostureRevisions())
}

func envSecurityPostureRows(cfg *config.Config) []config.SecurityPostureRow {
	if cfg == nil || cfg.Schema == config.SchemaVersion {
		return nil
	}
	return cfg.SecurityPostureStatusRows(securityPostureRevisions())
}
