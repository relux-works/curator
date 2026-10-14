# THE ONLY CURRENT INSTRUCTION — reviewer A of two (R138), RECORD-ONLY second opinion, for TASK-261014-2f0qnv (PL2). Reviewer B has already decided

**Sources:** the precondition `gw-PL2-brief.md`; curator-spec main `6648e18` environments §§1.3, 10.1, 10.4, 12, 13, manager §12.5, `conformance/environments-repair-preserve-lock/vectors.json`; this repository.

**Check:**
1. The refusal is decided before `beginOperation` and before any write, staging or store rebuild; no fragment; exact diagnostic and operator command per §10.1.
2. `--preserve-lock` without `--repair` is the existing usage error, raised before resolution.
3. The lock-changed race re-check exists under the repair lock and refuses with `environment_repair_failed`; the lock hash is otherwise unchanged; the current-home path takes no lock; plain `--repair` behaviour is unchanged.
4. Every vector case maps to a test that would fail on regression (reason about mutants: late refusal, dropped re-check, rebuild-before-refusal).
5. No live-home access; the remote-gate result is cited; no private identifiers or local paths.

RECORD-ONLY: attach `gw-PL2-review-A.md`. Do NOT call accept_cr or reject_cr. END YOUR TURN.
