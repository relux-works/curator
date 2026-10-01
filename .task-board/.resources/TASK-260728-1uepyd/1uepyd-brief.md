# TASK-260728-1uepyd — drive the rc.13 external-repository-acquisition vector (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, this task's README, and the audit `.research/260930_compiled-build-leaves-reconciliation.md` (row
1uepyd). The admission code is on main: internal/buildrepo/admission.go ExactSSHCommand / ValidateGitTool / AcquireNetwork,
local.go AdmitLocal. The gap is that curator-spec rc.13 (SPEC_PIN) publishes conformance/v1/vectors/external-repository-acquisition.json
(12 cases, 55 common_fetch_argv rows, 17 clean_environment rows, 11 forbidden_fetch_features rows) and curator does not consume it.
1. Add a conformance consumer that drives EVERY row of that vector through the production entry (AcquireNetwork / AdmitLocal, or the
   narrowest production seam that builds the fetch argv and environment). Compare the exact argv, the clean environment and the
   forbidden-feature refusals against the vector. Record coverage with the repo's conformancecoverage helper (driven / known-gap /
   bound / skipped), and add the exact pinned count row to .github/ci/conformance-case-counts.tsv.
2. Where curator does not match the vector, first decide whether it is a curator bug. If it is, fix it in production code. If the
   vector is wrong, stop and report it in the results; do not bend the test. Anything left must be an owned known-gap row with a
   precise reason. Nothing may skip silently.
3. Review the 3 existing bounds (local-config-and-refs 2, pack-index 1). Drive them if possible, otherwise restate each reason
   precisely.
4. Mutants, with real exit codes:
   - one extra fetch flag allowed;
   - one clean-environment variable leaked;
   - one forbidden feature accepted.

   Each must be killed.
5. New state reads go through internal/stateread. No Windows-reserved file names. No CHANGELOG/LOGBOOK: put the entry text in the
   results under "## CHANGELOG entry (for release prep)". Never spell any employer name.
Update the results with the per-table coverage, then run `task-board handoff TASK-260728-1uepyd --role developer`, then END YOUR TURN.
