# Explicitly excluded artifacts

Everything listed here exists on disk and must **not** be read as evidence.

## 1. The cancelled baseline race rerun — `gates-baseline/`

`bin/run-gates-baseline-race.sh` was launched at `BASELINE-RACE-START`
2026-07-29T20:01:45 to give the never-edited twin a race-atomicity number. It
was cancelled almost immediately by the orchestrator directive that opened this
re-entry cycle.

| file | state | verdict |
| --- | --- | --- |
| `gates-baseline/BASELINE-RACE-START` | 20:01:45 | run started |
| `gates-baseline/gate-race-atomicity-1.barrier` | `BARRIER_OK` | barrier taken |
| `gates-baseline/gate-race-atomicity-1.log` | **0 bytes** | killed mid-test |
| `gates-baseline/gate-race-atomicity-1.exit` | **absent** | **no result** |
| `gates-baseline/gate-race-atomicity-{2,3}.*` | absent | never started |
| `gates-baseline/gate-race-install-1.*` | absent | never started |
| `gates-baseline/BASELINE-RACE-DONE` | absent | driver never finished |

Under the evidence protocol a missing `.exit` means "killed or still running",
never "passed". **Excluded. Not restarted by this run.**

`gates-baseline/killed-RUN-18-58/` holds the same shape of debris from the
earlier 18:52 baseline driver — a `.barrier` and an empty `.log` for
`gate-race-atomicity-1`, no `.exit`. **Also excluded.**

## 2. What survives in `gates-baseline/`

Two gates completed before that driver was killed, and both are valid:

| gate | exit | wall |
| --- | ---: | ---: |
| `gate-transaction` | 0 | 15 s |
| `gate-atomicity-structure` | 0 | **306 s** |

Both have a `.barrier` reading `BARRIER_OK`, a non-empty `.log`, a `.seconds`
and an `.exit`. `gates-baseline/DRIVER-DONE` is absent and is **not** claimed.

Consequence for the margin argument: the same-session baseline has a **non-race**
atomicity number and **no race number**. Stated as such in the results document
rather than papered over with rfrdfo's cross-session figures.

## 3. `gates/gate-lint.exit` = 127

Missing-binary code, not a test result. Superseded by `gate-lint-abs`
(`evidence/lint-gate.md`).

## 4. `gates-partial-RUN-260729-313095/`

The archived partial run from an earlier cycle. Its `gate-atomicity-structure`
carries the driver's `99` barrier-refusal code — a non-result. Its other gates
are superseded by the complete `gates/` run and are kept only for audit.

## 5. Empty driver logs at the task root

`.baseline-driver-nohup.log`, `.baseline-race-driver.log`, `.chain-driver.log`
are all 0 bytes — nohup stubs for detached drivers, no content, no evidentiary
value.
