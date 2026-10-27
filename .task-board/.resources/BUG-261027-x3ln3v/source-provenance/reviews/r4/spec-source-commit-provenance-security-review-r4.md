**Source commit provenance — independent security rereview r4**

**ACCEPT_CODE_PENDING_HOSTED_GATES** at exact [curator-spec PR202](https://github.com/relux-works/curator-spec/pull/202) head `43d5831adb77f58e756cbdfa7eabffc1454922e8`. R3-01 is closed by the bounded fixture-error correction. No new finding remains after the complete changed-helper/test and cumulative security/consumer/policy sweep. This is code acceptance with actual Windows and all required hosted checks outstanding, not permission to land, adopt trust, release or claim main recovery.

| Source/publication boundary | Independently established evidence |
| --- | --- |
| Exact candidate | `43d5831adb77f58e756cbdfa7eabffc1454922e8` |
| Predecessor and retained intermediate | `357f2724c9ecc22db210dfa6d43abdb451ccfa44`, then `81d111a424d9bb1143df346bc7ba9162a9c34f7c`, then candidate; no rewritten predecessor |
| Base and fresh current main | `74a7dcc503c0fd7acd491299ab71290ae40e708e`; current main is an ancestor of the exact candidate |
| Canonical repository/default/feature | `relux-works/curator-spec`, `main`, `fix/source-commit-provenance`; remote configuration, default/PR APIs, advertisements, fetched refs and exact detached checkout agree. Final independent advertisement and PR source-metadata read still agree; PR remains open and same-repository |
| Signed objects | All six cumulative task objects independently verify local `G`, the same authorized source fingerprint, configured human author/committer, exact sole `Co-authored-by: Codex gpt-6.1-sol <noreply@openai.com>`, and no session metadata. GitHub commit API independently reports `verified: true`, `reason: valid` for all six; messages/parent identities match local objects |
| Complete remote/local parity | All 12 complete remote patches reconstruct byte-identical candidate blobs from base; cumulative published delta 1,125 additions/97 deletions. Complete predecessor-to-candidate correction is one file, `tools/test_verify_release_commit.py`, 245 additions/1 deletion |
| Full diff pins | Corrective diff SHA-256 `129e2188b836cdd8d538f222261dd10966cc84d39fa5678e52abea1e727f64f0`; full local base delta SHA-256 `80f767a08430adf61705392c09fd3b1de4a2677294d5972d825b3667944b7dfe` |
| Input review | Complete new producer handover: 14,986 bytes, SHA-256 `a1873a6995d40c836d36a95c39b92422207c027d38076c1b44ade3625ec60133`. Complete prior r3 report: 28,878 bytes, SHA-256 `195c72178ddef191dcf4bf0b73ffa684efd8ca896cc4476b48396912515f54c4`. Previous r3 brief and full original r1/r2/r3 scope retained. Producer COMMENT 5675005401 read as evidence, not acceptance |
| Isolation/preservation | Private Git directory, work tree and explicit command core.worktree binding were established before materialization. Clean exact detached checkout; stable HEAD/branch/status/index/both diffs and 11,028 pre-existing file hashes preserved, including all earlier reports/evidence/worktrees. No common core.worktree edit, source edit or broad reset/clean |

The six signed objects, in order, are `ec70b25b5c7b7b23be64a05dacd6534eee71ade3`, `41a6121151eea2cd4229f78fb2b4377f28370ef1`, `3d5f838717e0609fa8905c37b13b117c8d974a00`, `357f2724c9ecc22db210dfa6d43abdb451ccfa44`, `81d111a424d9bb1143df346bc7ba9162a9c34f7c`, and `43d5831adb77f58e756cbdfa7eabffc1454922e8`. Their common fingerprint is `SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA`. The source authority comes from the existing explicit owner mandate and independently matched public material/landed signature retained from r1, not from the producer adding an allowed-signers entry. Signature validity alone does not establish review, authorization or protection compliance.

| Finding | Severity, requirement, reproduction/effect and disposition |
| --- | --- |
| R3-01 | Medium; safe expected fixture-error diagnostics and exception state. The old helper exposed synthetic argv in a real timeout traceback and retained captured streams; its sanitized launch failure retained the original OS error through implicit context. **CLOSED at this exact head.** Expected OS launch and communication timeout errors now retain only safe classifications/count availability, leave their handlers, and raise replacement errors after bounded exact-child cleanup where applicable. Independent exact-old-byte controls still fail the intended confidentiality assertion, with no import/setup error. Nearby real exit-0 and check=False exit-7 output succeed; corrected real timeout reaches its deadline and reaps the owned direct child |
| R1-01 | Medium; production API error/token confidentiality. **CLOSED, preserved.** Both production helpers and their r2 transport controls are unchanged. Current named tests include real malformed-header rejection before connect, fixed helper/public diagnostics, known error families, valid in-memory HTTP parsing and redirect refusal. Earlier exact-old-byte failures and reject-all detectors remain pinned |
| R1-02 | Medium; unique source-policy decision identity. **CLOSED, preserved.** Decision0034 filename/heading/four policy references and PR text remain coherent. Fresh main and every other open PR inventory show no separate 0034. PR200 retains its own distinct design-only Decision0033 at `0d3bd4e8d7c33e6ba95de30673205936bf6c4e34`; no foreign source or semantics changed |

The earlier [r3 CHANGES_REQUESTED](https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5674843768) and [r2 code acceptance](https://github.com/relux-works/curator-spec/pull/202#pullrequestreview-5674261677) remain historical, exact-head records. This review neither alters their verdicts nor converts their intermediate red evidence into passing evidence.

| Complete changed surface | Review result, independent control and limit |
| --- | --- |
| Executable resolution and launch | Existing explicit shutil.which resolution, argv sequence, owned cwd/env and PIPE setup remain. Missing executable refuses without dispatch. Expected OSError stores only its class name; the fixed AssertionError is raised after leaving the handler. Safe tool-family/name labels do not identify a historical binary version |
| Launch exception confidentiality | Controlled sensitive PermissionError and FileNotFoundError inputs, plus an OSError constructor that resolves to PermissionError for errno 13, pass. Checks inspect str/repr, args, __dict__, formatted traceback, cmd/output/stdout/stderr, filename/filename2 and cause/context. Exact old helper fails the intended secret-absence assertion because the original launch error remains chained |
| Communication timeout | communicate(input, timeout=10) is retained. Timeout handler reads only availability and length/type of its standard partial output fields; it stores no original exception for later diagnostics. Real child-ready sentinel, observed actual deadline and exact child process state precede confidentiality acceptance |
| Availability versus amount | None is reported as unavailable, not zero or guessed emitted data. Empty bytes report zero observed bytes; bytes and text have distinct units. Independent audit distinguishes a two-byte UTF-8 sequence from one text character. Current real local timeout independently observes both API fields as 49 bytes; that result is local, not a Windows assumption |
| Exact-child cleanup success | Only the owned Popen child is killed; wait(timeout=5) returns the reported actual status. Real corrected child reached its body and 10.007-second deadline, then returned -9 and was reaped. Successful cleanup is established before the replacement AssertionError. No process-pattern kill or process-tree guarantee |
| Cleanup failure | Expected kill OSError, wait OSError and wait TimeoutExpired retain only stage/class and produce cleanup=unconfirmed. They claim neither an exit nor successful reaping. Independent controlled states verify kill-once, wait-once with five seconds, or no wait after kill failure. These are call-state tests; they do not prove OS cleanup under an actual denied kill or wait failure |
| Interruptions | Existing BaseException cleanup/propagation branch is unchanged. Controlled KeyboardInterrupt and SystemExit preserve the identical exception after one communicate(timeout=10), kill and wait(timeout=5). This proves the tested successful-cleanup call state, not arbitrary interruption/cleanup-failure combinations |
| Ordinary child results | Real exit 0 preserves stdout; check=False exit 7 preserves explicit nonzero output. Ordinary checked nonzero and silent failure still report only family/name/status/stream lengths. Missing/unlaunchable tools refuse. No reject-all implementation could satisfy these nearby valid controls |
| New capture_failure/assert_safe_failure helpers | Sensitive material is examined only in memory; saved observations contain booleans, classes, status and lengths. Assertions occur after the exception handler. Both exact-old-byte failing assertions have no attached raw cause/context, so the detector does not disclose the exception it detects |
| Four new tests | Complete launch-state, real-timeout/valid-child, cleanup-fault/partial-shape and interruption test methods were read and executed. Their real-process versus controlled-state claims match the assertions. No skip, platform exclusion or mock-only replacement of the existing cryptographic controls |
| Existing 39 tests and setup | AST equality after removing only modified SourceTrustTests.run_child and the newly added ExpectedFixtureErrorTests class proves the remaining module unchanged. Three real Ed25519 fixture keys, original signature/ref/trust/API/containment/tag checks and r3 environment checks remain intact; all 43 current named tests pass |
| Confidentiality scope | Expected standard-library launch/timeout families and ordinary failure receipts are covered. Traceback frame locals, arbitrary hostile library behavior, whole-process memory secrecy, unbounded process trees and a universal hard wall-clock guarantee are not claimed. Successful check=False intentionally returns requested output |

The exact old module from `357f2724c9ecc22db210dfa6d43abdb451ccfa44` was loaded byte-for-byte and its unchanged run_child method substituted into the current final launch/real-timeout detectors. Both produce the precise expected assertion failure, zero errors/skips. For the old real timeout, the child-ready sentinel, 10.008-second observed deadline, returned -9 and successful exact-PID reaping were established before the failed confidentiality assertion. Old timeout data had unsafe diagnostics/raw fields; old launch data had unsafe chained context. The corrected version produces safe diagnostics and no raw fields/cause/context. No generated sensitive fragment, raw failed traceback, private key or credential was persisted.

The Windows nuance is supported by an independent read of [CPython 3.12.10 subprocess.py](https://github.com/python/cpython/blob/v3.12.10/Lib/subprocess.py#L1603), SHA-256 `85d29b2bf0249f5436838298c9a60ee93508b1102e9ac43b001f8a7e7ae8f375`. The Windows reader-thread path can raise TimeoutExpired before collected output is assigned, leaving partial fields unavailable even after a child flushed. The fixture now matches observed API availability. This source read and controlled None/empty/bytes/text states do not constitute native Windows execution or identify the exact hosted Python patch version.

| Preserved cumulative security/consumer/policy surface | Evidence and practical boundary |
| --- | --- |
| Closed source trust parser | Exactly one source-commit principal, literal git namespace and valid Ed25519 base64/wire material. Missing/nonregular, malformed/duplicate, wrong principal/namespace/key type/material and unsupported inputs refuse. Current tests and pinned r1 regularity/symlink/FIFO/material cases apply to byte-identical production code |
| Maintainer source route | Original maintainer route remains available alongside source-only trust. Real synthetic source and release-key source commits succeed; no original private release key is used as fixture input |
| Every source route requires containment | Exact commit resolution and fixed origin/main ancestry precede both SSH and GitHub-created acceptance. Current unknown/unsigned/uncontained/missing-ref/wrong-target negatives pass. Pinned r1 query-error, replacement and PGP-route containment controls remain applicable; no early-trust success bypass |
| Source Git environment | Full inherited GIT_/XDG_ selector removal plus SSH socket/askpass removal; deliberate no-system/no-global/no-replacement policy, fixed cwd/trust inputs/SSH format/program. Current synthetic poison tests preserve full decoy bytes and metadata, absent external trace and caller env. Trusted installed Git/SSH executables and repository metadata remain assumptions; no arbitrary same-user attacker sandbox claim |
| Fixed reference/API authority | Full resolved SHA, fixed repository/default branch/API endpoint and request authority; conflicting context, unsupported caller override flags, wrong target/ref/query/error/unknown results refuse. Containment relies on the workflow's fresh origin/main fetch, not an authenticated claim from an arbitrary local ref |
| GitHub-created PGP fallback | Actual local signed-header dispatch, exact object identity, GitHub committer, verified true/reason valid and nonempty string signature/payload are required. Delegated GitHub validation is not local PGP verification, arbitrary user-signature authority, proof of PR review/squash topology/protection or current custody. API simulations demonstrate state/transport behavior only |
| API confidentiality/transport | Unchanged fixed errors raised outside handlers after OS/URL/HTTPException/ValueError handling. Current real newline-header rejection occurs before connect; known-family details and cause/context disappear. Fixed endpoint, 30-second timeout, actual in-memory HTTP valid responses and redirect refusals pass. No live token, request dump or authenticated mutation in fixtures |
| Optional merge-settings observer | Unchanged fixed-repository strict observer and context refusals; observation only. It neither changes settings nor compels rewriting an already signed fast-forward head; protection/review approval remain separate gates |
| CI consumer | Same main-only Release target provenance job/name/trigger; fresh default-ref fetch and exact GITHUB_SHA source verification. PR-only skip remains and is not qualification or recovery |
| Release/tag consumer | Annotated tag signature uses ONLY original maintainers.allowed_signers, with exact peeled tag/candidate agreement, fresh main fetch, source verification and all prior release gates. Current real source-key tag rejection is paired with release-key endorsement success; wrong-target agreement is rejected |
| Namespace separation | OpenSSH git namespace covers both commit and tag operations. The distinct consumer trust roots, not a fictitious commit-versus-tag namespace distinction, enforce source-only authorization |
| Packaging/every verifier consumer | Both separately scoped trust inputs remain packaged. The full original tracked/hidden consumer sweep and all touched adjacent reads remain applicable because only the fixture file changes. No source-key tag consumer, new packaging authority or archive/full-workflow execution claim |
| release_gate and adjacent tests | Exact unchanged release_gate.py, test_release_gate.py and test_allowed_signers.py were fully read in the original sweep; fresh byte pins preserve those results. Version/clean exact checkout/immutable historical source pins/independent review/ancestry/frozen diff/platform-assurance limits remain. Signature verification is a separate workflow gate. No release_gate test execution is claimed here |
| Governance/Release/Security/README/Decision | Explicit amendment of old squash-only procedure permits reviewed exact signed fast-forward and preserves the GitHub-created signed alternative. Separate tag endorsement, owner trust-policy decision and independent review remain required. No retrospective compliance fiction; verifier success alone does not prove human authorization/review/checks/protection |
| Decision allocation/publication | Current main and open PRs 203, 200, 198, 119 and 89 contain no other Decision0034. PR200 source/API head still match its original 0033 lock-refresh proposal. All four source-policy references, decision heading and current PR publication are consistent; historical 0033 reviews remain historical |
| Other authority | No existing key revocation/rotation, release/tag/platform-key/infra-exit/protection/status/version/window change. No branch rewrite, landing, adoption, release or native/host operation |

| Preserved r3 fixture/startup boundary | Evidence and limitation |
| --- | --- |
| Environment | Same explicit allowlist constructor and setup: owned absolute ProgramData/HOME/temp/cache/GOPATH, no broad live environment copy; caller GIT_/XDG_/SSH/askpass/native/user-profile state excluded before synthetic values. Three real observed keygens have empty passphrases and owned ProgramData/nonexistent socket |
| SYSTEMROOT and PATH | SYSTEMROOT is the deliberately retained Windows loader field; PATH selects trusted installed tools. Resolver PATH/PATHEXT/current-directory context and native tool roots remain trusted assumptions. Explicit argv/no shell does not claim arbitrary malicious PATH or Windows batch-file safety. Exact historical executable/version remains unknown |
| Native agent distinction | Pinned Microsoft startup only invents its default agent pipe when SSH_AUTH_SOCK is absent; synthetic signing receives an owned nonexistent socket. Product verification removes the socket and verifies signatures; signing and verification do not have identical agent behavior |
| Pinned startup source | Previously independently downloaded/matched Microsoft commit `274893f8d1fe86c4d128b514c3af4bc10cdc0c99` remains applicable because fixture_environment/setup are unchanged. [wmain_common.c](https://github.com/PowerShell/openssh-portable/blob/274893f8d1fe86c4d128b514c3af4bc10cdc0c99/contrib/win32/win32compat/wmain_common.c#L40), SHA-256 `248603baffe3e0403164176997a11f49c72fc66d9c1690df4f8c0ce584aab66f`; [w32fd.c](https://github.com/PowerShell/openssh-portable/blob/274893f8d1fe86c4d128b514c3af4bc10cdc0c99/contrib/win32/win32compat/w32fd.c#L195), `a964fb8994b4b778fe51db359fa72ba04e4c0cead122156c21f110958f899186`; [ssh-keygen.c](https://github.com/PowerShell/openssh-portable/blob/274893f8d1fe86c4d128b514c3af4bc10cdc0c99/ssh-keygen.c#L2716), `895098057c6360f533eb307c1a1552ddb06005d1647c7ff42c34e382bbcab716` |
| Old startup detector reuse | Exact old setup reached two real successful Git initializations before the intended missing-ProgramData mapped assertion. This is source-mapping sensitivity, not execution of an old Windows binary. Corrected real keygen/signature controls still pass; unchanged mapping need not be rerun |
| Historical Windows failure | Prior [Windows job](https://github.com/relux-works/curator-spec/actions/runs/40580017534/job/122457223176) failed setUpClass keygen with empty stderr. Supplied official failed-step output SHA-256 `8abd43a664bfe9a8ce5828ee280d13a515be528ec1b34a33866a670c38117d40`, 8,941 bytes. Binary/version, exact child return code and stdout remain UNKNOWN. A plausible ProgramData failure class is not proof of that exact historical cause or an external outage |

Independent commands used an empty-based allowlisted child environment, absolute disposable HOME/cwd/TMPDIR/TMP/TEMP/cache/XDG/GOPATH and no inherited Git/SSH/native selectors before deliberate synthetic settings. Installed Apple Git was directly selected to avoid the already documented shim-control failure; no toolchain/profile installation or repair occurred. Exact local argv/environment/PID receipts are retained privately; public command paths below are abbreviated.

| Independent command/control | Exit/result and detector meaning |
| --- | --- |
| `python3 -B <evidence>/run_controls.py named` | 0: exactly test_verify_release_commit, test_verify_release_merge_policy, test_allowed_signers; 43 tests, 0 failures, 0 errors, 0 skips. Three newly generated real fixture keys and real signed commit/tag operations; 14 full decoy/caller snapshots agree |
| `python3 -B <evidence>/run_controls.py old` | 0 detector: exactly two intended secret-absence assertion FAILs from exact predecessor run_child, zero errors/skips. Old real timeout reached body/deadline and exact child was reaped. Checker failures themselves have no raw context. Exit 0 means successful detection, not acceptance of old confidentiality |
| `python3 -B <evidence>/run_controls.py audit` | 0: eight independent controlled input/state rows, covering launch filename/context secrecy, honest unavailable/empty/bytes/text counts, successful reap status, kill OSError, reap OSError and reap timeout. Three launch inputs represent two actual exception classes because OSError(errno13) constructs PermissionError. Controlled cleanup is not OS-fault proof |
| `python3 -B <evidence>/verify_closure.py` | 0: all 12 remote patches reconstruct candidate; unchanged remainder AST; fresh decision allocation; stable index/checkout and 11,028 prior file hashes preserved; exact detached checkout clean; diff --check passes |
| Pinned r1 instrument evidence | Exact original verifier SHA-256 `423596a105e6dc2b12c44c9d804c2d4c48017cb16fc517cd23b838509f498ee7` fails intended early-containment and ambient-trust assertions. Accept-all and reject-all plants fail paired refusal/positive properties. No import/setup failure counted as detection |
| Pinned r2 instrument evidence | Original API bytes fail ten helper and ten stderr confidentiality assertions; four reject-all substitutions fail valid API controls; twelve independent API fault pairs pass. Relevant production bytes/current transport assertions unchanged; no need to rerun unrelated old plants |
| Pinned r3 instrument evidence | Startup-mapping and ordinary nonzero diagnostic old-byte controls remain applicable to unchanged setup/environment/current retained ordinary-output tests. R3-01 old red is additionally reproduced fresh above and remains preserved as historical evidence |
| Snapshot/state preservation | All 14 fresh full synthetic decoy snapshots include root/file metadata and bytes; caller environment agrees before/after. Atime is deliberately excluded because reading affects it. Trace absence and exact old resource/report preservation are checked separately |
| Direct processes | 95 independently recorded unique direct fixture children/runners, all absent at cleanup. Both real timeout children were killed/reaped by exact owned PID. This count is independent of producer and previous reviewer counts. Process lists/argv from other sessions were not inspected |
| Disk | 25,174,016-byte sampled task peak plus 458,752 bytes conservatively counted recent common-Git allocation = 25,632,768 bytes. With 1,048,576-byte publication/report reserve: 26,681,344 bytes, below 48 MiB STOP and 64 MiB hard budget. Final evidence growth is checked again; 50ms sampling is not a continuous filesystem quota claim |
| Bounds/cleanup | 1 MiB per-file OS limit, 60-second outer runner bound; candidate communicate 10 seconds and wait 5 seconds retained. Owned synthetic keys/repos/home/cache/runtime roots removed; clean review checkout and small text/source evidence retained. No live key/socket/repo/index/home/config used as a fixture input |

All earlier setup/instrument reds and producer intermediate evidence are preserved, including the original verifier/import/shim and patch-reconstruction issues already classified in prior reports. No failed setup is retrospectively a passing security control. Reused receipts are freshly hash-checked in prior-control-pins.json, SHA-256 `49da6765c55d1afc97d288404084ef3b473abfdd43b9f0505e85831eff52582d`; this records exact r1/r2/r3 evidence and why unchanged contracts permit reuse.

The selected recorded managed-profile/model/effort/resume fields matched the requested original reviewer context. No fallback, complete native configuration/argv inspection, new worker, native account/provider operation or host change occurred. Local execution was Darwin / Python 3.15.1 / Apple Git 2.54.0. No full test discovery, Go/product/module/build/package/race/repro/make/validate/regenerate/workflow/full gate/broad lint/crosscompile was run. Local results do not replace hosted Python 3.12 Windows or other required platform jobs.

The changed fixture module's predecessor blob is `5f437226b8d6f736bf280b7dd33ac8a8755c564c`, SHA-256 `2e06b226de3d37f078b476bc4075463892f7084ab1a9c5b9cf25e0a37c9f74b5`; candidate blob `0930f938b0a105f818859c7093a2bbf654cc62c4`, SHA-256 `464fd7b99a4b214115fc7b1beacfea6654ebe80e06ff9a14c776b384a9fe627a`. Every following input is freshly byte-identical to the r3 predecessor, preserving the accepted production r2 contracts and original release boundary:

| Input | Git blob | SHA-256 |
| --- | --- | --- |
| `tools/verify_release_commit.py` | `6de6aba39f78b600387afc4700b24c92f6995118` | `ad24bff5b90da99be0f977e3cdf57a14010389184c77bd13f6f50185262d1961` |
| `tools/verify_release_merge_policy.py` | `1cd3f5e64b1a40c462cb210e5deb16595ac84eb1` | `0f5617668e3374c3e3e1151310673de8ba9874b44b563b7f6e1f233af5e1dbcb` |
| `source.allowed_signers` | `8e3976ba08be6a9c0386193f32cf73aa00b3b397` | `a757c518fb974409858ead6c177bb332f08877d13e96420196eaf045d0f9927f` |
| `maintainers.allowed_signers` | `47f54d69ce219fa871602d77c554282c7d2c6207` | `12f4bfaf1bfcd954d30b0683f6b55b8ad7e19e5f60c88c95dbf0af702db82dc8` |
| `.github/workflows/ci.yml` | `78113b3efcfc056b610d16434223c951216ee672` | `e8e3a8d1133a0ac6a8a1110bec77264c76659e64a3dd632be9d495acf161b0c7` |
| `.github/workflows/release.yml` | `81701d420fb09f3a1ed9c652e53d70dc260685ed` | `260c8e03830c10d838b203157aa8e72d8f5b611e59c893e33f12caabc4496bbf` |
| `tools/release_gate.py` | `db0104261a5ba7028709946f60a66f9912aeff99` | `b6342d54a1f80e68f93206b4e97690157a999bc6eb37a43d6e4c52ef03a80dcc` |
| `tools/test_release_gate.py` | `5ce14d159ab8a6bd011ccffb54aeec1587fede0d` | `dfc6e27ec06bce182efee4e7ca709271ad3821777c18846e43edceae1b2170ab` |
| `tools/test_allowed_signers.py` | `3ad66a6a79971096b352c0a9cab72e5c7532d43f` | `5ab5484447c3aee4974c2aaf63a07ca0f12d42eb14260251a5dab1cf4e8415d5` |
| `tools/test_verify_release_merge_policy.py` | `e7ec3371e7af706fd3640256da871503b9a38edd` | `177e66eabc38807968853d43280dd48f3fd7c4b4a1b4f83bfc061b4f1ff4407b` |
| `decisions/0034-source-commit-and-release-tag-trust.md` | `5d3ad36d529630950aada62a56580f45e24dba3e` | `cfb09a7ce0704d7e3950d7b6c81a81fdda6f95994b508c822dc653ab575a7bae` |
| `GOVERNANCE.md` | `6e641811a6a5008728a1ba89156bf11e323c9c1b` | `1df692be2268871a75ebf7a3afaf5b744d4858f0a8123241ed6f8856e8488bec` |
| `RELEASE.md` | `3842918862e15732eb391fdb8f76833973c6c2ef` | `d428607b728757e79a0e7f6431e9eee22a993d9abed1e70e91dc56fde749a00d` |
| `README.md` | `cbcdaa06ccfb1c2252b93d9a959128fadb254d44` | `4255d1a0dabf9fffc83ce36d3274f8bcbe82a7a1bf75308b1d94a31129751362` |
| `SECURITY.md` | `c37572718da7ac8f32aedeb6b1971fa231b7e5c0` | `70eecfa1447a42beea58ef3dc4b210a270d1282043a23d0adbef3ddb5910457a` |

The original maintainer trust is also byte-identical to base and the original source candidate, not merely to the last fixture predecessor. Its source route remains valid while release tags continue to require that original root exclusively.

Exactly one hosted metadata snapshot was read at this head: [Specification run 40591288612](https://github.com/relux-works/curator-spec/actions/runs/40591288612) had Ubuntu/macOS/Windows Specification IN_PROGRESS, Formatting/Links SUCCESS and Release target provenance SKIPPED. [Implementation run 40591288157](https://github.com/relux-works/curator-spec/actions/runs/40591288157) showed all three platforms SUCCESS. These are the observed snapshot states, not a claim about subsequent completion. No CI polling/wait/rerun/cancel/log or artifact collector/download occurred. Native Windows execution remains UNVERIFIED in this review, and actual new Windows plus every required hosted check remain gates.

Historical actual-main [run 40569031707/job 122417945263](https://github.com/relux-works/curator-spec/actions/runs/40569031707/job/122417945263) remains genuinely failed. PR provenance SKIP is not qualification or recovery. Only after separately accepted exact-object landing can a NEW actual-main provenance pass establish recovery. No old-main status, prior review, task or goal is closed or rewritten.

```json
{
  "verdict": "ACCEPT_CODE_PENDING_HOSTED_GATES",
  "head": "43d5831adb77f58e756cbdfa7eabffc1454922e8",
  "base": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "current_main": "74a7dcc503c0fd7acd491299ab71290ae40e708e",
  "findings": [
    {
      "id": "R3-01",
      "severity": "medium",
      "status": "closed",
      "requirement": "Safe expected fixture errors without sensitive exception state",
      "evidence": "Corrected real timeout and launch state pass; exact old bytes fail both intended confidentiality assertions after valid child/deadline/cleanup controls; cleanup failures remain unconfirmed"
    },
    {
      "id": "R1-01",
      "severity": "medium",
      "status": "closed_preserved",
      "requirement": "Secret-free production API errors",
      "evidence": "Byte-identical production helpers; current named transport cases and pinned old-byte/reject-all controls"
    },
    {
      "id": "R1-02",
      "severity": "medium",
      "status": "closed_preserved",
      "requirement": "Unique source-policy decision number",
      "evidence": "Decision0034 and references preserved; fresh main/other open-PR inventory clear; PR200 Decision0033 unchanged"
    }
  ],
  "source_trust": {
    "fingerprint": "SHA256:rE/sRlfdFt/vT5y5yukDCsq4aFtC48pbWsq9z4hm8MA",
    "scope": "Commit-only plus existing maintainer source route; default-branch containment before every success",
    "adopted": false
  },
  "release_trust_unchanged": true,
  "native_or_hosted_gates": {
    "focused_tests": "43 pass, zero failures/errors/skips",
    "old_byte_detector": "two intended assertion failures, no setup/import errors",
    "native_windows": "UNVERIFIED; source mapping and Darwin controls do not substitute",
    "historical_windows_binary_status_stdout": "unknown",
    "specification_jobs": "all three IN_PROGRESS at sole exact-head snapshot",
    "pr_provenance": "SKIPPED; not qualification or recovery",
    "actual_main_recovery": "not established; historical failure remains red"
  },
  "next_checkpoint": "Root separately verifies real Windows and every required hosted check for this exact signed head, then applies the existing review/protection/exact-object landing gates. NEW actual-main provenance must pass after any separately accepted landing before recovery is claimed. No release, adoption or task/goal closure is authorized by this review."
}
```

This immutable report is the body for an exact-head COMMENT review, not a fabricated platform approval. Separate publication and exact body readback receipts, distinct sanitized review outcome and append-only note receipts record delivery while preserving all prior resource bytes/declarations/notes/status/assignee.
