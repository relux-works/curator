# Validation — TASK-261006-3ptm2x rework 2 (rev4): revision-3 findings F1/F2/F3

Revision-3 review (`TASK-261006-3ptm2x_review-verdict-rev3.md`,
RUN-261006-a39bf8) confirmed both rev1 P1 fixes and the audit ordering,
and requested changes with three P2 findings. This revision fixes all
three, adapts the reviewer's attack probes into maintained regression
tests, and names the killing mutants. No version is inferred from digest
bytes anywhere: every rule below keys off a recorded `hash_version`, a
lock declaration, the lane, the writer switch, an explicit caller
version, or the frozen shape.

## 1. Design note

### F1 — frozen v1 guard scans the full snapshot

`closure.ContentHashFor` scanned the context projection (`destination`)
after `whitelist.CopyContext` had omitted declared runtime/build roots
and non-whitelisted paths, so a NUL file in an excluded root escaped the
v1 guard while `ResolveDraft` still bound the tree to a frozen v1 lock
member. The fix scans the full frozen tree (`frozen`) with
`opaquescan.RefuseNULV1(frozen, VersionV1)` before filtering or hashing,
and keeps the projection for the hash only. The refusal keeps the
`source_member_invalid: frozen package context: audit.opaque.nul-byte`
shape, so existing message expectations are unchanged. `ResolveDraft`,
`RefreshDraft` (which fails before publishing, preserving the prior
lock/bindings), `LoadDraftFrozenNodes`, and the three
`internal/install/draftsources.go` callers all inherit the fix through
`ContentHashFor`. Inventory delta vs the rework-1 note: row 1b.5 now
reads "scans the FULL frozen snapshot before filtering or hashing".

No other v1 identity is computed on the resolve path before the guard:
the snapshot inventory digest, manifest digest, and closure-graph
identity are separate schemes outside `hashing.ContentSHA256`, which
the F3 observation confirms (refused resolves report zero).

### F2 — trust pins are a versioned carrier

`audit.Pin` wrote schema 1 without `hash_version` and `isPinned`
compared digest text only, so a legacy v1 pin authorized a v2 subject
at equal digest text — the finding-3 mechanism at the pin surface. The
fix mirrors the verdict carrier:

- v1 pins keep the byte-identical frozen schema-1 shape (no
  `hash_version` member) and authorize v1 reads only.
- v2 pins are schema 2 with `hash_version: 2`.
- New writer `audit.PinAtVersion(home, digest, version, reason, by)`;
  the legacy `audit.Pin` stays as the explicit v1 writer (all existing
  v1 pin tests pass unmodified). Unknown versions refuse; zero reads as
  v1, matching every other carrier reader.
- `isPinned(cfg, digest, version)` compares (version, digest); legacy
  absent-schema records read as v1; unknown schemas never authorize.
- `decideWithPins` passes the subject's resolved version;
  `internal/audit/sourceaudit.go` passes explicit v1 (frozen lane).
- CLI `audit --allow` records `hashing.WriteVersion()` explicitly — no
  digest inference. New pins are v2 under the current writer.

Note for the next reviewer: the rev3 probe
`TestReviewNewV2PinCarrierVersion` calls the versionless `Pin(...)` and
expects `hash_version: 2`. That API cannot exist without inferring the
version from digest bytes (forbidden by the task) or silently
re-versioning every legacy caller, so the maintained contract is the
explicit `PinAtVersion`: please re-probe through it
(`TestPinAtVersionWritesVersionedCarrier` is the maintained
equivalent). Same-digest-text v1+v2 pins share one digest-keyed
`trust.json`, mirroring the verdict file: last write wins, reads match
by version. Real collisions do not occur; synthetic equal-text fixtures
seed one version at a time.

CHANGELOG and the rework-1 validation note's pin claims are corrected:
legacy pins keep authorizing v1 reads only, and v2 identities need a
fresh pin after the cutover.

### F3 — refusal-before-hashing is observed, not inferred from shape

Revision 3 proved empty-digest/error-shape assertions cannot see a
discarded hash-before-refuse mutant (M-order passes all three old
tests). The fix is a test seam in `internal/hashing`:

- `CountV1Hashes(fn)` runs `fn` and reports how many v1 content-hash
  computations it performed (trees and in-memory file sets; v2 and
  failed dispatches never count). Dispatch wrappers delegate, so each
  computation counts exactly once. Production never calls it; it is not
  reentrant.
- Every guarded production entry gets a zero-assertion over a NUL tree
  plus a clean control requiring ≥1 computation, proving the seam is
  wired at that entry. A discard mutant (or the live hash moved above
  the guard) computes one v1 identity and fails the zero assertion.

The rework-1 and rev3 validation notes' "empty digest proves order"
claims are corrected in place (resources updated, marked "rev4 F3
correction").

## 2. Reviewer-probe mapping (all adapted into maintained tests)

| Rev3 probe | Maintained regression | Package |
|---|---|---|
| TestReviewFrozenV1RefusesExcludedNUL (runtime/build/unlisted) | TestContentHashForRefusesExcludedNUL | closure (local green) |
| TestReviewResolveDraftRefusesExcludedNUL | TestResolveDraftRefusesExcludedNUL | closure (local green) |
| (refresh + prior-lock preservation) | TestRefreshDraftExcludedNULPreservesPriorLock | closure (local green) |
| TestReviewV2RejectsLegacyPin | TestV2RejectsLegacyV1Pin | audit (local green) |
| TestReviewNewV2PinCarrierVersion | TestPinAtVersionWritesVersionedCarrier (via PinAtVersion; see API note above) | audit (local green) |
| (reverse + controls + schemaless) | TestV1RejectsV2Pin, TestPinVersionMatchAuthorizes, TestLegacySchemalessPinReadsAsV1 | audit (local green) |
| (CLI writer) | TestAuditAllowWritesWriterVersionPinCarrier | cmd/curator (hosted) |
| TestReviewSourceAuditNeverHashesNULV1 | TestCheckSourceAuditV1NULRefusalComputesNoV1Identity | audit (local green) |
| (gate + pipeline observation) | TestGateV1NULRefusalComputesNoV1Identity, TestAuditSubjectV1NULRefusalComputesNoV1Identity | audit (local green) |
| (currentness/store/profile observation) | TestCurrentV1NULRefusalComputesNoV1Identity, TestContentHashV1NULRefusalComputesNoV1Identity, TestGuardedReadersV1NULRefusalComputesNoV1Identity (3 subtests) | marker/contextstore/envprofile (local green) |
| (seam unit proof) | TestCountV1HashesObservesOnlyV1Computations | hashing (local green) |
| TestReviewLegacyMarkerCollisionAndDowngrade | already-maintained rev1 tests (rev3 confirmed passing) | marker (local green) |

F3 zero-assertions are also folded into the three closure F1 tests, so
the frozen path proves scope and order together.

## 3. Mutants (executed locally, all killed, all reverted)

- M-F1 (narrowing): move the full scan back to `destination`
  (projection-only scan). Killed: `TestContentHashForRefusesExcludedNUL`
  (3/3 shapes), `TestResolveDraftRefusesExcludedNUL` (3/3),
  `TestRefreshDraftExcludedNULPreservesPriorLock` — exit 1, then
  reverted.
- M-F2 (narrowing): `isPinned` ignores the recorded version (returns
  true for any pinned record). Killed: `TestV2RejectsLegacyV1Pin`,
  `TestV1RejectsV2Pin`, `TestLegacySchemalessPinReadsAsV1` — exit 1
  (`errs = []`, unauthorized admission), while
  `TestPinVersionMatchAuthorizes` still passes under the mutant as it
  must; then reverted.
- M-F3a (reviewer's exact M-order): discarded
  `ContentSHA256WithVersion(snapshot, nil, V1)` inside the
  `auditSubjectWithOpaquePaths` refusal branch. Killed: the two new
  auditSubject/CheckSourceAudit zero-tests — exit 1. The three old
  shape-tests still pass under it, reproducing the F3 evidence. Then
  reverted.
- M-F3b: same discard inside `gate()`'s opaque branch. Killed:
  `TestGateV1NULRefusalComputesNoV1Identity` ("computed 1 v1
  identities") — exit 1, then reverted. (M-F3a correctly does NOT trip
  the Gate test: the gate refuses before reaching the pipeline.)
- M-F3c: same discard inside `ContentHashFor`'s refusal branch. Killed:
  all three closure scope tests incl. every subtest — exit 1, then
  reverted.
- Post-revert: `grep MUTANT-F` over internal+cmd is empty; audit and
  closure NUL/pin suites re-run green.

Named but not executed as separate local mutants: per-site discard
mutants at marker.Current, contextstore.ContentHash, and the three
envprofile readers. Each shares the identical seam and pattern (counter
wraps the entry; the hashing line sits strictly after the guard), and
each zero-test carries a clean control proving the entry hashes through
the seam — a discard at any of them must increment. Stated as a bound,
not as executed evidence.

## 4. Local evidence (all through ~/.local/bin/mini-build-lock, GOFLAGS=-work)

syspolicyd running, successive crashes 26 → 26 across the session
(checked at start and end; no movement).

- `gofmt -l` on hashing/audit/closure/marker/contextstore/envprofile/
  opaquescan/contextaudit/cmd: clean, exit 0.
- `go vet` on the eight local packages: exit 0.
- `go vet ./internal/install ./cmd/curator` (compiles the
  hosted-pending tests): exit 0.
- `go build ./...`: exit 0.
- `go test ./internal/hashing -count=1`: ok, exit 0.
- `go test ./internal/audit -count=1` (full): ok, exit 0.
- `go test ./internal/closure -run 'NUL|Opaque|ContentHashFor|
  ResolveDraft|RefreshDraft|Refresh'`: ok (16 tests incl. 3 new),
  exit 0.
- `go test ./internal/marker -count=1` (full): ok, exit 0.
- `go test ./internal/opaquescan ./internal/contextaudit
  ./internal/contextstore -count=1`: all ok, exit 0.
- `go test ./internal/envprofile -run 'NUL|Opaque|GuardedReaders|
  StrictAuditMember|StorePins|SkillsOf|Colliding'`: ok (9 tests incl.
  the new F3 test), exit 0.
- Mutants M-F1/M-F2/M-F3a/M-F3b/M-F3c: each exit 1 on its killing
  test(s) as listed above, then reverted and re-run green.
- `golangci-lint run` on the seven library trees: 0 issues, exit 0;
  on cmd/curator + internal/install: 0 issues, exit 0.

## 5. Not run locally (hosted gate is the arbiter)

Per R193/R194 and the task brief, cmd/curator and internal/install
suites were written and compile-verified (vet exit 0, lint 0 issues)
but not executed here. The orchestrator runs on the hosted gate:

- env GOFLAGS=-work go test ./cmd/curator -run 'TestAuditAllowWritesWriterVersionPinCarrier' -count=1 -timeout=6m
- env GOFLAGS=-work go test ./internal/install -run 'NUL|Opaque|V1Reinstall|DraftLane' -count=1 -timeout=6m
- env GOFLAGS=-work go test ./cmd/curator -run 'NulOpaque|OpaqueNUL|StatusCheck|AuditAllow' -count=1 -timeout=6m
- the full hosted gate (scripts/remote-gate.sh) as the final arbiter

## 6. Stated bounds (unchanged unless noted)

- Scan-then-hash races (TOCTOU) between guard and hash remain out of
  scope; the managed install transaction still pins admitted inputs
  for its per-write recheck.
- Install/cmd production entries are covered by refusal tests plus
  delegation to the observed library entries; no v1-call counts are
  asserted at the install/cmd level (multi-hash planning surfaces make
  exact counts brittle there).
- The `audit --allow` flow takes an operator-supplied digest and the
  writer framing; it never guesses a version from bytes.
- CHANGELOG Unreleased extended (pin versioning + frozen full-snapshot
  scope). LOGBOOK.md and scripts/remote-gate.sh untouched (verified in
  git status: no output for either path).

## 7. Files changed (rework 2 on top of the rev3 candidate)

Production: internal/hashing/hashing.go (CountV1Hashes seam),
internal/closure/resolve.go (full-snapshot guard),
internal/audit/audit.go (PinAtVersion, versioned isPinned,
decideWithPins), internal/audit/sourceaudit.go (explicit v1 pin
reads), cmd/curator/main.go (`--allow` records WriteVersion),
CHANGELOG.md.
Tests: internal/hashing/v1observe_test.go (new),
internal/closure/nul_opaque_v1_scope_test.go (new),
internal/audit/opaque_v1_pins_test.go (new),
internal/audit/opaque_v1_nohash_test.go (new),
cmd/curator/nul_opaque_v1_pins_test.go (new, hosted);
extended: internal/marker/nul_opaque_v1_test.go,
internal/contextstore/nul_opaque_v1_test.go,
internal/envprofile/nul_opaque_v1_guard_test.go.
Board: this note (new); TASK-261006-3ptm2x_validation.md and
TASK-261006-3ptm2x_validation-rev3.md corrected in place (rev4
F1/F2/F3 markings).
