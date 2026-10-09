# THE ONLY CURRENT INSTRUCTION — TASK-260924-4mzun5 rework 4 (developer, loop breaker under tb-R226)

You are the strong producer (gpt-6-astra, max effort) brought in after three rework rounds. Revision 4 was reviewed: **changes_requested** with ONE blocking finding (F4, P1). Read `TASK-260924-4mzun5_review-verdict-rev4.md` first, then the reviewer probes `TASK-260924-4mzun5_review-install-probes-rev4.go`, `TASK-260924-4mzun5_review-cli-probes-rev4.go` and `TASK-260924-4mzun5_review-harness-rev4.txt`. Everything else in the candidate was found sound; keep it.

The Story workspace already holds the whole candidate (57 paths) on trunk 22cf8b4f. Do not restart from scratch. The orchestrator converged it onto 22cf8b4f after the N9 landing: CHANGELOG.md (both entries kept) and internal/audit/audit.go were three-way merged cleanly with N9's change (N9 records pin creation time in audit). Check that the merged audit.go builds (`go build ./...`) and that your hosted green run covers it.

## Fix F4 — an unrelated tag defeats strict same-tag movement refusal
At `internal/install/install.go` (moved-tag detection, around lines 1557–1565) the check infers continuity of the installed declaration from the presence of ANY live tag on the old package commit; an unrelated `archive` tag (lightweight or annotated) suppresses same-tag movement detection under StrictTags on project and global entries.
- Bind and recover the exact previous declared selection and ref from validated installed-generation evidence tied to the marker's lock digest, and distinguish movement of that same tag from an explicit declaration change. Do not infer a previous declaration from the presence or absence of arbitrary live tags.
- Preserve the closed v5/v6 marker carriers and the schema-1 re-resolution semantics; no new marker members unless the spec already defines them.
- Keep the existing changed-declaration admission and no-alias refusal controls.
- Add the reviewer's alias probes permanently at both production entries (project and global; lightweight and annotated).
- Add a narrowing mutant that waives same-tag refusal only when an unrelated tag still binds the prior commit; the new tests must kill it.
- The reviewer's separate P2 note is filed as BUG-261009-2s6t3y. Do not widen scope for it, but if your fix also makes the reviewer probe `TestReviewR4ChangedTagAfterOldTagDeleted` pass on both entries, keep it as a permanent regression and say so in your results.

## Evidence — hosted only
Never run `go test`, compiled test binaries or `go run` on this host (refused; they harm it). Compile-only checks (`go vet`, `go build`, `gofmt`) are fine.
From a disposable `git clone --shared` of the control root, push scratch branches and wait for CI (push all of them at once so the runs overlap; this run has a 120-minute limit) (`gh run watch <id> --exit-status`):
1. **green:** the exact candidate → full CI green;
2. **red:** the candidate with only the F4 production fix reverted → the alias regressions fail (name them);
3. **narrowing mutant** (waive same-tag refusal when an unrelated tag binds the prior commit) → the new tests fail (name them).
Record the run URLs and outcomes in your results resource, tick the checklist items with that evidence, delete the scratch branches.

Then `task-board handoff TASK-260924-4mzun5 --role developer` and END YOUR TURN. Do not edit LOGBOOK.md.
