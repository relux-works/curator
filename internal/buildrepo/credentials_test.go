package buildrepo

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// operatorMaterial creates the three kinds of object an operator selection may
// name and returns them already canonicalized, the way validation reports them.
func operatorMaterial(t *testing.T) (identity, agentSocket, knownHosts string) {
	t.Helper()
	root := t.TempDir()
	identity = writeOperatorFile(t, filepath.Join(root, "id_ed25519.pub"), "ssh-ed25519 AAAA operator\n")
	knownHosts = writeOperatorFile(t, filepath.Join(root, "known_hosts"), "example.test ssh-ed25519 AAAA\n")
	agentSocket = listenOperatorSocket(t, filepath.Join(shortTempDir(t), "agent.sock"))
	return identity, agentSocket, knownHosts
}

// shortTempDir is a scratch directory whose path fits a unix-domain socket
// address. The default per-test temporary directory does not on macOS, where
// it spells out the test name under a long per-user container.
func shortTempDir(t *testing.T) string {
	t.Helper()
	root, err := os.MkdirTemp("/tmp", "curator-buildrepo")
	if err != nil {
		t.Skipf("this host cannot create a short scratch directory: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })
	return root
}

func writeOperatorFile(t *testing.T, path, content string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return canonical(t, path)
}

func listenOperatorSocket(t *testing.T, path string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("an operator agent socket is a unix-domain rendezvous point")
	}
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Skipf("this host cannot create an agent socket: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	return canonical(t, path)
}

func canonical(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func managerPolicy(t *testing.T) SSHPolicy {
	t.Helper()
	root := t.TempDir()
	return SSHPolicy{
		Wrapper:         filepath.Join(root, "manager", "ssh-wrapper"),
		SSH:             filepath.Join(root, "tools", "ssh"),
		EmptyConfig:     filepath.Join(root, "manager", "ssh.config"),
		EmptyKnownHosts: filepath.Join(root, "manager", "empty_known_hosts"),
		ConnectTimeout:  15,
	}
}

// The operator's three authentication shapes have to survive the whole path
// from a selection to the argv the wrapper execs, because each one produces a
// different OpenSSH tail and only one of them is the pinned-agent form.
func TestEveryOperatorSelectionShapeReachesTheWrapperPolicy(t *testing.T) {
	identity, agentSocket, knownHosts := operatorMaterial(t)
	source, err := ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name        string
		credentials OperatorSSHCredentials
		wanted      []string
		refused     []string
	}{
		{
			name:        "identity only",
			credentials: OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts},
			wanted:      []string{"IdentitiesOnly=yes", "IdentityAgent=none", "-i " + identity},
			refused:     []string{"IdentityAgent=" + agentSocket},
		},
		{
			name:        "agent only",
			credentials: OperatorSSHCredentials{AgentSocket: agentSocket, KnownHosts: knownHosts},
			wanted:      []string{"IdentitiesOnly=no", "IdentityFile=none", "IdentityAgent=" + agentSocket},
			refused:     []string{"-i " + identity},
		},
		{
			name:        "agent pinned to one identity",
			credentials: OperatorSSHCredentials{Identity: identity, AgentSocket: agentSocket, KnownHosts: knownHosts},
			wanted:      []string{"IdentitiesOnly=yes", "IdentityAgent=" + agentSocket, "-i " + identity},
			refused:     []string{"IdentityFile=none", "IdentityAgent=none"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			policy, err := SSHPolicyFor(managerPolicy(t), source, testCase.credentials)
			if err != nil {
				t.Fatal(err)
			}
			if policy.ExpectedHost != "git@example.test" || policy.RepositoryPath != "skills/tool.git" {
				t.Fatalf("endpoint = %q %q", policy.ExpectedHost, policy.RepositoryPath)
			}
			if policy.KnownHosts != knownHosts {
				t.Fatalf("known hosts = %q, want %q", policy.KnownHosts, knownHosts)
			}
			argv := []string{policy.Wrapper, policy.ExpectedHost, "git-upload-pack 'skills/tool.git'"}
			command, err := ExactSSHCommand(policy, argv)
			if err != nil {
				t.Fatal(err)
			}
			joined := strings.Join(command, " ")
			for _, required := range testCase.wanted {
				if !strings.Contains(joined, required) {
					t.Errorf("SSH command missing %q in %s", required, joined)
				}
			}
			for _, forbidden := range testCase.refused {
				if strings.Contains(joined, forbidden) {
					t.Errorf("SSH command carries %q in %s", forbidden, joined)
				}
			}
		})
	}
}

func TestSSHEndpointKeepsTheSpellingGitHandsToSSH(t *testing.T) {
	for _, testCase := range []struct{ raw, host, path string }{
		{"git@example.test:skills/tool.git", "git@example.test", "skills/tool.git"},
		{"example.test:skills/tool", "example.test", "skills/tool"},
		{"ssh://git@example.test/skills/tool.git", "git@example.test", "/skills/tool.git"},
		{"ssh://example.test/skills/tool", "example.test", "/skills/tool"},
	} {
		source, err := ParseSource(testCase.raw)
		if err != nil {
			t.Fatalf("%s: %v", testCase.raw, err)
		}
		host, path, err := SSHEndpoint(source)
		if err != nil || host != testCase.host || path != testCase.path {
			t.Errorf("%s -> %q %q %v, want %q %q", testCase.raw, host, path, err, testCase.host, testCase.path)
		}
	}
	https, err := ParseSource("https://example.test/skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := SSHEndpoint(https); ErrorCode(err) != CodeIdentityInvalid {
		t.Fatalf("HTTPS endpoint error = %v, want %s", err, CodeIdentityInvalid)
	}
}

func TestSSHPolicyRefusesUnselectedAndUnverifiableRepositories(t *testing.T) {
	identity, _, knownHosts := operatorMaterial(t)
	source, err := ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SSHPolicyFor(managerPolicy(t), source, OperatorSSHCredentials{KnownHosts: knownHosts}); ErrorCode(err) != CodeSSHCredentialMissing {
		t.Fatalf("unselected repository error = %v, want %s", err, CodeSSHCredentialMissing)
	}
	// Known hosts alone authenticate nothing, and their absence leaves
	// StrictHostKeyChecking=yes with nothing to check against.
	if _, err := SSHPolicyFor(managerPolicy(t), source, OperatorSSHCredentials{Identity: identity}); ErrorCode(err) != CodeSSHCredentialMissing {
		t.Fatalf("hostless policy error = %v, want %s", err, CodeSSHCredentialMissing)
	}
}

func TestOperatorSSHCredentialsAdmitOnlyResolvedObjectsOfTheRightKind(t *testing.T) {
	identity, agentSocket, knownHosts := operatorMaterial(t)
	root := t.TempDir()
	directory := filepath.Join(root, "directory")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, credentials := range map[string]OperatorSSHCredentials{
		"relative identity":     {Identity: "id_ed25519"},
		"absent identity":       {Identity: filepath.Join(root, "absent")},
		"identity is directory": {Identity: directory},
		"agent is a file":       {AgentSocket: identity},
		"known hosts directory": {Identity: identity, KnownHosts: directory},
	} {
		if _, err := ValidateOperatorSSHCredentials(credentials); ErrorCode(err) != CodeIdentityInvalid {
			t.Errorf("%s: error = %v, want %s", name, err, CodeIdentityInvalid)
		}
	}
	// A link onto the material is resolved, not refused: a live agent socket
	// is conventionally reached through exactly that shape.
	link := filepath.Join(root, "agent.link")
	if err := os.Symlink(agentSocket, link); err != nil {
		t.Skipf("this host cannot create a symbolic link: %v", err)
	}
	admitted, err := ValidateOperatorSSHCredentials(OperatorSSHCredentials{
		Identity: identity, AgentSocket: link, KnownHosts: knownHosts,
	})
	if err != nil {
		t.Fatal(err)
	}
	if admitted.AgentSocket != agentSocket {
		t.Fatalf("agent socket = %q, want the resolved %q", admitted.AgentSocket, agentSocket)
	}
}

// The admission boundary is the last place an unselected SSH repository can be
// stopped, and it has to stop it even when a caller forgot to resolve anything.
func TestAcquireNetworkRefusesAnSSHRepositoryWithoutAnOperatorSelection(t *testing.T) {
	tool := realGitTool(t)
	tool.SSHWrapper = writeOperatorFile(t, filepath.Join(t.TempDir(), "ssh-wrapper"), "#!/bin/sh\nexit 1\n")
	source, err := ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	_, err = AcquireNetwork(context.Background(), NetworkRequest{
		Source: source,
		Lock:   LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)},
		Tool:   tool,
	})
	if ErrorCode(err) != CodeSSHCredentialMissing {
		t.Fatalf("unselected SSH acquisition error = %v, want %s", err, CodeSSHCredentialMissing)
	}
}
