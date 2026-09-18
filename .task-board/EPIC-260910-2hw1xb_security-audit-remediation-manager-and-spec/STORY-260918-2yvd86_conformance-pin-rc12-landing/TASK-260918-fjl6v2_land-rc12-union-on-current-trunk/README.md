# TASK-260918-fjl6v2: land-rc12-union-on-current-trunk

## Description
Apply the accepted rc.12 pin-promotion union (TASK-260917-16l2md revision 5, checkpoint 73fc8a4 vs base 3c45d4b, 40 files) onto the current curator trunk in a fresh Story workspace, using the three union resolutions and the test-only fix-up prepared and validated by TASK-260918-11f9l1 (resources), state and pin the closed status-row order, run the full suite at the rc.12 root, and hand off as the story-final Change Request.

## Scope
(define task scope)

## Acceptance Criteria
1. The working candidate = trunk + the 40-file union delta with per-file patch-id identical to 73fc8a4-vs-3c45d4b for every file except CHANGELOG.md, cmd/curator/envstatus.go and internal/envprofile/status.go (byte-identical to the TASK-260918-11f9l1 resolved files, themselves the union of both sides) and cmd/curator/main.go (auto-merge = checkpoint + S6 hunks); plus TASK-260918-11f9l1_fixup.patch (two S6 test files, +33 lines, no production code). 2. SPEC_PIN dced9b8 unchanged; no other change. 3. go build/vet/gofmt clean and the full suite green at the rc.12 root locally; hosted gate green at handoff. 4. Results carry the per-file identity proof, the closed output order of curator status / env status, and the gate transcripts; rule 8 hygiene.
