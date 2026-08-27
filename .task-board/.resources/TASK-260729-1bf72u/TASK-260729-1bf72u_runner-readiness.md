# TASK-260729-1bf72u runner readiness

Observed: 2026-07-29T12:05:59Z  
Revision 2 (Windows cleanup-guard correction only): 2026-07-29T12:26:50Z — see [Revision history](#revision-history)  
Repository: `/Users/iv/Developer/intranet/cocoaskills`  
Repository HEAD: `edce8816dda44bb121d661b7c4dea942558ce408` (`main`, clean, two commits behind `origin/main`)  
Scope: read-only macOS and `ssh win` qualification. No package, toolchain, registry, PATH, service, repository, source, or system configuration was changed.

## Outcome

| Surface | Outcome | Gate meaning |
| --- | --- | --- |
| macOS current Python development/test runner | **Ready with disk barrier** | Repository venv and current tests/typecheck are green. The data volume remains 98% used, with 26 GiB free after probe cleanup. |
| macOS future strict Go parity | **Blocked prerequisite** | Three native Go 1.25 candidates are discoverable, but none is the approved Go 1.25.12 patch with a recorded `curator-go-toolchain-v1` identity. Ambient `GOROOT` also contaminates the PATH-selected Homebrew command. |
| Windows SSH execution/transfer substrate | **Ready, security-qualified as elevated only** | Batch SSH, SFTP, SCP, SHA-256 verification, `cmd.exe`, Windows PowerShell 5.1, NTFS hardlinks/symlinks/junctions, case behavior, and child process launch work. The SSH token is elevated and has `SeCreateSymbolicLinkPrivilege`; this does not prove standard-user behavior. |
| Windows current Python CI | **Blocked prerequisite** | No runnable Python, Python launcher, Git, Git Bash, or PowerShell 7 is installed/on PATH. The WindowsApps Python aliases exit 49. |
| Windows future Go parity | **Blocked prerequisite** | No Go command, conventional Go tree, or approved absolute Go root exists. Operator installation/approval is required. |
| Linux `lev` | **Deferred, non-gating** | Explicitly owned by `TASK-260728-1skseh` / `TASK-260728-1e6811`; it was not inspected in this task. |
| New Go-parity tests, coverage, packaging, and product lint | **Deferred/non-gating for this read-only readiness task** | CocoaSkills HEAD has no Go module/source or Go-parity implementation. No source/test file was authorized. Coverage is not installed. Packaging writes `dist/`. No lint command is declared. |

The task itself is handoff-ready: the missing runner prerequisites are the required output, not an external blocker to completing this inventory.

## Repository and current CI contract

`pyproject.toml` requires Python 3.11+ and declares `pytest`, `mypy`, `build`, and `twine` in the development workflow. `.github/workflows/ci.yml` runs:

```text
python -m pip install --upgrade pip
python -m pip install -e ".[dev]"
CURATOR_CONFORMANCE_ROOT=<github-workspace>/protocol-spec/conformance/v1 python -m pytest -v
python -m mypy
python -m build
twine check dist/*
```

The test matrix is Ubuntu/macOS/Windows with Python 3.11, 3.12, 3.13, and 3.14. Linux typecheck/build are separate jobs. Windows CI configures:

```text
git config --global core.autocrlf false
git config --global core.symlinks false
```

Current HEAD contains no `go.mod` and no `*.go`. Go is a future native build-driver prerequisite, not a current CocoaSkills source language.

## macOS inventory

### Host, shell, paths, disk

| Check | Exit | Observed |
| --- | ---: | --- |
| `sw_vers` | 0 | macOS 26.5, build 25F71 |
| `uname -a` | 0 | Darwin 25.5.0, `RELEASE_ARM64_T6031` |
| `uname -m` | 0 | `arm64` |
| `sysctl -n machdep.cpu.brand_string` | 0 | Apple M3 Max |
| `id` | 0 | uid/euid 502 (`iv`), staff; user belongs to admin/developer groups |
| `printenv SHELL` | 0 | `/bin/zsh` |
| `zsh --version` | 0 | zsh 5.9 |
| `command -v bash` | 0 | `/opt/homebrew/bin/bash` |
| `bash --version` | 0 | GNU bash 5.3.9 |
| `/bin/bash --version` | 0 | Apple system GNU bash 3.2.57 |
| `/bin/sh -c '...; set -o'` | 0 | POSIX mode; `pipefail` is off by default |
| `zsh -dfc 'umask; ulimit -a'` | 0 | umask 022; max processes 10,666; open files 1,048,575; core size 0 |
| `launchctl limit` | 0 | soft maxfiles 256; soft maxproc 10,666 |
| `printenv TMPDIR` | 0 | `/var/folders/cz/jqbthtks55zbkpcdkfyk4bl80000gp/T/` |
| `getconf DARWIN_USER_TEMP_DIR` | 0 | same temp root |
| `df -h /System/Volumes/Data` before gates | 0 | 21 GiB free, 98% used |
| `df -h /System/Volumes/Data` after cleanup | 0 | 26 GiB free, 98% used |
| `df -h /private/tmp` | 0 | same APFS data volume |
| `mount` | 0 | APFS data volume is local, journaled, protected; system root is sealed/read-only |

Effective PATH:

```text
/Users/iv/.local/bin:/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/System/Cryptexes/App/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/local/bin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/bin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/appleinternal/bin:/pkg/env/global/bin:/Library/Apple/usr/bin:/usr/local/MacGPG2/bin:/Applications/VMware Fusion.app/Contents/Public:/usr/local/go/bin:/opt/homebrew/lib/node_modules/@openai/codex/node_modules/@openai/codex-darwin-arm64/vendor/aarch64-apple-darwin/codex-path:/Users/iv/.codex/tmp/arg0/codex-arg0uNiBfJ:/Users/iv/.local/bin:/Users/iv/.goenv/shims:/Users/iv/.goenv/bin:/Users/iv/.scripts/:/Users/iv/.mint/bin:/opt/homebrew/opt/ruby/bin:/Users/iv/.cargo/bin:/Applications/iTerm.app/Contents/Resources/utilities:/Users/iv/bin:/Users/iv/.cache/lm-studio/bin
```

### Python, Git, and current gates

| Check | Exit | Observed |
| --- | ---: | --- |
| `command -v python3` | 0 | `/opt/homebrew/bin/python3` |
| `python3 --version` | 0 | Python 3.14.4 |
| `/usr/bin/python3 --version` | 0 | Python 3.9.6; below project floor, do not use |
| Python platform/temp probe | 0 | executable `/opt/homebrew/opt/python@3.14/bin/python3.14`; CPython 3.14.4; macOS arm64; temp root matches above |
| `command -v py` | 1 | launcher absent; non-gating on macOS |
| `py --version` | 127 | expected-red: launcher absent |
| `.venv/bin/python --version` | 0 | Python 3.14.4 |
| `.venv/bin/python -m pytest --version` | 0 | pytest 9.0.3 |
| `.venv/bin/python -m mypy --version` | 0 | mypy 2.1.0 |
| `.venv/bin/python -m build --version` | 0 | build 1.5.0 |
| `.venv/bin/python -m twine --version` | 0 | twine 6.2.0 |
| `git --version` | 0 | Apple Git 2.50.1 |
| `command -v git` | 0 | `/usr/bin/git` |
| Full pytest gate below | **0** | **483 passed, 17 skipped in 93.39 s** |
| Strict mypy gate below | **0** | **Success: no issues found in 55 source files** |
| `.venv/bin/python -m coverage --version` | 1 | expected-red: coverage module absent; no coverage claim |
| `git diff --check` | 0 | whitespace clean |
| `git diff --cached --quiet` | 0 | nothing staged |
| `git status --short` | 0 | empty; source tree unchanged |

Executed pytest gate:

```text
env TMPDIR=/private/tmp/TASK-260729-1bf72u-pytest \
  PYTHONDONTWRITEBYTECODE=1 \
  .venv/bin/python -m pytest -v -p no:cacheprovider
```

Exit 0. The 17 skips include all shared protocol cases because this inventory did not receive `CURATOR_CONFORMANCE_ROOT`, plus the native Windows PowerShell test. This is green current portable coverage, not Go-parity/conformance qualification.

Executed strict typing gate:

```text
env TMPDIR=/private/tmp/TASK-260729-1bf72u-pytest \
  PYTHONDONTWRITEBYTECODE=1 \
  .venv/bin/python -m mypy \
  --cache-dir /private/tmp/TASK-260729-1bf72u-pytest/mypy-cache
```

Exit 0. The task temp root was removed using an exact asserted path and `test ! -e /private/tmp/TASK-260729-1bf72u-pytest` exited 0 afterward.

The ambient Homebrew Python lacked all four dev modules (`pytest`, `mypy`, `build`, `twine`): each `python3 -m <module> --version` exited 1. The repository venv is the qualified interpreter.

### macOS Go discovery and admission

| Candidate/check | Exit | Evidence | Assessment |
| --- | ---: | --- | --- |
| `command -v go` | 0 | `/opt/homebrew/bin/go` | Discovery only |
| `type -a go` | 0 | Homebrew, `/usr/local/go/bin/go`, goenv shim | Multiple ambient candidates |
| ambient `go version` | 0 | `go1.25.5 darwin/arm64` | Accepted family, stale patch |
| ambient `go env -json ...` | 0 | `GOROOT=/Users/iv/.goenv/versions/1.25.5`, `GOTOOLCHAIN=auto` | **Rejected for strict admission:** PATH launcher is Homebrew while ambient root is goenv; auto switching is enabled |
| clean Homebrew absolute probes | 0 each | `go1.25.5`; root `/opt/homebrew/Cellar/go/1.25.5/libexec`; `GOENV` empty; `GOTOOLCHAIN=local` | Native candidate only; user-owned and not approved/full-tree fingerprinted |
| clean goenv absolute probes | 0 each | `go1.25.5`; root `/Users/iv/.goenv/versions/1.25.5` | Native candidate only; user-owned and not approved/full-tree fingerprinted |
| clean official-package absolute probes | 0 each | `go1.25.1`; root `/usr/local/go` | Native official-package candidate, root-owned, but stale and not approved/full-tree fingerprinted |
| Homebrew `bin/go` SHA-256 | 0 | `11caa81dd6f1c2d6e2799f19265db880dd5a2c91b45703b4e871dbd90249f261` | Individual file only |
| goenv `bin/go` SHA-256 | 0 | `a84ad6c885e8a875475b9c6cea4f2e213f8ab449d43193520811c39b6007485c` | Individual file only |
| `/usr/local/go/bin/go` SHA-256 | 0 | `c5f484a7e96e11b7c8c3361fa0c3d1c7480d65bf6fc92a03516719f2250d2be8` | Individual file only |
| Homebrew `codesign -dv --verbose=4` | 0 | ad-hoc/linker-signed, no team ID | Not equivalent to official Developer ID provenance |
| goenv and `/usr/local/go` codesign probes | 0 each | Google LLC Developer ID, team `EQHXZ8M8AV`, hardened runtime | Signature is useful provenance, not full-tree identity |

Strict Go parity remains stopped until an operator selects a supported official Go 1.25.12 root and the manager computes/records `curator-go-toolchain-v1`. PATH, `go version`, binary SHA-256, package receipt, and code signature are insufficient substitutes.

### macOS filesystem/process/security

The corrected task-scoped temp probe exited 0:

```json
{"case_alias_exists":true,"case_sensitive":false,"child_exit":0,"cleanup_verified":true,"hardlink":{"link_count":2,"same_inode":true,"samefile":true},"mode":"0o700","symlink":{"is_symlink":true,"readlink":"Target.txt","reads_target":true,"samefile":true}}
```

The first version of this probe also exited 0 but used a brittle lexical `Path.resolve() == target` comparison and reported `resolves:false`; it was not used as capability evidence. The corrected probe used content read and `os.path.samefile`.

Security/signing checks:

| Check | Exit | Observed |
| --- | ---: | --- |
| `spctl --status` | 0 | assessments enabled |
| `csrutil status` | 0 | SIP enabled |
| `xcode-select -p` | 0 | Xcode 26.5 selected |
| `xcrun --find codesign` | 0 | `/usr/bin/codesign` |
| `security find-identity -v -p codesigning` | 0 | 9 valid signing identities; identity details intentionally omitted here |
| `/usr/bin/true` task child | 0 | process creation/wait works |

## Windows inventory through `ssh win`

### Host, shell, paths, disk

| Check | Exit | Observed |
| --- | ---: | --- |
| `ssh ... win hostname` | 0 | `DESKTOP-3PBO632` |
| `ssh ... win ver` | 0 | Microsoft Windows 10.0.19045.6456 |
| CIM OS inventory | 0 | Windows 10 Pro, build 19045, 64-bit |
| `set PROCESSOR_ARCHITECTURE` | 0 | AMD64 |
| computer system | 0 | Apple Inc. MacBookPro13,2 / Boot Camp |
| `whoami` | 0 | `desktop-3pbo632\admin` |
| administrator token check | 0 | `True` |
| PowerShell version | 0 | Windows PowerShell 5.1.19041.6456, Desktop edition |
| `where pwsh` | 1 | PowerShell 7 absent |
| `pwsh ...` | 1 | expected-red: command absent |
| `%COMSPEC%` | 0 | `C:\Windows\system32\cmd.exe` |
| OpenSSH `DefaultShell` registry value | 0 | absent; default SSH command shell is `cmd.exe` |
| `chcp` | 0 | code page 866; localized native diagnostics are not UTF-8 |
| `[Environment]::NewLine.Length` | 0 | 2 (CRLF) |
| fixed disk CIM inventory | 0 | `C:` NTFS, BOOTCAMP; 125,684,412,416 bytes total; 84,509,929,472 bytes free |

Remote PATH:

```text
C:\Program Files (x86)\Intel\iCLS Client\;C:\Program Files\Intel\iCLS Client\;C:\Windows\system32;C:\Windows;C:\Windows\System32\Wbem;C:\Windows\System32\WindowsPowerShell\v1.0\;C:\Windows\System32\OpenSSH\;C:\Program Files\Tailscale\;C:\Program Files (x86)\Windows Kits\10\Windows Performance Toolkit\;C:\Users\admin\AppData\Local\Programs\OpenAI\Codex\bin;C:\Users\admin\AppData\Local\Microsoft\WindowsApps;
```

`TEMP`, `TMP`, and `[IO.Path]::GetTempPath()` resolve to:

```text
C:\Users\admin\AppData\Local\Temp
```

### Missing Windows prerequisites

| Check | Exit | Finding |
| --- | ---: | --- |
| `where python` / `where python3` | 0 | only WindowsApps aliases |
| `python --version` / `python3 --version` | **49 each** | expected-red: aliases are not runnable Python |
| Python execution probe | **49** | no interpreter |
| `where py`, `py --version`, `py -0p` | **1 each** | launcher absent |
| Python HKCU/HKLM registry roots and `%LOCALAPPDATA%\Programs\Python` | 0 | all `False` |
| `where git`, `git --version` | **1 each** | Git absent |
| conventional 32/64-bit Git and Git Bash paths | 0 | all `False` |
| `where bash`, `bash --version` | **1 each** | Git Bash absent |
| `where go`, `go version` | **1 each** | Go absent |
| PowerShell `Get-Command go.exe -All` | 0 | empty |
| conventional Go root inventory | 0 | all roots and `bin\go.exe` candidates `False` |

Windows cannot run the current CocoaSkills pytest suite until an operator provisions Python 3.11-3.14 and Git. It cannot run future Go parity until an approved Go 1.25.12 root passes admission.

PowerShell 7 and Git Bash are useful but not separately gating: current project behavior explicitly supports Windows PowerShell 5.1 and `cmd.exe`. Git itself is gating because tests create real temporary repositories.

### Windows filesystem/process/security qualification

The unique probe directory was:

```text
C:\Users\admin\AppData\Local\Temp\TASK-260729-1bf72u-windows-fs-f53babe895a54b01ab13917197a508d1
```

The encoded PowerShell probe exited 0 and returned:

```json
{"created":true,"hardlink":true,"symlink":true,"junction":true,"case_alias_exists":true,"case_sensitive":false,"child_exit":0,"case_query_exit":0,"root_sddl":"O:BAG:S-1-5-21-2566270768-74292453-1896159616-513D:(A;OICIID;FA;;;SY)(A;OICIID;FA;;;BA)(A;OICIID;FA;;;S-1-5-21-2566270768-74292453-1896159616-1001)","fatal":null,"cleanup_verified":true,"symlink_attributes":"Archive, ReparsePoint","junction_attributes":"Directory, ReparsePoint"}
```

Interpretation:

- NTFS is case-insensitive for this directory.
- File hardlinks, file symlinks, directory junctions, reparse-point metadata, and child process launch/wait are available.
- The session is elevated. `whoami /priv` exited 0 and reports `SeCreateSymbolicLinkPrivilege` enabled, plus high-integrity administrative privileges. The Developer Mode registry probe returned no configured value. Do not generalize this result to a non-elevated runner.
- `Get-ExecutionPolicy -List` exited 0 with all scopes `Undefined`.
- `Get-AuthenticodeSignature $PSHOME\powershell.exe` exited 0 with status `Valid`.
- `Get-Command Start-Process,Stop-Process,Wait-Process` exited 0.
- `fsutil behavior query SymlinkEvaluation` exited 0; local-to-local and local-to-remote evaluation are enabled, remote-origin evaluation is disabled.

The probe directory was removed in `finally` and `cleanup_verified:true` proves absence.

### SSH execution and transfer

| Check | Exit | Observed |
| --- | ---: | --- |
| local `ssh -V` | 0 | OpenSSH 10.2p1, LibreSSL 3.3.6 |
| `ssh -G win` | 0 | user `admin`, port 22, `IdentitiesOnly yes`, public-key enabled, no agent forwarding, no control master |
| remote `ssh -V` | 0 | OpenSSH_for_Windows 9.5p1, LibreSSL 3.8.2 |
| `sftp -b /dev/null ... win` | 0 | SFTP subsystem handshake works |
| batch `ssh ... hostname` | 0 | non-interactive execution works |
| `scp -p ... win:C:/.../probe.txt` | 0 | SCP/SFTP transfer works |
| local `shasum -a 256` | 0 | `5969d12301e94ad7043f580ddaa7a8cc82f317ece9c41eddc35371616782a465` |
| remote `certutil -hashfile ... SHA256` | 0 | identical SHA-256 |
| remote exact `Remove-Item` | 0 | transfer probe removed |
| remote `Test-Path` | 0 | `False`, verified absent |
| local exact `unlink` and absence test | 0 each | local transfer marker removed and verified absent |

Every SSH connection warned that the negotiated exchange was not post-quantum. This is a transport hardening finding, not a functional runner blocker. Batch commands in this report explicitly use `-o BatchMode=yes -o ConnectTimeout=10`; the base `ssh -G` configuration itself says `batchmode no`.

## Official Windows Go 1.25.12 operator setup

Official facts were rechecked on 2026-07-29:

- Go downloads JSON: <https://go.dev/dl/?mode=json>
- Go installation guide: <https://go.dev/doc/install>
- Go toolchain selection: <https://go.dev/doc/toolchain>
- Go release policy/history: <https://go.dev/doc/devel/release>

The current accepted manager allowlist is family 1.25. Official Go currently publishes 1.25.12 and 1.26.5. Do not substitute 1.26 until the manager allowlist and conformance are deliberately expanded. Standard `GOTOOLCHAIN=auto` can find another toolchain on PATH or download one; parity probes must use `GOENV=off` and `GOTOOLCHAIN=local`.

### Option A: official amd64 MSI, recommended

Operator-only artifact:

```text
https://go.dev/dl/go1.25.12.windows-amd64.msi
SHA-256 45bc4ffd130e778374818551790abc2b4378dc5e89e46fcd114627ec9ebc1687
```

Manually download it from `go.dev`, verify it before opening, then follow the official interactive prompts. The normal x64 root is `C:\Program Files\Go`. Close and reopen shells after installation. This task did not download or execute it.

Pre-open verification:

```powershell
$Installer = "$HOME\Downloads\go1.25.12.windows-amd64.msi"
$Expected = '45BC4FFD130E778374818551790ABC2B4378DC5E89E46FCD114627EC9EBC1687'
$Actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $Installer).Hash
if ($Actual -ne $Expected) { throw "Go MSI SHA-256 mismatch: $Actual" }
Get-AuthenticodeSignature -LiteralPath $Installer |
  Format-List Status,StatusMessage,SignerCertificate
```

### Option B: official amd64 ZIP for an operator-owned immutable root

Operator-only artifact:

```text
https://go.dev/dl/go1.25.12.windows-amd64.zip
SHA-256 d5dc82da351b00e5eedd04f41356817d674cc4308131f0f638a5b14c5c3af4cb
```

Verify the ZIP, expand it into a new empty staging directory, then publish a real administrator-controlled directory such as `C:\Curator\Toolchains\go1.25.12`. Never overlay an existing Go tree. Do not use a repository, project runtime, junction, Chocolatey, Scoop, WindowsApps alias, or ambient PATH candidate as admission evidence.

Pre-extract verification:

```powershell
$Archive = "$HOME\Downloads\go1.25.12.windows-amd64.zip"
$Expected = 'D5DC82DA351B00E5EEDD04F41356817D674CC4308131F0F638A5B14C5C3AF4CB'
$Actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $Archive).Hash
if ($Actual -ne $Expected) { throw "Go ZIP SHA-256 mismatch: $Actual" }
```

### Exact post-install clean verification and fingerprints

Run from outside repositories and package/runtime roots:

```powershell
$ErrorActionPreference = 'Stop'
$ApprovedGoRoot = 'C:\Program Files\Go' # or the explicitly approved real ZIP root
$ApprovedGo = Join-Path $ApprovedGoRoot 'bin\go.exe'

$RootItem = Get-Item -Force -LiteralPath $ApprovedGoRoot
$GoItem = Get-Item -Force -LiteralPath $ApprovedGo
if (-not $RootItem.PSIsContainer) { throw 'Approved GOROOT is not a directory' }
if ($GoItem.PSIsContainer) { throw 'Approved go.exe is not a regular file' }
if (($RootItem.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
  throw 'Approved GOROOT must not be a reparse point'
}

Get-Acl -LiteralPath $ApprovedGoRoot | Format-List Owner,Sddl
Get-FileHash -Algorithm SHA256 -LiteralPath $ApprovedGo
Get-AuthenticodeSignature -LiteralPath $ApprovedGo |
  Format-List Status,StatusMessage,SignerCertificate

Remove-Item Env:\GOROOT -ErrorAction SilentlyContinue
$env:GOENV = 'off'
$env:GOTOOLCHAIN = 'local'
$env:LC_ALL = 'C'
$env:LANG = 'C'

& $ApprovedGo version
& $ApprovedGo env -json `
  GOROOT GOHOSTOS GOHOSTARCH GOOS GOARCH GOVERSION GOTOOLCHAIN GOENV `
  GOAMD64 GOTELEMETRY GOTELEMETRYDIR
if ($LASTEXITCODE -ne 0) { throw "Go clean probe failed: $LASTEXITCODE" }
```

Require:

```text
go version go1.25.12 windows/amd64
GOROOT == the approved real root
GOHOSTOS/GOOS == windows
GOHOSTARCH/GOARCH == amd64
GOTOOLCHAIN == local
GOENV disabled/empty under the clean probe
```

The official artifact SHA-256 and `go.exe` SHA-256 are two distinct fingerprints. Neither is the required full installed-tree identity.

There is deliberately **no fabricated full-tree command** in current CocoaSkills HEAD. `TASK-260720-3j8pp5` owns the independent Python implementation of the accepted byte-exact `curator-go-toolchain-v1` walk/framing algorithm and its public admission surface. Until that lands, the operator must record the checks above and keep native Go parity blocked. After it lands, the exact acceptance command must:

1. select the approved absolute root/executable without PATH;
2. run `telemetry off`, `version`, and the fixed `env -json` probe from an empty private environment;
3. compute `curator-go-toolchain-v1` over every directory, regular file, and allowed internal relative link plus normalized `go version`;
4. emit and persist `sha256:<64 lowercase hex>`;
5. recompute it after the last child exits and fail closed on drift.

Providing a made-up PowerShell aggregate hash here would be a forced fit and would not satisfy protocol identity.

## Native validation matrix

### Common barriers before either native run

1. Use a clean, immutable repository revision and an explicit immutable `CURATOR_CONFORMANCE_ROOT`.
2. Require manifest SHA-256:

   ```text
   b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
   ```

   This is the digest named by `TASK-260720-3s27te`. Do not silently substitute the older committed/default suite.
3. Require at least 20 GiB free for a full native parity run. The threshold is an operational guard derived from same-day 98% macOS disk pressure and earlier task evidence of build-cache failures. A focused test may use 10 GiB only with an explicit smaller selector and no full Go tree/cache work.
4. Allocate a unique task temp root outside the repository and every worktree.
5. Record only task-owned child PIDs. Do not use broad `pkill go`, `killall`, `taskkill /IM`, or cleanup globs.
6. Verify approved Go identity before and after the last child.
7. Stop before tests if Python, Git, approved Go, conformance digest, process barrier, temp barrier, or disk barrier fails.

### macOS exact commands

```bash
set -eu
set -o pipefail
REPO=/Users/iv/Developer/intranet/cocoaskills
CONFORMANCE_ROOT=/absolute/immutable/curator-spec/conformance/v1
EXPECTED_MANIFEST=b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c
RUN_TMP=/private/tmp/TASK-260729-1bf72u-native-macos

test ! -e "$RUN_TMP"
mkdir -m 700 "$RUN_TMP"
test -x "$REPO/.venv/bin/python"
test -x /absolute/approved/go1.25.12/bin/go
test -f "$CONFORMANCE_ROOT/manifest.json"
test "$(shasum -a 256 "$CONFORMANCE_ROOT/manifest.json" | awk '{print $1}')" = "$EXPECTED_MANIFEST"
python3 -c 'import shutil; assert shutil.disk_usage("/System/Volumes/Data").free >= 20 * 1024**3'

cd "$REPO"
test -z "$(git status --short)"
GOENV=off GOTOOLCHAIN=local /absolute/approved/go1.25.12/bin/go version
GOENV=off GOTOOLCHAIN=local /absolute/approved/go1.25.12/bin/go env -json \
  GOROOT GOHOSTOS GOHOSTARCH GOOS GOARCH GOVERSION GOTOOLCHAIN GOENV

CURATOR_CONFORMANCE_ROOT="$CONFORMANCE_ROOT" \
TMPDIR="$RUN_TMP" PYTHONDONTWRITEBYTECODE=1 \
.venv/bin/python -m pytest -v -p no:cacheprovider

TMPDIR="$RUN_TMP" PYTHONDONTWRITEBYTECODE=1 \
.venv/bin/python -m mypy --cache-dir "$RUN_TMP/mypy-cache"

git diff --check
git status --short
```

The single `awk` pipeline above must run under a shell with `pipefail` enabled if used as a formal gate. A pipeline-free alternative is:

```bash
python3 - "$CONFORMANCE_ROOT/manifest.json" "$EXPECTED_MANIFEST" <<'PY'
import hashlib
import pathlib
import sys
actual = hashlib.sha256(pathlib.Path(sys.argv[1]).read_bytes()).hexdigest()
if actual != sys.argv[2]:
    raise SystemExit(f"manifest SHA-256 mismatch: {actual}")
PY
```

Packaging is a separate clean-worktree gate because it writes `dist/`:

```bash
.venv/bin/python -m build
.venv/bin/python -m twine check dist/*
```

It was not executed by this read-only inventory.

### Windows exact commands after operator provisioning

Use Windows PowerShell 5.1, which is already present.

Two invariants make this block safe:

1. The host temp parent is captured **before** anything overrides `TEMP`/`TMP`. `$RunTmp` is built from that immutable value, and every cleanup guard compares against it — never against the live `$env:TEMP`.
2. Everything after the run root is created lives inside `try`, so the exact guarded cleanup in `finally` runs on a prerequisite failure, a gate failure, or a test failure — not only on success.

```powershell
$ErrorActionPreference = 'Stop'
$Repo = 'C:\runner\cocoaskills' # operator-materialized clean revision
$ConformanceRoot = 'C:\runner\protocol-spec\conformance\v1'
$ExpectedManifest = 'B6F56AACC0E37DCC6692F73F641BFF761E89B645ADFE20A47A06D81C6FDA204C'
$ApprovedGoRoot = 'C:\Program Files\Go'
$ApprovedGo = Join-Path $ApprovedGoRoot 'bin\go.exe'

# Capture the original host temp parent BEFORE the matrix overrides TEMP/TMP.
# $OriginalTempParent is the only value the cleanup guard is allowed to trust.
$OriginalTemp = $env:TEMP
$OriginalTmp = $env:TMP
$OriginalTempParent = [IO.Path]::GetFullPath($OriginalTemp).TrimEnd('\')
$RunTmp = Join-Path $OriginalTempParent 'TASK-260729-1bf72u-native-windows'
$RecordedPids = @() # populated by the harness with task-owned child PIDs only

function Invoke-TaskWindowsCleanup {
  [CmdletBinding()]
  param(
    [Parameter(Mandatory)][string]$RunTmp,
    [Parameter(Mandatory)][string]$OriginalTempParent,
    [int[]]$RecordedPids = @()
  )
  # 1. Stop only PIDs recorded by the task harness. Never match by process name.
  foreach ($TaskPid in $RecordedPids) {
    if (Get-Process -Id $TaskPid -ErrorAction SilentlyContinue) {
      Stop-Process -Id $TaskPid -ErrorAction Stop
    }
  }
  foreach ($TaskPid in $RecordedPids) {
    if (Get-Process -Id $TaskPid -ErrorAction SilentlyContinue) {
      throw "Recorded process still alive: $TaskPid"
    }
  }

  # 2. Validate the run root against the ORIGINAL host temp parent, not $env:TEMP.
  $ExpectedParent = [IO.Path]::GetFullPath($OriginalTempParent).TrimEnd('\')
  $ResolvedRunTmp = [IO.Path]::GetFullPath($RunTmp)
  if ([IO.Path]::GetDirectoryName($ResolvedRunTmp) -ne $ExpectedParent) {
    throw "Refusing cleanup outside the original TEMP parent: $ResolvedRunTmp"
  }
  if (-not [IO.Path]::GetFileName($ResolvedRunTmp).StartsWith('TASK-260729-1bf72u-')) {
    throw "Refusing cleanup outside task prefix: $ResolvedRunTmp"
  }

  # 3. Remove the exact run root, refusing reparse points.
  if (Test-Path -LiteralPath $ResolvedRunTmp) {
    $Item = Get-Item -Force -LiteralPath $ResolvedRunTmp
    if (($Item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) {
      throw "Refusing recursive cleanup of reparse point: $ResolvedRunTmp"
    }
    Remove-Item -LiteralPath $ResolvedRunTmp -Recurse -Force -ErrorAction Stop
  }

  # 4. Verify absence.
  if (Test-Path -LiteralPath $ResolvedRunTmp) {
    throw "Cleanup verification failed: $ResolvedRunTmp"
  }
}

if (Test-Path -LiteralPath $RunTmp) { throw "Temp root already exists: $RunTmp" }
New-Item -ItemType Directory -Path $RunTmp | Out-Null

try {
  if (-not (Test-Path -LiteralPath (Join-Path $Repo '.venv\Scripts\python.exe'))) {
    throw 'Qualified repository Python venv missing'
  }
  if (-not (Test-Path -LiteralPath $ApprovedGo)) { throw 'Approved Go missing' }
  if (-not (Get-Command git.exe -ErrorAction SilentlyContinue)) { throw 'Git missing' }
  $Manifest = Join-Path $ConformanceRoot 'manifest.json'
  if ((Get-FileHash -Algorithm SHA256 -LiteralPath $Manifest).Hash -ne $ExpectedManifest) {
    throw 'Conformance manifest SHA-256 mismatch'
  }
  if ((Get-PSDrive -Name C).Free -lt 20GB) { throw 'Less than 20 GiB free' }

  Set-Location -LiteralPath $Repo
  if (git status --short) { throw 'Repository is not clean' }

  Remove-Item Env:\GOROOT -ErrorAction SilentlyContinue
  $env:GOENV = 'off'
  $env:GOTOOLCHAIN = 'local'
  & $ApprovedGo version
  if ($LASTEXITCODE -ne 0) { throw "go version failed: $LASTEXITCODE" }
  & $ApprovedGo env -json GOROOT GOHOSTOS GOHOSTARCH GOOS GOARCH GOVERSION GOTOOLCHAIN GOENV
  if ($LASTEXITCODE -ne 0) { throw "go env failed: $LASTEXITCODE" }

  $env:CURATOR_CONFORMANCE_ROOT = $ConformanceRoot
  $env:TEMP = $RunTmp
  $env:TMP = $RunTmp
  $env:PYTHONDONTWRITEBYTECODE = '1'
  & (Join-Path $Repo '.venv\Scripts\python.exe') -m pytest -v -p no:cacheprovider
  if ($LASTEXITCODE -ne 0) { throw "pytest failed: $LASTEXITCODE" }

  git diff --check
  if ($LASTEXITCODE -ne 0) { throw "git diff --check failed: $LASTEXITCODE" }
  if (git status --short) { throw 'Repository changed during validation' }
}
finally {
  # Restore the original variables first, then run the exact guarded cleanup.
  $env:TEMP = $OriginalTemp
  $env:TMP = $OriginalTmp
  Invoke-TaskWindowsCleanup -RunTmp $RunTmp `
    -OriginalTempParent $OriginalTempParent `
    -RecordedPids $RecordedPids
}
```

The harness must also persist `$OriginalTempParent` and `$RunTmp` durably at run start, so a run aborted with its shell can still be cleaned up against the recorded original parent rather than a live `$env:TEMP`.

The current Windows host cannot enter this matrix: the qualified venv, Python, Git, repository, conformance root, and approved Go are absent.

### Windows cleanup-guard verification

The previous revision of this report built `$RunTmp` from `$env:TEMP`, then set `$env:TEMP = $RunTmp` inside the matrix, and the cleanup block recomputed the expected parent from that overwritten value. Independent review reproduced the defect, and this revision re-measured both the defect and the correction on `ssh win`.

Probe: `powershell -NoProfile -NonInteractive -EncodedCommand <base64 UTF-16LE>`, run standalone (no pipe), exit **0**. It reproduces the old guard read-only, then exercises the corrected pattern end-to-end on a uniquely named ephemeral root, deliberately throwing inside the `try` to prove `finally` still cleans up:

```json
{"old_expected_parent":"C:\\Users\\admin\\AppData\\Local\\Temp\\TASK-260729-1bf72u-native-windows",
 "old_actual_parent":"C:\\Users\\admin\\AppData\\Local\\Temp",
 "old_guard_passes":false,
 "original_temp_parent":"C:\\Users\\admin\\AppData\\Local\\Temp",
 "probe_run_tmp":"C:\\Users\\admin\\AppData\\Local\\Temp\\TASK-260729-1bf72u-cleanupguard-c07f8c35487542128b4afa419d039558",
 "created":true,"created_child":true,
 "simulated_failure":"simulated gate failure inside post-creation matrix",
 "finally_ran":true,
 "env_temp_at_cleanup":"C:\\Users\\admin\\AppData\\Local\\Temp\\TASK-260729-1bf72u-cleanupguard-c07f8c35487542128b4afa419d039558",
 "cleanup_expected_parent":"C:\\Users\\admin\\AppData\\Local\\Temp",
 "cleanup_actual_parent":"C:\\Users\\admin\\AppData\\Local\\Temp",
 "cleanup_guard_passes":true,
 "reparse_point":false,
 "cleanup_verified_absent":true,
 "env_restored":true}
```

Reading:

- `old_guard_passes:false` confirms the previous published cleanup would always refuse before removing the run root.
- `cleanup_guard_passes:true` with `cleanup_expected_parent == cleanup_actual_parent == C:\Users\admin\AppData\Local\Temp` confirms the corrected guard admits the exact run root while still comparing against the original immutable parent.
- `finally_ran:true` alongside a non-null `simulated_failure` confirms the guarded cleanup runs on a mid-matrix failure.
- `cleanup_verified_absent:true` and `env_restored:true` confirm removal and environment restoration.

An earlier identical run of the same probe (root `...-cleanupguard-1446ed1b7bb24644b03315f3f4969740`) returned the same verdicts. A separate read-only PowerShell verification exited **0** and returned `{"temp_root":"C:\\Users\\admin\\AppData\\Local\\Temp","probe_01_absent":true,"probe_02_absent":true,"task_prefixed_leftovers":[],"leftover_count":0}` — both ephemeral roots absent and no `TASK-260729-1bf72u-*` residue in the remote temp root.

Both published PowerShell blocks were additionally parse-checked, not executed, against the real Windows PowerShell 5.1 parser. Each was transferred to a uniquely named ephemeral remote path, parsed with `[System.Management.Automation.Language.Parser]::ParseFile`, then removed and verified absent in the same process:

| Block | Exit | Result |
| --- | ---: | --- |
| native Windows validation matrix | 0 | `{"parse_error_count":0,"parse_errors":[],"probe_removed_verified_absent":true}` |
| cleanup recovery invocation | 0 | `{"parse_error_count":0,"parse_errors":[],"probe_removed_verified_absent":true}` |

Parse success proves syntactic validity on the target shell only. It does not execute the matrix and does not qualify the still-missing Windows prerequisites.

No prerequisite was installed, and no system, registry, PATH, service, or repository state was changed by this correction.

### Stop and cleanup barriers

macOS:

```bash
# Stop: interrupt the foreground gate once, then wait for task-owned children.
# Escalate only recorded PIDs; never match by process name.
kill -TERM <recorded-task-pid>
# After a bounded wait, only if that exact PID is still the recorded process:
kill -KILL <recorded-task-pid>

python3 - "$RUN_TMP" <<'PY'
import os
import pathlib
import shutil
import sys
p = pathlib.Path(sys.argv[1])
if p.parent != pathlib.Path("/private/tmp") or not p.name.startswith("TASK-260729-1bf72u-"):
    raise SystemExit(f"refusing cleanup outside task temp: {p}")
if p.is_symlink():
    raise SystemExit(f"refusing recursive cleanup of symlink: {p}")
if p.exists():
    shutil.rmtree(p)
if p.exists() or p.is_symlink():
    raise SystemExit(f"cleanup verification failed: {p}")
PY
```

Windows stop and cleanup is exactly one thing: `Invoke-TaskWindowsCleanup`, defined once in the native matrix above. It stops only recorded PIDs, validates the run root against the original immutable temp parent, refuses reparse points, removes the exact root, and verifies absence. There is deliberately no second copy of that body here, so the in-run path and the recovery path cannot drift apart.

In-run: the matrix `finally` always invokes it, including on a prerequisite, gate, or test failure, after restoring `$env:TEMP`/`$env:TMP`.

Recovery for a run whose shell is gone — re-declare the function verbatim from the matrix, then call it with the values the harness recorded at run start:

```powershell
# $RecordedOriginalTempParent and $RecordedRunTmp come from the run's persisted record.
# Never recompute the expected parent from the live $env:TEMP: the matrix overrides
# TEMP/TMP to $RunTmp, so a live read makes the guard refuse the exact root it owns.
Invoke-TaskWindowsCleanup -RunTmp $RecordedRunTmp `
  -OriginalTempParent $RecordedOriginalTempParent `
  -RecordedPids $RecordedPids

if (Test-Path -LiteralPath $RecordedRunTmp) {
  throw "Cleanup verification failed: $RecordedRunTmp"
}
```

After either native run:

1. approved Go identity must match its pre-run value;
2. every recorded PID must be absent;
3. the exact task temp root must be absent;
4. repository `git status --short` must remain empty;
5. free disk must remain above the applicable floor.

## Probe anomalies and expected-red record

These failures are reported as failures, not passes:

| Command/probe | Exit | Rationale/correction |
| --- | ---: | --- |
| initial required board `set_status(...development)` | 1 | estimate-required refusal; Fibonacci 5 estimate was set, then the exact status command exited 0 |
| initial `ssh win "cmd /d /c ver"` and unquoted `cmd /d /c ver` | 1 each | remote default shell quoting produced a malformed `ver"` token; direct `ssh win ver` exited 0 |
| first local PowerShell-encoding helper | 1 | local zsh quoting parse error; here-doc encoder exited 0 |
| two PowerShell `Test-Path` checks containing `Program Files` | 1 each | remote `cmd.exe` consumed quoting; encoded conventional-root inventory exited 0 and found all absent |
| first remote `certutil` path with unquoted backslashes | 2 | local shell removed backslashes; forward-slash path exited 0 and matched SHA-256 |
| `tasklist /FI "PID eq 4"` | 1 | remote quoting changed filter syntax; discarded, not used as process evidence |
| ambient Python module checks | 1 each | global Python lacks pytest/mypy/build/twine; repository venv equivalents exited 0 |
| coverage readiness | 1 | coverage module absent; no coverage claim |
| Windows Python aliases | 49 | aliases are not runnable interpreter installations |
| Windows `py`, Git, Bash, Go, and `pwsh` executions | 1 | prerequisites absent |
| local `rm -f` marker cleanup | no process exit | execution guard rejected launch; exact `unlink` exited 0 and absence test exited 0 |
| revision 1 published Windows cleanup guard | n/a (documentation defect) | the guard recomputed the expected temp parent from the already-overwritten `$env:TEMP`, so `old_guard_passes` measured `false` on `ssh win`; it would always refuse before removing the run root. Corrected in revision 2 and re-measured `true` |
| first cleanup-guard probe delivery via `ssh win powershell -EncodedCommand` | 1 | remote `cmd.exe` rejected an 8,660-character command line as too long; the compacted 6,372-character encoding exited 0 |
| first cleanup-guard probe run | 0 | valid result, but the invocation was piped through `tee`; re-run standalone at exit 0 and only the standalone run is cited as gate evidence |

No failed probe was reclassified as green. Corrected commands and their real exits are recorded separately above.

## Revision history

| Revision | When | Change |
| --- | --- | --- |
| 1 | 2026-07-29T12:05:59Z | Initial macOS and `ssh win` readiness inventory |
| 2 | 2026-07-29T12:26:50Z | Windows cleanup-guard correction only: original temp parent captured before the `TEMP`/`TMP` override and used to construct `$RunTmp`; post-creation matrix wrapped in `try/finally`; guarded cleanup consolidated into a single `Invoke-TaskWindowsCleanup` definition validated against the immutable parent; correction re-measured on `ssh win`, both published blocks parse-checked at 0 errors, and every ephemeral probe path verified absent |

Revision 2 changed no other finding. No product source was edited, no prerequisite was installed, and the macOS suite was not rerun — the revision 1 macOS evidence (483 passed / 17 skipped at exit 0, strict mypy exit 0 across 55 source files) stands as recorded and is not re-asserted as fresh.
