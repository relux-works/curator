## Unreleased

## v0.15.0-rc.3 — 2026-10-04

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
  for the published cases. Mixed external/local builds stage external commands
  first, receipt-2 cache keys omit execution assurance while cache lookup still
  verifies its receipt, and package-selected signing or artifact/PATH output
  destinations receive explicit refusal diagnostics.

### Fixed

- Project install materializes skills at non-git product roots with a hygiene
  notice; unexpected Git failures still refuse installation. First-run help
  works without machine configuration, missing configuration points to
  bootstrap, and inactive profiles' unprovisioned homes no longer fail
  `env status --check`.
- Global-upgrade and GC build-cache sweeps preserve binaries used by live
  processes and refuse deletion when process inspection cannot establish
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
