# Review note — TASK-261001-3fgu9f rev1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest CR (base d0920353, 21 paths, gate green) against `launcher-muse-brief.md`, including its binding scope-reduction decision.

Known issues to rule on:
1. **LOGBOOK.md is a NEW file in the CR.** Producers never edit or create LOGBOOK.md. This alone means changes requested: remove it from the CR. Facts worth keeping belong in the task results or CHANGELOG.
2. **The CR appears to raise the agents-management pin to v0.5.33.** Two Claude goldens now carry `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION=false`.
   - Rule whether the pin bump is necessary for the muse rows. If it is, the golden changes are an expected, documented side effect: one CHANGELOG line, and the evidence states why.
   - If it is not necessary, revert the bump.
3. **Windows vet "fails on baseline and candidate".** Check that the candidate introduces no new Windows vet failures by diffing the two outputs.
4. **Scope.** Check each item:
   - the muse mapping row (ax provider "muse");
   - the v3 reader, with v1/v2 unchanged and HOME never set;
   - the mutants: HOME set → fail; v3 rejected → fail. Kill at least one yourself, with real exit codes;
   - the interactive refusal is a tested bound with the exact error, and it flips when the pinned plugin declares interactive;
   - the `permission_mode_unsupported` refusal is also stated as a bound.
5. **Hygiene.** No other stray files. Never spell any employer name.

Verdict: accept_cr, or changes requested with file:line.
