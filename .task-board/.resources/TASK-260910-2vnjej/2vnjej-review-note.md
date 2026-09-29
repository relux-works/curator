# Review note — TASK-260910-2vnjej S2 cross-registry root check (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev4 (base 213a53e5, tree 2742e187, 26 paths, gate green on every lane incl. Windows) against `2vnjej-sec-brief.md`,
`2vnjej-gatefix-1.md`, `2vnjej-gatefix-2.md` and curator-spec rc.13 registry.md (bootstrap trust from a signed checkpoint;
registry_bootstrap_tofu posture; cross-registry / mirror-group root comparison and equivocation; cite clauses) and the registry-client vectors.
Verify through the install path: (1) enabled registries in the same mirror group at the same log size compare accepted signed roots;
divergence → the specified refusal/diagnostic; advisory cases per spec; (2) bootstrap checkpoint: first use requires a valid signature,
later checkpoints preserve/advance the high-water mark (lower refused, equal-conflicting refused, higher advances); a present-but-unverifiable
checkpoint fails closed (never "absent") — reads via internal/stateread; (3) Windows: the checkpoint private-file check uses the owner-only
DACL helper (not mode bits) and fixtures create checkpoints through the product helper; the CRLF row; (4) vectors driven, Story gap rows
removed; mutants (root comparison skipped; high-water not enforced; unverifiable treated as absent; DACL check bypassed) killed with real exit
codes. No trunk revert, no CHANGELOG/LOGBOOK, no stray files. Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
