# Publish protocol 1.0.0-rc.4 release metadata

## Description
Update release-facing metadata for the schema 6 protocol candidate. Record the reader and writer compatibility transition, security impact, new conformance identity, and release gates without claiming that downstream implementations or releases exist before their real commits land.

## Scope
Work in curator-spec after validation gates and authoring documentation. Own README.md, COMPATIBILITY.md, CHANGELOG.md, and RELEASE.md plus only version metadata directly contained in those files. Set protocol version to 1.0.0-rc.4 and the actual change date. Reference claim v2 and the generated rc.4 suite. Do not change implementation pins, fabricate review acceptance, tag state, checksums, signatures, attestations, or released manager versions.

## Acceptance Criteria
README identifies 1.0.0-rc.4 and indexes schema 6, marker v2, receipt v1, claim v2, and the compile-only decision; COMPATIBILITY gives an explicit reader and writer matrix for schemas 1 through 6, marker v1 and v2, claim v1 rc.3 and claim v2 rc.4, unsupported old readers, and no migration requirement for existing schema 1 through 5 packages; CHANGELOG records declarative go-v1 builds, context exclusion, dry-run and transaction rules, conformance additions, and security tradeoffs; RELEASE adds rc.4 validation, deterministic regeneration, claim, decision, and interoperability prerequisites without marking unmet evidence complete; version strings agree with conformance manifest and claim v2; make validate and make regenerate-check pass.
