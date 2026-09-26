package main

import (
	"fmt"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

// cmdEnvMigrate runs the explicit credential migration step
// (environments §7.4, §10.1; manager §12.4): inspect inventories,
// plan prints the exact operations with their hash, and apply executes
// exactly the printed plan under the manager lock after the drift and
// conflict checks. Inspect and plan are read-only; only apply mutates.
// Apply requires the --expect hash of a prior --plan and prints the
// locked, revalidated plan before the first mutation; the library owns
// both guarantees, so a missing hash is a usage error here and a
// refusal in the library.
func (c cli) cmdEnvMigrate(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("env migrate")
	inspect := flags.Bool("inspect", false, "print the credential inventory per profile and environment")
	plan := flags.Bool("plan", false, "print the exact migration operations with the plan hash")
	apply := flags.Bool("apply", false, "execute the printed plan under the manager lock")
	profile := flags.String("profile", "", "migrate one profile (default: every installed profile)")
	env := flags.String("env", "", "migrate one environment (default: every registered environment)")
	expect := flags.String("expect", "", "the plan hash a prior --plan printed (required with --apply)")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env migrate --inspect|--plan|--apply [--profile <name>] [--env <env-id>] [--expect <plan-hash>]")
		_, _ = fmt.Fprintln(c.stderr, envAliasUsage)
		return exitUsage
	}
	phases := 0
	for _, selected := range []bool{*inspect, *plan, *apply} {
		if selected {
			phases++
		}
	}
	if phases != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env migrate needs exactly one of --inspect, --plan, or --apply")
		return exitUsage
	}
	// --expect anchors a printed plan to its apply: without an apply
	// there is no drift to check, mirroring --takeover with --repair.
	if *expect != "" && !*apply {
		_, _ = fmt.Fprintln(c.stderr, "curator: env migrate --expect applies only with --apply")
		return exitUsage
	}
	// Every apply requires a prior complete plan identity: without the
	// hash there is no printed plan to execute, so refuse before the
	// library runs (which refuses again for direct callers).
	if *apply && *expect == "" {
		_, _ = fmt.Fprintln(c.stderr, "curator: env migrate --apply needs --expect <plan-hash> from a prior --plan")
		return exitUsage
	}
	envID := ""
	if *env != "" {
		envID = envregistry.NormalizeEnvID(*env)
	}
	req := envprofile.MigrateRequest{
		Home:    cfg.Home(),
		Profile: *profile,
		EnvID:   envID,
		Machine: machineFromConfig(cfg),
		Expect:  *expect,
	}
	if *apply {
		// The library prints the locked, revalidated plan to stdout
		// before the first mutation (and refuses when the print
		// fails); this command only reports what the plan did.
		req.Print = c.stdout
		result, err := envprofile.ApplyMigration(req)
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		if result != nil && result.Recovered != nil {
			_, _ = fmt.Fprintf(c.stdout, "recovered interrupted migration apply %s: restored the prior state before executing\n", result.Recovered.Plan)
		}
		if result != nil && len(result.Applied) > 0 {
			_, _ = fmt.Fprintf(c.stdout, "applied migration plan %s: %d operations\n", result.Report.Hash, len(result.Applied))
			for _, op := range result.Applied {
				switch op.Kind {
				case envprofile.MigrateOpRelink:
					_, _ = fmt.Fprintf(c.stdout, "  relinked profile %s, environment %s, %s -> %s\n", op.Profile, op.EnvID, op.Path, op.To)
				case envprofile.MigrateOpUnlink:
					_, _ = fmt.Fprintf(c.stdout, "  unlinked profile %s, environment %s, %s\n", op.Profile, op.EnvID, op.Path)
				}
			}
			_, _ = fmt.Fprintln(c.stdout, "migration complete")
		}
		return exitOK
	}
	report, err := envprofile.PlanMigration(req)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	phase := envprofile.MigratePhasePlan
	if *inspect {
		phase = envprofile.MigratePhaseInspect
	}
	_, _ = fmt.Fprint(c.stdout, report.Text(phase))
	return exitOK
}
