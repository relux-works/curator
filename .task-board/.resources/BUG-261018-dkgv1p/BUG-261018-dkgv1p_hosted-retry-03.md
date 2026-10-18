# BUG-261018-dkgv1p — hosted gate retry after runner storage repair

Status: blocked. The full hosted gate was rerun on the unchanged signed candidate. A different runner infrastructure failure prevents green validation. No developer handoff or acceptance is claimed.

## Candidate and unchanged evidence

- Signed candidate: `233c8873889a9f25cd0e2d040a39bcf564a6540c`; signature verification returned `G` again.
- Tree: `796a42f78bf9fdad6dedd31e571653be096577a8`; rebuilt from all tracked and nonignored untracked worktree files with a private Git index and matched exactly.
- Managed Story HEAD remains `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`. The candidate remains uncommitted in the Story worktree. No source, tests, configuration or branch changed in this run.
- Previous local CLI, positive-control, passthrough, mutant, build, vet and lint evidence is reused from `BUG-261018-dkgv1p_validation.md`; none was rerun or relabeled as newly executed here. The previous report records Go 1.25.5 and golangci-lint 2.12.2. This run executes the hosted workflow with its own configured environments.
- Coverage remains **5 of 6 AC rows driven** by named production-entry tests; AC6 full-suite validation remains unsatisfied. The previous report preserves each production call site. No surface table was supplied in the brief; that remains a brief gap, not an implied coverage waiver.
- No previously committed reviewer test was removed. No source-text admission gate is introduced, so the source-text mutation checklist row remains inapplicable and unchecked. No rework diff was introduced.

## Commands executed and actual results

[Full CI workflow, attempt 3](https://github.com/relux-works/curator/actions/runs/39067488895/attempts/3).

| Command or process | Exit / result | Meaning |
| --- | --- | --- |
| `git log -1 233c8873889a9f25cd0e2d040a39bcf564a6540c --format='%H %T %G?'` | 0; signature `G` | Exact previously signed candidate verified |
| Private-index batch: `git read-tree HEAD`, `git add -A .`, `git write-tree` | Batch exit 0; exact tree above | Worktree matches candidate; normal index unchanged |
| `gh run rerun 39067488895 --repo relux-works/curator` | 0 | Full workflow rerun admitted; not a passing test verdict |
| Lint action binary extraction (`tar`) | **2** | `golangci-lint: Cannot open: Text file busy`; linter never ran |
| `gh api repos/relux-works/curator/actions/jobs/117501666614/logs --allow-escape-sequences` | 0 | Retrieved direct failed-job evidence |
| `gh run rerun 39067488895 --repo relux-works/curator --job 117501666614` | **1** | `job 117501666614 cannot be rerun` while workflow active |
| `gh run cancel 39067488895 --repo relux-works/curator` | 0 | Producer stopped remaining jobs after confirmed infrastructure failure |
| `gh run view 39067488895 --repo relux-works/curator --json headSha,status,conclusion,attempt,jobs` | 0 | Fresh terminal read: attempt 3, completed/cancelled, exact SHA |
| `git diff --check` | 0 | Candidate diff remains whitespace-clean |

No test or gate was piped through `tee`. An early `gh run view --job ... --log-failed` read reported that logs were unavailable while the run was active; it is not evidence of absence or success. An initial job-log read without `--allow-escape-sequences` was refused by the CLI output filter and supplied no evidence; the explicit permitted read above succeeded. Raw provider logs are retained only in ignored task-local storage; the sanitized failure above contains no runner paths or identities.

## Hosted job results

Cancelled jobs are incomplete verification, never green evidence. Skipped optional lanes are not tested lanes.

| Job | Result | Job ID |
| --- | --- | --- |
| Naming gate | success | 117501666403 |
| Interop conformance gate | success | 117501666557 |
| Gate self-test (windows-latest) | success | 117501666560 |
| Test (macos-latest) | cancelled | 117501666569 |
| Go 1.27.1 driver (ubuntu-latest) | success | 117501666570 |
| Test (windows-latest) | cancelled | 117501666588 |
| Gate self-test (ubuntu-latest) | success | 117501666593 |
| Lint | failure | 117501666614 |
| Race (macos-latest) | cancelled | 117501666623 |
| Go 1.25.5 driver (macos-latest) | cancelled | 117501666629 |
| Go 1.26.0 driver (ubuntu-latest) | success | 117501666630 |
| Gate self-test (macos-latest) | success | 117501666632 |
| Race (ubuntu-latest) | cancelled | 117501666653 |
| Go 1.25.5 driver (ubuntu-latest) | success | 117501666655 |
| Go 1.25.5 driver (windows-latest) | cancelled | 117501666686 |
| Test (ubuntu-latest) | cancelled | 117501666688 |
| Go 1.27.1 driver (macos-latest) | cancelled | 117501666697 |
| Go 1.27.1 driver (windows-latest) | cancelled | 117501666745 |
| Go 1.26.0 driver (macos-latest) | cancelled | 117501666748 |
| Go 1.26.0 driver (windows-latest) | cancelled | 117501666777 |
| Candidate suite (${{ matrix.os }}) | skipped | 117501667616 |
| Test (rose-air) | skipped | 117501687569 |

## Reused narrowing mutant evidence

Not rerun during this gate-only assignment. All controls below belong to the exact unchanged candidate and previous attached local evidence; no new mutant kill is claimed.

| Mutant | What it narrows the gate to | Named failing test | Result / survivor bound |
| --- | --- | --- | --- |
| `init-bypasses-help` | Guard remains for every built-in except `init` | `TestRegisteredSubcommandHelpNoWrites/init/1` | Child exit 1; killed |
| `only-leading-flags` | Guard recognizes only a leading help operand | `TestRegisteredSubcommandHelpNoWrites/init/2` | Child exit 1; killed |
| `help-crosses-separator` | Scan crosses `--` | `TestCommandHelpRunPassthrough` | Child exit 1; killed |

No surviving mutants are reported by the previous evidence. Snapshot coverage bounds and out-of-contract provider/internal-mode rows remain explicitly listed in the original validation report; this rerun adds no exclusions.

## Task-scoped logbook and blocker recovery

1. Storage repair allowed all three Linux Go driver qualification jobs to pass and Linux full/race jobs to begin behavioral testing. This does not establish full-suite success.
2. Lint job `117501666614` failed during installation of pinned golangci-lint 2.12.2: the action extracts to a shared binary destination which was busy. The direct job log reports `tar` exit 2 before linter execution. This is an external runner installation conflict, not a demonstrated candidate source failure.
3. A specific-job rerun was attempted and refused while the workflow was active. Remaining jobs were cancelled; the fresh terminal read is completed/cancelled. No local background command or pending observation is left running.
4. No shared runner process was killed, no shared binary or filesystem was modified, and no CI bypass or workflow edit was attempted. Existing full-suite checklist item stays unchecked.

Constraint: required hosted validation cannot be declared green while the pinned linter cannot be installed on the runner. The assumption that storage repair restored every required job was disproved by direct provider logs.

Options: (a) runner owner ensures the pinned linter installation destination is not concurrently executed/overwritten, then reruns failed/cancelled jobs on the same SHA; (b) runner owner supplies an approved isolated runner with the workflow's required platform/toolchain. A local lint pass cannot replace the required hosted workflow result, and a candidate/workflow workaround is outside this gate-only assignment.

Recommendation and exact external input: restore conflict-free pinned-linter installation on the runner (or provide equivalent isolated capacity), then retry failed/cancelled jobs on `233c8873889a9f25cd0e2d040a39bcf564a6540c`. No product or source change is requested. Full suite and landing properties remain unverified until the actual hosted gate succeeds.
