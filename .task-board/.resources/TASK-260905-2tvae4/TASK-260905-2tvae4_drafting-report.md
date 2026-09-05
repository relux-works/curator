# TASK-260905-2tvae4 drafting report: split manager-config schema-2 cases into their own vector family

## Commit
- Branch `draft/environments-manager-cli-1-1`, head `f61ee9a75cd1861a9993b0ee9ad4ad32a5ef3c9f`, exactly one commit past main `a68559b` (amended in place, force-with-lease pushed; previous heads 9af8af8 -> aba4d95 -> f61ee9a).
- Author Ivan Oparin <oparin@me.com>; SSH-signed (`git log --format=%G?` = U locally, no allowed_signers file; GitHub `commit.verification.verified` = true, reason `valid`).
- Intermediate head aba4d95 accidentally carried `tools/__pycache__/assurance.cpython-314.pyc`; removed before the final head.

## Files changed vs 9af8af8
| File | Change |
| --- | --- |
| conformance/v1/vectors/manager-config.json | restored to a68559b bytes (`git diff a68559b -- <file>` is empty) |
| conformance/v1/vectors/manager-config-v2.json | NEW: the 10 schema-2 cases plus `schema1-rejects-environments`, generator output |
| conformance/v1/manifest.json | regenerated (v2 family entry, manager-config.json hash back to a68559b) |
| release/1.0.0-rc.9.json | manifest pin regenerated to sha256:695d2ecf... |
| tools/generate-vectors/main.go | writes the schema-2 cases to manager-config-v2.json instead of appending to the schema-1 file |
| tools/validate.py | validate_manager_config_vectors reads both families; schema-1 file must carry schema_version 1 only; v2 file must be a non-empty list with at least one schema-2 case; names unique across both; per-case schema by schema_version; §12.1 knob/default cross-check unchanged |
| tools/test_validate.py | fixtures cover both files; new negatives: schema-2 case leaking into the frozen family, v2 without a schema-2 case, empty v2, name repeated across families |
| conformance/README.md | names the `vectors/manager-config-v2.json` family and the frozen schema-1 family |
| CHANGELOG.md | entry names the new vector file |

Shape note: the brief mentions an identity header (`schema_version`, `protocol_version`). The existing manager-config family is a bare case list with no header, so the v2 file mirrors that exact shape; each case's `input.schema_version` selects the schema.

Coverage ledger: `.github/ci/implementation-coverage.tsv` lists only consumed cases (no manager-config row at all), so it is unchanged. `implementation_coverage.py families` and `go` both upheld.

## Gates (each run as a standalone process, real exit codes)
| Command | Exit |
| --- | ---: |
| `go run ./tools/generate-vectors -root .` | 0 |
| `git diff a68559b -- conformance/v1/vectors/manager-config.json` | empty |
| `python3 tools/validate.py` (validated 58 schemas and 819 vector files) | 0 |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` (170 tests) | 0 |
| `make validate` (log .temp/make-validate-01.log) | 0 |
| `make regenerate-check` | 0 |
| `python3 tools/implementation_coverage.py families --root conformance/v1` (18 upheld) | 0 |
| pinned curator a3abcf34, `CURATOR_CONFORMANCE_ROOT=<worktree>/conformance/v1 go test -count=1 ./internal/interop` | 0 (`ok internal/interop 0.874s`) |
| same pin, full workflow package set with `-json`, then `implementation_coverage.py go --stream` (7 upheld) | 0 |
| NEGATIVE: same pin against a 9af8af8 root, `-run TestManagerConfigVectors` | 1 (FAIL on schema2-minimal-defaults, schema2-empty-environments-defaults, ...) — reproduces the hosted failure the split removes |

## Mutant / negative evidence (validator gate)
| Mutant | Narrows gate to | Failing test |
| --- | --- | --- |
| schema-2 case moved into manager-config.json | frozen family must be schema 1 only | test_schema_two_case_leaking_into_the_frozen_family_fails |
| v2 file with only schema-1 cases | v2 must carry a schema-2 case | test_v2_family_without_a_schema_two_case_fails |
| v2 file empty list | v2 non-empty | test_empty_v2_family_fails |
| case name duplicated across files | names unique across families | test_name_repeated_across_families_fails |
| forged valid flag / defaults drift / open object / etc. | unchanged from 9af8af8 | existing ManagerConfigVectorTests (18 total, all pass) |
No survivors among the mutants exercised. Bound: the tests do not prove the pinned Go manager reads only manager-config.json; that is established by the direct interop run above, not by the Python gate.

## PR #41 checks at head f61ee9a
Implementations run 33969621585: ubuntu-latest success, macos-latest success, windows-latest success. Specification (3 OS), Links, Formatting: pass. Release target provenance: skipping (as on main).
