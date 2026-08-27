# Prepare CocoaSkills Go parity release candidate

## Description
After all 17 Go parity delivery tasks are independently accepted and landed to ivanopcode/cocoaskills main, prepare a release-candidate package and publication plan without creating a git tag or GitHub Release until the user gives an explicit release command.

## Scope
CocoaSkills origin git@github.com:ivanopcode/cocoaskills.git only. Verify exact origin/main, choose the first non-conflicting v0.13.0 RC version, draft release notes and asset/checksum inventory, and record the exact publication commands and rollback checks. Read-only remote inspection is allowed. Do not create or push tags, publish a GitHub Release, modify another remote, or claim publication.

## Acceptance Criteria
All 17 Go parity delivery tasks are accepted and present on exact origin/main; the proposed RC version is absent remotely; release notes, asset/checksum inventory, exact target SHA, publication checklist and post-publication verification commands are attached as reviewed evidence. Repository state is unchanged by preparation: no tag or GitHub Release is created.
