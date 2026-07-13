// Command curator is the agent environment manager CLI (Spec §15).
package main

import (
	tea "github.com/charmbracelet/bubbletea"

	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/devsub"
	"github.com/relux-works/curator/internal/gitignore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/shell"
	"github.com/relux-works/curator/internal/skillcheck"
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
  status [path] [flags]    manifest vs installed state (--check, --json, --attest)
  list                     configured projects and declared skills
  project <subcommand>     add | resolve
  skill check <dir>        validate one skill package (--locale, --json)
  global <subcommand>      init | add | remove | list | status | install | update | upgrade
  hybrid <subcommand>      add | remove | list | status
  audit [target] [flags]   run audit, pin trust, or publish a signed record
  gc                       remove unreferenced runtime entries
  shell-init [shell]       print or cache an optional hook (auto, zsh, bash, powershell)
  ui                       terminal view over installed state
  config show              print the effective configuration
  --version                print the curator version
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return exitUsage
	}
	switch args[0] {
	case "--version", "version":
		fmt.Println("curator " + version.String())
		return exitOK
	case "init":
		return cmdInit(args[1:])
	case "bootstrap":
		return cmdBootstrap(args[1:])
	case "add":
		return cmdAdd(args[1:])
	case "remove":
		return cmdRemove(args[1:])
	case "install":
		return cmdInstall(args[1:])
	case "update":
		return cmdUpdate()
	case "upgrade":
		return cmdInstallMode(args[1:], true)
	case "status":
		return cmdStatus(args[1:])
	case "list":
		return cmdList()
	case "project":
		return cmdProject(args[1:])
	case "skill":
		if len(args) >= 2 && args[1] == "check" {
			return cmdSkillCheck(args[2:])
		}
	case "global":
		return cmdGlobal(args[1:])
	case "hybrid":
		return cmdHybrid(args[1:])
	case "audit":
		return cmdAudit(args[1:])
	case "gc":
		return cmdGC()
	case "shell-init":
		return cmdShellInit(args[1:])
	case "ui":
		return cmdUI()
	case "config":
		if len(args) >= 2 && args[1] == "show" {
			return cmdConfigShow()
		}
	case "-h", "--help", "help":
		fmt.Print(usage)
		return exitOK
	}
	fmt.Fprintf(os.Stderr, "curator: unknown command %q\n\n%s", args[0], usage)
	return exitUsage
}

func loadConfig() (*config.Config, int) {
	cfg, err := config.Load("", func(message string) {
		fmt.Fprintln(os.Stderr, "warning: "+message)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
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
		arg := args[index]
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

func cmdBootstrap(args []string) int {
	flags := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
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
		fmt.Fprintln(os.Stderr, "curator: bootstrap --if-missing and --force are mutually exclusive")
		return exitUsage
	}
	path := config.UserPath()
	if *ifMissing {
		if _, statErr := os.Stat(path); statErr == nil {
			fmt.Println("kept existing config:", path)
			return exitOK
		} else if !os.IsNotExist(statErr) {
			fmt.Fprintln(os.Stderr, "curator:", statErr)
			return exitFail
		}
	}
	if *skillsRoot == "" && !*nonInteractive {
		fmt.Fprint(os.Stderr, "skills_root: ")
		reader := bufio.NewReader(os.Stdin)
		value, readErr := reader.ReadString('\n')
		if readErr == nil {
			*skillsRoot = strings.TrimSpace(value)
		}
	}
	if *skillsRoot == "" {
		fmt.Fprintln(os.Stderr, "curator: bootstrap requires --skills-root")
		return exitUsage
	}
	if err := config.Bootstrap(path, *skillsRoot, *preferredLocale, splitNonEmpty(*defaultAgents), *force); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	fmt.Println("wrote", path)
	fmt.Println("shell profile changes are not required: agent skills can invoke project and global command shims directly")
	fmt.Println("optional bare commands for interactive use: curator shell-init --install")
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

func cmdInit(args []string) int {
	root := projectRootArg(args)
	path, err := manifest.EnsureEmpty(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	entries := adapters.RequiredGitignoreEntries(sortedKnownAgents())
	entries = append(entries, "Skillfile.dev.json")
	if err := gitignore.Append(filepath.Join(root, ".gitignore"), entries); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	fmt.Println("initialized", path)
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

func cmdAdd(args []string) int {
	flags := flag.NewFlagSet("add", flag.ContinueOnError)
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
		fmt.Fprintln(os.Stderr, "curator: add requires a skill name")
		return exitUsage
	}
	name := positional[0]
	refKind, refValue := "", ""
	for kind, value := range map[string]string{"tag": *tag, "branch": *branch, "revision": *revision} {
		if value != "" {
			if refKind != "" {
				fmt.Fprintln(os.Stderr, "curator: specify exactly one of --tag, --branch, --revision")
				return exitUsage
			}
			refKind, refValue = kind, value
		}
	}
	if refKind == "" {
		fmt.Fprintln(os.Stderr, "curator: specify exactly one of --tag, --branch, --revision")
		return exitUsage
	}
	rootArgs := positional[1:]
	if *project != "" {
		rootArgs = []string{*project}
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	targets, targetErr := selectProjectTargets(cfg, rootArgs, false)
	if targetErr != nil {
		fmt.Fprintln(os.Stderr, "curator:", targetErr)
		return exitUsage
	}
	root := targets[0].Root
	if err := manifest.AddDecl(root, name, refKind, refValue, *git, *source); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	return cmdInstall([]string{root})
}

func cmdRemove(args []string) int {
	flags := flag.NewFlagSet("remove", flag.ContinueOnError)
	project := flags.String("project", "", "project alias or path")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "curator: remove requires a skill name")
		return exitUsage
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	rootArgs := positional[1:]
	if *project != "" {
		rootArgs = []string{*project}
	}
	targets, targetErr := selectProjectTargets(cfg, rootArgs, false)
	if targetErr != nil {
		fmt.Fprintln(os.Stderr, "curator:", targetErr)
		return exitUsage
	}
	if err := manifest.RemoveDecl(targets[0].Root, positional[0]); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	fmt.Println("removed", positional[0])
	return exitOK
}

func installFlags(args []string) (install.Options, []string, bool, string, error) {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	all := flags.Bool("all", false, "operate on all configured projects")
	dryRun := flags.Bool("dry-run", false, "plan work without modifying files")
	fixGitignore := flags.Bool("fix-gitignore", false, "append missing managed gitignore entries")
	strictTags := flags.Bool("strict-tags", false, "fail if an installed tag moved to another commit")
	verbose := flags.Bool("verbose", false, "print detailed progress")
	var auditMode auditModeValue
	flags.Var(&auditMode, "audit", "run the audit gate in advisory or strict mode")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return install.Options{}, nil, false, "", err
	}
	return install.Options{
		DryRun: *dryRun, FixGitignore: *fixGitignore,
		StrictTags: *strictTags, Verbose: *verbose,
	}, positional, *all, auditMode.value, nil
}

func cmdInstall(args []string) int {
	return cmdInstallMode(args, false)
}

func cmdInstallMode(args []string, fetch bool) int {
	opts, rest, all, auditMode, err := installFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitUsage
	}
	cfg, code := loadConfig()
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
	targets, targetErr := selectProjectTargets(cfg, rest, all)
	if targetErr != nil {
		fmt.Fprintln(os.Stderr, "curator:", targetErr)
		return exitUsage
	}
	opts.Fetch = fetch && !opts.DryRun
	opts.FetchedRepos = map[string]bool{}
	exitCode := exitOK
	for _, target := range targets {
		result := install.Project(cfg, target.Root, target.Alias, opts)
		printResult(result)
		if result.Status == "failed" {
			exitCode = exitFail
		}
	}
	return exitCode
}

func printResult(result install.Result) {
	for _, message := range result.Messages {
		fmt.Println(message)
	}
	for _, message := range result.Errors {
		fmt.Fprintln(os.Stderr, "error:", message)
	}
}

func cmdUpdate() int {
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	entries, err := os.ReadDir(cfg.SkillsRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
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
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", entry.Name(), err)
			failed = true
			continue
		}
		fmt.Println("fetched", entry.Name())
	}
	if failed {
		return exitFail
	}
	return exitOK
}

func cmdStatus(args []string) int {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	all := flags.Bool("all", false, "operate on all configured projects")
	check := flags.Bool("check", false, "exit non-zero unless every skill is up to date")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	attest := flags.Bool("attest", false, "re-check installed skills against trusted registries")
	positional, err := parseInterspersed(flags, args)
	if err != nil {
		return exitUsage
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	targets, targetErr := selectProjectTargets(cfg, positional, *all)
	if targetErr != nil {
		fmt.Fprintln(os.Stderr, "curator:", targetErr)
		return exitUsage
	}
	if *attest {
		exitCode := exitOK
		for _, target := range targets {
			if code := cmdStatusAttest(cfg, target.Root, target.Alias, *jsonOut); code != exitOK {
				exitCode = code
			}
		}
		return exitCode
	}

	exitCode := exitOK
	jsonResults := make([]map[string]any, 0, len(targets))
	for _, target := range targets {
		result := install.Project(cfg, target.Root, target.Alias, install.Options{DryRun: true})
		if result.Status == "failed" {
			printResult(result)
			exitCode = exitFail
			continue
		}
		if result.Status == "skipped" {
			printResult(result)
			if *check {
				exitCode = exitFail
			}
			continue
		}
		drift := statusDrift(cfg, target.Root)
		jsonResults = append(jsonResults, map[string]any{"alias": target.Alias, "path": target.Root, "skills": drift})
		if !*jsonOut {
			names := make([]string, 0, len(drift))
			for name := range drift {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				fmt.Printf("%s: %s %s\n", target.Alias, name, drift[name])
			}
		}
		if *check {
			for _, state := range drift {
				if state != "up-to-date" {
					exitCode = exitFail
				}
			}
		}
	}
	if *jsonOut {
		var output any = jsonResults
		if len(jsonResults) == 1 {
			output = jsonResults[0]
		}
		payload, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(payload))
	}
	return exitCode
}

// statusDrift compares declared skills with installed markers.
func statusDrift(cfg *config.Config, projectRoot string) map[string]string {
	return scopeStatusDrift(cfg, projectRoot, filepath.Join(projectRoot, ".agents", "skills"))
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
			drift[decl.Name] = "not-installed"
			continue
		}
		recorded := marker.Read(installed)
		if recorded == nil {
			drift[decl.Name] = "invalid-marker"
			continue
		}
		if recorded.SchemaVersion != marker.SchemaVersion {
			drift[decl.Name] = "unsupported-marker"
			continue
		}
		actualHash, err := hashing.ContentSHA256(installed, nil)
		if err != nil || actualHash != recorded.ContentSHA256 {
			drift[decl.Name] = "content-drift"
			continue
		}
		repo := filepath.Join(cfg.SkillsRoot, filepath.FromSlash(decl.Source))
		resolved, err := gitops.Resolve(repo, decl.Ref.Kind, decl.Ref.Value)
		if err != nil {
			drift[decl.Name] = "unresolvable"
			continue
		}
		if recorded.RefKind == decl.Ref.Kind && recorded.Ref == decl.Ref.Value && recorded.Commit == resolved.Commit {
			drift[decl.Name] = "up-to-date"
		} else {
			drift[decl.Name] = "needs-install"
		}
	}
	return drift
}

func cmdStatusAttest(cfg *config.Config, projectRoot, alias string, jsonOut bool) int {
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
		fmt.Println(string(payload))
	} else {
		for _, item := range results {
			suffix := ""
			if item.Registry != "" {
				suffix = " via " + item.Registry
			}
			if item.Detail != "" {
				suffix += " (" + item.Detail + ")"
			}
			fmt.Printf("%s: %-24s %s%s\n", item.Scope, item.Skill, item.Result, suffix)
		}
	}
	if registry.HasRevocation(results) {
		return exitFail
	}
	return exitOK
}

func cmdList() int {
	cfg, code := loadConfig()
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
		fmt.Printf("%s\t%s\n", alias, project.Path)
		if projectManifest, err := manifest.Load(project.Path); err == nil && projectManifest != nil {
			for _, decl := range projectManifest.Skills {
				fmt.Printf("  %s %s %s\n", decl.Name, decl.Ref.Kind, decl.Ref.Value)
			}
		}
	}
	return exitOK
}

func cmdProject(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "curator: project requires a subcommand: add, resolve")
		return exitUsage
	}
	switch args[0] {
	case "add":
		flags := flag.NewFlagSet("project add", flag.ContinueOnError)
		agentsRaw := flags.String("agents", "", "comma-separated target agents")
		positional, err := parseInterspersed(flags, args[1:])
		if err != nil || len(positional) != 2 {
			fmt.Fprintln(os.Stderr, "curator: project add requires <alias> <path>")
			return exitUsage
		}
		cfg, code := loadConfig()
		if code != exitOK {
			return code
		}
		agents := splitNonEmpty(*agentsRaw)
		if len(agents) == 0 {
			agents = cfg.DefaultAgents
		}
		if err := config.AddProject(cfg.Path, positional[0], positional[1], agents); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		root, _ := filepath.Abs(positional[1])
		if _, err := manifest.EnsureEmpty(root); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		entries := append(adapters.RequiredGitignoreEntries(agents), "Skillfile.dev.json")
		if err := gitignore.Append(filepath.Join(root, ".gitignore"), entries); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		fmt.Printf("added project %s: %s\n", positional[0], root)
		return exitOK
	case "resolve":
		cfg, code := loadConfig()
		if code != exitOK {
			return code
		}
		positional := args[1:]
		if len(positional) == 0 {
			positional = []string{"."}
		}
		targets, err := selectProjectTargets(cfg, positional, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitUsage
		}
		target := targets[0]
		if _, err := os.Stat(filepath.Join(target.Root, manifest.Name)); err != nil {
			fmt.Fprintln(os.Stderr, "curator: Skillfile.json not found at or above", target.Root)
			return exitFail
		}
		fmt.Printf("alias: %s\npath: %s\nskillfile: %s\n", target.Alias, target.Root, filepath.Join(target.Root, manifest.Name))
		fmt.Printf("skills: %s\nbin: %s\n", filepath.Join(target.Root, ".agents", "skills"), filepath.Join(target.Root, ".agents", "bin"))
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "curator: unknown project subcommand %q\n", args[0])
		return exitUsage
	}
}

func cmdSkillCheck(args []string) int {
	flags := flag.NewFlagSet("skill check", flag.ContinueOnError)
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
		fmt.Println(string(payload))
	} else {
		for _, issue := range issues {
			fmt.Println(skillcheck.Format(issue))
		}
		if len(issues) == 0 {
			fmt.Println(dir + ": ok")
		}
	}
	if skillcheck.HasErrors(issues) {
		return exitFail
	}
	return exitOK
}

func cmdGlobal(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "curator: global requires a subcommand: init, add, remove, list, status, install, update, upgrade")
		return exitUsage
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	switch args[0] {
	case "init":
		path, err := install.GlobalInit(cfg.Home())
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		fmt.Println("initialized", path)
		return exitOK
	case "add":
		flags := flag.NewFlagSet("global add", flag.ContinueOnError)
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
			fmt.Fprintln(os.Stderr, "curator: specify exactly one of --tag, --branch, --revision")
			return exitUsage
		}
		if err := manifest.AddDecl(install.GlobalRoot(cfg.Home()), positional[0], refKind, refValue, *git, *source); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		return runGlobalInstall(cfg, nil)
	case "install":
		return runGlobalInstall(cfg, args[1:])
	case "remove":
		if len(args) < 2 {
			return exitUsage
		}
		if err := manifest.RemoveDecl(install.GlobalRoot(cfg.Home()), args[1]); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		return exitOK
	case "list":
		globalManifest, err := manifest.Load(install.GlobalRoot(cfg.Home()))
		if err != nil || globalManifest == nil {
			fmt.Println("no global Skillfile; run 'curator global init'")
			return exitOK
		}
		for _, decl := range globalManifest.Skills {
			fmt.Printf("%s %s %s\n", decl.Name, decl.Ref.Kind, decl.Ref.Value)
		}
		return exitOK
	case "status":
		drift := scopeStatusDrift(cfg, install.GlobalRoot(cfg.Home()), filepath.Join(install.GlobalRoot(cfg.Home()), "skills"))
		names := make([]string, 0, len(drift))
		for name := range drift {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Printf("global: %s %s\n", name, drift[name])
		}
		return exitOK
	case "update":
		return cmdUpdate()
	case "upgrade":
		return runGlobalInstallMode(cfg, args[1:], true)
	}
	fmt.Fprintf(os.Stderr, "curator: unknown global subcommand %q\n", args[0])
	return exitUsage
}

func runGlobalInstall(cfg *config.Config, args []string) int {
	return runGlobalInstallMode(cfg, args, false)
}

func runGlobalInstallMode(cfg *config.Config, args []string, fetch bool) int {
	opts, positional, all, auditMode, err := installFlags(args)
	if err != nil || len(positional) != 0 || all {
		if err == nil {
			err = fmt.Errorf("global install accepts flags only")
		}
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitUsage
	}
	if auditMode != "" {
		cfgCopy := *cfg
		cfgCopy.Audit = cfg.Audit
		cfgCopy.Audit.Enabled = true
		cfgCopy.Audit.Mode = auditMode
		cfg = &cfgCopy
	}
	opts.Fetch = fetch && !opts.DryRun
	opts.FetchedRepos = map[string]bool{}
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	result := install.Global(cfg, userHome, opts)
	printResult(result)
	if result.Status == "failed" {
		return exitFail
	}
	return exitOK
}

func cmdHybrid(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "curator: hybrid requires a subcommand: add, remove, list")
		return exitUsage
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	switch args[0] {
	case "add":
		flags := flag.NewFlagSet("hybrid add", flag.ContinueOnError)
		git := flags.String("git", "", "git clone URL")
		tag := flags.String("tag", "", "git tag")
		revision := flags.String("revision", "", "git revision")
		branch := flags.String("branch", "", "git branch")
		targets := flags.String("targets", "", "comma-separated targets (alias, absolute path, or glob)")
		target := flags.String("target", "", "target alias, absolute path, or glob")
		positional, err := parseInterspersed(flags, args[1:])
		if err != nil || len(positional) < 1 || (*targets == "" && *target == "") {
			fmt.Fprintln(os.Stderr, "curator: hybrid add requires a name and --target or --targets")
			return exitUsage
		}
		refKind, refValue := pickRef(*tag, *branch, *revision)
		if refKind == "" {
			fmt.Fprintln(os.Stderr, "curator: specify exactly one of --tag, --branch, --revision")
			return exitUsage
		}
		targetValues := []string{*target}
		if *targets != "" {
			targetValues = strings.Split(*targets, ",")
		}
		if err := scopes.AddHybridDecl(cfg.Home(), positional[0], refKind, refValue, *git, targetValues); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		return exitOK
	case "remove":
		if len(args) < 2 {
			return exitUsage
		}
		if err := scopes.RemoveHybridDecl(cfg.Home(), args[1]); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		return exitOK
	case "list":
		decls, err := scopes.LoadHybridDecls(cfg.Home())
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		for _, decl := range decls {
			fmt.Printf("%s %s %s targets=%s\n", decl.Decl.Name, decl.Decl.Ref.Kind, decl.Decl.Ref.Value, strings.Join(decl.Targets, ","))
		}
		return exitOK
	case "status":
		decls, err := scopes.LoadHybridDecls(cfg.Home())
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
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
			fmt.Printf("%s %s targets=%s\n", entry.Decl.Name, state, strings.Join(entry.Targets, ","))
		}
		return exitOK
	}
	fmt.Fprintf(os.Stderr, "curator: unknown hybrid subcommand %q\n", args[0])
	return exitUsage
}

func cmdAudit(args []string) int {
	flags := flag.NewFlagSet("audit", flag.ContinueOnError)
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
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	if *allow != "" {
		if *reason == "" {
			fmt.Fprintln(os.Stderr, "curator: --allow requires --reason")
			return exitUsage
		}
		path, err := audit.Pin(cfg.Home(), *allow, *reason, os.Getenv("USER"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		fmt.Println("pinned audit trust:", path)
		return exitOK
	}
	if *publish != "" {
		if *registryURL == "" {
			fmt.Fprintln(os.Stderr, "curator: --publish requires --registry")
			return exitUsage
		}
		bearer := *token
		if bearer == "" {
			bearer = os.Getenv("CURATOR_REGISTRY_TOKEN")
		}
		if bearer == "" {
			fmt.Fprintln(os.Stderr, "curator: --publish requires --token or CURATOR_REGISTRY_TOKEN")
			return exitUsage
		}
		payload, err := os.ReadFile(*publish) // #nosec G304 -- operator-supplied record path
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		response, err := registry.Publish(*registryURL, bearer, payload)
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		fmt.Println(response)
		return exitOK
	}

	cfgCopy := *cfg
	cfgCopy.Audit = cfg.Audit
	cfgCopy.Audit.Enabled = true
	cfg = &cfgCopy
	var targets []projectTarget
	if *all {
		if len(positional) > 0 || *global {
			fmt.Fprintln(os.Stderr, "curator: --all cannot be combined with a target or --global")
			return exitUsage
		}
		targets, err = selectProjectTargets(cfg, nil, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitUsage
		}
		targets = append(targets, projectTarget{Alias: "global", Root: install.GlobalRoot(cfg.Home())})
	} else if *global {
		if len(positional) > 0 {
			fmt.Fprintln(os.Stderr, "curator: --global cannot be combined with a project target")
			return exitUsage
		}
		targets = []projectTarget{{Alias: "global", Root: install.GlobalRoot(cfg.Home())}}
	} else {
		targets, err = selectProjectTargets(cfg, positional, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
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
				fmt.Printf("%s: %s\n", target.Alias, warning)
			}
			for _, message := range errors {
				fmt.Fprintf(os.Stderr, "%s: %s\n", target.Alias, message)
			}
			if len(warnings) == 0 && len(errors) == 0 {
				fmt.Printf("%s: audit clean\n", target.Alias)
			}
		}
	}
	if *jsonOut {
		payload, _ := json.MarshalIndent(outputs, "", "  ")
		fmt.Println(string(payload))
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

func cmdGC() int {
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	removed, err := scopes.CollectRuntime(cfg.Home())
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	for _, entry := range removed {
		fmt.Println("removed runtime", entry)
	}
	fmt.Printf("gc: %d runtime entr%s removed\n", len(removed), pluralY(len(removed)))
	return exitOK
}

func cmdShellInit(args []string) int {
	flags := flag.NewFlagSet("shell-init", flag.ContinueOnError)
	noGlobal := flags.Bool("no-global", false, "skip global env sourcing")
	installHook := flags.Bool("install", false, "cache the hook and print its optional profile source command")
	positional, err := parseInterspersed(flags, args)
	if err != nil || len(positional) > 1 {
		fmt.Fprintln(os.Stderr, "curator: shell-init accepts at most one shell: auto, zsh, bash, powershell")
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
		fmt.Fprintln(os.Stderr, "curator: unsupported shell:", shellName)
		return exitUsage
	}
	if *installHook {
		hookPath, installErr := shell.InstallHook(shellName, filepath.Dir(config.UserPath()), !*noGlobal)
		if installErr != nil {
			fmt.Fprintln(os.Stderr, "curator:", installErr)
			return exitFail
		}
		source, sourceErr := shell.SourceCommand(shellName, hookPath)
		if sourceErr != nil {
			fmt.Fprintln(os.Stderr, "curator:", sourceErr)
			return exitFail
		}
		fmt.Println("wrote", hookPath)
		fmt.Println("optional shell profile line:", source)
		return exitOK
	}
	hook, err := shell.Hook(shellName, !*noGlobal)
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitUsage
	}
	fmt.Print(hook)
	return exitOK
}

func cmdConfigShow() int {
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	fmt.Println(string(payload))
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

func cmdUI() int {
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	program := tea.NewProgram(ui.NewModel(ui.LoadState(cfg)))
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	return exitOK
}
