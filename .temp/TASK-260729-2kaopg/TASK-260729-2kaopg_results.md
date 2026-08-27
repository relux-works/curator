# TASK-260729-2kaopg — results

`curator global status` now reports compiled-command currentness. Work lives in
the task worktree `.temp/TASK-260729-2kaopg/worktree` (base `origin/main`
`17804ce`); nothing is staged or committed.

## The contract that was decided

The task asked for two explicit decisions. Both were made:

**1. `global status` gains `--json` and `--check`.** Both mean exactly what they
mean for `curator status`, over the same stable vocabulary and the same
classification in `cmd/curator/builds.go`. No second vocabulary was introduced.

**2. It runs a read-only global plan** — the same one `curator global install
--dry-run` runs. This is unavoidable if the existing vocabulary is to be reused:
the logical cache key is one opaque digest over the whole build input, and only
a plan derives the *current* one. A marker-only classification could report at
most build-command drift and context exposure, which is exactly the weaker
second vocabulary the task forbids.

The consequence is stated plainly in README: the command now resolves the
machine-wide closure and passes the read-only audit and registry gates. Verified
in `internal/install/global.go`, not assumed:

| Guard | Site | Effect under `DryRun` |
|---|---|---|
| commit deps | `global.go:38` | short-circuits before `Commit.resolve` |
| repository fetch | `global.go:129` | `FetchExisting: opts.Fetch && !opts.DryRun` → no fetch |
| audit gate | `global.go:193` | `audit.GateReadOnly` instead of `audit.Gate` |
| registry gate | `global.go:217` → `install.go:1086` | `persist=false` → no rollback-state migration, no trust-state write |
| live mutation | `global.go:283` | returns before `stageBuilds` and every commit path |

It runs no compiler and writes no installation target, cache entry, or trust
state. It does create an operation-private scratch workspace, which it releases.

### Two deliberate deviations from the project scope

1. **No `path` in the machine-readable document.** It carries `alias`, `skills`,
   and — only when the closure activates compiled commands — `builds`. The
   machine-wide scope has no operator-supplied root, `alias` already identifies
   it, and the manager home is never published. Two tests assert the manager
   home never appears in `global status --json`.

2. **Plain `global status` keeps its historical always-report / always-exit-zero
   contract.** The declared-skill map is still read straight from install
   markers (`scopeStatusDrift`), never from the plan, so a scope without
   compiled commands prints exactly the lines it always printed even when the
   plan refuses; a plan refusal is a `warning:` on stderr, not an `error:`.
   `--check` is the only surface that turns a verdict into a non-zero exit, and
   it fails closed twice over — once for every code that is not `up-to-date` or
   `current` (`checkFailed`), and once when the plan refused before it could
   describe every compiled command (`Status == "failed" && !BuildsComplete`),
   because such a run cannot even prove whether the scope *has* compiled state.

   This differs from `curator status`, where a refusal that leaves a command
   undescribed exits non-zero even without `--check`. That asymmetry is the
   price of AC compatibility: `curator status` has planned since before compiled
   commands existed and always exited non-zero on a failed plan, while
   `curator global status` never planned and always exited zero. Preserving the
   pre-existing declared-skill surface means preserving that exit contract.

## Implementation

`cmd/curator/main.go`:

- `statusScope` — one installed scope as seen by the read-only currentness
  surface (declaration root, own skills store, every store it reads).
  `projectStatusScope` keeps project behaviour byte-for-byte, including the
  machine-level hybrid store. `globalStatusScope` reads one store: hybrid
  declarations activate against a project, never against the global scope, and
  every node the global closure resolves — declared or transitively reached —
  installs into the global store itself.
- `statusReport` / `installedSkillDir` take that scope instead of a project
  root, so both scopes share one classification, one vocabulary, and one
  fail-closed verdict. `statusStores` was folded into the scope and removed.
- `cmdGlobalStatus` + `globalStatusPlan` implement the contract above, including
  the before/after marker fingerprint that makes compiled state which moved
  during classification report `build-state-changed` rather than an
  authoritative verdict.
- Usage now reads `status (--check, --json)` under `global` (`main.go:64`).

`README.md:194-222` replaces the predecessor's "Compiled diagnostics are a
project-scope surface" exclusion paragraph, in place, with the contract above.

Predecessor tests were updated only where the `statusReport` signature moved
(`cmd/curator/builds_test.go`, two call sites).

## Tests — `cmd/curator/global_status_test.go` (new)

| Test | Proves |
|---|---|
| `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` | an unchanged compiled global installation reports `current` with the full planned command (driver, build root, source dir, build source, target, key, artifact, cache outcome) and passes `--check`; six tampering cases each produce their own stable code and a non-zero `--check`: `build-source-drift`, `build-input-drift`+`unattributed`, `build-command-drift`, `corrupt-build-cache`, `missing-build-artifact`, `build-context-exposed`. Human output carries `state=` and `cause=`; the manager home never reaches `--json` |
| `TestGlobalStatusReportsAnUnusableToolchainPerCompiledCommand` | toolchain drift: one row per active command carrying `unusable-build-toolchain` and the stable `go-v1` boundary cause, the full selection/tested-family guidance on stderr as a `warning:`, no identity the plan never derived, and a non-zero `--check` |
| `TestGlobalStatusReportsATransitivelyResolvedCompiledCommand` | the single-store resolution finds a provider no declaration names, reports it, keeps it out of the declared-skill map, and lets its own drift fail `--check` |
| `TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands` | the no-compiled-command scope prints exactly `global: skill-a up-to-date\n`, nothing on stderr, exits zero, `--check` zero; `--json` is the two-key declared-skill document with no `builds`; ordinary `content-drift` keeps its own code and fails `--check` |
| `TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent` | a scope that declares nothing prints nothing and passes `--check` |
| `TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven` | an unresolvable declaration still publishes the declared-skill report and exits zero with a `warning:`, while `--check` refuses to call the scope current |
| `TestGlobalStatusRejectsPositionalArguments` | the machine-wide scope takes no target |

Every test redirects `HOME`/`USERPROFILE` to a temp directory, because a real
`global install` mirrors adapters and forwarding shims into the user home.

## Gates

All run from the task worktree as standalone processes; real exit codes below.
Logs in `.temp/TASK-260729-2kaopg/logs/`.

<!-- GATES -->

## Carried-forward boundary

TASK-260729-3jku56 (compiled install idempotence) is untouched by this task.
