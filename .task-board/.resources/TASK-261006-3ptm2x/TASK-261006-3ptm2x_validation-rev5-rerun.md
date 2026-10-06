# TASK-261006-3ptm2x rework 3 — recovery-run re-verification (RUN-261006-4c7a0b)

Recovery retry of RUN-261006-c36be2 (rework 3, rev5 candidate). The worktree
already contained the complete fix, tests, and CHANGELOG entry; this run made
zero production edits and verified the tree as found.

## Fix confirmed present (read, not inferred)

1. `internal/envprofile/store_boundary.go:222` — `opaquescan.RefuseNULV1(entry,
   version)` runs before the commit/state branch split in
   `storeEntryPinHashes`. `expectedStoreHash` stays first (Git tree-OID
   lookup / string format; computes no v1 identity), preserving the
   `pinned commit object unavailable` repair routing. Version source:
   `lock.ContentHashVersion()` via the `version` param — never digest bytes.
2. `internal/envprofile/switch.go:644` — `loadMaterial` guards each context
   member's full store entry before manifest/module reads. Covers all three
   production callers: `assembleHome` (managed.go:303), `materializeScope`
   (switch.go:481), `preflightInPlaceScope` (switch.go:518).
3. `Resolve` ordering: `validateNamedStorePins` (managed.go:2504) precedes
   `verifyHome`/`assembleHome` (2526/2586/2615), so the pin_hash refusal
   fires before any materialized v1 identity.

## Sibling-branch sweep (re-checked in this tree)

- `extractPinnedSnapshot` / `publishRebuiltStoreEntry`: `contextstore.ContentHash`
  is a writer-version guard+hash pair — no lock-version v1 hash. Held.
- `managed.go:skillsOf`: guards `packageRoot` before hashing. Held.
- `managed.go:mcpsOf`: manifest load only, no content identity. Held.
- `surfaceHash` callers: lock-versioned; inputs guarded upstream. Held.
- `contextaudit.DetectAtVersion`: v1→scan, v2→skip, unknown→refuse. Held.
- `internal/contextlock`: zero content-hash calls. Held.
- `identity_migration.go`, `status.go`: route via fixed
  `validateProfileStoreState`. Held.
- `audit.go` gate (205), `auditSubject` (312): v1-only scans; `detect`
  (632) is the frozen-v1 helper whose only production caller is the
  synthetic `runStaticCanary`. Held.
- `gitops` tree OIDs: separate algorithm, single caller is the fixed
  function. Held as classification, not an exemption.

## Tests re-run in THIS run (mini-build-lock, GOFLAGS=-work, real exit codes)

syspolicyd successive crashes 26 -> 26 across the session (no movement).

| command | go exit | time |
|---|---|---|
| `go test ./internal/envprofile -run 'TestResolveRefusesV1CommitPinnedContextNUL\|TestResolveCleanV1CommitContextStaysCurrent\|TestResolveAdmitsV2CommitPinnedContextNUL\|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL'` | 0 | 4.3s |
| `go test ./internal/envprofile -run 'NUL\|Nul\|Opaque\|GuardedReaders\|StrictAuditMember\|ValidateNamedStorePins\|SkillsOf'` | 0 | 7.7s |
| `go test ./internal/opaquescan ./internal/hashing ./internal/contextaudit` | 0 | 0.3+0.6+0.9s |
| `go test ./internal/audit -run 'NUL\|Nul\|Opaque\|Guard\|Frozen\|Pin\|Verdict\|Gate\|SourceAudit\|V1\|V2'` | 0 | 1.0s |
| `go vet` (envprofile, audit, contextaudit, opaquescan, hashing) | 0 | — |
| `go build ./...` | 0 | — |
| `gofmt -l` (all touched trees) | clean | — |

## Mutant re-executed in THIS run (then fully restored)

M1+M2 (= full rev4 bypass): guard narrowed to state branch only +
`loadMaterial` guard dropped. `go test` exit 1: both Resolve NUL shapes
fail — module shape returns a non-empty launch fragment with nil error
and computes 2 v1 identities, exactly the reviewer's finding; both
switch NUL shapes fail; clean controls pass. Restoration verified by
sha256 match on both files plus a repo-wide `MUTANT` grep (empty) and a
green re-run (exit 0). M1-alone pin-routing kill accepted from the
attached rev5 validation (not re-executed here).

## Reran vs accepted

- Reran here: all rows above, including the M1+M2 kill.
- Accepted from attached rev5/rev4 evidence (not re-runnable locally per
  R194): `cmd/curator` and `internal/install` suites and the hosted gate;
  M1-alone and M2-alone mutants; rev1–rev3 regression history.

## Other

- CHANGELOG `Unreleased` entry present (commit-pinned v1 contexts).
- No LOGBOOK.md or `scripts/remote-gate.sh` edits (absent from git status).
- Work left uncommitted in the story worktree for handoff snapshot.
