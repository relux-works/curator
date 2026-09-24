# TASK-260916-1xib1x — revision 5 independent review

Verdict: ACCEPTED. No outstanding finding in the requested revision-2 → revision-5 scope.

Candidate tree: abbbb066aec995be5932674d1140a2d80fef7965. Base: d275e2d9d23afdde75abecdc51971b663047ffd5. Previous reviewed tree: f684c415e35e7c9ebd716cf23c124bbbbf2bcc92. The delta is exactly 36 added lines in internal/audit/audit.go and 204 added lines in internal/audit/auditlabels_test.go; no deletions or changes to earlier label, CLI, installation, or conformance rows. Product working files were not modified. Tests and mutants ran only in a disposable candidate archive under the assigned worktree.

## Findings and verification

The writable cache-hit path now calls backfillScriptPolicies with cached findings and current subject policies. A missing/null script_policies member is backfilled; already-recorded policies remain untouched, preserving commands-less source-audit behavior. Cache versions/identity and decision computation are unchanged. Backfill serializes the same typed findings, without re-running detectors. Formatting can change, but findings values do not.

Independent Darwin amd64 / Go 1.26.0 / zsh command:

`go test -count=1 ./internal/audit`

Exit 0, 11.696s, after all mutations were restored and audit.go compared byte-for-byte with the candidate. This runs the named committed TestReviewerExistingVerdictGetsPolicyRecord, the original reviewer resource copied under TestReviewerOriginalExistingVerdictGetsPolicyRecord, all existing audit rows, and the new reviewer TestReviewerCurrentManifestBackfill.

The new attack uses old-format verdicts under a changed current command map, in both clean and non-empty cached-finding cases. Production GateReadOnly leaves verdict bytes unchanged; the report and ScriptPolicyRecords expose current policies. Production Gate then persists exactly current → explicit absence and guarded → script-worker-v1, drops the old command identity, preserves cached sentinel findings semantically, and returns identical warnings/errors to the read-only call. The sentinel finding does not exist in the snapshot, so recomputing detectors would fail the assertion. The existing mixed and no-warning rows also pass.

An initial baseline invocation overlapped the first disposable mutant and failed on missing backfill. It is discarded as contaminated, not claimed as a candidate failure or passing evidence. The final baseline above ran sequentially after mutation completion and byte restoration.

## Mutation evidence

3/3 attacks killed, 0 survivors; each uses go test -count=1 and returns exit 1 with a behavioral assertion failure, not a compile failure:

- Skip the cache-hit backfill call: both the committed named regression and original reviewer regression fail because script_policies remains absent.
- Narrow backfill to len(findings)==0: TestReviewerCurrentManifestBackfill/cached-finding fails. This proves non-empty cached findings are covered.
- Allow backfill regardless of persist: committed read-only regression and both new reviewer cases fail on changed stored bytes.

The exact output is attached as TASK-260916-1xib1x_rev5-mutants.log; reviewer source is attached separately. All production bytes restored. Prior revision-1/2 vector and five-mutant acceptance is retained per binding review note; those expensive CLI/global rows were not rerun in this scoped review. New coverage ratio is 2/2 changed-manifest cache shapes and 3/3 mutation attacks, not a full runtime qualification claim.

## Landing evidence reused

Independently queried https://github.com/relux-works/curator/actions/runs/35717146238: success. Head d1c6c63c6b2ded23a3691af592241558dc58ef57 resolves locally to abbbb066aec995be5932674d1140a2d80fef7965, exactly the candidate. Ubuntu/macOS/Windows tests, three gate self-tests, Ubuntu/macOS race, lint, naming and interop succeeded. Candidate suite and rose-air skipped; no rose-air/ARM64 claim. This is hosted job-status evidence, not independently replayed raw artifacts. Full landing suite was not rerun. git diff --check for the scoped delta passed.

Bounds: persisted non-null policy records are deliberately not refreshed; this review attacks old-format records as requested. Findings-cache identity is content-based, and normal manifest edits change that identity. Cache write failures retain the pre-existing best-effort behavior; no stronger durability claim. Platform-ledger registration remains R5 scope.

## Lifecycle / logbook note

2026-09-22: the revision-2 upgrade-cache finding is resolved. Read-only behavior and cached-findings preservation independently passed. No logbook executable is available; campaign rules prohibit LOGBOOK.md edits, so this attached task-scoped verdict records the result. Run goal query reports not goal-bound, and no directives are pending. Acceptance routes revision 5 to integrating; it is not a landing or done claim.
