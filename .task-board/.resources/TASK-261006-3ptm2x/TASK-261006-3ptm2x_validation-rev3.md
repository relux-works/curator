# Validation — TASK-261006-3ptm2x round 3: rev2 gate failures fixed

Supplements (does not replace) `TASK-261006-3ptm2x_validation.md` (rev2).
The rev2 candidate failed hosted gate run 37459410231 with 5 failing
leaves / 3 unique tests (`opaque-rev2-failures.md`). This round fixes all
three with no review-findings scope change. Checklist stays 15/15 with the
same recorded bounds (hosted gate arbitrates install/cmd suites).

## 1. Root causes

1. **Status plans never consult installed state.** `install.Project`
   and `install.Global` with `OperationStatus+DryRun` return at the
   dry-run step without looking at installed trees. Consequences:
   - `TestV1ReinstallRefusesNULAppearingAfterInstall` status-plan
     legs got `ok`/`planned` instead of the opaque refusal
     (`nul_opaque_v1_rework_test.go:67,74`).
   - `TestStatusCheckRefusesV1NULAppearingAfterInstall` exited
     `exitFail` via drift (`content-drift`) but with empty stderr
     (`nul_opaque_v1_rework_test.go:99`).
   - The rev2 code comments claiming "the opaque refusal itself
     surfaces through the install-status error in this same
     invocation" were false: no install-status error existed. This
     round makes that claim true.
2. **Draft refusal moved to resolve (mandated).** The finding-2 guard
   in `closure.ContentHashFor` fires inside `ResolveDraft`
   (`memberForNode`), so `TestDraftLaneStillBlocksNULBearingMember`'s
   build step (resolve with NUL present) now fails at lines 189/198 —
   before any install runs. The old expectation
   (`critical audit.opaque.nul-byte` from the audit gate) encodes
   pre-fix behavior: no lock can bind a NUL tree anymore, so the
   audit gate is unreachable for such trees. The test was updated to
   the mandated earlier refusal; the layer change is the fix, not a
   weakening (see §3).

## 2. Fixes

1. **Status-plan installed-tree guard.** New
   `refuseInstalledNULForStatus` (`internal/install/targets.go`),
   called from `projectAttempt` and `globalAttempt` after build
   planning, gated on `opts.Operation == OperationStatus` (the only
   production callers are the two status sites; plain `install
   --dry-run` keeps its shape). For each planned node it reads the
   recorded marker and runs `RefuseNULV1` at the recorded version;
   absent/unreadable/invalid markers are skipped (the drift
   classifiers report those states), v2 skips the scan. On refusal
   the plan fails with the opaque finding while `BuildsComplete`
   stays true, so CLI status prints the refusal to stderr plus the
   drift rows, and `--check` fails. Global scope covered identically.
2. **Draft test rewritten** (`nul_opaque_v2_test.go`):
   `explicit-resolve-refuses` and
   `refresh-refuses-and-keeps-prior-lock` (asserts lock bytes
   unchanged). Both pin `source_member_invalid` + the opaque id,
   proving the refusal fires at frozen-context load before any v1
   digest exists.
3. **CHANGELOG**: one clause added (status plans verify installed
   trees). No LOGBOOK or `scripts/remote-gate.sh` edits (verified in
   `git status`).

## 3. Mutants and stated bounds

- Delete the status-plan guard call → killed by the install
  status-plan legs (would return `ok`) and the cmd stderr assertion.
- Narrow the guard to top-level entries → killed by the nested
  `assets/a.bin` fixtures at install and cmd levels.
- Delete the `ContentHashFor` guard → `ResolveDraft` succeeds →
  killed by both draft subtests (`err == nil`).
- Hash-before-refuse at resolve is NOT distinguished by the
  resolve-level tests (both orders refuse); hash order is observed at
  every guarded entry with `hashing.CountV1Hashes` (rev4 F3
  correction: the rev3 text claimed the empty digest
  `TestContentHashForRefusesNULBeforeHashing` asserts proves order,
  but revision 3's M-order mutant survives that shape — only the
  zero-computation observation kills it).
- Grep audit: no other maintained test writes NUL bytes into an
  installed skill tree (remaining `\x00` uses are in-memory joins,
  hashes, or redaction fixtures), so the new plan refusal has no
  other in-repo caller to surprise.

## 4. Local evidence (all through `~/.local/bin/mini-build-lock`, `GOFLAGS=-work`)

syspolicyd running, successive crashes 26 → 26 across the session.

- `gofmt -l` on `internal/install/`, `cmd/curator/`,
  `internal/closure/`: clean, exit 0.
- `go vet ./internal/install/ ./cmd/curator/`: exit 0 (compiles
  production + test files, runs nothing).
- `go build ./...`: exit 0.
- `go test ./internal/audit -run 'NUL|Opaque|HashVersion|Verdict|CacheHit'
  -count=1`: ok 2.5s, exit 0.
- `go test ./internal/contextaudit -run 'NUL|Opaque|Version'
  -count=1`: ok, exit 0.
- `go test -count=1 ./internal/opaquescan ./internal/hashing`: both
  ok, exit 0.
- `golangci-lint run ./internal/install/...`: 0 issues, exit 0.

## 5. Not run locally (hosted gate is the arbiter)

Per R194 and the task brief, no `cmd/curator` or `internal/install`
test was executed here; `scripts/remote-gate.sh` was not run or
edited. The orchestrator runs on the hosted gate:

- `env GOFLAGS=-work go test ./internal/install -run
  'NUL|Opaque|V1Reinstall|DraftLane' -count=1 -timeout=6m`
- `env GOFLAGS=-work go test ./cmd/curator -run
  'NulOpaque|OpaqueNUL|StatusCheck' -count=1 -timeout=6m`
- the full hosted gate (`scripts/remote-gate.sh`) as the final arbiter
