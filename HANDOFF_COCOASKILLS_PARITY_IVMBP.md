# CocoaSkills Go parity handoff to ivmbp

## Objective

Complete CocoaSkills Go compiled-build parity with Curator. Finish all remaining
tracked delivery tasks and CI defects, obtain independent review, land accepted
Curator, curator-spec, and ivanopcode/cocoaskills changes to `main`, validate
macOS/Windows/Linux CI, and prepare the CocoaSkills release candidate.

Do not create or push tags and do not create GitHub Releases until the user gives
an explicit release command.

The interactive Codex session is the orchestrator. Delegated work may use both:

- Claude Opus 5: `--agent claude --model claude-opus-5`
- Codex Sol Medium: `--agent codex --model gpt-5.6-sol --reasoning-effort medium`

The board primary goal is active at `PRIMARY-GOAL-260731-v5a1is` revision 4 and
records this two-provider policy.

## Remote workspace

Connect with `ssh -A ivmbp`; the forwarded SSH agent currently provides GitHub
Git access. Use these prepared clean paths:

- Orchestration board: `/Users/iv/Developer/ReluxWorks/curator-parity`
- Protocol spec: `/Users/iv/Developer/ReluxWorks/curator-spec-parity`
- CocoaSkills canonical clone: `/Users/iv/Developer/Wildberries/cocoaskills`
- PR 19 task worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-12r55p/worktree`

The PR worktree uses local branch
`task/TASK-260720-12r55p-ivmbp-20260802`, tracking the existing remote PR branch
`task/TASK-260720-12r55p-shared-v6-vectors`. Push accepted rework explicitly:

```bash
git push origin HEAD:task/TASK-260720-12r55p-shared-v6-vectors
```

Do not clean, reset, or reuse these older dirty worktrees:

- `/Users/iv/Developer/ReluxWorks/curator`
- `/Users/iv/Developer/ReluxWorks/curator-spec`

They contain the previous migration branch and untracked historical board logs.

## Current delivery status

- Strict board metric: 14 of 17 Go parity delivery tasks are accepted `done`.
- `TASK-260720-12r55p` is `to-dev` after cycle-3 review.
- CocoaSkills PR: <https://github.com/ivanopcode/cocoaskills/pull/19>
- Exact reviewed signed head: `6e7742f0d28ad95ddd7d8e92364b84062571ad0b`
- Hosted CI run: `30743353816`
- Ubuntu 3.11-3.14, macOS 3.11-3.14, and strict mypy are green.
- Windows Python 3.11 failed at migration close; Windows Python 3.12-3.14 were
  still running. PR 19 was `UNSTABLE` at exact head `6e7742f`.
- The current head must not be merged even if those jobs later appear green;
  two independent reviewers requested changes from static reachable evidence.

After `TASK-260720-12r55p`, the critical path is:

1. `TASK-260720-3pemm6` cross-platform Go build E2E.
2. `TASK-260720-3s27te` integrated CocoaSkills schema-v6/Go verification.
3. `TASK-260730-2gtlzn` RC preparation only, with no tag or GitHub Release.

## Immediate blocking fix

The helper `_utime_portably()` added at PR head `6e7742f` repaired only one of
five Windows-unsupported calls. Four reachable calls remain in
`tests/protocol_lifecycle_observations.py` in `_observe_gc`:

- line 3055: `os.utime(rejected_entry, (1, 1), follow_symlinks=False)`
- line 3063: `os.utime(entry, (young_mtime, young_mtime), follow_symlinks=False)`
- line 3071: `os.utime(entry, (2, 2), follow_symlinks=False)`
- line 3112: `os.utime(other_entry, (1, 1), follow_symlinks=False)`

Required rework:

1. Route all four calls through `_utime_portably()`.
2. Add a source-level regression that rejects `follow_symlinks=False` outside
   the body of `_utime_portably()` in this observer module.
3. Run focused fallback plus representative GC vectors, strict mypy, build,
   Twine, signature/diff/no-workflow-drift guards, and the exact candidate suite.
4. Commit with a valid signature and push the existing PR branch.
5. Require terminal-green exact-head hosted Windows CI and a fresh independent
   reviewer verdict before landing PR 19.

Authoritative review resources on the task board:

- `TASK-260720-12r55p_review-verdict-cycle-3.md`
- `BUG-260802-1s021p_pr19-independent-review.md`

Both independently identify the same four reachable call sites. The earlier
provisional accept in the Opus review was superseded by its final
`CHANGES REQUESTED` verdict.

## Prepared downstream runbooks

The E2E preparation is accepted `done`:

- Board item: `BUG-260802-3ibgu1`
- Outcome: `BUG-260802-3ibgu1_csk-go-e2e-readiness-audit.md`
- Review: `BUG-260802-3ibgu1_review-verdict.md`
- Working copy: `.research/260802_csk-go-e2e-readiness-audit.md`

The integrated verification preparation is `to-review`:

- Board item: `BUG-260802-3fbn47`
- Outcome: `BUG-260802-3fbn47_csk-integrated-verification-readiness-audit.md`
- Working copy: `.research/260802_csk-integrated-verification-readiness-audit.md`

Review `BUG-260802-3fbn47` before using its runbook. Do not start the blocked
delivery task until `TASK-260720-3pemm6` is accepted and landed.

Candidate provenance remains:

- curator-spec commit: `432eb2ee1fe2d6b271e37269f867c8851c325539`
- manifest SHA-256:
  `12e58b82579645ba1ccafba49d3e2dd3216005ddf37ae63c68a9fafd46773071`
- committed CocoaSkills suite pin remains
  `0c81c1f8d5321d822be2a2817b05aea03e656e15`
- the two candidate trees are byte-equivalent; do not advance the release pin
  in the current delivery tasks.

## First commands on ivmbp

```bash
ssh -A ivmbp
export PATH="$HOME/.local/bin:$PATH"
cd /Users/iv/Developer/ReluxWorks/curator-parity

codex login status || codex login
claude auth status
gh auth status
task-board q --format compact \
  'get(TASK-260720-12r55p) { observability-task }; agents(stale=15)'

cd /Users/iv/Developer/Wildberries/cocoaskills
git fetch --prune origin
git status --short --branch

cd .temp/TASK-260720-12r55p/worktree
git status --short --branch
git rev-parse HEAD
git log -1 --show-signature
gh pr view 19 --json headRefOid,state,mergeStateStatus,statusCheckRollup
```

At migration time `codex login status`, `claude auth status`, and `gh auth
status` reported no active local login on `ivmbp`. Authenticate Codex before
resuming this session and authenticate Claude before delegating Opus runs.
`gh auth login` is needed for `gh pr`/CI operations; Git fetch and push can use
the forwarded SSH agent from `ssh -A ivmbp`. Do not copy credential files from
the source Mac.

The first delegated run should be a developer on `TASK-260720-12r55p`, using
either allowed pool. Give it the cycle-3 verdict resource and require it to keep
the existing worktree/branch and PR rather than creating a replacement PR.

After the new signed head is pushed, spawn a fresh independent reviewer. Land
the task and PR only after exact-head CI is terminal green and the reviewer marks
the board task `done`. Then fast-forward CocoaSkills `main`, remove only the
finished task worktree when it is safe, and start `TASK-260720-3pemm6` from the
accepted E2E runbook.

## Session resume

The current Codex thread is:

`019f4175-0d5c-7d43-8500-806f062aca9e`

Its session JSONL is copied to the matching path under
`/Users/iv/.codex/sessions/2026/07/08/` on `ivmbp`, with the previous remote
copy retained as a timestamped backup. Resume from the clean board worktree:

```bash
cd /Users/iv/Developer/ReluxWorks/curator-parity
codex resume 019f4175-0d5c-7d43-8500-806f062aca9e \
  "Continue CocoaSkills Go parity from HANDOFF_COCOASKILLS_PARITY_IVMBP.md. Verify current remote state first, then resume TASK-260720-12r55p rework through review, CI, landing, E2E, integrated verification, and RC preparation. Use Claude Opus 5 and Codex gpt-5.6-sol/medium for delegated work. Do not create tags or GitHub Releases."
```

If direct session resume fails, start a new Codex session from the same directory
with the quoted prompt above and treat this handoff file plus the task board as
authoritative.
