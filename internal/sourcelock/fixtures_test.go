package sourcelock

import (
	"encoding/json"
	"strings"
	"testing"
)

// Byte-copies of curator-spec
// conformance/draft-sources-v1/schema-cases/skillfile-lock-v1/*.json at
// main 871d11b. They pin JSON-schema shapes, not digests: every file
// carries the same placeholder lock_sha256, and the Git fixtures use a
// placeholder package directory that disagrees with the member directory.
// The normative member/package directory-agreement rule therefore rejects
// the Git fixtures semantically; the invalid fixtures must fail
// structurally, before integrity is even consulted.

const fixtureValidLocal = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "local-snapshot",
        "snapshot": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81"
}`

const fixtureValidGit = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "network-git",
        "repository": "example.org/kit",
        "directory": ".",
        "commit": {
          "object_format": "sha1",
          "hex": "0000000000000000000000000000000000000000"
        }
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81"
}`

const fixtureValidConfigured = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "configured-git",
        "source": "team/review",
        "commit": {
          "object_format": "sha1",
          "hex": "0000000000000000000000000000000000000000"
        },
        "directory": "."
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81"
}`

const fixtureInvalidCommitFormat = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "network-git",
        "repository": "example.org/kit",
        "directory": ".",
        "commit": {
          "object_format": "sha1",
          "hex": "0000000000000000000000000000000000000000000000000000000000000000"
        }
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81"
}`

const fixtureInvalidConfiguredEscape = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "configured-git",
        "source": "../review",
        "commit": {
          "object_format": "sha1",
          "hex": "0000000000000000000000000000000000000000"
        },
        "directory": "."
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81"
}`

const fixtureInvalidFakeCommit = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "local-snapshot",
        "snapshot": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
        "commit": "0000000000000000000000000000000000000000"
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81"
}`

const fixtureInvalidUnknownTopLevel = `{
  "schema_version": 1,
  "manifest_sha256": "sha256:b9b161ebadcac2d76b211fcfb38b74da89f62fc5b605e614588732555784f989",
  "members": [
    {
      "name": "review",
      "selection": 0,
      "directory": "agents/skills/review",
      "package": {
        "kind": "local-snapshot",
        "snapshot": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
      },
      "content_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111"
    }
  ],
  "lock_sha256": "sha256:04c57a18c2e1ccae26695420aa586ded8d15d7aace16d198bc5fbe4fbb060d81",
  "unexpected": true
}`

// withDigest rewrites a fixture's placeholder lock_sha256.
func withDigest(t *testing.T, fixture, digest string) []byte {
	t.Helper()
	var obj map[string]any
	if err := json.Unmarshal([]byte(fixture), &obj); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	obj["lock_sha256"] = digest
	payload, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return payload
}

func TestSchemaFixtures(t *testing.T) {
	// The local fixture is structurally valid: with a wrong digest the
	// only failure is integrity, proving structure passed.
	_, err := Parse(withDigest(t, fixtureValidLocal, "sha256:"+repeat("00", 32)))
	if err == nil || !strings.Contains(err.Error(), "lock_sha256 mismatch") {
		t.Fatalf("valid local fixture err = %v, want only a digest mismatch", err)
	}
	// The Git fixtures disagree on member/package directories, which the
	// normative text forbids even though JSON Schema cannot express it.
	for name, fixture := range map[string]string{"git": fixtureValidGit, "configured": fixtureValidConfigured} {
		_, err := Parse(withDigest(t, fixture, "sha256:"+repeat("00", 32)))
		if err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
			t.Fatalf("%s fixture err = %v, want source_member_invalid directory disagreement", name, err)
		}
		if strings.Contains(err.Error(), "lock_sha256 mismatch") {
			t.Fatalf("%s fixture failed integrity before structure: %v", name, err)
		}
	}
	// Every invalid fixture must fail structurally: integrity is never
	// consulted for a malformed lock.
	invalid := map[string]string{
		"commit-format":     fixtureInvalidCommitFormat,
		"configured-escape": fixtureInvalidConfiguredEscape,
		"fake-commit":       fixtureInvalidFakeCommit,
		"unknown-top-level": fixtureInvalidUnknownTopLevel,
	}
	for name, fixture := range invalid {
		_, err := Parse(withDigest(t, fixture, "sha256:"+repeat("00", 32)))
		if err == nil {
			t.Fatalf("invalid fixture %s accepted", name)
		}
		if strings.Contains(err.Error(), "lock_sha256 mismatch") {
			t.Fatalf("invalid fixture %s failed integrity before structure: %v", name, err)
		}
		if !strings.Contains(err.Error(), "source_member_invalid") && !strings.Contains(err.Error(), "source_selection_invalid") {
			t.Fatalf("invalid fixture %s err = %v, want a source diagnostic", name, err)
		}
	}
}
