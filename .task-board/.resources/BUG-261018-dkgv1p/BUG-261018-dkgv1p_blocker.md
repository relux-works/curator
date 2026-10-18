# Resolved infrastructure blocker

The previous runner blockers are resolved by full hosted attempt 4 on signed candidate `233c8873889a9f25cd0e2d040a39bcf564a6540c`: 20 jobs success, 2 optional skipped, 0 failures. See `BUG-261018-dkgv1p_hosted-retry-04.md` and https://github.com/relux-works/curator/actions/runs/39067488895/attempts/4 . No source change was required. The packet below is retained as historical failure evidence, not a current blocker.

---

# BUG-261018-dkgv1p — hosted validation blocked by runner disk capacity

## Constraint and evidence

The code/test candidate is `233c8873889a9f25cd0e2d040a39bcf564a6540c`, signed (`G`), tree `796a42f78bf9fdad6dedd31e571653be096577a8`. Story HEAD remains the recorded checkpoint; changes are uncommitted as required.

[Full CI workflow 39067488895](https://github.com/relux-works/curator/actions/runs/39067488895) cannot execute its required Linux verification because the self-hosted runners have exhausted writable disk space. This is direct provider-log evidence, not an inference from missing checks:

| Attempt | Job | Observed failure | Actual result |
| --- | --- | --- | --- |
| 1 | `Test (ubuntu-latest)`, job `117498057066` | `go vet` cannot create compiler/vet work files: `no space left on device` | Process exit 1 |
| 1 | `Go 1.25.5 driver (ubuntu-latest)`, job `117498057121` | Standard-library compilation cannot create its archive: `no space left on device` | Process exit 1 |
| 2 | `Test (ubuntu-latest)`, job `117498988846` | setup-go's `go env` cannot create its work directory: `no space left on device` | Job failure; inner process exit code not separately exposed |
| 2 | `Go 1.27.1 driver (ubuntu-latest)`, job `117498988889` | Compiler/vet writes fail: `no space left on device` | Process exit 1 |

The Linux Go 1.25.5/1.26.0/1.27.1 driver jobs failed across the retries. Actual tests for these jobs did not run. A failed compile or setup is not a passing behavioral test or a killed mutant.

Raw provider logs remain under the ignored task-local `.temp/BUG-261018-dkgv1p/` directory as `linux-test-failure-02.log`, `linux-driver-failure-02.log`, `linux-test-retry-failure-01.log` and `linux-driver-retry-failure-01.log`. Only sanitized findings are attached to the public board.

## Attempts and failed assumption

- The initial assumption was that the configured hosted gate had runnable Linux capacity. Direct ENOSPC failures disproved it.
- A targeted `gh run rerun 39067488895 --job 117498057066` exited 1. The REST retry endpoint independently returned HTTP 403 because the workflow was still running.
- Cancelled the remaining jobs, waited for the run to reach terminal cancellation, then ran `gh run rerun 39067488895 --failed`: exit 0, attempt 2 admitted on the identical candidate. The same disk-capacity failure recurred before behavioral testing.
- Requested cancellation of the remaining attempt-2 jobs after the repeated external failure; no pending job is represented as green.
- Local exact-candidate targeted help/regression/mutant suite, build, vet and lint all exit 0. These do not replace required full Linux/macOS/Windows hosted verification.

## Options and decision/input needed

1. **Recommended:** the runner owner frees space on the self-hosted Linux validation pool, then reruns the failed/cancelled workflow jobs on this exact candidate. This preserves the configured platform and validation contract.
2. An authorized infrastructure owner routes these jobs to healthy equivalent Linux capacity. This changes shared CI routing and should be handled separately from this CLI fix.
3. A native local mirror cannot establish the missing Linux platform evidence. No equivalent Linux runner has been provided to this run; reporting a native local suite as a full mirror would overclaim coverage.

Required external input: restore writable free space on the configured Linux validation runners, or provide approved equivalent Linux capacity. Do not weaken checks or treat cancellation as success. This run does not own the shared runner filesystem and cannot safely delete another job's caches/workspaces or alter the repository-wide routing policy.

After recovery, run `gh run rerun 39067488895 --failed` once the workflow is terminal. Recovery snapshot branch: `gate/BUG-261018-dkgv1p/261017-203714`. Keep the same signed candidate unless code, tests, configuration or environment identity changes. Inspect the full actual hosted result, update the validation resource, check the full-suite checklist row only if green, then perform the developer handoff. No review handoff is admitted yet.

## Logbook entry

The blocking condition is repeatable external ENOSPC in required hosted Linux jobs. Code and targeted evidence are preserved; the full-suite checklist item remains unchecked. No product workaround, shared-runner cleanup, source change to CI, fabricated check status, or direct control-root logbook edit was made.
