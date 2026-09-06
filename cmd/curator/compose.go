package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/config"
)

// isPathSource tells a path operand from a git URL syntactically, never by
// probing the filesystem (environments §9.1): an operand beginning with /,
// ./, or ../ (or a platform absolute-path spelling) is a path declaration;
// every other operand resolves as git under section 1.
func isPathSource(operand string) bool {
	if operand == "" {
		return false
	}
	if strings.HasPrefix(operand, "/") {
		return true
	}
	if strings.HasPrefix(operand, "./") || strings.HasPrefix(operand, `.\\`) {
		return true
	}
	if strings.HasPrefix(operand, "../") || strings.HasPrefix(operand, `..\\`) {
		return true
	}
	if operand == "." || operand == ".." {
		return true
	}
	if len(operand) >= 2 && isDriveLetter(operand[0]) && operand[1] == ':' {
		return true
	}
	if strings.HasPrefix(operand, `\\\\`) {
		return true
	}
	return false
}

func isDriveLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// cmdProfileCompose implements `profile compose <profile> add|remove|list`
// (cli/curator.md): it edits the machine overlays.<profile> list of
// manager-config schema 2. The lock moves only on profile update.
func (c cli) cmdProfileCompose(cfg *config.Config, args []string) int {
	if len(args) < 2 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> add|remove|list [<source> --range|--tag|--revision <ref>] [--weight <n>]")
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
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> add <source> --range|--tag|--revision <ref> [--directory <dir>] [--weight <n>]")
		return exitUsage
	}
	forms := 0
	for _, form := range []string{*rng, *tag, *revision} {
		if form != "" {
			forms++
		}
	}
	// A path overlay is an operator-local directory under the section 1
	// rules: it carries no requirement form and no directory. A git
	// overlay carries exactly one requirement form.
	if isPathSource(positional[0]) {
		if forms != 0 || *directory != "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: profile compose add of a path source takes no --range, --tag, --revision, or --directory")
			return exitUsage
		}
	} else if forms != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose add takes exactly one of --range, --tag, --revision")
		return exitUsage
	}
	entry := map[string]any{"source": positional[0]}
	switch {
	case *rng != "":
		entry["range"] = *rng
	case *tag != "":
		entry["tag"] = *tag
	default:
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
		entry["weight"] = parsed
	}
	object, err := config.ReadRaw(cfg.Path)
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
	overlays[profile] = append(list, entry)
	if schema, _ := object["schema_version"].(float64); schema == float64(config.SchemaVersion) {
		object["schema_version"] = float64(config.SchemaVersion2)
	}
	if _, err := config.WriteRaw(cfg.Path, object); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "added overlay %s to profile %s; the lock moves on profile update\n", positional[0], profile)
	return exitOK
}

func (c cli) cmdComposeRemove(cfg *config.Config, profile string, args []string) int {
	flags := c.newFlagSet("profile compose remove")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> remove <source>")
		return exitUsage
	}
	source := positional[0]
	object, err := config.ReadRaw(cfg.Path)
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
	if _, err := config.WriteRaw(cfg.Path, object); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "removed overlay %s from profile %s; the lock moves on profile update\n", source, profile)
	return exitOK
}

func (c cli) cmdComposeList(cfg *config.Config, profile string, args []string) int {
	if len(args) != 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: profile compose <profile> list takes no arguments")
		return exitUsage
	}
	decls := cfg.Env.Overlays[profile]
	names := make([]string, 0, len(decls))
	byKey := map[string]config.OverlayDeclaration{}
	for _, decl := range decls {
		key := decl.Source + "\x00" + decl.Range + "\x00" + decl.Tag + "\x00" + decl.Revision
		names = append(names, key)
		byKey[key] = decl
	}
	sort.Strings(names)
	for _, key := range names {
		decl := byKey[key]
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
