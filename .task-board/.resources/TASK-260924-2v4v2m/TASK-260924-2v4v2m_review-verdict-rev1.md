# TASK-260924-2v4v2m review verdict — CR rev1: ACCEPTED
Candidate tree ca48805350 (worktree == candidate, diff empty), base 1511b345. Reviewer: claude-opus-5-5 low.
- marker.go:473 v5LocalBuildShape = exactly the 7 members of curator-spec draft-sources-v1 install-marker-v5 go-v1 arm (additionalProperties:false); closedRawShape refuses count mismatch and null (rawJSONType null -> "").
- marker.go:840-853 production path validBuildState (called from Read) now dispatches go-v1 -> validV5LocalBuildShape; go-repository-v1 unchanged; only inside SchemaV5 branch, so v1-v4 untouched.
- marker.go:609 declared_tag uses identity.DraftSourceRefName — the same validator used by manifest/sources.go:115 (no second grammar).
- Tests drive marker.Read with raw rewritten bytes: repository/substituted/substitution:null, declared_tag "" and "release..next", plus valid tag preserved. No existing assertions changed; CHANGELOG only additive.
- Independent: go test ./internal/marker/ (disposable git-archive copy) ok.
- Mutants (re-applied by me): M1 local check -> `return ok`: KILLED. M3 tag check skipped for empty: KILLED. (M2 empty-only compare did not build — unused import; class covered by M3 + malformed row.)
Residual bound: no schema-level per-member value checks (sha256/policy grammar) in raw shape — decoded validation handles those, out of scope.
