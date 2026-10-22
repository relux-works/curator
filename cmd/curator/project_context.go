package main

import (
	"encoding/json"
	"fmt"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/projectcontext"
)

// cmdProjectContext implements only registration and admission controls.
// Provider launch and legacy project add/resolve/refresh keep their own lanes.
func (c cli) cmdProjectContext(verb string, args []string) int {
	flags := c.newFlagSet("project " + verb)
	var name, profile, environment, source, declaration, itemDigest, packageDigest string
	var optional bool
	switch verb {
	case "register":
		flags.StringVar(&name, "name", "", "operator project name (defaults to directory name)")
		flags.StringVar(&profile, "profile", "", "selected installed profile")
		flags.StringVar(&environment, "environment", "", "claude_code, codex_cli or muse")
	case "package":
		flags.StringVar(&source, "source", "", "one inventoried relative AGENTS.md or CLAUDE.md")
		flags.BoolVar(&optional, "optional", false, "explicitly allow this item to be omitted")
	case "preview":
		flags.StringVar(&declaration, "declaration", "", "closed version-1 candidate request; omit to show the retained package")
	case "approve":
		flags.StringVar(&itemDigest, "digest", "", "exact item SHA-256")
		flags.StringVar(&packageDigest, "package-digest", "", "exact review-package SHA-256")
	}
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	count := 1
	if verb == "rebind" || verb == "package" || verb == "approve" || verb == "revoke" {
		count = 2
	}
	if len(positional) != count || (verb == "register" && (profile == "" || environment == "")) || (verb == "package" && source == "") || (verb == "approve" && (itemDigest == "" || packageDigest == "")) {
		_, _ = fmt.Fprintln(c.stderr, "curator: invalid project", verb, "arguments; see --help")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	policy := envprofile.PolicyFromConfig(cfg).Precedence()
	store := projectcontext.Store{Home: cfg.Home(), Precedence: contextlock.Precedence{Winner: policy.Winner, Placement: policy.Placement}}
	var result any
	switch verb {
	case "register":
		result, err = store.Register(name, positional[0], profile, environment)
	case "rebind":
		result, err = store.Rebind(positional[0], positional[1])
	case "inspect":
		result, err = store.Inspect(positional[0])
	case "package":
		result, err = store.PackageInstruction(positional[0], positional[1], source, !optional)
	case "preview":
		if declaration != "" {
			var d projectcontext.Declaration
			d, err = projectcontext.ReadDeclaration(declaration)
			if err == nil {
				_, err = store.Preview(positional[0], d)
			}
		}
		if err == nil {
			result, err = store.ReviewContent(positional[0])
		}
	case "approve":
		result, err = store.Approve(positional[0], positional[1], itemDigest, packageDigest)
	case "revoke":
		err = store.Revoke(positional[0], positional[1])
		result = map[string]string{"item_id": positional[1], "status": "revoked"}
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, string(b))
	return exitOK
}
