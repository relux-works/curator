# TASK-260908-2g6x65 — A1 CLI delivery (SPEC §3) — producer results

Worktree `.temp/STORY-260908-xadoax/worktree`, checkpoint `484933b`, work left uncommitted per the story workspace contract. Ready for review.

## Delivered

- `internal/cli/cli.go` — the §3 parser. `cli.Parse(args, cli.Options{AxConfigured})` returns `cli.Invocation` (EnvID, Profile/Model/Effort/Name with *Set markers, SystemPrompt and AxProfile closed enums, Tracked, Native verbatim tail, Info for help/version) or `*cli.UsageError` (code `usage`, `cli.ExitCode = 2`). Ax-configured is caller-supplied; no ax.json read, no registry validation of env-id, no profile resolution, no process start.
- `cmd/curator-run/main.go` — `run()` wires `cli.Parse` with `AxConfigured: false` (ax.json family not yet delivered); usage errors print `curator-run: usage: <detail>` + usage text on stderr, exit 2; `--help/-h/--version` print `curator-run 0.0.0-dev (specification 0.2.1-draft)` on stdout, exit 0; a parsed launch is refused `not_implemented`, exit 1, nothing resolved or launched.
- `.github/workflows/ci.yml` — initial CI: fmt-check, build, vet, test, race on ubuntu-latest + macos-latest, `actions/checkout@v4`, `actions/setup-go@v5` with `go-version-file: go.mod` (convention of sibling `curator/.github/workflows/ci.yml`). No release/tag.
- `Makefile` — added `fmt-check`, `race`; `check` runs all. `README.md` — status now "partially implemented (§3 only)", tools table, CI note.
- `.scripts/cli-mutants.sh` — narrowing-mutant harness (14 mutants), restores the tree on exit.
- `internal/cli/testdata/cases.golden` — 30 accepted + 43 rejected shapes rendered through `Parse`.

## Parser rules chosen where §3 left detail open (documented on `Parse`)

- Informational flags: first one met left-to-right wins immediately; a usage error met earlier wins over a later `--help` (line read once, in order). `--help --version` → help; `--version --help` → version; `codex_cli --version --garbage` → version.
- Values: `--flag value` and `--flag=value`; a value that is absent, `--`, empty, or begins with `-` is "requires a value" (so `--profile --model m` does not swallow a flag as a value).
- Empty operand `""` before `--` is a usage error; a lone `-` is an unknown flag; a second `--` after the first is native.
- `--name` on untracked machine: validated, stored (`NameSet`), `Tracked=false` marks it ineffective. `--ax-profile` on untracked: vocabulary checked first, then the untracked usage error.

## AC coverage — 16 of 16 named behaviors driven through a production call site

Production call sites: `cli.Parse` (internal/cli/cli.go) and `run` (cmd/curator-run/main.go, `main` exits with its value).

| # | §3/§6 behavior | Test |
|---|---|---|
| 1 | post-`--` verbatim incl. empty operands, colliding flags, second `--` | `TestNativeTailVerbatim`, `TestParseAccepted/native_tail_*`, `TestRunParsedLaunchRefused` |
| 2 | repeated value flag → usage | `TestParseRejected/repeated_*` (6 flags), `TestRunUsageErrorsExit2` |
| 3 | unknown pre-`--` flag → usage, never forwarded | `TestParseRejected/unknown_*`, `lone_dash`, `triple_dash`, `TestRunUsageErrorsExit2` |
| 4 | missing value | `TestParseRejected/missing_value_*`, `value_spelled_as_flag`, `value_is_help_flag` |
| 5 | missing `<env-id>` | `TestParseRejected/no_arguments`, `only_double_dash`, `flags_only_no_env`, `native_tail_no_env` |
| 6 | stray operand | `TestParseRejected/stray_operand*` |
| 7 | `--system-prompt` closed vocabulary | `TestParseRejected/system_prompt_*`, `TestParseAccepted/system_prompt_*` |
| 8 | `--ax-profile` closed vocabulary | `TestParseRejected/ax_profile_bad_value_*`, `ax_profile_case_sensitive` |
| 9 | `--ax-profile` on untracked machine → usage | `TestParseRejected/ax_profile_(yolo_)on_untracked_machine`, `TestRunUsageErrorsExit2` |
| 10 | `--name` accepted-but-ineffective untracked | `TestNameIneffectiveUntracked`, `TestParseAccepted/name_on_untracked_machine_accepted_no_effect` |
| 11 | `--name` ax §2.1 grammar, 64-char bound | `TestParseRejected/name_*` (8), `TestParseAccepted/name_64_chars_on_tracked_machine` |
| 12 | informational flags deterministic, exit 0 | `TestParseAccepted/help_*`, `version_*`, `TestInformationalCarriesNothingElse`, `TestRunInformationalFlags` |
| 13 | usage exits 2, code line + usage text, nothing launched | `TestRunUsageErrorsExit2`, `TestUsageErrorShape` |
| 14 | env-id not registry-validated in parser | `TestParseAccepted/env-id_lookalike_is_not_validated_here` |
| 15 | goldens for accepted/rejected shapes | `TestGolden` (73 shapes) |
| 16 | spec version reported with build version | `TestSpecVersionPinned` |

## Negative evidence — narrowing mutants, 14 of 14 killed

`.scripts/cli-mutants.sh .temp/TASK-260908-2g6x65/mutants`, harness exit 0, full `go test ./... -v` per mutant (log `mutants-run-01.log`, per-mutant `M01..M14.log`, `summary.tsv`).

| Mutant | Gate narrowed to admit exactly one member | Killed by |
|---|---|---|
| M01 | repeat allowed for `--profile` only | `TestParseRejected/repeated_profile` |
| M02 | unknown flag `--nope` skipped | `TestParseRejected/unknown_flag` |
| M03 | empty value admitted | `TestParseRejected/missing_value_empty_string` |
| M04 | stray operand `resume` admitted | `TestParseRejected/stray_operand` |
| M05 | `--system-prompt prepend` admitted | `TestParseRejected/system_prompt_bad_value` |
| M06 | `--ax-profile unsafe` admitted | `TestParseRejected/ax_profile_bad_value_tracked` |
| M07 | untracked gate admits `standard` | `TestParseRejected/ax_profile_on_untracked_machine` |
| M08 | name length bound 64→65 | `TestParseRejected/name_65_chars` |
| M09 | name may start with `.` | `TestParseRejected/name_leading_dot` |
| M10 | empty argv passes the env-id gate | `TestParseRejected/no_arguments` |
| M11 | empty native operands dropped | `TestNativeTailVerbatim` |
| M12 | `--nope` tolerated before `--help` | `TestParseRejected/unknown_flag_before_help` |
| M13 | usage exit 2→1 (main) | `TestRunUsageErrorsExit2` |
| M14 | parsed launch exit 1→0 (main) | `TestRunParsedLaunchRefused` |

No gate inspects source text, so the token-preserving mutant rule is not applicable (stated bound).

## Validation (all rerun by me at the final tree)

| Command | Exit | Log |
|---|---|---|
| `make check` (build, fmt-check, vet, test, race) | 0 | `.temp/TASK-260908-2g6x65/make-check-01.log` |
| `.scripts/cli-mutants.sh` | 0 (14/14 KILLED) | `.temp/TASK-260908-2g6x65/mutants-run-01.log` |
| built binary smoke: `--version` 0; launch 1; stray/ax-profile/name/no-args/unknown 2 | as listed | `.temp/TASK-260908-2g6x65/binary-smoke-01.log` |

Lint: `gofmt -l` clean, `go vet` clean. No golangci-lint config exists in this repo; not added (out of scope, A3 owns release readiness). Hosted CI has not run yet: the workflow is new and lands with the parent's PR; green checks are the parent's to observe.

## Bounds and notes for the reviewer

- `AxConfigured` is hard-wired `false` in `main` until the §4.6/§4.7 ax.json story lands; the SPEC rule "configuration read fires before a usage error" is therefore not yet observable at the binary and belongs to that story.
- `--model`/`--effort` flag-vs-locked-default usage error (§4.3) is the defaults story's: the parser exposes `ModelSet`/`EffortSet` for it.
- Erratum F1 (`argv_suffix` wording) from the A0 review is composition-stage and untouched here.
- No logbook entry: no anomaly or decision beyond the documented parser rules.
