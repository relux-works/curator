package main

import (
	"fmt"
	"io"

	"github.com/relux-works/curator/internal/envprofile"
)

// printEnvStatus renders the profile × environment × surface matrix as
// human-readable rows. Machine-readable consumers use --json.
func printEnvStatus(stdout io.Writer, status *envprofile.Status) {
	for _, home := range status.Homes {
		state := "current"
		if !home.Current {
			state = "non-current"
		}
		provisioned := "unprovisioned"
		if home.Provisioned {
			provisioned = "provisioned"
		}
		_, _ = fmt.Fprintf(stdout, "%s %s: %s, %s, lock %s\n", home.Profile, home.Environment, state, provisioned, shortHash(home.LockHash))
		for _, surface := range home.Surfaces {
			_, _ = fmt.Fprintf(stdout, "  surface %s: %s\n", surface.Key, surface.State)
		}
		for _, finding := range home.Findings {
			_, _ = fmt.Fprintf(stdout, "  finding: %s\n", finding)
		}
		for _, warning := range home.Warnings {
			_, _ = fmt.Fprintf(stdout, "  warning: %s\n", warning)
		}
		if len(home.Passthrough) > 0 {
			_, _ = fmt.Fprintf(stdout, "  passthrough: %s\n", joinComma(home.Passthrough))
		}
		if len(home.Seeds) > 0 {
			_, _ = fmt.Fprintf(stdout, "  seeds: %s\n", joinComma(home.Seeds))
		}
		if len(home.SeedLinks) > 0 {
			_, _ = fmt.Fprintf(stdout, "  seed-links: %s\n", joinComma(home.SeedLinks))
		}
		if home.Backups > 0 {
			_, _ = fmt.Fprintf(stdout, "  backups: %d generations, oldest %s, newest %s\n", home.Backups, home.BackupsOldest, home.BackupsNewest)
		}
	}
	for _, scope := range status.Scopes {
		provisioned := "unprovisioned"
		if scope.Provisioned {
			provisioned = "provisioned"
		}
		_, _ = fmt.Fprintf(stdout, "scope %s: profile %s env %s native %s managed %s (%s)\n",
			scope.Scope, scope.Profile, scope.Environment, scope.Native, scope.Managed, provisioned)
	}
	for _, adapter := range status.Adapters {
		_, _ = fmt.Fprintf(stdout, "tool %s: recorded %s detected %s\n", adapter.ID, adapter.Recorded, adapter.Detected)
	}
	for _, target := range status.Targets {
		_, _ = fmt.Fprintf(stdout, "target %s: participating=%v (%s); %s\n", target.ID, target.Participating, target.Detail, target.Ungoverned)
	}
	for _, orphan := range status.Orphans {
		_, _ = fmt.Fprintf(stdout, "orphan: %s\n", orphan)
	}
	for _, note := range status.Notes {
		_, _ = fmt.Fprintf(stdout, "note: %s\n", note)
	}
}

func shortHash(hash string) string {
	if len(hash) > 19 {
		return hash[:19] + "…"
	}
	return hash
}

func joinComma(values []string) string {
	out := ""
	for i, value := range values {
		if i > 0 {
			out += ", "
		}
		out += value
	}
	return out
}
