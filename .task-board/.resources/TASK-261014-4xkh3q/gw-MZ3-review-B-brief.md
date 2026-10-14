# THE ONLY CURRENT INSTRUCTION — reviewer B of two (R138, cross-provider), the deciding review of TASK-261014-4xkh3q (W-MZ3)

**Sources:** the precondition `gw-MZ3-brief.md`; curator-spec eb1f08c environments §§7.1, 7.2, 7.9, 10.2, 12, 13; this repository.

**Check:**
1. Each of the brief's points 1-5 is implemented per the spec, with no muse special-casing in callers.
2. Muse monolithic output, the `referenced` refusal, drift and the status note are covered by tests that would fail on regression (reason about mutants).
3. The flag is never emitted without the materialized target; if argv lives elsewhere, the hand-off is precise.
4. The Muse root file starts with the generation header (precondition `MZ3-gate-note.md`), with a test asserting it.
5. No live-home access in code or tests; the remote-gate result is cited; signed commits.

Write your verdict before reading `gw-MZ3-review-A.md`. Accept (`accept_cr`) if no P0/P1 remains; otherwise request changes. END YOUR TURN.
