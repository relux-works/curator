# TASK-260923-em42lw — merge fix against 2elcdc (THE ONLY CURRENT INSTRUCTION)

Revision 8 fixed Windows but fails on macOS + ubuntu (Test and Race): internal/config TestSystemV2Refusals/isolated_direction
(environments_test.go:421) "err = <nil>, want mention of \"only toward shared\"". That refusal row is STALE: TASK-260923-2elcdc (landed on main,
isolated environment lock) made system-config-v2 environments.isolation ADMIT `isolated` (environments §12.2, manager §1 rule 1). Your merge
kept the pre-2elcdc test.
1. `git diff origin/main -- internal/config` over your working tree: for every hunk that is NOT your 0018 permissions / fragment-v2 work,
   take trunk's bytes (2elcdc) — in particular the isolation refusal/admission rows and any config parsing 2elcdc changed. List each file
   you reconciled. Your own changes stay.
2. Bounded runs: `go test ./internal/config -count=1` (+ -race), `go test ./cmd/curator -run 'EnvResolve|Permissions|Isolation' -count=1`.
3. `task-board m 'set_status(TASK-260923-em42lw, status=development)'` first; `git fetch origin main` and combine any newer trunk (keep both
   sides); refresh-candidate if needed. Append "Revision 9 — reconcile with 2elcdc", `resource update`, `task-board handoff TASK-260923-em42lw
   --role developer`; stay in the turn. If the loop detector refuses, stop and report. No CHANGELOG/LOGBOOK edit.
