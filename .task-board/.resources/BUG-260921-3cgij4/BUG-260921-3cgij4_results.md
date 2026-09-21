# BUG-260921-3cgij4 results — install-marker-v5 valid fixture vs build_source rule

## Rule (unchanged, not weakened)

`protocol/skillfile-sources.md:268` (marker schema 5, §4):

> Top-level `build_source` remains required exactly for active local `go-v1`
> commands and absent otherwise; an external-only build binds its source solely
> per external record.

Same rule in `protocol/core.md` §10 (markers v3/v4): top-level `build_source`
is REQUIRED exactly when at least one active local `go-v1` command exists and
MUST otherwise be absent. JSON Schema cannot express the cross-field rule, so
a document can be schema-valid yet normatively invalid. No normative text was
changed.

## Defect

`conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json`
(index `valid:true`) had `"builds": {}` together with a top-level
`build_source`. With zero builds there are no active local commands, so
`build_source` must be absent; the fixture violated its own contract. The
curator reader (`internal/marker/marker.go` `validBuildState`: empty builds
require `!sourcePresent && m.BuildSource == nil`) refuses it correctly, and
curator keeps the case as a documented bound (BUG-260920-2eg8nv: 111 driven /
4 bounds).

## Fix (per operator decision 2026-09-21)

1. `schema-cases/install-marker-v5/valid.json` — now carries an active local
   `go-v1` build, so its top-level `build_source` is legitimately required.
   Added one `builds` entry for `local-helper`, byte-identical to the `go-v1`
   record in the sibling `valid-local-mixed-builds.json` (closed v5 shape:
   `driver: go-v1`, `receipt_schema_version: 3`,
   `execution_policy: manager-worker-v1`, `cache_key`, `receipt_sha256`,
   `artifact_sha256`, `artifact_path: bin/local-helper`). All other members
   unchanged (`commands` still `[golden-tool, local-helper]`; builds keys ⊆
   commands, which is the reader's requirement — cf. script commands, which
   also produce no build entry per core §10).
2. `schema-cases/install-marker-v5/valid-no-builds.json` (new, `valid:true`) —
   the absent-otherwise branch: `"builds": {}` and NO top-level `build_source`.
   Note: `builds` itself is schema-required (marker v5 `required` retains the
   v2 rule "requires `build_roots` and a `builds` object, including empty
   values"), so the absent branch is encoded as empty-object, not key-absent.
3. `conformance/draft-sources-v1/index.json` — new entry for
   `valid-no-builds.json` (`valid:true`); `comment` fields added to both the
   `valid.json` and `valid-no-builds.json` entries citing
   `protocol/skillfile-sources.md:268` and BUG-260921-3cgij4. Index grows
   115 → 116 cases (positives 24 → 25 for the corpus; negatives unchanged at
   91). Extra `comment` keys are ignored by the README spec command (reads
   only schema/instance/valid) and by curator's `draftSchemaEntry` (unknown
   JSON fields ignored).

## Negative fixture: deliberately NOT added

The brief conditions the negative fixture on "if the corpus has the pattern".
It does not: every `index.json` negative is a JSON-Schema refusal, and the
README specification command asserts `validator.is_valid(...) == case['valid']`
for all 116 rows. A schema-valid but normatively-invalid document
(empty builds + `build_source`) marked `valid:false` would fail that gate by
construction. `semantic-cases.json` covers resolver/filesystem/security
outcomes (id/input/expected), not marker documents — no pattern there either.
The pin for the refusal lives on the curator side instead: `validBuildState`
plus the crossconformance bound row, which this fix converts to a drive (below).

## Before / after

Before (`valid.json`, normatively invalid):

```json
  "build_source": {
    "algorithm": "curator-build-source-v1",
    "content_sha256": "sha256:bbbb...bbbb"
  },
  "builds": {},
```

After (`valid.json`, present branch):

```json
  "build_source": {
    "algorithm": "curator-build-source-v1",
    "content_sha256": "sha256:bbbb...bbbb"
  },
  "builds": {
    "local-helper": {
      "artifact_path": "bin/local-helper",
      "artifact_sha256": "sha256:dddd...dddd",
      "cache_key": "sha256:1111...1111",
      "driver": "go-v1",
      "execution_policy": "manager-worker-v1",
      "receipt_schema_version": 3,
      "receipt_sha256": "sha256:eeee...eeee"
    }
  },
```

New sibling (`valid-no-builds.json`, absent branch): identical to before except
the top-level `build_source` member is removed.

Index delta:

```diff
   {
     "schema": "install-marker-v5.schema.json",
     "instance": "schema-cases/install-marker-v5/valid.json",
-    "valid": true
+    "valid": true,
+    "comment": "Present branch of protocol/skillfile-sources.md:268 ..."
+  },
+  {
+    "schema": "install-marker-v5.schema.json",
+    "instance": "schema-cases/install-marker-v5/valid-no-builds.json",
+    "valid": true,
+    "comment": "Absent branch of protocol/skillfile-sources.md:268 ..."
   },
```

## Evidence (real exit codes, this worktree)

- Draft-sources specification command (`conformance/draft-sources-v1/README.md`,
  run verbatim via `/tmp/draft_spec_check.py`): exit 0 —
  `Marker migration: 25/25 top-level fields; 2/2 build arms; narrowed refusal
  mutants: 18/18 detected`, `Schema cases: 116/116; negatives: 91/91; wire
  schemas: 8/8`, `Snapshot byte vectors: 3/3`.
- `python tools/validate.py` (uv-managed `jsonschema==4.25.1`, per
  `requirements-dev.txt`): exit 0 — `validated 62 schemas and 1119 vector files`.
- `python -B -m unittest discover -s tools -p 'test_*.py'`: exit 0 —
  `Ran 576 tests in 1280.597s / OK`.
- `go test ./tools/...`: exit 0 — `ok .../tools/generate-vectors`.
- Curator production-reader probe (transient `go test` in curator
  `internal/crossconformance`, same entry as `driveInstallMarkerV5Case`: write
  bytes as `.csk-install.json`, call `marker.Read`; probe file removed after):
  exit 0 — `valid.json: ACCEPTED`, `valid-no-builds.json: ACCEPTED`,
  `valid-local-mixed-builds.json: ACCEPTED` (control, unchanged).
- Control that the reader still discriminates: curator's pinned suite
  `TestDraftSourcesSchemaCases` (old corpus, pin `802caee`): exit 0 —
  `schema cases: 111 driven, 4 bounds, 115 total`, i.e. the OLD `valid.json`
  bytes are still refused (`got == nil` asserted by the bound row).

Note: dev dependencies were run via uv-managed ephemeral environments
(`uvx --with 'jsonschema==4.25.1'`) to match `requirements-dev.txt` without
adding untracked files to the worktree; `git status` shows only the three
intended paths.

## Curator follow-up (named, not done here)

After the next curator-spec pin promotion in curator:

1. Bump the vendored draft corpus (`internal/crossconformance/testdata/
   draft-sources-v1`: `DRAFT_SOURCES_PIN`, `MANIFEST.sha256`, corpus bytes) to
   a spec commit containing this fix.
2. Update `wantSchemaCases` 115 → 116 and add the expected driven row for
   `schema-cases/install-marker-v5/valid-no-builds.json`.
3. Convert the bound row for `schema-cases/install-marker-v5/valid.json` to a
   drive: delete the `draftSchemaBounds` entry and the `valid.json` special
   case in `driveInstallMarkerV5Case` (the `got != nil → Fatalf("convert the
   row to a drive")` arm), so the fixed fixture flows through the default
   `(got != nil) != entry.Valid` assertion. Expected post-state: 113 driven /
   3 bounds (remaining bounds: the three local-snapshot no-byte-reader cases).

The curator `SPEC_PIN`/vendored pin is a fixed commit, so this fix takes effect
there only after the pin bump; until then curator's suite stays green on the
old bytes with the documented bound.

## Revision 2 (republish unchanged, 2026-09-21)

revision 2 = revision 1 unchanged; gate rerun after a host exec-stall window.
