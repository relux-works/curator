//go:build unix

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Pausing only the actual descendant go build leaves profile sync available.
// After resuming the real compiler, both canonical and native owners publish.
// Removal followed by add also restores the compiled skill without concurrency.
// This exercises the production CLI, trusted toolchain and profile sync paths.
func TestGlobalNativeOwnerPausedBuild(t *testing.T) {
	fixture := setupNativeConcurrent(t, false)
	compiled := t.TempDir()
	writeCompiledSkillRepo(t, compiled)
	gitConfig := os.Getenv("GIT_CONFIG_GLOBAL")
	f, err := os.OpenFile(gitConfig, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("\n[url \"" + filepath.ToSlash(compiled) + "\"]\n\tinsteadOf = https://example.com/build-skill\n")
	if closeErr := f.Close(); err != nil || closeErr != nil {
		t.Fatalf("write fixture mapping: %v %v", err, closeErr)
	}
	writeFile(t, filepath.Join(fixture.home, ".curator", "global", "Skillfile.json"), `{"schema_version":1,"agents":["codex_cli","claude_code"],"skills":[{"name":"build-skill","git":"https://example.com/build-skill","tag":"v1"}]}`)
	child := startNativeCommand(t, fixture, "global", "install")
	pid := pauseNativeGoBuild(t, child)
	resumed := false
	t.Cleanup(func() {
		if !resumed {
			_ = syscall.Kill(pid, syscall.SIGCONT)
		}
	})
	fixture.command(t, 0, "profile", "sync")
	if _, err := os.Lstat(filepath.Join(fixture.home, ".curator", "global", "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("paused build published canonical state: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(fixture.home, ".codex", "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("paused build published native state: %v", err)
	}
	if err := syscall.Kill(pid, syscall.SIGCONT); err != nil {
		t.Fatal(err)
	}
	resumed = true
	child.wait(t)
	requireCompiledNativeOwner(t, fixture, child)
	fixture.command(t, 0, "global", "remove", "build-skill")
	fixture.command(t, 0, "global", "install")
	if _, err := os.Lstat(filepath.Join(fixture.home, ".curator", "global", "skills", "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("removed compiled skill remained canonical: %v", err)
	}
	fixture.command(t, 0, "global", "add", "build-skill", "--git", "https://example.com/build-skill", "--tag", "v1")
	requireCompiledNativeOwner(t, fixture, child)
}

func requireCompiledNativeOwner(t *testing.T, fixture nativeGlobalFixture, child *nativeCommand) {
	t.Helper()
	for _, path := range []string{".curator/global/skills/build-skill", ".codex/skills/build-skill", ".claude/skills/build-skill"} {
		if _, err := os.Stat(filepath.Join(fixture.home, filepath.FromSlash(path))); err != nil {
			t.Fatalf("resumed build did not publish %s: %v\n%s", path, err, child.output.String())
		}
	}
}

func pauseNativeGoBuild(t *testing.T, child *nativeCommand) int {
	t.Helper()
	deadline := time.Now().Add(40 * time.Second)
	for time.Now().Before(deadline) {
		// Read only process IDs, parent IDs and executable names globally. Read
		// arguments only for descendants owned by this fixture's CLI process.
		output, err := exec.Command("ps", "-axo", "pid=,ppid=,comm=").Output()
		if err != nil {
			t.Fatal(err)
		}
		descendants := map[int]bool{child.cmd.Process.Pid: true}
		rows := strings.Split(string(output), "\n")
		for changed := true; changed; {
			changed = false
			for _, row := range rows {
				fields := strings.Fields(row)
				if len(fields) < 3 {
					continue
				}
				pid, _ := strconv.Atoi(fields[0])
				parent, _ := strconv.Atoi(fields[1])
				if !descendants[parent] || descendants[pid] {
					continue
				}
				descendants[pid] = true
				changed = true
				if filepath.Base(fields[len(fields)-1]) != "go" {
					continue
				}
				args, _ := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=").Output()
				if !strings.Contains(string(args), " build ") {
					continue
				}
				if err := syscall.Kill(pid, syscall.SIGSTOP); err != nil {
					continue
				}
				t.Logf("paused owned go build PID %d", pid)
				return pid
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("actual go build was not observed before deadline")
	return 0
}
