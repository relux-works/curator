package testgitenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestMain cleanup removes a read-only module-cache tree. On Unix, it must
// also preserve external symlink targets and hard link contents/modes.
func TestIsolateCleansReadonlyHome(t *testing.T) {
	report := filepath.Join(t.TempDir(), "home")
	outside := t.TempDir()
	sentinel := filepath.Join(outside, "operator")
	if err := os.WriteFile(sentinel, []byte("preserved"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(outside, 0o500); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(outside, 0o700) }()
	before, err := os.Stat(outside)
	if err != nil {
		t.Fatal(err)
	}
	sentinelBefore, err := os.Stat(sentinel)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.v", "-test.run=^TestReadonlyHomeChild$")
	cmd.Env = append(os.Environ(), "CURATOR_TEST_READONLY_HOME="+report)
	if runtime.GOOS != "windows" {
		cmd.Env = append(cmd.Env, "CURATOR_TEST_EXTERNAL_HOME_LINK="+outside)
	}
	output, childErr := cmd.CombinedOutput()
	home, err := os.ReadFile(report)
	if err != nil {
		t.Fatalf("child did not prepare its read-only home: %v\n%s", err, output)
	}
	if _, err := os.Lstat(string(home)); !os.IsNotExist(err) {
		t.Errorf("read-only scratch home retained: %v", err)
	}
	after, err := os.Stat(outside)
	if err != nil || after.Mode() != before.Mode() {
		t.Errorf("external directory mode changed: %v", err)
	}
	if body, err := os.ReadFile(sentinel); err != nil || string(body) != "preserved" {
		t.Errorf("external sentinel changed: %q, %v", body, err)
	}
	if info, err := os.Stat(sentinel); err != nil || info.Mode() != sentinelBefore.Mode() {
		t.Errorf("external sentinel mode changed: %v", err)
	}
	if childErr != nil || !strings.Contains(string(output), "--- PASS: TestReadonlyHomeChild") {
		t.Fatalf("TestMain cleanup failed: %v\n%s", childErr, output)
	}
}

// The child leaves module-cache directory permissions for TestMain to clean;
// neither testing.T cleanup nor the fixture itself repairs those permissions.
func TestReadonlyHomeChild(t *testing.T) {
	report := os.Getenv("CURATOR_TEST_READONLY_HOME")
	if report == "" {
		return
	}
	home := os.Getenv("HOME")
	module := filepath.Join(home, "go", "pkg", "mod", "example.invalid", "module@v1.0.0")
	if err := os.MkdirAll(module, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "source.go"), []byte("package fixture\n"), 0o400); err != nil {
		t.Fatal(err)
	}
	if target := os.Getenv("CURATOR_TEST_EXTERNAL_HOME_LINK"); target != "" {
		if err := os.Symlink(target, filepath.Join(home, "external")); err != nil {
			t.Fatal(err)
		}
		if err := os.Link(filepath.Join(target, "operator"), filepath.Join(home, "external-hardlink")); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{module, filepath.Dir(module), filepath.Join(home, "go", "pkg", "mod")} {
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(report, []byte(home), 0o600); err != nil {
		t.Fatal(err)
	}
}
