## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260825-168m7o

## Blocks
- TASK-260825-3kb532
- TASK-260825-1d0eo5

## Checklist
- [x] add, login, list, remove implemented with validation failures covered
- [x] No token accepted as a command-line argument
- [x] Help documents precedence and the exposure warning
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
Reference material (read-only, shape and pitfalls only): a sibling manager ships this surface at /Users/iv/Developer/intranet/cocoaskills on main — installer.py holds the per-repository resolution, the host-bound override and the candidate prompt; cli.py holds the command; tests/test_build_https.py holds the test shapes. Do NOT copy code verbatim into Go and do NOT name that project in any comment, commit, document or board artifact: this repository's artifacts reference the Curator Protocol spec and this repository only. Pitfalls already paid for: an identity-unbound override discloses the token to every host in the closure, which core 12.2 now forbids without a host binding; a this-run-only prompt answer must not reach the saved config; anonymous HTTPS must keep working when nothing matches. The credential access layer landed in internal/gitcred (TASK-260825-1tgpcn) — consume it, do not reimplement it.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260824-46b375, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260824-46b375)
Implemented curator config build-https add/login/list/remove in cmd/curator/main.go, plus config.SetBuildHTTPS/RemoveBuildHTTPS in internal/config/write.go (mirroring the existing build_ssh Set/Remove pair). Built on top of the already-present (uncommitted) sibling work from TASK-260825-168m7o (build_https config field) and TASK-260825-1tgpcn (gitcred read/write via git credential) that were in the working tree. add selects exactly one source via --git-credentials/--keyring/--token-env (no literal token flag exists); login reads a token via hidden term.ReadPassword prompt when attached to a terminal, else one line from stdin, stores it through gitcred.Access.StoreScoped, and selects token=keyring for the scope; list shows source/username plus a live present=true/false probe (ReadScoped/ReadHost/os.Getenv depending on source); remove deletes the config scope and, only for a keyring-sourced scope, the stored token too (never touches an operator-owned git-credentials entry). Help text (buildHTTPSUsage) documents precedence against CURATOR_BUILD_HTTPS_TOKEN/_HOST and the core Spec section 12.2 disclosure warning for the identity-unbound override.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-46b375, pid=72917, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-2f4cf5, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-2f4cf5)
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-2f4cf5, pid=56395, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-09ae2e, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-09ae2e)
REVIEW VERDICT (RUN-260824-09ae2e): ACCEPTED. All AC verified first-person: add/login/list/remove with full invalid-invocation matrices and a real isolated credential-store round trip; no flag on the build-https surface carries a literal token (login reads hidden prompt or stdin; config layer rejects a pasted secret); help text checked against curator-spec core.md section 12.2 directly - host binding via CURATOR_BUILD_HTTPS_HOST plus the exposure warning, ordering test-pinned. Gates rerun by reviewer: go build 0, go vet 0, gofmt clean, golangci-lint 0 issues, go test internal/config+internal/gitcred ok, go test ./cmd/curator/... -timeout 30m ok (518.6s). Full evidence in TASK-260825-2gyhq8_review-verdict.md. Minor non-blocking notes recorded there (post-prompt username length validation edge in login; broker-time-only username charset rejection; one test probes the real credential helper read-only). Out-of-scope flag for orchestrator: LOGBOOK.md entry 0130 (TASK-260825-168m7o) names the external reference project path, which epic policy forbids in document artifacts - scrub before commit. Reviewer run supplies no commit_ack; commit-owning mover owns the scope commit. Prior reviewer run RUN-260824-2f4cf5 ended before its background test runs finished, hence no verdict from it.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-09ae2e, pid=34629, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260825-2gyhq8_spawn-log_-implementer--developer--claude-_RUN-260824-46b375.log](file://TASK-260825-2gyhq8/TASK-260825-2gyhq8_spawn-log_-implementer--developer--claude-_RUN-260824-46b375.log) — System spawn log captured by task-board
- [TASK-260825-2gyhq8_results.md](file://TASK-260825-2gyhq8/TASK-260825-2gyhq8_results.md) — Implementation notes and validation evidence for curator config build-https
- [TASK-260825-2gyhq8_spawn-log_-reviewer--reviewer--claude-_RUN-260824-2f4cf5.log](file://TASK-260825-2gyhq8/TASK-260825-2gyhq8_spawn-log_-reviewer--reviewer--claude-_RUN-260824-2f4cf5.log) — System spawn log captured by task-board
- [TASK-260825-2gyhq8_spawn-log_-reviewer--reviewer--claude-_RUN-260824-09ae2e.log](file://TASK-260825-2gyhq8/TASK-260825-2gyhq8_spawn-log_-reviewer--reviewer--claude-_RUN-260824-09ae2e.log) — System spawn log captured by task-board
- [TASK-260825-2gyhq8_review-verdict.md](file://TASK-260825-2gyhq8/TASK-260825-2gyhq8_review-verdict.md) — Reviewer verdict: accepted. AC verification, first-person gate results, minor non-blocking observations.

## Created
2026-08-24T21:23:40Z

## Last Update
2026-08-24T23:14:18Z

## Assigned To
[reviewer] reviewer (claude)
