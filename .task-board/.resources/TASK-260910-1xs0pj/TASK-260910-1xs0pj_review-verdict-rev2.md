# Review verdict — TASK-260910-1xs0pj revision 2 (CR-TASK-260910-1xs0pj-2)

**Verdict: CHANGES_REQUESTED → `to-dev`.**
Reviewer run RUN-260918-f29089 (claude-opus-5:max), 2026-09-19. Read-only review;
all mutation/probe work ran in a disposable clone, never in the Story worktree.

## Candidate identity (verified)

- Story worktree working tree (tracked + untracked, temp index `git write-tree`)
  = `9e0b7a17d74fb502fc9020b52d94e88f8a76deb5` = the Change Request candidate tree.
- Patch `TASK-260910-1xs0pj_change-request_rev2.patch` sha256
  `f65e1d4f686bced9f230f7a65a053643f35eb83537114ac270cd1d2a79f53bf0` (matches).
- Disposable clone: `git clone --no-checkout` control root → `checkout --detach e857e50`
  → `git apply --index` rev2 patch → commit; tree `9e0b7a17…` (exact).
- Hosted gate run 35383914475: `gh run view` headSha `60cb96ab1bedff98c79739fe54a7410a4c59ddd4`,
  conclusion `success`; `60cb96ab^{tree}` = `9e0b7a17…`, parent `e857e50…` (base). The gate
  ran the exact candidate. Windows/Ubuntu rows are verified by that gate only.

## Blocking finding — v5 Git markers carry the five replaced legacy fields (contract violation)

**Claim.** Revision 2 writes, and `marker.Read` admits, marker-schema-5 documents for Git
packages that carry top-level `source`, `git`, `ref_kind`, `ref`, `commit`. The accepted
contract replaces exactly those five with `package` and forbids them on every v5 arm.

**Normative source (curator-spec main, landed a4fcaf02 / current 802caee):**
- `protocol/skillfile-sources.md` §"schema-2 installations": "Its `package` replaces legacy
  `source/git/ref_kind/ref/commit` fields; its `lock_sha256` binds the installed selection
  and declared ref through the validated lock and matching manifest. The following migration
  is exhaustive". Table row: `source, git, ref_kind, ref, commit` → "Effective `package`;
  declaration/ref selection in the manifest and bound lock."
- `schemas/draft-sources-v1/install-marker-v5.schema.json`: `additionalProperties: false`;
  its property set has no `source`/`git`/`ref_kind`/`ref`/`commit`.
- `conformance/draft-sources-v1/README.md` specification check (exhaustive v4 migration
  audit): `replaced = {'source','git','ref_kind','ref','commit'}`;
  `assert set(v5['properties']) == (set(v4['properties']) - replaced) | {'package','lock_sha256'}`.
- Spec Git-arm positives `valid-attested-network.json`, `valid-legacy-substitution.json`
  carry the package alone; `invalid-unknown-top-level.json` shows any extra top-level key
  is invalid.

**Where (file:line, candidate tree):**
- `internal/marker/marker.go:316-318` — v5 `allowed` list now appends
  `"source", "git", "ref_kind", "ref", "commit"`.
- `internal/marker/marker.go:409-446` (`validV5Identity`) — the five are refused only for
  `local-snapshot`; Git arms admit `source` (portable path), `git`, and an all-or-nothing
  legacy triple. Base e857e50 (accepted 17ps6u rev4) refused all five on every arm
  (`for _, field := range []string{"source","git","ref_kind","ref","commit"} { if present → false }`).
- `internal/install/draftruntime.go:194-206` (`buildDraftMarker`) — for `member.Package.IsGit()`
  copies `node.Decl.Source`, `node.Decl.Git` and the resolved `Kind/Ref/Commit` into the v5
  marker; `Write` emits them (omitempty, non-empty).
- Tests pinning the violating shape: `internal/install/draftmarker_git_test.go:146-163`
  (`TestDraftGitInstallWritesMarkerV5` requires raw `source/git/ref_kind/ref/commit` present),
  `TestDraftGitMarkerIdentityRoundTrip`, `TestDraftMarkerArmGuards` Git rows;
  `internal/marker/marker_v5_git_test.go` `TestMarkerV5GitRoundTrip`,
  `TestMarkerV5GitIdentityValidation` (write-level acceptance of the triple).

**Reproduction (real production path, my run in the clone):**
1. Throwaway test in the clone (not in the candidate) calling `setupGitInstall` +
   `install.Project` (`draftRealInstall`, `DraftSourcesV1=true`) and dumping
   `.agents/skills/review/<marker>`: `go test -p 1 ./internal/install/ -run TestReviewDumpGitMarkerV5`
   → exit 0. Written document: `schema_version 5`, `source "review"`,
   `git "https://example.org/kit.git"`, `ref_kind "tag"`, `ref "v1"`,
   `commit "b39576f9…"`, `package {kind network-git, repository example.org/kit,
   commit {sha1, b39576f9…}, directory skills/review}`, `lock_sha256 sha256:4901334f…`.
2. jsonschema 4.25.1 Draft 2020-12 with the spec registry (`schemas/v1` + `schemas/draft-sources-v1`),
   validator `install-marker-v5.schema.json`:
   - control: spec positive `valid-attested-network.json` → **valid = True**
   - candidate-written v5 Git marker → **valid = False**:
     `$: Additional properties are not allowed ('commit', 'git', 'ref', 'ref_kind', 'source' were unexpected)`
   - the same document minus those five keys → **valid = True**
   (probe exit 0 = document invalid as asserted).

**Consequences.**
- AC "complete 25-field v4 migration … exactly as the table says": 20 of 25 fields match
  (enumeration below); the 5 replaced fields are neither dropped nor forbidden on Git arms.
- Closed raw-shape validation is now wider than the normative schema: `marker.Read` accepts
  documents the schema rejects, so the reader no longer fails closed on the five fields for
  Git arms (foreign bytes with `ref_kind` etc. are admitted where trunk refused them).
- Every v5 Git marker this build writes is non-conformant; the spec's own conformance
  check would refuse it. The hosted "Interop conformance gate" did not catch this because
  curator's CI pins `conformance/v1` (frozen), not the draft-sources-v1 corpus.

**Root cause — a directive/contract conflict the orchestrator must resolve (not human-only).**
Rework-1 required "a v5 marker for a Git package carries and reads back kind/ref/commit
exactly as the accepted 3vxe3y tests expect" and "make the five tests pass unchanged". Those
tests (cmd/curator/draft_sources_test.go:561-566, 645-650, 727-732, 899-904, 1292-1299;
landed 6645b9b, 2026-09-17) assert `marker.Read(dir).RefKind/Ref/Commit` (and `.Source`) —
the **legacy** marker shape Git draft members wrote under 17ps6u's explicitly *interim*
boundary ("Git draft members keep their accepted legacy shape until the integration leaf
migrates them"). This leaf is that migration. A spec-conformant v5 marker cannot read back
`RefKind == "tag"` from the installed directory alone: the package carries
repository+commit+directory and the declared ref lives in the manifest/lock, bound through
`lock_sha256`. So the directive's literal form and the accepted contract cannot both hold.
The accepted contract (operator-approved, landed) wins; the five tests must be re-pointed at
the v5 identity (same semantic — "tag vN at the locked commit" — v5 shape: package kind,
`package.commit.hex == commit`, `lock_sha256 == lock.LockSHA256`, declared ref from the lock
member). This is not a re-opening of 3vxe3y's accepted decision (identity is recorded and
compared); only the interim on-disk shape changes, exactly as 17ps6u and the spec foresaw.
Recommendation to the orchestrator: lift rework-1's "unchanged" constraint for these five
tests; no spec change is needed.

## Required rework (revision 3)

1. Restore the closed v5 shape: `validV5Identity` refuses `source/git/ref_kind/ref/commit`
   on **every** v5 arm (trunk's loop); drop the five from the v5 `allowed` list; delete
   `validLegacyTriple`'s v5 use (keep it for the legacy lane if you like);
   `buildDraftMarker` copies none of the five for Git arms. Keep everything else from rev1/rev2:
   attestation + substituted carried for Git arms, refused for local arms (builder fail-closed
   + reader), `draftRuntimeKeys`/`draftBuildPackages` covering all arms, receipt-3 binding.
2. Re-point the five 3vxe3y CLI tests at the v5 identity (needs the orchestrator's lift of
   the "unchanged" constraint — flagged above): schema 5, `Package.Kind`
   (`network-git` / `configured-git`), `Package.Commit.Hex == commit`, `LockSHA256 ==
   lock.LockSHA256`, and the lock member's declared ref `tag vN`. Do not weaken the
   semantic; only the field path changes.
3. Regression rows for the shape: at `install.Project` (both Git arms) the raw v5 document
   contains none of the five (mirror of trunk `TestMarkerV5LocalRoundTrip:87-91`); read-level
   negatives: foreign v5 Git bytes carrying each of the five (string, null, empty) are refused
   by `Read` and never `Current`. Optional but recommended: a test that validates a written
   Git-arm and local-arm v5 marker against the spec's `install-marker-v5.schema.json` when
   `CURATOR_CONFORMANCE_ROOT`-style env points at the draft-sources-v1 corpus (the
   `TestReadAuthoritativeMarkerV4SchemaCases` pattern), so the interop gate can catch this class.
4. Declared-ref currency through the lock, not the triple: re-point
   `TestDraftGitMarkerIdentityRoundTrip`'s "moved declared ref → non-current" row at the
   real mechanism (re-resolve to tag v2 → new lock → `lock_sha256` mismatch → `Current`
   false, reinstall re-stages), plus a narrowing mutant that drops the lock comparison
   (my MB shows `TestMarkerV5PlanMismatch/lock_sha256` and `TestMarkerV5Currentness` kill it).
5. Identity consumers under the v5 shape (grep item from rework-1, redo it against the
   corrected shape): `internal/install/install.go:1229 detectMovedTagsIn` reads
   `RefKind == "tag"` → for v5 members derive the declared tag from the lock member or
   record it as an explicit bound (the spec says draft tag moves are detected at explicit
   refresh; the warning never authorizes). `cmd/curator/main.go:1035 scopeStatusDrift`,
   `internal/registry/attest.go:41-50 AttestRoot`, `internal/ui/state.go:131-137` — state
   which compare the package for v5 today and which are deferred to TASK-260910-3eu4cy
   (Status/repair/refresh); do not leave "reads the legacy fields" as silent behaviour.
6. Evidence as before (narrow, bounded, exit codes): marker package; the five CLI tests
   (updated); the install subset; vet/gofmt; append "Revision 3" to results.md; handoff.

## 25-field v4 → v5 enumeration (independent source: `schemas/v1/install-marker-v4.schema.json`, 25 properties)

| v4 field | table says | candidate rev2 | ok |
|---|---|---|---|
| schema_version | `5` | `SchemaV5` const, required | ✓ |
| source | replaced by `package` (not a v5 property) | admitted+written on Git arms; forbidden local | ✗ |
| git | replaced (not a v5 property) | admitted+written on Git arms; forbidden local | ✗ |
| ref_kind | replaced (not a v5 property) | admitted+written on Git arms; forbidden local | ✗ |
| ref | replaced (not a v5 property) | admitted+written on Git arms; forbidden local | ✗ |
| commit | replaced (not a v5 property) | admitted+written on Git arms; forbidden local | ✗ |
| skill_schema_version | actual manifest version 1..8 | required; `<1 || >8` refused | ✓ |
| name, content_sha256, locale, agents, commands, dependencies, runtime_roots, build_roots, installed_at, files (10) | retained, required | all in the v5 required list; existing validators | ✓ |
| build_source, requirements, mcp_servers, activation, requirers (5) | retained, conditional/optional | in the v5 allowed list; existing conditional validators (build_source via validBuildState) | ✓ |
| attestation | retained for eligible Git; forbidden local | Git: carried exactly as the plan selected (`request.attestations`); local: builder fails closed `source_member_invalid`, reader refuses | ✓ |
| substituted | retained non-empty legacy id; forbidden local | Git: `node.Substituted` carried, non-empty validated; local: builder + reader refuse; new selectors cannot enable it (install/closure/strict gates, 2 rows) | ✓ |
| builds | retained, receipt version 3 | required; dufdai receipt-3 binding, now for all draft arms | ✓ |
| (new) package, lock_sha256 | required | required; closed package union shape; sha256 grammar | ✓ |

20/25 exact; 5/25 violate the table and the schema.

## What I verified independently (clone at tree 9e0b7a17; shell zsh→bash scripts, `set -o pipefail`)

| command | exit |
|---|---|
| `go build ./internal/marker/ ./internal/install/ ./cmd/curator/` | 0 |
| `go vet ./internal/marker ./internal/install` | 0 |
| `gofmt -l internal/marker internal/install cmd/curator` | 0, no output |
| `go test -p 1 ./internal/marker/ -count=1` (whole package) | 0 (1.4s) |
| `go test -p 1 ./cmd/curator/ -run 'TestProjectResolveGitAliasSelectionThroughCLI\|…LegacyConfiguredGit…\|…LegacyNetworkGit…\|…GitTagDiffersFromHEAD…\|…GitRealInstall…\|TestMarkerRefusalSeparatesUnsupportedFromInvalid' -count=1` | 0 (63.0s) |
| `go test -p 1 ./internal/install/ -run 'TestDraftGitInstallWritesMarkerV5\|TestDraftGitMarkerIdentityRoundTrip\|TestDraftGitRuntimeMaterializesUnderSourceV1Key\|TestDraftGitInstallCarriesRegistryAttestation\|TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker\|TestDraftLocalRejectsConfusedAttestation\|TestDraftGitBuildsPublishReceipt3\|TestDraftSelectorSubstitutionForbidden\|TestDraftMarkerArmGuards\|TestDraftRuntimeKeysCoverAllArms\|TestLegacyInstallUntouchedWhenDraftOff\|TestLegacyBuildsKeepReceipt1' -count=1 -v` | 0 (48.6s; 12/12 PASS) |
| schema probe (real install → jsonschema against install-marker-v5.schema.json) | document invalid (see above) |

Tests green ≠ contract met: the green rows pin the violating shape. Accepted from producer
evidence + hosted gate, not rerun by me: `TestDraftAudit*`, `TestDraftBuild*`, `TestDraftLocal*`,
`TestDraftPlan/SSH/Transport/Manager/Sources` subsets, full packages, Windows/Ubuntu lanes.

### Clauses that hold (positive + negative rows at the production entry)
- attestation present exactly when the plan selects evidence, equal incl. key absence:
  `TestDraftGitInstallCarriesRegistryAttestation` (+/−); revoked evidence refuses despite an
  attested marker, prior marker byte-identical: `TestDraftInstallRevokedEvidenceRefusesDespiteAttestedMarker` (−).
- local arm rejects attestation/substituted: builder fail-closed before staging
  (`TestDraftLocalRejectsConfusedAttestation`, nothing live), reader refuses foreign bytes
  (`TestMarkerV5LocalRejectsEvidenceSummaryBytes`, incl. the five legacy fields).
- new selectors cannot enable legacy substitution: `TestDraftSelectorSubstitutionForbidden`
  (draft-selectors-refuse `source_selection_invalid`, strict-refuses-first).
- package/lock binding: `TestDraftGitInstallWritesMarkerV5` (exact locked package + lock),
  `TestMarkerV5PlanMismatch` package/lock_sha256/registry/status/key/substituted rows
  non-current; legacy↔v5 interchange non-current both ways (`TestMarkerV5LegacyInterchange`).
- runtime key + receipt-3 migrate with the marker for all arms: `TestDraftRuntimeKeysCoverAllArms`,
  `TestDraftGitRuntimeMaterializesUnderSourceV1Key`, `TestDraftGitBuildsPublishReceipt3`.
- frozen lane byte-identical with the switch off: `TestLegacyInstallUntouchedWhenDraftOff`,
  `TestLegacyBuildsKeepReceipt1WithTheSwitchOff` (marker.go legacy branch unchanged except the
  `validLegacyTriple` extraction with identical predicate).
- summaries never authorize: all install gates run before staging; `Current` reports only.

### Narrowing mutants (mine, clone; file restored after each, final tree 9e0b7a17)
| mutant | narrow test | result |
|---|---|---|
| MA `validV5Identity` local arm refuses `substituted` only when attestation also present | `./internal/marker -run TestMarkerV5` | KILLED exit 1 (`…LocalRejectsEvidenceSummaryBytes/substituted`, `TestMarkerV5RefusesMalformedIdentity`) |
| MB `Current` keeps package comparison, drops `lock_sha256` comparison | `./internal/marker -run TestMarkerV5` | KILLED exit 1 (`TestMarkerV5PlanMismatch/lock_sha256`, `TestMarkerV5Currentness`) |
| MC `buildNodeMarker` passes `nil` attestation into the v5 builder (legacy lane keeps it) | `./internal/install -run 'TestDraftGitInstallCarriesRegistryAttestation\|TestDraftInstallRevoked…'` | KILLED exit 1 (`TestDraftGitInstallCarriesRegistryAttestation`) |
| MD identity triple dropped again (rev1 signature) | `./internal/install -run 'TestDraftGitInstallWritesMarkerV5\|TestDraftGitMarkerIdentityRoundTrip'` | KILLED exit 1 — note: killed by tests that pin the non-conformant shape; after rework the equivalent guard must be the lock binding (MB) and the schema-shape rows |
| ME builder local guard narrows the substitution refusal to attested nodes | `./internal/install -run 'TestDraftMarkerArmGuards\|TestDraftLocalRejectsConfusedAttestation'` | KILLED exit 1 (`TestDraftMarkerArmGuards`, helper-direct — production entry cannot reach the row; stated bound) |
| MF Git-arm `source` admitted without the PortablePath grammar | `./internal/marker -run TestMarkerV5` | KILLED exit 1 (`…GitIdentityValidation/{write,read}/bad-source`) — moot after rework |

No survivors; MD/MF become moot once the five fields are refused again.

## Minor observations (non-blocking, for the rework)
- `Current` never compared `Source`/`Git` (legacy lane either) — irrelevant once the fields go.
- `TestDraftMarkerArmGuards` is helper-direct; keep it labelled as the bound it is.
- Producer results.md is accurate about commands/exit codes; its claim "no internal/marker
  production change was needed" (rev1) was true and is the correct end state for the five fields.

## Host note
The host incident (syspolicyd crash → stranded exec of unsigned binaries; see the
orchestrator's EPIC note) made several shell calls stall >120 s during this review; every
result above comes from a completed process with a recorded exit code.
