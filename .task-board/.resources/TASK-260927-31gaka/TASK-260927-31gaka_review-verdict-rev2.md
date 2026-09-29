# TASK-260927-31gaka review verdict — rev2: ACCEPTED

Candidate tree 31e7a36f (base 86552087). The worktree matches the candidate (`git diff 31e7a36f` is empty). 7 paths, no stray files.

## Rev1 F1 (stray binary)
Fixed: `git diff ed98b5b1 31e7a36f` removes `curator` (20 MB) and changes one line: codex_seed_test.go:60 renames TestCodexSeedProvisioningAndStatus → TestCodexSeedRevisionAWholeCopyWarningAndPostureRegression. The body is unchanged and nothing references the old name. The comment at :57 still uses the old name (cosmetic, non-blocking).

## Review-note items
1. envregistry.go: `CodexSeedRevision = CodexSeedRevisionA` is the single switch, and the codex_cli adapter uses it. In managed.go gatherSeeds, revision A copies the payload unchanged (parse only). The test asserts the seeded bytes equal the native bytes. The record is {A, sorted names}. The `mcp_native_servers_ungoverned` warning fires only when len(names)>0. It names the servers, says "the next seed revision stops inheriting them", and gives the hint "Declare each server in the profile's MCP set, or accept the loss." Empty snapshot: no warning, but the record is still written (vectors a-without-servers-no-warning).
2. status.go: an A record lists the names with disposition "ungoverned". mcp_seed_unstripped is added only when a revision-B manager meets an A record. A home without a record reports "unknown" plus mcp_seed_unstripped naming the shipped revision. The pre-rule-home-unstripped-under-a/b vectors pass.
3. The B implementation is retained behind the unexported `codexSeedRevisionForTest` seam (Resolve + StatusOf; values other than A/B are rejected). The B tests are kept. conformance-gaps.tsv has 9 B rows owned by TASK-260927-1e5qqm. The vector runner returns FailureReason for B and fatals on unknown revisions.
4. See the mutants below. No new manager-state reads (the existing seed read path is reused), and TestManagerOwnedAbsenceReadsAreGuarded is in the passing envprofile run. Existing homes: TestCodexSeedSnapshotAndBytesDoNotRefreshOnRepair passes. No CHANGELOG/LOGBOOK edits.

## Independent runs (zsh, pipefail, git archive copy of 31e7a36f, CURATOR_CONFORMANCE_ROOT=curator-spec rc.13+1 conformance/v1)
- go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Guarded' → ok, EXIT 0
- go test ./cmd/curator -run 'EnvResolve|Marker|EnvStatus|Seed|Mcp' → ok 411s, EXIT 0
- TestEnvironmentsCodexSeedVectors -v: all 15 subcases PASS (A rows driven, B rows via the seam and gap-ledgered)
- M1 (A warning wrapped in `if false`): FAIL, EXIT 1 (RevisionA test + vectors) — killed
- M2 (A branch uses stripCodexSeedMCPServers): FAIL, EXIT 1 — killed
Hosted gate: accepted from the board (green); not rerun.
