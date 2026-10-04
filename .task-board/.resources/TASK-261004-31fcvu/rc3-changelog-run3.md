# Changelog

All notable implementation changes are recorded here.

## Unreleased

## 0.15.0-rc.3 - 2026-10-04

### Added

- Muse environment adapter and `launch-env-fragment-v3` (curator#100,
  curator-spec#121): managed XDG configuration, cache, data, and state parents
  preserve `HOME`; credential passthrough is validated without copying secrets.
  Muse prompt and MCP channels remain unverified and are refused.
- `launch-env-fragment-v2` carries Decision 0018 permission selections as
  `{mode, locked, source}`, including explicit yolo and locked native policy.
  Existing explicit `env migrate` credential ownership migration remains the
  route for credential conflicts; repair does not silently migrate ownership.
- `curator env unmanage --restore-backups` restores saved native context files
  when returning to ambient management, with a read-only dry run and conflict
  checks. The restore limitations found by the inline audit are listed below.
- `curator global adopt` explicitly takes ownership of conflicting global
  command shims after showing the plan, with dry-run support and rollback.
- Operator guidance for external build repositories and a second-operator
  bootstrap/profile/environment walkthrough. README and SECURITY now explain
  portable worker bounds and the separately installed verified-provider path;
  this release ships no verified provider.
- Release-readiness, spec-review, build-leaf reconciliation, and launch-context
  research records document qualification and ownership boundaries. Public-board
  guidance keeps other operators' product material private.
- Expanded production-entry regression coverage for profile reinstall,
  collection installs, bounded MCP declaration exposure, protected-store named
  absence, and nofollow parent writes. The test harness and CI isolate user and
  system Git configuration to keep host settings out of fixtures.

### Changed

- CI conformance pin advances to curator-spec `1.0.0-rc.14`, peeled tag commit
  `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, manifest SHA-256
  `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`.
  Candidate and released families retain exact case counts and owned gap
  accounting; the pin alone is not a conformance or release qualification.
- Content-hash v2 readers and versioned marker, context, environment, and
  registry carriers are supported with frozen v1 read compatibility. Production
  hashes are still written as `curator-content-v1`; the v2 writer remains off.
- Codex seed revision A ships for the first time in a tagged warning release:
  native `config.toml` is copied whole, inherited MCP servers are recorded and
  reported as ungoverned, and provisioning warns
  `mcp_native_servers_ungoverned`. Revision B remains deferred.
- Machine `security_posture` revision A also ships for the first time in a
  tagged warning release. The default remains `permissive`, with
  `security_posture_permissive` warnings; explicitly selected `hardened`
  applies its stricter defaults and locked policy. When a trusted registry is
  unreachable, permissive install/update names artifacts without registry
  evidence in a gate notice, while hardened refuses. Default-hardened revision
  B remains deferred.
- Global profile/environment operations serialize plan revalidation and
  publication under the manager-home mutation lock, publishing configuration,
  scope records, and managed homes in transaction order.
- External-repository acquisition and install lifecycle consumers now account
  for the published cases. Malformed repository requests refuse before Git is
  started; Windows cache artifacts retain their `.exe` suffix. Mixed
  external/local builds stage external commands first, receipt-2 cache keys
  omit execution assurance while cache lookup still verifies its receipt, and
  package-selected signing or artifact/PATH output destinations receive
  explicit refusal diagnostics.

### Fixed

- Project install materializes skills at non-git product roots with a hygiene
  notice; unexpected Git failures still refuse installation. First-run help
  works without machine configuration, missing configuration points to
  bootstrap, and inactive profiles' unprovisioned homes no longer fail
  `env status --check`.
- Global-upgrade and GC build-cache sweeps preserve binaries used by live
  processes and skip deletion when process inspection cannot establish
  safety. This does not repair the separate runtime-reference audit finding N1.
- Marker v3/v4 readers reject inconsistent external repository identities,
  substitution kinds, and effective revision widths; core v5 shares these
  cross-field checks.
- Unix HTTPS askpass requests the secret only after accepting the password
  prompt, preventing broken-pipe (`EPIPE`) transport errors on refusal paths.
- Hosted/self-hosted CI follow-ups preserve the selected Go and Node paths when
  adding Rust tools, verify toolchain/shim adoption, use the explicit rose-air
  runner label, and remove hard-link assumptions from worker identity tests.
  The naming gate ignores binary-patch payloads and machine echo records.

### Security

- Managed writes refuse symlinks/reparse points in parent routes and recheck
  boundaries before publication. Profile path sources and private stores now
  validate path kinds, ownership, permissions, and containment, including
  entries that disappear during a boundary walk. Read failures remain distinct
  from absence and cannot select an absence fallback.
- Audit-registry page/checkpoint records are bound to their protected store
  boundaries. Operator bootstrap checkpoints and mirror groups provide TOFU
  and equivocation checks, with their state surfaced by status commands.
- Profile source signer allowlists and required signers are enforced; updates
  surface system-prompt and MCP deltas for confirmation. Direct-only system
  module admission and trust-root provider status now honor their policy and
  report dropped, refused, missing, and unreadable rows accurately. Existing
  warning-stage provider, hook, and MCP passthrough policies remain in effect.
- Windows executable resolution proves platform ownership and every
  component-store hard-link origin before granting the captured System32
  exception; matching file identity alone is insufficient.
- Scoped HTTPS build-repository credentials now travel through a broker pipe
  rather than the child environment, with host-pinned credential selection.
- The release installer verifies attested or signed checksums and then the
  archive digest, refusing missing verifiers or failed verification. The
  explicit `CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1` emergency bypass warns.
  The interim content audit also reports NUL-containing opaque inputs rather
  than treating them as safely scanned text.

### Known issues

- Windows broker real-Git flake `TASK-260930-fp8vx7`: two historical failures
  were not reproduced in approximately 21,000 hosted passes. No root cause is
  established; rc.3 ships this as a documented risk.
- Content-hash v2 writing is deferred to rc.4, owned by `TASK-261003-1uzji7`;
  atomic v1→v2 profile hash migration must land first. The rc.14 snapshot
  v2-write case remains an owned known gap while production writers select v1.
- The 2026-10 inline security audit found N1–N4: runtime GC with incomplete
  references, expanded-snapshot budget bypass through repeated blobs, restored
  file permissions widened, and inability to restore a saved symlink. These
  are not fixed in rc.3. See [issue #106](https://github.com/relux-works/curator/issues/106)
  and [the audit report](docs/security-audit-2026-10-inline.md); remediation is
  tracked in `STORY-261004-3oognx`.
- B3 cache-prune PRs are excluded from this release.

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
