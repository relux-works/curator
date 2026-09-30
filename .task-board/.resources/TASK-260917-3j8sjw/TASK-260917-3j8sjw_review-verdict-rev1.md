# TASK-260917-3j8sjw — review verdict rev1: ACCEPTED

Reviewer run RUN-260930-e60ef5. CR-TASK-260917-3j8sjw-1 revision 1, base
`db32e7fd2f8446ba0d186e356290699c4a7d8cdd`, candidate tree
`13dd55e598ea81ee32bddf061ba215b0f5935d85`, 9 paths, +247/-11.

Normative source: curator-spec `b1a2efb6fa28d014968a2a8fd7641823b5f3cf28`
(PR #116), driven from a disposable clone at that exact commit.

## Gates I ran myself (real exit codes)

| command | exit |
|---|---|
| `pytest -q` with `CURATOR_CONFORMANCE_ROOT=<clone>/conformance/v1` | **0** — 240 passed |
| `mypy` (strict, `files = src/csk_registry`) | **0** — no issues in 15 source files |
| `pytest -k "high_water or upstream or rollback or inconsistent or checkpoint or boundary or idempot or backup or restore or recovery"` | **0** — 59 passed, 181 deselected |

Python 3.12.x venv, `pip install -e ".[dev]"` exit 0. The conformance suite is
`skipif` on `CURATOR_CONFORMANCE_ROOT`; every run above had it set, so the
spec-driven tests executed rather than skipped.

## 1. v1 stays v1, byte for byte — PROVEN, not asserted

I published the same v1 records under base `src` and candidate `src` with a
fixed Ed25519 seed and a fixed `created_at`, through the real
`_countersign` → `Store.append` path, then compared canonical bytes, entry
hashes and the derived snapshot:

- `schema_version: 1` explicit → canonical bytes, `entry_hash`, `seq` identical
- `schema_version` **absent** → identical (still normalized to 1 by
  `_countersign`, exactly as at base)
- snapshot `head`, `merkle_root`, `log_size`, `schema_version` identical

`diff` of the two dumps is empty. No history rewrite, no chain break. The DB
`_SCHEMA_VERSION` stays 4 (`store.py:82`) — no migration, matching the
deployment note. `validate_record` forbids `hash_version` on a schema-1 record,
so no new field can appear on frozen records.

All frozen v1 spec shapes are byte-identical between the old pin `47c3c8cb` and
`b1a2efb` (`registry-snapshot-v1`, `audit-record-v1`, `registry-log-entry-v1`,
`registry-bundle-v1`, `log-response-v2`, `records-response-v2`), so the pin move
cannot have shifted v1 expectations.

## 2. The gates refuse, and the refusal is proven by narrowing

Publication is gated at **one production choke point**: `validate_hash_version`
is called from `protocol.validate_record` (`protocol.py:190`, the POST
`/v1/records` and `import_bundle` path) *and* from `Store._append_locked`
(`store.py:894`), which is the single funnel for all four write paths —
`append`, `append_idempotent`, `append_imports`, `append_upstream_import`
(`store.py:873, 1158, 1339, 1425`). No append bypasses it.

Refusals verified: 400 `invalid_record` with diagnostic `hash_version_mismatch`,
and `log_size == 0` afterwards, i.e. nothing committed
(`tests/test_hash_version.py:56-70`). Matching refuses both directions, with and
without source identity in the query (`tests/test_hash_version.py:24-38`). Bundle
import refuses a v2 record inside a frozen schema-1 bundle and leaves the log
empty (`tests/test_hash_version.py:97-108`).

**Mutation matrix — 16 mutants, 14 killed, and 8 of the 10 kills are
*narrowing*, not deletion.** Each row is a full suite run.

| mutant | shape | exit | result |
|---|---|---|---|
| M1 records_page version WHERE clause removed | delete | 1 | killed (3 failed) |
| M2 WHERE always compares v1 | narrow | 1 | killed (5 failed) |
| M3 `hash_version` dropped from PARTITION BY | narrow | 1 | killed (1 failed) |
| M4 `validate_hash_version` → no-op | delete | 1 | killed (10 failed) |
| M5 gate keeps only the v1-carries-hash_version half | narrow | 1 | killed (6 failed) |
| M6 bundle schema-1-cannot-carry-v2 check removed | delete | 1 | killed (1 failed) |
| M7 export always emits bundle schema 1 | narrow | 1 | killed (3 failed) |
| M8 `_append_locked` gate removed | delete | 1 | killed (7 failed) |
| M9 query gate stops requiring `content_sha256` | narrow | 1 | killed (1 failed) |
| M10 selected version pinned to 1 | narrow | 1 | killed (3 failed) |
| M13 `_countersign` reverts to hardcoded `schema_version = 1` | narrow | 1 | killed (4 failed) |
| M15 query gate also admits `hash_version=3` | narrow | 1 | killed (1 failed) |
| M11 `hash_version` dropped from ORDER BY | narrow | 0 | **survived** — see bounds |
| M12 record `schema_version: 3` admitted | narrow | 0 | **survived** — pre-existing |
| M14 bundle `schema_version: true` admitted as 1 | narrow | 0 | **survived** — see bounds |
| M16 `records_page` argument guard removed | delete | 0 | **survived** — see bounds |

The brief's required mutant (version not compared in matching) is M1/M2/M3: all
killed. M2 is the stronger form — the comparison stays but always selects v1 —
and 5 tests fail.

P4 high-water and the R-series are intact: the 59-test targeted subset passes
(exit 0) and no high-water, checkpoint, boundary, idempotency, backup, restore
or recovery test needed changing in the diff.

## 3. Spec vectors: which apply, which do not

Spec PR #116 added **11** registry-service-relevant schema cases. This CR drives
**10 of 11** (measured from `schema-cases/index.json` at b1a2efb, not estimated):

driven by `test_hash_version_schema_vectors` — `audit-record-v1/invalid-hash-version-on-frozen-record`,
`audit-record-v2/{valid,invalid,invalid-v1-hash-version}`,
`registry-log-entry-v2/{valid,invalid}`, `registry-bundle-v2/{valid,invalid}`,
`log-response-v3/{valid,invalid}`. The two `audit-record-*` groups are driven
twice — once against the pinned JSON Schema and once through the real
`validate_record`, so schema and implementation are pinned to agree.

**Not driven:** `registry-log-entry-v1/invalid-v2-record-in-frozen-entry.json`
(1 of 11). The property is still exercised, and more strongly, through the live
service: `test_v2_service_outputs_match_pinned_shapes` asserts a real `/v1/log`
response carrying a v2 record is **invalid** against `log-response-v2`, whose
items `$ref` `registry-log-entry-v1`. Stated as a bound, not a defect.

**Do not apply to this service:**
- `vectors/content-hashes-v2.json` — curator-side tree framing. The registry
  never recomputes a content hash; both framings serialize as
  `sha256:<64 hex>` (spec `common.schema.json#/$defs/sha256`), so the registry
  cannot infer framing from digest bytes. This is why "declared version
  disagrees with its hash form" is correctly implemented as a declaration
  check, and the deployment note says so explicitly.
- `install-marker-v5`, `context-lock-v2`, `agent-environment-marker-v2/v3`,
  `manager-config-v3`, `launch-env-fragment-v2` hash-version cases — client and
  manager surfaces, no registry code path.
- `vectors/registry-service.json` `artifact_key` / `sort_key` — both changed to
  include `hash_version` at this pin, and neither was driven before this CR nor
  is driven now (the repo drives the file's `query_cases`, `idempotency_cases`,
  `transaction_cases`, `recovery_cases`, `checkpoint_cases`, `transport_cases`,
  `cache_cases`, `pagination`, `limits`, `records`, `snapshot`). `artifact_key`
  is covered behaviourally instead — M3 kills on the PARTITION BY. `sort_key`
  is the M11 bound below. The file's `query_cases` contain **no** version cases,
  so no spec query vector was left unimplemented.

## 4. Deployment note

`docs/content-hash-versions.md` — I checked all 16 factual claims against the
pinned spec and the code; every one holds. Notably correct and non-obvious:
that `registry-snapshot-v1` is frozen with `additionalProperties: false` and so
*cannot* carry a top-level `hash_version`, and that a snapshot instead commits
the versioned log through `head`/`merkle_root` because one log may hold both
framings. That is the right reading of the DoD's "snapshots carry hash_version",
not a shortcut around it.

CHANGELOG entry present under `## Unreleased` → `### Added`, linking the note.
No stray files (`git status` is exactly the 9 CR paths). No employer name
anywhere in the added text — I enumerated every capitalized token in the added
lines; the only proper noun is `Curator`. Nothing outside the version feature
changes behaviour: the one shared-code edit is `_countersign`, proven
byte-identical for v1 in §1.

## Stated bounds and non-blocking observations

Not defects. None gates acceptance; each is a fact a later reader should have.

1. **M11 — `ORDER BY ... hash_version ...` (`store.py:1236`) is unpinned.**
   Removing the term breaks no test. I tried to turn this into a real failure:
   I appended v2 *before* v1 for one artifact key and paged with `limit=1`.
   As-shipped returns `[1, 2]`; the mutant *also* returns `[1, 2]`, because the
   window's `PARTITION BY` already includes the version and SQLite emits
   partitions in that order. So the explicit term is currently redundant and
   ordering rests on an unspecified SQLite tie-break rather than on the clause.
   **I could not demonstrate a wrong output, so I report this as unknown, not
   as a bug.** Worth a follow-up test if `sort_key` is ever driven from the
   vector.
2. **M16 — the new `records_page` argument guard (`store.py:1207`) has no
   negative test.** Deleting it breaks nothing: every app call pre-validates,
   and `test_invalid_version_query` covers the same condition at the HTTP entry
   point. Defense-in-depth for direct store callers, unexercised.
3. **M12 — record `schema_version: 3` rejection is untested.** Pre-existing:
   base had no such test either (`git show <base>:tests/*` has no
   schema_version-rejection case), so this CR neither introduced nor widened
   the gap.
4. **M14 — `bundle.py:132` uses `type(...) is not int`, but a bundle
   `schema_version: true` is still admitted as 1** and has no test. The
   record-level bool rejection *is* tested (`(2, True)` case), so the coverage
   is inconsistent rather than absent.
5. **New SQLite JSON1 dependency.** `json_extract` is used for the first time in
   this repo (0 occurrences at base). Safe on every supported target — SQLite
   ships JSON functions unconditionally since 3.38, `python:3.12-slim` is well
   past that, and CI is green on ubuntu/macos/windows × 3.11/3.14 — but it is a
   new runtime requirement the deployment note does not mention.
6. **Style drift in `tests/test_hash_version.py`.** 56 single quotes and 0
   double quotes against a uniformly double-quoted suite; 0 comments or
   docstrings where the surrounding tests document each case's intent; no type
   annotations where `test_registry.py` / `test_protocol_conformance.py` use
   `tmp_path: Path` and `-> None`. Also a new cross-test-module import chain
   (`test_protocol_conformance` → `test_hash_version` → `test_registry`
   privates) with no precedent at base.
7. **Docs left slightly behind.** The README endpoint bullet for
   `GET /v1/records` lists `limit`/`cursor` but not `hash_version`, and the
   older conformance tests still assert `/v1/log` against `log-response-v2`
   although the spec endpoint table at this pin names `log-response-v3`. Both
   are still *true* for v1-only content, so neither is wrong today.

## Verdict

**Accepted.** Every spec-mandated gate — equal-version matching, publication
refusal, bundle-import validation — is implemented at a real production call
site, carries negative tests, and is proven by narrowing mutants rather than
only by deletion. v1 immutability is proven byte-for-byte against base rather
than asserted. 10 of 11 applicable new spec cases are driven from the pinned
clone, with the one omission named and independently covered. pytest exit 0,
mypy exit 0.
