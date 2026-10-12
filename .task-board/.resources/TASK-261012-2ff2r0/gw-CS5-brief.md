# THE ONLY CURRENT INSTRUCTION — TASK-261012-2ff2r0: apply the accepted finding F-23, operator custody (W-CS5, developer)

Repository: `curator-spec` (`cips/`). The preconditions are `gw4b-f23.md` (the finding), `gw4b-decisions.md` (the owner decision with its phone-first operator path, and R-OPERATOR) and `gw4-decisions.md` (earlier shared names). The owner decision takes precedence over the finding text. The CIPs stay Draft: do not change their status.

**Do this:**
1. In CIP-0011, align the operator and owner principals, the `curator broker` command surface (enrol, grant, revoke, status) and the slice-0 acceptance with R-OPERATOR:
   - widening acts need authentication plus confirmation and are refused from an agent session;
   - reads go through `swarma-observer`;
   - the phone path is the main path and the terminal sudo path the fallback.
2. Update the cross-CIP contract table if it lists operator principals.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README, CIP text and PlantUML diagrams. Use conservative PlantUML syntax; this host has no renderer.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Apply the owner decisions in `gw4b-decisions.md` and `gw4-decisions.md` exactly, and use the names fixed by its resolutions R-OPERATOR (and the wave-3 and wave-4 names) unchanged. Do not invent other wire fields, refusal codes or values. Where a format is unpublished, say so and add or extend an open-question entry ("would need …").
- Remove superseded text; do not leave it as an alternative. Update every figure that shows a changed behaviour, so no contradicting copy remains. Add a revision note at the top of the spec in its existing style.
- If a finding's evidence no longer matches the current text (line numbers may have moved), apply the finding to the current text. If it is already closed, report that with section numbers.

**Hand-off:** attach `TASK-261012-2ff2r0_gaps.md`. In it, give for each finding you touched the changed sections and figures, and list what stays open. Then run `task-board handoff TASK-261012-2ff2r0 --role developer` and END YOUR TURN.
