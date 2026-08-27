# TASK-260720-3j8pp5 exact-commit review verdict

Date: 2026-07-30
Reviewer run: `RUN-260730-fc0927`

## Verdict

**Accepted → `done`.**

Signed commit `1d28910f5bb276ff58e2a102e06968bd7640abe3` is safe to
land independently. The task-owned Windows failures are closed on Python
3.11–3.14. The only failures remaining in the exact-commit workflow are the
same eight `tests/test_build_source.py` failures owned by
`TASK-260720-3c0ss2` / PR #8.

No implementation changes are requested. The reviewer did not edit product
files, stage, commit, push, merge, or supply `commit_ack`.

## Exact commit and scope

- Worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`
- Branch: `task/TASK-260720-3j8pp5-toolchain-identity`
- Commit: `1d28910f5bb276ff58e2a102e06968bd7640abe3`
- Parent: `d5d16bfcaa2fe43dc994b819c2659512c4fd8f0a`
- Tree: `70bdfa9b0c5b957d7b5308e4ceef1de1d61b3e7c`
- Signature: good ECDSA signature for `oparin@me.com`, key
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`
- Commit scope: exactly:
  - `src/csk/builds/toolchain.py`
  - `tests/test_builds_toolchain.py`
- `git diff --check`: clean.
- Worktree after review: clean and still exactly at the reviewed commit.

Exact SHA-256 values:

- `src/csk/builds/toolchain.py`:
  `c7b5bd70d2784d2c57a8dc336035df010b40befe388dd8ed026b3b1d4d882edd`
- `tests/test_builds_toolchain.py`:
  `201ba9f2abe42eaa26a49f8d2786d5ce194b79bcf0cf51c7a0a2e877a5224360`
- Two-file `git diff --binary` from the parent:
  `71faf1fbd73c224f95f8ff26513a4595889f1bee9b9d7cc75924517baeb4e187`

All three values exactly match the independently accepted uncommitted
candidate recorded in
`TASK-260720-3j8pp5_windows-review-verdict.md`.

## Implementation review

The production change closes the Windows false-mutation defect without
weakening mutation detection:

- `os.scandir()` is used only to collect names while its iterator is open.
- Each initial entry identity is then captured from the live path using fresh
  `os.lstat(directory / component)`, avoiding cached Windows
  `DirEntry.stat()` metadata.
- Directories still receive a later path `lstat()` plus `_same_stat`
  revalidation.
- Regular files still require initial `lstat()` to match descriptor
  `os.fstat()` before reading, an exact byte count, and a final path `lstat()`.
- Links still require path identity and target revalidation.
- `ToolchainSession.verify()` still re-fingerprints the complete GOROOT and
  compares both the digest and frozen tree state before cleanup.
- The deterministic fake-`DirEntry` regression makes `.stat()` fail if called;
  the scan succeeds through fresh `os.lstat()`.
- The subprocess newline assertion now expects native CRLF only on Windows;
  protocol normalization continues accepting exactly LF or CRLF.

Exact-commit targeted mutation regression:

```text
4 passed in 0.03s
```

This covered the fake-`DirEntry` regression, file/directory digest mutation,
close-time mutation rejection with private-state cleanup, and release cleanup.

## GitHub exact-commit attribution

- PR #9: https://github.com/ivanopcode/cocoaskills/pull/9
- Run: https://github.com/ivanopcode/cocoaskills/actions/runs/30505740935
- PR head and run `headSha` both equal the reviewed commit
  `1d28910f5bb276ff58e2a102e06968bd7640abe3`.
- PR #9 is open, non-draft, mergeable, and contains one commit with the exact
  two-file scope above.
- Strict mypy passed.
- Ubuntu Python 3.11–3.14 all passed.
- macOS Python 3.11–3.14 all passed.

The prior baseline run
https://github.com/ivanopcode/cocoaskills/actions/runs/30503926948,
Windows Python 3.14 job
https://github.com/ivanopcode/cocoaskills/actions/runs/30503926948/job/90749459882,
failed these eight task-owned tests:

1. `test_establish_uses_only_exact_bootstrap_argv_and_clean_environment`
2. `test_probe_returns_frozen_snapshot_and_removes_private_root`
3. `test_context_manager_removes_private_root`
4. `test_capture_is_immutable_across_project_path_augmentation`
5. `test_default_runner_closes_stdin_and_shares_bounded_output_budget`
6. `test_file_and_directory_bytes_are_framed_and_mutate_identity`
7. `test_tree_mutation_before_close_fails_and_still_deletes_private_state`
8. `test_release_cleans_without_second_fingerprint`

Every one of those eight tests is `PASSED` in each exact-commit Windows job:

| Python | Job | Prior task-owned tests | Remaining failures |
|---|---:|---:|---:|
| 3.11 | `90754986892` | 8/8 passed | 8 source, 0 toolchain |
| 3.12 | `90754986928` | 8/8 passed | 8 source, 0 toolchain |
| 3.13 | `90754986917` | 8/8 passed | 8 source, 0 toolchain |
| 3.14 | `90754986918` | 8/8 passed | 8 source, 0 toolchain |

Each failed Windows summary is identical and reports:

```text
8 failed, 722 passed, 39 skipped
```

All eight failures are exclusively:

1. `tests/test_build_source.py::test_frozen_snapshot_identity_includes_root_marker_while_legacy_hash_does_not`
2. `tests/test_build_source.py::test_frozen_snapshot_matches_shared_binary_vector_and_ignores_mode_and_timestamp`
3. `tests/test_build_source.py::test_frozen_snapshot_rejects_non_portable_descendant`
4. `tests/test_build_source.py::test_frozen_snapshot_use_rechecks_after_last_child_and_rejects_mutation[bytes]`
5. `tests/test_build_source.py::test_frozen_snapshot_use_rechecks_after_last_child_and_rejects_mutation[tree]`
6. `tests/test_build_source.py::test_frozen_snapshot_use_rechecks_after_last_child_and_rejects_mutation[link]`
7. `tests/test_build_source.py::test_frozen_snapshot_rejects_root_replacement_even_with_identical_bytes`
8. `tests/test_build_source.py::test_frozen_snapshot_use_rechecks_before_callback`

That ownership is independently supported by:

- Board task `TASK-260720-3c0ss2` (`build-source-context-boundary`) remaining
  active at `to-review`.
- PR #8
  https://github.com/ivanopcode/cocoaskills/pull/8 at signed candidate
  `51d8713ad14a26bdc0bafc5216fbed173ba6009b`.
- PR #8 owns `src/csk/builds/source.py`,
  `src/csk/builds/_windows.py`, and `tests/test_build_source.py`; none is in
  PR #9.

The aggregate PR #9 workflow is red and its dependent `Build artifacts` job is
skipped solely because the four source-owned Windows jobs fail. That does not
invalidate this task's independent landing:

- the exact committed file and binary-diff hashes match the prior independent
  local acceptance;
- that acceptance already passed the full two-fixture-root suite, package
  build, Twine check, and packaged-source hash comparison on these exact
  bytes; and
- PR #9 changes no package metadata or source-task file.

## Independent exact-commit gates

Authoritative interpreter:
`/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python`
(Python 3.14.4, pytest 9.0.3, mypy 2.1.0).

| Gate | Exit | Result |
|---|---:|---|
| `git verify-commit 1d28910...` | 0 | Good ECDSA signature |
| Exact focused toolchain pytest | 0 | `63 passed in 0.33s` |
| Strict `python -m mypy` | 0 | No issues in 58 source files |
| Targeted mutation regression | 0 | `4 passed in 0.03s` |
| Exact scope/hash/diff check | 0 | Clean and byte-identical |

The default Homebrew Python was also probed and intentionally found unsuitable
for project gates because it lacks pytest, mypy, and Twine. No packages were
installed and no environment or product state was changed; validation used the
established repository development interpreter.

## Conclusion

PR #9 is an independently closed toolchain fix. Its task-owned behavior is
green on every supported OS/Python combination exercised by CI, including the
required Windows Python 3.11–3.14 matrix. The separate source failures must
remain with `TASK-260720-3c0ss2` / PR #8 and do not justify routing this task
back to development.
