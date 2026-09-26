# TASK-260923-em42lw — rework 3: refresh onto trunk aa093918 (THE ONLY CURRENT INSTRUCTION)

Revision 3 (pre-0018 vector accommodation, on base 48da2690 / SPEC_PIN rc.12) is withdrawn: TASK-260922-18ex37 landed and moved SPEC_PIN
to curator-spec dcc7f015, whose conformance/v1 already carries decision 0018 (`environments.permissions`). Trunk is now `aa093918` (many
landings since 48da2690, incl. 0017 cww1ov). Safety ref: refs/campaign/em42lw-rev3-delta-20260926.
1. `task-board m 'set_status(TASK-260923-em42lw, status=development)'`.
2. The old workspace (base 48da2690) could not be converged (conflict in internal/config/environments*) and was ABORTED (work preserved). Your
   new workspace starts on trunk. Re-apply your revision-3 delta: `git diff 48da2690 refs/campaign/em42lw-full-20260926 -- . ':!.task-board'
   ':!CHANGELOG.md' ':!LOGBOOK.md' | git apply --3way` (the ref is a full snapshot incl. previously untracked files — 37 paths — in the shared repository); resolve every conflict keeping trunk's content
   (0017, 18ex37 pin, etc.) plus your change; list each. Leave nothing staged. VERIFY `git diff --name-only HEAD` lists only this task's paths.
3. With the pinned root now at dcc7f015: if TestManagerConfigV2Vectors passes WITHOUT the pre-0018 accommodation, REMOVE the accommodation
   (and its rows); otherwise keep it scoped exactly as rework 2 required and say why. launch-env-fragment-v2 emission per the brief is
   unchanged; the lattice rows and mutants must still pass/kill.
4. (no refresh-candidate needed: the workspace is already on trunk).
5. Bounded runs: `go test ./internal/config/... ./internal/envprofile ./cmd/curator -run 'ManagerConfigV2|Fragment|EnvResolve|Permissions' -count=1`
   (+ -race on internal/config), go vet, `GOOS=windows go vet ./internal/config ./cmd/curator`. Real exit codes.
6. CHANGELOG POLICY: no CHANGELOG edit (entry text in results); no LOGBOOK edit. Append "Revision 4 — refresh onto aa093918", `resource update`,
   `task-board handoff TASK-260923-em42lw --role developer`; stay in the turn while the gate runs. A write-boundary `policy warn` block is a warning.
