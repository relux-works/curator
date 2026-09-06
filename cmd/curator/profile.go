package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
)

// cmdProfile dispatches the profile family (cli/curator.md): install, list,
// use, update, remove, sync. Composition (compose) edits machine
// configuration of manager-config schema 2, which is out of scope for this
// stage, so it refuses with guidance instead of a silent no-op.
func (c cli) cmdProfile(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile needs a subcommand: install | list | use | update | remove | sync")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	switch args[0] {
	case "install":
		return c.cmdProfileInstall(cfg, args[1:])
	case "list":
		return c.cmdProfileList(cfg, args[1:])
	case "use":
		return c.cmdProfileUse(cfg, args[1:])
	case "update":
		return c.cmdProfileUpdate(cfg, args[1:])
	case "remove":
		return c.cmdProfileRemove(cfg, args[1:])
	case "sync":
		return c.cmdProfileSync(cfg, args[1:])
	case "compose":
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose edits the machine overlays.<profile> list of manager-config schema 2, which this stage does not implement; the lock moves only on profile update")
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown profile subcommand %q\n", args[0])
	return exitUsage
}

func (c cli) cmdProfileInstall(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("profile install")
	directory := flags.String("directory", "", "directory within a git snapshot")
	rng := flags.String("range", "", "version range over tags")
	tag := flags.String("tag", "", "exact version tag")
	revision := flags.String("revision", "", "exact commit")
	as := flags.String("as", "", "profile name (default: the root package name)")
	use := flags.Bool("use", false, "activate the installed profile")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile install <git-url|path> [--directory <dir>] [--range <range>|--tag <tag>|--revision <commit>] [--as <name>] [--use]")
		return exitUsage
	}
	info, activated, updated, err := envprofile.Install(cfg.Home(), envprofile.InstallOptions{
		Operand: positional[0], Directory: *directory,
		Range: *rng, Tag: *tag, Revision: *revision, As: *as, Use: *use,
		Policy: envprofile.PolicyFromConfig(cfg),
	})
	for _, warning := range info.Warnings {
		_, _ = fmt.Fprintln(c.stderr, "warning:", warning)
	}
	for _, result := range info.Activation {
		if result.OK {
			_, _ = fmt.Fprintf(c.stdout, "%s: switched (%s)\n", result.Adapter, result.Home)
		} else {
			_, _ = fmt.Fprintf(c.stderr, "%s: %s\n", result.Adapter, result.Detail)
		}
	}
	if err != nil {
		// An install activation that could not materialize the whole
		// scope leaves the lock installed but the current unchanged;
		// report the partial switch like profile use does.
		if len(info.Activation) > 0 && info.Name != "" {
			_, _ = fmt.Fprintf(c.stdout, "installed profile %s (lock %s)\n", info.Name, info.LockHash)
		}
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	root, _ := info.Lock.RootMember()
	switch {
	case updated:
		_, _ = fmt.Fprintf(c.stdout, "updated profile %s (lock %s)\n", info.Name, info.LockHash)
	case activated:
		_, _ = fmt.Fprintf(c.stdout, "installed and activated profile %s (root %s %s, lock %s)\n", info.Name, root.Name, root.Version, info.LockHash)
	default:
		_, _ = fmt.Fprintf(c.stdout, "installed profile %s (root %s %s, lock %s); activate with 'curator profile use %s'\n", info.Name, root.Name, root.Version, info.LockHash, info.Name)
	}
	return exitOK
}

func (c cli) cmdProfileList(cfg *config.Config, args []string) int {
	if len(args) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile list takes no arguments")
		return exitUsage
	}
	profiles, err := envprofile.ListWithPolicy(cfg.Home(), envprofile.PolicyFromConfig(cfg))
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	machine, err := envprofile.Current(cfg.Home())
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	for _, info := range profiles {
		root, _ := info.Lock.RootMember()
		requirement := info.Source.Req.Range
		form := "range"
		if info.Source.Req.Tag != "" {
			requirement, form = info.Source.Req.Tag, "tag"
		} else if info.Source.Req.Revision != "" {
			requirement, form = info.Source.Req.Revision, "revision"
		}
		source := info.Source.Git
		switch info.Source.Kind {
		case envprofile.KindPath:
			source, form, requirement = info.Source.Path, "path", info.Source.Path
		case envprofile.KindLocal:
			source, form, requirement = "-", "local", "-"
		}
		markers := []string{}
		if machine == info.Name {
			markers = append(markers, "current")
		}
		for _, scope := range info.ScopedFor {
			markers = append(markers, scope+"="+info.Name)
		}
		sort.Strings(markers)
		_, _ = fmt.Fprintf(c.stdout, "%s\t%s\t%s\t%s\t%s %s\t%s\t%s\t%s\n",
			info.Name, root.Name, info.Source.Kind, source, form, requirement,
			root.Version, info.LockHash, strings.Join(markers, ","))
	}
	return exitOK
}

func (c cli) cmdProfileUse(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("profile use")
	env := flags.String("env", "", "narrow the switch to one environment")
	target := flags.String("target", "", "narrow the switch to one target")
	clearScope := flags.Bool("clear", false, "drop the scoped current profile")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile use <name> [--env <env-id>] [--target <target-id>] | profile use --clear --env <env-id>|--target <target-id>")
		return exitUsage
	}
	name := ""
	if len(positional) == 1 {
		name = positional[0]
	} else if len(positional) > 1 || !*clearScope || (*env == "" && *target == "") {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile use <name> [--env <env-id>] [--target <target-id>] | profile use --clear --env <env-id>|--target <target-id>")
		return exitUsage
	}
	results, err := envprofile.UseWithPolicy(cfg.Home(), name, *env, *target, *clearScope, envprofile.PolicyFromConfig(cfg))
	for _, result := range results {
		if result.OK {
			_, _ = fmt.Fprintf(c.stdout, "%s: switched (%s)\n", result.Adapter, result.Home)
		} else {
			_, _ = fmt.Fprintf(c.stderr, "%s: %s\n", result.Adapter, result.Detail)
		}
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}

func (c cli) cmdProfileUpdate(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("profile update")
	all := flags.Bool("all", false, "update every installed profile")
	positional, err := parseInterspersed(flags, args)
	if err != nil || (*all && len(positional) != 0) || (!*all && len(positional) > 1) {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile update [<name>|--all]")
		return exitUsage
	}
	policy := envprofile.PolicyFromConfig(cfg)
	var names []string
	if *all {
		profiles, err := envprofile.ListWithPolicy(cfg.Home(), policy)
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		for _, info := range profiles {
			if info.Name == envprofile.DefaultProfile {
				continue
			}
			names = append(names, info.Name)
		}
	} else if len(positional) == 1 {
		names = []string{positional[0]}
	} else {
		machine, err := envprofile.Current(cfg.Home())
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		if machine == "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: no current profile")
			return exitFail
		}
		names = []string{machine}
	}
	failed := false
	for _, name := range names {
		info, moved, err := envprofile.UpdateWithPolicy(cfg.Home(), name, policy)
		if err != nil {
			_, _ = fmt.Fprintf(c.stderr, "%s: %v\n", name, err)
			failed = true
			continue
		}
		for _, warning := range info.Warnings {
			_, _ = fmt.Fprintf(c.stderr, "%s: warning: %s\n", name, warning)
		}
		if moved {
			_, _ = fmt.Fprintf(c.stdout, "%s: updated (lock %s)\n", info.Name, info.LockHash)
		} else {
			_, _ = fmt.Fprintf(c.stdout, "%s: unchanged\n", info.Name)
		}
	}
	if failed {
		return exitFail
	}
	return exitOK
}

func (c cli) cmdProfileRemove(cfg *config.Config, args []string) int {
	flags := c.newFlagSet("profile remove")
	purge := flags.Bool("purge", false, "remove managed homes with markers and backups")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile remove <name> [--purge]")
		return exitUsage
	}
	if err := envprofile.Remove(cfg.Home(), positional[0], *purge); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "removed profile %s\n", positional[0])
	return exitOK
}

func (c cli) cmdProfileSync(cfg *config.Config, args []string) int {
	if len(args) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile sync takes no arguments")
		return exitUsage
	}
	results, err := envprofile.SyncWithPolicy(cfg.Home(), envprofile.PolicyFromConfig(cfg))
	for _, result := range results {
		if result.OK {
			_, _ = fmt.Fprintf(c.stdout, "%s: synced (%s)\n", result.Adapter, result.Home)
		} else {
			_, _ = fmt.Fprintf(c.stderr, "%s: %s\n", result.Adapter, result.Detail)
		}
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
