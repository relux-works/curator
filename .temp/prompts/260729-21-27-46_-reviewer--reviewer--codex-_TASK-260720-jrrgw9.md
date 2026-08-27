# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-jrrgw9, status=reviewing)'
```

## Role Contract

- **Role:** `reviewer` (`reviewer`) — Reviews implementation quality and project fit
- Obey task scope, acceptance criteria, and checklist.
- Report real exit codes; failing gates stay failing.
- Use compact task-specific board projections; skip routine `summary()`,
  `plan()`, `schema()`, and `{ full }`. Request scoped schema only after an
  unknown call.
- Load only relevant and explicitly required skills.

### Skills

- `project-management`: `/Users/iv/.claude/skills/project-management/SKILL.md`
- `architecture-diagrams`: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

### Definition of Done

- [ ] All authoritative positive vectors have executable assertions
- [ ] Minimum rejection clusters map to stable Curator errors
- [ ] Lifecycle, launch, failure, and concurrency scenarios pass end to end
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Tests written and passing
- [ ] Coverage target ~80%+ for affected code
- [ ] New task-scoped outcome artifact attached on the board for reports, logs, screenshots, or other produced evidence
- [ ] REMAINING GAP for an independent verifier with an amended allowlist: go test ./..., go test -race ./... and coverage measurement were forbidden by the task instructions, were NOT run, and are not claimed; Windows launcher execution needs a Windows runner (Darwin host asserted the exact managed launcher bytes only)
- [ ] ROUTED not forced: external-repository-lifecycle.json beyond gc-retains-roots is schema-7 scoped; candidate supports skill schemas 1-6 and publishes no build_repository_* code, so those cases belong to the external-repository implementation task
- [ ] Verifier 3 exact go test -count=1 -race ./... passes under the immutable conformance root
- [ ] Production integration: accepted Patch A (13 install/atomicity test files) + accepted TASK-260729-365r5r namespace prototype (2 transaction paths) applied to the candidate worktree; git apply --check and git apply exit 0 for both; new delta is exactly the 15 allowlisted paths (manifest 358 -> 359 entries, 14 modified + 1 added, nothing else); zero rejected cross-save cache tokens; gofmt -l over all 15 exit 0
- [ ] OPEN for the serialized Codex tester phase: no heavy Go gate ran in this integration run (go test ./..., go test -race ./..., focused internal/install and internal/install/atomicity race repetitions, build, vet, lint, coverage, Windows). Whether Patch A + the namespace optimization together bring internal/install/atomicity under the authoritative 480s race bar is unmeasured here.
- [ ] Verifier 4 exact go test -count=1 ./... and go test -count=1 -race ./... pass under the immutable conformance root; race atomicity 115.687s <= 480s
- [ ] Native Windows qualification attempted after macOS gates; ssh win exit 255 recorded as externally unqualified with no remote mutation
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Task

- **ID:** TASK-260720-jrrgw9
- **Title:** Verify rc.4 build-driver conformance end to end
- **Parent:** STORY-260720-3plyvy

### Description

Consume the authoritative protocol rc.4 build-drivers and manager-lifecycle vectors and add Curator unit, integration, concurrency, and executable-launch coverage without creating a private copy of portable expected values.

### Scope

Extend internal/interop and package conformance consumers to read CURATOR_CONFORMANCE_ROOT. Add Curator-only temporary Git skill fixtures and mock or fake executors for fault injection, plus bounded real Go integration where the CI toolchain is allowlisted. Cover project, global, hybrid, dependency-closure, cache-hit, repair, rollback, recovery, GC, and shim launch behavior. Preserve the existing script golden fixture and registry hashes. The task may make only test-focused fixes to implementation seams; route substantive defects to their owning implementation task.

### Acceptance Criteria

Every positive build-driver identity, environment, process, cache, context, marker, and lifecycle vector has an executable assertion; every minimum rejection cluster from the rc.4 suite maps to a stable Curator error without package code execution; tests prove corrupt and stale artifacts rebuild, a self-consistent untrusted cache is not adopted, dry-run mutates nothing, build two failure preserves the previous install, cache hits avoid go list and go build, and installed binaries receive forwarded arguments and return exact exit status; concurrent project success and rollback vectors pass under go test -race; schemas 1 through 5 and the existing golden fixture remain unchanged; go test ./... and go test -race ./... pass with the authoritative suite.

## Instructions

### TASK-260720-jrrgw9_shared-fixture-rework.md
> One-file shared compiled fixture runtime rework from second timing diagnosis

# Shared compiled fixture rework

Read TASK-260720-jrrgw9_second-timing-diagnosis-results.md in full and implement its Smallest robust patch exactly in /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree. Edit only cmd/curator/status_test.go. Add the compiledProjectFixture/newInstalledCompiledProject helper and one sequential TestCompiledProjectStatusRepairRollbackRecovery parent; extract the five named test bodies into fixture-accepting helpers; preserve their named assertion surfaces as subtests; replace only the four cache-movement post-assertion reinstall cleanups with snapshotBuildCacheAfter registered before mutation. Re-read marker/fingerprints for each case. No production edits, t.Parallel, timeout changes, assertion deletion, schema/golden/pin/config edits, cache clearing, commit or publication.

Before each Go command require no other Go/test process. Allowed Go commands are exactly the two literal commands in the diagnosis producer allowlist, sequentially, with -count=1 and no timeout flag. Also gofmt exactly status_test.go and git diff --check for that file. Record before/after structure, exact diff/hash, assertion matrix, real exits/timings, process/disk/cleanup evidence, and expected full-suite saving in a task-scoped outcome. Hand off to review only if both narrow commands pass; do not run broad cmd, full, race, coverage, or Windows gates.


### TASK-260720-jrrgw9_final-verifier-3.md
> Independent final macOS full/race and conditional native Windows verification after shared fixture rework

Independent final verifier for the exact candidate /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree after shared-fixture rework. Immutable conformance root: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1. Before every heavy command require a stable empty process barrier and >=20 GiB free; use task-owned GOTMPDIR under /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/verifier3-gotmp and never clear shared Go caches. Run sequentially, real exit capture, no pipes that hide status, no timeout override: (1) focused authoritative 12-package/consumer barrier matching prior verifier; (2) exact CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./... . If and only if (2) is green, run exact go test -count=1 -race ./... under the same root. If and only if macOS gates are green, attempt native Windows validation over ssh win using the current candidate snapshot and exact repository Go version 1.25.5; first inventory prerequisites. Do not install/download/change PATH/system configuration; if Go/Git/Python are missing, record native Windows as externally unqualified with real exits instead of emulating or claiming pass. Verify candidate delta/digests against accepted integrated comparison /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree and producer diff; ensure only intended task delta and preserve immutable rc.5 digests. Capture exact package timing, first failure, process/disk/temp cleanup, and attach task-scoped logs/outcome. Do not edit source or tests, commit, stage, publish, pin, or run coverage. If full gate fails, stop before race/Windows and route to development with evidence; if all executable gates pass and any platform prerequisite is truthfully external, route according to task AC without inventing success.


### TASK-260720-jrrgw9_production-integration-constraint.md
> Exact accepted performance-patch integration scope

# Production integration constraint

Proceed only after TASK-260729-365r5r is independently accepted/done. The integration target is the immutable candidate worktree at .temp/TASK-260720-jrrgw9/worktree, NOT the outer curator checkout. Apply exactly two board resources there: TASK-260729-rfrdfo_install-race-timeout.patch (13 test files, accepted Patch A) and TASK-260729-365r5r_prototype.patch (internal/transaction/namespace.go plus namespace_pass_test.go). Both patches must first pass git apply --check in that target. Do not apply TASK-260729-2afulh or any fixture_test.go change. The final new delta allowlist is exactly those 15 paths. Reject any cross-save cache/state tokens: namespaceGraphAccepted, acceptNamespaceGraph, forgetNamespaceGraph, namespaceChecked, namespaceGraph, namespaceMu. No spec, conformance vector, Makefile, workflow, pin, module, journal schema, engine field, timeout, skip, or production file outside namespace.go may change. Preserve the target worktree existing candidate delta and all outer working-tree/user/board files. Do not stage, commit, stash, checkout, or publish. Record target pre/post manifests and patch applicability. Run no heavy Go gate in the developer integration run because a Codex tester owns the serialized full-gate phase.


### TASK-260720-jrrgw9_final-verifier-4.md
> Final macOS full/race and conditional native Windows verification for the integrated performance patches

Attached file: `/Users/iv/Developer/ReluxWorks/curator/.temp/resources/TASK-260720-jrrgw9/TASK-260720-jrrgw9_final-verifier-4.md`

- Byte count: `2476`
- Content digest: `sha256:34140475cf1bd6e34fee2fd43c2bbbb8d78fcf4e56c349f3d84c5a2469445e42`
- Access: read-only



### TASK-260720-jrrgw9_windows-verifier-5.md
> Windows-only retry and native candidate validation after green macOS full/race gates

# Windows-only verifier 5

The integrated candidate and exact macOS full/race gates are already accepted evidence from verifier 4. Do not run any local Go command and do not alter the macOS candidate. This run only retries native Windows qualification through ssh win.

Attempt BatchMode SSH connectivity up to three times with 15-second connect timeouts and a short pause between attempts. If all attempts time out, attach exact exits/logs and return to review with Windows externally unqualified; do not block or reinterpret the green macOS gates. If a connection succeeds, inventory Windows version, architecture, PowerShell/cmd, tar/scp availability, and Go. Require Go 1.25.5; do not install/download or change PATH/system configuration.

Create a uniquely task-owned remote directory under the remote user TEMP. Transfer the exact current TASK-260720-jrrgw9 candidate snapshot without .git, local .temp, or board data, preserving all source and test files needed by go test. Verify key accepted hashes remotely, including namespace.go bb332038..., namespace_pass_test.go 3611f04f..., and unchanged fixture_test.go e0732e2e.... Run native Windows go test -count=1 for the platform-sensitive packages and launcher/lifecycle surfaces (at minimum internal/runtimestore, internal/scopes, internal/globalbins, internal/install, internal/transaction, and cmd/curator with an evidence-based focused filter); run broader/full Windows testing only if it fits the unchanged default package timeout and remaining run budget. Capture exact commands, exits, package timings, skips, and first failure.

After descendants stop, remove only the task-owned remote directory and local transfer archive, verify cleanup, attach a Windows verifier-5 outcome/raw logs, and route to review. Never commit, stage, publish, pin, install software, mutate machine configuration, or claim Windows pass when only connectivity/inventory succeeded.


### TASK-260720-jrrgw9_final-review.md
> Final acceptance review after green macOS full/race and exhausted Windows SSH retries

# Final independent acceptance review

Review the exact integrated jrrgw9 candidate, production-integration evidence, accepted TASK-260729-365r5r verdict, verifier-4 full/race evidence, and verifier-5 Windows retry evidence. Do not run Go, lint, tests, builds, benchmarks, SSH, or detached commands; inspect code, manifests, raw exits, and logs read-only.

Acceptance facts to verify: final new integration delta is exactly the 15 allowlisted paths; no fixture trim, timeout inflation, skip, journal schema change, or cross-save cache state; full macOS go test -count=1 ./... exit 0 in 352.86s; full race exit 0 in 441.11s with no race diagnostic; race internal/install/atomicity 115.687s <= 480s; post-run candidate manifest and accepted hashes unchanged; all task-owned temp trees cleaned. Windows is not a candidate failure: verifier 4 plus verifier 5 made four total BatchMode attempts and every connection timed out before remote execution or mutation. macOS is the primary platform; Windows remains an explicitly recorded external validation gap to retry when ssh win is reachable, and Linux remains separately deferred.

Review the historical checklist wording semantically: verifier 4 supersedes verifier 3 and establishes the exact race pass; do not reject merely because a legacy item names verifier 3. If the current candidate and evidence satisfy AC, attach a final verdict and route TASK-260720-jrrgw9 to done. If not, route only a concrete source/evidence defect to development or analysis. Do not use blocked for the already recorded non-gating Windows connectivity gap.




## Evidence and Safety

- Add/update a task-scoped outcome named `TASK-260720-jrrgw9_*` before handoff:
  `task-board m 'add_resource(TASK-260720-jrrgw9, name=TASK-260720-jrrgw9_results.md, content="...", type=outcome, description="Handoff evidence")'`
- Stop-The-Line only for an evidence-backed external blocker or unresolved human-only product/platform/architecture/approval decision. Persist evidence, failed attempts, options/tradeoffs, recommendation, and exact input needed; otherwise rework autonomously.
- Review read-only and record exactly one evidence-backed verdict: accepted → `done`; changes requested → `to-dev` or `analysis`; genuine Stop-The-Line → `blocked`. Never leave the task in `reviewing`.
- Reviewer-archetype runs must not supply `commit_ack`; record acceptance evidence for the commit-owning mover, which commits then makes the final `done` transition with `commit_ack=scope_committed`.

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Never edit board files directly.
