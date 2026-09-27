package envprofile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
)

type sourceSignerDeltaVectorFile struct {
	DeltaCases []sourceSignerDeltaVector `json:"delta_cases"`
}

type sourceSignerDeltaVector struct {
	Name       string                          `json:"name"`
	Conforming *bool                           `json:"conforming"`
	Flag       bool                            `json:"flag"`
	OldMembers []sourceSignerVectorMember      `json:"old_members"`
	NewMembers []sourceSignerVectorMember      `json:"new_members"`
	Snapshots  map[string]sourceSignerSnapshot `json:"snapshots"`
	Expected   struct {
		Lines     []string `json:"lines"`
		Trigger   []string `json:"trigger"`
		RevisionB struct {
			Diagnostic string `json:"diagnostic"`
			Proceeds   bool   `json:"proceeds"`
		} `json:"revision_b"`
	} `json:"expected"`
	Claimed struct {
		Proceeds bool `json:"proceeds"`
	} `json:"claimed"`
}

type sourceSignerVectorMember struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

type sourceSignerSnapshot struct {
	SystemModules []struct {
		Path         string   `json:"path"`
		Environments []string `json:"environments"`
		Bytes        string   `json:"bytes"`
	} `json:"system_modules"`
	MCP json.RawMessage `json:"mcp"`
}

// TestSystemDeltaVectors binds every pinned rc.13 delta case to the
// context-lock diff and snapshot comparison used by profile update. The CLI
// confirmation and publication path is separately driven by its golden.
func TestSystemDeltaVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-source-signers.json")) // #nosec G304 -- explicit conformance root
	if err != nil {
		t.Fatal(err)
	}
	var vectors sourceSignerDeltaVectorFile
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.DeltaCases) == 0 {
		t.Fatal("pinned source signer vector publishes no delta cases")
	}
	for _, tc := range vectors.DeltaCases {
		t.Run(tc.Name, func(t *testing.T) {
			home := t.TempDir()
			oldLock := &contextlock.Lock{Root: "root", Members: vectorLockMembers(tc.OldMembers)}
			newLock := &contextlock.Lock{Root: "root", Members: vectorLockMembers(tc.NewMembers)}
			for _, member := range append(append([]contextlock.Member{}, oldLock.Members...), newLock.Members...) {
				if member.Commit == "" {
					continue
				}
				snapshot, ok := tc.Snapshots[member.Commit]
				if !ok {
					t.Fatalf("vector snapshot for %s %s commit %s is absent", member.Kind, member.Name, member.Commit)
				}
				writeVectorSnapshot(t, home, member, snapshot)
			}
			lines, triggers, err := resolvedDelta(home, newGitManager(home), oldLock, newLock)
			if err != nil {
				t.Fatal(err)
			}
			proceeds := len(triggers) == 0 || tc.Flag
			if tc.Conforming != nil && !*tc.Conforming {
				if len(lines) == 0 || len(triggers) == 0 {
					t.Fatalf("non-conforming claim fixture produced no actual system delta: lines=%v triggers=%v", lines, triggers)
				}
				if proceeds || !tc.Claimed.Proceeds {
					t.Fatalf("revision-B gate proceeds=%t, claimed adversarially=%t; want refusal against the claim", proceeds, tc.Claimed.Proceeds)
				}
				return
			}
			if strings.Join(lines, "\n") != strings.Join(tc.Expected.Lines, "\n") {
				t.Errorf("delta lines = %#v, want %#v", lines, tc.Expected.Lines)
			}
			if strings.Join(triggers, "\n") != strings.Join(tc.Expected.Trigger, "\n") {
				t.Errorf("system-delta triggers = %#v, want %#v", triggers, tc.Expected.Trigger)
			}
			if proceeds != tc.Expected.RevisionB.Proceeds {
				t.Errorf("revision-B proceeds = %t, want %t", proceeds, tc.Expected.RevisionB.Proceeds)
			}
			if !proceeds && tc.Expected.RevisionB.Diagnostic != DiagSystemDeltaConfirmationRequired {
				t.Errorf("revision-B diagnostic = %q, want %q", tc.Expected.RevisionB.Diagnostic, DiagSystemDeltaConfirmationRequired)
			}
		})
	}
}

func vectorLockMembers(vector []sourceSignerVectorMember) []contextlock.Member {
	members := make([]contextlock.Member, 0, len(vector))
	for _, member := range vector {
		members = append(members, contextlock.Member{
			Kind: member.Kind, Name: member.Name, Version: member.Version,
			Commit: member.Commit, Source: "github.com/example/" + member.Name,
		})
	}
	return members
}

func writeVectorSnapshot(t *testing.T, home string, member contextlock.Member, snapshot sourceSignerSnapshot) {
	t.Helper()
	root := contextstore.EntryDir(home, member.Kind, member.Name, member.Commit)
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	switch member.Kind {
	case contextlock.KindContext:
		manifest := map[string]any{"schema_version": 1, "name": member.Name, "version": member.Version}
		modules := make([]any, 0, len(snapshot.SystemModules))
		for _, module := range snapshot.SystemModules {
			entry := map[string]any{"path": module.Path, "class": "system"}
			if module.Environments != nil {
				entry["environments"] = module.Environments
			}
			modules = append(modules, entry)
			path := filepath.Join(root, "context", filepath.FromSlash(module.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(module.Bytes), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if len(modules) > 0 {
			manifest["context"] = map[string]any{"modules": modules}
		}
		writeJSONFixture(t, filepath.Join(root, "agent-context.json"), manifest)
	case contextlock.KindMCP:
		var server map[string]any
		if err := json.Unmarshal(snapshot.MCP, &server); err != nil {
			t.Fatalf("decode MCP snapshot for %s: %v", member.Name, err)
		}
		if server == nil {
			t.Fatalf("MCP snapshot for %s has no declaration", member.Name)
		}
		writeJSONFixture(t, filepath.Join(root, "agent-mcp.json"), map[string]any{
			"schema_version": 1, "name": member.Name, "version": member.Version, "server": server,
		})
	}
}

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(fmt.Errorf("marshal fixture: %w", err))
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
