# Review verdict \u2014 TASK-260908-2g6x65, CR-TASK-260908-2g6x65-1 revision 1

Verdict: **accepted**. Reviewer: tracked reviewer run, 2026-09-08. Base `484933b`, candidate tree `ae48181`, worktree `.temp/STORY-260908-xadoax/worktree`. Toolchain go1.25.5 darwin/arm64 (CI pins go.mod `go 1.23` via setup-go@v5).

## Reran by the reviewer (exit codes real)

| Gate | Command | Exit |
|---|---|---:|
| build + gofmt + vet + test + race | `make check` (log `.temp/TASK-260908-2g6x65/make-check-01.log`) | 0 |
| mutant harness, 14 narrowing mutants | `.scripts/cli-mutants.sh .temp/TASK-260908-2g6x65/mutants` | 0, 14/14 KILLED, source restored byte-identical (cmp), `git status` unchanged |
| binary probes | built `cmd/curator-run`, 10 adversarial argv shapes | as expected (below) |

Probes at the production entry point (`main -> run -> cli.Parse`): `--help`/`--version` exit 0 with build+spec version; `--profile=x --profile y` \u2192 repeated-flag usage exit 2 (mixed forms detected); `--ax-profile yolo` untracked \u2192 usage exit 2; `--name=--x` and `--model -x` \u2192 missing-value usage exit 2; `--profile p --version` \u2192 info wins, exit 0; `--help=1` \u2192 unknown flag exit 2; `codex_cli -- --help` \u2192 tail uninspected, refused with exit 1 `not_implemented`, nothing launched.

## AC coverage: 13 of 13 SPEC \u00a73 rows driven

All at call site `cli.Parse` (unit + golden) and `run` (cmd tests): flags only before `--` (native tail keeps `--help`, `--`, empties, colliding flags: `TestNativeTailVerbatim`, goldens); first operand is env-id / second is stray (`stray operand*`); unknown pre-`--` flag never forwarded (`unknown flag*`, incl. before `--help`); repeated flag never last-wins (6 repeated_* cases); closed vocabularies incl. case sensitivity; `--name` grammar at parse time incl. 64/65 boundary, leading `.`/`_`, slash, space, unicode, newline; `--ax-profile` untracked usage error and `--name` untracked accepted-no-effect (`TestNameIneffectiveUntracked`); usage exit 2 with `usage` code line and usage text, no stdout (`TestRunUsageErrorsExit2`); informational exit 0 (`TestRunInformationalFlags`); env-id not registry-validated (`not_registered` accepted, per brief).

Negative evidence: 44 rejected shapes, every gate has a narrowing mutant (M01\u2013M12 parser, M13 exit code, M14 refusal exit) and the harness runs the whole behavioral suite, requiring the named test in the failures. No source-text-inspecting gate exists, so that DoD row does not apply.

## Side effects / honesty

Parser reads no file, starts no process; `run` only writes to the given writers. README status and package docs state \u00a73-only delivery, exit-1 `not_implemented` refusal, and that `--ax-profile` is always a usage error until ax.json (\u00a74.6) lands. `--version` reports build and spec version per \u00a78. CI: build/gofmt/vet/test/race on ubuntu+macos, checkout@v4 / setup-go@v5 / `go-version-file: go.mod`, matching sibling curator ci.yml conventions; no tag/release step. Makefile extended, not replaced. No LOGBOOK, no commits, no real ax.

## Notes (non-blocking, for the parent)

N1 `--flag=value` form is accepted (tested, deterministic, repeated-flag detection covers mixed forms). SPEC \u00a73 shows only the `--flag value` spelling and neither permits nor forbids `=`. Recommend the parent fold a one-line SPEC statement into the erratum set so the closed grammar names it explicitly; no code change requested.
N2 Values that are empty or begin with `-` are rejected as missing values. Consistent with every flag's value domain (name grammar, closed vocabularies, model/effort/profile identifiers); SPEC is silent, documented in `Parse` doc and usage error text.
N3 \u00a74.6 "configuration read fires before a usage error" is honoured structurally: `Options.AxConfigured` is an input to `Parse`; the ax.json read itself belongs to its owning story.
