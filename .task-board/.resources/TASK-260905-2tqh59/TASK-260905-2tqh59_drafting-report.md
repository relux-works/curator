# Drafting report: environments 1.1 follow-ups (TASK-260905-2tqh59)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-follow-ups`, branch `draft/environments-1-1-follow-ups`, base curator-spec main `fd237ba`. One signed commit `fcdb9ba` (ECDSA key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`, good signature; principal verified against the repository's `maintainers.allowed_signers`). 36 files, +1266/−79. No push, tag, PR, or LOGBOOK.md. The managed story worktree is untouched (empty story-side delta). Item 3 (`system-config-v2` lockable keys) is not in this task, as briefed.

## Item → change

| Item | File : section | Change |
| --- | --- | --- |
| 1a | `decisions/0012-context-packages-and-semver-locks.md` : Erratum item 1; Compatibility impact table | manager §12.4 row reads *bytes change* (isolation knob, liveness row, provisioning seeds; evidence `f61ee9a`); row annotated `[Erratum 2026-09-05, item 1]`, cells unedited |
| 1b | same : Erratum item 2; Decision 2 "Resolution" | two exact constraints peeling to different commits fail `context_range_conflict` (vector `exact-constraints-disagree`); sentence annotated |
| 1c | same : Erratum item 3; Decision 9 fragment example | descriptors omit `argument`; read as pre-revision (environments §10.2/§13); introducing sentence annotated, JSON unedited |
| 2 | `protocol/environments.md` §12.1 (after the knob table); `profiles/manager.md` §12.6 | `secret_material_waivers.pin` is bare lowercase hex, 40 (`commit`) or 64 (`state_sha256`), no `sha256:` prefix; matches `manager-config-v2` `$defs/secretMaterialWaiver.pin` → `common#/$defs/commit` (`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`) |
| 4 | `protocol/environments.md` §7.7 | new row `environment_form_unavailable` (warning; cross-references §5.7) |
| 5 N-a | `protocol/environments.md` §8.1 `linked` bullet | "(except the `claude_code` root-context surface, below)" |
| 5 N-b | `protocol/environments.md` §10.1 | a copied surface has no link target and is verified by the marker's content hash (§8.2), the hash §8.4 drift compares |
| 6 | `tools/validate.py` (`MANAGER_CONFIG_KNOB_ENUM_PATHS`, `environments_knob_rows/values`, cross-check in `validate_manager_config_vectors`); `tools/test_validate.py` `ManagerConfigVectorTests` | the seven closed-set knobs' enums must equal the §12.1 `Values` backticked set; tests `test_widened_enum_fails`, `test_narrowed_enum_fails`, `test_table_value_drifting_from_the_enum_fails`, `test_every_enum_knob_is_cross_checked` |
| 7 | `tools/generate-vectors/manager_config.go` | schema cases `invalid-overlay-range-grammar`, `invalid-overlay-tag-grammar`, `invalid-overlay-empty-source`; vectors `schema2-overlay-range-grammar`, `schema2-overlay-tag-grammar`, `schema2-overlay-empty-source` in `manager-config-v2.json`; `manager-config.json` byte-identical (`git diff --quiet` at `fd237ba`) |
| 8 | `profiles/manager.md` §12.5; `cli/curator.md` `curator run` paragraph | always-`--repair` rule cites environments §9.2 and §10.1; Decision 0013 D6.4 stays the citation for the provider column |
| 9 F1 | `tools/validate.py` wire semantics (`context-lock-v1`); `environments.go` case `invalid-required-by-self`; test "lock required_by self" | a member naming itself in `required_by` is rejected |
| 9 F2 | `schemas/v1/agent-environment-marker-v1.schema.json`; `tools/validate.py` wire semantics; `environments.go` cases `invalid-copy-outside-paths`, `invalid-root-context-missing-form`, `invalid-form-on-mcp-surface`, `invalid-surface-unknown-key`; test "marker copy outside its paths"; environments §8.2 clause | `surfaces` closed to `mcp`, `root-context`, `skills`, `system-prompt`; `form` required on `root-context` only; copies ⊆ paths in wire semantics. Five existing root-context negatives gained `form` so each stays single-reason |
| 9 F3 | `schemas/v1/launch-env-fragment-v1.schema.json` `absolutePath`; cases `invalid-env-dotdot-segment`, `invalid-path-prepend-dotdot-segment`, `invalid-system-prompt-path-dotdot-segment`; environments §10.2 clause | `..` segments rejected by pattern; containment stays the reader rule |
| 9 F4 | `environments.go` materialization input `note`; `vectors/environments.json` case `mcp-pi-none` | `note` explains `env_names` is recorded for information only on pi |
| — | `CHANGELOG.md` Unreleased → Changed | one entry covering the above |

## Anomaly found while gating

The positive marker case `valid-linked-symlink-fallback` recorded a `skills/pdf` copy on a skills surface whose `paths` was `[]`; the new copies ⊆ paths check rejected it (`.temp/validate-01.log`). The fixture, not the rule, was wrong: the case now lists `skills/pdf` in `paths`. Regenerated case bytes changed accordingly.

## Gates (run by me, standalone commands, real exit codes)

| Command | Exit | Evidence |
| --- | ---: | --- |
| `make validate` (baseline, `fd237ba`) | 0 | `.temp/validate-00.log` — 204 tests OK |
| `make validate` (first run after edits) | 2 | `.temp/validate-01.log` — the fixture anomaly above |
| `make validate` (final) | 0 | `.temp/validate-02.log` — `validated 59 schemas and 993 vector files`; `Ran 208 tests … OK`; `go test ./tools/...` ok |
| `make regenerate-check` (unstaged tree) | 2 | `.temp/regen-check-01.log` — expected: the target diffs the working tree against the index, and the regenerated files were unstaged |
| `make regenerate-check` (staged) | 0 | `.temp/regen-check-02.log` |
| `make regenerate-check` (post-commit `fcdb9ba`) | 0 | `.temp/regen-check-03.log` |
| `go vet ./tools/...` | 0 | inline |
| `git verify-commit HEAD` | 0 | good ECDSA signature; principal resolved with `-c gpg.ssh.allowedSignersFile=maintainers.allowed_signers` |
| `release/1.0.0-rc.9.json` | — | diff is exactly the two `manifest_sha256` lines |

Gate tails:

```
validate-02.log:  Ran 208 tests in 41.674s / OK / ok github.com/relux-works/curator-spec/tools/generate-vectors 0.520s
regen-check-03.log: git diff --exit-code -- conformance/v1 release/... (exit 0)
```

Python: venv `.temp/venv` (jsonschema 4.25.1, from `requirements-dev.txt`; the Makefile calls bare `python3`, so the venv `bin` was put first on `PATH`). Go 1.25.5.

## Mutant evidence (scratch copies under `.temp/mut`, real `tools/validate.py` entry point unless stated; log `.temp/mutants-01.log`)

| Mutant | Narrows the gate to | Named failing check | Bound if surviving |
| --- | --- | --- | --- |
| M1 widen `precedence.winner` enum in the schema | item 6 enum set equality | `validate.py`: "enum for precedence.winner is ['heavier', …]; section 12.1 states […]" (exit 1) | — |
| M2 delete the copies ⊆ paths check | F2 semantic rule | case `invalid-copy-outside-paths` "expected valid=False, got valid" (exit 1) | — |
| M3 delete the self-`required_by` check | F1 semantic rule | case `invalid-required-by-self` (exit 1) | — |
| M4 drop the fragment `..` `not` pattern | F3 | case `invalid-path-prepend-dotdot-segment` (exit 1) | — |
| M5 reopen marker `surfaces` to any key | F2 unknown surface keys | case `invalid-surface-unknown-key` (exit 1) | — |
| M6 admit `form` on every surface | F2 per-surface form | case `invalid-form-on-mcp-surface` (exit 1) | — |
| M7 relax the enum check to subset, plus widened schema | item 6, the check itself | `validate.py` alone: exit 0 (**survives that layer**); `test_validate.ManagerConfigVectorTests.test_widened_enum_fails` FAILS "ValidationFailure not raised" (exit 1), which `make validate` runs | bound: the bare `validate.py` run cannot detect a weakening of its own comparison; the unit test is the layer that pins it |
| M8 loosen `versionRange` pattern to `^.+$` | item 7 range grammar | case `invalid-overlay-range-grammar` (exit 1) | — |
| M9 admit an empty overlay `source` | item 7 | case `invalid-overlay-empty-source` (exit 1) | — |

Not mutated: the overlay `tag` grammar (`gitRefName` is `common.schema.json`, shared and covered elsewhere) — its new case `invalid-overlay-tag-grammar` is generated and validates as rejected, which proves reachability, not the class bound. The §7.7 row, the prose items, and the erratum have no executable gate beyond `validate.py`'s section-table parsers, which stayed green.

## Readings to confirm at review

1. Erratum item 2 states the failure as the step-2 empty effective constraint; environments §1.4 carries the same sentence and is not edited (the erratum says it is read the same way).
2. `form` is now *required* on `root-context` (§8.2 says "its form where the surface has one"); the reviewer's F2 asked for admission only under `root-context`, and every root-context surface has a form, so required is the stricter reading.
3. The three `..` negatives are single-reason: each path stays below `/manager/environments/` so the wire-semantics prefix check does not fire first.
