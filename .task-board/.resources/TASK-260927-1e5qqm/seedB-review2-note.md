# Reviewer instruction — TASK-260927-1e5qqm revision 2 (seed-B re-applied after converge)
Cross-provider review: the producer was muse max; the reviewer is codex sol high (R195: this flips a shipped default).
Revision 1 was ACCEPTED on 2026-10-05; revisions 2 and 3 are re-applications after main moved (the v2 writer flip, then posture-B). Verify the re-application, not the design again:
1. Semantic diff: the current revision against revision 1 differs only by the rebases. The same 36 gap rows are removed, nothing else in the TSV changed; the test intents in env_credential_marker_test.go and envstatus_test.go are re-applied without weakening any of main's new v1/v2 NUL-gate assertions.
2. The Codex seed revision B behaviour (B vectors driven, A homes unstripped under B, servers stripped, empty-table no-warning) holds at the production entry; use the hosted gate evidence for cmd/curator.
3. No stray edits; no LOGBOOK/CHANGELOG/remote-gate.sh changes.
accept_cr, or request changes with numbered findings. Do not edit files.
