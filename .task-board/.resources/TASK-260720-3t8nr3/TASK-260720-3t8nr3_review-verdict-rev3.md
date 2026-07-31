# TASK-260720-3t8nr3 review verdict rev 3 — ACCEPTED

Reviewer run: `RUN-260730-218cef`. Read-only review; no product code was modified.
Supersedes `TASK-260720-3t8nr3_review-verdict-rev2.md` (run `RUN-260730-bbd2a8`) and
`TASK-260720-3t8nr3_review-verdict.md` (run `RUN-260730-c38a81`), both retained for the
audit trail.

## Reviewed artifact

- Repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Worktree: `.temp/TASK-260720-3t8nr3/worktree`
- Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid`
- Worktree `HEAD` = `b3a5031ed551b27a298eef486a068b5175beaacc`. Re-fetched this cycle:
  `main` == `origin/main` == worktree `HEAD` == recorded base SHA.
- Uncommitted: 9 modified files (2276 insertions / 256 deletions) plus the untracked
  `tests/test_installer_transactions.py` (701 lines).
- Base preflight (board notes, `BASE PREFLIGHT 2026-07-30`) records the clean-clone
  fast-forward and both dependency handoffs (`TASK-260720-11yhth`, `TASK-260720-2x6mjn`)
  as `done` with outcome evidence before the worktree existed. DoD item 1 satisfied.

## Gates independently re-run by the reviewer

Interpreter `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python`; pytest
resolves `pythonpath = ["src"]` from the worktree. Every command ran as a standalone
process, no `tee`, no pipe chain.

| Gate | Command | Result |
| --- | --- | --- |
| Decisive full suite | `python -m pytest -q` | **1134 passed, 98 skipped in 1380.83s**, exit 0 (implementer claimed 1134/98 in 1362.58s — exact count match) |
| Strict typing | `python -m mypy` | Success: no issues found in 67 source files, exit 0 |
| Lint | `uvx ruff check` over the 9 changed sources/tests | All checks passed, exit 0 |
| Patch hygiene | `git diff --check` | clean, exit 0 |

Raw suite output: `.temp/review-218cef/full-suite.log` in the worktree. Every reported
figure reproduces. The implementer's evidence is honest, as in both prior cycles.

## Prior-cycle blocking finding — verified closed

### B3 — `adapter_mode: auto` decided by `$TMPDIR` — CLOSED

`_commit_materialization` (`src/csk/installer.py:1236`–`:1260`) now derives its staging
parent from `project.path.resolve(strict=False).parent`, with `config.path.parent`
(the manager home) as the fallback, and passes it as `dir=` to
`tempfile.TemporaryDirectory(prefix=".csk-materialization-plan-")`. The process temp
root is no longer consulted for materialization; `grep` confirms the prefix appears
exactly once in the tree, at the creation site.

Reproduced with the **same real macOS RAM disk method the rev-2 review used**
(`hdiutil attach -nomount ram://262144` + `diskutil erasevolume HFS+ CSKRAM`), running
the rev-2 reviewer's own scripts unmodified against the reworked worktree:

`.temp/review-bbd2a8/flip_probe.py` — same project, same config, three consecutive
installs, only `$TMPDIR` changes:

```
install#1 default TMPDIR : status=ok errors=[] tempdev=16777232 .claude/skills/skill-a symlink=True
install#2 TMPDIR=ramdisk : status=ok errors=[] tempdev=16777236 .claude/skills/skill-a symlink=True
install#3 default TMPDIR : status=ok errors=[] tempdev=16777232 .claude/skills/skill-a symlink=True
```

Rev 2 measured `True / False / True` on this exact script. The flip is gone.

`.temp/review-bbd2a8/xdev_probe.py` with `XDEV_WORK=/Volumes/CSKRAM` — project on the
RAM disk (`st_dev` 16777236), process temp on the root filesystem (16777232):

```
status: ok errors: []
tempdir dev: 16777232 project dev: 16777236
.claude/skills/skill-a -> symlink=True
.codex/skills/skill-a  -> symlink=True
```

The witness is now destination-anchored: a project on a foreign filesystem still gets
symlink mirrors. Coverage exists for both halves —
`tests/test_install.py::test_materialization_staging_is_private_and_anchored_to_project_filesystem`
points `tempfile.tempdir` at an unrelated location and asserts the real stage is a
cleaned-up sibling of the physical project outside both the checkout and the process temp
root, and `tests/test_adapters.py::test_transaction_auto_mode_rejects_cross_device_staging_without_probe`
supplies distinct device identities and asserts the probe is never attempted on the
cross-device branch (the rev-2 gap: the old vector could only exercise same-device).
`LOGBOOK.md` now records the anchor, the fallback, and the conservative copy decision.

## Prior-cycle non-blocking findings

**N5 (post-commit GC contention failed a committed install) — CLOSED.**
`installer.install` (`src/csk/installer.py:136`–`:148`) re-raises `locking.LockOrderError`
first, then treats `locking.LockError` as a `post-install garbage collection skipped`
message on a successful result. `LockOrderError` subclasses `LockError`
(`src/csk/locking.py:44`), so the ordering is correct and an internal lock-order defect
stays loud. `tests/test_install.py::test_post_commit_gc_lock_contention_does_not_fail_committed_install`
makes only the post-commit acquisition fail (keyed on the marker already existing) and
asserts the marker is live, `errors` is empty, and the message is present.

**N6 (staging placement) — CLOSED as part of B3, and recorded.** The O(total installed
state) copy is unchanged in size but no longer lands in a possibly RAM-backed system temp
directory. `LOGBOOK.md` states both the size tradeoff and the new location, and names
target-scoped copy-on-write staging as the deferred optimization.

**N7 (consumer ledger canonicalization) — RECORDED.** `LOGBOOK.md` and the results
artifact now state that `consumers.encode_consumers` resolves paths before sorting, why
(digest stability across lexical symlink routes), and that the first successful install
after upgrade may rewrite a legacy unresolved entry.

Separately checked and cleared: `_desired_consumers` (`src/csk/installer.py:1657`) drops
ledger entries whose project is missing or carries no markers. That looked like a new
regression against the pre-task append-only `consumers.record_consumer`, but
`gc.collect_runtime` (`src/csk/gc.py:40`–`:47`) already applied exactly that rule and ran
unconditionally after every non-dry-run install batch on pre-task `main`. The semantics
are preserved, only relocated into the transaction.

## New non-blocking observations

None of these blocks acceptance; all are recoverable and none contradicts a stated AC.

### N8 — a hard kill leaves the staging tree beside the project, and nothing sweeps it

Moving staging out of the system temp directory removes OS reaping. Measured with
`.temp/review-218cef/crash_litter_probe.py` (SIGKILL delivered from inside
`_stage_materialization`):

```
child returncode: -9 (-9 == SIGKILL)
staging root: /private/tmp/crash-37x_sd08/.csk-materialization-plan-xiyxbi8c
staging root still exists after crash: True
siblings of project after crash: ['.csk-materialization-plan-xiyxbi8c', 'cskhome', 'project', ...]
```

The leftover is hidden and correctness-neutral — no live surface is touched, and the next
install proceeds normally — but it is O(total installed state) in size and now sits in the
user's directory permanently. The project already has the machinery for this:
`gc.sweep_orphans` / `_is_dead_install_orphan` collect `.<name>.tmp-<pid>` leftovers by
liveness. Encoding the owning pid in the staging prefix and sweeping dead ones during the
existing post-commit GC would close it.

### N9 — the manager-home fallback can still degrade `auto` to `copy` cross-device

When the physical project parent cannot host staging, the fallback is the manager home,
and the same-device witness then compares the *home* filesystem against the destination.
Measured with `.temp/review-218cef/fallback_xdev_probe.py` (project on the RAM disk with
its parent chmod'd to `r-x`, manager home on the root filesystem):

```
status: ok  errors: []
staging roots: ['/tmp/xhome-cbu24dz8/.csk-materialization-plan-9_k_q26g']
claude mirror exists: True symlink: False
```

Same class as B3 — the destination fully supports symlinks and `auto` chose copy — but
unlike B3 it is deterministic (a function of filesystem layout, not of an environment
variable), it is the conservative direction, it requires an unwritable project parent, and
`LOGBOOK.md` states it explicitly. Recorded rather than hidden, so it is an accepted
residual. The same-device-writable fallback path itself works end-to-end
(`.temp/review-218cef/fallback_probe.py`: staging lands under `csk_home`, install ok,
mirror symlinked, no leftovers).

### N10 — for a managed project nested under another managed project, staging is created inside the outer checkout

`physical_project_parent` is the parent of the project, so installing `/repo/apps/web`
stages at `/repo/apps/.csk-materialization-plan-*` — inside `/repo`'s checkout if `/repo`
is itself a csk project. It is transient, hidden, outside every materialization target and
every generation-probe path (`_project_generation_probe`, `src/csk/installer.py:629`), so
nothing miscomputes. It is a small tension with N4's "never write into a live project", and
combined with N8 a crash makes that leftover permanent. Worth one line if the staging
anchor is revisited.

### N11 — `_link_probe` runs once per adapter mirror

`stage_project_adapter_targets` calls `_transaction_links_supported` inside the per-target
loop (`src/csk/adapters.py:308`–`:314`), so `auto` mode creates and unlinks a probe symlink
once per skill per agent instead of once per adapter root. Trivial cost, no correctness
impact.

## AC coverage — spot-checked independently

- **Audit before build, build before mutation.**
  `test_real_builds_run_provider_first_then_lexically_and_activate_marker_v2` asserts
  `events.index("audit") < events.index("build:alpha")` and
  `events.index("build:middle") < events.index("transaction:prepared")`, with build order
  `alpha, zeta, middle` (provider-first, then command-lexical). Confirmed structurally at
  `src/csk/installer.py:380`–`:490`: builds run outside the home lock, the lock covers only
  recover → generation/preimage/closure revalidation → publication → commit.
- **Failure leaves live surfaces unchanged.** `_tree_state` compares kind, mode, bytes and
  symlink targets recursively across `.agents`, `.claude`, `runtime`, `hybrid`, `builds`,
  `consumers.json`; asserted for the build-failure and publication-failure vectors.
- **Marker v2 / build-root isolation / immutable shims.** Asserted directly
  (`marker["schema_version"] == 2`, `build_roots == ["build"]`, `not (…/provider/build).exists()`,
  shims executed via `subprocess.run` and their stdout checked).
- **Consumer last + rollback.** `test_second_project_consumer_failure_rolls_back_without_touching_first`
  asserts `committed[-1] == "90-consumer"`, full tree restoration of *both* projects, and
  that the first project's ledger entry survives. Class ordering makes this structural, not
  incidental: adapter entries (`…/entry/<name>`) sort before the adapter ledger
  (`…/ledger`) inside `60-adapter-ledger`, and `90-consumer` is last overall.
- **Stale plans restart.** `test_commit_generation_change_restarts_complete_project_plan`
  asserts the audit gate ran exactly twice and the generation observation sequence was
  fully consumed.
- **Windows concurrency instruction (attached `main-windows-concurrency-flake.md`).**
  Re-verified: the vector is not skipped, not xfailed, keeps
  `assert all(not thread.is_alive() for thread in threads)` verbatim, and now adds
  per-worker `completed` events. Coordination is event-driven (both project locks held →
  A takes the home lock → B signals its attempt → A commits), bounded at 10s POSIX / 30s
  Windows with a `2×coordination + 5s` deadline off `time.monotonic()`. No module-level
  `pytestmark` in either `tests/test_transactions.py` or `tests/test_installer_transactions.py`;
  `POSIX_BUILD_VECTOR` is per-test on 5 of 7. CI (`.github/workflows/ci.yml`) runs the full
  suite on ubuntu/macos/windows × 3.11–3.14, so the Windows branch is exercised there.

## Lock-ownership change — reviewed, not a regression

`cli._dispatch` no longer wraps `install`/`upgrade` in `GlobalLock`. `GlobalLock`
subclasses `ManagerHomeLock` over the same `<home>/.lock` file
(`src/csk/locking.py:669`, `:750`), so mutual exclusion with `update` is preserved; the
window is deliberately narrowed to the commit phase, which is exactly what commit-time
generation and preimage revalidation plus plan restart exist to cover.

## Verdict

**accepted.**

Three review cycles, and every finding raised in the first two is closed or explicitly
recorded — none papered over. B3, the one blocking item from rev 2, is fixed at the right
layer: the capability witness is now anchored to the destination filesystem instead of an
unrelated environment variable, proven on a real second filesystem with the previous
reviewer's own scripts, and covered by vectors on both the anchoring and the cross-device
branch. N5 is fixed with a real regression vector; N6 and N7 are recorded in `LOGBOOK.md`
with their rationale. Every claimed gate reproduces exactly, including the 1134/98 full
suite.

N8–N11 are new, small, and none of them touches a stated AC or a live surface. N8 (crash
litter) is the one worth a follow-up item — the existing dead-pid orphan sweep is the
natural home for it.

Per the reviewer constraint, this run supplies no `commit_ack`. Acceptance evidence is
recorded here and in the board notes for the commit-owning mover, which commits the scope
and makes the final `done` transition with `commit_ack=scope_committed`.

Reviewer probe scripts left in the worktree for reproduction:
`.temp/review-218cef/crash_litter_probe.py`, `.temp/review-218cef/fallback_probe.py`,
`.temp/review-218cef/fallback_xdev_probe.py`, plus `.temp/review-218cef/full-suite.log`.
