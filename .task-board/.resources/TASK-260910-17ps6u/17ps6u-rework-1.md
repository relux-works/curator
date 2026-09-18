# Rework 1 — TASK-260910-17ps6u (revision 1 → 2): hosted gate failure

The revision-1 hosted gate (run 35306297978) failed on EVERY test/race lane with one test:
cmd/curator TestMarkerRefusalSeparatesUnsupportedFromInvalid/schema_from_a_newer_manager — builds_test.go:545:
`state = "invalid-marker", want "unsupported-marker"` for payload `{"schema_version":5,"name":"build-skill"}`.
Cause: this leaf makes marker schema 5 readable (marker.SupportedSchema now includes SchemaV5), so a schema-5 document that is not a valid marker is correctly "invalid-marker"; the test still used the literal 5 as "a schema from a newer manager". Your candidate is otherwise intact (the runtime shows it as CR revision 1, changes_requested); keep every production change as is.

Do exactly:
1. In cmd/curator/builds_test.go: make the "schema from a newer manager" row use a version strictly above every readable schema, derived from the marker package rather than a literal (e.g. `marker.SchemaV5+1` or an exported "newest readable" constant if one exists — do not add new exported API unless needed), and add the row "readable draft schema 5 that is still not a valid marker" → stateInvalidMarker. Update the comment above the readable-schema rows (3 and 4 → 3, 4 and 5). Keep the other rows unchanged.
2. Grep the repository tests for other places that treat schema 5 as unsupported/newer (`"schema_version":5`, `SchemaVersion: 5`, "newer manager") and fix them the same way; report the grep result in results.md even if empty.
3. Run narrowly with real exit codes: `go test -p 1 ./cmd/curator -run 'TestMarkerRefusalSeparatesUnsupportedFromInvalid|TestStatus.*Marker|Marker' -count=1 -timeout=180s` (background + tail), `go test -p 1 ./internal/marker -count=1`, `go vet ./cmd/curator`, gofmt. Then run the whole `go test ./cmd/curator -count=1` ONCE in the background (output to a file, tail it; it may take 10+ minutes here) and cite its exit code — the hosted gate failed on all lanes, so make sure no other cmd/curator case depends on the old band before handing off.
4. Append a "Revision 2" section to results.md (what changed, commands, exit codes), keep the checklist, then `task-board handoff TASK-260910-17ps6u --role developer`. No other scope changes; every tool call under 2 minutes.
