## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- TASK-260825-1lausy
- TASK-260825-2gyhq8

## Blocks
- TASK-260825-2fy132
- TASK-260825-1d0eo5

## Checklist
- [x] Candidate prompt with narrowest-scope default; discovery only lists, never uses
- [x] A this-run-only answer never reaches the saved config, on both credential surfaces
- [x] Non-terminal run continues anonymously for HTTPS
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
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Reference material (read-only, shape and pitfalls only): a sibling manager ships this surface at /Users/iv/Developer/intranet/cocoaskills on main — installer.py holds the candidate prompt and the persistence rule; tests/test_build_https.py and tests/test_build_ssh.py hold the test shapes, including the regression that a this-run-only answer must never reach the saved config. Do NOT copy code verbatim into Go and do NOT name that project in any comment, commit, document or board artifact. Consume the surfaces this epic already landed in the primary checkout: internal/gitcred for credential access, internal/config/buildhttps.go for scopes, internal/install/buildhttps.go for resolution. The producers of this epic work directly in the primary checkout /Users/iv/Developer/ReluxWorks/curator with unstaged changes; keep doing that so the landing task can assemble everything at once, and do not commit. Pitfall already paid for on the sibling: the prompt accumulates rules in memory to avoid re-prompting, so saving that accumulator writes run-only answers to disk; save only the entries the operator chose to persist, and check the SSH prompt path for the same latent bug.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260824-6f6eb7, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260824-6f6eb7)
agent completed: [implementer] developer (codex) (exit=124)
spawn run completed: codex (run=RUN-260824-6f6eb7, pid=95775, exit=124)
spawn run RUN-260824-6f6eb7 failed; operator action required; failure: run exceeded --timeout 50m0s and was terminated by the launcher
Finalize only: the previous run delivered the implementation and its tests into the primary checkout and wrote TASK-260825-3kb532_results.md, then the launcher terminated it at the timeout before it could finish the handoff. Do not redo the work. The orchestrator has run the gates on the combined tree: go build ./... clean and go test ./... -count=1 with 42 packages ok, exit 0, attached as TASK-260825-3kb532_orchestrator-full-suite.log. Verify the delivery matches the acceptance criteria, run gofmt and go vet and the linter if they are quick, tick the checklist, and hand off to review.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260825-e716f7, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260825-e716f7)
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260825-e716f7, pid=85755, exit=0)
No Change Request revision was published for TASK-260825-3kb532 (handoff_unsatisfied): the board is not at to-review
Producer delivery is complete: implementation and tests are in the primary checkout (internal/install/buildhttps.go and its test file, plus the external.go and CLI wiring), documented in TASK-260825-3kb532_results.md. Two consecutive runs were terminated by the launcher while waiting on their own re-run of the suite, so the orchestrator finalized the mechanical handoff on evidence it ran itself: go build clean, gofmt clean, go vet clean, go test ./... -count=1 with 42 packages ok and exit 0 (TASK-260825-3kb532_orchestrator-full-suite.log). Reviewer: the suite is already attached — do not re-run the whole suite to reach a verdict; targeted package runs are enough if you need them.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260825-18497d, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260825-18497d)
Review verdict: ACCEPTED (RUN-260825-18497d). All four AC verified with evidence in TASK-260825-3kb532_review-verdict.md: prompt default/this-run-only/abort tested on both surfaces; run-only never reaches config or credential store (byte-for-byte, real persist callback); the SSH unconditional-persist bug was real and is fixed via the shared save-gated scope question with no accumulator-wide save anywhere; gates rerun in the Story worktree (build, gofmt, vet, golangci-lint, 4 internal packages, targeted cmd/curator) all green on top of the attached orchestrator full-suite log. Three non-blocking notes recorded in the verdict (untested keyring rollback branch, untested scope-must-cover re-ask, run-only fixed at namespace scope). Commit/integration is the orchestrator step.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260825-18497d, pid=92013, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260825-3kb532_spawn-log_-implementer--developer--codex-_RUN-260824-6f6eb7.log](file://TASK-260825-3kb532/TASK-260825-3kb532_spawn-log_-implementer--developer--codex-_RUN-260824-6f6eb7.log) — System spawn log captured by task-board
- [TASK-260825-3kb532_results.md](file://TASK-260825-3kb532/TASK-260825-3kb532_results.md) — Implementation summary, negative tests, workspace anomaly, and validation exit codes
- [TASK-260825-3kb532_orchestrator-full-suite.log](file://TASK-260825-3kb532/TASK-260825-3kb532_orchestrator-full-suite.log)
- [TASK-260825-3kb532_spawn-log_-implementer--developer--claude-_RUN-260825-e716f7.log](file://TASK-260825-3kb532/TASK-260825-3kb532_spawn-log_-implementer--developer--claude-_RUN-260825-e716f7.log) — System spawn log captured by task-board
- [TASK-260825-3kb532_spawn-log_-reviewer--reviewer--claude-_RUN-260825-18497d.log](file://TASK-260825-3kb532/TASK-260825-3kb532_spawn-log_-reviewer--reviewer--claude-_RUN-260825-18497d.log) — System spawn log captured by task-board
- [TASK-260825-3kb532_review-verdict.md](file://TASK-260825-3kb532/TASK-260825-3kb532_review-verdict.md) — Reviewer verdict: accepted; AC-by-AC evidence, reran targeted gates in the Story worktree

## Created
2026-08-24T21:23:40Z

## Last Update
2026-08-25T00:48:16Z

## Assigned To
[reviewer] reviewer (claude)
