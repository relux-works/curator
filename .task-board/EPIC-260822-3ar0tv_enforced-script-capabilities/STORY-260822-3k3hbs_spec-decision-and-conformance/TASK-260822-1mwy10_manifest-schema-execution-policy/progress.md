## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- TASK-260822-1l4r4f

## Blocks
- TASK-260822-f4qv7w

## Checklist
- [x] Manifest schema admitting execution_policy on script commands committed on branch spec/sw-schema
- [x] Validation prose with defaults and rejection paths specified
- [x] Schema version numbering coordination with STORY-260822-1pm1c9 recorded in task notes
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
Own worktree: git -C /Users/iv/Developer/ReluxWorks/curator-spec worktree add .temp/STORY-260822-3k3hbs/schema-worktree -b spec/sw-schema origin/main. Commit without AI attribution, push -u origin spec/sw-schema, no PR — the landing task merges. Consume analysis.md of TASK-260822-1l4r4f. Note: decision 0009 (first-party module roots, landed b92b105) also needs a schema bump — coordinate numbering: script-worker execution_policy and module roots may share one schema version or take sequential ones; record the choice.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-c60b49, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-c60b49)
Manifest schema 8 landed on branch spec/sw-schema, commit ebfed81 (signed, no AI attribution), pushed, no PR — TASK-260822-c0rxj7 merges. Base origin/main=b92b105. Evidence: TASK-260822-1mwy10_results.md.

SCHEMA VERSION NUMBERING COORDINATION WITH STORY-260822-1pm1c9 — DECIDED: ONE SHARED BUMP. Schema 8 carries both decision 0008 (execution_policy on script commands) and decision 0009 (first-party module roots). Module roots takes no separate version; TASK-260822-3nvx91 extends the build-command branch of $defs.commandV8 rather than creating schema 9. Reasons: (1) no protocol release has ever carried two manifest schema versions (rc.4=6, rc.5=7); (2) the two surfaces are disjoint — script commands vs local build commands; (3) sequential 8/9 would leave schema 8 born superseded, since any manifest wanting both features must use the higher version; (4) sequential doubles the legacy rejection matrix and forces install-marker v4 AND v5, because skill_schema_version is a const per marker version; (5) no freeze is violated — COMPATIBILITY.md freezes released bytes, and schema 8 has not shipped; rc.5/rc.6/rc.7 metadata is unchanged by this commit.

WHAT SHIPPED. $defs.scriptCommandV8 = schema-7 script command + two OPTIONAL fields: execution_policy bound to const script-worker-v1, interpreter bound to enum [node-v1, python3-v1]. dependentRequired binds both directions. New $defs.commandV8 union. New schemas/v1/agent-skill-v8, csk-skill-v8, install-marker-v4. Validation prose with defaults and rejection paths in schemas/v1/README.md.

INSTALL-MARKER V4 — necessary consequence, not scope creep. install-marker-v3 pins skill_schema_version to const 7 and v2 to 0-6, so a schema-8 install had no marker version and schema 8 would have been uninstallable. Marker v4 is marker v3 with schema_version 4 and skill_schema_version 8 and no other difference, asserted by a test that strips those two properties and compares the documents.

INTERPRETER SET CONVERGED WITH THE PROSE TASK. The analysis left the exact interpreter set to the prose task. Schema 8 admits node-v1 and python3-v1 and no shell. origin/spec/sw-core-prose commit 41cf556 (TASK-260822-1f533i) independently states "Protocol 1.0 admits exactly python3-v1 and node-v1" and defers bash-v1/powershell-v1 for the same platform-resolution reason. No reconciliation needed. Their capability-default rule also matches the schema prose.

GATES, real exit codes, each a standalone process at ebfed81: python3 tools/validate.py = 0 (52 schemas, 656 vector files); python3 -B -m unittest discover -s tools -p test_*.py = 0 (94 tests, was 91); go test ./tools/... = 0; gofmt -l tools = 0 (no output); go run ./tools/generate-vectors -root . twice, each followed by git diff --exit-code over conformance/v1 and all four release/*.json = 0 all four times, so double regeneration is proven clean per GOVERNANCE. Not run: lychee links and release-provenance CI jobs (network/token), and make release-check (needs VERSION, release-time gate). Frozen-bytes evidence: git diff --name-only over conformance/v1/schema-cases reports exactly one modified file, index.json; the other 186 case changes are all new files, so no schema-1-through-7 case instance changed a byte.

CORE.MD DELTAS STILL OWED — not made here, protocol/core.md is TASK-260822-1f533i s file and origin/spec/sw-core-prose does not contain them yet. Without them the spec contradicts schema 8. (1) Section 4 preamble still says a manifest conforms to exactly one of agent-skill-v1 through v7 — must become v8, same for csk-skill-v7. (2) Section 4 schema table needs a row 8. (3) Section 4 version gates: schemas 2 through 7 reject unknown fields -> 2 through 8, plus a sentence that schemas 1 through 7 MUST reject execution_policy and interpreter. (4) Section 10: managers supporting schema 8 MUST read marker schemas 1-4 and write marker 4 for schema-8 mutations, plus a Marker v4 paragraph. (5) COMPATIBILITY.md rc.9 paragraph naming schema 8, marker v4, and the shared-bump decision — already inside TASK-260822-c0rxj7 AC.

HOOKS LEFT FOR TASK-260822-f4qv7w: mixed_build_cases in tools/generate-vectors/main.go maps manifest_schema to marker_version and has no schema8 row yet; install-marker-v4 has structural schema cases but no expected marker fixture under conformance/v1/expected/.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-c60b49, pid=85826, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-53fb45, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-53fb45)
REVIEWER VERDICT: ACCEPTED (RUN-260822-53fb45, not goal-bound). Subject spec/sw-schema @ ebfed81, base b92b105. Evidence: TASK-260822-1mwy10_review.md.

PROVENANCE. ebfed81 signature Good, ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM, oparin@me.com, verified against maintainers.allowed_signers. No AI attribution in the message. origin/spec/sw-schema = ebfed81, no PR as instructed.

GATES REPRODUCED INDEPENDENTLY at ebfed81 in a scratch detached worktree with a fresh venv (jsonschema 4.25.1), not the implementer clone: validate.py exit 0 (52 schemas, 656 vector files); unittest 94 tests OK; go test ./tools/... ok; go vet clean; gofmt -l tools empty; git diff --check and git show --check clean; go run ./tools/generate-vectors -root . twice, each followed by git diff --exit-code over conformance/v1 and release/, 0 on all four checks. The implementer counts reproduce exactly. Not run, same as reported: lychee links, release-provenance, make release-check.

FROZEN BYTES. 9 modified files, 185 added, 0 deleted. Under conformance only the two inventories changed (manifest.json, schema-cases/index.json); no schema-1..7 or marker-v1..v3 case instance changed a byte. release/1.0.0-rc.8.json pin update is REQUIRED, not a violation: validate.py freezes rc.5/rc.6/rc.7 by exact digest constants and separately asserts rc.8 candidate pin == live suite manifest digest; make regenerate-check lists all four for that reason. rc.5/6/7 bytes unchanged.

SCHEMA BEHAVIOUR PROBED INDEPENDENTLY (37 cases, not the repo vectors). Schema 8 accepts both fields, neither, node-v1, unix-only, win-only. Rejects policy-alone, interpreter-alone, null, "none", false, manager-worker-v1, hardened-worker-v1, script-worker-v2, bash-v1, pathless enforced command, the fields on system/go-v1/go-repository-v1 commands, and the fields at top level. Schemas 1..7 reject both fields top-level and per-command. ATTRIBUTION PROVEN: for each of v1..v7 the identical instance without the schema-8 fields ACCEPTs, so the legacy rejection is caused by the new fields, not incidental invalidity. The prose split is exact: v2..v7 via additionalProperties, v1 via validate.py wire semantics. Capability default annotations quoted in the README are factually correct.

STRUCTURAL DIFFS ARE MINIMAL. agent-skill-v8 vs v7 differ in exactly 4 lines ($id, title, schema_version const, commandV7->commandV8); same for csk-skill-v8. install-marker-v4 vs v3 differ in exactly 4 lines ($id, title, schema_version 3->4, skill_schema_version 7->8), asserted by the repo own document-equality test. commandV7 still refs the frozen scriptCommand.

CROSS-BRANCH COHERENCE CHECKED DIRECTLY, not taken on trust. origin/spec/sw-core-prose@41cf556 independently states: core.md:211 selection is an OPTIONAL per-command field on schema 8 or later (matches, resolves decision 0008 open question 1); core.md:256-257 and :282 admit exactly python3-v1 and node-v1 and defer bash-v1/powershell-v1 (matches scriptInterpreterV1, resolves open question 2); core.md:417 declared host globs are recorded and reported and mean no filtering (matches the README reporting-only reading, resolves open question 3); the capability derivation rule matches the README annotation-vs-derivation paragraph. The interpreter field is therefore the normative resolution of decision 0008 open question 2, not an unauthorised addition, and both halves of the story landed on the same answer.

OWED DELTAS VERIFIED GENUINELY STILL OWED on origin/spec/sw-core-prose@41cf556, so the implementer handoff notes are accurate and actionable: core.md:157 still says v1 through v7; :182 table stops at row 7; :186 still says schemas 2 through 7 reject unknown fields; :1539/:1554 still say read marker schemas 1,2,3 and marker v3 permits skill_schema_version through 7; COMPATIBILITY.md has no rc.9 paragraph. CHANGELOG.md has no Unreleased section so nothing is owed there pre-release.

ADVISORY FOR DOWNSTREAM, not defects and not blocking. (1) TASK-260822-3nvx91: the README says module roots extends the build-command branch reached from $defs.commandV8, but commandV8 build branches ref buildCommandV6 and repositoryBuildCommandV1, both SHARED with frozen commandV6/commandV7 — that extension must add a new $def, never edit buildCommandV6 in place, or it silently mutates schemas 6 and 7. (2) Whichever schema-8 branch lands second must regenerate: this branch already generated 185 schema-8 case files and the rc.8 pin for a schema 8 without module roots. (3) TASK-260822-f4qv7w: confirmed both flagged hooks — mixed_build_cases at tools/generate-vectors/main.go:1187 has rows only through schema7-*, and install-marker-v4 has 27 structural cases mirroring v3 exactly but no expected marker fixture under conformance/v1/expected/.

Reviewer supplied no commit_ack; the scope is already committed, signed, and pushed at ebfed81.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-53fb45, pid=16038, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-1mwy10_spawn-log_-implementer--developer--claude-_RUN-260822-c60b49.log](file://TASK-260822-1mwy10/TASK-260822-1mwy10_spawn-log_-implementer--developer--claude-_RUN-260822-c60b49.log) — System spawn log captured by task-board
- [TASK-260822-1mwy10_results.md](file://TASK-260822-1mwy10/TASK-260822-1mwy10_results.md) — Manifest schema 8 delivery: script-worker-v1 opt-in surface, install-marker-v4, defaults and rejection paths, shared-bump numbering decision with STORY-260822-1pm1c9, gate exit codes, and the core.md deltas still owed
- [TASK-260822-1mwy10_spawn-log_-reviewer--reviewer--claude-_RUN-260822-53fb45.log](file://TASK-260822-1mwy10/TASK-260822-1mwy10_spawn-log_-reviewer--reviewer--claude-_RUN-260822-53fb45.log) — System spawn log captured by task-board
- [TASK-260822-1mwy10_review.md](file://TASK-260822-1mwy10/TASK-260822-1mwy10_review.md) — Reviewer verdict (accepted) for manifest schema 8: independent gate reproduction, frozen-byte audit, 37-case schema behaviour probe with attribution controls, cross-branch coherence check against spec/sw-core-prose
- [TASK-260822-1mwy10_reviewer-gate-log.md](file://TASK-260822-1mwy10/TASK-260822-1mwy10_reviewer-gate-log.md) — Reviewer gate run log at ebfed81 (validate.py + unittest, pre-venv failure then clean run)

## Created
2026-08-22T16:00:35Z

## Last Update
2026-08-22T17:17:14Z

## Assigned To
[reviewer] reviewer (claude)
