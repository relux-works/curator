package main

import "testing"

func TestRunVersionExitsZero(t *testing.T) {
	if code := run([]string{"--version"}); code != 0 {
		t.Fatalf("run(--version) = %d, want 0", code)
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if code := run([]string{"frobnicate"}); code != 2 {
		t.Fatalf("run(frobnicate) = %d, want 2", code)
	}
}

func TestShellInitPrintsHooks(t *testing.T) {
	for _, shellName := range []string{"zsh", "bash", "powershell"} {
		if code := run([]string{"shell-init", shellName}); code != 0 {
			t.Fatalf("shell-init %s = %d", shellName, code)
		}
	}
	if code := run([]string{"shell-init", "fish"}); code != 2 {
		t.Fatalf("unsupported shell must be usage error")
	}
}

func TestSkillCheckOnTempDir(t *testing.T) {
	// an empty directory fails validation (missing SKILL.md)
	if code := run([]string{"skill", "check", t.TempDir()}); code != 1 {
		t.Fatalf("skill check on empty dir = %d, want 1", code)
	}
}
