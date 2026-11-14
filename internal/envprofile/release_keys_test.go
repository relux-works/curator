package envprofile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextresolve"
)

const platformReleasePublic = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIF9OaCfCki1tEczuA4hE2DaaToEpFteweRW3IBJkwbgT"
const curatorReleasePublic = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIF78aQWisJvEInj43zlVAGDUrzLpj36KEd6FPNqe4BvK"

// These offline public objects prove real tag-signature verification and
// source-specific key separation, without access to any private signing key.
// This does not certify package installation or tag/commit binding. Existing
// tag-or-commit and explicit-revision authority are preserved.
func TestReleaseTagPublicSignatures(t *testing.T) {
	platform := contextresolve.Signer{Type: "ssh", Key: platformReleasePublic}
	curator := contextresolve.Signer{Type: "ssh", Key: curatorReleasePublic}
	for _, tc := range []struct {
		name, fixture string
		allowed       []contextresolve.Signer
		valid         bool
		unsigned      bool
	}{
		{"platform-release-tag", "platform", []contextresolve.Signer{platform}, true, false},
		{"curator-family-tag", "curator", []contextresolve.Signer{curator}, true, false},
		{"commit-key-tag-refused", "commit-key", []contextresolve.Signer{platform, curator}, false, false},
		{"platform-on-curator-refused", "platform", []contextresolve.Signer{curator}, false, false},
		{"curator-on-platform-refused", "curator", []contextresolve.Signer{platform}, false, false},
		{"unsigned-platform-tag-refused", "platform", []contextresolve.Signer{platform}, false, true},
		{"tampered-platform-tag-refused", "platform", []contextresolve.Signer{platform}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			gitRun(t, repo, "init", "--bare")
			payload, err := os.ReadFile(filepath.Join("testdata", "release-tags", tc.fixture+".tag"))
			if err != nil {
				t.Fatal(err)
			}
			if tc.unsigned {
				payload = []byte(strings.Split(string(payload), "-----BEGIN SSH SIGNATURE-----")[0])
			}
			if strings.HasPrefix(tc.name, "tampered-") {
				payload = []byte(strings.Replace(string(payload), "\n-----BEGIN SSH SIGNATURE-----", "\ntampered\n-----BEGIN SSH SIGNATURE-----", 1))
			}
			cmd := exec.Command("git", "-C", repo, "hash-object", "-t", "tag", "-w", "--stdin")
			cmd.Stdin = strings.NewReader(string(payload))
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("import public tag: %v: %s", err, output)
			}
			beforeConfig, err := os.ReadFile(filepath.Join(repo, "config"))
			if err != nil {
				t.Fatal(err)
			}
			beforeRefs := fixtureGitOutput(t, repo, "for-each-ref")
			home := t.TempDir()
			evidence, err := newGitManager(home).verifyObjectSignature(repo, "tag", strings.TrimSpace(string(output)), tc.allowed)
			if err != nil || evidence.Present == tc.unsigned || evidence.Valid != tc.valid {
				t.Fatalf("signature = %+v, error = %v, want valid=%t", evidence, err, tc.valid)
			}
			fingerprints := map[string]string{
				"platform":   "ssh-ed25519 SHA256:x3t8hOxl+PixTw1eTqV+qCqSp88Vw7AEMrmGUvs213I",
				"curator":    "ssh-ed25519 SHA256:H9oEmLBNf0JIQvCEarIqbcSaUaQCidesH6rl2+rcl0c",
				"commit-key": "ssh-ed25519 SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA",
			}
			wantFingerprint := fingerprints[tc.fixture]
			if tc.unsigned {
				wantFingerprint = "unknown"
			}
			if got := contextresolve.FormatSigner(evidence.Signer); got != wantFingerprint {
				t.Fatalf("claimed signer=%s, want %s", got, wantFingerprint)
			}
			afterConfig, err := os.ReadFile(filepath.Join(repo, "config"))
			if err != nil || string(afterConfig) != string(beforeConfig) || fixtureGitOutput(t, repo, "for-each-ref") != beforeRefs {
				t.Fatalf("verification changed repository config or refs: %v", err)
			}
			if _, err := os.Lstat(filepath.Join(home, "profiles")); !os.IsNotExist(err) {
				t.Fatalf("verification published profile state: %v", err)
			}
			_, reported := verifiedAllowedSigner(contextresolve.CandidateSignatures{Tag: evidence}, tc.allowed)
			if reported != tc.valid {
				t.Fatalf("status signer admission=%t, want %t", reported, tc.valid)
			}
		})
	}
}
