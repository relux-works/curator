# Rework 1 — TASK-260910-5nrmtt (transport resolution executor)

Verdict rev1: CHANGES_REQUESTED (resource TASK-260910-5nrmtt_review-verdict-rev1.md; reproducers in TASK-260910-5nrmtt_review-reproducers.patch — apply them as regressions, do not weaken them). Three P1s, all must be fixed in the same candidate:

1. Resolved acquisition bypasses strict admission (transport.go: acquireNetworkFormat called directly). Route AcquireNetworkResolved through the same admission checks as AcquireNetwork (trusted git, canonical source, broker/wrapper presence, SSH credential selection, locked-commit + ref validation) with the shared timeout; bind the selected SSH credentials into the real wrapper policy per attempt. Keep the reviewer's 4 refusal cases as tests at the production entry point.
2. Substring classification authorizes forbidden fallback. Replace with positive identification of transport/authentication diagnostics; mixed, ambiguous, integrity/audit and truncated stderr evidence must stop after one fetch. Keep the reviewer's 3 cases.
3. Total deadline does not bound the child process tree. Bound cancellation and pipe draining for the whole process graph (process group / Setpgid + WaitDelay or equivalent, platform-guarded), and test the non-exec child shape with the reviewer's fixture (2 s budget must return well under 5 s).

Rules unchanged: narrow package tests only (internal/buildrepo and the packages you touch), evidence with exit codes, tick checklist, handoff rev2 from the Story worktree (STORY-260910-1bhj0g). This is the final leaf of the Story; keep the checkpointed policy leaf intact.
