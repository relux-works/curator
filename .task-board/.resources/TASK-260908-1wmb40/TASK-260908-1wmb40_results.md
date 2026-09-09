# TASK-260908-1wmb40 — a1-composition-delivery

Ready for review. Candidate is UNCOMMITTED on the managed Story branch.
Base: `18aeaed9af7dc5ffbe6cc79a4731a852fbb716da`; `HEAD..main` count at inspection: 0.
Host: Darwin arm64, Go 1.25.5, non-root (permission case executed, not skipped).

## Scope and implementation

- `internal/composition.Compose` consumes the actual released `agentic.Plan` and
  `LaunchRequest` types. Its narrow `ChildEnvironment` interface is satisfied by
  `agentic.System`; only `ChildEnv(nil, sameRequest)` is called, once. No BuildPlan
  or BuildLaunch occurs here. Original system/request pairing is a caller precondition.
- `Value` preserves Binary/WorkDir, exact ordered argv and raw stdin, with separate
  JSON transport fields. Full Plan.Env is the direct base; own/fragment/channel
  literals alone enter tracked data. Warnings carry names only and are returned
  independently of mode. Empty direct Env is non-nil to avoid os/exec inheritance.
- MCP descriptors come from validated fragments. PromptApplication is an already
  selected and encoded §5 input, not a new prompt selector or provider flag builder.
- `Value.CheckLaunchBoundary` is the exported fresh filesystem check, deliberately
  separate from Compose. Missing, dangling, permissions and nonregular are distinct;
  no flag removal, repair, fallback, process execution or ax invocation occurs.
- README records the production call-site obligations. Reserved path_prepend parser
  and hash handling are untouched; no managed command-root transform is invented.
- Only allowed direct module edge: skill-agents-management v0.5.10, source commit
  12f443d10bc217ca7a48e2edab19c739f441df9c, checked by `go mod download -json`.
  Its Go 1.25.5 minimum is reflected in go.mod. go-toml/v2 is its transitive
  dependency, brought in by the real released Codex ownership-surface test.

## AC coverage and production entry points

**10 of 12 expanded AC rows driven** through exported production APIs; the
remaining 2 rows are explicit later-story bounds. Within the requested composition
API scope this is 10 of 10. Tests are repository test files in the uncommitted
candidate, to be committed only by the parent after review per Story policy.

| Row | Behavior | Production call site | Named test / bound |
|---|---|---|---|
| 1 | Plan/prompt/MCP/native exact order, native uninspected | Compose | TestComposeOrderAndChannels |
| 2 | Binary and WorkDir retained, input independence | Compose | TestComposeOrderAndChannels |
| 3 | Full filtered environment, no removed-name re-admission | Compose | TestComposeEnvironmentBoundary; TestComposeReleasedChildEnvironment |
| 4 | Nil-parent same-request own names; no inherited secret serialization | Compose → ChildEnvironment.ChildEnv | TestComposeEnvironmentBoundary; TestComposeReleasedChildEnvironment |
| 5 | Disjoint literal/lookup names; merely inherited lookup retained; names-only warnings | Compose | TestComposeEnvironmentBoundary; TestComposeUnownedOverrideNoWarning |
| 6 | Unattached/attached-empty/UTF8/binary D4 and raw bytes | Compose | TestComposeStdin |
| 7 | MCP path + companions, name, variable, absence; no plan rebuilding | Compose | TestComposeOrderAndChannels; TestLaunchBoundaryUnengaged |
| 8 | Fresh missing/unreadable/regular codex layer refusal, flags retained | Value.CheckLaunchBoundary | TestLaunchBoundaryFilesystem; TestLaunchBoundaryLateReplacement |
| 9 | Ownership failure propagates, no candidate returned | Compose | TestComposeOwnershipFailure |
| 10 | Empty direct environment does not imply inherited underlay | Compose | TestComposeEmptyEnvironment |
| 11 | Full executable pipeline, stderr in both modes, final process/ax boundary invocation | Not wired (explicit task exclusion) | BOUND: execution Story must drive real main through both modes, invoke fresh probe immediately before launch with binary/§5 checks, and send tracked schema/extensions |
| 12 | Native Pi BuildLaunch admission against real operator tag | Not admitted here (explicit task exclusion) | BOUND: Pi-shaped plan fixtures prove only composition; later plan/execution Story must verify native Pi against the real new tag |

The filesystem probe has the usual path replacement window after inspection.
Post-open stat error/race branches are defensive and not deterministically driven;
no elimination of TOCTOU is claimed. The real unreadable test ran on this host;
it skips for root, and the mutant harness then reports a survivor/unverified case.
No new source-text inspection gate exists, so token-preserving source-gate attack
is not applicable. The mutant harness runs the behavioral suite, not a static scan.

## Validation performed directly

| Command | Exit | Evidence |
|---|---:|---|
| go test ./internal/composition -count=1 -v | 0 | focused-01.log (initial API suite) |
| go test ./internal/composition -count=1 -v | 0 | focused-02.log (adds real released Codex ChildEnv) |
| python3 .scripts/composition-mutants.py .temp/composition/mutants | 0 | mutants-01.log and mutants/*.log; each mutant test command exits 1, expected red |
| make check | 0 | make-check-01.log; run once after changes, build/fmt-check/vet/test/race all green |
| git diff --check | 0 | diff-check-01.log |

Each mutant executes `go test ./internal/composition -count=1 -v` directly,
with stdout/stderr sent to its log and its real exit recorded. The harness restores
original candidate bytes in finally and verifies restoration. All 9 are killed;
no survivor is presented as success. No gate is piped through tee.

| Mutant | What weakening admits | Named failing test | Exit | Survivor bound |
|---|---|---|---:|---|
| M1 | admit exactly inherited SECRET as a tracked literal | `TestComposeEnvironmentBoundary` | 1 | none |
| M2 | admit OWN as both literal and lookup | `TestComposeEnvironmentBoundary` | 1 | none |
| M3 | re-admit exactly plugin-removed REMOVED | `TestComposeEnvironmentBoundary` | 1 | none |
| M4 | admit one ChildEnv failure | `TestComposeOwnershipFailure` | 1 | none |
| M5 | admit absent layer while retaining all unreadable gates | `TestLaunchBoundaryFilesystem/missing` | 1 | none |
| M6 | admit dangling symlink after successful lstat | `TestLaunchBoundaryFilesystem/dangling` | 1 | none |
| M7 | admit directories but retain other nonregular rejection | `TestLaunchBoundaryFilesystem/directory` | 1 | none |
| M8 | admit permission-denied open only | `TestLaunchBoundaryFilesystem/unreadable` | 1 | none |
| M9 | collapse attached empty only to null | `TestComposeStdin/empty` | 1 | none |

## Inputs and operational notes

Read current SPEC §§4.5/4.6, Decision 0013 at d019f0e7179520b5c8dcde321c4fe51e04552f58,
and accepted-environment-evidence.md from TASK-260909-zg440c (the accepted
TASK-260908-1c0fwn analysis). No fresh provider runtime claims were inferred.

Initial required development status command exited 1 because the task lacked an
estimate; set estimate to Fibonacci 5 and repeated status successfully. Resource
projection guesses were unsupported, so resource inputs were located read-only.
An initial guessed module name github.com/relux-works/agents-management failed
with repository-not-found (exit 1); SPEC names skill-agents-management and that
exact allowed release downloaded successfully (exit 0). No workaround was used.

No main/SPEC/LOGBOOK/control-root/private-record edits, installs, releases, tags,
real model runs, ax calls, commits or home/config writes were performed. Findings
are attached on this board instead of LOGBOOK, as the task explicitly forbids
LOGBOOK edits. The parent owns independent review and signed delivery.
