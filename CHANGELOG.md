# Changelog

All notable implementation changes are recorded here.

## Unreleased

### Added

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
