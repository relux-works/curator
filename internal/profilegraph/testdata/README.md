# Profile graph contract fixtures

`small/` is entirely synthetic. The root requires one nested profile and one
skill; the composition adds one module-free overlay from a selected directory.
The two modules include an environment selector, a quote and a fenced example.
The fixture also admits one project instruction file. All remote locators use
the reserved `.invalid` namespace, and the commit ids are synthetic pins.

`graph.json` is the proposed wire example. `context-lock.json` is a canonical
lock consumed by the current materializer. `emitted/` holds its byte-exact
outputs for `claude_code` and `codex_cli`. The contract tests recreate those
outputs in memory and validate the schema, input transcript digest, identity
tuples and additive byte ledger. The listing's normalized content is separate
from the source `SKILL.md`; body and reference bytes are not in either emitted
document. Runtime loading remains unknown.

Run `GOFLAGS= go test ./internal/profilegraph -count=1` from the repository root.
Update the example and its measured hashes together when the contract changes.
There is no graph command or fixture capture from an installed home. Schema
and accounting controls do not qualify the future filesystem safety policy.
