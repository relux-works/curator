# TASK-260927-4pv4au — gate fix (THE ONLY CURRENT INSTRUCTION, with 4pv4au-brief.md)

Revision 1 (tree 9058518b) fails 3 tests on every lane (run 36344923451, go-test.json):
- cmd/curator TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands (status_test.go:1786): `status --json changed the historical
  document shape`.
- internal/scriptworker TestProductionBinaryLaunchesWhenHostProvides (derive_test.go:1154): `cannot decode the stub report "warning:
  security_posture_permissive: …"`.
- internal/install TestEnforcedInstallAndLaunchAtCLIEntry (scriptpolicy_test.go:712): `the installed launcher did not run the script:
  invalid character 'w' … warning: security_posture_permissive`.
The permissive warning leaks into machine-readable output and into the launch path. Fix per rc.13 (cite where the warning is emitted and
which "operation" means — profiles/manager.md §7.1, security-posture.json diagnostics/status_gates): warnings go to stderr only, never into
stdout of `--json` documents, script/launcher/worker protocol streams or the output of a launched command; the launch/exec path of a script
or tool does not emit it unless the spec names that surface. The posture row belongs where the spec puts it (env status); do not change the
legacy `status --json` document shape (add fields only if the spec's status_gates require it there, and then update the legacy-shape test
with the spec citation). Run the 3 tests plus `go test ./internal/config ./internal/envprofile -run 'Posture|Security|Hardened|Permissive'`
with real exit codes. Update the results resource (append "Revision 2 — warning surfaces") before handoff; WAIT for the gate; hand off only
green. No CHANGELOG/LOGBOOK edit.
