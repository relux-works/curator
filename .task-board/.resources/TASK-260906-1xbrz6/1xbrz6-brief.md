# TASK-260906-1xbrz6 — takeover closed set: import and global operations (curator-spec)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator-spec; work only in your assigned
Story worktree `.temp/STORY-260905-2z9pw4/worktree`. Read `campaign-producer-rules.md` first. Gate = the
board's validation command (spec-gate.sh = make validate steps; the runtime runs it once at handoff).

Context: `protocol/environments.md` §9.5 ("Takeover is not an operation of its own…") admits the takeover
flag on exactly five mutating operations (`profile install`, `profile use`, `profile sync`, `profile
update`, `env resolve --repair`). Review cycle 2 of TASK-260906-1hn93j observed two operations outside
that closed set that can also meet unmanaged files: (1) `profile import` (§9.6) — the import writes
nothing into a native home by itself, but its first-install ACTIVATION follows §9.1 and materializes
in-place surfaces; (2) §9.4 `global add` / `global install` materialize in place under the current
profile. Neither carries the flag. An operator blocked by `environment_surface_unmanaged_conflict` on
either can escape by running `profile sync --takeover` or `profile use --takeover` first and retrying.
The stage (c) implementation has ALREADY consumed the five-member enumeration (curator main), so this
leaf resolves the question by TEXT, not by widening.

Deliverable (orchestrator decision: keep the closed set): add to §9.5, next to the enumeration, one
normative sentence stating that `profile import` activation and the §9.4 global operations are outside
the set by design, that on `environment_surface_unmanaged_conflict` they fail closed exactly as §8.3
states, and that the recovery is the carrying operations named (`profile sync --takeover` / `profile use
--takeover`, then retry); mirror the same statement where §9.6 describes activation and where §9.4
describes in-place materialization (one sentence each, cross-referencing §9.5); make sure
`cli/curator.md` import/global rows and `profiles/manager.md` agree (no new flag on any row). Add a
schema/vector case only if an existing family enumerates the takeover-carrying operations (check
conformance for such a list; if it exists, the case must show import/global are not members). If, while
doing this, you find a concrete operator-facing gap that the escape path does NOT cover, do not widen the
set — record it in results.md as a decision packet for the orchestrator.

CHANGELOG entry (unreleased). Attach `TASK-260906-1xbrz6_results.md` (exact sentences added, files
touched, gate exit code) and hand off with `task-board handoff TASK-260906-1xbrz6 --role developer`.
The editorial follow-up (TASK-260906-3o75d6: state the takeover clause once in cli/curator.md) is a
separate leaf that runs after this one — do not do it here.
