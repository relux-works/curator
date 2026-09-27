# BUG-260916-3aco9f: profile-install-reinstall-drops-use-and-takeover

## Description
https://github.com/relux-works/curator/issues/73 — same defect as TASK-260907-187z6x (git same-source reinstall delegates to updateLocked and drops --use/--takeover; path reinstall was fixed by stage (c) rework 5). Sequencing 2026-09-22: 187z6x (running in STORY-1a2i5a) delivers the git-shape fix with eight run()-driven rows; this bug then covers the residual: verify every install shape incl. path reinstall with an unchanged lock under --use, add the stop-then-retry conformance vector (spec leaf if the vector family lives in curator-spec), and close issue #73 with the landed commits.

## Scope
internal/envprofile/envprofile.go installLocked / reinstallPathLocked and the git reinstall delegation; conformance vector for stop-then-retry

## Acceptance Criteria
--use and --takeover are honoured on every install shape including a reinstall of a recorded source; an unchanged lock still switches under --use; the retry after a takeover stop performs the takeover with notice and backup exactly like profile use --takeover; a conformance vector covers the stop-then-retry sequence; issue #73 closed with the landed commit
