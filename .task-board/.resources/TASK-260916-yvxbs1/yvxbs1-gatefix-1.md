# TASK-260916-yvxbs1 — gate fix (THE ONLY CURRENT INSTRUCTION, with yvxbs1-sec-brief.md and yvxbs1-decision-1.md)

Revision 1 (tree b61dd4e9, base eca2bf27) FAILED the hosted gate on every lane (run 36320324273) — you handed off a red revision again.
Orchestrator diagnosis from the go-test.json artifacts (76 top-level failures):
1. ~45 of them: `environment_store_untrusted: path "/tmp/acme-source" fails regular_types boundary check at "/tmp/acme-source": lstat
   /tmp/acme-source: no such file or directory` during `provision pi|codex_cli|claude_code` and repair (TestResolveProvisionRepair,
   TestResolveDriftRepair, TestResolvePassthroughLiveness, TestCredentialLink*, TestStatus*, TestOpencodeXDGSeeds, …). Your boundary check
   runs at provision/materialisation against the lock's recorded `source_path`, which is NOT read there (materialisation reads the store).
   Per rc.13 the protected-boundary validation of a path overlay applies where the path source is READ (profile install/update/resolve of a
   `path` package, and the store it produces) — cite the clause. Move the check to those read points; a recorded source_path that no longer
   exists must not make provisioning of an already-locked profile untrusted (if the spec says otherwise, cite it and update the fixtures).
2. Ledger consistency FAILED: `.github/ci/platform-cases.tsv` requires internal/envprofile TestSymlinkInPathIsSourceInvalid on linux/darwin
   but it is no longer compiled (renamed/deleted/build tag). Restore it or update the ledger row with the new name (the rename must keep the
   same assertion).
Run `go test ./internal/envprofile` split by -run groups (Resolve|Status|Credential|Seed|Path|Boundary|Overlay) and `go test ./cmd/curator
-run 'Profile|Path|Status'` with real exit codes, plus `scripts/check-ledger-consistency*` (or the gate's ledger step) locally.
Set status development; append "Revision 2 — boundary check at read points; ledger"; handoff and WAIT for the gate; read the validation log;
hand off only green. No CHANGELOG/LOGBOOK edit.
