// Command curator is the agent environment manager CLI (Spec §15).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitignore"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/shell"
	"github.com/relux-works/curator/internal/skillcheck"
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
  init [path]              create Skillfile.json and the managed gitignore block
  add <name> ...           add or replace a skill declaration, then install
  remove <name>            remove a skill declaration
  install [path] [flags]   apply Skillfile.json (see install -h)
  update                   fetch all source repositories under skills_root
  upgrade [path]           update, then install
  status [path] [flags]    manifest vs installed state (--check, --json, --attest)
  list                     configured projects and declared skills
  skill check <dir>        validate one skill package (--locale, --json)
  global <subcommand>      init | add | remove | list | install
  hybrid <subcommand>      add | remove | list
  audit ...                --allow <hash> --reason <text> | --publish <record> --registry <url>
  gc                       remove unreferenced runtime entries
  shell-init <shell>       print the shell hook (zsh, bash, powershell; --no-global)
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
	case "add":
		return cmdAdd(args[1:])
	case "remove":
		return cmdRemove(args[1:])
	case "install":
		return cmdInstall(args[1:])
	case "update":
		return cmdUpdate()
	case "upgrade":
		if code := cmdUpdate(); code != exitOK {
			return code
		}
		return cmdInstall(args[1:])
	case "status":
		return cmdStatus(args[1:])
	case "list":
		return cmdList()
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

func aliasFor(cfg *config.Config, projectRoot string) string {
	for alias, project := range cfg.Projects {
		if project.Path == projectRoot {
			return alias
		}
	}
	return filepath.Base(projectRoot)
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
	return agents
}

func cmdAdd(args []string) int {
	flags := flag.NewFlagSet("add", flag.ContinueOnError)
	git := flags.String("git", "", "git clone URL")
	source := flags.String("source", "", "source directory under skills_root")
	tag := flags.String("tag", "", "git tag")
	branch := flags.String("branch", "", "git branch")
	revision := flags.String("revision", "", "git revision")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	if flags.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "curator: add requires a skill name")
		return exitUsage
	}
	name := flags.Arg(0)
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
	root := projectRootArg(flags.Args()[1:])
	if err := manifest.AddDecl(root, name, refKind, refValue, *git, *source); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	return cmdInstall([]string{root})
}

func cmdRemove(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "curator: remove requires a skill name")
		return exitUsage
	}
	root := projectRootArg(args[1:])
	if err := manifest.RemoveDecl(root, args[0]); err != nil {
		fmt.Fprintln(os.Stderr, "curator:", err)
		return exitFail
	}
	fmt.Println("removed", args[0])
	return exitOK
}

func installFlags(args []string) (install.Options, []string, error) {
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	dryRun := flags.Bool("dry-run", false, "plan work without modifying files")
	fixGitignore := flags.Bool("fix-gitignore", false, "append missing managed gitignore entries")
	strictTags := flags.Bool("strict-tags", false, "fail if an installed tag moved to another commit")
	verbose := flags.Bool("verbose", false, "print detailed progress")
	if err := flags.Parse(args); err != nil {
		return install.Options{}, nil, err
	}
	return install.Options{
		DryRun: *dryRun, FixGitignore: *fixGitignore,
		StrictTags: *strictTags, Verbose: *verbose,
	}, flags.Args(), nil
}

func cmdInstall(args []string) int {
	opts, rest, err := installFlags(args)
	if err != nil {
		return exitUsage
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	root := projectRootArg(rest)
	result := install.Project(cfg, root, aliasFor(cfg, root), opts)
	printResult(result)
	if result.Status == "failed" {
		return exitFail
	}
	return exitOK
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
	check := flags.Bool("check", false, "exit non-zero unless every skill is up to date")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	attest := flags.Bool("attest", false, "re-check installed skills against trusted registries")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	cfg, code := loadConfig()
	if code != exitOK {
		return code
	}
	root := projectRootArg(flags.Args())
	alias := aliasFor(cfg, root)

	if *attest {
		return cmdStatusAttest(cfg, root, alias, *jsonOut)
	}

	result := install.Project(cfg, root, alias, install.Options{DryRun: true})
	if result.Status == "failed" {
		printResult(result)
		return exitFail
	}
	drift := statusDrift(cfg, root)
	if *jsonOut {
		payload, _ := json.MarshalIndent(map[string]any{"alias": alias, "skills": drift}, "", "  ")
		fmt.Println(string(payload))
	} else {
		for name, state := range drift {
			fmt.Printf("%s: %s %s\n", alias, name, state)
		}
	}
	if *check {
		for _, state := range drift {
			if state != "up-to-date" {
				return exitFail
			}
		}
	}
	return exitOK
}

// statusDrift compares declared skills with installed markers.
func statusDrift(cfg *config.Config, projectRoot string) map[string]string {
	drift := map[string]string{}
	projectManifest, err := manifest.Load(projectRoot)
	if err != nil || projectManifest == nil {
		return drift
	}
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	for _, decl := range projectManifest.Skills {
		installed := filepath.Join(skillsDir, decl.Name)
		if _, err := os.Stat(installed); err != nil {
			drift[decl.Name] = "not-installed"
			continue
		}
		repo := filepath.Join(cfg.SkillsRoot, filepath.FromSlash(decl.Source))
		resolved, err := gitops.Resolve(repo, decl.Ref.Kind, decl.Ref.Value)
		if err != nil {
			drift[decl.Name] = "unresolvable"
			continue
		}
		recorded := readMarkerCommit(installed)
		if recorded == resolved.Commit {
			drift[decl.Name] = "up-to-date"
		} else {
			drift[decl.Name] = "needs-install"
		}
	}
	return drift
}

func readMarkerCommit(installedDir string) string {
	payload, err := os.ReadFile(filepath.Join(installedDir, ".csk-install.json")) // #nosec G304
	if err != nil {
		return ""
	}
	var data struct {
		Commit string `json:"commit"`
	}
	_ = json.Unmarshal(payload, &data)
	return data.Commit
}

func cmdStatusAttest(cfg *config.Config, projectRoot, alias string, jsonOut bool) int {
	trusted := cfg.TrustedRegistries()
	registries := make([]registry.Registry, 0, len(trusted))
	for _, entry := range trusted {
		registries = append(registries, registry.Registry{Name: entry.Name, URL: entry.URL, PublicKeys: entry.PublicKeys})
	}
	fetch := registry.NewHTTPFetch(filepath.Join(cfg.Home(), "cache", "registry"), 0, 0, nil)
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
	for alias, project := range cfg.Projects {
		fmt.Printf("%s\t%s\n", alias, project.Path)
		if projectManifest, err := manifest.Load(project.Path); err == nil && projectManifest != nil {
			for _, decl := range projectManifest.Skills {
				fmt.Printf("  %s %s %s\n", decl.Name, decl.Ref.Kind, decl.Ref.Value)
			}
		}
	}
	return exitOK
}

func cmdSkillCheck(args []string) int {
	flags := flag.NewFlagSet("skill check", flag.ContinueOnError)
	localeValue := flags.String("locale", "", "validate against a locale")
	jsonOut := flags.Bool("json", false, "machine-readable output")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	dir := projectRootArg(flags.Args())
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
		fmt.Fprintln(os.Stderr, "curator: global requires a subcommand: init, add, remove, list, install")
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
		if err := flags.Parse(args[1:]); err != nil || flags.NArg() < 1 {
			return exitUsage
		}
		refKind, refValue := pickRef(*tag, *branch, *revision)
		if refKind == "" {
			fmt.Fprintln(os.Stderr, "curator: specify exactly one of --tag, --branch, --revision")
			return exitUsage
		}
		if err := manifest.AddDecl(install.GlobalRoot(cfg.Home()), flags.Arg(0), refKind, refValue, *git, ""); err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		fallthrough
	case "install":
		userHome, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "curator:", err)
			return exitFail
		}
		result := install.Global(cfg, userHome, install.Options{})
		printResult(result)
		if result.Status == "failed" {
			return exitFail
		}
		return exitOK
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
	}
	fmt.Fprintf(os.Stderr, "curator: unknown global subcommand %q\n", args[0])
	return exitUsage
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
		if err := flags.Parse(args[1:]); err != nil || flags.NArg() < 1 || *targets == "" {
			fmt.Fprintln(os.Stderr, "curator: hybrid add requires a name and --targets")
			return exitUsage
		}
		refKind, refValue := pickRef(*tag, *branch, *revision)
		if refKind == "" {
			fmt.Fprintln(os.Stderr, "curator: specify exactly one of --tag, --branch, --revision")
			return exitUsage
		}
		if err := scopes.AddHybridDecl(cfg.Home(), flags.Arg(0), refKind, refValue, *git, strings.Split(*targets, ",")); err != nil {
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
	}
	fmt.Fprintf(os.Stderr, "curator: unknown hybrid subcommand %q\n", args[0])
	return exitUsage
}

func cmdAudit(args []string) int {
	flags := flag.NewFlagSet("audit", flag.ContinueOnError)
	allow := flags.String("allow", "", "pin trust for a content hash")
	reason := flags.String("reason", "", "reason for --allow")
	publish := flags.String("publish", "", "signed audit record (JSON file) to submit")
	registryURL := flags.String("registry", "", "registry base URL for --publish")
	token := flags.String("token", "", "auditor token for --publish (or CURATOR_REGISTRY_TOKEN)")
	if err := flags.Parse(args); err != nil {
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
	fmt.Fprintln(os.Stderr, "curator: audit requires --allow or --publish")
	return exitUsage
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
	if err := flags.Parse(args); err != nil || flags.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "curator: shell-init requires a shell: zsh, bash, powershell")
		return exitUsage
	}
	hook, err := shell.Hook(flags.Arg(0), !*noGlobal)
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
