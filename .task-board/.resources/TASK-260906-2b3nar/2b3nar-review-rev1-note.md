# Review note for TASK-260906-2b3nar revision 1 (orchestrator, binding) — gitops per-component fold

Read the brief `2b3nar-brief.md` (precondition), the producer's `TASK-260906-2b3nar_results.md`, and
the source finding (N1 of the acquisition cycle-2 review,
`.task-board/.resources/TASK-260905-3r30t1/TASK-260905-3r30t1_review-verdict-rev5.md`). Reuse the
hosted gate evidence for the exact candidate tree; do not rerun the full landing suite.

`planWrites` now tracks every lower-cased ancestor prefix alongside the folded full paths, probes the
destination once through the existing `destinationFoldsCase`, and refuses a prefix (or file↔directory)
fold with the EXISTING `duplicate platform path in git snapshot: %q` class, in the pre-pass before
`writeBlobs` starts.

Judge:
1. **Completeness of the fold rule.** Attack it with shapes the producer did not list: three-way
   folds (`A/x`, `a/y`, `A/z`), a fold that appears only after a deeper component
   (`a/B/x` + `a/b/y` where `a` matches exactly), Unicode case folding (`İ`/`i`, `K`/K kelvin,
   NFC vs NFD) — decide whether `strings.ToLower` is the right predicate for the destination
   filesystem's folding and whether a mismatch is a finding or an accepted bound (APFS/HFS+ fold
   differently from NTFS; the pre-existing full-path check had the same predicate, so a NEW gap is
   the bar, not the inherited one). Also check the empty-path/root case and a tree whose only fold
   is between a gitlink/symlink entry and a directory.
2. **No behaviour change on case-sensitive destinations**: both spellings must still extract
   byte-exact; exact duplicates keep the `duplicate path` message. Prove it (the producer used a
   scratch case-sensitive APFS volume; if you cannot create one, say so and bound the claim).
3. **Atomicity**: a refused extraction must write NOTHING. Verify the refusal really is in the
   pre-pass for every new row, including the file-first ordering case (c), where the old code failed
   mid-stream with a raw `MkdirAll` error.
4. **Mutants**: re-run at least two of m1–m3 yourself and add one of your own (e.g. track prefixes
   but compare them case-sensitively, or probe the destination only for the full-path check).
5. **Ledger**: 4 rows added to `.github/ci/platform-cases.tsv` with host-capability tolerances;
   confirm the required/tolerated lanes match what the tests actually do and that no new skip class
   was introduced (`ledger-consistency.sh` must pass).
6. CHANGELOG `Fixed` entry present; production call site is `gitops.Extract`.

Recorded out-of-scope finding to judge, not to fix: the producer reports
`go test ./internal/envprofile/ -count=1` (full package) timing out in
`TestImportCorruptMarkerIsLoss`, passing alone in 7 s, with `import.go` unable to reach the changed
code. Say whether the evidence supports "unrelated pre-existing flake"; if it does, it is a separate
bug, not a blocker here.

Record exactly one verdict: `accept_cr(TASK-260906-2b3nar, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into the control root's LOGBOOK.md.
