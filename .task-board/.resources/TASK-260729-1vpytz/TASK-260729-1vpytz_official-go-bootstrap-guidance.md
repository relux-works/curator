# TASK-260729-1vpytz — official Go bootstrap guidance input

**Research date:** 2026-07-29  
**Consumers:** `TASK-260728-ypbuav` and CocoaSkills Windows Go qualification  
**Scope:** official installation/archive guidance for macOS, Windows, and
Linux; read-only local macOS, `ssh win`, and `ssh lev` inventory; mapping to accepted
`go-v1`. No software, PATH, profile, registry, product, catalog, pin, stage,
commit, or publication change was made.

## 1. Executive finding

The official Go catalog currently exposes two supported stable families:
**Go 1.26.5** and **Go 1.25.12**. The accepted Curator implementation
allowlists only family **1.25**. The Go 1.23 protocol floor is not an open-ended
compatibility promise.

Therefore:

1. Use the supported **Go 1.25.12** patch for current Curator/CocoaSkills
   `go-v1` qualification.
2. Reject Go 1.26 until that family passes the same `go-v1` conformance vectors
   and the manager allowlist is deliberately expanded.
3. Treat PATH as discovery only. Admission requires a manager-approved absolute
   `GOROOT`, its exact native `bin/go[.exe]`, a clean probe that returns the
   same root, and a full `curator-go-toolchain-v1` tree fingerprint.
4. Never download, install, upgrade, uninstall, or run package/installer
   instructions automatically. A missing/unqualified toolchain is a fail-closed
   operator action.

The Windows host is x64 Windows 10 Pro and has no Go executable on PATH, Go MSI
entry, or Go tree at the conventional roots inspected below. The Linux host is
Ubuntu 26.04 LTS x86_64 and likewise has no Go executable or conventional Go
root. Its OS repository currently offers Go 1.26, which is not an accepted
`go-v1` family. The local arm64
macOS host has three real 1.25-family trees—goenv 1.25.5, Homebrew 1.25.5, and
official `/usr/local/go` 1.25.1—but none is current or automatically trusted.

## 2. Current official distribution facts

All upstream facts were retrieved from official `go.dev` sources on
2026-07-29.

### 2.1 Versions and binary shapes

The official policy supports a major Go release until two newer major releases
exist. Both 1.26 and 1.25 are upstream-supported today, but only 1.25 is
accepted by the current manager.

| Family | Official patch | Upstream | Accepted `go-v1` |
| --- | --- | --- | --- |
| 1.25 | 1.25.12 | Supported stable | **Yes; recommended bootstrap** |
| 1.26 | 1.26.5 | Supported stable/latest | **No; qualify and allowlist first** |
| 1.23 / 1.24 | Older families | Protocol floor may be met | No; not allowlisted |
| <1.23, future/unknown, custom | Any | Irrelevant | No |

| Platform | Official architectures | Package shapes | Official/default root |
| --- | --- | --- | --- |
| macOS | `amd64`, `arm64` | `.pkg`, `.tar.gz` | Package: `/usr/local/go`; PATH file: `/etc/paths.d/go` |
| Windows | `386`, `amd64`, `arm64` | `.msi`, `.zip` | MSI: `Program Files` or `Program Files (x86)`; x64 normally `C:\Program Files\Go` |
| Linux | `386`, `amd64`, `arm64`, `armv6l`, `loong64`, `mips`, `mips64`, `mips64le`, `mipsle`, `ppc64`, `ppc64le`, `riscv64`, `s390x` | `.tar.gz` | `/usr/local/go` |

“Official distribution” here means exact artifacts from
`https://go.dev/dl/`. Homebrew, Chocolatey, Scoop, and OS repositories are not
the primary artifacts for this catalog. The official
`golang.org/dl/goX.Y.Z` wrapper and default Go toolchain switching can download
toolchains, so both are excluded by the no-auto-install boundary.

### 2.2 Selected Go 1.25.12 artifacts

The operator must recheck the live JSON before use:
`https://go.dev/dl/?mode=json`.

| Target | Official artifact | SHA-256 |
| --- | --- | --- |
| macOS arm64 pkg | [download](https://go.dev/dl/go1.25.12.darwin-arm64.pkg) | `a3d7891214fbcd1a31398ddfb2b6f68ebc01a0f05e5398d9b23e63bcea5d557e` |
| macOS arm64 archive | [download](https://go.dev/dl/go1.25.12.darwin-arm64.tar.gz) | `fa2c88bbcf64bd3b2aef355f026cfec6d3a4a01c132f999c8f8c964eb767164f` |
| macOS amd64 pkg | [download](https://go.dev/dl/go1.25.12.darwin-amd64.pkg) | `2c6b49d08e1ba340d0608d4c64ea5dbbd7f635763dec456fb87ad0d0d6629367` |
| macOS amd64 archive | [download](https://go.dev/dl/go1.25.12.darwin-amd64.tar.gz) | `00a2e743b82bccec03c51c4b0f7e46d5fec52184075fd6c5183c3bb39ae9fb00` |
| Windows amd64 MSI | [download](https://go.dev/dl/go1.25.12.windows-amd64.msi) | `45bc4ffd130e778374818551790abc2b4378dc5e89e46fcd114627ec9ebc1687` |
| Windows amd64 ZIP | [download](https://go.dev/dl/go1.25.12.windows-amd64.zip) | `d5dc82da351b00e5eedd04f41356817d674cc4308131f0f638a5b14c5c3af4cb` |
| Windows arm64 MSI | [download](https://go.dev/dl/go1.25.12.windows-arm64.msi) | `54e82ed1f76c6038f728e8025a830edfb4f0e30c3d7fd6442a2d21bb1c568e56` |
| Windows arm64 ZIP | [download](https://go.dev/dl/go1.25.12.windows-arm64.zip) | `054f046a5fa31fdcc9491cc19065cbf43bf521d805bbe298ae8d65dd981fca84` |
| Linux amd64 | [download](https://go.dev/dl/go1.25.12.linux-amd64.tar.gz) | `234828b7a89e0e303d2556310ee549fbcf253d28de937bac3da13d6294262ac1` |
| Linux arm64 | [download](https://go.dev/dl/go1.25.12.linux-arm64.tar.gz) | `8b5884aef89600aef5b0b051fb971f11f49bb996521e911f30f02a66884f7bd2` |

## 3. Accepted trust and compatibility mapping

### 3.1 Discovery is not admission

`command -v`, `type -a`, `where.exe`, installer-added PATH, version-manager
shims, symlinks, ambient `GOROOT`, and a plausible `go version` are discovery
evidence only.

The manager must receive an operator-owned selection before entering a
package-controlled directory:

```text
approved_goroot = canonical absolute real directory outside repository/runtime roots
approved_go     = <approved_goroot>/bin/go       # Unix
approved_go     = <approved_goroot>\bin\go.exe   # Windows
```

The launcher must be a regular native executable inside that exact tree, not a
wrapper. A clean `go env GOROOT` must resolve to the independently derived root.
Recommended catalog fields:

```yaml
family: "1.25"
version: "go1.25.12"
goos: "darwin|windows|linux"
goarch: "arm64|amd64|..."
distribution_url: "https://go.dev/dl/..."
distribution_sha256: "<official SHA-256>"
approved_goroot: "<canonical absolute root>"
approved_go: "<approved_goroot>/bin/go[.exe]"
toolchain_identity_algorithm: "curator-go-toolchain-v1"
toolchain_content_sha256: "sha256:<manager-computed full-tree digest>"
```

These fields must be immune to package, manifest, repository, runtime-root,
project `.agents/bin`, user PATH, and ambient environment override.

### 3.2 Artifact hash versus installed-tree fingerprint

The official SHA-256 authenticates the downloaded installer/archive. The
accepted `curator-go-toolchain-v1` algorithm separately binds every admitted
directory, regular file, allowed internal relative symlink, and normalized
`go version` record. It rejects special files, escaping/absolute links,
wrappers, and clean-probe root mismatch.

A `bin/go` hash, package receipt, MSI entry, or archive hash is not the full
installed-tree fingerprint. Recompute the full fingerprint after every install,
replacement, or upgrade and revalidate around use.

### 3.3 Required no-auto-install wording

> Curator and CocoaSkills MUST NOT download, install, upgrade, uninstall, or
> execute installer/archive/package-manager instructions for a Go toolchain.
> They MUST NOT accept a toolchain path, version, command, arguments, or
> instructions from a package, manifest, repository, runtime root, project
> shim, or ambient user PATH. If no manager-approved absolute root with a valid
> `curator-go-toolchain-v1` identity is available, the operation fails closed
> and displays operator-only guidance. Every probe/build uses the approved
> native executable directly with `GOTOOLCHAIN=local`; automatic switching or
> downloading is forbidden.

Official Go documentation says standard `GOTOOLCHAIN=auto` may search PATH and
download a newer toolchain. The accepted probe therefore starts clean with
`GOENV=off` and `GOTOOLCHAIN=local`.

## 4. Read-only inventory

### 4.1 Local macOS

**Observed:** 2026-07-29T00:40:09Z  
**Host:** macOS 26.5 build 25F71, Darwin arm64

```text
command -v go: /Users/iv/.goenv/shims/go
type -a go:
  /Users/iv/.goenv/shims/go
  /opt/homebrew/bin/go
  /usr/local/go/bin/go
  /Users/iv/.goenv/shims/go
ambient GOROOT: /Users/iv/.goenv/versions/1.25.5
```

| Candidate | Clean result | Launcher evidence | Assessment |
| --- | --- | --- | --- |
| `/Users/iv/.goenv/shims/go` | Dispatches to goenv 1.25.5 | 411-byte shell script; SHA-256 `968637fd94b7020b751cb5dc2669c160f3451bd0e1af03d50a011a917472b8fb` | Reject wrapper |
| `/Users/iv/.goenv/versions/1.25.5` | `go1.25.5 darwin/arm64`; matching root | Native Mach-O `bin/go`; SHA-256 `a84ad6c885e8a875475b9c6cea4f2e213f8ab449d43193520811c39b6007485c` | Family candidate; stale patch; not approved/fingerprinted |
| `/opt/homebrew/Cellar/go/1.25.5/libexec` via `/opt/homebrew/opt/go/libexec` | `go1.25.5 darwin/arm64`; clean root `/opt/homebrew/opt/go/libexec` | Native Mach-O; SHA-256 `11caa81dd6f1c2d6e2799f19265db880dd5a2c91b45703b4e871dbd90249f261`; `/opt/homebrew/bin/go` is a symlink | Family candidate; stale; PATH link is discovery only |
| `/usr/local/go` | `go1.25.1 darwin/arm64`; matching root | Native Mach-O; SHA-256 `c5f484a7e96e11b7c8c3361fa0c3d1c7480d65bf6fc92a03516719f2250d2be8`; root `root:wheel` | Official package but stale; not approved/fingerprinted |

`/etc/paths.d/go` contains `/usr/local/go/bin`; `pkgutil` reports
`org.golang.go`, version `go1.25.1`.

**Anomaly:** with ambient `GOROOT`, direct Homebrew and `/usr/local/go/bin/go`
calls both reported the goenv root. Empty-environment `GOENV=off
GOTOOLCHAIN=local` probes reported their correct roots. Even an explicit
executable is insufficient when inherited Go environment is admitted.

Only metadata, version/environment output, package receipts, and individual
launcher hashes were read. No candidate was mutated or promoted.

### 4.2 `ssh win`

**Observed:** 2026-07-29T00:38:38Z  
**Host:** Windows 10 Pro 10.0.19045, x64/AMD64, PowerShell 5.1.19041.6456

`Get-Command go.exe -All` and `where.exe go.exe` both returned absent. Process,
user, and machine PATH contain no Go directory. These roots are absent:

```text
C:\Program Files\Go
C:\Program Files (x86)\Go
C:\Go
C:\Users\admin\AppData\Local\Programs\Go
C:\Users\admin\go
C:\tools\go
C:\SDK\Go
C:\Users\admin\scoop\apps\go\current
C:\ProgramData\chocolatey\lib\golang\tools\go
C:\ProgramData\chocolatey\lib\golang\tools
```

Read-only checks also found no `GoProgrammingLanguage` registry key and no “Go
Programming Language” uninstall entry under HKLM/HKCU. There is no candidate
root to fingerprint. Windows qualification requires an operator-installed,
approved Go 1.25.12 amd64 root.

SSH separately warned that the connection lacked post-quantum key exchange.
This does not change the Go inventory result and is not toolchain evidence.

### 4.3 `ssh lev` Linux

**Observed:** 2026-07-29T00:48:47Z–00:48:59Z  
**Host:** Ubuntu 26.04 LTS (“Resolute Raccoon”), Linux
7.0.0-15-generic, x86_64, 64-bit, Bash login/process shell  
**User:** `lev` (`uid=1000`, ordinary home `/home/lev`, member of `sudo`)

The process PATH is:

```text
/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:
/usr/games:/usr/local/games:/snap/bin
```

`command -v go` returned absent. No Go file was found in the shallow read-only
scan of `/usr/local`, `/opt`, or `/home/lev`. These conventional candidates
were absent:

```text
/usr/local/go
/usr/bin/go
/usr/local/bin/go
/usr/lib/go
/usr/lib/go-*
/snap/go
/opt/go
/opt/curator/toolchains
/home/lev/.goenv
/home/lev/sdk
/home/lev/go
```

Relevant parent permissions are root-owned `0755` for `/usr/local`,
`/usr/local/bin`, `/usr/lib`, `/opt`, and `/snap`; `/home/lev` is user-owned
`0750`. `dpkg-query` reported no installed `golang*` package. `apt-cache policy
golang-go` reported `Installed: (none)` and candidate `2:1.26~1`.

Host mapping:

- `x86_64` selects the official `linux-amd64` archive, so the accepted manual
  bootstrap artifact is `go1.25.12.linux-amd64.tar.gz`.
- The Ubuntu repository's Go 1.26 candidate is upstream-current but outside
  the accepted 1.25 allowlist; product code must not install or select it.
- No candidate root exists to clean-probe or fingerprint. Linux `go-v1`
  qualification remains unavailable until an operator manually installs and
  approves a Go 1.25.12 amd64 tree.
- This task records a real host inventory, but still makes no claim that Linux
  execution controls or the full Linux qualification suite have passed.

## 5. Exact display-only operator steps

Curator/CocoaSkills may render these steps but must not execute them.

### 5.1 macOS arm64 package

1. Manually download
   `https://go.dev/dl/go1.25.12.darwin-arm64.pkg`.
2. Verify before opening:

   ```sh
   shasum -a 256 "$HOME/Downloads/go1.25.12.darwin-arm64.pkg"
   # a3d7891214fbcd1a31398ddfb2b6f68ebc01a0f05e5398d9b23e63bcea5d557e
   ```

3. Open the package and follow official prompts. It installs `/usr/local/go`
   and adds `/etc/paths.d/go`.
4. In a new terminal:

   ```sh
   APPROVED_GOROOT=/usr/local/go
   APPROVED_GO="$APPROVED_GOROOT/bin/go"
   test -d "$APPROVED_GOROOT"
   test ! -L "$APPROVED_GOROOT"
   test -f "$APPROVED_GO"
   test -x "$APPROVED_GO"
   file "$APPROVED_GO"

   env -i HOME=/var/empty TMPDIR=/tmp PATH=/usr/bin:/bin \
     GOENV=off GOTOOLCHAIN=local LC_ALL=C LANG=C \
     "$APPROVED_GO" version

   env -i HOME=/var/empty TMPDIR=/tmp PATH=/usr/bin:/bin \
     GOENV=off GOTOOLCHAIN=local LC_ALL=C LANG=C \
     "$APPROVED_GO" env -json GOROOT GOHOSTOS GOHOSTARCH GOOS GOARCH GOTOOLCHAIN GOENV
   ```

5. Require `go1.25.12`, `darwin/arm64`, and `/usr/local/go`; then run manager
   admission to compute/store the full fingerprint.

For an archive root, verify the matching official `.tar.gz`, extract into a new
empty staging directory, and publish a real directory such as
`/opt/curator/toolchains/go1.25.12`. Never extract over an existing tree.

### 5.2 Windows amd64 MSI

1. Manually download
   `https://go.dev/dl/go1.25.12.windows-amd64.msi`.
2. Verify before opening:

   ```powershell
   Get-FileHash -Algorithm SHA256 `
     -LiteralPath "$HOME\Downloads\go1.25.12.windows-amd64.msi"
   # 45BC4FFD130E778374818551790ABC2B4378DC5E89E46FCD114627EC9EBC1687
   ```

3. Open the MSI and follow official prompts. Normal x64 default:
   `C:\Program Files\Go`. Reopen terminals after install.
4. Verify:

   ```powershell
   $ApprovedGoRoot = 'C:\Program Files\Go'
   $ApprovedGo = Join-Path $ApprovedGoRoot 'bin\go.exe'
   Get-Item -Force -LiteralPath $ApprovedGoRoot
   Get-Item -Force -LiteralPath $ApprovedGo
   Get-FileHash -Algorithm SHA256 -LiteralPath $ApprovedGo

   Remove-Item Env:\GOROOT -ErrorAction SilentlyContinue
   $env:GOENV = 'off'
   $env:GOTOOLCHAIN = 'local'
   & $ApprovedGo version
   & $ApprovedGo env -json GOROOT GOHOSTOS GOHOSTARCH GOOS GOARCH GOTOOLCHAIN GOENV
   ```

5. Require `go1.25.12`, `windows/amd64`, and `C:\Program Files\Go`; then run
   manager admission and full-tree fingerprint.

For ZIP deployment, verify the official amd64 ZIP and expand it into a new,
administrator-owned real directory such as
`C:\Curator\Toolchains\go1.25.12`. Secure it against untrusted writes. Do not
trust `C:\Go` merely because it is conventional, or use a junction as the root.

### 5.3 Later Linux amd64/arm64

```text
x86_64  -> go1.25.12.linux-amd64.tar.gz
aarch64 -> go1.25.12.linux-arm64.tar.gz
```

```sh
sha256sum go1.25.12.linux-amd64.tar.gz
# 234828b7a89e0e303d2556310ee549fbcf253d28de937bac3da13d6294262ac1

sha256sum go1.25.12.linux-arm64.tar.gz
# 8b5884aef89600aef5b0b051fb971f11f49bb996521e911f30f02a66884f7bd2
```

Official default-root procedure:

```sh
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.12.linux-amd64.tar.gz
```

This is destructive and operator-only. Confirm the exact target and that no
manager process uses it. Official guidance explicitly forbids untarring a new
release over an existing `/usr/local/go` because that can break the install.
Post-install clean verification mirrors macOS with expected `linux/amd64` or
`linux/arm64`. A versioned `/opt/curator/toolchains/go1.25.12` is an operator
policy root, not an official default, and still needs full admission.

## 6. Upgrade, uninstall, and recovery

Upgrade:

- Never overlay archives; use a fresh tree.
- Verify the new official artifact first.
- Stop use of the old root before replacement.
- Re-run clean probes and recompute `curator-go-toolchain-v1`.
- Atomically update manager-owned catalog identity.
- A 1.25 patch change still needs a new fingerprint. Moving to 1.26 additionally
  needs family conformance and allowlist expansion.

Official uninstall guidance:

- Linux: remove `/usr/local/go`, remove its PATH entry, restart terminals.
- macOS package: remove `/usr/local/go` and `/etc/paths.d/go`, remove manual Go
  bin PATH entries, restart terminals.
- Windows MSI: use Control Panel or manually run
  `msiexec /x <exact-msi> /q`; the MSI route removes its environment variables.

Disable/remove the catalog record before uninstall. The product must not run
removal. Go user config/cache/module/bin locations (`GOENV`, `GOCACHE`,
`GOMODCACHE`, `GOBIN`) are separate and must not be deleted unless the operator
explicitly requests that additional data removal.

Missing/invalid-root behavior:

```text
do not search package/project/runtime PATH
do not download/install
report required family, host OS/arch, official URL/SHA, and config field
require operator installation/approval
rerun full manager admission
```

Unexpected fingerprint drift fails closed; never auto-accept the new digest.

## 7. Follow-on consumption

For `TASK-260728-ypbuav`, use sections 2–3 as the Go catalog row, section 5 as
display-only commands, and section 6 as recovery. The catalog must expose no
execution hook and must not permit package/repository override.

For CocoaSkills Windows qualification:

1. Operator manually installs official Go 1.25.12 amd64 MSI/ZIP.
2. Record official artifact SHA-256.
3. Supply exact absolute root through manager-owned configuration.
4. Prove native `go.exe`, clean GOROOT equality, `windows/amd64`,
   `GOTOOLCHAIN=local`, family 1.25, and full tree digest.
5. Only then run Windows Go parity/conformance.

The present absence of Go is an environment prerequisite, not justification
for PATH trust or automatic installation.

For later Linux qualification on `ssh lev`, the equivalent prerequisites are
the official Go 1.25.12 linux-amd64 archive checksum, a manually installed
root (`/usr/local/go` or an explicitly approved real versioned root), clean
`linux/amd64` probe equality, and the full installed-tree fingerprint. Do not
substitute Ubuntu's Go 1.26 package while the allowlist remains 1.25.

## 8. Fact-check and evidence

### Primary sources

| Official source | Claims |
| --- | --- |
| [Downloads JSON](https://go.dev/dl/?mode=json) | Versions, shapes, architectures, hashes |
| [Download/install](https://go.dev/doc/install) | Fresh Linux tree, macOS package root/PATH, Windows MSI default, version verification |
| [Manage installations](https://go.dev/doc/manage-install) | Downloading alternate versions, uninstall, separate user state |
| [Go toolchains](https://go.dev/doc/toolchain) | `auto` PATH/download behavior and `local` |
| [Release history/policy](https://go.dev/doc/devel/release) | Support policy and current patches |

Accepted local evidence:

- `.temp/TASK-260728-zb2s4z/release-probe/protocol/core.md` §§4.2, 8.2.
- `.temp/TASK-260728-zb2s4z/release-probe/profiles/manager.md` §2.2.
- `.temp/TASK-260720-1ljev5/worktree/internal/godriver/session.go`:
  `allowedFamilies={"1.25"}`, absolute selection, clean probes, fingerprint.

### Direct validations

| Validation | Exit | Result |
| --- | ---: | --- |
| Standalone `curl -fsSIL` on four official docs | 0 each | HTTP 200 |
| Standalone `curl -fsSIL` on selected macOS pkg, Windows MSI/ZIP, Linux amd64/arm64 archives | 0 each | HTTP 200 |
| `curl -fsSIL 'https://go.dev/dl/?mode=json'` | 56 | Expected-red: JSON endpoint returns HTTP 405 to HEAD |
| GET validation of download JSON | 0 | HTTP 200 |
| macOS inventory and clean probes | 0 | Section 4.1 |
| First SSH PowerShell attempt | 1 | Quoting/parser error; discarded; no mutation |
| Corrected encoded SSH inventory | 0 | Section 4.2 |
| Supplemental Windows roots/uninstall-key inventory | 0 | All absent |
| First Linux inventory | 127 | The read-only inventory had already established OS/arch/PATH/root absence, then an unescaped `dpkg-query` format variable tripped `set -u`; package conclusion discarded; no mutation |
| Corrected Linux package/root/permission inventory | 0 | No installed Go package/file; Ubuntu candidate is 1.26; parent permissions recorded |
| Artifact content/required-section assertions | 0 | All required platform, trust, inventory, and handoff terms present |
| `git diff --check -- LOGBOOK.md .research/260729_official-go-bootstrap-guidance.md` | 0 | No whitespace errors |
| Live JSON version/hash assertions with `curl` + `jq` under `pipefail` | 0 | Returned `true` for both versions and five selected artifact hashes |
| Accepted protocol-floor and implementation-allowlist assertions | 0 | Go 1.23 floor and exact `{"1.25"}` allowlist confirmed |

Claims were cross-checked across official download JSON, release history,
install/uninstall docs, and toolchain-selection docs. Compatibility was checked
against both accepted protocol text and implementation allowlist.

## 9. Recommendation

Publish a manager-owned **Go 1.25.12** entry with the exact official data above.
Install/replace a 1.25.12 macOS tree before new qualification evidence; do not
select the goenv shim, Homebrew link, or stale package tree automatically.
Require a human-installed/approved Go 1.25.12 amd64 root on Windows. The real
Ubuntu x86_64 host has now passed read-only inventory but has no Go; keep Linux
qualification unavailable until a manually installed 1.25.12 root passes
admission and native qualification.
