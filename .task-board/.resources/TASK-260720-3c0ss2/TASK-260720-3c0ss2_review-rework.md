# TASK-260720-3c0ss2 — installed-context currentness rework

## Provenance and scope

- CocoaSkills repository: `/Users/iv/Developer/intranet/cocoaskills`
- Clean, fetched `main`, task `HEAD`, and `origin/main`:
  `dd76b570f88339fd1d659c02950e68b17f6ba834`
- Existing task worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-3c0ss2/worktree`
- Branch: `task/TASK-260720-3c0ss2-build-source-identity`
- Dependency `TASK-260720-z9j4c9`: `done`
- Reviewer verdict addressed:
  `TASK-260720-3c0ss2_review-verdict.md`

This rework changes only installed-context currentness and its focused install
regression. It does not add receipts, cache publication, installer commit
logic, or any Go implementation or invocation.

## Reviewer finding and fix

The prior candidate could report a schema-6 installed tree as `up-to-date`
when a pre-exclusion build-root subtree was physically present and the marker's
legacy `content_sha256` had been recomputed to match it.

`_marker_is_current` now checks every declared build root before accepting the
installed content hash:

- any marker `files` entry at or below a build root invalidates currentness;
- any physical path at a declared build root, including a link observable by
  `lstat`, invalidates currentness;
- malformed marker `files` shapes fail closed;
- the ordinary reinstall path then replaces the tree with context selected by
  the build-root-aware whitelist.

The regression independently covers physical-only, marker-entry-only, and
combined pre-exclusion states. Each state is marker/hash-consistent before the
second install; that install no longer reports `up-to-date`, removes the build
root and stale marker entry, and a third install is then `up-to-date`.

## Exact validation ledger

All commands ran directly in the task worktree. The existing project Python
environment was selected with:

```text
PATH=/Users/iv/Developer/intranet/cocoaskills/.venv/bin:$PATH
```

| Command / gate | Result | Exit |
|---|---:|---:|
| Ambient `python -m pytest -q tests/test_install.py -k stale_build_root_forces_context_reinstall` | `python` unavailable | 127 |
| Ambient `python3 -m pytest -q tests/test_install.py -k stale_build_root_forces_context_reinstall` | pytest module unavailable | 1 |
| Project-env test-first regression | 3 failed as expected against the reviewed bug | 1 |
| Project-env post-fix regression | 3 passed, 43 deselected | 0 |
| Accepted rc.5/task-focused pytest over protocol, build source, hashing, whitelist, skillcheck, and install | 196 passed | 0 |
| Strict `python -m mypy` | Success; 57 source files | 0 |
| Full pytest with accepted `CURATOR_CONFORMANCE_ROOT` and `CURATOR_SCHEMA_V6_ROOT` | 705 passed, 1 skipped | 0 |
| `python -m build` | wheel and sdist built | 0 |
| `python -m twine check dist/*` | both distributions passed | 0 |
| Wheel inventory for `csk/builds/source.py` | present | 0 |
| Sdist inventory for `src/csk/builds/source.py` | present | 0 |
| `python -m compileall -q src/csk tests` | no diagnostics | 0 |
| `git diff --check` | no diagnostics | 0 |

Accepted conformance root:

```text
/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1
```

The corrected shared digest transcriptions are:

- empty snapshot:
  `sha256:3a518980ed122b2139e46152d9c4dda7426a42572f3235cde8cbe781566f5753`
- binary/empty/root-marker vector:
  `sha256:68008c9a1131c1295d78f4f7d184c3df5f7382a88d8d40333be7cf02b2ee4de9`

The original `TASK-260720-3c0ss2_results.md` outcome was revised with these
authoritative values and this rework's current gate ledger.

## Worktree state

The candidate remains unstaged and uncommitted. No branch was pushed, no tag
or release was created, and no Go command was run.
