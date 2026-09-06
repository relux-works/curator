# Review brief: stage (a) core — cycle 6 (F16; acceptance cycle)

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `834b40f6` (on the cycle-5 head `b6f00e1a`). Cycle-5 findings:
`TASK-260905-30zs8t_review-findings-stage-a-5.md`; author decision (option (a), machine pass skips
adapters carrying a scope record): `producer-brief-stage-a-rework-5.md`; rework report:
`TASK-260905-30zs8t_rework-report-5.md`.

1. **F16** — reproduce your own cycle-5 script: `env:codex_cli` scoped to profile B, machine current
   on A, machine-scope `profile use C` — the codex home stays on B, the other adapters move to C, and
   `profile list` reports exactly that; the same through `profile install <src> --use`; a scoped
   `profile use --env codex_cli D` still switches only that home. Run the narrowing mutant (machine
   pass ignores the scope record) and confirm a named test dies. Check `SyncWithPolicy` and
   `profile update` for the same shape — a scope record must not be overwritten there either.
2. **Regression** — the cumulative "verified, held" set from cycles 1–5, at least: F1 directory-addressed
   detector, F2 fresh-home default, F3 interrupted switch, F8 spellings, F9 gates with `audit.enabled`
   false, F10 system-locked migration, F12 `file://` rejection, F13 partial-scope activation, F14
   syntactic classification. This rework touched the switch path, which every one of those exercises.
3. **Gates** — the full set, plus `bash .github/ci/gate-selftest.sh`, the platform-case gate for the
   three GOOS values, and the vector families through `CURATOR_CONFORMANCE_ROOT`.
4. Commits signed by the repository's human identity.

This is the acceptance cycle. Block only on something that would ship a wrong artifact, bypass a gate,
or leave recorded state contradicting the bytes; everything smaller is a follow-up on
TASK-260906-1f2ng0, not another round. Read-only (scratch under the worktree's `.temp/`). Never write
into the control root. Findings resource `TASK-260905-30zs8t_review-findings-stage-a-6.md`.
Blocking/major → `development`; else explicit ACCEPT at `to-review` with `accept_cr`. Do not mark
done. `task-board handoff TASK-260905-30zs8t --role reviewer`.
