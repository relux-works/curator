# Document schema v6 authoring and CLI behavior

## Description
Update the schema index, conformance contract, and Curator CLI guide so an author or implementation developer can use schema 6 without inferring behavior from source code. Documentation must explain the complete declaration, toolchain and vendoring prerequisites, excluded build roots, cache and marker semantics, dry-run and status output, security limits, claim transition, and the process for standardizing future drivers.

## Scope
Work in curator-spec after normative prose, schemas, and vector names stabilize. Own schemas/v1/README.md, conformance/README.md, and cli/curator.md. Include one complete schema 6 example and link exact new schemas and vector files. Do not claim that managers already ship support, add generic build flags, or move cross-manager interoperability guidance from the downstream interop story into this protocol task.

## Acceptance Criteria
The schema index lists canonical and legacy v6, receipt v1, marker v2, and claim v2 and states that schemas 1 through 5 and claim v1 remain supported historical contracts; conformance docs define rc.4 manager evidence, v1 versus v2 claim handling, build fixture and vector execution, suite hash meaning, and required failure behavior; CLI docs explain install and upgrade build phase, fixed go-v1 prerequisites, build_roots context exclusion, cache status and currentness, dry-run results without compilation, diagnostics for unsupported or corrupt inputs, repair and GC implications, and no package argv or hook option; a future driver section requires a new identifier, closed schema, threat review, fixed process graph, cache identity, and vectors; links resolve and make validate passes.
