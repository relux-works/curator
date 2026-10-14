# BUG-261014-351fcr — reviewer B verdict, CR revision 1: ACCEPTED

Candidate tree 3524e8ad, base 84cf54b9. Source SHA-256 of both changed files in the worktree equals the digests in the gaps file. The verdict was written before reading gw-BF1-review-A.md.

No surface table was supplied with the brief. The rows below are the brief's five checks, each walked and given a result.

## Rows

| Row | Result | Attack run |
| --- | --- | --- |
| 1. Both call sites under hashing.WriteVersion(), same selection as global add --revision, no second rule, no path exemption | held | The diff is two lines: the direct DetectAtVersion call and Resolved.HashVersion = hashing.WriteVersion(). Grep shows the only remaining VersionV1 uses in envprofile.go are the unrelated waiver path (line 628) and the zero-value default in resolvedHashVersion (line 1680). contextresolve.go:902 uses the same hashing.WriteVersion(). No path-based branch exists. |
| 2. Migrated lock records hash version 2 under v2 | held | The lock is built in ensureDefault with HashVersion zero, and contextlock.lockVersions derives schema 2 / hash 2 from WriteVersion. The test reads the published lock through readLock and asserts schema 2 / hash 2, Validate(), member commit and source, and stored bytes. A mutant that reverts the lock to v1 is covered by that assertion and by the v1 cases. |
| 3. Tests fail on regression | held | I ran three mutants against TestMigratedGlobalSkillBinaryWriteVersion, each restored byte-for-byte afterwards. (A) Reverting only the direct DetectAtVersion call to V1 fails the test (exit 1). (B) Dropping Resolved.HashVersion fails the v2 subtests of the gzip-and-binary, gzip-only and binary-only cases in both directories. (C) Setting Resolved.HashVersion to an explicit V1 fails the same v2 subtests. The forced-v1 cases expect profile_source_invalid with opaque-file in the fixture directory and no published profile state. The gzip-only and binary-only cases also catch a file-type exemption. Directories .task-board/attachments and assets/deep rule out a path exemption. This is narrowing evidence, not delete-only: the NUL gate stays present in every mutant. |
| 4. network-undeclared follow-up | held | TestMigratedGlobalSkillNetworkUndeclaredFollowUp drives EnsureDefaultWithPolicy and reproduces the finding for scripts/fetch.sh. The gaps file records the exact error text and asks for a separate leaf. No fix was introduced. |
| 5. Hygiene | held | The test and gaps file contain no local paths or private identifiers. The test uses temp homes, local bare repos and the example.com identity rewritten to a local remote, and does not touch ~/.curator. The remote-gate result is cited: hosted run 38517800551 on snapshot f34a0db8, success, with the skipped lanes listed as skipped. |

AC coverage: 4 of 4 rows driven through EnsureDefaultWithPolicy -> ensureDefault -> migrateGlobalSkills.

## Re-run
go build ./..., go vet ./internal/envprofile/ and go test ./internal/envprofile/ -run TestMigratedGlobalSkill: all pass. I did not rerun the full suite locally (R223); I accepted the attached hosted result for that. I did not independently check the hosted run or the log digest.

## Findings
[] — no blocking findings.

## Notes
- The fixture reaches the migration by first parsing a branch declaration, then pinning its commit as a revision. migrateGlobalSkills still skips live branch declarations, as before; this is out of scope for the leaf.
- The network-undeclared refusal (strictAuditMember passes capabilities.ImplicitNone()) is confirmed and needs a separate leaf. It is not a defect of this candidate.
- Free hunt: I found nothing beyond the table. I did not attack concurrent writers or the case-sensitive lanes.
