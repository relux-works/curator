# Profile graph fixtures

`small/graph.json` and `dag-overlays/graph.json` are real command goldens.
`cmd/curator/profile_graph_test.go` creates synthetic immutable caches and
hash-verified local Git objects through `graphtest`, then invokes real CLI
dispatch. These fixtures contain no captured operator data.

The small golden replaces the original illustrative commit pins with real
content-addressed fixture objects, uses the actual build version, includes the
registered environment catalogue, and reports references as declared unread.
It has no project nodes because no project scope was selected. The original
wire example, its canonical lock and byte-exact emitted examples remain under
`small/contract-example.json`, `small/context-lock.json` and `small/emitted/`.
Those files retain the contract-round historical checks.

The second golden preserves two overlays sharing locator/revision but selecting
different directories and weights, and a nested DAG with one shared leaf.

Run `GOFLAGS= go test ./internal/profilegraph/... ./cmd/curator/...`.
Regenerate the command goldens explicitly with
`UPDATE_PROFILE_GRAPH_GOLDEN=1 GOFLAGS= go test ./cmd/curator -run '^TestProfileGraphCommandGoldens$' -count=1`.
The command itself never writes goldens or other files. Test logs go in `.temp/`.
