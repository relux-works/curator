package install

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
	"github.com/relux-works/curator/internal/skillspec"
)

func TestDraftTransportEnabled(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name  string
		value *string
		want  bool
	}{
		{"one enables", strPtr("1"), true},
		{"empty disables", strPtr(""), false},
		{"zero disables", strPtr("0"), false},
		{"true disables", strPtr("true"), false},
		{"two disables", strPtr("2"), false},
		{"unset disables", nil, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			getenv := func(string) string { return "" }
			if testCase.value != nil {
				getenv = func(string) string { return *testCase.value }
			}
			if got := DraftTransportEnabled(getenv); got != testCase.want {
				t.Fatalf("DraftTransportEnabled = %v, want %v", got, testCase.want)
			}
		})
	}
	if DraftTransportEnabled(nil) {
		t.Fatal("DraftTransportEnabled(nil) = true, want false")
	}
}

func strPtr(value string) *string { return &value }

func TestDraftTransportPlanConversion(t *testing.T) {
	t.Parallel()
	one := config.Resolution{Identity: "fixture.test/repository",
		Attempts: []config.Attempt{{URL: "https://fixture.test/repository.git"}}, Fallback: config.FallbackNone}
	plan, err := draftTransportPlan(one)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Identity != one.Identity || len(plan.Attempts) != 1 || plan.Attempts[0].URL != one.Attempts[0].URL ||
		plan.Attempts[0].Authentication != "" || plan.Fallback != buildrepo.TransportFallbackNone {
		t.Fatalf("legacy-shape plan = %+v", plan)
	}
	two := config.Resolution{Identity: "fixture.test/repository",
		Attempts: []config.Attempt{
			{URL: "https://fixture.test/repository.git", Authentication: "operator-acme"},
			{URL: "https://fixture.test/repository", Authentication: "operator-acme"},
		}, Fallback: config.FallbackAvailabilityAuth}
	plan, err = draftTransportPlan(two)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Attempts) != 2 || plan.Attempts[0].Authentication != "operator-acme" ||
		plan.Attempts[1].URL != "https://fixture.test/repository" || plan.Fallback != buildrepo.TransportFallbackAvailabilityAuth {
		t.Fatalf("two-endpoint plan = %+v", plan)
	}
	bad := config.Resolution{Identity: "fixture.test/repository",
		Attempts: []config.Attempt{{URL: "https://fixture.test/repository.git"}}, Fallback: "sometimes"}
	if _, err := draftTransportPlan(bad); err == nil || !strings.Contains(err.Error(), config.CodeRepositoryPolicyInvalid) {
		t.Fatalf("unknown fallback err = %v, want %s", err, config.CodeRepositoryPolicyInvalid)
	}
	mistranslated := config.Resolution{Identity: "fixture.test/repository",
		Attempts: []config.Attempt{{URL: "not a url"}}, Fallback: config.FallbackNone}
	if _, err := draftTransportPlan(mistranslated); err == nil {
		t.Fatal("malformed endpoint converted without an error")
	}
}

// TestDraftTransportAuth pins named provider admission at the production
// caller: a policy provider name resolves only through the operator's
// provider table and broker. An unconfigured name refuses before any fetch
// — it is never treated as anonymous and never receives another
// selection's material — while an explicitly anonymous provider fetches
// without broker traffic.
func TestDraftTransportAuth(t *testing.T) {
	bare, commit := draftBareFixture(t)
	lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}

	newDeps := func(tool buildrepo.GitTool, policy, providers string, reader buildrepo.ProviderSecretReader) ExternalDeps {
		return ExternalDeps{GitTool: tool, DraftTransportResolution: true, DraftPolicyPath: policy,
			DraftProvidersPath: providers, DraftProviderReader: reader,
			Audit: func(context.Context, buildrepo.AuditSubject) error { return nil }}
	}
	acquire := func(t *testing.T, deps ExternalDeps, tool buildrepo.GitTool, git, transport string) (*buildrepo.Snapshot, error) {
		t.Helper()
		return deps.acquireDraftNetwork(context.Background(), tool, git, transport, "fixture.test/repository", lock, "", "", "")
	}
	otherOnly := `{"schema_version":1,"providers":{"other":{"https":{"anonymous":true}}}}`

	t.Run("unconfigured named https refuses before any fetch", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		reader := &draftProviderReader{secrets: map[string]gitcred.HostCredential{
			"other": {Username: "other", Secret: "secret"},
		}}
		policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary))
		providers := draftProvidersFile(t, otherOnly)
		_, err := acquire(t, newDeps(tool, policy, providers, reader), tool, draftHTTPSPrimary, "https")
		if buildrepo.ErrorCode(err) != buildrepo.CodeRepositoryEndpointUnavailable {
			t.Fatalf("err = %v, want %s", err, buildrepo.CodeRepositoryEndpointUnavailable)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if len(reader.queries) != 0 {
			t.Fatalf("broker queries = %q, want none", reader.queries)
		}
	})

	t.Run("unconfigured named ssh refuses before any fetch", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftSSHPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftSSHPrimary))
		providers := draftProvidersFile(t, otherOnly)
		_, err := acquire(t, newDeps(tool, policy, providers, nil), tool, draftSSHPrimary, "ssh")
		if buildrepo.ErrorCode(err) != buildrepo.CodeRepositoryEndpointUnavailable {
			t.Fatalf("err = %v, want %s", err, buildrepo.CodeRepositoryEndpointUnavailable)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("distinct providers resolve their own broker material", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary:   {stderr: draftDNSStderr},
			draftHTTPSAlternate: {succeed: true, fileRepo: bare},
		})
		reader := &draftProviderReader{secrets: map[string]gitcred.HostCredential{
			"prov-a": {Username: "a", Secret: "secret-a"},
			"prov-b": {Username: "b", Secret: "secret-b"},
		}}
		policy := draftPolicyFile(t, draftPolicyEntryProviders(
			[2]string{draftHTTPSPrimary, "prov-a"}, [2]string{draftHTTPSAlternate, "prov-b"}))
		providers := draftProvidersFile(t, `{"schema_version":1,"providers":{`+
			`"prov-a":{"https":{"username":"a"}},"prov-b":{"https":{"username":"b"}}}}`)
		snapshot, err := acquire(t, newDeps(tool, policy, providers, reader), tool, draftHTTPSPrimary, "https")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 2 {
			t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		want := []string{"prov-a@fixture.test", "prov-b@fixture.test"}
		if fmt.Sprint(reader.queries) != fmt.Sprint(want) {
			t.Fatalf("broker queries = %q, want %q", reader.queries, want)
		}
	})

	t.Run("a silent provider receives no other provider's material", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary:   {succeed: true, fileRepo: bare},
			draftHTTPSAlternate: {succeed: true, fileRepo: bare},
		})
		reader := &draftProviderReader{secrets: map[string]gitcred.HostCredential{
			"prov-a": {Username: "a", Secret: "secret-a"},
		}}
		policy := draftPolicyFile(t, draftPolicyEntryProviders(
			[2]string{draftHTTPSPrimary, "prov-silent"}, [2]string{draftHTTPSAlternate, "prov-a"}))
		providers := draftProvidersFile(t, `{"schema_version":1,"providers":{`+
			`"prov-silent":{"https":{"username":"silent"}},"prov-a":{"https":{"username":"a"}}}}`)
		snapshot, err := acquire(t, newDeps(tool, policy, providers, reader), tool, draftHTTPSPrimary, "https")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		fetches := draftFetchLines(t, logPath)
		if len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftHTTPSAlternate+">") {
			t.Fatalf("fetches = %q, want exactly the second attempt", fetches)
		}
	})

	t.Run("explicit anonymous fetches without broker traffic", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary))
		providers := draftProvidersFile(t, `{"schema_version":1,"providers":{"operator-acme":{"https":{"anonymous":true}}}}`)
		snapshot, err := acquire(t, newDeps(tool, policy, providers, nil), tool, draftHTTPSPrimary, "https")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("invalid providers file fails before any fetch", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary))
		providers := draftProvidersFile(t, `not json`)
		_, err := acquire(t, newDeps(tool, policy, providers, nil), tool, draftHTTPSPrimary, "https")
		if err == nil || !strings.Contains(err.Error(), config.CodeRepositoryPolicyInvalid) {
			t.Fatalf("err = %v, want %s", err, config.CodeRepositoryPolicyInvalid)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("providerless resolution never opens the providers file", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, `{"schema_version":1,"repositories":{}}`)
		providers := draftProvidersFile(t, `not json`)
		snapshot, err := acquire(t, newDeps(tool, policy, providers, nil), tool, draftHTTPSPrimary, "https")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})
}

// draftProviderReader is a stub trusted broker: HTTPS secrets keyed by
// provider, with every query recorded for addressing assertions.
type draftProviderReader struct {
	secrets map[string]gitcred.HostCredential
	queries []string
}

func (r *draftProviderReader) ReadProvider(_ context.Context, provider, host string) (gitcred.HostCredential, bool) {
	r.queries = append(r.queries, provider+"@"+host)
	material, ok := r.secrets[provider]
	return material, ok
}

func operatorPath(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("material"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDraftPlanNeedsSSH(t *testing.T) {
	t.Parallel()
	https := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{{URL: "https://fixture.test/repository.git"}}}
	if draftPlanNeedsSSH(https) {
		t.Fatal("https-only plan needs SSH")
	}
	mixed := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{
		{URL: "https://fixture.test/repository.git"},
		{URL: "ssh://git@fixture.test/repository.git"},
	}}
	if !draftPlanNeedsSSH(mixed) {
		t.Fatal("mixed plan needs no SSH")
	}
	malformed := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{{URL: "not a url"}}}
	if draftPlanNeedsSSH(malformed) {
		t.Fatal("malformed plan needs SSH")
	}
}

func TestDraftPlanNamesProvider(t *testing.T) {
	t.Parallel()
	named := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{{URL: "https://fixture.test/repository.git", Authentication: "operator-acme"}}}
	if !draftPlanNamesProvider(named) {
		t.Fatal("named plan names no provider")
	}
	providerless := buildrepo.TransportPlan{Attempts: []buildrepo.TransportAttempt{{URL: "https://fixture.test/repository.git"}}}
	if draftPlanNamesProvider(providerless) {
		t.Fatal("providerless plan names a provider")
	}
	if draftPlanNamesProvider(buildrepo.TransportPlan{}) {
		t.Fatal("empty plan names a provider")
	}
}

func TestDraftSSHWrapperBase(t *testing.T) {
	base, cleanup, err := draftSSHWrapperBase()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if base.SSH == "" || base.EmptyConfig == "" || base.EmptyKnownHosts == "" || base.ConnectTimeout <= 0 {
		t.Fatalf("SSH base = %+v", base)
	}
	for _, path := range []string{base.SSH, base.EmptyConfig, base.EmptyKnownHosts} {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("base path %s is not an admitted regular file: %v", path, err)
		}
	}
	cleanup()
	if _, err := os.Lstat(base.EmptyConfig); !os.IsNotExist(err) {
		t.Fatalf("empty config survives cleanup: %v", err)
	}

	t.Run("missing ssh refuses later with a zero base", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		base, cleanup, err := draftSSHWrapperBase()
		defer cleanup()
		if err != nil {
			t.Fatal(err)
		}
		if base != (buildrepo.SSHWrapperBase{}) {
			t.Fatalf("base without ssh = %+v, want zero", base)
		}
	})
}

func TestDraftManagerExecutable(t *testing.T) {
	t.Parallel()
	executable, err := draftManagerExecutable()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(executable) {
		t.Fatalf("manager executable %q is not absolute", executable)
	}
	info, err := os.Lstat(executable)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("manager executable is not an admitted regular file: %v", err)
	}
}

// --- selection tests ---
//
// The default acquisition closure is driven through acquireDraftNetwork
// against a stand-in git: a POSIX shell wrapper around the real git that
// fails a fetch with fixture stderr or rewrites it to a fixture bare
// repository. Everything else is the real lane.

const (
	draftHTTPSPrimary   = "https://fixture.test/repository.git"
	draftHTTPSAlternate = "https://fixture.test/repository"
	draftSSHPrimary     = "ssh://git@fixture.test/repository.git"
)

const draftDNSStderr = "fatal: unable to access 'https://fixture.test/repository.git': Could not resolve host: fixture.test"

const draftTLSStderr = "fatal: unable to access 'https://fixture.test/repository.git': SSL certificate problem: self signed certificate"

type draftGitArm struct {
	succeed  bool
	fileRepo string
	stderr   string
}

func draftGitTool(t *testing.T, arms map[string]draftGitArm) (buildrepo.GitTool, string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; production code is platform-neutral")
	}
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	git, err = filepath.EvalSymlinks(git)
	if err != nil {
		t.Fatal(err)
	}
	execPath, err := exec.Command(git, "--exec-path").Output()
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(git, "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	wrapper := filepath.Join(root, "git-wrapper")
	logPath := filepath.Join(root, "argv.log")
	var script strings.Builder
	script.WriteString("#!/bin/sh\n")
	fmt.Fprintf(&script, "{ printf 'argv:'; for arg in \"$@\"; do printf ' <%%s>' \"$arg\"; done; printf ' | GIT_SSH=%%s\\n' \"${GIT_SSH-}\"; } >> %s\n", shellQuote(logPath))
	script.WriteString("is_fetch=0; for arg in \"$@\"; do if [ \"$arg\" = \"fetch\" ]; then is_fetch=1; fi; done\n")
	script.WriteString("if [ \"$is_fetch\" = 1 ]; then\n:\n")
	for _, url := range sortedDraftURLs(arms) {
		arm := arms[url]
		if !arm.succeed {
			fmt.Fprintf(&script, "for arg in \"$@\"; do if [ \"$arg\" = %s ]; then printf '%%s' %s >&2; exit 128; fi; done\n", shellQuote(url), shellQuote(arm.stderr))
		}
	}
	script.WriteString("fi\nargs=\"\"; for arg in \"$@\"; do\n")
	for _, url := range sortedDraftURLs(arms) {
		if arms[url].succeed {
			fmt.Fprintf(&script, "if [ \"$arg\" = %s ]; then arg=%s; fi\n", shellQuote(url), shellQuote("file://"+arms[url].fileRepo))
		}
	}
	script.WriteString("case \"$arg\" in protocol.https.allow=always|protocol.ssh.allow=always) arg=protocol.file.allow=always ;; esac\n")
	script.WriteString("args=\"$args\n$arg\"; done\noldifs=$IFS; IFS='\n'; set -- $args; IFS=$oldifs\n")
	fmt.Fprintf(&script, "exec %s \"$@\"\n", shellQuote(git))
	if err := os.WriteFile(wrapper, []byte(script.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	askPass, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return buildrepo.GitTool{Executable: wrapper, ExecPath: strings.TrimSpace(string(execPath)), AllowedVersions: []string{strings.TrimSpace(string(version))}, AskPass: askPass, SSHWrapper: askPass}, logPath
}

func sortedDraftURLs(arms map[string]draftGitArm) []string {
	urls := make([]string, 0, len(arms))
	for url := range arms {
		urls = append(urls, url)
	}
	for i := 1; i < len(urls); i++ {
		for j := i; j > 0 && urls[j] < urls[j-1]; j-- {
			urls[j], urls[j-1] = urls[j-1], urls[j]
		}
	}
	return urls
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func draftFetchLines(t *testing.T, logPath string) []string {
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

// draftBareFixture builds a bare repository holding one commit. Fixed dates
// keep the lock stable across runs.
func draftBareFixture(t *testing.T) (bare, commit string) {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	work := filepath.Join(root, "work")
	bare = filepath.Join(root, "remote.git")
	run := func(dir string, args ...string) string {
		t.Helper()
		command := exec.Command(git, args...)
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
	run("", "init", "--quiet", "--object-format=sha1", work)
	if err := os.WriteFile(filepath.Join(work, "hello.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(work, "add", "--", "hello.txt")
	run(work, "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
	commit = run(work, "rev-parse", "HEAD")
	run("", "clone", "--quiet", "--bare", "--", work, bare)
	return bare, commit
}

func draftPolicyFile(t *testing.T, payload string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source-policy.json")
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func draftPolicyEntry(urls ...string) string {
	var endpoints []string
	for _, url := range urls {
		endpoints = append(endpoints, fmt.Sprintf(`{"url":%q,"authentication":"operator-acme"}`, url))
	}
	return fmt.Sprintf(`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[%s],"fallback":"availability-auth"}}}`,
		strings.Join(endpoints, ","))
}

func draftPolicyEntryProviders(entries ...[2]string) string {
	var endpoints []string
	for _, entry := range entries {
		endpoints = append(endpoints, fmt.Sprintf(`{"url":%q,"authentication":%q}`, entry[0], entry[1]))
	}
	return fmt.Sprintf(`{"schema_version":1,"repositories":{"fixture.test/repository":{"endpoints":[%s],"fallback":"availability-auth"}}}`,
		strings.Join(endpoints, ","))
}

func draftProvidersFile(t *testing.T, payload string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source-providers.json")
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func draftSSHProvidersDoc(ssh buildrepo.OperatorSSHCredentials) string {
	return fmt.Sprintf(`{"schema_version":1,"providers":{"operator-acme":{"ssh":{"identity":%q,"known_hosts":%q}}}}`,
		ssh.Identity, ssh.KnownHosts)
}

func draftSSHCredentials(t *testing.T) buildrepo.OperatorSSHCredentials {
	t.Helper()
	return buildrepo.OperatorSSHCredentials{Identity: operatorPath(t, "id"), KnownHosts: operatorPath(t, "known_hosts")}
}

func TestAcquireDraftNetworkSelection(t *testing.T) {
	bare, commit := draftBareFixture(t)
	lock := buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}

	newDeps := func(tool buildrepo.GitTool, enabled bool, policy, providers string) ExternalDeps {
		return ExternalDeps{GitTool: tool, DraftTransportResolution: enabled, DraftPolicyPath: policy,
			DraftProvidersPath: providers,
			Audit:              func(context.Context, buildrepo.AuditSubject) error { return nil }}
	}
	// The policy names operator-acme everywhere; anonymous HTTPS keeps these
	// selection subtests off the broker (nil reader).
	anonymous := draftProvidersFile(t, `{"schema_version":1,"providers":{"operator-acme":{"https":{"anonymous":true}}}}`)
	// A present-but-invalid providers file proves the legacy lane never
	// consults the provider table.
	brokenProviders := draftProvidersFile(t, `not json`)

	t.Run("resolved falls back on availability then verifies", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary:   {stderr: draftDNSStderr},
			draftHTTPSAlternate: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary, draftHTTPSAlternate))
		snapshot, err := newDeps(tool, true, policy, anonymous).acquireDraftNetwork(context.Background(), tool,
			draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 2 {
			t.Fatalf("%d fetches, want 2:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("resolved succeeds on the first endpoint with one fetch", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary))
		snapshot, err := newDeps(tool, true, policy, anonymous).acquireDraftNetwork(context.Background(), tool,
			draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("resolved refuses a fail-closed class with one fetch", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary:   {stderr: draftTLSStderr},
			draftHTTPSAlternate: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary, draftHTTPSAlternate))
		_, err := newDeps(tool, true, policy, anonymous).acquireDraftNetwork(context.Background(), tool,
			draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
		if buildrepo.ErrorCode(err) != buildrepo.CodeSourceUnavailable {
			t.Fatalf("err = %v, want the legacy lane diagnostic", err)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("switch on without a policy runs the legacy lane", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		missing := filepath.Join(t.TempDir(), "absent-policy.json")
		snapshot, err := newDeps(tool, true, missing, brokenProviders).acquireDraftNetwork(context.Background(), tool,
			draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("switch off ignores a present policy", func(t *testing.T) {
		for _, payload := range []string{
			draftPolicyEntry(draftHTTPSAlternate, draftHTTPSPrimary),
			`not json`,
		} {
			tool, logPath := draftGitTool(t, map[string]draftGitArm{
				draftHTTPSPrimary:   {succeed: true, fileRepo: bare},
				draftHTTPSAlternate: {stderr: draftDNSStderr},
			})
			policy := draftPolicyFile(t, payload)
			snapshot, err := newDeps(tool, false, policy, brokenProviders).acquireDraftNetwork(context.Background(), tool,
				draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
			if err != nil {
				t.Fatalf("payload %q: %v", payload, err)
			}
			if snapshot.Commit != commit {
				t.Fatalf("payload %q: commit = %s, want %s", payload, snapshot.Commit, commit)
			}
			fetches := draftFetchLines(t, logPath)
			if len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftHTTPSPrimary+">") {
				t.Fatalf("payload %q: fetches = %q, want exactly the declared URL", payload, fetches)
			}
		}
	})

	t.Run("invalid policy fails before any fetch", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftHTTPSPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, `{"schema_version":1}`)
		_, err := newDeps(tool, true, policy, anonymous).acquireDraftNetwork(context.Background(), tool,
			draftHTTPSPrimary, "https", "fixture.test/repository", lock, "", "", "")
		if err == nil || !strings.Contains(err.Error(), config.CodeRepositoryPolicyInvalid) {
			t.Fatalf("err = %v, want %s", err, config.CodeRepositoryPolicyInvalid)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})

	t.Run("resolved ssh fetches behind the manager wrapper", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftSSHPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftSSHPrimary))
		ssh := draftSSHCredentials(t)
		tool.SSHCredentials = ssh
		providers := draftProvidersFile(t, draftSSHProvidersDoc(ssh))
		snapshot, err := newDeps(tool, true, policy, providers).acquireDraftNetwork(context.Background(), tool,
			draftSSHPrimary, "ssh", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		fetches := draftFetchLines(t, logPath)
		if len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if !strings.Contains(fetches[0], buildrepo.SSHWrapperName) {
			t.Fatalf("resolved ssh fetch kept the static wrapper:\n%s", fetches[0])
		}
	})

	t.Run("legacy ssh keeps the static wrapper", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftSSHPrimary: {succeed: true, fileRepo: bare},
		})
		policy := draftPolicyFile(t, draftPolicyEntry(draftSSHPrimary))
		ssh := draftSSHCredentials(t)
		// Production binds the repository selection into the tool before the
		// closure runs; the legacy lane admits only a bound selection.
		tool.SSHCredentials = ssh
		snapshot, err := newDeps(tool, false, policy, brokenProviders).acquireDraftNetwork(context.Background(), tool,
			draftSSHPrimary, "ssh", "fixture.test/repository", lock, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.Commit != commit {
			t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
		}
		fetches := draftFetchLines(t, logPath)
		if len(fetches) != 1 {
			t.Fatalf("%d fetches, want 1:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
		if strings.Contains(fetches[0], buildrepo.SSHWrapperName) {
			t.Fatalf("legacy ssh fetch used the manager wrapper:\n%s", fetches[0])
		}
		if !strings.Contains(fetches[0], tool.SSHWrapper) {
			t.Fatalf("legacy ssh fetch lost the static wrapper:\n%s", fetches[0])
		}
	})

	t.Run("resolved ssh without an ssh executable refuses without traffic", func(t *testing.T) {
		tool, logPath := draftGitTool(t, map[string]draftGitArm{
			draftSSHPrimary: {succeed: true, fileRepo: bare},
		})
		t.Setenv("PATH", t.TempDir())
		policy := draftPolicyFile(t, draftPolicyEntry(draftSSHPrimary))
		ssh := draftSSHCredentials(t)
		providers := draftProvidersFile(t, draftSSHProvidersDoc(ssh))
		_, err := newDeps(tool, true, policy, providers).acquireDraftNetwork(context.Background(), tool,
			draftSSHPrimary, "ssh", "fixture.test/repository", lock, "", "", "")
		if buildrepo.ErrorCode(err) != buildrepo.CodeIdentityInvalid || !strings.Contains(err.Error(), "SSH wrapper policy") {
			t.Fatalf("err = %v, want the SSH base refusal", err)
		}
		if fetches := draftFetchLines(t, logPath); len(fetches) != 0 {
			t.Fatalf("%d fetches, want 0:\n%s", len(fetches), strings.Join(fetches, "\n"))
		}
	})
}

// TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess runs on Windows
// only and proves the production caller reaches the resolved executor
// there: the switch is on and the policy is valid, yet acquisition refuses
// with the lane's typed platform diagnostic. The tool cannot run, so a
// refusal that came after any spawn would surface a different code. The
// CLI pipeline masks every acquisition failure as the lane diagnostic, so
// the typed code is asserted here, at the caller.
func TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows refusal probe is exercised on Windows")
	}
	policy := draftPolicyFile(t, draftPolicyEntry(draftHTTPSPrimary))
	// A tool that cannot run: any process spawn fails loudly, proving the
	// refusal precedes every Git call.
	tool := buildrepo.GitTool{Executable: filepath.Join(t.TempDir(), "no-such-git.exe")}
	missing := filepath.Join(t.TempDir(), "absent-providers.json")
	deps := ExternalDeps{DraftTransportResolution: true, DraftPolicyPath: policy, DraftProvidersPath: missing}
	_, err := deps.acquireDraftNetwork(context.Background(), tool,
		draftHTTPSPrimary, "https", "fixture.test/repository",
		buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, "", "", "")
	if buildrepo.ErrorCode(err) != buildrepo.CodeTransportResolutionUnsupportedPlatform {
		t.Fatalf("err = %v, want %s", err, buildrepo.CodeTransportResolutionUnsupportedPlatform)
	}
}

// TestDefaultAcquireFetchesTheDeclaredURL pins the repository URL threading:
// the default closure fetches ExternalSource.GitURL, never the configured
// repository name in Declared.Repository.
func TestDefaultAcquireFetchesTheDeclaredURL(t *testing.T) {
	bare, commit := draftBareFixture(t)
	tool, logPath := draftGitTool(t, map[string]draftGitArm{
		draftHTTPSPrimary: {succeed: true, fileRepo: bare},
	})
	deps := ExternalDeps{GitTool: tool,
		Audit: func(context.Context, buildrepo.AuditSubject) error { return nil }}
	source := ExternalSource{Skill: "skill-a", Repository: "tools", GitURL: draftHTTPSPrimary,
		Declared:  buildrepo.DeclaredState{Repository: "tools", Identity: "fixture.test/repository", Transport: "https", ObjectFormat: "sha1", Commit: commit},
		Effective: buildrepo.EffectiveState{IdentityKind: "network-git", Identity: "fixture.test/repository", Transport: "https", ObjectFormat: "sha1", Commit: commit}}
	request, err := externalPipelineRequest(deps, source, buildrepo.OperatorSSHCredentials{}, BuildHTTPSCredentials{},
		skillspec.Command{Name: "tool", Target: "tool"}, nil, nil, buildrepo.OperationDryRun, NewPortableBuildAuthority())
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := request.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Commit != commit {
		t.Fatalf("commit = %s, want %s", snapshot.Commit, commit)
	}
	fetches := draftFetchLines(t, logPath)
	if len(fetches) != 1 || !strings.Contains(fetches[0], "<"+draftHTTPSPrimary+">") {
		t.Fatalf("fetches = %q, want exactly the declared URL", fetches)
	}
}
