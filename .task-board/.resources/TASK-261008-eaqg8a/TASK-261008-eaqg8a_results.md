N7 developer results — TASK-261008-eaqg8a
(N7: install.Global reconciles a recognized manager-owned legacy forwarding link)

1. Production fix (boundary: install.Global via globalbins.StageForwarding/ownedTarget)
- internal/globalbins/globalbins.go: new ownedLegacyLink(path, canonical, platform).
  It returns true only for a live symlink that ownedTarget already recognizes
  (target resolves to the canonical shim); foreign or unrecognized links never
  qualify. No change to ownedTarget itself.
- internal/globalbins/stage.go: StageForwarding now records a recognized legacy
  link as a managed directory entry — ReplaceEntry for desired shims,
  RemoveEntry for stale removals — instead of a byte target. Everything else
  keeps the previous Replace/Remove path.
- Why this is the report's fix: the publication recheck
  (internal/staging/boundaries.go) refuses byte targets whose live final
  component is a link ("source_output_overlap: ... is now a link"). A KindEntry
  target is journaled, preimaged (DigestTarget digests the exact link
  destination string), verified, and rolled back as the link it is, through the
  same transaction machinery the adapter mirrors already use; the per-write
  RecheckOne accepts entry links. install.Global merges the forwarding plan
  into the journaled commit (internal/install/global.go, journalPlan), so the
  preimage is read under the home lock and post-prepare drift still fails.
- Foreign-file and changed-preimage refusals are untouched: non-links and
  unrecognized links never take the entry path (unmanagedConflict still
  excludes foreign bytes from the plan; byte-target preimage comparison still
  fails post-prepare swaps). The general prohibition on following links during
  managed writes is not weakened — only a link that ownedTarget recognizes AND
  that carries a ledger entry becomes an entry target.
- CHANGELOG.md: re-added the Unreleased/Fixed line from the attached
  N7-changelog.patch (operator-visible wording, task id cited).

2. Regression test (through the production entry, no "wave2" in the name)
- internal/install/global_legacy_link_test.go:
  TestGlobalReconcilesOwnedLegacyForwardingLink drives install.Global over four
  user-bin states: owned-regular, foreign-regular, owned-legacy-link (fixture
  emitted by globalbins.Refresh itself, as in the report), foreign-after-prepare
  (fault hook swaps foreign bytes in at PointPrepared). Owned states must
  reconcile to ok (the reconciled shim is asserted to be a regular file with
  the current forwarding launcher bytes); foreign bytes must be preserved and
  the post-prepare drift must fail.
- Windows: the owned-legacy-link subtest skips with reason "Unix legacy-link
  reconciliation is exercised on Linux and macOS", matching the existing
  platform-control skip class "(is|are) exercised (on|by)" in
  .github/ci/skip-classes.tsv (same convention as
  internal/globalbins/globalbins_test.go). The first hosted attempt used an
  unclassified skip reason and failed ONLY the windows platform-case gate; the
  reason was fixed and both hosted runs were repeated from the corrected
  candidate. No CI gate table was modified.

3. Local validation (R223: compile-only on the mini, zero local go test)
- go vet ./... → exit 0
- go build ./... → exit 0
- gofmt clean on all touched files
- No `go test` was executed locally in this run.

4. Hosted evidence (full CI matrix; run URLs below)
Adaptation note: .github/workflows/ci.yml push trigger covers only main and
gate/**, so a scratch/* push would never start CI. Both snapshots were pushed
as gate/TASK-261008-eaqg8a-*/<stamp> branches (same throwaway semantics as
scripts/remote-gate.sh: temp-index snapshot, worktree/index/HEAD untouched)
and deleted after recording. Run history persists by run id.
- GREEN, exact candidate 08fca25d (fix + test + CHANGELOG, tree byte-identical
  to the worktree, verified with cmp on all 4 files):
  https://github.com/relux-works/curator/actions/runs/37868830965
  conclusion: success. Every job green incl. Lint, all Test/Race lanes
  (ubuntu, macos, windows), drivers, gate self-tests, interop conformance.
  Ubuntu served stream: 0 failing tests; all four
  TestGlobalReconcilesOwnedLegacyForwardingLink subtests pass.
- RED, same candidate with ONLY the production fix reverted (fb403640;
  red-vs-green diff = internal/globalbins/globalbins.go + stage.go, 24
  insertions, nothing else):
  https://github.com/relux-works/curator/actions/runs/37868866996
  conclusion: failure. Failed jobs: the 4 unix Test/Race lanes only; windows
  Test, Lint, drivers, gate self-tests all passed. The ONLY failing test in
  the ubuntu served stream is
  internal/install :: TestGlobalReconcilesOwnedLegacyForwardingLink/owned-legacy-link;
  owned-regular, foreign-regular, and foreign-after-prepare all pass there.
- Superseded first pair (same mutant signal, kept as supporting evidence;
  branches deleted): green 37864145427 failed solely on the windows
  platform-case gate for the then-unclassified skip reason (all tests passed);
  red 37864686172 failed with exactly owned-legacy-link red and the three
  sibling subtests green on ubuntu.
  https://github.com/relux-works/curator/actions/runs/37864145427
  https://github.com/relux-works/curator/actions/runs/37864686172
- All four evidence branches deleted; `git ls-remote origin
  'refs/heads/gate/TASK-261008-eaqg8a-*'` returns empty.

5. Negative controls (report section N7)
- foreign-regular: foreign bytes preserved, green in green2 AND red2.
- foreign-after-prepare: preimage drift fails, foreign bytes preserved, green
  in green2 AND red2.
- owned-regular: reconciles to ok in both.

6. Working-tree state and notes
- Worktree left UNCOMMITTED with exactly 4 paths: M CHANGELOG.md,
  M internal/globalbins/globalbins.go, M internal/globalbins/stage.go,
  ?? internal/install/global_legacy_link_test.go. No commit made on the story
  branch (snapshot commits exist only as deleted remote gate branches).
- A stray local branch created by pushing the first red snapshot to the
  clone's origin (the worktree path) instead of GitHub was deleted; the commit
  was re-pushed to GitHub and the local ref removed. No other worktree
  pollution.
- LOGBOOK.md untouched per the brief ("No LOGBOOK edits"); the skip-gate
  lesson is recorded here instead.
- DoD item 10 (logbook) is therefore intentionally left unchecked: the brief
  forbids the edit, and everything relevant is in this resource.
