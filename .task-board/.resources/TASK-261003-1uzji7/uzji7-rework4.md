# THE ONLY CURRENT INSTRUCTION — TASK-261003-1uzji7 rework 4: make the v2 writer flip green on the hosted gate (developer, code; rc.4)

The orchestrator converged your Story workspace onto fresh main 77fabd45. Your rev3 changes are carried uncommitted. Snapshot: refs/campaign/uzji7-snapshot-20261005. Do NOT `git checkout` or reset anything in the workspace.

The rev3 hosted gate (run 37240734108) is RED: 46 failing tests on ubuntu and more on macOS. The extracted failure messages are in the attached resource `uzji7-rev3-failures.txt`. Read ALL of it first. The failures fall into four families:

A. **"install marker is invalid for schema 5"** (install/status/scopes/GC/crossconformance). After the flip, the writer emits install marker schema 5 (hash_version 2). Several readers and validators still reject schema 5 or never admit it: the scopes collector (GC sweep, every marker-schema row), status, the compiled-state readers, the draft replay and substitution paths. Fix the readers so they accept every schema the spec admits. Do NOT weaken validation for anything else.

B. **Legacy-lane marker schema pins** (TestStatusReportFindsASchema8InstallationCurrent: skill schema 7 wants marker 3, skill schema 8 wants marker 4; TestLegacyBuildsKeepReceipt1; TestLegacyMixedBuildProjectInstallProducesMarkerV3; TestGlobalMixedBuildStagesExternalBeforeLocal). Decide from curator-spec (rc.14 or later on spec main: protocol marker schemas and the hash-version sections) whether the v2 flip changes the marker schema for each skill-schema lane, or only the hash version inside the lane's schema. If the spec keeps per-lane schemas, the writer must keep them (with hash_version 2 where the schema carries it). Change a test pin ONLY if the spec says so, and cite the § in your notes.

C. **Environment-marker schema 1/2 parsing**: "versions 1 and 2 do not carry hash_version" breaks parsing of real schema-1 fixtures (TestEnvResolveKeepsSchema1BytesForMetadataOnly, TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome). Separately, the credential-record tests expect a schema-2 record but get Version:3 (TestEnvResolveCredentialRecord*). Legacy env markers must keep parsing byte-for-byte. Only newly written records move. Check against the spec's environment-marker versioning.

D. **Status currentness for legacy markers**: legacy (v1-hash) markers now classify as content-drift instead of current or needs-install (TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands, TestDraftStatusLegacyMarkerNeedsInstall, TestClassifyDraftMember…/legacy-marker, TestCLIEndToEndInstallStatusAndTamperCheck, TestCuratorStatusProviderPostureAndCheck, TestGlobalStatus*). Readers must compute the hash with the marker's OWN hash_version (v1 for legacy markers) and never compare a v1 marker against a v2 recomputation.

Rules:
- Rehash, never relabel (the original brief and 1foyf3 evidence still bind).
- Keep exact conformance counts.
- Do not edit LOGBOOK.md, CHANGELOG.md or scripts/remote-gate.sh.
- Host rules R193/R194 (attached host-rules.md) bind. cmd/curator, internal/install and the other fake-executable suites MUST NOT run locally; the hosted gate is the arbiter. You MAY run targeted tests for internal/marker, internal/envmarker, internal/hashing and internal/scopes ONLY through `~/.local/bin/mini-build-lock run uzji7 -- env GOFLAGS=-work go test ./internal/<pkg> -run '<regex>' -count=1 -timeout=6m`. If a sandbox prevents that, list the commands and do not run them.
- In the validation resource, write a table: family → root cause → fix (file:line) → the failing tests it should turn green → verified locally or hosted-pending.

Then `task-board handoff TASK-261003-1uzji7 --role developer` and END YOUR TURN.
