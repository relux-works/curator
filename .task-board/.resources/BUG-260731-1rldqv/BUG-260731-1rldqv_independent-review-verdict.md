# BUG-260731-1rldqv — independent review verdict

Reviewer run `RUN-260731-2bd01b`, read-only. Repository `ivanopcode/cocoaskills`,
PR 16, branch `task/TASK-260720-3t8nr3-transactional-project-hybrid`, head
`32737a8`.

**VERDICT: changes requested → `to-dev`.**

The Windows regression itself is fixed, correctly and at the root, and every
acceptance criterion about it is met and independently re-verified below. One
defect introduced by the same change is not: commit `7a66c73` removed the
fail-closed check that rejected a manager-home path occupied by a
non-directory, and the replacement silently accepts it. It is latent rather
than live — no current CLI path reaches it — but it sits in the very function
the commit added to establish the manager home's private state, and the fix is
two lines plus a test.

---

## 1. What I verified from primary sources

Every claim below was re-derived here, not carried over from the developer
report. All of them held.

| Check | Method | Result |
|---|---|---|
| PR 16 matrix green | `gh run view 30637483316 --json jobs` | **14/14 `success`** on head `32737a8`: 12 test cells (ubuntu/macos/windows × 3.11/3.12/3.13/3.14), `Type check / mypy strict`, `Build artifacts` |
| Head matches | `gh pr view 16` | `headRefOid = 32737a8b325506721f66a8b3d8e17f7f638c48d1`, `MERGEABLE` |
| Commits signed | `git log --format='%G?'` | `7a66c73`, `98ab7a2`, `32737a8` — all **`G`** (good signature) |
| Push scope | `git remote -v` | `origin` is `ivanopcode/cocoaskills` only; no other remote, no tags |
| mypy strict | `python -m mypy` at `32737a8` | **exit 0**, no issues in 67 source files |
| Full local suite | `pytest` at `32737a8`, macOS, CPython 3.11.6 | **exit 0 — 1145 passed, 100 skipped in 345.43s** (matches the reported 1145/100 exactly) |
| Targeted suites | `pytest test_transactions test_locking test_adapters test_installer_transactions` | **exit 0 — 156 passed, 4 skipped** |

Run history on the branch corroborates the provenance story: `30594273278`
(`8a02e17`) failed, `30624304158` (`7a66c73`), `30636027978` (`98ab7a2`) and
`30637483316` (`32737a8`) all succeeded.

## 2. The load-bearing claim — "the guard is not weakened" — checked, not accepted

`32737a8` replaces an O(n²) pairwise namespace scan with three index lookups.
That is the single riskiest edit in the change, so I tested the equivalence
claim directly rather than reasoning about it: I copied the **old** pairwise
algorithm verbatim out of `8a02e17` and ran it against the **shipped** indexed
scan over identical namespace sets on a real filesystem.

The candidate pool covers every shape the guard exists to catch — exact
duplicates, containment at several depths, a hardlink alias, a symlink to a
directory, a symlink to a file, a dangling symlink, absent paths, absent paths
under an absent parent, and `entry`/`bytes` mixtures of all of them.

```
exhaustive pair/triple sets checked: 1771
randomized sets checked:             4000   (sizes 4-9, order-shuffled)
disagreements:                       0
```

**5,771 namespace sets, zero divergence in the accept/reject decision.** The
indexed form is equivalent, and it is in one respect a superset: the old code
only reached the physical-identity check for pairs that had already survived
the parts checks, whereas the new pass indexes identity for every namespace.

I also confirmed the containment rewrite is complete rather than merely
plausible: `_namespace_parts` resolves through `_canonical_target_path`, so
every namespace is absolute and has at least an anchor component. The only
prefix `range(1, len(parts))` cannot reach is the empty one, which no namespace
can have. Nothing escapes the index.

Two smaller guard claims also hold on inspection:

- `_NamespaceProbe.identity()` sets `_identity_read = True` only after the
  `try` block completes, so a real `OSError` genuinely is not cached — it
  propagates on every call, as the docstring says.
- `_staging_tree_entry` reaches `_link_is_directory(info)` only inside
  `if stat.S_ISLNK(info.st_mode)`, so `info` is an `lstat` of the link itself
  and `st_file_attributes` describes the reparse point, not its destination.
  Adding `link_is_directory` to `_validate_staging_entry_modes` **strengthens**
  the staging guard, which the commit message understates.

## 3. The other "no-op on POSIX" claim — measured

`_permission_identity` changes what goes into the digest payload, and digests
are persisted in receipts and install markers, so "no-op on POSIX" is a
compatibility claim worth measuring rather than reading. I digested the same
tree — `tool.cmd`, `tool.exe`, `tool.bat`, `tool`, `plain.txt`, `ro.txt` across
modes `0o755/0o700/0o644/0o600/0o444` under a `0o750` directory — with the
module at `8a02e17` and at `32737a8`:

```
pre-fix  8a02e17: sha256:346f63ada809ccf52bc66cc662b30cf1689f1355cd7d45427513b7b5157b5251
post-fix 32737a8: sha256:346f63ada809ccf52bc66cc662b30cf1689f1355cd7d45427513b7b5157b5251
IDENTICAL_ON_POSIX: True
```

Byte-identical. No POSIX receipt or install marker is invalidated by this
change. `_permission_mode_identity` returns `mode` unchanged off Windows, and
reusing that existing helper rather than inventing a parallel Windows special
case is the right call architecturally — `_permission_mode_is_allowed` already
spoke that vocabulary.

## 4. Fit, coverage and honesty of the fixes

- **Producer-side, not guard-side.** `make_publication_source_private` is
  called at `installer.py:1107`, immediately before the **only**
  `CachePublication(` construction site in the tree (`installer.py:1116`), so
  coverage is complete and the publication guards themselves are untouched.
  Both routes by which a manager home comes into existence —
  `locking._prepare_manager_home` and `config._write_json_atomic` — go through
  `provision_new_manager_home`, so that coverage is complete too.
- **`make_publication_source_private` is safe with respect to the receipt.**
  It chmods after `artifact.metadata.sha256` is computed, but that digest and
  the publication-side re-hash are both content-only (`_hash_handle`), so a
  permission change cannot desynchronise them.
- **The test changes are faithful to the platform, not accommodations of it.**
  This is the thing the stop-the-line rule exists to catch, and the work passes
  it. `conftest.csk_home` now builds its home through the product's own
  `locking.provision_new_manager_home` instead of a bare `mkdir`, so the tests
  pass because the product establishes the state. `_stub_trusted_toolchain` now
  derives its target from the host and its artifact path from the product's own
  `build_metadata.derived_artifact_path`, replacing a hard-coded
  `goos="linux"` / `bin/<cmd>` that could only ever have published off Windows.
- **The diagnosis is corroborated from outside this repo.** The curator-side
  `BUG-260731-33v6zz` logbook entry independently found the same two Windows
  facts in the Go implementation — a `0x410` directory-junction reparse point,
  and "no manager home was ever protected" because a directory made by
  `os.MkdirAll` inherits its parent's ACEs. Two codebases, two languages, same
  platform truths, arrived at separately.
- **Documentation is complete and accurate** in the CocoaSkills `LOGBOOK.md`
  (80 lines across both stages), and the recorded open product decision —
  whether to add opt-in repair for a pre-existing Administrators-owned home —
  is correctly identified as a product call rather than taken unilaterally.

## 5. Rework required

### 5.1 `provision_new_manager_home` no longer fails closed on a non-directory home

`src/csk/locking.py:126-132` (introduced by `7a66c73`). The change rewrote

```python
csk_home.mkdir(mode=0o700, parents=True, exist_ok=True)
```

as

```python
csk_home.parent.mkdir(parents=True, exist_ok=True)
csk_home.mkdir(mode=0o700)
except FileExistsError:
    return
```

`Path.mkdir(exist_ok=True)` does not simply swallow `FileExistsError` — CPython's
implementation is `if not exist_ok or not self.is_dir(): raise`. It tolerates an
existing **directory** and re-raises for anything else. The bare
`except FileExistsError: return` drops that `is_dir()` condition, so a
non-directory occupying the manager-home path is now accepted.

Measured differentially, same probe on both commits:

| `~/.cocoaskills` is | `8a02e17` | `32737a8` |
|---|---|---|
| absent | OK, mode `0o700` | OK, mode `0o700` |
| an existing directory | OK, mode `0o700` | OK, mode `0o700` |
| a symlink → existing directory | OK | OK *(unchanged; `0o755` in both — pre-existing, not this change)* |
| **a symlink → missing directory** | `LockError: cannot create manager home` | **OK — home materialised at the link destination, mode `0o755`, never provisioned** |
| **a regular file** | `LockError: cannot create manager home` | **raw `FileExistsError` escapes** |

In the dangling-symlink row `GlobalLock` acquires successfully and creates
`.lock` at the link's destination. Both the `mode=0o700` create mode and, on
Windows, the entire `provision_manager_home` ownership/DACL stamp are skipped —
which is precisely the private state `7a66c73` exists to establish. The
function's own docstring claims "Only a home this call creates is provisioned,
so ownership drift on an established home still fails closed at inspection";
here a home this call did **not** create is accepted and used un-provisioned.

In the regular-file row the escaping `FileExistsError` is not in the exception
tuple `cli.main` catches (`cli.py:74-87`), so it surfaces as a traceback where
`LockError` would have printed `error: …` and returned `EXIT_LOCK`.

**Severity — stated honestly: latent, not live.** I checked reachability rather
than assuming it. Every `GlobalLock(cfg.path.parent)` call site in `cli.py`
(573, 592, 639) is preceded by `config.load_config()`, which fails first on a
non-directory home (`ConfigError` / `NotADirectoryError`). And `csk bootstrap`
already died with an unhandled `FileExistsError` on this input **before** the
change, so that path is not a regression. The defect is confined to
`provision_new_manager_home` / `_prepare_manager_home` as an API — which
`7a66c73` newly made a public, non-underscore, two-caller module function.

It is still rework rather than an accepted risk: this task's scope names
ownership and provenance checks explicitly and forbids weakening them, the
regression is inside the new ownership-provisioning code, and no test covers
the case — which is what let it through.

**Suggested repair** — restore exactly the condition `mkdir(exist_ok=True)`
applied, which keeps the symlink-to-directory row working as before:

```python
    except FileExistsError:
        # Path.mkdir(exist_ok=True) tolerates only an existing *directory*;
        # anything else at the home path must still fail closed.
        if not csk_home.is_dir():
            raise LockError(f"cannot create manager home: {csk_home}")
        return
```

Plus a regression test in `tests/test_locking.py` covering the two rows that
changed, with symlink-to-existing-directory as the positive control.

### 5.2 The second commit's cost contract is unpinned

`98ab7a2`'s contract — one canonicalisation per namespace — is pinned by
`test_namespace_independence_canonicalises_each_namespace_once` and
`test_namespace_independence_cost_stays_linear_in_targets`. `32737a8`'s
contract — that the **pairwise comparison** itself is gone — is not. Restoring
a pairwise scan over memoised probes would leave both existing cost tests
green while reinstating the 364,635 comparisons that commit removed, and on
Windows that was the dominant term. A counter on the comparison/lookup count,
in the same shape as the existing two tests, would close it.

### 5.3 Curator `LOGBOOK.md` entry stops at `7a66c73`

The curator-side entry (line 59) ends at "PR 16 full matrix `30624304158`
running on `7a66c73`". The two most transferable findings of the whole task are
missing from the cross-project record: the quadratic
`_validate_namespace_independence` (a POSIX-latent cost defect that only
Windows made fatal, worth knowing before the next `_save_journal`-adjacent
change) and the tooling anomaly that `gh run view --json jobs` reported
`completedAt` values contradicted by the job logs — 34 minutes against an
actual 8323.00s. Both are recorded in the CocoaSkills logbook and the board
notes; they belong in the curator logbook too.

## 6. Not blocking

- **Journal compatibility.** `link_is_directory` is persisted in
  `StagingTreeEntry` and is now recomputed differently on POSIX (`False` where
  `path.is_dir()` previously returned `True`), so a journal written by pre-fix
  code with a directory link would fail `_validate_staging_entry_modes` on
  resume. Unreachable in practice: entry targets with links arrive with
  `c4131bd`, which is not on `main`, so no released version can have written
  such a journal. Worth knowing if any part of this ever ships before the rest.
- **`config` → `locking` → `builds.cache` coupling** is now a two-level
  function-local import chain to avoid a cycle. It works and is commented, but
  it is a smell worth watching if a third such edge appears.
- **The open product decision** recorded by the developer — opt-in repair
  versus documented manual re-provisioning for a pre-existing
  Administrators-owned Windows home — is correctly scoped out of this task and
  should not be folded into the rework.

## 7. Routing

`to-dev`. This is ordinary, bounded rework: §5.1 is a two-line change plus a
test, §5.2 is one test, §5.3 is a logbook paragraph. Nothing here needs a human
decision, and nothing in §1-§4 needs redoing — the Windows fix, its evidence
and the green matrix all stand as delivered. On return, the reviewer needs only
the new tests red-then-green, `mypy` strict, and a green PR 16 matrix on the
new head.

Reviewer archetype is read-only, so no `commit_ack` is supplied and no code was
modified. All verification ran in throwaway `git worktree` checkouts of
`8a02e17` and `32737a8` under `/tmp`, since removed.
