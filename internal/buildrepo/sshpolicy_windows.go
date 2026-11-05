package buildrepo

import (
	"context"
	"golang.org/x/sys/windows"
	"os"
	"path/filepath"
)

// os/exec.ExtraFiles is unsupported on Windows. The explicit Windows spec
// alternative opens the bounded record in the operation-private wrapper dir.
func runGitCaptureSSH(ctx context.Context, executable, dir string, env []string, state string, args ...string) (string, bool, error) {
	return runGitCapture(ctx, executable, dir, setEnvironmentValue(env, EnvSSHWrapperState, state), args...)
}

func sshPolicyPayload(path string) ([]byte, error) {
	if !filepath.IsAbs(path) {
		return nil, os.ErrPermission
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 8192 {
		return nil, os.ErrPermission
	}
	return os.ReadFile(path) // #nosec G304 -- bounded record in the private manager wrapper directory.
}

// Git for Windows forwards only standard handles to helpers. Protect this
// operation-private record with an explicit current-user-only inheritable DACL;
// a mode argument alone does not protect a Windows policy record.
func protectSSHPolicyDirectory(path string) error {
	token := windows.GetCurrentProcessToken()
	user, err := token.GetTokenUser()
	if err != nil {
		return err
	}
	descriptor, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;" + user.User.Sid.String() + ")")
	if err != nil {
		return err
	}
	acl, _, err := descriptor.DACL()
	if err != nil {
		return err
	}
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil)
}
