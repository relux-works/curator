# STORY-260928-3gzr9m: nofollow-parent-walk-guard-row

## Description
E5 residual found in the 3ed9m3 rev2 review: mutants on internal/envprofile/nofollow.go (M1 Lstat->Stat in managedPath parent walk, line ~43; M2 disabling the symlink refusal at ~60/~111) SURVIVE on trunk d8e87bac against TestEnvironmentWriteNofollowVectors and TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing — later layers (nofollow open helpers, atomic writers) hold the rows. The parent-walk guard has no row that kills it on its own.

## Scope
(define story scope)

## Acceptance Criteria
rows that kill M1 and M2 individually (a planted parent-directory symlink on a managed write path refused with environment_write_would_follow_link before any later layer runs); hosted gate green
