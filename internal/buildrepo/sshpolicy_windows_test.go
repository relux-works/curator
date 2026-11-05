package buildrepo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

// A native inherited read handle reaches a direct child (positive control),
// but Git for Windows removes it from its helper child graph. No SSH runs.
// This establishes the reason for the explicit Windows policy-record transport.
func TestWindowsGitDoesNotForwardPolicyHandle(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	probe := filepath.Join(root, "git-policy-probe.exe")
	if err := copyBrokerExecutable(executable, probe); err != nil {
		t.Fatal(err)
	}
	policy, err := os.CreateTemp(t.TempDir(), "policy-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = policy.Close() }()
	if _, err := policy.WriteString("protected-policy"); err != nil {
		t.Fatal(err)
	}
	if err := windows.SetHandleInformation(windows.Handle(policy.Fd()), windows.HANDLE_FLAG_INHERIT, windows.HANDLE_FLAG_INHERIT); err != nil {
		t.Fatal(err)
	}
	run := func(command string, args ...string) string {
		t.Helper()
		child := exec.Command(command, args...)
		child.Env = append(os.Environ(), "CURATOR_TEST_POLICY_HANDLE="+strconv.FormatUint(uint64(policy.Fd()), 10))
		child.SysProcAttr = &syscall.SysProcAttr{AdditionalInheritedHandles: []syscall.Handle{syscall.Handle(policy.Fd())}}
		output, err := child.CombinedOutput()
		if err != nil {
			t.Fatalf("probe: %v %s", err, output)
		}
		return strings.TrimSpace(string(output))
	}
	if got := run(probe, "--curator-policy-handle-probe"); got != "protected-policy" {
		t.Fatalf("direct handle positive control: %q", got)
	}
	if got := run(git, "--exec-path="+root, "policy-probe", "--curator-policy-handle-probe"); got != "policy-handle-unavailable" {
		t.Fatalf("Git handle boundary: %q", got)
	}
	t.Log("direct child: protected-policy; Git helper: policy-handle-unavailable")
}

func runWindowsPolicyHandleProbe() bool {
	if len(os.Args) != 2 || os.Args[1] != "--curator-policy-handle-probe" {
		return false
	}
	handle, err := strconv.ParseUint(os.Getenv("CURATOR_TEST_POLICY_HANDLE"), 10, 64)
	if err != nil {
		os.Exit(91)
	}
	file := os.NewFile(uintptr(handle), "inherited-policy")
	data := make([]byte, len("protected-policy"))
	if _, err := file.ReadAt(data, 0); err != nil {
		fmt.Println("policy-handle-unavailable")
	} else {
		fmt.Println(string(data))
	}
	os.Exit(0)
	return true
}

// Starting from an explicitly foreign-writable fixture DACL, protection leaves
// exactly one inheritable current-user ACE and disables parent inheritance.
// The world-writable plant makes a no-op protection implementation fail.
func TestWindowsSSHPolicyDirectoryIsPrivate(t *testing.T) {
	root := t.TempDir()
	planted, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	acl, _, err := planted.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(root, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		t.Fatal(err)
	}
	if err := protectSSHPolicyDirectory(root); err != nil {
		t.Fatal(err)
	}
	descriptor, err := windows.GetNamedSecurityInfo(root, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatal(err)
	}
	control, _, err := descriptor.Control()
	if err != nil {
		t.Fatal(err)
	}
	if dacl == nil || dacl.AceCount != 1 || control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("policy DACL: %s", descriptor.String())
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(dacl, 0, &ace); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
	if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE || !sid.Equals(user.User.Sid) || ace.Header.AceFlags&(windows.OBJECT_INHERIT_ACE|windows.CONTAINER_INHERIT_ACE) != (windows.OBJECT_INHERIT_ACE|windows.CONTAINER_INHERIT_ACE) {
		t.Fatalf("policy trustee: %s", descriptor.String())
	}
	if err := os.WriteFile(filepath.Join(root, "policy"), []byte("private"), 0600); err != nil {
		t.Fatalf("current-user write control: %v", err)
	}
}
