# TASK-261021-mw4laf: muse-1-4-11-exclusion-replay

## Description
Run the Curator foreign-context exclusion replay P1-P4 against Muse 1.4.11 (1.4.11-R6667.1) for exec and interactive in sandboxed temporary homes, using the exact pin-store binary; publish a qualification record per form (qualified, or refused with evidence) and compare each row with 1.4.10. The operator launcher auto-updated to 1.4.11, so task-board muse stays fail-closed until this record exists.

## Scope
New .research/muse1411/ opt-in harness (reuse the muse1410 harness and fixtures where unchanged) and docs/qualification/muse-1.4.11-R6667.1.md. Exact binary from the task-board Muse pin store, provider echo, sandboxed offline. No runtime constants, registry pins, gates, credential policy, CHANGELOG.md or root README.md changes.

## Acceptance Criteria
1. P1-P4 replayed for exec and interactive, each row recorded with its observation and exit code.
2. The record states the environments 7.9 gate status per form and compares it with 1.4.10.
3. The binary digest 6c04ce17...84d9 is pinned and the harness refuses any other digest; without MUSE1411_BINARY all vendor tests skip before creating a fixture.
4. Offline controls hold: network and Keychain denied, no auth command, no remote inference.
5. A narrowing control shows that a planted foreign-skill leak fails its P-row.
6. go build, go vet and the default go test pass; the gated vendor run is recorded with exit codes; public-repo hygiene holds.
