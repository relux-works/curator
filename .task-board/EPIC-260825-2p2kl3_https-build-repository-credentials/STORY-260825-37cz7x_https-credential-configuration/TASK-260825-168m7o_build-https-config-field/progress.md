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
- TASK-260825-1lausy
- TASK-260825-2gyhq8
- TASK-260825-1d0eo5

## Checklist
- [x] build_https parses fail-closed and stores a token source, never a secret
- [x] Scope grammar and longest-prefix matcher reused from the SSH surface, not duplicated
- [x] Not present in the lockable system-configuration keys
- [x] go test ./internal/config green
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
spawn queued: [implementer] developer (claude) (run=RUN-260824-e08865, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260824-e08865)
Shipped internal/config/buildhttps.go: Config.BuildHTTPS map of scope -> {token?, token_env?, username?}, mirroring build_ssh. Token sources enumerated as TokenSourceGitCredentials="git-credentials" (operators own Git HTTPS credential for the host) and TokenSourceKeyring="keyring" (manager-namespaced entry a future login command stores) -- no prior enum existed in this repo, so these names were chosen to match the two-source language already in TASK-260825-1lausy/1tgpcn/2gyhq8 descriptions. token_env is the only field that may hold operator text (an env var name); a literal secret in token is rejected ("...; secrets never live in the config"). Scope grammar reused directly via ValidBuildSSHScope; longest-prefix match now lives in a shared generic longestScope[T] helper in buildssh.go used by both MatchBuildSSH and MatchBuildHTTPS. build_https added to managerKeys, deliberately NOT added to LockableKeys (same as build_ssh). No SetBuildHTTPS/RemoveBuildHTTPS write helpers added -- out of scope here, needed by TASK-260825-2gyhq8. go test ./internal/config, go vet ./..., golangci-lint run ./... all green. Downstream tasks should keep the git-credentials/keyring naming consistent.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-e08865, pid=67909, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-effc22, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-effc22)
Review verdict: ACCEPTED. All AC verified against the code, gates rerun independently (go build, go test ./internal/config -count=1, go vet, golangci-lint -- all green). Scope grammar and matcher genuinely reused via ValidBuildSSHScope + shared generic longestScope[T]; buildssh.go diff is behavior-preserving. build_https in managerKeys, not in LockableKeys, both pinned by tests. Literal-secret, exactly-one-source, grammar, and segment-boundary cases all covered. Evidence: TASK-260825-168m7o_review-verdict.md. No commit_ack from this reviewer run; uncommitted scope (buildhttps.go, buildhttps_test.go, buildssh.go, config.go) is for the commit-owning mover.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-effc22, pid=60653, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260825-168m7o_spawn-log_-implementer--developer--claude-_RUN-260824-e08865.log](file://TASK-260825-168m7o/TASK-260825-168m7o_spawn-log_-implementer--developer--claude-_RUN-260824-e08865.log) — System spawn log captured by task-board
- [TASK-260825-168m7o_results.md](file://TASK-260825-168m7o/TASK-260825-168m7o_results.md) — Design decisions, config shape, and validation evidence for build_https
- [TASK-260825-168m7o_spawn-log_-reviewer--reviewer--claude-_RUN-260824-effc22.log](file://TASK-260825-168m7o/TASK-260825-168m7o_spawn-log_-reviewer--reviewer--claude-_RUN-260824-effc22.log) — System spawn log captured by task-board
- [TASK-260825-168m7o_review-verdict.md](file://TASK-260825-168m7o/TASK-260825-168m7o_review-verdict.md) — Reviewer verdict: accepted, with independently rerun gates

## Created
2026-08-24T21:23:14Z

## Last Update
2026-08-24T22:03:08Z

## Assigned To
[reviewer] reviewer (claude)
