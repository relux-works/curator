package main

import (
	"flag"
	"fmt"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/envregistry"
)

// parseOriginAttestation retains presence separately from explicit empty input.
// Historical inputs are never filled from current HOME, selectors or links.
func parseOriginAttestation(flags *flag.FlagSet, args []string) (envorigin.AdoptRequest, error) {
	profile := flags.String("profile", "", "exact installed profile")
	target := flags.String("managed-target", "", "exact absolute managed home")
	lock := flags.String("lock-sha256", "", "exact installed sha256 lock identity")
	anchor := flags.String("operator-home", "", "verified original absolute account anchor")
	state := flags.String("selector-state", "", "original absent|present-empty|present-value")
	value := flags.String("selector-value", "", "explicit original selector value (including empty)")
	native := flags.String("native-root", "", "verified original absolute native root")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return envorigin.AdoptRequest{}, err
	}
	present := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { present[f.Name] = true })
	for _, name := range []string{"profile", "managed-target", "lock-sha256", "operator-home", "selector-state", "native-root"} {
		if !present[name] {
			return envorigin.AdoptRequest{}, fmt.Errorf("env adopt requires explicit --%s", name)
		}
	}
	if len(positional) != 1 {
		return envorigin.AdoptRequest{}, fmt.Errorf("env adopt requires one adapter id")
	}
	switch *state {
	case "absent", "present-empty", "present-value":
	default:
		return envorigin.AdoptRequest{}, fmt.Errorf("env adopt selector-state must be absent, present-empty or present-value")
	}
	if *state == "absent" && present["selector-value"] || *state != "absent" && !present["selector-value"] {
		return envorigin.AdoptRequest{}, fmt.Errorf("env adopt selector-state/value presence differs")
	}
	if *state == "present-empty" && *value != "" || *state == "present-value" && *value == "" {
		return envorigin.AdoptRequest{}, fmt.Errorf("env adopt selector-state/value content differs")
	}
	req := envorigin.AdoptRequest{Adapter: envregistry.NormalizeEnvID(positional[0]), Binding: envorigin.Binding{Profile: *profile, ManagedTarget: *target, LockSHA256: *lock}, OperatorHome: *anchor, SelectorState: *state, NativeRoot: *native}
	if present["selector-value"] {
		req.SelectorValue = value
	}
	return req, nil
}
func (c cli) cmdEnvAdopt(cfg *config.Config, args []string) int {
	attestation, err := parseOriginAttestation(c.newFlagSet("env adopt"), args)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	if err := envprofile.AdoptOrigin(envprofile.AdoptOriginRequest{Home: cfg.Home(), Attestation: attestation, Machine: machineFromConfig(cfg)}); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
