# THE ONLY CURRENT INSTRUCTION — TASK-261012-2uuu3e: apply the accepted cycle-2 review findings (W-CS4, developer)

Repository: `curator-spec` (`cips/`, `protocol/`, `cips/diagrams/`). The preconditions are `gw4-findings.md`, which holds the finding texts, and `gw4-decisions.md`, which holds the owner decisions and the shared names. For each finding, read its "Affected specs" line and apply the part for this repository. The CIPs stay Draft: do not change their status.

**Do this:**
1. **F-19 / R-WORKPLACE:**
   - In CIP-0009 D4.7, D5, D12 and its test plan, run the project-side executor `swarma-rb-exec` under a service principal and specify the `workplace` account class, with every control outside the worker's UID.
   - Add `swarma-rb/1` and `swarma-workplace-exec/1` to the publication register in `protocol/remote-worker.md` as unpublished.
   - Update the affected figures.
2. **F-15:** the dispatcher service readiness in CIP-0011: control and observation stay reachable while forward admission is closed.
3. **F-27 / R-NETBIND:** in CIP-0011 §6, the child's network profile is bound by machine-owned policy pinned at admission; the caller's text is not authority. The network-profiles lane publishes this binding as decision **N15** of `curator-network-profiles` (spec `spec/network-profiles.md`, under review, not yet landed): a root-owned binding keyed by Curator profile name and lock digest, which replaces CIP-0011 §6's `[bindings.profiles]` route. Cite N15 as the owner of the binding and mark that route as replaced; do not restate N15's details.
4. **F-14, F-21, F-24:** align any CIP-0011 text on renewal, time and service egress with the resolutions.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README, CIP text and PlantUML diagrams. Use conservative PlantUML syntax; this host has no renderer.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Apply the owner decisions in `gw4-decisions.md` exactly, and use the names fixed by its resolutions R-SESSION, R-REBIND, R-FATE, R-OUTCOME, R-TIME, R-LOCK, R-WORKPLACE, R-SVC-EGRESS, R-NETBIND and R-GEN (and the wave-3 names) unchanged. Do not invent other wire fields, refusal codes or values. Where a format is unpublished, say so and add or extend an open-question entry ("would need …").
- Remove superseded text; do not leave it as an alternative. Update every figure that shows a changed behaviour, so no contradicting copy remains. Add a revision note at the top of the spec in its existing style.
- If a finding's evidence no longer matches the current text (line numbers may have moved), apply the finding to the current text. If it is already closed, report that with section numbers.

**Hand-off:** attach `TASK-261012-2uuu3e_gaps.md`. In it, give for each finding you touched the changed sections and figures, and list what stays open. Then run `task-board handoff TASK-261012-2uuu3e --role developer` and END YOUR TURN.
