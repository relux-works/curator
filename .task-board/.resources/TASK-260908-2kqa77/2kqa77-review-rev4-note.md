# Review note for TASK-260908-2kqa77 revision 4 (orchestrator, binding) — the gate now uses yaml.v3

Revision 4 answers your revision-3 verdict (`TASK-260908-2kqa77_review-verdict-rev3.md`: R1 High —
a first-position block scalar or nested map forged a missing `skip_upload`; R2 — duplicate keys
admitted; R3 — the wiring pin accepted a commented `run:` and an `if: false` step). Per the
orchestrator's ruling the awk walk is DELETED and the check moved to Go:
`tools/goreleaserconfig/gate.go` (`Check`/`CheckFile` via `gopkg.in/yaml.v3`, exact string `auto`
for every `homebrew_casks[*].skip_upload`, `scoops[*].skip_upload`, `release.prerelease`),
`wiring.go` (`CheckWiring` parses `ci.yml` and requires exactly one LIVE lint step — no `if` key —
running `go test -count=1 ./tools/goreleaserconfig/`), table-driven tests (43 rows), your R1/R2
fixtures committed byte for byte, the lint step switched to the Go test, the shell gate removed.

Verify:
1. Your three findings are closed by EXECUTED rows: run your own R1 fixtures (block scalar first,
   nested map first, the committed-config bypass) and the R2 duplicate-key fixture — each must now
   fail with field + entry + observed value; the committed `.goreleaser.yml` must pass.
2. R3: comment out the step's `run:`, and separately add `if: false` to the step — `CheckWiring`
   must fail both; also check that deleting the step, renaming the command, or adding a second
   live step fails, and that the check cannot pass if `ci.yml` itself fails to parse.
3. Attack yaml.v3 usage itself: anchors/aliases (`skip_upload: *a`), merge keys (`<<:`), tags
   (`!!str auto` vs `!!bool`), multi-document input, and a flow-style mapping — decide which must
   fail and whether any forges a pass.
4. The Go test really runs in a job that has Go on every push (the `Gate self-test` job has no
   `setup-go`); the old shell gate and its self-test section are gone with no dangling reference.
5. CHANGELOG reflects the final design; the historical source-level grep pin is no longer claimed.

Record exactly one verdict: `accept_cr(TASK-260908-2kqa77, revision=4, evidence=<your outcome
resource>)` on ACCEPT, or changes-requested with file:line and reproduction. No LOGBOOK.md writes.
