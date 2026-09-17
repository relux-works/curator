# Brief — TASK-260910-2ohnjo: spec bounds for MCP env passthrough and declaration surfacing (S4)

Story `STORY-260910-1lf0m5` (bound-mcp-declaration-exposure), wave 1.
Worktree: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`.
Rules: `remediation-spec-producer-rules.md` (attached). Role: doc-writer (technical writer) of
normative spec text; no implementation.

## Finding (read it first)
`docs/security-audit-2026-09.md` S4 (spec) and `curator/docs/security-audit-2026-09.md`
S4 (manager): `stdio` MCP declaration packages (`protocol/environments.md` §2.2)
execute `command`+`args` at launch; `passable_env_names` defaults to `null` =
unbounded (§12.1, §10.3), `mcp_package_allowlist` defaults to empty = permits
all (§2.2, §12.1). A hostile declaration package runs arbitrary programs and
receives named operator secrets.

## Settled decisions (do not reopen)
- `passable_env_names` default becomes **empty** (opt-in per name). Decide and
  state whether `null` keeps meaning "unbounded" as an explicit operator choice
  (recommendation: keep `null` as the explicit, lockable-away unbounded value;
  absent knob = empty). Update §12.1 (default), §12.2 (already lockable), §2.2
  and §10.3 wording.
- A **loud warning** when `mcp_package_allowlist` is empty: a diagnostic emitted
  at `profile install`, `profile update` and in `env status` posture (working
  spelling `mcp_package_allowlist_empty`), stating that every declaration
  package in the closure is admitted.
- Surfacing: at `profile install` and `profile update` the manager MUST print,
  per MCP declaration package, the resolved stdio `command` + `args` (and the
  `env_names` it requests), before materialization; `env status` repeats them.
  Specify the exact output rows (closed columns) the way the document specifies
  other CLI output.
- Warn-first rollout (impact row "S4 passthrough default", migration hint names
  the variables): revision A keeps the unbounded default but warns for every
  passed operator variable not listed in `passable_env_names` (naming the
  variables and the knob); revision B makes the default empty and drops the
  unlisted names with a diagnostic. Specify both explicitly.
- Longer-term closed interpreter contract (audit item 4) is NOT in this task;
  note it as a future revision.

## Deliverable
1. §2.2 and §10.3 text: the new default, the opt-in semantics, the allowlist
   warning, the surfacing requirement, the two rollout profiles.
2. Diagnostics tables (§2.1 / §10.4 / §9.7 as applicable): the allowlist-empty
   warning, the revision-A passthrough warning, the revision-B drop diagnostic;
   `env status` posture rows.
3. §12.1 row updates (`passable_env_names` default; `mcp_package_allowlist`
   default annotation "empty = permits all, warned").
4. Conformance vectors under `conformance/v1/` (+ manifest): default resolution
   of `passable_env_names` under both profiles; allowlist-empty warning; the
   surfacing output bytes for one stdio declaration; schema cases for the knob
   default if the machine-config schema exists under `schemas/v1/`.
5. `CHANGELOG.md` Unreleased entry "S4: …" naming the two rollout steps.

## Out of scope
Implementation (`TASK-260910-gocke2`), E3 codex seed (`STORY-260916-1i1gfo`),
S1/S3 hardened defaults (`STORY-260910-2qmrb8`).

## Checklist and handoff
Tick the checklist items you satisfy; attach
`TASK-260910-2ohnjo_change-request_rev1.patch` and `TASK-260910-2ohnjo_evidence.md`
(with the `make validate` transcript), then
`task-board handoff TASK-260910-2ohnjo --role doc-writer`.
