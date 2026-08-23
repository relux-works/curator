# BUG-260801-nzpar0 — Windows held-root rename / WinError 87 diagnosis

Date: 2026-08-01
Role: researcher
Scope: CocoaSkills PR #18, exact heads `0314ab5939b51cf33f0fdbb68c9df90631e84f5a` and `4cd158990da1ae030d1115f9d0ca4cb3310cb60e`
Product-code changes by this task: none

## Executive finding

The repeated `ERROR_INVALID_PARAMETER` (`87`) is isolated to the Win32
`SetFileInformationByHandle(FileRenameInfo)` call path when CocoaSkills supplies
a non-NULL `FILE_RENAME_INFO.RootDirectory` and a simple relative target name.
It is not explained by the ctypes field offsets/padding, root-directory access
requested at `4cd1589`, root object type, source `DELETE` access/share mode,
UTF-16LE encoding, byte length, target component validity, or buffer size.

The strongest exact-platform evidence is the same five-test failure on hosted
Microsoft Windows Server 2025 (build `10.0.26100`) at both heads. The second
head changes only this rename path, reopens the destination root with exactly
`FILE_TRAVERSE | FILE_READ_ATTRIBUTES` (`0x000000a0`), verifies that handle as
the selected directory identity, and still receives error 87. Target-layout
compilation shows the current DWORD/union ctypes form and the legacy BOOLEAN
form are byte-compatible for `ReplaceIfExists == FALSE` on both Win32 and
Win64.

There is a material conflict inside Microsoft's own primary documentation:

- The [current Win32 `FILE_RENAME_INFO` page](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info)
  says a relative name can use a directory handle in `RootDirectory`.
- Microsoft's [2022 documentation correction](https://github.com/MicrosoftDocs/sdk-api/commit/ada04eef90bc7ebe441ce2ef938867d3a677d57d)
  instead says the field should be NULL; its commit rationale records that the
  field was effectively unusable and inconsistent in the then-current Win32
  implementation.
- The [2026 docs-only reversal](https://github.com/MicrosoftDocs/sdk-api/commit/d1debc569f40cda761474d216903f27c8aa7c7af)
  came through [PR #1123](https://github.com/MicrosoftDocs/sdk-api/pull/1123),
  whose submitter explicitly expressed uncertainty. Its attached patch changes
  only the Markdown contract and supplies no Windows implementation change or
  Server 2025 regression evidence.

The evidence therefore supports a Win32-wrapper/call-path incompatibility for
this relative-root form on the tested Server 2025 image. Because Win32 exposes
only error 87 and Microsoft documents no inverse Win32-error-to-NTSTATUS
mapping, this evidence does **not** prove which internal NTSTATUS (or pre-syscall
wrapper validation) produced 87.

### Minimal identity-safe recommendation

Replace only the failing Win32 invocation with the documented native
`NtSetInformationFile(..., FileRenameInformation=10)` call while retaining:

1. the exact already-open source handle with `DELETE` access;
2. the exact verified destination-root handle opened with
   `FILE_TRAVERSE | FILE_READ_ATTRIBUTES`;
3. the same simple UTF-16LE relative component and byte count;
4. `ReplaceIfExists = FALSE`; and
5. existing pre/post handle-identity validation.

Use the native `FILE_RENAME_INFORMATION` representation (`BOOLEAN`, aligned
`HANDLE`, `ULONG`, trailing UTF-16 code unit) and an exact `IO_STATUS_BLOCK`.
Treat `NTSTATUS >= 0` as success and translate a failing status with
`RtlNtStatusToDosError` to preserve the current exception surface. Classic
class 10 is sufficient; the Ex class and POSIX/replace flags are not needed.

Do **not** ship the diagnostic `RootDirectory=NULL` plus absolute-path form as
the correction. That form can demonstrate the wrapper boundary, but it resolves
the destination pathname at operation time and therefore cannot preserve the
held-root identity guarantee across a concurrent directory exchange.

## Evidence boundary

The assigned `ssh win` host was offline. The operator explicitly directed this
researcher not to wait for it and to hand off from exact hosted-Windows logs,
primary contracts, and a ready follow-up probe.

Exact connectivity attempts:

```text
$ ssh -o BatchMode=yes -o ConnectTimeout=10 win hostname
ssh: connect to host 100.120.84.42 port 22: Operation timed out
exit 255

$ ping -c 2 -W 1000 100.120.84.42
2 packets transmitted, 0 packets received, 100.0% packet loss
exit 2

$ tailscale ping -c 3 mbpro-win
... three timeouts; no reply
exit 1
```

`tailscale status` reported `mbpro-win` (`100.120.84.42`) offline, last seen one
day earlier. No live `ssh win` ABI/handle matrix was run, and this report does
not represent the ready probe's expected results as observed results. Exact
Server 2025 reproduction instead comes from the two PR jobs below.

## Exact-head Windows Server 2025 reproduction

The jobs are [0314ab5 / Python 3.13](https://github.com/ivanopcode/cocoaskills/actions/runs/30679822247/job/91314466425)
and [4cd1589 / Python 3.13](https://github.com/ivanopcode/cocoaskills/actions/runs/30680673688/job/91316908346).
GitHub identifies each run's `headSha` as the stated PR head. The checkout step
tests GitHub's merge of that exact head into unchanged base
`07655553cebcf867bbe58629de98e77644606c85`.

### Head `0314ab5`

Retrieval command (standalone shell with `pipefail`):

```bash
set -o pipefail
gh api repos/ivanopcode/cocoaskills/actions/jobs/91314466425/logs | \
  rg -n "Windows Server 2025|10\\.0\\.26100|Merge 0314ab5|Errno 87|5 failed, 1236 passed|Process completed with exit code 1"
```

Extraction exit: `0`. Native test process: expected-red exit `1`.

```text
Microsoft Windows Server 2025
HEAD is now at 208e0d5 Merge 0314ab5939b51cf33f0fdbb68c9df90631e84f5a into 07655553cebcf867bbe58629de98e77644606c85
OSError: [Errno 87] cannot atomically move open cache object to boundary-builds-b20192b73727819d: The parameter is incorrect.
OSError: [Errno 87] cannot atomically move open cache object to entry-77f971541a049d1866f2b6fd5da69d4575ec174e0f1b181d11521a6d09f2c0a9-938ed1ace6d050a7: The parameter is incorrect.
5 failed, 1236 passed, 185 skipped in 997.14s
Process completed with exit code 1.
```

### Head `4cd1589`

Retrieval command:

```bash
set -o pipefail
gh api repos/ivanopcode/cocoaskills/actions/jobs/91316908346/logs | \
  rg -n "Windows Server 2025|10\\.0\\.26100|Merge 4cd1589|Errno 87|5 failed, 1236 passed|Process completed with exit code 1"
```

Extraction exit: `0`. Native test process: expected-red exit `1`.

```text
Microsoft Windows Server 2025
10.0.26100
Image: windows-2025-vs2026
HEAD is now at acd99cd Merge 4cd158990da1ae030d1115f9d0ca4cb3310cb60e into 07655553cebcf867bbe58629de98e77644606c85
OSError: [Errno 87] cannot atomically move open cache object to boundary-builds-a3a1af6e84c996e7: The parameter is incorrect.
OSError: [Errno 87] cannot atomically move open cache object to entry-77f971541a049d1866f2b6fd5da69d4575ec174e0f1b181d11521a6d09f2c0a9-19dfe94ffc87f860: The parameter is incorrect.
5 failed, 1236 passed, 185 skipped in 847.13s
Process completed with exit code 1.
```

The first direct failure moves an ordinary file used as a deliberately invalid
cache boundary. The second moves a cache-entry directory. Repeating 87 for both
object kinds rules out a directory-only rename restriction.

## Hypothesis matrix

| Plausible cause | Exact evidence | Finding |
|---|---|---|
| Wrong ctypes layout or padding | Windows-target compiler layouts below; DWORD and BOOLEAN forms have identical offsets/sizes for both architectures | Ruled out for the zero/no-replace record used here |
| Wrong `RootDirectory` access | `4cd1589` opens a fresh root with `0x20 | 0x80`, then gets the same 87 as `0314ab5` | Ruled out as the cause of these runs |
| Wrong root handle type | `_open_raw_handle` succeeds, `_handle_details` requires `FILE_TYPE_DISK`, and `FileStandardInfo.directory` is checked by the protected-directory opener | Ruled out by the call path before rename |
| Source lacks `DELETE` | Source is opened with `0x00030080`, including `DELETE=0x00010000`; open succeeds | Ruled out |
| Source sharing conflict | Source uses share mask `0x7` (read/write/delete); a pre-existing incompatible handle would make `CreateFileW` fail with sharing violation before rename | Ruled out |
| Wrong encoding/length | Simple ASCII components become valid UTF-16LE; declared length is exact bytes (examples below) | Ruled out |
| Buffer too small or missing NUL storage | Win64 allocation is `24 + name_bytes`, as required by native docs; name starts at 20, leaving four zero bytes after it | Ruled out |
| Existing target with no-replace | Generated random targets are absent; documented collision errors are 80/183 at Win32 or `STATUS_OBJECT_NAME_COLLISION` natively, not repeated 87 | Ruled out for observed cases |
| Cross-volume rename | Both parents are below each test's same `tmp_path` root | Ruled out |
| Win32 relative-root handling | Only common feature left after the above; matches Microsoft's 2022 implementation warning and exact Server 2025 behavior | Leading diagnosis; raw internal status unavailable from Win32 |

## ABI and buffer proof

Microsoft's current [`FILE_RENAME_INFO`](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info)
header shape starts with a `BOOLEAN`/`DWORD` union, followed by `HANDLE`,
`DWORD FileNameLength`, and `WCHAR FileName[1]`. The native
[`FILE_RENAME_INFORMATION`](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/ns-ntifs-_file_rename_information)
has the corresponding `BOOLEAN`/`ULONG` union and states that the buffer must be
at least `sizeof(FILE_RENAME_INFORMATION) + FileNameLength` bytes.

Task-scoped target-layout source:
`.temp/BUG-260801-nzpar0/rename_abi_target.c` (also attached as an outcome).

```bash
clang -target x86_64-pc-windows-msvc -Xclang -fdump-record-layouts \
  -fsyntax-only .temp/BUG-260801-nzpar0/rename_abi_target.c
```

Exit `0`:

```text
FILE_RENAME_INFO_DWORD:
  ReplaceOrFlags=0 RootDirectory=8 FileNameLength=16 FileName=20
  sizeof=24 align=8
FILE_RENAME_INFO_BOOLEAN:
  ReplaceIfExists=0 RootDirectory=8 FileNameLength=16 FileName=20
  sizeof=24 align=8
```

```bash
clang -target i686-pc-windows-msvc -Xclang -fdump-record-layouts \
  -fsyntax-only .temp/BUG-260801-nzpar0/rename_abi_target.c
```

Exit `0`:

```text
FILE_RENAME_INFO_DWORD:
  ReplaceOrFlags=0 RootDirectory=4 FileNameLength=8 FileName=12
  sizeof=16 align=4
FILE_RENAME_INFO_BOOLEAN:
  ReplaceIfExists=0 RootDirectory=4 FileNameLength=8 FileName=12
  sizeof=16 align=4
```

At `4cd1589`, `_FileRenameInfo` uses `c_uint32`, `c_void_p`, `c_uint32`,
and `c_wchar * 1`. Native Windows Python has a two-byte `wchar_t`, so on the
tested 64-bit interpreter this is the first Win64 layout above. The buffer is
zero-created, `ReplaceIfExists` is zero, and therefore the three padding bytes
in the legacy BOOLEAN interpretation are also zero. Switching only the first
field from DWORD to BOOLEAN cannot alter this call's bytes or offsets.

For the two exact failing names:

```text
boundary-builds-a3a1af6e84c996e7
  characters=32, UTF-16LE bytes=64, Win64 buffer=24+64=88

entry-77f971541a049d1866f2b6fd5da69d4575ec174e0f1b181d11521a6d09f2c0a9-19dfe94ffc87f860
  characters=87, UTF-16LE bytes=174, Win64 buffer=24+174=198
```

The name begins at offset 20, so both buffers retain four zero bytes after the
encoded name. This accommodates even the Win32 page's separate statement that
`FileName` is NUL-terminated, while `FileNameLength` correctly excludes that
terminator. The current Win32 page and native contract both define the length
in bytes; the current Win32 page says a terminator is not required in the
length.

For implementation clarity, the native patch should use `c_uint16 * 1` for the
trailing Windows code unit rather than host-dependent `c_wchar * 1`; this is a
portability hardening for import/layout checks, not the cause on native Windows.

## Handle access, type, and identity

### Source

Both heads call `CreateFileW` for the exact source with:

```text
desired access = READ_CONTROL | FILE_READ_ATTRIBUTES | DELETE
               = 0x00020000 | 0x00000080 | 0x00010000
               = 0x00030080
share mode     = FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE
               = 0x00000007
flags          = FILE_FLAG_OPEN_REPARSE_POINT | FILE_FLAG_BACKUP_SEMANTICS
```

Microsoft's [CreateFileW contract](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew)
says `FILE_SHARE_DELETE` permits later delete access and that delete access
allows rename. It also says incompatible access/share combinations fail the
open with `ERROR_SHARING_VIOLATION`; these source opens succeeded. The native
rename contract independently requires `DELETE`, which is present.

### Destination root

At `0314ab5`, the already-held protected destination parent was passed. The
only `0314ab5..4cd1589` product diff reopens that same path with:

```text
FILE_EXECUTE | FILE_READ_ATTRIBUTES = 0x20 | 0x80 = 0x000000a0
```

The exact comparison command was:

```bash
git diff --unified=20 \
  0314ab5939b51cf33f0fdbb68c9df90631e84f5a \
  4cd158990da1ae030d1115f9d0ca4cb3310cb60e -- \
  src/csk/builds/cache_windows.py
```

It exited `0` and showed no change to the record, name, source handle, buffer,
or API call; the material addition was the `0xA0` root reopen plus
identity/final-path validation shown above.

For directories, Microsoft's [access-right constants](https://learn.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants)
define `FILE_EXECUTE` and `FILE_TRAVERSE` as the same `0x20` bit and
`FILE_READ_ATTRIBUTES` as `0x80`. The native rename page specifically calls
out traverse plus read-attribute for a `RootDirectory` handle. The second head
then compares volume/file ID and final path against the selected held parent,
requires a disk-directory object on open, and passes that exact live handle.
The unchanged error rules out both the prior rights hypothesis and an ordinary
wrong-object handle.

Microsoft's [`FILE_ID_INFO` contract](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_id_info)
states that volume serial plus 128-bit file ID identifies a file and is the
prescribed comparison for determining whether two handles represent the same
file. CocoaSkills uses that pair for its identity comparison.

The available CI logs do not print numeric handle values or a queried
`GrantedAccess` mask. The attached unrun matrix uses `NtQueryObject`,
`GetFileType`, `FileStandardInfo`, `FILE_ID_INFO`, and
`GetFinalPathNameByHandleW` to provide that evidence when a Windows host is
available. This limitation is explicit; source-requested access is not being
misreported as separately instrumented granted-access output.

## Why error 87 does not identify a native status

Microsoft defines [`ERROR_INVALID_PARAMETER` as 87](https://learn.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-).
[`RtlNtStatusToDosError`](https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-rtlntstatustodoserror)
maps NTSTATUS to a Win32 system error and explicitly documents that there is no
inverse function. Therefore the Win32 result cannot prove whether the wrapper
rejected the record before a syscall or translated one of multiple native
statuses.

As corroborating—not local-NTFS-dispositive—evidence, Microsoft's
[`MS-FSCC FileRenameInformation` table](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/1d2673a8-8fb9-4868-920a-775ccaa30cf8)
distinguishes invalid name/length (`STATUS_INVALID_PARAMETER`), missing delete
access (`STATUS_ACCESS_DENIED`), existing no-replace target
(`STATUS_OBJECT_NAME_COLLISION`), cross-device destination, and record-length
mismatch. It also says a nonzero root is invalid for a **network** operation;
the failing CI paths are local `C:` temporary paths, so that network-only rule
does not itself explain these runs.

The native patch is diagnostically better as well as identity-safe:
[`NtSetInformationFile`](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntsetinformationfile)
returns the raw status, takes an `IO_STATUS_BLOCK`, documents class 10 and its
`DELETE` requirement, and explicitly instructs user-mode callers to use the
`NtSetInformationFile` name.

## Concrete patch shape (recommendation only)

No product code was edited by this research task. The smallest correction to
`4cd1589` is:

1. Add `ntdll.NtSetInformationFile` and `ntdll.RtlNtStatusToDosError` ctypes
   signatures.
2. Define native `FILE_RENAME_INFORMATION` with:
   `c_ubyte`, naturally aligned `c_void_p`, `c_uint32`, `c_uint16 * 1`.
3. Define `IO_STATUS_BLOCK` as a pointer-sized union of signed 32-bit status or
   pointer, followed by pointer-sized `Information`.
4. Keep the `4cd1589` dedicated `rename_root` open with `0x000000a0`, including
   its identity/final-path comparison and post-call revalidation.
5. Pass `source.value`, the IO status block, the existing buffer, its full byte
   size, and information class `10`.
6. Preserve `ReplaceIfExists=0`; translate a failing native collision into the
   existing `FileExistsError` retry path and all other statuses into the
   existing Windows error surface.

If instead the long-held destination handle is passed directly, ensure that
handle was originally opened with `FILE_TRAVERSE | FILE_READ_ATTRIBUTES`.
The current general protected-directory opener requests `GENERIC_READ`, which
does not by itself document the exact specialized root-right pair. Retaining
the narrow `4cd1589` rename-root open avoids broadening unrelated handles.

An absolute pathname with NULL root is not an identity-safe substitute:
prevalidation proves what the pathname selected before the call, and
postvalidation proves what a handle selected afterward, but neither binds the
kernel's destination lookup to the already-held directory during the rename.

## Reproducible follow-up matrix

The task-scoped `rename_matrix.py` outcome is syntax-checked and ready for a
native Windows host. It creates one PID-scoped temporary tree, prints JSONL,
and removes that tree. It compares:

- Win32 class 3 relative root using modern/legacy/packed layouts;
- byte length, character length, and length-including-NUL variants;
- `sizeof + bytes` versus `offsetof(FileName) + bytes` buffers;
- Win32 NULL-root absolute DOS and extended paths;
- native class 10 relative-root file and directory moves;
- root access variants (`READ_ATTRIBUTES`, `TRAVERSE`, both, generic read);
- source without `DELETE`, a no-replace collision, wrong root type, and an
  incompatible share control; and
- requested/granted access, disk/directory type, final path, volume/file ID,
  raw NTSTATUS, translated WinError, and before/after source identity.

Intended commands when `ssh win` returns (not run in this task):

```powershell
python rename_matrix.py
cl /nologo /W4 rename_abi.c /Fe:rename_abi.exe
.\rename_abi.exe
```

Expected outcomes are deliberately not asserted here. The decisive acceptance
for the recommendation is: Win32 relative-root repeats 87 while native class
10 succeeds with the same source/root identities; the collision control fails
without replacement; missing `DELETE` and invalid root controls return their
own native statuses.

## Validation ledger

All gates were launched directly without `tee`. Pipelines used for log
selection set `pipefail`.

| Command / evidence | Exit | Interpretation |
|---|---:|---|
| `ssh -o BatchMode=yes -o ConnectTimeout=10 win hostname` | 255 | Host offline; unavailable probe, not a pass |
| `ping -c 2 -W 1000 100.120.84.42` | 2 | No packets returned; unavailable probe, not a pass |
| `tailscale ping -c 3 mbpro-win` | 1 | No peer reply; unavailable probe, not a pass |
| 0314 job-log extraction shown above | 0 | Evidence retrieved |
| 0314 native pytest process | 1 | Expected-red exact-head reproduction: five failures, including error 87 |
| 4cd job-log extraction shown above | 0 | Evidence retrieved |
| 4cd native pytest process | 1 | Expected-red exact-head reproduction after root-right fix |
| Exact `0314ab5..4cd1589` source diff | 0 | Root-right/identity change isolated |
| x86_64 Windows-target Clang ABI layout | 0 | Layout evidence produced |
| i686 Windows-target Clang ABI layout | 0 | Layout evidence produced |
| `python3 -m py_compile .temp/BUG-260801-nzpar0/rename_matrix.py` | 0 | Follow-up probe syntax valid |
| UTF-16LE example encoder command | 0 | Exact character/byte counts produced |
| Initial artifact trailing-whitespace check | 1 | Three Markdown hard-break lines were reported; spaces removed |
| Artifact trailing-whitespace rerun | 0 | Clean after correction |
| `git diff --check -- LOGBOOK.md` | 0 | Logbook edit has no whitespace errors |
| `task-board validate` | 0 | Command itself passed; it also reported 1,227 pre-existing board-wide broken-link/status/resource findings outside this task |

One discarded diagnostic extraction used `rg -m` with `pipefail`; `rg` closed
the upstream stream after its match limit and the pipeline truthfully exited
`141`. It was not treated as a gate. The same extraction was rerun without
`-m` and exited `0`, producing the evidence quoted above.

## Decision and residual risk

Recommendation: use native class 10 with the verified `0xA0` root handle and
preserve all identity/no-replace checks. This is the narrowest correction that
honors the platform's documented held-root contract and the product's race
boundary.

Residual risk: the live native class-10 contrast could not be executed because
`ssh win` was offline, and the exact CI heads predate the recommended native
patch. The next Windows CI/head must record raw NTSTATUS and pass the ordinary
move, root-exchange, and no-replace collision tests before the product change
is accepted. This report hands off a diagnosis and minimal patch direction,
not evidence that the subsequent implementation has passed Windows CI.
