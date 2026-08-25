# Flight Logbook

> Institutional memory. Concise, factual, high-signal.
> Newest entries first. One block per insight.

## 2026-08-25

### 0640 — `BUG-260825-11nmd5`: relaxing one directive turned an early-exit scan into a bypass for another

- CLASS: a byte scanner that stops at the *first* matching needle is only safe while every needle carries the *same* verdict. `scanSourceDirectives` (`internal/godriver/graph.go`) looked for `//go:cgo_import_dynamic` and `//go:generate` and returned on whichever it hit first. That was harmless for as long as both were rejected unconditionally.
- REGRESSION INTRODUCED BY THE FIX ABOVE IT: PR 40 (`c9fe49c`) made `//go:generate` *exempt* inside a materialized vendor tree. The scan's early exit then became a hole — a `//go:generate` in the first 64 KiB window set `matched = 2` and terminated the read, so a `//go:cgo_import_dynamic` in any later window of the same file was never seen, and the carve-out admitted the package. Reproduced end to end through `Build()` on an audited, non-replaced vendored module: diagnostic code `""`, build succeeded.
- THE GO COMPILER IS NOT A BACKSTOP HERE: `cmd/compile/internal/noder` permits `//go:cgo_import_dynamic` for general use (the comment names Solaris code in `golang.org/x/sys/unix`), and `/usr/lib/libSystem.B.dylib` satisfies its argument check. Preflight was the only thing standing there.
- FIX: the scan now resolves by **severity, not by first hit**. Only `//go:cgo_import_dynamic` — which nothing weaker can override — ends the read early; a `//go:generate` hit is recorded and the file is still read to EOF. The three severities are named constants (`directiveNone` / `directiveCgoImportDynamic` / `directiveGenerate`) so the call site reads as a verdict rather than as `matched == 1` / `matched == 2`. Cost: files that carry `//go:generate` are read whole, bounded by the frozen build source.
- FINDING, THE NARROWING MUTANT IS WHAT PROVES THE BOUND: reintroducing `return true` on the generate branch is delete-only and proves nothing about the class. The mutant that matters gates the cgo check on `matched == directiveNone`, i.e. keeps the "keep scanning" behavior but lets a recorded generate suppress a later cgo hit. It is killed by `TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows/generate_before_cgo` and by the end-to-end case. Four mutants applied and reverted, all killed, including one that removes the carve-out entirely and reddens PR 40's own allowed-side test — proof the hardening did not quietly undo the relaxation.
- SPEC BASIS: profiles/manager.md §2.3 is a *containment* predicate — an active non-standard `GoFiles` file is scanned as exact bytes and rejected if it **contains** the directive, wherever in the file it sits. An early-exit scan silently weakens containment to "contains, in the prefix before some other token". Worth reading every other early-exit byte scan in the tree with that sentence in hand.
- SCOPE: `internal/godriver/graph.go`, `internal/godriver/graph_test.go`, `internal/godriver/moduleroots_test.go`. Branch `fix/BUG-260825-11nmd5-directive-scan-shortcircuit` off `680f6a6`.


### 0415 — Marker-schema banding: one predicate, or every reader drifts

- ROOT CAUSE: three readers each carried their own inequality on the install-marker schema instead of asking the marker package. `classifySkillBuilds` (`cmd/curator/builds.go:365`) admitted only the written schema; `markerRefusal` (`builds.go:545`) admitted schemas 1-2; `marks.absorb` (`internal/scopes/gc.go:213`) admitted 2-3. Marker v4 failed all three.
- FINDING: `marker.Current` was correct throughout, because it asked the shared `buildBearingSchema` predicate. The three that drifted were the three that hand-rolled the check. The predicate existed; it was just unexported.
- FINDING: the GC one was the dangerous half and went unreported. A marker v4 contributed no live build reference, so a maintenance pass could delete protected cache entries a live schema-8 installation was still running from. Found only by grepping the whole class, not from the bug report.
- FINDING: a stale banding check produces a *self-contradictory* remedy, which is the tell. "install marker schema 4 cannot describe a compiled command; reinstall to record marker schema 2" — the manager would never write schema 2 for that band.
- FIX: exported `marker.BuildBearingSchema` / `marker.SupportedSchema`, added `marker.NewestSchemaVersion`; all three readers ask them. `internal/marker/marker.go`, `cmd/curator/builds.go`, `internal/scopes/gc.go`.
- STATUS: PR #41, resolved pending merge.

### 0416 — Two tests asserted the defect instead of catching it

- REGRESSION: `TestMarkerRefusalSeparatesUnsupportedFromInvalid` pinned `{"schema_version":3}` as `unsupported-marker` — i.e. it asserted that a schema this manager reads is unreadable. The test passed for the whole life of the bug and would have blocked the fix.
- FINDING: the status matrix case named "marker schema cannot be read by this manager" pinned `marker.SchemaVersion + 1`. That is *written* + 1, not *readable* + 1. The moment a newer schema became readable it stopped testing the unsupported band and started testing a readable one. Pin `NewestSchemaVersion + 1`.
- NOTE: general shape — a negative test anchored to the value a system *writes* silently stops being a negative test when the *read* band widens. Anchor negative schema/version tests to the edge of the accepted band.
- SCOPE: `cmd/curator/builds_test.go`, `cmd/curator/status_test.go`.

### 0417 — Conformance suite cannot catch a one-schema reader

- FINDING: `curator-spec` at pin `0ed5c691` (rc.9) proves the marker write side and the currentness side separately and never crosses them with the schema as the variable.
- FINDING: `expected/install-marker-v4.json` pins what a schema-8 install must *record*, and says nothing about what a reader must then *accept*. `vectors/manager-lifecycle.json` → `status_cases` has two cases: `compiled-installation-current` lists `marker-schema` among its `validated` steps but never parameterises which schema; `compiled-currentness-failure-matrix` enumerates 14 `independent_conditions` and none is a marker-schema condition. The `go-build-skill` fixture carries no `schema_version` field at all. `gc_cases` has no marker-schema-parameterised liveness vector.
- NOTE: a manager can be conformant on every published vector while reporting a successful install as needs-install. Gap written up for the suite owner in the BUG-260825-1l1st9 board artifact.
- STATUS: pending — needs suite-owner action, not a curator change.

## 2026-08-23


### 1307 — Ratified: operator credential selections are never lockable

`build_ssh` stays outside LockableKeys by decision of the spec owner (2026-08-23): credential material is operator-owned, and a system configuration must not select or constrain it. The manager-profile system-configuration clause now states this explicitly; the peer implementation records the same rule at its lockable-keys definition.
