## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- TASK-260822-96m5pj

## Blocks
- TASK-260822-b0wg3a

## Checklist
- [x] config build-ssh add/list/remove implemented with grammar validation and sorted list output
- [x] Help documents precedence flags > env > config scopes
- [x] CLI tests incl. validation failures; go test green
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
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-d2b78c, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-d2b78c)
Ready for review. Implemented `curator config build-ssh add|list|remove` plus a real `config` dispatcher (cmdConfig replaces the inline `config show` case in run()).

Behaviour: add validates through config.ValidateBuildSSH (same grammar as parseBuildSSH) BEFORE reading the config, so a malformed invocation is exit 2 and is never attributed to the config file; it replaces the whole entry under a scope (a merge would leave a stale agent selection able to authenticate) and prints added|replaced. list prints one sorted <scope>TAB<fields> line per scope; with no scopes stdout stays empty and the note goes to stderr. remove of a scope not recorded in the user config is exit 1 with a message naming the config path; removing the last scope drops the whole build_ssh field rather than leaving {}.

--agent [SOCKET] uses the parseInterspersed AcceptsOptionalValue hook and claims the next token only when it reads as a credential path, so `add --agent git.example.com` keeps the scope positional. --agent=false is rejected, matching the config grammar.

New config API: SetBuildSSH / RemoveBuildSSH in internal/config/write.go, ValidBuildSSHPath exported in buildssh.go.

REVIEWER NOTE / cross-task constraint: the help text now commits to the CURATOR_BUILD_SSH_* env prefix (matching CURATOR_CONFIG / CURATOR_SYSTEM_CONFIG / CURATOR_REGISTRY_TOKEN) and a test pins the precedence ordering flags > env > config scopes. No flags/env resolution layer exists yet - TASK-260822-2505vo must adopt that prefix or update buildSSHUsage.

Verification, each run standalone with its real exit code: go build ./... = 0; go vet ./... = 0; gofmt -l cmd internal = 0 (no output); go test -count=1 ./... = 0; golangci-lint run = 0 (0 issues); go build -o bin/curator ./cmd/curator = 0. Mutation check: patching SetBuildSSH to merge instead of replace turns the suite red, so the replacement claim is covered. NOT run: make ci-test / check-ci - they require CURATOR_CONFORMANCE_ROOT pointing at a materialised conformance/v1 checkout, which is not available in this session; CI runs them from the committed pin.

Working tree is uncommitted by policy (no auto-commit). Files: cmd/curator/main.go, cmd/curator/main_test.go, internal/config/write.go, internal/config/write_test.go, internal/config/buildssh.go, LOGBOOK.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-d2b78c, pid=15097, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-0dff48, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-0dff48)
REVIEW VERDICT: accepted (RUN-260822-0dff48, read-only). Evidence: TASK-260822-3pkc80_review.md.

Gates re-run independently by the reviewer, each standalone with its real exit code: go build ./... = 0; go vet ./... = 0; gofmt -l cmd internal = no output; go test -count=1 ./... = 0 (all packages ok); go test -count=1 ./cmd/... ./internal/config/... = 0; golangci-lint run = 0 issues. NOT run: make ci-test / check-ci (need CURATOR_CONFORMANCE_ROOT; not in this task AC).

Beyond the unit tests, the built binary was exercised against four scratch config files: bare --agent does not swallow the scope (add --agent alpha.example.com keeps the scope positional); a re-add drops the previous agent selection from the JSON rather than merging it; remove of an unconfigured scope is exit 1 naming the user config path; removing the last scope drops the whole build_ssh field; list is sorted <scope>TAB<fields> with empty stdout and a stderr note when nothing is configured; 13 malformed invocations all exit 2 and leave the file byte-identical. config show marshals BuildSSH, so the field is visible in the effective configuration for free.

AC met: add/list/remove with grammar validation via ValidateBuildSSH before the config is read (malformed invocation = exit 2, never attributed to the config file); sorted list; help states precedence and TestConfigHelpDocumentsPrecedenceAndSubcommands pins the ordering by string index rather than by word presence; validation-failure tests present; go test green. Fit: SetBuildSSH/RemoveBuildSSH follow AddProject exactly (readObject -> mutate -> Parse as pre-write gate -> writeObjectAtomic 0600), and TASK-260822-96m5pj logbook warning about calling ValidateBuildSSH before BuildSSHObject is honoured and pinned by the socket-no-agent test case.

Non-blocking, carried forward: (1) LIVE CROSS-TASK CONSTRAINT - the help commits to CURATOR_BUILD_SSH_* with no resolver behind it; TASK-260822-2505vo (currently development) must adopt that prefix or update buildSSHUsage in the same change, and the story text also mentions a CSK_ alias that the help does not document. (2) A system-config build_ssh is fully masked by the operator first add (whole-key override per Spec 7.2, same as projects/allowed_sources) and build_ssh is a manager key but not in LockableKeys - worth a line from TASK-260822-4p3dcq. (3) list is not space-safe for paths containing spaces (config show is the machine surface). (4) list ignores trailing args while remove rejects them - matches the repo argless-command convention. (5) The pre-write Parse gate validates the user object alone, inherited verbatim from AddProject.

Reviewer archetype supplies no commit_ack; the working tree stays uncommitted by policy and the commit-owning mover carries this scope into the Story/Bug commit.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-0dff48, pid=79928, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-3pkc80_spawn-log_-implementer--developer--claude-_RUN-260822-d2b78c.log](file://TASK-260822-3pkc80/TASK-260822-3pkc80_spawn-log_-implementer--developer--claude-_RUN-260822-d2b78c.log) — System spawn log captured by task-board
- [TASK-260822-3pkc80_implementation-notes.md](file://TASK-260822-3pkc80/TASK-260822-3pkc80_implementation-notes.md) — CLI surface, exit-code contract, env-prefix constraint for the resolver task, tests and verification evidence
- [TASK-260822-3pkc80_verification.log](file://TASK-260822-3pkc80/TASK-260822-3pkc80_verification.log) — go test -count=1 ./... (exit 0) and golangci-lint run (exit 0) output
- [TASK-260822-3pkc80_spawn-log_-reviewer--reviewer--claude-_RUN-260822-0dff48.log](file://TASK-260822-3pkc80/TASK-260822-3pkc80_spawn-log_-reviewer--reviewer--claude-_RUN-260822-0dff48.log) — System spawn log captured by task-board
- [TASK-260822-3pkc80_review.md](file://TASK-260822-3pkc80/TASK-260822-3pkc80_review.md) — Reviewer verdict: accepted; independent gate re-run, live CLI behavioural verification, AC mapping, cross-task env-prefix constraint

## Created
2026-08-22T16:12:06Z

## Last Update
2026-08-22T17:25:22Z

## Assigned To
[reviewer] reviewer (claude)
