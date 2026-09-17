package buildrepo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// testFakeSSHBase names the test-binary copies that stand in for the SSH
// executable in wrapper tests. A "-fail<N>" suffix exits N instead of
// echoing argv, proving exit-code mapping.
const testFakeSSHBase = "curator-test-fake-ssh"

func isTestFakeSSH(argv0 string) bool {
	name := filepath.Base(argv0)
	if runtime.GOOS == "windows" {
		name = strings.TrimSuffix(strings.ToLower(name), ".exe")
	}
	return name == testFakeSSHBase || strings.HasPrefix(name, testFakeSSHBase+"-fail")
}

func testFakeSSHMain(argv []string) int {
	name := filepath.Base(argv[0])
	if runtime.GOOS == "windows" {
		name = strings.TrimSuffix(strings.ToLower(name), ".exe")
	}
	if rest, ok := strings.CutPrefix(name, testFakeSSHBase+"-fail"); ok {
		code, err := strconv.Atoi(rest)
		if err != nil {
			return 1
		}
		return code
	}
	for _, arg := range argv {
		fmt.Println(arg)
	}
	return 0
}

func testFakeSSH(t *testing.T, suffix string) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	name := testFakeSSHBase + suffix
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	destination := filepath.Join(t.TempDir(), name)
	if err := copyBrokerExecutable(executable, destination); err != nil {
		t.Fatal(err)
	}
	return destination
}

func testSSHBase(t *testing.T, ssh string) SSHWrapperBase {
	t.Helper()
	dir := t.TempDir()
	write := func(name string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	return SSHWrapperBase{SSH: ssh, EmptyConfig: write("ssh.config"), EmptyKnownHosts: write("empty_known_hosts"), ConnectTimeout: 15}
}

func testSSHMaterial(t *testing.T) (identity, knownHosts string) {
	t.Helper()
	identity = filepath.Join(t.TempDir(), "id_example")
	knownHosts = filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(identity, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(knownHosts, []byte("hosts"), 0o600); err != nil {
		t.Fatal(err)
	}
	return identity, knownHosts
}

func TestBindSSHWrapperPolicy(t *testing.T) {
	identity, knownHosts := testSSHMaterial(t)
	resolve := func(path string) string {
		t.Helper()
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			t.Fatal(err)
		}
		return resolved
	}
	identity, knownHosts = resolve(identity), resolve(knownHosts)
	ssh := testFakeSSH(t, "")

	t.Run("scp-like endpoint", func(t *testing.T) {
		source, err := ParseSource("git@example.test:skills/tool.git")
		if err != nil {
			t.Fatal(err)
		}
		tool := GitTool{
			SSHCredentials: OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts},
			SSHBase:        testSSHBase(t, ssh),
		}
		policy, err := bindSSHWrapperPolicy(tool, source)
		if err != nil {
			t.Fatal(err)
		}
		if policy.ExpectedHost != "git@example.test" || policy.RepositoryPath != "skills/tool.git" {
			t.Fatalf("endpoint = %q %q", policy.ExpectedHost, policy.RepositoryPath)
		}
		if policy.Identity != identity || policy.KnownHosts != knownHosts || policy.SSH != ssh || policy.ConnectTimeout != 15 {
			t.Fatalf("policy = %+v", policy)
		}
	})

	t.Run("ssh url keeps its leading slash", func(t *testing.T) {
		source, err := ParseSource("ssh://git@example.test/skills/tool.git")
		if err != nil {
			t.Fatal(err)
		}
		tool := GitTool{
			SSHCredentials: OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts},
			SSHBase:        testSSHBase(t, ssh),
		}
		policy, err := bindSSHWrapperPolicy(tool, source)
		if err != nil {
			t.Fatal(err)
		}
		if policy.RepositoryPath != "/skills/tool.git" {
			t.Fatalf("repository path = %q", policy.RepositoryPath)
		}
	})

	source, err := ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	valid := GitTool{
		SSHCredentials: OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts},
		SSHBase:        testSSHBase(t, ssh),
	}
	for _, testCase := range []struct {
		name string
		tool GitTool
		code string
	}{
		{"zero base", GitTool{SSHCredentials: valid.SSHCredentials}, CodeIdentityInvalid},
		{"relative ssh", GitTool{SSHCredentials: valid.SSHCredentials, SSHBase: SSHWrapperBase{SSH: "ssh", EmptyConfig: valid.SSHBase.EmptyConfig, EmptyKnownHosts: valid.SSHBase.EmptyKnownHosts, ConnectTimeout: 15}}, CodeIdentityInvalid},
		{"dangling ssh", GitTool{SSHCredentials: valid.SSHCredentials, SSHBase: SSHWrapperBase{SSH: filepath.Join(t.TempDir(), "absent"), EmptyConfig: valid.SSHBase.EmptyConfig, EmptyKnownHosts: valid.SSHBase.EmptyKnownHosts, ConnectTimeout: 15}}, CodeIdentityInvalid},
		{"directory ssh", GitTool{SSHCredentials: valid.SSHCredentials, SSHBase: SSHWrapperBase{SSH: t.TempDir(), EmptyConfig: valid.SSHBase.EmptyConfig, EmptyKnownHosts: valid.SSHBase.EmptyKnownHosts, ConnectTimeout: 15}}, CodeIdentityInvalid},
		{"zero timeout", GitTool{SSHCredentials: valid.SSHCredentials, SSHBase: SSHWrapperBase{SSH: ssh, EmptyConfig: valid.SSHBase.EmptyConfig, EmptyKnownHosts: valid.SSHBase.EmptyKnownHosts}}, CodeIdentityInvalid},
		{"unselected credentials", GitTool{SSHBase: valid.SSHBase}, CodeSSHCredentialMissing},
		{"known hosts alone select nothing", GitTool{SSHCredentials: OperatorSSHCredentials{KnownHosts: knownHosts}, SSHBase: valid.SSHBase}, CodeSSHCredentialMissing},
		{"identity without pinned host keys", GitTool{SSHCredentials: OperatorSSHCredentials{Identity: identity}, SSHBase: valid.SSHBase}, CodeSSHCredentialMissing},
		{"dangling identity", GitTool{SSHCredentials: OperatorSSHCredentials{Identity: filepath.Join(t.TempDir(), "absent"), KnownHosts: knownHosts}, SSHBase: valid.SSHBase}, CodeIdentityInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := bindSSHWrapperPolicy(testCase.tool, source); ErrorCode(err) != testCase.code {
				t.Fatalf("error = %v, want %s", err, testCase.code)
			}
		})
	}
}

func TestMaterializeSSHWrapperRoundTrip(t *testing.T) {
	identity, knownHosts := testSSHMaterial(t)
	source, err := ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	tool := GitTool{
		SSHCredentials: OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts},
		SSHBase:        testSSHBase(t, testFakeSSH(t, "")),
	}
	policy, err := bindSSHWrapperPolicy(tool, source)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	wrapper, state, err := materializeSSHWrapper(t.TempDir(), executable, policy)
	if err != nil {
		t.Fatal(err)
	}
	if !IsSSHWrapperInvocation(wrapper) {
		t.Fatalf("materialized wrapper %q does not dispatch", wrapper)
	}
	loaded, ok := readSSHWrapperState(state)
	if !ok {
		t.Fatal("materialized state does not read back")
	}
	if loaded != policyWithWrapper(policy, wrapper) {
		t.Fatalf("state = %+v", loaded)
	}
	// The state carries the policy shape only: paths, endpoint, and
	// timeout, with no secret field to smuggle one through.
	payload, err := os.ReadFile(state)
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]any
	if err := json.Unmarshal(payload, &keys); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"Wrapper": true, "SSH": true, "ExpectedHost": true, "RepositoryPath": true,
		"EmptyConfig": true, "KnownHosts": true, "EmptyKnownHosts": true, "Identity": true,
		"AgentSocket": true, "ConnectTimeout": true}
	if len(keys) != len(want) {
		t.Fatalf("state keys = %v", keys)
	}
	for key := range keys {
		if !want[key] {
			t.Fatalf("state carries %q", key)
		}
	}
	if _, _, err := materializeSSHWrapper(t.TempDir(), "relative/curator", policy); ErrorCode(err) != CodeIdentityInvalid {
		t.Fatalf("relative wrapper source error = %v, want %s", err, CodeIdentityInvalid)
	}
}

func policyWithWrapper(policy SSHPolicy, wrapper string) SSHPolicy {
	policy.Wrapper = wrapper
	return policy
}

func TestRunSSHWrapperAnswersOnlyTheBoundTuple(t *testing.T) {
	identity, knownHosts := testSSHMaterial(t)
	resolve := func(path string) string {
		t.Helper()
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			t.Fatal(err)
		}
		return resolved
	}
	identity, knownHosts = resolve(identity), resolve(knownHosts)
	source, err := ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	ssh := testFakeSSH(t, "")
	tool := GitTool{
		SSHCredentials: OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts},
		SSHBase:        testSSHBase(t, ssh),
	}
	policy, err := bindSSHWrapperPolicy(tool, source)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, state, err := materializeSSHWrapper(t.TempDir(), executable, policy)
	if err != nil {
		t.Fatal(err)
	}
	getenv := func(name string) string {
		if name == EnvSSHWrapperState {
			return state
		}
		return ""
	}
	tuple := []string{"git@example.test", "git-upload-pack 'skills/tool.git'"}

	t.Run("pinned argv reaches ssh", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if code := RunSSHWrapper(tuple, getenv, strings.NewReader("stdin-probe"), &stdout, &stderr); code != 0 {
			t.Fatalf("code = %d, stderr = %q", code, stderr.String())
		}
		argv := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		joined := strings.Join(argv, " ")
		// The fake echoes exactly what the wrapper execed: the fixed
		// argv, the pinned material, and the bound tuple.
		for _, marker := range []string{ssh, "-F", "BatchMode=yes", "StrictHostKeyChecking=yes",
			"PasswordAuthentication=no", "ProxyCommand=none", "IdentitiesOnly=yes", "IdentityAgent=none",
			"-i " + identity, "UserKnownHostsFile=" + knownHosts, "git@example.test", tuple[1]} {
			if !strings.Contains(joined, marker) {
				t.Errorf("ssh argv lacks %q:\n%s", marker, joined)
			}
		}
		if argv[0] != ssh || argv[len(argv)-2] != "git@example.test" || argv[len(argv)-1] != tuple[1] {
			t.Errorf("ssh argv endpoints wrong:\n%s", joined)
		}
	})

	t.Run("exit code maps through", func(t *testing.T) {
		failing := policy
		failing.SSH = testFakeSSH(t, "-fail3")
		_, failingState, err := materializeSSHWrapper(t.TempDir(), executable, failing)
		if err != nil {
			t.Fatal(err)
		}
		failingGetenv := func(name string) string {
			if name == EnvSSHWrapperState {
				return failingState
			}
			return ""
		}
		var stdout, stderr bytes.Buffer
		if code := RunSSHWrapper(tuple, failingGetenv, strings.NewReader(""), &stdout, &stderr); code != 3 {
			t.Fatalf("code = %d, want 3", code)
		}
	})

	for _, testCase := range []struct {
		name string
		args []string
	}{
		{"foreign host", []string{"other@example.test", tuple[1]}},
		{"foreign path", []string{tuple[0], "git-upload-pack 'other/repo.git'"}},
		{"bare command", []string{tuple[0], "git-upload-pack skills/tool.git"}},
		{"extra argument", []string{tuple[0], tuple[1], "extra"}},
		{"missing command", []string{tuple[0]}},
		{"no arguments", nil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := RunSSHWrapper(testCase.args, getenv, strings.NewReader(""), &stdout, &stderr); code != 1 {
				t.Fatalf("code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("refused tuple produced %q", stdout.String())
			}
		})
	}

	for _, testCase := range []struct {
		name  string
		state func() string
	}{
		{"absent state", func() string { return filepath.Join(t.TempDir(), "absent") }},
		{"directory state", func() string { return t.TempDir() }},
		{"empty state variable", func() string { return "" }},
		{"garbage state", func() string { return writeSSHStateFile(t, "not json") }},
		{"unknown field", func() string {
			return writeSSHStateFile(t, `{"Wrapper":"/abs/w","SSH":"/abs/ssh","ExpectedHost":"h","RepositoryPath":"r","EmptyConfig":"/abs/c","KnownHosts":"/abs/k","EmptyKnownHosts":"/abs/e","Identity":"/abs/i","AgentSocket":"","ConnectTimeout":15,"Secret":"x"}`)
		}},
		{"relative ssh", func() string {
			mutated := policy
			mutated.SSH = "ssh"
			return writeSSHStatePolicy(t, mutated)
		}},
		{"empty endpoint", func() string {
			mutated := policy
			mutated.ExpectedHost = ""
			return writeSSHStatePolicy(t, mutated)
		}},
		{"zero timeout", func() string {
			mutated := policy
			mutated.ConnectTimeout = 0
			return writeSSHStatePolicy(t, mutated)
		}},
		{"foreign wrapper name", func() string {
			mutated := policy
			mutated.Wrapper = filepath.Join(t.TempDir(), "something-else")
			return writeSSHStatePolicy(t, mutated)
		}},
		{"control character in path", func() string {
			mutated := policy
			mutated.Identity = "/abs/id\ninjected"
			return writeSSHStatePolicy(t, mutated)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			broken := func(name string) string {
				if name == EnvSSHWrapperState {
					return testCase.state()
				}
				return ""
			}
			var stdout, stderr bytes.Buffer
			if code := RunSSHWrapper(tuple, broken, strings.NewReader(""), &stdout, &stderr); code != 1 {
				t.Fatalf("code = %d, want 1", code)
			}
		})
	}
}

func writeSSHStateFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ssh-wrapper-state.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeSSHStatePolicy(t *testing.T, policy SSHPolicy) string {
	t.Helper()
	payload, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	return writeSSHStateFile(t, string(payload)+"\n")
}

func TestIsSSHWrapperInvocation(t *testing.T) {
	if !IsSSHWrapperInvocation(filepath.Join("manager-wrappers", SSHWrapperName)) {
		t.Fatal("private wrapper copy does not dispatch")
	}
	abs, err := filepath.Abs(filepath.Join("manager-wrappers", SSHWrapperName))
	if err != nil {
		t.Fatal(err)
	}
	if !IsSSHWrapperInvocation(abs) {
		t.Fatal("absolute wrapper copy does not dispatch")
	}
	for _, argv0 := range []string{HTTPSBrokerName, "ssh", "curator", SSHWrapperName + ".bak", "x" + SSHWrapperName} {
		if IsSSHWrapperInvocation(argv0) {
			t.Fatalf("%q dispatches as the SSH wrapper", argv0)
		}
	}
}
