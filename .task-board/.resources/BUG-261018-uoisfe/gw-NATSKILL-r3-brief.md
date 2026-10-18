# THE ONLY CURRENT INSTRUCTION — BUG-261018-uoisfe: rework round 3 (developer)

The deciding review of CR revision 1 requested changes. Read the attached verdict `BUG-261018-uoisfe_review-verdict-rev1.md`. Eleven of twelve rows held; only row 12 is broken.

**Do this.** Fix the single finding `unrelated-takeover-argument-discard`: `--takeover` must be accepted by exactly the five named operations (`profile install --use`, `profile use`, `profile sync`, `profile update`, `env resolve --repair`) and refused as a usage error (exit 2) by every other command, `curator list --takeover` included. Add a behavioral test that sweeps every registered command that is not one of the five and asserts exit 2 with the same command passing without the flag, plus a narrowing mutant showing the test fails when one command silently accepts the flag. Stay inside the bounds of `gw-NATSKILL-r2-brief.md`; every rule there still holds. Do not reopen accepted rows. Do not edit the root `README.md`.

Re-run the full local suite and the hosted gate on the exact new signed candidate.

**Hand-off:**
1. Update `BUG-261018-uoisfe_gaps.md` with a "Rework round 3" section: the finding and how it was fixed, with file and line.
2. Run `task-board handoff BUG-261018-uoisfe --role developer`.
3. END YOUR TURN.
