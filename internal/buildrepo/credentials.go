package buildrepo

import (
	"os"
	"path/filepath"
	"strings"
)

// CodeSSHCredentialMissing reports an SSH build repository the operator has
// selected no credentials for. Ambient SSH state is never a fallback: a fetch
// that succeeds because some unrelated key happened to be loaded in the
// operator's agent is a fetch nobody authorized.
const CodeSSHCredentialMissing = "build_repository_ssh_credential_missing"

// OperatorSSHCredentials is one operator SSH selection for one external build
// repository. It is operator-owned: no manifest, descriptor, repository,
// substitution, or marker may select it (Spec §12.2).
//
// The three shapes an SSH invocation admits are an identity alone, an agent
// alone, and an agent pinned to one named identity. SSHPolicyFor carries each
// of them into the wrapper policy unchanged.
type OperatorSSHCredentials struct {
	// Identity is the identity file offered to the destination. Together with
	// AgentSocket it pins the single key the agent may offer.
	Identity string
	// AgentSocket is the agent socket the destination authenticates against.
	AgentSocket string
	// KnownHosts carries the operator host keys for this repository. The fetch
	// pins StrictHostKeyChecking=yes and has no other source of truth.
	KnownHosts string
}

// Selected reports whether the operator named any authentication material.
// Known hosts alone authenticate nothing, so they do not count as a selection.
func (c OperatorSSHCredentials) Selected() bool {
	return c.Identity != "" || c.AgentSocket != ""
}

// ValidateOperatorSSHCredentials resolves every operator-named path and proves
// it is the admitted kind of object.
//
// A symbolic link is resolved rather than refused: a live agent socket is
// conventionally a stable link onto a per-session rendezvous point. What must
// hold is that the resolved target exists and is the right kind.
func ValidateOperatorSSHCredentials(credentials OperatorSSHCredentials) (OperatorSSHCredentials, error) {
	for _, field := range []struct {
		label  string
		socket bool
		value  *string
	}{
		{"SSH identity", false, &credentials.Identity},
		{"SSH agent socket", true, &credentials.AgentSocket},
		{"SSH known hosts", false, &credentials.KnownHosts},
	} {
		if *field.value == "" {
			continue
		}
		if !filepath.IsAbs(*field.value) {
			return OperatorSSHCredentials{}, admissionError(CodeIdentityInvalid, "%s path must be absolute", field.label)
		}
		target, err := filepath.EvalSymlinks(*field.value)
		if err != nil {
			return OperatorSSHCredentials{}, admissionError(CodeIdentityInvalid, "%s is unavailable", field.label)
		}
		info, err := os.Lstat(target)
		if err != nil {
			return OperatorSSHCredentials{}, admissionError(CodeIdentityInvalid, "%s is unavailable", field.label)
		}
		admitted, kind := info.Mode().IsRegular(), "regular file"
		if field.socket {
			admitted, kind = info.Mode()&os.ModeSocket != 0, "socket"
		}
		if !admitted {
			return OperatorSSHCredentials{}, admissionError(CodeIdentityInvalid, "%s is not an admitted %s", field.label, kind)
		}
		*field.value = target
	}
	return credentials, nil
}

// SSHEndpoint returns the exact host and repository path Git hands to the SSH
// program. Git passes `user@host` verbatim and quotes the path exactly as
// written: an `ssh://` source keeps its leading slash, an scp-like source does
// not.
func SSHEndpoint(source Source) (string, string, error) {
	if source.Transport != "ssh" {
		return "", "", admissionError(CodeIdentityInvalid, "repository transport is not SSH")
	}
	if remainder, found := strings.CutPrefix(source.Git, "ssh://"); found {
		boundary := strings.Index(remainder, "/")
		if boundary <= 0 {
			return "", "", admissionError(CodeIdentityInvalid, "SSH source has no repository path")
		}
		return remainder[:boundary], remainder[boundary:], nil
	}
	host, path, found := strings.Cut(source.Git, ":")
	if !found || host == "" || path == "" {
		return "", "", admissionError(CodeIdentityInvalid, "SSH source has no repository path")
	}
	return host, path, nil
}

// SSHPolicyFor completes a manager-owned wrapper policy with the exact endpoint
// of one SSH source and the operator's credential selection. Every path in base
// belongs to the manager; the operator owns only what the returned policy
// authenticates with, and the known-hosts file it verifies the host against.
func SSHPolicyFor(base SSHPolicy, source Source, credentials OperatorSSHCredentials) (SSHPolicy, error) {
	if !credentials.Selected() {
		return SSHPolicy{}, admissionError(CodeSSHCredentialMissing,
			"SSH build repositories require an operator identity or agent")
	}
	host, repositoryPath, err := SSHEndpoint(source)
	if err != nil {
		return SSHPolicy{}, err
	}
	admitted, err := ValidateOperatorSSHCredentials(credentials)
	if err != nil {
		return SSHPolicy{}, err
	}
	policy := base
	policy.ExpectedHost, policy.RepositoryPath = host, repositoryPath
	policy.Identity, policy.AgentSocket = admitted.Identity, admitted.AgentSocket
	if admitted.KnownHosts != "" {
		policy.KnownHosts = admitted.KnownHosts
	}
	if policy.KnownHosts == "" {
		return SSHPolicy{}, admissionError(CodeSSHCredentialMissing,
			"SSH build repositories pin StrictHostKeyChecking and require operator host keys")
	}
	return policy, nil
}
