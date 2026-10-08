# Reviewer instruction — TASK-260927-25hk87 revision 3 (posture-B re-applied after converge)
Cross-provider review: the producer was muse max; the reviewer is codex sol high (R195: this flips the shipped security-posture default to hardened).
Revision 2 was ACCEPTED on 2026-10-05 and went stale only because main rewrote `cmd/curator/env_credential_marker_test.go`. Verify the re-application, not the design again:
1. Semantic diff: revision 3 against revision 2 differs only by the rebase; the test intent in env_credential_marker_test.go is re-applied without weakening any of main's new v1/v2 NUL-gate assertions.
2. The hardened default still holds at the production entry (security_posture conformance cases driven, gap rows gone); use the hosted gate evidence for cmd/curator.
3. No stray edits; no LOGBOOK/CHANGELOG/remote-gate.sh changes.
accept_cr, or request changes with numbered findings. Do not edit files.
