# Developer outcome — author language-adapter guide

## Delivered

Created only docs/authoring-language-adapters.md. The guide is author-oriented and covers the supported_build predicate; C0-C7 evidence flow; selection-neutral capture and exact target binding; the closureexec seam and unchanged guard allowlist; reject-by-default source analysis across spelling, position, build-setting-kind, and macro-oracle-input axes; the verified OBSERVED-READ boundary; CGP05/CGP10; all seven obligations; rejection-matrix integration; per-ecosystem concerns; and validation/evidence expectations. It links the consumer conformance document instead of reproducing its profile, diagnostic, migration, and matrix tables.

## Grounding

- Predicate and delivery boundary: .spec/skill-facing-cli-source-closure.md and docs/source-closure-adapter-conformance.md.
- C0-C7, graph authority, and canonical identity: internal/closuregraph.
- Permit/receipt process boundary: internal/closureexec/acquisition.go and portable_runner.go plus adapter guard tests.
- Recursive payload policy: internal/artifactpolicy.
- Worked adapters: internal/rustsource, npmsource, pnpmsource, yarnclassicsource, yarnmodernsource, swiftpmsource, swiftpminterop, and swiftpmbuild.
- Reject-by-default lesson and observed-read boundary: swiftpminterop parser/build-setting/read-set tests and swiftpmbuild/readset.go.
- Seven obligations, 19 rejection vectors, independent CCJ-1 oracle, and committed export: internal/crossconformance.
- CI and evidence: .github/workflows/ci.yml, .github/ci/test-gate.sh, Makefile, and CONTRIBUTING.md.

## Validation

- Ruby Markdown structure, UTF-8, trailing-whitespace, fence, heading, and 28-link check: exit 0. Missing worktree targets were verified read-only against codex/legacy-board-repair because the Story worktree predates their integration.
- Required-content and English-language check over 24 anchors: exit 0.
- git diff --check: exit 0.
- Scope audit: the only new file is docs/authoring-language-adapters.md; the 10 modified cmd/curator files were present before this run and remain untouched. README.md and CONTRIBUTING.md are unchanged.
- No go build or go test was run, as the docs-only brief explicitly forbids those gates.

No files were staged or committed.