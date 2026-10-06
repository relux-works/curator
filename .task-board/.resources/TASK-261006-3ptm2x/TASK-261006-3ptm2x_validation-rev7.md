# TASK-261006-3ptm2x rework 3 follow-up — validation (rev7)

## Scope of this run: test-fixture fix only, product code untouched

The rev4 P1 (`v1-commit-context-bypass`) product fix from rev5 is intact and
unchanged in this run:

1. `internal/envprofile/store_boundary.go` — `storeEntryPinHashes` runs
   `opaquescan.RefuseNULV1(entry, version)` before the commit/state branch
   split (verified present, lines 215-224).
2. `internal/envprofile/switch.go` — `loadMaterial` runs the per-member
   full-entry `RefuseNULV1` guard before manifest/module reads (verified
   present).

The rev5 16-row sibling-branch sweep holds unchanged: no product file was
edited in this run (`git diff` product files identical to rev5/rev6 state;
only `internal/envprofile/nul_opaque_v1_commit_test.go` changed).

## What was fixed and why

Rev5 and rev6 hosted gates failed identically on ubuntu-latest only (macOS and
Windows green): `TestResolveCleanV1CommitContextStaysCurrent` and both
`TestResolveAdmitsV2CommitPinnedContextNUL` subtests failed at fixture setup
with `environment_home_stale: passthrough entry .credentials.json is detached:
link targets <tmp>/002/.credentials.json, expected <tmp>/...`.

Root cause (matches the orchestrator rev6 diagnosis): `commitResolveRequest`
created a NEW native home (`t.TempDir()`) on every call, and the two-call
tests (provision with `Repair=true`, then bare re-resolve) built two
requests. On Linux the claude_code `.credentials.json` passthrough is a
file link into the native home (`internal/envregistry/envregistry.go:254`),
so provisioning linked native #1 while the bare re-check expected native #2.
On macOS the credential lives in the Keychain (empty passthrough map), so
the tests passed there.

Fix in `internal/envprofile/nul_opaque_v1_commit_test.go`:
- New `commitResolveHomes` fixture: native, launch, and XDG dirs created ONCE
  per subtest via `newCommitResolveHomes(t)`; `homes.request(home, profile)`
  builds each `ResolveRequest` over the pinned dirs. All five
  `commitResolveRequest` call sites converted; the old constructor is gone
  (`grep commitResolveRequest` returns nothing).
- The pinned native home is seeded with a live `.credentials.json`, as in
  `seedLiveNativeCredentials`: harmless on macOS, keeps the Linux file-link
  passthrough live instead of detached-pending.
- `TestSwitchMaterializeRefusesV1CommitPinnedContextNUL` native seam now
  returns one fixed dir instead of a fresh `t.TempDir()` per call.

## Tests (production entry, `internal/envprofile`)

- `TestResolveRefusesV1CommitPinnedContextNUL` (module/excluded): legacy v1
  lock + commit-pinned NUL context + hand-built current v1 marker. `Resolve`
  refuses with `*storeEntryFailure` (`pin_hash`) carrying
  `audit.opaque.nul-byte`, zero v1 hashes.
- `TestResolveCleanV1CommitContextStaysCurrent`: clean v1 home provisioned
  under the v1 writer, re-resolved bare: current, fragment emitted,
  v1 counter > 0 (seam-wiring proof).
- `TestResolveAdmitsV2CommitPinnedContextNUL` (v2-module/v2-excluded):
  provisioned and re-resolved under v2: current, zero v1 hashes.
- `TestSwitchMaterializeRefusesV1CommitPinnedContextNUL`
  (module/excluded/clean): production `materializeScopeWithNativeHome`.
  NUL shapes refuse with the opaque finding and zero v1 hashes; clean
  materializes with v1 > 0.

## Mutants

- M1 (guard moved below the `member.Commit` return in
  `storeEntryPinHashes` = rev4 code): re-verified RED with the new fixture
  via temporary edit, then byte-restored (`git diff --stat` confirms the
  15-line guard is back). Both shapes FAIL with the pin-routing assertion:
  outcome degrades to `environment_home_stale` via the loadMaterial guard
  instead of the `pin_hash` refusal. Kills the mutant the rework3 brief
  names.
- M1+M2 (= full rev4 bypass) and M2 (loadMaterial guard dropped): product
  code unchanged since rev5, where both were verified red; not re-run.

## Local evidence (mini-build-lock, GOFLAGS=-work)

syspolicyd `running` throughout; successive crashes 26 -> 26 (no change).
`go test`/`go vet`/`go build` real status taken from their ok/FAIL markers
(output was tailed; pipeline exit 0 in all green rows).

| command | go-tool status | time |
|---|---|---|
| `go test ./internal/envprofile -run 'TestResolveRefusesV1CommitPinnedContextNUL\|TestResolveCleanV1CommitContextStaysCurrent\|TestResolveAdmitsV2CommitPinnedContextNUL\|TestSwitchMaterializeRefusesV1CommitPinnedContextNUL' -v` | ok, exit 0 | 7.2s |
| `go test ./internal/envprofile -run 'NUL\|Nul\|Opaque\|GuardedReaders\|StrictAuditMember\|ValidateNamedStorePins\|SkillsOf'` | ok, exit 0 | 11.6s |
| `go test ./internal/opaquescan ./internal/hashing` | ok, exit 0 | 0.6s + 0.3s |
| `go test ./internal/contextaudit` | ok, exit 0 | 0.3s |
| `go test ./internal/audit -run 'NUL\|Nul\|Opaque\|Guard\|Frozen\|Pin\|Verdict\|Gate\|SourceAudit\|V1\|V2'` | ok, exit 0 | 0.8s |
| M1 mutant: `go test ./internal/envprofile -run TestResolveRefusesV1CommitPinnedContextNUL` | FAIL, exit 1 (expected-red; mutant killed) | 1.1s |
| restore check: same test after byte-restore | ok, exit 0 | 1.1s |
| `go vet` envprofile, opaquescan, hashing, audit, contextaudit | clean, exit 0 | — |
| `go build ./...` | clean, exit 0 | — |
| `golangci-lint run internal/envprofile/...` | 0 issues, exit 0 | — |
| `gofmt -l` touched test file | clean | — |

Not run locally per R194 (hosted gate is arbiter): `cmd/curator`,
`internal/install`, and any suite creating fake executables.

## Other

- CHANGELOG `Unreleased` entry (covering commit-pinned v1 contexts) kept as
  is. No LOGBOOK or `scripts/remote-gate.sh` edits.
- Work left uncommitted in the story worktree for handoff snapshot.
