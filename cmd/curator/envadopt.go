package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

func (c cli) cmdEnvAdopt(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("env adopt")
	profile := flags.String("profile", "", "exact installed profile (required)")
	operator := flags.String("operator-home", "", "original operator-home anchor (required, explicit)")
	native := flags.String("native-root", "", "original absolute native root (required, explicit)")
	state := flags.String("selector-state", "", "original selector state: absent | present-empty | present-value")
	value := flags.String("selector-value", "", "original selector value (required for either present state)")
	positional, err := parseInterspersed(flags, args)
	provided := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	usage := func() int {
		_, _ = fmt.Fprintln(c.stderr, "curator: env adopt <env-id> --profile <name> --operator-home <path> --native-root <path> --selector-state absent|present-empty|present-value [--selector-value <original-value>]")
		return exitUsage
	}
	if err != nil || len(positional) != 1 || *profile == "" || *operator == "" || *native == "" || *state == "" {
		return usage()
	}
	var original *string
	switch *state {
	case "absent":
		if provided["selector-value"] {
			return usage()
		}
	case "present-empty":
		if !provided["selector-value"] || *value != "" {
			return usage()
		}
		original = value
	case "present-value":
		if !provided["selector-value"] || *value == "" {
			return usage()
		}
		original = value
	default:
		return usage()
	}
	launchDir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	changed, err := envprofile.AdoptOrigin(envprofile.OriginAdoptRequest{Resolve: envprofile.ResolveRequest{Home: cfg.Home(), Profile: *profile, EnvID: envregistry.NormalizeEnvID(positional[0]), LaunchDir: launchDir, Machine: machineFromConfig(cfg), Policy: envprofile.PolicyFromConfig(cfg)}, OperatorHome: *operator, NativeRoot: *native, SelectorState: *state, SelectorValue: original})
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	if changed {
		_, _ = fmt.Fprintln(c.stdout, "provisioning origin adopted")
	} else {
		_, _ = fmt.Fprintln(c.stdout, "provisioning origin already matches")
	}
	return exitOK
}
