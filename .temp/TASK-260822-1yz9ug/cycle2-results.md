# TASK-260822-1yz9ug — rework cycle 2 results: decision 0009 prose corrections landed

Run `RUN-260822-ec0450`, role `doc-writer`. Status: **ready for review**.

**Landed:** `be7861c` on `curator-spec` `main`, single-parent squash of
[PR 26](https://github.com/relux-works/curator-spec/pull/26), decision-only,
amending `decisions/0009-first-party-module-roots.md` in place.
**No revert** of `b92b105` (PR 24) or `3dc9ca6` (PR 25). The Decision section's
rules were not touched this cycle; the diff is 12 insertions / 9 deletions in one file.

## Scope: the three sentences of `TASK-260822-1yz9ug_review-verdict-2.md`

Each finding was re-derived first hand before editing, not taken from the verdict.

### F3 — Consequences bullet contradicted the Context

Was: *"restores the `replace` directives and vendor trees its `main` branch currently lacks"*.

Verified against `relux-works/skill-project-management` `origin/main`:

| Module | `replace` directives |
| --- | --- |
| `tools/board-cli/go.mod` | 3 — `pkg/board`, `pkg/remoteconfig`, `pkg/providerlimits` |
| `tools/board-tui/go.mod` | 1 — `pkg/remoteconfig` |
| `pkg/providerlimits/go.mod` | 1 — `pkg/remoteconfig` |

`git ls-tree -r --name-only origin/main | grep -c 'tools/board-cli/vendor/'` → `0`;
same for `tools/board-tui/vendor/` → `0`. `.gitignore:64-67` excludes `tools/*/vendor/`
with the stale-`-mod=vendor` rationale in-file. So `main` lacks **only** the vendor trees,
exactly as the amended Context says two screens above.

Now reads: *"restores the vendor trees its `main` branch currently lacks, keeps the
`replace` directives it already has, and declares its module roots, replacing the
unreplaced pre-vendored shape it ships at `ca5c4fd3`."*

### F4 — "The Decision section is unchanged" was false, about the load-bearing edit

`git diff b92b105 3dc9ca6 -- decisions/0009-first-party-module-roots.md` hunk headers:
`-2,7`, `-15,15`, **`-111,9`**, `-146,11`, `-172,7`, `-194,9`, `-222,9`, `-245,3`.

Section boundaries in `b92b105` (`grep -n '^## '`): `## Decision` at line **68**,
`## Rejected alternatives` at line **136**. Old line 111 falls inside `## Decision`, at
point 5. The hunk removed the false laundering claim and added the clause bounding it
(*"It does not, and by construction cannot, reach a package that presents no replacement
at all; open question 7 records what is left open."*).

The scoping predicate itself — "the exceptions are scoped to results whose module carries
no replacement" — is present on both sides of the diff. So the rules genuinely did not
move; the sentence bounding what point 5 delivers did. The fix states that precisely
rather than widening the claim.

Now reads: *"The Decision section's rules are unchanged; point 5 gains the bound that its
scoping does not reach packages presenting no replacement."*

The `3dc9ca6` commit message carries the same wording and is immutable — left alone, as routed.
The `LOGBOOK.md` 2052 entry carried it too and **was** corrected (see below).

### F5 — the `ca5c4fd3` no-replace claim was over-broad

Was: *"There the repository carries no `replace` directive at all."*

`git show ca5c4fd3:pkg/board/go.mod` carries one:

```
replace github.com/relux-works/skill-agent-facing-api/agentquery => ../../../skill-agent-facing-api/agentquery
```

Three levels up, outside the repository. Repo-wide count at `ca5c4fd3` (non-vendor `go.mod`s):
`pkg/board` 1, `pkg/remoteconfig` 0, `tools/board-cli` 0, `tools/board-tui` 0,
`tools/board-regress` 0, `tools/board-server` 0.

The operative claim survives — `replace` in a non-main module takes no part in resolution,
`pkg/board` is vendored at that revision, and the manager-visible `go list` stream still
presents `Module.Replace == nil` throughout — but the narrower statement is also the
stronger one, since it is the build root's `go.mod` the manager reads.

Now reads: *"There neither build root carries a `replace` directive."*

## Logbook

- **Corrected** the 2052 entry's "Decision section byte-unchanged" to the F4 wording, with
  an explicit in-line note that the original claim was falsified by the diff.
- **Added** entry `2205` recording all three findings, the three-cycle pattern (every defect
  on this task has been an unqualified prose claim about state one command falsifies, while
  the normative rules survived every cycle untouched), and two method traps below.

## Gates

Run on `bb064ae` in the landing worktree. Each standalone, real exit code, no pipe chains:

| Gate | rc |
| --- | --- |
| `python tools/validate.py` | **0** — 49 schemas, 471 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | **0** — 91 tests |
| `go test ./tools/...` | **0** |
| `go run ./tools/generate-vectors -root .` | **0** |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8}.json` | **0** |
| `gofmt -l tools` | **0** — empty output |
| `git diff --check` | **0** |
| `lychee` with the exact `ci.yml` arguments over `**/*.md` | **0** — 40 OK, 0 errors |

PR 26: **8 checks pass** (`Formatting`, `Links`, `Specification` ×3, `Implementations` ×3).
`Release target provenance` skips on pull requests by design (`ci.yml:87`) and ran
**green post-merge** on `be7861c`, which is the gate that verifies the signed commit.

Commit signed: `%G? = G`, `Ivan Oparin <oparin@me.com>`. No AI attribution.

## Cleanup

- Remote branch `spec/module-roots-decision-cycle2` deleted; local branch deleted.
- Worktree `.temp/STORY-260822-1pm1c9/worktree` removed; the story temp dir is empty.
- venv kept **outside** the worktree at `curator/.temp/TASK-260822-1yz9ug/venv`
  (`curator-spec` has no `.gitignore`); `git status` in the worktree showed exactly one
  modified file before staging.
- `gh pr merge --squash --delete-branch` **partially failed**: the remote merge succeeded but
  the local post-merge checkout aborted with `fatal: 'main' is already used by worktree at
  /Users/iv/Developer/ReluxWorks/curator-spec`, so `--delete-branch` never ran. Branch deleted
  manually. Verify merge state with `gh pr view` rather than trusting that command's exit code.

## Anomalies worth carrying

- **Disk hit 100%** mid-task (126Mi free of 926Gi). The harness could not write a Bash tool's
  output file; every call failed `ENOSPC` until space was freed. `go clean -cache` recovered
  **34Gi** (it exits non-zero with "directory not empty" while still reclaiming). Check
  `df -h /` before blaming tooling when tool calls fail for no apparent reason.
- **`${PIPESTATUS[0]}` is a bashism** and expands to empty under this zsh, so
  `cmd | tail; echo rc=${PIPESTATUS[0]}` reports *no* exit code. Redirect gate output to a log
  file and read `$?`.

## Still live, not owned by this task

- **`EPIC-260822-18ylpq` note** still says the snapshot shipped with the replaces *stripped*
  and that `2ed3acd` restored them. Both were disproved in cycle 1 and the landed decision
  explicitly denies the first. It is the input `TASK-260822-3nvx91` reads — align before that
  task spawns.
- **Numbering collision, third cycle running.** `draft/TASK-260728-12pnm1-rust-driver`,
  `draft/TASK-260728-168smo-kotlin-native-driver` and `draft/TASK-260728-1yhuqi-swift-driver`
  hold `0009`/`0010`/`0011` locally against a `main` that now owns `0009`. `decisions/`
  already carries a duplicate `0005` from this exact failure mode.
