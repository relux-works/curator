# CIP-NNNN: Manager tool trust and comparison follow-ups

- **Status:** Draft — design pending; implementation is not authorized by this document. CIP number awaits the spec owner.
- **Owner:** ivan-curator (orchestrator); decision: operator.
- **Created:** 2026-10-04.
- **Related:** TASK-261004-3qgmvx — csk-gap-analysis-follow-ups-design; parent STORY-261004-3e03l7 — design-csk-gap-follow-ups; [curator#87 — Go family allowlist and toolchain policy](https://github.com/relux-works/curator/issues/87).
- **Affects:** Protocol Core §§2, 4.1, 6.2, 8.2; Manager Profile §§2.1–2.2, 3, 4.2, 7, 12.3; curator acquisition, installation, launchers, audit and Go qualification.

## Summary

All seven reported behaviors remain present on current main, with materially different meanings: tool lookup and audit attribution cross trust boundaries; reserved names and extraction totals lack policy; dependency-directory precedence follows the existing specification; selective global install is an absent feature; newer Go families are deliberately refused. Recommend an operator-selected sequence led by manager tool resolution and explicit unsupported-backend refusal, with specification changes accompanying each affected behavior. Keep audit execution design with its existing owner, and qualify actual Go families before widening admission. This is a research handoff, not permission to implement or a release/conformance claim.

## Motivation and research boundary

An installed skill must not supply the manager's Git or SSH discovery executable. An operator selecting an audit backend must be able to tell whether it ran. Updating one global skill should not silently remove unrelated installations, and snapshot admission needs a finite resource policy.

The decision to unblock is which follow-ups to prioritize and which normative rules to adopt. Research budget: one initial worker slice within 60 minutes; one research document and one small evidence packet, each below 64 KiB; no archive and no second serial research prerequisite. Grammar freezing is not applicable: this draft proposes decisions, not a new accepted wire format. The first consuming production slice is safe Git resolution through the existing install/acquisition entry points.

The public-safe task extract was sufficient; the private comparison report was neither requested nor copied. No product source, tests, configuration or LOGBOOK.md in the worktree is changed. All experimental sources, repositories, configuration, HOME, Go caches and subprocess temporary roots were scratch-only. No real credential, token, SSH agent or Keychain entry was read or exercised. Logbook recording is explicitly inapplicable under the task-specific prohibition; important findings are recorded here and in task notes instead.

### Frozen evidence identities

- Curator main: [ca1b776fb580ec0cee0173bf150daf063023aeaa](https://github.com/relux-works/curator/tree/ca1b776fb580ec0cee0173bf150daf063023aeaa). HEAD, local main, origin/main and fresh public `refs/heads/main` agreed at observation. This supersedes comparison baseline `876127f`. Every curator file:line below refers to this commit, not a claim about a later main.
- Spec: **v1.0.0-rc.14**, tag object `661bead088186db70c9f57ad0b115305ec2bef35`, peeled commit **43bf0a2506d5c354a73bbc3ea4623d4653db10c7**. Read exact public raw files and independently checked the tag. Curator pins the same revision in `.github/workflows/ci.yml:38–43`.
- Normative sources: [Protocol Core at rc.14](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/core.md), [Manager Profile at rc.14](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md). Spec line numbers below are from these files.
- Probe host: darwin/amd64, Go **1.26.0**. Other OSes and a real Go 1.27 toolchain were not tested. Evidence packet: **TASK-261004-3qgmvx_probe-evidence.json**, including scratch probe sources, source digests, commands, observations and real exit codes.

## Current state: seven dispositions

### 3. Manager tools can resolve through skill shims — CONFIRMED, high

`internal/gitops/gitops.go:225–240`, `:596–605`, and `:813–817` invoke bare `git`; the ordinary install route reaches resolution at `internal/closure/closure.go:361` and extraction through `internal/snapshot/snapshot.go:131`. `internal/install/buildsshcandidates.go:125–145` uses LookPath then final-target EvalSymlinks for ssh-add. Its child already has a small environment; environment filtering does not validate the selected executable. Another relevant selection site is `cmd/curator/main.go:1975–1985`.

**Measured:** production `gitops.Resolve` executed a synthetic Git in `.agents/bin`, both directly and through an outside → shim-directory → outside symlink chain, and accepted its synthetic commit. The ssh-add helper accepted one synthetic identity from a shim despite a nonexistent synthetic socket. The latter drives the helper, not the discovery menu. Absolute shim paths alone suffice; no claim that Go accepts every relative-PATH shape is needed.

**Design:** one resolver for manager-owned tool invocations, with explicit forbidden roots, absolute search entries, component-by-component link checks and launch-boundary identity rechecks. Skip empty/relative search entries; reject a candidate touching a forbidden directory at any hop. Keep separately pinned Git distributions and closed worker toolchains at their stronger trust level. Proposed normative text: **S1**. A path policy prevents this demonstrated class; it does not authenticate every arbitrary executable in an operator-supplied directory.

### 4. Ordinary closure has no reserved system/manager names — CONFIRMED, medium; high when composed with item 3

`internal/closure/closure.go:233–245` enforces one exact active-name owner without a reserved-name check. `internal/identifiers/identifiers.go:27–55` rejects portable filename hazards, including Windows device names; it does not reserve tools. Production project dry-run admitted **4/4** candidate names: `git`, `curator`, `ssh-add`, `GiT.ExE`. This does not establish behavior for every collision or filesystem.

Core §4.1 (`core.md:203–213`) requires one owner but supplies no system-name list. Core §2 (`:51–68`) otherwise compares identifiers case-sensitively and reserves Windows device basenames. Add an explicit security comparison rule without silently changing general identifier equality. **Design:** reject reserved exported script/build names at planning, including a conflict with declared system requirements; maintain a small normative system set plus an implementation-owned manager namespace. Proposed text and exact initial set: **S2**. Reserving all executables found on the local PATH would be nondeterministic and too broad.

### 6. Declared dependency directories enter the launcher prefix — CONFIRMED; spec violation REFUTED, medium

`internal/install/install.go:1022–1066` starts with the scope bin directory, resolves declared system commands/dependencies and appends their containing directories. `internal/install/targets.go:191–199` supplies that list during real staging; `internal/runtimestore/runtimestore.go:161–172,190–205` puts the whole list before inherited PATH on Unix and Windows.

**Measured:** install with PATH ordered preferred-directory → dependency-directory, then execute the published command with only the preferred directory and system directories inherited. An unrelated executable sharing the dependency directory wins over the preferred executable. The launcher returned `dependency`, not `preferred`.

Manager §3 (`manager.md:662–669`) expressly requires the current prefix. **Design:** manager-generated runtime/bin directories, then inherited PATH, then deduplicated resolved dependency directories. Dependencies remain discoverable with a sparse PATH, but caller choices take precedence. This changes ordinary launcher behavior; enforced script-worker PATH remains separate under §3.1 (`:671–676`). Proposed text: **S3**. Coordinate shell activation with TASK-261004-1z2pgb — shell-hook-no-source-and-path-append-design; this item concerns command launchers, not that sibling's shell hook.

### 8. Configured audit backend is unused but labels static verdicts — CONFIRMED, high assurance risk

`internal/config/config.go:963–1011` accepts a backend name and raw backend definitions. In the audited production package, selected backend/model are used for cache naming and stored labels (`internal/audit/audit.go:413–415,465–478`); cache misses run deterministic detection (`:307–325`), and the canary is static (`:195–205`). No backend dispatch occurs on that path.

**Measured:** `audit.Gate`, strict mode, fresh scratch cache, backend `synthetic-unimplemented`: zero blocking errors and one persisted verdict labelled with that backend. No backend ran. This demonstrates false attribution and absent refusal; it does not demonstrate actual cloud egress or credential leakage.

Manager §7 (`manager.md:1106–1119`) says backend invocation is optional and specifies backend-failure policy. Clarify optional capability versus an explicitly selected unsupported capability. **Design:** the immediate implementation refuses an enabled, selected unsupported backend before cache reuse or audit; `null` explicitly means static analysis only. Runtime failure of a supported backend keeps the existing strict/advisory distinction. A later real backend requires environment, canary, request-size and egress controls. Proposed text: **S4**. Detailed design and secret transport remain with TASK-261004-hy8zmn — audit-token-argv-and-backend-env-allowlist-design.

### 9. Global install lacks `--only NAME` — CONFIRMED feature gap, medium operational impact

`cmd/curator/main.go:714–741,1810–1817` has no selector. The scratch production CLI returned **exit 2**, `flag provided but not defined: -only`. Global planning resolves the whole manifest (`internal/install/global.go:128–150`), derives the desired whole state (`:393–428`), and removes stale entries outside that set (`:445–449`). Filtering nodes alone would therefore threaten retained state.

Manager §4.2 (`manager.md:1037–1043`) defines the global scope but no selector; profile-scoped global state also has obligations in §12.3 (`:2723–2738`). **Design:** select one direct skill from the current frozen global/profile selection; resolve its dependency closure, union it with authenticated retained installation state, and commit one transaction. Do not silently advance a profile lock or synchronize unrelated environment members. Reject shared-dependency conflicts instead of widening selection. Proposed retained-state contract: **S5**. Until this exists, retaining current whole-scope behavior is a valid product choice.

### 10. Go 1.26/1.27 are unqualified/refused — CONFIRMED, medium availability and maintenance impact

`internal/godriver/session.go:42,79–89,245–257` permits only family 1.25. A real production `godriver.Probe` of host Go 1.26.0 returned `unsupported_go_family`; actual 1.27 execution remains unmeasured, although the same closed map excludes it. [Issue #87](https://github.com/relux-works/curator/issues/87) is open and describes the same policy problem. The [Go release policy/history](https://go.dev/doc/devel/release) lists 1.26 and 1.27 and supports a family until two newer major releases exist; preserving only 1.25 is no longer a sustainable support policy.

**Additional spec mismatch:** Manager §2.2 (`manager.md:193–201`) still mandates support for a tested Go **1.23** family while permitting tested additional families; current code admits neither 1.23 nor 1.26/1.27. This sentence needs reconciliation, not just an allowlist edit.

**Design:** retain a release-owned explicit tested-family matrix; qualify 1.26 and 1.27 separately against pinned go-v1 vectors on supported platforms, then admit them. Never use a minimum-version rule or a package-controlled range to authorize an unknown toolchain. Full version and tree digest already enter build/cache identity (`internal/buildmeta/models.go:67–72,101–109,326–328`; Core §8.2, `core.md:1540–1576`), so adding families need not invent another identity field. Source registry audit remains a statement about source, not proof of binaries built by every compiler. Proposed text: **S6**. No qualification claim is made here.

### Follow-up: aggregate legacy extraction budgets — CONFIRMED, medium denial-of-service risk

`internal/gitops/gitops.go:28,675–699` caps each file at 512 MiB but has no total-byte or file-count budget. `listTree` buffers the entire listing before parsing (`:596–605`), then retains every entry. `Extract` calls it before planning/writing (`:523–556`). Snapshot creation and cache authentication both reach extraction (`internal/snapshot/snapshot.go:91,131`), so an existing cache does not remove the exposure.

**Bounded probe:** `planWrites` accepted five distinct 512 MiB metadata entries: **2,684,354,560 declared bytes**. No blob contents were allocated or extracted. This helper-only observation confirms the missing proposed 2 GiB preflight bound, not a measured disk-exhaustion exploit. No large file-count probe ran; the count finding is static.

Core §6.2 (`core.md:1221–1225`) delegates documented implementation limits without requiring these aggregates. The stronger external-repository path already has 200,000 files and 2 GiB expanded-object defaults (`internal/buildrepo/admission.go:68–85,738,1058`); those are not inherited by legacy extraction and have different accounting. **Design:** explicit per-snapshot count, per-output-path byte total and bounded listing parser, shared by creation and reauthentication. Proposed text: **S7**.

## Evidence and verification

All commands below ran directly as standalone processes, without a pipe or tee. The evidence packet supplies exact commands and scratch sources. Each `go test` used `-count=1 -timeout=180s -v`; packages are under `./internal/`. No prior attached test result was accepted as a substitute, and no full product or multi-platform suite was run.

| Command / mask | Real exit | Observation and interpretation |
|---|---:|---|
| `go build -o <scratch>/curator ./cmd/curator` | 0 | Production CLI compiled from the pinned source export. |
| `go test ... ./internal/gitops -run '^TestResearchGitRejectsSkillShim$'` | 1 | Expected-red: both direct and intermediate-link candidates executed. |
| `go test ... ./internal/install -run '^TestResearch(ReservedCommandsRejected\|SystemDependencyDoesNotPreemptAmbient\|SSHAddRejectsSkillShim)$'` | 1 | Expected-red: 4 reserved-name candidates admitted; dependency sibling won; synthetic ssh-add executed. |
| `go test ... ./internal/audit -run '^TestResearchConfiguredBackendMustRunOrRefuse$'` | 1 | Expected-red: unsupported selected backend did not refuse and labelled static output. |
| `curator global install --only sample` in scratch HOME/config | 2 | Expected refusal: option is absent. This is feature evidence, not a passing install. |
| `go test ... ./internal/godriver -run '^TestResearchHostGoQualification$'` | 1 | Expected-red: real Go 1.26.0 refused; no qualification performed. |
| `go test ... ./internal/gitops -run '^TestResearchAggregateSnapshotBudget$'` | 1 | Expected-red against the proposed 2 GiB bound, helper only. |
| `go test ... ./internal/install -run '^(TestEndToEndInstall\|TestRuntimeLauncherCapturesDeclaredSystemDependency\|TestMissingSystemCommandFails)$'` | 0 | Three existing controls passed, including current spec behavior and missing-command refusal. |
| `go test ... ./internal/gitops -run '^(TestCloneAndResolve\|TestExtractProducesExactTree)$'` | 0 | Two existing Git/extraction controls passed. |
| `go test ... ./internal/audit -run '^(TestGateModes\|TestCanaryFires)$'` | 0 | Two existing audit controls passed. |

Coverage is **7/7 items inspected**, **6/7 exercised via a production API or CLI**, with the extraction total **1/7 helper-only**. The ssh-add subcase is also helper-only. There were 11 distinct expected-red Go assertion scenarios and 7 passing existing top-level control tests. These are research observations, not production regression coverage for the proposed changes.

Two preliminary attempts were excluded from that coverage: an `EnsureRepo` probe exited 0 but never called Git (it only stats `.git`), and the first CLI probe exited 1 because its explicitly selected synthetic system config was missing. The corrected probes above reached `Resolve` and flag parsing. Both preliminary logs and exit codes remain in the evidence packet. The Git resolver probe was repeated after adding the aggregate scratch test to the same file; it again exited 1, and both executions are preserved. Failed web-cache reads were replaced with successful unauthenticated reads of the exact public spec revision; unsuccessful exploratory board projections/searches are not validation evidence.

Artifact checks: the scoped naming gate and `git diff --check` exited 0. The artifact verifier exited 0 and checked all seven required finding sections, 21 concrete source-anchor bounds, evidence digests, byte budgets, publication-path patterns and the absence of tracked source/LOGBOOK changes. Its first version exited 1 on an unjustified minimum of 25 citations; that arbitrary count was replaced by one concrete source anchor per required finding while retaining every path/line-bound check. This verifier does not establish the semantic truth of a claim; the source reads and probes above do that.

## Security considerations and compatibility

The attacker controls package command names, manifests, scripts and repository sizes; previously installed package shims may already be on PATH. The operator controls tool directories and audit configuration. Absolute spelling alone does not cross this trust boundary safely: symlink ancestry, directory identity and replacement races matter. Resolve from a neutral directory, bound link traversal, reject unreadable/ambiguous checks, and recheck before spawn. No portable pathname-only design should claim elimination of a hostile same-user replacement race.

The reserved-name list is defense in depth, not an exhaustive list of all OS executables. Ordinary skill scripts still execute under their documented policy; changing PATH precedence does not make them sandboxed. Budget checks must precede buffering or writing the data they bound. Count duplicate blob IDs once per output path for extracted bytes, use overflow-safe arithmetic, and stop/join extraction children before discarding private staging.

Existing reserved exports will require renaming and a managed reinstall; do not silently delete or rename commands. Rebuild generated ordinary shims when the new PATH contract is adopted, preserving argument/exit behavior. Ignore or invalidate legacy cache entries that purport to come from an unexecuted backend; never relabel them as backend evidence. Keep `null` static mode usable. A selected unsupported backend becomes a visible configuration failure. `--only` is additive and leaves whole-scope commands unchanged. Go family admission preserves existing version/digest separation; retention or retirement of 1.25 must be explicit. Per-snapshot limits may newly reject large repositories and must name the exceeded metric.

## Design options and tradeoffs

| Option | Mechanism | Benefit | Cost / security limit |
|---|---|---|---|
| A. Document current behavior and defer changes | Preserve rc.14 and whole-scope installs; publish limitations | Smallest compatibility cost | Confirmed tool execution and audit misattribution remain. Not recommended. |
| B. Staged contracts and bounded implementation leaves | Resolver and backend refusal first; adopt reserve/PATH/budget rules; qualify Go; decide selector separately | Addresses evidenced trust failures without starting a backend platform; each change is independently reviewable | Requires coordinated spec revisions and explicit migration notices. Recommended. |
| C. One broad hardening/feature release | Deliver all seven, real backends and selector together | One migration event | Couples security fixes to backend egress and retained-state design; highest testing and review cost. |

### Recommendation

Choose **B**, subject to operator prioritization. First close manager tool selection and unsupported-backend attribution, then apply extraction budgets and explicit reserved-name policy. Put Go family qualification next or alongside those independent leaves because the current family is aging. Change launcher ordering only with the Manager §3 amendment. Ship global `--only` only if its productivity value warrants the retained-state transaction work. Do not launch implementation from this research handoff.

The default controls should be mandatory safe tool selection, a closed reserved-name rule, static audit when backend is `null`, explicit refusal of unsupported selected backends, fixed documented extraction ceilings, and release-qualified Go families. No manifest escape hatch. The operator retains trusted tool-directory selection and explicit backend choice; selective installation is an explicit CLI action. Detailed new backend configuration belongs to its existing design task.

## Open questions for the operator

1. **Priority:** approve option B's sequence, or prioritize Go availability ahead of the independent naming/PATH changes? **Recommend:** tool trust and backend attribution first, budgets and Go qualification next; selector later.
2. **Tool compatibility:** should arbitrary absolute operator PATH directories remain eligible after managed-directory exclusion, or require an explicit trust-root list for every manager tool? **Recommend:** existing trusted selection first, sanitized absolute PATH as the compatibility fallback for legacy tools; disclose that this does not attest tool authenticity.
3. **Reserved names:** accept S2's initial closed set, manager provider namespace, and repeated extension stripping including `.com`? **Recommend:** yes; reserve active declared system requirements too, without enumerating the whole installed OS.
4. **Missing dependencies and unsupported backend:** reject the selected scope atomically, including dry-run, rather than partially publishing unrelated members of that scope? **Recommend:** yes; unsupported backend is a configuration error in both audit modes, distinct from a supported backend's runtime failure.
5. **Launcher ordering:** accept that an inherited executable may now override a resolved dependency and that a sparse PATH still gets the dependency suffix? **Recommend:** yes, for ordinary launchers only.
6. **Selective installs:** authorize a current-profile, one-direct-skill selector with frozen unrelated state and rejection of conflicting shared dependency changes? **Recommend:** yes as the first version if this feature is prioritized; no implicit profile update or prune.
7. **Go policy:** replace the obsolete mandatory 1.23 sentence with a published tested-family matrix and qualify 1.26/1.27? **Recommend:** yes; decide 1.25 retirement explicitly, never infer support for later families.
8. **Budget values:** adopt 200,000 output files, 2 GiB extracted bytes, 512 MiB per file and 64 MiB listing bytes per snapshot? **Recommend:** these initial fixed ceilings, with 4,096 path bytes and depth 128 to bound records. They are design choices, not measured corpus percentiles or whole-operation limits.

## Specification changes: normative sketches

These are proposed decisions for a successor to rc.14, not edits to published rc.14 artifacts.

**S1 — Core §6.2 / Manager §2.1, manager tool resolution.** Before invoking a manager tool, a manager MUST resolve one absolute executable independently of package data. Empty and relative search entries MUST NOT participate. The selected entry, every traversed link/reparse hop and its directory ancestry MUST remain outside package sources, runtime trees, project/hybrid/global skill bins, managed environment roots and directories used to publish skill forwarding shims. Unreadable checks, cycles or exceeded traversal bounds MUST refuse. The manager MUST recheck executable identity at launch; stronger existing pinned-distribution and closed-worker requirements remain binding. Directory exclusion and file/link comparison MUST use the host's path semantics.

**S2 — Core §4.1 / Manager §2.1, reserved exports and failure scope.** Before publication, a manager MUST reject an active exported script/build name whose ASCII-lowercased spelling, after repeatedly removing a terminal `.exe`, `.cmd`, `.bat`, `.ps1` or `.com`, matches a reserved name. Initial shared set: `git`, `ssh`, `ssh-add`, `ssh-keygen`, `gpg`, `sh`, `bash`, `zsh`, `cmd`, `powershell`, `pwsh`, `env`, `sudo`, `doas`, `go`, `python`, `python3`, `node`, `npm`, `npx`, `curl`, `wget`. Implementations MUST also reserve their manager basename/provider namespace (for curator: `curator` and `curator-*`) and the normalized names of declared system executable requirements in the selected closure. System requirement declarations themselves remain allowed. General identifier equality is unchanged. Every declared system requirement of a node selected into the operation closure, including a context-only node, MUST be checked; a missing requirement MUST fail that scope's plan with its hint before publication. Independent scopes may report independent outcomes, with a nonzero aggregate exit if any fail.

**S3 — Manager §3, ordinary launcher PATH.** An ordinary command launcher MUST compose PATH in this order: manager-generated command/runtime directories required by the command; inherited PATH; directories of safely resolved declared system dependencies. Package-declared dependency directories MUST NOT enter the manager prefix. Duplicate added directories are suppressed using platform path equality; unset or empty inherited PATH adds no empty search component. Argument forwarding, child exit status, direct invocation and §3.1 enforced-worker rules remain unchanged.

**S4 — Manager §7, backend truth, child environment and secret transport.** An enabled or explicitly requested non-null selected backend MUST be implemented; otherwise it MUST cause `audit_backend_unsupported` before cache reuse or publication. New backend findings MUST come from actual execution under the declared policy. Cache reuse MUST bind a prior successful execution of that backend/model to the same versioned input and policy identity; a static-only entry is not such evidence. Static findings MUST identify their actual producer and MUST NOT attest an unexecuted backend/model. Backend children MUST receive an environment constructed from an explicit versioned allowlist, private HOME/config/temp roots and only deliberately scoped credentials; ambient environment inheritance is forbidden. A backend MUST pass its own health/canary contract before its findings count, and unavailable canary evidence MUST NOT count as success. Encoded requests MUST fit the configured byte cap (current default 1 MiB, maximum 10 MiB), and cloud egress MUST require explicit public-source classification, cloud permission and prior secret redaction. Secret values MUST NOT be accepted in CLI argv or echoed; dedicated stdin/descriptor or scoped credential-store transport is preferred. Any retained environment-token input is captured deliberately and excluded from unrelated children. The precise CLI migration and backend contract are owned by TASK-261004-hy8zmn — audit-token-argv-and-backend-env-allowlist-design; current argv exposure is at `cmd/curator/main.go:2118–2120`. Supported-backend runtime failures retain strict-block/advisory-warning behavior, with no successful-backend label.

**S5 — Manager §4.2 / §12.3, selective retained state.** `global install --only NAME` MUST name one existing direct skill in the current selected global/profile state. It MUST operate from a frozen manifest/profile-lock generation, without advancing that lock, fetching unrelated sources or synchronizing unrelated environment members. The replacement set is the selected skill and required dependency closure. The desired published state MUST be that replacement set union authenticated retained state, including context, markers, runtimes, adapters, forwarding shims and live build references. Unselected installed members, including otherwise-stale members, MUST remain unchanged; this operation is not prune. Dependencies needed by any retained consumer remain live. A conflicting shared version/command, unknown name, unreadable retained state or changed generation MUST refuse the whole selected transaction. Dry-run MUST describe replacements and retention without writes. A later full-scope operation remains the explicit reconciliation path; retained state is not evidence that those members passed a fresh audit.

**S6 — Manager §2.2 / Core §8.2, Go support policy.** A manager release MUST publish its tested Go family/platform matrix and MUST admit a family only after qualification against the pinned go-v1 vectors. It MUST refuse an unqualified family with the observed family, tested set and an operator-selected-toolchain remedy; it MUST NOT auto-download or accept a package-selected compiler. Each admitted execution MUST retain the normalized full Go version and toolchain-content digest in build/cache/receipt identity. A source audit MUST NOT be presented as compiled-artifact attestation. Replace, rather than silently reinterpret, the current mandatory 1.23 sentence.

**S7 — Core §6.2 / Manager §2.1, snapshot budgets.** Every legacy raw-tree extraction and reauthentication MUST enforce documented per-snapshot maximum file count, aggregate output bytes and bounded tree-listing bytes in addition to the per-file cap. The initial manager policy is 200,000 files, 2 GiB aggregate bytes, 512 MiB per file and 64 MiB listing bytes, with path/depth bounds stated above. Each output path contributes its full size even when multiple paths share a blob. Accumulation MUST reject arithmetic overflow and over-limit input before persistent publication; the listing parser MUST enforce bounds while consuming its stream rather than after unbounded buffering. Blob framing and copied sizes MUST agree with the bounded plan; any failure MUST stop children and discard private staging, leaving installed state intact. Exactly-at-limit inputs are admitted. These bounds concern one snapshot, not total resources of a multi-snapshot operation.

## Implementation leaves and test plan implied by option B

No new board implementation tasks are created by this draft. Sizes are relative estimates: S = one bounded change, M = several coupled call sites, L = a new transactional feature. Spec decisions precede or land in lockstep with each owning leaf.

| Leaf | Size / dependency | First production slice and acceptance evidence |
|---|---|---|
| Tool resolver | M; S1 | Ordinary install → closure → Git resolve/extract plus SSH discovery and external Git selection. Direct shim, outside→shim→outside, parent-directory links, relative/empty PATH, unreadable path, loop, replacement and safe-tool controls on supported OSes. Mutants removing intermediate-hop or parent checks must turn named tests red. |
| Backend refusal and truthful cache | S; S4 and existing audit-design owner | CLI audit/install → Gate, both fresh and cached. Unknown selected backend in strict/advisory/dry-run refuses; null static mode still works; old falsely labelled cache cannot authorize it. Actual backend execution remains a separate later leaf with request N/N+1, canary, child-env sentinel and denied-egress cases. |
| Legacy extraction budget | M; S7 | Snapshot Get and reauthentication → Extract, through real Git fixtures and rollback. Count/byte/listing N and N+1; overflow; repeated blob IDs at distinct paths; cleanup after stream failure. Narrow each bound independently, not only remove the check. Keep CI fixtures small via explicit test limits. |
| Reserved names | S; S2 | Project/global/hybrid planning rejects normalized names before staging; ordinary nonreserved exports and system declarations succeed. Case variants, each suffix, repeated suffixes and retained-state conflicts. Derive expected reserved vectors from the adopted spec set. |
| Go 1.26 qualification | M; S6 | Real go-v1 acquisition/build/receipt/currentness on supported OS/architecture cells; no network compiler fallback; cross-toolchain cache miss and unknown-family refusal. Record measured matrix cells, patch versions and vector counts. |
| Go 1.27 qualification | M; S6, independently qualified | Same production slice and matrix; a green 1.26 lane is not evidence for 1.27. Include retirement/diagnostic policy for 1.25. |
| Ordinary launcher ordering | M; S3, shared resolver | Install and execute published Unix/Windows script and build shims. Ambient collision wins; dependency remains available with sparse/unset PATH; quoting, argument/exit preservation and enforced-worker behavior retain coverage. Mutant moving a dependency into the prefix must fail. |
| Global selective install | L; explicit product approval of S5 | CLI selector → frozen plan → serialized union publication → status/GC. Update A while B's content/markers/shims stay byte-identical; shared dependency conflict, removed-but-retained B, malformed marker, command collision, concurrent generation change and publication rollback. Freeze profile state explicitly. |

Items outside this packet remain with their current owners: BUG-261004-bknio5 — gc-sweeps-live-runtime-on-uncertain-marks; the separately tracked trust-pin work; TASK-261004-1z2pgb — shell-hook-no-source-and-path-append-design; and TASK-261004-hy8zmn — audit-token-argv-and-backend-env-allowlist-design. No evidence here accepts those items or another implementation's behavior.
