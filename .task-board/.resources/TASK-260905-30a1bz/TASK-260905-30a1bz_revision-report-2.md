# TASK-260905-30a1bz — revision report 2 (rework 1 of ax PR #1)

Subject: `relux-works/agent-session-manager-spec` PR #1, branch
`draft/curator-environment-integration`. Previous head `2c7f642`; new head
`c6270a37ae3a8573f18fbb5dae2b78c58525cf64`, pushed by plain push
(`2c7f642..c6270a3`, no force). PR https://github.com/relux-works/agent-session-manager-spec/pull/1
state OPEN, `mergedAt` null, `headRefOid` = local HEAD. PR body updated
(`gh pr edit 1 --body-file`): case counts, mutant counts, rework section. Not merged.
`VERSION`, `CHANGELOG.md`, `RELEASE_NOTES.md`, `README.md`: `git diff --stat main..HEAD`
on those paths is empty. Frozen map: only the `SPEC.md` digest moved
(`6bbefa4d…` → `f474b10b…`). Nothing written into the control root.

## Commits (on top of `2c7f642`, no rewrite; `git log --oneline main..HEAD` = c6270a3, 6144ff5, 2c7f642, ef91985, d7075e1)

| Commit | Subject | Signature |
| --- | --- | --- |
| `6144ff5248cb1d4787cd32fa68ad1bc8a85d1d3d` | fixtures: launch-plan determinism case and exact boundary cases | Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM; Author Ivan Oparin <oparin@me.com> |
| `c6270a37ae3a8573f18fbb5dae2b78c58525cf64` | validation: gate-narrowing expected-red mutants for the launch-plan gate | same key, same author |

## Finding → disposition

| Finding | Disposition | Where |
| --- | --- | --- |
| F1 (major) determinism row unbound | Option (a). `LAUNCH-PLAN-DETERMINISM-NEG` added with optional case member `plugin_answers: {planning_launch_argv, step_4_launch_argv}` (named after §13.1's "planning call" and "step 4's launch"). Gate: `planning_launch_argv` must equal the resolved final argv (else fixture reference error, i.e. the fixture cannot lie about the planning answer); `step_4_launch_argv != planning` → `provider_protocol_error`, `details {reason: "launch_argv_mismatch", argv_index: <first differing index>}`. Added to `required_negatives`; `provider_protocol_error` added to the required observed-code set. Mutants: gate `determinism-admitted`, gate `determinism-length-only`, fixture `determinism-relabeled`. | `scripts/validate_launch_plan.py`, fixture, `test_expected_red.sh` |
| F2 (major) extensions fixture 203 over | `LAUNCH-PLAN-EXTENSIONS-NEG` re-sized to exactly 65,537 persisted canonical bytes; new `LAUNCH-PLAN-EXTENSIONS-POS` at exactly 65,536 (required positive). Sizes computed by `.temp/rework1/build_cases.py` using the validator's own `canonical` and the gate's persisted object (caller extensions + 4 Curator keys + `ax.launch-plan-request` with the real 71-char digest). Mutants: gate `EXTENSIONS_MAX_BYTES` 65536→65537 (NEG admitted → red); fixture `extensions-pos-widened` (+1 byte on the POS padding → refused → red). Existing `extensions-bound-relaxed` kept. | fixture, gate, suite |
| F3 (minor) unproven bounds + schema | Negatives (all in `required_negatives`): `LAUNCH-PLAN-SCHEMA-NEG` (`field: "schema"`), `LAUNCH-PLAN-ARGV-ELEMENTS-BOUND-NEG` (final argv 129 = base 1 + 128), `LAUNCH-PLAN-ARGV-ELEMENT-BYTES-BOUND-NEG` (4,097-byte element; not in the brief, added as it is a gate constant too), `LAUNCH-PLAN-ARGV-BYTES-BOUND-NEG` (65,537 total: `codex` 5 + 15×4,096 + 4,092), `LAUNCH-PLAN-ENV-NAMES-BOUND-NEG` (65 names), `LAUNCH-PLAN-LITERAL-BOUND-NEG` (4,097-byte value), `LAUNCH-PLAN-EXTENSIONS-KEYS-BOUND-NEG` (60 caller keys → 65 persisted; the brief's "65 extension keys" is read as the persisted count because that is what the gate bounds — 65 caller keys would be 70 persisted and a +1 narrowing would not catch it). Positive `LAUNCH-PLAN-EXTENSIONS-KEYS-POS` at 64 persisted. Each +1 narrowing is an expected-red gate mutant. | fixture, gate, suite |
| N1 (note) pycache | Gate mutants `rm -rf scripts/__pycache__` in the copy and run under `PYTHONDONTWRITEBYTECODE=1`; the scratch proofs below ran the same way. | `test_expected_red.sh` |

SPEC change: Appendix D.4 launch-plan bullet now lists the required narrowing mutants
(bound +1, unknown `schema`, step-4 argv mismatch). No other SPEC text changed.

## Fixture sizes (from `build_cases.py`, validator JCS `canonical`)

| Case | Persisted canonical bytes | Padding value length |
| --- | ---: | ---: |
| `LAUNCH-PLAN-EXTENSIONS-POS` | 65,536 | 64,997 |
| `LAUNCH-PLAN-EXTENSIONS-NEG` | 65,537 | 64,998 |

Fixture totals: 5 positive, 23 negative (was 3 / 15).

## Narrowing proofs on a scratch copy (`.temp/rework1/narrowings.sh`, `PYTHONDONTWRITEBYTECODE=1`, `__pycache__` excluded; log `.temp/rework1/narrowings.log`)

| Mutant on `validate_launch_plan.py` | Result |
| --- | --- |
| `EXTENSIONS_MAX_BYTES` 65536→65537 | RED exit 1: LAUNCH-PLAN-EXTENSIONS-NEG was admitted |
| `EXTENSIONS_MAX_KEYS` 64→65 | RED exit 1: LAUNCH-PLAN-EXTENSIONS-KEYS-BOUND-NEG was admitted |
| `ARGV_MAX_ELEMENTS` 128→129 | RED exit 1: LAUNCH-PLAN-ARGV-ELEMENTS-BOUND-NEG was admitted |
| `ARGV_ELEMENT_MAX_BYTES` 4096→4097 | RED exit 1: LAUNCH-PLAN-ARGV-ELEMENT-BYTES-BOUND-NEG was admitted |
| `ARGV_TOTAL_MAX_BYTES` 65536→65537 | RED exit 1: LAUNCH-PLAN-ARGV-BYTES-BOUND-NEG was admitted |
| `ENV_MAX` 64→65 | RED exit 1: LAUNCH-PLAN-ENV-NAMES-BOUND-NEG was admitted |
| `LITERAL_MAX_BYTES` 4096→4097 | RED exit 1: LAUNCH-PLAN-LITERAL-BOUND-NEG was admitted |
| `STDIN_MAX_BYTES` 65536→65537 | RED exit 1: LAUNCH-PLAN-STDIN-BOUND-NEG was admitted |
| `schema` check removed | RED exit 1: LAUNCH-PLAN-SCHEMA-NEG was admitted |
| determinism check disabled | RED exit 1: LAUNCH-PLAN-DETERMINISM-NEG was admitted |
| determinism compared by length only | RED exit 1: LAUNCH-PLAN-DETERMINISM-NEG was admitted |

The same eleven mutants are now permanent expected-red mutations (`run_launch_plan_gate_mutation`).

## Gate exit codes at `c6270a3` content (each run as a standalone process; logs under `.temp/rework1/`)

| Command | Exit | Evidence |
| --- | ---: | --- |
| `./scripts/validate_spec.py` | 0 | Launch-plan ledger gate_classes=6, positive_cases=5, negative_cases=23 (`validate.log`) |
| `./scripts/test_expected_red.sh` | 0 | 327 passed, 0 failed out of 327 mutations (`expected-red.log`; 13 new) |
| `./run_validation.sh` | 0 | "Validation successful: all contracts, diagrams, and publication artifacts are fresh" (`run-validation.log`) |
| `git diff --check` | 0 | clean |

Note: the four gates ran on the working tree immediately before the two commits; the committed
content is byte-identical to what was validated (`git status` clean after commit, no further edits).

## Unverified / not done

- No `ax` implementation exists; the fixture cases run against the reference gate only, as before.
- The expected-red suite ran once end-to-end (about 25 minutes); it was not repeated after the push.
