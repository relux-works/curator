package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/envprofile"
)

type sourceSignerVectorFile struct {
	VerificationCases []struct {
		Name     string `json:"name"`
		Expected struct {
			Verdict    string `json:"verdict"`
			Diagnostic string `json:"diagnostic"`
		} `json:"expected"`
	} `json:"verification_cases"`
}

func loadSourceSignerVector(t *testing.T, name string) (string, string) {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-source-signers.json")) // #nosec G304 -- explicit conformance root
	if err != nil {
		t.Fatal(err)
	}
	var vectors sourceSignerVectorFile
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, vector := range vectors.VerificationCases {
		if vector.Name == name {
			return vector.Expected.Verdict, vector.Expected.Diagnostic
		}
	}
	t.Fatalf("pinned rc.13 signer vector %q is absent", name)
	return "", ""
}

func generateSSHKey(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen is unavailable")
	}
	private := filepath.Join(t.TempDir(), "signer")
	if output, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", private).CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen: %v\n%s", err, output)
	}
	output, err := exec.Command("ssh-keygen", "-y", "-f", private).Output()
	if err != nil {
		t.Fatal(err)
	}
	return private, strings.TrimSpace(string(output))
}

func signerContextRepo(t *testing.T, signingKey string, signCommit, signTag bool) (string, string) {
	t.Helper()
	repo := t.TempDir()
	writeGitRepoFile(t, repo, "agent-context.json", `{"schema_version":1,"name":"signed","version":"1.0.0"}`+"\n")
	runGitRepo(t, repo, "init")
	runGitRepo(t, repo, "add", ".")
	commitArgs := []string{}
	if signCommit {
		commitArgs = append(commitArgs, "-c", "gpg.format=ssh", "-c", "user.signingkey="+signingKey)
	}
	commitArgs = append(commitArgs, "commit")
	if signCommit {
		commitArgs = append(commitArgs, "-S")
	}
	commitArgs = append(commitArgs, "-m", "release")
	runGitRepo(t, repo, commitArgs...)
	commit := strings.TrimSpace(string(mustGitOutput(t, repo, "rev-parse", "HEAD")))
	tagArgs := []string{}
	if signTag {
		tagArgs = append(tagArgs, "-c", "gpg.format=ssh", "-c", "user.signingkey="+signingKey)
	}
	tagArgs = append(tagArgs, "tag")
	if signTag {
		tagArgs = append(tagArgs, "-s", "-m", "release")
	}
	tagArgs = append(tagArgs, "v1.0.0")
	runGitRepo(t, repo, tagArgs...)
	return repo, commit
}

func mustGitOutput(t *testing.T, repo string, args ...string) []byte {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = repo
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	return output
}

func configureSignerPolicy(source stubConfigSource, canonical, publicKey string, required bool) stubConfigSource {
	source.cfg.Env.SourceSigners = map[string][]config.SourceSigner{
		canonical: {{Type: "ssh", Key: publicKey + " operator@example"}},
	}
	source.cfg.Env.RequireSourceSigners = required
	return source
}

// TestSourceSignerVectorsAtProfileInstall binds the pinned unsigned, wrong
// signer, accepted SSH tag, and empty-allowlist vectors to the production
// profile install entry point and real Git signature verification.
func TestSourceSignerVectorsAtProfileInstall(t *testing.T) {
	requireGit(t)
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen is unavailable")
	}
	allowedPrivate, allowedPublic := generateSSHKey(t)
	wrongPrivate, _ := generateSSHKey(t)
	tests := []struct {
		name       string
		vector     string
		signCommit bool
		signTag    bool
		wrong      bool
		empty      bool
	}{
		{name: "unsigned-refused", vector: "unsigned-refused"},
		{name: "wrong-signer-refused", vector: "wrong-signer-refused", signTag: true, wrong: true},
		{name: "ssh-tag-signature-accepted", vector: "ssh-tag-signature-accepted", signTag: true},
		{name: "empty-allowlist-signed-refused", vector: "empty-allowlist-signed-refused", signTag: true, empty: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verdict, diagnostic := loadSourceSignerVector(t, tc.vector)
			source, home := profileHome(t)
			signingKey := allowedPrivate
			if tc.wrong {
				signingKey = wrongPrivate
			}
			repo, _ := signerContextRepo(t, signingKey, tc.signCommit, tc.signTag)
			const rawSource = "https://example.com/signed"
			serveGitRepos(t, map[string]string{rawSource: repo})
			if tc.empty {
				source.cfg.Env.SourceSigners = map[string][]config.SourceSigner{"example.com/signed": {}}
			} else {
				source = configureSignerPolicy(source, "example.com/signed", allowedPublic, false)
			}
			args := []string{"profile", "install", rawSource}
			code, stdout, stderr := runProfile(t, source, args...)
			if verdict == "accepted" {
				if code != exitOK {
					t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
				}
				if _, err := os.Stat(filepath.Join(envprofile.ProfileDir(home, "signed"), "lock.json")); err != nil {
					t.Fatalf("accepted install did not publish its lock: %v", err)
				}
				if tc.name == "ssh-tag-signature-accepted" {
					statusCode, statusJSON, statusErr := runProfile(t, source, "env", "status", "--json")
					if statusCode != exitOK {
						t.Fatalf("status = %d: %s\n%s", statusCode, statusJSON, statusErr)
					}
					var status envprofile.Status
					if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
						t.Fatal(err)
					}
					if len(status.SourceSignerPosture) == 0 || status.SourceSignerPosture[0].State != "enforced" || !status.SourceSignerPosture[0].Current || !strings.HasPrefix(status.SourceSignerPosture[0].Signer, "ssh-ed25519 SHA256:") {
						t.Fatalf("status signer posture = %+v", status.SourceSignerPosture)
					}
					humanCode, humanStatus, humanErr := runProfile(t, source, "env", "status")
					if humanCode != exitOK || !strings.Contains(humanStatus, "require_source_signers: false") ||
						!strings.Contains(humanStatus, "update_confirmation: B-flip") ||
						!strings.Contains(humanStatus, "source_signers profile signed context signed example.com/signed: enforced (ssh-ed25519 SHA256:") {
						t.Fatalf("human status = %d\n%s\n%s\n%s", humanCode, humanStatus, humanErr, statusErr)
					}
				}
				return
			}
			if code != exitFail || !strings.Contains(stderr, diagnostic) {
				t.Fatalf("install = %d, diagnostic %q missing\nstdout:\n%s\nstderr:\n%s", code, diagnostic, stdout, stderr)
			}
			if _, err := os.Stat(filepath.Join(envprofile.ProfileDir(home, "signed"), "lock.json")); !os.IsNotExist(err) {
				t.Fatalf("refused signer candidate left a lock: err=%v", err)
			}
		})
	}
}

// TestRevisionDoesNotBorrowTagSignature narrows the exact-revision rule: a
// signed tag cannot authorize an unsigned peeled commit selected by revision.
func TestRevisionDoesNotBorrowTagSignature(t *testing.T) {
	requireGit(t)
	allowedPrivate, allowedPublic := generateSSHKey(t)
	repo, commit := signerContextRepo(t, allowedPrivate, false, true)
	source, home := profileHome(t)
	const rawSource = "https://example.com/signed"
	serveGitRepos(t, map[string]string{rawSource: repo})
	source = configureSignerPolicy(source, "example.com/signed", allowedPublic, false)
	code, stdout, stderr := runProfile(t, source, "profile", "install", rawSource, "--revision", commit)
	if code != exitFail || !strings.Contains(stderr, "context_source_unsigned") {
		t.Fatalf("revision install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if _, err := os.Stat(filepath.Join(envprofile.ProfileDir(home, "signed"), "lock.json")); !os.IsNotExist(err) {
		t.Fatalf("unsigned revision unexpectedly published a lock: %v", err)
	}
}
