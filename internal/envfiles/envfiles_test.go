package envfiles

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteProjectShapesAndSourcing(t *testing.T) {
	project := t.TempDir()
	if err := WriteProject(project); err != nil {
		t.Fatal(err)
	}
	sh, err := os.ReadFile(filepath.Join(project, ".agents", "env.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(sh)
	for _, needle := range []string{"CSK_PROJECT_ROOT", ".agents/bin", "BASH_SOURCE", "ZSH_VERSION"} {
		if !strings.Contains(text, needle) {
			t.Fatalf("env.sh lacks %q:\n%s", needle, text)
		}
	}
	ps1, err := os.ReadFile(filepath.Join(project, ".agents", "env.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ps1), "CSK_PROJECT_ROOT") {
		t.Fatalf("env.ps1 lacks project root:\n%s", ps1)
	}

	if runtime.GOOS == "windows" {
		return
	}
	// sourcing under bash resolves the project root and prepends PATH
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("no bash")
	}
	cmd := exec.Command(bash, "-c", ". '"+filepath.Join(project, ".agents", "env.sh")+"'; printf '%s\n%s' \"$CSK_PROJECT_ROOT\" \"$PATH\"")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.SplitN(string(out), "\n", 2)
	resolved, _ := filepath.EvalSymlinks(project)
	gotRoot, _ := filepath.EvalSymlinks(lines[0])
	if gotRoot != resolved {
		t.Fatalf("CSK_PROJECT_ROOT = %q, want %q", lines[0], project)
	}
	if !strings.HasPrefix(lines[1], lines[0]+"/.agents/bin:") {
		t.Fatalf("PATH not prepended: %q", lines[1])
	}
}

func TestWriteGlobal(t *testing.T) {
	home := t.TempDir()
	if err := WriteGlobal(home); err != nil {
		t.Fatal(err)
	}
	sh, err := os.ReadFile(filepath.Join(home, "global", "env.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sh), "CSK_GLOBAL_ROOT") {
		t.Fatalf("global env.sh:\n%s", sh)
	}
	if !strings.Contains(string(sh), "BASH_SOURCE") || !strings.Contains(string(sh), "ZSH_VERSION") {
		t.Fatalf("global env.sh does not locate itself:\n%s", sh)
	}
	if _, err := os.Stat(filepath.Join(home, "global", "bin")); err != nil {
		t.Fatal("global bin dir missing")
	}
	if runtime.GOOS != "windows" {
		bash, err := exec.LookPath("bash")
		if err != nil {
			t.Skip("no bash")
		}
		cmd := exec.Command(bash, "-c", ". '"+filepath.Join(home, "global", "env.sh")+"'; printf '%s\n%s' \"$CSK_GLOBAL_ROOT\" \"$PATH\"")
		out, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.SplitN(string(out), "\n", 2)
		expectedInfo, expectedErr := os.Stat(filepath.Join(home, "global"))
		actualInfo, actualErr := os.Stat(lines[0])
		if expectedErr != nil || actualErr != nil || !os.SameFile(expectedInfo, actualInfo) {
			t.Fatalf("CSK_GLOBAL_ROOT = %q, want %q", lines[0], filepath.Join(home, "global"))
		}
		if !strings.HasPrefix(lines[1], lines[0]+"/bin:") {
			t.Fatalf("global PATH not prepended: %q", lines[1])
		}
	}
}
