# Review verdict — TASK-260910-1xya7x revision 3 (CR-TASK-260910-1xya7x-3)

Reviewer run: RUN-260919-5cc262 (claude-opus-5, reviewer archetype), 2026-09-20.
Verdict: **CHANGES_REQUESTED** (revision 3 → 4), routed `to-dev`. The harness and the corpus
execution are sound and every measured claim I could re-derive holds (§1–§7); what fails the binding
review note is the *record*: (a) the semantic runner has no executed-count check — mutant H2 (a
silently dropped case) SURVIVES (§8), the schema runner's equivalent is killed; (b) four of the five
"semantic bounds" are driven rows that FAIL the draft expectation (product gaps, one security-
relevant) and the results file files them under "Explicit bounds (never passing)" — hidden failures
by the note's definition (F1); (c) the platform coverage record omits the per-lane numbers (F5) and
the CLI "transaction restore" row is a prepare-time refusal, not a rollback (F7). The required rework
(§11) is small and test/report-only; the acceptance-relevant facts are all measured below so the
producer needs no re-investigation. F1.1–F1.3 are product follow-ups OUTSIDE this leaf's scope line
(surfaced by this leaf, not caused by it) — the orchestrator should open them regardless of rev4.

## 1. Exact candidate identity (verified, not trusted)

- Story worktree `.temp/STORY-260910-1cnwwp/worktree`: HEAD `71e353e6` (stbg4d rev2 checkpoint on
  base `97a161da`); temp-index `git read-tree HEAD && git add -A internal/ cmd/ && git write-tree`
  = `2d400147c6868064d56e01e360f7e1e472427187` = the CR candidate tree OID. Worktree not modified.
- Patch `TASK-260910-1xya7x_change-request_rev3.patch` sha256
  `3b587c9c5f65c6c4b7219658a72cc690ef492b5bace423ade5e658b59d83052e` (matches the CR record).
- Disposable clones `/tmp/1xya7x-rev3/cand` and `/tmp/1xya7x-rev3/probe`: `git checkout --detach
  97a161da && git apply --index <rev3 patch> && git commit` → `HEAD^{tree}` = `2d400147…` in both;
  every driver prints `git write-tree` before and after and both clones ended at `2d400147…`.
- Leaf delta vs the checkpointed sibling (`git diff --name-status 71e353e6 2d400147`): **153 paths,
  all `A` (13 759 insertions, 0 deletions, 0 modifications)** — 18 `internal/crossconformance/
  draftsources_*_test.go`, `internal/testcli/cli.go`, and the vendored corpus under
  `internal/crossconformance/testdata/draft-sources-v1/`. No production file, no `conformance/v1`,
  no `schemas/v1`, no release/SPEC_PIN file is touched by this leaf. Frozen v1 goldens are
  byte-identical by construction.
- Hosted gate run 35469802052 (`gh run view --json headSha`): head `1b995a85a3a8…`,
  `git rev-parse 1b995a85^{tree}` = `2d400147…` — the gate ran the exact candidate. Jobs: Test
  ubuntu 4.4 min, macOS 12.6 min, Windows **70.5 min** (all success); Race/Lint/Naming/Gate
  self-test/Interop conformance gate all success; rose-air and Candidate suite skipped
  (rose-air lane unverified, as the producer states).

## 2. Corpus pin — verified against curator-spec

- `DRAFT_SOURCES_PIN` = `802caee548ddc8b19408746d26c7972d39b39cc2` = curator-spec `main` HEAD
  ("Close the dotfile-manager heuristic…"). `git archive 802caee conformance/draft-sources-v1
  schemas/draft-sources-v1` vs the vendored `corpus/` tree: `diff -rq` **identical** (0 differences).
- `TestDraftSourcesPin` re-hashes every vendored byte against `MANIFEST.sha256` (no missing/extra
  files) — passes; the pin constant in `draftsources_corpus_test.go` equals the pin file.
- Accepted-contract subset (a4fcaf0, the AC's 102/73): `ac-schema-cases.txt` (102) == the a4fcaf0
  `index.json` instances, same order; `ac-semantic-cases.txt` (73) == the a4fcaf0 semantic ids, same
  order. `a4fcaf0..802caee` is purely additive: +13 `source-policy-v2` schema cases (+ its schema),
  +21 `v2-*` semantic cases; every a4fcaf0 schema entry (schema/instance/valid) and every a4fcaf0
  semantic case (full JSON) is unchanged at the pin; `snapshot-cases.json` byte-identical (3 vectors).
  So "115/94/3 at the pin" ⊇ "102/73/3 of the AC" exactly as the producer states.

## 3. Independent execution on this host (macOS 15.6, go1.26.0 darwin/amd64, load 7–11)

Driver `/tmp/1xya7x-rev3/run-full.sh` (`#!/bin/bash`, `set -o pipefail`) in the `cand` clone:

| step | result |
|---|---|
| `gofmt -l internal/crossconformance internal/testcli` | exit 0, 0 files |
| `go vet ./internal/crossconformance/ ./internal/testcli/` | exit 0 |
| `go build ./...` | exit 0 |
| `go test -p 1 -count=1 -timeout=1800s -v -run 'Draft' ./internal/crossconformance/` | **exit 0**, `ok … 451.758s` |

Measured from the `-v` log: **243 `=== RUN` / 243 `--- PASS` / 0 SKIP / 0 FAIL**; subtests executed:
`TestDraftSourcesSchemaCases/*` **115/115**, `TestDraftSourcesSnapshotVectors/*` **3/3**,
`TestDraftSourcesSemanticCases/*` **94/94** (every corpus id, none sampled); harness summary line
`schema cases: 110 driven, 5 bounds, 115 total`; 10 `BOUND` log lines (5 schema + 5 semantic).
Pre-existing (v1) crossconformance tests: `go test -v -skip 'Draft' ./internal/crossconformance/`
in the `probe` clone — see §8 (result appended there); the hosted gate ran the whole package green
on all three OSes.

## 4. Per-lane coverage from the gate ledgers (`test/observed-cases.tsv`, `skips-observed.tsv`)

| lane | package time | schema | snapshot | semantic | other draft skips |
|---|---|---|---|---|---|
| macOS (hosted) | 214 s | 115/115 | 3/3 | 94/94 pass, 0 skip | none |
| ubuntu (hosted) | 111 s | 115/115 | 3/3 | 93 pass, 1 skip (`case-alias`: host-capability "case sensitivity") | none |
| windows (hosted) | **2447 s** | 115/115 | **0/3 skipped** (platform-control "Windows does not expose portable executable permission bits…") | **45 pass, 49 skip** ("test transport wrapper is POSIX-only") | `CLILocalSkillScriptDependencies/shims_execute_the_frozen_runtime` ("executes POSIX skill commands"), `CLIInstallRestoresPriorState` ("this process can write through a read-only directory…") |
| rose-air | — | unverified (lane skipped) | | | |

Every observed skip reason matches a declared class in `.github/ci/skip-classes.tsv` verbatim
(platform-case gate green); no new class. The 49 Windows semantic skips are exactly the rows that
need the POSIX fake-`git` shim: 13 transport (`fallback-*`, `pinned-auth`,
`endpoint-identity-mismatch`), `attested-network-current`, 8 `attestation-evidence-*` (these skip
at layer 3 only — layers 0–2, registry.Resolve/install.Project/RefreshDraft, DID execute on Windows
but Go reports the subtest as skipped), 6 `marker-plan-mismatch-*`, 21 `v2-*`. The precedent for the
class is the accepted cmd/curator transport suites (`draft_transport_test.go`, `draft_sources_test.go`).
Windows therefore proves: all 115 schema readers, the boundary/selection/capture/currentness
install-level rows, and all 18 external-evidence rows; it does not prove the transport fetch loop or
the snapshot byte vectors (the executable bit is part of the pinned digest). **The producer's results
file does not state these numbers** ("ubuntu/windows: hosted gate only") — recorded here (F5).

Windows budget: `internal/crossconformance` 2447 s (`TestDraftSourcesSemanticCases` 2275 s, of which
the 17 `external-evidence-mismatch-*` rows cost 100–240 s each ≈ 2000 s — serial install+repair of
the external build pipeline; macOS 117 s for the same test). Within the 120m per-package budget; the
Windows job wall time (70.5 min) is still bounded by `internal/install` (3808 s in this run). Note for
landing (F8): the package is now the second-longest Windows package; the external rows use separate
temp dirs and could take `t.Parallel()` if it grows.

## 5. Production-entry audit of every row (what actually reaches production)

Schema (115): `skillfile-v2` → `manifest.LoadWithOptions(DraftSourcesV1)`; `source-policy-v1/v2` →
`config.ParseSourcePolicy` (+ `repository_policy_invalid` class asserted on negatives);
`skillfile-lock-v1` → `sourcelock.Parse`; `install-marker-v5` → `marker.Read`; `build-receipt-v3` →
`protocoljson.MarshalCanonical` + `buildmeta.DecodeReceipt`; `source-audit-v1` →
`audit.ParseSourceAudit`; `local-snapshot-v1` (3) → bound (no byte reader; shape via Capture vectors).
No JSON-Schema engine is executed (disclosed). Bound set asserted in both directions; each bound row
asserts the bound is still real (a fixed gap flips the row). Two "driven" positives —
`skillfile-lock-v1/valid-git.json`, `valid-configured-git.json` — are schema-valid but refused by
`sourcelock.Parse` with `source_member_invalid` because member `directory` ≠ package `directory`
(spec: "these two directory fields MUST agree"); the harness asserts the refusal (inverted, prose-
derived oracle) and counts them as driven — consistent treatment would list them with the marker
`valid.json` bound (F6, minor). The two marker bounds are the pre-existing reader gaps F1/F2 from the
1xs0pj review (`invalid-external-missing-substituted` accepted; `valid.json` refused by the
all-or-nothing rule) — unchanged trunk behaviour, correctly not counted as passes.

Snapshot (3): `snapshot.PrepareLocalAcquisition` + `snapshot.Capture` over a materialised tree;
per-file sha/executable and the inventory digest equal the pinned values; SKILL.md frozen across
vectors; 3 distinct identities. Unix-only (see §4).

Semantic (94) — entry per group:
- boundary 1–7: `snapshot.ValidateLocalPackage` (production admission gate; called by
  `PrepareLocalAcquisition` ← `closure/resolve.go:579`) for 6 rows; `write-boundary-retarget` through
  `install.Project` with a `transaction.Hooks` observer that swaps the output between prepare and
  commit → `source_output_overlap` + digest-identical rollback.
- selection 8–12: `manifest.LoadWithOptions` (unknown alias), `manifest.Expand` (missing/invalid
  members), `sourcelock.New/Parse` (duplicate names — the corpus input is lock `members`), install
  (frozen membership; fixture weakness in F7).
- capture/audit 13–20: `snapshot.Capture/PublishLocal/OpenLocal`, `audit.CheckSourceAudit`,
  `closure.ResolveDraft/RefreshDraft`, `install.Project`; `capture-mutation` is the one real bound
  (`captureAfterCopyHook` is unexported — verified).
- transport 21–33: compiled CLI `project resolve` with a default-deny fake `git` on PATH: clone log,
  exit code, stderr class + `endpoint 2: not attempted`, no URL leak, lock present/absent, alternate
  commit bound; plus `TestDraftTransportFallbackGate` on every OS.
- currentness/marker 34–39, 51–56: compiled CLI `project resolve`/`install`/`status`/`status --check`
  over a signed httptest registry stub + fake git; `install.Project` for the strict/forbidden
  substitution refusals with before/after tree digests.
- evidence 41–50: 8 rows in 4 layers (`registry.Resolve` static fetch → `install.Project` fresh +
  repair over the live stub, digest-preserved → `closure.RefreshDraft` cannot launder → CLI `status`
  nonzero/read-only on unix); 2 rows (wrong-name, wrong-context) driven at `registry.Resolve` and
  FAILING the draft expectation (F1).
- external 40, 57–73: baseline `install.Project` with the fake toolchain/builder (`xbBuildDeps`) and
  the real protected cache; each recorded field mutated; repair via `install.Project` re-derives the
  honest record (protected artifact bytes re-verified; `input.package` tamper forces a rebuild —
  builder call count asserted). The **status half is not driven** at a production entry: the
  arm-vs-receipt comparison is done by the test, and the CLI half is a single shared fail-closed row
  (`TestDraftExternalStatusFailsClosedWithoutSource`, `build_repository_source_unavailable`) — disclosed
  by the producer as "external-CLI status half fail-closed" (F4).
- v2 74–94: `config.ParseSourcePolicy` + `config.ResolveRepositoryEndpoints` on every OS, the
  compiled CLI on unix (clone log = zero attempts for every refusal; canonical lock identity for
  positives); external lane 92–94 via `install.Project` + `ExternalDeps{GitTool: wrapper,
  DraftTransportResolution}` (mirror fetched exactly once per resolution) and
  `buildrepo.ValidateTransportPlan` for the strict-lane class (the install-level diagnostic is masked
  as `build_repository_source_unavailable` — logged, F9); 2 rows (`v2-alias-resolution`,
  `v2-user-insteadof-ignored`) driven and FAILING the draft expectation (F1).

End-to-end (AC clauses): `TestDraftSourcesCLILocalSkillScriptDependencies` (compiled `curator`:
bootstrap → project add → resolve → install; runtime store `runtime/<name>/<SourceV1Key>` frozen
bytes, no links; marker v5 binds package + `lock_sha256`; adapter mirror; shims reach the protected
store and execute on unix); `TestDraftSourcesBrokerAskpassDispatch` (production binary copied under
`HTTPSBrokerName`: 2 prompts answered, 9 silent refusals + the plain-binary control);
`TestDraftSourcesCLIInstallRestoresPriorState` (see F7 — the failure is at transaction *prepare*);
`TestDraftSourcesCLIProjectPathWithSpaceAndUnicode` (resolve + `install --dry-run` only, F7),
`TestDraftSourcesCLISymlinkMemberRefused`, `TestDraftSourcesCrossCompile` (GOOS=windows/linux builds).

## 6. Narrowing mutants (probe clone; `git checkout --` restore is safe there — candidate committed)

| id | mutation (production unless noted) | mask | result |
|---|---|---|---|
| M1 | `buildrepo.AllowSecondAttempt` also advances on `FailureTLS` | `TestDraftTransportFallbackGate|…/fallback-tls$` | exit 1 **KILLED** (both) |
| M2 | `cmd/curator/draft_status.go:134` drops `recorded.LockSHA256 != lockSHA256` | `…/marker-plan-mismatch-lock_sha256$` | exit 1 **KILLED** |
| M3 | `draft_status.go:145` `recorded.Substituted != ""` → `false` | `…/marker-plan-mismatch-substituted$` | exit 1 **KILLED** |
| M4 | `registry.Matches` admits every record (`if true`) | `…/attestation-evidence-wrong-commit$` | exit 1 **KILLED** |
| M5 / M5b | `snapshot/boundaries.go` physical-path overlap check neutralised (M5b: + admitted-input check) | `symlink-managed|managed-source|case-alias` | exit 0 **equivalent** — the lexical check at `boundaries.go:153` (`staging.IsOutputPath(candidate…)`) resolves file identity itself, so layers 153/177/356 are redundant |
| M5c | all three overlap layers neutralised | same | exit 1 **KILLED** (managed-source, symlink-managed, case-alias fail) |
| H1 (harness) | schema runner iterates `entries[1:]` (silent drop) | `SchemaCases|CorpusCounts|Pin` | exit 1 **KILLED**: `driven(109)+bound(5) != total(115)` |
| H2 (harness) | semantic runner iterates `cases[1:]` (silent drop, line 28 only) | `SemanticCases|SemanticCoverage|CorpusCounts` | see §8 |

Producer-reported mutants (rev1 all killed; rev3 lock-exclusion mutant killed) were not re-run; the
table above is my own evidence.

## 7. Review probes (all PASS in the probe clone, exit 0, 21.0 s; source in the evidence bundle)

- **A** frozen-membership with `include:["*"]`: lock bound to `review`, live tree grows `docs`,
  `install.Project` publishes exactly `[review]` — frozen consumption holds; the producer's row
  (`include:["review"]`) cannot distinguish freezing from selection (F7).
- **B** missing-snapshot at `install.Project`: store removed after resolve → `failed`,
  `source_snapshot_unavailable`, tree digests unchanged, store not re-created — the producer's row
  stops at `snapshot.OpenLocal` (F7).
- **C** spaced + unicode project path with a REAL `install` (script skill): shim written, executes
  (`provider-ok`), `status --check` exit 0 — the producer's row stops at `--dry-run` (F7).

## 8. Late results (appended after the background runs finished)

**H2 — SURVIVED.** `draftsources_semantic_test.go:28` mutated to `for _, c := range cases[1:]` (only
the runner loop; the coverage test untouched): `go test -p 1 -count=1 -v -run
'TestDraftSourcesSemanticCases|TestDraftSourcesSemanticCoverage|TestDraftSourcesCorpusCounts'
./internal/crossconformance/` → **exit 0**, `ok … 413.024s`, **93** `--- PASS:
TestDraftSourcesSemanticCases/` subtests executed for a 94-case corpus and nothing failed.
`TestDraftSourcesCorpusCounts` counts the corpus (94) and `TestDraftSourcesSemanticCoverage` matches
drivers ↔ ids, but no test observes how many rows actually ran, so a runner edit that filters cases
(e.g. a per-OS `continue`) would shrink the matrix silently. The schema runner's counter kills the
same mutation (H1). This is the "case counter that drops a case → killed" check from the review note
and it is not met for the semantic corpus.

**v1 regressions (pre-existing crossconformance tests) — green.** `go test -p 1 -count=1 -v -skip
'Draft' ./internal/crossconformance/` in the probe clone (probe file removed first, tree
`2d400147…`): **exit 0**, `ok … 18.524s`, 21 top-level tests pass, 0 fail, 0 skip; 2 subtests skipped
(`TestCrossAdapterConformance/{normative-suite,project-every-path}/rust`, host-capability "no
operator-approved Cargo descriptor for native target x86_64-apple-darwin" — the same two skips the
hosted ubuntu/windows ledgers show). The guard tests (`TestIntegrationSurfaceStartsNoProcessOutside
TheSharedSeams`, `TestIntegrationProductionSourceImportsNoRepositoryPackage`) pass with the new
files, i.e. the `internal/testcli` seam keeps the package free of `os/exec`.

## 9. Findings

**F1 — 4 of the 5 "semantic bounds" are driven rows that FAIL the draft expectation (product
conformance gaps, not bounds).** The results file lists them under "Explicit bounds (never passing)";
only `capture-mutation` is a real bound. The harness pins each gap signature (the row fails when the
gap is fixed — an xfail mechanism, acceptable) and the code comments state the real behaviour, but
the results file never says that the product FAILS these four corpus cases — a reader of the board
record learns "bounds", not "non-conformant, one security-relevant". By the review note's definition
these are hidden failures, and the headline "94 / 89 driven / 5 bounds" overstates conformance. Corrected tally below. The four gaps (all
verified by me at the production entry, all outside this leaf's scope line):
  1. `attestation-evidence-wrong-name`, `attestation-evidence-wrong-context`: `registry.Matches`
     (`internal/registry/registry.go:297`) is `content == content || (identity && commit)` — the
     record name is never compared and a wrong context hash is accepted when identity+commit match.
     Draft §4 ("Network-Git members may use existing registry evidence only with exact name,
     canonical repository, commit and context hash matching") is not implemented on the draft lane
     (no exact-match layer anywhere in `internal/install`/`registry`/`audit`; grep verified). M4 shows
     the rows would detect a fix. → follow-up product task.
  2. `v2-alias-resolution`: alias substitution is loader-only. `config.ResolveRepositoryEndpoints`
     carries `ResolvedHost/ResolvedPort`, but the executor fetches `sources[index]` parsed from
     `attempt.URL` (`internal/buildrepo/transport.go:502, 931`) and only *records* the resolved host
     as provenance; CLI resolve clones the listed URL. The corpus expects "connect-to-
     mirror.corp.example:8443". The transport-revision-2 alias feature therefore never changes the
     connection in production. → follow-up product task (touches STORY-v58b5y's accepted claims).
  3. `v2-user-insteadof-ignored` (**security-relevant**): `curator project resolve` on a literal
     `git:` source inherits the user's git configuration; a hostile `url.<evil>.insteadOf=<declared>`
     redirects the clone and the lock binds the evil commit (the row asserts exactly this). The
     resolved lane isolates user config (`TestUserConfigIgnoredByResolvedLane`), the `gitops.Clone`
     lane does not. → follow-up product task with priority; until then the draft literal-URL lane is
     not conformant to repository-transport §(user configuration never consulted).
  Corrected semantic tally at the pin (94): **87 driven & passing; 4 driven & failing (F1.1–F1.3);
  1 real bound (capture-mutation); 2 driven & failing + 1 bound of those are inside the AC's 73.**
  Of the 87, 17 external-evidence rows + `external-only-current` are half-driven (repair half at
  `install.Project`, status half by test-side comparison, F4), and `runtime-only-refresh`/
  `build-only-refresh` assert the new package identity but not the cache identity (covered by the
  accepted `internal/install` `TestDraftBuildKeyFollowsThePackageIdentity`).

**F2 — the results file's semantic headline is therefore inaccurate as a conformance claim.** It does
say "no release qualification claim"; R3 in §11 carries the corrected numbers into the record.

**F3 — the two pre-existing marker reader gaps (schema bounds) stay open**:
`invalid-external-missing-substituted` accepted by `marker.Read` (BUG candidate already noted in the
1xs0pj review), `valid.json` refused (spec-corpus quirk). Correctly reported as bounds.

**F4 — external-evidence status half not driven** (18 rows): disclosed by the producer as a design
bound; the CLI status verdict for an external mismatch is never observed from production (the
shared fail-closed row proves only "no source → nonzero, no mutation"). Acceptable as a bound.

**F5 — platform coverage not quantified by the producer** for ubuntu/windows; §4 above is the record.

**F6 — classification nits**: 2 inverted-oracle lock positives counted as driven (§5);
`testcli.RequireGit` skips with the undeclared reason "git is not available" (cannot fire on hosted
runners; would be fatal in the platform-case gate if it did — acceptable, but declare or `t.Fatal`).

**F7 — test-strength gaps (behaviour verified by my probes, rows weaker than they claim)**:
`frozen-membership` (vacuous include), `missing-snapshot` (OpenLocal level, "live untouched" is
unfalsifiable there), `CLIProjectPathWithSpaceAndUnicode` (dry-run only),
`TestDraftSourcesCLIInstallRestoresPriorState` — its refusal is `prepare the install transaction:
mkdir …/.codex/skills/.curator-txn-….desired: permission denied`, i.e. the transaction fails at
*prepare* before any replacement is staged, so the "half-replaced generation rolls back" narrative
is not what runs and the prior-state assertion is near-vacuous; real rollback proof remains the
Go-API hook row (`write-boundary-retarget`) and the accepted install-package per-class sweep. The
semantic runner has no executed-count ratio (schema runner has one) — see H2 in §8.

**F8 — Windows cost** (§4): 41 min for the package, dominated by the serial external rows.

**F9 — diagnostic masking observed** (product, minor): install-level external refusals report
`build_repository_source_unavailable` for a `build_repository_identity_invalid` plan (logged by the
rows; the true class is pinned at `ValidateTransportPlan`).

**Design note** — `internal/testcli` exists so the crossconformance guard
(`TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams`, no `os/exec` in package files)
stays intact; precedent `internal/testtoolchain`; nothing under `internal`/`cmd` production imports
it (grep verified). Acceptable.

**Interpretation bound** — "compiled CLI" was read by the producer as the compiled `curator`
binary; no local `go-v1` compiled skill command is driven end to end in this leaf (external
`go-repository-v1` is, via `install.Project` + fake builder; local go-v1 stays covered by the accepted
install-package tests).

## 10. What is verified and holds (do not redo in rev4)

Exact tree and gate identity (§1); pin identical to curator-spec 802caee and the a4fcaf0 subset
exact (§2); 115/3/94 executed on this host with 0 skips and on macOS hosted, per-lane numbers (§3–§4);
production entries per group (§5); narrowing production mutants M1–M4, M5c and harness mutant H1
killed (§6); probes A/B/C show the underlying behaviour holds where rows are weak (§7); v1
regressions green (§8); no production/spec/frozen file touched; Windows within budget (§4).

## 11. Required for revision 4 (test/report only; no production change)

R1 `internal/crossconformance/draftsources_semantic_test.go:28-40` — count executed rows and their
   outcome class and assert `executed == len(cases)` (and `== wantSemanticCases`) at the end of
   `TestDraftSourcesSemanticCases`; log one measured ratio line like the schema suite's
   (`semantic cases: <driven-pass> driven, <known-gap> known-gap, <bound> bound, <total> total`).
   Reproduction of the gap: apply the H2 mutation above → today exit 0 with 93 rows; after R1 it
   must fail.
R2 Distinguish "cannot be driven" from "driven, fails the draft expectation": add
   `semanticKnownGap(t, c, reason)` (or equivalent) and use it instead of `semanticBound` at
   `draftsources_semantic_evidence_test.go:490` and `:506` (wrong-name, wrong-context) and
   `draftsources_semantic_v2_test.go` (`driveV2AliasResolution`, `driveV2UserInsteadOfBound`);
   keep the gap-signature assertions exactly as they are (they flip on a fix). Only
   `capture-mutation` stays a bound.
R3 `TASK-260910-1xya7x_results.md` (rev4 section): replace the semantic headline with the corrected
   tally — 94 pinned = 87 driven-pass + 4 driven-fail (known product gaps: attestation-evidence-
   wrong-name, -wrong-context, v2-alias-resolution, v2-user-insteadof-ignored) + 1 bound; AC subset
   73 = 70 + 2 + 1 — under a heading that says "conformance failures / product gaps", with the
   three follow-up items from F1 named; record the per-lane coverage from §4 verbatim (macOS
   115/3/94; ubuntu 115/3/93+1 host-capability skip; windows 115 schema, 0/3 snapshot
   platform-control, 45/94 semantic + 49 platform-control skips, 2 e2e skips; rose-air unverified)
   and the Windows package time (2447 s, semantic 2275 s); note the two inverted-oracle lock
   positives (§5) next to the schema bounds.
R4 `draftsources_transaction_e2e_test.go` (header comment + results wording): state what the row
   proves — a transaction *prepare* refusal (`prepare the install transaction: mkdir
   …/.curator-txn-…desired: permission denied`) leaves the prior generation intact — and point the
   rollback proof at `write-boundary-retarget` (Go-API hook after PointPrepared) and the accepted
   install-package sweep; or make the failure land after prepare. Either is acceptable; a wording fix
   is enough.

Optional (recommended, not acceptance-gating): fold the §7 probe fixtures into the rows
(`frozen-membership` → `include:["*"]`; `missing-snapshot` → `install.Project` after removing the
store; `CLIProjectPathWithSpaceAndUnicode` → real `install` + shim execution + `status --check`);
give `testcli.RequireGit` a declared reason or `t.Fatal`; consider `t.Parallel()` for the external
rows (F8). Evidence for rev4: the usual narrow macOS run (`go vet`, `-run 'Draft'` in the background,
gofmt) with exit codes; Windows via the hosted gate only.

## 12. Follow-ups for the orchestrator (outside this leaf; open regardless of rev4)

1. **Security**: draft literal-URL resolve lane honours user git config (`url.insteadOf`) —
   `cmd/curator project resolve` / `gitops.Clone` path; the resolved lane already isolates
   (`TestUserConfigIgnoredByResolvedLane`). Corpus `v2-user-insteadof-ignored`; row in
   `draftsources_semantic_v2_test.go` reproduces (lock binds the evil commit).
2. Draft §4 exact evidence matching (name + canonical repository + commit + context) is not
   implemented; `registry.Matches` OR-matching accepts wrong-name/wrong-context records. Corpus
   `attestation-evidence-wrong-name`, `-wrong-context`; M4 shows the rows detect a fix.
3. Transport revision 2 alias substitution never reaches the connection (`buildrepo/transport.go`
   fetches `attempt.URL`; resolved host recorded as provenance only). Corpus `v2-alias-resolution`.
4. Marker reader admits an external `go-repository-v1` build record without `substituted`
   (schema bound `invalid-external-missing-substituted`; BUG candidate from the 1xs0pj review).
5. Minor: install-level external refusal masks `build_repository_identity_invalid` as
   `build_repository_source_unavailable` (F9).

## Evidence bundle

`TASK-260910-1xya7x_review-rev3-evidence.tar.gz`: `logs/` (full.status, draft.out summary, probes.out,
mutants.status, mutant5b.status, mutant-M5c, hmutants.status, h2.status, v1.status, per-lane
counts), driver scripts, `zz_review_probes_test.go`, gate ledger extracts.
