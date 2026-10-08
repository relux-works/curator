# THE ONLY CURRENT INSTRUCTION — curator v0.15.0-rc.5 release notes (developer, docs)
The operator approved releasing the B changes right away (2026-10-08). The orchestrator tags main right after this lands. Write the consolidated `## v0.15.0-rc.5 — 2026-10-08` section of CHANGELOG.md.
**Source of truth:** `git log v0.15.0-rc.4..origin/main`. Read it in full; board-record commits ("Record … board state") carry no product change.
**Do:**
1. Turn `## Unreleased` into the rc.5 section and leave a fresh empty `## Unreleased` above it. Group the entries as Added / Changed / Fixed / Security / Known issues, in the style of the rc.4 section.
2. Reconcile against the history. Expected content, verify each against its commits:
   - **Security:** the security-posture revision B flip: the hardened posture is now the DEFAULT (TASK-260927-25hk87). State what changes for an operator who relied on the previous default and how to keep the old behaviour, exactly as the code and docs describe it (read `docs/` and the test names in the commit; do not invent flags).
   - **Changed:** the Codex seed revision B flip (TASK-260927-1e5qqm): managed Codex homes no longer inherit native MCP servers; the profile MCP table is the only source. State the operator-visible effect and any migration note the commit's docs carry.
   - Anything else in the range (CI gate tweaks, conformance-gap rows, docs) under the matching heading, or omitted if not operator-visible.
   Never claim anything that is not in the history.
3. **Known issues**, stated plainly:
   - the Windows broker real-Git flake (TASK-260930-fp8vx7) is still a documented risk;
   - the board-close gap for the Go qualification task is a tooling issue only (spm#537); no product impact.
4. Change ONLY CHANGELOG.md. Never edit LOGBOOK.md. No version constants (GoReleaser ldflags set the version).
**Evidence:** in the outcome, list the commit subjects you reconciled against and map each entry to its commits.
Then `task-board handoff  --role developer` and END YOUR TURN.
