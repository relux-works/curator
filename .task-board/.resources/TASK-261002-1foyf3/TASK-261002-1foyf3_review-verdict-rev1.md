# TASK-261002-1foyf3 — rc14-pin-and-v2-writer-cutover: revision 1 review

Verdict: accepted under the binding pin-only rescope. No blocking findings. Acceptance covers the rc.14 conformance pin with v1 writers retained, not a v2 writer cutover.

CR: CR-TASK-261002-1foyf3-1 revision 1. Base 68210eccfd656e770cf9101c2b927ceda01359df; candidate tree d529101a03721c331f3d0131d7da4a26f4a380db. Fresh origin main advertisement and exact-ref fetch both resolve to the base. Working-tree tracked contents match the candidate. No product code modified by reviewer.

## Swept surfaces

| Surface | Review result |
| --- | --- |
| Workflow pin and docs | Final peeled rc.14 commit 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 agrees across workflow, docs and count-ledger comment; no withdrawn daf15ec8 pin in delta. |
| Release identity | Fresh tag advertisement: object 661bead088186db70c9f57ad0b115305ec2bef35, peeled 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. Tagged-tree manifest and clean checkout both hash to 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5; commands exit 0. |
| Writer selection and compatibility | EnableV2Writers remains false. Only comments/test failure wording change in hashing. Frozen v1 default and existing v2 seam tests pass. No marker/install/context/envprofile flip adaptations in delta; rc14_cutover_test.go absent. |
| Gap ownership and exact counts | Exactly one rc.14 snapshot-acquisition/cases gap: byte-exact-snapshot, owner TASK-261003-1uzji7, exact agreed migration reason. Count remains 1. No rows removed. Nine unrelated seed/posture B rows retained (10 total rc.14 rows); “one gap” refers to snapshot family, not all rc.14 families. |
| Production acquisition and negatives | Snapshot test drives gitops.Extract, called from snapshot.go:91/131 and closure.go:510. Both autocrlf settings check bytes, paths, unexpanded export-subst and line endings. rc.14 uses WriteVersion; historical vectors use frozen v1. Existing coverage negatives reject missing classification, unlisted failure, passing gap and vanished case/count. |
| Tally | Producer's attached verbose rc.14 run measures 0 driven + 1 known-gap + 0 bound + 0 skipped = 1/1. Reviewer reran entire required package list against rc.14 successfully; ledger regression verifies owner, reason and exact tally. This is accounted coverage, not a claim that v2 snapshot hashing passes. |
| Protected files and constants | CHANGELOG.md, LOGBOOK.md, rc.8 release_pin.go unchanged. CodexSeedRevision and SecurityPostureRevision remain A. No unrelated implementation changes. |
| Architecture | Independent conformance pin advances while immutable module pin and write format remain fixed. Migration ownership is explicit. No new production behavior or bypass introduced. |

## Hosted gate and retained producer evidence

CR validation resource TASK-261002-1foyf3_change-request_rev1-validation.log records remote-gate exit 0. Independently queried https://github.com/relux-works/curator/actions/runs/37150222861 : success, head bf1524f170ef5053e9c453abccbda1635a430ca9, whose tree is exactly d529101a03721c331f3d0131d7da4a26f4a380db. Default Test matrix Ubuntu/macOS/Windows, Race Ubuntu/macOS, lint, interop gate, naming and gate self-tests passed. Workflow and fetched macOS job log confirm SPEC_PIN and actual spec checkout at 43bf0a25. Thus hosted default rc.14 matrix IS included. Separate workflow-dispatch Candidate suite and self-hosted rose-air were skipped; no claim is made for those lanes.

Accepted producer evidence from TASK-261002-1foyf3_pin-only-results.md: build exit 0; initial stale-cache lint exit 1 followed by isolated-cache pinned lint exit 0 (also independently corroborated by hosted lint); verbose snapshot tally above. Did not rerun full hosted suites or lint locally. git diff --check and scope assertions exit 0.

## Independent bounded verification

All reviewer Go calls used -work and the shared directory lock, and ran to completion before verdict. Required command: go test -work ./internal/hashing ./internal/marker ./internal/conformancecoverage ./internal/interop/... -count=1.

| Source/corpus | Exit | Duration seconds | syspolicyd crashes before → after |
| --- | --- | --- | --- |
| Candidate / actual rc.14 checkout | 0 | 11.06 | 381 → 381 |
| Base 68210ecc / producer's rc.13 archive corpus | 1, expected reproduction | 20.34 | 381 → 381 |
| Candidate / actual rc.13 checkout | 0 | 5.96 | 382 → 382 |

Base was materialized into an isolated temporary source directory without altering the Story branch. Same full command and same archive corpus reproduce the original subst.txt mismatch: actual 65 bytes, SHA-256 a1ab8edbd48667c39da619a7cc4bad17fe5e96eb218da72de150ae7fa4e93849; expected 40 bytes, ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc. Therefore the archive-materialization failure predates the delta. Actual rc.13 checkout is 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065 with manifest be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca, and candidate passes against it.

Host limitation: service was running before/after each call, but crash count increased between the base call and final compatibility call. The inherited wrapper checked running state and launched the final call before the increased count was noticed; the requested five-minute cooldown was not applied for that call. No further Go work was run. Final compatibility exited 0 with stable 382 → 382; hosted success provides independent validation. This operational deviation is recorded rather than represented as full host-rule compliance.

Run goal queried before verdict: not goal-bound. LOGBOOK.md deliberately unchanged under binding rescope; review findings and limits live in this outcome. Acceptance routes to integrating through accept_cr; reviewer does not commit, acknowledge or mark done.

### review-rc14

```text
WORK=<retained-work>
ok  	github.com/relux-works/curator/internal/hashing	6.561s
ok  	github.com/relux-works/curator/internal/marker	9.227s
ok  	github.com/relux-works/curator/internal/conformancecoverage	8.435s
ok  	github.com/relux-works/curator/internal/interop	8.774s
ok  	github.com/relux-works/curator/internal/interop/environments	9.758s
```

### review-base-rc13-archive

```text
WORK=<retained-work>
ok  	github.com/relux-works/curator/internal/hashing	9.845s
ok  	github.com/relux-works/curator/internal/marker	6.406s
ok  	github.com/relux-works/curator/internal/conformancecoverage	5.551s
ok  	github.com/relux-works/curator/internal/interop	7.022s
--- FAIL: TestConformanceSnapshotAcquisition (0.37s)
    --- FAIL: TestConformanceSnapshotAcquisition/byte-exact-snapshot (0.37s)
        snapshot_acquisition_test.go:97: fixture subst.txt on disk (sha256:a1ab8edbd48667c39da619a7cc4bad17fe5e96eb218da72de150ae7fa4e93849, 65 bytes) does not match the vector (sha256:ec9a6c8c260b3977dad2f1cba09414599879ac5c3e7353c3205a40d5fbe122bc, 40 bytes); the checkout normalized it
    coverage.go:233: published-case coverage: failing published case snapshot-acquisition/cases/byte-exact-snapshot is not listed in the gap ledger: case test failed
    coverage.go:235: published cases snapshot-acquisition/cases: 0 driven, 0 known-gap, 0 bound, 0 skipped, 0 total
FAIL
FAIL	github.com/relux-works/curator/internal/interop/environments	7.995s
FAIL
```

### review-rc13-checkout

```text
WORK=<retained-work>
ok  	github.com/relux-works/curator/internal/hashing	1.177s
ok  	github.com/relux-works/curator/internal/marker	4.212s
ok  	github.com/relux-works/curator/internal/conformancecoverage	2.288s
ok  	github.com/relux-works/curator/internal/interop	3.235s
ok  	github.com/relux-works/curator/internal/interop/environments	4.532s
```
