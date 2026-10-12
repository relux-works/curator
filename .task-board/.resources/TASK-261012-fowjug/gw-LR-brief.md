# THE ONLY CURRENT INSTRUCTION — TASK-261012-fowjug: spec-gap wave, launcher docs hygiene (W-LR, developer)

Repository: curator-agent-launcher.

**Do this:**
1. A board outcome file was committed into the repository as `diagrams/TASK-261011-1pvp8b_figures.md`. Turn it into `diagrams/README.md`, an index of the figures in `diagrams/`: figure → spec section → phase.
2. Remove board-specific wording such as task IDs and outcome-handoff text.
3. Update any link to the old path in `SPEC.md`, `README.md` or the docs.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README and PlantUML diagrams. Use conservative PlantUML syntax; there is no renderer on this host.
- Do not run tests or builds yourself; the configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Do not invent wire fields, refusal codes or values. Where a format is unpublished, name it as unpublished and add or extend an open-question entry.
- Update every figure that shows a changed behaviour, so no contradicting copy remains.
- Add a revision note at the top of the spec in the existing style.

**Hand-off:** attach `TASK-261012-fowjug_gaps.md` (one paragraph), then `task-board handoff TASK-261012-fowjug --role developer` and END YOUR TURN.
