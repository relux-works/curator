# TASK-260729-1bf72u Windows cleanup-guard correction evidence

Rework cycle 1. Captured: 2026-07-29T12:26:50Z-2026-07-29T12:35Z UTC on `ssh win` (DESKTOP-3PBO632, Windows 10.0.19045.6456, Windows PowerShell 5.1.19041.6456).

Corrected report: `TASK-260729-1bf72u_runner-readiness.md`

| Revision | SHA-256 |
| --- | --- |
| 1 (reviewed, changes requested) | `3afa0fe0dd66ef79eec9dac00dad973a5e2e3fc8a8a2dfaf61179f1bbc1ec0ef` |
| 2 (this correction) | `d6e1ce11508bd88b8ae356839b071eba75a19e46f434ad53f11b3a47dcc239ad` |

## Defect

Revision 1 built `$RunTmp` from `$env:TEMP`, then the native matrix set `$env:TEMP = $RunTmp`, and the cleanup block recomputed the expected parent from that overwritten value. The guard therefore always refused before removing the exact run root. Cleanup also ran only on the success path.

## Correction

1. `$OriginalTemp`/`$OriginalTmp`/`$OriginalTempParent` are captured before any `TEMP`/`TMP` override.
2. `$RunTmp` is constructed from `$OriginalTempParent`.
3. Everything after `New-Item` is wrapped in `try`; `finally` restores `$env:TEMP`/`$env:TMP` and calls the exact guarded cleanup.
4. The guard validates `[IO.Path]::GetDirectoryName($ResolvedRunTmp)` against `$OriginalTempParent`, enforces the task name prefix, refuses reparse points, removes the exact root, then verifies absence.
5. The guarded cleanup exists once as `Invoke-TaskWindowsCleanup`; the recovery path calls the same definition instead of a second copy.

## Commands and real exits

| # | Command | Exit | Result |
| --- | --- | ---: | --- |
| 1 | `ssh -o BatchMode=yes -o ConnectTimeout=10 win hostname` | 0 | `DESKTOP-3PBO632` |
| 2 | cleanup-guard probe, first delivery (8,660-char encoded command) | 1 | expected-red: remote `cmd.exe` rejected the command line as too long |
| 3 | cleanup-guard probe, compacted (6,372-char encoded command), piped through `tee` | 0 | valid but pipe-delivered; not cited as gate evidence |
| 4 | cleanup-guard probe, standalone, no pipe | 0 | JSON below; authoritative evidence |
| 5 | remote absence/leftover verification | 0 | both probe roots absent, 0 task-prefixed leftovers |
| 6 | matrix block parse-check on Windows PowerShell 5.1 | 0 | 0 parse errors; probe file removed and verified absent |
| 7 | recovery block parse-check on Windows PowerShell 5.1 | 0 | 0 parse errors; probe file removed and verified absent |
| 8 | final remote leftover sweep after all probes | 0 | 0 task-prefixed leftovers |

## Probe 4 raw output (standalone run, exit 0)

```json
{"old_expected_parent":"C:\\Users\\admin\\AppData\\Local\\Temp\\TASK-260729-1bf72u-native-windows","old_actual_parent":"C:\\Users\\admin\\AppData\\Local\\Temp","old_guard_passes":false,"original_temp_parent":"C:\\Users\\admin\\AppData\\Local\\Temp","probe_run_tmp":"C:\\Users\\admin\\AppData\\Local\\Temp\\TASK-260729-1bf72u-cleanupguard-c07f8c35487542128b4afa419d039558","created":true,"created_child":true,"simulated_failure":"simulated gate failure inside post-creation matrix","finally_ran":true,"env_temp_at_cleanup":"C:\\Users\\admin\\AppData\\Local\\Temp\\TASK-260729-1bf72u-cleanupguard-c07f8c35487542128b4afa419d039558","cleanup_expected_parent":"C:\\Users\\admin\\AppData\\Local\\Temp","cleanup_actual_parent":"C:\\Users\\admin\\AppData\\Local\\Temp","cleanup_guard_passes":true,"reparse_point":false,"cleanup_verified_absent":true,"env_restored":true}
```

`old_guard_passes:false` reproduces the reviewer's finding. `cleanup_guard_passes:true` with `finally_ran:true` and a non-null `simulated_failure` proves the corrected guard admits the run root and still runs on a mid-matrix failure. `cleanup_verified_absent:true` and `env_restored:true` prove removal and environment restoration.

## Probe 5 / 8 raw output (exit 0 each)

```json
{"temp_root":"C:\Users\admin\AppData\Local\Temp","probe_01_absent":true,"probe_02_absent":true,"task_prefixed_leftovers":[],"leftover_count":0}
```

## Probe 6 / 7 raw output

```json
{"path":"C:/Users/admin/AppData/Local/Temp/TASK-260729-1bf72u-parse-0f7fa4778f2a48a896b231daef482394.ps1","parse_error_count":0,"parse_errors":[],"probe_removed_verified_absent":true}
{"path":"C:/Users/admin/AppData/Local/Temp/TASK-260729-1bf72u-parse-977eab93d61146cf9d802df07791ab19.ps1","parse_error_count":0,"parse_errors":[],"probe_removed_verified_absent":true}
```

## Cleanup-guard probe source (as executed)

```powershell
$ErrorActionPreference = 'Stop'
$r = [ordered]@{}
$OldRunTmp = Join-Path $env:TEMP 'TASK-260729-1bf72u-native-windows'
$OldExp = [IO.Path]::GetFullPath($OldRunTmp).TrimEnd('\')
$OldRes = [IO.Path]::GetFullPath($OldRunTmp)
$r['old_expected_parent'] = $OldExp
$r['old_actual_parent'] = [IO.Path]::GetDirectoryName($OldRes)
$r['old_guard_passes'] = ([IO.Path]::GetDirectoryName($OldRes) -eq $OldExp)
$OrigParent = [IO.Path]::GetFullPath($env:TEMP).TrimEnd('\')
$SavedTemp = $env:TEMP
$SavedTmp = $env:TMP
$r['original_temp_parent'] = $OrigParent
$RunTmp = Join-Path $OrigParent ('TASK-260729-1bf72u-cleanupguard-' + [guid]::NewGuid().ToString('N'))
$r['probe_run_tmp'] = $RunTmp
if (Test-Path -LiteralPath $RunTmp) { throw "exists: $RunTmp" }
New-Item -ItemType Directory -Path $RunTmp | Out-Null
Set-Content -LiteralPath (Join-Path $RunTmp 'payload.txt') -Value 'probe' -Encoding ASCII
$r['created'] = [bool](Test-Path -LiteralPath $RunTmp)
$r['created_child'] = [bool](Test-Path -LiteralPath (Join-Path $RunTmp 'payload.txt'))
$r['simulated_failure'] = $null
$r['finally_ran'] = $false
try {
  $env:TEMP = $RunTmp
  $env:TMP = $RunTmp
  throw 'simulated gate failure inside post-creation matrix'
} catch {
  $r['simulated_failure'] = $_.Exception.Message
} finally {
  $r['finally_ran'] = $true
  $r['env_temp_at_cleanup'] = $env:TEMP
  $Res = [IO.Path]::GetFullPath($RunTmp)
  $r['cleanup_expected_parent'] = $OrigParent
  $r['cleanup_actual_parent'] = [IO.Path]::GetDirectoryName($Res)
  $Ok = ([IO.Path]::GetDirectoryName($Res) -eq $OrigParent)
  $r['cleanup_guard_passes'] = $Ok
  if (-not $Ok) { throw "Refusing cleanup outside original TEMP parent: $Res" }
  if (-not [IO.Path]::GetFileName($Res).StartsWith('TASK-260729-1bf72u-')) { throw "Refusing cleanup outside task prefix: $Res" }
  $r['reparse_point'] = $false
  if (Test-Path -LiteralPath $Res) {
    $It = Get-Item -Force -LiteralPath $Res
    $r['reparse_point'] = (($It.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0)
    if ($r['reparse_point']) { throw "Refusing recursive cleanup of reparse point: $Res" }
    Remove-Item -LiteralPath $Res -Recurse -Force -ErrorAction Stop
  }
  $r['cleanup_verified_absent'] = -not (Test-Path -LiteralPath $Res)
  $env:TEMP = $SavedTemp
  $env:TMP = $SavedTmp
  $r['env_restored'] = (($env:TEMP -eq $SavedTemp) -and ($env:TMP -eq $SavedTmp))
}
$r | ConvertTo-Json -Compress
```

## Scope honesty

- No prerequisite was installed or downloaded. Windows Python, Git, Go, and PowerShell 7 remain absent and are still reported blocked.
- No product source, registry, PATH, service, or repository state was changed.
- The macOS suite was not rerun. Revision 1 macOS evidence (483 passed / 17 skipped at exit 0; strict mypy exit 0 over 55 source files) stands as recorded and is not re-asserted as fresh.
- Coverage, product lint, and new Go-parity tests remain unavailable/deferred and are not claimed green.
- Remote writes were limited to uniquely named ephemeral probe paths under the host temp root, each removed and verified absent; the final sweep found 0 residue.
