# Landed delivery evidence reconciliation — rework verification

## Context

Board task: `BUG-260810-2oxt8b` (`reconcile-landed-delivery-evidence`).

This researcher pass independently verifies the supported legacy-status repair and the resulting Curator board state after the first reviewer requested changes. The work is metadata and research only: no product code, website work, adapter implementation, commit, push, pull request, or merge was performed.

Evidence was checked on 2026-08-10 through task-scoped `task-board` queries, read-only `gh` queries, and read-only `git` inspection. Remote `main` resolved to `124654a051f60995e826b75ab43a3c4f3359f5ff`, and `agent/curator-conformance-review` resolved to `8f9b90a31ea04006680866fc05cb7d1f9a675d2a`; both matched the local remote-tracking refs.

## Key findings

1. The legacy-status repair path is available. [`skill-project-management` PR #4](https://github.com/relux-works/skill-project-management/pull/4), `Add legacy status normalization`, is `MERGED` at merge commit `8dc0b71490214fe5ead6bf9cfde9574df084fd91`; installed `task-board` reports `0.24.3-3-g8dc0b71`.
2. All four historical delivery items now read `done` through the CLI. Their corresponding evidence resources remain attached, including the formerly blocked PR #5 resource on `STORY-260713-72b914`.
3. Parent aggregation is correct: the shell, manager, and registry delivery parents are `done`; the conformance-review Story and its Epic remain active only because this Bug is in its producer cycle. They should become `to-review` on researcher handoff and `done` only after reviewer acceptance of this Bug.
4. PRs #5, #6, and #7 remain merged, their merge commits are ancestors of current `origin/main`, and each cited CI run has six completed successful jobs.
5. PR #1 remains closed without merge. Its implementation is represented on main by common ancestry and stable patch equivalence; its only non-equivalent branch commit is board-only closure evidence.
6. Adapter planning remains a separate backlog direction with Swift, Kotlin, and C as the baseline and evidence-based discovery before considering more languages. The dedicated website remains an empty backlog Epic.
7. Strict board validation passed with exit code 0 and reported `valid=true`, `errors=[]`, and `warnings=[]` after the repair state was inspected.

## Task-scoped evidence matrix

| Board item | Required evidence and exact transition | Independent verification | Current disposition |
| --- | --- | --- | --- |
| `TASK-260713-c7a18d` (`harden-shell-activation`) | PR #6 evidence supported `reviewing` → `done` | [PR #6](https://github.com/relux-works/curator/pull/6) is `MERGED`; head `7adce172da7c480ba1f2e601e93312a0f18dbb39`; merge `7cde9b9bcd9fae26ddd2af3dd29aa380d0c4e12e`; [CI 29263194544](https://github.com/relux-works/curator/actions/runs/29263194544) is `completed/success` with six successful jobs; merge-ancestor command exited 0 | `done`; evidence resource `TASK-260713-c7a18d_PR6-delivery-evidence.md` attached; parent `STORY-260713-9f4c2b` and `EPIC-260712-785993` are `done` |
| `STORY-260713-b4e219` (`seamless-manager-lifecycle`) | PR #7 evidence supported legacy empty Story `backlog` → `done`, without fabricating a leaf task | [PR #7](https://github.com/relux-works/curator/pull/7) is `MERGED`; head `a78545cd2ca02b025be1af72ed2fe4a5b5a9210d`; merge `20fd90c5fbc113331e5cd365ce48c507c912dafd`; [CI 29282453876](https://github.com/relux-works/curator/actions/runs/29282453876) is `completed/success` with six successful jobs; merge-ancestor command exited 0 | `done`; evidence resource `STORY-260713-b4e219_PR7-delivery-evidence.md` attached; shared parent `EPIC-260712-785993` is `done` |
| `TASK-260713-7a9c1e` (`review-and-correct`) | PR #1 closure and patch equivalence supported `reviewing` → `done` | [PR #1](https://github.com/relux-works/curator/pull/1) is `CLOSED`, `mergedAt=null`, head `8f9b90a31ea04006680866fc05cb7d1f9a675d2a`; both [implementation CI 29213237436](https://github.com/relux-works/curator/actions/runs/29213237436) and [closure CI 29213318140](https://github.com/relux-works/curator/actions/runs/29213318140) are `completed/success` with six successful jobs | `done`; evidence resource `TASK-260713-7a9c1e_PR1-equivalence-evidence.md` attached; parent aggregation waits only on this Bug's review cycle |
| `STORY-260713-72b914` (`production-profile-conformance`) | Supported normalization of legacy raw status to canonical `backlog`, then PR #5 evidence supported `backlog` → `done`; its only-child Epic should aggregate to `done` | [PR #5](https://github.com/relux-works/curator/pull/5) is `MERGED`; head `e423ce3f2838d48e9543e5c39e9e04428ee100d1`; merge `75e802aa6dd65757d4763490d3950470fb49b150`; [CI 29270137737](https://github.com/relux-works/curator/actions/runs/29270137737) is `completed/success` with six successful jobs; merge-ancestor command exited 0 | `done`; all four checklist items are checked; `STORY-260713-72b914_PR5-delivery-evidence.md` is attached; `EPIC-260712-d77d32` is `done` with `blockedBy=[]` |
| `BUG-260810-2oxt8b` (`reconcile-landed-delivery-evidence`) | Researcher `analysis` → `to-review`; reviewer `reviewing` → `done` only if this evidence and all acceptance gates remain accepted | Current research outcome, exact status/resource checks, and strict validation provide the producer evidence | Ready for producer handoff; `STORY-260713-d0d4e8` and `EPIC-260712-c8ac0f` should follow this Bug to `to-review`, then aggregate to `done` only after acceptance |

The transition history for the repaired legacy values is sourced from the task's rework brief. This pass independently verifies the resulting canonical CLI state, resources, checklist, parent aggregation, installed repair version, and the merged repair PR; it does not claim direct observation of the earlier mutation invocation.

## Closed conformance branch fact-check

The current remote and local branch refs agree. These standalone checks all exited 0:

- `git merge-base origin/main origin/agent/curator-conformance-review` returned `d60f8d065dee6461512637f0c85c34c369574fa1`.
- `git merge-base --is-ancestor d60f8d065dee6461512637f0c85c34c369574fa1 origin/main` confirmed that shared implementation history is on main.
- `git cherry -v origin/main origin/agent/curator-conformance-review` returned `- 294aa8725cb6a3a9f3102896bfa3f418c7408e3e` and `+ 8f9b90a31ea04006680866fc05cb7d1f9a675d2a`.
- Patch-ID commands ran under `zsh -o pipefail`; branch commit `294aa8725cb6a3a9f3102896bfa3f418c7408e3e` and main commit `3efdbca5790baf0a04b9b9d3fc0fe9bfb46d0363` both produced stable patch ID `fdba908b17a82392386c79dfe25ae3ac79bbf2e5`.
- `git diff-tree --no-commit-id --name-status -r 8f9b90a31ea04006680866fc05cb7d1f9a675d2a` listed four `.task-board` paths and no product-code path.

Inference: the closed PR #1 branch contains no missing product implementation. Main contains the shared history literally and the remaining code change by patch equivalence; unique commit `8f9b90a...` records closure metadata only.

## CI evidence summary

Every cited `gh run view` command exited 0. Each run is `completed/success` and has six successful jobs:

| Delivery | Run | Head SHA | Successful jobs |
| --- | --- | --- | --- |
| PR #1 implementation | [29213237436](https://github.com/relux-works/curator/actions/runs/29213237436) | `294aa8725cb6a3a9f3102896bfa3f418c7408e3e` | Interop golden gate, Lint, Naming gate, Ubuntu, macOS, Windows |
| PR #1 closure | [29213318140](https://github.com/relux-works/curator/actions/runs/29213318140) | `8f9b90a31ea04006680866fc05cb7d1f9a675d2a` | Interop golden gate, Lint, Naming gate, Ubuntu, macOS, Windows |
| PR #5 | [29270137737](https://github.com/relux-works/curator/actions/runs/29270137737) | `e423ce3f2838d48e9543e5c39e9e04428ee100d1` | Interop conformance gate, Lint, Naming gate, Ubuntu, macOS, Windows |
| PR #6 | [29263194544](https://github.com/relux-works/curator/actions/runs/29263194544) | `7adce172da7c480ba1f2e601e93312a0f18dbb39` | Interop conformance gate, Lint, Naming gate, Ubuntu, macOS, Windows |
| PR #7 | [29282453876](https://github.com/relux-works/curator/actions/runs/29282453876) | `a78545cd2ca02b025be1af72ed2fe4a5b5a9210d` | Interop conformance gate, Lint, Naming gate, Ubuntu, macOS, Windows |

No local product test suite was rerun because the task explicitly forbids product implementation and is limited to board metadata. The historical delivery-test claim is based on the directly queried GitHub CI records above.

## Backlog and scope verification

- `EPIC-260810-271m92` (`skill-facing-cli-adapters`) remains `backlog`.
- Its only child, `STORY-260810-rn6fg1` (`discover-skill-cli-language-inventory`), remains `backlog` with `task_class=research`.
- The adapter baseline is Swift, Kotlin, and C.
- Additional languages require discovery of the real skill/CLI inventory first. Rust, Python, Node/TypeScript, Dart, and .NET are candidates to evaluate, not commitments.
- The eventual adapter contract covers build/install, executable resolution, invocation/version probing, packaging/runtime expectations, and conformance tests.
- `EPIC-260713-c12fbe` (`dedicated-website`) remains `backlog` and has no children.

## Validation and recommendation

`task-board validate --json` ran as a standalone command after the repaired board state and evidence attachments were inspected. It exited 0 and returned:

```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

The board metadata matches the acceptance criteria and existing architecture: historical delivery is represented by evidence on the existing items, parent state is derived correctly, the adapter direction is isolated in backlog, and website work remains untouched. The recommendation is to hand `BUG-260810-2oxt8b` to review; the reviewer should accept only if the cited state and strict validation remain unchanged.

## References

- [Curator PR #1 — Correct Curator v0.1 conformance gaps](https://github.com/relux-works/curator/pull/1)
- [Curator PR #5 — Harden registry client production behavior](https://github.com/relux-works/curator/pull/5)
- [Curator PR #6 — Make command activation shell-neutral](https://github.com/relux-works/curator/pull/6)
- [Curator PR #7 — Implement seamless manager lifecycle](https://github.com/relux-works/curator/pull/7)
- [Task-board PR #4 — Add legacy status normalization](https://github.com/relux-works/skill-project-management/pull/4)
- [Closure evidence commit `8f9b90a...`](https://github.com/relux-works/curator/commit/8f9b90a31ea04006680866fc05cb7d1f9a675d2a)
- [Main patch-equivalent commit `3efdbca...`](https://github.com/relux-works/curator/commit/3efdbca5790baf0a04b9b9d3fc0fe9bfb46d0363)
