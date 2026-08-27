## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260825-1tgpcn

## Blocks
- TASK-260825-1d0eo5

## Checklist
- [x] Broker answers only the two Git prompts, only for the pinned host; everything else exits silently
- [x] State file carries host and username only; secret rides only in the fetch children's environment
- [x] GIT_ASKPASS and core.askPass both point at the wrapper
- [x] Anonymous HTTPS path unchanged; private fetch verified end to end against a real repository
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Reference material (read-only, shape and pitfalls only): a sibling manager ships this surface at /Users/iv/Developer/intranet/cocoaskills on main — src/csk/https_broker.py is the answer function and its fail-closed boundary, src/csk/git_admission.py holds broker materialization and the GIT_ASKPASS plus core.askPass wiring. Do NOT copy code verbatim into Go and do NOT name that project in any comment, commit, document or board artifact. Pitfalls already paid for: core.askPass overrides GIT_ASKPASS so both must be set; the fetch environment has an empty PATH and a private HOME, so everything the broker needs must be resolved by the manager beforehand; an HTTPS URL needs its .git suffix or the service answers 301 and the fetch refuses redirects; the state file must carry the pinned host and username and never a secret. Consume internal/gitcred rather than reimplementing credential access.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260824-ba9b31, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260824-ba9b31)
Implemented manager-binary HTTPS askpass wrapper, secret-free host/username state, per-fetch-only secret environment, and install-to-fetch credential binding. Real TLS Basic Auth Git repository test passed. Gates: focused broker/install tests 0; full go test ./... 0; full golangci-lint 0; native and Windows amd64 builds 0; scoped diff-check 0. Initial lifecycle transition was dependency-gated, then estimate-gated; dependency is now done and task entered development normally. Evidence attached in TASK-260825-3n4bjj_results.md and TASK-260825-3n4bjj_go-test-full-01.log; important process-graph decision recorded in LOGBOOK.md 0056.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260824-ba9b31, pid=56437, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-6cea80, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-6cea80)
REVIEW VERDICT: ACCEPTED (RUN-260824-6cea80). All five ACs verified with independent evidence: real-TLS Basic-Auth E2E green; foreign host/prompt/state fail closed (unit matrix + reviewer probes of the real built binary incl. symlinked/relative/unknown-field state — all silent exit 1); anonymous path broker-free and prompt-refusing; secret redacted across %v/%+v/%#v and nested GitTool; independent full go test 42/42 exit 0, golangci-lint 0 issues, vet/gofmt/no-broad-suppression clean. Reviewed behavior change recorded in LOGBOOK 0316: productionGitTool no longer honors ambient GIT_ASKPASS as askpass source — deliberate, load-bearing for the broker boundary and Spec core 12.2. Evidence: TASK-260825-3n4bjj_review-verdict.md, TASK-260825-3n4bjj_review-go-test-full-01.log. No commit_ack supplied; commit-owning mover is TASK-260825-1d0eo5.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-6cea80, pid=34714, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260825-3n4bjj_spawn-log_-implementer--developer--codex-_RUN-260824-ba9b31.log](file://TASK-260825-3n4bjj/TASK-260825-3n4bjj_spawn-log_-implementer--developer--codex-_RUN-260824-ba9b31.log) — System spawn log captured by task-board
- [TASK-260825-3n4bjj_results.md](file://TASK-260825-3n4bjj/TASK-260825-3n4bjj_results.md) — HTTPS broker implementation, fail-closed behavior, real-repository authentication, and validation evidence
- [TASK-260825-3n4bjj_go-test-full-01.log](file://TASK-260825-3n4bjj/TASK-260825-3n4bjj_go-test-full-01.log) — Full go test ./... output; command exited 0
- [TASK-260825-3n4bjj_spawn-log_-reviewer--reviewer--claude-_RUN-260824-6cea80.log](file://TASK-260825-3n4bjj/TASK-260825-3n4bjj_spawn-log_-reviewer--reviewer--claude-_RUN-260824-6cea80.log) — System spawn log captured by task-board
- [TASK-260825-3n4bjj_review-verdict.md](file://TASK-260825-3n4bjj/TASK-260825-3n4bjj_review-verdict.md) — Reviewer verdict ACCEPTED: AC-by-AC evidence, independent gate reruns, real-binary broker probes
- [TASK-260825-3n4bjj_review-go-test-full-01.log](file://TASK-260825-3n4bjj/TASK-260825-3n4bjj_review-go-test-full-01.log) — Reviewer's independent full go test run; 42/42 ok, exit 0

## Created
2026-08-24T21:23:14Z

## Last Update
2026-08-24T23:17:54Z

## Assigned To
[reviewer] reviewer (claude)
