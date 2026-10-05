# TASK-261004-3pvg2k — introduce-cips-directory-and-process

Developer evidence for RUN-261004-06bd1d, 2026-10-04.

## Changes

- Added cips/README.md with purpose, issue/Decision/normative relationships, sequential never-reused numbering, lifecycle, required sections, and index.
- Added cips/TEMPLATE.md preserving the supplied binding metadata and section list.
- Added CIP-0001-curator-improvement-proposals.md with Status: Accepted (operator, 2026-10-04), all template sections, alternatives, and the adoption boundary.
- Reserved CIP-0002 through CIP-0006 as In preparation index rows with the five supplied titles. Reservations are explicitly distinct from lifecycle statuses and have no broken links to absent files.
- README.md and GOVERNANCE.md each link the index/process, template, and CIP-0001.

Exactly five docs files changed. No changes to protocol/, profiles/, schemas/, conformance/, release/, LOGBOOK.md, or CHANGELOG.md. No runtime behavior or tooling changes, so no new runtime tests were added. Existing validation and explicit documentation acceptance checks cover this scope. No commit was created; changes remain uncommitted in the isolated curator-spec Story worktree for review.

## Validation run directly by this developer

All gate commands ran as standalone processes with no tee or shell pipeline. No previously attached validation evidence was substituted.

1. `env PATH="$PWD/.temp/TASK-261004-3pvg2k-venv/bin:$PATH" make validate` — **exit 0**.
   - `python3 tools/validate.py`: validated 73 schemas and 1294 vector files, including local Markdown links; exit 0 as shown by make advancing to the next recipe.
   - `python3 -B -m unittest discover -s tools -p 'test_*.py'`: 672 tests in 555.100 seconds, OK; exit 0 as shown by make advancing to Go.
   - `go test ./tools/...`: github.com/relux-works/curator-spec/tools/generate-vectors passed in 6.487 seconds; make returned exit 0.
   - Python dependencies were installed from requirements-dev.txt in an ignored task-scoped venv. Venv creation and dependency installation each exited 0.
2. `git diff --check` — **exit 0**.
3. `git diff --exit-code HEAD -- protocol profiles schemas conformance release LOGBOOK.md CHANGELOG.md` — **exit 0** (no changes).
4. Standalone `python3` documentation acceptance assertions — **exit 0**:
   - Required section headings: 12/12 in the template and 12/12 in CIP-0001; template title and all five supplied metadata fields present.
   - Required index rows/statuses/titles: 6/6; CIP-0001 Accepted, CIP-0002..0006 In preparation.
   - Required root discovery links: 6/6 across README.md and GOVERNANCE.md.
   - Changed-path allowlist: exactly the five intended docs files, including all untracked files.
   - Final newlines and no trailing spaces/tabs in those five files.
   - Manifest bytes equal HEAD and exact SHA-256 equals the required baseline.

Core manifest SHA-256 before and after:
`6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`.

The documentation assertions were an inline acceptance check; they did not add a production validator or claim runtime/platform coverage. No expected-red gate or skipped required check.

## Decisions and scope

The earlier repository-binding blocker is resolved: this run uses the isolated curator-spec Story worktree. LOGBOOK.md was not edited because the binding task brief explicitly prohibits it; decisions and evidence are recorded here and in board notes. The conditional logbook checklist is not applicable under that instruction. No CHANGELOG entry is required: the repository describes its changelog as notable protocol changes, and this process changes only documentation.

The separately attached TASK-261004-3pvg2k_docs.patch contains the full five-file documentation change for review. Operator acceptance of the process is from the binding instruction, not inferred from this developer handoff. Future CIP acceptance does not bypass governance or attest that implementation work has shipped.
