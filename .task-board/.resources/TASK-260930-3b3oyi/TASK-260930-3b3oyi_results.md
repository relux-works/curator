# TASK-260930-3b3oyi — rework 1, response to review revision 2

All repository changes remain uncommitted in the assigned story worktree. No CHANGELOG or LOGBOOK file edits.

## Findings addressed

F1: `cli.childEnviron` replaces inherited markers with `CURATOR_INTERNAL_POSTURE_WARNED=<own pid>` once the invocation has warned. `cli.loadConfig` honours it only when it equals `os.Getppid()` and that PID is greater than 1. Accepted ancestry marks the invocation warned so direct descendants can rebind the marker. User-set `1` and a different PID both warn. The prescribed bound is direct curator dispatch: an intervening shell (`curator run → shell → curator`) breaks the binding and the grandchild warns again, as the production comment states. This is an immediate-parent binding, not cryptographic authentication; a caller deliberately supplying its actual PID can reproduce the binding.

Named regression: `TestPostureWarningOncePerProcessTree`, including `user-set-one`, `not-parent-pid`, and real native `curator run env status --json` dispatch. Binary rows pin the exact warning and exit 0; the nested row requires child JSON. No Windows skip or weakened assertions.

F2: `wantsHelp` considers only the group operand (`--help`, `-h`, `help`). Group help works without config. Subcommand help reaches its original FlagSet after config loading; without config it refuses with the bootstrap hint and exit 1. Documentation explicitly distinguishes these paths.

Named regressions: `TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet` (exact original FlagSet stderr and exit 2 for both `env resolve --help` and `global add --help`) and `TestGlobalInitSubcommandHelpWithoutConfigRefuses` (five refusal rows). The old generic-help rows were replaced with group-level rows.

## Production-entry evidence

Built `/tmp/TASK-260930-3b3oyi-curator` from the candidate. Probes used throwaway HOME and native/config locations. Raw argv, stdout, stderr and subprocess exit codes are attached as `TASK-260930-3b3oyi_production-rework1.json` and `TASK-260930-3b3oyi_scope-rework1.json`.

| Row | Exit | Observed behavior |
| --- | --- | --- |
| `global init`, clean HOME | 1 | Exact absence line and bootstrap hint below |
| `global --help`, `profile --help`, `env --help`, no config | 0 each | Group usage on stdout, stderr empty |
| `env resolve --help`, no config | 1 | Same absence diagnostic and hint |
| Documented bootstrap example, with `--skills-root "$HOME/skills"` | 0 | Config written |
| `global init` after bootstrap | 0 | Global Skillfile initialized |
| `env resolve --help`, with config | 2 | `Usage of env resolve:` and original flags, command and alias usage |
| `global add --help`, with config | 2 | `Usage of global add:` and original flags |
| `env resolve codex_cli --profile acme`, fresh profile | 1 | Exact stale diagnostic below |
| `env status --check`, current acme has one unprovisioned home | 1 | Current-scope finding fails check |
| `env status --check`, repaired acme, unused default unprovisioned | 0 | All default findings remain printed |
| User marker `1`, standalone `env status --json` | 0 | Exactly one posture warning |
| Non-parent PID marker, standalone `env status --json` | 0 | Exactly one posture warning |
| Native `curator run env status --json` | 0 | Exactly one posture warning across parent and child, child JSON present |

Exact newline-terminated missing-config stderr (path normalized only):

```text
curator: global config not found: <HOME>/.curator/config.json
curator: run `curator bootstrap --skills-root <dir>` first (see docs/cli.md#bootstrap)
```

Exact newline-terminated stale diagnostic, following the posture warning:

```text
curator: environment_home_stale: home unprovisioned; rerun with --repair
```

Exact newline-terminated posture warning:

```text
warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults before revision B flips the default
```

## Commands run by this developer

- `go test -p 1 ./cmd/curator -count=1 -timeout 4m -v -run '^(TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses|TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestPostureWarningOncePerProcessTree)$'`: exit 0, 18.766 s, 4/4 top-level tests passed.
- Required broad mask: `go test -p 1 ./cmd/curator -count=1 -timeout 9m -run 'Bootstrap|GlobalInit|EnvStatus|EnvResolve|Posture'`: exit 1, 540.678 s, timeout in existing `TestEnvStatusMissingAndUnreadableKeepRecord`; no assertion failure reported. This command is failing and is not claimed green.
- `go test -p 1 ./internal/config ./internal/envprofile -count=1 -timeout 5m -run 'Status|Stale|NonCurrent|Resolve'`: exit 1; config passed in 0.857 s; envprofile timed out in 301.159 s during `TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer/absent_marker_is_known_unprovisioned`. No assertion failure reported; not claimed green.
- Bounded internal rerun: `go test -p 1 ./internal/envprofile -count=1 -timeout 3m -v -run '^(TestResolveStaleUnprovisioned|TestResolveProvisionRepair|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusProfileMembersAndUnregistered|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent)$'`: exit 0, 43.669 s, 5/5 tests passed, including the test active at the timeout.
- `GOOS=windows go vet -p 1 ./cmd/curator ./internal/...`: exit 0.
- `GOOS=windows go test -p 1 -c ./cmd/curator -o /tmp/TASK-260930-3b3oyi-curator.test.exe`: exit 0; compile check, not Windows runtime.
- `go build -p 1 ./...`: exit 0.
- `go build -p 1 -o /tmp/TASK-260930-3b3oyi-curator ./cmd/curator`: exit 0.
- `golangci-lint run`: exit 0, 0 issues.
- `git diff --check`: exit 0.
- `test -z "$(gofmt -l cmd/curator/main.go cmd/curator/firstrun_ux_test.go)"`: exit 0.
- `bash .github/ci/no-broad-suppression.sh`: exit 0.

## Evidence limits and history

Windows runtime and the full hosted gate were not run by this developer on this macOS host. The prior review verified the Windows fix and reported the previous candidate hosted gate green; that is accepted prior evidence, not a claim that this rework is hosted-green. The Windows row still uses a native manager binary, with no skip. Prior gate-fix artifacts already preserve run 36778472331: Go tests exited 0, platform-case gate exited 1 due to the old Windows posture-row skip. Those artifacts were not re-downloaded here. The hosted gate remains the arbiter. Broader local masks timed out under host load; only named bounded reruns are claimed green. No full protocol-conformance result is inferred from these checks.

The first mutation attempt launched before source-copy setup returned: exit 1, `go.mod file not found`; no mutant proof. The first `-timeout 3m` mutation attempt exited 1 after 240.106 s while waiting for the cross-process host GOROOT lock in TestMain, before running assertions; also no mutant proof. Its bounded rerun exited 1 in 59.820 s with actual assertion failures in both forged-marker rows (see mutant evidence below). No lock bypass or assertion changes. An unnecessary repeat with `-timeout 9m` was launched before this successful mutant log was inspected; its own waiting test child was terminated after the assertion-based mutant result was recovered. The repeat returned exit 1 (`signal: terminated`), and is not mutant evidence.

## Entry text (no CHANGELOG/LOGBOOK edit)

Bound posture-warning suppression to the dispatching immediate parent PID; generic and unrelated-PID markers retain the warning. Restored subcommand flag help and exit 2 after bootstrap while keeping group help available without config. Direct curator descendants share one warning; a shell-interposed grandchild warns again. Added exact production-entry regressions for forged markers and help routing, retaining the native Windows row.

## Narrowing mutants

Mutants run in a separate `/tmp/TASK-260930-3b3oyi-mutant-rework1` source copy; the candidate remains untouched.

F1 mutant M-posture-nonempty: replace the parent-bound predicate in production `cli.loadConfig` with `os.Getenv(postureWarnedEnv) != ""`. This narrows the warning's coverage to unmarked invocations while retaining suppression and ordinary/nested controls. `go test -p 1 ./cmd/curator -count=1 -timeout 3m -v -run '^TestPostureWarningOncePerProcessTree$'`: exit 1, 59.820 s. Both `user-set-one` and `not-parent-pid` FAIL on their exact-warning assertions (stderr was empty); the nested and standalone controls still succeeded. Killed negative rows: 2/2. Raw evidence: `TASK-260930-3b3oyi_mutant-posture-rework1.log`.

F2 mutant M-help-subcommand-interception: replace the group-only return in `wantsHelp` with a scan for help anywhere in the group argv, reproducing the rejected routing and narrowing the available per-subcommand help. With F1 restored in the mutant copy, `go test -p 1 ./cmd/curator -count=1 -timeout 9m -v -run '^(TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestGlobalInitSubcommandHelpWithoutConfigRefuses)$'`: exit 1, 61.314 s. Both configured FlagSet rows FAIL (exit 0/group stdout instead of exit 2/FlagSet stderr); the config-free refusal test also FAILS. Killed configured help rows: 2/2. Raw evidence: `TASK-260930-3b3oyi_mutant-help-rework1.log`.

## Clean candidate task-scope rerun

`go test -p 1 ./cmd/curator -count=1 -timeout 9m -v -run '^(TestGlobalInitCleanHomeRefusesWithBootstrapHint|TestGlobalInitMalformedConfigHasNoBootstrapHint|TestGlobalInitHelpWithoutConfig|TestGlobalInitSubcommandHelpWithoutConfigRefuses|TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet|TestBootstrapDocsExampleRuns|TestEnvResolveStaleHintsRepair|TestEnvStatusCheckCurrentScopeOnly|TestPostureWarningOncePerProcessTree)$'`: **exit 0**, 68.344 s, **9/9** top-level first-run tests passed. This is the unmutated story candidate after both mutation checks; it is not a passing rerun of the broader mask. Raw log: `TASK-260930-3b3oyi_targeted-rework1.log`. Combined with the bounded internal check, the task-scope test evidence is **14/14** top-level tests green. Checklist “Tests green” refers to these explicit task-scope runs; the broader timeout commands remain red and are reported above. Both mutant rejections are expected-red exit 1, caused by the specified regression assertions, not reported as passing tests.

Source SHA-256 at handoff preparation:
- `cmd/curator/main.go`: `71c06a5e6a525a7fa6e7fa27ee82d9b1d4ecc234e79ae44da1c8307ab99fc5ba`
- `cmd/curator/firstrun_ux_test.go`: `56c30f1f683a93c057aa088b90c9f1e40f008335bc3e33acda2c7cc9b8b8bebb`
- `docs/cli.md`: `b0163f184712ad756f10c0b334ee05ed4665ed602dc15b98aadc799803fb1bc9`

Review revision 2's rejection is already attached on the board; this revision answers both findings and routes through developer handoff to review. The new validation, production-entry and mutation artifacts preserve this run's measured evidence.
