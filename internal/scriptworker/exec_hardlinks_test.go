package scriptworker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func fixtureWindowsExecOrigin(t *testing.T, target, alias string, owned bool) {
	t.Helper()
	origin := windowsExecHardlinkOrigin{OwnerSID: trustedInstallerSID, Links: []string{target, alias}, Count: 2}
	if !owned {
		origin.OwnerSID = "S-1-5-21-1-2-3-1001"
	}
	previous := inspectWindowsExecHardlinks
	inspectWindowsExecHardlinks = func(file *os.File, path string) (windowsExecHardlinkOrigin, error) {
		if path != target || file.Name() != target {
			return windowsExecHardlinkOrigin{}, errors.New("unexpected executable at OS seam")
		}
		return origin, nil
	}
	t.Cleanup(func() { inspectWindowsExecHardlinks = previous })
}

// Every observation reaches resolveExecForPlatform, also used by DeriveProfile.
// VerifyExec must repeat the origin proof even when content has not changed.
func TestWindowsExecHardlinkOriginEvidence(t *testing.T) {
	root := mustPhysical(t, t.TempDir())
	target := filepath.Join(root, "System32", "cmd.exe")
	alias := filepath.Join(root, "WinSxS", "cmd.exe")
	writeTestFile(t, target, []byte("executable fixture"), 0o755)
	if err := os.MkdirAll(filepath.Dir(alias), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(target, alias); err != nil {
		t.Fatal(err)
	}
	good := windowsExecHardlinkOrigin{OwnerSID: trustedInstallerSID, Links: []string{target, alias}, Count: 2}
	previous := inspectWindowsExecHardlinks
	t.Cleanup(func() { inspectWindowsExecHardlinks = previous })
	cases := []struct {
		name string
		edit func(*windowsExecHardlinkOrigin)
		err  error
		want bool
	}{
		{name: "trusted-installer", want: true},
		{name: "local-system", edit: func(o *windowsExecHardlinkOrigin) { o.OwnerSID = localSystemSID }, want: true},
		{name: "unknown-owner", edit: func(o *windowsExecHardlinkOrigin) { o.OwnerSID = "" }},
		{name: "ordinary-owner", edit: func(o *windowsExecHardlinkOrigin) { o.OwnerSID = "S-1-5-21-1-2-3-1001" }},
		{name: "read-failure", err: errors.New("access denied")},
		{name: "partial-enumeration", edit: func(o *windowsExecHardlinkOrigin) { o.Count = 3 }},
		{name: "duplicate-alias", edit: func(o *windowsExecHardlinkOrigin) { o.Links[1] = target }},
		{name: "missing-target", edit: func(o *windowsExecHardlinkOrigin) { o.Links[0] = alias + "-other" }},
		{name: "relative-alias", edit: func(o *windowsExecHardlinkOrigin) { o.Links[1] = "WinSxS/cmd.exe" }},
		{name: "store-prefix-sibling", edit: func(o *windowsExecHardlinkOrigin) { o.Links[1] = filepath.Join(root, "WinSxS-other", "cmd.exe") }},
		{name: "third-link-outside-store", edit: func(o *windowsExecHardlinkOrigin) {
			o.Count = 3
			o.Links = append(o.Links, filepath.Join(root, "OtherLinks", "cmd.exe"))
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			origin := good
			origin.Links = append([]string(nil), good.Links...)
			if testCase.edit != nil {
				testCase.edit(&origin)
			}
			inspectWindowsExecHardlinks = func(*os.File, string) (windowsExecHardlinkOrigin, error) { return origin, testCase.err }
			identity, found, err := resolveExecForPlatform("cmd.exe", nil, nil, "windows", []string{"SYSTEMROOT=" + root}, true)
			if err != nil || found != testCase.want {
				t.Fatalf("production resolve found=%t err=%v, want %t", found, err, testCase.want)
			}
			if found {
				if err := VerifyExec(identity); err != nil {
					t.Fatal(err)
				}
				origin.OwnerSID = ""
				if err := VerifyExec(identity); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
					t.Fatalf("changed owner at launch boundary: %v", err)
				}
				origin = good
				origin.Links = []string{target, filepath.Join(root, "OtherLinks", "cmd.exe")}
				if err := VerifyExec(identity); DiagnosticCode(err) != CodeWorkerIdentityInvalid {
					t.Fatalf("changed link origin at launch boundary: %v", err)
				}
			}
		})
	}
}
