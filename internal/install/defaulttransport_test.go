package install

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
)

// These cases exercise real Git and upload-pack through a capture SSH fixture
// on the default acquisition lane and real project/global install entrypoints.
// The builder is injected; they prove fetch isolation and publication, not
// native compiler containment or cryptographic SSH authentication.
func TestDefaultTransportInstallScopes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only; this POSIX capture fixture cannot execute on Windows; portable admission and native wrapper witnesses run separately")
	}
	for _, scope := range []string{"project", "global"} {
		for _, mode := range []string{"selected", "pinned-agent", "missing", "wrong-pin", "transport", "missing-wrapper", "no-host-keys", "missing-key-file", "missing-hosts-file", "no-ssh", "ref-moved", "https"} {
			t.Run(scope+"/"+mode, func(t *testing.T) {
				e := newEnv(t)
				writeLegacyBuildSkill(t, e, 7, []string{"go-repository-v1"})
				sourceRoot := t.TempDir()
				for _, f := range lifecycleSnapshot("golden-tool").Files {
					e.write(sourceRoot, f.Path, string(f.Content))
				}
				e.git(sourceRoot, "init", "-q", "-b", "main")
				e.git(sourceRoot, "add", ".")
				e.git(sourceRoot, "commit", "-qm", "fixture")
				pin := strings.TrimSpace(e.gitOutput(sourceRoot, "rev-parse", "HEAD"))
				if mode == "wrong-pin" {
					pin = strings.Repeat("f", 40)
				}
				url := draftSSHPrimary
				if mode == "https" {
					url = draftHTTPSPrimary
				}
				skillRoot := filepath.Join(e.skillsRoot, "golden-skill")
				payload, err := os.ReadFile(filepath.Join(skillRoot, "csk-skill.json"))
				if err != nil {
					t.Fatal(err)
				}
				var manifest map[string]any
				if err := json.Unmarshal(payload, &manifest); err != nil {
					t.Fatal(err)
				}
				repository := map[string]any{"git": url, "locked_commit": map[string]any{"object_format": "sha1", "hex": pin}}
				if mode == "ref-moved" {
					e.git(sourceRoot, "tag", "v1")
					repository["tag"] = "v1"
					repository["locked_commit"] = map[string]any{"object_format": "sha1", "hex": strings.Repeat("f", 40)}
				}
				manifest["build_repositories"] = map[string]any{"golden-tools": repository}
				payload, err = json.Marshal(manifest)
				if err != nil {
					t.Fatal(err)
				}
				e.write(skillRoot, "csk-skill.json", string(payload))
				e.git(skillRoot, "add", ".")
				e.git(skillRoot, "commit", "-qm", "fixture source")
				e.git(skillRoot, "tag", "-f", "v1")
				tool, logPath := defaultCaptureTool(t, sourceRoot, mode)
				credentials := tool.SSHCredentials
				external := ExternalDeps{GitTool: tool, Audit: func(context.Context, buildrepo.AuditSubject) error { return nil }}
				if mode != "missing" {
					external.BuildSSH.RunWide = BuildSSHFlags{Identity: credentials.Identity, KnownHosts: credentials.KnownHosts, Agent: credentials.AgentSocket}
				}
				deps, _ := lifecycleBuildDeps(t)
				options := Options{Platform: installPlatform(), Build: deps, External: external}
				before := snapshotState(t, e)
				var result Result
				shim := filepath.Join(e.project, ".agents", "bin", shimName("golden-tool"))
				if scope == "global" {
					e.globalDeclare("golden-skill")
					result = Global(e.cfg, t.TempDir(), options)
					shim = filepath.Join(GlobalRoot(e.home), "bin", shimName("golden-tool"))
				} else {
					result = e.install(options)
				}
				text := fmt.Sprintf("%+v", result)
				log, _ := os.ReadFile(logPath)
				switch mode {
				case "selected", "pinned-agent", "https":
					if result.Status != "ok" {
						fixtureStderr, _ := os.ReadFile(logPath + ".git-stderr")
						t.Fatalf("install: %s; fixture stderr=%s", text, fixtureStderr)
					}
					if _, err := os.Lstat(shim); err != nil {
						t.Fatal(err)
					}
					if mode == "selected" || mode == "pinned-agent" {
						for _, want := range []string{"<-F>", "<" + tool.SSHBase.EmptyConfig + ">", "<IdentityAgent=" + func() string {
							if credentials.AgentSocket != "" {
								return credentials.AgentSocket
							}
							return "none"
						}() + ">", "<IdentitiesOnly=yes>", "<" + credentials.Identity + ">", "<StrictHostKeyChecking=yes>", "<git@fixture.test>"} {
							if !strings.Contains(string(log), want) {
								t.Errorf("missing %s in %s", want, log)
							}
						}
						if strings.Contains(string(log), "ambient") {
							t.Fatalf("ambient material reached transport: %s", log)
						}
					}
				default:
					code, reason := buildrepo.CodeSourceUnavailable, "transport_failed"
					if mode == "missing" {
						code, reason = buildrepo.CodeSSHCredentialMissing, ""
					}
					if mode == "missing-wrapper" {
						code, reason = buildrepo.CodeCredentialPolicyInvalid, "exact manager wrapper"
					}
					if mode == "no-host-keys" {
						code, reason = buildrepo.CodeCredentialPolicyInvalid, "require operator host keys"
					}
					if mode == "missing-key-file" {
						code, reason = buildrepo.CodeCredentialPolicyInvalid, "SSH identity is unavailable"
					}
					if mode == "missing-hosts-file" {
						code, reason = buildrepo.CodeCredentialPolicyInvalid, "SSH known hosts is unavailable"
					}
					if mode == "no-ssh" {
						code, reason = buildrepo.CodeCredentialPolicyInvalid, "no ssh executable found"
					}
					if mode == "ref-moved" {
						code, reason = buildrepo.CodeRefMoved, ""
					}
					if mode == "wrong-pin" {
						reason = "locked_ref_unavailable"
					}
					if result.Status == "ok" || !strings.Contains(text, code) || !strings.Contains(text, reason) {
						stderr, _ := os.ReadFile(logPath + ".git-stderr")
						t.Fatalf("want %s %s: %s; fixture stderr=%s", code, reason, text, stderr)
					}
					if mode == "wrong-pin" && strings.Contains(text, "transport_failed") {
						t.Fatal(text)
					}
					if err := stateUnchanged(t, before, snapshotState(t, e), "credential refusal"); err != nil {
						t.Fatal(err)
					}
					if _, err := os.Lstat(shim); !os.IsNotExist(err) {
						t.Fatalf("refusal published shim: %v", err)
					}
					if (mode == "missing" || mode == "missing-wrapper" || mode == "no-host-keys" || mode == "missing-key-file" || mode == "missing-hosts-file" || mode == "no-ssh") && len(log) != 0 {
						t.Fatalf("refusal reached transport: %s", log)
					}
				}
			})
		}
	}
}

func defaultCaptureTool(t *testing.T, repository, mode string) (buildrepo.GitTool, string) {
	t.Helper()
	// Credential admission resolves symlinks (including macOS /var). Capture
	// expectations name that same selected file, rather than its path alias.
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write := func(name, content string) string {
		p := filepath.Join(root, name)
		if err := os.WriteFile(p, []byte(content), 0o700); err != nil {
			t.Fatal(err)
		}
		return p
	}
	identity := write("selected-key", "fixture selected key\n")
	knownHosts := write("known_hosts", "fixture host\n")
	empty := write("empty-config", "")
	log := filepath.Join(root, "ssh.log")
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	git, err = filepath.EvalSymlinks(git)
	if err != nil {
		t.Fatal(err)
	}
	// The fixture rejects any key other than the selected one and requires the
	// empty -F config. It never launches OpenSSH or reaches an external host.
	ssh := write("capture-ssh", fmt.Sprintf(`#!/bin/sh
{ for arg in "$@"; do printf '<%%s>' "$arg"; done; printf '\n'; } >> %s
[ ! -e "$HOME/.ssh/config" ] || { printf 'ambient config reached fixture' > %s; printf 'ambient config reached fixture' >&2; exit 97; }
[ -z "${SSH_AUTH_SOCK+x}" ] || { printf 'ambient agent reached fixture' > %s; printf 'ambient agent reached fixture' >&2; exit 96; }
[ "$1" = '-F' ] && [ "$2" = %s ] || exit 91
selected=0
while [ "$#" -gt 2 ]; do
 if [ "$1" = '-i' ]; then [ "$2" = %s ] || exit 92; selected=1; fi
 shift
done
[ "$selected" = 1 ] || exit 93
[ "$1" = 'git@fixture.test' ] && [ "$2" = "git-upload-pack '/repository.git'" ] || exit 94
%s
%s upload-pack %s 2>%s
status=$?
/bin/cat %s >&2
exit "$status"
`, shellQuote(log), shellQuote(log+".git-stderr"), shellQuote(log+".git-stderr"), shellQuote(empty), shellQuote(identity), func() string {
		if mode == "transport" {
			return "printf 'ssh: connect to host fixture.test port 22: Connection refused\\n' >&2; exit 128"
		}
		return ""
	}(), shellQuote(git), shellQuote(repository), shellQuote(log+".stderr"), shellQuote(log+".stderr")))
	execPath, err := exec.Command(git, "--exec-path").Output()
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(git, "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	manager, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	tool := buildrepo.GitTool{Executable: git, ExecPath: strings.TrimSpace(string(execPath)), AllowedVersions: []string{strings.TrimSpace(string(version))}, AskPass: manager, SSHWrapper: manager, SSHCredentials: buildrepo.OperatorSSHCredentials{Identity: identity, KnownHosts: knownHosts}, SSHBase: buildrepo.SSHWrapperBase{SSH: ssh, EmptyConfig: empty, EmptyKnownHosts: empty, ConnectTimeout: 15}}
	if mode == "pinned-agent" {
		// Unix socket addresses are bounded independently of filesystem paths.
		// The selected socket is fixture-owned; the capture never connects.
		socketDir, err := os.MkdirTemp("/tmp", "curator-test-agent-")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(socketDir) })
		socket := filepath.Join(socketDir, "agent.sock")
		listener, err := net.Listen("unix", socket)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = listener.Close() })
		socket, err = filepath.EvalSymlinks(socket)
		if err != nil {
			t.Fatal(err)
		}
		tool.SSHCredentials.AgentSocket = socket
	}
	if mode == "wrong-pin" {
		tool.Executable = write("capture-git", fmt.Sprintf("#!/bin/sh\n%s \"$@\" 2>%s\nstatus=$?\n/bin/cat %s >&2\nexit \"$status\"\n", shellQuote(git), shellQuote(log+".git-stderr"), shellQuote(log+".git-stderr")))
	}
	if mode == "missing-wrapper" {
		tool.SSHWrapper = ""
	}
	if mode == "https" {
		tool, _ = draftGitTool(t, map[string]draftGitArm{draftHTTPSPrimary: {succeed: true, fileRepo: repository}})
	}
	switch mode {
	case "no-host-keys":
		tool.SSHCredentials.KnownHosts = ""
	case "missing-key-file":
		tool.SSHCredentials.Identity += "-missing"
	case "missing-hosts-file":
		tool.SSHCredentials.KnownHosts += "-missing"
	case "no-ssh":
		// A zero base requests production discovery; an unavailable explicit
		// operator path exercises its actual fail-closed no-executable path.
		t.Setenv("CURATOR_BUILD_SSH_EXECUTABLE", filepath.Join(root, "missing-ssh"))
		tool.SSHBase = buildrepo.SSHWrapperBase{}
	}
	// Plant ambient hooks that would fail or leak into argv if inherited. Paths
	// are wholly fixture-owned; no real SSH executable or real home is touched.
	t.Setenv("SSH_AUTH_SOCK", filepath.Join(root, "ambient-agent"))
	ehome := t.TempDir()
	if err := os.Mkdir(filepath.Join(ehome, ".ssh"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ehome, ".ssh", "config"), []byte("Host *\n IdentityFile /fixture/ambient-identity\n ProxyCommand /fixture/ambient-proxy\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", ehome)
	return tool, log
}
