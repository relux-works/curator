# Review note for TASK-260922-cww1ov revision 2 (orchestrator, binding) — F-C3, FINAL leaf of STORY-260922-1cenbr

Revision 2 = rework 1 for your revision-1 verdict (`TASK-260922-cww1ov_review-verdict-rev1.md`;
brief `cww1ov-rework-1.md`). Scope of the delta ONLY:
1. **R1 (your P2)**: the no-copy row now scans the ENTIRE temporary manager home — state, journal,
   backup and temp paths included — not just environments/profiles/native roots, and also scans at
   the interrupted-apply hook so transient journal content is covered; a narrowing mutant that
   copies ONLY into manager state must be killed by the named production-entry row. Re-run your own
   `review-state-secret-copy` mutation (attached `TASK-260922-cww1ov_review-mutants.py`, third
   mutation): it must now FAIL the row.
2. **Bookkeeping**: the ledger recount (you measured 37 all-platform rows with Windows
   host-capability tolerance + 5 all-platform/no-tolerance + 4 Unix-only = 46), the
   `.github/ci/platform-cases.tsv` ENOTDIR comment (three vs four), and a COMPLETE results resource
   (the previous one ended in a literal truncation marker).

Everything you accepted at revision 1 stands — the two 0017 hazards, both dangling states, the
migration end-to-end rows, the refusal classes with the codex admission table, and
`resolve --repair` never migrating silently. Do not re-open them; judge the delta and the claims it
makes.

Also check: the allowance for legitimate lock/journal metadata is narrow (no path allowlist
that could hide a real copy), no accepted test was weakened, and the new scan cannot pass vacuously
(e.g. plant the credential bytes under the manager home yourself and watch the row fail).

Record exactly one verdict: `accept_cr(TASK-260922-cww1ov, revision=2, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction for each finding. Do not write into the control root's LOGBOOK.md.
