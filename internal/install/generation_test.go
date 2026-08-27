package install

import (
	"os"
	"path/filepath"
	"testing"
)

// A declaration observation is sound only if its generation belongs to the bytes
// the parser consumed. These cases lock that binding at the read itself, so a
// future refactor cannot go back to digesting the path as a second operation and
// reopen the A -> B -> A window (see generation.go).

// writeDocument writes one declaration payload the way every supported manifest
// writer does: os.WriteFile, in place, on the inode the path already names.
func writeDocument(t *testing.T, path, payload string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
}

// generationOfBytes reports the generation the manager would record for payload,
// derived from a separate file that holds exactly those bytes. It never calls the
// digest helper under test directly, so an assertion against it cannot pass by
// agreeing with itself.
func generationOfBytes(t *testing.T, payload []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "reference.json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	reference, err := readDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	return reference.generation
}

// onceAfterOpen installs a one-shot read-window hook and proves it fired.
func onceAfterOpen(t *testing.T, target string, change func()) {
	t.Helper()
	fired := false
	afterDocumentOpen = func(path string) {
		if fired || path != target {
			return
		}
		fired = true
		change()
	}
	t.Cleanup(func() {
		afterDocumentOpen = nil
		if !fired {
			t.Fatalf("the read window never opened for %s", target)
		}
	})
}

// TestReadDocumentBindsGenerationToBytesRewrittenInPlace covers the writer shape
// Curator actually ships: os.WriteFile truncates and rewrites the inode the
// reader already holds open, so the read consumes the *new* bytes. The recorded
// generation has to be the generation of those bytes and not of whatever the path
// held when the read started.
func TestReadDocumentBindsGenerationToBytesRewrittenInPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Skillfile.json")
	writeDocument(t, path, `{"generation": "A"}`)
	transient := `{"generation": "B", "padding": "wider than A"}`
	onceAfterOpen(t, path, func() { writeDocument(t, path, transient) })

	parsed, err := readDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(parsed.payload) != transient {
		t.Fatalf("the read did not consume the in-place rewrite, so this case proves nothing: %q", parsed.payload)
	}
	if want := generationOfBytes(t, parsed.payload); parsed.generation != want {
		t.Fatalf("the generation does not belong to the bytes the parser gets: got %s, bytes are %s",
			parsed.generation, want)
	}
}

// TestReadDocumentBindsGenerationToBytesReplacedByRename covers the other writer
// shape: an atomic rename leaves the open handle on the old inode, so the read
// consumes the *previous* bytes while the path already holds the new ones. The
// generation must follow the bytes that were returned, which is exactly what a
// path digest taken around the read cannot do — and the mismatch it leaves is
// what makes the under-lock recheck restart.
func TestReadDocumentBindsGenerationToBytesReplacedByRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Skillfile.json")
	settled := `{"generation": "A"}`
	writeDocument(t, path, settled)
	replacement := `{"generation": "B"}`
	onceAfterOpen(t, path, func() {
		staged := filepath.Join(dir, "staged.json")
		writeDocument(t, staged, replacement)
		if err := os.Rename(staged, path); err != nil {
			t.Fatal(err)
		}
	})

	parsed, err := readDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(parsed.payload) != settled {
		t.Fatalf("the rename should have left the read on the old inode: %q", parsed.payload)
	}
	if want := generationOfBytes(t, parsed.payload); parsed.generation != want {
		t.Fatalf("the generation does not belong to the bytes the parser gets: got %s, bytes are %s",
			parsed.generation, want)
	}
	if current := documentGeneration(path); current == parsed.generation {
		t.Fatal("the recheck cannot see the replacement, so a stale closure would commit")
	}
}

// TestDeclarationGenerationIsContentAddressed is the property the ABA cases
// depend on: a generation identifies bytes, so a byte-identical restoration is
// indistinguishable from never having changed, and only binding the generation to
// the parsed bytes can catch a transient generation in between.
func TestDeclarationGenerationIsContentAddressed(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "Skillfile.json")
	if absent := documentGeneration(path); absent != documentAbsent {
		t.Fatalf("an absent document must be explicit: %s", absent)
	}

	writeDocument(t, path, `{"generation": "A"}`)
	settled := documentGeneration(path)
	if settled == documentAbsent {
		t.Fatal("a present document must not read as absent")
	}

	writeDocument(t, path, `{"generation": "B"}`)
	if changed := documentGeneration(path); changed == settled {
		t.Fatal("different bytes must produce different generations")
	}

	writeDocument(t, path, `{"generation": "A"}`)
	if restored := documentGeneration(path); restored != settled {
		t.Fatal("byte-identical content must produce the same generation")
	}

	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if rechecked := documentGeneration(path); rechecked != settled {
		t.Fatal("a mode change selects nothing installed and must not restart a run")
	}
}

// TestDocumentGenerationIsStableWhenUnreadable keeps the recheck honest about
// state it cannot read: it has to reproduce the same marker rather than restart
// forever or, worse, compare unequal markers and give up after MaxRestarts.
func TestDocumentGenerationIsStableWhenUnreadable(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "Skillfile.json")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	first := documentGeneration(path)
	if first == documentAbsent {
		t.Fatal("a directory in place of a document is not absence")
	}
	if second := documentGeneration(path); second != first {
		t.Fatalf("an unreadable document must observe stably: %s then %s", first, second)
	}
}

// TestObservationsRecheckDocumentsAgainstTheParsedBytes drives the recheck
// directly: an observation recorded from a transient generation must report a
// change once the file settles back, which is the ABA restart the install cases
// below rely on. A marker observation keeps its path-digest reader.
func TestObservationsRecheckDocumentsAgainstTheParsedBytes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "Skillfile.json")
	writeDocument(t, path, `{"generation": "A"}`)
	transient := generationOfBytes(t, []byte(`{"generation": "B"}`))

	observed := newObservations()
	observed.observeDocument(projectManifestKey, path, transient)
	if restart := observed.recheck(BuildPlan{}, nil); restart == nil {
		t.Fatal("a generation the file never settled at must restart closure resolution")
	}

	observed = newObservations()
	settled, err := readDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	observed.observeDocument(projectManifestKey, path, settled.generation)
	marker := filepath.Join(dir, "marker.json")
	writeDocument(t, marker, "{}")
	observed.observe("marker/project/skill-a", marker)
	if restart := observed.recheck(BuildPlan{}, nil); restart != nil {
		t.Fatalf("unchanged state must not restart: %v", restart)
	}

	writeDocument(t, marker, `{"changed": true}`)
	if restart := observed.recheck(BuildPlan{}, nil); restart == nil {
		t.Fatal("a changed marker must still restart through its path digest")
	}
}
