# TASK-261008-1c6bvv — Architecture review of CIP-0008/0009/0010

Research handoff, 2026-10-08. Scope: the donor worker design in curator-spec PR #134; no code or specification changes.

**Recommendation: revise before adopting the security and deployment claims.** The architecture has useful boundaries, but the proposed recipes do not yet establish them. The strongest blockers are the same-UID trusted bridge, incomplete harness lockdown, conflation of carrier and signing identities, and the assertion that sharing one Codex credential file serializes refresh. The one-host MVP also needs a cross-user credential broker that CIP-0010 explicitly defers.

P1 means a contradiction or missing contract blocks the advertised security or required behavior. P2 means the affected workflow or compatibility claim needs correction before qualification. These are design findings, not claims that an exploit was executed.

## Summary of findings

| ID | Severity | Location | Finding |
|---|---|---|---|
| F1 | P1 | C8 R3–R5 | Intake rejects the very flags required by the composer; the tail denylist is incomplete. |
| F2 | P1 | C8 R3, R7 | Codex mapping omits measured controls and mixes CLI and app-server contracts. |
| F3 | P1 | C8 R3, Security | Claude settings-source suppression does not establish the claimed hook, plugin and context boundary. |
| F4 | P2 | C8 R3; C9 D1 | Muse log suppression removes measured history and budget accounting. |
| F5 | P2 | C8 R3–R4, Current state | Release qualification and remaining local effects are overstated. |
| F6 | P2 | C8 R5, Compatibility; C10 C3 | Plan extensions, fragments and closed schema revisions are conflated. |
| F7 | P1 | C8 R6; C9 D1; C10 C2–C3 | Proposed locks contradict permission and credential-selection rules. |
| F8 | P2 | C8 Compatibility; C9 D1 | Remote-worker entry and composition ownership need an explicit decision amendment. |
| F9 | P1 | C8 R5, Security; C9 D2 | Rechecking an old plan cannot detect changed machine policy or prove applied confinement. |
| F10 | P1 | C9 D1–D2, Security | The same UID defeats bridge, key, audit and egress separation. |
| F11 | P1 | C9 D2, D4 | OS recipes do not yet constitute a usable, measured deployment boundary. |
| F12 | P1 | C9 D1, D3–D4 | An MCP declaration pin and an appended PATH do not pin the executed binary. |
| F13 | P1 | C9 D4, D6 | Bootstrap needs a precise trust chain, key roles and replay-resistant transcript. |
| F14 | P1 | C9 D1, D4–D6 | Forced-command and privileged-helper authority can escape the intended boundary. |
| F15 | P1 | C9 D5 | Framing, reconnect and cancellation lack enforceable state and resource rules. |
| F16 | P1 | C8 Security; C9 D3, Compatibility | OS-user and container backends do not already share a conformant execution contract. |
| F17 | P1 | C8 R6; C9 D7 | Stop, retire, revocation and identity destruction have incompatible semantics. |
| F18 | P1 | C9 D3–D5 | A project-held carrier credential does not authorize signing as the donor-held worker key. |
| F19 | P1 | C9 D5 | Mailbox delivery needs host admission, durable cursors and explicit disconnected behavior. |
| F20 | P1 | C10 Options, C2, Security | One Codex file is not one live refresher; the claimed lock is unsupported. |
| F21 | P2 | C10 C1–C4 | Bare homes, native-store sharing and node storage contradict one another. |
| F22 | P1 | C10 C3, C5, Security | The executor capability needs protected source binding and effective-auth qualification. |
| F23 | P1 | C10 Options, C6 | CIP-0003 safeguards and Decision 0017 gates cannot be silently superseded. |
| F24 | P2 | C10 C3–C4, Security | Expiry, error classification, resume and /login enforcement remain unqualified. |
| F25 | P2 | C10 C6–C7; C9 rollout | The proposed sequencing omits the broker needed for the owner’s MVP. |

## Evidence basis and citation key

The aliases below identify repositories, exact revisions and files. References in findings add the section, heading, symbol or diagram step. Private source documents and raw machine output are not republished here.

| Alias | Repository and revision | File or scope |
|---|---|---|
| C8 | curator-spec, 449a33dbd03e1519b17d3278a50af9713372e5d2 | cips/CIP-0008-remote-worker-launch-mode.md |
| C9 | same draft revision | cips/CIP-0009-donor-side-deployment-and-bridge.md |
| C10 | same draft revision | cips/CIP-0010-credentials-setup-token-and-inherited-auth.md |
| CD | same draft revision | cips/diagrams/donor-side.dsl; donor-first-deployment.puml; join-handshake.puml; remote-worker-turn.puml; remote-worker-retire.puml |
| MARKET | same draft revision | .research/261008_freelance-agents-marketplace.md |
| ENV | curator-spec main, 210d3619051248cf9b87a7bfe02e5a9142a6b7cf | protocol/environments.md, rc.14 context |
| D13 | same main revision | decisions/0013-execution-ownership-and-launch-plans.md |
| D17 | same main revision | decisions/0017-environment-credential-modes.md |
| D18 | same main revision | decisions/0018-curator-run-permission-interface.md |
| D19 | same main revision | decisions/0019-fragment-consumers-and-one-construction-site.md |
| D21 | same main revision | decisions/0021-sessions-enter-through-curator-run.md |
| C2 | same main revision | cips/CIP-0002-project-context-in-managed-launches.md |
| C3 | same main revision | cips/CIP-0003-claude-managed-home-credential-modes.md |
| C6 | same main revision | cips/CIP-0006-legacy-provider-settings-and-mcp-opt-outs.md |
| OWNER | wiki, 40bacd224ab5 | research/briefs/2026-10-08-curator-remote-worker-donor.md |
| ARCH | wiki main, a25f9d6edea174f5a99fcf7b274e40465ea6c7d4 | session-host/architecture.ru.md |
| CALL | same wiki revision | research/call-digest-2026-10-07.ru.md |
| AUTH | same wiki revision | research/harness-auth/RESEARCH.md |
| EXPIRY | same wiki revision | research/harness-auth/PLAN-expiry-check.ru.md |
| VIS | swarm-platform-architecture main, 35ca3c19dd1c1e5126ad9d03cb0f7a56bbef3ca4 | vision/VISION.md, v1.3 §0 |
| LCH | same architecture revision | vision/contracts/launch.md |
| ADM | same architecture revision | vision/contracts/admission-and-delivery.md |
| VD | same architecture revision | diagrams/plantuml/sequence/external-agent-onboarding.puml; remote-workplace-attach.puml; remote-worker-start.puml; mailbox-delivery.puml; keeper-sign.puml |
| RWH | remote-worker-harness, 71103c6cad46d6b05a7791bfb9f8820da8da5025 | docs/session-mode.md; docs/claude-mode.md; docs/muse-mode.md; docs/worker-mode.md; named implementation symbols below |
| RWP | remote-workplace, cb24ea04c223d4bd045e51f22e08fdad6ca9f74e | docs/protocol.md, v2.1 |
| CUR | curator, 75ab9a71a9b9049ec4242b4acdb60c3fead6aff1 | internal/envregistry/envregistry.go; internal/envprofile/managed.go; cmd/curator/main.go; cmd/curator/umbrella.go |
| HELP | direct scratch-home observations, 2026-10-08 | claude --help; codex --help; codex exec --help; muse --help; muse exec --help |
| PRIMARY | public primary documents checked 2026-10-08 | RFC 4035 §4.9.3; OpenBSD sshd(8), AUTHORIZED_KEYS FILE FORMAT; OpenAI Codex, CI/CD authentication, “Operational rules that matter”; Codex configuration reference, web_search and features.web_search entries |

D13 retains a proposed status and records the adopted D19/D21 amendments. Its preserved proposal is a design contract reference, not a reason to override the adopted decisions. The ENV text is the current normative baseline. AUTH and EXPIRY explicitly distinguish source inspection from outstanding live probes; this review preserves that distinction.

### Verification performed and limits

- Six read-only repository checkouts and retrieval of OWNER at its requested revision succeeded, each with exit 0. PR #134 head/base were checked. A fresh remote-main check matched CUR; source review therefore used current Curator main, not an assumed worktree baseline.
- All five HELP commands ran as standalone subprocesses with a scratch HOME and scratch harness/XDG configuration roots, an otherwise cleared environment with the executable search path supplied, and updater suppression. Each exited 0: 5/5 help probes. Codex printed a scratch-directory PATH-alias warning; help availability is the only property credited. No login, logout, authentication or model invocation occurred.
- Installed release identification was consistent with the C8 targets: Claude 2.1.293, Codex 0.159.0 and Muse 1.4.3. This used executable metadata/static identification, not a live conformance run. The Muse launcher’s updater-suppression branch was also inspected. A help option proves an interface exists, not that the effective launch disables the behavior.
- RWH’s documented live results and mutation checks are accepted as existing evidence, with their stated releases and modes. They were not rerun. New end-to-end lockdown qualifications in this review: **0/3 harnesses**. Help surfaces compared: **3/3 harnesses**. No Go build/test, credential-store access, packet capture or live authentication experiment was performed.
- Every requested source group, including all five VD diagrams, all CD diagrams and MARKET, was read. Findings below are either a direct textual contradiction, a source-backed implementation mismatch, or an explicitly identified unproven contract.
- Report checks: the standalone Python structure/privacy check exited 0 (25/25 findings, 7/7 question sections, required fields present); git diff --check exited 0. Pattern scanning is a supplement to manual review, not a proof that arbitrary secret content can be detected. No expected-failure gate is reported as passing.

## A. Per-harness lockdown and remote-auto

**Answer:** remote-auto is a sound separate permission policy if it preauthorizes only admitted remote tools and denies other approvals. It cannot be expressed as a bypass mode. The proposed table does not yet remove or contain every local capability at the named releases. R3 must define the effective configuration and process boundary; R4 must constrain caller input. Neither flag omission nor an empty working directory proves absence. [C8 R2–R4; HELP, Claude --restricted and --permission-prompts; RWH docs/session-mode.md, “Keeping built-in tools off the remote host”.]

### F1 — P1: distinguish generated policy from caller overrides

**CIP:** C8 R3–R5. **Contradicting evidence:** C8 R3 itself; HELP option descriptions; D19 Decision 1.

R5 refuses a lockdown plan containing any R4 argument. R3 necessarily emits those arguments: Claude --tools, --allowedTools, --mcp-config and --permission-mode; Codex -s, -a and restricted -c keys; Muse --approval-mode. A correctly composed plan therefore fails the literal intake rule.

R4 also leaves advertised local-effect paths unclassified. Examples include Claude --brief, --worktree/--tmux and prompt-file options; Codex -C/--cd, -p/--profile, --worktree, image and output-file options and configuration outside the listed -c keys; Muse --workspace, --worktree, --image, --prompt-file, --base-url and --approval-judge. HELP advertises these surfaces; this review did not demonstrate an escape through them.

**Concrete fix:** parse a closed, versioned caller-argument grammar before construction. Validate the fully expanded executable, arguments, configuration and plugin contribution against the canonical generated policy at intake. Generated restricted flags must be required, while caller replacements, aliases, duplicate flags, alternate sources and unknown options are refused. Test both a valid generated plan and each conflicting tail through the production launch entry.

### F2 — P1: restore the whole measured Codex recipe and choose a transport

**CIP:** C8 R3, R7. **Contradicting evidence:** RWH docs/session-mode.md, “Keeping built-in tools off the remote host”, “Known gaps”; internal/session/config.go, disabledFeatures, renderConfig and sanitizeCatalog; HELP, codex exec --ignore-user-config and codex --strict-config; PRIMARY, Codex configuration reference.

RWH removes multi_agent_version, empties experimental_supported_tools, clears supports_search_tool, disables 23 feature switches, disables request-user-input, and sets root web_search = "disabled". C8 copies only part of that recipe and uses the deprecated features.web_search switch. This does not reproduce the measured tool surface.

The measured adapter uses app-server and thread-level dynamic tools. C8 combines exec-specific switches, a Curator MCP profile, loopback app-server and supervisor approval handling without specifying the actual invocation or protocol. An exec --ignore-user-config launch also skips the base configuration file where RWH puts its hardening. --strict-config validates keys; it does not suppress other configuration sources.

**Concrete fix:** define separate qualified CLI/exec and app-server recipes, or ship only the measured app-server path initially. Bind the sanitized catalog, all feature controls, selected model and configuration sources into the policy. For MCP, prove remote calls work while native approvals are denied; denial of every app-server approval is not automatically an MCP preapproval contract. Preserve the dispatch tripwire as detection: it acts after dispatch. Treat missing tool-surface audit evidence as unknown/refusal, not an empty safe surface.

### F3 — P1: Claude’s suppressed settings cannot also enforce the policy

**CIP:** C8 R3, Security (“Hostile repository”). **Contradicting evidence:** HELP, Claude --restricted, --settings, --bare and --disable-slash-commands; C2, native-surface inventory and project-context design; C6, legacy-settings ownership; CUR envregistry.go, Claude Seeds.

--restricted ignores vendor user/project/local settings while retaining vendor managed policy and explicit --settings. A settings.json inside a Curator-managed home is not thereby vendor managed policy. R3 relies on that file to disable hooks and plugin sync without establishing that it is read. An empty cwd also does not establish absence of ancestor/home instructions, installed skills or native configuration. The statement that skills are only prompts is too broad for a surface that can include executable extensions.

**Concrete fix:** select one protected, explicit settings source; document precedence against actual managed policy and refuse incompatible policy. Disable and measure hooks, plugins, sync, instruction discovery and skill execution separately. Retain strict MCP even with an empty declaration set. Qualify the ordinary OAuth-token launch without --bare, since --bare suppresses OAuth. The print-only --permission-prompts none branch and the interactive dontAsk branch need distinct vectors; never spell bypassPermissions with --restricted.

### F4 — P2: Muse’s measured continuity and budget source are session logs

**CIP:** C8 R3 Muse row; C9 D1 supervisor guards. **Contradicting evidence:** RWH docs/muse-mode.md, “Shape”, “Settings” and measured read-boundary discussion; HELP, muse exec --no-session-log.

RWH launches exec once per turn with a stable session identifier and relies on persisted history. Its usage accounting includes main-turn and reminder-subagent log records. C8 disables session logs but C9 imports those budget and multi-turn guarantees. No replacement history/usage channel is specified. The older serve-mode evidence cannot establish a notification stream for the installed release.

**Concrete fix:** either retain protected, quota-bounded logs outside the untrusted process’s authority where feasible, with redaction and lifecycle rules, or qualify an alternative source of history and usage before claiming those guards. State any stateless limitation. Independently keep OS read confinement mandatory: RWH measured reads outside --workspace and access to the login copy inside the permitted sandbox; output redaction is not protection against encoded secret content.

### F5 — P2: pin a qualified release tuple and describe local activity honestly

**CIP:** C8 Current state, R3–R4, R6–R8. **Contradicting evidence:** RWH docs/session-mode.md, release-specific tests and known gaps; docs/claude-mode.md and docs/muse-mode.md; CUR envregistry.go, CodexSeedRevision and Registry; CUR managed.go, gatherSeeds and buildFragment.

The Claude row says “≥ 2.1.293”, Codex says “0.159”, but R4 rejects unknown releases. Current Curator still selects Codex seed revision A, records older verified harness versions, and has no Muse MCP adapter in the registry. RWH evidence for other releases is not qualification of these Curator launch paths.

RWH also isolates native discovery with a scratch HOME, whereas a Curator launch preserves HOME. A dedicated donor account may make that safe, but the tested source inventory must be rebuilt for that account; a scratch-home help result does not establish it. [ENV §10.2; LCH §8; RWH docs/muse-mode.md, “Tool isolation”.]

**Concrete fix:** qualify an exact tuple of executable digest, release, platform, transport, model/catalog, settings policy and sandbox revision. Turn updates into an explicit requalification step. A write-denied binary location does not prevent updater checks, downloads or another launcher path. Change “nothing can act on this machine” to a precise promise about permitted local runtime/state and remote tools. C8 R8’s limits are correct; its banner must reflect them.

The local-capability coverage is:

| Surface | Evidence and R3/R4 coverage | Required disposition |
|---|---|---|
| Updaters | Claude R3 names suppression variables, explicitly pending verification. Muse R3 only denies binary writes; the inspected launcher has MUSE_NO_AUTO_UPDATE branches. R4 is not an updater boundary. | Pin the actual launcher/binary path; disable, observe and confine update activity. |
| Telemetry/prefetch/network | Claude R3 lists variables; RWH Codex renderConfig also disables analytics. Blocking model tools does not block harness-originated traffic. | Measured provider/endpoint allowlist and trusted proxy/firewall; no claim of zero traffic from flags alone. |
| Logs/state | Claude persists transcripts; Codex --ephemeral changes persistence; Muse --no-session-log affects continuity and accounting. | Explicit permitted writable roots, quotas, retention and recovery per mode; see F4/F24. |
| MCP/settings/hooks/plugins | Claude retains managed/explicit settings; Codex recipe omits RWH hooks/plugins controls; current Curator seed A can preserve native MCP. | Effective-source inventory and qualified empty-MCP behavior, not just a declaration pin. |
| WebSearch/hosted tools | Claude --tools empty is the intended removal; Codex needs the full feature/catalog recipe; Muse --disable-web-tools is advertised. | Audit model-visible tools and block independent network paths. A provider-hosted tool still violates an “only remote-tools” tool set. |
| Context/file reads | Empty cwd is not a native-source inventory. Muse retains read/search; image/prompt-file/worktree options create further surfaces. | Closed argument policy plus OS read boundary with explicit runtime exceptions. |
| Harness shell-outs | Auth helpers, hooks, updater launchers and other internal subprocesses are distinct from model tools. RWH Claude evidence itself uses an OS credential helper. | Include child execution and effective helpers in the sandbox and supply-chain model; absence remains unproven until measured. |

## B. Specification placement and launch ownership

**Answer:** posture belongs in machine configuration under ENV §12.1, with a profile binding. It must not become executable policy supplied by package bytes under §10.3. A third permissions value is sensible under D18. Posture belongs in a versioned launch contract and receipt, not the profile lock hash. That placement still requires the amendments below.

### F6 — P2: extensions are not fragment metadata, and additive is not universally compatible

**CIP:** C8 R5, Compatibility and Specification changes; C10 C3. **Contradicting evidence:** D13 Decisions 3.2, 3.3 and 6.4; ENV §§10.1–10.3; C2/C3/C6, schema-change sections.

D13 §6.4 defines launch-plan extensions, not a closed fragment metadata object. Its outer plan shape is closed and version checked, while reverse-DNS extensions have an extensibility contract. A new top-level credential member does not fit the old outer schema. C8 alternately says the pin covers posture and correctly says only the plan does.

**Concrete fix:** specify the typed fragment change separately from the required plan-extension semantics. Make support for lockdown mandatory at intake; an old reader ignoring an extension cannot count as enforcement. Version the outer plan if credential is top-level, or define an explicitly negotiated extension. Coordinate the v4 changes proposed by C2/C3/C6/C8/C10 in one schema and migration table. Preserve profile-pin meaning and refuse unsupported mandatory capabilities precisely.

### F7 — P1: lock posture without granting system policy credential-selection power

**CIP:** C8 R6; C9 D1; C10 C2–C3. **Contradicting evidence:** ENV §12.2, permission direction and credential rule; D18 Decision 3; C3, “Configuration precedence and defaults”; ENV §10.3.

ENV currently permits a permissions lock only toward native. Locking remote-auto needs a named amendment. More seriously, C10 makes node:<label> lockable even though §12.2 prohibits locks selecting or constraining credential material. Allowing an environment entry in a package to name a node also crosses the operator-owned source boundary.

**Concrete fix:** amend the permission lock contract explicitly for validated lockdown plus remote-auto. Keep the account/source binding in operator-owned machine configuration. A package may declare a non-authoritative requirement but cannot select a credential. Preserve the no-source-selection lock rule unless a separate reviewed decision changes it; do not smuggle that change in through a label.

### F8 — P2: reconcile the new entry point with adopted composition ownership

**CIP:** C8 Compatibility, “curator run --mode remote-worker”; C9 D1 and new command group. **Contradicting evidence:** D19 Decisions 1–2; D21 Decisions 1–2 and Context; LCH §§1–2, §9; CUR main.go command dispatch and umbrella.go, cmdUmbrella/resolveProvider.

D19 puts all native spelling in agents-management, permits several fragment consumers, and excludes headless curator run. D21 says a session host consumes a composed plan and never composes. LCH §9 proposes changing the headless boundary; it is not evidence that the Curator decision has already changed. “Remote worker” also mixes target/posture with LCH’s independent interactive/headless and transport axes.

**Concrete fix:** adopt the needed decision amendment before implementation. Name the composer, fragment consumer and final process owner for local MVP and donor launches. Have the supervisor consume the qualified plan, or explicitly grant it a narrowly defined consumer role using the same construction site. Keep native grammar out of the Curator manager. Use a target/posture field separate from interaction mode; do not create a second flag compiler.

### F9 — P1: plan immutability does not enforce policy freshness

**CIP:** C8 R5 and Security (“TOCTOU”); C9 D2. **Contradicting evidence:** LCH §§7–8, final-executor binding and remote reconstruction; ARCH §§7.4–7.5, live grant/policy checks; C8’s own compose-once sequence.

If the executor checks only the old plan, a later machine-config change does not alter its digest. The claimed automatic refusal at resume has no causal path. A sandbox-profile digest similarly proves the requested bytes, not that a particular process is under that profile.

**Concrete fix:** bind plans to a trusted policy generation and check the current generation at final exec/resume. Specify invalidation, active-session stop behavior and rotation rules. Include the actual policy/backend capability in the executor’s acceptance record, with boot/setup qualification and negative enforcement evidence. Distinguish donor-supervisor assertions from remote attestation; the project cannot infer a hostile donor’s kernel state from a profile pin.

## C. Bridge, handshake, OS boundary and retire

### F10 — P1: the same UID is not the proposed trusted-server boundary

**CIP:** C9 D1–D2, D8 and Security; C8 R6 and Security. **Contradicting evidence:** ARCH §§7.2–7.5, separate keeper/broker identities and peer credentials; RWH docs/claude-mode.md, same-user Keychain trust anchor; docs/muse-mode.md, readable sandbox login material.

Harness, MCP server, supervisor and possibly proxy share one UID. UID-based firewall rules cannot distinguish their sockets by which process deserves a destination; destination rules allow that destination to the shared UID. Socket ownership also cannot distinguish equal owners. The shared home permits the process to attack writable audit/state and, on the file fallback, read the private key. A stored secret in a keyring is not automatically a non-exportable signing/decryption service. D8’s assertion that the harness user cannot change any control state conflicts with its ownership of that state.

**Concrete fix:** separate the untrusted harness UID from a trusted bridge/signer/supervisor identity. Use a narrow authenticated local RPC with peer credentials and a minimal stdio relay; keep keys, lifecycle state, binaries and audit authority outside the harness’s write/read reach. Keep the donor’s stop operation independent. If this split is deferred, scope the first release to the weaker trusted-process posture and withdraw the compromised-process protection claim. The documented threat model requires the split before a strong donor guarantee.

### F11 — P1: make OS setup an executable capability, not an illustrative recipe

**CIP:** C9 D2, D4. **Contradicting evidence:** AUTH §E, Keychain execution-context limits; ARCH §7.5; RWH docs/muse-mode.md, measured Seatbelt boundary; ENV §7.4, writable native authentication state.

A hidden no-login account, a login-keychain item and a launchd user agent surviving logout need a specified bootstrap/session context. They are not established by the cited Muse process sandbox. The Linux example binds the managed home read-only while harness state and personal Codex refresh need writes; ProtectSystem is not a general read-confidentiality boundary. A read boundary cannot be “where available” for a harness that keeps local read tools. Literal denial of all reads outside home/scratch also needs explicit exceptions for the executable and runtime libraries.

**Concrete fix:** qualify one platform first with an explicit service domain, account lifecycle, credential-unlock strategy, protected executable/library roots, writable native-state roots and mandatory read confinement. Record both positive usability and negative file/network/IPC tests at the actual supervisor entry. Refuse unsupported hosts. Specify transactional provisioning and rollback after each failed step, without deleting a pre-existing account or relaxing policy.

### F12 — P1: bind the executable independently of the declaration package

**CIP:** C9 D1, D3–D4; CD donor-first-deployment. **Contradicting evidence:** ENV §2.2, canonical source allowlist, stdio command and manager boundary; CUR umbrella.go, activeProviderRevision and resolveRevisionA; C9’s own appended-PATH description.

A locked MCP declaration pins declaration content, not whichever executable wins PATH lookup. Appending a binary directory permits earlier shadowing. The proposed mcp_package_allowlist value is a package name, whereas ENV requires canonical source identities. ENV also says the manager does not install or launch the MCP server binary; donor provisioning must not silently reinterpret ordinary materialization as executable installation. Current umbrella revision A can warn rather than refuse a provider outside trusted roots.

**Concrete fix:** use canonical source allowlist entries and separately pin/verify trusted executable bytes at final launch, including the umbrella provider and socket relay. Give the launcher a protected resolution domain with no writable predecessor. Put installation in an explicit operator-owned provisioning contract, or amend the manager boundary deliberately. Record release provenance and fail on changed, unreadable or shadowed executables.

### F13 — P1: specify the bootstrap transcript and actual cryptographic roles

**CIP:** C9 D4, D6, Security; CD join-handshake and donor-first-deployment. **Contradicting evidence:** ARCH §§7.2, 7.10; ADM §§1–2; VD external-agent-onboarding; PRIMARY RFC 4035 §4.9.3.

The human fingerprint comparison is valuable and must remain mandatory. A self-signature proves possession, not that the possessor is the intended donor. “20+ characters” does not define random entropy; the invitation code is a one-use bearer bootstrap capability despite the “no secret in the invitation” statement.

The package is signed by an operator but verified against a root fingerprint without spelling the operator certificate chain, scope and validity. The same unnamed key pair appears to serve signing, SSH and X25519 decryption. The draft does not define algorithm-specific keys, their binding, or whether the chosen keyring can perform those operations without exporting material. A client-supplied signed nonce is not a server challenge or an atomic invitation-consumption protocol.

The diagram’s DNS AD-bit check is insufficient unless it comes from a trusted validating resolver over a protected channel, or validation is local. DNSSEC, SSHFP and TLSA need explicit secure/bogus/indeterminate handling, service/algorithm matching and rotation semantics.

**Concrete fix:** define a versioned, domain-separated canonical transcript binding invitation identifier, project, role, both parties’ keys, server challenge, expiry, algorithms and package digest. Use cryptographically random codes, attempt limits, atomic claims and replay tombstones; retries may recover only the same authenticated enrollment. Use standard authenticated encryption and separately certified signing/encryption/SSH roles. Require a root-validated operator chain and an independently authenticated fingerprint channel. Specify trusted DNS validation or choose an explicit out-of-band pin bootstrap for the first release. Never downgrade on validation failure.

### F14 — P1: protect SSH restrictions and bound the root helper’s inputs

**CIP:** C9 D1, D4–D6. **Contradicting evidence:** CALL C9; ARCH §§7.4–7.5; PRIMARY sshd(8), AUTHORIZED_KEYS FILE FORMAT; RWP §9, tool authority boundary.

Donor-to-project SSH is the correct direction. restrict and a forced command are useful, but if the workplace user can edit its own effective authorized_keys, a remote shell tool can add an unrestricted key. A bare forced-command name also needs trusted resolution. Restricting sudoers to one executable is not enough if that helper accepts arbitrary users, paths, policy text or deletion targets.

**Concrete fix:** keep effective SSH authorization and command configuration outside workplace write authority; disable unneeded forwarding, agent/X11 access, PTY and user startup/environment hooks explicitly in the qualified server policy. Execute an absolute, protected bridge binary. Give the helper a closed operation schema, authenticated caller/grant binding, fixed templates, allocated account identifiers and safe path handling. Test symlink/path substitution, arbitrary-user deletion, argument injection and partial setup recovery. Recheck host pins on every reconnect; never add trust-on-first-use fallback.

### F15 — P1: define framed-channel state and failure semantics

**CIP:** C9 D5. **Contradicting evidence:** RWP §§6, 9 and 14; ADM §5; ARCH §7.9, Remote Bash lifecycle; RWH docs/worker-mode.md, budget and process guards.

Length-prefixed JSON plus request IDs, deadlines and output caps does not define a safe protocol. Missing pieces include a bounded frame before allocation, input caps, canonical field/types and duplicate-key handling, binary encoding, version negotiation, principal/connection-scoped IDs, sequence rules, backpressure and method authorization. A caller’s timeout/output values cannot widen server policy. cancel/stdin need ownership and live-job checks; env_allow is not permission to inject arbitrary environment variables.

Reconnect also creates an uncertain-result problem: an exec or mailbox post may have taken effect before its reply was lost. Killing an SSH child is not proof that detached descendants have stopped.

**Concrete fix:** specify a closed protocol and authoritative server minima/caps, per-session quotas and revocation checks. Scope cancellation to a job owned by that binding. Give side effects stable idempotency identities or report an unknown result without replaying them. Use process-group/cgroup or equivalent lifetime ownership plus an independent lease deadline, and prove descendant cleanup on transport loss. Version rejection, truncated frames, duplicate IDs, stale cancels, oversized input/output and reconnect-after-effect need negative vectors.

### F16 — P1: name the backend contract instead of borrowing container conformance

**CIP:** C8 Security layer 5; C9 D3 and Compatibility. **Contradicting evidence:** RWP §1 and §9, including §9.2 and §9.11; RWH docs/worker-mode.md, workplace policy and review handoff; VIS §0 and CALL C9; VD remote-worker-start.

RWP v2.1’s named guarantees belong to its workplace path, including its execution API, policy intersection, audit/export and cleanup. It explicitly distinguishes execution outside that path. A new OS-user Remote Bash executor does not inherit those guarantees because its tool names are similar. C9 exposes all six tools, while RWH/RWP intersect tools with policy; a read-only worker must not automatically gain bash/write/patch.

There is also an upstream inconsistency: VIS v1.3 and CALL C9 choose an OS-user workplace, while remote-worker-start still depicts a container path. This is a source reconciliation issue, not proof that either backend is already interchangeable.

**Concrete fix:** choose and name the first backend; specify its equivalent required guarantees and explicit differences. Separate mailbox control methods from the remote tool set, intersect remote tools with server policy, and retain policy identity, audit, output review/export and cleanup semantics. Map an adapter to RWP only after qualification; do not label arbitrary direct execution a conformant workplace. Update the stale architecture diagram when that decision is adopted.

### F17 — P1: separate stopping work from destroying the worker identity

**CIP:** C8 R6; C9 D7; CD remote-worker-retire. **Contradicting evidence:** ARCH §§7.2, 7.4 and 7.9; VD remote-workplace-attach; ADM §2, key revocation/freshness.

C9 deletes the donor keyring item yet says the same pinned worker identity can return. That requires retaining the key or a separately authorized rotation, not deletion. C8’s deletion of a donor carrier identity also contradicts C9’s no-carrier-credential design. Removing an SSH key prevents a new login but does not by itself stop an established channel or its jobs.

**Concrete fix:** define separate transitions for pause/stop, workplace retirement, provider-credential removal, trust-key revocation and destructive purge. Persist stop/revocation before service restart can race it. Revoke grants and channel authorization, close existing sessions, terminate all owned descendants, then archive only permitted noncredential artifacts and release the account/firewall state after process absence. Preserve identity for return or enroll a new key through a root-authorized rotation. Donor purge must communicate the revocation outcome; loss of the channel alone is not evidence of remote deletion.

## D. Mailbox and carrier identity

**Answer:** keeping the worker’s carrier credential exclusively in the project bridge is consistent with CALL C2/C9, VIS M3 and ARCH §7.10. The donor should access only its scoped mailbox through the tool/control channel. This does not move ownership of the external worker’s signing key. [ADM §§1–2, 4–7; VD mailbox-delivery and keeper-sign.]

### F18 — P1: carrier credentials and trust signatures are different authorities

**CIP:** C9 D3 mailbox_post, D4 key generation, D5 tasks/results; CD remote-worker-turn. **Contradicting evidence:** ARCH §7.2, external-key ownership; ADM §§1–2 and §7; VD external-agent-onboarding and keeper-sign.

The donor-generated worker key is pinned and said never to leave its keyring. The project bridge then signs outbound envelopes “with the worker’s identity”. A project-held carrier account cannot produce that worker-key signature. The internal keeper flow does not silently delegate an external private key. Audience projection after signing would also change the signed bytes.

**Concrete fix:** choose either donor-side signing of the exact final projected envelope through the trusted local signer, or an explicit root-authorized project-side delegate identity with a visible, scoped delegation chain. Keep carrier authentication distinct in the schema and audit. Do not represent a bridge assertion as a direct worker signature. Test alteration after projection and use of a valid carrier credential without signing authority.

### F19 — P1: delivery needs admission and recovery, not only notifications

**CIP:** C9 D5 tasks/turns and mailbox frames. **Contradicting evidence:** ADM §§3–6; ARCH §7.10, durable mailbox/cursor ordering and three entry channels; LCH §6; VD mailbox-delivery.

“An admitted mailbox item becomes the next turn” collapses authenticated data and permission to inject a session turn. ADM requires host-side admission and binding even after upstream verification. A signature or post grant alone does not authorize arbitrary user-turn delivery. A mailbox notification should not become a command into the donor.

When SSH is down, the project mailbox can remain durable, but donor notification, reads, posts and remote tools are unavailable. The draft specifies neither durable cursor recovery nor what the supervisor does with a still-running harness. Replaying an uncertain outbound post or treating a pipe write as applied delivery breaks the delivery contract.

**Concrete fix:** carry the verification/admission record and have the donor host validate target binding, expiry, revocation/fence and delivery mode. Deliver authorized control turns through the qualified session-input adapter; expose ordinary data through mailbox tools. Keep notifications bounded and advisory. Define pause/deadline behavior on disconnect, durable high-water cursors with acknowledgement after persistence, reconnect catch-up, deduplication and an application receipt distinct from transport receipt. Preserve pending proposals with stable IDs; never fall back to giving the donor a carrier credential.

## E. Credentials, nodes and executor capability

**Answer:** a protected reference resolved at final exec is the right general design. A Claude setup-token is a useful enrollment path, and excluding --bare is correct. The proposal must retain distinctions between enrolled external authorization, native refresh state, native-store isolation and account policy. An environment-delivered bearer token remains readable by the harness; it is not a non-extractable lease. [AUTH §§A–B, E; C3 Recommendation and runtime rules; LCH §7.]

The recorded Claude native Keychain service is scoped by CLAUDE_CONFIG_DIR; sharing an OS Keychain does not make two managed homes share an item. An enrolled setup-token avoids copying that native refresh family, but has its own lifetime and manual vendor-revocation limits. Codex’s in-place file writer preserves link identity; it says nothing about concurrent refresh ownership. These are separate facts, not interchangeable authentication strategies. [ENV §7.4; AUTH §§A–B, E; D17 Q1/Q7.]

### F20 — P1: one file does not serialize Codex refresh

**CIP:** C10 Options for Codex, C2 personal row and Security (“Rotation hazards”). **Contradicting evidence:** RWH docs/session-mode.md, “Login: the single auth owner”; internal/authowner/owner.pl, flock protocol; AUTH §§B, E; ENV §7.4, per-harness writers; PRIMARY, Codex CI/CD authentication, operational rules.

RWH explicitly records that multiple app-servers sharing one auth.json can race refresh. Its solution uses one refresh owner, an explicit auth.json.rwh-lock, external access tokens for tenants and no tenant auth.json. Its documented eight-caller test produces one refresh with the lock and eight without it. This is stronger and directly contrary evidence to C10’s “one file = one live refresher”. The .auth.json.lock mentioned in C10 is not evidence of native Codex serialization; ENV associates that lock shape with Muse.

**Concrete fix:** either serialize all refresh-capable personal-account launches with a real account lease, including native clients, or qualify RWH’s single-owner/external-token app-server adapter at the chosen release. Preserve the in-place file-link liveness checks, but do not count them as concurrency control. PLAN step 6’s file/copy experiment is not this race qualification. The older RWH live result is reusable evidence, not a new 0.159 qualification.

### F21 — P2: retain separate source-mode and store-isolation axes

**CIP:** C10 C1–C4 and Compatibility. **Contradicting evidence:** ENV §7.4; D17 Q4/Q5/Q7; C3, “Configuration precedence and defaults”; AUTH §§B, E.

C2 defines bare as no seed and no link, then mandates a Codex auth-file link. C1 and Security say material lives only in a keyring except where unavailable, but the personal Codex row needs a writable file even on a keyring-capable platform. --from-native also cannot assume an active native file exists when the effective store is keyring. A new isolation value masks these different policies.

**Concrete fix:** keep isolation as native-store sharing, with a separate credential mode/source binding and a harness-specific storage/channel table. Represent Codex’s shared native refresh store honestly, or use the token-only tenant alternative. Specify absent versus unreadable/native-keyring cases with no silent export or fallback. Adopt changes through inspect → plan → apply migration and a coordinated marker schema, preserving existing homes until explicit adoption.

### F22 — P1: capability negotiation is necessary but does not authorize a secret

**CIP:** C10 C3, C5 and Security; C8 Security credential claim. **Contradicting evidence:** LCH §7; C3 Recommendation, launch selection and runtime rules; AUTH §§A–B, E; ENV §10.3, passable_env_names scope; D13 Decision 6.3 and open environment-filtering question.

credential-injection/1 is a good versioned gate. A self-declared capability or caller-supplied source_ref is not a protected authorization binding. Higher-precedence inherited auth variables, helpers or provider endpoints can override the chosen node or send authorization to the wrong destination. passable_env_names bounds the MCP declaration’s named pass-through; an empty list does not sanitize all inherited harness environment or every MCP child.

Per-launch environment injection also gives the token to the process. A compromised harness can copy it or pass it to a child. C8’s “must not carry a credential it can exfiltrate in bulk” is not established by eliminating a store copy.

The stronger “no Keychain item is read” assertion in C10 C2 also needs the effective-source check proposed in EXPIRY step 2; choosing an environment channel alone is not evidence that every native lookup was suppressed.

**Concrete fix:** re-resolve a protected binding at final exec for executor peer identity, profile, harness, account, channel, endpoint, approved executable and policy generation. Require behavioral qualification behind the capability. Reserve injected names, reject conflicting sources, separate harness and MCP-child environments, and distinguish absent/denied/expired/malformed reads without fallback. Describe the owner-approved initial trusted-process credential posture accurately; stronger non-extraction requires a separately approved mechanism, not an environment variable.

### F23 — P1: preserve CIP-0003’s gates and amend no-copy policy explicitly

**CIP:** C10 Options for Codex, C6 and Compatibility. **Contradicting evidence:** C3 Current state, capability table, Recommendation and item 5; D17 Q1/Q4/Q7; AUTH §§A–B, G and “Confidence and gaps”; EXPIRY step 6.

“Claude cannot share a file on macOS at all” is stronger than the evidence. C3 records a plaintext fallback and synthetic linked-file reads, while correctly withholding refresh-safety qualification. Its legacy Linux handling and read-order/refresh gate cannot be discarded on that premise. Replacing item 5 is appropriate only once an executor actually implements the protected runtime contract.

Codex keyring identity being derived from CODEX_HOME is supported by source inspection in AUTH, while D17 Q4 retains its older assumption pending the specified probe. That discrepancy needs an explicit revision and evidence label. The stripped-copy option also remains prohibited by D17 Q7 without a new consent policy, despite being discussed as a workaround in AUTH.

**Concrete fix:** adopt C3 and C10 together with a clause-by-clause disposition: retained defaults, precedence, protected-store reads, migration, no fallback, live qualification and legacy behavior. Record per-home keyring source evidence without claiming the pending live probe passed. For any native credential copy, separately amend Q7 with operator-owned profile/environment, source/destination roles, purpose, reason and expiry. Enrollment of a deliberately supplied setup-token is not permission to extract an existing native login.

### F24 — P2: qualify error and recovery behavior instead of declaring it

**CIP:** C10 C3–C4 and Security (/login); C8 R3 persistence switches. **Contradicting evidence:** AUTH §E, lifecycle; EXPIRY steps 2–4; C3 runtime rules; HELP, Codex --ephemeral and Claude print-mode options.

Enrollment time is not necessarily a setup-token’s mint time. A token imported late in its lifetime cannot be reported as fresh for another year. Opaque TLS traffic does not give the executor HTTP status automatically, and a 403 is not uniquely revocation. EXPIRY explicitly proposes measuring the observable error and resume path; those measurements are not results.

C8’s ephemeral Codex mode and Muse log suppression also conflict with a generic native --resume promise. Finally, /login is an input handled inside a running interactive harness, not a native argument that the launcher’s R4 table can intercept.

**Concrete fix:** distinguish known expiry, supplied mint metadata and unknown expiry. Define and qualify typed harness error classification, preserving uncertainty when permission/quota/auth causes cannot be separated. Park and notify without indefinite retry; verify context recovery for each actual persistence mode. Constrain admitted session inputs or ship only a mode where auth commands are unavailable; do not promise a launch-time flag table enforces every later interactive command.

## F. MVP, first donor deployment and cuts

### F25 — P2: sequence the local broker before claiming the MVP is unblocked

**CIP:** C10 C6–C7; C9 Implementation plan. **Contradicting evidence:** CALL C15 and plan §4; ARCH §§7.2–7.5; AUTH §E, different OS user; LCH §7.

The owner’s first MVP is one subscription, one machine, a chain of OS users, Claude/Codex, certificate admission and mailbox delivery, without a remote harness. C10 says it unblocks that MVP while explicitly excluding cross-user leases. A per-user node cannot, by itself, deliver authorization across those user boundaries. Cross-user delivery also does not inherently authorize concurrent use or credential copying.

**Concrete fix:** give the minimal peer-authenticated broker/final-executor interface its own prerequisite slice. Prove one authorized chain end to end before generalizing node kinds or deploying a donor. Keep personal Codex concurrency serialized unless the qualified single-owner adapter is deliberately chosen.

The minimum useful milestones are:

1. **One-host MVP:** separate keeper, broker and worker identities; root/admission/grant checks; one protected credential channel per harness; final-exec source binding; a composed-plan host; durable mailbox and exact-byte signatures; one task spanning the user chain. Check rejection of wrong-peer/wrong-account/expired requests and preserve explicit human enrollment. [CALL C15; ARCH §§7.2–7.5; LCH §§6–7; ADM §§2–6.]
2. **First donor:** one qualified OS and harness tuple, trusted local bridge/supervisor split, donor-owned stop control, bounded egress, protected executable resolution, enrollment transcript and SSH restrictions, one backend, tool-policy intersection, budgets, disconnect recovery and retire. Verify these at the real init/launch/tool/stop paths, including failure halfway through provisioning. [C8 R7; C9 D1–D8; RWH docs/worker-mode.md; RWP §9.]
3. **Acceptance evidence:** measure the visible tool set against the complete permitted set, alongside enforcement negatives. Unknown audit coverage, unknown releases, a failed settings read or missing sandbox support must refuse. Narrow the gate in mutation checks so a broader unauthorized class demonstrably fails. Reuse only the stated portion of earlier RWH evidence. [C8 Test plan; RWH docs/session-mode.md, layer table and known gaps.]

**Cut from the first delivery:** marketplace/reputation/price mechanics, federation rollout, generic multi-provider node graphs, enterprise-token paths outside the owner’s first slice, multi-OS parity, and automatic backend interchangeability. Defer Muse donor support if its read boundary and persistence/accounting cannot be qualified promptly. An out-of-band pinned bootstrap can reduce first-release DNS/DANE implementation scope if adopted explicitly; it must preserve authenticated key comparison and replay protection. Do not cut grant enforcement, independent donor stop, signing-key ownership, credential-source checks or job cleanup. MARKET’s future-market discussion provides no measured evidence that these trust or subscription constraints disappear.

## G. Decisions that are sound and should stay

- **Machine-owned posture, package-owned context.** Keep posture under ENV §12.1 and preserve §10.3. Record effective launch policy separately from the profile pin. [C8 R1/R5; ENV §§10.3, 12.1; LCH §8.]
- **A distinct remote-auto policy.** It should preapprove only the authorized remote tool set, deny everything else that would prompt, and never request native bypass. Unknown capabilities/releases should refuse. [C8 R2/R4; D18; HELP, Claude --restricted.]
- **One native grammar construction site.** Centralize harness mapping in agents-management and test the actual consumer/executor paths. [C8 R3/R7; D19 Decisions 1–2.]
- **The donor remains the machine authority.** No launch flag protects against its owner; the donor needs independent visibility and stop/revoke control. OS confinement is the boundary for the untrusted process. [C8 R6/R8; ARCH §7.5.]
- **Donor-initiated SSH and mailbox through the bridge.** No project-initiated general command surface on the donor, no donor carrier credential, no model-facing enrollment/key/config tool, and no direct board access. Preserve the distinction between transport messages and authorized session turns. [C9 D3–D6; CALL C2/C9; ARCH §§7.9–7.10; ADM.]
- **Human-confirmed enrollment and pinned public identities.** Generate donor private keys locally, perform independent fingerprint confirmation, grant narrowly with expiry, and retain explicit revocation. The protocol details need F13’s corrections; the direction is sound. [C9 D4/D6; ARCH §7.2; VD external-agent-onboarding.]
- **Secrets resolved only at the final executor.** Plans, fragments, receipts and node records should contain nonsecret references and metadata. Capability refusal is the correct rollout default until the executor is qualified. [C10 C1/C3; C3 item 5; LCH §7.]
- **Claude setup-token enrollment without native-store extraction.** Preserve per-home interactive behavior, exclude --bare for OAuth, and keep vendor enrollment/revocation with the human. Do not reinterpret this as a short-lived exchange capability. [C10 Options/C2/C4; AUTH §§A, E; D17 Q7.]
- **No provider interception in the initial design.** Keep the owner’s chosen initial credential posture and bounded concurrency explicit; a node does not multiply entitlement or prove non-extraction. [C10 C7; ARCH §7.3; AUTH §§E–F.]
- **Reuse measured RWH guards with their actual dependencies.** Keep allow-from, rate/token/lifetime controls, tool-surface observation and audit, while naming their transport, log and single-auth-owner assumptions. [C9 D1; RWH docs/worker-mode.md; docs/session-mode.md; docs/muse-mode.md.]

The report is ready for review. Proposed fixes are recommendations; no policy amendment, live-security qualification or product acceptance is implied.
