# Android OS research — reviewer A record

Task: TASK-261010-3751vf — research-android-as-host-os
Change Request: CR-TASK-261010-3751vf-1, revision 1.
Candidate tree: `defcc10615407e848f65345f6f47a23bc9f94873`.

**Independent recommendation: accept. No P0/P1 findings. Record-only: reviewer B must reconcile this record and decide acceptance.**

Role resolution: the prompt carries both A and B briefs. The task launch notes explicitly identify this tracked run as R138 reviewer A, record-only. No other reviewer record was read. No accept_cr, reject_cr, done transition, or commit acknowledgement was supplied.

## Swept surfaces

| Surface | Result |
| --- | --- |
| Citations and version scope | 16/16 selected load-bearing claims checked against primary source text below; Android code pinned to android-16.0.0_r1, systemd v257, and Fedora CoreOS docs at the report’s exact commit. Rolling pages are identified as rolling, not immutable evidence. |
| Mapping and verdict | Covers the overview’s M1–M6, S1–S7, B1–B6, added services and ten principles. Host recommendation follows the UID, harness, key and lifetime gaps; APK, native image and guest guarantees remain distinct. |
| Platform contracts | Reviewed pinned module contract passages and wiki chapters 02–11 plus architecture sections on module ownership/principles. Correctly gives PTY/app-server and renewal to the runner, preserves generation checks and current authority, and treats broker-delivered vendor tokens as exposed to the harness. |
| Experiment | Proposed only; one hosted worker, two hours, no full AOSP build; failure to provision is an environment result. Shell-domain success is not app/native-domain qualification, and hardware trust remains untested. No working-machine fallback. |
| Scope and hygiene | Exact base-to-candidate delta is one 291-line research file. Working artifact equals candidate and attached research outcome byte-for-byte. No LOGBOOK delta, secrets, personal paths, local hostnames or session links found in manual inspection. |

## Claim checks

| Claim | Primary source | Finding |
| --- | --- | --- |
| UID/GID collision | [AID](https://android.googlesource.com/platform/system/core/+/android-16.0.0_r1/libcutils/include/private/android_filesystem_config.h) | Android 16 r1 defines app IDs 10000–19999 and external/shared GIDs spanning 30000–59999. The report correctly limits the conflict to the same-number UID/GID contract. |
| Package allocation | [APPIDS](https://android.googlesource.com/platform/frameworks/base/+/android-16.0.0_r1/services/core/java/com/android/server/pm/AppIdSettingMap.java) | acquireAndRegisterNewAppId allocates and registers package setting identities; this is not an arbitrary persistent OS-account API. |
| App-data execution | [APP-EXEC](https://android.googlesource.com/platform/system/sepolicy/+/android-16.0.0_r1/private/app_neverallows.te) | The execute_no_trans neverallow covers modern untrusted apps with named legacy/run-as exclusions. The report limits this to the ordinary modern app deployment. |
| Key algorithms | [KEYMINT](https://android.googlesource.com/platform/hardware/interfaces/+/android-16.0.0_r1/security/keymint/aidl/android/hardware/security/keymint/IKeyMintDevice.aidl) | The pinned interface requires TEE Curve25519 and explicitly excludes it for StrongBox; the exposed API does not offer BIP-32/SLIP-0010 child derivation. |
| Keystore access | [KEYSTORE](https://source.android.com/docs/security/features/keystore) | The documented daemon stores encrypted key blobs; operations run through KeyMint and its trusted application. APP/SELINUX namespaces support the claimed access separation. |
| Runner lifetime | [INIT-REAP](https://android.googlesource.com/platform/system/core/+/android-16.0.0_r1/init/service.cpp) | Service::Reap invokes KillProcessGroup(SIGKILL) for normal services and modern oneshot cleanup. Detachment alone is not restart-survival evidence. |
| Memory pressure | [LMKD](https://source.android.com/docs/core/perf/lmkd) | The documented daemon selects processes for killing under memory pressure; it does not preserve session state. |
| Mobile background lifetime | [DOZE](https://developer.android.com/training/monitoring-device-state/doze-standby) | The documentation defers app CPU/network work and restricts network access during Doze. The research explicitly distinguishes native init daemons. |
| Verified boot | [AVB](https://source.android.com/docs/security/features/verifiedboot/avb) | The documentation describes signed dm-verity metadata, bootloader integration and rollback protection; these do not establish application-state recovery. |
| A/B recovery | [AB](https://source.android.com/docs/core/ota/ab) | The inactive slot is updated and boot failure can fall back to the previous OS. The research does not confuse this with mutable-ledger rollback. |
| VM userspace | [MICRODROID](https://source.android.com/docs/core/virtualization/microdroid) | Microdroid is a mini Android payload OS with Bionic and Binder; it is not a ready-made glibc workstation. |
| Dynamic identities on Linux | [SYSTEMD](https://github.com/systemd/systemd/blob/v257/man/systemd.exec.xml) | v257 documents recycled dynamic UIDs, persistent-file hazards and descriptor-passing caveats. StateDirectory does not provide Swarma generation semantics. |
| Fedora CoreOS MAC | [FCOS-SELINUX](https://github.com/coreos/fedora-coreos-docs/blob/1fec1acb5437220c559b8833e0ef1d643edd918b/modules/ROOT/pages/selinux.adoc) | The pinned document explicitly states SELinux enforcing mode by default and supports custom policy. |
| Fedora CoreOS updates | [FCOS-UPDATE](https://github.com/coreos/fedora-coreos-docs/blob/1fec1acb5437220c559b8833e0ef1d643edd918b/modules/ROOT/pages/auto-updates.adoc) | The pinned document identifies OSTree deployments, rpm-ostree, Zincati and reboot finalization coordination. |
| Flatcar MAC default | [FLATCAR-SELINUX](https://www.flatcar.org/docs/latest/security/encryption/selinux/) | The inspected rolling guide explicitly says enforcement is not default. The report correctly asks for chosen-image verification rather than treating old examples as universal release evidence. |
| Hosted probe feasibility | [CUTTLEFISH](https://source.android.com/docs/devices/cuttlefish/get-started) | The getting-started guide requires virtualization support and matching host/image artifacts. The proposed environment-failure outcome is therefore necessary and correctly scoped. |

## Architecture evidence

- [UM](https://github.com/relux-works/swarma-user-manager/blob/92bdf1d59478b8fd28d2d98cfe9fb95716d10e8e/spec/helper.md): §4.1 fixes same-number UID/GID ranges, generations and reuse cleanup; launcher contract requires an adapter rather than dropping authentication.
- [CB](https://github.com/relux-works/swarma-credential-broker/blob/f910e68f882900f03ffd036ecc2481f35bf8988b/spec/broker.md): §4.1–4.2 binds kernel peer identity and active generation; §5.5 distinguishes declared from enforced network policy; delivered tokens cannot be retroactively erased.
- [DISP](https://github.com/relux-works/swarma-dispatcher/blob/9cec4115b8666aa1f3325a792359135a9df5a2f7/spec/dispatcher.md): §3 and §5–7 preserve caller-chain intersection, durable effect identity, capacity ownership and confirmed cleanup.
- [SH](https://github.com/relux-works/swarma-session-host/blob/23d290dac24c0fc0c0283df8d384a54054c7fff7/spec/session-host.md): §3.2 explicitly assigns PTY/app-server and credential renewal to the per-agent runner; §3.4 preserves fences, epochs and operator update locks.
- [ARCH](https://github.com/relux-works/wiki/blob/2e30edbd54977b306ab53497663fcd94a5ca6cb6/session-host/architecture.ru.md): The module and principle inventories agree with the research mapping; the research correctly gives newer contracts precedence over the older overview.

## Findings

```yaml
findings:
  - id: A-N1
    severity: note
    severity_reason: "P2: rolling documentation and a virtual-device probe cannot qualify a production image or physical key/boot stack; this is already disclosed and does not block acceptance."
    disposition: "Keep exact-image defaults and physical trust verification in the later implementation gate."
blocking_findings: 0
independent_recommendation: accepted
```

## Evidence bounds and access anomalies

No builds, tests, emulators or experiments were run. Runtime gates executed: 0. This is a documentation/source review, not a green runtime suite. No prior attached runtime evidence was accepted as proof.

The public module specifications and pinned Android/systemd/Fedora sources were retrieved independently. Access-controlled wiki input was retrieved through authenticated repository reads for this review; source copies are not attached to the public board. One browser read of the Doze URL failed internally; the same official page with the English-language query loaded successfully. Initial outcome retrieval by basename failed because the registered resource uses a task-scoped nested name; the registered resource was then retrieved and compared successfully. Neither anomaly was treated as source absence.

Reviewed artifact SHA-256: `d0b00f13537585e849e7c400a53a71e9074517be543132ca011485a294dc4547`.

Final role action: attach this record as TASK-261010-3751vf/android-review-A.md and return to the coordinator for the separate cross-provider deciding review. LOGBOOK is unchanged under the explicit brief exception.
