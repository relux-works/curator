# Landed delivery evidence reconciliation

## Context

Board task: `BUG-260810-2oxt8b` (`reconcile-landed-delivery-evidence`).

This audit reconciles historical Curator delivery metadata with work already present on `origin/main`. It covers PRs #5, #6, and #7, the closed PR #1 conformance branch, exact reviewer status recommendations, parent aggregation, and a separate adapter backlog direction. It intentionally makes no product-code changes and does not perform the historical terminal transitions; those belong to the reviewer verdict.

Evidence was collected on 2026-08-10 with read-only `gh` and `git` commands. `git ls-remote` resolved remote `main` to `124654a051f60995e826b75ab43a3c4f3359f5ff` and `agent/curator-conformance-review` to `8f9b90a31ea04006680866fc05cb7d1f9a675d2a`; both matched the local remote-tracking refs.

## Key findings

1. PRs [#5](https://github.com/relux-works/curator/pull/5), [#6](https://github.com/relux-works/curator/pull/6), and [#7](https://github.com/relux-works/curator/pull/7) are merged. Each PR head has one completed successful CI run with the same six required jobs: interoperability/conformance, lint, naming, Ubuntu, macOS, and Windows. Each merge commit is an ancestor of current `origin/main`.
2. PR [#1](https://github.com/relux-works/curator/pull/1) is closed without merge, but its implementation is already on `origin/main`: history through `d60f8d0...` is a literal ancestor, and the remaining code commit `294aa87...` has the same stable patch ID as main commit `3efdbca...`. The only non-equivalent branch commit is `8f9b90a...`, which changes four board-metadata files to record CI closure and historical statuses; it changes no product code.
3. The terminal transitions should remain reviewer-owned. Existing leaf Tasks `TASK-260713-c7a18d` and `TASK-260713-7a9c1e` have producer evidence and can move from `reviewing` to `done` after review. Legacy empty Stories `STORY-260713-b4e219` and `STORY-260713-72b914` map directly to merged PR delivery; no historical implementation task should be invented.
4. A separate backlog Epic and discovery Story now capture Swift, Kotlin, and C as the baseline, evidence-based discovery before adding languages, and Rust/Python/Node-TypeScript/Dart/.NET as candidates only. The dedicated website Epic remains untouched in `backlog`.
5. One board-storage anomaly remains important for review: `STORY-260713-72b914` is projected by structured queries as `backlog`, but its legacy `progress.md` stores `in-progress`. Current `task-board` 0.24.3 rejects that raw value on resource or status mutation even though `task-board validate` reports the board valid. The PR #5 evidence therefore lives in this task-scoped outcome rather than on that historical Story. No direct board-file edit was made.

## Evidence matrix and exact reviewer transitions

| Board item | Current structured status | Delivery and CI evidence | Main-line or equivalence evidence | Exact recommendation | Expected aggregation |
| --- | --- | --- | --- | --- | --- |
| `TASK-260713-c7a18d` (`harden-shell-activation`) | `reviewing` | PR [#6](https://github.com/relux-works/curator/pull/6) merged `2026-07-13T16:58:27Z`; head `7adce172da7c480ba1f2e601e93312a0f18dbb39`; [CI 29263194544](https://github.com/relux-works/curator/actions/runs/29263194544) completed successfully with all six jobs | Merge `7cde9b9bcd9fae26ddd2af3dd29aa380d0c4e12e`; ancestor check against `origin/main` exited 0 | Reviewer: `reviewing` → `done` | Only-child `STORY-260713-9f4c2b` auto-aggregates to `done`; `EPIC-260712-785993` becomes `done` after `STORY-260713-b4e219` is also accepted |
| `STORY-260713-b4e219` (`seamless-manager-lifecycle`) | `backlog` | PR [#7](https://github.com/relux-works/curator/pull/7) merged `2026-07-13T20:29:52Z`; head `a78545cd2ca02b025be1af72ed2fe4a5b5a9210d`; [CI 29282453876](https://github.com/relux-works/curator/actions/runs/29282453876) completed successfully with all six jobs | Merge `20fd90c5fbc113331e5cd365ce48c507c912dafd`; ancestor check against `origin/main` exited 0 | Reviewer: legacy empty Story `backlog` → `done`; do not invent a leaf task | `EPIC-260712-785993` becomes `done` once this Story and `STORY-260713-9f4c2b` are both terminal |
| `STORY-260713-72b914` (`production-profile-conformance`) | projected `backlog`; raw legacy value `in-progress` | PR [#5](https://github.com/relux-works/curator/pull/5) merged `2026-07-13T17:26:03Z`; head `e423ce3f2838d48e9543e5c39e9e04428ee100d1`; [CI 29270137737](https://github.com/relux-works/curator/actions/runs/29270137737) completed successfully with all six jobs | Merge `75e802aa6dd65757d4763490d3950470fb49b150`; ancestor check against `origin/main` exited 0 | Reviewer target: legacy Story → `done` after a task-board-supported normalization of `in-progress`; do not edit the board file directly or fabricate a leaf task | Only-child `EPIC-260712-d77d32` auto-aggregates to `done` |
| `TASK-260713-7a9c1e` (`review-and-correct`) | `reviewing` | PR [#1](https://github.com/relux-works/curator/pull/1) closed without merge; [CI 29213237436](https://github.com/relux-works/curator/actions/runs/29213237436) succeeded at implementation head `294aa87...`; [CI 29213318140](https://github.com/relux-works/curator/actions/runs/29213318140) succeeded at closure head `8f9b90a...`; both have all six jobs | Merge base `d60f8d065dee6461512637f0c85c34c369574fa1` is an ancestor of main. `294aa87...` and main `3efdbca...` share stable patch ID `fdba908b17a82392386c79dfe25ae3ac79bbf2e5`. Unique `8f9b90a...` is metadata-only closure evidence | Reviewer: `reviewing` → `done` | `STORY-260713-d0d4e8` and `EPIC-260712-c8ac0f` auto-aggregate to `done` only after sibling `BUG-260810-2oxt8b` is also reviewed and accepted |
| `BUG-260810-2oxt8b` (`reconcile-landed-delivery-evidence`) | `analysis` during production | This research document plus task-scoped outcome and board evidence | No product delivery claim; metadata/research scope only | Producer handoff: `analysis` → `to-review`; reviewer: `reviewing` → `done` if accepted | Together with `TASK-260713-7a9c1e`, closes the remaining active children of `STORY-260713-d0d4e8` |

The legacy empty-Story recommendations are compatible with `task-board.config.json`, which explicitly sets `validation.allow_legacy_empty_done_containers=true`. PR-head tree inspection confirms that PR #5 and PR #7 created only each Story's README/progress files, not leaf Tasks.

## Closed conformance branch fact-check

The following independent checks converge on the same conclusion:

- `gh pr view 1 ...` exited 0 and returned `state=CLOSED`, `mergedAt=null`, `headRefOid=8f9b90a31ea04006680866fc05cb7d1f9a675d2a`.
- `git merge-base origin/main origin/agent/curator-conformance-review` exited 0 and returned `d60f8d065dee6461512637f0c85c34c369574fa1`.
- `git merge-base --is-ancestor d60f8d065dee6461512637f0c85c34c369574fa1 origin/main` exited 0.
- `git cherry -v origin/main origin/agent/curator-conformance-review` exited 0 and returned exactly one equivalent commit (`- 294aa87...`) plus one unique commit (`+ 8f9b90a...`).
- Patch-ID commands used `zsh -o pipefail`; both exited 0 and returned `fdba908b17a82392386c79dfe25ae3ac79bbf2e5` for branch `294aa87...` and main `3efdbca...`.
- `git diff-tree --no-commit-id --name-status -r 8f9b90a...` exited 0 and listed only four `.task-board` paths. The commit message says it records the green six-check run and closure statuses.
- The implementation-head and closure-head GitHub Actions runs both returned `completed/success`, with all six named jobs successful.

Inference: closure is justified by landed code and patch equivalence, but the unique closure commit itself should not be treated as a missing product patch or merged solely to copy historical status metadata.

## Adapter backlog direction

Created without starting implementation:

- `EPIC-260810-271m92` (`skill-facing-cli-adapters`) — `backlog`.
- `STORY-260810-rn6fg1` (`discover-skill-cli-language-inventory`) — `backlog`, `task_class=research`.

The baseline languages are Swift, Kotlin, and C. Discovery must inspect the actual skill/CLI inventory—skills, manifests, build metadata, launchers, packaging, and runtime requirements—before adding more languages. Rust, Python, Node/TypeScript, Dart, and .NET are evaluation candidates, not commitments.

The eventual adapter contract must cover:

- build and installation;
- executable resolution;
- invocation and version probing;
- packaging and runtime expectations;
- conformance tests.

`EPIC-260713-c12fbe` (`dedicated-website`) remains `backlog`, has no new children, and received no mutation.

## Board anomaly / logbook entry

Attaching PR #5 evidence directly to `STORY-260713-72b914` failed with exit 1:

```text
reading progress: parsing status: unknown status: in-progress
```

An attempted `set_status(STORY-260713-72b914, status=backlog)` also failed with exit 1 for the same reason. Structured `get(...)` nevertheless projects the item as `backlog` with an empty checklist, and `task-board validate` exits 0. Raw read-only inspection shows the historical checklist still contains three checked items and one unchecked OS-matrix item. PR #5 CI supplies the missing OS-matrix evidence, but the checklist/status cannot be mutated through this CLI until legacy `in-progress` is normalized by a supported task-board path.

Recommended handling: preserve this finding in review, avoid direct `.task-board` edits, and use a task-board release or explicit CLI migration that can atomically normalize the legacy value before applying the reviewer verdict. The PR #5 evidence remains available in this task-scoped outcome.

## Validation record

- Initial `task-board validate`: exit 0, `Board is valid. No issues found.`
- Post-reconciliation validation after the adapter backlog and historical evidence mutations: exit 0, `Board is valid. No issues found.`
- A final strict validation is required after outcome attachment and checklist updates, immediately before the researcher handoff; its real exit code is reported in the handoff summary.
- No product build/test commands were run because the assignment explicitly limits work to Curator board metadata and read-only delivery evidence.

## References

- [PR #1 — Correct Curator v0.1 conformance gaps](https://github.com/relux-works/curator/pull/1)
- [PR #1 implementation-head CI](https://github.com/relux-works/curator/actions/runs/29213237436)
- [PR #1 closure-head CI](https://github.com/relux-works/curator/actions/runs/29213318140)
- [Closure evidence commit `8f9b90a...`](https://github.com/relux-works/curator/commit/8f9b90a31ea04006680866fc05cb7d1f9a675d2a)
- [Main patch-equivalent commit `3efdbca...`](https://github.com/relux-works/curator/commit/3efdbca5790baf0a04b9b9d3fc0fe9bfb46d0363)
- [PR #5 — Harden registry client production behavior](https://github.com/relux-works/curator/pull/5)
- [PR #5 CI](https://github.com/relux-works/curator/actions/runs/29270137737)
- [PR #6 — Make command activation shell-neutral](https://github.com/relux-works/curator/pull/6)
- [PR #6 CI](https://github.com/relux-works/curator/actions/runs/29263194544)
- [PR #7 — Implement seamless manager lifecycle](https://github.com/relux-works/curator/pull/7)
- [PR #7 CI](https://github.com/relux-works/curator/actions/runs/29282453876)
