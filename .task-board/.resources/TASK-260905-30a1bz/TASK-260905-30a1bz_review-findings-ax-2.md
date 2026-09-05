# TASK-260905-30a1bz — review cycle 2 verdict: ACCEPT (with one minor, non-blocking note)

Subject: ax PR #1, branch `draft/curator-environment-integration`, head `c6270a37ae3a8573f18fbb5dae2b78c58525cf64`
(rework commits `6144ff5`, `c6270a3` on cycle-1 head `2c7f642`). Reviewed diff: `git diff 2c7f642..c6270a3`
(SPEC.md +8/-6 in Appendix D.4 only; fixture; `validate_launch_plan.py`; `validate_spec.py` SPEC digest only; `test_expected_red.sh`).

## Change Request CR-TASK-260905-30a1bz-2 rev 2: `repository_delta=empty` — why that is correct

The task's deliverable lives in a different repository (`relux-works/agent-session-manager-spec`, worktree
`/Users/iv/Developer/ReluxWorks/.temp/ax-curator-integration/worktree`). The producer brief explicitly forbids writing
into the curator-spec control root and the story branch carries no curator-spec change. The outcome is the pushed
ax branch plus the attached reports. An empty delta on `task-board/story/STORY-260905-3v5dn3` is therefore the
expected shape, not a missing deliverable.

## F1–F3 of review-findings-ax-1 (verified, not read)

| Finding | Verified how | Result |
| --- | --- | --- |
| F1 determinism | `LAUNCH-PLAN-DETERMINISM-NEG` present with `plugin_answers.planning_launch_argv` / `step_4_launch_argv` (differ at index 4); gate refuses `provider_protocol_error {reason: launch_argv_mismatch, argv_index: 4}`; in `required_negatives`; `provider_protocol_error` in required observed-code set; suite mutants `determinism-relabeled`, `determinism-admitted`, `determinism-length-only` | resolved |
| F2 extensions size | Recomputed on the live gate with `validate_spec.canonical` capturing the persisted object at the bound check: `LAUNCH-PLAN-EXTENSIONS-POS` = 65,536 bytes (admitted), `LAUNCH-PLAN-EXTENSIONS-NEG` = 65,537 bytes (refused `field: extensions`); mutants `extensions-pos-widened` and gate `EXTENSIONS_MAX_BYTES` 65536→65537 present | resolved |
| F3 bounds + schema | Recomputed: final argv 129 elements; total argv 65,537 bytes (17 elements); max element 4,097 bytes; 65 `env_names`; 4,097-byte literal; 65 persisted extension keys (POS sibling 64); `LAUNCH-PLAN-SCHEMA-NEG` uses `urn:ax:schema:launch-plan` with `field: schema`. All in `required_negatives`. | resolved |
| N1 pycache | Suite `mutate_launch_plan_gate` does `rm -rf __pycache__` and runs under `PYTHONDONTWRITEBYTECODE=1` | resolved |

## Narrowings rerun myself on a scratch copy (`PYTHONDONTWRITEBYTECODE=1`, `__pycache__` removed)

All eleven producer mutants go red with the exact diagnostic: `EXTENSIONS_MAX_BYTES`+1, `EXTENSIONS_MAX_KEYS`+1,
`ARGV_MAX_ELEMENTS`+1, `ARGV_ELEMENT_MAX_BYTES`+1, `ARGV_TOTAL_MAX_BYTES`+1, `ENV_MAX`+1, `LITERAL_MAX_BYTES`+1,
`STDIN_MAX_BYTES`+1, schema check removed, determinism disabled, determinism length-only — each exit 1, diagnostic hit 1.
`expect_fail` asserts the named diagnostic, not merely a nonzero exit.

## Additional attack (my own mutant)

`if step_4 != planning` → `if step_4[base_length:] != planning[base_length:]` SURVIVES (validate_spec exit 0). The only
determinism fixture mismatches inside the caller suffix, so a gate that ignores base-argv drift at step 4 is not
caught. **Minor, non-blocking**: the brief asked for one mismatch pair and got it; §13.1 text is unchanged and correct.
Recommend for a later cycle: a second determinism case whose mismatch sits at index < `base_argv_length`.

## Gates rerun by me at `c6270a3` (worktree clean before and after)

| Command | Exit |
| --- | ---: |
| `./scripts/validate_spec.py` | 0 (ledger gate_classes=6, positive_cases=5, negative_cases=23) |
| `./scripts/test_expected_red.sh` | 0 (327 passed, 0 failed out of 327 mutations; ran under PYTHONDONTWRITEBYTECODE=1) |
| `./run_validation.sh` | 0 ("Validation successful") |
| `git diff --check` | 0 |

Frozen map: only `SPEC.md` digest changed (`git diff 2c7f642..c6270a3 -- scripts/validate_spec.py` = 2 lines).
`VERSION`/`CHANGELOG.md`/`RELEASE_NOTES.md`/`README.md`: `git diff --stat main..HEAD` on those paths empty.

## Regression check on the cycle-1 items

Fixture: `profile_mappings`, `base_argv`, `curator_extension_keys`, schema unchanged; no case removed; only
`LAUNCH-PLAN-EXTENSIONS-NEG` modified (resize). SPEC change limited to the D.4 mutant list. Gate change adds the
optional `plugin_answers` member and determinism check after the extensions bound; existing refusal order untouched.

## Delivery

`git log --oneline main..HEAD` = c6270a3, 6144ff5, 2c7f642, ef91985, d7075e1 (no rewrite). Both new commits:
Good "git" signature for oparin@me.com, ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM, author Ivan Oparin.
`gh pr view 1`: headRefOid = c6270a3…, state OPEN, mergedAt null, body cites Decision 0013, proposes v0.6.0, has the
rework section with the 65,536/65,537 numbers.

Verdict: ACCEPT — `accept_cr(TASK-260905-30a1bz, revision=2)`. Not marked done. Nothing written into the control root;
scratch under the ax worktree `.temp/review2/`.
