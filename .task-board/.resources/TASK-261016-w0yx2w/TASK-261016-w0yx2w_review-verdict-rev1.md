# Review verdict — SNAP-PUBLISH revision 1

Task: TASK-261016-w0yx2w — implement retained-handle snapshot destination publication.
Candidate: CR-TASK-261016-w0yx2w-1 revision 1, tree `2e20b225f2794c9e524ee69332ed156424c612d8`, base `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`.

## Scope and normative basis

Read current curator-spec main `f9970ddf01ff6b52faaed3721ef49ce9dd9c5ca7`: managed-home-snapshots sections 2/4/5, environments section 5, and CIP-0014 DestinationBoundary/Idempotence rows. Absent destination or fully verified exact prior only; private unpredictable staging; no-replace publication; complete post-publication verification; failure preserves published output without rollback. Same-user mutation inside private staging is outside the guarantee.

Coverage is **5 of 5 leaf AC rows driven through the library production API**. Named tests and call sites are preserved in the attached sweep and producer Round 3 matrix. **0 of 5 CLI rows driven**, explicitly excluded by this leaf's current instructions. A caller search confirms there is no materialize integration; no CLI qualification is inferred from helper tests. The coordinator must supply complete protected roots, authorized capture bytes and verified execute metadata. Windows has explicit refusal only.

The 21 changed paths match the reviewed tree. Workflow routing, runner documentation, README and CI-routing tests outside the original package seam are explicitly justified by the CI-owner Round 3 brief. No product code was changed by the reviewer; supplemental tests run in a scratch archive of the exact candidate. Fresh origin/main equals the CR base.

## Swept surfaces

The brief supplies four paired rows, not a catalog-keyed surface table. All four are retained in supplied order; no row is omitted.

| Row | Result | Public entry attacks |
|---|---|---|
| Admission and protected identities | held | OpenDestination: linked ancestors/target, shared mode, wrong owner, kernel ACL, absent/empty boundaries, protected equality/inside/containing aliases with positive private controls |
| Replacement and destination competition | held | Revalidate/Publish: ancestor and identical-target replacement; mkdir/file/link/prior competitors after last check before real no-replace syscall; retained-parent rename window |
| Actual prior identity | held | MatchPrior/Publish: modified bytes/modes/extras/missing metadata/unreadable/unknown version/new generation; exact prior no-op positive control |
| Staging failure, cleanup and verification | held | Publish: context interruption, real kernel permission failure, substituted staging root, foreign sibling link, post-rename bytes/size/mode/extra/metadata/link mismatch without rollback |

`held` describes only these executed attacks, not proof of absence. All 16 narrowing controls produced their intended named child failure, exit 1; the harness exited 0. In particular M2, M6, no-replace, cleanup scope and post-verification weakening were exercised. These are behavioral child runs, not static-token assertions.

Free hunt found no blocking defect. Supplemental `TestReviewerPublicationRetry` covers clean retry after cancellation, refusal to adopt a competing identical snapshot, and refusal after a published mismatch while retaining the extra file. `TestReviewerStageRootBeforeRename` substitutes the top-level staging name before the syscall and verifies identity refusal without destructive rollback. Both named tests and all retry subtests passed under `-race`.

## Evidence

- `TASK-261016-w0yx2w_review-validation-rev1.txt`: focused package with race detector, 16 narrowing controls, supplemental attacks and CI-routing LF/CRLF controls, all exit 0.
- `TASK-261016-w0yx2w_review-probes-rev1.go`: supplemental executable attack file. Reproduce in an archive of the pinned candidate at `internal/envsnapshot/reviewer_retry_test.go`, then `go test -count=1 -race -run '^TestReviewer(PublicationRetry|StageRootBeforeRename)$' -v ./internal/envsnapshot`.
- `TASK-261016-w0yx2w_review-sweep-rev1.md`: surface sweep and free-hunt notes.
- Local runtime: darwin/arm64, Go 1.25.5. Linux/macOS hosted suite qualification and Windows refusal are bounded separately by terminal hosted evidence.

The initial supplemental test-file write used a wrong scratch-relative path. Its subsequent no-tests invocation is excluded from evidence; the corrected run explicitly executed both named tests. No setup failure is counted as a negative-control kill.

## Terminal hosted rerun

Reviewer command `sh scripts/remote-gate.sh` exited **0**. [Run 39082985905](https://github.com/relux-works/curator/actions/runs/39082985905) completed successfully: **20 successful jobs, 2 conditional skips**. Every job and step was inspected. Skips are not passes. Signed detached snapshot `c42a6707ce7001003db3951c851e79eb7f2ab434` has signature G and exactly the candidate tree. The gate completed and removed its temporary branch; no validation process remains running.

Downloaded and parsed all five complete test streams. Snapshot test/subtest passes: Linux test/race 254/254; macOS test/race 252/252; Windows 104. All have zero snapshot failures/skips and a terminal package pass. Windows TestDestinationUnqualifiedPlatform passed; this qualifies refusal, not Windows publication. Linux race has one explicitly classified dependency-download banner; all remaining lines parse as JSON.

The reviewer independently reran focused tests, narrowing controls, supplemental probes, routing tests and the full hosted gate. The older producer green is corroborating evidence only, not a substitute for this rerun. Linux snapshot mutant execution is not claimed; the 16 snapshot narrowing controls were executed locally on macOS.

## Verdict and logbook

**ACCEPT** revision 1, accepted but not landed. No blocking finding reproduced. The four surface rows are held and the free hunt is empty of findings. Conditional rejection-routing checklist item is not applicable because this verdict accepts; no `done` or `commit_ack` transition is requested.

```json
{
  "verdict": "accepted",
  "revision": 1,
  "candidate_tree": "2e20b225f2794c9e524ee69332ed156424c612d8",
  "findings": [],
  "notes": [
    "Library-only acceptance: CLI integration and complete coordinator inventory remain separate leaves.",
    "Windows is qualified for explicit refusal only.",
    "The brief has no catalog-keyed surface table or explicit review budgets; all four supplied paired rows were swept.",
    "Same-user mutation inside private staging is outside the normative guarantee. Supplemental root-substitution and retry attacks found no blocking defect."
  ],
  "surfaces": [
    {"row": "admission-protected-identity", "result": "held"},
    {"row": "replacement-competition", "result": "held"},
    {"row": "prior-identity", "result": "held"},
    {"row": "staging-cleanup-verification", "result": "held"}
  ],
  "free_hunt": [],
  "hosted_run": "https://github.com/relux-works/curator/actions/runs/39082985905"
}
```
