# THE ONLY CURRENT INSTRUCTION — reviewer B of two (R138, cross-provider), the deciding review of the Change Request of TASK-261013-2borgq (W-LR7r2, docs only, read-only)

**Revision 2.** Revision 1 failed the hosted gate only on `TestReleaseVersionPinned`: the test pinned the pre-tag README and CHANGELOG strings. The orchestrator expanded the scope by exactly that test's pinned entries (`gw-LR7r2-brief.md`). Check that the test edit is limited to those entries, that the gate is green on all four lanes, and that the docs delta still meets the brief.


**Sources:**
- the preconditions `gw-LR7-brief.md`, `gw6-decisions.md` and the attached finding records;
- `gw5-decisions.md`, `gw4b-decisions.md` and `gw4-decisions.md`;
- this repository's current text;
- the cited sections of the neighbouring module specs.

**Check:**
1. Every finding in `gw-LR7-brief.md` is applied in this module's sections, or truthfully reported as needing no change, with section numbers.
2. Each change meets the finding's requirement and stays within the brief's bound. The optional owner policies stay off. No implementation or qualification is claimed.
3. The shared names are used unchanged. No wire fields, refusal codes or values were invented. Gaps are named, with open questions.
4. The security properties survive the edit:
   - fail-closed on unknown;
   - no automatic retry after an ambiguous effect;
   - unknown is never absent;
   - narrowing acts never widen authority;
   - recovery after an emergency stop is explicit.
5. The figures match the text, no contradicting copy remains, and every new or changed figure is linked.
6. The outcome `TASK-261013-2borgq_gaps.md` is complete and true.
7. Public hygiene: no first names, host names, local absolute paths, session links, board-resource links, F-numbers or review-record slugs.
8. No collateral edits: the base-to-candidate diff touches only what the findings need. Watch for mechanical substitutions, such as `UTF-8` mangled by an F-number strip.

No tests or builds on this host.

Do checks 1–8 yourself and write your verdict BEFORE you read `gw-LR7r2-review-A.md`. Then reconcile, and check every P0/P1 that A raised.

Accept (`accept_cr`) if no P0/P1 remains; P2 issues are notes under tb-R226. Otherwise request changes, listing the exact defects. END YOUR TURN.
