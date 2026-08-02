# TASK-260720-th0jdi — Review rework evidence (RUN-260801-a7ef1b)

## Provenance and handoff boundary

- Task: `TASK-260720-th0jdi` — Implement build currentness, repair, and GC.
- Clean recorded base and merge base:
  `07655553cebcf867bbe58629de98e77644606c85`.
- Previously reviewed implementation head:
  `f3c1254e4fe7958c720cbef096a4ef00103d43a2`.
- Signed rework commits:
  - `0314ab5939b51cf33f0fdbb68c9df90631e84f5a` — journal,
    retirement-race, orphan, and native-coverage hardening;
  - `4cd158990da1ae030d1115f9d0ca4cb3310cb60e` — diagnostic
    Win32 root-access isolation;
  - `f7f65995fd934b2bc446812363380f806afe195e` — native
    `NtSetInformationFile` correction;
  - `f99b5bc378ea7c404db0d04e4205f86b979a9177` — dedicated
    `FILE_TRAVERSE | FILE_READ_ATTRIBUTES` destination-root binding;
  - `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09` — test-harness cleanup
    ordering and fail-safe warning expectation correction after native Windows
    execution proved the product path.
- `git show --show-signature HEAD` reports a good ECDSA signature for
  `oparin@me.com`.
- Existing pull request: [ivanopcode/cocoaskills#18](https://github.com/ivanopcode/cocoaskills/pull/18).
- Authoritative exact-head run:
  [30683162729](https://github.com/ivanopcode/cocoaskills/actions/runs/30683162729)
  for `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09` — success. Python
  3.11–3.14 passed on Ubuntu, macOS, and Windows; strict mypy and artifact
  build also passed. `gh run watch --exit-status` exited `0`.
- No merge, tag, or release was performed.

## Reviewer findings addressed

### Per-journal and per-context marker generations

Transaction journals expose one marker-generation group for every context
target. Each group covers its live, staged, staged-source, backup, rollback,
cleanup, and cleanup-tomb paths. GC snapshots every path before and after
reading, requires at least one valid marker in each group, and merges references
only after all groups are proven safe.

A vanished generation group, corrupt or unreadable marker, path race,
journal-root race, or journal filesystem `OSError` makes the whole mark phase
uncertain. Runtime, snapshot, and build sweeps are all suppressed. Regression
coverage includes intact staged state, all generations vanished, disappearance
during scan, forced journal-root error, and a valid project target alongside
vanished global and hybrid targets so one context cannot mask another.

### Identity-bound protected-entry retirement

The POSIX collector carries classified directory state into quarantine,
selects through the held driver descriptor, and verifies the quarantined inode
before deletion. A pathname exchange is restored or retained; a driver-root
exchange can affect only the entry reached through the previously held root
descriptor and never a replacement live root.

The Windows collector opens the exact classified source with `DELETE`, reopens
the selected quarantine root with the documented narrow access mask
`FILE_TRAVERSE | FILE_READ_ATTRIBUTES` (`0xA0`), verifies its volume/file ID and
final path against the already-held parent, and calls native
`NtSetInformationFile(FileRenameInformation=10)`. The record uses native
`BOOLEAN`/aligned `HANDLE`/`ULONG`/UTF-16 widths, `ReplaceIfExists=FALSE`, an
exact `IO_STATUS_BLOCK`, and `RtlNtStatusToDosError` for failing statuses.

The native destination-root handle remains open across the syscall, so both
source and destination are kernel-bound. A platform-neutral contract test
asserts the source handle, root handle and `0xA0` rights, class 10, UTF-16 byte
length, buffer size, and no-replace flag. Native Windows tests exchange the
entry, driver root, and destination quarantine root at deterministic points;
young or replacement state is retained and never deleted.

### Exact runtime orphan compatibility

Runtime orphan matching recognizes only manager-owned legacy and indexed
temporary/backup names plus indexed stale names. PID and allocation indexes
must have canonical numeric shape. Live-PID entries and malformed,
leading-zero, or unrelated dotfiles are retained; only exact dead-PID names
are swept.

### Native Windows status and GC coverage

The fake-build fixture derives the native target and `.exe` suffix on Windows.
An unskipped status vector installs through the protected backend and proves a
fresh marker-v2 build current in the object model and JSON. Build-GC
grace/mark/sweep integration also runs natively. Backend tests cover normal
old/young collection, publication quarantine, entry/root/destination-root
exchange, fault retention, and exact-object behavior.

### Accepted currentness and repair behavior preserved

Marker-v1 schemas 1–5 remain current under their existing rules. Marker-v2
status rederives build source from the selected persistent raw snapshot and
static context boundary, then checks ref, installed bytes, roots/context
exclusion, source identity, toolchain, target, `manager-worker-v1` execution
policy, full cache key, protected provenance, canonical receipt, artifact path,
hash and size, marker build fields, and managed shim.

Policy-less legacy and reserved-hardened keys remain non-current and are
rebuilt. Capability evidence remains result-only. Normal install remains the
repair operation and rebuilds from a freshly frozen/revalidated snapshot
without adopting candidate bytes.

## Independent Windows diagnosis and fact-checking

The task consumed the independent board outcome
`BUG-260801-nzpar0_windows-handle-rename-diagnosis.md`. It isolated the two
expected-red Server 2025 runs, verified x86/x64 record layouts, ruled out
source access/share, root type/access, UTF-16 length, buffer size, collision,
and cross-volume hypotheses, and recommended the native class-10 call with the
dedicated held root used by the final patch.

Primary sources:

- Microsoft documents native class 10, `DELETE`, `IO_STATUS_BLOCK`, raw
  status, and the user-mode `NtSetInformationFile` name in
  [`NtSetInformationFile`](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntsetinformationfile).
- Microsoft documents the native layout, held `RootDirectory`, relative name,
  byte length, no-replace behavior, and minimum buffer size in
  [`FILE_RENAME_INFORMATION`](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_rename_information).
- The current Win32 page advertises relative-root support in
  [`FILE_RENAME_INFO`](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info),
  while Microsoft's [2022 correction](https://github.com/MicrosoftDocs/sdk-api/commit/ada04eef90bc7ebe441ce2ef938867d3a677d57d)
  records that the Win32 field was effectively unusable. The
  [2026 docs-only reversal](https://github.com/MicrosoftDocs/sdk-api/commit/d1debc569f40cda761474d216903f27c8aa7c7af)
  supplied no Server 2025 implementation evidence. Both exact hosted runs
  observed error 87 on the Win32 relative-root path.
- Microsoft documents `FILE_EXECUTE == FILE_TRAVERSE == 0x20` and
  `FILE_READ_ATTRIBUTES == 0x80` in
  [file access-right constants](https://learn.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants).
- Task behavior was checked against Curator manager requirements for
  [status and locked fail-safe GC](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L657-L684),
  [cache acceptance/rebuild](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L329-L340),
  and [normal-install repair](https://github.com/relux-works/curator-spec/blob/f5d7673039226ab81de2f4f87e2155ae995c4df3/profiles/manager.md#L444-L463).

The requested `ssh win` host was externally unavailable. `ssh` exited `255`
after TCP timeout, ICMP ping exited `2`, and `tailscale ping` exited `1`; the
operator explicitly directed the researcher not to wait. The ready native
matrix is attached to the diagnostic bug, but its predicted output is not
reported as observed evidence. The exact-head GitHub Actions Windows jobs are
the authoritative native gate.

## Frozen-diff local validation

Every gate below was launched as a standalone process without `tee` or an
exit-status-obscuring pipe.

- Focused status/maintenance/transaction/backend suite — exit `0`:
  `114 passed, 27 skipped in 90.05s`.
- Full `python -m pytest -q` — exit `0`:
  `1206 passed, 104 skipped in 273.14s`.
- `python -m mypy --strict src/csk` — exit `0`:
  no issues in 68 source files.
- `python -m compileall -q src tests` — exit `0`.
- `python -m build --outdir
  /var/folders/v5/c7x5kxfd2dvc6gmb1772xdp00000gn/T/csk-TASK-260720-th0jdi-exact-head-build.XXXXXX.IeB7RGyPUz`
  — exit `0`; sdist
  and wheel built.
- `python -m twine check
  /var/folders/v5/c7x5kxfd2dvc6gmb1772xdp00000gn/T/csk-TASK-260720-th0jdi-exact-head-build.XXXXXX.IeB7RGyPUz/*`
  — exit `0`; both
  artifacts passed.
- Platform-neutral native-call contract regression — exit `0`:
  `1 passed in 0.09s`.
- `git diff --check` and committed-diff checks — exit `0`.
- Signed commit verification for `870daa3` — good ECDSA signature.
- Push of `870daa3` to the existing PR branch — exit `0`.
- Exact-head hosted run `30683162729` — success; its direct
  `gh run watch --exit-status` gate exited `0`.
- `gh pr checks 18` — exit `0`; all 14 declared checks passed.
- `gh pr view 18` — exit `0`; the PR is open, non-draft, mergeable, and its
  head is exactly `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09`.

The optional Ruff probes each exited `1` because Ruff is not installed or
declared by this project (`No module named ruff`). They were not treated as
project lint gates; strict mypy is the declared lint/type gate and is green.

## Expected-red and superseded evidence

These results are intentionally reported as failures/cancellations:

- First rework-focused run — exit `1`: `2 failed, 88 passed, 25 skipped`.
  The tests lacked one import and attempted a sealed-directory exchange
  without first applying a mutable test mode. Direct rerun after correcting
  those test defects exited `0`, followed by the green focused/full gates.
- Diagnostic single-test `--pdb` run — exit `2` after intentionally quitting
  the debugger; it was not a validation gate.
- Exact run [30679822247](https://github.com/ivanopcode/cocoaskills/actions/runs/30679822247)
  at `0314ab5` — failure: the representative Windows job exited `1` with
  `5 failed, 1236 passed, 185 skipped`; all five failures reached Win32
  `SetFileInformationByHandle(FileRenameInfo)` error 87.
- Exact run [30680673688](https://github.com/ivanopcode/cocoaskills/actions/runs/30680673688)
  at `4cd1589` — failure: the representative Windows job exited `1` with
  the same `5 failed, 1236 passed, 185 skipped` and error 87 after the narrow
  root reopen, isolating the Win32 call path rather than the access mask.
- The local watcher for superseded run
  [30682140560](https://github.com/ivanopcode/cocoaskills/actions/runs/30682140560)
  was interrupted with exit `130` after new evidence required the dedicated
  native root handle. GitHub then canceled that remote run through workflow
  concurrency when `f99b5bc` was pushed. It is not counted as a gate.
- Exact run [30682551287](https://github.com/ivanopcode/cocoaskills/actions/runs/30682551287)
  at `f99b5bc` — failure. Native Windows retirement reached and passed the
  corrected class-10 product path; the Python 3.13 job then reported
  `1 failed, 1242 passed, 185 skipped, 1 error`. Both defects were confined to
  tests: one platform-neutral mock survived into autouse cleanup because of
  monkeypatch teardown order, and one race regression expected an earlier
  warning even though the broader fail-safe correctly reported
  `protected boundary is uncertain`. Commit `870daa3` corrects only those
  harness expectations.

## Review handoff condition

The exact-head `30683162729` matrix and artifact build are green, the frozen
worktree is clean, and every requested developer gate has passed. The branch
remains unmerged and no tag or release has been created.
