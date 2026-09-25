# TASK-260924-1aa9wb — Skillfile schema 2 lane on by default (THE ONLY CURRENT INSTRUCTION)

Operator decision (2026-09-24): `CURATOR_DRAFT_SOURCES_V1` becomes default-ON. Control root: curator; your Story worktree only.
Read `campaign-producer-rules.md`. Landing gate = hosted CI at handoff. Source of truth: curator-spec main
`protocol/skillfile-sources.md` ("Skillfile schema 1 retains its exact meaning … No on-disk migration is implicit").
Today: `internal/install/draftsources.go` — `DraftSourcesEnabled` is true only for exactly "1"; callers:
`cmd/curator/project_resolve.go:67,85`, `draft_status.go:41`, `draft_help.go:90`, `main.go:1181`, `internal/install/install.go:234,390`.
Do exactly the task description and AC:
- unset / empty / any value other than "0" → lane ON; exactly "0" → OFF (transitional opt-out; document as deprecated). Keep the
  name of the variable; update help/docs wording from "opt-in draft" to "default; set =0 to opt out".
- PROVE schema-1 byte identity with the lane on vs off: run the existing v1 corpus / golden CLI outputs both ways and diff (the
  stbg4d review recipe did exactly this for "switch on vs off" — reuse it); any difference is a defect to fix, not to accept.
- A schema-2 project works end to end with NO environment variable through the real CLI entry.
- Update tests that asserted "schema 2 refused without the switch" to assert the =0 opt-out instead; never delete them.
- Narrowing mutants in a disposable copy: default back to off; "=0" ignored → each killed.
- CHANGELOG, README, docs/cli.md, docs/troubleshooting.md.
Bounded local runs (≤10 min/call; per package). Attach results, `resource update`, check DoD, `task-board handoff TASK-260924-1aa9wb --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
