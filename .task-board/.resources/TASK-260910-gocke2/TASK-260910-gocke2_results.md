# TASK-260910-gocke2 results

## Outcome

The assigned branch already contains the S4 production implementation, CLI integration, and RC13 vector runner. This task's working-tree change strengthens the allowlist-warning tests in `internal/envprofile/surfacing_test.go`; it does not modify production code. The test additions drive empty and non-empty allowlist behavior through install, update, and status, and prove the empty-allowlist warning does not make a current status non-current.

Normative source: curator-spec v1.0.0-rc.13, `protocol/environments.md` §§2.2, 2.3, 9.1/9.2, 10.3, 12, and `conformance/v1/vectors/environments-env-passthrough.json`.

## Files and acceptance coverage

- `internal/envprofile/surfacing_test.go:142-192`: `TestAllowlistWarningOperations` checks `mcp_package_allowlist_empty` is emitted on install, update, and env status when the effective allowlist is empty, and omitted for a non-empty allowlist.
- `internal/envprofile/surfacing_test.go:194-242`: `TestEmptyAllowlistWarningLeavesCurrentStatusCurrent` creates current installed profile/environment surfaces, establishes a quiet non-empty-allowlist control, then proves that adding the empty-allowlist warning leaves `Status.NonCurrent` false.
- Existing `internal/envfragment/envfragment.go:217-240,252-315` implements both profiles behind `ActiveS4Profile`, ships `s4-warn`, applies the absent/null distinction, excludes reserved names, and emits the variable names and migration hint for the warning profile or dropped-name diagnostic for enforce.
- Existing `internal/envprofile/managed.go:1943-1956` calls the resolver from the production resolution path and carries its warning.
- Existing `internal/envprofile/surfacing.go:12-22,24-58,61-86` implements the empty-allowlist warning and declaration-row construction/emission. `internal/envprofile/envprofile.go:758-790,908-935` places install/update rows after the audit/admission gate and before lock publication. `internal/contextmaterialize/mcp.go:143-178` formats compact closed-column rows.
- Existing `internal/envprofile/status.go:304-317,770-789` reports active S4 posture, effective passable names, allowlist warning, and scoped declaration rows. `cmd/curator/envstatus.go:35-45` renders the posture and rows; `cmd/curator/profile_surfacing_test.go:74-119` drives the install CLI and checks stdio output, warning, and status posture.
- Existing `internal/envprofile/envpassthrough_conformance_test.go:20-98` reads the pinned vector family from `CURATOR_CONFORMANCE_ROOT` and drives production resolution, config parsing, install/update/status, row formatting, and publication ordering. The RC13 root provides the family, so no root-content skip is used.

The S4 resolver and surfacing implementation were present at the task's base commit `0ffe2e1db9400dd18de3c3a0241275295c481b0e`; this task's patch is intentionally limited to the two additional warning assertions/tests described above.

## Conformance ledger

`.github/ci/conformance-gaps.tsv` has 70 data rows at base and 70 after this task; there were 0 rows owned by `STORY-260910-1lf0m5` or `TASK-260910-gocke2` before and 0 after. No S4 rows were present to remove or reattribute. Two manager-config security-posture cases remain assigned to `STORY-260910-2qmrb8` (S1/S3); they are outside this task's scope and were not modified.

## Validation transcripts

All commands below ran in the worktree. The conformance root was extracted from the exact curator-spec rc.13 pin `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` into `/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/curator-spec-rc13.dZH4Ic/conformance/v1`.

- `go build ./...` — exit 0.
- `go vet ./...` — exit 0.
- `gofmt -l .` — exit 0; it reports pre-existing unformatted Go attachments under `.task-board/.resources/**`, which are board artifacts and were not edited. `git ls-files -z -- '*.go' ':(exclude).task-board/**' | xargs -0 gofmt -l` — exit 0, no product Go files listed.
- `golangci-lint run` — exit 0, 0 issues.
- `git diff --check` — exit 0.
- `go test ./internal/config/... -count=1` — exit 0.
- `go test ./internal/envfragment/... -count=1` — exit 0.
- `go test ./internal/contextmaterialize/... -count=1` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/curator-spec-rc13.dZH4Ic/conformance/v1 go test ./internal/envprofile -run '^TestEnvironmentsEnvPassthroughVectors$' -count=1 -v` — exit 0; all 25 vector cases driven, with 0 gaps, bounds, or skips.
- `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/curator-spec-rc13.dZH4Ic/conformance/v1 go test ./internal/envprofile -run '^(TestEnvironmentsEnvPassthroughVectors|TestInstallSurfacesMCPDeclarations|TestInstallSurfacingSilentWithoutMCP|TestUpdateSurfacesCandidateMCPSet|TestAllowlistWarningOperations|TestResolvePassthroughKnobs|TestStatusS4Posture|TestSurfacingUnreadableManifest|TestInstallEmitsSurfacingBeforePublication|TestUpdateEmitsSurfacingBeforePublication|TestReinstallEmitsSurfacingBeforePublication|TestSurfacingEmittedDespitePublicationFailure|TestSurfacingSinkWriteFailureIsNonFatal|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent)$' -count=1` — exit 0 (114.558s).
- `go test ./internal/envprofile -run '^(TestAllowlistWarningOperations|TestEmptyAllowlistWarningLeavesCurrentStatusCurrent)$' -count=1` — exit 0 (20.462s).
- `CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T/curator-spec-rc13.dZH4Ic/conformance/v1 go test ./cmd/curator -run '^(TestProfileInstallListsStdioCommands|TestProfileUpdateListsNewDeclaration|TestProfileInstallWarnsEmptyAllowlist|TestProfileInstallSurfacesBeforePublication|TestProfileUpdateSurfacesBeforePublication|TestEnvStatusS4Posture|TestEnvStatusEffectivePassableList)$' -count=1` — exit 0 (40.320s).
- `go test ./internal/envprofile/... -count=1` — exit 1 after 554.239s; the command was interrupted at the bounded shell time limit (`signal: interrupt`). Treat the broad package run as incomplete, not passing. The focused envprofile runs above passed.

The full hosted landing gate was not run locally; the task rules assign it to the handoff runtime.

## Mutation evidence

Temporary production-source mutations were restored after each run. Each mutant was killed (real exit 1):

1. Suppress the absent-knob `s4-warn` warning — `TestResolvePassthroughProfiles/warn-absent-passes-with-warning`.
2. Make absent-knob `s4-enforce` unbounded — `TestResolvePassthroughProfiles/enforce-absent-drops-all`.
3. Treat explicit null as empty — `TestResolvePassthroughProfiles/warn-null-unbounded-silent`.
4. Disable reserved-name exclusion — `TestResolvePassthroughProfiles/warn-absent-reserved-excluded`.
5. Mute empty-allowlist warning — `TestAllowlistWarningOperations`, including install/update/status.
6. Suppress the CLI allowlist warning — install, update, and `env status` CLI warning tests.
7. Change the formatted surfacing column `args=` to `argv=` — `TestFormatDeclarationRows`.
8. Move install surfacing after `op.publish` — `TestInstallEmitsSurfacingBeforePublication`.
9. Move update surfacing after `op.publish` — `TestUpdateEmitsSurfacingBeforePublication`.
10. Report `s4-enforce` in posture — `TestStatusS4Posture`.
11. Mark a current status non-current when the empty-allowlist warning is present — the new `TestEmptyAllowlistWarningLeavesCurrentStatusCurrent`.

The initial status fixture for mutant 11 already had unrelated non-current rows and did not kill the mutant. The added clean-current fixture closes that evidence gap; the narrowed mutant then exited 1.

## Release-prep entry

## CHANGELOG entry (for release prep)

S4: ship the warning release (`s4-warn`) for MCP environment passthrough. With an absent `passable_env_names` knob, passthrough remains unbounded and emits `mcp_env_passthrough_unlisted`; list the named variables to keep passing them after the flip. `s4-enforce`, which drops unlisted names, follows in a later release.

## Handoff notes

No CHANGELOG or LOGBOOK file was edited, per the current campaign rules. No production changes were required in this worktree because the implementation was already present at the assigned base. The only worktree delta is the added warning test coverage in `internal/envprofile/surfacing_test.go`.
