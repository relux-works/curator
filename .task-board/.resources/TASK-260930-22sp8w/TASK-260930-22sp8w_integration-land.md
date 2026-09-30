# TASK-260930-22sp8w — integration precondition confirmation (rev3, producer run RUN-260930-6211a0)

Revision 3 is ACCEPTED. Per the bound-developer integration binding, the producer
does NOT run `worktree integrate` or `worktree checkpoint`, does NOT set status,
and does NOT call the generic `handoff`. The runner performs the bound landing
synchronously after this run exits. This artifact confirms landing preconditions only.

## Preconditions (observed 2026-09-30, worktree)

- Board status: `integrating` (verified via `task-board q 'get(TASK-260930-22sp8w) { id status }'`).
- Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260930-12oimr/worktree`
- Branch: `task-board/story/STORY-260930-12oimr`
- HEAD: `b4b08a1993d240b5dd24d6929cafe2b51e72637c` (rev3 base; no producer commit on top — delta left UNCOMMITTED as required for handoff snapshot).
- Uncommitted delta vs HEAD: 35 files, 600 insertions, 6 deletions (`git diff --stat HEAD -- . ':!.task-board'`).
- Spawn directives for RUN-260930-6211a0: none.
- No file changed by this run; no integrate/checkpoint/handoff/status executed by the producer.

## Uncommitted path list (35)

- .github/ci/gate-selftest.sh
- .github/ci/test-gate.sh
- cmd/curator/gitenv_hostile_test.go
- cmd/curator/main_test.go
- cmd/curator/status_test.go
- internal/audit/gitenv_main_test.go
- internal/buildrepo/httpsbroker_test.go
- internal/closure/gitenv_main_test.go
- internal/config/gitenv_main_test.go
- internal/contextpkg/gitenv_main_test.go
- internal/crossconformance/gitenv_main_test.go
- internal/devsub/gitenv_main_test.go
- internal/envmarker/gitenv_main_test.go
- internal/envprofile/network_fixture_test.go
- internal/gitcred/gitcred_test.go
- internal/gitignore/gitenv_main_test.go
- internal/gitops/gitenv_main_test.go
- internal/godriver/main_test.go
- internal/identity/gitenv_main_test.go
- internal/install/atomicity/gitenv_main_test.go
- internal/install/scriptpolicy_test.go
- internal/interop/environments/gitenv_main_test.go
- internal/manifest/gitenv_main_test.go
- internal/marker/gitenv_main_test.go
- internal/rustsource/main_test.go
- internal/scriptpolicy/gitenv_main_test.go
- internal/scriptworker/main_test.go
- internal/skillspec/gitenv_main_test.go
- internal/snapshot/gitenv_main_test.go
- internal/sourcelock/gitenv_main_test.go
- internal/swiftpmbuild/gitenv_main_test.go
- internal/swiftpminterop/gitenv_main_test.go
- internal/swiftpmsource/gitenv_main_test.go
- internal/testgitenv/testgitenv.go
- internal/yarnmodernsource/gitenv_main_test.go

## For the runner

Ready for the bound synchronous landing:
`task-board worktree integrate STORY-260930-12oimr --cr TASK-260930-22sp8w --revision 3`
using this run's immutable role/archetype and revision binding. The runner records
the landing evidence.
