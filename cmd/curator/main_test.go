package main

import "testing"

func TestRunVersionExitsZero(t *testing.T) {
	if code := run([]string{"--version"}); code != 0 {
		t.Fatalf("run(--version) = %d, want 0", code)
	}
}

func TestRunNoArgsExitsUsage(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Fatalf("run() = %d, want 2", code)
	}
}

func TestRunUnknownCommandExitsUsage(t *testing.T) {
	if code := run([]string{"frobnicate"}); code != 2 {
		t.Fatalf("run(frobnicate) = %d, want 2", code)
	}
}
