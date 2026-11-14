package envprofile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envregistry"
)

// Real SSH commit/tag signatures verify through pipes, tampered signed bytes
// fail, and the mutating verifier remains usable. No private material is read
// during the status phase; fixture key creation is completed beforehand.
func TestReadPuritySSHSignatures(t *testing.T) {
	if runtime.GOOS == "plan9" {
		t.Skip("read-only signature input transport unavailable")
	}
	pinHomes(t)
	private, public := statusSignerKey(t)
	repo := statusSignerRepo(t, private)
	ids := newGitIdentities(t)
	remote := ids.serve(repo, "https://example.com/signed")
	home := gitFixtureHome(t)
	policy := Policy{SourceSigners: map[string][]contextresolve.Signer{"example.com/signed": {{Type: "ssh", Key: public}}}}
	if _, _, _, err := Install(home, InstallOptions{Operand: remote, Tag: "v1.0.0", Policy: policy}); err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	t.Setenv("TMPDIR", scratch)
	before := readWitnesses(t, home)
	tempBefore := readWitnesses(t, scratch)
	status, err := StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), Policy: policy, Detect: func(envregistry.Adapter) string { return "unknown" }})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range status.SourceSignerPosture {
		if row.Source == "example.com/signed" {
			found = true
			if !row.Current || row.Signer == "unknown" || row.Verification != "verified" {
				t.Errorf("SSH verification posture = %+v", row)
			}
		}
	}
	if !found {
		t.Fatal("signed profile absent from posture")
	}
	assertReadWitnesses(t, home, before)
	assertReadWitnesses(t, scratch, tempBefore)
	for _, objectType := range []string{"commit", "tag"} {
		object := "HEAD"
		if objectType == "tag" {
			object = "v1.0.0"
		}
		payload, err := exec.Command("git", "-C", repo, "cat-file", "-p", object).Output()
		if err != nil {
			t.Fatal(err)
		}
		allowed := []contextresolve.Signer{{Type: "ssh", Key: public}}
		evidence, err := verifyReadOnlyObjectSignature(objectType, payload, allowed)
		if err != nil || !evidence.Valid {
			t.Fatalf("valid %s signature = %+v, %v", objectType, evidence, err)
		}
		_, _, splitErr := splitGitSignedPayload(objectType, payload)
		if splitErr != nil {
			t.Fatal(splitErr)
		}
		t.Logf("%s real SSH %s valid control PASS", runtime.GOOS, objectType)
		// Alter a signed header, preserving the complete embedded signature.
		var changed string
		if objectType == "commit" {
			changed = strings.Replace(string(payload), "tree ", "tree 0", 1)
		} else {
			changed = strings.Replace(string(payload), "object ", "object 0", 1)
		}
		evidence, err = verifyReadOnlyObjectSignature(objectType, []byte(changed), allowed)
		if err != nil || evidence.Valid {
			t.Fatalf("tampered %s signature admitted: %+v, %v", objectType, evidence, err)
		}
	}
	t.Logf("%s real SSH tampered commit/tag controls PASS", runtime.GOOS)
	manager := newGitManager(home)
	pinRaw, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	pin := strings.TrimSpace(string(pinRaw))
	evidence, err := manager.verifyObjectSignature(repo, "commit", pin, []contextresolve.Signer{{Type: "ssh", Key: public}})
	if err != nil || !evidence.Valid {
		t.Fatalf("mutating signature verifier = %+v, %v", evidence, err)
	}
}

// Missing public keyring is explicitly unavailable; it is never an invalid
// signature verdict or permission to initialize GNUPGHOME.
func TestReadPurityGPGUnavailable(t *testing.T) {
	base := t.TempDir()
	t.Setenv("GNUPGHOME", filepath.Join(base, "absent"))
	before := readWitnesses(t, base)
	payload := []byte("tree abc\ngpgsig -----BEGIN PGP SIGNATURE-----\n fixture\n -----END PGP SIGNATURE-----\n\nmessage\n")
	evidence, err := verifyReadOnlyObjectSignature("commit", payload, nil)
	if !errors.Is(err, errSignerVerificationUnavailable) || evidence.Valid {
		t.Fatalf("missing keyring = %+v, %v", evidence, err)
	}
	assertReadWitnesses(t, base, before)
	// An unsafe keyring link cannot select a different public-key source.
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Join(base, "absent")); err == nil {
		before = readWitnesses(t, base)
		evidence, err = verifyReadOnlyObjectSignature("commit", payload, nil)
		if !errors.Is(err, errSignerVerificationUnavailable) || evidence.Valid {
			t.Fatalf("linked keyring = %+v, %v", evidence, err)
		}
		assertReadWitnesses(t, base, before)
	}
}

// An unchanged signed pin with a missing verifier retains pin currency and
// reports unavailable explicitly; enforced failures retain their false verdict.
func TestReadPuritySignerVerdictsKeepPin(t *testing.T) {
	if runtime.GOOS == "plan9" {
		t.Skip("read-only signature input transport unavailable")
	}
	pinHomes(t)
	private, public := statusSignerKey(t)
	repo := statusSignerRepo(t, private)
	home := gitFixtureHome(t)
	ids := newGitIdentities(t)
	remote := ids.serve(repo, "https://example.com/signed")
	policy := Policy{SourceSigners: map[string][]contextresolve.Signer{"example.com/signed": {{Type: "ssh", Key: public}}}}
	info, _, _, err := Install(home, InstallOptions{Operand: remote, Policy: policy})
	if err != nil {
		t.Fatal(err)
	}
	req := StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), Policy: policy, Detect: func(envregistry.Adapter) string { return "unknown" }}
	good := sourceSignerPosture(req, []Info{info})
	if len(good) != 1 || good[0].Verification != "verified" || !good[0].Current {
		t.Fatalf("valid signer verdict = %+v", good)
	}
	// A policy naming another key retains the established enforced-failure verdict.
	_, wrongPublic := statusSignerKey(t)
	req.Policy.SourceSigners = map[string][]contextresolve.Signer{"example.com/signed": {{Type: "ssh", Key: wrongPublic}}}
	failed := sourceSignerPosture(req, []Info{info})
	if len(failed) != 1 || failed[0].Verification != "failed" || failed[0].Current {
		t.Fatalf("failed signer verdict = %+v", failed)
	}
	req.Policy = policy
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	onlyGit := t.TempDir()
	if runtime.GOOS == "windows" {
		data, err := os.ReadFile(git)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(onlyGit, "git.exe"), data, 0o700); err != nil {
			t.Fatal(err)
		}
	} else if err := os.Symlink(git, filepath.Join(onlyGit, "git")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", onlyGit)
	before := readWitnesses(t, home)
	missing := sourceSignerPosture(req, []Info{info})
	if len(missing) != 1 || missing[0].Verification != "unavailable" || !missing[0].Current || missing[0].Diagnostic != "signer_verification_unavailable" || !strings.Contains(missing[0].Hint, "curator profile update signed") {
		t.Fatalf("unavailable signer verdict = %+v", missing)
	}
	if len(missing[0].Hint) > 1200 || strings.ContainsAny(missing[0].Hint, "\r\n\x1b") {
		t.Fatalf("unavailable signer hint was not bounded and escaped: %q", missing[0].Hint)
	}
	assertReadWitnesses(t, home, before)
	// The production matrix must not count the same unavailable evidence as green.
	status, err := StatusOf(req)
	if err != nil {
		t.Fatal(err)
	}
	if !status.NonCurrent {
		t.Fatal("unavailable verification counted as green")
	}
}
