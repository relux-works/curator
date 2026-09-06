package main

import (
	"fmt"
	"io"

	"github.com/relux-works/curator/internal/envprofile"
)

// printEnvStatus renders the profile × environment × surface matrix as
// human-readable rows. Machine-readable consumers use --json.
func printEnvStatus(stdout io.Writer, status *envprofile.Status) {
	// §12.2: env status reports the locked require_current_profile
	// requirement.
	if status.RequireCurrentProfile != nil {
		locked := "unlocked"
		if status.RequireCurrentLocked {
			locked = "locked"
		}
		_, _ = fmt.Fprintf(stdout, "require_current_profile: %s (%s)\n", *status.RequireCurrentProfile, locked)
	}
	for _, home := range status.Homes {
		state := "current"
		if !home.Current {
			state = "non-current"
		}
		provisioned := "unprovisioned"
		if home.Provisioned {
			provisioned = "provisioned"
		}
		_, _ = fmt.Fprintf(stdout, "%s %s: %s, %s, mode %s, form %s, lock %s\n", home.Profile, home.Environment, state, provisioned, home.Mode, home.Form, shortHash(home.LockHash))
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
		if len(home.SeededProjects) > 0 {
			_, _ = fmt.Fprintf(stdout, "  seeded-projects: %s\n", joinComma(home.SeededProjects))
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
		_, _ = fmt.Fprintf(stdout, "target %s (%s): participating=%v (%s); %s\n", target.ID, target.Adapter, target.Participating, target.Detail, target.Ungoverned)
	}
	for _, profile := range status.Profiles {
		_, _ = fmt.Fprintf(stdout, "profile %s: lock %s, precedence winner=%s placement=%s\n",
			profile.Profile, shortHash(profile.LockHash), profile.Precedence.Winner, profile.Precedence.Placement)
		for _, member := range profile.Members {
			_, _ = fmt.Fprintf(stdout, "  member %s %s weight %d\n", member.Kind, member.Name, member.Weight)
		}
	}
	for _, id := range status.UnregisteredEnvironments {
		_, _ = fmt.Fprintf(stdout, "unregistered environment in machine configuration: %s\n", id)
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
