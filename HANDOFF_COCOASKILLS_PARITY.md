# CocoaSkills Go Parity Orchestration Handoff

Checkpoint date: 2026-07-31 (Asia/Tbilisi)

This file and branch preserve orchestration state. **Do not merge
`handoff/cocoaskills-parity-20260731` into `main`.** Product changes live in
their task branches and pull requests.

## Goal and operating constraints

- Codex session/goal ID: `019f4175-0d5c-7d43-8500-806f062aca9e`.
- Goal: finish CocoaSkills Go build-driver parity with Curator, including
  independent review, cross-platform CI, landing, and RC preparation.
- Codex is the orchestrator only. Use Claude Opus 5 child runs for
  implementation, testing, and independent review.
- Land accepted pull requests autonomously into `main`.
- Push CocoaSkills only to `ivanopcode/cocoaskills`.
- Do not create tags or GitHub Releases until the user explicitly commands it.
- Rust, Swift, and Kotlin implementation remain backlog. Their specification
  coverage is in scope, but only Go implementation parity is in the active goal.
- Prepare the CocoaSkills release candidate after all seven remaining delivery
  tasks are accepted and landed.
- The goal was marked `blocked` only for the host migration. Resume it on the
  new host before starting work.

All local child runs were intentionally cancelled for migration. A cancelled
run is not a review verdict. No implementation agent should still be running on
the old Mac.

## Repositories

Clone these exact remotes:

```bash
mkdir -p ~/Developer/ReluxWorks ~/Developer/Wildberries
cd ~/Developer/ReluxWorks
gh repo clone relux-works/curator
gh repo clone relux-works/curator-spec
cd ~/Developer/Wildberries
gh repo clone ivanopcode/cocoaskills
```

Use the checkpoint branch only as the orchestrator checkout:

```bash
cd ~/Developer/ReluxWorks/curator
git fetch origin --prune
git checkout handoff/cocoaskills-parity-20260731
```

Create task-scoped worktrees under `.temp/<TASK-ID>/worktree` for product work.
Never implement product changes directly on the checkpoint branch.

## Current pull requests

### Curator

- PR 12: <https://github.com/relux-works/curator/pull/12>, head
  `c7bc8900fb9901b1ea50c01646e1f2d84f5b6780`, task
  `BUG-260731-27h1yc`, status `reviewing`. Task-owned Windows package evidence
  is green, but the whole Windows job is red on packages owned elsewhere.
  Reviewer `RUN-260731-5000ae` was cancelled for migration. Spawn a fresh
  independent Opus reviewer and require an explicit scoped verdict.
- PR 13: <https://github.com/relux-works/curator/pull/13>, head
  `c6cb2dff29844aa1554b41f7f52f598130567b48`, task
  `BUG-260731-33v6zz`, status `development`. CI run `30630564858` is green on
  Ubuntu/macOS/lint/race and red on Windows. Download and classify the final
  Windows evidence, then route focused rework or review. The cancelled
  developer was `RUN-260731-6b26c6`.
- PR 14: <https://github.com/relux-works/curator/pull/14>, signed head
  `d345420109a9d043546d7cdb7b78a13d0bc19137`, task
  `BUG-260731-3a5q1p`, status `to-review`. Local rc.6 combined gate and focused
  install/race/lint/gofmt/vet passed. CI run `30631473821` has a red Windows
  Test job and was still finishing macOS when this checkpoint was written.
  Spawn an independent Opus reviewer after collecting current CI evidence.
- Comparison-only branch `task/BUG-260731-fs3dht-windows-goroot`, signed head
  `8aa5810`, preserves an alternative GOROOT hardening approach. Do not open or
  merge it unless the PR 13 reviewer needs a concrete comparison.

### CocoaSkills

- PR 16: <https://github.com/ivanopcode/cocoaskills/pull/16>, head
  `7a66c73ebb30822df40c87856f2301f1f409b735`, CI run `30624304158`.
  At checkpoint time Ubuntu and macOS (all Python versions) plus mypy were
  green; all four Windows cells had remained in progress for about two hours.
  Inspect their current state before rerunning anything.
- Task `BUG-260731-2rhy74` is deliberately blocked on whole-PR Windows green.
  Once PR 16 is green, independently reviewed, and landed, complete the marker
  fixture/spec landing and unblock `TASK-260720-g7kgox`.

### Curator specification

- PR 14: <https://github.com/relux-works/curator-spec/pull/14>, head
  `b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb`, 8/8 checks green.
- PR 15: <https://github.com/relux-works/curator-spec/pull/15>, head
  `2629aecff19a33e8cd1b5ebcfd898894ff1eeae0`, 8/8 checks green.
- Land only after the corresponding implementation pins and acceptance
  evidence are final. Curator-spec may be released when ready; do not tag it
  until the user gives the release command.

## Remaining CocoaSkills delivery path

Story `STORY-260720-1uv5gi` has 16 of 23 leaves done (69.6%). Seven delivery
tasks remain, in this dependency order:

1. `TASK-260720-g7kgox` - transactional global install
2. `TASK-260720-th0jdi` - build currentness, repair, and GC
3. `TASK-260720-12r55p` - shared v6 vector consumer
4. `TASK-260720-akf5kh` - schema v6 user docs (parallel with `12r55p`)
5. `TASK-260720-3pemm6` - cross-platform Go build E2E
6. `TASK-260720-3s27te` - integrated CocoaSkills v6 verification
7. `TASK-260730-2gtlzn` - publish CocoaSkills Go parity RC

The immediate critical path is: finish/review/land the Windows fixes and PR 16,
accept `BUG-260731-2rhy74`, then start `TASK-260720-g7kgox` with Opus 5.

## New-host bootstrap

The MacBook Air is `ivan-macbook-air-m1` / `100.78.9.45`, user `iv`.
GitHub CLI auth was transferred and verified for `ivanopcode`. Codex CLI
`0.146.0` is installed at `~/.local/bin/codex`.

Activate the persistent signing agent in every new shell:

```bash
export SSH_AUTH_SOCK="$(cat ~/.ssh/agent/cocoaskills-parity.sock)"
ssh-add -l
git config --global user.signingkey ~/.ssh/ivanopcode.pub
git config --global gpg.format ssh
git config --global commit.gpgsign true
```

The expected signing-key fingerprint is
`SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`. Verify a signed temporary
commit before publishing product work.

The migration copies these runtime artifacts to the new host:

- `~/.local/bin/task-board`
- `~/.agents/skills/project-management`
- `~/.agents/.instructions`
- the Codex session file and its session-index entry
- task-scoped dry-run evidence under `.temp/BUG-260731-3a5q1p/logs`

Validate the environment before spawning:

```bash
export PATH="$HOME/.local/bin:$PATH"
task-board q 'project_config()'
task-board q --format compact 'summary() { observability-project }'
gh auth status
git verify-commit HEAD
```

Inspect project spawn policy and model resolution before every first launch.
Spawn implementation and review with explicit Claude/Opus selection, preserve
producer -> reviewer -> rework loops, and land only an accepted task.

## Resume command

Run from the Curator checkpoint checkout on the MacBook Air:

```bash
cd ~/Developer/ReluxWorks/curator
export PATH="$HOME/.local/bin:$PATH"
export SSH_AUTH_SOCK="$(cat ~/.ssh/agent/cocoaskills-parity.sock)"
codex resume 019f4175-0d5c-7d43-8500-806f062aca9e \
  'Продолжаем цель «доводим cocoaskills до curator parity» после миграции на ivan-macbook-air-m1. Сними цель с blocked, прочитай HANDOFF_COCOASKILLS_PARITY.md, проверь GitHub CI/PR и task-board, затем перезапусти только Claude Opus 5 reviewer/developer runs по критическому пути. Все принятые PR приземляй автономно в main. Не создавай теги или GitHub Releases.'
```

First resumed actions:

1. Re-query PR 12, 13, 14 and CocoaSkills PR 16; do not trust the timestamped CI
   snapshot above when newer evidence exists.
2. Confirm `task-board agents()` has no live inherited runs.
3. Spawn fresh Opus review for PR 12 and PR 14, and route PR 13 from its latest
   Windows artifact.
4. Monitor the four long-running Windows cells on CocoaSkills PR 16. Diagnose
   rather than blindly rerunning if they are still stuck.
5. Land accepted Curator/CocoaSkills/spec work in dependency order, update the
   board, and start the seven-task delivery chain.

## Checkpoint caveats

- The checkpoint commit contains historical board resources and evidence. Some
  imported evidence has trailing whitespace and patch payloads; this makes
  `git diff --check` noisy on the checkpoint branch. Product branches must still
  pass normal formatting and diff checks.
- Raw `*_spawn-log_*` files were intentionally excluded from the checkpoint
  commit because they are large runtime streams, including one file above
  50 MiB. Structured board state, notes, outcomes, and task artifacts are kept.
- The old local worktrees are not required for continuation because all useful
  product commits are pushed. Recreate worktrees from the published branches.
