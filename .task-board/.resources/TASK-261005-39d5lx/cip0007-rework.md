# THE ONLY CURRENT INSTRUCTION — TASK-261005-39d5lx rework 1 (researcher; curator-spec)
The reviewer (sol high, RUN-261005-fa0397) requested changes with six numbered findings: three High and three Medium. Read the attached `TASK-261005-39d5lx_review-verdict-rev1.md` IN FULL and address every finding:
1. **Medium.** Remove both vendor or organisation names (CIP lines 18–19); use generic CLI descriptions. Correct the evidence note's false "none" claim with a real rescan count.
2. **High.** Separate first-match registry selection from federation-wide revocation.
3. **High.** Bind upstream verification to an operator-trusted publisher and subject, not just to "a valid signature".
4. **High.** Distinguish snapshot rollback protection from tool downgrade policy and replay authorization.
5. **Medium.** Correct which manifest the lock hash covers, and specify versioned extensions.
6. **Medium.** Make option B's runtime-exposure contract internally consistent, and preserve the execution controls.
Keep the Draft status and the template shape. Cite a spec § for every security claim, or mark the claim as a proposal. In the evidence note, add a table that maps finding → change → location.
Rerun `python -B tools/validate.py` (exit code) and the name/path scan (counts only).
Then `task-board handoff TASK-261005-39d5lx --role researcher` and END YOUR TURN.
