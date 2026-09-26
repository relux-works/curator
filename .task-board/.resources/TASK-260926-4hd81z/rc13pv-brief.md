# TASK-260926-4hd81z — accept the rc.13 suite protocol version (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Blocking curator-spec v1.0.0-rc.13 (PR #97): spec Implementations run
36216044011 fails only `internal/scriptpolicy TestScriptExecutionPolicyIdentityMatchesTheSuite`: "the pinned suite protocol_version =
\"1.0.0-rc.13\", want the rc.12-pinned script-worker-v1 protocol 1.0.0-rc.9".
1. Change the check to a closed set {1.0.0-rc.9, 1.0.0-rc.13} (versions whose script-worker-v1 identity is unchanged) with a comment
   saying why; everything else it binds stays.
2. Run `go test ./internal/scriptpolicy -count=1` against curator's pinned suite (default) AND with CURATOR_CONFORMANCE_ROOT pointing at
   conformance/v1 of curator-spec PR #97 head f6bd748c (fetch refs/heads/release/v1.0.0-rc.13 into $TMPDIR) — both exit 0.
3. Mutant: accept any version → a crafted rc.14 label must fail (kill it; real exit codes).
No CHANGELOG/LOGBOOK edit (entry text in results). Attach results, check DoD, `task-board handoff TASK-260926-4hd81z --role developer`. A write-boundary
`policy warn` block is a warning.
