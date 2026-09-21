# Review verdict — BUG-260920-2eg8nv revision 1: ACCEPT

Reviewer: claude-opus-5 (RUN-260920-99cb50), 2026-09-20. Read-only review; the Story
worktree was never modified (temp-index `write-tree` before and after the review =
`05b7aee1d0182414a898469c4ec6087318c1966b`). Evidence bundle:
`BUG-260920-2eg8nv_review-rev1-evidence.tar.gz` (probes, drivers, raw logs).

## 1. Exact candidate tree

- CR-BUG-260920-2eg8nv-1, base `9352a87a…`, candidate tree `05b7aee1…`; patch resource
  sha256 `5f1ec50a…` (matches the CR record).
- Disposable clone `/tmp/2eg8nv-review/cand` = base + `git apply --index` patch + commit →
  `HEAD^{tree}` = `05b7aee1…`. Every rerun, probe and mutant below ran there; the clone
  ended every driver at the candidate tree with a clean status (logged).
- Gate: run 35505872964 (success, all 11 required jobs green) headSha `8f424445…`, whose
  parent is the base and whose tree is `05b7aee1…` — the gate tested exactly this candidate.
- Delta = 3 paths (`internal/marker/marker.go`, `internal/marker/marker_v5_builds_test.go`,
  `internal/crossconformance/draftsources_schema_test.go`); checkpointed siblings
  3ukdk4/2wyzde/3v7x6j untouched.

## 2. Schema verification (my own, jsonschema 4.25.1 + referencing, README registry recipe)

`schemacheck.py` / `schemacheck.out`, rc=0, spec pin 802caee:

- marker-v5 corpus vs `install-marker-v5.schema.json`: **38/38 agree with index.json**.
- `invalid-external-missing-substituted.json`: schema-invalid; a schema mutant that removes
  `substituted` from the external arm's `required` admits it → the refusal is exactly the
  absent Boolean (the AC's class).
- `valid.json`: schema-valid; `builds: {}` + top-level `build_source`. The schema has no
  conditional/dependent rule mentioning `build_source` (v4 or v5: none; top-level
  `build_source` is optional). `valid.json` minus `build_source` is also schema-valid. Among
  the 7 `valid*.json` marker fixtures it is the ONLY one whose top-level `build_source`
  presence does not track an active local `go-v1` build. Normative text
  (`protocol/skillfile-sources.md` §4: "Top-level `build_source` remains required exactly for
  active local `go-v1` commands and absent otherwise"; `protocol/core.md` marker v2–v4 rules)
  makes the document normatively inconsistent. The producer's determination is confirmed;
  the reader is right; the attached discrepancy text
  (`BUG-260920-2eg8nv_curator-spec-discrepancy.md`) is accurate (pin, index entry, quoted
  rule, isolation probe, suggested fix) and contains no workaround. Row stays a documented bound.

## 3. Reader corpus ratio and valid.json isolation (probe `TestZZReviewCorpusRatio`, `TestZZReviewValidJSONCause`)

- Base tree: `marker.Read` agrees **36/38** (F1 admitted, valid.json refused).
- Candidate: **37/38** — the only disagreement is `valid.json` (index valid, reader refuses).
- valid.json: as-is → refused; `build_roots: []` alone → still refused; minus top-level
  `build_source` → readable (also with `build_roots: []`, also minus `requirers`). The refusal
  is isolated to the single `build_source` member, as the producer reported.

## 4. Narrow reruns in the clone (bash driver, `set -o pipefail`, rc captured before logging)

| command | rc |
|---|---|
| `gofmt -l internal/marker/ internal/crossconformance/` | 0, no files |
| `go vet ./internal/marker/ ./internal/crossconformance/` | 0 |
| `golangci-lint run ./internal/marker/... ./internal/crossconformance/...` | 0, `0 issues.` |
| `go test -count=1 -p 1 -v ./internal/marker/` (whole package: v1–v4 legacy, v5 identity/schema pins, builds) | 0 — 44 top-level PASS; `TestMarkerV5ExternalBuildClosedShape` 34/34 subtests PASS |
| `go test -count=1 -p 1 -v -run 'TestDraftSourcesSchemaCases$\|TestDraftSourcesPin$\|TestDraftSourcesCorpusCounts$' ./internal/crossconformance/` | 0 — 115/115 subtests PASS, ratio line `schema cases: 111 driven, 4 bounds, 115 total`; `invalid-external-missing-substituted.json` PASS as a DRIVEN row (default `(got != nil) != entry.Valid` branch), `valid.json` PASS as the retained bound |

Executed-count guards intact: `len(entries) == wantSchemaCases` (115), `driven+bound == total`,
every `draftSchemaBounds` key must be present in the index. 4 bounds = 3 local-snapshot + valid.json.

## 5. Hosted gate evidence (artifacts `test-evidence-{ubuntu,windows,macos}-latest`, run 35505872964)

Per lane (`observed-cases.tsv`): `TestMarkerV5ExternalBuildClosedShape/*` 34 subtests, 34 pass;
`…/invalid-external-missing-substituted.json` pass; `…/install-marker-v5/valid.json` pass;
115 schema-case subtests; ratio line `111 driven, 4 bounds, 115 total` on all three OSes
(Windows included — the marker package is pure Go, no POSIX wrapper skip).

## 6. Production-entry probes (probe file `zz_review_2eg8nv_test.go`, every row through `marker.Read` + `marker.Current`)

Same probe on base and candidate (`rows-base.txt` / `rows-cand.txt`):

| row | base | candidate |
|---|---|---|
| missing-substituted (F1) | readable | **refused** |
| substituted-null | readable | **refused** |
| substitution-null | readable | **refused** |
| declared-tag-null | readable | **refused** |
| local-path substitution carrying `ref` | readable | **refused** |
| case-variant keys `Substituted` / `DECLARED_TAG` / `Substitution: null` (Go's decoder matches case-insensitively) | readable | **refused** |
| nested case-variant `effective_identity.KIND`, `substitution.REF`, `substitution.TYPE: null` | readable | **refused** |
| duplicate `substituted` key, trailing data, record-is-array, `receipt_schema_version: 3.0`, `substituted: []` | refused | refused |
| control (intact writer output) | readable | readable |

For every refused document `Current(dir, expected)` returned `(false, nil)` (asserted): `Read == nil`
short-circuits `Current` at `marker.go:1086`; on the status lane a nil read of a supported
schema is classified by `markerRefusal` (`cmd/curator/builds.go:537`) as the typed
`invalid-marker` state. Never current.

## 7. Narrowing mutants (clone, `git checkout --` restore, hash-verified edit and restore, committed tests only)

| mutant | result |
|---|---|
| M1 drop `"substituted"` from `v5ExternalBuildRequired` (the AC mutant) | **killed**: unit `missing-substituted` FAIL; schema row `invalid-external-missing-substituted.json` FAIL (rc=1 both) |
| M2 ignore the call-site result in `validBuildState` | **killed**: 5 unit rows (missing-substituted, substituted-null, substitution-null, declared-tag-null, local-path-with-ref) + schema row |
| M3 drop the per-member JSON type check | **killed**: substituted-null, declared-tag-null |
| M5 drop the `local-path` closed shape | **killed**: local-path-with-ref |
| M8 empty the required loop | **killed**: missing-substituted + schema row |
| M4 drop the raw `substituted == present(substitution)` conditional | survives committed tests — **equivalent**: with the type loop intact the decoded `validDriverBuild` check is the same predicate. Proof: M4 + dropping the decoded check → `substituted-true-missing-substitution` and `substituted-false-with-substitution` FAIL. Not a gap. |
| M6 drop `!allowed` (foreign member refusal) | survives committed tests (the `foreign-field*` rows are refused earlier by `DisallowUnknownFields`); its surviving class is a null-valued case-variant member (`"Substitution": null`), killed by my probe. Semantically inert on decode. Residual, see §8. |
| M7 drop the nested identity closed shape | survives committed tests; killed by my `effective_identity.KIND` probe (case-variant nested key that the decoder would honour). Residual, see §8. |
| M9 drop the `network-git` closed shape | survives committed tests; killed by my `substitution.TYPE: null` probe. Residual, see §8. |

M2 also establishes the exact new reach of the raw check over the decoder + decoded validation:
absent/null `substituted`, null `substitution`, null `declared_tag`, `local-path` with `ref`,
and case-variant member keys — the class the AC names, plus the case-variant class the
candidate covers by construction.

## 8. Residuals (non-blocking; recorded as bounds for follow-up, not defects of this leaf)

1. Case-variant member keys (top-level `Substitution: null`, nested `KIND`, `TYPE: null`) are
   refused by the candidate but not pinned by a committed row; the committed foreign-member
   rows die at the decoder, so M6/M7/M9 survive the suite. Suggest three rows in a later
   hygiene leaf (probe file in the bundle has them).
2. `declared_tag: ""` is admitted on both trees (value grammar; `validDriverBuild` never checks
   the tag pattern). Pre-existing, outside this AC ("value grammars stay the job of the decoded
   validation").
3. The local `go-v1` arm's raw shape is still not closed: `"repository": null`, `"substituted":
   null`, `"substitution": null` on a `go-v1` record decode to zero values and are admitted
   (probe `TestZZReviewLocalArmNullResidual`; unknown members are refused by the decoder). Outside
   this AC (external arm only; "accepted local-arm rows unchanged") — candidate for a follow-up.
4. Producer results.md counts "31 negative rows … 35/35 subtests"; the committed test has 30
   negatives + 4 controls = 34 subtests (local, ubuntu, windows, macos all report 34). Reporting
   slip only.

## 9. Fit and scope

The external-arm check mirrors the accepted package-arm discipline (`validV5PackageShape` /
`closedRawShape` / `rawJSONType`, now with `boolean`/`number`), sits in the existing v5
per-record branch of `validBuildState`, and is v5-only: v1–v4 markers and the v5 local arm take
the same paths as before (whole marker package green, legacy/local rows unchanged). Only the
required loop refuses an absent `substituted`, so the AC mutant is a one-line kill rather than a
survivor behind a redundant refusal. Comments match the package's density and idiom.

## Verdict

ACCEPT — `accept_cr(BUG-260920-2eg8nv, revision=1, evidence=BUG-260920-2eg8nv_review-verdict-rev1.md)`.
The valid.json discrepancy text is ready for the orchestrator to raise on curator-spec.
