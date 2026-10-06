# TASK-261006-3ptm2x rework 3 — validation (rev5)

Fix for review finding `v1-commit-context-bypass` (rev4 verdict, RUN-261006-9289ae):
commit-pinned v1 context re-checks skipped the full-snapshot opaque guard and
computed/trusted v1 identities over NUL.

## Fix (2 guards, no version inference from digests)

1. `internal/envprofile/store_boundary.go` — `storeEntryPinHashes`: the
   `opaquescan.RefuseNULV1(entry, version)` scan now runs before the
   commit/state branch split, so commit-pinned members are refused before
   any trust. `expectedStoreHash` stays first: it computes no v1 identity
   (Git tree-OID lookup / string format), preserving the
   `pinned commit object unavailable` repair routing. The old comment
   claiming commit pins are "outside this rule" is replaced: the Git tree
   OID is a separate algorithm, but the same snapshot feeds lock-declared
   v1 materialized identities downstream, so the scan applies.
   Version source: `lock.ContentHashVersion()` via the `version` param.
2. `internal/envprofile/switch.go` — `loadMaterial`: per context member,
   `RefuseNULV1(entry, lock.ContentHashVersion())` over the full store
   entry runs before the manifest/module reads. This covers the three
   production callers that reach v1 surface identities without a preceding
   pin check (`assembleHome`, `materializeScope`, `preflightInPlaceScope`).
   Unknown versions fail closed inside `RefuseNULV1` (same as before).

## Sibling-branch sweep (early-return pattern)

| # | Branch | Result |
|---|--------|--------|
| 1 | `store_boundary.go:storeEntryPinHashes` commit branch | FIXED (guard before return) |
| 2 | `store_boundary.go:storeEntryPinHashes` state branch | held (guard precedes hash; unchanged order) |
| 3 | `store_boundary.go:extractPinnedSnapshot` | held: `contextstore.ContentHash` is a writer-version guard+hash pair; no lock-version v1 hash on this path |
| 4 | `store_boundary.go:publishRebuiltStoreEntry` | held: same writer-version pair, twice |
| 5 | `store_boundary.go:rebuildUntrustedStoreEntry` | held: revalidates via fixed `validateNamedStorePins` |
| 6 | `switch.go:loadMaterial` | FIXED (per-member full-entry guard) |
| 7 | `managed.go:skillsOf` | held: already guards `packageRoot` before hashing |
| 8 | `managed.go:mcpsOf` | held: manifest load only, computes no content identity |
| 9 | `managed.go:surfaceHash`, `switch.go:836,903` | held: lock-versioned; inputs guarded upstream by fixes 1-2 + skillsOf |
| 10 | `contextaudit.Detect` (frozen v1) / `DetectAtVersion` | held: version dispatch, v1 scan unconditional, no early return |
| 11 | `internal/contextlock` | held: no content-hash calls at all (lock hash is plain SHA-256, not §8 framing) |
| 12 | `contextstore.ContentHash/EnsureState/EnsureGit` | held: writer-version-consistent guard+hash |
| 13 | `identity_migration.go` | held: explicit V2 rehash of state entries; entry via fixed `validateProfileStoreState` |
| 14 | `envprofile.go:strictAuditMember` | held: already guards before revocation-identity hash |
| 15 | `status.go` store-state paths | held: via fixed `validateProfileStoreState` |
| 16 | `gitops.TreeObjectIDFromDir/PinnedCommitTree` | held as algorithm classification (Git OIDs, not §8 identities); single caller is the fixed function |

## Tests (`internal/envprofile/nul_opaque_v1_commit_test.go`, production entry)

- `TestResolveRefusesV1CommitPinnedContextNUL` (module/excluded): legacy v1
  lock + commit-pinned NUL context + hand-built current v1 marker (the
  reviewer's probe shape). `Resolve` must refuse with a `*storeEntryFailure`
  (`pin_hash`) carrying `audit.opaque.nul-byte`, with zero v1 hashes.
- `TestResolveCleanV1CommitContextStaysCurrent`: clean v1 home provisioned
  under the v1 writer, re-resolved bare under the restored writer: current,
  fragment emitted, v1 counter > 0 (seam-wiring proof).
- `TestResolveAdmitsV2CommitPinnedContextNUL` (v2-module/v2-excluded):
  provisioned and re-resolved under v2: current, zero v1 hashes.
  Controls are production-provisioned (not hand-built markers) so they stay
  consistent on every platform: on Linux, claude_code defaults to shared
  isolation with a file-link passthrough entry, which a hand-built empty
  marker cannot satisfy.
- `TestSwitchMaterializeRefusesV1CommitPinnedContextNUL`
  (module/excluded/clean): production `materializeScopeWithNativeHome`
  under the v1 writer. NUL shapes refuse with the opaque finding and zero
  v1 hashes; clean materializes with v1 > 0.

## Mutants (all verified red by temporary edit, then restored)

- M1 (guard below the Commit return in `storeEntryPinHashes`, = rev4 code):
  Resolve test FAILS both shapes — outcome degrades to stale-via-loadMaterial
  instead of the pin_hash refusal; the pin-routing assertion kills it.
  Observed: `environment_home_stale ... audit.opaque.nul-byte ...; want the
  refusal routed through the pin_hash store check`.
- M1+M2 (= full rev4 bypass): Resolve test FAILS — module shape returns a
  non-empty launch fragment with nil error and computes 2 v1 identities,
  matching the reviewer's finding exactly. Switch test NUL shapes FAIL.
- M2 (loadMaterial guard dropped): switch test FAILS both NUL shapes
  (materializes over NUL); clean still passes.
- Over-blocking direction: clean-v1 and both v2 controls fail if the NUL
  rule is applied unconditionally or to v2 (provision/refusal inversion).

## Local evidence (mini-build-lock, GOFLAGS=-work)

syspolicyd running throughout; successive crashes 26 -> 26 (no change).

| command | exit | time |
|---|---|---|
| `go test ./internal/envprofile -run 'TestResolveRefusesV1CommitPinnedContextNUL\|TestResolveCleanV1CommitContextStaysCurrent\|TestResolveAdmitsV2CommitPinnedContextNUL\|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL'` | 0 | 4.634s |
| `go test ./internal/envprofile -run 'NUL\|Nul\|Opaque\|OpaqueFile\|GuardedReaders\|StrictAuditMember\|ValidateNamedStorePins\|SkillsOf'` | 0 | 7.795s |
| `go test ./internal/envprofile -run 'TestResolve\|TestStoreBoundary\|TestUse\|TestTakeover\|TestSync\|TestForeignSymlink\|TestMaterialize'` | 0 | 86.833s |
| `go test ./internal/opaquescan ./internal/hashing` | 0 | 0.379s + 0.597s |
| `go test ./internal/contextaudit` | 0 | 0.283s |
| `go test ./internal/audit -run 'NUL\|Nul\|Opaque\|OpaqueFile\|Guard\|Frozen\|Pin\|Verdict\|Gate\|SourceAudit\|V1\|V2'` | 0 | 0.784s |
| `go vet` (envprofile, audit, contextaudit, opaquescan, hashing) | 0 | — |
| `go build ./...` | 0 | — |
| `golangci-lint run internal/envprofile/...` | 0 | 0 issues |
| `gofmt -l` (touched packages) | clean | — |
| M1 / M1+M2 / M2 mutant runs | FAIL (expected; mutants killed) | 1.1s / 1.9s / 1.3s |

Not run locally per R194 (hosted gate is arbiter): `cmd/curator`,
`internal/install`, and any suite creating fake executables.

## Other

- CHANGELOG `Unreleased` extended (commit-pinned v1 contexts refused like
  other v1 paths). No LOGBOOK or `scripts/remote-gate.sh` edits.
- Work left uncommitted in the story worktree for handoff snapshot.
