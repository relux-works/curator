# BUG-260921-3cgij4 — review verdict, Change Request CR-BUG-260921-3cgij4-2 (revision 2)

Reviewer run: RUN-260921-670be2 (claude-opus-5/max). Repository: curator-spec.
Base OID `802caee548ddc8b19408746d26c7972d39b39cc2`; candidate tree OID
`d7d83c7f81cfe2a1787fc8af246b779a17205e47`; patch sha256
`9201386c92c5beb536adc42a48cd7b0d779f8452ecb272aafbb0817b39f3e09f` (recomputed, matches).
Story worktree blobs for the three changed paths match the candidate tree exactly
(index.json `effd940`, valid.json `cda43b9`, valid-no-builds.json `47e0c4c`).

## Verdict: ACCEPT

Exactly one verdict branch: `accept_cr(BUG-260921-3cgij4, revision=2, evidence=BUG-260921-3cgij4_review-verdict-rev2.md)`.

## What was judged (per 3cgij4-review-rev2-note.md)

1. **Normative rule untouched; both fixtures satisfy it.** `git diff --stat base..candidate -- protocol schemas docs`
   is empty. The rule at `protocol/skillfile-sources.md:268` ("Top-level `build_source` remains required exactly
   for active local `go-v1` commands and absent otherwise") is cited by exact line number in both index comments
   (verified: line 268 in the candidate tree).
   - `valid.json` (present branch): one `builds.local-helper` entry, byte-identical to the local `go-v1` record in
     the pre-existing accepted fixture `valid-local-mixed-builds.json`; closed v5 shape (`driver: go-v1`,
     `receipt_schema_version: 3`, `execution_policy: manager-worker-v1`, `cache_key`, `receipt_sha256`,
     `artifact_sha256`, `artifact_path: bin/local-helper`, no extra members); `local-helper` ∈ `commands`;
     `build_roots` non-empty; top-level `build_source` therefore legitimately required and present.
   - `valid-no-builds.json` (absent branch): `builds: {}` and no top-level `build_source`. `builds` is
     schema-required (`required` list of install-marker-v5.schema.json), so "absent" is correctly encoded as the
     empty object. The two fixtures differ in exactly the two rule-bearing members (diffed after key-sorting) —
     a minimal pair.
2. **Index consistency.** 115 → 116 rows (positives 24 → 25, negatives 91 unchanged); no duplicate instances;
   every indexed instance exists; every file under `schema-cases/` is indexed; the new row sits directly after
   `valid.json`. `comment` is a new key for this index (0 uses at base, 2 now); the consumers ignore it — the
   README specification command reads only `schema`/`instance`/`valid`, and curator's `loadDraftIndex` uses plain
   `json.Unmarshal` (no `DisallowUnknownFields`).
3. **Negative fixture for the exact defect — infeasible in this corpus's pattern; recorded as a bound.**
   Every `index.json` negative is a JSON-Schema refusal and the README specification command asserts
   `Draft202012Validator(...).is_valid(doc) == case['valid']` per row. Measured, not argued: mutant G6 (the
   exact-defect document — empty `builds` plus top-level `build_source` — marked `valid: true`) is GREEN under
   that gate; mutant G7 (the same document marked `valid: false`) is RED (`AssertionError` on that row). So the
   cross-field rule is not schema-expressible and cannot be pinned as an `index.json` negative without changing
   the gate. `semantic-cases.json` is a labelled id/input/expected corpus (no marker documents); it carries the
   positive `external-only-current` case with `top_level_build_source: absent` but no refusal case for either
   side of the top-level rule. Adding one there would be a different mechanism (a curator semantic driver plus
   `wantACSemantic`/ledger changes) and is outside this leaf's scope (schema-cases fixtures + index); named
   below as an optional follow-up, not required. The negative pin lives at the consumer: curator
   `internal/marker.validBuildState`, exercised below.
4. **Curator follow-up in results.md is precise.** Names resolve in curator @ `50d3ca1`:
   `internal/crossconformance/testdata/draft-sources-v1/DRAFT_SOURCES_PIN` (= `802caee`, i.e. exactly this CR's
   base — the vendored corpus is the pre-fix state), `MANIFEST.sha256`, `wantSchemaCases = 115` in
   `draftsources_corpus_test.go:26`, the `draftSchemaBounds` row for `valid.json` and the `valid.json` arm of
   `driveInstallMarkerV5Case` in `draftsources_schema_test.go`. Expected post-state 113 driven / 3 bounds is
   arithmetically right (116 − 3 local-snapshot bounds). Minor addition for the follow-up: `ac-schema-cases.txt`
   (102 rows, a subset that must exist in the pinned corpus) may optionally gain the new instance; the
   `MANIFEST.sha256` must be regenerated with the corpus bytes.
5. **Nothing else in the patch.** `git diff --stat` = exactly the three paths (59 insertions, 2 deletions).

## Evidence — what I reran myself (disposable materialisation of the candidate tree: `git archive d7d83c7… | tar -x` into `/tmp/review-3cgij4-rev2`; venv `jsonschema==4.25.1`, `referencing 0.37.0`, Python 3.14.6; Go 1.26.0)

| Check | Result |
| --- | --- |
| README specification command (`conformance/draft-sources-v1/README.md`, extracted verbatim, sha256 `65c6bb09…25dc`) | exit 0 — `Marker migration: 25/25; 2/2 build arms; narrowed refusal mutants: 18/18`, `Schema cases: 116/116; negatives: 91/91; wire schemas: 8/8`, `Snapshot byte vectors: 3/3` |
| `python3 tools/validate.py` | exit 0 — `validated 62 schemas and 1119 vector files` (11.3 s) |
| `python3 -B -m unittest discover -s tools -p 'test_*.py'` | exit 0 — `Ran 576 tests in 1188.710s` / `OK` (started 13:09:21Z, ended 13:29:11Z; 0 FAIL/ERROR lines; run as one background job with the log polled, log `/tmp/review-3cgij4-rev2-unittest.log`) |
| `go test -count=1 ./tools/...` | exit 0 — `ok github.com/relux-works/curator-spec/tools/generate-vectors 1.112s` |

Gate-reach proof (narrowing mutants against the README specification command, scratch copies, all restored):

| Mutant | Expected | Observed |
| --- | --- | --- |
| G1 `valid-no-builds.json` row flipped to `valid:false` | red | red (AssertionError on that row) |
| G2 `valid.json` row flipped to `valid:false` | red | red |
| G3 unknown top-level key added to `valid-no-builds.json` | red | red |
| G4 `artifact_path` dropped from the `local-helper` build (closed shape) | red | red |
| G5 extra `commit` member added to the `local-helper` build (closed shape) | red | red |
| G6 exact-defect document (empty builds + build_source) marked `valid:true` | green (schema blind to the rule) | green — `116/116` |
| G7 exact-defect document marked `valid:false` | red (negative fixture infeasible) | red |

Curator pinned-harness control (disposable clone, vendored corpus at `DRAFT_SOURCES_PIN=802caee`, i.e. the pre-fix
corpus): `go test -count=1 -run 'TestDraftSourcesSchemaCases$' ./internal/crossconformance/` → PASS,
`schema cases: 111 driven, 4 bounds, 115 total`, bound row for `valid.json` still active — the harness state the
follow-up starts from.

Consumer production-entry proof (curator @ `50d3ca17505623b44abb418e9bba5c95bdd91623`, disposable clone
`/tmp/review-3cgij4-curator`, probe test in `internal/marker` driving the same entry as
`crossconformance.driveInstallMarkerV5Case`: bytes written as `.csk-install.json`, `marker.Read(dir)`;
`go test -count=1 -run TestReview3cgij4Probe ./internal/marker/` → PASS, 7/7):

| Document | `marker.Read` | Meaning |
| --- | --- | --- |
| new `valid.json` | ACCEPTED | present branch admitted |
| new `valid-no-builds.json` | ACCEPTED | absent branch admitted |
| `valid-local-mixed-builds.json` (control) | ACCEPTED | unchanged control |
| old `valid.json` @ `802caee` | REFUSED | the defect is still refused (reader discriminates) |
| mutant A: new `valid.json` minus top-level `build_source` | REFUSED | local go-v1 present + source absent → refused |
| mutant B: new `valid-no-builds.json` plus top-level `build_source` | REFUSED | the exact defect rebuilt from the sibling → refused |
| mutant C: `local-helper.driver` = `go-repository-v1` (incomplete external record) | REFUSED | closed-shape control |

Accepted from attached evidence without replay: none of the verdict rests on the producer's logs; the rev2
validation log (`BUG-260921-3cgij4_change-request_rev2-validation.log`, exit 0, 576 tests / 1116 s) is
consistent with what I observed.

## Bounds stated

- `make validate` does not reach `conformance/draft-sources-v1`: `tools/validate.py` roots at
  `conformance/v1` and there are 0 references to `draft-sources-v1` under `tools/` or the Makefile. Its green
  proves the released corpus is unaffected; the discriminating gate for this change is the README
  specification command, which was run and mutant-tested above.
- The corpus pins the cross-field rule only positively (one fixture per branch); the refusal side is pinned
  by the consumer reader (curator `validBuildState`), not by curator-spec — see item 3.
- Observation, not a defect and not introduced by this CR: `activation.commands` lists only `golden-tool`
  while `local-helper` carries the go-v1 build; this mirrors the pre-existing accepted
  `valid-local-mixed-builds.json`, and neither the schema nor the reader ties `builds` keys to
  `activation.commands` (the reader requires `builds` keys ⊆ `commands`). Fixtures are structural examples
  with synthetic digests per the README.

## Checklist items closed by this review

Implementation matches AC — yes (all five AC clauses verified above). Solution fits project architecture —
yes (corpus fixed, rule untouched, consumer pin left where it belongs). Tests green — yes (table above).
Verdict branch — accept; no `commit_ack` supplied; integration is the producer-side/orchestrator step.
