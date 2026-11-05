//go:build darwin || linux

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

// Real install entrypoints must propagate Ctrl-C/termination into Build. The
// worker is causally paused before compilation: cancellation must kill it even
// while stopped and must publish no compiled skill. Separate driver witnesses
// cover manager SIGKILL and running compiler/tool descendants.
func TestNativeInstallSignalCancellation(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	for _, signal := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(strconv.Itoa(int(signal)), func(t *testing.T) {
			fixture := setupNativeConcurrent(t, false)
			compiled := t.TempDir()
			writeFile(t, filepath.Join(compiled, "assets", "build-tool", "cmd", "tool", nativeBuildPauseMarker), "pause\n")
			writeCompiledSkillRepo(t, compiled)
			configureNativeCompiledSkill(t, fixture, compiled)
			child := startNativeCommand(t, fixture, "global", "install")
			worker := waitNativeBuildWorker(t, child)
			t.Cleanup(func() { _ = syscall.Kill(worker, syscall.SIGKILL) })
			if err := child.cmd.Process.Signal(signal); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-child.done:
				child.finished = true
				if err == nil || !strings.Contains(child.output.String(), "process_canceled") {
					t.Fatalf("install signal diagnostic: %v\n%s", err, child.output.String())
				}
			case <-time.After(10 * time.Second):
				t.Fatal("install did not cancel promptly")
			}
			deadline := time.Now().Add(10 * time.Second)
			for syscall.Kill(worker, 0) == nil && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}
			if syscall.Kill(worker, 0) == nil {
				state, _ := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(worker)).Output()
				if !strings.Contains(string(state), "Z") {
					t.Fatalf("interrupted install left worker %d running: %s", worker, state)
				}
			}
			for _, path := range []string{".curator/global/skills/build-skill", ".codex/skills/build-skill", ".claude/skills/build-skill"} {
				if _, err := os.Lstat(filepath.Join(fixture.home, filepath.FromSlash(path))); !os.IsNotExist(err) {
					t.Fatalf("interrupted install published %s: %v", path, err)
				}
			}
		})
	}
}
