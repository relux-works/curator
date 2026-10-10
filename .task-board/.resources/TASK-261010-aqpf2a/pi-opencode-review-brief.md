# THE ONLY CURRENT INSTRUCTION — review the research Change Request of TASK-261010-aqpf2a (reviewer, read-only)

Review Change Request rev1 of TASK-261010-aqpf2a: the outcome `TASK-261010-aqpf2a_pi-opencode-tool-lockdown.md`, also committed as `.research/261010_pi-opencode-tool-lockdown.md`. It researches whether Pi and OpenCode can run with no locally acting tools (curator-spec CIP-0008 lockdown, PR #134). The research brief is the precondition `pi-opencode-lockdown-brief.md`.

Check:
1. **Campaign rule: producers never edit `LOGBOOK.md`.** The rev1 patch modifies `LOGBOOK.md`. That alone requires changes: the Change Request must contain only `.research/261010_pi-opencode-tool-lockdown.md`.
2. Content against the brief: a verdict table first, then sections A–F answered; the upstream projects identified with evidence.
3. Sample at least 8 cited claims across A–E, weighted towards the removal mechanisms (§B) and the proposed configurations (§C2, §C4): the cited upstream file at the cited commit (read-only `gh api`) says what the report claims. Name every citation that does not.
4. The report claims no execution on this host: no harness installation or run, no test suite, no login.
5. No secrets, personal paths, host names or employer names in the report or the patch.

Do not run builds, tests or harnesses on this host.

Verdict: request changes. If items 2–5 hold (minor citation drift is P2 and does not block under tb-R226), the ONLY required change is item 1, stated exactly as "drop the LOGBOOK.md hunk; keep the .research file byte-identical", and say explicitly in the verdict that the content is otherwise accepted, so the next review only has to confirm that delta. Otherwise list every blocking defect as well. In the verdict findings block, `severity` must be one of `bypass|regression|robustness|note`; put the P-level in `severity_reason`. Then END YOUR TURN.
