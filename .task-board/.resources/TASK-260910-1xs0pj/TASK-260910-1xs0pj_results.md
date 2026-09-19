# TASK-260910-1xs0pj — migrate install markers with full currentness (handoff evidence)

Candidate: uncommitted working tree of `.temp/STORY-260910-1s75e1/worktree`
on top of checkpoint `e857e50d75a8dba8dc5ff0345243b775e194339c`.
Shell: bash; every gate below ran as a standalone process with a real exit
code. No commits on the Story branch.

## Changed paths

```
 M internal/install/draftruntime.go
 M internal/install/draftruntime_test.go
 M internal/install/install.go
?? internal/install/draftmarker_git_test.go
?? internal/marker/marker_v5_git_test.go
```

## What it implements (skillfile-sources.md protocol §4, "schema-2 installations")

This leaf removes the accepted interim boundary the wave-3 siblings left
behind ("Git draft members keep their accepted legacy shape until the
integration leaf migrates them"): every schema-2 installation now records
marker schema 5, regardless of skill manifest version.

- `internal/install/install.go` (`buildNodeMarker`): every locked draft
  member — local-snapshot, network-git and configured-git — stages
  through `buildDraftMarker` (marker v5) with the frozen package and the
  binding `lock_sha256`. The local-only condition is gone; frozen v1
  nodes still record the legacy marker byte-identically. The registry
  attestation the effective plan selected is passed into the builder.
- `internal/install/draftruntime.go` (`buildDraftMarker`): carries the
  preserved Git evidence summaries — `attestation` exactly as the
  effective plan selected it, `substituted` exactly the operator
  identifier the node carries — and fails closed with
  `source_member_invalid` when either would land on a `local-snapshot`
  member, which admits neither (no network identity, no selector to
  substitute). Returns an error now; the call site propagates it, so the
  commit aborts during private staging and no live state is touched.
- `internal/install/draftruntime.go` (`draftRuntimeKeys`,
  `draftBuildPackages`): both selections cover every locked member. The
  marker, the source-v1 runtime key and the receipt-3 binding migrate
  together, never one without the other — GC already marks every v5
  runtime leaf from the recorded package, and marker-5 build entries
  already bind receipt version 3, so migrating the marker alone would
  strand live trees and lie about receipts.
- `internal/marker`: no production change. Read/write/validate (closed
  raw-shape per arm, attestation/substituted forbidden for
  local-snapshot) and `Current` (package, lock, attestation,
  substitution compared against the effective plan in addition to every
  retained core comparison) already implement the full 25-field v4
  migration table from 17ps6u/dufdai; this leaf adds the Git-arm
  currentness/mismatch/evidence rows that table requires.
- `internal/install/draftruntime_test.go`: the interim-boundary test
  `TestDraftRuntimeKeysSkipGitMembers` is replaced by
  `TestDraftRuntimeKeysCoverAllArms` — the expectation change IS this
  task (the sibling documented the skip as interim "until marker and key
  migrate together"). All three arms are pinned to
  `SourceV1Key(package digest)`.

Accepted gates reused unchanged (not re-opened): closure
`BuildExpanded` refuses acquired `Substituted != ""`; install refuses
any dev-substitution manifest with draft selectors
(`source_selection_invalid`); strict audit refuses substituted installs
at step 4 before any planning; frozen consumption never sets
`Substituted`; `resolveRegistries` exact-matches identity/commit/context
and fails revoked/strict-unknown; `checkDraftSourceAudit` binds every
member package before cache/compiler work. All install gates run before
staging unconditionally — there is no up-to-date early exit ahead of
them — so a recorded summary can only report idempotence, never
authorize.

## Tests (production entries named)

Production entry `marker.Read`/`marker.Current`/`marker.Write`
(`internal/marker/marker_v5_git_test.go`):
- `TestMarkerV5GitRoundTrip` (+): network-git and configured-git v5 with
  attestation + substituted round-trip; raw bytes carry no replaced
  legacy field and every retained migration-table field; identical
  restaging is current (attested-network-current).
- `TestMarkerV5PlanMismatch` (−, 8 rows + control): registry, status,
  key_id, attestation-absent, substituted, substituted-absent, package,
  lock_sha256 mismatches are all non-current (the six
  marker-plan-mismatch semantic cases plus absence rows).
- `TestMarkerV5LocalRejectsEvidenceSummaryBytes` (−): foreign local-arm
  bytes carrying attestation/substituted are refused by Read and never
  current.
- `TestMarkerV5LegacyInterchange` (−, both directions): legacy-recorded
  vs v5 expectation and v5-recorded vs legacy expectation are both
  non-current; migration and rollback both reinstall.

Production entry `install.Project` with `DraftSourcesV1=true`
(`internal/install/draftmarker_git_test.go`):
- `TestDraftGitInstallWritesMarkerV5` (+): real network-git install
  records schema 5 binding the exact locked package + lock; raw bytes
  carry no legacy source field; reinstall reports up-to-date.
- `TestDraftGitRuntimeMaterializesUnderSourceV1Key` (+): Git script
  runtime materializes from the frozen tree under the namespaced package
  key as copied bytes (no links), context excludes the runtime root,
  shim reaches the protected store, marker is v5 network-git; reinstall
  is up-to-date.
- `TestDraftGitInstallCarriesRegistryAttestation` (+/−): recorded
  attestation equals the selected evidence exactly (registry/status/key
  presence, incl. key absence); changed evidence reinstalls and records
  the new summary.
- `TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker` (−):
  revoked evidence refuses the reinstall; the prior attested marker is
  preserved byte-identically (summary never authorizes).
- `TestDraftLocalRejectsConfusedAttestation` (−): evidence selected for
  a local member fails closed with `source_member_invalid` and stages
  nothing live.
- `TestDraftGitBuildsPublishReceipt3` (+): Git member with local go-v1
  and external go-repository-v1 commands publishes both under the
  receipt-3 namespaces only, marker records bind receipt 3 with every
  retained external field, reinstall takes exact cache hits without
  compiling.
- `TestDraftSelectorSubstitutionForbidden` (−, 2 rows): from-selector
  project + operator substitution manifest refuses with
  `source_selection_invalid` and publishes nothing; strict audit refuses
  even earlier. (selector-substitution-forbidden,
  legacy-substitution-strict shapes.)
- `TestDraftMarkerArmGuards` (builder seam): Git carries attestation and
  substituted; local + either is `source_member_invalid`. This seam is
  the narrowest honest one: the rows are unreachable through
  install.Project because the closure/install/strict gates refuse
  substitutions first and local nodes resolve no registry identity.

## Commands and exit codes (all run by me in this session)

| command | exit |
|---|---|
| `go build ./internal/install/ ./internal/marker/` | 0 |
| `go vet ./internal/marker/ ./internal/install/` | 0 |
| `gofmt -l internal/marker internal/install cmd/curator` | 0, no output |
| `git diff --check` | 0 |
| `golangci-lint run ./internal/marker/... ./internal/install/...` | 0 (0 issues) |
| `go test -p 1 ./internal/marker/ -count=1` (whole package) | 0 (2.4s) |
| `go test -p 1 ./internal/scopes/ -count=1` (whole package) | 0 (1.7s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftAudit'` | 0 (257s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftGit\|TestDraftMarker\|TestDraftRuntimeKeys\|TestDraftInstall\|TestDraftSelector\|TestDraftSources\|TestDraftMarkerPackage'` | 0 (54s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftBuild\|TestDraftLocal'` | 0 (82s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftPlan\|TestDraftSSH\|TestDraftTransport\|TestDraftManager\|TestLegacyInstallUntouchedWhenDraftOff\|TestLegacyBuildsKeepReceipt1'` | 0 (10s) |
| `go test -p 1 ./cmd/curator/ -run TestMarkerRefusalSeparatesUnsupportedFromInvalid` | 0 (0.5s) |

The full `-run TestDraft` single invocation exceeds the host's shell
bound (audit subset alone takes 257s on this host); the four subset rows
above cover every `TestDraft*` test plus the two legacy controls, all
green. Not run: the full install/cmd packages and the whole repository
suite (host stalls; the remote gate runs the landing suite once at
handoff). Legacy byte-identity rests on the untouched frozen lane
(`buildMarker`, nil-lock paths) plus `TestLegacyInstallUntouchedWhenDraftOff`
and `TestLegacyBuildsKeepReceipt1WithTheSwitchOff`.

## Narrowing mutants (all killed, files restored, rebuild verified)

| mutant | test that killed it | result |
|---|---|---|
| M1 `buildNodeMarker` routes Git back to legacy `buildMarker` | `TestDraftGitInstallWritesMarkerV5` | FAIL (killed) |
| M2 `buildDraftMarker` drops the attestation parameter | `TestDraftGitInstallCarriesRegistryAttestation` | FAIL (killed) |
| M3 `draftRuntimeKeys` skips Git members again | `TestDraftRuntimeKeysCoverAllArms` + `TestDraftGitRuntimeMaterializesUnderSourceV1Key` | FAIL (killed) |
| M4 `marker.Current` ignores attestation | `TestMarkerV5PlanMismatch` | FAIL (killed) |
| M5-narrow local guard keeps attestation check, drops substitution check | `TestDraftMarkerArmGuards` ("local marker admitted a development substitution") | FAIL (killed) |
| M6 `draftBuildPackages` skips Git members again | `TestDraftGitBuildsPublishReceipt3` | FAIL (killed) |

No survivors. (An M5 delete-variant died on build failure via the unused
import instead of behavior, so it was redone as the narrowing mutant
above.)

## Findings, decisions, bounds

- No `internal/marker` production change was needed: validation and
  `Current` already implement the complete migration table; the gap was
  entirely that Git draft members never reached the v5 builder. The
  marker-level deliverable is the Git currentness/mismatch/evidence
  coverage.
- `legacy-substitution-current` on the draft lane is unreachable by
  accepted design (resolve passes nil substitutions; install refuses any
  substitution manifest with draft selectors), so `node.Substituted` is
  always empty for draft installs today. The marker still preserves the
  field for Git arms (validation allows, currentness compares, the
  builder carries it) and the frozen v1 substituted lane is untouched;
  the builder-seam test pins the plumbing for where it applies.
- v5 local builds skip `VerifyShim` in `currentBuilds`, exactly like v2
  (only v3/v4 require it); no production caller drives `Current` with
  build evidence (install re-stages build-bearing nodes through cache
  hits), and the evidence machinery is covered by the schema-agnostic v2
  tests. Left as the deliberate v2 precedent, not changed speculatively.
- `detectMovedTagsIn` reads only `RefKind == "tag"` markers, so v5
  markers skip the advisory warning; draft tag moves are detected at
  explicit refresh, and the warning never authorizes. Unchanged.
- Out of scope for the follow-up (TASK-260910-3eu4cy owns Status,
  repair/refresh, runtimestore): `curator status` drift comparison and
  `registry.AttestRoot` still read legacy identity fields, so they do
  not yet compare/re-resolve v5 Git packages; no registry, UI, CLI or
  transaction file was touched here.
- Windows: no new skips and no POSIX-only fixtures added (no shim
  execution, symlinks, or executable-bit assertions; script fixtures
  declare both `unix_path` and `win_path` via the shared helper).
  Windows/Ubuntu proof comes from the hosted gate at handoff.
- LOGBOOK.md edits are forbidden by campaign rules; findings are
  recorded here per sibling precedent.

---

# Revision 2 — Git-arm identity restored (rework-1)

Cause of the rev1 gate failure: `buildDraftMarker` staged a pure package
binding for Git members, so the installed v5 marker read back
`RefKind/Ref/Commit = "  "` and the five accepted 3vxe3y CLI tests failed
on the marker identity. The status/currentness readers (`marker.Current`
triple comparison, `detectMovedTagsIn`, `scopeStatusDrift`,
`registry.AttestRoot`, `ui.skillsUnder`) all compare the declared
identity fields, so the gap was never a bound — it was a regression of
accepted behaviour on this leaf's scope line.

## What changed since rev1

```
 M internal/marker/marker.go            (v5 Git identity validation)
 M internal/install/draftruntime.go     (builder preserves Git identity)
 M internal/install/install.go          (comment only since rev1)
 M internal/install/draftruntime_test.go (unchanged since rev1)
 M internal/install/draftmarker_git_test.go (identity assertions + regression row)
 M internal/marker/marker_v5_git_test.go   (dual-write contract)
```

- `internal/marker/marker.go` (`validV5Identity`): a Git package may
  preserve the accepted declared identity — `source` (optional, portable
  path), `git` (optional non-empty, via the existing helper), and the
  `ref_kind/ref/commit` triple all-or-nothing under exactly the legacy
  grammar (new shared `validLegacyTriple`, also used by the legacy lane
  with identical semantics). A local snapshot still forbids all five
  plus attestation/substituted. Pure package bindings without any
  preserved identity stay readable (trunk `TestMarkerV5GitArmShape` and
  the closed-shape suite pass unchanged).
- `internal/install/draftruntime.go` (`buildDraftMarker`): for Git arms
  the builder copies `Source`/`Git` and the complete resolved triple
  from the frozen node — the same values the legacy marker carried —
  never fabricated or partial (an incomplete ref stages a pure package
  binding). Local arms copy nothing. Every rev1 local-arm decision is
  untouched.
- `marker.Current` needed no change: it already compares the triple
  (declared-ref moves refuse currency), then package + lock + attestation
  + substituted. All other identity consumers needed no change either;
  they read the v5 shape correctly once the fields are present (grep
  below).

## Consumer grep (rework item 3)

```
internal/install/install.go:1229  detectMovedTagsIn: recorded.RefKind/Ref/Commit (advisory warning restored for v5 Git tags)
internal/install/targets.go:44    stageNode via marker.Current (triple + package + lock + attestation + substituted)
internal/marker/marker.go:940     Current triple comparison (declared-ref moves non-current)
cmd/curator/main.go:1035          scopeStatusDrift: recorded vs declared ref (restored; 3eu4cy still owns Status)
internal/registry/attest.go:41-50 AttestRoot: recorded.Commit + recorded.Git (network-git re-resolves; configured-git
                                  stays unattestable via empty Git, exactly as the legacy lane)
internal/ui/state.go:131-137      skillsUnder display of Ref/Commit (read-only; meaningful again for v5 Git)
internal/scopes/gc.go:219         installed.Package.Digest() (package-based; unaffected)
```

No registry, UI, CLI or transaction production file was touched. The rev1
"status drift / AttestRoot still read legacy identity fields" bound is
closed by this revision: those readers now see the preserved identity in
v5 Git markers without code changes.

## Tests added/changed in rev2

- `TestDraftGitMarkerIdentityRoundTrip` (new, production entry
  `install.Project` + `marker.Current`): real Git install records schema
  5 with `tag v1 <locked commit>` + source/endpoint; identical restaging
  is current; a moved declared ref is non-current.
- `TestDraftGitInstallWritesMarkerV5`: raw bytes now must carry
  source/git/ref_kind/ref/commit + package + lock_sha256, and the triple
  must equal the locked selection.
- `TestDraftGitRuntimeMaterializesUnderSourceV1Key`: marker triple must
  equal declared tag v1 over the package commit.
- `TestDraftMarkerArmGuards` (builder seam): Git rows now carry a
  resolved identity and pin it; a new row pins that an unresolved Git
  node stages a pure package binding with no partial identity.
- `TestMarkerV5GitRoundTrip`: fixtures carry full per-arm identity;
  round-trip, raw presence and currentness pinned for both Git arms.
- `TestMarkerV5GitIdentityValidation` (new): 7 write-level + 12
  read-level rows — partial triple, bad kind, empty ref, malformed
  commit, bad/empty/null source, null kind/commit all refused by Write
  and by Read from foreign bytes, never current; pure-package control
  stays readable.
- `TestMarkerV5PlanMismatch`: +3 declared-ref rows (ref, ref-kind,
  commit with package unchanged → non-current).
- `TestMarkerV5LegacyInterchange`: strengthened — the legacy expectation
  now names the exact preserved triple and is still non-current both
  directions (schema mismatch dominates).
- `TestMarkerV5LocalRejectsEvidenceSummaryBytes`: +5 foreign-bytes rows
  (source/ref_kind/ref/commit/git refused for local-snapshot).

## Commands and exit codes (all run by me in this session, shell bash)

| `go test -p 1 ./cmd/curator/ -run 'TestProjectResolveGitAliasSelectionThroughCLI\|TestProjectResolveLegacyConfiguredGitThroughCLI\|TestProjectResolveLegacyNetworkGitThroughCLI\|TestProjectResolveGitTagDiffersFromHEADThroughCLI\|TestProjectResolveGitRealInstallThroughCLI\|TestMarkerRefusalSeparatesUnsupportedFromInvalid' -count=1` | 0 (58.9s) |

The five rework-named tests pass unchanged (no expectation edits), plus
the marker-refusal separator. Two earlier attempts in this session
failed before running a single test with `acquire package host GOROOT
test lock: context deadline exceeded` — a cross-process macOS test
fixture lock held by another agent's full-suite `cmd/curator` package
run on this shared host (PID 62261, alive 21:29–23:00); the third
attempt acquired it after that run's package timeout and went green.
The lock is taken in `cmd/curator` TestMain, so no narrower mask could
bypass it.

| `go test -p 1 ./internal/marker/ -count=1` (whole package) | 0 (1.4s) |
| `go test -p 1 ./internal/marker/ -run TestMarkerV5 -count=1` | 0 (0.8s) |
| `go test -p 1 ./internal/install/ -run TestDraftMarkerArmGuards\|TestDraftRuntimeKeysCoverAllArms` | 0 (0.9s) |
| `go test -p 1 ./internal/install/ -run TestDraftGitInstallWritesMarkerV5\|TestDraftGitMarkerIdentityRoundTrip` | 0 (10.4s) |
| `go test -p 1 ./internal/install/ -run TestDraftGitRuntimeMaterializes…\|TestDraftGitInstallCarriesRegistryAttestation\|TestDraftInstallRevoked…\|TestDraftLocalRejectsConfusedAttestation` | 0 (22.5s) |
| `go test -p 1 ./internal/install/ -run TestDraftGitBuildsPublishReceipt3\|TestDraftSelectorSubstitutionForbidden` | 0 (19.4s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftAudit[A-M]'` | 0 (65.6s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftAudit[N-Z]'` | 0 (211.2s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftBuild\|TestDraftLocal'` | 0 (118.4s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftPlan\|TestDraftSSH\|TestDraftTransport\|TestDraftManager\|TestDraftSources\|TestDraftMarker\|TestDraftInstall'` | 0 (19.9s) |
| `go test -p 1 ./internal/install/ -run 'TestLegacyInstallUntouchedWhenDraftOff\|TestLegacyBuildsKeepReceipt1'` | 0 (5.1s) |
| `go test -p 1 ./internal/scopes/ -count=1` | 0 (1.3s) |
| `go build ./internal/marker/ ./internal/install/ ./cmd/curator/` | 0 |
| `go vet ./cmd/curator ./internal/marker ./internal/install` | 0 |
| `gofmt -l internal/marker internal/install cmd/curator` | 0, no output |
| `git diff --check` | 0 |
| `golangci-lint run ./internal/marker/... ./internal/install/...` | 0 (0 issues) |

Together the install rows above cover all 57 `TestDraft*` tests plus
both legacy frozen-lane controls. Not run: the full install/cmd packages
and the whole repository suite (host stalls; the remote gate runs the
landing suite once at handoff).

## Narrowing mutants (rev2; all killed, files restored, rebuild verified)

| mutant | test that killed it | result |
|---|---|---|
| M7 `validV5Identity` narrows the triple check to commit-only (kind/ref unchecked) | `TestMarkerV5GitIdentityValidation` write+read/bad-ref-kind, read/empty-ref, read/null-ref-kind | FAIL (killed, exit 1) |
| M8 `buildDraftMarker` drops the triple copy, keeps Source/Git | `TestDraftGitInstallWritesMarkerV5` (`marker identity = \"review\" \"https://…\"   , want … tag v1 …`), `TestDraftGitMarkerIdentityRoundTrip`, `TestDraftMarkerArmGuards` | FAIL (killed, exit 1) |
| M1r `buildNodeMarker` routes Git back to legacy `buildMarker` | `TestDraftGitInstallWritesMarkerV5` (schema 2, nil package) | FAIL (killed, exit 1) |

M8 reproduces the exact rev1 gate signature (`marker identity = \"  \"`),
so the regression row guards the accepted 3vxe3y contract directly.
Rev1 mutants M1–M6 killed in the prior session; M1 is re-proven here as
M1r against the moved code, M4's target (`Current` attestation
comparison) is untouched by rev2. No survivors.

## Findings, decisions, bounds (rev2 deltas)

- The v5 Git marker is a dual binding: the frozen package (+ lock) is
  authoritative for selection identity and cache/GC keys, while the
  preserved declared triple keeps every status/currentness/attestation
  reader comparing the same selection the legacy marker carried. The
  package can never contradict the triple on the real path (both derive
  from the locked commit), and a mismatch from foreign bytes fails
  closed at Read or refuses currency at Current.
- Trunk `TestMarkerV5GitArmShape` and the closed-shape suite (pure Git
  v5 without legacy fields) pass unchanged by design: preserved
  identity is optional in validation, mandatory in the builder whenever
  the node resolves a complete ref (which frozen consumption guarantees
  for every locked Git member: draft alias roots via bindings,
  legacy roots via declaration, transitives via requirer recovery).
- `detectMovedTagsIn` re-enters for v5 Git tags (advisory warning only,
  read-only, never authorizes) — the rev1 "skips the warning" note is
  superseded.
- Windows: no new skips, no POSIX-only fixtures (same as rev1).

---

# Revision 3 — closed v5 shape restored (rework-2 + rev2 verdict)

ORCHESTRATOR RULING (binding, supersedes rework-1 point 2 "make the five
tests pass unchanged"): the accepted contract (curator-spec
protocol/skillfile-sources.md §"schema-2 installations",
schemas/draft-sources-v1/install-marker-v5.schema.json with
additionalProperties:false, the conformance README migration audit)
REPLACES the five legacy top-level fields source/git/ref_kind/ref/commit
with `package` + `lock_sha256` on EVERY v5 arm. The five cmd/curator
TestProjectResolve*ThroughCLI tests (3vxe3y) asserted the interim legacy
shape that 17ps6u explicitly left "until the integration leaf migrates
them" — this leaf is that migration. This revision re-points those tests
at the v5 identity (same semantic: "tag vN at the locked commit", v5
shape: package kind, `package.commit.hex`, `lock_sha256`, declared ref
from the bound manifest) and returns the marker to the closed v5 shape.
No spec change. Identity is still recorded and compared; only the
on-disk shape changes, exactly as the spec foresaw.

## What changed since rev2

```
 M internal/marker/marker.go               (closed v5 shape restored)
 M internal/install/draftruntime.go        (builder stages no legacy identity)
 M internal/install/install.go             (comment only since rev2; rev1 routing kept)
 M cmd/curator/draft_sources_test.go       (five 3vxe3y tests re-pointed at v5)
 M internal/install/draftmarker_git_test.go (closed shape + lock mechanism)
 M internal/marker/marker_v5_git_test.go   (closed shape + Git read negatives)
?? internal/marker/marker_v5_schema_test.go (schema drift/validation)
?? internal/marker/testdata/draft-sources-v1/*.schema.json (vendored spec)
```

- `internal/marker/marker.go` (`validV5Identity`, line 399): the five are
  refused on EVERY v5 arm again (trunk's loop); the five are dropped from
  the v5 `allowed` list; `validLegacyTriple` (line 425) keeps serving the
  legacy lane only under an identical predicate. Production v5 validation
  is back to the trunk e857e50 behavior (comment-only delta).
- `internal/install/draftruntime.go` (`buildDraftMarker`, line 118): the
  Git-arm Source/Git/triple copy is deleted; the builder stages the
  package + lock binding only. Kept from rev1/rev2: attestation +
  substituted carried for Git arms and refused for local arms (builder
  fail-closed + reader), `draftRuntimeKeys`/`draftBuildPackages` on all
  arms, receipt-3 binding.
- `cmd/curator/draft_sources_test.go`: new `assertV5GitMarker` helper
  (line 492) binds an installed v5 Git marker to its lock generation
  (schema 5; locked package kind, repository/source/directory, commit
  object format + hex; `lock_sha256 == lock.LockSHA256`; the five absent
  from the decoded marker AND the raw bytes; declared tag from the
  manifest the lock binds via `CheckStale`). Re-pointed assertions:
  - `TestProjectResolveGitAliasSelectionThroughCLI` (line 623):
    network-git, tag v2 at the locked commit (dry-run still names
    "review tag v2"; locked commit == tagged commit asserted).
  - `TestProjectResolveLegacyConfiguredGitThroughCLI` (line 704, default
    + custom-source): configured-git, `package.source` == the configured
    path ("review" / "custom/review"), tag v1 at the locked commit.
  - `TestProjectResolveLegacyNetworkGitThroughCLI` (line 779, default +
    custom-source): network-git, tag v1 at the locked commit.
  - `TestProjectResolveGitTagDiffersFromHEADThroughCLI` (line 942):
    network-git, tag v1 at the locked (non-HEAD) commit.
  - `TestProjectResolveGitRealInstallThroughCLI` (line 1334):
    network-git, tag v1 at the locked commit (lock now read from disk
    for the binding check); pinned reinstall without the repository
    still passes.
- `internal/install/draftmarker_git_test.go`:
  - `TestDraftGitInstallWritesMarkerV5`: closed-shape assertions (the
    five empty + absent from raw bytes; package + lock_sha256 present),
    package binds the locked member, `lock_sha256 == lock.LockSHA256`,
    declared tag v1 via the lock→manifest binding (`CheckStale`).
  - `TestDraftGitMarkerLockBindingRoundTrip` (line 197; renamed from
    ...IdentityRoundTrip, verdict item 4): declared-ref currency through
    the lock — install at v1, identical restaging current, re-resolve to
    tag v2 through the production closure (new upstream commit + tag,
    manifest + lock + bindings rewritten like an explicit refresh),
    `Current` false against the moved generation, reinstall re-stages
    and binds the moved lock, second reinstall up-to-date.
  - `TestDraftGitRuntimeMaterializesUnderSourceV1Key`: closed shape +
    lock binding instead of the triple.
  - `TestDraftMarkerArmGuards` (builder seam): Git carries attestation /
    substituted on the package binding with no legacy identity; local +
    either is `source_member_invalid`.
- `internal/marker/marker_v5_git_test.go`:
  - Fixtures carry package + attestation + substituted only.
  - `TestMarkerV5GitRoundTrip`: raw bytes carry NONE of the five;
    retained fields present; identical restaging current.
  - `TestMarkerV5GitRefusesLegacyIdentityBytes` (line 126; replaces
    ...GitIdentityValidation): Write refuses each of the five on both
    Git arms; Read refuses foreign bytes carrying each of the five as
    string, null and empty (2 arms × 5 fields × 3 forms) and never
    Current. This is the required read-level negative for the shape.
  - `TestMarkerV5PlanMismatch`: declared-ref rows removed (a v5
    expectation never carries the triple); currency flows through
    package + `lock_sha256` (kept rows).
  - `TestMarkerV5LegacyInterchange`: kept (both directions non-current).
- `internal/marker/marker_v5_schema_test.go` + vendored schemas (rework
  item 3): `testdata/draft-sources-v1/install-marker-v5.schema.json`
  (sha256 ab7aeb9acab3200c8e96ed757a8294973842346ec12fb9748884c77cd299b3ab,
  byte-identical to curator-spec 802caee) and
  `source-types-v1.schema.json`
  (sha256 fbb389bbfe5e2d05293e669e196d150f4fbaef1601ae239f8e7054c200621c15,
  byte-identical). `TestMarkerV5TopLevelMatchesAcceptedSchema` (line 67)
  pins the closed 22-member top level, the 15-member required set, the
  absence of the five, and the local-arm attestation/substituted forbids
  against the vendored schema. `TestMarkerV5PackageArmsMatchAcceptedSchema`
  (line 179) pins the in-code `v5PackageArms` table against the vendored
  source-types arms. `TestMarkerV5WrittenMarkersValidateAgainstAcceptedSchema`
  (line 328) validates a written marker of every arm (network-git,
  configured-git, local-snapshot) against the vendored schema, with a
  control row proving the check refuses a spliced-back `source` (not
  vacuous). Value grammars behind $refs stay covered by `validMarker`
  unit rows; this check pins the closed shape, the class that regressed.

## 25-field v4 → v5 enumeration (rev3; source: schemas/v1/install-marker-v4.schema.json)

25/25 exact (rev2 verdict counted 20/25; the 5 violations are closed):
schema_version → 5; source/git/ref_kind/ref/commit → replaced by
`package` (refused on every v5 arm at Write and Read, absent from every
installed marker); skill_schema_version 1..8; the 10 retained required
and 5 retained conditional/optional fields unchanged;
attestation/substituted retained for Git, refused for local;
builds → receipt 3; package + lock_sha256 required.

## Identity consumers under the v5 shape (verdict item 5; grep redone)

```
internal/install/targets.go:44    stageNode via marker.Current (package + lock + attestation + substituted; v5-correct, no change)
internal/marker/marker.go:971     Current triple comparison (empty==empty on v5; declared-ref moves surface via lock_sha256; no change)
internal/install/install.go:1218  detectMovedTagsIn: recorded.RefKind=="tag" never matches v5 (advisory warning skips v5; read-only, never authorizes)
internal/scopes/gc.go:219         installed.Package.Digest() (package-based; v5-correct, no change)
cmd/curator/main.go:1035          scopeStatusDrift: legacy triple comparison (schema-2 manifests fail manifest.Load without opt-in → empty drift; whole draft status surface deferred to TASK-260910-3eu4cy)
internal/registry/attest.go:41-50 AttestRoot: v5 markers report "unattestable: marker lacks commit or hash" (read-only; install-time attestation gate unchanged; re-resolution deferred to 3eu4cy)
internal/ui/state.go:137-138      skillsUnder: Ref/Commit display empty for v5 (display-only; package display deferred to 3eu4cy)
```

Explicit bound (no code change, per the verdict's "or record it as an
explicit bound" alternative): draft tag moves are detected at explicit
refresh (re-resolve → new lock → marker non-current → reinstall), never
via the advisory moved-tag warning. The warning, status drift,
status --attest re-resolution and TUI display for v5 belong to
TASK-260910-3eu4cy (Status/repair/refresh), which owns that surface.

## Commands and exit codes (all run by me in this session, shell bash)

Runner note (host incident, see bounds): syspolicyd is crash-looping on
this host (27 consecutive crashes observed, throttleTimeout 1200) and
fresh test binaries are SIGKILLed in waves; the `go test` wrapper also
stalls. Evidence below ran via `go test -c` + direct binary execution
from the package directory (required for testdata); exit codes are the
binaries' real exits. Attempts dying with 137 were retried and never
counted as results.

| command | exit |
|---|---|
| `go build ./internal/marker/ ./internal/install/ ./cmd/curator/` | 0 |
| `go build ./...` (whole repo) | 0 |
| `go vet ./internal/marker/ ./internal/install/ ./cmd/curator/` | 0 |
| `gofmt -l internal/marker internal/install cmd/curator` | 0, no output |
| `git diff --check` | 0 |
| `golangci-lint run ./internal/marker/... ./internal/install/...` | 0 (0 issues) |
| marker full package `-test.count=1` (incl. all TestMarkerV5* + schema tests) | 0 PASS |
| install core: `TestDraftGitInstallWritesMarkerV5` (5.1s) + `TestDraftGitMarkerLockBindingRoundTrip` (8.2s) | 0 PASS |
| install evidence: `TestDraftGitRuntimeMaterializesUnderSourceV1Key` (8.4s), `TestDraftGitInstallCarriesRegistryAttestation` (7.2s), `TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker` (6.3s), `TestDraftLocalRejectsConfusedAttestation` (0.6s) | 0 PASS |
| install misc: `TestDraftGitBuildsPublishReceipt3` (11.2s), `TestDraftSelectorSubstitutionForbidden` (0.3s), `TestDraftMarkerArmGuards`, `TestDraftRuntimeKeysCoverAllArms`, `TestDraftMarkerPackageDigestMatchesLockDigest` | 0 PASS |
| install `TestDraftAudit[A-H]` (8) / `[I-Q]` (3) / `[R-Z]` (5) | 0 PASS (16/16) |
| install `TestDraftBuild\|TestDraftLocal` (15) | 0 PASS |
| install `TestDraftPlan\|TestDraftSSH\|TestDraftTransport\|TestDraftManager\|TestDraftSources\|TestDraftMarker\|TestDraftInstall\|TestLegacyInstallUntouchedWhenDraftOff\|TestLegacyBuildsKeepReceipt1` (21) | 0 PASS |
| scopes full package `-test.count=1` | 0 PASS |
| CLI `TestProjectResolveGitAliasSelectionThroughCLI` (5.8s) | 0 PASS |
| CLI `TestProjectResolveLegacyConfiguredGitThroughCLI` (14.8s, 2 modes) | 0 PASS |
| CLI `TestProjectResolveLegacyNetworkGitThroughCLI` (11.5s, 2 modes) | 0 PASS |
| CLI `TestProjectResolveGitTagDiffersFromHEADThroughCLI` (6.7s) | 0 PASS |
| CLI `TestProjectResolveGitRealInstallThroughCLI` alone (9.5s) | 0 PASS |
| CLI `TestMarkerRefusalSeparatesUnsupportedFromInvalid` (0.01s) | 0 PASS |
| final fresh-binary confirmation (restored tree): marker full PASS; install focused 4/4 PASS | 0 |

Mask audit: the install rows above cover 57/57 `TestDraft*` tests plus
both legacy frozen-lane controls. Not run: the full install/cmd packages
and the whole repository suite (host stalls; the remote gate runs the
landing suite once at handoff). One `TestProjectResolveGitRealInstallThroughCLI`
attempt in a combined run timed out after 4m40s stuck spawning git in
the gitignore gate (pre-marker path) during a host exec-stall wave; it
passed alone on retry (9.5s). Environmental flake, not a product hang:
the hang point precedes every code path this leaf touches.

## Narrowing mutants (rev3; all killed, files restored, restore verified)

| mutant | test that killed it | result |
|---|---|---|
| M1 `validV5Identity` + v5 `allowed` re-admit `source` on Git arms (rev2 shape) | `TestMarkerV5GitRefusesLegacyIdentityBytes` write/source + all 6 read/source rows | FAIL (killed, exit 1) |
| M2 reader drops the local `substituted` refusal, keeps `attestation` | `TestMarkerV5LocalRejectsEvidenceSummaryBytes/substituted`, `TestMarkerV5RefusesMalformedIdentity/substituted` (attestation rows still pass) | FAIL (killed, exit 1) |
| M3 `buildDraftMarker` drops the local substitution guard, keeps attestation | `TestDraftMarkerArmGuards` (`local marker admitted a development substitution`); `TestDraftLocalRejectsConfusedAttestation` still passes | FAIL (killed, exit 1) |
| M4 `Current` keeps the package comparison, drops `lock_sha256` | `TestMarkerV5PlanMismatch/lock_sha256`, `TestMarkerV5Currentness` | FAIL (killed, exit 1) |

No survivors. `TestDraftGitMarkerLockBindingRoundTrip` does NOT kill M4
(the moved generation differs in package as well as lock), which is the
intended layering the verdict foresaw: the install test pins the
re-resolve → new lock → non-current → restage mechanism end to end,
while the marker rows pin the `lock_sha256` comparison itself. M3's
killer is the builder seam by accepted design (the production entry
cannot reach local+substituted: closure/install/strict gates refuse
first). Restore verified after each mutant (`grep MUTANT` empty,
`git diff --stat` back to the 5-file + 4-untracked set, rebuild OK).

## Findings, decisions, bounds (rev3 deltas)

- Rev2's "dual binding" design note is superseded and withdrawn: v5 is
  package + lock only, on every arm. The rev1 shape (pure package
  binding) was the conformant one; rev3 is rev1's shape plus the
  lock→manifest declared-ref binding made explicit in tests.
- The `detectMovedTagsIn` bound returns to its rev1 form: v5 markers
  skip the advisory moved-tag warning (read-only in any case); draft tag
  moves surface at explicit refresh through the lock comparison, proven
  by `TestDraftGitMarkerLockBindingRoundTrip`.
- `Current`'s triple comparison is intentionally unchanged: it is
  empty==empty on the v5 lane (both sides schema-closed) and still
  separates legacy↔v5 interchange in both directions.
- The schema-validation test pins the closed SHAPE against the vendored
  spec (top-level sets, required sets, package arm shapes, local-arm
  forbids); value grammars behind $refs ($defs in v4/common) stay
  covered by `validMarker` unit rows. The hosted Interop gate pins
  conformance/v1 only, so this in-repo test is what guards the draft
  shape until the gate widens.
- Windows: no new skips and no POSIX-only fixtures added (no shim
  execution, symlinks, or executable-bit assertions; script fixtures
  declare both `unix_path` and `win_path` via the shared helper).
  Windows/Ubuntu proof comes from the hosted gate at handoff.
- LOGBOOK.md edits are forbidden by campaign rules; findings are
  recorded here per sibling precedent.
