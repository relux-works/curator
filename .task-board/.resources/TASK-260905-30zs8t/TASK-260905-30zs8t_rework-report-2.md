# TASK-260905-30zs8t — rework report 2 (F8, F9)

Branch `feat/agent-environments-stage-a`, head `314ae748` on top of the
cycle-2 head `ac9d0037` (one signed commit, `G Ivan Oparin <oparin@me.com>`).
No rewrite, no push, no PR. Worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`.

Authority: curator-spec main `f39f4a9` (`protocol/environments.md` rev 1.1,
`schemas/v1`, `conformance/v1`, `cli/curator.md`, Decision 0012).

## Finding → disposition

### F8 (blocking) — canonical source identity — FIXED

`canonicalGit` now normalizes through `internal/identity.Parse` at every
boundary the brief names:

- `Install` (`envprofile.go`): the install record `Source.Git` stores the
  canonical identity; a malformed network source is rejected with
  `profile_source_invalid` before any fetch.
- Requirement reader (`gitsource.go` `packageOf`): every
  `requires.{contexts,skills,mcp}` source canonicalizes on entry to
  resolution, so two spellings never produce a spurious
  `context_source_mismatch`.
- `migrateGlobalSkills` (`envprofile.go`): migrated skill members pin the
  canonical identity; malformed sources rejected.
- Agreement check (`contextresolve.go` `selectName`): comparison is over
  `canonicalForAgreement` (core §6.1 identity; file:// compares as raw so
  distinct test remotes stay distinct).
- `gitManager.Identity` returns the canonical identity and records the
  raw→canonical mapping; the cache (`repoDir`) is keyed by canonical, so
  four spellings share one clone and one store entry. `ensureRepo` clones
  through the recorded raw when present, else `https://canonical`.
- Marker (`switch.go` `markerSource`): renders the canonical identity;
  old records carrying a raw URL heal on the next switch.

File:// remotes carry no network identity (`identity.Parse` returns "") and
pass through as their trimmed raw URL: the hermetic test shim, bypassing
the network allowlist like any local source, keying its own cache entry.
Production git sources are network identities (`host/path`).

### F9 (blocking) — source allowlist + strict audit — FIXED

- `gitManager` carries `allowedSources`; `ensureRepo` gates every clone
  through `gateSource` (the `closure.gateSource` shape: empty permits all,
  local/file bypasses, otherwise `identity.Allowed` segment-aware) with
  `profile_source_invalid` before any network clone. `Install` and
  `UpdateWithPolicy` attach the policy; `migrateGlobalSkills` loads it from
  the manager-home config file.
- `contextresolve.Input.MCPAllowlist` is populated from the policy on both
  install and update paths, so `mcp_package_not_allowed` can fire in
  production and in tests (previously no producer anywhere).
- `strictAuditMember` runs the manager §7 audit in strict mode over every
  member: `audit.CanaryPasses` (always blocking), raw-tree hashing via
  `hashing.ContentSHA256`, `audit.RevocationFor` over the canonical
  identity + raw clone URL (state hash for path members), and the
  deterministic detectors via `audit.Gate` with a forced-strict config —
  regardless of any enabled flag. An advisory profile install does not
  exist. `auditAndStore` and the update pre-check both call it; migrated
  skills call it too.
- The CLI (`cmd/curator/profile.go`) passes `allowed_sources` and
  `audit.revocations` from the loaded machine configuration into install
  and update.
- New `audit` exports: `RevocationFor` (pure revocation matcher) and
  `CanaryPasses` (static-canary probe).

### F2 non-blocking note (global-skill migration warning)

Recorded, not implemented, per the brief: an operator is not warned when a
branch-pinned or local global skill does not migrate. For the stage that
wires the operator surface.

### F7 (recorded bound, unchanged)

Scoped `secret_material_waivers` still pass `nil`; no machine-config
surface until manager-config schema 2.

## Stated bounds in stage (a)

- The MCP package allowlist has no schema-1 machine-config surface
  (manager-config schema 2, environments §12.1). The CLI passes an empty
  list (permits all); tests inject via `Policy`. Stated in the `Policy`
  doc and here, not silent.
- File:// git remotes are a test-only shim (no network identity,
  schema-invalid locks/markers, allowlist bypass). Production sources are
  canonical network identities. New network-identity tests use
  `insteadOf` and assert schema-shaped identities.
- `fetchRaw` first-spelling-wins: every spelling of one repository clones
  the same bytes; transport selection follows the first observed spelling.

## New tests (all through production entry points)

`internal/envprofile/envprofile_f8f9_test.go`:

- `TestCanonicalIdentityUnifiesSpellings` — one repo under `https`
  (uppercase host + `.git`), scp-style, and `ssh://` via `insteadOf`:
  one canonical `example.com/acme` in the lock and the marker, one
  `lock_sha256`, `lock.Validate` + `marker.Validate` + `ValidCanonical`.
- `TestCanonicalIdentityRejectsMalformedSource` — explicit-port URL
  refused with `profile_source_invalid` + the port reason, no clone.
- `TestSourceAllowlistRefusesBeforeClone` — outside `allowed_sources`
  refused with the allowlist reason, `profile-repos` empty.
- `TestRevokedSourceIsRefused` — `source:` revocation refused with the
  revocation reason.
- `TestMCPAllowlistRefusesOutsidePackage` — outside MCP allowlist refused
  with `mcp_package_not_allowed` via `Install` → `Resolve`.
- `TestStrictAuditCanaryPasses` — canary precondition.
- `TestMigratedSkillSourceIsCanonical` — migration boundary carries the
  source identity.

`internal/contextresolve/contextresolve_test.go`:

- `TestCanonicalSourceAgreementUnifiesSpellings` — https vs scp spellings
  of one identity agree and resolve; different hosts stay
  `context_source_mismatch`. Production call site: `Resolve`.

## Mutant table (each run, each kills a named test)

| Mutant | What it narrows the gate to | Named test that fails |
|---|---|---|
| F8: `canonicalGit` parses but returns the raw operand (old behaviour) | Admits every spelling as a distinct identity | `TestCanonicalIdentityUnifiesSpellings` FAILS (lock source `https://EXAMPLE.com/acme.git`, want `example.com/acme`) |
| F8: agreement compares raw declared strings (no `canonicalForAgreement`) | Admits spelling disagreement as mismatch | `TestCanonicalSourceAgreementUnifiesSpellings` FAILS (`context_source_mismatch` on one identity) |
| F8: malformed operand passed through (no `Parse` error) | Admits malformed network sources to the clone | `TestCanonicalIdentityRejectsMalformedSource` FAILS (no port reason) |
| F9: `gateSource` always allows | Admits outside-allowlist sources to the clone | `TestSourceAllowlistRefusesBeforeClone` FAILS (clone attempted, no allowlist reason) |
| F9: `strictAuditMember` result ignored in `auditAndStore` (old behaviour: `Detect`-only) | Admits revoked sources | `TestRevokedSourceIsRefused` FAILS (`err = <nil>`) |
| F9: `Input.MCPAllowlist` not populated from policy | Admits outside-allowlist MCP packages | `TestMCPAllowlistRefusesOutsidePackage` FAILS (`err = <nil>`) |
| F9: `runStaticCanary` forced false | Detectors produce nothing | `TestStrictAuditCanaryPasses` FAILS and `TestCanaryFires` (audit) FAILS |

No survivors. (A narrower revocation-only mutant disabling just the
explicit `RevocationFor` still blocks via `audit.Gate`'s own revocation
check — defense in depth, not an admission.)

## Gate outputs (real exit codes, standalone processes)

- `go build ./...` — 0 (also `GOOS=windows` 0, `GOOS=linux` 0)
- `go vet ./...` — 0
- `gofmt -l cmd internal` — clean
- `golangci-lint run` on envprofile + contextresolve + audit — 0 issues
- `go test -count=1 -race` on envprofile, contextresolve, audit, identity
  — all `ok` (envprofile 29.3s)
- Vector families with
  `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1` — 5 top-level
  PASS (detectors, header, monolithic, resolution, versions), exactly 7
  stage-deferred sub-skips (3 referenced-* + 4 mcp-*), 0 FAIL
- `bash .github/ci/gate-selftest.sh` — 81 passed, 0 failed (exit 0)
- `bash .github/ci/no-broad-suppression.sh` — ok (exit 0)
- `bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger` — ok,
  103 rows across linux/darwin/windows (exit 0)
- `go test -count=1 -timeout 30m ./cmd/curator/` — `ok … 273.524s`
  (exit 0, first-hand in this session)

## Files

- `cmd/curator/profile.go` — pass policy from machine config
- `internal/audit/audit.go` — `RevocationFor`, `CanaryPasses` exports
- `internal/contextresolve/contextresolve.go` — canonical agreement +
  helper
- `internal/contextresolve/contextresolve_test.go` — agreement test
- `internal/envprofile/envprofile.go` — `Policy`, canonical install /
  migration, `UpdateWithPolicy`, strict audit, file policy loader
- `internal/envprofile/gitsource.go` — canonical cache + raw mapping,
  allowlist gate, canonical requirement reader
- `internal/envprofile/switch.go` — canonical marker source
- `internal/envprofile/envprofile_f8f9_test.go` — new (7 tests)

Signed commit `314ae748` (`G`). No push, no tag, no PR.
