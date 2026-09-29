# TASK-260929-2wgjam integration preconditions (bound developer run, curator)

Revision 2 ACCEPTED per integration instruction binding (`CR-TASK-260929-2wgjam-2` rev 2).
Landing is the runner's synchronous step after this producer exits; this run performs
no landing itself.

Preconditions confirmed read-only (exit 0 on each check):

- Board: `TASK-260929-2wgjam` status=`integrating`; `STORY-260929-1s4r14` status=`integrating` (via `task-board q get`, exit 0).
- Worktree branch: `task-board/story/STORY-260929-1s4r14`, HEAD `64b12189f131d58dada1999e7a930de09e634e8a` (no producer commit on top of checkpoint).
- Working tree: exactly 3 uncommitted modifications, no untracked additions by this run:
  - `M .github/ci/gate-selftest.sh`
  - `M .github/ci/install-rust-toolchain.sh`
  - `M docs/self-hosted-runner-setup.md`
- No `set_status`, no `handoff`, no `worktree checkpoint`/`worktree integrate` executed by this run (per binding: runner lands synchronously; `integrate-land.md` refusal path not triggered because landing was not attempted here).
- No repo file changed by this run; no gate re-run here — gate evidence remains with the accepted revision and the runner's gate run.

Handed off to review via the bound runner landing (board stays `integrating`; only the integration transaction may write `done`).
