# Rework 3 — TASK-260910-14hsti (local source and output boundaries)

Verdict rev4: CHANGES_REQUESTED (resource TASK-260910-14hsti_review-verdict-rev4.md). Orchestrator scope ruling, binding for this cycle:

Finding 2 (per-write recheck) — fully in scope: the source contract's acceptance says "recheck at publication". One Recheck before journalPlan is not that. Carry the boundary snapshot into the transaction plan / engine so the physical-identity comparison runs immediately before EACH publication write (and rollback of unpublished state on refusal); add the real-journal fault-injection regression that swaps a parent after planning and before a later write. No stubJournal-only proof.

Finding 1 (production data flow) — in scope for every production caller that exists today:
- adapter destination planning: production adapter groups (internal/install/install.go around :752) must populate Admitted, and `scopeTargets.boundaries` must be populated on the real install path, so the publication guard is live for actual installs — test through the public production entry (the real stageTargets, not an injected stub).
- local acquisition: wire `PrepareLocalAcquisition` into every existing production path that acquires a local path source (the current path-source snapshot capture used by profile compose/path overlays), gated by the draft/opt-in switch as the wave note requires; the legacy path stays byte-identical when the switch is off (test that).
- OUT of scope for this leaf: the draft package-snapshot store and its acquisition entry point belong to TASK-260910-16k7xy (capture-and-store-local-package-snapshots). If a call site only exists there, do not stub it; list the exact function and file in results.md under "Bounds" so 16k7xy wires it. Do not describe in-scope wiring as "deferred".

Keep the Windows eager identity pin and the same-spelling regressions. Narrow tests only (internal/snapshot, staging, privatedir, adapters, install — use -run filters; the host stalls under whole-package parallel runs, so run install tests with -p 1). Evidence with exit codes, mutants for the two findings, checklist, handoff rev5 from the Story worktree.
