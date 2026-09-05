# Review findings: launcher SPEC `0.2.0-draft` (cycle 2 — confirm the minor fixes)

Subject: `curator-agent-launcher` worktree `.worktrees/curator-agent-launcher-spec-0.2`, branch
`draft/spec-0.2`, head `e19eb9f` (commits `b025e3e`, `e19eb9f` on the cycle-1-accepted `ffe9b68`).
Scope reviewed: exactly `git diff ffe9b68..e19eb9f` — `SPEC.md` only, +19/−8. Reviewer run `RUN-260905-2df817`.

## Verdict: ACCEPT (no blocking or major findings)

## Cycle-1 minor findings

| # | Finding | Result at `e19eb9f` |
|---|---|---|
| 1 | §4.3 level 2 lock sentence contradicted itself ("used even against flags … flags are a usage error") | **resolved** — now: "a `--model`/`--effort` flag for a member the machine entry sets is a `usage` error naming the locked member". Single clause; consistent with the §6 usage row "a flag overriding a locked default". |
| 2 | operator-over-machine override stated per entry while the rest of §4.3 is per member | **resolved** — "the operator file overrides the machine file per member: an operator entry that sets only `model` leaves a machine `effort` for the same env-id in force", explicitly tied to the section's per-member resolution. |
| 3 | "integration configured" had no named source, load-bearing for the `--ax-profile` usage error | **resolved** — §4.6 names `ax.json` in the §4.3 level-2 directories, closed schema `{ "schema": "curator-run-ax-v1", "enabled": <boolean> }` with unknown members rejected, machine file over operator file, absent in both = not configured, unreadable/unparsable = `defaults_config_invalid` never "not configured", read once before argument handling completes. `defaults.json`'s own closed schema is untouched (separate sibling file, no new member). Consistent with §3 usage rules (usage errors "resolve nothing" — a configuration read is not resolution) and with §6 invariant 1 (fallback never fires on a failed read). |

Reuse of `defaults_config_invalid` for `ax.json`: judged acceptable, no distinct code needed. The
family semantics are identical (a launcher-owned machine-configuration file exists but cannot be
read, parsed, or violates its closed schema), and §6 invariant 1 already lists the code. A separate
`ax_config_invalid` would buy only a name and cost a §6 row; not worth it at draft stage.

## Residual minors (non-blocking; fold into 0.2.1 or the next touch)

1. **minor / §6 `defaults` row** — the row's description cites only "§4.3: a machine-configuration
   file … names an unknown env-id or member". It should also name `ax.json` (§4.6) since that
   file now raises the same code, so a reader mapping the code back finds both sources.
2. **minor / §9** — the docs-confidence item lists the `defaults.json` location and lock rule as
   "this document's own"; `ax.json` is the same kind of drafting choice and belongs in that bullet.
3. **note / §4.6 vs §3** — when `ax.json` is unparsable *and* the command line has a usage error,
   which fires first is unstated. Both are terminal, so no launch escapes either way; ordering can
   be pinned when the parser exists.

## Gates rerun by the reviewer at `e19eb9f`

```text
go build ./...
go vet ./...
go test ./... -count=1
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	3.113s
exit=0
```

`git diff --check ffe9b68..e19eb9f` clean. Worktree clean. Only `SPEC.md` changed in scope.
Negative evidence for the single gate (`TestSpecVersionPinned` rejects a mutated constant) was
established in cycle 1 at `ffe9b68`; this delta touches no Go file, so it is accepted from that
record and not rerun as a mutant.

## Commits

`b025e3e` and `e19eb9f`, both `Good "git" signature` ED25519 for ivan@relux.works, author
Ivan Oparin <ivan@relux.works>. Not the producer's commits (orchestrator applied the fixes per the
cycle-2 brief). PR status on relux-works/curator-agent-launcher not checked by this reviewer
(brief mentions `gh pr list`; landing is the orchestrator's step).

## On the Change Request's empty repository delta

`CR-TASK-260905-3ewdq0-1` revision 1 has `repository_delta=empty` against the curator-spec story
worktree and is already in state `accepted` from cycle 1. That remains the right outcome: every
deliverable of this leaf lives in the sibling repository `curator-agent-launcher` (`draft/spec-0.2`,
now `e19eb9f`), and the brief forbade writing anything into curator-spec or the control root. No
curator-spec file was supposed to change, and none did.
