# TASK-260916-hxr6qv revision 4 review verdict

ACCEPT — exact candidate 895ec9c2fa05ff20a3f77f5530f2970ba767989f, base 62ea2d2ced3fac7c3c5f7a2ff4e887e0f92f4540. Reviewer RUN-260917-b53b5d. No source edits or commits.

Independently compared changed files to candidate blobs: 26/26 equal. git diff --check exit 0. Read repository-transport sections 6–7, rework requirements, previous verdict and revision-4 implementation/evidence. The Linux correction at cmd/curator/draft_transport_provenance_test.go:195–255 preserves the real non-dry-run production install and sink. Unsupported build hosts must return the exact inventory refusal after acquisition; five actual sink records and two successful mirror records remain mandatory. Supported hosts require successful installation and six records. Canonical identity, listed/resolved mirror fields and secret checks still execute. No new blocking defect established.

Hosted verification: independently queried gh run view 35200643034 (exit 0): success, head 27c52ba6575b945390a8ff55f758778fd488e39f. Independently resolved that commit tree to the exact candidate above. Attached revision-4 validation log ends exit 0: eleven jobs green including Ubuntu test/race, macOS test/race, Windows test and lint. Candidate suite and rose-air skipped. Full gate not rerun.
https://github.com/relux-works/curator/actions/runs/35200643034

Independent reviewer evidence reused from predecessor RUN-260917-90ba1c, whose verdict could not be persisted due to host startup stalls: exact-candidate production sink tests passed, exit 0, 72.735s; dry-run-only overlay failed at the intended nil-sink assertion, exit 1, 21.861s. Read its preserved report at .temp/review-hxr6qv-r4/TASK-260916-hxr6qv_review-verdict-rev4.md and dryonly.log. Measured independent narrowing attack: 1/1 killed. This is independent reviewer evidence, not producer-only acceptance.

Current-run reruns: production cmd mask, install/buildrepo/config revision-2 and compatibility mask, removed-sink and dry-run-only overlays were launched with -p 1 -count=1 -timeout=90s. They stalled without results; terminated, not counted green or as mutation kills. sample of curator.test showed 12 KiB footprint at _dyld_start before Go runtime startup. Reattempts using internal linking are separately reported below. No host settings, signing metadata, installed binaries or product sources changed. Existing alternate task-board binary with --no-update-check works and is used solely to persist this review.

Reused producer evidence: revision-3 attempt-cap narrowing plus mirror/alias failure-class mutants (3/3 killed; failure classes helper/loader level); revision-4 removed-sink and dry-run-only mutants (2/2 killed). Earlier review independently passed revision-2/refusal and schema-1 compatibility checks; no product delta in this test-only rework. These do not establish universal mutation coverage.

Bounds: Windows skips the POSIX production fixture. Linux exercises acquisition and its diagnostics before expected build refusal, not successful build artifacts. Portable-file scans have a 1 MiB bound and do not assert every artifact family exists; not a universal absence proof. No rose-air result claimed. Host-stalled checks are unverified, not assertion failures. The requested Linux gate regression is verified by the exact-tree hosted run.

All 13 board checklist items checked on read. Goal query: not goal-bound. No directives. Acceptance uses accept_cr revision=4 only, no commit_ack or done transition; integration remains producer-owned. Host anomaly is recorded here and in board notes; campaign prohibits LOGBOOK.md edits.

Internal-linking retries (buildrepo narrow shard and removed-sink cmd overlay) also produced no assertion output and were terminated with descendants; neither counted as passing. All this reviewer’s verification processes stopped before lifecycle completion.
