# Reviewer verdict for TASK-260811-3ksxig

Verdict: **changes requested -> to-dev**

## Review authority

- Reviewer run: `RUN-260823-872c60`
- `task-board spawn goal RUN-260823-872c60`: no active goal; the run is not goal-bound.
- Reviewed producer outcome: `TASK-260811-3ksxig_implementation-report.md`
- Reviewed implementation scope: `internal/pnpmsource`, the pnpm-related artifact-policy text additions, tests, and documentation.
- No product code was modified by this reviewer.

## Required changes

### 1. Declared patches cannot survive materialization reconciliation

`validateMaterializedTree` compares every installed package byte-for-byte with
the unpatched admitted tarball inventory (`internal/pnpmsource/materialize.go`,
lines 492-504). A correctly applied pnpm patch necessarily changes that
inventory, so real patched materialization is rejected with
`closure_integrity_mismatch`. The positive patch test only parses and captures
the patch (`conformance_test.go`, lines 308-335); it never derives a store or
materializes the patched graph. Its placeholder `manager-hash-a` also does not
prove manager-hash reconciliation.

Required rework: model patch identity in the exact snapshot/package context,
verify the manager patch hash according to the pinned pnpm profile, derive the
expected patched file inventory from admitted tarball + admitted patch under a
declared/receipted transform, and add a positive end-to-end patched
materialization test plus stale-hash/content-drift negatives.

### 2. Installed-tree graph reconciliation admits undeclared extra content

The virtual-store walk ignores `filepath.WalkDir` errors and silently skips
directories without a readable valid `package.json`
(`internal/pnpmsource/materialize.go`, lines 443-476). Those directories remain
in the final materialized tree, while only recognized package roots contribute
to the observed package set. Therefore an extra `.pnpm/.../node_modules/...`
tree with missing/malformed metadata can evade `closure_graph_incomplete` and
become admitted source. This violates shared S07 and the requirement that an
installed layout cannot widen or repair the lock graph.

Required rework: propagate every walk/read/JSON error, reject every unclaimed
virtual-store entry/member, reconcile the complete owned layout to exact
snapshot identities, and add negative fixtures for metadata-less, malformed,
and otherwise unclaimed extra virtual-store content.

### 3. The closed YAML parser accepts trailing documents

`decodeClosedYAML` decodes only the first YAML document and checks that
document's root count, but never requires EOF from the decoder
(`internal/pnpmsource/lock.go`, lines 261-270). The second typed decode has the
same one-document behavior. A review probe containing a valid 9.0 lock followed
by `---` and a conflicting 10.0 document returned `version="9.0" err=<nil>`.
The hidden bytes affect only the raw digest, not the parsed/canonical graph.

Required rework: require exactly one YAML document and EOF for both lock and
workspace configuration parsing; add multi-document/trailing-content negative
tests with `closure_lock_format_unsupported`.

### 4. Exact peer-context replay is not verified

Parsing creates peer edges from snapshot-key parentheses, but materialization
validation collapses selected snapshots to `PackageKey -> count` and observed
packages to `name@version -> count` (`materialize.go`, lines 435-490). It does
not bind each virtual-store directory to its exact snapshot key or verify the
installed peer dependency links. Two peer-context instances can therefore be
miswired or substituted while preserving package/version multiplicity.

Required rework: reconcile each installed virtual-store instance and its peer
links against the exact selected snapshot/peer context, and add two-context
positive and swapped/missing-peer-link negative fixtures.

## Validation evidence

| Command | Result |
| --- | --- |
| `go test -count=1 -cover ./internal/pnpmsource` | pass; 79.3% statements |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/pnpmsource` | pass |
| `go vet ./internal/pnpmsource ./internal/artifactpolicy ./internal/nodesource` | pass |
| `golangci-lint run ./internal/pnpmsource ./internal/artifactpolicy` | pass; 0 issues |
| `git diff --check` | pass |
| closed-lock review probe | parser incorrectly accepted a trailing second YAML document |

Passing current tests does not satisfy the acceptance criteria because the
patch materialization path, complete installed-layout ownership, exact peer
linkage, and multi-document rejection are not covered and the implementation
fails those security properties by inspection/probe.

## Routing

This is ordinary implementation rework with no external or human-only blocker.
Route to `to-dev`; a new reviewer cycle is required after fixes and new vectors.
