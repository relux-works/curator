# Delta review note — TASK-260916-3oh0u8 rev3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev2 was ACCEPTED on content. Rev3 re-applies it on trunk eca2bf27 (tree 1feb9d60, base eca2bf27, gate green). Orchestrator checks done:
same 6-path set; 4 paths byte-identical to rev2 (platform-cases.tsv, status_test.go, umbrella.go, umbrella_test.go); on the 2 intersecting
paths (cmd/curator/main.go, cmd/curator/envstatus.go) rev3 has 0 lines present in neither trunk nor rev2, and 0 trunk-new lines dropped;
2n0233's registry-posture rows (formatRegistryPosture, ReadBoundaryPosture, registryRows) and rev2's provider posture (providerPostureForConfig,
provider_diagnostic) are both present.
Confirm by reading the combined hunks of those two files (ordering/semantics: status --check currentness combines provider AND registry
posture; JSON payload carries both), run `go build ./...` and `go test ./cmd/curator -run 'Umbrella|Provider|EnvStatus|Status|Boundary|Attest'`
with real exit codes. accept_cr or changes requested with file:line. No LOGBOOK.md.
