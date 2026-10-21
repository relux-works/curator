# CLI grammar and UX proposal

Task: TASK-261021-i2f2uq. Review date: 2026-10-21. Proposal only; no implementation or specification acceptance is claimed.

The highest-value change is a coherent help and registration path, followed by separating station observation from operator mutations. Keep `curator run` as the documented front door and retain `curator-run` as the supported direct executable. Those spellings already have an explicit equivalence contract; their coexistence is not itself a defect. [L1] §2

## Evidence and limits

Reviewed snapshots:

| Surface | Snapshot / observation | Evidence |
| --- | --- | --- |
| Curator source | `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`; equal to fetched `origin/main` | Local Git ref comparison |
| Curator executable | `v0.15.0-rc.7`; 72 registered command topics probed with `--help`, all exit 0 | Attached `TASK-261021-i2f2uq_help-evidence.md`, Curator sections; [C1] |
| Launcher source and review build | `dd70beb42239fc80f81bcfd512358f8d3844b615`; built executable reports `0.3.0`, specification `0.5.1-draft` | Attached help evidence; [L3], [L4] |
| FRONT-DOOR | Revision 7 draft; its opening says nothing implemented, while README explicitly says partially implemented | [L1] opening, [L2] Front doors |
| task-board registration | Installed command help, not a registration execution | Attached help evidence, task-board sections |

Source permalinks below freeze the review. Runtime probes exercise help and rejected syntax only. No real agent launch, registry mutation, package installation, station stop, credential inspection, or hosted-manager qualification was performed. Launcher user-config probes use task-local fixture directories. They establish local behavior, not cross-platform qualification. Source review covers the CLI docs and their linked source/transport troubleshooting sections; the older onboarding guide is supplemental. [C2] Shared flags, Skillfile schema 2 project sources; [C7] Skillfile source and transport diagnostics; [C8]

Curator-spec and skill-project-management are change destinations below, not independently inspected snapshots. Their contract references come from the launcher's explicit citations. In particular, a FRONT-DOOR proposal is not evidence that those upstream contracts have already changed. [L1] §0.1; [L5] §§1, 4.8

## 1. Inventory

### 1.1 Curator built-ins

The following table exhausts the 72 `commandRegistry` entries, including parent help topics. It records their advertised help; handler admission differences and additional accepted flags are listed immediately below. Prefix every command with `curator`. `<>` means an operand, `[]` optional syntax, and `|` alternatives. All rows also accept `-h` / `--help` before `--`; `curator help <topic>` is the corresponding explicit help form. `help` itself and top-level `--version` / `-h` / `--help` are additional informational forms. [C1]; [C3] cli.run

**Exit family C:** 0 on success, 1 on failure or blocked result, 2 on usage. `status --check`, `global status --check` and `env status --check` use nonzero status for their documented non-current sets, so an unchecked status success is not a universal health assertion. Provider dispatch inherits the provider result, rather than restricting it to 0/1/2; `run` uses that provider exit family. All 72 observed help invocations returned 0. Parent groups require a child for ordinary operation. [C3] exit constants, cmdStatus, parseGlobalStatusOptions; [C5] cmdEnvStatus; [C3] runProvider; [C6] discovery; [C2] curator status

| Command / topic | Operands and command flags | Meaning | Exits |
| --- | --- | --- | --- |
| `bootstrap` | `[--skills-root <dir>] [--default-agents <ids>] [--preferred-locale <locale>] [--force] [--if-missing] [--non-interactive]` | Create the machine configuration and default directories. | C |
| `init` | `[path]` | Create Skillfile.json and the managed .gitignore block in the project directory. | C |
| `add` | `<name> [--git <url>] [--branch <ref>] [--tag <ref>] [--revision <ref>] [--source <dir>] [--project <path>]` | Add or replace a skill declaration, then install it. | C |
| `remove` | `<name> [path] [--project <path>]` | Remove a skill declaration. | C |
| `install` | `[path] [--all] [--dry-run] [--fix-gitignore] [--strict-tags] [--verbose] [--audit[=advisory\|strict]] [--build-ssh-identity <path>] [--build-ssh-agent <socket>] [--build-ssh-known-hosts <path>]` | Apply Skillfile.json to the selected project. | C |
| `update` | `` | Fetch all source repositories under skills_root. | C |
| `upgrade` | `[path] [--all] [--dry-run] [--fix-gitignore] [--strict-tags] [--verbose] [--audit[=advisory\|strict]] [--build-ssh-identity <path>] [--build-ssh-agent <socket>] [--build-ssh-known-hosts <path>]` | Fetch the selected dependency closure, then install it. | C |
| `status` | `[path] [--all] [--check] [--json] [--attest]` | Report manifest, installed, and compiled state. | C |
| `list` | `` | List configured projects and declared skills. | C |
| `project` | `<add\|resolve\|refresh> [arguments]` | Register projects or resolve their dependency closures. | C |
| `project add` | `<alias> <path> [--agents <ids>]` | Register a project and initialize its declarative files. | C |
| `project resolve` | `[path]` | Resolve transitive dependencies for a project closure. | C |
| `project refresh` | `[path]` | Refresh the frozen project source lock. | C |
| `skill` | `check <dir> [--locale <locale>] [--json]` | Validate a skill package. | C |
| `skill check` | `<dir> [--locale <locale>] [--json]` | Validate one skill package. | C |
| `global` | `<init\|add\|remove\|list\|status\|install\|adopt\|update\|upgrade> [arguments]` | Manage machine-wide skill installations. | C |
| `global init` | `` | Create the global Skillfile.json and managed directories. | C |
| `global add` | `<name> [--git <url>] [--branch <ref>] [--tag <ref>] [--revision <ref>] [--source <dir>]` | Add a global skill declaration and install it. | C |
| `global remove` | `<name>` | Remove a global skill declaration. | C |
| `global list` | `` | List global skill declarations. | C |
| `global status` | `[--check] [--json]` | Report global installed and compiled state. | C |
| `global install` | `[--dry-run] [--audit[=advisory\|strict]]` | Install global skill declarations. | C |
| `global adopt` | `<command> [--dry-run]` | Adopt an existing global command. | C |
| `global update` | `` | Fetch global source repositories. | C |
| `global upgrade` | `[--dry-run] [--audit[=advisory\|strict]]` | Fetch the global dependency closure and install it. | C |
| `profile` | `<install\|import\|list\|use\|update\|remove\|sync\|compose> [arguments]` | Manage environment profiles. | C |
| `profile install` | `<source> [--directory <dir>] [--range <range>\|--tag <tag>\|--revision <ref>] [--as <name>] [--use] [--takeover] [--confirm-system-delta]` | Install an environment profile from a source package. | C |
| `profile import` | `[--as <name>] [--allow-lossy] [--use]` | Import the managed home as an environment profile. | C |
| `profile list` | `` | List installed environment profiles. | C |
| `profile use` | `<name> [--env <env-id>] [--target <target-id>] [--clear] [--takeover]` | Activate or clear an environment profile. | C |
| `profile update` | `[<name>] [--all] [--takeover] [--confirm-system-delta]` | Update installed environment profiles. | C |
| `profile remove` | `<name> [--purge]` | Remove an installed environment profile. | C |
| `profile sync` | `[--takeover]` | Synchronize active environment profiles. | C |
| `profile compose` | `<profile> <add\|remove\|list> [arguments]` | Compose an environment profile from installed profiles. | C |
| `profile compose <profile> add` | `<source> [--range <range>\|--tag <tag>\|--revision <ref>] [--directory <dir>] [--weight <n>]` | Add an overlay to a composed profile. | C |
| `profile compose <profile> remove` | `<source>` | Remove an overlay from a composed profile. | C |
| `profile compose <profile> list` | `` | List members of a composed profile. | C |
| `env` | `<resolve\|status\|config\|migrate\|unmanage> [arguments]` | Inspect and manage environment homes. | C |
| `env resolve` | `<env-id> [--profile <name>] [--repair [--preserve-lock] [--dry-run]] [--takeover] [--format json\|env\|shell]` | Resolve an environment fragment; --repair provisions or repairs the managed home. | C |
| `env status` | `[--check] [--json]` | Report environment home state. | C |
| `env config` | `<show\|set\|unset> [<knob> [<value>]]` | Inspect or edit machine environment configuration. | C |
| `env config show` | `[<knob>]` | Print effective environment configuration. | C |
| `env config set` | `<knob> <value>` | Set a machine environment configuration knob. | C |
| `env config unset` | `<knob>` | Unset a machine environment configuration knob. | C |
| `env migrate` | `--inspect\|--plan\|--apply [--profile <name>] [--env <env-id>] [--expect <plan-hash>]` | Migrate existing environment installations. | C |
| `env unmanage` | `[--restore-backups] [--env <env-id>] [--target <target-id>]` | Restore files from managed home backups. | C |
| `run` | `<env-id> [arguments] [-- <passthrough arguments>]` | Dispatch to the installed curator-run provider; arguments after -- are passed unchanged. | C |
| `hybrid` | `<add\|remove\|list\|status> [arguments]` | Manage hybrid-scope skill declarations. | C |
| `hybrid add` | `<name> [--git <url>] [--branch <ref>\|--tag <ref>\|--revision <ref>] --target <target>\|--targets <targets>` | Add a hybrid skill declaration. | C |
| `hybrid remove` | `<name>` | Remove a hybrid skill declaration. | C |
| `hybrid list` | `` | List hybrid skill declarations. | C |
| `hybrid status` | `` | Report hybrid installed state. | C |
| `audit` | `[target] [--all] [--global] [--json] [--allow <hash> --reason <reason>] [--publish <file> --registry <url> --token <token>]` | Audit skills, pin trust, or publish a signed record. | C |
| `gc` | `` | Remove unreferenced runtime entries. | C |
| `shell-init` | `[auto\|zsh\|bash\|powershell] [--install] [--no-global]` | Print or cache an optional shell hook. | C |
| `hook` | `<approve\|approvals\|revoke> [arguments]` | Manage trust approvals for project environment files. | C |
| `hook approve` | `<path>` | Approve a project environment file. | C |
| `hook approvals` | `` | List recorded project environment file approvals. | C |
| `hook revoke` | `<path>` | Revoke a project environment file approval. | C |
| `ui` | `` | Open the terminal view of installed state. | C |
| `config` | `<show\|build-ssh\|build-https> [arguments]` | Inspect and edit the machine configuration. | C |
| `config show` | `` | Print the effective machine configuration. | C |
| `config build-ssh` | `<add\|list\|remove> [arguments]` | Manage SSH credentials for external build repositories. | C |
| `config build-ssh add` | `<scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]` | Add or replace SSH credentials for a repository scope. | C |
| `config build-ssh list` | `` | List configured SSH credential scopes. | C |
| `config build-ssh remove` | `<scope>` | Remove an SSH credential scope. | C |
| `config build-https` | `<add\|login\|list\|remove> [arguments]` | Manage HTTPS credentials for external build repositories. | C |
| `config build-https add` | `<scope> [--git-credentials] [--keyring] [--token-env NAME] [--username <name>]` | Configure HTTPS credentials for a repository scope. | C |
| `config build-https login` | `<scope> [--username <name>]` | Store HTTPS credentials interactively for a repository scope. | C |
| `config build-https list` | `` | List configured HTTPS credential scopes. | C |
| `config build-https remove` | `<scope>` | Remove an HTTPS credential scope. | C |
| `version` | `` | Print the curator version. | C |

Inventory corrections and boundary rules:

- `profile use` has two actual forms: `<name> [--env E | --target T] [--takeover]`, and `--clear (--env E | --target T) [--takeover]`. The generic synopsis above makes the name appear mandatory even for clearing. [C4] profileUseUsage, cmdProfileUse
- `profile compose <profile> add` also accepts `--tag` and `--revision`, omitted from the registered synopsis. A Git overlay requires exactly one of range/tag/revision; a path overlay admits none and rejects `--directory`. [C9] cmdComposeAdd
- `--audit` supports advisory/strict through its own parser; reference syntax is not a promise that every flag accepts arbitrary optional values. `config build-https add` requires exactly one credential-source selector. `--token` is the existing audit publish flag, not a recommendation to put credentials in command history. [C3] cmdInstallMode, cmdConfigBuildHTTPSAdd, cmdAudit; [C2] curator config build-https
- `hybrid add` requires exactly one of `--tag`, `--branch`, or `--revision`, though the registered synopsis makes those flags look optional. It accepts interspersed flags through the shared parser. [C3] cmdHybrid, pickRef, parseInterspersed
- `global add` parses `--source` and `--branch` but rejects them in its profile-scoped handler; the valid form requires `--git` plus exactly one of `--tag` or `--revision`. `global install` and `global upgrade` actually reuse install flags: in addition to advertised dry-run/audit, they accept `--fix-gitignore`, `--strict-tags`, `--verbose`, and the three `--build-ssh-*` value flags; `--all` and positional paths are rejected. `global update` dispatches to the exact same `cmdUpdate()` as root update, rather than a global-only fetch. [C3] cmdGlobal, runGlobalInstallModeWithProfile, installFlags
- Environment aliases are `claude` to `claude_code`, and `codex` to `codex_cli`; output and stored identities stay canonical. These are input convenience aliases, not new registry IDs. [C10]
- Unknown top-level subcommands can be external `curator-<name>` providers. Their commands/flags/exits are provider-owned, so there is no finite built-in inventory of arbitrary installed extensions. `skill check` is built-in while other `skill` verbs can dispatch externally; `session` is mentioned as an external shim in the provider architecture. No uninspected external verbs are asserted here. [C1] skill handler; [C6] discovery and naming

### 1.2 Launcher, Werkstatt, and manager registry

Every row below supports the two equivalent entry prefixes `curator run` and `curator-run` by design. Runtime help equivalence is currently broken, as finding F2 details. This is the implemented parser inventory, compared against FRONT-DOOR §2, rather than an assumption that all draft features are active. [L1] §2; [L3] Usage, Parse; [L4] run

**Exit family L:** 0 success; 2 syntax/selected configuration-invalid refusals; 1 ordinary operational refusal; 6 hosted protocol/capability/provider/terminal/stdin refusal; 16 hosted policy refusal. Native child exit statuses propagate, with signals represented as 128 + signal. The same code `network_scope_unsupported` is 1 on the direct network path and 16 on the hosted path. Preserve that intentional asymmetry until a versioned contract changes it. [L6] ExitForCode; [L2] Diagnostics and exit codes

| Command | Operands and flags | Exits / constraints | Evidence |
| --- | --- | --- | --- |
| `<env-id>` | Launch options in the next table; optional `resume [SES-HANDLE]` or `--resume <id>`; `[-- native args...]` | L; typed resume never creates a fresh session | [L3]; [L1] §5.3 |
| `werkstatt up <id>` | `--attach`, `--new`, `--task-class C [--difficulty D]` | L; always hosted, delegated trial gate applies; difficulty needs task-class | [L7]; [L1] §§5.8, 6.2 |
| `werkstatt down <id>` | `--force` | L; graceful busy refusal 1; unknown liveness refuses; force still requires process identity/generation proof | [L1] §§5.2, 6.6 |
| `werkstatt recycle <id>` | None | L; graceful down then new generation under one lock; draining timeout refuses 1 and starts nothing | [L1] §6.6 |
| `werkstatt status [<id>]` | `--json`; with ID: `--set-task T` or `--clear-task`, `--queue-depth N`, `--rebind-manager` or `--abandon-operation` | L; dispatch edits need a non-null session; queue depth 0..2^53-1; snapshot can be stale and still exit 0 | [L7]; [L1] §§6.2, 6.5, 6.6 |
| `werkstatt list` | `--json` | L; same status row shape; no ID | [L7]; [L1] §6.6 |
| `session-manager list` | `--json` | L; registry display, not manager launch | [L8]; [L9] |
| `session-manager show <id>` | `--json` | L; invalid or unregistered ID 2 | [L8]; [L9] |
| `session-manager register` | Required `--manifest <path\|->`; optional `--select` or `--no-select` | L; invalid/shadowed registry 2; lock busy 1; selection policy below | [L8]; [L9]; [L1] §4.3 |
| `session-manager select <id>` | None | L; explicit persistent selection; unknown ID 2 | [L8]; [L1] §4.3 |
| `session-manager unregister <id>` | None | L; selected removal leaves no selection, never silently selects another manager | [L1] §4.3; [L9] |
| `migrate` | `--dry-run` | L; plan then apply; destination conflict 2; dry-run has no writes | [L1] §13; [L3] Parse |
| `--help` / `-h` | Top level or environment launch grammar | Observed 0 with absent fixture policy; 1 with malformed fixture `launcher.json`; administrative nested help observed 2 | Help evidence; [L3]; [L4] run |
| `--version` | Informational | Observed 0 with absent fixture policy; configuration can prevent information output | Help evidence; [L4] run |

Launch flags, all before `--`; the native tail is preserved and uninspected at the parser boundary. Repeated flags and mutually exclusive host or permissions spellings refuse rather than using last-wins. [L3] Parse

| Flags / selectors | Values, defaults, and limits | Evidence |
| --- | --- | --- |
| `--hosted`, `--native`, `--untracked` | One host slot; untracked equals native; no flag follows machine `host_mode` (`auto` if absent) | [L3]; [L1] §3 |
| `--board-goal`, `--a2a <identity>`, `--detach` | Hosted-only; board goal absent means none; identity `[a-z0-9._=-]{1,64}`; plain doorbell default heartbeat 1m, six wakes/hour | [L3]; [L2] Hosted doorbells; [L1] §3.3 |
| `resume [SES-HANDLE]`, `--resume <id>` | Latest/manager handle versus native provider identity; hosted-only typed intent; native tail cannot choose the hosted resume target | [L3]; [L1] §5.3 |
| `--profile <name>` | Current profile when absent; environment resolve always repairs | [L3]; [L5] §4.1 |
| `--system-prompt <append\|replace>` | Opt-in absent by default; channel availability can refuse | [L3]; [L5] §5 |
| `--model <model>`, `--effort <effort>` | Explicit values above configured defaults, admitted by spawn plane | [L3]; [L5] §4.3 |
| `--permissions <native\|yolo>`, `--yolo` | One explicit permission slot; yolo alias retained; legacy `-d` / `--danger` rejected | [L3] |
| `--name <session-name>` | `[A-Za-z0-9][A-Za-z0-9._-]{0,63}`; ax default env plus UTC timestamp; ignored when untracked | [L3] Usage |
| `--ax-profile <standard\|yolo>` | ax only; absent uses ax default; native/untracked rejects this selector | [L3] Usage, Parse |
| `--network <profile>` | Direct execution; ax tracked and hosted restrictions apply | [L3]; [L5] §4.4b |
| `--network-policy <policy>` | `optimistic` default, `strict`, or `pinned:<binary SHA-256>` | [L3] |
| `--network-known-bad <path>` | Explicit known-bad list; read or validation failure refuses | [L3]; [L5] §4.4b |

`werkstatt`, `session-manager`, and `migrate` are reserved command operands, not environment IDs. Environment aliases match Curator's aliases. `muse` is supported subject to fragment/runtime capability gates; `opencode` is currently unsupported by this launcher. The broker/one-shot exec contract in SPEC §10 is explicitly a target amendment with unpublished CLI/exit decisions, so it is not an implemented command omitted from this table. [L1] §2; [L3]; [L5] §10; [L2] Target one-shot exec mode

### 1.3 Task-board registration seam

This inventory covers the requested registration seam, not all unrelated task-board commands. Task brief, Inputs

| Command | Flags | Exits and status of evidence |
| --- | --- | --- |
| `task-board session-manager` | `-h` / `--help` | Observed help 0; advertises `register` |
| `task-board session-manager register --for curator-run` | `--select` or `--no-select`; `-h` / `--help`; inherited `--board-dir`, `--context-profile`, `--insecure`, `--json`, `--no-update-check`, `--remote`, `--remote-board` | Observed help 0. Actual registration and wrapper failure exit mapping not exercised or specified in its help; do not infer complete mapping from the launcher's exits. |

Evidence: attached help transcript, task-board sections. The exposed command says it pipes the installed Task-board manifest to curator-run; FRONT-DOOR requires task-board to emit the manifest and never write the registry itself. Plugin RPC entry points such as `task-board session-manager plugin <op>` are a technical interface in FRONT-DOOR §12, not advertised human registration verbs. [L1] §§4.3, 12

The numeric refusal map matters more than adding new synonyms. Current registry errors split into invalid/shadowed/unknown ID (2) and lock contention (1); station errors split into record/layout/context validation (2) and lifecycle/availability failures (1). Hosted capability and policy are 6 and 16. Full code/detail vocabularies remain at FRONT-DOOR §11 and `diagnostics/plugin.go`; preserve them in any grammar migration. [L1] §11; [L6]; [L10]

## 2. Findings

Each recommendation below is proposed, not an instruction to mutate existing policy.

| ID | Finding and evidence | Consequence / recommendation |
| --- | --- | --- |
| F1 | **Documentation maturity is contradictory.** FRONT-DOOR's opening says nothing implemented; README describes a partially implemented front door and executable Werkstatt lifecycle; parser/build expose it. Curator's root help omits `profile import` although nested profile help lists it. [L1] opening; [L2] Front doors, Werkstatt lifecycle; [C3] usage; [C1] | Label implemented, gated, and proposed sections per surface; generate root summaries from the registry. Do not make the draft look wholly implemented or wholly absent. |
| F2 | **Provider help is swallowed; nested launcher help refuses.** `curator run session-manager register --help` exits 0 with generic run help. Direct `curator-run session-manager register --help` and complete `werkstatt status example --help` exit 2. Root launcher help exits 0. [Help evidence; L3; L4 run; C3 cli.run] | Delegate recognized provider topic help through the trusted provider resolution boundary, and add dedicated nested help before any configuration/mutation path. Missing provider should leave useful umbrella help plus an actionable diagnostic. |
| F3 | **Help depends on live policy in the launcher, and the reference contradicts itself.** Malformed fixture `launcher.json` makes `curator-run --help` exit 1. Curator docs Shared flags promise nested help 0 without config, but Global scope still claims nested help needs config and exits 2. All 72 Curator topics actually returned 0. [Probe evidence; L4 run; C2 Shared flags, curator global init] | Make help/version static and read-only, even when configuration is broken; fix obsolete reference prose. Preserve a separate policy diagnostic on real launches. |
| F4 | **Two front-door spellings are intentional; hints choose inconsistently.** FRONT-DOOR §4.3 uses direct register while rule 6 suggests umbrella select; the resolution notices also choose umbrella. Equivalence is explicit in §2. [L1] §§2, 3.2, 4.3 | Document umbrella everywhere for humans, identify direct executable once for package/plugin authors, and test both. Correcting the hint alone does not solve F2. |
| F5 | **`status` mixes observation, dispatch edits, recovery, and reconciliation.** Its flags set/clear tasks, change queue depth, replace manager bindings, and abandon operations; ordinary status also refreshes liveness and can reconcile pending work. [L1] §§5.7, 6.2, 6.5, 6.6; [L2] Werkstatt lifecycle | Keep bounded observation/reconciliation semantics explicit; move deliberate dispatch and recovery mutations to named verbs. Do not promise that current status is a filesystem-pure read. |
| F6 | **`werkstatt list` and status without ID have the same documented shape.** Both return station record, liveness, manager status, a2a evidence and stale marker. [L1] §6.6; [L7] | Prefer `list` for all and `status <id>` for one, retaining no-ID status as an alias. Avoid inventing different payloads simply to justify both. |
| F7 | **Update/upgrade/refresh/resolve are distinct but hard to discover.** Top-level update fetches all source repositories; upgrade fetches a selected dependency closure and installs. Schema-2 resolve freezes sources, refresh explicitly advances them, install materializes the lock; neither launch nor status rescans sources. [C1]; [C2] curator update, curator upgrade, Skillfile schema 2 project sources; [C11] | Keep the distinct effects, label them in help, and resist collapsing them into a single update that silently advances locks. `project list` can clarify the current root list. |
| F8 | **Selection language and optional activation need clear scope.** `profile use` changes active profiles, manager `select` changes launcher selection; install `--use` is convenience, whereas run `--profile` is a launch override. Clear-scope syntax and compose tag/revision help are inaccurate. [C1]; [C4]; [C9]; [L1] §4.3 | Reuse `use` for manager selection with an alias from select; keep scoped profile use and per-launch override distinct. Correct the synopsis before larger renames. |
| F9 | **First registration has hidden authority and lifecycle effects.** First manager auto-selects; later no-TTY registration adds without selecting, while TTY registration asks. A writable system directory chooses the system target; an existing system registry shadows user mutations. Same-ID re-registration updates its timestamp. Werkstatt owner bindings include executable and registration timestamp, so automatic re-registration can invalidate an old owner binding. [L1] §§4.2, 4.3, 6.2; [L9] Service.Run | Do not re-register on every launch or choose write scope from incidental privileges in a lazy path. Missing registration and updating an existing manager require different behavior. |
| F10 | **Install-time registration has no package-hook path after setup removal.** FRONT-DOOR §§4.3/12 assume a manager installer runs registration. Curator promises zero package-code execution during install. The brief stipulates setup.sh removal. [C12] What Curator manages; [L1] §§4.3, 12; Brief §4 | Make explicit registration the baseline; evaluate constrained first-use registration in section 4. Do not reintroduce package scripts as an implicit hook. |
| F11 | **Generic administrative errors lose local guidance.** Both registry and station parsers reduce many errors to 'expected ... with its documented arguments'; nested help offers no recovery. Some runtime messages say only 'Session resume refused', with useful distinction surviving in the code. [L8]; [L7]; [L10] | Keep stable codes and add a specific offending flag/operand plus the smallest valid help or remediation command. Use distinct human text for not-found versus ambiguity. |
| F12 | **Status JSON success does not imply established liveness.** Lock timeout gives a stale snapshot and exit 0, sometimes null liveness; unreadable manager state is unknown; a2a presence remains separate. `--abandon-operation` explicitly declares an override and stops nothing remote. [L1] §§6.5, 6.6 | Agent dispatch must require fresh manager evidence and a usable session, never `exit == 0`, a2a presence, or override-down alone. Put this rule next to status examples. |
| F13 | **Defaults vary by caller but must not vary by guessed identity.** Registry registration prompts only on a TTY, first-manager selection does not prompt, launches never ask about managers, and board-goal binding is opt-in. Compatibility task-board launch aliases add board-goal explicitly. [L1] §§3.3, 4.3, 12 | Interactive setup may ask once; automation must specify selection intent. No hidden 'agent mode' that changes security, host choice or board ownership. Document the different workflows in the table below. |
| F14 | **Some apparent surplus grammar carries contract meaning.** `--untracked` equals `--native`; `--yolo` abbreviates permissions yolo. However resume handle and provider ID are distinct selectors, task-board's `--for curator-run` identifies the registry owner, and `--` protects native arguments. [L3]; [L1] §§4.3, 5.3; task-board help | Deprecate only redundant advertised spellings. Retain typed selectors, `--for`, and the native boundary; shortening them would erase useful information. |
| F15 | **Machine policy and repair side effects are underexposed.** Auto with a selected manager goes hosted subject to activation; no manager falls back natively with a notice. Native skips registry contact. Plain launch resolves with repair; Werkstatt up can repair the home and write hooks before start. [L1] §§3.2, 3.3, 5.8, 6.2 | Show effective host route and its source in diagnostics; say which local writes can precede a refusal. Never turn an actual hosted refusal into a weaker native launch. |
| F16 | **Output and scope syntax is irregular but not all differences need new flags.** Status/check commands use `--json`; fragment resolution uses `--format json\|env\|shell` with JSON default; env config show is already JSON. Project roots use `[path]` or `--project`, while `--all` is not a universal scope selector. [C1]; [C2] Shared flags; [C5]; [C13] | Standardize the contract in help: text plus optional JSON for reports, explicit format for fragments, one target form per command, and no implication that every flag is global. Add parity only where the operation supports it. |
| F18 | **Help is not always an admission contract.** Global add advertises source/branch that the handler rejects; compose help omits tag/revision; global install/upgrade admit more flags than listed. Global update is an alias of root update despite a narrower help summary. [C1]; [C3] cmdGlobal, installFlags; [C9] | Generate or cross-check help against handler contracts, show forbidden combinations, and describe global update as an alias rather than claiming narrower scope. |
| F17 | **Umbrella discovery and registry execution have different trust rules.** Current Curator revision A uses PATH with a warning outside trust roots, while registered managers store resolved absolute executables and never use repository/profile/env endpoints. [C6] activeProviderRevision, resolveRevisionA; [L1] §4.2 | A lazy registrar must not borrow revision A's permissive PATH selection. Trusted discovery for registration requires its own explicit contract; see section 4. |

### Human and agent workflows

These are recommended workflows under the same authority model, not newly implemented defaults. Existing no-prompt launch and board-goal rules remain authoritative. [L1] §§3.2, 3.3, 4.3

| Situation | Human | Agent / automation |
| --- | --- | --- |
| Explicit manager registration | May answer a one-time TTY selection question | Pass `--select` or `--no-select`; a needed product choice goes to its operator |
| Launch with configured policy | Use policy default, with a one-line route explanation | Same policy; pin `--hosted` or `--native` when execution placement is part of the task |
| Board goal | Pass `--board-goal` for a board-bound session | Pass only when the task authorizes that board ownership; never infer from current directory |
| Structured status | Text can summarize source, stale and unknown | Use `--json`; validate fresh manager evidence and required capability/session fields |
| Recovery | Inspect bound owner before explicit recovery action | Never infer consent to abandon/rebind from timeout; report exact unresolved owner or operation |
| Refusal | Concrete code, cause, one safe next command | Same code and exit, stable structured fields where a versioned JSON error contract is introduced; no retry that changes semantics |

## 3. Proposed grammar

### Rules and compatibility

Use `noun verb operands`, with a small regular vocabulary: list/show/status for views; add/remove for declarations; install/update for materialization; use for selection. Preserve specialist verbs where they represent different effects: resolve, refresh, migrate, register, up/down/recycle. The root project shortcuts and global scope prefix are useful existing conventions, so a complete rename into a new hierarchy is not a high-value first step. [Proposal based on F5-F8, F14-F16]

Alias policy A: existing root project shortcuts, `curator-run`, environment aliases, and typed provider interfaces remain supported indefinitely. Alias policy B: old CLI spellings for newly advertised verbs remain for at least two minor releases and 90 days, whichever is longer, then require an explicit compatibility review before removal. Advertise only the canonical form; a migration notice may appear once on interactive stderr, never contaminate JSON or run on every background call. This duration is a proposal, not an existing release promise. Wire IDs, config keys, diagnostic codes and environment IDs do not change with CLI aliases. [Proposal; current wire boundaries: L1 §§4-5, 11]

Prefix `curator` is included below; `run ...` rows retain direct `curator-run ...` equivalence. Rows marked keep are deliberate decisions against unnecessary churn.

| Noun | Now | Proposed | Reason / compatibility |
| --- | --- | --- | --- |
| profile | `profile install SOURCE [--use]`, `profile import [--use]` | Keep | Installation and optional activation are useful separate effects; A. [C1] |
| profile | `profile use NAME [--clear]` generic help | `profile use NAME [--env E\|--target T]`; separate synopsis `profile use --clear (--env E\|--target T)` | Match existing scoped clearing without a new verb. [C4] |
| profile | `profile compose PROFILE add SOURCE` | `profile overlay add PROFILE SOURCE` | Make noun/verb discoverable before operands; B alias for compose. [C9] |
| profile | `profile compose PROFILE remove SOURCE` / `list` | `profile overlay remove PROFILE SOURCE` / `list PROFILE` | Same ordering and one overlay noun; B. [C9] |
| profile | `profile update [NAME] --all`, `sync`, `remove [--purge]`, `list` | Keep; show `[NAME\|--all]` as alternatives | Distinct lock/home effects; the handler already rejects simultaneous NAME and all. [C1]; [C4] |
| env | `env resolve E --repair [--preserve-lock] [--dry-run]` | Keep for integrations; advertise the read versus repair distinction | Avoid breaking the launch fragment seam; JSON remains default and dry-run requires repair. [C5] |
| env | `env status [--check] [--json]` | Keep | An existing useful matrix/check interface. [C5] |
| env | `env config show\|set\|unset ...` | Keep; document typed knob values and effective source | Avoid duplicating configuration ownership in launcher commands. [C13] |
| env | `env migrate --inspect\|--plan\|--apply [--expect HASH]` | `env migrate inspect\|plan\|apply [--expect HASH]` | Actions become verbs; apply retains its required expected hash; B for mode flags. [C1]; [C14] |
| env | `env unmanage [--restore-backups] [--env E\|--target T]` | Keep | Name conveys removal of management; do not relabel it uninstall or imply credentials are removed. [C5] |
| project | `list` | `project list`, root `list` alias | Put configured projects under their noun; retain skill summaries and A alias. [C3] cmdList |
| project | `project add ALIAS PATH`, `init [PATH]` | Keep | Registration and initializing a directory are different operations. [C1]; [C3] |
| project | `project resolve [PATH]`, `project refresh [PATH]` | Keep | Freeze versus advance is a real lock contract; show resolve/refresh then install workflow. [C11] |
| project | Root `add`, `remove`, `install`, `upgrade`, `status` | Keep existing project shortcuts; document scope beside examples | Avoid a bulk rename that changes no behavior; A. [C1] |
| project | `remove NAME [PATH] [--project PATH]` | Advertise `remove NAME [--project TARGET]`; retain positional path alias | One advertised target position; B for redundant path form. [C1]; [C3] cmdRemove |
| project | `update` | Keep, describe as fetch all configured source repositories | Do not silently make it install, refresh, or update only one project. [C1]; [C3] cmdUpdate |
| run | `curator-run ...` and `curator run ...` mixed in prose | `curator run ...` in operator docs; direct executable in provider contracts | One documented front door, both supported; A. [L1] §2 |
| run | Generic `run --help`; nested help is swallowed | `run --help` and `run <topic> --help` from provider-owned static topics | Same help through both entry forms; fallback umbrella synopsis when provider is missing. [F2] |
| run | `--native`, `--untracked` | Advertise `--native`; keep untracked alias | Removes confusing ax-era synonym without changing host semantics; A alias. [L3] |
| run | `--permissions native\|yolo`, `--yolo` | Advertise explicit permissions; retain yolo shorthand | Keep convenience without inventing a third permission grammar; A. [L3] |
| run | `resume [HANDLE]`, `--resume NATIVE-ID` | Keep, explicitly label manager handle versus provider identity | They are not interchangeable IDs; preserve refusal-only resume. [L1] §5.3 |
| run | `--board-goal`, `--a2a`, `--detach`, network/prompt/model flags | Keep and group in help by hosted, context, model, network | Discover constraints without adding mode-specific wrappers. [L3] |
| werkstatt | `up`, `down [--force]`, `recycle` | Keep | Small, meaningful lifecycle vocabulary; recycle remains graceful and never silently forces. [L1] §6.6 |
| werkstatt | `status [ID]`, `list` | Advertise `status ID`, `list`; retain no-ID status alias | One/all split, same schema with array cardinality retained; B. [L1] §6.6 |
| werkstatt | `status ID --set-task T\|--clear-task [--queue-depth N]` | `werkstatt update ID [--task T\|--clear-task] [--queue-depth N]` | Deliberate dispatch metadata mutation has a verb; preserve non-null-session gate; B for old flags. [L1] §6.5 |
| werkstatt | `status ID --rebind-manager\|--abandon-operation` | `werkstatt recover ID --rebind-manager\|--abandon-operation` | Put ownership override under explicit recovery; require exactly one action; B for old flags. [L1] §§5.7, 6.2 |
| session-manager | `list`, `show ID` | Keep | Registry inspection, distinct from session liveness. [L8]; [L9] |
| session-manager | `select ID` | `use ID` | Reuse selection verb from profile; B alias from select. [L1] §4.3; [C4] |
| session-manager | `register --manifest PATH\|- [--select\|--no-select]` | Keep, add explicit `--scope user\|system` | Machine layer is an authority decision; default to user, refuse existing system authority as today; B for implicit target selection. [L1] §4.3 |
| session-manager | `unregister ID` | Keep, same explicit scope as register/use | A protocol registry entry is not an installed package; no binary removal implied. [L1] §4.3 |
| session-manager | `task-board session-manager register --for curator-run` | Keep as the explicit bootstrap command | The manager emits its own manifest; `--for` declares the destination, not an extra launch front door. [L1] §§4.3, 12 |
| global | `global init\|add\|remove\|list\|status\|install\|adopt\|update\|upgrade` | Keep except advertise root `update` instead of redundant `global update`; retain its A alias | Stable scope, accurate global-add restrictions and install flags; global update has no narrower implementation. [C1] |
| global | `bootstrap --if-missing --non-interactive ...` | Keep | Already expresses idempotent machine initialization and noninteractive failure. [C1]; [C2] curator bootstrap |
| global | `audit [--allow ...\|--publish ...]` | Advertise `audit [TARGET]`, `audit allow HASH --reason TEXT`, `audit publish FILE --registry URL` | Separate inspection from trust mutation/publishing; B for mode flags; new publish syntax must use a non-argv credential input contract. [C3] cmdAudit |
| global | `hook approve`, `hook approvals`, `hook revoke` | `hook approve`, `hook list`, `hook revoke` | Consistent listing verb; B alias for approvals. [C1] |
| global | `config build-ssh` / `config build-https` | Keep; grouped under build credentials in help | Avoid a broad credential hierarchy rename while source types differ. [C1]; [C2] Operator configuration |
| global | `shell-init`, `gc`, `ui`, `version` | Keep; advertise `--version` as the portable info flag | Small utilities do not need another noun layer. [C1]; [C3] cli.run |

For audit publish, a future token-file/stdin mechanism needs bounded reading, no credential echo, and a nonconflicting payload channel; it is a separate contract change, not an invented currently supported flag. Credentials and transport behavior are outside this proposal's implementation scope. [Proposal based on C3 cmdAudit]

## 4. Registration without setup.sh

### 4.1 Baseline that works without an install hook

After installing the two binaries, the operator explicitly runs:

```text
task-board session-manager register --for curator-run --select
```

Use `--no-select` when preserving the selection is intended. The current wrapper emits the installed manager manifest and invokes curator-run's registry command. Curator installs declarative packages without executing their code, so installation itself cannot be responsible for running this wrapper. No setup script, shell profile edit, or package execution hook is required for this explicit path. Keep package installation success separate from registration success. [task-board registration help; L1 §§4.3, 12; C12 What Curator manages; Brief §4]

The ownership is straightforward: Curator verifies and installs; task-board knows its own plugin entry point and emits the manifest; curator-run alone validates and atomically writes its registry and selection; the selected manager owns sessions. An orchestrator may perform the explicit registration when authorized, but an agent should not substitute a different manager or selection after a refusal. [L1] §§1, 4.2, 4.3, 5.2, 12

### 4.2 Evaluate lazy registration by curator-run

**Recommendation:** ship the explicit command as the guaranteed hook-free route. Lazy registration on first hosted use is a useful optional follow-up only with a launcher-owned trusted discovery contract. An empty registry does not tell curator-run which executable to ask for a manifest. Searching a workspace, a profile, the ambient PATH, or a board-supplied executable would introduce authority that §4.2 explicitly excludes. Current Curator provider revision A is therefore not a suitable registration trust source. [L1] §4.2; [C6] activeProviderRevision, resolveRevisionA

| Option | Benefit | Cost / decision |
| --- | --- | --- |
| Explicit wrapper command | Available grammar, deterministic choice, no new discovery or install behavior | One documented post-install command; recommended baseline. [L1] §4.3 |
| Manager self-registration on its first invocation | Manager already knows its binary location; can reuse the wrapper | Does not fix `curator run` first when task-board has never run; arbitrary task-board read/help commands should not start mutating launcher config. [Proposal; current wrapper seam: L1 §12] |
| Launcher lazy registration from a trusted catalog | Can make first hosted use work without a separate setup command | Needs a new catalog/trust contract, bounded manifest request, transaction rules, and explicit amendments to no-manager auto behavior. [Proposal constrained by L1 §§3.2, 4.2, 4.3] |
| Launcher searches PATH or project files | Low initial implementation effort | Reject: turns untrusted discovery into manager authority and defeats the no-repository-endpoint rule. [L1] §4.2; F17 |

A clean lazy design, if approved, is as follows. These are proposed requirements, not claims about the current implementation.

1. **Install only inert metadata.** A signed distribution may ship a closed launcher-owned catalog of approved manager IDs and trusted executable roots/identities. Package install copies validated metadata and binaries; it runs no package code. Do not let a package nominate an arbitrary executable or extend trusted roots. Independent installations without such metadata use the explicit wrapper. A new manifest-only task-board command would be needed if the launcher consumes stdout directly; do not call the current registration wrapper recursively. [Proposal; constraints: C12; L1 §§4.2, 12]
2. **Trigger narrowly.** Never register on help/version, status/list, migration preview, explicit native, or machine native mode. On an actual explicit hosted launch or `werkstatt up`, after syntax and applicable policy admission, inspect the registry. Auto-empty launch is also technically a first hosted-use opportunity, but changing its current immediate-native result requires an explicit §3.2 amendment; leave it native until that amendment is accepted. [Proposal; existing routing: L1 §3.2]
3. **Respect existing authority.** Invalid/unreadable registry refuses; it is not absence. A selected manager is used without discovery. A nonempty unselected registry requires an explicit use/select decision and is never repaired by picking a manager. An existing system registry is authoritative; no user write, sudo, or fallback layer. A lazy path always targets user scope and never gains system scope merely because the invoking process can write there. [Proposal preserving L1 §§4.2, 4.3]
4. **Discover a bounded, trusted candidate.** Only a catalog-approved installed manager can be queried. Resolve and verify its absolute executable, request a manifest through a pure bounded operation, reject unexpected ID/executable/protocol/argv, and validate before any registry write. No workspace/profile/env endpoint, arbitrary PATH lookup, network fetch, installer execution, daemon start, or credential access. Multiple candidates require explicit selection rather than order-dependent automatic choice. No candidate gives the existing no-manager refusal/notice plus the explicit registration hint. [Proposal; trust foundation: L1 §§4.2, 5.6]
5. **Make first registration idempotent.** For an empty user registry and exactly one approved candidate, add and select once under the existing lock/atomic-write contract. Re-read under lock; a concurrent established selection wins and is never overwritten. Do not issue `register` on each launch: its timestamp is part of old station owner bindings. Existing entries with changed executables need an explicit update/rebind workflow; do not rewrite them merely to make a launch proceed. [Proposal; existing transaction/owner contracts: L1 §§4.3, 6.2; L9]
6. **Never prompt inside launch.** Disclose the successful first registration on stderr once. Ambiguous candidates or authority conflicts return a concrete explicit-registration action. Human and agent launches follow the same rule; interactive prompting belongs in an explicit registration/setup action. Do not retry failures as native or as a different manager. Current `--no-select` semantics remain available in explicit registration. [Proposal preserving L1 §§3.2, 4.3]
7. **Separate committed registration from launch outcome.** Successful registration may remain if later launch admission refuses; report these as two results. Never imply no changes when registration committed, and never claim a session exists merely because registration succeeded. Failure before registry commit leaves old bytes and selection unchanged. Do not roll back a committed selection over another process's subsequent update. [Proposal; existing atomic registration: L1 §4.3]

This makes lazy registration an explicit launcher feature, not a disguised package install hook. It requires an owner decision on catalog authority and on whether auto-empty behavior should change. Until then, document the explicit wrapper as the complete supported path, rather than claiming a hook-free automatic path already exists. [Proposal; L1 §§0.1, 3.2, 4.3]

### 4.3 Required implementation acceptance cases

These cases specify future work. They were not implemented or run in this proposal task. Each new gate must have nearby valid controls and a narrowing mutant, consistent with the development brief.

| Contract | Positive control | Negative / invariant | Narrowing mutant |
| --- | --- | --- | --- |
| Trusted discovery | Approved installed manager emits a valid matching manifest | Workspace/PATH-only, wrong executable/ID, malformed or oversized response starts no manager and writes no registry | Remove root or executable identity validation; planted untrusted candidate must be rejected by the test |
| Existing authority | Explicitly selected approved manager launches normally | Invalid registry, unselected nonempty registry, system shadow, or declined selection never triggers replacement | Treat malformed registry as empty or bypass the system-layer guard |
| Trigger | Explicit hosted first use registers exactly once | Native/help/version/status perform no registrar call and no registry write | Remove native/info short-circuit |
| Atomicity and races | Two independent valid operations preserve both acknowledged entries | Concurrent selection is not overwritten; failed validation/write leaves previous bytes and selection intact | Remove under-lock reread or replace atomic write with partial publication |
| Ownership | Existing station binding remains readable after ordinary later launches | Repeated launch does not change registered_at; changed executable never silently adopts an older session | Re-register every launch or ignore binding timestamp |
| No fallback | Successfully selected manager receives exactly one admitted request | Refusal/timeout/unknown operation never launches native or a second manager/session | Turn hosted refusal into native retry |
| Help | Every canonical/alias topic is static help 0 with absent and malformed config | Help starts no registrar, provider session, hook, repair, or daemon; tail after `--` remains operand data | Move help behind config/registration or scan native-tail help |

## 5. Top five changes by value and ownership

These are implementation slices for later authorization. No repository edits or PRs were made for this review.

| Rank | Change and value | curator | curator-spec | curator-agent-launcher | skill-project-management |
| ---: | --- | --- | --- | --- | --- |
| 1 | **One help contract through both front doors.** Fix actual discoverability failures before renaming verbs. | Trusted provider help delegation, generated root summaries, cli.md corrections, positive/negative help probes. | Amend umbrella/help contract in environments §11 and CLI reference if provider delegation needs new routing syntax. | Static nested help for all station/registry/migrate verbs; help/version before config; help equivalence checks. | Keep registration help accurate and establish wrapper refusal exit contract; no need to redesign unrelated board commands. |
| 2 | **Hook-free registration with explicit ownership.** Makes setup removal operable without hidden install execution. | Document post-install wrapper; preserve zero package execution. | Record no-hook registration lifecycle and trusted discovery boundary; amend any install-time assumptions. | Registry UX, explicit scope; optional catalog-backed lazy registration only after the decisions in §4. | Keep manifest emission/registry delegation; replace installer assumptions with explicit registration; add pure manifest output only if lazy design is approved. |
| 3 | **Separate status, dispatch edits, and recovery.** Prevents an apparently observational command from authorizing an ownership override. | Provider discovery/help links and docs only. | Reflect the operator/observation boundary where launcher references are normative; do not change wire liveness rules implicitly. | Add update/recover verbs and old-flag aliases; retain binding CAS, operation tombstones, stale/unknown semantics. | Preserve manager refusal/generation contracts; coordinate any docs or plugin-adapter assumptions, not a new registry writer. |
| 4 | **Actionable diagnostics and truthful status semantics.** Reduces unsafe retries and agent false positives. | Keep check scope explicit; link source/transport remedies. | Preserve exit/code compatibility; publish any new structured error schema and semantic check mode before use. | Specific parse errors and remediation; route-source disclosure; explain stale/unknown/override and intentional network exit asymmetry. | Establish registration wrapper exit/error propagation; preserve runtime refusal codes rather than flattening them. |
| 5 | **Small grammar cleanup with a measured alias window.** Improves regularity without a big command migration. | Fix profile clear/compose help, add project list alias, consider overlay and migrate verb forms, hook list alias. | Update CLI grammar and alias policy; stable IDs/config/wire remain unchanged. | Manager use alias; canonical umbrella examples; de-emphasize untracked; one/all station help. | Adopt canonical remediation spellings; retain the explicit --for boundary and compatibility launch aliases. |

Evidence for ownership: Curator source owns its built-ins and umbrella; launcher SPEC owns CLI/composition, FRONT-DOOR §§4/6/11 own registry/stations/diagnostics; FRONT-DOOR §12 allocates the task-board wrapper/plugin and points to curator-spec's upstream contracts. Specific upstream file edits need confirmation against their then-current trees, which were not present as reviewed snapshots here. [C1]; [C6]; [L5] §§2-6; [L1] §§0.1, 4, 6, 11, 12

## Open decisions

1. Whether to adopt catalog-backed lazy registration and who owns catalog trust/updates. Explicit registration is usable without resolving this.
2. Whether auto-empty launch should attempt trusted first registration before its current native notice. This changes §3.2 and must not be introduced as an incidental UX tweak.
3. Whether to adopt the proposed overlay/migration/use renames at all; help and registration fixes do not depend on them. Alias duration proposed here needs release-owner acceptance.
4. The task-board wrapper's complete numeric failure-exit mapping needs publication or source-backed verification; help alone does not establish it.
5. Which draft clauses are accepted upstream and which runtime activation gates remain. The contradictory FRONT-DOOR banner should be corrected by its owner rather than interpreted as approval.

These are decisions for follow-up slices, not blockers to delivery of this proposal. [F1, F9-F10, F14; §4]

## Source index

All links point to reviewed repository snapshots. Section/function qualifiers in citations identify the exact evidence within each file. Runtime observations are in the separately attached help evidence.

[C1]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/command_help.go
[C2]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/docs/cli.md
[C3]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/main.go
[C4]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/profile.go
[C5]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/env.go
[C6]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/umbrella.go
[C7]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/docs/troubleshooting.md
[C8]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/docs/second-operator.md
[C9]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/compose.go
[C10]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/envalias.go
[C11]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/draft_help.go
[C12]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/README.md
[C13]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/envconfig.go
[C14]: https://github.com/relux-works/curator/blob/cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7/cmd/curator/envmigrate.go
[L1]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/FRONT-DOOR.md
[L2]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/README.md
[L3]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/internal/cli/cli.go
[L4]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/cmd/curator-run/main.go
[L5]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/SPEC.md
[L6]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/internal/diagnostics/diagnostics.go
[L7]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/internal/werkstattcommand/command.go
[L8]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/internal/cli/session_manager.go
[L9]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/internal/sessionmanager/service.go
[L10]: https://github.com/relux-works/curator-agent-launcher/blob/dd70beb42239fc80f81bcfd512358f8d3844b615/internal/diagnostics/plugin.go
