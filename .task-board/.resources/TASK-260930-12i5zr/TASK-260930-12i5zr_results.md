# TASK-260930-12i5zr results

## Suite identities

Both suite manifests retain the protocol label `1.0.0-rc.13`, so the coverage ratchets select suites by the SHA-256 of `conformance/v1/manifest.json`.

| Suite | Source | Commit | Manifest SHA-256 |
|---|---|---|---|
| rc.13 / SPEC_PIN | curator-spec rc.13 | `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` | `be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca` |
| PR #116 candidate | `spec-2vapkz-content-hash-v2` | `526a9aa067a02d4d0b576dd7176c1137e2d7487d` | `950ee74ad148615c273fe95bbb93f1bc0f9bdf2ea2bd9395f1dc8e3601419e60` |

## Per-family results

Counts are exact whole-family pins. `driven / known-gap / count` cells describe the entire family on rows with full-family tallies. For the four delta-only rows, the first two values classify only the newly added case; the count remains the exact size of the whole indexed family. `—` means the candidate-only case/family is absent from that suite. The source corpus rows also include the separately stated bound count.

| Family | rc.13: driven / known-gap / count | Candidate: driven / known-gap / count | Scope |
|---|---:|---:|---|
| `marker/install-marker-v4/schema-cases` | 22 / 5 / 27 | 23 / 5 / 28 | Full family; the new frozen-shape case is driven. |
| `agent-environment-marker-v2/schema-cases` | 26 / 0 / 26 | 27 / 0 / 27 | Full family; the new frozen-shape case is driven. |
| `skillfile-sources-v1/schema-cases` | 118 / 0 / 121 (+3 bound) | 121 / 7 / 131 (+3 bound) | Full sibling corpus; candidate’s 7 v2-only cases are owned known gaps. |
| `marker/install-marker-v3/schema-cases` | — / — / 27 | 1 / 0 / 28 | Delta-only: added frozen marker case. |
| `context-lock-v1/schema-cases` | — / — / 31 | 1 / 0 / 32 | Delta-only: added frozen lock case. |
| `audit-record-v1/schema-cases` | — / — / 2 | 1 / 0 / 3 | Delta-only: added frozen record case. |
| `registry-log-entry-v1/schema-cases` | — / — / 2 | 1 / 0 / 3 | Delta-only: added frozen embedded-record case. |
| `agent-environment-marker-v3/schema-cases` | not published | 0 / 28 / 28 | Candidate-only v2 family; all cases are known gaps. |
| `audit-record-v2/schema-cases` | not published | 0 / 3 / 3 | Candidate-only v2 family; all cases are known gaps. |
| `context-lock-v2/schema-cases` | not published | 0 / 4 / 4 | Candidate-only v2 family; all cases are known gaps. |
| `marker/install-marker-v5/schema-cases` | not published | 0 / 29 / 29 | Candidate-only v2 family; all cases are known gaps. |
| `log-response-v3/schema-cases` | not published | 0 / 2 / 2 | Candidate-only v2 family; all cases are known gaps. |
| `manager-config-v3/schema-cases` | not published | 0 / 5 / 5 | Candidate-only v2 family; all cases are known gaps. |
| `registry-bundle-v2/schema-cases` | not published | 0 / 2 / 2 | Candidate-only v2 family; all cases are known gaps. |
| `registry-log-entry-v2/schema-cases` | not published | 0 / 2 / 2 | Candidate-only v2 family; all cases are known gaps. |
| `content-hashes-v2/vectors` | not published | 0 / 5 / 5 | Five executable vectors; metadata is excluded from the vector count. |

The candidate’s 87 new v2 gap rows (75 main-root schema cases, 7 sibling-corpus schema cases, and 5 hash vectors) all name `TASK-260917-2tx81l`. They activate only under the candidate manifest identity. rc.13 reports no v2 gap rows.

## Frozen-shape negatives

The candidate negatives for marker v3/v4, context-lock v1, environment marker v2, and audit-record v1 are rejected at their production readers. For each frozen object with a single new top-level `hash_version`, a control that removes only that member is accepted. The sibling source corpus also drives the frozen marker v5, lock v1, and source-audit v1 cases with accepted controls. The registry-log-entry v1 case rejects its embedded v2 record through `ParseRecord`.

## Failures and bounded runs

The first candidate package run exposed these exact stale-pin failures, both fixed by selecting exact count tables by manifest digest:

- `marker/install-marker-v4/schema-cases` published 28 cases, pinned 27.
- `agent-environment-marker-v2/schema-cases` published 27 cases, pinned 26.

The initial all-at-once `internal/crossconformance` run was interrupted (exit 1) while Git-backed semantic cases were running. A first grouped semantic attempt was also interrupted at 76.6 seconds (exit 1) to split the work further. All cross-conformance test functions were then run with bounded `-run` masks under both roots: schema/corpus consumers, digest and transport cases, all five semantic batches plus matrix coverage, CLI/playbook/cross-compile cases, external status/replay/snapshot/lifecycle cases, and protocol/export/adapter cases. Each replacement shard exited 0 on both roots.

A broad exploratory context-lock sweep also hit the pre-existing `invalid-source-with-git-suffix.json` case before the test was narrowed to the PR-added hash-version negative. That older case is outside this task’s added-case delta; it was not changed or counted as a passing classification. The added context-lock frozen-shape negative and the normal `internal/contextlock` package suite pass on both roots.

An expanded all-tests command over additional schema-reader packages was interrupted (exit 1) after more than five minutes in `internal/envprofile`’s `TestInstallSurfacesSystemModuleWarning`, which does not read conformance cases. The root-aware tests in those additional packages were rerun directly and passed on both roots. No full `internal/envprofile` package result is claimed.

## Verification and exit codes

- Main conformance package set (`conformancecoverage`, `marker`, `contextlock`, `envmarker`, `registry`, `manifest`, `skillspec`, `config`): rc.13 exit 0; candidate exit 0.
- Additional root-aware conformance consumers (`envprofile`, `envfragment`, `devsub`, `buildrepo`, `buildmeta`, `sourcelock`) with their schema/vector test-name mask: rc.13 exit 0; candidate exit 0.
- Full `internal/conformancecoverage` package after the ownership assertion: rc.13 exit 0; candidate exit 0.
- `TestContentHashV2*` gap classification: rc.13 exit 0 (candidate-only families absent); candidate exit 0 with exact per-family gap tallies above and 87/87 rows owned by `TASK-260917-2tx81l`.
- Count mutant: candidate marker v4 pin changed from 28 to 29; matching-root `go test ./internal/marker -run '^TestReadAuthoritativeMarkerV4SchemaCases$' -count=1` failed with exit 1 and `publishes 28 cases, want pinned count 29`. Restored the table (`cmp` exit 0); the same candidate test then passed with exit 0.
- `make lint`: exit 0 (`0 issues`).
- `go vet` over changed and conformance-reader packages: exit 0.
- `git diff --check`: exit 0. `gofmt -l` over changed Go files: exit 0 with no files listed.

No content-hash-v2 implementation was added. No CHANGELOG or LOGBOOK files were edited.
