// The hosted-only control fetches through the product wrapper using OpenSSH 8.9.
// A correct key and pinned host succeed; wrong-key, unknown-host and changed-key attempts
// must report distinct static policy reasons and return no snapshot.
package main

import (
	"context"
	"fmt"
	"github.com/relux-works/curator/internal/buildrepo"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if buildrepo.IsSSHWrapperInvocation(os.Args[0]) {
		os.Exit(buildrepo.RunSSHWrapper(os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr))
	}
	root := os.Args[1]
	commit := os.Args[2]
	user := os.Args[3]
	git, err := buildrepo.SelectTrustedExecutable("git", "")
	must(err)
	ssh, err := buildrepo.SelectTrustedExecutable("ssh", "")
	must(err)
	probe := func(args ...string) string {
		cmd := exec.Command(git, args...)
		cmd.Env = buildrepo.DiscoveryEnvironment(os.Environ())
		out, err := cmd.Output()
		must(err)
		return strings.TrimSpace(string(out))
	}
	fingerprint, err := buildrepo.ExecutableFingerprint(git)
	must(err)
	manager, err := os.Executable()
	must(err)
	source, err := buildrepo.ParseSource(user + "@127.0.0.1:" + filepath.ToSlash(filepath.Join(filepath.Base(root), "repository.git")))
	must(err)
	tool := buildrepo.GitTool{Executable: git, Fingerprint: fingerprint, ExecPath: probe("--exec-path"), AllowedVersions: []string{probe("--version")}, SSHWrapper: manager,
		SSHBase: buildrepo.SSHWrapperBase{SSH: ssh, EmptyConfig: filepath.Join(root, "empty-config"), EmptyKnownHosts: filepath.Join(root, "empty-hosts"), ConnectTimeout: 10}}
	for _, c := range []struct{ name, key, hosts, reason string }{{"key-fetch", "identity", "known-hosts", ""}, {"wrong-key", "wrong-identity", "known-hosts", "authentication_failed"}, {"unknown-host", "identity", "empty-hosts", "host_key_verification_failed"}, {"changed-key", "identity", "changed-hosts", "host_key_verification_failed"}} {
		tool.SSHCredentials = buildrepo.OperatorSSHCredentials{Identity: filepath.Join(root, c.key), KnownHosts: filepath.Join(root, c.hosts)}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		snapshot, err := buildrepo.AcquireNetwork(ctx, buildrepo.NetworkRequest{Source: source, Lock: buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: commit}, Tool: tool})
		cancel()
		if c.reason == "" {
			if err != nil || snapshot == nil || snapshot.Commit != commit {
				panic(fmt.Sprintf("positive key fetch: %v", err))
			}
		} else {
			if snapshot != nil || buildrepo.ErrorCode(err) != buildrepo.CodeCredentialPolicyInvalid || !strings.Contains(fmt.Sprint(err), "("+c.reason+")") {
				panic(fmt.Sprintf("%s: %v", c.name, err))
			}
		}
		fmt.Printf("PASS %s: %v\n", c.name, err)
	}
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
