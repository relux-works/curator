package buildrepo

// Per-attempt SSH wrapper policies for bounded transport resolution.
//
// The strict lane admits one static SSH wrapper per tool, but a resolved
// acquisition may attempt two endpoints with two different credential
// selections. Each SSH attempt therefore binds its endpoint and its
// selected credentials into a real wrapper policy through the existing
// SSHPolicyFor builder and runs the fetch behind a private wrapper copy
// pinned to that policy, mirroring the per-fetch HTTPS credential broker:
// the manager executable is copied under a private basename the manager
// binary dispatches, and the policy travels in a secret-free state file.
// Paths name authentication material; no secret is copied or written.

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

const (
	// SSHWrapperName is the basename of the manager-owned per-attempt SSH
	// wrapper copy. The manager binary dispatches this basename before its
	// public CLI.
	SSHWrapperName = "curator-build-ssh-wrapper"
	// EnvSSHWrapperState names the manager-owned, secret-free wrapper-state
	// file for one fetch.
	EnvSSHWrapperState = "CURATOR_BUILD_SSH_WRAPPER_STATE"
)

// SSHWrapperBase is the manager-owned base of every per-attempt SSH wrapper
// policy: the SSH executable and the empty configuration the fixed argv
// pins, plus the single connection timeout. It carries no endpoint and no
// credentials; the resolved executor completes it per attempt with the
// endpoint source and the attempt's operator credential selection. The
// wiring task populates it beside the per-repository credential selection;
// a zero base binds no SSH policy and every SSH attempt refuses.
type SSHWrapperBase struct {
	SSH             string
	EmptyConfig     string
	EmptyKnownHosts string
	ConnectTimeout  int
}

// bindSSHWrapperPolicy completes the real per-attempt wrapper policy for
// one SSH source from the manager base and the attempt's bound credential
// selection. A selection that cannot form a policy — unknown endpoint,
// unresolvable paths, or no pinned host keys — refuses here, before any
// fetch traffic, instead of authenticating against ambient SSH state.
func bindSSHWrapperPolicy(tool GitTool, source Source) (SSHPolicy, error) {
	for name, value := range map[string]string{
		"SSH executable":        tool.SSHBase.SSH,
		"SSH empty config":      tool.SSHBase.EmptyConfig,
		"SSH empty known hosts": tool.SSHBase.EmptyKnownHosts,
	} {
		if !filepath.IsAbs(value) {
			return SSHPolicy{}, admissionError(CodeIdentityInvalid, "SSH wrapper policy %s is not configured", name)
		}
		info, err := os.Lstat(value)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return SSHPolicy{}, admissionError(CodeIdentityInvalid, "SSH wrapper policy %s is not an admitted absolute regular file", name)
		}
	}
	if tool.SSHBase.ConnectTimeout <= 0 {
		return SSHPolicy{}, admissionError(CodeIdentityInvalid, "SSH wrapper policy has no connection timeout")
	}
	return SSHPolicyFor(SSHPolicy{
		SSH:             tool.SSHBase.SSH,
		EmptyConfig:     tool.SSHBase.EmptyConfig,
		EmptyKnownHosts: tool.SSHBase.EmptyKnownHosts,
		ConnectTimeout:  tool.SSHBase.ConnectTimeout,
	}, source, tool.SSHCredentials)
}

// IsSSHWrapperInvocation reports whether argv0 names the private per-attempt
// SSH wrapper copy.
func IsSSHWrapperInvocation(argv0 string) bool {
	name := filepath.Base(argv0)
	if runtime.GOOS == "windows" {
		name = strings.TrimSuffix(strings.ToLower(name), ".exe")
	}
	return name == SSHWrapperName
}

// RunSSHWrapper executes one Git SSH-wrapper invocation against the bound
// per-attempt policy: it answers only Git's exact wrapper tuple for the
// pinned endpoint and otherwise fails silently. The SSH child inherits the
// wrapper's environment — the fetch environment, already clean — and its
// exit code is reported back to Git.
func RunSSHWrapper(args []string, getenv func(string) string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 2 || getenv == nil || stdin == nil || stdout == nil || stderr == nil {
		return 1
	}
	policy, ok := readSSHWrapperState(getenv(EnvSSHWrapperState))
	if !ok {
		return 1
	}
	command, err := ExactSSHCommand(policy, []string{policy.Wrapper, args[0], args[1]})
	if err != nil {
		return 1
	}
	child := exec.Command(command[0], command[1:]...) // #nosec G204 -- absolute manager-admitted SSH executable; fixed policy argv.
	child.Stdin, child.Stdout, child.Stderr = stdin, stdout, stderr
	if err := child.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		return 1
	}
	return 0
}

// materializeSSHWrapper copies the manager executable under the private
// wrapper basename and writes the bound per-attempt policy beside it. The
// state file carries paths only, never secrets.
func materializeSSHWrapper(root, executable string, policy SSHPolicy) (string, string, error) {
	if !filepath.IsAbs(executable) {
		return "", "", admissionError(CodeIdentityInvalid, "SSH wrapper source is not absolute")
	}
	managerRoot := filepath.Join(root, "manager-wrappers")
	if err := os.Mkdir(managerRoot, 0o700); err != nil {
		return "", "", err
	}
	name := SSHWrapperName
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	wrapper := filepath.Join(managerRoot, name)
	if err := copyBrokerExecutable(executable, wrapper); err != nil {
		return "", "", err
	}
	policy.Wrapper = wrapper
	payload, err := json.Marshal(policy)
	if err != nil {
		return "", "", err
	}
	payload = append(payload, '\n')
	statePath := filepath.Join(managerRoot, "ssh-wrapper-state.json")
	if err := os.WriteFile(statePath, payload, 0o600); err != nil {
		return "", "", err
	}
	return wrapper, statePath, nil
}

// readSSHWrapperState loads one bound wrapper policy. Every malformed,
// unreadable, or unadmitted state fails silently: the wrapper answers
// nothing without a complete policy.
func readSSHWrapperState(path string) (SSHPolicy, bool) {
	if !filepath.IsAbs(path) {
		return SSHPolicy{}, false
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > 8192 {
		return SSHPolicy{}, false
	}
	payload, err := os.ReadFile(path) // #nosec G304 -- absolute manager-owned path supplied only to the fetch child.
	if err != nil {
		return SSHPolicy{}, false
	}
	var policy SSHPolicy
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return SSHPolicy{}, false
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return SSHPolicy{}, false
	}
	for _, value := range []string{policy.Wrapper, policy.SSH, policy.EmptyConfig, policy.KnownHosts, policy.EmptyKnownHosts} {
		if !validSSHWrapperPath(value) {
			return SSHPolicy{}, false
		}
	}
	if !IsSSHWrapperInvocation(policy.Wrapper) {
		return SSHPolicy{}, false
	}
	for _, value := range []string{policy.Identity, policy.AgentSocket} {
		if value != "" && !validSSHWrapperPath(value) {
			return SSHPolicy{}, false
		}
	}
	if policy.ExpectedHost == "" || policy.RepositoryPath == "" || policy.ConnectTimeout <= 0 {
		return SSHPolicy{}, false
	}
	for _, value := range []string{policy.ExpectedHost, policy.RepositoryPath} {
		for _, r := range value {
			if unicode.IsControl(r) {
				return SSHPolicy{}, false
			}
		}
	}
	return policy, true
}

// validSSHWrapperPath admits one absolute wrapper-state path. Spaces stay
// admissible — operator paths legitimately contain them — but control
// characters never reach an argv.
func validSSHWrapperPath(value string) bool {
	if value == "" || !filepath.IsAbs(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
