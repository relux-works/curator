# TASK-260928-q5100t — spec: global add/install lock publication ordering (THE ONLY CURRENT INSTRUCTION)

curator-spec repository. Operator decision 2026-09-28 on the TASK-260906-1xbrz6 decision packet
(`TASK-260906-1xbrz6_global-recovery-decision-packet.md`): ALTERNATIVE 1. Keep the five-operation takeover carrier set closed (no new flag
on global add/install). State normatively in protocol/environments.md §9.4 (and wherever the manager transaction rules for global
operations live — cite them; mirror the publish-before-rematerialize order §9.2 already gives `profile update`, steps 4–5):
- `global add` and `global install` publish the extended profile lock before in-place materialization;
- when a surface write meets `environment_surface_unmanaged_conflict`, the published extended lock is KEPT (not rolled back), so the §9.4/
  §9.5 recovery — `profile sync --takeover` or `profile use --takeover`, then retry — materializes the updated skill set;
- the operation reports the conflict naming the surface; nothing else changes.
Add conformance vectors (a new case in the existing global/takeover vector family, or a small new family with manifest/schema wiring as the
repo's tools require) for: lock published before surfaces; lock preserved on conflict; sync --takeover after the conflict materializes it.
Run the repo validators (`make` targets / tools/validate.py) with real exit codes. CHANGELOG entry under Unreleased (this repo keeps its
CHANGELOG). No LOGBOOK.md. Update the results resource, `task-board handoff TASK-260928-q5100t --role developer`, END YOUR TURN (the
runner publishes and gates). Write only inside your Story worktree.
