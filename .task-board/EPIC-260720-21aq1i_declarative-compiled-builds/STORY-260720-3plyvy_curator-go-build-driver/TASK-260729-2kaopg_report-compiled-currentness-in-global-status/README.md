# TASK-260729-2kaopg: report-compiled-currentness-in-global-status

## Description
Make `curator global status` report compiled-command currentness, and decide the --json/--check parity contract for the global scope.

## Scope
Own the cmd/curator global status branch and its CLI tests. TASK-260720-1nlmvv deliberately excluded this surface: its listed CLI surfaces were install, upgrade, dry-run, status, status --json, status --check, and gc, and the exclusion is now documented in README. Global install and upgrade do plan and commit compiled commands, so a compiled global installation currently has no currentness surface at all. Reuse the stable currentness vocabulary and classification in cmd/curator/builds.go rather than inventing a second one. Decide explicitly whether global status gains --json and --check, and whether it may run a read-only global plan (which brings audit and registry read-only gates into a command that today only reads markers).

## Acceptance Criteria
curator global status reports one diagnostic per active compiled command using the same stable codes as curator status; the chosen --json/--check contract is documented in README; an unchanged compiled global installation reports current and a drifted one does not; the pre-existing declared-skill output stays compatible for a global scope without compiled commands.
