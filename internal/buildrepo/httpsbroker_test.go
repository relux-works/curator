package buildrepo

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	legacyHTTPSBrokerSecretEnv = "CURATOR_BUILD_HTTPS_ASKPASS_SECRET"
	testBrokerSpawnGrandchild  = "CURATOR_TEST_HTTPS_BROKER_SPAWN_GRANDCHILD"
	testBrokerReportPath       = "CURATOR_TEST_HTTPS_BROKER_REPORT"
	testBrokerGrandchild       = "CURATOR_TEST_HTTPS_BROKER_GRANDCHILD"
	testBrokerFetchChild       = "CURATOR_TEST_HTTPS_BROKER_FETCH_CHILD"
	testExpectedBrokerSecret   = "fetch-only-secret"
)

func TestMain(m *testing.M) {
	if os.Getenv(testBrokerGrandchild) == "1" {
		code := 0
		presence := "absent"
		if _, ok := os.LookupEnv(legacyHTTPSBrokerSecretEnv); ok {
			presence = "present"
		}
		if err := os.WriteFile(os.Getenv(testBrokerReportPath), []byte(presence), 0o600); err != nil {
			code = 2
		}
		os.Exit(code)
	}
	if os.Getenv(testBrokerFetchChild) == "1" {
		os.Exit(runHTTPSBrokerFetchChild())
	}
	if IsHTTPSBrokerInvocation(os.Args[0]) {
		var output bytes.Buffer
		code := RunHTTPSCredentialBroker(os.Args[1:], os.Getenv, &output)
		if reportPath := os.Getenv(testBrokerReportPath); reportPath != "" {
			presence := "absent"
			if _, ok := os.LookupEnv(legacyHTTPSBrokerSecretEnv); ok {
				presence = "present"
			}
			answer := "wrong"
			if output.String() == testExpectedBrokerSecret+"\n" {
				answer = "received"
			}
			if err := os.WriteFile(reportPath+".helper", []byte(presence), 0o600); err != nil {
				code = 1
			}
			if err := os.WriteFile(reportPath+".answer", []byte(answer), 0o600); err != nil {
				code = 1
			}
			if os.Getenv(testBrokerSpawnGrandchild) == "1" {
				command := exec.Command(os.Args[0])
				command.Env = setEnvironmentValue(os.Environ(), testBrokerGrandchild, "1")
				command.Env = setEnvironmentValue(command.Env, testBrokerReportPath, reportPath+".grandchild")
				if err := command.Run(); err != nil {
					code = 1
				}
			}
		}
		_, _ = os.Stdout.Write(output.Bytes())
		os.Exit(code)
	}
	if IsSSHWrapperInvocation(os.Args[0]) {
		os.Exit(RunSSHWrapper(os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr))
	}
	if isTestFakeSSH(os.Args[0]) {
		os.Exit(testFakeSSHMain(os.Args))
	}
	os.Exit(m.Run())
}

func runHTTPSBrokerWithTestTransport(t *testing.T, statePath, secret string, args []string, mutate func(map[string]string)) (int, string) {
	t.Helper()
	command := exec.Command(os.Args[0])
	transport, err := NewHTTPSBrokerSecretTransport(command, secret)
	if err != nil {
		t.Fatal(err)
	}
	environment := map[string]string{
		EnvHTTPSBrokerState:     statePath,
		EnvHTTPSBrokerTransport: testHTTPSBrokerTransportValue(transport),
	}
	if mutate != nil {
		mutate(environment)
	}
	serveCtx, cancelServe := context.WithCancel(context.Background())
	serveDone := make(chan error, 1)
	go func() { serveDone <- transport.Serve(serveCtx) }()
	var output bytes.Buffer
	code := RunHTTPSCredentialBroker(args, func(name string) string { return environment[name] }, &output)
	cancelServe()
	serveErr := <-serveDone
	closeErr := transport.Close()
	if serveErr != nil || closeErr != nil {
		t.Fatalf("test HTTPS broker transport: %v", errors.Join(serveErr, closeErr))
	}
	return code, output.String()
}

func TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts(t *testing.T) {
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	credentials := NewHTTPSCredentials("git.example.test", "oauth2", "broker-secret")
	wrapper, statePath, err := materializeHTTPSCredentialBroker(root, executable, credentials)
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name   string
		args   []string
		mutate func(map[string]string)
		want   string
		code   int
	}{
		{name: "username", args: []string{"Username for 'https://git.example.test': "}, want: "oauth2\n"},
		{name: "password", args: []string{"Password for 'https://oauth2@git.example.test': "}, want: "broker-secret\n"},
		{name: "foreign host", args: []string{"Username for 'https://other.example.test': "}, code: 1},
		{name: "foreign prompt", args: []string{"Token for 'https://git.example.test': "}, code: 1},
		{name: "extra argument", args: []string{"Username for 'https://git.example.test': ", "extra"}, code: 1},
		{name: "absent secret transport", args: []string{"Username for 'https://git.example.test': "}, mutate: func(environment map[string]string) { delete(environment, EnvHTTPSBrokerTransport) }, code: 1},
		{name: "malformed secret transport", args: []string{"Username for 'https://git.example.test': "}, mutate: func(environment map[string]string) { environment[EnvHTTPSBrokerTransport] = "not-a-transport" }, code: 1},
		{name: "absent state", args: []string{"Username for 'https://git.example.test': "}, mutate: func(environment map[string]string) { environment[EnvHTTPSBrokerState] = filepath.Join(root, "absent") }, code: 1},
		{name: "unreadable state shape", args: []string{"Username for 'https://git.example.test': "}, mutate: func(environment map[string]string) { environment[EnvHTTPSBrokerState] = root }, code: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			code, output := runHTTPSBrokerWithTestTransport(t, statePath, "broker-secret", testCase.args, testCase.mutate)
			if code != testCase.code || output != testCase.want {
				t.Fatalf("code=%d output=%q, want code=%d output=%q", code, output, testCase.code, testCase.want)
			}
		})
	}
	assertBrokerExecutableReleased(t, wrapper)
}

func TestHTTPSCredentialBrokerRejectsEmptyPasswordPipe(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, statePath, err := materializeHTTPSCredentialBroker(t.TempDir(), executable,
		NewHTTPSCredentials("git.example.test", "oauth2", "must-not-be-used"))
	if err != nil {
		t.Fatal(err)
	}
	code, output := runHTTPSBrokerWithTestTransport(t, statePath, "", []string{"Password for 'https://oauth2@git.example.test': "}, nil)
	if code != 1 || output != "" {
		t.Fatalf("empty secret pipe returned code=%d output=%q, want fail-closed exit 1 and empty output", code, output)
	}
}

func TestHTTPSBrokerStateContainsHostAndUsernameOnly(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	const secret = "must-never-reach-state"
	wrapper, statePath, err := materializeHTTPSCredentialBroker(t.TempDir(), executable,
		NewHTTPSCredentials("git.example.test", "oauth2", secret))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(wrapper) != filepath.Dir(statePath) {
		t.Fatalf("wrapper %q and state %q are not manager-owned siblings", wrapper, statePath)
	}
	payload, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), secret) || string(payload) != "{\"host\":\"git.example.test\",\"username\":\"oauth2\"}\n" {
		t.Fatal("HTTPS broker state does not contain only host and username")
	}
	for _, diagnostic := range []string{
		fmt.Sprintf("%v", NewHTTPSCredentials("git.example.test", "oauth2", secret)),
		fmt.Sprintf("%+v", NewHTTPSCredentials("git.example.test", "oauth2", secret)),
		fmt.Sprintf("%#v", NewHTTPSCredentials("git.example.test", "oauth2", secret)),
	} {
		if strings.Contains(diagnostic, secret) || !strings.Contains(diagnostic, "<redacted>") {
			t.Fatalf("credential diagnostic = %q", diagnostic)
		}
	}
	assertBrokerExecutableReleased(t, wrapper)
}

func TestHTTPSBrokerPipeSurvivesFetchAndAskpassExecOnEveryPlatform(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	reportPath := filepath.Join(root, "askpass-report")
	wrapper, statePath, err := materializeHTTPSCredentialBroker(root, executable,
		NewHTTPSCredentials("fixture.test", "oauth2", testExpectedBrokerSecret))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable)
	command.Env = append(cleanDiscoveryEnvironment(),
		testBrokerFetchChild+"=1",
		testBrokerReportPath+"="+reportPath,
		testBrokerSpawnGrandchild+"=1",
		"GIT_ASKPASS="+wrapper,
		EnvHTTPSBrokerState+"="+statePath,
	)
	if err := runCommandWithHTTPSSecret(command, testExpectedBrokerSecret); err != nil {
		t.Fatalf("fetch and askpass pipe flow: %v", err)
	}
	for _, report := range []struct{ suffix, want string }{
		{suffix: ".fetch", want: "absent"},
		{suffix: ".answer", want: "received"},
		{suffix: ".helper", want: "absent"},
		{suffix: ".grandchild", want: "absent"},
	} {
		contents, err := os.ReadFile(reportPath + report.suffix)
		if err != nil {
			t.Fatalf("read askpass %s report: %v", report.suffix, err)
		}
		if string(contents) != report.want {
			t.Fatalf("askpass %s report = %q, want %q", report.suffix, contents, report.want)
		}
	}
	assertBrokerExecutableReleased(t, wrapper)
}

func assertBrokerExecutableReleased(t *testing.T, wrapper string) {
	t.Helper()
	if err := os.Remove(wrapper); err != nil {
		t.Fatalf("remove materialized HTTPS broker before TempDir cleanup: %v", err)
	}
}

func TestPrivateHTTPSBrokerAuthenticatesRealGitRepository(t *testing.T) {
	fixture := makeGitFixture(t, "sha1", false)
	runTestGit(t, fixture.bare, realGitPath(t), "update-server-info")

	const username, secret = "oauth2", testExpectedBrokerSecret
	reportPath := filepath.Join(t.TempDir(), "private-https-askpass-report")
	var authenticated int
	files := http.FileServer(http.Dir(fixture.bare))
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user, password, ok := request.BasicAuth()
		if !ok || user != username || password != secret {
			writer.Header().Set("WWW-Authenticate", `Basic realm="curator-test"`)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		authenticated++
		http.StripPrefix("/repository.git", files).ServeHTTP(writer, request)
	}))
	defer server.Close()

	certificate := filepath.Join(t.TempDir(), "tls-ca.pem")
	der := server.Certificate().Raw
	if _, err := x509.ParseCertificate(der); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(certificate, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	host := strings.TrimPrefix(server.URL, "https://")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrapper, statePath, err := materializeHTTPSCredentialBroker(t.TempDir(), executable,
		NewHTTPSCredentials(host, username, secret))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	gitArgs := []string{
		"-c", "credential.helper=",
		"-c", "core.askPass=" + wrapper,
		"-c", "http.sslVerify=true",
	}
	if runtime.GOOS == "windows" {
		// Git for Windows defaults to schannel, which ignores http.sslCAInfo;
		// the test-scoped CA is only honoured by the openssl backend.
		gitArgs = append(gitArgs, "-c", "http.sslBackend=openssl")
	}
	gitArgs = append(gitArgs,
		"-c", "http.sslCAInfo="+certificate,
		"-c", "http.followRedirects=false",
		"ls-remote", "--exit-code", server.URL+"/repository.git")
	cmd := exec.CommandContext(ctx, realGitPath(t), gitArgs...)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS="+wrapper,
		EnvHTTPSBrokerState+"="+statePath)
	cmd.Env = append(cmd.Env,
		testBrokerSpawnGrandchild+"=1",
		testBrokerReportPath+"="+reportPath,
	)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err = runCommandWithHTTPSSecret(cmd, secret)
	if err != nil {
		t.Fatalf("private HTTPS ls-remote: %v\n%s", err, output.String())
	}
	if authenticated == 0 || !strings.Contains(output.String(), fixture.commit) {
		t.Fatalf("authenticated requests=%d output=%s", authenticated, output.String())
	}
	for _, entry := range cmd.Env {
		if strings.Contains(entry, secret) || strings.HasPrefix(entry, legacyHTTPSBrokerSecretEnv+"=") {
			t.Fatalf("fetch environment carries HTTPS secret material in %q", strings.SplitN(entry, "=", 2)[0])
		}
	}
	for _, report := range []struct{ suffix, want string }{
		{suffix: ".helper", want: "absent"},
		{suffix: ".answer", want: "received"},
		{suffix: ".grandchild", want: "absent"},
	} {
		contents, err := os.ReadFile(reportPath + report.suffix)
		if err != nil {
			t.Fatalf("read real-git askpass %s report: %v", report.suffix, err)
		}
		if string(contents) != report.want {
			t.Fatalf("real-git askpass %s report = %q, want %q", report.suffix, contents, report.want)
		}
	}
	assertBrokerExecutableReleased(t, wrapper)
}

func TestSelectedHTTPSFetchSecretUsesPipeAndNeverEntersTheProcessEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the HTTPS test transport wrapper is POSIX-only")
	}
	fixture := makeGitFixture(t, "sha1", false)
	reportPath := filepath.Join(t.TempDir(), "https-askpass-report")
	fetchHook := fmt.Sprintf(`fetch_command=0
for arg in "$@"; do [ "$arg" = "fetch" ] && fetch_command=1; done
if [ "$fetch_command" = "1" ]; then
  export %s=1
  export %s=%s
  "$GIT_ASKPASS" "Password for 'https://oauth2@fixture.test': " >/dev/null || exit 93
fi`, testBrokerSpawnGrandchild, testBrokerReportPath, shellQuote(reportPath))
	tool, logPath := fakeHTTPGitToolWithFetchHook(t, fixture.bare, fetchHook)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	tool.AskPass = executable
	tool.HTTPSCredentials = NewHTTPSCredentials("fixture.test", "oauth2", "fetch-only-secret")
	source, err := ParseSource("https://fixture.test/repository.git")
	if err != nil {
		t.Fatal(err)
	}
	_, acquireErr := AcquireNetwork(context.Background(), NetworkRequest{
		Source: source,
		Lock:   LockedCommit{ObjectFormat: "sha1", Hex: fixture.commit},
		Tool:   tool,
	})
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if acquireErr != nil {
		t.Errorf("authenticated production fetch failed: %v", acquireErr)
	}
	var fetchLine string
	for _, line := range strings.Split(string(log), "\n") {
		if strings.Contains(line, " fetch ") {
			fetchLine = line
			break
		}
	}
	if fetchLine == "" || !strings.Contains(fetchLine, "core.askPass=") {
		t.Errorf("fetch argv did not set core.askPass: %s", log)
	}
	for _, required := range []string{
		"http.sslVerify=true", "http.followRedirects=false", "https://fixture.test/repository.git",
	} {
		if !strings.Contains(fetchLine, required) {
			t.Errorf("fetch lost hardened argument %q: %s", required, fetchLine)
		}
	}
	if !strings.Contains(fetchLine, "secret=0 state=1 handle=1") || !strings.Contains(fetchLine, "askpass=") || !strings.Contains(fetchLine, HTTPSBrokerName) {
		t.Errorf("fetch environment must carry metadata and a pipe handle, never the secret: %s", fetchLine)
	}
	var environmentAskPass, configuredAskPass string
	for _, field := range strings.Fields(fetchLine) {
		if value, ok := strings.CutPrefix(field, "askpass="); ok {
			environmentAskPass = value
		}
		if value, ok := strings.CutPrefix(field, "core.askPass="); ok {
			configuredAskPass = value
		}
	}
	if environmentAskPass == "" || configuredAskPass != environmentAskPass {
		t.Errorf("GIT_ASKPASS=%q core.askPass=%q, want the same wrapper", environmentAskPass, configuredAskPass)
	}
	for _, line := range strings.Split(string(log), "\n") {
		if line == "" || strings.Contains(line, " fetch ") {
			continue
		}
		if strings.Contains(line, "secret=1") || strings.Contains(line, "state=1") || strings.Contains(line, "handle=1") {
			t.Errorf("non-fetch Git child received broker material: %s", line)
		}
	}
	if strings.Contains(string(log), "fetch-only-secret") {
		t.Error("fetch diagnostics contain the HTTPS secret")
	}
	for _, report := range []struct{ suffix, want string }{
		{suffix: ".answer", want: "received"},
		{suffix: ".helper", want: "absent"},
		{suffix: ".grandchild", want: "absent"},
	} {
		contents, err := os.ReadFile(reportPath + report.suffix)
		if err != nil {
			t.Errorf("read askpass %s report: %v", report.suffix, err)
			continue
		}
		if string(contents) != report.want {
			t.Errorf("askpass %s report = %q, want %q", report.suffix, contents, report.want)
		}
	}
}

func TestAnonymousHTTPSArgumentsAndEnvironmentRemainUnchanged(t *testing.T) {
	paths := privatePaths{root: "/private", home: "/private/home", config: "/private/config", path: "/private/path"}
	tool := GitTool{ExecPath: "/git/libexec", AskPass: "/manager/askpass"}
	beforeArgs := strictFetchArgs("/repo", "/hooks", tool.AskPass, "https", "https://example.test/repo.git", "abc:refs/curator/locked")
	beforeEnv := cleanGitEnvironment(paths, tool, "https")
	if tool.HTTPSCredentials.Selected() {
		t.Fatal("zero-value credentials unexpectedly selected authentication")
	}
	afterArgs := strictFetchArgs("/repo", "/hooks", tool.AskPass, "https", "https://example.test/repo.git", "abc:refs/curator/locked")
	afterEnv := cleanGitEnvironment(paths, tool, "https")
	if strings.Join(beforeArgs, "\x00") != strings.Join(afterArgs, "\x00") || strings.Join(beforeEnv, "\x00") != strings.Join(afterEnv, "\x00") {
		t.Fatal("anonymous HTTPS argv or environment changed")
	}
	for _, entry := range afterEnv {
		if strings.HasPrefix(entry, EnvHTTPSBrokerState+"=") || strings.HasPrefix(entry, legacyHTTPSBrokerSecretEnv+"=") || strings.HasPrefix(entry, EnvHTTPSBrokerTransport+"=") {
			t.Fatalf("anonymous environment contains broker material: %q", entry)
		}
	}
}
