## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260728-zb2s4z

## Blocks
- TASK-260728-pwbr32
- TASK-260728-1ph8rs
- TASK-260728-2u5u14
- TASK-260728-2spy93
- TASK-260728-1g0z69

## Checklist
- [x] Every normative, schema, registry, fixture and generated reference uses skill-build.json; curator-build.json is rejected with no compatibility alias
- [x] Receipt, marker, cache and candidate identities bind the neutral descriptor while schemas 1-6 remain byte-stable
- [x] Deterministic regeneration, compatibility and release gates pass with task-scoped outcome evidence
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
Execution base: continue from the accepted curator-spec worktree /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree. Preserve its accepted uncommitted candidate, do not stage, commit, publish or mutate predecessor worktrees. This is an unreleased schema-7 rename: skill-build.json becomes the sole name and curator-build.json must have no alias.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-29c355, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-29c355)
Descriptor renamed to manager-neutral skill-build.json across normative prose, schemas, registry entries, generated cases, vectors, docs, release metadata and candidate digest. No alias: the retired literal appears nowhere in the tree (guards assemble the stem from parts so they scan their own source). Frozen schema-6 curator-build-source-v1/-v2 digest algorithm deliberately untouched - it is bound into the rc.4 byte-frozen install-marker-v2 valid case. New candidate pin sha256:9ba9b8ec... (predecessor sha256:58f8d229...). External cache key 07dd911a -> 4abc903b, a miss not an alias; local go-v1 identities byte-identical. All 9 rc.4 byte-frozen digests re-verified. Gates green standalone: generate 0, validate 0 (42 schemas / 422 files), python unittest 0 (29 tests), go test 0, go vet 0, gofmt 0, compileall 0, git diff --check 0. Clean probe: regenerate-check x2 = 0, release-check VERSION=1.0.0-rc.5 = 0. Three negative probes proved the Go guard, the Python validator and the release gate each reject the retired name. Deliberately not added: absent-descriptor hard-failure rule (no diagnostic-table row exists; beyond a rename). Note for reviewers: zb2s4z_results.md states the predecessor pin as bfe49f25 - that is the pre-rework value; the accepted on-disk predecessor pins 58f8d229, matching zb2s4z_rework-1-results.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260728-29c355, pid=92627, exit=0)
Reviewer focus: independently verify that skill-build.json is the sole repository-root descriptor and the retired curator-build stem/path/schema/case id has no alias or surviving reference outside the byte-frozen curator-build-source-v1/v2 algorithm namespace; schema-1 descriptor semantics and ownership are unchanged; all schema 1-6 frozen artifacts are byte-identical; regenerate twice and independently recompute the rc.5 conformance manifest digest/pins; run Python, Go, formatting, compatibility and clean-checkout release gates; do not stage, commit, publish, advance a release pin, or claim platform validation.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-e07352, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-e07352)
Reviewer verdict: accepted. Independent validation, Python/Go tests, frozen-artifact comparisons, neutral-name rejection guards, deterministic regeneration, digest recomputation and clean rc.5 release gates all passed. Evidence: TASK-260728-2kp3tv_review-verdict.md. No commit, release, pin advancement or platform claim was made.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-e07352, pid=3316, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260728-2kp3tv_spawn-log_-implementer--developer--claude-_RUN-260728-29c355.log](file://TASK-260728-2kp3tv/TASK-260728-2kp3tv_spawn-log_-implementer--developer--claude-_RUN-260728-29c355.log) — System spawn log captured by task-board
- [TASK-260728-2kp3tv_results.md](file://TASK-260728-2kp3tv/TASK-260728-2kp3tv_results.md) — Developer handoff evidence: manager-neutral skill-build.json descriptor rename across the unreleased schema-7/rc.5 candidate, guards, identities and gate exit codes
- [TASK-260728-2kp3tv_spawn-log_-reviewer--reviewer--codex-_RUN-260728-e07352.log](file://TASK-260728-2kp3tv/TASK-260728-2kp3tv_spawn-log_-reviewer--reviewer--codex-_RUN-260728-e07352.log) — System spawn log captured by task-board
- [TASK-260728-2kp3tv_review-verdict.md](file://TASK-260728-2kp3tv/TASK-260728-2kp3tv_review-verdict.md) — Reviewer acceptance evidence for the skill-build.json descriptor rename

## Created
2026-07-28T09:07:03Z

## Last Update
2026-07-28T17:00:00Z

## Assigned To
[reviewer] reviewer (codex)
