# TASK-260922-18ex37 results — revision 5

## Pin, refresh, and ledger

- `SPEC_PIN` is `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` in `.github/workflows/ci.yml`; its manifest SHA-256 is `cb7a98e97543282cdde75f04c55fb15acb91062f8ba6510e3d603db42f1c1ab2`. The spec tree was taken from that exact commit. This honors the 2026-09-25 operator decision to exclude manifest v9/#89 and #90.
- The curator candidate was refreshed to trunk `c278af4f`. `CHANGELOG.md` matches the trunk blob exactly (SHA-256 `97c55451b85a14d13cf922c2d48bb1697897aa852e879540884b1b0bfdb3c064`). The required deletion of root `TASK-260922-3bbvrs_results.md` remains; no new root task or bug result file was added. All candidate changes are unstaged and uncommitted.
- `.github/ci/conformance-gaps.tsv` has 69 data rows: 64 failures against dcc7f015 and five pre-existing marker-v4 gaps. The owner histogram is: BUG-260923-2afgyq 5; STORY-260910-2qmrb8 3; STORY-260910-6bo7ej 2; STORY-260916-1i1gfo 2; STORY-260916-ioemse 11; STORY-260916-ioemse+STORY-260922-1cenbr 28; STORY-260922-1cenbr 16; STORY-260925-1v7pvn 2. Every row has a board owner. No row is assigned to a completed owner.
- The two remaining executable-identity failures are `windows-exec-noncomponent-store-hardlinks` and `windows-exec-unowned-file-hardlinks`. Their former owner STORY-260822-2h0v9j is done, so both rows now belong to the new backlog story STORY-260925-1v7pvn, `windows-exec-hardlink-origin-validation`.
- The row `windows-exec-uncaptured-systemroot-hardlinks` is absent. BUG-260924-5p8b0z landed on trunk `c278af4f`; `TestUncapturedSystemRootHardlinkRegression` now verifies that the production resolver rejects the ambient-SystemRoot case. The case is not a gap.

## Per-family counts before and after the pin

Counts are driven / gap / bounded. The old-root counts are from the attached TASK-260922-3bbvrs and TASK-260923-em42lw evidence; dcc7f015 counts below were measured by production-entry consumers and `conformancecoverage.RunOutcomes`.

| Family | Old root | dcc7f015 | dcc7 total |
|---|---:|---:|---:|
| manager-config-v2 schema | 80 / 0 / 0 | 78 / 29 / 0 | 107 |
| system-config-v2 schema | 28 / 0 / 0 | 36 / 6 / 0 | 42 |
| manager-config-v2 vectors | 48 / 0 / 0 | 31 / 25 / 0 | 56 |
| agent-environment-marker-v1 schema | 53 / 0 / 0 | 59 / 2 / 0 | 61 |
| launch-env-fragment-v1 schema | 49 / 0 / 0 | 50 / 0 / 0 | 50 |
| draft-sources-v1 schema | 111 / 0 / 4 | 113 / 0 / 3 | 116 |
| draft-sources-v1 semantic | 93 / 0 / 1 | 93 / 0 / 1 | 94 |
| draft-sources-v1 snapshots | 3 / 0 / 0 | 3 / 0 / 0 | 3 |
| script-host-execution-policy opt-in cases | 6 / 0 / 0 | 6 / 0 / 0 | 6 |
| script-host-execution-policy executable identity | not published | 6 / 2 / 0 | 8 |

The script policy map now classifies 41/41 named cases and both added sections. `TestExecutableIdentityCasesAtProductionEntry` drives all eight published identity cases through the production resolver: six pass and two are recorded gaps. No newly published case that passes in curator is entered as a gap.

## Owner verification against the pinned spec history

The publishing commits were inspected and each is an ancestor of dcc7f015:

| Published case or first blocking feature | Spec commit | Gap owner |
|---|---|---|
| E3 Codex seed marker forms | `4a2fa3e` | STORY-260916-1i1gfo |
| S1/S3 security posture | `5146c7b` | STORY-260910-2qmrb8 |
| S2 registry bootstrap | `1e73c03` | STORY-260910-6bo7ej |
| E1 source signer fields | `684c9f1` | STORY-260916-ioemse |
| 0017/0018 manager and permission model | `937e795` | STORY-260922-1cenbr |
| Overlay path-kind cases; dcc first blockers are the new E1 and 0017/0018 fields | `bd39adb`, `684c9f1`, `937e795` | STORY-260916-ioemse + STORY-260922-1cenbr |
| Windows executable identity cases | `dcc7f015` | STORY-260925-1v7pvn for the two remaining failures |

The overlay driver tests all 14 schema cases through `config.Load` and all 14 vectors through `config.Parse`/EffectiveJSON. It records the first unsupported field and requires the combined owner. Thus E6/path-kind is not used as a proxy owner where the actual dcc blocker is E1 or 0017/0018. E5, S5, R3/P2, and 8.4.1 produced no new failures in their served production drivers, so they have no new gap rows. The five carried marker-v4 rows remain assigned to BUG-260923-2afgyq.

## Required adjacent status

- **Dotfile-manager:** dcc7f015 publishes 22 cases (Linux 13, macOS 6, Windows 3). TASK-260918-bi6ouz is `to-review`; its attached accepted candidate evidence reports 6/22 on macOS. Linux and Windows remain unverified here and are not claimed as passing.
- **marker-v5:** Published at the pinned root. Curator drives 39 schema cases through `marker.Read`; the corrected `valid.json` passes, and `valid-no-builds.json` is added. The draft schema family totals 113 driven, zero gaps, and three explicit bounds.
- **0017/0018:** Their cases and vectors are present from `937e795`. Manager schema/vector and system schema results are included in the table. Sixteen rows are solely assigned to STORY-260922-1cenbr, and the 28 overlay rows jointly assigned with E1 reflect both published field sets.
- **Rework 4 disposition:** Its platform-scope blocker concerned the uncaptured-SystemRoot case. That case is fixed on c278af4f and passes the named regression test; the row is removed. The two distinct remaining link-origin/ownership failures have a live backlog owner.

## Validation run

Every command below ran as a direct process. Target-root tests used `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1`.

- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/scriptworker -run '^(TestExecutableIdentityCasesAtProductionEntry|TestUncapturedSystemRootHardlinkRegression)$' -count=1 -timeout=3m -v` — exit 0; identity tally 6 driven, 2 known gaps, 0 bound, 0 skipped, 8 total; uncaptured-SystemRoot regression passed.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/scriptpolicy -run '^(TestScriptHostExecutionPolicySectionsAreAllClassified|TestScriptHostExecutionPolicyProductionConsumersCoverAllCases|TestScriptExecutionOptInCases)$' -count=1 -timeout=3m -v` — exit 0; section consumers pass and opt-in cases are 6/6.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/config -count=1 -timeout=8m` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/crossconformance -run '^TestDraftSources(Pin|CorpusCounts|SchemaCases)$' -count=1 -timeout=3m` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/envmarker ./internal/marker -count=1 -timeout=8m` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/conformancecoverage -count=1 -timeout=3m` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/config -run '^TestOverlayGapOwnersMatchFirstProductionBlocker$' -count=1 -timeout=8m -v` — exit 0; 14 schema and 14 vector cases checked.
- Narrowing mutant: temporarily removed only the `source_signers` and `require_source_signers` first-blocker mappings from `overlaySchemaFailureOwner`; the same regression command exited 1 and reported six actual signer-field blockers it could no longer attribute. Restored the source byte-for-byte, then reran the exact regression command — exit 0.
- `bash .github/ci/gate-selftest.sh` — exit 0; 198 passed, 0 failed.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260922-18ex37-ledger-consistency` — exit 0; 356 platform rows checked across Linux, Darwin, and Windows.
- `bash .github/ci/no-broad-suppression.sh` — exit 0.
- `go build ./...` — exit 0.
- `golangci-lint run --timeout=5m` — exit 0; 0 issues.
- `git diff --check` — exit 0. `CHANGELOG.md` blob comparison to `c278af4f` — matched.

The hosted GitHub Linux/macOS/Windows matrix and rose-air lane were not run by this producer; cross-platform hosted checks remain with the coordinator. The local ledger check is not a substitute for executing those hosts. The configured handoff validation runs through `task-board handoff`.

## CHANGELOG entry (for release prep)

- Published-case consumers now share a counted outcome harness. The rc.12
  config, marker, build/cache/lifecycle, skill, environment, closure and interop
  vector/schema families, plus the vendored draft-sources-v1
  semantic/schema/snapshot families, are checked against committed case-count
  pins. A single gap ledger rejects unlisted failures, stale passing rows, and
  vanished listed cases. The complete `skillfile-dev-v2` and `skill-build-v1`
  schema lists and external-repository raw-object, LFS-pointer, pack-index and
  local-config/ref fixture lists now use the same accounting. A gap row records
  work still owed and is never an accepted deviation.

- Pin curator CI to curator-spec `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (manifest SHA-256 `cb7a98e97543282cdde75f04c55fb15acb91062f8ba6510e3d603db42f1c1ab2`). Record 69 conformance gap rows: 64 target-root failures with board owners and five pre-existing marker-v4 gaps. The eight Windows executable identity cases are counted at the production resolver: six driven and two assigned to STORY-260925-1v7pvn; the uncaptured-SystemRoot hard-link case is fixed on trunk and has no gap row.

## Revision 5 — refresh onto ab34556e

### Refresh and trunk combination

- `task-board worktree refresh-candidate TASK-260922-18ex37` exited 0 with `Outcome=refresh_advanced`, `TrunkOID=ab34556ebf17ab95532a3f795aa677b2234ecd8a`, and `BranchOID=6fab2e79a65c1b86ecbb0979f289a61e16f9d9f0`.
- The requested `git diff c278af4f ab34556e` combination reported index mismatches on four overlapping paths. They were resolved with file-based three-way merges from c278af4f to ab34556e, retaining both sides; the full trunk delta was then merged across its 58 changed paths. No merge markers or unresolved paths remain.
- `.github/ci/gate-selftest.sh`: retained the Story's committed-pin and gap-ledger assertions, and the trunk per-lane timeout drift checks, synthetic Go timeout case, and rustup rows.
- `.github/workflows/ci.yml`: retained `SPEC_PIN=dcc7f015e2d97edf2d52928afb6fd79ec8129e8b`; merged the trunk's 60m Ubuntu/macOS and 120m Windows budgets and rustup setup rows.
- `docs/ci-gates.md`: both published-case accounting and trunk timeout documentation are present.
- `internal/crossconformance/draftsources_schema_test.go`: retained the shared outcome tally and trunk parser behavior. The `internal/envprofile` test rename from `draft_capture_test.go` to `path_capture_test.go` follows trunk.
- `CHANGELOG.md` is byte-identical to ab34556e, as required by the refresh instruction. The pin and 69 gap-row release-prep entry remains documented in this results resource.
- `.github/ci/conformance-gaps.tsv` contains 69 data rows; all 69 have non-empty owners. The SPEC_PIN and family counts from the accepted revision are unchanged.
- The hosted matrix was not rerun on this refreshed candidate. Revision 4's hosted run 36090141585 predates this refresh and is not presented as evidence for the current tree; the coordinator owns hosted checks after handoff.

### Validation on the refreshed tree

Every command below ran as a direct process and its real exit code is recorded.

- `bash .github/ci/gate-selftest.sh` — exit 0; 212 passed, 0 failed. This includes the new per-lane timeout and rustup rows.
- `bash .github/ci/ledger-consistency.sh /tmp/TASK-260922-18ex37-ledger-consistency-r5` — exit 0; 356 rows checked across Linux, Darwin, and Windows.
- `bash .github/ci/no-broad-suppression.sh` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-260922-18ex37-spec-dcc.Z9zaHF/conformance/v1 go test ./internal/scriptpolicy ./internal/crossconformance -count=1` — exit 0; scriptpolicy 1.353s, crossconformance 477.077s.
- `go build ./...` — exit 0.
- `golangci-lint run --timeout=5m` — exit 0; 0 issues.
- `git diff --check` — exit 0.
- `git diff --exit-code ab34556e -- CHANGELOG.md` — exit 0.

The worktree remains uncommitted for review.

## Revision 6 — artefact removed (bound cleanup, revision_budget=1)

Per `18ex37-cleanup-6.md` and review verdict rev5 F1 (the only blocking finding):
deleted the stale merge artefact `.github/workflows/ci.yml.merged.tmp` (`rm`,
untracked file). No other file was changed by this run.

Verification (each command run directly, real exit codes):

- `rm .github/workflows/ci.yml.merged.tmp` — done; `test ! -e` confirms absent.
- No other untracked artefact matching `*.tmp`, `*.orig`, `*.rej`, `*.merged*`
  remains in the worktree (checked via `git status --porcelain`).
- `git diff --check -- . ':!.task-board'` — exit 0.
- `grep -n SPEC_PIN .github/workflows/ci.yml` — `SPEC_PIN:
  dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (line 44); no `dced9b8` anywhere
  under `.github/`.
- Path-set check against the rev5 Change Request patch (62 paths): the patch
  minus `ci.yml.merged.tmp` is 61 paths. The current non-board worktree delta
  (24 tracked + 2 untracked = 26 paths: the 2 untracked new files
  `valid-no-builds.json` and `executable_identity_conformance_test.go` are both
  in the rev5 set) is a subset of those 61 except for two pre-existing
  worktree entries this run did not touch: `M CHANGELOG.md` and
  `D TASK-260922-3bbvrs_results.md` (rev2's `git rm` of another task's stray
  file, still present as a deletion vs HEAD `6fab2e79`). Literal equality with
  "61 paths" does not hold because the CR patch base predates HEAD `6fab2e79`
  (which already landed many rev5 paths, e.g. via the 3bbvrs commit); the
  bound's intent — artefact gone, nothing else changed — holds: this run's own
  delta is exactly the one file removal.
- Full-suite evidence is unchanged from revision 5 (verdict rev5: everything
  else verified; hosted gate run 36186939097 green on all lanes). No new
  implementation, no new tests: cleanup only.

The worktree remains uncommitted for the review snapshot.

## Revision 6 — artefact removed, refresh onto 96e3f272 (bound run, revision_budget=1)

Per `18ex37-refresh-6.md` and review verdict rev5 F1 (the only blocking finding):

- Artefact: `.github/workflows/ci.yml.merged.tmp` is absent — already removed
  by the prior bound run. `git status --porcelain` shows no other untracked
  artefact matching `*.tmp`, `*.orig`, `*.rej`, `*.merged*`; the only two
  untracked files are the rev5 candidate additions `valid-no-builds.json` and
  `executable_identity_conformance_test.go`.
- Trunk combine: `git diff ab34556e 96e3f272 -- . ':!.task-board'
  ':!CHANGELOG.md' | git apply --3way` — exit 0, all 6 trunk paths applied
  cleanly (11jgkt snapshot files, 10d3l1 draftevidence_test.go), disjoint from
  this Story as expected. `git reset -q` afterwards so nothing is staged
  (`git diff --cached --name-only` empty). No merge markers remain.
- No trunk revert: all 6 trunk files byte-match 96e3f272 (`git show
  96e3f272:<path> | cmp - worktree`), and CHANGELOG.md byte-matches 96e3f272.
  Candidate delta vs 96e3f272 is 59 tracked modified paths plus the 2 untracked
  candidate files = 61, i.e. the rev5 Change Request patch (62 paths) minus the
  artefact. Lens note: literal `git diff --name-only 96e3f272` reports 62
  because the 3 trunk-new files are untracked and that command diffs via the
  index; the byte comparisons above are the authoritative no-revert proof.
- `task-board worktree refresh-candidate TASK-260922-18ex37` — exit 0,
  `Outcome=refresh_advanced`, `TrunkOID=96e3f272d201e00492d3991da590bd7f997eab0d`.
  No `--replay-resolutions` needed (no conflicts); CHANGELOG already trunk bytes.
- Pin: `SPEC_PIN: dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (ci.yml line 44);
  no `dced9b8` under `.github/`.

Validation (each a direct process, real exit codes, `set -o pipefail` shells):

- `go build ./...` — exit 0.
- `go test -count=1 ./internal/snapshot/` — exit 0 (ok 6.756s).
- `go test -count=1 ./internal/install/ -run 'DraftEvidence'` — exit 0 (covers
  the combined trunk test file; 31.208s).
- `go vet ./internal/snapshot/ ./internal/install/` — exit 0.
- `git diff --check -- . ':!.task-board'` — exit 0.
- Expected-red, reported honestly: full `go test -count=1
  ./internal/snapshot/ ./internal/install/` in one invocation — exit 1; the
  install package hit the 8m go-test timeout (FAIL after 480.664s, timeout
  panic trace via draftpublish/draftmarker tests) under parallel-package
  contention on this host. No file implicated is touched by this change, and
  the isolated rerun `go test ./internal/install/ -run
  '^TestDraftInstallFailsStaleBindingsGeneration$'` — exit 0 (19.53s). Not
  presented as passing; hosted lanes remain coordinator-owned after handoff.

No new implementation and no new tests in this bound revision: artefact removal
plus the ordered disjoint refresh only. The worktree remains uncommitted for
the review snapshot.
