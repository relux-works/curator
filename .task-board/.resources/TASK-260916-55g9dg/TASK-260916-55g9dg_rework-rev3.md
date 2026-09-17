# Rework brief — TASK-260916-55g9dg, revision 3 (answers review verdict rev2)

Read `TASK-260916-55g9dg_review-verdict-rev2.md` first; it is the authority
for this round. Everything it marks as implemented/passing stays. Same Story
worktree; rules in `remediation-manager-producer-rules.md`.

## 1 (high) — remove the forbidden vector-comparison workaround
`internal/config/environments_conformance_test.go:33-66,224`
(`postPinKnobs` / `prunePostRevisionKnobs`) deletes the new knobs from the
actual normalized output whenever the root's expected vector lacks them.
Delete that mechanism and restore the exact byte comparison. Yes, this makes
`TestManagerConfigV2Vectors` fail on the hosted gate's pinned rc.11 root —
that is the known, orchestrator-owned pin-lag blocker (`spec-pin-lag-hold.md`);
it is resolved by promoting `SPEC_PIN`, never in this test. State the
expected failure in your results.

## 2 (medium) — reject an explicit null `system_module_waivers`
`internal/config/environments.go:281` treats a present `null` as absence.
The landed schema (`manager-config-v2.schema.json:501`) requires an array;
§12.1 says list with empty default. Reject `null` with the configuration
error the closed grammar uses for a wrong type, keep absent = `[]`, and add
a committed regression test at the `Load` production entry point.

## 3 (medium) — explicit root-content coverage for the admission subset
`internal/interop/environments/context_materialization_test.go:170-175`
iterates whatever cases the root has, so at rc.11 the five admission cases
silently yield nothing. Add a dedicated driver for the admission subset
(select the five cases by name/family) that runs all of them when the root
publishes them and otherwise takes the explicit `root-content` skip
(`t.Skipf("%s publishes no system-module admission cases", root)`), with a
matching `.github/ci/platform-cases.tsv` row (class `root-content`), like the
existing `internal/godriver` precedent.

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l internal/ cmd/`,
`go test -count=1` for `./internal/config/... ./internal/contextmaterialize/...
./internal/envprofile/... ./internal/interop/environments/... ./cmd/curator/...`
with the conformance root at
`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`
where everything must pass; against the rc.11 root only the manager-config
vector comparison may fail). Update `TASK-260916-55g9dg_results.md`
(1–3 closure with file:line, transcripts, the expected pinned-root failure),
tick the checklist, then `task-board handoff TASK-260916-55g9dg --role developer`.
