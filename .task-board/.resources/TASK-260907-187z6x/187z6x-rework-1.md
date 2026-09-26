# TASK-260907-187z6x rework 1 → revision 2 (orchestrator brief, binding)

Verdict: CHANGES_REQUESTED on revision 1 (`TASK-260907-187z6x_review-verdict-rev1.md`): one medium
finding plus an evidence correction. The reviewer VERIFIED the rest — parity is the right reading of
environments §9.1/§9.2/§9.5 and cli/curator.md (no reinstall carve-out), the path implementation and
activation helpers are byte-unchanged, takeover safety is a §8.3.1 write, and both mutants (yours
and theirs) behave. Do not re-open any of that.

## R1 (medium) — a partial activation reports "installed", not "updated"
`cmd/curator/profile.go:128-133` prints `installed profile …` unconditionally when activation
returns an error, ignoring the returned `updated` flag. Row 2 of your suite
(`cmd/curator/profile_git_reinstall_test.go:134`, `not-current/use without takeover`) discards
stdout and so never noticed: on the exact candidate it prints
`installed profile groot (lock sha256:…)` where §9.1's reinstall reporting rule — and your own
results claim that "every case is reported as an update" — require `updated profile groot (…)`.
The reviewer's reproduction is a Go overlay that adds `updatedLine(t, stdout)` to row 2; it exits 1
on the candidate.

Fix both halves:
1. make the reporting honour the update result when activation fails partially (the operand was a
   reinstall: the lock/source record WAS updated, only the switch did not complete) — keep the exit
   code and the diagnostics exactly as they are, this is a wording defect, not a behaviour change;
2. give row 2 the same operator-line assertion the seven success rows have, so the shape is pinned;
   add a narrowing mutant (restore the unconditional `installed`) that the row kills.

Check the path-root sibling still reports identically (the reviewer's bound: path behaviour must not
change).

## R2 — evidence correction (no code change)
Your results describe row 2 as "nothing written, no backup" without qualification. The actual stdout
shows `codex_cli`, `opencode` and `pi` DID switch — which is exactly what pinned environments §9.2
requires (attempt every adapter, report the partial). Rewrite that row's description to say what
really happens: the blocked `claude_code` home is the one that is neither written nor backed up,
the other adapters switch, and the command reports the partial with
`environment_surface_unmanaged_conflict` + `profile_use_partial` and exit 1.

## Rules
Continue from the revision-1 tree (no checkout/clean/stash). Re-run the eight-row suite plus the
path sibling with real exit codes, add the mutant row, append a "Revision 2" section to
`TASK-260907-187z6x_results.md`, then `task-board handoff TASK-260907-187z6x --role developer`.
Publish only on a green gate. Note: trunk has moved to 48da2690 (the rc.12 conformance pin, SPEC_PIN
= dced9b8); if the base refresh brings rows that previously skipped into execution, name them.
