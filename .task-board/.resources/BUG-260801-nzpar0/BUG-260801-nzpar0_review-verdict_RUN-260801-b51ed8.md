# BUG-260801-nzpar0 reviewer verdict — RUN-260801-b51ed8

Verdict branch: **accepted**

## Outcome

The diagnostic work satisfies the task after the recorded orchestrator route override that accepted exact hosted-Windows isolation instead of waiting for the offline `ssh win` host. The report is technically cautious: it proves the repeated Win32 error boundary and rules out the listed ABI, rights, type, access/share, encoding, length, sizing, collision, and cross-volume hypotheses, while explicitly leaving the unrun native class-10 contrast as residual implementation risk.

## Independent evidence

- GitHub run metadata independently resolves run `30679822247` to head `0314ab5939b51cf33f0fdbb68c9df90631e84f5a` and run `30680673688` to head `4cd158990da1ae030d1115f9d0ca4cb3310cb60e`. Both Python 3.13 jobs ran `python -m pytest -v` on Microsoft Windows Server 2025 `10.0.26100`, image `windows-2025-vs2026`, and ended with the same five failures and native process exit 1. The two direct failures report WinError 87 for both a regular-file boundary move and a directory quarantine move.
- `gh api repos/ivanopcode/cocoaskills/compare/0314ab5939b51cf33f0fdbb68c9df90631e84f5a...4cd158990da1ae030d1115f9d0ca4cb3310cb60e` reports one commit and one changed product file. Its patch only replaces the original destination root with a fresh `FILE_READ_ATTRIBUTES | FILE_EXECUTE` (`0xa0`) root, verifies identity/final path, and revalidates after the same `SetFileInformationByHandle` call.
- Exact source inspection confirms source desired access `READ_CONTROL | FILE_READ_ATTRIBUTES | DELETE` (`0x00030080`), share-all (`0x7`), disk-object checking, directory checking on the selected destination parent, and volume/file-ID plus final-path identity checks.
- Reviewer reruns of both Clang MSVC target layouts exited 0. Win64 DWORD and BOOLEAN forms both produce offsets `0/8/16/20`, size 24, alignment 8; Win32 forms both produce `0/4/8/12`, size 16, alignment 4. Read-only Python compilation of the attached matrix exited 0.
- The board report and `.research/260801_windows-handle-rename-winerror87.md` have the same SHA-256: `32a89ca998232935c9cb830556f52c3d0d4c01e65698295d96cd695e720e8b13`. `git diff --check` passed, and no Curator product source/test/module files are changed by this task.
- The handoff is present on `TASK-260720-th0jdi` and names the retained `0xa0` root, native class 10, exact ABI, raw status translation, and no-replace requirement.

## Primary-contract fact check

- Current Win32 `FILE_RENAME_INFO` documents a relative `FileName` with a directory `RootDirectory`, byte length, and no required terminator: https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info
- Native `FILE_RENAME_INFORMATION` documents the BOOLEAN/ULONG union, held target root, traverse plus read-attribute rights, simple relative name, DELETE requirement, no-replace behavior, and minimum buffer size: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_rename_information
- `NtSetInformationFile` documents information class 10, the IO status block, DELETE access, raw status return, and the user-mode `NtSetInformationFile` name: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntsetinformationfile
- `RtlNtStatusToDosError` documents NTSTATUS-to-system-error translation and the absence of an inverse mapping: https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-rtlntstatustodoserror
- `FILE_ID_INFO` documents volume serial plus 128-bit file ID as the handle-identity comparison: https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_id_info
- MicrosoftDocs commit `ada04eef90bc7ebe441ce2ef938867d3a677d57d` records the 2022 Win32 implementation warning; commit `d1debc569f40cda761474d216903f27c8aa7c7af` is the later docs-only reversal. The report accurately identifies the conflict and does not infer an OS fix from the documentation change.

## Architecture and risk

The recommended correction is the smallest identity-safe boundary change: retain the exact open source and verified destination-root handles, replace only the failing Win32 wrapper call with native `NtSetInformationFile(FileRenameInformation=10)`, keep `ReplaceIfExists=FALSE`, use the exact native ABI and `IO_STATUS_BLOCK`, and translate failures through `RtlNtStatusToDosError`. A NULL-root absolute path would reintroduce pathname resolution during the operation and would not preserve held-root identity.

Residual risk is explicit and acceptable for this diagnosis task: `ssh win` remained offline, so granted-access/raw-NTSTATUS output and a successful native class-10 control are not claimed. The attached matrix is the required next Windows implementation gate. Exact-head product pytest remains expected-red evidence of the diagnosed defect; the diagnostic artifact validations themselves are green.

No product edit, commit, merge, tag, or release was performed by this task.