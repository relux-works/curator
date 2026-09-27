# TASK-260916-1zgucp — E1 gate fix (THE ONLY CURRENT INSTRUCTION, with 1zgucp-sec-brief.md)

Revisions 1–2 fail the hosted gate (latest run in the rev2 validation log) on three concrete things:
1. internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded (state_read_guard_test.go:221): "internal/envprofile/profiledelta.go:deltaMemberRoot
   tests not-exist after os.Stat without a seam route or reviewed allowlist reason". Route that read through the internal/stateread seam
   (absent vs unreadable, environments §8.4.1) — do NOT allow-list it unless it genuinely cannot be unreadable, with a reason.
2. internal/envprofile TestUpdatePathWithoutStatePinIsSourceInvalid: now attempts a real fetch ("repository 'https://example.com/pk/' not
   found") instead of refusing with profile_source_invalid naming the missing state pin. Your signer-allowlist enforcement runs BEFORE the
   state-pin validation; restore the order so the missing-pin refusal happens before any network/source access (fail-closed, no I/O).
3. internal/envprofile TestSignerPostureVectorsAtStatus/enforced-names-verified-signer: "fatal: empty ident name … not allowed" — the test
   fixture runs git commit without an identity on CI runners. Set user.name/user.email (or GIT_AUTHOR_*/GIT_COMMITTER_* env) in the fixture.
4. Trunk moved to `97ca3370` (3aco9f, 2r1upt, em42lw, 20o9dk, rc.2 …): `git fetch origin main`, combine (keep both sides; any trunk-added reader
   onto the seam), `task-board worktree refresh-candidate TASK-260916-1zgucp` if publication would refuse; VERIFY no trunk revert.
5. `task-board m 'set_status(TASK-260916-1zgucp, status=development)'` first; bounded runs of internal/envprofile (split) + -race on the
   three tests; append "Revision 3 — gate fix", `resource update`, handoff; stay in the turn. If the loop detector refuses, stop and report.
No CHANGELOG/LOGBOOK edit.
