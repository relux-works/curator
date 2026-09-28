# TASK-260916-yvxbs1 — gate fix 2 (THE ONLY CURRENT INSTRUCTION, with yvxbs1-sec-brief.md, yvxbs1-decision-1.md)

Good progress: revision 2 (tree 1df50e2f) is down to TWO failures on every lane (run 36328835395, go-test.json):
- internal/envprofile TestUpdateDefaultIsBlocked (envprofile_test.go:494): got `profile_not_found: manager_state_absent: …/profiles/default/
  source.json: no installed profile "default"`, want `profile_update_blocked`.
- cmd/curator TestProfileListMigrationHonoursSystemPolicy/update-default (profile_test.go:697): same.
Your change moved the source.json read (via stateread) ahead of the `default`-profile special case in the update path. Restore the order:
the `default` profile's update refusal (`profile_update_blocked`) is decided before any source.json read, exactly as on trunk; the new
path-source boundary validation runs only for profiles that have a path source. Change nothing else. Run those two tests plus
`go test ./internal/envprofile -run 'Update|Path|Boundary|Overlay'` and `go test ./cmd/curator -run 'Profile'` with real exit codes.
Set status development; append "Revision 3 — default update ordering restored"; handoff and WAIT for the gate; hand off only green.
No CHANGELOG/LOGBOOK edit.
