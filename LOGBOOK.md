# Flight Logbook

> Institutional memory. Concise, factual, high-signal.
> Newest entries first. One block per insight.

## 2026-08-25

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
