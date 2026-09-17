# CHANGES_REQUESTED — TASK-260910-16k7xy revision 4

Candidate tree: 36b334603d7e02edce04d87938b0b00661e00dd4.
Base: 37e38d95a3f5bf1849f0e3d08dc2bf40217437f7.
Checkpoint HEAD: a4181f9 (accepted 14hsti boundary work).
All 31 candidate changed paths were compared byte-for-byte with the working files: 31/31 match. Working delta is precisely the six capture-leaf paths. No repository code, index, branch or commits changed by this review. Experiments use Go overlays in ignored .temp/review-16k7xy only.

## Findings and required rework

1. P1 — internal/snapshot/capture.go:246-250 fails to revalidate file identities. The contract section 3 explicitly requires identities as well as bytes and membership. BuildInventory retains only path/hash/execute, so replacing a file after copying with a different inode containing the same bytes/mode passes. A deterministic scheduling hook after the copy loop atomically replaces a.md with an identical-byte new regular file; Capture returns nil error. TestReviewIdentityReplacementDuringCapture fails. Retain and compare physical file identities across capture, with safe reads and refusal on replacement. Add a deterministic production-entry regression for same-byte replacement; do not merely compare digests.

2. P1 — internal/snapshot/capture.go:315-338 rereads staging after authenticating it, then publishes the new copy without hashing that copy. A deterministic hook immediately after the pre-publication inventory verification changes a.md from before to after. PublishLocal returns success and leaves a target addressed by the old digest; OpenLocal immediately refuses it as source_snapshot_changed. This creates corrupt content-addressed store state even though the envprofile caller subsequently refuses installation. Verify the actual copied publication tree before rename, fail mutation with source_snapshot_changed, and leave no published target on refusal. Add a regression for mutation between validation and the publication copy, not only mutation before PublishLocal starts.

3. P2 — internal/envprofile/draft_capture_test.go:42-57 does not prove consumption of the captured tree. Mutation happens after Install returns. A narrowing mutant retaining capture/publication but assigning the frozen source only when stored == "" makes the normal install use dir; all TestDraftInstall tests still pass (exit 0). Move the test mutation to after capture and before contextstore consumption, drive Install, assert captured bytes and kill that mutant. Likewise TestCaptureDetectsLiveContentChange compares two separate explicit captures, not a race within capture. Narrowing the post-copy live digest comparison to file-count comparison survives the entire Capture|Vector|Publish|Open shard. Add a real concurrent-content-change refusal at the acquisition/Install boundary. Producer's original evidence labels two complete clause deletions as narrowing mutants; those are not narrowing proofs.

These are ordinary implementation/test rework, not external blockers. No re-opening of accepted 14hsti decisions is requested.

## Independent commands and observed outcomes

Shell: zsh; commands run directly, no pipes masking exit codes. Every started process completed before verdict.

- go test -p 1 ./internal/snapshot -run 'Capture|Vector|Publish|Open' -count=1: exit 0, 1.564s.
- go test -p 1 ./internal/envprofile -run 'Draft|Path' -count=1: exit 0, 76.580s (includes legacy path coverage selected by this mask).
- go vet ./internal/snapshot ./internal/envprofile: exit 0.
- golangci-lint run ./internal/snapshot/...: exit 0, "0 issues."
- gofmt -l on all six leaf paths: exit 0, no output.
- go test -p 1 -overlay .temp/review-16k7xy/probe.json ./internal/snapshot -run TestReview -count=1: exit 1; 2/2 deterministic race refusal probes fail.
  - capture accepted a replaced file identity during capture.
  - PublishLocal accepted mutated copy; subsequent OpenLocal=source_snapshot_changed: stored snapshot does not match its recorded bytes.
- go test -p 1 -overlay .temp/review-16k7xy/mutant-capture.json ./internal/snapshot -run 'Capture|Vector|Publish|Open' -count=1: exit 0, 1.745s; mutant survives.
- go test -p 1 -overlay .temp/review-16k7xy/mutant-install.json ./internal/envprofile -run DraftInstall -count=1: exit 0, 11.508s; mutant survives.

Narrowing mutant score: killed 0/2; surviving 2/2. This measures these two attacks only, not all refusal clauses. Instrumented race probes only inject deterministic file mutations at existing scheduling boundaries; they do not remove validation. Source copies and overlay mappings are attached in TASK-260910-16k7xy_review-probes-rev4.tar.gz. To use elsewhere, extract and regenerate Replace absolute paths from the mapping basenames under the new checkout. No restore was needed: candidate files were never mutated.

## Accepted evidence and bounds

- Production wiring inspected: Install/stateForPath -> admitPathSource -> PrepareLocalAcquisition -> Capture -> PublishLocal -> OpenLocal -> contextstore.EnsureState. Overlay declarations use stateForPath as well. Draft flag preserves legacy branch.
- All 3/3 normative vectors match the accepted conformance snapshot-cases.json and pass locally on macOS. Dirty/untracked Git, no synthesized commit, absent/tampered snapshot refusals pass the snapshot shard.
- Windows skips are gated on GOOS == windows and begin with the approved platform-control sentence. Windows normative executable vectors remain explicitly deferred under the latest review brief, not claimed passing.
- Reused hosted evidence: TASK-260910-16k7xy_change-request_rev4-validation.log records run 35160776161 success, exit 0. Gate commit a6e3b9f9cbc5e7f0d4d6f56a0f8e3a9c5f38b357 resolves locally to exactly candidate tree 36b334603d7e02edce04d87938b0b00661e00dd4. No full suite rerun. Ubuntu/macOS/Windows jobs reported successful; rose-air and candidate suite reported skipped. These are attached CI results, not independently executed platform checks.
- Runtime/build vector hashes are covered; end-to-end runtime/build/dependency consumption beyond the inspected envprofile wiring is not established by this review. The constant-only TestVectorsFreezeContextAcrossRuntimeAndBuildEdits does not add execution coverage.

## Lifecycle

spawn goal returned no active goal (run not goal-bound). No pending directives. Verdict: changes_requested, route to to-dev. Acceptance, architecture and complete-test-coverage checklist claims remain unapproved. Review findings recorded in task notes and this outcome; repository LOGBOOK.md untouched per campaign prohibition.
