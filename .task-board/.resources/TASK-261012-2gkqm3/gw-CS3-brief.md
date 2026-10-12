# THE ONLY CURRENT INSTRUCTION — TASK-261012-2gkqm3: apply the accepted review findings (W-CS3, developer)

Repository: `curator-spec` (`cips/`, `protocol/`, `cips/diagrams/`). The preconditions are `gw3-findings.md`, which holds the finding texts, and `gw3-decisions.md`, which holds the owner decisions and the shared names. The CIPs stay Draft: do not change their status.

**Do this:**
1. **F-7.** In CIP-0009 D5/D6, its security and test sections, `protocol/remote-worker.md` and `cips/diagrams/remote-worker-turn.puml`, separate remote effect claims from the ten-minute result cache:
   - a durable claim of the binding-generation-scoped operation identity plus the canonical-argument digest, made before mutation;
   - a known claim without a result stays pending or `unknown_result` and is never re-executed automatically; mismatching arguments are refused;
   - authenticated new-versus-retry admission with a bounded sequence or epoch floor;
   - claims are discarded only when their identities can never be admitted again;
   - claims are reconciled before recovery opens admission.

   The wire stays unpublished. Add the qualification cases. Reuse the R-OBS observation states where they apply; do not add a service.
2. **CIP-0011.** Align its phase, client and acceptance text with:
   - F-5: dispatcher socket service from the first slice, no command mode;
   - F-3: requester-signed orchestrator delegation;
   - F-8: the `revocation.check` verdict;
   - F-11: the single vendor-refresh writer as a slice-1 condition;
   - F-1: delete at retirement where the CIP mentions archiving.
3. **F-4.** Remove launch-ticket text anywhere in the CIPs and the cross-CIP contract table.

**Rules:**
- Public repository: no first names, host names, local absolute paths, session links or board-resource links.
- Docs only: spec text, README, CIP text and PlantUML diagrams. Use conservative PlantUML syntax; this host has no renderer.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.
- Apply the owner decisions in `gw3-decisions.md` exactly, and use the names fixed by its resolutions R-OBS, R-VERDICT, R-START, R-DELEG, R-MAINT and R-FD3 unchanged. Do not invent other wire fields, refusal codes or values. Where a format is unpublished, say so and add or extend an open-question entry ("would need …").
- Remove superseded text; do not leave it as an alternative. Update every figure that shows a changed behaviour, so no contradicting copy remains. Add a revision note at the top of the spec in its existing style.
- If a finding's evidence no longer matches the current text (line numbers may have moved), apply the finding to the current text. If it is already closed, report that with section numbers.

**Hand-off:** attach `TASK-261012-2gkqm3_gaps.md`. In it, give for each finding you touched the changed sections and figures, and list what stays open. Then run `task-board handoff TASK-261012-2gkqm3 --role developer` and END YOUR TURN.
