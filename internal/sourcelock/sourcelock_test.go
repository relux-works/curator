package sourcelock

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Golden vectors below were cross-checked against an independent Python
// CCJ-1/SHA-256 implementation (/tmp/ccj_check.py): the Go digests must
// equal the Python digests byte for byte.

// goldenManifest is intentionally pretty-printed with unsorted keys: the
// digest runs over the parsed value, not the bytes.
const goldenManifest = `{
  "skills": [
    {"name": "review", "from": "team", "directory": "skills/review"},
    {"from": "project", "directory": "skills", "include": ["docs"]}
  ],
  "schema_version": 2,
  "sources": {
    "team": {"tag": "v1.2.0", "git": "https://example.org/kit.git"},
    "project": {"path": "./agents"}
  }
}`

const (
	goldenManifestSHA256 = "sha256:f38287608ffb1dc5ba62508a9afb6c9633544b38e013255f2a0ef724cc327066"
	goldenLockSHA256     = "sha256:f8541eccf52e7f2c479b1648bf386af4f554540272fde09fbe1741304889871b"
	goldenLocalDigest    = "sha256:ec6b66e6d2ad7cf7ae1cedebf98bb577292ae09821e956ce124fca5f6fba38ec"
	goldenNetworkDigest  = "sha256:c98205a2e34d49dbf04222c6fc399cc99745ff55461dc56e65aea937f02618bc"
	goldenConfigDigest   = "sha256:9d406e20e4faa6d678b04b5ff4dfa99791938a57dcc615e0823d93f19747d681"
)

func index(i int) *int { return &i }

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// goldenMembers returns the three-arm closure in deliberately unsorted
// order: New must sort it into canonical order.
func goldenMembers() []Member {
	local, err := LocalPackage("sha256:" + repeat("11", 32))
	if err != nil {
		panic(err)
	}
	network, err := NetworkGitPackage("example.org/kit", Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}, "skills/review")
	if err != nil {
		panic(err)
	}
	configured, err := ConfiguredGitPackage("team/audit-helper", Commit{ObjectFormat: "sha256", Hex: repeat("ab", 32)})
	if err != nil {
		panic(err)
	}
	return []Member{
		{Name: "review", Selection: index(0), Directory: "skills/review", Package: network, ContentSHA256: "sha256:" + repeat("33", 32)},
		{Name: "docs", Selection: index(1), Directory: "skills/docs", Package: local, ContentSHA256: "sha256:" + repeat("22", 32)},
		{Name: "audit-helper", Selection: nil, Directory: ".", Package: configured, ContentSHA256: "sha256:" + repeat("44", 32)},
	}
}

func mustGoldenLock(t *testing.T) *Lock {
	t.Helper()
	lock, err := New(goldenManifestSHA256, goldenMembers())
	if err != nil {
		t.Fatalf("New golden lock: %v", err)
	}
	return lock
}

func mustCanonical(t *testing.T, value any) []byte {
	t.Helper()
	canonical, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		t.Fatalf("marshal canonical: %v", err)
	}
	return canonical
}

func TestManifestDigestGolden(t *testing.T) {
	payload := []byte(goldenManifest)
	parsed, err := manifest.ParseBytesWithOptions(payload, "Skillfile.json", manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatalf("golden manifest must parse as draft schema 2: %v", err)
	}
	if parsed.SchemaVersion != 2 || len(parsed.Skills) != 2 {
		t.Fatalf("unexpected golden manifest shape: %+v", parsed)
	}
	digest, err := ManifestDigest(payload)
	if err != nil {
		t.Fatalf("ManifestDigest: %v", err)
	}
	if digest != goldenManifestSHA256 {
		t.Fatalf("manifest digest = %s, want %s", digest, goldenManifestSHA256)
	}
}

func TestNewSortsMembersAndSealsDigest(t *testing.T) {
	lock := mustGoldenLock(t)
	var names []string
	for _, member := range lock.Members {
		names = append(names, member.Name)
	}
	if strings.Join(names, ",") != "audit-helper,docs,review" {
		t.Fatalf("member order = %v, want canonical UTF-8 order", names)
	}
	if lock.LockSHA256 != goldenLockSHA256 {
		t.Fatalf("lock_sha256 = %s, want %s", lock.LockSHA256, goldenLockSHA256)
	}
	digest, err := lock.Digest()
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	if digest != goldenLockSHA256 {
		t.Fatalf("Digest = %s, want %s", digest, goldenLockSHA256)
	}
	if err := lock.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestPackageDigestsGolden(t *testing.T) {
	lock := mustGoldenLock(t)
	want := map[string]string{
		"docs":         goldenLocalDigest,
		"review":       goldenNetworkDigest,
		"audit-helper": goldenConfigDigest,
	}
	for _, member := range lock.Members {
		digest, err := member.Package.Digest()
		if err != nil {
			t.Fatalf("Digest %s: %v", member.Name, err)
		}
		if digest != want[member.Name] {
			t.Fatalf("package digest %s = %s, want %s", member.Name, digest, want[member.Name])
		}
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := PathIn(dir)
	if filepath.Base(path) != FileName || FileName != "Skillfile.lock.json" {
		t.Fatalf("lock file name = %s", path)
	}
	want := mustGoldenLock(t)
	if err := Write(path, want); err != nil {
		t.Fatalf("Write: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat lock: %v", err)
	}
	// Windows only honors the owner-write bit; exact Unix modes are
	// asserted off Windows, following the repository convention.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o644 {
		t.Fatalf("lock mode = %o, want 644", info.Mode().Perm())
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	if err := protocoljson.RequireCanonical(payload); err != nil {
		t.Fatalf("lock bytes are not canonical: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(mustCanonical(t, got.Object())) != string(mustCanonical(t, want.Object())) {
		t.Fatalf("round trip changed the lock value")
	}
	if got.Members[1].Selection == nil || *got.Members[1].Selection != 1 {
		t.Fatalf("selection index did not survive the round trip")
	}
	if got.Members[0].Selection != nil {
		t.Fatalf("null transitive selection became %v", *got.Members[0].Selection)
	}
}

func TestReadMissing(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), FileName))
	if !os.IsNotExist(err) {
		t.Fatalf("missing lock err = %v, want IsNotExist", err)
	}
}

func TestParseAcceptsPrettyPrinted(t *testing.T) {
	want := mustGoldenLock(t)
	pretty, err := json.MarshalIndent(want.Object(), "", "  ")
	if err != nil {
		t.Fatalf("indent: %v", err)
	}
	got, err := Parse(pretty)
	if err != nil {
		t.Fatalf("Parse pretty-printed lock: %v", err)
	}
	if string(mustCanonical(t, got.Object())) != string(mustCanonical(t, want.Object())) {
		t.Fatalf("pretty-printed parse changed the lock value")
	}
}

func TestCheckStale(t *testing.T) {
	lock := mustGoldenLock(t)
	if err := lock.CheckStale([]byte(goldenManifest)); err != nil {
		t.Fatalf("fresh lock reported stale: %v", err)
	}
	if err := lock.CheckStaleDigest(goldenManifestSHA256); err != nil {
		t.Fatalf("fresh digest reported stale: %v", err)
	}
	changed := strings.Replace(goldenManifest, `"v1.2.0"`, `"v1.2.1"`, 1)
	if err := lock.CheckStale([]byte(changed)); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatalf("changed manifest err = %v, want source_lock_stale", err)
	}
	if err := lock.CheckStaleDigest("sha256:" + repeat("00", 32)); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatalf("wrong digest err = %v, want source_lock_stale", err)
	}
	// Whitespace alone is not a manifest change: the digest runs over the value.
	relaxed := strings.ReplaceAll(goldenManifest, " ", "")
	if err := lock.CheckStale([]byte(relaxed)); err != nil {
		t.Fatalf("whitespace-only change reported stale: %v", err)
	}
}

func lockPayload(t *testing.T, mutate func(obj map[string]any)) []byte {
	t.Helper()
	var obj map[string]any
	raw, err := json.Marshal(mustGoldenLock(t).Object())
	if err != nil {
		t.Fatalf("marshal base lock: %v", err)
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal base lock: %v", err)
	}
	if mutate != nil {
		mutate(obj)
	}
	payload, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal mutated lock: %v", err)
	}
	return payload
}

func memberAt(obj map[string]any, i int) map[string]any {
	return obj["members"].([]any)[i].(map[string]any)
}

func TestMalformedLocks(t *testing.T) {
	cases := []struct {
		name      string
		payload   func(t *testing.T) []byte
		wantClass string
	}{
		{"garbage", func(*testing.T) []byte { return []byte("{not json") }, "source_selection_invalid"},
		{"trailing-data", func(*testing.T) []byte { return []byte("{} {}") }, "source_selection_invalid"},
		{"duplicate-keys", func(*testing.T) []byte {
			return []byte(`{"schema_version":1,"schema_version":1,"manifest_sha256":"x","members":[],"lock_sha256":"y"}`)
		}, "source_selection_invalid"},
		{"top-level-array", func(*testing.T) []byte { return []byte("[]") }, "source_selection_invalid"},
		{"unknown-top-level", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { obj["unexpected"] = true })
		}, "source_selection_invalid"},
		{"schema-version-2", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { obj["schema_version"] = float64(2) })
		}, "source_selection_invalid"},
		{"schema-version-string", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { obj["schema_version"] = "1" })
		}, "source_selection_invalid"},
		{"manifest-digest-bare", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { obj["manifest_sha256"] = repeat("ab", 32) })
		}, "source_selection_invalid"},
		{"members-missing", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { delete(obj, "members") })
		}, "source_selection_invalid"},
		{"member-unknown-field", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 0)["location"] = "/tmp/x" })
		}, "source_member_invalid"},
		{"member-machine-endpoint", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 2)["endpoint"] = "https://example.org/kit.git" })
		}, "source_member_invalid"},
		{"member-missing-selection", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { delete(memberAt(obj, 0), "selection") })
		}, "source_selection_invalid"},
		{"selection-negative", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 1)["selection"] = float64(-1) })
		}, "source_selection_invalid"},
		{"selection-fraction", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 1)["selection"] = 1.5 })
		}, "source_selection_invalid"},
		{"selection-string", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 1)["selection"] = "1" })
		}, "source_selection_invalid"},
		{"name-empty", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 0)["name"] = "" })
		}, "source_member_invalid"},
		{"name-reserved", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 0)["name"] = "CON" })
		}, "source_member_invalid"},
		{"directory-absolute", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 1)["directory"] = "/tmp/x" })
		}, "source_member_invalid"},
		{"directory-glob", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 1)["directory"] = "skills/*" })
		}, "source_member_invalid"},
		{"content-bare", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { memberAt(obj, 0)["content_sha256"] = repeat("44", 32) })
		}, "source_member_invalid"},
		{"lock-digest-missing", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) { delete(obj, "lock_sha256") })
		}, "source_selection_invalid"},
		{"lock-digest-tampered-member", func(t *testing.T) []byte {
			return lockPayload(t, func(obj map[string]any) {
				memberAt(obj, 0)["content_sha256"] = "sha256:" + repeat("55", 32)
			})
		}, "source_selection_invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.payload(t))
			if err == nil {
				t.Fatalf("Parse accepted %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantClass) {
				t.Fatalf("err = %v, want class %s", err, tc.wantClass)
			}
		})
	}
}

func TestDuplicateNamesConflict(t *testing.T) {
	members := goldenMembers()
	members = append(members, members[0])
	if _, err := New(goldenManifestSHA256, members); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
		t.Fatalf("exact duplicate err = %v, want source_name_conflict", err)
	}
	folded := goldenMembers()
	alias := folded[0]
	alias.Name = "REVIEW"
	alias.ContentSHA256 = "sha256:" + repeat("34", 32)
	folded = append(folded, alias)
	if _, err := New(goldenManifestSHA256, folded); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
		t.Fatalf("case-fold duplicate err = %v, want source_name_conflict", err)
	}
	payload := lockPayload(t, func(obj map[string]any) {
		dup := memberAt(obj, 0)
		obj["members"] = []any{dup, dup}
	})
	if _, err := Parse(payload); err == nil || !strings.Contains(err.Error(), "source_name_conflict") {
		t.Fatalf("parsed duplicate err = %v, want source_name_conflict", err)
	}
}

func TestUnsortedMembersRejected(t *testing.T) {
	lock := mustGoldenLock(t)
	raw, err := json.Marshal(lock.Object())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	members := obj["members"].([]any)
	obj["members"] = []any{members[2], members[1], members[0]}
	payload, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	_, err = Parse(payload)
	if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
		t.Fatalf("unsorted members err = %v, want source_selection_invalid", err)
	}
}

func TestThreeArmsRoundTrip(t *testing.T) {
	sha1 := Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}
	sha256 := Commit{ObjectFormat: "sha256", Hex: repeat("ab", 32)}
	local, err := LocalPackage("sha256:" + repeat("11", 32))
	if err != nil {
		t.Fatalf("LocalPackage: %v", err)
	}
	network, err := NetworkGitPackage("example.org/kit", sha1, "skills/review")
	if err != nil {
		t.Fatalf("NetworkGitPackage: %v", err)
	}
	configured, err := ConfiguredGitPackage("team/review", sha256)
	if err != nil {
		t.Fatalf("ConfiguredGitPackage: %v", err)
	}
	if local.IsGit() || !network.IsGit() || !configured.IsGit() {
		t.Fatalf("arm predicates confused: %+v %+v %+v", local, network, configured)
	}
	lock, err := New(goldenManifestSHA256, []Member{
		{Name: "a-local", Selection: index(0), Directory: ".", Package: local, ContentSHA256: "sha256:" + repeat("01", 32)},
		{Name: "b-network", Selection: index(1), Directory: "skills/review", Package: network, ContentSHA256: "sha256:" + repeat("02", 32)},
		{Name: "c-configured", Selection: nil, Directory: ".", Package: configured, ContentSHA256: "sha256:" + repeat("03", 32)},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	parsed, err := Parse(mustCanonical(t, lock.Object()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	for _, name := range []string{"a-local", "b-network", "c-configured"} {
		want, _ := lock.Find(name)
		got, ok := parsed.Find(name)
		if !ok || !got.equalRecord(want) {
			t.Fatalf("member %s did not survive the round trip", name)
		}
	}
}

func packageAt(obj map[string]any, i int) map[string]any {
	return memberAt(obj, i)["package"].(map[string]any)
}

func TestCrossArmConfusionRejected(t *testing.T) {
	// Member indexes in canonical order: 0 audit-helper (configured),
	// 1 docs (local), 2 review (network).
	cases := []struct {
		name   string
		mutate func(obj map[string]any)
	}{
		{"local-with-commit", func(obj map[string]any) {
			packageAt(obj, 1)["commit"] = map[string]any{"object_format": "sha1", "hex": repeat("0", 40)}
		}},
		{"local-with-repository", func(obj map[string]any) {
			packageAt(obj, 1)["repository"] = "example.org/kit"
		}},
		{"local-with-source", func(obj map[string]any) {
			packageAt(obj, 1)["source"] = "team/docs"
		}},
		{"local-with-directory", func(obj map[string]any) {
			packageAt(obj, 1)["directory"] = "."
		}},
		{"local-missing-snapshot", func(obj map[string]any) {
			delete(packageAt(obj, 1), "snapshot")
		}},
		{"local-bare-snapshot", func(obj map[string]any) {
			packageAt(obj, 1)["snapshot"] = repeat("11", 32)
		}},
		{"network-with-snapshot", func(obj map[string]any) {
			packageAt(obj, 2)["snapshot"] = "sha256:" + repeat("11", 32)
		}},
		{"network-with-source", func(obj map[string]any) {
			packageAt(obj, 2)["source"] = "team/review"
		}},
		{"network-missing-commit", func(obj map[string]any) {
			delete(packageAt(obj, 2), "commit")
		}},
		{"network-missing-directory", func(obj map[string]any) {
			delete(packageAt(obj, 2), "directory")
		}},
		{"network-endpoint-smuggled", func(obj map[string]any) {
			packageAt(obj, 2)["url"] = "https://example.org/kit.git"
		}},
		{"configured-with-repository", func(obj map[string]any) {
			packageAt(obj, 0)["repository"] = "example.org/kit"
		}},
		{"configured-with-snapshot", func(obj map[string]any) {
			packageAt(obj, 0)["snapshot"] = "sha256:" + repeat("11", 32)
		}},
		{"configured-nonroot-directory", func(obj map[string]any) {
			packageAt(obj, 0)["directory"] = "sub"
		}},
		{"unknown-kind", func(obj map[string]any) {
			packageAt(obj, 1)["kind"] = "git"
		}},
		{"commit-extra-field", func(obj map[string]any) {
			packageAt(obj, 2)["commit"].(map[string]any)["snapshot"] = "sha256:" + repeat("11", 32)
		}},
		{"commit-missing-hex", func(obj map[string]any) {
			delete(packageAt(obj, 2)["commit"].(map[string]any), "hex")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(lockPayload(t, tc.mutate))
			if err == nil {
				t.Fatalf("Parse accepted %s", tc.name)
			}
			if !strings.Contains(err.Error(), "source_member_invalid") && !strings.Contains(err.Error(), "source_selection_invalid") {
				t.Fatalf("err = %v, want a source selection/member diagnostic", err)
			}
		})
	}
	// A hand-built struct must not smuggle Git fields into a local arm either.
	smuggled := Package{Kind: KindLocalSnapshot, Snapshot: "sha256:" + repeat("11", 32), Commit: Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}}
	if err := smuggled.Validate("package"); err == nil {
		t.Fatalf("Validate accepted a local package carrying a commit")
	}
	if _, err := LocalPackage(repeat("11", 32)); err == nil {
		t.Fatalf("LocalPackage accepted a bare-hex snapshot")
	}
}

func TestCommitValidation(t *testing.T) {
	cases := []struct {
		name   string
		commit Commit
		valid  bool
	}{
		{"sha1", Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}, true},
		{"sha256", Commit{ObjectFormat: "sha256", Hex: repeat("ab", 32)}, true},
		{"sha1-long", Commit{ObjectFormat: "sha1", Hex: repeat("0", 64)}, false},
		{"sha256-short", Commit{ObjectFormat: "sha256", Hex: repeat("0", 40)}, false},
		{"uppercase", Commit{ObjectFormat: "sha1", Hex: repeat("A", 40)}, false},
		{"prefixed", Commit{ObjectFormat: "sha1", Hex: "sha256:" + repeat("0", 40)}, false},
		{"snapshot-in-commit", Commit{ObjectFormat: "sha256", Hex: ""}, false},
		{"bad-format", Commit{ObjectFormat: "md5", Hex: repeat("0", 32)}, false},
		{"empty", Commit{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NetworkGitPackage("example.org/kit", tc.commit, "skills/review")
			if tc.valid && err != nil {
				t.Fatalf("valid commit rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("invalid commit accepted")
			}
		})
	}
}

func TestDirectoryAgreement(t *testing.T) {
	network, err := NetworkGitPackage("example.org/kit", Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}, "skills/review")
	if err != nil {
		t.Fatalf("NetworkGitPackage: %v", err)
	}
	disagree := Member{Name: "review", Selection: index(0), Directory: "skills/other", Package: network, ContentSHA256: "sha256:" + repeat("33", 32)}
	if err := disagree.Validate("members[0]"); err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("directory disagreement err = %v, want source_member_invalid", err)
	}
	agree := disagree
	agree.Directory = "skills/review"
	if err := agree.Validate("members[0]"); err != nil {
		t.Fatalf("agreeing directories rejected: %v", err)
	}
	configured, err := ConfiguredGitPackage("team/review", Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)})
	if err != nil {
		t.Fatalf("ConfiguredGitPackage: %v", err)
	}
	legacy := Member{Name: "review", Selection: index(0), Directory: "sub", Package: configured, ContentSHA256: "sha256:" + repeat("33", 32)}
	if err := legacy.Validate("members[0]"); err == nil {
		t.Fatalf("configured-git with non-root member directory accepted")
	}
	legacy.Directory = "."
	if err := legacy.Validate("members[0]"); err != nil {
		t.Fatalf("legacy root directory rejected: %v", err)
	}
}

func TestRepositoryValidation(t *testing.T) {
	commit := Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}
	valid := []string{"example.org/kit", "github.com/acme/kit", "git.example-codec-1.internal/a/b/c"}
	for _, repo := range valid {
		if _, err := NetworkGitPackage(repo, commit, "."); err != nil {
			t.Fatalf("repository %q rejected: %v", repo, err)
		}
	}
	invalid := []string{
		"",
		"Example.org/kit",
		"example.org:8080/kit",
		"example.org/kit?.git",
		"example.org/kit with space",
		"example.org/kit%20x",
		"example.org",
		"example.org/../kit",
		"https://example.org/kit",
		"example.org/kit.git",
		"github.com/acme/kit.git",
		"example.org/" + repeat("a", 4096),
	}
	for _, repo := range invalid {
		if _, err := NetworkGitPackage(repo, commit, "."); err == nil {
			t.Fatalf("repository %q accepted", repo)
		}
	}
}

// TestNoncanonicalRepositoryRejected pins repository-transport revision 1
// §1 for the lock: a member repository MUST already be canonical, so a
// terminal ".git" is rejected by New, Write, Parse and Read alike. The
// persisted payload carries a correctly recomputed lock_sha256, proving
// the rejection is normalization (source_member_invalid on the
// repository), not an integrity mismatch.
func TestNoncanonicalRepositoryRejected(t *testing.T) {
	commit := Commit{ObjectFormat: "sha1", Hex: repeat("0", 40)}
	if _, err := NetworkGitPackage("example.org/kit.git", commit, "skills/review"); err == nil ||
		!strings.Contains(err.Error(), "source_member_invalid") || !strings.Contains(err.Error(), "repository") {
		t.Fatalf("NetworkGitPackage .git err = %v, want source_member_invalid on repository", err)
	}
	members := goldenMembers()
	for i := range members {
		if members[i].Package.Kind == KindNetworkGit {
			members[i].Package.Repository = "example.org/kit.git"
		}
	}
	if _, err := New(goldenManifestSHA256, members); err == nil ||
		!strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("New .git err = %v, want source_member_invalid", err)
	}
	// Canonical member order isolates the repository as the only defect.
	sort.SliceStable(members, func(i, j int) bool {
		return bytes.Compare([]byte(members[i].Name), []byte(members[j].Name)) < 0
	})
	raw := &Lock{ManifestSHA256: goldenManifestSHA256, Members: members}
	digest, err := raw.digest()
	if err != nil {
		t.Fatalf("digest of .git lock: %v", err)
	}
	raw.LockSHA256 = digest
	assertNormalization := func(stage string, err error) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), "source_member_invalid") ||
			!strings.Contains(err.Error(), "repository") {
			t.Fatalf("%s err = %v, want source_member_invalid on repository", stage, err)
		}
		if strings.Contains(err.Error(), "lock_sha256 mismatch") {
			t.Fatalf("%s failed integrity instead of normalization: %v", stage, err)
		}
	}
	assertNormalization("Write", Write(PathIn(t.TempDir()), raw))
	payload := mustCanonical(t, raw.Object())
	assertNormalization("Parse", func() error { _, err := Parse(payload); return err }())
	diskPath := PathIn(t.TempDir())
	if err := os.WriteFile(diskPath, payload, 0o644); err != nil {
		t.Fatalf("stage .git lock on disk: %v", err)
	}
	assertNormalization("Read", func() error { _, err := Read(diskPath); return err }())
	// Positive control: the canonical spelling without ".git" stays valid.
	if _, err := NetworkGitPackage("example.org/kit", commit, "skills/review"); err != nil {
		t.Fatalf("canonical repository rejected: %v", err)
	}
}

func TestSelectionSemantics(t *testing.T) {
	local, err := LocalPackage("sha256:" + repeat("11", 32))
	if err != nil {
		t.Fatalf("LocalPackage: %v", err)
	}
	// Collection siblings share one selection index; a transitive member
	// carries null. Index 0 must never read as null.
	lock, err := New(goldenManifestSHA256, []Member{
		{Name: "sib-a", Selection: index(1), Directory: "skills/sib-a", Package: local, ContentSHA256: "sha256:" + repeat("22", 32)},
		{Name: "sib-b", Selection: index(1), Directory: "skills/sib-b", Package: local, ContentSHA256: "sha256:" + repeat("23", 32)},
		{Name: "root-zero", Selection: index(0), Directory: ".", Package: local, ContentSHA256: "sha256:" + repeat("24", 32)},
		{Name: "transitive", Selection: nil, Directory: ".", Package: local, ContentSHA256: "sha256:" + repeat("25", 32)},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	parsed, err := Parse(mustCanonical(t, lock.Object()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	zero, _ := parsed.Find("root-zero")
	if zero.Selection == nil || *zero.Selection != 0 {
		t.Fatalf("selection 0 confused with null")
	}
	transitive, _ := parsed.Find("transitive")
	if transitive.Selection != nil {
		t.Fatalf("null selection became %d", *transitive.Selection)
	}
	asRoot := transitive
	asRoot.Selection = index(2)
	if err := parsed.CheckMembership([]Member{asRoot}); err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
		t.Fatalf("null-vs-index substitution err = %v, want source_member_invalid", err)
	}
}

func TestEmptyMembers(t *testing.T) {
	lock, err := New(goldenManifestSHA256, nil)
	if err != nil {
		t.Fatalf("empty lock rejected: %v", err)
	}
	parsed, err := Parse(mustCanonical(t, lock.Object()))
	if err != nil {
		t.Fatalf("Parse empty lock: %v", err)
	}
	if len(parsed.Members) != 0 {
		t.Fatalf("empty lock gained members")
	}
	if err := parsed.CheckNames(nil); err != nil {
		t.Fatalf("empty coverage rejected: %v", err)
	}
	if err := parsed.CheckNames([]string{"x"}); err == nil || !strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("empty lock coverage err = %v, want source_member_missing", err)
	}
}

func TestCheckNames(t *testing.T) {
	lock := mustGoldenLock(t)
	if err := lock.CheckNames([]string{"review", "docs", "audit-helper"}); err != nil {
		t.Fatalf("exact coverage rejected: %v", err)
	}
	// A plan name outside the lock is missing; a locked name outside the
	// plan is stale. Callers pass manifest expansion, never live
	// filesystem members, so live extras never widen the frozen set.
	if err := lock.CheckNames([]string{"docs", "review", "audit-helper", "docs-extra"}); err == nil || !strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("extra plan name err = %v, want source_member_missing", err)
	}
	if err := lock.CheckNames([]string{"review"}); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatalf("extra locked member err = %v, want source_lock_stale", err)
	}
}

func TestCheckMembership(t *testing.T) {
	lock := mustGoldenLock(t)
	if err := lock.CheckMembership(goldenMembers()); err != nil {
		t.Fatalf("exact membership rejected: %v", err)
	}
	missing := goldenMembers()[:2]
	if err := lock.CheckMembership(missing); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatalf("narrow plan err = %v, want source_lock_stale", err)
	}
	extra := append(append([]Member(nil), goldenMembers()...), Member{Name: "stowaway"})
	if err := lock.CheckMembership(extra); err == nil || !strings.Contains(err.Error(), "source_member_missing") {
		t.Fatalf("stowaway err = %v, want source_member_missing", err)
	}
	differ := func(mutate func(*Member)) []Member {
		members := append([]Member(nil), goldenMembers()...)
		for i := range members {
			if members[i].Name == "review" {
				mutate(&members[i])
			}
		}
		return members
	}
	cases := map[string][]Member{
		"directory": differ(func(m *Member) { m.Directory = "skills/other" }),
		"commit": differ(func(m *Member) {
			m.Package.Commit = Commit{ObjectFormat: "sha1", Hex: repeat("1", 40)}
		}),
		"content":   differ(func(m *Member) { m.ContentSHA256 = "sha256:" + repeat("99", 32) }),
		"selection": differ(func(m *Member) { m.Selection = index(7) }),
		"kind": differ(func(m *Member) {
			m.Package.Kind = KindLocalSnapshot
			m.Package.Snapshot = "sha256:" + repeat("11", 32)
		}),
	}
	for name, expected := range cases {
		if err := lock.CheckMembership(expected); err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
			t.Fatalf("%s drift err = %v, want source_member_invalid", name, err)
		}
	}
}

func TestFind(t *testing.T) {
	lock := mustGoldenLock(t)
	member, ok := lock.Find("docs")
	if !ok || member.Name != "docs" {
		t.Fatalf("Find missed a locked member")
	}
	if _, ok := lock.Find("nope"); ok {
		t.Fatalf("Find returned an unlocked member")
	}
	var nilLock *Lock
	if _, ok := nilLock.Find("docs"); ok {
		t.Fatalf("Find on nil lock succeeded")
	}
}

func TestNilGuards(t *testing.T) {
	var lock *Lock
	if err := lock.Validate(); err == nil {
		t.Fatalf("nil Validate accepted")
	}
	if _, err := lock.Digest(); err == nil {
		t.Fatalf("nil Digest accepted")
	}
	if err := lock.CheckStale([]byte(goldenManifest)); err == nil {
		t.Fatalf("nil CheckStale accepted")
	}
	if err := lock.CheckNames(nil); err == nil {
		t.Fatalf("nil CheckNames accepted")
	}
	if err := lock.CheckMembership(nil); err == nil {
		t.Fatalf("nil CheckMembership accepted")
	}
	if err := Write(filepath.Join(t.TempDir(), FileName), nil); err == nil {
		t.Fatalf("Write nil lock accepted")
	}
	if _, err := ManifestDigest(nil); err == nil {
		t.Fatalf("ManifestDigest accepted empty input")
	}
	if _, err := ManifestDigest([]byte(`[]`)); err == nil {
		t.Fatalf("ManifestDigest accepted a non-object")
	}
}

func objectKeys(value map[string]any) []string {
	var keys []string
	for key := range value {
		keys = append(keys, key)
	}
	return keys
}

func TestMachineSeparation(t *testing.T) {
	lock := mustGoldenLock(t)
	top := map[string]bool{}
	for _, key := range objectKeys(lock.Object()) {
		top[key] = true
	}
	for _, key := range []string{"schema_version", "manifest_sha256", "members", "lock_sha256"} {
		if !top[key] || len(top) != 4 {
			t.Fatalf("lock object keys = %v", objectKeys(lock.Object()))
		}
	}
	for _, member := range lock.Members {
		obj := member.object()
		fields := map[string]bool{}
		for _, key := range objectKeys(obj) {
			fields[key] = true
		}
		for _, key := range []string{"name", "selection", "directory", "package", "content_sha256"} {
			if !fields[key] || len(fields) != 5 {
				t.Fatalf("member %s object keys = %v", member.Name, objectKeys(obj))
			}
		}
		packageObj := member.Package.object()
		pfields := map[string]bool{}
		for _, key := range objectKeys(packageObj) {
			pfields[key] = true
		}
		var want []string
		switch member.Package.Kind {
		case KindLocalSnapshot:
			want = []string{"kind", "snapshot"}
		case KindNetworkGit:
			want = []string{"kind", "repository", "commit", "directory"}
		case KindConfiguredGit:
			want = []string{"kind", "source", "commit", "directory"}
		}
		for _, key := range want {
			if !pfields[key] || len(pfields) != len(want) {
				t.Fatalf("package %s object keys = %v", member.Name, objectKeys(packageObj))
			}
		}
	}
	// No absolute path, endpoint, or credential-adjacent field may appear
	// in the portable bytes.
	raw := string(mustCanonical(t, lock.Object()))
	for _, forbidden := range []string{"/tmp/", "https://", "ssh://", "location", "endpoint", "authentication", ".git"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("lock bytes contain %q", forbidden)
		}
	}
}

func TestRebindingPreservesIdentity(t *testing.T) {
	lock := mustGoldenLock(t)
	member, _ := lock.Find("docs")
	first, err := member.Package.Digest()
	if err != nil {
		t.Fatalf("Digest: %v", err)
	}
	again, err := member.Package.Digest()
	if err != nil || again != first {
		t.Fatalf("package digest unstable: %s vs %s", first, again)
	}
	// The same lock generation binds fresh machine locations on two
	// machines without changing portable identity. Locations come from
	// t.TempDir so they are absolute on every platform.
	for _, machine := range []string{"machine-a", "machine-b"} {
		location := filepath.Join(t.TempDir(), machine, "skills")
		bindings, err := NewBindings(lock.LockSHA256, map[string]SourceBinding{
			"project": {Location: location},
		})
		if err != nil {
			t.Fatalf("NewBindings: %v", err)
		}
		if err := bindings.CheckFresh(lock); err != nil {
			t.Fatalf("rebound bindings rejected: %v", err)
		}
	}
}

func TestContextLockConfusion(t *testing.T) {
	// A context lock (environments capability) must never parse as a
	// package lock, and a package lock must never parse as a context lock.
	context := &contextlock.Lock{
		Root: "app",
		Members: []contextlock.Member{
			{Kind: contextlock.KindContext, Name: "app", Version: "1.0.0", StateHash: repeat("aa", 32), RequiredBy: []string{}},
		},
	}
	contextBytes, err := context.Canonical()
	if err != nil {
		t.Fatalf("context Canonical: %v", err)
	}
	if _, err := Parse(contextBytes); err == nil {
		t.Fatalf("package lock parser accepted a context lock")
	}
	packageBytes := mustCanonical(t, mustGoldenLock(t).Object())
	if _, err := contextlock.Parse(packageBytes); err == nil {
		t.Fatalf("context lock parser accepted a package lock")
	}
}
