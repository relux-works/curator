# TASK-260720-12r55p reviewer verdict — cycle 2 — CHANGES REQUESTED

Reviewed CocoaSkills PR 19 at exact signed head `ba250bfc4dfe104a160eadd5b5f4e340693bf892` over exact signed base `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`.

## Verdict

CHANGES REQUESTED. Route to `to-dev` for ordinary conformance-harness rework and require another independent reviewer cycle. This is not a stop-the-line boundary.

## Blocking findings

### P1 — build-driver rejection bindings still manufacture expected outcomes instead of observing the corresponding CocoaSkills boundary

The new exhaustive name table closes unknown names, but it remains the source of both the boundary and expected error. A fresh read-only mutation audit changed `input.condition` in every rejection vector that carries it: all 75 of 75 altered cases still passed. More importantly, behavioral sabotage probes showed that the adapter can pass when the product seam does the wrong thing:

- `unknown-driver` passed when `skillspec.load_skill_spec` was replaced with an unrelated `SkillSpecError`; `_probe_skill_spec_rejection` only requires some exception, and `assert_build_rejection_case` retains the hard-coded expected code.
- `wrong-go-executable-path` passed when `toolchain._select_toolchain` returned `untrusted_go_executable`; lines 660–663 explicitly accept the wrong code and then return `toolchain_executable_mismatch`.
- `artifact-hash-mismatch` passed while `cache.cache_for_manager_home` was patched to fail if backend inspection occurred. The adapter verifies an unchanged valid receipt, creates a generic corrupt `CacheInspection`, and returns the table value without inspecting a mismatched artifact.
- The same hard-coded return pattern remains for multiple process, toolchain, cache, and context cases at `tests/protocol_conformance_adapters.py:602-610`, `:691-714`, and `:719-751`.

This does not satisfy behavior-sensitive coverage of all 77 rejection cases or the requirement to consume vector inputs without replacing them with hard-coded fixtures/results.

Required repair: make each name binding execute the relevant CocoaSkills validator/backend with the case's concrete vector data where published and independently materialized condition fixtures otherwise. Derive the observed error/status from that execution, assert the vector error/result/reuse/no-execution fields against the observed trace, and remove paths that return `_BUILD_REJECTION_BINDINGS[name][1]` without observing that exact product outcome. Add exhaustive mutation/sabotage coverage so an unrelated exception, wrong product error, omitted cache inspection, or unbound condition cannot pass.

### P1 — manager lifecycle remains predominantly a self-consistency reader, not implementation conformance

`_LIFECYCLE_CASE_FIELDS` verifies exact top-level field presence, but `assert_manager_lifecycle_case` mostly compares vector fields to copied literals or to other fields in the same vector. A fresh one-leaf-at-a-time audit over all 32 cases changed 352 scalar leaves: 104 mutations still passed.

Representative unbound normative values include:

- every `compiled-cache-miss-is-read-only.forbidden_persistent_effects[*]`;
- every `all-source-and-trust-gates-before-build.required_before_toolchain_or_cache[*]`;
- the first three private-build failure `events` and most `forbidden_effects`;
- `interrupted-global-journal-recovered-by-transaction-id.journal_transaction_id` and `triggering_project`;
- every repair `independent_conditions[*]` and `required_pipeline[0:9]`;
- the first seven currentness `validated[*]` fields and every failure-matrix condition;
- all six `deterministic-target-order-and-consumer-last.expected_commit_order[*]` values;
- `compiled-currentness-failure-matrix.result`.

The adapter imports no transaction engine, installer/dry-run planner, build-currentness/status implementation, repair pipeline, or launcher implementation for these clusters. The two rollback booleans now fail mutation, but that proves only literal assertions, not rollback behavior.

Required repair: bind every one of the 32 cases to the corresponding CocoaSkills seam or reusable existing test helper and compare an observed trace/state transition to every normative vector field. Cover bootstrap, publication, cross-project state, dry-run, GC, launch, planning, private builds, recovery, repair, status, commit, rollback, and upgrade. Add an exhaustive field-mutation test or an explicit reviewed normative-field classification so future fields cannot silently become metadata-only.

## Revalidation of the other prior findings

- Fixed environment and argv: repaired. All 28 environment values and all 60 scalar leaves across the five argv/cwd/name/source-awareness records failed independent mutation, and the adapter captures the real three parent probes plus worker plan.
- Manifest authentication: repaired for the in-scope candidate inventory. Required vector, selected schema-case, go-build fixture, and 11 build-driver artifact bytes are admitted through manifest membership and digest checks; mutated bytes fail.
- Build-source/toolchain named seams: the previously cited link, special-file, mutation, missing-LF, multiple-LF, resolution, escape, dangling, framing, and fingerprint checks now execute CocoaSkills validators. These repairs do not cure the rejection and lifecycle blockers above.
- Windows projection: repaired. No new skip, xfail, `os.name`, or platform bypass exists; native links are target-first, native `readlink` must denote the exact protocol target, and real fingerprint/resolution validators still run.

## Passing evidence

- Exact-root local gate: `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 .venv/bin/python -m pytest -q tests/test_protocol_conformance.py` — exit 0, 452 passed.
- Strict mypy: `.venv/bin/python -m mypy` — exit 0, no issues in 68 source files.
- Hosted run `30691018727` is exact head `ba250bfc4...`, conclusion success, 14/14 jobs green across Ubuntu/macOS/Windows Python 3.11–3.14, strict mypy, and build artifacts.
- PR 19 is open, mergeable, clean, exact head `ba250bfc4...`, exact base `dacccaaf...`.
- All four task commits have good signatures; task worktree is clean; `.github` has no diff.
- Curator-spec commit `432eb2ee1fe2d6b271e37269f867c8851c325539` is GitHub-verified; manifest digest is `sha256:12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`; the rc.6 release record matches and remains candidate-only.
- `0c81c1f8..432eb2ee` is byte-identical for `conformance` and `release/1.0.0-rc.6.json`; no workflow pin, tag, GitHub Release, released-suite pin, or conformance claim changed.

Green gates establish stability, but the mutation/sabotage evidence proves that explicit acceptance criteria remain unenforced.