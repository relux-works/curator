## Status
to-dev

## Review
required

## Task Class
code

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [ ] One exported discriminator matches the landed schema and never probes the filesystem
- [ ] A path overlay is declarable from the config reader, resolveOverlay and profile compose add
- [ ] A path overlay carrying range, tag, branch, revision or directory stays profile_source_invalid
- [ ] profile install operand classification uses the same helper with no silent kind change
- [ ] The stale bound and ledger rows 303-304 are retired and truthful
- [ ] The candidate lane against curator-spec main with CI_REQUIRE_FULL_ROOT=1 is green on all three runners
- [ ] Every new refusal is driven through run() and killed by a narrowing mutant
- [ ] The two-root gate table is run sequentially with observed exit codes
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [ ] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [ ] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [ ] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn agent resolution: Agent selection: muse via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=muse; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (muse) (run=RUN-260906-339c1c, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260906-339c1c)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-339c1c, pid=72365, exit=1)
spawn autonomous recovery: run RUN-260906-339c1c queued successor RUN-260906-02c4c9 (attempt 1/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-02c4c9)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-02c4c9, pid=72764, exit=1)
spawn autonomous recovery: run RUN-260906-02c4c9 queued successor RUN-260906-eab036 (attempt 2/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-eab036)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-eab036, pid=72972, exit=1)
spawn autonomous recovery: run RUN-260906-eab036 queued successor RUN-260906-b966cb (attempt 3/3, model=muse-spark-1.3-contributor): spawned agent exited with code 1
spawn run started: [implementer] developer (muse) (run=RUN-260906-b966cb)
agent completed: [implementer] developer (muse) (exit=1)
spawn run completed: muse (run=RUN-260906-b966cb, pid=73172, exit=1)
recovery parked after 3 successor attempts for chain RUN-260906-339c1c; operator action required; last failure: spawned agent exited with code 1

## Precondition Resources
- [producer-brief-consume-overlay-rule.md](file://TASK-260906-19gjyw/producer-brief-consume-overlay-rule.md) — Producer brief: port the git-source-only overlay rule into the Go reader and make the candidate lane green against spec main

## Outcome Resources
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-339c1c.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-339c1c.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-02c4c9.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-02c4c9.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-eab036.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-eab036.log) — System spawn log captured by task-board
- [TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-b966cb.log](file://TASK-260906-19gjyw/TASK-260906-19gjyw_spawn-log_-implementer--developer--muse-_RUN-260906-b966cb.log) — System spawn log captured by task-board

## Created
2026-09-06T19:20:27Z

## Last Update
2026-09-06T22:51:42Z

## Assigned To
[implementer] developer (muse)
