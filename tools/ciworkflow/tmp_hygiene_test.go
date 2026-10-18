package ciworkflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const restoreCommand = "bash scripts/restore-runner-tmp.sh"
const restoreCondition = "always() && runner.environment == 'self-hosted' && runner.os == 'Linux'"

func checkTemporaryTreeRestore(data []byte) error {
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return err
	}
	count := 0
	for name, job := range wf.Jobs {
		route := fmt.Sprint(job.RunsOn)
		if !strings.Contains(route, "LINUX_SELF_HOSTED") && !strings.Contains(route, "relux-linux") {
			continue
		}
		count++
		if len(job.Steps) == 0 {
			return fmt.Errorf("%s: missing final tmp restore", name)
		}
		last := job.Steps[len(job.Steps)-1]
		if last.Run != restoreCommand || last.If != restoreCondition || last.Shell != "bash" || !last.Continue {
			return fmt.Errorf("%s: missing guarded final tmp restore", name)
		}
	}
	if count != 7 {
		return fmt.Errorf("expected seven Linux-routable jobs, got %d", count)
	}
	return nil
}

// Every current Linux-routable job ends with the guarded non-blocking shared script call.
func TestLinuxCIWorkflowTemporaryTreeRestore(t *testing.T) {
	if err := checkTemporaryTreeRestore(workflowBytes(t)); err != nil {
		t.Fatal(err)
	}
}

// Per-job removal and token-preserving invocation/condition/order plants are refused by the live workflow gate.
func TestLinuxCIWorkflowTemporaryTreeRestorePlants(t *testing.T) {
	data := workflowBytes(t)
	if err := checkTemporaryTreeRestore(data); err != nil {
		t.Fatal(err)
	}
	var wf workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatal(err)
	}
	for name, job := range wf.Jobs {
		if !strings.Contains(fmt.Sprint(job.RunsOn), "LINUX_SELF_HOSTED") {
			continue
		}
		t.Run(name, func(t *testing.T) {
			for _, plant := range []string{"remove", "echo", "success-only", "hosted", "not-final", "blocking", "wrong-shell"} {
				t.Run(plant, func(t *testing.T) {
					var mutant workflow
					if err := yaml.Unmarshal(data, &mutant); err != nil {
						t.Fatal(err)
					}
					j := mutant.Jobs[name]
					n := len(j.Steps) - 1
					switch plant {
					case "remove":
						j.Steps = j.Steps[:n]
					case "echo":
						j.Steps[n].Run = "echo '" + restoreCommand + "'"
					case "success-only":
						j.Steps[n].If = strings.Replace(restoreCondition, "always()", "success()", 1)
					case "hosted":
						j.Steps[n].If = strings.Replace(restoreCondition, "self-hosted", "github-hosted", 1)
					case "not-final":
						j.Steps = append(j.Steps, step{Run: "echo later"})
					case "blocking":
						j.Steps[n].Continue = false
					case "wrong-shell":
						j.Steps[n].Shell = "sh"
					}
					mutant.Jobs[name] = j
					raw, err := yaml.Marshal(mutant)
					if err != nil {
						t.Fatal(err)
					}
					if err := checkTemporaryTreeRestore(raw); err == nil || !strings.Contains(err.Error(), name+":") {
						t.Fatalf("plant admitted: %v", err)
					}
				})
			}
		})
	}
}

func restoreScript() string {
	if path := os.Getenv("CI_TMP_SCRIPT_UNDER_TEST"); path != "" {
		return path
	}
	return "../../scripts/restore-runner-tmp.sh"
}

func runRestore(script string, env []string) ([]byte, error) {
	cmd := exec.Command("bash", script)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "TMPDIR=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, env...)
	return cmd.CombinedOutput()
}

// The real script heals both current and stale read-only trees without deleting or following external links.
func TestRunnerTemporaryTreeSelfHeal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Linux chmod contract is exercised on Linux and macOS")
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("untouched"), 0o400); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(outside, 0o600) })
	for _, name := range []string{"current", "stale"} {
		dir := filepath.Join(root, name, "nested")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "payload"), []byte(name), 0o400); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(dir, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	out, err := runRestore(restoreScript(), []string{"TMPDIR=" + root})
	if err != nil {
		t.Fatalf("restore exit: %v: %s", err, out)
	}
	for _, name := range []string{"current", "stale"} {
		dir := filepath.Join(root, name, "nested")
		for _, path := range []string{dir, filepath.Join(dir, "payload")} {
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm()&0o200 == 0 {
				t.Errorf("not restored: %s: %v", path, err)
			}
		}
		data, err := os.ReadFile(filepath.Join(dir, "payload"))
		if err != nil || string(data) != name {
			t.Errorf("file deleted or changed: %v", err)
		}
	}
	if info, err := os.Stat(outside); err != nil || info.Mode().Perm()&0o200 != 0 {
		t.Errorf("external symlink target changed: %v", err)
	}
	if info, err := os.Lstat(filepath.Join(root, "link")); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Error("link deleted")
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatalf("tree not removable: %v", err)
	}
}

// Unsafe roots log refusal and exit zero before chmod; a chmod spy keeps filesystem-root plants harmless.
func TestRunnerTemporaryTreeRefusals(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Linux shell contract is exercised on Linux and macOS")
	}
	root := t.TempDir()
	spy := t.TempDir()
	record := filepath.Join(t.TempDir(), "called")
	if err := os.WriteFile(filepath.Join(spy, "chmod"), []byte("#!/bin/sh\nprintf called > \"$CHMOD_RECORD\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("x"), 0o400); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(file, 0o600) })
	link := filepath.Join(root, "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		value *string
	}{
		{"unset", nil}, {"empty", stringPointer("")}, {"root", stringPointer("/")}, {"root-alias", stringPointer("/tmp/..")},
		{"double-slash-root", stringPointer("//")},
		{"file", &file}, {"missing", stringPointer(filepath.Join(root, "absent"))}, {"symlink", &link},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := []string{"PATH=" + spy + string(os.PathListSeparator) + os.Getenv("PATH"), "CHMOD_RECORD=" + record}
			if tc.value != nil {
				env = append(env, "TMPDIR="+*tc.value)
			}
			out, err := runRestore(restoreScript(), env)
			if err != nil || !strings.Contains(string(out), "refusing") {
				t.Fatalf("refusal exit/log: %v: %s", err, out)
			}
			if _, err := os.Stat(record); !os.IsNotExist(err) {
				t.Fatal("refused root reached chmod")
			}
			if info, err := os.Stat(file); err != nil || info.Mode().Perm()&0o200 != 0 {
				t.Fatal("refusal changed fixture")
			}
		})
	}
}
func stringPointer(s string) *string { return &s }

// Canonicalization preserves a trailing pathname newline and never chmods the adjacent trimmed path.
func TestRunnerTemporaryTreeRootTrailingNewline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Linux shell contract is exercised on Linux and macOS")
	}
	base := t.TempDir()
	root := filepath.Join(base, "runner-tmp\n")
	outside := strings.TrimSuffix(root, "\n")
	t.Cleanup(func() {
		for _, path := range []string{root, outside} {
			_ = os.Chmod(path, 0o700)
		}
	})
	for _, path := range []string{root, outside} {
		if err := os.Mkdir(path, 0o500); err != nil {
			t.Fatal(err)
		}
	}
	out, err := runRestore(restoreScript(), []string{"TMPDIR=" + root})
	if err != nil {
		t.Fatalf("restore: %v: %s", err, out)
	}
	if info, err := os.Stat(root); err != nil || info.Mode().Perm()&0o200 == 0 {
		t.Fatalf("newline root not restored: %v: %s", err, out)
	}
	if info, err := os.Stat(outside); err != nil || info.Mode().Perm()&0o200 != 0 {
		t.Fatalf("trimmed sibling changed: %v", err)
	}
	if err := os.RemoveAll(root); err != nil {
		t.Fatal(err)
	}
}

// Permission errors are observable but cannot turn the best-effort final step into a job failure.
func TestRunnerTemporaryTreeChmodFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Linux shell contract is exercised on Linux and macOS")
	}
	spy := t.TempDir()
	if err := os.WriteFile(filepath.Join(spy, "chmod"), []byte("#!/bin/sh\nexit 13\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	out, err := runRestore(restoreScript(), []string{"TMPDIR=" + t.TempDir(), "PATH=" + spy + string(os.PathListSeparator) + os.Getenv("PATH")})
	if err != nil || !strings.Contains(string(out), "could not restore") {
		t.Fatalf("chmod error handling: %v: %s", err, out)
	}
}

// Each script plant retains chmod and its root gate but weakens one boundary; the behavioral test must fail by name.
func TestRunnerTemporaryTreeNarrowingControls(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Linux shell contract is exercised on Linux and macOS")
	}
	raw, err := os.ReadFile("../../scripts/restore-runner-tmp.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []struct{ name, old, replacement, test, finding string }{
		{"current-only", `chmod -R u+w -- "$root"`, `chmod -R u+w -- "$root/current"`, "TestRunnerTemporaryTreeSelfHeal", "not restored"},
		{"admit-root-alias", `if [[ "$root" == / || "$root" -ef / ]]; then`, `if [[ "$root" == / && "${TMPDIR-}" == / ]]; then`, "TestRunnerTemporaryTreeRefusals/root-alias", "refusal exit/log"},
		{"admit-double-slash-root", `if [[ "$root" == / || "$root" -ef / ]]; then`, `if [[ "$root" == / ]]; then`, "TestRunnerTemporaryTreeRefusals/double-slash-root", "refusal exit/log"},
		{"trim-path-newline", `root=$(cd -- "$root" && pwd -P && printf '.')`, `root=$(cd -- "$root" && pwd -P)`, "TestRunnerTemporaryTreeRootTrailingNewline", "newline root not restored"},
	} {
		t.Run(p.name, func(t *testing.T) {
			if strings.Count(string(raw), p.old) != 1 {
				t.Fatal("plant target absent or ambiguous")
			}
			script := filepath.Join(t.TempDir(), "restore.sh")
			if err := os.WriteFile(script, []byte(strings.Replace(string(raw), p.old, p.replacement, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^"+p.test+"$", "-test.v")
			cmd.Env = append(os.Environ(), "CI_TMP_SCRIPT_UNDER_TEST="+script)
			out, runErr := cmd.CombinedOutput()
			exit, ok := runErr.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: "+strings.Split(p.test, "/")[0]) || !strings.Contains(string(out), p.finding) {
				t.Fatalf("survivor or invalid plant: %v: %s", runErr, out)
			}
			t.Logf("%s killed by %s; behavioral child exit 1 (expected red)", p.name, p.test)
		})
	}
}
