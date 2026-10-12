# THE ONLY CURRENT INSTRUCTION — reviewer B of two (R138, cross-provider), the deciding review of the Change Request of TASK-261012-32vhet (W-LR4, docs only, read-only)

**Sources:** the preconditions `gw-LR4-brief.md`, `gw4-decisions.md` and `gw4-findings.md`; this repository's current spec; the cited sections of the neighbouring module specs.

**Check:**
1. Every item of `gw-LR4-brief.md` is done, or truthfully reported as already closed.
2. The owner decisions F-16, F-20 and F-21 are applied exactly; the F-1, F-4 and F-5 decisions still hold; superseded text is removed.
3. The shared names from the resolutions are used unchanged. Nothing beyond them is invented: no wire fields, refusal codes or values. Gaps are named, with open questions.
4. The findings' security properties survive the edit:
   - fail-closed on unknown;
   - no automatic retry after an ambiguous effect;
   - ancestor revocation reaches descendants;
   - point-in-time verdicts are not grants.
5. The figures match the text; no contradicting copy remains.
6. The outcome `TASK-261012-32vhet_gaps.md` is complete and true.
7. Public hygiene: no first names, host names, local absolute paths, session links or board-resource links.

No tests or builds on this host. Do checks 1–7 yourself and write your verdict BEFORE you read `gw-LR4-review-A.md`. Then reconcile, and check every P0/P1 that A raised. Accept (`accept_cr`) if no P0/P1 remains; P2 issues are notes under tb-R226. Otherwise request changes, listing the exact defects. END YOUR TURN.
