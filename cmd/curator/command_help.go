package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/version"
)

// commandRegistry owns the built-in command dispatch and its help topics. Child
// topics describe the verbs dispatched by their parent; external umbrella
// providers own their own help. Help never calls a handler or loads config.
type commandSpec struct {
	name     string
	synopsis string
	summary  string
	detail   string
	run      func(cli, []string) int
}

var commandRegistry = []commandSpec{
	{"bootstrap", "[--skills-root <dir>] [--default-agents <ids>] [--preferred-locale <locale>] [--force] [--if-missing] [--non-interactive]", "Create the machine configuration and default directories.", "", cli.cmdBootstrap},
	{"init", "[path]", "Create Skillfile.json and the managed .gitignore block in the project directory.", "", cli.cmdInit},
	{"add", "<name> [--git <url>] [--branch <ref>] [--tag <ref>] [--revision <ref>] [--source <dir>] [--project <path>]", "Add or replace a skill declaration, then install it.", "", cli.cmdAdd},
	{"remove", "<name> [path] [--project <path>]", "Remove a skill declaration.", "", cli.cmdRemove},
	{"install", "[path] [--all] [--dry-run] [--fix-gitignore] [--strict-tags] [--verbose] [--audit[=advisory|strict]] [--build-ssh-identity <path>] [--build-ssh-agent <socket>] [--build-ssh-known-hosts <path>]", "Apply Skillfile.json to the selected project.", draftWorkflowSection, cli.cmdInstall},
	{"update", "", "Fetch all source repositories under skills_root.", "", func(c cli, _ []string) int { return c.cmdUpdate() }},
	{"upgrade", "[path] [--all] [--dry-run] [--fix-gitignore] [--strict-tags] [--verbose] [--audit[=advisory|strict]] [--build-ssh-identity <path>] [--build-ssh-agent <socket>] [--build-ssh-known-hosts <path>]", "Fetch the selected dependency closure, then install it.", draftWorkflowSection, func(c cli, args []string) int { return c.cmdInstallMode(args, true) }},
	{"status", "[path] [--all] [--check] [--json] [--attest]", "Report manifest, installed, and compiled state.", draftWorkflowSection, cli.cmdStatus},
	{"list", "", "List configured projects and declared skills.", "", func(c cli, _ []string) int { return c.cmdList() }},
	{"project", "<add|resolve|refresh|register|rebind|inspect|package|preview|approve|revoke> [arguments]", "Register projects or resolve their dependency closures.", "", cli.cmdProject},
	{"project add", "<alias> <path> [--agents <ids>]", "Register a project and initialize its declarative files.", "", nil},
	{"project register", "<path> [--name <name>] --profile <profile> --environment <environment>", "Bind a checkout to a new opaque manager identity without modifying its repository.", "Admission only; launch integration is not advertised.\n", nil},
	{"project rebind", "<checkout-id> <path>", "Retire a checkout binding and allocate a fresh identity without carrying approvals.", "", nil},
	{"project inspect", "<checkout-id>", "Report bounded inert project candidates; import no permissions or memory.", "", nil},
	{"project package", "<checkout-id> <item-id> --source <relative-instruction-file> [--optional]", "Explicitly package one inventoried AGENTS.md or CLAUDE.md for review.", "Preserves exact LF Markdown bytes; does not approve or launch.\n", nil},
	{"project preview", "<checkout-id> [--declaration <file>]", "Show the exact itemized review package, snapshots, selected bytes and digest.", "Without --declaration, shows the protected retained package. Never executes, fetches, installs or connects.\n", nil},
	{"project approve", "<checkout-id> <item-id> --digest <item-digest> --package-digest <review-digest>", "Record approval of one exact protected item within manager policy.", "Repository approval files and hashes alone have no authority.\n", nil},
	{"project revoke", "<checkout-id> <item-id>", "Revoke the protected item approval.", "", nil},

	{"project resolve", "[path]", "Resolve transitive dependencies for a project closure.", projectResolveUsage, nil},
	{"project refresh", "[path]", "Refresh the frozen project source lock.", projectResolveUsage, nil},
	{"skill", "check <dir> [--locale <locale>] [--json]", "Validate a skill package.", "", func(c cli, args []string) int {
		if len(args) > 0 && args[0] == "check" {
			return c.cmdSkillCheck(args[1:])
		}
		return c.runProvider(append([]string{"skill"}, args...))
	}},
	{"skill check", "<dir> [--locale <locale>] [--json]", "Validate one skill package.", "", nil},
	{"global", "<init|add|remove|list|status|install|adopt|update|upgrade> [arguments]", "Manage machine-wide skill installations.", globalHelp, cli.cmdGlobal},
	{"global init", "", "Create the global Skillfile.json and managed directories.", "", nil},
	{"global add", "<name> [--git <url>] [--branch <ref>] [--tag <ref>] [--revision <ref>] [--source <dir>]", "Add a global skill declaration and install it.", "", nil},
	{"global remove", "<name>", "Remove a global skill declaration.", "", nil},
	{"global list", "", "List global skill declarations.", "", nil},
	{"global status", "[--check] [--json]", "Report global installed and compiled state.", "", nil},
	{"global install", "[--dry-run] [--audit[=advisory|strict]]", "Install global skill declarations.", "", nil},
	{"global adopt", "<command> [--dry-run]", "Adopt an existing global command.", "", nil},
	{"global update", "", "Fetch global source repositories.", "", nil},
	{"global upgrade", "[--dry-run] [--audit[=advisory|strict]]", "Fetch the global dependency closure and install it.", "", nil},
	{"profile", "<install|import|list|use|update|remove|sync|compose|mcp> [arguments]", "Manage environment profiles.", "Requires the global config: " + configNotFoundHint + ".\n", cli.cmdProfile},
	{"profile install", "<source> [--directory <dir>] [--range <range>|--tag <tag>|--revision <ref>] [--as <name>] [--use] [--takeover] [--confirm-system-delta]", "Install an environment profile from a source package.", "", nil},
	{"profile import", "[--as <name>] [--allow-lossy] [--use]", "Import the managed home as an environment profile.", "", nil},
	{"profile list", "", "List installed environment profiles.", "", nil},
	{"profile use", "<name> [--env <env-id>] [--target <target-id>] [--clear] [--takeover]", "Activate or clear an environment profile.", envAliasUsage + "\n", nil},
	{"profile update", "[<name>] [--all] [--takeover] [--confirm-system-delta]", "Update installed environment profiles.", "", nil},
	{"profile remove", "<name> [--purge]", "Remove an installed environment profile.", "", nil},
	{"profile mcp", "<install|status|repair|restore|replace|remove|release> <profile> [flags]", "Inspect or first-install profile package MCP records.", "Ownership actions refuse until the separate authority capability is available.\n", nil},
	{"profile mcp install", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "First-install absent records or refuse unavailable ownership actions.", "", nil},
	{"profile mcp status", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "Inspect profile MCP records without recovery or writes.", "", nil},
	{"profile mcp repair", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "First-install absent records or refuse unavailable ownership actions.", "", nil},
	{"profile mcp restore", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "Refuse ownership actions until the authority capability is available.", "", nil},
	{"profile mcp replace", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "Refuse ownership actions until the authority capability is available.", "", nil},
	{"profile mcp remove", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "Refuse ownership actions until the authority capability is available.", "", nil},
	{"profile mcp release", "<profile> [--env <id,id>] [--scope profile] [--dry-run] [--takeover]", "Refuse ownership actions until the authority capability is available.", "", nil},
	{"profile sync", "[--takeover]", "Synchronize active environment profiles.", "", nil},
	{"profile compose", "<profile> <add|remove|list> [arguments]", "Compose an environment profile from installed profiles.", "", nil},
	{"profile compose <profile> add", "<source> [--range <range>|--tag <tag>|--revision <ref>] [--directory <dir>] [--weight <n>]", "Add an overlay to a composed profile.", "", nil},
	{"profile compose <profile> remove", "<source>", "Remove an overlay from a composed profile.", "", nil},
	{"profile compose <profile> list", "", "List members of a composed profile.", "", nil},
	{"env", "<resolve|status|config|migrate|unmanage|trust> [arguments]", "Inspect and manage environment homes.", "Requires the global config: " + configNotFoundHint + ".\n", cli.cmdEnv},
	{"env trust", "<key|grant|revoke|list|read> [arguments]", "Record operator policy trust and digest-listed hook consent.", envTrustUsage, nil},
	{"env trust key", "--principal <id> --public-key <base64-ed25519>", "Enroll the operator public key once in private state.", "A different replacement key is refused.\n", nil},
	{"env trust grant", "--command <signed-command.json>", "Record one authenticated policy-grant event.", "Requires the approving operator and actual actor signatures.\n", nil},
	{"env trust revoke", "--command <signed-command.json>", "Revoke every grant derived from the approved entry.", "Revocation never terminates an already running session.\n", nil},
	{"env trust list", "", "Inspect private policy entries and event history without writes.", "", nil},
	{"env trust read", "<env> <absolute-project-root> --checkout <id> --lane <id> --profile <name> --lock-sha256 <hex>", "Read current per-start consent or refuse without usable output.", "Production hook/destination adapters remain unqualified; even empty hooks refuse.\n", nil},
	{"env resolve", "<env-id> [--profile <name>] [--repair [--preserve-lock] [--dry-run]] [--takeover] [--format json|env|shell]", "Resolve an environment fragment; --repair provisions or repairs the managed home.", envAliasUsage + "\n", nil},
	{"env status", "[--check] [--json]", "Report environment home state.", envAliasUsage + "\n", nil},
	{"env config", "<show|set|unset> [<knob> [<value>]]", "Inspect or edit machine environment configuration.", envKnobAliasUsage + "\n", nil},
	{"env config show", "[<knob>]", "Print effective environment configuration.", envKnobAliasUsage + "\n", nil},
	{"env config set", "<knob> <value>", "Set a machine environment configuration knob.", envKnobAliasUsage + "\n", nil},
	{"env config unset", "<knob>", "Unset a machine environment configuration knob.", envKnobAliasUsage + "\n", nil},
	{"env migrate", "--inspect|--plan|--apply [--profile <name>] [--env <env-id>] [--expect <plan-hash>]", "Migrate existing environment installations.", "", nil},
	{"env unmanage", "[--restore-backups] [--env <env-id>] [--target <target-id>]", "Restore files from managed home backups.", envAliasUsage + "\n", nil},
	{"run", "<env-id> [arguments] [-- <passthrough arguments>]", "Dispatch to the installed curator-run provider; arguments after -- are passed unchanged.", envAliasUsage + "\n", func(c cli, args []string) int { return c.runProvider(append([]string{"run"}, args...)) }},
	{"hybrid", "<add|remove|list|status> [arguments]", "Manage hybrid-scope skill declarations.", "", cli.cmdHybrid},
	{"hybrid add", "<name> [--git <url>] [--branch <ref>|--tag <ref>|--revision <ref>] --target <target>|--targets <targets>", "Add a hybrid skill declaration.", "", nil},
	{"hybrid remove", "<name>", "Remove a hybrid skill declaration.", "", nil},
	{"hybrid list", "", "List hybrid skill declarations.", "", nil},
	{"hybrid status", "", "Report hybrid installed state.", "", nil},
	{"audit", "[target] [--all] [--global] [--json] [--allow <hash> --reason <reason>] [--publish <file> --registry <url> --token <token>]", "Audit skills, pin trust, or publish a signed record.", "", cli.cmdAudit},
	{"gc", "", "Remove unreferenced runtime entries.", "", func(c cli, _ []string) int { return c.cmdGC() }},
	{"shell-init", "[auto|zsh|bash|powershell] [--install] [--no-global]", "Print or cache an optional shell hook.", "", cli.cmdShellInit},
	{"hook", "<approve|approvals|revoke> [arguments]", "Manage trust approvals for project environment files.", "", cli.cmdHook},
	{"hook approve", "<path>", "Approve a project environment file.", "", nil},
	{"hook approvals", "", "List recorded project environment file approvals.", "", nil},
	{"hook revoke", "<path>", "Revoke a project environment file approval.", "", nil},
	{"ui", "", "Open the terminal view of installed state.", "", func(c cli, _ []string) int { return c.cmdUI() }},
	{"config", "<show|build-ssh|build-https> [arguments]", "Inspect and edit the machine configuration.", configUsage, cli.cmdConfig},
	{"config show", "", "Print the effective machine configuration.", "", nil},
	{"config build-ssh", "<add|list|remove> [arguments]", "Manage SSH credentials for external build repositories.", buildSSHUsage, nil},
	{"config build-ssh add", "<scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]", "Add or replace SSH credentials for a repository scope.", buildSSHUsage, nil},
	{"config build-ssh list", "", "List configured SSH credential scopes.", buildSSHUsage, nil},
	{"config build-ssh remove", "<scope>", "Remove an SSH credential scope.", buildSSHUsage, nil},
	{"config build-https", "<add|login|list|remove> [arguments]", "Manage HTTPS credentials for external build repositories.", buildHTTPSUsage, nil},
	{"config build-https add", "<scope> [--git-credentials] [--keyring] [--token-env NAME] [--username <name>]", "Configure HTTPS credentials for a repository scope.", buildHTTPSUsage, nil},
	{"config build-https login", "<scope> [--username <name>]", "Store HTTPS credentials interactively for a repository scope.", buildHTTPSUsage, nil},
	{"config build-https list", "", "List configured HTTPS credential scopes.", buildHTTPSUsage, nil},
	{"config build-https remove", "<scope>", "Remove an HTTPS credential scope.", buildHTTPSUsage, nil},
	{"version", "", "Print the curator version.", "", func(c cli, _ []string) int {
		_, _ = fmt.Fprintln(c.stdout, "curator "+version.String())
		return exitOK
	}},
}

func (command commandSpec) helpText() string {
	text := "usage: curator " + command.name
	if command.synopsis != "" {
		text += " " + command.synopsis
	}
	return text + "\n\n" + command.summary + "\n\n" + command.detail
}

// hasHelpFlag stops at the argument boundary, leaving provider operands intact.
func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

// Validate consent before dispatch: argument-free wrappers and providers must
// not silently discard it. Arguments after -- belong to the invoked program.
func rejectsTakeover(args []string) bool {
	var takeover, use, repair, dryRun bool
	for _, arg := range args {
		if arg == "--" {
			break
		}
		name, value, assigned := strings.Cut(arg, "=")
		enabled := !assigned
		if assigned {
			enabled, _ = strconv.ParseBool(value)
		}
		switch name {
		case "--takeover", "-takeover":
			takeover = true
		case "--use", "-use":
			use = enabled
		case "--repair", "-repair":
			repair = enabled
		case "--dry-run", "-dry-run":
			dryRun = enabled
		}
	}
	if !takeover {
		return false
	}
	// Help topics use the same command and activation admission as dispatch.
	if len(args) > 0 && args[0] == "help" {
		args = args[1:]
	}
	command, _ := registeredCommand(args)
	if command != nil {
		switch command.name {
		case "profile mcp", "profile mcp install", "profile mcp status", "profile mcp repair", "profile mcp restore", "profile mcp replace", "profile mcp remove", "profile mcp release", "profile use", "profile sync", "profile update":
			return false
		case "env resolve":
			return !repair || dryRun
		case "profile install":
			return !use
		}
	}
	return true
}

// registeredCommand selects the deepest contiguous command path. Help flags
// may precede a child verb; ordinary operands and -- end command discovery.
func registeredCommand(args []string) (*commandSpec, int) {
	var words []string
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg != "-h" && arg != "--help" {
			words = append(words, arg)
		}
	}
	var selected *commandSpec
	depth := 0
	for i := range commandRegistry {
		command := &commandRegistry[i]
		path := strings.Fields(command.name)
		if len(path) > len(words) || len(path) <= depth {
			continue
		}
		matches := true
		for index, part := range path {
			if part != words[index] && (!strings.HasPrefix(part, "<") || !strings.HasSuffix(part, ">")) {
				matches = false
				break
			}
		}
		if matches {
			selected, depth = command, len(path)
		}
	}
	return selected, depth
}
