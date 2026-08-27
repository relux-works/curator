# Review verdict: changes requested

Route: `to-dev`.

## Acceptance-blocking finding

`TestManagerAndSystemConfigV1ExposeNoBuildPolicyOverrides` and `TestRegistryAndAuditSchemasRemainSourceEvidenceOnly` pin declared property names and scan those declarations for forbidden names, but `assertPropertySet` never asserts `additionalProperties: false`. `assertNoDeclaredProperties` likewise only walks declared `properties`; it does not guard schema closure. Consequently, changing any relevant root or nested declared object from closed to open would admit arbitrary `driver`, `build_policy`, `receipt_provenance`, or similar fields while every new compatibility test still passes. That does not satisfy the AC requiring fixed manager/system surfaces and registry/audit evidence boundaries to be guarded against expansion.

Required rework: explicitly assert closure for all relevant root and nested declared object shapes (following the referenced manager/audit schemas), or pin the 12 fixed manager/system/registry/audit schema files to their frozen origin/main hashes. Retain the exact property-set and forbidden-name assertions. Then rerun all required gates and two deterministic regeneration passes.

## Verified evidence

- Task delta against the accepted predecessor is limited to `tools/generate-vectors/main_test.go` and the approved `protocol/core.md` sentence.
- All 48 frozen legacy files are byte-identical to `57c1f56846d221ecc55786bd3c2467ec32f11730`.
- `go test -count=1 ./tools/generate-vectors`, `make validate`, `go vet ./tools/...`, `gofmt -d`, and `git diff --check` pass.
- In a disposable copy, two `make regenerate` passes and two `make regenerate-check` passes produced no conformance diff and the same aggregate conformance hash `8eb88a753b7a5cafb678dfb46dbf8fb4657fcb4a56a69cdbe1d39ebb3ae04f32`.

No implementation files were modified during review.