//go:build !windows

package buildrepo

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
)

// Git inherits descriptor 3 and passes it to the exact binary wrapper. The
// original is read-only; Git's standard streams keep their protocol roles.
func runGitCaptureSSH(ctx context.Context, executable, dir string, env []string, state string, args ...string) (string, bool, error) {
	policy, err := os.Open(state)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = policy.Close() }()
	return runGitCapturePolicy(ctx, executable, dir, setEnvironmentValue(env, EnvSSHWrapperState, "3"), policy, args...)
}

func sshPolicyPayload(value string) ([]byte, error) {
	number, err := strconv.Atoi(value)
	if err != nil || number < 3 {
		return nil, fmt.Errorf("invalid policy descriptor")
	}
	fd, err := syscall.Dup(number)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), "manager-policy")
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 8192 {
		return nil, fmt.Errorf("invalid policy descriptor")
	}
	return io.ReadAll(io.NewSectionReader(file, 0, 8193))
}

func protectSSHPolicyDirectory(string) error { return nil }
