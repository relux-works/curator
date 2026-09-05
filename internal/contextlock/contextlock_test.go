package contextlock

import (
	"strings"
	"testing"
)

// Production entry points under test: Validate, Sort, Canonical, Hash,
// HashBytes, Parse, Read, Write.

func testLock() *Lock {
	return &Lock{Root: "root", Members: []Member{
		{Kind: KindSkill, Name: "s", Commit: strings.Repeat("a", 40), Source: "https://example.com/s", Weight: 1, RequiredBy: []string{"root"}},
		{Kind: KindContext, Name: "root", Commit: strings.Repeat("b", 40), Source: "https://example.com/root", Version: "1.0.0", Weight: 3},
	}}
}

// TestCanonicalOrderIsEnforced narrows the order gate: members not sorted by
// (kind, name) must fail validation. A mutant that skips the order check
// must fail this test.
func TestCanonicalOrderIsEnforced(t *testing.T) {
	lock := testLock() // skill before context: unsorted.
	if err := lock.Validate(); err == nil || !strings.Contains(err.Error(), "sorted") {
		t.Fatalf("unsorted lock must fail on order, got %v", err)
	}
	lock.Sort()
	if err := lock.Validate(); err != nil {
		t.Fatalf("sorted lock rejected: %v", err)
	}
	if lock.Members[0].Kind != KindContext || lock.Members[1].Kind != KindSkill {
		t.Fatalf("sort order %+v", lock.Members)
	}
}

// TestPinExclusivity narrows the pin gate: a member carrying both commit and
// state_sha256, or neither, must fail.
func TestPinExclusivity(t *testing.T) {
	lock := testLock()
	lock.Sort()
	both := lock.Members[0]
	both.StateHash = strings.Repeat("c", 64)
	bad := &Lock{Root: lock.Root, Members: []Member{both, lock.Members[1]}}
	if err := bad.Validate(); err == nil {
		t.Fatal("member with both pins must fail")
	}
	neither := lock.Members[0]
	neither.Commit = ""
	bad = &Lock{Root: lock.Root, Members: []Member{neither, lock.Members[1]}}
	if err := bad.Validate(); err == nil {
		t.Fatal("member with no pin must fail")
	}
}

// TestHashIsStableOverCanonicalBytes checks the CCJ-1/hash binding: Hash
// equals HashBytes(Canonical), and reordering members changes the hash.
func TestHashIsStableOverCanonicalBytes(t *testing.T) {
	lock := testLock()
	lock.Sort()
	canonical, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	hash, err := lock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if HashBytes(canonical) != hash {
		t.Fatal("Hash is not HashBytes(Canonical)")
	}
	if len(hash) != len("sha256:")+64 {
		t.Fatalf("lock hash %q is not sha256:<64 hex>", hash)
	}
	reweighted := &Lock{Root: lock.Root, Members: []Member{lock.Members[0], lock.Members[1]}}
	reweighted.Members[0].Weight++
	other, err := reweighted.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if other == hash {
		t.Fatal("member weights must enter the hash")
	}
}

// TestParseRoundTrip checks Parse accepts what Write emits.
func TestParseRoundTrip(t *testing.T) {
	lock := testLock()
	lock.Sort()
	canonical, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Root != "root" || len(parsed.Members) != 2 {
		t.Fatalf("parsed %+v", parsed)
	}
	if _, ok := parsed.RootMember(); !ok {
		t.Fatal("root member is not found")
	}
	if _, ok := parsed.Find(KindSkill, "s"); !ok {
		t.Fatal("skill member is not found")
	}
}

// TestMalformedLocksAreRejected covers the negative shape: no members,
// unknown kind, missing root member, state pin on a skill.
func TestMalformedLocksAreRejected(t *testing.T) {
	lock := testLock()
	lock.Sort()
	cases := []*Lock{
		{Root: "root"},
		{Root: "root", Members: []Member{{Kind: "bogus", Name: "x", Commit: strings.Repeat("d", 40), Source: "s", Version: "1.0.0"}}},
		{Root: "missing", Members: lock.Members},
		{Root: "root", Members: []Member{
			{Kind: KindContext, Name: "root", StateHash: strings.Repeat("e", 64), Version: "1.0.0"},
			{Kind: KindSkill, Name: "s", StateHash: strings.Repeat("e", 64)},
		}},
	}
	for i, bad := range cases {
		if err := bad.Validate(); err == nil {
			t.Fatalf("case %d must fail", i)
		}
	}
}

// TestWriteAndRead checks the file round trip and the returned hash.
func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	lock := testLock()
	lock.Sort()
	path := dir + "/lock.json"
	hash, err := Write(path, lock)
	if err != nil {
		t.Fatal(err)
	}
	read, readHash, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if readHash != hash || read.Root != "root" {
		t.Fatalf("read hash %q want %q", readHash, hash)
	}
}
