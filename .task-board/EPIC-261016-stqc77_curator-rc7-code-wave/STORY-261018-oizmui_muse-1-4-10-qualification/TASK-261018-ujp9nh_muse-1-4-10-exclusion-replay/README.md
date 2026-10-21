# TASK-261018-ujp9nh: muse-1-4-10-exclusion-replay

## Description
Run the Curator foreign-context exclusion replay P1-P4 against Muse 1.4.10 for exec and interactive forms in temporary homes; publish a qualification record per form (qualified or refused with evidence). Starts after the shared Muse quota resets on 2026-10-19.

## Scope
New .research/muse1410/ opt-in harness and docs/qualification/muse-1.4.10-R6427.1.md, mirroring the 1.4.8 record. Exact vendor binary muse-bin-1.4.10-R6427.1 through MUSE1410_BINARY, provider echo, sandbox-exec with network and Keychain denied, disposable homes. No runtime constants, registry pins, qualification gates, credential policy, CHANGELOG.md or root README.md changes.

## Acceptance Criteria
1. P1-P4 replayed for exec and for interactive, each row recorded with its observation and exit code.
2. The record states the environments 7.9 gate status per form, and states whether the 1.4.8 interactive failure persists in 1.4.10.
3. The binary digest is recorded and the harness refuses any other digest; without MUSE1410_BINARY all vendor tests skip before creating a fixture.
4. Offline controls hold: network and Keychain are denied, there is no auth command, and there is no remote inference.
5. A narrowing control shows that a planted foreign-skill leak fails its P-row.
6. go build, go vet and the default go test pass; the gated vendor run is recorded with exit codes; public-repo hygiene holds.
