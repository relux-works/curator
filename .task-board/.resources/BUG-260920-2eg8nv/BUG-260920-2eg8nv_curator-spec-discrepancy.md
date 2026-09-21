# Discrepancy report for curator-spec — install-marker-v5 `valid.json`

For the orchestrator to raise on curator-spec (schema-corpus bug, not a
reader bug).

## Fixture

- Corpus pin: `802caee548ddc8b19408746d26c7972d39b39cc2`
- File: `conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json`
- Index entry: `{"schema": "install-marker-v5.schema.json",
  "instance": "schema-cases/install-marker-v5/valid.json", "valid": true}`

## Defect

The document carries `"builds": {}` together with a top-level `build_source`
(`{"algorithm": "curator-build-source-v1", "content_sha256": "sha256:bbbb..."}`)
and `"build_roots": ["build"]`.

`protocol/skillfile-sources.md` states the normative rule:

> Top-level `build_source` remains required exactly for active local `go-v1`
> commands and absent otherwise; an external-only build binds its source
> solely per external record.

With empty `builds` there are no active local `go-v1` commands, so the
top-level `build_source` must be absent. The fixture violates the normative
contract of its own schema: it is JSON-Schema-valid (the schema cannot express
the cross-field presence rule) but normatively inconsistent. Every other
`valid-*.json` marker fixture in the same directory pairs `build_source` with
an active local build; `valid.json` is the outlier.

## Consumer impact

The curator marker reader (`marker.Read`, `validBuildState`) enforces the
normative rule and refuses the document (`nil`, never current). The
crossconformance schema row is therefore held as an explicit bound
("normatively inconsistent fixture (empty builds with build_source); reader
refusal is correct") instead of a driven pass. No workaround was added on the
consumer side.

## Probe

- `valid.json` as-is → `marker.Read` returns `nil` (refused).
- `valid.json` with only the top-level `build_source` member removed →
  `marker.Read` admits the document.

The refusal is isolated to that single member.

## Suggested fix (spec side)

Either remove the top-level `build_source` (and, if desired, the now
meaningless `"build_roots": ["build"]`) from `valid.json`, or give the fixture
an active local `go-v1` build so the `build_source` is required. The former is
the minimal change consistent with the fixture's empty-builds shape.
