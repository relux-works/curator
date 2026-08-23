## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260822-2505vo
- TASK-260822-3pkc80

## Blocks
- TASK-260822-4p3dcq
- TASK-260823-qwr5w9

## Checklist
- [x] Precheck before any fetch; interactive candidate menu with agent key count and ~/.ssh *.pub; default agent+pin
- [x] Persistence only after explicit scope choice; abort path covered
- [x] Non-interactive fail-closed message carries ready-to-run add commands from detected candidates
- [x] Dry-run reports per-repository credential source; go test green
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
spawn queued: [implementer] developer (claude) (run=RUN-260822-e5785a, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-e5785a)
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260822-e5785a, pid=3467, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-46ddd9, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-46ddd9)
Work continues in .temp/TASK-260822-b0wg3a/worktree on branch task/TASK-260822-b0wg3a-install-precheck, based off origin/main 6a9b201 (the checkpoint branch has no internal/install/external.go). The branch carries the predecessor stack 96m5pj + 2505vo + 3pkc80 verbatim, then adds the precheck: internal/install/buildsshcandidates.go (agent socket + ssh-add -l key count degrading to unknown, ~/.ssh/*.pub listing capped at 8, ready-to-run add commands) and internal/install/buildsshprompt.go (two-question flow: credential then scope, persist only after both). Gates so far on the final tree: go build 0, go vet 0, gofmt clean, golangci-lint v2.12.2 0 issues, .github/ci/gate-selftest.sh 75 passed 0 failed, 39 named buildssh tests pass. Full suite running; cmd/curator alone needs ~10 min (toolchain fingerprinting), so a 5m -timeout reports a false hang.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-46ddd9, pid=21288, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-fad8ab, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-fad8ab)
reviewer verdict RUN-260822-fad8ab: ACCEPTED (read-only, not goal-bound). All AC met. Provenance re-verified independently: rebuilt the 2505vo baseline from origin/main 6a9b201 + the accepted patch and diffed the whole tree — all seven shared files byte-identical, so the delta charged here is exactly buildsshcandidates.go + buildsshprompt.go + 3 test files, edits to install/buildssh.go, install.go, global.go, the cmd/curator resolver wiring, and the go.mod term promotion. The attached final.patch re-diffs identical to the working tree. Gates re-run at first hand in the worktree: go build 0, go vet 0, gofmt clean, golangci-lint v2.12.2 0 issues, gate-selftest 75 passed 0 failed, go test -count=1 -timeout 30m ./... exit 0 (41 packages, cmd/curator 486.238s), -race clean on the 36 named precheck/prompt/candidate tests. Coverage on the new code 78.9-100% per function. Four non-blocking findings, none of which weakens the fail-closed guarantee: (1) the persist closure writes config.SetBuildSSH(cfg.Path,...) but never updates cfg.BuildSSH, so install --all re-asks per target; (2) identity-only candidates built from *.pub cannot authenticate under IdentityAgent=none — implementer flagged it, confirmed against admission.go:150-168, docs item for TASK-260822-4p3dcq; (3) defaultBuildSSHScope can derive a scope that ValidBuildSSHScope rejects, because PortableComponent admits runes scopeSegmentRE does not — the prompt then rejects its own default and the diagnostic emits an unrunnable command; (4) nit: Sscanf(%d) accepts trailing garbage in the menu answer. Evidence: TASK-260822-b0wg3a_review-verdict-RUN-fad8ab.md, TASK-260822-b0wg3a_review-fullsuite-RUN-fad8ab.log, LOGBOOK 2026-08-23 0330. Nothing committed: the reviewed tree is the staged index of .temp/TASK-260822-b0wg3a/worktree; the commit-owning mover lands the attached patch.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-fad8ab, pid=55355, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-b0wg3a_spawn-log_-implementer--developer--claude-_RUN-260822-e5785a.log](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_spawn-log_-implementer--developer--claude-_RUN-260822-e5785a.log) — System spawn log captured by task-board
- [TASK-260822-b0wg3a_spawn-log_-implementer--developer--claude-_RUN-260822-46ddd9.log](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_spawn-log_-implementer--developer--claude-_RUN-260822-46ddd9.log) — System spawn log captured by task-board
- [TASK-260822-b0wg3a_results.md](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_results.md) — Implementation notes: discovery/prompt/precheck design, AC-to-test map, gate results, the cmd/curator timeout trap, and the .pub identity-only observation
- [TASK-260822-b0wg3a_final.patch](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_final.patch) — Full worktree diff against origin/main 6a9b201: the 96m5pj+2505vo+3pkc80 stack plus this task's install precheck, candidates, and prompt
- [TASK-260822-b0wg3a_full-suite-final.log](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_full-suite-final.log)
- [TASK-260822-b0wg3a_spawn-log_-reviewer--reviewer--claude-_RUN-260822-fad8ab.log](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_spawn-log_-reviewer--reviewer--claude-_RUN-260822-fad8ab.log) — System spawn log captured by task-board
- [TASK-260822-b0wg3a_review-verdict-RUN-fad8ab.md](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_review-verdict-RUN-fad8ab.md) — Reviewer verdict RUN-260822-fad8ab: accepted; provenance re-verified against the 2505vo baseline, all gates re-run, four non-blocking findings
- [TASK-260822-b0wg3a_review-fullsuite-RUN-fad8ab.log](file://TASK-260822-b0wg3a/TASK-260822-b0wg3a_review-fullsuite-RUN-fad8ab.log) — Reviewer's independent go test -count=1 -timeout 30m ./... run: exit 0, 41 packages ok, cmd/curator 486.238s

## Created
2026-08-22T16:12:06Z

## Last Update
2026-08-22T22:03:03Z

## Assigned To
[reviewer] reviewer (claude)
