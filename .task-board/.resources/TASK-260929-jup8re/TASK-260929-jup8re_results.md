# TASK-260929-jup8re results — naming gate: content-free reports + historical CI-echo exemption

## Change (uncommitted in the Story worktree)
- `.github/ci/naming-gate.sh`: hits are now printed as `path:line` only, for both the full-name and short-name scans. The line content is never printed.
- The exemption `ECHO` skips a candidate line only if it matches one of two alternatives:
  1. `Employer name gate(\t|\\t)[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:.]+Z \./` (the signature from the brief, verbatim).
  2. **Deviation from the brief (needed for AC5):** `…[0-9-]*T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?Z \./`, where `…` is U+2026. The only hit left on origin/main fc499a96 is the uyak0e `progress.md:74`, a runner excerpt truncated inside the log prefix (`  log: …-29T01:51:52.8528587Z ./.task-board/.resources/...patch:209085:<base85>`). The step name and year are gone, so alternative 1 cannot match it. Alternative 2 requires the ellipsis truncation marker directly followed by the timestamp tail, `Z` and ` ./`.
  - The binary-patch exemption is unchanged.
- `.github/ci/gate-selftest.sh`: rows f–k were added. Names are assembled at runtime.
  - (f) progress.md-style signature line → 0
  - (g) JSON `\t`-escaped signature line → 0
  - (h) step name without timestamp or `./` → 1
  - (j) prefix-truncated echo → 0
  - (k) prose with `…` but no timestamp tail → 1
  - (i) planted words → gate rc 1. Stdout+stderr contains neither word, and `./docs/notes.md:1` is present.

## Evidence (real exit codes)
- Naming rows via the extracted-section harness on the real gate: 16/16 ok, rc=0.
- Full `bash .github/ci/gate-selftest.sh`: 281 passed, 0 failed, rc=0.
- `bash .github/ci/naming-gate.sh <git archive origin/main extraction>` (origin/main = fc499a96, uyak0e records present): **rc=0**.
- The gate on the worktree itself: rc=0.
- Each mutant is a one-line change to the gate. The naming rows were run against each:
  - content printed again: rc=1, row (i) "no line content" FAIL
  - signature relaxed to just the step name: rc=1, row (h) FAIL
  - truncation alternative relaxed to just `…`: rc=1, row (k) FAIL
  - exemption removed: rc=1, rows (f), (g) and (j) FAIL. On the origin/main archive it gives rc=1 with 19 hits.
- Lint: shellcheck is not installed on this host, so it was NOT run. Substitute checks: `bash -n` on the gate rc=0, `bash -n` on the selftest rc=0, `python3 -m py_compile` on the gate's embedded python rc=0.
- Logbook: the brief forbids LOGBOOK edits and the CLI has no logbook command. The deviation decision is recorded in the task notes and in this artifact.

## Residuals (stated bounds)
- A line deliberately forged with either exact machine signature is exempt. Only machine-shaped CI-log echoes can be exempt; prose is never exempt, as rows (h) and (k) show.
- Alternative 2 has no step name, so it is broader than alternative 1. Any line with `…<date-tail>T<hh:mm:ss>Z ./` is exempt, whatever the step.
- The ledger TSV named in the task title was not built. The attached brief (the only current instruction) specifies the signature exemption instead.
