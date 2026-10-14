# TASK-261014-2f0qnv — recovery evidence


## Successor recovery and handoff validation

No production, test or configuration bytes changed during recovery. SHA-256 verification of all ten candidate paths matched the attached identity, and the previously green snapshot `6706a888fdcc4777fa92b8a809cab0554940f2ee` and handoff snapshot `9ba1c76ada3ed140f595c3ae0c8b3b45ad2322cc` share tree `265d93f4093227dfa6e1a05ea49c72dbb5a3551e`.

The first automatic Change Request validation ran `sh scripts/remote-gate.sh` and exited **1**: hosted run https://github.com/relux-works/curator/actions/runs/38551230578 failed in `TestCompiledProjectRepairsCorruptCompiledState/install` at `status_test.go:894` (empty status result). The other jobs succeeded; all ten preserve-lock vector cases and all task-specific tests passed in the failed macOS job's uploaded JSON stream. This failed gate remains recorded as failed, despite the earlier green run. Its failure was not classified as an external CI failure; no local-mirror fallback was used. The hosted failure's cause remains unknown.

The successor personally reran these standalone commands on the unchanged candidate (Go 1.25.5, darwin/arm64):

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go build ./...` | 0 | recovery-build-01.log |
| `go vet ./...` | 0 | recovery-vet-01.log |
| `go test ./internal/envprofile -run 'TestRepairPreserveLock|TestPreserveLock|TestPlainRepairMigrationKeepsLockPrivate' -count=1 -v` | 0 | recovery-env-01.log; 10 of 10 vectors driven |
| `go test ./cmd/curator -run '^TestCompiledProjectRepairsCorruptCompiledState$' -count=1 -v` (inherited PATH) | 1 | recovery-compiled-01.log; both subcases failed at status --check after repair, a different assertion from the hosted failure |
| Same compiled-state command with PATH limited to the active Go toolchain and system tools | 0 | recovery-compiled-controlled-02.log; original install and upgrade test names/assertions unchanged |
| `go test ./cmd/curator -run 'TestEnvResolvePreserveLock|TestGlobalInitEnvResolveSubcommandHelp' -count=1 -v` (same controlled PATH) | 0 | recovery-cli-01.log |
| `git diff --check` | 0 | recovery-diff-check-01.log |
| `gofmt -l` on the eight changed/new Go files | 0 | recovery-gofmt-01.log; empty output |

The earlier ten-mutant results are reused, with their named failing tests and bounds above, because the candidate source and test bytes are unchanged. No new mutant or full local suite was run by the successor. The managed Story checkpoint remains `84cf54b94670ae9881f91b1c147feb4b75afddee`; observed local main `352c5c460739ec5849d677cad62c76875ed49f81` is six commits ahead. Evidence here is for the checkpoint candidate, not current trunk. No branch switch, commit or rebase was performed; upstream convergence belongs to the orchestrator.

The required `task-board handoff TASK-261014-2f0qnv --role developer` is being retried after these diagnostics. Its generated Change Request validation outcome is the authoritative result of that new hosted gate; the earlier green run alone does not satisfy a failed handoff. This recovery note is attached before invoking the synchronous handoff.
