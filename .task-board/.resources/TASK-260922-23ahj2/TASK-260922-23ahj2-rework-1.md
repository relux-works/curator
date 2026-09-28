# TASK-260922-23ahj2 rework 1 (orchestrator, binding)

Verdict rev1: CHANGES_REQUESTED (TASK-260922-23ahj2_review-verdict-rev1.md). Ownership fix,
F-M1/F-L1/F-A1, headless set and the no-D5-violation statement are accepted. Three sentences from
the binding addendum are missing — insert them VERBATIM where the verdict names the lines, no other
change; continue from the revision-1 tree (no checkout/clean/stash); republish on a green gate.

1. C1 — `decisions/0018-curator-run-permission-interface.md` (~350–361) and results.md F-M1/F-L1:
   add "agents-management owns the LaunchRequest permission-mode member for LaunchModeInteractive
   with positive and negative interactive goldens per tool release; the module's argvguard forbids
   spelling argv grammar in two places, which is why F-M1 owns the mapping and F-L1 carries no argv
   grammar." Mirror it in the F-M1 AC. Add to 0018 Compatibility and results: "0017/0018 block
   nothing in task-board: tracked children spell bypass in the module's exec plugins themselves;
   0018 serves curator run now and tracked sessions later through ax (F-A1)."
2. C2 — `decisions/0017-environment-credential-modes.md` (rationale ~85, choice ~166),
   `protocol/environments.md` §7.4 (~1486), results.md F-C1/F-C3: add "On the operator Mac,
   ~/.curator/environments/default/pi/auth.json points to ~/.pi/auth.json; that target does not
   exist, the real credential is ~/.pi/agent/auth.json, and env status does not show the dangling
   passthrough as detached." Extend F-C1: "env status and resolve report dangling or mis-targeted
   credential links as detached, with environment_credential_conflict-class wording instead of
   silence." Extend F-C3: "The production-entry suite covers a recorded link whose native target
   does not exist, with a narrowing mutant proving this bound." Keep F-C2 (bytes-preserved
   migration) as is.
3. C3 — `decisions/0017…` (~167), `protocol/environments.md` (~1467), `profiles/manager.md`
   (~2746): add "The operator's machines have no cli_auth_credentials_store key in config.toml, so
   the platform default file applies and isolated is admitted; the operator's profile design relies
   on that platform default and on nothing more from 0017." Keep the keyring/auto refusal.
Append "Revision 2" to results.md listing the three insertions with file:line.
