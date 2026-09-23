# TASK-260906-2b3nar review verdict — revision 4 (refresh-only) — ACCEPTED

Reviewer: claude (reviewer run), 2026-09-23. Scope per 2b3nar-review-rev4-note.md: verify the refresh only; fold logic was accepted at rev1.

## Evidence (re-executed by reviewer)
- Worktree tree (temp index over HEAD 48da2690 + `add -A`) = `0e5edef1d4c9eaac4525a9e88b3c765f520e4f5e` = CR candidate tree.
- Green gate run 35744083039 head f17d77fd → tree `0e5edef1…` (gh api), conclusion success: gate ran on the exact candidate tree.
- Rev1 vs rev4 patch, per file, hunk bodies (minus `index`/`@@` lines) identical for all 4 paths: platform-cases.tsv, dirfold_test.go, gitops.go, CHANGELOG.md. The CHANGELOG fold entry appears exactly once; the trunk rc.12 content comes from the base (48da2690), so both entries are present.
- SPEC_PIN = dced9b8317e0e8af79edf2d0539b32bd22b6c85b (.github/workflows/ci.yml:53), inherited from base. The gitops/snapshot rows are not vector-consuming; they passed identically.
- Red gates rev2 (35735392681) and rev3 (35739653315): downloaded test-evidence-windows-latest/go-test.json. The ONLY `fail` actions are internal/managerlock TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked (+ package). `go list -deps ./internal/managerlock` has 0 internal/gitops deps. So this is BUG-260922-6chzf9, unrelated.
- zsh, `set -o pipefail`: `go test ./internal/gitops ./internal/snapshot -count=1` → ok/ok, exit=0.
- `go test ./internal/gitops -count=1 -v -run 'Fold|DirFold|Prefix'` → exit=0: RefusesDirectoryComponentFold PASS, RefusesFileDirectoryFold PASS, RefusesNestedDirectoryComponentFold PASS, AdmitsDirectoryComponentFoldWhenCaseSensitive SKIP (named reason: this host's temp dir is case-insensitive; same bound as rev1 — the hosted linux lane executes it).
- ledger-consistency.sh needs a gate evidence dir, so I did not rerun it locally. I accept the green hosted run's platform-case gate for it.

## Findings
None blocking. Mutants were not rerun because the logic is byte-identical to rev1, where they were accepted.

Verdict: ACCEPT revision 4.

## Recording outcome (appended)
`accept_cr(TASK-260906-2b3nar, revision=4, …)` was REFUSED by the runtime: `validation_not_bound_to_tree: … the validation evidence carries no source tree identity … revalidation is required (candidate_tree_oid=0e5edef1…, evidence_tree_oid=0e5edef1…)`. The content verdict above is ACCEPT, but the board cannot record it against this validation record. Routed to `to-dev`: the only thing needed is to republish/revalidate (handoff again) so that the validation evidence is bound to the tree. No code change is requested. The next reviewer can reuse this verdict's checks after confirming the new tree is still 0e5edef1….
