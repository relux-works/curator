# Review B brief: BUG-261017-47xdyt, install must not delete or replace content it does not own

Decide accept or reject. Read the bug, `gw-OWNPRES-brief.md`, the triage, and the spec clauses it cites.

1. Reproduce item A yourself in a private scratch HOME on the candidate: unowned siblings are named in a refusal and stay byte-identical; foreign native links stay valid.
2. Item C2: an unowned selected destination refuses by name with no partial change; owned updates and absent installs still work.
3. Ownership comes from preceding manager state and markers, not names or file presence; unreadable or invalid evidence refuses.
4. Mixed owned and unowned case proves owned stale entries are still removed (no reject-everything).
5. Global and hybrid callers keep their contracts; transaction and GC invariants hold.
6. Mutants fail for the intended reason; full gate green; public hygiene holds.

Reject with numbered concrete fixes; accept only when all six hold.
