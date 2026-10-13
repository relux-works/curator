# Wave 6 decisions (owner, 2026-10-13)

**Owner decision, verbatim:** «окей, давай принимаем все 44». It accepts all cycle-3 architecture-review findings, F-29 … F-72, which the owner received with the implementation plan.

**What this wave applies.** It covers the findings that block the first stages (P0, S0a, and W, the workspace lifecycle), plus text drift against decisions already taken. Each finding lands as its minimal fix, before its first consumer (implementation plan §13). The findings for later stages land with those stages' contract packages, not now:
- S0b, the one-host MVP;
- S2, enforced networking;
- T, board delivery;
- A, the first donor workplace;
- H, the session-host services.

**Optional owner policies stay unchosen and off:**
- standing discard of refused exports;
- the local-reviewer and requester-exclusion rules;
- metadata confidentiality hardening.

**Bounds that apply to every lane:**
- The watcher freeze lifecycle is v1. This wave publishes the v0 scope honestly and lists the v1 items as open.
- The alert route gets an explicit status-only phase for v0. Containment never waits for delivery.
- The two ALLSTOP findings stay distinct acceptance cases.
- `workspace-result/1` keeps its field set.
- No implementation, qualification or activation is claimed anywhere.

**The records are private review material.** Public repositories cite module sections and shared rule names only, with no F-numbers, record slugs, seat names or event ids. Never remove identifiers mechanically: a pattern such as `F-<digits>` also matches inside `UTF-8`. Simply add none, and leave every other passage byte-identical.
