# BUG-260731-1rldqv — review rework

Round `RUN-260731-711b96` (developer). Answers the three changes requested by
independent reviewer `RUN-260731-2bd01b`. Repository `ivanopcode/cocoaskills`,
PR 16, branch `task/TASK-260720-3t8nr3-transactional-project-hybrid`.

One signed commit, `f8b90a5`, parent `32737a8`. Nothing in the original Windows
fix, its evidence, or the green matrix that produced it was touched. None of
the three items needed a human decision, so none was escalated.

---

## Rework 1 — `provision_new_manager_home` fails closed again

`src/csk/locking.py:126`. The function has to know whether *this* call created
the home, because only a home it creates may be stamped private; an established
home must fail closed on ownership drift at inspection instead.
`mkdir(exist_ok=True)` cannot answer that question, which is why `7a66c73`
rewrote it as a plain `mkdir` with `except FileExistsError: return`.

That is not what the flag does. CPython's body is
`if not exist_ok or not self.is_dir(): raise` — it tolerates an existing
*directory* and re-raises for anything else. The bare `except` dropped the
condition, so a non-directory at the home path became acceptable.

```python
    except FileExistsError as exc:
        if not csk_home.is_dir():
            raise LockError(f"cannot create manager home: {csk_home}") from exc
        return
```

**Verified differentially, not by reading.** The same five-shape probe was run
against a throwaway worktree of `8a02e17` (pre-fix baseline) and against the
fixed module, through the shared entry point `_prepare_manager_home`:

| `~/.cocoaskills` is | `8a02e17` | `32737a8` (defect) | `f8b90a5` (fixed) |
|---|---|---|---|
| absent | OK, `0o700` | OK, `0o700` | OK, `0o700` |
| an existing directory | OK, `0o755` | OK, `0o755` | OK, `0o755` |
| a symlink → existing directory | OK, `0o755` | OK, `0o755` | OK, `0o755` |
| **a symlink → missing directory** | `LockError` | **OK, never provisioned** | `LockError` |
| **a regular file** | `LockError` | **OK, silently accepted** | `LockError` |

All five rows now match the pre-fix baseline exactly. The two adopted rows are
unchanged, so the condition did not over-tighten.

**One correction to the verdict, on a detail.** The verdict recorded the
regular-file row on `32737a8` as "raw `FileExistsError` escapes".
`provision_new_manager_home` itself returned *cleanly* on that input — measured
above. The traceback the reviewer saw comes from a later `mkdir` once locking
went on to use the accepted non-home. Both are fail-open and the fix is the
same, but `test_locking_refuses_to_bind_to_a_home_path_taken_by_a_file` now
pins the `LockError`-not-`OSError` boundary that `cli.main` depends on, which
is the part of that observation that mattered.

**Tests** — `tests/test_locking.py`, five, one per shape:

| test | asserts |
|---|---|
| `test_provisioning_adopts_an_established_home_directory` | positive control: adopted, not remade, mode untouched |
| `test_provisioning_adopts_a_home_reached_through_a_symlink` | positive control: resolves to a directory, so still adopted |
| `test_provisioning_rejects_a_home_path_taken_by_a_file` | `LockError`, and the file is not disturbed |
| `test_provisioning_rejects_a_home_symlinked_to_a_missing_directory` | `LockError`, and the destination is not materialised |
| `test_locking_refuses_to_bind_to_a_home_path_taken_by_a_file` | the rejection reaches the caller as `LockError` |

Red/green verified by reverting only `src/csk/locking.py`: **exit 1, 3 failed,
2 passed** — the three rejections fail, and the two positive controls pass
either way, which is what shows they are controls rather than duplicates.

## Rework 2 — the pairwise scan is now pinned as gone

`tests/test_transactions.py`. Memoising the probes (`98ab7a2`) and indexing
them (`32737a8`) were two separate wins, and only the first had a tripwire.

The new test counts every read of the state a namespace is compared by — its
normalized parts and its physical identity — plus every equality between two
parts tuples, at three equally spaced sizes. Equal namespace steps must produce
equal work steps.

| namespaces | indexed (shipped) | pairwise over the same memoised probes |
|---|---|---|
| 32 | 156 | 3,500 |
| 60 | 296 | 12,446 |
| 88 | 436 | 26,880 |
| first differences | 140, 140 | 8,946, 14,434 |

Red proof: the pairwise body was actually restored in `transactions.py` and the
namespace suite re-run — **exit 1, 1 failed, 18 passed**. It fails this test
*alone*; both canonicalisation tests stay green, exactly as the reviewer
predicted they would.

## Rework 3 — curator `LOGBOOK.md`

New entry carries the two transferable findings the earlier entry stopped
short of:

- The quadratic `_validate_namespace_independence` was never Windows-specific.
  It canonicalised **both** operands per comparison and re-ran on every journal
  save, twice per 32 KiB staging chunk. One install measured **750,620**
  canonicalisations, 182 s of its 210 s, *on macOS*. Windows opens a handle per
  path component for the same question, so only there did a quadratic pass with
  a cheap constant become an outage. A POSIX-latent cost defect that a slower
  syscall promotes — worth looking for before the next change near a hot
  revalidation path.
- `gh run view --json jobs` returned `completedAt` values implying the three
  green Windows cells took ~34 minutes. The job **logs** said `8323.00s`. The
  API field was wrong; the logs are authoritative.

CocoaSkills `LOGBOOK.md` carries the two code reworks, including the
generalisable one: an `exist_ok=True` → `except FileExistsError` rewrite is
never a refactor, because `pathlib` attaches a type check to that flag which
the bare `except` discards.

---

## Evidence — real exit codes, each command run standalone

Local, macOS, CPython 3.11.6:

| command | exit | result |
|---|---|---|
| `python -m mypy` | **0** | no issues in 67 source files |
| `pytest tests/test_locking.py -k "provisioning or refuses_to_bind"` | **0** | 5 passed |
| same, with `src/csk/locking.py` reverted | **1** | 3 failed, 2 passed — *red proof* |
| `pytest tests/test_transactions.py -k namespace` | **0** | 19 passed |
| same, with the pairwise scan restored | **1** | 1 failed, 18 passed — *red proof* |
| `pytest` transactions + locking + adapters + installer_transactions | **0** | 162 passed, 4 skipped (156 baseline + 6) |
| `pytest` (full suite) | **0** | **1151 passed, 100 skipped** in 386.22s |

The full-suite baseline at `32737a8` was 1145 / 100, so the delta is exactly
the six new tests. No dedicated linter is configured in `pyproject.toml`; mypy
strict is the project's static gate and is green locally and in CI.

CI — PR 16 run `30641011440` on head `f8b90a5`: **`conclusion = success`, 14/14
jobs green** — 12 test cells (ubuntu / macOS / windows × 3.11 / 3.12 / 3.13 /
3.14), `Type check / mypy strict`, `Build artifacts`.

Windows cells, read from the **job logs** rather than the API timing field:

| cell | result | duration |
|---|---|---|
| 3.11 | 1217 passed, 152 skipped | 672.01s |
| 3.12 | 1217 passed, 152 skipped | 851.24s |
| 3.13 | 1217 passed, 152 skipped | 493.38s |
| 3.14 | 1217 passed, 152 skipped | 660.22s |

`32737a8` reported 1211 / 152 on every Windows cell, so the delta is **+6 and 0
new skips** — all six new tests actually executed on Windows.

## Verified Windows platform fact, previously assumed the other way

Both symlink tests ran on `windows-latest` rather than skipping, which settles
a question I could not answer locally with `ssh win` offline: **`CreateDirectoryW`
on a dangling symlink returns `ERROR_ALREADY_EXISTS`; it does not reparse to the
link destination and create a directory there.** I had reasoned it might, since
the final component of a path is normally reparsed when `FILE_OPEN_REPARSE_POINT`
is absent, and had a platform-tolerant version of the test ready to push. The
runner says otherwise, so the strict assertion — `LockError` on every platform —
stands, and the stronger test is the one that shipped.

The four `test_provisioning_*` cases and `test_locking_refuses_to_bind_*` are
`PASSED` in the 3.11 job log at 59%, and `test_namespace_independence_never_walks_the_pairs`
at 99%.

`ssh win` (tailscale `mbpro-win`) was offline for this round as well —
`connect to host 100.120.84.42 port 22: Operation timed out` — so
`windows-latest` runners were again the only Windows harness.

## Still open, unchanged and out of scope

A Windows manager home created by an elevated shell before `7a66c73` stays
Administrators-owned and fails closed. Adopting it automatically would blunt
the drift guard. Explicit opt-in repair versus documented manual
re-provisioning remains a product decision, and the reviewer was explicit that
it must not be folded into this rework. It was not.
