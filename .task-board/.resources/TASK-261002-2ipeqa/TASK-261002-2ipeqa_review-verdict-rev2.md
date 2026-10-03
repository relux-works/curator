# TASK-261002-2ipeqa — accept-rc14-candidate-suite: revision 2 review

Verdict: changes_requested; route to-dev. No source changes made.
Reviewed exact base 64345d71aa539ef71b9f0b4d3fb9b28d1c9d3fb7 to candidate tree 7d4914a0ad91333b90923db62547b348c50d6389, all 16 paths.

## Finding P1 — candidate validation is missing
Location: .github/workflows/ci.yml:550-552 and scripts/remote-gate.sh (push-triggered validation).
The revision-2 hosted log reports success, but explicitly reports Candidate suite skipped. Independently queried GitHub run 37000203531 (gh exit 0): success, candidate job skipped. Its head 2c76441b22a3f7296feb883ca54b88aaff7e2a25 resolves to the exact candidate tree above, so tree attribution is correct; suite attribution is not. The workflow admits candidate validation only on workflow_dispatch with candidate_ref/root, whereas the remote gate pushes a gate branch. Default SPEC_PIN remains rc.13. Thus the green default matrix does not establish a green rc.14 matrix or the spec's exact candidate Implementations check.

The attached Revision 2 results explicitly list the exact six-package Implementations command and consumption gate as unrun. Revision 1 passes are historical and cannot establish this changed revision. This is an evidence gap, not a demonstrated product defect or external blocker.

Required rework: validate the same frozen candidate tree on hosted runners with candidate_ref=e3a88cedba7a844c594348ee89654471db086f98 and candidate_manifest_sha256=6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5; attach an actually executed green candidate lane and run the spec command against that root: go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy, followed by tools/implementation_coverage.py go --stream <real-json>. Attach real exits and authenticated tree/root identities, then hand off for another review. Keep SPEC_PIN and writer switch unchanged. No additional local stalled-test retries requested.

## Swept surfaces
| Surface | Result |
| --- | --- |
| Digest and corpus | Independently authenticated manifest and all 1294 manifest entries from candidate checkout at e3a88ced; Python exit 0, 0.182s. |
| Independent family counts | agent-environment-marker-v3/schema-cases 29; launch-env-fragment-v3/schema-cases 36; environments-muse/cases 16. All match pins. Independently recomputed 3/103 families; ledger totals 103 families / 1870 entries. Producer's separate 103/103 audit is supporting evidence, not my independent recomputation. |
| Gaps | Nine rc.14 rows, exclusively posture/provisioning. All seven fixed cases absent (five marker cases have v3/v4 representations). Four stale historical-candidate Windows rows removed consistently with the trunk fix. |
| Older suites | All prior count rows byte-for-byte unchanged (Python assertion exit 0); dispatch retains rc.13/hash-v2/Muse distinctions. Historical production suite replay not performed by reviewer. |
| Consumers and negatives | Reviewed all changed test consumers and coverage.go dispatch; rc.14 uses implemented hash-v2 plus Muse. Closed unknown/near digest and protocol refusal tests remain. No implementation defect found in this sweep. Fresh mutation testing not performed. |
| Switches and hygiene | Writer false; ci.yml, internal/hashing and LOGBOOK unchanged against base (git diff exit 0). One CHANGELOG addition retains trunk entries. Whitespace check exit 0. |
| Validation | Hosted default matrix/lint green, candidate skipped. Accepted producer rev2 candidate/default targeted exits 0 as attached evidence; not rerun myself. Exact candidate spec shape unverified on rev2. |

## Host and lifecycle
Reviewer started no Go build/test/vet/run and no local lint/build; therefore no new long-command duration or crash pair is claimed. Read-only launchctl observed running, successive crashes 369 (exit 0). Binding host decision and prior two stalled attempts respected. No LOGBOOK write. Findings persisted here and in board notes.
Queried spawn goal: run is not goal-bound. Changes requested is the sole verdict branch; ordinary validation rework routes to-dev, not blocked.
