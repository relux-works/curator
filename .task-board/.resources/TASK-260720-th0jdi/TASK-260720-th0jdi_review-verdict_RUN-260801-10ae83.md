# TASK-260720-th0jdi independent review verdict

## Verdict

**Accepted.** Route the task from `reviewing` to `done`.

No correctness, safety, architecture, or acceptance-criteria defect remains at the reviewed head. This reviewer made no product-code, test, merge, tag, or release change and does not supply `commit_ack`.

## Reviewed provenance

- Repository: CocoaSkills, PR #18
- Base: `07655553cebcf867bbe58629de98e77644606c85`
- Reviewed head: `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09`
- Merge-base: exactly the declared base; the base is an ancestor of the reviewed head.
- Worktree: clean at the reviewed head.
- Commits: six commits in the range; each reports a good ECDSA signature for `oparin@me.com`.
- Canonical local `main`: clean, fast-forward aligned with `origin/main`, and at the declared base.
- Dependency `TASK-260720-g7kgox`: persisted as `done` before review.
- PR #18: open, non-draft, mergeable, merge-state `CLEAN`, and its remote head is exactly the reviewed SHA.
- GitHub Actions run 30683162729: completed successfully at exactly the reviewed SHA with 14/14 successful jobs: twelve test matrix jobs (Python 3.11–3.14 on Ubuntu, macOS, and Windows), strict mypy, and build artifacts. All four native Windows test jobs completed their test step successfully.

## Revalidation of the four prior findings

### 1. Per-context journal-generation uncertainty — resolved

`TransactionEngine.referenced_install_marker_groups` now returns a separate generation group for every relevant context target and includes live, staged, staged-source, backup, rollback, cleanup-sidecar, and tomb paths. GC scans each group before and after, marks only references from a fully valid group, and treats either a generation-state change or the absence of any valid marker generation for that target as uncertainty. Any such warning stops all runtime, snapshot, and protected-build sweeping.

The focused tests cover a valid project generation that cannot mask vanished global and hybrid targets, a wholly vanished generation, a generation-path race, a journal filesystem error, and corrupt journal/marker fail-safe retention.

### 2. Exact-object retirement on POSIX and Windows — resolved

POSIX GC retains the classified entry identity and uses held root/entry descriptors for the rename into quarantine. It compares the moved object's device, inode, owner, link count, mode, and timestamps to the classified identity, and restores or conservatively retains a mismatched object. Entry-path and driver-root exchange regressions prove a later pathname occupant is not retired.

Windows GC opens the selected source object with `DELETE` plus inspection rights, compares it with the previously classified object state, and renames that exact held source handle. The destination is rebound through a narrow held root handle and checked against the originally held quarantine identity/final path. Native entry exchange, live driver-root exchange, and destination quarantine-root exchange tests prove replacement objects remain live and an already-moved exact object is retained in the detached quarantine rather than deleted.

### 3. Indexed orphan names — resolved

The orphan matcher recognizes the canonical legacy and indexed transaction shapes for `.tmp`, `.backup`, and `.stale`, including PID/index variants, while rejecting malformed or merely similar live names. Focused tests cover live/dead PID behavior, indexed names, legacy names, and malformed shapes.

### 4. Native Windows status, GC, fault, and race coverage — resolved

The status suite includes a native-backend fresh-current test. The Windows backend suite includes native provenance/corruption/fault checks, lookup/publication races, old-versus-young GC behavior, quarantine behavior, source-entry exchange, live-root exchange, and destination-root exchange. These tests are guarded only when `os.name != "nt"`; the four successful Windows CI jobs therefore exercise the native branches. Commit `870daa3` is test-only: it restores the platform-neutral monkeypatch before native tests and updates the entry-exchange warning assertion to the actual protected-boundary fail-safe diagnostic.

## Native Windows rename ABI audit

The implementation's direct native call agrees with Microsoft's documented `NtSetInformationFile` / `FILE_RENAME_INFORMATION` contract:

- It uses information class 10 (`FileRenameInformation`).
- The structure fields and widths are `BOOLEAN ReplaceIfExists`, `HANDLE RootDirectory`, `ULONG FileNameLength`, and UTF-16 `WCHAR FileName`; the filename length is supplied in bytes and the buffer includes the structure plus the filename bytes.
- The source is an already-held handle opened with `DELETE`, so the operation targets the inspected source object instead of re-resolving its pathname.
- The destination-root reopening requests only `FILE_EXECUTE | FILE_READ_ATTRIBUTES` (`0x20 | 0x80`), then proves its object identity and final path match the held quarantine root before passing that handle as `RootDirectory`.
- `ReplaceIfExists` is false. Native collisions are translated to `FileExistsError` and retried with a fresh random child; other NT status failures are translated through `RtlNtStatusToDosError`.
- After success, source and destination-root handles are revalidated and the moved source must be the named direct child of the held destination root before any deletion is allowed.

Primary references:
- Microsoft, NtSetInformationFile: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntsetinformationfile
- Microsoft, FILE_RENAME_INFORMATION: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_rename_information
- Microsoft, File access rights constants: https://learn.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants

## Acceptance-criteria coverage

- Marker v1 schemas 1–5 preserve legacy currentness and runtime/snapshot GC compatibility.
- Marker v2 currentness recomputes the selected raw-snapshot identity and frozen context boundary; compares ref, installed content, build roots/source, toolchain, native target, complete cache key, canonical receipt, protected provenance, artifact path/hash/size, marker fields, and managed shim; and rechecks the marker/snapshot boundaries against concurrent change.
- `policy.execution_policy=manager-worker-v1` is explicitly described as a transitive cache-key/receipt dimension. Conformance vectors independently confirm the legacy key `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48` omits the policy and the reserved-hardened key `sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037` uses `hardened-worker-v1`; both are tested as non-current and rebuilt rather than adopted.
- Missing, corrupt, wrong-target, wrong-toolchain, wrong-policy, unsupported, context-leaking, or untrusted build state is non-current in text and JSON and makes `--check` fail.
- Capability evidence is attached only to the status result and is absent from the currentness input; an evidence error is surfaced without changing a clean verdict.
- A normal reinstall is the repair path. It replans from a revalidated frozen snapshot and publishes/reuses only protected bytes that pass the full expected-input, canonical-receipt, provenance, and artifact verification.
- Project and global status share the currentness implementation; global and hybrid marker-v2 stores are covered.
- GC runs under the manager-home lock, marks valid marker-v2 build keys plus every safe in-flight journal generation across project/global/hybrid stores, conservatively returns before every sweep on incomplete evidence, preserves marker-v1 runtime/snapshot compatibility, applies the grace rule, and lets the protected backend retire only proven unreferenced entries.
- The implementation fits the existing closure, installer, transaction, build-cache, shim, and CLI boundaries; it introduces no package-controlled repair hook or separate execution-policy mechanism.

Protocol fact-check: CocoaSkills was compared with Curator Protocol commit `f5d7673039226ab81de2f4f87e2155ae995c4df3`, especially `spec/v1/manager.md`, decision 0004, and `conformance/v1/vectors/build-drivers.json`. Those sources define the complete build input/receipt and protected-boundary checks, normal-install repair, fail-safe journal-aware GC, policy execution identity, and the two non-current policy vectors reflected above.

## Independent verification

- Focused maintenance suite:
  `python -m pytest tests/test_build_currentness.py tests/test_build_gc.py tests/test_status.py tests/test_gc.py tests/test_installer_transactions.py tests/test_global_install_transactions.py tests/test_build_cache_posix.py tests/test_build_cache_windows.py -q`
  → **114 passed, 27 skipped in 94.41s**.
- Strict typing: `python -m mypy` → **Success: no issues found in 68 source files**.
- Patch hygiene: `git diff --check <base>..<head>` → clean.
- Exact-head CI: https://github.com/ivanopcode/cocoaskills/actions/runs/30683162729
- Reviewed PR: https://github.com/ivanopcode/cocoaskills/pull/18

## Findings

No remaining blocking or non-blocking implementation finding. The appropriate explicit reviewer branch is **accepted → done**.
