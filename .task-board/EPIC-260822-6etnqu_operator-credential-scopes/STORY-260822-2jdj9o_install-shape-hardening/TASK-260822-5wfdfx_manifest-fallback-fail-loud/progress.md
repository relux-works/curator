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
- [x] Regression: canonical manifest schema 99 + runtime.json alongside errors with upgrade hint, no fallback
- [x] Fallback reachable only when no manifest exists; go test green
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
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-c1807e, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-c1807e)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-b89c6b, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-b89c6b)
Tests-only change; internal/skillspec/parse.go untouched (git diff empty). The fail-loud rule already held: Load returns loadCskSkill(cskPath) error included (parse.go:28), so the agents/runtime.json branch is unreachable while a manifest exists. Added to internal/skillspec/parse_test.go: TestBrokenCskSkillNeverFallsBackToLegacy (5 cases: schema_version 99, malformed JSON, top-level array, zero-byte manifest, schema 3 without capabilities — each with a VALID agents/runtime.json planted alongside, asserting error + nil spec); TestNewerSchemaVersionReportsUpgradeHint (*verr.Error on path schema_version, message contains UpgradeHint and 99); TestLegacyFallbackNeedsAbsentManifest (positive control — same payload loads only once csk-skill.json is removed, proving precedence, not a broken fallback file). Mutation-checked: swallowing the manifest error in Load turned all 6 tests red with SourceFile:agents/runtime.json leaking; parse.go restored and re-verified. Gates: go build ./... 0, go vet ./... 0, go test ./... 0 (31 pkg, 0 FAIL), golangci-lint run 0 (0 issues), gofmt -l clean. FOR REVIEWER: task description says agent-skill.json; the repo canonical name is csk-skill.json and agent-skill.json exists nowhere in the tree — tests use the real name.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-b89c6b, pid=16299, exit=0)
Fail-loud manifest precedence, handed to review.

FINDING (split result): the parse-error half was already correct at HEAD -- a present csk-skill.json with malformed JSON, schema_version 99 or an invalid body has always been terminal, never a fallback. The real defect was presence detection: Load probed with os.Stat and treated EVERY stat failure as absence, so a dangling manifest symlink loaded agents/runtime.json (SourceFile: agents/runtime.json, legacy commands) and an unreadable snapshot returned the empty pure-context spec.

FIX internal/skillspec/parse.go: precedence now goes through manifestPresent(), which uses os.Lstat and returns an error for any failure that is not fs.ErrNotExist. Applied to the legacy manifest too. Fallback is now reachable only when no manifest entry exists at all.

TESTS internal/skillspec/parse_test.go, 8 regression tests, all run (none skipped) on darwin: schema-99 + runtime.json alongside -> error path schema_version carrying UpgradeHint and the value 99; 5 broken bodies with a parseable fallback planted alongside; unreadable manifest; manifest-as-directory; dangling manifest symlink; unreadable snapshot; dangling legacy manifest; plus the positive control that the same legacy payload loads once the manifest is removed.

MUTATION CHECK (throwaway module copy, live checkout never mutated): swallowing the parse error -> exit 1, 12 cases; reverting only presence to os.Stat -> exit 1, exactly the 3 presence tests, schema-99 tests still green.

VALIDATION (real exit codes): go test -count=1 ./... 0; skillspec + all 5 Load consumers 0; go vet ./... 0; GOOS=windows/linux vet on skillspec 0; gofmt -l cmd internal empty; golangci-lint run 0 issues; go build 0.
NOT GREEN, NOT MINE: make ledger-check exit 2 on six pre-existing rows naming packages absent from the module (internal/transaction, internal/install/atomicity, internal/godriver). .github/ untouched.
NOT RUN: make ci-test / check-ci / candidate-* -- need CURATOR_CONFORMANCE_ROOT, unset on this host.

NAMING: the description says agent-skill.json; this repo reads csk-skill.json (docs/implementation-plan.md:41). The agent-skill.json rename lives in the rc.5 fixture root and is tracked by LOGBOOK 1615 / TASK-260720-z9j4c9, so the tests target the name this build implements.
ANOMALY: two implementer spawns (RUN-260822-c1807e, RUN-260822-b89c6b) ran concurrently on this one task and both edited parse_test.go in the same checkout; reconciled by hand. Same hazard as TASK-260822-96m5pj today.
Artifacts: TASK-260822-5wfdfx_manifest-fallback-fail-loud.md, TASK-260822-5wfdfx_mutant-2-stat-presence.log.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-c1807e, pid=13779, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-694018, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-694018)
REVIEW VERDICT: ACCEPTED (RUN-260822-694018, read-only, no code touched).

AC verified clause by clause on this branch. Regression tests present in internal/skillspec (8, all confirmed to RUN on darwin, zero skips). schema_version 99 + a valid agents/runtime.json planted alongside errors with *verr.Error on path schema_version carrying UpgradeHint and 99, and returns a nil spec. Positive control TestLegacyFallbackNeedsAbsentManifest proves the fallback loads only once the manifest is removed, so the failures come from precedence and not from a broken fallback payload.

BOTH MUTANTS REPRODUCED INDEPENDENTLY in a throwaway module copy under .temp/TASK-260822-5wfdfx-mutation (live checkout never mutated). Mutant 1, parse.go reverted to HEAD: exit 1, exactly the 3 presence tests fail, schema-99 tests stay green - which is itself the proof that the parse-error half was already correct and only presence detection was broken. Mutant 2, swallow the manifest parse error: exit 1, 19 failures including all 5 broken-manifest subtests and the upgrade-hint test. Matches the attached mutant log line for line.

GATES RE-RUN STANDALONE: go build ./... 0; go vet ./... 0; GOOS=windows and GOOS=linux vet on skillspec 0; gofmt -l cmd internal empty; go test -count=1 ./... 0 (31 ok, 0 FAIL); golangci-lint run 0 issues.

SCOPE JUDGMENT: the parse.go change is not scope creep. The story clause reads present-but-unreadable OR newer-schema, and the AC reads fallback reachable only when no manifest file exists at all - presence detection is exactly what those sentences are about. Both Load consumers (internal/closure/closure.go:288, internal/skillcheck/skillcheck.go:32) already treat the error as fatal, so the new error paths surface through existing channels.

F1 CARRY FORWARD, THE FINDING THAT OUTLIVES THIS TASK: main has already replaced this Load entirely (agent-skill.json canonical, csk-skill.json legacy alias, dual-file conflict detection). Its presence helper is os.Stat + os.IsNotExist at main:internal/skillspec/parse.go:91-100, so MAIN STILL CARRIES THE EXACT DEFECT FIXED HERE, across three probes instead of two, plus ManifestSourcePath which discards the error outright. This branch fix does not merge into it: the function it patches does not exist there and the tests hardcode csk-skill.json where main uses CanonicalManifestName/LegacyManifestName. Recommend a separate follow-up task against main.

F2: LOGBOOK:220 is wrong to call agent-skill.json a stale alias in upstream task text. curator-spec/protocol/core.md:35-40 makes it the canonical implementation-neutral filename and csk-skill.json a reserved legacy READ alias that writers must not emit; main implements exactly that. The implementer decision to test csk-skill.json was still correct (the canonical name exists nowhere on this branch) - only the justification needs correcting. Residual value confirmed narrow but real: main TestCanonicalManifestResolution/invalid_peer_does_not_fall_back plants no agents/runtime.json, tests no schema 99 and no presence failure, so this coverage is new.

F3: stale artifact. TASK-260822-5wfdfx_results.md from the first concurrent spawn says No production code was changed, falsified by the delivered diff; LOGBOOK 2016 carries the same superseded claim that LOGBOOK 2025 corrects nine minutes later. Authoritative artifact is TASK-260822-5wfdfx_manifest-fallback-fail-loud.md. Recommend the mover update results.md in place rather than leave two board artifacts telling different stories.

Note: errors.Is(ENOTDIR, fs.ErrNotExist) is false on darwin (probed), so Load on a regular file now errors instead of returning the empty spec. Correct, and it converges with main behavior.

HANDOFF, NO commit_ack FROM THIS REVIEWER: commit scope is exactly internal/skillspec/parse.go and internal/skillspec/parse_test.go plus the LOGBOOK entries for this task. The working tree carries unrelated sibling-task work in cmd/curator/, internal/closure/, internal/config/, internal/install/ and internal/skillcheck/ that must not be swept in. The commit-owning mover commits that scope, then makes the final done transition with commit_ack=scope_committed.

Artifacts: TASK-260822-5wfdfx_review-verdict.md. Logbook entry 2106.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-694018, pid=14346, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-5wfdfx_spawn-log_-implementer--developer--claude-_RUN-260822-c1807e.log](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_spawn-log_-implementer--developer--claude-_RUN-260822-c1807e.log) — System spawn log captured by task-board
- [TASK-260822-5wfdfx_spawn-log_-implementer--developer--claude-_RUN-260822-b89c6b.log](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_spawn-log_-implementer--developer--claude-_RUN-260822-b89c6b.log) — System spawn log captured by task-board
- [TASK-260822-5wfdfx_results.md](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_results.md) — Regression tests pinning manifest fail-loud + no-fallback, with mutation check and gate exit codes
- [TASK-260822-5wfdfx_manifest-fallback-fail-loud.md](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_manifest-fallback-fail-loud.md) — Implementation notes: fail-loud manifest precedence fix, regression tests, mutation check and validation exit codes
- [TASK-260822-5wfdfx_mutant-2-stat-presence.log](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_mutant-2-stat-presence.log) — go test -v stream of the pre-fix os.Stat presence mutant: only the three presence tests fail
- [TASK-260822-5wfdfx_spawn-log_-reviewer--reviewer--claude-_RUN-260822-694018.log](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_spawn-log_-reviewer--reviewer--claude-_RUN-260822-694018.log) — System spawn log captured by task-board
- [TASK-260822-5wfdfx_review-verdict.md](file://TASK-260822-5wfdfx/TASK-260822-5wfdfx_review-verdict.md) — Reviewer verdict accepted: AC checked clause by clause, both mutants reproduced independently, gates re-run; three carry-forward findings incl. main still carrying the os.Stat presence defect

## Created
2026-08-22T16:12:28Z

## Last Update
2026-08-22T17:08:40Z

## Assigned To
[reviewer] reviewer (claude)
