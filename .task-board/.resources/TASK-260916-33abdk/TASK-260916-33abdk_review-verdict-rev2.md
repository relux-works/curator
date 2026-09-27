# Review verdict — TASK-260916-33abdk rev2 (CR-TASK-260916-33abdk-2): ACCEPTED

Candidate: base 55b94af2, tree a1a14805 (Story worktree re-hashed via a temp index = a1a1480502d8…, exact match). 11 paths, all product/test/.github/ci; no CHANGELOG/LOGBOOK, no stray files.

## Spec check (curator-spec v1.0.0-rc.13, 23435129)
- environments.md §7.4 "Codex seed MCP rule", Revision B: config.toml is copied with `mcp_servers` and every `mcp_servers.*` sub-table removed, the result MUST parse as TOML, and the stripped names are reported once as `mcp_native_servers_not_inherited`. Implemented in managed.go `stripCodexSeedMCPServers` + gatherSeeds, which runs on the production Resolve/provision path. Invalid TOML fails closed (`environment_seed_unreadable`) before the home is published.
- §7.4: "Both revisions write the codex_seed_record … the record is still written with an empty snapshot." The spec therefore REQUIRES an empty record at provisioning. The gatefix note's "omit when empty" would be non-conformant. The producer's rev2 fix is correct: the record is written at provisioning, carried over from `prior` on repair, and kept in the schema-1 legacy projection. A metadata-only repair therefore keeps schema-1 bytes, and TestEnvResolveKeepsSchema1BytesForMetadataOnly passes. A pre-rule schema-1 marker without a record also stays unchanged, because the record is set only on provision or from prior; §7.4 says "an existing home keeps its bytes" and status reports `mcp_seed_unstripped`.
- §8.2 marker: the record is not limited to schema 2. rc.13 schema-case valid-codex-seed-record*.json sits on the schema-1 `valid` base, so removing the `requires schema 2` check is correct. The two Story-owned gap rows now pass and leave conformance-gaps.tsv (count before 2 → after 0 for STORY-260916-1i1gfo). The closed shape is unchanged: revision is limited to A|B, and unknown-field/empty-name cases are still rejected.
- §7.7/§12 env status: prints a per-home sorted names list plus disposition, and a manager `codex-seed` row. A records under a B manager report ungoverned + `mcp_seed_unstripped`; an absent record reports unstripped.

## Bounds judgement (6 bounds = 3 provisioning + 3 posture, all `revision`/`revision_shipped` = A)
These bounds are legitimate. They cover the behaviour of a manager that SHIPS revision A. A manager ships exactly one revision, and this one ships B (registry CodexSeedRevision=B). No other leaf can drive those cases in this binary. A-record homes under the shipped B manager ARE driven.
Residual for the orchestrator (not rework for this leaf): §7.4 says "a manager MUST ship revision A before revision B". Whether curator's release history counts as having shipped the A warning release is a release-sequencing decision for the orchestrator/human.

## Independent reruns (zsh, real exit codes)
- `CURATOR_CONFORMANCE_ROOT=<git archive 23435129>/conformance/v1 go test ./internal/envprofile -run 'Seed|Mcp|MCP|Codex|Status|Guarded' -count=1 -v` → exit 0 (82 PASS, includes TestManagerOwnedAbsenceReadsAreGuarded). Coverage: provisioning 4 driven/3 bound/7; posture 5 driven/3 bound/8.
- `go test ./cmd/curator -run 'EnvResolve|Marker|Credential|Seed|Mcp|EnvStatus|PrintEnvStatus' -count=1` → exit 0.

## Mutants (disposable copy in $TMPDIR)
| Mutant | Result |
|---|---|
| M1 strip removed (return payload before delete) | exit 1 — ProvisioningAndStatus, StripsInlineMCPTable, CodexSeedVectors fail |
| M2 status B-report removed (return nil) | envprofile exit 1 (b-home-lists-not-inherited); cmd/curator EnvStatus exit 1 (TestEnvStatusReportsStrippedNativeCodexServers) |
| M3 provisioning warning suppressed | exit 1 — three tests fail |
All three are killed.

Hosted gate: rev2 validation is green per the orchestrator's review note. I did not re-run the gate myself.
