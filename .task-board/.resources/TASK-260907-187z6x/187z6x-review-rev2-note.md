# Review note for TASK-260907-187z6x revision 2 (orchestrator, binding) — rework 1

Revision 2 answers your revision-1 verdict (`TASK-260907-187z6x_review-verdict-rev1.md`): R1 — the
install error branch in `cmd/curator/profile.go` now honours the `updated` flag, so a partially
activated same-source reinstall prints `updated profile <name> (lock …)` (exit code and diagnostics
unchanged), and row 2 now asserts that operator line; the two neither-flag rows gained the missing
no-backup assertions. R2 — the row-2 description now says the blocked `claude_code` home is the one not
written or backed up while the other adapters switch (§9.2 partial). Everything you VERIFIED at
revision 1 (normative parity, activation semantics, takeover safety, both mutants, ledger) stands.

Judge the delta: re-run your own row-2 overlay (must now PASS), add a mutant that restores the
unconditional `installed` line (must be killed by row 2), and confirm the path-root sibling still
reports identically. Trunk moved to 48da2690 (rc.12 pin) — check the refreshed base did not change
any row's behaviour.

Record exactly one verdict: `accept_cr(TASK-260907-187z6x, revision=2, evidence=<your outcome
resource>)` on ACCEPT, or changes-requested with file:line and reproduction. No LOGBOOK.md writes.
