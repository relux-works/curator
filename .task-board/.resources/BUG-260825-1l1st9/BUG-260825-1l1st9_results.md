# BUG-260825-1l1st9 — marker v4 compiled-command status banding

PR: https://github.com/relux-works/curator/pull/41
Branch: `fix/BUG-260825-1l1st9-marker-v4-compiled-banding`
Commit: `564371c` (signed, G)

## Workspace note

The STORY-260822-2lvw0e worktree was at `903af23`, which predates the entire
schema-8 / marker-v4 subsystem — `PolicySchemaVersion` does not exist there and
the bug is not reproducible. `origin/main` was at `272b203`. The fix was
therefore built in a task-scoped worktree
`.temp/BUG-260825-1l1st9/worktree` forked from `origin/main`, and the story
branch was left untouched.

## Root cause

Three readers each carried their own inequality on the install-marker schema
instead of asking the marker package which schemas carry a build record.

| Reader | Admitted | Consequence |
| --- | --- | --- |
| `classifySkillBuilds` (`cmd/curator/builds.go:365`) | schema 2 only | the reported escape; marker v3 refused too |
| `markerRefusal` (`cmd/curator/builds.go:545`) | schemas 1, 2 | schemas 3 and 4 reported "from a newer manager" |
| `marks.absorb` (`internal/scopes/gc.go:213`) | schemas 2, 3 | marker v4 contributed no live build reference — GC could delete cache entries a live schema-8 install was running from |

`marker.Current` was already correct: it asks `buildBearingSchema`.

## Fix

`buildBearingSchema` / `supportedMarkerSchema` exported as
`marker.BuildBearingSchema` / `marker.SupportedSchema`; all three readers ask
them. Added `marker.NewestSchemaVersion`, pinned to be the real maximum of the
readable band.

The remedy string no longer names a schema to record — that was the
self-contradiction ("schema 4 … record marker schema 2").

## Two existing tests asserted the defect — corrected, not deleted

- `TestMarkerRefusalSeparatesUnsupportedFromInvalid` claimed schema 3 was
  unreadable.
- The status matrix case `marker schema cannot be read by this manager` pinned
  `marker.SchemaVersion + 1`, so it stopped testing the unsupported band the
  moment a newer schema became readable.

## Tests added

- `internal/marker/schema_band_test.go` — both predicates pinned exhaustively
  over `[-1, NewestSchemaVersion+3]`, plus `NewestSchemaVersion` is the maximum.
- `TestClassifySkillBuildsAcceptsEveryBuildBearingMarkerSchema` — schemas 2/3/4
  current; schemas 0/1/5 refused; remedy must never name an older schema.
- `TestStatusReportFindsASchema8InstallationCurrent` — drives `statusReport`
  (the production call site behind `curator status` / `curator global status`)
  over markers produced by the real `marker.Write` at bands 6/7/8, asserting
  the writer really produced schemas 2/3/4.
- `TestCollectMarksBuildKeysFromEveryBuildBearingMarkerSchema` — GC marks live
  keys from all three schemas; a schema-1 install adds none.
- `TestAbsorbKeysBuildLivenessOnTheSchemaBandNotOnAPopulatedBuildsMap` — the
  GC delete-direction bound, unreachable through a marker on disk.
- `TestMarkerRefusalSeparatesUnsupportedFromInvalid` extended: unsupported
  detail must name the real newest readable schema.
- New status matrix case: a readable schema whose document is invalid must be
  `invalid-marker`, never `unsupported-marker`.

## Mutation evidence

7/7 mutants caught — see `BUG-260825-1l1st9_mutation-evidence.txt`. Both
directions on every gate: narrowing a band turns the positive cases red,
deleting a band turns the negative cases red.

## Gates run locally (darwin/arm64)

`CURATOR_CONFORMANCE_ROOT` = pinned rc.9 suite (`SPEC_PIN` `0ed5c691`,
manifest `sha256:803918bf…b44403`), materialized so no gate ran silently
narrowed.

| Gate | Exit |
| --- | ---: |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l .` | clean |
| `golangci-lint run` | 0 (0 issues) |
| `go test ./internal/...` (42 pkgs, 14m25s) | 0 |
| `go test ./cmd/curator/` (324s) | 0 |
| `ledger-consistency.sh` (80 rows) | 0 |
| `no-broad-suppression.sh` | 0 |
| `gate-selftest.sh` (81 passed) | 0 |

Not run locally: the Windows and Linux lanes, the race lanes, the interop and
naming gates. Those run on the PR.

## Live verification (AC)

Before the fix, on the marker that produced the escape:

```
global: project-management needs-install
global: project-management.task-board ... cache=cache-hit state=needs-install
  detail="install marker schema 4 cannot describe a compiled command; reinstall to record marker schema 2"
```

After, same marker, **no reinstall**:

```
global: project-management up-to-date
global: project-management.task-board     ... cache=cache-hit state=current
global: project-management.task-board-tui ... cache=cache-hit state=current
global: project-management.tb-sessiond    ... cache=cache-hit state=current
```

Marker still `schema_version 4`, `commit 3958813`.

## Conformance-suite gap

Recorded for the suite owner in
`BUG-260825-1l1st9_conformance-gap.md`. Summary: the suite proves the write
side (`expected/install-marker-v4.json`) and the currentness side
(`status_cases`) separately and never crosses them with the schema as the
variable, so no vector would fail a manager that bands on one schema.
