package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

// machineFromConfig threads the effective machine knobs used by environment
// resolution and status: passable_env_names with its presence bit, and the
// isolation map, pre-overlay user values, and system-lock source.
func machineFromConfig(cfg *config.Config) envregistry.MachineConfig {
	machine := envregistry.DefaultMachineConfig()
	if cfg == nil {
		return machine
	}
	machine.PassableEnvNames = cfg.Env.PassableEnvNames
	machine.PassableEnvNamesSet = cfg.Env.PassableEnvNamesSet
	machine.Isolation = cfg.Env.Isolation
	machine.UserIsolation = cfg.UserIsolation
	machine.IsolationLocked = cfg.Locked["environments.isolation"]
	machine.IsolationLockSource = cfg.SystemConfigPath
	machine.Permissions = cfg.Env.Permissions
	machine.PermissionsLocked = cfg.Locked["environments.permissions"]
	return machine
}

// profileNativeHomeResolver keeps injected CLI homes consistent across the
// profile and legacy global materializers. Production commands leave the
// resolver nil and honor the adapter environment variables directly.
func (c cli) profileNativeHomeResolver() func(string) (string, error) {
	if c.userHome == nil {
		return nil
	}
	userHome, err := c.userHome()
	if err != nil {
		return func(string) (string, error) { return "", err }
	}
	return func(id string) (string, error) {
		switch id {
		case envregistry.ClaudeCode:
			return filepath.Join(userHome, ".claude"), nil
		case envregistry.CodexCLI:
			return filepath.Join(userHome, ".codex"), nil
		case envregistry.OpenCode:
			return filepath.Join(userHome, ".config", "opencode"), nil
		case envregistry.Muse:
			return filepath.Join(userHome, ".config", "muse"), nil
		case envregistry.Pi:
			return filepath.Join(userHome, ".pi"), nil
		default:
			return "", fmt.Errorf("%s: unregistered environment %q", envregistry.DiagUnknown, id)
		}
	}
}

func (c cli) cmdEnv(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env needs a subcommand: resolve | status | config | migrate | unmanage")
		return exitUsage
	}
	if wantsHelp(args) {
		_, _ = fmt.Fprint(c.stdout, "usage: curator env <resolve|status|config|migrate|unmanage> [flags]\n\nRequires the global config: "+configNotFoundHint+".\n")
		return exitOK
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	switch args[0] {
	case "resolve":
		return c.cmdEnvResolve(cfg, args[1:])
	case "status":
		return c.cmdEnvStatus(cfg, args[1:])
	case "config":
		return c.cmdEnvConfig(cfg, args[1:])
	case "migrate":
		return c.cmdEnvMigrate(cfg, args[1:])
	case "unmanage":
		return c.cmdEnvUnmanage(cfg, args[1:])
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown env subcommand %q\n", args[0])
	return exitUsage
}

func (c cli) cmdEnvUnmanage(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("env unmanage")
	restore := flags.Bool("restore-backups", false, "restore the newest takeover backup generation")
	environment := flags.String("env", "", "unmanage one environment (default: every environment)")
	target := flags.String("target", "", "unmanage one secondary fixed-home target")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env unmanage [--restore-backups] [--env <env-id>] [--target <target-id>]")
		return exitUsage
	}
	result, err := envprofile.Unmanage(envprofile.UnmanageRequest{
		Home:           cfg.Home(),
		EnvID:          *environment,
		TargetID:       *target,
		RestoreBackups: *restore,
		BackupLstat:    c.unmanageBackupLstat,
		BackupReadDir:  c.unmanageBackupReadDir,
	})
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	for _, home := range result.Homes {
		_, _ = fmt.Fprintf(c.stdout, "unmanaged %s: removed %d surface(s), restored %d backup file(s)\n",
			home.Environment, len(home.Removed), len(home.Restored))
		if !*restore {
			_, _ = fmt.Fprintf(c.stderr, "notice: any backup generations remain at %s\n", home.BackupPath)
		}
	}
	return exitOK
}

func (c cli) cmdEnvResolve(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("env resolve")
	profile := flags.String("profile", "", "profile name (default: the current profile)")
	repair := flags.Bool("repair", false, "provision or repair the managed home under the mutation lock")
	dryRun := flags.Bool("dry-run", false, "report store rebuilds without modifying state (requires --repair)")
	takeover := flags.Bool("takeover", false, "take over the unmanaged files the repair would write (only with --repair)")
	format := flags.String("format", "json", "fragment format: json | env | shell")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env resolve <env-id> [--profile <name>] [--repair [--dry-run]] [--takeover] [--format json|env|shell]")
		_, _ = fmt.Fprintln(c.stderr, envAliasUsage)
		return exitUsage
	}
	// --takeover applies only with --repair (cli/curator.md): without a
	// repair there is no write to take over.
	if *takeover && !*repair {
		_, _ = fmt.Fprintln(c.stderr, "curator: env resolve --takeover applies only with --repair")
		return exitUsage
	}
	if *dryRun && !*repair {
		_, _ = fmt.Fprintln(c.stderr, "curator: env resolve --dry-run requires --repair")
		return exitUsage
	}
	if *dryRun && *takeover {
		_, _ = fmt.Fprintln(c.stderr, "curator: env resolve --takeover is not available with --dry-run")
		return exitUsage
	}
	launchDir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	policy := envprofile.PolicyFromConfig(cfg)
	policy.Takeover = *takeover
	result, err := envprofile.Resolve(envprofile.ResolveRequest{
		Home:      cfg.Home(),
		Profile:   *profile,
		EnvID:     envregistry.NormalizeEnvID(positional[0]),
		LaunchDir: launchDir,
		Machine:   machineFromConfig(cfg),
		Repair:    *repair,
		DryRun:    *dryRun,
		Format:    *format,
		Policy:    policy,
	})
	if result != nil {
		for _, warning := range result.Warnings {
			_, _ = fmt.Fprintln(c.stderr, "warning:", warning)
		}
		if result.Notice != "" {
			_, _ = fmt.Fprintln(c.stderr, result.Notice)
		}
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprint(c.stdout, string(result.Document))
	return exitOK
}

func (c cli) cmdEnvStatus(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("env status")
	check := flags.Bool("check", false, "exit non-zero when any row is non-current")
	asJSON := flags.Bool("json", false, "render the matrix as JSON")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env status [--check] [--json]")
		return exitUsage
	}
	launchDir, _ := os.Getwd()
	status, err := envprofile.StatusOf(envprofile.StatusRequest{
		Home:      cfg.Home(),
		Machine:   machineFromConfig(cfg),
		LaunchDir: launchDir,
		Policy:    envprofile.PolicyFromConfig(cfg),
	})
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	status.SecurityPostureRows = envSecurityPostureRows(cfg)
	attachProviderPosture(cfg, status)
	attachRegistryPosture(cfg, status)
	if *asJSON {
		payload, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		_, _ = fmt.Fprintln(c.stdout, string(payload))
	} else {
		printEnvStatus(c.stdout, status)
	}
	if *check && status.NonCurrent {
		return exitFail
	}
	return exitOK
}
