# TASK-260822-1yz9ug — review cycle 3 verdict: ACCEPTED

Reviewer run `RUN-260822-e17587`, 2026-08-22. Read-only; no file in either
repository was modified by this run.

## Verdict

**ACCEPTED.** The three cycle-2 findings (F3, F5, F4) are fixed in the exact
form the verdict requested, each fix re-derived here from first-hand commands
rather than taken from the implementer's report. No further defect was found
in the landed document.

Final state: `decisions/0009-first-party-module-roots.md` at curator-spec
`main` `be7861c`.

## What was verified

### Landing mechanics

| Check | Evidence |
|---|---|
| PR 26 merged, decision-only | `state=MERGED`, `mergedAt 2026-08-22T17:19:01Z`, one file changed, +12/-9 |
| Squash, not merge commit | `git rev-list --parents -n1 be7861c` → single parent `3dc9ca6` |
| Base is `main` | `baseRefName=main`, head `spec/module-roots-decision-cycle2` |
| All required PR checks green | 8 pass: Formatting, Links, Specification ×3 OS, Implementations ×3 OS. `Release target provenance` = `skipping`, correct: `ci.yml:87` is `if: github.event_name != 'pull_request'` |
| Post-merge main green | Both push workflows `success` on `be7861c`; within Specification CI the `Release target provenance` job itself is `success` — the signed-target gate did run on `main` |
| Branch and worktree removed | `git branch -a --list '*module-roots*'` empty; no `STORY-260822-1pm1c9` entry in `git worktree list` |
| No AI attribution | commit body scanned for co-authored/claude/generated-with/anthropic → none |
| Signature shape | `%G? = E`, GitHub `web-flow` squash author — identical to `b92b105`, `3dc9ca6` and `a2d44eb`; the repo's own provenance gate accepted it post-merge |
| Numbering | `decisions/` on `main` holds 0001–0009 (with the pre-existing duplicate 0005); 0009 was genuinely free at branch time |

### F3 — Consequences first-consumer bullet

Landed text: "restores the vendor trees its `main` branch currently lacks,
keeps the `replace` directives it already has, and declares its module roots".

Re-derived against `relux-works/skill-project-management` `origin/main`:

- `tools/board-cli/go.mod` lines 21/23/25 — three directory-form replaces
  (`pkg/board`, `pkg/remoteconfig`, `pkg/providerlimits`).
- `tools/board-tui/go.mod` line 50 — `pkg/remoteconfig` alone.
- `pkg/providerlimits/go.mod` line 7 — `pkg/remoteconfig`.
- `git ls-tree -r --name-only origin/main` → **0** paths under
  `tools/board-cli/vendor/` and **0** under `tools/board-tui/vendor/`;
  `.gitignore:67` is `tools/*/vendor/`, with the `-mod=vendor` rationale in
  the comment at lines 65–66.

`main` has the directives and lacks only the vendor trees. The bullet now
says exactly that, and no longer contradicts the Context two screens above.

(`tools/board-server/go.mod:34` also carries a replace, but `board-server` is
not a build root — `agent-skill.json.build_roots` is exactly
`["tools/board-cli", "tools/board-tui"]` on both revisions — so the
document's scoping is right to omit it.)

### F5 — Context scope at `ca5c4fd3`

Landed text: "There neither build root carries a `replace` directive."

Every `go.mod` at `ca5c4fd3`, replace-line counts:
`pkg/board` **1** (→ `../../../skill-agent-facing-api/agentquery`, outside
the repository), `pkg/remoteconfig` 0, `tools/board-cli` 0,
`tools/board-regress` 0, `tools/board-server` 0, `tools/board-tui` 0.

The old "carries no `replace` directive at all" was over-broad; the new
sentence is true and is the stronger claim, since the build root's `go.mod`
is what the manager reads.

The surrounding Context claims hold exactly:
`tools/board-cli/go.mod` requires `pkg/board v0.0.0` and
`pkg/remoteconfig v0.0.0` as ordinary requirements (no `providerlimits` —
that module did not exist at `ca5c4fd3`, `git ls-tree ca5c4fd3 pkg/` returns
`pkg/board` and `pkg/remoteconfig` only, so naming two is correct, not an
omission); `vendor/modules.txt` lines 24–36 list both as `## explicit` with
no `=>`; 79 vendored files sit under
`vendor/github.com/relux-works/skill-project-management/pkg/`. The on-disk
snapshot `~/.curator/cache/project-management/ca5c4fd3…/snapshot` matches the
git tree (0 replaces in both build roots, both `vendor/modules.txt` present)
— confirming "not a packaging-time rewrite". `ca5c4fd3` is an ancestor of
`main`, 117 commits back, so "has since drifted" is right.

### F4 — Status section

Landed text: "The Decision section's rules are unchanged; point 5 gains the
bound that its scoping does not reach packages presenting no replacement."

`git diff b92b105 3dc9ca6` carries a hunk `@@ -111,9 +136,11 @@`; `## Decision`
spans lines 68–135 at `b92b105`, so the hunk is inside it. Reading the hunk:
the rule sentence ("the exceptions are scoped to results whose module carries
no replacement") is present on **both** sides; what changed is the trailing
justification — the false laundering guarantee out, the explicit bound
("It does not, and by construction cannot, reach a package that presents no
replacement at all") in. The new Status wording describes that precisely.

`LOGBOOK.md` 2052 entry now carries the corrected sentence plus an inline
note that the original "byte-unchanged" claim was falsified by the diff; new
entry 2205 records all three findings. The `3dc9ca6` commit message still
carries the old claim and is correctly left immutable.

## Additional claims checked this cycle (not previously verified)

- **"Rather than restructure, it dropped its `replace` directives and
  vendored its own modules"** (Rejected alternatives) — supported by history,
  not just by present state: `3dfb2d2` (2026-08-07, *"Add curator Go
  vendoring for skill"*) deletes exactly the three replace directives from
  `tools/board-cli/go.mod`, 6 deletions, nothing else. Found with
  `git log --full-history -S`.
- **"a non-standard result with a non-nil `Module.Replace` fails as
  inconsistent vendor metadata"** — exact in both implementations:
  cocoaskills `src/csk/builds/go_v1.py:980` (`module.get("Replace") is not
  None` → `GoV1Error("vendor_metadata_inconsistent", …)`) and curator `main`
  `internal/godriver/build.go:517`, with `build_test.go:130` naming the
  "replaced module" case for that code.
- **"It builds under `go-v1` today"** — three receipts under
  `~/.curator/cache/build/go-v1/` for `task-board`, `tb-sessiond`
  (build root `tools/board-cli`) and `task-board-tui` (build root
  `tools/board-tui`), each with a produced `bin/` artifact, `module_mode:
  vendor`, sharing one `curator-build-source-v1` digest.
- **Governance shape** — `GOVERNANCE.md:61-66` requires context, decision,
  alternatives, compatibility impact, security impact; all five sections are
  present. No decisions index exists in the repo (`0008` is referenced from
  no file either), so no index update was owed.
- **Document read end to end** (327 lines). Points 1–8 cover the `modules`
  declaration, the bijection with directory-form replace directives, the
  admitted directive form, containment, the scan-surface extension, external
  dependencies, unchanged `curator-build-source-v1` identity per 8.1, and the
  failure boundary. Both DoD-required rejections are present (implicit
  manager input; repository consolidation), plus four more. Open question 7
  carries the F2 residual forward to `TASK-260822-3nvx91`.

## Definition of Done

All eight authoritative items (1–8; items 9–21 are the duplicate set the
implementer recorded on the board) are satisfied and independently verified
above. Items 9–21 restate them and are checked truthfully.

## Handed to the commit-owning mover

This reviewer run supplies no `commit_ack`. Acceptance evidence is this
artifact. The spec change is already merged (`be7861c`); what remains
uncommitted is the **curator-side** working tree — `LOGBOOK.md` (entries 2205
and the 2052 correction) and the board files. The commit-owning mover commits
that scope and makes the enforced `done` transition with
`commit_ack=scope_committed`.

## Carried forward — not blocking this task

1. **`EPIC-260822-18ylpq` note is still wrong in two places**, and it is the
   input `TASK-260822-3nvx91` reads. It says `ca5c4fd3` "shipped with the
   replace directives STRIPPED" — the landed decision explicitly denies a
   packaging-time rewrite, and the committed tree at that revision is already
   replace-free — and it attributes the restoration of the replaces on `main`
   to `2ed3acd`, which cycle 1 disproved (`2ed3acd` removed an `agentquery`
   replace; the first-party replaces returned via `1e655eb`). Board metadata
   under the orchestrator, outside this task's deliverable, but it should be
   aligned before `TASK-260822-3nvx91` spawns or that task inherits a premise
   the decision it implements contradicts.
2. **Numbering collision, live for a third cycle.**
   `draft/TASK-260728-12pnm1-rust-driver`,
   `draft/TASK-260728-168smo-kotlin-native-driver` and
   `draft/TASK-260728-1yhuqi-swift-driver` hold `0009`/`0010`/`0011` locally
   against a `main` that now owns `0009`. `decisions/` already carries a
   duplicate `0005` from exactly this failure mode.

## Note on the three-cycle pattern

Every defect on this task, across three cycles, was an unqualified prose
sentence about state — the consumer's manifest, its `go.mod`, the document's
own diff — that one command falsifies. The normative rules in `## Decision`
survived all three cycles substantively untouched. The implementer's 2205
logbook entry records this; it is worth carrying into
`TASK-260822-3nvx91`, where the same repositories get described again.
