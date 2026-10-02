# Review note — TASK-261001-3bsyvh rev4 re-apply identity review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev4 re-applies the accepted rjxrgs content onto the current trunk. The gate is green. The normative content is `refs/campaign/rjxrgs-rev1-20261001` (fe2b5f61, the accepted rjxrgs rev1 on 5432c85f). Your own earlier identity verdict covered rev1, and `refs/campaign/3bsyvh-rev3-20261001` (d540fbb1) holds the accepted rev3 on e87d488b.

Verify:
1. For every non-`.github/ci/*.tsv` path, rev3 equals rjxrgs rev1 in substance. Use a two-way +/- line comparison of the per-path patches. Differences may come only from trunk context; prove it with blob or merge-tree comparison.
2. The three `.github/ci/*.tsv` files hold the union of trunk rows and rjxrgs rows. Counts are exact: independently confirm `go test ./internal/conformancecoverage -count=1` exits 0, and recompute one changed count row.
3. Check for stray files, LOGBOOK, and the employer name; never spell it.

accept_cr (citing the rjxrgs rev1 substantive verdict) or changes requested with the exact hunks.
Trunk now also carries the Muse conformance rows; the tsv union must keep them.
