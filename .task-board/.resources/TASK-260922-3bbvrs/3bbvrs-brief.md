# TASK-260922-3bbvrs — gap ledger for published-case consumers (curator)

Control root: curator; work only in your Story worktree (STORY-260922-2goxjs). Read `campaign-producer-rules.md`
and this task's description/AC (exact). Landing gate = hosted CI, once, at handoff.

Generalise the existing harness in `internal/crossconformance/draftsources_semantic_test.go` (each case classified
driven | known-gap | bound | skipped; the four counts tallied and the total asserted against the published case
count) to the consumers that render EVERY published case — `internal/config` `TestManagerConfigV2Vectors` and its
system-config-v2 sibling first; then find any other family that iterates a published case list and list it.
Add ONE committed ledger (suggest `.github/ci/conformance-gaps.tsv`: family, case id, owning board element,
one-line reason), read by those tests. The ratchet is the point: a known-gap row that starts PASSING must FAIL the
gate until removed from the ledger; an unlisted failing case fails; a vanished case fails (tally mismatch).
At the CURRENT pin (rc.12) the ledger may be empty — prove the mechanism with fixtures and three narrowing
mutants (drop the tally, accept an unlisted gap, let a passing gap stay listed), each killed by a named test.
Do NOT move SPEC_PIN (that is TASK-260922-18ex37). Docs: a gap row is an owed implementation, never an accepted
deviation. CHANGELOG. Attach `TASK-260922-3bbvrs_results.md` and hand off.
