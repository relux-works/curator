# Review brief: stage (b) — cycle 2 (rework of F1–F10)

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, branch
`feat/agent-environments-stage-b`, head `7bca4d4b`; rebased onto curator main (which now carries stage (a)
as `981b1eeb`) and five rework commits on the cycle-1 head. PR https://github.com/relux-works/curator/pull/60.
Cycle-1 findings: `TASK-260906-2g0bgq_review-findings-stage-b-1.md`; author decisions:
`producer-brief-stage-b-rework-1.md`; rework report: `TASK-260906-2g0bgq_rework-report-1.md`.

Verify by driving, not reading — your own scripts, as in cycle 1:
1. **F1** — on a simulated macOS below the pinned release, `shared` is admitted and a managed
   `claude_code` home provisions; `isolated` is still refused; at or above the pin both behave as §7.4
   says. Run the narrowing mutant (refuse `shared` on darwin regardless of the pin) and confirm the new
   matrix rows kill it.
2. **F2** — write through a link into every manager-authored surface (root context, system prompt, each
   MCP file) for every adapter and confirm resolve reports the home stale and emits no fragment; confirm
   the link-identity fast path still applies to targets under `contextstore.Root`; run the mutant that
   keeps the link check and drops the byte check.
3. **F3** — `golangci-lint run ./...` is 0 issues on this head, and the rework report's gate lines are
   the commands' own output.
4. **F4–F6** — the boundary case below the environments root's parent; the always-copied claude_code
   root context asserted under **both** forms with its recorded reason; the referenced-form staleness
   check testing the approval key itself, with its negative case.
5. **F7–F10** — the second §7.6 secondary target, the unreadable `config.toml` failing rather than
   defaulting to `file`, the three §12 `env status` rows, and `--format env|shell` emitting exactly the
   fragment's `env` names.
6. **Regression** — the cycle-1 "verified, held" set: the conformance subset byte-exact with zero skips,
   the fragment well-formed for every adapter and both forms, the marker contents, the §10.3 boundary.
   The rework touched the read path and the registry, so re-drive them.
7. **Gates** — the full set including `bash .github/ci/gate-selftest.sh`, the platform-case gate for the
   three GOOS values, and the vector families; then check `gh pr checks 60`. Stage (a) shipped a
   Windows-only fixture defect (native paths in git config values) that no Unix lane could see — look
   for that class in every fixture this stage added before the hosted lane does.
8. Signed commits by the repository's human identity; scope limited to stage (b).

Read-only (scratch under the worktree's `.temp/`). Never write into the control root. Findings resource
`TASK-260906-2g0bgq_review-findings-stage-b-2.md`. Blocking/major → `development`; else explicit ACCEPT
at `to-review` with `accept_cr`. Do not mark done.
`task-board handoff TASK-260906-2g0bgq --role reviewer`.
