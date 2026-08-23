# TASK-260822-1yz9ug — rework results: decision 0009 amended and landed

**Run:** `RUN-260822-2bf138` (doc-writer rework)
**Date:** 2026-08-22
**Landed:** curator-spec `main` `3dc9ca6b938db3eea0931190cd471b8d66e66238`, single-parent squash of
[PR 25](https://github.com/relux-works/curator-spec/pull/25)
**Amends:** `decisions/0009-first-party-module-roots.md` (originally landed at `b92b105`, PR 24)
**No revert of `b92b105`.** The Decision section is byte-unchanged.

Supersedes nothing; complements `TASK-260822-1yz9ug_results.md` (the original landing) and
answers `TASK-260822-1yz9ug_review-verdict.md`.

## Routing addressed

The rework routing assigned points 1–4 of the review's five requested changes; point 5 (the
`EPIC-260822-18ylpq` note) was handled by the orchestrator.

| # | Requested | Where it landed |
| --- | --- | --- |
| 1 | Replace the three `type: "system"` statements with the verified consumer shape | Context (rewritten, two new paragraphs), Rejected alternatives, Consequences |
| 2 | Fix per-module attribution | Context paragraph 2 |
| 3 | Add the strip-and-pre-vendor shape as a rejected alternative with real failure modes | Rejected alternatives (new entry, three sub-points) |
| 4 | Resolve F2 — narrow the claim **or** record the residual | **Both**: point 5, Compatibility impact and Security impact narrowed; residual recorded as open question 7 |
| — | Amendment recorded in the document itself | Status |

## Independent verification

Every reviewer claim was re-derived at first hand before writing; two supporting details did not
survive and were dropped rather than repeated.

### Confirmed

| Claim | Command | Result |
| --- | --- | --- |
| No `type: "system"` command ever | `git log --all -S'"type": "system"' -- agent-skill.json` | empty — never existed |
| `origin/main` manifest is build/`go-v1` | `git show origin/main:agent-skill.json` | schema 6, 3 commands, all `"type": "build"` / `go-v1` |
| `board-cli` carries all three replaces | `git show origin/main:tools/board-cli/go.mod \| grep -c '^replace'` | `3` |
| `board-tui` carries one | `git show origin/main:tools/board-tui/go.mod \| grep -c '^replace'` | `1` (`pkg/remoteconfig`) |
| `pkg/providerlimits` replaces `pkg/remoteconfig` | `git show origin/main:pkg/providerlimits/go.mod` | confirmed |
| No vendor tree on `main` | `git ls-tree --name-only origin/main tools/board-cli/` | no `vendor` entry |
| `.gitignore` excludes `tools/*/vendor/` | `git show origin/main:.gitignore` | present, with the stale-vendor-tree rationale in-file |
| `ca5c4fd3` snapshot: zero replaces | `grep '^replace' <snapshot>/tools/board-cli/go.mod` | zero |
| `ca5c4fd3` snapshot: zero `=>` in modules.txt | `grep '=>' <snapshot>/tools/board-cli/vendor/modules.txt` | zero; both first-party modules listed `## explicit; go 1.25.5` |
| First-party sources vendored | `ls <snapshot>/.../skill-project-management/pkg/` | `board`, `remoteconfig` |
| `ca5c4fd3` is what is pinned | `skill-project-management/Skillfile.json:8` | `ca5c4fd3dec4d30c4c9360bc81f268314e544dd1` |

Snapshot path: `~/.curator/cache/project-management/ca5c4fd3dec4d30c4c9360bc81f268314e544dd1/snapshot`.

### Not confirmed — corrected before writing

1. **"the packaged snapshot ships with the `replace` directives *stripped*"** implies a
   packaging-time rewrite. There is none. `git show ca5c4fd3:tools/board-cli/go.mod` has zero
   `replace` lines and `git ls-tree ca5c4fd3 tools/board-cli/` lists a committed `vendor`
   directory — the **committed tree at that revision is already in that shape**, and the
   installed snapshot is a faithful copy. The amended Context says so explicitly, because a
   manager-side reader would otherwise go looking for a transformation step that does not exist.
2. **"`2ed3acd` added the directory-form replaces"** is wrong. `2ed3acd` (2026-07-30) *removed*
   an `agentquery` replace and touched nothing first-party; `b6c7404` (2026-08-20) touched only
   `.gitignore`, `README.md`, `SKILL.md` and `references/`. The replaces returned to `main` via
   `1e655eb` (2026-08-18, a trunk carry). None of this belongs in a spec decision, so the
   amended text cites **no commit archaeology at all** — only the two present-state facts:
   `main` has the replaces and no vendor tree; `ca5c4fd3` has neither problem.

   Method note: `git log -S<pat> -- <path>` under default history simplification hid `1e655eb`
   entirely; only `--full-history` surfaced it. Pickaxe on a merge-heavy branch is not proof of
   absence.

## The F2 resolution, and why both halves

The review offered "narrow the security claim **or** record the residual as an open question".
Both were done, deliberately:

- **Narrowing is not optional.** Point 5's claim that scoping decision 0005's exceptions to
  unreplaced results stops `go mod vendor` laundering package-controlled assembly and
  `cgo_import_dynamic` into the build is false as written. A package with no `replace` directive
  presents `Module.Replace == nil`, declares no `modules`, never triggers the bijection, and
  keeps the exceptions. Leaving that sentence standing would put a security guarantee in an
  accepted decision that the mechanism does not deliver. Point 5, the Compatibility impact bullet
  and the Security impact paragraph now state the bounded guarantee: the scoping reaches the
  newly admitted shape and nothing else.
- **Narrowing alone would bury the gap.** The route stays open, it is pre-existing rather than
  introduced by 0009, and it is what the named first consumer ships today. Closing it needs a
  *mechanism* — nothing in the fixed `go list` stream separates first-party from third-party
  vendored code once the replacement is gone — so it is a design question for the normative
  change, not something a decision about the declared form can answer. Open question 7 puts it
  in front of `TASK-260822-3nvx91` explicitly instead of leaving it implied.

The Decision section was left untouched throughout: nothing in F1 or F2 changes what the protocol
should do, only what the document may claim.

## Gates

Run on `5cb234b` before push, each as a standalone process, real exit codes:

| Gate | Exit code | Note |
| --- | --- | --- |
| `python tools/validate.py` | 0 | 49 schemas, 471 vector files |
| `python -B -m unittest discover -s tools -p 'test_*.py'` | 0 | 91 tests, 120.5s |
| `go test ./tools/...` | 0 | |
| `go run ./tools/generate-vectors -root .` | 0 | |
| `git diff --exit-code -- conformance/v1 release/1.0.0-rc.{5,6,7,8}.json` | 0 | deterministic regeneration |
| `gofmt -l tools` | 0 | empty output |
| `git diff --check` | 0 | |

PR 25 checks: `Formatting`, `Links`, `Specification` ×3 (ubuntu/macos/windows), `Implementations`
×3 — all **pass**. `Release target provenance` reported `skipping`, which is by design:
`.github/workflows/ci.yml:87` gates it on `github.event_name != 'pull_request'`.

Commit `5cb234b` signed with the maintainer key (`%G?` = `G`, `oparin@me.com`), matching
`maintainers.allowed_signers`. No AI attribution in the commit or PR body.

## Cleanup

- Worktree `curator-spec/.temp/STORY-260822-1pm1c9/amendment-worktree` removed; parent directory
  removed.
- Local branch `spec/module-roots-decision-amendment` deleted.
- Remote branch `spec/module-roots-decision-amendment` deleted.
- Verified: `git worktree list` has no match, `git branch --list` empty, `git ls-remote --heads
  origin 'spec/module-roots*'` empty.

## Anomalies and warnings

- **Host lacks `jsonschema`.** `python3 tools/validate.py` fails at exit 1 with
  `ModuleNotFoundError: No module named 'jsonschema'`. A throwaway venv from
  `requirements-dev.txt` inside the worktree works, but it **must be removed before `git add`**,
  along with `tools/__pycache__/` which the unittest run creates untracked.
- **The unittest gate takes ~120s**, which exceeds a 2-minute default tool timeout. Budget for
  it or it reads as a hang.
- **Numbering collision warning, still live.** `draft/TASK-260728-12pnm1-rust-driver`,
  `draft/TASK-260728-168smo-kotlin-native-driver` and `draft/TASK-260728-1yhuqi-swift-driver`
  hold `0009`, `0010`, `0011` locally and now collide with `main`. `decisions/` already carries a
  duplicate at `0005` from the same failure mode. Any of those branches landing without
  renumbering produces a third collision.
