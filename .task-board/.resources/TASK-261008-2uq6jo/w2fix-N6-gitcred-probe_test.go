package gitcred

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Uses the real Git executable with only a synthetic helper in a temporary home.
func TestWave2CredentialAnswerBound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture helper is a POSIX shell script")
	}
	for _, n := range []int{32, maxAnswerBytes - 80, maxAnswerBytes + 128} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			access, home := realGitAccess(t, "[credential]\n\thelper =\n")
			helper := filepath.Join(home, "helper")
			reply := filepath.Join(home, "reply")
			secret := strings.Repeat("x", n)
			if err := os.WriteFile(reply, []byte("username=synthetic\npassword="+secret+"\n\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(helper, []byte("#!/bin/sh\n/bin/cat '"+reply+"'\n"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[credential]\n\thelper =\n\thelper = "+helper+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			got, ok := access.ReadHost(context.Background(), "fixture.invalid")
			t.Logf("ReadHost requested_secret_bytes=%d returned_secret_bytes=%d present=%v", n, len(got.Secret), ok)
			if n > maxAnswerBytes {
				if ok {
					t.Fatalf("oversized helper answer accepted as credential (truncated=%v)", got.Secret != secret)
				}
				return
			}
			if !ok || got.Secret != secret {
				t.Fatal("bounded answer did not survive unchanged")
			}
		})
	}
}

// Access.ReadHost is the production entry. This controlled Git executable tests
// the exact cap including format bytes, without relying on Git output ordering.
func TestWave2CredentialExactFrameBound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture")
	}
	for _, delta := range []int{-1, 0, 1} {
		t.Run(fmt.Sprint(delta), func(t *testing.T) {
			root := t.TempDir()
			output := filepath.Join(root, "answer")
			git := filepath.Join(root, "git")
			prefix := "username=synthetic\npassword="
			n := maxAnswerBytes + delta - len(prefix) - 2
			if err := os.WriteFile(output, []byte(prefix+strings.Repeat("x", n)+"\n\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(git, []byte("#!/bin/sh\n/bin/cat '"+output+"'\n"), 0700); err != nil {
				t.Fatal(err)
			}
			got, ok := (Access{Git: git, Home: root, Environ: []string{}}).ReadHost(context.Background(), "fixture.invalid")
			t.Logf("answer_bytes=%d present=%v secret_bytes=%d", maxAnswerBytes+delta, ok, len(got.Secret))
			if delta > 0 {
				if ok {
					t.Fatal("cap+1 frame was accepted")
				}
				return
			}
			if !ok || len(got.Secret) != n {
				t.Fatal("at/below-cap answer failed")
			}
		})
	}
}
