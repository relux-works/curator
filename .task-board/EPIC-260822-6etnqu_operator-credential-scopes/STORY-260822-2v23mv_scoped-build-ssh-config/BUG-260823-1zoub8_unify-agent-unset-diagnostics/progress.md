## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(2))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Both paths that reach the agent-unset condition emit one diagnostic
- [x] The unified message carries the protocol error code and keeps the operator remedy
- [x] The HTTPS sibling surface checked for the same split and aligned
- [x] Coverage pins the single message for both paths; go test green
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Tree selection: work from a branch cut from origin/main; the primary checkout and the .temp story worktrees hold pre-landing copies and must not be used as a source. The condition is: an operator asked for the SSH agent but SSH_AUTH_SOCK is unset. It surfaces through two different paths — the run-wide credential capture and a build_ssh scope selection — with two different messages, and only one carries a protocol error code a caller can act on. Unify them on one diagnostic that names the condition once and stays machine-readable, and keep the remedy text that tells the operator what to do. This surface now has an HTTPS sibling landed on 2026-08-25 (internal/config/buildhttps.go, internal/install/buildhttps.go): check whether the same split exists there and align both rather than fixing one. Add coverage that pins the single message for both paths. The orchestrator owns the commit and the pull request.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260826-415a10, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260826-415a10)
Developer implementation: unified run-wide and scoped SSH agent-unset failures on exact build_repository_ssh_credential_missing diagnostic; added exact-equality coverage through resolveBuildSSH. HTTPS sibling on origin/main has no analogous duplicated condition (anonymous run-wide absence versus distinct configured-source failures). Validation: targeted regression 0, internal/install 0, full go test rerun 0 after initializing the pinned worktree submodule, golangci-lint 0, build 0, gofmt/diff checks 0. Evidence: BUG-260823-1zoub8_results.md. Existing LOGBOOK.md already records both the agent-diagnostic split and fresh-worktree submodule setup anomaly.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260826-415a10, pid=88660, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260826-f98118, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260826-f98118)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260826-f98118, pid=28197, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260823-1zoub8_spawn-log_-implementer--developer--codex-_RUN-260826-415a10.log](file://BUG-260823-1zoub8/BUG-260823-1zoub8_spawn-log_-implementer--developer--codex-_RUN-260826-415a10.log) — System spawn log captured by task-board
- [BUG-260823-1zoub8_results.md](file://BUG-260823-1zoub8/BUG-260823-1zoub8_results.md) — Developer implementation and validation evidence
- [BUG-260823-1zoub8_change-request_rev1.patch](file://BUG-260823-1zoub8/BUG-260823-1zoub8_change-request_rev1.patch) — Change Request CR-BUG-260823-1zoub8-1 revision 1 candidate patch (repository_delta=present, 2 changed paths)
- [BUG-260823-1zoub8_spawn-log_-reviewer--reviewer--codex-_RUN-260826-f98118.log](file://BUG-260823-1zoub8/BUG-260823-1zoub8_spawn-log_-reviewer--reviewer--codex-_RUN-260826-f98118.log) — System spawn log captured by task-board
- [BUG-260823-1zoub8_review-verdict.md](file://BUG-260823-1zoub8/BUG-260823-1zoub8_review-verdict.md) — Reviewer acceptance verdict and validation evidence for CR revision 1

## Created
2026-08-23T01:05:22Z

## Last Update
2026-08-26T21:19:01Z

## Assigned To
[reviewer] reviewer (codex)
