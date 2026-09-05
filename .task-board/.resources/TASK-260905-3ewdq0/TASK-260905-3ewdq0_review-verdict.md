# Review findings: launcher SPEC `0.2.0-draft` (cycle 2 — confirm the minor fixes)

Reviewer run `RUN-260905-9241aa` (recovery attempt of `RUN-260905-2df817`, whose findings of the
same name were attached but whose lifecycle did not close; this file replaces them with an
independent re-review).

Subject: `curator-agent-launcher`, head `e19eb9f` (commits `b025e3e`, `e19eb9f` on the
cycle-1-accepted `ffe9b68`). The brief's worktree `.worktrees/curator-agent-launcher-spec-0.2`
no longer exists — PR #2 (`draft/spec-0.2`) was merged at 07:47Z and `main` of the main checkout
is `e19eb9f`; the review was done from that checkout read-only and gates from a `git archive`
scratch export. Scope: exactly `git diff ffe9b68..e19eb9f` — `SPEC.md` only, +19/−8,
`git diff --check` clean.

## Verdict: ACCEPT (no blocking or major findings)

## Cycle-1 minor findings 1–3

| # | Finding | Result at `e19eb9f` |
|---|---|---|
| 1 | §4.3 lock sentence contradicted itself | **resolved** — single clause: a `--model`/`--effort` flag for a member the locked machine entry sets is a `usage` error naming the locked member. Matches the §6 usage row ("a flag overriding a locked default"). |
| 2 | operator-over-machine override per entry vs per member | **resolved** — "overrides the machine file per member: an operator entry that sets only `model` leaves a machine `effort` for the same env-id in force"; consistent with the §4.3 preamble's per-member rule. |
| 3 | "integration configured" had no source | **resolved** — §4.6 names `ax.json` beside `defaults.json` (§4.3 level-2 directories), closed schema `{ "schema": "curator-run-ax-v1", "enabled": <boolean> }`, unknown members rejected, machine file decides when present, else operator, absent in both = not configured, unreadable/unparsable = `defaults_config_invalid` never "not configured", read once before argument handling completes. `defaults.json`'s closed schema is untouched (separate file). Consistent with §3 ("usage errors resolve nothing": a config read is not resolution) and §6 invariant 1 (fallback never fires on a failed read). |

Reuse of `defaults_config_invalid` for `ax.json`: acceptable, no distinct code needed. The
condition is identical (a launcher-owned configuration file exists but is unreadable or violates
its closed schema) and the code already sits in §6 invariant 1. A separate code buys a name only.

## Attack notes (gate behaviour, not just reading)

- Bypass path for `--ax-profile yolo` on an untracked machine: none — §3 makes it `usage`, and
  the fact it depends on is now a named read that precedes argument handling.
- Absence vs read failure for `ax.json`: distinguished explicitly; a malformed file cannot be
  mistaken for "untracked" (the dangerous direction: silently losing tracking).
- Version gate: `make check` exit 0 at `e19eb9f`; mutating `specVersion` to `0.2.1-draft` in the
  scratch export makes `TestSpecVersionPinned` fail (exit 1). Rerun by this reviewer.

## Residual minors (non-blocking; fold into 0.2.1 or the next touch)

1. **minor / §4.6 `ax.json`** — `"enabled": false` is never stated to mean "not configured";
   only absence is. Add: a present file with `enabled: false` is not configured (untracked mode).
2. **minor / §4.6 vs §4.3** — `ax.json` precedence is machine-over-operator while
   `defaults.json` is operator-over-machine unless locked. Defensible (tracking is machine
   policy) but the inversion is unexplained; one clause of rationale.
3. **minor / §6 `defaults` row** — cites only §4.3; `ax.json` (§4.6) now raises the same code
   and should be named there.
4. **minor / §9** — the docs-confidence bullet lists `defaults.json` as "this document's own";
   `ax.json` is the same kind of drafting choice and belongs in it.
5. **note** — when `ax.json` is unparsable and the command line also has a usage error, which
   diagnostic fires first is unstated. Both are terminal; pin when the parser exists.

## Gates rerun by this reviewer (scratch export of `e19eb9f`)

```text
go build ./...
go vet ./...
go test ./... -count=1
ok  	github.com/relux-works/curator-agent-launcher/cmd/curator-run	1.322s
make check exit=0
```

`specVersion = "0.2.0-draft"` (`main.go:18`), README line 24 `0.2.0-draft`, `TestSpecVersionPinned`
pins the same string. Main checkout of the launcher left untouched (`git status` clean).

## Commits

`b025e3e`, `e19eb9f`: both `Good "git" signature` ED25519 for ivan@relux.works, author
Ivan Oparin. Orchestrator-applied fixes, not producer commits. PR #2 is MERGED (landing was
the orchestrator's step; noted, not judged here).

## On the Change Request's empty repository delta

`CR-TASK-260905-3ewdq0-1` revision 1 carries `repository_delta=empty` against the curator-spec
story worktree and is already `accepted` from cycle 1. That is the right outcome: the brief
scoped every edit to the sibling repository `curator-agent-launcher` and forbade writing into
curator-spec or the control root. No curator-spec file was meant to change and none did; the
deliverable is verified at `e19eb9f` in that repository.
