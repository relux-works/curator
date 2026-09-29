# TASK-260929-1kze6z results

## Diff summary
- NEW `.github/ci/naming-gate.sh`: patterns assembled from parts; `grep -rIliE` lists candidate files, python3 rescans each file line by line and reports every match except lines inside a `GIT binary patch` block (to next `diff --git ` / EOF) that are `literal N`/`delta N`, empty, or base85-shaped (`^[A-Za-z][0-9A-Za-z!#$%&()*+;<=>?@^_`{|}~-]+$`). grep exit >1 or python failure -> exit 2 (fail closed). Optional root arg.
- `.github/workflows/ci.yml`: "Employer name gate" step now `run: bash .github/ci/naming-gate.sh`.
- `.github/ci/gate-selftest.sh`: naming rows a-e + reason matches (8 asserts).
- CHANGELOG: not touched (Unreleased has no CI subsection).
- Note: first attempt used `grep -Z -n` record parsing; BSD grep emitted records without the NUL separator, so the design switched to file-list + python rescan.

## Local exit codes
- `bash .github/ci/naming-gate.sh` (current tree) -> 0
- `bash .github/ci/gate-selftest.sh` (full) -> 0, 273 passed, 0 failed
- shellcheck: not installed on host, not run.

## Selftest rows
| row | input | want | got |
|---|---|---|---|
| a | short word as whole word (`;..;`) on base85 line in binary block | 0 | 0 |
| b | short word on `+` text line of .patch | 1 + short-name msg | 1, msg matched |
| c | short word on space-containing line inside binary block | 1 + short-name msg | 1, msg matched |
| d | full name (uppercased) in .md | 1 + full-name msg | 1, msg matched |
| e | clean tree | 0 | 0 |

## Mutants
| mutant | change | killed by |
|---|---|---|
| no-skip | exemption branch -> `elif False:` | row (a) FAIL; real tree exit 1 |
| no-shape-check | `B85.match(line)` -> `True` | row (c) + its reason row FAIL |
Both restored; baseline re-run 8/8, real tree 0.
