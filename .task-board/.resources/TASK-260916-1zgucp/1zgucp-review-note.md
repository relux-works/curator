# Review note — TASK-260916-1zgucp E1 source signers + system delta confirmation (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest revision (rev3, gate green) against `1zgucp-sec-brief.md`, `1zgucp-gatefix-1.md` and curator-spec v1.0.0-rc.13 (E1:
source_signers allowlist, require_source_signers, system-delta confirmation; cite clauses). Verify through the production entry:
1. Signer allowlist enforced during resolution (contextresolve/contextlock): unsigned/unknown signer refused under enforcement, posture
   reported in env status; the missing-state-pin refusal happens BEFORE any network/source access (fail-closed, no I/O).
2. Profile update prints the resolved-version delta and refuses without --confirm-system-delta when the delta touches class: system modules
   or MCP declarations.
3. The stateread seam used for profiledelta.go:deltaMemberRoot (guard test passes); signer-posture test fixture sets a git identity.
4. rc.13 E1 vectors (122 cases per the goal list) driven; E1-owned gap rows removed (before/after histogram).
5. Kill at least two mutants (signer check removed / confirmation skipped). No revert of trunk; no CHANGELOG/LOGBOOK; no stray files.
Also: the producer's earlier stray write into the control root was cleaned by the orchestrator — confirm nothing of it is missing from the
candidate. Focused bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
