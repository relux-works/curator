# Gate fix note — TASK-261001-2yvag1 (orchestrator; read before re-handoff)

Hosted gate runs 36813214802 (rev1) and 36821035096 (rev2) are red on all five lanes. The same 6 tests fail in each (from the go-test.json artifacts):
- cmd/curator:
  - TestEnvStatusMissingAndUnreadableKeepRecord (hook_posture_test.go:374: `env status --check without approvals = 1`)
  - TestEnvStatusRegistryBoundaryPostureAndCheck
  - TestEnvStatusReportsShellHookTrustPosture
  - TestEnvStatusUnreadableApprovalStateSurfaced
- internal/envprofile:
  - TestManagerOwnedAbsenceReadsAreGuarded (state_read_guard_test.go:215: `internal/envprofile/muse.go:*ResolveRequest.inspectMuseAuth tests not-exist after function-value os.Lstat/os.Readlink without a seam route or reviewed allowlist reason`)
  - TestEmptyAllowlistWarningLeavesCurrentStatusCurrent (surfacing_test.go:230: the control fixture is no longer current because the new `muse` environment shows Provisioned:false)

Cause: adding muse as a known environment makes the existing status fixtures non-current, and the new muse auth inspection reads state outside the guarded seam.

Fix:
1. Route inspectMuseAuth's Lstat/Readlink through the existing stateread seam. Do not add an allowlist entry unless a reviewed reason is truly needed.
2. Make an unprovisioned muse not break "current" status for profiles that never enabled muse, the same way other optional environments behave (spec d373078a decides). Do not just edit fixtures to hide a real behaviour change: if the status semantics legitimately change, state it and update the fixtures with the reason.

Reproduce locally:
`go test ./cmd/curator -run 'TestEnvStatus' -count=1` and `go test ./internal/envprofile -run 'TestManagerOwnedAbsenceReadsAreGuarded|TestEmptyAllowlistWarning' -count=1`.
Record real exit codes before re-handoff.
