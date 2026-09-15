# TASK-260909-2vy977 producer evidence rev2 (v0.5.11 native-Pi completion)

State: ready for review. Candidate is UNCOMMITTED on the managed Story
branch `task-board/story/STORY-260908-1wxjbs`; HEAD remains `ae8676c`,
no commit, rebase, reset, or stash performed. This rev2 supersedes the
rev1 outcome only where rev1 assumed v0.5.10: rev1 file-resolution,
diagnostics, and entry-framing evidence stands and is preserved.

## Review-verdict F1/F2/F3 resolution

- F1 (missing v0.5.11 + refusal-only Pi coverage): resolved. The
  operator-published tag was verified (below), `go.mod` pins the real
  `v0.5.11`, and Pi now resolves successfully end to end:
  `pi` -> `gpt-5.6-sol` / `max` via `pi-openai`, all lineup origin.
  `TestRunPiUnresolvable` is replaced by `TestRunPiResolvesNativeLineup`;
  `TestCompleteLineupTops`, `TestRunLineupEnvsPrintGroupBeforeRefusal`,
  and `TestRunMapping` all drive 3 of 3 environments successfully.
- F2 (registration claim): corrected, not defended. The rev1 claim that
  a native-Pi declaration needs zero launcher changes was wrong: v0.5.11
  declares THREE runtimes for `pi-native`, the exactly-one declaration
  rule would refuse Pi forever, the `pinative` system plugin was never
  registered, and the `google` vendor's own registration contract
  requires the `gemini-cli` and `antigravity` systems. All three are now
  wired (`internal/defaults/lineup.go` `NewRegistry`) and positively
  tested. The legacy `pi` wrapper system is still deliberately
  unregistered; `ResolveRuntime("pi")` and `ResolveRuntime("pi-native")`
  (a system id, not a runtime id) both refuse.
- F3 (validation accounting): corrected. This outcome reports exactly
  what ran: narrow package tests during iteration, the full 34-mutant
  harness once on the finished tree, and `make check` exactly once at
  the end (exit via `PIPESTATUS[0]`, pipeline stated, not hidden).

## Upstream dependency evidence (personally verified 2026-09-09)

- `git ls-remote https://github.com/relux-works/skill-agents-management`
  `v0.5.*`: `0ea486e46765ecf12fff7c2ac526e12da02e95ed refs/tags/v0.5.11`
  and `a2a6e9f377f62a5872d99ecdfff0d1690e385f2a refs/tags/v0.5.11^{}`.
- Fetched tag object verified: `git tag -v v0.5.11` ->
  `Good "git" signature for oparin@me.com with ECDSA key
  SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`,
  tagger Ivan Oparin, `Native Pi interactive support`, peeled commit
  `a2a6e9f...` (PR23 head).
- API inspected at the tag: `pinative` system (`ID pi-native`,
  interactive + dry-run only, `--model <vendor>/<id>` plus
  `--thinking <effort>`), frozen runtimes `pi-anthropic`, `pi-openai`,
  `pi-google` (`docs/consuming-the-module.md`: "the launcher resolves
  one of the frozen pi-anthropic, pi-openai or pi-google runtimes"),
  `RuntimeDeclarations()` in sorted id order, `ResolveRuntime`
  requiring both system and vendor plugins, vendor `Register`
  requiring every agentic system its rows mention
  (`ErrUnknownAgenticSystem`).
- Scores are vendor-local (openai top 120, anthropic 85, google 60), so
  no cross-vendor comparison is meaningful; the union top is
  deterministic and module-derived, printed with its origin on stderr,
  and operator-overridable per member. This caveat is in the code
  comment, not hidden.
- Pin: `go get ...@v0.5.11`, `go.mod` diff is the one version line,
  `go.sum` gains the two v0.5.11 hashes, `go mod verify` -> `all
  modules verified`, no `replace`, no `go.work`.

## Design: union candidates, contributor binding

`Complete` gathers driven rows across EVERY runtime declared for the
mapped system (registry sorted-declaration order) and ranks the union
with `vendorplugin.Lineup`; the bound `Runtime` is the contributor of
the selected row. Single-runtime systems behave exactly as before. New
narrow gates: a configured model or the union top driven by several
runtimes binds none (`defaults_unresolvable`); a declared runtime that
fails to materialize poisons its whole system. Observed pins
(re-verify on pin change): `claude-fable-5-1`/`high` via `claude`,
`gpt-6-astra`/`max` via `codex`, `gpt-5.6-sol`/`max` via `pi-openai`
(per-runtime pi tops: anthropic `claude-fable-5`/80,
openai `gpt-5.6-sol`/120, google `gemini-3.1-pro-preview`/50 effortless).

## Behavioral AC coverage (12 of 12 implementation rows driven)

`run` is `cmd/curator-run.run`; `Complete` is `defaults.Files.Complete`.

| AC row | Production call site | Named test |
|---|---|---|
| Real tagged lineup fallback v0.5.11 | `Complete` via `run` | TestCompleteLineupTops (3 envs), TestNewRegistryResolvesLaunchableSystems (5 runtimes) |
| Levels 1-2 file/flag resolution preserved | `Complete` via `run` | prior suite untouched (20 file mutants re-killed); TestCompletePerMemberFill |
| Per-member fill, only unset members | `Complete` via `run` | TestCompletePerMemberFill (+pi-model-binds-contributor-runtime) |
| Row recommendation, not display pick | `Complete` | TestCompleteLineupTops, PerMemberFill/model-takes-own-row-effort |
| No-effort semantics (None/unknown unset) | `Complete` | TestCompleteNoEffortSemantics (+pi-effortless-row binds pi-google) |
| No completion past refusal/failure | `Complete` via `run` | TestCompleteNoFirePastFailure, TestRunDiagnosticsContract/usage-locked-flag |
| Origin line-group at every launch | `EmitGroup` in `run` | TestResolvedLine, TestRunPiResolvesNativeLineup, TestRunMapping (3 envs) |
| Group ordered before plan-request point | `run` | TestRunLineupEnvsPrintGroupBeforeRefusal (3 envs), TestRunMapping |
| Locks/failures refuse with no launch | `run` | TestRunDiagnosticsContract (13 rows incl. unresolvable via bare registry) |
| Typed failure, no invented model | `Complete` via `run` | TestCompleteUnresolvable, TestCompleteAmbiguousRuntime (2 subtests), TestCompleteNilRegistry, TestCompleteRuntimeResolutionFailure (2 subtests), TestCompleteWrongVendorModels, TestCompleteMappingRefusal |
| All three envs through the real module | `NewRegistry`+`Complete` in `run` | per-env tests above; 3 of 3 environment rows successful |
| Pin real verified tag, no override | `go.mod`/`go.sum` | one-line version diff, `go mod verify` exit 0, no replace, no go.work |

Stated bounds (declared, not driven): B1 `ErrEffortMissing`->`plan_refused`
and `BuildLaunch` admission belong to TASK-260908-2so46q. B2
provider-limits verdict gate belongs to 2so46q. B3 composition/exec
wiring is a later story (`not_implemented` remains). B4
`NewRegistry` error branches (unreachable with the real module).
B5 SPEC history lines naming v0.5.10 as past evidence are preserved as
evidence, not claims. No source-text-inspection gate was added (typed
`SystemID` + exact id equality only), so the token-preserving-mutant
clause is not applicable.

## Commands personally executed (real exits)

| Command | Real exit | Evidence |
|---|---|---|
| `go test ./internal/defaults/ ./cmd/curator-run/ -count=1` (finished tree) | 0 | both ok |
| `python3 .scripts/defaults-mutants.py .temp/TASK-260909-2vy977/mutants` | 0 | 34 of 34 killed, per-mutant logs + summary.tsv |
| `make check 2>&1 \| tail -25; echo MAKE_EXIT:${PIPESTATUS[0]}` | 0 | build, fmt-check, vet, all tests, all race tests green; exit is make's own via PIPESTATUS |
| `go mod verify` | 0 | all modules verified |
| `git diff --exit-code` guard | n/a | HEAD ae8676c intact; all changes uncommitted (status below) |

Mutant deltas vs rev1: `empty-lineup-admit` removed (branch provably
unreachable: non-empty candidates imply non-empty Lineup; the empty
gate lives only in `len(cands)==0`, still mutated by
`unresolvable-admit-empty`); `ambiguity-first-wins` replaced by
`ambiguity-model-admit` (>1 to >2, killed by
.../configured-model) and `top-contributor-admit` (!=1 to <1, killed by
.../lineup-top); `findrow-fuzzy`, `runtime-drop`,
`resolution-failure-admit` (now skip-past-failure, killed by the new
partial-failure subtest), `lineup-bottom`, `unresolvable-admit-empty`
re-anchored to the union code. Every new gate has its narrowing
mutant; all narrowing (never delete-only).

## Files changed (uncommitted, HEAD ae8676c intact)

- `go.mod`/`go.sum`: v0.5.10 -> real v0.5.11 only
- `internal/defaults/lineup.go`: pinative+gemini+agy systems, google
  vendor, union candidates + contributor binding, multi-contributor and
  poisoned-binding refusals
- `internal/defaults/lineup_test.go`: 3-env tops (+pi pin), pi
  contributor/effortless subtests, ambiguity x2, partial-failure,
  no-declaration unresolvable
- `cmd/curator-run/main.go`: package doc now states §4.3 delivered
- `cmd/curator-run/main_test.go`: pi success coverage x3 env rows,
  bare-registry unresolvable contract row, unicode-boundary expectation
- `cmd/curator-run/gate_framing_test.go`, `internal/diagnostics/*`:
  preserved rev1 bytes, still valid
- `.scripts/defaults-mutants.py`: 7 anchors reworked (above)
- `README.md`: v0.5.11 + union wording; `SPEC.md` §4.2: tag published,
  §4.3 implemented, §4.4 still later

Platform: macOS arm64, go1.25.5. No hosted CI, installs, daemon
restarts, real ax, runtime-home edits, tags, releases, or main writes.
