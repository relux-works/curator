package envprofile

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/manifest"
)

// migrationRevisionFixture starts with a legacy branch declaration, then pins
// that branch's commit as a revision. migrateGlobalSkills deliberately leaves
// live branch declarations to the direct-skill pipeline; the audit regression
// is at the next migration of the newly revision-pinned declaration.
func migrationRevisionFixture(t *testing.T, files map[string]string) (home, commit string) {
	t.Helper()
	ids := newGitIdentities(t)
	repo := gitRepo(t, files, "v1.0.0")
	gitRun(t, repo, "branch", "-M", "main")
	bare := filepath.Join(t.TempDir(), "skill.git")
	gitRun(t, repo, "clone", "--bare", repo, bare)
	cmd := exec.Command("git", "rev-parse", "refs/heads/main")
	cmd.Dir = bare
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	commit = strings.TrimSpace(string(out))
	operand := ids.serve(bare, "https://example.com/skills/binary-skill")
	home = t.TempDir()
	pinHomes(t)
	global := filepath.Join(home, "global")
	if err := os.MkdirAll(global, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDeclaration := func(ref, value string) {
		t.Helper()
		payload, err := json.Marshal(map[string]any{
			"schema_version": 1,
			"skills": []any{map[string]string{
				"name": "binary-skill", "git": operand, ref: value,
			}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(global, "Skillfile.json"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeDeclaration("branch", "main")
	legacy, err := manifest.Load(global)
	if err != nil || legacy == nil || len(legacy.Skills) != 1 || legacy.Skills[0].Ref.Kind != "branch" {
		t.Fatalf("legacy declaration = %+v, err = %v", legacy, err)
	}
	writeDeclaration("revision", commit)
	return home, commit
}

func migrationGzip(t *testing.T) string {
	t.Helper()
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write([]byte("attachment\n")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(compressed.Bytes(), []byte{0}) {
		t.Fatal("gzip fixture must carry a NUL byte")
	}
	return compressed.String()
}

// Production entry: EnsureDefaultWithPolicy -> ensureDefault ->
// migrateGlobalSkills -> DetectAtVersion and strictAuditMember. Reverting
// either audit's version independently must fail the v2 migration. The v1
// cases also rule out a directory exemption and assert no lock publication.
func TestMigratedGlobalSkillBinaryWriteVersion(t *testing.T) {
	gzipPayload := migrationGzip(t)
	for _, directory := range []string{".task-board/attachments", "assets/deep"} {
		for _, payloads := range []struct {
			name  string
			files map[string]string
		}{
			{"gzip-and-binary", map[string]string{"attachment.gz": gzipPayload, "small.bin": "a\x00b"}},
			{"gzip-only", map[string]string{"attachment.gz": gzipPayload}},
			{"binary-only", map[string]string{"small.bin": "a\x00b"}},
		} {
			for _, v2 := range []bool{true, false} {
				versionName := "v1"
				if v2 {
					versionName = "v2"
				}
				t.Run(directory+"/"+payloads.name+"/"+versionName, func(t *testing.T) {
					previous := hashing.EnableV2Writers
					hashing.EnableV2Writers = v2
					t.Cleanup(func() { hashing.EnableV2Writers = previous })
					files := map[string]string{"README.md": "binary skill\n"}
					for path, payload := range payloads.files {
						files[directory+"/"+path] = payload
					}
					home, commit := migrationRevisionFixture(t, files)
					err := EnsureDefaultWithPolicy(home, Policy{})
					if !v2 {
						if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) ||
							!strings.Contains(err.Error(), contextaudit.ClassOpaqueFile) || !strings.Contains(err.Error(), directory+"/") {
							t.Fatalf("v1 migration = %v, want profile_source_invalid with opaque-file in %s", err, directory)
						}
						assertMigrationUnpublished(t, home)
						return
					}
					if err != nil {
						t.Fatalf("v2 migration: %v", err)
					}
					lock, _, err := readLock(home, DefaultProfile)
					if err != nil {
						t.Fatal(err)
					}
					if lock.HashVersion != 2 || lock.SchemaVersion != contextlock.SchemaVersion2 {
						t.Fatalf("migrated lock version = %d/%d, want schema 2/hash 2", lock.SchemaVersion, lock.HashVersion)
					}
					if err := lock.Validate(); err != nil {
						t.Fatalf("migrated lock: %v", err)
					}
					if len(lock.Members) != 2 {
						t.Fatalf("migrated members = %+v, want root and skill", lock.Members)
					}
					member, ok := lock.Find(contextlock.KindSkill, "binary-skill")
					if !ok || member.Commit != commit || member.Source != "example.com/skills/binary-skill" {
						t.Fatalf("migrated skill = %+v, present = %v", member, ok)
					}
					entry := contextstore.EntryDir(home, contextlock.KindSkill, "binary-skill", commit)
					for path, want := range files {
						got, err := os.ReadFile(filepath.Join(entry, filepath.FromSlash(path)))
						if err != nil || string(got) != want {
							t.Fatalf("stored %s: err = %v, bytes preserved = %v", path, err, string(got) == want)
						}
					}
				})
			}
		}
	}
}

func assertMigrationUnpublished(t *testing.T, home string) {
	t.Helper()
	for _, path := range []string{sourcePath(home, DefaultProfile), lockPath(home, DefaultProfile)} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("refused migration published profile state: %v", err)
		}
	}
}

// Characterize the follow-up through the same production migration entry.
// Binary data passes the v2 audits, but the implicit-none capability contract
// still blocks the URL. This test records that bound, not a remedy for it.
func TestMigratedGlobalSkillNetworkUndeclaredFollowUp(t *testing.T) {
	previous := hashing.EnableV2Writers
	hashing.EnableV2Writers = true
	t.Cleanup(func() { hashing.EnableV2Writers = previous })
	home, _ := migrationRevisionFixture(t, map[string]string{
		"README.md":        "network skill\n",
		"assets/small.bin": "a\x00b",
		"scripts/fetch.sh": "curl https://downloads.example.com/archive\n",
	})
	err := EnsureDefaultWithPolicy(home, Policy{})
	if err == nil || !strings.Contains(err.Error(), DiagSourceInvalid) ||
		!strings.Contains(err.Error(), "audit.capability.network-undeclared") ||
		!strings.Contains(err.Error(), "scripts/fetch.sh") {
		t.Fatalf("migration = %v, want strict network-undeclared finding", err)
	}
	assertMigrationUnpublished(t, home)
	t.Logf("follow-up refusal: %v", err)
}
