# TASK-260927-31gaka results

## Change
- The registry selects revision A. Provisioning preserves the native config.toml bytes, records the sorted native mcp_servers names in codex_seed_record, and emits mcp_native_servers_ungoverned only when that snapshot is non-empty. The warning names the servers, says the next seed revision stops inheriting them, and gives the profile MCP migration hint.
- env status reports the shipped revision and lists the recorded names per managed Codex home as ungoverned. Pre-rule homes now name revision A in the re-provision hint.
- Revision B stripping and status behavior remain covered through an internal revision override seam. No manager-owned reads were added.
- The rc.13 text in protocol/environments.md §§7.4, 7.7, 8.2 and the §12 diagnostic rows was followed. The conformance vectors were archived from CI SPEC_PIN commit 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065; the vector file SHA-256 matched e8fefe0f7eb8fe5692d3a5858367fbe3254f79c7acd4144635f726a35d081c83.

## Conformance coverage
- Revision A provisioning: 3/7 driven through Resolve; revision B: 4/7 exercised through the B seam and recorded as owned gaps.
- Revision A posture: 3/8 driven through StatusOf; revision B-shipped posture rows: 5/8 exercised through the B seam and recorded as owned gaps.
- All nine B-shipped gap rows name TASK-260927-1e5qqm. The verbose vector run reported these ratios and passed.

## Verification
- Pinned rc.13 vector test with -count=1 -v: exit 0.
- Focused Codex provisioning, status, vector, invalid TOML and stateread guard tests: exit 0.
- CLI env resolve/status tests for the A warning and whole-file copy: exit 0.
- internal/envregistry tests: exit 0.
- TestManagerOwnedAbsenceReadsAreGuarded: exit 0.
- go build -o TMP/curator ./cmd/curator: exit 0.
- go vet ./internal/envprofile ./internal/envregistry: exit 0.
- golangci-lint run ./internal/envprofile ./internal/envregistry ./cmd/curator: exit 0, 0 issues.
- gofmt file check and git diff --check: exit 0.

## Mutation checks
- Warning-removed mutant: the rc.13 vector test exited 1 because the A server warning count was zero instead of one. Killed.
- Strip-under-A mutant: the rc.13 vector test exited 1 because seeded_has_mcp_servers was false instead of true. Killed.
- managed.go was restored from the saved source copy and cmp verified byte-for-byte: exit 0.

## Non-passing validation attempts
- The broad go test ./internal/envprofile -count=1 run timed out after the package 10-minute limit and exited 1, with TestSurfacingSinkWriteFailureIsNonFatal active. It is not reported as passing; the focused Codex/status/guard mask passed.
- An exploratory curator-spec-pin invocation with the rc.13 conformance commit exited 1: that command checks the separate rc.8 product release pin f8c405aa3ad0a39d260c2ed93684e55c5a346359. The rc.13 vector source itself was archived from the workflow pin and its hash verified as recorded above.
- Two early focused test runs exited 1 due to test assertion text and an obsolete revision-B assertion; both were corrected, and the final focused runs passed.

## CHANGELOG entry (for release prep)
Ship Codex seed revision A: preserve the native config.toml, record native MCP server names, and warn that the next revision stops inheriting them with profile MCP migration guidance.

No CHANGELOG.md or LOGBOOK.md edits were made. The owned B gaps, local timeout, pin distinction and validation evidence are recorded here.
## Revision 1 — re-applied on 86552087
- Re-applied `refs/campaign/3qf8er-impl-20260928` against the 6bd98d49 base onto 86552087. Resolved the sole `internal/envprofile/status.go` conflict by retaining trunk's stateread seams and the revision test seam. Final changed paths are exactly the seven expected paths.
- Reran `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Guarded'`: exit 0 (112.051s); this includes `TestManagerOwnedAbsenceReadsAreGuarded`.
- First re-apply run of `go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp'`: exit 1 (443.997s), because the schema-1 metadata-only CLI fixture expected B while the unoverridden production path now ships A. Updated that fixture to expect A; B behavior remains covered through the revision seam.
- Reran `go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp'`: exit 0 (311.972s).
- `go build ./cmd/curator`: exit 0. `golangci-lint run ./internal/envprofile ./internal/envregistry ./cmd/curator`: exit 0, 0 issues. `git diff --check HEAD`: exit 0. `gofmt -d` over the changed Go files: exit 0 with no output.
- The revision-A/B conformance vector outcomes and the two mutation kills above are retained from the previously attached verification and were not rerun during this re-apply. Hosted gate is pending the handoff runner; no hosted result is claimed here.

## Revision 2 — stray binary removed; review regression proof
- Addressed review finding F1 by removing the 20,249,680-byte root `curator` Mach-O. Final `git status --short` and `git diff --name-only origin/main -- . ':!.task-board'` show only the seven expected leaf paths; `git ls-files --others --exclude-standard` and the CHANGELOG/LOGBOOK diff check were empty. Path audit exit 0.
- Renamed the existing production-path test to `TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression`; its assertions cover the warning, whole-file byte copy, `codex_seed_record`, and `StatusOf` posture. Clean run before and after mutation restores: `go test ./internal/envprofile -run '^TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression$' -count=1`, exit 0.
- Warning-removed mutant: narrowed warning emission to revision B only. The named test failed because `mcp_native_servers_ungoverned` count was 0 instead of 1; expected mutant-kill exit 1. Restored `managed.go` from a saved copy and `cmp -s` passed, exit 0.
- Strip-under-A narrowing mutant: routed revision A through `stripCodexSeedMCPServers`. The named test failed because managed `config.toml` omitted inherited `mcp_servers`; expected mutant-kill exit 1. Restored `managed.go` from the saved copy and `cmp -s` passed, exit 0.
- Focused package verification: `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Guarded' -count=1`, exit 0 (85.620s), including the stateread guard.
- CLI entry-point verification: `go test ./cmd/curator -run '^(TestPrintEnvStatusShowsCodexSeedPostureAndNames|TestEnvStatusReportsUngovernedNativeCodexServers|TestEnvResolveCopiesAndReportsInlineNativeCodexMCPTable|TestEnvResolveKeepsSchema1BytesForMetadataOnly|TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome)$' -count=1`, exit 0 (27.530s).
- A broader `go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp' -count=1` was interrupted before Go's 10-minute test timeout at about 9:29; real exit 1. It is not claimed as passing; the named CLI tests above were run separately and passed.
- `go build -o /tmp/TASK-260927-31gaka_curator ./cmd/curator`: exit 0. `golangci-lint run ./internal/envprofile ./internal/envregistry ./cmd/curator`: exit 0, 0 issues. `gofmt -d` over all changed Go files: exit 0, no output. `git diff --check HEAD`: exit 0.
- The rc.13 vector test was not rerun in this session because `CURATOR_CONFORMANCE_ROOT` is unset (the environment lookup exited 1). The prior attached results record the pinned rc.13 vector run, its SHA-256, and exit 0; vector implementation is unchanged apart from the separate test rename. The hosted gate is not claimed locally; the corrected handoff runner publishes the CR and runs it after this turn ends.

No CHANGELOG.md or LOGBOOK.md edits were made. The CHANGELOG entry for release prep remains above.
