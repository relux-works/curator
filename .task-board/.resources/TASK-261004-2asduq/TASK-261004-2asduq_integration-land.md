# Integration preconditions

Task: TASK-261004-2asduq — launch-command-environment-fragment-design. Accepted change request revision 1. This is pre-landing evidence, not a landing success claim.

The binding integration assignment supersedes earlier manual integration and generic lifecycle instructions. No status mutation, handoff, checkpoint or integrate command was executed. The runner owns synchronous landing after this run exits. No repository files were changed by this run.

Checks executed directly:
- task-board q get projection: exit 0; status integrating. An initial unsupported task query exited 1; corrected to get, which exited 0.
- task-board spawn status: exit 0; current run RUN-261005-6a2bd3 running, researcher/analyst.
- task-board worktree status: exit 0; revision 1 accepted, repository_delta=present, two changed paths; current run holds the Story lease. Worktree exists on its managed branch, base main, tip 54bed271b7609bf206a04369202473c430d0d96a. Reports dirty worktree and held lease; these were not bypassed. Also reports one unrelated uncommitted board activity file, zero unpushed commits and zero unpublished closures.
- git status --short: exit 0; only the two expected research documents are untracked.
- git diff --exit-code: exit 0; no tracked unstaged changes.
- git diff --cached --exit-code: exit 0; no staged changes.
- shasum -a 256 on both research documents: exit 0.

Document SHA-256: 26209305bc16504c040e3c8fabae5d698bb4ede91df3f140e7c77a1852741d17 (.research/261004_CIP-0002-project-context-in-managed-launches.md).
Evidence SHA-256: a83845b01a832483eb953fa1c7638f7505d31cbe5ceaf1fdd3b89dff4d8b36a2 (.research/261004_CIP-0002-project-context-in-managed-launches_evidence.md).

Acceptance is taken from the authoritative CLI status, not re-established by this run. No builds, tests or content review were rerun in this integration-only assignment. Candidate equality, fresh protected-authority checks and landing remain the runner transaction responsibility. No credentials were read and LOGBOOK.md was untouched.