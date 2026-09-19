# Review verdict — TASK-260910-1xs0pj revision 3 (CR-TASK-260910-1xs0pj-3)

Reviewer run: RUN-260918-3eadd5 (claude-opus-5, reviewer archetype, read-only).
Reviewed: 2026-09-19 (UTC 00:00–00:15), macOS host e11-1, load 6–14.

## Verdict: ACCEPT

`accept_cr(TASK-260910-1xs0pj, revision=3, evidence=TASK-260910-1xs0pj_review-verdict-rev3.md)`

Revision 3 restores the closed v5 shape on every arm exactly as the accepted spec
(curator-spec `protocol/skillfile-sources.md` §4 "schema-2 installations",
`schemas/draft-sources-v1/install-marker-v5.schema.json`, `additionalProperties:false`)
and the orchestrator ruling (1xs0pj-rework-2.md) require; the rev2 defect (v5 Git markers
carrying `source/git/ref_kind/ref/commit`) is closed at the writer, the reader and the
production entry, and real written documents of every arm validate against the spec schema
with a real JSON Schema validator. All 25 v4 fields now migrate exactly as the table says.
Non-blocking findings and bounds are listed at the end for the orchestrator.

## Candidate identity (verified, not taken from the producer)

- Story worktree working tree (temporary index `git read-tree HEAD && git add -A internal/ cmd/
  && git write-tree`) = `63c32caedb01fa50620fd39f52d6c74ed6adbd53` = CR candidate tree. ✓
- Patch `TASK-260910-1xs0pj_change-request_rev3.patch` sha256
  `27fbc6084d3c87cb9ac1b1cd94fe8f4daf837bc13bdcde143d6a80639f6a5338` (matches the CR). ✓
- Disposable clones (`git clone --no-checkout` of the control root, `checkout --detach
  e857e50`, `git apply --index` rev3 patch, commit): `HEAD^{tree}` =
  `63c32caedb01fa50620fd39f52d6c74ed6adbd53` in both the test clone
  (`/tmp/1xs0pj-rev3-review`) and the mutant clone (`/tmp/1xs0pj-rev3-mut`). Every probe and
  mutant ran there; the Story worktree was never written. Both clones end at the candidate
  tree (`git write-tree` = `63c32cae…`). ✓
- Hosted gate: GitHub run 35402478828 conclusion `success`; `headSha`
  `5cc4590fabc7ee9e99d5a7356219792a87ddd435` → `git cat-file -p` shows
  `tree 63c32caedb01fa50620fd39f52d6c74ed6adbd53`, `parent e857e50d…` (= base). Jobs: Naming
  gate, Lint, Interop conformance gate, Test/Race on ubuntu + macos, Test on windows, three
  gate self-tests — all `success`; `Test (rose-air)` and `Candidate suite` `skipped`
  (self-hosted lane not run on gate snapshots — bound). ✓
- Vendored schemas `internal/marker/testdata/draft-sources-v1/{install-marker-v5,source-types-v1}.schema.json`
  are byte-identical (`cmp`) to curator-spec main 802caee (landed in a4fcaf0):
  sha256 `ab7aeb9a…` / `fbb389bb…`. ✓

## Independent reruns (disposable clone, bash driver with `set -o pipefail`, real exit codes)

| command | exit |
|---|---|
| `go vet ./internal/marker ./internal/install ./cmd/curator` | 0 |
| `gofmt -l internal/marker internal/install cmd/curator` | 0, no output |
| `go test -p 1 ./internal/marker -count=1 -timeout=300s` | 0 (`ok … 0.875s`) |
| `go test -p 1 ./internal/install -run 'TestDraft\|TestLegacyInstallUntouchedWhenDraftOff' -count=1 -timeout=500s` | 0 (`ok … 296.198s`) |
| `go test -p 1 ./cmd/curator -run 'TestProjectResolve' -count=1 -timeout=400s` | 0 (`ok … 61.899s`) |
| the five re-pointed 3vxe3y tests alone, `-v` (see addendum at the end) | see addendum |
| `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1 go test ./internal/marker -run TestReadAuthoritativeMarkerV4SchemaCases` | PASS (frozen v4 corpus through the unchanged legacy reader) |

Host note: no exec stalls this time (load 6–14, 77 GiB free); nothing was retried.

## Spec-schema validation of REAL written markers (rev2's defect class)

Throwaway tests in the clone drove the production entries and dumped the marker bytes;
`/tmp/1xs0pj-rev3-venv` (jsonschema 4.25.1 + referencing 0.37.0),
`Draft202012Validator(install-marker-v5.schema.json, registry over schemas/v1 +
schemas/draft-sources-v1)` per the conformance README recipe:

| document (entry point) | package.kind | attestation | builds | result |
|---|---|---|---|---|
| `install.Project` network-git (setupGitInstall, ResolveAttest keyed) | network-git | yes (key_id) | 0 | VALID |
| `install.Project` network-git plain | network-git | no | 0 | VALID |
| `install.Project` network-git script member (runtime root) | network-git | no | 0 | VALID |
| `install.Project` network-git build member (go-v1 + go-repository-v1, receipt 3) | network-git | no | 2 | VALID |
| `install.Project` local-snapshot (draftLocalCollectionPayload) | local-snapshot | no | 0 | VALID |
| CLI `curator project resolve` + `install` legacy configured entry `{name, tag}` | configured-git | no | 0 | VALID |

Controls: the spec corpus `schema-cases/install-marker-v5` — 38/38 documents agree with the
validator (7 `valid-*` accepted, 31 `invalid-*` refused), so the validator is not vacuous.
`validate.py` exit 0. The dumped network-git-attested document carries exactly
`package{kind,repository,commit{object_format,hex},directory}`, `lock_sha256`, `attestation`
and the retained fields; none of `source/git/ref_kind/ref/commit` is present.

## Spec corpus through the production reader (`marker.Read`) — measured

38 corpus documents written as `.csk-install.json` and read: **36/38 agree**.
- 31 negatives: 30 refused (incl. `invalid-unknown-top-level`, `invalid-local-attestation`,
  `invalid-local-substituted`, `invalid-fake-commit`, every attestation and external-record
  negative except one), all 7 positives except one accepted.
- Disagreement 1 — `invalid-external-missing-substituted.json` is ACCEPTED by `Read`
  (external `go-repository-v1` build record without the `substituted` boolean decodes as
  `false`; `validDriverBuild`, trunk 20sx61 code, does not close the build-record raw
  shape). Pre-existing on trunk, outside this leaf's delta (marker.go delta = comment +
  `validLegacyTriple` extraction) and outside its scope line (build records are dufdai/17ps6u
  territory); the writer always emits `substituted` for external records
  (`Build.MarshalJSON`), and `Current` compares records structurally. Recorded as
  finding F1 below — follow-up, not blocking.
- Disagreement 2 — `valid.json` (builds `{}` with top-level `build_source` present) is
  REFUSED by `Read`: trunk's all-or-nothing rule (`validBuildState`: no builds → no
  build_source) implements the prose "`build_source` … required exactly for active local
  go-v1 commands and absent otherwise"; the corpus positive is schema-valid but
  prose-inconsistent. Curator never writes such a document. Not a candidate defect (F2:
  spec-corpus quirk to raise with curator-spec).

## Acceptance clauses — positive and negative rows at the production entry

1. **Marker v5 on every schema-2 installation, package replaces the five** —
   `install.go:1039-1060` routes every locked member to `buildDraftMarker`
   (`draftruntime.go:118-192`), which sets none of the five; `validV5Identity`
   (`marker.go:399-420`) refuses each of the five present on ANY v5 arm and the v5 `allowed`
   list (`marker.go:351-357`) is exactly the 22 schema properties (15 required + 7 optional).
   Positive: `TestDraftGitInstallWritesMarkerV5`, `TestDraftGitRuntimeMaterializesUnderSourceV1Key`,
   `TestDraftGitBuildsPublishReceipt3`, the five re-pointed CLI tests, my six dumps. Negative:
   `TestMarkerV5GitRefusesLegacyIdentityBytes` (Write ×5 fields ×2 arms; Read ×5 fields ×3
   forms ×2 arms, never Current), `TestMarkerV5LocalRejectsEvidenceSummaryBytes` (7 rows),
   `TestMarkerV5WrittenMarkersValidateAgainstAcceptedSchema/control` (spliced `source` refused).
2. **Git-arm identity round-trips (3vxe3y semantic in v5 shape)** — `assertV5GitMarker`
   (`cmd/curator/draft_sources_test.go:492-548`): schema 5, locked package kind/repository/
   source/directory/commit{object_format,hex}, `lock_sha256 == lock.LockSHA256`, the five
   absent from the decoded marker AND the raw bytes, lock binds the declaring manifest
   (`CheckStale`), locked commit == the tagged commit. Applied in all five tests (alias v2,
   legacy configured ×2 modes, legacy network ×2 modes, tag≠HEAD, real install + offline
   pinned reinstall). Reran: pass.
3. **Declared-ref currency through the lock** — `TestDraftGitMarkerLockBindingRoundTrip`:
   identical restaging current; re-resolve to tag v2 through `closure.ResolveDraft` → new
   lock → `Current` false → reinstall re-stages and binds the moved lock → next reinstall
   up-to-date. My probe `TestReviewLockOnlyChangeRestages` (clone only): a lock-only change
   (second member from the same source/tag; review package byte-identical) → `Current`
   false and `install.Project` re-stages review and rebinds `lock_sha256` — PASS (10.6 s).
4. **Attestation preserved for Git, refused for local** — positive
   `TestDraftGitInstallCarriesRegistryAttestation` (exact registry/status/key, key absence
   significant, evidence change re-stages); negative at the production entry
   `TestDraftLocalRejectsConfusedAttestation` (`source_member_invalid … cannot carry a
   registry attestation`, nothing staged — verified `Lstat` absent); reader negative
   `TestMarkerV5LocalRejectsEvidenceSummaryBytes/attestation`. Never authorizes:
   `TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker` (revoked evidence refuses the
   reinstall, prior marker byte-identical).
5. **Legacy `substituted` preserved for Git, refused for local; new selectors cannot enable
   it** — builder carries `node.Substituted` on Git arms and refuses it on local
   (`draftruntime.go:132-139`); reader refuses it on local (`marker.go:415-417`); `Current`
   compares it (`marker.go:973`). `TestDraftSelectorSubstitutionForbidden` (install.Project:
   `source_selection_invalid: development substitutions are not admitted with draft
   selectors`, nothing published; strict audit refuses first). Bound B1: the Git-arm
   positive is reachable only at the builder seam (`TestDraftMarkerArmGuards`) because the
   accepted 3vxe3y gate (`install.go:384-387`, landed 6645b9b) refuses every substitution on
   the draft lane — the field can be read, compared and refused, never produced today.
6. **Package/lock binding, all mismatch/evidence cases** — `Current` (`marker.go:956-981`,
   trunk, unchanged): package DeepEqual, `lock_sha256`, locale, agents, activation,
   substituted, mcp, attestation, builds, content hash. `TestMarkerV5PlanMismatch` (registry,
   status, key_id, attestation-absent, substituted, substituted-absent, package, lock_sha256),
   `TestMarkerV5LegacyInterchange` (both directions non-current), `TestMarkerV5Currentness`.
7. **Summaries never authorize** — readers report only (`stageNode` → `marker.Current`;
   `detectMovedTagsIn` advisory; `AttestRoot` reports `unattestable` for v5); clause 4's
   revoked-evidence row; strict audit refuses before cache reads (clause 5).
8. **Legacy lanes byte-identical with the switch off** — legacy writer/validator delta is
   the `validLegacyTriple` extraction (identical predicate by De Morgan) and comments;
   `buildMarker` untouched. Empirical: a schema-2 marker (all optional fields, git,
   attestation, substituted) and a schema-4 marker written by base e857e50 and by the
   candidate hash identically (sha256 `80d491da…`, 951 B; `7ed23e6c…`, 544 B).
   `TestLegacyInstallUntouchedWhenDraftOff` pass; frozen v4 corpus through the reader pass;
   hosted legacy suites green on three OSes.
9. **Runtime key / receipt migrate with the marker** — `draftRuntimeKeys` and
   `draftBuildPackages` cover every arm (`draftruntime.go:24-56`);
   `TestDraftGitRuntimeMaterializesUnderSourceV1Key`, `TestDraftGitBuildsPublishReceipt3`,
   `TestDraftRuntimeKeysCoverAllArms`. My GC probe (scopes, clone only): a Git-arm v5 marker
   keeps its source-v1 leaf and the old commit-keyed leaf is swept, no warnings — PASS.

## 25-field v4 → v5 migration table (enumerated; all 25 match)

| v4 field | spec v5 | candidate | ✓ |
|---|---|---|---|
| schema_version | `5` | `Write` sets `SchemaV5` when `Package != nil` (marker.go:862); required | ✓ |
| source | replaced by `package` | refused on every arm (marker.go:400-404), not in allowed list, builder never sets | ✓ |
| git | replaced | same | ✓ |
| ref_kind | replaced; declared ref in manifest + bound lock | same; CLI/install rows assert the ref through `lock.CheckStale(manifest)` | ✓ |
| ref | replaced | same | ✓ |
| commit | replaced by `package.commit` | same; `package.commit.hex == locked commit` asserted | ✓ |
| skill_schema_version | actual manifest version 1..8 | builder `node.Spec.SchemaVersion`; reader refuses <1 or >8 for v5 (marker.go:366) | ✓ |
| name | retained, required | required list | ✓ |
| content_sha256 | retained, required | required; grammar `sha256:` hex | ✓ |
| locale | retained, required (nullable) | required; `validNullableLocale` | ✓ |
| agents | retained, required | required; sorted identifier set | ✓ |
| commands | retained, required | required | ✓ |
| dependencies | retained, required | required | ✓ |
| runtime_roots | retained, required | required | ✓ |
| build_roots | retained, required | required (v5 list) | ✓ |
| installed_at | retained, required | required; `validTimestamp` | ✓ |
| files | retained, required | required | ✓ |
| build_source | retained conditional | optional in allowed list; all-or-nothing with local go-v1 builds (`validBuildState`) | ✓ |
| requirements | retained optional | optional | ✓ |
| mcp_servers | retained optional | optional | ✓ |
| activation | retained optional | optional | ✓ |
| requirers | retained optional | optional | ✓ |
| attestation | retained for eligible Git; forbidden local | Git carried (builder + reader + Current); local refused at builder (fail-closed, no write) and reader | ✓ |
| substituted | retained non-empty legacy id; forbidden local | Git carried/compared; local refused at builder and reader; non-empty enforced (`validOptionalNonEmptyString`) | ✓ (B1) |
| builds | retained records with receipt version 3 | builder forces `ReceiptSchemaVersion=3` + execution policy on every entry; `validV5Build` (trunk) | ✓ |
| (new) package | required, closed arms | `validV5PackageShape` closed per arm + nested commit (trunk) | ✓ |
| (new) lock_sha256 | required | required; `sha256:` hex grammar | ✓ |

Required-set parity: candidate v5 required list (15) == schema `required` (15);
allowed list (22) == schema `properties` (22). `TestMarkerV5TopLevelMatchesAcceptedSchema`
and `TestMarkerV5PackageArmsMatchAcceptedSchema` pin both against the vendored bytes.

## Narrowing mutants (mutant clone; each restored with `git checkout --`, final tree = candidate)

| id | mutant | killer | result |
|---|---|---|---|
| M-A | reader loop refuses the five only on local arms + allowed list re-admits them (rev2 shape) | `TestMarkerV5GitRefusesLegacyIdentityBytes` | KILLED exit 1 |
| M-A2 | M-A + builder copies Source/Git/RefKind/Ref/Commit for Git members | `TestDraftGitInstallWritesMarkerV5:147`, `TestDraftGitRuntimeMaterializesUnderSourceV1Key:343` (`install.Project`, raw-bytes/decoded identity) | KILLED exit 1 |
| M-B | `Current` drops the `lock_sha256` comparison | marker: `TestMarkerV5PlanMismatch`, `TestMarkerV5Currentness`; install (production entry): my `TestReviewLockOnlyChangeRestages:87` | KILLED exit 1 (both) |
| M-C1 | builder drops the local-arm attestation guard only | `TestDraftLocalRejectsConfusedAttestation:461` (install still fails closed via the reader: `install marker is invalid for schema 5`, diagnostic narrowed), `TestDraftMarkerArmGuards:531` | KILLED exit 1 |
| M-C2 | reader drops the local-arm attestation refusal only | `TestMarkerV5LocalRejectsEvidenceSummaryBytes`, `TestMarkerV5RefusesMalformedIdentity` | KILLED exit 1 |
| M-C3 | builder drops the local-arm substitution guard only | `TestDraftMarkerArmGuards:536` (builder seam — B1) | KILLED exit 1 |
| M-D | reader loop refuses the five only as non-empty strings | — | SURVIVED exit 0, **equivalent**: `onlyFields(raw, allowed)` (allowed list without the five) still refuses null/empty; proven by M-D2 |
| M-D2 | M-D + allowed list re-admits the five | `…/read/<arm>/<field>/null` and `/empty` rows (8 shown) | KILLED exit 1 |
| M-J | allowed list re-admits the five, loop deleted (delete-only control) | write ×5, read rows, local rows | KILLED exit 1 |
| M-K | reader local-arm attestation refusal narrowed to non-null objects | — | SURVIVED exit 0, **equivalent**: the shared attestation rule (`rawObject` must be an object) refuses `null`; probe with M-K applied: local marker + `"attestation":null` / `"substituted":null` / `"attestation":{}` all refused by `Read`, never Current — PASS |
| M-E | builder drops the Git-arm attestation | `TestDraftGitInstallCarriesRegistryAttestation:387` (production entry) | KILLED exit 1 |
| M-F | `Current` drops the attestation comparison | marker `TestMarkerV5PlanMismatch`; install `TestDraftGitInstallCarriesRegistryAttestation:396` (changed evidence must re-stage) | KILLED exit 1 (both) |
| M-I | `Current` drops the substituted comparison | `TestMarkerV5PlanMismatch` | KILLED exit 1 |
| M-G | builder never carries `substituted` on Git arms | `TestDraftMarkerArmGuards:527` (seam — B1) | KILLED exit 1 |
| M-H | `draftRuntimeKeys` skips Git members (interim boundary) | `TestDraftGitRuntimeMaterializesUnderSourceV1Key:326`, `TestDraftRuntimeKeysCoverAllArms:557` | KILLED exit 1 |

No genuine survivor: the two survivors are behaviour-equivalent mutants of redundant
defense layers, each demonstrated by a combined mutant or a direct probe.

## Identity consumers under the v5 shape (grep redone over non-test code)

- `internal/install/targets.go:44` `stageNode` → `marker.Current`: compares package + lock +
  attestation + substituted — v5-correct.
- `internal/install/install.go:1218-1234` `detectMovedTagsIn`: legacy advisory warning,
  never matches v5 (`RefKind == ""`), read-only; draft tag moves surface through explicit
  refresh → new lock → non-current (clause 3). Bound, as the spec allows.
- `internal/scopes/gc.go:211-229`: v5 marks the source-v1 leaf from `Package.Digest()`
  (probe above, Git arm included).
- `cmd/curator/main.go:1000-1045` `scopeStatusDrift`: `manifest.Load` without the draft
  option → schema-2 manifests yield an empty drift map; `internal/registry/attest.go:41-50`
  `AttestRoot`: v5 → `unattestable: marker lacks commit or hash` (conservative, never
  "attested"); `internal/ui/state.go:131-138`: blank ref/commit display. All read-only,
  none authorizes; the draft status/attest/TUI surface belongs to TASK-260910-3eu4cy
  ("Status, repair and refresh enforce exact currentness") — B2.
- Also found (not in the producer's grep): `internal/envprofile/import.go:419-429`
  `recoverSkill` — profile import recovers a declaration only from a legacy marker
  (`m.Git`/`m.Source` + ref + commit); a v5 installation is simply not recoverable (falls
  through, fails closed, fabricates nothing). Outside this leaf; B3.
  (`internal/contextlock` and `internal/envprofile/managed.go` matches are other types.)

## Findings (non-blocking) and bounds

- F1 (pre-existing, follow-up): `marker.Read` accepts a v5/v4 external `go-repository-v1`
  build record that omits the `substituted` boolean (spec corpus
  `invalid-external-missing-substituted.json`; spec: "An external record MUST NOT omit these
  fields"). Trunk `validDriverBuild` (20sx61) — not in this leaf's delta or scope line; the
  writer always emits the field. Recommend a small BUG item (close the external build-record
  raw shape: `substituted` required, `substitution` typed) under the build-record owner.
- F2 (spec corpus): `schema-cases/install-marker-v5/valid.json` is schema-valid but violates
  the prose all-or-nothing `build_source` rule; Curator's reader (correctly) refuses it.
  Worth a curator-spec note; no Curator change.
- F3 (minor test hygiene): `assertV5GitMarker` checks the declared tag with
  `strings.Contains(manifest, "tag":"vN")`, which is trivially true in the alias test (both
  tags appear); the alias semantic there is pinned by the dry-run report (`review tag v2`) and
  the selection index. Could parse the manifest member the lock binds; not required.
- B1: Git-arm `substituted` positive is unreachable through `install.Project` (accepted
  3vxe3y draft-lane refusal); proven at the builder seam + reader.
- B2: draft status drift / `status --attest` / TUI display for v5 deferred to 3eu4cy.
- B3: envprofile import of v5 installations (fails closed) — outside scope.
- B4: Windows and Ubuntu rows verified by the hosted gate only; rose-air lane skipped on the
  gate snapshot; no new skip classes in the candidate (no POSIX-only fixtures added).
- B5: value grammars behind the schema's `$ref`s are covered by `validMarker` unit rows and
  my jsonschema run over real documents, not by the in-repo `validateV5AgainstSchema`
  (closed-shape pin only) — as the candidate states.

## Addendum — five re-pointed CLI tests, verbose rerun
(appended below by the driver before the verdict was attached)

`go test -p 1 ./cmd/curator -run 'TestProjectResolveGitAliasSelectionThroughCLI|TestProjectResolveLegacyConfiguredGitThroughCLI|TestProjectResolveLegacyNetworkGitThroughCLI|TestProjectResolveGitTagDiffersFromHEADThroughCLI|TestProjectResolveGitRealInstallThroughCLI' -count=1 -timeout=300s -v` → exit 0, `ok … 53.351s`; all five ran (none skipped on macOS):
Alias (7.63s), LegacyConfigured default/custom-source (5.63s/5.76s), LegacyNetwork default/custom-source (7.06s/5.89s), TagDiffersFromHEAD (7.17s), RealInstall incl. offline pinned reinstall (13.49s) — PASS.

Shell note: driver scripts ran under `#!/bin/bash` with `set -o pipefail` and logged real exit codes; ad-hoc zsh tool calls that piped `go test` through `grep` report the pass/fail from the `ok`/`FAIL` lines (zsh has no `PIPESTATUS`).
Review clone and mutant clone both end at tree `63c32caedb01fa50620fd39f52d6c74ed6adbd53` with no leftover probe files; the Story worktree was not modified.
