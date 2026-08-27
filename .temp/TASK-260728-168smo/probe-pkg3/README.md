# TASK-260728-168smo — cycle 3 reproduction package

Everything needed to re-run the `windows/amd64` qualification of decision 0010
and to re-read the macOS evidence the exclusion rests on.

## Layout

| Path | Contents |
|---|---|
| `win/` | every PowerShell script run on the Windows host, plus the two transport helpers |
| `win-evidence/` | the raw host logs those scripts wrote, and the parsed ETW trace |
| `mac-cycle2/` | the cycle-2 macOS sandbox profiles, the two iterative-denial enumerators, their raw logs and the final allow file — retained because correction C1 of the evidence log is a re-reading of exactly these |

## Transport

`psrun.sh <script.ps1>` runs a short script through
`powershell -EncodedCommand` over `ssh win`. It is limited by the ~8 KiB
command-line ceiling; a longer script silently fails with
"the command line is too long", which is why the four suites use the second
helper.

`psfile.sh <script.ps1>` copies the script to `C:\kn168\<name>` with `scp` and
runs `powershell -NoProfile -NonInteractive -ExecutionPolicy Bypass -File`.
Use this for anything non-trivial.

## The decisive runs, in order

| Script | What it establishes |
|---|---|
| `probe1.ps1` | host inventory: no `java`, no `~/.konan`, admin, 78.7 GB free, `logman`/`tracerpt`/`certutil`/`netsh` present |
| `p1-fetch.ps1` | downloads and checksum-verifies the Kotlin/Native release and a Temurin JDK against their published digests |
| `p2-bundle.ps1` | unpacks both into the `curator-kotlin-bundle-v1` layout; proves the distribution has **0** `*.exe` |
| `p3-probe.ps1` | reads `run_konan.bat` and `konanc.bat`; runs both Stage A probe vectors; extracts the `mingw_x64` / `linux_x64` / `remote:internal` keys from `konan.properties` |
| `p4-hydrate.ps1` | the one online hydration run; records the four dependencies and their byte counts |
| `p5.ps1` | curated tree digest, 27,867 files |
| `suiteA.ps1` | writable-data-dir compile, the one-file mutation, digest restoration |
| `suiteA3.ps1` | the `.extracted` requirement, and the read-only / overlay sequence with a corrected principal |
| `suiteB.ps1` | first ETW-traced compile |
| `suiteC.ps1` | probes against an empty data dir, `@`-token table, launcher control |
| **`final.ps1`** | **the consolidated qualification run: R1–R10.** This is the script to re-run |
| `cleanup.ps1` | lifts the deny ACE, stops the ETW session, deletes the firewall rule, removes `C:\kn168`, and verifies each |

`fixacl.ps1` and `diag.ps1` are kept because they record two real traps rather
than accidents: `icacls /deny "...(W,D,DC)"` denies `SYNCHRONIZE` as part of the
simple right `W` and makes the JDK unrunnable, and a per-user deny does not stop
a delete when the same account is in `BUILTIN\Administrators`. Both are written
up as obligation K-11.

## Re-running

```bash
./psfile.sh p1-fetch.ps1     # ~430 MB of verified downloads
./psfile.sh p2-bundle.ps1
./psfile.sh p3-probe.ps1
./psfile.sh p4-hydrate.ps1   # ~465 MB, the only online step, ~70 s
./psfile.sh p5.ps1           # ~4 min, hashes 2.4 GB
./psfile.sh final.ps1        # R1-R10, ~15 min
./psfile.sh cleanup.ps1
```

`final.ps1` is self-contained apart from the bundle: it lifts and re-applies the
ACL itself, rebuilds the overlay, takes its own ETW trace, parses the PE import
directory in-process without invoking any tool, adds and deletes its own
firewall rule, and recomputes the tree digest at the end. The baseline digest
`63d96ff7c488e713dedbf7029237cfc6cd030ae4c1caf11c8ba2274395badae3` is hard-coded
in R1 and R10 as the comparison target; on a re-run with the same two verified
archives it must match.

## Reading the ETW trace

`win-evidence/etw-f.xml` is `tracerpt` XML output for the R3 trace.
`ProcessStart` is `EventID` 1 from provider `Microsoft-Windows-Kernel-Process`.
Each event carries `ProcessID`, `ParentProcessID` and `ImageName` as an NT
device path; `final.ps1` resolves those to drive letters with `QueryDosDevice`
and then takes the transitive parent-PID closure of the compiler JVM. Two
images are below it, both inside the bundle.

## Correcting the macOS record

`mac-cycle2/run7-enumeration.log` is the run seeded from an empty allow file;
`run8-enumeration.log` is the run seeded with its four discoveries.
`allowed-externals.txt` is `run8`'s final file and legitimately holds six paths.
The union across both runs is seven. The set sufficient for a successful compile
is six. See correction C1 in the evidence log.
