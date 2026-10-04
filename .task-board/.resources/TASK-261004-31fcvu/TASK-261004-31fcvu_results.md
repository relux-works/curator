# TASK-261004-31fcvu — rc3-release-notes: developer rework, revision 2

Ready for review. Only CHANGELOG.md differs from the recorded base; all work remains uncommitted. LOGBOOK.md is untouched. This outcome supersedes revision 1's duplicate/disposition claims.

## Review findings addressed

- P1: all 27 previously deleted entries are restored byte-for-byte in original order under `### Shipped in 0.15.0-rc.2 but not recorded in its notes` at the end of rc.3. The required explanatory sentence attributes these changes to rc.2. Independent section-scoped lookup finds **0/27 released twins**, while **27/27** occur in the tag's Unreleased section. Presence somewhere in a tag is not evidence of a released-note duplicate. No previous entry is now classified as dropped.
- P2: the concrete internal runner identifier is replaced with “select the explicit self-hosted runner label.” Scanned the entire rc.3 section, including historical entries, for the known identifier, machine-name patterns and personal user paths; no match. Earlier release bytes remain unchanged as required.

The fresh empty Unreleased and normalized rc.3 heading/date remain. All bytes from the rc.2 heading through EOF match the base. The rc.14 pin is `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, manifest `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`; production writes remain `curator-content-v1`. Known issues explicitly retain fp8vx7 as risk, rc.4 v2 writer deferral, N1–N4 unfixed (#106), and B3 exclusion.

## Direct validation and bounds

Commands were standalone processes; no tee or pipeline masks a gate result.

| Command | Actual exit | Result |
| --- | --- | --- |
| `python3 /tmp/TASK-261004-31fcvu_validate.py` | 0 | 54/54 mapped non-administrative commits; 258/258 board-only commits; 32/32 entry dispositions; 27/27 exact historical texts/order; prior release bytes; pin/write policy; known issues; public wording; CHANGELOG-only scope. |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --regression` | 0 | Three named tests; 27/27 single-entry omissions rejected; Unreleased cannot serve as released evidence; original and historical-subsection runner injections rejected. |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --regression --narrow-release-search` | 1 | Expected-red narrowing mutant: the gate refuses only entries absent from the whole rc.2 tag, allowing tag Unreleased to masquerade as released evidence. Both preservation regressions fail (all 27 omissions admitted). |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --regression --narrow-public-scan` | 1 | Expected-red narrowing mutant: only the fresh rc.3 prose is scanned, excluding the historical subsection. `test_internal_runner_identifier_rejected_in_rc3` fails on the omitted surface. |
| `python3 /tmp/TASK-261004-31fcvu_validate.py --candidate /tmp/TASK-261004-31fcvu_rev1.md` | 1 | Expected-red previous candidate: rejects exactly entries 5–20 and 22–32, all 27 review omissions. |
| `git diff --check` | 0 | Tracked diff whitespace validation. |
| `env -u GOROOT GOMAXPROCS=2 go build -p 2 -o /tmp/TASK-261004-31fcvu_curator ./cmd/curator` | 0 | Build rerun in this rework; output outside repository. |

The first document check exited 1 on entry 3 because its consolidation-token comparison did not normalize wrapped prose. Corrected that comparison only; historical preservation remains byte-exact. The corrected validator exited 0 twice. An initial run against the previous candidate exited 1 at subsection structure; after moving the preservation check ahead of structure, the recorded replay exits 1 specifically for all 27 omissions. The P1 narrowing mutant also exited 1 before its reporting was condensed; the attached log records the current concise run. Normal regression runs exited 0 before and after these validator fixes. Diff checks exited 0 on both executions. No failed gate is reported as green.

No product-code behavior changed. Tests and validator are board outcome artifacts, preserving the task's CHANGELOG-only repository scope. Go unit tests, full conformance/platform/race suites and release qualification were **not run** because this rework changes documentation only. The roughly 21,000 hosted passes are historical incident evidence from the readiness report, not a new locally executed test. Revision 1 reviewer content spot checks (11 commits) are accepted prior evidence, not rerun here; the entire current subject map and 258 administrative path boundaries were rerun independently. Source base and history range are unchanged.

## Regression contract

`ReleaseNotesRegression.test_prior_unreleased_requires_released_twin_or_verbatim_rc3` derives the 27 historical texts from base Unreleased, removes each individually from the actual candidate file, and requires the same preservation gate used by `validate` to reject it. `test_prior_unreleased_in_ambient_unreleased_is_not_released` proves neither tag nor current Unreleased can provide released-section evidence. `test_internal_runner_identifier_rejected_in_rc3` uses the label derived from unchanged base history, preventing disclosure in the test artifact.

Five entries (1–4 and 21) retain the original brief's explicit rc.3 consolidations. Their exact source texts are in the ledger; validation checks the designated group, a unique bullet and required content for each. Entry 2 deliberately supersedes stale “keep rc.13 pinned” wording with the released pin and the owned v2-write gap. This is an explicit transformation, not a duplicate/deletion exception. Every other entry must be verbatim in rc.3 or have a section-scoped released twin; the restored block also has a strict byte/order assertion. Coverage is 27/27 exact historical preservations + 5/5 explicit consolidations = 32/32, with no silently omitted entries.

The validator proves document preservation, structure and subject-map identity; semantic completeness of every commit is bounded by the reviewed subject-to-entry map and prior content inspection, not automatically proved by substring checks. Runner scanning detects the base's known label and declared private-path/machine-name patterns; the whole rc.3 section was also read. The public subject list neutralizes the internal identifier; per-subject SHA-256 values preserve exact original subject identity in the JSON without copying the identifier.

## Current-base disposition ledger

No dropped entries remain. Each source line range refers to base CHANGELOG.md; exact full text is retained in the companion JSON.

| Entry | Source lines | Entry begins | Disposition |
| --- | --- | --- | --- |
| 1 | 7–7 | Marker v3/v4 readers reject inconsistent external repository identities, substitution kinds, and effective revision widths. | represented: Fixed 3 |
| 2 | 8–8 | Accept the rc.14 candidate conformance digest with exact counts and gap accounting, including the snapshot v2-write gap owned by the pin cut-over; keep rc.13 pinned and v2 writers off. | represented: Changed 1 and Known issues 2 (stale rc.13 wording superseded by released rc.14 pin; v1 writes retained) |
| 3 | 9–9 | Windows executable resolution now proves platform ownership and every component-store hard-link origin before granting the captured System32 exception. | represented: Security 4 |
| 4 | 11–11 | Unix HTTPS askpass requests the secret only after accepting the password prompt, preventing broken-pipe transport errors on refusal paths. | represented: Fixed 4 |
| 5 | 15–32 | R5 script-worker-v1 runtime conformance qualification. All 33 named | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 6 | 33–49 | E4: umbrella provider lookup from trust roots (warning release, | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 7 | 50–64 | R4 script audit warning classes for `script-worker-v1` (manager profile | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 8 | 65–100 | R3 native probes, capability evidence, and preflight for | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 9 | 101–119 | R2 declaration-derived enforcement for `script-worker-v1`. Every enforced | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 10 | 120–136 | R1 script manager/worker invocation path (`script-worker-v1`). The manager | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 11 | 137–145 | Scoped HTTPS credentials for external build repositories. A `build_https` | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 12 | 146–151 | Operator documentation for scoped HTTPS build-repository token sources, | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 13 | 152–155 | Schema-8 first-party module roots for the `go-v1` driver: a build root may | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 14 | 156–182 | S6: shell-hook trust gate — warning release (`A-warning`). The POSIX and | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 15 | 183–192 | CI guard for the GoReleaser rc channel values. | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 16 | 193–205 | E2: direct-only `class: system` modules. Only the system modules of direct | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 17 | 206–223 | S4 warning release (`s4-warn`, audit finding S4): `profile install`, | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 18 | 224–230 | Conformance pin → v1.0.0-rc.12 (`dced9b8`): the hosted gate now runs | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 19 | 234–240 | The install transaction engine now caches canonical namespace resolutions | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 20 | 241–248 | A `go-v1` build root whose `vendor/modules.txt` carries a directory | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 21 | 252–252 | Project install now materializes skills at non-git product roots with a hygiene notice, while unexpected Git failures refuse installation. | represented: Fixed 1 |
| 22 | 254–265 | E4: the user-bin shim directory counts as manager-published — and | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 23 | 266–281 | Git snapshots are extracted from the object database (`git ls-tree -r -z` | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 24 | 282–291 | Git snapshot extraction now folds directory components per component when | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 25 | 293–300 | Status no longer reports a successfully installed schema-8 skill as | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 26 | 301–304 | Garbage collection no longer drops the live build references of a marker v4. | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 27 | 305–308 | A marker document at a readable schema that is nonetheless invalid is now | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 28 | 309–313 | A snapshot-extraction spawn failure of `git cat-file --batch` is now | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 29 | 314–328 | Draft Skillfile acquisition no longer consults user or system Git | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 30 | 329–350 | Draft literal-URL acquisition now builds git's environment from an | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 31 | 351–363 | Audit-registry snapshot verification no longer mistakes a snapshot | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |
| 32 | 364–371 | A `git check-ignore` spawn failure in the managed `.gitignore` gate is no | restored_verbatim: Shipped in rc.2 but absent from its released notes; restored verbatim in original order at the end of rc.3, explicitly attributed to rc.2. |

## Full non-administrative commit-subject → entry map

Compared against the full `git log v0.15.0-rc.2..origin/main` list: **54/54** hash/subject identities verified.

| Commit | Subject | rc.3 entry |
| --- | --- | --- |
| `6bb3c626c718b3f9aec5e1cf96aa5561de7bbf40` | STORY-261002-1r1hz7: STORY-261002-1r1hz7: curator-rc14-pin-and-v2-writer | Changed 1: released rc.14 pin/digest; Known issues 2: snapshot v2-write gap |
| `81fe98582dc2f50beb1d58e2d24970d9db1020b8` | Add the 2026-10 inline security audit report (EN + RU) | Known issues 3: inline audit N1–N4, #106, remediation story |
| `291688e4ed51c2619ca467e268943959f56a9124` | STORY-261003-3m6kh1: STORY-261003-3m6kh1: rc14-lockstep-carrier | Changed 1: rc.14 candidate exact counts/gap accounting (superseded rc.13 pin wording removed) |
| `f17ea7331c258b59db84fc6ea83872cf638b51a0` | Neutralise second-operator board elements and add tb-R163 (tb-R162 W4) | Added 6: public-board privacy guidance; board resource cleanup is administrative |
| `c224404db414f6eeb705ce62bfe0833d6b74cb42` | STORY-261002-1pd460: STORY-261002-1pd460: build-cache-sweep-live-executables | Fixed 2: global-upgrade/GC live-executable build-cache sweep |
| `ecc0ce7d506567422388408f41221d6cfa0bc71a` | STORY-260822-2lvw0e: STORY-260822-2lvw0e: curator-go-module-roots | Fixed 3: marker v3/v4/core-v5 cross-field validation (subject does not describe its actual payload) |
| `d989fd05f285a31b8ec343ed17c9d795b09e42d9` | STORY-261002-1w63le: STORY-261002-1w63le: spec-owner-review-triage | Added 6: spec-review research records |
| `ff52d4cbecfc2aeb0e38deff922a35435330ce16` | STORY-261002-2327ef: STORY-261002-2327ef: release-readiness-rc14-rc3 | Added 6: release-readiness research records |
| `a228444194512c162dfee211de051116b4504fc5` | STORY-260925-1v7pvn: STORY-260925-1v7pvn: windows-exec-hardlink-origin-validation | Security 4: Windows owner and all hard-link origin checks |
| `e5d632ac34c2e75aea6f27ef8c58e7b6d61e0d92` | STORY-261001-2peo2f: STORY-261001-2peo2f: rjxrgs-carrier | Changed 6: external lifecycle consumers, staging order, receipt-2 key, refusal diagnostics |
| `9b4459db7876f644574635c3c07339340d8cb774` | STORY-261001-1rrh6z: STORY-261001-1rrh6z: non-git-project-root-install | Fixed 1: non-git project root install |
| `e3b5048200c750eec3b96c8fee8bfe08181a1c2b` | STORY-261001-qabjyj: STORY-261001-qabjyj: research-docs-carrier-no-logbook | Added 6: launch-context research records |
| `e4a6a8d5df879183c4d8b5c648aa440bd50ff7dc` | STORY-261001-1xuwlu: STORY-261001-1xuwlu: muse-environment-curator | Added 1: Muse adapter, launch-env-fragment-v3, XDG parents and refused unverified channels |
| `0ef54f547deec35b1ebaae6fd9b7297d49d3f380` | STORY-261001-17ali8: STORY-261001-17ali8: askpass-pipe-refusal-epipe | Fixed 4: Unix askpass EPIPE refusal path |
| `87e21c0cfc533c854604dec69db056813c7967a3` | STORY-261001-38ijl9: STORY-261001-38ijl9: second-operator-guide-carrier | Added 5: second-operator walkthrough |
| `3265bc79d739eebde4d19a4117bf691586b8cd50` | STORY-260930-klqikd: STORY-260930-klqikd: content-hash-v2-curator | Changed 2: hash-v2 readers/versioned carriers and v1 writes |
| `88b76ac9afa35f201f54648589aa9e67fed7066f` | STORY-260930-3feoy5: STORY-260930-3feoy5: ext-repo-acquisition-conformance | Changed 6: external acquisition consumers and malformed-request pre-Git refusal |
| `dad3a0896e0e9e799620fecc1d64a9dbae450914` | STORY-260728-1ojb1p: STORY-260728-1ojb1p: curator-external-build-repositories | Added 5: external-repository guide; Changed 6: Windows artifact.exe cache naming |
| `a96218acb827e6b55446758ba9a5315c4a9403b7` | STORY-260930-9k3uil: STORY-260930-9k3uil: second-operator-quick-fixes | Fixed 1: first-run/bootstrap/inactive-profile UX |
| `67842e1275eca951bc941298083060c26bf0d980` | STORY-260930-15hioz: STORY-260930-15hioz: store-boundary-named-absence-rows | Added 7: protected-store named-absence tests |
| `4ee7121e3025ea9a33231d54ece5eb3af33d91b5` | STORY-260930-12oimr: STORY-260930-12oimr: test-git-config-isolation | Added 7: hostile Git-config isolation in tests/CI |
| `71360fcbf052462a29ddaa4dbb800dceb016fac0` | STORY-260930-jzq0dx: STORY-260930-jzq0dx: accept-content-hash-v2-candidate-suite | Changed 1: candidate content-hash-v2 exact counts and owned gaps |
| `6735fec729aaa670f21812fd7cc558c980db5ebb` | STORY-260928-9vt338: STORY-260928-9vt338: manager-interim-nul-opaque-rule | Security 6: NUL-containing opaque-input audit |
| `3901ca5c5b3a4c7e95bc7a676634554a2e9a20ce` | STORY-260930-2o0ybs: STORY-260930-2o0ybs: compiled-builds-board-reconciliation | Added 6: build-leaf reconciliation research |
| `b4e4be2ecceadc431c3ac8efb45fce88900a1c35` | STORY-260930-1soi12: STORY-260930-1soi12: [internal runner identifier omitted]-lane-runner-label | Fixed 5: explicit self-hosted runner label |
| `7eaf09ec367d78ca285d24ce08b6812b3c2df529` | STORY-260910-6bo7ej: STORY-260910-6bo7ej: tofu-and-equivocation-mitigations | Security 2: TOFU bootstrap checkpoints and mirror-group equivocation checks |
| `777b7fa568e8e22187b6d9176fa1685b3787213d` | STORY-260910-148pj1: STORY-260910-148pj1: profile-store-protected-boundary | Security 1: protected profile/private-store boundary checks |
| `600c2148f237de2a9b1434b93f614af45e431c77` | STORY-260929-1s4r14: STORY-260929-1s4r14: rust-lane-path-shadows-go | Fixed 5: Rust tool install preserves Go/Node PATH |
| `d5af3bd1d0b9dd83a8464c515fa7395a07eb5f11` | STORY-260910-234vmx: STORY-260910-234vmx: supply-chain-and-credential-hardening | Security 6: release installer checksum and archive verification |
| `29ebed123f535d89704ff06e4cde0a5ed0e3813b` | STORY-260928-1t6bto: STORY-260928-1t6bto: askpass-secret-via-pipe | Security 5: scoped HTTPS secret broker pipe |
| `2d59c6458664fd8ea74553a88c7fc43f9c42fa07` | STORY-260928-rp2r1j: STORY-260928-rp2r1j: top-level-sandbox-posture-docs | Added 5: README/SECURITY portable bounds and separately installed verified provider |
| `793581d503dbf5037240a743be8511ce3de741c0` | STORY-260929-1rfp7e: STORY-260929-1rfp7e: naming-gate-machine-echo-records | Fixed 5: naming gate ignores machine echo records |
| `7a4d18acc8ef0cba81d8c3562ba976cd926bc9f5` | STORY-260928-7eowfl: STORY-260928-7eowfl: path-boundary-walk-vanishing-entry | Security 1: disappearing-entry boundary walk |
| `dc16cf6200ca154619ecb4e62962a8e46c828f90` | STORY-260929-3f6aym: STORY-260929-3f6aym: naming-gate-binary-patch-noise | Fixed 5: naming gate ignores binary patches |
| `3f60f7f008ffaa5bc836d2202d204a5bc482d7f4` | Add the security_posture model (revision A) and the unreachable-registry gate (TASK-260927-4pv4au, TASK-260910-1sapuy via carrier TASK-260928-36r9k5) | Changed 4: security_posture A and unreachable-registry behavior |
| `aca32ebc179116c23b1995ec5648686b8dd6ce5a` | STORY-260928-3gzr9m: STORY-260928-3gzr9m: nofollow-parent-walk-guard-row | Added 7: nofollow parent-write production-entry coverage |
| `7444178daf67a96309eee0ec62847baf20d9dcfa` | Ship Codex seed revision A before revision B (TASK-260927-31gaka via carrier TASK-260928-2iu83q) | Changed 3: first tagged Codex seed A warning release |
| `7bc05184b9e7c459f4c9d321d6593cbe7f0ff8dc` | STORY-260928-lpnvkn: STORY-260928-lpnvkn: global-operation-lock-publication-ordering | Changed 5: global manager-home lock and publication ordering |
| `46175696b666fc525b70445b67bb53993ea761e9` | STORY-260928-1xu5sf: STORY-260928-1xu5sf: [internal runner identifier omitted]-scriptworker-identity-tests | Fixed 5: worker identity tests remove hard-link assumption |
| `fdb196970c38357c4f7ef76aee9d04d691bd055d` | STORY-260916-wgt8vz: STORY-260916-wgt8vz: path-kind-admission-and-boundary | Security 1: path-kind, ownership, permission and containment admission |
| `1308f02575308801f97db54fc5f1d55cbd16cc99` | STORY-260923-11vn9k: STORY-260923-11vn9k: [internal runner identifier omitted]-toolchain-and-shim-adoption | Fixed 5: self-hosted toolchain and shim adoption checks |
| `92f4e7b52670b87dfaf9e57e417a98e93cb8ad85` | STORY-260916-73a5zg: STORY-260916-73a5zg: managed-write-nofollow-rule | Security 1: managed-write nofollow boundary and publication recheck |
| `354729269439022ef2ee40ea5d51cd7cf6910d6b` | STORY-260928-16hi30: STORY-260928-16hi30: curator-global-adopt | Added 4: global adopt plan/dry-run/rollback |
| `c0400b1a78ee7ea9738f3fd52a5fe26de75f2570` | STORY-260916-ioemse: STORY-260916-ioemse: source-signer-allowlist-and-update-delta | Security 3: source signer allowlist/required signer and update deltas |
| `ead882c8da784176cfd6e7d131f3359dc2d0733a` | STORY-260916-1ll22r: STORY-260916-1ll22r: absence-vs-read-failure-collapse | Security 1: absence versus failed reads; no absence fallback on read error |
| `4b25b7951e8b243a2ce003bfaa8102d25965982b` | STORY-260916-1i1gfo: STORY-260916-1i1gfo: codex-seed-mcp-tables-residual | Changed 3: inherited Codex MCP seed records/status (revision A supersedes residual seed behavior) |
| `d292185ed8f1da7032081cfaaf2601dff3f82c3f` | STORY-260916-2otjbn: STORY-260916-2otjbn: umbrella-provider-trust-roots | Security 3: provider trust-root warning/status/refusal policy |
| `890d598a76499b28bff41193a80dc316a669967f` | STORY-260910-25yc0h: STORY-260910-25yc0h: records-boundary-binding | Security 2: page/checkpoint protected-store binding |
| `3ec26d9643ee4227f16a7906ca38cae103006c37` | STORY-260916-2d9coh: STORY-260916-2d9coh: direct-only-system-modules | Security 3: direct-only system modules and dropped/refused status |
| `28b61779aea43528a693fbb812a65d9f2a9c2b28` | STORY-260927-22m88w: STORY-260927-22m88w: env-unmanage-restore-backups | Added 3: env unmanage restore backups |
| `59a0ad1ae6fa67a74a373d661612894f9282a8f0` | STORY-260910-1lf0m5: STORY-260910-1lf0m5: bound-mcp-declaration-exposure | Added 7: bounded MCP declaration production-entry regression tests |
| `36ce784ebb32f14bf28bbd54767faf9ed869a416` | STORY-260916-8ql03k: STORY-260916-8ql03k: profile-install-reinstall-defects | Added 7: profile install/reinstall regression matrix |
| `69705ba44bae8880bfe04b60fb5235174cc0537d` | STORY-260924-3gd2d6: STORY-260924-3gd2d6: playbook-collection-acceptance | Added 7: collection install acceptance regressions |
| `1a57c71cb5545ff2561c8f2e8956e0387191a369` | STORY-260923-1lu2o3: STORY-260923-1lu2o3: 0018-curator-fragment-permissions | Added 2: Decision 0018 launch-env-fragment-v2 permissions; env migrate is explicitly described as existing rc.2 behavior |

## Full administrative subject list

258/258 commits independently checked to change only `.task-board/` paths; no product bullet is required for board lifecycle records.

- `876127f7c714e01092f43c8950dc879421c52461` Record STORY-261002-1r1hz7 board state
- `a779594cf5bda66b5dceb4e99a686994eb6126d5` Record STORY-261002-145yl1 board state
- `892014d0ae0573393d17a022277d075b77aa6e0b` Record STORY-261004-3oognx board state
- `68210eccfd656e770cf9101c2b927ceda01359df` Record STORY-261002-2x32ly board state
- `3a8dbf134b2774a9b9f768b2d38d2bb870b3d7f8` Record STORY-261003-3m6kh1 board state
- `1c756da29c0e7fe9708fbb5fd65675be77747aa0` Record STORY-261003-3m6kh1 board state
- `c58d6556636f6b7e5c6cf6c421aa15117b33fb21` Record STORY-261002-1pd460 board state
- `d310b658cd3899833158e16ed9dc95f792e3dc26` Record STORY-261002-1pd460 board state
- `c18143c3679af3957f9fe85462c529bde4ae1e59` Record STORY-260822-2lvw0e board state
- `2c7eb4eb3708cfde47af9a1f3f1a6ccf80d33392` Record STORY-260822-2lvw0e board state
- `c474e53e116a6141190e07caa4389baea779f25e` Record STORY-261002-1w63le board state
- `57e91e644601c46d1568a2f0053690875f62dab3` Record STORY-261002-1w63le board state
- `14d56554a9418359139401e2230fe29db2858321` Record STORY-261002-j60t04 board state
- `743438c5a1aa6d22313fbc14f81d38e0a4d8d6b6` Record STORY-261002-2rckyl board state
- `4e9d553f5c0dae393a749025f7d4999acd89a7bc` Record STORY-261002-2327ef board state
- `d47b64d3e56e89637951fabd6df6a98a1f984ac0` Record STORY-261002-2327ef board state
- `a2649b81609b6ec62bbe2c6cfddbe7406ae235d0` Record EPIC-260822-3ar0tv board state
- `764ad01e54f43aea55bd5b6fb72ee36fb80102c0` Record STORY-260925-1v7pvn board state
- `5e00dc6c905cf3df209f5f9aba131c5ebf7cec0f` Record STORY-261002-yf1tkx board state
- `2c2d0fdb500d906482b3b5258e6f0cd2a8935d91` Record STORY-261001-qabjyj board state
- `565877d07a179255cad61ba5239c0c2978ef4418` Record STORY-261001-luaymu board state
- `8ba1257e0b259e951ac7397cf82afe5f946e7306` Record STORY-261001-38ijl9 board state
- `6cfc430ce0ba9f69bf1f051459c5731ca83cc608` Record STORY-261001-30wj56 board state
- `21fa9efdd52d7f555cc11076f1bf5991d8c83f2b` Record STORY-261001-2peo2f board state
- `877cbb89c4b32118a4d45400ac2a45e3e7474619` Record STORY-261001-1xuwlu board state
- `c236a8921857f9d74b8246990b9a758a687ba5ad` Record STORY-261001-1rrh6z board state
- `8f9309b4029545cfb5d9d2f1d9a7219c008dba09` Record STORY-261001-17ali8 board state
- `c11674780e5b1218cd38e0b46f50bd6c6169fae3` Record STORY-260930-klqikd board state
- `bda7e9dd4fa1e14d683c79f520d980737730a5ef` Record STORY-260930-jzq0dx board state
- `46ce23a9f43c54a561255cdcfbf3bcde59850265` Record STORY-260930-9k3uil board state
- `10570fd6fef13cf677d1126c4a4df3f473dfd890` Record STORY-260930-3feoy5 board state
- `312ef83fd3d6720de1659ad56a2162bcda0bbe9b` Record STORY-260930-2o0ybs board state
- `953a536e0d14d435854bc2aef981db1eb9b69758` Record STORY-260930-2ekpps board state
- `04e61e53e3c637568ce1048735963e6e316ee517` Record STORY-260930-1soi12 board state
- `544d41bddaf1f65426c9f5f585202018f9f0650b` Record STORY-260930-15hioz board state
- `439fa5bd95f2c20096886d7ab8209445a81d23fa` Record STORY-260930-12oimr board state
- `ae3d0cba836c4c17d098ea95f9f0ac0e38cac73e` Record STORY-260929-3f6aym board state
- `682d23f621333d55206604a08bb6e6dc7904f7af` Record STORY-260929-1s4r14 board state
- `1b328f32cae06e0d0818eea21a31c962d705b2dc` Record STORY-260929-1rfp7e board state
- `315f3c7a27d366158a1341490bedf4a067c1ad80` Record STORY-260928-rp2r1j board state
- `d709eeadf664e1c198e4b626d65a45dc84e969bf` Record STORY-260928-lpnvkn board state
- `01ea1752ea90fdf82f32cdb2c76f570e91f3b205` Record STORY-260928-9vt338 board state
- `86855904140cae7d610a837e5429356c20538da7` Record STORY-260928-7eowfl board state
- `f25f3a6f3eb2c2661a4101034574a413eb85ae09` Record STORY-260928-3pdcpd board state
- `4838c87d4c00de3db0f9353a4f3282685f61b9dd` Record STORY-260928-3gzr9m board state
- `c74b6552047f1d19b68a02df77a0090afb89b3f9` Record STORY-260928-3fjnux board state
- `57ada3ad084c4c984e2f54869d0c6f8b673a7b81` Record STORY-260928-1xu5sf board state
- `98b3da69bf12d5670d607edfe08f19989f4ff510` Record STORY-260928-1t6bto board state
- `dc42af2a21caa743c2e0a2afd3d7a9bfe67e0c60` Record STORY-260927-3qf8er board state
- `546e1a36d557968eadcedaea3058803f81f7fa92` Record STORY-260927-22m88w board state
- `b99957e619b09510951bde35b804a2019a9992fb` Record STORY-260926-nn2j3l board state
- `26df73a8f4de76f79d57b806a1cd8881faa5f266` Record STORY-260924-txgta4 board state
- `4db2755be75be712ecc272a4288a21cf80f48214` Record STORY-260924-iafjfs board state
- `9e414c85626a38a2eccd0b3058b10b65db2e0ee8` Record STORY-260924-3gd2d6 board state
- `ec1a6ac7d74a85b7f3407aa8ce9c8b1e9069dea6` Record STORY-260924-3eywt2 board state
- `2b6c2e4b545792b3632e9fb0cf63f89e43eb4f7d` Record STORY-260924-2tyzhh board state
- `63210376c3f1bbb3a541d83979ec629c33a48e93` Record STORY-260924-2go2bz board state
- `6c50f578b867d7811d27e9bd5ee2f5ff30288280` Record STORY-260924-1oyh2m board state
- `9eea5931614fa86f2a6305290168a7b66b3f4306` Record STORY-260923-vkxt08 board state
- `45d09f549ff59211180a49edfa61f0689be4a58e` Record STORY-260923-laeycm board state
- `94219021d80da7f884e448f23b00382527f62dae` Record STORY-260923-3vwgy4 board state
- `805c4d272c8edff235d33e042449adb4573c3471` Record STORY-260923-3qwrnl board state
- `771c4c6c94a624463843441319d9d746d3779b30` Record STORY-260923-2mla0q board state
- `c77995afc9837d733ba2092764ecc4ff8bfce479` Record STORY-260923-2btaia board state
- `d2798aed37f9f0ede32d6bccfeb14ae055bb5867` Record STORY-260923-1v3no2 board state
- `9bcd0acbfa89007be190207b3b7bb85cdac04028` Record STORY-260923-1lu2o3 board state
- `ad68bccaaf8ac0498b40f8bcbc7a1b5517844b6c` Record STORY-260923-11vn9k board state
- `67f9036398b55f7f5465f65b8222cc1a561b0cf6` Record STORY-260922-2goxjs board state
- `0c9d70efc7262f588d0fcbbe38846862061416aa` Record STORY-260922-1cenbr board state
- `917c6d4390c5cb9a074b43ed52564e194da46774` Record STORY-260919-37szes board state
- `3f807564424e0dd535cd387a3d6a0e166c5c83cf` Record STORY-260918-2yvd86 board state
- `b279ef3eba7ada30876d8afbe4fe0b90dc6b17e7` Record STORY-260917-3w3lvj board state
- `90ebc37f09dd7f81e2e4bb49ca163814700ab558` Record STORY-260916-wgt8vz board state
- `b86223c4bdbcfa0fe9c02a9af90d2b066c0448cc` Record STORY-260916-v58b5y board state
- `e8227e418e6bda1771ed6dff9d56fc680b009ddb` Record STORY-260916-prdjid board state
- `cd168bc43e1dbb4face450b8b4d2030f51f206ce` Record STORY-260916-ioemse board state
- `6c585bd24b956880681bebec48d8e56403bfdafc` Record STORY-260916-8ql03k board state
- `b29e49742425bda20a3a9762a35483b0e1fc56ea` Record STORY-260916-73a5zg board state
- `e66e8bd02efe4b3c5fac296d47bf04a7931505da` Record STORY-260916-2otjbn board state
- `6e10d2b5fc548a26e41bc35c15d40a736d7496a9` Record STORY-260916-2d9coh board state
- `28ab233380ee589290f389c6140e8842a477fe54` Record STORY-260916-1on1d2 board state
- `12c4cd2b895c6bd9dd786bac6cf859dc325ea57f` Record STORY-260916-1nc5dc board state
- `815c11d755e774bf5e180cd56c859e63e7231964` Record STORY-260916-1ll22r board state
- `9865a793e8b253f597f95f8906e59262faadee19` Record STORY-260916-1i1gfo board state
- `dda562b411a5ebc9475a3af6eeb11ee19b5ff392` Record STORY-260915-3w11un board state
- `0dee04d96e6d7b23f0373c1b700eb0998075f02d` Record STORY-260910-6bo7ej board state
- `1a8f29b44df98084c1bd356070e8b3e990fa3faa` Record STORY-260910-3vxe3y board state
- `18b988aaef8bcb61d52e0649a392338914b24a12` Record STORY-260910-35tbgb board state
- `24184fa4f6ffe88683cc217df5b133b71fdb6ef8` Record STORY-260910-2qmrb8 board state
- `320d1dbdd53ce0a61d0a10eb6f6594f581b16cee` Record STORY-260910-2awkzu board state
- `78ac5749a7f110b27bec853c79259924efd8cd9f` Record STORY-260910-25yc0h board state
- `a03f00c6be89c94466b079ceb82dc1796c9266d6` Record STORY-260910-234vmx board state
- `e6993820c2e0fe7bcf1abdfe40b74eeb48ef6d03` Record STORY-260910-20sx61 board state
- `fda54772581a6185711a2b3b9782cc484fff698b` Record STORY-260910-1s75e1 board state
- `9782297ea14cfe1b25ec3e11c141aa1f4ff26759` Record STORY-260910-1lf0m5 board state
- `968555ee3b467c14d9ffd476272b56f0e8924f93` Record STORY-260910-1bhj0g board state
- `012386344b5cffca09b7b2e1dc901c670543214e` Record STORY-260910-197y84 board state
- `7753d1bd656321c831d401439d93a5e5d2e2b204` Record STORY-260910-148pj1 board state
- `67fd23ae56ecaab091e8309938b52572f96cc6c3` Record STORY-260908-sd6xkr board state
- `6305a5c567ac7a1eab0f76e6376bffbba9463d14` Record STORY-260908-g7o5zw board state
- `15d31b3a9d39dbd78e3295934053f7711fb3dca7` Record STORY-260908-3ry7gf board state
- `4910c5de358991fea0199b0783fdd31e557963b6` Record STORY-260908-2u6nly board state
- `ee2d1a7002948ddbed12bd4c2fad4a8f33cc6224` Record STORY-260907-2bddfc board state
- `fb751260763335ac12724ccce827520c04f55c81` Record STORY-260906-1a2i5a board state
- `3eb0ad024d81d1e4926261a9c3d758d86d530e1e` Record STORY-260905-2qvzwk board state
- `d5f2fb1bc524da0ed8eb5153d8fa930f99ce2250` Record STORY-260905-1n0iy8 board state
- `1f58a78c07fc30d483dab96e2d27aa681d3f9626` Record STORY-260822-2h0v9j board state
- `55430c9b2dcd49c9f282597ca0dd5c5f1de6e368` Record STORY-260728-1ojb1p board state
- `aeab533469d5647c2cbe5c8bbc126d6ea2573147` Record STORY-260930-2ekpps board state
- `c9fa063c4290afa49db3b6b16629b9122c501073` Record STORY-261001-2peo2f board state
- `850f474853f32b2447ff9432f5c5eb06fd5eae27` Record STORY-261001-2peo2f board state
- `dc8661c2dd54f25070fbe73bad276e9dabc629c5` Record STORY-261001-13bv7q board state
- `afdce865d668dd21231f94369e1e5791f36f15ad` Record STORY-261001-1rrh6z board state
- `16dc54b3c5ba80e6d61826e51b188a8371840537` Record STORY-261001-1rrh6z board state
- `02a13a423ee6800557a170c84a86bfbac582864d` Record STORY-261001-luaymu board state
- `55a668865a12cf8623dc83a21fea8bc5f768894f` Record STORY-261001-30wj56 board state
- `b8e40058096dc77cabe23c8cd6168341e834729c` Record STORY-261001-qabjyj board state
- `1587aa7a9e13c46103d970b54b1389bcfac28f0f` Record STORY-261001-qabjyj board state
- `5ce440c2abc82ff4ac24d278be54813e6e1c7fc9` Record STORY-261001-3jexth board state
- `2d1f27db7f8424e8c9102dd088b3f42df9dcdf2c` Record STORY-261001-1xuwlu board state
- `1dcdb2cd4c7a02f73e731b1814dc77ba84c4b9a3` Record STORY-261001-1xuwlu board state
- `e412a502880295dcd96f9918b205f6d931295a1e` Record STORY-261001-17ali8 board state
- `f0119a8b56437ef1e5c8ad264afa0445f2c14e45` Record STORY-261001-17ali8 board state
- `56032115ff937922a237fd51a2ff7f6a98b3c232` Record STORY-261001-38ijl9 board state
- `fa8099bd240d6bc622a7d37b4d80c6126890fbec` Record STORY-261001-38ijl9 board state
- `e87d488b8fd892bc86c037ae5e325e596df8e745` Record STORY-260930-klqikd board state
- `eb69624269d9ae7a03dee26dd353959f301045d8` Record STORY-260930-klqikd board state
- `bd126a9acdc51b6061917ba8c4d7d26a7abafd41` Record STORY-260930-3feoy5 board state
- `905e7f484f02cfa8d088571c16ac45bf29ed2b1e` Record STORY-260930-3feoy5 board state
- `54d4aface89e22bf4a4116505eba3ecfe668efa8` Record STORY-260728-1ojb1p board state
- `8c0a763ae79dffc3df096c7eecbfff94c29f56bf` Record STORY-260728-1ojb1p board state
- `8fb36a68e13ba4c0ff1ed62682885de934b7ee60` Record STORY-260930-9k3uil board state
- `3204f51eecaa17bf2c86cca55ba60036a1d867e9` Record STORY-260930-9k3uil board state
- `cd46746a83b29dd46b902acfec9cc23dbae30b10` Record STORY-260728-1ojb1p board state
- `4cad82b156b5012efe91cbe07135770a63c72fb6` Record STORY-261001-54zgrq board state
- `c803afd7e9ed9fe016f85be10fc366cd6dccc0e1` Record STORY-261001-2ub9kw board state
- `bab2433ba115a7eafb2298dba6ef16154e2cc62e` Record STORY-260930-15hioz board state
- `47c83f5ef843949a8ab265530c29101d0e828ffc` Record STORY-260728-19nx3g board state
- `5ed5c4e13671ceafc7992bc4bef074e0e027c72d` Record STORY-260930-15hioz board state
- `5432c85f61aff7776ae69b7215093d3c47562e1b` Record STORY-260917-3w3lvj board state
- `30b3d6781924733ed7120c8adac5183b877f506a` Record STORY-260930-31zrk5 board state
- `5c82acc051011823ee21ad14a0d015432abd984e` Record STORY-260917-hbuawd board state
- `fc2b16bb971c89aef6dfca85aad1b285b9bc18ca` Record STORY-260930-dmzzad board state
- `0d1d78a2ffe683cf235613b686e5552a37870beb` Record STORY-260930-12oimr board state
- `3a43635271f95811474a3e1e0096829c6106c89b` Record STORY-260930-12oimr board state
- `8db5b2354536b6f3c907dcb8c3a889f5d7f2a099` Record STORY-260930-jzq0dx board state
- `a7ea331617a0960ddb5b76f7a05d2b100dd761ef` Record STORY-260930-jzq0dx board state
- `b7740d5eb8f6eea24e859d43b37bb10047251bb0` Record STORY-260928-9vt338 board state
- `eed740494fab07d001ea24d45860f96e1cf717f5` Record STORY-260928-9vt338 board state
- `b4b08a1993d240b5dd24d6929cafe2b51e72637c` Record STORY-260930-2o0ybs board state
- `14710e627a3fd78eaffb7a40d2ed837b20fb7fab` Record STORY-260930-2o0ybs board state
- `92c2d7afd42fce30db71e8443e189a260daf5f2f` Record STORY-260930-1soi12 board state
- `39a2f2fdd120f6fd852f701c0b40d5414ddf3245` Record STORY-260930-1soi12 board state
- `cbe520789de69d01f2018d0d103490bf12b876b7` Record STORY-260720-3plyvy board state
- `38d69d2f47e31b7a3a301fb12b1dff8e67cef93f` Record STORY-260928-3o6kp9 board state
- `f33c2eebc8f07bc22c3b6682592b7e9e7bbbea63` Record STORY-260910-6bo7ej board state
- `0e3169bb6a20d39e58af361a25554b447df8731c` Record STORY-260910-6bo7ej board state
- `cadaaf2e39f6de8005682b3d9b4e8209b6a3d072` Record STORY-260910-148pj1 board state
- `bdb7741353cbb4b75ed3b9e52b001f0660563465` Record STORY-260910-148pj1 board state
- `d50a0bf4797eb73e353407b9334e4a4a8135821f` Record STORY-260929-1s4r14 board state
- `478a7eaf10759ec2d0827d576d694ff4a92cff68` Record STORY-260929-1s4r14 board state
- `d7b85aeb08f1c064cba4eeef1e6dc2615cbb0ed3` Record STORY-260910-234vmx board state
- `be2126819a7490aa788e0b9836e4cbce1d3886e3` Record STORY-260910-234vmx board state
- `f30c2b34229d657062ae03d43f7adb97ecb0e3ea` Record STORY-260928-1t6bto board state
- `85a7133ab7c6b079ec976260f361d65de20df7ef` Record STORY-260928-1t6bto board state
- `64b12189f131d58dada1999e7a930de09e634e8a` Record STORY-260928-rp2r1j board state
- `934093b00b2257aa7aa88d4270f5c10333718618` Record STORY-260928-rp2r1j board state
- `0a6382883557793a23a2719dfc6eedcf6030a39a` Record STORY-260929-1rfp7e board state
- `450861c17b1bb69a1c9bd34a7beb4214fdda35e7` Record STORY-260929-1rfp7e board state
- `fc499a96258f3bd4bdafc581910ad6266be4739a` Record STORY-260928-7eowfl board state
- `10821e673007d4bc4b2eb2cecc6fedbbdb00b162` Record STORY-260928-7eowfl board state
- `cea992e28d5fd8277cc79d40a7e3e8415cf5596d` Record STORY-260929-3f6aym board state
- `ee1e1076334b096899f34d83e2cf1577cfa49c2d` Record STORY-260929-3f6aym board state
- `6a7deb11dc92229c0380d2ea64cf302b86a51383` Record STORY-260929-3f6aym board state
- `7f2fb6b837f9b4d09b7cf7edb865764fdf374adf` Record STORY-260928-lpnvkn board state
- `ee88d1bf38819ea4f17335c7967ffe023950ea2f` Record STORY-260928-3pdcpd board state
- `fa373b53171704edfc86ba3358d6a7dc603cfd6b` Record STORY-260928-3gzr9m board state
- `ec514ce81988ff310441710a89b9fcb55e945293` Record STORY-260928-3fjnux board state
- `8093c6e3af0b6a845876b717bcc3be7a1c592b19` Record STORY-260928-1xu5sf board state
- `8145a8e5c7f4ff491c8dc7b5ec999779d910576b` Record STORY-260928-16hi30 board state
- `397653d74289b7f79db4315b549fe81686dbbdb9` Record STORY-260927-3qf8er board state
- `38fd673e8fbf38c863c0942b6f29b72be9f271a4` Record STORY-260927-22m88w board state
- `0bb3667cccb70140b9b8fba123b22a04421ddd2b` Record STORY-260926-p1s3y7 board state
- `7b31c9a1f46424018ebb0c8b08fc1a83f26d9159` Record STORY-260926-nn2j3l board state
- `ee6c7577cfe80800e5b473eb7d8028065545a607` Record STORY-260926-1y8pjx board state
- `5e17fbd865dc6134e5c6548c8da0598cfc56b866` Record STORY-260924-txgta4 board state
- `92d616465caac257335593c9ee0583aa21f93a92` Record STORY-260924-iafjfs board state
- `9ce95f573700c830875d8442ecf2c95df030005f` Record STORY-260924-7187cn board state
- `b8d57231bf7abb05fccfc03f8686e096175144d2` Record STORY-260924-3gd2d6 board state
- `1a9919beedfb3dbd75016115e83b83fc864367b1` Record STORY-260924-3eywt2 board state
- `77f45611e7ab21e48ee0df0864bd6b6ee69a8478` Record STORY-260924-360z95 board state
- `c4e381f0a2c151699657095a7463e79a1f51d404` Record STORY-260924-2tyzhh board state
- `d6caafa8b09629a86cb38c6de6d63789aba71fd2` Record STORY-260924-2ptqzm board state
- `689409350a503de1c14502c1309c55805b1aa68c` Record STORY-260924-2go2bz board state
- `5437e42d6fe36654f0caa45bea3717e1bfa592eb` Record STORY-260924-2e6xct board state
- `ebd576ac60dae79d1a17a2a57a4412234a38bcd2` Record STORY-260924-1oyh2m board state
- `11c8a81bdcd2fd69c842088f1da1e9ea4963a440` Record STORY-260924-1ckno7 board state
- `35d39a19a942c492ec6bb72f3ef1ab3ef2bd88e2` Record STORY-260924-15f1yi board state
- `2f32308a8ba14cdf823983585f00ff18af87e574` Record STORY-260923-ys7rg9 board state
- `430429b20ab54c508e9e40f254ac50a01170b4c6` Record STORY-260923-vkxt08 board state
- `010c1bf14292294f5c30dd1370e97f1e77ecd350` Record STORY-260923-laeycm board state
- `f6b9f233a7b68f39ec60df6dd765d3d5ef4ce617` Record STORY-260923-3vwgy4 board state
- `0e3907edb2c551e6814a98f5449acf61cf30cb70` Record STORY-260923-3qwrnl board state
- `0fe31cb97a7abc62e5ad747eac93235b5bb99dce` Record STORY-260923-2mla0q board state
- `5460a1f27ef9f4e6a94cabc206f4a09435a1b9d4` Record STORY-260923-2bty9j board state
- `ee393f960f10d02a2565ca7438e2cfca95c61edb` Record STORY-260923-2btaia board state
- `b37533c7528a4714962608a9187b98ee942329fd` Record STORY-260923-1v3no2 board state
- `fb85ef70ab500b0ee91f78b46ddcdef6d91130d2` Record STORY-260923-1lu2o3 board state
- `b1d3b509bbfc7510c68d2dfc138d1b4fb72b8743` Record STORY-260923-11vn9k board state
- `eb2610bc07a01488109249c9d15ea9339a5e389e` Record STORY-260922-h3epwn board state
- `3acbeec33074a0f499903600fbaf3fc9bc3770b2` Record STORY-260922-39hxog board state
- `eb26289c13824dd6b8dd21d69d2485744472344d` Record STORY-260922-2goxjs board state
- `398f897ab3005a0c647a30158e17ffae07943c6d` Record STORY-260922-1cenbr board state
- `75d4ecf99f5ea3e8c99e4413d4406685f38dd885` Record STORY-260922-188t6n board state
- `4ef8bf2c9e1d5ee1813c27d1b8f28ff4d0d869ab` Record STORY-260921-atwkfi board state
- `d3f14c19b815092a91bf83ca0e72dc3a2a725ced` Record STORY-260921-3z0fgr board state
- `50cc1ccc9c9c9188e5fe88ed89db5a6422094b13` Record STORY-260919-37szes board state
- `5ca4f3a895f964e9727b79a79cb462e110e87702` Record STORY-260918-2yvd86 board state
- `1c45eec96196d54dcbb814bef57dee689ec3fd6c` Record STORY-260917-3w3lvj board state
- `acb99355dc9ec1cfe0ea507a6fd58f45194b38b7` Record STORY-260916-wgt8vz board state
- `56ac207ce3f8740626ebf11867197964f726197f` Record STORY-260916-ioemse board state
- `65713b796fa3c69e4ac7b6b085117d8fd84b8a2e` Record STORY-260916-8ql03k board state
- `7fceb41567694defb2ddb58a16370e8762e259a3` Record STORY-260916-73a5zg board state
- `f5b73ad14d9edd24be66d2a24cfb0a19af72ba73` Record STORY-260916-33vuzm board state
- `96a96b4c499468749cedc606e8b27974848c0395` Record STORY-260916-2otjbn board state
- `993ad709d3ccbaa4a5986a9a8e69f6ec11fb432a` Record STORY-260916-2d9coh board state
- `76edc66d2775fd7f7603ef322e7632378f03ac62` Record STORY-260916-1ll22r board state
- `097d8ff49f4b0a7595aeec9181780d373feafb44` Record STORY-260916-1i1gfo board state
- `6125033faa2dedee0823f4f313398f8b046d041a` Record STORY-260915-3w11un board state
- `adaade4243c2e93d67a94c2121db6853064ec5cf` Record STORY-260910-2qmrb8 board state
- `6df669d951662f78b8dbc7f5a58f8f455f16bbc0` Record STORY-260910-25yc0h board state
- `337a66042a242861ba5093b7d48871857c0612bf` Record STORY-260910-1lf0m5 board state
- `93524468cf8b6f0d10a7a46064fe9d37f2658514` Record STORY-260908-g7o5zw board state
- `3fb81e1a6f88b1cf32cc1e6a3d1fef90461dd3fb` Record STORY-260907-2bddfc board state
- `db19d4c326fe1dde55589e3cbe24c77505e7e600` Record STORY-260906-1a2i5a board state
- `fee552bf9fcfca4a0ccfcfa87146a616771398c9` Record STORY-260905-3l0fav board state
- `12c09c50d98c285a2ccac2616826a8fbd4eaef4a` Record STORY-260905-2z9pw4 board state
- `224c187f723517a99373d9dea6f2818600f5ff22` Record STORY-260905-2qvzwk board state
- `9a244c45fd312768177606e2dcb4887485e14232` Record STORY-260905-1n0iy8 board state
- `805b4c020016f426c449b3ec262084cf330d22aa` Record STORY-260822-2h0v9j board state
- `a6fd75b7d02bc33282ccc7185641ba4a25e526d1` Record EPIC-260910-16qce1 board state
- `ba51e4aa20657916987d1a0511dc2bbf1fae1406` Record STORY-260928-3gzr9m board state
- `213a53e5c701ef16b961f3398c7f2f0b82c98094` Record STORY-260928-lpnvkn board state
- `56392800b8adc9ae2ffa49d5ae01101241fe6fe9` Record STORY-260928-1xu5sf board state
- `d8e87bacda3bb4cd9010801156646f75de9354ce` Record STORY-260916-wgt8vz board state
- `2252ebee05b5d93c982b288aa6de57dcc4dea6a9` Record STORY-260923-11vn9k board state
- `e4f4fe8653df21500752c8c06baf0d0114021fab` Record STORY-260916-73a5zg board state
- `97e856425b1aeafe533e86e33e7f9dfd873d9b50` Record STORY-260916-ioemse board state
- `86552087f9a5ba27a9148ee84f579cc8b0b1fcbe` Record STORY-260916-1ll22r board state
- `6bd98d49e9aaa65db0937ef0112af0351eddbf08` Record STORY-260916-1i1gfo board state
- `aa7d8d0904499be242c44d6d000889bf63975bbd` Record STORY-260916-2otjbn board state
- `eca2bf27edaec03227ce8485953d96df9d7ef48c` Record STORY-260910-25yc0h board state
- `d41da0fbf210f6bc0ea1a5cef3d81a12af9eb439` Record STORY-260916-2d9coh board state
- `38c68570b10f7772dd9e7c4adbe10644ba4c6805` Record STORY-260927-22m88w board state
- `55b94af251d72fe78637e1af9b51fbc807e7ed67` Record STORY-260910-1lf0m5 board state
- `97ca33704e7f2bca27e457ca78323c6faae97034` Record STORY-260916-8ql03k board state
- `0ffe2e1db9400dd18de3c3a0241275295c481b0e` Record STORY-260924-3gd2d6 board state
- `0be1c20e41bb3d9aec268aeea358a53eed7ff79c` Record STORY-260923-1lu2o3 board state

## Task logbook (repository LOGBOOK.md untouched)

- Review regression: revision 1 treated tag Unreleased presence as released-note duplication. That disposition and its evidence predicate were invalid; corrected in results, JSON and the shared validation gate.
- Decision: preserve all 27 original texts, including interim stage language, under an explicit rc.2 attribution. Do not merge them into new rc.3 feature claims or alter earlier release sections.
- Decision: keep the five original-brief consolidations and the truthful rc.14/v1-writer statement; no blanket restoration of the stale rc.13 pin wording.
- Verification: the named preservation regression detects every individual omission; a narrowed tag-wide predicate fails it. The public-wording regression also fails a scan narrowed to exclude historical text.
- Handoff: findings and test artifacts are attached on the task before developer handoff; no changes are committed here.
