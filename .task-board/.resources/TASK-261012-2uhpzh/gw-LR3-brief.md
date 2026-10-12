# THE ONLY CURRENT INSTRUCTION — TASK-261012-2uhpzh: apply the accepted review findings (W-LR3, developer)

Repository: `curator-agent-launcher` (`SPEC.md`, `diagrams/`, README). The preconditions are `gw3-findings.md`, which holds the finding texts, and `gw3-decisions.md`, which holds the owner decisions and the shared names.

**Do this:**
1. **F-6 / R-FD3.** Rewrite §10.1 so it references the user manager's single descriptor-lifecycle contract and does not restate a conflicting order:
   - only caller-owned 0–2 and read-only FD3;
   - no service, broker or root handle in the executor;
   - the staging requirement: completely staged, bounded, immutable bytes preferred; a pipe only with a published capacity, feeder, deadline and cleanup order.

   Keep the intake, digest, framing and read-deadline publication gates, and add the joint publication of the FD3 mechanism as a gate before activation. Keep the 48/80 KiB bounds.
2. **F-4 (owner decision).** Remove references to launch tickets, ticket redemption and a host role in §10 and its figures. Only the dispatcher invokes the privileged launcher.
3. **F-5.** Fix any reference to a dispatcher command mode.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README, CIP text and PlantUML diagrams. Use conservative PlantUML syntax; this host has no renderer.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Apply the owner decisions in `gw3-decisions.md` exactly, and use the names fixed by its resolutions R-OBS, R-VERDICT, R-START, R-DELEG, R-MAINT and R-FD3 unchanged. Do not invent other wire fields, refusal codes or values. Where a format is unpublished, say so and add or extend an open-question entry ("would need …").
- Remove superseded text; do not leave it as an alternative. Update every figure that shows a changed behaviour, so no contradicting copy remains. Add a revision note at the top of the spec in its existing style.
- If a finding's evidence no longer matches the current text (line numbers may have moved), apply the finding to the current text. If it is already closed, report that with section numbers.

**Hand-off:** attach `TASK-261012-2uhpzh_gaps.md`. In it, give for each finding you touched the changed sections and figures, and list what stays open. Then run `task-board handoff TASK-261012-2uhpzh --role developer` and END YOUR TURN.
