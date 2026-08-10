# Reconciliation brief

Work only on Curator board metadata. Do not edit product code, commit, push, open or merge pull requests.

## Delivery evidence to verify

- PR #6, `Make command activation shell-neutral`: https://github.com/relux-works/curator/pull/6 — merged; expected mapping: `TASK-260713-c7a18d` can move from `reviewing` to `done` after evidence review.
- PR #7, manager lifecycle delivery: https://github.com/relux-works/curator/pull/7 — merged; expected mapping: `STORY-260713-b4e219` can move from `backlog` to `done` after evidence review.
- PR #5, production profile conformance: https://github.com/relux-works/curator/pull/5 — merged; expected mapping: `STORY-260713-72b914` can move from `backlog` to `done` after evidence review.
- PR #1, `Correct Curator v0.1 conformance gaps`: https://github.com/relux-works/curator/pull/1 — closed without merge, with the historical six-platform/check suite green. Its branch is `origin/agent/curator-conformance-review`. Current `origin/main` already contains the implementation patches by patch equivalence; the only substantive unique branch commit previously observed was `8f9b90a31ea04006680866fc05cb7d1f9a675d2a`, which records closure evidence. Re-verify this before recommending `TASK-260713-7a9c1e` as `done`.

Use read-only `gh` and `git` inspection to verify current facts. Record URLs, relevant commit IDs, check conclusions, and any inference explicitly. Never fabricate a historical implementation task.

## Board outcome

Attach a task-scoped outcome to `BUG-260810-2oxt8b` with an evidence matrix and exact recommended transitions. Add/update evidence resources on the historical board items when useful. Do not make their terminal transitions; leave those for the reviewer verdict.

Capture, but do not start, a separate backlog direction for skill-facing CLI adapters:

- baseline languages: Swift, Kotlin, and C;
- define discovery of the real skill/CLI inventory before adding more languages;
- candidates to evaluate, not pre-commit: Rust, Python, Node/TypeScript, Dart, and .NET;
- adapter contract should eventually cover build/install, executable resolution, invocation/version probing, packaging/runtime expectations, and conformance tests.

Prefer one backlog Epic plus one discovery Story. Leave `EPIC-260713-c12fbe` (`dedicated-website`) in `backlog` and do not create website implementation tasks.

Validate the board before the producer handoff. Keep `.task-board/.goals/` and `.temp/` out of git commits.
