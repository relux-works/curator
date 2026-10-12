# THE ONLY CURRENT INSTRUCTION — TASK-261012-32vhet: apply the accepted cycle-2 review findings (W-LR4, developer)

Repository: `curator-agent-launcher` (`SPEC.md`, `diagrams/`, README). The preconditions are `gw4-findings.md`, which holds the finding texts, and `gw4-decisions.md`, which holds the owner decisions and the shared names. For each finding, read its "Affected specs" line and apply the part for this repository.

**Do this:**
1. **F-16 (owner decision):** in the workspace materialization contract, a child in another account gets a standalone repository: private `.git`, common directory, refs, index, configuration and hooks.
   - Never inherit a `.git` backlink, executable local configuration or credential helpers.
   - Copy/reflink optimisations keep this rule.
   - Optional object sharing is immutable and read-only.
   - Linked worktrees stay allowed only within one account.
   - Who transfers source and results across homes stays an open question.
2. **F-13:** in §10.4, reference the single termination scope.
3. **F-18:** reference the user manager's fate table where §10 describes process lifetime.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README, CIP text and PlantUML diagrams. Use conservative PlantUML syntax; this host has no renderer.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Apply the owner decisions in `gw4-decisions.md` exactly, and use the names fixed by its resolutions R-SESSION, R-REBIND, R-FATE, R-OUTCOME, R-TIME, R-LOCK, R-WORKPLACE, R-SVC-EGRESS, R-NETBIND and R-GEN (and the wave-3 names) unchanged. Do not invent other wire fields, refusal codes or values. Where a format is unpublished, say so and add or extend an open-question entry ("would need …").
- Remove superseded text; do not leave it as an alternative. Update every figure that shows a changed behaviour, so no contradicting copy remains. Add a revision note at the top of the spec in its existing style.
- If a finding's evidence no longer matches the current text (line numbers may have moved), apply the finding to the current text. If it is already closed, report that with section numbers.

**Hand-off:** attach `TASK-261012-32vhet_gaps.md`. In it, give for each finding you touched the changed sections and figures, and list what stays open. Then run `task-board handoff TASK-261012-32vhet --role developer` and END YOUR TURN.
