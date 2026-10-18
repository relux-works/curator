# BUG-261018-dkgv1p — full hosted gate retry after runner HOME isolation

Status: ready for review. Attempt 4 of the full hosted workflow passed on the unchanged signed candidate: **20 jobs success, 2 optional jobs skipped, 0 failures or cancellations**. The previous infrastructure failures remain historical failures, not passing evidence.

## Candidate and scope

- Signed candidate: `233c8873889a9f25cd0e2d040a39bcf564a6540c`; `git verify-commit` exited 0 before dispatch and after the terminal verdict.
- Candidate tree: `796a42f78bf9fdad6dedd31e571653be096577a8`; a private-index reconstruction matched before dispatch and again after all jobs passed.
- Managed Story HEAD stays at checkpoint `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`. All candidate changes remain UNCOMMITTED in the managed worktree. No source, test, configuration, branch or index change was introduced by this retry.
- Prior local CLI, positive-control, passthrough, mutant, build, vet and lint evidence is reused from `BUG-261018-dkgv1p_validation.md`, bound to this unchanged candidate. Those commands were not rerun locally in this gate-only assignment. Hosted jobs were actually rerun with the workflow's configured toolchains and environments.

## Commands and real exit codes

[Full CI workflow, attempt 4](https://github.com/relux-works/curator/actions/runs/39067488895/attempts/4).

| Command / process | Real exit / result | Meaning |
| --- | ---: | --- |
| `gh run rerun 39067488895 --repo relux-works/curator` | 0 | Full workflow rerun admitted on the original SHA |
| `git verify-commit 233c8873889a9f25cd0e2d040a39bcf564a6540c` | 0, twice | Existing candidate signature remains valid |
| Private-index `git read-tree HEAD`, `git add -A .`, `git write-tree` | 0, before and after | Exact candidate tree; normal index untouched |
| `gh run view 39067488895 --repo relux-works/curator --json attempt,headSha,headBranch,status,conclusion,url,jobs` | 0 | Fresh terminal provider evidence: attempt 4, completed/success, exact SHA |
| `gh run watch 39067488895 --repo relux-works/curator --exit-status` | 0 | Terminal full-gate verdict |
| `git diff --check` | 0, before and after | Candidate remains whitespace-clean |
| Direct active-job log read | 1; HTTP 404 | Logs were unavailable while active; no test verdict inferred |
| Initial read-only workflow observer | 143 | Observer intentionally stopped; hosted workflow was not cancelled |
| Two bounded read-only observers | 124 each | Observation windows expired; hosted workflow continued; these are not test passes |

All validation and observation commands ran as standalone processes with direct output redirection, without `tee` or status-masking pipelines. The Python identity/terminal assertions exited 0. No new commit, push, local full-suite replay or CI-status override was performed.

## Hosted job results

Every mandatory job passed. The two skipped jobs are configuration-dependent optional lanes, not passing test claims.

| Job | Result | Provider job |
| --- | --- | --- |
| Naming gate | success | [117509494941](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509494941) |
| Lint | success | [117509495043](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495043) |
| Test (ubuntu-latest) | success | [117509495070](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495070) |
| Gate self-test (windows-latest) | success | [117509495083](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495083) |
| Gate self-test (ubuntu-latest) | success | [117509495094](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495094) |
| Interop conformance gate | success | [117509495095](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495095) |
| Go 1.26.0 driver (windows-latest) | success | [117509495113](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495113) |
| Test (windows-latest) | success | [117509495117](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495117) |
| Race (ubuntu-latest) | success | [117509495119](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495119) |
| Go 1.27.1 driver (windows-latest) | success | [117509495132](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495132) |
| Gate self-test (macos-latest) | success | [117509495143](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495143) |
| Go 1.26.0 driver (macos-latest) | success | [117509495164](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495164) |
| Test (macos-latest) | success | [117509495167](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495167) |
| Go 1.25.5 driver (windows-latest) | success | [117509495178](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495178) |
| Go 1.25.5 driver (ubuntu-latest) | success | [117509495183](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495183) |
| Race (macos-latest) | success | [117509495189](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495189) |
| Go 1.26.0 driver (ubuntu-latest) | success | [117509495196](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495196) |
| Go 1.27.1 driver (macos-latest) | success | [117509495204](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495204) |
| Go 1.27.1 driver (ubuntu-latest) | success | [117509495210](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495210) |
| Go 1.25.5 driver (macos-latest) | success | [117509495251](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495251) |
| Optional self-hosted Test lane (disabled by configuration) | skipped | [117509495885](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509495885) |
| Candidate suite (${{ matrix.os }}) | skipped | [117509496015](https://github.com/relux-works/curator/actions/runs/39067488895/job/117509496015) |

The full Test lanes run `go vet`, ledger consistency, and `bash .github/ci/test-gate.sh` on Linux, macOS and Windows. Both Race lanes run that test/platform gate with `GO_TEST_FLAGS=-race`. Lint includes pinned golangci-lint and the workflow configuration gates. All nine Go driver qualification/build jobs passed. This supplies the previously missing full-suite and landing-gate property evidence.

## Acceptance coverage and bounds

**5 of 6 AC rows driven through the production CLI by named tests; AC6 is separately satisfied by the signed candidate and the green full hosted workflow.** The detailed production call-site map remains in `BUG-261018-dkgv1p_validation.md`: `main -> run -> cli.run` for AC1–3 and AC5, and `cli.run -> runProvider -> cmdUmbrella` for AC4. The tests enumerate 72 registered help topics and 490 help invocations; the positive init control proves command execution still writes the required files without help.

- No surface table was supplied in the brief. This remains a brief gap, not a claim of table-derived coverage.
- No reviewer test was removed or renamed by this retry; candidate identity is unchanged and all hosted suites passed.
- No rework diff was introduced. Original scope stays in `cmd/curator`, plus the two CLI documentation files already identified in the validation report.
- Out-of-contract rows remain external dynamically discovered provider names and hidden worker/shim modes, under AC1's limitation to registered subcommands. The registered provider passthrough boundary remains tested under AC4. No other AC row is excluded.
- Filesystem snapshots prove persistent paths, sizes, modes and bytes in the isolated roots; they do not observe transient writes removed before exit, timestamps or writes outside those roots.
- The help gate inspects argv, not source text. The conditional source-text gate checklist row was reviewed as **not applicable**, rather than claimed as a newly executed static-gate mutant. Existing mutants preserve and narrow the behavioral guard.

## Reused narrowing mutant evidence

These controls were previously executed on this exact candidate and attached in the validation report. They are not claimed as a new local run here. There are no reported survivors.

| Mutant | What it narrows the gate to | Named failing test | Result / survivor bound |
| --- | --- | --- | --- |
| `init-bypasses-help` | Guard remains for every built-in except `init` | `TestRegisteredSubcommandHelpNoWrites/init/1` | Child exit 1; killed |
| `only-leading-flags` | Only a leading help operand is recognized | `TestRegisteredSubcommandHelpNoWrites/init/2` | Child exit 1; killed |
| `help-crosses-separator` | Help scan crosses `--` | `TestCommandHelpRunPassthrough` | Child exit 1; killed |

## Task-scoped logbook

1. The runner operator's HOME isolation repair was followed by a full rerun of the exact original signed candidate. Lint installed and executed successfully; the prior binary-extraction conflict did not recur.
2. Every required job reached success, including all three full Test lanes and both Race lanes. No runner infrastructure failure or source-caused failure occurred in attempt 4.
3. All observations were kept within bounded foreground calls. Read-only observers were stopped or expired without cancelling the hosted run. No background observer remains running.
4. Candidate tree and Story checkpoint were checked again after terminal success. The worktree remains ready for the managed handoff snapshot, with no producer commit on the Story branch.
5. This outcome supersedes the active blocker status recorded in earlier reports. Historical ENOSPC and linter-installation failures remain failures; they are not relabeled as green. No direct control-root logbook edit was made.
