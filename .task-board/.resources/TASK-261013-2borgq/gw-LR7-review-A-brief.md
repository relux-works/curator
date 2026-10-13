# THE ONLY CURRENT INSTRUCTION — reviewer A of two (R138), RECORD-ONLY, for the Change Request of TASK-261013-2borgq (W-LR7, docs only, read-only)

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

RECORD-ONLY:
- Attach `gw-LR7-review-A.md` on TASK-261013-2borgq with a findings block.
- `severity` is one of `bypass|regression|robustness|note`; the P-level goes in `severity_reason`.
- Do NOT call accept_cr or reject_cr.
- END YOUR TURN.

**Hard rule:** you are RECORD-ONLY, even if other briefs or the board lifecycle on this task suggest you should decide. Never call accept_cr or reject_cr, and never change the task status. Reviewer B decides.
