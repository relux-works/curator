## Status
done

## Assigned To
[reviewer] reviewer (codex)

## Created
2026-07-20T02:11:48Z

## Last Update
2026-07-22T02:12:10Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Lock identities and acquisition order are deterministic
- [x] Unix and Windows subprocess contention tests pass
- [x] Dry-run proves no lock state is created
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Protocol candidate handoff: TASK-260720-3ag6pi is blocked only on a real landed rc.4 release commit, which cannot be created under this goal. Its independent reviewer confirmed all candidate validation, integrity, compatibility, regeneration, and safety evidence. Consume /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree explicitly as the candidate conformance root (suite SHA sha256:70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae). This does not constitute release or pin evidence; see epic precondition EPIC-260720-21aq1i_protocol-candidate-handoff.md.
Curator integration precondition: create /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zl1cj/worktree from exact origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 and import only the complete reviewer-accepted product diff from /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-4bd0it/worktree. Exclude .temp, board/config, planning/research, diagrams, binaries, caches, alternate indexes, and unrelated files; do not import blocked TASK-260720-1zntv0-only changes. Do not commit or stage. Own only internal/managerlock with build-tagged Unix/Windows implementations and tests; do not implement journals, target swaps, installer orchestration, or GC. Enforce canonical unsigned-UTF-8 project ordering, optional key-before-home constraints, OS-release semantics, subprocess contention/abnormal-exit coverage, and a dry-run API path that creates no lock state. Record exact provenance, native Unix runtime, Windows compile/runtime availability, race/vet/build and no-staging evidence honestly.
spawn queued: [implementer] developer (codex) (run=RUN-260721-3a6c79, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260721-3a6c79)
Logbook 2026-07-21 — exact origin/main 17804cea worktree imported the byte-identical reviewer-accepted TASK-260720-4bd0it product diff (tracked binary-diff SHA-256 ed23bdd5db679114bcdb4ea53134f84d602d0bbf6c80820aef61f653fc66a342), then added only internal/managerlock. Decision: stable hashed lock files live below manager home and are never deleted; an operation state machine plus process-local reservations enforces canonical project order, at most one optional key released before home, and home-only recovery/GC. Native Unix subprocess contention, independent-project/key concurrency, cancellation, abnormal-exit, 82.1% focused race coverage, make check, full race, native build, Linux/Windows compile, scoped lint (0 issues), gofmt, diff, provenance, and no-staging gates pass. Native Windows runtime is unavailable on Darwin. Full lint exposes 64 inherited issues outside managerlock in the accepted imported diff; those ownership-external files were not changed. Evidence: TASK-260720-1zl1cj_results.md and attached logs.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-3a6c79, pid=63205, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260721-b31526, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260721-b31526)
Review changes requested: canonicalAbsolute falls back to an unresolved spelling for a missing home. Once the first lock acquisition creates that home below a symlinked ancestor, the same configured path resolves differently; homeLockPath hashes the changed spelling, so independent managers can hold distinct files instead of one exclusive manager-home lock, and process order state is split. See TASK-260720-1zl1cj_review-verdict.md for reproduction, required prefix-canonicalization rework, regression tests, and passing gates.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-b31526, pid=78097, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260721-059184, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260721-059184)
Logbook 2026-07-21 review rework — fixed unstable first-use manager-home identity by resolving the longest existing prefix before appending nonexistent components. Regression proves a pre-creation manager and post-creation subprocess contend through an aliased ancestor; portable coverage verifies stable identity without symlink privileges. Full race/check/build/cross-compile gates pass; no staging. Native Windows runtime and golangci-lint executable unavailable in this Darwin run; Windows test binary compilation and Go vet/gofmt passed.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-059184, pid=83014, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260721-cb0754, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260721-cb0754)
Review cycle 2 changes requested: the canonical-prefix fix appends missing components verbatim, so case-variant configured homes created from an absent path on case-insensitive Windows retain different Manager.home values. homeLockPath hashes those strings and stateForHome keys on them, allowing distinct lock files and split in-process ordering for one physical manager home. Add a native Windows case-alias first-use subprocess regression and stabilize identity/exclusion without breaking case-sensitive filesystems. All available Darwin race/check/vet/gofmt/diff/no-staging gates and Windows/Linux compilation pass; native Windows runtime unavailable. Evidence: TASK-260720-1zl1cj_review-verdict-cycle-2.md
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-cb0754, pid=88486, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260721-9b6bd9, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260721-9b6bd9)
Logbook 2026-07-21 review rework cycle 3 — resolved Windows first-use case-alias identity splitting. Windows canonicalization now queries per-directory case sensitivity: ordinary case-insensitive paths normalize the complete existing-plus-missing identity, while case-sensitive directories preserve spelling. Added a Windows-only pre-creation alias regression covering shared Home/lock path/process state, key-before-home enforcement, subprocess contention, release, and post-creation stability. Focused race coverage is 82.4%; make check, full uncached race, native build, Linux/Windows test compilation, vet, gofmt, diff, and no-staging gates pass. Native Windows runtime is unavailable on this Darwin host; golangci-lint is not installed, while task-defined go vet/gofmt lint gates pass. Evidence: TASK-260720-1zl1cj_review-rework-cycle-3.md
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-9b6bd9, pid=94222, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260721-05beec, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260721-05beec)
Review cycle 3 changes requested: identity_windows.go applies one leaf directory case-sensitivity flag to the entire canonical path. Windows semantics are per-directory, so a case-sensitive parent can contain distinct Foo/foo child homes whose own flags are case-insensitive; both are uppercased to one identity, sharing process state/lock root and potentially redirecting state to a third spelling. The existing Windows regression skips case-sensitive roots. Rework requires component-wise or equivalent stable physical canonicalization plus native Windows regressions for mixed parent/child flags and multi-component first use. Darwin focused race coverage 82.4%, full race, make check/build, vet/gofmt/diff/no-staging, and Linux/Windows compilation pass; native Windows runtime unavailable. Evidence: TASK-260720-1zl1cj_review-verdict-cycle-3.md
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-05beec, pid=2713, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260721-d3ab7f, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260721-d3ab7f)
Logbook 2026-07-21 review rework cycle 4 — replaced leaf-wide Windows case normalization with containing-directory component semantics and acquisition-time physical identity stabilization before process reservation or lock-path selection. This avoids Foo/foo collisions under sensitive parents, prevents uppercase sibling redirection, and safely handles multi-component first use with observed created-directory flags. Added mixed-flag native Windows regressions plus a concurrent first-use race regression. Focused race coverage is 82.8 percent; make check, full uncached race, native build, Linux/Windows test and CLI compilation, vet, gofmt, diff, and no-staging gates pass. Native Windows runtime remains unavailable on Darwin with no configured runner/emulator; attached the exact Windows test executable. golangci-lint is absent. Evidence: TASK-260720-1zl1cj_review-rework-cycle-4.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-d3ab7f, pid=8677, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260721-b0afb8, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260721-b0afb8)
Logbook 2026-07-21 review cycle 4 — no remaining code or architecture defect found. Independent Darwin focused race coverage 82.8 percent, full race, make check, native build, Windows vet and test compilation, Linux test compilation, scoped golangci-lint 0 issues, gofmt, diff, and no-staging gates pass. Acceptance is stop-the-line blocked solely on the explicit native Windows subprocess test requirement: no runnable Windows environment exists locally, and existing windows-latest CI requires an unauthorized commit/push or PR. Recommended input is approval for a temporary task branch/PR, or a native Windows runner with go test -race -cover ./internal/managerlock -count=1 -v and go test ./... logs. Evidence: TASK-260720-1zl1cj_review-verdict-cycle-4.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-b0afb8, pid=18764, exit=0)
Native Windows runner available 2026-07-22 — SSH alias win reaches admin@100.120.84.42 (DESKTOP-3PBO632, Windows 10.0.19045). Port 22 and native execution are available; Go/Git are absent from remote PATH. Use the attached exact TASK-260720-1zl1cj_managerlock-windows-amd64.test.exe or independently cross-compile the candidate test binary locally, transfer with scp, execute native Windows subprocess/coverage-compatible tests, preserve exact provenance, remove remote temporary artifacts, and attach a distinct native-runtime verdict. Do not commit or stage.
spawn queued: [reviewer] reviewer (codex) (run=RUN-260722-89cde4, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260722-89cde4)
Logbook 2026-07-22 review cycle 5 — changes requested. Native Windows 10.0.19045 executed the exact current candidate binary (SHA-256 e76fb84cf807bcbfb671ad1fd51e4c90837b9d56e237f8f74bf6490f74b43e92) and returned FAIL: TestMissingHomeUsesCanonicalExistingPrefix and TestMissingHomeBelowSymlinkKeepsIdentityAndContends assert Unix-style preserved casing instead of the package Windows canonical identity. Windows case-alias and subprocess contention/abnormal-exit coverage passed; two mixed per-directory case-sensitivity tests skipped because the runner reports the feature unsupported. Darwin focused race coverage 82.8%, full race, make check, build, vet, cross-compilation, gofmt, diff, and no-staging gates pass. Remote artifacts removed and cleanup verified. Rework portable expectations without weakening Windows canonicalization, then rerun the full native Windows suite and reviewer cycle. Evidence: TASK-260720-1zl1cj_review-verdict-cycle-5.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260722-89cde4, pid=20901, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260722-e45bf6, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260722-e45bf6)
Logbook 2026-07-22 review rework cycle 6 — corrected the two portable canonical-prefix test expectations to use platform canonical identity without changing product behavior. Native Windows 10.0.19045 complete managerlock suite now exits 0 using binary SHA-256 cff6319218783241284f1ea237b4abdb052b5bd33b347297d1e2f10e204c896f; all available tests pass, with two per-directory case-sensitivity tests skipped because the runner reports unsupported. Darwin focused race coverage 82.8%, full race, make check, build, vet, formatting, diff, Linux/Windows compilation, no-staging, and remote cleanup gates pass. golangci-lint is unavailable. Evidence: TASK-260720-1zl1cj_review-rework-cycle-6.md and TASK-260720-1zl1cj_native-windows-cycle-6.log.
Cycle-6 native runtime evidence is linked as TASK-260720-1zl1cj_native-windows-cycle-6-verified.log; the earlier unlinked filename was superseded during resource attachment recovery.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260722-e45bf6, pid=28331, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260722-99ec17, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260722-99ec17)
Logbook 2026-07-22 review cycle 6 — accepted. Independent inspection found no remaining correctness, scope, or architecture defect. Fresh native Windows binary SHA-256 cfd9ef67ee1186df08286f95a068d75893052162e65ce6270a8c58d00eabd428 passed the complete package suite at 82.5 percent coverage; two unsupported per-directory case-sensitivity tests skipped, while Windows case-alias and all subprocess contention/abnormal-exit paths passed. Darwin focused race coverage 82.8 percent, full race, make check, build, Windows vet/test compilation, Linux test compilation, formatting/diff, and no-staging gates pass. Initial PowerShell flag parsing ran no tests; corrected cmd.exe rerun passed. Remote cleanup verified. Evidence: TASK-260720-1zl1cj_review-verdict-cycle-6.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260722-99ec17, pid=38797, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-1zl1cj_spawn-log_-implementer--developer--codex-.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_spawn-log_-implementer--developer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-1zl1cj_results.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_results.md) — Implementation provenance, lock behavior, platform coverage, validation, and lint-boundary evidence
- [TASK-260720-1zl1cj_managerlock-race.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_managerlock-race.log) — Native subprocess, abnormal-exit, race, and coverage validation log
- [TASK-260720-1zl1cj_make-check.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_make-check.log) — Successful full go vet, test, and formatting validation log
- [TASK-260720-1zl1cj_full-race.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_full-race.log) — Successful uncached repository-wide race test log
- [TASK-260720-1zl1cj_full-lint-anomaly.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_full-lint-anomaly.log) — Inherited full-repository lint findings outside managerlock ownership; scoped package lint is clean
- [TASK-260720-1zl1cj_spawn-log_-reviewer--reviewer--codex-.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_spawn-log_-reviewer--reviewer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-1zl1cj_review-verdict.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-verdict.md) — Independent review verdict, canonical-home defect evidence, required regression, and passing validation
- [TASK-260720-1zl1cj_review-rework-results.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-rework-results.md) — Canonical-prefix review rework and validation evidence
- [TASK-260720-1zl1cj_review-verdict-cycle-2.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-verdict-cycle-2.md) — Second-cycle independent review verdict, remaining Windows first-use canonicalization defect, required regression, and passing validation
- [TASK-260720-1zl1cj_review-rework-cycle-3.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-rework-cycle-3.md) — Third-cycle Windows case-alias canonicalization rework and validation evidence
- [TASK-260720-1zl1cj_review-verdict-cycle-3.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-verdict-cycle-3.md) — Third-cycle independent review verdict, remaining per-directory Windows canonicalization defect, regression requirements, and passing validation
- [TASK-260720-1zl1cj_review-rework-cycle-4.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-rework-cycle-4.md) — Cycle-4 Windows component-wise identity rework, regression coverage, validation, and platform boundary
- [TASK-260720-1zl1cj_managerlock-windows-amd64.test.exe](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_managerlock-windows-amd64.test.exe) — Cycle-6 Windows amd64 managerlock test executable; SHA-256 cff6319218783241284f1ea237b4abdb052b5bd33b347297d1e2f10e204c896f
- [TASK-260720-1zl1cj_review-verdict-cycle-4.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-verdict-cycle-4.md) — Cycle-4 review verdict, passing gates, and native Windows execution stop-line evidence
- [TASK-260720-1zl1cj_review-verdict-cycle-5.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-verdict-cycle-5.md) — Cycle-5 native Windows review verdict, exact failing tests, passing gates, and required rework
- [TASK-260720-1zl1cj_review-rework-cycle-6.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-rework-cycle-6.md) — Cycle-6 portable expectation rework and complete native Windows validation
- [TASK-260720-1zl1cj_native-windows-cycle-6-verified.log](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_native-windows-cycle-6-verified.log) — Verified native Windows cycle-6 verbose package test, SHA-256 provenance, and cleanup evidence
- [TASK-260720-1zl1cj_review-verdict-cycle-6.md](file://TASK-260720-1zl1cj/TASK-260720-1zl1cj_review-verdict-cycle-6.md) — Cycle-6 independent accepted review with native Windows and full validation evidence
