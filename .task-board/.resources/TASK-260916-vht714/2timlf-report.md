# TASK-260916-2timlf — Credential and permission modes

Research date: 2026-09-16; e11-1, Darwin 24.6.0 x86_64. Ready for review. Research only; no repository changes, credential writes, logins, real ax calls or bypass sessions.

## 1. Current state: environment × platform and e11-1 observations

**Key findings:** credential modes already exist as environments.isolation.<profile>.<env-id>. Current macOS Claude deliberately refuses shared. Native Pi's default auth root differs from Curator's resolver. Tracked native bypass is forbidden by the current launcher/ax contract. Credential migration has two inspection-derived hazards.

Citation roots are read-only local snapshots under /Users/administrator/Developer/ReluxWorks/:

| Prefix | Checkout | HEAD |
|---|---|---|
| C | curator/curator | 18f05497f6ad4db243279de643535716bbea6ec6 |
| S | curator/curator-spec | a68854d54725862f6f696019ef2e569f3ec29cd6 |
| L | curator/curator-agent-launcher | b34e1e27dbe97155682ce013948a0cc226280844 |
| A | skill-agents-management | 7f0b6bc69c4388841e4aea17309416e9c1393ab5 |
| I | relux-agents-infra-main | 459742ea67e3c6b84169520b92d74fe7f73e3002 |

B5 denotes board outcome TASK-260908-yl5x3k_onboarding-evidence-rev2.md under C .task-board/.resources/TASK-260908-yl5x3k/. H denotes attached TASK-260916-2timlf_native-help.txt, with command names, exit codes and numbered output. Citations establish source behavior or prior observations, not newly executed platform tests.

**Spec-location correction:** S profiles/manager.md:1008 §7 is source audit. Managed homes/credentials are now manager §12, especially §12.4 (:2343). Adapter §7 and knob §12.1 are in S protocol/environments.md:944,2240.

| Environment/platform | Store and strategy | Mode and evidence bound |
|---|---|---|
| Claude / macOS ≥ recorded 2.1.261 | Keychain service Claude Code-credentials, suffixed for CLAUDE_CONFIG_DIR; JSON fallback under config dir. Darwin passthrough empty; generated .claude.json seed only. | Isolated default; shared refused. C internal/envregistry/envregistry.go:193-209,388-408; S protocol/environments.md:1074,1101-1123. Installed 2.1.273 help verified; suffix algorithm accepted pinned evidence, not re-extracted here. |
| Claude / older macOS | Same store family; older directory scoping not established. Registry still links nothing. | Code defaults/permits shared, refuses isolated. Compatibility branch is not proof of sharing. C registry:388-408. |
| Claude / Linux | Native ~/.claude/.credentials.json linked under shared; isolated omits link. | Shared default; isolated available. Refresh write behavior unverified/expected to detach, S environments:1074-1099. No Linux runtime validation here. |
| Claude / Windows | Native user-profile .claude/.credentials.json per vendor docs; no Windows/default passthrough entry. | Code falls through to shared default and accepts isolated, but sharing is not implemented/verified. Windows reserved, S environments:1074; C registry:193-196,388-415. |
| Codex / macOS | ~/.codex/auth.json, relocated by CODEX_HOME. Exact native config keyring suppresses links; otherwise auth.json file-link. config.toml copied once. | Shared default, isolated omits links. e11-1 file mode worked. C registry:213-240; C internal/envprofile/managed.go:490-547,563-590. Cross-home keyring sharing not established here. |
| Codex / Linux | Same default row: CODEX_HOME auth.json or OS keyring. | Same code behavior; runtime/keyring sharing unverified here. C registry:225-228. |
| Codex / Windows | Same default row and auth-file role. | Runtime, symlink privileges, refresh and keyring scope unverified here. |
| Pi / macOS | Pi 0.84.2 documents ~/.pi/agent/auth.json, relocated by PI_CODING_AGENT_DIR. Curator resolves .pi and links ~/.pi/auth.json instead. | Shared default; isolated omits links. C registry:263-286; C internal/envprofile/switch.go:75-80; installed Pi docs/providers.md:26,111,139 and H. Root discrepancy matters before login diagnosis. |
| Pi / Linux | Same home-variable/default row. | Same discrepancy by inspection; no runtime verification. |
| Pi / Windows | Same default row. | Platform and symlink behavior unverified. |
| OpenCode / macOS | Auth in XDG_DATA_HOME, outside swapped XDG_CONFIG_HOME; no credential links/seeds. | Ambient shared; isolated refused. Absent in B5; launcher refuses environment. C registry:243-260,389-391; L internal/mapping/mapping.go:21-32. |
| OpenCode / Linux | Same adapter/default row; ambient XDG data auth. | Shared-only contract; docs-confidence, not host-tested. |
| OpenCode / Windows | No distinct Windows auth verification. | Same GOOS fallback; actual behavior unknown. Do not label XDG behavior tested Windows support. |

Managed homes are ~/.curator/environments/<profile>/<env>; OpenCode adds opencode/ below its parent (B5:1011-1014). Empty passthrough does not demonstrate authentication success.

[Claude authentication](https://code.claude.com/docs/en/authentication) documents macOS Keychain, JSON fallback on failed Keychain write, and directory-specific Keychain selection. [Codex authentication](https://learn.chatgpt.com/docs/auth) documents file/keyring/auto/ephemeral. Auto may use keyring although Curator links a file for it. Current docs do not prove every pinned binary. Installed Pi 0.84.2 docs/providers.md:26,139 documents its auth path and mode 0600.

### B5 evidence versus this run

- B5:796-806: Claude prompt **exit 1**, Not logged in; passthrough empty. This proves /login was required, not that an operator subsequently performed it.
- B5:808-840: Codex prompt **exit 0**, replied OK; link at :594. Approval never coexisted with workspace-write sandbox: no prompts is not unrestricted execution. Figma OAuth and skill-fixture warnings did not prevent the answer.
- B5:842-865: Pi prompt **exit 1**, no Anthropic API key; link to ~/.pi/auth.json, size 2. B5 calls it an empty object but says contents were not dumped. Size alone does not prove contents: recorded empty/two-byte store, no usable credential in that run.
- Fresh metadata-only Python check, **exit 0**: Claude passthrough [], seeds [.claude.json]; Codex auth.json file-link, seeds [config.toml]; Pi auth.json file-link, seeds []. Targets remain ~/.codex/auth.json and ~/.pi/auth.json. Native sizes: Claude JSON 509, Codex JSON 3864, .pi/auth.json 2, **.pi/agent/auth.json 127 bytes**. No auth contents read. The latter file does not establish account/provider/usability, but disproves a broad claim that no native Pi auth file exists.
- Fresh version output: Claude 2.1.273; Codex 0.153.4; Pi 0.84.2. Registry pins 2.1.261, 0.153.2, 0.84.2 (C registry:199,231,280). No login/refresh repeated.

### Seeds, backups, takeover and hazards

C managed.go:551-605 copies non-credential seeds once; absent skipped, unreadable stops before writes. Claude seed is generated and repair adds project trust entries (:623-630; S environments:1140-1166), never OAuth data. Copied Codex config may retain permission settings; launcher defaults alone cannot establish approval posture.

applyPlan backs up replaced regular managed-surface files, skips symlinks, refuses existing backup generations (:823-850). Provisioning conflicts require takeover (:1645-1666). Credential links are separate in finalizeMarker (:927-942), outside ordinary surface backup. **Takeover is not credential-copy consent.**

Two inspection-derived defect candidates need production-entry tests:

1. Shared → isolated: effectivePassthrough returns empty (:500-502); finalizeMarker only loops over new links then records empty (:929-942). Removal's recorded set comes only from surfaces (:1670-1677). The inspected path does not remove the old credential link. Later verification compares empty sets (:1343-1346), without checking unrecorded auth files. Sharing can remain behind an isolated configuration.
2. Repair/shared migration: finalizeMarker removes an existing wanted-link path before symlinking (:935-939). A detached regular or isolated credential can be unlinked without credential-specific preservation, contrary to S manager:2365-2368's leave-bytes-untouched promise. No live mutation reproduced this.

checkPassthrough checks link identity, not target existence/authentication (:1330-1357). Current marker is not an authentication attestation.

## 2. Credential-mode options: knob, strategy, security and migration

| Knob option | Tradeoff |
|---|---|
| Existing environments.isolation.<profile>.<env-id>: shared or isolated | **Recommend canonical v1 spelling.** Already parsed/resolved; separate accounts per profile. Describe credential-store sharing, not security sandbox. |
| New environments.credential_mode.<env-id> | Simple environment-wide default, but needs profile override layer. If added later, place below existing per-profile entries and version schemas. |
| Rename to environments.credentials.<env-id>.mode | Extensible, but needs compatibility/precedence/lock migration; unnecessary now. |
| Per-run credential flag | Risks implicit auth migration of existing home. Defer; launch should not silently change credential ownership. |

Existing lock rules: S manager:51-73; S environments:2285-2297; C internal/config/environments.go:570-598,796-816 permit system isolation map **only toward shared**. A lock replaces the whole profile/env map; unlocked operator map wins whole over system default. Not permission for fleet credential selection. Shared lock on current macOS Claude must still fail unsupported. Enforced isolation requires a reviewed policy change; do not call credential selectors universally lockable.

| Strategy | Applicability and consequences |
|---|---|
| file-link | Codex file mode, Pi after root correction, Claude Linux provisionally. No manager secret copies. Logout/account changes affect every home. In-place writes preserve link; rename-over detaches and requires non-destructive conflict handling. Missing vs unreadable differs. Never copy credentials as symlink fallback. |
| keychain-shared / shared keyring namespace | Only with verified native credential selector independent of managed config. Same Keychain does not mean same item. Claude selects by config dir; Codex cross-home scope needs proof. No token export/injection or HOME workaround. |
| copy-at-provision with operator waiver | Bootstrap, **not ongoing sharing**. Copies diverge/refresh tokens may invalidate each other. Needs tool idle/lock, private destination/recovery, no reseeding or store/log/hash of secrets. Current spec forbids it; cannot treat credentials as existing seeds. |
| isolated | Fresh supported home, no link, fresh login. Same OS user can access other files. Ambient keys/helpers/cloud credentials and auth selectors may defeat account separation. OpenCode config swap cannot isolate auth. |

Installed Claude binary contains CLAUDE_SECURESTORAGE_CONFIG_DIR at byte offset 80210816 (read-only search exit 0). This proves **string presence only**, not semantics/precedence/support. Full contiguous Claude Code-credentials was not found. SHA-256 path suffix (dash plus first eight hex characters) remains accepted pinned evidence S environments:1074; vendor docs independently confirm directory scoping. Validate a supported override before shared admission; never export Keychain secrets into JSON.

secret_material_waivers is **not credential-copy consent**: S environments:1566-1582,2267-2278 waives an audited member pin/file/byte-span finding. It does not authorize extraction or revise manager §12.4. Future copy consent needs separate operator-owned profile/env, source/destination roles, purpose, reason, expiry; no token values/digests. Explicitly revise the no-copy boundary.

passable_env_names (S environments:2130-2135,2264) bounds MCP env-name passthrough, not all inherited provider credentials. L internal/composition/composition.go:52-104 includes inherited environment and overlays. It is not a universal isolation filter. Strict account isolation needs policy for auth variables/helpers/keyring/native config, refusing unsupported backends.

Recommended migration: explicit inspect/plan/apply, separate from resolve --repair.

1. Under manager lock inspect prior marker, store roles, backend/version, topology and requested mode. No secret reads for path classification. Unknown/unreadable refuses, not absence. Include both Pi roots; sizes cannot select account.
2. Preserve effective mode on upgrade. Do not silently change Claude account or Pi link. Missing old mode metadata requires inventory, not inference from empty passthrough.
3. Shared → isolated: unlink only recorded symlink still targeting declared native store; preserve native bytes. Refuse detached/unrecorded conflicts. No secret copy; fresh login; reconcile inherited selectors before isolation claim.
4. Isolated → shared: refuse conflicting credential files/Keychain state. Operator chooses account/disposition or fresh home. General repair must not delete isolated credentials.
5. Backend/root change requires explicit plan/capability verification. Native config.toml cannot represent managed config after provisioning. Define auto/ephemeral; current exact-keyring regex does not.
6. Atomically publish non-secret mode/strategy/source-role/version/provenance after reconciliation. Verify absence and presence; protect rollback and journal. General backups must not follow auth symlinks/archive credentials.

## 3. Permission interface: mappings, refusals and provenance

L internal/cli/cli.go:31-70,193-212,215-269,296-307 accepts profile, system-prompt, model, effort, name, ax-profile, then preserves native args after --. Repeated/unknown flags fail usage. ax-profile standard|yolo controls tracking; untracked usage fails. L README:113-132; SPEC:141,623-629 forbids deriving ax yolo from native argv.

A is the Go module, not legacy wrapper implementation. Interactive builders omit bypass (A pkg/agentic/systems/claude/args.go:56-58; codex/args.go:44-46). Legacy source found in I:

| Legacy wrapper | Expansion/evidence |
|---|---|
| agents-infra claude -d / --danger / --yolo | --dangerously-skip-permissions. I tools/agents-infra/internal/infra/claude_launch.go:12,419-427. agents-infra claude --print-config -d exited 0, printed expansion and source wrapper:-d. |
| agents-infra codex -d / --danger / --yolo | --dangerously-bypass-approvals-and-sandbox. I tools/agents-infra/internal/infra/codex_launch.go:16,567-575. agents-infra codex --print-config --yolo exited 0, printed expansion/source wrapper:--yolo. |

Neither print-config launched an agent. Aliases apply before wrapper --. Native Claude -d means debug, so verbatim forwarding is not parity. Legacy project yolo defaults exist; explicit Claude permission mode suppresses them (I Claude:556-589). Do not copy defaulting.

### Exact installed flags

H includes full standalone help, each exit 0, Claude 2.1.273, Codex 0.153.4, Pi 0.84.2. No bypass execution.

| Tool | Exact flags/semantics | Mapping |
|---|---|---|
| Claude | --dangerously-skip-permissions bypasses all permission checks. | Yolo untracked; no external sandbox/organization-policy guarantee. |
| Claude | --allow-dangerously-skip-permissions makes bypass selectable, not enabled by default. --permission-mode accepts acceptEdits, auto, bypassPermissions, manual, dontAsk, plan. | Allow flag is not yolo; acceptEdits/dontAsk not unrestricted equivalents. |
| Codex interactive AND exec | --dangerously-bypass-approvals-and-sandbox skips confirmation and executes without sandboxing. | Yolo; verify exec insertion in production tests. |
| Codex interactive | -a/--ask-for-approval: on-request, never. -s/--sandbox: read-only, workspace-write, danger-full-access. | Never alone is not yolo; sandbox alone does not select approvals. |
| Codex interactive AND exec | --approve-for-me uses automatic review with workspace-write sandbox. --dangerously-bypass-hook-trust bypasses persisted hook trust. | Neither equivalent; never add hook-trust bypass. Exec help does not list -a. |
| Pi | No permission-bypass flag in full help. --approve/-a trusts project-local files; --no-approve/-na ignores them. | Refuse yolo; do not map --approve or disable extensions. |
| OpenCode | Absent in B5, refused by launcher mapping. | env_unsupported; no invented mapping. |

Pi installed README:502 says no permission popups. [Pi upstream README](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/README.md) describes extension gates and no built-in MCP; no Curator MCP channel follows. [Codex CLI reference](https://learn.chatgpt.com/docs/developer-commands?surface=cli) supports bypass semantics; installed help governs exact version. Historical full-auto/untrusted/yolo alias spellings are not exposed in installed help and are not verified canonical mappings.

### Interface options

| Option | Assessment |
|---|---|
| --yolo boolean, optional -d/--danger | Familiar parity; -d confuses native debug. Possible alias after semantics settle. |
| --permissions native or yolo | **Recommended.** Native means no launcher override, not guaranteed prompts. Omission equals native. Yolo requests declared native bypass, not outside-policy bypass. |
| --permissions standard or yolo | Standard suggests a uniform safe posture tools do not share. Avoid without exact definition. |
| ask/auto/never/yolo or split sandbox/approval | False equivalence/version-dependent grammar; defer. |
| Reuse --ax-profile yolo | Reject: tracked-only concept, distinct ownership. |

Rules:

- Yolo explicit per invocation only; never derive from profile/model/prompt/credentials/env/defaults/ax. Keep defaults.json closed to model/effort (L SPEC:324-355); reject permission members. Native stored settings can still relax native mode.
- Resolve env/version/invocation capability before mutation/launch. Unsupported: proposed permission_mode_unsupported. Invalid/repeated/conflicting: usage exit 2. Failed probe is not known support. Reviewed capability table, not runtime help parsing.
- Explicit yolo rejects conflicting native selectors: Claude permission-mode/restricted/duplicates; Codex sandbox/approval/config-policy/competing bypass. Cover equals/separate forms, aliases, exec placement. Provider grammar distinguishes prompt text from flags. Unknown policy forms cannot yield false resolved-policy claims.
- Native/omitted preserves raw argument contract and guards. Raw bypass may remain available untracked: interface is UX, not security perimeter. Record raw requests separately.

### ax interplay

L SPEC:623-629 forbids native bypass in composed requests and says ax plugins refuse it from native argv. L internal/plan/plan.go:227 requests interactive; composition.go:104 appends native args; execution/execution.go:46,74,139 serializes tracked or directly launches untracked. No real ax call tested.

| Selection | Recommendation |
|---|---|
| Untracked/native | Existing direct launch; ax-profile remains usage error. |
| Untracked/yolo, verified Claude/Codex | One canonical mapped flag after conflict checks. |
| Tracked/native, ax absent/standard/yolo | Existing ax path; no native policy derived from ax. Report separately. |
| Tracked/typed yolo, even ax-profile yolo | **Refuse until versioned ax capability admits it.** No fallback to untracked; names do not prove equivalence. |
| Pi/yolo | Refuse both modes. |

Full tracked parity needs approved ax-owned permission representation/validation or reviewed native-mapping admission. Launcher argv alone violates contract. Open implementation decision, not a research blocker.

Audit requested mode/origin (typed/raw/absent), mapping/version, tracking, independent ax profile, credential mode/strategy, config source/lock and refusal. Separate requested/resolved from effective enforcement. No full argv/auth/inherited-env dumps. Existing tracked schema remains until versioned extension admission (L SPEC:631-640). Untracked stderr can print non-secret resolution, not pretend ax audit.

## 4. Recommendations and spec/implementation touchpoints

**Credentials:** retain existing knob/refusals; first repair migration reconciliation and Pi root. Keep macOS Claude isolated until independent namespace verified. Defer copying. Report backend/link health separately from authentication.

**Permissions:** explicit permissions native or yolo, default native; one reviewed provider mapping owner. No default-yolo config. Refuse Pi/OpenCode/unknown/tracked-yolo under current contract. Preserve ax-profile. Optional --yolo alias; decide -d separately.

| Area | Touchpoints |
|---|---|
| Credential spec | S protocol/environments.md §7.1 homes, §7.4 strategies/isolation/seeds/migration, §7.7 diagnostics, §7.9 versions, §8.3 backups, §9 onboarding/unmanage, §10.1 repair, §12.1 knobs/§12.2 locks. S profiles/manager.md §1 ownership/locks, §12.4 credentials. Version Pi change. |
| Schemas/vectors | Config v2 if grammar changes; marker/status if mode/provenance added. Frozen v1 unchanged. Negative vectors for waiver misuse/profile auth selection/unsupported backend/credential conflict. |
| Curator | internal/envregistry capabilities; internal/config/environments.go locks; internal/envprofile/switch.go:75-80 root; managed.go effectivePassthrough/codexKeyring/finalizeMarker/checkPassthrough and backups; envmarker/status. |
| Permission spec | L SPEC §3 CLI, §4.2 mapping, §4.3 defaults, §4.4 ownership, §4.5 composition, §4.6 ax, §4.7 config, §6 diagnostics, §9 bounds; README. Decision 0013 D3.6/D5/D6.4 before tracked bypass. |
| Launcher | internal/cli parse; mapping/capabilities; internal/plan intent; composition placement/conflicts; execution refusal/provenance; defaults closed schema. Single flag owner. |
| Module/ax | If mapping in A, reviewed interactive capability, unchanged headless defaults. Ax plugin/schema admission and policy ownership must change together. |

Follow-up tests **not run here**: production provision/resolve/repair temporary stores for link removal/conflict preservation/detached files/missing vs unreadable/interruptions/Pi roots. Real launcher fake-process entry for mappings/refusals/equals/raw conflicts/defaults/tracked refusal. Narrowing mutants. Actual Linux/macOS/Windows lanes before coverage claims.

## 5. Open questions and verification accounting

1. Is isolated separate store paths or strict account separation including ambient auth? Recommend bounded store meaning pending full auth-source contract; never filesystem isolation.
2. Is Claude secure-storage override supported/stable? String presence insufficient; authorized disposable-account namespace/login/refresh evidence required.
3. Does Codex keyring identity depend on CODEX_HOME? What about auto/ephemeral/diverged config? Host proves file mode only.
4. Which Pi root wins when both exist? Inventory/operator choice/preservation, not size inference.
5. Full tracked yolo parity? Ax protocol/capability design first; refuse meanwhile.
6. Fleet enforced isolated? Current system schema only shared; reviewed policy revision required.
7. Copying? Recommend no this increment; separate consent/lifecycle, not content waiver.
8. Legacy aliases? Long canonical plus optional --yolo; -d is Claude debug after --.
9. Later versions? Versioned capabilities; fail closed unknown explicit-yolo. Revisit at-or-above credential assumptions.

Shell zsh. No manual landing/test suite; only /tmp artifacts. Standalone native help redirected directly, no tee: claude --help, codex --help, codex exec --help, pi --help each **exit 0**. Legacy print-config each **0**. Metadata Python **0**. B5 not rerun: accepted Claude **1**, Codex **0**, Pi **1**; auth failures remain failures.

Navigation failures, not gates: invalid task-board task query **1**, corrected get; guessed Pi bin/pi **127**, corrected PATH; guessed spec/source rg **2**; no-match rg **1**; envconfig glob zsh **1**. Not evidence of feature absence. Initial artifact tool calls failed JavaScript parsing before execution; ls confirmed absent **1**, then apply_patch recovered.

Independently verified help/source/non-secret metadata. Not reverified OAuth/refresh/Keychain writes/bypass execution/ax runtime/Linux/Windows/upstream identity. HEADs local snapshots, not integration-base attestations. No repository candidate/worktree provisioning.

Logbook-equivalent record (campaign forbids LOGBOOK.md edits): retain knob, refuse unsupported modes, investigate link retention/deletion hazards, correct Pi root, separate native permissions/ax. Also recorded in board notes. Researcher handoff follows assignment's final command; nested developer role conflicts with it.

Artifact verification: standalone Python assertions exited 0 (five ordered AC sections, report below 40 KB, four successful help captures, required native flags and secret-policy topics present). git status --short exited 0 with empty output. These are artifact/integrity checks, not behavioral security tests.
