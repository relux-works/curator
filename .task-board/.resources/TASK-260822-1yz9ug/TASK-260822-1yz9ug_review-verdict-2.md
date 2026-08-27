# TASK-260822-1yz9ug — review verdict (cycle 2): changes requested

Reviewer run `RUN-260822-5f6beb`, read-only. Verdict branch: **changes requested → `to-dev`**.

Scope of this cycle: the amendment landed at `3dc9ca6` (PR 25) answering
`TASK-260822-1yz9ug_review-verdict.md`. **Nothing here asks for a revert of `b92b105`
or `3dc9ca6`.** The normative core is unchanged and sound; the remaining defects are three
sentences of prose that do not match verifiable state, two of them introduced by the
amendment itself.

## The rework landed, and the hard parts landed correctly

| Cycle-1 request | Result | Evidence (re-derived first hand) |
| --- | --- | --- |
| 1. Replace the three `type: "system"` statements | done | All three gone. New Context says the manifest "was never the obstacle"; `origin/main:agent-skill.json` is schema 6, `build_roots ["tools/board-cli","tools/board-tui"]`, three commands `task-board` / `tb-sessiond` / `task-board-tui`, every one `"type": "build"` + `go-v1`. All three revisions that ever touched the file (`1e655eb`, `cc251cf`, `3dfb2d2`, under `--all --full-history`) contain only `"type": "build"` |
| 2. Per-module attribution | done | `origin/main:tools/board-cli/go.mod` = 3 replaces (`pkg/board`, `pkg/remoteconfig`, `pkg/providerlimits`); `tools/board-tui/go.mod` = 1 (`pkg/remoteconfig`); `pkg/providerlimits/go.mod` requires + replaces `pkg/remoteconfig`. Matches the amended text exactly |
| 3. Strip-and-pre-vendor as a rejected alternative | done | New entry with three failure modes, correctly framed as "the status quo this decision displaces" |
| 4. Resolve F2 | done, both halves | Point 5, Compatibility impact and Security impact narrowed to the delivered guarantee; open question 7 records the residual with the right diagnosis (needs a mechanism, not a predicate) |
| — Amendment recorded in the doc | done | Status section, though see F4 |

Landing mechanics: PR 25 `MERGED`, `mergeCommit 3dc9ca6`, single parent `b92b105` (squash);
8 required checks pass (`Formatting`, `Links`, `Specification` ×3, `Implementations` ×3),
`Release target provenance` `skipping` by design (`ci.yml:87`); both post-merge `main`
workflows green on `3dc9ca6`. No `spec/module-roots*` branch local or remote, no
`.temp/STORY-260822-1pm1c9` worktree. Signature/author shape (`%G? = E`, `98310998+ivanopcode@users.noreply.github.com`)
is identical to `b92b105` and `a2d44eb` — GitHub squash-merge, the repo's established pattern,
not a regression. `GOVERNANCE.md` requires context/decision/alternatives/compatibility/security
and imposes no supersede-vs-amend rule, so amending in place is admissible.

Both of the implementer's corrections to cycle 1 hold, and I confirm them:

- **Not a packaging-time rewrite.** `git show ca5c4fd3:tools/board-cli/go.mod` has zero
  `replace` lines and `git ls-tree ca5c4fd3 tools/board-cli/` lists a committed `vendor`
  directory. `vendor/modules.txt` has zero `=>` lines and lists both first-party modules as
  `## explicit`. The committed tree is already in that shape; the snapshot is a copy.
  Cycle 1's "stripped" was wrong, and dropping it was right.
- **Chronology.** `ca5c4fd3` (2026-08-07) is an ancestor of `origin/main`, 117 commits behind,
  so "has since drifted" is accurate. The amended text citing no commit archaeology at all is
  the right call.

Two further claims I checked because they are load-bearing and were not spot-checked in cycle 1:

- "It builds under `go-v1` today" — substantiated. `~/.curator/cache/build/go-v1/` holds three
  receipts whose `bin/` entries are exactly `task-board`, `tb-sessiond`, `task-board-tui`.
- "Its replaced modules are rejected outright as inconsistent vendor metadata" — exact:
  `cocoaskills/src/csk/builds/go_v1.py:980` raises `vendor_metadata_inconsistent` on
  `module.get("Replace") is not None`.
- `main` has no vendor tree under either build root; `.gitignore:64-67` excludes
  `tools/*/vendor/` with the stale-`-mod=vendor` rationale in-file, as quoted.

## F3 — the Consequences bullet contradicts the Context two screens above it

> First consumer: `skill-project-management` restores the `replace` directives and vendor trees
> its `main` branch currently lacks and declares its module roots […]

`main` does **not** lack the `replace` directives. It has four of them
(`tools/board-cli` ×3, `tools/board-tui` ×1, plus `pkg/providerlimits` ×1), which the amended
Context states correctly in its own second paragraph. What `main` lacks is the vendor trees.
The relative clause covers both conjuncts, so the bullet asserts the opposite of the Context
about half its subject — and it is the bullet that tells the first consumer what to do.

Fix: `…restores the vendor trees its `main` branch currently lacks, keeps the `replace`
directives it already has, and declares its module roots, replacing the unreplaced pre-vendored
shape it ships at `ca5c4fd3`.` (The parked branch `task/go-v1-switch` at `36c1e02` on that repo
already carries exactly that shape, per the `EPIC-260822-18ylpq` note.)

## F4 — "The Decision section is unchanged" is false, and false about the load-bearing edit

The Status section, the `3dc9ca6` commit message, the logbook entry and the board note all
state the Decision section is unchanged (the artifact says "byte-unchanged"). `git diff
b92b105 3dc9ca6` contains a hunk at `@@ -111,9 +136,11 @@` — inside `## Decision`, point 5 —
replacing three lines with five.

This is not a stray whitespace edit. The removed text was the false security claim
("so that `go mod vendor` cannot launder …"), and the added text is the qualification that
bounds it: *"It does not, and by construction cannot, reach a package that presents no
replacement at all; open question 7 records what is left open."* That clause changes what the
Decision asserts, not merely how it argues — which is precisely why the amendment was correct
to make it. The normative rules (the scoping predicate itself) are indeed unchanged; the
sentence as written is wrong on both readings.

Consequence if left: `TASK-260822-3nvx91` is told the normative source of truth did not move,
while the one sentence bounding point 5's guarantee did. Anyone diffing to confirm gets a
document contradicted by its own history in one command.

Fix: `The Decision section's rules are unchanged; point 5 gains the bound that its scoping
does not reach packages presenting no replacement.` Same correction to the logbook entry;
the commit message is immutable and needs nothing.

## F5 — "the repository carries no `replace` directive at all" is over-broad at `ca5c4fd3`

> There the repository carries no `replace` directive at all.

`git show ca5c4fd3:pkg/board/go.mod` carries one:

```
replace github.com/relux-works/skill-agent-facing-api/agentquery => ../../../skill-agent-facing-api/agentquery
```

Three levels up — outside the repository, pointing at a sibling checkout on a developer
machine. Repo-wide at `ca5c4fd3`: `pkg/board` 1, everything else 0.

The operative claim survives: the two build roots carry zero replaces, `replace` directives in
non-main modules take no part in resolution, and `pkg/board` is vendored at that revision, so
the manager-visible `go list` stream still presents `Module.Replace == nil` throughout. Only
the scope of the sentence is wrong, and it is wrong in the same direction cycle 1 flagged —
an unqualified claim about a consumer repository that one `git show` falsifies.

Fix: `There neither build root carries a `replace` directive.` — which is also the stronger
statement, since it is the build root's `go.mod` the manager reads.

## Requested changes

One decision-only PR to `curator-spec` `main`, same landing recipe, amending
`decisions/0009-first-party-module-roots.md` in place. Three sentences:

1. **F3** — Consequences bullet: `main` lacks the vendor trees, not the `replace` directives.
2. **F4** — Status: say the Decision section's *rules* are unchanged and name point 5's added
   bound. Correct the 2052 logbook entry the same way.
3. **F5** — Context: scope the no-replace claim to the build roots.

Nothing else in the document needs touching, and no `.md` outside `decisions/` is affected.
If the maintainer would rather not spend a third decision-only PR and CI cycle on three
sentences, folding them into `TASK-260822-3nvx91`'s normative PR is a defensible alternative —
that is a routing call, not a review finding, and it does not change that the sentences are
wrong today.

## Notes carried forward, not blocking this task

- **The `EPIC-260822-18ylpq` note still carries the two details the implementer disproved.** It
  says the snapshot "shipped with the replace directives STRIPPED" and that `2ed3acd` restored
  them. The amended decision explicitly denies the first ("This is not a packaging-time
  rewrite"), and the second is wrong (`2ed3acd` removed an `agentquery` replace; the first-party
  replaces returned via `1e655eb`). The note is the orchestrator's artifact and is the input
  `TASK-260822-3nvx91` reads, so it should be aligned with the landed text before that task
  spawns, or the normative producer inherits a framing the decision it implements refutes.
- **Numbering collision, still live.** `draft/TASK-260728-12pnm1-rust-driver`,
  `draft/TASK-260728-168smo-kotlin-native-driver` and `draft/TASK-260728-1yhuqi-swift-driver`
  hold `0009`/`0010`/`0011` locally against a `main` that now owns `0009`. `decisions/` already
  carries a duplicate `0005` from this failure mode.
- **Method note worth keeping.** In `zsh`, `git show "$c:agent-skill.json"` silently applies the
  `:a` history modifier and resolves to an absolute path — the command fails with a mangled
  filename that reads like a missing revision. Brace it: `"${c}:agent-skill.json"`.
- **Board hygiene.** Checklist items 9–21 duplicate DoD items 1–8; the mutation DSL exposes no
  removal. Items 1–8 are authoritative. Not a defect in the work.
