# TASK-260916-3l60rn — review verdict, revision 1

Verdict: **changes_requested**. Route to `to-dev`; do not accept CR revision 1.

Reviewed the exact curator-spec Story worktree at base `e8b53a003256433761cebce6080d6a955d777f25`, the producer brief/rules, producer evidence and patch, and security audit E6 / Appendix B. No candidate code, index, or LOGBOOK was edited. All probes modified Python objects in memory or used a disposable `/tmp` copy. The run is not goal-bound (`task-board spawn goal "$TASK_BOARD_RUN_ID"`: `Active Goal: none`).

## Findings requiring correction

### R1 — inconsistent dry-run outcome for the new path-directory failure class

`protocol/environments.md:766` calls the new failure **“entry-class `environment_store_untrusted`”**, and :769 says **“There is no rebuild”**. However §10.1 :2469–2473 still says **“Dry-run evaluation of an entry-class failure reports `would-rebuild-untrusted-store`”**, and the §10.4 table at :2691 explicitly describes that diagnostic as **“a real operation would rebuild it”**. The new entry-class member therefore enters a row that promises the action §4 forbids. The E6 vectors carry `rebuild_planned: false`, but none identifies a dry-run operation and resolves this contradiction.

Correction: explicitly exclude path-source-directory failures from the rebuild dry-run branch in §4/§10.1/§10.4, state their diagnostic and no-rebuild result consistently, and add a pinned dry-run conformance case with a negative replacement test. Preserve the settled operator-repair rule and the existing store-entry rules.

### R2 — boundary gate can narrow to system-module carriers without a failing test

The settled rule applies to directories regardless of content class. Yet both no-system cases (`conformance/v1/vectors/environments-path-kind-admission.json:107` and :134) have all five checks true; all five boundary-failure cases carry system modules. Consequently this in-memory narrowing of the actual validator survives:

```python
# tools/validate.py:6657, inside validate_environments_path_kind_admission_vectors
trusted, failing = e6_boundary_trusted(inputs) if case["carries_system_modules"] else (True, [])
```

The same narrowed function was also invoked through the actual `validate.main()` entry point: `validated 62 schemas and 1096 vector files`, `NARROWED_MAIN_EXIT: 0`. Independent result: all **21/21** `PathKindAdmissionVectorTests` still pass. This is a narrowing probe, not deletion of the gate. Published coverage is 5/5 individual failure checks for system-module overlays, but **0/2 required no-system entry classes (overlay and onboarding import) have boundary-refusal cases**. The no-system positive cases prove admission only. A manager that skips the directory check for these inputs is not distinguished by this corpus.

Correction: add pinned refusal scenarios for an untrusted no-system overlay and an untrusted no-system onboarding import, plus corresponding trusted controls and rule-7 replacement tests. Ensure the narrowing above fails the production validator and its regression suite. This does not require changing manager implementation.

## Per-deliverable review

Paths/lines below refer to the candidate curator-spec tree.

| Requirement | Evidence / result |
| --- | --- |
| Patch equals worktree | Both `git diff HEAD | git patch-id --stable` and the downloaded patch yield `a94da8807edb592c87a9773475d0c009dad1c896`. Worktree diff SHA-256 `48b99eb3137e94c652147da5f18d3fdb784cce53576b31d3ecbd53c76bf0c881`. No comparison with moving origin/main. |
| Dated Decision 0012 amendment | `decisions/0012-context-packages-and-semver-locks.md:127`: `Amendment (2026-09-18, E6)`; three amendment items and original-passage markers. Preserves historical body rather than silently rewriting it. |
| Git-only MCP; hard error naming package/declaration | `protocol/environments.md:526`: “An MCP declaration package MUST resolve from a `git` source”; :529–534 “MUST NOT carry”, new diagnostic, “never admitted, never warned-through”. §2.1 table :456 includes the identical spelling. |
| E1 rationale | §2.2 :534–537: “no canonical identity” / “never verified” / “allowlist is total over canonical identities”. |
| Path system modules require direct naming AND boundary | §3 :636–651 applies E2 “as is” and requires the contract “at every resolve and before any materialization”. Impossible transitive resolution is explicitly identified; synthetic transitive admission case preserves E2 error policy. |
| Five checks and integrity baseline | §4 :738–759 enumerates ownership, permissions/DACL, containment, regular types, link safety; :761–765 preserves store-entry state pin and does not rehash/adopt live directory bytes. |
| No fragment / non-current / no rebuild | §4 :766–771 states all three, but R1 requires aligning dry-run reporting. |
| Overlays/imports regardless of content | §3 :650–652, §6 :1120–1126 and §9.6 :2371–2374 state directory checking and MCP prohibition. R2: missing refusal evidence for no-system overlay/import. |
| Diagnostics closed and identical | New spelling `mcp_declaration_path_source_refused` agrees across Decision, §2.1/§2.2/§13, CHANGELOG, vectors and validator. Reuses `environment_store_untrusted` and `context_system_module_transitive`. §1.1 has no new source-parser diagnostic; §3.1 already tables E2. No new knob/schema/lock key; §12.1/§12.2 need no new admission key. R1 is an outcome inconsistency, not spelling drift. |
| Posture | §12 :2877–2884 adds source directory, path and failing check to store-trust row. |
| Direct rollout | §4 :773–775: “not configurable” and extension rollout “direct (impact row E6: under the hood)”; CHANGELOG.md:10 names E6. No warn-first requirement applies. |
| Required vectors and manifest | New family contains git MCP control, three path MCP refusals, admitted system overlay, five distinct boundary failures, synthetic transitive case, and three non-conforming observations. Manifest :4287 registers its digest. All named existing branches have discriminating input pins in `tools/validate.py:6479–6498`; replacement/narrowing regressions in `tools/test_validate.py:2232` onward pass. R2 describes the missing content-class branch. |
| Existing vectors preserved | Independently compared HEAD bytes for all 1,095 pre-existing conformance files other than manifest: 1,095/1,095 identical. |
| Scope | Eight changed files: Decision, protocol, CHANGELOG, new vector, manifest, generated rc.9 manifest pins, validator and its tests. No manager implementation, schema, LOGBOOK, or proposals 0014–0018 changed. Test/validator changes are explicitly required by campaign rule 7. |
| Empty curator CR | Verified `git diff --exit-code 640a9df19ba94295cce059562101599996761f4b 57d2578ea9d1644d8a8931fdcdfbcb8ce4e15545` exits 0. This is expected: the deliverable is in curator-spec and attached as the matching spec patch. Empty curator delta itself is not a rejection reason; the two findings concern the actual spec deliverable. No merge/landing is claimed. |

## Independent validation and bounded-run accounting

Shell for requested make command: `/bin/bash`, `set -o pipefail`, spec Story worktree:

```sh
PATH="/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH" make validate
```

Observed:

```text
python3 tools/validate.py
validated 62 schemas and 1096 vector files
python3 -B -m unittest discover -s tools -p 'test_*.py'
```

The complete make invocation was intentionally interrupted after substantive rework findings, during `test_release_gate` (KeyboardInterrupt; make exit 2), and replaced with focused validation. **No independent full-suite pass is claimed.** The producer reports 428 tests / 874.937 seconds / exit 0; that remains producer evidence, not a reviewer replay. No test assertion failure had been observed at interruption. A subsequent broader `python3 -B -m unittest discover -s tools -p test_validate.py` run was also stopped after approximately 3.5 minutes (exit 130, during `test_require_source_signers_default_drifting_from_the_table_fails`); it is incomplete and contributes no full-suite pass claim. The completed 21-test E6 run, production-gate run, regeneration and Go test are the independent results relied on for this rejection. An initial focused invocation accidentally used the curator working directory and failed discovery; rerun in curator-spec is the result below.

```text
python3 -B -m unittest discover -s tools -p test_validate.py -k PathKindAdmissionVectorTests
Ran 21 tests in 0.012s
OK
exit 0

go test ./tools/...
ok github.com/relux-works/curator-spec/tools/generate-vectors 3.113s
exit 0
```

Regeneration ran in a disposable copy of the candidate's tracked files so the reviewer did not write generated files into the candidate:

```text
make regenerate
go run ./tools/generate-vectors -root .
exit 0
Regeneration byte-equivalence: 1102 / 1102
Changed: []
```

This is a candidate-relative regeneration proof. The additional exact `make regenerate-check` execution uses the disposable copy as `GIT_WORK_TREE` and an isolated `/tmp/TASK-260916-3l60rn-regen.index` populated from the candidate; no candidate index or files are modified. This checks generation against the reviewed candidate rather than treating its intended uncommitted additions as drift. Exact target result:

```text
make regenerate-check
go run ./tools/generate-vectors -root .
git diff --exit-code -- conformance/v1 release/1.0.0-rc.5.json release/1.0.0-rc.6.json release/1.0.0-rc.7.json release/1.0.0-rc.8.json release/1.0.0-rc.9.json
exit 0
```

The in-memory narrowing probe printed:

```text
Ran 21 tests in 0.010s
OK
NARROWED boundary gate skips no-system sources; suite green: True
exit 0
```

## Handoff

Ordinary spec/conformance rework, no external blocker or human decision. Address R1 and R2, regenerate, attach the revised patch/evidence and run bounded complete validation before a new reviewer cycle. Findings are recorded here and on the board; campaign prohibition on LOGBOOK edits takes precedence over the generic logbook instruction.
