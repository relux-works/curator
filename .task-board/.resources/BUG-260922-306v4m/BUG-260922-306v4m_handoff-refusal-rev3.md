# BUG-260922-306v4m handoff refusal — rev3 (exact refusal, work stopped per republish protocol)

`task-board handoff BUG-260922-306v4m --role developer` (RUN-260923-b4c8b5)
was REFUSED. Per `republish-tree-bound-evidence.md` ("If the handoff
refuses for any reason, attach the exact refusal as an outcome resource
and stop — do not edit the checklist, do not check items by proxy, do
not change code"), no checklist item was touched and no file was changed.
The refusal text below is verbatim (reason line, samples window, and all
35 `[VIOLATION]` paths; the 20333 `[reported]` tail lines name other runs
and foreign worktrees and are omitted as non-violations).

```text
run_wrote_outside_worktree: run RUN-260923-b4c8b5 element BUG-260922-306v4m policy warn violated=true
samples: 2026-09-23T12:08:26.544846Z .. 2026-09-23T12:26:14.228849Z
paths:
  [VIOLATION] .task-board/.activity/BUG-260922-6chzf9/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/BUG-260923-11jgkt/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/BUG-260923-3mazfw/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/STORY-260906-1a2i5a/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/STORY-260922-188t6n/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/STORY-260923-3vwgy4/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/STORY-260923-laeycm/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/STORY-260923-vkxt08/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/TASK-260906-2b3nar/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/TASK-260907-187z6x/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/TASK-260908-1bfk8y/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/TASK-260922-1ejkxv/events.ndjson (unattributed)
  [VIOLATION] .task-board/.activity/TASK-260923-xq4pjj/events.ndjson (unattributed)
  [VIOLATION] .task-board/.element-move-journal.json (unattributed)
  [VIOLATION] .task-board/.resources/BUG-260923-11jgkt/campaign-producer-rules.md (unattributed)
  [VIOLATION] .task-board/.resources/BUG-260923-11jgkt/flake-brief.md (unattributed)
  [VIOLATION] .task-board/.resources/BUG-260923-3mazfw/campaign-producer-rules.md (unattributed)
  [VIOLATION] .task-board/.resources/BUG-260923-3mazfw/flake-brief.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260905-2qvzwk_stage-a-acquisition-byte-exact/TASK-260906-2b3nar_fold-gate-directory-components/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260906-1a2i5a_stage-c-composition-path-kind-and-import/TASK-260907-187z6x_git-same-source-reinstall-drops-use-and-takeover/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260907-2bddfc_promote-the-released-suite-pin/TASK-260908-1bfk8y_stale-pin-comment-in-platform-exclusions/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260922-188t6n_0017-0018-spec-follow-ups/TASK-260922-1ejkxv_fleet-isolated-policy/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260908-2wp8wn_curator-run-and-infra-migration/STORY-260915-3w11un_workstation-and-checkout-readiness/TASK-260923-xq4pjj_adopt-hand-written-task-board-shim/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-3vwgy4_ci-flake-macos-git-eacces/BUG-260923-3mazfw_hosted-macos-git-spawn-eacces-recurs/README.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-3vwgy4_ci-flake-macos-git-eacces/BUG-260923-3mazfw_hosted-macos-git-spawn-eacces-recurs/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-3vwgy4_ci-flake-macos-git-eacces/README.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-3vwgy4_ci-flake-macos-git-eacces/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-laeycm_ci-flake-windows-snapshot-concurrency/BUG-260923-11jgkt_windows-snapshot-concurrent-get-sharing-violation/README.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-laeycm_ci-flake-windows-snapshot-concurrency/BUG-260923-11jgkt_windows-snapshot-concurrent-get-sharing-violation/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-laeycm_ci-flake-windows-snapshot-concurrency/README.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-laeycm_ci-flake-windows-snapshot-concurrency/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-vkxt08_ci-flake-fixes/BUG-260922-6chzf9_windows-managerlock-tiny-deadline-flake/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-vkxt08_ci-flake-fixes/BUG-260923-3mazfw_hosted-macos-git-spawn-eacces-recurs/README.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-vkxt08_ci-flake-fixes/BUG-260923-3mazfw_hosted-macos-git-spawn-eacces-recurs/progress.md (unattributed)
  [VIOLATION] .task-board/EPIC-260905-29w4hn_agent-environments-execution/STORY-260923-vkxt08_ci-flake-fixes/progress.md (unattributed)
```

## Facts of this run (for the operator clearing the violation)

- This run's only control-root writes were attributed board mutations:
  `set_status(BUG-260922-306v4m, status=development)` (activity seq 52)
  and `resource add BUG-260922-306v4m_republish-rev3.md` (activity seq 53).
  No board file was edited directly; all scratch went to `/tmp`.
- Every `[VIOLATION]` path above belongs to a different element, is marked
  `(unattributed)`, and falls inside this run's wall window, during which
  dozens of concurrent campaign runs were mutating the board on this host
  (observed: concurrent `task-board handoff` / `spawn wait` processes for
  other tasks, e.g. TASK-260922-cww1ov, TASK-260923-1fgsrb,
  BUG-260923-11jgkt, all long-running simultaneously).
- The worktree diff is byte-identical to the accepted rev1/rev2 patch
  (`cmp` clean; 4 files: `.github/ci/gate-selftest.sh`,
  `.github/ci/install-rust-toolchain.sh`, `CHANGELOG.md`,
  `docs/self-hosted-runner-setup.md`); no code changed in this run.
- Fresh validation on this tree: `bash -n` passes on both scripts;
  `bash .github/ci/gate-selftest.sh` → 224 passed, 0 failed, exit 0.
- Board state now: status `to-review` (activity seq 54, the handoff's
  status transition persisted); no CR revision 3 was published.
- Exact input needed: operator clearance of the `run_wrote_outside_worktree`
  warn on RUN-260923-b4c8b5 (or a quiet-window re-handoff), then reviewer
  launch / checkpoint / integration can proceed.
