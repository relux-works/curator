package transaction

import (
	"strings"
	"testing"
)

func orderingRoot() RemovalEntry {
	return RemovalEntry{Kind: "directory", Mode: 0o755}
}

func orderingFile(path string) RemovalEntry {
	return RemovalEntry{
		RelativePath: path,
		Kind:         "file",
		Mode:         0o644,
		Digest:       "sha256:" + strings.Repeat("ab", 32),
	}
}

// TestValidateRemovalEntriesOrdersByUnsignedBytes pins the ordering contract of
// a removal manifest. The check walks the manifest forward carrying the
// preceding path, so this fixes what that walk has to reject: any pair that is
// equal or descending in unsigned-byte order, including the pair where a
// signed-byte reading would disagree.
func TestValidateRemovalEntriesOrdersByUnsignedBytes(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		entries []RemovalEntry
		accept  bool
	}{
		{
			name:    "root alone",
			entries: []RemovalEntry{orderingRoot()},
			accept:  true,
		},
		{
			name:    "strictly ascending",
			entries: []RemovalEntry{orderingRoot(), orderingFile("a"), orderingFile("b"), orderingFile("c")},
			accept:  true,
		},
		{
			// "~" is 0x7e and "é" opens with 0xc3. Unsigned-byte order puts
			// the tilde first; a signed reading would put it last, so this
			// pair separates the two orderings.
			name:    "high byte sorts after ascii",
			entries: []RemovalEntry{orderingRoot(), orderingFile("~"), orderingFile("é")},
			accept:  true,
		},
		{
			name:    "high byte before ascii is descending",
			entries: []RemovalEntry{orderingRoot(), orderingFile("é"), orderingFile("~")},
			accept:  false,
		},
		{
			name:    "duplicate path",
			entries: []RemovalEntry{orderingRoot(), orderingFile("a"), orderingFile("a")},
			accept:  false,
		},
		{
			name:    "descending pair",
			entries: []RemovalEntry{orderingRoot(), orderingFile("b"), orderingFile("a")},
			accept:  false,
		},
		{
			name:    "descending at the tail",
			entries: []RemovalEntry{orderingRoot(), orderingFile("a"), orderingFile("c"), orderingFile("b")},
			accept:  false,
		},
		{
			// A prefix is strictly shorter and therefore strictly smaller;
			// the reverse pairing is the descending one.
			name:    "prefix before extension",
			entries: []RemovalEntry{orderingRoot(), orderingFile("a"), orderingFile("ab")},
			accept:  true,
		},
		{
			name:    "extension before prefix",
			entries: []RemovalEntry{orderingRoot(), orderingFile("ab"), orderingFile("a")},
			accept:  false,
		},
		{
			name:    "root repeated",
			entries: []RemovalEntry{orderingRoot(), orderingRoot()},
			accept:  false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateRemovalEntries(testCase.entries)
			if testCase.accept && err != nil {
				t.Fatalf("canonical manifest was rejected: %v", err)
			}
			if !testCase.accept {
				if err == nil {
					t.Fatal("a manifest out of strict unsigned-byte order was accepted")
				}
				if !strings.Contains(err.Error(), "strict unsigned-byte order") {
					t.Fatalf("rejection names the wrong defect: %v", err)
				}
			}
		})
	}
}

// TestValidateRemovalEntriesStaysFailClosedWithoutARoot keeps the degenerate
// inputs on the rejecting side and, because the ordering walk no longer reads
// the slice behind the current index, proves they cannot fault.
func TestValidateRemovalEntriesStaysFailClosedWithoutARoot(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		entries []RemovalEntry
	}{
		{name: "nil manifest", entries: nil},
		{name: "empty manifest", entries: []RemovalEntry{}},
		{name: "first entry is not the root", entries: []RemovalEntry{orderingFile("a")}},
		{name: "root is not first", entries: []RemovalEntry{orderingFile("a"), orderingRoot()}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("validation faulted instead of rejecting: %v", recovered)
				}
			}()
			if err := validateRemovalEntries(testCase.entries); err == nil {
				t.Fatal("a manifest without a leading root entry was accepted")
			}
		})
	}
}

// TestValidateRemovalEntriesRejectsUnorderedTextDefects keeps the two halves of
// the ordering condition independent: a path that is not valid text is refused
// even when the manifest is otherwise ascending.
func TestValidateRemovalEntriesRejectsUnorderedTextDefects(t *testing.T) {
	entries := []RemovalEntry{orderingRoot(), orderingFile("a"), orderingFile("b\x00c")}
	if err := validateRemovalEntries(entries); err == nil {
		t.Fatal("a manifest carrying a NUL in a path was accepted")
	}
	entries = []RemovalEntry{orderingRoot(), orderingFile("a"), orderingFile("b" + string([]byte{0xff}))}
	if err := validateRemovalEntries(entries); err == nil {
		t.Fatal("a manifest carrying invalid UTF-8 in a path was accepted")
	}
}
