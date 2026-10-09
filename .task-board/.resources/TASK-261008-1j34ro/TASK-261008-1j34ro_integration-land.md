Integration producer preflight for N8, revision 4. Run RUN-261009-30f2c0.

Latest Integration Assignment supersedes earlier direct-integration and generic status/handoff instructions. Landing is reserved to the runner after this producer exits. No integration or checkpoint command was invoked.

Observed: task and Story status integrating; CR-TASK-261008-1j34ro-4 accepted, kind story_final, producer role developer, archetype implementer. Current workspace lease belongs to this run. HEAD and recorded CR base both f2ed88a615f648f4c2fd8beca6a1c4d3a89c5ffc. Accepted candidate tree 36d74898f16ed9dda196b4e12e6760676d0abcf7.

Working-tree content comparison against accepted candidate: 6/6 changed files MATCH (CHANGELOG.md, cmd/curator/audit_allow_test.go, cmd/curator/main.go, internal/audit/audit.go, internal/audit/pin_digest_test.go, internal/hashing/hashing.go). git status lists exactly these six paths. Comparison command exited 0. No source files changed by this run.

Read checks: worktree status, task/Story status query, acceptance activity query, spawn status and directives all exited 0. Initial query using unsupported changeRequest projection exited 1; corrected projection succeeded. No directives recorded.

No build, tests, validation gates, status mutation or generic handoff executed, as required by this integration assignment. Existing revision-4 acceptance and evidence retained. Protected-authority freshness and transactional landing remain runner checks; this preflight does not claim landing success.