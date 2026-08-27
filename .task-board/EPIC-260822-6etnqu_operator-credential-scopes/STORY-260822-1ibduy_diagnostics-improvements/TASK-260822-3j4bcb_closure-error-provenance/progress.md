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
- [x] Closure manifest parse errors name node, resolved ref, and chain
- [x] Test with broken transitive manifest asserts the text; protocol codes unchanged; go test green
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
spawn queued: [implementer] developer (claude) (run=RUN-260822-52b441, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-52b441)
Implementation ready for review.

Production change is one wrap in internal/closure/closure.go:288 (resolveNode):
  invalid skill manifest for <name> <kind> <ref> -> <short commit> (via <chain>): <verr>
Measured: invalid skill manifest for leaf tag v1 -> d0775b5f70d0 (via <project> -> mid -> leaf): commands: must be an object
Before: leaf: commands: must be an object (no ref, no chain; the name prefix read like a verr field path).

Placement: provenance is added at the call site, not inside skillspec.Load. The other caller is skillcheck.go:32, which backs standalone curator skill check and emits protocol issue skill.manifest_invalid; wording pushed down would leak closure-only context into a command with no closure and change a protocol payload. skill.manifest_invalid and its csk-skill.json path are unchanged, and skillcheck_test.go now asserts the boundary.

Install planning path: skillspec.Load has exactly two call sites repo-wide, so this single wrap covers install.Project, install.Global and cmdAudit -- all reach manifest parsing only via closure.Build. install.validateNodes -> skillcheck.Validate re-parses the same snapshot but is unreachable for manifest errors (closure.Build fails first on the same directory with the same deterministic parser), so it was left alone rather than given a second divergent format.

Tests: TestTransitiveManifestErrorNamesNodeRefAndChain (internal/closure, asserts name+ref+chain and that the declaring skill is not blamed), TestBrokenTransitiveManifestNamesTheBrokenSkill (internal/install, same provenance through Result.Errors), extended TestValidateMissingSkillAndInvalidManifest (internal/skillcheck, protocol code unchanged). New harness helper harness.brokenSkill. Red run captured before the fix.

Gates, each a standalone process: go build ./... exit 0; go test ./... exit 0; go vet ./... exit 0; golangci-lint run exit 0 (0 issues); gofmt -l ./cmd ./internal clean.
NOT run: make ci-test / check-ci -- they require CURATOR_CONFORMANCE_ROOT, which is unset in this environment; conformance_test.go self-skips without it. Note gofmt -l . at repo root is unusable here because .temp/ holds an unpacked go1.25.1 tree.

Artifacts: TASK-260822-3j4bcb_results.md, TASK-260822-3j4bcb_go-test.log, TASK-260822-3j4bcb_golangci-lint.log. Logbook entry 2026-08-22 2038. Changes are uncommitted per repo policy.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-52b441, pid=74943, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-bc9f3a, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-bc9f3a)
REVIEW VERDICT: ACCEPTED (RUN-260822-bc9f3a, read-only).

AC verified independently, not taken from the implementer log:
- name+ref+chain: TestTransitiveManifestErrorNamesNodeRefAndChain (closure_test.go:365) and TestBrokenTransitiveManifestNamesTheBrokenSkill (install_test.go:598) both re-run with -count=1, PASS. Install test binds the chain to closure.ProjectEdge rather than a literal.
- protocol codes unchanged: skill.manifest_invalid appears repo-wide only in skillcheck.go:34 and its test; payload and csk-skill.json path untouched; skillcheck_test.go pins the boundary.
- go test green.

Gates re-run by the reviewer: go build ./... 0; go test ./... no failures; go test -count=1 on internal/{closure,install,skillcheck,skillspec} + cmd/... all ok; go vet ./... 0; gofmt -l ./cmd ./internal empty; golangci-lint run 0 issues.

Conformance not run (CURATOR_CONFORMANCE_ROOT unset). Residual risk checked directly instead: internal/closure/conformance_test.go asserts provider order only, no error-string assertion; no vector or doc pins the closure manifest-error text; skillspec conformance untouched since skillspec.Load message is unchanged. Low residual risk, CI Implementations lane is the real gate.

Architecture fit confirmed: message shape mirrors sibling version conflict / cannot resolve messages in the same file. Call-site placement verified as load-bearing (skillcheck.go:32 is a protocol surface with no closure). Coverage claim verified: skillspec.Load has 2 call sites; all three closure.Build callers (install.go:179, global.go:59, main.go:1139) propagate err verbatim with no prefix; validateNodes -> skillcheck.Validate is genuinely unreachable for manifest errors.

Non-blocking (no rework): (1) message omits node.Substituted, so a dev-substituted node reads as revision HEAD -> <commit> with no hint it came from Skillfile.dev.json — outside AC, short commit still identifies what was parsed; (2) install test duplicates repo setup inline, justified since env.skill has no dependency/broken-manifest knobs; (3) the negative guard on "manifest for mid" is vacuous against the old text, but the positive assertions carry the red-green signal and the red run was captured.

HANDOFF TO COMMIT-OWNING MOVER: reviewer archetype supplies no commit_ack. Working tree carries unrelated uncommitted work — internal/skillspec/parse.go + parse_test.go belong to TASK-260822-5wfdfx, internal/config/config.go to the cocoaskills parity line. Commit ONLY internal/closure/closure.go, internal/closure/closure_test.go, internal/install/install_test.go, internal/skillcheck/skillcheck_test.go and this task LOGBOOK entry, then make the final done transition with commit_ack=scope_committed.

Evidence: TASK-260822-3j4bcb_review-verdict.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-bc9f3a, pid=15164, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-3j4bcb_spawn-log_-implementer--developer--claude-_RUN-260822-52b441.log](file://TASK-260822-3j4bcb/TASK-260822-3j4bcb_spawn-log_-implementer--developer--claude-_RUN-260822-52b441.log) — System spawn log captured by task-board
- [TASK-260822-3j4bcb_results.md](file://TASK-260822-3j4bcb/TASK-260822-3j4bcb_results.md) — Closure manifest error provenance: repro, change, placement decision, tests, gate exit codes
- [TASK-260822-3j4bcb_golangci-lint.log](file://TASK-260822-3j4bcb/TASK-260822-3j4bcb_golangci-lint.log) — golangci-lint run: exit 0, 0 issues
- [TASK-260822-3j4bcb_go-test.log](file://TASK-260822-3j4bcb/TASK-260822-3j4bcb_go-test.log) — go test ./...: exit 0, all packages ok
- [TASK-260822-3j4bcb_spawn-log_-reviewer--reviewer--claude-_RUN-260822-bc9f3a.log](file://TASK-260822-3j4bcb/TASK-260822-3j4bcb_spawn-log_-reviewer--reviewer--claude-_RUN-260822-bc9f3a.log) — System spawn log captured by task-board
- [TASK-260822-3j4bcb_review-verdict.md](file://TASK-260822-3j4bcb/TASK-260822-3j4bcb_review-verdict.md) — Reviewer verdict: accepted; AC-by-AC evidence, reviewer-re-run gate exit codes, conformance residual-risk check, non-blocking observations

## Created
2026-08-22T16:12:28Z

## Last Update
2026-08-22T16:45:07Z

## Assigned To
[reviewer] reviewer (claude)
