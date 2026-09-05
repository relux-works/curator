# TASK-260905-30a1bz — review findings, ax PR #1 revision, cycle 1

Subject: `relux-works/agent-session-manager-spec` PR #1, branch
`draft/curator-environment-integration`, head `2c7f642` (commits `ef91985`, `2c7f642` on
top of `d7075e1`, base main `28bf96d`). Authority: curator-spec Decision 0013 at `83de1a5`.
Reviewed diff: `git diff d7075e1..2c7f642`. Reviewer worktree access read-only; scratch and
mutant copies under the ax worktree `.temp/review/`.

## Verdict: CHANGES REQUESTED (two majors, one minor). Route: `development`.

The Change Request on the curator-spec side (`CR-TASK-260905-30a1bz-1`, `repository_delta=empty`)
is correct as an empty delta: this leaf's deliverable lives in the ax repository, and the
brief forbids writes into curator-spec. The emptiness is not a finding. The findings below
are against the ax branch head `2c7f642`.

## Gates rerun by the reviewer (ax worktree, head `2c7f642`)

| Command | Exit | Evidence |
| --- | ---: | --- |
| `./scripts/validate_spec.py` | 0 | 286/286; launch-plan ledger gate_classes=6, positive=3, negative=15 (`.temp/review/validate.log`) |
| `./run_validation.sh` | 0 | structurizr-cli + plantuml present; "Validation successful" (`.temp/review/run-validation.log`) |
| `./scripts/test_expected_red.sh` | 0 | 314 passed, 0 failed out of 314 mutations (`.temp/review/expected-red.log`) |
| `git diff --check` | 0 | clean |

Delivery shape verified: `git log --oneline main..HEAD` = `2c7f642`, `ef91985`, `d7075e1`
(no rewrite); both new commits carry `Good "git" signature for oparin@me.com`, author
`Ivan Oparin <oparin@me.com>`; `gh pr view 1 --json headRefOid` = `2c7f642` = local head =
`origin/draft/curator-environment-integration`; PR state OPEN, base `main`, not merged; PR
body cites Decision 0013 and proposes v0.6.0. Frozen map: only `SPEC.md` digest moved.
`VERSION`, `CHANGELOG.md`, `RELEASE_NOTES.md`, `README.md` untouched (diff empty).

## Decision 7 items, verified in SPEC text

| Item | Verified | Where |
| --- | --- | --- |
| 1 | yes — merge paragraph gone; "adds no member" withdrawn; `--launch-plan` operation described | §7.5 |
| 2 | yes — `stdin` row + Launch Stdin object (`utf-8`/`base64url`, ≤65,536 decoded, non-secret, absent/null = terminal); `system-modules` boolean; `profile-pin` = `sha256:` lock hash; Decision 8 pre/post distinguishability | §5.1 |
| 3 | yes — nine names in order; manifest/probe 1.1.0 consequence stated; `SpawnPlan.stdin`; `resume.launch_plan` (row + example); verbatim translation, no reorder/dedup/rewrite/second spelling; `capability_unavailable` `details.capability` before invocation | §7.3, §7.4, §7.5 |
| 4 | yes — planning-role `launch` before step 2 for both forms; final argv in `launch_plan.argv`; `ax.launch-plan-request` `{form, base_argv_length, request_digest}` with no suffix copy; residual bound → `launch_plan_invalid` `field: "extensions"`; determinism → `provider_protocol_error` | §13.1 |
| 5 | yes — refuse-on-drift when `system-modules: true` (`policy_refused`, exit 16, `environment_drift`); warn otherwise; strict mode kept; failed resolution distinct; like-with-like | §13.10 |
| 6 | yes — CCJ-1 canonical bytes of the parsed fragment, not pretty-printed output | §5.1 |
| 7 | yes — grammar row; `--task-board` exclusivity and `argv`+`--profile yolo` = `invalid_arguments`; document table; `launch_plan_invalid` exit class 2 with `details.field`; secrets → `secret_policy_violation` exit 16; profile-flag MUST rule keyed on §7.7 with `reason: "profile_flag"` and `details.argv_index`, "in `argv` or in `argv_suffix`"; `curator session` note kept | §14.1, §15.3 |
| 8 | partly — §1.5 row, §15.3 code, Appendix D.2 row and D.4 bullet, A.1 row and A.12 table citing Decision 0013, fixture with 3 positive and 15 negative cases; v0.6.0 in PR body only. See F1/F2 for the fixture defects | §1.5, §15.3, A.1, A.12, D.2, D.4, `fixtures/launch_plan_request_conformance.json` |

ax invariants: lease/fencing/checkpoint/materialization text untouched; §1.6 rules untouched;
§13.2 task-board launch untouched (only the grammar block gains a separate row); the three
original Curator keys kept.

## Gate attack (scratch copy of `scripts/`, `fixtures/`, `SPEC.md` under `.temp/review/mut/`, bytecode cache disabled)

Narrowing mutants applied to `scripts/validate_launch_plan.py` and the gate rerun:

| Mutant | Result |
| --- | --- |
| profile-flag check only at the first caller index | RED (PROFILE-FLAG-NEG admitted) |
| profile-flag check skipped for the `argv` form | RED (ARGV-FORM-NEG admitted) |
| profile-flag long form only, alias dropped | RED (ALIAS-NEG admitted) |
| `STDIN_MAX_BYTES` +1 | RED (STDIN-BOUND-NEG at 65,537 admitted) |
| capability check skipped for `argv` form | RED |
| `ax.launch-plan-request` collision unchecked | RED |
| env_literals/env_names disjointness dropped | RED |
| `argv` + `--profile yolo` admitted | RED |
| stdin secret check dropped | RED |
| unknown member admitted | RED |
| `EXTENSIONS_MAX_BYTES` +1 | **GREEN — survives** (F2) |
| `EXTENSIONS_MAX_BYTES` +203 | RED |
| `ARGV_TOTAL_MAX_BYTES` +1 | **GREEN — survives** (F3) |
| `ARGV_MAX_ELEMENTS` +1 | **GREEN — survives** (F3) |
| `ENV_MAX` +1 | **GREEN — survives** (F3) |
| `LITERAL_MAX_BYTES` +1 | **GREEN — survives** (F3) |
| `EXTENSIONS_MAX_KEYS` +1 | **GREEN — survives** (F3) |
| `schema` value check removed | **GREEN — survives** (F3) |

## Findings

### F1 — MAJOR — §13.1 binds `LAUNCH-PLAN-DETERMINISM-NEG` to a fixture that does not contain it

- Section: SPEC.md §13.1, conformance table ("The caller-plan path has these required
  conformance cases; Appendix D binds them to `fixtures/launch_plan_request_conformance.json`"),
  row `LAUNCH-PLAN-DETERMINISM-NEG`; Appendix D.2 "Launch Plan request" row ("a
  planning/launch argv mismatch"); Appendix D.4 ("every `LAUNCH-PLAN-*` case from
  `fixtures/launch_plan_request_conformance.json`").
- What is wrong: `grep -c DETERMINISM fixtures/launch_plan_request_conformance.json` = 0. The
  fixture carries six of the seven table IDs; the gate's `required_negatives` set omits the
  seventh, so the coverage class does not notice. The normative binding in the SPEC is false
  for one row, and the revision report's "Unverified" list does not say so (it says the
  cases run against the reference gate, which is true only for the six present).
- Fix (either): (a) add a `LAUNCH-PLAN-DETERMINISM-NEG` case whose document carries a
  `planning_argv`/`launch_argv` pair (or equivalent plugin-answer members) and extend the
  reference gate to compare step-4 argv to the planning answer, refusing with
  `provider_protocol_error`, plus a required-negative entry and an expected-red mutation that
  admits the mismatch; or (b) reword §13.1 so that the determinism row is bound to the
  implementation suite of Appendix D.4 and not to the fixture file, and say in D.2/D.4 that
  the fixture binds the six document-level cases. Option (a) matches Decision 7 item 4 ("each
  case is proven by a test that fails when the gate admits the input").

### F2 — MAJOR — the extensions-bound fixture is not the one-byte-over object D.4 requires

- Section: SPEC.md Appendix D.4 ("Narrowing mutants MUST … admit a persisted-extensions
  object one byte over the Section 1.6 bound … and each MUST fail"); fixture case
  `LAUNCH-PLAN-EXTENSIONS-NEG`.
- What is wrong: the persisted extensions object of that case (caller extensions + four
  Curator keys + `ax.launch-plan-request`) canonicalizes to 65,739 bytes, 203 over the bound.
  A gate narrowed by one byte (`EXTENSIONS_MAX_BYTES = 65537`) stays green; only a +203
  narrowing goes red. The fixture therefore cannot serve as the witness for the mutant the
  SPEC itself mandates, and the producer's `extensions-bound-relaxed` expected-red mutation
  relabels the expectation rather than narrowing the bound, so it does not detect this
  either. The stdin case shows the right shape (65,537 decoded bytes, +1 mutant red).
- Fix: size the case so the persisted canonical object is exactly 65,537 bytes (compute
  with the same JCS `canonical` the gate uses, including the `request_digest` value length),
  and add a positive sibling at exactly 65,536 so the bound is pinned from both sides; add an
  expected-red mutation that widens the fixture object by one byte below the bound.

### F3 — MINOR — the remaining §5.1/§14.1 limits and the `schema` value are enforced by the gate but proven by nothing

- Section: `scripts/validate_launch_plan.py` constants `ARGV_TOTAL_MAX_BYTES`,
  `ARGV_MAX_ELEMENTS`, `ENV_MAX`, `LITERAL_MAX_BYTES`, `EXTENSIONS_MAX_KEYS`, and the
  `document.get("schema") != SCHEMA` check; SPEC §14.1 ("a reader MUST reject an unknown
  member, an unknown `schema`, and a `schema_version` other than `1.0.0`").
- What is wrong: no fixture case exercises any of these bounds or an unknown `schema` value
  (every case document carries the correct `schema`; only `schema_version` is varied). Each
  +1 narrowing and the removed `schema` check survive. Decision 7 item 4 names only the three
  required negatives, so this is not a Decision 0013 gap, but the gate advertises "reference
  implementation of the Section 14.1 document validation" and Appendix D.2 lists "Unknown
  member, schema, or version" as required mutations.
- Fix: one boundary negative per constant (129 elements, 65,537 total argv bytes, 65 env
  names, a 4,097-byte literal, 65 extension keys) and an unknown-`schema` negative
  (`field: "schema"`).

### N1 — NOTE — reviewer harness pitfall, not a producer defect

Same-length source edits within one second reuse a stale `__pycache__` entry (pyc records
mtime in seconds plus size). Two of my mutants first appeared green for that reason and went
red once bytecode caching was disabled. The producer's expected-red mutations edit the JSON
fixture and the SPEC text, which are not cached, so this does not affect their results.

## What the next producer must deliver

New signed commits on top of `2c7f642` (no rewrite), F1 and F2 resolved, F3 recommended,
`validate_spec.py` / `test_expected_red.sh` / `git diff --check` / `run_validation.sh` exit 0,
plain push so PR #1 updates in place, PR body updated, PR not merged, updated revision report
attached, `task-board handoff TASK-260905-30a1bz --role developer`.
