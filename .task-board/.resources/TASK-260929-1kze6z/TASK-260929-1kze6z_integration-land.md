# TASK-260929-1kze6z integration-land — producer preconditions confirmation (rev 1)

Binding: developer (implementer), run RUN-260929-cb91df, STORY-260929-3f6aym.
Revision 1 is accepted. Producer did NOT run `task-board worktree integrate`
(the runner performs the bound landing synchronously after this run exits),
so there is no integrate log and no refusal to report — nothing was attempted.
Board left at `integrating`; no status or handoff mutation by producer.
No file was changed by this run (read-only verification only).

## Observed preconditions

- Task TASK-260929-1kze6z status: `integrating`; story STORY-260929-3f6aym
  status: `integrating` (board query, this session).
- Worktree branch `task-board/story/STORY-260929-3f6aym`, HEAD `7f2fb6b8`
  (a board-state record commit, not a producer commit). Producer made zero
  commits on this branch.
- Uncommitted delta only (ready for the landing snapshot):
  - `M .github/ci/gate-selftest.sh`
  - `M .github/workflows/ci.yml`
  - `?? .github/ci/naming-gate.sh`

## Delta summary (uncommitted, accepted rev 1)

- `.github/workflows/ci.yml`: the inline "Employer name gate" block is
  replaced by one line calling `bash .github/ci/naming-gate.sh`, with a
  comment pointing at the script and its selftest rows.
- `.github/ci/naming-gate.sh` (new): both patterns assembled from parts at
  runtime; scans the tree while exempting only `literal N` / `delta N`,
  blank, and base85-shaped lines inside `GIT binary patch` blocks
  (block ends at next `diff --git ` line or EOF). All other lines —
  including non-base85 lines inside a block and text lines in `.patch`
  files — are still scanned.
- `.github/ci/gate-selftest.sh`: rows (a)–(e) with failure-reason match —
  binary-noise pass, patch-text fail, non-base85-in-block fail, full-name
  fail, clean pass.
- CHANGELOG.md: `Unreleased` carries Added/Changed/Fixed, no CI subsection,
  so the brief's conditional line was not triggered; unchanged.

## Local verification rerun by this run (real exit codes)

- `bash .github/ci/naming-gate.sh` on the current tree → exit 0.
- `bash .github/ci/gate-selftest.sh` (full suite) → 273 passed, 0 failed,
  exit 0; all 8 naming rows ok, including each fail row matching its
  naming-reason message.
- Runtime-assembled pattern scan of the tree: the full name occurs nowhere;
  the short name as a word occurs only inside the known 31gaka binary-patch
  resource (the exempted base85 noise); the gate's exit 0 confirms every hit
  there is exempted noise.

## Not rerun by this run

- Mutant probes (no-skip killed by row (a); no-shape-check killed by row
  (c)): accepted from the producer evidence already attached to the task's
  results resource. Replaying them would require editing files, which this
  integration run must not do.

Ready for the runner's bound landing; board remains `integrating` for the
integration transaction to write `done`.
