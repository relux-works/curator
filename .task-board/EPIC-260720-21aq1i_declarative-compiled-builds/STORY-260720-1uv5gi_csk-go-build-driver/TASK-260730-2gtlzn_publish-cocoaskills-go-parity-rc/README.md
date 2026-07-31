# TASK-260730-2gtlzn: publish-cocoaskills-go-parity-rc

## Description
After all 17 core Go parity delivery tasks are independently accepted and landed to ivanopcode/cocoaskills main, verify the exact remote main commit, create the first non-conflicting v0.13.0 release-candidate tag, publish the GitHub prerelease, and verify repository and release assets/state.

## Scope
CocoaSkills origin git@github.com:ivanopcode/cocoaskills.git only. No wb remote operations. Publication starts only after the 17 delivery-task gate; separate interop/platform release-grade work does not block this RC.

## Acceptance Criteria
All 17 Go parity delivery tasks are done and present on origin/main; the selected RC tag points exactly to verified origin/main; GitHub prerelease is published and observable; no other remote is modified.
