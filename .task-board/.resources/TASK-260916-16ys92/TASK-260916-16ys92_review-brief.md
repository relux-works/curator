# Review brief — TASK-260916-16ys92 (launcher provider-path line, E4), review round 1

You are the independent reviewer of a curator-agent-launcher change produced
for `TASK-260916-16ys92` (story `STORY-260916-2otjbn`, wave 1 of the 2026-09
security-audit remediation). Repository: curator-agent-launcher (this
control root); the candidate is the Story worktree
`<control-root>/.temp/STORY-260916-2otjbn/worktree` = the published Change
Request revision 1 (`TASK-260916-16ys92_change-request_rev1.patch`). Read the
producer brief `TASK-260916-16ys92_brief.md` and results
`TASK-260916-16ys92_results.md` first. Models: reviewers gpt-6-astra low.

## What to verify (read-only; do not edit the candidate)
1. **Contract**: SPEC.md §2 and §4.3 record the provider-path line
   (`curator-run: provider: path=<absolute path>`, symlink-resolved own
   executable, fallback on resolution error that never fails the launch),
   the producer's path-only decision with its rationale and the deferred
   closed origin set; §8.1 revision history and the specification changelog
   bumped (`0.3.0-draft` → `0.4.0-draft`); CHANGELOG Unreleased E4 entry.
2. **Implementation** matches the SPEC text exactly: emitted in the same
   line-group at every launch before the plan request, folded with the
   existing framing rule (a hostile path with newlines cannot split the
   group — try one), no new flags, §4.3 model/effort semantics untouched,
   `pi` unchanged beyond the line.
3. **Tests/goldens**: pipeline goldens show the line with host-specific
   values normalized the way the repository already does; unit tests for
   fold and fallback; re-run `go build ./... && go vet ./... && gofmt -l . &&
   go test ./...` yourself from the worktree (`set -o pipefail`; quote exit
   codes) and, if the repository has lint config, the linter CI uses.
4. **Scope**: `git diff --stat origin/main` lists only SPEC, CHANGELOG,
   the line-group code, goldens and tests; no ax, no defaults changes.

## Verdict
Record `TASK-260916-16ys92_review-verdict-rev1.md` (task outcome) with the
per-item table (file:line quotes), transcripts and findings; then either
`task-board m 'accept_cr(TASK-260916-16ys92, revision=1, evidence=TASK-260916-16ys92_review-verdict-rev1.md)'`
or a changes-requested verdict routed with `set_status(TASK-260916-16ys92, status=to-dev)`
listing the concrete corrections.
