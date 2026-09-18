# TASK-260918-11f9l1: combine-rc12-union-with-moved-trunk

## Description
Replay the accepted, checkpointed rc.12 pin-promotion union (TASK-260917-16l2md revision 5, checkpoint 73fc8a4 on base 3c45d4b) onto the moved curator trunk (S6 squash 64cacfc: shell-hook trust rows in envstatus.go / status.go / main.go / CHANGELOG; d00fe7a: rose-air rustup lane) via task-board worktree refresh-candidate with checkpoint-bound replay resolutions, combine the status-row output order so both the S6 and the E2/E4/S4 tests pass, and hand off the combined tree as the story-final Change Request.

## Scope
(define task scope)

## Acceptance Criteria
1. refresh-candidate replays checkpoint 73fc8a4 onto the fresh trunk authority with explicit resolutions for CHANGELOG.md, cmd/curator/envstatus.go and internal/envprofile/status.go (union of both sides; nothing dropped). 2. The combined tree keeps every accepted rev-5 hunk of the union and every S6 hunk byte-identical except the three resolved regions; row order in curator status / env status stated and pinned by tests; all S6 and E2/E4/S4 tests green at the rc.12 root. 3. SPEC_PIN stays dced9b8; no build outputs in the candidate. 4. Results record the replay transcript, the per-file identity proof and the gate transcripts.
