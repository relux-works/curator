# BUG-260922-306v4m integration check — fresh outcome evidence (RUN-260923-6fc54f, 2026-09-23)

Integration run for accepted Change Request `CR-BUG-260922-306v4m-3` revision 3 (role `developer`,
archetype `implementer`). Verify-only: **no file was edited in this run**. The tree is left
uncommitted on the Story branch for the bound landing. Status left at `integrating`; no `handoff`
and no status change issued in this run, per the integration assignment.

## 1. Landing preconditions confirmed

- Board: `task-board q 'get(BUG-260922-306v4m)'` → `status=integrating` (this session; the opening
  `set_status(..., status=integrating)` was a no-op `integrating`→`integrating`).
- Worktree: branch `task-board/story/STORY-260915-3w11un`, HEAD `48da2690`; `git status --short`
  shows exactly the accepted 4-file delta, all uncommitted, no untracked files added by this run:
  - `M .github/ci/install-rust-toolchain.sh`
  - `M .github/ci/gate-selftest.sh`
  - `M CHANGELOG.md`
  - `M docs/self-hosted-runner-setup.md`
  - `git diff --stat`: 236 insertions, 2 deletions across those 4 files.
- No commits made in this run (`git rev-parse --short HEAD` still `48da2690`).
- Scope respected: failure-path diagnostics only; success-path lines untouched
  (`GITHUB_PATH` writes at install script lines 48–50 and 139–142, `rust-pin: using rustup at`
  line 143, resolution loop lines 55–69 — all identical apart from the inserted diagnostics
  function + guarded call).
- Docs paragraph present (`docs/self-hosted-runner-setup.md:37`, "diagnostics block before the
  remedy note"); CHANGELOG entry present (`CHANGELOG.md:132`, "runner diagnostics block").
- No secrets in the block: only `RUNNER_NAME`, `RUNNER_OS`, `hostname`, `whoami`, `HOME`,
  `CARGO_HOME` (set/defaulted), `HOMEBREW_PREFIX`, searched `PATH`, per-candidate lines,
  bounded `rust`/`cargo` listings, `command -v` / `type -a` — never the whole environment.

## 2. Validations run in THIS session (real exit codes, shell `bash`)

| # | Command (standalone, no `tee`) | Exit | Result |
|---|---|---|---|
| 1 | `bash -n .github/ci/install-rust-toolchain.sh` | 0 | syntax clean |
| 2 | `bash -n .github/ci/gate-selftest.sh` | 0 | syntax clean |
| 3 | `bash .github/ci/gate-selftest.sh` (with `set -o pipefail`, tailed; real status via `PIPESTATUS[0]`) | 0 | `gate-selftest: 224 passed, 0 failed` — includes the BUG rows: absent-fixture block + remedy-last, defaulted-`CARGO_HOME` 3-candidate count, 20-of-25 bounded listing, narrowing mutant killed, both success fixtures silent |
| 4 | Failure probe, isolated fixtures under `/tmp/integ-306v4m` (filtered `PATH`, empty `CARGO_HOME`/`HOMEBREW_PREFIX`, `RUNNER_NAME=fake-rose-air`, `RUNNER_OS=fake-macOS`) | 1 | block printed, remedy sentence is the last line, 4 candidate lines |
| 5 | Success probe, rustup/rustc/cargo fakes only under `CARGO_HOME/bin` (filtered `PATH`, `RUSTUP_HOME` elsewhere) | 0 | `rust-pin: using rustup at /tmp/integ-306v4m/only-cargo/bin/rustup`, no diagnostics block, `GITHUB_PATH` exactly the shim dir, fake log shows `toolchain install 1.92.0 --profile minimal` |

Fresh failure-probe excerpt (exit 1, last line is the unchanged remedy):

```text
rust-pin: rustup not found; runner diagnostics:
rust-pin:   RUNNER_NAME=fake-rose-air
rust-pin:   RUNNER_OS=fake-macOS
rust-pin:   hostname=e11-1.macminivault.com
rust-pin:   whoami=administrator
rust-pin:   HOME=/Users/administrator
rust-pin:   CARGO_HOME=/tmp/integ-306v4m/empty-cargo (set)
rust-pin:   HOMEBREW_PREFIX=/tmp/integ-306v4m/empty-brew
rust-pin:   PATH=/tmp/integ-306v4m/empty-cargo/bin:/usr/bin:/bin
rust-pin:   candidate /tmp/integ-306v4m/empty-cargo/bin/rustup: absent
rust-pin:   candidate /tmp/integ-306v4m/empty-brew/bin/rustup: absent
rust-pin:   candidate /opt/homebrew/bin/rustup: absent
...
rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane
```

Fresh success-probe excerpt (exit 0):

```text
rust-pin: using rustup at /tmp/integ-306v4m/only-cargo/bin/rustup
GITHUB_PATH: /tmp/integ-306v4m/only-cargo/bin   (exactly one line)
fake log:    rustup toolchain install 1.92.0 --profile minimal
diagnostics block: absent (grep count 0)
```

## 3. Not run, and why

- `go build` / `go vet` / `golangci-lint`: no Go files in the delta (2 shell + 2 markdown); the
  configured gate for this scope is `gate-selftest.sh`, run green above. Shell syntax (`bash -n`,
  exit 0 each) is the relevant static check for the changed scripts.
- Full landing suite / hosted CI: per the brief the runtime runs it once at handoff; not run
  manually here.
- `task-board worktree status/obligations/integrating` (read-only inspect): attempted, hung
  without output, terminated — not a precondition; the `get` + `git` evidence above stands.

## 4. Close-out

- One read-only inspect (`worktree status/obligations/integrating`) was terminated without output;
  nothing else was backgrounded. All `/tmp/integ-306v4m` probes are scratch (outside the repo).
- No `worktree checkpoint` / `worktree integrate` executed (runner-owned). No `handoff` called.
- Tree verified unchanged and uncommitted; board left at `integrating` for the bound landing
  transaction, which alone may write `done`.

revision 4 = revision 3 unchanged; re-verified under the integration run so the landing evidence is tree-bound.
