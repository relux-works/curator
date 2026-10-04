# Integration preconditions

Run: RUN-261004-422d56. No landing command was invoked: the current integration assignment delegates the synchronous transaction to the runner after producer exit. No status change or generic handoff was invoked.

Read-only evidence:
- Board query exited 0: bug and parent Story are integrating.
- worktree status exited 0: CR-BUG-261004-bknio5-1 revision 1 is accepted, kind story_final; producer role developer, archetype implementer; candidate tree 31d96981112f3a682357b8064b28ed8ba5c1d42b. Active workspace lease belongs to this run.
- Branch tip equals checkpoint 934952a45953587a1d4184b692b3fb4ee401e732.
- git status --short exited 0: only cmd/curator/gc_test.go, internal/scopes/gc.go, internal/scopes/gc_conservative_test.go are modified.
- git diff --exit-code 31d96981112f3a682357b8064b28ed8ba5c1d42b -- cmd/curator/gc_test.go internal/scopes/gc.go internal/scopes/gc_conservative_test.go exited 0: all three files match the accepted snapshot.
- git diff --check exited 0.
- Spawn directives query exited 0: no directives.

No repository files changed by this run. No local Go build, tests, vet or lint executed, per hosted-evidence restriction. Existing acceptance and hosted evidence were not rerun. Initial query requesting unsupported changeRequests field exited 1; corrected query exited 0.

Bounds: status reports unrelated uncommitted board activity for TASK-261002-2ipeqa and holder_run_record_known=false, while the active lease and spawn status identify this running producer. These were not repaired or bypassed. Fresh authority, delivery, candidate identity and transactional landing gates remain runner-owned; landing success is not asserted here.
