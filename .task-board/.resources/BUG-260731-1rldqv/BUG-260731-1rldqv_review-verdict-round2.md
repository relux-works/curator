# BUG-260731-1rldqv — independent review verdict, round 2

**Run** `RUN-260731-d23d45` (reviewer archetype, read-only).
**Subject** rework round `RUN-260731-711b96`, signed commit `f8b90a5`, parent `32737a8`.
**Repository** `ivanopcode/cocoaskills`, PR 16, branch
`task/TASK-260720-3t8nr3-transactional-project-hybrid`.

## VERDICT: ACCEPTED

All three changes requested by `RUN-260731-2bd01b` are delivered and were
re-verified from primary sources, not read off the developer report. Every
number in `BUG-260731-1rldqv_review-rework.md` that I could independently
measure reproduced exactly. No new defect surfaced, and nothing in the original
Windows fix or the green matrix was disturbed.

Reviewer archetype: **no `commit_ack` supplied**. Acceptance evidence is
recorded below for the commit-owning mover.

---

## 1. Provenance and publication

| check | result |
|---|---|
| PR 16 head | `f8b90a5edf9bef4ffabd42b2338f399fc07b0b42` — the exact commit reviewed |
| PR state | `OPEN`, base `main`, `MERGEABLE` |
| signatures | `7a66c73` / `98ab7a2` / `32737a8` / `f8b90a5` all `%G? = G`, key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` |
| remotes | `origin` → `ivanopcode/cocoaskills` only; no second remote |
| refs carrying `f8b90a5` | the PR branch and its remote tracking ref, nothing else |
| tags | **no tag contains any of the four fix commits**; `git ls-remote --tags` shows only the pre-existing `v0.1.x` releases |
| probe branches | `git ls-remote --heads origin` = `main`, the PR branch, and the pre-existing `feat/csk-v0.10-hybrid-scope`. The disposable probe branches are gone. |
| `f8b90a5` diffstat | `LOGBOOK.md +31`, `src/csk/locking.py +8/-1`, `tests/test_locking.py +82`, `tests/test_transactions.py +107` — exactly as reported, nothing else touched |

## 2. Rework 1 — `provision_new_manager_home` fails closed again — CONFIRMED

Re-measured the five-shape probe myself, through the shared entry point
`_prepare_manager_home`, against three throwaway worktrees under `/tmp`:

| `~/.cocoaskills` is | `8a02e17` (pre-fix) | `32737a8` (defect) | `f8b90a5` (shipped) |
|---|---|---|---|
| absent | OK, `0o700` | OK, `0o700` | OK, `0o700` |
| an existing directory | OK, `0o755` | OK, `0o755` | OK, `0o755` |
| symlink → existing directory | OK, `0o755` | OK, `0o755` | OK, `0o755` |
| **symlink → missing directory** | `LockError` | **returns cleanly, `is_dir=False`** | `LockError` |
| **a regular file** | `LockError` | **returns cleanly, `is_dir=False`, mode `0o644`** | `LockError` |

`f8b90a5` agrees with the `8a02e17` baseline on all five rows. The two adoption
rows are untouched, so the restored condition did not over-tighten.

**The developer's correction to my predecessor's verdict is right.** Calling
`provision_new_manager_home` directly on the regular-file input at `32737a8`
returns *cleanly* — it does not let a raw `FileExistsError` escape. I reproduced
that directly, and separately reproduced the traceback: with the fix reverted,
`test_locking_refuses_to_bind_to_a_home_path_taken_by_a_file` dies at
`os.mkdir(self, mode)` **downstream** of the accepted non-home. Both states are
fail-open, so the fix is unchanged, and the new `GlobalLock` test now pins the
`LockError`-not-`OSError` boundary `cli.main` depends on.

**Red proof reproduced.** Reverting only the `except FileExistsError` body in my
worktree: `pytest tests/test_locking.py -k "provisioning or refuses_to_bind"` →
**exit 1, 3 failed, 2 passed**. The three rejections fail; the two adoption
tests pass either way, which is what makes them controls rather than duplicates.
Restored: **5 passed**.

Reviewed for over-reach: `Path.is_dir()` swallows `OSError` and returns `False`,
so an unreadable home also fails closed — the safe direction. The
`mkdir` → `is_dir()` window is the same TOCTOU CPython's own
`mkdir(exist_ok=True)` carries, so this restores parity rather than introducing
anything.

Fixture fidelity preserved: `tests/conftest.py:60` calls
`provision_new_manager_home` on an **absent** path, so it still takes the create
branch and is fully provisioned. The condition changes no test's coverage.

## 3. Rework 2 — the pairwise scan is pinned as gone — CONFIRMED

`test_namespace_independence_never_walks_the_pairs` measures at 4 / 8 / 12
targets, which is 32 / 60 / 88 namespaces. I re-measured on my own host:

| namespaces | shipped indexed scan | my measurement |
|---|---|---|
| 32 | 156 | **156** |
| 60 | 296 | **296** |
| 88 | 436 | **436** |

The report's table is labelled by namespaces, not targets — the numbers are
exact, not approximate. I also measured with **fresh** monkeypatch fixtures per
size (the shipped test reuses one fixture across three nested patches) and got
byte-identical counts, which rules out the nesting distorting the contract.

**Red proof reproduced independently.** I restored `98ab7a2`'s pairwise
`_namespaces_overlap` scan over the same memoised probes and re-ran the
namespace suite: **exit 1, 1 failed, 18 passed** —
`assert (7364 - 2128) == (15736 - 7364)`. It fails this test *alone*; both
canonicalisation cost tests stay green. That is precisely the hole the rework
was asked to close.

## 4. Guard equivalence — independently re-derived, not carried over

The riskiest edit in the whole change is `32737a8`'s replacement of the O(n²)
pairwise scan with three index lookups. I did not accept my predecessor's
differential; I ran my own against the pairwise scan I had just restored, over
identical namespace sets on a real filesystem:

- 18 filesystem shapes: plain file, directory, child-under-directory, absent,
  absent-under-absent-parent, hardlink alias, symlink→dir, symlink→file,
  dangling symlink, three nesting depths, five unused paths.
- **3,867 sets** — all pairs and triples of those shapes, every
  same-path-twice duplicate, plus 3,000 randomized sets of size 2–6.
- Compared on `(raised?, exception type name)` — each module defines its own
  exception classes, so comparison is by name.

**Zero disagreements.** Independent corroboration of the 5,771-set result in
the round-1 verdict, by a different harness and a different shape pool.

## 5. No test was weakened to accommodate the platform

`git diff 8a02e17 f8b90a5 -- tests/` adds exactly three skip constructs:

- 2 × `pytest.skip("symlinks unavailable")` — host-capability guards in the two
  **new** symlink tests. The Windows job logs prove they **did not fire**.
- 1 × `@pytest.mark.skipif(os.name != "posix")` on a **new** POSIX-only privacy
  test. Additive.

No pre-existing test acquired a skip marker.

**Windows skip set-diff, `main b3a5031` vs `f8b90a5` (Python 3.11 job logs).**
Six tests are skipped at head and not at base. All six are new-file or new-test,
never previously running: five in `tests/test_installer_transactions.py`, a file
that **does not exist on `main`**, gated by `POSIX_BUILD_VECTOR` which `c4131bd`
introduced — the PR's own baseline, not the fix; and one new POSIX-only test in
`test_build_cache_posix.py` added by `7a66c73`. **Nothing that ran before is
skipped now.**

## 6. Acceptance criteria — re-proved at the PR head

**CI run `30641011440`, head `f8b90a5` (= the PR head): `conclusion = success`,
14/14 jobs** — 12 test cells, `Type check / mypy strict`, `Build artifacts`.

Test counts pulled from the raw job logs via
`gh api repos/.../actions/jobs/<id>/logs`, not from a summary:

| lane | result | duration |
|---|---|---|
| windows 3.11 / 3.12 / 3.13 / 3.14 | 1217 passed, 152 skipped | 672.01 / 851.24 / 493.38 / 660.22 s |
| ubuntu ×4 | 1295 passed, 74 skipped | 107.60 – 311.69 s |
| macOS ×4 | 1319 passed, 50 skipped | 136.46 – 165.98 s |

Collection totals 1369 on every platform, so no lane is quietly collecting less.
Against `32737a8`'s 1211 / 152 on Windows the delta is **+6 passed, 0 new
skips** — all six new tests executed on Windows. Confirmed by name in the logs:
the four `test_provisioning_*` cases and
`test_locking_refuses_to_bind_to_a_home_path_taken_by_a_file` are `PASSED` at
59 %, `test_namespace_independence_never_walks_the_pairs` `PASSED` at 99 %.

**Original failure signatures are gone.** Grepping the Windows 3.11 job log:
`transaction target changed while digesting` = 0, `WinError 5` = 0,
`cache_publication_invalid` = 0, `FAILED` = 0, `ERROR ` = 0.

**Every originally-failing file passes on Windows**, none skipped wholesale:

| file | PASSED | SKIPPED | FAILED |
|---|---|---|---|
| test_activation_modes | 6 | 2 | 0 |
| test_audit_cli | 12 | 0 | 0 |
| test_closure_install | 12 | 0 | 0 |
| test_dev_substitution | 7 | 0 | 0 |
| test_gc | 8 | 0 | 0 |
| test_global_install | 33 | 6 | 0 |
| test_hybrid_scope | 9 | 0 | 0 |
| test_install | 52 | 6 | 0 |
| test_mcp_dependencies | 24 | 0 | 0 |
| test_status | 2 | 0 | 0 |

That is the AC's "the failing tests pass on windows-latest for Python 3.11,
3.12, 3.13 and 3.14 without weakening the corruption guard", proved case by
case rather than by the job's green tick.

**Reviewer-local, macOS, CPython 3.11.6, real exit codes, each command
standalone in a throwaway worktree of `f8b90a5`:**

| command | exit | result |
|---|---|---|
| `python -m mypy` | **0** | no issues in 67 source files |
| `pytest tests/test_locking.py -k "provisioning or refuses_to_bind"` | **0** | 5 passed |
| same, locking fix reverted | **1** | 3 failed, 2 passed — *red proof* |
| `pytest tests/test_transactions.py -k namespace` | **0** | 19 passed |
| same, pairwise scan restored | **1** | 1 failed, 18 passed — *red proof* |
| `pytest` (full suite) | **0** | **1151 passed, 100 skipped** in 371.89 s |

The full-suite figure matches the reported 1151 / 100 exactly, against a
1145 / 100 baseline at `32737a8` — the delta is exactly the six new tests. No
dedicated linter is configured in `pyproject.toml`; mypy strict is the project's
static gate and is green locally and in CI.

## 7. Rework 3 — curator `LOGBOOK.md` — CONFIRMED

The new `1810` entry carries both transferable findings the earlier entry
stopped short of: the quadratic `_validate_namespace_independence` as a
POSIX-latent cost defect (750,620 canonicalisations, 182 s of 210 s, on macOS)
that only a slower syscall promoted to an outage, and the
`gh run view --json jobs` `completedAt` anomaly (implied ~34 min against an
actual 8323.00 s in the job logs). It also records the `CreateDirectoryW`
reparse fact and the `gh run view --log` corollary. CocoaSkills `LOGBOOK.md`
carries the two code reworks, including the generalisable one.

It is **uncommitted in the curator working tree**, the way previous rounds left
it, for the orchestrator to land — consistent with the accepted sibling
`BUG-260731-27h1yc`.

## 8. Non-blocking observations, for the record

- **A symlinked home is still adopted un-provisioned.** `f8b90a5` deliberately
  keeps the symlink→existing-directory row accepting, matching the `8a02e17`
  baseline. On Windows that home skips the DACL stamp and will fail closed
  later at the build cache. This is the same fail-closed shape as the open
  product decision below, not a new hole.
- **The cost test's nested monkeypatches are subtle but sound.** Three calls to
  `_count_namespace_pass_work` share one fixture, so each patch wraps the last.
  I verified by measuring with fresh fixtures per size and getting identical
  counts. Worth a comment if that helper is extended.
- **Still open, correctly not folded in:** a Windows manager home created by an
  elevated shell before `7a66c73` stays `BUILTIN\Administrators`-owned and
  fails closed. Automatic adoption would blunt the drift guard. Explicit opt-in
  repair versus documented manual re-provisioning remains a **product
  decision**, not rework, and was properly kept out of this round.
- **Tooling, confirmed again this round:** `gh run view --log` returned nothing
  useful; `gh api repos/.../actions/jobs/<id>/logs` returned the full ~233 KB
  per job. That is the endpoint to use.

## 9. For the commit-owning mover

- Accept and land PR 16 at head `f8b90a5`. Base `main` `b3a5031`, `MERGEABLE`,
  all six branch commits GPG-good.
- Commit the curator `LOGBOOK.md` working-tree change that carries the
  cross-project entry.
- No tag or GitHub Release was created and none is implied by this scope.

Reviewer verification ran entirely in throwaway `git worktree` checkouts of
`8a02e17`, `32737a8` and `f8b90a5` under `/tmp`, since removed. No repository
file was modified, and no `commit_ack` is supplied.
