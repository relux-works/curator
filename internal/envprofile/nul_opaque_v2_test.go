package envprofile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/pathboundary"
)

// TestInstallAdmitsV1CollidingSkillTreesUnderV2 drives Install through
// contextresolve.Resolve -> auditAndStore -> auditMember with the v1
// framing collision pair. Under v2 the NUL bytes are ordinary data
// (Spec §8): both members install under a v2 lock. Git skill members are
// commit-pinned, so the twin-distinguishing proof for the shared audit
// gate lives in the audit and install packages; here the v1 collision
// preamble ties the admitted fixtures to the pair the v1 lane refuses.
// The v1 refusal of the same pair lives in nul_opaque_test.go.
func TestInstallAdmitsV1CollidingSkillTreesUnderV2(t *testing.T) {
	enableV2WritersForTest(t)
	pinHomes(t)
	ids := newGitIdentities(t)
	single := map[string]string{
		"assets/a.bin": "x\x00docs/b.md\x00y\x00z",
	}
	split := map[string]string{
		"assets/a.bin": "x",
		"docs/b.md":    "y\x00z",
	}
	left, right := t.TempDir(), t.TempDir()
	writeV1Files(t, left, single)
	writeV1Files(t, right, split)
	leftHash, err := hashing.ContentSHA256(left, nil)
	if err != nil {
		t.Fatal(err)
	}
	rightHash, err := hashing.ContentSHA256(right, nil)
	if err != nil {
		t.Fatal(err)
	}
	if leftHash != rightHash {
		t.Fatalf("v1 framing did not collide: %s != %s", leftHash, rightHash)
	}

	home := t.TempDir()
	skillARepo := gitRepo(t, single, "v1.0.0")
	skillAOperand := ids.serve(skillARepo, "https://example.com/v2-single-skill")
	skillBRepo := gitRepo(t, split, "v1.0.0")
	skillBOperand := ids.serve(skillBRepo, "https://example.com/v2-split-skill")
	rootRepo := gitRepo(t, twinDependencyRootFiles(skillAOperand, skillBOperand), "v1.0.0")
	rootOperand := ids.serve(rootRepo, "https://example.com/v2-twin-root")

	info, _, _, err := Install(home, InstallOptions{Operand: rootOperand})
	if err != nil {
		t.Fatalf("v2 Install of NUL-bearing twins: %v", err)
	}
	if info.Lock.ContentHashVersion() != hashing.VersionV2 {
		t.Fatalf("lock content hash version = %d, want v2", info.Lock.ContentHashVersion())
	}
	if memberA, ok := lockMember(info.Lock, "ska"); !ok || memberA.Kind != contextlock.KindSkill {
		t.Fatalf("installed lock lacks skill member ska: %+v", info.Lock.Members)
	}
	if memberB, ok := lockMember(info.Lock, "skb"); !ok || memberB.Kind != contextlock.KindSkill {
		t.Fatalf("installed lock lacks skill member skb: %+v", info.Lock.Members)
	}
}

// TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2 mirrors the v1
// refusal of a deep NUL file with the v2 admission: the context member
// installs and publishes its profile source.
func TestInstallAdmitsDeepNULFileInContextPathSnapshotUnderV2(t *testing.T) {
	enableV2WritersForTest(t)
	home := t.TempDir()
	pinHomes(t)
	source := t.TempDir()
	writeGitFile(t, source, "agent-context.json", `{"schema_version": 1, "name": "nul-context", "version": "1.0.0",`+
		`"context": {"modules": [{"path": "a.md"}]}}`+"\n")
	writeGitFile(t, source, "context/a.md", "clean\n")
	deep := filepath.Join(source, "assets", "deep", "image.unsupported")
	if err := os.MkdirAll(filepath.Dir(deep), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(deep, []byte{1, 0, 2}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(source); err != nil {
		t.Fatal(err)
	}

	info, _, _, err := Install(home, InstallOptions{Operand: source})
	if err != nil {
		t.Fatalf("v2 Install of deep-NUL context: %v", err)
	}
	if info.Lock.ContentHashVersion() != hashing.VersionV2 {
		t.Fatalf("lock content hash version = %d, want v2", info.Lock.ContentHashVersion())
	}
	if _, err := readSource(home, "nul-context"); err != nil {
		t.Fatalf("admitted context install published no profile source: %v", err)
	}
}

func twinDependencyRootFiles(skillAOperand, skillBOperand string) map[string]string {
	return map[string]string{
		"agent-context.json": `{"schema_version": 1, "name": "twin-root", "version": "1.0.0",` +
			`"context": {"modules": [{"path": "a.md"}]},` +
			`"requires": {"skills": {"ska": {"git": "` + skillAOperand + `", "range": "*"},` +
			`"skb": {"git": "` + skillBOperand + `", "range": "*"}}}}` + "\n",
		"context/a.md": "clean\n",
	}
}
