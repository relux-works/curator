# Integration preconditions — hardened security posture default
Task: TASK-260927-25hk87
Run: RUN-261008-cb3bcc
Date: 2026-10-08

Latest integration assignment followed: runner owns synchronous landing. No integrate, checkpoint, generic handoff, status mutation, or repository file edit performed.

Fresh task query: integrating (exit 0).
Fresh worktree status: CR-TASK-260927-25hk87-2 revision 2 accepted, kind story_final, producer developer/implementer; current lease RUN-261008-cb3bcc (exit 0).
Base and branch tip: ff8f75a0209928bfeae94306f9171c8c8159cb87.
Accepted candidate: af82499fd08af44beeb05f50162fa107fafa2f4b.
Observed 18 tracked modifications plus the expected untracked internal/config/security_posture_test.go; matches all 19 accepted paths.
Standalone git diff --exit-code against candidate, excluding that untracked file: exit 0.
Untracked file hash and accepted blob both 9e5ac4a33cbad47f7e20239f57e5240be38a67c0 (hash-object and rev-parse exit 0).
Standalone git diff --check: exit 0.
No delta in scripts/remote-gate.sh, LOGBOOK.md or CHANGELOG.md.
Run status and directives queries: exit 0; running developer/implementer, no directives.

Accepted existing evidence, not rerun: review-verdict-rev2.md and review-hosted-rev2.json (both resource reads exit 0). Hosted gate https://github.com/relux-works/curator/actions/runs/37293008806 records success, 20 executed jobs green and 2 declared skips for exactly the accepted candidate tree. Reviewer config and CLI targeted suites each recorded exit 0; posture counts config 12 driven/5 bound/0 gaps/0 skips and CLI 13 driven/4 bound/0 gaps/0 skips; all 6 B vectors driven. Existing review confirms warning release prerequisite and unchanged CodexSeedRevision.
No new code, tests or configuration changes; build/test/lint not rerun by this integration producer.
Bound: fresh protected-trunk authority, delivery freshness and transactional landing eligibility must be rechecked by the runner; recorded workspace authority is historical, not a fresh remote proof. Landing has not been performed or claimed here.
