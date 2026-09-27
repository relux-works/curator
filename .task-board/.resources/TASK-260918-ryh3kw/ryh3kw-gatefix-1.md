# TASK-260918-ryh3kw — gate fix (THE ONLY CURRENT INSTRUCTION, with ryh3kw-sec-brief.md and ryh3kw-decision-1.md)

Your handoff said the Windows fix was "locally validated" but the published revision 2 (tree 3cfcc758 — identical to your worktree) FAILED the
hosted gate: `TASK-260918-ryh3kw_change-request_rev2-validation.log`, run 36296139672, Test (windows-latest) — ~30 required cases fail across
cmd/curator, internal/install, internal/install/atomicity, internal/transaction, internal/interop/environments, internal/shell, internal/adapters
(e.g. TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable/missing_operand, TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault/
absent_config, TestDryRunTouchesNothing, TestEndToEndInstall, TestConformanceEnvironments*). Pattern: on Windows, genuinely ABSENT paths are now
classified unreadable (fail closed where absence is legal). Read the full failing logs (`gh run view 36296139672 --log-failed` from the
worktree) and fix the classification so Windows absence (ERROR_FILE_NOT_FOUND / ERROR_PATH_NOT_FOUND, and a missing parent) is absence, while
real read failures stay unreadable. If a unix-only distinction (ENOTDIR component) cannot be made on Windows, state the platform bound and keep
the row unix-only with a reason — never break absence.
Also: trunk moved to 38c68570 (TASK-260927-1wc76r restore-backups landed) — `git fetch origin main`, combine keeping both sides; the 2 restore
vectors you parked as gaps for 1wc76r must now be driven or remain attributed to what actually blocks them. VERIFY no trunk revert
(`git diff --name-only origin/main` shows only your paths). `task-board m 'set_status(TASK-260918-ryh3kw, status=development)'` first; bounded
local runs; the HOSTED gate result on the published revision is the only proof — do not claim green without it. Append "Revision 3 — Windows
absence classification", `resource update`, handoff; stay in the turn. If the loop detector refuses, stop and report. No CHANGELOG/LOGBOOK edit.
