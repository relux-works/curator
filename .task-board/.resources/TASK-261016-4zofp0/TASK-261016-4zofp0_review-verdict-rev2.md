# SNAP-LOCK revision 2 review verdict

**ACCEPT**, capture-stage scope. Findings: **0**. This accepts the candidate; it does not claim landing or delivery of the deferred closure/CLI integrations.

Element: TASK-261016-4zofp0 — implement-rc7-snapshot-nonrecovering-capture.
Base: `dccac3199aa5c543c589c129f5ef2bc437523a2a`.
Candidate tree: `f0b08b96c097fcf57c1927355f9874079df23e5b` (repository delta present, 16 paths).

## Verdict data

```json
{
  "verdict": "accepted",
  "revision": 2,
  "candidate_tree": "f0b08b96c097fcf57c1927355f9874079df23e5b",
  "findings": [],
  "notes": [
    "Scope is the capture transaction, not SNAP-CLOSURE inventory or SNAP-CLI publication: 5/5 capture-stage AC rows driven, 0/2 CLI forms qualified.",
    "No catalog surface table or review budget was supplied; the five mandatory paired rows were swept sequentially.",
    "The bounded local callback contract is cooperative; arbitrary callback blocking is not preempted.",
    "Windows POSIX sentinel/permission and some held-object replacement plants remain explicitly unavailable; this review ran on darwin/arm64.",
    "Fresh upstream contains disjoint changes; combined-head validation and landing are still required from the delivery owner."
  ],
  "surface_results": [
    {
      "row": "R1",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-sweep-rev2.md"
    },
    {
      "row": "R2",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-sweep-rev2.md"
    },
    {
      "row": "R3",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-sweep-rev2.md"
    },
    {
      "row": "R4",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-sweep-rev2.md"
    },
    {
      "row": "R5",
      "result": "held",
      "evidence": "TASK-261016-4zofp0_review-sweep-rev2.md"
    }
  ],
  "free_hunt": []
}
```

## Contract and scope

Read curator-spec main at `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`: managed-home-snapshots sections 1 and 5, environments sections 9.2/10.1, and CIP-0014 consistency steps 1–5. No conflict with the supplied capture leaf was identified. Recovery observation is non-recovering, selection is repeated under the shared lock, content verification avoids vendor execution, and object identity revalidation supplements fragment equality.

Coverage is **5 of 5 capture-stage AC rows driven**, with named tests and production call sites in the retained coverage map and the attached review sweep. **0 of 2 CLI forms are qualified by this leaf**; the explicit SNAP-CLOSURE/SNAP-CLI integration split is accepted, not described as shipped materialize functionality. The callback must provide bounded local work, protected handles and the admitted closure. No archive, destination, credential-provider or live vendor behavior is qualified here.

All five supplied paired surface rows are held under the named attacks in `TASK-261016-4zofp0_review-sweep-rev2.md`. Results were appended after each row before the next row ran. Held means those attacks did not reproduce; it does not prove absence of defects. Free hunt has no findings: reviewer-only late corrupt journal injection and repair/dry-run/preserve-lock input attacks reached CaptureMaterialize, refused with the intended diagnostics and preserved fixture state beside successful controls.

The 16 changed paths match the declared implementation and the producer's justified seam exceptions: managed/status verification, managerlock existing-state acquisition, transaction recovery instrumentation, platform test fixtures, README and signed remote-gate snapshots. Existing tests were retained. No product code was edited by this review. The M3 plant and reviewer-only tests used an isolated archive of the exact candidate tree.

## Independent execution

- `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run 'Test(MaterializeCapture|PendingRecovery|AcquireExistingHomeOnly|CaptureLock)' -count=1`: exit 0.
- Five sequential row-specific commands, each with `-count=1 -timeout=2m`: exit 0 each; full commands and logs are in the attached sweep and execution evidence.
- `go test ./internal/envprofile -run '^TestReviewCapture' -count=1 -timeout=2m`: exit 0, isolated candidate plus the attached reviewer test file.
- M3 narrowing plant retains profile-lock object validation but omits other object identities. `go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/aba$' -count=1 -timeout=2m`: exit 1 at the intended assertion (capture accepted instead of environment_materialize_capture_drift). The harness verified that failure and restored the original bytes. This is behavioral, not a static-token checker.
- The other eight producer narrowing-control results were inspected and reused under the revision-2 unchanged source/test/configuration identity, not rerun by this reviewer. Source equality and the same Go 1.25.5 darwin/arm64 platform support reuse; M3 was independently repeated as required by this leaf.
- `git diff --check`: exit 0. All 16 working candidate blobs were rechecked against the exact tree and remained identical.
- `sh scripts/remote-gate.sh`, after loading the configured agent Git environment: **actual process exit 0**. [Reviewer hosted run](https://github.com/relux-works/curator/actions/runs/38895758815): **20 enabled jobs success**, two workflow-optional jobs skipped. The signed snapshot `ce2172f43935fc7604cac4ce44f7470e2bbe9371` has the exact candidate tree and signature `G`. This full gate was rerun by the reviewer, not inferred from the producer's earlier green.

Evidence: `TASK-261016-4zofp0_review-execution-rev2.md`, `TASK-261016-4zofp0_review-sweep-rev2.md`, `TASK-261016-4zofp0_review-extra-rev2_test.go`, `TASK-261016-4zofp0_review-gate-rev2.json`, `TASK-261016-4zofp0_review-identity-rev2.json`.

## Delivery bounds

Fresh upstream `58b2aac337f73fb313893611d9de28e9a4fcff0a` is ahead of the candidate base, with no changed-path overlap. This review covers the exact supplied candidate, not a future combined landing tree. Delivery owner must satisfy current-head PR/check/signature obligations. Reviewer goal read returned no bound goal. The live merged checklist readback has 30/30 checked. Acceptance uses accept_cr revision 2, never commit_ack or a reviewer done transition.
