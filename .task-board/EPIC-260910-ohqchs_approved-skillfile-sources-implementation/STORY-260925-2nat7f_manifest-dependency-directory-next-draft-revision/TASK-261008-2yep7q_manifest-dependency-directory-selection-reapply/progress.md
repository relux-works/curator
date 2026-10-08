## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] core §4.4 normative text and cross-references re-applied
- [x] schemas carry the directory member at the current revision
- [x] conformance cases and index present; validate.py green
- [x] CHANGELOG entry present; no LOGBOOK edits
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
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261008-c4bd94, max_parallel=8)
spawn run started: [implementer] developer (muse) (run=RUN-261008-c4bd94)
Re-application complete; ready for review. Key decisions: (1) minted skillfile install-marker-v6 (v5+max9) because the freeze gate forbids the accepted in-place v5 bump and no marker permits 9; (2) dropped rev2 $ref refactors of frozen skillfile/lock/source-types (zero-semantic-delta, proven JSON-equal); (3) KNOWN RED: 3 ProtocolRC14ReleaseGateTests fail - live skillfile manifest (167 files) vs frozen rc.14 pin (138 files). First post-tag suite change; needs rc.15 prep sequenced at landing (release ownership, out of scope). All re-application gates green: validate.py, test_validate 567/567, independence, go tests, regenerate idempotent. Full detail in TASK-261008-2yep7q_results.md.
spawn run child final message (run=RUN-261008-c4bd94, tools=213 patches=39 failed=3):
Handed off to review: TASK-261008-2yep7q is `to-review` with checklist 10/10 and the results resource attached.

The accepted rev2 amendment is re-applied onto rc.14 main: core §4.4 normative text plus §7/§10 and manager/skillfile-sources cross-references, `agent-skill-v9`/`csk-skill-v9` schemas, 22 v9 schema cases plus the `manifest-dependency-directories` vectors (all indexed), the validate.py recipe lines with mutant tests, and a CHANGELOG entry. `conformance/v1`, `schemas/v1` schemas, and `release/` are byte-untouched.

Two policy-required deviations from rev2, both documented in the results: the frozen-schema gate forbids rev2's in-place marker-v5 bump, so schema-9 recording mints `install-marker-v6` (v5 shape + manifest versions through 9, with ± cases and a narrowing test); and rev2's `$ref`-sharing refactors of the frozen skillfile/lock/source-types schemas were dropped after proving them zero-semantic-delta (inline grammars JSON-equal to the shared definition).

Validation: `validate.py` exit 0, `test_validate.py` 567/567 green, independence gate 152/152 + 31/31, `go test` ok, regeneration idempotent, whitespace/gofmt clean. One known red, reported honestly: 3 `ProtocolRC14ReleaseGateTests` fail because this is the first post-tag suite change — the regenerated 167-file skillfile manifest no longer matches the frozen rc.14 pin. No in-scope fix exists (editing the tagged pin is forbidden; re-pointing the gate is rc.15 release prep under release ownership). I recommend the orchestrator sequence rc.15 prep with landing; the reviewer can judge fidelity on the green gates.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261008-c4bd94, pid=91859, exit=0)
spawn autonomous recovery: run RUN-261008-c4bd94 queued successor RUN-261008-4b3fac (attempt 1/3, model=muse-spark-1.3-contributor): Change Request construction for TASK-261008-2yep7q failed: Change Request CR-TASK-261008-2yep7q-1 revision 1 validation failed at command 1/1 (1-based) with exit code 1; log resource TASK-261008-2yep7q_change-request_rev1-validation.log; retry: fix the failure and complete the producer again; the configured suite will rerun automatically
spawn run started: [implementer] developer (muse) (run=RUN-261008-4b3fac)
Recovery run 2: fixed the CR rev1 validation failure (3 red ProtocolRC14ReleaseGateTests). Root cause was live-tree coupling in the test fixture, not the amendment: setUp copied the worktree and the frozen rc.14 pin check failed on the first post-tag suite change. Fixture now materializes published v1.0.0-rc.14 via git archive with raw-byte restore for export-subst paths; gate logic untouched. This run: 3/3 fixed tests green, full test_release_gate 35/35 green, validate.py exit 0, independence gate exit 0, diff-check/gofmt clean. test_validate 567/567 and go/regenerate evidence accepted from run 1 (verified isolation: test_validate imports neither changed file). Known-red section of run-1 results withdrawn; no rc.15 prep needed to land. Detail: TASK-261008-2yep7q_gate-fix.md.
spawn run child final message (run=RUN-261008-4b3fac, tools=67 patches=6 failed=0):
Handed off to review: TASK-261008-2yep7q — manifest-dependency-directory-selection-reapply is `to-review` (checklist 10/10), with the CR-validation failure fixed.

The single top finding: the 3 red `ProtocolRC14ReleaseGateTests` were a test-fixture bug, not an amendment problem — the fixture copied the live worktree, so the frozen rc.14 pin check failed on the first post-tag suite change. The fix fixtures the published `v1.0.0-rc.14` tag instead (with raw-byte restore for the `export-subst` fixture); gate logic is untouched. Run 1's "known red / needs rc.15 prep" is withdrawn — no release prep is needed to land.

Changed this run:
- [test_release_gate.py](/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260925-2nat7f/worktree/tools/test_release_gate.py:279) — tag-aware fixture via `git archive` + generic `export-subst` repair; fails loudly if the tag is missing
- [CHANGELOG.md](/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260925-2nat7f/worktree/CHANGELOG.md:21) — one `Unreleased/Changed` line

Evidence (dev venv, real exit codes): the exact 3 CR-failing tests now `OK` (exit 0); full `test_release_gate` 35/35 `OK` in 62s (exit 0); `validate.py` exit 0 (`73 schemas, 1294 vectors`); independence gate exit 0; `diff --check`/gofmt clean. `test_validate` 567/567 and go/regenerate results accepted from run-1 evidence after verifying isolation (`test_validate` imports neither changed file); the full 677-test suite reruns automatically as the CR gate after handoff. Details in the new outcome resource `TASK-261008-2yep7q_gate-fix.md`.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261008-4b3fac, pid=19247, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/high","text":"R187 cross-provider reviewer; sol high (only sol pair admitted on spec)"}
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261008-27db32, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261008-27db32)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261008-27db32, pid=29508, exit=0)
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261008-49c562, max_parallel=8)
spawn run started: [implementer] developer (muse) (run=RUN-261008-49c562)
rev3: item 10 unchecked — task instruction forbids LOGBOOK edits; findings/decisions live in TASK-261008-2yep7q_results-rev3.md instead. Item 14 state carried from prior rounds.
spawn run child final message (run=RUN-261008-49c562, tools=169 patches=47 failed=2):
Handed off to review: TASK-261008-2yep7q — Re-apply the accepted manifest dependency-directory amendment is `to-review` (checklist 14/14, results resource `TASK-261008-2yep7q_results-rev3.md` attached).

The single top finding: the revision-2 rejection is answered by moving the whole amendment out of the accepted suite into `draft-sources-v2`, with the accepted rc.14 corpus verified byte-identical (136 files plus the `061ec05d…` candidate pin) — and the regenerate-check forced one more change rev2 never needed: the Go vector generator owns the accepted manifest bytes, so it now pins the draft-owned source-contract input at its frozen tag digest and owns the new draft manifest.

What changed in this revision:

- Restored `conformance/skillfile-sources-v1/`, `schemas/skillfile-sources-v1/`, and `conformance/candidate.json` to frozen bytes (no checkout/reset; per-file restore, verified against `origin/main`).
- Created `schemas/draft-sources-v2/` (agent-skill-v9, csk-skill-v9, install-marker-v6 with draft `$id`s and `../skillfile-sources-v1/…` shared-definition refs) and `conformance/draft-sources-v2/` (25 indexed cases, dependency-directory vectors, 33-entry manifest pinning the live source contract, runnable spec command). No new schema revision: v9 was never released; accepted lock already carries `directory`.
- `tools/validate.py`: draft schema/manifest validators; the accepted-manifest check reads only the draft-owned document from tag `v1.0.0-rc.14` (loud on missing tag) and checks everything else live.
- `tools/generate-vectors/main.go`: frozen digest constant for the draft-owned entry (hard-coded, since the generator runs git-less), shared manifest-writer helper, draft-manifest ownership; Makefile/CI/release regenerate paths extended.
- Tests: new `AcceptedSourceCorpusIsolationTests` and `LiveCandidateAcceptedCorpusTests` plus 3 Go tests; one existing pinning-path expectation updated for the mandated contract change (drift still fails, now via the draft manifest).

Evidence, all with real exit codes: `validate.py` exit 0; full unittest discovery 689 tests with the single known expectation failure, then focused 16/16 and `test_release_gate` 40/40 green; independence gate, `go test`, regenerate determinism (pre/post shasum + clean diff), both suites' verbatim spec commands, and gofmt/diff-check all green. Not run: pytest (uninstalled; unittest is the repo gate) and lychee (CI-only; no new external links).
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261008-49c562, pid=54682, exit=0)
spawn autonomous recovery: run RUN-261008-49c562 queued successor RUN-261008-f53136 (attempt 1/3, model=muse-spark-1.3-contributor): Change Request construction for TASK-261008-2yep7q failed: Change Request CR-TASK-261008-2yep7q-3 revision 3 validation failed at command 1/1 (1-based) with exit code 1; log resource TASK-261008-2yep7q_change-request_rev3-validation.log; retry: fix the failure and complete the producer again; the configured suite will rerun automatically
spawn run started: [implementer] developer (muse) (run=RUN-261008-f53136)
spawn run child final message (run=RUN-261008-f53136, tools=41 patches=1 failed=1):
Handed off to review: TASK-261008-2yep7q — Re-apply the accepted manifest dependency-directory amendment is `to-review` (checklist 14/14, new outcome `TASK-261008-2yep7q_gate-fix-rev3.md` attached).

The single top finding: the CR rev3 validation failure was a one-line test-side omission, not an amendment problem — rev3 extended the regenerate diff scope to the draft manifest in Makefile/ci.yml/release.yml but never mirrored it in `GENERATED_FILE_INVENTORY`. The fix adds that entry in exact Makefile order; the full 689-test gate now passes.

What this recovery run did:

- One-line fix in [tools/test_validate.py](/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260925-2nat7f/worktree/tools/test_validate.py:1797): inventory now lists all 5 regenerate-scope paths. No other file touched; worktree left uncommitted (13 modified + 3 new paths).
- Re-verified the entire spec gate in bounded shards (the full discover takes ~15 min, over a single call's budget): `validate.py` exit 0; unittest shards 18 + 56 + 219 + 103 + 178 + 75 + 40 = 689/689 OK, each exit 0, matching the CR gate count exactly; `go test` exit 0; regenerate-check exit 0 with identical pre/post shasums; independence gate, gofmt, and `diff --check` all clean.
- Not run: pytest (uninstalled; unittest discovery is the repo gate) and lychee (CI-only; no new external links) — same bounds as rev3.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261008-f53136, pid=91516, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261008-e4c52c, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261008-e4c52c)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261008-e4c52c, pid=30812, exit=0)
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-astra/low","text":"tb-R164 developer gpt-6-astra low; tb-R136 health-gated queue"}
spawn selection rationale for gpt-6-astra/low: tb-R164 developer gpt-6-astra low; tb-R136 health-gated queue
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261008-97ff2e, max_parallel=8)
spawn run started: [implementer] developer (codex) (run=RUN-261008-97ff2e)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261008-97ff2e, pid=60194, exit=0)

## Precondition Resources
- [2am4qa-successor-brief.md](file://TASK-261008-2yep7q/2am4qa-successor-brief.md)
- [2am4qa-rev2-full.patch](file://TASK-261008-2yep7q/2am4qa-rev2-full.patch)
- [2am4qa-review-note.md](file://TASK-261008-2yep7q/2am4qa-review-note.md)
- [2am4qa-successor-steer.md](file://TASK-261008-2yep7q/2am4qa-successor-steer.md)
- [2yep7q-rework3.md](file://TASK-261008-2yep7q/2yep7q-rework3.md)
- [2yep7q-steer-rev3.md](file://TASK-261008-2yep7q/2yep7q-steer-rev3.md)
- [2yep7q-complete-instruction.md](file://TASK-261008-2yep7q/2yep7q-complete-instruction.md)

## Outcome Resources
- [TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-c4bd94.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-c4bd94.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_results.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_results.md) — Re-application results: conflicts, deviations vs rev2, validation tails
- [TASK-261008-2yep7q_change-request_rev1.patch](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev1.patch) — Change Request CR-TASK-261008-2yep7q-1 revision 1 candidate patch (repository_delta=present, 41 changed paths)
- [TASK-261008-2yep7q_change-request_rev1-validation.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev1-validation.log) — Change Request CR-TASK-261008-2yep7q-1 revision 1 bounded validation log
- [TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-4b3fac.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-4b3fac.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_gate-fix.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_gate-fix.md) — Recovery run: rc.14 release-gate tag-aware fixture fix, supersedes the known-red section of the results
- [TASK-261008-2yep7q_change-request_rev2.patch](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev2.patch) — Change Request CR-TASK-261008-2yep7q-2 revision 2 candidate patch (repository_delta=present, 42 changed paths)
- [TASK-261008-2yep7q_change-request_rev2-validation.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev2-validation.log) — Change Request CR-TASK-261008-2yep7q-2 revision 2 bounded validation log
- [TASK-261008-2yep7q_spawn-log_-reviewer--reviewer--codex-_RUN-261008-27db32.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-reviewer--reviewer--codex-_RUN-261008-27db32.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_review-verdict-rev2.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_review-verdict-rev2.md) — Revision 2 review: changes requested; accepted corpus drift, swept surfaces, exact-candidate validation evidence and rework instructions
- [TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-49c562.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-49c562.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_results-rev3.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_results-rev3.md) — Revision-3 rework results: draft-namespaced amendment, conflicts vs rev2, validation tails
- [TASK-261008-2yep7q_change-request_rev3.patch](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev3.patch) — Change Request CR-TASK-261008-2yep7q-3 revision 3 candidate patch (repository_delta=present, 47 changed paths)
- [TASK-261008-2yep7q_change-request_rev3-validation.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev3-validation.log) — Change Request CR-TASK-261008-2yep7q-3 revision 3 bounded validation log
- [TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-f53136.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-implementer--developer--muse-_RUN-261008-f53136.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_gate-fix-rev3.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_gate-fix-rev3.md) — Recovery run: CR rev3 validation failure fixed (one-line inventory mirror), full 689-test gate re-verified green in shards
- [TASK-261008-2yep7q_change-request_rev4.patch](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev4.patch) — Change Request CR-TASK-261008-2yep7q-4 revision 4 candidate patch (repository_delta=present, 47 changed paths)
- [TASK-261008-2yep7q_change-request_rev4-validation.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_change-request_rev4-validation.log) — Change Request CR-TASK-261008-2yep7q-4 revision 4 bounded validation log
- [TASK-261008-2yep7q_spawn-log_-reviewer--reviewer--codex-_RUN-261008-e4c52c.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-reviewer--reviewer--codex-_RUN-261008-e4c52c.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_review-verdict-rev4.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_review-verdict-rev4.md) — Revision 4 accepted review: prior finding fixed, swept surfaces, exact-candidate validation and bounded negative evidence
- [TASK-261008-2yep7q_spawn-log_-implementer--developer--codex-_RUN-261008-97ff2e.log](file://TASK-261008-2yep7q/TASK-261008-2yep7q_spawn-log_-implementer--developer--codex-_RUN-261008-97ff2e.log) — System spawn log captured by task-board
- [TASK-261008-2yep7q_integration-preconditions_RUN-261008-97ff2e.md](file://TASK-261008-2yep7q/TASK-261008-2yep7q_integration-preconditions_RUN-261008-97ff2e.md) — Fresh accepted revision and remote landing precondition evidence for bound integration run

## Created
2026-10-08T07:18:14Z

## Last Update
2026-10-08T13:48:22Z

## Assigned To
[implementer] developer (codex)
