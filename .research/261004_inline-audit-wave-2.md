# Curator robustness review, wave 2: confirmed findings N6–N9

Report date: 2026-10-04. Task: **TASK-261004-2wvbzz — inline-audit-wave-2-uncovered-surfaces**. Research handoff: **ready for review**.

Four invariant failures were reproduced on temporary local fixtures. This extends the [2026-10-03 report](../docs/security-audit-2026-10-inline.md); it is targeted evidence, not a certification of the repository. No product code was changed.

## Identity, method, and boundaries

- Curator: `e5489b6ba22c9e9cf7a925e03f3653d115188999`.
- Separate launcher repository: `d0920353556dd0a3cea2864c616c230915977bc7`. Its results are attributed separately; curator-run does not ship in Curator.
- Specification: Curator's actual CI pin, rc.14 commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`. The pin creation-time requirement was also checked in rc.13. The newer hash-version migration is outside this pass; N5 is not claimed as a new finding.
- Both source trees were copied using `git archive`, with new probe tests added only to scratch copies. Byte comparison covered **9,641 Curator tracked files and 153 launcher tracked files: zero differences**. The worktree deliverable is this report; probes and evidence are board resources. No commits, branch operations, installations, or product fixes were performed.
- Reported toolchain: `go version go1.26.0 darwin/amd64`. All Go validation processes used `env -u GOROOT GOFLAGS=-work GOMAXPROCS=2 go test -p 2`, `-count=1`, and bounded package/test selections. No gate used a pipe or `tee`; shell redirection retained the gate's exit status.
- Fixtures use temporary manager homes, real local Git repositories, synthetic credentials, a loopback registry, and local recording child processes. No external registry fetch or real credential store was used. Public logs redact temporary/personal paths; credential assertions log lengths, never secret values.
- Work was performed inline without sub-agents. Every cited test was run in this session; no prior attached test result was accepted as a substitute.

The decision this report supports is which production boundaries need regression fixes. N6 and N7 deserve functional remediation first; N8 and N9 need pin-writer validation and provenance work. The attached probes are the first regression slice for those changes. No further research prerequisite is proposed.

## N6. Oversized credential output is silently truncated and accepted

**Confirmed. Medium functional risk: a helper response can become a different credential while being reported present.**

Promised invariant: the helper's response has a 64-KiB bound, and an accepted credential is the helper's actual answer. The production call is `gitcred.Access.ReadHost` → `Access.call` → the real Git credential process. The [bounded buffer](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/gitcred/gitcred.go#L402) keeps only the remaining prefix but returns the original write length with no error. [Access.call](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/gitcred/gitcred.go#L258) therefore treats a successful Git exit as a valid response and parses the prefix.

`TestWave2CredentialAnswerBound` uses real Git, an isolated `.gitconfig`, and a local helper returning synthetic repeated bytes. `TestWave2CredentialExactFrameBound` additionally addresses exact frame lengths through `ReadHost` using a controlled Git executable, explicitly separating the framing boundary from Git's output formatting.

| Probe | Observed result | Expected assertion |
|---|---|---|
| Real Git, secret length 32 | Present, exact bytes retained | Pass |
| Real Git, secret length 65,456 | Present, exact bytes retained | Pass |
| Real Git, secret length 65,664 | **Present, returned length 65,472; bytes truncated** | Fail |
| Controlled output, total frame 65,535 / 65,536 bytes | Both accepted unchanged | Pass / pass |
| Controlled output, total frame 65,537 bytes | **Accepted after output truncation** | Fail |

Go process exit: **1**, commands 01 and 07. The failure is expected-red evidence, not a passing gate. The real Git child succeeded; the library exposes presence rather than its child status. Four adjacent prompting, namespace, near-miss, and failed-persistence controls passed in command 10, exit **0**.

This is not an observed memory-exhaustion defect: retained output is bounded. No authentication using the truncated value was attempted. A fix should retain an overflow flag, refuse the whole answer on overflow, and preserve acceptance at the exact limit.

## N7. A recognized manager-owned forwarding link cannot be reconciled

**Confirmed on Unix. Medium compatibility risk: a global install fails on an owned legacy shim.**

Promised invariant: both current regular forwarding launchers and the earlier manager-owned symlink form remain recognized and replaceable. [ownedTarget](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/globalbins/globalbins.go#L351) explicitly documents this compatibility. However, [StageForwarding](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/globalbins/stage.go#L133) records forwarding replacements as byte targets. The [publication recheck](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/staging/boundaries.go#L459) then refuses the owned link as `source_output_overlap: … is now a link`.

`TestWave2GlobalOwnership` calls `install.Global` with real manager locks, staging, ownership ledger, and transaction engine. In the retained reproduction, the legacy link is emitted by `globalbins.Refresh` itself; it is not merely a hand-written approximation. `Refresh` is used to create the legacy fixture, while the failing live operation is the current `Global` entry.

| Existing shim state | Production Global result | Invariant assertion |
|---|---|---|
| Owned regular launcher | `ok` | Pass |
| User replaces owned launcher with foreign regular bytes | `ok`; foreign bytes preserved | Pass |
| Owned link emitted by Refresh, with ownership ledger | **`failed`, `source_output_overlap`** | Fail |
| Foreign regular bytes introduced after transaction preparation | `failed`, preimage-digest mismatch; foreign bytes preserved | Pass |

Go process exit: **1**, command 07. Command 04 reproduced the same failure using an explicitly constructed legacy link. `Global` is a library entry and returns a status, not an OS exit code. No foreign-file overwrite was observed. This does not establish that every previously released installation uses this link form.

A fix should reconcile the recognized link as an owned directory entry, carrying its exact target through backup/rollback, while retaining foreign-file and changed-preimage refusals. It should not weaken the general prohibition on following links during managed writes.

## N8. `audit --allow` accepts a path in place of a content digest

**Confirmed. Low local-input validation risk: pin state escapes its intended audit namespace.**

Promised invariant: an audit pin addresses a content identity. The [CLI](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/cmd/curator/main.go#L2111) requires a reason but passes `--allow` directly to `audit.Pin`. [trustDir](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/audit/audit.go#L409) joins the value beneath `home/audit`; [hashing.Normalize](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/hashing/hashing.go#L250) normalizes spelling but validates neither hex syntax nor path components.

`TestWave2AuditPinNamespace` drives the real `run` CLI dispatcher. All resulting writes remain inside its temporary manager home.

| Input | CLI exit | Result |
|---|---|---|
| Valid 64-hex digest, nonempty reason | 0 | Record under `home/audit/<digest>/trust.json` in the companion pin test |
| `--allow ../wave2-outside-audit --reason "synthetic approval"` | **0** | **Creates `home/wave2-outside-audit/trust.json`** |

The required refusal assertion fails. Go process exit: **1**, commands 02, 07, and 12. The exact filename remains `trust.json`; this is not a demonstrated arbitrary-filename write. The operator supplies the malformed flag, and this pass established no package-controlled route to it or privilege escalation.

A fix should parse a supported content identity before filesystem access, refuse all non-digests, and verify containment before writing any pin state.

## N9. A successful pin omits its required creation time

**Confirmed. Low provenance/conformance risk.**

The [pinned manager source-audit contract](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L1122) requires a pin to record content identity, operator identity, reason, and creation time. The creation-time requirement already exists in [rc.13](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/profiles/manager.md#L1097). [Pin](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/audit/audit.go#L379) writes no timestamp field.

| Probe | CLI exit | Stored result | Assertion |
|---|---|---|---|
| `TestWave2AuditPinCreationTime`, valid digest, synthetic operator and reason | **0** | Record exists, creation time absent | Fail |

Go process exit: **1**, commands 02, 07, and 12. Inspection of the complete writer object confirms no creation time under another key; this is not merely a field-name mismatch. This pass does not claim an expiry bypass—no pin expiry was exercised. A fix should write a defined creation timestamp and test its validity alongside the existing pin fields.

## Coverage: production evidence versus reading

Ratios below describe the enumerated probes or pinned corpus, never the whole possible input space. Green refusal tests deliberately supplied invalid inputs; red finding tests assert the required safe behavior.

| Surface and invariant | Dynamic evidence and measured coverage | Read-only or unproven bound |
|---|---|---|
| Transaction target namespaces are independent | New `Engine.Prepare`/`Commit` probes under a real home lock: **4/4** cases—disjoint, duplicate path, symlink-parent alias, hardlink alias. Existing save-time alias, recovery alias, and rollback tests also passed. Commands 05, 06, 09. | Exhaustive path-exchange schedules, crash combinations, and every filesystem are unproven. |
| Concurrent installs publish consistent snapshots/state | `Project`: two projects reach staging together, share one home/source, then both install and execute their shims successfully. `Capture`: **4/4** stable/byte-change/replaced-inode/membership cases. `PublishLocal`: **8/8** concurrent writers, then `OpenLocal` authenticates matching bytes. Both concurrency tests pass `-race`, command 08, exit 0. | Bounded in-process interleavings; not a cross-process stress test or proof against all filesystem races. No concurrent source edit during the two-Project scenario. |
| Global writes replace only owned shims | **4/4** states exercised: **3 pass, 1 red N7**. Late foreign replacement is preserved on failure. Command 07. | Stale-removal races and Windows forwarding wrappers were not dynamically exercised. |
| Managed context writes do not follow foreign links | **10/11** pinned nofollow corpus cases dynamically driven through `UseWithPolicy`/`Resolve`, **1 explicitly bound**, no skips; command 06, exit 0. Direct rendered-document symlink control passed in command 05. | Bound row: unauthorized, unowned-link backup has no production operation that requests it; adjacent takeover paths were exercised. Not all credential-link and backup variants were rerun. |
| Credential helper output is bounded and intact; prompts stay disabled | **6/6** answer-size cases run: **4 pass, 2 red N6**. Four existing prompt/namespace/read-back controls passed, command 10. | Oversized stderr is discarded but helper CPU/lifetime under unlimited output was not stressed. Duplicate keys, every malformed answer grammar, and platform keychains are unproven. |
| Worker frames stay within opening-frame size/format rules | `godriver.RunWorker` and `scriptworker.RunWorker`: **14/14** refusal cases pass; each worker returns **3**. Zero/oversized lengths, short header, truncated payload, unknown field/kind, trailing JSON. Commands 01 and 09. | These are direct hidden-mode production entries, not successful authenticated worker sessions. No allocation of a 64-MiB at-limit frame, full permit-state sweep, or complete canonical-JSON grammar proof. |
| npm/tar limits and content policy apply before admission | `Parse` → `CaptureAndAdmit` → real capture store, SRI, recursive artifact policy, embedded metadata: **8/8** cases pass. Valid input and 255-byte component admitted; SRI drift, 2-MiB high-expansion fixture, traversal, SRI-correct second gzip stream, WebAssembly leaf, 256-byte component refused. Commands 04 and 09. | **2/10** declared numeric limit fields touched: component length and expansion ratio; **1/10** tested exactly at limit and limit+1. No heavy raw/leaf/aggregate budget test. Raw-file reads and project copying precede policy admission, so a pre-admission memory/disk bound is **unknown**. Offline npm execution/cache replay after admission was not run. |
| Umbrella discovery follows its selected trust-root revision | **14/14** rc.14 cases exercised for both A and B through the resolver, command 06. Real `run` dispatch verified warning-phase behavior and exact argv/environment transport, including empty/newline/quote/literal-dollar values. Commands 05, 06, 09. | The shipped switch is **revision A**: PATH selection outside trust roots warns and runs. That deliberate rollout behavior is not a new finding. Revision B is not claimed as active. Native Windows lookup was not run. |
| Launcher fragment and permission transport preserves the admitted decision | Separate launcher: **5/5** new `run` probes plus **31/31** existing permission rows pass, commands 03 and 11. Native argv/environment values survive; profile permission mode reaches typed admission; force-native/yolo conflict exits 2 before build; trailing fragment and foreign PATH injection exit 1 before build/child. Manager `env resolve`: **4/4** permission rows plus schema assertions pass, command 06. | Local recording provider and scripted resolver in the launcher fixture; manager producer and launcher consumer were exercised separately. No single combined Curator→launcher→real-agent deployment or live agent behavior claim. |
| Pins/revocations obey source-audit decision precedence | **10/10** rows run through both `Gate` and `GateReadOnly`: **20/20** expected decisions; command 13. Includes strict threshold/off, advisory, old schema pin requirement, pin over findings, hash revocation over pin, and source/Git globs. `Project` also refuses a signed loopback-registry revocation despite a local pin. Draft install pin/revocation cases and **4/4** broken-report cases pass. | N8/N9 cover pin writing, not an authorization bypass by remote content. Cloud backends, disabled-audit policy, canary failure injection, all severity thresholds, expiry, and hash-version migration are not established by this table. |
| Windows-specific paths | **0 native dynamic cases**; source reading only, below. | No Windows conformance or runtime safety claim. |

The npm `inspectTarball` helper alone is not reported as a limit bypass: the SRI-correct artifact probes reached the real policy chain preceding it. Conversely, downstream rejection does not establish a limit on resources already spent reading/copying upstream inputs.

## Windows reading

- Transaction identity uses [completeNamespaceIdentity](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/transaction/namespace_identity_windows.go) to force Go's deferred identity read and fail closed if it cannot be established. Snapshot capture uses [staging.FileIdentity](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/snapshot/identity_windows.go), rather than relying on later `SameFile` reads.
- [Durable rename](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/transaction/durability_windows.go) uses `MoveFileEx` with write-through; replacement separately adds replace-existing. Some directory flush failures are tolerated by design. Crash durability remains untested here.
- [npm junction handling](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/npmsource/linkentry_windows.go) recognizes irregular directory reparse points and normalizes the junction before downstream containment. Native junction exchange, volume boundaries, and DACL behavior are unproven.
- [Windows launcher wrappers](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/runtimestore/runtimestore.go#L175) explicitly document that `%VAR%` inside forwarded arguments cannot be preserved through `cmd.exe call`'s second expansion. This is a stated contract bound, not a newly reproduced defect. Umbrella PATHEXT behavior was read, not run on Windows.
- The [HTTPS broker named pipe](https://github.com/relux-works/curator/blob/e5489b6ba22c9e9cf7a925e03f3653d115188999/internal/buildrepo/httpsbroker_pipe_windows.go#L213) constructs a protected current-user DACL. Its client accumulates data until pipe end without an independent total-byte cap. Whether a harmful stream can arise across the protected transport is **unknown**; no Windows red probe exists, so this is not numbered as a confirmed finding.
- The manager permission-fragment test records a Windows absolute-path schema gap and skips only that schema assertion on Windows. This run was on macOS and does not discharge that gap.

## Command ledger and evidence

The attached evidence bundle contains exact argv, working-tree identities, exit codes, and redacted full output for every numbered test process. The first failed harness runs remain available; they were not converted to passing evidence.

| Command/log | Validation selection | Actual exit | Interpretation |
|---|---|---:|---|
| 01-framing | New credential and both worker probes | **1** | Expected red N6; worker packages pass |
| 02-surfaces | Initial install/snapshot/npm/audit/CLI probes | **1** | Install fixture did not compile (`OnStaged` signature); snapshot/npm/audit pass; N8/N9 red. Initial second-stream row reached only SRI, then was strengthened |
| 03-launcher | Five new launcher probes and 31 permission rows | **0** | All pass |
| 04-install-npm | Corrected install fixture and strengthened 8-row npm chain | **1** | Concurrent installs pass; N7 red; npm passes |
| 05-existing-production | Existing namespace/rollback, nofollow, umbrella, permission emission | **1** | Incorrect relative conformance-root path; selected non-corpus tests pass. No corpus pass claimed |
| 06-production-rerun | Absolute pinned conformance root; namespace, nofollow, umbrella, permission emission | **0** | Corrected checks pass; nofollow has one explicit bound |
| 07-confirmed-red | Credential, global ownership, CLI pins | **1** | N6–N9 reproduced; legacy link now emitted by Refresh |
| 08-race | Concurrent Project installs and PublishLocal with `-race` | **0** | Both pass; no race detector report |
| 09-green-probes | Worker, npm, audit, capture, namespace, real umbrella transport | **0** | All selected probes pass |
| 10-credential-controls | Prompting, host namespace, provider near miss, unpersisted write | **0** | Four controls pass |
| 11-launcher-probes | Retained/formatted five launcher probes | **0** | All pass |
| 12-pin-red | Retained CLI pin probes | **1** | N8/N9 reproduced |
| 13-audit-paths | Expanded 10-row table, registry-over-pin, draft pins/revocations/broken reports | **0** | All pass |

`gofmt` exited **0** on the retained probe files. Source-integrity comparison exited **0**, with zero tracked-byte differences. No repository-wide test pass is claimed; only the selections above were run. Expected-red tests remain red in the attached suite and must not be used as a green acceptance gate for a product fix.

Artifact verification exited **0**: all 11 probe-file digests, 13 command exits against their logs, four finding reproductions, source-integrity record, public-path redaction, and report whitespace were checked. `git diff --check` exited **0**.

Task-scoped outcome resources:

- `TASK-261004-2wvbzz_results.md`: this report, updated before handoff.
- `TASK-261004-2wvbzz_probes.tar.gz`: exact Go probes at their respective repository paths, replay README, and SHA-256 manifest. Intended consumer: the follow-up reviewer and regression-fix implementer.
- `TASK-261004-2wvbzz_evidence.tar.gz`: command ledger, redacted logs, source-integrity record, and verification scripts.
- `TASK-261004-2wvbzz_logbook.md`: findings, interpretation decisions, and harness corrections. The repository's `LOGBOOK.md` was not edited, as the current brief requires.

The findings establish four bounded failures. The remaining surfaces have the explicit no-new-finding bounds above; unmeasured behavior remains unknown.
