# TASK-260910-24cuys — review handoff

## Candidate and contract

Base/checkpoint HEAD: `4f27ccb21fd7c7b5f449466c6c858bf9b8108940`.
Uncommitted candidate; the developer handoff publishes the authoritative CR tree.
The accompanying candidate-sha256 artifact identifies all 49 changed/new source,
test and fixture files, including unchanged copies of 41 published vectors.
Specification checkout: `3535d63ea80f97bba2fcb6e1f06996cfc25cf7df`, containing
landed source contract `a4fcaf02`. Read protocol/skillfile-sources.md,
protocol/repository-transport.md and the draft skillfile schema with v1 references.
Task/Story outcomes were inspected: no earlier implementation outcome was attached.

## Implementation

- Reader-owned `ParseOptions.DraftSourcesV1` admits unreleased schema 2 via
  `LoadWithOptions`, `ParseBytesWithOptions`, and `ParseWithOptions`.
- Existing Load/ParseBytes/Parse remain schema-1 readers. No CLI/release capability
  claim is added; later resolver work must explicitly opt in.
- Source union preserves literal native paths, Git endpoint spelling, explicit
  canonical repository identity and exact tag/branch/revision declarations.
- Decl.Selector distinguishes individual and collection selections from legacy
  configured-root declarations. Declaration order, include/exclude and directory
  spellings are retained without acquisition or expansion.
- Unknown aliases/fields, mixed arms, unsafe directories, invalid refs/endpoints,
  duplicate direct names and unsupported versions fail during parsing.
- All existing legacy parsing fields/defaults remain, including source relative
  to configured skills_root and the legacy endpoint/ref grammar.
- Existing protocoljson duplicate-key/Unicode admission remains in the byte
  entry point. Existing skillspec/package capability behavior is unchanged.
- A draft-specific identity ref validator uses the schema's Unicode scalar
  length bound; the build validator's legacy byte-length limit is unchanged.

## Direct validation (zsh, standalone commands, no pipes)

All commands below were run by this producer; no prior evidence was accepted.

| Command | Exit | Evidence |
|---|---:|---|
| `go test ./internal/manifest` (initial and after new fixtures) | 0 each | Existing tests and then draft cases green |
| `go test ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec` | 0 | Scoped suites green |
| `go build ./cmd/curator` | 0 | CLI compiles; generated binary removed |
| `go vet ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec` | 0 | Ran before and after mutations |
| `go test ./internal/closure -run 'Test' -count=1` | 0 | Earlier regression run; superseded by restored-candidate rerun below |
| `go test ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec ./internal/closure -count=1` | 0 | Restored candidate: all five packages green (closure 39.234s) |
| `go build -o .temp/TASK-260910-24cuys-curator ./cmd/curator` | 0 | Restored candidate CLI build |
| `git diff --check` | 0 | No whitespace errors |
| `test -z "$(gofmt -l internal/manifest/manifest.go internal/manifest/sources.go internal/manifest/sources_test.go internal/identity/draft_sources.go)"` | 0 | All changed Go files formatted |
| `golangci-lint run ./internal/manifest/... ./internal/identity/...` | 127 | NOT RUN: executable unavailable; lint checklist left unchecked |

The full landing suite was not run manually; handoff owns its single execution.
Other platforms are unverified. Some existing external-conformance tests skip
without CURATOR_CONFORMANCE_ROOT; draft corpus tests are committed and unconditional.
No package installation, runtime-home changes, live credential export or ax calls.

## Measured negative evidence

Published schema-labeled cases: **41/41** driven through production
`manifest.LoadWithOptions`, with **41/41** copied payloads verified byte-identical.
Additional tests drive `ParseBytesWithOptions` for capability admission, every
legacy field, source union retention, unsafe selectors, JSON duplicate keys,
case-sensitive alias lookup, exact refs, URL arms and Unicode ref bounds.

Narrowing mutants caught: **3/3**, **0 survivors among these three experiments**.

| Narrowing mutation | Direct command | Real exit/result |
|---|---|---|
| Directory glob gate checks `*?` but stops checking brackets | `go test ./internal/manifest -run '^TestDraftNegativeAdmission$' -count=1` | 1, expected failure: admitted `a[b]` |
| Ref gate rejects only exact `..`, admitting embedded `..` | `go test ./internal/manifest -run '^TestDraftRefGrammarBounds$' -count=1` | 1, expected failure: admitted tag and branch `a..b` |
| Capability gate admits minimal v2 objects without opt-in | `go test ./internal/manifest -run '^TestDraftCapabilityAdmission$' -count=1` | 1, expected failure: default byte and file readers admitted v2 |

All mutations were reverted before the final scoped test/build/vet runs and
candidate hash capture. These experiments do not establish exhaustive mutation
coverage of every parser predicate.

## Bounds and review needs

Schema-backed means the published fixture labels are the oracle; Go tests do not
execute a JSON Schema engine. Python's jsonschema import was unavailable
(ModuleNotFoundError); no dependency was installed. Schema files are preserved
with fixtures for independent inspection. This parsing task does not establish
filesystem containment/symlink behavior, member discovery, snapshots/locks,
transport policy/fallback, audit or installation. Source paths are deliberately
not probed by the parser. Independent review must verify the exact published CR.

See the task-scoped logbook outcome for the ref-length decision and lint limit.

## Handoff blocker

`task-board handoff TASK-260910-24cuys --role developer` exited **1**:
`unchecked checklist items [5] (Lint clean): handoff evidence missing`.
No CR was published by that attempt and the task is not ready for review yet.
The `golangci-lint` command is unavailable (exit 127). A corrected bounded search
of existing /usr/local, ~/.local, ~/go and ~/Developer roots for a binary named
`golangci-lint` returned exit 1/no matches (excluded node_modules, .git,
.task-board and .temp). Earlier searches returned exit 2 due to absent
/opt/homebrew and are not absence evidence. This is not an exhaustive host search.

Campaign rules explicitly prohibit installs in this run. No install was attempted;
no lint checklist claim or handoff bypass was made. Required external input:
provision an approved golangci-lint binary (or its existing absolute path), or
reroute this candidate to an environment that has it. Recommended: provision the
project-compatible binary and resume the same task/worktree; run lint, address any
findings, rerun affected tests/build, update evidence and retry developer handoff.
The alternative is an explicit owner change to the required validation policy,
which weakens the gate and is not recommended. The implementation is retained
uncommitted with its attached evidence and exact-file hashes.


---

# TASK-260910-24cuys — resumed producer validation

Resumed RUN-260915-952d97 at base 4f27ccb21fd7c7b5f449466c6c858bf9b8108940.
All 49 candidate files matched the prior attached SHA-256 inventory before edits
(`shasum -a 256 -c <attached inventory>`, exit 0). The prior implementation,
fixture provenance, and three mutation experiments were inspected in the attached
results. Those mutation experiments were NOT rerun in this continuation; their
results remain historical evidence with the bounds already stated there.

The provisioned golangci-lint 2.12.2 resolved the external blocker. Initial
`golangci-lint run ./internal/manifest/... ./internal/identity/...` exited 1 with
two staticcheck QF1001 findings. Applied equivalent De Morgan transformations
to the hex-character and collection-member predicates in sources.go. No other
production/test/fixture bytes changed; the updated 49-file inventory supersedes
the earlier candidate hash inventory. No commits or runtime-home edits made.

## Commands personally rerun (zsh, set -o pipefail for Go/lint, no pipes)

| Command | Exit |
|---|---:|
| `go test ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec -count=1` (before lint fix) | 0 |
| `go build -o .temp/TASK-260910-24cuys-curator ./cmd/curator` (before fix) | 0 |
| `golangci-lint run ./internal/manifest/... ./internal/identity/...` (after fix) | 0; zero issues |
| `go test ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec ./internal/closure -count=1` (after fix) | 0; all five packages |
| `go build -o .temp/TASK-260910-24cuys-curator ./cmd/curator` (after fix) | 0 |
| `go vet ./internal/manifest ./internal/identity ./internal/protocoljson ./internal/skillspec` | 0 |
| `git diff --check` | 0 |
| `test -z "$(gofmt -l internal/manifest/manifest.go internal/manifest/sources.go internal/manifest/sources_test.go internal/identity/draft_sources.go)"` | 0 |

The unconditional draft corpus remains 41/41 schema-labeled vectors through
LoadWithOptions. Schema-engine execution, installation behavior and other
platforms remain unverified as described in the prior results. No full landing
suite was run manually; runtime handoff owns that execution. Independent exact-CR
review remains required. This addendum supersedes the prior lint/handoff blocker.
