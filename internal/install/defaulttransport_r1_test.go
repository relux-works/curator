package install

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildrepo"
)

// A package-provided ssh first on inherited PATH cannot become the selected
// executable. Explicitly selecting that project command also refuses; an
// operator-selected fixture outside managed roots is the positive control.
// Only selection is exercised: no real SSH client is ever started.
func TestR1ProjectBinSSHSelected(t *testing.T) {
	bin := filepath.Join(t.TempDir(), ".agents", "bin")
	if err := os.MkdirAll(bin, 0755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(bin, "ssh")
	if err := os.WriteFile(foreign, []byte("foreign package program"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	t.Setenv("CURATOR_BUILD_SSH_EXECUTABLE", "")
	inherited, release, selectErr := draftSSHWrapperBase()
	release()
	if selectErr != nil || inherited.SSH == foreign {
		t.Fatalf("PATH selected project ssh: %+v %v", inherited, selectErr)
	}
	t.Setenv("CURATOR_BUILD_SSH_EXECUTABLE", foreign)
	base, cleanup, err := draftSSHWrapperBase()
	defer cleanup()
	if err != nil || base.SSH != "" {
		t.Fatalf("project SSH selected: %+v %v", base, err)
	}
	trusted := filepath.Join(t.TempDir(), "ssh")
	if err := os.WriteFile(trusted, []byte("operator fixture"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_BUILD_SSH_EXECUTABLE", trusted)
	base, cleanup, err = draftSSHWrapperBase()
	defer cleanup()
	if err != nil || base.SSH != trusted {
		t.Fatalf("trusted control: %+v %v", base, err)
	}
	if got, _ := os.ReadFile(foreign); string(got) != "foreign package program" {
		t.Fatal("foreign executable changed")
	}
}

// Missing credentials refuse on the actual default acquisition path before
// any Git/SSH probe on every platform, including Windows. No network or writes.
func TestDefaultTransportMissingCredentialsPortable(t *testing.T) {
	source, err := buildrepo.ParseSource("git@example.test:skills/tool.git")
	if err != nil {
		t.Fatal(err)
	}
	_, err = acquireDefaultNetwork(context.Background(), buildrepo.NetworkRequest{Source: source, Lock: buildrepo.LockedCommit{ObjectFormat: "sha1", Hex: strings.Repeat("a", 40)}})
	if buildrepo.ErrorCode(err) != buildrepo.CodeSSHCredentialMissing {
		t.Fatalf("default-lane missing credential: %v", err)
	}
}

// A missing, relative or non-file manager wrapper is blocked credential policy,
// not source grammar. A valid absolute regular file is admitted without running
// it; refusal preserves the file and exposes no fixture path.
func TestDraftManagerExecutableAdmission(t *testing.T) {
	root := t.TempDir()
	valid := filepath.Join(root, "manager")
	if err := os.WriteFile(valid, []byte("manager fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	if got, err := draftManagerExecutableFrom(valid); err != nil || got != valid {
		t.Fatalf("positive: %q %v", got, err)
	}
	for _, path := range []string{"relative-manager", valid + "-absent", root} {
		got, err := draftManagerExecutableFrom(path)
		if got != "" || buildrepo.ErrorCode(err) != buildrepo.CodeCredentialPolicyInvalid || strings.Contains(err.Error(), root) {
			t.Fatalf("policy refusal: %q %v", got, err)
		}
	}
	if got, err := os.ReadFile(valid); err != nil || string(got) != "manager fixture" {
		t.Fatalf("preimage changed: %q %v", got, err)
	}
}
