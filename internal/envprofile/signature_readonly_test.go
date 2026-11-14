package envprofile

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/gitops"
)

// Real SSH commit/tag signatures verify through the local transport, tampered signed bytes
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
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, scratch)
	}
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
	assertSignatureScratch(t, scratch, tempBefore)
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
	t.Run("present", func(t *testing.T) {
		good := sourceSignerPosture(req, []Info{info})
		if len(good) != 1 || good[0].Verification != "verified" || !good[0].Current {
			t.Fatalf("valid signer verdict = %+v", good)
		}
	})
	// A policy naming another key retains the established enforced-failure verdict.
	_, wrongPublic := statusSignerKey(t)
	req.Policy.SourceSigners = map[string][]contextresolve.Signer{"example.com/signed": {{Type: "ssh", Key: wrongPublic}}}
	failed := sourceSignerPosture(req, []Info{info})
	if len(failed) != 1 || failed[0].Verification != "failed" || failed[0].Current {
		t.Fatalf("failed signer verdict = %+v", failed)
	}
	req.Policy = policy
	member := info.Lock.Members[0]
	cache := newReadOnlyGitManager(home).repoDir(member.Source)
	readObject := func() ([]byte, error) {
		cmd := exec.Command("git", "-C", cache, "cat-file", "-p", member.Commit)
		cmd.Env = gitops.ReadOnlyEnv()
		return cmd.CombinedOutput()
	}
	signed, err := readObject()
	if err != nil || !hasGitSignature("commit", signed) {
		t.Fatalf("signed object control: %v\n%s", err, signed)
	}
	// Forward to the original Git installation, retaining its executable/DLL
	// layout on Windows. Only the copied forwarder is visible on PATH.
	installMissingPinGit(t, "forward", "")
	forwarder, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", filepath.Dir(forwarder))
	if path, err := exec.LookPath("ssh-keygen"); err == nil || !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("missing-verifier fixture exposes ssh-keygen: %q, %v", path, err)
	}
	if got, err := readObject(); err != nil || !bytes.Equal(got, signed) {
		t.Fatalf("missing-verifier fixture cannot read exact signed object: %v\n%s", err, got)
	}
	t.Logf("%s real Git forwarder reads exact signed object with ssh-keygen absent PASS", runtime.GOOS)
	before := readWitnesses(t, home)
	missing := sourceSignerPosture(req, []Info{info})
	if len(missing) != 1 || missing[0].Verification != "unavailable" || !missing[0].Current || missing[0].Diagnostic != "signer_verification_unavailable" || !strings.Contains(missing[0].Hint, "curator profile update signed") {
		t.Fatalf("unavailable signer verdict = %+v", missing)
	}
	if len(missing[0].Hint) > 1200 || strings.ContainsAny(missing[0].Hint, "\r\n\x1b") {
		t.Fatalf("unavailable signer hint was not bounded and escaped: %q", missing[0].Hint)
	}
	assertReadWitnesses(t, home, before)
	t.Logf("%s missing SSH verifier verdict unavailable/current PASS", runtime.GOOS)
	// The production matrix must not count the same unavailable evidence as green.
	status, err := StatusOf(req)
	if err != nil {
		t.Fatal(err)
	}
	if !status.NonCurrent {
		t.Fatal("unavailable verification counted as green")
	}
}

// A real allowed tag over an unsigned commit is current only for the Git
// namespace. A valid signature in another namespace is failed/non-current,
// rather than operationally unavailable; managed-state witnesses stay intact.
func TestReadPuritySSHNamespaceVerdict(t *testing.T) {
	pinHomes(t)
	private, public := statusSignerKey(t)
	repo := statusSignerRepo(t, private)
	if out, err := exec.Command("git", "-C", repo, "-c", "user.name=Curator Test", "-c", "user.email=curator@example.test", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "unsigned namespace fixture").CombinedOutput(); err != nil {
		t.Fatalf("unsigned fixture: %v %s", err, out)
	}
	commit, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	tag, err := exec.Command("git", "-C", repo, "cat-file", "-p", "v1.0.0").Output()
	if err != nil {
		t.Fatal(err)
	}
	data, _, err := splitGitSignedPayload("tag", tag)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	lines[0] = "object " + strings.TrimSpace(string(commit))
	data = []byte(strings.Join(lines, "\n"))
	ids := newGitIdentities(t)
	remote := ids.serve(repo, "https://example.com/namespace")
	for _, namespace := range []string{"git", "not-git"} {
		t.Run(namespace, func(t *testing.T) {
			name := "v1.0.0"
			body := append([]byte(nil), data...)
			signer := exec.Command("ssh-keygen", "-Y", "sign", "-n", namespace, "-f", private)
			signer.Stdin = strings.NewReader(string(body))
			signature, err := signer.Output()
			if err != nil {
				t.Fatal(err)
			}
			object := exec.Command("git", "-C", repo, "hash-object", "-t", "tag", "-w", "--stdin")
			object.Stdin = strings.NewReader(string(append(body, signature...)))
			oid, err := object.Output()
			if err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command("git", "-C", repo, "update-ref", "refs/tags/"+name, strings.TrimSpace(string(oid))).CombinedOutput(); err != nil {
				t.Fatalf("tag fixture: %v %s", err, out)
			}
			home := gitFixtureHome(t)
			if _, _, _, err := Install(home, InstallOptions{Operand: remote, Tag: name}); err != nil {
				t.Fatal(err)
			}
			before := readWitnesses(t, home)
			policy := Policy{SourceSigners: map[string][]contextresolve.Signer{"example.com/namespace": {{Type: "ssh", Key: public}}}}
			status, err := StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), Policy: policy, Detect: func(envregistry.Adapter) string { return "unknown" }})
			if err != nil {
				t.Fatal(err)
			}
			if len(status.SourceSignerPosture) != 1 {
				t.Fatalf("posture rows: %+v", status.SourceSignerPosture)
			}
			row := status.SourceSignerPosture[0]
			want := "verified"
			if namespace != "git" {
				want = "failed"
			}
			if row.Verification != want || row.Current != (namespace == "git") {
				t.Fatalf("namespace rejection became unavailable/current: %+v want %s", row, want)
			}
			assertReadWitnesses(t, home, before)
			t.Logf("%s real SSH namespace %s verdict %s PASS", runtime.GOOS, namespace, want)
		})
	}
}

// Empty global/system configuration and a discovery ceiling supply no Git
// identity. The namespace fixture still reaches both real signer verdicts;
// the shared negative control first proves implicit identity is refused.
func TestReadPuritySSHNamespaceWithoutAmbientIdentity(t *testing.T) {
	runFixturesWithoutAmbientIdentity(t, []string{"TestReadPuritySSHNamespaceVerdict"})
}
