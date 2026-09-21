# BUG-260921-3cgij4: install-marker-v5-valid-fixture-violates-build-source-rule

## Description
conformance/draft-sources-v1/schema-cases/install-marker-v5/valid.json (index valid:true) has builds: {} together with a top-level build_source. protocol/skillfile-sources.md:268 says top-level build_source remains required exactly for active local go-v1 commands and absent otherwise, so with zero builds the document is normatively invalid although JSON-Schema-valid (the cross-field rule is not expressible in the schema). The curator reader (internal/marker validBuildState) refuses it correctly; curator keeps this case as a documented bound (BUG-260920-2eg8nv results: 111 driven / 4 bounds). Operator decision 2026-09-21: keep valid.json as the fixture WITH an active local go-v1 build (add a builds entry so build_source is legitimately required), and add a sibling positive fixture without builds and without build_source so the absent-otherwise branch is covered; regenerate/adjust the index; do not weaken the normative rule.

## Scope
curator-spec conformance/draft-sources-v1/schema-cases/install-marker-v5 fixtures + index; schema/cases regeneration checks (make validate)

## Acceptance Criteria
valid.json carries an active local go-v1 build consistent with its build_source; a second positive fixture covers builds absent + build_source absent; index updated; make validate green; the discrepancy text is cross-referenced in the case comment; a curator follow-up (pin bump + driven row) is named in results
