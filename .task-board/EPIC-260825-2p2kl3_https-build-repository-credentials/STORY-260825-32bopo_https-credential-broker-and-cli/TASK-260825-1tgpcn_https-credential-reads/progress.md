## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- TASK-260825-3n4bjj
- TASK-260825-1d0eo5

## Checklist
- [x] Every read and write goes through git credential with interactive prompting disabled
- [x] Operator home pinned for the helper lookup; namespaced entry distinct from the operator's own
- [x] A write that persists nothing is caught by reading it back, with platform guidance in the message
- [x] go test green
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
Reference material for this epic (read-only, and only as a source of shape and pitfalls): a sibling manager already ships this surface at /Users/iv/Developer/intranet/cocoaskills on main — src/csk/build_https.py (scope grammar, token sources, credential reads with read-back verification), src/csk/https_broker.py (the answer function and its fail-closed boundary), src/csk/git_admission.py (broker materialization, GIT_ASKPASS and core.askPass wiring), src/csk/installer.py (per-repository resolution, host-bound override, candidate prompt), src/csk/cli.py (the command), tests/test_build_https.py (46 tests worth copying the shape of). Read it to avoid rediscovering the pitfalls; do NOT copy code verbatim into Go, and do NOT name that project in any code comment, commit message, document or board artifact — this repository's artifacts reference the Curator Protocol spec and this repository only. Known pitfalls already paid for there: platform secret tools do not work on Windows so the operator Git credential machinery is the only cross-platform mechanism; a credential helper can report a successful write while persisting nothing; an HTTPS URL in a manifest needs its .git suffix or the service answers 301 and the fetch refuses redirects; core.askPass overrides GIT_ASKPASS so both must be set; the fetch environment has an empty PATH and a private HOME, so anything the manager resolves must be resolved before the fetch.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260824-e8091d, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260824-e8091d)
Implemented in internal/gitcred (new, dependency-free package): Access.ReadHost / ReadScoped / StoreScoped / DeleteScoped / Discover, plus NamespaceUsername, ScopeHost, OperatorHome. Every call is git credential fill|approve|reject — no platform branch. Prompting is closed off four ways (GIT_TERMINAL_PROMPT=0, GCM_INTERACTIVE=never, -c credential.interactive=false, -c core.askPass= with GIT_ASKPASS/SSH_ASKPASS *removed* rather than emptied, since git reads an empty askpass var as unset and falls through). HOME and USERPROFILE are both pinned to the operator home; the rest of the environment is inherited because a helper is session-bound. Manager entries live under username curator-build-https:<scope>. Writes and deletes are both proved by reading back; failure raises with the store to configure on the running platform plus the other platforms and the env-var alternative. Two guards for non-collision: a host read refuses an answer whose username carries the namespace prefix, and a scoped read refuses an answer for a different username. Reads never error — absent material, helper failure, missing git and a 15s hang all report nothing here. 17 tests, no skips: a stand-in git (this test binary re-executed) with selectable helper defects, plus real git with its built-in store helper in a pinned temp home with system config excluded. Evidence: go test -timeout 30m ./... exit 0 (42 pkgs), go test -race ./internal/gitcred ok, go vet ./... exit 0, golangci-lint run 0 issues, gofmt clean on tracked files. Findings in LOGBOOK.md 2026-08-25 0158.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-e8091d, pid=67919, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-de8abc, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-de8abc)
REVIEW cycle 1 (RUN-260824-de8abc): CHANGES REQUESTED -> to-dev. All AC items verified and hold (prompting closed off, home pinned, read-back catches a non-persisting helper with platform guidance, namespacing collision-free; reviewer re-ran package tests + race, build, vet, golangci-lint, gofmt — all green; nothing else imports the package). One CONFIRMED defect: the documented 15s bound does not hold when the credential *helper* (git's child) hangs — exec.CommandContext kills git only, the orphaned helper keeps the inherited stderr pipe open, and with cmd.WaitDelay=0 cmd.Run() blocks until the helper exits. Repro: 500ms timeout, 20s grandchild holding stderr -> Run returned after 20.3s; with WaitDelay set -> 1.5s (.temp/TASK-260825-1tgpcn/waitdelay-repro/). TestACallIsBounded misses it because modeHang hangs the fake git itself, not a grandchild. Requested: set cmd.WaitDelay in Access.call + add a grandchild-hang harness mode asserting the bound. Nothing else — do not touch env construction, refusal guards, read-back or messages. Full verdict: resource TASK-260825-1tgpcn_review-verdict.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-de8abc, pid=60664, exit=0)
Rework: apply exactly the change the reviewer requested in TASK-260825-1tgpcn_review-verdict.md and nothing else. Set cmd.WaitDelay in Access.call so Run returns within a bound after the context kill instead of blocking on pipes an orphaned helper holds, and extend the stand-in git with a hang mode whose grandchild inherits stderr and outlives the kill, asserting the elapsed bound the way TestACallIsBounded does. The reviewer's repro lives at .temp/TASK-260825-1tgpcn/waitdelay-repro/main.go. Do not touch the environment construction, the namespacing, or the read-back verification: all were verified and accepted. Correct the 15s-bound claim in the logbook entry if the fix changes it.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260824-bbc036, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260824-bbc036)
Rework cycle 2 (RUN-260824-bbc036): applied exactly the reviewer request. internal/gitcred/gitcred.go — new drainDelay=2s constant and cmd.WaitDelay=drainDelay in Access.call, so a call costs at most Timeout+drainDelay instead of waiting on pipes an orphaned helper still holds after the context kill. internal/gitcred/gitcred_test.go — modeHangGrandchild (stand-in git spawns a helper and blocks on it) plus modeSleeper (that helper: the test binary re-executed, handed the stand-in git own stderr, sleeping 30s, recording nothing), and TestACallIsBoundedWhenTheHelperOutlivesGit asserting the elapsed bound. Regression-proved: with the WaitDelay line removed the new test fails at 30.008s; with it, 2.21s. TestACallIsBounded untouched, still 0.21s. Environment construction, refusal guards, read-back and message text untouched as instructed. Validation, each standalone: go build ./... 0; go test -count=1 ./internal/gitcred/ 0; go test -race -count=1 ./internal/gitcred/ 0 (21.3s); go test -timeout 30m -count=1 ./... 0 (42 pkgs ok); go vet ./... 0; golangci-lint run ./internal/gitcred/... 0 issues; gofmt -l internal/ cmd/ clean. NOT green and NOT mine: golangci-lint run ./... exits 1 with 5 issues in cmd/curator/main.go:2179,2181 (errcheck) and internal/install/buildhttps.go:46,63,87 (revive) — both files written at 02:16-02:17 by sibling tasks still in flight, left for their owners. LOGBOOK 0158 corrected: the 15s-bound claim was false, replaced by a FINDING with the mechanism and the measured numbers. Evidence: TASK-260825-1tgpcn_results-cycle2.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-bbc036, pid=72644, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-e5c23e, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-e5c23e)
REVIEW cycle 2 (RUN-260824-e5c23e): ACCEPTED -> done. The rework is exactly what cycle 1 requested and nothing else: drainDelay=2s + cmd.WaitDelay in Access.call (gitcred.go:62,275), modeHangGrandchild/modeSleeper harness modes and TestACallIsBoundedWhenTheHelperOutlivesGit. Reviewer reproduced the regression proof independently in a scratch copy: without the WaitDelay line the new test fails at 30.02s (call outlives the orphaned helper), with it 2.23s PASS; the standalone repro shows 20.3s vs 1.5s. All gates re-run by the reviewer: package tests 18/18, -race ok, flake screen of both bounded-call tests 5x under heavy load 10/10, build 0, vet 0, golangci-lint ./... now 0 issues tree-wide (the 5 sibling-owned issues the implementer reported were since fixed by their owners), gofmt clean, consumers (internal/install, internal/config) race-green. Full-tree go test: 41/42 ok incl cmd/curator (882.7s under 3-suite contention) and gitcred; the single FAIL is TestRegistryAttestationLandsInMarker — a pre-existing, load-sensitive registry snapshot timestamp tolerance (registry_e2e_test.go untouched, snapshot.go:159 untouched, passed 3x solo today in the same tree). Not attributable to this delivery; flagged for the orchestrator as its own fix candidate. Logbook 0223 correction verified. Acceptance evidence in TASK-260825-1tgpcn_review-verdict-cycle2.md; no commit_ack from this reviewer — the delivery (internal/gitcred/, LOGBOOK entries) awaits the commit-owning mover.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-e5c23e, pid=56298, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260825-1tgpcn_spawn-log_-implementer--developer--claude-_RUN-260824-e8091d.log](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_spawn-log_-implementer--developer--claude-_RUN-260824-e8091d.log) — System spawn log captured by task-board
- [TASK-260825-1tgpcn_results.md](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_results.md) — Operator HTTPS credential access through git credential: design decisions, observed helper behaviour, test harnesses and validation evidence
- [TASK-260825-1tgpcn_spawn-log_-reviewer--reviewer--claude-_RUN-260824-de8abc.log](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_spawn-log_-reviewer--reviewer--claude-_RUN-260824-de8abc.log) — System spawn log captured by task-board
- [TASK-260825-1tgpcn_review-verdict.md](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_review-verdict.md) — Reviewer verdict cycle 1: changes requested — unbounded hang past DefaultTimeout when the credential helper (git's child) hangs; WaitDelay repro + exact one-line fix
- [TASK-260825-1tgpcn_spawn-log_-implementer--developer--claude-_RUN-260824-bbc036.log](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_spawn-log_-implementer--developer--claude-_RUN-260824-bbc036.log) — System spawn log captured by task-board
- [TASK-260825-1tgpcn_results-cycle2.md](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_results-cycle2.md) — Rework after review cycle 1: WaitDelay bound + grandchild-hang harness, regression proof, validation exit codes
- [TASK-260825-1tgpcn_spawn-log_-reviewer--reviewer--claude-_RUN-260824-e5c23e.log](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_spawn-log_-reviewer--reviewer--claude-_RUN-260824-e5c23e.log) — System spawn log captured by task-board
- [TASK-260825-1tgpcn_review-verdict-cycle2.md](file://TASK-260825-1tgpcn/TASK-260825-1tgpcn_review-verdict-cycle2.md) — Review cycle 2 verdict: ACCEPTED. Rework verified (WaitDelay bound + grandchild-hang test), regression proof reproduced independently, all gates re-run; one unrelated pre-existing registry flake documented for the orchestrator.

## Created
2026-08-24T21:23:14Z

## Last Update
2026-08-24T22:54:46Z

## Assigned To
[reviewer] reviewer (claude)
