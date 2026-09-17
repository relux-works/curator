# TASK-260916-55g9dg — hosted gate failure on Change Request revision 1 (run 35176838317)

Extracted by the orchestrator from the CI evidence of
https://github.com/relux-works/curator/actions/runs/35176838317. macOS and
Windows lanes are green; only the Linux lanes (Test and Race, ubuntu-latest)
fail, and only one test:

```
internal/envprofile TestStatusReportsPolicyAndDropped
admission_test.go:300: drop warnings made the row non-current:
  [environment_passthrough_detached: passthrough entry .credentials.json is detached]
```

So the status row you assert as current carries an unrelated warning on
Linux: the test fixture's passthrough entry `.credentials.json` is reported
detached there (a symlink/passthrough materialization difference between the
Linux runner and macOS — check how the fixture creates the passthrough entry
and what `environment_passthrough_detached` requires on Linux), and your
assertion "drop warnings never make the row non-current" is evaluated against
ALL warnings of the row. Fix the fixture so the passthrough entry is
materialized/attached on Linux (or assert specifically that no
`context_system_module_dropped`-class warning makes the row non-current
while other warning classes are filtered out of that particular assertion,
if the spec §12 text is about drop warnings only). Note also that the
manager-config vector comparison did NOT fail for E2 on the pinned root —
keep it that way (see `spec-pin-lag-hold.md` for the E4/S4 situation; do not
add knob keys to the normalized defaults output unless the spec vector at
the pinned root carries them).

Re-run `go test ./internal/envprofile/...` (ideally on Linux via the gate)
and hand off again; the runtime re-runs the hosted gate.
