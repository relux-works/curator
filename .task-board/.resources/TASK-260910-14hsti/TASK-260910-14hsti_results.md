# TASK-260910-14hsti — review handoff

## Candidate and contract

- Story worktree base HEAD: `12f1287ee0fb538f9ca004dd53b870e824e5baf2` (STORY-260910-197y84 parser+collections integrated).
- Uncommitted candidate; the handoff snapshot publishes the authoritative tree.
- Spec checkout: curator-spec main `871d11b`. Read protocol/skillfile-sources.md (§2, §5),
  protocol/repository-transport.md (revisions 1+2 — revision 1 semantics only),
  schemas/draft-sources-v1/source-policy-v1/v2, conformance/draft-sources-v1
  (semantic-cases.json, README), docs/skillfile-sources.md.
- Scope kept to `internal/snapshot`, `internal/staging`, `internal/privatedir`,
  `internal/adapters` plus tests. No manifest/closure/transaction/config changes;
  frozen v1 schemas untouched; diagnostics named exactly per spec §5.

## Implementation

- `internal/staging/boundaries.go`: `Canonicalize` (absolute+clean+EvalSymlinks
  with missing-prefix handling, fail-closed), `Within` (canonical string prefix
  plus SameFile ancestor walk for symlink and actual-FS case equivalence),
  `ContainsGit` (case-insensitive), `IsOutputPath`, `Plan.Snapshot` (plan-time
  physical identity; entry targets resolve the parent only, never the final link)
  and `Plan.Recheck` (retarget since Snapshot, byte-target final-link refusal,
  entry link-destination check, admitted-overwrite refusal in both directions).
- `internal/snapshot/boundaries.go`: `ValidateLocalPackage` (alias known,
  directory portable, source exists, pre-resolution output prune check, resolve,
  source containment else `source_selection_invalid`, post-resolution output check,
  root-package detection via SameFile against project root, declared runtime/build
  disjointness for non-root), `ValidateRootInputs` (unknown alias, duplicates,
  portable/link-free/per-entry existence and regular-or-dir type, `.git` and
  output disjointness both directions, pairwise entry overlap, SKILL.md + effective
  manifest + declared runtime/build coverage via skillspec), `EnumerateInputs`
  (full tree or root selections; deterministic output/.git pruning; link and
  special-file refusal; pruned-root/entry is an error). Broad `path: "."` with a
  safe selected subdirectory stays valid; only the selected package is judged.
- `internal/adapters/boundaries.go`: `ProjectOutputRoots` (.agents plus every
  AgentPaths dir), `ValidateDestinations` (each live path's full canonical and
  entry-location canonical against admitted inputs both directions, plus live link
  destination into admitted). `internal/adapters/adapters.go`: `unmanagedConflict`
  now fails closed on non-NotExist inspection errors and forces symlink/marker
  proof when the on-disk spelling is a case-variant of the requested name
  (`casingAliasUnmanaged`); exact-match ledger adoption unchanged.
- `internal/privatedir/staging.go`: `TempStaging` (MkdirTemp + Protect + Validate,
  no residue on failure) for private snapshot capture staging.

## Direct validation (bash, standalone commands, `set -o pipefail` where piped)

All commands run by this producer on the restored candidate; no prior evidence accepted.

| Command | Exit | Evidence |
|---|---:|---|
| `go test -count=1 ./internal/staging/ ./internal/snapshot/ ./internal/adapters/ ./internal/privatedir/` | 0 | All four touched packages green (staging 0.48s, snapshot 2.51s, adapters 0.82s, privatedir 1.54s) |
| `go test -count=1 ./internal/snapshot/ -run TestValidateCaseAlias -v` | 0 | PASS without skip: host volume is case-insensitive, direct `.AGENTS` semantic refused |
| `go vet ./internal/staging/ ./internal/snapshot/ ./internal/adapters/ ./internal/privatedir/` | 0 | Clean |
| `test -z "$(gofmt -l internal/staging internal/snapshot internal/adapters internal/privatedir)"` | 0 | `FMT_CLEAN`, no files listed |
| `go build -o /tmp/curator-14hsti ./cmd/curator` (binary removed) | 0 | CLI compiles |
| `git diff --check` | 0 | No whitespace errors |
| `golangci-lint run ./internal/staging/... ./internal/snapshot/... ./internal/adapters/... ./internal/privatedir/...` (initial) | 1 | Expected failure: 6 issues (3 package-comments, 1 redefines-builtin-id `real`, 2 gosec G602); fixed, see below |
| `golangci-lint run ...` (after fixes) | 0 | `0 issues.` |

The full landing suite was not run manually; the handoff owns the single remote-gate
execution. Other platforms are unverified. No installs, daemon restarts, tags,
releases, LOGBOOK.md edits, runtime-home changes, live credential export, or ax calls.

## Measured negative evidence

Every acceptance row is driven through a production entry point
(`snapshot.ValidateLocalPackage` / `ValidateRootInputs` / `EnumerateInputs`,
`staging.Plan.Snapshot`+`Recheck`, `adapters.StageProject`+`ValidateDestinations`,
`privatedir.TempStaging`):

- broad-root allow; managed-source / symlink-managed / case-alias refuse
  (`source_output_overlap`); write-boundary retarget refuses at recheck;
  root-without-inputs refuses; selector escape refuses (`source_selection_invalid`);
  unknown root-input alias refuses (`source_alias_unknown`); duplicate/overlapping/
  non-portable/linked root entries refuse; missing entries refuse
  (`source_member_missing`); special files and links in enumeration refuse
  (`source_member_invalid`); SKILL.md/manifest/declared-root coverage gaps refuse;
  admitted-overwrite both directions refuse at staging recheck and adapter
  validation; unmanaged takeover refuses via `StageProject` while the disjoint
  admitted check passes (allowlist does not authorize takeover).

Narrowing mutants, all killed (**3/3**, 0 survivors), bytes restored (`diff` clean):

| Narrowing mutation | Direct command | Real exit/result |
|---|---|---|
| Staging `checkAdmittedOverlap` keeps destination-within-admitted, drops reverse | `go test ./internal/staging/ -run TestPlanRecheckRefusesAdmittedOverwriteBothDirections` | 1, expected failure: admitted-inside-destination admitted |
| `IsOutputPath` keeps output containment, drops `.git` pruning | `go test ./internal/staging/ -run TestIsOutputPathPrunesGitAndOutputs` and `go test ./internal/snapshot/ -run TestEnumerateInputsPrunesAndRejects` | 1 and 1, expected failures: `.git` admitted in both packages |
| Adapters `ValidateDestinations` keeps destination-within-admitted, drops reverse | `go test ./internal/adapters/ -run TestValidateDestinationsRefusesAdmittedOverwrite` | 1, expected failure: reverse admitted-inside-destination admitted |

## Bounds and review needs

- Unicode-normalization aliases of missing paths are a stated blind spot
  (documented on `Within`): existing paths are covered through SameFile, absent
  NFD-equivalent spellings are not probed.
- `privatedir.TempStaging` narrowing mutants were not attacked: it is a
  MkdirTemp+Protect+Validate wrapper whose failure modes are platform-dependent
  (Unix MkdirTemp already yields 0700); tests prove private+validated success paths.
- Root-input "every required context input" is bounded to SKILL.md, the effective
  manifest file, and declared runtime/build roots; additional pipeline context
  inputs beyond those must be listed by the operator but are not individually
  cross-checked against the manifest here.
- Revision 2 policy shapes are hooks only: root-input validation is
  version-agnostic and no endpoint port/mirror/alias behaviour was added.
- Independent review must verify the exact published CR tree and rerun the narrow
  suites plus the remote gate.
