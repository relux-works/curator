# TASK-261008-eaqg8a — N7 legacy forwarding link review, revision 1

Verdict: accepted. No blocking findings or implementation rework.

Reviewed CR-TASK-261008-eaqg8a-1 revision 1 against report section N7 and its four-state probe. Base: 3d395ffe72ec979e2ef1d3d792655a7f405785cc. Candidate tree: 5633d0b614ba54996d460984008699222a3be5d7. Reviewed all four changed paths; repository code remains unchanged by this reviewer.

## Swept surfaces (6/6)

| Surface | Finding and evidence |
| --- | --- |
| Production wiring | install.Global stages through globalbins.StageForwarding, merges its plan, and journalPlan maps staging.KindEntry to transaction.KindEntry with DigestTarget preimages. The permanent TestGlobalReconcilesOwnedLegacyForwardingLink drives Global, including the Refresh-generated legacy fixture. |
| Ownership admission | Desired shims still require a ledger entry and ownedTarget via unmanagedConflict. Stale removals still require the ledger and ownedTarget. Only a recognized live symlink receives ReplaceEntry/RemoveEntry; regular targets retain byte handling. |
| Exact link backup/rollback | DigestTarget hashes the raw Readlink destination. Transaction backup renames the entry and rechecks its digest; rollback renames the backup back. Existing exact-link commit/rollback, removal, recovery, and strict-byte-kind tests pass in the hosted artifact. No general nofollow guard was weakened. |
| Foreign-file and prepared-preimage controls | Both report refusals pass in the candidate and production-revert mutant. Foreign bytes are preserved; the post-prepare drift requires a failed Global result. Owned regular forwarding remains green. |
| Cross-platform regression and gate | Candidate full hosted matrix is green, including lint, Test on Linux/macOS/Windows and Race on Linux/macOS. Unix legacy-link subtest is explicitly skipped on Windows using an admitted platform-control reason. |
| Operator-facing documentation | Unreleased/Fixed describes recognized legacy-link reconciliation transactionally and retained foreign/preimage refusals accurately. |

## Independently verified hosted evidence

- Candidate green: https://github.com/relux-works/curator/actions/runs/37868830965
  Commit 08fca25d1e4f4325563bc6917c52dfbfae196ec8; GitHub commit API reports exactly the candidate tree above.
- Handoff validation green: https://github.com/relux-works/curator/actions/runs/37873617879
  Commit 6470446441e7d862fe1923fda7c42bacc477d3ac; GitHub commit API also reports exactly the same candidate tree.
- Production-revert mutant red: https://github.com/relux-works/curator/actions/runs/37868866996
  Commit fb403640cd1c6031088ff8cd0cbc093c5370a94b; tree 29dd1bf318d531af19f359c271c9650996e3cb2b.

Downloaded and parsed the green/red test-evidence-ubuntu-latest go-test-served.json artifacts myself. Candidate: N7 states 4/4 pass, zero failing named tests. Mutant: owned-legacy-link fails with Global status failed; owned-regular, foreign-regular, foreign-after-prepare pass (3/3 controls). The only failing leaf is TestGlobalReconcilesOwnedLegacyForwardingLink/owned-legacy-link, plus its parent test.

Both artifacts also report pass for TestEntryKindDigestsLinksAndBytesWhileByteKindStaysStrict, TestCommitAndRollbackRestoreALinkExactly (4/4 rollback leaves), TestEntryRemovalRestoresTheExactLink, and TestRecoveryFinishesAPreparedLinkTransaction.

Mutant reconstructed independently in a disposable shared clone: retain candidate test and CHANGELOG, restore ONLY internal/globalbins/globalbins.go and stage.go from the CR base. Its computed tree exactly equals the hosted red tree (24 removed production lines). Compile-only go vet on internal/globalbins and internal/install passed in that clone. Hosted suite results above are accepted from the producer's exact snapshots and independently inspected; I did not rerun hosted suites or execute any local go test.

## Checks rerun by this reviewer

- Candidate: env -u GOROOT GOMAXPROCS=2 go vet ./... — exit 0.
- Candidate: env -u GOROOT GOMAXPROCS=2 go build ./... — exit 0.
- Changed Go files: gofmt -l — empty; git diff --check — exit 0.
- Candidate file contents match the CR tree, including the untracked test blob.
- Fresh origin HEAD advertises main at the CR base; exact-ref fetch in the disposable clone matches that OID. No upstream overlap or convergence needed.
- Task gate scratch branches are deleted.
- spawn goal queried before verdict: run is not goal-bound.

Bounds: production rollback of this specific Global legacy-link fixture and stale-removal race schedules are not newly exercised by the four-state regression. Exact-link rollback/recovery is supported by unchanged transaction tests and reviewed entry-kind wiring. Windows legacy symlinks are outside the Unix probe's dynamic claim. No exhaustive filesystem race or crash-schedule claim.

Checklist review items are satisfied by the evidence above. Conditional nonacceptance checklist item is not applicable to this accepted branch; this task-scoped verdict records the branch. LOGBOOK remains untouched per the task brief. Acceptance routes to integrating, not done; landing belongs to the tracked producer integration run.
