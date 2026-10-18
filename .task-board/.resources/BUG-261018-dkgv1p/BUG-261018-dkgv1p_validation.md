# BUG-261018-dkgv1p — subcommand-help-flag-dispatches-command

Status: ready for review. Full hosted attempt 4 passed on the exact unchanged signed candidate: 20 jobs success, 2 optional jobs skipped, no failures or cancellations. See `BUG-261018-dkgv1p_hosted-retry-04.md` for commands, exit codes and every job result. Earlier infrastructure failures below remain historical failures, not reviewer acceptance.

## Candidate

- Story checkpoint: `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`; fresh advertised origin HEAD and fetched origin/main matched it before implementation.
- Signed detached validation commit: `233c8873889a9f25cd0e2d040a39bcf564a6540c`; `git log -1 --format='%G?'` returned `G`.
- Candidate tree: `796a42f78bf9fdad6dedd31e571653be096577a8`.
- Story branch stays at its checkpoint with the candidate UNCOMMITTED, as required by the managed Story assignment. Integration and branch commits belong to the orchestrator.
- Go `1.25.5`, native `darwin/arm64`; golangci-lint `2.12.2`.

## Change

The built-in registry now drives top-level dispatch and 72 command/help topics, including nested verbs. `cli.run` answers help before invoking a handler or loading config. `help <command>` shares the dedicated renderer. Help recognition stops at `--`; `run` retains its real provider dispatch and unchanged passthrough operands. Existing help tests retain their original names and now assert the required exit-0 stdout contract. No reviewer test was removed.

Changes are limited to `cmd/curator` plus `README.md` and `docs/cli.md` to document the changed help behavior and test commands. This is the first producer revision, not review rework.

## Commands and real exit codes

All test/build/lint commands ran as standalone processes, with output redirected to task-local logs, without `tee` or an exit-masking pipeline.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run 'Test(RegisteredSubcommandHelpNoWrites\|CommandHelp.*\|GlobalInit.*Help.*\|Draft.*Help.*\|ConfigSubcommandDispatch)$' -count=1 -v` | 0 | `help-tests-05.log`; 72 topics, 490 help invocations; init, passthrough and mutant controls green |
| `GOLANGCI_LINT_CACHE="$PWD/.temp/BUG-261018-dkgv1p/lint-cache" make lint` | 0 | `lint-05.log`; 0 issues |
| `go build ./...` | 0 | `build-04.log` |
| `go vet ./...` | 0 | `vet-04.log` |
| `git diff --check` | 0 | `diff-check-01.log` |
| Hosted full CI | 0 | [workflow 39067488895, attempt 4](https://github.com/relux-works/curator/actions/runs/39067488895/attempts/4), exact signed candidate above; 20 jobs success, 2 optional skipped |

`CURATOR_TEST_CLI_HELPER=1` is the repository's existing TestMain protocol for child tests that do not mutate or fingerprint GOROOT; git isolation remains active. It avoids the shared host toolchain lock for these narrow CLI tests. Full hosted jobs use their normal workflow environment and suite selection.

The configured full gate is `sh scripts/remote-gate.sh`. To respect the headless bounded-command rule, its signed detached snapshot publication and hosted workflow observation were executed in separate bounded commands. No altered CI status is created. Full validation is now green in attempt 4; the retry-04 outcome records the terminal evidence. The following attempts remain historical failures. Attempts 1 and 2 both have required Linux job failures (`no space left on device`); remaining jobs were cancelled after the repeated external failure. See the attached `BUG-261018-dkgv1p_blocker.md` for job IDs, retry results and exact recovery input.

## Behavioral acceptance coverage

**5 of 6 AC rows driven through the production CLI by named tests.** AC6 is an operational validation/signature requirement, verified by the commands and hosted workflow above, rather than a CLI behavior test.

| AC | Named test | Production call site / assertion |
| --- | --- | --- |
| 1 | `TestRegisteredSubcommandHelpNoWrites` | Built binary `main -> run -> cli.run`; both flags at every command-path position and after ordinary/invalid operands, before `--`; exit 0, dedicated output, empty stderr, equal before/after tree paths, sizes, modes and file hashes across cwd/HOME/XDG_CONFIG_HOME/CURATOR_HOME/TMPDIR |
| 2 | `TestRegisteredSubcommandHelpNoWrites` | Built binary `main -> cli.run` explicit help branch; flag output must equal `help <command>` byte-for-byte |
| 3 | `TestRegisteredSubcommandHelpNoWrites` | Enumerates the same `commandRegistry` used by `cli.run` for built-in dispatch; duplicate names fail; covers all 72 topics without a separate command-name list |
| 4 | `TestCommandHelpRunPassthrough` | Built binary `main -> cli.run -> runProvider -> cmdUmbrella`; native provider returns actual argv as JSON, asserting literal `--`, `-h`, `--help`, multiple operands and help before the boundary |
| 5 | `TestCommandHelpNarrowingMutants/init-bypasses-help` | Go overlay weakens the production guard for `init` only; nested table runs rebuild and execute the mutated production CLI, and must fail with child exit 1 and the named behavioral failure |
| 6 | Operational validation | Full hosted attempt 4 passed all mandatory jobs, including suite/race/lint; signed detached snapshot verified; no producer commit on the managed Story branch |

Positive control: `TestCommandHelpInitPositiveControl` drives `main -> cli.run -> cmdInit` without help and asserts nonempty `Skillfile.json` and `.gitignore`. Its top-level help/version controls assert zero persistent writes too. `TestCommandHelpUnknownTopic` proves an explicit unknown help topic is an exit-2 usage error through `run -> cli.run`.

## Narrowing mutant evidence

Each guard remains present. Only child exit 1 plus the named behavioral failure counts as a kill. The parent mutant suite exits 0 only after those expected-red outcomes are observed. All three were killed; there are no surviving mutants.

| Mutant | What it narrows the gate to | Named test that fails | Child exit / survivor bound |
| --- | --- | --- | --- |
| `init-bypasses-help` | Every built-in except `init` still gets the dispatch guard | `TestRegisteredSubcommandHelpNoWrites/init/1` | 1; killed |
| `only-leading-flags` | Help recognized only when the first operand is a help flag | `TestRegisteredSubcommandHelpNoWrites/init/2` | 1; killed |
| `help-crosses-separator` | The help scan continues past `--` | `TestCommandHelpRunPassthrough` | 1; killed |

Reproduce all controls: `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestCommandHelpNarrowingMutants$' -count=1 -v` (exit 0). Inner commands are `go test -count=1 -timeout=3m -v -run '^TestRegisteredSubcommandHelpNoWrites$/^init$' .` for the first two, and `go test -count=1 -timeout=3m -v -run '^TestCommandHelpRunPassthrough$' .` for the third, with a private GOFLAGS overlay inherited by the CLI build. Each inner command exits 1. The overlays never edit production files.

## Bounds and out-of-contract rows

- **Brief gap:** no attack-surface table was supplied. No table-derived coverage map is claimed or silently waived; the AC map and mutant table above record the actual coverage.
- Snapshot equality proves persistent paths, sizes, modes and file bytes; it cannot observe transient writes removed before exit. Timestamps, access times, syscall auditing and writes outside the five isolated roots are not measured by this test.
- External dynamically discovered provider names and hidden internal worker/shim modes are outside AC1's registered subcommand set. The registered `run` provider boundary is exercised; independent external provider help remains provider-owned.
- No other AC rows are treated as out of contract. AC6 is now satisfied by attempt 4 and the verified signed candidate.
- The help gate examines argv, not source text. The source-text-gate mutation checklist row is inapplicable; overlays exercise behavior in the real built CLI.
- Existing reviewer tests are retained by original name. Two old tests intentionally changed expectations that contradicted the new help contract (config refusal and FlagSet exit 2); this is not a waiver of non-help behavior.

## Task-scoped logbook

1. Root cause: top-level dispatch passed `init -h` into `cmdInit`; `projectRootArg` ignored the flag and initialization wrote the project files.
2. First local targeted invocation was terminated by the producer while waiting for the shared GOROOT lock: real exit 143, no green evidence. Later narrow invocations use the established helper protocol described above.
3. Initial mutant harness invocation exited 1 because GOFLAGS quoted only the overlay value. This was a setup failure, not a killed mutant. Quoting the entire overlay argument fixed it; all subsequent controls produced the required named behavioral failures.
4. Initial `make lint` exited 2 (underlying linter failure): one new De Morgan simplification plus stale cached paths/issues from another removed worktree. The new condition was corrected; a private lint cache gave 0 issues. No foreign cache or worktree was edited.
5. Two earlier hosted snapshots were cancelled by the producer after self-review corrected help operands and retained the original nested subtest names; both are superseded and are not validation evidence. The current signed snapshot and workflow are identified above. Its own two attempts encountered repeated external Linux ENOSPC failures; remaining jobs were cancelled after that evidence was captured.
6. Public outcome contains only synthetic test details and repository-relative evidence names. Raw local logs remain ignored under `.temp/BUG-261018-dkgv1p/`; no direct control-root logbook edit was made.

7. Hosted attempt 4 passed on the unchanged signed candidate after the runner HOME repair: 20 successful jobs, 2 optional skipped, 0 failures. See `BUG-261018-dkgv1p_hosted-retry-04.md`; no source/test/configuration change was made during that retry.
