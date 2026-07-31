package transaction

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mustSymlink creates one link or skips the test on a platform without them,
// and returns the destination the host actually recorded for it.
//
// A link's content is path text in the host's own syntax. os.Symlink rewrites
// the destination with filepath.FromSlash before it reaches the filesystem,
// because '/' is not a separator inside a Windows reparse point, so the string
// a later os.Readlink returns is the host's spelling of what was asked for and
// not the argument itself. Expectations below are that recorded string, which
// keeps "restored exactly" an exact comparison against the destination that was
// really there rather than against a POSIX literal only one family of hosts can
// satisfy.
func mustSymlink(t *testing.T, destination, path string) string {
	t.Helper()
	if err := os.Symlink(destination, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	return mustReadlink(t, path)
}

func mustReadlink(t *testing.T, path string) string {
	t.Helper()
	destination, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("expected a symbolic link at %s: %v", path, err)
	}
	return destination
}

func entryTarget(t *testing.T, class, identifier, live, source string) Target {
	t.Helper()
	preimage, err := DigestTarget(KindEntry, live)
	if err != nil {
		t.Fatal(err)
	}
	return Target{
		Class: class, Identifier: identifier, Kind: KindEntry,
		LivePath: live, StagedSource: source, PreimageDigest: preimage,
	}
}

// TestEntryKindDigestsLinksAndBytesWhileByteKindStaysStrict pins the boundary
// between the two kinds: only an entry target may be a link, and a link's digest
// is exactly its destination string.
func TestEntryKindDigestsLinksAndBytesWhileByteKindStaysStrict(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	mustWrite(t, file, "bytes")
	link := filepath.Join(root, "link")
	mustSymlink(t, "file", link)

	if _, err := DigestTarget(KindBytes, link); err == nil {
		t.Fatal("a byte target accepted a symbolic link")
	}
	linkDigest, err := DigestTarget(KindEntry, link)
	if err != nil {
		t.Fatal(err)
	}
	if !validDigest(linkDigest) || linkDigest == DigestAbsent {
		t.Fatalf("link digest = %q", linkDigest)
	}

	// The destination string is the whole content: a link to different bytes
	// under the same name digests the same, and a re-pointed link does not.
	other := filepath.Join(root, "other-link")
	mustSymlink(t, "file", other)
	if again, err := DigestTarget(KindEntry, other); err != nil || again != linkDigest {
		t.Fatalf("identical destinations digest differently: %q vs %q (%v)", again, linkDigest, err)
	}
	repointed := filepath.Join(root, "repointed")
	mustSymlink(t, "elsewhere", repointed)
	if changed, err := DigestTarget(KindEntry, repointed); err != nil || changed == linkDigest {
		t.Fatalf("a different destination digests the same: %q (%v)", changed, err)
	}

	// An entry target that is not a link is still digested as bytes, so one
	// target can carry a directory today and a link tomorrow.
	fileDigest, err := DigestTarget(KindEntry, file)
	if err != nil {
		t.Fatal(err)
	}
	if byteDigest, err := DigestPath(file); err != nil || byteDigest != fileDigest {
		t.Fatalf("entry digest of a regular file = %q, want the byte digest %q (%v)", fileDigest, byteDigest, err)
	}
	if absent, err := DigestTarget(KindEntry, filepath.Join(root, "missing")); err != nil || absent != DigestAbsent {
		t.Fatalf("absent entry digest = %q, %v", absent, err)
	}
}

// TestCommitAndRollbackRestoreALinkExactly proves an owned link is replaced and
// restored with its exact destination, including across a mode transition.
func TestCommitAndRollbackRestoreALinkExactly(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		prepare func(t *testing.T, live string) string
		stage   func(t *testing.T, source string)
	}{
		{
			name:    "absent to link",
			prepare: func(*testing.T, string) string { return "" },
			stage:   func(t *testing.T, source string) { mustSymlink(t, "../canonical/skill", source) },
		},
		{
			name: "link to repointed link",
			prepare: func(t *testing.T, live string) string {
				return "link:" + mustSymlink(t, "../old/skill", live)
			},
			stage: func(t *testing.T, source string) { mustSymlink(t, "../canonical/skill", source) },
		},
		{
			name: "link to copied tree",
			prepare: func(t *testing.T, live string) string {
				return "link:" + mustSymlink(t, "../old/skill", live)
			},
			stage: func(t *testing.T, source string) { mustWrite(t, filepath.Join(source, "SKILL.md"), "copied") },
		},
		{
			name: "copied tree to link",
			prepare: func(t *testing.T, live string) string {
				mustWrite(t, filepath.Join(live, "SKILL.md"), "old-copy")
				return "tree:old-copy"
			},
			stage: func(t *testing.T, source string) { mustSymlink(t, "../canonical/skill", source) },
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			for _, outcome := range []string{"commit", "rollback"} {
				t.Run(outcome, func(t *testing.T) {
					root := t.TempDir()
					liveRoot := filepath.Join(root, "adapter")
					mustMkdirAll(t, liveRoot)
					live := filepath.Join(liveRoot, "skill")
					before := testCase.prepare(t, live)
					source := filepath.Join(root, "stage", "skill")
					mustMkdirAll(t, filepath.Dir(source))
					testCase.stage(t, source)
					desired, err := DigestTarget(KindEntry, source)
					if err != nil {
						t.Fatal(err)
					}

					var engine *Engine
					if outcome == "rollback" {
						engine = mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{
							Fault: func(event Event) error {
								if event.Point == PointTargetCommitted {
									return errors.New("injected failure after the entry committed")
								}
								return nil
							},
						}))
					} else {
						engine = mustEngine(t, filepath.Join(root, "home"))
					}
					journal, err := engine.Prepare(testLock{}, Plan{
						TransactionID:   "txn-entry",
						ProjectIdentity: "/project",
						Targets:         []Target{entryTarget(t, "60-adapter-ledger", "mirror", live, source)},
					})
					if err != nil {
						t.Fatal(err)
					}
					commitErr := engine.Commit(testLock{}, journal.TransactionID)

					if outcome == "commit" {
						if commitErr != nil {
							t.Fatal(commitErr)
						}
						if got, err := DigestTarget(KindEntry, live); err != nil || got != desired {
							t.Fatalf("committed entry digest = %q, want %q (%v)", got, desired, err)
						}
						return
					}
					if commitErr == nil {
						t.Fatal("the injected failure did not fail the commit")
					}
					assertEntryRestored(t, live, before)
					// Nothing of the transaction may survive a rollback.
					entries, err := os.ReadDir(liveRoot)
					if err != nil {
						t.Fatal(err)
					}
					for _, entry := range entries {
						if strings.HasPrefix(entry.Name(), ".curator-txn-") {
							t.Fatalf("rollback left the sidecar %s behind", entry.Name())
						}
					}
				})
			}
		})
	}
}

func assertEntryRestored(t *testing.T, live, before string) {
	t.Helper()
	switch {
	case before == "":
		if _, err := os.Lstat(live); !os.IsNotExist(err) {
			t.Fatalf("rollback left an entry where none existed: %v", err)
		}
	case strings.HasPrefix(before, "link:"):
		if got := mustReadlink(t, live); got != strings.TrimPrefix(before, "link:") {
			t.Fatalf("restored link destination = %q, want %q", got, strings.TrimPrefix(before, "link:"))
		}
	default:
		info, err := os.Lstat(live)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("restored entry is not the prior directory: %v", err)
		}
		if got := mustRead(t, filepath.Join(live, "SKILL.md")); got != strings.TrimPrefix(before, "tree:") {
			t.Fatalf("restored tree content = %q, want %q", got, strings.TrimPrefix(before, "tree:"))
		}
	}
}

// TestRecoveryFinishesAPreparedLinkTransaction proves an interrupted link
// commit is completed by home-scoped recovery rather than left half-applied.
func TestRecoveryFinishesAPreparedLinkTransaction(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	liveRoot := filepath.Join(root, "adapter")
	mustMkdirAll(t, liveRoot)
	live := filepath.Join(liveRoot, "skill")
	before := mustSymlink(t, "../old/skill", live)
	source := filepath.Join(root, "stage", "skill")
	mustMkdirAll(t, filepath.Dir(source))
	prepared := mustSymlink(t, "../canonical/skill", source)

	engine := mustEngine(t, home)
	if _, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-entry-recovery",
		ProjectIdentity: "/project",
		Targets:         []Target{entryTarget(t, "60-adapter-ledger", "mirror", live, source)},
	}); err != nil {
		t.Fatal(err)
	}
	if got := mustReadlink(t, live); got != before {
		t.Fatalf("preparation already replaced the live link: %q", got)
	}
	if err := mustEngine(t, home).Recover(testLock{}); err != nil {
		t.Fatal(err)
	}
	if got := mustReadlink(t, live); got != prepared {
		t.Fatalf("recovery left the link at %q, want the prepared destination %q", got, prepared)
	}
}

// TestEntryRemovalRestoresTheExactLink proves a removal target — how a stale
// adapter mirror leaves — puts the exact prior link back when a later class
// fails, and removes it when the transaction succeeds.
func TestEntryRemovalRestoresTheExactLink(t *testing.T) {
	for _, outcome := range []string{"commit", "rollback"} {
		t.Run(outcome, func(t *testing.T) {
			root := t.TempDir()
			liveRoot := filepath.Join(root, "adapter")
			mustMkdirAll(t, liveRoot)
			stale := filepath.Join(liveRoot, "stale")
			staleDestination := mustSymlink(t, "../canonical/stale", stale)
			ledgerLive := filepath.Join(liveRoot, "ledger.json")
			mustWrite(t, ledgerLive, "old-ledger")
			ledgerSource := filepath.Join(root, "stage", "ledger.json")
			mustWrite(t, ledgerSource, "new-ledger")

			var engine *Engine
			if outcome == "rollback" {
				engine = mustEngine(t, filepath.Join(root, "home"), WithHooks(Hooks{
					Fault: func(event Event) error {
						if event.Point == PointAfterBackup && event.Class == "80-removal" {
							return errors.New("injected failure at the removal class")
						}
						return nil
					},
				}))
			} else {
				engine = mustEngine(t, filepath.Join(root, "home"))
			}

			ledgerPreimage, err := DigestPath(ledgerLive)
			if err != nil {
				t.Fatal(err)
			}
			stalePreimage, err := DigestTarget(KindEntry, stale)
			if err != nil {
				t.Fatal(err)
			}
			journal, err := engine.Prepare(testLock{}, Plan{
				TransactionID:   "txn-entry-removal",
				ProjectIdentity: "/project",
				Targets: []Target{
					{
						Class: "60-adapter-ledger", Identifier: "ledger",
						LivePath: ledgerLive, StagedSource: ledgerSource, PreimageDigest: ledgerPreimage,
					},
					{
						Class: "80-removal", Identifier: "adapter/stale", Kind: KindEntry,
						LivePath: stale, PreimageDigest: stalePreimage,
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			commitErr := engine.Commit(testLock{}, journal.TransactionID)

			if outcome == "commit" {
				if commitErr != nil {
					t.Fatal(commitErr)
				}
				if _, err := os.Lstat(stale); !os.IsNotExist(err) {
					t.Fatalf("the stale link survived a successful commit: %v", err)
				}
				if got := mustRead(t, ledgerLive); got != "new-ledger" {
					t.Fatalf("ledger = %q, want the committed replacement", got)
				}
				return
			}
			if commitErr == nil {
				t.Fatal("the injected failure did not fail the commit")
			}
			if got := mustReadlink(t, stale); got != staleDestination {
				t.Fatalf("restored stale link = %q, want its exact prior destination %q", got, staleDestination)
			}
			if got := mustRead(t, ledgerLive); got != "old-ledger" {
				t.Fatalf("ledger = %q, want the restored preimage", got)
			}
		})
	}
}

// TestEntryTargetsDoNotAliasTheirDestination proves the namespace guard reads an
// owned entry as itself: a mirror link and the canonical directory it points at
// are independent targets, while two byte targets aliasing one object are still
// refused.
func TestEntryTargetsDoNotAliasTheirDestination(t *testing.T) {
	root := t.TempDir()
	canonical := filepath.Join(root, "canonical", "skill")
	mustWrite(t, filepath.Join(canonical, "SKILL.md"), "old")
	adapterRoot := filepath.Join(root, "adapter")
	mustMkdirAll(t, adapterRoot)
	mirror := filepath.Join(adapterRoot, "skill")
	mustSymlink(t, "../canonical/skill", mirror)

	canonicalSource := filepath.Join(root, "stage", "canonical")
	mustWrite(t, filepath.Join(canonicalSource, "SKILL.md"), "new")
	mirrorSource := filepath.Join(root, "stage", "mirror")
	stagedDestination := mustSymlink(t, "../canonical/skill", mirrorSource)

	canonicalDigest, err := DigestPath(canonical)
	if err != nil {
		t.Fatal(err)
	}
	engine := mustEngine(t, filepath.Join(root, "home"))
	journal, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-entry-alias",
		ProjectIdentity: "/project",
		Targets: []Target{
			{
				Class: "10-context", Identifier: "skill",
				LivePath: canonical, StagedSource: canonicalSource, PreimageDigest: canonicalDigest,
			},
			entryTarget(t, "60-adapter-ledger", "mirror", mirror, mirrorSource),
		},
	})
	if err != nil {
		t.Fatalf("a mirror link and its destination were rejected as aliases: %v", err)
	}
	if err := engine.Commit(testLock{}, journal.TransactionID); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, filepath.Join(canonical, "SKILL.md")); got != "new" {
		t.Fatalf("canonical content = %q, want the committed replacement", got)
	}
	if got := mustReadlink(t, mirror); got != stagedDestination {
		t.Fatalf("mirror link = %q, want the committed destination %q", got, stagedDestination)
	}
}

// TestEntryTargetRejectsGenerationExpectations keeps the one combination the
// kind cannot express out of the journal: a link has no generation file.
func TestEntryTargetRejectsGenerationExpectations(t *testing.T) {
	root := t.TempDir()
	liveRoot := filepath.Join(root, "adapter")
	mustMkdirAll(t, liveRoot)
	live := filepath.Join(liveRoot, "skill")
	source := filepath.Join(root, "stage", "skill")
	mustMkdirAll(t, filepath.Dir(source))
	mustSymlink(t, "../canonical/skill", source)

	engine := mustEngine(t, filepath.Join(root, "home"))
	if _, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-entry-generation",
		ProjectIdentity: "/project",
		Targets: []Target{{
			Class: "60-adapter-ledger", Identifier: "mirror", Kind: KindEntry,
			LivePath: live, StagedSource: source, ExpectedGeneration: "v1",
		}},
	}); err == nil {
		t.Fatal("an entry target with a generation expectation was accepted")
	}
	if _, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-entry-kind",
		ProjectIdentity: "/project",
		Targets: []Target{{
			Class: "60-adapter-ledger", Identifier: "mirror", Kind: TargetKind("bogus"),
			LivePath: live, StagedSource: source, PreimageDigest: DigestAbsent,
		}},
	}); err == nil {
		t.Fatal("an unknown target kind was accepted")
	}
}

// TestEntryStagingRefusesLinksInsideATree keeps the dangerous case refused: only
// the entry itself may be a link, never something buried in a copied tree.
func TestEntryStagingRefusesLinksInsideATree(t *testing.T) {
	root := t.TempDir()
	liveRoot := filepath.Join(root, "adapter")
	mustMkdirAll(t, liveRoot)
	live := filepath.Join(liveRoot, "skill")
	source := filepath.Join(root, "stage", "skill")
	mustWrite(t, filepath.Join(source, "SKILL.md"), "content")
	mustSymlink(t, "../../../etc/passwd", filepath.Join(source, "escape"))

	engine := mustEngine(t, filepath.Join(root, "home"))
	if _, err := engine.Prepare(testLock{}, Plan{
		TransactionID:   "txn-entry-nested-link",
		ProjectIdentity: "/project",
		Targets: []Target{{
			Class: "60-adapter-ledger", Identifier: "mirror", Kind: KindEntry,
			LivePath: live, StagedSource: source, PreimageDigest: DigestAbsent,
		}},
	}); err == nil {
		t.Fatal("a staged tree containing a link was accepted for an entry target")
	}
}
