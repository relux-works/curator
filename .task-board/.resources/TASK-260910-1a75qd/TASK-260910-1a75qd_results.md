# TASK-260910-1a75qd results — source lock model and validation

Producer: developer. Base: story worktree `task-board/story/STORY-260910-3vxe3y`
at `12f1287ee0fb538f9ca004dd53b870e824e5baf2`, uncommitted diff limited to the
new package `internal/sourcelock/` (5 files, +2269 lines). No other package
touched; frozen v1 schemas untouched; draft/opt-in only.

Contract: curator-spec main `871d11b` (`protocol/skillfile-sources.md`,
`protocol/repository-transport.md` revision 1, `schemas/draft-sources-v1/*`,
`conformance/draft-sources-v1/*`). Revision 2 transport (ports/mirrors/
aliases) excluded: the lock carries no endpoint properties at all, so v2
needs no lock change.

## What was built

New package `internal/sourcelock` (`sourcelock.go`, `bindings.go`).
`internal/managerlock` was deliberately NOT reused: it owns OS file locks,
and putting the package lock there would invite exactly the confusion the
acceptance forbids. `protocoljson` needed no change: CCJ-1 + `Validate`
already supply the digest path, and the lock consumes them.

- `Package`: disjoint `local-snapshot` / `network-git` / `configured-git`
  arms with per-arm constructors, strict validation, CCJ-1 `Canonical` and
  content `Digest`. Cross-arm fields are rejected; a snapshot digest can
  never land in a commit field (commit hex is bare, format-length-bound).
- `Lock` + `Member`: `Skillfile.lock.json` (schema 1) with UTF-8-name
  ordering, root selection indexes vs null transitive selection,
  member/package directory agreement for both Git arms, `manifest_sha256`
  over the entire parsed Skillfile, and `lock_sha256` self-integrity.
- Lifecycle: `New` (sort + seal), `Parse` (strict, unknown fields
  rejected, structure before integrity), `Read`/`Write` (canonical bytes,
  atomic, 0644), `PathIn`, `Find`.
- Gates: `CheckStale`/`CheckStaleDigest` (`source_lock_stale`),
  `CheckNames` / `CheckMembership` (exact frozen-set comparison:
  `source_member_missing`, `source_member_invalid`, `source_lock_stale`).
- `Bindings`: machine-private per-alias record (canonical absolute
  location + root-input snapshot of the operator's source-policy config
  for that lock generation), 0600 atomic persistence, `CheckFresh`
  generation binding. Source-policy.json itself stays owned by its own
  task; bindings only snapshot what refresh compares.
- Context-lock boundary: package doc + `TestContextLockConfusion` prove
  the two locks reject each other's bytes both directions.

Diagnostic mapping (spec §5 classes): envelope/shape/selection/integrity
→ `source_selection_invalid`; member/package records and directory
disagreement → `source_member_invalid`; duplicate (incl. case-folded)
names → `source_name_conflict`; manifest/bindings/plan drift →
`source_lock_stale`; plan member outside the lock →
`source_member_missing`; bad binding alias → `source_alias_unknown`.

## Acceptance coverage (each row driven through a production entry point)

| Acceptance row | Test (production call site) |
|---|---|
| Persist Skillfile.lock.json, exact identities, frozen selection | `TestWriteReadRoundTrip` (`Write`→`Read`), `TestNewSortsMembersAndSealsDigest` (`New`/`Digest`) |
| Machine path binding separated from portable identity | `TestMachineSeparation` (`Object` key sets + byte scan), malformed `location`/`endpoint`/`url` cases via `Parse`, `TestBindingsRoundTrip` |
| Stale lock | `TestCheckStale` (`CheckStale`, changed tag → `source_lock_stale`; whitespace-only → fresh) |
| Malformed lock | `TestMalformedLocks` (24 cases via `Parse`), `TestSchemaFixtures` (4 invalid fixtures fail structurally) |
| Package membership | `TestCheckNames`, `TestCheckMembership`, `TestDuplicateNamesConflict`, `TestUnsortedMembersRejected`, `TestEmptyMembers`, `TestFind` |
| Three arms distinguished | `TestThreeArmsRoundTrip`, `TestCrossArmConfusionRejected` (17 cases), `TestCommitValidation`, `TestRepositoryValidation`, `TestDirectoryAgreement` |
| Context lock never confused | `TestContextLockConfusion` (`Parse` ↔ `contextlock.Parse` cross-rejection) |
| Rebinding preserves identity | `TestRebindingPreservesIdentity` (two locations, one `Digest`, `CheckFresh`) |
| Selection 0 vs null, shared collection index | `TestSelectionSemantics` |

Golden digests (manifest, lock, 3 package identities) were cross-checked
against an independent Python CCJ-1 implementation (`/tmp/ccj_check.py`)
before being pinned; Go and Python agree byte for byte.

## Gates (real exit codes, narrow scope per host rules)

- `go test -count=1 ./internal/sourcelock/` → exit 0 (`ok`, 76 run
  lines incl. subtests, 0 failures). Neighbor packages not rerun: the
  diff adds a new package and edits no existing file.
- `go vet ./internal/sourcelock/` → exit 0.
- `gofmt -l internal/sourcelock/` → clean (no output).
- `golangci-lint run ./internal/sourcelock/` (configured gate:
  misspell/revive/gosec) → exit 0, `0 issues` (two interim revive
  unused-parameter findings fixed in test table closures).
- `go build ./...` → exit 0 (module builds; whole-module `go test` not
  run per the narrow-test host rule).

## Mutants (bytes restored, `cmp` clean afterwards)

- M1 narrowing: directory agreement on network-git only → `go test
  -run TestDirectoryAgreement` FAILS (exit 1). Killed.
- M2 narrowing: duplicate detection exact-only (no case fold) → `go test
  -run TestDuplicateNamesConflict` FAILS (exit 1). Killed.
- M3 deletion: skip `lock_sha256` comparison → tampered-member + fixture
  integrity assertions FAIL (exit 1). Killed.
- M4 narrowing: `CheckNames` ignores extra locked members → `go test
  -run TestCheckNames` FAILS (exit 1). Killed.
- Survivors: none.

## Notable decision (recorded here instead of LOGBOOK.md, which the
campaign boundary forbids me to edit)

The normative sentence "for both Git identity kinds these two directory
fields MUST agree" is enforced: member.directory must equal
package.directory for `network-git` and `configured-git`. The
`valid-git` / `valid-configured-git` schema fixtures disagree
(`agents/skills/review` vs `.`) and carry a placeholder `lock_sha256`;
they pin JSON-schema shapes (JSON Schema cannot express cross-field
equality), not semantic validity, so the implementation rejects that
shape with `source_member_invalid`. `TestSchemaFixtures` pins this
stance; a reviewer disagreeing with the reading should say so before
the refresh task builds on it.

## Bounds

- Snapshot capture/availability (`source_snapshot_changed`,
  `source_snapshot_unavailable`), refresh orchestration, markers,
  receipts, registry evidence, and source-policy parsing belong to
  sibling tasks; this package exposes the lock representation and
  validation they consume.
- No live network, credential, or runtime-home use; all tests hermetic
  (`t.TempDir`, inline fixture copies with provenance comments).

---

# Revision 2 (2026-09-16) — Windows CI repair

CR-1 rev1 remote gate failed ONLY on windows-latest (`go test` exit 1;
ubuntu/macos/lint/race/conformance all green). Root cause, verified
against the Go toolchain source
(`internal/filepathlite/path_windows.go: IsAbs` requires a volume
name): the bindings tests used Unix-only location literals
(`/work/skills`, `/work/vendor/kit`, `/elsewhere`, `/machine-a/...`),
which are not absolute paths on Windows, so `NewBindings` rejected
them and the fixtures hit `t.Fatalf`. The mode assertions
(`Perm() == 0o600/0o644`) are likewise unrepresentable on Windows,
which honors only the owner-write bit. Production code was already
correct (platform-native `filepath.IsAbs`/`Clean` for the
machine-private record); the fix is test-only, two files:

- `internal/sourcelock/bindings_test.go`: new `mustBindingsFixture`
  builds locations from `t.TempDir()` (absolute + Clean-stable on
  every platform); mode assertion guarded by
  `runtime.GOOS != "windows"` per repo convention
  (`internal/config/write_test.go`); raw-bytes location check now
  uses a CCJ-encoded needle so backslash paths match; stale-generation
  fixture uses a TempDir location; new regression test
  `TestBindingsLocationIsPlatformAbsolute`.
- `internal/sourcelock/sourcelock_test.go`: lock mode assertion
  guarded the same way; `TestRebindingPreservesIdentity` uses
  TempDir locations per machine.

Negative-path literals (`/etc/x`, `/tmp/x`, `work/kit`, `/work//kit`)
needed no change: `identifiers.PortablePath` is pure string logic,
identical on all platforms, and those cases assert rejection only.

## Gates, this run (real exit codes, narrow scope)

- `go test -count=1 ./internal/sourcelock/` → exit 0
  (`ok ... 0.444s` on final bytes).
- `go vet ./internal/sourcelock/` → exit 0.
- `gofmt -l internal/sourcelock/` → clean (no output).
- `golangci-lint run ./internal/sourcelock/` → exit 0, `0 issues`.
- `GOOS=windows go vet ./internal/sourcelock/` → exit 0.
- `GOOS=windows go test -c ./internal/sourcelock/` → exit 0
  (Windows test binary compiles; no Wine/host execution available
  here — execution evidence comes from the rev2 remote gate).
- `go build ./...` → exit 0. Whole-module `go test` not run per the
  narrow-test host rule.

## Mutants, this run (bytes restored, `cmp` clean, sha256 `2c6ba0eb…bffe` before and after)

- M1 narrowing: directory agreement on network-git only → `go test
  -run TestDirectoryAgreement` FAILS (exit 1,
  `configured-git with non-root member directory accepted`). Killed.
- M2 narrowing: duplicate detection exact-only (no case fold) →
  `go test -run TestDuplicateNamesConflict` FAILS (exit 1,
  `case-fold duplicate err = <nil>`). Killed.
- Survivors: none.

## Diff scope

`git status`: only `?? internal/sourcelock/` (5 files). No other
package touched; frozen v1 schemas untouched; draft/opt-in only.
Production files are byte-identical to the rev1 candidate — the only
delta is the two test files above.

---

# Revision 3 (2026-09-16) — reviewer P1 rework: canonical repository normalization

Reviewer verdict `TASK-260910-1a75qd_review-verdict-rev2.md`
(CHANGES_REQUESTED) reproduced that `internal/sourcelock/sourcelock.go`
`validRepository` delegated to `identity.ValidCanonical`, which checks
path grammar but not repository-transport revision 1 §1 normalization:
a lock member with repository `example.org/kit.git` was accepted by
New, Write and Read, permitting a non-canonical portable identity.

## Fix (scoped to the lock module; no legacy widening)

- `internal/sourcelock/sourcelock.go`: `validRepository` now rejects
  any value with a terminal `.git` suffix before consulting
  `identity.ValidCanonical` (which already covers lowercase host,
  case-sensitive path, and transport/username rejection). Legacy
  identity semantics elsewhere are untouched; the diagnostic stays
  `source_member_invalid` on the `.repository` path per spec §5.
- `internal/sourcelock/sourcelock_test.go`:
  - `TestRepositoryValidation`: `example.org/kit.git` and
    `github.com/acme/kit.git` added to the invalid table.
  - New `TestNoncanonicalRepositoryRejected`: `NetworkGitPackage`,
    `New`, `Write`, `Parse` and `Read` must all reject
    `example.org/kit.git` with `source_member_invalid` naming the
    repository. The persisted payload carries a recomputed
    `lock_sha256` (via the package's own digest preimage), and the
    test fails if the error is a `lock_sha256 mismatch`, so the
    rejection is credited to normalization, never integrity. `Read`
    is driven by staging the canonical bytes directly on disk
    (bypassing `Write`'s own gate). Canonical `example.org/kit`
    positive control kept.

## Gates, this run (narrow scope per host rules; shell `bash` with `set -o pipefail`, real exit codes)

- `go test -count=1 ./internal/sourcelock/` → exit 0
  (`ok ... 0.530s` on final bytes; 0.444–0.530s across runs).
- `go test -count=1 ./internal/sourcelock/ -run 'TestNoncanonicalRepositoryRejected|TestRepositoryValidation'` → exit 0.
- `go vet ./internal/sourcelock/` → exit 0.
- `gofmt -l internal/sourcelock/` → exit 0, clean (no output).
- `golangci-lint run ./internal/sourcelock/` (configured gate) → exit 0, `0 issues`.
- `go build ./...` → exit 0. Whole-module `go test` not run per the
  narrow-test host rule; neighbor packages untouched.
- `GOOS=windows go vet ./internal/sourcelock/` → exit 0.
- `git status`: only `?? internal/sourcelock/` (5 files, 2 changed
  this revision). No other package touched; frozen v1 schemas
  untouched; draft/opt-in only.

## Mutants, this run (bytes restored, `sha256sum -c` OK on `04b25953…c2616`)

- M0 deletion (authentic pre-fix failure): suffix guard removed →
  `TestRepositoryValidation` (`repository "example.org/kit.git"
  accepted`) and `TestNoncanonicalRepositoryRejected`
  (`err = <nil>`) FAIL — the rev2 hole reproduced on the new tests.
  Killed. (Observed via FAIL lines; the `| tail` pipeline masked the
  numeric exit on that one invocation — every other gate below quotes
  `${PIPESTATUS[0]}`.)
- M1 narrowing: guard weakened to `HasSuffix "/.git"` (bare component
  only) → `go test -run 'TestNoncanonicalRepositoryRejected|
  TestRepositoryValidation'` exit 1, both tests FAIL. Killed.
- M2 narrowing: guard gated to single-level paths
  (`strings.Count(value, "/") == 1 && …`) → `go test -run
  TestRepositoryValidation` exit 1,
  `repository "github.com/acme/kit.git" accepted`. Killed.
- Survivors: none.

## Scope notes

- `identity.ValidCanonical` and every other existing file are
  byte-untouched; the `.git` rule lives only in the scoped lock
  module per the rework brief. Uppercase `.GIT` stays accepted
  (path is case-sensitive; core strips exactly lowercase `.git`),
  matching `identity.Parse`'s case-sensitive `TrimSuffix`.
- Bounds from revisions 1–2 unchanged (snapshot capture, refresh
  orchestration, markers/receipts, source-policy parsing remain
  sibling work; `Bindings.Validate` checks lexical shape only).
