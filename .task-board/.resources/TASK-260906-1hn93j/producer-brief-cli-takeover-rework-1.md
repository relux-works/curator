# Producer brief: rework 1 — close the takeover gap in the spec, then publish both CLI rows

## Verdict

Review cycle 1 returned CHANGES REQUESTED with one blocking finding, two major and one minor. Read
`TASK-260906-1hn93j_review-findings-cli-1.md` in full before touching anything. The reviewer's
argument is accepted by the orchestrator and is not open for re-litigation: the standalone
`curator env takeover --takeover` shape is not merely unsourced, it is excluded by §9.5, and the row
it produced documents a refusal branch the published command cannot reach.

The reviewer offered two fixes and correctly noted that fix (a) alone still publishes an enumeration
`environments.md` does not state. Neither of its options is therefore taken as written. The
orchestrator's decision is to **close the spec gap first, in this same batch**, and then publish the
rows on the closed sentence. Your scope is widened accordingly.

## Scope, widened

You now edit three files instead of one.

### 1. `protocol/environments.md` §9.5 — close the takeover shape

Replace the sentence

> Takeover of a specific unmanaged file outside onboarding requires the explicit takeover flag and
> performs the same notice and backup; without the flag, section 8.3 applies and the operation fails
> rather than overwrite.

with text that fixes exactly these three things and nothing more:

- **Shape.** The takeover is a flag carried by a mutating operation, never an operation of its own.
  Say so in the document's own voice, so that "the operation" in the refusal clause has an antecedent.
- **Enumeration.** The flag is accepted on exactly the mutating operations this section already names
  as onboarding triggers — `profile install`, `profile use`, `profile sync`, `profile update`, and
  `env resolve --repair` — and on no other operation. That set is closed. It is the only candidate set
  the document supplies, and a closed set is what the revision-1 discipline of this specification
  requires everywhere else; do not widen it and do not leave it open.
- **Per-file semantics preserved.** The takeover covers the unmanaged files the carrying operation
  would write — no more. §8.3's "a second takeover of the same path" and §9.5's "a specific unmanaged
  file" must both still read true. The flag never selects a scope of its own.

Keep the surviving clauses intact: the same notice and backup as onboarding, and the without-flag
failure with `environment_surface_unmanaged_conflict` under §8.3. Change no diagnostic name, add no
new diagnostic, and touch no other section. The document stays revision 1 (its header says why), so
there is no revision line to bump.

### 2. `profiles/manager.md` — match

The manager profile carries the same ambiguous sentence (the reviewer cites it around line 2295:
"Takeover of a specific unmanaged file outside onboarding requires the explicit takeover flag ...
without the flag the section 12.2 ledger discipline fails the operation rather than overwrite").
Bring it into agreement with the amended §9.5 in the manager's own voice — the manager states the
manager-side obligation and cites the environments section; it does not restate the rule normatively.

### 3. `cli/curator.md` — publish the rows

- **Delete** the standalone `curator env takeover` row and its example line entirely.
- **Add `[--takeover]`** to the five rows the amended §9.5 enumerates: `profile install`,
  `profile use` (both published forms — judge whether `--clear` can meet unmanaged state and say so
  in the report), `profile sync`, `profile update`, and `env resolve`. One short clause per row,
  sourced to the amended sentence. On `env resolve`, state its relationship to `--repair` exactly as
  the amended §9.5 does — the flag matters only where the operation writes.
- **Fix the import row** per finding 3: `curator profile import [--as <name>] [--allow-lossy] [--use]`,
  with the §9.1 activation clause the install row already carries (`--use` takes no name; first
  install activates and says so). §9.6's "Activation follows the section 9.1 rules without magic" is
  the source; this is not a free spelling.
- Update the examples block: the takeover example becomes one that carries the flag on a real
  operation.
- `cli/curator.md` mentions onboarding nowhere today. Adding the flag to five rows is the first place
  those rows say why they can meet unmanaged state at all. Keep that to the minimum each row needs;
  the index still is not a second normative source.

### 4. `CHANGELOG.md`

Revise the `## Unreleased` entry: the takeover is now a normative amendment to §9.5 and manager.md
plus a CLI-surface change, not the addition of a new command. Put each half under the heading the
repository's convention gives it.

## What the review found and you must not repeat

- **Finding 2 (major).** Do not publish `--env`/`--target` for a takeover, and do not invent a path
  operand instead. Under the amended sentence the takeover's scope is the carrying operation's scope
  and no scope grammar is needed.
- **Finding 4 (minor).** Your verification bound claimed "no reference to `curator.md` or `CHANGELOG`
  anywhere under `tools/`". The `CHANGELOG` half is false (`tools/release_gate.py:650,722` and its
  tests), and a filename grep cannot establish the `curator.md` half because
  `tools/validate.py:3311` `validate_local_links()` rglobs every `*.md`. The conclusion happened to
  hold; the method could not establish it. In this report, state bounds you actually established, by
  the method that establishes them.
- The reviewer dismissed Q1, Q3 and Q4 as non-defects because `cli/curator.md:3` declares itself
  informative and says other managers may use different command names and flags. Accept that: keep
  `--takeover`, `--allow-lossy`, `profile import` and `--as` as published. Q2 is subsumed by Q5, and
  Q5 is what this rework closes in the spec rather than in the index.

## Delivery

**Exactly one signed commit past curator-spec main `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`** —
amend or `git reset --soft` your existing `f013e0c` so the branch carries one commit, not two. A
Change Request candidate must be exactly one single-parent commit past the Story base or the board
refuses it. Human identity, no `LOGBOOK.md`, no stray file.

Gates: `make validate` and `make regenerate-check` green (venv under `.temp/venv`) — run them
yourself and quote the tails with observed exit codes. State explicitly in the report whether the
§9.5 amendment touches any schema, vector, diagnostic table or conformance case; if it does, that is
a finding to raise, not a change to make silently. Do not push, tag, or open a PR.

Attach `TASK-260906-1hn93j_rework-report-1.md`: finding → resolution table for all four findings; the
amended §9.5 and manager sentences quoted in full, before and after; a clause → source-sentence table
for every published row clause, now sourced to the amended text; the gate tails; and honest
verification bounds. Then `task-board handoff TASK-260906-1hn93j --role developer`. Never write into
the control root.
