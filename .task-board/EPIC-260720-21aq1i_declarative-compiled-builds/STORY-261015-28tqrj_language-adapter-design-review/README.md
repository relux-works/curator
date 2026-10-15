# STORY-261015-28tqrj: language-adapter-design-review

## Description
Backlog: revisit the design of the language source-closure adapters and decide where to strengthen and where to simplify. Observations that triggered this: (1) go-v1, the path real skills use for compiled commands, is documented as an unchanged compatibility baseline and is not driven by the cross-adapter conformance suite or the closuregraph records that the six newer adapters share; (2) the SwiftPM, Rust and four Node adapters carry roughly twice the code of the Go path while few installed skills ship sources in those ecosystems, and the SwiftPM ones are macOS-oriented, so Linux SwiftPM coverage proves little in practice; (3) the hardened guarantees (network denial, read-only toolchain, private build-root writes) stay deferred for every adapter in portable mode and no verified provider ships. Not scheduled; to be picked up after the rc.7 normative wave.

## Scope
Design review only. No code change. Inputs: docs/source-closure-adapter-conformance.md, docs/authoring-language-adapters.md, internal/godriver, internal/buildsource, internal/closuregraph, internal/crossconformance, and the adapter packages.

## Acceptance Criteria
A written design note that (a) states for each adapter whether to keep, freeze, simplify or retire it, with the demand evidence behind that call; (b) says whether and how go-v1 joins the closuregraph records and the cross-adapter suite, with a cost estimate; (c) names which deferred hardened guarantees matter first; (d) proposes follow-up leaves for the accepted changes.
