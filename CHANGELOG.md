# Changelog

All notable implementation changes are recorded here.

## Unreleased

### Added

- R5 script-worker-v1 runtime conformance qualification. All 33 named
  behavioral vector cases and 11 mandatory controls map to registered
  production-entry rows; launch, audit, control, and evidence cases are
  required on Ubuntu, macOS, and Windows hosted lanes. Every declared exec
  grant must resolve through the manager-owned search directories or Launch
  refuses before the worker starts; Windows default lookup uses the captured
  manager SYSTEMROOT and copies verified System32 hard-linked binaries into
  the private PATH farm. The native inventory
  matrix is Linux: process teardown, file-size, and handle controls available,
  with cgroup, Landlock, and network namespace controls host-conditional;
  macOS: teardown, file-size, and handle controls available, with aggregate
  process/memory, descendant exec, filesystem, and network controls
  unavailable; Windows: Job Object teardown/process/memory and inherited
  handle controls available, with file-size, descendant exec, filesystem,
  and network controls unavailable. Linux conditional controls remain
  per-invocation probes. Piped input remains capped at 64 MiB and combined
  captured output at 16 MiB; interactive and pass-through streams are outside
  this bounded model.
- E4: umbrella provider lookup from trust roots (warning release,
  revision A). Unknown subcommands still resolve `curator-<name>` on the
  ambient `PATH`, but the manager now computes the trust verdict against
  the trust roots — the manager install directory, then the new
  machine-configuration `provider_directories` list (environments
  §11/§12.1, lockable under §12.2) — and warns
  `subcommand_provider_outside_trust_roots` with the resolved path, the
  roots consulted, and the migration hint when the selection lies
  outside them. Providers inside manager-published or managed
  directories stay refused (`subcommand_provider_untrusted`), an
  unreadable trust root fails (`subcommand_provider_root_unreadable`,
  never absence), and `env status` reports the resolved provider path
  with its trust verdict per discovered provider — always `curator-run`
  and `curator-session` — with refused, missing, and unreadable rows
  non-current for `--check`. Revision B (trust roots only, `PATH`
  never selects) follows in a later release; this story blocks proposal
  0016 / `path_prepend`.
- R4 script audit warning classes for `script-worker-v1` (manager profile
  §7). Every declared-only script command — schema 7 and schema 8 without
  `execution_policy` — now warns `script-command-declared-only` through
  `curator audit`, the install-time audit gate, and `skill check`; every
  enforced command with declared `network` hosts warns
  `script-command-unfiltered-declared-network` (reporting-only, no portable
  filtering applied or claimed). Both classes are always warnings in every
  mode and never block: declared-only skills install exactly as before and
  enforced skills keep their native launchers. Enforced commands are never
  labelled declared-only, and the network label applies only to enforced
  commands with non-empty declared hosts. For every script command the
  audit record additionally carries the effective execution-policy
  identity or its explicit absence (`audit info` lines, `script_policies`
  in `audit --json`, and the stored verdict), independent of warning
  eligibility. Operator guidance in `docs/troubleshooting.md`.
- R3 native probes, capability evidence, and preflight for
  `script-worker-v1`. Every enforced invocation probes the exhaustive
  eight-control native inventory once, before the worker starts, with no
  host-label, cache, or configuration substitution; applies exactly the
  `available`/`host-conditional`-present controls (process-group and
  Job Object teardown with exact limits, `RLIMIT_FSIZE`, handle hygiene,
  and on Linux delegated cgroup v2 bounds, Landlock exec denial and
  write confinement over the derived path set, and a network namespace
  without interfaces); and returns exactly one closed result-only
  `script-capability-evidence-v1` record, which the parent validates
  against its own probe before permitting the run. A host that cannot
  provide a mandatory control refuses install and invocation with
  `script_execution_control_unavailable` before any worker starts;
  record contradictions refuse with
  `script_execution_capability_evidence_invalid`, and deferred-guarantee
  or foreign-policy claims with
  `script_execution_hardened_claim_forbidden`. The 11-control
  implementation table is complete and the R2 table-injection seam is
  removed, so enforced commands admit at `skill check`, install as
  native launchers, and run end to end when the host provides the
  mandatory controls. The invocation record and derivation report are
  available through the new operator-selected `script_diagnostics_dir`
  machine configuration. Stream behaviour stays bounded (64 MiB stdin
  refusal, 16 MiB capture); piped or file stdin is fully buffered, terminal
  stdin is null-bound, and stdout/stderr share the capture budget. Overflow
  reports through the worker and refuses the invocation without forwarding partial capture; live
  pass-through and interactive streams are unsupported. On Linux, a derived
  filesystem path that cannot be ruled
  refuses the invocation fail-closed
  (`script_execution_worker_protocol_invalid` naming the control and
  the offending path) rather than running with a probed-present
  control left unenforced; write confinement grants the derived path
  set, the operation-private area, and the null device, handling every
  filesystem mutation right the probed Landlock ABI provides (write,
  truncation, entry creation/removal/reparenting, device ioctl) while
  reads stay unrestricted.
- R2 declaration-derived enforcement for `script-worker-v1`. Every enforced
  invocation derives its containment profile from the declared capabilities,
  deny by default: a manager-built environment (empty bootstrap plus
  manager-set values plus exactly the non-reserved `env_read` names, with
  the portable, platform, and per-interpreter reserved sets enforced and an
  interpreter without a reserved set refused), a manager-built `PATH` over a
  manager-owned directory exposing exactly the resolved interpreter and the
  manager-resolved declared exec names, offline network configuration with
  proxy/resolver scrubbing when the derived network is none, a
  manager-selected working directory with the private temporary,
  configuration, and cache roots bound through the platform environment,
  reporting-only declared network hosts, and secret identifiers that never
  resolve to values. The worker revalidates the derived profile and starts
  the interpreter only after the parent's permit frame. Enforced commands
  install as native launchers (a manager copy plus a sidecar contract — no
  shell, `.cmd`, or symlink shim) that replay the manager role, and install
  records each command's derivation in its result messages. Operator
  documentation for the `script_interpreters` bindings in
  `docs/script-interpreters.md`.
- R1 script manager/worker invocation path (`script-worker-v1`). The manager
  resolves the closed `node-v1`/`python3-v1` interpreter identifiers from the
  new operator-trusted `script_interpreters` machine-configuration mapping
  only (never repository, runtime root, `.agents/bin`, user `PATH`, or
  manifest), re-executes the installed manager in the fixed hidden
  `__curator-script-worker-v1` mode with a fresh session nonce, identity
  recheck at the launch boundary, explicit stream binding, a private runtime
  area, and worker-domain teardown, reusing the go-v1 worker's executable
  identity primitives. Admission now preflights the 11 mandatory portable
  controls: unsupported policies keep refusing
  `script_execution_policy_unsupported`, while `node-v1`/`python3-v1`
  commands refuse `script_execution_control_unavailable` naming the missing
  controls. Interpreter bindings must name native executable images (the
  `.exe` itself on Windows); wrapper scripts and batch files are refused
  at resolution and at the worker before the interpreter runs. Enforced
  launch remains refused until the R2/R3 control set is complete; no
  enforced script launches uncontained.
- Scoped HTTPS credentials for external build repositories. A `build_https`
  configuration section maps a source scope to a token source — the operator's
  own Git credential for the host, a manager-namespaced keyring entry, or a
  named environment variable — and the `curator config build-https` command
  (`add`, `login`, `list`, `remove`) manages those selections. Selection is
  resolved per repository by canonical source identity before the first fetch
  and is never selectable by a package; a private HTTPS fetch is answered by a
  manager-owned, host-pinned askpass broker, and an uncovered repository stays
  anonymous (`Spec core §12.2`).
- Operator documentation for scoped HTTPS build-repository token sources,
  credential resolution, and the `curator config build-https` command.
  `CURATOR_BUILD_HTTPS_TOKEN` without `CURATOR_BUILD_HTTPS_HOST` is
  identity-unbound and may be offered to every HTTPS build-repository
  host reached by the run; bind it to one host or use a `build_https` scope
  (`Spec core §12.2`).
- Schema-8 first-party module roots for the `go-v1` driver: a build root may
  replace declared module directories elsewhere in the snapshot, and those
  directories join the directive, cgo, and assembly scan surface
  (Protocol Core §4.2.3).
- S6: shell-hook trust gate — warning release (`A-warning`). The POSIX and
  PowerShell hooks now verify a project `.agents/env.sh` / `.agents/env.ps1`
  against the manager-home approval state before sourcing it, and warn once
  per shell session with `shell_hook_env_unapproved` (no record) or
  `shell_hook_env_changed` (bytes differ; re-approval required) naming the
  absolute path and `curator hook approve <path>`. Files the manager itself
  writes are recorded as `approved_by: manager` at write time and source
  silently. This release only warns: unapproved and changed files are still
  sourced. Migration: run `curator hook approve <path>` for each project env
  file the warning names (a project-local approval record is ignored — only
  the manager-home record counts). The operator approval surface is
  `curator hook approve <path>` (records the current bytes as
  `approved_by: operator`, fails without recording when the file is absent
  or unreadable), `curator hook approvals` (read-only listing), and
  `curator hook revoke <path>` (removes the record; revoking a missing
  record leaves state unchanged and exits 0). `curator status` and
  `curator env status` report the shell-hook trust posture per known env
  file (`approved`, `shell_hook_env_unapproved`, `shell_hook_env_changed`);
  a recorded file whose bytes are missing or unreadable keeps its row
  with an explicit qualifier and is non-current under `--check`, as a
  changed file is, while an unapproved file whose bytes read stays a
  warning row. `status --json` gains `shell_hook_trust` (plus
  `shell_hook_trust_warnings` when approval-state warnings exist, so an
  unreadable approval state surfaces in JSON and fails `--check`). The
  enforcing revision (`B-enforcing`,
  refuse without sourcing) follows in a later release
  (Manager profile §8).
- CI guard for the GoReleaser rc channel values.
  `tools/goreleaserconfig` parses `.goreleaser.yml` with `gopkg.in/yaml.v3`
  (not a text search) and requires every `homebrew_casks`/`scoops` entry's
  `skip_upload` and `release.prerelease` to be exactly the string `auto`,
  case-sensitively, failing with the field, entry, and observed value.
  `goreleaser check` validates names and types only and accepts every
  wrong-value mutant, and an unset key is what published v0.14.0-rc.1 to
  the tap and bucket. The check runs as a Go test in the lint lane on
  every push (and in every `go list ./...` lane); `gate-selftest.sh`
  pins the lint wiring structurally (TASK-260908-2kqa77).
- E2: direct-only `class: system` modules. Only the system modules of direct
  packages — the root, the active overlays, and the packages their
  `requires.contexts` name — plus packages admitted by a
  `system_module_waivers` entry reach the system-prompt output and the
  launch fragment; transitive system modules are refused. The machine knob
  `transitive_system_modules` selects `drop` (default, non-breaking: the
  module is skipped at materialization with the
  `context_system_module_dropped` warning naming package and module) or
  `error` (opt-in strictness: resolution fails with
  `context_system_module_transitive` and the lock is left unchanged). The
  knob locks to `error` only; waivers are not lockable. `curator env
  status` reports the effective policy value with every dropped module by
  package and path (Protocol environments §3, §5.5, §12).
- S4 warning release (`s4-warn`, audit finding S4): `profile install`,
  `profile update`, and `env status` now surface every resolved MCP
  declaration package — package, version, transport, stdio command, args,
  and requested `env_names` — as one `mcp-declaration` row printed after
  the audit gate passes and before the lock is published; an empty MCP
  package allowlist warns `mcp_package_allowlist_empty`; and a resolution
  that passes an operator variable outside the configured
  `passable_env_names` list — or any passed variable when the knob is
  absent — warns `mcp_env_passthrough_unlisted` naming the variables and
  the knob with the migration hint ("list the named variables to keep
  passing them after the flip"). `env status` postures the active S4
  profile with the effective `passable_env_names`. An explicit
  `passable_env_names: null` stays unbounded with no warning. The
  enforcing profile (`s4-enforce`: absent knob is empty, unlisted names
  dropped with `mcp_env_passthrough_dropped`) is implemented behind the
  same option and follows in a later release; this release keeps the
  pre-S4 unbounded behaviour and only warns
  (Spec environments §2.2, §2.3, §10.3, §12).
- Conformance pin → v1.0.0-rc.12 (`dced9b8`): the hosted gate now runs
  the manager-config-v2, environments (with the E2 system-module
  admission cases), umbrella-provider-resolution, and
  environments-env-passthrough vectors at the rc.12 root, and the four
  `root-content` skip paths the E2/E4/S4 candidates carried for those
  families are removed — an absent family now fails instead of
  skipping.

### Changed

- The install transaction engine now caches canonical namespace resolutions
  per transaction across journal saves in one per-write recheck epoch
  (invalidated by `checkBoundary` before every publication write; legacy
  journals without a recheck keep walking every save). Behaviour-preserving
  performance fix: the draft failure-at-every-target-class sweep and late
  rollbacks spend far less time in `filepath.EvalSymlinks` on Windows, with
  every rollback and boundary proof unchanged.
- A `go-v1` build root whose `vendor/modules.txt` carries a directory
  replacement the command does not declare is now refused with
  `build_module_root_directive_undeclared`. §4.2.3 requires a command with an
  absent or empty `modules` list to have an *empty* effective replace set, and
  `go mod vendor` materialises an annotation for an **unused** `replace`
  directive exactly as it does for a used one. A schema-6 or schema-7 skill
  that carried an unused directory `replace` therefore built before and now
  fails; declare the directory under `modules`, or drop the directive.

### Fixed

- E4: the user-bin shim directory counts as manager-published — and
  refuses providers under revision A — only once the manager has
  actually published shims there (the ownership ledger exists) or the
  operator declared it via `CURATOR_GLOBAL_USER_BIN` (now honored by
  the lookup; previously ignored). A merely selected PATH entry holds
  no manager-written content, so providers there warn
  `subcommand_provider_outside_trust_roots` and stay current instead
  of refusing; on Windows, where temp trees sit below the user
  profile, the selector otherwise claimed ordinary PATH directories.
  Trust-root and refused-directory comparisons now also match by
  filesystem identity, so an 8.3 short spelling and a case variant
  name the same directory (environments §11).
- Git snapshots are extracted from the object database (`git ls-tree -r -z`
  plus `git cat-file --batch`) instead of `git archive`, so every regular file
  carries exactly its committed blob bytes. `git archive` applied
  `core.autocrlf`, `text`/`eol`, and `export-subst` to its output, which made
  snapshot content hashes depend on the acquiring machine's git configuration
  and the repository's `.gitattributes` (Protocol environments §1.2, core §6.2,
  §6.5). The skills snapshot cache and closure scratch snapshots both use the
  new path; symlinks, gitlinks, path escapes, `.git` components,
  platform-path collisions, and oversize blobs are refused, and `100755`
  keeps its executable bit. Every refusal that needs no blob bytes is decided
  from the `ls-tree -l` listing before `cat-file` starts, so a refused
  extraction writes nothing and never leaves `cat-file` blocked on its
  output pipe; a failure while streaming terminates and drains the child
  before waiting on it and removes what the call wrote. Closure scratch
  snapshots are extracted into a sibling staging directory and renamed into
  place only on success.
- Git snapshot extraction now folds directory components per component when
  gating platform-path collisions. The gate compared only folded full paths,
  so two tree entries whose directories fold together but whose basenames
  differ (`Dir/x.txt` + `dir/y.txt`) were admitted and landed in one physical
  directory on a case-folding filesystem, silently losing the committed tree's
  identity. Every ancestor prefix of every planned target is now tracked
  folded, and a prefix that folds onto another prefix or onto a planned file
  is refused with the existing `duplicate platform path` diagnostic before
  any byte is written; case-sensitive destinations still extract both
  spellings with exact bytes.

- Status no longer reports a successfully installed schema-8 skill as
  `needs-install`. Every reader that decides whether a recorded compiled
  command is knowable now bands on the whole build-bearing marker schema set
  (2, 3 and 4) instead of only the schema the release writes, so a schema-7 or
  schema-8 installation is reported `current` on the marker it actually wrote.
  The old remedy was self-contradictory as well: it told an operator holding a
  marker v4 to reinstall so the manager would record marker schema 2, a schema
  it would never write for that band.
- Garbage collection no longer drops the live build references of a marker v4.
  A schema-8 installation's recorded cache keys went unmarked, so a
  maintenance pass could delete protected cache entries the installation was
  still running from.
- A marker document at a readable schema that is nonetheless invalid is now
  reported as an invalid document rather than as one from a newer manager.
  Schemas 3 and 4 are read by this release, so `upgrade the manager` was never
  the remedy for them.
- A snapshot-extraction spawn failure of `git cat-file --batch` is now
  reported with its operation context (`git cat-file --batch failed in ...`)
  like every other product git spawn, instead of surfacing the bare fork/exec
  text. The underlying cause is preserved in the message and the error chain;
  success-path behaviour is unchanged.
- Draft Skillfile acquisition no longer consults user or system Git
  configuration when cloning or fetching a literal `git:` source
  (`project resolve`/`project refresh`). A hostile
  `url.<evil>.insteadOf` (or `pushInsteadOf`, URL rewriting,
  includes, credential helper, `core.sshCommand`) in the operator's
  git configuration previously redirected the clone and the lock
  bound the attacker's commit; the lane now runs every clone and
  fetch with user and system configuration isolated, independent of
  `HOME`, with no interactive prompt. Credentials come from the
  invoking environment only (SSH agent via `SSH_AUTH_SOCK`,
  `GIT_ASKPASS`, `GIT_SSH`/`GIT_SSH_COMMAND`): a private HTTPS
  source that authenticated through a configured credential helper
  now fails `repository_endpoint_unavailable` instead of prompting —
  provide a non-interactive `GIT_ASKPASS` program or an SSH endpoint
  with an agent (docs/cli.md, docs/troubleshooting.md).
- Draft literal-URL acquisition now builds git's environment from an
  explicit allow-list and runs SSH under a curator-owned command
  (`project resolve`/`project refresh`). Per-invocation ambient
  overrides (`GIT_SSH_COMMAND`, `GIT_SSH`, `GIT_PROXY_COMMAND`,
  `GIT_EXEC_PATH`, every other `GIT_*` override) and the proxy
  environment previously reached git, and the real ssh read the
  user's `~/.ssh/config` (Host aliases, `ProxyCommand`); only the
  contract-listed environment now reaches git (`PATH`,
  `HOME`/`USERPROFILE` for default `known_hosts`/identities, temp
  dirs, `TZ`, Windows process essentials, `SSH_AUTH_SOCK`,
  `GIT_ASKPASS`), transport pins (`http.sslVerify`,
  `http.followRedirects=false`, empty `credential.helper`,
  `GIT_PROTOCOL_FROM_USER=0`) hold per invocation, and ssh runs with
  an empty config (`-F`, `BatchMode`, `StrictHostKeyChecking=yes`,
  `ProxyCommand=none`, `ProxyJump=none`, forwarding disabled;
  host-key validation never disabled). Operator consequences: SSH
  host aliases, `ProxyCommand`, and `IdentityFile` entries are not
  read — use an agent (or a default-named key) and the literal host
  name; unknown hosts fail closed (seed `known_hosts` with
  `ssh-keyscan` or a first manual `ssh`); proxy environment is not
  honoured. `GIT_ASKPASS` stays the only HTTPS credential channel
  (docs/cli.md, docs/troubleshooting.md).
- Audit-registry snapshot verification no longer mistakes a snapshot
  published while a fetch is in flight for a future-dated
  snapshot. The future-timestamp bound was evaluated against a clock
  reading taken before the snapshot fetch, so under a literal zero
  clock skew any whole-second boundary crossed during the fetch
  excluded every trusted registry with `every trusted audit registry
  served a tampered snapshot` instead of the evidence verdict the
  fetch actually returned. The bound now tolerates the checker's own
  latency since it sampled its clock, measured monotonically; a timestamp
  genuinely ahead of the post-fetch clock plus skew is still refused
  with the same class and text, and the stale check is unchanged. This
  removes the nondeterministic refusal-class flip on both the draft
  and legacy lanes (BUG-260920-2d9gfv).
- A `git check-ignore` spawn failure in the managed `.gitignore` gate is no
  longer reported as "generated paths are not ignored by git". A git that
  cannot be executed (a spawn error or a missing git binary) returns a `git
  check-ignore failed ...` error carrying the tool diagnostic, and the
  install refuses with it (failed) instead of skipping with the policy
  message. Git's own verdicts, including "not a repository", are policy
  outcomes as before. The not-ignored message and the success path are
  unchanged.

## 0.15.0-rc.2 - 2026-09-26

### Added

- Draft schema-9 skill-manifest dependencies may select a skill package below a
  pinned repository with `directory`. Directory grammar, containment,
  `SKILL.md` name checks, closure unification, lock identity, and source-audit
  records all bind the selected folder; omitted `directory` remains the
  repository root.

- Skillfile schema 2 source handling is supported by default for projects.
  Schema 1 retains its exact meaning, and no on-disk migration is implicit.

- Schema-2 installs and upgrades replay missing locked snapshots from declared Git or path sources, validate package identity and content hashes before publication, and preserve the committed `Skillfile.lock.json`.

- Production-entry tests for draft registry evidence that differs only in
  repository identity or commit, including prior-state preservation and
  diagnostic redaction at `install.Project`.

- Published-case consumers now share a counted outcome harness. The rc.12
  config, marker, build/cache/lifecycle, skill, environment, closure and interop
  vector/schema families, plus the vendored draft-sources-v1
  semantic/schema/snapshot families, are checked against committed case-count
  pins. A single gap ledger rejects unlisted failures, stale passing rows, and
  vanished listed cases. The complete `skillfile-dev-v2` and `skill-build-v1`
  schema lists and external-repository raw-object, LFS-pointer, pack-index and
  local-config/ref fixture lists now use the same accounting. A gap row records
  work still owed and is never an accepted deviation.

- Skillfile schema 2 collections can select every skill from a repository subdirectory with `directory`, `include`, and `exclude` selectors.

- Skill manifest dependencies can select a skill from a repository subdirectory with `dependencies.skills[].directory`.

- Production-entry test suite for the Decision 0017 credential-link
  repairs and the explicit credential migration (tests only, no behavior
  change). Forty-six rows on temporary stores drive `Resolve`,
  `StatusOf`, `PlanMigration`, `ApplyMigration`, and `curator env
  resolve|migrate`: both 0017 hazards, the mis-targeted vs
  dangling-to-declared split, the inspect → plan → apply migration with
  journal/recovery and no secret copies, every repair refusal class
  with the codex admission table, and no silent repair — each refusal
  proven by a narrowing mutant (conflict admitted means the test
  fails). Every row is registered in
  `.github/ci/platform-cases.tsv`; symlink-dependent rows probe the
  host and skip under the existing `host-capability` vocabulary where
  the Windows host forbids unprivileged creation.

- Explicit credential migration: `curator env migrate
  --inspect|--plan|--apply` (Spec environments §7.4/§10.1, manager
  §12.4/§12.5, Decision 0017). Inspect inventories the old marker, every
  recorded link target, and both Pi roots (`~/.pi/auth.json` and
  `~/.pi/agent/auth.json`) per profile and environment, read-only and
  without reading credential bytes; plan prints the exact operations
  (relink a recorded symlink at the declared native path, unlink a
  stale recorded link) with a plan hash covering the marker identities;
  apply requires the `--expect` hash of a prior plan, prints the locked,
  revalidated plan before the first mutation, and executes exactly the
  printed plan under the manager-home mutation lock through a durable
  migration journal (temp-link + atomic rename per relink), refusing on
  plan drift and on conflicts. A syscall failure mid-op or a failed
  marker publication reverts the links; an interrupted apply leaves the
  journal for the next `--apply --expect <plan-hash>` to recover to the
  prior state before executing. Conflicts that need the operator — an
  isolated→shared account choice, a regular file at a link path, two
  live Pi credentials — refuse with `environment_credential_conflict`
  naming the exact out-of-band decision. No step copies, moves, or
  rewrites credential bytes; the effective mode is preserved and the v1
  marker still records path and strategy only.

Managed-home credential records now carry schema-2 backend metadata and provenance, with pathless records for linkless strategies.

### Changed

- Draft marker-v5 build records now validate the closed raw shape of local
  `go-v1` builds, refusing external-only fields even when their JSON value is
  `null`. External `declared_tag` values must be non-empty valid Git ref names;
  marker schemas 1 through 4 keep their existing rules.

Increase the Go per-package test timeout to 60 minutes on Ubuntu and macOS CI lanes while retaining 120 minutes on Windows.

- Pin curator CI to curator-spec `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (manifest SHA-256 `cb7a98e97543282cdde75f04c55fb15acb91062f8ba6510e3d603db42f1c1ab2`). Record 69 conformance gap rows: 64 target-root failures with board owners and five pre-existing marker-v4 gaps. The eight Windows executable identity cases are counted at the production resolver: six driven and two assigned to STORY-260925-1v7pvn; the uncaptured-SystemRoot hard-link case is fixed on trunk and has no gap row.

- Environments profile follow-ups: path installs now report distinct
  `profile_source_path_missing` and `profile_source_path_unreadable`
  diagnostics; only a genuinely absent machine config selects the default
  policy; machine switches clear scope records equal to the new default in
  the same journaled publish; and the CLI reports when a switch touched no
  adapter homes or when an installed profile remains after activation is
  refused.

Accept the rc.13 suite label for the unchanged script-worker-v1 execution-policy identity while retaining a closed protocol-version set.

- `codex_cli` `isolated` is now admitted under effective `file` credential
  storage only (Spec environments §7.4, Decision 0017). The native
  `config.toml` is parsed as TOML and only the top-level
  `cli_auth_credentials_store` key counts, in any valid spelling; an
  absent file or key resolves to the platform default `file`, so
  `isolated` stays admitted; `keyring` or `auto` is refused with
  `environment_isolated_unsupported`; a selector outside the verified
  `file`/`keyring`/`auto` set fails closed with
  `environment_credential_unsupported` at provisioning and repair, and a
  `config.toml` that does not parse fails closed naming the file instead
  of reading as absent.

- New `pi` managed homes link `auth.json` at the corrected native root
  `~/.pi/agent` (Spec environments §7.4, Decision 0017 choice 3). Existing
  homes linked at the old `~/.pi/auth.json` target are reported detached
  with `environment_credential_conflict`-class wording; the move to the
  agent root runs as the explicit `curator env migrate` step, which
  repair points at instead of re-pointing silently.

- `env resolve --repair` no longer migrates credential ownership (Spec
  environments §10.1, manager §12.4). A recorded link aimed at the wrong
  native target (for example a `pi` home still linked at the pre-0017
  root) and a stale recorded link left behind by a mode change now
  refuse with `environment_credential_conflict` pointing at `curator env
  migrate --plan` / `--apply --expect <plan-hash>` instead of being
  re-pointed or unlinked silently;
  previously repair moved them itself. Repair still re-links an absent
  path and replaces an empty directory, and still refuses a regular
  file, a foreign link, or a non-empty directory without touching bytes.

System environments.isolation locks can enforce either shared or isolated. Silent locked settings are honored, conflicting explicit shared requests fail closed, and existing shared passthrough requires the explicit F-C2 migration.

Update Curator's Skillfile sources v1 conformance to the released corpus, verify locked Git object formats during replay, use current resolved endpoints on refresh, reject SCP-like alias ports before I/O, and enforce repository+commit revocation under advisory policy.

### Fixed

- The managerlock subprocess deadline row now uses a context that is already
  expired and asserts that no project lock file was created, removing the
  platform-speed race from an uncontended 1 ns acquisition test.

- `git check-ignore` now retries one child-start `EACCES` after 100 ms, then
  returns persistent permission errors so installation still fails closed.

Narrow the Windows System32 executable hard-link allowance to resolutions backed by an explicitly captured manager SYSTEMROOT; ambient values do not grant the exception.

- Stabilized the audit-registry exact future-bound test on Windows by creating
  its empty rollback-state catalog before calling the checker. The checker's
  elapsed-time allowance includes cache initialization, whose durable file
  sync could take longer than the one-second test edge. The test still drives
  `CheckSnapshotsWithPolicy` with an isolated cache and retains the same
  inside, exact-bound, one-second-past, and far-future assertions.

- Concurrent Windows snapshot readers now retry destination authentication
  three times at 10 ms intervals when the immutable destination open returns
  `ERROR_SHARING_VIOLATION` (32) during publication. Access denied, other read
  errors, persistent sharing violations, and a destination with different
  contents still fail closed as an immutable-commit conflict.

- Credential-link repairs are fix-first and never destroy bytes (Spec
  environments §7.4/§10.1, manager §12.4, Decision 0017). A stale recorded
  link (`shared`→`isolated`, or a store gone ambient) and a mis-targeted
  recorded link (for example a `pi` home still aimed at the pre-0017
  native root) refuse with `environment_credential_conflict` naming the
  path and pointing at the explicit `curator env migrate` step, removing
  nothing — previously the path was unconditionally removed and
  re-linked. A regular file, a foreign link, or a directory at a link
  path refuses the same way. A correctly targeted link whose native
  target does not exist yet is the distinct detached-pending state — the
  normal pre-login shape — reported as a warning by provisioning, repair,
  and bare resolve, which all succeed loudly; it is never stale and never
  a conflict. A native target that cannot be inspected stays a stale
  conflict/inspection diagnostic.

- Manager-owned absence-sensitive reads now use typed absent and unreadable
  outcomes through a shared filesystem seam. The deny-by-default reader audit
  measures guarded and reviewed sites and fails on an unreviewed absence
  collapse. Status, import, snapshot, purge, cleanup, registry, and runtime
  inventory readers preserve read failures instead of invoking absence
  fallbacks; newly refreshed enforced-shim inventory reads use the same seam
  (environments §8.4.1).

- A same-source reinstall of a git root honours `--use` and `--takeover`
  exactly as a first install does. The reinstall re-resolves as an update
  and then runs the install row's activation — activating when the machine
  has no current or the operator passes `--use`, with `--takeover`
  covering the unmanaged files the install would write — instead of
  delegating to the update and reporting success for no work. Previously
  the §9.5 stop-and-retry `profile install <git-url> --use --takeover`
  accepted both flags, did nothing, and exited 0 saying `updated profile`;
  a retry that still meets unmanaged state without `--takeover` now fails
  loudly with the stop diagnostic (Protocol environments §9.1, §9.5).
  Path-root reinstall behaviour is unchanged.

- The `Install pinned Rust toolchain via rustup` lane step now prints a
  runner diagnostics block before its remedy note when no probed location
  holds `rustup`: `RUNNER_NAME`, `RUNNER_OS`, `hostname`, `whoami`, `HOME`,
  `CARGO_HOME` (set or defaulted), `HOMEBREW_PREFIX`, the searched `PATH`,
  one executable/exists-not-executable/absent line per probed candidate, a
  bounded rust/cargo listing of the searched bin directories, and
  `command -v` / `type -a` for `rustup`. The bare remedy sentence could not
  distinguish a different runner, a different service user, or a `rustup`
  outside the probed paths (rose-air runs 35663049586, 35725359745); the
  next red run names its evidence. Exit code, remedy sentence, and the
  success path are unchanged.

## 0.12.5 - 2026-07-14

### Added

- Shared conformance coverage for canonical, legacy, dual-file, conflict,
  invalid-manifest, and runtime-fallback resolution.

### Changed

- `agent-skill.json` is now the implementation-neutral canonical skill
  manifest filename; `csk-skill.json` remains a protocol 1.x read alias.
- Diagnostics and authoring guidance now point new packages to the canonical
  filename.

### Fixed

- Dual manifests are validated independently and accepted only when their JSON
  values are equal; mismatches fail closed with
  `conflicting_skill_manifests`.
- An invalid modern manifest no longer falls through to another filename or
  `agents/runtime.json`.

## 0.12.4 - 2026-07-13

### Added

- Idempotent `bootstrap --if-missing` for repository-managed onboarding.
- Self-contained POSIX and Windows command launchers that carry skill and
  declared system dependency paths without shell profile setup.

### Changed

- `upgrade` fetches only the selected project or global dependency closure and
  deduplicates repositories shared by multi-project operations.
- Install and upgrade dry runs use temporary resolution state and no longer
  mutate persistent source checkouts, caches, audit or registry state, runtime
  state, configuration, or installation artifacts.
