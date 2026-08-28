// Command curator is the agent environment manager CLI (Spec §15).
package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"

	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/gitignore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/managerlock"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/rustsource"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/shell"
	"github.com/relux-works/curator/internal/skillcheck"
	"github.com/relux-works/curator/internal/transaction"
	"github.com/relux-works/curator/internal/ui"
	"github.com/relux-works/curator/internal/version"
)

// Exit codes: 0 ok, 1 failure or blocked result, 2 usage.
const (
	exitOK    = 0
	exitFail  = 1
	exitUsage = 2
)

const usage = `curator: agent environment manager

Usage:
  curator <command> [arguments]

Commands:
  bootstrap [flags]         create the machine configuration
  init [path]              create Skillfile.json and the managed gitignore block
  add <name> ...           add or replace a skill declaration, then install
  remove <name>            remove a skill declaration
  install [path] [flags]   apply Skillfile.json (see install -h)
  update                   fetch all source repositories under skills_root
  upgrade [path]           fetch the selected dependency closure, then install
  status [path] [flags]    manifest, installed, and compiled state (--check, --json, --attest)
  list                     configured projects and declared skills
  project <subcommand>     add | resolve
  skill check <dir>        validate one skill package (--locale, --json)
  global <subcommand>      init | add | remove | list | status (--check, --json) | install | update | upgrade
  hybrid <subcommand>      add | remove | list | status
  audit [target] [flags]   run audit, pin trust, or publish a signed record
  gc                       remove unreferenced runtime entries
  shell-init [shell]       print or cache an optional hook (auto, zsh, bash, powershell)
  ui                       terminal view over installed state
  config <subcommand>      show | build-ssh | build-https (see config build-ssh -h, config build-https -h)
  --version                print the curator version
`

func main() {
	if handled, code := rustsource.DispatchInternalWorker(os.Args[1:], os.Stdin, os.Stdout); handled {
		os.Exit(code)
	}
	if buildrepo.IsHTTPSBrokerInvocation(os.Args[0]) {
		os.Exit(buildrepo.RunHTTPSCredentialBroker(os.Args[1:], os.Getenv, os.Stdout))
	}
	// The fixed hidden go-v1 build worker is an implementation boundary, not a
	// user-visible command surface. It is dispatched before any other parsing,
	// requires exactly this one manager-owned argument, and is never reachable
	// through a package file, manifest value, environment value, PATH lookup,
	// shell, or user option.
	if len(os.Args) == 2 && os.Args[1] == godriver.WorkerMode {
		os.Exit(godriver.RunWorker(os.Stdin, os.Stdout))
	}
	// Resolve the environment-backed user path once at the process boundary.
	// The command core receives an explicit source so independent invocations
	// never consult CURATOR_CONFIG themselves.
	os.Exit(run(os.Args[1:], fileConfigSource(config.UserPath()), os.Stdout, os.Stderr))
}

// configSource is the injectable configuration seam used by one CLI invocation.
// Path is also used by commands that create config before it can be loaded.
type configSource interface {
	Path() string
	Load(warn func(string)) (*config.Config, error)
}

type fileConfigSource string

func (source fileConfigSource) Path() string { return string(source) }

func (source fileConfigSource) Load(warn func(string)) (*config.Config, error) {
	return config.Load(source.Path(), warn)
}

type cli struct {
	config   configSource
	stdout   io.Writer
	stderr   io.Writer
	userHome func() (string, error)
}

func (c cli) newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(c.stderr)
	return flags
}

// run executes one isolated CLI invocation. Its configuration source and
// output writers are explicit so callers can run concurrent invocations without
// mutating CURATOR_CONFIG, os.Stdout, or os.Stderr.
func run(args []string, source configSource, stdout, stderr io.Writer) int {
	command := cli{config: source, stdout: stdout, stderr: stderr, userHome: os.UserHomeDir}
	return command.run(args)
}

func (c cli) run(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(c.stderr, usage)
		return exitUsage
	}
	switch args[0] {
	case "--version", "version":
		_, _ = fmt.Fprintln(c.stdout, "curator "+version.String())
		return exitOK
	case "init":
		return c.cmdInit(args[1:])
	case "bootstrap":
		return c.cmdBootstrap(args[1:])
	case "add":
		return c.cmdAdd(args[1:])
	case "remove":
		return c.cmdRemove(args[1:])
	case "install":
		return c.cmdInstall(args[1:])
	case "update":
		return c.cmdUpdate()
	case "upgrade":
		return c.cmdInstallMode(args[1:], true)
	case "status":
		return c.cmdStatus(args[1:])
	case "list":
		return c.cmdList()
	case "project":
		return c.cmdProject(args[1:])
	case "skill":
		if len(args) >= 2 && args[1] == "check" {
			return c.cmdSkillCheck(args[2:])
		}
	case "global":
		return c.cmdGlobal(args[1:])
	case "hybrid":
		return c.cmdHybrid(args[1:])
	case "audit":
		return c.cmdAudit(args[1:])
	case "gc":
		return c.cmdGC()
	case "shell-init":
		return c.cmdShellInit(args[1:])
	case "ui":
		return c.cmdUI()
	case "config":
		return c.cmdConfig(args[1:])
	case "-h", "--help", "help":
		_, _ = fmt.Fprint(c.stdout, usage)
		return exitOK
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown command %q\n\n%s", args[0], usage)
	return exitUsage
}

func (c cli) loadConfig() (*config.Config, int) {
	cfg, err := c.config.Load(func(message string) {
		_, _ = fmt.Fprintln(c.stderr, "warning: "+message)
	})
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return nil, exitFail
	}
	return cfg, exitOK
}

func projectRootArg(args []string) string {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			if abs, err := filepath.Abs(arg); err == nil {
				return abs
			}
			return arg
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// parseInterspersed lets commands accept flags before or after positional
// arguments. The standard flag package stops at the first positional token,
// while the informative CLI surface documents forms such as
// `install <path> --dry-run` and `add <name> --tag <ref>` (Spec §15).
func parseInterspersed(flags *flag.FlagSet, args []string) ([]string, error) {
	var flagArgs, positional []string
	for index := 0; index < len(args); index++ {
		arg := args[index] // #nosec G602 -- index is bounded by the loop condition and only advanced after an explicit bounds check
		if arg == "--" {
			for trailing := index + 1; trailing < len(args); trailing++ {
				positional = append(positional, args[trailing])
			}
			break
		}
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}

		flagArgs = append(flagArgs, arg)
		nameValue := strings.TrimLeft(arg, "-")
		name, _, hasEquals := strings.Cut(nameValue, "=")
		definition := flags.Lookup(name)
		if definition == nil || hasEquals {
			continue
		}
		if optional, ok := definition.Value.(interface{ AcceptsOptionalValue(string) bool }); ok &&
			index+1 < len(args) && optional.AcceptsOptionalValue(args[index+1]) {
			index++
			flagArgs[len(flagArgs)-1] = arg + "=" + args[index]
			continue
		}
		boolean, isBoolean := definition.Value.(interface{ IsBoolFlag() bool })
		if isBoolean && boolean.IsBoolFlag() {
			continue
		}
		if index+1 >= len(args) {
			return nil, fmt.Errorf("flag needs an argument: %s", arg)
		}
		index++
		flagArgs = append(flagArgs, args[index])
	}
	if err := flags.Parse(flagArgs); err != nil {
		return nil, err
	}
	return positional, nil
}

type auditModeValue struct {
	value string
}

func (value *auditModeValue) String() string { return value.value }

func (value *auditModeValue) Set(raw string) error {
	if raw == "true" {
		raw = "advisory"
	}
	if raw != "advisory" && raw != "strict" {
		return fmt.Errorf("audit mode must be advisory or strict")
	}
	value.value = raw
	return nil
}

func (*auditModeValue) IsBoolFlag() bool { return true }

func (*auditModeValue) AcceptsOptionalValue(raw string) bool {
	return raw == "advisory" || raw == "strict"
}

func aliasFor(cfg *config.Config, projectRoot string) string {
	for alias, project := range cfg.Projects {
		if project.Path == projectRoot {
			return alias
		}
	}
	return filepath.Base(projectRoot)
}

type projectTarget struct {
	Alias string
	Root  string
}

func selectProjectTargets(cfg *config.Config, positional []string, all bool) ([]projectTarget, error) {
	if all {
		if len(positional) > 0 {
			return nil, fmt.Errorf("--all cannot be combined with a project target")
		}
		aliases := make([]string, 0, len(cfg.Projects))
		for alias := range cfg.Projects {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)
		targets := make([]projectTarget, 0, len(aliases))
		for _, alias := range aliases {
			targets = append(targets, projectTarget{Alias: alias, Root: cfg.Projects[alias].Path})
		}
		if len(targets) == 0 {
			return nil, fmt.Errorf("--all requested but no projects are configured")
		}
		return targets, nil
	}
	if len(positional) > 1 {
		return nil, fmt.Errorf("expected at most one project target")
	}
	if len(positional) == 1 {
		if project, present := cfg.Projects[positional[0]]; present {
			return []projectTarget{{Alias: positional[0], Root: project.Path}}, nil
		}
	}
	root := projectRootArg(positional)
	root = nearestProjectRoot(root)
	return []projectTarget{{Alias: aliasFor(cfg, root), Root: root}}, nil
}

func nearestProjectRoot(start string) string {
	original := projectRootArg([]string{start})
	root := original
	if info, err := os.Stat(root); err == nil && !info.IsDir() {
		root = filepath.Dir(root)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, manifest.Name)); err == nil {
			return root
		}
		parent := filepath.Dir(root)
		if parent == root {
			return original
		}
		root = parent
	}
}

func (c cli) cmdBootstrap(args []string) int {
	flags := c.newFlagSet("bootstrap")
	skillsRoot := flags.String("skills-root", "", "directory containing skill repositories")
	preferredLocale := flags.String("preferred-locale", "", "preferred locale")
	defaultAgents := flags.String("default-agents", "codex_cli", "comma-separated default agents")
	force := flags.Bool("force", false, "overwrite an existing configuration")
	ifMissing := flags.Bool("if-missing", false, "create configuration only when absent")
	nonInteractive := flags.Bool("non-interactive", false, "fail instead of prompting for missing values")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) > 0 {
		return exitUsage
	}
	if *force && *ifMissing {
		_, _ = fmt.Fprintln(c.stderr, "curator: bootstrap --if-missing and --force are mutually exclusive")
		return exitUsage
	}
	path := c.config.Path()
	if *ifMissing {
		if _, statErr := os.Stat(path); statErr == nil {
			_, _ = fmt.Fprintln(c.stdout, "kept existing config:", path)
			return exitOK
		} else if !os.IsNotExist(statErr) {
			_, _ = fmt.Fprintln(c.stderr, "curator:", statErr)
			return exitFail
		}
	}
	if *skillsRoot == "" && !*nonInteractive {
		_, _ = fmt.Fprint(c.stderr, "skills_root: ")
		reader := bufio.NewReader(os.Stdin)
		value, readErr := reader.ReadString('\n')
		if readErr == nil {
			*skillsRoot = strings.TrimSpace(value)
		}
	}
	if *skillsRoot == "" {
		_, _ = fmt.Fprintln(c.stderr, "curator: bootstrap requires --skills-root")
		return exitUsage
	}
	if err := config.Bootstrap(path, *skillsRoot, *preferredLocale, splitNonEmpty(*defaultAgents), *force); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, "wrote", path)
	_, _ = fmt.Fprintln(c.stdout, "shell profile changes are not required: agent skills can invoke project and global command shims directly")
	_, _ = fmt.Fprintln(c.stdout, "optional bare commands for interactive use: curator shell-init --install")
	return exitOK
}

func splitNonEmpty(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (c cli) cmdInit(args []string) int {
	root := projectRootArg(args)
	path, err := manifest.EnsureEmpty(root)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	entries := adapters.RequiredGitignoreEntries(sortedKnownAgents())
	entries = append(entries, "Skillfile.dev.json")
	if err := gitignore.Append(filepath.Join(root, ".gitignore"), entries); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, "initialized", path)
	return exitOK
}

func sortedKnownAgents() []string {
	known := adapters.KnownAgents()
	var agents []string
	for agent := range known {
		agents = append(agents, agent)
	}
	sort.Strings(agents)
	return agents
}

func (c cli) cmdAdd(args []string) int {
	flags := c.newFlagSet("add")
	git := flags.String("git", "", "git clone URL")
	source := flags.String("source", "", "source directory under skills_root")
	tag := flags.String("tag", "", "git tag")
	branch := flags.String("branch", "", "git branch")
	revision := flags.String("revision", "", "git revision")
	project := flags.String("project", "", "project alias or path")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	if len(positional) < 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: add requires a skill name")
		return exitUsage
	}
	name := positional[0]
	refKind, refValue := "", ""
	for kind, value := range map[string]string{"tag": *tag, "branch": *branch, "revision": *revision} {
		if value != "" {
			if refKind != "" {
				_, _ = fmt.Fprintln(c.stderr, "curator: specify exactly one of --tag, --branch, --revision")
				return exitUsage
			}
			refKind, refValue = kind, value
		}
	}
	if refKind == "" {
		_, _ = fmt.Fprintln(c.stderr, "curator: specify exactly one of --tag, --branch, --revision")
		return exitUsage
	}
	rootArgs := positional[1:]
	if *project != "" {
		rootArgs = []string{*project}
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	targets, targetErr := selectProjectTargets(cfg, rootArgs, false)
	if targetErr != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", targetErr)
		return exitUsage
	}
	root := targets[0].Root
	if err := manifest.AddDecl(root, name, refKind, refValue, *git, *source); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return c.cmdInstall([]string{root})
}

func (c cli) cmdRemove(args []string) int {
	flags := c.newFlagSet("remove")
	project := flags.String("project", "", "project alias or path")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) < 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: remove requires a skill name")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	rootArgs := positional[1:]
	if *project != "" {
		rootArgs = []string{*project}
	}
	targets, targetErr := selectProjectTargets(cfg, rootArgs, false)
	if targetErr != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", targetErr)
		return exitUsage
	}
	if err := manifest.RemoveDecl(targets[0].Root, positional[0]); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, "removed", positional[0])
	return exitOK
}

func (c cli) installFlags(args []string) (install.Options, []string, bool, string, error) {
	flags := c.newFlagSet("install")
	all := flags.Bool("all", false, "operate on all configured projects")
	dryRun := flags.Bool("dry-run", false, "plan work without modifying files")
	fixGitignore := flags.Bool("fix-gitignore", false, "append missing managed gitignore entries")
	strictTags := flags.Bool("strict-tags", false, "fail if an installed tag moved to another commit")
	verbose := flags.Bool("verbose", false, "print detailed progress")
	var auditMode auditModeValue
	flags.Var(&auditMode, "audit", "run the audit gate in advisory or strict mode")
	sshIdentity := flags.String("build-ssh-identity", "",
		"identity file for external SSH build repositories (or "+install.EnvBuildSSHIdentity+")")
	sshAgent := flags.String("build-ssh-agent", "",
		"agent socket for external SSH build repositories, or \""+install.BuildSSHAgentAuto+
			"\" for your own agent (or "+install.EnvBuildSSHAgent+")")
	sshKnownHosts := flags.String("build-ssh-known-hosts", "",
		"host keys external SSH build repositories are verified against (or "+install.EnvBuildSSHKnownHosts+")")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return install.Options{}, nil, false, "", err
	}
	return install.Options{
		DryRun: *dryRun, FixGitignore: *fixGitignore,
		StrictTags: *strictTags, Verbose: *verbose,
		BuildSSH: install.BuildSSHFlags{
			Identity: *sshIdentity, Agent: *sshAgent, KnownHosts: *sshKnownHosts,
		},
	}, positional, *all, auditMode.value, nil
}

func (c cli) cmdInstall(args []string) int {
	return c.cmdInstallMode(args, false)
}

func (c cli) cmdInstallMode(args []string, fetch bool) int {
	opts, rest, all, auditMode, err := c.installFlags(args)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	if auditMode != "" {
		cfgCopy := *cfg
		cfgCopy.Audit = cfg.Audit
		cfgCopy.Audit.Enabled = true
		cfgCopy.Audit.Mode = auditMode
		cfg = &cfgCopy
	}
	authority, err := preflightCLIExecution(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	opts.Build.Assurance = authority
	targets, targetErr := selectProjectTargets(cfg, rest, all)
	if targetErr != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", targetErr)
		return exitUsage
	}
	opts.Fetch = fetch && !opts.DryRun
	opts.FetchedRepos = map[string]bool{}
	opts.External = productionExternalDeps(cfg, opts.DryRun)
	opts.External.BuildSSH = install.CaptureBuildSSHSelection(cfg, opts.BuildSSH, os.Getenv)
	opts.External.BuildSSH.Resolve = c.operatorBuildSSHResolver(cfg, opts.DryRun)
	opts.External.BuildHTTPS.Resolve = c.operatorBuildHTTPSResolver(cfg, opts.DryRun)
	exitCode := exitOK
	for _, target := range targets {
		result := install.Project(cfg, target.Root, target.Alias, opts)
		c.printResult(result)
		if !opts.DryRun {
			c.printRepairNotices(result)
		}
		if result.Status == "failed" {
			exitCode = exitFail
		}
	}
	return exitCode
}

func (c cli) printResult(result install.Result) {
	for _, message := range result.Messages {
		_, _ = fmt.Fprintln(c.stdout, message)
	}
	c.printResultErrors(result)
}

// printResultErrors reports the failure surface of a result: the redacted
// failure text plus, when a go-v1 trust boundary refused, the operator guidance
// that belongs to it — which selection mechanisms exist and which release
// families this manager tested.
func (c cli) printResultErrors(result install.Result) {
	c.printFailures(result, "error:")
}

// printStatusRefusal is the read-only reporting form of the same surface.
// `status` uses it when the refusal is already published per command in the
// stable vocabulary: stdout stays one report document, and the same detail is
// a warning on standard error rather than an error, because the command did
// produce the report it was asked for.
func (c cli) printStatusRefusal(result install.Result) {
	c.printFailures(result, "warning:")
}

func (c cli) printFailures(result install.Result, prefix string) {
	for _, message := range result.Errors {
		_, _ = fmt.Fprintln(c.stderr, prefix, message)
	}
	if guidance := goToolchainGuidance(result.BuildDiagnostic); guidance != "" {
		_, _ = fmt.Fprintln(c.stderr, prefix, guidance)
	}
}

func (c cli) cmdUpdate() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	entries, err := os.ReadDir(cfg.SkillsRoot)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	failed := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repo := filepath.Join(cfg.SkillsRoot, entry.Name())
		if _, err := os.Stat(filepath.Join(repo, ".git")); err != nil {
			continue
		}
		if err := gitops.Fetch(repo); err != nil {
			_, _ = fmt.Fprintf(c.stderr, "warning: %s: %v\n", entry.Name(), err)
			failed = true
			continue
		}
		_, _ = fmt.Fprintln(c.stdout, "fetched", entry.Name())
	}
	if failed {
		return exitFail
	}
	return exitOK
}

func (c cli) cmdStatus(args []string) int {
	flags := c.newFlagSet("status")
	all := flags.Bool("all", false, "operate on all configured projects")
	check := flags.Bool("check", false, "exit non-zero unless every skill is up to date")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	attest := flags.Bool("attest", false, "re-check installed skills against trusted registries")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	authority, err := preflightCLIExecution(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	targets, targetErr := selectProjectTargets(cfg, positional, *all)
	if targetErr != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", targetErr)
		return exitUsage
	}
	if *attest {
		exitCode := exitOK
		for _, target := range targets {
			if code := c.cmdStatusAttest(cfg, target.Root, target.Alias, *jsonOut); code != exitOK {
				exitCode = code
			}
		}
		return exitCode
	}

	exitCode := exitOK
	jsonResults := make([]map[string]any, 0, len(targets))
	for _, target := range targets {
		// Installed markers are fingerprinted before the read-only plan and
		// again after classification, so compiled state that moved during the
		// whole window is reported as such instead of silently deciding the
		// verdict from a plan that was already stale.
		scope := projectStatusScope(cfg, target.Root, target.Alias)
		before := markerDigests(scope.stores...)
		result := install.Project(cfg, target.Root, target.Alias, install.Options{DryRun: true, Build: install.BuildDeps{Assurance: authority}, External: productionExternalDeps(cfg, true)})
		if result.Status == "failed" {
			// A read-only plan that refused, yet still described every compiled
			// command it was asked about, has produced a currentness verdict — and
			// reporting a verdict is not itself a failure. The rows below carry it
			// in the stable vocabulary, and `--check` is the surface that turns a
			// non-current verdict into a non-zero exit.
			//
			// Everything else keeps the historical behaviour exactly: a failure
			// that describes no compiled command, or describes only some of them,
			// reports the error and exits non-zero rather than publishing a
			// silently partial report.
			if !result.BuildsComplete {
				c.printResult(result)
				exitCode = exitFail
				continue
			}
			c.printStatusRefusal(result)
		}
		if result.Status == "skipped" {
			c.printResult(result)
			if *check {
				exitCode = exitFail
			}
			continue
		}
		drift, builds := statusReport(cfg, scope, factsList(result.Builds), before)
		payload := map[string]any{"alias": target.Alias, "path": target.Root, "skills": drift}
		// A closure without compiled commands keeps the historical object
		// exactly: no build key appears at all.
		if len(builds) > 0 {
			payload["builds"] = builds
		}
		jsonResults = append(jsonResults, payload)
		if !*jsonOut {
			names := make([]string, 0, len(drift))
			for name := range drift {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				_, _ = fmt.Fprintf(c.stdout, "%s: %s %s\n", target.Alias, name, drift[name])
			}
			for _, build := range builds {
				_, _ = fmt.Fprintf(c.stdout, "%s: %s\n", target.Alias, build.Describe())
			}
		}
		if *check && checkFailed(drift, builds) {
			exitCode = exitFail
		}
	}
	if *jsonOut {
		var output any = jsonResults
		if len(jsonResults) == 1 {
			output = jsonResults[0]
		}
		payload, _ := json.MarshalIndent(output, "", "  ")
		_, _ = fmt.Fprintln(c.stdout, string(payload))
	}
	return exitCode
}

// statusScope is one installed scope seen by the read-only currentness surface:
// the declaration document it compares against, the store its own skills are
// installed into, and every store a command it activates may live in.
//
// It exists so the project scope and the machine-wide scope share one
// classification, one stable vocabulary, and one fail-closed verdict rather than
// growing a second, weaker one.
type statusScope struct {
	alias        string
	manifestRoot string
	skillsDir    string
	// stores is fingerprinted for the classification race window and searched
	// for the installed directory of a compiled command's skill.
	stores []string
}

// projectStatusScope is one project. It also reads the machine-level hybrid
// store, because a hybrid declaration activates against a project and its
// installed node is reachable from here.
func projectStatusScope(cfg *config.Config, projectRoot, alias string) statusScope {
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	return statusScope{
		alias:        alias,
		manifestRoot: projectRoot,
		skillsDir:    skillsDir,
		stores:       []string{skillsDir, scopes.HybridSkillsRoot(cfg.Home())},
	}
}

// globalStatusScope is the machine-wide scope. It consults no hybrid store:
// hybrid declarations activate against a project, never against the global
// scope, and every node the global closure resolves — declared or transitively
// reached — is installed into the global store itself.
func globalStatusScope(cfg *config.Config) statusScope {
	root := install.GlobalRoot(cfg.Home())
	return statusScope{
		alias:        "global",
		manifestRoot: root,
		skillsDir:    filepath.Join(root, "skills"),
		stores:       []string{filepath.Join(root, "skills")},
	}
}

// statusDrift compares declared skills with installed markers.
func statusDrift(cfg *config.Config, projectRoot string) map[string]string {
	return scopeStatusDrift(cfg, projectRoot, filepath.Join(projectRoot, ".agents", "skills"))
}

// statusReport is the complete read-only currentness verdict of one scope: the
// historical per-skill drift map plus one diagnostic row per compiled command
// the closure activates.
//
// Compiled state can only demote a skill that every ordinary check already
// accepted. A skill that is missing, tampered, unresolvable, or behind its
// declaration keeps that more actionable code.
func statusReport(
	cfg *config.Config,
	scope statusScope,
	builds []buildFacts,
	before map[string]string,
) (map[string]string, []buildReport) {
	drift := scopeStatusDrift(cfg, scope.manifestRoot, scope.skillsDir)
	if len(builds) == 0 {
		return drift, nil
	}

	bySkill := map[string][]buildFacts{}
	var skills []string
	for _, facts := range builds {
		if _, seen := bySkill[facts.Skill]; !seen {
			skills = append(skills, facts.Skill)
		}
		bySkill[facts.Skill] = append(bySkill[facts.Skill], facts)
	}
	sort.Strings(skills)

	type verdict struct {
		skill     string
		installed string
		facts     []buildFacts
		state     string
		rows      []buildReport
	}
	verdicts := make([]verdict, 0, len(skills))
	for _, skill := range skills {
		facts := bySkill[skill]
		installed, present := installedSkillDir(scope, skill)
		if !present {
			verdicts = append(verdicts, verdict{skill: skill, installed: installed, facts: facts,
				state: stateNotInstalled,
				rows: plannedRows(facts, stateNotInstalled, "",
					"the compiled command belongs to a skill that is not installed")})
			continue
		}
		state, rows := classifySkillBuilds(installed, marker.Read(installed), facts)
		verdicts = append(verdicts, verdict{skill: skill, installed: installed, facts: facts, state: state, rows: rows})
	}

	// The second look closes the whole window: planning and every classification
	// above have run, so compiled state that differs from the state this verdict
	// was derived from makes the verdict stale, not authoritative. Both halves of
	// that state are re-read — the install markers this run fingerprinted, and
	// the protected cache evidence every compiled row was classified from —
	// because either one can move on its own.
	after := markerDigests(scope.stores...)
	movedCache := recheckBuildCache(cfg.Home(), builds)
	var reports []buildReport
	for _, item := range verdicts {
		if changed := changedDuringCheck(item.facts, before, after, item.installed, movedCache); changed != "" {
			item.state = buildStateChanged
			item.rows = plannedRows(item.facts, buildStateChanged, "", changed)
		}
		reports = append(reports, item.rows...)
		demoteSkill(drift, item.skill, item.state)
	}
	return drift, reports
}

// changedDuringCheck reports why one skill's compiled verdict is not
// authoritative, or an empty string when nothing it was derived from moved.
func changedDuringCheck(
	facts []buildFacts,
	before, after map[string]string,
	installedDir string,
	movedCache map[string]bool,
) string {
	if markerMoved(before, after, installedDir) {
		return "the install marker changed while status was classifying it; re-run status"
	}
	for _, item := range facts {
		if movedCache[item.Skill+"."+item.Command] {
			return "protected build cache state changed while status was classifying it; re-run status"
		}
	}
	return ""
}

// demoteSkill lowers one declared skill from up-to-date to a compiled-state
// code. It never overwrites a code an ordinary check already produced.
func demoteSkill(drift map[string]string, skill, state string) {
	if state == "" || currentCode(state) {
		return
	}
	if recorded, declared := drift[skill]; declared && recorded == stateUpToDate {
		drift[skill] = state
	}
}

// installedSkillDir resolves the store that holds one installed skill: the
// scope's own context store, or another store the scope reads for a node no
// declaration in it reaches.
func installedSkillDir(scope statusScope, skill string) (string, bool) {
	for _, store := range scope.stores {
		candidate := filepath.Join(store, skill)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
	}
	return filepath.Join(scope.skillsDir, skill), false
}

func scopeStatusDrift(cfg *config.Config, manifestRoot, skillsDir string) map[string]string {
	drift := map[string]string{}
	projectManifest, err := manifest.Load(manifestRoot)
	if err != nil || projectManifest == nil {
		return drift
	}
	for _, decl := range projectManifest.Skills {
		installed := filepath.Join(skillsDir, decl.Name)
		if _, err := os.Stat(installed); err != nil {
			drift[decl.Name] = stateNotInstalled
			continue
		}
		recorded := marker.Read(installed)
		if recorded == nil {
			// Marker schema 1 stays readable, so an unchanged schema 1 through 5
			// installation remains current under it. Only a marker the reader
			// refuses reaches here, and the refusal itself carries the stable
			// code: a schema this manager cannot read, a build driver outside the
			// closed set, or an invalid document. A schema-1 marker that has to
			// describe a compiled command is demoted later, by the build
			// classification that can actually see the plan.
			state, _ := markerRefusal(installed)
			drift[decl.Name] = state
			continue
		}
		actualHash, err := hashing.ContentSHA256(installed, nil)
		if err != nil || actualHash != recorded.ContentSHA256 {
			drift[decl.Name] = stateContentDrift
			continue
		}
		repo := filepath.Join(cfg.SkillsRoot, filepath.FromSlash(decl.Source))
		resolved, err := gitops.Resolve(repo, decl.Ref.Kind, decl.Ref.Value)
		if err != nil {
			drift[decl.Name] = stateUnresolvable
			continue
		}
		if recorded.RefKind == decl.Ref.Kind && recorded.Ref == decl.Ref.Value && recorded.Commit == resolved.Commit {
			drift[decl.Name] = stateUpToDate
		} else {
			drift[decl.Name] = stateNeedsInstall
		}
	}
	return drift
}

func (c cli) cmdStatusAttest(cfg *config.Config, projectRoot, alias string, jsonOut bool) int {
	trusted := cfg.TrustedRegistries()
	registries := make([]registry.Registry, 0, len(trusted))
	for _, entry := range trusted {
		registries = append(registries, registry.Registry{Name: entry.Name, URL: entry.URL, PublicKeys: entry.PublicKeys})
	}
	fetch := registry.NewHTTPFetchWithPolicy(
		filepath.Join(cfg.Home(), "cache", "registry"),
		time.Duration(cfg.Audit.CacheTTLSeconds)*time.Second,
		time.Duration(cfg.Audit.OfflineGraceSeconds)*time.Second,
		nil,
	)
	results := registry.AttestRoot(alias, filepath.Join(projectRoot, ".agents", "skills"), registries, fetch)
	if jsonOut {
		payload, _ := json.MarshalIndent(results, "", "  ")
		_, _ = fmt.Fprintln(c.stdout, string(payload))
	} else {
		for _, item := range results {
			suffix := ""
			if item.Registry != "" {
				suffix = " via " + item.Registry
			}
			if item.Detail != "" {
				suffix += " (" + item.Detail + ")"
			}
			_, _ = fmt.Fprintf(c.stdout, "%s: %-24s %s%s\n", item.Scope, item.Skill, item.Result, suffix)
		}
	}
	if registry.HasRevocation(results) {
		return exitFail
	}
	return exitOK
}

func (c cli) cmdList() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	aliases := make([]string, 0, len(cfg.Projects))
	for alias := range cfg.Projects {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		project := cfg.Projects[alias]
		_, _ = fmt.Fprintf(c.stdout, "%s\t%s\n", alias, project.Path)
		if projectManifest, err := manifest.Load(project.Path); err == nil && projectManifest != nil {
			for _, decl := range projectManifest.Skills {
				_, _ = fmt.Fprintf(c.stdout, "  %s %s %s\n", decl.Name, decl.Ref.Kind, decl.Ref.Value)
			}
		}
	}
	return exitOK
}

func (c cli) cmdProject(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: project requires a subcommand: add, resolve")
		return exitUsage
	}
	switch args[0] {
	case "add":
		flags := c.newFlagSet("project add")
		agentsRaw := flags.String("agents", "", "comma-separated target agents")
		positional, err := parseInterspersed(flags, args[1:])
		if err != nil || len(positional) != 2 {
			_, _ = fmt.Fprintln(c.stderr, "curator: project add requires <alias> <path>")
			return exitUsage
		}
		cfg, code := c.loadConfig()
		if code != exitOK {
			return code
		}
		agents := splitNonEmpty(*agentsRaw)
		if len(agents) == 0 {
			agents = cfg.DefaultAgents
		}
		if err := config.AddProject(cfg.Path, positional[0], positional[1], agents); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		root, _ := filepath.Abs(positional[1])
		if _, err := manifest.EnsureEmpty(root); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		entries := append(adapters.RequiredGitignoreEntries(agents), "Skillfile.dev.json")
		if err := gitignore.Append(filepath.Join(root, ".gitignore"), entries); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		_, _ = fmt.Fprintf(c.stdout, "added project %s: %s\n", positional[0], root)
		return exitOK
	case "resolve":
		cfg, code := c.loadConfig()
		if code != exitOK {
			return code
		}
		positional := args[1:]
		if len(positional) == 0 {
			positional = []string{"."}
		}
		targets, err := selectProjectTargets(cfg, positional, false)
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitUsage
		}
		target := targets[0]
		if _, err := os.Stat(filepath.Join(target.Root, manifest.Name)); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator: Skillfile.json not found at or above", target.Root)
			return exitFail
		}
		_, _ = fmt.Fprintf(c.stdout, "alias: %s\npath: %s\nskillfile: %s\n", target.Alias, target.Root, filepath.Join(target.Root, manifest.Name))
		_, _ = fmt.Fprintf(c.stdout, "skills: %s\nbin: %s\n", filepath.Join(target.Root, ".agents", "skills"), filepath.Join(target.Root, ".agents", "bin"))
		return exitOK
	default:
		_, _ = fmt.Fprintf(c.stderr, "curator: unknown project subcommand %q\n", args[0])
		return exitUsage
	}
}

func (c cli) cmdSkillCheck(args []string) int {
	flags := c.newFlagSet("skill check")
	localeValue := flags.String("locale", "", "validate against a locale")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	dir := projectRootArg(positional)
	issues := skillcheck.Validate(dir, *localeValue)
	if *jsonOut {
		payload, _ := json.MarshalIndent(issues, "", "  ")
		_, _ = fmt.Fprintln(c.stdout, string(payload))
	} else {
		for _, issue := range issues {
			_, _ = fmt.Fprintln(c.stdout, skillcheck.Format(issue))
		}
		if len(issues) == 0 {
			_, _ = fmt.Fprintln(c.stdout, dir+": ok")
		}
	}
	if skillcheck.HasErrors(issues) {
		return exitFail
	}
	return exitOK
}

func (c cli) cmdGlobal(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: global requires a subcommand: init, add, remove, list, status, install, update, upgrade")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	switch args[0] {
	case "init":
		path, err := install.GlobalInit(cfg.Home())
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		_, _ = fmt.Fprintln(c.stdout, "initialized", path)
		return exitOK
	case "add":
		flags := c.newFlagSet("global add")
		git := flags.String("git", "", "git clone URL")
		tag := flags.String("tag", "", "git tag")
		revision := flags.String("revision", "", "git revision")
		branch := flags.String("branch", "", "git branch")
		source := flags.String("source", "", "source directory under skills_root")
		positional, err := parseInterspersed(flags, args[1:])
		if err != nil || len(positional) < 1 {
			return exitUsage
		}
		refKind, refValue := pickRef(*tag, *branch, *revision)
		if refKind == "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: specify exactly one of --tag, --branch, --revision")
			return exitUsage
		}
		if err := manifest.AddDecl(install.GlobalRoot(cfg.Home()), positional[0], refKind, refValue, *git, *source); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		return c.runGlobalInstall(cfg, nil)
	case "install":
		return c.runGlobalInstall(cfg, args[1:])
	case "remove":
		if len(args) < 2 {
			return exitUsage
		}
		if err := manifest.RemoveDecl(install.GlobalRoot(cfg.Home()), args[1]); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		return exitOK
	case "list":
		globalManifest, err := manifest.Load(install.GlobalRoot(cfg.Home()))
		if err != nil || globalManifest == nil {
			_, _ = fmt.Fprintln(c.stdout, "no global Skillfile; run 'curator global init'")
			return exitOK
		}
		for _, decl := range globalManifest.Skills {
			_, _ = fmt.Fprintf(c.stdout, "%s %s %s\n", decl.Name, decl.Ref.Kind, decl.Ref.Value)
		}
		return exitOK
	case "status":
		return c.cmdGlobalStatus(cfg, args[1:])
	case "update":
		return c.cmdUpdate()
	case "upgrade":
		return c.runGlobalInstallMode(cfg, args[1:], true)
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown global subcommand %q\n", args[0])
	return exitUsage
}

// cmdGlobalStatus is the read-only currentness surface of the machine-wide
// scope. It carries the same contract as `curator status`: the same stable
// codes, the same optional machine-readable document, and the same fail-closed
// `--check`.
//
// Compiled currentness cannot be read off an install marker alone — the logical
// key is a digest over the whole build input, and only a plan derives the
// current one — so this command runs the same read-only global plan
// `curator global install --dry-run` runs. That resolves the closure and passes
// the read-only audit and registry gates; it runs no compiler and writes no
// installation target, cache entry, or trust state.
//
// The pre-existing surface is preserved exactly. The declared-skill report is
// still derived straight from install markers, never from the plan, so a global
// scope without compiled commands prints the lines it always printed and still
// exits zero — including when the plan itself refused. Reporting and verdict
// stay separate here too: `--check` is the only surface that turns a
// non-current, or an unprovable, verdict into a non-zero exit.
//
// The command is three phases — parse the request, acquire the plan once,
// classify and render from it — so the whole reporting contract can be driven
// from an already acquired plan without a second one being derived for it.
func (c cli) cmdGlobalStatus(cfg *config.Config, args []string) int {
	opts, code := c.parseGlobalStatusOptions(args)
	if code != exitOK {
		return code
	}
	authority, err := preflightCLIExecution(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return c.reportGlobalStatus(cfg, opts, func(cfg *config.Config) (install.Result, bool) {
		return c.globalStatusPlanWithAuthority(cfg, authority)
	})
}

// globalStatusOptions is the parsed request of one `curator global status`
// invocation: which document to render, and whether the verdict is fail-closed.
type globalStatusOptions struct {
	check   bool
	jsonOut bool
}

// parseGlobalStatusOptions is the first phase. The machine-wide scope takes no
// target: it is one scope, so a stray path is a usage error rather than
// something silently ignored.
func (c cli) parseGlobalStatusOptions(args []string) (globalStatusOptions, int) {
	flags := c.newFlagSet("global status")
	check := flags.Bool("check", false, "exit non-zero unless every skill and compiled command is current")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return globalStatusOptions{}, exitUsage
	}
	if len(positional) > 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: global status accepts flags only")
		return globalStatusOptions{}, exitUsage
	}
	return globalStatusOptions{check: *check, jsonOut: *jsonOut}, exitOK
}

// globalStatusAcquire is the second phase: it produces the read-only plan the
// compiled verdict is derived from, and reports whether compiled state stayed
// unprovable. A command run always passes globalStatusPlan, so every invocation
// classifies a plan this run acquired itself.
type globalStatusAcquire func(*config.Config) (install.Result, bool)

// reportGlobalStatus is the third phase: it classifies the acquired plan against
// installed state, renders the requested document, and applies the fail-closed
// verdict. Acquisition is a parameter so the phase can be driven from a plan
// that was already acquired, without the classification, the rendering, or the
// verdict differing in any way from a command run.
func (c cli) reportGlobalStatus(cfg *config.Config, opts globalStatusOptions, acquire globalStatusAcquire) int {
	scope := globalStatusScope(cfg)
	// Installed markers are fingerprinted before the read-only plan and again
	// after classification, so compiled state that moved during the whole window
	// is reported as such instead of published as an authoritative verdict.
	before := markerDigests(scope.stores...)
	result, unprovable := acquire(cfg)
	if result.Status == "failed" {
		// The declared-skill report was still produced, so the refusal is a
		// warning on standard error rather than an error, exactly as it is for a
		// project status that still published every compiled row.
		c.printStatusRefusal(result)
	}

	drift, builds := statusReport(cfg, scope, factsList(result.Builds), before)
	if opts.jsonOut {
		// The machine-wide document carries no `path`: the scope has no
		// operator-supplied root, `alias` already identifies it, and the manager
		// home is never published. A scope without compiled commands produces the
		// declared-skill document with no `builds` key at all.
		payload := map[string]any{"alias": scope.alias, "skills": drift}
		if len(builds) > 0 {
			payload["builds"] = builds
		}
		document, _ := json.MarshalIndent(payload, "", "  ")
		_, _ = fmt.Fprintln(c.stdout, string(document))
	} else {
		names := make([]string, 0, len(drift))
		for name := range drift {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			_, _ = fmt.Fprintf(c.stdout, "%s: %s %s\n", scope.alias, name, drift[name])
		}
		for _, build := range builds {
			_, _ = fmt.Fprintf(c.stdout, "%s: %s\n", scope.alias, build.Describe())
		}
	}
	if opts.check && (unprovable || checkFailed(drift, builds)) {
		return exitFail
	}
	return exitOK
}

// globalStatusPlan runs the read-only machine-wide plan and reports whether
// compiled state stayed unprovable.
//
// Unprovable means the plan refused before it could describe every compiled
// command the closure activates, so this run cannot even tell whether the scope
// has compiled state. It never suppresses the declared-skill report; it is the
// fail-closed input of `--check`. A scope with no machine-wide Skillfile is not
// unprovable: it declares nothing and activates nothing.
func (c cli) globalStatusPlan(cfg *config.Config) (install.Result, bool) {
	return c.globalStatusPlanWithAuthority(cfg, install.NewPortableBuildAuthority())
}

func (c cli) globalStatusPlanWithAuthority(cfg *config.Config, authority *install.BuildAuthority) (install.Result, bool) {
	userHome, err := c.userHome()
	if err != nil {
		result := install.Result{Alias: "global", Path: install.GlobalRoot(cfg.Home()), Status: "failed"}
		result.Errors = append(result.Errors, fmt.Sprintf(
			"could not resolve the user home the machine-wide scope mirrors into: %v", err))
		return result, true
	}
	result := install.Global(cfg, userHome, install.Options{DryRun: true, Build: install.BuildDeps{Assurance: authority}, External: productionExternalDeps(cfg, true)})
	return result, result.Status == "failed" && !result.BuildsComplete
}

func (c cli) runGlobalInstall(cfg *config.Config, args []string) int {
	return c.runGlobalInstallMode(cfg, args, false)
}

func (c cli) runGlobalInstallMode(cfg *config.Config, args []string, fetch bool) int {
	opts, positional, all, auditMode, err := c.installFlags(args)
	if err != nil || len(positional) != 0 || all {
		if err == nil {
			err = fmt.Errorf("global install accepts flags only")
		}
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	if auditMode != "" {
		cfgCopy := *cfg
		cfgCopy.Audit = cfg.Audit
		cfgCopy.Audit.Enabled = true
		cfgCopy.Audit.Mode = auditMode
		cfg = &cfgCopy
	}
	authority, err := preflightCLIExecution(context.Background(), cfg)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	opts.Build.Assurance = authority
	opts.Fetch = fetch && !opts.DryRun
	opts.FetchedRepos = map[string]bool{}
	opts.External = productionExternalDeps(cfg, opts.DryRun)
	opts.External.BuildSSH = install.CaptureBuildSSHSelection(cfg, opts.BuildSSH, os.Getenv)
	opts.External.BuildSSH.Resolve = c.operatorBuildSSHResolver(cfg, opts.DryRun)
	opts.External.BuildHTTPS.Resolve = c.operatorBuildHTTPSResolver(cfg, opts.DryRun)
	userHome, err := c.userHome()
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	result := install.Global(cfg, userHome, opts)
	c.printResult(result)
	if !opts.DryRun {
		c.printRepairNotices(result)
	}
	if result.Status == "failed" {
		return exitFail
	}
	return exitOK
}

// operatorBuildSSHResolver supplies the interactive credential prompt when
// this process actually has an operator in front of it.
//
// A dry run never returns one: it reports what a run would do, and persisting
// a credential the operator was asked for mid-report would make the read-only
// surface write. A non-interactive run never returns one either, which is the
// fail-closed path: the precheck then names every uncovered repository and the
// exact commands that would cover it.
func (c cli) operatorBuildSSHResolver(cfg *config.Config, dryRun bool) install.BuildSSHResolver {
	stderrFile, terminalOutput := c.stderr.(*os.File)
	if dryRun || !terminalOutput || !attachedToTerminal(os.Stdin) || !attachedToTerminal(stderrFile) {
		return nil
	}
	// The prompt speaks on stderr so a caller redirecting stdout still sees
	// the questions it is being asked.
	return install.InteractiveBuildSSHResolver(os.Stdin, c.stderr,
		func(credential config.BuildSSHCredential) error {
			_, err := config.SetBuildSSH(cfg.Path, credential)
			return err
		})
}

// operatorBuildHTTPSResolver offers unmatched HTTPS repositories an explicit
// operator choice. A headless run keeps this nil and therefore continues over
// anonymous HTTPS; a dry run remains read-only even on a terminal.
func (c cli) operatorBuildHTTPSResolver(cfg *config.Config, dryRun bool) install.BuildHTTPSResolver {
	stderrFile, terminalOutput := c.stderr.(*os.File)
	if dryRun || !terminalOutput || !attachedToTerminal(os.Stdin) || !attachedToTerminal(stderrFile) {
		return nil
	}
	access := gitcred.Access{}
	return install.InteractiveBuildHTTPSResolver(os.Stdin, c.stderr,
		func() (string, error) { return readBuildHTTPSToken(os.Stdin, c.stderr) },
		persistPromptedBuildHTTPS(cfg, access))
}

func persistPromptedBuildHTTPS(cfg *config.Config, access gitcred.Access) install.PersistBuildHTTPS {
	return func(ctx context.Context, credential config.BuildHTTPSCredential, secret string) error {
		if credential.Token == config.TokenSourceKeyring {
			host := gitcred.ScopeHost(credential.Scope)
			if err := access.StoreScoped(ctx, credential.Scope, host, secret); err != nil {
				return err
			}
			if _, err := config.SetBuildHTTPS(cfg.Path, credential); err != nil {
				_ = access.DeleteScoped(ctx, credential.Scope, host)
				return err
			}
			return nil
		}
		_, err := config.SetBuildHTTPS(cfg.Path, credential)
		return err
	}
}

// attachedToTerminal reports whether one standard stream is a real terminal.
//
// The character-device test alone is not enough: `< /dev/null` is a character
// device, and treating it as a terminal would make a scripted run block on a
// prompt nobody can answer instead of failing closed.
func attachedToTerminal(file *os.File) bool {
	return term.IsTerminal(file.Fd())
}

// productionExternalDeps binds schema-7 repository work to manager/operator
// state. Discovery happens before package data is consulted; an unavailable
// tool remains a zero-value fail-closed dependency and affects only closures
// that actually activate go-repository-v1.
func productionExternalDeps(cfg *config.Config, dryRun bool) install.ExternalDeps {
	deps := install.ExternalDeps{GitTool: productionGitTool()}
	// Read-only surfaces get the environment and configured scopes; a command
	// that also parses --build-ssh-* flags overrides this with them.
	deps.BuildSSH = install.CaptureBuildSSHSelection(cfg, install.BuildSSHFlags{}, os.Getenv)
	deps.BuildHTTPS = install.CaptureBuildHTTPSSelection(cfg, os.Getenv)
	deps.AuditWarnings = func(_ context.Context, subject buildrepo.AuditSubject) ([]string, error) {
		candidate := audit.Subject{
			Name: subject.Declared.Repository, Source: subject.Declared.Identity,
			Git: subject.Declared.Identity, Commit: subject.Effective.Commit,
			Snapshot: subject.SnapshotRoot, SchemaVersion: 7,
			Capabilities: capabilities.ImplicitNone(),
		}
		var auditWarnings, auditErrs []string
		if dryRun {
			auditWarnings, auditErrs = audit.GateReadOnly(cfg, []audit.Subject{candidate})
		} else {
			auditWarnings, auditErrs = audit.Gate(cfg, []audit.Subject{candidate})
		}
		if len(auditErrs) != 0 {
			return auditWarnings, errors.New(strings.Join(auditErrs, "; "))
		}
		return auditWarnings, nil
	}
	return deps
}

func productionGitTool() buildrepo.GitTool {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return buildrepo.GitTool{}
	}
	gitPath, err = filepath.EvalSymlinks(gitPath)
	if err != nil {
		return buildrepo.GitTool{}
	}
	run := func(args ...string) string {
		command := exec.Command(gitPath, args...) // #nosec G204 -- resolved operator tool, fixed discovery arguments.
		output, runErr := command.Output()
		if runErr != nil || len(output) > 4096 {
			return ""
		}
		return strings.TrimSpace(string(output))
	}
	execPath := run("--exec-path")
	version := run("--version")
	if execPath == "" || version == "" {
		return buildrepo.GitTool{}
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return buildrepo.GitTool{}
	}
	tool := buildrepo.GitTool{Executable: gitPath, ExecPath: execPath, AllowedVersions: []string{version}}
	if executable, executableErr := os.Executable(); executableErr == nil {
		tool.AskPass = admittedOperatorFile(executable)
	}
	tool.SSHWrapper = admittedOperatorFile(os.Getenv("GIT_SSH"))
	return tool
}

func admittedOperatorFile(name string) string {
	if name == "" {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(name)
	if err != nil || !filepath.IsAbs(resolved) {
		return ""
	}
	info, err := os.Lstat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return ""
	}
	return resolved
}

func (c cli) cmdHybrid(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(c.stderr, "curator: hybrid requires a subcommand: add, remove, list")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	switch args[0] {
	case "add":
		flags := c.newFlagSet("hybrid add")
		git := flags.String("git", "", "git clone URL")
		tag := flags.String("tag", "", "git tag")
		revision := flags.String("revision", "", "git revision")
		branch := flags.String("branch", "", "git branch")
		targets := flags.String("targets", "", "comma-separated targets (alias, absolute path, or glob)")
		target := flags.String("target", "", "target alias, absolute path, or glob")
		positional, err := parseInterspersed(flags, args[1:])
		if err != nil || len(positional) < 1 || (*targets == "" && *target == "") {
			_, _ = fmt.Fprintln(c.stderr, "curator: hybrid add requires a name and --target or --targets")
			return exitUsage
		}
		refKind, refValue := pickRef(*tag, *branch, *revision)
		if refKind == "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: specify exactly one of --tag, --branch, --revision")
			return exitUsage
		}
		targetValues := []string{*target}
		if *targets != "" {
			targetValues = strings.Split(*targets, ",")
		}
		if err := scopes.AddHybridDecl(cfg.Home(), positional[0], refKind, refValue, *git, targetValues); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		return exitOK
	case "remove":
		if len(args) < 2 {
			return exitUsage
		}
		if err := scopes.RemoveHybridDecl(cfg.Home(), args[1]); err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		return exitOK
	case "list":
		decls, err := scopes.LoadHybridDecls(cfg.Home())
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		for _, decl := range decls {
			_, _ = fmt.Fprintf(c.stdout, "%s %s %s targets=%s\n", decl.Decl.Name, decl.Decl.Ref.Kind, decl.Decl.Ref.Value, strings.Join(decl.Targets, ","))
		}
		return exitOK
	case "status":
		decls, err := scopes.LoadHybridDecls(cfg.Home())
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		store := scopes.HybridSkillsRoot(cfg.Home())
		for _, entry := range decls {
			state := "not-installed"
			installed := filepath.Join(store, entry.Decl.Name)
			if recorded := marker.Read(installed); recorded != nil {
				actual, hashErr := hashing.ContentSHA256(installed, nil)
				if hashErr != nil || actual != recorded.ContentSHA256 {
					state = "content-drift"
				} else {
					state = "installed"
				}
			}
			_, _ = fmt.Fprintf(c.stdout, "%s %s targets=%s\n", entry.Decl.Name, state, strings.Join(entry.Targets, ","))
		}
		return exitOK
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown hybrid subcommand %q\n", args[0])
	return exitUsage
}

func (c cli) cmdAudit(args []string) int {
	flags := c.newFlagSet("audit")
	all := flags.Bool("all", false, "audit all configured projects and global skills")
	global := flags.Bool("global", false, "audit global skills")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	allow := flags.String("allow", "", "pin trust for a content hash")
	reason := flags.String("reason", "", "reason for --allow")
	publish := flags.String("publish", "", "signed audit record (JSON file) to submit")
	registryURL := flags.String("registry", "", "registry base URL for --publish")
	token := flags.String("token", "", "auditor token for --publish (or CURATOR_REGISTRY_TOKEN)")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	if *allow != "" {
		if *reason == "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: --allow requires --reason")
			return exitUsage
		}
		path, err := audit.Pin(cfg.Home(), *allow, *reason, os.Getenv("USER"))
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		_, _ = fmt.Fprintln(c.stdout, "pinned audit trust:", path)
		return exitOK
	}
	if *publish != "" {
		if *registryURL == "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: --publish requires --registry")
			return exitUsage
		}
		bearer := *token
		if bearer == "" {
			bearer = os.Getenv("CURATOR_REGISTRY_TOKEN")
		}
		if bearer == "" {
			_, _ = fmt.Fprintln(c.stderr, "curator: --publish requires --token or CURATOR_REGISTRY_TOKEN")
			return exitUsage
		}
		payload, err := os.ReadFile(*publish) // #nosec G304 -- operator-supplied record path
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		response, err := registry.Publish(*registryURL, bearer, payload)
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitFail
		}
		_, _ = fmt.Fprintln(c.stdout, response)
		return exitOK
	}

	cfgCopy := *cfg
	cfgCopy.Audit = cfg.Audit
	cfgCopy.Audit.Enabled = true
	cfg = &cfgCopy
	var targets []projectTarget
	if *all {
		if len(positional) > 0 || *global {
			_, _ = fmt.Fprintln(c.stderr, "curator: --all cannot be combined with a target or --global")
			return exitUsage
		}
		targets, err = selectProjectTargets(cfg, nil, true)
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitUsage
		}
		targets = append(targets, projectTarget{Alias: "global", Root: install.GlobalRoot(cfg.Home())})
	} else if *global {
		if len(positional) > 0 {
			_, _ = fmt.Fprintln(c.stderr, "curator: --global cannot be combined with a project target")
			return exitUsage
		}
		targets = []projectTarget{{Alias: "global", Root: install.GlobalRoot(cfg.Home())}}
	} else {
		targets, err = selectProjectTargets(cfg, positional, false)
		if err != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", err)
			return exitUsage
		}
	}

	type auditOutput struct {
		Scope    string   `json:"scope"`
		Warnings []string `json:"warnings"`
		Errors   []string `json:"errors"`
	}
	outputs := make([]auditOutput, 0, len(targets))
	exitCode := exitOK
	for _, target := range targets {
		warnings, errors := auditTarget(cfg, target)
		outputs = append(outputs, auditOutput{Scope: target.Alias, Warnings: warnings, Errors: errors})
		if len(errors) > 0 {
			exitCode = exitFail
		}
		if !*jsonOut {
			for _, warning := range warnings {
				_, _ = fmt.Fprintf(c.stdout, "%s: %s\n", target.Alias, warning)
			}
			for _, message := range errors {
				_, _ = fmt.Fprintf(c.stderr, "%s: %s\n", target.Alias, message)
			}
			if len(warnings) == 0 && len(errors) == 0 {
				_, _ = fmt.Fprintf(c.stdout, "%s: audit clean\n", target.Alias)
			}
		}
	}
	if *jsonOut {
		payload, _ := json.MarshalIndent(outputs, "", "  ")
		_, _ = fmt.Fprintln(c.stdout, string(payload))
	}
	return exitCode
}

func auditTarget(cfg *config.Config, target projectTarget) ([]string, []string) {
	projectManifest, err := manifest.Load(target.Root)
	if err != nil {
		return nil, []string{err.Error()}
	}
	if projectManifest == nil {
		return nil, []string{"Skillfile.json not found"}
	}
	substitutions := map[string]devsub.Substitution{}
	if target.Alias != "global" {
		substitutions, err = devsub.Load(target.Root)
		if err != nil {
			return nil, []string{err.Error()}
		}
	}
	nodes, err := closure.Build(closure.Options{
		SkillsRoot: cfg.SkillsRoot, Home: cfg.Home(), AllowedSources: cfg.AllowedSources,
	}, projectManifest, substitutions)
	if err != nil {
		return nil, []string{err.Error()}
	}
	subjects := make([]audit.Subject, 0, len(nodes))
	for _, node := range nodes {
		subjects = append(subjects, audit.Subject{
			Name: node.Name, Source: node.Decl.Source, Git: node.Decl.Git,
			Commit: node.Resolved.Commit, Snapshot: node.Snapshot,
			SchemaVersion: node.Spec.SchemaVersion, Capabilities: node.Spec.Capabilities,
		})
	}
	return audit.Gate(cfg, subjects)
}

// cmdGC runs maintenance under the exclusive manager-home mutation lock, so it
// serializes with every install, rollback, and recovery. Incomplete
// transactions are recovered first, and their build references are marked, so
// an installation interrupted by a crash keeps the artifacts it will finish
// with.
func (c cli) cmdGC() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	home := cfg.Home()
	manager, err := managerlock.New(home)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	lock, err := manager.AcquireHomeOnly(context.Background(), false)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator: acquire the manager-home lock:", err)
		return exitFail
	}
	result, err := collectUnderLock(home, lock)
	if closeErr := lock.Close(); closeErr != nil && err == nil {
		err = fmt.Errorf("release the manager-home lock: %w", closeErr)
	}
	for _, warning := range result.Warnings {
		_, _ = fmt.Fprintln(c.stderr, "curator: warning:", warning)
	}
	for _, entry := range result.RemovedRuntime {
		_, _ = fmt.Fprintln(c.stdout, "removed runtime", entry)
	}
	for _, key := range result.RemovedBuilds {
		_, _ = fmt.Fprintln(c.stdout, "removed build", key)
	}
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "gc: %d runtime entr%s removed, %d build entr%s removed\n",
		len(result.RemovedRuntime), pluralY(len(result.RemovedRuntime)),
		len(result.RemovedBuilds), pluralY(len(result.RemovedBuilds)))
	return exitOK
}

func collectUnderLock(home string, lock *managerlock.HomeLock) (scopes.MaintenanceResult, error) {
	engine, err := transaction.New(home)
	if err != nil {
		return scopes.MaintenanceResult{}, fmt.Errorf("open the install transaction journal: %w", err)
	}
	if err := engine.Recover(lock); err != nil {
		return scopes.MaintenanceResult{}, fmt.Errorf("recover incomplete install transactions: %w", err)
	}
	journalKeys, err := engine.ReferencedBuildKeys(lock)
	if err != nil {
		return scopes.MaintenanceResult{}, fmt.Errorf("read in-flight build references: %w", err)
	}
	return scopes.Collect(scopes.MaintenanceRequest{Home: home, Lock: lock, JournalKeys: journalKeys})
}

func (c cli) cmdShellInit(args []string) int {
	flags := c.newFlagSet("shell-init")
	noGlobal := flags.Bool("no-global", false, "skip global env sourcing")
	installHook := flags.Bool("install", false, "cache the hook and print its optional profile source command")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) > 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: shell-init accepts at most one shell: auto, zsh, bash, powershell")
		return exitUsage
	}
	shellName := "auto"
	if len(positional) == 1 {
		shellName = positional[0]
	}
	if shellName == "auto" {
		shellName = shell.Detect(nil, "")
	}
	if shellName != "zsh" && shellName != "bash" && shellName != "powershell" {
		_, _ = fmt.Fprintln(c.stderr, "curator: unsupported shell:", shellName)
		return exitUsage
	}
	if *installHook {
		hookPath, installErr := shell.InstallHook(shellName, filepath.Dir(c.config.Path()), !*noGlobal)
		if installErr != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", installErr)
			return exitFail
		}
		source, sourceErr := shell.SourceCommand(shellName, hookPath)
		if sourceErr != nil {
			_, _ = fmt.Fprintln(c.stderr, "curator:", sourceErr)
			return exitFail
		}
		_, _ = fmt.Fprintln(c.stdout, "wrote", hookPath)
		_, _ = fmt.Fprintln(c.stdout, "optional shell profile line:", source)
		return exitOK
	}
	hook, err := shell.Hook(shellName, !*noGlobal)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	_, _ = fmt.Fprint(c.stdout, hook)
	return exitOK
}

const configUsage = `curator config: inspect and edit the machine configuration

Usage:
  curator config show                  print the effective configuration
  curator config build-ssh <subcommand>
                                       operator SSH credentials for external
                                       build repositories (add | list | remove)
  curator config build-https <subcommand>
                                       operator HTTPS tokens for external
                                       build repositories (add | login | list | remove)
`

const buildSSHUsage = `curator config build-ssh: operator SSH credentials for external build repositories

Usage:
  curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
  curator config build-ssh list
  curator config build-ssh remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

An entry names an agent, an identity file, or both:
  --agent                  the agent socket the operator's environment provides
  --agent SOCKET           a named agent socket
  --identity PATH          an identity file offered to the destination
  --agent --identity PATH  an agent pinned to that one identity
  --known-hosts PATH       a known-hosts file for this scope
Every path must be absolute or start with '~/'.

add replaces whatever was recorded under the same scope; remove reports a scope
that is not configured as an error; list prints one line per scope, sorted.

Precedence: command-line flags override CURATOR_BUILD_SSH_* environment values,
which override the scopes configured here. Credentials are operator-owned: no
manifest, descriptor, repository, substitution, or marker can select them.

An install that reaches an SSH build repository nothing covers stops before it
fetches anything. With a terminal attached it lists the agent and ~/.ssh/*.pub
files found on this host and asks which to record and under which scope; with
no terminal, and on a dry run, it prints those same candidates as ready-to-run
add commands and fails with build_repository_ssh_credential_missing.
`

func (c cli) cmdConfig(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(c.stderr, configUsage)
		return exitUsage
	}
	switch args[0] {
	case "show":
		return c.cmdConfigShow()
	case "build-ssh":
		return c.cmdConfigBuildSSH(args[1:])
	case "build-https":
		return c.cmdConfigBuildHTTPS(args[1:])
	case "-h", "--help", "help":
		_, _ = fmt.Fprint(c.stdout, configUsage)
		return exitOK
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown config subcommand %q\n\n%s", args[0], configUsage)
	return exitUsage
}

// buildSSHAgentValue is the --agent selector. Bare, it selects the agent
// socket the operator's environment provides; with a value it names one
// socket. Only the affirmative spelling exists, matching the config grammar,
// where "agent": false would be a second way to write an identity-only entry.
type buildSSHAgentValue struct {
	selected bool
	socket   string
}

func (value *buildSSHAgentValue) String() string {
	if value.socket != "" {
		return value.socket
	}
	if value.selected {
		return "true"
	}
	return ""
}

func (value *buildSSHAgentValue) Set(raw string) error {
	switch raw {
	case "true":
		value.selected = true
		return nil
	case "false":
		return fmt.Errorf("--agent %s", config.BuildSSHAgentRule)
	}
	if !config.ValidBuildSSHPath(raw) {
		return fmt.Errorf("agent socket %s", config.BuildSSHPathRule)
	}
	value.selected, value.socket = true, raw
	return nil
}

func (*buildSSHAgentValue) IsBoolFlag() bool { return true }

// AcceptsOptionalValue claims the next token only when it reads as a socket
// path, so `--agent <scope>` keeps the scope positional.
func (*buildSSHAgentValue) AcceptsOptionalValue(raw string) bool {
	return config.ValidBuildSSHPath(raw)
}

func (c cli) cmdConfigBuildSSH(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(c.stderr, buildSSHUsage)
		return exitUsage
	}
	switch args[0] {
	case "add":
		return c.cmdConfigBuildSSHAdd(args[1:])
	case "list":
		return c.cmdConfigBuildSSHList()
	case "remove":
		return c.cmdConfigBuildSSHRemove(args[1:])
	case "-h", "--help", "help":
		_, _ = fmt.Fprint(c.stdout, buildSSHUsage)
		return exitOK
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown config build-ssh subcommand %q\n\n%s", args[0], buildSSHUsage)
	return exitUsage
}

func (c cli) cmdConfigBuildSSHAdd(args []string) int {
	flags := c.newFlagSet("config build-ssh add")
	agent := &buildSSHAgentValue{}
	flags.Var(agent, "agent", "select an SSH agent, optionally by socket path")
	identity := flags.String("identity", "", "identity file offered to the destination")
	knownHosts := flags.String("known-hosts", "", "known-hosts file for this scope")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	if len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: config build-ssh add requires exactly one <scope>")
		return exitUsage
	}
	credential := config.BuildSSHCredential{
		Scope:       positional[0],
		Agent:       agent.selected,
		AgentSocket: agent.socket,
		Identity:    *identity,
		KnownHosts:  *knownHosts,
	}
	// Checked before the config is read, so a malformed invocation is a usage
	// error rather than a failure attributed to the config file.
	if err := config.ValidateBuildSSH(credential); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	replaced, err := config.SetBuildSSH(cfg.Path, credential)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	verb := "added"
	if replaced {
		verb = "replaced"
	}
	_, _ = fmt.Fprintf(c.stdout, "%s build_ssh scope %s: %s\n", verb, credential.Scope, formatBuildSSH(credential))
	return exitOK
}

func (c cli) cmdConfigBuildSSHList() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	scopes := cfg.BuildSSHScopes()
	if len(scopes) == 0 {
		// On stderr, so a caller parsing the listing sees an empty stdout
		// rather than a line that names no scope.
		_, _ = fmt.Fprintln(c.stderr, "curator: no build_ssh scopes are configured")
		return exitOK
	}
	for _, scope := range scopes {
		_, _ = fmt.Fprintf(c.stdout, "%s\t%s\n", scope, formatBuildSSH(cfg.BuildSSH[scope]))
	}
	return exitOK
}

func (c cli) cmdConfigBuildSSHRemove(args []string) int {
	if len(args) != 1 || strings.HasPrefix(args[0], "-") {
		_, _ = fmt.Fprintln(c.stderr, "curator: config build-ssh remove requires exactly one <scope>")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	if err := config.RemoveBuildSSH(cfg.Path, args[0]); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintf(c.stdout, "removed build_ssh scope %s\n", args[0])
	return exitOK
}

// formatBuildSSH renders one credential in the operator's own spelling, so a
// listed line names the paths the config file carries rather than resolved
// ones.
func formatBuildSSH(credential config.BuildSSHCredential) string {
	var parts []string
	switch {
	case credential.AgentSocket != "":
		parts = append(parts, "agent="+credential.AgentSocket)
	case credential.Agent:
		parts = append(parts, "agent")
	}
	if credential.Identity != "" {
		parts = append(parts, "identity="+credential.Identity)
	}
	if credential.KnownHosts != "" {
		parts = append(parts, "known_hosts="+credential.KnownHosts)
	}
	return strings.Join(parts, " ")
}

const buildHTTPSUsage = `curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

A token is never accepted as a command-line argument, and the config never
stores one as a literal value. add names where the token is read from:
  --git-credentials  the operator's own Git HTTPS credential for the scope's
                      host, read through 'git credential fill'
  --keyring           the token already stored for this scope by login
  --token-env NAME    an environment variable read at process entry
  --username NAME     sent alongside the resolved token; defaults to "token"

login reads a token from a hidden prompt, or one line from stdin when it is
not attached to a terminal, stores it through the operator's own Git
credential machinery under a scope-namespaced entry, and selects it for the
scope the same way 'add --keyring' would.

add and login replace whatever was recorded under the same scope; remove
reports a scope that is not configured as an error, and also deletes the
scope's stored token when it selected one; list prints one line per scope,
sorted, and reports whether the token it names actually resolves right now.

Precedence: CURATOR_BUILD_HTTPS_TOKEN, optionally pinned to one host by
CURATOR_BUILD_HTTPS_HOST, covers every repository on that host ahead of the
scopes configured here; unset, or once that host is exhausted, the longest
matching build_https scope applies; an HTTPS repository nothing covers is not
an error, unlike SSH — it fetches anonymously. Credentials are operator-owned:
no manifest, descriptor, repository, substitution, or marker can select them.

Disclosure warning (Spec core §12.2): CURATOR_BUILD_HTTPS_TOKEN without
CURATOR_BUILD_HTTPS_HOST is not bound to any one repository identity — it is
offered to every private HTTPS build repository host this run reaches. Set
CURATOR_BUILD_HTTPS_HOST to bind it to one host, or prefer a build_https
scope, unless every one of those hosts is meant to receive that token.
`

func (c cli) cmdConfigBuildHTTPS(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(c.stderr, buildHTTPSUsage)
		return exitUsage
	}
	switch args[0] {
	case "add":
		return c.cmdConfigBuildHTTPSAdd(args[1:])
	case "login":
		return c.cmdConfigBuildHTTPSLogin(args[1:])
	case "list":
		return c.cmdConfigBuildHTTPSList()
	case "remove":
		return c.cmdConfigBuildHTTPSRemove(args[1:])
	case "-h", "--help", "help":
		_, _ = fmt.Fprint(c.stdout, buildHTTPSUsage)
		return exitOK
	}
	_, _ = fmt.Fprintf(c.stderr, "curator: unknown config build-https subcommand %q\n\n%s", args[0], buildHTTPSUsage)
	return exitUsage
}

func (c cli) cmdConfigBuildHTTPSAdd(args []string) int {
	flags := c.newFlagSet("config build-https add")
	gitCredentials := flags.Bool("git-credentials", false,
		"use the operator's own Git HTTPS credential for this scope's host")
	keyring := flags.Bool("keyring", false, "use the token already stored for this scope (see login)")
	tokenEnv := flags.String("token-env", "", "environment variable read for the token at process entry")
	username := flags.String("username", "", "username sent alongside the resolved token")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	if len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: config build-https add requires exactly one <scope>")
		return exitUsage
	}
	token, envName, err := pickBuildHTTPSSource(*gitCredentials, *keyring, *tokenEnv)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	credential := config.BuildHTTPSCredential{
		Scope: positional[0], Token: token, TokenEnv: envName, Username: *username,
	}
	// Checked before the config is read, so a malformed invocation is a usage
	// error rather than a failure attributed to the config file.
	if err := config.ValidateBuildHTTPS(credential); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	replaced, err := config.SetBuildHTTPS(cfg.Path, credential)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	verb := "added"
	if replaced {
		verb = "replaced"
	}
	_, _ = fmt.Fprintf(c.stdout, "%s build_https scope %s: %s\n", verb, credential.Scope, formatBuildHTTPS(credential))
	return exitOK
}

// pickBuildHTTPSSource resolves the add flags to exactly one config source, so
// a request naming zero or several is a usage error rather than a silent
// pick among them.
func pickBuildHTTPSSource(gitCredentials, keyring bool, tokenEnv string) (token, env string, err error) {
	count := 0
	if gitCredentials {
		token, count = config.TokenSourceGitCredentials, count+1
	}
	if keyring {
		token, count = config.TokenSourceKeyring, count+1
	}
	if tokenEnv != "" {
		env, count = tokenEnv, count+1
	}
	if count != 1 {
		return "", "", errors.New(
			"config build-https add requires exactly one of --git-credentials, --keyring, or --token-env")
	}
	return token, env, nil
}

func (c cli) cmdConfigBuildHTTPSLogin(args []string) int {
	flags := c.newFlagSet("config build-https login")
	username := flags.String("username", "", "username sent alongside the stored token")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	if len(positional) != 1 {
		_, _ = fmt.Fprintln(c.stderr, "curator: config build-https login requires exactly one <scope>")
		return exitUsage
	}
	scope := positional[0]
	if !config.ValidBuildSSHScope(scope) {
		_, _ = fmt.Fprintln(c.stderr, "curator: scope", scope, config.BuildSSHScopeRule)
		return exitUsage
	}
	// The scope is validated, and the config is loaded, before a token is
	// read: a doomed invocation must not first make the operator type a
	// secret into a hidden prompt for nothing.
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	token, err := readBuildHTTPSToken(os.Stdin, c.stderr)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitUsage
	}
	access := gitcred.Access{}
	if err := access.StoreScoped(context.Background(), scope, gitcred.ScopeHost(scope), token); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	credential := config.BuildHTTPSCredential{Scope: scope, Token: config.TokenSourceKeyring, Username: *username}
	replaced, err := config.SetBuildHTTPS(cfg.Path, credential)
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	verb := "logged in and selected"
	if replaced {
		verb = "replaced the login for"
	}
	_, _ = fmt.Fprintf(c.stdout, "%s build_https scope %s: %s\n", verb, scope, formatBuildHTTPS(credential))
	return exitOK
}

// readBuildHTTPSToken reads a token without ever accepting it as a
// command-line argument, matching core 12.2: a hidden prompt when both ends
// are a real terminal, otherwise one line from stdin, so a scripted login can
// pipe a token in without echoing it to a terminal that is not there.
func readBuildHTTPSToken(in *os.File, errOut io.Writer) (string, error) {
	var raw string
	errFile, terminalOutput := errOut.(*os.File)
	if terminalOutput && attachedToTerminal(in) && attachedToTerminal(errFile) {
		_, _ = fmt.Fprint(errOut, "token: ")
		read, err := term.ReadPassword(in.Fd())
		_, _ = fmt.Fprintln(errOut)
		if err != nil {
			return "", fmt.Errorf("reading token: %w", err)
		}
		raw = string(read)
	} else {
		scanner := bufio.NewScanner(in)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", fmt.Errorf("reading token: %w", err)
			}
		}
		raw = scanner.Text()
	}
	token := strings.TrimRight(raw, "\r\n")
	if token == "" {
		return "", errors.New("token must be a non-empty single line")
	}
	return token, nil
}

func (c cli) cmdConfigBuildHTTPSList() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	scopes := cfg.BuildHTTPSScopes()
	if len(scopes) == 0 {
		// On stderr, so a caller parsing the listing sees an empty stdout
		// rather than a line that names no scope.
		_, _ = fmt.Fprintln(c.stderr, "curator: no build_https scopes are configured")
		return exitOK
	}
	access := gitcred.Access{}
	for _, scope := range scopes {
		credential := cfg.BuildHTTPS[scope]
		present := buildHTTPSPresent(context.Background(), access, credential)
		_, _ = fmt.Fprintf(c.stdout, "%s\t%s present=%t\n", scope, formatBuildHTTPS(credential), present)
	}
	return exitOK
}

// buildHTTPSPresent reports whether the material one credential selection
// names actually resolves right now: the stored token behind a keyring
// selection can have been dropped by the operator's own credential store
// since it was recorded, the same way the operator's own Git credential for a
// git-credentials selection can have been removed, or a token_env variable
// can have gone unset.
func buildHTTPSPresent(ctx context.Context, access gitcred.Access, credential config.BuildHTTPSCredential) bool {
	host := gitcred.ScopeHost(credential.Scope)
	switch credential.Token {
	case config.TokenSourceKeyring:
		_, ok := access.ReadScoped(ctx, credential.Scope, host)
		return ok
	case config.TokenSourceGitCredentials:
		_, ok := access.ReadHost(ctx, host)
		return ok
	default:
		return credential.TokenEnv != "" && os.Getenv(credential.TokenEnv) != ""
	}
}

func (c cli) cmdConfigBuildHTTPSRemove(args []string) int {
	if len(args) != 1 || strings.HasPrefix(args[0], "-") {
		_, _ = fmt.Fprintln(c.stderr, "curator: config build-https remove requires exactly one <scope>")
		return exitUsage
	}
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	credential := cfg.BuildHTTPS[args[0]]
	if err := config.RemoveBuildHTTPS(cfg.Path, args[0]); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	if credential.Token == config.TokenSourceKeyring {
		access := gitcred.Access{}
		if !access.DeleteScoped(context.Background(), args[0], gitcred.ScopeHost(args[0])) {
			_, _ = fmt.Fprintf(c.stderr,
				"curator: warning: could not confirm the stored token for %s was removed\n", args[0])
		}
	}
	_, _ = fmt.Fprintf(c.stdout, "removed build_https scope %s\n", args[0])
	return exitOK
}

// formatBuildHTTPS renders one credential in the operator's own spelling —
// where the token is read from, never the token itself — so a listed line
// names the selection the config file carries rather than a resolved secret.
func formatBuildHTTPS(credential config.BuildHTTPSCredential) string {
	var parts []string
	switch credential.Token {
	case config.TokenSourceGitCredentials:
		parts = append(parts, "source=git-credentials")
	case config.TokenSourceKeyring:
		parts = append(parts, "source=keyring")
	}
	if credential.TokenEnv != "" {
		parts = append(parts, "token_env="+credential.TokenEnv)
	}
	if credential.Username != "" {
		parts = append(parts, "username="+credential.Username)
	}
	return strings.Join(parts, " ")
}

func (c cli) cmdConfigShow() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	_, _ = fmt.Fprintln(c.stdout, string(payload))
	return exitOK
}

func pickRef(tag, branch, revision string) (string, string) {
	kind, value := "", ""
	for k, v := range map[string]string{"tag": tag, "branch": branch, "revision": revision} {
		if v != "" {
			if kind != "" {
				return "", ""
			}
			kind, value = k, v
		}
	}
	return kind, value
}

func pluralY(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

func (c cli) cmdUI() int {
	cfg, code := c.loadConfig()
	if code != exitOK {
		return code
	}
	program := tea.NewProgram(ui.NewModel(ui.LoadState(cfg)), tea.WithOutput(c.stdout))
	if _, err := program.Run(); err != nil {
		_, _ = fmt.Fprintln(c.stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
