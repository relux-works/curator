# Review verdict (reviewer B, deciding) — CR-TASK-261010-3751vf-1 rev1

Verdict: ACCEPT. No P0/P1 finding. Five P2 notes (tb-R226). Written before reading reviewer A's record `android-review-A.md`.

Delta: exactly one added file, `.research/261010_android-os-base.md` (291 lines, 60622 bytes). No `LOGBOOK.md` change. Hygiene grep (local paths, home dirs, host names, emails, key/token patterns, session links) found nothing; the only hits were the words "session host" and the public wiki/spec URLs. No tests/builds/emulators were run on this host; verification was read-only retrieval of pinned sources.

## 1. Citations — 22 load-bearing claims checked, all match the pinned source
Pinned AOSP `android-16.0.0_r1` (fetched via googlesource):
1. UID allocation: `android_filesystem_config.h` — AID_APP 10000–19999; AID_CACHE_GID 20000–29999; EXT_GID 30000–39999; EXT_CACHE_GID 40000–49999; SHARED_GID 50000–59999; ISOLATED 90000–99999; USER_OFFSET 100000. Study's overlap claim with Swarma's 30000–59999 is exact.
2. `AppIdSettingMap.java` — app IDs indexed from FIRST_APPLICATION_UID, registered per package (supports "package identity allocation tied to registration").
3. App-data exec: `app_neverallows.te` — `neverallow ... {app_data_file privapp_data_file}:file execute_no_trans` with exclusions only `untrusted_app_25`, `untrusted_app_27`, `runas_app`. Study's "legacy target-SDK workaround" wording is accurate.
4. KeyMint: `IKeyMintDevice.aidl` — TEE must support P-224/256/384/521 and curve 25519 (Ed25519/X25519); "STRONGBOX IKeyMintDevices do not support curve 25519", StrongBox "must support P_256 and no other curves". No derivation operation in the interface (grep). Study exact.
5. init: `service.cpp` `Service::Reap` — kills the process group (`KillProcessGroup(SIGKILL)`) for non-oneshot/restart services, and for oneshot services when vendor Android version >= R. Study's hedge ("relevant normal paths") is correct; `killProcessGroup` is cgroup-based so "setsid is not an escape" is a fair inference.
6. netd: `netd.c` — `if (is_system_uid(uid)) return PASS/false` exemptions; `BACKGROUND_MATCH` cleared for `ifindex == 1` (loopback). Matches the study's two statements.
7. PTY/Binder/socket: `pty.cpp` has `openpty`/`forkpty`/`login_tty`; `Binder.getCallingUid()`/`clearCallingIdentity` present; `android_net_LocalSocketImpl.cpp` references `SO_PEERCRED`.
8. FGS: developer.android.com — dataSync and mediaProcessing 6 h / 24 h, tracked separately, targeting Android 15+. Exact.
9. AVB/A-B: AVB page lists rollback protection and dm-verity metadata generation; A/B page: unused slot kept as fallback, update_verifier marks boot successful after dm-verity check, AVB not required for A/B. Study's "boot rollback is not data rollback; anti-rollback forbids downgrade" is consistent.
10. Microdroid: Bionic-based payloads, native binary in APK; Java limited. Study exact.
11. lmkd: "reacts to high memory pressure by killing the least essential processes". Exact.
12. Build requirements: 64-bit x86 Linux, ≥400 GB disk, 64 GB RAM. Exact. Cuttlefish: x86_64 and ARM64, KVM, host package from same build, reboot after install (kernel modules/udev). Exact — and the study's env-failure fallback handles the reboot problem on hosted runners.
13. systemd v257 `DynamicUser=`: "UID/GIDs are recycled after a unit is terminated"; warns about leftover files and AF_UNIX directory-fd passing. Exact.
14. Fedora CoreOS (docs commit 1fec1acb): "SELinux enabled in enforcing mode"; OSTree/rpm-ostree/Zincati, reboot strategies. Exact.
15. Flatcar: "implements SELinux, but currently does not enforce SELinux protections by default"; reboot resets to permissive. Study's "enforcement is not default" is exact.
16. Talos v1.11: Secure Boot signs systemd-boot and UKI; TPM2 disk encryption sealed to PCR 7. Exact.
17. Node v22.14.0 BUILDING.md: "Android is not a supported platform", no CI testing. Exact.
18. Codex README @806d973: musl Linux x86_64/aarch64 + macOS artifacts, no Android. Exact.
19. Claude Code setup: macOS/Windows/Ubuntu/Debian/Alpine only; npm installs the same native binary, which "does not itself invoke Node". Exact.

## 2. Verdict follows from the mapping, scoped by deployment model
The headline answer is explicitly conditional on APK / custom AOSP / Linux guest (lines 7, 173–177, 199); the Linux-guest and Microdroid results are never credited to native Android. The verdict table (187–190) follows from the tables: conflict items (UID range, runner reap, key curve, harness) drive "do not select AOSP now"; fit items drive node and reference-design "yes". "Conditional option" for an appliance carries stated preconditions. Claims about weaker Linux/macOS fault tolerance are explicitly declined (line 165).

## 3. Platform specs read correctly
UM spec: agent UIDs 30000–59999, services 900–999 Linux, sudo + `SUDO_UID` entry, never-reused generation — all as the study states. SH spec §3.2 ownership table: the per-session runner (`swarma-session-runner`) owns the PTY master/app-server transport and the credential lease client; controller re-adopts under a new epoch; explicit stop stays stopped; the study's SH-2 reading and its note that the Oct 7 overview is superseded are correct. Broker/dispatcher/book B02–B11 (access-controlled wiki): not independently re-read by me beyond the public specs — see note N2.

## 4. First experiment
Bounded (one worker, two hours, no full AOSP build), hosted-CI only, "never fall back to a working machine", pinned build ID + digests, passed/attempted counted per lane, expected rejections preserve nonzero status, no permissive SELinux, and the three outcomes include "environment prevented the probe" explicitly not counted as evidence against Android. Honest and executable.

## 5. Scope and hygiene — pass (see Delta above).

## Findings
- N1 (note): ~40 rolling documentation pages are cited by access date, not pinned. Disclosed in the register; acceptable for a conceptual study.
- N2 (note): book chapters B02–B11 are access-controlled; only the four public specs were independently re-read here.
- N3 (note): Step 1 of the experiment depends on hosted runners offering KVM and tolerating a Cuttlefish host-package reboot; the study provides a hosted-VM fallback. A first implementer should expect an environment outcome and budget for it.
- N4 (note): Step 3 (APK PTY probe) is a modest scope extension of a probe whose headline is the harness binary; it is gated as optional by the "otherwise mark unknown" rule.
- N5 (note): Bottlerocket, NixOS, ChromiumOS/minijail rows were not re-fetched in this review; their cells are hedged ("do not infer", "separate qualification") and make no unqualified guarantee claim.

No P0, no P1.

## Reconciliation with reviewer A
`android-review-A.md` does not exist on the task: reviewer A's run (RUN-261010-383b33) ended without a record and its auto-recovery successor was cancelled by the operator. There are therefore no A findings to reconcile and no P0/P1 raised by A. The verdict above stands on B's independent checks alone; the absence of A's cross-check is a process gap, not evidence about the study.

Decision: accept_cr (no P0/P1 remains; N1–N5 are P2 notes under tb-R226).
