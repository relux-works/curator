# TASK-260905-30zs8t — rework report 3 (F10, F11, F12)

Branch `feat/agent-environments-stage-a`, head `dea3f5ac` on top of the
cycle-3 head `314ae748` (one signed commit, `G oparin@me.com`, 10 files,
+752/−147). No rewrite, no push, no tag, no PR. Worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`.

Authority: curator-spec main `f39f4a9` (`protocol/environments.md` rev 1.1,
`schemas/v1`, `conformance/v1`, `cli/curator.md`, Decision 0012).

## Finding → disposition

### F10 (blocking) — migration enforces the loaded machine policy — FIXED

The migration path no longer re-parses configuration into a weaker copy:

- `envprofile.PolicyFromConfig(cfg)` derives `{AllowedSources, Revocations}`
  from the CLI's already-loaded `cfg` (system overlay with locked keys and
  the `CURATOR_CONFIG` override included).
- The CLI threads it into every profile operation: `Install` and
  `UpdateWithPolicy` (as before), plus now `ListWithPolicy`,
  `UseWithPolicy`, `SyncWithPolicy` (`cmd/curator/profile.go`; `profile
  update --all` lists under the same policy).
- `ensureDefault`/`migrateGlobalSkills` take the policy explicitly; new
  `EnsureDefaultWithPolicy`, `ListWithPolicy`, `UseWithPolicy`,
  `SyncWithPolicy` carry it, and `updateLocked`/`resyncCurrentScopes` pass it
  through.
- The bare entry points (`EnsureDefault`, `List`, `Use`, `Sync`, `Update`)
  resolve the policy through `loadMachinePolicy`, which goes through
  `config.Load(config.UserPath(), nil)` — system overlay included. A missing
  file yields an empty policy (absence is legitimate); any other read or
  parse failure fails the caller with `profile_source_invalid` — a failed
  read is never an empty policy. Production always calls with
  `home == cfg.Home()`; library callers on any other home must use the
  `WithPolicy` variants (stated on each bare entry point).
- `loadPolicyForHome` (direct `home/config.json` read, no error return) is
  deleted.

Regression cover (all drive production entry points on isolated homes):

- `TestMigrationHonoursSystemLockedAllowlist` — bare `EnsureDefault` with
  `CURATOR_CONFIG` permissive and `CURATOR_SYSTEM_CONFIG` locking
  `allowed_sources` (+ `audit`): refused with the allowlist reason, no
  profile record, no clone.
- `TestMigrationHonoursSystemLockedRevocation` — system locks the allowlist
  permissive and revokes the skill identity: refused with the revocation
  reason (proves migration-path revocation through the loader).
- `TestMigrationSucceedsWhenSystemPolicyPermits` — mirror success case.
- `TestMigrationHonoursConfigPathOverride` — bare `List` with
  `CURATOR_CONFIG=<home>/curator.json` (no `config.json`): the override
  gates apply.
- `TestMachinePolicyLoadFailureFailsMigration` — malformed config fails the
  migration with the configuration reason.
- `TestProfileListMigrationHonoursSystemPolicy` (`cmd/curator`, through
  `run()` with a real `fileConfigSource` + `CURATOR_SYSTEM_CONFIG`):
  `profile list` refused — proves the CLI threading end to end.

### F11 (major) — narrowing tests for the F9 gates — FIXED

- `TestSourceAllowlistRejectsAdjacentPathSegment` — allowlist
  `example.com/org` vs operand `example.com/org-evil/pkg` (same host,
  adjacent segment), refused with the allowlist reason before any clone.
  Kills the host-only-match mutant (M-F11-hostonly).
- `TestRevokedSkillMemberIsRefused` / `TestRevokedMCPMemberIsRefused` — a
  revoked `skill` / `mcp` member required by a clean root, refused with the
  revocation reason through `Install`. Kill the kind-gated-revocation mutant
  (M-F11-kindgated).
- `TestStrictAuditCanaryFailureBlocksInstall` — drives `Install` with the
  canary forced to fail through a `canaryPasses` seam on
  `strictAuditMember`, asserting refusal with the canary reason. Kills the
  guard-removed mutant (M-F11-nocanary). The seam is documented at its
  declaration; production behavior is unchanged.
- `TestMigratedSkillSourceIsCanonical` fixed as specified: it now declares
  `https://EXAMPLE.com/skills/hello.git`, migrates through
  `EnsureDefaultWithPolicy`, and asserts the canonical
  `example.com/skills/hello` plus `identity.ValidCanonical` plus
  `lock.Validate()`. The old assertion on the raw `file://` URL (which the
  old `canonicalGit` passed unchanged) is gone.

### F12 (major) — `file://` git operands rejected — FIXED (full rejection)

- `canonicalGit` rejects any `file:`-prefixed operand (any case) with
  `profile_source_invalid` ("carries no network identity"). This is the same
  boundary that rejects malformed network sources, and it covers the
  operand (`Install`, `rootInput`), the requirement reader (`Identity`,
  hence every transitive `requires.*.git`), and `migrateGlobalSkills`.
- `gateSource` rejects `file://` as well (defense in depth at the clone
  boundary); the old allowlist bypass for it is gone.
- All `file://`-based tests converted to `insteadOf` fixtures serving fake
  network identities from local repos (new shared helper
  `newGitIdentities`/`serve` in `network_fixture_test.go`):
  `TestInstallGitResolvesNetworkIdentity` (renamed from
  `TestInstallGitFromFileRemote`, now also asserts the canonical lock
  source + `lock.Validate()`),
  `TestInstallSurfacesUnresolvedMCPCommand`,
  `TestInstallDirectoryRejectsSecretUnderSubdirectory`,
  `TestInstallTransitiveDirectoryRejectsSecret`,
  `TestEnsureDefaultMigratesGlobalSkills` (now via `EnsureDefaultWithPolicy`
  + canonical assertion), `TestRevokedSourceIsRefused`,
  `TestMCPAllowlistRefusesOutsidePackage`,
  `TestMigratedSkillSourceIsCanonical`,
  `TestProfileUpdatePinnedTagIsUnchanged` (CLI). The only remaining `file://`
  strings are refusal-test inputs and the local side of `insteadOf`
  rewrites. The CLI-gate-only fallback was not needed.
- Docs updated: `gitsource.go` (cache keying, `ensureRepo`, `Identity`,
  `packageOf`, `canonicalRequirementSource`), `canonicalGit`,
  `contextresolve.go` (`canonicalForAgreement`),
  `switch.go` (`markerSource` — pre-rejection records pass through as
  written; such records can only predate the rejection).
- Package-test hermeticity: new `TestMain` in `network_fixture_test.go`
  pins `CURATOR_CONFIG` at an absent path and neutralizes
  `CURATOR_SYSTEM_CONFIG`, so the loader never observes the operator's real
  configuration; loader tests override with `t.Setenv`.

## Mutant table (each run in a scratch copy, each kills a named test)

| Mutant (gate stays present, admits one class member) | Named test that fails |
|---|---|
| F10: loader reads the user file but drops the system overlay | `TestMigrationHonoursSystemLockedAllowlist` FAILS (skill migrates) |
| F10: loader maps a parse failure to an empty policy | `TestMachinePolicyLoadFailureFailsMigration` FAILS (migration succeeds) |
| F11: `gateSource` matches on the host only, ignoring the path | `TestSourceAllowlistRejectsAdjacentPathSegment` FAILS (install succeeds) |
| F11: revocation applies only when `Kind == context` | `TestRevokedSkillMemberIsRefused` + `TestRevokedMCPMemberIsRefused` FAIL (both install) |
| F11: canary guard removed from `strictAuditMember` | `TestStrictAuditCanaryFailureBlocksInstall` FAILS (install succeeds) |
| F12: `canonicalGit` passes `file://` through (old behaviour) | `TestCanonicalGitRejectsFileRemote` FAILS (operand admitted at the boundary; the `Install`-level tests still hold through the `gateSource` backstop — that layering is why the boundary carries its own pinning test) |
| F12: `gateSource` bypasses `file://` (old behaviour) | `TestGateSourceRejectsFileRemote` FAILS |

No survivors. New layer-pinning tests: `TestCanonicalGitRejectsFileRemote`
(`file:///`, `FILE:///`, `file://host/…`),
`TestGateSourceRejectsFileRemote` (fires even with an empty allowlist);
production-path tests: `TestFileOperandIsRefused` (both cases, no record,
no clone), `TestFileRequirementSourceIsRefused` (transitive, no dependency
clone), `TestProfileInstallFileOperandIsRefused` (CLI).

## Gate outputs (real exit codes, standalone processes, this session)

- `go build ./...` — 0 (also `GOOS=windows` 0, `GOOS=linux` 0)
- `go vet ./...` — 0
- `gofmt -l cmd internal` — clean
- `golangci-lint run` on envprofile + contextresolve + cmd/curator — 0 issues
- `go test -count=1 -race` on envprofile, contextresolve, identity, audit —
  all `ok` (envprofile 24.9s)
- Vector families with
  `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1` —
  `TestConformanceContextDetectors`, `TestConformanceEnvironmentsHeader`,
  `TestConformanceEnvironmentsMonolithic`, `TestConformanceContextResolution`,
  `TestConformanceContextVersions` all PASS; exactly the 7 known
  stage-deferred sub-skips (3 referenced-* + 4 mcp-*), 0 FAIL
- `bash .github/ci/gate-selftest.sh` — 81 passed, 0 failed (exit 0)
- `bash .github/ci/no-broad-suppression.sh` — ok (exit 0)
- `bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger` — ok,
  103 rows across linux/darwin/windows (exit 0)
- Full suite in bounded sequential calls (no backgrounding): internal
  batches 1–4 + atomicity all `ok`; `./cmd/curator` in four `-run` quarters
  (`[A-C]` 185.5s, `[D-I]` 39.9s, `[M-S]` 33.1s, `[T-Z]` 0.4s) all `ok`
  (exit 0 each) — 110 tests, no skips added
- `platform-case-gate.sh` on a recorded `-json` stream of envprofile +
  contextresolve + interop: zero in-stream failures; 7 skips recorded, all
  `stage-deferred`/`tolerated-by-ledger` (the known mcp/referenced sets);
  the 78 `required case never ran` rows are packages outside the scoped
  stream — the full matrix remains CI's hosted job, as in cycle 3

## Stated bounds (kept, plus the reviewer's note)

- MCP package allowlist still has no schema-1 machine-config surface
  (manager-config schema 2, out of scope); CLI passes empty (permits all);
  enforcement path exists and is tested. Unchanged, judged defensible.
- `fetchRaw` first-spelling-wins unchanged; added the reviewer's note to the
  `gitManager` doc: a transitive requirement declared `git@host:org/dep` is
  canonicalized before any raw is recorded and therefore clones over
  `https://`, which can surprise an SSH-only operator.
- The old `file:// test-only shim` bound is **withdrawn and replaced by the
  rejection** — no operator-reachable path produces a schema-invalid
  artifact anymore.
- F7 scoped waivers and the F2 migration warning stay carried to the stage
  that lands the operator surfaces, as before.

## Files

- `internal/envprofile/envprofile.go` — `PolicyFromConfig`,
  `loadMachinePolicy` (replaces `loadPolicyForHome`), policy threading
  through `ensureDefault`/`migrateGlobalSkills`/`updateLocked`/
  `resyncCurrentScopes`, `EnsureDefaultWithPolicy`, `ListWithPolicy`,
  `Update` loading, `canonicalGit` file rejection + `isFileRemote`,
  `canaryPasses` seam
- `internal/envprofile/gitsource.go` — `gateSource` file rejection, doc
  updates (cache, `ensureRepo`, `Identity`, `packageOf`,
  `canonicalRequirementSource`)
- `internal/envprofile/switch.go` — `UseWithPolicy`, `SyncWithPolicy`,
  `markerSource` doc
- `internal/contextresolve/contextresolve.go` — `canonicalForAgreement` doc
- `cmd/curator/profile.go` — policy threaded into list/use/update/sync
  (install already was)
- Tests: new `envprofile_f10f11f12_test.go` (5 F10 + 4 F11 + 4 F12),
  new `network_fixture_test.go` (`TestMain` isolation + `gitIdentities`
  helper); converted `envprofile_test.go`, `envprofile_f8f9_test.go`,
  `cmd/curator/profile_test.go` (+ `TestProfileInstallFileOperandIsRefused`,
  `TestProfileListMigrationHonoursSystemPolicy`)

Signed commit `dea3f5ac` (`G`). No push, no tag, no PR. Nothing written
into the control root.
