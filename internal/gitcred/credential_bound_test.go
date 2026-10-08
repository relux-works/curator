package gitcred

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCredentialAnswerBoundRefusesOversized runs the real Git executable
// against its built-in store helper, seeded only with a synthetic
// .git-credentials file in a temporary home. The store helper ships with Git
// on every platform, so this reproduction runs on Windows too.
func TestCredentialAnswerBoundRefusesOversized(t *testing.T) {
	for _, n := range []int{32, maxAnswerBytes - 256, maxAnswerBytes + 128} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			access, home := realGitAccess(t, "[credential]\n\thelper = store\n")
			secret := strings.Repeat("x", n)
			line := "https://synthetic:" + secret + "@fixture.invalid\n"
			if err := os.WriteFile(filepath.Join(home, ".git-credentials"), []byte(line), 0600); err != nil {
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

// TestCredentialExactFrameBound drives the production entry Access.ReadHost
// against a controlled Git executable that answers with an exact byte count,
// without relying on any helper's output formatting. The controlled Git is
// this test binary re-executed through the package's fake-git fixture, so the
// exact cap-1/cap/cap+1 frames run on every platform without a shell script.
func TestCredentialExactFrameBound(t *testing.T) {
	for _, delta := range []int{-1, 0, 1} {
		t.Run(fmt.Sprint(delta), func(t *testing.T) {
			access, dir := fakeAccess(t, modeFixedAnswer)
			prefix := "username=synthetic\npassword="
			n := maxAnswerBytes + delta - len(prefix) - 2
			if err := os.WriteFile(filepath.Join(dir, "answer.bin"), []byte(prefix+strings.Repeat("x", n)+"\n\n"), 0600); err != nil {
				t.Fatal(err)
			}
			got, ok := access.ReadHost(context.Background(), "fixture.invalid")
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
