# TASK-260916-16ys92 results: launcher prints its resolved provider path

Story: STORY-260916-2otjbn (umbrella-provider-trust-roots), wave 1.
Repo: curator-agent-launcher. SPEC bumped `0.3.0-draft` → `0.4.0-draft`.

## Decision: path-only provider line (no origin suffix)

The brief allowed printing only the path when nothing distinguishes
umbrella-dispatched from direct invocations. Verified: the umbrella
dispatches `curator run …` by exec'ing `curator-<name>` off `PATH` with
the remaining argv verbatim and passes no argv/env marker the launcher
can read (no `os.Getenv`/`CURATOR_*`/umbrella signal anywhere in
`internal/` or `cmd/`; SPEC §2 states Curator carries no knowledge
beyond the discovery rule). Inventing origin detection (e.g. parent
process sniffing) would be a forced fit, so this revision prints the
path-only form and the SPEC records the rationale plus a deferred
closed origin set for a future revision once the umbrella passes a
marker. No new flags; all §4.3 model/effort semantics unchanged; `pi`
behavior unchanged beyond the added line.

## Line-group contract (SPEC §4.3)

At every launch, before the plan request, stderr carries, in order:

1. `curator-run: provider: path=<absolute path>` — own executable via
   `os.Executable` + `filepath.EvalSymlinks`; on any failure
   (lookup error, eval error, empty or non-absolute result) the
   diagnostic-safe fallback `path=unavailable`; never fails the launch.
2. `curator-run: defaults: model=… (…) effort=… (…)` — unchanged.

Both values fold with the existing framing rule (CR/CRLF→LF, LF→LF+2
spaces); neither line is a §6 diagnostic line.

## Per-file changes

- `SPEC.md`: §2 gains the E4 trust-root dispatch paragraph
  (curator-spec `0da4020`, PR #62; install dir then
  `provider_directories`; warn rev A / refuse rev B; path-only
  rationale); §4.3 line-group rewritten as the ordered two-line
  contract (form, resolution, fallback, folding, non-diagnostic
  invariant, deferred origin set); header + §8 intro bumped to
  `0.4.0-draft`; §8.1 gains the `0.4.0-draft` row; Specification
  changelog gains the 2026-09-17 E4 entry.
- `internal/defaults/lineup.go`: added `ProviderUnavailable`,
  `ResolveProviderPath` / `ResolveProviderPathWith` (injectable
  executable+eval seam), `ProviderLine` (folded, empty→fallback),
  `EmitGroupWithProvider` (provider line, then defaults line);
  `EmitGroup` now emits both lines via the production resolver.
  `Line()`/`Describe()` unchanged.
- `cmd/curator-run/main.go`: `specVersion` → `0.4.0-draft`;
  `launchDeps` gains injectable `providerPath func() string` (nil =
  production resolver); `run` emits the group via
  `EmitGroupWithProvider` after `Complete`, before the plan request.
- `internal/defaults/lineup_test.go`: `TestResolvedLine` updated to
  the two-line group via `EmitGroupWithProvider`; added
  `TestProviderLineFold` (form, fallback, CR/CRLF/hostile folding,
  never diagnostic), `TestResolveProviderPathFallback` (8-case table:
  executable error/empty, eval error/empty/relative, nil funcs,
  absolute passthrough), `TestResolveProviderPathProduction`
  (never empty; absolute or fallback; renders and stays
  non-diagnostic).
- `cmd/curator-run/main_test.go`: `testDeps` injects deterministic
  `/test/bin/curator-run`; all line-group prefix assertions updated to
  provider-then-defaults; mapping/alias/unresolvable refusals assert
  the provider line is absent before the group stage;
  `TestSpecVersionPinned` → `0.4.0-draft`; added
  `TestRunProviderFallbackNeverFailsLaunch` (empty→fallback and
  hostile→folded at the real entry point; launch proceeds to
  `plan_refused`, exactly one diagnostic line).
- `cmd/curator-run/defaults_test.go`: per-member, Pi-preference, and
  failure-ordering assertions updated for the two-line group; failure
  cases assert no provider line before the group.
- `cmd/curator-run/pipeline_test.go`: `entryFixture` injects
  `<fixture>/curator-run` so goldens normalize it to `<ROOT>` like all
  other host paths.
- `cmd/curator-run/testdata/pipeline-*.golden` (6 files): regenerated
  ONLY via `UPDATE_PIPELINE_GOLDENS=1`; each diff proven a pure
  insertion of `curator-run: provider: path=<ROOT>/curator-run\n`
  before the defaults line (verified per file: new minus that insert
  equals old, byte-for-byte).
- `cmd/curator-run/testdata/help.golden`, `README.md` (2 version
  strings): version pins bumped to `0.4.0-draft` for consistency
  (required by the SPEC bump; help golden first line only).
- `CHANGELOG.md`: header SPEC pin → `0.4.0-draft`; Added gains the
  "E4: …" entry for the provider-path line.

## Validation transcripts (worktree root, `set -o pipefail`)

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l .` → no output (clean) → exit 0
- `go test ./internal/defaults/ -run
  'TestResolvedLine|TestProviderLineFold|TestResolveProviderPath' -v`
  → all pass → exit 0
- `go test ./cmd/curator-run/ -run
  'TestRun|TestSpec|TestExecutable|TestGate'` → ok → exit 0
- `go test ./cmd/curator-run/ -run TestProductionPipelineGoldens`
  (before regen) → FAIL as expected (authentic golden mismatch on the
  new stderr line) → exit 1
- `UPDATE_PIPELINE_GOLDENS=1 go test ./cmd/curator-run/ -run
  'TestProductionPipelineGoldens|TestProductionAliasEquivalence'` →
  ok → exit 0 (documented regen mechanism only)
- `go test ./cmd/curator-run/ -run
  'TestProductionPipelineGoldens|TestProductionAliasEquivalence'`
  (after regen) → ok → exit 0
- `go test ./... -count=1` → all 11 packages ok → exit 0
  (cmd/curator-run 28.3s, axconfig, cli, composition, defaults,
  diagnostics, execution 15.5s, fragment, mapping, plan,
  systemprompt)

The configured landing suite was not run (runtime runs it once at
handoff, per the brief). Nothing was committed, pushed, or branched;
all work is uncommitted in the Story worktree for handoff snapshot.

## Checklist mapping

- SPEC §2 + §4.3 record the provider-path line (path-only form with
  rationale, symlink-resolved executable, fallback), §8.1 +
  specification changelog bumped — done.
- Launcher emits the provider line in the §4.3 group at every launch
  before the plan request, folded; launch never fails on resolution
  errors — done.
- Pipeline goldens show the provider line; fold + fallback unit tests;
  transcripts above; CHANGELOG E4 entry — done.
- Code per description/AC; tests written and passing; lint (vet+gofmt)
  clean; build green; outcome artifact attached (this file).
