# TASK-260910-3kvq02 — review handoff evidence

## Candidate

Base: `65c6f1ec4b18c2374a9381028d12b78149114e7a`.
Candidate tree: `a26dc0d83293435de0dd22a2c64700ec8cc10dee`.
The attached candidate JSON records SHA-256 for all six changed/new paths.
An isolated index was used to derive the tree; the managed branch was not
committed, switched or reset. Existing untracked worktree board artifacts were
left untouched.

## Contract and existing outcomes inspected

Read current curator-spec `protocol/skillfile-sources.md`,
`protocol/repository-transport.md`, `docs/skillfile-sources.md`, and draft
semantic cases. Read this task's earlier authorization-plan outcome and the
accepted parser's `TASK-260910-24cuys_review-verdict-rev1.md`. The historical
planning-only blocker is superseded by the attached implementation authorization.
No earlier test evidence is claimed as a run of this candidate.

## Delivered behavior and production paths

- `manifest.Expand`: parsed draft selectors become individual skill units in
  declaration order, then UTF-8 folder byte order. Relative local paths use
  the declaring Skillfile directory; native path contents are literal. Git
  aliases require an acquisition-owned tree and do not fetch during expansion.
- Existing parser alias/name/path/include/exclude grammar is reused. Literal
  existence/type precedes exclusion. Wildcards discover immediate directories
  only; overlap with literal includes deduplicates by folder. Missing excluded
  literals, empty results and invalid discovered metadata/packages fail.
- SKILL.md name, description and optional triggers are validated; individual
  names and optional schema-1 machine-manifest names must agree. Metadata links,
  special files and non-directory fallback parents fail before metadata reads.
- Physical selector containment uses symlink resolution and SameFile ancestry.
  Project managed outputs, configured machine outputs and Git metadata are
  pruned; explicit selections into them fail. Output inspection failures are
  not treated as absent boundaries. Portable destination collisions fail across
  individual, collection and legacy declarations.
- `closure.BuildExpanded`: expanded acquired nodes feed the existing dependency
  queue, command narrowing and provider-first topological sort. Diamond
  dependencies remain one provider with separate consumer edges; version,
  repository and destination conflicts fail. Local identity is never represented
  as a Git commit. A Git alias must supply one commit for all members.
- Legacy `closure.Build` refuses unresolved selectors before acquisition; its
  existing valid schema-1 behavior remains covered by package regressions.

## Direct validation (zsh, standalone processes, no tee/pipeline)

All following commands ran against the candidate bytes and exited **0**:

1. `go test -count=1 ./internal/manifest ./internal/closure ./internal/identifiers`
   — manifest 0.841s, closure 18.187s, identifiers 2.533s.
2. `go vet ./internal/manifest ./internal/closure ./internal/identifiers`.
3. `golangci-lint run ./internal/manifest/... ./internal/closure/... ./internal/identifiers/...`
   — 0 issues.
4. `go build ./...`.
5. `git diff --check`.
6. Formatting assertion over the five changed Go files — exit 0.

Earlier development failures, not represented as green evidence:
- First `go test -count=1 ./internal/manifest ./internal/closure`: exit 1;
  wildcard pruning of a link into output was performed too late. Fixed.
- Next `go test -count=1 ./internal/manifest ./internal/closure ./internal/identifiers`:
  exit 1; new dependency fixtures incorrectly used `tag` instead of the existing
  `ref:{kind,value}` contract. Fixed the fixtures, retaining production grammar.
- Initial scoped golangci-lint: exit 1, one G304 for the fixed metadata filename.
  Added the narrow contained-path justification; subsequent lint exits 0.
- Focused expansion/closure runs after those corrections exited 0.

## Gate attacks

Attached mutation harness and raw per-mutant outputs record each exact `go test
-count=1` command and real exit code. **8/9 narrowing mutants killed**, each by
an assertion failure with exit **1**, not a compile failure. Sites: explicit
members before excludes, case-equivalent destinations, discovered names,
description types, physical escape, optional manifest identity, alias commit
consistency, and output read errors.

One survivor, exit **0**: narrowing the metadata symlink clause to the canonical
manifest. This is behaviorally subsumed: the subsequent regular-file check
rejects symlink mode for either manifest, and the directory check rejects it
for an intermediate fallback parent. The test still refuses the link. This is
not proof of every refusal clause; only these nine mutations were measured.
Candidate file SHA-256 values were checked after restoration and match exactly.

## Explicit bounds

Host evidence: Darwin amd64 only. Other OS/architecture lanes are unverified.
`CURATOR_CONFORMANCE_ROOT` was unset; optional external legacy-vector tests
skip. Bundled draft parser fixtures and package regressions ran. This is not a
claim that all 73 campaign semantic cases execute through installation.

The production entry points tested here are the internal expansion and closure
APIs. The acquisition callback is a trusted internal boundary; test fixtures
supply immutable temporary trees, not a production local-snapshot implementation.
No CLI opts into schema 2 yet. Local snapshot capture, root-input policy, lock
replay, repository transport, audit and publication are separate work; no
installation, launch, authentication or cross-platform claim is made. Required
machine output roots must be supplied by the acquisition owner. The integration
boundary is documented in `docs/draft-source-expansion.md`.

The full configured landing suite was not manually run. The handoff runtime
owns that suite and the Change Request publication exactly once. Independent
review remains required.

Post-mutation restoration check: `go test -count=1 ./internal/manifest
./internal/closure -run 'TestExpand|TestBuildExpanded'` also exited **0**
(manifest 2.912s, closure 16.474s), after matching all candidate hashes.
