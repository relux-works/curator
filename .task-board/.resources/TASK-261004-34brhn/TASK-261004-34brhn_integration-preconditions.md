# Integration preconditions

Task: TASK-261004-34brhn — research-claude-login-transfer-modes.
Parent: STORY-261004-1pwxri — claude-macos-managed-home-login-modes.
Date: 2026-10-05. Bound run: RUN-261005-38ed02; researcher / analyst.

The latest Integration Assignment supersedes earlier status, handoff and manual integration commands. No status mutation, generic handoff, checkpoint or integrate was invoked. No repository file was changed by this run.

Fresh observations:
- Scoped board query exited 0: task and story both integrating.
- activity(TASK-261004-34brhn, kinds=change_request, limit=6, order=descending) exited 0: CR-TASK-261004-34brhn-2 revision 2 transitioned ready to accepted at 2026-10-05T05:18:36.109866Z; latest returned CR event.
- Tracked spawn status exited 0: current run running, researcher / analyst.
- git status --short exited 0: exactly two untracked research documents, no other changes. git rev-parse HEAD exited 0: ca1b776fb580ec0cee0173bf150daf063023aeaa.
- Standalone git diff --exit-code exited 0. Standalone git diff --cached --exit-code exited 0. Standalone git diff --exit-code -- LOGBOOK.md exited 0.
- git hash-object on .research/261004_CIP-0003-claude-managed-home-credential-modes.md and its _evidence.md companion exited 0: a81c141ed2ded73e8ab886a21f20c798a8ad37c4 and 0f90d54fc0a54772a96450b500db73f4ad5c29ac respectively. These identify observed files; no equality with the accepted snapshot is claimed.

Bounds: prior research validation and review remain historical; no research probes or Go tests were rerun. No credentials were accessed. Protected authority freshness, candidate equality, final-leaf delivery kind and transaction admission remain runner-owned checks, not proven by this packet. No landing success is claimed.

CLI discovery anomalies: task-board change-request --help and task-board cr --help each exited 1 (unknown command); neither changed state. Skill/reference and worktree/resource help reads exited 0.

Producer precondition inspection is handed to the synchronous bound runner for authoritative validation and landing. Keep integrating until that transaction resolves.