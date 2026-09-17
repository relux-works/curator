package main

// Draft transport resolution at the CLI entry point.
//
// These tests drive `install --dry-run` through run() with a stand-in git on
// PATH: a POSIX shell script that fails a fetch with fixture stderr or
// rewrites it to a fixture bare repository. Everything else — manifest and
// closure evaluation, credential selection, the strict lane, audit,
// receipting — is the production path. None of these tests run in parallel:
// they mutate process environment per run.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
)

const (
	draftCLIHTTPS    = "https://fixture.test/https-tools.git"
	draftCLIHTTPSAlt = "https://fixture.test/https-tools"
	draftCLISSH      = "ssh://git@fixture.test/ssh-tools.git"
	draftCLISSHAlt   = "ssh://git@fixture.test/ssh-tools"
)

const draftCLIDNSStderr = "fatal: unable to access 'https://fixture.test/https-tools.git': Could not resolve host: fixture.test"

const draftCLITLSStderr = "fatal: unable to access 'https://fixture.test/https-tools.git': SSL certificate problem: self signed certificate"

type draftCLIFixture struct {
	t             *testing.T
	root          string
	configPath    string
	policyPath    string
	providersPath string
	httpsBare     string
	httpsLock     string
	sshBare       string
	sshLock       string
	staticSSH     string
	identity      string
	knownHosts    string
	origPath      string
}

type draftCLIArm struct {
	succeed  bool
	fileRepo string
	stderr   string
}

func setupDraftCLI(t *testing.T) *draftCLIFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	fixture := &draftCLIFixture{t: t, root: t.TempDir(), origPath: os.Getenv("PATH")}
	skillsRoot := filepath.Join(fixture.root, "skills")
	project := filepath.Join(fixture.root, "project")
	fixture.configPath = filepath.Join(fixture.root, "home", "config.json")
	fixture.policyPath = filepath.Join(fixture.root, "home", "source-policy.json")
	fixture.providersPath = filepath.Join(fixture.root, "home", "source-providers.json")
	for _, dir := range []string{filepath.Join(skillsRoot, "skill-a"), project, filepath.Join(fixture.root, "home")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	fixture.httpsBare, fixture.httpsLock = draftCLIExternalRepo(t, fixture.root, "https-remote")
	fixture.sshBare, fixture.sshLock = draftCLIExternalRepo(t, fixture.root, "ssh-remote")
	draftCLISkillRepo(t, skillsRoot, fixture.httpsLock, fixture.sshLock)
	if err := os.WriteFile(filepath.Join(project, manifest.Name), []byte(
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v1"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".agents/\n.codex/skills/\nSkillfile.dev.json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	draftCLIGit(t, "", "init", "--quiet", "-b", "main", project)
	if err := config.Bootstrap(fixture.configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(fixture.configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	fixture.staticSSH = filepath.Join(fixture.root, "static-ssh-wrapper")
	if err := os.WriteFile(fixture.staticSSH, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	fixture.identity = filepath.Join(fixture.root, "id_example")
	if err := os.WriteFile(fixture.identity, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	fixture.knownHosts = filepath.Join(fixture.root, "known_hosts")
	if err := os.WriteFile(fixture.knownHosts, []byte("hosts"), 0o600); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func draftCLIGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	command.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.test",
		"GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.test",
		"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z", "GIT_COMMITTER_DATE=2006-01-02T15:04:05Z",
		"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

// draftCLIExternalRepo builds a bare repository holding one external-build
// descriptor commit. Fixed dates keep the lock stable across runs.
func draftCLIExternalRepo(t *testing.T, root, name string) (bare, lock string) {
	t.Helper()
	work := filepath.Join(root, name+"-work")
	bare = filepath.Join(root, name+".git")
	draftCLIGit(t, "", "init", "--quiet", "-b", "main", "--object-format=sha1", work)
	files := map[string]string{
		"skill-build.json":       `{"schema_version":1,"targets":{"tool":{"driver":"go-repository-v1","build_root":"tools","source_dir":"tools/cmd/tool"}}}`,
		"tools/go.mod":           "module example.test/tool\n",
		"tools/cmd/tool/main.go": "package main\nfunc main() {}\n",
	}
	for name, content := range files {
		path := filepath.Join(work, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	draftCLIGit(t, work, "add", ".")
	draftCLIGit(t, work, "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "external")
	lock = draftCLIGit(t, work, "rev-parse", "HEAD")
	draftCLIGit(t, "", "clone", "--quiet", "--bare", "--", work, bare)
	return bare, lock
}

func draftCLISkillRepo(t *testing.T, skillsRoot, httpsLock, sshLock string) {
	t.Helper()
	repo := filepath.Join(skillsRoot, "skill-a")
	skillManifest := map[string]any{
		"schema_version": 7,
		"capabilities":   map[string]any{},
		"build_repositories": map[string]any{
			"https-tools": map[string]any{"git": draftCLIHTTPS, "locked_commit": map[string]any{"object_format": "sha1", "hex": httpsLock}},
			"ssh-tools":   map[string]any{"git": draftCLISSH, "locked_commit": map[string]any{"object_format": "sha1", "hex": sshLock}},
		},
		"commands": map[string]any{
			"https-cmd": map[string]any{"type": "build", "driver": "go-repository-v1", "repository": "https-tools", "target": "tool"},
			"ssh-cmd":   map[string]any{"type": "build", "driver": "go-repository-v1", "repository": "ssh-tools", "target": "tool"},
		},
	}
	payload, err := json.Marshal(skillManifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "agent-skill.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("---\nname: skill-a\n---\n# Skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	draftCLIGit(t, "", "init", "--quiet", "-b", "main", repo)
	draftCLIGit(t, repo, "add", ".")
	draftCLIGit(t, repo, "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "skill")
	draftCLIGit(t, repo, "tag", "v1")
}

// installDraftFakeGit writes a stand-in git named `git` onto a fresh PATH
// directory and returns its argv log path. Callers scope PATH per run.
// creds arms provider HTTPS secrets by provider name for `credential fill`
// calls from the manager's broker reader; an unarmed provider stays silent
// (exit 1, no material). Every fill request is logged as a `cred:` line so
// tests can assert exactly which providers the run addressed.
func (f *draftCLIFixture) installDraftFakeGit(t *testing.T, arms map[string]draftCLIArm, creds map[string]string) (fakeDir, logPath string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	fakeDir = t.TempDir()
	wrapper := filepath.Join(fakeDir, "git")
	logPath = filepath.Join(fakeDir, "argv.log")
	urls := make([]string, 0, len(arms))
	for url := range arms {
		urls = append(urls, url)
	}
	sort.Strings(urls)
	credProviders := make([]string, 0, len(creds))
	for provider := range creds {
		credProviders = append(credProviders, provider)
	}
	sort.Strings(credProviders)
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	script.WriteString("{ printf 'argv:'; for arg in \"$@\"; do printf ' <%s>' \"$arg\"; done; printf ' | GIT_SSH=%s\\n' \"${GIT_SSH-}\"; } >> " + draftCLIShellQuote(logPath) + "\n")
	script.WriteString("is_cred=0; want_fill=0; for arg in \"$@\"; do if [ \"$arg\" = \"credential\" ]; then is_cred=1; fi; if [ \"$arg\" = \"fill\" ]; then want_fill=1; fi; done\n")
	script.WriteString("if [ \"$is_cred\" = 1 ] && [ \"$want_fill\" = 1 ]; then\n")
	script.WriteString("username=\"\"; while IFS= read -r line; do case \"$line\" in username=*) username=${line#username=};; esac; done\n")
	script.WriteString("{ printf 'cred:<%s>\\n' \"$username\"; } >> " + draftCLIShellQuote(logPath) + "\n")
	script.WriteString("case \"$username\" in\n")
	for _, provider := range credProviders {
		script.WriteString(draftCLIShellQuote(gitcred.ProviderNamespacePrefix+provider) + ") printf 'username=%s\\npassword=%s\\n' \"$username\" " + draftCLIShellQuote(creds[provider]) + "; exit 0;;\n")
	}
	script.WriteString("*) exit 1;;\nesac\nfi\n")
	script.WriteString("if [ \"$is_cred\" = 1 ]; then exit 1; fi\n")
	script.WriteString("is_fetch=0; for arg in \"$@\"; do if [ \"$arg\" = \"fetch\" ]; then is_fetch=1; fi; done\n")
	script.WriteString("if [ \"$is_fetch\" = 1 ]; then\n:\n")
	for _, url := range urls {
		if arms[url].succeed {
			continue
		}
		script.WriteString("for arg in \"$@\"; do if [ \"$arg\" = " + draftCLIShellQuote(url) + " ]; then printf '%s' " + draftCLIShellQuote(arms[url].stderr) + " >&2; exit 128; fi; done\n")
	}
	script.WriteString("fi\nargs=\"\"; for arg in \"$@\"; do\n")
	for _, url := range urls {
		if !arms[url].succeed {
			continue
		}
		script.WriteString("if [ \"$arg\" = " + draftCLIShellQuote(url) + " ]; then arg=" + draftCLIShellQuote("file://"+arms[url].fileRepo) + "; fi\n")
	}
	script.WriteString("case \"$arg\" in protocol.https.allow=always|protocol.ssh.allow=always) arg=protocol.file.allow=always ;; esac\n")
	script.WriteString("args=\"$args\n$arg\"; done\noldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\n")
	script.WriteString("exec " + draftCLIShellQuote(realGit) + " \"$@\"\n")
	if err := os.WriteFile(wrapper, []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	return fakeDir, logPath
}

func draftCLIFetchLines(t *testing.T, logPath string) []string {
	t.Helper()
	payload, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("reading git log: %v", err)
	}
	var fetches []string
	for _, line := range strings.Split(string(payload), "\n") {
		if strings.Contains(line, "<fetch>") {
			fetches = append(fetches, line)
		}
	}
	return fetches
}

func draftCLIPolicy(entries map[string][]string, fallback string) string {
	repositories := make([]string, 0, len(entries))
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		var endpoints []string
		for _, url := range entries[key] {
			endpoints = append(endpoints, `{"url":"`+url+`","authentication":"operator-acme"}`)
		}
		repositories = append(repositories, `"`+key+`":{"endpoints":[`+strings.Join(endpoints, ",")+`],"fallback":"`+fallback+`"}`)
	}
	return `{"schema_version":1,"repositories":{` + strings.Join(repositories, ",") + `}}`
}

// draftCLIPolicyProviders builds a policy whose endpoints each name their own
// provider: entries maps an identity to (url, provider) pairs.
func draftCLIPolicyProviders(entries map[string][][2]string, fallback string) string {
	repositories := make([]string, 0, len(entries))
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		var endpoints []string
		for _, endpoint := range entries[key] {
			endpoints = append(endpoints, `{"url":"`+endpoint[0]+`","authentication":"`+endpoint[1]+`"}`)
		}
		repositories = append(repositories, `"`+key+`":{"endpoints":[`+strings.Join(endpoints, ",")+`],"fallback":"`+fallback+`"}`)
	}
	return `{"schema_version":1,"repositories":{` + strings.Join(repositories, ",") + `}}`
}

// draftCLIOperatorProviders configures operator-acme for both transports the
// matrix exercises: explicitly anonymous HTTPS and the fixture SSH selection.
func draftCLIOperatorProviders(f *draftCLIFixture) string {
	return fmt.Sprintf(`{"schema_version":1,"providers":{"operator-acme":{"https":{"anonymous":true},"ssh":{"identity":%q,"known_hosts":%q}}}}`,
		f.identity, f.knownHosts)
}

// draftCLICredFills returns the `cred:` lines the fake git logged: one per
// `credential fill` request, in request order.
func draftCLICredFills(t *testing.T, logPath string) []string {
	t.Helper()
	payload, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("reading git log: %v", err)
	}
	var fills []string
	for _, line := range strings.Split(string(payload), "\n") {
		if strings.HasPrefix(line, "cred:") {
			fills = append(fills, line)
		}
	}
	return fills
}

var (
	draftCLIHexRE       = regexp.MustCompile(`\b[0-9a-f]{64}\b|\b[0-9a-f]{40}\b`)
	draftCLIBuildRepoRE = regexp.MustCompile(`curator-buildrepo-[0-9]+`)
	draftCLISSHDirRE    = regexp.MustCompile(`curator-draft-ssh-[0-9]+`)
	draftCLIGoBuildRE   = regexp.MustCompile(`go-build[0-9]+`)
	draftCLITagSHARE    = regexp.MustCompile(`tag v1 [0-9a-f]+`)
)

// normalizeDraftCLI replaces every host- and run-specific byte: temporary
// paths (in resolved and unresolved spellings), per-run private-state
// directory numbers, and full object IDs.
func normalizeDraftCLI(t *testing.T, root, value string) string {
	t.Helper()
	prefixes := []string{root}
	if resolved, err := filepath.EvalSymlinks(root); err == nil && resolved != root {
		prefixes = append(prefixes, resolved)
	}
	tmp := os.TempDir()
	prefixes = append(prefixes, tmp)
	if resolved, err := filepath.EvalSymlinks(tmp); err == nil && resolved != tmp {
		prefixes = append(prefixes, resolved)
	}
	sort.Slice(prefixes, func(i, j int) bool { return len(prefixes[i]) > len(prefixes[j]) })
	for _, prefix := range prefixes {
		prefix = strings.TrimRight(prefix, `/\`)
		value = strings.ReplaceAll(value, prefix, "$T")
	}
	value = draftCLIBuildRepoRE.ReplaceAllString(value, `curator-buildrepo-$$N`)
	value = draftCLISSHDirRE.ReplaceAllString(value, `curator-draft-ssh-$$N`)
	value = draftCLIGoBuildRE.ReplaceAllString(value, `go-build$$N`)
	value = draftCLITagSHARE.ReplaceAllString(value, `tag v1 $$S`)
	value = draftCLIHexRE.ReplaceAllString(value, `$$H`)
	return value
}

func draftCLIShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// runDraftInstall runs one CLI invocation with per-run process environment.
// policy nil means no policy file; providers nil means no providers file;
// switchOn sets the draft switch to "1".
func (f *draftCLIFixture) runDraftInstall(t *testing.T, fakeDir, logPath string, switchOn bool, policy, providers *string) (int, string, string, []string) {
	t.Helper()
	saved := map[string]*string{}
	set := map[string]string{
		"PATH":                          fakeDir + string(os.PathListSeparator) + f.origPath,
		"GIT_SSH":                       f.staticSSH,
		"CURATOR_BUILD_SSH_IDENTITY":    f.identity,
		"CURATOR_BUILD_SSH_KNOWN_HOSTS": f.knownHosts,
	}
	unset := []string{"SSH_AUTH_SOCK", "CURATOR_BUILD_SSH_AGENT", "CURATOR_BUILD_HTTPS_TOKEN", "CURATOR_BUILD_HTTPS_HOST", install.EnvDraftTransportResolution}
	if switchOn {
		set[install.EnvDraftTransportResolution] = "1"
	}
	for key := range set {
		if value, ok := os.LookupEnv(key); ok {
			held := value
			saved[key] = &held
		} else {
			saved[key] = nil
		}
	}
	for _, key := range unset {
		if _, seen := saved[key]; seen {
			continue
		}
		if value, ok := os.LookupEnv(key); ok {
			held := value
			saved[key] = &held
		} else {
			saved[key] = nil
		}
	}
	defer func() {
		for key, held := range saved {
			if held == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *held)
			}
		}
	}()
	for key, value := range set {
		if err := os.Setenv(key, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range unset {
		if _, overridden := set[key]; overridden {
			continue
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	if policy == nil {
		if err := os.Remove(f.policyPath); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	} else if err := os.WriteFile(f.policyPath, []byte(*policy), 0o644); err != nil {
		t.Fatal(err)
	}
	if providers == nil {
		if err := os.Remove(f.providersPath); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	} else if err := os.WriteFile(f.providersPath, []byte(*providers), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := capture(t, f.configPath, "install", "app", "--dry-run")
	return code, stdout, stderr, draftCLIFetchLines(t, logPath)
}

func draftCLISuccessArms(f *draftCLIFixture) map[string]draftCLIArm {
	return map[string]draftCLIArm{
		draftCLIHTTPS: {succeed: true, fileRepo: f.httpsBare},
		draftCLISSH:   {succeed: true, fileRepo: f.sshBare},
	}
}

// draftCLILegacyRun formats one legacy run for the golden: exit code,
// normalized stdout, and normalized fetch lines.
func draftCLILegacyRun(t *testing.T, root string, code int, stdout string, fetches []string) string {
	t.Helper()
	var out strings.Builder
	_, _ = fmt.Fprintf(&out, "exit: %d\n", code)
	out.WriteString("--- stdout ---\n")
	out.WriteString(normalizeDraftCLI(t, root, stdout))
	if !strings.HasSuffix(stdout, "\n") {
		out.WriteString("\n")
	}
	out.WriteString("--- fetches ---\n")
	out.WriteString(normalizeDraftCLI(t, root, strings.Join(fetches, "\n")))
	out.WriteString("\n")
	return out.String()
}

func TestDraftTransportLegacyGolden(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}
	fixture := setupDraftCLI(t)
	policy := draftCLIPolicy(map[string][]string{
		"fixture.test/https-tools": {draftCLIHTTPS},
		"fixture.test/ssh-tools":   {draftCLISSH},
	}, "none")
	// A present-but-invalid providers file on every legacy run proves the
	// legacy lane never consults the provider table.
	broken := `not json`

	run := func(switchOn bool, policy *string) (int, string, string, []string) {
		t.Helper()
		fakeDir, logPath := fixture.installDraftFakeGit(t, draftCLISuccessArms(fixture), nil)
		return fixture.runDraftInstall(t, fakeDir, logPath, switchOn, policy, &broken)
	}

	code, stdout, stderr, fetches := run(false, &policy)
	if code != exitOK {
		t.Fatalf("switch off = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("switch-off stderr:\n%s", stderr)
	}
	if len(fetches) != 2 {
		t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
	}
	golden := draftCLILegacyRun(t, fixture.root, code, stdout, fetches)

	want, err := os.ReadFile(filepath.Join("testdata", "draft-transport-legacy.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if golden != string(want) {
		t.Fatalf("legacy run differs from the committed golden:\n%s", golden)
	}
	for _, fetch := range fetches {
		if strings.Contains(fetch, buildrepo.SSHWrapperName) {
			t.Fatalf("legacy fetch used the manager wrapper:\n%s", fetch)
		}
	}

	// Switch on without a policy, and switch off without one, must produce
	// the same bytes on the same fixture.
	code2, stdout2, stderr2, fetches2 := run(true, nil)
	if code2 != exitOK || stderr2 != "" {
		t.Fatalf("switch on without policy = %d\nstdout:\n%s\nstderr:\n%s", code2, stdout2, stderr2)
	}
	if other := draftCLILegacyRun(t, fixture.root, code2, stdout2, fetches2); other != golden {
		t.Fatalf("switch on without policy differs from the legacy golden:\n%s", other)
	}
	code3, stdout3, stderr3, fetches3 := run(false, nil)
	if code3 != exitOK || stderr3 != "" {
		t.Fatalf("switch off without policy = %d\nstdout:\n%s\nstderr:\n%s", code3, stdout3, stderr3)
	}
	if other := draftCLILegacyRun(t, fixture.root, code3, stdout3, fetches3); other != golden {
		t.Fatalf("switch off without policy differs from the legacy golden:\n%s", other)
	}
}

func TestDraftTransportResolvedMatrix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}

	t.Run("fallback on availability then success", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, map[string]draftCLIArm{
			draftCLIHTTPS:    {stderr: draftCLIDNSStderr},
			draftCLIHTTPSAlt: {succeed: true, fileRepo: fixture.httpsBare},
			draftCLISSH:      {succeed: true, fileRepo: fixture.sshBare},
		}, nil)
		policy := draftCLIPolicy(map[string][]string{
			"fixture.test/https-tools": {draftCLIHTTPS, draftCLIHTTPSAlt},
			"fixture.test/ssh-tools":   {draftCLISSH},
		}, "availability-auth")
		operator := draftCLIOperatorProviders(fixture)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &operator)
		if code != exitOK {
			t.Fatalf("resolved = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if len(fetches) != 3 {
			t.Fatalf("%d fetches, want 3:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if !strings.Contains(fetches[0], "<"+draftCLIHTTPS+">") || !strings.Contains(fetches[1], "<"+draftCLIHTTPSAlt+">") {
			t.Fatalf("https attempts out of order:\n%s", strings.Join(fetches, "\n"))
		}
		if !strings.Contains(fetches[2], "<"+draftCLISSH+">") || !strings.Contains(fetches[2], buildrepo.SSHWrapperName) {
			t.Fatalf("ssh attempt missed the manager wrapper:\n%s", strings.Join(fetches, "\n"))
		}
	})

	t.Run("fail-closed class refuses with one fetch", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, map[string]draftCLIArm{
			draftCLIHTTPS:    {stderr: draftCLITLSStderr},
			draftCLIHTTPSAlt: {succeed: true, fileRepo: fixture.httpsBare},
			draftCLISSH:      {succeed: true, fileRepo: fixture.sshBare},
		}, nil)
		policy := draftCLIPolicy(map[string][]string{
			"fixture.test/https-tools": {draftCLIHTTPS, draftCLIHTTPSAlt},
			"fixture.test/ssh-tools":   {draftCLISSH},
		}, "availability-auth")
		operator := draftCLIOperatorProviders(fixture)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &operator)
		if code == exitOK {
			t.Fatalf("refusal exited 0\nstdout:\n%s", stdout)
		}
		if combined := stdout + stderr; !strings.Contains(combined, buildrepo.CodeSourceUnavailable) {
			t.Fatalf("refusal lacks the lane diagnostic:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
		if len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftCLIHTTPS+">") {
			t.Fatalf("fetches = %q, want exactly the refused first attempt", fetches)
		}
	})

	t.Run("invalid policy fails before any fetch", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, draftCLISuccessArms(fixture), nil)
		policy := `{"schema_version":1}`
		operator := draftCLIOperatorProviders(fixture)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &operator)
		if code == exitOK {
			t.Fatalf("invalid policy exited 0\nstdout:\n%s", stdout)
		}
		// The pipeline masks every acquisition failure as the lane
		// diagnostic; the repository_policy_invalid class itself is pinned
		// at install level. What matters here is the failure with no fetch.
		if combined := stdout + stderr; !strings.Contains(combined, buildrepo.CodeSourceUnavailable) {
			t.Fatalf("invalid policy lacks the lane diagnostic:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
		if len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})
}

// TestDraftTransportProviderAdmission pins named provider admission through
// the CLI entry point: an unknown provider name refuses before any fetch
// traffic, two distinct configured providers each resolve only their own
// broker material, and an explicitly anonymous provider fetches without
// broker traffic.
func TestDraftTransportProviderAdmission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}

	t.Run("unknown https provider refuses before any fetch", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, draftCLISuccessArms(fixture), nil)
		policy := draftCLIPolicyProviders(map[string][][2]string{
			"fixture.test/https-tools": {{draftCLIHTTPS, "prov-unknown"}},
			"fixture.test/ssh-tools":   {{draftCLISSH, "operator-acme"}},
		}, "none")
		operator := draftCLIOperatorProviders(fixture)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &operator)
		if code == exitOK {
			t.Fatalf("unknown provider exited 0\nstdout:\n%s", stdout)
		}
		if combined := stdout + stderr; !strings.Contains(combined, buildrepo.CodeSourceUnavailable) {
			t.Fatalf("refusal lacks the lane diagnostic:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
		if len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if fills := draftCLICredFills(t, logPath); len(fills) != 0 {
			t.Fatalf("broker fills = %q, want none", fills)
		}
	})

	t.Run("unknown ssh provider refuses without fetching", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, draftCLISuccessArms(fixture), nil)
		policy := draftCLIPolicyProviders(map[string][][2]string{
			"fixture.test/https-tools": {{draftCLIHTTPS, "operator-acme"}},
			"fixture.test/ssh-tools":   {{draftCLISSH, "prov-unknown"}},
		}, "none")
		operator := draftCLIOperatorProviders(fixture)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &operator)
		if code == exitOK {
			t.Fatalf("unknown provider exited 0\nstdout:\n%s", stdout)
		}
		if combined := stdout + stderr; !strings.Contains(combined, buildrepo.CodeSourceUnavailable) {
			t.Fatalf("refusal lacks the lane diagnostic:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
		}
		if len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftCLIHTTPS+">") {
			t.Fatalf("fetches = %q, want exactly the configured https attempt", fetches)
		}
	})

	t.Run("distinct providers keep their own material", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, map[string]draftCLIArm{
			draftCLIHTTPS:    {succeed: true, fileRepo: fixture.httpsBare},
			draftCLIHTTPSAlt: {succeed: true, fileRepo: fixture.httpsBare},
			draftCLISSH:      {succeed: true, fileRepo: fixture.sshBare},
		}, map[string]string{"prov-a": "secret-a"})
		policy := draftCLIPolicyProviders(map[string][][2]string{
			"fixture.test/https-tools": {{draftCLIHTTPS, "prov-silent"}, {draftCLIHTTPSAlt, "prov-a"}},
			"fixture.test/ssh-tools":   {{draftCLISSH, "operator-acme"}},
		}, "availability-auth")
		providers := fmt.Sprintf(`{"schema_version":1,"providers":{`+
			`"prov-silent":{"https":{"username":"silent"}},`+
			`"prov-a":{"https":{"username":"a"}},`+
			`"operator-acme":{"https":{"anonymous":true},"ssh":{"identity":%q,"known_hosts":%q}}}}`,
			fixture.identity, fixture.knownHosts)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &providers)
		if code != exitOK {
			t.Fatalf("resolved = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if len(fetches) != 2 {
			t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if !strings.Contains(fetches[0], "<"+draftCLIHTTPSAlt+">") {
			t.Fatalf("silent provider fetched its endpoint:\n%s", strings.Join(fetches, "\n"))
		}
		if !strings.Contains(fetches[1], "<"+draftCLISSH+">") {
			t.Fatalf("ssh attempt missing:\n%s", strings.Join(fetches, "\n"))
		}
		fills := draftCLICredFills(t, logPath)
		want := []string{
			"cred:<" + gitcred.ProviderNamespacePrefix + "prov-silent>",
			"cred:<" + gitcred.ProviderNamespacePrefix + "prov-a>",
		}
		if fmt.Sprint(fills) != fmt.Sprint(want) {
			t.Fatalf("broker fills = %q, want %q", fills, want)
		}
	})

	t.Run("explicit anonymous stays anonymous", func(t *testing.T) {
		fixture := setupDraftCLI(t)
		fakeDir, logPath := fixture.installDraftFakeGit(t, draftCLISuccessArms(fixture), nil)
		policy := draftCLIPolicyProviders(map[string][][2]string{
			"fixture.test/https-tools": {{draftCLIHTTPS, "anon"}},
			"fixture.test/ssh-tools":   {{draftCLISSH, "operator-acme"}},
		}, "none")
		providers := fmt.Sprintf(`{"schema_version":1,"providers":{`+
			`"anon":{"https":{"anonymous":true}},`+
			`"operator-acme":{"https":{"anonymous":true},"ssh":{"identity":%q,"known_hosts":%q}}}}`,
			fixture.identity, fixture.knownHosts)
		code, stdout, stderr, fetches := fixture.runDraftInstall(t, fakeDir, logPath, true, &policy, &providers)
		if code != exitOK {
			t.Fatalf("resolved = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
		}
		if len(fetches) != 2 {
			t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if fills := draftCLICredFills(t, logPath); len(fills) != 0 {
			t.Fatalf("broker fills = %q, want none", fills)
		}
	})
}

func TestProductionBinaryDispatchesSSHWrapper(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "curator"+productionBinarySuffix())
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build production curator: %v\n%s", err, output)
	}
	wrapperName := buildrepo.SSHWrapperName + productionBinarySuffix()
	wrapper := filepath.Join(t.TempDir(), wrapperName)
	payload, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(wrapper, payload, 0o700); err != nil {
		t.Fatal(err)
	}

	writeFile := func(name string, content []byte) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	emptyConfig := writeFile("ssh.config", nil)
	knownHosts := writeFile("known_hosts", []byte("hosts"))
	emptyKnownHosts := writeFile("empty_known_hosts", nil)
	identity := writeFile("id_example", []byte("key"))
	sshLog := filepath.Join(t.TempDir(), "ssh-argv.log")

	fakeSSH := ""
	if runtime.GOOS != "windows" {
		fakeSSH = filepath.Join(t.TempDir(), "ssh")
		script := "#!/bin/sh\nfor arg in \"$@\"; do printf '<%s>' \"$arg\"; done >> " + draftCLIShellQuote(sshLog) + "\nprintf '\\n' >> " + draftCLIShellQuote(sshLog) + "\nexit 3\n"
		if err := os.WriteFile(fakeSSH, []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
	} else {
		fakeSSH = filepath.Join(t.TempDir(), "ssh.exe")
		if err := os.WriteFile(fakeSSH, []byte("MZ"), 0o700); err != nil {
			t.Fatal(err)
		}
	}

	policy := map[string]any{
		"Wrapper": wrapper, "SSH": fakeSSH,
		"ExpectedHost": "git@fixture.test", "RepositoryPath": "/repository.git",
		"EmptyConfig": emptyConfig, "KnownHosts": knownHosts, "EmptyKnownHosts": emptyKnownHosts,
		"Identity": identity, "ConnectTimeout": 15,
	}
	state, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	statePath := writeFile("ssh-wrapper-state.json", append(state, '\n'))

	run := func(args []string, stateValue string) (int, string) {
		t.Helper()
		command := exec.Command(wrapper, args...)
		command.Env = []string{"PATH=" + os.Getenv("PATH"), buildrepo.EnvSSHWrapperState + "=" + stateValue}
		output, err := command.CombinedOutput()
		if err == nil {
			return 0, string(output)
		}
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode(), string(output)
		}
		t.Fatalf("wrapper invocation failed: %v\n%s", err, output)
		return -1, ""
	}

	if runtime.GOOS != "windows" {
		t.Run("passthrough", func(t *testing.T) {
			code, output := run([]string{"git@fixture.test", "git-upload-pack '/repository.git'"}, statePath)
			if code != 3 {
				t.Fatalf("wrapper exit = %d, want the ssh exit 3\n%s", code, output)
			}
			logged, err := os.ReadFile(sshLog)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"<-F>", "<" + emptyConfig + ">", "<BatchMode=yes>", "<git@fixture.test>", "<git-upload-pack '/repository.git'>"} {
				if !strings.Contains(string(logged), want) {
					t.Fatalf("ssh argv lacks %s:\n%s", want, logged)
				}
			}
		})
	}

	for _, testCase := range []struct {
		name string
		args []string
	}{
		{"missing state refuses", []string{"git@fixture.test", "git-upload-pack '/repository.git'"}},
		{"wrong tuple refuses", []string{"other@fixture.test", "git-upload-pack '/repository.git'"}},
		{"single arg refuses before CLI parsing", []string{"--version"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			stateValue := ""
			if testCase.name == "wrong tuple refuses" {
				stateValue = statePath
			}
			if code, output := run(testCase.args, stateValue); code != 1 || output != "" {
				t.Fatalf("refusal = (%d, %q), want silent exit 1", code, output)
			}
		})
	}

	t.Run("malformed state refuses", func(t *testing.T) {
		bad := writeFile("bad-state.json", []byte("{broken"))
		if code, output := run([]string{"git@fixture.test", "git-upload-pack '/repository.git'"}, bad); code != 1 || output != "" {
			t.Fatalf("refusal = (%d, %q), want silent exit 1", code, output)
		}
	})
}

// Windows reaches the resolved lane through the same production caller, but
// the CLI pipeline masks every acquisition failure as the lane diagnostic
// (buildrepo.RunPipeline), so the lane's typed platform refusal is asserted
// at install level by
// TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess instead of here.
