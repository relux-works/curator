package main

import (
	"encoding/json"
	"fmt"
	"os"

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

func (c cli) cmdEnv(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env needs a subcommand: resolve | status | config | migrate")
		return exitUsage
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
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown env subcommand %q\n", args[0])
	return exitUsage
}

func (c cli) cmdEnvResolve(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("env resolve")
	profile := flags.String("profile", "", "profile name (default: the current profile)")
	repair := flags.Bool("repair", false, "provision or repair the managed home under the mutation lock")
	takeover := flags.Bool("takeover", false, "take over the unmanaged files the repair would write (only with --repair)")
	format := flags.String("format", "json", "fragment format: json | env | shell")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: env resolve <env-id> [--profile <name>] [--repair] [--takeover] [--format json|env|shell]")
		_, _ = fmt.Fprintln(c.stderr, envAliasUsage)
		return exitUsage
	}
	// --takeover applies only with --repair (cli/curator.md): without a
	// repair there is no write to take over.
	if *takeover && !*repair {
		_, _ = fmt.Fprintln(c.stderr, "curator: env resolve --takeover applies only with --repair")
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
	attachProviderPosture(cfg, status)
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
