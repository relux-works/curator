# TASK-260720-akf5kh producer evidence

## Change identity

- CocoaSkills base: `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09`
- Branch: `task/TASK-260720-akf5kh-schema-v6-user-docs`
- Signed commit: `dacccaaf3ed18740a4d501fe8a3bfec64644c03e`
- Pull request: https://github.com/ivanopcode/cocoaskills/pull/20
- Exact-head CI: https://github.com/ivanopcode/cocoaskills/actions/runs/30686365518
- Accepted protocol source: signed tag `v1.0.0-rc.5` at
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`

The commit changes only `README.md`, `README.ru.md`, `ARCHITECTURE.md`,
`ARCHITECTURE.ru.md`, `SECURITY.md`, `SECURITY.ru.md`,
`docs/skill-authoring.md`, `CHANGELOG.md`, and `LOGBOOK.md`. The worktree was
clean after the commit. No product code, test, workflow, or configuration file
changed. No merge, release, or tag was created.

## Delivered contract

The English source documentation and maintained Russian mirrors now describe
the mixed schema-6 manifest, build-root exclusion and containment, the fixed
vendor-only Go 1.25 implementation family against the Go 1.23 protocol floor,
the normative four-node `manager-worker-v1` boundary, exact native control
inventory and result-only capability evidence, deferred hardened guarantees,
macOS/Windows support and explicitly owned Linux follow-ups, logical versus
physical cache identity, protected-state provenance, transaction and dry-run
ordering, status and reinstall repair, locked GC, and direct Unix/Windows
activation. The README command and local validation sections remain present.

## Green evidence (real exit codes)

- `git verify-tag v1.0.0-rc.5`: exit 0; good signature. `git rev-list -n 1`
  resolved the tag to the accepted commit above (exit 0).
- Fenced JSON parser: exit 0; 25 JSON examples across 7 documentation files.
- Mixed manifest semantic check: exit 0; the examples in both READMEs and the
  authoring guide are identical and accepted by `csk.skillspec.load_skill_spec`.
- Local Markdown link/anchor check: exit 0; 73 links across 9 files.
- Schema-6 contract coverage check: exit 0 across 7 source documents.
- Code-to-document constant check: exit 0 for Go family, policy, graph,
  resource bounds, inventory/availability, deferred guarantees, cache layout,
  and GC grace.
- Focused documentation-relevant behavior suite: exit 0; 629 passed and 70
  skipped across schema, toolchain, build, cache, activation, currentness,
  transaction, status, CLI, and shim tests.
- `python -m mypy`: exit 0; no issues in 68 source files.
- `git diff --check` and `git diff --cached --check`: exit 0.
- `python -m build --outdir .../dist-handoff`: exit 0.
- `python -m twine check .../dist-handoff/*`: exit 0; wheel and sdist passed.
- `git commit -S`: exit 0; `git verify-commit HEAD`: exit 0 with a good
  signature.
- Branch push: exit 0; remote head equals the signed commit.
- Settled `gh pr checks 20`: exit 0. Exact-head CI run 30686365518 completed
  successfully with 14/14 jobs: 12 Python 3.11-3.14 jobs on Ubuntu, macOS, and
  Windows, strict mypy, and distribution build/metadata checks.

## Red and diagnostic evidence (reported as red)

- The host had no `python` command (exit 127), and the initial system
  `python3 -m pytest --version` and `python3 -m mypy --version` probes each
  exited 1 because those modules were absent. A task-scoped virtual environment
  was created and the development dependencies installed (both exit 0).
- An early JSON gate exited 1 on a multi-document fence; the independent
  fragments were relabeled as text. A later ad-hoc parser also exited 1 because
  its regex ignored indented Markdown closing fences; the checker was corrected
  before the green 25-example run.
- Early mirror/coverage checks exited 1 for a checker quoting syntax error and
  then for missing Security mirror identifiers and numeric limits. The checker
  syntax was corrected and the genuine documentation gaps were added.
- Expanded coverage checker attempts exited 1 while overly literal English
  tokens were applied to equivalent prose and Russian translations. The gate
  was narrowed to stable protocol identifiers plus language-appropriate prose;
  the resulting contract check exited 0.
- The out-of-scope-term `rg` diagnostic exited 1 as expected because it found
  no rc.6, schema-7, repository-v7, receipt-v2, or marker-v3 terms.
- Hosted-check polls exited 8 while jobs were pending. Two interactive watcher
  processes exited 1 while the workflow itself remained active and failure-free;
  these were monitor-process failures, not test failures. The settled exact-head
  query later exited 0 with all 14 jobs green.

## Review notes

The task logbook entry records the protocol-floor versus trusted-family,
logical-portability versus csk-layout, receipt-consistency versus provenance,
platform scope, and deferred-hardening decisions. Review should compare only
against the accepted rc.5 protocol tag and landed CocoaSkills base; later
protocol revisions are intentionally outside this task.
