package buildrepo

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
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

func TestMain(m *testing.M) {
	if IsHTTPSBrokerInvocation(os.Args[0]) {
		os.Exit(RunHTTPSCredentialBroker(os.Args[1:], os.Getenv, os.Stdout))
	}
	os.Exit(m.Run())
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
	environment := map[string]string{
		EnvHTTPSBrokerState:  statePath,
		EnvHTTPSBrokerSecret: "broker-secret",
	}
	getenv := func(name string) string { return environment[name] }

	for _, testCase := range []struct {
		name   string
		args   []string
		mutate func()
		want   string
		code   int
	}{
		{name: "username", args: []string{"Username for 'https://git.example.test': "}, want: "oauth2\n"},
		{name: "password", args: []string{"Password for 'https://oauth2@git.example.test': "}, want: "broker-secret\n"},
		{name: "foreign host", args: []string{"Username for 'https://other.example.test': "}, code: 1},
		{name: "foreign prompt", args: []string{"Token for 'https://git.example.test': "}, code: 1},
		{name: "extra argument", args: []string{"Username for 'https://git.example.test': ", "extra"}, code: 1},
		{name: "absent secret", args: []string{"Username for 'https://git.example.test': "}, mutate: func() { environment[EnvHTTPSBrokerSecret] = "" }, code: 1},
		{name: "absent state", args: []string{"Username for 'https://git.example.test': "}, mutate: func() { environment[EnvHTTPSBrokerState] = filepath.Join(root, "absent") }, code: 1},
		{name: "unreadable state shape", args: []string{"Username for 'https://git.example.test': "}, mutate: func() { environment[EnvHTTPSBrokerState] = root }, code: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			environment[EnvHTTPSBrokerState] = statePath
			environment[EnvHTTPSBrokerSecret] = "broker-secret"
			if testCase.mutate != nil {
				testCase.mutate()
			}
			var output bytes.Buffer
			code := RunHTTPSCredentialBroker(testCase.args, getenv, &output)
			if code != testCase.code || output.String() != testCase.want {
				t.Fatalf("code=%d output=%q, want code=%d output=%q", code, output.String(), testCase.code, testCase.want)
			}
		})
	}
	assertBrokerExecutableReleased(t, wrapper)
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
		t.Fatalf("state = %q", payload)
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

func assertBrokerExecutableReleased(t *testing.T, wrapper string) {
	t.Helper()
	if err := os.Remove(wrapper); err != nil {
		t.Fatalf("remove materialized HTTPS broker before TempDir cleanup: %v", err)
	}
}

func TestPrivateHTTPSBrokerAuthenticatesRealGitRepository(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the local TLS fixture relies on the platform Git shipped on Unix CI")
	}
	fixture := makeGitFixture(t, "sha1", false)
	runTestGit(t, fixture.bare, realGitPath(t), "update-server-info")

	const username, secret = "oauth2", "private-repository-secret"
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
	cmd := exec.CommandContext(ctx, realGitPath(t),
		"-c", "credential.helper=",
		"-c", "core.askPass="+wrapper,
		"-c", "http.sslVerify=true",
		"-c", "http.sslCAInfo="+certificate,
		"-c", "http.followRedirects=false",
		"ls-remote", "--exit-code", server.URL+"/repository.git")
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS="+wrapper,
		EnvHTTPSBrokerState+"="+statePath,
		EnvHTTPSBrokerSecret+"="+secret)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("private HTTPS ls-remote: %v\n%s", err, output)
	}
	if authenticated == 0 || !strings.Contains(string(output), fixture.commit) {
		t.Fatalf("authenticated requests=%d output=%s", authenticated, output)
	}
}

func TestSelectedHTTPSFetchEnvironmentIsScopedAndOverridesBothAskPassSurfaces(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the HTTPS test transport wrapper is POSIX-only")
	}
	fixture := makeGitFixture(t, "sha1", false)
	tool, logPath := fakeHTTPGitTool(t, fixture.bare)
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
	if _, err := AcquireNetwork(context.Background(), NetworkRequest{
		Source: source,
		Lock:   LockedCommit{ObjectFormat: "sha1", Hex: fixture.commit},
		Tool:   tool,
	}); err != nil {
		t.Fatal(err)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	var fetchLine string
	for _, line := range strings.Split(string(log), "\n") {
		if strings.Contains(line, " fetch ") {
			fetchLine = line
			break
		}
	}
	if fetchLine == "" || !strings.Contains(fetchLine, "core.askPass=") {
		t.Fatalf("fetch argv did not set core.askPass: %s", log)
	}
	for _, required := range []string{
		"http.sslVerify=true", "http.followRedirects=false", "https://fixture.test/repository.git",
	} {
		if !strings.Contains(fetchLine, required) {
			t.Fatalf("fetch lost hardened argument %q: %s", required, fetchLine)
		}
	}
	if !strings.Contains(fetchLine, "secret=1 state=1") || !strings.Contains(fetchLine, "askpass=") || !strings.Contains(fetchLine, HTTPSBrokerName) {
		t.Fatalf("fetch environment did not point GIT_ASKPASS at the materialized wrapper: %s", fetchLine)
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
		t.Fatalf("GIT_ASKPASS=%q core.askPass=%q, want the same wrapper", environmentAskPass, configuredAskPass)
	}
	for _, line := range strings.Split(string(log), "\n") {
		if line == "" || strings.Contains(line, " fetch ") {
			continue
		}
		if strings.Contains(line, "secret=1") || strings.Contains(line, "state=1") {
			t.Fatalf("non-fetch Git child received broker material: %s", line)
		}
	}
	if strings.Contains(string(log), "fetch-only-secret") {
		t.Fatalf("diagnostic log disclosed the secret: %s", log)
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
		if strings.HasPrefix(entry, EnvHTTPSBrokerState+"=") || strings.HasPrefix(entry, EnvHTTPSBrokerSecret+"=") {
			t.Fatalf("anonymous environment contains broker material: %q", entry)
		}
	}
}
