# TASK-260729-2kaopg — compiled currentness in `curator global status`

Follow-up to TASK-260720-1nlmvv, which deliberately excluded the machine-wide
scope: its CLI surfaces were `install`, `upgrade`, `--dry-run`, `status`,
`status --json`, `status --check`, and `gc`. Global install and upgrade do plan
and commit compiled commands, so a compiled global installation had no
currentness surface at all.

## Where the work lives

- Worktree: `.temp/TASK-260729-2kaopg/worktree`, base `origin/main 17804ce`.
- Seeded by mirroring the preserved predecessor tree
  `.temp/TASK-260720-1nlmvv/worktree` (read-only; never modified, staged, or
  committed). Nothing is staged or committed in either tree.

## The contract that was decided

`curator global status` gains `--json` and `--check`. Both mean exactly what
they mean for `curator status`, over the same stable vocabulary and the same
classification in `cmd/curator/builds.go`; no second vocabulary was introduced.

It runs the same **read-only** plan `curator global install --dry-run` runs.
This was the load-bearing decision the task asked for. It is unavoidable if the
existing vocabulary is to be reused: the logical cache key is one opaque digest
over the whole build input, and only a plan derives the *current* one. A
marker-only classification could report at most build-command drift and context
exposure, which is exactly the second, weaker vocabulary the task forbids.

The consequence is stated plainly in README: the command now passes the
read-only audit and registry gates and resolves the machine-wide closure. It
runs no compiler and writes no installation target, cache entry, or trust state
(`Fetch` stays false under `DryRun`, so it also performs no repository fetch).

### Two deliberate deviations from the project scope

1. **No `path` in the machine-readable document.** It carries `alias`,
   `skills`, and — only when the closure activates compiled commands —
   `builds`. The machine-wide scope has no operator-supplied root, `alias`
   already identifies it, and the manager home is never published. Two tests
   assert the manager home never appears in `global status --json`.

2. **Plain `global status` keeps its historical always-report / always-exit-zero
   contract.** The declared-skill map is still read straight from install
   markers (`scopeStatusDrift`), never from the plan, so:
   - a scope without compiled commands prints exactly the lines it always
     printed, even when the plan refuses;
   - a plan refusal is a `warning:` on standard error, not an `error:`;
   - `--check` is the only surface that turns a verdict into a non-zero exit.

   `--check` fails closed twice over: for every code that is not `up-to-date`
   or `current` (`checkFailed`), and when the plan refused before it could
   describe every compiled command (`Status == "failed" && !BuildsComplete`),
   because such a run cannot prove the scope is current — it cannot even prove
   whether the scope has compiled state.

   A machine-wide scope with no `Skillfile.json` declares and activates nothing:
   it prints nothing, and `--check` has nothing to refuse.

This differs from `curator status`, where a refusal that leaves a command
undescribed exits non-zero even without `--check`. That asymmetry is the price
of AC compatibility: `curator status` has planned since before compiled
commands existed and always exited non-zero on a failed plan, while
`curator global status` never planned and always exited zero. Preserving the
pre-existing declared-skill surface means preserving that exit contract, and
`--check` — a flag that did not exist on this command before — carries the
fail-closed verdict instead.

## Implementation

`cmd/curator/main.go`:

- `statusScope` — one installed scope as seen by the read-only currentness
  surface (declaration root, own skills store, every store it reads).
  `projectStatusScope` keeps the project behaviour byte-for-byte, including the
  machine-level hybrid store. `globalStatusScope` reads one store: hybrid
  declarations activate against a project, never against the global scope, and
  every node the global closure resolves — declared or transitively reached —
  installs into the global store itself.
- `statusReport` / `installedSkillDir` now take that scope instead of a project
  root, so both scopes share one classification, one vocabulary, and one
  fail-closed verdict. `statusStores` was folded into the scope and removed.
- `cmdGlobalStatus` + `globalStatusPlan` implement the contract above,
  including the before/after marker fingerprint that makes compiled state which
  moved during classification report `build-state-changed` rather than an
  authoritative verdict.
- Usage now reads `status (--check, --json)` under `global`.

`README.md`: the "Compiled diagnostics are a project-scope surface" paragraph is
replaced by the contract above.

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

Predecessor tests were updated only where the `statusReport` signature moved
(`cmd/curator/builds_test.go`, two call sites).

## Gates

See the gate table in the results artifact.

## Carried-forward boundary

TASK-260729-3jku56 (compiled install idempotence) is untouched by this task.
