# TASK-260811-2gazym R1-R2 developer rework evidence

Status: developer evidence prepared for independent review

Run: `RUN-260811-04961c`

Authoritative run goal at the latest implementation checkpoint:
`GOAL-260811-7861b2` revision 1, resolved scope `TASK-260811-2gazym`

Review policy: required

Date: 2026-08-11

## Reviewed finding input

This rework addresses exactly R1 and R2 from
`TASK-260811-2gazym_review-verdict_RUN-260811-0f5cd8.md`, SHA-256
`fbfb103cb1505693b9aa39bba253db2552137eb225f24e197424d3e81ddc17c0`.
The orchestrator's follow-up directive also required the same pre-read bound
for BSD `ar` `#1/<len>` extended names.

## R1: causal trust authority

- Removed the exported `ManagerVerifier`, `ManagerVerifierConfig`,
  `ExternalToolchainPolicy`, staging-session, and output-verification API. An
  adapter can no longer create a verifier from caller-selected roots, plan
  digests, or expected bytes and receive trust authority.
- Removed production record-to-authorization issuer helpers. The production
  package exposes only `ToolchainAuthorization` and
  `LocalOutputAuthorization` interfaces with package-private methods. External
  packages cannot implement them, and `NewService` can verify but cannot mint
  either trust role.
- Clarified `LocalOutputAuthorization` as the opaque protected-executor
  receipt seam. No current production path manufactures it. The positive A08
  semantic seam therefore remains fail-closed until downstream protected
  execution supplies a causally valid receipt; tests use package-only receipt
  values solely to exercise the closed role codec and decision semantics.
- Added the external-package regression
  `TestExternalCallerCreatedRootsAndCopiedObjectsStayUntrusted`. A caller-built
  toolchain root receives `artifact_toolchain_untrusted`; adapter start count
  stays zero. A pre-existing ELF object copied into caller-created staging
  receives `artifact_local_output_unreceipted`; cache publication count stays
  zero.
- Routed reusable `T04/copied-preexisting` and
  `T04/hard-link-preexisting` through the public production
  `AdmitLocalOutput` path without a receipt. Package-private boolean mutation
  is no longer their conformance proof. Both exact manifests now correctly
  equal base T04 because caller assertions contribute no role evidence.
- The global compiled-dependency deny and unavailable `verified-binary-v1`
  behavior are unchanged.

## R2: declared metadata limits before payload reads

- Added non-mutating `preflightLeaf` validation so declared sizes reject before
  payload allocation while actual-read accounting remains honest.
- PAX local/global and GNU long-name/long-link tar metadata now run declared
  single-leaf and emitted-byte preflight before `readExactAt`, parsing, or
  hashing. Successfully read metadata is charged before parsing, so malformed
  metadata retains bounded unmanifested-byte evidence.
- GNU `ar` `//` string tables now pass leaf and emitted-byte preflight before
  allocation. The already-read table is reused for structural validation.
- BSD `ar` `#1/<len>` names now preflight the parsed name length for leaf,
  emitted-byte, and definite path-length bounds before `resolveARName` can
  read it. Effective member content accounting remains separate from header
  name metadata.
- Added bounded invalid-metadata precedence cases for PAX local/global and GNU
  long-name/link records. Added sparse default-limit cases for a 257 MiB PAX
  record, GNU string table, and BSD extended name. Each asserts the first
  deterministic limit diagnostic, exact synthetic path, and less than 32 MiB
  allocation before rejection.

## Exact conformance

`conformance.Cases()` still publishes exactly 182 A/C/F/T/V cases. The corpus
source SHA-256 is
`87a5cb6afb1c120cf75979cccd57fe2702c9a7dd74bee22dfa80418e1f26750e`.
Only the two T04 expected digests changed, both to the existing base T04 digest
`sha256:ba6bd3ed536380d1e181e6435f956de9ef65e913a2dffcbcd932643f582a6ce8`.
All other pinned outcomes and digests remain unchanged.

## Validation evidence

Every command below ran directly as a standalone process. No gate was piped
through `tee` or had its status obscured.

| Command | Exit | Evidence |
| --- | ---: | --- |
| focused R1/R2 adversarial command, including external caller, tar metadata, GNU string table, BSD extended name, forged receipt, and malformed metadata cases | 0 | passed in 0.593s |
| `go test -count=1 ./internal/artifactpolicy -run '^TestReusableArtifactManifestV1ConformanceCorpus$'` | 0 | all 182 exact cases passed in 0.810s |
| `go test -short -count=1 ./internal/artifactpolicy/...` | 0 | artifactpolicy passed in 8.192s |
| `go test -count=1 ./internal/artifactpolicy/...` | 0 | artifactpolicy passed in 31.731s |
| targeted final-source `go test -race` over R1/R2 and exact corpus | 0 | passed in 5.797s |
| `go test -race -count=1 ./internal/artifactpolicy/...` | 0 | full package race suite passed in 230.938s |
| `go test -short -count=1 -cover ./internal/artifactpolicy/...` | 0 | implementation package coverage 75.3%; fixture-only conformance subpackage 0.0% |
| `go vet ./internal/artifactpolicy/...` | 0 | no findings |
| `go build ./internal/artifactpolicy/...` | 0 | package compiled |
| `gofmt -l internal/artifactpolicy` | 0 | no files listed |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./internal/artifactpolicy/...` | 0 | pinned v2.12.2, `0 issues.` |
| public caller-configured issuer symbol absence check | 0 | no `NewManagerVerifier`, config, staging, or verifier API remains |
| `go test -count=1 ./internal/buildsource ./internal/buildmeta ./internal/buildcache ./internal/godriver ./internal/buildrepo ./internal/install/...` | 0 | Go baseline regressions passed; godriver 42.091s, buildrepo 19.901s, install 78.974s, atomicity 86.960s |
| `go test -count=1 ./...` | 0 | every repository package passed; cmd/curator 348.361s, artifactpolicy 140.395s, closuregraph 12.412s, godriver 67.314s, install 108.286s, atomicity 110.217s |
| `go vet ./...` | 0 | no findings |
| `go build ./...` | 0 | repository compiled |
| `/Users/iv/go/1.25.5/bin/golangci-lint run ./...` | 0 | `0 issues.`; one non-failing processor warning references a deleted stale integration worktree under `/private/tmp` |
| `git diff --check` | 0 | no tracked whitespace errors |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

## Truthful repair-loop failures

- The first exact-corpus rerun exited 1 because the two T04 digests still bound
  the removed caller-minted role evidence. Semantic class, decision, code, and
  path assertions reached the intended result. The two exact digests were
  audited and updated; the same command now exits 0.
- The first short package rerun exited 1 after an initial refactor used a
  zero-sized seal type (distinct pointers are permitted to compare equal in Go)
  and successful metadata preflight mutated leaf accounting before malformed
  bytes were charged. A non-zero seal identity was restored, declared
  preflight was separated from actual-read accounting, and the same gate now
  exits 0.

Neither failure is represented as a green checklist item.

## Source-stability evidence

- Artifactpolicy sorted file/content fingerprint:
  `d9a811e9f5ffadfcd4fda90b6a13b69ad3ba5a7a068e70fb1d12cd46a0bf7e1d`.
- The exclusive repository sequence covered 362 sorted `.go`, `go.mod`, and
  `go.sum` files. Its fingerprint was
  `b6bb88c574f46754fba91e588a8c2215be68fb14f465acd5c37233b4833c204e`
  before full test, immediately afterward, and after full vet/build/lint.
- Concurrent closuregraph work was accepted and source-stable before the full
  lane was released. No sibling source changed during artifactpolicy gates.

## Review focus

Independent review should verify that no public caller-configured trust issuer
remains; external packages cannot implement the opaque receipt interfaces;
T04 uses the public fail-closed path; and all PAX/GNU/BSD metadata allocation
paths execute declared leaf/emitted/path preflight before their first payload
read. Kotlin remains excluded.
