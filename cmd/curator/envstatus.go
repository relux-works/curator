package main

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/config"
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
	for _, warning := range status.ShellHookTrustWarnings {
		_, _ = fmt.Fprintf(stdout, "shell-hook-trust: warning: %s\n", warning)
	}
	for _, row := range status.ShellHookTrust {
		_, _ = fmt.Fprintln(stdout, formatTrustRow(row))
	}
	// §12 posture: the active S4 profile with the effective
	// passable_env_names, the machine-level warnings, and the §2.3
	// surfacing rows for the current profile of each reported scope.
	_, _ = fmt.Fprintf(stdout, "s4_profile: %s, passable_env_names: %s\n", status.S4Profile, formatPassable(status.PassableEnvNames))
	for _, warning := range status.Warnings {
		_, _ = fmt.Fprintf(stdout, "warning: %s\n", warning)
	}
	for _, scope := range status.MCPDeclarations {
		_, _ = fmt.Fprintf(stdout, "scope %s profile %s mcp-declarations:\n", scope.Scope, scope.Profile)
		for _, row := range scope.Rows {
			_, _ = fmt.Fprintln(stdout, row)
		}
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
	for _, provider := range status.Providers {
		_, _ = fmt.Fprintf(stdout, "provider %s: %s\n", provider.Name, describeProvider(provider))
	}
	for _, target := range status.Targets {
		_, _ = fmt.Fprintf(stdout, "target %s (%s): participating=%v (%s); %s\n", target.ID, target.Adapter, target.Participating, target.Detail, target.Ungoverned)
	}
	for _, profile := range status.Profiles {
		_, _ = fmt.Fprintf(stdout, "profile %s: lock %s, precedence winner=%s placement=%s, transitive_system_modules=%s\n",
			profile.Profile, shortHash(profile.LockHash), profile.Precedence.Winner, profile.Precedence.Placement, profile.TransitiveSystemModules)
		for _, member := range profile.Members {
			_, _ = fmt.Fprintf(stdout, "  member %s %s weight %d\n", member.Kind, member.Name, member.Weight)
		}
		for _, dropped := range profile.DroppedSystemModules {
			_, _ = fmt.Fprintf(stdout, "  dropped system module %s %s\n", dropped.Package, dropped.Path)
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

// describeProvider renders one §12 umbrella provider row: the resolved
// absolute provider path with its trust verdict. A refused or failed
// provider row is non-current; the revision-A outside-roots warning row
// stays current.
func describeProvider(provider envprofile.ProviderState) string {
	state := "current"
	if !provider.Current {
		state = "non-current"
	}
	switch provider.Verdict {
	case envprofile.ProviderTrusted:
		return fmt.Sprintf("%s (trusted, %s)", derefProvider(provider.Resolved), state)
	case envprofile.ProviderOutsideTrustRoots:
		return fmt.Sprintf("%s (%s, %s)", derefProvider(provider.Resolved), derefProvider(provider.Diagnostic), state)
	case envprofile.ProviderRefused:
		return fmt.Sprintf("%s (%s, %s)", derefProvider(provider.RefusedPath), derefProvider(provider.Diagnostic), state)
	case envprofile.ProviderUnreadable:
		return fmt.Sprintf("unreadable trust root %s (%s, %s)", derefProvider(provider.UnreadableDirectory), derefProvider(provider.Diagnostic), state)
	default:
		return fmt.Sprintf("missing (%s, %s)", derefProvider(provider.Diagnostic), state)
	}
}

// derefProvider renders an optional provider field, never empty.
func derefProvider(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}

// attachProviderPosture resolves the §12 umbrella provider rows for the
// loaded machine configuration and folds their currency into the
// matrix: a refused, missing, or unreadable provider row is
// non-current, so env status --check fails while one stands.
func attachProviderPosture(cfg *config.Config, status *envprofile.Status) {
	var projectBins []string
	for _, project := range cfg.Projects {
		projectBins = append(projectBins, filepath.Join(project.Path, ".agents", "bin"))
	}
	sort.Strings(projectBins)
	status.Providers = providerPosture(cfg.Home(), cfg.Env.ProviderDirectories, projectBins)
	for _, provider := range status.Providers {
		if !provider.Current {
			status.NonCurrent = true
		}
	}
}

// formatPassable renders the effective passable_env_names: unbounded for
// a nil list, otherwise the compact JSON array.
func formatPassable(names []string) string {
	if names == nil {
		return "unbounded"
	}
	out := "["
	for i, name := range names {
		if i > 0 {
			out += ","
		}
		out += `"` + name + `"`
	}
	return out + "]"
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
