# TASK-260924-1nh93t review verdict — CR rev2: CHANGES REQUESTED (to-dev)

Reviewed candidate tree d6590e2b vs base 850ac393 (11 paths). Reviewer: claude-opus-5-5.

## Blocking finding — audit table misses a SPEC §4 clause (no public API exists)
SPEC 0.5.0-draft §4.3 (launcher worktree SPEC.md, permission-mode block): every untracked launch prints, **before admission**,
`curator-run: permissions=<mode> source=<...> mapped=<flag or none>` and "The `mapped` value is supplied by agents-management; this SPEC names no provider flag."
- The module exposes no public API returning the mapped native flag (or "none") for (system, release, mode). `ReleaseCapability` (pkg/agentic/system.go:461) carries only Release/Grammar/YoloSupported; `Plan`/`LaunchProvenance` (pkg/agentic/plan.go:15-57) carry no mapping; the spelling lives only in plugin-private consts (e.g. pkg/agentic/systems/claude/policy.go). The launcher cannot spell it itself (no provider spelling outside plugins), and it must print the line before admission, so it cannot scrape Plan.Argv either.
- TASK-260924-1nh93t_results.md audit row "§4.3–§4.5" claims satisfied but does not cover `mapped=`. This is exactly the missing-row class the review note warns about (would stop F-L1b again).
Required: add an audit row for §4.3 `mapped=` provenance and a public, release-versioned API (e.g. `Registry.PermissionMapping(system, toolRelease, mode) -> {Flag string|none, Grammar}` via an optional plugin capability reusing the same `verifiedReleases` rows), fail-closed typed errors for unknown system / unverified release, `ErrPermissionModeUnsupported`→"none"/typed per pi-native, tests per environment, narrowing mutants, README + fold into the single CHANGELOG bullet.

## Verified OK (keep)
- Classifier `agentic.Registry.ClassifyNonInteractiveArgs` (pkg/agentic/noninteractive.go:54): per-release rows, same grammar, typed errors (ErrUnknownSystem, ErrNativeArgsClassifierUnsupported, ErrPermissionModeUnverifiedRelease, ErrNativeArgsClassificationIndeterminate), `--` stop via internal/nativeargs, codex exec/e placements; no second grammar.
- My mutant: codex classifier substitutes pinned release "0.153.2" for caller input → `go test ./pkg/agentic -run 'NonInteractive|Classif' -count=1` exit 1 (noninteractive_test.go:136) — KILLED.
- zsh, pipefail: `go test ./pkg/agentic/...` exit 0; `go vet ./pkg/agentic/...` exit 0. CHANGELOG: one new Unreleased bullet, released entries untouched. Spelling guards (codex/muse argvguard) green in the run above.
- Minor (non-blocking): claude/pinative treat `--print=false` as print; acceptable fail-toward-headless (native default) but worth a comment/test.
