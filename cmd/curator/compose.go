package main

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/identity"
)

// cmdProfileCompose implements `profile compose <profile> add|remove|list`
// (cli/curator.md, environments §12.1): add and remove edit the machine
// overlays.<profile> list of manager-config schema 2 and, for an installed
// profile with a lock, rebuild that lock from the complete resulting
// composition under the same held profile operation before the declaration
// is written. Before a profile has a lock only the declaration changes.
func (c cli) cmdProfileCompose(cfg *config.Config, args []string) int {
	if len(args) < 2 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> add|remove|list [<source> [--range|--tag|--revision <ref>]] [--weight <n>]")
		return exitUsage
	}
	profile, action := args[0], args[1]
	switch action {
	case "add":
		return c.cmdComposeAdd(cfg, profile, args[2:])
	case "remove":
		return c.cmdComposeRemove(cfg, profile, args[2:])
	case "list":
		return c.cmdComposeList(cfg, profile, args[2:])
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown profile compose action %q\n", action)
	return exitUsage
}

func (c cli) cmdComposeAdd(cfg *config.Config, profile string, args []string) int {
	flags := c.newFlagSet("profile compose add")
	rng := flags.String("range", "", "version range over tags")
	tag := flags.String("tag", "", "exact version tag")
	revision := flags.String("revision", "", "exact commit")
	directory := flags.String("directory", "", "directory within the overlay snapshot")
	weight := flags.String("weight", "", "machine-assigned weight (default: overlay_default_weight)")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> add <source> [--range|--tag|--revision <ref>] [--directory <dir>] [--weight <n>] (a git source takes exactly one requirement form; a path source takes none)")
		return exitUsage
	}
	forms := 0
	for _, form := range []string{*rng, *tag, *revision} {
		if form != "" {
			forms++
		}
	}
	// The requirement form is required only for a git source
	// (manager-config-v2 $defs/overlay, environments §1 and §12.1). The
	// kind comes from the one discriminator, the same helper the config
	// reader and resolveOverlay use, so a row written here parses back as
	// the kind this row decided.
	source := positional[0]
	switch identity.ClassifySource(source) {
	case identity.SourceGit:
		if forms != 1 {
			_, _ = fmt.Fprintln(c.stderr, "curator: profile compose add takes exactly one of --range, --tag, --revision for a git source")
			return exitUsage
		}
	case identity.SourcePath:
		if forms != 0 {
			_, _ = fmt.Fprintln(c.stderr, "curator: a path overlay carries no --range, --tag, or --revision")
			return exitUsage
		}
		if *directory != "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: a path overlay carries no --directory")
			return exitUsage
		}
	default:
		_, _ = fmt.Fprintf(c.stderr, "curator: overlay source %q is neither a git source nor a path\n", source)
		return exitUsage
	}
	entry := map[string]any{"source": source}
	switch {
	case *rng != "":
		entry["range"] = *rng
	case *tag != "":
		entry["tag"] = *tag
	case *revision != "":
		entry["revision"] = *revision
	}
	if *directory != "" {
		entry["directory"] = *directory
	}
	if *weight != "" {
		parsed, err := strconv.Atoi(*weight)
		if err != nil || parsed < 0 {
			_, _ = fmt.Fprintln(c.stderr, "curator: --weight must be a non-negative integer")
			return exitUsage
		}
		entry["weight"] = float64(parsed)
	}
	object, err := config.ReadRawOrDefault(cfg.Path)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	env, _ := object["environments"].(map[string]any)
	if env == nil {
		env = map[string]any{}
		object["environments"] = env
	}
	overlays, _ := env["overlays"].(map[string]any)
	if overlays == nil {
		overlays = map[string]any{}
		env["overlays"] = overlays
	}
	var list []any
	if raw, present := overlays[profile]; present {
		list, _ = raw.([]any)
	}
	// Rerunning the same add after a crash between the lock and declaration
	// writes must converge (environments §12.1), so an identical declaration
	// is not appended twice: the lock is still rebuilt from the list as is.
	if !containsOverlayEntry(list, entry) {
		list = append(list, entry)
	}
	overlays[profile] = list
	if schema, _ := object["schema_version"].(float64); schema == float64(config.SchemaVersion) {
		config.PromoteSchemaPreservingPosture(object, config.SchemaVersion2, cfg.EffectiveSecurityPosture())
	}
	return c.composePublish(cfg, profile, object, "added overlay "+source+" to profile "+profile)
}

// containsOverlayEntry reports whether list already carries a declaration
// equal to entry in every field (source, form, directory, weight).
func containsOverlayEntry(list []any, entry map[string]any) bool {
	for _, item := range list {
		existing, _ := item.(map[string]any)
		if existing != nil && reflect.DeepEqual(existing, entry) {
			return true
		}
	}
	return false
}

func (c cli) cmdComposeRemove(cfg *config.Config, profile string, args []string) int {
	flags := c.newFlagSet("profile compose remove")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> remove <source>")
		return exitUsage
	}
	source := positional[0]
	object, err := config.ReadRawOrDefault(cfg.Path)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	env, _ := object["environments"].(map[string]any)
	var list []any
	if env != nil {
		overlays, _ := env["overlays"].(map[string]any)
		if overlays != nil {
			list, _ = overlays[profile].([]any)
		}
	}
	kept := make([]any, 0, len(list))
	for _, item := range list {
		entry, _ := item.(map[string]any)
		if entry == nil {
			kept = append(kept, item)
			continue
		}
		if name, _ := entry["source"].(string); name != source {
			kept = append(kept, item)
		}
	}
	if len(kept) == len(list) {
		_, _ = fmt.Fprintf(c.stderr, "curator: profile %s has no overlay %q\n", profile, source)
		return exitFail
	}
	overlays, _ := env["overlays"].(map[string]any)
	overlays[profile] = kept
	return c.composePublish(cfg, profile, object, "removed overlay "+source+" from profile "+profile)
}

// composePublish validates the complete resulting declaration, rebuilds the
// installed profile's lock from it, and writes the declaration — in that
// order, under one held profile operation (environments §12.1). Nothing is
// written when the resulting composition fails to validate, resolve, or
// audit. An uninstalled profile, or one without a lock, gets only the
// declaration; its first install resolves the overlays jointly.
func (c cli) composePublish(cfg *config.Config, profile string, object map[string]any, summary string) int {
	candidate, err := config.Parse(object, cfg.Path)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	policy := envprofile.PolicyFromConfig(cfg)
	resulting := envprofile.PolicyFromConfig(candidate).Overlays[profile]
	overlays := make(map[string][]envprofile.OverlaySpec, len(policy.Overlays)+1)
	for name, specs := range policy.Overlays {
		overlays[name] = specs
	}
	overlays[profile] = resulting
	policy.Overlays = overlays
	result, err := envprofile.Compose(cfg.Home(), profile, envprofile.ComposeOptions{
		Policy: policy, SurfacingSink: c.stdout, WarningSink: c.stderr,
		PublishDeclaration: func() error {
			_, err := config.WriteRaw(cfg.Path, object)
			return err
		},
	})
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	// A locked overlays_allowed:false empties every overlay list at
	// resolution with the manager §1 warning, so a declaration written here
	// is inert until the policy changes; the lock is rebuilt root-only. The
	// list row already warns; the write row must not stay silent about the
	// same fact.
	if !cfg.Env.OverlaysAllowed {
		_, _ = fmt.Fprintln(c.stderr, "warning: overlays_allowed is false: the declaration is inert; resolution joins the root alone")
	}
	switch {
	case !result.LockRebuilt:
		_, _ = fmt.Fprintf(c.stdout, "%s; the profile has no lock, so only the declaration changed; the first install resolves the overlays jointly\n", summary)
	case result.LockMoved:
		_, _ = fmt.Fprintf(c.stdout, "%s; lock %s\n", summary, result.Info.LockHash)
	default:
		_, _ = fmt.Fprintf(c.stdout, "%s; lock unchanged (%s)\n", summary, result.Info.LockHash)
	}
	return exitOK
}

func (c cli) cmdComposeList(cfg *config.Config, profile string, args []string) int {
	if len(args) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> list takes no arguments")
		return exitUsage
	}
	// A locked overlays_allowed:false empties every overlay list at
	// resolution with the manager §1 warning, so a declaration the file
	// still carries is inert. The list prints the file's contents verbatim
	// (it edits the file, not the effective state) but says so, or the
	// informative row would mislead.
	if !cfg.Env.OverlaysAllowed && len(cfg.Env.Overlays[profile]) > 0 {
		_, _ = fmt.Fprintln(c.stderr, "warning: overlays_allowed is false: the listed declarations are inert; resolution joins the root alone")
	}
	// Sort a copy to keep configuration order intact. Equal source/ref keys
	// still represent separate declarations, with their own directory/weight.
	decls := slices.Clone(cfg.Env.Overlays[profile])
	key := func(decl config.OverlayDeclaration) string {
		return decl.Source + "\x00" + decl.Range + "\x00" + decl.Tag + "\x00" + decl.Revision
	}
	sort.SliceStable(decls, func(i, j int) bool { return key(decls[i]) < key(decls[j]) })
	for _, decl := range decls {
		form := ""
		switch {
		case decl.Range != "":
			form = "range=" + decl.Range
		case decl.Tag != "":
			form = "tag=" + decl.Tag
		case decl.Revision != "":
			form = "revision=" + decl.Revision
		default:
			form = "path"
		}
		weight := "(default)"
		if decl.Weight != nil {
			weight = strconv.Itoa(*decl.Weight)
		}
		directory := "-"
		if decl.Directory != "" {
			directory = decl.Directory
		}
		_, _ = fmt.Fprintf(c.stdout, "%s\t%s\t%s\t%s\n", decl.Source, form, directory, weight)
	}
	return exitOK
}
