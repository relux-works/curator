package contextlock

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
)

func TestUnversionedLockObjectFollowsManagedWriterSwitch(t *testing.T) {
	prior := hashing.EnableV2Writers
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
	hashing.EnableV2Writers = false
	legacy := (&Lock{Root: "root"}).Object()
	if legacy["schema_version"] != SchemaVersion {
		t.Fatalf("default lock schema = %v, want %d", legacy["schema_version"], SchemaVersion)
	}
	if _, present := legacy["hash_version"]; present {
		t.Fatal("default lock shape carries hash_version")
	}

	hashing.EnableV2Writers = true
	versioned := (&Lock{Root: "root"}).Object()
	if versioned["schema_version"] != SchemaVersion2 || versioned["hash_version"] != 2 {
		t.Fatalf("enabled lock versions = schema:%v hash:%v, want 2/2", versioned["schema_version"], versioned["hash_version"])
	}
}

func TestResolvedDeltaIncludesStatePinWhenHashVersionChanges(t *testing.T) {
	member := Member{Kind: KindContext, Name: "root", Version: "1.0.0", StateHash: "sha256:" + strings.Repeat("a", 64)}
	legacy := &Lock{SchemaVersion: SchemaVersion, Root: "root", Members: []Member{member}}
	versioned := &Lock{SchemaVersion: SchemaVersion2, HashVersion: 2, Root: "root", Members: []Member{member}}

	deltas := ResolvedDelta(legacy, versioned)
	if len(deltas) != 1 || deltas[0].Old == nil || deltas[0].New == nil {
		t.Fatalf("state pin version change produced deltas %+v, want one moved member", deltas)
	}
}
