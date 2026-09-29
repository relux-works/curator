# Review note — TASK-260910-2vnjej rev6: identity review of a re-apply (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

You ACCEPTED rev4 (base 213a53e5, tree 2742e187, 26 paths, saved as `refs/campaign/6bo7ej-rev4-20260929`). Trunk moved, and the re-apply
conflicted in `.github/ci/conformance-gaps.tsv`, `internal/install/install.go` and `internal/registry/snapshot.go`. Rev6 (base cea992e2,
tree 12b6c386, 26 paths) is green on every lane. Content was accepted already; review ONLY fidelity.
1. The path sets are equal:
   `git diff --name-only 213a53e5 refs/campaign/6bo7ej-rev4-20260929 -- . ':!.task-board'` vs
   `git diff --name-only cea992e2 12b6c386 -- . ':!.task-board'`.
2. For every path except the 3 conflict paths, the +/- line multisets are identical (sort | diff). Report any difference.
3. On the 3 conflict paths, check three buckets and nothing else:
   (a) lines in neither side;
   (b) trunk lines dropped, i.e. lines that trunk's cea992e2 version has but rev6 lacks and rev4 did not remove;
   (c) rows that trunk REMOVED from conformance-gaps.tsv but rev6 re-added.
   Each must be empty or explained by a both-sides resolution. Show the resolution hunks in install.go and snapshot.go and confirm that
   both trunk's behaviour and the S2 cross-registry root check survive.
4. Run `go test ./internal/registry` and `go test ./internal/install -run 'Registry|Snapshot|Root|Mirror|Checkpoint|Bootstrap'`, and give
   the real exit codes.
accept_cr, or changes requested with file:line. No LOGBOOK.md. Do not spell any employer name anywhere.
