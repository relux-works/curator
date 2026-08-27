## Status
done

## Assigned To
[reviewer] reviewer (codex)

## Created
2026-07-20T02:11:47Z

## Last Update
2026-07-20T23:42:44Z

## Blocked By
- (none)

## Blocks
- TASK-260720-3mrm4z
- TASK-260720-1zntv0
- TASK-260720-4bd0it
- TASK-260720-3itlly

## Checklist
- [x] Build-source digest matches authoritative framing vectors
- [x] Frozen snapshots reject mutation and invalid file types
- [x] Tampered snapshot-cache entries rebuild without changing legacy hashes
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
Curator implementation worktree precondition: create /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-256kj1/worktree from exact origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8. Do not copy the dirty shared checkout or any .planning, .research, .task-board, diagrams, task-board.config.json, .temp, binaries, caches, alternate indexes, or unrelated user files. Use /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree only as the explicit candidate conformance root; do not copy it into Curator or call it a released/pinned spec. Do not commit or stage. Record exact base, changed files, vector identities consumed, focused/full tests, race/vet results applicable to scope, and preservation of legacy hashing.
spawn queued: [implementer] developer (codex) (run=RUN-260720-efb4fc, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260720-efb4fc)
Implementation checkpoint: working in the required detached worktree at origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8. Added internal/buildsource validation/token API with exact curator-build-source-v1 framing, marker inclusion, portable path and platform-collision checks, link/special-file rejection, directory-aware mutation rechecks, and callback ordering guards. Snapshot cache now validates a freshly archived immutable commit before inspecting/reusing a cache hit and atomically stages replacements for corrupt/incomplete entries. Candidate conformance root consumed explicitly at suite SHA sha256:70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae. Focused conformance tests and focused race tests pass; full validation remains in progress.
Logbook 2026-07-21 — immutable build-source implementation: origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 required a new retained-root validation token rather than reusing legacy ContentSHA256, because the protocol identity includes root .csk-install.json and length framing while mutation checks must also include directory structure. Snapshot cache presence is no longer trusted: every lookup first validates a fresh git archive of the immutable commit, then reuses only an equivalent cached tree or stages a replacement. Candidate-focused buildsource/snapshot/hashing and race tests pass; baseline make check, build, Windows compile, and diff checks pass. An entire-repository run with candidate conformance enabled exposed only an expected downstream internal/interop manager-lifecycle reader gap on origin/main; all other packages passed, and the normal full suite passes. golangci-lint is unavailable, so lint evidence is go vet plus repository gofmt via make check. Full details are attached as TASK-260720-256kj1_results.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-efb4fc, pid=36927, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260720-bfb79e, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260720-bfb79e)
Reviewer verdict 2026-07-21 — changes requested. Build-source framing, marker separation, mutation handling, focused race tests, make check, and Windows compile gates pass. Stop-ship defect: repair of a present tampered snapshot is not atomic. The target-to-backup then staged-to-target sequence exposes a missing live path and races concurrent repairers. Public-API stress evidence over 100 rounds with 12 concurrent snapshot.Get calls: 19 caller errors and 7198 missing-target observations, including ENOENT, directory-not-empty, failed restore, and validation of a missing publication. Existing concurrency coverage tests only a cold miss. Full evidence and rework acceptance are attached as TASK-260720-256kj1_review-verdict.md. Route: to-dev, not blocked.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-bfb79e, pid=53196, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260720-b10bdc, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260720-b10bdc)
Rework logbook 2026-07-21 — replaced the reviewed target-to-backup repair with cross-process OS locking and portable immutable-generation publication. Cold snapshots retain the historical path; repair validates and publishes a sibling generation, atomically switches a bounded regular-file pointer, retains the immediately retired generation for existing readers, and prunes older generations best-effort. This avoids the Windows API constraint that existing non-empty directories cannot be atomically replaced. The pre-fix regression reproduced missing-path and caller failures; the rework passes 100 rounds with 24 concurrent mixed callers, focused race tests, candidate buildsource/snapshot/hashing conformance, make check, build, Linux/Windows compile, diff, and legacy-hashing preservation gates. Evidence updated in TASK-260720-256kj1_results.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn completion blocked: no new task-scoped outcome artifact was attached. Add an outcome resource named like TASK-260720-256kj1_results.md and then set status back to to-review.
spawn run completed: codex (run=RUN-260720-b10bdc, pid=61627, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260720-ebe648, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260720-ebe648)
Revalidation logbook 2026-07-21 — independently reran the existing atomic-repair rework after the prior spawn hook rejected handoff for lack of a new outcome artifact. Candidate-focused conformance, 100-round 24-caller tampered-cache stress, focused race, coverage, make check, native build, Windows/Linux compile, diff/gofmt, and legacy-hashing preservation gates all pass. No caller errors or missing-live-path observations occurred. Fresh evidence attached as TASK-260720-256kj1_atomic-repair-revalidation.md; no new defect or forced-fit constraint found.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-ebe648, pid=85251, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260720-8eac6d, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260720-8eac6d)
Reviewer logbook 2026-07-21 — accepted after second-cycle rework. Independently verified exact curator-build-source-v1 framing and marker separation, frozen-token mutation rejection, immutable-source-first cache validation, and the replacement generation publication design. The prior present-repair race is closed: 100 rounds with 24 mixed callers passed with zero caller errors and zero missing historical live-path observations; focused race, candidate conformance, make check, native build, Windows/Linux compile, formatting/diff, coverage, and legacy-hashing isolation gates pass. Final evidence is attached as TASK-260720-256kj1_final-review-verdict.md. Candidate-suite evidence remains non-release and non-pin evidence.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-8eac6d, pid=94303, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-256kj1_spawn-log_-implementer--developer--codex-.log](file://TASK-260720-256kj1/TASK-260720-256kj1_spawn-log_-implementer--developer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-256kj1_results.md](file://TASK-260720-256kj1/TASK-260720-256kj1_results.md) — Implementation, atomic concurrent-repair rework, conformance vectors, and verification evidence
- [TASK-260720-256kj1_spawn-log_-reviewer--reviewer--codex-.log](file://TASK-260720-256kj1/TASK-260720-256kj1_spawn-log_-reviewer--reviewer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-256kj1_review-verdict.md](file://TASK-260720-256kj1/TASK-260720-256kj1_review-verdict.md) — Independent reviewer verdict, validation evidence, and required atomic snapshot-repair rework
- [TASK-260720-256kj1_atomic-repair-revalidation.md](file://TASK-260720-256kj1/TASK-260720-256kj1_atomic-repair-revalidation.md) — Fresh validation evidence for the reviewer-requested atomic concurrent snapshot repair
- [TASK-260720-256kj1_final-review-verdict.md](file://TASK-260720-256kj1/TASK-260720-256kj1_final-review-verdict.md) — Accepted second-cycle review verdict with independent conformance, atomic-repair, race, full-suite, coverage, build, portability, and legacy-hash evidence
