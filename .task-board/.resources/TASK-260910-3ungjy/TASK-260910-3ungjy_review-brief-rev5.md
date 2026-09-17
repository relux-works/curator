# Review brief — TASK-260910-3ungjy (S6 approval commands and status posture, curator manager), review round 3 (Change Request revision 5)

You are the independent reviewer of a curator implementation produced for
`TASK-260910-3ungjy` (story `STORY-260910-2awkzu`, second and final leaf; the
first leaf `TASK-260910-1952mz` is checkpointed on the Story branch at
`d489ae0`). Read, in this order: `remediation-manager-producer-rules.md`, the
producer brief `TASK-260910-3ungjy_brief.md`, the orchestrator's gate analysis
`TASK-260910-3ungjy_gate-failure-rev1.md`, the producer results
`TASK-260910-3ungjy_results.md`, the published Change Request patch
`TASK-260910-3ungjy_change-request_rev5.patch` (revision 1 failed the hosted
gate; revision 2 passed on all lanes — see its validation log), and the
landed spec `profiles/manager.md` §8.2/§8.3/§8.4/§8.6/§8.7 in the read-only
curator-spec checkout `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`.

## Where the candidate is
The managed Story worktree `<control-root>/.temp/STORY-260910-2awkzu/worktree`
on branch `task-board/story/STORY-260910-2awkzu` holds the exact candidate
(checkpoint `d489ae0` + the uncommitted revision-2 delta). Do not edit it and
leave NO files in it (the rev-6 review of the sibling left a `.review/` copy
inside the worktree and broke the checkpoint digest — use a disposable copy
under `/tmp` for builds, tests and mutants).

## What to verify
1. **§8.3 commands**: `curator hook approve <path>` resolves the operand to
   the same canonical identity the hooks and `hookapproval` use (absolute,
   symlinks resolved, Windows spelling folded), fails without recording when
   the file is absent or unreadable — two distinct facts, never conflated —
   records `{ path, sha256, approved_by: "operator", approved_at }` of the
   current bytes, re-records after a change, publishes atomically through
   the existing helper; `curator hook approvals` lists read-only and never
   rewrites or "repairs" a malformed record; `curator hook revoke <path>`
   removes the record and leaves state byte-identical when there is none,
   reporting so. Quote file:line.
2. **§8.6 posture**: `curator status` and `curator env status` list each
   known project env file with state approved / unapproved
   (`shell_hook_env_unapproved`) / changed (`shell_hook_env_changed`), the
   absolute path and `approved_by` for recorded files; `--check` treats a
   changed file as non-current and an unapproved file as a warning row. The
   `status --json` document gained the additive `shell_hook_trust` key: the
   legacy-shape pin was updated deliberately (still proves no `builds` key;
   asserts exactly the four keys and the row shape) and CHANGELOG names the
   addition — confirm it was not merely loosened.
3. **Unreadable-state tests on every platform**: the rev-1 Windows failure
   (a regular file at the manager-home path reads as absence on Windows) is
   fixed by making unreadability real on all platforms (directory at the
   state file path or a lock), not by a GOOS skip; §8.4 "unreadable is never
   absence" holds at every new read site.
4. **Independent validation** from a disposable copy with
   `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`:
   `go build ./... && go vet ./... && gofmt -l internal cmd` and the narrow
   packages (`./cmd/curator/... ./internal/hookapproval/... ./internal/shell/...
   ./internal/envfiles/...`), `set -o pipefail`, exit codes quoted; the
   sibling's §8.7 vector test still passes; an end-to-end check: approve →
   generated hook sources silently; revoke → warns again (run the emitted
   hook under `sh`/`bash`, and `pwsh` if present); at least two narrowing
   mutants (e.g. approve records without reading bytes; revoke deletes the
   whole state file) caught by committed tests.
5. **Scope and hygiene**: no change to the hook text or trust decision logic
   of the first leaf (report if touched), no `SPEC_PIN`/vendored spec, no
   unrelated edits, shipped profile stays `A-warning`, CLI docs updated, no
   writes outside the worktree.

## Verdict
Record `TASK-260910-3ungjy_review-verdict-rev5.md` (task outcome) with the
per-item table, transcripts, mutants and findings; then exactly one of
`task-board m 'accept_cr(TASK-260910-3ungjy, revision=5, evidence=TASK-260910-3ungjy_review-verdict-rev5.md)'`
or a changes-requested verdict routed with `set_status(TASK-260910-3ungjy, status=to-dev)`
listing the concrete corrections. Never accept on the producer's evidence
alone; never edit the candidate.


## Round 3 specifics (revision 5 — the combination review)
Revision 4 was ACCEPTED (`TASK-260910-3ungjy_review-verdict-rev4.md`) but its
integration was refused with `integration_base_moved`: trunk advanced with
`STORY-260910-3vxe3y` (`6645b9b`, package-lock-and-frozen-resolution), which
changed `cmd/curator/main.go` — a path this candidate also changes — and the
board demoted the revision to stale because "no one has looked at the
combination". Revision 5 is that combination: the Story workspace was
refreshed onto fresh trunk (`worktree refresh-candidate`), the rev-4 work
re-applied, and republished; its hosted gate is green on all lanes.
`TASK-260910-3ungjy_results.md` ("Revision 5") names the exact points in
`cmd/curator/main.go` where trunk's package-lock hunks sit next to this
task's `hook` subcommand and status rows.

Verify FIRST the combination: `git log -1 6645b9b -- cmd/curator/main.go`
and the current file — trunk's hunks intact, this task's hunks intact, the
subcommand dispatch and status-row wiring coherent (no shadowed cases, no
duplicated registration, no lost flag), the package-lock tests of
`STORY-260910-3vxe3y` and this task's tests both green in one run. Then the
rev-4→rev-5 diff must be nothing but that re-anchoring (compare the two
change-request patches modulo context). The R1–R3 posture closures accepted
in round 2 still hold (spot-check the three production-entry tests). Do not
leave files in the worktree.
